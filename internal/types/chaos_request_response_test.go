package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldCreateChaosRequest(t *testing.T) {
	require.Equal(t, 1, NewChaosRequest("", 1).ExecutionTimes)
}

func Test_ShouldCreateChaosResponse(t *testing.T) {
	require.Equal(t, 1, NewChaosResponse(nil, 1, 1).Failed)
}
