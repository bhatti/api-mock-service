package pm

import (
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// postman model adopted from https://github.com/rbretecher/go-postman-collection

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

// PostmanConverter converter
type PostmanConverter struct {
	config    *types.Configuration
	variables map[string]string
	execs     []string
	auth      map[string][]*PostmanAuthParam
	started   time.Time
	ended     time.Time
}

// ConvertPostmanToScenarios builds scenarios from Postman contents
func ConvertPostmanToScenarios(
	config *types.Configuration,
	collection *PostmanCollection,
	started time.Time,
	ended time.Time,
) (scenarios []*types.APIScenario) {
	converter := buildConverter(config, started, ended)
	converter.addVariables(collection.Variables)
	converter.addEvents(collection.Events)
	converter.addAuth(collection.Auth)
	for _, item := range collection.Items {
		scenarios = append(scenarios, converter.itemToScenario(collection.Info.Name, item)...)
	}
	return
}

func (c *PostmanConverter) itemToScenario(
	group string,
	items *PostmanItems,
) (scenarios []*types.APIScenario) {
	c.addVariables(items.Variables)
	c.addEvents(items.Events)
	c.addAuth(items.Auth)
	if items.IsGroup() {
		group = items.Name
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
		scenarios = append(scenarios, c.itemToScenario(group, item)...)
	}
	return
}

func (c *PostmanConverter) toScenario(
	items *PostmanItems,
	res *PostmanResponse,
	group string,
) (scenario *types.APIScenario, err error) {
	if items.Request == nil {
		return nil, fmt.Errorf("no request for %s", items.Name)
	}
	if items.Request.URL == nil {
		return nil, fmt.Errorf("no request url for %s", items.Name)
	}
	raw := replaceTemplateVariables(items.Request.URL.Raw)
	if b, err := fuzz.ParseTemplate("", []byte(raw), c.variables); err == nil {
		raw = string(b)
	} else {
		log.WithFields(log.Fields{
			"Raw":       raw,
			"Error":     err,
			"Variables": c.variables,
		}).Warnf("failed to parse template for url")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

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
	scenario.Name = items.Name + " " + items.ID
	scenario.Description = items.Description
	if group != "" {
		scenario.Group = group + "_" + scenario.Group
		scenario.Tags = append(scenario.Tags, group)
	}
	for k, v := range c.variables {
		if v != "" {
			scenario.Request.Variables[k] = v
		}
	}
	for _, variable := range items.Variables {
		scenario.Request.Variables[variable.Name] = variable.Value
	}
	for k, vals := range c.auth {
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

	return scenario, nil
}

func (c *PostmanConverter) addVariables(variables []*PostmanVariable) {
	for _, next := range variables {
		c.variables[next.KeyName()] = next.Value
	}
}

func (c *PostmanConverter) addEvents(events []*PostmanEvent) {
	for _, event := range events {
		if event.Script == nil || event.Disabled {
			continue
		}
		for _, exec := range event.Script.Exec {
			exec = strings.TrimSpace(exec)
			if exec == "" {
				continue
			}
			found := false
			for _, next := range c.execs {
				if exec == next {
					found = true
					break
				}
			}
			if found {
				continue
			}
			if strings.Contains(exec, "pm.variables.set") ||
				strings.Contains(exec, "pm.variables.get") ||
				strings.Contains(exec, "pm.environment.get") ||
				strings.Contains(exec, "pm.request.headers.add") {
				c.execs = append(c.execs, exec)
			} else if strings.HasPrefix(exec, "/*") ||
				strings.HasPrefix(exec, "*") ||
				strings.HasPrefix(exec, "//") {
				// ignore
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
				// ignore
			} else {
				log.WithFields(log.Fields{
					"Exec":      exec,
					"Variables": c.variables,
				}).Debugf("unknown postman event could not be imported")
			}
		}
	}
}

func (c *PostmanConverter) addAuth(auth *PostmanAuth) {
	if auth != nil && auth.GetParams() != nil {
		c.auth[string(auth.Type)] = auth.GetParams()
	}
}

func (c *PostmanConverter) handleVariableEvents(name string, headers map[string][]string) {
	for _, exec := range c.execs {
		c.handleEvent(name, exec, headers)
	}
}

func (c *PostmanConverter) handleEvent(name string, exec string, headers map[string][]string) {
	if strings.Contains(exec, "pm.variables.set") {
		var re = regexp.MustCompile(`pm.variables.set.[' ]+(\w+)[', ]+(.+)'\)`)
		partsStr := strings.ReplaceAll(re.ReplaceAllString(exec, `$1=$2`), "'", "")
		parts := strings.Split(partsStr, "=")
		if partsStr != exec && len(parts) == 2 {
			c.variables[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	} else if strings.Contains(exec, "pm.request.headers.add") {
		var re = regexp.MustCompile(`pm.request.headers.add..key[:' ]+([^']+)[', ]+value[',: ]+(.+)}.*`)
		partsStr := strings.ReplaceAll(re.ReplaceAllString(exec, `$1=$2`), "'", "")
		parts := strings.Split(partsStr, "=")
		if partsStr != exec && len(parts) == 2 {
			if strings.Contains(parts[1], "pm.variables.get") {
				re = regexp.MustCompile(`.*pm.variables.get\((.+)\).*`)
				varName := strings.TrimSpace(strings.ReplaceAll(re.ReplaceAllString(parts[1], `$1`), "'", ""))
				if c.variables[varName] == "" {
					log.WithFields(log.Fields{
						"Exec":            exec,
						"Variables":       c.variables,
						"PostmanVariable": varName,
					}).Warnf("unknown variable %s in postman event", varName)
				} else {
					re = regexp.MustCompile(`[+ ]*pm.variables.get\((.+)\)`)
					parts[1] = re.ReplaceAllString(parts[1], c.variables[varName])
				}
				re = regexp.MustCompile(`[ +]*pm.info.requestName`)
				parts[1] = re.ReplaceAllString(parts[1], name)
				headers[strings.TrimSpace(parts[0])] = []string{strings.TrimSpace(parts[1])}
			} else if strings.Contains(parts[1], "pm.environment.get") {
				re = regexp.MustCompile(`.*pm.environment.get\((.+)\).*`)
				varName := strings.TrimSpace(strings.ReplaceAll(re.ReplaceAllString(parts[1], `$1`), "'", ""))
				if os.Getenv(varName) == "" {
					log.WithFields(log.Fields{
						"Exec":        exec,
						"EnvVariable": varName,
					}).Warnf("unknown env variable %s in postman event", varName)
				} else {
					re = regexp.MustCompile(`[+ ]*pm.environment.get\((.+)\)`)
					parts[1] = re.ReplaceAllString(parts[1], os.Getenv(varName))
				}
				re = regexp.MustCompile(`[ +]*pm.info.requestName`)
				parts[1] = re.ReplaceAllString(parts[1], name)
				headers[strings.TrimSpace(parts[0])] = []string{strings.TrimSpace(parts[1])}
			} else {
				headers[strings.TrimSpace(parts[0])] = []string{strings.TrimSpace(parts[1])}
			}
		}
	}
}

func replaceTemplateVariables(s string) string {
	var re = regexp.MustCompile(`{{(\w+)}}`)
	return re.ReplaceAllString(s, `{{.$1}}`)
}

func buildConverter(config *types.Configuration, started time.Time, ended time.Time) *PostmanConverter {
	converter := &PostmanConverter{
		config:    config,
		variables: make(map[string]string),
		execs:     make([]string, 0),
		auth:      make(map[string][]*PostmanAuthParam),
		started:   started,
		ended:     ended,
	}
	return converter
}
