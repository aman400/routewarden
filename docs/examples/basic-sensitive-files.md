# Example 1: Basic Sensitive File Blocking

This scenario protects a web application against reconnaissance and exposure of critical infrastructure files using RouteWarden's built-in rule dictionary.

---

## Configuration Preview

::: code-group

```yaml [File (YAML)]
# dynamic_conf.yml
http:
  middlewares:
    warden-shield:
      plugin:
        routewarden:
          enabled: true
          enableDefaultPatterns: true
          response:
            mode: json
            statusCode: 403
            body: '{"error":"Forbidden","message":"Sensitive path blocked by RouteWarden"}'

  routers:
    webapp-router:
      rule: "Host(`localhost`)"
      entryPoints:
        - web
      middlewares:
        - warden-shield
      service: webapp-service

  services:
    webapp-service:
      loadBalancer:
        servers:
          - url: "http://webapp:80"
```

```toml [File (TOML)]
# dynamic_conf.toml
[http.routers.webapp-router]
  rule = "Host(`localhost`)"
  entryPoints = ["web"]
  middlewares = ["warden-shield"]
  service = "webapp-service"

[http.services.webapp-service.loadBalancer]
  [[http.services.webapp-service.loadBalancer.servers]]
    url = "http://webapp:80"

[http.middlewares.warden-shield.plugin.routewarden]
  enabled = true
  enableDefaultPatterns = true

[http.middlewares.warden-shield.plugin.routewarden.response]
  mode = "json"
  statusCode = 403
  body = '{"error":"Forbidden","message":"Sensitive path blocked by RouteWarden"}'
```

```bash [CLI]
# Traefik Docker Compose Labels / CLI equivalent
- "traefik.enable=true"
- "traefik.http.routers.webapp.rule=Host(`localhost`)"
- "traefik.http.routers.webapp.entrypoints=web"
- "traefik.http.routers.webapp.middlewares=warden-shield"
- "traefik.http.middlewares.warden-shield.plugin.routewarden.enabled=true"
- "traefik.http.middlewares.warden-shield.plugin.routewarden.enableDefaultPatterns=true"
- "traefik.http.middlewares.warden-shield.plugin.routewarden.response.mode=json"
- "traefik.http.middlewares.warden-shield.plugin.routewarden.response.statusCode=403"
- 'traefik.http.middlewares.warden-shield.plugin.routewarden.response.body={"error":"Forbidden","message":"Sensitive path blocked by RouteWarden"}'
```

:::

---

## Docker Compose Example

```yaml
services:
  traefik:
    image: traefik:v3.1
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--experimental.plugins.routewarden.modulename=github.com/routewarden/traefik-warden"
      - "--experimental.plugins.routewarden.version={{version}}"
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
