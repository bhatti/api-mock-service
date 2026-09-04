package controller

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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

func newReportTestController(t *testing.T) (*ProducerContractController, *contract.ProducerExecutor) {
	config := types.BuildTestConfig()
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	executor := contract.NewProducerExecutor(mockScenarioRepository, groupConfigRepository, client)
	webServer := web.NewStubWebServer()
	ctrl := NewProducerContractController(executor, webServer)
	return ctrl, executor
}

func Test_ShouldReturnJUnitXML_WithReport(t *testing.T) {
	ctrl, executor := newReportTestController(t)
	report := &types.ProducerContractResponse{
		Results:   map[string]any{"POST /users": ""},
		Errors:    map[string]string{"DELETE /users": "not found"},
		Succeeded: 1,
		Failed:    1,
	}
	executor.SetLastReport("test-group", report)

	ctx := web.NewStubContext(&http.Request{})
	ctx.Params["group"] = "test-group"

	err := ctrl.getReportJUnit(ctx)
	require.NoError(t, err)

	raw, ok := ctx.Result.([]byte)
	require.True(t, ok)
	require.True(t, strings.Contains(string(raw), "<?xml"))
	require.True(t, strings.Contains(string(raw), "test-group"))

	var testsuites struct {
		XMLName xml.Name `xml:"testsuites"`
	}
	require.NoError(t, xml.Unmarshal(raw, &testsuites))
}

func Test_ShouldReturnJUnitXML_NoReport(t *testing.T) {
	ctrl, _ := newReportTestController(t)

	ctx := web.NewStubContext(&http.Request{})
	ctx.Params["group"] = "missing-group"

	err := ctrl.getReportJUnit(ctx)
	require.NoError(t, err)

	m, ok := ctx.Result.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "missing-group", m["group"])
	require.Contains(t, m["message"], "no report data")
}

func Test_ShouldReturnJUnitXML_EmptyGroup(t *testing.T) {
	ctrl, _ := newReportTestController(t)
	ctx := web.NewStubContext(&http.Request{})
	err := ctrl.getReportJUnit(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "group not specified")
}

func Test_ShouldReturnJSONSummary_WithReport(t *testing.T) {
	ctrl, executor := newReportTestController(t)
	report := &types.ProducerContractResponse{
		Results:   map[string]any{"GET /items": ""},
		Errors:    map[string]string{},
		Succeeded: 1,
		Failed:    0,
		SecuritySummary: &types.SecuritySummary{
			TotalFindings: 3,
			BySeverity:    map[string]int{"high": 2, "medium": 1},
			ByCategory:    map[string]int{"CWE-89": 2, "CWE-79": 1},
			PassedChecks:  []string{"ssti", "xxe"},
		},
	}
	executor.SetLastReport("summary-group", report)

	ctx := web.NewStubContext(&http.Request{})
	ctx.Params["group"] = "summary-group"

	err := ctrl.getReportSummary(ctx)
	require.NoError(t, err)

	raw, ok := ctx.Result.([]byte)
	require.True(t, ok)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(raw, &parsed))
	require.Equal(t, "summary-group", parsed["group"])
	require.Equal(t, float64(1), parsed["succeeded"])
	require.Equal(t, float64(0), parsed["failed"])
	require.NotNil(t, parsed["securitySummary"])
}

func Test_ShouldReturnJSONSummary_NoReport(t *testing.T) {
	ctrl, _ := newReportTestController(t)

	ctx := web.NewStubContext(&http.Request{})
	ctx.Params["group"] = "empty-group"

	err := ctrl.getReportSummary(ctx)
	require.NoError(t, err)

	m, ok := ctx.Result.(map[string]any)
	require.True(t, ok)
	require.Contains(t, m["message"], "no report data")
}

func Test_ShouldReturnJSONSummary_EmptyGroup(t *testing.T) {
	ctrl, _ := newReportTestController(t)
	ctx := web.NewStubContext(&http.Request{})
	err := ctrl.getReportSummary(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "group not specified")
}
