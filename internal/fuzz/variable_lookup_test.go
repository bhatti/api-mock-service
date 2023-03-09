package fuzz

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldFindVariable(t *testing.T) {
	require.Nil(t, FindVariable("k", nil))
	require.Nil(t, FindVariable("k", ""))
	require.NotNil(t, FindVariable("k", map[string]string{"k": "1"}))
	require.Nil(t, FindVariable("k", map[string]string{"x": "1"}))
	require.NotNil(t, FindVariable("k.a", map[string]any{"k": map[string]string{"a": "b"}}))
}

func Test_ShouldFindVariableArray(t *testing.T) {
	res := FindVariable("k.a", []any{
		map[string]any{"k": map[string]string{"a": "b"}},
		map[string]any{"k": map[string]string{"a": "c"}},
	}).([]any)
	require.Equal(t, 2, len(res))
	for _, next := range res {
		m := next.(map[string]string)
		require.True(t, m["a"] != "")
	}
}

func Test_ShouldCompareVariable(t *testing.T) {
	require.False(t, VariableEquals("k", nil, 1))
	require.False(t, VariableEquals("k", "", 1))
	require.True(t, VariableEquals("k", map[string]string{"k": "1"}, "1"))
	require.False(t, VariableEquals("k", map[string]string{"x": "1"}, "1"))
	require.True(t, VariableEquals("k.a", map[string]any{"k": map[string]string{"a": "b"}}, "b"))
}

func Test_ShouldCompareVariableNumber(t *testing.T) {
	require.Equal(t, float64(0), VariableNumber("k", nil))
	require.Equal(t, float64(2), VariableNumber("k", map[string]string{"k": "2"}))
	require.Equal(t, float64(0), VariableNumber("k", map[string]string{"x": "2"}))
	require.True(t, VariableNumber("k.a", map[string]any{"k": map[string]string{"a": "3.1"}}) > 3.0)
}

func Test_ShouldContainsVariable(t *testing.T) {
	require.False(t, VariableContains("k", nil, 1))
	require.False(t, VariableContains("k", []any{""}, 1))
	require.True(t, VariableContains("k", []any{"1"}, map[string]string{"k": "1"}))
	require.False(t, VariableContains("k", []any{"1"}, map[string]string{"x": "1"}))
	require.True(t, VariableContains("k.a", []any{"b"}, map[string]any{"k": map[string]string{"a": "abc"}}))
	require.False(t, VariableContains("k.a", []any{"abcd"}, map[string]any{"k": map[string]string{"a": "abc"}}))
}

func Test_ShouldFindVariableInMap(t *testing.T) {
	target := map[string]any{
		"attributes": []any{
			map[string]any{"value": "e8b84c3c", "enabled": true, "date": 1.671221246959e+09, "username": "doris@indecens.org"},
			map[string]any{"value": "fa135431", "enabled": true, "date": 1.671221246959e+09, "username": "minora@indecens.org"},
		},
	}
	res := FindVariable("attributes.username", target).([]any)
	require.Equal(t, 2, len(res))
}

func Test_ShouldFindVariableInMapArray(t *testing.T) {
	target := map[string]any{
		"attributes": []any{
			map[string][]any{"username": {"doris@indecens.org"}},
			map[string][]any{"username": {"minora@indecens.org"}},
		},
	}
	res := FindVariable("attributes.username", target).([]any)
	require.Equal(t, 2, len(res))
}
