# Example 4: IP & CIDR Subnet Whitelisting

RouteWarden allows you to declare trusted IPs and subnets (`allowedIps`) to bypass path blocking. This is ideal for internal management portals, company VPN gateways, and authorized vulnerability scanners.

---

## How IP Resolution Works

RouteWarden evaluates client IPs in the following priority order:
1. **`X-Forwarded-For`** header (first IP in list)
2. **`X-Real-IP`** header
3. Socket **`RemoteAddr`**

Both exact IPv4/IPv6 addresses (`127.0.0.1`, `2001:db8::1`) and CIDR blocks (`10.0.0.0/8`, `192.168.1.0/24`) are supported.

---

## Docker Compose Configuration

```yaml
services:
  traefik:
    image: traefik:v3.1
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
      - "--experimental.plugins.routewarden.modulename=github.com/aman400/routewarden"
      - "--experimental.plugins.routewarden.version={{version}}"
    ports:
      - "80:80"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"

  admin-service:
    image: nginx:alpine
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.admin.rule=Host(`admin.localhost`)"
      - "traefik.http.routers.admin.entrypoints=web"
      - "traefik.http.routers.admin.middlewares=admin-shield"

      # RouteWarden Configuration with IP Whitelist
      - "traefik.http.middlewares.admin-shield.plugin.routewarden.enabled=true"
      - "traefik.http.middlewares.admin-shield.plugin.routewarden.pathPatterns=(?i)^/admin(/.*)?$,(?i)^/metrics(/.*)?$"
      - "traefik.http.middlewares.admin-shield.plugin.routewarden.allowedIps=10.0.0.0/8,192.168.1.100"
      - "traefik.http.middlewares.admin-shield.plugin.routewarden.response.mode=json"
      - "traefik.http.middlewares.admin-shield.plugin.routewarden.response.statusCode=403"
      - "traefik.http.middlewares.admin-shield.plugin.routewarden.response.body={\"error\":\"Forbidden\",\"message\":\"Restricted to authorized IP/VPN\"}"
```

---

## Testing Verification

```bash
# Request without whitelisted IP (Blocked)
curl -i -H "Host: admin.localhost" http://localhost/admin
# HTTP/1.1 403 Forbidden

# Request originating from allowed corporate subnet via proxy (Allowed)
curl -i -H "Host: admin.localhost" -H "X-Forwarded-For: 10.5.20.1" http://localhost/admin
# Passes cleanly to upstream container!
```
