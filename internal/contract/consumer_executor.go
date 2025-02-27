package contract

import (
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
	overrides := make(map[string]any)
	for k, v := range c.Request().Header {
		overrides[k] = v[0]
	}
	for k, v := range c.QueryParams() {
		overrides[k] = v[0]
	}
	if form, err := c.FormParams(); err == nil {
		for k, v := range form {
			overrides[k] = v[0]
		}
	}

	key, err := web.BuildMockScenarioKeyData(c.Request())
	if err != nil {
		return web.HandleError(c, err)
	}
	matchedScenario, respBody, _, err := cx.ExecuteWithKey(c.Request(), c.Response().Header(), key, overrides)
	if err != nil {
		return web.HandleError(c, err)
	}
	return c.Blob(
		matchedScenario.Response.StatusCode,
		matchedScenario.Response.ContentType(""),
		respBody)
	return err
}

// ExecuteWithKey request and replays stubbed response
func (cx *ConsumerExecutor) ExecuteWithKey(
	req *http.Request,
	respHeaders http.Header,
	key *types.APIKeyData,
	overrides map[string]any) (matchedScenario *types.APIScenario, respBytes []byte,
	sharedVariables map[string]any, err error) {
	started := time.Now()

	matchedScenario, err = cx.scenarioRepository.Lookup(key, overrides)
	if err != nil {
		return
	}
	if matchedScenario.NextRequest != "" && len(matchedScenario.Response.AddSharedVariables) > 0 {
		nextScenario, err := cx.scenarioRepository.LookupByName(matchedScenario.NextRequest, overrides)
		if err != nil {
			return nil, nil, nil,
				fmt.Errorf("next request key: %s not found: %s", key.Name, err)
		}
		_, sharedVariables, err = cx.execute(req, respHeaders, nextScenario, started)
		if err != nil {
			return nil, nil, nil, err
		}
		for k, v := range sharedVariables {
			if strVal, ok := v.(string); ok && matchedScenario.Request.Variables[k] == "" {
				matchedScenario.Request.Variables[k] = strVal
				req.Header.Set(k, strVal)
			}
		}
	}
	if err != nil {
		return nil, nil, nil, err
	}

	respBytes, sharedVariables, err = cx.execute(req, respHeaders, matchedScenario, started)
	return
}

func (cx *ConsumerExecutor) execute(
	req *http.Request,
	respHeaders http.Header,
	matchedScenario *types.APIScenario,
	started time.Time) ([]byte, map[string]any, error) {
	return AddMockResponse(
		req,
		req.Header,
		respHeaders,
		matchedScenario,
		started,
		time.Now(),
		cx.config,
		cx.scenarioRepository,
		cx.fixtureRepository,
		cx.groupConfigRepository,
	)
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
) (respBody []byte, sharedVariables map[string]any, err error) {
	var inBody []byte
	inBody, req.Body, err = utils.ReadAll(req.Body)
	if err == nil && len(inBody) > 0 {
		scenario.Request.Contents = string(inBody)
		scenario.Request.ExampleContents = string(inBody)
	}
	sharedVariables = make(map[string]any)

	respHeaders.Add(types.MockScenarioHeader, scenario.Name)
	respHeaders.Add(types.MockScenarioPath, scenario.Path)
	respHeaders.Add(types.MockRequestCount, fmt.Sprintf("%d", scenario.RequestCount))

	{
		// check request assertions
		templateParams, queryParams, postParams, reqHeaders := scenario.Request.BuildTemplateParams(
			req,
			scenario.ToKeyData().MatchGroups(scenario.Path),
			reqHeaders,
			make(map[string]any))
		reqContents, err := fuzz.UnmarshalArrayOrObject(inBody)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal request body for (%s) due to %w", scenario.Name, err)
		}
		if err = scenario.Request.Assert(queryParams, postParams, reqHeaders, reqContents, templateParams); err != nil {
			return nil, nil, err
		}
	}

	for k, vals := range scenario.Response.Headers {
		for _, val := range vals {
			respHeaders.Add(k, val)
		}
	}

	// Embedding this check to return chaos response based on group config
	if b := CheckChaosForScenarioGroup(groupConfigRepository, scenario, respHeaders); b != nil {
		return b, sharedVariables, nil
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
		return nil, nil, types.NewNotFoundError("proxy server skipping local lookup due to record-mode")
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
		if code, err := strconv.Atoi(reqHeaders.Get(types.MockResponseStatus)); err == nil {
			scenario.Response.StatusCode = code
		}
	}

	//if scenario.Response.StatusCode == 0 {
	//	scenario.Response.StatusCode = 200
	//}

	// Build output from contents-file or contents property
	respBody = []byte(scenario.Response.Contents)
	if scenario.Response.ContentsFile != "" {
		respBody, err = fixtureRepository.Get(
			scenario.Method,
			scenario.Response.ContentsFile,
			scenario.Path)
	}
	respHeaders.Set(types.ContentLengthHeader, fmt.Sprintf("%d", len(respBody)))

	_ = handleSharedVariables(scenario, respBody, map[string]any{},
		groupConfigRepository.Variables(scenario.Group), sharedVariables, respHeaders)

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
