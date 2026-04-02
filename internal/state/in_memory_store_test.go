package state

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_InMemoryStateStore_CurrentStateNewSession(t *testing.T) {
	store := NewInMemoryStateStore()
	require.Equal(t, "", store.CurrentState("sess-1"))
}

func Test_InMemoryStateStore_TransitionFromInitial(t *testing.T) {
	store := NewInMemoryStateStore()
	// Brand-new session: transition from any from-state succeeds.
	err := store.Transition("sess-1", "created", "active")
	require.NoError(t, err)
	require.Equal(t, "active", store.CurrentState("sess-1"))
}

func Test_InMemoryStateStore_TransitionFromMatchingState(t *testing.T) {
	store := NewInMemoryStateStore()
	require.NoError(t, store.Transition("sess-1", "", "created"))
	require.NoError(t, store.Transition("sess-1", "created", "active"))
	require.Equal(t, "active", store.CurrentState("sess-1"))
}

func Test_InMemoryStateStore_TransitionFailsWrongState(t *testing.T) {
	store := NewInMemoryStateStore()
	require.NoError(t, store.Transition("sess-1", "", "active"))
	err := store.Transition("sess-1", "created", "deleted") // wrong from-state
	require.Error(t, err)
	require.Equal(t, "active", store.CurrentState("sess-1")) // state unchanged
}

func Test_InMemoryStateStore_SetAndGet(t *testing.T) {
	store := NewInMemoryStateStore()
	store.Set("sess-1", "orderId", "ord-42")
	val, ok := store.Get("sess-1", "orderId")
	require.True(t, ok)
	require.Equal(t, "ord-42", val)
}

func Test_InMemoryStateStore_GetMissingKey(t *testing.T) {
	store := NewInMemoryStateStore()
	val, ok := store.Get("sess-1", "missing")
	require.False(t, ok)
	require.Nil(t, val)
}

func Test_InMemoryStateStore_Reset(t *testing.T) {
	store := NewInMemoryStateStore()
	require.NoError(t, store.Transition("sess-1", "", "active"))
	store.Set("sess-1", "key", "value")
	store.Reset("sess-1")
	require.Equal(t, "", store.CurrentState("sess-1"))
	_, ok := store.Get("sess-1", "key")
	require.False(t, ok)
}

func Test_InMemoryStateStore_EmptySessionIDIsNoOp(t *testing.T) {
	store := NewInMemoryStateStore()
	require.Equal(t, "", store.CurrentState(""))
	require.NoError(t, store.Transition("", "a", "b"))
	store.Set("", "k", "v")
	_, ok := store.Get("", "k")
	require.False(t, ok)
}

func Test_InMemoryStateStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryStateStore()
	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			sid := "sess-concurrent"
			_ = store.Transition(sid, "", "state")
			store.Set(sid, "counter", n)
			store.CurrentState(sid)
			store.Get(sid, "counter")
		}(i)
	}
	wg.Wait()
}

func Test_InMemoryStateStore_MultipleSessionsIsolated(t *testing.T) {
	store := NewInMemoryStateStore()
	require.NoError(t, store.Transition("sess-A", "", "created"))
	require.NoError(t, store.Transition("sess-B", "", "deleted"))
	require.Equal(t, "created", store.CurrentState("sess-A"))
	require.Equal(t, "deleted", store.CurrentState("sess-B"))
	store.Reset("sess-A")
	require.Equal(t, "", store.CurrentState("sess-A"))
	require.Equal(t, "deleted", store.CurrentState("sess-B"))
}
