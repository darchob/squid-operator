# Networking

Every port, Service and network policy the operator actually creates or ships.

## Squid instance traffic

### Service

One `Service` per `SquidInstance`, named after it, built by `pkg/squid/networking.go`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: squid-sample
  namespace: healthcare-tools
spec:
  type: LoadBalancer
  ports:
    - name: http
      protocol: TCP
      port: 3128
      targetPort: 3128
  selector:
    app.kubernetes.io/name: squid-sample
```

- The type is **always `LoadBalancer`** — not configurable. On a cluster without a load-balancer
  controller the `EXTERNAL-IP` stays `<pending>`; the `NodePort` and `ClusterIP` still work.
- Port `3128` only. Squid's own `http_port` directive lives in your `SquidConfigs` rules; changing it
  there without changing the container port breaks the probes.
- Once the load balancer reports ingress, the controller copies it into
  `status.loadBalancer` on the instance.

There is **no Ingress** support: no Ingress object is created and the `SquidInstance` spec has no
ingress field. HTTP proxying is L4 traffic to `3128`; an HTTP Ingress is the wrong shape for it
anyway. Expose it through the `LoadBalancer` Service, or add a `Service` of your own.

### Pod ports and probes

The container declares one port and both probes target it:

```yaml
ports:
  - name: http
    containerPort: 3128
    protocol: TCP
livenessProbe:
  tcpSocket: { port: http }
readinessProbe:
  tcpSocket: { port: http }
```

TCP-only probes: a Squid that accepts connections but denies every request still reads as healthy.

### Restricting who can use the proxy

Two independent layers:

- **Squid ACLs**, in `SquidConfigs` rules — the real access control, see
  [Configuration](../user-guide/configuration.md).
- **NetworkPolicy**, written by you. The operator ships no policy for instance pods. Select them
  through the labels it applies:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: squid-sample-clients
  namespace: healthcare-tools
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: squid-sample
      app.kubernetes.io/managed-by: squid-operator
  policyTypes: [Ingress]
  ingress:
    - from:
        - podSelector:
            matchLabels:
              proxy-client: "true"
      ports:
        - protocol: TCP
          port: 3128
```

Remember the egress side: a forward proxy needs to reach the internet, and DNS. A default-deny
egress policy in the namespace will silently break it.

## Operator traffic

| Port   | Served by                     | Configured with                                            |
| ------ | ----------------------------- | ---------------------------------------------------------- |
| `9443` | Admission webhook server      | controller-runtime default; certs at `/tmp/k8s-webhook-server/serving-certs` |
| `8081` | `healthz` / `readyz`          | `--health-probe-bind-address` (default `:8081`)            |
| `8080` | Prometheus metrics (plain)    | `--metrics-bind-address` (default `:8080`; `0` disables)   |
| `8443` | Prometheus metrics (TLS)      | `--metrics-bind-address=:8443 --metrics-secure`             |

The Helm chart starts the manager with `--metrics-bind-address=0`, so metrics are off unless you
change it. The `Service` in front of the webhook is `<release>-webhook-service:443 → 9443`.

HTTP/2 is disabled on both the metrics and webhook servers unless `--enable-http2` is passed, to
avoid the Stream Cancellation / Rapid Reset CVEs (GHSA-qppj-fm5r-hxr3, GHSA-4374-p667-p6c8).

### Webhook certificates

cert-manager issues the serving certificate (`config/certmanager/`, or `templates/certificats.yaml`
in the chart) and injects the CA bundle into both webhook configurations through the
`cert-manager.io/inject-ca-from` annotation. Both configurations use `failurePolicy: Fail`: if the
webhook is unreachable, creating or updating `SquidInstance` / `SquidConfigs` objects fails closed.

### Shipped network policies

`config/network-policy/` (applied only if you include that kustomize component) restricts ingress to
the manager pod:

- `allow-metrics-traffic` — port `8443` from namespaces labelled `metrics: enabled`
- `allow-webhook-traffic` — port `443` from namespaces labelled `webhook: enabled`

The second one matters: with that policy active, `SquidInstance`/`SquidConfigs` objects can only be
created from namespaces carrying `webhook: enabled`, because the API server call to the webhook is
blocked otherwise.

## Validation Jobs

The `SquidConfigs` webhook runs `squid -k parse` in a Job in the **resource's own namespace**, using
the instance's image. Its pod needs to pull that image, so it is subject to the namespace's egress
policies and to any `imagePullSecrets` requirements. A blocked or slow pull shows up as a 2-minute
admission timeout.

## DNS

The operator sets no DNS configuration on the pods; they use the cluster resolver. Squid's own
resolution behaviour is configured through rules, for example:

```yaml
spec:
  rules: |
    dns_nameservers 10.96.0.10
    dns_v4_first on
```

## Troubleshooting

| Symptom                              | Check                                                                   |
| ------------------------------------ | ----------------------------------------------------------------------- |
| `EXTERNAL-IP` stuck `<pending>`      | Cluster has no load-balancer controller; use the `NodePort` or `ClusterIP` |
| Proxy reachable but everything denied | Squid ACLs in your `SquidConfigs`, not the network layer                 |
| `context deadline exceeded` on apply | Webhook unreachable — netpol, webhook Service, cert-manager             |
| Endpoints empty                      | Pods not ready — TCP probe on `3128` failing, check pod logs            |

## Next steps

- [Components](components.md) — where each object is built
- [Monitoring](../user-guide/monitoring.md) — scraping the manager metrics
- [Security](../development/security.md) — RBAC and hardening
