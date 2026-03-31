package oapi

import (
	"context"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_ShouldHandleAdditionalStringFormats(t *testing.T) {
	formats := []string{"password", "byte", "binary", "ipv4", "ipv6", "hostname", "int32", "int64", "float", "double"}
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	for _, format := range formats {
		prop := Property{Name: "field", Type: "string", Format: format}
		val := prop.Value(dataTempl)
		require.NotNil(t, val, "format %q should produce non-nil value", format)
		m, ok := val.(map[string]string)
		require.True(t, ok, "format %q should return map[string]string, got %T", format, val)
		require.NotEmpty(t, m["field"], "format %q map value should not be empty", format)
	}
}

func Test_ShouldReturnMapForAllFormatsWithIncludeType(t *testing.T) {
	formats := []string{
		"date", "date-time", "uri", "email", "uuid", "phone", "host", "hostname",
		"ulid", "airport", "locale", "country", "zip", "ip", "ipv4", "ipv6",
		"isbn10", "isbn13", "ssn", "password", "byte", "binary", "int32", "int64", "float", "double",
	}
	dataTempl := fuzz.NewDataTemplateRequest(true, 1, 1)
	for _, format := range formats {
		prop := Property{Name: "field", Type: "string", Format: format}
		val := prop.Value(dataTempl)
		require.NotNil(t, val, "IncludeType=true format %q should produce non-nil value", format)
		_, ok := val.(map[string]string)
		require.True(t, ok, "IncludeType=true format %q should return map[string]string, got %T", format, val)
	}
}

func Test_ShouldBuildProperty(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/twilio_accounts_v1.yaml")
	require.NoError(t, err)
	loader := &openapi3.Loader{Context: context.Background(), IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromData(data)
	require.NoError(t, err)
	reqParam := doc.Paths["/v1/Credentials/AWS"].Get.Parameters[0]
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	reqProperty := schemaToProperty(reqParam.Value.Name, true, reqParam.Value.In, reqParam.Value.Schema.Value, dataTempl)
	require.Equal(t, float64(1), reqProperty.Min)
	require.Equal(t, float64(1000), reqProperty.Max)
	require.Equal(t, "integer", reqProperty.Type)
	require.Contains(t, reqProperty.String(), reqProperty.Name)

	for k, resParam := range doc.Paths["/v1/Credentials/AWS"].Get.Responses["200"].Value.Content.Get("application/json").Schema.Value.Properties {
		resProperty := schemaToProperty(k, false, "body", resParam.Value, dataTempl)
		require.Contains(t, resProperty.String(), k)
		require.NotNil(t, resProperty.Value(dataTempl))
	}

}
