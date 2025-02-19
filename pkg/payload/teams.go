package payload

import (
	"fmt"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target/formatting"
)

func (p *PolicyReportResultPayload) ToTeams(result []v1alpha2.PolicyReportResult) *adaptivecard.Message {
	header := adaptivecard.NewContainer()

	header.AddElement(false, adaptivecard.NewTitleTextBlock(formatting.ResourceString(p.Result.GetResource()), true))

	header.AddElement(false, adaptivecard.NewTextBlock(fmt.Sprintf("Received %d new Policy Report Results", len(results)), true))

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

func newFactSet() adaptivecard.Element {
	factSet := adaptivecard.Element{
		Type: adaptivecard.TypeElementFactSet,
	}

	return factSet
}

func newFactSetPointer() *adaptivecard.Element {
	factSet := newFactSet()

	return &factSet
}

func newSubTitle(title string) adaptivecard.Element {
	text := adaptivecard.NewTextBlock(title, true)
	text.Weight = adaptivecard.WeightBolder
	text.IsSubtle = true

	return text
}

func MapToColumnSet(list map[string]string) adaptivecard.Element {
	i := 0

	first := adaptivecard.NewColumn()
	first.Items = append(first.Items, newFactSetPointer())

	second := adaptivecard.NewColumn()
	second.Items = append(second.Items, newFactSetPointer())

	propBlock := adaptivecard.NewColumnSet()
	propBlock.Columns = []adaptivecard.Column{first, second}

	for property, value := range list {
		index := i % 2

		propBlock.Columns[index].Items[0].Facts = append(propBlock.Columns[index].Items[0].Facts, adaptivecard.Fact{
			Title: helper.Title(property),
			Value: value,
		})

		i++
	}

	return propBlock
}
