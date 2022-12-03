package oapi

import (
	"context"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func Test_ShouldParseValidTwitterOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/twitter.yaml")
	require.NoError(t, err)
	specs, err := Parse(context.Background(), data)
	require.NoError(t, err)
	require.Equal(t, 112, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario()
		require.NoError(t, err)
		out, err := yaml.Marshal(scenario)
		require.NoError(t, err)
		require.True(t, len(out) > 0)
	}
}

func Test_ShouldParseValidJobsOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	specs, err := Parse(context.Background(), data)
	require.NoError(t, err)
	require.Equal(t, 27, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario()
		require.NoError(t, err)
		require.Equal(t, ".+", scenario.Request.MatchHeaders["x-api-key"])
		_, err = yaml.Marshal(scenario)
		require.NoError(t, err)
	}
}

func Test_ShouldParseValidTwilioOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/twilio_accounts_v1.yaml")
	require.NoError(t, err)
	specs, err := Parse(context.Background(), data)
	require.NoError(t, err)
	require.Equal(t, 10, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario()
		require.NoError(t, err)
		_, err = yaml.Marshal(scenario)
		require.NoError(t, err)
	}
}

func Test_ShouldParseValidPetsOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/pets.yaml")
	//data, err := os.ReadFile("../../fixtures/oapi/plaid.yaml")
	require.NoError(t, err)
	specs, err := Parse(context.Background(), data)
	require.NoError(t, err)
	require.Equal(t, 10, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario()
		require.NoError(t, err)
		_, err = yaml.Marshal(scenario)
		require.NoError(t, err)
	}
}

func Test_ShouldParseValidLambdaOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/lambda.yaml")
	require.NoError(t, err)
	specs, err := Parse(context.Background(), data)
	require.NoError(t, err)
	require.Equal(t, 22, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario()
		require.NoError(t, err)
		_, err = yaml.Marshal(scenario)
		require.NoError(t, err)
	}
}

func Test_ShouldParseValidVimeoOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/vimeo.yaml")
	require.NoError(t, err)
	specs, err := Parse(context.Background(), data)
	require.NoError(t, err)
	require.Equal(t, 620, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario()
		require.NoError(t, err)
		_, err = yaml.Marshal(scenario)
		require.NoError(t, err)
	}
}
