package repository

import (
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldSaveAndGetGroupConfig(t *testing.T) {
	// GIVEN a contents repository
	groupConfigRepository, err := NewFileGroupConfigRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	names := []string{"root", "global", "test1", "test1abc", "test1abc1", "test2"}
	for _, name := range names {
		gc := &types.GroupConfig{
			Variables:                        map[string]string{name + "v1": "val"},
			MeanTimeBetweenAdditionalLatency: 8,
			MeanTimeBetweenFailure:           4,
			MaxAdditionalLatencySecs:         3,
			HTTPErrors:                       []int{400, 401},
		}
		// WHEN saving group config
		err = groupConfigRepository.Save(name, gc)
		// THEN it should succeed
		require.NoError(t, err)

		// AND should return saved scenario
		loaded, err := groupConfigRepository.Load(name)
		require.NoError(t, err)
		require.Equal(t, gc.Variables, loaded.Variables)
		require.Equal(t, gc.MeanTimeBetweenFailure, loaded.MeanTimeBetweenFailure)
		require.Equal(t, gc.MeanTimeBetweenAdditionalLatency, loaded.MeanTimeBetweenAdditionalLatency)
		require.Equal(t, gc.MaxAdditionalLatencySecs, loaded.MaxAdditionalLatencySecs)
		require.Equal(t, gc.HTTPErrors, loaded.HTTPErrors)
	}
	vars := groupConfigRepository.Variables("test1abc")
	require.Equal(t, 4, len(vars))
	vars = groupConfigRepository.Variables("")
	require.Equal(t, 2, len(vars))
}
