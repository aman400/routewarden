# Example 1: Basic Sensitive File Blocking

This scenario protects a web application against reconnaissance and exposure of critical infrastructure files using RouteWarden's built-in rule dictionary.

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
      - "--experimental.plugins.routewarden.version=v0.2.0"
    ports:
      - "80:80"
      - "8080:8080"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"

  webapp:
    image: nginx:alpine
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.webapp.rule=Host(`localhost`)"
      - "traefik.http.routers.webapp.entrypoints=web"
      - "traefik.http.routers.webapp.middlewares=warden-shield"

      # RouteWarden Setup
      - "traefik.http.middlewares.warden-shield.plugin.routewarden.enabled=true"
      - "traefik.http.middlewares.warden-shield.plugin.routewarden.enableDefaultPatterns=true"
      - "traefik.http.middlewares.warden-shield.plugin.routewarden.response.mode=json"
      - "traefik.http.middlewares.warden-shield.plugin.routewarden.response.statusCode=403"
      - "traefik.http.middlewares.warden-shield.plugin.routewarden.response.body={\"error\":\"Forbidden\",\"message\":\"Sensitive path blocked by RouteWarden\"}"
```

---

## Verification Commands

```bash
# Legitimate homepage access (Allowed)
curl -I http://localhost/
# Output: HTTP/1.1 200 OK

# Probing for environment secrets (Blocked)
curl -i http://localhost/.env
# Output: HTTP/1.1 403 Forbidden
# {"error":"Forbidden","message":"Sensitive path blocked by RouteWarden"}

# Probing for Git repository details (Blocked)
curl -i http://localhost/.git/config
# Output: HTTP/1.1 403 Forbidden
```
