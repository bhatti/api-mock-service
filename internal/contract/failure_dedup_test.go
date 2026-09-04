package contract

import (
	"fmt"
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeduplicateFailures_EmptyInput(t *testing.T) {
	clusters := DeduplicateFailures(nil, nil)
	assert.Empty(t, clusters)

	clusters = DeduplicateFailures(map[string]string{}, nil)
	assert.Empty(t, clusters)
}

func TestDeduplicateFailures_AllUnique(t *testing.T) {
	errors := map[string]string{
		"mutation-1": "status 500 didn't match expected value 200",
		"mutation-2": "assertion failed: field 'name' expected string",
		"mutation-3": "completely different error",
	}
	details := map[string]*types.ContractValidationDetail{
		"mutation-1": {URL: "POST /users", StatusCode: 500, ExpectedStatusCode: 200},
		"mutation-2": {URL: "GET /users/1", StatusCode: 200},
		"mutation-3": {URL: "DELETE /users/1", StatusCode: 404},
	}

	clusters := DeduplicateFailures(errors, details)
	assert.Equal(t, 3, len(clusters))
}

func TestDeduplicateFailures_GroupsSameErrorPattern(t *testing.T) {
	errors := make(map[string]string)
	details := make(map[string]*types.ContractValidationDetail)

	for i := 0; i < 50; i++ {
		name := fmt.Sprintf("mutation-%d", i)
		if i < 30 {
			errors[name] = "status 500 didn't match expected value 200"
			details[name] = &types.ContractValidationDetail{URL: "POST /users", StatusCode: 500, ExpectedStatusCode: 200}
		} else if i < 45 {
			errors[name] = "status 400 didn't match expected value 200"
			details[name] = &types.ContractValidationDetail{URL: "POST /users", StatusCode: 400, ExpectedStatusCode: 200}
		} else {
			errors[name] = "assertion failed: field 'name' expected string"
			details[name] = &types.ContractValidationDetail{URL: "GET /users/1", StatusCode: 200}
		}
	}

	clusters := DeduplicateFailures(errors, details)
	require.Equal(t, 3, len(clusters))
	assert.Equal(t, 30, clusters[0].Count)
	assert.Equal(t, 15, clusters[1].Count)
	assert.Equal(t, 5, clusters[2].Count)
}

func TestDeduplicateFailures_SortedByCountDescending(t *testing.T) {
	errors := map[string]string{
		"m1": "error a",
		"m2": "error a",
		"m3": "error b",
	}
	details := map[string]*types.ContractValidationDetail{
		"m1": {URL: "POST /a"},
		"m2": {URL: "POST /a"},
		"m3": {URL: "POST /b"},
	}

	clusters := DeduplicateFailures(errors, details)
	require.True(t, len(clusters) >= 2)
	assert.GreaterOrEqual(t, clusters[0].Count, clusters[len(clusters)-1].Count)
}

func TestDeduplicateFailures_InjectionCategory(t *testing.T) {
	errors := map[string]string{
		"sec-sqli-1": "injection finding(s)",
	}
	details := map[string]*types.ContractValidationDetail{
		"sec-sqli-1": {
			URL:        "POST /login",
			StatusCode: 500,
			InjectionFindings: []types.InjectionFinding{
				{Type: "sqli-error-mysql", Severity: "critical", CWE: "CWE-89"},
			},
		},
	}

	clusters := DeduplicateFailures(errors, details)
	require.Equal(t, 1, len(clusters))
	assert.Equal(t, "injection", clusters[0].Key.ErrorCategory)
	assert.Equal(t, "critical", clusters[0].Severity)
}

func TestDeduplicateFailures_SchemaViolationCategory(t *testing.T) {
	errors := map[string]string{
		"schema-1": "schema violation",
	}
	details := map[string]*types.ContractValidationDetail{
		"schema-1": {
			URL: "GET /api",
			SchemaViolations: []types.SchemaViolation{
				{Field: "age", Message: "expected integer"},
			},
		},
	}

	clusters := DeduplicateFailures(errors, details)
	require.Equal(t, 1, len(clusters))
	assert.Equal(t, "schema-violation", clusters[0].Key.ErrorCategory)
	assert.Equal(t, "medium", clusters[0].Severity)
}

func TestDeduplicateFailures_AffectedMutationsTracked(t *testing.T) {
	errors := map[string]string{
		"m1": "same error",
		"m2": "same error",
		"m3": "same error",
	}
	details := map[string]*types.ContractValidationDetail{
		"m1": {URL: "POST /x", StatusCode: 500},
		"m2": {URL: "POST /x", StatusCode: 500},
		"m3": {URL: "POST /x", StatusCode: 500},
	}

	clusters := DeduplicateFailures(errors, details)
	require.Equal(t, 1, len(clusters))
	assert.Equal(t, 3, len(clusters[0].AffectedMutations))
}

func TestDeduplicateFailures_NoDetails(t *testing.T) {
	errors := map[string]string{
		"m1": "some error",
		"m2": "some error",
	}

	clusters := DeduplicateFailures(errors, nil)
	require.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))
	assert.Equal(t, 2, clusters[0].Count)
}

func TestClassifyError(t *testing.T) {
	assert.Equal(t, "status-mismatch", classifyError("status 500 didn't match expected value 200", nil))
	assert.Equal(t, "assertion", classifyError("response assertion failed", nil))
	assert.Equal(t, "other", classifyError("unknown error", nil))
	assert.Equal(t, "injection", classifyError("any", &types.ContractValidationDetail{
		InjectionFindings: []types.InjectionFinding{{Type: "sqli"}},
	}))
	assert.Equal(t, "schema-violation", classifyError("any", &types.ContractValidationDetail{
		SchemaViolations: []types.SchemaViolation{{Field: "x"}},
	}))
}

func TestHashPrefix(t *testing.T) {
	h1 := hashPrefix("hello world", 512)
	h2 := hashPrefix("hello world", 512)
	h3 := hashPrefix("different", 512)
	assert.Equal(t, h1, h2)
	assert.NotEqual(t, h1, h3)
	assert.Equal(t, 16, len(h1))
}
