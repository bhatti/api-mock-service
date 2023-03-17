package controller

import (
	"encoding/json"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types/archive"
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
	webserver.POST("/_exec_history/har", ctrl.postExecHistoryHar)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// getExecHistoryNames handler
// swagger:route GET /_exec_history/names exec-history getExecHistoryNames
// Fetches history of api scenarios by group.
// responses:
//
//	200: execHistoryNamesResponse
func (ehc *ExecHistoryController) getExecHistoryNames(c web.APIContext) error {
	res := ehc.scenarioRepository.HistoryNames(c.QueryParam("group"))
	if res == nil {
		res = make([]string, 0)
	}
	return c.JSON(http.StatusOK, res)
}

// postExecHistoryHar handler
// swagger:route POST /_exec_history/har exec-history postExecHistoryHar
// Uploads HAR contents and generates scenario and history.
// responses:
//
//	200: postExecHistoryHarResponse
func (ehc *ExecHistoryController) postExecHistoryHar(c web.APIContext) (err error) {
	har := &archive.Har{}
	err = json.NewDecoder(c.Request().Body).Decode(har)
	if err != nil {
		return err
	}
	if err = ehc.scenarioRepository.SaveHar(har); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

// getExecHistoryHar handler
// swagger:route GET /_exec_history/har exec-history getExecHistoryHar
// Fetches execution history in the format of HTTP HAR.
// responses:
//
//	200: execHistoryHarResponse
func (ehc *ExecHistoryController) getExecHistoryHar(c web.APIContext) error {
	name := c.QueryParam("name")
	group := c.QueryParam("group")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize := 20
	res, err := ehc.scenarioRepository.LoadHar(name, group, page, pageSize)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

// ********************************* Swagger types ***********************************

// swagger:response execHistoryHarResponse
// ExecHistory history in the format of HAR format
type execHistoryHarResponse struct {
	// in:body
	Body []archive.HarLog
}

// swagger:parameters getExecHistoryHar
// The parameters for HAR query parameters
type execHistoryNamesParams struct {
	// in:query
	Group string `json:"group"`
	// in:query
	Page string `json:"page"`
}

// swagger:response execHistoryNamesResponse
// response for history of scenario names
type execHistoryNamesResponseBody struct {
	// in:body
	Body []string
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
}
