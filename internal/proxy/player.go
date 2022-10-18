package proxy

import (
	"io"
	"time"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
)

// MockScenarioName header
const MockScenarioName = "Mock-Scenario"

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
	reqBody := []byte{}

	if c.Request().Body != nil {
		reqBody, err = io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
	}

	key := &types.MockScenarioKeyData{
		Method:      types.MethodType(c.Request().Method),
		Name:        c.Request().Header.Get(MockScenarioName),
		Path:        c.Request().URL.Path,
		QueryParams: c.Request().URL.RawQuery,
		ContentType: c.Request().Header.Get(types.ContentTypeHeader),
		Contents:    string(reqBody),
	}

	matchedScenario, err := p.scenarioRepository.Lookup(key)
	if err != nil {
		return err
	}

	for k, vals := range matchedScenario.Response.Headers {
		for _, val := range vals {
			c.Response().Header().Add(k, val)
		}
	}
	c.Response().Header().Add(MockScenarioName, matchedScenario.Name)
	if matchedScenario.WaitBeforeReply > 0 {
		time.Sleep(matchedScenario.WaitBeforeReply)
	}
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
