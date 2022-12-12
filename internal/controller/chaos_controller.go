package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/chaos"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"github.com/bhatti/api-mock-service/internal/web"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// MockChaosController structure
type MockChaosController struct {
	executor *chaos.Executor
}

// NewMockChaosController instantiates controller for executing chaos client
func NewMockChaosController(
	executor *chaos.Executor,
	webserver web.Server) *MockChaosController {
	ctrl := &MockChaosController{
		executor: executor,
	}

	webserver.POST("/_chaos/:group", ctrl.PostMockChaosGroupScenario)
	webserver.POST("/_chaos/:method/:name/:path", ctrl.PostMockChaosScenario)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// PostMockChaosGroupScenario handler
// swagger:route POST /_chaos/{group} chaos PostMockChaosGroupScenario
// Plays chaos client for a scenario by group
// responses:
//
//	200: mockScenarioChaosResponse
func (mcc *MockChaosController) PostMockChaosGroupScenario(c web.APIContext) (err error) {
	group := c.Param("group")
	if group == "" {
		return fmt.Errorf("scenario group not specified")
	}
	chaosReq, err := buildChaosRequest(c)
	if err != nil {
		return err
	}
	dataTemplate := types.NewDataTemplateRequest(false, 1, 1)
	res := mcc.executor.ExecuteByGroup(context.Background(), group, dataTemplate, chaosReq)
	return c.JSON(http.StatusOK, res)
}

// PostMockChaosScenario handler
// swagger:route POST /_chaos/{method}/{name}/{path} chaos PostMockChaosScenario
// Plays chaos client for a scenario by name
// responses:
//
//	200: mockScenarioChaosResponse
func (mcc *MockChaosController) PostMockChaosScenario(c web.APIContext) (err error) {
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
	}).Debugf("chaos mocking scenario...")
	chaosReq, err := buildChaosRequest(c)
	if err != nil {
		return err
	}
	dataTemplate := types.NewDataTemplateRequest(true, 1, 1)
	res := mcc.executor.Execute(context.Background(), keyData, dataTemplate, chaosReq)
	return c.JSON(http.StatusOK, res)
}

// ********************************* Swagger types ***********************************

// swagger:parameters PostMockChaosScenario PostMockChaosGroupScenario
// The params for mock-scenario based on OpenAPI v3
type mockScenarioChaosCreateParams struct {
	// in:body
	Body types.ChaosRequest
}

// MockScenario body for update
// swagger:response mockScenarioChaosResponse
type mockScenarioChaosResponseBody struct {
	// in:body
	Body types.ChaosResponse
}

func buildChaosRequest(c web.APIContext) (types.ChaosRequest, error) {
	b, _, err := utils.ReadAll(c.Request().Body)
	if err != nil {
		return types.ChaosRequest{}, err
	}
	chaosReq := types.ChaosRequest{}
	err = json.Unmarshal(b, &chaosReq)
	if err != nil {
		return types.ChaosRequest{}, err
	}
	if chaosReq.BaseURL == "" {
		return types.ChaosRequest{}, fmt.Errorf("baseURL is not specified in %s", b)
	}
	if chaosReq.ExecutionTimes <= 0 {
		chaosReq.ExecutionTimes = 5
	}
	chaosReq.Overrides = make(map[string]any)
	for k, v := range c.QueryParams() {
		chaosReq.Overrides[k] = v[0]
	}
	return chaosReq, nil
}
