package utils

import (
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"testing"
	"time"
)

const mockPath = "//abc//\\def/123/"

func Test_ShouldParsePredicateForNthRequest(t *testing.T) {
	keyData1 := buildScenario(types.Post, "test1", mockPath, 1).ToKeyData()
	keyData2 := buildScenario(types.Post, "test2", mockPath, 2).ToKeyData()
	require.True(t, MatchScenarioPredicate(keyData1, keyData2, 0))
	keyData1.MatchQueryParams = map[string]string{"a": `\d+`, "b": "abc"}
	keyData2.MatchQueryParams = map[string]string{"a": `\d+`, "b": "abc"}
	keyData1.Predicate = `{{NthRequest 3}}`
	require.True(t, MatchScenarioPredicate(keyData1, keyData2, 0))
	require.False(t, MatchScenarioPredicate(keyData1, keyData2, 2))
	require.True(t, MatchScenarioPredicate(keyData1, keyData2, 3))
}

func Test_ShouldMatchScenarioPredicate(t *testing.T) {
	keyData := &types.MockScenarioKeyData{}
	require.True(t, MatchScenarioPredicate(keyData, &types.MockScenarioKeyData{}, 0))
	keyData.Predicate = `{{NthRequest 3}}`
	require.True(t, MatchScenarioPredicate(keyData, &types.MockScenarioKeyData{}, 0))
	require.False(t, MatchScenarioPredicate(keyData, &types.MockScenarioKeyData{}, 2))
	require.True(t, MatchScenarioPredicate(keyData, &types.MockScenarioKeyData{}, 3))
}

func Test_ShouldParseScenarioTemplate(t *testing.T) {
	scenarioFiles := []string{
		"../../fixtures/scenario1.yaml",
		"../../fixtures/scenario2.yaml",
		"../../fixtures/scenario3.yaml",
		"../../fixtures/account.yaml",
	}
	for _, scenarioFile := range scenarioFiles {
		// GIVEN a mock scenario loaded from YAML
		b, err := os.ReadFile(scenarioFile)
		require.NoError(t, err)

		// WHEN parsing YAML for contents tag
		body, err := fuzz.ParseTemplate("../../fixtures", b,
			map[string]any{"ETag": "abc", "Page": 1, "PageSize": 10, "Nonce": 1, "SleepSecs": 5})

		// THEN it should not fail
		require.NoError(t, err)
		scenario := types.MockScenario{}
		// AND it should return valid mock scenario
		err = yaml.Unmarshal(body, &scenario)
		if err != nil {
			t.Logf("faile parsing %s\n%s\n", scenarioFile, body)
		}
		require.NoError(t, err)
		// AND it should have expected contents

		require.Contains(t, scenario.Response.Headers["ETag"], "abc")
		require.Contains(t, scenario.Response.ContentType(), "application/json",
			fmt.Sprintf("%v in %s", scenario.Response.Headers, scenarioFile))
	}
}

func Test_ShouldParseCustomerStripeTemplate(t *testing.T) {
	// GIVEN a mock scenario loaded from YAML
	b, err := os.ReadFile("../../fixtures/stripe-customer.yaml")
	require.NoError(t, err)

	// WHEN parsing YAML for contents tag
	body, err := fuzz.ParseTemplate("../../fixtures", b,
		map[string]any{"ETag": "abc", "Page": 1, "PageSize": 10, "Nonce": 1, "SleepSecs": 5})

	// THEN it should not fail
	require.NoError(t, err)
	scenario := types.MockScenario{}
	// AND it should return valid mock scenario
	err = yaml.Unmarshal(body, &scenario)
	require.NoError(t, err)
	// AND it should have expected contents

	require.Equal(t, "Bearer sk_test_[0-9a-fA-F]{10}$", scenario.Request.MatchHeaders["Authorization"])
	require.Contains(t, scenario.Response.ContentType(), "application/json")
}

func Test_ShouldParseCommentsTemplate(t *testing.T) {
	// GIVEN a mock scenario loaded from YAML
	b, err := os.ReadFile("../../fixtures/list_comments.yaml")
	require.NoError(t, err)

	// WHEN parsing YAML for contents tag
	body, err := fuzz.ParseTemplate("../../fixtures", b, map[string]any{})

	// THEN it should not fail
	require.NoError(t, err)
	scenario := types.MockScenario{}
	// AND it should return valid mock scenario
	err = yaml.Unmarshal(body, &scenario)
	require.NoError(t, err)
	// AND it should have expected contents
	require.True(t, scenario.Response.StatusCode == 200 || scenario.Response.StatusCode == 400)
	require.Contains(t, scenario.Response.ContentType(), "application/json")
}

func Test_ShouldParseDevicesTemplate(t *testing.T) {
	// GIVEN a mock scenario loaded from YAML
	b, err := os.ReadFile("../../fixtures/devices.yaml")
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		// WHEN parsing YAML for contents tag
		body, err := fuzz.ParseTemplate("../../fixtures", b,
			map[string]any{"ETag": "abc", "page": i, "pageSize": 5, fuzz.RequestCount: i})

		// THEN it should not fail
		require.NoError(t, err)
		scenario := types.MockScenario{}
		// AND it should return valid mock scenario
		err = yaml.Unmarshal(body, &scenario)
		require.NoError(t, err)
		// AND it should have expected contents
		if i%10 == 0 {
			require.True(t, scenario.Response.StatusCode == 500 || scenario.Response.StatusCode == 501)
		} else {
			require.True(t, scenario.Response.StatusCode == 200 || scenario.Response.StatusCode == 400)
		}
		require.Contains(t, scenario.Response.ContentType(), "application/json")
	}
}
func buildScenario(method types.MethodType, name string, path string, n int) *types.MockScenario {
	return &types.MockScenario{
		Method:      method,
		Name:        name,
		Path:        path,
		Group:       path,
		Description: name,
		Request: types.MockHTTPRequest{
			MatchQueryParams: map[string]string{"a": `\d+`, "b": "abc"},
			MatchHeaders: map[string]string{
				types.ContentTypeHeader: "application/json",
				"ETag":                  `\d{3}`,
			},
		},
		Response: types.MockHTTPResponse{
			Headers: map[string][]string{
				"ETag":                  {strconv.Itoa(n)},
				types.ContentTypeHeader: {"application/json"},
			},
			Contents:   "test body",
			StatusCode: 200,
		},
		WaitBeforeReply: time.Duration(1) * time.Second,
	}
}
