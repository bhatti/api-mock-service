package controller

import (
	"fmt"
	"net/http"
	"net/url"
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
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	player := proxy.NewPlayer(mockScenarioRepository, fixtureRepository)
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
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	player := proxy.NewPlayer(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Get, fmt.Sprintf("books_get_%d", i), "/api/books/:topic/:id", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business/202")
	require.NoError(t, err)
	// WHEN looking up todos by PUT with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "GET",
		URL:    u,
		Header: make(http.Header),
	})

	// WHEN invoking GET proxy API
	err = ctrl.getRoot(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldPlayDeleteProxyRequests(t *testing.T) {
	// GIVEN repository, player and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	player := proxy.NewPlayer(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Delete, fmt.Sprintf("books_delete_%d", i), "/api/books/:topic/:id", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business/202")
	require.NoError(t, err)
	// WHEN looking up todos by PUT with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "DELETE",
		URL:    u,
		Header: make(http.Header),
	})

	// WHEN invoking DELETE proxy API
	err = ctrl.deleteRoot(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldPlayPostProxyRequests(t *testing.T) {
	// GIVEN repository, player and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	player := proxy.NewPlayer(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Post, fmt.Sprintf("books_post_%d", i), "/api/books/:topic", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business")
	require.NoError(t, err)
	// WHEN looking up todos by POST with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "POST",
		URL:    u,
		Header: make(http.Header),
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
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	player := proxy.NewPlayer(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Put, fmt.Sprintf("books_put_%d", i), "/api/books/:topic/:id", i)))
	}
	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)
	u, err := url.Parse("https://books.com/api/books/topic/business/202")
	require.NoError(t, err)
	// WHEN looking up todos by PUT with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "PUT",
		URL:    u,
		Header: make(http.Header),
	})

	// WHEN invoking PUT proxy API
	err = ctrl.putRoot(ctx)

	// THEN it should return stubbed response
	require.NoError(t, err)
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func buildScenario(method types.MethodType, name string, path string, n int) *types.MockScenario {
	return &types.MockScenario{
		Method:      method,
		Name:        name,
		Path:        path,
		Description: name,
		Request: types.MockHTTPRequest{
			QueryParams: fmt.Sprintf("a=1&b=2&n=%d", n),
			ContentType: "application/json",
			Headers: map[string][]string{
				"ETag": {"981"},
			},
		},
		Response: types.MockHTTPResponse{
			Headers: map[string][]string{
				"ETag": {"123"},
			},
			ContentType: "application/json",
			Contents:    "test body",
			StatusCode:  200,
		},
		WaitBeforeReply: time.Duration(1) * time.Second,
	}
}
