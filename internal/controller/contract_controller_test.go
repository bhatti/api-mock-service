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
	_ = apiScenarioContractCreateParams{}
	_ = apiScenarioContractResponseBody{}
	_ = postProducerContractHistoryParams{}
	_ = postProducerContractGroupScenarioParams{}
}

func Test_ShouldFailPostContractScenarioWithoutMethod(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewProducerContractController(executor, webServer)

	reader := io.NopCloser(bytes.NewReader([]byte("test")))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario with without method, name and path
	err = ctrl.postProducerContractScenarioByPath(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid method")
}

func Test_ShouldFailPostContractScenarioWithoutName(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewProducerContractController(executor, webServer)

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
	err = ctrl.postProducerContractScenarioByPath(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "scenario name")
}

func Test_ShouldFailPostContractScenarioWithoutPath(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewProducerContractController(executor, webServer)

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
	err = ctrl.postProducerContractScenarioByPath(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "path not specified")
}

func Test_ShouldFailPostContractScenarioWithoutBaseURL(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
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
	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewProducerContractController(executor, webServer)

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
	err = ctrl.postProducerContractScenarioByPath(ctx)

	// THEN it should fail
	require.NoError(t, err)
	res := ctx.Result.(*types.ProducerContractResponse)
	require.Equal(t, 1, len(res.Errors))
	for _, err := range res.Errors {
		require.Contains(t, err, "http URL is not valid ")
	}
}

func Test_ShouldPostContractScenarioWithoutGroup(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND a valid scenario
	_, err = saveTestScenario("../../fixtures/get_todo.yaml", mockScenarioRepository)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	todo := ` { } `
	client.AddMapping("GET", "https://localhost/todos/10", web.NewStubHTTPResponse(200, todo))
	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewProducerContractController(executor, webServer)

	contractReq := types.NewProducerContractRequest("https://localhost", 1, 0)
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
	err = ctrl.postProducerContractGroupScenario(ctx)

	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldPostContractScenarioByHistory(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
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
	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewProducerContractController(executor, webServer)

	contractReq := types.NewProducerContractRequest("https://localhost", 1, 0)
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
	err = ctrl.postProducerContractHistoryByGroup(ctx)

	// THEN it should not fail
	require.NoError(t, err)
	res := ctx.Result.(*types.ProducerContractResponse)
	require.Equal(t, 0, len(res.Errors))
}

func Test_ShouldPostContractScenarioWithGroup(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
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
	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewProducerContractController(executor, webServer)

	contractReq := types.NewProducerContractRequest("https://localhost", 1, 0)
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
	err = ctrl.postProducerContractGroupScenario(ctx)

	// THEN it should not fail
	require.NoError(t, err)
	res := ctx.Result.(*types.ProducerContractResponse)
	require.Equal(t, 0, len(res.Errors))
}

func Test_ShouldPostContractScenarioWithMethodNamePath(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
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
	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewProducerContractController(executor, webServer)

	contractReq := types.NewProducerContractRequest("https://localhost", 1, 0)
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
			"X-Api-Key": {fuzz.RandRegex(`[\x20-\x7F]{1,32}`)},
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
	err = ctrl.postProducerContractScenarioByPath(ctx)

	// THEN it should not fail
	require.NoError(t, err)
	res := ctx.Result.(*types.ProducerContractResponse)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 0, len(res.Errors), fmt.Sprintf("errors %v", res.Errors))
}

func saveTestScenario(name string, repo repository.APIScenarioRepository) (*types.APIScenario, error) {
	// GIVEN a mock scenario loaded from YAML
	b, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	scenario := types.APIScenario{}
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
