package controller

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"

	"github.com/stretchr/testify/require"
)

func Test_InitializeSwaggerStructsForMockOAPIScenarioController(t *testing.T) {
	_ = mockScenarioOAPICreateParams{}
	_ = mockScenarioOAPIResponseBody{}
}

func Test_ShouldFailPostScenarioWithBadOAPIInput(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockOAPIController(mockScenarioRepository, webServer)
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
	ctrl := NewMockOAPIController(mockScenarioRepository, webServer)
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
	ctrl := NewMockOAPIController(mockScenarioRepository, webServer)
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
