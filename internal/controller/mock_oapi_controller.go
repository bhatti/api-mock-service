package controller

import (
	"context"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/oapi"
	"github.com/bhatti/api-mock-service/internal/utils"
	"gopkg.in/yaml.v3"
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

	webserver.GET("/_oapi/:group", ctrl.GetOpenAPISpecsByGroup)
	webserver.GET("/_oapi/:method/:name/:path", ctrl.GetOpenAPISpecsByScenario)
	webserver.POST("/_oapi", ctrl.PostMockOAPIScenario)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// GetOpenAPISpecsByGroup handler
// swagger:route GET /_oapi/{group} open-api GetOpenAPISpecsByGroup
// Generates OpenAPI specs for the scenario group
// responses:
//
//	200: mockOapiSpecIResponse
func (moc *MockOAPIController) GetOpenAPISpecsByGroup(c web.APIContext) (err error) {
	group := c.Param("group")
	if group == "" {
		return fmt.Errorf("scenario group not specified")
	}
	allByGroup := moc.mockScenarioRepository.LookupAllByGroup(group)
	var scenarios []*types.MockScenario
	for _, keyData := range allByGroup {
		scenario, err := moc.getScenario(keyData)
		if err != nil {
			return err
		}
		scenarios = append(scenarios, scenario)
	}
	b, err := oapi.MarshalScenarioToOpenAPI(group, "", scenarios...)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, "application/json", b)
}

// GetOpenAPISpecsByScenario handler
// swagger:route GET /_oapi/{method}/{name}/{path} open-apiGetOpenAPISpecsByScenario
// Generates OpenAPI specs for the scenario
// responses:
//
//	200: mockOapiSpecIResponse
func (moc *MockOAPIController) GetOpenAPISpecsByScenario(c web.APIContext) (err error) {
	method, err := types.ToMethod(c.Param("method"))
	if err != nil {
		return fmt.Errorf("method name not specified in %s due to %w", c.Request().URL, err)
	}
	name := c.Param("name")
	if name == "" {
		return fmt.Errorf("scenario name not specified in %s", c.Request().URL)
	}
	path := c.Param("path")
	if path == "" {
		return fmt.Errorf("path not specified in %s", c.Request().URL)
	}
	keyData := &types.MockScenarioKeyData{
		Method:           method,
		Name:             name,
		Path:             path,
		MatchQueryParams: make(map[string]string),
	}
	scenario, err := moc.getScenario(keyData)
	if err != nil {
		return err
	}
	b, err := oapi.MarshalScenarioToOpenAPI("", "", scenario)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, "application/json", b)
}

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
	dataTempl := fuzz.NewDataTemplateRequest(true, 1, 1)
	specs, err := oapi.Parse(context.Background(), data, dataTempl)
	if err != nil {
		return err
	}
	scenarios := make([]*types.MockScenario, 0)
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
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

// MockScenario body for update
// swagger:response mockOapiSpecIResponse
type mockOapiSpecIResponseBody struct {
	// in:body
	Body []byte
}

func (moc *MockOAPIController) getScenario(keyData *types.MockScenarioKeyData) (scenario *types.MockScenario, err error) {
	b, err := moc.mockScenarioRepository.LoadRaw(keyData.Method, keyData.Name, keyData.Path)
	if err != nil {
		return nil, err
	}
	scenario = &types.MockScenario{}
	if err = yaml.Unmarshal(b, scenario); err != nil {
		scenario, err = moc.mockScenarioRepository.Lookup(keyData, nil)
		if err != nil {
			return nil, err
		}
	}
	return scenario, err
}
