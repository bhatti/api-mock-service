// SPDX-License-Identifier: MIT

// Package state provides session-scoped state management for stateful scenario testing.
// It enables CREATE → READ → DELETE and other ordered workflow scenarios by tracking
// the current state of each session identified by X-Session-ID.
package state

// StateStore manages session state and key-value data for stateful scenario testing.
type StateStore interface {
	// CurrentState returns the current state for the given session.
	// Returns "" if the session is unknown (treated as initial state).
	CurrentState(sessionID string) string

	// Transition moves a session from one state to another.
	// Returns an error if the session is not in fromState.
	Transition(sessionID, fromState, toState string) error

	// Set stores an arbitrary value under key for the session.
	Set(sessionID, key string, val any)

	// Get retrieves a value stored under key for the session.
	Get(sessionID, key string) (any, bool)

	// Reset clears all state for a session (useful for test teardown).
	Reset(sessionID string)
}
