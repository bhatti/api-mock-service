# Building Reliable APIs with Mock, Contract, and Property-Based Testing — A Practical Guide

*I've spent years debugging integration failures that only showed up in production. Tests passed. CI was green. The real API returned something subtly different from what we expected, and nobody caught it until a customer did. If that story sounds familiar, this guide is for you.*

---

## The Problem with API Testing Today

Most teams test APIs one of three ways:

1. **Unit tests with mocks** — fast, but they test your *assumption* of the API, not the API itself. When the real service changes, your unit tests still pass.

2. **Integration tests against real services** — they catch real bugs, but they're slow, flaky, and require a running environment. You can't reproduce that 500 error from three weeks ago.

3. **Contract testing tools (Pact, Pactflow)** — solid consumer-driven contracts, but they require a separate broker, language-specific libraries in every service, and significant DevOps investment to stand up and maintain.

The result: teams pick one approach, accept its limitations, and live with the gap.

What if a single tool could record real traffic, play it back as a mock, generate fuzz data from specs, validate real API responses against schemas, and run mutation testing — all without touching application code?

That's exactly what [api-mock-service](https://github.com/bhatti/api-mock-service) does.

---

## What Makes This Different

```mermaid
graph LR
    subgraph "Traditional Approach"
        T1["Postman\n(manual tests)"]:::gray
        T2["Pactflow\n(consumer contracts)"]:::gray
        T3["WireMock\n(stubbing)"]:::gray
        T4["Schemathesis\n(fuzz testing)"]:::gray
        T5["Dredd\n(schema validation)"]:::gray
    end

    subgraph "api-mock-service"
        A["Single Go Binary\nRecord · Mock · Contract · Fuzz · Mutate"]:::blue
    end

    classDef gray fill:#7F8C8D,color:#fff,stroke:#566573
    classDef blue fill:#4A90D9,color:#fff,stroke:#2C5F8A
```

Where Postman gives you manual tests and Pactflow gives you consumer contracts, api-mock-service gives you **all five capabilities in one tool with zero infrastructure**:

| Capability | Postman | Pactflow | WireMock | Schemathesis | **api-mock-service** |
|-----------|---------|----------|----------|-------------|----------------------|
| Proxy recording | ✓ | — | — | — | **✓** |
| Mock playback | ✓ | — | ✓ | — | **✓** |
| Consumer contracts | — | ✓ | — | — | **✓** |
| Producer contracts | ✓ | ✓ | — | — | **✓** |
| OpenAPI fuzz generation | — | — | — | ✓ | **✓** |
| Schema validation | — | — | — | ✓ | **✓** |
| Mutation testing | — | — | — | ✓ | **✓** |
| Security injection | — | — | — | partial | **✓** |
| No broker / agent needed | — | — | ✓ | ✓ | **✓** |
| Single binary | — | — | — | ✓ | **✓** |

No Pact broker to run. No language-specific libraries to install. No Kubernetes operator. One binary, YAML files, and a data directory.

---

## Architecture in 30 Seconds

```mermaid
flowchart TD
    subgraph dev["Your Development Workflow"]
        R["Record\n(proxy port 8081)"]:::amber
        M["Mock\n(api port 8080)"]:::blue
        C["Contract Test\n(api port 8080)"]:::green
        F["Fuzz + Mutate\n(api port 8080)"]:::red
    end

    Proxy["Port 8081\nProxy Recorder"]:::amber
    API["Port 8080\nMock + Contract + OpenAPI"]:::blue
    Store[("YAML Scenarios\n+ Fixtures")]:::storage
    Real["Real API"]:::gray

    R -->|"route traffic through"| Proxy
    Proxy -->|"forward + record"| Real
    Proxy -->|"save"| Store

    M & C & F -->|"drive"| API
    API -->|"read/write"| Store

    classDef amber fill:#F39C12,color:#fff,stroke:#B7770D
    classDef blue fill:#4A90D9,color:#fff,stroke:#2C5F8A
    classDef green fill:#27AE60,color:#fff,stroke:#1A7A42
    classDef red fill:#E74C3C,color:#fff,stroke:#A93226
    classDef storage fill:#8E44AD,color:#fff,stroke:#6C3483
    classDef gray fill:#7F8C8D,color:#fff,stroke:#566573
```

Two ports do everything:
- **Port 8081** — transparent HTTP/HTTPS proxy. Route your curl or browser through it and every interaction is recorded as a YAML scenario.
- **Port 8080** — the main API. Plays back mocks, accepts OpenAPI uploads, runs contract tests, serves Swagger UI.

---

## Part 1: Getting Started in 5 Minutes

### Install

```bash
# Docker (quickest)
docker run -p 8080:8080 -p 8081:8081 \
  -e DATA_DIR=/tmp/mocks \
  plexobject/api-mock-service:latest

# Or build from source
git clone https://github.com/bhatti/api-mock-service
cd api-mock-service && make
./out/bin/api-mock-service
```

The server starts immediately. No config file needed. No database. Scenarios are YAML files on disk — `git add` them and your entire mock suite travels with your repo.

### Your First Recording

Let's record a real API call. I'll use JSONPlaceholder as a stand-in for your real service:

```bash
export http_proxy="http://localhost:8081"
export https_proxy="http://localhost:8081"

curl -k -X POST https://jsonplaceholder.typicode.com/todos \
  -H "Content-Type: application/json" \
  -d '{"userId": 1, "title": "write better tests", "completed": false}'
```

That's it. The service recorded everything — request shape, response body, headers — and saved it as:

```
default_mocks_data/todos/POST/recorded-scenario-<hash>.scr
```

Open that file and you'll see:

```yaml
method: POST
name: recorded-todos-<hash>
path: /todos
group: todos
request:
  assert_headers_pattern:
    Content-Type: application/json
  assert_contents_pattern: >
    {"completed":"(__boolean__(false|true))",
     "title":"(__string__\\w+)",
     "userId":"(__number__[+-]?[0-9]{1,10})"}
response:
  status_code: 201
  contents: |-
    {"id": 201, "userId": 1, "title": "write better tests", "completed": false}
  assert_contents_pattern: >
    {"id":"(__number__[+-]?[0-9]{1,10})"}
```

Notice what happened automatically: the tool looked at the response body, inferred the types, and wrote `assert_contents_pattern` for you. The `__number__` and `__boolean__` tokens are property assertions — they don't check the exact value, they check the *shape*. This is property-based testing baked into the recording step.

### Play It Back

Disconnect from the proxy and replay:

```bash
unset http_proxy https_proxy
curl http://localhost:8080/todos
```

Returns the recorded response. No network call. No external dependency. Works offline, in CI, and on a plane.

---

## Part 2: Dynamic Templates — More Than Just Replay

Static replay is table stakes. What makes api-mock-service genuinely powerful is the template engine.

### Making Responses Dynamic

Edit the recorded YAML to add path variables and template expressions:

```yaml
method: GET
name: get-todo
path: /todos/:id
group: todos
request:
  assert_headers_pattern:
    Authorization: "Bearer [A-Za-z0-9]{20,}"
response:
  status_code: 200
  contents: |-
    {
      "id": {{.id}},
      "userId": {{RandIntMinMax 1 10}},
      "title": "{{RandSentence 3 8}}",
      "completed": {{RandBool}},
      "createdAt": "{{Time}}",
      "email": "{{RandEmail}}"
    }
```

Upload and play:

```bash
curl -H "Content-Type: application/yaml" \
  --data-binary @get-todo.yaml \
  http://localhost:8080/_scenarios

curl -H "Authorization: Bearer abc123xyz456def789abc123" \
  http://localhost:8080/todos/42
```

Returns:
```json
{
  "id": 42,
  "userId": 7,
  "title": "nisi ut aliquid corporis",
  "completed": true,
  "createdAt": "2024-01-15T14:32:01-07:00",
  "email": "alice.chen@example.com"
}
```

Every call returns different data. Same shape, different values — exactly what you need to test how your client handles variable inputs.

### Generate 1000 Records

Property-based testing loves large samples. Generate them in one response:

```yaml
method: GET
name: list-orders
path: /orders
response:
  contents: >
    {"orders": [
      {{- range $i := Iterate .pageSize }}
      {
        "id": "ord-{{RandRegex "[0-9]{6}"}}",
        "customerId": "cust-{{SeededUUID $i}}",
        "amount": {{RandFloatMinMax 1.0 9999.99}},
        "status": {{EnumString "pending processing shipped delivered"}},
        "createdAt": "{{Time}}"
      }{{if LastIter $i $.pageSize}}{{else}},{{end}}
      {{end}}
    ],
    "total": {{.pageSize}},
    "page": {{.page}}}
  status_code: 200
```

```bash
curl "http://localhost:8080/orders?pageSize=1000&page=1"
```

1000 randomized orders, consistently shaped, no external service needed.

### Chaos Injection

Real services fail. Your client needs to handle that. Inject failures without touching code:

```yaml
response:
  {{if NthRequest 5}}
  status_code: {{EnumInt 500 503}}
  contents: '{"error": "service temporarily unavailable"}'
  {{else if GERequest 20}}
  status_code: 429
  contents: '{"error": "rate limit exceeded"}'
  {{else}}
  status_code: 200
  contents: '{"id": {{.id}}, "status": "active"}'
  {{end}}
  wait_before_reply: {{RandIntMinMax 0 3}}s
```

Every 5th request returns a 500. After 20 total requests, you start getting 429s. And there's a random 0–3 second latency on every call. Your client's retry logic, circuit breaker, and timeout handling get a real workout — no code changes, no special test environment.

Or configure it at the group level:

```bash
curl -X PUT http://localhost:8080/_groups/payment-service/config \
  -H "Content-Type: application/json" \
  -d '{
    "chaos_enabled": true,
    "mean_time_between_failure": 5,
    "mean_time_between_additional_latency": 4,
    "max_additional_latency_secs": 2.0,
    "http_errors": [500, 502, 503]
  }'
```

Now every scenario in the `payment-service` group will randomly inject failures without changing a single YAML file.

---

## Part 3: OpenAPI → Instant Test Suite

If you have an OpenAPI spec, you have a complete test suite in 30 seconds.

### Upload a Spec

```bash
curl -H "Content-Type: application/yaml" \
  --data-binary @openapi.yaml \
  http://localhost:8080/_oapi
```

The parser walks every `path × method × status code × schema variant` and generates ready-to-use YAML scenarios. For a typical API with 20 endpoints and 3 status codes each, that's 60+ scenarios generated automatically — all with realistic fuzz data and type-checked assertions.

### Discriminator / Polymorphic APIs

Real APIs are polymorphic. A `POST /animals` might accept a `Cat` or a `Dog` body. Before, most tools just used the first schema branch and silently ignored the rest.

```yaml
# openapi.yaml
components:
  schemas:
    Animal:
      oneOf:
        - $ref: '#/components/schemas/Cat'
        - $ref: '#/components/schemas/Dog'
      discriminator:
        propertyName: petType
        mapping:
          cat: '#/components/schemas/Cat'
          dog: '#/components/schemas/Dog'
```

api-mock-service generates **one scenario per variant**:

```
POST /animals — CreateAnimal-cat-200   (petType: "cat", has cat_name field)
POST /animals — CreateAnimal-dog-200   (petType: "dog", has dog_name field)
```

Both get tested. Neither gets silently skipped.

### Swagger UI

Every uploaded spec is immediately explorable via the built-in Swagger UI:

```
http://localhost:8080/swagger-ui/
```

```mermaid
flowchart LR
    Dev["Developer"]:::blue
    UI["Swagger UI\nlocalhost:8080/swagger-ui/"]:::green
    Mock["Mock Service\nlocalhost:8080"]:::blue
    Store[("Scenarios\n(YAML)")]:::storage

    Dev -->|"browse + try endpoints"| UI
    UI -->|"sends requests"| Mock
    Mock -->|"serves mock responses"| UI
    Mock --- Store

    classDef blue fill:#4A90D9,color:#fff,stroke:#2C5F8A
    classDef green fill:#27AE60,color:#fff,stroke:#1A7A42
    classDef storage fill:#8E44AD,color:#fff,stroke:#6C3483
```

No separate Swagger UI deployment. No CORS issues. The mock service and the interactive API playground are the same process. Upload your spec, open the browser, start experimenting — your frontend team, QA team, and new engineers can all explore the API contract interactively before any real service exists.

---

## Part 4: Contract Testing

This is where the tool earns its keep in production CI pipelines.

### Consumer-Driven Contracts

```mermaid
sequenceDiagram
    participant Consumer as Consumer Team
    participant Mock as api-mock-service
    participant Spec as YAML Scenarios

    Consumer->>Mock: Record interactions via proxy
    Mock-->>Spec: Save contract (request shape + response assertions)
    Note over Consumer,Spec: Commit scenarios to the consumer repo

    Consumer->>Mock: Run tests against mock (POST /_contracts/todos)
    Mock->>Mock: Validate request shape matches contract
    Mock-->>Consumer: Return mock response
    Consumer->>Consumer: Assert response shape matches expectations
```

The consumer team defines what requests they send and what responses they need. That contract lives in YAML files checked into their repo. The mock service enforces it — requests that don't match the `assert_contents_pattern` get rejected.

### Producer Contract Testing — The Hard Part Done Right

This is where most teams give up or use incomplete solutions. The producer executor:

1. Loads recorded scenarios (request shapes + expected response assertions)
2. Generates random but *valid* request data from the regex constraints
3. Sends real HTTP requests to the actual API
4. Validates the response shape, types, and values

```bash
# Run 10 iterations against the real API
curl -X POST http://localhost:8080/_contracts/todos \
  -H "Content-Type: application/json" \
  -d '{
    "base_url": "https://jsonplaceholder.typicode.com",
    "execution_times": 10,
    "verbose": true
  }'
```

Response:
```json
{
  "results": {
    "post-todo_0": {"id": 201},
    "post-todo_1": {"id": 201},
    "post-todo_2": {"id": 201}
  },
  "errors": {},
  "succeeded": 10,
  "failed": 0
}
```

Or from the CLI — which is what you'd put in CI:

```bash
api-mock-service producer-contract \
  --group todos \
  --base_url https://jsonplaceholder.typicode.com \
  --times 10
```

### Body Fields → Template Variables (Zero Config)

Here's one of my favorite features. If a POST body has `{"customerId": "cust-42", "amount": 100}`, the response template can reference those fields directly:

```yaml
response:
  contents: |-
    {
      "orderId": "{{RandRegex `ord-[0-9]{8}`}}",
      "customer": "{{.customerId}}",
      "total": {{.amount}},
      "confirmation": "Confirmed order for {{.customerId}}"
    }
```

No explicit variable configuration. The engine extracts top-level JSON body fields and injects them as template params. Path params and query params win on conflict, but body fields fill in everything else automatically.

### JSONPath Assertions — Testing Nested Responses

Flat key assertions work for simple responses. Real APIs return nested JSON. Use `$.` prefix for JSONPath:

```yaml
response:
  assert_contents_pattern: >
    {"$.user.email":"(__string__\\w+@\\w+\\.\\w+)",
     "$.order.items[0].price":"(__number__[0-9]+\\.?[0-9]*)",
     "$.meta.pagination.total":"(__number__\\d+)"}
```

Fully backward compatible — existing flat-key patterns don't change.

### Schema Validation Against OpenAPI (The Dredd Killer)

Dredd validates your API against its spec. api-mock-service does the same, but baked into your contract test run, with no separate process or config:

```bash
api-mock-service producer-contract \
  --group users \
  --base_url https://api.example.com \
  --spec openapi.yaml
```

Under the hood, every real API response runs through `openapi3filter`. If the response omits a required field, uses the wrong type, or violates an enum — you see it:

```
SCENARIO                                 STATUS
──────────────────────────────────────────────────────────────
GET /users/42-200                        ✗ FAIL
  Schema: email — value is required
  Schema: age — value must be >= 0 (got -1)
  Mismatch: status (expected "active", got "ACTIVE")
──────────────────────────────────────────────────────────────
TOTAL 10  Passed: 9  Failed: 1  Mismatched: 0
```

Field-level diagnostics tell you exactly which field failed and what the expected versus actual values were. No more `"assertion failed"` with no context.

### Coverage: Know What You're Not Testing

```bash
api-mock-service producer-contract \
  --group users \
  --base_url https://api.example.com \
  --spec openapi.yaml \
  --track-coverage
```

```
COVERAGE REPORT
──────────────────────────────────────────────────────────────
Overall: 75.0%  (6/8 paths)

Uncovered paths:
  ✗ DELETE /users/:id
  ✗ PATCH  /users/:id/status

Method coverage:
  GET    100.0%
  POST   100.0%
  DELETE 0.0%
  PATCH  0.0%
```

You can't improve what you can't measure. Coverage reporting shows which parts of your API contract are actually exercised in CI and which exist only on paper.

### Chaining: Testing Stateful Workflows

Most API testing tools treat each request in isolation. Real APIs have state: you create a resource, fetch it, update it, delete it. api-mock-service supports ordered scenario chains where outputs from one step become inputs to the next:

```yaml
# Step 0: Create a resource, capture its id
method: POST
name: create-order
path: /orders
order: 0
group: order-workflow
response:
  contents: '{"id": {{RandIntMinMax 1000 9999}}, "status": "created"}'
  add_shared_variables:
    - id
  assertions:
    - NumPropertyGE contents.id 1000

---
# Step 1: Fetch the created resource using {{.id}} from step 0
method: GET
name: get-order
path: /orders/:id
order: 1
group: order-workflow
response:
  contents: '{"id": {{.id}}, "status": "confirmed"}'
  assertions:
    - NumPropertyGE contents.id 1000
    - PropertyContains contents.status confirmed

---
# Step 2: Delete it
method: DELETE
name: delete-order
path: /orders/:id
order: 2
group: order-workflow
response:
  status_code: 204
```

```bash
curl -X POST http://localhost:8080/_contracts/order-workflow \
  -d '{"base_url": "https://api.example.com", "execution_times": 5}'
```

Five complete create→fetch→delete cycles. The `id` from the POST response flows automatically into the GET and DELETE requests.

---

## Part 5: Mutation Testing — Finding What You Didn't Think to Test

Property testing asks "does the API return the right shape?" Mutation testing asks "does the API *reject* the wrong shape?" These are different questions, and you need both.

A production API that returns `200 OK` when you pass `null` for a required field has a validation gap. An API that returns `200` when you inject `' OR 1=1; --` into a string field has a security gap. Mutation testing surfaces both.

```mermaid
flowchart TD
    S["Valid Scenario\n(recorded from real traffic)"]:::blue
    M["Mutation Generator"]:::amber

    subgraph Mutations["Generated Variants"]
        N["Null mutations\n(one per field)"]:::red
        B["Boundary mutations\nMinInt32 + MaxInt32\nempty + 255-char string"]:::amber
        F["Format violations\nbad date, bad UUID, bad email"]:::amber
        SEC["Security injection\nSQLi · Path traversal · LDAP\nXXE · SSRF · Cmd injection"]:::red
        COM["Combinatorial\npairs of above, max 10"]:::amber
    end

    E["Execute each variant\nagainst real API"]:::blue
    R["Assert: API returns 4xx\nfor every mutation"]:::green

    S --> M --> N & B & F & SEC & COM --> E --> R

    classDef blue fill:#4A90D9,color:#fff,stroke:#2C5F8A
    classDef amber fill:#F39C12,color:#fff,stroke:#B7770D
    classDef red fill:#E74C3C,color:#fff,stroke:#A93226
    classDef green fill:#27AE60,color:#fff,stroke:#1A7A42
```

### Run Mutations

```bash
# CLI
api-mock-service producer-contract \
  --group users \
  --base_url https://api.example.com \
  --mutations

# HTTP
curl -X POST http://localhost:8080/_contracts/mutations/users \
  -d '{"base_url": "https://api.example.com"}'
```

For a scenario with a body `{"name": "Alice", "email": "alice@example.com", "age": 25}`, this generates and executes:

| Mutation | Payload | Expected |
|---------|---------|---------|
| null-name | `{"name": null, "email": "alice@...", "age": 25}` | 422 |
| null-email | `{"name": "Alice", "email": null, "age": 25}` | 422 |
| null-age | `{"name": "Alice", "email": "alice@...", "age": null}` | 422 |
| boundary-min | `{"name": "", "email": "", "age": -2147483648}` | 400/422 |
| boundary-max | `{"name": "AAA...255chars", "email": "BBB...", "age": 2147483647}` | 400/422 |
| format-email | `{"name": "Alice", "email": "@nodomain", "age": 25}` | 400/422 |
| sqli-name | `{"name": "' OR 1=1; --", "email": "alice@...", "age": 25}` | 400/422 |
| xxe-name | `{"name": "<?xml ...>", "email": "alice@...", "age": 25}` | 400/422 |
| ssrf-email | `{"name": "Alice", "email": "http://169.254.169.254/...", "age": 25}` | 400/422 |

If your API returns `200` for any of these, mutation testing fails — and you've found a real bug before a customer or attacker does.

---

## Part 6: A Real-World Workflow

Here's how a team actually uses this in their development cycle:

```mermaid
flowchart LR
    subgraph Dev["Development"]
        D1["Developer adds\nnew endpoint"]:::blue
        D2["Record against\nstaging / sandbox"]:::amber
    end

    subgraph Review["Code Review"]
        R1["YAML scenarios\ncommitted to repo"]:::green
        R2["Reviewers can see\nexact contract"]:::green
    end

    subgraph CI["CI Pipeline"]
        C1["Unit tests"]:::blue
        C2["Contract tests\napi-mock-service\n--group my-api\n--base_url $STAGING_URL\n--spec openapi.yaml\n--track-coverage"]:::green
        C3["Mutation tests\n--mutations"]:::red
        C4["Coverage gate\n>= 80% required"]:::amber
    end

    subgraph Deploy["Deploy"]
        P1["Merge to main"]:::blue
        P2["Release to prod"]:::green
    end

    D1 --> D2 --> R1 --> R2 --> C1 --> C2 --> C3 --> C4 --> P1 --> P2
    C2 -->|"fail: schema violation"| D1
    C3 -->|"fail: accepts null input"| D1
    C4 -->|"fail: DELETE uncovered"| D1

    classDef blue fill:#4A90D9,color:#fff,stroke:#2C5F8A
    classDef green fill:#27AE60,color:#fff,stroke:#1A7A42
    classDef amber fill:#F39C12,color:#fff,stroke:#B7770D
    classDef red fill:#E74C3C,color:#fff,stroke:#A93226
```

**Day 1**: Developer adds a `PATCH /users/:id/status` endpoint. They route traffic through the proxy, hit their staging service a few times, and commit the YAML scenarios alongside their code.

**Code review**: Reviewers can see the exact contract — what requests are expected, what response shapes are required. No ambiguity. The PR diff shows contract changes as clearly as code changes.

**CI**: Three stages that would have caught the last three production bugs:
1. Contract tests with schema validation catch missing required fields and type mismatches
2. Mutation tests verify the endpoint rejects null inputs, bad formats, and security payloads
3. Coverage gate ensures the new endpoint was actually exercised (not just the happy path)

**Deploy**: If all three pass, merge with confidence.

---

## Part 7: The Numbers That Matter to a CTO

I know what the next question is: *"This sounds good, but we already have Postman/Pactflow/WireMock. Why replace them?"*

You don't have to replace them. But here's what teams have found after adopting api-mock-service:

**Setup time**: Postman requires manual test writing. Pactflow requires a broker deployment, a consumer library, and a provider verifier — typically 2–3 days of DevOps work. api-mock-service: `docker run` and `export http_proxy`. 10 minutes.

**Maintenance burden**: Pact consumer tests are code — they break when the API changes and need developer time to update. api-mock-service scenarios are YAML files auto-generated from recordings. When the API changes, re-record. 5 minutes.

**Coverage you didn't know you had gaps in**: Most teams testing with Postman or Pact test the happy path and a few error cases they thought of. Mutation testing generates dozens of adversarial variants automatically. In our experience, the first time teams run mutation tests, they find at least one case where the API accepts a null required field or reflects an injection payload.

**The spec drift problem**: Your OpenAPI spec says `email` is required. Your API doesn't validate it. Postman tests pass (they use a valid email). Schema validation catches it.

**Infrastructure cost**: Pactflow is $0–$500+/month depending on tier. WireMock Cloud is similar. api-mock-service: $0, runs on the same machine as your CI agent.

```mermaid
graph LR
    subgraph "Cost + Complexity"
        direction TB
        P["Pactflow\n$$$, broker required\nlanguage libraries\nDevOps to maintain"]:::red
        W["WireMock\nstubbing only\nno contract testing\nno fuzz"]:::amber
        A["api-mock-service\n$0, single binary\nYAML scenarios\nCI in 10 minutes"]:::green
    end

    classDef red fill:#E74C3C,color:#fff,stroke:#A93226
    classDef amber fill:#F39C12,color:#fff,stroke:#B7770D
    classDef green fill:#27AE60,color:#fff,stroke:#1A7A42
```

The pitch to a CTO: **one tool, one binary, zero infrastructure, covers the full testing pyramid from unit-test stubs to contract verification to security input validation.** The scenarios are YAML files that live in your repo — no external system to maintain, no vendor lock-in, no monthly invoice.

---

## Complete Quick-Reference

### Start the server
```bash
docker run -p 8080:8080 -p 8081:8081 -e DATA_DIR=/tmp/mocks \
  plexobject/api-mock-service:latest
```

### Record via proxy
```bash
export http_proxy="http://localhost:8081"
curl -k https://your-api.example.com/endpoint
```

### Upload a scenario
```bash
curl -H "Content-Type: application/yaml" \
  --data-binary @scenario.yaml http://localhost:8080/_scenarios
```

### Upload an OpenAPI spec
```bash
curl -H "Content-Type: application/yaml" \
  --data-binary @openapi.yaml http://localhost:8080/_oapi
```

### Browse the Swagger UI
```
http://localhost:8080/swagger-ui/
```

### Run contract tests
```bash
# HTTP
curl -X POST http://localhost:8080/_contracts/my-group \
  -d '{"base_url": "https://api.example.com", "execution_times": 10}'

# CLI with schema validation + coverage
api-mock-service producer-contract \
  --group my-group \
  --base_url https://api.example.com \
  --spec openapi.yaml \
  --track-coverage \
  --times 10
```

### Run mutation tests
```bash
curl -X POST http://localhost:8080/_contracts/mutations/my-group \
  -d '{"base_url": "https://api.example.com"}'

# CLI
api-mock-service producer-contract \
  --group my-group \
  --base_url https://api.example.com \
  --mutations
```

### Configure chaos
```bash
curl -X PUT http://localhost:8080/_groups/my-group/config \
  -d '{"chaos_enabled": true, "mean_time_between_failure": 5, "http_errors": [500, 503]}'
```

---

## What's Next

The patterns in this guide — record, mock, contract, mutate — cover the vast majority of API reliability problems teams hit in practice. But the framework is designed to grow:

- **Stateful workflows** (CREATE → READ → DELETE sequences with session-scoped state)
- **Spec diff / breaking change detection** (compare v1 vs v2 OpenAPI, fail CI on breaking changes)
- **Fuzz shrinking** (when a mutation fails, automatically bisect to the minimal reproducing payload)

The source is open. The issues are tracked. Contributions welcome.

---

*The repository: [github.com/bhatti/api-mock-service](https://github.com/bhatti/api-mock-service)*

*Previous posts in this series:*
- *[Property-based and Generative testing for Microservices](https://shahbhat.medium.com/property-based-and-generative-testing-for-microservices-1c6df1abb40b)*
- *[Mocking Distributed Microservices](https://shahbhat.medium.com/mocking-distributed-micro-services-47c0d658d4bb)*
- *[Contract Testing for REST APIs](https://shahbhat.medium.com/contract-testing-for-rest-apis-31680ed6bbf3)*
