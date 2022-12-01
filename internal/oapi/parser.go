package oapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/getkin/kin-openapi/openapi3"
)

// Parse parses Open-API and generates mock scenarios
func Parse(ctx context.Context, data []byte) (specs []*APISpec, err error) {
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}

	doc, err := loader.LoadFromData(data)
	if err != nil {
		doc = &openapi3.T{}
		if err = json.Unmarshal(data, doc); err != nil {
			return nil, fmt.Errorf("failed to parse open-api with size %d due to %w", len(data), err)
		}
		if err := loader.ResolveRefsIn(doc, nil); err != nil {
			return nil, fmt.Errorf("failed to resolve refs in open-api with size %d due to %w", len(data), err)
		}
	}

	for k, v := range doc.Paths {
		for _, spec := range ParseAPISpec(types.Delete, k, v.Delete) {
			specs = append(specs, spec)
		}
		for _, spec := range ParseAPISpec(types.Get, k, v.Get) {
			specs = append(specs, spec)
		}
		for _, spec := range ParseAPISpec(types.Post, k, v.Post) {
			specs = append(specs, spec)
		}
		for _, spec := range ParseAPISpec(types.Put, k, v.Put) {
			specs = append(specs, spec)
		}
		for _, spec := range ParseAPISpec(types.Patch, k, v.Patch) {
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
				Regex:       ref.Value.BearerFormat,
			}
			if ref.Value.In == "header" {
				spec.Request.Headers = append(spec.Request.Headers, prop)
			} else if ref.Value.In == "query" {
				spec.Request.QueryParams = append(spec.Request.Headers, prop)
			}
		}
	}
	return
}
