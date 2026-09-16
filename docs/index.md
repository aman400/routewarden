---
layout: home

hero:
  name: "RouteWarden"
  text: "High-Performance Traefik Middleware"
  tagline: "Ultra-fast sensitive path defense, anti-evasion normalization, IP whitelisting, and multi-action responses."
  image:
    src: /icon.svg
    alt: RouteWarden Logo
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: Examples & Wiki
      link: /examples/overview
    - theme: alt
      text: View on GitHub
      link: https://github.com/aman400/routewarden

features:
  - icon: 🛡️
    title: Zero-Config Sensitive File Blocking
    details: Automatically guards against unauthorized access to .env, .git, .aws, backups (.sql, .bak), config files (.yaml, .ini), logs, and debug endpoints.
  - icon: ⚡
    title: Advanced Anti-Evasion Engine
    details: Defeats multi-layer URL encoding, directory traversal (../), IIS backslashes (\), semicolon matrix parameters, and null byte injections.
  - icon: 🌐
    title: IP & CIDR Subnet Whitelisting
    details: Bypass blocking for corporate VPNs, office IPs, or developer subnets with support for X-Forwarded-For, X-Real-IP, and socket RemoteAddr.
  - icon: 🎭
    title: Multi-Mode Response Engine
    details: Custom JSON payloads, branded HTML 404/403 pages, Cloudflare Turnstile/hCaptcha verification challenges, URL redirects, or silent drops.
  - icon: 🚀
    title: Pure Go & Yaegi Native
    details: Zero third-party dependencies outside the Go standard library. 100% compliant with Traefik's Yaegi interpreter.
  - icon: 🐳
    title: Docker & Kubernetes Native
    details: Drop-in support for Traefik v2/v3, Docker Compose labels (global & per-service), and Kubernetes IngressRoute CRDs.
---

## Quick Look

Protecting your entire infrastructure with RouteWarden takes just a few labels in Docker Compose:

```yaml
services:
  traefik:
    image: traefik:v3.1
    command:
      - "--experimental.plugins.routewarden.modulename=github.com/aman400/routewarden"
      - "--experimental.plugins.routewarden.version={{version}}"
      - "--entrypoints.web.http.middlewares=global-warden@docker"
    labels:
      # Enable RouteWarden middleware (Default: true)
      - "traefik.http.middlewares.global-warden.plugin.routewarden.enabled=true"
      # Block built-in sensitive files: .env*, .git, .aws, .sql, .bak, etc. (Default: true)
      - "traefik.http.middlewares.global-warden.plugin.routewarden.enableDefaultPatterns=true"
      # (Optional) Additional custom regex patterns to block (Default: [])
      - "traefik.http.middlewares.global-warden.plugin.routewarden.pathPatterns=(?i)^/admin(/.*)?$,(?i)^/api/internal(/.*)?$"
      # (Optional) Safe exception overrides to always allow (Default: robots.txt, ads.txt, sitemap.xml, .well-known/*)
      - "traefik.http.middlewares.global-warden.plugin.routewarden.allowPatterns=(?i)^/api/internal/health$,(?i)^/robots\\.txt$"
      # (Optional) Trusted IP / CIDR subnet bypass (Default: [])
      - "traefik.http.middlewares.global-warden.plugin.routewarden.allowedIps=10.0.0.0/8"
      # Response mode: text, json, html, captcha, redirect (Default: text, StatusCode: 403)
      - "traefik.http.middlewares.global-warden.plugin.routewarden.response.mode=json"
```
