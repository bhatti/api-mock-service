package contract

import (
	"context"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/pm"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/oapi"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/stretchr/testify/require"
)

var baseURL = "https://mocksite.local"

func Test_ShouldNotExecuteNonexistentScenario(t *testing.T) {
	// GIVEN scenario repository
	config := types.BuildTestConfig()
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1)
	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, web.NewHTTPClient(config, web.NewAuthAdapter(config)))
	// THEN it should execute saved scenario
	res := executor.Execute(context.Background(), &http.Request{}, &types.APIKeyData{}, dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors[""], `could not lookup matching API`)
}

func Test_ShouldExecuteScenariosByHistory(t *testing.T) {
	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND a valid scenario
	scenario1, err := saveTestScenario("../../fixtures/create_user.yaml", scenarioRepository)
	require.NoError(t, err)
	require.Equal(t, "application/json", scenario1.Request.Headers["Content-Type"])
	require.Equal(t, "application/json", scenario1.Response.Headers["Content-Type"][0])
	require.Equal(t, "keep-alive", scenario1.Response.Headers["Connection"][0])
	_, err = saveTestScenario("../../fixtures/get_user.yaml", scenarioRepository)
	require.NoError(t, err)
	_, err = saveTestScenario("../../fixtures/users.yaml", scenarioRepository)
	require.NoError(t, err)

	// AND valid template for random data
	contractReq := types.NewProducerContractRequest(baseURL, 1)
	client := web.NewStubHTTPClient()
	client.AddMapping("POST", baseURL+"/users", web.NewStubHTTPResponse(200,
		`{"User": {"Directory": "my_dir", "Username": "my_user@foo.cc", "DesiredDeliveryMediums": ["EMAIL"]}}`))
	client.AddMapping("GET", baseURL+"/users/1", web.NewStubHTTPResponse(200,
		`{"User": {"Directory": "my_dir2", "Username": "my_user2@foo.cc", "DesiredDeliveryMediums": ["EMAIL"]}}`))
	client.AddMapping("GET", baseURL+"/users", web.NewStubHTTPResponse(200,
		`{"User": {"Directory": "my_dir3", "Username": "my_user3@foo.cc", "DesiredDeliveryMediums": ["EMAIL"]}}`))
	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, client)
	// THEN it should execute saved scenario
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	res := executor.ExecuteByHistory(context.Background(), &http.Request{}, "user_group", dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 0, len(res.Errors), fmt.Sprintf("%v", res.Errors))
}

func Test_ShouldExecuteChainedGroupScenarios(t *testing.T) {
	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND a valid scenario
	_, err = saveTestScenario("../../fixtures/create_user.yaml", scenarioRepository)
	require.NoError(t, err)
	_, err = saveTestScenario("../../fixtures/get_user.yaml", scenarioRepository)
	require.NoError(t, err)
	_, err = saveTestScenario("../../fixtures/users.yaml", scenarioRepository)
	require.NoError(t, err)

	// AND valid template for random data
	contractReq := types.NewProducerContractRequest(baseURL, 1)
	client := web.NewStubHTTPClient()
	client.AddMapping("POST", baseURL+"/users", web.NewStubHTTPResponse(200,
		`{"User": {"Directory": "my_dir", "Username": "my_user@foo.cc", "DesiredDeliveryMediums": ["EMAIL"]}}`))
	client.AddMapping("GET", baseURL+"/users/1", web.NewStubHTTPResponse(200,
		`{"User": {"Directory": "my_dir2", "Username": "my_user2@foo.cc", "DesiredDeliveryMediums": ["EMAIL"]}}`))
	client.AddMapping("GET", baseURL+"/users", web.NewStubHTTPResponse(200,
		`{"User": {"Directory": "my_dir3", "Username": "my_user3@foo.cc", "DesiredDeliveryMediums": ["EMAIL"]}}`))
	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, client)
	// THEN it should execute saved scenario
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	res := executor.ExecuteByGroup(context.Background(), &http.Request{}, "user_group", dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 0, len(res.Errors), fmt.Sprintf("%v", res.Errors))
}

func Test_ShouldExecuteGetTodo(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/get_todo.yaml", scenarioRepository)
	require.NoError(t, err)
	scenario.Path = "/todos/10"
	scenario.Response.Assertions = []string{"VariableContains contents.id 10", "VariableContains contents.title illo"}
	err = scenarioRepository.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1)
	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, web.NewHTTPClient(config, web.NewAuthAdapter(config)))
	// THEN it should execute saved scenario
	res := executor.Execute(context.Background(), &http.Request{}, scenario.ToKeyData(), dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 0, len(res.Errors), fmt.Sprintf("%v", res.Errors))
}

func Test_ShouldExecutePutPosts(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/put_posts.yaml", scenarioRepository)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1)

	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, web.NewHTTPClient(config, web.NewAuthAdapter(config)))
	// THEN it should execute saved scenario
	res := executor.Execute(context.Background(), &http.Request{}, scenario.ToKeyData(), dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 0, len(res.Errors))
}

func Test_ShouldNotExecutePutPostsWithBadHeaderAssertions(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/put_posts.yaml", scenarioRepository)
	require.NoError(t, err)
	// AND a bad assertion
	scenario.Request.Headers[types.AuthorizationHeader] = "AWS4-HMAC-SHA256"
	scenario.Response.Assertions = types.AddAssertion(
		scenario.Response.Assertions, "VariableContains headers.Content-Type application/xjson")
	err = scenarioRepository.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1)

	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, web.NewHTTPClient(config, web.NewAuthAdapter(config)))
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), &http.Request{}, scenario.ToKeyData(), dataTemplate, contractReq)
	for k, err := range res.Errors {
		t.Log(k, err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["put_posts"], `failed to assert response '{{VariableContains "headers.Content-Type"`)
}

func Test_ShouldNotExecutePutPostsWithBadHeaders(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/put_posts.yaml", scenarioRepository)
	require.NoError(t, err)
	// AND bad matching header
	scenario.Response.AssertHeadersPattern[types.ContentTypeHeader] = "application/xjson"
	err = scenarioRepository.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1)

	// AND executor
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, web.NewHTTPClient(config, web.NewAuthAdapter(config)))

	// WHEN executing scenario
	res := executor.Execute(context.Background(), &http.Request{}, scenario.ToKeyData(), dataTemplate, contractReq)
	// THEN it should not execute saved scenario
	for k, err := range res.Errors {
		t.Log(k, err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["put_posts"], `didn't match required response header Content-Type with regex application/xjson`)
}

func Test_ShouldNotExecutePutPostsWithMissingRequestHeaders(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/put_posts.yaml", scenarioRepository)
	require.NoError(t, err)
	// AND missing header
	scenario.Request.Headers["Content-Type"] = "blah"
	err = scenarioRepository.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1)

	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, web.NewHTTPClient(config, web.NewAuthAdapter(config)))
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), &http.Request{}, scenario.ToKeyData(), dataTemplate, contractReq)
	for k, err := range res.Errors {
		t.Log(k, err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["put_posts"], `didn't match required request header 'Content-Type' with regex 'application/x-www-form-urlencoded'`)
}

func Test_ShouldNotExecutePutPostsWithMissingResponseHeaders(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/put_posts.yaml", scenarioRepository)
	require.NoError(t, err)
	// AND missing header
	scenario.Response.AssertHeadersPattern["Abc-Content-Type"] = "application/xjson"
	err = scenarioRepository.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1)

	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, web.NewHTTPClient(config, web.NewAuthAdapter(config)))
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), &http.Request{}, scenario.ToKeyData(), dataTemplate, contractReq)
	for k, err := range res.Errors {
		t.Log(k, err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["put_posts"], `failed to find required response header Abc-Content-Type with regex application/xjson`)
}

func Test_ShouldExecutePostProductScenario(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/save_product.yaml", scenarioRepository)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	product := `{"category":"BOOKS","id":"123","inventory":"10","name":"toy 1","price":{"amount":12,"currency":"USD"}}`
	client.AddMapping("POST", baseURL+"/products", web.NewStubHTTPResponse(200, product))

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest(baseURL, 1)
	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, client)
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), &http.Request{}, scenario.ToKeyData(), dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 0, len(res.Errors))
}

func Test_ShouldExecuteGetTodoWithBadAssertions(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/get_comment.yaml", scenarioRepository)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	todo := `
        {
          "postId": 1,
          "id": 22,
          "name": "id labore ex et quam laborum",
          "email": "Eliseo@gardner.biz",
          "body": "laudantium enim quasi est quidem magnam voluptate ipsam eos\ntempora quo necessitatibus\ndolor quam autem quasi\nreiciendis et nam sapiente accusantium"
        }
`
	client.AddMapping("GET", baseURL+"/comments/1", web.NewStubHTTPResponse(200, todo))

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest(baseURL, 1)
	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, client)
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), &http.Request{}, scenario.ToKeyData(), dataTemplate, contractReq)
	for k, err := range res.Errors {
		t.Log(k, err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["get_comment"], `failed to assert response '{{VariableContains "contents.id" "1"}}`)
}

func Test_ShouldExecuteGetTodoWithBadStatus(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/get_comment.yaml", scenarioRepository)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	todo := `{} `
	client.AddMapping("GET", baseURL+"/comments/1", web.NewStubHTTPResponse(400, todo))

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest(baseURL, 1)
	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, client)
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), &http.Request{}, scenario.ToKeyData(), dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["get_comment"], `failed to execute request with status 400`)
}

func Test_ShouldExecuteJobsOpenAPIWithInvalidStatus(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND mock scenarios from open-api specifications
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)

	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
	contractReq := types.NewProducerContractRequest(baseURL, 5)
	specs, _, err := oapi.Parse(context.Background(), &types.Configuration{}, b, dataTemplate)

	require.NoError(t, err)

	// AND save specs
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTemplate)
		require.NoError(t, err)
		if scenario.Response.StatusCode == 200 {
			scenario.Group = "bad_v1_job"
		}
		err = scenarioRepository.Save(scenario)
		require.NoError(t, err)
	}
	// AND valid template for random data
	// AND mock web client
	client, data := buildJobsTestClient("AC1234567890", "BAD", "", 200)
	contractReq.Params = data
	contractReq.Verbose = true
	// AND executor
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, client)
	// WHEN executing scenario
	res := executor.ExecuteByGroup(context.Background(), &http.Request{}, "bad_v1_job", dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
		// THEN it should fail to execute
		require.Contains(t, err, "key 'jobStatus' - value 'BAD' didn't match regex")
	}
}

func Test_ShouldExecuteJobsOpenAPI(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN mock scenarios from open-api specifications
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	// AND scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest(baseURL, 5)
	specs, _, err := oapi.Parse(context.Background(), &types.Configuration{}, b, dataTemplate)
	require.NoError(t, err)

	for i, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTemplate)
		scenario.Path = "/good" + scenario.Path
		require.NoError(t, err)
		scenario.Group = fmt.Sprintf("good_spec_%d", i)
		// WHEN saving scenario to mock scenario repository
		err = scenarioRepository.Save(scenario)
		// THEN it should save scenario
		require.NoError(t, err)
		// WITH mock web client
		client, data := buildJobsTestClient("AC1234567890", "RUNNING", "/good", scenario.Response.StatusCode)
		contractReq.Verbose = i%2 == 0
		contractReq.Params = data
		// AND executor
		executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, client)
		// AND should return saved scenario
		saved, err := scenarioRepository.Lookup(scenario.ToKeyData(), nil)
		require.NoError(t, err)
		if scenario.Response.StatusCode != saved.Response.StatusCode {
			t.Fatalf("unexpected status %d != %d", scenario.Response.StatusCode, saved.Response.StatusCode)
		}
		// WHEN executing scenario
		res := executor.Execute(context.Background(), &http.Request{}, saved.ToKeyData(), dataTemplate, contractReq)
		for _, err := range res.Errors {
			t.Log(err)
		}
		// THEN it should succeed
		require.Equal(t, 0, len(res.Errors), fmt.Sprintf("spec %d == %v", i, res.Errors))
	}
}

func Test_ShouldParseRequestBody(t *testing.T) {
	scenario := &types.APIScenario{}
	scenario.Request.AssertContentsPattern = `{"PoolId": "us-west-2", "Username": "christina.perdit@sonaret.gov", "UserAttributes": [{"Name": "email", "Value": "christina.perdit@sonaret.gov"}], "DesiredDeliveryMediums": ["EMAIL"]}`
	str, _ := buildRequestBody(scenario)
	require.Contains(t, str, "christina.perdit@sonaret.gov")
}

func buildJobsTestClient(jobID string, jobStatus string, prefixPath string, httpStatus int) (web.HTTPClient, map[string]any) {
	client := web.NewStubHTTPClient()
	job := `
{
  "add": [
    "id1"
  ],
  "attributeMap": {
    "attr1": true,
    "attr2": "hello"
  },
  "completed": false,
  "jobId": "` + jobID + `",
  "jobStatus": "` + jobStatus + `",
  "name": "test-job",
  "records": 5000,
  "remaining": 2000,
  "remove": [
    "none"
  ]
}
`
	jobStatusReply := `
{
  "jobId": "` + jobID + `",
  "jobStatus": "` + jobStatus + `"
}
`

	client.AddMapping("GET", baseURL+prefixPath+"/v1/jobs/"+jobID, web.NewStubHTTPResponse(httpStatus, job))
	client.AddMapping("GET", baseURL+prefixPath+"/v1/jobs", web.NewStubHTTPResponse(httpStatus, `[`+job+`]`))
	client.AddMapping("POST", baseURL+prefixPath+"/v1/jobs", web.NewStubHTTPResponse(httpStatus, jobStatusReply))
	client.AddMapping("POST", baseURL+prefixPath+"/v1/jobs/"+jobID+"/cancel", web.NewStubHTTPResponse(httpStatus, jobStatusReply))
	client.AddMapping("POST", baseURL+prefixPath+"/v1/jobs/"+jobID+"/pause", web.NewStubHTTPResponse(httpStatus, jobStatusReply))
	client.AddMapping("POST", baseURL+prefixPath+"/v1/jobs/"+jobID+"/resume", web.NewStubHTTPResponse(httpStatus, jobStatusReply))
	client.AddMapping("POST", baseURL+prefixPath+"/v1/jobs/"+jobID+"/state", web.NewStubHTTPResponse(httpStatus, jobStatusReply))
	client.AddMapping("POST", baseURL+prefixPath+"/v1/jobs/"+jobID+"/state/"+jobStatus, web.NewStubHTTPResponse(httpStatus, jobStatusReply))
	return client, map[string]any{"jobId": jobID, "state": jobStatus}
}

func Test_ShouldExecutePostmanAPISuiteAsProducer(t *testing.T) {
	// GIVEN configuration and repositories
	config := types.BuildTestConfig()
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)

	// Define base URL for API server
	baseURL := "https://api.example.com"

	// Load and parse the Postman collection
	file, err := os.Open("../../fixtures/postman_basic.json")
	require.NoError(t, err)
	defer file.Close()

	// Parse collection
	collection, err := pm.ParseCollection(file)
	require.NoError(t, err)
	require.Equal(t, "API Testing Suite", collection.Info.Name)

	// Convert to scenarios and store them
	scenarios, vars := pm.ConvertPostmanToScenarios(config, collection, time.Now(), time.Now())
	require.NotEmpty(t, scenarios)
	require.NotEmpty(t, vars.Variables)

	// Create test variables
	apiKey := "test-api-key-123"
	resourceID := "res-12345"
	jws := "mock-jws-token"
	accessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IlRlc3QgVXNlciIsImlhdCI6MTUxNjIzOTAyMn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	// Set variables
	vars.Variables["api_key"] = apiKey
	vars.Variables["resource_id"] = resourceID
	vars.Variables["jws"] = jws
	vars.Variables["base_url"] = baseURL
	vars.Variables["client_name"] = "test-client"
	vars.Variables["org_id"] = "test-org"

	// Create a stub HTTP client
	client := web.NewStubHTTPClient()

	// Configure stub responses for each API endpoint

	// Auth endpoint
	client.AddMapping("POST", baseURL+"/auth/token", web.NewStubHTTPResponse(200,
		map[string]interface{}{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
		}).WithHeader("Content-Type", "application/json"))

	// Create Resource endpoint
	client.AddMapping("POST", baseURL+"/api/resources", web.NewStubHTTPResponse(200, // 201
		map[string]interface{}{
			"id":          resourceID,
			"name":        "test resource",
			"description": "test description",
			"created_at":  time.Now().Format(time.RFC3339),
		}).WithHeader("Content-Type", "application/json"))

	// Get Resource endpoint
	client.AddMapping("GET", baseURL+"/api/resources/"+resourceID, web.NewStubHTTPResponse(200,
		map[string]interface{}{
			"id":          resourceID,
			"name":        "test resource",
			"description": "test description",
		}).WithHeader("Content-Type", "application/json"))

	// Update Resource endpoint
	client.AddMapping("PATCH", baseURL+"/api/resources/"+resourceID, web.NewStubHTTPResponse(200,
		map[string]interface{}{
			"id":          resourceID,
			"name":        "updated name",
			"description": "test description",
			"updated_at":  time.Now().Format(time.RFC3339),
		}).WithHeader("Content-Type", "application/json"))

	// Delete Resource endpoint
	client.AddMapping("DELETE", baseURL+"/api/resources/"+resourceID,
		web.NewStubHTTPResponse(200, nil, "Content-Type", "application/json")) // 204

	// Save all scenarios
	for _, scenario := range scenarios {
		// Update BaseURL to empty to use the one from contractReq
		scenario.BaseURL = ""

		// Save scenario
		err = scenarioRepository.Save(scenario)
		require.NoError(t, err)
	}

	// Save variables
	err = scenarioRepository.SaveVariables(vars)
	require.NoError(t, err)

	// Create producer executor with stub client
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, client)

	// --- Test 1: Test Individual APIs ---
	t.Run("Individual API Tests", func(t *testing.T) {
		// Find auth scenario
		var authScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Name, "Get JWT Token") {
				authScenario = s
				break
			}
		}
		require.NotNil(t, authScenario, "Auth scenario not found")

		// Test auth API
		t.Run("Auth API Test", func(t *testing.T) {
			// Set up data template and contract request
			dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
			contractReq := types.NewProducerContractRequest(baseURL, 1)
			contractReq.Headers.Set("Content-Type", "application/json")
			contractReq.Headers.Set("x-api-key", apiKey)
			contractReq.Params = map[string]interface{}{
				"jws": jws,
			}

			// Execute the auth scenario
			res := executor.Execute(
				context.Background(),
				&http.Request{},
				authScenario.ToKeyData(),
				dataTemplate,
				contractReq,
			)

			// Verify results
			require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Auth API errors: %v", res.Errors))
			mapResp := getFirstMap(res.Results)
			require.NotEmpty(t, mapResp, "No results returned from auth API")

			// Check for token in response
			require.Contains(t, mapResp, "access_token", "Response missing access_token")
			require.Equal(t, accessToken, mapResp["access_token"], "Unexpected access token")
		})

		// Test CRUD operations
		for _, scenario := range scenarios {
			if !strings.Contains(scenario.Group, "CRUD Operations") {
				continue
			}

			t.Run(fmt.Sprintf("CRUD Test: %s", scenario.Name), func(t *testing.T) {
				// Set up data template and contract request
				dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
				contractReq := types.NewProducerContractRequest(baseURL, 1)
				contractReq.Headers.Set("Content-Type", "application/json")
				contractReq.Headers.Set("x-api-key", apiKey)
				contractReq.Headers.Set("Authorization", "Bearer "+accessToken)
				contractReq.Params = map[string]interface{}{
					"resource_id":  resourceID,
					"name":         "test resource",
					"description":  "test description",
					"access_token": accessToken,
				}

				// If this is the update test, update the name parameter
				if strings.Contains(scenario.Name, "Update") {
					contractReq.Params["name"] = "updated name"
				}

				// Execute the scenario
				res := executor.Execute(
					context.Background(),
					&http.Request{},
					scenario.ToKeyData(),
					dataTemplate,
					contractReq,
				)

				// Verify results
				require.Equal(t, 0, len(res.Errors), fmt.Sprintf("%s errors: %v", scenario.Name, res.Errors))

				// Skip checking results for DELETE since it doesn't return content
				if scenario.Method != types.Delete {
					require.NotEmpty(t, res.Results, "No results returned from "+scenario.Name)

					// Get the first result

					// Additional checks based on operation
					respMap := getFirstMap(res.Results)
					switch {
					case strings.Contains(scenario.Name, "Create"):
						require.Equal(t, resourceID, respMap["id"], "Resource ID mismatch: %v", res.Results)
						require.Equal(t, "test resource", respMap["name"], "Resource name mismatch: %v", res.Results)
					case strings.Contains(scenario.Name, "Get"):
						require.Equal(t, resourceID, respMap["id"], "Resource ID mismatch: %v", respMap)
						require.Equal(t, "test resource", respMap["name"], "Resource name mismatch: %v", res.Results)
					case strings.Contains(scenario.Name, "Update"):
						require.Equal(t, resourceID, respMap["id"], "Resource ID mismatch: %v", res.Results)
						require.Equal(t, "updated name", respMap["name"], "Updated name mismatch: %v", res.Results)
					}
				}
			})
		}
	})

	// --- Test 2: Test By Group ---
	t.Run("Execute By Group", func(t *testing.T) {
		// Set up data template and contract request
		dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
		contractReq := types.NewProducerContractRequest(baseURL, 1)
		contractReq.Headers.Set("Content-Type", "application/json")
		contractReq.Headers.Set("x-api-key", apiKey)
		contractReq.Headers.Set("Authorization", "Bearer "+accessToken)
		contractReq.Params = map[string]interface{}{
			"resource_id":  resourceID,
			"name":         "test resource",
			"description":  "test description",
			"access_token": accessToken,
			"jws":          jws,
		}

		// Execute by Auth group
		t.Run("Auth Group", func(t *testing.T) {
			res := executor.ExecuteByGroup(
				context.Background(),
				&http.Request{},
				"API Testing Suite_Authentication_auth_token",
				dataTemplate,
				contractReq,
			)

			require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Auth group errors: %v", res.Errors))
			require.NotEmpty(t, res.Results, "No results returned from Auth group")
		})

		// Execute by CRUD group
		t.Run("CRUD Group", func(t *testing.T) {
			res := executor.ExecuteByGroup(
				context.Background(),
				&http.Request{},
				"API Testing Suite_CRUD Operations_api_resources",
				dataTemplate,
				contractReq,
			)

			require.Equal(t, 0, len(res.Errors), fmt.Sprintf("CRUD group errors: %v", res.Errors))
			require.NotEmpty(t, res.Results, "No results returned from CRUD group")
		})
	})

	// --- Test 3: Full API Flow Test ---
	t.Run("Full API Flow", func(t *testing.T) {
		// Initial setup
		dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
		contractReq := types.NewProducerContractRequest(baseURL, 1)
		contractReq.Headers.Set("Content-Type", "application/json")
		contractReq.Headers.Set("x-api-key", apiKey)
		contractReq.Params = map[string]interface{}{
			"jws": jws,
		}

		// STEP 1: Auth - Get token
		t.Log("Step 1: Authenticate and get token")
		var authScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Name, "Get JWT Token") {
				authScenario = s
				break
			}
		}
		require.NotNil(t, authScenario)

		authRes := executor.Execute(
			context.Background(),
			&http.Request{},
			authScenario.ToKeyData(),
			dataTemplate,
			contractReq,
		)
		require.Equal(t, 0, len(authRes.Errors), fmt.Sprintf("Auth API errors: %v", authRes.Errors))

		// Extract token from response
		resMap := getFirstMap(authRes.Results)
		extractedToken, exists := resMap["access_token"]
		require.True(t, exists, "Response missing access_token: %v", authRes.Results)
		extractedTokenStr := fmt.Sprintf("%v", extractedToken)
		require.Equal(t, accessToken, extractedTokenStr, "Unexpected access token")

		// Update contract request with token
		contractReq.Headers.Set("Authorization", "Bearer "+extractedTokenStr)
		contractReq.Params["access_token"] = extractedTokenStr

		// STEP 2: Create Resource
		t.Log("Step 2: Create a resource")
		var createScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Name, "Create Resource") {
				createScenario = s
				break
			}
		}
		require.NotNil(t, createScenario)

		createRes := executor.Execute(
			context.Background(),
			&http.Request{},
			createScenario.ToKeyData(),
			dataTemplate,
			contractReq,
		)
		require.Equal(t, 0, len(createRes.Errors), fmt.Sprintf("Create API errors: %v", createRes.Errors))

		// Extract resource ID
		resMap = getFirstMap(createRes.Results)
		extractedID, exists := resMap["id"]
		require.True(t, exists, "Create response missing id")
		extractedIDStr := fmt.Sprintf("%v", extractedID)
		require.Equal(t, resourceID, extractedIDStr, "Unexpected resource ID")

		// Update resource ID in parameters
		contractReq.Params["resource_id"] = extractedIDStr

		// STEP 3: Get Resource
		t.Log("Step 3: Get the created resource")
		var getScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Name, "Get Resource") {
				getScenario = s
				break
			}
		}
		require.NotNil(t, getScenario)

		getRes := executor.Execute(
			context.Background(),
			&http.Request{},
			getScenario.ToKeyData(),
			dataTemplate,
			contractReq,
		)
		require.Equal(t, 0, len(getRes.Errors), fmt.Sprintf("Get API errors: %v", getRes.Errors))

		resMap = getFirstMap(getRes.Results)
		// Verify resource data
		require.Equal(t, resourceID, resMap["id"], "Resource ID mismatch in GET")
		require.Equal(t, "test resource", resMap["name"], "Resource name mismatch in GET")

		// STEP 4: Update Resource
		t.Log("Step 4: Update the resource")
		var updateScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Name, "Update Resource") {
				updateScenario = s
				break
			}
		}
		require.NotNil(t, updateScenario)

		// Add update params
		contractReq.Params["name"] = "updated name"

		updateRes := executor.Execute(
			context.Background(),
			&http.Request{},
			updateScenario.ToKeyData(),
			dataTemplate,
			contractReq,
		)
		require.Equal(t, 0, len(updateRes.Errors), fmt.Sprintf("Update API errors: %v", updateRes.Errors))

		resMap = getFirstMap(updateRes.Results)
		// Verify updated resource
		require.Equal(t, resourceID, resMap["id"], "Resource ID mismatch in UPDATE")
		require.Equal(t, "updated name", resMap["name"], "Updated name mismatch")

		// STEP 5: Delete Resource
		t.Log("Step 5: Delete the resource")
		var deleteScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Name, "Delete Resource") {
				deleteScenario = s
				break
			}
		}
		require.NotNil(t, deleteScenario)

		deleteRes := executor.Execute(
			context.Background(),
			&http.Request{},
			deleteScenario.ToKeyData(),
			dataTemplate,
			contractReq,
		)
		require.Equal(t, 0, len(deleteRes.Errors), fmt.Sprintf("Delete API errors: %v", deleteRes.Errors))
	})
}

func getFirstMap(m map[string]any) map[string]any {
	for _, v := range m {
		if resp, ok := v.(map[string]any); ok {
			return resp
		} else if strResp, ok := v.(map[string]string); ok {
			resp := make(map[string]any)
			for kk, vv := range strResp {
				resp[kk] = vv
			}
			return resp
		}
	}
	return m
}

func saveTestScenario(
	name string,
	scenarioRepo repository.APIScenarioRepository,
) (*types.APIScenario, error) {
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
	err = scenarioRepo.Save(&scenario)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse("http://localhost:8080")
	if err != nil {
		return nil, err
	}
	err = scenarioRepo.SaveHistory(&scenario, u.String(), time.Now(), time.Now().Add(time.Second))
	if err != nil {
		return nil, err
	}
	return &scenario, nil
}
