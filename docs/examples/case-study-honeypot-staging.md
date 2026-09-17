# Case Study: Honeypot Deflection & Tarpit Scanning Sink

This case study demonstrates how to use RouteWarden to deflect reconnaissance bots and malicious vulnerability scanners into honeypots or silent TCP drops.

---

## The Threat Model

Public IPv4 and IPv6 addresses receive continuous automated requests looking for `.env`, `/phpinfo.php`, `/.git`, and common vulnerable endpoints. 

While returning an HTTP `403 Forbidden` or `404 Not Found` works, scanners will often continue iterating through hundreds of file paths, consuming reverse proxy bandwidth and generating thousands of log lines.

---

## Strategy A: Silent Connection Drops (`silentDrop: true`)

Rather than allocating memory buffers and sending an HTTP status response, RouteWarden's `silentDrop` mode closes the underlying TCP connection immediately (or returns an empty payload).

```yaml
http:
  middlewares:
    scanner-drop:
      plugin:
        routewarden:
          enabled: true
          enableDefaultPatterns: true
          # Close connection immediately on probe attempts
          silentDrop: true
```

### Result:
- Port scanners receive a connection reset (`TCP RST` or EOF).
- Automated vulnerability tools flag the endpoint as dead or unresponsive, prompting them to abandon the host.
- Zero server bandwidth spent delivering HTML error bodies.

---

## Strategy B: External Honeypot Deflection (`mode: redirect`)

When an attacker accesses any sensitive file pattern, RouteWarden can issue an HTTP `302/307 Redirect` to an external honeypot, a public loopback (`http://127.0.0.1`), or an FBI/IC3 reporting endpoint:

```yaml
http:
  middlewares:
    honeypot-deflect:
      plugin:
        routewarden:
          enabled: true
          enableDefaultPatterns: true
          response:
            mode: redirect
            statusCode: 307
            redirectUrl: "https://honeypot.internal.corp/capture"
            headers:
              X-RouteWarden-Deflected: "true"
```

---

## Strategy C: Staging & Preview Environment Cloaking

For pull-request preview environments (e.g., `pr-142.staging.example.com`), competitors or automated crawlers shouldn't index unreleased code:

```yaml
http:
  middlewares:
    staging-guard:
      plugin:
        routewarden:
          enabled: true
          # Block everything by default
          pathPatterns:
            - '^/.*$'
          # Disable standard public exemptions (robots.txt, sitemap.xml)
          enableDefaultAllowPatterns: false
          # Allow exclusively developer and office subnets
          allowedIps:
            - "10.0.0.0/8"
            - "100.64.0.0/10" # Tailscale
            - "203.0.113.50/32" # Corporate NAT IP
          response:
            mode: json
            statusCode: 404
            body: '{"error":"Not Found"}'
```
