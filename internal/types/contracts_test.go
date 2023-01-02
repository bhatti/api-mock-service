package types

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldCreateContractRequest(t *testing.T) {
	require.Equal(t, 1, NewContractRequest("", 1).ExecutionTimes)
}

func Test_ShouldCreateContractResponse(t *testing.T) {
	require.Equal(t, 0, NewContractResponse().Failed)
}

func Test_ShouldAddContractResponse(t *testing.T) {
	res := NewContractResponse()
	res.Add("test", nil, nil)
	require.Equal(t, 1, res.Succeeded)
	res.Add("test", 1, nil)
	require.Equal(t, 2, res.Succeeded)
	res.Add("test", 1, errors.New("test"))
	require.Equal(t, 2, res.Succeeded)
	require.Equal(t, 1, res.Failed)
}
