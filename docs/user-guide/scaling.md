# Scaling

## Manual scaling

`spec.replicas` is passed straight to the Deployment:

```yaml
apiVersion: squid.cdk.clara.net/v1
kind: SquidInstance
metadata:
  name: squid-sample
  namespace: healthcare-tools
spec:
  replicas: 3
  image:
    repository: ubuntu/squid
    tag: latest
  storageClassName: standard
```

```bash
kubectl patch squidinstance squid-sample -n healthcare-tools \
  --type=merge -p '{"spec":{"replicas":3}}'
kubectl rollout status deploy/squid-sample -n healthcare-tools
```

!!! warning "`kubectl scale` does not work"
    The CRD declares no `scale` subresource, so `kubectl scale squidinstance ...` fails with
    `the server could not find the requested resource`. Patch or edit `spec.replicas` instead.

Editing the instance re-renders the Deployment and stamps a fresh
`squid-operator.kubernetes.io/restartedAt`, so a replica change also rolls the existing pods.

## Autoscaling

`spec.hpaSpec` accepts a full `autoscaling/v1` `HorizontalPodAutoscalerSpec`:

```yaml
spec:
  hpaSpec:
    minReplicas: 1
    maxReplicas: 5
    targetCPUUtilizationPercentage: 80
    scaleTargetRef:
      apiVersion: apps/v1
      kind: Deployment
      name: squid-sample
```

!!! danger "Not implemented"
    **The operator does not reconcile `spec.hpaSpec`.** No HPA is created, and the field's presence
    changes nothing. `managedResources` in `internal/controller/squidinstance_controller.go` lists
    Deployment, ConfigMap, ServiceAccount, Service and PVC only. Do not rely on it.

To autoscale today, create the HPA yourself against the operator-managed Deployment:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: squid-sample
  namespace: healthcare-tools
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: squid-sample          # same name as the SquidInstance
  minReplicas: 1
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 80
```

!!! warning "It fights the controller"
    The instance controller rewrites the Deployment — including `spec.replicas` from
    `SquidInstance.spec.replicas` — on every reconcile. An external HPA and the operator will
    overwrite each other's replica count, and each reconcile also triggers a rollout. Until HPA
    support is reconciled properly, treat autoscaling as unsupported and size `spec.replicas` for
    peak load. CPU-target autoscaling also requires resource requests, which the operator does not
    set (see [Configuration](configuration.md#what-is-not-configurable)).

## What scaling actually gives you

Squid replicas sit behind one `LoadBalancer` Service on port `3128`. Consequences:

- **No shared cache.** Each replica caches independently (memory only — there is no cache volume),
  so the hit ratio drops as replicas increase. Use `cache_peer` rules if you need sibling caching.
- **No session affinity.** The Service has no `sessionAffinity`, so consecutive requests from a
  client can land on different replicas. Matters for authentication helpers with per-connection
  state.
- **Shared log volume.** The PVC is `ReadWriteMany` and every replica appends to the same files in
  `/var/log/squid`. More replicas means more write contention and faster growth toward the
  hardcoded `10Gi`.
- **Config is identical everywhere.** All replicas mount the same ConfigMap.

## Sizing

The operator sets no requests or limits, so pods are `BestEffort` unless the namespace has a
`LimitRange`. Add one for predictable scheduling. Rough starting points for a general-purpose
forward proxy, to be validated against your traffic:

| Load                    | Replicas | CPU request | Memory request        |
| ----------------------- | -------- | ----------- | --------------------- |
| Dev / low               | 1        | 100m        | 256Mi                 |
| Moderate                | 2–3      | 500m        | 512Mi + `cache_mem`   |
| High                    | 3+       | 1           | 1Gi + `cache_mem`     |

Memory must cover `cache_mem` from your rules plus Squid's per-connection overhead; a `cache_mem`
larger than the limit gets the pod OOM-killed.

## Storage

The PVC is created once at `10Gi` `ReadWriteMany` and is **never updated** by the controller (PVC
updates are deliberately skipped). Growing it means resizing the PVC directly, if the storage class
allows expansion:

```bash
kubectl patch pvc squid-sample -n healthcare-tools \
  --type=merge -p '{"spec":{"resources":{"requests":{"storage":"20Gi"}}}}'
```

## Monitoring the effect

```bash
kubectl get squidinstance squid-sample -n healthcare-tools     # HEALTH
kubectl get pods -n healthcare-tools -l app.kubernetes.io/name=squid-sample
kubectl top pods -n healthcare-tools -l app.kubernetes.io/name=squid-sample
kubectl exec -n healthcare-tools deploy/squid-sample -- tail /var/log/squid/access.log
```

There are no Squid-level metrics exported — see [Monitoring](monitoring.md).

## Next steps

- [Monitoring](monitoring.md) — what is observable today
- [Configuration](configuration.md) — cache and connection tuning
- [Troubleshooting](troubleshooting.md) — pods not becoming ready
