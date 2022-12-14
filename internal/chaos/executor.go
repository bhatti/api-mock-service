package chaos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
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
	chaosReq types.ChaosRequest,
) *types.ChaosResponse {
	started := time.Now()
	scenarioKey.Name = ""
	res := types.NewChaosResponse(nil, 0, 0)
	log.WithFields(log.Fields{
		"Component":    "Tester",
		"Scenario":     scenarioKey,
		"ChaosRequest": chaosReq,
	}).Infof("execute BEGIN")

	for i := 0; i < chaosReq.ExecutionTimes; i++ {
		scenario, err := x.scenarioRepository.Lookup(scenarioKey)
		if err != nil {
			res.Errors = append(res.Errors, err.Error())
			return res
		}
		url := chaosReq.BaseURL + scenario.Path
		err = x.execute(ctx, url, scenario, chaosReq.Overrides, dataTemplate, chaosReq)
		res.Add(err)
		time.Sleep(scenario.WaitBeforeReply)
	}

	elapsed := time.Since(started).String()
	log.WithFields(log.Fields{
		"Component":    "Tester",
		"Scenario":     scenarioKey,
		"ChaosRequest": chaosReq,
		"Elapsed":      elapsed,
		"Errors":       len(res.Errors),
	}).Infof("execute COMPLETED")
	return res
}

// ExecuteByGroup an API with mock data
func (x *Executor) ExecuteByGroup(
	ctx context.Context,
	group string,
	dataTemplate fuzz.DataTemplateRequest,
	chaosReq types.ChaosRequest,
) *types.ChaosResponse {
	started := time.Now()
	scenarioKeys := x.scenarioRepository.LookupAllByGroup(group)
	res := types.NewChaosResponse(nil, 0, 0)
	log.WithFields(log.Fields{
		"Component":    "Tester",
		"Group":        group,
		"ChaosRequest": chaosReq,
		"Request":      chaosReq,
	}).Infof("execute-by-group BEGIN")

	for _, scenarioKey := range scenarioKeys {
		if chaosReq.Verbose {
			log.WithFields(log.Fields{
				"Component":    "Tester",
				"Scenario":     scenarioKey,
				"ChaosRequest": chaosReq,
				"Request":      chaosReq,
			}).Infof("execute-by-group-key BEGIN")
		}

		for i := 0; i < chaosReq.ExecutionTimes; i++ {
			scenario, err := x.scenarioRepository.Lookup(scenarioKey)
			if err != nil {
				res.Errors = append(res.Errors, err.Error())
				return res
			}
			url := chaosReq.BaseURL + scenario.Path
			err = x.execute(ctx, url, scenario, chaosReq.Overrides, dataTemplate, chaosReq)
			res.Add(err)
			time.Sleep(scenario.WaitBeforeReply)
		}

		elapsed := time.Since(started).String()
		if chaosReq.Verbose {
			log.WithFields(log.Fields{
				"Component":    "Tester",
				"Scenario":     scenarioKey,
				"ChaosRequest": chaosReq,
				"Elapsed":      elapsed,
				"Errors":       len(res.Errors),
				"Request":      chaosReq,
			}).Infof("execute-by-group-key COMPLETED")
		}
	}
	elapsed := time.Since(started).String()
	log.WithFields(log.Fields{
		"Component":    "Tester",
		"Group":        group,
		"ChaosRequest": chaosReq,
		"Elapsed":      elapsed,
		"Errors":       len(res.Errors),
		"Request":      chaosReq,
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
	chaosRequest types.ChaosRequest,
) (err error) {
	started := time.Now()
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
	if err != nil {
		return err
	}
	elapsed := time.Since(started).String()
	var resBytes []byte
	if resBytes, resBody, err = utils.ReadAll(resBody); err != nil {
		return err
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
		return fmt.Errorf("failed to execute request with status %d due to %s", statusCode, resBytes)
	}

	var resContents any
	if resContents, err = updateTemplateParams(templateParams, scenario, resBytes, resHeaders, statusCode); err != nil {
		return err
	}

	if chaosRequest.Verbose {
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
			return fmt.Errorf("failed to find required header %s with regex %s", k, v)
		}
		match, err := regexp.MatchString(v, actualHeader[0])
		if err != nil {
			return fmt.Errorf("failed to fuzz required header %s with regex %s and actual value %s due to %w",
				k, v, actualHeader[0], err)
		}
		if !match {
			return fmt.Errorf("didn't match required header %s with regex %s and actual value %s",
				k, v, actualHeader[0])
		}
	}

	if scenario.Response.MatchContents != "" {
		regex := make(map[string]string)
		err := json.Unmarshal([]byte(scenario.Response.MatchContents), &regex)
		if err != nil {
			return fmt.Errorf("failed to unmarshal response '%s' regex due to %w", scenario.Response.MatchContents, err)
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
			return fmt.Errorf("failed to validate response due to %w", err)
		}
	}

	for _, assertion := range scenario.Response.Assertions {
		assertion = normalizeAssertion(assertion)
		b, err := utils.ParseTemplate("", []byte(assertion), templateParams)
		if err != nil {
			return fmt.Errorf("failed to parse assertion %s due to %w", assertion, err)
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
			return fmt.Errorf("failed to assert '%s' with value '%s'", assertion, b)
		}
	}
	return nil
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
	templateParams[types.RequestCount] = fmt.Sprintf("%d", scenario.RequestCount)
	contents, err := fuzz.UnmarshalArrayOrObject(resBytes)
	if err != nil {
		return nil, err
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
	if scenario.Request.MatchContents != "" {
		contents = scenario.Request.MatchContents
	} else if scenario.Request.ExampleContents != "" {
		contents = scenario.Request.ExampleContents
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
	res = fuzz.PopulateRandomData(res)
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
	for k, v := range scenario.Request.ExamplePathParams {
		templateParams[k] = v
		queryParams[k] = v
	}
	for k, v := range scenario.Request.ExampleQueryParams {
		templateParams[k] = v
		queryParams[k] = v
	}
	for k, v := range scenario.Request.MatchQueryParams {
		templateParams[k] = regexValue(v)
		queryParams[k] = regexValue(v)
	}
	for k, v := range scenario.Request.ExampleHeaders {
		templateParams[k] = v
		reqHeaders[k] = []string{v}
	}
	for k, v := range scenario.Request.MatchHeaders {
		templateParams[k] = regexValue(v)
		reqHeaders[k] = []string{regexValue(v)}
	}
	// Find any params for query params and path variables
	for k, v := range scenario.ToKeyData().MatchGroups(scenario.Path) {
		templateParams[k] = v
	}
	for k, v := range overrides {
		templateParams[k] = v
	}
	return
}

func regexValue(val string) string {
	if strings.HasPrefix(val, "__") || strings.HasPrefix(val, "(") {
		return fuzz.RandRegex(val)
	}
	return val
}
