# Examples & Cookbook Overview

Browse ready-to-run configurations and production blueprints for RouteWarden across different deployment targets.

---

## Scenario Index

| Scenario | Description | Target |
|---|---|---|
| [1. Basic Sensitive Files](/examples/basic-sensitive-files) | Shield `.env`, `.git`, backups, and configs with custom JSON errors. | Docker Compose / Traefik |
| [2. Global EntryPoint Shield](/examples/docker-compose-global) | Protect all microservices and routes automatically at the Traefik entrypoint. | Docker Compose |
| [3. Service-Level Docker Compose](/examples/docker-compose-service) | Tailor rules, custom regexes, and safe allowlists per service. | Docker Compose |
| [4. IP / Subnet Whitelisting](/examples/ip-whitelisting) | Allow internal corporate VPNs, office IPs, and developer subnets. | Docker Compose / Traefik |
| [5. Captcha Verification Challenge](/examples/captcha) | Challenge clients via Cloudflare Turnstile or hCaptcha on sensitive routes. | Docker Compose / Traefik |
| [6. Kubernetes IngressRoute](/examples/kubernetes) | Production IngressRoute and Middleware CRD setup for Traefik Kubernetes. | Kubernetes CRD |

---

## Production Case Studies

| Case Study | Focus & Threat Model | Protection Pattern |
|---|---|---|
| [1. Immich Dual-Router Shield](/examples/case-study-immich) | Public photo/video sharing without exposing login/admin APIs. | Dual Traefik Routers + 404 Masking |
| [2. Zero-Trust Webhooks](/examples/case-study-webhooks) | Secure Stripe/GitHub webhook ingress from unauthorized HTTP injection. | Provider CIDR Whitelisting + Silent Drop |
| [3. Observability & Metrics Cloaking](/examples/case-study-observability) | Prevent public harvesting of Prometheus `/metrics` and `/actuator`. | VPC / Internal Scraper IP Exemption |
| [4. CMS & WordPress Brute-Force Shield](/examples/case-study-cms-shield) | Eliminate credential-stuffing on `wp-login.php`, `xmlrpc.php`, and `/admin`. | Cloudflare Turnstile / hCaptcha Challenge |
| [5. Password Vaults (Vaultwarden)](/examples/case-study-vaultwarden) | Public mobile password sync while restricting `/admin` to Tailscale/WireGuard. | VPN Subnet Filter + 404 Error Cloaking |
| [6. Honeypots & Staging Cloaking](/examples/case-study-honeypot-staging) | Neutralize scanning bots and hide pull-request preview clusters from crawlers. | TCP RST (`silentDrop`) & 307 Deflection |

---

## In-Repo Runnable Code

All examples are checked directly into the [`examples/`](https://github.com/routewarden/traefik-warden/tree/main/examples) directory of the RouteWarden GitHub repository. You can clone the repo and run any scenario in seconds:

```bash
git clone https://github.com/routewarden/traefik-warden.git
cd routewarden/examples/01-basic-sensitive-files
docker compose up -d
```
