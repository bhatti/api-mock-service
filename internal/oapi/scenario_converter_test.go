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

func Test_ShouldConvertAnyToSchema(t *testing.T) {
	b := `
{
  "account": "21212423423",
  "boo": [
    true,
    false
  ],
  "id": "us-west2_test1",
  "items": [
    1.1,
    2
  ],
  "locations": [
    {
      "lat": 12.5,
      "lng": 12
    }
  ],
  "logs": [
    {
      "config": {
        "timeout": 5
      },
      "created_at": 123,
      "id": 1,
      "name": "jake"
    },
    {
      "config": {
        "timeout": 6
      },
      "created_at": 234,
      "id": 2,
      "name": "larry"
    }
  ],
  "name": "sample-id5",
  "regions": [
    "us-east-2",
    "us-west-2"
  ],
  "taxes": [
    123,
    14
  ]
}`
	obj, err := fuzz.UnmarshalArrayOrObject([]byte(b))
	require.NoError(t, err)
	schema := anyToSchema(obj)
	require.NotNil(t, schema)
}

func Test_ShouldParseJobsOpenAPI(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN mock scenarios from open-api specifications
	b, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	// AND scenario repository
	repo, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)

	// AND valid template for random data
	dataTemplate := fuzz.NewDataTemplateRequest(false, 1, 2)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, b, dataTemplate)

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
	scenarios := make([]*types.APIScenario, 0)
	methods := []types.MethodType{types.Post, types.Get, types.Put, types.Delete,
		types.Options, types.Head, types.Patch, types.Connect, types.Trace}
	for i, key := range repo.LookupAllByGroup("SpecConvert") {
		scenario, err := repo.Lookup(key, nil)
		require.NoError(t, err)
		if i > 0 {
			scenario.BaseURL = "https://localhost:8080"
		}
		scenario.Method = methods[i%len(methods)]
		scenarios = append(scenarios, scenario)
	}
	_, err = MarshalScenarioToOpenAPI("t-api", "t-version", scenarios...)
	require.NoError(t, err)
}
