package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	log "github.com/sirupsen/logrus"
)

// MockScenarioController structure
type MockScenarioController struct {
	mockScenarioRepository repository.MockScenarioRepository
	oapiRepository         repository.OAPIRepository
}

// NewMockScenarioController instantiates controller for updating mock-scenarios
func NewMockScenarioController(
	mockScenarioRepository repository.MockScenarioRepository,
	oapiRepository repository.OAPIRepository,
	webserver web.Server) *MockScenarioController {
	ctrl := &MockScenarioController{
		mockScenarioRepository: mockScenarioRepository,
		oapiRepository:         oapiRepository,
	}

	webserver.GET("/_scenarios", ctrl.listMockScenarioPaths)
	webserver.GET("/_scenarios/history", ctrl.mockScenarioHistory)
	webserver.GET("/_scenarios/:method/names/:path", ctrl.getMockNames)
	webserver.GET("/_scenarios/groups", ctrl.getGroups)
	webserver.GET("/_scenarios/:method/:name/:path", ctrl.getMockScenario)
	webserver.POST("/_scenarios", ctrl.postMockScenario)
	webserver.DELETE("/_scenarios/:method/:name/:path", ctrl.deleteMockScenario)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// postMockScenario handler
// swagger:route POST /_scenarios mock-scenarios postMockScenario
// Creates new mock scenario based on request body.
// responses:
//
//	200: mockScenarioResponse
func (msc *MockScenarioController) postMockScenario(c web.APIContext) (err error) {
	scenario := &types.MockScenario{}
	if c.Request().Header.Get(types.ContentTypeHeader) != "application/yaml" {
		err = json.NewDecoder(c.Request().Body).Decode(scenario)
		if err = msc.mockScenarioRepository.Save(scenario); err != nil {
			return err
		}
	} else {
		if err = msc.mockScenarioRepository.SaveRaw(c.Request().Body); err != nil {
			return err
		}
	}
	return c.JSON(http.StatusOK, scenario)
}

// mockScenarioHistory handler
// swagger:route GET /_scenarios/history mock-scenarios mockScenarioHistory
// Fetches history of mock scenarios
// responses:
//
//	200: mockHistoryResponse
func (msc *MockScenarioController) mockScenarioHistory(c web.APIContext) error {
	res := msc.mockScenarioRepository.HistoryNames(c.QueryParam("group"))
	if res == nil {
		res = make([]string, 0)
	}
	return c.JSON(http.StatusOK, res)
}

// listMockScenarioPaths handler
// swagger:route GET /_scenarios mock-scenarios listMockScenario
// List paths of all scenarios
// responses:
//
//	200: mockScenarioPathsResponse
func (msc *MockScenarioController) listMockScenarioPaths(c web.APIContext) error {
	res := make(map[string]*types.MockScenarioKeyData)
	for _, next := range msc.mockScenarioRepository.ListScenarioKeyData(c.QueryParam("group")) {
		res[fmt.Sprintf("/_scenarios/%s/%s%s", next.Method, next.Name, next.Path)] = next
	}
	return c.JSON(http.StatusOK, res)
}

// getMockScenario handler
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
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	log.WithFields(log.Fields{
		"Method": method,
		"Name":   name,
		"Path":   path,
	}).Debugf("getting mock scenario...")

	b, err := msc.mockScenarioRepository.LoadRaw(method, name, path)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, string(b))
}

// swagger:route GET /_scenarios/groups mock-scenarios getGroups
// Returns mock scenario groups
// responses:
//
//	200: mockGroupsResponse
func (msc *MockScenarioController) getGroups(c web.APIContext) error {
	groups := msc.mockScenarioRepository.GetGroups()
	for _, name := range msc.oapiRepository.GetNames() {
		if name == "" {
			continue
		}
		dup := false
		for _, group := range groups {
			if name == group {
				dup = true
				break
			}
		}
		if !dup {
			groups = append(groups, name)
		}
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i] < groups[j]
	})

	return c.JSON(http.StatusOK, groups)
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

// deleteMockScenario handler
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

// MockScenario groups
// swagger:response mockGroupsResponse
type mockGroupsResponseBody struct {
	// in:body
	Body []string
}

// MockScenario history scenario names
// swagger:response mockHistoryResponse
type mockHistoryResponseBody struct {
	// in:body
	Body []string
}

// MockScenario summary and paths
// swagger:response mockScenarioPathsResponse
type mockScenarioPathsResponseBody struct {
	// in:body
	Body map[string]*types.MockScenarioKeyData
}

// swagger:parameters deleteMockScenario getMockScenario
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
