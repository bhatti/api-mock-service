package fuzz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ExtractJSONPath_DotNotation(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{"id": "u123"},
	}
	require.Equal(t, "u123", ExtractJSONPath("user.id", data))
}

func Test_ExtractJSONPath_ArrayIndex(t *testing.T) {
	data := map[string]any{
		"items": []any{
			map[string]any{"name": "first"},
			map[string]any{"name": "second"},
		},
	}
	require.Equal(t, "first", ExtractJSONPath("items[0].name", data))
	require.Equal(t, "second", ExtractJSONPath("items[1].name", data))
}

func Test_ExtractJSONPath_RootAnchor(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{"email": "a@b.com"},
	}
	// "$.user.email" strips the "$." prefix
	require.Equal(t, "a@b.com", ExtractJSONPath("$.user.email", data))
}

func Test_ExtractJSONPath_MissingReturnsNil(t *testing.T) {
	data := map[string]any{"x": "val"}
	require.Nil(t, ExtractJSONPath("y.z", data))
}

func Test_IsJSONPathExpression_DollarDot(t *testing.T) {
	require.True(t, IsJSONPathExpression("$.user.id"))
	require.True(t, IsJSONPathExpression("$.items[0]"))
}

func Test_IsJSONPathExpression_ArrayBracket(t *testing.T) {
	require.True(t, IsJSONPathExpression("items[0].name"))
}

func Test_IsJSONPathExpression_FlatKey(t *testing.T) {
	require.False(t, IsJSONPathExpression("userId"))
	require.False(t, IsJSONPathExpression("user.id")) // dot notation without "$." is flat key, not jsonpath
}

func Test_ValidateRegexMap_JSONPathKey(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{"email": "alice@example.com"},
	}
	regex := map[string]string{
		"$.user.email": `__string__\w+@\w+\.\w+`,
	}
	err := ValidateRegexMap(data, regex)
	require.NoError(t, err)
}

func Test_ValidateRegexMap_JSONPathKey_Mismatch(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{"email": "not-an-email"},
	}
	regex := map[string]string{
		"$.user.email": `__string__\w+@\w+\.\w+`,
	}
	err := ValidateRegexMap(data, regex)
	require.Error(t, err)
}

func Test_ValidateRegexMap_MixedJSONPathAndFlat(t *testing.T) {
	data := map[string]any{
		"status": "active",
		"user":   map[string]any{"id": "u42"},
	}
	regex := map[string]string{
		"status":    "__string__active",
		"$.user.id": "__string__u42",
	}
	err := ValidateRegexMap(data, regex)
	require.NoError(t, err)
}

func Test_ValidateRegexMap_ArrayIndex(t *testing.T) {
	data := map[string]any{
		"items": []any{"apple", "banana"},
	}
	regex := map[string]string{
		"items[0]": "__string__apple",
	}
	err := ValidateRegexMap(data, regex)
	require.NoError(t, err)
}
