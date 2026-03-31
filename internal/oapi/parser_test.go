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

func Test_ShouldPlaceQueryAPIKeyInQueryParams(t *testing.T) {
	// Construct minimal OpenAPI spec with apiKey security scheme in: query
	oapiJSON := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "TestAPI", "version": "1.0"},
		"paths": {
			"/items": {
				"get": {
					"operationId": "listItems",
					"responses": {"200": {"description": "ok"}}
				}
			}
		},
		"components": {
			"securitySchemes": {
				"apiKeyQuery": {
					"type": "apiKey",
					"name": "api_key",
					"in": "query"
				}
			}
		}
	}`)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, oapiJSON, dataTempl)
	require.NoError(t, err)
	require.Equal(t, 1, len(specs))

	spec := specs[0]
	// api_key should be in QueryParams, not in Headers
	found := false
	for _, qp := range spec.Request.QueryParams {
		if qp.Name == "api_key" {
			found = true
		}
	}
	require.True(t, found, "api_key security scheme with in:query should be in QueryParams")
	// Headers should not contain api_key
	for _, h := range spec.Request.Headers {
		require.NotEqual(t, "api_key", h.Name, "api_key with in:query must not appear in Headers")
	}
}

func Test_ShouldNotDuplicateAllOfChildren(t *testing.T) {
	// Construct schema with allOf containing 2 sub-schemas
	oapiJSON := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "AllOfTest", "version": "1.0"},
		"paths": {
			"/items": {
				"get": {
					"operationId": "getItem",
					"responses": {
						"200": {
							"description": "ok",
							"content": {
								"application/json": {
									"schema": {
										"allOf": [
											{"properties": {"id": {"type": "string"}}},
											{"properties": {"name": {"type": "string"}}}
										]
									}
								}
							}
						}
					}
				}
			}
		}
	}`)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, oapiJSON, dataTempl)
	require.NoError(t, err)
	require.Equal(t, 1, len(specs))

	body := specs[0].Response.Body
	require.Len(t, body, 1)

	// Count children — should be exactly 2 (id + name), not 6 (triplicated bug)
	seen := make(map[string]int)
	for _, child := range body[0].Children {
		seen[child.Name]++
	}
	for name, count := range seen {
		require.Equal(t, 1, count, "property %q should appear exactly once, got %d (duplicate allOf bug)", name, count)
	}
}

func Test_ShouldParseOneOfAndAnyOfSchemas(t *testing.T) {
	oapiJSON := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "OneOfAnyOf", "version": "1.0"},
		"paths": {
			"/items": {
				"get": {
					"operationId": "getItem",
					"responses": {
						"200": {
							"description": "ok",
							"content": {
								"application/json": {
									"schema": {
										"oneOf": [
											{"properties": {"cat_name": {"type": "string"}}},
											{"properties": {"dog_name": {"type": "string"}}}
										]
									}
								}
							}
						}
					}
				}
			}
		}
	}`)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, oapiJSON, dataTempl)
	require.NoError(t, err)
	require.Equal(t, 1, len(specs))

	body := specs[0].Response.Body
	require.Len(t, body, 1)
	// First oneOf branch (cat_name) should be captured; second branch ignored
	require.NotEmpty(t, body[0].Children, "oneOf properties should be captured")
	require.Equal(t, "cat_name", body[0].Children[0].Name)
}

func Test_ShouldPopulateTagsFromOpenAPIOperation(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/pets.yaml")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)
	require.NoError(t, err)

	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		require.NotEmpty(t, scenario.Tags, "every scenario should have at least one tag")
	}
}
