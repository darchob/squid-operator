# Monitoring

What is observable today, and what is not. The operator exports **no Squid metrics** — there is no
exporter sidecar and no `/metrics` endpoint on the proxy pods. Observability comes from three
places: resource status, Kubernetes events, and the controller-runtime metrics of the manager.

## Resource status

```bash
kubectl get squidinstance -n healthcare-tools
```

```
NAME           HEALTH
squid-sample   Done
```

`HEALTH` (`.status.health`) is `Done` when every managed resource reconciled, `Error` when one
failed. The API also defines `Pending`, `Deployed` and `Merged`, which the controller does not
currently set.

```bash
kubectl get squidconfigs -n healthcare-tools
```

```
NAME             STATE
allow-localnet   Merged
```

`STATE` (`.status.state`) is `Merged`, `Applying`, `Deletion` or `Failed`.

Load balancer address, once assigned:

```bash
kubectl get squidinstance squid-sample -n healthcare-tools \
  -o jsonpath='{.status.loadBalancer.ingress[*].ip}'
```

!!! note
    Neither kind publishes `metav1.Condition`s, so there is no `Ready` condition and
    `kubectl wait --for=condition=Ready` does not work. Poll the printer columns instead:
    `kubectl wait --for=jsonpath='{.status.health}'=Done squidinstance/squid-sample`.

`HEALTH: Done` only means the operator successfully applied its resources. It says nothing about
whether Squid is serving traffic — check the Deployment and pod readiness for that.

## Events

Both controllers emit events on the custom resource:

| Reason    | Emitted when                                      |
| --------- | ------------------------------------------------- |
| `Created` | A managed resource was created (instance)         |
| `Updated` | A managed resource was updated (instance), or a ConfigMap was merged (configs) |

```bash
kubectl describe squidinstance squid-sample -n healthcare-tools
kubectl get events -n healthcare-tools --sort-by='.lastTimestamp'
```

## Controller metrics

The manager serves standard controller-runtime metrics — reconcile counts, durations, errors, work
queue depth, Go runtime, client-go request latencies. There are **no operator-specific metrics**.

Useful series (all provided by controller-runtime, labelled by `controller="squidinstance"` /
`"squidconfigs"`):

| Metric                                            | Meaning                        |
| ------------------------------------------------- | ------------------------------ |
| `controller_runtime_reconcile_total`              | Reconciles, by result          |
| `controller_runtime_reconcile_errors_total`       | Reconcile errors               |
| `controller_runtime_reconcile_time_seconds`       | Reconcile duration histogram   |
| `workqueue_depth`                                 | Pending reconcile requests     |
| `controller_runtime_active_workers`               | Concurrent reconciles          |

### Enabling the endpoint

Metrics are bound to `--metrics-bind-address`, default `:8080`. **The Helm chart starts the manager
with `--metrics-bind-address=0`, which disables them.** The chart exposes no value to override the
manager arguments, so enabling metrics means either editing
`deploy/squid-operator/templates/deployment.yaml`, or deploying with kustomize (`make deploy`),
where `config/default` wires the metrics `Service` and its TLS patch.

For TLS-protected metrics, pass `--metrics-secure` and bind `:8443`; `config/certmanager/certificate-metrics.yaml`
and `config/prometheus/monitor_tls_patch.yaml` provide the certificate and the scrape config.

### Scraping

`config/prometheus/monitor.yaml` ships a `ServiceMonitor` for the manager:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: controller-manager-metrics-monitor
spec:
  endpoints:
    - path: /metrics
      port: http
      scheme: http
  selector:
    matchLabels:
      control-plane: controller-manager
```

It requires the Prometheus Operator CRDs and a metrics `Service` (`config/default/metrics_service.yaml`).
If `config/network-policy/` is applied, scraping is only allowed from namespaces labelled
`metrics: enabled`.

## Logs

Operator:

```bash
kubectl logs -n squid-operator deploy/squid-operator -f
kubectl logs -n squid-operator-system deploy/squid-operator-controller-manager -f   # make deploy
```

The manager runs zap in development mode (`Development: true` in `cmd/main.go`): human-readable,
verbose. Logger names to grep for: `squidInstances`, `squidConfigs`, `squidinstance-resource`,
`squidconfigs-resource`, `setup`.

Squid instances write to the PVC, not stdout:

```bash
kubectl exec -n healthcare-tools deploy/squid-sample -- tail -f /var/log/squid/access.log
kubectl exec -n healthcare-tools deploy/squid-sample -- tail -f /var/log/squid/cache.log
```

To ship them, run a log collector that mounts the same `ReadWriteMany` PVC, or add a sidecar in
`pkg/squid/workloads.go`.

## Squid's own statistics

No exporter is deployed, but the manager interface is available inside the pod if your rules enable
it:

```yaml
spec:
  rules: |
    acl localhost src 127.0.0.1/32
    http_access allow manager localhost
    http_access deny manager
```

```bash
kubectl exec -n healthcare-tools deploy/squid-sample -- squidclient -h 127.0.0.1 -p 3128 mgr:info
kubectl exec -n healthcare-tools deploy/squid-sample -- squidclient -h 127.0.0.1 -p 3128 mgr:mem
```

For Prometheus-visible proxy metrics you would need a `squid-exporter` sidecar, which the operator
does not create.

## Suggested alerts

Based only on metrics that exist:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: squid-operator-alerts
  namespace: squid-operator
spec:
  groups:
    - name: squid-operator
      rules:
        - alert: SquidOperatorReconcileErrors
          expr: rate(controller_runtime_reconcile_errors_total{job=~".*squid.*"}[5m]) > 0
          for: 10m
          labels:
            severity: warning
          annotations:
            summary: Squid operator reconcile errors
            description: "Controller {{ $labels.controller }} is failing to reconcile."

        - alert: SquidOperatorDown
          expr: up{job=~".*squid-operator.*"} == 0
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: Squid operator is not being scraped
```

For instance health, alert on the Deployment rather than the custom resource — `kube-state-metrics`
does not know about `SquidInstance`:

```yaml
- alert: SquidInstanceUnavailable
  expr: kube_deployment_status_replicas_available{deployment=~"squid-.*"} == 0
  for: 5m
  labels:
    severity: critical
```

## Dashboards

Useful panels with the data available:

1. **Operator** — reconcile rate and errors per controller, reconcile duration p95, work queue depth.
2. **Instances** — Deployment available vs desired replicas, pod restarts, CPU/memory from
   `kube-state-metrics` and cAdvisor.
3. **Validation** — number of `validate-*` Jobs, failures (`kube_job_status_failed`), which tracks
   rejected configuration.

Cache hit ratio, request rate and response times need a Squid exporter that is not part of this
project.

## Next steps

- [Troubleshooting](troubleshooting.md) — acting on what you observe
- [Networking](../architecture/networking.md) — ports and policies
- [Security](../development/security.md) — securing the metrics endpoint
