# Changelog

All notable changes to the **RouteWarden** Traefik middleware plugin will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[v0.1.0]: https://github.com/aman400/routewarden/releases/tag/v0.1.0
