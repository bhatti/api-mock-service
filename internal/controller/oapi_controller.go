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

// OAPIController structure
type OAPIController struct {
	internalOAPI       embed.FS
	scenarioRepository repository.APIScenarioRepository
	oapiRepository     repository.OAPIRepository
}

// NewOAPIController instantiates controller for updating api-scenarios based on OpenAPI v3
func NewOAPIController(
	internalOAPI embed.FS,
	scenarioRepository repository.APIScenarioRepository,
	oapiRepository repository.OAPIRepository,
	webserver web.Server) *OAPIController {
	ctrl := &OAPIController{
		internalOAPI:       internalOAPI,
		scenarioRepository: scenarioRepository,
		oapiRepository:     oapiRepository,
	}

	webserver.GET("/_oapi", ctrl.getOpenAPISpecsByGroup)
	webserver.GET("/_oapi/", ctrl.getOpenAPISpecsByGroup)
	webserver.GET("/_oapi/:group", ctrl.getOpenAPISpecsByGroup)
	webserver.GET("/_oapi/history/:name", ctrl.getOpenAPISpecsByHistory)
	webserver.GET("/_oapi/:method/:name/:path", ctrl.getOpenAPISpecsByScenario)
	webserver.POST("/_oapi", ctrl.postMockOAPIScenario)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// getOpenAPISpecsByGroup handler
// swagger:route GET /_oapi/{group} open-api getOpenAPISpecsByGroup
// Generates OpenAPI specs by group of API scenarios.
// responses:
//
//	200: apiOapiSpecIResponse
func (moc *OAPIController) getOpenAPISpecsByGroup(c web.APIContext) (err error) {
	u := c.Request().URL
	group := c.Param("group")
	var b []byte
	if group == "_internal" || group == "" {
		b, err = moc.internalOAPI.ReadFile("docs/openapi.json")
	} else {
		b, err = moc.oapiRepository.LoadRaw(group)
		if err != nil {
			allByGroup := moc.scenarioRepository.LookupAllByGroup(group)
			if len(allByGroup) == 0 {
				allByGroup = moc.scenarioRepository.LookupAllByPath(
					strings.ReplaceAll(c.Request().URL.Path, "/_oapi", ""))
			}
			var scenarios []*types.APIScenario
			for _, keyData := range allByGroup {
				scenario, err := moc.getScenario(keyData, c.QueryParam("raw") == "true")
				if err != nil {
					return err
				}
				scenarios = append(scenarios, scenario)
			}
			b, err = oapi.MarshalScenarioToOpenAPI(group, "", scenarios...)
		}
	}
	if err != nil {
		return err
	}

	b = []byte(strings.ReplaceAll(string(b), oapi.MockServerBaseURL, u.Scheme+"://"+u.Host))
	return c.Blob(http.StatusOK, "application/json", b)
}

// getOpenAPISpecsByHistory handler
// swagger:route GET /_oapi/history/{name} open-api getOpenAPISpecsByHistory
// Generates OpenAPI specs based on name of API scenario from execution history.
// responses:
//
//	200: apiOapiSpecIResponse
func (moc *OAPIController) getOpenAPISpecsByHistory(c web.APIContext) (err error) {
	u := c.Request().URL
	name := c.Param("name")
	if name == "" {
		return fmt.Errorf("history name not specified in %s", c.Request().URL)
	}
	scenarios, err := moc.scenarioRepository.LoadHistory(name, "", 0, 100)
	if err != nil {
		return err
	}
	if len(scenarios) == 0 {
		return fmt.Errorf("history not found for %s", name)
	}
	b, err := oapi.MarshalScenarioToOpenAPI(scenarios[0].Name, "", scenarios[0])
	if err != nil {
		return err
	}
	b = []byte(strings.ReplaceAll(string(b), oapi.MockServerBaseURL, u.Scheme+"://"+u.Host))
	return c.Blob(http.StatusOK, "application/json", b)
}

// getOpenAPISpecsByScenario handler
// swagger:route GET /_oapi/{method}/{name}/{path} open-api getOpenAPISpecsByScenario
// Generates OpenAPI specs for the API scenario by method, name and path.
// responses:
//
//	200: apiOapiSpecIResponse
func (moc *OAPIController) getOpenAPISpecsByScenario(c web.APIContext) (err error) {
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
	keyData := &types.APIKeyData{
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
// Creates new api scenarios based on Open API v3 uploaded by user.
// responses:
//
//	200: apiScenarioOAPIResponse
func (moc *OAPIController) postMockOAPIScenario(c web.APIContext) (err error) {
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
	scenarios := make([]*types.APIScenario, 0)
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		if err != nil {
			return err
		}
		err = moc.scenarioRepository.Save(scenario)
		if err != nil {
			return err
		}
		scenarios = append(scenarios, scenario)
	}
	if len(specs) > 0 {
		err = moc.oapiRepository.SaveRaw(specs[0].Title, data)
		if err != nil {
			return err
		}
	}
	return c.JSON(http.StatusOK, scenarios)
}

// ********************************* Swagger types ***********************************

// swagger:parameters postMockOAPIScenario
// The params for api-scenario based on OpenAPI v3
type apiScenarioOAPICreateParams struct {
	// in:body
	Body []byte
}

// APIScenario body for update
// swagger:response apiScenarioOAPIResponse
type apiScenarioOAPIResponseBody struct {
	// in:body
	Body types.APIScenario
}

// The params for group
// swagger:parameters getOpenAPISpecsByGroup
type getOpenAPISpecsByGroupParams struct {
	// in:path
	Group string `json:"group"`
}

// getOpenAPISpecsByHistoryParams params for name
// swagger:parameters getOpenAPISpecsByHistory
type getOpenAPISpecsByHistoryParams struct {
	// Name of open-api spec
	// in: path
	Name string `json:"name"`
}

// The params for name
// swagger:parameters getOpenAPISpecsByScenario
type getOpenAPISpecsByScenarioParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Name string `json:"name"`
	// in:path
	Path string `json:"path"`
}

// APIScenario body for update
// swagger:response apiOapiSpecIResponse
type apiOapiSpecIResponseBody struct {
	// in:body
	Body []byte
}

func (moc *OAPIController) getScenario(keyData *types.APIKeyData, raw bool) (scenario *types.APIScenario, err error) {
	if raw {
		b, err := moc.scenarioRepository.LoadRaw(keyData.Method, keyData.Name, keyData.Path)
		if err != nil {
			return nil, err
		}
		scenario = &types.APIScenario{}
		if err = yaml.Unmarshal(b, scenario); err == nil {
			return scenario, err
		}
	}
	return moc.scenarioRepository.Lookup(keyData, nil)
}
