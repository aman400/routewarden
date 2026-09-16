# Getting Started with RouteWarden (v0.1.x)

::: warning Legacy Version Notice
You are currently viewing the documentation for the **v0.1.x series**.  
To read the documentation for the latest release, please switch to [Latest (v0.2.x)](/guide/getting-started).
:::

**RouteWarden v0.1.x** provides core sensitive file and path regex blocking for Traefik with support for custom JSON, HTML, Captcha, Redirect, and Text responses.

---

## Installation & Traefik Setup

### 1. Static Configuration (`traefik.yml`)

```yaml
experimental:
  plugins:
    routewarden:
      moduleName: github.com/aman400/routewarden
      version: v0.1.0
```

---

## Dynamic Configuration Example (`dynamic_conf.yml`)

```yaml
http:
  middlewares:
    route-shield:
      plugin:
        routewarden:
          enabled: true
          enableDefaultPatterns: true
          response:
            mode: json
            statusCode: 403
            body: '{"error":"Forbidden","version":"v0.1.0"}'

  routers:
    app-router:
      rule: "Host(`example.com`)"
      entryPoints:
        - web
      middlewares:
        - route-shield
      service: app-service
```
