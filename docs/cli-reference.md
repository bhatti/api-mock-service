# CLI Reference

api-mock-service ships a single binary with multiple subcommands.

## Installation

```bash
# Build from source
make && ./out/bin/api-mock-service

# Docker
docker run -p 8080:8080 -p 8081:8081 \
  -e HTTP_PORT=8080 -e PROXY_PORT=8081 -e DATA_DIR=/tmp/mocks \
  plexobject/api-mock-service:latest

# go install
go install github.com/bhatti/api-mock-service@latest
```

---

## `api-mock-service` — Start the Server

Starts the mock service on two ports: HTTP playback/API on `httpPort` and proxy recorder on `proxyPort`.

```
api-mock-service [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--httpPort` | int | `8080` | HTTP port for mock playback and API endpoints |
| `--proxyPort` | int | `8081` | Proxy recorder port |
| `--dataDir` | string | `default_mocks_data` | Directory to store scenario YAML files |
| `--config` | string | — | Path to config file |

### Examples

```bash
# Start with custom ports and data dir
api-mock-service --httpPort 9090 --proxyPort 9091 --dataDir /var/mocks

# Start with config file
api-mock-service --config /etc/api-mock/config.yaml
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `HTTP_PORT` | Same as `--httpPort` |
| `PROXY_PORT` | Same as `--proxyPort` |
| `DATA_DIR` | Same as `--dataDir` |
| `ASSET_DIR` | Directory for static assets served at `/_assets` |
| `HISTORY_DIR` | Directory for execution history |

---

## `api-mock-service producer-contract` — Run Contract Tests

Runs producer-driven contract tests against a real API. Loads scenarios from the local data directory, generates fuzz request data, sends real HTTP requests, and validates responses.

```
api-mock-service producer-contract [flags]
```

### Flags

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--base_url` | string | — | yes | Base URL of the real API to test |
| `--group` | string | — | yes* | Group of scenarios to run (*required unless `--scenario` is set) |
| `--scenario` | string | — | no | Path to a specific scenario YAML file to run |
| `--times` | int | `10` | no | Number of execution iterations per scenario |
| `--verbose` | bool | `false` | no | Log request/response bodies |
| `--dataDir` | string | — | no | Data directory (overrides default) |
| `--spec` | string | — | no | **[Plan A]** Path to OpenAPI spec file for response schema validation |
| `--track-coverage` | bool | `false` | no | **[Plan A]** Include OpenAPI coverage report in output (requires `--spec`) |
| `--mutations` | bool | `false` | no | **[Plan A]** Run mutation testing instead of normal contract execution (requires `--group`) |

### Examples

#### Run contracts for a group

```bash
api-mock-service producer-contract \
  --group todos \
  --base_url https://jsonplaceholder.typicode.com \
  --times 10
```

#### Run a specific scenario file

```bash
api-mock-service producer-contract \
  --scenario fixtures/get_todo.yaml \
  --base_url https://jsonplaceholder.typicode.com
```

#### Validate responses against OpenAPI schema

```bash
api-mock-service producer-contract \
  --group my-api \
  --base_url https://api.example.com \
  --spec openapi.yaml \
  --times 5
```

#### Track OpenAPI path coverage

```bash
api-mock-service producer-contract \
  --group my-api \
  --base_url https://api.example.com \
  --spec openapi.yaml \
  --track-coverage \
  --times 10
```

Output includes:
```
COVERAGE REPORT
──────────────────────────────────────────────────────────────
Overall: 87.5%  (7/8 paths)

Uncovered paths:
  ✗ DELETE /users/:id

Method coverage:
  GET    100.0%
  POST   75.0%
```

#### Run mutation testing

```bash
api-mock-service producer-contract \
  --group my-api \
  --base_url https://api.example.com \
  --mutations
```

Mutation testing requires `--group`. It generates corrupted request variants (null fields, boundary values, format violations, security payloads) and verifies the API rejects them.

```
──────────────────────────────────────────────────────────────
SCENARIO                                 STATUS
──────────────────────────────────────────────────────────────
create-user-null-email_0                 ✓ PASS
create-user-boundary-min_0               ✓ PASS
create-user-sqli-name_0                  ✗ FAIL
  Schema: Response contained injection payload
──────────────────────────────────────────────────────────────
TOTAL 42  Passed: 41  Failed: 1  Mismatched: 0
```

#### Combine mutations + schema validation

```bash
api-mock-service producer-contract \
  --group my-api \
  --base_url https://api.example.com \
  --spec openapi.yaml \
  --mutations \
  --track-coverage
```

### Output Format

The CLI prints a colored table (ANSI codes, TTY-detected — plain text in CI):

```
──────────────────────────────────────────────────────────────
SCENARIO                                 STATUS     LATENCY
──────────────────────────────────────────────────────────────
GET /todos-200                           ✓ PASS
POST /todos-201                          ✗ FAIL
  Missing: id, userId
  Mismatch: completed (expected false, got "false")
  Schema: status 422 expected, got 200
──────────────────────────────────────────────────────────────
TOTAL 10  Passed: 9  Failed: 1  Mismatched: 0
```

Colors: green = PASS, red = FAIL, yellow = warning. Disabled automatically in non-TTY environments (CI, pipes).

---

## `api-mock-service contract` — Consumer Contract Client

Runs consumer contract tests (legacy alias, prefer `producer-contract`).

```bash
api-mock-service contract \
  --base_url https://jsonplaceholder.typicode.com \
  --group todos \
  --times 10
```

---

## `api-mock-service config` — Show Configuration

Prints the active configuration.

```bash
api-mock-service config
```

---

## `api-mock-service version` — Show Version

```bash
api-mock-service version
```

---

## Global Flags

Available on all subcommands:

| Flag | Description |
|------|-------------|
| `--config` | Path to config file |
| `--dataDir` | Data directory for scenarios |
| `-h, --help` | Help for the command |

---

## Related Docs

- [API Reference](api-reference.md) — HTTP endpoints for the same operations
- [Contract Testing](contract-testing.md) — contract patterns and request body fields
- [Fuzz & Property Testing](fuzz-property-testing.md) — mutation strategies
- [OpenAPI Guide](openapi-guide.md) — using `--spec` and coverage
