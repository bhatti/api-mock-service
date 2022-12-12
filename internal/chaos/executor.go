package chaos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"github.com/bhatti/api-mock-service/internal/web"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
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
	dataTemplate types.DataTemplateRequest,
	chaosReq types.ChaosRequest,
) *types.ChaosResponse {
	started := time.Now()
	scenarioKey.Name = ""
	scenario, err := x.scenarioRepository.Lookup(scenarioKey)
	res := types.NewChaosResponse(nil, 0, 0)
	if err != nil {
		res.Errors = append(res.Errors, err)
		return res
	}
	url := chaosReq.BaseURL + scenario.Path
	log.WithFields(log.Fields{
		"Component": "Tester",
		"Scenario":  scenario,
		"URL":       url,
	}).Infof("execute BEGIN")

	for i := 0; i < chaosReq.ExecutionTimes; i++ {
		err := x.execute(ctx, url, scenario, chaosReq.Overrides, dataTemplate, chaosReq)
		if err != nil {
			res.Errors = append(res.Errors, err)
			res.Failed++
		} else {
			res.Succeeded++
		}
		time.Sleep(scenario.WaitBeforeReply)
	}
	elapsed := time.Since(started).String()
	log.WithFields(log.Fields{
		"Component": "Tester",
		"URL":       url,
		"Scenario":  scenario,
		"Elapsed":   elapsed,
		"Errors":    len(res.Errors),
	}).Infof("execute COMPLETED")
	return res
}

// ExecuteByGroup an API with mock data
func (x *Executor) ExecuteByGroup(
	ctx context.Context,
	group string,
	dataTemplate types.DataTemplateRequest,
	chaosReq types.ChaosRequest,
) *types.ChaosResponse {
	started := time.Now()
	scenarioKeys := x.scenarioRepository.LookupAllByGroup(group)
	res := types.NewChaosResponse(nil, 0, 0)
	for _, scenarioKey := range scenarioKeys {
		scenario, err := x.scenarioRepository.Lookup(scenarioKey)
		if err != nil {
			res.Errors = append(res.Errors, err)
			continue
		}
		url := chaosReq.BaseURL + scenario.Path
		log.WithFields(log.Fields{
			"Component": "Tester",
			"Scenario":  scenario,
			"URL":       url,
			"Request":   chaosReq,
		}).Infof("execute-by-group BEGIN")

		for i := 0; i < chaosReq.ExecutionTimes; i++ {
			err := x.execute(ctx, url, scenario, chaosReq.Overrides, dataTemplate, chaosReq)
			if err != nil {
				res.Errors = append(res.Errors, err)
				res.Failed++
			} else {
				res.Succeeded++
			}
			time.Sleep(scenario.WaitBeforeReply)
		}

		elapsed := time.Since(started).String()
		log.WithFields(log.Fields{
			"Component": "Tester",
			"URL":       url,
			"Scenario":  scenario,
			"Elapsed":   elapsed,
			"Errors":    len(res.Errors),
			"Request":   chaosReq,
		}).Infof("execute-by-group COMPLETED")
	}
	return res
}

// execute an API with mock data
func (x *Executor) execute(
	ctx context.Context,
	url string,
	scenario *types.MockScenario,
	overrides map[string]any,
	dataTemplate types.DataTemplateRequest,
	chaosRequest types.ChaosRequest,
) (err error) {
	started := time.Now()
	templateParams := buildTemplateParams(scenario, overrides)
	if utils.RandNumMinMax(1, 100) < 20 {
		dataTemplate = dataTemplate.WithMaxMultiplier(utils.RandNumMinMax(2, 5))
	}
	for k, v := range templateParams {
		url = strings.ReplaceAll(url, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	headers := make(map[string][]string)
	params := make(map[string]string)
	reqContents, reqBody := buildRequestBody(scenario, dataTemplate)

	statusCode, resBody, resHeaders, err := x.client.Handle(
		ctx, url, string(scenario.Method), headers, params, reqBody)
	if err != nil {
		return err
	}
	elapsed := time.Since(started).String()
	var resBytes []byte
	if resBytes, resBody, err = utils.ReadAll(resBody); err != nil {
		return err
	}

	var resContents interface{}
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
			return fmt.Errorf("failed to match required header %s with regex %s and actual value %s due to %w",
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
		err = utils.ValidateRegexMap(resContents, regex)
		if err != nil {
			log.WithFields(log.Fields{
				"Component":  "Tester",
				"URL":        url,
				"Scenario":   scenario,
				"StatusCode": statusCode,
				"Elapsed":    elapsed,
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
	statusCode int) (interface{}, error) {
	templateParams[types.RequestCount] = fmt.Sprintf("%d", scenario.RequestCount)
	contents, err := utils.UnmarshalArrayOrObject(resBytes)
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
	dataTemplate types.DataTemplateRequest,
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
	res, err := utils.UnmarshalArrayOrObjectAndExtractTypesAndMarshal(contents, dataTemplate)
	if err != nil {
		log.WithFields(log.Fields{
			"Component": "Tester",
			"Scenario":  scenario,
			"Error":     err,
		}).Infof("failed to unmarshal request")
		return "", nil
	}
	return res, io.NopCloser(bytes.NewReader([]byte(res)))
}

func buildTemplateParams(
	scenario *types.MockScenario,
	overrides map[string]any,
) map[string]any {
	templateParams := make(map[string]any)
	for _, env := range os.Environ() {
		parts := strings.Split(env, "=")
		if len(parts) == 2 {
			templateParams[parts[0]] = parts[1]
		}
	}
	for k, v := range scenario.Request.ExamplePathParams {
		templateParams[k] = utils.RandRegex(v)
	}
	for k, v := range scenario.Request.ExampleQueryParams {
		templateParams[k] = utils.RandRegex(v)
	}
	for k, v := range scenario.Request.ExampleHeaders {
		templateParams[k] = utils.RandRegex(v)
	}
	// Find any params for query params and path variables
	for k, v := range scenario.ToKeyData().MatchGroups(scenario.Path) {
		templateParams[k] = v
	}
	for k, v := range overrides {
		templateParams[k] = v
	}
	return templateParams
}
