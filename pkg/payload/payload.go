package payload

import (
	"fmt"
	"strings"
	"time"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/http"
	"github.com/slack-go/slack"
)

var (
	keyReplacer   = strings.NewReplacer(".", "_", "]", "", "[", "")
	labelReplacer = strings.NewReplacer("/", "")
)

type EmailMsg struct {
	Recipients []string `json:"recipients"`
	Attachment []byte   `json:"attachment"`
	CC         []string `json:"cc"`
	Bcc        []string `json:"bcc"`
	Body       string   `json:"body"`
	Subject    string   `json:"subject"`
}

type Payload interface {
	GetID() string
	Body() http.Result
	ToLoki() Stream
	ToTelegram(chatId string) (string, error)
	ToTeams() adaptivecard.Container
	ToSlack(channel string) *slack.Attachment
	ToDiscord() DiscordPayload
	BlobStorageKey(string) string
	KinesisKey() string
	AddCustomFields(map[string]string)
	ToGoogleChat() (*GCPayload, error)
	ToEmail() (EmailMsg, error)
}

type PolicyReportResultPayload struct {
	Result v1alpha2.PolicyReportResult
}

func (p *PolicyReportResultPayload) GetID() string {
	return p.Result.GetID()
}

func (p *PolicyReportResultPayload) AddCustomFields(fieldMap map[string]string) {
	props := make(map[string]string, 0)

	for property, value := range fieldMap {
		props[property] = value
	}

	for property, value := range p.Result.Properties {
		props[property] = value
	}

	p.Result.Properties = props
}

func (p *PolicyReportResultPayload) BlobStorageKey(prefix string) string {
	t := time.Unix(p.Result.Timestamp.Seconds, int64(p.Result.Timestamp.Nanos))
	return fmt.Sprintf("%s/%s/%s-%s-%s.json", prefix, t.Format("2006-01-02"), p.Result.Policy, p.Result.ID, t.Format(time.RFC3339Nano))
}

func (p *PolicyReportResultPayload) Body() http.Result {
	return http.NewJSONResult(p.Result)
}

func (s *PolicyReportResultPayload) KinesisKey() string {
	t := time.Unix(s.Result.Timestamp.Seconds, int64(s.Result.Timestamp.Nanos))
	return fmt.Sprintf("%s-%s-%s", s.Result.Policy, s.Result.ID, t.Format(time.RFC3339Nano))
}
