package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/stretchr/testify/require"
)

func Test_ShouldLookupPutMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	player := NewPlayer(scenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, scenarioRepository.Save(buildScenario(types.Put, fmt.Sprintf("todo_put_%d", i), "/api/todos/:id", i)))
	}
	u, err := url.Parse("https://jsonplaceholder.typicode.com/blah")
	// WHEN looking up non-existing API
	ctx := web.NewStubContext(
		&http.Request{
			Method: "PUT",
			URL:    u,
			Header: http.Header{MockWaitBeforeReply: []string{"1"}, MockResponseStatus: []string{"0"}},
		},
	)
	// THEN it should not find it
	err = player.Handle(ctx)
	require.Error(t, err)

	u, err = url.Parse("https://jsonplaceholder.typicode.com/api/todos/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by PUT with different query param
	ctx = web.NewStubContext(&http.Request{
		Method: "PUT",
		URL:    u,
		Header: http.Header{MockWaitBeforeReply: []string{"1"}, MockResponseStatus: []string{"0"}},
	})
	err = player.Handle(ctx)
	require.NoError(t, err)
	// THEN it should find it
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldLookupPostMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	player := NewPlayer(scenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, scenarioRepository.Save(buildScenario(types.Post, fmt.Sprintf("book_post_%d", i), "/api/:topic/books", i)))
	}

	u, err := url.Parse("https://books.com/api/scifi/books/13?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by POST with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "POST",
		URL:    u,
		Header: make(http.Header),
	})
	err = player.Handle(ctx)
	require.NoError(t, err)
	// THEN it should find it
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldLookupGetMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	player := NewPlayer(scenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, scenarioRepository.Save(buildScenario(types.Get, fmt.Sprintf("books_get_%d", i), "/api/books/:topic/:id", i)))
	}
	// WHEN looking up non-existing API
	u, err := url.Parse("https://books.com/v2/topic/business/blah")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{
		Method: "GET",
		URL:    u,
		Header: make(http.Header),
	})
	// THEN it should not find it
	err = player.Handle(ctx)
	require.Error(t, err)

	u, err = url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by PUT with different query param
	ctx = web.NewStubContext(&http.Request{
		Method: "GET",
		URL:    u,
		Header: make(http.Header),
	})
	err = player.Handle(ctx)
	require.NoError(t, err)
	// THEN it should find it
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldLookupDeleteMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository and player
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	player := NewPlayer(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Delete, fmt.Sprintf("books_delete_%d", i), "/api/books/:topic/:id", i)))
	}
	// WHEN looking up non-existing API
	u, err := url.Parse("https://books.com/business/books/202")
	ctx := web.NewStubContext(&http.Request{
		Method: "DELETE",
		URL:    u,
		Header: make(http.Header),
	})
	// THEN it should not find it
	err = player.Handle(ctx)
	require.Error(t, err)

	// WHEN looking up todos by PUT with different query param
	u, err = url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	ctx = web.NewStubContext(&http.Request{
		Method: "DELETE",
		URL:    u,
		Header: make(http.Header),
	})
	err = player.Handle(ctx)
	require.NoError(t, err)
	// THEN it should find it
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldLookupDeleteMockScenariosWithBraces(t *testing.T) {
	// GIVEN a mock scenario repository and player
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	player := NewPlayer(mockScenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(buildScenario(types.Delete, fmt.Sprintf("books_delete_%d", i), "/api/books/{topic}/{id}", i)))
	}
	// WHEN looking up non-existing API
	u, err := url.Parse("https://books.com/business/books/202")
	ctx := web.NewStubContext(&http.Request{
		Method: "DELETE",
		URL:    u,
		Header: make(http.Header),
	})
	// THEN it should not find it
	err = player.Handle(ctx)
	require.Error(t, err)

	// WHEN looking up todos by PUT with different query param
	u, err = url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	ctx = web.NewStubContext(&http.Request{
		Method: "DELETE",
		URL:    u,
		Header: make(http.Header),
	})
	err = player.Handle(ctx)
	require.NoError(t, err)
	// THEN it should find it
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldLookupPutMockScenariosWithBraces(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	player := NewPlayer(scenarioRepository, fixtureRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, scenarioRepository.Save(buildScenario(types.Put, fmt.Sprintf("todo_put_%d", i), "/api/todos/{id}", i)))
	}
	u, err := url.Parse("https://jsonplaceholder.typicode.com/blah")
	// WHEN looking up non-existing API
	ctx := web.NewStubContext(
		&http.Request{
			Method: "PUT",
			URL:    u,
			Header: make(http.Header),
		},
	)
	// THEN it should not find it
	err = player.Handle(ctx)
	require.Error(t, err)

	u, err = url.Parse("https://jsonplaceholder.typicode.com/api/todos/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by PUT with different query param
	ctx = web.NewStubContext(&http.Request{
		Method: "PUT",
		URL:    u,
		Header: make(http.Header),
	})
	err = player.Handle(ctx)
	require.NoError(t, err)
	// THEN it should find it
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
			MatchQueryParams: map[string]string{"a": `\d+`, "b": "abc"},
			MatchContentType: "application/json",
		},
		Response: types.MockHTTPResponse{
			Headers: map[string][]string{
				"ETag": {strconv.Itoa(n)},
			},
			ContentType: "application/json",
			Contents:    "test body",
			StatusCode:  200,
		},
		WaitBeforeReply: time.Duration(1) * time.Second,
	}
}
