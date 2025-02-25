package pm

import (
	"encoding/json"
	"io"
	"os"

	"github.com/bhatti/api-mock-service/internal/types"
	log "github.com/sirupsen/logrus"
)

// PostmanEnvironment represents the structure of a Postman environment file
type PostmanEnvironment struct {
	ID     string               `json:"id"`
	Name   string               `json:"name"`
	Values []PostmanEnvVariable `json:"values"`
}

// PostmanEnvVariable represents a variable in a Postman environment
type PostmanEnvVariable struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

// ParseEnvironment parses a Postman environment file into a PostmanEnvironment struct
func ParseEnvironment(reader io.Reader) (*PostmanEnvironment, error) {
	decoder := json.NewDecoder(reader)
	var environment PostmanEnvironment
	if err := decoder.Decode(&environment); err != nil {
		return nil, err
	}
	return &environment, nil
}

// ConvertEnvironmentToAPIVariables converts a PostmanEnvironment to APIVariables
func ConvertEnvironmentToAPIVariables(environment *PostmanEnvironment) *types.APIVariables {
	// Create a new APIVariables instance
	apiVars := &types.APIVariables{
		Name:      environment.Name,
		Variables: make(map[string]string),
	}

	// Convert each Postman environment variable to APIVariables
	for _, envVar := range environment.Values {
		// Only include enabled variables
		if envVar.Enabled {
			apiVars.Variables[envVar.Key] = envVar.Value
		}
	}

	return apiVars
}

// LoadAndConvertPostmanEnvironment loads a Postman environment file and converts it to APIVariables
func LoadAndConvertPostmanEnvironment(filePath string) (*types.APIVariables, error) {
	// Open the environment file
	file, err := os.Open(filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"FilePath": filePath,
			"Error":    err,
		}).Error("Failed to open Postman environment file")
		return nil, err
	}
	defer file.Close()

	// Parse the environment file
	environment, err := ParseEnvironment(file)
	if err != nil {
		log.WithFields(log.Fields{
			"FilePath": filePath,
			"Error":    err,
		}).Error("Failed to parse Postman environment file")
		return nil, err
	}

	// Convert to APIVariables
	return ConvertEnvironmentToAPIVariables(environment), nil
}
