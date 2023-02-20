package controller

import (
	"bytes"
	"embed"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"

	"github.com/stretchr/testify/require"
)

// docs holds our open-api specifications
//
//go:embed docs
var internalOAPI embed.FS

func Test_InitializeSwaggerStructsForMockOAPIScenarioController(t *testing.T) {
	_ = mockScenarioOAPICreateParams{}
	_ = mockScenarioOAPIResponseBody{}
	_ = mockOapiSpecIResponseBody{}
}

func Test_ShouldFailPostScenarioWithBadOAPIInput(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockOAPIController(internalOAPI, mockScenarioRepository, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario with without method, name and path
	err = ctrl.PostMockOAPIScenario(ctx)

	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldCreateTwitterMockScenarioFromOAPI(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockOAPIController(internalOAPI, mockScenarioRepository, webServer)
	b, err := os.ReadFile("../../fixtures/oapi/twitter.yaml")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario from Open API
	err = ctrl.PostMockOAPIScenario(ctx)

	// THEN it should return saved scenario
	require.NoError(t, err)
	arrScenarios := ctx.Result.([]*types.MockScenario)
	require.Equal(t, 112, len(arrScenarios))
}

func Test_ShouldCreatePetsMockScenarioFromOAPI(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockOAPIController(internalOAPI, mockScenarioRepository, webServer)
	b, err := os.ReadFile("../../fixtures/oapi/pets.yaml")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario from Open API
	err = ctrl.PostMockOAPIScenario(ctx)

	// THEN it should return saved scenario
	require.NoError(t, err)
	arrScenarios := ctx.Result.([]*types.MockScenario)
	require.Equal(t, 10, len(arrScenarios))
}

func Test_ShouldDownloadInternalSpecs(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockOAPIController(internalOAPI, mockScenarioRepository, webServer)
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN fetching open-api specs without group
	ctx.Params["group"] = "_internal"
	ctx.Request().URL, err = url.Parse("http://localhost:8080")
	require.NoError(t, err)
	// WHEN fetching open-api specs
	err = ctrl.GetOpenAPISpecsByGroup(ctx)
	// THEN it should return saved scenario
	require.NoError(t, err)
	blob := ctx.Result.([]byte)
	require.True(t, len(blob) > 0)
}

func Test_ShouldGetOpenAPIByGroup(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockOAPIController(internalOAPI, mockScenarioRepository, webServer)
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario from Open API
	err = ctrl.PostMockOAPIScenario(ctx)

	// THEN it should return saved scenario
	require.NoError(t, err)

	arrScenarios := ctx.Result.([]*types.MockScenario)
	// WHEN fetching open-api specs without group
	err = ctrl.GetOpenAPISpecsByGroup(ctx)
	//  THEN it should fail
	require.Error(t, err)
	ctx.Params["group"] = arrScenarios[0].Group
	ctx.Request().URL, err = url.Parse("http://localhost:8080")
	// WHEN fetching open-api specs
	err = ctrl.GetOpenAPISpecsByGroup(ctx)
	// THEN it should return saved scenario
	require.NoError(t, err)
	blob := ctx.Result.([]byte)
	require.True(t, len(blob) > 0)
}

func Test_ShouldGetOpenAPIByScenario(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockOAPIController(internalOAPI, mockScenarioRepository, webServer)
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario from Open API
	err = ctrl.PostMockOAPIScenario(ctx)

	// THEN it should return saved scenario
	require.NoError(t, err)

	arrScenarios := ctx.Result.([]*types.MockScenario)

	// WHEN fetching open-api specs without method, path, name
	err = ctrl.GetOpenAPISpecsByScenario(ctx)
	// THEN it should fail
	require.Error(t, err)

	ctx.Params["method"] = string(arrScenarios[0].Method)
	err = ctrl.GetOpenAPISpecsByScenario(ctx)
	// THEN it should fail
	require.Error(t, err)
	ctx.Params["path"] = arrScenarios[0].Path
	err = ctrl.GetOpenAPISpecsByScenario(ctx)
	// THEN it should fail
	require.Error(t, err)

	ctx.Params["name"] = arrScenarios[0].Name
	// WHEN fetching open-api specs
	err = ctrl.GetOpenAPISpecsByScenario(ctx)
	// THEN it should return saved scenario
	require.NoError(t, err)
	blob := ctx.Result.([]byte)
	require.True(t, len(blob) > 0)
}
