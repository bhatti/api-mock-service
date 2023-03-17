package controller

import (
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types/har"
	"github.com/bhatti/api-mock-service/internal/web"
	"net/http"
	"strconv"
)

// ExecHistoryController structure
type ExecHistoryController struct {
	scenarioRepository repository.APIScenarioRepository
}

// NewExecHistoryController instantiates controller for execution history
func NewExecHistoryController(
	scenarioRepository repository.APIScenarioRepository,
	webserver web.Server) *ExecHistoryController {
	ctrl := &ExecHistoryController{
		scenarioRepository: scenarioRepository,
	}

	webserver.GET("/_exec_history/names", ctrl.getExecHistoryNames)
	webserver.GET("/_exec_history/har", ctrl.getExecHistoryHar)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// getExecHistoryNames handler
// swagger:route GET /_exec_history/names exec-history getExecHistoryNames
// Fetches history of api scenarios by group.
// responses:
//
//	200: execHistoryNamesResponse
func (msc *ExecHistoryController) getExecHistoryNames(c web.APIContext) error {
	res := msc.scenarioRepository.HistoryNames(c.QueryParam("group"))
	if res == nil {
		res = make([]string, 0)
	}
	return c.JSON(http.StatusOK, res)
}

// getExecHistoryHar handler
// swagger:route GET /_exec_history/har exec-history getExecHistoryHar
// Fetches execution history in the format of HTTP HAR.
// responses:
//
//	200: execHistoryNamesResponse
func (msc *ExecHistoryController) getExecHistoryHar(c web.APIContext) error {
	name := c.QueryParam("name")
	group := c.QueryParam("group")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize := 20
	res, err := msc.scenarioRepository.LoadHar(name, group, page, pageSize)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

// ********************************* Swagger types ***********************************

// ExecHistory history scenario names
// swagger:response execHistoryNamesResponse
type execHistoryNamesResponse struct {
	// in:body
	Body []har.Log
}

// swagger:parameters execHistoryNamesResponse
// The parameters for HAR query parameters
type execHistoryNamesParams struct {
	// in:query
	Group string `json:"group"`
	// in:query
	Page string `json:"page"`
}

// ExecHistory history scenario names
// swagger:response execHistoryNamesResponse
type execHistoryNamesResponseBody struct {
	// in:body
	Body []string
}
