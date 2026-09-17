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
| [💡 Case Study: Dual-Router Shield (Immich)](/examples/case-study-immich) | Public photo/video sharing with blocked admin/login endpoints and private VPN router. | Production Architecture |

---

## In-Repo Runnable Code

All examples are checked directly into the [`examples/`](https://github.com/aman400/routewarden/tree/main/examples) directory of the RouteWarden GitHub repository. You can clone the repo and run any scenario in seconds:

```bash
git clone https://github.com/aman400/routewarden.git
cd routewarden/examples/01-basic-sensitive-files
docker compose up -d
```
