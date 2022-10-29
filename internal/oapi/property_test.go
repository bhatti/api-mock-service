package oapi

import (
	"context"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_ShouldBuildProperty(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/twilio_accounts_v1.yaml")
	require.NoError(t, err)
	loader := &openapi3.Loader{Context: context.Background(), IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromData(data)
	require.NoError(t, err)
	reqParam := doc.Paths["/v1/Credentials/AWS"].Get.Parameters[0]
	reqProperty := schemaToProperty(reqParam.Value.Name, true, reqParam.Value.In, reqParam.Value.Schema.Value)
	require.Equal(t, float64(1), reqProperty.Min)
	require.Equal(t, float64(1000), reqProperty.Max)
	require.Equal(t, "integer", reqProperty.Type)
	require.Contains(t, reqProperty.String(), reqProperty.Name)

	for k, resParam := range doc.Paths["/v1/Credentials/AWS"].Get.Responses["200"].Value.Content.Get("application/json").Schema.Value.Properties {
		resProperty := schemaToProperty(k, false, "body", resParam.Value)
		require.Contains(t, resProperty.String(), k)
		require.NotNil(t, resProperty.Value())
	}

}
