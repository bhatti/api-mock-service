# Examples: All New Capabilities

This document provides runnable examples for every capability added by the security testing, injection detection, and CI/CD reporting features. Start the mock service first:

```bash
# Start the service
./out/bin/api-mock-service

# Or via Docker
docker run -p 8080:8080 -p 8081:8081 \
  -e DATA_DIR=/tmp/mocks plexobject/api-mock-service:latest
```

---

## 1. Grammar-Based Security Payload Generators

Eight injection payload generators are available in scenario templates. Each has a random (`Rand*`) and deterministic (`SeededRand*`) variant.

### Example: Custom Security Test Scenario

```yaml
# File: mock_tests/api_contracts/security-demo/POST/create-user.yaml
method: POST
name: security-sqli-test
path: /api/users
group: security-demo
request:
    headers:
        Content-Type: application/json
    contents: '{"name":"{{RandSQLi}}","email":"{{RandXSS}}","role":"user"}'
response:
    status_code: 400
    contents: '{"error":"invalid input"}'
```

### All Available Generators

| Template Function | Attack Class | Example Payloads |
|---|---|---|
| `{{RandSQLi}}` / `{{SeededRandSQLi}}` | SQL Injection | `' OR 1=1; --`, `1 UNION SELECT username FROM users`, `'; WAITFOR DELAY '0:0:5'; --` |
| `{{RandXSS}}` / `{{SeededRandXSS}}` | Cross-Site Scripting | `<script>alert(1)</script>`, `<img onerror=alert(1) src=x>`, `javascript:alert(1)` |
| `{{RandPathTraversal}}` / `{{SeededRandPathTraversal}}` | Path Traversal | `../../../etc/passwd`, `....//....//etc/shadow`, `%2e%2e%2f` |
| `{{RandSSTI}}` / `{{SeededRandSSTI}}` | Server-Side Template Injection | `{{7*7}}`, `${7*7}`, `<%= 7*7 %>` |
| `{{RandCmdInjection}}` / `{{SeededRandCmdInjection}}` | Command Injection | `; cat /etc/passwd`, `` `whoami` ``, `$(id)` |
| `{{RandNoSQLi}}` / `{{SeededRandNoSQLi}}` | NoSQL Injection | `{"$gt":""}`, `{"$ne":null}`, `'; db.users.drop(); '` |
| `{{RandLDAPi}}` / `{{SeededRandLDAPi}}` | LDAP Injection | `*)(uid=*))(|(uid=*`, `admin)(|(password=*))` |
| `{{RandXXE}}` / `{{SeededRandXXE}}` | XML External Entity | `<!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]>` |

### Using Seeded Variants for Reproducibility

```yaml
contents: '{"name":"{{SeededRandSQLi 42}}","email":"{{SeededRandXSS 42}}"}'
# Same seed (42) produces the same payload every time
```

---

## 2. Mutation Testing with Security Injection

Run mutation testing against a real API. Mutations automatically include null fields, boundary values, format violations, AND security injection payloads from all 8 generators.

### HTTP API

```bash
# Run mutations against the "security-demo" group
curl -X POST http://localhost:8080/_contracts/mutations/security-demo \
  -H "Content-Type: application/json" \
  -d '{
    "base_url": "https://api.example.com",
    "execution_times": 3,
    "run_mutations": true,
    "verbose": true
  }'
```

### CLI

```bash
api-mock-service producer-contract \
  --group security-demo \
  --base_url https://api.example.com \
  --mutations
```

### Response Structure

```json
{
  "results": { "POST /api/users create-user-201": "" },
  "errors": {
    "POST /api/users create-user-sqli-name-201": "actual status 500 != expected 201"
  },
  "succeeded": 8,
  "failed": 3,
  "securitySummary": {
    "totalFindings": 2,
    "bySeverity": { "high": 1, "medium": 1 },
    "byCategory": { "CWE-89": 1, "CWE-79": 1 },
    "passedChecks": ["ssti", "cmd-injection", "nosqli", "ldapi", "xxe", "path-traversal"]
  }
}
```

---

## 3. Multiple Mutation Rounds

Run multiple independent rounds of mutation payloads for better coverage. Each round generates fresh randomized payloads.

```bash
curl -X POST http://localhost:8080/_contracts/mutations/security-demo \
  -H "Content-Type: application/json" \
  -d '{
    "base_url": "https://api.example.com",
    "execution_times": 3,
    "run_mutations": true,
    "mutation_rounds": 5
  }'
```

With `mutation_rounds: 5`, each scenario gets 5 independent rounds of mutation payloads, dramatically increasing the chance of finding edge-case vulnerabilities.

---

## 4. Injection Detection (Response Analysis)

Every mutation response is automatically scanned for evidence that the injection succeeded.

### Detection Categories

| Category | What It Detects | CWE |
|---|---|---|
| SQL Injection | MySQL/PostgreSQL/SQLite/MSSQL/Oracle error signatures | CWE-89 |
| NoSQL Injection | MongoDB/CouchDB/Cassandra error patterns | CWE-943 |
| LDAP Injection | javax.naming/LDAPException signatures | CWE-90 |
| Reflected XSS | Payload appears unescaped in response | CWE-79 |
| Reflected SSTI | Template syntax reflected back | CWE-1336 |
| Command Injection | Command output in response | CWE-78 |
| Info Disclosure | Stack traces, file paths, server version strings | CWE-200 |
| Unhandled Input | Unexpected 500 after injection payload | CWE-20 |

### Example Finding

```json
{
  "endpoint": "POST /api/users",
  "field": "name",
  "payload": "' OR 1=1; --",
  "category": "sql-error-signature",
  "severity": "high",
  "evidence": "You have an error in your SQL syntax near '1=1'",
  "cwe": "CWE-89",
  "owasp": "WSTG-INPV-05"
}
```

---

## 5. Blind Injection Detection (Timing-Based)

For attacks that produce no visible error (blind SQL injection, blind SSRF), the detector compares response times against a baseline.

### How It Works

1. **Phase 1**: A clean run establishes baseline response time per scenario
2. **Phase 2**: Mutation payloads are sent; if a payload containing timing keywords (`SLEEP`, `WAITFOR`, `pg_sleep`, `BENCHMARK`) causes a response time > baseline * threshold, it's flagged

### Configure Sensitivity

```bash
curl -X POST http://localhost:8080/_contracts/mutations/security-demo \
  -H "Content-Type: application/json" \
  -d '{
    "base_url": "https://api.example.com",
    "execution_times": 3,
    "run_mutations": true,
    "timing_threshold_multiplier": 5.0
  }'
```

Default threshold is 3.0x. A value of 5.0 means a response must be 5x slower than the baseline to be flagged.

### Example Finding

```json
{
  "category": "blind-timing",
  "payload": "'; WAITFOR DELAY '0:0:5'; --",
  "evidence": "response 5200ms exceeded baseline 50ms * threshold 3.0 (150ms)",
  "cwe": "CWE-89"
}
```

---

## 6. Sequence-Level Mutations

Beyond per-field mutations, sequence mutations test operation ordering and idempotency.

### Four Strategies

| Strategy | What It Tests | Example |
|---|---|---|
| **Reversed** | Order dependency | DELETE -> GET -> POST instead of POST -> GET -> DELETE |
| **Skip Step** | Missing prerequisites | POST -> DELETE (skipping GET) |
| **Duplicate Step** | Idempotency | POST -> POST -> GET -> DELETE |
| **Method Swap** | HTTP method confusion | PUT <-> PATCH, GET <-> POST |

### Enabling Sequence Mutations

Sequence mutations run automatically during `/_contracts/mutations/:group` when the group has 2+ scenarios. No additional configuration needed.

```bash
# Groups with 2+ scenarios automatically get sequence mutations
curl -X POST http://localhost:8080/_contracts/mutations/security-demo \
  -H "Content-Type: application/json" \
  -d '{
    "base_url": "https://api.example.com",
    "execution_times": 3,
    "run_mutations": true
  }'
```

---

## 7. Auto-Discovered Request Chains (Dependency Graph)

When you upload an OpenAPI spec, the service automatically discovers operation dependencies by matching response fields to request parameters.

### How It Works

```bash
# Upload an OpenAPI spec
curl -X POST http://localhost:8080/_oapi \
  -H "Content-Type: application/yaml" \
  --data-binary @petstore.yaml
```

The dependency graph uses confidence-weighted name matching:

| Match Type | Confidence | Example |
|---|---|---|
| Exact name | 1.0 | Response `id` -> Request path param `id` |
| Singular/plural | 0.9 | Response `petId` -> Request `pet_id` |
| ID + path | 0.8 | Response `id` in `/pets` -> Request `petId` |
| Semantic | 0.7 | Response `created_id` -> Request `resourceId` |

### Result

Operations get a `DependencyOrder` field. Contract testing and mutation testing execute them in dependency order (POST first, then GET/PUT, then DELETE), ensuring test flows match real usage patterns.

---

## 8. Failure Deduplication

Mutation testing can produce hundreds of failures from the same root cause. Failure deduplication clusters them by `(statusCode, errorCategory, endpoint, responseHash)`.

### Example

Running 50 mutation rounds might produce 200 failures, but deduplication reveals just 3 root causes:

```json
{
  "failureClusters": [
    {
      "key": {
        "statusCode": 500,
        "errorCategory": "injection",
        "endpoint": "POST /users"
      },
      "count": 150,
      "severity": "critical",
      "sampleError": "actual status 500 != expected 201",
      "affectedMutations": ["sqli-name", "sqli-email", "xss-name"]
    },
    {
      "key": {
        "statusCode": 422,
        "errorCategory": "schema-violation",
        "endpoint": "POST /users"
      },
      "count": 40,
      "severity": "medium",
      "sampleError": "actual status 422 != expected 201"
    }
  ]
}
```

---

## 9. JUnit XML Export for CI/CD

Export mutation test results as JUnit XML for integration with Jenkins, GitHub Actions, GitLab CI, etc.

### Step 1: Run Mutations

```bash
curl -X POST http://localhost:8080/_contracts/mutations/my-api \
  -H "Content-Type: application/json" \
  -d '{"base_url": "https://api.example.com", "execution_times": 3, "run_mutations": true}'
```

### Step 2: Fetch JUnit XML

```bash
curl http://localhost:8080/_reports/my-api/junit -o test-results.xml
```

### Example JUnit XML

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="my-api" tests="10" failures="2" errors="0">
  <testsuite name="my-api" tests="10" failures="2">
    <properties>
      <property name="succeeded" value="8"/>
      <property name="failed" value="2"/>
      <property name="injection_findings" value="3"/>
      <property name="failure_clusters" value="1"/>
    </properties>
    <testcase name="POST /users create-user-201" classname="my-api"/>
    <testcase name="POST /users create-user-sqli-name-201" classname="my-api">
      <failure message="actual status 500 != expected 201"/>
    </testcase>
  </testsuite>
</testsuites>
```

### GitHub Actions Integration

```yaml
- name: Run API contract tests
  run: |
    curl -X POST http://localhost:8080/_contracts/mutations/my-api \
      -H "Content-Type: application/json" \
      -d '{"base_url": "${{ env.API_URL }}", "run_mutations": true}'
    curl http://localhost:8080/_reports/my-api/junit -o test-results.xml

- name: Publish test results
  uses: dorny/test-reporter@v1
  with:
    name: API Security Tests
    path: test-results.xml
    reporter: java-junit
```

---

## 10. JSON Security Summary Export

Get a structured JSON summary of mutation test results with security findings, failure clusters, and top errors.

```bash
curl http://localhost:8080/_reports/my-api/summary | jq .
```

### Example Response

```json
{
  "group": "my-api",
  "succeeded": 8,
  "failed": 2,
  "total": 10,
  "securitySummary": {
    "totalFindings": 3,
    "bySeverity": { "high": 2, "medium": 1 },
    "byCategory": { "CWE-89": 2, "CWE-79": 1 },
    "passedChecks": ["ssti", "cmd-injection", "nosqli", "ldapi", "xxe", "path-traversal"]
  },
  "failureClusters": [
    {
      "key": { "statusCode": 500, "errorCategory": "injection", "endpoint": "POST /users" },
      "count": 2,
      "severity": "critical"
    }
  ],
  "topErrors": [
    "actual status 500 != expected 201 (POST /users create-user-sqli-name-201)"
  ]
}
```

---

## 11. OWASP/CWE Classification

All injection findings are automatically classified with CWE identifiers and OWASP WSTG test IDs:

| CWE | OWASP Test ID | Category |
|---|---|---|
| CWE-89 | WSTG-INPV-05 | SQL Injection |
| CWE-79 | WSTG-INPV-01 | Cross-Site Scripting |
| CWE-78 | WSTG-INPV-12 | Command Injection |
| CWE-22 | WSTG-ATHZ-01 | Path Traversal |
| CWE-90 | WSTG-INPV-06 | LDAP Injection |
| CWE-611 | WSTG-INPV-07 | XML External Entity |
| CWE-1336 | WSTG-INPV-18 | Server-Side Template Injection |
| CWE-943 | WSTG-INPV-05 | NoSQL Injection |
| CWE-200 | WSTG-INFO-02 | Information Disclosure |
| CWE-20 | WSTG-INPV-17 | Improper Input Validation |

The `passedChecks` field in the security summary lists vulnerability classes where no findings were detected, confirming your API handles those attack vectors safely.

---

## End-to-End Walkthrough

Here is a complete walkthrough combining all capabilities:

```bash
# 1. Start the mock service
./out/bin/api-mock-service &

# 2. Upload scenarios (or use proxy recording)
curl -H "Content-Type: application/yaml" \
  --data-binary @mock_tests/api_contracts/security-demo/POST/create-user.yaml \
  http://localhost:8080/_scenarios

# 3. Upload an OpenAPI spec (enables dependency discovery + schema validation)
curl -X POST http://localhost:8080/_oapi \
  -H "Content-Type: application/yaml" \
  --data-binary @openapi.yaml

# 4. Run mutation testing with security injection and timing analysis
curl -X POST http://localhost:8080/_contracts/mutations/security-demo \
  -H "Content-Type: application/json" \
  -d '{
    "base_url": "https://api.example.com",
    "execution_times": 3,
    "run_mutations": true,
    "track_coverage": true,
    "spec_content": "'"$(cat openapi.yaml)"'",
    "timing_threshold_multiplier": 3.0,
    "mutation_rounds": 3
  }'

# 5. Export JUnit XML for CI
curl http://localhost:8080/_reports/security-demo/junit -o results.xml

# 6. Get JSON security summary
curl http://localhost:8080/_reports/security-demo/summary | jq .

# 7. Check coverage
curl http://localhost:8080/_coverage/security-demo | jq .
```

---

## Related Documentation

- [Contract Testing Guide](contract-testing.md) — Full contract testing reference
- [Fuzz & Property Testing](fuzz-property-testing.md) — Mutation strategies, shrinking, injection detection
- [How-To Guide](how-to-guide.md) — 33 cookbook recipes
- [API Reference](api-reference.md) — All HTTP endpoints
- [OpenAPI Guide](openapi-guide.md) — Spec upload, dependency discovery
