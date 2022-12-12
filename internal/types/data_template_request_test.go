package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldCreateDataTemplateRequest(t *testing.T) {
	req := NewDataTemplateRequest(false, 1, 1)
	require.Equal(t, 1, req.MinMultiplier)
	req = req.WithMaxMultiplier(2)
	require.Equal(t, 2, req.MaxMultiplier)
	req = req.WithInclude(true)
	require.True(t, req.IncludeType)
}
