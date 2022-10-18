package utils

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"

	"github.com/stretchr/testify/require"
)

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
	_, err := ParseTemplate("../../fixtures", b, map[string]interface{}{"Digest": "123"})
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
	_, err := ParseTemplate("../../fixtures", b, map[string]interface{}{})

	// THEN it should not fail without params
	require.NoError(t, err)

	// AND it should not fail with params
	_, err = ParseTemplate("../../fixtures", b, map[string]interface{}{"Account": "x123", "Money": 12, "Note": "ty"})
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
	_, err := ParseTemplate("../../fixtures", b, map[string]interface{}{"t_av": 10.0, "device_id": "ABC"})

	// THEN it should not fail
	require.NoError(t, err)
}

func Test_ShouldParseScenarioTemplate(t *testing.T) {
	scenarioFiles := []string{
		"../../fixtures/scenario1.yaml",
		"../../fixtures/scenario2.yaml",
		"../../fixtures/scenario3.yaml",
	}
	for _, scenarioFile := range scenarioFiles {
		// GIVEN a mock scenario loaded from YAML
		b, err := ioutil.ReadFile(scenarioFile)
		require.NoError(t, err)

		// WHEN parsing YAML for contents tag
		body, err := ParseTemplate("../../fixtures", b,
			map[string]interface{}{"ETag": "abc", "Page": 1, "PageSize": 10, "Nonce": 1, "SleepSecs": 5})
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
		require.Contains(t, scenario.Response.ContentType, "application/json")
	}
}

func Test_ShouldValidateFileName(t *testing.T) {
	require.True(t, validFileName("file-name-13"))
	require.True(t, validFileName("file_name_123.txt"))
	require.False(t, validFileName("file_name_12..txt"))
	require.False(t, validFileName("/file_name_12..txt"))
	require.False(t, validFileName("./file_name_12..txt"))
}
