# Local Development & Deployment

This guide explains how to develop, test, and run RouteWarden locally as a Traefik plugin without publishing to GitHub or the Traefik Plugin Catalog.

---

## 1. How Traefik Local Plugins Work

Traefik allows loading plugins directly from a local directory on your filesystem using the `experimental.localPlugins` configuration key.

When using `localPlugins`, Traefik requires the source code to be mounted in a specific directory structure matching Go's module namespace:

```text
plugins-local/
└── src/
    └── github.com/
        └── routewarden/
            └── traefik-warden/
                ├── .traefik.yml
                ├── config.go
                ├── ip_filter.go
                ├── path_normalizer.go
                ├── response_handler.go
                └── routewarden.go
```

---

## 2. Local Traefik Configuration Preview

::: code-group

```yaml [File (YAML)]
# traefik.yml (Static)
experimental:
  localPlugins:
    routewarden:
      moduleName: github.com/routewarden/traefik-warden

# dynamic_conf.yml (Dynamic Middleware & Router)
http:
  middlewares:
    local-warden:
      plugin:
        routewarden:
          enabled: true
          enableDefaultPatterns: true
          response:
            mode: json
            statusCode: 403
            body: '{"error":"Forbidden","environment":"local-dev"}'

  routers:
    app-router:
      rule: "Host(`localhost`)"
      entryPoints:
        - web
      middlewares:
        - local-warden
      service: app-service
```

```toml [File (TOML)]
# traefik.toml (Static)
[experimental.localPlugins.routewarden]
  moduleName = "github.com/routewarden/traefik-warden"

# dynamic_conf.toml (Dynamic Middleware & Router)
[http.routers.app-router]
  rule = "Host(`localhost`)"
  entryPoints = ["web"]
  middlewares = ["local-warden"]
  service = "app-service"

[http.middlewares.local-warden.plugin.routewarden]
  enabled = true
  enableDefaultPatterns = true

[http.middlewares.local-warden.plugin.routewarden.response]
  mode = "json"
  statusCode = 403
  body = '{"error":"Forbidden","environment":"local-dev"}'
```

```bash [CLI]
# Traefik CLI flags
traefik \
  --api.insecure=true \
  --providers.docker=true \
  --entrypoints.web.address=:80 \
  --experimental.localplugins.routewarden.modulename=github.com/routewarden/traefik-warden \
  --log.level=DEBUG
```

:::

---

## 3. Local Docker Compose Setup

Here is a complete, ready-to-run `docker-compose.yml` for developing and testing RouteWarden locally:

```yaml
services:
  traefik:
    image: traefik:v3.1
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      
      # Declare RouteWarden as a LOCAL plugin:
      - "--experimental.localPlugins.routewarden.modulename=github.com/routewarden/traefik-warden"
      
      # Log level debug helps verify plugin loading
      - "--log.level=DEBUG"
    ports:
      - "80:80"
      - "8080:8080" # Traefik Web UI & Middleware Explorer
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
      # Mount your local repository directory into Traefik's plugins-local path:
      - ".:/plugins-local/src/github.com/routewarden/traefik-warden:ro"

  webapp:
    image: nginx:alpine
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.webapp.rule=Host(`localhost`)"
      - "traefik.http.routers.webapp.entrypoints=web"
      - "traefik.http.routers.webapp.middlewares=local-warden"

      # Middleware configuration
      - "traefik.http.middlewares.local-warden.plugin.routewarden.enabled=true"
      - "traefik.http.middlewares.local-warden.plugin.routewarden.enableDefaultPatterns=true"
      - "traefik.http.middlewares.local-warden.plugin.routewarden.response.mode=json"
      - "traefik.http.middlewares.local-warden.plugin.routewarden.response.statusCode=403"
      - "traefik.http.middlewares.local-warden.plugin.routewarden.response.body={\"error\":\"Forbidden\",\"environment\":\"local-dev\"}"
```

---

## 4. Step-by-Step Local Walkthrough

### Step 1: Start the Cluster
From the root of the RouteWarden repository:

```bash
docker compose up
```

### Step 2: Verify Plugin Initialization
Watch the Traefik startup logs. You should see Traefik's Yaegi interpreter successfully compiling the local plugin:

```text
level=info msg="Loading plugin: routewarden with module: github.com/routewarden/traefik-warden"
level=info msg="Plugin routewarden loaded successfully"
```

### Step 3: Inspect via Traefik Dashboard
Open your browser to: `http://localhost:8080/dashboard/#/http/middlewares`

You will see `local-warden@docker` listed with its active configuration.

### Step 4: Test Blocking Live
```bash
# 1. Test normal request (Allowed)
curl -I http://localhost/
# Output: HTTP/1.1 200 OK

# 2. Test sensitive .env file (Blocked)
curl -i http://localhost/.env
# Output: HTTP/1.1 403 Forbidden
# {"status":403,"message":"Blocked by local RouteWarden plugin"}

# 3. Test URL-encoded evasion (%2eenv) (Blocked)
curl -i "http://localhost/%2eenv"
# Output: HTTP/1.1 403 Forbidden
```

### Step 5: Live Code Iteration
Because the repository root is mounted with `- .:/plugins-local/src/github.com/routewarden/traefik-warden:ro`, whenever you modify Go code, simply restart Traefik to recompile:

```bash
docker compose restart traefik
```
*(No Docker rebuild or external plugin download required!)*
