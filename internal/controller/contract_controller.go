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
	executor *contract.Executor
}

// NewContractController instantiates controller for executing contract client
func NewContractController(
	executor *contract.Executor,
	webserver web.Server) *ContractController {
	ctrl := &ContractController{
		executor: executor,
	}

	webserver.POST("/_contracts/:group", ctrl.PostMockContractGroupScenario)
	webserver.POST("/_contracts/:method/:name/:path", ctrl.PostMockContractScenario)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// PostMockContractGroupScenario handler
// swagger:route POST /_contracts/{group} contract PostMockContractGroupScenario
// Plays contract client for a scenario by group
// responses:
//
//	200: mockScenarioContractResponse
func (mcc *ContractController) PostMockContractGroupScenario(c web.APIContext) (err error) {
	group := c.Param("group")
	if group == "" {
		return fmt.Errorf("scenario group not specified")
	}
	contractReq, err := buildContractRequest(c)
	if err != nil {
		return err
	}
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 1)
	res := mcc.executor.ExecuteByGroup(context.Background(), group, dataTemplate, contractReq)
	return c.JSON(http.StatusOK, res)
}

// PostMockContractScenario handler
// swagger:route POST /_contracts/{method}/{name}/{path} contract PostMockContractScenario
// Plays contract client for a scenario by name
// responses:
//
//	200: mockScenarioContractResponse
func (mcc *ContractController) PostMockContractScenario(c web.APIContext) (err error) {
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
	res := mcc.executor.Execute(context.Background(), keyData, dataTemplate, contractReq)
	return c.JSON(http.StatusOK, res)
}

// ********************************* Swagger types ***********************************

// swagger:parameters PostMockContractScenario PostMockContractGroupScenario
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

func buildContractRequest(c web.APIContext) (types.ContractRequest, error) {
	b, _, err := utils.ReadAll(c.Request().Body)
	if err != nil {
		return types.ContractRequest{}, err
	}
	contractReq := types.ContractRequest{}
	err = json.Unmarshal(b, &contractReq)
	if err != nil {
		return types.ContractRequest{}, err
	}
	if contractReq.BaseURL == "" {
		return types.ContractRequest{}, fmt.Errorf("baseURL is not specified in %s", b)
	}
	if contractReq.ExecutionTimes <= 0 {
		contractReq.ExecutionTimes = 5
	}
	contractReq.Overrides = make(map[string]any)
	for k, v := range c.Request().Header {
		contractReq.Overrides[k] = v[0]
	}
	for k, v := range c.QueryParams() {
		contractReq.Overrides[k] = v[0]
	}
	return contractReq, nil
}
