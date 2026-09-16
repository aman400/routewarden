# Getting Started with RouteWarden

**RouteWarden** is a high-performance Traefik middleware plugin written in pure Go, designed to intercept and block unauthorized reconnaissance, directory probing, and access to sensitive files before requests ever hit your backend services.

---

## Key Capabilities

- **Automated Sensitive Asset Shielding**: Blocks attempts to access environment configurations (`.env`), VCS repositories (`.git`, `.svn`), credentials (`.aws`, `.ssh`), database dumps (`.sql`, `.bak`), application configurations (`.yaml`, `.conf`, `.ini`), and debug panels (`phpinfo.php`, `/actuator`).
- **Anti-Evasion Engine**: Proactively detects and decodes layered URL encoding tricks (`%252e%252e`), semicolon path matrix parameters (`/;param/.env`), backslash separators (`\..\`), and null bytes (`%00`).
- **IP & CIDR Subnet Allowlisting**: Exempts internal networks, VPN gateways, and developer machines from path blocking.
- **Custom Responses & Captcha**: Return custom JSON error structures, custom branded HTML 404 pages, or challenge clients via **Cloudflare Turnstile**, **hCaptcha**, or **reCAPTCHA**.

---

## Supported Traefik Versions

| Traefik Version | Compatibility | Notes |
|---|---|---|
| **Traefik v3.x** (v3.0, v3.1, v3.2+) | ✅ **Supported** | Full support for Traefik v3 runtime, CLI flags, Docker Compose labels, and IngressRoute CRDs. |
| **Traefik v2.x** (v2.8 – v2.11+) | ✅ **Supported** | Fully compatible with Traefik v2 plugin mechanism. |
| **Traefik v1.x** | ❌ **Not Supported** | External Yaegi middleware plugins are not available in Traefik v1. |

---

## Installation & Traefik Setup

### 1. Static Configuration

Declare RouteWarden in Traefik's plugins configuration:

::: code-group

```yaml [File (YAML)]
# traefik.yml
experimental:
  plugins:
    routewarden:
      moduleName: github.com/aman400/routewarden
      version: {{version}}
```

```toml [File (TOML)]
# traefik.toml
[experimental.plugins.routewarden]
  moduleName = "github.com/aman400/routewarden"
  version = "{{version}}"
```

```bash [CLI]
traefik \
  --experimental.plugins.routewarden.modulename=github.com/aman400/routewarden \
  --experimental.plugins.routewarden.version={{version}}
```

:::

> **Local Development (`localPlugins`)**:
> ::: code-group
> ```yaml [File (YAML)]
> experimental:
>   localPlugins:
>     routewarden:
>       moduleName: github.com/aman400/routewarden
> ```
> ```toml [File (TOML)]
> [experimental.localPlugins.routewarden]
>   moduleName = "github.com/aman400/routewarden"
> ```
> ```bash [CLI]
> traefik --experimental.localplugins.routewarden.modulename=github.com/aman400/routewarden
> ```
> :::

---

### 2. Dynamic Configuration

Configure the RouteWarden middleware and attach it to your router:

::: code-group

```yaml [File (YAML)]
# dynamic_conf.yml
http:
  middlewares:
    route-shield:
      plugin:
        routewarden:
          enabled: true
          enableDefaultPatterns: true
          allowedIps:
            - "127.0.0.1"
            - "10.0.0.0/8"
          response:
            mode: json
            statusCode: 403
            body: '{"error":"Forbidden","message":"Sensitive route protected by RouteWarden"}'

  routers:
    app-router:
      rule: "Host(`app.example.com`)"
      entryPoints:
        - web
      middlewares:
        - route-shield
      service: app-service
```

```toml [File (TOML)]
# dynamic_conf.toml
[http.routers.app-router]
  rule = "Host(`app.example.com`)"
  entryPoints = ["web"]
  middlewares = ["route-shield"]
  service = "app-service"

[http.middlewares.route-shield.plugin.routewarden]
  enabled = true
  enableDefaultPatterns = true
  allowedIps = ["127.0.0.1", "10.0.0.0/8"]

[http.middlewares.route-shield.plugin.routewarden.response]
  mode = "json"
  statusCode = 403
  body = '{"error":"Forbidden","message":"Sensitive route protected by RouteWarden"}'
```

```bash [CLI]
# Note: Dynamic configurations in Traefik can also be declared via Docker Compose labels or CLI
traefik \
  --entrypoints.web.address=:80 \
  --entrypoints.web.http.middlewares=route-shield@docker
```

:::

---

## Next Steps

- Explore [System Architecture](/guide/architecture) to understand the request inspection pipeline.
- View the complete [Configuration Reference](/reference/configuration).
- Check the [Examples & Wiki Cookbook](/examples/overview) for production Docker Compose & Kubernetes blueprints.
