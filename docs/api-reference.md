# API Reference

All endpoints are served on **port 8080** (default). The proxy recorder runs on **port 8081**.

Base URL: `http://localhost:8080`

---

## Mock Playback

### `* /:path`

Any HTTP method — GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS — on any path is matched against stored scenarios. The mock service selects the best-matching scenario and returns the rendered response.

**Match priority:**
1. Exact method + path + `X-Mock-Scenario` header
2. Method + path + parameter matching (`assert_headers_pattern`, `assert_query_params_pattern`, `assert_contents_pattern`)
3. Method + path (round-robin across matching scenarios)
4. 404 if no match

**Request headers:**

| Header | Description |
|--------|-------------|
| `X-Mock-Scenario: <name>` | Select specific scenario by name |
| `X-Mock-Response-Status: 503` | Override response status code |
| `X-Mock-Wait-Before-Reply: 2s` | Inject artificial delay |
| `X-Session-ID: <id>` | Session identifier for stateful workflows — enables state machine transitions across requests (see [Contract Testing](contract-testing.md#stateful-workflows)) |

**Response headers (always present):**

| Header | Description |
|--------|-------------|
| `X-Mock-Path` | Path template matched |
| `X-Mock-Scenario` | Scenario name selected |
| `X-Mock-Request-Count` | Number of times this scenario has been called |

---

## Scenario Management

### `POST /_scenarios`

Upload a YAML scenario file.

**Request:**
```
Content-Type: application/yaml
Body: <YAML scenario>
```

**Response:** `200 OK`
```json
{"name": "my-scenario", "path": "/v1/orders/:id", "method": "GET"}
```

**Example:**
```bash
curl -H "Content-Type: application/yaml" \
  --data-binary @my-scenario.yaml \
  http://localhost:8080/_scenarios
```

---

### `GET /_scenarios`

List all stored scenarios (summary).

**Response:** `200 OK`
```json
{
  "/_scenarios/GET/my-scenario/v1/orders/:id": {
    "method": "GET",
    "name": "my-scenario",
    "path": "/v1/orders/:id",
    "assert_query_params_pattern": {},
    "assert_headers_pattern": {},
    "LastUsageTime": 1672531200,
    "RequestCount": 42
  }
}
```

---

### `GET /_scenarios/:method/:name/:path`

Get a specific scenario by method, name, and path.

**Example:**
```bash
curl http://localhost:8080/_scenarios/GET/my-scenario/v1/orders/:id
```

---

### `DELETE /_scenarios/:method/:name/:path`

Delete a specific scenario.

**Example:**
```bash
curl -X DELETE http://localhost:8080/_scenarios/GET/my-scenario/v1/orders/:id
```

---

## Group Management

### `GET /_groups`

List all scenario groups.

**Response:** `200 OK` — array of group names.

---

### `GET /_groups/:group`

Get all scenarios in a group.

---

### `PUT /_groups/:group/config`

Set group variables and chaos configuration.

**Request body:**
```json
{
  "variables": {
    "env": "staging",
    "apiVersion": "v2"
  },
  "chaos_enabled": false,
  "mean_time_between_failure": 5,
  "mean_time_between_additional_latency": 4,
  "max_additional_latency_secs": 2.5,
  "http_errors": [400, 500, 503]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `variables` | `map[string]string` | Variables injected into all scenarios in this group |
| `chaos_enabled` | bool | Enable chaos injection |
| `mean_time_between_failure` | int | ~1/N requests will get an HTTP error |
| `mean_time_between_additional_latency` | int | ~1/N requests will get extra latency |
| `max_additional_latency_secs` | float | Max latency to add (seconds) |
| `http_errors` | `[]int` | HTTP status codes to return on error injection |

Use `global` as the group name to share variables across all scenarios.

---

## OpenAPI

### `POST /_oapi`

Upload an OpenAPI 3.x spec (YAML or JSON). Generates scenario files for every path × method × status code × discriminator variant.

**Request:**
```
Content-Type: application/yaml   (or application/json)
Body: <OpenAPI spec>
```

**Response:** `200 OK`
```json
{"scenarios": 42, "updated": "<modified spec bytes>"}
```

**Example:**
```bash
curl -H "Content-Type: application/yaml" \
  --data-binary @openapi.yaml \
  http://localhost:8080/_oapi
```

---

### `GET /_oapi`

Download the most recently uploaded spec.

---

## Contract Testing

All contract endpoints accept a `ProducerContractRequest` body:

```json
{
  "base_url": "https://api.example.com",
  "execution_times": 5,
  "verbose": false,
  "track_coverage": false,
  "run_mutations": false,
  "spec_content": ""
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `base_url` | string | — | Target API base URL |
| `execution_times` | int | 5 | Runs per scenario |
| `verbose` | bool | false | Log request/response details |
| `track_coverage` | bool | false | Include coverage report in response |
| `run_mutations` | bool | false | Run in mutation mode |
| `spec_content` | string | — | Inline OpenAPI YAML/JSON for schema validation |
| `dry_run` | bool | false | List scenarios that would run without executing them |

**Response format:**

```json
{
  "results": {
    "<scenario-name>_<iteration>": { /* response body fields */ }
  },
  "errors": {
    "<scenario-name>_<iteration>": "error message"
  },
  "error_details": {
    "<scenario-name>_<iteration>": {
      "summary": "assertion failed",
      "scenario": "get-user",
      "url": "https://api.example.com/users/42",
      "statusCode": 200,
      "expectedStatusCode": 201,
      "missingFields": ["id"],
      "valueMismatches": {
        "status": {"expected": "active", "actual": "pending"}
      },
      "headerMismatches": {},
      "schemaViolations": [
        {"field": "email", "message": "value is required"}
      ]
    }
  },
  "succeeded": 8,
  "failed": 1,
  "mismatched": 0,
  "coverage": {
    "totalPaths": 8,
    "coveredPaths": 7,
    "coveragePercentage": 87.5,
    "uncoveredPaths": ["DELETE /users/:id"],
    "methodCoverage": {"GET": 100.0, "DELETE": 0.0}
  },
  "metrics": {}
}
```

---

### `POST /_contracts/:group`

Execute producer contract tests for all scenarios in a group.

**Example:**
```bash
curl -X POST http://localhost:8080/_contracts/todos \
  -H "Content-Type: application/json" \
  -d '{"base_url": "https://jsonplaceholder.typicode.com", "execution_times": 3}'
```

---

### `POST /_contracts/:method/:name/:path`

Execute a single scenario by method, name, and path.

**Example:**
```bash
curl -X POST http://localhost:8080/_contracts/GET/get-todo/todos/:id \
  -d '{"base_url": "https://jsonplaceholder.typicode.com"}'
```

---

### `POST /_contracts/history/:group`

Execute contracts by previous execution history. Omit `:group` to run all history.

**Example:**
```bash
curl -X POST http://localhost:8080/_contracts/history/todos \
  -d '{"base_url": "https://jsonplaceholder.typicode.com"}'
```

---

### `POST /_contracts/mutations/:group`

Generate and execute mutation variants for all scenarios in a group. Tests API robustness against null fields, boundary values, format violations, and security payloads.

**Example:**
```bash
curl -X POST http://localhost:8080/_contracts/mutations/my-api \
  -H "Content-Type: application/json" \
  -d '{"base_url": "https://api.example.com", "execution_times": 1}'
```

Mutation strategies: null fields, combinatorial pairs, format boundary (date/uuid/email/uri), boundary values (min+max), security injection (SQLi/path traversal/LDAP/XXE/SSRF/command injection).

See [Fuzz & Property Testing](fuzz-property-testing.md) for details.

---

### `GET /_coverage/:group`

Returns the last coverage summary from the most recent `ExecuteByGroup` run for the group. Requires that the run used `track_coverage: true` with a spec.

**Example:**
```bash
curl http://localhost:8080/_coverage/my-api
```

**Response (coverage available):**
```json
{
  "totalPaths": 8,
  "coveredPaths": 7,
  "coveragePercentage": 87.5,
  "uncoveredPaths": ["DELETE /users/:id"],
  "methodCoverage": {"GET": 100.0, "POST": 75.0}
}
```

**Response (no data yet):**
```json
{
  "group": "my-api",
  "message": "no coverage data available — run producer-contract with track_coverage:true and spec_content first"
}
```

---

### `POST /_oapi/diff`

Compare two OpenAPI specs and report breaking and non-breaking changes. Returns **409 Conflict** when breaking changes are detected (CI-friendly).

**Request body:**
```json
{
  "base": "<OpenAPI YAML or JSON string>",
  "head": "<OpenAPI YAML or JSON string>"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/_oapi/diff \
  -H "Content-Type: application/json" \
  -d "{\"base\": $(cat v1.yaml | jq -Rs .), \"head\": $(cat v2.yaml | jq -Rs .)}"
```

**Response (200 OK — no breaking changes):**
```json
{
  "breakingChanges": [],
  "nonBreakingChanges": [{"path": "/products", "changeType": "added-path", "severity": "non-breaking"}],
  "addedPaths": ["/products"],
  "removedPaths": []
}
```

**Response (409 Conflict — breaking changes detected):**
```json
{
  "breakingChanges": [
    {"path": "/users", "method": "GET", "field": "id", "changeType": "type-change", "before": "integer", "after": "string", "severity": "breaking"}
  ],
  "nonBreakingChanges": [],
  "addedPaths": [],
  "removedPaths": ["/orders/batch"]
}
```

---

## History

### `POST /_history/:method/:name/:path`

Save an execution result to history.

---

### `GET /_history`

List execution history.

---

### `GET /_history/har`

Download execution history as HAR format.

---

### `POST /_history/har`

Import scenarios from a HAR file. Response bodies are automatically analyzed to generate type-aware `assert_contents_pattern` assertions — no manual configuration required.

**Example:**
```bash
curl -X POST http://localhost:8080/_history/har \
  --data-binary @my-recording.har
```

---

### `GET /_history/postman`

Download execution history as Postman Collection format.

---

### `POST /_history/postman`

Import scenarios from a Postman Collection. Response bodies are automatically analyzed to generate type-aware `assert_contents_pattern` assertions — no manual configuration required.

**Example:**
```bash
curl -X POST http://localhost:8080/_history/postman \
  -H "Content-Type: */*" \
  --data-binary @collection.json
```

---

## Fixtures

### `GET /_fixtures/:method/:name/:path`

Retrieve a fixture file.

---

### `POST /_fixtures/:method/:name/:path`

Upload a fixture file. The fixture becomes available to templates as `{{ FileProperty "name" "key" }}` etc.

**Example:**
```bash
# Upload text fixture (available in templates as SeededFileLine "lines.txt" $n)
curl -H "Content-Type: text/plain" \
  --data-binary @lines.txt \
  http://localhost:8080/_fixtures/GET/lines.txt/devices

# Upload YAML fixture
curl -H "Content-Type: application/yaml" \
  --data-binary @props.yaml \
  http://localhost:8080/_fixtures/GET/props.yaml/devices

# Upload binary image
curl -H "Content-Type: image/png" \
  --data-binary @logo.png \
  http://localhost:8080/_fixtures/GET/logo.png/images/logo
```

---

## Proxy

### `GET|POST|PUT|DELETE|PATCH /_proxy`

Forward a request to a remote API and record the interaction.

**Request headers:**

| Header | Required | Description |
|--------|----------|-------------|
| `X-Mock-Url` | yes | Full URL of the real API endpoint to call |

**Example:**
```bash
curl -H "X-Mock-Url: https://api.stripe.com/v1/customers/cus_xxx/cash_balance" \
     -H "Authorization: Bearer sk_test_xxx" \
     http://localhost:8080/_proxy
```

---

## UI & Health

### `GET /_ui`

Embedded Swagger UI. Browse to `http://localhost:8080/_ui` in your browser.

---

### `GET /_health`

Health check endpoint.

**Response:** `200 OK` `{"status": "ok"}`

---

## Proxy Recorder (Port 8081)

Set your HTTP client proxy to `http://localhost:8081`. All traffic is forwarded to the real server and recorded as scenarios automatically.

```bash
export http_proxy="http://localhost:8081"
export https_proxy="http://localhost:8081"
```

You may need to disable TLS verification (`curl -k`) or import `ca_cert.pem` for HTTPS recording.

---

## Related Docs

- [Mock Guide](mock-guide.md) — recording, playback, templates
- [Contract Testing](contract-testing.md) — contract execution and options
- [OpenAPI Guide](openapi-guide.md) — spec upload and schema validation
- [CLI Reference](cli-reference.md) — command-line interface
