package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldValidateConfiguration(t *testing.T) {
	// GIVEN an configuration
	config, err := NewConfiguration(8080, "/dir", "/asset", &Version{})
	// WHEN validating config
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "/dir", config.DataDir)
}
