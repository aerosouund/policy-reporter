package payload

import (
	"fmt"
	"strings"
	"time"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target/formatting"
)

var (
	keyReplacer   = strings.NewReplacer(".", "_", "]", "", "[", "")
	labelReplacer = strings.NewReplacer("/", "")
)

type Payload interface {
	CreationTimestamp() time.Time
	Body() map[string]interface{}
	ToLoki(map[string]string) Stream
	ToTelegram()
	ToTeams()
	ToSlack()
	ToDiscord()
	BlobStorageKey(string) string
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

func (s *client) newMessage(resource *corev1.ObjectReference, results []v1alpha2.PolicyReportResult) *adaptivecard.Message {
	header := adaptivecard.NewContainer()

	if resource != nil {
		header.AddElement(false, adaptivecard.NewTitleTextBlock(formatting.ResourceString(resource), true))
	} else {
		header.AddElement(false, adaptivecard.NewTitleTextBlock("New PolicyReport Results", true))
	}

	header.AddElement(false, adaptivecard.NewTextBlock(fmt.Sprintf("Received %d new Policy Report Results", len(results)), true))

	if len(s.customFields) > 0 {
		header.AddElement(false, MapToColumnSet(s.customFields))
	}

	card := adaptivecard.NewCard()
	card.SetFullWidth()
	card.AddContainer(true, header)

	for _, result := range results {
		stats := newFactSet()
		stats.Facts = append(stats.Facts, adaptivecard.Fact{Title: "Status", Value: string(result.Result)})

		if result.Severity != "" {
			stats.Facts = append(stats.Facts, adaptivecard.Fact{Title: "Severity", Value: string(result.Severity)})
		}

		policy := fmt.Sprintf("Policy: %s", result.Policy)

		if result.Rule != "" {
			policy = fmt.Sprintf("%s/%s", policy, result.Rule)
		}

		r := adaptivecard.NewContainer()
		r.Separator = true
		r.Spacing = adaptivecard.SpacingLarge
		r.AddElement(false, newSubTitle(policy))
		r.AddElement(false, adaptivecard.NewTextBlock(result.Category, true))
		r.AddElement(false, stats)
		r.AddElement(false, adaptivecard.NewTextBlock(result.Message, true))

		if len(result.Properties) > 0 {
			r.AddElement(false, MapToColumnSet(result.Properties))
		}

		card.AddContainer(false, r)
	}

	msg := adaptivecard.NewMessage()
	msg.Attach(card)

	return msg
}

func (s *PolicyReportResultPayload) ToTelegram()

func (s *PolicyReportResultPayload) ToSlack()
