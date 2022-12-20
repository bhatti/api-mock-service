package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"

	"github.com/stretchr/testify/require"
)

func Test_InitializeSwaggerStructsForMockScenarioController(t *testing.T) {
	_ = mockNamesResponseBody{}
	_ = mockNamesParams{}
	_ = mockScenarioCreateParams{}
	_ = mockScenarioResponseBody{}
	_ = mockScenarioIDParams{}
	_ = mockScenarioPathsResponseBody{}
}

func Test_ShouldFailPostScenarioWithoutMethodNameOrPath(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockScenarioController(mockScenarioRepository, webServer)
	data := []byte("test data")
	reader := io.NopCloser(bytes.NewReader(data))
	u, err := url.Parse("http://localhost:8080?a=1&b=abc")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{Body: reader, URL: u})
	ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}}

	// WHEN creating mock scenario without method, name and path
	err = ctrl.PostMockScenario(ctx)

	// THEN it should fail without name
	require.Error(t, err)
	ctx.Params["method"] = "POST"
	err = ctrl.PostMockScenario(ctx)
	require.Error(t, err)

	// THEN it should fail without path
	require.Error(t, err)
	ctx.Params["name"] = "data1"
	err = ctrl.PostMockScenario(ctx)
	require.Error(t, err)
}

func Test_ShouldFailGetScenarioWithoutMethodNameOrPath(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockScenarioController(mockScenarioRepository, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	u, err := url.Parse("http://localhost:8080?a=1&b=abc")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{Body: reader, URL: u})
	ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}}

	// WHEN getting mock scenario without method, name and path
	err = ctrl.GetMockScenario(ctx)

	// THEN it should fail without name
	require.Error(t, err)
	ctx.Params["method"] = "GET"
	err = ctrl.GetMockScenario(ctx)
	require.Error(t, err)

	// THEN it should fail without path
	require.Error(t, err)
	ctx.Params["name"] = "data1"
	err = ctrl.GetMockScenario(ctx)
	require.Error(t, err)
}

func Test_ShouldFailGetScenarioNamesWithoutMethodNameOrPath(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockScenarioController(mockScenarioRepository, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	u, err := url.Parse("http://localhost:8080?a=1&b=abc")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{Body: reader, URL: u})
	ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}}

	// WHEN getting mock scenario without method, name and path
	err = ctrl.getMockNames(ctx)

	// THEN it should fail without name
	require.Error(t, err)
	ctx.Params["method"] = "GET"
	err = ctrl.getMockNames(ctx)
	require.Error(t, err)

	// THEN it should fail without path
	require.Error(t, err)
	ctx.Params["name"] = "data1"
	err = ctrl.getMockNames(ctx)
	require.Error(t, err)
}

func Test_ShouldFailDeleteScenarioWithoutMethodNameOrPath(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockScenarioController(mockScenarioRepository, webServer)
	data := []byte("test data")
	reader := io.NopCloser(bytes.NewReader(data))
	u, err := url.Parse("http://localhost:8080?a=1&b=abc")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{Body: reader, URL: u})
	ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}}

	// WHEN deleting mock scenario without method, name and path
	err = ctrl.DeleteMockScenario(ctx)

	// THEN it should fail without name
	require.Error(t, err)
	ctx.Params["method"] = "DELETE"
	err = ctrl.DeleteMockScenario(ctx)
	require.Error(t, err)

	// THEN it should fail
	require.Error(t, err)
	ctx.Params["name"] = "data1"
	err = ctrl.DeleteMockScenario(ctx)
	require.Error(t, err)
}

func Test_ShouldCreateAndGetMockScenario(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockScenarioController(mockScenarioRepository, webServer)
	scenario := buildScenario(types.Post, "test1", "/path1", 1)
	b, err := json.Marshal(scenario)
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	u, err := url.Parse("http://localhost:8080?a=1&b=abc")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{Body: reader, Method: string(scenario.Method), URL: u})
	ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}, "Content-Type": []string{"application/json"}}

	// WHEN creating mock scenario
	err = ctrl.PostMockScenario(ctx)

	// THEN it should return saved scenario
	require.NoError(t, err)
	savedScenario := ctx.Result.(*types.MockScenario)
	require.NoError(t, scenario.ToKeyData().Equals(savedScenario.ToKeyData()))

	// WHEN getting mock scenario by path
	ctx.Params["method"] = string(savedScenario.Method)
	ctx.Params["name"] = savedScenario.Name
	ctx.Params["path"] = "/path1"
	ctx.Params["a"] = "b"
	err = ctrl.GetMockScenario(ctx)

	// THEN it should not fail
	require.NoError(t, err)
	res := ctx.Result.(string)
	require.True(t, len(res) > 0)

	// AND it should not fail with yaml output
	ctx.Request().Header = map[string][]string{types.ContentTypeHeader: {"application/yaml"}, "Auth": {"01234567890"}}
	err = ctrl.GetMockScenario(ctx)
	// THEN it should not fail
	require.NoError(t, err)

	strScenario := ctx.Result.(string)
	require.Contains(t, strScenario, "method: POST")
}

func Test_ShouldCreateAndGetMockScenarioWithYAML(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockScenarioController(mockScenarioRepository, webServer)
	scenario := buildScenario(types.Post, "test1", "/path1", 1)
	b, err := yaml.Marshal(scenario)
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	u, err := url.Parse("http://localhost:8080?a=1&b=abc")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{Body: reader, Method: string(scenario.Method), URL: u})
	ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}, "Content-Type": []string{"application/yaml"}}

	// WHEN creating mock scenario
	err = ctrl.PostMockScenario(ctx)

	// THEN it should return saved scenario
	require.NoError(t, err)
	savedScenario := ctx.Result.(*types.MockScenario)
	require.Equal(t, "", string(savedScenario.Method))

	// WHEN getting mock scenario by path
	ctx.Params["method"] = string(scenario.Method)
	ctx.Params["name"] = scenario.Name
	ctx.Params["path"] = "/path1"
	ctx.Params["a"] = "b"
	err = ctrl.GetMockScenario(ctx)

	// THEN it should not fail
	require.NoError(t, err)
	str := ctx.Result.(string)
	require.True(t, len(str) > 0)
}

func Test_ShouldCreateAndGetMockNames(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockScenarioController(mockScenarioRepository, webServer)
	u, err := url.Parse("http://localhost:8080?a=1&b=abc")
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		scenario := buildScenario(types.Post, fmt.Sprintf("abc_%d", i), "/123/456", i)
		b, err := json.Marshal(scenario)
		require.NoError(t, err)
		reader := io.NopCloser(bytes.NewReader(b))
		ctx := web.NewStubContext(&http.Request{Body: reader, URL: u})
		ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}}

		// WHEN creating mock scenario
		err = ctrl.PostMockScenario(ctx)

		// THEN it should return saved scenario
		require.NoError(t, err)
	}

	// WHEN getting mock scenario by path
	ctx := web.NewStubContext(&http.Request{URL: u})
	ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}}
	ctx.Params["method"] = "Post"
	ctx.Params["path"] = "/123/456"
	err = ctrl.getMockNames(ctx)

	// THEN it should not fail
	require.NoError(t, err)
	names := ctx.Result.([]string)
	require.Equal(t, 10, len(names))
}

func Test_ShouldCreateAndDeleteMockScenario(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockScenarioController(mockScenarioRepository, webServer)

	// WHEN creating mock scenario
	scenario := buildScenario(types.Post, "test2", "/abc/123/456", 1)
	b, err := json.Marshal(scenario)
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	u, err := url.Parse("http://localhost:8080?a=1&b=abc")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{Body: reader, URL: u})
	ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}}
	err = ctrl.PostMockScenario(ctx)

	// THEN it should succeed
	require.NoError(t, err)
	savedScenario := ctx.Result.(*types.MockScenario)
	require.NoError(t, scenario.ToKeyData().Equals(savedScenario.ToKeyData()))

	// WHEN deleting mock scenario
	ctx.Params["method"] = string(savedScenario.Method)
	ctx.Params["name"] = savedScenario.Name
	ctx.Params["path"] = savedScenario.NormalPath('/')
	err = ctrl.DeleteMockScenario(ctx)
	// THEN it should succeed
	require.NoError(t, err)

	// AND get API should fail
	err = ctrl.GetMockScenario(ctx)

	// THEN it should not fail
	require.Error(t, err)
}

func Test_ShouldListMockScenario(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockScenarioController(mockScenarioRepository, webServer)

	// WHEN creating mock scenario
	scenario := buildScenario(types.Post, "test2", "/abc/123/456", 1)
	b, err := json.Marshal(scenario)
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	u, err := url.Parse("http://localhost:8080?a=1&b=abc")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{Body: reader, URL: u})
	ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}}
	err = ctrl.PostMockScenario(ctx)
	require.NoError(t, err)

	// THEN it should be able to list
	err = ctrl.ListMockScenarioPaths(ctx)
	require.NoError(t, err)
	scenarios := ctx.Result.(map[string]*types.MockScenarioKeyData)
	require.True(t, len(scenarios) > 0)
}
