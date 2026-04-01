package contract

import (
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// buildMinimalOpenAPIDoc builds a minimal openapi3.T with a single GET /users path.
func buildMinimalOpenAPIDoc(paths map[string][]string) *openapi3.T {
	doc := &openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "Test", Version: "1.0"},
		Paths:   openapi3.Paths{},
	}
	for path, methods := range paths {
		item := &openapi3.PathItem{}
		op := &openapi3.Operation{
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Value: &openapi3.Response{},
				},
			},
		}
		for _, method := range methods {
			switch method {
			case "GET":
				item.Get = op
			case "POST":
				item.Post = op
			case "PUT":
				item.Put = op
			case "DELETE":
				item.Delete = op
			}
		}
		doc.Paths[path] = item
	}
	return doc
}

func Test_Coverage_PopulatedWhenTrackingEnabled(t *testing.T) {
	doc := buildMinimalOpenAPIDoc(map[string][]string{
		"/users": {"GET"},
	})
	scenarios := []*types.APIScenario{
		{Method: "GET", Path: "/users", Response: types.APIResponse{StatusCode: 200}},
	}
	analyzer := NewOpenAPIContractCoverage(doc, scenarios)
	report := analyzer.Analyze()
	require.NotNil(t, report)
	require.Equal(t, 1, report.TotalPaths)
	require.Equal(t, 1, report.CoveredPaths)
	require.InDelta(t, 100.0, report.Coverage, 0.01)
}

func Test_Coverage_UncoveredPathsListed(t *testing.T) {
	doc := buildMinimalOpenAPIDoc(map[string][]string{
		"/users":  {"GET"},
		"/orders": {"POST"},
	})
	// Only cover /users GET, leave /orders POST uncovered
	scenarios := []*types.APIScenario{
		{Method: "GET", Path: "/users", Response: types.APIResponse{StatusCode: 200}},
	}
	analyzer := NewOpenAPIContractCoverage(doc, scenarios)
	report := analyzer.Analyze()
	require.Equal(t, 2, report.TotalPaths)
	require.Equal(t, 1, report.CoveredPaths)
	require.NotEmpty(t, report.UncoveredPaths)
	found := false
	for _, p := range report.UncoveredPaths {
		if p == "POST /orders" {
			found = true
		}
	}
	require.True(t, found, "POST /orders should appear in uncovered paths")
}

func Test_Coverage_MethodCoverageCalculated(t *testing.T) {
	doc := buildMinimalOpenAPIDoc(map[string][]string{
		"/users":  {"GET", "POST"},
		"/orders": {"GET"},
	})
	// Cover all GETs, no POSTs
	scenarios := []*types.APIScenario{
		{Method: "GET", Path: "/users", Response: types.APIResponse{StatusCode: 200}},
		{Method: "GET", Path: "/orders", Response: types.APIResponse{StatusCode: 200}},
	}
	analyzer := NewOpenAPIContractCoverage(doc, scenarios)
	report := analyzer.Analyze()
	require.NotNil(t, report.MethodCoverage)
	// GET should have some positive coverage
	getGov, hasGet := report.MethodCoverage["GET"]
	require.True(t, hasGet)
	require.Greater(t, getGov, 0.0)
}

func Test_Coverage_FieldCoverageByScenario(t *testing.T) {
	// Test that coverageReportToSummary maps _coverage suffix entries correctly
	report := &CoverageReport{
		TotalPaths:   1,
		CoveredPaths: 1,
		Coverage:     100.0,
		PathCoverage: map[string]float64{"GET /users": 100.0},
	}
	results := map[string]any{
		"my-scenario_coverage": &ContractCoverage{
			ScenarioName:    "my-scenario",
			CoveragePercent: 85.0,
		},
		"unrelated-key": "some-value",
	}
	summary := coverageReportToSummary(report, results)
	require.NotNil(t, summary)
	require.Contains(t, summary.FieldCoverageByScenario, "my-scenario")
	require.InDelta(t, 85.0, summary.FieldCoverageByScenario["my-scenario"], 0.01)
	require.NotContains(t, summary.FieldCoverageByScenario, "unrelated-key")
}

func Test_Coverage_NilReportReturnsNil(t *testing.T) {
	require.Nil(t, coverageReportToSummary(nil, nil))
}
