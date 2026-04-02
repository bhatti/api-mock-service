// SPDX-License-Identifier: MIT

package state

import (
	"fmt"
	"sync"
)

// InMemoryStateStore is a goroutine-safe in-process implementation of StateStore.
// It holds all session state in memory; state is lost on process restart.
type InMemoryStateStore struct {
	mu     sync.RWMutex
	states map[string]string            // sessionID → current state name
	data   map[string]map[string]any    // sessionID → key → value
}

// NewInMemoryStateStore returns a ready-to-use InMemoryStateStore.
func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{
		states: make(map[string]string),
		data:   make(map[string]map[string]any),
	}
}

// CurrentState implements StateStore.
func (s *InMemoryStateStore) CurrentState(sessionID string) string {
	if sessionID == "" {
		return ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.states[sessionID]
}

// Transition implements StateStore.
// If the session has no recorded state (empty string) it is treated as matching any fromState,
// so the first transition always succeeds.
func (s *InMemoryStateStore) Transition(sessionID, fromState, toState string) error {
	if sessionID == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.states[sessionID]
	// Allow transition when session is brand-new (current == "") OR state matches
	if current != "" && current != fromState {
		return fmt.Errorf("state transition rejected for session %q: current=%q want-from=%q",
			sessionID, current, fromState)
	}
	s.states[sessionID] = toState
	return nil
}

// Set implements StateStore.
func (s *InMemoryStateStore) Set(sessionID, key string, val any) {
	if sessionID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data[sessionID] == nil {
		s.data[sessionID] = make(map[string]any)
	}
	s.data[sessionID][key] = val
}

// Get implements StateStore.
func (s *InMemoryStateStore) Get(sessionID, key string) (any, bool) {
	if sessionID == "" {
		return nil, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m, ok := s.data[sessionID]; ok {
		v, found := m[key]
		return v, found
	}
	return nil, false
}

// Reset implements StateStore.
func (s *InMemoryStateStore) Reset(sessionID string) {
	if sessionID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, sessionID)
	delete(s.data, sessionID)
}
