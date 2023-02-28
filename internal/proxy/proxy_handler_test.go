package proxy

import (
	"bytes"
	"encoding/json"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func Test_ShouldNotStartProxyServer(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	config := &types.Configuration{DataDir: "../../mock_tests", ProxyPort: -1}
	handler := NewProxyHandler(config, web.NewAWSSigner(config), scenarioRepository, fixtureRepository, web.NewWebServerAdapter())
	require.Error(t, handler.Start())
}

func Test_ShouldNotHandleProxyRequestWithNotFoundError(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	u, err := url.Parse("http://localhost:8080/path2")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
	}
	config := &types.Configuration{DataDir: "../../mock_tests", ProxyPort: 8081}
	handler := NewProxyHandler(config, web.NewAWSSigner(config), scenarioRepository, fixtureRepository, web.NewWebServerAdapter())
	_, res := handler.handleRequest(req, nil)
	require.Nil(t, res)
}

func Test_ShouldNotHandleProxyRequestWithValidationError(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	u, err := url.Parse("http://localhost:8080/path?a=b")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
	}
	config := &types.Configuration{DataDir: "../../mock_tests", ProxyPort: 8081}
	handler := NewProxyHandler(config, web.NewAWSSigner(config), scenarioRepository, fixtureRepository, web.NewWebServerAdapter())
	_, res := handler.handleRequest(req, nil)
	require.Nil(t, res)
}

func Test_ShouldHandleProxyRequest(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	scenario := buildScenario(types.Post, "todos", "/api/todos", 0)
	require.NoError(t, scenarioRepository.Save(scenario))

	u, err := url.Parse("http://localhost:8080/api/todos?a=3&b=abc")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"application/json"}},
	}
	config := &types.Configuration{DataDir: "../../mock_tests", ProxyPort: 8081}
	handler := NewProxyHandler(config, web.NewAWSSigner(config), scenarioRepository, fixtureRepository, web.NewWebServerAdapter())
	_, res := handler.handleRequest(req, nil)
	require.NotNil(t, res)
}

func Test_ShouldHandleProxyRequestFixturesWithAdapter(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	scenario := buildScenario(types.Post, "todos", "/api/todos", 0)
	require.NoError(t, scenarioRepository.Save(scenario))

	adapter := web.NewWebServerAdapter()
	adapter.GET("/_fixtures/:method/fixtures/:path", adapterHandler)
	adapter.GET("/_fixtures/:method/:name/:path", adapterHandler)
	adapter.POST("/_fixtures/:method/:name/:path", adapterHandler)
	adapter.DELETE("/_fixtures/:method/:name/:path", adapterHandler)
	config := &types.Configuration{DataDir: "../../mock_tests", ProxyPort: 8081}
	proxy := NewProxyHandler(config, web.NewAWSSigner(config), scenarioRepository, fixtureRepository, adapter)

	methodPaths := map[string]string{
		"/_fixtures/POST/fixtures/my/path?a=1&b=1": "GET",
		"/_fixtures/POST/myname/my/path?a=2&b=1":   "GET",
		"/_fixtures/POST/myname/my/path?a=3&b=1":   "POST",
		"/_fixtures/POST/myname/my/path?a=4&b=1":   "DELETE",
	}
	for path, method := range methodPaths {
		u, err := url.Parse("http://localhost:8080" + path)
		require.NoError(t, err)
		req := &http.Request{
			URL:    u,
			Method: method,
			Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
		}
		_, res := proxy.handleRequest(req, nil)
		require.NotNil(t, res)
		params := make(map[string]string)
		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		err = json.Unmarshal(b, &params)
		require.NoError(t, err)
		require.Equal(t, "POST", params["http-method"])
		require.Equal(t, "my/path", params["http-path"])
		require.Equal(t, "1", params["query-b"])
		if strings.Contains(path, "myname") {
			require.Equal(t, "myname", params["http-name"])
		}
	}
}

func adapterHandler(c web.APIContext) error {
	res := make(map[string]string)
	res["url-path"] = c.Request().URL.Path
	res["http-path"] = c.Param("path")
	res["http-method"] = c.Param("method")
	res["http-name"] = c.Param("name")
	for k, v := range c.Request().URL.Query() {
		res["query-"+k] = v[0]
	}
	return c.JSON(200, res)
}

func Test_ShouldHandleProxyRequestScenariosWithAdapter(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	scenario := buildScenario(types.Post, "todos", "/api/todos", 0)
	require.NoError(t, scenarioRepository.Save(scenario))

	adapter := web.NewWebServerAdapter()
	adapter.GET("/_scenarios", adapterHandler)
	adapter.GET("/_scenarios/:method/names/:path", adapterHandler)
	adapter.GET("/_scenarios/:method/:name/:path", adapterHandler)
	adapter.POST("/_scenarios", adapterHandler)
	adapter.DELETE("/_scenarios/:method/:name/:path", adapterHandler)
	adapter.POST("/_oapi", adapterHandler)
	config := &types.Configuration{DataDir: "../../mock_tests", ProxyPort: 8081}
	proxy := NewProxyHandler(config, web.NewAWSSigner(config), scenarioRepository, fixtureRepository, adapter)

	methodPaths := map[string]string{
		"/_scenarios?a=1&b=1":                     "GET",
		"/_scenarios/POST/names/my/path?a=1&b=1":  "GET",
		"/_scenarios/POST/myname/my/path?a=2&b=1": "GET",
		"/_scenarios?a=3&b=1":                     "POST",
		"/_scenarios/POST/myname/my/path?a=4&b=1": "DELETE",
		"/_oapi?a=3&b=1":                          "POST",
	}
	for path, method := range methodPaths {
		u, err := url.Parse("http://localhost:8080" + path)
		require.NoError(t, err)
		req := &http.Request{
			URL:    u,
			Method: method,
			Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
		}
		_, res := proxy.handleRequest(req, nil)
		require.NotNil(t, res)
		params := make(map[string]string)
		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		err = json.Unmarshal(b, &params)
		require.NoError(t, err)
		if strings.Contains(path, "POST") {
			require.Equal(t, "POST", params["http-method"])
		}
		if strings.Contains(path, "my/path") {
			require.Equal(t, "my/path", params["http-path"])
		}
		if strings.Contains(path, "myname") {
			require.Equal(t, "myname", params["http-name"])
		}
		require.Equal(t, "1", params["query-b"])
	}
}

func Test_ShouldHandleProxyResponseWithoutRequestBody(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	u, err := url.Parse("http://localhost:8080/api/todos?a=b")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
	}
	res := &http.Response{
		Request: req,
		Header:  http.Header{},
	}
	config := &types.Configuration{DataDir: "../../mock_tests", ProxyPort: 8081}
	handler := NewProxyHandler(config, web.NewAWSSigner(config), scenarioRepository, fixtureRepository, web.NewWebServerAdapter())
	res = handler.handleResponse(res, nil)
	require.NotNil(t, res)
}

func Test_ShouldHandleProxyResponseWithoutResponse(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	config := &types.Configuration{DataDir: "../../mock_tests", ProxyPort: 8081}
	handler := NewProxyHandler(config, web.NewAWSSigner(config), scenarioRepository, fixtureRepository, web.NewWebServerAdapter())
	require.Nil(t, handler.handleResponse(nil, nil))
}

func Test_ShouldHandleProxyResponseWithoutRequest(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	res := &http.Response{}
	config := &types.Configuration{DataDir: "../../mock_tests", ProxyPort: 8081}
	handler := NewProxyHandler(config, web.NewAWSSigner(config), scenarioRepository, fixtureRepository, web.NewWebServerAdapter())
	res = handler.handleResponse(res, nil)
	require.NotNil(t, res)
}

func Test_ShouldHandleProxyResponseWithoutResponseBody(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	u, err := url.Parse("http://localhost:8080/api/todos?a=b")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
		Body:   io.NopCloser(bytes.NewReader([]byte("test"))),
	}
	res := &http.Response{
		Request: req,
		Header:  http.Header{},
	}
	config := &types.Configuration{DataDir: "../../mock_tests", ProxyPort: 8081}
	handler := NewProxyHandler(config, web.NewAWSSigner(config), scenarioRepository, fixtureRepository, web.NewWebServerAdapter())
	res = handler.handleResponse(res, nil)
	require.NotNil(t, res)
}

func Test_ShouldHandleProxyResponseWithRequestAndResponseBody(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	u, err := url.Parse("http://localhost:8080/api/todos?a=b")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
		Body:   io.NopCloser(bytes.NewReader([]byte("test"))),
	}
	res := &http.Response{
		Request: req,
		Body:    io.NopCloser(bytes.NewReader([]byte("test"))),
		Header:  http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
	}
	config := &types.Configuration{DataDir: "../../mock_tests", ProxyPort: 8081}
	handler := NewProxyHandler(config, web.NewAWSSigner(config), scenarioRepository, fixtureRepository, web.NewWebServerAdapter())
	res = handler.handleResponse(res, nil)
	require.NotNil(t, res)
	req.Header[types.MockRecordMode] = []string{types.MockRecordModeDisabled}
	res = handler.handleResponse(res, nil)
	require.NotNil(t, res)
}
