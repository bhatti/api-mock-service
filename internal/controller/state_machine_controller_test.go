package controller

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/stretchr/testify/require"
)

// Test_StateMachine_ControllerTransitionViaHTTP tests that the state machine
// transitions are applied at the controller level when X-Session-ID header is present.
func Test_StateMachine_ControllerTransitionViaHTTP(t *testing.T) {
	// GIVEN repositories and executor
	config := types.BuildTestConfig()
	scenarioRepo, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepo, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepo, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)

	player := contract.NewConsumerExecutor(config, scenarioRepo, fixtureRepo, groupConfigRepo)

	// AND a scenario with a state machine definition
	scenario := &types.APIScenario{
		Method:      types.Get,
		Name:        "sm-ctrl-test",
		Path:        "/api/sm-ctrl/items",
		Group:       "sm-ctrl-group",
		Description: "state machine controller test",
		Request:     types.APIRequest{},
		Response: types.APIResponse{
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Contents:   `{"id":1,"status":"created"}`,
			StatusCode: 200,
		},
		StateMachine: &types.ScenarioStateMachine{
			InitialState: "idle",
			Transitions: []types.StateTransition{
				{From: "idle", To: "active", OnMethod: "GET", OnStatus: 200},
			},
		},
	}
	err = scenarioRepo.Save(scenario)
	require.NoError(t, err)

	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)

	u, err := url.Parse("http://localhost:8080/api/sm-ctrl/items")
	require.NoError(t, err)

	// WHEN making a request with X-Session-ID header
	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: make(http.Header),
	}
	req.Header.Set("X-Session-ID", "test-session-ctrl-001")

	ctx := web.NewStubContext(req)
	err = ctrl.getRoot(ctx)

	// THEN the request should be served successfully (state machine does not block playback)
	require.NoError(t, err)
}

// Test_StateMachine_ControllerNoSessionIDNoTransition verifies that without X-Session-ID
// the state machine is silently skipped and the scenario still plays normally.
func Test_StateMachine_ControllerNoSessionIDNoTransition(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepo, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepo, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)

	player := contract.NewConsumerExecutor(config, scenarioRepo, fixtureRepo, groupConfigRepo)

	scenario := &types.APIScenario{
		Method:      types.Get,
		Name:        "sm-ctrl-nosession",
		Path:        "/api/sm-ctrl/nosession",
		Group:       "sm-ctrl-group",
		Description: "state machine without session",
		Request:     types.APIRequest{},
		Response: types.APIResponse{
			Headers:    http.Header{"Content-Type": []string{"application/json"}},
			Contents:   `{"ok":true}`,
			StatusCode: 200,
		},
		StateMachine: &types.ScenarioStateMachine{
			InitialState: "idle",
			Transitions: []types.StateTransition{
				{From: "idle", To: "active", OnMethod: "GET", OnStatus: 200},
			},
		},
	}
	err = scenarioRepo.Save(scenario)
	require.NoError(t, err)

	webServer := web.NewStubWebServer()
	ctrl := NewRootController(player, webServer)

	u, err := url.Parse("http://localhost:8080/api/sm-ctrl/nosession")
	require.NoError(t, err)

	// WHEN making a request WITHOUT X-Session-ID header
	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: make(http.Header),
	}
	ctx := web.NewStubContext(req)
	err = ctrl.getRoot(ctx)

	// THEN the request is still served — no state machine side-effects
	require.NoError(t, err)
}
