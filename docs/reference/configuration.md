# Configuration Reference

This reference covers all configuration options available in RouteWarden.

---

## Core Options

| Key | Type | Default | Description |
|---|---|---|---|
| `enabled` | `bool` | `true` | Enables or disables the middleware. When `false`, all traffic passes through. |
| `enableDefaultPatterns` | `bool` | `true` | Enables built-in protection for `.env*`, `.git`, `.aws`, `.sql`, backups, and logs. |
| `enableDefaultAllowPatterns` | `bool` | `true` | Enables built-in allowlist exemptions (`/robots.txt`, `/sitemap.xml`, `/ads.txt`, `/security.txt`, `/.well-known/*`). Set to `false` to disable. |
| `pathPatterns` | `[]string` | `[]` | List of custom regular expressions to block (matches against normalized path). |
| `blockPatterns` | `[]string` | `[]` | Alias for `pathPatterns`. |
| `allowPatterns` | `[]string` | `[]` | Additional custom regex patterns to explicitly allow even if matching blocked rules. |
| `allowedIps` | `[]string` | `[]` | Whitelisted IPv4/IPv6 addresses or CIDR subnets (e.g., `10.0.0.0/8`, `127.0.0.1`). |
| `checkQuery` | `bool` | `false` | Also inspects the URL raw query string for blocked patterns. |
| `statusCode` | `int` | `403` | Default HTTP status code when request is blocked (legacy shortcut). |

---

## Default Allow Patterns

By default, RouteWarden allows standard public informational files and ACME certificate verification paths:

```regex
(?i)^/robots\.txt$
(?i)^/ads\.txt$
(?i)^/security\.txt$
(?i)^/\.well-known(/.*)?$
```

---

## Response Configuration (`response`)

> 📖 **Deep Dive**: For an in-depth breakdown of all 11+ response behaviors, limitations, attacker impacts, and security cautions, visit the **[Response Modes Reference](/reference/response-modes)**.

| Key | Type | Default | Description |
|---|---|---|---|
| `mode` | `string` | `"json"` | Response mode: `json`, `html`, `text`, `xml`, `redirect`, `captcha`, `silentDrop`, `gzipBomb` (`bomb`), `tarpit`, `fakeSuccess` (`decoy`), `rateLimitChallenge` (`ratelimit`), `proxy` (`mirror`), or `infiniteStream` (`garbagestream`). |
| `statusCode` | `int` | `403` | HTTP status code returned to client (use `200` for honeypots / deception, `429` for rate limit challenge). |
| `body` | `string` | `""` | Response body for `json`, `html`, `xml`, or `text` mode. |
| `headers` | `map[string]string` | `{}` | Custom HTTP response headers injected into blocked responses. |
| `redirectUrl` | `string` | `""` | Target URL when `mode: redirect`. |
| `proxyUrl` | `string` | `""` | Target backend honeypot URL when `mode: proxy` (transparent reverse-proxy). |
| `captcha` | `object` | `{}` | Captcha challenge options when `mode: captcha`. |
| `gzipBombMB` | `int` | `10` | Uncompressed stream size in Megabytes when `mode: gzipBomb` (expands ~1000x on client memory). |
| `retryAfterSeconds` | `int` | `300` | Value for `Retry-After` header when `mode: rateLimitChallenge`. |
| `tarpitDelayMs` | `int` | `1000` | Milliseconds between trickle bytes when `mode: tarpit` (stalls scanner connections). |
| `tarpitMaxDurationSeconds` | `int` | `60` | Maximum seconds before terminating stalled connection in `mode: tarpit`. |
| `streamSizeMB` | `int` | `50` | Total garbage data size in Megabytes when `mode: infiniteStream`. |

### Captcha Options (`response.captcha`)

| Key | Type | Default | Description |
|---|---|---|---|
| `provider` | `string` | `"turnstile"` | Captcha provider: `turnstile`, `hcaptcha`, or `recaptcha`. |
| `siteKey` | `string` | `""` | Public site key for the captcha widget. |
| `title` | `string` | `"Verification"` | Heading displayed on the verification challenge page. |
| `template` | `string` | `""` | Optional custom HTML template string override. |
