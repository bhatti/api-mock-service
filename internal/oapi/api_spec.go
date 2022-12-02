package oapi

import (
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
)

// APISpec structure
type APISpec struct {
	ID          string
	Description string
	Path        string
	Method      types.MethodType
	Request     Request
	Response    Response
}

// ParseAPISpec converts open-api operation to API specs
func ParseAPISpec(method types.MethodType, path string, op *openapi3.Operation) (specs []*APISpec) {
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
			property := schemaToProperty(param.Value.Name, true, param.Value.In, param.Value.Schema.Value)
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
		respHeaders := extractHeaders(resp.Value.Headers)

		for resContentType, resMedia := range resp.Value.Content {
			if len(reqContent) > 0 {
				for reqContentType, reqMedia := range reqContent {
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
							ContentType: reqContentType,
						},
						Response: Response{
							Headers:     respHeaders,
							ContentType: resContentType,
							Body:        []Property{schemaToProperty("", false, "body", resMedia.Schema.Value)},
							StatusCode:  parseResponseStatus(status),
						},
					}
					if op.RequestBody != nil && op.RequestBody.Value != nil {
						spec.Request.ContentType = reqContentType
						spec.Request.Body = append(spec.Request.Body,
							schemaToProperty("", true, "body", reqMedia.Schema.Value))
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
						ContentType: "",
					},
					Response: Response{
						Headers:     respHeaders,
						ContentType: resContentType,
						Body:        []Property{schemaToProperty("", false, "body", resMedia.Schema.Value)},
						StatusCode:  parseResponseStatus(status),
					},
				}
				specs = append(specs, spec)
			}
		}
	}
	return
}

// BuildMockScenario builds mock scenario from API spec
func (api *APISpec) BuildMockScenario() (*types.MockScenario, error) {
	req, err := api.Request.buildMockHTTPRequest()
	if err != nil {
		return nil, err
	}
	res, err := api.Response.buildMockHTTPResponse()
	if err != nil {
		return nil, err
	}

	spec := &types.MockScenario{
		Description:     api.Description,
		Method:          api.Method,
		Path:            api.Path,
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

func marshalPropertyValue(params []Property) (out []byte, err error) {
	out = []byte{}
	arr := propertyValue(params)
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

func propertyValue(params []Property) (res []interface{}) {
	for _, param := range params {
		val := param.Value()
		if val != nil {
			res = append(res, val)
		}
	}
	return
}

func schemaToProperty(name string, matchRequest bool, in string, schema *openapi3.Schema) Property {
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
	if schema.Items != nil {
		property.SubType = schema.Items.Value.Type
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
		for name, prop := range schema.Items.Value.Properties {
			property.Children = append(property.Children, schemaToProperty(name, matchRequest, in, prop.Value))
		}
	}
	for name, prop := range schema.Properties {
		property.Children = append(property.Children, schemaToProperty(name, matchRequest, in, prop.Value))
	}
	if schema.AdditionalProperties != nil {
		for name, prop := range schema.AdditionalProperties.Value.Properties {
			property.Children = append(property.Children, schemaToProperty(name, matchRequest, in, prop.Value))
		}
	}
	if property.In == "body" {
		log.WithFields(log.Fields{
			"component": "Property",
			"Name":      property.Name,
			"In":        property.In,
			"Children":  len(property.Children),
			"Type":      property.Type,
			"Value":     property.Value(),
		}).Debugf("parsing property")
	}

	return property
}

func extractHeaders(headers openapi3.Headers) (res []Property) {
	for k, header := range headers {
		if header.Value.Schema == nil {
			continue
		}
		if header.Value.Name == "" {
			header.Value.Name = k
		}
		property := schemaToProperty(header.Value.Name, false, header.Value.In, header.Value.Schema.Value)
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
