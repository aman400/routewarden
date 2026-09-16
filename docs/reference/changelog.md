# Changelog & Migration Guide

All notable changes to the **RouteWarden** Traefik middleware plugin are documented below, along with breaking changes and migration advice between versions.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and RouteWarden adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.2.x Series] — Latest

The `v0.2.x` release series introduces CIDR/IP whitelisting, comprehensive anti-evasion hardening, a multi-mode response engine, and an interactive documentation site.

### Breaking Changes & Upgrade Considerations

::: danger Breaking Changes in v0.2.x
1. **Config Key Renaming (`blockPatterns` ➔ `pathPatterns`)**:
   - In `v0.1.0`, `blockPatterns` was used in some examples. In `v0.2.x`, `pathPatterns` is the primary configuration key. Although `blockPatterns` is retained as a backward-compatible alias in Go, configuring `pathPatterns` is recommended.
2. **Normalized Path Matching**:
   - Starting in `v0.2.0`, incoming paths are strictly canonicalized and anti-evasion decoded before regex evaluation. If your custom regex in `v0.1.x` relied on matched raw URL-encoded characters (such as `%2e` or `%2f`), it will no longer match because paths are decoded prior to inspection. Regexes should match raw plain path segments.
3. **Response Header Structure**:
   - Custom response headers in `response.headers` are now strictly validated against standard HTTP header formatting.
:::

### Version Differences (v0.1.x vs v0.2.x)

| Feature / Capability | v0.1.x | v0.2.x | Notes / Details |
|---|---|---|---|
| **IP / CIDR Whitelisting** | ❌ Not available | ✅ **`allowedIps`** | Whitelist IPs or subnets (e.g. `10.0.0.0/8`, `192.168.1.100`) to bypass blocking. |
| **Client IP Resolution** | ❌ None | ✅ **`X-Forwarded-For` & `X-Real-IP`** | Accurately tracks origin IP through reverse proxies and load balancers. |
| **Response Modes** | `json`, `html`, `text`, `redirect` | `json`, `html`, `text`, `redirect`, **`captcha`**, **`silentDrop`** | Fully integrated Cloudflare Turnstile & hCaptcha challenge templates. |
| **Path Anti-Evasion** | Basic URL decode | Multi-layer decode, dot-segment traversal, IIS backslash & matrix param scrubbing | Neutralizes `%252e%252e`, `/;param/.env`, and `\\` evasion vectors. |
| **Test Suite Coverage** | ~60% basic tests | **92.6% statement coverage** | Per-component isolation tests with Yaegi conformance and race detection. |
| **Documentation** | Readme only | Interactive VitePress Wiki + Version Switching | Live searchable documentation with unified code tabs and live examples. |

---

### [v0.2.1] - 2026-09-16

#### Added
- **Interactive Documentation & Wiki Site (VitePress)**:
  - Official documentation site hosted on GitHub Pages ([`https://aman400.github.io/routewarden/`](https://aman400.github.io/routewarden/)).
  - Client-side full-text search, dark/light theme, and synchronized multi-format code previews (YAML, TOML, CLI).
  - Version switching across documentation branches (`v0.2.x` and `v0.1.x`).
- **In-Repo Examples Suite (`examples/`)**:
  - `01-basic-sensitive-files`: Quickstart shielding backend services against `.env`, `.git`, backups, and configs.
  - `02-global-entrypoint-shield`: Global entrypoint middleware shielding all services across Traefik without per-service labels.
  - `03-ip-whitelist-vpn`: Bypassing security checks for trusted CIDR / VPN networks.
  - `04-captcha-challenge`: Verification challenges with Cloudflare Turnstile and hCaptcha.
  - `05-kubernetes-ingressroute`: Kubernetes Traefik `Middleware` and `IngressRoute` CRD manifests.
- **Automated GitHub Pages CI/CD Pipeline**:
  - Added `.github/workflows/deploy-docs.yml` using GitHub Actions and `@actions/deploy-pages`.

#### Changed
- **Streamlined `README.md`**:
  - Simplified landing page with quickstart returning 404 Not Found error payloads.
  - Concise configuration summary table and badges linking to the documentation wiki.

---

### [v0.2.0] - 2026-09-16

#### Added
- **IP & CIDR Subnet Whitelisting (`allowedIps`)**:
  - Added `allowedIps` configuration supporting IPv4 addresses, IPv6 addresses, and CIDR subnet masks (e.g., `127.0.0.1`, `10.0.0.0/8`, `2001:db8::/32`).
  - Implemented client IP resolution with proxy forwarding support (`X-Forwarded-For`, `X-Real-IP`, and socket `RemoteAddr`).
  - Requests originating from whitelisted IPs/subnets bypass sensitive route blocking and proceed directly to downstream services.
- **Architectural Modularization**:
  - Split core plugin into clean decoupled components:
    - `config.go`: Schemas, default regex rules, and builder factory.
    - `ip_filter.go`: Dedicated IP address and CIDR subnet evaluation engine.
    - `path_normalizer.go`: Anti-evasion path normalizer and sanitizer.
    - `response_handler.go`: Multi-mode response engine (JSON, HTML, Captcha, Redirect, Text, Silent Drop).
    - `routewarden.go`: Middleware coordinator implementing Traefik's `http.Handler`.
- **Per-File Test Suites & Integration Pipeline**:
  - Split test coverage into dedicated files: `config_test.go`, `ip_filter_test.go`, `path_normalizer_test.go`, `response_handler_test.go`, and `routewarden_test.go`.
  - Added `integration_test.go` simulating a multi-middleware Traefik pipeline.
  - Increased statement test coverage to **92.6%**.
- **Branding & Visual Assets**:
  - Minimalist animated SVG line-art icon (`assets/icon.svg`).
  - GitHub social preview banner (`assets/banner.png`).
  - Architecture diagram (`assets/architecture.png`).

---

## [v0.1.x Series] — Legacy

### [v0.1.0] - 2026-09-16

#### Added
- **Core Middleware Engine**:
  - Traefik middleware conforming to Yaegi interpreter specifications using Go standard library (`net/http`, `regexp`, `context`).
  - Factory functions `CreateConfig()` and `New()`.
- **Sensitive Path & Extension Blocking**:
  - Default rule set for blocking `.env*`, `.git`, `.svn`, `.aws`, `.ssh`, backups (`.bak`, `.backup`, `.sql`, `.tar.gz`, `.zip`), configs (`.conf`, `.config`, `.ini`, `.yaml`, `.yml`), logs (`.log`), and debug/status endpoints (`phpinfo.php`, `/actuator/*`).
  - Configurable `pathPatterns` and `blockPatterns` for custom regex matching.
  - Configurable `allowPatterns` override list (defaults include `/robots.txt`, `/ads.txt`, `/security.txt`, and `/.well-known/*`).
- **Initial Response Actions**:
  - Support for `json`, `html`, `redirect`, and `text` modes.
  - Configurable status code (default `403`) and response headers.
- **Initial Anti-Evasion**:
  - Basic URL unescaping, backslash normalization, and semicolon matrix parameter stripping.
