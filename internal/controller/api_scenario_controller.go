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

// APIScenarioController structure
type APIScenarioController struct {
	scenarioRepository repository.APIScenarioRepository
	oapiRepository     repository.OAPIRepository
}

// NewAPIScenarioController instantiates controller for updating api-scenarios
func NewAPIScenarioController(
	scenarioRepository repository.APIScenarioRepository,
	oapiRepository repository.OAPIRepository,
	webserver web.Server) *APIScenarioController {
	ctrl := &APIScenarioController{
		scenarioRepository: scenarioRepository,
		oapiRepository:     oapiRepository,
	}

	webserver.GET("/_scenarios", ctrl.listAPIScenarioPaths)
	webserver.GET("/_scenarios/:method/names/:path", ctrl.getAPIScenarioNames)
	webserver.GET("/_scenarios/groups", ctrl.getAPIGroups)
	webserver.GET("/_scenarios/:method/:name/:path", ctrl.getAPIScenario)
	webserver.POST("/_scenarios", ctrl.postMockScenario)
	webserver.DELETE("/_scenarios/:method/:name/:path", ctrl.deleteAPIScenario)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// postMockScenario handler
// swagger:route POST /_scenarios api-scenarios postMockScenario
// Creates new api scenario based on request body.
// responses:
//
//	200: apiScenarioResponse
func (msc *APIScenarioController) postMockScenario(c web.APIContext) (err error) {
	scenario := &types.APIScenario{}
	if c.Request().Header.Get(types.ContentTypeHeader) != "application/yaml" {
		err = json.NewDecoder(c.Request().Body).Decode(scenario)
		if err != nil {
			return err
		}
		if err = msc.scenarioRepository.Save(scenario); err != nil {
			return err
		}
	} else {
		if err = msc.scenarioRepository.SaveRaw(c.Request().Body); err != nil {
			return err
		}
	}
	return c.JSON(http.StatusOK, scenario)
}

// listAPIScenarioPaths handler
// swagger:route GET /_scenarios api-scenarios listMockScenario
// List paths of all scenarios with group if available.
// responses:
//
//	200: apiScenarioPathsResponse
func (msc *APIScenarioController) listAPIScenarioPaths(c web.APIContext) error {
	res := make(map[string]*types.APIKeyData)
	for _, next := range msc.scenarioRepository.ListScenarioKeyData(c.QueryParam("group")) {
		res[fmt.Sprintf("/_scenarios/%s/%s%s", next.Method, next.Name, next.Path)] = next
	}
	return c.JSON(http.StatusOK, res)
}

// getAPIScenario handler
// swagger:route GET /_scenarios/{method}/{name}/{path} api-scenarios getAPIScenario
// Finds an existing api scenario based on method, name and path.
// responses:
//
//	200: apiScenarioResponse
func (msc *APIScenarioController) getAPIScenario(c web.APIContext) error {
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
	}).Debugf("getting api scenario...")

	b, err := msc.scenarioRepository.LoadRaw(method, name, path)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, string(b))
}

// swagger:route GET /_scenarios/groups api-scenarios getAPIGroups
// Finds api scenario groups.
// responses:
//
//	200: apiGroupsResponse
func (msc *APIScenarioController) getAPIGroups(c web.APIContext) error {
	groups := msc.scenarioRepository.GetGroups()
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

// swagger:route GET /_scenarios/{method}/names/{path} api-scenarios getAPIScenarioNames
// Finds api scenario names by method and path.
// responses:
//
//	200: apiNamesResponse
func (msc *APIScenarioController) getAPIScenarioNames(c web.APIContext) error {
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
	}).Infof("getting api scenario names...")
	names, err := msc.scenarioRepository.GetScenariosNames(method, path)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, names)
}

// deleteAPIScenario handler
// swagger:route DELETE /_scenarios/{method}/{name}/{path} api-scenarios getAPIScenario
// Deletes an existing api scenario based on id.
// responses:
//
//	200: emptyResponse
func (msc *APIScenarioController) deleteAPIScenario(c web.APIContext) error {
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
	}).Infof("deleting api scenario...")
	err = msc.scenarioRepository.Delete(method, name, path)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

// ********************************* Swagger types ***********************************

// swagger:parameters postMockScenario
// The params for api-scenario
type apiScenarioCreateParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Name string `json:"name"`
	// in:path
	Path string `json:"path"`
	// in:body
	Body types.APIScenario
}

// APIScenario body for update
// swagger:response apiScenarioResponse
type apiScenarioResponseBody struct {
	// in:body
	Body types.APIScenario
}

// APIScenario names
// swagger:response apiNamesResponse
type apiNamesResponseBody struct {
	// in:body
	Body []string
}

// APIScenario groups
// swagger:response apiGroupsResponse
type apiGroupsResponseBody struct {
	// in:body
	Body []string
}

// APIScenario summary and paths
// swagger:response apiScenarioPathsResponse
type apiScenarioPathsResponseBody struct {
	// in:body
	Body map[string]*types.APIKeyData
}

// swagger:parameters deleteAPIScenario getAPIScenario
// The parameters for finding api-scenario by method, path and name
type apiScenarioIDParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Name string `json:"name"`
	// in:path
	Path string `json:"path"`
}

// swagger:parameters getAPIScenarioNames
// The parameters for finding api-scenario names by path and method
type apiNamesParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Path string `json:"path"`
}
