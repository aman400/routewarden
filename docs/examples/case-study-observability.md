# Case Study: Prometheus & Actuator Observability Cloaking

This case study demonstrates how to protect internal telemetry, metrics scrapers, and diagnostic dumps from public disclosure while preserving uninterrupted collection by internal monitoring pipelines (Prometheus, Datadog, Grafana).

---

## The Threat Model

Most modern microservices expose operational metrics and health dashboards out-of-the-box:
- **Prometheus Scrape Endpoints**: `/metrics`
- **Spring Boot Actuator**: `/actuator`, `/actuator/env`, `/actuator/heapdump`, `/actuator/loggers`
- **Go Runtime Profiling**: `/debug/pprof`, `/debug/vars`

When exposed to the public internet, these endpoints leak proprietary architecture details, server environment variables, memory layouts, and query patterns to competitive reconnaissance bots and exploit kits.

---

## The Solution: Ingress Masking via RouteWarden

RouteWarden intercepts all requests directed at diagnostic and metrics paths:
- **Public Traffic**: Receives a cloaked `404 Not Found` response.
- **Authorized Scrapers**: Requests originating from the internal monitoring cluster (e.g. Prometheus pod CIDR `10.244.0.0/16` or VPC subnet) bypass the filter and receive live metrics.

---

## Traefik Configuration

::: code-group

```yaml [File (YAML)]
# dynamic_conf.yml
http:
  middlewares:
    metrics-cloak:
      plugin:
        routewarden:
          enabled: true
          enableDefaultPatterns: true
          # Guard metrics, profiling, and actuator endpoints
          pathPatterns:
            - '(?i)^/(metrics|server-metrics|telemetry)(/.*)?$'
            - '(?i)^/actuator(/.*)?$'
            - '(?i)^/debug/(pprof|vars)(/.*)?$'
          # Allow internal Prometheus scraper & Kubernetes VPC
          allowedIps:
            - "10.0.0.50/32"    # Dedicated Prometheus server IP
            - "10.244.0.0/16"   # Internal Kubernetes Pod Network
            - "127.0.0.1"       # Localhost diagnostic agent
          response:
            mode: json
            statusCode: 404
            body: '{"error":"Not Found","message":"The requested URL was not found on this server"}'

  routers:
    app-router:
      rule: "Host(`app.example.com`)"
      entryPoints:
        - websecure
      middlewares:
        - metrics-cloak
      service: app-service
```

```toml [File (TOML)]
# dynamic_conf.toml
[http.routers.app-router]
  rule = "Host(`app.example.com`)"
  entryPoints = ["websecure"]
  middlewares = ["metrics-cloak"]
  service = "app-service"

[http.middlewares.metrics-cloak.plugin.routewarden]
  enabled = true
  enableDefaultPatterns = true
  pathPatterns = [
    "(?i)^/(metrics|server-metrics|telemetry)(/.*)?$",
    "(?i)^/actuator(/.*)?$",
    "(?i)^/debug/(pprof|vars)(/.*)?$"
  ]
  allowedIps = ["10.0.0.50/32", "10.244.0.0/16", "127.0.0.1"]

[http.middlewares.metrics-cloak.plugin.routewarden.response]
  mode = "json"
  statusCode = 404
  body = '{"error":"Not Found","message":"The requested URL was not found on this server"}'
```

```bash [CLI]
# Docker Compose Labels equivalent
- "traefik.http.routers.app.rule=Host(`app.example.com`)"
- "traefik.http.routers.app.middlewares=metrics-cloak"
- "traefik.http.middlewares.metrics-cloak.plugin.routewarden.enabled=true"
- "traefik.http.middlewares.metrics-cloak.plugin.routewarden.pathPatterns=(?i)^/(metrics|server-metrics)(/.*)?$,(?i)^/actuator(/.*)?$,(?i)^/debug/pprof(/.*)?$"
- "traefik.http.middlewares.metrics-cloak.plugin.routewarden.allowedIps=10.0.0.50/32,10.244.0.0/16,127.0.0.1"
- "traefik.http.middlewares.metrics-cloak.plugin.routewarden.response.mode=json"
- "traefik.http.middlewares.metrics-cloak.plugin.routewarden.response.statusCode=404"
```

:::
