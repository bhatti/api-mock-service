package proxy

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/stretchr/testify/require"
)

func Test_ShouldNotRecordWithoutMockURL(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	client.AddMapping("GET", "https://jsonplaceholder.typicode.com/todos/10", web.NewStubHTTPResponse(200, `
	{
		"userId": 1,
		"id": 10,
		"title": "illo est ratione doloremque quia maiores aut",
		"completed": true
	  }
	`))
	recorder := NewRecorder(client, mockScenarioRepository)
	ctx := web.NewStubContext(&http.Request{Method: "GET"})

	// WHEN invoking Handle without MockUrl
	err = recorder.Handle(ctx)

	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldRecordGetProxyRequests(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	body := strings.TrimSpace(`
	{
		"userId": 1,
		"id": 10,
		"title": "illo est ratione doloremque quia maiores aut",
		"completed": true
	  }
	`)
	client.AddMapping("GET", "https://jsonplaceholder.typicode.com/todos/10", web.NewStubHTTPResponse(200, body))
	recorder := NewRecorder(client, mockScenarioRepository)
	ctx := web.NewStubContext(&http.Request{
		Method: "GET",
		Header: map[string][]string{
			MockURL: {"https://jsonplaceholder.typicode.com/todos/10"},
		},
	})

	// WHEN invoking GET proxy API
	err = recorder.Handle(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, body, string(saved))
}

func Test_ShouldRecordDeleteProxyRequests(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	client.AddMapping("DELETE", "https://jsonplaceholder.typicode.com/todos/101", web.NewStubHTTPResponse(200, "{}"))
	recorder := NewRecorder(client, mockScenarioRepository)
	ctx := web.NewStubContext(&http.Request{
		Method: "DELETE",
		Header: map[string][]string{
			MockURL: {"https://jsonplaceholder.typicode.com/todos/101"},
		},
	})

	// WHEN invoking DELETE proxy API
	err = recorder.Handle(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "{}", string(saved))
}

func Test_ShouldRecordPostProxyRequests(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	reqBody := []byte(strings.TrimSpace(`{"userId": 101, "title": "Buy milk", "completed": False}`))
	resBody := strings.TrimSpace(`
{
  "{\"userId\": 101, \"title\": \"Buy milk\", \"completed\": False}": "",
  "id": 201
}
	`)
	client.AddMapping("PUT", "https://jsonplaceholder.typicode.com/todos/202", web.NewStubHTTPResponse(200, resBody))
	recorder := NewRecorder(client, mockScenarioRepository)
	reader := io.NopCloser(bytes.NewReader(reqBody))
	ctx := web.NewStubContext(&http.Request{
		Method: "PUT",
		Header: map[string][]string{
			MockURL: {"https://jsonplaceholder.typicode.com/todos/202"},
		},
		Body: reader,
	})

	// WHEN invoking POST proxy API
	err = recorder.Handle(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, resBody, string(saved))
}

func Test_ShouldRecordPutProxyRequests(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	reqBody := []byte(strings.TrimSpace(`{"id": 202, "userId": 505, "title": "Buy milk", "completed": False}`))
	resBody := strings.TrimSpace(`
{
  "{\"id\": 202, \"userId\": 505, \"title\": \"Buy milk\", \"completed\": False}": "",
  "id": 2
}
	`)
	client.AddMapping("POST", "https://jsonplaceholder.typicode.com/todos/2", web.NewStubHTTPResponse(200, resBody))
	recorder := NewRecorder(client, mockScenarioRepository)
	reader := io.NopCloser(bytes.NewReader(reqBody))
	ctx := web.NewStubContext(&http.Request{
		Method: "POST",
		Header: map[string][]string{
			MockURL: {"https://jsonplaceholder.typicode.com/todos/2"},
		},
		Body: reader,
	})

	// WHEN invoking PUT proxy API
	err = recorder.Handle(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, resBody, string(saved))
}
