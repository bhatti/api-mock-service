package contract

import (
	"context"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/pm"
	"gopkg.in/yaml.v3"
	"net/http"
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
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1, 0)
	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, web.NewHTTPClient(config, web.NewAuthAdapter(config)))
	// THEN it should execute saved scenario
	res := executor.Execute(context.Background(), &http.Request{}, &types.APIKeyData{}, dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, res.Mismatched)
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
	contractReq := types.NewProducerContractRequest(baseURL, 1, 0)
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
	contractReq := types.NewProducerContractRequest(baseURL, 1, 0)
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
	scenario.Response.Assertions = []string{"PropertyContains contents.id 10", "PropertyContains contents.title illo"}
	err = scenarioRepository.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1, 0)
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
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1, 0)

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
		scenario.Response.Assertions, "PropertyContains headers.Content-Type application/xjson")
	err = scenarioRepository.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1, 0)

	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, web.NewHTTPClient(config, web.NewAuthAdapter(config)))
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), &http.Request{}, scenario.ToKeyData(), dataTemplate, contractReq)
	for k, err := range res.Errors {
		t.Log(k, err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["put_posts"], `failed to assert response '{{PropertyContains "headers.Content-Type"`)
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
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1, 0)

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
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1, 0)

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
	contractReq := types.NewProducerContractRequest("https://jsonplaceholder.typicode.com", 1, 0)

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
	contractReq := types.NewProducerContractRequest(baseURL, 1, 0)
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
	contractReq := types.NewProducerContractRequest(baseURL, 1, 0)
	// WHEN executing scenario
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, client)
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), &http.Request{}, scenario.ToKeyData(), dataTemplate, contractReq)
	for k, err := range res.Errors {
		t.Log(k, err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["get_comment"], `failed to assert response '{{PropertyContains "contents.id" "1"}}`)
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
	contractReq := types.NewProducerContractRequest(baseURL, 1, 0)
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
	contractReq := types.NewProducerContractRequest(baseURL, 5, 0)
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
	contractReq := types.NewProducerContractRequest(baseURL, 5, 0)
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
			contractReq := types.NewProducerContractRequest(baseURL, 1, 0)
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
				contractReq := types.NewProducerContractRequest(baseURL, 1, 0)
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
		contractReq := types.NewProducerContractRequest(baseURL, 1, 0)
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
		contractReq := types.NewProducerContractRequest(baseURL, 1, 0)
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

func Test_ShouldExecuteTransferAPISuiteAsProducer(t *testing.T) {
	// GIVEN configuration and repositories
	config := types.BuildTestConfig()
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)

	// Define base URL for API server
	baseURL := "https://api.example.com"

	// Load and parse the OpenAPI spec
	data, err := os.ReadFile("../../fixtures/oapi/transfer.json")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := oapi.Parse(context.Background(), &types.Configuration{}, data, dataTempl)
	scenarios := make([]*types.APIScenario, len(specs))
	// Save all scenarios
	// Create test variables for substitution in URL paths and request bodies
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	accessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IlRlc3QgVXNlciIsImlhdCI6MTUxNjIzOTAyMn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	transferId := "transfer-12345"
	accountId := "account-67890"

	for i, spec := range specs {
		scenarios[i], err = spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		if strings.Contains(scenarios[i].Name, "getAuthToken") {
			scenarios[i].Group = "Authentication"
		} else if strings.Contains(scenarios[i].Path, "transfer") {
			scenarios[i].Group = "Transfer"
		} else if strings.Contains(scenarios[i].Path, "asset") {
			scenarios[i].Group = "Assets"
		} else if strings.Contains(scenarios[i].Path, "restrictions") {
			scenarios[i].Group = "Restrictions"
		}
		// Set variables
		scenarios[i].Request.Variables["client_id"] = clientID
		scenarios[i].Request.Variables["client_secret"] = clientSecret
		scenarios[i].Request.Variables["transfer_id"] = transferId
		scenarios[i].Request.Variables["account_id"] = accountId
		scenarios[i].Request.Variables["base_url"] = baseURL
		_, err = yaml.Marshal(scenarios[i])
		require.NoError(t, err)
		require.NoError(t, scenarioRepository.Save(scenarios[i]))
	}

	// Create a stub HTTP client
	client := web.NewStubHTTPClient()

	// Configure stub responses for each API endpoint

	// Auth endpoint
	client.AddMapping("POST", baseURL+"/v1/auth/token", web.NewStubHTTPResponse(200,
		map[string]interface{}{
			"accessToken": accessToken,
			"tokenType":   "Bearer",
			"expiresIn":   3600,
			"scope":       "read write",
			"issuedAt":    time.Now().Format(time.RFC3339),
		}).WithHeader("Content-Type", "application/json"))

	// List Transfers endpoint
	client.AddMapping("GET", baseURL+"/v1/transfers", web.NewStubHTTPResponse(200,
		map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":            transferId,
					"controlNumber": "CN-12345",
					"status":        "PENDING",
					"direction":     "INBOUND",
					"transferType":  "PARTIAL",
					"sourceAccount": map[string]interface{}{
						"id":              "src-acc-1",
						"number":          "123456789",
						"institutionName": "Source Bank",
					},
					"destinationAccount": map[string]interface{}{
						"id":              "dst-acc-1",
						"number":          "987654321",
						"institutionName": "Destination Bank",
					},
					"createdAt":  time.Now().Format(time.RFC3339),
					"updatedAt":  time.Now().Format(time.RFC3339),
					"totalValue": 5000.00,
				},
			},
			"pagination": map[string]interface{}{
				"totalItems":   1,
				"totalPages":   1,
				"currentPage":  1,
				"itemsPerPage": 20,
			},
			"meta": map[string]interface{}{
				"requestId":      "req-abc-123",
				"timestamp":      time.Now().Format(time.RFC3339),
				"processingTime": 0.235,
				"version":        "v1",
			},
		}).WithHeader("Content-Type", "application/json"))

	// Create Transfer endpoint
	client.AddMapping("POST", baseURL+"/v1/transfers", web.NewStubHTTPResponse(201,
		map[string]interface{}{
			"id":            transferId,
			"controlNumber": "CN-12345",
			"status":        "PENDING",
			"source": map[string]interface{}{
				"accountId":       "src-acc-1",
				"accountNumber":   "123456789",
				"institutionName": "Source Bank",
			},
			"destination": map[string]interface{}{
				"accountId":     "dst-acc-1",
				"accountNumber": "987654321",
			},
			"transferType":            "PARTIAL",
			"createdAt":               time.Now().Format(time.RFC3339),
			"estimatedCompletionDate": time.Now().Add(time.Hour * 24 * 5).Format(time.RFC3339),
			"validations": []map[string]interface{}{
				{
					"type":    "account_verification",
					"status":  "PASSED",
					"message": "Account verification completed successfully",
				},
			},
			"nextActions": []string{"upload_documents", "approve_transfer"},
		}).WithHeader("Content-Type", "application/json"))

	// Get Transfer details endpoint
	client.AddMapping("GET", baseURL+"/v1/transfers/{transferId}", web.NewStubHTTPResponse(200,
		map[string]interface{}{
			"id":                transferId,
			"controlNumber":     "CN-12345",
			"clientReferenceId": "client-ref-001",
			"status":            "PENDING",
			"direction":         "INBOUND",
			"transferType":      "PARTIAL",
			"source": map[string]interface{}{
				"id":          "src-acc-1",
				"number":      "123456789",
				"accountType": "INDIVIDUAL",
				"institution": map[string]interface{}{
					"id":        "inst-1",
					"name":      "Source Bank",
					"dtcNumber": "1234",
					"type":      "BROKER",
				},
			},
			"destination": map[string]interface{}{
				"id":          "dst-acc-1",
				"number":      "987654321",
				"accountType": "INDIVIDUAL",
				"institution": map[string]interface{}{
					"id":        "inst-2",
					"name":      "Destination Bank",
					"dtcNumber": "5678",
					"type":      "BROKER",
				},
			},
			"assets": []map[string]interface{}{
				{
					"id":        "asset-1",
					"assetType": "EQUITY",
					"identifiers": map[string]interface{}{
						"cusip":  "12345ABC",
						"symbol": "AAPL",
					},
					"name":         "Apple Inc.",
					"quantity":     10,
					"price":        150.00,
					"marketValue":  1500.00,
					"currency":     "USD",
					"positionType": "LONG",
					"status":       "PENDING",
				},
			},
			"createdAt":  time.Now().Format(time.RFC3339),
			"updatedAt":  time.Now().Format(time.RFC3339),
			"totalValue": 1500.00,
		}).WithHeader("Content-Type", "application/json"))

	// Update Transfer endpoint
	client.AddMapping("PUT", baseURL+"/v1/transfers/{transferId}", web.NewStubHTTPResponse(200,
		map[string]interface{}{
			"id":            transferId,
			"controlNumber": "CN-12345",
			"status":        "IN_PROGRESS",
			"source": map[string]interface{}{
				"accountId":       "src-acc-1",
				"accountNumber":   "123456789",
				"institutionName": "Source Bank",
			},
			"destination": map[string]interface{}{
				"accountId":     "dst-acc-1",
				"accountNumber": "987654321",
			},
			"transferType":            "PARTIAL",
			"createdAt":               time.Now().Format(time.RFC3339),
			"estimatedCompletionDate": time.Now().Add(time.Hour * 24 * 5).Format(time.RFC3339),
			"validations": []map[string]interface{}{
				{
					"type":    "approval_verification",
					"status":  "PASSED",
					"message": "Approval completed successfully",
				},
			},
			"nextActions": []string{"wait_for_completion"},
		}).WithHeader("Content-Type", "application/json"))

	// Cancel Transfer endpoint
	client.AddMapping("DELETE", baseURL+"/v1/transfers/{transferId}",
		web.NewStubHTTPResponse(204, nil).WithHeader("Content-Type", "application/json"))

	// List Assets endpoint
	client.AddMapping("GET", baseURL+"/v1/assets", web.NewStubHTTPResponse(200,
		map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":        "asset-1",
					"assetType": "EQUITY",
					"identifiers": map[string]interface{}{
						"cusip":  "12345ABC",
						"symbol": "AAPL",
					},
					"name":         "Apple Inc.",
					"quantity":     10,
					"price":        150.00,
					"marketValue":  1500.00,
					"currency":     "USD",
					"positionType": "LONG",
					"status":       "PENDING",
				},
			},
			"pagination": map[string]interface{}{
				"totalItems":   1,
				"totalPages":   1,
				"currentPage":  1,
				"itemsPerPage": 20,
			},
			"meta": map[string]interface{}{
				"requestId":      "req-abc-123",
				"timestamp":      time.Now().Format(time.RFC3339),
				"processingTime": 0.235,
				"version":        "v1",
			},
		}).WithHeader("Content-Type", "application/json"))

	// Get Restrictions endpoint
	client.AddMapping("GET", baseURL+"/v1/restrictions", web.NewStubHTTPResponse(200,
		map[string]interface{}{
			"restrictions": map[string]interface{}{
				"blacklist": []map[string]interface{}{
					{
						"cusip":       "999BLACKLISTED",
						"symbol":      "BLCK",
						"description": "Blacklisted security",
						"reason":      "Regulatory restriction",
						"accountId":   accountId,
						"createdAt":   time.Now().Format(time.RFC3339),
						"createdBy":   "system",
					},
				},
				"graylist":  []map[string]interface{}{},
				"whitelist": []map[string]interface{}{},
				"assetTypes": []map[string]interface{}{
					{
						"assetType":  "OPTION",
						"restricted": true,
						"reason":     "Account not option-enabled",
						"accountId":  accountId,
						"createdAt":  time.Now().Format(time.RFC3339),
						"createdBy":  "system",
					},
				},
			},
			"meta": map[string]interface{}{
				"requestId":      "req-abc-123",
				"timestamp":      time.Now().Format(time.RFC3339),
				"processingTime": 0.235,
				"version":        "v1",
			},
		}).WithHeader("Content-Type", "application/json"))

	// Create Restrictions endpoint
	client.AddMapping("POST", baseURL+"/v1/restrictions", web.NewStubHTTPResponse(201,
		map[string]interface{}{
			"accountId": accountId,
			"appliedRestrictions": map[string]interface{}{
				"blacklist":  1,
				"graylist":   0,
				"whitelist":  0,
				"assetTypes": 1,
			},
			"createdAt": time.Now().Format(time.RFC3339),
			"createdBy": "test-user",
		}).WithHeader("Content-Type", "application/json"))

	// Save all scenarios
	for _, scenario := range scenarios {
		// Update BaseURL to empty to use the one from contractReq
		scenario.BaseURL = ""

		// Save scenario
		err = scenarioRepository.Save(scenario)
		require.NoError(t, err)
	}

	// Create producer executor with stub client
	executor := NewProducerExecutor(scenarioRepository, groupConfigRepository, client)

	// --- Test 1: Test Individual APIs ---
	t.Run("Individual API Tests", func(t *testing.T) {
		// Find auth scenario
		var authScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Path, "/auth/token") && s.Method == types.Post {
				authScenario = s
				break
			}
		}
		require.NotNil(t, authScenario, "Auth scenario not found")

		// Test auth API
		t.Run("Auth API Test", func(t *testing.T) {
			// Set up data template and contract request
			dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
			contractReq := types.NewProducerContractRequest(baseURL, 1, 200)
			contractReq.Headers.Set("Content-Type", "application/json")
			contractReq.Params = map[string]interface{}{
				"client_id":     clientID,
				"client_secret": clientSecret,
			}

			// Execute the auth scenario
			keyData := authScenario.ToKeyData()
			keyData.Name = ""
			contractReq.MatchResponseCode = 200
			res := executor.Execute(
				context.Background(),
				&http.Request{},
				keyData,
				dataTemplate,
				contractReq,
			)

			// Verify results
			require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Auth API errors: %v", res.Errors))
			mapResp := getFirstMap(res.Results)
			require.NotEmpty(t, mapResp, "No results returned from auth API")

			// Check for token in response
			require.Contains(t, mapResp, "accessToken", "Response missing accessToken")
			require.Equal(t, accessToken, mapResp["accessToken"], "Unexpected access token")
		})

		// Test transfer API endpoints
		t.Run("Transfers API Tests", func(t *testing.T) {
			// Find transfer scenarios
			var listTransfersScenario, createTransferScenario, getTransferScenario, updateTransferScenario, cancelTransferScenario *types.APIScenario

			for _, s := range scenarios {
				if strings.Contains(s.Path, "/transfers") {
					switch s.Method {
					case types.Get:
						if !strings.Contains(s.Path, "/transfers/") {
							listTransfersScenario = s
						} else {
							getTransferScenario = s
						}
					case types.Post:
						createTransferScenario = s
					case types.Put:
						updateTransferScenario = s
					case types.Delete:
						cancelTransferScenario = s
					}
				}
			}

			require.NotNil(t, listTransfersScenario, "List transfers scenario not found")
			require.NotNil(t, createTransferScenario, "Create transfer scenario not found")
			require.NotNil(t, getTransferScenario, "Get transfer scenario not found")
			require.NotNil(t, updateTransferScenario, "Update transfer scenario not found")
			require.NotNil(t, cancelTransferScenario, "Cancel transfer scenario not found")

			// Common setup for transfer tests
			dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
			contractReq := types.NewProducerContractRequest(baseURL, 1, 200)
			contractReq.Headers.Set("Content-Type", "application/json")
			contractReq.Headers.Set("Authorization", "Bearer "+accessToken)
			contractReq.Params = map[string]interface{}{
				"transfer_id":  transferId,
				"account_id":   accountId,
				"access_token": accessToken,
			}

			// Test List Transfers
			t.Run("List Transfers", func(t *testing.T) {
				keyData := listTransfersScenario.ToKeyData()
				keyData.Name = ""
				res := executor.Execute(
					context.Background(),
					&http.Request{},
					keyData,
					dataTemplate,
					contractReq,
				)
				require.Equal(t, 0, len(res.Errors), fmt.Sprintf("List Transfers errors: %v", res.Errors))
				require.Equal(t, 1, res.Succeeded)
				mapResp := getFirstMap(res.Results)
				require.NotEmpty(t, mapResp, "No results returned from List Transfers API")

				// Verify we got data array
				require.Contains(t, mapResp, "data", "Response missing data array")
				require.Contains(t, mapResp, "pagination", "Response missing pagination")
			})

			// Test Create Transfer
			t.Run("Create Transfer", func(t *testing.T) {
				// Add transfer-specific parameters
				contractReq.Params["sourceAccount"] = map[string]interface{}{
					"number":        "123456789",
					"institutionId": "inst-1",
				}
				contractReq.Params["destinationAccount"] = map[string]interface{}{
					"number": "987654321",
				}
				contractReq.Params["transferType"] = "PARTIAL"

				keyData := createTransferScenario.ToKeyData()
				keyData.Name = ""
				contractReq.MatchResponseCode = 201
				res := executor.Execute(
					context.Background(),
					&http.Request{},
					keyData,
					dataTemplate,
					contractReq,
				)
				require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Create Transfer errors: %v", res.Errors))
				mapResp := getFirstMap(res.Results)
				require.NotEmpty(t, mapResp, "No results returned from Create Transfer API")

				// Verify specific fields
				require.Equal(t, transferId, mapResp["id"], "Transfer ID mismatch")
				require.Equal(t, "PENDING", mapResp["status"], "Status should be PENDING")
			})

			// Test Get Transfer
			t.Run("Get Transfer", func(t *testing.T) {
				keyData := getTransferScenario.ToKeyData()
				keyData.Name = ""
				contractReq.MatchResponseCode = 200
				res := executor.Execute(
					context.Background(),
					&http.Request{},
					keyData,
					dataTemplate,
					contractReq,
				)
				require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Get Transfer errors: %v", res.Errors))
				mapResp := getFirstMap(res.Results)
				require.NotEmpty(t, mapResp, "No results returned from Get Transfer API")

				// Verify specific fields
				require.Equal(t, transferId, mapResp["id"], "Transfer ID mismatch")
				require.Contains(t, mapResp, "assets", "Response missing assets array")
			})

			// Test Update Transfer
			t.Run("Update Transfer", func(t *testing.T) {
				// Add update-specific parameters
				contractReq.Params["action"] = "APPROVE"
				contractReq.Params["approvalDetails"] = map[string]interface{}{
					"approvalLevel": "STANDARD",
					"comment":       "Approved by test",
				}

				keyData := updateTransferScenario.ToKeyData()
				keyData.Name = ""
				res := executor.Execute(
					context.Background(),
					&http.Request{},
					keyData,
					dataTemplate,
					contractReq,
				)
				require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Update Transfer errors: %v", res.Errors))
				mapResp := getFirstMap(res.Results)
				require.NotEmpty(t, mapResp, "No results returned from Update Transfer API")

				// Verify updated status
				require.Equal(t, "IN_PROGRESS", mapResp["status"], "Status should be IN_PROGRESS after approval")
			})

			// Test Cancel Transfer
			t.Run("Cancel Transfer", func(t *testing.T) {
				keyData := cancelTransferScenario.ToKeyData()
				keyData.Name = ""
				contractReq.MatchResponseCode = 204
				res := executor.Execute(
					context.Background(),
					&http.Request{},
					keyData,
					dataTemplate,
					contractReq,
				)
				require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Cancel Transfer errors: %v", res.Errors))
				// DELETE responses are often empty
			})
		})

		// Test assets API
		t.Run("Assets API Test", func(t *testing.T) {
			var listAssetsScenario *types.APIScenario
			for _, s := range scenarios {
				if strings.Contains(s.Path, "/assets") && s.Method == types.Get {
					listAssetsScenario = s
					break
				}
			}
			require.NotNil(t, listAssetsScenario, "List assets scenario not found")

			dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
			contractReq := types.NewProducerContractRequest(baseURL, 1, 200)
			contractReq.Headers.Set("Content-Type", "application/json")
			contractReq.Headers.Set("Authorization", "Bearer "+accessToken)
			contractReq.Params = map[string]interface{}{
				"account_id": accountId,
			}

			keyData := listAssetsScenario.ToKeyData()
			keyData.Name = ""
			contractReq.MatchResponseCode = 200
			res := executor.Execute(
				context.Background(),
				&http.Request{},
				keyData,
				dataTemplate,
				contractReq,
			)
			require.Equal(t, 0, len(res.Errors), fmt.Sprintf("List Assets errors: %v", res.Errors))
			mapResp := getFirstMap(res.Results)
			require.NotEmpty(t, mapResp, "No results returned from List Assets API")

			// Verify assets data
			require.Contains(t, mapResp, "data", "Response missing data array")
		})

		// Test restrictions API endpoints
		t.Run("Restrictions API Tests", func(t *testing.T) {
			var getRestrictionsScenario, createRestrictionsScenario *types.APIScenario
			for _, s := range scenarios {
				if strings.Contains(s.Path, "/restrictions") {
					if s.Method == types.Get {
						getRestrictionsScenario = s
					} else if s.Method == types.Post {
						createRestrictionsScenario = s
					}
				}
			}
			require.NotNil(t, getRestrictionsScenario, "Get restrictions scenario not found")
			require.NotNil(t, createRestrictionsScenario, "Create restrictions scenario not found")

			dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
			contractReq := types.NewProducerContractRequest(baseURL, 1, 200)
			contractReq.Headers.Set("Content-Type", "application/json")
			contractReq.Headers.Set("Authorization", "Bearer "+accessToken)
			contractReq.Params = map[string]interface{}{
				"account_id": accountId,
			}

			// Test Get Restrictions
			t.Run("Get Restrictions", func(t *testing.T) {
				keyData := getRestrictionsScenario.ToKeyData()
				keyData.Name = ""
				contractReq.MatchResponseCode = 200
				res := executor.Execute(
					context.Background(),
					&http.Request{},
					keyData,
					dataTemplate,
					contractReq,
				)
				require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Get Restrictions errors: %v", res.Errors))
				mapResp := getFirstMap(res.Results)
				require.NotEmpty(t, mapResp, "No results returned from Get Restrictions API")

				// Verify restrictions data
				require.Contains(t, mapResp, "restrictions", "Response missing restrictions object")
			})

			// Test Create Restrictions
			t.Run("Create Restrictions", func(t *testing.T) {
				// Add restriction-specific parameters
				contractReq.Params["blacklist"] = []map[string]interface{}{
					{
						"identifier": map[string]interface{}{
							"type":  "CUSIP",
							"value": "999BLACKLISTED",
						},
						"description": "Blacklisted security",
						"reason":      "Regulatory restriction",
					},
				}
				contractReq.Params["assetTypes"] = []map[string]interface{}{
					{
						"assetType":  "OPTION",
						"restricted": true,
						"reason":     "Account not option-enabled",
					},
				}

				keyData := createRestrictionsScenario.ToKeyData()
				keyData.Name = ""
				contractReq.MatchResponseCode = 201

				res := executor.Execute(
					context.Background(),
					&http.Request{},
					keyData,
					dataTemplate,
					contractReq,
				)
				require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Create Restrictions errors: %v", res.Errors))
				mapResp := getFirstMap(res.Results)
				require.NotEmpty(t, mapResp, "No results returned from Create Restrictions API")

				// Verify applied restrictions
				require.Contains(t, mapResp, "appliedRestrictions", "Response missing appliedRestrictions")
				restrictions := mapResp["appliedRestrictions"].(map[string]interface{})
				require.Equal(t, float64(1), restrictions["blacklist"], "Expected 1 blacklist restriction applied")
				require.Equal(t, float64(1), restrictions["assetTypes"], "Expected 1 assetType restriction applied")
			})
		})
	})

	// --- Test 2: Test By Group ---
	t.Run("Execute By Group", func(t *testing.T) {
		// Set up data template and contract request
		dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
		contractReq := types.NewProducerContractRequest(baseURL, 1, 200)
		contractReq.Headers.Set("Content-Type", "application/json")
		contractReq.Headers.Set("Authorization", "Bearer "+accessToken)
		contractReq.Params = map[string]interface{}{
			"client_id":     clientID,
			"client_secret": clientSecret,
			"transfer_id":   transferId,
			"account_id":    accountId,
		}

		contractReq.MatchResponseCode = 200
		// Execute Authentication group
		t.Run("Authentication Group", func(t *testing.T) {
			res := executor.ExecuteByGroup(
				context.Background(),
				&http.Request{},
				"Authentication",
				dataTemplate,
				contractReq,
			)
			require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Authentication group errors: %v", res.Errors))
			require.Equal(t, 1, res.Succeeded)
		})

		// Execute Transfers group
		t.Run("Transfers Group", func(t *testing.T) {
			// Add transfer-specific parameters
			contractReq.Params["sourceAccount"] = map[string]interface{}{
				"number":        "123456789",
				"institutionId": "inst-1",
			}
			contractReq.Params["destinationAccount"] = map[string]interface{}{
				"number": "987654321",
			}
			contractReq.Params["transferType"] = "PARTIAL"
			contractReq.Params["action"] = "APPROVE"
			contractReq.Params["approvalDetails"] = map[string]interface{}{
				"approvalLevel": "STANDARD",
				"comment":       "Approved by test",
			}

			res := executor.ExecuteByGroup(
				context.Background(),
				&http.Request{},
				"Transfers",
				dataTemplate,
				contractReq,
			)
			require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Transfers group errors: %v", res.Errors))
		})

		// Execute Assets group
		t.Run("Assets Group", func(t *testing.T) {
			res := executor.ExecuteByGroup(
				context.Background(),
				&http.Request{},
				"Assets",
				dataTemplate,
				contractReq,
			)
			require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Assets group errors: %v", res.Errors))
			require.NotEmpty(t, res.Results, "No results returned from Assets group")
		})

		// Execute Restrictions group
		t.Run("Restrictions Group", func(t *testing.T) {
			// Add restriction-specific parameters
			contractReq.Params["blacklist"] = []map[string]interface{}{
				{
					"identifier": map[string]interface{}{
						"type":  "CUSIP",
						"value": "999BLACKLISTED",
					},
					"description": "Blacklisted security",
					"reason":      "Regulatory restriction",
				},
			}
			contractReq.Params["assetTypes"] = []map[string]interface{}{
				{
					"assetType":  "OPTION",
					"restricted": true,
					"reason":     "Account not option-enabled",
				},
			}
			res := executor.ExecuteByGroup(
				context.Background(),
				&http.Request{},
				"Restrictions",
				dataTemplate,
				contractReq,
			)
			require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Restrictions group errors: %v", res.Errors))
			require.NotEmpty(t, res.Results, "No results returned from Restrictions group")
		})
	})

	// --- Test 3: Full API Flow Test ---
	t.Run("Full API Flow", func(t *testing.T) {
		// Initial setup
		dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
		contractReq := types.NewProducerContractRequest(baseURL, 1, 200)
		contractReq.Headers.Set("Content-Type", "application/json")
		contractReq.Params = map[string]interface{}{
			"client_id":     clientID,
			"client_secret": clientSecret,
		}

		// STEP 1: Auth - Get token
		t.Log("Step 1: Authenticate and get token")
		var authScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Path, "/auth/token") && s.Method == types.Post {
				authScenario = s
				break
			}
		}
		require.NotNil(t, authScenario)
		keyData := authScenario.ToKeyData()
		contractReq.MatchResponseCode = 200
		authRes := executor.Execute(
			context.Background(),
			&http.Request{},
			keyData,
			dataTemplate,
			contractReq,
		)
		require.Equal(t, 0, len(authRes.Errors), fmt.Sprintf("Auth API errors: %v", authRes.Errors))

		// Extract token from response
		resMap := getFirstMap(authRes.Results)
		extractedToken, exists := resMap["accessToken"]
		require.True(t, exists, "Response missing accessToken: %v", authRes.Results)
		extractedTokenStr := fmt.Sprintf("%v", extractedToken)
		require.Equal(t, accessToken, extractedTokenStr, "Unexpected access token")

		// Update contract request with token
		contractReq.Headers.Set("Authorization", "Bearer "+extractedTokenStr)
		contractReq.Params["access_token"] = extractedTokenStr

		// STEP 2: Create Transfer
		t.Log("Step 2: Create a transfer")
		var createTransferScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Path, "/transfers") && s.Method == types.Post {
				createTransferScenario = s
				break
			}
		}
		require.NotNil(t, createTransferScenario)

		// Add transfer creation parameters
		contractReq.Params["sourceAccount"] = map[string]interface{}{
			"number":        "123456789",
			"institutionId": "inst-1",
		}
		contractReq.Params["destinationAccount"] = map[string]interface{}{
			"number": "987654321",
		}
		contractReq.Params["transferType"] = "PARTIAL"
		contractReq.Params["assets"] = []map[string]interface{}{
			{
				"assetType": "EQUITY",
				"identifiers": map[string]interface{}{
					"cusip":  "12345ABC",
					"symbol": "AAPL",
				},
				"quantity":     10,
				"positionType": "LONG",
			},
		}

		keyData = createTransferScenario.ToKeyData()
		keyData.Name = ""
		contractReq.MatchResponseCode = 201
		createRes := executor.Execute(
			context.Background(),
			&http.Request{},
			keyData,
			dataTemplate,
			contractReq,
		)
		require.Equal(t, 0, len(createRes.Errors), fmt.Sprintf("Create Transfer API errors: %v", createRes.Errors))

		// Extract transfer ID
		resMap = getFirstMap(createRes.Results)
		extractedID, exists := resMap["id"]
		require.True(t, exists, "Create response missing id")
		extractedIDStr := fmt.Sprintf("%v", extractedID)
		require.Equal(t, transferId, extractedIDStr, "Unexpected transfer ID")

		// Update transfer ID in parameters
		contractReq.Params["transfer_id"] = extractedIDStr

		// STEP 3: Get Transfer
		t.Log("Step 3: Get the created transfer")
		var getTransferScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Path, "/transfers/") && s.Method == types.Get {
				getTransferScenario = s
				break
			}
		}
		require.NotNil(t, getTransferScenario)

		keyData = getTransferScenario.ToKeyData()
		keyData.Name = ""
		contractReq.MatchResponseCode = 200
		getRes := executor.Execute(
			context.Background(),
			&http.Request{},
			keyData,
			dataTemplate,
			contractReq,
		)
		require.Equal(t, 0, len(getRes.Errors), fmt.Sprintf("Get Transfer API errors: %v", getRes.Errors))

		// STEP 4: Update Transfer (approve)
		t.Log("Step 4: Update the transfer (approve)")
		var updateTransferScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Path, "/transfers/") && s.Method == types.Put {
				updateTransferScenario = s
				break
			}
		}
		require.NotNil(t, updateTransferScenario)

		// Add update params
		contractReq.Params["action"] = "APPROVE"
		contractReq.Params["approvalDetails"] = map[string]interface{}{
			"approvalLevel": "STANDARD",
			"comment":       "Approved by test",
		}

		updateRes := executor.Execute(
			context.Background(),
			&http.Request{},
			updateTransferScenario.ToKeyData(),
			dataTemplate,
			contractReq,
		)
		require.Equal(t, 0, len(updateRes.Errors), fmt.Sprintf("Update Transfer API errors: %v", updateRes.Errors))

		// STEP 5: Get restrictions to check if assets are transferable
		t.Log("Step 5: Check restrictions for the assets")
		var getRestrictionsScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Path, "/restrictions") && s.Method == types.Get {
				getRestrictionsScenario = s
				break
			}
		}
		require.NotNil(t, getRestrictionsScenario)

		// Add account ID parameter
		contractReq.Params["account_id"] = accountId

		restrictionsRes := executor.Execute(
			context.Background(),
			&http.Request{},
			getRestrictionsScenario.ToKeyData(),
			dataTemplate,
			contractReq,
		)
		require.Equal(t, 0, len(restrictionsRes.Errors), fmt.Sprintf("Get Restrictions API errors: %v", restrictionsRes.Errors))

		// STEP 6: Cancel Transfer
		t.Log("Step 6: Cancel the transfer")
		var cancelTransferScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Path, "/transfers/") && s.Method == types.Delete {
				cancelTransferScenario = s
				break
			}
		}
		require.NotNil(t, cancelTransferScenario)

		deleteRes := executor.Execute(
			context.Background(),
			&http.Request{},
			cancelTransferScenario.ToKeyData(),
			dataTemplate,
			contractReq,
		)
		require.Equal(t, 0, len(deleteRes.Errors), fmt.Sprintf("Cancel Transfer API errors: %v", deleteRes.Errors))

		// Success - we've completed the full API flow
		t.Log("Successfully completed full API flow test")
	})

	// --- Test 4: Contract Coverage ---
	t.Run("Contract Coverage Tests", func(t *testing.T) {
		// Find a transfer scenario to analyze coverage
		var transferScenario *types.APIScenario
		for _, s := range scenarios {
			if strings.Contains(s.Path, "/transfers/") && s.Method == types.Get {
				transferScenario = s
				break
			}
		}
		require.NotNil(t, transferScenario)

		// Run with coverage tracking
		dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
		contractReq := types.NewProducerContractRequest(baseURL, 1, 200)
		contractReq.Headers.Set("Content-Type", "application/json")
		contractReq.Headers.Set("Authorization", "Bearer "+accessToken)
		contractReq.Params = map[string]interface{}{
			"transfer_id": transferId,
		}
		contractReq.TrackCoverage = true // Enable coverage tracking

		res := executor.Execute(
			context.Background(),
			&http.Request{},
			transferScenario.ToKeyData(),
			dataTemplate,
			contractReq,
		)
		require.Equal(t, 0, len(res.Errors), fmt.Sprintf("Transfer coverage test errors: %v", res.Errors))

		// Verify coverage results
		//coverageKey := transferScenario.Name + "_coverage"
		//require.Contains(t, res.Results, coverageKey, "Coverage metrics not found in results")

		// Get coverage statistics
		// TODO fix
		//stats, err := executor.GetContractStats(transferScenario.Name)
		//require.NoError(t, err, "Error getting contract stats")
		//t.Logf("Contract stats: %d executions, %.2f%% success rate, %.2f ms avg latency",
		//	stats.TotalExecutions, stats.SuccessRate, stats.AverageLatency)
	})
}
