<div align="center">
  <img src="assets/icon.svg" alt="RouteWarden Logo" width="140" height="140" />
  <h1>RouteWarden</h1>
  <p><strong>High-performance Traefik middleware to stop sensitive file exposure (.env, .git, backups), neutralize path-evasion attacks, whitelist IPs, and serve custom error/captcha responses before requests reach your backend.</strong></p>
</div>

<p align="center">
  <a href="https://github.com/aman400/routewarden/releases"><img src="https://img.shields.io/github/v/release/aman400/routewarden?color=blue" alt="GitHub Release" /></a>
  <a href="https://traefik.io"><img src="https://img.shields.io/badge/Traefik-v2.x%20%7C%20v3.x-24A1C1.svg?logo=traefik&logoColor=white" alt="Traefik Compatibility: v2.x | v3.x" /></a>
  <a href="https://pkg.go.dev/github.com/aman400/routewarden"><img src="https://pkg.go.dev/badge/github.com/aman400/routewarden.svg" alt="Go Reference" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT" /></a>
  <a href="https://goreportcard.com/report/github.com/aman400/routewarden"><img src="https://goreportcard.com/badge/github.com/aman400/routewarden" alt="Go Report Card" /></a>
  <a href="https://aman400.github.io/routewarden/"><img src="https://img.shields.io/badge/Docs-VitePress%20Wiki-6366f1.svg" alt="Documentation Site" /></a>
</p>

---

> 📖 **Full Documentation, Guides & Wiki**: [https://aman400.github.io/routewarden/](https://aman400.github.io/routewarden/)  
> 📂 **Runnable Scenarios**: [`examples/`](examples/) *(Docker Compose & Kubernetes CRDs)*

---

## Supported Traefik Versions

| Traefik Version | Status | Notes |
|---|---|---|
| **Traefik v3.x** (v3.0, v3.1, v3.2+) | ✅ **Fully Supported** | Standard Yaegi runtime, Docker labels & Kubernetes CRDs |
| **Traefik v2.x** (v2.8 – v2.11+) | ✅ **Fully Supported** | Compatible with standard plugin mechanism |
| **Traefik v1.x** | ❌ **Not Supported** | Plugins are not supported in Traefik v1 |

---

## What is RouteWarden?

**RouteWarden** is a lightweight Traefik middleware written in pure Go (with zero external dependencies) that intercepts and blocks requests before they reach your backend:

- 🛡️ **Zero-Config Defense**: Blocks `.env*`, `.git`, `.aws`, `.sql`, `.bak`, `.conf`, `.yaml`, logs, and debug endpoints.
- ⚡ **Anti-Evasion**: Normalizes double-URL encoding (`%252e%252e`), semicolon matrix params (`/;param/.env`), and Windows backslashes (`\`).
- 🌐 **IP & CIDR Whitelist**: Bypass blocking for corporate VPNs, office IPs, or developer subnets (`10.0.0.0/8`).
- 🎭 **Flexible Responses**: Return custom **404 Not Found**, **403 Forbidden**, custom JSON, HTML, **302 Redirect**, or interactive **Turnstile / hCaptcha / reCAPTCHA** challenges.

---

## Quick Start (404 Response Example)

The cleanest way to handle reconnaissance bots is returning a standard **404 Not Found** so attackers believe the file does not exist.

### Option A: Docker Compose

```yaml
services:
  traefik:
    image: traefik:v3.1
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
      - "--experimental.plugins.routewarden.modulename=github.com/aman400/routewarden"
      - "--experimental.plugins.routewarden.version=v0.2.2"
    ports:
      - "80:80"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"

  webapp:
    image: nginx:alpine
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.webapp.rule=Host(`localhost`)"
      - "traefik.http.routers.webapp.entrypoints=web"
      - "traefik.http.routers.webapp.middlewares=warden-shield"

      # RouteWarden Configuration
      # (Default: true) Enable or disable middleware
      - "traefik.http.middlewares.warden-shield.plugin.routewarden.enabled=true"
      # (Default: true) Block sensitive files (.env*, .git, .aws, .sql, .bak, .log, configs)
      - "traefik.http.middlewares.warden-shield.plugin.routewarden.enableDefaultPatterns=true"
      # (Optional) Custom regex patterns to block (Default: [])
      - "traefik.http.middlewares.warden-shield.plugin.routewarden.pathPatterns=(?i)^/admin(/.*)?$,(?i)^/api/internal(/.*)?$"
      # (Optional) Safe exception overrides to allow (Default: robots.txt, ads.txt, sitemap.xml, .well-known/*)
      - "traefik.http.middlewares.warden-shield.plugin.routewarden.allowPatterns=(?i)^/api/internal/health$,(?i)^/robots\\.txt$"
      # Return a clean 404 response (Default mode: text, Default statusCode: 403)
      - "traefik.http.middlewares.warden-shield.plugin.routewarden.response.mode=text"
      - "traefik.http.middlewares.warden-shield.plugin.routewarden.response.statusCode=404"
      - "traefik.http.middlewares.warden-shield.plugin.routewarden.response.body=404 page not found"
```

---

### Option B: Traefik Dynamic Configuration (`dynamic_conf.yml`)

#### 1. Static Configuration (`traefik.yml`)
```yaml
experimental:
  plugins:
    routewarden:
      moduleName: github.com/aman400/routewarden
      version: v0.2.2
```

#### 2. Dynamic Configuration (`dynamic_conf.yml`)
```yaml
http:
  middlewares:
    warden-404:
      plugin:
        routewarden:
          enabled: true                # Default: true
          enableDefaultPatterns: true  # Default: true (.env*, .git, .aws, .sql, .bak, etc.)
          # (Optional) Custom regex patterns to block (Default: [])
          pathPatterns:
            - '(?i)^/admin(/.*)?$'
            - '(?i)^/api/internal(/.*)?$'
          # (Optional) Safe exceptions to allow (Default: robots.txt, ads.txt, sitemap.xml, .well-known/*)
          allowPatterns:
            - '(?i)^/api/internal/health$'
            - '(?i)^/robots\.txt$'
          # (Optional) Trusted developer/VPN IP bypass (Default: [])
          allowedIps:
            - "127.0.0.1"
            - "10.0.0.0/8"
          # Response action (Default mode: text, Default statusCode: 403)
          response:
            mode: text
            statusCode: 404
            body: "404 page not found"

  routers:
    app-router:
      rule: "Host(`app.example.com`)"
      entryPoints:
        - web
      middlewares:
        - warden-404
      service: app-service
```

---

## Basic Configuration Options

| Option | Type | Default | Description |
|---|---|---|---|
| `enabled` | `bool` | `true` | Turn the middleware on or off. |
| `enableDefaultPatterns` | `bool` | `true` | Block common sensitive files (`.env*`, `.git`, `.aws`, `.sql`, `.bak`, `.log`, configs). |
| `pathPatterns` | `[]string` | `[]` | Additional custom regex patterns to block (e.g. `['(?i)^/admin/.*']`). |
| `allowPatterns` | `[]string` | `[...]` | Safe regex overrides (defaults: `/robots.txt`, `/ads.txt`, `/.well-known/*`). |
| `allowedIps` | `[]string` | `[]` | Whitelisted IPv4/IPv6 addresses or CIDR subnets (e.g. `127.0.0.1`, `10.0.0.0/8`). |
| `checkQuery` | `bool` | `false` | Also inspect query parameters for blocked patterns. |
| `response.mode` | `string` | `"text"` | Action on block: `"text"`, `"json"`, `"html"`, `"captcha"`, `"redirect"`, or `"silentDrop"`. |
| `response.statusCode` | `int` | `403` | HTTP status code returned to client (e.g. `404`, `403`, `401`, `429`). |
| `response.body` | `string` | `""` | Custom payload returned in the response body. |

> 💡 For the complete list of settings (including Captcha providers, custom HTML templates, and header injection), visit the **[Full Configuration Reference](https://aman400.github.io/routewarden/reference/configuration)**.

---

## Documentation & Advanced Examples

For in-depth setup guides, anti-evasion architecture, and ready-to-run blueprints, visit our **[Documentation Wiki](https://aman400.github.io/routewarden/)**:

- 📖 **[Getting Started & Installation Guide](https://aman400.github.io/routewarden/guide/getting-started)**
- 🏛️ **[System Architecture & Pipeline](https://aman400.github.io/routewarden/guide/architecture)**
- 💻 **[Local Development & Testing Guide](https://aman400.github.io/routewarden/guide/local-deployment)**
- 🧪 **[Automated Testing & Coverage Architecture](https://aman400.github.io/routewarden/guide/testing)**
- ⚙️ **[Full Configuration Options Table](https://aman400.github.io/routewarden/reference/configuration)**
- 🎯 **[Custom Path Patterns & Regex Guide](https://aman400.github.io/routewarden/reference/custom-paths)**
- 🛡️ **[Anti-Evasion Engine (Encoding, Matrix Params, Traversals)](https://aman400.github.io/routewarden/reference/anti-evasion)**
- 🚀 **[Global EntryPoint Shield Cookbook](https://aman400.github.io/routewarden/examples/docker-compose-global)**
- 🌐 **[IP & CIDR Subnet Whitelisting Cookbook](https://aman400.github.io/routewarden/examples/ip-whitelisting)**
- 🤖 **[Cloudflare Turnstile & hCaptcha Challenges](https://aman400.github.io/routewarden/examples/captcha)**
- ☸️ **[Kubernetes IngressRoute CRD Example](https://aman400.github.io/routewarden/examples/kubernetes)**

---

## License

This project is licensed under the [MIT License](LICENSE).
