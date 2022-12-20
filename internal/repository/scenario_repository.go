package repository

import (
	"github.com/bhatti/api-mock-service/internal/types"
	"io"
)

// MockScenarioRepository defines data store for mock-scenarios
type MockScenarioRepository interface {
	// GetScenariosNames returns mock scenarios for given Method and Path
	GetScenariosNames(
		method types.MethodType,
		path string) ([]string, error)

	// SaveRaw saves raw data assuming to be yaml format
	SaveRaw(input io.ReadCloser) (err error)

	// SaveYaml saves as yaml data
	SaveYaml(key *types.MockScenarioKeyData, payload []byte) (err error)

	// LoadRaw loads matching scenario
	LoadRaw(method types.MethodType, name string, path string) (b []byte, err error)

	// Save MockScenario
	Save(scenario *types.MockScenario) (err error)

	// Delete removes a mock scenario
	Delete(
		method types.MethodType,
		scenarioName string,
		path string) error

	// LookupAll finds matching scenarios
	LookupAll(key *types.MockScenarioKeyData) ([]*types.MockScenarioKeyData, int)

	// LookupAllByGroup finds matching scenarios by group
	LookupAllByGroup(group string) []*types.MockScenarioKeyData

	// Lookup finds top matching scenario that hasn't been used recently
	Lookup(target *types.MockScenarioKeyData, data map[string]any) (*types.MockScenario, error)

	// ListScenarioKeyData returns keys for all scenarios
	ListScenarioKeyData(group string) []*types.MockScenarioKeyData
}
