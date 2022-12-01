package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ShouldValidateBuildMockScenarioKeyData(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	// WHEN creating key data
	// THEN it should succeed
	keyData := scenario.ToKeyData()

	require.Equal(t, "path1-test1-abc", scenario.NormalName())
	require.Equal(t, "", keyData.PathPrefix(0))
	require.Equal(t, "/path1", keyData.PathPrefix(1))
	require.Equal(t, "/path1/test1", keyData.PathPrefix(2))
	require.Equal(t, "/path1/test1/abc", keyData.PathPrefix(3))
	require.Equal(t, "/path1/test1/abc", keyData.PathPrefix(4))
}

func Test_ShouldNotValidateEmptyMockScenarioKeyData(t *testing.T) {
	// GIVEN a empty mock keyData
	keyData := &MockScenarioKeyData{}
	// WHEN validating keyData
	// THEN it should fail
	require.Error(t, keyData.Validate())
	keyData.Method = Get
	require.Error(t, keyData.Validate())
	keyData.Path = "/path1//\\\\//test1/2///"
	require.Error(t, keyData.Validate())
	keyData.Name = "test1"
	require.NoError(t, keyData.Validate())
	require.True(t, keyData.String() != "")
	require.True(t, keyData.PartialMethodPathKey() != "")
	require.True(t, keyData.MethodNamePathPrefixKey() != "")
}

func Test_ShouldCompareNotEqualsMockScenarioKeyData(t *testing.T) {
	keyData := &MockScenarioKeyData{}
	keyData.Path = "/xxx"
	require.Error(t, keyData.Equals(buildScenario().ToKeyData()))
	keyData.Method = Post
	require.Error(t, keyData.Equals(buildScenario().ToKeyData()))
	keyData.Name = "scenario"
	require.Error(t, keyData.Equals(buildScenario().ToKeyData()))
	keyData.Path = "/path1/test1/abc"
	require.NoError(t, keyData.Equals(buildScenario().ToKeyData()))
}

func Test_ShouldValidateWithoutPath(t *testing.T) {
	keyData := &MockScenarioKeyData{}
	keyData.Method = Post
	require.Error(t, keyData.Equals(buildScenario().ToKeyData()))
	keyData.Path = "/path1/test1/abc"
	require.Error(t, keyData.Equals(buildScenario().ToKeyData()))
	keyData.Name = "scenario"
	require.NoError(t, keyData.Equals(buildScenario().ToKeyData()))
}

func Test_ShouldCompareEqualsMockScenarioKeyData(t *testing.T) {
	keyData1 := buildScenario().ToKeyData()
	keyData2 := buildScenario().ToKeyData()
	require.NoError(t, keyData1.Equals(keyData2))
	keyData1.MatchHeaders["abc"] = "000"
	require.Error(t, keyData1.Equals(keyData2))
	keyData1.MatchContents = "content1"
	require.Error(t, keyData1.Equals(keyData2))
	keyData1.MatchContentType = "yaml"
	require.Error(t, keyData1.Equals(keyData2))
	keyData1.MatchQueryParams["xyz"] = "111"
	require.Error(t, keyData1.Equals(keyData2))
}
