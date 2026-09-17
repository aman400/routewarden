# Case Study: Zero-Trust Webhook Ingress (Stripe & GitHub)

This case study demonstrates how to protect payment gateways and webhook ingestion endpoints from unauthorized tampering, forged payload injection, and internal route disclosure.

---

## The Threat Model

Modern SaaS and e-commerce applications rely heavily on external webhooks from services like **Stripe**, **GitHub**, **Shopify**, **Paddle**, and **Slack**.

Because webhook handlers must be reachable over the public internet, developers frequently place their entire payment or integration microservice behind Traefik. This exposes several risks:
- **Reconnaissance of Internal Endpoints**: Attackers probe `/webhooks/debug`, `/webhooks/replay`, or `/webhooks/test`.
- **Forged Event Injection**: Direct HTTP spam to the webhook receiver attempting to trigger expensive verification workflows or crash background queues.
- **Credential & Config Exposure**: Accidental leakage of `.env` or deployment configurations on webhook servers.

---

## The Solution: Two-Layer Zero-Trust Defense

With RouteWarden, you enforce a two-tier gatekeeper at the Traefik edge:
1. **IP Range Restriction (`allowedIps`)**: Only official CIDR blocks published by Stripe/GitHub can send payloads to `/webhooks/.*`.
2. **Exact Allowlist (`allowPatterns`)**: Only designated production webhook endpoints (e.g. `/webhooks/stripe/v1`) are permitted; all internal or debug paths are blocked with `404 Not Found` or dropped silently (`silentDrop`).

---

## Traefik Configuration

::: code-group

```yaml [File (YAML)]
# dynamic_conf.yml
http:
  middlewares:
    webhook-shield:
      plugin:
        routewarden:
          enabled: true
          enableDefaultPatterns: true
          # Block everything under /webhooks by default
          pathPatterns:
            - '(?i)^/webhooks(/.*)?$'
          # Allow ONLY the verified production webhook handler
          allowPatterns:
            - '(?i)^/webhooks/stripe/v1$'
          # Restrict to official Stripe Webhook IP ranges
          allowedIps:
            - "3.18.12.63/32"
            - "3.130.192.231/32"
            - "13.235.14.237/32"
            - "13.235.122.149/32"
            - "35.154.171.200/32"
          response:
            mode: silentDrop # Drop unauthorized scanner connections immediately

  routers:
    webhook-router:
      rule: "Host(`api.example.com`) && PathPrefix(`/webhooks`)"
      entryPoints:
        - websecure
      middlewares:
        - webhook-shield
      service: payment-service
```

```toml [File (TOML)]
# dynamic_conf.toml
[http.routers.webhook-router]
  rule = "Host(`api.example.com`) && PathPrefix(`/webhooks`)"
  entryPoints = ["websecure"]
  middlewares = ["webhook-shield"]
  service = "payment-service"

[http.middlewares.webhook-shield.plugin.routewarden]
  enabled = true
  enableDefaultPatterns = true
  pathPatterns = ["(?i)^/webhooks(/.*)?$"]
  allowPatterns = ["(?i)^/webhooks/stripe/v1$"]
  allowedIps = [
    "3.18.12.63/32",
    "3.130.192.231/32",
    "13.235.14.237/32",
    "13.235.122.149/32",
    "35.154.171.200/32"
  ]

[http.middlewares.webhook-shield.plugin.routewarden.response]
  mode = "silentDrop"
```

```bash [CLI]
# Docker Compose Labels equivalent
- "traefik.http.routers.webhook.rule=Host(`api.example.com`) && PathPrefix(`/webhooks`)"
- "traefik.http.routers.webhook.middlewares=webhook-shield"
- "traefik.http.middlewares.webhook-shield.plugin.routewarden.enabled=true"
- "traefik.http.middlewares.webhook-shield.plugin.routewarden.pathPatterns=(?i)^/webhooks(/.*)?$"
- "traefik.http.middlewares.webhook-shield.plugin.routewarden.allowPatterns=(?i)^/webhooks/stripe/v1$"
- "traefik.http.middlewares.webhook-shield.plugin.routewarden.allowedIps=3.18.12.63/32,3.130.192.231/32,13.235.14.237/32"
- "traefik.http.middlewares.webhook-shield.plugin.routewarden.response.mode=silentDrop"
```

:::
