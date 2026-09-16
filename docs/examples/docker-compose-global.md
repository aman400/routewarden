# Example 2: Global EntryPoint Shield

Attaching RouteWarden directly to Traefik's entrypoint provides unified, cluster-wide protection across all services without requiring repetitive labels on individual containers.

---

## Docker Compose Configuration

```yaml
services:
  traefik:
    image: traefik:v3.1
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--experimental.plugins.routewarden.modulename=github.com/aman400/routewarden"
      - "--experimental.plugins.routewarden.version={{version}}"
      # Attach global-warden middleware to all incoming traffic on entrypoint
      - "--entrypoints.web.http.middlewares=global-warden@docker"
    ports:
      - "80:80"
      - "8080:8080"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
    labels:
      - "traefik.enable=true"
      - "traefik.http.middlewares.global-warden.plugin.routewarden.enabled=true"
      - "traefik.http.middlewares.global-warden.plugin.routewarden.enableDefaultPatterns=true"
      - "traefik.http.middlewares.global-warden.plugin.routewarden.allowedIps=127.0.0.1,10.0.0.0/8"
      - "traefik.http.middlewares.global-warden.plugin.routewarden.response.mode=json"
      - "traefik.http.middlewares.global-warden.plugin.routewarden.response.statusCode=403"
      - "traefik.http.middlewares.global-warden.plugin.routewarden.response.body={\"error\":\"Forbidden\",\"scope\":\"global-shield\"}"

  service-frontend:
    image: nginx:alpine
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.frontend.rule=Host(`frontend.localhost`)"
      - "traefik.http.routers.frontend.entrypoints=web"

  service-backend:
    image: nginx:alpine
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.backend.rule=Host(`api.localhost`)"
      - "traefik.http.routers.backend.entrypoints=web"
```

Both `frontend.localhost` and `api.localhost` are guarded immediately.
