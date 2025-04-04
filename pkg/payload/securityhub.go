package payload

import (
	"context"
	"fmt"
	"time"

	hub "github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/payload/scutils"
)

var schema = toPointer("2018-10-08")

type HubClient interface {
	BatchImportFindings(ctx context.Context, params *hub.BatchImportFindingsInput, optFns ...func(*hub.Options)) (*hub.BatchImportFindingsOutput, error)
	GetFindings(ctx context.Context, params *hub.GetFindingsInput, optFns ...func(*hub.Options)) (*hub.GetFindingsOutput, error)
	BatchUpdateFindings(ctx context.Context, params *hub.BatchUpdateFindingsInput, optFns ...func(*hub.Options)) (*hub.BatchUpdateFindingsOutput, error)
}

func toPointer[T any](v T) *T {
	return &v
}

func (p *BatchPolr) ToSecurityHubFindings(scConf scutils.SecurityHubConfig) []types.AwsSecurityFinding {
	results := []v1alpha2.PolicyReportResult{}

	for _, pl := range p.Results {
		results = append(results, pl.Result)
	}

	return helper.Map(results, func(result v1alpha2.PolicyReportResult) types.AwsSecurityFinding {
		generator := result.Policy
		if generator == "" {
			generator = result.Rule
		}

		title := generator
		if result.HasResource() {
			title = fmt.Sprintf("%s: %s", title, result.ResourceString())
		}

		t := time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos))

		return types.AwsSecurityFinding{
			Id:            toPointer(result.GetID()),
			AwsAccountId:  &scConf.AccountID,
			SchemaVersion: schema,
			ProductArn:    &scConf.ProductARN,
			GeneratorId:   toPointer(fmt.Sprintf("%s/%s", result.Source, generator)),
			Types:         []string{mapType(result.Source)},
			CreatedAt:     toPointer(t.Format("2006-01-02T15:04:05.999999999Z07:00")),
			UpdatedAt:     toPointer(t.Format("2006-01-02T15:04:05.999999999Z07:00")),
			Severity: &types.Severity{
				Label: MapSeverity(result.Severity),
			},
			Title:       &title,
			Description: &result.Message,
			ProductName: &scConf.ProductName,
			CompanyName: &scConf.CompanyName,
			Compliance: &types.Compliance{
				Status: types.ComplianceStatusFailed,
			},
			Workflow: &types.Workflow{
				Status: types.WorkflowStatusNew,
			},
			Resources: []types.Resource{
				{
					Type:      toPointer("Other"),
					Region:    &scConf.Region,
					Partition: types.PartitionAws,
					Id:        mapResourceID(result),
					// Details: &types.ResourceDetails{
					// 	Other: mapOtherDetails(polr, result),
					// },
				},
			},
			RecordState: types.RecordStateActive,
		}
	})
}

func MapSeverity(s v1alpha2.PolicySeverity) types.SeverityLabel {
	switch s {
	case v1alpha2.SeverityInfo:
		return types.SeverityLabelInformational
	case v1alpha2.SeverityLow:
		return types.SeverityLabelLow
	case v1alpha2.SeverityMedium:
		return types.SeverityLabelMedium
	case v1alpha2.SeverityHigh:
		return types.SeverityLabelHigh
	case v1alpha2.SeverityCritical:
		return types.SeverityLabelCritical
	default:
		return types.SeverityLabelInformational
	}
}

func mapType(source string) string {
	if source == "" {
		return "Software and Configuration Checks/Kubernetes Policies"
	}
	return "Software and Configuration Checks/Kubernetes Policies/" + source
}

func mapResourceID(result v1alpha2.PolicyReportResult) *string {
	if result.HasResource() {
		res := result.GetResource()
		if res.UID != "" {
			return toPointer(string(res.UID))
		}

		return toPointer(result.ResourceString())
	}
	return toPointer(result.GetID())
}

func mapOtherDetails(polr v1alpha2.ReportInterface, result v1alpha2.PolicyReportResult) map[string]string {
	details := map[string]string{
		"Source":   result.Source,
		"Category": result.Category,
		"Policy":   result.Policy,
		"Rule":     result.Rule,
		"Result":   string(result.Result),
		"Report":   polr.GetKey(),
	}

	if len(c.customFields) > 0 {
		for property, value := range c.customFields {
			details[property] = value
		}

		for property, value := range result.Properties {
			details[property] = value
		}
	}

	if result.HasResource() {
		res := result.GetResource()

		if res.APIVersion != "" {
			details["Resource APIVersion"] = res.APIVersion
		}
		if res.Kind != "" {
			details["Resource Kind"] = res.Kind
		}
		if res.Namespace != "" {
			details["Resource Namespace"] = res.Namespace
		}
		if res.Name != "" {
			details["Resource Name"] = res.Name
		}
		if res.UID != "" {
			details["Resource UID"] = string(res.UID)
		}
	}

	return details
}
