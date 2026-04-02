# How-To Guide — api-mock-service Cookbook

Each section solves one concrete problem. Copy the relevant snippet and adapt it to your API. For deeper reference, follow the links to the dedicated docs.

---

## Contents

1. [Start the server](#1-start-the-server)
2. [Record real traffic through the proxy](#2-record-real-traffic-through-the-proxy)
3. [Play back a recorded mock](#3-play-back-a-recorded-mock)
4. [Write a mock scenario by hand](#4-write-a-mock-scenario-by-hand)
5. [Generate dynamic response data with templates](#5-generate-dynamic-response-data-with-templates)
6. [Route requests to different scenarios based on request content](#6-route-requests-to-different-scenarios-based-on-request-content)
7. [Chain multi-step scenarios (create → read → delete)](#7-chain-multi-step-scenarios-create--read--delete)
8. [Import a HAR or Postman collection and get instant assertions](#8-import-a-har-or-postman-collection-and-get-instant-assertions)
9. [Import an OpenAPI spec and get instant mocks](#9-import-an-openapi-spec-and-get-instant-mocks)
10. [Mock polymorphic / discriminator APIs (oneOf / anyOf)](#10-mock-polymorphic--discriminator-apis-oneof--anyof)
11. [Inject request body fields into response templates](#11-inject-request-body-fields-into-response-templates)
12. [Run producer contract tests against a real API](#12-run-producer-contract-tests-against-a-real-api)
13. [Assert nested response fields with JSONPath](#13-assert-nested-response-fields-with-jsonpath)
14. [Validate responses against an OpenAPI schema](#14-validate-responses-against-an-openapi-schema)
15. [Get field-level diagnostics when a test fails](#15-get-field-level-diagnostics-when-a-test-fails)
16. [Track which API paths were exercised (coverage)](#16-track-which-api-paths-were-exercised-coverage)
17. [Run mutation testing to find validation gaps](#17-run-mutation-testing-to-find-validation-gaps)
18. [Shrink a failing mutation to its minimal reproducer](#18-shrink-a-failing-mutation-to-its-minimal-reproducer)
19. [Test stateful workflows (session-scoped state machine)](#19-test-stateful-workflows-session-scoped-state-machine)
20. [Detect breaking API changes in CI](#20-detect-breaking-api-changes-in-ci)
21. [Preview scenarios without executing (dry run)](#21-preview-scenarios-without-executing-dry-run)
22. [Inject chaos: errors and latency](#22-inject-chaos-errors-and-latency)
23. [Serve static files and binary fixtures](#23-serve-static-files-and-binary-fixtures)
24. [Share variables across all scenarios in a group](#24-share-variables-across-all-scenarios-in-a-group)
25. [Use deterministic / seeded test data](#25-use-deterministic--seeded-test-data)

---

## 1. Start the Server

**Problem:** I want to run the mock service locally.

```bash
# From source
make && ./out/bin/api-mock-service

# Docker
docker run -p 8080:8080 -p 8081:8081 \
  -e DATA_DIR=/tmp/mocks \
  plexobject/api-mock-service:latest

# Custom ports and data directory
./api-mock-service \
  --httpPort 9090 \
  --proxyPort 9091 \
  --dataDir /var/mocks
```

**Ports:**
- `8080` — mock playback, contract testing, OpenAPI, all management APIs
- `8081` — transparent proxy recorder (point your HTTP client here)

**Health check:**
```bash
curl http://localhost:8080/_health
# {"status": "ok"}
```

→ [CLI Reference](cli-reference.md)

---

## 2. Record Real Traffic Through the Proxy

**Problem:** I want to capture real API calls and automatically convert them to mock scenarios.

```bash
# Set proxy environment variables
export http_proxy="http://localhost:8081"
export https_proxy="http://localhost:8081"

# Make requests — they are transparently forwarded and saved
curl -k https://api.example.com/orders/42
curl -k -X POST https://api.example.com/orders \
  -H "Content-Type: application/json" \
  -d '{"customerId":"cust-1","amount":99}'
```

Scenarios are saved automatically under `default_mocks_data/`.

**From code (Python):**
```python
import requests
proxies = {'http': 'http://localhost:8081', 'https': 'http://localhost:8081'}
resp = requests.get('https://api.example.com/orders/42', proxies=proxies, verify=False)
```

**Via the `/_proxy` passthrough endpoint (no proxy env needed):**
```bash
curl -H "X-Mock-Url: https://api.example.com/orders/42" \
     -H "Authorization: Bearer token" \
     http://localhost:8080/_proxy
```

→ [Mock Guide — Recording](mock-guide.md)

---

## 3. Play Back a Recorded Mock

**Problem:** I want to replay recorded traffic without hitting the real API.

```bash
# Any request on port 8080 is matched against stored scenarios
curl http://localhost:8080/orders/42

# Force a specific scenario by name
curl -H "X-Mock-Scenario: get-order-200" \
  http://localhost:8080/orders/42

# Override the response status code for this request
curl -H "X-Mock-Response-Status: 503" \
  http://localhost:8080/orders/42

# Inject an artificial delay
curl -H "X-Mock-Wait-Before-Reply: 500ms" \
  http://localhost:8080/orders/42
```

**Match priority:**
1. Exact match on `X-Mock-Scenario` header
2. Method + path + matching request assertions
3. Method + path (round-robin across multiple matching scenarios)
4. 404 if nothing matches

→ [Mock Guide — Playback](mock-guide.md), [API Reference](api-reference.md)

---

## 4. Write a Mock Scenario by Hand

**Problem:** I want to create a scenario without recording real traffic.

```yaml
# save as get-user.yaml
name: get-user
method: GET
path: /users/:id
group: users
description: Returns a user by ID
request:
  assert_headers_pattern:
    Authorization: "Bearer .+"
response:
  status_code: 200
  headers:
    Content-Type: ["application/json"]
  contents: >
    {
      "id": {{.id}},
      "name": "{{RandName}}",
      "email": "{{RandEmail}}",
      "createdAt": "{{Time}}"
    }
  assert_contents_pattern: >
    {"id": "__number__\\d+",
     "name": "__string__\\w+",
     "email": "__string__\\S+@\\S+"}
```

```bash
# Upload it
curl -X POST -H "Content-Type: application/yaml" \
  --data-binary @get-user.yaml \
  http://localhost:8080/_scenarios

# Test it
curl -H "Authorization: Bearer tok123" \
  http://localhost:8080/users/42
# {"id": 42, "name": "James Riley", "email": "jriley@example.com", "createdAt": "2026-04-02T..."}

# List all scenarios
curl http://localhost:8080/_scenarios

# Delete it
curl -X DELETE http://localhost:8080/_scenarios/GET/get-user/users/:id
```

→ [Mock Guide](mock-guide.md), [API Reference — Scenario Management](api-reference.md#scenario-management)

---

## 5. Generate Dynamic Response Data with Templates

**Problem:** I want mock responses with realistic random data, not static fixtures.

```yaml
name: create-order
method: POST
path: /orders
group: orders
response:
  status_code: 201
  contents: >
    {
      "orderId":    "{{UUID}}",
      "customerId": "cust-{{RandIntMinMax 1000 9999}}",
      "amount":     {{RandFloatMinMax 1.0 10000.0}},
      "currency":   "{{RandCurrencyCode}}",
      "email":      "{{RandEmail}}",
      "phone":      "{{RandPhone}}",
      "city":       "{{RandCity}}",
      "country":    "{{RandCountryCode}}",
      "createdAt":  "{{Time}}",
      "expiresAt":  "{{RandFutureDate}}",
      "checksum":   "{{RandSHA256}}",
      "status":     "{{EnumString "pending processing shipped"}}"
    }
```

**Key template categories:**

| Category | Examples |
|----------|---------|
| Identity | `{{UUID}}`, `{{ULID}}`, `{{RandName}}`, `{{RandEmail}}`, `{{RandPhone}}` |
| Auth | `{{RandUsername}}`, `{{RandPassword}}`, `{{RandSlug}}` |
| Numbers | `{{RandIntMinMax 1 100}}`, `{{RandFloatMinMax 0.0 1.0}}` |
| Strings | `{{RandString 20}}`, `{{RandRegex "^[A-Z]{3}[0-9]{6}$"}}` |
| Pick one | `{{EnumString "a b c"}}`, `{{EnumInt 200 201 204}}` |
| Location | `{{RandCity}}`, `{{RandCountry}}`, `{{RandUSPostal}}`, `{{RandLatitude}}`, `{{RandLongitude}}`, `{{RandTimezone}}` |
| Time | `{{Time}}`, `{{Date}}`, `{{RandFutureDate}}`, `{{RandPastDate}}`, `{{RandUnixTimestamp}}` |
| Network | `{{RandIP}}`, `{{RandIPv6}}`, `{{RandMACAddress}}`, `{{RandPort}}` |
| Crypto | `{{RandSHA256}}`, `{{RandMD5}}`, `{{RandBase64}}` |
| Financial | `{{RandCurrencyCode}}`, `{{RandCreditCard}}` |
| Versioning | `{{RandSemver}}`, `{{RandMimeType}}`, `{{RandFilename}}`, `{{RandHTTPStatus}}` |
| UI | `{{RandHexColor}}`, `{{RandRGBColor}}` |
| Boolean | `{{RandBool}}` |
| From file | `{{RandFileLine "names.txt"}}`, `{{FileProperty "cfg.yaml" "key"}}` |

**Loop to build arrays:**
```yaml
contents: >
  {"items": [
    {{- range $i := Iterate .count}}
    {"id": {{$i}}, "name": "{{SeededName $i}}"}
    {{if LastIter $i $.count}}{{else}},{{end}}
    {{- end}}
  ]}
```

→ [Mock Guide — Template Functions](mock-guide.md)

---

## 6. Route Requests to Different Scenarios Based on Request Content

**Problem:** I want `/orders` to return different responses depending on query params, headers, or request body.

```yaml
# Scenario 1: Premium customers (header-based routing)
name: get-orders-premium
method: GET
path: /orders
group: orders
request:
  assert_headers_pattern:
    X-Customer-Tier: "premium"
response:
  status_code: 200
  contents: '{"orders": [], "limit": 1000, "tier": "premium"}'

---
# Scenario 2: Standard customers
name: get-orders-standard
method: GET
path: /orders
group: orders
request:
  assert_headers_pattern:
    X-Customer-Tier: "(standard|free)"
response:
  status_code: 200
  contents: '{"orders": [], "limit": 10, "tier": "standard"}'

---
# Scenario 3: Paginated (query param)
name: get-orders-page2
method: GET
path: /orders
group: orders
request:
  assert_query_params_pattern:
    page: "[2-9]|[1-9][0-9]+"
response:
  status_code: 200
  contents: '{"orders": [], "page": {{.page}}, "hasMore": false}'
```

**Conditional response within one scenario:**
```yaml
# Use NthRequest to alternate: fail every 5th request
response:
  {{if NthRequest 5}}
  status_code: 429
  contents: '{"error": "rate limited"}'
  {{else}}
  status_code: 200
  contents: '{"orders": []}'
  {{end}}
```

→ [Mock Guide — Request Assertions](mock-guide.md)

---

## 7. Chain Multi-Step Scenarios (create → read → delete)

**Problem:** I want to test a workflow where step 2 uses data from step 1's response.

```yaml
# Step 1: Create (order: 0)
name: create-product
method: POST
path: /products
order: 0
group: product-lifecycle
request:
  contents: '{"name":"{{RandWord 1 1}}","price":{{RandFloatMinMax 1.0 999.0}}}'
response:
  status_code: 201
  contents: '{"id": {{RandIntMinMax 100 9999}}, "name": "Widget", "price": 29.99}'
  add_shared_variables:
    - id         # captures "id" from response for use in subsequent steps
  assertions:
    - NumPropertyGE contents.id 100

---
# Step 2: Read (order: 1) — uses {{.id}} from step 1
name: get-product
method: GET
path: /products/:id
order: 1
group: product-lifecycle
response:
  status_code: 200
  contents: '{"id": {{.id}}, "name": "Widget", "price": 29.99}'
  assertions:
    - NumPropertyGE contents.id 100

---
# Step 3: Delete (order: 2)
name: delete-product
method: DELETE
path: /products/:id
order: 2
group: product-lifecycle
response:
  status_code: 204
  contents: ""
```

```bash
# Run the full chain 3 times
curl -X POST http://localhost:8080/_contracts/product-lifecycle \
  -H "Content-Type: application/json" \
  -d '{"base_url": "https://api.example.com", "execution_times": 3}'

# CLI
api-mock-service producer-contract \
  --group product-lifecycle \
  --base_url https://api.example.com \
  --times 3
```

→ [Contract Testing — Chaining](contract-testing.md)

---

## 8. Import a HAR or Postman Collection and Get Instant Assertions

**Problem:** I have recorded browser traffic (HAR) or a Postman collection and want contract assertions automatically generated.

```bash
# Import HAR (exported from Chrome DevTools → Network → Save all as HAR)
curl -X POST http://localhost:8080/_history/har \
  --data-binary @recording.har

# Import Postman collection
curl -X POST http://localhost:8080/_history/postman \
  -H "Content-Type: */*" \
  --data-binary @collection.json
```

For a recorded response like `{"id":42,"email":"user@example.com","active":true}`, the import automatically generates:

```yaml
response:
  assert_contents_pattern: >
    {"id": "__number__\\d+",
     "email": "__string__\\S+",
     "active": "__boolean__(true|false)"}
```

**Run contract tests against the real API using imported scenarios:**
```bash
api-mock-service producer-contract \
  --group my-service \
  --base_url https://api.example.com \
  --times 5
```

Any future API change that breaks the recorded shape fails the test.

→ [API Reference — History](api-reference.md#history), [Fuzz & Property Testing — HAR/Postman Import](fuzz-property-testing.md)

---

## 9. Import an OpenAPI Spec and Get Instant Mocks

**Problem:** I have an OpenAPI spec and want mock scenarios generated for every endpoint.

```bash
# Upload the spec — generates scenarios for all paths × methods × status codes
curl -X POST -H "Content-Type: application/yaml" \
  --data-binary @openapi.yaml \
  http://localhost:8080/_oapi
# {"scenarios": 42, ...}

# Download the spec back (for verification)
curl http://localhost:8080/_oapi > current-spec.yaml

# List what was generated
curl http://localhost:8080/_scenarios | jq 'keys'
```

**What gets generated per endpoint:**
- One scenario per `(path, method, status code)` combination
- Response body templates using schema-inferred fuzz functions
- Type-aware `assert_contents_pattern` assertions from schema constraints
- For `oneOf`/`anyOf`, one scenario per variant (see recipe 10)

```bash
# Play back a generated scenario
curl http://localhost:8080/users/42
# {"id": 42, "name": "Jill Torres", "email": "jtorres@example.net", ...}

# Browse via Swagger UI
open http://localhost:8080/_ui
```

→ [OpenAPI Guide](openapi-guide.md)

---

## 10. Mock Polymorphic / Discriminator APIs (oneOf / anyOf)

**Problem:** My API uses `oneOf` or `anyOf` and I need separate mocks for each variant.

Given a spec with:
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

After `POST /_oapi`, two scenarios are created: `CreateAnimal-cat-201` and `CreateAnimal-dog-201`.

```bash
# Play back cat variant — petType is set automatically
curl -H "X-Mock-Scenario: CreateAnimal-cat-201" \
  -X POST http://localhost:8080/animals \
  -d '{"petType":"cat","name":"Whiskers"}'
# {"petType":"cat","name":"Mittens","indoor":true}

# Play back dog variant
curl -H "X-Mock-Scenario: CreateAnimal-dog-201" \
  -X POST http://localhost:8080/animals \
  -d '{"petType":"dog","name":"Rex"}'
# {"petType":"dog","name":"Buddy","breed":"Labrador"}

# Run contract tests against the real API — both variants exercised
api-mock-service producer-contract \
  --group CreateAnimal \
  --base_url https://api.example.com \
  --spec animals.yaml
```

Without a discriminator, variants are auto-named `variant0`, `variant1`, etc. Schemas without `oneOf`/`anyOf` are unaffected.

→ [OpenAPI Guide — Discriminator Support](openapi-guide.md)

---

## 11. Inject Request Body Fields into Response Templates

**Problem:** I want the response to echo fields from the request body without any configuration.

```yaml
name: create-order
method: POST
path: /orders
group: orders
response:
  status_code: 201
  contents: >
    {
      "orderId":    "{{UUID}}",
      "customerId": "{{.customerId}}",
      "amount":     {{.amount}},
      "currency":   "{{.currency}}",
      "status":     "pending"
    }
```

Send:
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customerId":"cust-42","amount":150,"currency":"EUR"}'
```

Response:
```json
{"orderId":"3f4a7b8c-...","customerId":"cust-42","amount":150,"currency":"EUR","status":"pending"}
```

**Rules:**
- Top-level string, number, and boolean fields inject automatically — no YAML changes needed
- Path params win over body fields if names conflict (`{{.id}}` → path param, not body)
- Nested objects: `{{.address}}` gives the whole sub-object; individual nested fields are not flattened

→ [Contract Testing — Body→Template Injection](contract-testing.md)

---

## 12. Run Producer Contract Tests Against a Real API

**Problem:** I want to drive a real API with the recorded scenarios and verify the responses still match.

```bash
# HTTP endpoint
curl -X POST http://localhost:8080/_contracts/orders \
  -H "Content-Type: application/json" \
  -d '{
    "base_url": "https://api.example.com",
    "execution_times": 5,
    "verbose": false
  }'

# CLI (equivalent)
api-mock-service producer-contract \
  --group orders \
  --base_url https://api.example.com \
  --times 5

# Run a single specific scenario
api-mock-service producer-contract \
  --scenario fixtures/get_order.yaml \
  --base_url https://api.example.com
```

**CLI table output:**
```
──────────────────────────────────────────────────────────────
SCENARIO                                 STATUS     LATENCY
──────────────────────────────────────────────────────────────
GET /orders/:id-200                      ✓ PASS     45ms
POST /orders-201                         ✗ FAIL     12ms
  Missing: orderId
  Mismatch: status (expected pending, got processing)
──────────────────────────────────────────────────────────────
TOTAL 10  Passed: 9  Failed: 1  Mismatched: 0
```

**Execute by recorded history (runs in the order requests were originally made):**
```bash
curl -X POST http://localhost:8080/_contracts/history/orders \
  -d '{"base_url": "https://api.example.com"}'
```

→ [Contract Testing](contract-testing.md), [CLI Reference](cli-reference.md)

---

## 13. Assert Nested Response Fields with JSONPath

**Problem:** My response has nested objects and arrays and I need assertions beyond top-level fields.

```yaml
response:
  assert_contents_pattern: >
    {"$.user.id":             "__number__\\d+",
     "$.user.email":          "__string__\\S+@\\S+\\.\\S+",
     "$.order.items[0].price":"__number__[0-9]+\\.?[0-9]*",
     "$.order.status":        "(pending|confirmed|shipped)",
     "$.metadata.tags":       "(__array__3)"}
```

**Syntax reference:**

| Key syntax | Matches |
|------------|---------|
| `$.user.email` | `response.user.email` |
| `$.items[0].name` | First element of `items` array, field `name` |
| `$.a.b.c` | Arbitrarily nested path |

**Mixing flat keys and JSONPath in the same scenario:**
```yaml
assert_contents_pattern: >
  {"id":              "__number__\\d+",
   "$.address.city":  "__string__[A-Z][a-z]+",
   "$.address.zip":   "__string__\\d{5}"}
```

Flat keys use direct field lookup; `$.` keys traverse the JSON tree. Both styles are backward compatible and can be mixed freely.

→ [Contract Testing — JSONPath Assertions](contract-testing.md), [Fuzz & Property Testing](fuzz-property-testing.md)

---

## 14. Validate Responses Against an OpenAPI Schema

**Problem:** I want the contract test to fail when the real API returns a response that violates its own OpenAPI schema — wrong types, missing required fields, etc.

```bash
# HTTP: provide spec inline as spec_content
curl -X POST http://localhost:8080/_contracts/orders \
  -H "Content-Type: application/json" \
  -d "{
    \"base_url\": \"https://api.example.com\",
    \"execution_times\": 3,
    \"spec_content\": $(cat openapi.yaml | jq -Rs .)
  }"

# CLI: provide spec as a file
api-mock-service producer-contract \
  --group orders \
  --base_url https://api.example.com \
  --spec openapi.yaml \
  --times 5
```

Schema violations appear in `schemaViolations` in the response (see recipe 15) and in the CLI output:

```
POST /orders-201                         ✗ FAIL
  Schema: amount must be >= 0
  Schema: email is required
```

→ [Contract Testing — Schema Validation](contract-testing.md), [OpenAPI Guide](openapi-guide.md)

---

## 15. Get Field-Level Diagnostics When a Test Fails

**Problem:** A contract test fails with a vague error. I need to know exactly which fields were wrong.

```bash
curl -X POST http://localhost:8080/_contracts/orders \
  -d '{"base_url":"https://api.example.com","execution_times":1}' \
  | jq '.error_details'
```

```json
{
  "create-order_0": {
    "summary": "assertion failed",
    "scenario": "create-order",
    "url": "https://api.example.com/orders",
    "statusCode": 200,
    "expectedStatusCode": 201,
    "missingFields": ["orderId"],
    "valueMismatches": {
      "status": {"expected": "pending", "actual": "processing"}
    },
    "headerMismatches": {
      "Content-Type": {"expected": "application/json", "actual": "text/plain"}
    },
    "schemaViolations": [
      {"field": "amount", "message": "value must be a number"}
    ]
  }
}
```

**Reading the output:**
- `missingFields` — fields in `assert_contents_pattern` that were absent from the real response
- `valueMismatches` — fields present but with the wrong value (expected vs. actual)
- `headerMismatches` — same for response headers
- `schemaViolations` — fields that violate the OpenAPI schema (only when `spec_content` is provided)
- `statusCode` vs. `expectedStatusCode` — quick check when the wrong HTTP status was returned

The `errors` map (legacy) still contains a plain-text summary for backward compatibility.

→ [Contract Testing — Field-Level Diagnostics](contract-testing.md), [API Reference — Response Format](api-reference.md)

---

## 16. Track Which API Paths Were Exercised (Coverage)

**Problem:** I want to know which OpenAPI paths my contract tests covered and which ones were missed.

```bash
# Run contracts with coverage tracking (requires --spec)
api-mock-service producer-contract \
  --group my-service \
  --base_url https://api.example.com \
  --spec openapi.yaml \
  --track-coverage \
  --times 10
```

**CLI output:**
```
COVERAGE REPORT
──────────────────────────────────────────────────────────────
Overall: 87.5%  (7/8 paths)

Uncovered paths:
  ✗ DELETE /users/:id

Method coverage:
  GET    100.0%
  POST    75.0%
  DELETE   0.0%
```

**Retrieve the last coverage report via HTTP (without re-running):**
```bash
curl http://localhost:8080/_coverage/my-service
```
```json
{
  "totalPaths": 8,
  "coveredPaths": 7,
  "coveragePercentage": 87.5,
  "uncoveredPaths": ["DELETE /users/:id"],
  "methodCoverage": {"GET": 100.0, "POST": 75.0, "DELETE": 0.0}
}
```

**HTTP (inline spec):**
```bash
curl -X POST http://localhost:8080/_contracts/my-service \
  -d "{\"base_url\":\"https://api.example.com\",
       \"track_coverage\":true,
       \"spec_content\":$(cat openapi.yaml | jq -Rs .)}"
```

→ [Contract Testing — Coverage Reporting](contract-testing.md), [API Reference — Coverage Endpoint](api-reference.md)

---

## 17. Run Mutation Testing to Find Validation Gaps

**Problem:** I want to verify that the real API correctly rejects malformed, null, boundary, and security-injection inputs.

```bash
# HTTP
curl -X POST http://localhost:8080/_contracts/mutations/my-service \
  -H "Content-Type: application/json" \
  -d '{"base_url": "https://api.example.com", "execution_times": 1}'

# CLI
api-mock-service producer-contract \
  --group my-service \
  --base_url https://api.example.com \
  --mutations
```

**What gets generated for each scenario:**

| Strategy | What it tests |
|----------|--------------|
| **Null fields** | Each field set to `null` — expects 400/422 |
| **Combinatorial** | Pairs of (boundary value, null field) — up to 10 pairs |
| **Format boundary** | Invalid dates, UUIDs, emails, URIs |
| **Boundary values** | Min: `MinInt32`, `""` — Max: `MaxInt32`, 255-char string |
| **Security injection** | SQLi, path traversal, LDAP, command injection, SSRF, XXE, XSS |

**Example output:**
```
──────────────────────────────────────────────────────────────
SCENARIO                                 STATUS
──────────────────────────────────────────────────────────────
create-order-null-email_0                ✓ PASS  (API returned 422)
create-order-boundary-min_0              ✓ PASS  (API returned 400)
create-order-sqli-customerId_0           ✗ FAIL  (API returned 200 — accepted injection)
──────────────────────────────────────────────────────────────
TOTAL 42  Passed: 41  Failed: 1
```

**Combine with schema validation:**
```bash
api-mock-service producer-contract \
  --group my-service \
  --base_url https://api.example.com \
  --mutations \
  --spec openapi.yaml \
  --track-coverage
```

→ [Contract Testing — Mutation Testing](contract-testing.md), [Fuzz & Property Testing](fuzz-property-testing.md)

---

## 18. Shrink a Failing Mutation to Its Minimal Reproducer

**Problem:** A mutation test failed with a large payload. I want to find the smallest input that still triggers the failure.

```bash
api-mock-service producer-contract \
  --group payments \
  --base_url https://api.example.com \
  --mutations \
  --shrink
```

**Example output:**
```
SHRINK ANALYSIS
──────────────────────────────────────────────────────────────
Shrinking POST /payments-sqli-amount_0 ...
  Original body: {"amount":"' OR 1=1; --","currency":"USD","customerId":"cust-99"}
  ✓ Reduced in 14 attempts
  Minimal body:  {"amount":"' OR 1=1; --"}
──────────────────────────────────────────────────────────────
```

**What this tells you:** The API accepts SQL injection in `amount` regardless of other fields. The minimal reproducer is the exact payload to include in a bug report.

**Strategies (run in order):**
1. **Field removal** — delta debugging: remove one field at a time, keep if failure persists
2. **String shortening** — binary-search string length to minimum that still fails
3. **Array shrinking** — remove elements one at a time
4. **Numeric reduction** — halve large boundary values until failure stops

→ [Contract Testing — Fuzz Shrinking](contract-testing.md), [Fuzz & Property Testing — Fuzz Shrinking](fuzz-property-testing.md)

---

## 19. Test Stateful Workflows (Session-Scoped State Machine)

**Problem:** I need to test a CREATE → READ → UPDATE → DELETE workflow where each step depends on the previous.

**Step 1: Define scenarios with a `state_machine` block**

```yaml
# create-order.yaml
name: create-order
method: POST
path: /orders
group: order-lifecycle
response:
  status_code: 201
  contents: '{"orderId":"{{UUID}}","status":"pending"}'
state_machine:
  transitions:
    - from: ""          # any starting state (new session)
      to: "created"
      on_method: POST
      on_status: 201
      extract_key: "$.orderId"   # saves orderId into session store
```

```yaml
# get-order.yaml
name: get-order
method: GET
path: /orders/:id
group: order-lifecycle
response:
  status_code: 200
  contents: '{"orderId":"{{.orderId}}","status":"pending"}'
state_machine:
  initial_state: "created"       # only matches when session is in this state
  transitions:
    - from: "created"
      to: "viewed"
      on_method: GET
      on_status: 200
```

```yaml
# delete-order.yaml
name: delete-order
method: DELETE
path: /orders/:id
group: order-lifecycle
response:
  status_code: 204
  contents: ""
state_machine:
  initial_state: "viewed"
  transitions:
    - from: "viewed"
      to: "deleted"
      on_method: DELETE
      on_status: 204
```

**Step 2: Run the workflow — same `X-Session-ID` ties the steps together**

```bash
SESSION="session-$(date +%s)"

# CREATE → state transitions to "created", orderId extracted
curl -s -X POST http://localhost:8080/orders \
  -H "X-Session-ID: $SESSION" \
  -H "Content-Type: application/json" \
  -d '{"customerId":"cust-1","amount":99}' | jq .
# {"orderId":"abc-123","status":"pending"}

# READ → only matches in "created" state → transitions to "viewed"
curl -s http://localhost:8080/orders/abc-123 \
  -H "X-Session-ID: $SESSION" | jq .
# {"orderId":"abc-123","status":"pending"}

# DELETE → only matches in "viewed" state → transitions to "deleted"
curl -s -X DELETE http://localhost:8080/orders/abc-123 \
  -H "X-Session-ID: $SESSION"
# 204 No Content

# READ again → no scenario matches "deleted" state → 404
curl -s http://localhost:8080/orders/abc-123 \
  -H "X-Session-ID: $SESSION"
# 404 Not Found
```

Each unique `X-Session-ID` gets its own isolated state — parallel tests never interfere.

→ [Contract Testing — Stateful Workflows](contract-testing.md), [API Reference — X-Session-ID](api-reference.md)

---

## 20. Detect Breaking API Changes in CI

**Problem:** I want CI to fail when a new OpenAPI spec introduces a breaking change.

**CLI (for CI pipelines):**
```bash
api-mock-service compare-specs \
  --base api/v1.yaml \
  --head api/v2.yaml \
  --fail-on-breaking   # exits with code 2 if breaking changes found
```

**Output:**
```
SPEC DIFF REPORT
──────────────────────────────────────────────────────────────
Removed paths (BREAKING):
  - /orders/batch

Breaking changes:
  ✗ [GET /users] type-change: id integer → string
  ✗ [POST /orders] new-required-param: "currency" is now required

Non-breaking changes:
  ~ [GET /products] added-path
──────────────────────────────────────────────────────────────
breaking=2  non-breaking=1  added-paths=1  removed-paths=1
```

**HTTP endpoint (returns 409 on breaking changes — useful in webhook/scripted CI):**
```bash
curl -X POST http://localhost:8080/_oapi/diff \
  -H "Content-Type: application/json" \
  -d "{\"base\": $(cat v1.yaml | jq -Rs .), \"head\": $(cat v2.yaml | jq -Rs .)}"
# 200 OK → no breaking changes
# 409 Conflict → breaking changes (body contains full diff)
```

**GitHub Actions integration:**
```yaml
- name: Check for breaking API changes
  run: |
    api-mock-service compare-specs \
      --base main/openapi.yaml \
      --head pr/openapi.yaml \
      --fail-on-breaking
```

**Breaking change categories:** removed path/method, optional param promoted to required, new required param, field type change, response field removed, enum narrowed, format change.

→ [Contract Testing — Spec Diff](contract-testing.md), [CLI Reference — compare-specs](cli-reference.md), [API Reference — POST /_oapi/diff](api-reference.md)

---

## 21. Preview Scenarios Without Executing (Dry Run)

**Problem:** I want to verify which scenarios are registered before running the full test suite, especially in CI.

```bash
# CLI
api-mock-service producer-contract \
  --group payments \
  --base_url https://api.example.com \
  --dry-run

# HTTP
curl -X POST http://localhost:8080/_contracts/payments \
  -d '{"base_url":"https://api.example.com","dry_run":true}'
```

**Output:**
```
DRY RUN — no requests will be sent
──────────────────────────────────────────────────────────────
SCENARIO                              METHOD  PATH
──────────────────────────────────────────────────────────────
create-payment-201                    POST    /payments
get-payment-200                       GET     /payments/:id
refund-payment-200                    POST    /payments/:id/refund
cancel-payment-200                    DELETE  /payments/:id
──────────────────────────────────────────────────────────────
4 scenarios would run
```

**Use in CI to gate on scenario existence before running the full suite:**
```yaml
- name: Validate scenario inventory
  run: |
    api-mock-service producer-contract \
      --group payments --base_url https://api.example.com --dry-run
- name: Run contract tests
  run: |
    api-mock-service producer-contract \
      --group payments --base_url https://api.example.com --times 5
```

→ [Contract Testing — Dry Run](contract-testing.md), [CLI Reference](cli-reference.md)

---

## 22. Inject Chaos: Errors and Latency

**Problem:** I want to test how my application behaves when a downstream API is unreliable.

**Method A: Group chaos config (no YAML changes needed)**

```bash
curl -X PUT http://localhost:8080/_groups/payments/config \
  -H "Content-Type: application/json" \
  -d '{
    "chaos_enabled": true,
    "mean_time_between_failure": 5,
    "mean_time_between_additional_latency": 4,
    "max_additional_latency_secs": 2.5,
    "http_errors": [500, 502, 503]
  }'
```

Effect: ~1 in 5 requests returns a random error; ~1 in 4 requests gets up to 2.5s extra latency.

**Method B: Template-based conditionals in the scenario**

```yaml
response:
  {{if NthRequest 3}}
  status_code: 503
  contents: '{"error": "service temporarily unavailable"}'
  wait_before_reply: 2s
  {{else}}
  status_code: 200
  contents: '{"status": "ok", "latency": "{{RandIntMinMax 10 200}}ms"}'
  {{end}}
```

**Method C: Multiple scenarios on the same path (round-robin)**

```yaml
# success (name: get-payment-ok)
response: {status_code: 200, contents: '{"status":"paid"}'}

# failure (name: get-payment-timeout — same path/method)
response: {status_code: 504, contents: '{"error":"gateway timeout"}'}
```

The mock service rotates between matching scenarios automatically.

→ [Mock Guide — Chaos Testing](mock-guide.md)

---

## 23. Serve Static Files and Binary Fixtures

**Problem:** My API returns binary files (PDFs, images) or I want to use large fixture data in templates.

**Upload a fixture:**
```bash
# Text fixture (lines usable via RandFileLine / SeededFileLine)
curl -H "Content-Type: text/plain" \
  --data-binary @product-names.txt \
  http://localhost:8080/_fixtures/GET/product-names.txt/products

# YAML property fixture (usable via FileProperty)
curl -H "Content-Type: application/yaml" \
  --data-binary @config.yaml \
  http://localhost:8080/_fixtures/GET/config.yaml/products

# Binary file (returned directly as response body)
curl -H "Content-Type: image/png" \
  --data-binary @logo.png \
  http://localhost:8080/_fixtures/GET/logo.png/images/logo
```

**Use fixtures in a scenario:**
```yaml
# Text fixture — random or seeded line
contents: '{"productName": "{{RandFileLine "product-names.txt"}}"}'

# YAML property file
contents: '{"token": {{FileProperty "config.yaml" "apiToken"}}}'

# Serve binary file directly (no template body)
response:
  status_code: 200
  content_type: image/png
  contents_file: logo.png
```

**Retrieve a fixture:**
```bash
curl http://localhost:8080/_fixtures/GET/logo.png/images/logo
```

**Serve static assets (no upload needed — drop files in `ASSET_DIR`):**
```bash
cp report.pdf default_assets/
curl http://localhost:8080/_assets/default_assets/report.pdf
```

→ [Mock Guide — Fixtures](mock-guide.md)

---

## 24. Share Variables Across All Scenarios in a Group

**Problem:** I want to inject environment-specific values (base URLs, API keys, feature flags) into all scenarios without editing each file.

```bash
# Set group-level variables (injected as template params in all scenarios in the group)
curl -X PUT http://localhost:8080/_groups/payments/config \
  -H "Content-Type: application/json" \
  -d '{
    "variables": {
      "apiVersion": "v2",
      "region": "us-east-1",
      "featureFlag": "enabled"
    }
  }'

# Use "global" to share across ALL groups
curl -X PUT http://localhost:8080/_groups/global/config \
  -d '{"variables": {"environment": "staging", "supportEmail": "ops@example.com"}}'
```

Scenarios can reference these as `{{.apiVersion}}`, `{{.region}}`, `{{.environment}}`, etc.:

```yaml
contents: >
  {
    "version": "{{.apiVersion}}",
    "region":  "{{.region}}",
    "support": "{{.supportEmail}}"
  }
```

→ [API Reference — Group Management](api-reference.md), [Mock Guide](mock-guide.md)

---

## 25. Use Deterministic / Seeded Test Data

**Problem:** I want mock responses to return the same data every run so snapshot tests and recordings stay stable.

Use `Seeded*` template functions instead of `Rand*`:

```yaml
contents: >
  {
    "id":        "{{SeededUUID 42}}",
    "name":      "{{SeededName 0}}",
    "city":      "{{SeededCity 1}}",
    "random":    {{SeededRandom 7}},
    "active":    {{SeededBool 3}}
  }
```

**Same seed → same value on every run:**
```bash
# First call
curl http://localhost:8080/users/1
# {"id":"...-42-fixed-uuid","name":"Michael Chen","city":"Denver",...}

# Second call — identical
curl http://localhost:8080/users/1
# {"id":"...-42-fixed-uuid","name":"Michael Chen","city":"Denver",...}
```

**Seeded functions reference:**

| Function | Description |
|----------|-------------|
| `{{SeededUUID N}}` | Deterministic UUID v4 from seed N |
| `{{SeededName N}}` | Deterministic full name |
| `{{SeededCity N}}` | Deterministic city |
| `{{SeededRandom N}}` | Deterministic float 0–1 |
| `{{SeededBool N}}` | Deterministic boolean |
| `{{SeededFileLine "file.txt" N}}` | Line N from a fixture file |

→ [Mock Guide — Template Functions](mock-guide.md)

---

## Related Docs

| Guide | What it covers |
|-------|----------------|
| [Mock Guide](mock-guide.md) | Recording, playback, templates, fixtures, chaos |
| [Contract Testing](contract-testing.md) | Consumer + producer contracts, JSONPath, schema validation, stateful workflows, mutations, coverage, spec diff |
| [Fuzz & Property Testing](fuzz-property-testing.md) | Mutation strategies, security injection, shrinking, HAR/Postman import |
| [OpenAPI Guide](openapi-guide.md) | Spec upload, discriminator/oneOf/anyOf, Swagger UI, coverage |
| [API Reference](api-reference.md) | Every HTTP endpoint with request/response shapes |
| [CLI Reference](cli-reference.md) | Every command and flag |
