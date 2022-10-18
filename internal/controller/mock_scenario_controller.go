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

	webserver.GET("/_scenarios/:method/scenarios/:path", ctrl.getMockNames)
	webserver.GET("/_scenarios/:method/:name/:path", ctrl.getMockScenario)
	webserver.POST("/_scenarios", ctrl.postMockScenario)
	webserver.DELETE("/_scenarios/:method/:name/:path", ctrl.deleteMockScenario)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// swagger:route POST /_scenarios mock-scenarios postMockScenario
// Creates new mock scenario based on request body.
// responses:
//
//	200: mockScenarioResponse
func (msc *MockScenarioController) postMockScenario(c web.APIContext) (err error) {
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

// swagger:route GET /_scenarios/{method}/{name}/{path} mock-scenarios getMockScenario
// Finds an existing mock scenario based on method, name and path
// responses:
//
//	200: mockScenarioResponse
func (msc *MockScenarioController) getMockScenario(c web.APIContext) error {
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
	}).Infof("getting mock scenario...")
	scenario, err := msc.mockScenarioRepository.Get(
		method,
		name,
		path,
		nil)
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

// swagger:route GET /_scenarios/{method}/scenarios/{path} mock-scenarios getMockNames
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

// swagger:route DELETE /_scenarios/{method}/{name}/{path} mock-scenarios getMockScenario
// Deletes an existing mock scenario based on id.
// responses:
//
//	200: emptyResponse
func (msc *MockScenarioController) deleteMockScenario(c web.APIContext) error {
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

// swagger:parameters postMockScenario
// The params for mock-scenario
type mockScenarioCreateParams struct {
	// in:body
	Body types.MockScenario
}

// swagger:parameters putMockScenario
// The params for mock-scenario
type mockScenarioUpdateParams struct {
	// in:path
	Name string `json:"name"`
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

// swagger:parameters deleteMockScenario getMockScenario
// The parameters for finding mock-scenario by path, method and name
type mockScenarioIDParams struct {
	// in:path
	Name string `json:"nam"`
	// in:path
	Path string `json:"path"`
}

// swagger:parameters getMockNames
// The parameters for finding mock-scenario names by path and method
type mockNamesParams struct {
	// in:path
	Path string `json:"path"`
}
