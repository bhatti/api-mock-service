package repository

import (
	"fmt"
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"

	"github.com/stretchr/testify/require"
)

const FixturePath = "//xyz//\\def/123/"

func Test_ShouldSaveAndGetFixtures(t *testing.T) {
	// GIVEN a contents repository
	fixtureRepository, err := NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// WHEN saving contents
	data := []byte("data contents")
	err = fixtureRepository.Save(types.Post, "data_0", FixturePath, data)
	// THEN it should succeed
	require.NoError(t, err)

	// AND should return saved scenario
	saved, err := fixtureRepository.Get(types.Post, "data_0", FixturePath)
	require.NoError(t, err)
	require.Equal(t, data, saved)
}

func Test_ShouldListMockFixtureNames(t *testing.T) {
	// GIVEN a mock fixture repository
	fixtureRepository, err := NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND a set of fixtures
	for i := 0; i < 10; i++ {
		err = fixtureRepository.Save(types.Get, fmt.Sprintf("data_%d", i), FixturePath, []byte(fmt.Sprintf("test -data for %d", i)))
		require.NoError(t, err)
	}
	// WHEN listing fixtures
	names, err := fixtureRepository.GetFixtureNames(types.Get, FixturePath)
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		require.Equal(t, fmt.Sprintf("data_%d", i), names[i])
		err = fixtureRepository.Delete(types.Get, names[i], FixturePath)
		require.NoError(t, err)
	}
}

func Test_ShouldNotGetAfterDeletingFixtures(t *testing.T) {
	// GIVEN a mock contents repository
	fixtureRepository, err := NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// WHEN saving contents
	data := []byte("data contents")
	err = fixtureRepository.Save(types.Delete, "data1", FixturePath, data)
	// THEN it should succeed
	require.NoError(t, err)

	// AND should return saved scenario
	saved, err := fixtureRepository.Get(types.Delete, "data1", FixturePath)
	require.NoError(t, err)
	require.Equal(t, data, saved)

	// But WHEN DELETING the mock scenario
	err = fixtureRepository.Delete(types.Delete, "data1", FixturePath)
	require.NoError(t, err)

	// THEN GET should fail
	_, err = fixtureRepository.Get(types.Delete, "data1", FixturePath)
	require.Error(t, err)
}
