package har

import (
	"encoding/json"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_ShouldBuildHar(t *testing.T) {
	// GIVEN a valid scenario
	scenario := types.BuildTestScenario(types.Post, "test -name", "/path", 0)
	scenario.Group = "har-group"
	scenario.Request.Headers = map[string]string{
		types.ContentTypeHeader: "application/json 1.1",
	}
	scenario.Request.QueryParams = map[string]string{
		"abc": "123",
	}
	config := types.BuildTestConfig()
	har := BuildHar(config, scenario, "http://host:100", "localhost:8080", time.Now(), time.Now().Add(time.Second))
	j, _ := json.Marshal(har)
	require.True(t, len(j) > 0)
}
