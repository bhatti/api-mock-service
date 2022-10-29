package controller

import (
	"context"
	"github.com/bhatti/api-mock-service/internal/oapi"
	"io"
	"net/http"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
)

// MockOAPIController structure
type MockOAPIController struct {
	mockScenarioRepository repository.MockScenarioRepository
}

// NewMockOAPIController instantiates controller for updating mock-scenarios based on OpenAPI v3
func NewMockOAPIController(
	mockScenarioRepository repository.MockScenarioRepository,
	webserver web.Server) *MockOAPIController {
	ctrl := &MockOAPIController{
		mockScenarioRepository: mockScenarioRepository,
	}

	webserver.POST("/_oapi", ctrl.postMockOAPIScenario)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// swagger:route POST /_oapi mock-scenarios postMockOAPIScenario
// Creates new mock scenarios based on Open API v3
// responses:
//
//	200: mockScenarioOAPIResponse
func (moc *MockOAPIController) postMockOAPIScenario(c web.APIContext) (err error) {
	data, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	_ = c.Request().Body.Close()
	specs, err := oapi.Parse(context.Background(), data)
	if err != nil {
		return err
	}
	scenarios := make([]*types.MockScenario, 0)
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario()
		if err != nil {
			return err
		}
		err = moc.mockScenarioRepository.Save(scenario)
		if err != nil {
			return err
		}
		scenarios = append(scenarios, scenario)
	}
	return c.JSON(http.StatusOK, scenarios)
}

// ********************************* Swagger types ***********************************

// swagger:parameters postMockOAPIScenario
// The params for mock-scenario
type mockScenarioOAPICreateParams struct {
	// in:body
	Body []byte
}

// MockScenario body for update
// swagger:response mockScenarioOAPIResponse
type mockScenarioOAPIResponseBody struct {
	// in:body
	Body types.MockScenario
}
