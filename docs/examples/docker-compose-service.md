# Example 3: Service-Level Docker Compose

When individual microservices require custom regex rules, sensitive directory exceptions, query inspection, or dedicated error payloads, configure RouteWarden at the service router level.

---

## Configuration Preview

::: code-group

```yaml [File (YAML)]
# dynamic_conf.yml
http:
  middlewares:
    service-warden:
      plugin:
        routewarden:
          enabled: true
          enableDefaultPatterns: true
          checkQuery: true
          pathPatterns:
            - '(?i)^/admin(/.*)?$'
            - '(?i)^/api/internal(/.*)?$'
          allowPatterns:
            - '(?i)^/robots\.txt$'
            - '(?i)^/\.well-known(/.*)?$'
          allowedIps:
            - "192.168.1.0/24"
            - "10.10.0.0/16"
          response:
            mode: json
            statusCode: 403
            body: '{"error":"access_denied","service":"web"}'
            headers:
              X-Protected-By: "RouteWarden"

  routers:
    web-router:
      rule: "Host(`example.com`)"
      entryPoints:
        - web
      middlewares:
        - service-warden
      service: web-service
```

```toml [File (TOML)]
# dynamic_conf.toml
[http.routers.web-router]
  rule = "Host(`example.com`)"
  entryPoints = ["web"]
  middlewares = ["service-warden"]
  service = "web-service"

[http.middlewares.service-warden.plugin.routewarden]
  enabled = true
  enableDefaultPatterns = true
  checkQuery = true
  pathPatterns = ["(?i)^/admin(/.*)?$", "(?i)^/api/internal(/.*)?$"]
  allowPatterns = ["(?i)^/robots\\.txt$", "(?i)^/\\.well-known(/.*)?$"]
  allowedIps = ["192.168.1.0/24", "10.10.0.0/16"]

[http.middlewares.service-warden.plugin.routewarden.response]
  mode = "json"
  statusCode = 403
  body = '{"error":"access_denied","service":"web"}'

[http.middlewares.service-warden.plugin.routewarden.response.headers]
  X-Protected-By = "RouteWarden"
```

```bash [CLI]
# Docker Compose Labels / CLI equivalent
- "traefik.enable=true"
- "traefik.http.routers.web.rule=Host(`example.com`)"
- "traefik.http.routers.web.entrypoints=web"
- "traefik.http.routers.web.middlewares=service-warden"
- "traefik.http.middlewares.service-warden.plugin.routewarden.enabled=true"
- "traefik.http.middlewares.service-warden.plugin.routewarden.enableDefaultPatterns=true"
- "traefik.http.middlewares.service-warden.plugin.routewarden.checkQuery=true"
- "traefik.http.middlewares.service-warden.plugin.routewarden.pathPatterns=(?i)^/admin(/.*)?$,(?i)^/api/internal(/.*)?$"
- "traefik.http.middlewares.service-warden.plugin.routewarden.allowPatterns=(?i)^/robots\\.txt$,(?i)^/\\.well-known(/.*)?$"
- "traefik.http.middlewares.service-warden.plugin.routewarden.allowedIps=192.168.1.0/24,10.10.0.0/16"
- "traefik.http.middlewares.service-warden.plugin.routewarden.response.mode=json"
- "traefik.http.middlewares.service-warden.plugin.routewarden.response.statusCode=403"
- 'traefik.http.middlewares.service-warden.plugin.routewarden.response.body={"error":"access_denied","service":"web"}'
- "traefik.http.middlewares.service-warden.plugin.routewarden.response.headers.X-Protected-By=RouteWarden"
```

:::

---

## Docker Compose Example

```yaml
services:
  web:
    image: my-web-app:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.web.rule=Host(`example.com`)"
      - "traefik.http.routers.web.entrypoints=web"
      # Attach service-specific middleware
      - "traefik.http.routers.web.middlewares=service-warden"

      # RouteWarden Middleware Definition
      - "traefik.http.middlewares.service-warden.plugin.routewarden.enabled=true"
      - "traefik.http.middlewares.service-warden.plugin.routewarden.enableDefaultPatterns=true"
      - "traefik.http.middlewares.service-warden.plugin.routewarden.checkQuery=true"
      # Block internal/debug routes specifically for this application
      - "traefik.http.middlewares.service-warden.plugin.routewarden.pathPatterns=(?i)^/admin(/.*)?$,(?i)^/api/internal(/.*)?$"
      # Safe exceptions for public robot & ACME challenges
      - "traefik.http.middlewares.service-warden.plugin.routewarden.allowPatterns=(?i)^/robots\\.txt$,(?i)^/\\.well-known(/.*)?$"
      # Trusted internal office network
      - "traefik.http.middlewares.service-warden.plugin.routewarden.allowedIps=192.168.1.0/24,10.10.0.0/16"
      # Custom JSON response structure
      - "traefik.http.middlewares.service-warden.plugin.routewarden.response.mode=json"
      - "traefik.http.middlewares.service-warden.plugin.routewarden.response.statusCode=403"
      - "traefik.http.middlewares.service-warden.plugin.routewarden.response.body={\"error\":\"access_denied\",\"service\":\"web\"}"
      - "traefik.http.middlewares.service-warden.plugin.routewarden.response.headers.X-Protected-By=RouteWarden"
```
