package targetconfig

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	tcv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/clientset/versioned"
	tcinformer "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/informers/externalversions"
	"github.com/kyverno/policy-reporter/pkg/target"
)

type TargetConfigClient struct {
	TargetFactory target.Factory
	TargetClients *target.Collection
	Logger        *zap.Logger

	tcClient  tcv1alpha1.Interface
	informer  cache.SharedIndexInformer
	tcCount   int
	hasSynced bool
}

type EventType string

const (
	DeleteTcEvent = "delete"
	CreateTcEvent = "create"
)

type TcEvent struct {
	Type                EventType
	Targets             *target.Collection
	RestartPolrInformer bool
}

func NewTargetConfigClient(tcClient tcv1alpha1.Interface, f target.Factory, targets *target.Collection, logger *zap.Logger) *TargetConfigClient {
	return &TargetConfigClient{
		TargetFactory: f,
		TargetClients: targets,
		Logger:        logger,
		tcClient:      tcClient,
	}
}

func (c *TargetConfigClient) TargetConfigCount() int {
	return c.tcCount
}

func (c *TargetConfigClient) CreateInformer(targetChan chan TcEvent, addFn, delFn func(interface{}), upFn func(interface{}, interface{})) error {
	tcInformer := tcinformer.NewSharedInformerFactory(c.tcClient, 0)
	inf := tcInformer.Policyreporter().V1alpha1().TargetConfigs().Informer()
	c.informer = inf

	tcs, err := c.tcClient.PolicyreporterV1alpha1().TargetConfigs("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	c.tcCount = len(tcs.Items)
	c.configureInformer(addFn, delFn, upFn)
	return nil
}

func (c *TargetConfigClient) HasSynced() bool {
	return c.hasSynced
}

func (c *TargetConfigClient) Run(stopChan chan struct{}) {
	go c.informer.Run(stopChan)

	if !cache.WaitForCacheSync(stopChan, c.informer.HasSynced) {
		c.Logger.Error("Failed to sync target config cache")
		return
	}

	c.hasSynced = true
	c.Logger.Info("target config cache synced")
}

func (c *TargetConfigClient) configureInformer(addFn, delFn func(interface{}), upFn func(interface{}, interface{})) {
	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addFn,
		UpdateFunc: upFn,
		DeleteFunc: delFn,
	})
}

func AddFn(c *TargetConfigClient, targetChan chan TcEvent) func(interface{}) {
	return func(obj interface{}) {
		tc := obj.(*v1alpha1.TargetConfig)
		c.Logger.Info(fmt.Sprintf("new target: %s", tc.Name))

		t, err := c.TargetFactory.CreateSingleClient(tc)
		if err != nil {
			c.Logger.Error("unable to create target from TargetConfig: " + err.Error())
			return
		}

		c.TargetClients.AddTarget(tc.Name, t)
		targetChan <- TcEvent{Type: CreateTcEvent, Targets: c.TargetClients, RestartPolrInformer: !tc.Spec.SkipExisting}
	}
}

func DelFn(c *TargetConfigClient, targetChan chan TcEvent) func(interface{}) {
	return func(obj interface{}) {
		tc := obj.(*v1alpha1.TargetConfig)
		c.Logger.Info(fmt.Sprintf("deleting target: %s", tc.Name))

		c.TargetClients.RemoveTarget(tc.Name)
		targetChan <- TcEvent{Type: DeleteTcEvent, Targets: c.TargetClients}
	}
}

func UpFn(c *TargetConfigClient, targetChan chan TcEvent) func(interface{}, interface{}) {
	return func(oldObj, newObj interface{}) {
		tc := newObj.(*v1alpha1.TargetConfig)
		c.Logger.Info(fmt.Sprintf("update target: %s", tc.Name))

		t, err := c.TargetFactory.CreateSingleClient(tc)
		if err != nil {
			c.Logger.Error("unable to create target from TargetConfig: " + err.Error())
			return
		}

		c.TargetClients.AddTarget(tc.Name, t)
		targetChan <- TcEvent{Type: CreateTcEvent, Targets: c.TargetClients, RestartPolrInformer: !tc.Spec.SkipExisting}

	}
}
