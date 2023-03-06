package controller

import (
	"context"
	"embed"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/oapi"
	"github.com/bhatti/api-mock-service/internal/utils"
	"gopkg.in/yaml.v3"
	"net/http"
	"strings"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
)

// MockOAPIController structure
type MockOAPIController struct {
	internalOAPI           embed.FS
	mockScenarioRepository repository.MockScenarioRepository
}

// NewMockOAPIController instantiates controller for updating mock-scenarios based on OpenAPI v3
func NewMockOAPIController(
	internalOAPI embed.FS,
	mockScenarioRepository repository.MockScenarioRepository,
	webserver web.Server) *MockOAPIController {
	ctrl := &MockOAPIController{
		internalOAPI:           internalOAPI,
		mockScenarioRepository: mockScenarioRepository,
	}

	webserver.GET("/_oapi/", ctrl.getOpenAPISpecsByGroup)
	webserver.GET("/_oapi/history/:name", ctrl.getOpenAPISpecsByHistory)
	webserver.GET("/_oapi/:group", ctrl.getOpenAPISpecsByGroup)
	webserver.GET("/_oapi/:method/:name/:path", ctrl.getOpenAPISpecsByScenario)
	webserver.POST("/_oapi", ctrl.postMockOAPIScenario)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// getOpenAPISpecsByGroup handler
// swagger:route GET /_oapi/{group} open-api getOpenAPISpecsByGroup
// Generates OpenAPI specs for the scenario group
// responses:
//
//	200: mockOapiSpecIResponse
func (moc *MockOAPIController) getOpenAPISpecsByGroup(c web.APIContext) (err error) {
	u := c.Request().URL
	group := c.Param("group")
	var b []byte
	if group == "_internal" {
		b, err = moc.internalOAPI.ReadFile("docs/openapi.json")
	} else {
		allByGroup := moc.mockScenarioRepository.LookupAllByGroup(group)
		if len(allByGroup) == 0 {
			allByGroup = moc.mockScenarioRepository.LookupAllByPath(
				strings.ReplaceAll(c.Request().URL.Path, "/_oapi", ""))
		}
		var scenarios []*types.MockScenario
		for _, keyData := range allByGroup {
			scenario, err := moc.getScenario(keyData, c.QueryParam("raw") == "true")
			if err != nil {
				return err
			}
			scenarios = append(scenarios, scenario)
		}
		b, err = oapi.MarshalScenarioToOpenAPI(group, "", scenarios...)
	}
	if err != nil {
		return err
	}

	b = []byte(strings.ReplaceAll(string(b), oapi.MockServerBaseURL, u.Scheme+"://"+u.Host))
	return c.Blob(http.StatusOK, "application/json", b)
}

// getOpenAPISpecsByHistory handler
// swagger:route GET /_oapi/history/{name} open-api getOpenAPISpecsByHistory
// Generates OpenAPI specs for the scenario history
// responses:
//
//	200: mockOapiSpecIResponse
func (moc *MockOAPIController) getOpenAPISpecsByHistory(c web.APIContext) (err error) {
	u := c.Request().URL
	name := c.Param("name")
	if name == "" {
		return fmt.Errorf("history name not specified in %s", c.Request().URL)
	}
	scenario, err := moc.mockScenarioRepository.LoadHistory(name)
	if err != nil {
		return err
	}
	b, err := oapi.MarshalScenarioToOpenAPI(scenario.Name, "", scenario)
	if err != nil {
		return err
	}
	b = []byte(strings.ReplaceAll(string(b), oapi.MockServerBaseURL, u.Scheme+"://"+u.Host))
	return c.Blob(http.StatusOK, "application/json", b)
}

// getOpenAPISpecsByScenario handler
// swagger:route GET /_oapi/{method}/{name}/{path} open-apiGetOpenAPISpecsByScenario
// Generates OpenAPI specs for the scenario
// responses:
//
//	200: mockOapiSpecIResponse
func (moc *MockOAPIController) getOpenAPISpecsByScenario(c web.APIContext) (err error) {
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
		Method:                   method,
		Name:                     name,
		Path:                     path,
		AssertQueryParamsPattern: make(map[string]string),
	}
	scenario, err := moc.getScenario(keyData, c.QueryParam("raw") == "true")
	if err != nil {
		return err
	}
	b, err := oapi.MarshalScenarioToOpenAPI("", "", scenario)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, "application/json", b)
}

// postMockOAPIScenario handler
// swagger:route POST /_oapi open-api postMockOAPIScenario
// Creates new mock scenarios based on Open API v3
// responses:
//
//	200: mockScenarioOAPIResponse
func (moc *MockOAPIController) postMockOAPIScenario(c web.APIContext) (err error) {
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

// swagger:parameters postMockOAPIScenario
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

func (moc *MockOAPIController) getScenario(keyData *types.MockScenarioKeyData, raw bool) (scenario *types.MockScenario, err error) {
	if raw {
		b, err := moc.mockScenarioRepository.LoadRaw(keyData.Method, keyData.Name, keyData.Path)
		if err != nil {
			return nil, err
		}
		scenario = &types.MockScenario{}
		if err = yaml.Unmarshal(b, scenario); err == nil {
			return scenario, err
		}
	}
	return moc.mockScenarioRepository.Lookup(keyData, nil)
}
