# System Architecture

RouteWarden operates as an in-line HTTP middleware within Traefik's proxy pipeline. Every incoming request undergoes a strict, multi-stage inspection lifecycle before being either passed downstream or intercepted.

---

## Architectural Flow

![RouteWarden Architecture](/architecture.png)

```
Incoming Request
       │
       ▼
┌─────────────────────────┐
│  IP Whitelist Evaluator │ ─── (Matches Allowed IP/CIDR?) ───► Forward to Upstream
└─────────────────────────┘
       │ No
       ▼
┌─────────────────────────┐
│  Path Anti-Evasion      │
│  - Multi-layer Unescape │
│  - Semicolon Strip      │
│  - Backslash Normalize  │
│  - Traversal Canonical  │
└─────────────────────────┘
       │ Normalized Candidates
       ▼
┌─────────────────────────┐
│ Allowlist Regex Engine  │ ─── (Matches allowPatterns?)   ───► Forward to Upstream
└─────────────────────────┘
       │ No
       ▼
┌─────────────────────────┐
│ Sensitive Matcher       │
│ - Built-in Patterns     │ ─── (No Match)                 ───► Forward to Upstream
│ - Custom pathPatterns   │
└─────────────────────────┘
       │ Matches Forbidden Pattern
       ▼
┌─────────────────────────┐
│ Response Handler        │ ───► JSON / HTML / Captcha / Redirect / Silent Drop
└─────────────────────────┘
```

---

## Modular Component Design

RouteWarden is constructed with clean, decoupled Go components adhering to Yaegi interpreter specifications:

1. **`ip_filter.go` (`IPFilter`)**:
   - Parses exact IPv4 (`127.0.0.1`), IPv6 (`::1`), and CIDR networks (`10.0.0.0/8`).
   - Resolves client IP by prioritizing proxy headers (`X-Forwarded-For`, `X-Real-IP`) and falling back to socket `RemoteAddr`.
2. **`path_normalizer.go` (`PathNormalizer`)**:
   - Extracts URL paths, unescapes recursive percent-encoding (`%252e%252e` ➔ `..`), strips semicolon matrix parameters (`/;param/.env`), normalizes Windows backslashes (`\`), and generates canonical candidate paths.
3. **`config.go` (`Config`)**:
   - Defines plugin parameters, default sensitive regex sets, allowlist patterns, and compiler factories.
4. **`response_handler.go` (`ResponseHandler`)**:
   - Manages HTTP response emission for custom JSON, HTML, redirect codes, and embedded Captcha challenges (**Turnstile**, **hCaptcha**, **reCAPTCHA**).
5. **`routewarden.go` (`RouteWarden`)**:
   - Implements Traefik's standard `http.Handler` interface.
