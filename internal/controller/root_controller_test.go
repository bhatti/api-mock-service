package controller

import (
	"fmt"
	"github.com/bhatti/api-mock-service/internal/contract"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/stretchr/testify/require"
)

func Test_ShouldNotPlayNonExistingAPI(t *testing.T) {
	// GIVEN repository, consumerExecutor and controller for mock scenario
	_ = rootPathParams{}
	config := types.BuildTestConfig()
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	player := contract.NewConsumerExecutor(config, mockScenarioRepository, fixtureRepository, groupConfigRepository)
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

func Test_ShouldPlayProxyRequestsByMethod(t *testing.T) {
	cases := []struct {
		method     types.MethodType
		httpMethod string
		path       string
		handler    string
	}{
		{types.Get, "GET", "/api/books/:topic/:id", "getRoot"},
		{types.Delete, "DELETE", "/api/books/:topic/:id", "deleteRoot"},
		{types.Post, "POST", "/api/books/:topic", "postRoot"},
		{types.Put, "PUT", "/api/books/:topic/:id", "putRoot"},
		{types.Connect, "Connect", "/api/books/:topic/:id", "connectRoot"},
		{types.Head, "Head", "/api/books/:topic/:id", "headRoot"},
		{types.Options, "Options", "/api/books/:topic/:id", "optionsRoot"},
		{types.Patch, "Patch", "/api/books/:topic/:id", "patchRoot"},
		{types.Trace, "Trace", "/api/books/:topic/:id", "traceRoot"},
	}

	for _, tc := range cases {
		t.Run(tc.httpMethod, func(t *testing.T) {
			config := types.BuildTestConfig()
			mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
			require.NoError(t, err)
			fixtureRepository, err := repository.NewFileFixtureRepository(config)
			require.NoError(t, err)
			groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
			require.NoError(t, err)
			player := contract.NewConsumerExecutor(config, mockScenarioRepository, fixtureRepository, groupConfigRepository)
			for i := 0; i < 3; i++ {
				require.NoError(t, mockScenarioRepository.Save(buildScenario(tc.method, fmt.Sprintf("books_%s_%d", tc.httpMethod, i), tc.path, i)))
			}
			webServer := web.NewStubWebServer()
			ctrl := NewRootController(player, webServer)

			urlStr := "https://books.com/api/books/topic/business/202?a=123&b=abc"
			if tc.method == types.Post {
				urlStr = "https://books.com/api/books/topic/business?a=12&b=abc"
			}
			u, err := url.Parse(urlStr)
			require.NoError(t, err)

			ctx := web.NewStubContext(&http.Request{
				Method: tc.httpMethod,
				URL:    u,
				Header: map[string][]string{types.ContentTypeHeader: {"application/yaml"}, "Auth": {"01234567890"}},
			})

			handlerFunc := map[string]func(web.APIContext) error{
				"getRoot":     ctrl.getRoot,
				"deleteRoot":  ctrl.deleteRoot,
				"postRoot":    ctrl.postRoot,
				"putRoot":     ctrl.putRoot,
				"connectRoot": ctrl.connectRoot,
				"headRoot":    ctrl.headRoot,
				"optionsRoot": ctrl.optionsRoot,
				"patchRoot":   ctrl.patchRoot,
				"traceRoot":   ctrl.traceRoot,
			}[tc.handler]

			err = handlerFunc(ctx)
			require.NoError(t, err)
			saved := ctx.Result.([]byte)
			require.Equal(t, "test body", string(saved))
		})
	}
}

func buildScenario(method types.MethodType, name string, path string, n int) *types.APIScenario {
	return &types.APIScenario{
		Method:      method,
		Name:        name,
		Path:        path,
		Description: name,
		Group:       "root-group",
		Request: types.APIRequest{
			AssertQueryParamsPattern: map[string]string{"a": `\d+`, "b": "abc"},
			AssertHeadersPattern: map[string]string{
				types.ContentTypeHeader: "application/(json|yaml)",
				"Auth":                  "[0-9a-z]{10}",
			},
		},
		Response: types.APIResponse{
			Headers: map[string][]string{
				types.ETagHeader:        {strconv.Itoa(n)},
				types.ContentTypeHeader: {"application/json"},
			},
			Contents:   "test body",
			StatusCode: 200,
		},
		WaitBeforeReply: 0,
	}
}
