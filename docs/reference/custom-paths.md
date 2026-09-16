# Custom Path Configuration & Regex Guide

RouteWarden provides a flexible regular expression matching engine allowing you to define custom blocking rules (`pathPatterns` / `blockPatterns`) and safe overrides (`allowPatterns`).

---

## 1. How Path Matching Works

Before regular expressions are evaluated, RouteWarden runs every request path through its **Anti-Evasion Engine**:
- Decodes layered percent-encoding (`%252e` ➔ `.`)
- Strips semicolon matrix parameters (`/;param=1/admin` ➔ `/admin`)
- Normalizes Windows backslashes (`\admin` ➔ `/admin`)
- Cleans directory traversals (`/static/../admin` ➔ `/admin`)

Regex patterns are evaluated against the clean normalized path (and optionally the query string if `checkQuery: true`).

---

## 2. Defining Block Patterns (`pathPatterns`)

You can supply one or more regular expressions to block. `pathPatterns` and `blockPatterns` are interchangeable aliases.

### Syntax & Flags
RouteWarden uses Go's standard `regexp` syntax (RE2).
- **Case-Insensitive Flag**: Always prefix with `(?i)` unless you strictly require case sensitivity.
- **Root/Segment Anchoring**: Use `(^|/)` or `^/` to ensure you match full path segments rather than accidental substrings.

### Quick Pattern Cheat Sheet

| Defense Goal | Recommended Regex | Example Blocked URLs |
|---|---|---|
| **Environment Files** | `(?i)(^|/)(\.env.*)$` | `/.env`, `/.env.production`, `/.env.local` |
| **Source Control (Git/SVN)** | `(?i)(^|/)\.(git|svn|hg)(/.*\|$)` | `/.git/config`, `/.git/HEAD`, `/.svn/entries` |
| **Cloud & SSH Keys** | `(?i)(^|/)\.(aws|ssh|kube|docker)(/.*\|$)` | `/.aws/credentials`, `/.ssh/id_rsa`, `/.kube/config` |
| **Database Dumps** | `(?i).*\.(sql|dump|sqlite3?|db|rdb)$` | `/backup.sql`, `/data.dump`, `/users.sqlite` |
| **Archives & Backups** | `(?i).*\.(tar|tar\.gz|tgz|zip|rar|7z|bak)$` | `/site.zip`, `/db.backup`, `/code.tar.gz` |
| **Configurations** | `(?i).*\.(conf|config|ini|yaml|yml)$` | `/app.conf`, `/server.ini`, `/config.yaml` |
| **Application Logs** | `(?i).*\.(log|txt)$` *(pair with allowPatterns)* | `/error.log`, `/debug.txt`, `/access.log` |
| **PHP & CGI Exploits** | `(?i).*\.(php[0-9]?|phtml|cgi|asp|aspx|jsp)$` | `/index.php`, `/upload.phtml`, `/test.cgi` |
| **Internal / Admin APIs** | `(?i)^/api/(internal|admin|private)(/.*)?$` | `/api/internal/users`, `/api/admin/delete` |
| **Actuator & Metrics** | `(?i)^/(actuator|metrics|heapdump|env)(/.*)?$` | `/actuator/health`, `/metrics`, `/heapdump` |
| **Debug & Server Status** | `(?i)(^|/)(phpinfo|server-status|server-info)` | `/phpinfo.php`, `/server-status` |
| **Swagger / API Docs** | `(?i)^/(swagger|swagger-ui|api-docs)(/.*)?$` | `/swagger-ui.html`, `/v2/api-docs` |
| **Node / Python Locks** | `(?i)(^|/)(package-lock\.json|yarn\.lock|Pipfile)$` | `/package-lock.json`, `/yarn.lock` |

---

### Common Recipe Cookbooks

#### Recipe 1: Modern SPA / React / Vue / Next.js Shield
Blocks reconnaissance of server-side artifacts while permitting normal frontend routing:
```yaml
pathPatterns:
  - '(?i)(^|/)(\.env.*|\.git.*|\.aws.*)$'
  - '(?i).*\.(sql|bak|backup|conf|ini|yaml|yml|log)$'
  - '(?i)(^|/)(next\.config\.js|tsconfig\.json|package\.json|package-lock\.json)$'
```

#### Recipe 2: Python / Django / FastAPI Shield
Prevents exposure of virtualenvs, SQLite databases, and test artifacts:
```yaml
pathPatterns:
  - '(?i)(^|/)(__pycache__|\.pytest_cache|\.venv|venv)(/.*)?$'
  - '(?i).*\.(pyc|pyd|sqlite3?|db|log)$'
  - '(?i)(^|/)(requirements\.txt|Pipfile.*|poetry\.lock)$'
```

#### Recipe 3: PHP / WordPress / CMS Hardening
Stops brute-forcing and common scanning bots:
```yaml
pathPatterns:
  - '(?i)(^|/)(wp-login\.php|wp-admin|xmlrpc\.php|wp-config\.php)$'
  - '(?i)(^|/)(phpmyadmin|pma|adminer\.php|info\.php|phpinfo\.php)$'
  - '(?i).*\.(php|phtml|php3|php4|php5|phps|cgi)$'
```

#### Recipe 4: Java / Spring Boot Microservice Shield
Protects Spring actuator management ports and memory dumps:
```yaml
pathPatterns:
  - '(?i)^/(actuator|metrics|heapdump|trace|env|prometheus)(/.*)?$'
  - '(?i)^/(h2-console|swagger-ui.*|v[23]/api-docs)(/.*)?$'
```

#### Recipe 5: Microservice Internal API Isolation
Restricts internal endpoints from being reached via public ingress:
```yaml
pathPatterns:
  - '(?i)^/api/(internal|admin|management|debug)(/.*)?$'
```

---

## 3. Predefined Sample Application Blueprints

Below are complete, production-tested RouteWarden configurations designed for specific popular application stacks:

### 🌟 Blueprint A: WordPress / WooCommerce Store
Stops XML-RPC amplification attacks, wp-config exposure, and brute-force bot scans on wp-login:

::: code-group

```yaml [File (YAML)]
# dynamic_conf.yml
http:
  middlewares:
    wp-warden:
      plugin:
        routewarden:
          enabled: true
          enableDefaultPatterns: true
          pathPatterns:
            - '(?i)(^|/)(xmlrpc\.php|wp-config\.php|install\.php|license\.txt|readme\.html)$'
          allowPatterns:
            - '(?i)^/wp-content/uploads/.*'
            - '(?i)^/robots\.txt$'
          allowedIps:
            - "203.0.113.50"
          response:
            mode: text
            statusCode: 404
            body: "404 Not Found"

  routers:
    wp-router:
      rule: "Host(`shop.example.com`)"
      entryPoints:
        - web
      middlewares:
        - wp-warden
      service: wp-service
```

```toml [File (TOML)]
# dynamic_conf.toml
[http.routers.wp-router]
  rule = "Host(`shop.example.com`)"
  entryPoints = ["web"]
  middlewares = ["wp-warden"]
  service = "wp-service"

[http.middlewares.wp-warden.plugin.routewarden]
  enabled = true
  enableDefaultPatterns = true
  pathPatterns = ["(?i)(^|/)(xmlrpc\\.php|wp-config\\.php|install\\.php|license\\.txt|readme\\.html)$"]
  allowPatterns = ["(?i)^/wp-content/uploads/.*", "(?i)^/robots\\.txt$"]
  allowedIps = ["203.0.113.50"]

[http.middlewares.wp-warden.plugin.routewarden.response]
  mode = "text"
  statusCode = 404
  body = "404 Not Found"
```

```bash [CLI]
# Docker Compose Labels / CLI equivalent
- "traefik.enable=true"
- "traefik.http.routers.wp.rule=Host(`shop.example.com`)"
- "traefik.http.routers.wp.middlewares=wp-warden"
- "traefik.http.middlewares.wp-warden.plugin.routewarden.enabled=true"
- "traefik.http.middlewares.wp-warden.plugin.routewarden.enableDefaultPatterns=true"
- "traefik.http.middlewares.wp-warden.plugin.routewarden.pathPatterns=(?i)(^|/)(xmlrpc\\.php|wp-config\\.php|install\\.php|license\\.txt|readme\\.html)$"
- "traefik.http.middlewares.wp-warden.plugin.routewarden.allowPatterns=(?i)^/wp-content/uploads/.*,(?i)^/robots\\.txt$"
- "traefik.http.middlewares.wp-warden.plugin.routewarden.allowedIps=203.0.113.50"
- "traefik.http.middlewares.wp-warden.plugin.routewarden.response.mode=text"
- "traefik.http.middlewares.wp-warden.plugin.routewarden.response.statusCode=404"
- "traefik.http.middlewares.wp-warden.plugin.routewarden.response.body=404 Not Found"
```

:::

### 🌟 Blueprint B: Next.js / React / SvelteKit Full-Stack App
Protects internal server assets, environment secrets, and build manifests:

```yaml
services:
  nextjs-app:
    image: my-nextjs-app:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.nextjs.rule=Host(`app.example.com`)"
      - "traefik.http.routers.nextjs.middlewares=nextjs-warden"

      - "traefik.http.middlewares.nextjs-warden.plugin.routewarden.enabled=true"
      - "traefik.http.middlewares.nextjs-warden.plugin.routewarden.enableDefaultPatterns=true"
      # Block build configs, package locks, and server logs
      - "traefik.http.middlewares.nextjs-warden.plugin.routewarden.pathPatterns=(?i)(^|/)(next\\.config\\.js|tsconfig\\.json|package\\.json|package-lock\\.json|yarn\\.lock)$"
      # Allow static chunks and images
      - "traefik.http.middlewares.nextjs-warden.plugin.routewarden.allowPatterns=(?i)^/_next/static/.*,(?i)^/favicon\\.ico$"
      - "traefik.http.middlewares.nextjs-warden.plugin.routewarden.response.mode=json"
      - "traefik.http.middlewares.nextjs-warden.plugin.routewarden.response.statusCode=403"
      - "traefik.http.middlewares.nextjs-warden.plugin.routewarden.response.body={\"error\":\"Forbidden\"}"
```

### 🌟 Blueprint C: Python / Django / FastAPI Backend
Guards virtual environment directories, SQLite database files, and Django management endpoints:

```yaml
services:
  django-api:
    image: my-django-app:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.django.rule=Host(`api.example.com`)"
      - "traefik.http.routers.django.middlewares=django-warden"

      - "traefik.http.middlewares.django-warden.plugin.routewarden.enabled=true"
      - "traefik.http.middlewares.django-warden.plugin.routewarden.enableDefaultPatterns=true"
      # Block Python byte-code, virtualenvs, SQLite dumps, and settings files
      - "traefik.http.middlewares.django-warden.plugin.routewarden.pathPatterns=(?i)(^|/)(__pycache__|\\.venv|venv|local_settings\\.py|manage\\.py)$"
      # Exempt public static files & media
      - "traefik.http.middlewares.django-warden.plugin.routewarden.allowPatterns=(?i)^/static/.*,(?i)^/media/.*"
      # Office VPN bypass
      - "traefik.http.middlewares.django-warden.plugin.routewarden.allowedIps=10.0.0.0/8"
      - "traefik.http.middlewares.django-warden.plugin.routewarden.response.mode=json"
      - "traefik.http.middlewares.django-warden.plugin.routewarden.response.statusCode=403"
```

### 🌟 Blueprint D: Spring Boot / Java Cloud Microservice
Shields internal Actuator management metrics, trace dumps, and H2 database consoles:

```yaml
services:
  spring-service:
    image: my-spring-app:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.spring.rule=Host(`service.internal.example.com`)"
      - "traefik.http.routers.spring.middlewares=spring-warden"

      - "traefik.http.middlewares.spring-warden.plugin.routewarden.enabled=true"
      - "traefik.http.middlewares.spring-warden.plugin.routewarden.enableDefaultPatterns=true"
      # Block Spring debug consoles, heapdumps, and environment variables
      - "traefik.http.middlewares.spring-warden.plugin.routewarden.pathPatterns=(?i)^/(actuator|metrics|heapdump|trace|env|h2-console)(/.*)?$"
      # Exempt only the public liveness health check
      - "traefik.http.middlewares.spring-warden.plugin.routewarden.allowPatterns=(?i)^/actuator/health$"
      # Response
      - "traefik.http.middlewares.spring-warden.plugin.routewarden.response.mode=json"
      - "traefik.http.middlewares.spring-warden.plugin.routewarden.response.statusCode=403"
      - "traefik.http.middlewares.spring-warden.plugin.routewarden.response.body={\"error\":\"Forbidden\",\"scope\":\"actuator-protected\"}"
```

### 🌟 Blueprint E: PHP / Laravel Application
Protects `.env`, Artisan CLI files, storage logs, and debug toolbars:

```yaml
services:
  laravel-app:
    image: my-laravel-app:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.laravel.rule=Host(`laravel.example.com`)"
      - "traefik.http.routers.laravel.middlewares=laravel-warden"

      - "traefik.http.middlewares.laravel-warden.plugin.routewarden.enabled=true"
      - "traefik.http.middlewares.laravel-warden.plugin.routewarden.enableDefaultPatterns=true"
      # Block artisan, composer files, and storage logs
      - "traefik.http.middlewares.laravel-warden.plugin.routewarden.pathPatterns=(?i)(^|/)(artisan|composer\\.json|composer\\.lock|package\\.json|\\.env.*)$"
      # Allow public compiled assets
      - "traefik.http.middlewares.laravel-warden.plugin.routewarden.allowPatterns=(?i)^/(css|js|images|storage)/.*"
      - "traefik.http.middlewares.laravel-warden.plugin.routewarden.response.mode=text"
      - "traefik.http.middlewares.laravel-warden.plugin.routewarden.response.statusCode=404"
      - "traefik.http.middlewares.laravel-warden.plugin.routewarden.response.body=404 page not found"
```



---

## 3. Allowing Safe Endpoints (`allowPatterns`)

The `allowPatterns` list takes precedence over both built-in default patterns and your custom `pathPatterns`. If a path matches **any** regex in `allowPatterns`, RouteWarden immediately permits the request to pass downstream.

### Default Built-in Allow Rules
By default, RouteWarden automatically whitelists:
```regex
(?i)^/robots\.txt$
(?i)^/sitemap.*\.xml$
(?i)^/ads\.txt$
(?i)^/security\.txt$
(?i)^/\.well-known(/.*)?$
```

### Adding Custom Exceptions
For example, if you block all `*.yaml` files or `/api/*`, but need to allow a public spec file or public health check:

```yaml
pathPatterns:
  - '(?i).*\.ya?ml$'
  - '(?i)^/api/(internal|admin).*'

allowPatterns:
  # Allow public OpenAPI spec despite .yaml block
  - '(?i)^/api/v1/openapi\.ya?ml$'
  # Allow specific public health endpoint
  - '(?i)^/api/internal/health$'
```

---

## 4. Query String Inspection (`checkQuery`)

By default (`checkQuery: false`), RouteWarden inspects only the URL path. If attackers attempt to smuggle sensitive files via query parameters (e.g. `?file=../../.env` or `?redirect=phpinfo.php`), enable `checkQuery`:

```yaml
checkQuery: true
pathPatterns:
  - '(?i)(\.env|phpinfo|backup\.sql)'
```

---

## 5. Docker Compose Configuration Example

```yaml
services:
  webapp:
    image: nginx:alpine
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.webapp.rule=Host(`app.example.com`)"
      - "traefik.http.routers.webapp.middlewares=custom-shield"

      # RouteWarden Middleware Definition
      - "traefik.http.middlewares.custom-shield.plugin.routewarden.enabled=true"
      # Block custom internal endpoints
      - "traefik.http.middlewares.custom-shield.plugin.routewarden.pathPatterns=(?i)^/api/(internal|admin)(/.*)?$,(?i).*\.(sql|dump)$"
      # Exempt public health check
      - "traefik.http.middlewares.custom-shield.plugin.routewarden.allowPatterns=(?i)^/api/internal/health$"
      # Response
      - "traefik.http.middlewares.custom-shield.plugin.routewarden.response.mode=json"
      - "traefik.http.middlewares.custom-shield.plugin.routewarden.response.statusCode=403"
      - "traefik.http.middlewares.custom-shield.plugin.routewarden.response.body={\"error\":\"Forbidden\",\"message\":\"Restricted path pattern\"}"
```
