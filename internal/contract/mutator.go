package contract

import (
	"encoding/json"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"math"
	"strings"
)

// ContractMutator creates variations of a contract to test robustness
type ContractMutator struct {
	scenario  *types.APIScenario
	mutations []*types.APIScenario
}

// NewContractMutator creates a new mutator
func NewContractMutator(scenario *types.APIScenario) *ContractMutator {
	return &ContractMutator{
		scenario:  scenario,
		mutations: make([]*types.APIScenario, 0),
	}
}

// GenerateMutations Generate mutations to test boundary conditions and edge cases
func (m *ContractMutator) GenerateMutations() []*types.APIScenario {
	// Create a copy with missing optional fields
	m.createMissingFieldsMutation()

	// Create a copy with boundary values
	m.createBoundaryValuesMutation()

	// Create a copy with malformed data
	m.createMalformedDataMutation()

	return m.mutations
}

// Implementation example: create a mutation with boundary values
func (m *ContractMutator) createBoundaryValuesMutation() {
	scenario := *m.scenario
	scenario.Name = scenario.Name + "-boundary-values"

	// For each numeric field in request body, create versions with min/max values
	if m.scenario.Request.Contents != "" {
		var requestBody map[string]interface{}
		if err := json.Unmarshal([]byte(m.scenario.Request.Contents), &requestBody); err == nil {
			applyBoundaryValues(requestBody)
			if newBody, err := json.Marshal(requestBody); err == nil {
				scenario.Request.Contents = string(newBody)
				m.mutations = append(m.mutations, &scenario)
			}
		}
	}
}

// createMissingFieldsMutation creates a scenario with optional fields removed
func (m *ContractMutator) createMissingFieldsMutation() {
	// Deep copy the scenario
	scenarioCopy := *m.scenario
	scenarioCopy.Name = scenarioCopy.Name + "-missing-fields"

	// Parse request body to identify optional fields
	if m.scenario.Request.Contents != "" {
		var requestBody map[string]interface{}
		if err := json.Unmarshal([]byte(m.scenario.Request.Contents), &requestBody); err == nil {
			// Remove non-required fields based on assertion patterns
			for field := range requestBody {
				// Check if the field is required in assertion patterns
				isRequired := false
				if m.scenario.Request.AssertContentsPattern != "" {
					var assertPattern map[string]interface{}
					if err := json.Unmarshal([]byte(m.scenario.Request.AssertContentsPattern), &assertPattern); err == nil {
						_, isRequired = assertPattern[field]
					}
				}

				if !isRequired {
					delete(requestBody, field)
				}
			}

			if newBody, err := json.Marshal(requestBody); err == nil {
				scenarioCopy.Request.Contents = string(newBody)
				m.mutations = append(m.mutations, &scenarioCopy)
			}
		}
	}
}

// createMalformedDataMutation creates scenarios with malformed data
func (m *ContractMutator) createMalformedDataMutation() {
	// Create several malformed variations
	malformations := []struct {
		name    string
		mutator func(map[string]interface{}) map[string]interface{}
	}{
		{
			name: "overflow-strings",
			mutator: func(data map[string]interface{}) map[string]interface{} {
				result := make(map[string]interface{})
				for k, v := range data {
					switch val := v.(type) {
					case string:
						// Create very long string
						result[k] = strings.Repeat("X", 10000)
					case map[string]interface{}:
						result[k] = overflowStrings(val)
					default:
						result[k] = v
					}
				}
				return result
			},
		},
		{
			name: "invalid-types",
			mutator: func(data map[string]interface{}) map[string]interface{} {
				result := make(map[string]interface{})
				for k, v := range data {
					switch v.(type) {
					case string:
						result[k] = 12345 // Replace string with number
					case float64, int, int64:
						result[k] = "not-a-number" // Replace number with string
					case bool:
						result[k] = "not-a-boolean" // Replace boolean with string
					case map[string]interface{}:
						result[k] = "not-an-object" // Replace object with string
					default:
						result[k] = v
					}
				}
				return result
			},
		},
		{
			name: "special-chars",
			mutator: func(data map[string]interface{}) map[string]interface{} {
				result := make(map[string]interface{})
				for k, v := range data {
					switch val := v.(type) {
					case string:
						// Add special characters
						result[k] = val + "'\"><script>alert(1)</script>"
					case map[string]interface{}:
						result[k] = addSpecialChars(val)
					default:
						result[k] = v
					}
				}
				return result
			},
		},
	}

	// Apply each malformation
	for _, malformation := range malformations {
		if m.scenario.Request.Contents != "" {
			var requestBody map[string]interface{}
			if err := json.Unmarshal([]byte(m.scenario.Request.Contents), &requestBody); err == nil {
				// Apply the malformation
				mutatedBody := malformation.mutator(requestBody)

				// Create new scenario with the mutated body
				scenarioCopy := *m.scenario
				scenarioCopy.Name = scenarioCopy.Name + "-" + malformation.name

				if newBody, err := json.Marshal(mutatedBody); err == nil {
					scenarioCopy.Request.Contents = string(newBody)
					// For malformed data, expect an error response
					scenarioCopy.Response.StatusCode = 400
					m.mutations = append(m.mutations, &scenarioCopy)
				}
			}
		}
	}
}

// Helper functions for createMalformedDataMutation
func overflowStrings(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		switch val := v.(type) {
		case string:
			result[k] = strings.Repeat("X", 10000)
		case map[string]interface{}:
			result[k] = overflowStrings(val)
		default:
			result[k] = v
		}
	}
	return result
}

func addSpecialChars(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		switch val := v.(type) {
		case string:
			result[k] = val + "'\"><script>alert(1)</script>"
		case map[string]interface{}:
			result[k] = addSpecialChars(val)
		default:
			result[k] = v
		}
	}
	return result
}

// Helper function for boundary value mutation
func applyBoundaryValues(obj map[string]interface{}) {
	for key, value := range obj {
		switch v := value.(type) {
		case float64:
			// Apply a boundary condition - use maximum value
			obj[key] = math.MaxFloat64
		case map[string]interface{}:
			applyBoundaryValues(v)
		case []interface{}:
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					applyBoundaryValues(itemMap)
					v[i] = itemMap
				}
			}
		}
	}
}

// GenerateMutatedScenarios creates test variations of a scenario
func GenerateMutatedScenarios(scenario *types.APIScenario) ([]*types.APIScenario, error) {
	mutations := make([]*types.APIScenario, 0)

	// 1. Edge case testing - create a scenario with minimum values
	minScenario := *scenario
	minScenario.Name = scenario.Name + "-min-values"

	// Modify request body to use minimum values
	if reqBody, err := fuzz.UnmarshalArrayOrObject([]byte(scenario.Request.Contents)); err == nil {
		minBody := applyMinValues(reqBody)
		if bodyJSON, err := json.Marshal(minBody); err == nil {
			minScenario.Request.Contents = string(bodyJSON)
		}
	}
	mutations = append(mutations, &minScenario)

	// 2. Edge case testing - create a scenario with maximum values
	maxScenario := *scenario
	maxScenario.Name = scenario.Name + "-max-values"

	// Modify request body to use maximum values
	if reqBody, err := fuzz.UnmarshalArrayOrObject([]byte(scenario.Request.Contents)); err == nil {
		maxBody := applyMaxValues(reqBody)
		if bodyJSON, err := json.Marshal(maxBody); err == nil {
			maxScenario.Request.Contents = string(bodyJSON)
		}
	}
	mutations = append(mutations, &maxScenario)

	// 3. Error case testing - missing required fields
	if requiredFields := extractRequiredFields(scenario); len(requiredFields) > 0 {
		for _, field := range requiredFields {
			errorScenario := *scenario
			errorScenario.Name = scenario.Name + "-missing-" + field
			errorScenario.Response.StatusCode = 400 // Expect validation error

			// Modify request to remove the required field
			if reqBody, err := fuzz.UnmarshalArrayOrObject([]byte(scenario.Request.Contents)); err == nil {
				removePath(reqBody, field)
				if bodyJSON, err := json.Marshal(reqBody); err == nil {
					errorScenario.Request.Contents = string(bodyJSON)
				}
			}
			mutations = append(mutations, &errorScenario)
		}
	}

	return mutations, nil
}

// Helper functions for mutation testing
func applyMinValues(obj interface{}) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = applyMinValues(val)
		}
		return result
	case []interface{}:
		if len(v) > 0 {
			// For arrays, keep just one element with min values
			return []interface{}{applyMinValues(v[0])}
		}
		return v
	case float64:
		return 0.0
	case int:
		return 0
	case string:
		if len(v) > 0 {
			// Minimum valid string (often 1 character)
			return v[:1]
		}
		return ""
	default:
		return v
	}
}

func applyMaxValues(obj interface{}) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = applyMaxValues(val)
		}
		return result
	case []interface{}:
		if len(v) > 0 {
			// For arrays, create multiple elements with max values
			result := make([]interface{}, 10)
			for i := 0; i < 10; i++ {
				result[i] = applyMaxValues(v[0])
			}
			return result
		}
		return v
	case float64:
		return math.MaxFloat32
	case int:
		return math.MaxInt32
	case string:
		// Create a long string (but not too long to avoid breaking APIs)
		return strings.Repeat(v+"X", 50)
	default:
		return v
	}
}

func extractRequiredFields(scenario *types.APIScenario) []string {
	// Extract fields that appear to be required based on assertion patterns
	requiredFields := make([]string, 0)

	if scenario.Request.AssertContentsPattern != "" {
		var assertPattern map[string]interface{}
		if err := json.Unmarshal([]byte(scenario.Request.AssertContentsPattern), &assertPattern); err == nil {
			for field := range assertPattern {
				requiredFields = append(requiredFields, field)
			}
		}
	}

	return requiredFields
}

func removePath(obj interface{}, path string) {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return
	}

	if m, ok := obj.(map[string]interface{}); ok {
		if len(parts) == 1 {
			delete(m, parts[0])
			return
		}

		if nested, ok := m[parts[0]].(map[string]interface{}); ok {
			removePath(nested, strings.Join(parts[1:], "."))
		}
	}
}
