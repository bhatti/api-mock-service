package pm

import (
	"encoding/json"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"net/url"
	"os"
	"testing"
	"time"
)

func Test_ShouldBuildPostman(t *testing.T) {
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
	u, err := url.Parse("https://localhost:8080")
	require.NoError(t, err)
	scenario.BaseURL = u.String()
	c := ConvertScenariosToPostman(scenario.Name, scenario)
	j, _ := json.Marshal(c)
	require.True(t, len(j) > 0)

	scenarios := ConvertPostmanToScenarios(config, c, time.Time{}, time.Time{})
	require.Equal(t, 1, len(scenarios))
}

func Test_ShouldParseRegexAndReplaceBackVariables(t *testing.T) {
	s := replaceTemplateVariables(`{{BaseUri_PreRegion_Pool}}{{UserRegion}}{{BaseUri_PostRegion}}`)
	require.Equal(t, "{{.BaseUri_PreRegion_Pool}}{{.UserRegion}}{{.BaseUri_PostRegion}}", s)
}

// See https://developer.twitter.com/en/docs/tutorials/postman-getting-started
func Test_ShouldReplacePostmanEvents(t *testing.T) {
	execs := []string{
		"pm.variables.set('ApiTargetNamespace', 'IdentityService.')",
		"pm.variables.set('Region', 'us-west-2')",
		"pm.request.headers.add({key: 'Content-Type', value: 'application/x-amz-json-1.1' })",
		"pm.request.headers.add({key: 'X-Target1', value: pm.variables.get('ApiTargetNamespace')+pm.info.requestName })",
		"pm.request.headers.add({key: 'X-Target2', value: pm.info.requestName + pm.variables.get('ApiTargetNamespace')+pm.info.requestName })",
		"const [userId] = pm.environment.get('access_token').split('-');",
		"pm.request.headers.add({key: 'X-Target3', value: pm.info.requestName + pm.environment.get('access_token')+pm.info.requestName })",
	}
	os.Setenv("access_token", "-abc123")
	config := types.BuildTestConfig()
	headers := make(map[string][]string)
	c := buildConverter(config, time.Now(), time.Now())
	for _, exec := range execs {
		c.handleEvent("test-name", exec, headers)
	}
	require.Equal(t, "IdentityService.", c.variables["ApiTargetNamespace"])
	require.Equal(t, "us-west-2", c.variables["Region"])
	require.Equal(t, "application/x-amz-json-1.1", headers[types.ContentTypeHeader][0])
	require.Equal(t, "IdentityService.test-name", headers["X-Target1"][0])
	require.Equal(t, "test-nameIdentityService.test-name", headers["X-Target2"][0])
	require.Equal(t, "test-name-abc123test-name", headers["X-Target3"][0])
}
