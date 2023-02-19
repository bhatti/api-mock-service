package oapi

import (
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
	"regexp"
	"strings"
)

// MarshalScenarioToOpenAPI converts open-api specs into json
func MarshalScenarioToOpenAPI(title string, version string, scenarios ...*types.MockScenario) ([]byte, error) {
	t := ScenarioToOpenAPI(title, version, scenarios...)
	return t.MarshalJSON()
}

// ScenarioToOpenAPI convert scenarios to open-api specs
func ScenarioToOpenAPI(title string, version string, scenarios ...*types.MockScenario) *openapi3.T {
	root := &openapi3.T{
		OpenAPI: "3.0.2",
		Info:    &openapi3.Info{Title: title, Version: version},
		Components: openapi3.Components{
			SecuritySchemes: make(openapi3.SecuritySchemes),
			Schemas:         make(openapi3.Schemas),
		},
		Paths: make(openapi3.Paths),
		Servers: openapi3.Servers{
			&openapi3.Server{
				URL: "http://localhost:8000",
			},
		},
	}
	ops := make(map[string]*openapi3.Operation)
	for _, scenario := range scenarios {
		path := &openapi3.PathItem{
			Summary:     "",
			Description: "",
		}
		op := ops[scenario.MethodPath()]
		if op == nil {
			op = &openapi3.Operation{
				Summary:     scenario.Name,
				Description: scenario.Description,
				OperationID: sanitizeScenarioName(scenario.Name),
			}
			ops[scenario.MethodPath()] = op
		}
		reqRef, reqBody := updateScenarioRequest(scenario, op)
		resRef, resBody := updateScenarioResponse(scenario, op)
		if reqBody != nil && len(reqBody.Properties) > 0 {
			root.Components.Schemas[reqRef] = &openapi3.SchemaRef{
				Value: reqBody,
			}
		}
		if resBody != nil && len(resBody.Properties) > 0 {
			root.Components.Schemas[resRef] = &openapi3.SchemaRef{
				Value: resBody,
			}
		}
		for name, auth := range scenario.Authentication {
			root.Components.SecuritySchemes[name] = &openapi3.SecuritySchemeRef{
				Value: &openapi3.SecurityScheme{
					Type:             auth.Type,
					Name:             auth.Name,
					In:               auth.In,
					Scheme:           auth.Scheme,
					BearerFormat:     auth.Format,
					OpenIdConnectUrl: auth.URL,
				},
			}
		}
		switch scenario.Method {
		case types.Post:
			path.Post = op
		case types.Get:
			path.Get = op
		case types.Put:
			path.Put = op
		case types.Delete:
			path.Delete = op
		case types.Options:
			path.Options = op
		case types.Head:
			path.Head = op
		case types.Patch:
			path.Patch = op
		case types.Connect:
			path.Connect = op
		case types.Option:
			path.Options = op
		case types.Trace:
			path.Trace = op
		}
		root.AddOperation(scenario.Path, string(scenario.Method), op)
	}
	return root
}

func buildParameter(k string, v string, in string) *openapi3.Parameter {
	v, _ = sanitizeRegexValue(v)
	return &openapi3.Parameter{
		Name:     k,
		In:       in,
		Required: in == "path",
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "string",
				//Pattern: v,
				Example: fuzz.RandRegex(v),
			},
		},
	}
}

func buildParameterRef(k string, v string, in string) *openapi3.ParameterRef {
	return &openapi3.ParameterRef{
		Value: buildParameter(k, v, in),
	}
}

func addParameter(parameters openapi3.Parameters, k string, v string, in string) openapi3.Parameters {
	for _, next := range parameters {
		if next.Value.Name == k && next.Value.In == in {
			return parameters
		}
	}
	return append(parameters, buildParameterRef(k, v, in))
}

func updateScenarioRequest(scenario *types.MockScenario, op *openapi3.Operation) (string, *openapi3.Schema) {
	for k, v := range scenario.Request.Headers {
		op.Parameters = addParameter(op.Parameters, k, v, "header")
	}
	for k, v := range scenario.Request.QueryParams {
		op.Parameters = addParameter(op.Parameters, k, v, "query")
	}
	for k, v := range scenario.Request.PathParams {
		op.Parameters = addParameter(op.Parameters, k, v, "path")
	}

	res, _ := fuzz.UnmarshalArrayOrObject([]byte(scenario.Request.AssertContentsPatternOrContent()))
	body := anyToSchema(res)
	ref := sanitizeScenarioName(scenario.Name) + "Request"
	if body != nil && len(body.Properties) > 0 {
		op.RequestBody = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Required: true,
				Content: openapi3.Content{
					scenario.Request.ContentType("application/json"): &openapi3.MediaType{
						Schema: &openapi3.SchemaRef{
							//Value: body,
							Ref: "#/components/schemas/" + ref,
						},
						Example: scenario.Request.ExampleContents,
					},
				},
			},
		}
	}
	return ref, body
}

func updateScenarioResponse(scenario *types.MockScenario, op *openapi3.Operation) (string, *openapi3.Schema) {
	if op.Responses == nil {
		op.Responses = make(openapi3.Responses)
	}
	resp := &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Headers: make(openapi3.Headers),
			Content: make(openapi3.Content),
		},
	}
	op.Responses[fmt.Sprintf("%d", scenario.Response.StatusCode)] = resp
	for k, v := range scenario.Response.Headers {
		resp.Value.Headers[k] = &openapi3.HeaderRef{
			Value: &openapi3.Header{
				Parameter: *buildParameter(k, v[0], "header"),
			},
		}
	}
	res, _ := fuzz.UnmarshalArrayOrObject([]byte(scenario.Response.AssertContentsPatternOrContent()))
	body := anyToSchema(res)
	ref := sanitizeScenarioName(scenario.Name) + "Response"

	if body != nil && len(body.Properties) > 0 && scenario.Response.StatusCode == 200 {
		resp.Value.Content[scenario.Response.ContentType("application/json")] = &openapi3.MediaType{
			Schema: &openapi3.SchemaRef{
				Ref: "#/components/schemas/" + ref,
				//Value: body,
			},
			Example: scenario.Response.ExampleContents,
		}
	}
	return ref, body
}

func sanitizeScenarioName(name string) string {
	return regexp.MustCompile(`-\d{3}-.*`).ReplaceAllString(name, "")
}

func sanitizeRegexValue(val any) (string, any) {
	strVal := fmt.Sprintf("%v", val)
	if reflect.TypeOf(val).String() == "string" {
		if strings.Contains(strVal, fuzz.PrefixTypeNumber) || strings.Contains(strVal, "RandNum") {
			strVal = ""
			if strings.Contains(strVal, ".") {
				val = 0.0
			} else {
				val = 0
			}
		} else if strings.Contains(strVal, fuzz.PrefixTypeBoolean) {
			strVal = ""
			val = false
		} else if strings.Contains(strVal, fuzz.PrefixTypeObject) {
			strVal = ""
			val = make(map[string]string)
		} else if strings.Contains(strVal, fuzz.PrefixTypeArray) {
			strVal = ""
			if strings.Contains(strVal, fuzz.PrefixTypeBoolean) {
				val = make([]bool, 0)
			} else if strings.Contains(strVal, fuzz.PrefixTypeNumber) {
				val = make([]float64, 0)
			} else {
				val = make([]string, 0)
			}
		} else if strings.Contains(strVal, "{{") {
			strVal = ""
			val = false
		}
	}

	return fuzz.StripTypeTags(strVal), val
}

// anyToProperty
func anyToSchema(val any) *openapi3.Schema {
	if val == nil {
		return nil
	}
	maxLen := uint64(100)
	strVal, val := sanitizeRegexValue(val)
	switch val.(type) {
	case map[string]string:
		hm := val.(map[string]string)
		prop := &openapi3.Schema{
			Description: "object-string-map " + strVal,
			Properties:  make(openapi3.Schemas),
			Type:        "object",
		}
		for k := range hm {
			prop.Properties[k] = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: "string",
				},
			}
		}
		return prop
	case map[string]any:
		hm := val.(map[string]any)
		prop := &openapi3.Schema{
			Description: "object-any-map " + strVal,
			Properties:  make(openapi3.Schemas),
			Type:        "object",
		}
		for k, v := range hm {
			grandChild := anyToSchema(v)
			if grandChild != nil {
				prop.Properties[k] = &openapi3.SchemaRef{
					Value: grandChild,
				}
			}
		}
		return prop
	case []bool:
		prop := &openapi3.Schema{
			Description: "bool-array " + strVal,
			Type:        "array",
			Items: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type: "boolean",
			}},
			MaxItems: &maxLen,
		}
		return prop
	case []string:
		arr := val.([]string)
		prop := &openapi3.Schema{
			Description: "string-array " + strVal,
			Type:        "array",
			Items: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type: "string",
			}},
			MaxItems: &maxLen,
		}
		for _, v := range arr {
			prop.Items.Value.Example = v
			prop.Items.Value.Pattern = v
		}
		return prop
	case []any:
		arr := val.([]any)
		prop := &openapi3.Schema{
			Description: "any-array " + strVal,
			Type:        "array",
			Items: &openapi3.SchemaRef{Value: &openapi3.Schema{
				Properties: make(openapi3.Schemas),
			}},
			MaxItems: &maxLen,
		}
		for _, v := range arr {
			grandChild := anyToSchema(v)
			if grandChild != nil {
				prop.Items.Value.Example = v
				if grandChild.Type == "integer" || grandChild.Type == "float" ||
					grandChild.Type == "string" || grandChild.Type == "bool" {
					prop.Items.Value.Pattern = grandChild.Pattern
				}
				prop.Items.Value.Type = grandChild.Type
				//prop.Items.Value.Properties[fmt.Sprintf("%d", i)] = &openapi3.SchemaRef{
				//	Value: grandChild,
				//}
			}
		}
		return prop
	case bool:
		return &openapi3.Schema{
			Type:        "bool",
			Description: strVal,
		}
	case int:
		return &openapi3.Schema{
			Type:        "integer",
			Description: strVal,
		}
	case int8:
		return &openapi3.Schema{
			Type:        "integer",
			Description: strVal,
		}
	case int16:
		return &openapi3.Schema{
			Type:        "integer",
			Description: strVal,
		}
	case int32:
		return &openapi3.Schema{
			Type:        "integer",
			Description: strVal,
		}
	case int64:
		return &openapi3.Schema{
			Type:        "integer",
			Description: strVal,
		}
	case uint:
		return &openapi3.Schema{
			Type:        "integer",
			Description: strVal,
		}
	case uint8:
		return &openapi3.Schema{
			Type:        "integer",
			Description: strVal,
		}
	case uint16:
		return &openapi3.Schema{
			Type:        "integer",
			Description: strVal,
		}
	case uint32:
		return &openapi3.Schema{
			Type:        "integer",
			Description: strVal,
		}
	case uint64:
		return &openapi3.Schema{
			Type:        "integer",
			Description: strVal,
		}
	case float32:
		return &openapi3.Schema{
			Type:        "float",
			Description: strVal,
		}
	case float64:
		return &openapi3.Schema{
			Type:        "float",
			Description: strVal,
		}
	case string:
		return &openapi3.Schema{
			Type:    "string",
			Pattern: strVal,
			Example: fuzz.RandRegex(strVal),
		}
	default:
		return &openapi3.Schema{
			Description: "unknown " + reflect.TypeOf(val).String() + " - " + strVal,
			Type:        "object",
		}
	}
}
