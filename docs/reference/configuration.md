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

| Key | Type | Default | Description |
|---|---|---|---|
| `mode` | `string` | `"json"` | Response mode: `json`, `html`, `captcha`, `redirect`, `text`, or `silentDrop`. |
| `statusCode` | `int` | `403` | HTTP status code returned to client. |
| `body` | `string` | `""` | Response body for `json`, `html`, or `text` mode. |
| `headers` | `map[string]string` | `{}` | Custom HTTP response headers injected into blocked responses. |
| `redirectUrl` | `string` | `""` | Target URL when `mode: redirect`. |
| `captcha` | `object` | `{}` | Captcha challenge options when `mode: captcha`. |

### Captcha Options (`response.captcha`)

| Key | Type | Default | Description |
|---|---|---|---|
| `provider` | `string` | `"turnstile"` | Captcha provider: `turnstile`, `hcaptcha`, or `recaptcha`. |
| `siteKey` | `string` | `""` | Public site key for the captcha widget. |
| `title` | `string` | `"Verification"` | Heading displayed on the verification challenge page. |
| `template` | `string` | `""` | Optional custom HTML template string override. |
