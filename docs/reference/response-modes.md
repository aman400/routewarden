# Response Modes Reference

RouteWarden features a modular response engine that dictates exactly how Traefik handles blocked requests. Rather than only offering a generic static 403 error page, RouteWarden supports **11 deterministic response modes** ranging from standard REST error bodies to interactive bot challenges and offensive honeypot traps.

---

## Response Modes at a Glance

| Mode | Category | Target Use Case | Default Status Code |
|---|---|---|---|
| [`text`](#1-plain-text-text) | Standard Error | Minimalist plain text response | `403 Forbidden` |
| [`json`](#2-json-payload-json) | Standard Error | Clean REST API error objects for single-page apps | `403 Forbidden` |
| [`html`](#3-custom-html-html) | Standard Error | Branded corporate or application error pages | `403 Forbidden` |
| [`xml`](#4-xml-error-xml) | Enterprise | SOAP, XML-RPC, and legacy enterprise backends | `403 Forbidden` |
| [`redirect`](#5-url-redirect-redirect) | Redirection | External honeypots, loopback sinkholes, or portals | `302 Found` |
| [`captcha`](#6-interactive-challenge-captcha) | Verification | Cloudflare Turnstile, hCaptcha, or Google reCAPTCHA | `403 Forbidden` |
| [`silentDrop`](#7-silent-tcp-drop-silentdrop) | Passive Defense | Instant TCP socket teardown to confuse port scanners | N/A (TCP RST/EOF) |
| [`gzipBomb`](#8-gzip-decompression-bomb-gzipbomb-bomb) | Active Defense | Memory exhaustion trap that crashes crawler RAM | `200 OK` |
| [`tarpit`](#9-reverse-slowloris-tarpit-tarpit) | Active Defense | Extremely slow trickle stream to tie up bot sockets | `200 OK` |
| [`fakeSuccess`](#10-synthetic-honeypot-decoy-fakesuccess-decoy) | Deception | Realistic fake `.env`, `actuator`, `git`, or `phpinfo` | `200 OK` |
| [`rateLimitChallenge`](#11-rate-limit-backoff-ratelimitchallenge-ratelimit) | Throttling | HTTP 429 with compliant `Retry-After` header | `429 Too Many Requests` |
| [`proxy`](#12-transparent-honeypot-proxy-proxy-mirror) | Forensics | Transparent reverse-proxy into canary honeypot container | Upstream status |
| [`infiniteStream`](#13-infinite-garbage-stream-infinitestream) | Active Defense | Endless pseudo-random characters to fill crawler storage | `200 OK` |

---

## 1. Plain Text (`text`)

The simplest and lightest mode. Returns a raw text string with `Content-Type: text/plain; charset=utf-8` and `X-Content-Type-Options: nosniff`.

### Example Configuration

```yaml
response:
  mode: text
  statusCode: 403
  body: "Access Denied: You do not have permission to view this resource."
```

### Limitations & Caveats
- Browsers render unstyled plain text.
- Automated API clients that expect JSON will receive an unparsed string.

---

## 2. JSON Payload (`json`)

Returns structured JSON data with `Content-Type: application/json` and `X-Content-Type-Options: nosniff`. If `body` is omitted, RouteWarden generates a standard JSON error structure containing the HTTP status code and error description.

### Example Configuration

```yaml
response:
  mode: json
  statusCode: 404
  body: '{"status":"error","code":404,"message":"The requested endpoint does not exist"}'
```

### Limitations & Caveats
- Ensure `body` contains valid JSON syntax; RouteWarden injects the raw string directly into the HTTP response stream.

---

## 3. Custom HTML (`html`)

Renders custom HTML pages with `Content-Type: text/html; charset=utf-8`. If `body` is left empty, RouteWarden serves a clean minimalist dark-mode fallback page.

### Example Configuration

```yaml
response:
  mode: html
  statusCode: 403
  body: |
    <!DOCTYPE html>
    <html>
      <head><title>Access Restricted</title></head>
      <body style="font-family: sans-serif; text-align: center; padding: 4rem;">
        <h1>403 Restricted Zone</h1>
        <p>This endpoint is monitored and protected by RouteWarden.</p>
      </body>
    </html>
```

---

## 4. XML Error (`xml`)

Emits structured XML documents with `Content-Type: application/xml; charset=utf-8`. Essential for SOAP web services, legacy payment gateways, or enterprise Java/Spring backends where upstream clients fail if returned non-XML payloads.

### Example Configuration

```yaml
response:
  mode: xml
  statusCode: 403
  body: |
    <?xml version="1.0" encoding="UTF-8"?>
    <soap:Fault xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
      <faultcode>soap:Client</faultcode>
      <faultstring>Access to protected endpoint is denied</faultstring>
    </soap:Fault>
```

---

## 5. URL Redirect (`redirect`)

Redirects the client to another URL using HTTP 301, 302, 307, or 308. If the configured `statusCode` is outside the 300–308 range, RouteWarden automatically falls back to `302 Found`. If `redirectUrl` is omitted, it defaults to `/`.

### Example Configuration

```yaml
response:
  mode: redirect
  statusCode: 307
  redirectUrl: "https://honeypot.corp.example.com/capture"
  headers:
    X-RouteWarden-Action: "deflected"
```

### Limitations & Caveats
- Advanced scanners detect 30x status codes and record that the endpoint redirected rather than interacting with the destination. For transparent deflection without changing client URLs, use [`proxy`](#12-transparent-honeypot-proxy-proxy-mirror).

---

## 6. Interactive Challenge (`captcha`)

Serves a self-contained security verification page prompting users to complete a CAPTCHA challenge before proceeding. RouteWarden comes with pre-built dark-mode templates for **Cloudflare Turnstile**, **hCaptcha**, and **Google reCAPTCHA**.

### Example Configuration

```yaml
response:
  mode: captcha
  statusCode: 403
  captcha:
    provider: turnstile # or "hcaptcha", "recaptcha"
    siteKey: "0x4AAAAAAtestkey123"
    title: "RouteWarden Security Verification"
```

### Limitations & Caveats
- Validating the submitted captcha token on POST requires a backend authentication service or an accompanying Traefik forward-auth plugin. RouteWarden serves the challenge interface directly at the edge.

---

## 7. Silent TCP Drop (`silentDrop`)

Rather than allocating HTTP response buffers, RouteWarden hijacks the underlying TCP connection and terminates it immediately (`TCP RST` / connection reset) using Go's `http.Hijacker`.

### Example Configuration

```yaml
response:
  mode: silentDrop
# Alternatively via top-level flag:
silentDrop: true
```

### Why Use Silent Drops?
- **Zero Server Bandwidth**: No HTTP headers or status codes are sent over the wire.
- **Port Scanner Disruption**: Automated scanners (like `masscan` or `nmap`) receive an unexpected connection reset, causing them to flag the port or endpoint as dead.

---

## 8. Gzip Decompression Bomb (`gzipBomb` / `bomb`)

An active-defense countermeasure. RouteWarden serves an HTTP `200 OK` response with `Content-Encoding: gzip` containing a stream of repeating zero-byte blocks.

```yaml
response:
  mode: gzipBomb # Alias: "bomb"
  statusCode: 200 # Looks like an enticing 200 OK to crawlers
  gzipBombMB: 10 # 10MB compressed stream expands to ~10GB in client RAM
```

### How It Works
1. A 10 MB stream of zeroes compresses into just a few kilobytes over the wire.
2. Automated crawlers (`dirsearch`, `nikto`, `gobuster`, Python scrapers) automatically decompress HTTP responses in system memory.
3. Expanding 10 MB to **10+ GB in memory** triggers an Out-Of-Memory (OOM) crash in the scanner process or freezes its worker thread pool.

::: danger CRITICAL WARNING: Do Not Attach Globally
Standard web browsers and legitimate search engine spiders (**Googlebot**, **Bingbot**) automatically decompress gzip encoding.
- **NEVER** configure `gzipBomb` on global entrypoints or legitimate content routes.
- **Always** keep `enableDefaultAllowPatterns: true` so `/robots.txt` and `/sitemap.xml` are exempt.
- **Only** apply to explicit exploit probe paths (`^/\.env`, `^/\.git`, `^/wp-login\.php`, `^/phpmyadmin`).
:::

---

## 9. Reverse Slowloris Tarpit (`tarpit`)

Stalls the requesting bot's TCP connection by sending an HTTP `200 OK` header and trickling individual single bytes at slow, deliberate intervals.

### Example Configuration

```yaml
response:
  mode: tarpit
  statusCode: 200
  tarpitDelayMs: 1000 # Send 1 byte every 1000ms (1 second)
  tarpitMaxDurationSeconds: 120 # Force connection close after 2 minutes
```

### How It Works & Attacker Impact
Automated vulnerability scanners operate with limited thread concurrency pools (typically 10 to 50 threads). When an attacker probes 20 honeypot endpoints, RouteWarden holds all 20 worker threads hostage for 120 seconds, grinding their scanning operations to a halt.

### Limitations & Traefik Considerations
- Holding open connections consumes an open file descriptor in Traefik. Ensure `tarpitMaxDurationSeconds` is set to a reasonable limit (e.g. 60–120s) so your Traefik instance does not exhaust connection limits.

---

## 10. Synthetic Honeypot Decoy (`fakeSuccess` / `decoy`)

Feeds automated vulnerability scanners convincing synthetic mock data so they record a false-positive and waste time attempting to exploit dummy credentials.

### Example Configuration

```yaml
response:
  mode: fakeSuccess # Alias: "decoy"
  statusCode: 200
```

### Out-of-the-Box Synthetic Payloads

| Probed Endpoint | RouteWarden Synthetic Response Payload |
|---|---|
| `/.env` | Realistic Laravel `.env` with dummy MySQL credentials and base64 app key |
| `/actuator/health` | Spring Boot Actuator health JSON with dummy disk space metrics |
| `/.git/HEAD` | Standard Git repository pointer (`ref: refs/heads/master`) |
| `/phpinfo.php` | Sanitized mock HTML document displaying PHP 8.2 system specs |
| `/wp-login.php` | Clean WordPress administration login form |
| Other endpoints | Generic REST success JSON `{"status":"success","data":{"id":1,"active":true}}` |

You can override any synthetic payload by supplying custom text in the `body` option.

---

## 11. Rate Limit Challenge (`rateLimitChallenge` / `ratelimit`)

Returns an HTTP `429 Too Many Requests` status code with an injected standard `Retry-After: <seconds>` header.

### Example Configuration

```yaml
response:
  mode: rateLimitChallenge # Aliases: "ratelimit", "backoff"
  statusCode: 429
  retryAfterSeconds: 300 # Instructs client to wait 5 minutes
  body: '{"error":"Too Many Requests","retryAfter":300,"message":"Throttled by RouteWarden"}'
```

### Attacker Impact
Polite web crawlers, search indexers, and compliance-oriented scrapers respect the `Retry-After` header and will pause all outbound requests until the cooldown window expires.

---

## 12. Transparent Honeypot Proxy (`proxy` / `mirror`)

Transparently reverse-proxies the unauthorized request into an internal honeypot or forensic capture container (e.g. Cowrie, OpenCanary, or an isolated sandbox) without issuing an HTTP redirect.

### Example Configuration

```yaml
response:
  mode: proxy # Alias: "mirror"
  proxyUrl: "http://canary-honeypot:8080"
```

### Why Use Transparent Proxying?
- **Attacker Blindness**: Attackers cannot detect that they were redirected because the URL, IP, and TLS handshake remain identical.
- **Forensic Payload Collection**: Captures full POST bodies, exploit payloads, SQL injection strings, and evasion attempts inside an isolated environment for security analysis.

---

## 13. Infinite Garbage Stream (`infiniteStream` / `garbagestream`)

Sends an HTTP `200 OK` response with `Content-Type: application/octet-stream` and continuously streams pseudo-random garbage characters at full TCP connection speed.

### Example Configuration

```yaml
response:
  mode: infiniteStream # Alias: "garbagestream"
  statusCode: 200
  streamSizeMB: 100 # Stream 100 Megabytes of random data
```

### Attacker Impact
- Fills local disk storage if the crawler saves HTTP response files to disk.
- Freezes or crashes regex parsers, HTML tokenizers, and JSON decoders that attempt to read the entire body into memory before parsing.
