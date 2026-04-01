package oapi

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"gopkg.in/yaml.v2"
	"strings"
)

// Parse parses Open-API and generates api scenarios.
// Returns specs, the re-serialized spec bytes, the parsed openapi3.T document, and any error.
func Parse(ctx context.Context, config *types.Configuration, data []byte,
	dataTemplate fuzz.DataTemplateRequest) (specs []*APISpec, updated []byte, doc *openapi3.T, err error) {
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}

	// Normalize OpenAPI 3.1 features (e.g. type arrays) before passing to the
	// kin-openapi 3.0 loader, which cannot handle them natively.
	normalized := normalizeOpenAPI31(data)

	doc, err = loader.LoadFromData(normalized)
	if err != nil {
		doc = &openapi3.T{}
		if err = json.Unmarshal(normalized, doc); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse open-api with size %d due to %w", len(data), err)
		}
		if err := loader.ResolveRefsIn(doc, nil); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to resolve refs in open-api with size %d due to %w", len(data), err)
		}
	}

	addServers(config, doc)

	var title string
	if doc.Info != nil {
		title = doc.Info.Title
		if doc.Info.Version != "" {
			title += "_V" + doc.Info.Version
		}
	}
	if title == "" {
		title = fmt.Sprintf("OAPI_%x", sha1.Sum(data))
	}

	for k, v := range doc.Paths {
		for _, spec := range ParseAPISpecWithDiscriminatorVariants(title, types.Delete, k, v.Delete, dataTemplate) {
			specs = append(specs, spec)
		}
		for _, spec := range ParseAPISpecWithDiscriminatorVariants(title, types.Get, k, v.Get, dataTemplate) {
			specs = append(specs, spec)
		}
		for _, spec := range ParseAPISpecWithDiscriminatorVariants(title, types.Post, k, v.Post, dataTemplate) {
			specs = append(specs, spec)
		}
		for _, spec := range ParseAPISpecWithDiscriminatorVariants(title, types.Put, k, v.Put, dataTemplate) {
			specs = append(specs, spec)
		}
		for _, spec := range ParseAPISpecWithDiscriminatorVariants(title, types.Patch, k, v.Patch, dataTemplate) {
			specs = append(specs, spec)
		}
	}
	for _, ref := range doc.Components.SecuritySchemes {
		for _, spec := range specs {
			prop := Property{
				Name:        ref.Value.Name,
				Description: ref.Value.Description,
				Type:        ref.Value.Type,
				In:          ref.Value.In,
				Pattern:     ref.Value.BearerFormat,
			}
			if ref.Value.In == "header" {
				spec.Request.Headers = append(spec.Request.Headers, prop)
			} else if ref.Value.In == "query" {
				spec.Request.QueryParams = append(spec.Request.QueryParams, prop)
			}
			spec.SecuritySchemes = doc.Components.SecuritySchemes
		}
	}
	// Use kin-openapi's own marshaler so circular $ref schemas are serialized back
	// as "$ref" strings rather than as infinite Go pointer traversals, which would
	// cause yaml.Marshal (and encoding/json.Marshal) to loop forever.
	updated, err = doc.MarshalJSON()
	return
}

// BuildRouter builds a legacy kin-openapi router from a parsed OpenAPI document.
// The router is used with openapi3filter.FindRoute for response schema validation.
func BuildRouter(doc *openapi3.T) (routers.Router, error) {
	return legacy.NewRouter(doc)
}

func addServers(config *types.Configuration, doc *openapi3.T) {
	envServers := make(map[string]string)
	for _, env := range config.TestEnvironments {
		envServers[env] = ""
	}
	mockServerFound := processExistingServers(doc.Servers, envServers)
	if !mockServerFound {
		doc.Servers = append(doc.Servers, &openapi3.Server{URL: MockServerBaseURL})
	}
	for _, server := range envServers {
		if server != "" {
			doc.Servers = append(doc.Servers, &openapi3.Server{URL: server})
		}
	}
}

func processExistingServers(servers openapi3.Servers, envServers map[string]string) bool {
	mockServerFound := false
	var templateEnv, templateURL string
	for _, server := range servers {
		if strings.Contains(server.URL, MockServerBaseURL) {
			mockServerFound = true
		}
		for env := range envServers {
			if strings.Contains(server.URL, env) {
				envServers[env] = server.URL
				templateEnv = env
				templateURL = server.URL
				delete(envServers, env) // it already exists so we will skip it when adding servers
			}
		}
	}
	if templateEnv != "" && templateURL != "" {
		for env, server := range envServers {
			if server == "" {
				envServers[env] = strings.ReplaceAll(templateURL, templateEnv, env)
			}
		}
	}
	return mockServerFound
}

// normalizeOpenAPI31 converts OpenAPI 3.1 constructs that kin-openapi v0.106 (OpenAPI 3.0 only)
// cannot parse into their 3.0-equivalent forms:
//
//   - type arrays ["string","null"] → type:"string", nullable:true
//   - openapi version "3.1.x" → "3.0.3"
//
// The input may be JSON or YAML; output is always JSON so the loader can handle it uniformly.
func normalizeOpenAPI31(data []byte) []byte {
	var doc map[string]interface{}

	// Try JSON first, then YAML.
	if err := json.Unmarshal(data, &doc); err != nil {
		var yamlDoc interface{}
		if err2 := yaml.Unmarshal(data, &yamlDoc); err2 != nil {
			return data
		}
		// yaml.v2 may produce map[interface{}]interface{}; convert to map[string]interface{}.
		converted := deepConvertYAMLMap(yamlDoc)
		var ok bool
		doc, ok = converted.(map[string]interface{})
		if !ok {
			return data
		}
	}

	// Downgrade 3.1.x → 3.0.3 so the loader validates correctly.
	if v, ok := doc["openapi"].(string); ok && strings.HasPrefix(v, "3.1") {
		doc["openapi"] = "3.0.3"
	}

	normalizeTypeArrays(doc)

	out, err := json.Marshal(doc)
	if err != nil {
		return data
	}
	return out
}

// normalizeTypeArrays recursively walks a parsed JSON/YAML document and converts
// any schema "type" that is a JSON array (OpenAPI 3.1 / JSON Schema style) into
// a single string type with nullable:true when "null" is one of the variants.
func normalizeTypeArrays(v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		if typeField, ok := val["type"]; ok {
			if arr, ok := typeField.([]interface{}); ok {
				nonNull := ""
				hasNull := false
				for _, item := range arr {
					if s, ok := item.(string); ok {
						if s == "null" {
							hasNull = true
						} else if nonNull == "" {
							nonNull = s
						}
					}
				}
				if nonNull != "" {
					val["type"] = nonNull
					if hasNull {
						val["nullable"] = true
					}
				} else {
					delete(val, "type")
				}
			}
		}
		for _, child := range val {
			normalizeTypeArrays(child)
		}
	case []interface{}:
		for _, item := range val {
			normalizeTypeArrays(item)
		}
	}
}

// deepConvertYAMLMap converts map[interface{}]interface{} (produced by yaml.v2) into
// map[string]interface{} so it can be handled uniformly with JSON-decoded documents.
func deepConvertYAMLMap(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, child := range val {
			out[fmt.Sprintf("%v", k)] = deepConvertYAMLMap(child)
		}
		return out
	case []interface{}:
		for i, item := range val {
			val[i] = deepConvertYAMLMap(item)
		}
	}
	return v
}
