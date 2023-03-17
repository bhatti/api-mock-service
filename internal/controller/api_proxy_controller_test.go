package controller

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/bhatti/api-mock-service/internal/proxy"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"

	"github.com/stretchr/testify/require"
)

func Test_ShouldNotRecordWithoutMockURL(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	client.AddMapping("GET", "https://jsonplaceholder.typicode.com/todos/10", web.NewStubHTTPResponse(200, `
	{
		"userId": 1,
		"id": 10,
		"title": "my test title1",
		"completed": true
	  }
	`))
	recorder := proxy.NewRecorder(config, client, mockScenarioRepository)
	webServer := web.NewStubWebServer()
	ctrl := NewAPIProxyController(recorder, webServer)
	ctx := web.NewStubContext(&http.Request{Method: "GET"})

	// WHEN invoking GET proxy API without MockUrl
	err = ctrl.getAPIProxy(ctx)

	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldRecordGetProxyRequests(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	body := strings.TrimSpace(`
	{
		"userId": 1,
		"id": 10,
		"title": "my test title 5",
		"completed": true
	  }
	`)
	client.AddMapping("GET", "https://jsonplaceholder.typicode.com/todos/10", web.NewStubHTTPResponse(200, body))
	recorder := proxy.NewRecorder(config, client, mockScenarioRepository)
	webServer := web.NewStubWebServer()
	ctrl := NewAPIProxyController(recorder, webServer)
	u, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{
		URL:    u,
		Method: "GET",
		Header: map[string][]string{
			types.MockURL: {"https://jsonplaceholder.typicode.com/todos/10"},
		},
	})

	// WHEN invoking GET proxy API
	err = ctrl.getAPIProxy(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, body, string(saved))
}

func Test_ShouldRecordDeleteProxyRequests(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	client := web.NewStubHTTPClient()
	client.AddMapping("DELETE", "https://jsonplaceholder.typicode.com/todos/101", web.NewStubHTTPResponse(200, "{}"))
	recorder := proxy.NewRecorder(config, client, mockScenarioRepository)
	webServer := web.NewStubWebServer()
	ctrl := NewAPIProxyController(recorder, webServer)
	u, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{
		Method: "DELETE",
		URL:    u,
		Header: map[string][]string{
			types.MockURL: {"https://jsonplaceholder.typicode.com/todos/101"},
		},
	})

	// WHEN invoking DELETE proxy API
	err = ctrl.deleteAPIProxy(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "{}", string(saved))
}

func Test_ShouldRecordPostProxyRequests(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
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
	recorder := proxy.NewRecorder(config, client, mockScenarioRepository)
	webServer := web.NewStubWebServer()
	ctrl := NewAPIProxyController(recorder, webServer)
	reader := io.NopCloser(bytes.NewReader(reqBody))
	u, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{
		Method: "PUT",
		URL:    u,
		Header: map[string][]string{
			types.MockURL: {"https://jsonplaceholder.typicode.com/todos/202"},
		},
		Body: reader,
	})

	// WHEN invoking POST proxy API
	err = ctrl.postAPIProxy(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, resBody, string(saved))
}

func Test_ShouldRecordPutProxyRequests(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
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
	recorder := proxy.NewRecorder(config, client, mockScenarioRepository)
	webServer := web.NewStubWebServer()
	ctrl := NewAPIProxyController(recorder, webServer)
	reader := io.NopCloser(bytes.NewReader(reqBody))
	u, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{
		Method: "POST",
		URL:    u,
		Header: map[string][]string{
			types.MockURL: {"https://jsonplaceholder.typicode.com/todos/2"},
		},
		Body: reader,
	})

	// WHEN invoking PUT proxy API
	err = ctrl.putAPIProxy(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, resBody, string(saved))
}
