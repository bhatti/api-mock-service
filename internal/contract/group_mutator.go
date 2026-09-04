package contract

import (
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
)

// GroupMutator generates sequence-level mutations for a group of scenarios.
// While ContractMutator mutates individual request payloads, GroupMutator
// mutates the execution order and composition of scenario sequences.
type GroupMutator struct {
	scenarios []*types.APIScenario
	maxPerStrategy int
}

// NewGroupMutator creates a new group mutator.
func NewGroupMutator(scenarios []*types.APIScenario) *GroupMutator {
	return &GroupMutator{
		scenarios:      scenarios,
		maxPerStrategy: 10,
	}
}

// GenerateSequenceMutations returns mutated groups of scenarios.
// Each inner slice is a complete sequence to execute as a unit.
func (gm *GroupMutator) GenerateSequenceMutations() [][]*types.APIScenario {
	if len(gm.scenarios) < 2 {
		return nil
	}

	var result [][]*types.APIScenario
	result = append(result, gm.reversed()...)
	result = append(result, gm.skipStep()...)
	result = append(result, gm.duplicateStep()...)
	result = append(result, gm.methodSwap()...)
	return result
}

func (gm *GroupMutator) reversed() [][]*types.APIScenario {
	reversed := make([]*types.APIScenario, len(gm.scenarios))
	for i, s := range gm.scenarios {
		clone := *s
		clone.Name = s.Name + "-seq-reversed"
		reversed[len(gm.scenarios)-1-i] = &clone
	}
	return [][]*types.APIScenario{reversed}
}

func (gm *GroupMutator) skipStep() [][]*types.APIScenario {
	var groups [][]*types.APIScenario
	limit := len(gm.scenarios)
	if limit > gm.maxPerStrategy {
		limit = gm.maxPerStrategy
	}
	for i := 0; i < limit; i++ {
		group := make([]*types.APIScenario, 0, len(gm.scenarios)-1)
		for j, s := range gm.scenarios {
			if j == i {
				continue
			}
			clone := *s
			clone.Name = s.Name + "-seq-skip-" + gm.scenarios[i].SafeName()
			group = append(group, &clone)
		}
		groups = append(groups, group)
	}
	return groups
}

func (gm *GroupMutator) duplicateStep() [][]*types.APIScenario {
	var groups [][]*types.APIScenario
	limit := len(gm.scenarios)
	if limit > gm.maxPerStrategy {
		limit = gm.maxPerStrategy
	}
	for i := 0; i < limit; i++ {
		group := make([]*types.APIScenario, 0, len(gm.scenarios)+1)
		for j, s := range gm.scenarios {
			clone := *s
			if j == i {
				clone.Name = s.Name + "-seq-dup-first"
			}
			group = append(group, &clone)
			if j == i {
				clone2 := *s
				clone2.Name = s.Name + "-seq-dup-second"
				group = append(group, &clone2)
			}
		}
		groups = append(groups, group)
	}
	return groups
}

func (gm *GroupMutator) methodSwap() [][]*types.APIScenario {
	swaps := map[types.MethodType]types.MethodType{
		types.Put:   types.Patch,
		types.Patch: types.Put,
		types.Get:   types.Delete,
	}

	var groups [][]*types.APIScenario
	count := 0
	for i, s := range gm.scenarios {
		if count >= gm.maxPerStrategy {
			break
		}
		swappedMethod, ok := swaps[s.Method]
		if !ok {
			continue
		}
		group := make([]*types.APIScenario, len(gm.scenarios))
		for j, orig := range gm.scenarios {
			clone := *orig
			if j == i {
				clone.Method = swappedMethod
				clone.Name = orig.Name + "-seq-method-" + string(swappedMethod)
				clone.Response.StatusCode = fuzz.RandIntMinMax(400, 409)
			}
			group[j] = &clone
		}
		groups = append(groups, group)
		count++
	}
	return groups
}
