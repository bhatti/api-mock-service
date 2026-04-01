package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/stretchr/testify/require"
)

// ─── Mutations endpoint ────────────────────────────────────────────────────────

func Test_ShouldFailPostContractMutationsWithoutGroup(t *testing.T) {
	config := types.BuildTestConfig()
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, web.NewStubHTTPClient())
	ctrl := NewProducerContractController(executor, web.NewStubWebServer())

	contractReq := types.NewProducerContractRequest("https://localhost", 1, 0)
	data, err := json.Marshal(contractReq)
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{Body: io.NopCloser(bytes.NewReader(data))})
	ctx.Params["group"] = ""

	err = ctrl.postProducerContractMutationsByGroup(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "group not specified")
}

func Test_ShouldPostContractMutationsByGroupWithNoScenarios(t *testing.T) {
	config := types.BuildTestConfig()
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, web.NewStubHTTPClient())
	ctrl := NewProducerContractController(executor, web.NewStubWebServer())

	contractReq := types.NewProducerContractRequest("https://localhost", 1, 0)
	data, err := json.Marshal(contractReq)
	require.NoError(t, err)
	u, _ := url.Parse("http://localhost:8080/_contracts/mutations/no-such-group")
	ctx := web.NewStubContext(&http.Request{Body: io.NopCloser(bytes.NewReader(data)), URL: u})
	ctx.Params["group"] = "no-such-group-xyz"

	err = ctrl.postProducerContractMutationsByGroup(ctx)
	require.NoError(t, err)
	res := ctx.Result.(*types.ProducerContractResponse)
	// No scenarios → no mutations to execute
	require.Equal(t, 0, res.Succeeded+res.Failed+res.Mismatched)
}

func Test_ShouldPostContractMutationsByGroupWithScenario(t *testing.T) {
	config := types.BuildTestConfig()
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)

	// Save a scenario with a JSON body so mutations can be generated
	scenario := &types.APIScenario{
		Method: types.Post,
		Name:   "ctrl-mutation-scenario",
		Path:   "/ctrl-mutation",
		Group:  "ctrl-mutation-group",
		Request: types.APIRequest{
			Contents: `{"username":"bob","age":25}`,
			Headers:  map[string]string{"Content-Type": "application/json"},
		},
		Response: types.APIResponse{StatusCode: 400, Contents: `{"error":"bad"}`},
	}
	require.NoError(t, mockScenarioRepository.Save(scenario))

	client := web.NewStubHTTPClient()
	client.AddMapping("POST", "https://localhost/ctrl-mutation",
		web.NewStubHTTPResponse(400, `{"error":"bad"}`))

	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, client)
	ctrl := NewProducerContractController(executor, web.NewStubWebServer())

	contractReq := types.NewProducerContractRequest("https://localhost", 1, 0)
	data, err := json.Marshal(contractReq)
	require.NoError(t, err)
	u, _ := url.Parse("http://localhost:8080/_contracts/mutations/ctrl-mutation-group")
	ctx := web.NewStubContext(&http.Request{Body: io.NopCloser(bytes.NewReader(data)), URL: u})
	ctx.Params["group"] = "ctrl-mutation-group"

	err = ctrl.postProducerContractMutationsByGroup(ctx)
	require.NoError(t, err)
	res := ctx.Result.(*types.ProducerContractResponse)
	total := res.Succeeded + res.Failed + res.Mismatched
	require.Greater(t, total, 0, "mutations should have been executed")
}

// ─── spec_content support ─────────────────────────────────────────────────────

const minimalSpecContent = `
openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths:
  /todos/:id:
    get:
      operationId: getTodo
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                  title:
                    type: string
`

func Test_ShouldPostContractGroupScenarioWithSpecContent(t *testing.T) {
	config := types.BuildTestConfig()
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	_, err = saveTestScenario("../../fixtures/get_todo.yaml", mockScenarioRepository)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	client.AddMapping("GET", "https://localhost/todos/10",
		web.NewStubHTTPResponse(200, `{"userId":1,"id":10,"title":"illo test","completed":true}`))

	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, client)
	ctrl := NewProducerContractController(executor, web.NewStubWebServer())

	contractReq := types.NewProducerContractRequest("https://localhost", 1, 0)
	contractReq.SpecContent = minimalSpecContent
	data, err := json.Marshal(contractReq)
	require.NoError(t, err)
	u, _ := url.Parse("http://localhost:8080/_contracts/todos")
	ctx := web.NewStubContext(&http.Request{
		Body: io.NopCloser(bytes.NewReader(data)),
		URL:  u,
		Header: map[string][]string{
			"Mock-Url":  {"https://jsonplaceholder.typicode.com/todos/10"},
			"x-api-key": {fuzz.RandRegex(`[\x20-\x7F]{1,32}`)},
		},
	})
	ctx.Params["group"] = "todos"

	// Should succeed without error — specAwareExecutor parses spec and attaches it
	err = ctrl.postProducerContractGroupScenario(ctx)
	require.NoError(t, err)
	res := ctx.Result.(*types.ProducerContractResponse)
	require.NotNil(t, res)
}

func Test_ShouldPostContractGroupScenarioWithInvalidSpecContent(t *testing.T) {
	config := types.BuildTestConfig()
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	_, err = saveTestScenario("../../fixtures/get_todo.yaml", mockScenarioRepository)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	client.AddMapping("GET", "https://localhost/todos/10",
		web.NewStubHTTPResponse(200, `{"userId":1,"id":10,"title":"illo test","completed":true}`))

	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, client)
	ctrl := NewProducerContractController(executor, web.NewStubWebServer())

	contractReq := types.NewProducerContractRequest("https://localhost", 1, 0)
	contractReq.SpecContent = "this is not valid yaml or json {"
	data, err := json.Marshal(contractReq)
	require.NoError(t, err)
	u, _ := url.Parse("http://localhost:8080/_contracts/todos")
	ctx := web.NewStubContext(&http.Request{
		Body: io.NopCloser(bytes.NewReader(data)),
		URL:  u,
		Header: map[string][]string{
			"Mock-Url":  {"https://jsonplaceholder.typicode.com/todos/10"},
			"x-api-key": {fuzz.RandRegex(`[\x20-\x7F]{1,32}`)},
		},
	})
	ctx.Params["group"] = "todos"

	// Invalid spec should fall back gracefully — execution continues without schema validation
	err = ctrl.postProducerContractGroupScenario(ctx)
	require.NoError(t, err, "invalid spec_content should be ignored gracefully")
}

// ─── specAwareExecutor helper ─────────────────────────────────────────────────

func Test_specAwareExecutor_ReturnsBaseWhenNoSpec(t *testing.T) {
	config := types.BuildTestConfig()
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	base := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, web.NewStubHTTPClient())

	contractReq := &types.ProducerContractRequest{SpecContent: ""}
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)

	result := specAwareExecutor(base, contractReq, dataTemplate)
	require.Same(t, base, result, "when SpecContent is empty, specAwareExecutor should return the original executor")
}

func Test_specAwareExecutor_ReturnsNewCopyWhenSpecProvided(t *testing.T) {
	config := types.BuildTestConfig()
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	base := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, web.NewStubHTTPClient())

	contractReq := &types.ProducerContractRequest{SpecContent: minimalSpecContent}
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)

	result := specAwareExecutor(base, contractReq, dataTemplate)
	require.NotSame(t, base, result, "when SpecContent is provided, specAwareExecutor should return a new copy")
}

// ─── ProducerContractRequest new fields ───────────────────────────────────────

func Test_ProducerContractRequest_UnmarshalWithSpecContent(t *testing.T) {
	raw := `{"base_url":"https://api.example.com","execution_times":3,"spec_content":"openapi: 3.0.3"}`
	var req types.ProducerContractRequest
	require.NoError(t, json.Unmarshal([]byte(raw), &req))
	require.Equal(t, "https://api.example.com", req.BaseURL)
	require.Equal(t, 3, req.ExecutionTimes)
	require.Equal(t, "openapi: 3.0.3", req.SpecContent)
}

func Test_ProducerContractRequest_UnmarshalWithRunMutations(t *testing.T) {
	raw := `{"base_url":"https://api.example.com","run_mutations":true}`
	var req types.ProducerContractRequest
	require.NoError(t, json.Unmarshal([]byte(raw), &req))
	require.True(t, req.RunMutations)
}

func Test_ProducerContractRequest_MarshalPreservesNewFields(t *testing.T) {
	req := &types.ProducerContractRequest{
		BaseURL:       "https://api.example.com",
		SpecContent:   "openapi: 3.0.3",
		RunMutations:  true,
		TrackCoverage: true,
	}
	data, err := json.Marshal(req)
	require.NoError(t, err)
	var decoded types.ProducerContractRequest
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, req.SpecContent, decoded.SpecContent)
	require.Equal(t, req.RunMutations, decoded.RunMutations)
	require.Equal(t, req.TrackCoverage, decoded.TrackCoverage)
}

func Test_ProducerContractRequest_SpecContentOmittedWhenEmpty(t *testing.T) {
	req := &types.ProducerContractRequest{BaseURL: "https://api.example.com"}
	data, err := json.Marshal(req)
	require.NoError(t, err)
	// spec_content should not appear in JSON when empty (omitempty)
	require.NotContains(t, string(data), "spec_content")
}
