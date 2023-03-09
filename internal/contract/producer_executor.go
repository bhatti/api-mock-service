package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/metrics"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/bhatti/api-mock-service/internal/fuzz"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"github.com/bhatti/api-mock-service/internal/web"
	log "github.com/sirupsen/logrus"
)

// ProducerExecutor structure
type ProducerExecutor struct {
	scenarioRepository repository.MockScenarioRepository
	client             web.HTTPClient
}

// NewProducerExecutor executes contracts for producers
func NewProducerExecutor(
	scenarioRepository repository.MockScenarioRepository,
	client web.HTTPClient) *ProducerExecutor {
	return &ProducerExecutor{
		scenarioRepository: scenarioRepository,
		client:             client,
	}
}

// Execute an API with mock data
func (x *ProducerExecutor) Execute(
	ctx context.Context,
	req *http.Request,
	scenarioKey *types.MockScenarioKeyData,
	dataTemplate fuzz.DataTemplateRequest,
	contractReq *types.ContractRequest,
) *types.ContractResponse {
	started := time.Now()
	sli := metrics.NewMetrics()
	sli.RegisterHistogram(scenarioKey.SafeName())
	res := types.NewContractResponse()
	log.WithFields(log.Fields{
		"Component":       "ProducerExecutor",
		"Scenario":        scenarioKey,
		"ContractRequest": contractReq.String(),
	}).Infof("execute BEGIN")

	for i := 0; i < contractReq.ExecutionTimes; i++ {
		scenario, err := x.scenarioRepository.Lookup(scenarioKey, contractReq.Overrides())
		if err != nil {
			res.Add(scenarioKey.Name, nil, err)
			res.Metrics = sli.Summary()
			return res
		}
		url := scenario.BuildURL(contractReq.BaseURL)
		resContents, err := x.execute(ctx, req, url, scenario, contractReq, dataTemplate, sli)
		res.Add(scenario.Name, resContents, err)
		time.Sleep(scenario.WaitBeforeReply)
	}

	res.Metrics = sli.Summary()
	elapsed := time.Since(started).String()
	log.WithFields(log.Fields{
		"Component":       "ProducerExecutor",
		"Scenario":        scenarioKey,
		"ContractRequest": contractReq.String(),
		"Elapsed":         elapsed,
		"Errors":          len(res.Errors),
		"Metrics":         res.Metrics,
	}).Infof("execute COMPLETED")
	return res
}

// ExecuteByHistory executes execution history for an API with mock data
func (x *ProducerExecutor) ExecuteByHistory(
	ctx context.Context,
	req *http.Request,
	group string,
	dataTemplate fuzz.DataTemplateRequest,
	contractReq *types.ContractRequest,
) *types.ContractResponse {
	started := time.Now()
	execHistory := x.scenarioRepository.HistoryNames(group)
	res := types.NewContractResponse()
	log.WithFields(log.Fields{
		"Component":       "ProducerExecutor",
		"Group":           group,
		"ContractRequest": contractReq.String(),
		"History":         len(execHistory),
	}).Infof("execute-by-history BEGIN")

	sli := metrics.NewMetrics()
	registered := make(map[string]bool)

	for i := 0; i < contractReq.ExecutionTimes; i++ {
		for _, scenarioName := range execHistory {
			scenario, err := x.scenarioRepository.LoadHistory(scenarioName)
			if err != nil {
				res.Add(fmt.Sprintf("%s_%d", scenario.Name, i), nil, err)
				res.Metrics = sli.Summary()
				return res
			}
			if !registered[scenario.SafeName()] {
				sli.RegisterHistogram(scenario.SafeName())
			}
			url := scenario.BuildURL(contractReq.BaseURL)
			resContents, err := x.execute(ctx, req, url, scenario, contractReq, dataTemplate, sli)
			res.Add(fmt.Sprintf("%s_%d", scenario.Name, i), resContents, err)
			time.Sleep(scenario.WaitBeforeReply)
		}
	}

	elapsed := time.Since(started).String()
	res.Metrics = sli.Summary()
	log.WithFields(log.Fields{
		"Component":       "ProducerExecutor",
		"Group":           group,
		"ContractRequest": contractReq.String(),
		"Elapsed":         elapsed,
		"Errors":          len(res.Errors),
		"Metrics":         res.Metrics,
	}).Infof("execute-by-history COMPLETED")
	return res
}

// ExecuteByGroup executes an API with mock data
func (x *ProducerExecutor) ExecuteByGroup(
	ctx context.Context,
	req *http.Request,
	group string,
	dataTemplate fuzz.DataTemplateRequest,
	contractReq *types.ContractRequest,
) *types.ContractResponse {
	started := time.Now()
	scenarioKeys := x.scenarioRepository.LookupAllByGroup(group)
	res := types.NewContractResponse()
	log.WithFields(log.Fields{
		"Component":       "ProducerExecutor",
		"Group":           group,
		"ContractRequest": contractReq.String(),
		"ScenarioKeys":    scenarioKeys,
	}).Infof("execute-by-group BEGIN")

	sort.Slice(scenarioKeys, func(i, j int) bool {
		return scenarioKeys[i].Order < scenarioKeys[j].Order
	})
	sli := metrics.NewMetrics()
	for _, scenarioKey := range scenarioKeys {
		sli.RegisterHistogram(scenarioKey.SafeName())
	}

	for i := 0; i < contractReq.ExecutionTimes; i++ {
		for _, scenarioKey := range scenarioKeys {
			scenario, err := x.scenarioRepository.Lookup(scenarioKey, contractReq.Overrides())
			if err != nil {
				res.Add(fmt.Sprintf("%s_%d", scenarioKey.Name, i), nil, err)
				res.Metrics = sli.Summary()
				return res
			}
			url := scenario.BuildURL(contractReq.BaseURL)
			resContents, err := x.execute(ctx, req, url, scenario, contractReq, dataTemplate, sli)
			res.Add(fmt.Sprintf("%s_%d", scenarioKey.Name, i), resContents, err)
			time.Sleep(scenario.WaitBeforeReply)
		}
	}

	elapsed := time.Since(started).String()
	res.Metrics = sli.Summary()
	log.WithFields(log.Fields{
		"Component":       "ProducerExecutor",
		"Group":           group,
		"ContractRequest": contractReq.String(),
		"Elapsed":         elapsed,
		"Errors":          len(res.Errors),
		"ScenarioKeys":    scenarioKeys,
		"Metrics":         res.Metrics,
	}).Infof("execute-by-group COMPLETED")
	return res
}

// execute an API with mock data
func (x *ProducerExecutor) execute(
	ctx context.Context,
	req *http.Request,
	url string,
	scenario *types.MockScenario,
	contractReq *types.ContractRequest,
	dataTemplate fuzz.DataTemplateRequest,
	sli *metrics.Metrics,
) (res any, err error) {
	started := time.Now().UnixMilli()
	templateParams, queryParams, reqHeaders := scenario.Request.BuildTemplateParams(
		req,
		scenario.ToKeyData().MatchGroups(scenario.Path),
		contractReq.Headers,
		contractReq.Params)
	if fuzz.RandIntMinMax(1, 100) < 20 {
		dataTemplate = dataTemplate.WithMaxMultiplier(fuzz.RandIntMinMax(2, 5))
	}
	for k, v := range templateParams {
		url = strings.ReplaceAll(url, "{"+k+"}", fmt.Sprintf("%v", v))
	}

	reqBodyStr, reqBody := buildRequestBody(scenario)

	{
		// check request assertions
		reqContents, err := fuzz.UnmarshalArrayOrObject([]byte(reqBodyStr))
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal request body for (%s) due to %w", scenario.Name, err)
		}

		if err = scenario.Request.Assert(queryParams, reqHeaders, reqContents, templateParams); err != nil {
			return nil, err
		}
	}

	log.WithFields(log.Fields{
		"Component":   "ProducerExecutor",
		"URL":         url,
		"Scenario":    scenario,
		"Headers":     reqHeaders,
		"QueryParams": queryParams,
	}).Debugf("before execute")

	statusCode, resBody, resHeaders, err := x.client.Handle(
		ctx, url, string(scenario.Method), reqHeaders, queryParams, reqBody)
	elapsed := time.Now().UnixMilli() - started
	sli.AddHistogram(scenario.SafeName(), float64(elapsed)/1000.0, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke %s for %s (%s) due to %w", scenario.Name, url, scenario.Method, err)
	}

	var resBytes []byte
	if resBytes, resBody, err = utils.ReadAll(resBody); err != nil {
		return nil, fmt.Errorf("failed to read response body for %s due to %w", scenario.Name, err)
	}

	fields := log.Fields{
		"Component":   "ProducerExecutor",
		"URL":         url,
		"Scenario":    scenario,
		"StatusCode":  statusCode,
		"Headers":     reqHeaders,
		"QueryParams": queryParams,
		"Elapsed":     elapsed}

	// response contents
	resContents, err := fuzz.UnmarshalArrayOrObject(resBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response for %s due to %w", scenario.Name, err)
	}

	templateParams[fuzz.RequestCount] = fmt.Sprintf("%d", scenario.RequestCount)
	templateParams["status"] = statusCode
	templateParams["elapsed"] = elapsed

	if contractReq.Verbose {
		fields["Request"] = reqBodyStr
		fields["Response"] = resContents
		fields["ResponseBytes"] = string(resBytes)
	}

	if statusCode != scenario.Response.StatusCode {
		log.WithFields(fields).Warnf("failed to execute request, actual status %d != %d (scenario) for %s",
			statusCode, scenario.Response.StatusCode, scenario.Path)
		return nil, fmt.Errorf("failed to execute request with status %d didn't match expected value %d for %s (%s)",
			statusCode, scenario.Response.StatusCode, scenario.Name, scenario.Path)
	}

	if contractReq.Verbose {
		log.WithFields(fields).Infof("executed request")
	}
	if err = scenario.Response.Assert(resHeaders, resContents, templateParams); err != nil {
		return nil, err
	}

	if resContents != nil {
		if contractReq.Params == nil {
			contractReq.Params = map[string]any{}
		}
		pipeProperties := map[string]any{}

		for _, propName := range scenario.Response.PipeProperties {
			val := fuzz.FindVariable(propName, resContents)
			if val != nil {
				n := strings.Index(propName, ".")
				propName = propName[n+1:]
				contractReq.Params[propName] = val
				pipeProperties[propName] = val
			}
		}
		resContents = pipeProperties
	}
	return resContents, nil
}

func buildRequestBody(
	scenario *types.MockScenario,
) (string, io.ReadCloser) {
	var contents string
	if scenario.Request.Contents != "" {
		contents = scenario.Request.Contents
	} else if scenario.Request.AssertContentsPattern != "" {
		contents = scenario.Request.AssertContentsPattern
	}
	if contents == "" {
		return "", nil
	}
	res, err := fuzz.UnmarshalArrayOrObject([]byte(contents))
	if err != nil {
		log.WithFields(log.Fields{
			"Component": "ProducerExecutor",
			"Scenario":  scenario,
			"Error":     err,
		}).Infof("failed to unmarshal request")
		return "", nil
	}
	j, err := json.Marshal(fuzz.GenerateFuzzData(res))
	if err != nil {
		log.WithFields(log.Fields{
			"Component": "ProducerExecutor",
			"Scenario":  scenario,
			"Error":     err,
		}).Infof("failed to marshal populated request")
		return "", nil
	}
	return string(j), utils.NopCloser(bytes.NewReader(j))
}
