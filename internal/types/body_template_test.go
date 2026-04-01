package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_BodyFields_InjectedAsTemplateParams(t *testing.T) {
	params := map[string]any{}
	body := []byte(`{"userId":"u1","score":42}`)
	InjectBodyFieldsAsTemplateParams(params, body)
	require.Equal(t, "u1", params["userId"])
	require.Equal(t, float64(42), params["score"])
}

func Test_BodyFields_PathParamsWin(t *testing.T) {
	params := map[string]any{
		"userId": "path-value",
	}
	body := []byte(`{"userId":"body-value","other":"x"}`)
	InjectBodyFieldsAsTemplateParams(params, body)
	// path param wins
	require.Equal(t, "path-value", params["userId"])
	// non-conflicting body field is injected
	require.Equal(t, "x", params["other"])
}

func Test_BodyFields_NonJSONBodyIsIgnored(t *testing.T) {
	params := map[string]any{}
	body := []byte(`plain text body`)
	// Should not panic or error
	InjectBodyFieldsAsTemplateParams(params, body)
	require.Empty(t, params)
}

func Test_BodyFields_EmptyBodyIsIgnored(t *testing.T) {
	params := map[string]any{}
	InjectBodyFieldsAsTemplateParams(params, nil)
	require.Empty(t, params)
	InjectBodyFieldsAsTemplateParams(params, []byte{})
	require.Empty(t, params)
}

func Test_BodyFields_NestedNotFlattened(t *testing.T) {
	params := map[string]any{}
	body := []byte(`{"user":{"id":"x","name":"alice"}}`)
	InjectBodyFieldsAsTemplateParams(params, body)
	// "user" is injected as a map, not flattened
	require.Contains(t, params, "user")
	require.NotContains(t, params, "id")
	require.NotContains(t, params, "name")
	userMap, ok := params["user"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "x", userMap["id"])
}
