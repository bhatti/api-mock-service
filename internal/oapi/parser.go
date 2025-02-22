package oapi

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v2"
	"strings"
)

// Parse parses Open-API and generates api scenarios
func Parse(ctx context.Context, config *types.Configuration, data []byte,
	dataTemplate fuzz.DataTemplateRequest) (specs []*APISpec, updated []byte, err error) {
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}

	doc, err := loader.LoadFromData(data)
	if err != nil {
		doc = &openapi3.T{}
		if err = json.Unmarshal(data, doc); err != nil {
			return nil, nil, fmt.Errorf("failed to parse open-api with size %d due to %w", len(data), err)
		}
		if err := loader.ResolveRefsIn(doc, nil); err != nil {
			return nil, nil, fmt.Errorf("failed to resolve refs in open-api with size %d due to %w", len(data), err)
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
		for _, spec := range ParseAPISpec(title, types.Delete, k, v.Delete, dataTemplate) {
			specs = append(specs, spec)
		}
		for _, spec := range ParseAPISpec(title, types.Get, k, v.Get, dataTemplate) {
			specs = append(specs, spec)
		}
		for _, spec := range ParseAPISpec(title, types.Post, k, v.Post, dataTemplate) {
			specs = append(specs, spec)
		}
		for _, spec := range ParseAPISpec(title, types.Put, k, v.Put, dataTemplate) {
			specs = append(specs, spec)
		}
		for _, spec := range ParseAPISpec(title, types.Patch, k, v.Patch, dataTemplate) {
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
				spec.Request.QueryParams = append(spec.Request.Headers, prop)
			}
			spec.SecuritySchemes = doc.Components.SecuritySchemes
		}
	}
	updated, err = yaml.Marshal(doc)
	return
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
