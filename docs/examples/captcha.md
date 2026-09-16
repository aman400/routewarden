# Example 5: Captcha Challenge (Turnstile / hCaptcha / reCAPTCHA)

Instead of dropping connections or returning static error codes, RouteWarden can serve interactive Captcha challenges on sensitive paths using **Cloudflare Turnstile**, **hCaptcha**, or **Google reCAPTCHA**.

---

## Is a `captcha.html` File Required?

> [!TIP]
> **No external `captcha.html` file is required!**  
> RouteWarden has a **built-in, mobile-responsive dark-mode HTML template** embedded directly into the Go binary. When `mode: captcha` is enabled, RouteWarden automatically:
> 1. Injects the official provider JavaScript SDK (`https://js.hcaptcha.com/1/api.js` for hCaptcha or Cloudflare/Google equivalent).
> 2. Renders the appropriate widget container (`<div class="h-captcha" data-sitekey="..."></div>`).
> 3. Populates your custom title and site key.
>
> *(Optional: If you ever want to override the design with your own custom layout, you can pass an HTML template string into `response.captcha.template`).*

---

## Supported Providers

| Provider | `response.captcha.provider` | Injected SDK Script | Widget Class |
|---|---|---|---|
| **hCaptcha** | `hcaptcha` | `https://js.hcaptcha.com/1/api.js` | `<div class="h-captcha">` |
| **Cloudflare Turnstile** | `turnstile` | `https://challenges.cloudflare.com/turnstile/v0/api.js` | `<div class="cf-turnstile">` |
| **Google reCAPTCHA v2** | `recaptcha` | `https://www.google.com/recaptcha/api.js` | `<div class="g-recaptcha">` |

---

## 1. hCaptcha Configuration Example

This example protects `/admin` and `/login` with **hCaptcha** (using the official hCaptcha test site key `10000000-ffff-ffff-ffff-000000000001`):

::: code-group

```yaml [File (YAML)]
# dynamic_conf.yml
http:
  middlewares:
    hcaptcha-barrier:
      plugin:
        routewarden:
          enabled: true
          pathPatterns:
            - '(?i)^/(admin|login)(/.*)?$'
          response:
            mode: captcha
            statusCode: 403
            captcha:
              provider: "hcaptcha"
              siteKey: "10000000-ffff-ffff-ffff-000000000001"
              title: "Human Verification (hCaptcha)"

  routers:
    app-router:
      rule: "Host(`app.example.com`)"
      entryPoints:
        - web
      middlewares:
        - hcaptcha-barrier
      service: app-service
```

```toml [File (TOML)]
# dynamic_conf.toml
[http.routers.app-router]
  rule = "Host(`app.example.com`)"
  entryPoints = ["web"]
  middlewares = ["hcaptcha-barrier"]
  service = "app-service"

[http.middlewares.hcaptcha-barrier.plugin.routewarden]
  enabled = true
  pathPatterns = ["(?i)^/(admin|login)(/.*)?$"]

[http.middlewares.hcaptcha-barrier.plugin.routewarden.response]
  mode = "captcha"
  statusCode = 403

[http.middlewares.hcaptcha-barrier.plugin.routewarden.response.captcha]
  provider = "hcaptcha"
  siteKey = "10000000-ffff-ffff-ffff-000000000001"
  title = "Human Verification (hCaptcha)"
```

```bash [CLI]
# Docker Compose Labels / CLI equivalent
- "traefik.enable=true"
- "traefik.http.routers.app.rule=Host(`app.example.com`)"
- "traefik.http.routers.app.middlewares=hcaptcha-barrier"
- "traefik.http.middlewares.hcaptcha-barrier.plugin.routewarden.enabled=true"
- "traefik.http.middlewares.hcaptcha-barrier.plugin.routewarden.pathPatterns=(?i)^/(admin|login)(/.*)?$"
- "traefik.http.middlewares.hcaptcha-barrier.plugin.routewarden.response.mode=captcha"
- "traefik.http.middlewares.hcaptcha-barrier.plugin.routewarden.response.statusCode=403"
- "traefik.http.middlewares.hcaptcha-barrier.plugin.routewarden.response.captcha.provider=hcaptcha"
- "traefik.http.middlewares.hcaptcha-barrier.plugin.routewarden.response.captcha.siteKey=10000000-ffff-ffff-ffff-000000000001"
- "traefik.http.middlewares.hcaptcha-barrier.plugin.routewarden.response.captcha.title=Human Verification (hCaptcha)"
```

:::

---

## 2. Cloudflare Turnstile Configuration Example

::: code-group

```yaml [File (YAML)]
# dynamic_conf.yml
http:
  middlewares:
    turnstile-barrier:
      plugin:
        routewarden:
          enabled: true
          pathPatterns:
            - '(?i)^/login(/.*)?$'
            - '(?i)^/reset-password(/.*)?$'
          response:
            mode: captcha
            statusCode: 403
            captcha:
              provider: "turnstile"
              siteKey: "1x00000000000000000000AA"
              title: "Security Verification Required"

  routers:
    login-router:
      rule: "Host(`login.example.com`)"
      entryPoints:
        - web
      middlewares:
        - turnstile-barrier
      service: login-service
```

```toml [File (TOML)]
# dynamic_conf.toml
[http.routers.login-router]
  rule = "Host(`login.example.com`)"
  entryPoints = ["web"]
  middlewares = ["turnstile-barrier"]
  service = "login-service"

[http.middlewares.turnstile-barrier.plugin.routewarden]
  enabled = true
  pathPatterns = ["(?i)^/login(/.*)?$", "(?i)^/reset-password(/.*)?$"]

[http.middlewares.turnstile-barrier.plugin.routewarden.response]
  mode = "captcha"
  statusCode = 403

[http.middlewares.turnstile-barrier.plugin.routewarden.response.captcha]
  provider = "turnstile"
  siteKey = "1x00000000000000000000AA"
  title = "Security Verification Required"
```

```bash [CLI]
# Docker Compose Labels / CLI equivalent
- "traefik.enable=true"
- "traefik.http.routers.login.rule=Host(`login.example.com`)"
- "traefik.http.routers.login.middlewares=turnstile-barrier"
- "traefik.http.middlewares.turnstile-barrier.plugin.routewarden.enabled=true"
- "traefik.http.middlewares.turnstile-barrier.plugin.routewarden.pathPatterns=(?i)^/login(/.*)?$,(?i)^/reset-password(/.*)?$"
- "traefik.http.middlewares.turnstile-barrier.plugin.routewarden.response.mode=captcha"
- "traefik.http.middlewares.turnstile-barrier.plugin.routewarden.response.statusCode=403"
- "traefik.http.middlewares.turnstile-barrier.plugin.routewarden.response.captcha.provider=turnstile"
- "traefik.http.middlewares.turnstile-barrier.plugin.routewarden.response.captcha.siteKey=1x00000000000000000000AA"
- "traefik.http.middlewares.turnstile-barrier.plugin.routewarden.response.captcha.title=Security Verification Required"
```

:::

---

## 3. Dynamic YAML Example (`dynamic_conf.yml`)

For file-based Traefik setups using hCaptcha:

```yaml
http:
  middlewares:
    hcaptcha-shield:
      plugin:
        routewarden:
          enabled: true
          pathPatterns:
            - '(?i)^/portal/.*'
          response:
            mode: captcha
            statusCode: 403
            captcha:
              provider: "hcaptcha"
              siteKey: "10000000-ffff-ffff-ffff-000000000001"
              title: "Verification Challenge"
```

