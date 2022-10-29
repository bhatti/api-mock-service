package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldOutputVersion(t *testing.T) {
	// GIVEN a version
	version := NewVersion("1", "xx", "date")
	// The output should not be empty
	require.True(t, version.Output(true) != "")
	require.True(t, version.Output(false) != "")
	require.True(t, version.String() != "")
}
