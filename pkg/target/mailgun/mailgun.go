package mailgun

import (
	"context"

	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/mailgun/mailgun-go/v4"
	"go.uber.org/zap"
)

type Options struct {
	target.ClientOptions
	CustomFields map[string]string
	Sender       string
	Mg           mailgun.Mailgun
}

func (c *client) Send(p payload.Payload) {
	emailMsg, err := p.ToEmail()
	if err != nil {
		zap.L().Error(c.Name()+": email conversion error", zap.Error(err))
		return
	}

	for _, recip := range emailMsg.Recipients {
		msg := mailgun.NewMessage(c.sender, emailMsg.Subject, emailMsg.Body, recip)
		c.mg.Send(context.TODO(), msg)
		zap.L().Info(c.Name() + ": email sent to " + recip)
	}
}

type client struct {
	target.BaseClient
	customFields map[string]string
	sender       string
	mg           mailgun.Mailgun
}

func (c *client) Type() target.ClientType {
	return target.SingleSend
}

func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.CustomFields,
		options.Sender,
		options.Mg,
	}
}
