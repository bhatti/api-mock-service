package pm

import (
	"encoding/json"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"os"
	"strings"
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

	scenarios, vars := ConvertPostmanToScenarios(config, c, time.Time{}, time.Time{})
	require.Len(t, scenarios, 1)
	require.Len(t, vars.Variables, 2)
}

func Test_ShouldParseRegexAndReplaceBackVariables(t *testing.T) {
	s := replaceTemplateVariables(`{{BaseUri_PreRegion_Pool}}{{UserRegion}}{{BaseUri_PostRegion}}`)
	require.Equal(t, "{{.BaseUri_PreRegion_Pool}}{{.UserRegion}}{{.BaseUri_PostRegion}}", s)
}

// Test_ShouldProcessPostmanScripts replaces Test_ShouldReplacePostmanEvents
func Test_ShouldProcessPostmanScripts(t *testing.T) {
	// Setup test
	config := types.BuildTestConfig()
	converter := NewPostmanConverter(config, time.Now(), time.Now())
	processor := NewScriptConverter(converter.context)
	headers := make(http.Header)

	// Set environment variable
	_ = os.Setenv("access_token", "-abc123")

	// Test scripts
	scripts := []struct {
		exec     string
		expected map[string]string
	}{
		{
			exec: "pm.variables.set('ApiTargetNamespace', 'IdentityService.')",
			expected: map[string]string{
				"ApiTargetNamespace": "IdentityService.",
			},
		},
		{
			exec: "pm.variables.set('Region', 'us-west-2')",
			expected: map[string]string{
				"Region": "us-west-2",
			},
		},
		{
			exec: `pm.request.headers.add({key: "Content-Type", value: "application/x-amz-json-1.1" })`,
			expected: map[string]string{
				"Content-Type": "application/x-amz-json-1.1",
			},
		},
		{
			exec: `pm.request.headers.add({key: "X-Target1", value: pm.variables.get("ApiTargetNamespace")+pm.info.requestName })`,
			expected: map[string]string{
				"X-Target1": "IdentityService.test-name",
			},
		},
		{
			exec: `pm.request.headers.add({key: "X-Target2", value: pm.info.requestName + pm.variables.get("ApiTargetNamespace")+pm.info.requestName })`,
			expected: map[string]string{
				"X-Target2": "test-nameIdentityService.test-name",
			},
		},
		{
			exec: `pm.request.headers.add({key: "X-Target3", value: pm.info.requestName + pm.environment.get("access_token")+pm.info.requestName })`,
			expected: map[string]string{
				"X-Target3": "test-name-abc123test-name",
			},
		},
	}

	// Process each script
	for _, s := range scripts {
		processor.ProcessScript(s.exec, "test-name", headers)
	}

	// Verify collection variables
	require.Equal(t, "IdentityService.", converter.context.CollectionVars["ApiTargetNamespace"])
	require.Equal(t, "us-west-2", converter.context.CollectionVars["Region"])

	// Verify headers
	require.Equal(t, "application/x-amz-json-1.1", headers.Get("Content-Type"))
	require.Equal(t, "IdentityService.test-name", headers.Get("X-Target1"), headers)
	require.Equal(t, "test-nameIdentityService.test-name", headers.Get("X-Target2"), headers)
	require.Equal(t, "test-name-abc123test-name", headers.Get("X-Target3"), headers)
}

func Test_ShouldProcessPostmanCollection(t *testing.T) {
	config := types.BuildTestConfig()
	file, err := os.Open("../../fixtures/twitter_postman.json")
	require.NoError(t, err)
	c, err := ParseCollection(file)
	require.NoError(t, err)
	scenarios, vars := ConvertPostmanToScenarios(config, c, time.Now(), time.Now())
	require.Len(t, scenarios, 137)
	require.Len(t, vars.Variables, 3)

	// Additional verification of script processing
	for _, scenario := range scenarios {
		// Verify variables are processed
		require.NotEmpty(t, scenario.Request.Variables)

		// Verify headers are processed
		if len(scenario.Request.Headers) > 0 {
			require.NotEmpty(t, scenario.Request.Headers)
		}

		// Verify scripts are converted to assertions
		if len(scenario.Request.Assertions) > 0 {
			matched := false
			for _, assertion := range scenario.Request.Assertions {
				if strings.HasPrefix(assertion, "PropertyMatches") { // PreRequest
					matched = true
				}
			}
			require.True(t, matched)
		}
	}
}

func Test_ShouldProcessPostmanBasicCollection(t *testing.T) {
	config := types.BuildTestConfig()
	file, err := os.Open("../../fixtures/postman_basic.json")
	require.NoError(t, err)

	// Parse collection
	c, err := ParseCollection(file)
	require.NoError(t, err)

	// Basic collection validation
	require.Equal(t, "API Testing Suite", c.Info.Name)
	require.Equal(t, 2, len(c.Items))  // Auth and CRUD folders
	require.Equal(t, 2, len(c.Events)) // Global pre-request and test scripts

	// Convert to scenarios
	scenarios, vars := ConvertPostmanToScenarios(config, c, time.Now(), time.Now())
	require.Len(t, scenarios, 5) // 1 auth + 4 CRUD operations
	require.Len(t, vars.Variables, 6)

	// GIVEN scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)

	// Test Auth Scenario specifically
	var authScenario *types.APIScenario
	for _, s := range scenarios {
		if strings.Contains(s.Name, "Get JWT Token") {
			authScenario = s
			break
		}
	}
	require.NotNil(t, authScenario)

	// ---- VERIFY SCRIPT CONVERSION -----

	// Verify global scripts are converted for each scenario
	for _, scenario := range scenarios {
		// Check if the global test script is in the description
		hasGlobalScript := strings.Contains(scenario.Response.Description, "Response time is acceptable")

		err = scenarioRepository.Save(scenario)
		require.NoError(t, err)

		// Verify response time assertion was converted (if global test script exists)
		if hasGlobalScript {
			responseTimeFound := false
			for _, assertion := range scenario.Response.Assertions {
				if strings.HasPrefix(assertion, "ResponseTimeMillisLE") {
					responseTimeFound = true
					require.Contains(t, assertion, "1000", "Response time limit not found in %s", scenario.Name)
					break
				}
			}
			require.True(t, responseTimeFound, "ResponseTimeMillisLE assertion not found %s: %v",
				scenario.Name, scenario.Response.Assertions)

			// Verify Content-Type assertion was converted
			ctAssertionFound := false
			for _, assertion := range scenario.Response.Assertions {
				if strings.Contains(assertion, "headers.Content-Type") {
					ctAssertionFound = true
					require.Contains(t, assertion, "application/json", "Content-Type value not found in %s", scenario.Name)
					break
				}
			}
			require.True(t, ctAssertionFound, "Content-Type assertion not found in %s: %v",
				scenario.Name, scenario.Response.Assertions)
		}

		// Verify conditional access token handling (this is a global prerequest script)
		if !strings.Contains(scenario.Name, "Get JWT Token") {
			require.Contains(t, scenario.NextRequest, "access_token",
				"Predicate should contain access_token condition: %s", scenario.NextRequest)
		}

	}

	// Verify auth scenario specific scripts

	// 1. Verify datetime converted from pre-request script
	require.Contains(t, authScenario.Request.Variables, "datetime",
		"datetime variable not found: %v", authScenario.Request.Variables)
	require.Equal(t, "{{ISODatetime}}", authScenario.Request.Variables["datetime"],
		"datetime not converted to ISODatetime")

	// 2. Verify environment variables extracted
	require.Contains(t, authScenario.Request.Variables, "client_name",
		"client_name variable not found: %v", authScenario.Request.Variables)
	require.Contains(t, authScenario.Request.Variables, "org_id",
		"org_id variable not found: %v", authScenario.Request.Variables)

	// 3. Verify status code assertion converted
	require.Equal(t, 200, authScenario.Response.StatusCode,
		"Status code not set from test script")
	statusCodeAssertionFound := false
	for _, assertion := range authScenario.Response.Assertions {
		if strings.HasPrefix(assertion, "ResponseStatusMatches") {
			statusCodeAssertionFound = true
			require.Contains(t, assertion, "200", "Status code value not found")
			break
		}
	}
	require.True(t, statusCodeAssertionFound, "Status code assertion not found: %v",
		authScenario.Response.Assertions)

	// 4. Verify access_token extraction
	require.Contains(t, authScenario.Response.AddSharedVariables, "access_token",
		"access_token not added to AddSharedVariables: %v", authScenario.Response.AddSharedVariables)

	// Verify original scripts are preserved in descriptions
	require.Contains(t, authScenario.Request.Description, "datetime",
		"Pre-request script not preserved in description")
	require.Contains(t, authScenario.Response.Description, "Status code is 200",
		"Test script not preserved in description")

	// ---- END SCRIPT CONVERSION TESTS ----

	// Verify auth request details
	require.Equal(t, "POST", string(authScenario.Method))
	require.Equal(t, "/auth/token", authScenario.Path)

	// Verify auth headers
	contentTypeFound := false
	apiKeyFound := false
	for key, values := range authScenario.Request.Headers {
		if strings.EqualFold(key, "Content-Type") {
			require.Contains(t, values, "application/json", authScenario.Request.Headers)
			contentTypeFound = true
		}
		if strings.EqualFold(key, "x-api-key") {
			require.Contains(t, values, "{{.api_key}}", authScenario.Request.Headers)
			apiKeyFound = true
		}
	}
	require.True(t, contentTypeFound, "Content-Type header not found", authScenario.Request.Headers)
	require.True(t, apiKeyFound, "x-api-key header not found", authScenario.Request.Headers)

	// Test CRUD Scenarios
	var createScenario *types.APIScenario
	for _, s := range scenarios {
		if strings.Contains(s.Name, "Create Resource") {
			createScenario = s
			break
		}
	}
	require.NotNil(t, createScenario)

	// Verify bearer auth
	require.Contains(t, createScenario.Authentication, "bearer")
	require.Equal(t, "{{.access_token}}", createScenario.Authentication["bearer"].Format)

	// Verify variables are populated
	require.NotEmpty(t, createScenario.Request.Variables)
	require.Contains(t, createScenario.Request.Variables, "base_url", createScenario.Request.Variables)
	require.Contains(t, createScenario.Request.Variables, "api_key", createScenario.Request.Variables)

	// Test CRUD operations paths
	var paths []string
	for _, scenario := range scenarios {
		if strings.Contains(scenario.Group, "CRUD Operations") {
			paths = append(paths, scenario.Path)
		}
	}
	require.Contains(t, paths, "/api/resources")
	require.Contains(t, paths, "/api/resources/{{.resource_id}}")

	// Verify all scenarios have variables
	for _, scenario := range scenarios {
		require.NotEmpty(t, scenario.Request.Variables,
			"Scenario %s has no variables", scenario.Name)
	}

	// Verify request body for create operation
	for _, scenario := range scenarios {
		if strings.Contains(scenario.Name, "Create Resource") {
			require.Contains(t, scenario.Request.Contents, "test resource")
			require.Contains(t, scenario.Request.Contents, "test description")
		}
	}

	// Verify HTTP methods for CRUD operations
	methodMap := make(map[string]bool)
	for _, scenario := range scenarios {
		if strings.Contains(scenario.Group, "CRUD Operations") {
			methodMap[string(scenario.Method)] = true
		}
	}
	require.True(t, methodMap["POST"])
	require.True(t, methodMap["GET"])
	require.True(t, methodMap["PATCH"])
	require.True(t, methodMap["DELETE"])
}

func Test_ScriptConverter_PreRequestScript(t *testing.T) {
	// Setup
	scenario := &types.APIScenario{
		Predicate: "{{NthRequest 1}}",
		Request: types.APIRequest{
			Variables: make(map[string]string),
		},
	}

	// Add pre-request script to description
	preReqScript := `if (!pm.environment.get('access_token')) {
    postman.setNextRequest('Get JWT Token');
}
const datetime = new Date().toISOString();
const payload = {
    'name': pm.environment.get('client_name'),
    'organization': pm.environment.get('org_id'),
    'datetime': datetime
};`

	// Convert scripts
	config := types.BuildTestConfig()
	converter := NewScriptConverter(NewPostmanConverter(config, time.Now(), time.Now()).context)
	converter.ConvertPreRequestScript(preReqScript, scenario)

	// Verify predicate was updated
	require.Contains(t, scenario.NextRequest, "access_token")
	require.Contains(t, scenario.NextRequest, "Get JWT Token")

	// Verify datetime variable was added
	require.Contains(t, scenario.Request.Variables, "datetime")
	require.Equal(t, "{{ISODatetime}}", scenario.Request.Variables["datetime"])

	// Verify environment variables were added
	require.Contains(t, scenario.Request.Variables, "client_name")
	require.Contains(t, scenario.Request.Variables, "org_id")
}

func Test_ScriptConverter_TestScript(t *testing.T) {
	// Setup
	scenario := &types.APIScenario{
		Response: types.APIResponse{
			Assertions:           []string{},
			AddSharedVariables:   []string{},
			AssertHeadersPattern: make(map[string]string),
		},
		Request: types.APIRequest{
			Variables: make(map[string]string),
		},
	}

	// Add test script to description
	testScript := `pm.test('Response time is acceptable', () => {
    pm.expect(pm.response.responseTime).to.be.below(1000);
});
pm.test('Response has valid structure', () => {
    pm.expect(pm.response.headers.get('Content-Type')).to.include('application/json');
});
pm.environment.set('access_token', responseBody.access_token);
pm.test('Status code is 200', function() {
    pm.response.to.have.status(200);
});`

	// Convert scripts
	config := types.BuildTestConfig()
	converter := NewScriptConverter(NewPostmanConverter(config, time.Now(), time.Now()).context)
	converter.ConvertTestScript(testScript, scenario)

	// Verify response time assertion was added
	responseTimeFound := false
	for _, assertion := range scenario.Response.Assertions {
		if assertion == "ResponseTimeMillisLE 1000" {
			responseTimeFound = true
			break
		}
	}
	require.True(t, responseTimeFound, "ResponseTimeMillisLE assertion not found: %v", scenario.Response.Assertions)

	// Verify status code was set
	require.Equal(t, 200, scenario.Response.StatusCode)

	// Verify content type assertion was added
	contentTypeFound := false
	for _, assertion := range scenario.Response.Assertions {
		if assertion == "PropertyMatches headers.Content-Type application/json" {
			contentTypeFound = true
			break
		}
	}
	require.True(t, contentTypeFound, "Content-Type assertion not found")

	// Verify content type header pattern was added
	require.Equal(t, "application/json", scenario.Response.AssertHeadersPattern["Content-Type"])

	// Verify set variable was added
	require.Contains(t, scenario.Response.AddSharedVariables, "access_token")

	// Verify variable reference was added
	require.NotContains(t, scenario.Request.Variables, "access_token", scenario.Name)
}

func Test_ExtractPathAfterBaseUrl(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple base_url with path",
			input:    "{{base_url}}/auth/token",
			expected: "/auth/token",
		},
		{
			name:     "base_url with multiple slashes",
			input:    "{{base_url}}///auth/token",
			expected: "/auth/token",
		},
		{
			name:     "base_url with path and query params",
			input:    "{{base_url}}/api/resources?param=value",
			expected: "/api/resources",
		},
		{
			name:     "base_url with path variables",
			input:    "{{base_url}}/api/resources/{{resource_id}}",
			expected: "/api/resources/{{resource_id}}",
		},
		{
			name:     "Plain URL without base_url",
			input:    "https://example.com/api/test",
			expected: "/api/test",
		},
		{
			name:     "Just a path",
			input:    "/api/test",
			expected: "/api/test",
		},
		{
			name:     "Path without leading slash",
			input:    "api/test",
			expected: "/api/test",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractPathAfterBaseUrl(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func Test_ParseRawUrl(t *testing.T) {
	// Test with various PostmanItems configurations

	// Test with base_url in Raw
	items1 := &PostmanItems{
		Request: &PostmanRequest{
			URL: &PostmanURL{
				Raw: "{{base_url}}/auth/token",
			},
		},
	}
	u1 := parseRawUrl(items1)
	require.Equal(t, "/auth/token", u1.Path)

	// Test with Path components
	items2 := &PostmanItems{
		Request: &PostmanRequest{
			URL: &PostmanURL{
				Path: []string{"auth", "token"},
			},
		},
	}
	u2 := parseRawUrl(items2)
	require.Equal(t, "/auth/token", u2.Path)

	// Test with resource_id variable in path
	items3 := &PostmanItems{
		Request: &PostmanRequest{
			URL: &PostmanURL{
				Raw: "{{base_url}}/api/resources/{{resource_id}}",
			},
		},
	}
	u3 := parseRawUrl(items3)
	require.Equal(t, "/api/resources/{{resource_id}}", u3.Path)

	// Test with query parameters
	items4 := &PostmanItems{
		Request: &PostmanRequest{
			URL: &PostmanURL{
				Raw: "{{base_url}}/api/search?q=test&limit=10",
			},
		},
	}
	u4 := parseRawUrl(items4)
	require.Equal(t, "/api/search", u4.Path)
	require.Contains(t, u4.RawQuery, "limit=10", u4.RawQuery)
	require.Contains(t, u4.RawQuery, "q=test", u4.RawQuery)

	// Test with no URL information
	items5 := &PostmanItems{
		Request: &PostmanRequest{
			URL: &PostmanURL{},
		},
	}
	u5 := parseRawUrl(items5)
	require.Equal(t, "/", u5.Path)
}

// Test for variable syntax conversion
func Test_VariableSyntaxConversion(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple variable",
			input:    "{{api_key}}",
			expected: "{{.api_key}}",
		},
		{
			name:     "Already converted variable",
			input:    "{{.api_key}}",
			expected: "{{.api_key}}",
		},
		{
			name:     "Template function",
			input:    "{{NthRequest 1}}",
			expected: "{{NthRequest 1}}",
		},
		{
			name:     "Conditional template",
			input:    "{{if not .access_token}}{{RequestByName \"Get JWT Token\"}}{{else}}{{NthRequest 1}}{{end}}",
			expected: "{{if not .access_token}}{{RequestByName \"Get JWT Token\"}}{{else}}{{NthRequest 1}}{{end}}",
		},
		{
			name:     "Mixed content",
			input:    "This is a test with {{api_key}} and {{base_url}}",
			expected: "This is a test with {{.api_key}} and {{.base_url}}",
		},
		{
			name:     "JSON with variables",
			input:    "{\"auth\": \"{{api_key}}\", \"token\": \"{{access_token}}\"}",
			expected: "{\"auth\": \"{{.api_key}}\", \"token\": \"{{.access_token}}\"}",
		},
		{
			name:     "Function with variables",
			input:    "{{PropertyMatches headers.Authorization {{access_token}}}}",
			expected: "{{PropertyMatches headers.Authorization {{.access_token}}}}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := convertVariableToGoTemplate(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

// Test scenario variables conversion
func Test_EnhanceScenarioVariables(t *testing.T) {
	// Create a test scenario with postman variable syntax
	scenario := &types.APIScenario{
		Name: "Test Scenario",
		Path: "/api/test",
		Request: types.APIRequest{
			Variables: map[string]string{
				"api_key":      "{{api_key}}",
				"access_token": "{{access_token}}",
				"template_fn":  "{{NthRequest 1}}",
			},
			Headers: map[string]string{
				"Authorization": "Bearer {{access_token}}",
				"X-API-Key":     "{{api_key}}",
			},
			Contents: "{\"token\": \"{{access_token}}\"}",
		},
		Authentication: map[string]types.APIAuthorization{
			"bearer": {
				Format: "{{access_token}}",
			},
		},
		Predicate: "{{if not .access_token}}{{RequestByName \"Get JWT Token\"}}{{else}}{{NthRequest 1}}{{end}}",
	}

	// Apply variable enhancement
	enhanceScenarioVariables(scenario)

	// Check that variables were converted correctly
	require.Equal(t, "{{.api_key}}", scenario.Request.Variables["api_key"])
	require.Equal(t, "{{.access_token}}", scenario.Request.Variables["access_token"])
	require.Equal(t, "{{NthRequest 1}}", scenario.Request.Variables["template_fn"]) // Should not be changed

	// Check headers
	require.Equal(t, "Bearer {{.access_token}}", scenario.Request.Headers["Authorization"])
	require.Equal(t, "{{.api_key}}", scenario.Request.Headers["X-API-Key"])

	// Check contents
	require.Equal(t, "{\"token\": \"{{.access_token}}\"}", scenario.Request.Contents)

	// Check auth
	require.Equal(t, "{{.access_token}}", scenario.Authentication["bearer"].Format)

	// Predicate should be preserved as is (template functions)
	require.Equal(t, "{{if not .access_token}}{{RequestByName \"Get JWT Token\"}}{{else}}{{NthRequest 1}}{{end}}", scenario.Predicate)
}
