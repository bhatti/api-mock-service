package contract

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"testing"

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
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewContractRequest("https://jsonplaceholder.typicode.com", 1)
	// WHEN executing scenario
	config := &types.Configuration{DataDir: "../../mock_tests"}
	executor := NewExecutor(repo, web.NewHTTPClient(config, web.NewAWSSigner(config)))
	// THEN it should execute saved scenario
	res := executor.Execute(context.Background(), &types.MockScenarioKeyData{}, dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors[""], `could not lookup matching API`)
}

func Test_ShouldExecuteChainedGroupScenarios(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND a valid scenario
	_, err = saveTestScenario("../../fixtures/create_user.yaml", repo)
	require.NoError(t, err)
	_, err = saveTestScenario("../../fixtures/get_user.yaml", repo)
	require.NoError(t, err)
	_, err = saveTestScenario("../../fixtures/users.yaml", repo)
	require.NoError(t, err)

	// AND valid template for random data
	contractReq := types.NewContractRequest(baseURL, 1)
	client := web.NewStubHTTPClient()
	client.AddMapping("POST", baseURL+"/users", web.NewStubHTTPResponse(200,
		`{"User": {"Directory": "my_dir", "Username": "my_user@foo.cc", "DesiredDeliveryMediums": ["EMAIL"]}}`))
	client.AddMapping("GET", baseURL+"/users/1", web.NewStubHTTPResponse(200,
		`{"User": {"Directory": "my_dir2", "Username": "my_user2@foo.cc", "DesiredDeliveryMediums": ["EMAIL"]}}`))
	client.AddMapping("GET", baseURL+"/users", web.NewStubHTTPResponse(200,
		`{"User": {"Directory": "my_dir3", "Username": "my_user3@foo.cc", "DesiredDeliveryMediums": ["EMAIL"]}}`))
	// WHEN executing scenario
	executor := NewExecutor(repo, client)
	// THEN it should execute saved scenario
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	res := executor.ExecuteByGroup(context.Background(), "user_group", dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 0, len(res.Errors), fmt.Sprintf("%v", res.Errors))
}

func Test_ShouldExecuteGetTodo(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/get_todo.yaml", repo)
	require.NoError(t, err)
	scenario.Path = "/todos/10"
	scenario.Response.Assertions = []string{"VariableContains contents.id 10", "VariableContains contents.title illo"}
	err = repo.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewContractRequest("https://jsonplaceholder.typicode.com", 1)
	// WHEN executing scenario
	config := &types.Configuration{DataDir: "../../mock_tests"}
	executor := NewExecutor(repo, web.NewHTTPClient(config, web.NewAWSSigner(config)))
	// THEN it should execute saved scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 0, len(res.Errors), fmt.Sprintf("%v", res.Errors))
}

func Test_ShouldExecutePutPosts(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/put_posts.yaml", repo)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewContractRequest("https://jsonplaceholder.typicode.com", 1)

	// WHEN executing scenario
	config := &types.Configuration{DataDir: "../../mock_tests"}
	executor := NewExecutor(repo, web.NewHTTPClient(config, web.NewAWSSigner(config)))
	// THEN it should execute saved scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 0, len(res.Errors))
}

func Test_ShouldNotExecutePutPostsWithBadHeaderAssertions(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/put_posts.yaml", repo)
	require.NoError(t, err)
	// AND a bad assertion
	scenario.Request.Headers[web.Authorization] = "AWS4-HMAC-SHA256"
	scenario.Response.Assertions = append(scenario.Response.Assertions, "VariableContains headers.Content-Type application/xjson")
	err = repo.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewContractRequest("https://jsonplaceholder.typicode.com", 1)

	// WHEN executing scenario
	config := &types.Configuration{DataDir: "../../mock_tests"}
	executor := NewExecutor(repo, web.NewHTTPClient(config, web.NewAWSSigner(config)))
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["put_posts"], `failed to assert '{{VariableContains "headers.Content-Type"`)
}

func Test_ShouldParseRegexValue(t *testing.T) {
	require.Equal(t, "__1", regexValue("__1"))
	require.Equal(t, "1", regexValue("(1)"))
	require.Equal(t, "1", regexValue("1"))
}

func Test_ShouldNotExecutePutPostsWithBadHeaders(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/put_posts.yaml", repo)
	require.NoError(t, err)
	// AND bad matching header
	scenario.Response.AssertHeadersPattern[types.ContentTypeHeader] = "application/xjson"
	err = repo.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewContractRequest("https://jsonplaceholder.typicode.com", 1)

	// AND executor
	config := &types.Configuration{DataDir: "../../mock_tests"}
	executor := NewExecutor(repo, web.NewHTTPClient(config, web.NewAWSSigner(config)))

	// WHEN executing scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, contractReq)
	// THEN it should not execute saved scenario
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["put_posts"], `didn't match required header Content-Type with regex application/xjson`)
}

func Test_ShouldNotExecutePutPostsWithMissingHeaders(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/put_posts.yaml", repo)
	require.NoError(t, err)
	// AND missing header
	scenario.Response.AssertHeadersPattern["Abc-Content-Type"] = "application/xjson"
	err = repo.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewContractRequest("https://jsonplaceholder.typicode.com", 1)

	// WHEN executing scenario
	config := &types.Configuration{DataDir: "../../mock_tests"}
	executor := NewExecutor(repo, web.NewHTTPClient(config, web.NewAWSSigner(config)))
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["put_posts"], `failed to find required header Abc-Content-Type`)
}

func Test_ShouldExecutePostProductScenario(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/save_product.yaml", repo)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	product := `{"category":"BOOKS","id":"123","inventory":"10","name":"toy 1","price":{"amount":12,"currency":"USD"}}`
	client.AddMapping("POST", baseURL+"/products", web.NewStubHTTPResponse(200, product))

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewContractRequest(baseURL, 1)
	// WHEN executing scenario
	executor := NewExecutor(repo, client)
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 0, len(res.Errors))
}

func Test_ShouldExecuteGetTodoWithBadAssertions(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/get_comment.yaml", repo)
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
	contractReq := types.NewContractRequest(baseURL, 1)
	// WHEN executing scenario
	executor := NewExecutor(repo, client)
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["get_comment"], `failed to assert '{{VariableContains "contents.id" "1"}}`)
}

func Test_ShouldExecuteGetTodoWithBadStatus(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/get_comment.yaml", repo)
	require.NoError(t, err)

	client := web.NewStubHTTPClient()
	todo := `{} `
	client.AddMapping("GET", baseURL+"/comments/1", web.NewStubHTTPResponse(400, todo))

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewContractRequest(baseURL, 1)
	// WHEN executing scenario
	executor := NewExecutor(repo, client)
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors["get_comment"], `failed to execute request with status 400`)
}

func Test_ShouldExecuteJobsOpenAPIWithInvalidStatus(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND mock scenarios from open-api specifications
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)

	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
	contractReq := types.NewContractRequest(baseURL, 5)
	specs, err := oapi.Parse(context.Background(), b, dataTemplate)
	require.NoError(t, err)

	// AND save specs
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTemplate)
		require.NoError(t, err)
		if scenario.Response.StatusCode == 200 {
			scenario.Group = "bad_v1_job"
		}
		err = repo.Save(scenario)
		require.NoError(t, err)
	}
	// AND valid template for random data
	// AND mock web client
	client, data := buildJobsTestClient("AC1234567890", "BAD", "", 200)
	contractReq.Overrides = data
	contractReq.Verbose = true
	// AND executor
	executor := NewExecutor(repo, client)
	// WHEN executing scenario
	res := executor.ExecuteByGroup(context.Background(), "bad_v1_job", dataTemplate, contractReq)
	for _, err := range res.Errors {
		t.Log(err)
		// THEN it should fail to execute
		require.Contains(t, err, "key 'jobStatus' - value 'BAD' didn't match regex")
	}
}

func Test_ShouldExecuteJobsOpenAPI(t *testing.T) {
	// GIVEN mock scenarios from open-api specifications
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	// AND scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	contractReq := types.NewContractRequest(baseURL, 5)
	specs, err := oapi.Parse(context.Background(), b, dataTemplate)
	require.NoError(t, err)

	for i, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTemplate)
		scenario.Path = "/good" + scenario.Path
		require.NoError(t, err)
		scenario.Group = fmt.Sprintf("good_spec_%d", i)
		// WHEN saving scenario to mock scenario repository
		err = repo.Save(scenario)
		// THEN it should save scenario
		require.NoError(t, err)
		// WITH mock web client
		client, data := buildJobsTestClient("AC1234567890", "RUNNING", "/good", scenario.Response.StatusCode)
		contractReq.Verbose = true
		contractReq.Overrides = data
		// AND executor
		executor := NewExecutor(repo, client)
		// AND should return saved scenario
		saved, err := repo.Lookup(scenario.ToKeyData(), nil)
		require.NoError(t, err)
		if scenario.Response.StatusCode != saved.Response.StatusCode {
			t.Fatalf("unexpected status %d != %d", scenario.Response.StatusCode, saved.Response.StatusCode)
		}
		// WHEN executing scenario
		res := executor.Execute(context.Background(), saved.ToKeyData(), dataTemplate, contractReq)
		for _, err := range res.Errors {
			t.Log(err)
		}
		// THEN it should succeed
		require.Equal(t, 0, len(res.Errors), fmt.Sprintf("spec %d == %v", i, res.Errors))
	}
}

func Test_ShouldParseRequestBody(t *testing.T) {
	scenario := &types.MockScenario{}
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
