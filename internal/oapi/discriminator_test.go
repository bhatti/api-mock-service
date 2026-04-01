package oapi

import (
	"context"
	"strings"
	"testing"

	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// discriminatorYAML has oneOf + explicit mapping → should generate 2 specs
const discriminatorYAML = `
openapi: "3.0.3"
info:
  title: Animals API
  version: "1.0"
paths:
  /animals:
    post:
      operationId: createAnimal
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                animalType:
                  type: string
      responses:
        "200":
          description: Created
          content:
            application/json:
              schema:
                oneOf:
                  - $ref: '#/components/schemas/Cat'
                  - $ref: '#/components/schemas/Dog'
                discriminator:
                  propertyName: animalType
                  mapping:
                    cat: '#/components/schemas/Cat'
                    dog: '#/components/schemas/Dog'
components:
  schemas:
    Cat:
      type: object
      properties:
        animalType:
          type: string
        meow:
          type: string
    Dog:
      type: object
      properties:
        animalType:
          type: string
        bark:
          type: string
`

// noDiscriminatorYAML uses oneOf without a discriminator mapping
const noDiscriminatorYAML = `
openapi: "3.0.3"
info:
  title: Simple API
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
                properties:
                  id:
                    type: string
`

func Test_Discriminator_GeneratesOneScenarioPerVariant(t *testing.T) {
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, _, err := Parse(context.Background(), &types.Configuration{}, []byte(discriminatorYAML), dataTempl)
	require.NoError(t, err)
	require.NotEmpty(t, specs)
	// Should have 2 specs: one per variant (cat and dog)
	require.Equal(t, 2, len(specs), "expected one spec per discriminator variant, got %d", len(specs))
}

func Test_Discriminator_VariantNamesFromMapping(t *testing.T) {
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, _, err := Parse(context.Background(), &types.Configuration{}, []byte(discriminatorYAML), dataTempl)
	require.NoError(t, err)
	require.Equal(t, 2, len(specs))

	// Variant IDs should contain "cat" and "dog" from the discriminator mapping
	hascat, hasdog := false, false
	for _, s := range specs {
		if strings.Contains(s.ID, "cat") {
			hascat = true
		}
		if strings.Contains(s.ID, "dog") {
			hasdog = true
		}
	}
	require.True(t, hascat, "expected a 'cat' variant spec; got IDs: %v", specIDs(specs))
	require.True(t, hasdog, "expected a 'dog' variant spec; got IDs: %v", specIDs(specs))
}

func Test_Discriminator_FallsBackToSingleSpecWithoutDiscriminator(t *testing.T) {
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, _, err := Parse(context.Background(), &types.Configuration{}, []byte(noDiscriminatorYAML), dataTempl)
	require.NoError(t, err)
	require.NotEmpty(t, specs, "expected at least one spec for a plain schema")
}

func Test_DiscriminatorVariants_ReturnsNilForPlainSchema(t *testing.T) {
	schema := &openapi3.Schema{Type: "object"}
	require.Nil(t, DiscriminatorVariants(schema))
}

func Test_DiscriminatorVariants_AutoNamesVariants(t *testing.T) {
	catSchema := &openapi3.Schema{Type: "object"}
	dogSchema := &openapi3.Schema{Type: "object"}
	schema := &openapi3.Schema{
		OneOf: openapi3.SchemaRefs{
			{Value: catSchema},
			{Value: dogSchema},
		},
	}
	variants := DiscriminatorVariants(schema)
	require.Equal(t, 2, len(variants))
	require.Equal(t, "variant0", variants[0].Name)
	require.Equal(t, "variant1", variants[1].Name)
}

func specIDs(specs []*APISpec) []string {
	ids := make([]string, len(specs))
	for i, s := range specs {
		ids[i] = s.ID
	}
	return ids
}
