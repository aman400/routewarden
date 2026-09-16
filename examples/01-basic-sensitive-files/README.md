# Example 01: Basic Sensitive File Blocking

This example demonstrates using RouteWarden to protect a backend web service from exposure of sensitive files (such as `.env`, `.git`, `.aws`, `.sql`, `.bak`, `.log`, `.conf`, and lockfiles).

## Architecture

```
Internet Request ───► Traefik (:80) ───► RouteWarden Plugin ───► Nginx Backend
                                                │
                                       (Sensitive file?)
                                       ├── Yes ──► HTTP 403 (Blocked JSON)
                                       └── No  ──► Passes to Backend
```

## Running the Example

```bash
docker compose up -d
```

Open Traefik Dashboard at [http://localhost:8080](http://localhost:8080).

## Verification

### 1. Legitimate Traffic (Allowed)
```bash
curl -I http://localhost/
# Expected: HTTP/1.1 200 OK
```

### 2. Sensitive `.env` Probe (Blocked)
```bash
curl -i http://localhost/.env
# Expected: HTTP/1.1 403 Forbidden
# {"error":"Forbidden","message":"Sensitive path blocked by RouteWarden"}
```

### 3. Git Repository Probe (Blocked)
```bash
curl -i http://localhost/.git/config
# Expected: HTTP/1.1 403 Forbidden
```

### 4. Database Dump Probe (Blocked)
```bash
curl -i http://localhost/backup.sql
# Expected: HTTP/1.1 403 Forbidden
```

### 5. URL-Encoded Evasion Attempt (Blocked)
```bash
curl -i "http://localhost/%2eenv"
# Expected: HTTP/1.1 403 Forbidden
```
