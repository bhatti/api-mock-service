package contract

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/bhatti/api-mock-service/internal/types"
)

// ReportExporter converts ProducerContractResponse into CI/CD-friendly formats.
type ReportExporter struct{}

// junitTestSuites is the top-level JUnit XML container.
type junitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	XMLName    xml.Name         `xml:"testsuite"`
	Name       string           `xml:"name,attr"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Errors     int              `xml:"errors,attr"`
	Time       float64          `xml:"time,attr"`
	Timestamp  string           `xml:"timestamp,attr"`
	Properties []junitProperty  `xml:"properties>property,omitempty"`
	TestCases  []junitTestCase  `xml:"testcase"`
}

type junitTestCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      float64       `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Body    string `xml:",chardata"`
}

type junitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// ExportJUnitXML produces JUnit XML from a ProducerContractResponse.
func (e *ReportExporter) ExportJUnitXML(group string, response *types.ProducerContractResponse) ([]byte, error) {
	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	suite := junitTestSuite{
		Name:      group,
		Tests:     response.Succeeded + response.Failed,
		Failures:  response.Failed,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	suite.Properties = append(suite.Properties,
		junitProperty{Name: "succeeded", Value: fmt.Sprintf("%d", response.Succeeded)},
		junitProperty{Name: "failed", Value: fmt.Sprintf("%d", response.Failed)},
	)
	if response.Coverage != nil {
		suite.Properties = append(suite.Properties,
			junitProperty{Name: "coverage_percent", Value: fmt.Sprintf("%.1f", response.Coverage.Coverage)},
		)
	}
	if response.SecuritySummary != nil {
		suite.Properties = append(suite.Properties,
			junitProperty{Name: "injection_findings", Value: fmt.Sprintf("%d", response.SecuritySummary.TotalFindings)},
		)
	}
	if len(response.FailureClusters) > 0 {
		suite.Properties = append(suite.Properties,
			junitProperty{Name: "failure_clusters", Value: fmt.Sprintf("%d", len(response.FailureClusters))},
		)
	}

	for key := range response.Results {
		tc := junitTestCase{
			Name:      key,
			Classname: group,
		}
		if t, ok := response.Metrics[key]; ok {
			tc.Time = t
		}
		suite.TestCases = append(suite.TestCases, tc)
	}

	for key, errMsg := range response.Errors {
		tc := junitTestCase{
			Name:      key,
			Classname: group,
			Failure: &junitFailure{
				Message: truncateForXML(errMsg, 500),
				Type:    "ContractValidationError",
				Body:    errMsg,
			},
		}
		if t, ok := response.Metrics[key]; ok {
			tc.Time = t
		}
		suite.TestCases = append(suite.TestCases, tc)
	}

	var totalTime float64
	for _, m := range response.Metrics {
		totalTime += m
	}
	suite.Time = totalTime

	suites := junitTestSuites{Suites: []junitTestSuite{suite}}
	out, err := xml.MarshalIndent(suites, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JUnit XML: %w", err)
	}
	return append([]byte(xml.Header), out...), nil
}

// JSONSummary is a structured summary of a contract test run for CI/CD dashboards.
type JSONSummary struct {
	Group           string                  `json:"group"`
	Timestamp       string                  `json:"timestamp"`
	Succeeded       int                     `json:"succeeded"`
	Failed          int                     `json:"failed"`
	Total           int                     `json:"total"`
	PassRate        float64                 `json:"passRate"`
	Coverage        *types.CoverageSummary  `json:"coverage,omitempty"`
	SecuritySummary *types.SecuritySummary  `json:"securitySummary,omitempty"`
	FailureClusters []types.FailureCluster  `json:"failureClusters,omitempty"`
	TopErrors       []string                `json:"topErrors,omitempty"`
}

// ExportJSONSummary produces a concise JSON summary from a ProducerContractResponse.
func (e *ReportExporter) ExportJSONSummary(group string, response *types.ProducerContractResponse) ([]byte, error) {
	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	total := response.Succeeded + response.Failed
	var passRate float64
	if total > 0 {
		passRate = float64(response.Succeeded) / float64(total) * 100.0
	}

	summary := JSONSummary{
		Group:           group,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		Succeeded:       response.Succeeded,
		Failed:          response.Failed,
		Total:           total,
		PassRate:        passRate,
		Coverage:        response.Coverage,
		SecuritySummary: response.SecuritySummary,
		FailureClusters: response.FailureClusters,
	}

	limit := 10
	if len(response.Errors) < limit {
		limit = len(response.Errors)
	}
	i := 0
	for _, msg := range response.Errors {
		if i >= limit {
			break
		}
		summary.TopErrors = append(summary.TopErrors, truncateForXML(msg, 300))
		i++
	}

	return json.MarshalIndent(summary, "", "  ")
}

func truncateForXML(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
