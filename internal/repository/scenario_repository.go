package repository

import (
	"github.com/bhatti/api-mock-service/internal/types"
	"io"
)

// APIScenarioRepository defines data store for api-scenarios
type APIScenarioRepository interface {
	// HistoryNames returns list of api scenarios names
	HistoryNames(group string) []string

	// SaveHistory saves history APIScenario
	SaveHistory(scenario *types.APIScenario, url string) (err error)

	// LoadHistory loads scenario
	LoadHistory(name string) (*types.APIScenario, error)

	// GetGroups returns api scenarios groups
	GetGroups() []string

	// GetScenariosNames returns api scenarios for given Method and Path
	GetScenariosNames(
		method types.MethodType,
		path string) ([]string, error)

	// SaveRaw saves raw data assuming to be yaml format
	SaveRaw(input io.ReadCloser) (err error)

	// SaveYaml saves as yaml data
	SaveYaml(key *types.APIKeyData, payload []byte) (err error)

	// LoadRaw loads matching scenario
	LoadRaw(method types.MethodType, name string, path string) (b []byte, err error)

	// Save APIScenario
	Save(scenario *types.APIScenario) (err error)

	// Delete removes a api scenario
	Delete(
		method types.MethodType,
		scenarioName string,
		path string) error

	// LookupAll finds matching scenarios
	LookupAll(key *types.APIKeyData) ([]*types.APIKeyData, int)

	// LookupAllByGroup finds matching scenarios by group
	LookupAllByGroup(group string) []*types.APIKeyData

	// LookupAllByPath finds matching scenarios by path
	LookupAllByPath(path string) []*types.APIKeyData

	// Lookup finds top matching scenario that hasn't been used recently
	Lookup(target *types.APIKeyData, data map[string]any) (*types.APIScenario, error)

	// ListScenarioKeyData returns keys for all scenarios
	ListScenarioKeyData(group string) []*types.APIKeyData
}
