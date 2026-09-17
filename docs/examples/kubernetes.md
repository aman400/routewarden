# Example 6: Kubernetes IngressRoute (Traefik CRD)

Deploy RouteWarden inside Kubernetes clusters using Traefik's Custom Resource Definitions (`Middleware` and `IngressRoute`).

---

## Static Cluster Configuration

Ensure the plugin is enabled in Traefik's Helm chart values:

```yaml
additionalArguments:
  - "--experimental.plugins.routewarden.modulename=github.com/routewarden/traefik-warden"
  - "--experimental.plugins.routewarden.version={{version}}"
```

---

## Manifests

### `middleware.yaml`
```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: routewarden-k8s-shield
  namespace: default
spec:
  plugin:
    routewarden:
      enabled: true
      enableDefaultPatterns: true
      checkQuery: true
      allowedIps:
        - "10.0.0.0/8"
        - "172.16.0.0/12"
      response:
        mode: json
        statusCode: 403
        body: '{"error":"Forbidden","source":"k8s-routewarden-crd"}'
```

### `ingressroute.yaml`
```yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: secured-app-ingress
  namespace: default
spec:
  entryPoints:
    - web
    - websecure
  routes:
    - match: Host(`production.example.com`)
      kind: Rule
      services:
        - name: app-service
          port: 80
      middlewares:
        - name: routewarden-k8s-shield
```

---

## Deployment Commands

```bash
# Apply both CRDs
kubectl apply -f middleware.yaml
kubectl apply -f ingressroute.yaml

# Inspect middleware status
kubectl get middleware routewarden-k8s-shield -o yaml
```
