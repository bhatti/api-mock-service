package contract

import (
	"context"
	"net/http"
	"testing"

	"github.com/bhatti/api-mock-service/internal/oapi"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// schemaValidationYAML is a minimal spec with a required field.
const schemaValidationYAML = `
openapi: "3.0.3"
info:
  title: Test API
  version: "1.0"
paths:
  /items:
    get:
      operationId: getItem
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                required:
                  - id
                properties:
                  id:
                    type: string
                  name:
                    type: string
`

func Test_ValidateResponse_PassesForValidResponse(t *testing.T) {
	doc := mustParseOAPIDoc(t, []byte(schemaValidationYAML))
	router, err := oapi.BuildRouter(doc)
	require.NoError(t, err)

	executor := &ProducerExecutor{
		openAPIDoc:    doc,
		openAPIRouter: router,
	}
	report := &ContractDiffReport{
		MissingFields:    []string{},
		ExtraFields:      []string{},
		TypeMismatches:   map[string]string{},
		ValueMismatches:  map[string]ValueMismatch{},
		HeaderMismatches: map[string]ValueMismatch{},
		ExpectedFields:   map[string]interface{}{},
		ActualFields:     map[string]interface{}{},
	}
	err = executor.validateResponseAgainstSchema(
		context.Background(),
		"https://mocksite.local/items",
		"GET",
		200,
		[]byte(`{"id":"abc","name":"test"}`),
		http.Header{"Content-Type": []string{"application/json"}},
		report,
	)
	require.NoError(t, err)
	require.Empty(t, report.SchemaViolations, "valid response should produce no violations")
}

func Test_ValidateResponse_SkipsUnknownRoute(t *testing.T) {
	doc := mustParseOAPIDoc(t, []byte(schemaValidationYAML))
	router, err := oapi.BuildRouter(doc)
	require.NoError(t, err)

	executor := &ProducerExecutor{
		openAPIDoc:    doc,
		openAPIRouter: router,
	}
	report := &ContractDiffReport{
		MissingFields:    []string{},
		ExtraFields:      []string{},
		TypeMismatches:   map[string]string{},
		ValueMismatches:  map[string]ValueMismatch{},
		HeaderMismatches: map[string]ValueMismatch{},
		ExpectedFields:   map[string]interface{}{},
		ActualFields:     map[string]interface{}{},
	}
	// /unknown-path is not in the spec
	err = executor.validateResponseAgainstSchema(
		context.Background(),
		"https://mocksite.local/unknown-path",
		"GET",
		200,
		[]byte(`{"foo":"bar"}`),
		http.Header{"Content-Type": []string{"application/json"}},
		report,
	)
	// Should not return an error — gracefully skips unknown routes
	require.NoError(t, err)
	require.Empty(t, report.SchemaViolations)
}

func Test_ValidateResponse_NilRouterIsNoop(t *testing.T) {
	executor := &ProducerExecutor{
		openAPIDoc:    nil,
		openAPIRouter: nil,
	}
	report := &ContractDiffReport{
		MissingFields:    []string{},
		ExtraFields:      []string{},
		TypeMismatches:   map[string]string{},
		ValueMismatches:  map[string]ValueMismatch{},
		HeaderMismatches: map[string]ValueMismatch{},
		ExpectedFields:   map[string]interface{}{},
		ActualFields:     map[string]interface{}{},
	}
	err := executor.validateResponseAgainstSchema(
		context.Background(),
		"https://mocksite.local/items",
		"GET",
		200,
		[]byte(`{}`),
		http.Header{},
		report,
	)
	require.NoError(t, err)
	require.Empty(t, report.SchemaViolations)
}

func Test_WithOpenAPISpec_SetsFields(t *testing.T) {
	doc := mustParseOAPIDoc(t, []byte(schemaValidationYAML))
	router, err := oapi.BuildRouter(doc)
	require.NoError(t, err)

	original := &ProducerExecutor{}
	result := original.WithOpenAPISpec(doc, router)
	// WithOpenAPISpec returns a new copy — original should be unchanged
	require.NotEqual(t, original, result, "WithOpenAPISpec should return a new copy, not the original")
	require.Nil(t, original.openAPIDoc, "original executor should not be modified")
	require.NotNil(t, result.openAPIDoc, "copy should have doc set")
	require.NotNil(t, result.openAPIRouter, "copy should have router set")
}

// mustParseOAPIDoc parses a YAML byte slice into an openapi3.T or fails the test.
func mustParseOAPIDoc(t *testing.T, data []byte) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(data)
	require.NoError(t, err)
	require.NoError(t, doc.Validate(context.Background()))
	return doc
}
