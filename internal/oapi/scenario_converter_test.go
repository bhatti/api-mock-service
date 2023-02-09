package oapi

import (
	"context"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_ShouldParseJobsOpenAPI(t *testing.T) {
	// GIVEN mock scenarios from open-api specifications
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	// AND scenario repository
	repo, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	specs, err := Parse(context.Background(), b, dataTemplate)
	require.NoError(t, err)

	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTemplate)
		scenario.Group = "SpecConvert"
		require.NoError(t, err)
		// WHEN saving scenario to mock scenario repository
		err = repo.Save(scenario)
		// THEN it should save scenario
		require.NoError(t, err)
		// AND should return saved scenario
		_, err = repo.Lookup(scenario.ToKeyData(), nil)
		require.NoError(t, err)
	}
	scenarios := make([]*types.MockScenario, 0)
	for _, key := range repo.LookupAllByGroup("SpecConvert") {
		scenario, err := repo.Lookup(key, nil)
		require.NoError(t, err)
		scenarios = append(scenarios, scenario)
	}
	out, err := MarshalScenarioToOpenAPI("t-api", "t-version", scenarios...)
	require.NoError(t, err)
	_ = os.WriteFile("../../output.json", out, 0644)
}
