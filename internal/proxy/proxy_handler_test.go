package proxy

import (
	"bytes"
	"encoding/json"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/elazarl/goproxy"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

func Test_ShouldNotStartProxyServer(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)
	config.ProxyPort = -1
	handler := NewProxyHandler(config,
		web.NewAuthAdapter(config), scenarioRepository, fixtureRepository, groupConfigRepository, web.NewWebServerAdapter())
	require.Error(t, handler.Start())
}

func Test_ShouldNotHandleProxyRequestWithNotFoundError(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	u, err := url.Parse("http://localhost:8080/path2")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
	}
	handler := NewProxyHandler(config,
		web.NewAuthAdapter(config), scenarioRepository, fixtureRepository, groupConfigRepository, web.NewWebServerAdapter())
	_, res := handler.handleRequest(req, &goproxy.ProxyCtx{})
	require.Nil(t, res)
}

func Test_ShouldNotHandleProxyRequestWithValidationError(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	u, err := url.Parse("http://localhost:8080/path?a=b")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
	}
	handler := NewProxyHandler(config,
		web.NewAuthAdapter(config), scenarioRepository, fixtureRepository, groupConfigRepository, web.NewWebServerAdapter())
	_, res := handler.handleRequest(req, &goproxy.ProxyCtx{})
	require.Nil(t, res)
}

func Test_ShouldHandleProxyRequest(t *testing.T) {
	config := types.BuildTestConfig()
	config.Debug = true
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)
	err = groupConfigRepository.Save("my-group5", &types.GroupConfig{ChaosEnabled: true})
	require.NoError(t, err)

	scenario := types.BuildTestScenario(types.Post, "todos", "/vabc5/api/todos", 0)
	scenario.Group = "my-group5"
	require.NoError(t, scenarioRepository.Save(scenario))

	u, err := url.Parse("http://localhost:8080/vabc5/api/todos?a=3&b=abc")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{
			"X1":                    []string{"val1"},
			types.ETagHeader:        []string{"123"},
			types.ContentTypeHeader: []string{"application/json"}},
	}
	handler := NewProxyHandler(config,
		web.NewAuthAdapter(config), scenarioRepository, fixtureRepository, groupConfigRepository, web.NewWebServerAdapter())
	_, res := handler.handleRequest(req, &goproxy.ProxyCtx{})
	require.NotNil(t, res)
	_, res = handler.handleRequest(req, &goproxy.ProxyCtx{})
	require.NotNil(t, res)
}

func Test_ShouldHandleProxyRequestFixturesWithAdapter(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	scenario := types.BuildTestScenario(types.Post, "todos", "/api/todos", 0)
	require.NoError(t, scenarioRepository.Save(scenario))

	adapter := web.NewWebServerAdapter()
	adapter.GET("/_fixtures/:method/fixtures/:path", adapterHandler)
	adapter.GET("/_fixtures/:method/:name/:path", adapterHandler)
	adapter.POST("/_fixtures/:method/:name/:path", adapterHandler)
	adapter.DELETE("/_fixtures/:method/:name/:path", adapterHandler)
	proxy := NewProxyHandler(config,
		web.NewAuthAdapter(config), scenarioRepository, fixtureRepository, groupConfigRepository, adapter)

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
		_, res := proxy.handleRequest(req, &goproxy.ProxyCtx{})
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
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	scenario := types.BuildTestScenario(types.Post, "todos", "/api/todos", 0)
	require.NoError(t, scenarioRepository.Save(scenario))

	adapter := web.NewWebServerAdapter()
	adapter.GET("/_scenarios", adapterHandler)
	adapter.GET("/_scenarios/:method/names/:path", adapterHandler)
	adapter.GET("/_scenarios/:method/:name/:path", adapterHandler)
	adapter.POST("/_scenarios", adapterHandler)
	adapter.DELETE("/_scenarios/:method/:name/:path", adapterHandler)
	adapter.POST("/_oapi", adapterHandler)
	proxy := NewProxyHandler(config,
		web.NewAuthAdapter(config), scenarioRepository, fixtureRepository, groupConfigRepository, adapter)

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
		_, res := proxy.handleRequest(req, &goproxy.ProxyCtx{})
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
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
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
	handler := NewProxyHandler(config,
		web.NewAuthAdapter(config), scenarioRepository, fixtureRepository, groupConfigRepository, web.NewWebServerAdapter())
	res = handler.handleResponse(res, &goproxy.ProxyCtx{})
	require.NotNil(t, res)
}

func Test_ShouldHandleProxyResponseWithoutResponse(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	handler := NewProxyHandler(config,
		web.NewAuthAdapter(config), scenarioRepository, fixtureRepository, groupConfigRepository, web.NewWebServerAdapter())
	require.Nil(t, handler.handleResponse(nil, &goproxy.ProxyCtx{}))
}

func Test_ShouldHandleProxyResponseWithoutRequest(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	res := &http.Response{}
	handler := NewProxyHandler(config,
		web.NewAuthAdapter(config), scenarioRepository, fixtureRepository, groupConfigRepository, web.NewWebServerAdapter())
	res = handler.handleResponse(res, &goproxy.ProxyCtx{})
	require.NotNil(t, res)
}

func Test_ShouldHandleProxyCondition(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
	require.NoError(t, err)

	handler := NewProxyHandler(config,
		web.NewAuthAdapter(config), scenarioRepository, fixtureRepository, groupConfigRepository, web.NewWebServerAdapter())
	proxyCond := handler.proxyCondition()
	u, err := url.Parse("https://localhost:8080")
	require.NoError(t, err)
	require.True(t, proxyCond(&http.Request{URL: u}, nil))
	u, err = url.Parse("https://localhost:8080/index.html")
	require.NoError(t, err)
	require.False(t, proxyCond(&http.Request{URL: u}, nil))
	u, err = url.Parse("https://localhost:8080/index.txt")
	require.NoError(t, err)
	require.False(t, proxyCond(&http.Request{URL: u}, nil))
	config.ProxyURLFilter = "(abc|123)"
	u, err = url.Parse("https://localhost:8080")
	require.NoError(t, err)
	require.False(t, proxyCond(&http.Request{URL: u}, nil))
	u, err = url.Parse("https://localhost:8080/abcd")
	require.NoError(t, err)
	require.True(t, proxyCond(&http.Request{URL: u}, nil))
	u, err = url.Parse("https://localhost:8080/1234")
	require.NoError(t, err)
	require.True(t, proxyCond(&http.Request{URL: u}, nil))
	u, err = url.Parse("https://localhost:8080/234")
	require.NoError(t, err)
	require.False(t, proxyCond(&http.Request{URL: u}, nil))
}

func Test_SaveProxyCert(t *testing.T) {
	err := saveProxyCert()
	require.NoError(t, err)
	err = os.Remove("cert.pem")
	require.NoError(t, err)
}

func Test_ShouldHandleProxyResponseWithoutResponseBody(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
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
	handler := NewProxyHandler(config,
		web.NewAuthAdapter(config), scenarioRepository, fixtureRepository, groupConfigRepository, web.NewWebServerAdapter())
	res = handler.handleResponse(res, &goproxy.ProxyCtx{})
	require.NotNil(t, res)
}

func Test_ShouldHandleProxyResponseWithRequestAndResponseBody(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(types.BuildTestConfig())
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
	handler := NewProxyHandler(config,
		web.NewAuthAdapter(config), scenarioRepository, fixtureRepository, groupConfigRepository, web.NewWebServerAdapter())
	res = handler.handleResponse(res, &goproxy.ProxyCtx{})
	require.NotNil(t, res)
	req.Header[types.MockRecordMode] = []string{types.MockRecordModeDisabled}
	res = handler.handleResponse(res, &goproxy.ProxyCtx{})
	require.NotNil(t, res)
}
