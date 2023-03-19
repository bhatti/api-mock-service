package controller

import (
	"encoding/json"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/types/archive"
	"github.com/bhatti/api-mock-service/internal/types/postman"
	"github.com/bhatti/api-mock-service/internal/web"
	"net/http"
	"strconv"
	"time"
)

const pageSize = 50

// APIHistoryController structure
type APIHistoryController struct {
	config             *types.Configuration
	scenarioRepository repository.APIScenarioRepository
}

// NewAPIHistoryController instantiates controller for execution history
func NewAPIHistoryController(
	config *types.Configuration,
	scenarioRepository repository.APIScenarioRepository,
	webserver web.Server) *APIHistoryController {
	ctrl := &APIHistoryController{
		config:             config,
		scenarioRepository: scenarioRepository,
	}

	webserver.GET("/_history", ctrl.getExecHistory)
	webserver.GET("/_history/names", ctrl.getExecHistoryNames)
	webserver.GET("/_history/har", ctrl.getExecHistoryHar)
	webserver.POST("/_history/har", ctrl.postExecHistoryHar)
	webserver.GET("/_history/postman", ctrl.getExecHistoryPostman)
	webserver.POST("/_history/postman", ctrl.postExecHistoryPostman)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// getExecHistory handler
// swagger:route GET /_history api-history getExecHistory
// Fetches history of api scenarios by name or group.
// responses:
//
//	200: getExecHistoryResponse
func (ehc *APIHistoryController) getExecHistory(c web.APIContext) error {
	res := ehc.scenarioRepository.HistoryNames(c.QueryParam("group"))
	if res == nil {
		res = make([]string, 0)
	}
	return c.JSON(http.StatusOK, res)
}

// getExecHistoryNames handler
// swagger:route GET /_history/names api-history getExecHistoryNames
// Fetches history of api scenarios by group.
// responses:
//
//	200: execHistoryNamesResponse
func (ehc *APIHistoryController) getExecHistoryNames(c web.APIContext) error {
	res := ehc.scenarioRepository.HistoryNames(c.QueryParam("group"))
	if res == nil {
		res = make([]string, 0)
	}
	return c.JSON(http.StatusOK, res)
}

// postExecHistoryHar handler
// swagger:route POST /_history/har api-history postExecHistoryHar
// Uploads HAR contents and generates scenario and history.
// responses:
//
//	200: postExecHistoryHarResponse
func (ehc *APIHistoryController) postExecHistoryHar(c web.APIContext) (err error) {
	har := &archive.Har{}
	err = json.NewDecoder(c.Request().Body).Decode(har)
	if err != nil {
		return err
	}
	scenarios := archive.ConvertHarToScenarios(ehc.config, har)
	for _, scenario := range scenarios {
		u, err := scenario.GetURL("")
		if err != nil {
			return err
		}
		if err = ehc.scenarioRepository.SaveHistory(scenario, u.String(), scenario.StartTime, scenario.EndTime); err != nil {
			return err
		}

	}
	return c.NoContent(http.StatusOK)
}

// getExecHistoryHar handler
// swagger:route GET /_history/har api-history getExecHistoryHar
// Fetches execution history in the format of HTTP HAR.
// responses:
//
//	200: execHistoryHarResponse
func (ehc *APIHistoryController) getExecHistoryHar(c web.APIContext) error {
	name := c.QueryParam("name")
	group := c.QueryParam("group")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	scenarios, err := ehc.scenarioRepository.LoadHistory(name, group, page, pageSize)
	if err != nil {
		return err
	}
	har := archive.ConvertScenariosToHar(ehc.config, nil, time.Time{}, time.Time{}, scenarios...)
	return c.JSON(http.StatusOK, har)
}

// postExecHistoryPostman handler
// swagger:route POST /_history/postman api-history postExecHistoryPostman
// Uploads HAR contents and generates scenario and history.
// responses:
//
//	200: postExecHistoryPostmanResponse
func (ehc *APIHistoryController) postExecHistoryPostman(c web.APIContext) (err error) {
	collection := &postman.Collection{}
	err = json.NewDecoder(c.Request().Body).Decode(collection)
	if err != nil {
		return err
	}
	scenarios := postman.ConvertPostmanToScenarios(ehc.config, collection, time.Time{}, time.Time{})
	for _, scenario := range scenarios {
		u, err := scenario.GetURL("")
		if err != nil {
			return err
		}
		if err = ehc.scenarioRepository.Save(scenario); err != nil {
			return err
		}
		if err = ehc.scenarioRepository.SaveHistory(scenario, u.String(), scenario.StartTime, scenario.EndTime); err != nil {
			return err
		}
	}
	return c.JSON(http.StatusOK, postExecHistoryHarResponseBody{Count: len(scenarios)})
}

// getExecHistoryPostman handler
// swagger:route GET /_history/postman api-history getExecHistoryPostman
// Fetches execution history in the format of HTTP HAR.
// responses:
//
//	200: execHistoryPostmanResponse
func (ehc *APIHistoryController) getExecHistoryPostman(c web.APIContext) error {
	name := c.QueryParam("name")
	group := c.QueryParam("group")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	scenarios, err := ehc.scenarioRepository.LoadHistory(name, group, page, pageSize)
	if err != nil {
		return err
	}
	collection := postman.ConvertScenariosToPostman("", scenarios...)
	return c.JSON(http.StatusOK, collection)
}

// ********************************* Swagger types ***********************************

// swagger:response execHistoryNamesResponse
// response for history of scenario names
type execHistoryNamesResponseBody struct {
	// in:body
	Body []string
}

// swagger:parameters getExecHistoryHar
// The parameters for HAR query parameters
type execHistoryHarParams struct {
	// in:query
	Name string `json:"name"`
	// in:query
	Group string `json:"group"`
	// in:query
	Page string `json:"page"`
}

// swagger:response execHistoryHarResponse
// ExecHistory history in the format of HAR format
type execHistoryHarResponse struct {
	// in:body
	Body archive.Har
}

// swagger:parameters postExecHistoryHar
// The parameters for uploading har contents
type postExecHistoryHarParams struct {
	// in:body
	Body archive.Har
}

// swagger:response postExecHistoryHarResponse
// response for uploading HAR
type postExecHistoryHarResponseBody struct {
	// in:body
	Count int `json:"count"`
}

// swagger:parameters getExecHistory
// The parameters for history
type getExecHistoryParams struct {
	// in:query
	Name string `json:"name"`
	// in:query
	Group string `json:"group"`
	// in:query
	Page string `json:"page"`
}

// swagger:response getExecHistoryResponse
// response for history of scenario history
type getExecHistoryResponseBody struct {
	// in:body
	Body []*types.APIScenario
}

// swagger:parameters getExecHistoryPostman
// The parameters for HAR query parameters
type execHistoryPostmanParams struct {
	// in:query
	Name string `json:"name"`
	// in:query
	Group string `json:"group"`
	// in:query
	Page string `json:"page"`
}

// swagger:response execHistoryPostmanResponse
// ExecHistory history in the format of HAR format
type execHistoryPostmanResponse struct {
	// in:body
	Body postman.Collection
}

// swagger:parameters postExecHistoryPostman
// The parameters for uploading har contents
type postExecHistoryPostmanParams struct {
	// in:body
	Body postman.Collection
}

// swagger:response postExecHistoryPostmanResponse
// response for uploading postman
type postExecHistoryPostmanResponseBody struct {
}
