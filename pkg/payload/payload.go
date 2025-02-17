package payload

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	corev1 "k8s.io/api/core/v1"
)

var (
	keyReplacer   = strings.NewReplacer(".", "_", "]", "", "[", "")
	labelReplacer = strings.NewReplacer("/", "")
)

type Payload interface {
	// CreationTimestamp() time.Time
	Body() map[string]interface{}
	ToLoki(map[string]string) Stream
	ToTelegram() (string, error)
	// ToTeams()
	ToSlack()
	// ToDiscord()
	BlobStorageKey(string) string
	KinesisKey() string
}

type PolicyReportResultPayload struct {
	Result v1alpha2.PolicyReportResult
}

func (p *PolicyReportResultPayload) BlobStorageKey(prefix string) string {
	t := time.Unix(p.Result.Timestamp.Seconds, int64(p.Result.Timestamp.Nanos))
	return fmt.Sprintf("%s/%s/%s-%s-%s.json", prefix, t.Format("2006-01-02"), p.Result.Policy, p.Result.ID, t.Format(time.RFC3339Nano))
}

// should be the equivalent of get json body
func (p *PolicyReportResultPayload) Body() {}

func (s *PolicyReportResultPayload) KinesisKey() string {
	t := time.Unix(s.Result.Timestamp.Seconds, int64(s.Result.Timestamp.Nanos))
	return fmt.Sprintf("%s-%s-%s", s.Result.Policy, s.Result.ID, t.Format(time.RFC3339Nano))
}

func (s *PolicyReportResultPayload) ToLoki(customFields map[string]string) Stream {
	timestamp := time.Now()
	if s.Result.Timestamp.Seconds != 0 {
		timestamp = time.Unix(s.Result.Timestamp.Seconds, int64(s.Result.Timestamp.Nanos))
	}

	labels := map[string]string{
		"status":    string(s.Result.Result),
		"policy":    s.Result.Policy,
		"createdBy": "policy-reporter",
	}

	if s.Result.Rule != "" {
		labels["rule"] = s.Result.Rule
	}
	if s.Result.Category != "" {
		labels["category"] = s.Result.Category
	}
	if s.Result.Severity != "" {
		labels["severity"] = string(s.Result.Severity)
	}
	if s.Result.Source != "" {
		labels["source"] = s.Result.Source
	}
	if s.Result.HasResource() {
		res := s.Result.GetResource()
		if res.APIVersion != "" {
			labels["apiVersion"] = res.APIVersion
			labels["kind"] = res.Kind
			labels["name"] = res.Name
		}
		if res.UID != "" {
			labels["uid"] = string(res.UID)
		}
		if res.Namespace != "" {
			labels["namespace"] = res.Namespace
		}
	}

	for property, value := range s.Result.Properties {
		labels[keyReplacer.Replace(property)] = labelReplacer.Replace(value)
	}

	for label, value := range customFields {
		labels[keyReplacer.Replace(label)] = labelReplacer.Replace(value)
	}

	return Stream{
		Values: []Value{[]string{fmt.Sprintf("%v", timestamp.UnixNano()), "[" + strings.ToUpper(string(s.Result.Severity)) + "] " + s.Result.Message}},
		Stream: labels,
	}
}

func (s *PolicyReportResultPayload) ToTelegram(chatID string) (string, error) {
	// if len(e.customFields) > 0 {
	// 	props := make(map[string]string, 0)

	// 	for property, value := range e.customFields {
	// 		props[property] = value
	// 	}

	// 	for property, value := range result.Properties {
	// 		props[property] = value
	// 	}

	// 	result.Properties = props
	// }

	var textBuffer bytes.Buffer

	ttmpl, err := template.New("telegram").Funcs(template.FuncMap{"escape": escape}).Parse(notificationTempl)
	if err != nil {
		return "", err
	}

	var res *corev1.ObjectReference
	if s.Result.HasResource() {
		res = s.Result.GetResource()
	}

	err = ttmpl.Execute(&textBuffer, values{
		Result:   s.Result,
		Time:     time.Now(),
		Resource: res,
	})
	if err != nil {
		return "", err
	}

	return textBuffer.String(), nil
}

func (s *PolicyReportResultPayload) ToSlack()
