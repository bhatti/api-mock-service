package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/metrics"
	"io"
	"os"
	"regexp"
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

// Executor structure
type Executor struct {
	scenarioRepository repository.MockScenarioRepository
	client             web.HTTPClient
}

// NewExecutor instantiates controller for updating mock-scenarios
func NewExecutor(
	scenarioRepository repository.MockScenarioRepository,
	client web.HTTPClient) *Executor {
	return &Executor{
		scenarioRepository: scenarioRepository,
		client:             client,
	}
}

// Execute an API with mock data
func (x *Executor) Execute(
	ctx context.Context,
	scenarioKey *types.MockScenarioKeyData,
	dataTemplate fuzz.DataTemplateRequest,
	contractReq types.ContractRequest,
) *types.ContractResponse {
	started := time.Now()
	sli := metrics.NewMetrics()
	sli.RegisterHistogram(scenarioKey.SafeName())
	res := types.NewContractResponse()
	log.WithFields(log.Fields{
		"Component":       "Tester",
		"Scenario":        scenarioKey,
		"ContractRequest": contractReq,
	}).Infof("execute BEGIN")

	for i := 0; i < contractReq.ExecutionTimes; i++ {
		scenario, err := x.scenarioRepository.Lookup(scenarioKey, contractReq.Overrides)
		if err != nil {
			res.Add(scenarioKey.Name, nil, err)
			res.Metrics = sli.Summary()
			return res
		}
		url := contractReq.BaseURL + scenario.Path
		resContents, err := x.execute(ctx, url, scenario, contractReq.Overrides, dataTemplate, contractReq, sli)
		res.Add(scenario.Name, resContents, err)
		time.Sleep(scenario.WaitBeforeReply)
	}
	res.Metrics = sli.Summary()
	elapsed := time.Since(started).String()
	log.WithFields(log.Fields{
		"Component":       "Tester",
		"Scenario":        scenarioKey,
		"ContractRequest": contractReq,
		"Elapsed":         elapsed,
		"Errors":          len(res.Errors),
		"Metrics":         res.Metrics,
	}).Infof("execute COMPLETED")
	return res
}

// ExecuteByGroup an API with mock data
func (x *Executor) ExecuteByGroup(
	ctx context.Context,
	group string,
	dataTemplate fuzz.DataTemplateRequest,
	contractReq types.ContractRequest,
) *types.ContractResponse {
	started := time.Now()
	scenarioKeys := x.scenarioRepository.LookupAllByGroup(group)
	res := types.NewContractResponse()
	log.WithFields(log.Fields{
		"Component":       "Tester",
		"Group":           group,
		"ContractRequest": contractReq,
		"Request":         contractReq,
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
			scenario, err := x.scenarioRepository.Lookup(scenarioKey, contractReq.Overrides)
			if err != nil {
				res.Add(fmt.Sprintf("%s_%d", scenarioKey.Name, i), nil, err)
				res.Metrics = sli.Summary()
				return res
			}
			url := contractReq.BaseURL + scenario.Path
			resContents, err := x.execute(ctx, url, scenario, contractReq.Overrides, dataTemplate, contractReq, sli)
			res.Add(fmt.Sprintf("%s_%d", scenarioKey.Name, i), resContents, err)
			time.Sleep(scenario.WaitBeforeReply)
		}
	}

	elapsed := time.Since(started).String()
	res.Metrics = sli.Summary()
	log.WithFields(log.Fields{
		"Component":       "Tester",
		"Group":           group,
		"ContractRequest": contractReq,
		"Elapsed":         elapsed,
		"Errors":          len(res.Errors),
		"ScenarioKeys":    scenarioKeys,
		"Metrics":         res.Metrics,
	}).Infof("execute-by-group COMPLETED")
	return res
}

// execute an API with mock data
func (x *Executor) execute(
	ctx context.Context,
	url string,
	scenario *types.MockScenario,
	overrides map[string]any,
	dataTemplate fuzz.DataTemplateRequest,
	contractRequest types.ContractRequest,
	metrics *metrics.Metrics,
) (res any, err error) {
	started := time.Now().UnixMilli()
	templateParams, queryParams, reqHeaders := buildTemplateParams(scenario, overrides)
	if fuzz.RandNumMinMax(1, 100) < 20 {
		dataTemplate = dataTemplate.WithMaxMultiplier(fuzz.RandNumMinMax(2, 5))
	}
	for k, v := range templateParams {
		url = strings.ReplaceAll(url, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	reqContents, reqBody := buildRequestBody(scenario)

	statusCode, resBody, resHeaders, err := x.client.Handle(
		ctx, url, string(scenario.Method), reqHeaders, queryParams, reqBody)
	elapsed := time.Now().UnixMilli() - started
	metrics.AddHistogram(scenario.SafeName(), float64(elapsed)/1000.0, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke %s for %s (%s) due to %w", scenario.Name, url, scenario.Method, err)
	}
	var resBytes []byte
	if resBytes, resBody, err = utils.ReadAll(resBody); err != nil {
		return nil, fmt.Errorf("failed to read response body for %s due to %w", scenario.Name, err)
	}

	if statusCode >= 300 {
		log.WithFields(log.Fields{
			"Component":  "Tester",
			"URL":        url,
			"Scenario":   scenario,
			"StatusCode": statusCode,
			"Elapsed":    elapsed,
			"Headers":    reqHeaders,
			"Request":    reqContents,
			"Response":   string(resBytes)}).Warnf("failed to execute request")
		return nil, fmt.Errorf("failed to execute request with status %d due to %s", statusCode, resBytes)
	}

	var resContents any
	if resContents, err = updateTemplateParams(templateParams, scenario, resBytes, resHeaders, statusCode); err != nil {
		return nil, err
	}

	if contractRequest.Verbose {
		log.WithFields(log.Fields{
			"Component":  "Tester",
			"URL":        url,
			"Scenario":   scenario,
			"StatusCode": statusCode,
			"Elapsed":    elapsed,
			"Headers":    reqHeaders,
			"Request":    reqContents,
			"Response":   resContents}).Infof("executed request")
	}

	for k, v := range scenario.Response.MatchHeaders {
		actualHeader := resHeaders[k]
		if len(actualHeader) == 0 {
			return nil, fmt.Errorf("failed to find required header %s with regex %s", k, v)
		}
		match, err := regexp.MatchString(v, actualHeader[0])
		if err != nil {
			return nil, fmt.Errorf("failed to fuzz required header %s with regex %s and actual value %s due to %w",
				k, v, actualHeader[0], err)
		}
		if !match {
			return nil, fmt.Errorf("didn't match required header %s with regex %s and actual value %s",
				k, v, actualHeader[0])
		}
	}

	if scenario.Response.MatchContents != "" {
		regex := make(map[string]string)
		err := json.Unmarshal([]byte(scenario.Response.MatchContents), &regex)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response '%s' regex due to %w", scenario.Response.MatchContents, err)
		}
		err = fuzz.ValidateRegexMap(resContents, regex)
		if err != nil {
			log.WithFields(log.Fields{
				"Component":  "Tester",
				"URL":        url,
				"Scenario":   scenario,
				"StatusCode": statusCode,
				"Elapsed":    elapsed,
				"Headers":    reqHeaders,
				"Request":    reqContents,
				"Response":   resContents}).Warnf("failed to validate resposne")
			return nil, fmt.Errorf("failed to validate response due to %w", err)
		}
	}

	for _, assertion := range scenario.Response.Assertions {
		assertion = normalizeAssertion(assertion)
		b, err := fuzz.ParseTemplate("", []byte(assertion), templateParams)
		if err != nil {
			return nil, fmt.Errorf("failed to parse assertion %s due to %w", assertion, err)
		}
		if string(b) == "true" {
			log.WithFields(log.Fields{
				"Component":  "Tester",
				"URL":        url,
				"Scenario":   scenario,
				"StatusCode": statusCode,
				"Assertion":  assertion,
				"Elapsed":    elapsed,
				"Output":     string(b)}).Debugf("successfully asserted test")
		} else {
			log.WithFields(log.Fields{
				"Component":  "Tester",
				"URL":        url,
				"Scenario":   scenario,
				"StatusCode": statusCode,
				"Assertion":  assertion,
				"Elapsed":    elapsed,
				"Params":     templateParams,
				"Headers":    reqHeaders,
				"Request":    reqContents,
				"Response":   resContents,
				"Output":     string(b)}).Warnf("failed to assert test")
			return nil, fmt.Errorf("failed to assert '%s' with value '%s'",
				assertion, b)
		}
	}
	if resContents != nil {
		if overrides == nil {
			overrides = map[string]any{}
		}
		pipeProperties := map[string]any{}

		for _, propName := range scenario.Response.PipeProperties {
			val := fuzz.FindVariable(propName, resContents)
			if val != nil {
				n := strings.Index(propName, ".")
				propName = propName[n+1:]
				overrides[propName] = val
				pipeProperties[propName] = val
			}
		}
		resContents = pipeProperties
	}
	return resContents, nil
}

func normalizeAssertion(assertion string) string {
	if !strings.HasPrefix(assertion, "{{") {
		parts := strings.Split(assertion, " ")
		var sb strings.Builder
		sb.WriteString("{{")
		for i, next := range parts {
			if i > 0 {
				sb.WriteString(fmt.Sprintf(` "%s"`, next))
			} else {
				sb.WriteString(next)
			}
		}
		sb.WriteString("}}")
		assertion = sb.String()
	}
	return assertion
}

func updateTemplateParams(
	templateParams map[string]any,
	scenario *types.MockScenario,
	resBytes []byte,
	resHeaders map[string][]string,
	statusCode int) (any, error) {
	templateParams[fuzz.RequestCount] = fmt.Sprintf("%d", scenario.RequestCount)
	contents, err := fuzz.UnmarshalArrayOrObject(resBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response for %s due to %w", scenario.Name, err)
	}
	if contents != nil {
		templateParams["contents"] = contents
	}
	flatHeaders := make(map[string]string)
	for k, v := range resHeaders {
		flatHeaders[k] = v[0]
	}
	templateParams["headers"] = flatHeaders
	templateParams["status"] = statusCode
	return contents, nil
}

func buildRequestBody(
	scenario *types.MockScenario,
) (string, io.ReadCloser) {
	var contents string
	if scenario.Request.Contents != "" {
		contents = scenario.Request.Contents
	} else if scenario.Request.MatchContents != "" {
		contents = scenario.Request.MatchContents
	}
	if contents == "" {
		return "", nil
	}
	res, err := fuzz.UnmarshalArrayOrObject([]byte(contents))
	if err != nil {
		log.WithFields(log.Fields{
			"Component": "Tester",
			"Scenario":  scenario,
			"Error":     err,
		}).Infof("failed to unmarshal request")
		return "", nil
	}
	res = fuzz.GenerateFuzzData(res)
	j, err := json.Marshal(res)
	if err != nil {
		log.WithFields(log.Fields{
			"Component": "Tester",
			"Scenario":  scenario,
			"Error":     err,
		}).Infof("failed to marshal populated request")
		return "", nil
	}
	return string(j), io.NopCloser(bytes.NewReader(j))
}

func buildTemplateParams(
	scenario *types.MockScenario,
	overrides map[string]any,
) (templateParams map[string]any, queryParams map[string]string, reqHeaders map[string][]string) {
	templateParams = make(map[string]any)
	queryParams = make(map[string]string)
	reqHeaders = make(map[string][]string)
	for _, env := range os.Environ() {
		parts := strings.Split(env, "=")
		if len(parts) == 2 {
			templateParams[parts[0]] = parts[1]
		}
	}
	for k, v := range scenario.Request.PathParams {
		templateParams[k] = v
		queryParams[k] = v
	}
	for k, v := range scenario.Request.MatchQueryParams {
		templateParams[k] = regexValue(v)
		queryParams[k] = regexValue(v)
	}
	for k, v := range scenario.Request.QueryParams {
		templateParams[k] = v
		queryParams[k] = v
	}
	for k, v := range scenario.Request.MatchHeaders {
		templateParams[k] = regexValue(v)
		reqHeaders[k] = []string{regexValue(v)}
	}
	for k, v := range scenario.Request.Headers {
		templateParams[k] = v
		reqHeaders[k] = []string{v}
	}
	// Find any params for query params and path variables
	for k, v := range scenario.ToKeyData().MatchGroups(scenario.Path) {
		templateParams[k] = v
	}
	for k, v := range overrides {
		templateParams[k] = v
		queryParams[k] = fmt.Sprintf("%v", v)
	}
	return
}

func regexValue(val string) string {
	if strings.HasPrefix(val, "__") || strings.HasPrefix(val, "(") {
		return fuzz.RandRegex(val)
	}
	return val
}
