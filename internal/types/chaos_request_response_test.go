package types

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldCreateChaosRequest(t *testing.T) {
	require.Equal(t, 1, NewChaosRequest("", 1).ExecutionTimes)
}

func Test_ShouldCreateChaosResponse(t *testing.T) {
	require.Equal(t, 0, NewChaosResponse().Failed)
}

func Test_ShouldAddChaosResponse(t *testing.T) {
	res := NewChaosResponse()
	res.Add("test", nil, nil)
	require.Equal(t, 1, res.Succeeded)
	res.Add("test", 1, nil)
	require.Equal(t, 2, res.Succeeded)
	res.Add("test", 1, errors.New("test"))
	require.Equal(t, 2, res.Succeeded)
	require.Equal(t, 1, res.Failed)
}
