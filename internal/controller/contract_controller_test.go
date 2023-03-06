package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"

	"github.com/stretchr/testify/require"
)

func Test_InitializeSwaggerStructsForMockContractScenarioController(t *testing.T) {
	_ = mockScenarioContractCreateParams{}
	_ = mockScenarioContractResponseBody{}
}

func Test_ShouldFailPostContractScenarioWithoutMethod(t *testing.T) {
	config := buildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(config)
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	executor := contract.NewProducerExecutor(mockScenarioRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewContractController(executor, webServer)

	reader := io.NopCloser(bytes.NewReader([]byte("test")))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario with without method, name and path
	err = ctrl.PostMockContractScenario(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid method")
}

func Test_ShouldFailPostContractScenarioWithoutName(t *testing.T) {
	config := buildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(config)
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	executor := contract.NewProducerExecutor(mockScenarioRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewContractController(executor, webServer)

	reader := io.NopCloser(bytes.NewReader([]byte("test")))
	ctx := web.NewStubContext(&http.Request{
		Body:   reader,
		Method: "POST",
		Header: map[string][]string{
			"Mock-Url":  {"https://jsonplaceholder.typicode.com/todos/10"},
			"x-api-key": {fuzz.RandRegex(`[\x20-\x7F]{1,32}`)},
		},
	})
	ctx.Params["method"] = "POST"

	// WHEN creating mock scenario with without method
	err = ctrl.PostMockContractScenario(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "scenario name")
}

func Test_ShouldFailPostContractScenarioWithoutPath(t *testing.T) {
	config := buildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(config)
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	executor := contract.NewProducerExecutor(mockScenarioRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewContractController(executor, webServer)

	reader := io.NopCloser(bytes.NewReader([]byte("test")))
	ctx := web.NewStubContext(&http.Request{
		Body:   reader,
		Method: "POST",
		Header: map[string][]string{
			"Mock-Url":  {"https://jsonplaceholder.typicode.com/todos/10"},
			"x-api-key": {fuzz.RandRegex(`[\x20-\x7F]{1,32}`)},
		},
	})
	ctx.Params["method"] = "POST"
	ctx.Params["name"] = "name"

	// WHEN creating mock scenario with without method
	err = ctrl.PostMockContractScenario(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "path not specified")
}

func Test_ShouldFailPostContractScenarioWithoutBaseURL(t *testing.T) {
	config := buildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(config)
	require.NoError(t, err)
	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/get_todo.yaml", mockScenarioRepository)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	todo := `
{
  "userId": 1,
  "id": 10,
  "title": "illo est ratione doloremque quia maiores aut",
  "completed": true
}
`
	client.AddMapping("GET", "https://localhost/todos/10", web.NewStubHTTPResponse(200, todo))
	executor := contract.NewProducerExecutor(mockScenarioRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewContractController(executor, webServer)

	reader := io.NopCloser(bytes.NewReader([]byte("{}")))
	u, err := url.Parse("http://localhost:8080/_contracts/GET/todo-get/todos/10")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{
		Body:   reader,
		Method: "POST",
		URL:    u,
		Header: map[string][]string{
			"Mock-Url":  {"https://jsonplaceholder.typicode.com/todos/10"},
			"x-api-key": {fuzz.RandRegex(`[\x20-\x7F]{1,32}`)},
		},
	})
	ctx.Params["method"] = string(scenario.Method)
	ctx.Params["name"] = scenario.Name
	ctx.Params["path"] = "/todos/10"

	// WHEN creating mock scenario with without method, name and path
	err = ctrl.PostMockContractScenario(ctx)

	// THEN it should fail
	require.Error(t, err)
	require.Contains(t, err.Error(), "baseURL is not specified")
}

func Test_ShouldPostContractScenarioWithoutGroup(t *testing.T) {
	config := buildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(config)
	require.NoError(t, err)
	// AND a valid scenario
	_, err = saveTestScenario("../../fixtures/get_todo.yaml", mockScenarioRepository)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	todo := ` { } `
	client.AddMapping("GET", "https://localhost/todos/10", web.NewStubHTTPResponse(200, todo))
	executor := contract.NewProducerExecutor(mockScenarioRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewContractController(executor, webServer)

	contractReq := types.NewContractRequest("https://localhost", 1)
	data, err := json.Marshal(contractReq)
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	u, err := url.Parse("http://localhost:8080/_contracts/GET/todo-get/todos/10")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{
		Body:   reader,
		Method: "POST",
		URL:    u,
		Header: map[string][]string{
			"Mock-Url":  {"https://jsonplaceholder.typicode.com/todos/10"},
			"x-api-key": {fuzz.RandRegex(`[\x20-\x7F]{1,32}`)},
		},
	})
	ctx.Params["group"] = ""

	// WHEN creating mock scenario without group
	err = ctrl.PostMockContractGroupScenario(ctx)

	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldPostContractScenarioWithGroup(t *testing.T) {
	config := buildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(config)
	require.NoError(t, err)
	// AND a valid scenario
	_, err = saveTestScenario("../../fixtures/get_todo.yaml", mockScenarioRepository)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	todo := `
{
  "userId": 1,
  "id": 10,
  "title": "illo est ratione doloremque quia maiores aut",
  "completed": true
}
`
	client.AddMapping("GET", "https://localhost/todos/10", web.NewStubHTTPResponse(200, todo))
	executor := contract.NewProducerExecutor(mockScenarioRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewContractController(executor, webServer)

	contractReq := types.NewContractRequest("https://localhost", 1)
	data, err := json.Marshal(contractReq)
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	u, err := url.Parse("http://localhost:8080/_contracts/GET/todo-get/todos/10")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{
		Body:   reader,
		Method: "POST",
		URL:    u,
		Header: map[string][]string{
			"Mock-Url":  {"https://jsonplaceholder.typicode.com/todos/10"},
			"x-api-key": {fuzz.RandRegex(`[\x20-\x7F]{1,32}`)},
		},
	})
	ctx.Params["group"] = "/todos/10"

	// WHEN creating mock scenario with group
	err = ctrl.PostMockContractGroupScenario(ctx)

	// THEN it should not fail
	require.NoError(t, err)
	res := ctx.Result.(*types.ContractResponse)
	require.Equal(t, 0, len(res.Errors))
}

func Test_ShouldPostContractScenarioWithMethodNamePath(t *testing.T) {
	config := buildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(config)
	require.NoError(t, err)
	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/get_todo.yaml", mockScenarioRepository)
	require.NoError(t, err)
	scenario.Path = "/todos/10"
	err = mockScenarioRepository.Save(scenario)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	todo := `
{
  "userId": 15,
  "id": 10,
  "title": "illo est ratione doloremque quia maiores aut",
  "completed": true
}
`
	client.AddMapping("GET", "https://localhost/todos/10", web.NewStubHTTPResponse(200, todo))
	executor := contract.NewProducerExecutor(mockScenarioRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewContractController(executor, webServer)

	contractReq := types.NewContractRequest("https://localhost", 1)
	data, err := json.Marshal(contractReq)
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	u, err := url.Parse("http://localhost:8080/_contracts/GET/todo-get/todos/10")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{
		Body:   reader,
		Method: "POST",
		URL:    u,
		Header: map[string][]string{
			"Mock-Url":  {"https://jsonplaceholder.typicode.com/todos/10"},
			"x-api-key": {fuzz.RandRegex(`[\x20-\x7F]{1,32}`)},
		},
		Form: map[string][]string{
			"id": {"10"},
		},
	})
	ctx.Params["method"] = string(scenario.Method)
	ctx.Params["name"] = scenario.Name
	ctx.Params["path"] = "/todos/10"
	ctx.Params["id"] = "10"

	// WHEN creating mock scenario with method, name and path
	err = ctrl.PostMockContractScenario(ctx)

	// THEN it should not fail
	require.NoError(t, err)
	res := ctx.Result.(*types.ContractResponse)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 0, len(res.Errors), fmt.Sprintf("errors %v", res.Errors))
}

func saveTestScenario(name string, repo repository.MockScenarioRepository) (*types.MockScenario, error) {
	// GIVEN a mock scenario loaded from YAML
	b, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	scenario := types.MockScenario{}
	// AND valid template for random data
	err = yaml.Unmarshal(b, &scenario)
	if err != nil {
		return nil, err
	}
	err = repo.Save(&scenario)
	if err != nil {
		return nil, err
	}
	return &scenario, nil
}
