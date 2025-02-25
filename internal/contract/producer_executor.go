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
	scenarioRepository    repository.APIScenarioRepository
	groupConfigRepository repository.GroupConfigRepository
	client                web.HTTPClient
}

// NewProducerExecutor executes contracts for producers
func NewProducerExecutor(
	scenarioRepository repository.APIScenarioRepository,
	groupConfigRepository repository.GroupConfigRepository,
	client web.HTTPClient) *ProducerExecutor {
	return &ProducerExecutor{
		scenarioRepository:    scenarioRepository,
		groupConfigRepository: groupConfigRepository,
		client:                client,
	}
}

// Execute an API with fuzz data request
func (px *ProducerExecutor) Execute(
	ctx context.Context,
	req *http.Request,
	scenarioKey *types.APIKeyData,
	dataTemplate fuzz.DataTemplateRequest,
	contractReq *types.ProducerContractRequest,
) *types.ProducerContractResponse {
	started := time.Now()
	sli := metrics.NewMetrics()
	sli.RegisterHistogram(scenarioKey.SafeName())
	contractResponse := types.NewProducerContractResponse()
	if contractReq.Verbose {
		log.WithFields(log.Fields{
			"Component":               "ProducerExecutor",
			"Scenario":                scenarioKey,
			"ProducerContractRequest": contractReq.String(),
		}).Infof("execute BEGIN")
	}

	for i := 0; i < contractReq.ExecutionTimes; i++ {
		scenario, err := px.scenarioRepository.Lookup(scenarioKey, contractReq.Overrides())
		if err != nil {
			contractResponse.Add(scenarioKey.Name, nil, err)
			contractResponse.Metrics = sli.Summary()
			return contractResponse
		}
		url := scenario.BuildURL(contractReq.BaseURL)
		resContents, err := px.execute(ctx, req, url, scenario, contractReq, contractResponse, dataTemplate, sli)
		contractResponse.Add(scenario.Name, resContents, err)
		time.Sleep(scenario.WaitBeforeReply)
	}

	contractResponse.Metrics = sli.Summary()
	elapsed := time.Since(started).String()
	if contractReq.Verbose {
		log.WithFields(log.Fields{
			"Component":               "ProducerExecutor",
			"Scenario":                scenarioKey,
			"ProducerContractRequest": contractReq.String(),
			"Elapsed":                 elapsed,
			"Errors":                  len(contractResponse.Errors),
			"Metrics":                 contractResponse.Metrics,
		}).Infof("execute COMPLETED")
	}
	return contractResponse
}

// ExecuteByHistory executes execution history for an API with fuzz data request
func (px *ProducerExecutor) ExecuteByHistory(
	ctx context.Context,
	req *http.Request,
	group string,
	dataTemplate fuzz.DataTemplateRequest,
	contractReq *types.ProducerContractRequest,
) *types.ProducerContractResponse {
	started := time.Now()
	execHistory := px.scenarioRepository.HistoryNames(group)
	contractResponse := types.NewProducerContractResponse()
	log.WithFields(log.Fields{
		"Component":               "ProducerExecutor",
		"Group":                   group,
		"ProducerContractRequest": contractReq.String(),
		"History":                 len(execHistory),
	}).Infof("execute-by-history BEGIN")

	sli := metrics.NewMetrics()
	registered := make(map[string]bool)

	for i := 0; i < contractReq.ExecutionTimes; i++ {
		for _, scenarioName := range execHistory {
			scenarios, err := px.scenarioRepository.LoadHistory(scenarioName, "", 0, 100)
			if err != nil {
				contractResponse.Add(fmt.Sprintf("%s_%d", scenarioName, i), nil, err)
				contractResponse.Metrics = sli.Summary()
				return contractResponse
			}
			for _, scenario := range scenarios {
				if !registered[scenario.SafeName()] {
					sli.RegisterHistogram(scenario.SafeName())
				}
				url := scenario.BuildURL(contractReq.BaseURL)
				resContents, err := px.execute(ctx, req, url, scenario, contractReq, contractResponse, dataTemplate, sli)
				contractResponse.Add(fmt.Sprintf("%s_%d", scenario.Name, i), resContents, err)
				time.Sleep(scenario.WaitBeforeReply)
			}
		}
	}

	elapsed := time.Since(started).String()
	contractResponse.Metrics = sli.Summary()
	log.WithFields(log.Fields{
		"Component":               "ProducerExecutor",
		"Group":                   group,
		"ProducerContractRequest": contractReq.String(),
		"Elapsed":                 elapsed,
		"Errors":                  len(contractResponse.Errors),
		"Metrics":                 contractResponse.Metrics,
	}).Infof("execute-by-history COMPLETED")
	return contractResponse
}

// ExecuteByGroup executes an API with fuzz data request
func (px *ProducerExecutor) ExecuteByGroup(
	ctx context.Context,
	req *http.Request,
	group string,
	dataTemplate fuzz.DataTemplateRequest,
	contractReq *types.ProducerContractRequest,
) *types.ProducerContractResponse {
	started := time.Now()
	scenarioKeys := px.scenarioRepository.LookupAllByGroup(group)
	contractResponse := types.NewProducerContractResponse()
	log.WithFields(log.Fields{
		"Component":               "ProducerExecutor",
		"Group":                   group,
		"ProducerContractRequest": contractReq.String(),
		"ScenarioKeys":            scenarioKeys,
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
			scenario, err := px.scenarioRepository.Lookup(scenarioKey, contractReq.Overrides())
			if err != nil {
				contractResponse.Add(fmt.Sprintf("%s_%d", scenarioKey.Name, i), nil, err)
				contractResponse.Metrics = sli.Summary()
				return contractResponse
			}
			url := scenario.BuildURL(contractReq.BaseURL)
			resContents, err := px.execute(ctx, req, url, scenario, contractReq, contractResponse, dataTemplate, sli)
			contractResponse.Add(fmt.Sprintf("%s_%d", scenarioKey.Name, i), resContents, err)
			time.Sleep(scenario.WaitBeforeReply)
		}
	}

	elapsed := time.Since(started).String()
	contractResponse.Metrics = sli.Summary()
	log.WithFields(log.Fields{
		"Component":               "ProducerExecutor",
		"Group":                   group,
		"ProducerContractRequest": contractReq.String(),
		"Elapsed":                 elapsed,
		"Errors":                  len(contractResponse.Errors),
		"ScenarioKeys":            scenarioKeys,
		"Metrics":                 contractResponse.Metrics,
	}).Infof("execute-by-group COMPLETED")
	return contractResponse
}

// execute an API with fuzz data request
func (px *ProducerExecutor) execute(
	ctx context.Context,
	req *http.Request,
	url string,
	scenario *types.APIScenario,
	contractReq *types.ProducerContractRequest,
	contractRes *types.ProducerContractResponse,
	dataTemplate fuzz.DataTemplateRequest,
	sli *metrics.Metrics,
) (res any, err error) {
	if req == nil {
		return nil, fmt.Errorf("http request is not specified")
	}
	if url == "" {
		return nil, fmt.Errorf("http URL is not specified")
	}
	if scenario == nil {
		return nil, fmt.Errorf("scenario is not specified")
	}
	if contractReq == nil {
		return nil, fmt.Errorf("contract request is not specified")
	}
	if contractRes == nil {
		return nil, fmt.Errorf("contract response is not specified")
	}
	if !strings.HasPrefix(url, "http") {
		return nil, fmt.Errorf("http URL is not valid %s, scenario url %s", url, scenario.BaseURL)
	}

	started := time.Now().UnixMilli()
	templateParams, queryParams, postParams, reqHeaders := scenario.Request.BuildTemplateParams(
		req, scenario.ToKeyData().MatchGroups(scenario.Path),
		contractReq.Headers, contractReq.Params)
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

		if err = scenario.Request.Assert(queryParams, postParams, reqHeaders, reqContents, templateParams); err != nil {
			return nil, fmt.Errorf("request assertion for scenario %s (%s) [headers: %v] failed: %s",
				scenario.Name, scenario.Description, req.Header, err)
		}
	}

	log.WithFields(log.Fields{
		"Component":   "ProducerExecutor",
		"URL":         url,
		"Scenario":    scenario,
		"Headers":     reqHeaders,
		"QueryParams": queryParams,
	}).Debugf("before execute")

	statusCode, httpVersion, resBody, resHeaders, err := px.client.Handle(
		ctx, url, string(scenario.Method), reqHeaders, queryParams, reqBody)
	elapsed := time.Now().UnixMilli() - started
	sli.AddHistogram(scenario.SafeName(), float64(elapsed)/1000.0, nil)
	contractRes.URLs[url] = contractRes.URLs[url] + 1
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
		"HTTPVersion": httpVersion,
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
		sharedVariables := make(map[string]any)
		// TODO should we return filtered response from shared variables
		_ = handleSharedVariables(scenario, resContents, contractReq.Params,
			px.groupConfigRepository.Variables(scenario.Group), sharedVariables, resHeaders)
	}
	return resContents, nil
}

func handleSharedVariables(scenario *types.APIScenario, resContents any,
	params map[string]any, groupVariables map[string]string,
	sharedVariables map[string]any, resHeaders http.Header) any {
	if resContents == nil {
		return nil
	}
	for k, v := range groupVariables {
		sharedVariables[k] = v
	}

	for _, propName := range scenario.Response.AddSharedVariables {
		val := fuzz.FindVariable(propName, resContents)
		if val != nil {
			n := strings.Index(propName, ".")
			propName = propName[n+1:]
			params[propName] = val
			sharedVariables[propName] = val
		} else {
			vals := resHeaders[propName]
			if len(vals) > 0 {
				params[propName] = vals[0]
				sharedVariables[propName] = vals[0]
			}
		}
	}
	for _, propName := range scenario.Response.DeleteSharedVariables {
		delete(params, propName)
		delete(sharedVariables, propName)
	}
	return sharedVariables
}

func buildRequestBody(
	scenario *types.APIScenario,
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
	j, err := json.MarshalIndent(fuzz.GenerateFuzzData(res), "", "  ")
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
