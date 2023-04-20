package contract

import (
	"errors"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/web"
)

// ConsumerExecutor structure
type ConsumerExecutor struct {
	config                *types.Configuration
	scenarioRepository    repository.APIScenarioRepository
	fixtureRepository     repository.APIFixtureRepository
	groupConfigRepository repository.GroupConfigRepository
}

// NewConsumerExecutor instantiates controller for updating api-scenarios
func NewConsumerExecutor(
	config *types.Configuration,
	scenarioRepository repository.APIScenarioRepository,
	fixtureRepository repository.APIFixtureRepository,
	groupConfigRepository repository.GroupConfigRepository,
) *ConsumerExecutor {
	return &ConsumerExecutor{
		config:                config,
		scenarioRepository:    scenarioRepository,
		fixtureRepository:     fixtureRepository,
		groupConfigRepository: groupConfigRepository,
	}
}

// Execute request and replays stubbed response
func (cx *ConsumerExecutor) Execute(c web.APIContext) (err error) {
	started := time.Now()
	key, err := web.BuildMockScenarioKeyData(c.Request())
	if err != nil {
		return err
	}

	overrides := make(map[string]any)
	for k, v := range c.QueryParams() {
		overrides[k] = v[0]
	}
	if form, err := c.FormParams(); err == nil {
		for k, v := range form {
			overrides[k] = v[0]
		}
	}
	matchedScenario, matchErr := cx.scenarioRepository.Lookup(key, overrides)
	if matchErr != nil {
		var validationErr *types.ValidationError
		var notFoundErr *types.NotFoundError
		if errors.As(matchErr, &validationErr) {
			return c.String(400, matchErr.Error())
		} else if errors.As(matchErr, &notFoundErr) {
			return c.String(404, matchErr.Error())
		}
		return matchErr
	}

	respBody, err := AddMockResponse(
		c.Request(),
		c.Request().Header,
		c.Response().Header(),
		matchedScenario,
		started,
		time.Now(),
		cx.config,
		cx.scenarioRepository,
		cx.fixtureRepository,
		cx.groupConfigRepository,
	)
	if err != nil {
		return err
	}
	return c.Blob(
		matchedScenario.Response.StatusCode,
		matchedScenario.Response.ContentType(""),
		respBody)
}

// AddMockResponse method is shared so it cannot be instance method
func AddMockResponse(
	req *http.Request,
	reqHeaders http.Header,
	respHeaders http.Header,
	scenario *types.APIScenario,
	started time.Time,
	ended time.Time,
	config *types.Configuration,
	scenarioRepository repository.APIScenarioRepository,
	fixtureRepository repository.APIFixtureRepository,
	groupConfigRepository repository.GroupConfigRepository,
) (respBody []byte, err error) {
	var inBody []byte
	inBody, req.Body, err = utils.ReadAll(req.Body)
	if err == nil && len(inBody) > 0 {
		scenario.Request.Contents = string(inBody)
		scenario.Request.ExampleContents = string(inBody)
	}

	{
		// check request assertions
		templateParams, queryParams, postParams, reqHeaders := scenario.Request.BuildTemplateParams(
			req,
			scenario.ToKeyData().MatchGroups(scenario.Path),
			reqHeaders,
			make(map[string]any))
		reqContents, err := fuzz.UnmarshalArrayOrObject(inBody)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal request body for (%s) due to %w", scenario.Name, err)
		}
		if err = scenario.Request.Assert(queryParams, postParams, reqHeaders, reqContents, templateParams); err != nil {
			return nil, err
		}
	}

	for k, vals := range scenario.Response.Headers {
		for _, val := range vals {
			respHeaders.Add(k, val)
		}
	}
	respHeaders.Add(types.MockScenarioHeader, scenario.Name)
	respHeaders.Add(types.MockScenarioPath, scenario.Path)
	respHeaders.Add(types.MockRequestCount, fmt.Sprintf("%d", scenario.RequestCount))

	// Embedding this check to return chaos response based on group config
	if b := CheckChaosForScenarioGroup(groupConfigRepository, scenario, respHeaders); b != nil {
		return b, nil
	}

	if config.RecordOnly || req.Header.Get(types.MockRecordMode) == types.MockRecordModeEnabled {
		log.WithFields(log.Fields{
			"Component":        "ConsumerExecutor-AddMockResponse",
			"ConfigRecordOnly": config.RecordOnly,
			"Host":             req.Host,
			"Path":             req.URL,
			"Method":           req.Method,
			"Scenario":         scenario.Name,
			"Group":            scenario.Group,
			//"Headers":          req.Header,
		}).Infof("proxy server skipped local lookup due to record-mode")
		return nil, types.NewNotFoundError("proxy server skipping local lookup due to record-mode")
	}

	// Override wait time from request header
	if reqHeaders.Get(types.MockWaitBeforeReply) != "" {
		scenario.WaitBeforeReply, _ = time.ParseDuration(reqHeaders.Get(types.MockWaitBeforeReply))
	}
	if scenario.WaitBeforeReply > 0 {
		log.WithFields(log.Fields{
			"Component": "ConsumerExecutor-AddMockResponse",
			"Scenario":  scenario.Name,
			"Group":     scenario.Group,
			"Delay":     scenario.WaitBeforeReply,
		}).Infof("scenario sleep wait")
		time.Sleep(scenario.WaitBeforeReply)
	}
	// Override response status from request header
	if reqHeaders.Get(types.MockResponseStatus) != "" {
		scenario.Response.StatusCode, _ = strconv.Atoi(reqHeaders.Get(types.MockResponseStatus))
	}
	if scenario.Response.StatusCode == 0 {
		scenario.Response.StatusCode = 200
	}
	// Build output from contents-file or contents property
	respBody = []byte(scenario.Response.Contents)
	if scenario.Response.ContentsFile != "" {
		respBody, err = fixtureRepository.Get(
			scenario.Method,
			scenario.Response.ContentsFile,
			scenario.Path)
	}
	respHeaders.Set(types.ContentLengthHeader, fmt.Sprintf("%d", len(respBody)))

	if err == nil {
		scenario.Response.Contents = string(respBody)
		if scenario.Request.Headers == nil {
			scenario.Request.Headers = make(map[string]string)
		}
		for k, vals := range reqHeaders {
			for _, val := range vals {
				scenario.Request.Headers[k] = val
			}
		}
		if scenario.Response.Headers == nil {
			scenario.Response.Headers = make(map[string][]string)
		}
		for k, vals := range respHeaders {
			scenario.Response.Headers[k] = vals
		}

		err = scenarioRepository.SaveHistory(scenario, req.URL.String(), started, ended)
	}

	return
}

// CheckChaosForScenarioGroup helper method
func CheckChaosForScenarioGroup(
	groupConfigRepository repository.GroupConfigRepository,
	scenario *types.APIScenario,
	respHeaders http.Header) []byte {
	if groupConfig, err := groupConfigRepository.Load(scenario.Group); err == nil {
		respHeaders.Add(types.MockChaosEnabled, fmt.Sprintf("%v", groupConfig.ChaosEnabled))
		delay := groupConfig.GetDelayLatency()
		if delay > 0 {
			log.WithFields(log.Fields{
				"Component":   "ConsumerExecutor-AddMockResponse",
				"Scenario":    scenario.Name,
				"Group":       scenario.Group,
				"GroupConfig": groupConfig,
				"Delay":       delay,
			}).Debugf("chaos sleep wait")
			time.Sleep(delay)
		}
		status := groupConfig.GetHTTPStatus()
		if status >= 300 {
			scenario.Response.StatusCode = status
			return []byte("injected fault from consumer-executor")
		}
	}
	return nil
}
