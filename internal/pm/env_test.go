package pm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ShouldConvertPostmanEnvironment(t *testing.T) {
	// Create a temporary file with test Postman environment
	tempDir := t.TempDir()
	envFilePath := filepath.Join(tempDir, "test-env.json")

	// Create test environment content
	envJSON := `{
		"id": "test-environment",
		"name": "Test Environment",
		"values": [
			{
				"key": "base_url",
				"value": "https://api.example.com",
				"enabled": true
			},
			{
				"key": "api_key",
				"value": "your-api-key",
				"enabled": true
			},
			{
				"key": "access_token",
				"value": "",
				"enabled": true
			},
			{
				"key": "client_name",
				"value": "test-client",
				"enabled": true
			},
			{
				"key": "org_id",
				"value": "test-org",
				"enabled": true
			},
			{
				"key": "disabled_var",
				"value": "should-not-appear",
				"enabled": false
			}
		]
	}`

	// Write the environment JSON to the temp file
	err := os.WriteFile(envFilePath, []byte(envJSON), 0644)
	require.NoError(t, err)

	// Test parsing the environment file
	t.Run("Parse Environment File", func(t *testing.T) {
		file, err := os.Open(envFilePath)
		require.NoError(t, err)
		defer file.Close()

		env, err := ParseEnvironment(file)
		require.NoError(t, err)
		require.Equal(t, "test-environment", env.ID)
		require.Equal(t, "Test Environment", env.Name)
		require.Equal(t, 6, len(env.Values))
	})

	// Test converting to APIVariables
	t.Run("Convert to APIVariables", func(t *testing.T) {
		apiVars, err := LoadAndConvertPostmanEnvironment(envFilePath)
		require.NoError(t, err)

		// Check name
		require.Equal(t, "Test Environment", apiVars.Name)

		// Check variables
		require.Equal(t, 5, len(apiVars.Variables))
		require.Equal(t, "https://api.example.com", apiVars.Variables["base_url"])
		require.Equal(t, "your-api-key", apiVars.Variables["api_key"])
		require.Equal(t, "", apiVars.Variables["access_token"])
		require.Equal(t, "test-client", apiVars.Variables["client_name"])
		require.Equal(t, "test-org", apiVars.Variables["org_id"])

		// Verify disabled variables are not included
		_, hasDisabled := apiVars.Variables["disabled_var"]
		require.False(t, hasDisabled, "Disabled variables should not be included")
	})

	// Test with empty environment
	t.Run("Empty Environment", func(t *testing.T) {
		emptyEnvPath := filepath.Join(tempDir, "empty-env.json")
		emptyEnvJSON := `{
			"id": "empty-environment",
			"name": "Empty Environment",
			"values": []
		}`
		err := os.WriteFile(emptyEnvPath, []byte(emptyEnvJSON), 0644)
		require.NoError(t, err)

		apiVars, err := LoadAndConvertPostmanEnvironment(emptyEnvPath)
		require.NoError(t, err)
		require.Equal(t, "Empty Environment", apiVars.Name)
		require.Equal(t, 0, len(apiVars.Variables))
	})

	// Test with malformed JSON
	t.Run("Malformed JSON", func(t *testing.T) {
		malformedPath := filepath.Join(tempDir, "malformed.json")
		err := os.WriteFile(malformedPath, []byte(`{malformed json`), 0644)
		require.NoError(t, err)

		_, err = LoadAndConvertPostmanEnvironment(malformedPath)
		require.Error(t, err)
	})

	// Test with non-existent file
	t.Run("Non-existent File", func(t *testing.T) {
		_, err = LoadAndConvertPostmanEnvironment(filepath.Join(tempDir, "does-not-exist.json"))
		require.Error(t, err)
	})
}
