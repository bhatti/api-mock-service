package oapi

import (
	"context"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/getkin/kin-openapi/openapi3"
)

// Parse parses Open-API and generates mock scenarios
func Parse(ctx context.Context, data []byte) (specs []*APISpec, err error) {
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, err
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
	return
}
