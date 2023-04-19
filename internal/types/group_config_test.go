package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldNotCalculateGroupFailureProbabilityWithDisabled(t *testing.T) {
	// GIVEN a group config
	gc := &GroupConfig{}
	countFail := 0
	for i := 0; i < 100; i++ {
		status := gc.GetHTTPStatus()
		if status >= 300 {
			countFail++
		}
	}
	require.Equal(t, 0, countFail)
	countLatency := 0
	for i := 0; i < 100; i++ {
		t := gc.GetDelayLatency()
		if t > 0 {
			countLatency++
		}
	}
	require.Equal(t, 0, countLatency)
}

func Test_ShouldCalculateGroupFailureProbability(t *testing.T) {
	// GIVEN a group config
	gc := &GroupConfig{ChaosEnabled: true}
	countFail := 0
	for i := 0; i < 100; i++ {
		status := gc.GetHTTPStatus()
		if status >= 300 {
			countFail++
		}
	}
	require.True(t, countFail > 0)
	countLatency := 0
	for i := 0; i < 100; i++ {
		t := gc.GetDelayLatency()
		if t > 0 {
			countLatency++
		}
	}
	require.True(t, countLatency > 0)
}
