package contract

import (
	"context"
	"net/http"
	"testing"

	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/stretchr/testify/require"
)

// Test_ExecuteMutationsByGroup_GeneratesAndExecutesMutations verifies that the executor
// generates mutation variants from each scenario in the group and executes them.
func Test_ExecuteMutationsByGroup_GeneratesAndExecutesMutations(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepo, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)

	scenario := &types.APIScenario{
		Method: types.Post,
		Name:   "mutation-exec-scenario",
		Path:   "/mutation-exec",
		Group:  "mutation-exec-group",
		Request: types.APIRequest{
			Contents: `{"name":"Alice","email":"alice@example.com","age":30}`,
			Headers:  map[string]string{"Content-Type": "application/json"},
		},
		Response: types.APIResponse{
			StatusCode: 400,
			Contents:   `{"error":"bad request"}`,
		},
	}
	require.NoError(t, scenarioRepo.Save(scenario))

	client := web.NewStubHTTPClient()
	client.AddMapping("POST", "https://mutationapi.local/mutation-exec",
		web.NewStubHTTPResponse(400, `{"error":"bad request"}`))

	executor := NewProducerExecutor(scenarioRepo, groupConfigRepo, client)
	contractReq := types.NewProducerContractRequest("https://mutationapi.local", 1, 0)
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)

	res := executor.ExecuteMutationsByGroup(context.Background(), &http.Request{}, "mutation-exec-group", dataTemplate, contractReq)
	require.NotNil(t, res)
	total := res.Succeeded + res.Failed + res.Mismatched
	require.Greater(t, total, 0, "expected mutations to be generated and executed")
}

// Test_ExecuteMutationsByGroup_HandlesEmptyGroup verifies graceful handling when
// the group has no scenarios.
func Test_ExecuteMutationsByGroup_HandlesEmptyGroup(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepo, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	executor := NewProducerExecutor(scenarioRepo, groupConfigRepo, client)
	contractReq := types.NewProducerContractRequest("https://mutationapi.local", 1, 0)
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)

	res := executor.ExecuteMutationsByGroup(context.Background(), &http.Request{}, "nonexistent-mutation-group-xyz", dataTemplate, contractReq)
	require.NotNil(t, res)
	require.Equal(t, 0, res.Succeeded+res.Failed+res.Mismatched,
		"empty group should produce no results")
}

// Test_ExecuteMutationsByGroup_ReturnsNonNilMetrics verifies that Metrics is always
// populated (not nil) after mutation execution.
func Test_ExecuteMutationsByGroup_ReturnsNonNilMetrics(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepo, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	executor := NewProducerExecutor(scenarioRepo, groupConfigRepo, client)
	contractReq := types.NewProducerContractRequest("https://metricsapi.local", 1, 0)
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)

	res := executor.ExecuteMutationsByGroup(context.Background(), &http.Request{}, "nonexistent-metrics-group-xyz", dataTemplate, contractReq)
	require.NotNil(t, res)
	require.NotNil(t, res.Metrics)
}

// Test_Execute_InjectsBodyFieldsIntoTemplateParams verifies that JSON request body
// fields are injected into template params so {{.fieldName}} resolves in response
// templates during producer contract execution.
func Test_Execute_InjectsBodyFieldsIntoTemplateParams(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepo, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)

	scenario := &types.APIScenario{
		Method: types.Post,
		Name:   "body-inject-scenario",
		Path:   "/inject",
		Group:  "inject-group",
		Request: types.APIRequest{
			Contents: `{"customerId":"cust-42","amount":100}`,
			Headers:  map[string]string{"Content-Type": "application/json"},
		},
		Response: types.APIResponse{
			StatusCode: 201,
			Contents:   `{"created":true}`,
		},
	}
	require.NoError(t, scenarioRepo.Save(scenario))

	client := web.NewStubHTTPClient()
	client.AddMapping("POST", "https://injectapi.local/inject",
		web.NewStubHTTPResponse(201, `{"created":true}`))

	executor := NewProducerExecutor(scenarioRepo, groupConfigRepo, client)
	contractReq := types.NewProducerContractRequest("https://injectapi.local", 1, 0)
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)

	res := executor.Execute(context.Background(), &http.Request{}, scenario.ToKeyData(), dataTemplate, contractReq)
	for k, v := range res.Errors {
		t.Logf("error %s: %s", k, v)
	}
	require.Equal(t, 0, len(res.Errors), "body field injection should not break scenario execution")
}

// Test_WithOpenAPISpec_ReturnsDistinctCopies verifies that each call to WithOpenAPISpec
// returns a new executor copy, so concurrent HTTP handlers each get an independent instance.
func Test_WithOpenAPISpec_ReturnsDistinctCopies(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepo, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	client := web.NewStubHTTPClient()

	original := NewProducerExecutor(scenarioRepo, groupConfigRepo, client)

	// Pass nil doc/router — WithOpenAPISpec should still return a new copy
	copy1 := original.WithOpenAPISpec(nil, nil)
	copy2 := original.WithOpenAPISpec(nil, nil)

	require.NotSame(t, original, copy1, "WithOpenAPISpec must return a new instance")
	require.NotSame(t, original, copy2, "each call should return a distinct instance")
	require.NotSame(t, copy1, copy2, "two calls must produce two distinct instances")

	// Original's openAPIDoc field must remain nil (not mutated)
	require.Nil(t, original.openAPIDoc, "original executor must not be modified")
}
