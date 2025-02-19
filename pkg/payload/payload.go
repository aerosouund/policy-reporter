package payload

import (
	"fmt"
	"strings"
	"time"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/slack-go/slack"
)

var (
	keyReplacer   = strings.NewReplacer(".", "_", "]", "", "[", "")
	labelReplacer = strings.NewReplacer("/", "")
)

type Payload interface {
	// CreationTimestamp() time.Time
	GetID() string

	Body() map[string]interface{}
	ToLoki(map[string]string) Stream
	ToTelegram(chatId string) (string, error)
	// ToTeams()
	ToSlack(channel string) *slack.Attachment
	ToDiscord(customFields map[string]string) DiscordPayload // todo: custom fields
	BlobStorageKey(string) string
	KinesisKey() string
	ToGoogleChat()
}

type PolicyReportResultPayload struct {
	Result v1alpha2.PolicyReportResult
}

func (p *PolicyReportResultPayload) GetID() string {
	return p.Result.GetID()
}

func (p *PolicyReportResultPayload) BlobStorageKey(prefix string) string {
	t := time.Unix(p.Result.Timestamp.Seconds, int64(p.Result.Timestamp.Nanos))
	return fmt.Sprintf("%s/%s/%s-%s-%s.json", prefix, t.Format("2006-01-02"), p.Result.Policy, p.Result.ID, t.Format(time.RFC3339Nano))
}

// should be the equivalent of get json body
func (p *PolicyReportResultPayload) Body() map[string]interface{} {
	return map[string]interface{}{}
}

func (s *PolicyReportResultPayload) KinesisKey() string {
	t := time.Unix(s.Result.Timestamp.Seconds, int64(s.Result.Timestamp.Nanos))
	return fmt.Sprintf("%s-%s-%s", s.Result.Policy, s.Result.ID, t.Format(time.RFC3339Nano))
}
