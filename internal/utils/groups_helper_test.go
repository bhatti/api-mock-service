package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ShouldNormalizeGroup(t *testing.T) {
	require.Equal(t, "root", NormalizeGroup("", ""))
	require.Equal(t, "title", NormalizeGroup("title", ""))
	require.Equal(t, "path1", NormalizeGroup("", "/path1/{test}"))
	require.Equal(t, "path1", NormalizeGroup("", "/path1/:test"))
	require.Equal(t, "path1_path2", NormalizeGroup("", "/path1/path2"))
	require.Equal(t, "root", NormalizeGroup("", "/"))
}
