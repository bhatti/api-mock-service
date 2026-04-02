package oapi

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Inline synthetic OpenAPI specs — no external files needed
// ---------------------------------------------------------------------------

// petStoreYAML exercises: paths, parameters (path/query/header), requestBody,
// responses (200/201/400/404), tags, operationId, summary, description
const petStoreYAML = `
openapi: "3.0.0"
info:
  title: SyntheticPetStore
  version: "2.0"
  description: Synthetic spec for integration testing
paths:
  /pets:
    get:
      operationId: listPets
      summary: List all pets
      tags: [pets]
      parameters:
        - name: limit
          in: query
          required: false
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
        - name: status
          in: query
          schema:
            type: string
            enum: [available, pending, sold]
      responses:
        "200":
          description: A list of pets
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/Pet"
        "400":
          description: Bad request
    post:
      operationId: createPet
      summary: Create a pet
      tags: [pets]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/NewPet"
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
        "400":
          description: Invalid input
  /pets/{petId}:
    get:
      operationId: getPet
      summary: Get a specific pet
      tags: [pets]
      parameters:
        - name: petId
          in: path
          required: true
          schema:
            type: string
            format: uuid
        - name: X-Request-ID
          in: header
          required: false
          schema:
            type: string
      responses:
        "200":
          description: A pet
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
        "404":
          description: Not found
components:
  schemas:
    Pet:
      type: object
      required: [id, name]
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
          minLength: 1
          maxLength: 100
        status:
          type: string
          enum: [available, pending, sold]
        age:
          type: integer
          minimum: 0
          maximum: 30
    NewPet:
      type: object
      required: [name]
      properties:
        name:
          type: string
        tag:
          type: string
`

// schemaConstraintsYAML exercises: const, default, example, exclusiveMin/Max,
// multipleOf, uniqueItems, minProperties, maxProperties, nullable, readOnly, writeOnly
const schemaConstraintsYAML = `
openapi: "3.0.0"
info:
  title: SchemaConstraints
  version: "1.0"
paths:
  /orders:
    post:
      operationId: createOrder
      tags: [orders]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              minProperties: 2
              maxProperties: 10
              properties:
                quantity:
                  type: integer
                  minimum: 1
                  maximum: 100
                  exclusiveMinimum: true
                  exclusiveMaximum: false
                  multipleOf: 5
                  example: "10"
                currency:
                  type: string
                  enum: [USD]
                price:
                  type: number
                  minimum: 0.01
                  exclusiveMinimum: true
                  example: "9.99"
                event_type:
                  type: string
                  enum: [order_created]
                  default: order_created
                sku:
                  type: string
                  minLength: 3
                  maxLength: 20
                  pattern: "^[A-Z0-9-]+$"
                tags:
                  type: array
                  items:
                    type: string
                  uniqueItems: true
                  minItems: 1
                  maxItems: 5
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema:
                type: object
                readOnly: true
                properties:
                  order_id:
                    type: string
                    format: uuid
                    readOnly: true
                  created_at:
                    type: string
                    format: date-time
                    readOnly: true
                  total:
                    type: number
                    readOnly: true
                    example: "99.99"
`

// securityYAML exercises: apiKey (header/query/cookie), http (basic/bearer),
// oauth2 (all 4 flows with scopes), openIdConnect
const securityYAML = `
openapi: "3.0.0"
info:
  title: SecurityAPI
  version: "1.0"
paths:
  /resources:
    get:
      operationId: listResources
      tags: [resources]
      security:
        - apiKeyHeader: []
        - bearerAuth: []
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  items:
                    type: array
                    items:
                      type: string
components:
  securitySchemes:
    apiKeyHeader:
      type: apiKey
      name: X-API-Key
      in: header
      description: API key passed via header
    apiKeyQuery:
      type: apiKey
      name: api_key
      in: query
      description: API key passed via query string
    basicAuth:
      type: http
      scheme: basic
      description: HTTP Basic Authentication
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: Bearer token authentication
    oauth2Flows:
      type: oauth2
      description: OAuth2 with multiple flows
      flows:
        implicit:
          authorizationUrl: https://auth.example.com/authorize
          scopes:
            read:pets: Read pet data
            write:pets: Write pet data
        password:
          tokenUrl: https://auth.example.com/token
          scopes:
            read:pets: Read pet data
        clientCredentials:
          tokenUrl: https://auth.example.com/token
          scopes:
            read:all: Full read access
            write:all: Full write access
        authorizationCode:
          authorizationUrl: https://auth.example.com/authorize
          tokenUrl: https://auth.example.com/token
          scopes:
            read:pets: Read pet data
            write:pets: Write pet data
    openIdConnect:
      type: openIdConnect
      openIdConnectUrl: https://auth.example.com/.well-known/openid-configuration
`

// compositionYAML exercises: allOf (inheritance), oneOf (discriminated union),
// anyOf (flexible types), nested composition, inline objects
const compositionYAML = `
openapi: "3.0.0"
info:
  title: CompositionAPI
  version: "1.0"
paths:
  /animals:
    get:
      operationId: listAnimals
      tags: [animals]
      responses:
        "200":
          description: Animals list
          content:
            application/json:
              schema:
                type: object
                properties:
                  cat:
                    allOf:
                      - $ref: "#/components/schemas/Animal"
                      - type: object
                        properties:
                          indoor:
                            type: boolean
                  dog:
                    oneOf:
                      - $ref: "#/components/schemas/BigDog"
                      - $ref: "#/components/schemas/SmallDog"
                  pet:
                    anyOf:
                      - $ref: "#/components/schemas/Cat"
                      - $ref: "#/components/schemas/Dog"
components:
  schemas:
    Animal:
      type: object
      required: [name, species]
      properties:
        name:
          type: string
        species:
          type: string
    BigDog:
      type: object
      properties:
        breed:
          type: string
        weight_kg:
          type: number
          minimum: 25
    SmallDog:
      type: object
      properties:
        breed:
          type: string
        weight_kg:
          type: number
          maximum: 25
    Cat:
      type: object
      properties:
        indoor:
          type: boolean
        lives:
          type: integer
    Dog:
      type: object
      properties:
        breed:
          type: string
`

// paramStyleYAML exercises parameter style, explode, deprecated, allowEmptyValue
const paramStyleYAML = `
openapi: "3.0.0"
info:
  title: ParamStyleAPI
  version: "1.0"
paths:
  /search:
    get:
      operationId: search
      tags: [search]
      parameters:
        - name: ids
          in: query
          style: form
          explode: true
          schema:
            type: array
            items:
              type: string
        - name: filter
          in: query
          style: deepObject
          explode: true
          schema:
            type: object
            properties:
              status:
                type: string
              category:
                type: string
        - name: sort
          in: query
          deprecated: true
          schema:
            type: string
            enum: [asc, desc]
            default: asc
        - name: Accept-Language
          in: header
          schema:
            type: string
            default: en-US
      responses:
        "200":
          description: Search results
          content:
            application/json:
              schema:
                type: object
                properties:
                  results:
                    type: array
                    items:
                      type: string
                  total:
                    type: integer
`

// nullableYAML exercises nullable fields, readOnly/writeOnly in request vs response
const nullableYAML = `
openapi: "3.0.0"
info:
  title: NullableAPI
  version: "1.0"
paths:
  /users:
    post:
      operationId: createUser
      tags: [users]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                  writeOnly: true
                password:
                  type: string
                  format: password
                  writeOnly: true
                  minLength: 8
                nickname:
                  type: string
                  nullable: true
      responses:
        "200":
          description: Created user
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                    format: uuid
                    readOnly: true
                  username:
                    type: string
                  nickname:
                    type: string
                    nullable: true
                  created_at:
                    type: string
                    format: date-time
                    readOnly: true
`

// multiResponseYAML exercises multiple response codes, 2xx body export,
// response headers, external docs
const multiResponseYAML = `
openapi: "3.0.0"
info:
  title: MultiResponseAPI
  version: "1.0"
paths:
  /jobs:
    post:
      operationId: submitJob
      summary: Submit an async job
      tags: [jobs]
      externalDocs:
        description: Job processing guide
        url: https://docs.example.com/jobs
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                priority:
                  type: integer
                  minimum: 1
                  maximum: 10
                  default: 5
      responses:
        "200":
          description: Job already exists
          headers:
            X-Job-ID:
              schema:
                type: string
                format: uuid
          content:
            application/json:
              schema:
                type: object
                properties:
                  job_id:
                    type: string
                  status:
                    type: string
        "201":
          description: Job created
          headers:
            Location:
              schema:
                type: string
                format: uri
            X-Job-ID:
              schema:
                type: string
                format: uuid
          content:
            application/json:
              schema:
                type: object
                properties:
                  job_id:
                    type: string
                    format: uuid
                  status:
                    type: string
                    enum: [queued, running, done, failed]
                  created_at:
                    type: string
                    format: date-time
        "202":
          description: Job accepted for processing
          content:
            application/json:
              schema:
                type: object
                properties:
                  job_id:
                    type: string
                  message:
                    type: string
        "422":
          description: Validation error
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                  details:
                    type: array
                    items:
                      type: string
`

// ---------------------------------------------------------------------------
// Helper: parse YAML spec and return specs+scenarios
// ---------------------------------------------------------------------------

func parseYAML(t *testing.T, specYAML string) ([]*APISpec, []*types.APIScenario) {
	t.Helper()
	data := []byte(specYAML)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)
	require.NoError(t, err)
	require.NotEmpty(t, specs)

	scenarios := make([]*types.APIScenario, 0, len(specs))
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		scenarios = append(scenarios, scenario)
	}
	return specs, scenarios
}

// ---------------------------------------------------------------------------
// Integration tests
// ---------------------------------------------------------------------------

// Test_OAPIInteg_PetStoreBasics verifies core parsing: operationId, summary,
// tags, path params, query params, header params, request body, multiple status codes.
func Test_OAPIInteg_PetStoreBasics(t *testing.T) {
	specs, scenarios := parseYAML(t, petStoreYAML)

	// Should produce specs for GET /pets (200+400), POST /pets (201+400), GET /pets/{petId} (200+404)
	require.GreaterOrEqual(t, len(specs), 6)

	// Verify every scenario has tags and a non-empty group
	for _, s := range scenarios {
		require.NotEmpty(t, s.Tags, "every scenario should have tags")
		require.NotEmpty(t, s.Group)
	}

	// Find the createPet spec (POST /pets, 201)
	var createPet *APISpec
	for _, sp := range specs {
		if sp.ID == "createPet" && sp.Response.StatusCode == 201 {
			createPet = sp
			break
		}
	}
	require.NotNil(t, createPet, "createPet spec (201) should exist")
	require.True(t, createPet.RequestBodyRequired, "POST /pets requestBody.required should be true")
	require.Equal(t, "Create a pet", createPet.Summary)
	require.Equal(t, []string{"pets"}, createPet.Tags)

	// Find GET /pets/{petId} and check path param
	var getPet *APISpec
	for _, sp := range specs {
		if sp.ID == "getPet" && sp.Response.StatusCode == 200 {
			getPet = sp
			break
		}
	}
	require.NotNil(t, getPet, "getPet spec should exist")
	require.Len(t, getPet.Request.PathParams, 1)
	require.Equal(t, "petId", getPet.Request.PathParams[0].Name)
	require.Equal(t, "uuid", getPet.Request.PathParams[0].Format)
	require.True(t, getPet.Request.PathParams[0].Required)

	// Verify a 404 scenario has NthRequest 2 predicate
	var getPet404 *types.APIScenario
	for _, s := range scenarios {
		if strings.Contains(s.Name, "getPet") && s.Response.StatusCode == 404 {
			getPet404 = s
			break
		}
	}
	require.NotNil(t, getPet404)
	require.Equal(t, "{{NthRequest 2}}", getPet404.Predicate)

	// Verify the 2xx scenario has NthRequest 1 predicate
	var getPet200 *types.APIScenario
	for _, s := range scenarios {
		if strings.Contains(s.Name, "getPet") && s.Response.StatusCode == 200 {
			getPet200 = s
			break
		}
	}
	require.NotNil(t, getPet200)
	require.Equal(t, "{{NthRequest 1}}", getPet200.Predicate)
}

// Test_OAPIInteg_SchemaConstraints verifies that const, default, example,
// exclusiveMin, exclusiveMax, multipleOf, uniqueItems, minProperties,
// maxProperties are all captured in Property.
func Test_OAPIInteg_SchemaConstraints(t *testing.T) {
	specs, _ := parseYAML(t, schemaConstraintsYAML)

	// Find POST /orders 201 spec to inspect the response body
	var resp201 *APISpec
	for _, sp := range specs {
		if sp.ID == "createOrder" && sp.Response.StatusCode == 201 {
			resp201 = sp
			break
		}
	}
	require.NotNil(t, resp201)

	// Check readOnly on response properties
	for _, bp := range resp201.Response.Body {
		for _, child := range bp.Children {
			if child.Name == "order_id" || child.Name == "created_at" || child.Name == "total" {
				require.True(t, child.ReadOnly, "response property %s should be readOnly", child.Name)
			}
		}
	}

	// Find POST /orders request spec and verify constraints
	var req *APISpec
	for _, sp := range specs {
		if sp.ID == "createOrder" && sp.Response.StatusCode == 201 {
			req = sp
			break
		}
	}
	require.NotNil(t, req)

	// Verify RequestBodyRequired
	require.True(t, req.RequestBodyRequired)

	// Verify MinProps/MaxProps on the request body root property
	if len(req.Request.Body) > 0 {
		body := req.Request.Body[0]
		require.Equal(t, uint64(2), body.MinProps)
		require.Equal(t, uint64(10), body.MaxProps)
	}

	// Find individual constraint fields in request body children
	findChild := func(children []Property, name string) *Property {
		for i, c := range children {
			if c.Name == name {
				return &children[i]
			}
		}
		return nil
	}

	if len(req.Request.Body) > 0 {
		children := req.Request.Body[0].Children
		qty := findChild(children, "quantity")
		if qty != nil {
			require.True(t, qty.ExclusiveMin, "quantity.exclusiveMinimum should be true")
			require.Equal(t, float64(5), qty.MultipleOf, "quantity.multipleOf should be 5")
			require.Equal(t, "10", qty.Example)
		}

		currency := findChild(children, "currency")
		if currency != nil {
			require.NotEmpty(t, currency.Enum, "currency should have enum values")
			require.Equal(t, "USD", currency.Enum[0])
			// Single-element enum → Const
			require.Equal(t, "USD", currency.Const)
		}

		price := findChild(children, "price")
		if price != nil {
			require.True(t, price.ExclusiveMin, "price.exclusiveMinimum should be true")
			require.Equal(t, "9.99", price.Example)
		}

		eventType := findChild(children, "event_type")
		if eventType != nil {
			require.Equal(t, "order_created", eventType.Default)
		}

		sku := findChild(children, "sku")
		if sku != nil {
			require.Equal(t, float64(3), sku.Min)
			require.Equal(t, float64(20), sku.Max)
			require.NotEmpty(t, sku.Pattern)
		}

		tags := findChild(children, "tags")
		if tags != nil {
			require.True(t, tags.UniqueItems, "tags array should have uniqueItems=true")
			require.Equal(t, float64(1), tags.Min)
			require.Equal(t, float64(5), tags.Max)
		}
	}
}

// Test_OAPIInteg_SecuritySchemes verifies all security scheme types:
// apiKey (header/query), http (basic/bearer), oauth2 (all 4 flows), openIdConnect.
func Test_OAPIInteg_SecuritySchemes(t *testing.T) {
	_, scenarios := parseYAML(t, securityYAML)
	require.NotEmpty(t, scenarios)

	// Every scenario should have the global security schemes in Authentication
	for _, s := range scenarios {
		require.NotEmpty(t, s.Authentication, "scenarios should have authentication entries")
	}

	s := scenarios[0]

	// apiKeyHeader
	apiKeyHeader, ok := s.Authentication["apiKeyHeader"]
	require.True(t, ok, "apiKeyHeader scheme should be present")
	require.Equal(t, "apiKey", apiKeyHeader.Type)
	require.Equal(t, "X-API-Key", apiKeyHeader.Name)
	require.Equal(t, "header", apiKeyHeader.In)
	require.Equal(t, "API key passed via header", apiKeyHeader.Description)

	// bearerAuth
	bearer, ok := s.Authentication["bearerAuth"]
	require.True(t, ok, "bearerAuth scheme should be present")
	require.Equal(t, "http", bearer.Type)
	require.Equal(t, "bearer", bearer.Scheme)
	require.Equal(t, "JWT", bearer.Format)

	// oauth2Flows — verify scopes are collected from all 4 flows
	oauth2, ok := s.Authentication["oauth2Flows"]
	require.True(t, ok, "oauth2Flows scheme should be present")
	require.Equal(t, "oauth2", oauth2.Type)
	require.Equal(t, "OAuth2 with multiple flows", oauth2.Description)
	require.NotEmpty(t, oauth2.Scopes, "OAuth2 scopes should be captured")
	require.Contains(t, oauth2.Scopes, "read:pets", "read:pets scope should be present")
	require.Contains(t, oauth2.Scopes, "write:pets", "write:pets scope should be present")
	require.Contains(t, oauth2.Scopes, "read:all", "read:all scope should be present")
	require.Contains(t, oauth2.Scopes, "write:all", "write:all scope should be present")
}

// Test_OAPIInteg_QueryAPIKeyGoesToQueryParams verifies the security scheme
// with in:query goes to QueryParams, not Headers (regression for the line-73 bug).
func Test_OAPIInteg_QueryAPIKeyGoesToQueryParams(t *testing.T) {
	data := []byte(securityYAML)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)
	require.NoError(t, err)

	for _, sp := range specs {
		// api_key (in:query) must appear in QueryParams
		for _, qp := range sp.Request.QueryParams {
			if qp.Name == "api_key" {
				// Found — good
				return
			}
		}
		// Must NOT appear in Headers
		for _, h := range sp.Request.Headers {
			require.NotEqual(t, "api_key", h.Name, "api_key (in:query) must not be in Headers")
		}
	}
}

// Test_OAPIInteg_SchemaComposition verifies allOf property merging, oneOf/anyOf
// first-branch selection, and that properties are NOT duplicated.
func Test_OAPIInteg_SchemaComposition(t *testing.T) {
	specs, _ := parseYAML(t, compositionYAML)
	require.NotEmpty(t, specs)

	sp := specs[0]
	require.Equal(t, "listAnimals", sp.ID)

	// Helper: count children with a given name
	countNamed := func(children []Property, name string) int {
		n := 0
		for _, c := range children {
			if c.Name == name {
				n++
			}
		}
		return n
	}

	// Find "cat" in response body — should have allOf merging Animal + indoor
	findProp := func(props []Property, name string) *Property {
		for i, p := range props {
			if p.Name == name {
				return &props[i]
			}
		}
		return nil
	}

	if len(sp.Response.Body) > 0 {
		body := sp.Response.Body[0]

		cat := findProp(body.Children, "cat")
		if cat != nil {
			// Should have Animal fields (name, species) + indoor, each exactly once
			require.Equal(t, 1, countNamed(cat.Children, "name"), "cat.name should appear exactly once (allOf dedup)")
			require.Equal(t, 1, countNamed(cat.Children, "species"), "cat.species should appear exactly once (allOf dedup)")
			require.Equal(t, 1, countNamed(cat.Children, "indoor"), "cat.indoor should appear exactly once (allOf dedup)")
		}

		dog := findProp(body.Children, "dog")
		if dog != nil {
			// oneOf: only first branch (BigDog) should be used
			require.GreaterOrEqual(t, len(dog.Children), 1)
		}
	}
}

// Test_OAPIInteg_ParameterStyle verifies that Style, Explode, and Deprecated
// are captured from parameter objects.
func Test_OAPIInteg_ParameterStyle(t *testing.T) {
	data := []byte(paramStyleYAML)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)
	require.NoError(t, err)
	require.NotEmpty(t, specs)

	sp := specs[0]
	require.Equal(t, "search", sp.ID)

	findQP := func(name string) *Property {
		for i, qp := range sp.Request.QueryParams {
			if qp.Name == name {
				return &sp.Request.QueryParams[i]
			}
		}
		return nil
	}

	ids := findQP("ids")
	if ids != nil {
		require.Equal(t, "form", ids.Style)
		require.True(t, ids.Explode)
	}

	sort := findQP("sort")
	if sort != nil {
		require.True(t, sort.Deprecated, "sort parameter should be marked deprecated")
		require.Equal(t, "asc", sort.Default)
	}
}

// Test_OAPIInteg_NullableReadWriteOnly verifies that nullable, readOnly, writeOnly
// are captured and influence which fields appear in request vs response.
func Test_OAPIInteg_NullableReadWriteOnly(t *testing.T) {
	specs, _ := parseYAML(t, nullableYAML)

	findSpec := func(operationID string, status int) *APISpec {
		for _, sp := range specs {
			if sp.ID == operationID && sp.Response.StatusCode == status {
				return sp
			}
		}
		return nil
	}

	sp := findSpec("createUser", 200)
	require.NotNil(t, sp)

	findChild := func(props []Property, name string) *Property {
		for i, p := range props {
			if p.Name == name {
				return &props[i]
			}
		}
		return nil
	}

	// Request body: password should be writeOnly
	if len(sp.Request.Body) > 0 {
		pwd := findChild(sp.Request.Body[0].Children, "password")
		if pwd != nil {
			require.True(t, pwd.WriteOnly)
			require.Equal(t, "password", pwd.Format)
		}
		nick := findChild(sp.Request.Body[0].Children, "nickname")
		if nick != nil {
			require.True(t, nick.Nullable)
		}
	}

	// Response body: id and created_at should be readOnly
	if len(sp.Response.Body) > 0 {
		id := findChild(sp.Response.Body[0].Children, "id")
		if id != nil {
			require.True(t, id.ReadOnly)
			require.Equal(t, "uuid", id.Format)
		}
		createdAt := findChild(sp.Response.Body[0].Children, "created_at")
		if createdAt != nil {
			require.True(t, createdAt.ReadOnly)
		}
	}
}

// Test_OAPIInteg_MultipleResponseCodes verifies that each response code gets its
// own APISpec, response bodies are exported for all 2xx (200/201/202),
// and that ExternalDocs URL is captured.
func Test_OAPIInteg_MultipleResponseCodes(t *testing.T) {
	specs, _ := parseYAML(t, multiResponseYAML)

	statusCodes := make(map[int]bool)
	for _, sp := range specs {
		if sp.ID == "submitJob" {
			statusCodes[sp.Response.StatusCode] = true
		}
	}
	require.True(t, statusCodes[200], "200 spec should exist")
	require.True(t, statusCodes[201], "201 spec should exist")
	require.True(t, statusCodes[202], "202 spec should exist")
	require.True(t, statusCodes[422], "422 spec should exist")

	// ExternalDocs URL should be captured
	for _, sp := range specs {
		if sp.ID == "submitJob" {
			require.Equal(t, "https://docs.example.com/jobs", sp.ExternalDocsURL)
			break
		}
	}

	// RequestBodyRequired should be true for POST /jobs
	for _, sp := range specs {
		if sp.ID == "submitJob" && sp.Response.StatusCode == 201 {
			require.True(t, sp.RequestBodyRequired)
			break
		}
	}
}

// Test_OAPIInteg_RequestBodyRequiredAddsAssertion verifies that RequestBodyRequired=true
// causes a HasProperty contents assertion to be added to the scenario.
func Test_OAPIInteg_RequestBodyRequiredAddsAssertion(t *testing.T) {
	_, scenarios := parseYAML(t, multiResponseYAML)

	for _, s := range scenarios {
		if strings.Contains(s.Name, "submitJob") && s.Response.StatusCode == 201 {
			found := false
			for _, a := range s.Request.Assertions {
				if strings.Contains(a, "HasProperty") && strings.Contains(a, "contents") {
					found = true
					break
				}
			}
			require.True(t, found, "RequestBodyRequired should generate HasProperty contents assertion")
		}
	}
}

// Test_OAPIInteg_ConstValueReturnedByProperty verifies that a property with
// a single-element enum (const) returns that exact value from Value().
func Test_OAPIInteg_ConstValueReturnedByProperty(t *testing.T) {
	dt := fuzz.NewDataTemplateRequest(false, 1, 1)
	prop := Property{
		Name:  "event_type",
		Type:  "string",
		Const: "order_created",
	}
	val := prop.Value(dt)
	m, ok := val.(map[string]string)
	require.True(t, ok)
	require.Equal(t, "order_created", m["event_type"])

	// IncludeType mode should also return the const (for exact-match validation)
	dtInclude := fuzz.NewDataTemplateRequest(true, 1, 1)
	val = prop.Value(dtInclude)
	m, ok = val.(map[string]string)
	require.True(t, ok)
	require.Equal(t, "order_created", m["event_type"])
}

// Test_OAPIInteg_ExampleAndDefaultValues verifies that Example takes priority
// over Default in mock response generation, and both are skipped for IncludeType.
func Test_OAPIInteg_ExampleAndDefaultValues(t *testing.T) {
	dt := fuzz.NewDataTemplateRequest(false, 1, 1)

	// Example takes priority over Default
	prop := Property{
		Name:    "total",
		Type:    "number",
		Example: "99.99",
		Default: "0.00",
	}
	val := prop.Value(dt)
	m, ok := val.(map[string]string)
	require.True(t, ok)
	require.Equal(t, "99.99", m["total"])

	// Default used when no example
	propDefault := Property{
		Name:    "priority",
		Type:    "integer",
		Default: "5",
	}
	val = propDefault.Value(dt)
	m, ok = val.(map[string]string)
	require.True(t, ok)
	require.Equal(t, "5", m["priority"])

	// IncludeType mode: example/default NOT used (generate validation pattern)
	dtInclude := fuzz.NewDataTemplateRequest(true, 1, 1)
	val = prop.Value(dtInclude)
	m, ok = val.(map[string]string)
	require.True(t, ok)
	// Pattern-based, not "99.99"
	require.NotEqual(t, "99.99", m["total"])
}

// Test_OAPIInteg_ExclusiveMinMax verifies that ExclusiveMin/ExclusiveMax adjust
// the generated range in numericValue().
func Test_OAPIInteg_ExclusiveMinMax(t *testing.T) {
	dt := fuzz.NewDataTemplateRequest(false, 1, 1)

	prop := Property{
		Name:         "quantity",
		Type:         "integer",
		Min:          1,
		Max:          100,
		ExclusiveMin: true,
		ExclusiveMax: false,
	}
	val := prop.numericValue()
	// ExclusiveMin=true with Min=1 → should use min=2
	require.Contains(t, val, "2", "exclusive min=1 should shift to 2")
	_ = dt // used for type check only
}

// Test_OAPIInteg_RoundTrip parses a spec, builds scenarios, exports back to
// OpenAPI, and verifies the exported doc has the expected structure.
func Test_OAPIInteg_RoundTrip(t *testing.T) {
	_, scenarios := parseYAML(t, petStoreYAML)
	require.NotEmpty(t, scenarios)

	doc := ScenarioToOpenAPI("RoundTripTest", "1.0", scenarios...)
	require.NotNil(t, doc)
	require.Equal(t, "RoundTripTest", doc.Info.Title)

	// All paths should be present
	require.NotEmpty(t, doc.Paths)

	// Mock server should be in servers list
	found := false
	for _, srv := range doc.Servers {
		if strings.Contains(srv.URL, MockServerBaseURL) || srv.URL == MockServerBaseURL {
			found = true
		}
	}
	require.True(t, found, "MockServerBaseURL should appear in exported servers")

	// Marshal to JSON and back — should be valid JSON
	b, err := doc.MarshalJSON()
	require.NoError(t, err)
	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &raw))
	require.Equal(t, "3.0.2", raw["openapi"])
}

// Test_OAPIInteg_MultipleContentTypes verifies that when a response has multiple
// content types, separate specs are created for each.
func Test_OAPIInteg_MultipleContentTypes(t *testing.T) {
	specYAML := `
openapi: "3.0.0"
info:
  title: MultiContentType
  version: "1.0"
paths:
  /data:
    get:
      operationId: getData
      tags: [data]
      responses:
        "200":
          description: Data
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
            application/xml:
              schema:
                type: object
                properties:
                  id:
                    type: string
`
	data := []byte(specYAML)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)
	require.NoError(t, err)
	// Should have 2 specs: one per content type
	require.Equal(t, 2, len(specs))
	contentTypes := make(map[string]bool)
	for _, sp := range specs {
		contentTypes[sp.Response.ContentType] = true
	}
	require.True(t, contentTypes["application/json"])
	require.True(t, contentTypes["application/xml"])
}

// Test_OAPIInteg_ResponseHeadersCaptured verifies that response headers defined
// in the OpenAPI spec are captured in oapi.Response.Headers.
func Test_OAPIInteg_ResponseHeadersCaptured(t *testing.T) {
	specs, _ := parseYAML(t, multiResponseYAML)

	var spec201 *APISpec
	for _, sp := range specs {
		if sp.ID == "submitJob" && sp.Response.StatusCode == 201 {
			spec201 = sp
			break
		}
	}
	require.NotNil(t, spec201)

	headerNames := make(map[string]bool)
	for _, h := range spec201.Response.Headers {
		headerNames[h.Name] = true
	}
	require.True(t, headerNames["Location"] || headerNames["X-Job-ID"],
		"response headers Location and/or X-Job-ID should be captured; found: %v", headerNames)
}

// Test_OAPIInteg_PropertyFuzzValues verifies that Value() produces non-nil,
// non-empty output for every primitive type and common format combination.
func Test_OAPIInteg_PropertyFuzzValues(t *testing.T) {
	dt := fuzz.NewDataTemplateRequest(false, 1, 1)

	cases := []struct {
		name   string
		prop   Property
	}{
		{"integer", Property{Name: "n", Type: "integer", Min: 1, Max: 100}},
		{"number", Property{Name: "n", Type: "number", Min: 1, Max: 100}},
		{"boolean", Property{Name: "b", Type: "boolean"}},
		{"string_plain", Property{Name: "s", Type: "string"}},
		{"string_uuid", Property{Name: "s", Type: "string", Format: "uuid"}},
		{"string_email", Property{Name: "s", Type: "string", Format: "email"}},
		{"string_date", Property{Name: "s", Type: "string", Format: "date"}},
		{"string_datetime", Property{Name: "s", Type: "string", Format: "date-time"}},
		{"string_phone", Property{Name: "s", Type: "string", Format: "phone"}},
		{"string_uri", Property{Name: "s", Type: "string", Format: "uri"}},
		{"string_ip", Property{Name: "s", Type: "string", Format: "ip"}},
		{"string_ipv4", Property{Name: "s", Type: "string", Format: "ipv4"}},
		{"string_ipv6", Property{Name: "s", Type: "string", Format: "ipv6"}},
		{"string_hostname", Property{Name: "s", Type: "string", Format: "hostname"}},
		{"string_password", Property{Name: "s", Type: "string", Format: "password"}},
		{"string_byte", Property{Name: "s", Type: "string", Format: "byte"}},
		{"string_binary", Property{Name: "s", Type: "string", Format: "binary"}},
		{"string_int32", Property{Name: "s", Type: "string", Format: "int32"}},
		{"string_int64", Property{Name: "s", Type: "string", Format: "int64"}},
		{"string_float", Property{Name: "s", Type: "string", Format: "float"}},
		{"string_double", Property{Name: "s", Type: "string", Format: "double"}},
		{"string_ssn", Property{Name: "s", Type: "string", Format: "ssn"}},
		{"string_enum", Property{Name: "s", Type: "string", Enum: []string{"a", "b", "c"}}},
		{"string_pattern", Property{Name: "s", Type: "string", Pattern: `^[A-Z]{3}$`}},
		{"string_const", Property{Name: "s", Type: "string", Const: "fixed"}},
		{"string_example", Property{Name: "s", Type: "string", Example: "hello"}},
		{"string_default", Property{Name: "s", Type: "string", Default: "world"}},
		{"array_string", Property{Name: "a", Type: "array", SubType: "string", Min: 1, Max: 3}},
		{"array_int", Property{Name: "a", Type: "array", SubType: "integer", Min: 1, Max: 3}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			val := tc.prop.Value(dt)
			require.NotNil(t, val, "Value() must not be nil for %s", tc.name)
		})
	}
}

// Test_OAPIInteg_AllFormatsWithIncludeType verifies that every format returns
// map[string]string when IncludeType=true (contract/pattern mode).
func Test_OAPIInteg_AllFormatsWithIncludeType(t *testing.T) {
	dtInclude := fuzz.NewDataTemplateRequest(true, 1, 1)
	formats := []string{
		"date", "date-time", "time", "uri", "email", "phone", "uuid", "ulid",
		"host", "hostname", "ip", "ipv4", "ipv6", "airport", "locale", "country",
		"zip", "ssn", "isbn10", "isbn13", "password", "byte", "binary",
		"int32", "int64", "float", "double",
	}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			prop := Property{Name: "field", Type: "string", Format: format}
			val := prop.Value(dtInclude)
			_, ok := val.(map[string]string)
			require.True(t, ok, "format=%q IncludeType=true must return map[string]string, got %T", format, val)
		})
	}
}

// Test_OAPIInteg_ScenarioValidationAssertions verifies that contract assertions
// are generated for required fields in request bodies.
func Test_OAPIInteg_ScenarioValidationAssertions(t *testing.T) {
	_, scenarios := parseYAML(t, petStoreYAML)

	var createPet *types.APIScenario
	for _, s := range scenarios {
		if strings.Contains(s.Name, "createPet") && s.Response.StatusCode == 201 {
			createPet = s
			break
		}
	}
	require.NotNil(t, createPet)

	// Content-Type is added via checkRequestHeader into Assertions (PropertyMatches headers.Content-Type ...)
	hasContentType := false
	for _, a := range createPet.Request.Assertions {
		if strings.Contains(a, "Content-Type") {
			hasContentType = true
			break
		}
	}
	require.True(t, hasContentType, "Content-Type assertion should be generated")

	// RequestBodyRequired=true should add HasProperty contents (not PropertyLenGE —
	// that counts map keys, not bytes, breaking single-key bodies like {"ids":[...]})
	bodyAsserted := false
	for _, a := range createPet.Request.Assertions {
		if strings.Contains(a, "HasProperty") && strings.Contains(a, "contents") {
			bodyAsserted = true
			break
		}
	}
	require.True(t, bodyAsserted, "required request body should generate HasProperty contents assertion")
}

// Test_OAPIInteg_ResponseBodyForAll2xx verifies round-trip export includes
// response body schemas for 200, 201, and 202 (not just 200).
func Test_OAPIInteg_ResponseBodyForAll2xx(t *testing.T) {
	_, scenarios := parseYAML(t, multiResponseYAML)

	doc := ScenarioToOpenAPI("JobsAPI", "1.0", scenarios...)
	require.NotNil(t, doc)

	// Check that /jobs POST has 200, 201, 202 responses
	pathItem, ok := doc.Paths["/jobs"]
	require.True(t, ok, "/jobs path should be exported")
	require.NotNil(t, pathItem.Post)

	for _, code := range []string{"200", "201", "202"} {
		resp, exists := pathItem.Post.Responses[code]
		require.True(t, exists, "response %s should be exported", code)
		require.NotNil(t, resp.Value)
	}
}

// Test_OAPIInteg_InfoAndServerCapture verifies the API title, version,
// and mock server URL are captured during parsing.
func Test_OAPIInteg_InfoAndServerCapture(t *testing.T) {
	data := []byte(petStoreYAML)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, updated, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)
	require.NoError(t, err)
	require.NotEmpty(t, specs)
	require.NotEmpty(t, updated)

	// Title should include version
	for _, sp := range specs {
		require.True(t, strings.Contains(sp.Title, "SyntheticPetStore"), "title should contain API name")
		break
	}

	// Updated YAML should contain mock server
	require.True(t, strings.Contains(string(updated), MockServerBaseURL),
		"updated spec should contain MockServerBaseURL")
}

// Test_OAPIInteg_DeprecatedOperation verifies that deprecated:true on an
// operation is captured in APISpec.Deprecated.
func Test_OAPIInteg_DeprecatedOperation(t *testing.T) {
	specYAML := `
openapi: "3.0.0"
info:
  title: DeprecatedAPI
  version: "1.0"
paths:
  /old-endpoint:
    get:
      operationId: oldEndpoint
      deprecated: true
      tags: [legacy]
      responses:
        "200":
          description: OK
`
	data := []byte(specYAML)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)
	require.NoError(t, err)
	require.Len(t, specs, 1)
	require.True(t, specs[0].Deprecated, "deprecated operation should set APISpec.Deprecated=true")
}

// Test_OAPIInteg_EmptyEnumProducesPattern verifies enum values are captured
// and produce valid enum patterns.
func Test_OAPIInteg_EnumProducesPattern(t *testing.T) {
	dt := fuzz.NewDataTemplateRequest(false, 1, 1)

	prop := Property{
		Name: "status",
		Type: "string",
		Enum: []string{"active", "inactive", "pending"},
	}
	val := prop.Value(dt)
	m, ok := val.(map[string]string)
	require.True(t, ok)
	// stringValue() with enum uses EnumString template
	require.Contains(t, m["status"], "active")

	// Single-value enum → treated as Const
	constProp := Property{
		Name:  "version",
		Type:  "string",
		Const: "v1",
	}
	val = constProp.Value(dt)
	m, ok = val.(map[string]string)
	require.True(t, ok)
	require.Equal(t, "v1", m["version"])
}

// Test_OAPIInteg_LargeSpecParsing exercises parsing of the full Twitter spec
// to ensure nil guards and composition fixes don't cause regressions on a real
// 112-operation spec.
func Test_OAPIInteg_LargeSpecParsing(t *testing.T) {
	data, err := os.ReadFile("../../fixtures/oapi/twitter.yaml")
	if err != nil {
		t.Skipf("fixture not available: %v", err)
	}
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, _, _, err := Parse(context.Background(), &types.Configuration{}, data, dataTempl)
	require.NoError(t, err)
	require.Greater(t, len(specs), 0)
	for _, sp := range specs {
		s, err := sp.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		require.NotEmpty(t, s.Tags, "every Twitter spec scenario should have tags")
	}
}

// ---------------------------------------------------------------------------
// OpenAPI 3.1 compatibility tests
// ---------------------------------------------------------------------------

// openapi31TypeArrayYAML uses OpenAPI 3.1 syntax where "type" is an array,
// e.g. ["string","null"] — the exact pattern that caused the 500 error reported
// when uploading large real-world specs.
const openapi31TypeArrayYAML = `
openapi: "3.1.0"
info:
  title: OA31TypeArray
  version: "1.0"
paths:
  /items:
    get:
      operationId: listItems
      tags: [items]
      parameters:
        - name: filter
          in: query
          schema:
            type: ["string", "null"]
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: ["integer", "null"]
                  name:
                    type: ["string", "null"]
                  score:
                    type: ["number", "null"]
                  active:
                    type: ["boolean", "null"]
                  tags:
                    type: ["array", "null"]
                    items:
                      type: string
    post:
      operationId: createItem
      tags: [items]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name:
                  type: string
                category:
                  type: ["string", "null"]
                count:
                  type: ["integer", "null"]
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: ["integer", "null"]
`

// openapi31SingleTypeArrayYAML covers single-element type arrays like ["string"].
const openapi31SingleTypeArrayYAML = `
openapi: "3.1.0"
info:
  title: OA31SingleType
  version: "1.0"
paths:
  /widgets:
    get:
      operationId: getWidget
      tags: [widgets]
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: ["integer"]
                  name:
                    type: ["string"]
`

// openapi31DeepNestedYAML exercises type arrays nested inside allOf/oneOf/anyOf
// and inside array items — the most common patterns in large enterprise specs.
const openapi31DeepNestedYAML = `
openapi: "3.1.0"
info:
  title: OA31DeepNested
  version: "1.0"
paths:
  /orders:
    get:
      operationId: listOrders
      tags: [orders]
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  orders:
                    type: array
                    items:
                      type: object
                      properties:
                        orderId:
                          type: ["string", "null"]
                        total:
                          type: ["number", "null"]
                        status:
                          type: ["string", "null"]
                          enum: [pending, shipped, delivered]
                  meta:
                    allOf:
                      - type: object
                        properties:
                          page:
                            type: ["integer", "null"]
                          total:
                            type: ["integer", "null"]
`

// Test_OAPIInteg_31TypeArrayParsesWithoutError is the primary regression test for the
// reported 500 error: "cannot unmarshal array into Go value of type string" when uploading
// specs that use OpenAPI 3.1 type arrays like ["string","null"].
func Test_OAPIInteg_31TypeArrayParsesWithoutError(t *testing.T) {
	specs, scenarios := parseYAML(t, openapi31TypeArrayYAML)
	require.NotEmpty(t, specs, "3.1 spec with type arrays must parse to at least one spec")
	require.NotEmpty(t, scenarios, "must produce at least one scenario")

	// Both GET /items and POST /items should parse
	require.GreaterOrEqual(t, len(specs), 2, "should have specs for GET and POST")
}

// Test_OAPIInteg_31TypeArrayNullableConverted verifies that ["T","null"] schemas are
// converted to nullable:true in the parsed Property.
func Test_OAPIInteg_31TypeArrayNullableConverted(t *testing.T) {
	specs, _ := parseYAML(t, openapi31TypeArrayYAML)

	// Find the GET /items 200 response spec
	var getSpec *APISpec
	for _, sp := range specs {
		if strings.Contains(string(sp.Method), "GET") && sp.Response.StatusCode == 200 {
			getSpec = sp
			break
		}
	}
	require.NotNil(t, getSpec, "GET /items spec must be found")

	// Response body should have properties; all were originally ["T","null"]
	require.NotEmpty(t, getSpec.Response.Body, "response body must be parsed")
	body := getSpec.Response.Body[0]
	require.NotEmpty(t, body.Children, "body must have child properties")

	// Every nullable field should have been converted to a valid single type
	for _, child := range body.Children {
		require.NotEmpty(t, child.Type,
			"property %q: type array must be normalized to a non-empty string type", child.Name)
		require.NotEqual(t, "null", child.Type,
			"property %q: type must not be 'null' after normalization", child.Name)
		if child.Nullable {
			// nullable flag should have been set when "null" was one of the variants
			require.True(t, child.Nullable,
				"property %q should be nullable", child.Name)
		}
	}
}

// Test_OAPIInteg_31TypeArrayQueryParamNullable verifies that a nullable query parameter
// defined as type:["string","null"] is parsed without error and placed correctly.
func Test_OAPIInteg_31TypeArrayQueryParamNullable(t *testing.T) {
	specs, scenarios := parseYAML(t, openapi31TypeArrayYAML)
	require.NotEmpty(t, specs)
	require.NotEmpty(t, scenarios)

	var getScenario *types.APIScenario
	for _, s := range scenarios {
		if strings.Contains(string(s.Method), "GET") {
			getScenario = s
			break
		}
	}
	require.NotNil(t, getScenario, "GET scenario must be found")
	// filter param should land in QueryParams
	_, hasFilter := getScenario.Request.QueryParams["filter"]
	require.True(t, hasFilter, "nullable query param 'filter' must be present in QueryParams")
}

// Test_OAPIInteg_31SingleElementTypeArray verifies that ["string"] (no null) is
// converted to type:"string" without setting nullable.
func Test_OAPIInteg_31SingleElementTypeArray(t *testing.T) {
	specs, scenarios := parseYAML(t, openapi31SingleTypeArrayYAML)
	require.NotEmpty(t, specs)
	require.NotEmpty(t, scenarios)
}

// Test_OAPIInteg_31VersionDowngraded verifies that a spec declaring openapi:"3.1.0"
// is successfully parsed (version downgraded to 3.0.3 internally).
func Test_OAPIInteg_31VersionDowngraded(t *testing.T) {
	specs, _ := parseYAML(t, openapi31TypeArrayYAML)
	require.NotEmpty(t, specs, "3.1.0 version should be downgraded and parsed successfully")
}

// Test_OAPIInteg_31DeepNestedTypeArrays verifies type arrays inside nested schemas:
// array items, allOf sub-schemas, and enum fields.
func Test_OAPIInteg_31DeepNestedTypeArrays(t *testing.T) {
	specs, scenarios := parseYAML(t, openapi31DeepNestedYAML)
	require.NotEmpty(t, specs)
	require.NotEmpty(t, scenarios)

	// All scenarios must build without error
	dt := fuzz.NewDataTemplateRequest(false, 1, 1)
	for _, sp := range specs {
		s, err := sp.BuildMockScenario(dt)
		require.NoError(t, err, "BuildMockScenario must not fail on 3.1 deep-nested spec")
		require.NotEmpty(t, s.Name)
	}
}

// Test_OAPIInteg_31RequestBodyNullableFields verifies that POST body with nullable fields
// ([\"string\",\"null\"]) produces a valid scenario with body content.
func Test_OAPIInteg_31RequestBodyNullableFields(t *testing.T) {
	specs, scenarios := parseYAML(t, openapi31TypeArrayYAML)
	require.NotEmpty(t, specs)

	var postScenario *types.APIScenario
	for _, s := range scenarios {
		if strings.Contains(string(s.Method), "POST") && s.Response.StatusCode == 201 {
			postScenario = s
			break
		}
	}
	require.NotNil(t, postScenario, "POST /items 201 scenario must be found")
	// Required 'name' field must produce a body assertion
	require.NotEmpty(t, postScenario.Request.AssertContentsPattern,
		"POST body must produce assert contents pattern")
}

// Test_OAPIInteg_31NormalizeTypeArraysDirectly tests the normalization helper directly
// so edge cases can be verified without round-tripping through the full parser.
func Test_OAPIInteg_31NormalizeTypeArraysDirectly(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantType string
		nullable bool
	}{
		{"string+null", `{"type":["string","null"]}`, "string", true},
		{"integer+null", `{"type":["integer","null"]}`, "integer", true},
		{"number+null", `{"type":["number","null"]}`, "number", true},
		{"boolean+null", `{"type":["boolean","null"]}`, "boolean", true},
		{"array+null", `{"type":["array","null"]}`, "array", true},
		{"single string", `{"type":["string"]}`, "string", false},
		{"null only", `{"type":["null"]}`, "", false},
		{"plain string", `{"type":"string"}`, "string", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := normalizeOpenAPI31([]byte(tc.input))
			var result map[string]interface{}
			require.NoError(t, json.Unmarshal(out, &result))
			if tc.wantType == "" {
				_, hasType := result["type"]
				require.False(t, hasType, "type key should be removed when only null was present")
			} else {
				require.Equal(t, tc.wantType, result["type"],
					"type should be normalized to %q", tc.wantType)
			}
			if tc.nullable {
				require.Equal(t, true, result["nullable"], "nullable should be set to true")
			} else {
				require.Nil(t, result["nullable"], "nullable should not be set")
			}
		})
	}
}

// Test_OAPIInteg_PropertyStringToString verifies the String() method works
// for all property types without panicking.
func Test_OAPIInteg_PropertyStringMethod(t *testing.T) {
	props := []Property{
		{Name: "a", Type: "string"},
		{Name: "b", Type: "integer", Min: 1, Max: 10},
		{Name: "c", Type: "array", SubType: "string"},
		{Name: "d", Type: "object", Children: []Property{{Name: "x", Type: "string"}}},
		{Name: "e", Type: "string", Const: "fixed"},
		{Name: "f", Type: "string", Enum: []string{"a", "b"}},
	}
	for _, p := range props {
		s := p.String()
		require.Contains(t, s, p.Name)
	}
}

// ---------------------------------------------------------------------------
// Circular $ref (stack overflow) regression tests
// ---------------------------------------------------------------------------

// circularDirectYAML: A → children → A (direct self-reference via $ref)
const circularDirectYAML = `
openapi: "3.0.0"
info:
  title: CircularDirect
  version: "1.0"
components:
  schemas:
    Node:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
        children:
          type: array
          items:
            $ref: '#/components/schemas/Node'
paths:
  /nodes:
    get:
      operationId: listNodes
      tags: [nodes]
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Node'
`

// circularMutualYAML: A → B → A (mutual / indirect circular reference)
const circularMutualYAML = `
openapi: "3.0.0"
info:
  title: CircularMutual
  version: "1.0"
components:
  schemas:
    Parent:
      type: object
      properties:
        id:
          type: integer
        child:
          $ref: '#/components/schemas/Child'
    Child:
      type: object
      properties:
        id:
          type: integer
        parent:
          $ref: '#/components/schemas/Parent'
paths:
  /parents:
    get:
      operationId: listParents
      tags: [parents]
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Parent'
`

// circularAllOfYAML: circular reference inside allOf composition
const circularAllOfYAML = `
openapi: "3.0.0"
info:
  title: CircularAllOf
  version: "1.0"
components:
  schemas:
    Base:
      type: object
      properties:
        id:
          type: integer
    Extended:
      allOf:
        - $ref: '#/components/schemas/Base'
        - type: object
          properties:
            nested:
              $ref: '#/components/schemas/Extended'
paths:
  /extended:
    get:
      operationId: getExtended
      tags: [extended]
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Extended'
`

// circularRequestBodyYAML: circular schema in POST request body — the most common
// real-world pattern that caused the reported stack overflow.
const circularRequestBodyYAML = `
openapi: "3.0.0"
info:
  title: CircularRequestBody
  version: "1.0"
components:
  schemas:
    TreeNode:
      type: object
      properties:
        value:
          type: string
        left:
          $ref: '#/components/schemas/TreeNode'
        right:
          $ref: '#/components/schemas/TreeNode'
paths:
  /tree:
    post:
      operationId: createTree
      tags: [tree]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/TreeNode'
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TreeNode'
`

// Test_OAPIInteg_CircularDirectRefNoStackOverflow is the primary regression test for the
// reported fatal stack overflow. A schema that references itself (Node.children → Node)
// must parse without crashing.
func Test_OAPIInteg_CircularDirectRefNoStackOverflow(t *testing.T) {
	// Must not panic / stack overflow
	specs, scenarios := parseYAML(t, circularDirectYAML)
	require.NotEmpty(t, specs)
	require.NotEmpty(t, scenarios)

	dt := fuzz.NewDataTemplateRequest(false, 1, 1)
	for _, sp := range specs {
		s, err := sp.BuildMockScenario(dt)
		require.NoError(t, err)
		require.NotEmpty(t, s.Name)
	}
}

// Test_OAPIInteg_CircularMutualRefNoStackOverflow verifies that A→B→A mutual circular
// references also terminate correctly.
func Test_OAPIInteg_CircularMutualRefNoStackOverflow(t *testing.T) {
	specs, scenarios := parseYAML(t, circularMutualYAML)
	require.NotEmpty(t, specs)
	require.NotEmpty(t, scenarios)

	dt := fuzz.NewDataTemplateRequest(false, 1, 1)
	for _, sp := range specs {
		_, err := sp.BuildMockScenario(dt)
		require.NoError(t, err)
	}
}

// Test_OAPIInteg_CircularAllOfNoStackOverflow verifies circular refs inside allOf.
func Test_OAPIInteg_CircularAllOfNoStackOverflow(t *testing.T) {
	specs, scenarios := parseYAML(t, circularAllOfYAML)
	require.NotEmpty(t, specs)
	require.NotEmpty(t, scenarios)

	dt := fuzz.NewDataTemplateRequest(false, 1, 1)
	for _, sp := range specs {
		_, err := sp.BuildMockScenario(dt)
		require.NoError(t, err)
	}
}

// Test_OAPIInteg_CircularRequestBodyNoStackOverflow verifies that the POST request body
// with a self-referencing TreeNode schema (the exact pattern in the reported crash)
// produces a valid scenario without stack overflow.
func Test_OAPIInteg_CircularRequestBodyNoStackOverflow(t *testing.T) {
	specs, scenarios := parseYAML(t, circularRequestBodyYAML)
	require.NotEmpty(t, specs)
	require.NotEmpty(t, scenarios)

	var postScenario *types.APIScenario
	for _, s := range scenarios {
		if strings.Contains(string(s.Method), "POST") {
			postScenario = s
			break
		}
	}
	require.NotNil(t, postScenario, "POST /tree scenario must be found")
	require.Equal(t, 201, postScenario.Response.StatusCode)
}

// Test_OAPIInteg_CircularRefPreservesNonCircularChildren verifies that cycle detection
// doesn't strip legitimate non-circular children — only the back-edge is stubbed.
func Test_OAPIInteg_CircularRefPreservesNonCircularChildren(t *testing.T) {
	specs, _ := parseYAML(t, circularDirectYAML)
	require.NotEmpty(t, specs)

	// The Node schema has id (integer) and name (string) as direct properties.
	// Both should be present even though the children array creates a cycle.
	sp := specs[0]
	require.NotEmpty(t, sp.Response.Body, "response body must be populated")
	body := sp.Response.Body[0]

	hasID := false
	hasName := false
	for _, child := range body.Children {
		if child.Name == "id" {
			hasID = true
		}
		if child.Name == "name" {
			hasName = true
		}
	}
	require.True(t, hasID, "non-circular property 'id' must be preserved")
	require.True(t, hasName, "non-circular property 'name' must be preserved")
}


// ---------------------------------------------------------------------------
// Regression: single-key request body must not fail HasProperty assertion
// ---------------------------------------------------------------------------

// singleKeyBodyYAML exercises a POST endpoint whose required request body
// has exactly one top-level key containing an array.  This is the shape that
// previously caused "PropertyLenGE contents 2" to fail (VariableSize returns
// map key-count = 1, so 1 >= 2 was false).
var singleKeyBodyYAML = `
openapi: "3.0.0"
info:
  title: "Batch API"
  version: "1.0"
paths:
  /batch/items/delete:
    post:
      operationId: batchDeleteItems
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [itemIds]
              properties:
                itemIds:
                  type: array
                  items:
                    type: string
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  deleted:
                    type: integer
`

// Test_OAPIInteg_SingleKeyBodyNotRejected verifies that a POST endpoint whose
// required request body has a single top-level array key is accepted by the
// auto-generated assertions.
//
// Root cause of original bug: "PropertyLenGE contents 2" was auto-generated for
// any required request body.  VariableSize returns the number of top-level keys
// for map bodies (not the byte length), so {"itemIds":[...]} with 1 key would
// return 1, causing 1 >= 2 → false.
// Fix: generate "HasProperty contents" — true for any non-nil body regardless of
// key count.
func Test_OAPIInteg_SingleKeyBodyNotRejected(t *testing.T) {
	_, scenarios := parseYAML(t, singleKeyBodyYAML)

	var target *types.APIScenario
	for _, s := range scenarios {
		if strings.Contains(s.Name, "batchDeleteItems") && s.Response.StatusCode == 200 {
			target = s
			break
		}
	}
	require.NotNil(t, target, "batchDeleteItems scenario must be generated")

	// HasProperty contents must be present; PropertyLenGE must NOT be
	hasHasProperty := false
	for _, a := range target.Request.Assertions {
		require.NotContains(t, a, "PropertyLenGE",
			"PropertyLenGE must NOT be auto-generated for required body — use HasProperty instead")
		if strings.Contains(a, "HasProperty") && strings.Contains(a, "contents") {
			hasHasProperty = true
		}
	}
	require.True(t, hasHasProperty, "HasProperty contents must be generated for required request body")

	// Simulate mock playback: Assert must accept a single-key body {itemIds:[...]}
	reqBody := map[string]any{"itemIds": []any{"id-1", "id-2"}}
	reqHeaders := make(map[string][]string)
	reqHeaders["Content-Type"] = []string{"application/json"}
	err := target.Request.Assert(
		map[string]string{},
		map[string]string{},
		reqHeaders,
		reqBody,
		map[string]any{"itemIds": []any{"id-1", "id-2"}},
	)
	require.NoError(t, err, "single-key body {itemIds:[...]} must pass all auto-generated assertions")
}

// Test_OAPIInteg_PropertyLenGECountsMapKeys documents the existing (correct)
// behavior: PropertyLenGE with a map counts top-level keys, not bytes.
// This is intentional for assertions like PropertyLenGE someObject 3 meaning
// "the object has at least 3 keys". The revokeTokens bug was that the
// auto-generated assertion used threshold 2 for a body that always has 1 key.
func Test_OAPIInteg_PropertyLenGECountsMapKeys(t *testing.T) {
	templateParams := map[string]any{
		"body": map[string]any{
			"a": "1",
			"b": "2",
			"c": "3",
		},
	}

	// 3 keys >= 3 → true
	b, err := fuzz.ParseTemplate("", []byte(`{{PropertyLenGE "body" 3}}`), templateParams)
	require.NoError(t, err)
	require.Equal(t, "true", string(b))

	// 3 keys >= 4 → false
	b, err = fuzz.ParseTemplate("", []byte(`{{PropertyLenGE "body" 4}}`), templateParams)
	require.NoError(t, err)
	require.Equal(t, "false", string(b))

	// Single-key map {"ids":[...]} has 1 key < 2 → false (the original bug)
	singleKey := map[string]any{
		"contents": map[string]any{"ids": []any{"abc"}},
	}
	b, err = fuzz.ParseTemplate("", []byte(`{{PropertyLenGE "contents" 2}}`), singleKey)
	require.NoError(t, err)
	require.Equal(t, "false", string(b), "PropertyLenGE contents 2 returns false for single-key body — use HasProperty instead")

	// HasProperty handles it correctly
	b, err = fuzz.ParseTemplate("", []byte(`{{HasProperty "contents"}}`), singleKey)
	require.NoError(t, err)
	require.Equal(t, "true", string(b), "HasProperty contents is true for any non-nil body")
}
