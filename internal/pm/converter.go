package pm

import (
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	log "github.com/sirupsen/logrus"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// postman model adopted from https://github.com/rbretecher/go-postman-collection

// PostmanConverter converter
type PostmanConverter struct {
	config  *types.Configuration
	context *PostmanContext
	started time.Time
	ended   time.Time
}

// NewPostmanConverter creates a new converter instance
func NewPostmanConverter(
	config *types.Configuration,
	started time.Time,
	ended time.Time,
) *PostmanConverter {
	return &PostmanConverter{
		config:  config,
		context: NewPostmanContext(),
		started: started,
		ended:   ended,
	}
}

// ConvertScenariosToPostman builds Postman contents
func ConvertScenariosToPostman(
	name string,
	scenarios ...*types.APIScenario,
) (c *PostmanCollection) {
	c = CreateCollection(name, name)
	scenariosByGroup := make(map[string][]*types.APIScenario)
	for _, scenario := range scenarios {
		list := append(scenariosByGroup[scenario.Group], scenario)
		scenariosByGroup[scenario.Group] = list
	}
	for groupName, groupScenarios := range scenariosByGroup {
		group := c.AddItemGroup(groupName)
		for _, scenario := range groupScenarios {
			if item, err := toItem(scenario); err == nil {
				group.AddItem(item)
			} else {
				log.WithFields(log.Fields{
					"Group": groupName,
					"Error": err,
				}).Warnf("failed to convert scenario to item")
			}
		}
	}
	return
}

func toItem(scenario *types.APIScenario) (*PostmanItems, error) {
	u, err := scenario.GetURL("http://0.0.0.0")
	if err != nil {
		return nil, err
	}
	return CreatePostmanItem(PostmanItem{
		Name:        scenario.Name,
		Description: scenario.Description,
		Request:     toRequest(string(scenario.Method), u, scenario.Request),
		Responses:   []*PostmanResponse{toResponse(scenario.Response)},
	}), nil
}

func toResponse(res types.APIResponse) *PostmanResponse {
	return &PostmanResponse{
		Headers: &PostmanHeaderList{Headers: buildHeadersArray(res.Headers)},
		Cookies: make([]*PostmanCookie, 0),
		Body:    res.Contents,
		Status:  "",
		Code:    res.StatusCode,
		Name:    "",
	}
}

func toRequest(method string, u *url.URL, req types.APIRequest) *PostmanRequest {
	return &PostmanRequest{
		URL:    buildURL(u),
		Method: types.MethodType(method),
		Auth:   nil, // TODO auth
		Header: buildHeaders(req.Headers),
		Body: &PostmanRequestBody{
			Raw:      req.Contents,
			FormData: req.PostParams,
		},
	}
}

// ConvertPostmanToScenarios builds scenarios from Postman contents
func ConvertPostmanToScenarios(
	config *types.Configuration,
	collection *PostmanCollection,
	started time.Time,
	ended time.Time,
) (scenarios []*types.APIScenario, apiVariables *types.APIVariables) {
	converter := NewPostmanConverter(config, started, ended)
	converter.addVariables(collection.Variables)
	converter.addEvents(collection.Events)
	converter.addAuth(collection.Auth)
	apiVariables = &types.APIVariables{
		Name:      collection.Info.Name,
		Variables: make(map[string]string),
	}

	for _, item := range collection.Items {
		if apiVariables.Name == "" {
			apiVariables.Name = item.Name
		}
		scenarios = append(scenarios, converter.itemToScenarios(collection.Info.Name, item)...)
	}
	for _, v := range collection.Variables {
		if !v.Disabled && v.Name != "" {
			apiVariables.Variables[v.Name] = v.Value
		}
	}
	for _, scenario := range scenarios {
		scenario.VariablesFile = apiVariables.Name
		for k, v := range scenario.Request.Variables {
			if strings.Contains(v, "{{") {
				apiVariables.Variables[k] = ""
			}
		}
	}
	return
}

func (c *PostmanConverter) itemToScenarios(group string, items *PostmanItems) (scenarios []*types.APIScenario) {
	c.addVariables(items.Variables)
	c.addEvents(items.Events)
	c.addAuth(items.Auth)
	if items.IsGroup() {
		if group != "" {
			group = group + "_" + items.Name
		} else {
			group = items.Name
		}
	} else if len(items.Responses) > 0 {
		for _, res := range items.Responses {
			scenario, err := c.toScenario(items, res, group)
			if err == nil {
				scenarios = append(scenarios, scenario)
			} else {
				log.WithFields(log.Fields{
					"entry":           items,
					"PostmanResponse": res,
					"Error":           err,
				}).Warnf("failed to convert item to scenario")
			}
		}
	} else {
		scenario, err := c.toScenario(items, &PostmanResponse{Headers: &PostmanHeaderList{}}, group)
		if err == nil {
			scenarios = append(scenarios, scenario)
		} else {
			log.WithFields(log.Fields{
				"entry": items,
				"Error": err,
			}).Warnf("failed to convert item to scenario without response")
		}
	}
	for _, item := range items.Items {
		scenarios = append(scenarios, c.itemToScenarios(group, item)...)
	}
	return
}

func (c *PostmanConverter) toScenario(items *PostmanItems, res *PostmanResponse, group string,
) (scenario *types.APIScenario, err error) {
	if items.Request == nil {
		return nil, fmt.Errorf("no request for %s", items.Name)
	}
	if items.Request.URL == nil {
		return nil, fmt.Errorf("no request url for %s", items.Name)
	}

	// Initialize variables first
	c.addVariables(items.Variables)
	c.addEvents(items.Events)

	u := parseRawUrl(items)
	//raw := replaceTemplateVariables(items.Request.URL.Raw)
	//if b, err := fuzz.ParseTemplate("", []byte(raw), c.context.CollectionVars); err == nil {
	//	raw = string(b)
	//} else {
	//	log.WithFields(log.Fields{
	//		"Raw":       raw,
	//		"Error":     err,
	//		"Variables": c.context.CollectionVars,
	//	}).Warnf("failed to parse template for url")
	//}
	//u, err := url.Parse(raw)
	//if err != nil {
	//	return nil, err
	//}

	headers := items.Request.headersMap()
	c.handleVariableEvents(items.Name, headers)

	scenario, err = types.BuildScenarioFromHTTP(
		c.config,
		"pm-",
		u,
		string(items.Request.Method),
		"", // group
		"", // req-http-version
		"", // res-http-version
		items.Request.bodyText(),
		res.bodyText(),
		u.Query(),
		items.Request.formParams(),
		headers,
		items.Request.contentType(),
		res.headersMap(),
		res.contentType(),
		res.Code,
		c.started,
		c.ended,
	)
	if err != nil {
		return nil, err
	}
	scenario.Name = strings.TrimSpace(items.Name + " " + items.ID)
	scenario.Description = items.Description
	if group != "" {
		scenario.Group = group + "_" + scenario.Group
		scenario.Tags = append(scenario.Tags, group)
	}

	// Initialize variables map
	if scenario.Request.Variables == nil {
		scenario.Request.Variables = make(map[string]string)
	}

	// Add default variables before enhancement
	scenario.Request.Variables["scenario_name"] = scenario.Name
	scenario.Request.Variables["scenario_group"] = scenario.Group
	if len(scenario.Tags) > 0 {
		scenario.Request.Variables["scenario_tag"] = scenario.Tags[0]
	}

	c.enhanceScenario(scenario, items)
	enhanceScenarioVariables(scenario)

	return scenario, nil
}

func (c *PostmanConverter) enhanceScenario(scenario *types.APIScenario, items *PostmanItems) {
	// Add URL variables from raw URL
	if items.Request != nil && items.Request.URL != nil {
		raw := items.Request.URL.Raw
		varPattern := regexp.MustCompile(`{{([^}]+)}}`)
		matches := varPattern.FindAllStringSubmatch(raw, -1)
		for _, match := range matches {
			if len(match) > 1 {
				varName := strings.TrimSpace(match[1])
				scenario.Request.Variables[varName] = "{{." + varName + "}}"
			}
		}
	}

	// Add variables from all scopes
	for k, v := range c.context.CollectionVars {
		if v != "" {
			scenario.Request.Variables[k] = v
		}
	}

	for k, v := range c.context.Environment {
		if v != "" {
			scenario.Request.Variables[k] = v
		}
	}

	for _, variable := range items.Variables {
		scenario.Request.Variables[variable.Name] = variable.Value
	}

	// Add URL query params as variables if they exist
	if len(items.Request.URL.Query) > 0 {
		for _, q := range items.Request.URL.Query {
			if q.Value != "" {
				scenario.Request.Variables[q.Key] = q.Value // should we use query prefix
			}
		}
	}

	// Add required base variables if not present
	requiredVars := []string{"base_url", "api_key"}
	for _, varName := range requiredVars {
		if _, exists := scenario.Request.Variables[varName]; !exists {
			scenario.Request.Variables[varName] = "{{." + varName + "}}"
		}
	}

	// This prevents scenarios from referencing tokens they're supposed to create
	//for _, setVar := range scenario.Response.AddSharedVariables {
	//	if setVar == "access_token" {
	//		// If this scenario sets access_token in response, remove it from request
	//		delete(scenario.Request.Variables, "access_token")
	//		break
	//	}
	//}

	// Convert scripts using ScriptConverter
	scriptConverter := NewScriptConverter(c.context)

	// Add pre-request scripts as assertions
	preScripts := c.context.GetScripts("prerequest")
	for _, script := range preScripts {
		scenario.Request.Description += utils.ToYAMLComment(fmt.Sprintf("PreRequest:%s", script))
	}

	if len(preScripts) > 0 {
		combinedScript := strings.Join(preScripts, "\n")
		scriptConverter.ConvertPreRequestScript(combinedScript, scenario)
	}

	// Add test scripts as response assertions
	testScripts := c.context.GetScripts("test")
	for _, script := range testScripts {
		scenario.Response.Description += utils.ToYAMLComment(fmt.Sprintf("Test:%s", script))
	}
	if len(testScripts) > 0 {
		combinedScript := strings.Join(testScripts, "\n")
		scriptConverter.ConvertTestScript(combinedScript, scenario)
	}

	// Add settings
	if c.context.Settings.TimeoutMS > 0 {
		// scenario.WaitBeforeReply = time.Duration(c.context.Settings.TimeoutMS) * time.Millisecond
	}

	// Handle authentication
	if items.Request.Auth != nil {
		switch items.Request.Auth.Type {
		case "bearer":
			if len(items.Request.Auth.Bearer) > 0 {
				for _, param := range items.Request.Auth.Bearer {
					if param.Key == "token" {
						scenario.Authentication["bearer"] = types.APIAuthorization{
							Type:   "bearer",
							Name:   "Authorization",
							In:     "header",
							Scheme: "bearer",
							Format: fmt.Sprintf("%v", param.Value),
						}
						break
					}
				}
			}
		case "apiKey":
			// Handle API key auth
			if len(items.Request.Auth.APIKey) > 0 {
				for _, param := range items.Request.Auth.APIKey {
					scenario.Authentication["apiKey"] = types.APIAuthorization{
						Type:   "apiKey",
						Name:   param.Key,
						In:     "header", //param.In,
						Format: fmt.Sprintf("%v", param.Value),
					}
					break
				}
			}
		case "basic":
			// Handle basic auth
			scenario.Authentication["basic"] = types.APIAuthorization{
				Type:   "basic",
				Name:   "Authorization",
				In:     "header",
				Scheme: "basic",
			}
		}
	}

	for k, vals := range c.context.Auth {
		for _, v := range vals {
			scenario.Authentication[k] = types.APIAuthorization{
				Type:   k,
				Name:   "Authorization",
				Scheme: v.Key,
				In:     "header",
				Format: fmt.Sprintf("%v", v.Value),
			}
		}
	}
}

// enhanceScenarioVariables processes all variables in a scenario to use Go template syntax
func enhanceScenarioVariables(scenario *types.APIScenario) {
	if strings.Contains(scenario.Path, "{{") && !strings.Contains(scenario.Path, "{{.") {
		scenario.Path = convertVariableToGoTemplate(scenario.Path)
	}
	// Process request variables
	for k, v := range scenario.Request.Variables {
		if v != "" && strings.Contains(v, "{{") && !strings.Contains(v, "{{.") {
			scenario.Request.Variables[k] = convertVariableToGoTemplate(v)
		}
	}

	// Process authentication format strings
	for authType, auth := range scenario.Authentication {
		if auth.Format != "" && strings.Contains(auth.Format, "{{") && !strings.Contains(auth.Format, "{{.") {
			auth.Format = convertVariableToGoTemplate(auth.Format)
			scenario.Authentication[authType] = auth
		}
	}

	// Process request headers
	for k, v := range scenario.Request.Headers {
		if strings.Contains(v, "{{") && !strings.Contains(v, "{{.") {
			scenario.Request.Headers[k] = convertVariableToGoTemplate(v)
		}
	}

	// Process request contents if it contains variables
	if strings.Contains(scenario.Request.Contents, "{{") && !strings.Contains(scenario.Request.Contents, "{{.") {
		scenario.Request.Contents = convertVariableToGoTemplate(scenario.Request.Contents)
	}

	// Also process example contents
	if strings.Contains(scenario.Request.ExampleContents, "{{") && !strings.Contains(scenario.Request.ExampleContents, "{{.") {
		scenario.Request.ExampleContents = convertVariableToGoTemplate(scenario.Request.ExampleContents)
	}
}

func (c *PostmanConverter) addVariables(variables []*PostmanVariable) {
	for _, next := range variables {
		if next != nil && next.KeyName() != "" {
			c.context.CollectionVars[next.KeyName()] = next.Value
		}
	}
}

func (c *PostmanConverter) addEvents(events []*PostmanEvent) {
	scriptConverter := NewScriptConverter(c.context)

	for _, event := range events {
		if event.Script == nil || event.Disabled {
			continue
		}
		scriptType := string(event.Listen)
		for _, exec := range event.Script.Exec {
			// Extract environment variables from script
			envVarPattern := regexp.MustCompile(`pm\.environment\.get\(['"](\w+)['"]\)`)
			matches := envVarPattern.FindAllStringSubmatch(exec, -1)
			for _, match := range matches {
				if len(match) > 1 {
					varName := match[1]
					c.context.Environment[varName] = "{{." + varName + "}}"
				}
			}
			// Store in Scripts map
			c.context.AddScript(event.Script.Name, scriptType, exec)

			// Process script commands
			scriptConverter.ProcessScript(exec, "", nil)
		}
	}
}

func (c *PostmanConverter) handleVariableEvents(name string, headers map[string][]string) {
	scriptConverter := NewScriptConverter(c.context)
	// Process all stored scripts that might affect headers
	for _, exec := range c.context.ScriptExecs {
		scriptConverter.ProcessScript(exec, name, headers)
	}
}

func (c *PostmanConverter) addAuth(auth *PostmanAuth) {
	if auth != nil && auth.GetParams() != nil {
		c.context.Auth[string(auth.Type)] = auth.GetParams()
	}
}

func replaceTemplateVariables(s string) string {
	var re = regexp.MustCompile(`{{(\w+)}}`)
	return re.ReplaceAllString(s, `{{.$1}}`)
}

// parseRawUrl processes a Postman URL and correctly extracts just the path component
func parseRawUrl(items *PostmanItems) *url.URL {
	var u *url.URL

	// Get the raw URL string
	rawURL := ""
	if items.Request.URL.Raw != "" {
		rawURL = items.Request.URL.Raw
	} else if len(items.Request.URL.Path) > 0 {
		// Join path components if Raw is not available
		pathParts := make([]string, len(items.Request.URL.Path))
		for i, part := range items.Request.URL.Path {
			pathParts[i] = part
		}
		rawURL = "/" + strings.Join(pathParts, "/")
	} else {
		// Default to root path if no URL information is available
		return &url.URL{Path: "/"}
	}

	// Check if there are query parameters in the raw URL
	var queryString string
	if idx := strings.Index(rawURL, "?"); idx >= 0 {
		queryString = rawURL[idx+1:] // Save the query string
	}

	// Extract the path component after the base_url variable
	path := extractPathAfterBaseUrl(rawURL)

	// Create a URL with just the path component
	u = &url.URL{Path: path}

	// Extract path variables for scenario variables
	pathVars := make(map[string]string)
	varPattern := regexp.MustCompile(`{{([^}]+)}}`)
	matches := varPattern.FindAllStringSubmatch(path, -1)
	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			pathVars[varName] = "{{." + varName + "}}"
		}
	}

	// Add query parameters if present in the URL.Query field
	if items.Request.URL.Query != nil && len(items.Request.URL.Query) > 0 {
		query := url.Values{}
		for _, q := range items.Request.URL.Query {
			query.Add(q.Key, q.Value)
		}
		u.RawQuery = query.Encode()
	} else if queryString != "" {
		// If no Query field but we extracted a query string from the raw URL
		u.RawQuery = queryString
	}

	return u
}

// extractPathAfterBaseUrl extracts the path component after any {{base_url}} variable
func extractPathAfterBaseUrl(rawURL string) string {
	// Return "/" for empty inputs
	if rawURL == "" {
		return "/"
	}

	// Pattern to match {{base_url}} with optional protocol and domain parts
	baseUrlPattern := regexp.MustCompile(`{{base_url}}(?:/+)?`)

	// If base_url is in the raw URL, extract just the path component
	if baseUrlPattern.MatchString(rawURL) {
		// Split by {{base_url}} and get everything after it
		parts := baseUrlPattern.Split(rawURL, 2)
		if len(parts) > 1 {
			path := parts[1]

			// Remove any query parameters
			if idx := strings.Index(path, "?"); idx >= 0 {
				path = path[:idx]
			}

			// Ensure the path starts with a slash
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			return path
		}
	}

	// If we can't find base_url or can't extract the path, try to parse as a URL
	if parsedURL, err := url.Parse(rawURL); err == nil {
		path := parsedURL.Path

		// Ensure the path starts with a slash even after parsing
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		// Remove any query parameters (should be handled by url.Parse already)
		if idx := strings.Index(path, "?"); idx >= 0 {
			path = path[:idx]
		}

		return path
	}

	// Fallback: return the raw string with a leading slash if needed
	if !strings.HasPrefix(rawURL, "/") {
		rawURL = "/" + rawURL
	}

	// Remove any query parameters
	if idx := strings.Index(rawURL, "?"); idx >= 0 {
		rawURL = rawURL[:idx]
	}

	return rawURL
}

// ProcessURLPath processes and normalizes URL paths
// This is a simpler version of the function you already have
func ProcessURLPath(raw string) string {
	if raw == "" {
		return "/"
	}

	// Ensure leading slash
	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}

	// Remove trailing slash if present
	raw = strings.TrimRight(raw, "/")

	return raw
}
