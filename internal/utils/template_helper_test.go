package utils

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"

	"github.com/stretchr/testify/require"
)

func Test_ShouldParsePredicateFalse(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{false}}`)
	// WHEN parsing template
	out, err := ParseTemplate("", b, map[string]any{})
	// THEN it should not fail
	require.NoError(t, err)
	require.Equal(t, "false", string(out))
}

func Test_ShouldParsePredicateTrue(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{false}}`)
	// WHEN parsing template
	out, err := ParseTemplate("", b, map[string]any{})
	// THEN it should not fail
	require.NoError(t, err)
	require.Equal(t, "false", string(out))
}

func Test_ShouldParsePredicateMissing(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{GERequest 3}}`)
	// WHEN parsing template
	out, err := ParseTemplate("", b, map[string]any{})
	// THEN it should not fail
	require.NoError(t, err)
	require.Equal(t, "false", string(out))
}

func Test_ShouldParsePredicateMismatch(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{GERequest 3}}`)
	// WHEN parsing template
	out, err := ParseTemplate("", b, map[string]any{types.RequestCount: 1})
	// THEN it should not fail
	require.NoError(t, err)
	require.Equal(t, "false", string(out))
}

func Test_ShouldParsePredicateMatch(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{GERequest 3}}`)
	// WHEN parsing template
	out, err := ParseTemplate("", b, map[string]any{types.RequestCount: 3})
	// THEN it should not fail
	require.NoError(t, err)
	require.Equal(t, "true", string(out))
}

func Test_ShouldParseAdd(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{Add 3 5}}`)
	// WHEN parsing template
	out, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should add numbers
	require.NoError(t, err)
	require.Equal(t, "8", string(out))
}

func Test_ShouldParseUnescape(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{Unescape "test"}}`)
	// WHEN parsing unescape
	out, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "test", string(out))
}

func Test_ShouldParseRandNumMinMax(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandNumMinMax 1 10}}`)
	// WHEN parsing random number
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandNumMax(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandNumMax 1}}`)
	// WHEN parsing random number
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseUdid(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{Udid}}`)
	// WHEN parsing udid
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseSeededUdid(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{SeededUdid 3}}`)
	// WHEN parsing udid
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandCity(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandCity}}`)
	// WHEN parsing city
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseSeededCity(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{SeededCity 3}}`)
	// WHEN parsing city
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandBool(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandBool}}`)
	// WHEN parsing bool
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseSeededBool(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{SeededBool 5}}`)
	// WHEN parsing bool
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseEnumString(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{EnumString "one" "two"}}`)
	// WHEN parsing name
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseEnumInt(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{EnumInt 3 5}}`)
	// WHEN parsing name
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseDateTime(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{Time}}{{Date}}{{TimeFormat "2006"}}`)
	// WHEN parsing name
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseTimeFormat(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{TimeFormat "mm"}}`)
	// WHEN parsing name
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandCountries(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandCountry}}{{RandCountryCode}}`)
	// WHEN parsing name
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseSeededCountries(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{SeededCountry 1}}{{SeededCountryCode 1}}`)
	// WHEN parsing name
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandName(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandName}}`)
	// WHEN parsing name
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseSeededName(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{SeededName 11}}`)
	// WHEN parsing name
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandString(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandString 10}}`)
	// WHEN parsing string
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandStringMinMax(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandStringMinMax 1 10}}`)
	// WHEN parsing string
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandStringArrayMinMax(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandStringArrayMinMax 1 10}}`)
	// WHEN parsing string array
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)

	// GIVEN a template string with 0 max
	b = []byte(`{{RandStringArrayMinMax 1 0}}`)
	// WHEN parsing string array
	_, err = ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandIntArrayMinMax(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandIntArrayMinMax 1 10}}`)
	// WHEN parsing string array
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)

	// GIVEN a template string with 0 max
	b = []byte(`{{RandIntArrayMinMax 1 0}}`)
	// WHEN parsing string array
	_, err = ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandRegex(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandRegex "\\d{3}"}}`)
	// WHEN parsing string regex
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandRegexPhone(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandPhone}}`)
	// WHEN parsing string regex
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandRegexEmail(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandEmail}}`)
	// WHEN parsing string regex
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandRegexHost(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandHost}}`)
	// WHEN parsing string regex
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandRegexURL(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandURL}}`)
	// WHEN parsing string regex
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseInt(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{Int 3}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseFloat(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{Float 3}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseLT(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{LT 3 5}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseLE(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{LE 3 5}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseEQ(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{EQ 3 5}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseGT(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{GT 3 5}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseGE(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{GE 3 5}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseNth(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{Nth 3 5}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseYAMLFileProperty(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{YAMLFileProperty "../../fixtures/props.yaml" "token"}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseFilePropertyJson(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{JSONFileProperty "../../fixtures/props.yaml" "token"}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseFileProperty(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{FileProperty "../../fixtures/props.yaml" "token"}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandFileLine(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandFileLine "../../fixtures/lines.txt"}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseRandSeededFileLine(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{SeededFileLine "../../fixtures/lines.txt" 3}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParsePredicateForNthRequest(t *testing.T) {
	keyData := &types.MockScenarioKeyData{}
	require.True(t, MatchScenarioPredicate(keyData, &types.MockScenarioKeyData{}, 0))
	keyData.Predicate = `{{NthRequest 3}}`
	require.True(t, MatchScenarioPredicate(keyData, &types.MockScenarioKeyData{}, 0))
	require.False(t, MatchScenarioPredicate(keyData, &types.MockScenarioKeyData{}, 2))
	require.True(t, MatchScenarioPredicate(keyData, &types.MockScenarioKeyData{}, 3))
}

func Test_ShouldParseRandDict(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandDict}}`)
	// WHEN parsing int
	out, err := ParseTemplate("", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
	dict := make(map[string]any)
	err = json.Unmarshal(out, &dict)
	require.NoError(t, err)
	require.True(t, len(dict) >= 1)
}

func Test_ShouldParseNthRequest(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{NthRequest 3}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseNthRequestWithData(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{NthRequest 3}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{types.RequestCount: 1})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseGERequestWithStringData(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{GERequest 3}}`)
	// WHEN parsing without count
	_, err := ParseTemplate("../../fixtures", b, map[string]string{})
	// THEN it should succeed
	require.NoError(t, err)
	// WHEN parsing int
	_, err = ParseTemplate("../../fixtures", b, map[string]string{types.RequestCount: "1"})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseGERequestWithData(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{GERequest 3}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{types.RequestCount: 1})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseLTRequestWithStringData(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{LTRequest 3}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]string{types.RequestCount: "1"})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseLTRequestWithData(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{LTRequest 3}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]any{types.RequestCount: 1})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseWord(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandWord 1 10}}`)
	// WHEN parsing int
	_, err := ParseTemplate("", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseSentence(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandSentence 1 10}}`)
	// WHEN parsing int
	_, err := ParseTemplate("", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseParagraph(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{RandParagraph 1 10}}`)
	// WHEN parsing int
	_, err := ParseTemplate("", b, map[string]any{})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldParseNthRequestWithStringData(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{NthRequest 3}}`)
	// WHEN parsing int
	_, err := ParseTemplate("../../fixtures", b, map[string]string{types.RequestCount: "1"})
	// THEN it should succeed
	require.NoError(t, err)
}

func Test_ShouldNotHaveExtraSpace(t *testing.T) {
	// GIVEN a template string
	b := []byte(`
<script type="text/javascript">
    function load_{{.Digest}}() {
        document.getElementById("log_btn_{{.Digest}}").hidden = true;
        let xmlhttp = new XMLHttpRequest();
        xmlhttp.open("GET", "{{.DashboardRawURL}}", false);
        xmlhttp.send();
        document.getElementById("logs_{{.Digest}}").textContent = xmlhttp.responseText;
    }
</script>
`)

	// WHEN parsing template
	_, err := ParseTemplate("../../fixtures", b, map[string]any{"Digest": "123"})
	// THEN it should not fail
	require.NoError(t, err)
}

func Test_ShouldFailOnNoVariables(t *testing.T) {
	// GIVEN a template string
	b := []byte(`
{{with .Account -}}
Account: {{.}}
{{- end}}
Money: {{.Money}}
{{if .Note -}}
Note: {{.Note}}
{{- end}}
`)

	// WHEN parsing template
	_, err := ParseTemplate("../../fixtures", b, map[string]any{})

	// THEN it should not fail without params
	require.NoError(t, err)

	// AND it should not fail with params
	_, err = ParseTemplate("../../fixtures", b, map[string]any{"Account": "x123", "Money": 12, "Note": "ty"})
	require.NoError(t, err)
}

func Test_ShouldParseExpression(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{"device_id": {{.device_id}}, "description": "
  {{if lt .t_av 30.0}}
    Current temperature is {{.t_av}}, it's normal."
  {{else if ge .t_av 30.0}}
    Current temperature is {{.t_av}}, it's high."
  {{end}}
}`)

	// WHEN parsing template
	_, err := ParseTemplate("../../fixtures", b, map[string]any{"t_av": 10.0, "device_id": "ABC"})

	// THEN it should not fail
	require.NoError(t, err)
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
		body, err := ParseTemplate("../../fixtures", b,
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
	body, err := ParseTemplate("../../fixtures", b,
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
	body, err := ParseTemplate("../../fixtures", b, map[string]any{})

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
		body, err := ParseTemplate("../../fixtures", b,
			map[string]any{"ETag": "abc", "page": i, "pageSize": 5, types.RequestCount: i})

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

func Test_ShouldParseArrayVariableSizeEQTrue(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableSizeEQ "contents.arr" 4}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 10, "title": "hello world", "completed": true, "arr": []int{1, 2, 3, 4}}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "true", string(b))
}

func Test_ShouldParseStringArrayVariableSizeEQTrue(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableSizeEQ "contents.arr" 2}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 10, "title": "hello world", "completed": true, "arr": []string{"one", "two"}}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "true", string(b))
}

func Test_ShouldParseArrayVariableSizeEQFalse(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableSizeEQ "contents.arr" 5}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 10, "title": "hello world", "completed": true, "arr": []float64{1, 2, 3, 4}}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "false", string(b))
}

func Test_ShouldParseVariableSizeEQTrue(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableSizeEQ "contents" 4}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 10, "title": "hello world", "completed": true}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "true", string(b))
}

func Test_ShouldParseVariableSizeEQFalse(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableSizeEQ "contents" 5}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 10, "title": "hello world", "completed": true}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "false", string(b))
}

func Test_ShouldParseVariableSizeLE(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableSizeLE "contents" 4}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 10, "title": "hello world", "completed": true}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "true", string(b))
}

func Test_ShouldParseVariableSizeGE(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableSizeGE "contents" 1}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 10, "title": "hello world", "completed": true}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "true", string(b))
}

func Test_ShouldParseVariableEquals(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableEquals "contents.id" "10"}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 10, "title": "hello world", "completed": true}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "true", string(b))
}

func Test_ShouldParseVariableNotEquals(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableEquals "contents.id" "10"}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 101, "title": "hello world", "completed": true}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "false", string(b))
}

func Test_ShouldParseVariableContains(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableContains "contents.id" "10"}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 10, "title": "hello world", "completed": true}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "true", string(b))
}

func Test_ShouldParseVariableNotContains(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableContains "contents.id" "20"}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 10, "title": "hello world", "completed": true}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "false", string(b))
}

func Test_ShouldParseVariableNotPartialContains(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableContains "contents.id" "10"}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 201, "title": "hello world", "completed": true}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "false", string(b))
}

func Test_ShouldParseVariablePartialContains(t *testing.T) {
	// GIVEN a template string
	b := []byte(`{{VariableContains "contents.id" "10"}}`)
	// AND contents
	contents := map[string]any{"userId": 1, "id": 101, "title": "hello world", "completed": true}
	// WHEN parsing string
	b, err := ParseTemplate("", b, map[string]any{"contents": contents})
	// THEN it should succeed
	require.NoError(t, err)
	require.Equal(t, "true", string(b))
}

func Test_ShouldValidateFileName(t *testing.T) {
	require.True(t, validFileName("file-name-13"))
	require.True(t, validFileName("file_name_123.txt"))
	require.False(t, validFileName("file_name_12..txt"))
	require.False(t, validFileName("/file_name_12..txt"))
	require.False(t, validFileName("./file_name_12..txt"))
}

func Test_ShouldConvertToInt64(t *testing.T) {
	require.Equal(t, int64(0), toInt64(nil))
	require.Equal(t, int64(10), toInt64(int64(10)))
	require.Equal(t, int64(10), toInt64(int32(10)))
	require.Equal(t, int64(10), toInt64(10))
	require.Equal(t, int64(10), toInt64(uint(10)))
	require.Equal(t, int64(10), toInt64("10"))
}

func Test_ShouldConvertToInt(t *testing.T) {
	require.Equal(t, 0, toInt(nil))
	require.Equal(t, 10, toInt(int64(10)))
	require.Equal(t, 10, toInt(int32(10)))
	require.Equal(t, 10, toInt(10))
	require.Equal(t, 10, toInt(uint(10)))
}

func Test_ShouldConvertToFloat64(t *testing.T) {
	var f32 float32 = 10
	var f64 float64 = 10
	var i64 int64 = 10
	var u64 uint64 = 10
	require.Equal(t, float64(0), ToFloat64(nil))
	require.Equal(t, float64(10), ToFloat64(float64(10)))
	require.Equal(t, float64(10), ToFloat64(float32(10)))
	require.Equal(t, float64(10), ToFloat64(10))
	require.Equal(t, float64(10), ToFloat64(uint(10)))
	require.Equal(t, float64(10), ToFloat64(int32(10)))
	require.Equal(t, float64(10), ToFloat64(int64(10)))
	require.Equal(t, float64(10), ToFloat64(uint64(10)))
	require.Equal(t, float64(10), ToFloat64(&f32))
	require.Equal(t, float64(10), ToFloat64(&f64))
	require.Equal(t, float64(10), ToFloat64(&i64))
	require.Equal(t, float64(10), ToFloat64(&u64))
}

func Test_ShouldFindVariable(t *testing.T) {
	require.Nil(t, findVariable("k", nil))
	require.Nil(t, findVariable("k", ""))
	require.NotNil(t, findVariable("k", map[string]string{"k": "1"}))
	require.Nil(t, findVariable("k", map[string]string{"x": "1"}))
	require.NotNil(t, findVariable("k.a", map[string]any{"k": map[string]string{"a": "b"}}))
}

func Test_ShouldCompareVariable(t *testing.T) {
	require.False(t, variableEquals("k", nil, 1))
	require.False(t, variableEquals("k", "", 1))
	require.True(t, variableEquals("k", map[string]string{"k": "1"}, "1"))
	require.False(t, variableEquals("k", map[string]string{"x": "1"}, "1"))
	require.True(t, variableEquals("k.a", map[string]any{"k": map[string]string{"a": "b"}}, "b"))
}
