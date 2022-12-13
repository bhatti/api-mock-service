package controller

import (
	"bytes"
	"encoding/json"
	"github.com/bhatti/api-mock-service/internal/chaos"
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

func Test_InitializeSwaggerStructsForMockChaosScenarioController(t *testing.T) {
	_ = mockScenarioChaosCreateParams{}
	_ = mockScenarioChaosResponseBody{}
}

func Test_ShouldFailPostChaosScenarioWithoutMethod(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	executor := chaos.NewExecutor(mockScenarioRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewMockChaosController(executor, webServer)

	reader := io.NopCloser(bytes.NewReader([]byte("test")))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario with without method, name and path
	err = ctrl.PostMockChaosScenario(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid method")
}

func Test_ShouldFailPostChaosScenarioWithoutBaseURL(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
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
	executor := chaos.NewExecutor(mockScenarioRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewMockChaosController(executor, webServer)

	reader := io.NopCloser(bytes.NewReader([]byte("{}")))
	u, err := url.Parse("http://localhost:8080/_chaos/GET/todo-get/todos/10")
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
	err = ctrl.PostMockChaosScenario(ctx)

	// THEN it should fail
	require.Error(t, err)
	require.Contains(t, err.Error(), "baseURL is not specified")
}

func Test_ShouldPostChaosScenarioWithoutGroup(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND a valid scenario
	_, err = saveTestScenario("../../fixtures/get_todo.yaml", mockScenarioRepository)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	todo := ` { } `
	client.AddMapping("GET", "https://localhost/todos/10", web.NewStubHTTPResponse(200, todo))
	executor := chaos.NewExecutor(mockScenarioRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewMockChaosController(executor, webServer)

	chaosReq := types.NewChaosRequest("https://localhost", 1)
	data, err := json.Marshal(chaosReq)
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	u, err := url.Parse("http://localhost:8080/_chaos/GET/todo-get/todos/10")
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
	err = ctrl.PostMockChaosGroupScenario(ctx)

	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldPostChaosScenarioWithGroup(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
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
	executor := chaos.NewExecutor(mockScenarioRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewMockChaosController(executor, webServer)

	chaosReq := types.NewChaosRequest("https://localhost", 1)
	data, err := json.Marshal(chaosReq)
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	u, err := url.Parse("http://localhost:8080/_chaos/GET/todo-get/todos/10")
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
	err = ctrl.PostMockChaosGroupScenario(ctx)

	// THEN it should not fail
	require.NoError(t, err)
	res := ctx.Result.(*types.ChaosResponse)
	require.Equal(t, 0, len(res.Errors))
}

func Test_ShouldPostChaosScenarioWithMethodNamePath(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
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
	executor := chaos.NewExecutor(mockScenarioRepository, client)

	webServer := web.NewStubWebServer()
	ctrl := NewMockChaosController(executor, webServer)

	chaosReq := types.NewChaosRequest("https://localhost", 1)
	data, err := json.Marshal(chaosReq)
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	u, err := url.Parse("http://localhost:8080/_chaos/GET/todo-get/todos/10")
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

	// WHEN creating mock scenario with method, name and path
	err = ctrl.PostMockChaosScenario(ctx)

	// THEN it should not fail
	require.NoError(t, err)
	res := ctx.Result.(*types.ChaosResponse)
	require.Equal(t, 0, len(res.Errors))
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
