package proxy

import (
	"errors"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"net/http"
	"strconv"
	"time"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/web"
)

// Player structure
type Player struct {
	scenarioRepository repository.MockScenarioRepository
	fixtureRepository  repository.MockFixtureRepository
}

// NewPlayer instantiates controller for updating mock-scenarios
func NewPlayer(
	mockScenarioRepository repository.MockScenarioRepository,
	fixtureRepository repository.MockFixtureRepository,
) *Player {
	return &Player{
		scenarioRepository: mockScenarioRepository,
		fixtureRepository:  fixtureRepository,
	}
}

// Handle request and replays stubbed response
func (p *Player) Handle(c web.APIContext) (err error) {
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

	respBody, err := addMockResponse(c.Request().Header, c.Response().Header(), matchedScenario, p.fixtureRepository)
	if err != nil {
		return err
	}
	return c.Blob(
		matchedScenario.Response.StatusCode,
		matchedScenario.Response.ContentType(),
		respBody)
}

func addMockResponse(
	reqHeader http.Header,
	respHeader http.Header,
	matchedScenario *types.MockScenario,
	fixtureRepository repository.MockFixtureRepository) (respBody []byte, err error) {
	for k, vals := range matchedScenario.Response.Headers {
		for _, val := range vals {
			respHeader.Add(k, val)
		}
	}
	respHeader.Add(types.MockScenarioName, matchedScenario.Name)
	respHeader.Add(types.MockScenarioPath, matchedScenario.Path)
	respHeader.Add(types.MockRequestCount, fmt.Sprintf("%d", matchedScenario.RequestCount))
	// Override wait time from request header
	if reqHeader.Get(types.MockWaitBeforeReply) != "" {
		matchedScenario.WaitBeforeReply, _ = time.ParseDuration(reqHeader.Get(types.MockWaitBeforeReply))
	}
	if matchedScenario.WaitBeforeReply > 0 {
		time.Sleep(matchedScenario.WaitBeforeReply)
	}
	// Override response status from request header
	if reqHeader.Get(types.MockResponseStatus) != "" {
		matchedScenario.Response.StatusCode, _ = strconv.Atoi(reqHeader.Get(types.MockResponseStatus))
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
	return
}
