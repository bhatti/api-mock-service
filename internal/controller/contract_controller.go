package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"github.com/bhatti/api-mock-service/internal/web"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// ProducerContractController structure for producer driven contracts
type ProducerContractController struct {
	executor *contract.ProducerExecutor
}

// NewProducerContractController instantiates controller for executing contract client
func NewProducerContractController(
	executor *contract.ProducerExecutor,
	webserver web.Server) *ProducerContractController {
	ctrl := &ProducerContractController{
		executor: executor,
	}

	webserver.POST("/_contracts/:group", ctrl.postProducerContractGroupScenario)
	webserver.POST("/_contracts/history/:group", ctrl.postProducerContractHistoryByGroup)
	webserver.POST("/_contracts/history", ctrl.postProducerContractHistoryByGroup)
	webserver.POST("/_contracts/:method/:name/:path", ctrl.postProducerContractScenarioByPath)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// postProducerContractHistoryByGroup handler
// swagger:route POST /_contracts/history/{group} producer-contract postProducerContractHistoryByGroup
// Invokes service api-contract using executed history of consumer contracts
// responses:
//
//	200: apiScenarioContractResponse
func (mcc *ProducerContractController) postProducerContractHistoryByGroup(c web.APIContext) (err error) {
	group := c.Param("group")
	contractReq, err := buildContractRequest(c)
	if err != nil {
		return err
	}
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
	res := mcc.executor.ExecuteByHistory(context.Background(), c.Request(), group, dataTemplate, contractReq)
	return c.JSON(http.StatusOK, res)
}

// postProducerContractGroupScenario handler
// swagger:route POST /_contracts/{group} producer-contract postProducerContractGroupScenario
// Invokes service api-contract by group of api contracts
// responses:
//
//	200: apiScenarioContractResponse
func (mcc *ProducerContractController) postProducerContractGroupScenario(c web.APIContext) (err error) {
	group := c.Param("group")
	if group == "" {
		return fmt.Errorf("scenario group not specified")
	}
	contractReq, err := buildContractRequest(c)
	if err != nil {
		return err
	}
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
	res := mcc.executor.ExecuteByGroup(context.Background(), c.Request(), group, dataTemplate, contractReq)
	return c.JSON(http.StatusOK, res)
}

// postProducerContractScenarioByPath handler
// swagger:route POST /_contracts/{method}/{name}/{path} producer-contract postProducerContractScenarioByPath
// Plays contract client for a scenario by name
// Invokes service api-contract by method, contracts-name and path
// responses:
//
//	200: apiScenarioContractResponse
func (mcc *ProducerContractController) postProducerContractScenarioByPath(c web.APIContext) (err error) {
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
	}).Debugf("contract api scenario...")
	contractReq, err := buildContractRequest(c)
	if err != nil {
		return err
	}
	dataTemplate := fuzz.NewDataTemplateRequest(true, 1, 1)
	res := mcc.executor.Execute(context.Background(), c.Request(), keyData, dataTemplate, contractReq)
	return c.JSON(http.StatusOK, res)
}

// ********************************* Swagger types ***********************************

// swagger:parameters postProducerContractScenarioByPath
// The params for api-contract based on OpenAPI v3
type apiScenarioContractCreateParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Name string `json:"name"`
	// in:path
	Path string `json:"path"`
	// in:body
	Body types.ProducerContractRequest
}

// The params for api-contract based on OpenAPI v3
// swagger:parameters postProducerContractGroupScenario
type postProducerContractGroupScenarioParams struct {
	// in:path
	Group string `json:"group"`
	// in:body
	Body types.ProducerContractRequest
}

// The params for api-contract based on OpenAPI v3
// swagger:parameters postProducerContractHistoryByGroup
type postProducerContractHistoryParams struct {
	// in:path
	Group string `json:"group"`
	// in:body
	Body types.ProducerContractRequest
}

// APIScenario body for update
// swagger:response apiScenarioContractResponse
type apiScenarioContractResponseBody struct {
	// in:body
	Body types.ProducerContractResponse
}

func buildContractRequest(c web.APIContext) (*types.ProducerContractRequest, error) {
	b, _, err := utils.ReadAll(c.Request().Body)
	if err != nil {
		return nil, err
	}
	contractReq := &types.ProducerContractRequest{}
	err = json.Unmarshal(b, &contractReq)
	if err != nil {
		return nil, err
	}
	// contractReq.BaseURL may be nil
	if contractReq.ExecutionTimes <= 0 {
		contractReq.ExecutionTimes = 5
	}
	contractReq.Params = make(map[string]any)
	contractReq.Headers = c.Request().Header
	for k, v := range c.QueryParams() {
		contractReq.Params[k] = v[0]
	}
	return contractReq, nil
}
