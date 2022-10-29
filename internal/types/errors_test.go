package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldBuildValidationError(t *testing.T) {
	// GIVEN a mismatch error
	err := NewValidationError("test error")
	// THEN it should match message
	require.Error(t, err)
	require.Equal(t, "test error", err.Error())
}

func Test_ShouldBuildNotFoundError(t *testing.T) {
	// GIVEN a mismatch error
	err := NewNotFoundError("test error")
	// THEN it should match message
	require.Error(t, err)
	require.Equal(t, "test error", err.Error())
}
