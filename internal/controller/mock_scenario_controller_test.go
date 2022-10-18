package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
}

func Test_ShouldFailPostScenarioWithoutMethodNameOrPath(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockScenarioController(mockScenarioRepository, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario without method, name and path
	err = ctrl.postMockScenario(ctx)

	// THEN it should fail without name
	require.Error(t, err)
	ctx.Params["method"] = "POST"
	err = ctrl.postMockScenario(ctx)
	require.Error(t, err)

	// THEN it should fail without path
	require.Error(t, err)
	ctx.Params["name"] = "data1"
	err = ctrl.postMockScenario(ctx)
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
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN getting mock scenario without method, name and path
	err = ctrl.getMockScenario(ctx)

	// THEN it should fail without name
	require.Error(t, err)
	ctx.Params["method"] = "GET"
	err = ctrl.getMockScenario(ctx)
	require.Error(t, err)

	// THEN it should fail without path
	require.Error(t, err)
	ctx.Params["name"] = "data1"
	err = ctrl.getMockScenario(ctx)
	require.Error(t, err)
}

func Test_ShouldFailDeleteScenarioWithoutMethodNameOrPath(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockScenarioController(mockScenarioRepository, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN deleting mock scenario without method, name and path
	err = ctrl.deleteMockScenario(ctx)

	// THEN it should fail without name
	require.Error(t, err)
	ctx.Params["method"] = "DELETE"
	err = ctrl.deleteMockScenario(ctx)
	require.Error(t, err)

	// THEN it should fail
	require.Error(t, err)
	ctx.Params["name"] = "data1"
	err = ctrl.deleteMockScenario(ctx)
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
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario
	err = ctrl.postMockScenario(ctx)

	// THEN it should return saved scenario
	require.NoError(t, err)
	savedScenario := ctx.Result.(*types.MockScenario)
	require.NoError(t, scenario.Equals(savedScenario))

	// WHEN getting mock scenario by path
	ctx.Params["method"] = string(savedScenario.Method)
	ctx.Params["name"] = savedScenario.Name
	ctx.Params["path"] = savedScenario.NormalPath('/')
	err = ctrl.getMockScenario(ctx)

	// THEN it should not fail
	require.NoError(t, err)
	getScenario := ctx.Result.(*types.MockScenario)
	require.NoError(t, scenario.Equals(getScenario))
}

func Test_ShouldCreateAndGetMockNames(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockScenarioController(mockScenarioRepository, webServer)
	for i := 0; i < 10; i++ {
		scenario := buildScenario(types.Post, fmt.Sprintf("abc_%d", i), "/123/456", i)
		b, err := json.Marshal(scenario)
		require.NoError(t, err)
		reader := io.NopCloser(bytes.NewReader(b))
		ctx := web.NewStubContext(&http.Request{Body: reader})

		// WHEN creating mock scenario
		err = ctrl.postMockScenario(ctx)

		// THEN it should return saved scenario
		require.NoError(t, err)
	}

	// WHEN getting mock scenario by path
	ctx := web.NewStubContext(&http.Request{})
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
	ctx := web.NewStubContext(&http.Request{Body: reader})
	err = ctrl.postMockScenario(ctx)

	// THEN it should succeed
	require.NoError(t, err)
	savedScenario := ctx.Result.(*types.MockScenario)
	require.NoError(t, scenario.Equals(savedScenario))

	// WHEN deleting mock scenario
	ctx.Params["method"] = string(savedScenario.Method)
	ctx.Params["name"] = savedScenario.Name
	ctx.Params["path"] = savedScenario.NormalPath('/')
	err = ctrl.deleteMockScenario(ctx)
	// THEN it should succeed
	require.NoError(t, err)

	// AND get API should fail
	err = ctrl.getMockScenario(ctx)

	// THEN it should not fail
	require.Error(t, err)
}
