package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/bhatti/api-mock-service/internal/proxy"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/stretchr/testify/require"
)

func Test_ShouldNotPlayNonExistingAPI(t *testing.T) {
	// GIVEN repository, player and controller for mock scenario
	_ = rootPathParams{}
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(buildTestConfig())
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(buildTestConfig())
	require.NoError(t, err)
	player := proxy.NewConsumerExecutor(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Delete, fmt.Sprintf("books_delete_%d", i), "/api/books/:topic/:id", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)

	u, err := url.Parse("https://jsonplaceholder.typicode.com/blah")
	// WHEN looking up non-existing API
	ctx := web.NewStubContext(
		&http.Request{
			Method: "PUT",
			URL:    u,
			Header: make(http.Header),
		},
	)
	err = ctrl.deleteRoot(ctx)

	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldPlayGetProxyRequests(t *testing.T) {
	// GIVEN repository, player and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(buildTestConfig())
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(buildTestConfig())
	require.NoError(t, err)
	player := proxy.NewConsumerExecutor(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Get, fmt.Sprintf("books_get_%d", i), "/api/books/:topic/:id", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by GET with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "GET",
		URL:    u,
		Header: make(http.Header),
	})

	// WHEN invoking GET proxy API
	err = ctrl.getRoot(ctx)
	// THEN it should fail without headers
	require.Error(t, err)

	// WHEN looking up todos by GET with header
	ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}, types.ContentTypeHeader: {"application/yaml"}}

	// WHEN invoking GET proxy API
	err = ctrl.getRoot(ctx)
	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldPlayDeleteProxyRequests(t *testing.T) {
	// GIVEN repository, player and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(buildTestConfig())
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(buildTestConfig())
	require.NoError(t, err)
	player := proxy.NewConsumerExecutor(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Delete, fmt.Sprintf("books_delete_%d", i), "/api/books/:topic/:id", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by PUT with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "DELETE",
		URL:    u,
		Header: make(http.Header),
	})
	ctx.Request().Header = map[string][]string{types.ContentTypeHeader: {"application/yaml"}, "Auth": {"01234567890"}}

	// WHEN invoking DELETE proxy API
	err = ctrl.deleteRoot(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldPlayPostProxyRequests(t *testing.T) {
	// GIVEN repository, player and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(buildTestConfig())
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(buildTestConfig())
	require.NoError(t, err)
	player := proxy.NewConsumerExecutor(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Post, fmt.Sprintf("books_post_%d", i), "/api/books/:topic", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business?a=12&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by POST with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "POST",
		URL:    u,
		Header: map[string][]string{types.ContentTypeHeader: {"application/yaml"}, "Auth": {"01234567890"}},
	})

	// WHEN invoking POST proxy API
	err = ctrl.postRoot(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldPlayPutProxyRequests(t *testing.T) {
	// GIVEN repository, player and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(buildTestConfig())
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(buildTestConfig())
	require.NoError(t, err)
	player := proxy.NewConsumerExecutor(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Put, fmt.Sprintf("books_put_%d", i), "/api/books/:topic/:id", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business/202?a=12&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by PUT with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "PUT",
		URL:    u,
		Header: map[string][]string{types.ContentTypeHeader: {"application/yaml"}, "Auth": {"01234567890"}},
	})

	// WHEN invoking PUT proxy API
	err = ctrl.putRoot(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldPlayConnectProxyRequests(t *testing.T) {
	// GIVEN repository, player and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(buildTestConfig())
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(buildTestConfig())
	require.NoError(t, err)
	player := proxy.NewConsumerExecutor(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Connect, fmt.Sprintf("books_Connect_%d", i), "/api/books/:topic/:id", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by Connect with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "Connect",
		URL:    u,
		Header: map[string][]string{types.ContentTypeHeader: {"application/yaml"}, "Auth": {"01234567890"}},
	})

	// WHEN invoking Connect proxy API
	err = ctrl.connectRoot(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldPlayHeadProxyRequests(t *testing.T) {
	// GIVEN repository, player and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(buildTestConfig())
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(buildTestConfig())
	require.NoError(t, err)
	player := proxy.NewConsumerExecutor(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Head, fmt.Sprintf("books_Head_%d", i), "/api/books/:topic/:id", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by Head with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "Head",
		URL:    u,
		Header: map[string][]string{types.ContentTypeHeader: {"application/yaml"}, "Auth": {"01234567890"}},
	})

	// WHEN invoking Head proxy API
	err = ctrl.headRoot(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldPlayOptionsProxyRequests(t *testing.T) {
	// GIVEN repository, player and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(buildTestConfig())
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(buildTestConfig())
	require.NoError(t, err)
	player := proxy.NewConsumerExecutor(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Options, fmt.Sprintf("books_Options_%d", i), "/api/books/:topic/:id", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by Options with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "Options",
		URL:    u,
		Header: map[string][]string{types.ContentTypeHeader: {"application/yaml"}, "Auth": {"01234567890"}},
	})

	// WHEN invoking Options proxy API
	err = ctrl.optionsRoot(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldPlayPatchProxyRequests(t *testing.T) {
	// GIVEN repository, player and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(buildTestConfig())
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(buildTestConfig())
	require.NoError(t, err)
	player := proxy.NewConsumerExecutor(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Patch, fmt.Sprintf("books_Patch_%d", i), "/api/books/:topic/:id", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by Patch with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "Patch",
		URL:    u,
		Header: map[string][]string{types.ContentTypeHeader: {"application/yaml"}, "Auth": {"01234567890"}},
	})

	// WHEN invoking Patch proxy API
	err = ctrl.patchRoot(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldPlayTraceProxyRequests(t *testing.T) {
	// GIVEN repository, player and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(buildTestConfig())
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(buildTestConfig())
	require.NoError(t, err)
	player := proxy.NewConsumerExecutor(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Trace, fmt.Sprintf("books_Trace_%d", i), "/api/books/:topic/:id", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by Trace with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "Trace",
		URL:    u,
		Header: map[string][]string{types.ContentTypeHeader: {"application/yaml"}, "Auth": {"01234567890"}},
	})

	// WHEN invoking Trace proxy API
	err = ctrl.traceRoot(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func buildTestConfig() *types.Configuration {
	return &types.Configuration{
		DataDir:                  "../../mock_tests",
		HistoryDir:               "../../mock_history",
		MaxHistory:               5,
		ProxyPort:                8081,
		AssertQueryParamsPattern: "target",
		AssertHeadersPattern:     "target",
	}
}

func buildScenario(method types.MethodType, name string, path string, n int) *types.MockScenario {
	return &types.MockScenario{
		Method:      method,
		Name:        name,
		Path:        path,
		Description: name,
		Request: types.MockHTTPRequest{
			AssertQueryParamsPattern: map[string]string{"a": `\d+`, "b": "abc"},
			AssertHeadersPattern: map[string]string{
				types.ContentTypeHeader: "application/(json|yaml)",
				"Auth":                  "[0-9a-z]{10}",
			},
		},
		Response: types.MockHTTPResponse{
			Headers: map[string][]string{
				"ETag":                  {strconv.Itoa(n)},
				types.ContentTypeHeader: {"application/json"},
			},
			Contents:   "test body",
			StatusCode: 200,
		},
		WaitBeforeReply: time.Duration(1) * time.Second,
	}
}
