package contract

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
)

func baseScenario() *types.APIScenario {
	return &types.APIScenario{
		Name: "test-scenario",
		Request: types.APIRequest{
			Contents: `{"name":"Alice","age":30,"email":"alice@example.com","createdAt":"2024-01-01","userId":"u1"}`,
		},
		Response: types.APIResponse{
			StatusCode: 200,
		},
	}
}

func Test_NullFieldMutations_OnePerField(t *testing.T) {
	scenario := baseScenario()
	mutator := NewContractMutator(scenario)

	var requestBody map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(scenario.Request.Contents), &requestBody))
	fieldCount := len(requestBody)

	mutations := mutator.GenerateMutations()
	nullMutations := make([]*types.APIScenario, 0)
	for _, m := range mutations {
		if strings.Contains(m.Name, "-null-") {
			nullMutations = append(nullMutations, m)
		}
	}
	require.Equal(t, fieldCount, len(nullMutations), "expected one null mutation per field")
	for _, m := range nullMutations {
		require.Equal(t, 422, m.Response.StatusCode, "null mutations should expect 422")
	}
}

func Test_CombinatorialMutations_MaxTen(t *testing.T) {
	scenario := baseScenario()
	mutator := NewContractMutator(scenario)

	mutations := mutator.GenerateMutations()
	comboMutations := make([]*types.APIScenario, 0)
	for _, m := range mutations {
		if strings.Contains(m.Name, "-combo-") {
			comboMutations = append(comboMutations, m)
		}
	}
	require.LessOrEqual(t, len(comboMutations), 10, "combinatorial mutations should be capped at 10")
	require.Greater(t, len(comboMutations), 0, "should have at least one combinatorial mutation")
}

func Test_FormatBoundary_DateFields(t *testing.T) {
	scenario := &types.APIScenario{
		Name: "test-scenario",
		Request: types.APIRequest{
			// "startDate" contains "date" → triggers date format mutations
			Contents: `{"startDate":"2024-01-01"}`,
		},
		Response: types.APIResponse{StatusCode: 200},
	}
	mutator := NewContractMutator(scenario)
	mutations := mutator.GenerateMutations()

	dateFormatMutations := make([]*types.APIScenario, 0)
	for _, m := range mutations {
		if strings.Contains(m.Name, "-format-date-") {
			dateFormatMutations = append(dateFormatMutations, m)
		}
	}
	require.NotEmpty(t, dateFormatMutations, "should generate date format mutations")
	for _, m := range dateFormatMutations {
		require.Equal(t, 400, m.Response.StatusCode)
	}
}

func Test_FormatBoundary_UUIDFields(t *testing.T) {
	scenario := &types.APIScenario{
		Name: "test-scenario",
		Request: types.APIRequest{
			Contents: `{"userId":"abc-123-def"}`,
		},
		Response: types.APIResponse{StatusCode: 200},
	}
	mutator := NewContractMutator(scenario)
	mutations := mutator.GenerateMutations()

	// "userId" ends with "id" suffix so it should trigger uuid format mutations
	uuidMutations := make([]*types.APIScenario, 0)
	for _, m := range mutations {
		if strings.Contains(m.Name, "-format-uuid-") {
			uuidMutations = append(uuidMutations, m)
		}
	}
	require.NotEmpty(t, uuidMutations, "should generate uuid format mutations for id-suffixed fields")
}

func Test_SecurityInjection_ContainsSQLAndPathTraversal(t *testing.T) {
	scenario := &types.APIScenario{
		Name: "test-scenario",
		Request: types.APIRequest{
			Contents: `{"username":"alice","path":"/home"}`,
		},
		Response: types.APIResponse{StatusCode: 200},
	}
	mutator := NewContractMutator(scenario)
	mutations := mutator.GenerateMutations()

	secMutations := make([]*types.APIScenario, 0)
	for _, m := range mutations {
		if strings.Contains(m.Name, "-sec-") {
			secMutations = append(secMutations, m)
		}
	}
	require.NotEmpty(t, secMutations, "should generate security injection mutations")

	hasSQLi := false
	hasPathTraversal := false
	for _, m := range secMutations {
		if strings.Contains(m.Name, "sqli") {
			hasSQLi = true
		}
		if strings.Contains(m.Name, "path-traversal") {
			hasPathTraversal = true
		}
	}
	require.True(t, hasSQLi, "should include SQL injection mutations")
	require.True(t, hasPathTraversal, "should include path traversal mutations")
}

func Test_BoundaryMutations_BothMinAndMax(t *testing.T) {
	scenario := baseScenario()
	mutator := NewContractMutator(scenario)
	mutations := mutator.GenerateMutations()

	hasMin := false
	hasMax := false
	for _, m := range mutations {
		if strings.Contains(m.Name, "-boundary-min") {
			hasMin = true
		}
		if strings.Contains(m.Name, "-boundary-max") {
			hasMax = true
		}
	}
	require.True(t, hasMin, "should generate min boundary mutation")
	require.True(t, hasMax, "should generate max boundary mutation")
}

func Test_GenerateMutations_EmptyBodyProducesNoMutations(t *testing.T) {
	scenario := &types.APIScenario{
		Name:     "empty-body",
		Request:  types.APIRequest{},
		Response: types.APIResponse{StatusCode: 200},
	}
	mutator := NewContractMutator(scenario)
	mutations := mutator.GenerateMutations()
	// missing-fields mutation is skipped when body is empty
	// boundary/null/combinatorial/format/security all skip too
	require.Empty(t, mutations)
}
