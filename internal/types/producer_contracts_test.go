package types

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldCreateContractRequest(t *testing.T) {
	require.Equal(t, 1, NewProducerContractRequest("", 1, 0).ExecutionTimes)
}

func Test_ShouldCreateContractResponse(t *testing.T) {
	require.Equal(t, 0, NewProducerContractResponse().Failed)
}

func Test_ShouldAddContractResponse(t *testing.T) {
	res := NewProducerContractResponse()
	res.Add("test", nil, nil)
	require.Equal(t, 1, res.Succeeded)
	res.Add("test", 1, nil)
	require.Equal(t, 2, res.Succeeded)
	res.Add("test", 1, errors.New("test"))
	require.Equal(t, 2, res.Succeeded)
	require.Equal(t, 1, res.Failed)
}

func Test_ShouldAddContractRequest(t *testing.T) {
	req := NewProducerContractRequest("url", 5, 0)
	req.Headers["h"] = []string{"val"}
	req.Params["p"] = "val"
	require.Equal(t, 2, len(req.Overrides()))
}
