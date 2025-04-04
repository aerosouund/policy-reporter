package payload

import (
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/kyverno/policy-reporter/pkg/payload/scutils"
)

type BatchPolr struct {
	Results []PolicyReportResultPayload
}

type BatchPayload interface {
	ToPayloadSlice() []Payload
	ToSecurityHubFindings(scConf scutils.SecurityHubConfig) []types.AwsSecurityFinding
	Filter([]Payload) BatchPayload
}

func NewBatchPayload(pls []Payload) BatchPayload {
	return &BatchPolr{
		Results: pls,
	}
}

func (p *BatchPolr) Filter(existing []Payload) BatchPayload {
	filtered := []PolicyReportResultPayload{}
	mapping := make(map[string]bool)

	for _, p := range existing {
		mapping[p.GetID()] = true
	}

	for _, p := range p.Results {
		if _, ok := mapping[p.GetID()]; ok {
			continue
		}
		filtered = append(filtered, p)
	}

	return &BatchPolr{
		Results: filtered,
	}
}

func (p *BatchPolr) ToPayloadSlice() []Payload {
	pls := []Payload{}
	for _, p := range p.Results {
		pls = append(pls, &p)
	}
	return pls
}
