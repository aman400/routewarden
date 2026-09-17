# Testing & Verification

RouteWarden contains a comprehensive, multi-layer testing architecture designed to guarantee correctness, concurrency safety, and total compliance with Traefik's Yaegi interpreter constraints.

---

## Running Automated Tests

### 1. Run All Tests with Race Detector
RouteWarden handles concurrent web requests under load. Always execute tests with the Go race detector enabled:

```bash
go test -v -race ./...
```

### 2. Check Statement Code Coverage
Measure statement coverage across the codebase:

```bash
go test -cover ./...
```

To generate and view an interactive HTML coverage heatmap in your browser:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

*(Current statement coverage stands at **>92%** across all modules).*

### 3. Run Specific Test Suites
You can target individual component suites directly:

```bash
# Run only security evasion tests
go test -v -run TestRouteWarden_SecurityEvasionVectors ./...

# Run only IP / CIDR whitelist evaluation tests
go test -v -run TestIPFilter ./...

# Run only multi-middleware integration pipeline tests
go test -v -run TestPipeline ./...
```

---

## Test Architecture & Modular Coverage

RouteWarden isolates unit testing per component and uses integration tests to validate full HTTP request cycles:

| Test File | Scope & Focus |
|---|---|
| [`config_test.go`](https://github.com/routewarden/traefik-warden/blob/main/config_test.go) | Default settings, regex dictionaries, validation of `CreateConfig()` and factory defaults. |
| [`path_normalizer_test.go`](https://github.com/routewarden/traefik-warden/blob/main/path_normalizer_test.go) | Recursive URL unescaping (`%252e%252e`), semicolon matrix parameters (`/;param/.env`), Windows backslash normalization, and null byte injection attacks. |
| [`ip_filter_test.go`](https://github.com/routewarden/traefik-warden/blob/main/ip_filter_test.go) | Exact IPv4/IPv6 addresses, CIDR subnet matching, parsing `X-Forwarded-For`, `X-Real-IP`, and invalid IP syntax error handling. |
| [`response_handler_test.go`](https://github.com/routewarden/traefik-warden/blob/main/response_handler_test.go) | Custom JSON formatting, HTML templates, Turnstile/hCaptcha/reCAPTCHA markup generation, redirects, silent drops, and status codes. |
| [`routewarden_test.go`](https://github.com/routewarden/traefik-warden/blob/main/routewarden_test.go) | Middleware integration, default patterns, custom regex rules, allowlist overrides, query inspection, and disabled mode pass-through. |
| [`integration_test.go`](https://github.com/routewarden/traefik-warden/blob/main/integration_test.go) | Full multi-middleware pipeline simulation (`Tracing ➡️ RouteWarden ➡️ Upstream Backend Service`). |

---

## Continuous Integration (GitHub Actions)

Every pull request and push to `main` runs RouteWarden's automated CI pipeline:
- Validates code formatting with `gofmt`
- Compiles with Go 1.22+ and 1.23+
- Executes `go test -v -race ./...`
