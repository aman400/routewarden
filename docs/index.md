---
layout: home

hero:
  name: "RouteWarden"
  text: "High-Performance Traefik Middleware"
  tagline: "Stop sensitive file leaks (.env, .git, backups), neutralize path-evasion attacks, whitelist IPs, and serve custom error/captcha responses before requests reach your backend."
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
      link: https://github.com/routewarden/traefik-warden

features:
  - icon: 🛡️
    title: Anti-Probing & Scanner Defense
    details: Automatically intercepts automated bots and vulnerability crawlers probing for exposed credentials, backups, and administrative endpoints.
  - icon: ⚡
    title: Advanced Anti-Evasion Engine
    details: Defeats multi-layer URL encoding, directory traversal (../), IIS backslashes (\), semicolon matrix parameters, and null byte injections.
  - icon: 🌐
    title: IP & CIDR Subnet Whitelisting
    details: Bypass blocking for corporate VPNs, office IPs, or developer subnets with support for X-Forwarded-For, X-Real-IP, and socket RemoteAddr.
  - icon: 🎭
    title: Multi-Mode Response Engine
    details: Custom JSON payloads, branded HTML 404/403 pages, Cloudflare Turnstile/hCaptcha verification challenges, URL redirects, silent drops, or active defense gzip bombs.
  - icon: 🚀
    title: Pure Go & Yaegi Native
    details: Zero third-party dependencies outside the Go standard library. 100% compliant with Traefik's Yaegi interpreter.
  - icon: 🐳
    title: Docker & Kubernetes Native
    details: Drop-in support for Traefik v2/v3, Docker Compose labels (global & per-service), and Kubernetes IngressRoute CRDs.
---

## What is RouteWarden?

**RouteWarden** is an ultra-fast, zero-dependency Traefik middleware plugin built in pure Go. It acts as an **in-line security shield** deployed at your edge router or ingress controller, safeguarding downstream microservices and web applications from **reconnaissance probing**, accidental sensitive data exposure, and path evasion attacks.

Every internet-connected IP is continuously bombarded by automated crawlers, Shodan/Censys scanners, and credential-harvesting bots searching for `.env` files, `.git` credential databases, database backups, admin consoles, and leaked cloud credentials. RouteWarden intercepts and neutralizes these probing attempts at the Traefik proxy layer **before they ever hit your upstream containers or touch your backend logs**.

---

## Request Inspection Lifecycle

RouteWarden evaluates every inbound HTTP request across four deterministic security stages:

<div class="home-pipeline">
  <div class="pipeline-step">
    <div class="pipeline-num">Stage 1</div>
    <div class="pipeline-title">IP / CIDR Whitelisting</div>
    <div class="pipeline-desc">Evaluates origin IP against <code>allowedIps</code> via socket <code>RemoteAddr</code>, <code>X-Forwarded-For</code>, or <code>X-Real-IP</code>. Trusted VPN/office IPs bypass checks immediately.</div>
  </div>
  <div class="pipeline-step">
    <div class="pipeline-num">Stage 2</div>
    <div class="pipeline-title">Anti-Evasion Normalization</div>
    <div class="pipeline-desc">Unescapes double URL encoding (<code>%252e%252e</code>), strips semicolon matrix parameters (<code>/;param/.env</code>), normalizes IIS backslashes (<code>\</code>), and scrubs null bytes.</div>
  </div>
  <div class="pipeline-step">
    <div class="pipeline-num">Stage 3</div>
    <div class="pipeline-title">Dual Match Engine</div>
    <div class="pipeline-desc">Matches against safe overrides (<code>allowPatterns</code>) before testing built-in sensitive dictionaries (<code>.env</code>, <code>.git</code>, backups, configs) and custom <code>pathPatterns</code>.</div>
  </div>
  <div class="pipeline-step">
    <div class="pipeline-num">Stage 4</div>
    <div class="pipeline-title">Multi-Action Response</div>
    <div class="pipeline-desc">Emits custom JSON, branded HTML, redirect honeypots, silent TCP drops, interactive <b>Cloudflare Turnstile</b> / <b>hCaptcha</b> challenges, or active defense <b>Gzip Bombs</b>.</div>
  </div>
</div>

---

## Common Attack Vectors Blocked Out-of-the-Box

Without requiring any custom regex rules, RouteWarden's `enableDefaultPatterns: true` guards against the most critical OWASP information disclosure vulnerabilities:

<div class="attack-grid">
  <div class="attack-card">
    <h4>🔐 Environment & Secrets</h4>
    <p>Blocks <code>/.env</code>, <code>/.env.local</code>, <code>/.env.production</code>, <code>/.aws/credentials</code>, and <code>/.ssh/id_rsa</code>.</p>
  </div>
  <div class="attack-card">
    <h4>📁 Version Control Repos</h4>
    <p>Prevents source code leakage via <code>/.git/config</code>, <code>/.git/HEAD</code>, <code>/.svn/entries</code>, and <code>/.hg/</code>.</p>
  </div>
  <div class="attack-card">
    <h4>💾 Database Dumps & Backups</h4>
    <p>Catches accidental exposure of <code>/dump.sql</code>, <code>/db.bak</code>, <code>/backup.tar.gz</code>, and <code>/site.zip</code>.</p>
  </div>
  <div class="attack-card">
    <h4>⚙️ Configuration & Diagnostics</h4>
    <p>Guards server manifests like <code>/config.yaml</code>, <code>/app.ini</code>, <code>/phpinfo.php</code>, and Spring <code>/actuator/*</code>.</p>
  </div>
</div>

---

## Anti-Evasion Capabilities

Attackers frequently encode paths or use reverse-proxy edge cases to evade simple string-matching rules. RouteWarden eliminates these evasion vectors before matching:

| Evasion Technique | Raw Attacker Payload | RouteWarden Normalized Candidate | Protection Action |
|---|---|---|---|
| **Double URL Encoding** | `/%252e%252e/%252eenv` | `/.env` | 🛡️ **Blocked** |
| **Semicolon Matrix Traversal** | `/public;param=1/..;param=2/.env` | `/.env` | 🛡️ **Blocked** |
| **Windows / IIS Backslash** | `/static\..\.git\config` | `/.git/config` | 🛡️ **Blocked** |
| **Null Byte Injection** | `/.env%00.png` | `/.env` | 🛡️ **Blocked** |
| **Dot-Segment Traversal** | `/images/../.aws/credentials` | `/.aws/credentials` | 🛡️ **Blocked** |

---

## Flexible Response Engines

When a sensitive route is intercepted, you decide how Traefik responds to the client:

- **`json`**: Return clean JSON payloads with customizable status codes (e.g. `403 Forbidden` or `404 Not Found`) and custom error messages.
- **`html`**: Render branded warning or company error pages with embedded styling.
- **`text`**: Emit lightweight plain-text error messages.
- **`xml`**: Output standard XML formatted error bodies (`<Error><Status>403</Status>...</Error>`) for SOAP and enterprise services.
- **`captcha`**: Present human verification challenges using **Cloudflare Turnstile**, **hCaptcha**, or **Google reCAPTCHA** without needing any backend captcha server.
- **`redirect`**: Silently deflect attackers to an external honeypot, logging sink, or warning site.
- **`silentDrop`**: Close the TCP connection immediately without emitting any response payload to confuse automated port scanners.
- **`gzipBomb`** *(alias: `bomb`)*: Stream compressed zero-byte blocks that expand ~1000× (e.g. 10 MB expands to ~10 GB in client RAM) with negligible server bandwidth, forcing memory exhaustion (OOM) on vulnerability crawlers (`nikto`, `gobuster`, `dirsearch`).
- **`tarpit`**: Reverse Slowloris defense that trickles individual bytes at slow intervals to tie up scanner socket pools and concurrency workers for minutes.
- **`fakeSuccess`** *(alias: `decoy`)*: Return realistic synthetic honeypot data (`.env` credentials, mock Spring Actuator metrics, dummy `git/HEAD`, fake `wp-login`) to fool attackers and log early warning telemetry.
- **`rateLimitChallenge`** *(alias: `ratelimit`)*: Issue an HTTP `429 Too Many Requests` with a compliant `Retry-After` header to force automated scrapers to back off.
- **`proxy`** *(alias: `mirror`)*: Transparently reverse-proxy probing traffic to an internal forensics/canary container without tipping off the attacker with a 302 redirect.
- **`infiniteStream`** *(alias: `garbagestream`)*: Stream endless chunks of pseudo-random data to fill client disks and crash parsing buffers.

<div class="bomb-callout">
  <div class="bomb-callout-header">
    <span class="bomb-badge">Active Defense</span>
    <h3 class="bomb-callout-title">💣 Gzip Bomb Mode (Bot Counter-Offensive)</h3>
  </div>
  <p class="bomb-callout-body">
    Tired of vulnerability scanners filling your logs? When an unauthorized bot scans for <code>.env</code>, <code>.git</code>, or <code>wp-login.php</code>, RouteWarden can return an enticing <code>200 OK</code> response with <code>Content-Encoding: gzip</code>. The tiny wire stream expands <b>~1000x in the crawler's memory</b> (10 MB expands to ~10 GB), triggering an immediate Out-Of-Memory (OOM) crash in scanning tools like <code>dirsearch</code>, <code>nikto</code>, or Python-based scrapers without consuming server resources.
  </p>
  <div class="bomb-callout-code">
    response: { mode: "gzipBomb", statusCode: 200, gzipBombMB: 10 }
  </div>
  <div class="bomb-callout-warning">
    <span>⚠️</span>
    <span><strong>Caution with Search Engines:</strong> Never attach <code>gzipBomb</code> globally or to legitimate content URLs. Standard web browsers and legitimate search engine spiders (Googlebot, Bingbot) decompress gzip streams automatically. Only target high-confidence malicious probe endpoints (e.g. <code>/.env</code>, <code>/.git</code>, <code>wp-login.php</code>) and ensure <code>enableDefaultAllowPatterns: true</code> remains active so <code>/robots.txt</code> and <code>/sitemap.xml</code> are never bombed.</span>
  </div>
</div>

---

## 60-Second Quickstart

Get RouteWarden running on your Traefik instance in under a minute:

::: code-group

```yaml [File (YAML)]
# dynamic_conf.yml
http:
  middlewares:
    global-warden:
      plugin:
        routewarden:
          enabled: true
          enableDefaultPatterns: true
          # Optional custom regex patterns to guard
          pathPatterns:
            - '(?i)^/admin(/.*)?$'
            - '(?i)^/api/internal(/.*)?$'
          # Safe exceptions (always allowed)
          allowPatterns:
            - '(?i)^/api/internal/health$'
            - '(?i)^/robots\.txt$'
          # Whitelisted VPN or office IPs
          allowedIps:
            - "10.0.0.0/8"
            - "192.168.1.100"
          # Response configuration
          response:
            mode: json
            statusCode: 404
            body: '{"error":"Not Found","message":"The requested resource does not exist"}'

  routers:
    app-router:
      rule: "Host(`example.com`)"
      entryPoints:
        - web
      middlewares:
        - global-warden
      service: app-service
```

```toml [File (TOML)]
# dynamic_conf.toml
[http.routers.app-router]
  rule = "Host(`example.com`)"
  entryPoints = ["web"]
  middlewares = ["global-warden"]
  service = "app-service"

[http.middlewares.global-warden.plugin.routewarden]
  enabled = true
  enableDefaultPatterns = true
  pathPatterns = ["(?i)^/admin(/.*)?$", "(?i)^/api/internal(/.*)?$"]
  allowPatterns = ["(?i)^/api/internal/health$", "(?i)^/robots\\.txt$"]
  allowedIps = ["10.0.0.0/8", "192.168.1.100"]

[http.middlewares.global-warden.plugin.routewarden.response]
  mode = "json"
  statusCode = 404
  body = '{"error":"Not Found","message":"The requested resource does not exist"}'
```

```bash [CLI]
# Docker Compose Labels / CLI equivalent
- "traefik.http.middlewares.global-warden.plugin.routewarden.enabled=true"
- "traefik.http.middlewares.global-warden.plugin.routewarden.enableDefaultPatterns=true"
- "traefik.http.middlewares.global-warden.plugin.routewarden.pathPatterns=(?i)^/admin(/.*)?$,(?i)^/api/internal(/.*)?$"
- "traefik.http.middlewares.global-warden.plugin.routewarden.allowPatterns=(?i)^/api/internal/health$,(?i)^/robots\\.txt$"
- "traefik.http.middlewares.global-warden.plugin.routewarden.allowedIps=10.0.0.0/8,192.168.1.100"
- "traefik.http.middlewares.global-warden.plugin.routewarden.response.mode=json"
- "traefik.http.middlewares.global-warden.plugin.routewarden.response.statusCode=404"
- 'traefik.http.middlewares.global-warden.plugin.routewarden.response.body={"error":"Not Found","message":"The requested resource does not exist"}'
```

:::

---

## Production Case Studies

Real-world deployment patterns demonstrating how engineering teams and self-hosters protect their applications using RouteWarden:

<div class="attack-grid">
  <div class="attack-card">
    <h4>📸 <a href="/examples/case-study-immich">Immich Photo Sharing</a></h4>
    <p>Public photo/album sharing while strictly cloaking administrative, login, and user management APIs under a 404.</p>
  </div>
  <div class="attack-card">
    <h4>💳 <a href="/examples/case-study-webhooks">Zero-Trust Webhooks</a></h4>
    <p>Lock down Stripe/GitHub payment webhook ingress using official provider IP CIDRs and silent TCP drops.</p>
  </div>
  <div class="attack-card">
    <h4>📊 <a href="/examples/case-study-observability">Metrics & Actuator Cloaking</a></h4>
    <p>Shield Prometheus <code>/metrics</code> and Spring Boot <code>/actuator</code> from public scanners while keeping internal scrapers active.</p>
  </div>
  <div class="attack-card">
    <h4>📝 <a href="/examples/case-study-cms-shield">WordPress & CMS Shield</a></h4>
    <p>Defeat brute-force and XML-RPC attacks on <code>wp-login.php</code> using interactive Cloudflare Turnstile / hCaptcha challenges.</p>
  </div>
  <div class="attack-card">
    <h4>🔐 <a href="/examples/case-study-vaultwarden">Password Vaults (Bitwarden)</a></h4>
    <p>Allow public mobile password sync while restricting <code>/admin</code> strictly to WireGuard or Tailscale subnets.</p>
  </div>
  <div class="attack-card">
    <h4>💣 <a href="/examples/case-study-honeypot-staging">Honeypots & Active Gzip Bomb</a></h4>
    <p>Crash scanning bots with <code>gzipBomb</code> decompression traps, reset TCP connections with <code>silentDrop</code>, and cloak staging preview clusters.</p>
  </div>
</div>

---

## Ready to Explore?

- Check out the [Getting Started Guide](/guide/getting-started) for step-by-step installation instructions.
- Learn about the [System Architecture](/guide/architecture) and how RouteWarden processes requests.
- Explore the [Examples Cookbook](/examples/overview) for Docker Compose and Kubernetes manifests.
- Browse all [Production Case Studies](/examples/case-study-immich) for practical production blueprints.
- Read the [Anti-Evasion Security Deep Dive](/reference/anti-evasion) for security test results.
