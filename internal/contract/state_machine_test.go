package contract

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/state"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
)

// Test_StateMachine_StateStoreIsInMemoryByDefault verifies the executor is wired correctly.
func Test_StateMachine_StateStoreIsInMemoryByDefault(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, _ := repository.NewFileAPIScenarioRepository(config)
	fixtureRepo, _ := repository.NewFileFixtureRepository(config)
	groupConfigRepo, _ := repository.NewFileGroupConfigRepository(config)

	executor := NewConsumerExecutor(config, scenarioRepo, fixtureRepo, groupConfigRepo)
	require.NotNil(t, executor.stateStore)
	_, ok := executor.stateStore.(*state.InMemoryStateStore)
	require.True(t, ok)
}

// Test_StateMachine_TransitionAppliedOnMethodAndStatusMatch verifies the state transitions
// when method and status match.
func Test_StateMachine_TransitionAppliedOnMethodAndStatusMatch(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, _ := repository.NewFileAPIScenarioRepository(config)
	fixtureRepo, _ := repository.NewFileFixtureRepository(config)
	groupConfigRepo, _ := repository.NewFileGroupConfigRepository(config)
	executor := NewConsumerExecutor(config, scenarioRepo, fixtureRepo, groupConfigRepo)

	scenario := &types.APIScenario{
		Method: types.Get,
		Name:   "test-order",
		Path:   "/api/orders/1",
		Response: types.APIResponse{StatusCode: 200},
		StateMachine: &types.ScenarioStateMachine{
			InitialState: "",
			Transitions: []types.StateTransition{
				{
					From:     "",
					To:       "created",
					OnMethod: "GET",
					OnStatus: 200,
				},
			},
		},
	}

	u, _ := url.Parse("https://example.com/api/orders/1")
	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: make(http.Header),
	}
	req.Header.Set(SessionIDHeader, "sess-001")

	executor.applyStateMachineTransitions(req, scenario, []byte(`{"id":1}`))
	require.Equal(t, "created", executor.stateStore.CurrentState("sess-001"))
}

// Test_StateMachine_NoTransitionWhenStatusMismatch checks that state is unchanged when
// the response status doesn't match the transition condition.
func Test_StateMachine_NoTransitionWhenStatusMismatch(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, _ := repository.NewFileAPIScenarioRepository(config)
	fixtureRepo, _ := repository.NewFileFixtureRepository(config)
	groupConfigRepo, _ := repository.NewFileGroupConfigRepository(config)
	executor := NewConsumerExecutor(config, scenarioRepo, fixtureRepo, groupConfigRepo)

	scenario := &types.APIScenario{
		Method: types.Get,
		Name:   "test-fail",
		Path:   "/api/orders/1",
		Response: types.APIResponse{StatusCode: 404}, // doesn't match OnStatus: 200
		StateMachine: &types.ScenarioStateMachine{
			Transitions: []types.StateTransition{
				{From: "", To: "created", OnStatus: 200},
			},
		},
	}

	u, _ := url.Parse("https://example.com/api/orders/1")
	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: make(http.Header),
	}
	req.Header.Set(SessionIDHeader, "sess-002")

	executor.applyStateMachineTransitions(req, scenario, []byte(`{}`))
	require.Equal(t, "", executor.stateStore.CurrentState("sess-002"))
}

// Test_StateMachine_NoTransitionWithoutSessionID verifies empty session is ignored.
func Test_StateMachine_NoTransitionWithoutSessionID(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, _ := repository.NewFileAPIScenarioRepository(config)
	fixtureRepo, _ := repository.NewFileFixtureRepository(config)
	groupConfigRepo, _ := repository.NewFileGroupConfigRepository(config)
	executor := NewConsumerExecutor(config, scenarioRepo, fixtureRepo, groupConfigRepo)

	scenario := &types.APIScenario{
		Method:   types.Get,
		Name:     "no-session",
		Path:     "/api/items/1",
		Response: types.APIResponse{StatusCode: 200},
		StateMachine: &types.ScenarioStateMachine{
			Transitions: []types.StateTransition{{From: "", To: "viewed", OnStatus: 200}},
		},
	}

	u, _ := url.Parse("https://example.com/api/items/1")
	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: http.Header{}, // No session header
	}

	// Should not panic, state store should remain empty
	executor.applyStateMachineTransitions(req, scenario, []byte(`{}`))
	require.Equal(t, "", executor.stateStore.CurrentState(""))
}

// Test_StateMachine_ExtractKeyFromResponse verifies that ExtractKey stores the value.
func Test_StateMachine_ExtractKeyFromResponse(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, _ := repository.NewFileAPIScenarioRepository(config)
	fixtureRepo, _ := repository.NewFileFixtureRepository(config)
	groupConfigRepo, _ := repository.NewFileGroupConfigRepository(config)
	executor := NewConsumerExecutor(config, scenarioRepo, fixtureRepo, groupConfigRepo)

	scenario := &types.APIScenario{
		Method:   types.Post,
		Name:     "create-order",
		Path:     "/api/orders",
		Response: types.APIResponse{StatusCode: 200},
		StateMachine: &types.ScenarioStateMachine{
			Transitions: []types.StateTransition{
				{
					From:       "",
					To:         "created",
					OnMethod:   "POST",
					OnStatus:   200,
					ExtractKey: "$.orderId",
				},
			},
		},
	}

	u, _ := url.Parse("https://example.com/api/orders")
	req := &http.Request{
		Method: "POST",
		URL:    u,
		Header: make(http.Header),
	}
	req.Header.Set(SessionIDHeader, "sess-extract-1")
	respBody := []byte(`{"orderId":"ord-99","status":"pending"}`)

	executor.applyStateMachineTransitions(req, scenario, respBody)

	require.Equal(t, "created", executor.stateStore.CurrentState("sess-extract-1"))
	val, ok := executor.stateStore.Get("sess-extract-1", "orderId")
	require.True(t, ok)
	require.Equal(t, "ord-99", val)
}

// Test_StateMachine_NilStateMachine verifies no panic when scenario has no state machine.
func Test_StateMachine_NilStateMachine(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, _ := repository.NewFileAPIScenarioRepository(config)
	fixtureRepo, _ := repository.NewFileFixtureRepository(config)
	groupConfigRepo, _ := repository.NewFileGroupConfigRepository(config)
	executor := NewConsumerExecutor(config, scenarioRepo, fixtureRepo, groupConfigRepo)

	scenario := &types.APIScenario{
		Method:       types.Get,
		Name:         "no-state-machine",
		Path:         "/api/simple",
		Response:     types.APIResponse{StatusCode: 200},
		StateMachine: nil, // no state machine
	}

	u, _ := url.Parse("https://example.com/api/simple")
	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: make(http.Header),
	}
	req.Header.Set(SessionIDHeader, "sess-nil")

	// Must not panic
	require.NotPanics(t, func() {
		executor.applyStateMachineTransitions(req, scenario, []byte(`{}`))
	})
}

// Test_StateMachine_MultipleTransitions checks that only the first matching transition fires.
func Test_StateMachine_MultipleTransitions(t *testing.T) {
	config := types.BuildTestConfig()
	scenarioRepo, _ := repository.NewFileAPIScenarioRepository(config)
	fixtureRepo, _ := repository.NewFileFixtureRepository(config)
	groupConfigRepo, _ := repository.NewFileGroupConfigRepository(config)
	executor := NewConsumerExecutor(config, scenarioRepo, fixtureRepo, groupConfigRepo)

	// Set session to "created" first
	require.NoError(t, executor.stateStore.Transition("sess-multi", "", "created"))

	scenario := &types.APIScenario{
		Method:   types.Put,
		Name:     "update-order",
		Path:     "/api/orders/1",
		Response: types.APIResponse{StatusCode: 200},
		StateMachine: &types.ScenarioStateMachine{
			Transitions: []types.StateTransition{
				{From: "created", To: "updated", OnStatus: 200},
				{From: "updated", To: "done", OnStatus: 200}, // should not fire this time
			},
		},
	}

	u, _ := url.Parse("https://example.com/api/orders/1")
	req := &http.Request{
		Method: "PUT",
		URL:    u,
		Header: make(http.Header),
	}
	req.Header.Set(SessionIDHeader, "sess-multi")

	executor.applyStateMachineTransitions(req, scenario, []byte(`{}`))
	require.Equal(t, "updated", executor.stateStore.CurrentState("sess-multi"))
}
