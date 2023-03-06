package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ShouldValidateConfiguration(t *testing.T) {
	// GIVEN a configuration
	config, err := NewConfiguration(8080, 8081, "/dir", "/asset", "/history", &Version{})
	// WHEN validating config
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "/dir", config.DataDir)
}

func Test_ShouldNotValidateConfiguration(t *testing.T) {
	// GIVEN a configuration
	_, err := NewConfiguration(8080, 8080, "/dir", "/asset", "/history", &Version{})
	// WHEN validating config
	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldMatchHeader(t *testing.T) {
	// GIVEN a configuration
	c, err := NewConfiguration(8080, 8081, "/dir", "/asset", "/history", &Version{})
	require.NoError(t, err)
	// WHEN matching header with empty regex
	// THEN it should not match
	require.False(t, c.AssertHeader("test"))
	c.AssertHeadersPattern = "test\\d"
	// WHEN matching header with empty input
	// THEN it should not match
	require.False(t, c.AssertHeader(""))
	// WHEN matching header with non matching input
	// THEN it should not match
	require.False(t, c.AssertHeader("test"))
	// WHEN matching header with matching input
	// THEN it should not match
	require.True(t, c.AssertHeader("test1"))
}

func Test_ShouldMatchQueryParameters(t *testing.T) {
	// GIVEN a configuration
	c, err := NewConfiguration(8080, 8081, "/dir", "/asset", "/history", &Version{})
	require.NoError(t, err)
	// WHEN matching query params with empty regex
	// THEN it should not match
	require.False(t, c.AssertQueryParams("test"))
	c.AssertQueryParamsPattern = "test\\d"
	// WHEN matching query params with empty input
	// THEN it should not match
	require.False(t, c.AssertQueryParams(""))
	// WHEN matching query params with non matching input
	// THEN it should not match
	require.False(t, c.AssertQueryParams("test"))
	// WHEN matching query params with matching input
	// THEN it should not match
	require.True(t, c.AssertQueryParams("test1"))
	require.True(t, c.AssertQueryParams("Test1"))
}
