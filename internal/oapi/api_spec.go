package oapi

import (
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"regexp"
	"strconv"
	"strings"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
)

// APISpec structure
type APISpec struct {
	Title               string
	ID                  string
	Summary             string
	Description         string
	Path                string
	Method              types.MethodType
	Tags                []string
	Deprecated          bool
	ExternalDocsURL     string
	RequestBodyRequired bool
	SecuritySchemes     openapi3.SecuritySchemes
	Request             Request
	Response            Response
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

	// Extract all parameters first
	reqHeaders, queryParams, pathParams := extractParameters(op.Parameters, dataTemplate)

	// Handle request body content types
	reqContent := make(map[string]*openapi3.MediaType)
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		reqContent = op.RequestBody.Value.Content
	}

	// Process responses
	for status, resp := range op.Responses {
		if resp.Value == nil {
			continue
		}

		respHeaders := extractHeaders(resp.Value.Headers, dataTemplate)

		// Handle no content responses
		if len(resp.Value.Content) == 0 {
			// Maintain backward compatibility by still looping through reqContent
			if len(reqContent) > 0 {
				for reqContentType, reqMedia := range reqContent {
					headers := append([]Property{}, reqHeaders...) // Create copy of headers
					headers = append(headers, Property{
						Name:    "Content-Type",
						Pattern: reqContentType,
						Type:    "string",
						In:      "header",
					})

					spec := createBaseSpec(op, title, method, path, status)
					spec.Request = Request{
						Headers:     headers,
						QueryParams: queryParams,
						PathParams:  pathParams,
						Body:        make([]Property, 0),
					}
					if op.RequestBody != nil && op.RequestBody.Value != nil {
						spec.Request.Body = append(spec.Request.Body,
							schemaToProperty("", true, "body", reqMedia.Schema.Value, dataTemplate))
					}
					spec.Response = Response{
						Headers:    respHeaders,
						StatusCode: parseResponseStatus(status),
					}
					specs = append(specs, spec)
				}
			} else {
				// No request content, create single spec
				spec := createBaseSpec(op, title, method, path, status)
				spec.Request = Request{
					Headers:     reqHeaders,
					QueryParams: queryParams,
					PathParams:  pathParams,
					Body:        make([]Property, 0),
				}
				spec.Response = Response{
					Headers:    respHeaders,
					StatusCode: parseResponseStatus(status),
				}
				specs = append(specs, spec)
			}
			continue
		}

		// Handle responses with content
		for resContentType, resMedia := range resp.Value.Content {
			if resMedia.Schema == nil || resMedia.Schema.Value == nil {
				continue
			}

			if len(reqContent) > 0 {
				for reqContentType, reqMedia := range reqContent {
					headers := append([]Property{}, reqHeaders...) // Create copy of headers
					headers = append(headers, Property{
						Name:    "Content-Type",
						Pattern: reqContentType,
						Type:    "string",
						In:      "header",
					})

					spec := createBaseSpec(op, title, method, path, status)
					spec.Request = Request{
						Headers:     headers,
						QueryParams: queryParams,
						PathParams:  pathParams,
						Body:        make([]Property, 0),
					}
					if op.RequestBody != nil && op.RequestBody.Value != nil {
						spec.Request.Body = append(spec.Request.Body,
							schemaToProperty("", true, "body", reqMedia.Schema.Value, dataTemplate))
					}
					spec.Response = Response{
						Headers:     respHeaders,
						ContentType: resContentType,
						Body: []Property{schemaToProperty("", false,
							"body", resMedia.Schema.Value, dataTemplate)},
						StatusCode: parseResponseStatus(status),
					}
					specs = append(specs, spec)
				}
			} else {
				// No request content, create single spec
				spec := createBaseSpec(op, title, method, path, status)
				spec.Request = Request{
					Headers:     reqHeaders,
					QueryParams: queryParams,
					PathParams:  pathParams,
					Body:        make([]Property, 0),
				}
				spec.Response = Response{
					Headers:     respHeaders,
					ContentType: resContentType,
					Body: []Property{schemaToProperty("", false,
						"body", resMedia.Schema.Value, dataTemplate)},
					StatusCode: parseResponseStatus(status),
				}
				specs = append(specs, spec)
			}
		}
	}

	return specs
}

// BuildMockScenario builds api scenario from open-API spec
func (api *APISpec) BuildMockScenario(dataTemplate fuzz.DataTemplateRequest) (*types.APIScenario, error) {
	req, err := api.Request.buildMockHTTPRequest(dataTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to build request for api scenario %s - %s due to %w", api.Path, api.Method, err)
	}
	res, err := api.Response.buildMockHTTPResponse(dataTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to build response for api scenario due to %w", err)
	}

	tags := api.Tags
	if len(tags) == 0 {
		tags = []string{types.NormalizeGroup(api.Title, api.Path)}
	}
	spec := &types.APIScenario{
		Description:     api.Description,
		Method:          api.Method,
		Path:            api.Path,
		Group:           types.NormalizeGroup(api.Title, api.Path),
		Tags:            tags,
		Request:         req,
		Response:        res,
		WaitBeforeReply: 0,
		Authentication:  make(map[string]types.APIAuthorization),
	}
	for name, scheme := range api.SecuritySchemes {
		auth := types.APIAuthorization{
			Type:        scheme.Value.Type,
			Name:        scheme.Value.Name,
			In:          scheme.Value.In,
			Format:      scheme.Value.BearerFormat,
			Scheme:      scheme.Value.Scheme,
			URL:         scheme.Value.OpenIdConnectUrl,
			Description: scheme.Value.Description,
		}
		if scheme.Value.Flows != nil {
			auth.Scopes = extractOAuthScopes(scheme.Value.Flows)
		}
		spec.Authentication[name] = auth
	}
	// If the request body is required, assert it is present.
	// Note: PropertyLenGE contents N counts top-level keys for object bodies —
	// not byte length — so it fails for single-key objects like {"ids":[...]}.
	// HasProperty contents correctly checks that the body is non-nil regardless of type.
	if api.RequestBodyRequired {
		spec.Request.Assertions = types.AddAssertion(spec.Request.Assertions, "HasProperty contents")
	}
	if res.StatusCode >= 300 {
		spec.Predicate = "{{NthRequest 2}}"
	} else {
		spec.Predicate = "{{NthRequest 1}}"
	}
	spec.SetName(api.ID + "-")
	return spec, nil
}

func marshalPropertyValueWithTypes(params []Property, dataTemplate fuzz.DataTemplateRequest, stripQuotes bool) (out string, err error) {
	matchContents, err := marshalPropertyValue(params, dataTemplate.WithInclude(true), stripQuotes)
	if err != nil {
		return "", fmt.Errorf("failed to marshal params due to %w", err)
	}
	res, err := fuzz.UnmarshalArrayOrObject(matchContents)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal params object/array '%s' due to %w", matchContents, err)
	}
	j, err := json.MarshalIndent(fuzz.FlatRegexMap(res), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal params flat map due to %w", err)
	}
	return string(j), nil
}

func marshalPropertyValue(params []Property, dataTemplate fuzz.DataTemplateRequest, stripQuotes bool) (out []byte, err error) {
	out = []byte{}
	arr := propertyValue(params, dataTemplate)
	if len(arr) > 1 {
		out, err = json.MarshalIndent(arr, "", "  ")
	} else if len(arr) > 0 {
		out, err = json.MarshalIndent(arr[0], "", "  ")
	}
	if stripQuotes {
		return stripNumericBooleanQuotes(out), nil
	}
	return out, nil

}

func stripNumericBooleanQuotes(b []byte) []byte {
	re := regexp.MustCompile(`"{{(RandIntMinMax \d \d|RandStringArrayMinMax \d \d|RandDict|RandBool)}}"`)
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

// schemaValueToString converts an OpenAPI schema value (interface{}) to a string for storage
func schemaValueToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// constFromSchema extracts a single const value: OpenAPI encodes const as a single-element enum
func constFromSchema(schema *openapi3.Schema) string {
	if len(schema.Enum) == 1 {
		return fmt.Sprintf("%v", schema.Enum[0])
	}
	return ""
}

// ptrUint64 safely dereferences a *uint64
func ptrUint64(p *uint64) uint64 {
	if p == nil {
		return 0
	}
	return *p
}

// maxSchemaDepth is the maximum recursion depth for schema traversal. Real-world schemas
// are rarely deeper than 10 levels; 20 is a safe ceiling that stops infinite recursion
// caused by circular $ref schemas regardless of whether kin-openapi reuses schema pointers.
const maxSchemaDepth = 20

func schemaToProperty(
	name string,
	matchRequest bool,
	in string,
	schema *openapi3.Schema,
	dataTemplate fuzz.DataTemplateRequest) Property {
	return schemaToPropertyV(name, matchRequest, in, schema, dataTemplate, make(map[*openapi3.Schema]bool), 0)
}

// schemaToPropertyV is the cycle-safe implementation.
// visited tracks schema pointers already on the current call stack (pointer identity).
// depth is an independent depth counter that stops recursion regardless of pointer reuse.
// Both guards together handle all circular $ref patterns in OpenAPI 3.0 and 3.1 specs.
func schemaToPropertyV(
	name string,
	matchRequest bool,
	in string,
	schema *openapi3.Schema,
	dataTemplate fuzz.DataTemplateRequest,
	visited map[*openapi3.Schema]bool,
	depth int) Property {
	if depth > maxSchemaDepth || visited[schema] {
		// Circular reference or max depth — return a stub to break the cycle.
		return Property{Name: name, Type: schema.Type, In: in, matchRequest: matchRequest, Children: make([]Property, 0)}
	}
	visited[schema] = true
	defer func() { delete(visited, schema) }()

	property := Property{
		Name:         name,
		Title:        schema.Title,
		Description:  schema.Description,
		Type:         schema.Type,
		Format:       schema.Format,
		Pattern:      schema.Pattern,
		In:           in,
		Children:     make([]Property, 0),
		matchRequest: matchRequest,
		Nullable:     schema.Nullable,
		ReadOnly:     schema.ReadOnly,
		WriteOnly:    schema.WriteOnly,
		Deprecated:   schema.Deprecated,
		Const:        constFromSchema(schema),
		Default:      schemaValueToString(schema.Default),
		Example:      schemaValueToString(schema.Example),
		UniqueItems:  schema.UniqueItems,
		ExclusiveMin: schema.ExclusiveMin,
		ExclusiveMax: schema.ExclusiveMax,
		MultipleOf:   fuzz.ToFloat64(schema.MultipleOf),
		MinProps:     schema.MinProps,
		MaxProps:     ptrUint64(schema.MaxProps),
	}
	if property.Type == "integer" || property.Type == "number" || property.Type == "float" {
		property.Min = fuzz.ToFloat64(schema.Min)
		property.Max = fuzz.ToFloat64(schema.Max)
	} else if property.Type == "string" {
		property.Min = fuzz.ToFloat64(schema.MinLength)
		property.Max = fuzz.ToFloat64(schema.MaxLength)
	} else if property.Type == "array" {
		property.Min = fuzz.ToFloat64(schema.MinItems)
		property.Max = fuzz.ToFloat64(schema.MaxItems)
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
			if next.Value == nil {
				continue
			}
			childProperty := schemaToPropertyV(name, matchRequest, in, next.Value, dataTemplate, visited, depth+1)
			property.Children = append(property.Children, childProperty)
		}
		addAllAnySchemaToPropertyV(schema.Items.Value, &property, matchRequest, in, dataTemplate, visited, depth+1)
	}
	addAllAnySchemaToPropertyV(schema, &property, matchRequest, in, dataTemplate, visited, depth+1)
	for name, prop := range schema.Properties {
		if prop.Value == nil {
			continue
		}
		property.Children = append(property.Children, schemaToPropertyV(name, matchRequest, in, prop.Value, dataTemplate, visited, depth+1))
	}
	if schema.AdditionalProperties != nil && schema.AdditionalProperties.Value != nil {
		for name, prop := range schema.AdditionalProperties.Value.Properties {
			if prop.Value == nil {
				continue
			}
			property.Children = append(property.Children, schemaToPropertyV(name, matchRequest, in, prop.Value, dataTemplate, visited, depth+1))
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
	addAllAnySchemaToPropertyV(schema, property, matchRequest, in, dataTemplate, make(map[*openapi3.Schema]bool), 0)
}

func addAllAnySchemaToPropertyV(
	schema *openapi3.Schema,
	property *Property,
	matchRequest bool, in string,
	dataTemplate fuzz.DataTemplateRequest,
	visited map[*openapi3.Schema]bool,
	depth int,
) {
	// allOf: merge all sub-schemas (intersection type — all properties required)
	for _, next := range schema.AllOf {
		if next.Value == nil {
			continue
		}
		if property.SubType == "" {
			property.SubType = next.Value.Type
		}
		for name, prop := range next.Value.Properties {
			if prop.Value == nil {
				continue
			}
			property.Children = append(property.Children, schemaToPropertyV(name, matchRequest, in, prop.Value, dataTemplate, visited, depth+1))
		}
	}
	// oneOf: use first branch as representative mock schema (mutually exclusive variants)
	for _, next := range schema.OneOf {
		if next.Value == nil {
			continue
		}
		if property.SubType == "" {
			property.SubType = next.Value.Type
		}
		for name, prop := range next.Value.Properties {
			if prop.Value == nil {
				continue
			}
			property.Children = append(property.Children, schemaToPropertyV(name, matchRequest, in, prop.Value, dataTemplate, visited, depth+1))
		}
		break // only first oneOf branch needed for mock generation
	}
	// anyOf: use first branch as representative mock schema (one or more valid variants)
	for _, next := range schema.AnyOf {
		if next.Value == nil {
			continue
		}
		if property.SubType == "" {
			property.SubType = next.Value.Type
		}
		for name, prop := range next.Value.Properties {
			if prop.Value == nil {
				continue
			}
			property.Children = append(property.Children, schemaToPropertyV(name, matchRequest, in, prop.Value, dataTemplate, visited, depth+1))
		}
		break // only first anyOf branch needed for mock generation
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
		inValue := header.Value.In
		if inValue == "" {
			inValue = "header"
		}
		property := schemaToProperty(header.Value.Name, false, inValue, header.Value.Schema.Value, dataTemplate)
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

func createBaseSpec(op *openapi3.Operation, title string, method types.MethodType, path string, status string) *APISpec {
	externalDocsURL := ""
	if op.ExternalDocs != nil {
		externalDocsURL = op.ExternalDocs.URL
	}
	requestBodyRequired := false
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		requestBodyRequired = op.RequestBody.Value.Required
	}
	return &APISpec{
		ID:                  op.OperationID,
		Summary:             op.Summary,
		Description:         op.Description,
		Method:              method,
		Path:                path,
		Title:               title,
		Tags:                op.Tags,
		Deprecated:          op.Deprecated,
		ExternalDocsURL:     externalDocsURL,
		RequestBodyRequired: requestBodyRequired,
	}
}

func extractParameters(params openapi3.Parameters, dataTemplate fuzz.DataTemplateRequest) (reqHeaders, queryParams, pathParams []Property) {
	reqHeaders = make([]Property, 0)
	queryParams = make([]Property, 0)
	pathParams = make([]Property, 0)

	for _, param := range params {
		if param.Value == nil || param.Value.Schema == nil || param.Value.Schema.Value == nil {
			continue
		}

		if param.Value.Name == "" {
			param.Value.Name = param.Ref
		}

		property := schemaToProperty(param.Value.Name, true, param.Value.In, param.Value.Schema.Value, dataTemplate)
		property.Required = param.Value.Required
		property.Deprecated = param.Value.Deprecated
		if param.Value.Style != "" {
			property.Style = param.Value.Style
		}
		if param.Value.Explode != nil {
			property.Explode = *param.Value.Explode
		}

		switch param.Value.In {
		case "path":
			pathParams = append(pathParams, property)
		case "header":
			reqHeaders = append(reqHeaders, property)
		case "query":
			queryParams = append(queryParams, property)
		}
	}

	return reqHeaders, queryParams, pathParams
}

// DiscriminatorVariant holds a named variant schema from a oneOf/anyOf discriminator.
type DiscriminatorVariant struct {
	Name   string
	Schema *openapi3.Schema
}

// DiscriminatorVariants returns named variants for a oneOf/anyOf schema.
// Uses discriminator.mapping keys when present; otherwise derives names from $ref paths
// or falls back to "variant0", "variant1", etc.
// Returns nil when the schema has no oneOf/anyOf branches.
func DiscriminatorVariants(schema *openapi3.Schema) []DiscriminatorVariant {
	if schema == nil {
		return nil
	}
	branches := schema.OneOf
	if len(branches) == 0 {
		branches = schema.AnyOf
	}
	if len(branches) == 0 {
		return nil
	}

	// Build reverse mapping: $ref → variant name from discriminator.mapping
	refToName := make(map[string]string)
	if schema.Discriminator != nil {
		for name, ref := range schema.Discriminator.Mapping {
			refToName[ref] = name
		}
	}

	variants := make([]DiscriminatorVariant, 0, len(branches))
	for i, branch := range branches {
		if branch.Value == nil {
			continue
		}
		name := fmt.Sprintf("variant%d", i)
		if branch.Ref != "" {
			if n, ok := refToName[branch.Ref]; ok {
				name = n
			} else {
				// Derive name from $ref: "#/components/schemas/Cat" → "cat"
				parts := strings.Split(branch.Ref, "/")
				if last := parts[len(parts)-1]; last != "" {
					name = strings.ToLower(last)
				}
			}
		}
		variants = append(variants, DiscriminatorVariant{Name: name, Schema: branch.Value})
	}
	return variants
}

// hasDiscriminatorInResponses returns true if any response body schema has oneOf/anyOf variants.
func hasDiscriminatorInResponses(op *openapi3.Operation) bool {
	if op == nil {
		return false
	}
	for _, resp := range op.Responses {
		if resp.Value == nil {
			continue
		}
		for _, media := range resp.Value.Content {
			if media.Schema != nil && media.Schema.Value != nil {
				if len(DiscriminatorVariants(media.Schema.Value)) > 0 {
					return true
				}
			}
		}
	}
	return false
}

// ParseAPISpecWithDiscriminatorVariants is like ParseAPISpec but generates one *APISpec
// per discriminator variant when a response body schema contains oneOf/anyOf.
// For schemas without discriminator variants, behavior is identical to ParseAPISpec.
func ParseAPISpecWithDiscriminatorVariants(
	title string,
	method types.MethodType,
	path string,
	op *openapi3.Operation,
	dataTemplate fuzz.DataTemplateRequest,
) []*APISpec {
	if op == nil {
		return nil
	}
	if !hasDiscriminatorInResponses(op) {
		return ParseAPISpec(title, method, path, op, dataTemplate)
	}

	specs := make([]*APISpec, 0)
	reqHeaders, queryParams, pathParams := extractParameters(op.Parameters, dataTemplate)

	reqContent := make(map[string]*openapi3.MediaType)
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		reqContent = op.RequestBody.Value.Content
	}

	buildSpec := func(status string, respHeaders []Property, resContentType string, resSchema *openapi3.Schema, variantName string) {
		addReqContent := func(headers []Property, reqMedia *openapi3.MediaType) *APISpec {
			spec := createBaseSpec(op, title, method, path, status)
			spec.Request = Request{
				Headers:     headers,
				QueryParams: queryParams,
				PathParams:  pathParams,
				Body:        make([]Property, 0),
			}
			if reqMedia != nil && op.RequestBody != nil && op.RequestBody.Value != nil {
				spec.Request.Body = append(spec.Request.Body,
					schemaToProperty("", true, "body", reqMedia.Schema.Value, dataTemplate))
			}
			if resSchema != nil {
				spec.Response = Response{
					Headers:     respHeaders,
					ContentType: resContentType,
					Body:        []Property{schemaToProperty("", false, "body", resSchema, dataTemplate)},
					StatusCode:  parseResponseStatus(status),
				}
			} else {
				spec.Response = Response{
					Headers:    respHeaders,
					StatusCode: parseResponseStatus(status),
				}
			}
			if variantName != "" {
				spec.ID = spec.ID + "-" + variantName
			}
			return spec
		}

		if len(reqContent) > 0 {
			for reqContentType, reqMedia := range reqContent {
				headers := append([]Property{}, reqHeaders...)
				headers = append(headers, Property{Name: "Content-Type", Pattern: reqContentType, Type: "string", In: "header"})
				specs = append(specs, addReqContent(headers, reqMedia))
			}
		} else {
			specs = append(specs, addReqContent(reqHeaders, nil))
		}
	}

	for status, resp := range op.Responses {
		if resp.Value == nil {
			continue
		}
		respHeaders := extractHeaders(resp.Value.Headers, dataTemplate)

		if len(resp.Value.Content) == 0 {
			buildSpec(status, respHeaders, "", nil, "")
			continue
		}

		for resContentType, resMedia := range resp.Value.Content {
			if resMedia.Schema == nil || resMedia.Schema.Value == nil {
				continue
			}
			variants := DiscriminatorVariants(resMedia.Schema.Value)
			if len(variants) == 0 {
				buildSpec(status, respHeaders, resContentType, resMedia.Schema.Value, "")
			} else {
				for _, variant := range variants {
					buildSpec(status, respHeaders, resContentType, variant.Schema, variant.Name)
				}
			}
		}
	}
	return specs
}

// extractOAuthScopes collects scope name→description from all OAuth2 flows
func extractOAuthScopes(flows *openapi3.OAuthFlows) map[string]string {
	scopes := make(map[string]string)
	if flows == nil {
		return scopes
	}
	merge := func(s map[string]string) {
		for k, v := range s {
			scopes[k] = v
		}
	}
	if flows.Implicit != nil {
		merge(flows.Implicit.Scopes)
	}
	if flows.Password != nil {
		merge(flows.Password.Scopes)
	}
	if flows.ClientCredentials != nil {
		merge(flows.ClientCredentials.Scopes)
	}
	if flows.AuthorizationCode != nil {
		merge(flows.AuthorizationCode.Scopes)
	}
	return scopes
}
