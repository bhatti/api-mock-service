package proxy

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
	scenarioRepository repository.MockScenarioRepository
	fixtureRepository  repository.MockFixtureRepository
}

// NewConsumerExecutor instantiates controller for updating mock-scenarios
func NewConsumerExecutor(
	mockScenarioRepository repository.MockScenarioRepository,
	fixtureRepository repository.MockFixtureRepository,
) *ConsumerExecutor {
	return &ConsumerExecutor{
		scenarioRepository: mockScenarioRepository,
		fixtureRepository:  fixtureRepository,
	}
}

// Execute request and replays stubbed response
func (p *ConsumerExecutor) Execute(c web.APIContext) (err error) {
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

	respBody, err := addMockResponse(
		c.Request(),
		c.Request().Header,
		c.Response().Header(),
		matchedScenario,
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

// this method is shared so it cannot be instance method
func addMockResponse(
	req *http.Request,
	reqHeaders http.Header,
	respHeaders http.Header,
	matchedScenario *types.MockScenario,
	scenarioRepository repository.MockScenarioRepository,
	fixtureRepository repository.MockFixtureRepository,
) (respBody []byte, err error) {
	var inBody []byte
	inBody, req.Body, err = utils.ReadAll(req.Body)
	if err == nil && len(inBody) > 0 {
		matchedScenario.Request.Contents = string(inBody)
		matchedScenario.Request.ExampleContents = string(inBody)
	}

	{
		// check request assertions
		overrides := make(map[string]any)
		templateParams, queryParams, reqHeaders := matchedScenario.Request.BuildTemplateParams(
			req,
			matchedScenario.ToKeyData().MatchGroups(matchedScenario.Path),
			overrides)
		reqContents, err := fuzz.UnmarshalArrayOrObject(inBody)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal request body for (%s) due to %w", matchedScenario.Name, err)
		}
		if err = matchedScenario.Request.Assert(queryParams, reqHeaders, reqContents, templateParams); err != nil {
			return nil, err
		}
	}

	for k, vals := range matchedScenario.Response.Headers {
		for _, val := range vals {
			respHeaders.Add(k, val)
		}
	}
	respHeaders.Add(types.MockScenarioName, matchedScenario.Name)
	respHeaders.Add(types.MockScenarioPath, matchedScenario.Path)
	respHeaders.Add(types.MockRequestCount, fmt.Sprintf("%d", matchedScenario.RequestCount))
	// Override wait time from request header
	if reqHeaders.Get(types.MockWaitBeforeReply) != "" {
		matchedScenario.WaitBeforeReply, _ = time.ParseDuration(reqHeaders.Get(types.MockWaitBeforeReply))
	}
	if matchedScenario.WaitBeforeReply > 0 {
		time.Sleep(matchedScenario.WaitBeforeReply)
	}
	// Override response status from request header
	if reqHeaders.Get(types.MockResponseStatus) != "" {
		matchedScenario.Response.StatusCode, _ = strconv.Atoi(reqHeaders.Get(types.MockResponseStatus))
	}
	if matchedScenario.Response.StatusCode == 0 {
		matchedScenario.Response.StatusCode = 200
	}
	// Build output from contents-file or contents property
	respBody = []byte(matchedScenario.Response.Contents)
	if matchedScenario.Response.ContentsFile != "" {
		respBody, err = fixtureRepository.Get(
			matchedScenario.Method,
			matchedScenario.Response.ContentsFile,
			matchedScenario.Path)
	}

	if err == nil {
		matchedScenario.Response.Contents = string(respBody)
		if matchedScenario.Request.Headers == nil {
			matchedScenario.Request.Headers = make(map[string]string)
		}
		for k, vals := range reqHeaders {
			for _, val := range vals {
				matchedScenario.Request.Headers[k] = val
			}
		}
		if matchedScenario.Response.Headers == nil {
			matchedScenario.Response.Headers = make(map[string][]string)
		}
		for k, vals := range respHeaders {
			matchedScenario.Response.Headers[k] = vals
		}
		err = scenarioRepository.SaveHistory(matchedScenario)
	}

	return
}
