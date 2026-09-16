# Changelog

All notable changes to the **RouteWarden** Traefik middleware plugin will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.2.1] - 2026-09-16

### Added
- **Interactive Documentation & Wiki Site (VitePress)**:
  - Launched official documentation site hosted on GitHub Pages ([`https://aman400.github.io/routewarden/`](https://aman400.github.io/routewarden/)).
  - Built-in client-side full-text search, responsive dark/light mode, and copy-paste code blocks.
  - Comprehensive guides covering:
    - **Getting Started & Traefik Setup** (`/guide/getting-started`)
    - **System Architecture & Request Pipeline** (`/guide/architecture`)
    - **Local Development & Deployment Walkthrough** (`/guide/local-deployment`)
    - **Automated Testing & Coverage Architecture** (`/guide/testing`)
    - **Anti-Evasion Engine Security Analysis** (`/reference/anti-evasion`)
    - **Full Configuration Options Reference** (`/reference/configuration`)
- **Ready-to-Run In-Repo Examples Suite (`examples/`)**:
  - `01-basic-sensitive-files`: Quickstart protecting backend Nginx against `.env`, `.git`, backups, and configs.
  - `02-global-entrypoint-shield`: Global entrypoint middleware shielding all services across the cluster without per-container labels.
  - `03-ip-whitelist-vpn`: Bypassing security checks for trusted CIDR / office IP networks.
  - `04-captcha-challenge`: Side-by-side **hCaptcha** and **Cloudflare Turnstile** verification challenges (confirming no external `captcha.html` file is required).
  - `05-kubernetes-ingressroute`: Production Traefik Kubernetes `Middleware` and `IngressRoute` CRD manifests.
- **Automated GitHub Pages CI/CD Pipeline**:
  - Added [`.github/workflows/deploy-docs.yml`](.github/workflows/deploy-docs.yml) using `@actions/deploy-pages` to build and publish docs on push.

### Changed
- **Streamlined `README.md`**:
  - Simplified landing page with a direct quickstart for returning **404 Not Found** error payloads.
  - Added concise **Basic Configuration Options** summary table.
  - Added prominent documentation badges and deep links to the VitePress documentation wiki.

---

## [v0.2.0] - 2026-09-16

### Added
- **IP & CIDR Subnet Whitelisting (`allowedIps`)**:
  - Added `allowedIps` configuration supporting exact IPv4 addresses, IPv6 addresses, and CIDR subnet masks (e.g., `127.0.0.1`, `10.0.0.0/8`, `2001:db8::/32`).
  - Implemented client IP resolution with proxy forwarding support (`X-Forwarded-For`, `X-Real-IP`, and socket `RemoteAddr`).
  - Requests originating from whitelisted IPs/subnets bypass sensitive route blocking and proceed directly to downstream services.
- **Architectural Refactoring & Modularization**:
  - Split core plugin implementation into modular, decoupled components:
    - `config.go`: Configuration schemas, default regex rules, and builder factory.
    - `ip_filter.go`: Dedicated IP address and CIDR subnet evaluation engine.
    - `path_normalizer.go`: Comprehensive anti-evasion path sanitizer.
    - `response_handler.go`: Multi-mode response engine (JSON, HTML, Captcha, Redirect, Text, Silent Drop).
    - `routewarden.go`: Clean middleware coordinator implementing Traefik's `http.Handler` interface.
- **Comprehensive Per-File Test Suites & Integration Pipeline**:
  - Split test coverage into dedicated files: `config_test.go`, `ip_filter_test.go`, `path_normalizer_test.go`, `response_handler_test.go`, and `routewarden_test.go`.
  - Added `integration_test.go` simulating a full multi-middleware Traefik pipeline (Tracing ➡️ RouteWarden ➡️ Backend Service).
  - Increased statement test coverage to **92.6%**.
- **Branding & Visual Assets**:
  - Added minimalist animated SVG line-art icon ([`assets/icon.svg`](assets/icon.svg)).
  - Added GitHub social preview banner ([`assets/banner.png`](assets/banner.png)).
  - Added technical architecture flow diagram ([`assets/architecture.png`](assets/architecture.png)).
  - Updated Traefik plugin catalog manifest (`.traefik.yml`) to reference vector SVG icon.

---

## [v0.1.0] - 2026-09-16

### Added
- **Core Middleware Engine**:
  - Traefik middleware implementation conforming to Yaegi interpreter specifications using Go standard library (`net/http`, `regexp`, `context`, etc.).
  - Factory functions `CreateConfig()` and `New()`.
- **Sensitive Path & Extension Blocking**:
  - Out-of-the-box rule set for blocking `.env*`, `.git`, `.svn`, `.aws`, `.ssh`, backups (`.bak`, `.backup`, `.sql`, `.tar.gz`, `.zip`), configs (`.conf`, `.config`, `.ini`, `.yaml`, `.yml`), logs (`.log`), lockfiles, and debug/status endpoints (`phpinfo.php`, `/actuator/*`).
  - Configurable `pathPatterns` and `blockPatterns` for custom regex matching.
  - Configurable `allowPatterns` override list (defaults include `/robots.txt`, `/ads.txt`, `/security.txt`, and `/.well-known/*`).
- **Flexible Response Actions**:
  - **`json`**: Return custom JSON payloads with automatic `application/json` Content-Type and custom HTTP status codes.
  - **`html`**: Return custom branded HTML warning or error pages with `text/html`.
  - **`captcha`**: Present responsive Captcha challenges with support for **Cloudflare Turnstile**, **hCaptcha**, **Google reCAPTCHA**, and custom challenge templates.
  - **`redirect`**: Redirect blocked requests to honeypot or security warning URLs.
  - **`text`**: Standard text responses.
  - **`silentDrop`**: Immediate TCP connection termination or empty payload.
  - Custom HTTP status codes (e.g., 401, 403, 404, 418, 429).
  - Custom response headers injection (`response.headers`).
- **Anti-Evasion & Security Hardening**:
  - Multi-layer iterative URL unescaping to defeat double URL encoding bypasses (`%252e%252e`, `%252eenv`).
  - Semicolon matrix parameter handling and segmentation to defeat reverse-proxy bypasses (`/;param/.env`, `/endpoint;jsessionid=.../.env`).
  - Backslash normalization to prevent Windows/IIS-style path separator evasion (`/static\..\.env`).
  - Encoded null byte protection (`%00`).
  - Query parameter inspection when `checkQuery` is enabled.
- **Developer & Deployment Artifacts**:
  - Traefik plugin catalog manifest (`.traefik.yml`).
  - GitHub Actions CI workflow for automated testing with race detection (`.github/workflows/test.yml`).
  - Comprehensive table-driven unit tests and security evasion test suite (`routewarden_test.go`).
  - Documentation and configuration examples for Traefik v2/v3, Docker Compose, and Kubernetes IngressRoute (`README.md`).

[v0.2.1]: https://github.com/aman400/routewarden/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/aman400/routewarden/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/aman400/routewarden/releases/tag/v0.1.0

