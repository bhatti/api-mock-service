package oapi

import (
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"regexp"
	"strconv"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
)

// APISpec structure
type APISpec struct {
	Title       string
	ID          string
	Description string
	Path        string
	Method      types.MethodType
	Request     Request
	Response    Response
}

// ParseAPISpec converts open-api operation to API specs
func ParseAPISpec(
	title string,
	method types.MethodType,
	path string,
	op *openapi3.Operation,
	dataTemplate fuzz.DataTemplateRequest) (specs []*APISpec) {
	specs = make([]*APISpec, 0)
	if op == nil {
		return
	}
	reqContent := make(map[string]*openapi3.MediaType)
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		reqContent = op.RequestBody.Value.Content
	}
	reqHeaders := make([]Property, 0)
	queryParams := make([]Property, 0)
	pathParams := make([]Property, 0)
	for _, param := range op.Parameters {
		if param.Value.Name == "" {
			param.Value.Name = param.Ref
		}
		if param.Value.Schema != nil {
			property := schemaToProperty(param.Value.Name, true, param.Value.In, param.Value.Schema.Value, dataTemplate)
			if param.Value.In == "path" {
				pathParams = append(pathParams, property)
			} else if param.Value.In == "header" {
				reqHeaders = append(reqHeaders, property)
			} else {
				queryParams = append(queryParams, property)
			}
		}
	}

	for status, resp := range op.Responses {
		respHeaders := extractHeaders(resp.Value.Headers, dataTemplate)

		for resContentType, resMedia := range resp.Value.Content {
			if len(reqContent) > 0 {
				for reqContentType, reqMedia := range reqContent {
					reqHeaders = append(reqHeaders, Property{Name: "ContentsType", Regex: reqContentType, Type: "string", In: "header"})
					spec := &APISpec{
						Title:       title,
						ID:          op.OperationID,
						Description: op.Description,
						Method:      method,
						Path:        path,
						Request: Request{
							Headers:     reqHeaders,
							QueryParams: queryParams,
							PathParams:  pathParams,
							Body:        make([]Property, 0),
						},
						Response: Response{
							Headers:     respHeaders,
							ContentType: resContentType,
							Body: []Property{schemaToProperty("", false,
								"body", resMedia.Schema.Value, dataTemplate)},
							StatusCode: parseResponseStatus(status),
						},
					}
					if op.RequestBody != nil && op.RequestBody.Value != nil {
						spec.Request.Body = append(spec.Request.Body,
							schemaToProperty("", true,
								"body", reqMedia.Schema.Value, dataTemplate))
					}
					specs = append(specs, spec)
				}
			} else {
				spec := &APISpec{
					ID:          op.OperationID,
					Description: op.Description,
					Method:      method,
					Path:        path,
					Request: Request{
						Headers:     reqHeaders,
						QueryParams: queryParams,
						PathParams:  pathParams,
						Body:        make([]Property, 0),
					},
					Response: Response{
						Headers:     respHeaders,
						ContentType: resContentType,
						Body: []Property{schemaToProperty("", false,
							"body", resMedia.Schema.Value, dataTemplate)},
						StatusCode: parseResponseStatus(status),
					},
				}
				specs = append(specs, spec)
			}
		}
	}
	return
}

// BuildMockScenario builds mock scenario from API spec
func (api *APISpec) BuildMockScenario(dataTemplate fuzz.DataTemplateRequest) (*types.MockScenario, error) {
	req, err := api.Request.buildMockHTTPRequest(dataTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to build request for mock scenario %s - %s due to %w", api.Path, api.Method, err)
	}
	res, err := api.Response.buildMockHTTPResponse(dataTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to build response for mock scenario due to %w", err)
	}

	spec := &types.MockScenario{
		Description:     api.Description,
		Method:          api.Method,
		Path:            api.Path,
		Group:           utils.NormalizeGroup(api.Title, api.Path),
		Request:         req,
		Response:        res,
		WaitBeforeReply: 0,
	}
	if res.StatusCode >= 300 {
		spec.Predicate = "{{NthRequest 2}}"
	}
	spec.Name = api.ID + "-" + spec.Digest()
	return spec, nil
}

func marshalPropertyValueWithTypes(params []Property, dataTemplate fuzz.DataTemplateRequest) (out string, err error) {
	matchContents, err := marshalPropertyValue(params, dataTemplate.WithInclude(true))
	if err != nil {
		return "", fmt.Errorf("failed to marshal params due to %w", err)
	}
	res, err := fuzz.UnmarshalArrayOrObject(matchContents)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal params object/array '%s' due to %w", matchContents, err)
	}
	j, err := json.Marshal(fuzz.FlatRegexMap(res))
	if err != nil {
		return "", fmt.Errorf("failed to marshal params flat map due to %w", err)
	}
	return string(j), nil
}

func marshalPropertyValue(params []Property, dataTemplate fuzz.DataTemplateRequest) (out []byte, err error) {
	out = []byte{}
	arr := propertyValue(params, dataTemplate)
	if len(arr) > 1 {
		out, err = json.Marshal(arr)
	} else if len(arr) > 0 {
		out, err = json.Marshal(arr[0])
	}
	return stripQuotes(out), nil
}

func stripQuotes(b []byte) []byte {
	re := regexp.MustCompile(`"{{(RandNumMinMax \d \d|RandStringArrayMinMax \d \d|RandDict|RandBool)}}"`)
	return []byte(re.ReplaceAllString(string(b), `{{$1}}`))
}

func propertyValue(params []Property, dataTemplate fuzz.DataTemplateRequest) (res []any) {
	for _, param := range params {
		val := param.Value(dataTemplate)
		if val != nil {
			res = append(res, val)
		}
	}
	return
}

func schemaToProperty(
	name string,
	matchRequest bool,
	in string,
	schema *openapi3.Schema,
	dataTemplate fuzz.DataTemplateRequest) Property {
	property := Property{
		Name:         name,
		Description:  schema.Description,
		Type:         schema.Type,
		Format:       schema.Format,
		Regex:        schema.Pattern,
		In:           in,
		Children:     make([]Property, 0),
		matchRequest: matchRequest,
	}
	if property.Type == "integer" || property.Type == "float" {
		property.Min = utils.ToFloat64(schema.Min)
		property.Max = utils.ToFloat64(schema.Max)
	} else if property.Type == "string" {
		property.Min = utils.ToFloat64(schema.MinLength)
		property.Max = utils.ToFloat64(schema.MaxLength)
	} else if property.Type == "array" {
		property.Min = utils.ToFloat64(schema.MinItems)
		property.Max = utils.ToFloat64(schema.MaxItems)
	}
	if schema.Enum != nil {
		property.Enum = make([]string, len(schema.Enum))
		for i, next := range schema.Enum {
			property.Enum[i] = fmt.Sprintf("%v", next)
		}
	}
	if schema.Items != nil && schema.Items.Value != nil {
		property.SubType = schema.Items.Value.Type
		for name, next := range schema.Items.Value.Properties {
			childProperty := schemaToProperty(name, matchRequest, in, next.Value, dataTemplate)
			property.Children = append(property.Children, childProperty)
		}
		addAllAnySchemaToProperty(schema.Items.Value, &property, matchRequest, in, dataTemplate)
	}
	addAllAnySchemaToProperty(schema, &property, matchRequest, in, dataTemplate)
	for name, prop := range schema.Properties {
		property.Children = append(property.Children, schemaToProperty(name, matchRequest, in, prop.Value, dataTemplate))
	}
	if schema.AdditionalProperties != nil {
		for name, prop := range schema.AdditionalProperties.Value.Properties {
			property.Children = append(property.Children, schemaToProperty(name, matchRequest, in, prop.Value, dataTemplate))
		}
	}
	if property.In == "body" {
		log.WithFields(log.Fields{
			"component": "Property",
			"Name":      property.Name,
			"In":        property.In,
			"Children":  len(property.Children),
			"Type":      property.Type,
			"Value":     property.Value(dataTemplate),
		}).Debugf("parsing property")
	}

	return property
}

func addAllAnySchemaToProperty(
	schema *openapi3.Schema,
	property *Property,
	matchRequest bool, in string,
	dataTemplate fuzz.DataTemplateRequest,
) {
	for _, next := range schema.AllOf {
		property.SubType = next.Value.Type
		for name, prop := range next.Value.Properties {
			property.Children = append(property.Children, schemaToProperty(name, matchRequest, in, prop.Value, dataTemplate))
		}
	}
	// TODO add support for any-of/one-of at the property
	for _, next := range schema.AllOf {
		property.SubType = next.Value.Type
		for name, prop := range next.Value.Properties {
			property.Children = append(property.Children, schemaToProperty(name, matchRequest, in, prop.Value, dataTemplate))
		}
	}
	for _, next := range schema.AllOf {
		property.SubType = next.Value.Type
		for name, prop := range next.Value.Properties {
			property.Children = append(property.Children, schemaToProperty(name, matchRequest, in, prop.Value, dataTemplate))
		}
	}
}

func extractHeaders(
	headers openapi3.Headers,
	dataTemplate fuzz.DataTemplateRequest,
) (res []Property) {
	for k, header := range headers {
		if header.Value.Schema == nil {
			continue
		}
		if header.Value.Name == "" {
			header.Value.Name = k
		}
		property := schemaToProperty(header.Value.Name, false, header.Value.In, header.Value.Schema.Value, dataTemplate)
		res = append(res, property)
	}
	return
}

func parseResponseStatus(status string) int {
	code, err := strconv.Atoi(status)
	if err != nil {
		code = 200
	}
	return code
}
