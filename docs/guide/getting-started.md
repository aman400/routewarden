# Getting Started with RouteWarden

**RouteWarden** is a high-performance Traefik middleware plugin written in pure Go, designed to intercept and block unauthorized reconnaissance, directory probing, and access to sensitive files before requests ever hit your backend services.

---

## Key Capabilities

- **Automated Sensitive Asset Shielding**: Blocks attempts to access environment configurations (`.env`), VCS repositories (`.git`, `.svn`), credentials (`.aws`, `.ssh`), database dumps (`.sql`, `.bak`), application configurations (`.yaml`, `.conf`, `.ini`), and debug panels (`phpinfo.php`, `/actuator`).
- **Anti-Evasion Engine**: Proactively detects and decodes layered URL encoding tricks (`%252e%252e`), semicolon path matrix parameters (`/;param/.env`), backslash separators (`\..\`), and null bytes (`%00`).
- **IP & CIDR Subnet Allowlisting**: Exempts internal networks, VPN gateways, and developer machines from path blocking.
- **Custom Responses & Captcha**: Return custom JSON error structures, custom branded HTML 404 pages, or challenge clients via **Cloudflare Turnstile**, **hCaptcha**, or **reCAPTCHA**.

---

## Installation & Traefik Setup

### 1. Static Configuration (`traefik.yml`)

Declare RouteWarden in Traefik's plugins section:

```yaml
experimental:
  plugins:
    routewarden:
      moduleName: github.com/aman400/routewarden
      version: v0.2.0
```

If you are developing locally:

```yaml
experimental:
  localPlugins:
    routewarden:
      moduleName: github.com/aman400/routewarden
```

---

### 2. Dynamic Configuration (`dynamic_conf.yml`)

```yaml
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

---

## Next Steps

- Explore [System Architecture](/guide/architecture) to understand the request inspection pipeline.
- View the complete [Configuration Reference](/reference/configuration).
- Check the [Examples & Wiki Cookbook](/examples/overview) for production Docker Compose & Kubernetes blueprints.
