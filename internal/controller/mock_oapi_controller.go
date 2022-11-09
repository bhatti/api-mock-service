package controller

import (
	"context"
	"github.com/bhatti/api-mock-service/internal/oapi"
	"github.com/bhatti/api-mock-service/internal/utils"
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

	webserver.POST("/_oapi", ctrl.PostMockOAPIScenario)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// PostMockOAPIScenario handler
// swagger:route POST /_oapi open-api PostMockOAPIScenario
// Creates new mock scenarios based on Open API v3
// responses:
//
//	200: mockScenarioOAPIResponse
func (moc *MockOAPIController) PostMockOAPIScenario(c web.APIContext) (err error) {
	var data []byte
	data, c.Request().Body, err = utils.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
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

// swagger:parameters PostMockOAPIScenario
// The params for mock-scenario based on OpenAPI v3
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
