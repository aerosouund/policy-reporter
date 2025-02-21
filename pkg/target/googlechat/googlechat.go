package googlechat

import (
	"github.com/kyverno/policy-reporter/pkg/http"
	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
)

// Options to configure the Discord target
type Options struct {
	target.ClientOptions
	Webhook      string
	Headers      map[string]string
	CustomFields map[string]string
	HTTPClient   http.Client
}

type client struct {
	target.BaseClient
	webhook      string
	headers      map[string]string
	customFields map[string]string
	client       http.Client
}

func (e *client) Send(result payload.Payload) {
	if len(e.customFields) > 0 {
		result.AddCustomFields(e.customFields)
	}
	payload, err := result.ToGoogleChat()
	// handle error ?

	req, err := http.CreateJSONRequest("POST", e.webhook, payload)
	if err != nil {
		return
	}

	for header, value := range e.headers {
		req.Header.Set(header, value)
	}

	resp, err := e.client.Do(req)
	http.ProcessHTTPResponse(e.Name(), resp, err)
}

func (e *client) Type() target.ClientType {
	return target.SingleSend
}

// NewClient creates a new loki.client to send Results to Elasticsearch
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Webhook,
		options.Headers,
		options.CustomFields,
		options.HTTPClient,
	}
}
