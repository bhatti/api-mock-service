package types

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_ShouldValidateProperMockScenario(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	// WHEN validating scenario
	// THEN it should succeed
	require.NoError(t, scenario.Validate())
	require.Equal(t, "path1/test1/abc", scenario.NormalPath('/'))
}

func Test_ShouldValidateBuildMockScenarioKeyData(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	// WHEN creating key data
	// THEN it should succeed
	keyData := scenario.ToKeyData()

	require.Equal(t, "", keyData.PathPrefix(0))
	require.Equal(t, "/path1", keyData.PathPrefix(1))
	require.Equal(t, "/path1/test1", keyData.PathPrefix(2))
	require.Equal(t, "/path1/test1/abc", keyData.PathPrefix(3))
	require.Equal(t, "/path1/test1/abc", keyData.PathPrefix(4))
}

func Test_ShouldNotValidateEmptyMockScenario(t *testing.T) {
	// GIVEN a empty mock scenario repository
	scenario := &MockScenario{}
	// WHEN validating scenario
	// THEN it should fail
	require.Error(t, scenario.Validate())
	scenario.Method = Get
	scenario.Path = "/path1//\\\\//test1/2///"
	require.Error(t, scenario.Validate())
	scenario.Name = "test1"
	require.Error(t, scenario.Validate())
	scenario.Response.Contents = "test"
	require.NoError(t, scenario.Validate())
	require.Equal(t, "path1/test1/2", scenario.NormalPath('/'))
}

func buildScenario() *MockScenario {
	scenario := &MockScenario{
		Method:      Post,
		Name:        "scenario",
		Path:        "/path1/\\\\//test1//abc////",
		Description: "",
		Request: MockHTTPRequest{
			QueryParams: "a=1&b=2",
			Headers: map[string][]string{
				"CTag": {"981"},
			},
		},
		Response: MockHTTPResponse{
			Headers: map[string][]string{
				"ETag": {"123"},
			},
			ContentType: "application/json",
			Contents:    "test body",
			StatusCode:  200,
		},
		WaitBeforeReply: time.Duration(1) * time.Second,
	}
	return scenario
}
