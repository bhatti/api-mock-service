package contract

import (
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func threeScenarioGroup() []*types.APIScenario {
	return []*types.APIScenario{
		{Name: "create-user", Method: types.Post, Path: "/users"},
		{Name: "get-user", Method: types.Get, Path: "/users/{id}"},
		{Name: "update-user", Method: types.Put, Path: "/users/{id}"},
	}
}

func TestGroupMutator_SingleScenario(t *testing.T) {
	gm := NewGroupMutator([]*types.APIScenario{{Name: "only"}})
	result := gm.GenerateSequenceMutations()
	assert.Empty(t, result, "single scenario should produce no sequence mutations")
}

func TestGroupMutator_EmptyInput(t *testing.T) {
	gm := NewGroupMutator(nil)
	assert.Empty(t, gm.GenerateSequenceMutations())
}

func TestGroupMutator_Reversed(t *testing.T) {
	scenarios := threeScenarioGroup()
	gm := NewGroupMutator(scenarios)
	groups := gm.reversed()

	require.Len(t, groups, 1)
	reversed := groups[0]
	require.Len(t, reversed, 3)
	assert.Contains(t, reversed[0].Name, "update-user")
	assert.Contains(t, reversed[1].Name, "get-user")
	assert.Contains(t, reversed[2].Name, "create-user")
}

func TestGroupMutator_SkipStep(t *testing.T) {
	scenarios := threeScenarioGroup()
	gm := NewGroupMutator(scenarios)
	groups := gm.skipStep()

	require.Len(t, groups, 3)
	for _, group := range groups {
		assert.Len(t, group, 2, "each skip group should have N-1 scenarios")
	}

	firstGroup := groups[0]
	for _, s := range firstGroup {
		assert.NotContains(t, s.Name, "create-user-seq-skip")
	}
}

func TestGroupMutator_DuplicateStep(t *testing.T) {
	scenarios := threeScenarioGroup()
	gm := NewGroupMutator(scenarios)
	groups := gm.duplicateStep()

	require.Len(t, groups, 3)
	for _, group := range groups {
		assert.Len(t, group, 4, "each duplicate group should have N+1 scenarios")
	}

	firstGroup := groups[0]
	dupCount := 0
	for _, s := range firstGroup {
		if s.Name == "create-user-seq-dup-first" || s.Name == "create-user-seq-dup-second" {
			dupCount++
		}
	}
	assert.Equal(t, 2, dupCount, "first group should have 2 copies of first scenario")
}

func TestGroupMutator_MethodSwap(t *testing.T) {
	scenarios := threeScenarioGroup()
	gm := NewGroupMutator(scenarios)
	groups := gm.methodSwap()

	require.NotEmpty(t, groups)
	hasSwapped := false
	for _, group := range groups {
		for _, s := range group {
			if s.Method == types.Delete || s.Method == types.Patch {
				hasSwapped = true
			}
		}
	}
	assert.True(t, hasSwapped, "should have swapped methods")
}

func TestGroupMutator_MethodSwap_PutToPatch(t *testing.T) {
	scenarios := threeScenarioGroup()
	gm := NewGroupMutator(scenarios)
	groups := gm.methodSwap()

	hasPatch := false
	for _, group := range groups {
		for _, s := range group {
			if s.Method == types.Patch {
				hasPatch = true
				assert.Contains(t, s.Name, "seq-method-PATCH")
			}
		}
	}
	assert.True(t, hasPatch, "PUT should be swapped to PATCH")
}

func TestGroupMutator_GenerateSequenceMutations_Total(t *testing.T) {
	scenarios := threeScenarioGroup()
	gm := NewGroupMutator(scenarios)
	groups := gm.GenerateSequenceMutations()

	require.NotEmpty(t, groups)
	assert.True(t, len(groups) >= 7, "expected at least 7 mutation groups (1 reversed + 3 skip + 3 dup)")
}

func TestGroupMutator_DoesNotMutatOriginal(t *testing.T) {
	scenarios := threeScenarioGroup()
	originalNames := make([]string, len(scenarios))
	for i, s := range scenarios {
		originalNames[i] = s.Name
	}

	gm := NewGroupMutator(scenarios)
	_ = gm.GenerateSequenceMutations()

	for i, s := range scenarios {
		assert.Equal(t, originalNames[i], s.Name, "original scenarios should not be modified")
	}
}
