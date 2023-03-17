package contract

import (
	"errors"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/web"
)

// ConsumerExecutor structure
type ConsumerExecutor struct {
	scenarioRepository repository.APIScenarioRepository
	fixtureRepository  repository.APIFixtureRepository
}

// NewConsumerExecutor instantiates controller for updating api-scenarios
func NewConsumerExecutor(
	scenarioRepository repository.APIScenarioRepository,
	fixtureRepository repository.APIFixtureRepository,
) *ConsumerExecutor {
	return &ConsumerExecutor{
		scenarioRepository: scenarioRepository,
		fixtureRepository:  fixtureRepository,
	}
}

// Execute request and replays stubbed response
func (p *ConsumerExecutor) Execute(c web.APIContext) (err error) {
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
	matchedScenario, err := p.scenarioRepository.Lookup(key, overrides)
	if err != nil {
		var validationErr *types.ValidationError
		var notFoundErr *types.NotFoundError
		if errors.As(err, &validationErr) {
			return c.String(400, err.Error())
		} else if errors.As(err, &notFoundErr) {
			return c.String(404, err.Error())
		}
		return err
	}

	respBody, err := AddMockResponse(
		c.Request(),
		c.Request().Header,
		c.Response().Header(),
		matchedScenario,
		started,
		time.Now(),
		p.scenarioRepository,
		p.fixtureRepository,
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
	scenarioRepository repository.APIScenarioRepository,
	fixtureRepository repository.APIFixtureRepository,
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
	respHeaders.Add(types.MockScenarioName, scenario.Name)
	respHeaders.Add(types.MockScenarioPath, scenario.Path)
	respHeaders.Add(types.MockRequestCount, fmt.Sprintf("%d", scenario.RequestCount))
	// Override wait time from request header
	if reqHeaders.Get(types.MockWaitBeforeReply) != "" {
		scenario.WaitBeforeReply, _ = time.ParseDuration(reqHeaders.Get(types.MockWaitBeforeReply))
	}
	if scenario.WaitBeforeReply > 0 {
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

		err = scenarioRepository.SaveHistory(scenario, req.URL, started, ended)
	}

	return
}
