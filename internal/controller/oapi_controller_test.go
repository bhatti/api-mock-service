package controller

import (
	"bytes"
	"embed"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

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
	_ = apiScenarioOAPICreateParams{}
	_ = apiScenarioOAPIResponseBody{}
	_ = apiOapiSpecIResponseBody{}
	_ = getOpenAPISpecsByGroupParams{}
	_ = getOpenAPISpecsByHistoryParams{}
	_ = getOpenAPISpecsByScenarioParams{}
}

func Test_ShouldFailPostScenarioWithBadOAPIInput(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	oapiRepository, err := repository.NewFileOAPIRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewOAPIController(internalOAPI, mockScenarioRepository, oapiRepository, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario with without method, name and path
	err = ctrl.postMockOAPIScenario(ctx)

	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldCreateTwitterMockScenarioFromOAPI(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	oapiRepository, err := repository.NewFileOAPIRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewOAPIController(internalOAPI, mockScenarioRepository, oapiRepository, webServer)
	b, err := os.ReadFile("../../fixtures/oapi/twitter.yaml")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario from Open API
	err = ctrl.postMockOAPIScenario(ctx)

	// THEN it should return saved scenario
	require.NoError(t, err)
	arrScenarios := ctx.Result.([]*types.APIKeyData)
	require.Equal(t, 112, len(arrScenarios))
}

func Test_ShouldCreatePetsMockScenarioFromOAPI(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	oapiRepository, err := repository.NewFileOAPIRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewOAPIController(internalOAPI, mockScenarioRepository, oapiRepository, webServer)
	b, err := os.ReadFile("../../fixtures/oapi/pets.yaml")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario from Open API
	err = ctrl.postMockOAPIScenario(ctx)

	// THEN it should return saved scenario
	require.NoError(t, err)
	arrScenarios := ctx.Result.([]*types.APIKeyData)
	require.Equal(t, 10, len(arrScenarios))
}

func Test_ShouldDownloadScenarioHistory(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	oapiRepository, err := repository.NewFileOAPIRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewOAPIController(internalOAPI, mockScenarioRepository, oapiRepository, webServer)
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	u, err := url.Parse("http://localhost:8080?a=1&b=abc")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	scenario := buildScenario(types.Post, "test1", "/path1", 1)
	err = mockScenarioRepository.SaveHistory(scenario, u.String(), time.Now(), time.Now())
	require.NoError(t, err)

	names := mockScenarioRepository.HistoryNames(scenario.Group)
	require.True(t, len(names) > 0)
	// WHEN fetching open-api specs without name
	ctx.Request().URL, err = url.Parse("http://localhost:8080")
	require.NoError(t, err)
	err = ctrl.getOpenAPISpecsByHistory(ctx)
	// THEN it should fail
	require.Error(t, err)

	// WHEN fetching open-api specs with name
	ctx.Params["name"] = names[0]
	err = ctrl.getOpenAPISpecsByHistory(ctx)
	// THEN it should return saved scenario
	require.NoError(t, err)
	blob := ctx.Result.([]byte)
	require.True(t, len(blob) > 0)
}

func Test_ShouldDownloadInternalSpecs(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	oapiRepository, err := repository.NewFileOAPIRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewOAPIController(internalOAPI, mockScenarioRepository, oapiRepository, webServer)
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN fetching open-api specs without group
	ctx.Params["group"] = "_internal"
	ctx.Request().URL, err = url.Parse("http://localhost:8080")
	require.NoError(t, err)
	// WHEN fetching open-api specs
	err = ctrl.getOpenAPISpecsByGroup(ctx)
	// THEN it should return saved scenario
	require.NoError(t, err)
	blob := ctx.Result.([]byte)
	require.True(t, len(blob) > 0)
}

func Test_ShouldGetOpenAPIByGroup(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	oapiRepository, err := repository.NewFileOAPIRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewOAPIController(internalOAPI, mockScenarioRepository, oapiRepository, webServer)
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	ctx := web.NewStubContext(&http.Request{
		Body: reader,
	})
	ctx.Request().URL, err = url.Parse("http://localhost:8080")

	// WHEN creating mock scenario from Open API
	err = ctrl.postMockOAPIScenario(ctx)

	// THEN it should return saved scenario
	require.NoError(t, err)

	arrScenarios := ctx.Result.([]*types.APIKeyData)
	// WHEN fetching open-api specs without group
	err = ctrl.getOpenAPISpecsByGroup(ctx)
	//  THEN it should not fail
	require.NoError(t, err)
	ctx.Params["group"] = arrScenarios[0].Group
	// WHEN fetching open-api specs
	err = ctrl.getOpenAPISpecsByGroup(ctx)
	// THEN it should return saved scenario
	require.NoError(t, err)
	blob := ctx.Result.([]byte)
	require.True(t, len(blob) > 0)

	ctx.Params["group"] = "unknown"
	// WHEN fetching open-api specs
	err = ctrl.getOpenAPISpecsByGroup(ctx)
	// THEN it should return saved scenario
	require.NoError(t, err)
	blob = ctx.Result.([]byte)
	require.True(t, len(blob) > 0)
}

func Test_ShouldGetOpenAPIByScenario(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	oapiRepository, err := repository.NewFileOAPIRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewOAPIController(internalOAPI, mockScenarioRepository, oapiRepository, webServer)
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock scenario from Open API
	err = ctrl.postMockOAPIScenario(ctx)

	// THEN it should return saved scenario
	require.NoError(t, err)

	arrScenarios := ctx.Result.([]*types.APIKeyData)

	// WHEN fetching open-api specs without method, path, name
	err = ctrl.getOpenAPISpecsByScenario(ctx)
	// THEN it should fail
	require.Error(t, err)

	ctx.Params["method"] = string(arrScenarios[0].Method)
	err = ctrl.getOpenAPISpecsByScenario(ctx)
	// THEN it should fail
	require.Error(t, err)
	ctx.Params["path"] = arrScenarios[0].Path
	err = ctrl.getOpenAPISpecsByScenario(ctx)
	// THEN it should fail
	require.Error(t, err)

	ctx.Params["name"] = arrScenarios[0].Name
	// WHEN fetching open-api specs
	err = ctrl.getOpenAPISpecsByScenario(ctx)
	// THEN it should return saved scenario
	require.NoError(t, err)
	blob := ctx.Result.([]byte)
	require.True(t, len(blob) > 0)
}
