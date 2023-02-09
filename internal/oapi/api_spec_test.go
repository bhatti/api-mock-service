package oapi

import (
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func Test_ShouldStripQuotes(t *testing.T) {
	contents := `{"job":{"add":"{{RandStringArrayMinMax 1 1}}","attributeMap":"{{RandDict}}","completed":"{{RandBool}}","jobId":"{{RandStringMinMax 0 0}}","jobStatus":"{{EnumString PENDING RUNNING SUCCEEDED CANCELED FAILED}}","name":   "{{RandStringMinMax 0 0}}","records":"{{RandNumMinMax 0 0}}","remaining":"{{RandNumMinMax 0 0}}","remove":"{{RandStringArrayMinMax 1 1}}"}}`
	out := stripNumericBooleanQuotes([]byte(contents))
	expected := `{"job":{"add":{{RandStringArrayMinMax 1 1}},"attributeMap":{{RandDict}},"completed":{{RandBool}},"jobId":"{{RandStringMinMax 0 0}}","jobStatus":"{{EnumString PENDING RUNNING SUCCEEDED CANCELED FAILED}}","name":   "{{RandStringMinMax 0 0}}","records":{{RandNumMinMax 0 0}},"remaining":{{RandNumMinMax 0 0}},"remove":{{RandStringArrayMinMax 1 1}}}}`
	require.Equal(t, expected, string(out))
}

func Test_ShouldParseAndGenerateSaveCustomerTemplate(t *testing.T) {
	// GIVEN a mock scenario loaded from YAML
	b, err := os.ReadFile("../../fixtures/save_customer.yaml")
	require.NoError(t, err)

	// WHEN parsing YAML for contents tag
	body, err := fuzz.ParseTemplate("../../fixtures", b, map[string]any{})

	// THEN it should not fail
	require.NoError(t, err)
	scenario := types.MockScenario{}
	// AND it should return valid mock scenario
	err = yaml.Unmarshal(body, &scenario)
	require.NoError(t, err)
	// AND it should have expected contents

	require.Contains(t, scenario.Response.ContentType(""), "application/json")

	obj, err := fuzz.UnmarshalArrayOrObject([]byte(scenario.Request.Contents))
	require.NoError(t, err)
	obj = fuzz.GenerateFuzzData(obj)
	require.NotNil(t, obj)
}
