package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
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
		"title": "my test title 2",
		"completed": true
	  }
	`))

	recorder := NewRecorder(&types.Configuration{ProxyPort: 8081}, client, mockScenarioRepository)
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
		"title": "my test title 4",
		"completed": true
	  }
	`)
	client.AddMapping("GET", "https://jsonplaceholder.typicode.com/todos/10", web.NewStubHTTPResponse(200, body))
	recorder := NewRecorder(&types.Configuration{ProxyPort: 8081}, client, mockScenarioRepository)
	u, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{
		Method: "GET",
		URL:    u,
		Header: map[string][]string{
			types.MockURL: {"https://jsonplaceholder.typicode.com/todos/10"},
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
	recorder := NewRecorder(&types.Configuration{ProxyPort: 8081}, client, mockScenarioRepository)
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
	recorder := NewRecorder(&types.Configuration{ProxyPort: 8081}, client, mockScenarioRepository)
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
	recorder := NewRecorder(&types.Configuration{ProxyPort: 8081}, client, mockScenarioRepository)
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
	err = recorder.Handle(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, resBody, string(saved))
}

func Test_ShouldSaveMockResponse(t *testing.T) {
	// GIVEN a mock scenario repository
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	u, err := url.Parse("http://localhost:8080/path?a=b&target=2")
	require.NoError(t, err)

	resHeaders := http.Header{"target": []string{"val1"}, types.ContentTypeHeader: []string{"json"}}
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: resHeaders,
	}
	_, err = saveMockResponse(
		&types.Configuration{ProxyPort: 8081, MatchQueryRegex: "target", MatchHeaderRegex: "target"}, u, req, []byte("test"), []byte("test"),
		resHeaders, 404, mockScenarioRepository)
	require.NoError(t, err)
}

func Test_ShouldRecordRealPostProxyRequests(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	config := &types.Configuration{DataDir: "../../mock_tests"}
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(config)
	require.NoError(t, err)
	client := web.NewHTTPClient(config)
	reqBody := []byte(`{ "userId": 1, "id": 1, "title": "sunt aut", "body": "quia et rem eveniet architecto" }`)
	recorder := NewRecorder(&types.Configuration{ProxyPort: 8081}, client, mockScenarioRepository)
	reader := io.NopCloser(bytes.NewReader(reqBody))
	u, err := url.Parse("https://jsonplaceholder.typicode.com/posts")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{
		Method: "POST",
		URL:    u,
		Header: map[string][]string{
			"X-Mock-Url": {"https://jsonplaceholder.typicode.com/posts"},
		},
		Body: reader,
	})
	// WHEN invoking POST proxy API
	err = recorder.Handle(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Contains(t, string(saved), "id")
}
