# RouteWarden Local Testing Playground

This directory provides an instant, self-contained local testing environment for RouteWarden with Traefik using `localPlugins`.

## Why `localPlugins`?
- **Zero integrity check errors**: Bypasses Traefik's public plugin catalog.
- **Instant reload**: Mounts your live local Go code directly into Traefik.
- **No network/rate-limits**: Works completely offline.

---

## Quick Start

### 1. Start Traefik & Sample Backend
```bash
cd samples
docker compose up -d
```

### 2. Follow Logs in Real Time
```bash
docker compose logs -f traefik
```
You will immediately see RouteWarden initialize:
```text
[DEBUG] routewarden [routewarden]: initialized (enabled=true, debug=true, securityLog=true, blockPatterns=12, allowPatterns=5, mode=fakeSuccess)
```

Whenever an attacker probes a blocked route (such as `/.env` or `/wp-login.php`), RouteWarden emits a structured JSON security audit log to stdout, ready for CrowdSec or SIEM parsers:
```json
{"action":"fakeSuccess","client_ip":"192.168.1.50","method":"GET","path":"/.env","pattern":"(?i)(^|/)(\\.env.*|.*\\.(txt|log|bak|backup|sql|conf|config|ini|yaml|yml))$","plugin":"routewarden","reason":"path_blocked","request_uri":"/.env","timestamp":"2026-09-19T16:15:00Z","type":"routewarden_block","user_agent":"curl/7.88.1"}
```

### 3. Run Automated Tests
```bash
./test.sh
```
This tests all 11 defensive modes and validates the emission of structured JSON security audit logs.


---

## Testing Different Response Modes

You can test any response mode by editing `docker-compose.yml` under `traefik.http.middlewares.routewarden.plugin.routewarden`:

### 1. Honeypot Decoy Mode (`fakeSuccess`)
```yaml
traefik.http.middlewares.routewarden.plugin.routewarden.mode: "fakeSuccess"
```
Test with:
```bash
curl http://localhost/.env
```
Returns `200 OK` with realistic fake Laravel/MySQL credentials.

---

### 2. Silent Drop Mode (`silentDrop`)
```yaml
traefik.http.middlewares.routewarden.plugin.routewarden.silentdrop: "true"
```
Test with:
```bash
curl -v http://localhost/.env
```
Abruptly closes the socket (`Empty reply from server` / stream reset).

---

### 3. JSON Error Mode
```yaml
traefik.http.middlewares.routewarden.plugin.routewarden.mode: "json"
traefik.http.middlewares.routewarden.plugin.routewarden.statuscode: "403"
```
Test with:
```bash
curl http://localhost/.env
```
Returns structured `{"error":"Forbidden","status":403,...}`.

---

### 4. Gzip Bomb Mode
```yaml
traefik.http.middlewares.routewarden.plugin.routewarden.mode: "gzipBomb"
```

---

## Teardown
```bash
docker compose down
```
