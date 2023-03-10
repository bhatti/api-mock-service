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

// ContractController structure for producer driven contracts
type ContractController struct {
	executor *contract.ProducerExecutor
}

// NewContractController instantiates controller for executing contract client
func NewContractController(
	executor *contract.ProducerExecutor,
	webserver web.Server) *ContractController {
	ctrl := &ContractController{
		executor: executor,
	}

	webserver.POST("/_contracts/:group", ctrl.postMockContractGroupScenario)
	webserver.POST("/_contracts/history/:group", ctrl.postMockContractHistory)
	webserver.POST("/_contracts/history", ctrl.postMockContractHistory)
	webserver.POST("/_contracts/:method/:name/:path", ctrl.postMockContractScenario)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// postMockContractHistory handler
// swagger:route POST /_contracts/history/{group} contract postMockContractHistory
// Plays contract client for a scenario by history
// responses:
//
//	200: mockScenarioContractResponse
func (mcc *ContractController) postMockContractHistory(c web.APIContext) (err error) {
	group := c.Param("group")
	contractReq, err := buildContractRequest(c)
	if err != nil {
		return err
	}
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
	res := mcc.executor.ExecuteByHistory(context.Background(), c.Request(), group, dataTemplate, contractReq)
	return c.JSON(http.StatusOK, res)
}

// postMockContractGroupScenario handler
// swagger:route POST /_contracts/{group} contract postMockContractGroupScenario
// Plays contract client for a scenario by group
// responses:
//
//	200: mockScenarioContractResponse
func (mcc *ContractController) postMockContractGroupScenario(c web.APIContext) (err error) {
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

// postMockContractScenario handler
// swagger:route POST /_contracts/{method}/{name}/{path} contract postMockContractScenario
// Plays contract client for a scenario by name
// responses:
//
//	200: mockScenarioContractResponse
func (mcc *ContractController) postMockContractScenario(c web.APIContext) (err error) {
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
	}).Debugf("contract mocking scenario...")
	contractReq, err := buildContractRequest(c)
	if err != nil {
		return err
	}
	dataTemplate := fuzz.NewDataTemplateRequest(true, 1, 1)
	res := mcc.executor.Execute(context.Background(), c.Request(), keyData, dataTemplate, contractReq)
	return c.JSON(http.StatusOK, res)
}

// ********************************* Swagger types ***********************************

// swagger:parameters postMockContractScenario postMockContractGroupScenario
// The params for mock-scenario based on OpenAPI v3
type mockScenarioContractCreateParams struct {
	// in:body
	Body types.ContractRequest
}

// MockScenario body for update
// swagger:response mockScenarioContractResponse
type mockScenarioContractResponseBody struct {
	// in:body
	Body types.ContractResponse
}

func buildContractRequest(c web.APIContext) (*types.ContractRequest, error) {
	b, _, err := utils.ReadAll(c.Request().Body)
	if err != nil {
		return nil, err
	}
	contractReq := &types.ContractRequest{}
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
