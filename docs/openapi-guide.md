# OpenAPI Guide

api-mock-service can import OpenAPI 3.x specifications and automatically generate YAML scenario files for every path, method, status code, and schema variant. The generated scenarios include fuzz-ready response templates and regex-based assertions derived from the spec schema.

## Uploading a Spec

```bash
curl -H "Content-Type: application/yaml" \
  --data-binary @my-api.yaml \
  http://localhost:8080/_oapi
```

JSON specs also work:
```bash
curl -H "Content-Type: application/json" \
  --data-binary @my-api.json \
  http://localhost:8080/_oapi
```

## What Gets Generated

For each `path × method × status code` combination the parser creates a YAML scenario with:

- **Response template** using fuzz functions matched to the schema types:
  - `string` + `pattern` → `{{RandRegex "..."}}`
  - `string` + `format: date-time` → `{{Time}}`
  - `string` + `format: uri` → `{{RandURL}}`
  - `string` + `format: email` → `{{RandEmail}}`
  - `integer` → `{{RandIntMinMax min max}}`
  - `number` → `{{RandFloatMinMax min max}}`
  - `boolean` → `{{RandBool}}`
  - `array` → iterates with `Iterate`

- **`assert_contents_pattern`** with type tokens auto-derived from the schema:
  - `__string__^AC[0-9a-fA-F]{32}$` for a patterned string
  - `__number__[+-]?[0-9]{1,10}` for integers
  - `__boolean__(false|true)` for booleans

Example input spec fragment:
```yaml
paths:
  /v1/AuthTokens/Promote:
    post:
      responses:
        '200':
          content:
            application/json:
              schema:
                properties:
                  account_sid:
                    type: string
                    pattern: "^AC[0-9a-fA-F]{32}$"
                  auth_token:
                    type: string
                  url:
                    type: string
                    format: uri
```

Generated scenario:
```yaml
method: POST
name: UpdateAuthTokenPromotion-<hash>
path: /v1/AuthTokens/Promote
group: v1_AuthTokens_Promote
response:
  status_code: 200
  contents: >
    {"account_sid":"{{RandRegex `^AC[0-9a-fA-F]{32}$`}}",
     "auth_token":"{{RandStringMinMax 0 64}}",
     "url":"{{RandURL}}"}
  assert_contents_pattern: >
    {"account_sid":"(__string__^AC[0-9a-fA-F]{32}$)",
     "auth_token":"(__string__\\w+)"}
```

## Discriminator / oneOf / anyOf Support

The parser generates **one scenario per variant**, so every branch is tested — no variant is silently ignored.

### Without a Discriminator

Given a schema like:
```yaml
oneOf:
  - $ref: '#/components/schemas/Cat'
  - $ref: '#/components/schemas/Dog'
```

Generates:
- `CreateAnimal-variant0-201` — scenario using the Cat schema
- `CreateAnimal-variant1-201` — scenario using the Dog schema

### With a Discriminator

```yaml
oneOf:
  - $ref: '#/components/schemas/Cat'
  - $ref: '#/components/schemas/Dog'
discriminator:
  propertyName: petType
  mapping:
    cat: '#/components/schemas/Cat'
    dog: '#/components/schemas/Dog'
```

Generates:
- `CreateAnimal-cat-201` — scenario with `petType: "cat"` + Cat fields
- `CreateAnimal-dog-201` — scenario with `petType: "dog"` + Dog fields

The discriminator field is automatically set to the correct value in the response template.

### Complete Example

Upload this spec:

```yaml
openapi: "3.0.0"
info:
  title: Pet Store
  version: "1.0"
paths:
  /animals:
    post:
      operationId: CreateAnimal
      requestBody:
        required: true
        content:
          application/json:
            schema:
              oneOf:
                - $ref: '#/components/schemas/Cat'
                - $ref: '#/components/schemas/Dog'
              discriminator:
                propertyName: petType
                mapping:
                  cat: '#/components/schemas/Cat'
                  dog: '#/components/schemas/Dog'
      responses:
        '201':
          content:
            application/json:
              schema:
                oneOf:
                  - $ref: '#/components/schemas/Cat'
                  - $ref: '#/components/schemas/Dog'
                discriminator:
                  propertyName: petType
                  mapping:
                    cat: '#/components/schemas/Cat'
                    dog: '#/components/schemas/Dog'
components:
  schemas:
    Cat:
      type: object
      required: [petType, name, indoor]
      properties:
        petType: {type: string}
        name:    {type: string}
        indoor:  {type: boolean}
    Dog:
      type: object
      required: [petType, name, breed]
      properties:
        petType: {type: string}
        name:    {type: string}
        breed:   {type: string}
```

```bash
curl -X POST -H "Content-Type: application/yaml" \
  --data-binary @animals.yaml \
  http://localhost:8080/_oapi
# {"scenarios": 2, ...}
```

Two scenarios are created:

```bash
# Play back the Cat variant
curl -H "X-Mock-Scenario: CreateAnimal-cat-201" \
  -X POST http://localhost:8080/animals \
  -d '{"petType":"cat","name":"Whiskers"}'
# {"petType":"cat","name":"Mittens","indoor":true}

# Play back the Dog variant
curl -H "X-Mock-Scenario: CreateAnimal-dog-201" \
  -X POST http://localhost:8080/animals \
  -d '{"petType":"dog","name":"Rex"}'
# {"petType":"dog","name":"Buddy","breed":"Labrador"}
```

Run contract tests against a real server — both variants are tested automatically:

```bash
api-mock-service producer-contract \
  --group CreateAnimal \
  --base_url https://api.petstore.example.com \
  --spec animals.yaml \
  --times 3
```

### Backward Compatibility

APIs without `oneOf`/`anyOf` are unaffected — they generate exactly one scenario as before.

## Listing Generated Scenarios

```bash
# List all scenarios (includes generated ones)
curl http://localhost:8080/_scenarios

# Filter by group (URL-encoded path prefix)
curl "http://localhost:8080/_scenarios?group=v1_AuthTokens_Promote"
```

## Playing Back Generated Scenarios

Once uploaded, generated scenarios serve mock responses immediately:

```bash
curl -X POST http://localhost:8080/v1/AuthTokens/Promote
```

Returns a fully dynamic response:
```json
{
  "account_sid": "ACF3A7ea7f5c90f6482CEcA77BED07Fb91",
  "auth_token": "PaC7rKdGER73rXNi...",
  "url": "https://Billy.com"
}
```

## Swagger UI

The mock service includes an embedded Swagger UI. After uploading a spec, browse to:

```
http://localhost:8080/_ui
```

You can also run a standalone Swagger UI pointing at the mock service's spec endpoint:

```bash
docker run -p 8080:8080 \
  -e BASE_URL=/swagger \
  -e SWAGGER_JSON=/data/openapi.yaml \
  -v $(pwd)/fixtures/oapi:/data \
  swaggerapi/swagger-ui
```

## Using the Spec for Contract Validation

Providing a spec at contract test time enables response schema validation: each real API response is checked against the OpenAPI schema using `openapi3filter`. Schema violations surface as structured errors, not just "assertion failed".

### CLI

```bash
api-mock-service producer-contract \
  --group my-api \
  --base_url https://api.example.com \
  --spec openapi.yaml \
  --track-coverage
```

### HTTP request body

```json
{
  "base_url": "https://api.example.com",
  "execution_times": 5,
  "track_coverage": true,
  "spec_content": "openapi: 3.0.3\ninfo:\n  title: My API\n  version: 1.0\npaths:\n  ..."
}
```

### What Gets Validated

- **Required fields** — fields marked `required` in the spec that are absent in the response
- **Type mismatches** — e.g., a field declared as `integer` that arrives as a string
- **Format violations** — `date-time`, `uuid`, `email`, `uri` format checks
- **Enum violations** — values outside declared enum set

### Graceful Fallback

- Routes not in the spec are silently skipped (no false failures for undocumented endpoints)
- If `spec_content` cannot be parsed, contract execution continues without schema validation (a warning is logged)

## Coverage Reporting

When `--track-coverage` is set, the tool compares executed scenario paths against all paths in the spec and reports:

```
COVERAGE REPORT
──────────────────────────────────────────────────────────────
Overall: 75.0%  (6/8 paths)

Uncovered paths:
  ✗ DELETE /users/:id
  ✗ PATCH  /users/:id

Method coverage:
  GET    100.0%
  POST   100.0%
  DELETE 0.0%
  PATCH  0.0%
```

HTTP response:
```json
{
  "coverage": {
    "totalPaths": 8,
    "coveredPaths": 6,
    "coveragePercentage": 75.0,
    "uncoveredPaths": ["DELETE /users/:id", "PATCH /users/:id"],
    "methodCoverage": {
      "GET": 100.0,
      "POST": 100.0,
      "DELETE": 0.0,
      "PATCH": 0.0
    }
  }
}
```

## Downloading the Spec

```bash
# Get the uploaded spec back
curl http://localhost:8080/_oapi
```

## Related Docs

- [Contract Testing](contract-testing.md) — schema validation, coverage, mutations
- [CLI Reference](cli-reference.md) — `--spec` and `--track-coverage` flags
- [API Reference](api-reference.md) — `/_oapi` endpoint details
- [Fuzz & Property Testing](fuzz-property-testing.md) — mutation strategies
