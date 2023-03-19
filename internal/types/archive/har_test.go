package archive

import (
	"encoding/json"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
	"time"
)

func Test_ShouldBuildHar(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a valid scenario
	scenario := types.BuildTestScenario(types.Post, "test -name", "/path", 0)
	scenario.Group = "archive-group"
	scenario.Request.Headers = map[string]string{
		types.ContentTypeHeader: "application/json 1.1",
	}
	scenario.Request.QueryParams = map[string]string{
		"abc": "123",
	}
	u, err := url.Parse("https://localhost:8080" + scenario.Path)
	require.NoError(t, err)
	har := ConvertScenariosToHar(config, u, time.Now(), time.Now().Add(time.Second), scenario)
	j, _ := json.Marshal(har)
	require.True(t, len(j) > 0)

	scenarios := ConvertHarToScenarios(config, har)
	require.Equal(t, 1, len(scenarios))
}

func Test_ShouldConvertHar(t *testing.T) {
	config := types.BuildTestConfig()
	// AND a valid scenario
	scenario := types.BuildTestScenario(types.Post, "test-name", "/path", 0)
	scenario.Group = "archive-group"
	scenario.Request.Headers = map[string]string{
		types.ContentTypeHeader: "application/json 1.1",
	}
	scenario.Request.QueryParams = map[string]string{
		"abc": "123",
	}
	u, err := url.Parse("https://localhost:8080" + scenario.Path)
	require.NoError(t, err)
	har := ConvertScenariosToHar(config, u, time.Now(), time.Now().Add(time.Second), scenario)
	require.Equal(t, 1, len(har.Log.Entries))
}
