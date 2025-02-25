package pm

import (
	"github.com/bhatti/api-mock-service/internal/types"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ScriptConverter converts Postman scripts to API scenario format
type ScriptConverter struct {
	context *PostmanContext

	variableSetRegex    *regexp.Regexp
	variableGetRegex    *regexp.Regexp
	environmentGetRegex *regexp.Regexp
	headerAddRegex      *regexp.Regexp

	conditionalPattern  *regexp.Regexp
	dateTimePattern     *regexp.Regexp
	responseTimePattern *regexp.Regexp
	statusCodePattern   *regexp.Regexp
	contentTypePattern  *regexp.Regexp
	setEnvPattern       *regexp.Regexp
}

// NewScriptConverter creates a new script converter
// that's backward compatible with ScriptProcessor
func NewScriptConverter(context *PostmanContext) *ScriptConverter {
	return &ScriptConverter{
		context: context,

		variableSetRegex:    regexp.MustCompile(`pm\.variables\.set\(['"](\w+)['"],\s*['"]([^'"]+)['"]\)`),
		headerAddRegex:      regexp.MustCompile(`pm\.request\.headers\.add\({key:\s*['"]([^'"]+)['"],\s*value:\s*(.+?)}\)`),
		variableGetRegex:    regexp.MustCompile(`pm\.variables\.get\(['"](\w+)['"]\)`),
		environmentGetRegex: regexp.MustCompile(`pm\.environment\.get\(['"](\w+)['"]\)`),

		conditionalPattern:  regexp.MustCompile(`if\s*\(\s*!\s*pm\.environment\.get\(['"](\w+)['"]\)\s*\)\s*\{[^}]*setNextRequest\(['"]([^'"]+)['"]\)`),
		dateTimePattern:     regexp.MustCompile(`const\s+(\w+)\s*=\s*new\s+Date\(\)\.toISOString\(\)`),
		responseTimePattern: regexp.MustCompile(`pm\.expect\s*\(\s*pm\.response\.responseTime\s*\)\.to\.be\.below\s*\(\s*(\d+)\s*\)`),
		statusCodePattern:   regexp.MustCompile(`pm\.response\.to\.have\.status\s*\(\s*(\d+)\s*\)`),
		contentTypePattern:  regexp.MustCompile(`pm\.expect\s*\(\s*pm\.response\.headers\.get\s*\(['"]Content-Type['"]\)\s*\)\.to\.include\s*\(['"]([^'"]+)['"]\)`),
		setEnvPattern:       regexp.MustCompile(`pm\.environment\.set\s*\(['"](\w+)['"]\s*,\s*([^)]+)\)`),
	}
}

// ProcessScript processes a script command
func (c *ScriptConverter) ProcessScript(exec string, name string, headers http.Header) {
	exec = strings.TrimSpace(exec)
	if exec == "" {
		return
	}

	// Check if this script has already been processed
	found := false
	for _, next := range c.context.ScriptExecs {
		if exec == next {
			found = true
			break
		}
	}
	if found {
		return
	}

	// New check: Skip script with setNextRequest that references the current scenario
	if name != "" && strings.Contains(exec, "setNextRequest") {
		// Extract the request name from setNextRequest
		setNextRequestPattern := regexp.MustCompile(`setNextRequest\(['"]([^'"]+)['"]\)`)
		match := setNextRequestPattern.FindStringSubmatch(exec)
		if len(match) >= 2 {
			nextRequestName := match[1]
			// If it references the current scenario, skip this script
			if nextRequestName == name {
				log.WithFields(log.Fields{
					"ScenarioName": name,
					"Script":       exec,
				}).Debugf("skipping setNextRequest script that references current scenario")
				return
			}
		}
	}

	if strings.Contains(exec, "pm.variables.set") ||
		strings.Contains(exec, "pm.variables.get") ||
		strings.Contains(exec, "pm.environment.get") ||
		strings.Contains(exec, "pm.request.headers.add") {
		c.handleEvent(name, exec, headers)
		// Add to script execs only if it's a supported type
		c.context.ScriptExecs = append(c.context.ScriptExecs, exec)
	} else if strings.HasPrefix(exec, "/*") ||
		strings.HasPrefix(exec, "*") ||
		strings.HasPrefix(exec, "//") {
		// ignore comments
	} else if strings.Contains(exec, "pm.collectionVariables.unset") ||
		strings.Contains(exec, "pm.response.json") ||
		strings.Contains(exec, "pm.response.code") ||
		strings.Contains(exec, "console.") ||
		strings.Contains(exec, "if (") ||
		strings.Contains(exec, "}") ||
		strings.Contains(exec, "pm.request.url") ||
		strings.Contains(exec, "const ") ||
		strings.Contains(exec, "let ") ||
		strings.Contains(exec, "return") ||
		strings.Contains(exec, "pm.collectionVariables.set") {
		// ignore unsupported commands
	} else {
		log.WithFields(log.Fields{
			"Exec":      exec,
			"Variables": c.context.CollectionVars,
		}).Debugf("unknown postman event could not be imported")
	}
}

// handleEvent processes script events
func (c *ScriptConverter) handleEvent(name string, exec string, headers http.Header) {
	if strings.Contains(exec, "pm.variables.set") {
		matches := c.variableSetRegex.FindStringSubmatch(exec)
		if len(matches) == 3 {
			varName := strings.TrimSpace(matches[1])
			varValue := strings.TrimSpace(matches[2])
			c.context.CollectionVars[varName] = varValue
		}
	} else if strings.Contains(exec, "pm.request.headers.add") {
		matches := c.headerAddRegex.FindStringSubmatch(exec)
		if len(matches) == 3 {
			headerName := strings.TrimSpace(matches[1])
			headerValue := c.processHeaderValue(matches[2], name)
			if headerValue != "" && headers != nil {
				headers.Set(headerName, headerValue)
			}
		}
	}
}

// processHeaderValue processes a header value
func (c *ScriptConverter) processHeaderValue(value string, requestName string) string {
	// Remove surrounding quotes if present
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)

	// Process variables
	if strings.Contains(value, "pm.variables.get") {
		matches := c.variableGetRegex.FindStringSubmatch(value)
		if len(matches) == 2 {
			varName := strings.TrimSpace(matches[1])
			if varValue, exists := c.context.CollectionVars[varName]; exists {
				re := regexp.MustCompile(`[+ ]*pm\.variables\.get\(['"]` + varName + `['"]\)`)
				value = re.ReplaceAllString(value, varValue)
			} else {
				log.WithFields(log.Fields{
					"Variables":       c.context.CollectionVars,
					"PostmanVariable": varName,
				}).Warnf("unknown variable %s in postman event", varName)
				re := regexp.MustCompile(`[+ ]*pm\.variables\.get\(['"]` + varName + `['"]\)`)
				value = re.ReplaceAllString(value, "")
			}
		}
	}

	// Process environment variables
	if strings.Contains(value, "pm.environment.get") {
		matches := c.environmentGetRegex.FindStringSubmatch(value)
		if len(matches) == 2 {
			varName := strings.TrimSpace(matches[1])
			if envValue := os.Getenv(varName); envValue != "" {
				re := regexp.MustCompile(`[+ ]*pm\.environment\.get\(['"]` + varName + `['"]\)`)
				value = re.ReplaceAllString(value, envValue)
			} else {
				log.WithFields(log.Fields{
					"EnvVariable": varName,
				}).Warnf("unknown env variable %s in postman event", varName)
				re := regexp.MustCompile(`[+ ]*pm\.environment\.get\(['"]` + varName + `['"]\)`)
				value = re.ReplaceAllString(value, "")
			}
		}
	}

	// Replace request name
	if strings.Contains(value, "pm.info.requestName") {
		re := regexp.MustCompile(`[ +]*pm\.info\.requestName`)
		value = re.ReplaceAllString(value, requestName)
	}

	// Clean up any remaining artifacts
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " + ", "")
	value = strings.ReplaceAll(value, "+", "")
	value = strings.Trim(value, `"'`)

	return value
}

// ConvertPreRequestScript converts a pre-request script to API scenario format
func (c *ScriptConverter) ConvertPreRequestScript(script string, scenario *types.APIScenario) {
	// Check for conditional next request
	conditionalMatches := c.conditionalPattern.FindStringSubmatch(script)

	if len(conditionalMatches) >= 3 {
		varName := conditionalMatches[1]
		nextRequest := strings.TrimSpace(conditionalMatches[2])
		// Skip adding predicate if it would reference itself
		if nextRequest != scenario.Name {
			// Set next request to load first
			scenario.NextRequest = "{{if not ." + varName + "}}" + nextRequest + "{{end}}"
		} else {
			log.WithFields(log.Fields{
				"ScenarioName": scenario.Name,
				"NextRequest":  nextRequest,
			}).Debugf("skipping predicate that would reference itself")
		}
	}

	// Check for datetime variables
	dateTimeMatches := c.dateTimePattern.FindAllStringSubmatch(script, -1)
	for _, match := range dateTimeMatches {
		if len(match) >= 2 {
			varName := match[1]
			// Add ISODatetime function
			scenario.Request.Variables[varName] = "{{ISODatetime}}"
		}
	}

	// Extract any environment variable references
	envMatches := c.environmentGetRegex.FindAllStringSubmatch(script, -1)
	for _, match := range envMatches {
		if len(match) >= 2 {
			varName := match[1]
			// Add to variables if not already present
			if _, exists := scenario.Request.Variables[varName]; !exists {
				scenario.Request.Variables[varName] = "{{." + varName + "}}"
			}
		}
	}
}

// ConvertTestScript converts a test script to API scenario format
func (c *ScriptConverter) ConvertTestScript(script string, scenario *types.APIScenario) {
	// Check for response time assertions
	responseTimeMatches := c.responseTimePattern.FindStringSubmatch(script)
	if len(responseTimeMatches) >= 2 {
		// Add ResponseTimeMillisLE assertion
		scenario.Response.Assertions = updateOrAppendAssertion(
			scenario.Response.Assertions,
			"ResponseTimeMillisLE "+responseTimeMatches[1],
			"ResponseTimeMillisLE \\d+",
		)
	}

	// Check for status code assertions
	statusCodeMatches := c.statusCodePattern.FindStringSubmatch(script)
	if len(statusCodeMatches) >= 2 {
		statusCode, err := strconv.Atoi(statusCodeMatches[1])
		if err == nil {
			// Set status code
			scenario.Response.StatusCode = statusCode
			// Add ResponseStatusMatches assertion
			scenario.Response.Assertions = updateOrAppendAssertion(
				scenario.Response.Assertions,
				"ResponseStatusMatches "+statusCodeMatches[1],
				"ResponseStatusMatches \\d+",
			)
		}
	}

	// Check for content type assertions
	contentTypeMatches := c.contentTypePattern.FindStringSubmatch(script)
	if len(contentTypeMatches) >= 2 {
		contentType := contentTypeMatches[1]
		// Add VariableMatches assertion for Content-Type
		assertion := "VariableMatches headers.Content-Type " + contentType
		scenario.Response.Assertions = updateOrAppendAssertion(
			scenario.Response.Assertions,
			assertion,
			"VariableMatches headers.Content-Type .*",
		)

		// Also add to assert headers pattern
		scenario.Response.AssertHeadersPattern["Content-Type"] = contentType
	}

	// Check for environment variable setting
	setEnvMatches := c.setEnvPattern.FindAllStringSubmatch(script, -1)
	for _, match := range setEnvMatches {
		if len(match) >= 3 {
			varName := match[1]
			varSource := match[2]

			// Add to shared variables
			scenario.Response.AddSharedVariables = appendIfNotExists(scenario.Response.AddSharedVariables, varName)

			// If it's from response body, add to variables
			if strings.Contains(varSource, "responseBody") {
				// Extract the property path
				parts := strings.Split(varSource, ".")
				if len(parts) > 1 {
					propertyPath := strings.Join(parts[1:], ".")
					// Add to variables
					scenario.Request.Variables[varName] = "{{." + propertyPath + "}}"
				} else {
					// Just add the variable
					scenario.Request.Variables[varName] = "{{." + varName + "}}"
				}
				// This prevents scenarios from referencing their own output variables
				delete(scenario.Request.Variables, varName)
			}
		}
	}
}

// Helper function to update or append an assertion
func updateOrAppendAssertion(assertions []string, newAssertion, pattern string) []string {
	// Create regex from pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		// If regex compilation fails, just append
		return append(assertions, newAssertion)
	}

	// Check if there's an existing assertion that matches the pattern
	for i, assertion := range assertions {
		if re.MatchString(assertion) {
			// Update existing assertion
			assertions[i] = newAssertion
			return assertions
		}
	}

	// No match found, append new assertion
	return append(assertions, newAssertion)
}

// Helper function to append to a slice if not already exists
func appendIfNotExists(slice []string, value string) []string {
	for _, item := range slice {
		if item == value {
			return slice
		}
	}
	return append(slice, value)
}

// convertVariableToGoTemplate converts Postman variable syntax to Go template syntax
// It handles these cases:
// 1. {{variable}} -> {{.variable}} for simple variables
// 2. Preserves {{function ...}} format for Go template functions
// 3. Handles nested variables properly
func convertVariableToGoTemplate(input string) string {
	if input == "" {
		return input
	}

	// Special case: if the input already contains Go template functions/syntax, process carefully
	if containsTemplateFunction(input) {
		return processTemplateWithFunctions(input)
	}

	// For simple cases without template functions, just convert all variables
	re := regexp.MustCompile(`{{([^.}][^}]*)}}`)
	result := re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract the variable name (remove the {{ and }})
		varName := match[2 : len(match)-2]
		// Add the dot prefix for variables
		return "{{." + varName + "}}"
	})

	return result
}

// processTemplateWithFunctions handles complex templates containing functions
func processTemplateWithFunctions(input string) string {
	// First identify all template tags
	tagPattern := regexp.MustCompile(`{{([^{}]+)}}`)

	// Replace only the simple variable tags, not function tags
	result := tagPattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract tag content
		content := match[2 : len(match)-2]

		// Skip if it already has a dot prefix
		if strings.HasPrefix(content, ".") {
			return match
		}

		// Skip if it's a template keyword or function
		if isTemplateKeywordOrFunction(content) {
			return match
		}

		// It's a simple variable, add the dot
		return "{{." + content + "}}"
	})

	// Handle nested variables inside complex expressions
	// Find variables inside functions like {{VariableMatches x {{var}}}}
	nestedVarPattern := regexp.MustCompile(`{{([^{}]+){{([^.{}][^{}]*)}}([^{}]*)}}`)
	for nestedVarPattern.MatchString(result) {
		result = nestedVarPattern.ReplaceAllStringFunc(result, func(match string) string {
			return nestedVarPattern.ReplaceAllString(match, "{{$1{{.$2}}$3}}")
		})
	}

	return result
}

// isTemplateKeywordOrFunction checks if a string is a template keyword or function
func isTemplateKeywordOrFunction(content string) bool {
	// Trim spaces
	trimmed := strings.TrimSpace(content)

	// Template keywords
	keywords := []string{"if", "else", "end", "range", "with", "define", "template", "block"}

	// Check if it starts with a keyword
	for _, keyword := range keywords {
		if strings.HasPrefix(trimmed, keyword+" ") || trimmed == keyword {
			return true
		}
	}

	// Check if it's likely a function (starts with uppercase or contains function syntax)
	if len(trimmed) > 0 && (trimmed[0] >= 'A' && trimmed[0] <= 'Z') {
		return true
	}

	// Look for function call syntax with quotes or parentheses
	if strings.Contains(trimmed, "(") || strings.Contains(trimmed, "\"") || strings.Contains(trimmed, "'") {
		return true
	}

	// Specific known functions
	knownFunctions := []string{
		"NthRequest",
		"RequestByName",
		"ISODatetime",
		"VariableMatches",
	}

	for _, funcName := range knownFunctions {
		if strings.HasPrefix(trimmed, funcName) {
			return true
		}
	}

	return false
}

// containsTemplateFunction checks if a string contains any Go template functions
func containsTemplateFunction(input string) bool {
	// Template keywords
	keywords := []string{"if", "else", "end", "range", "with", "define", "template", "block"}

	// Look for template keywords
	for _, keyword := range keywords {
		if strings.Contains(input, "{{"+keyword+" ") || strings.Contains(input, "{{ "+keyword+" ") {
			return true
		}
	}

	// Known functions
	functionPatterns := []string{
		"{{NthRequest",
		"{{RequestByName",
		"{{ISODatetime",
		"{{VariableMatches",
	}

	for _, pattern := range functionPatterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}

	return false
}
