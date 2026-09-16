# Example 02: Global EntryPoint Shield

This architecture configures RouteWarden directly on Traefik's entrypoint (`--entrypoints.web.http.middlewares=global-warden@docker`), instantly protecting **every container and route** in your cluster without requiring duplicate middleware labels on every single service.

## Running the Example

```bash
docker compose up -d
```

## Verification

Both `frontend.localhost` and `api.localhost` are automatically shielded:

```bash
# Test Frontend Route
curl -i -H "Host: frontend.localhost" http://localhost/.env
# Expected: HTTP/1.1 403 Forbidden {"error":"Forbidden","scope":"global-shield"}

# Test API Route
curl -i -H "Host: api.localhost" http://localhost/.git/HEAD
# Expected: HTTP/1.1 403 Forbidden {"error":"Forbidden","scope":"global-shield"}
```
