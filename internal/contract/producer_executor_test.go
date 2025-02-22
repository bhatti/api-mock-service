package contract

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"os"
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
