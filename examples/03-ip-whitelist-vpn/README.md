# Example 03: IP & CIDR Subnet Whitelisting

This example configures RouteWarden to protect sensitive endpoints (`/admin`, `/metrics`, and default sensitive files), while allowing verified corporate VPNs and internal subnets (`10.0.0.0/8`, `192.168.1.100`) to pass through without interruption.

## Running the Example

```bash
docker compose up -d
```

## Verification

### 1. Request from Untrusted IP (Blocked)
```bash
curl -i -H "Host: admin.localhost" http://localhost/admin
# Expected: HTTP/1.1 403 Forbidden
# {"error":"Forbidden","message":"Restricted to trusted VPN / Internal IPs"}
```

### 2. Request from Whitelisted IP via Proxy (`X-Forwarded-For`) (Allowed)
```bash
curl -i -H "Host: admin.localhost" -H "X-Forwarded-For: 10.0.0.25" http://localhost/admin
# Expected: Passes through to upstream service!
```
