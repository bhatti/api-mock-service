package proxy

import (
	"errors"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"strconv"
	"time"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/web"
)

// MockRequestCount header
const MockRequestCount = "Mock-Request-Count"

// MockResponseStatus header
const MockResponseStatus = "Mock-Response-Status"

// MockWaitBeforeReply header
const MockWaitBeforeReply = "Mock-Wait-Before-Reply"

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
	key, err := web.BuildMockScenarioKeyData(c)
	if err != nil {
		return err
	}

	matchedScenario, err := p.scenarioRepository.Lookup(key)
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

	for k, vals := range matchedScenario.Response.Headers {
		for _, val := range vals {
			c.Response().Header().Add(k, val)
		}
	}
	c.Response().Header().Add(types.MockScenarioName, matchedScenario.Name)
	c.Response().Header().Add(MockRequestCount, fmt.Sprintf("%d", matchedScenario.RequestCount))
	// Override wait time from request header
	if c.Request().Header.Get(MockWaitBeforeReply) != "" {
		matchedScenario.WaitBeforeReply, _ = time.ParseDuration(c.Request().Header.Get(MockWaitBeforeReply))
	}
	if matchedScenario.WaitBeforeReply > 0 {
		time.Sleep(matchedScenario.WaitBeforeReply)
	}
	// Override response status from request header
	if c.Request().Header.Get(MockResponseStatus) != "" {
		matchedScenario.Response.StatusCode, _ = strconv.Atoi(c.Request().Header.Get(MockResponseStatus))
	}
	if matchedScenario.Response.StatusCode == 0 {
		matchedScenario.Response.StatusCode = 200
	}

	// Build output from contents-file or contents property
	var respBody []byte
	if matchedScenario.Response.ContentsFile != "" {
		respBody, err = p.fixtureRepository.Get(
			matchedScenario.Method,
			matchedScenario.Response.ContentsFile,
			matchedScenario.Path)
		if err != nil {
			return err
		}
	} else {
		respBody = []byte(matchedScenario.Response.Contents)
	}
	return c.Blob(
		matchedScenario.Response.StatusCode,
		matchedScenario.Response.ContentType,
		respBody)
}
