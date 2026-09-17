# Case Study: Self-Hosted Cloud & Vault Protection (Vaultwarden & Nextcloud)

This case study demonstrates how to secure critical self-hosted services like **Vaultwarden** (Bitwarden-compatible password manager) and **Nextcloud** without interfering with day-to-day mobile synchronization or public file sharing.

---

## The Threat Model: Vaultwarden

Vaultwarden is one of the most popular self-hosted password managers. To allow mobile apps and browser extensions to sync passwords, the `/api/*` and `/identity/*` endpoints must be reachable over the internet.

However, the `/admin` portal (which allows creating/deleting accounts, viewing server logs, and changing master secrets) represents an existential security risk if probed by external attackers.

### The Objective
1. **Public Traffic**: Can reach `/api/*`, `/identity/*`, and the web vault for ordinary password synchronization.
2. **Admin Portal (`/admin`)**: Fully locked down and invisible (returning `404 Not Found`) unless the connection originates from a trusted internal VPN subnet (e.g., Tailscale `100.64.0.0/10` or WireGuard).

---

## Traefik Configuration

::: code-group

```yaml [File (YAML)]
# dynamic_conf.yml
http:
  middlewares:
    vaultwarden-shield:
      plugin:
        routewarden:
          enabled: true
          enableDefaultPatterns: true
          # Intercept the administrative console
          pathPatterns:
            - '(?i)^/admin(/.*)?$'
          # Allow ONLY internal WireGuard & Tailscale VPN addresses
          allowedIps:
            - "100.64.0.0/10"   # Tailscale CGNAT range
            - "10.8.0.0/24"     # WireGuard VPN subnet
            - "127.0.0.1"       # Localhost
          response:
            mode: json
            statusCode: 404
            body: '{"error":"Not Found","message":"The requested resource was not found"}'

  routers:
    vault-router:
      rule: "Host(`vault.example.com`)"
      entryPoints:
        - websecure
      middlewares:
        - vaultwarden-shield
      service: vault-service
```

```toml [File (TOML)]
# dynamic_conf.toml
[http.routers.vault-router]
  rule = "Host(`vault.example.com`)"
  entryPoints = ["websecure"]
  middlewares = ["vaultwarden-shield"]
  service = "vault-service"

[http.middlewares.vaultwarden-shield.plugin.routewarden]
  enabled = true
  enableDefaultPatterns = true
  pathPatterns = ["(?i)^/admin(/.*)?$"]
  allowedIps = ["100.64.0.0/10", "10.8.0.0/24", "127.0.0.1"]

[http.middlewares.vaultwarden-shield.plugin.routewarden.response]
  mode = "json"
  statusCode = 404
  body = '{"error":"Not Found","message":"The requested resource was not found"}'
```

```bash [CLI]
# Docker Compose Labels equivalent
- "traefik.http.routers.vault.rule=Host(`vault.example.com`)"
- "traefik.http.routers.vault.middlewares=vaultwarden-shield"
- "traefik.http.middlewares.vaultwarden-shield.plugin.routewarden.enabled=true"
- "traefik.http.middlewares.vaultwarden-shield.plugin.routewarden.pathPatterns=(?i)^/admin(/.*)?$"
- "traefik.http.middlewares.vaultwarden-shield.plugin.routewarden.allowedIps=100.64.0.0/10,10.8.0.0/24,127.0.0.1"
- "traefik.http.middlewares.vaultwarden-shield.plugin.routewarden.response.mode=json"
- "traefik.http.middlewares.vaultwarden-shield.plugin.routewarden.response.statusCode=404"
```

:::
