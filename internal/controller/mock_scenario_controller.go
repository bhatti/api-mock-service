package controller

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	log "github.com/sirupsen/logrus"
)

// MockScenarioController structure
type MockScenarioController struct {
	mockScenarioRepository repository.MockScenarioRepository
}

// NewMockScenarioController instantiates controller for updating mock-scenarios
func NewMockScenarioController(
	mockScenarioRepository repository.MockScenarioRepository,
	webserver web.Server) *MockScenarioController {
	ctrl := &MockScenarioController{
		mockScenarioRepository: mockScenarioRepository,
	}

	webserver.GET("/_scenarios", ctrl.ListMockScenarioPaths)
	webserver.GET("/_scenarios/:method/names/:path", ctrl.getMockNames)
	webserver.GET("/_scenarios/:method/:name/:path", ctrl.GetMockScenario)
	webserver.POST("/_scenarios", ctrl.PostMockScenario)
	webserver.DELETE("/_scenarios/:method/:name/:path", ctrl.DeleteMockScenario)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// PostMockScenario handler
// swagger:route POST /_scenarios mock-scenarios PostMockScenario
// Creates new mock scenario based on request body.
// responses:
//
//	200: mockScenarioResponse
func (msc *MockScenarioController) PostMockScenario(c web.APIContext) (err error) {
	scenario := &types.MockScenario{}
	if c.Request().Header.Get(types.ContentTypeHeader) == "application/yaml" {
		if err = msc.mockScenarioRepository.SaveRaw(c.Request().Body); err != nil {
			return err
		}
	} else {
		err = json.NewDecoder(c.Request().Body).Decode(scenario)
		if err = msc.mockScenarioRepository.Save(scenario); err != nil {
			return err
		}
	}
	return c.JSON(http.StatusOK, scenario)
}

// ListMockScenarioPaths handler
// swagger:route GET /_scenarios mock-scenarios listMockScenario
// List paths of all scenarios
// responses:
//
//	200: mockScenarioPathsResponse
func (msc *MockScenarioController) ListMockScenarioPaths(c web.APIContext) error {
	res := make(map[string]*types.MockScenarioKeyData)
	for _, next := range msc.mockScenarioRepository.ListScenarioKeyData(c.QueryParam("group")) {
		res[fmt.Sprintf("/_scenarios/%s/%s%s", next.Method, next.Name, next.Path)] = next
	}
	return c.JSON(http.StatusOK, res)
}

// GetMockScenario handler
// swagger:route GET /_scenarios/{method}/{name}/{path} mock-scenarios GetMockScenario
// Finds an existing mock scenario based on method, name and path
// responses:
//
//	200: mockScenarioResponse
func (msc *MockScenarioController) GetMockScenario(c web.APIContext) error {
	method, err := types.ToMethod(c.Param("method"))
	if err != nil {
		return err
	}
	name := c.Param("name")
	if name == "" {
		return fmt.Errorf("scenario name not specified")
	}
	path := c.Param("path")
	if path == "" {
		return fmt.Errorf("path not specified")
	}
	keyData, err := web.BuildMockScenarioKeyData(c.Request())
	if err != nil {
		return err
	}
	keyData.Method = method
	keyData.Name = name
	keyData.Path = path
	log.WithFields(log.Fields{
		"Method": method,
		"Name":   name,
		"Path":   path,
	}).Debugf("getting mock scenario...")

	scenario, err := msc.mockScenarioRepository.Lookup(keyData, nil)
	if err != nil {
		return err
	}
	if c.Request().Header.Get(types.ContentTypeHeader) == "application/yaml" {
		b, err := yaml.Marshal(scenario)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(b))
	}
	return c.JSON(http.StatusOK, scenario)
}

// swagger:route GET /_scenarios/{method}/names/{path} mock-scenarios getMockNames
// Returns mock scenario names
// responses:
//
//	200: mockNamesResponse
func (msc *MockScenarioController) getMockNames(c web.APIContext) error {
	method, err := types.ToMethod(c.Param("method"))
	if err != nil {
		return err
	}
	path := c.Param("path")
	if path == "" {
		return fmt.Errorf("path not specified")
	}
	log.WithFields(log.Fields{
		"Method": method,
		"Path":   path,
	}).Infof("getting mock scenario names...")
	names, err := msc.mockScenarioRepository.GetScenariosNames(method, path)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, names)
}

// DeleteMockScenario handler
// swagger:route DELETE /_scenarios/{method}/{name}/{path} mock-scenarios GetMockScenario
// Deletes an existing mock scenario based on id.
// responses:
//
//	200: emptyResponse
func (msc *MockScenarioController) DeleteMockScenario(c web.APIContext) error {
	method, err := types.ToMethod(c.Param("method"))
	if err != nil {
		return err
	}
	name := c.Param("name")
	if name == "" {
		return fmt.Errorf("scenario name not specified")
	}
	path := c.Param("path")
	if path == "" {
		return fmt.Errorf("path not specified")
	}
	log.WithFields(log.Fields{
		"Method": method,
		"Name":   name,
		"Path":   path,
	}).Infof("deleting mock scenario...")
	err = msc.mockScenarioRepository.Delete(method, name, path)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

// ********************************* Swagger types ***********************************

// swagger:parameters PostMockScenario
// The params for mock-scenario
type mockScenarioCreateParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Name string `json:"name"`
	// in:path
	Path string `json:"path"`
	// in:body
	Body types.MockScenario
}

// MockScenario body for update
// swagger:response mockScenarioResponse
type mockScenarioResponseBody struct {
	// in:body
	Body types.MockScenario
}

// MockScenario names
// swagger:response mockNamesResponse
type mockNamesResponseBody struct {
	// in:body
	Body []string
}

// MockScenario summary and paths
// swagger:response mockScenarioPathsResponse
type mockScenarioPathsResponseBody struct {
	// in:body
	Body map[string]*types.MockScenarioKeyData
}

// swagger:parameters DeleteMockScenario GetMockScenario
// The parameters for finding mock-scenario by method, path and name
type mockScenarioIDParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Name string `json:"name"`
	// in:path
	Path string `json:"path"`
}

// swagger:parameters getMockNames
// The parameters for finding mock-scenario names by path and method
type mockNamesParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Path string `json:"path"`
}
