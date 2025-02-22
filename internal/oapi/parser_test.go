package oapi

import (
	"context"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func Test_ShouldParseValidTwitterOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/twitter.yaml")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)
	require.NoError(t, err)
	require.Equal(t, 112, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		out, err := yaml.Marshal(scenario)
		require.NoError(t, err)
		require.True(t, len(out) > 0)
	}
}

func Test_ShouldParseAndConvertValidDescribeAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/describe-job.json")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)

	require.NoError(t, err)
	require.Len(t, specs, 6)
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		require.True(t, scenario.Request.Headers["x-api-key"] != "")
		_, err = yaml.Marshal(scenario)
		require.NoError(t, err)
	}
}

func Test_ShouldParseValidJobsOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/jobs-openapi.json")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)

	require.NoError(t, err)
	require.Equal(t, 32, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		require.True(t, scenario.Request.Headers["x-api-key"] != "")
		_, err = yaml.Marshal(scenario)
		require.NoError(t, err)
	}
}

func Test_ShouldParseValidTwilioOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/twilio_accounts_v1.yaml")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)

	require.NoError(t, err)
	require.Len(t, specs, 13)
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		_, err = yaml.Marshal(scenario)
		require.NoError(t, err)
	}
}

func Test_ShouldParseValidPetsOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/pets.yaml")
	//data, err := os.ReadFile("../../fixtures/oapi/plaid.yaml")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)

	require.NoError(t, err)
	require.Equal(t, 10, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		_, err = yaml.Marshal(scenario)
		require.NoError(t, err)
	}
}

func Test_ShouldParseValidLambdaOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/lambda.yaml")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)

	require.NoError(t, err)
	require.Len(t, specs, 30)
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		_, err = yaml.Marshal(scenario)
		require.NoError(t, err)
	}
}

func Test_ShouldParseProductsOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/post-product.json")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)

	require.NoError(t, err)
	require.Equal(t, 4, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		out, err := yaml.Marshal(scenario)
		require.NoError(t, err)
		require.True(t, len(out) > 0)
	}
}

func Test_ShouldParseGetCustomerOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/get-customer.json")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)

	require.NoError(t, err)
	require.Equal(t, 1, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		out, err := yaml.Marshal(scenario)
		require.NoError(t, err)
		require.True(t, len(out) > 0)
	}
}

func Test_ShouldParseGetCustomersOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/get-customers.json")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)

	require.NoError(t, err)
	require.Equal(t, 1, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		out, err := yaml.Marshal(scenario)
		require.NoError(t, err)
		require.True(t, len(out) > 0)
	}
}

func Test_ShouldParseSaveCustomersOpenAPI(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/save-customer.json")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)

	require.NoError(t, err)
	require.Equal(t, 1, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		out, err := yaml.Marshal(scenario)
		require.NoError(t, err)
		require.True(t, len(out) > 0)
	}
}
