# Example 3: Service-Level Docker Compose

When individual microservices require custom regex rules, sensitive directory exceptions, query inspection, or dedicated error payloads, configure RouteWarden at the service router level.

---

## Docker Compose Configuration

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
