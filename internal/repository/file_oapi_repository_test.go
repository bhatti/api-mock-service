package repository

import (
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_ShouldSaveAndGetOAPI(t *testing.T) {
	// GIVEN a contents repository
	oapiRepository, err := NewFileOAPIRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// WHEN saving contents
	b, err := os.ReadFile("../../fixtures/oapi/twilio_accounts_v1.yaml")
	err = oapiRepository.SaveRaw("test1", b)
	// THEN it should succeed
	require.NoError(t, err)

	// AND should return saved scenario
	saved, err := oapiRepository.Load("test1")
	require.NoError(t, err)
	require.Equal(t, "Twilio - Accounts", saved.Info.Title)

	// WHEN saving again as an object
	err = oapiRepository.Save("test1", saved)
	// THEN it should succeed
	require.NoError(t, err)

	// AND should return saved scenario
	_, err = oapiRepository.LoadRaw("test1")
	require.NoError(t, err)
}
