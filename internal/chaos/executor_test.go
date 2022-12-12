package chaos

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

func Test_ShouldExecuteGetTodo(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/get_todo.yaml", repo)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	chaosReq := types.NewChaosRequest("https://jsonplaceholder.typicode.com", 1)
	// WHEN executing scenario
	executor := NewExecutor(repo, web.NewHTTPClient(&types.Configuration{DataDir: "../../mock_tests"}))
	// THEN it should execute saved scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, chaosReq)
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
	chaosReq := types.NewChaosRequest("https://jsonplaceholder.typicode.com", 1)

	// WHEN executing scenario
	executor := NewExecutor(repo, web.NewHTTPClient(&types.Configuration{DataDir: "../../mock_tests"}))
	// THEN it should execute saved scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, chaosReq)
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
	scenario.Response.Assertions = append(scenario.Response.Assertions, "VariableContains headers.Content-Type application/xjson")
	err = repo.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	chaosReq := types.NewChaosRequest("https://jsonplaceholder.typicode.com", 1)

	// WHEN executing scenario
	executor := NewExecutor(repo, web.NewHTTPClient(&types.Configuration{DataDir: "../../mock_tests"}))
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, chaosReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors[0].Error(), `failed to assert '{{VariableContains "headers.Content-Type"`)
}

func Test_ShouldNotExecutePutPostsWithBadHeaders(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/put_posts.yaml", repo)
	require.NoError(t, err)
	// AND bad matching header
	scenario.Response.MatchHeaders["Content-Type"] = "application/xjson"
	err = repo.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	chaosReq := types.NewChaosRequest("https://jsonplaceholder.typicode.com", 1)

	// AND executor
	executor := NewExecutor(repo, web.NewHTTPClient(&types.Configuration{DataDir: "../../mock_tests"}))

	// WHEN executing scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, chaosReq)
	// THEN it should not execute saved scenario
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors[0].Error(), `didn't match required header Content-Type with regex application/xjson`)
}

func Test_ShouldNotExecutePutPostsWithMissingHeaders(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND a valid scenario
	scenario, err := saveTestScenario("../../fixtures/put_posts.yaml", repo)
	require.NoError(t, err)
	// AND missing header
	scenario.Response.MatchHeaders["Abc-Content-Type"] = "application/xjson"
	err = repo.Save(scenario)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	chaosReq := types.NewChaosRequest("https://jsonplaceholder.typicode.com", 1)

	// WHEN executing scenario
	executor := NewExecutor(repo, web.NewHTTPClient(&types.Configuration{DataDir: "../../mock_tests"}))
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, chaosReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors[0].Error(), `failed to find required header Abc-Content-Type`)
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
	chaosReq := types.NewChaosRequest(baseURL, 1)
	// WHEN executing scenario
	executor := NewExecutor(repo, client)
	// THEN it should not execute saved scenario
	res := executor.Execute(context.Background(), scenario.ToKeyData(), dataTemplate, chaosReq)
	for _, err := range res.Errors {
		t.Log(err)
	}
	require.Equal(t, 1, len(res.Errors))
	require.Contains(t, res.Errors[0].Error(), `failed to assert '{{VariableContains "contents.id" "1"}}`)
}

func Test_ShouldExecuteJobsOpenAPIWithInvalidStatus(t *testing.T) {
	// GIVEN scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND mock scenarios from open-api specifications
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)

	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
	chaosReq := types.NewChaosRequest(baseURL, 5)
	specs, err := oapi.Parse(context.Background(), b, dataTemplate)
	require.NoError(t, err)

	// AND save specs
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTemplate)
		require.NoError(t, err)
		err = repo.Save(scenario)
		require.NoError(t, err)
	}
	// AND valid template for random data
	// AND mock web client
	client, data := buildJobsTestClient("AC1234567890", "BAD")
	chaosReq.Overrides = data
	chaosReq.Verbose = true
	// AND executor
	executor := NewExecutor(repo, client)
	// WHEN executing scenario
	res := executor.ExecuteByGroup(context.Background(), "v1_jobs", dataTemplate, chaosReq)
	for _, err := range res.Errors {
		t.Log(err)
		// THEN it should fail to execute
		require.Contains(t, err.Error(), "key 'jobStatus' - value 'BAD' didn't match regex")
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
	chaosReq := types.NewChaosRequest(baseURL, 5)
	specs, err := oapi.Parse(context.Background(), b, dataTemplate)
	require.NoError(t, err)

	// AND mock web client
	client, data := buildJobsTestClient("AC1234567890", "RUNNING")
	chaosReq.Overrides = data
	// AND executor
	executor := NewExecutor(repo, client)
	for _, spec := range specs {
		// WHEN saving scenario to mock scenario repository
		repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
		require.NoError(t, err)
		// THEN it should save scenario
		scenario, err := spec.BuildMockScenario(dataTemplate)
		require.NoError(t, err)
		err = repo.Save(scenario)
		require.NoError(t, err)
		// AND should return saved scenario
		saved, err := repo.Lookup(scenario.ToKeyData())
		require.NoError(t, err)

		// WHEN executing scenario
		res := executor.Execute(context.Background(), saved.ToKeyData(), dataTemplate, chaosReq)
		for _, err := range res.Errors {
			t.Log(err)
		}
		// THEN it should succeed
		require.Equal(t, 0, len(res.Errors), fmt.Sprintf("%v", res.Errors))
	}
}

func buildJobsTestClient(jobID string, jobStatus string) (web.HTTPClient, map[string]any) {
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
	client.AddMapping("GET", baseURL+"/v1/jobs/"+jobID, web.NewStubHTTPResponse(200, job))
	client.AddMapping("GET", baseURL+"/v1/jobs", web.NewStubHTTPResponse(200, `[`+job+`]`))
	client.AddMapping("POST", baseURL+"/v1/jobs", web.NewStubHTTPResponse(200, jobStatusReply))
	client.AddMapping("POST", baseURL+"/v1/jobs/"+jobID+"/cancel", web.NewStubHTTPResponse(200, jobStatusReply))
	client.AddMapping("POST", baseURL+"/v1/jobs/"+jobID+"/pause", web.NewStubHTTPResponse(200, jobStatusReply))
	client.AddMapping("POST", baseURL+"/v1/jobs/"+jobID+"/resume", web.NewStubHTTPResponse(200, jobStatusReply))
	client.AddMapping("POST", baseURL+"/v1/jobs/"+jobID+"/state", web.NewStubHTTPResponse(200, jobStatusReply))
	client.AddMapping("POST", baseURL+"/v1/jobs/"+jobID+"/state/"+jobStatus, web.NewStubHTTPResponse(200, jobStatusReply))
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
