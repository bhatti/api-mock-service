package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldValidateConfiguration(t *testing.T) {
	// GIVEN a configuration
	config, err := NewConfiguration(8080, 8081, "/dir", "/asset", &Version{})
	// WHEN validating config
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "/dir", config.DataDir)
}

func Test_ShouldNotValidateConfiguration(t *testing.T) {
	// GIVEN a configuration
	_, err := NewConfiguration(8080, 8080, "/dir", "/asset", &Version{})
	// WHEN validating config
	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldMatchHeader(t *testing.T) {
	// GIVEN a configuration
	c, err := NewConfiguration(8080, 8081, "/dir", "/asset", &Version{})
	require.NoError(t, err)
	// WHEN matching header with empty regex
	// THEN it should not match
	require.False(t, c.MatchHeader("test"))
	c.MatchHeaderRegex = "test\\d"
	// WHEN matching header with empty input
	// THEN it should not match
	require.False(t, c.MatchHeader(""))
	// WHEN matching header with non matching input
	// THEN it should not match
	require.False(t, c.MatchHeader("test"))
	// WHEN matching header with matching input
	// THEN it should not match
	require.True(t, c.MatchHeader("test1"))
}
