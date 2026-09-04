package contract

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestResponse() *types.ProducerContractResponse {
	resp := types.NewProducerContractResponse()
	resp.Add("scenario-ok", map[string]any{"status": "ok"}, nil)
	resp.Add("scenario-fail", nil, &ContractValidationError{
		OriginalError: assert.AnError,
		Scenario:      "scenario-fail",
		URL:           "POST /api",
	})
	resp.Metrics["scenario-ok"] = 0.15
	resp.Metrics["scenario-fail"] = 0.42
	return resp
}

func TestReportExporter_ExportJUnitXML_Basic(t *testing.T) {
	e := &ReportExporter{}
	resp := newTestResponse()

	xmlBytes, err := e.ExportJUnitXML("test-group", resp)
	require.NoError(t, err)

	xmlStr := string(xmlBytes)
	assert.True(t, strings.HasPrefix(xmlStr, "<?xml"))
	assert.Contains(t, xmlStr, `name="test-group"`)
	assert.Contains(t, xmlStr, `tests="2"`)
	assert.Contains(t, xmlStr, `failures="1"`)

	var suites junitTestSuites
	err = xml.Unmarshal(xmlBytes[len(xml.Header):], &suites)
	require.NoError(t, err)
	require.Len(t, suites.Suites, 1)

	suite := suites.Suites[0]
	assert.Equal(t, 2, suite.Tests)
	assert.Equal(t, 1, suite.Failures)
}

func TestReportExporter_ExportJUnitXML_WithCoverageAndSecurity(t *testing.T) {
	e := &ReportExporter{}
	resp := newTestResponse()
	resp.Coverage = &types.CoverageSummary{Coverage: 85.5}
	summary := types.SecuritySummary{TotalFindings: 3}
	resp.SecuritySummary = &summary
	resp.FailureClusters = []types.FailureCluster{{Count: 5}}

	xmlBytes, err := e.ExportJUnitXML("group-with-extras", resp)
	require.NoError(t, err)

	xmlStr := string(xmlBytes)
	assert.Contains(t, xmlStr, `name="coverage_percent"`)
	assert.Contains(t, xmlStr, `value="85.5"`)
	assert.Contains(t, xmlStr, `name="injection_findings"`)
	assert.Contains(t, xmlStr, `value="3"`)
	assert.Contains(t, xmlStr, `name="failure_clusters"`)
	assert.Contains(t, xmlStr, `value="1"`)
}

func TestReportExporter_ExportJUnitXML_NilResponse(t *testing.T) {
	e := &ReportExporter{}
	_, err := e.ExportJUnitXML("group", nil)
	assert.Error(t, err)
}

func TestReportExporter_ExportJUnitXML_EmptyResponse(t *testing.T) {
	e := &ReportExporter{}
	resp := types.NewProducerContractResponse()

	xmlBytes, err := e.ExportJUnitXML("empty-group", resp)
	require.NoError(t, err)
	assert.Contains(t, string(xmlBytes), `tests="0"`)
}

func TestReportExporter_ExportJSONSummary_Basic(t *testing.T) {
	e := &ReportExporter{}
	resp := newTestResponse()

	jsonBytes, err := e.ExportJSONSummary("test-group", resp)
	require.NoError(t, err)

	var summary JSONSummary
	err = json.Unmarshal(jsonBytes, &summary)
	require.NoError(t, err)

	assert.Equal(t, "test-group", summary.Group)
	assert.Equal(t, 1, summary.Succeeded)
	assert.Equal(t, 1, summary.Failed)
	assert.Equal(t, 2, summary.Total)
	assert.InDelta(t, 50.0, summary.PassRate, 0.1)
	assert.NotEmpty(t, summary.Timestamp)
}

func TestReportExporter_ExportJSONSummary_WithClusters(t *testing.T) {
	e := &ReportExporter{}
	resp := newTestResponse()
	resp.FailureClusters = []types.FailureCluster{
		{Key: types.FailureClusterKey{Endpoint: "POST /api"}, Count: 3},
	}

	jsonBytes, err := e.ExportJSONSummary("test-group", resp)
	require.NoError(t, err)

	var summary JSONSummary
	err = json.Unmarshal(jsonBytes, &summary)
	require.NoError(t, err)
	require.Len(t, summary.FailureClusters, 1)
	assert.Equal(t, 3, summary.FailureClusters[0].Count)
}

func TestReportExporter_ExportJSONSummary_NilResponse(t *testing.T) {
	e := &ReportExporter{}
	_, err := e.ExportJSONSummary("group", nil)
	assert.Error(t, err)
}

func TestReportExporter_ExportJSONSummary_AllSucceeded(t *testing.T) {
	e := &ReportExporter{}
	resp := types.NewProducerContractResponse()
	resp.Add("s1", "ok", nil)
	resp.Add("s2", "ok", nil)

	jsonBytes, err := e.ExportJSONSummary("group", resp)
	require.NoError(t, err)

	var summary JSONSummary
	err = json.Unmarshal(jsonBytes, &summary)
	require.NoError(t, err)
	assert.InDelta(t, 100.0, summary.PassRate, 0.1)
	assert.Empty(t, summary.TopErrors)
}

func TestReportExporter_ExportJSONSummary_TopErrorsCapped(t *testing.T) {
	e := &ReportExporter{}
	resp := types.NewProducerContractResponse()
	for i := 0; i < 20; i++ {
		resp.Add("fail-"+string(rune('A'+i)), nil, assert.AnError)
	}

	jsonBytes, err := e.ExportJSONSummary("group", resp)
	require.NoError(t, err)

	var summary JSONSummary
	err = json.Unmarshal(jsonBytes, &summary)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(summary.TopErrors), 10)
}

func TestTruncateForXML(t *testing.T) {
	assert.Equal(t, "short", truncateForXML("short", 100))
	long := strings.Repeat("x", 600)
	result := truncateForXML(long, 500)
	assert.Equal(t, 503, len(result))
	assert.True(t, strings.HasSuffix(result, "..."))
}
