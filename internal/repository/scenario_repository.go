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

	// Save MockScenario
	Save(scenario *types.MockScenario) (err error)

	// Delete removes a mock senario
	Delete(
		method types.MethodType,
		scenarioName string,
		path string) error

	// LookupAll finds matching scenarios
	LookupAll(key *types.MockScenarioKeyData) ([]*types.MockScenarioKeyData, int)

	// Lookup finds top matching scenario that hasn't been used recently
	Lookup(target *types.MockScenarioKeyData) (*types.MockScenario, error)

	// ListScenarioKeyData returns keys for all scenarios
	ListScenarioKeyData() []*types.MockScenarioKeyData
}
