# Basic Usage

Everyday operations against the two custom resources. See [Quick Start](../getting-started/quickstart.md)
for the first deployment.

## Create a proxy

```yaml title="instance.yaml"
apiVersion: squid.cdk.clara.net/v1
kind: SquidInstance
metadata:
  name: squid-sample
  namespace: healthcare-tools
spec:
  replicas: 2
  image:
    repository: ubuntu/squid
    tag: latest
  storageClassName: standard   # must support ReadWriteMany
```

```bash
kubectl apply -f instance.yaml
kubectl get squidinstance -n healthcare-tools
```

```
NAME           HEALTH
squid-sample   Done
```

`spec.replicas`, `spec.image` and `spec.storageClassName` are all required by the CRD. There is no
field for resources, ports, service type or Ingress — see [Configuration](configuration.md#what-is-not-configurable).

## Add configuration

Rules live in `SquidConfigs` objects, each targeting an instance through an annotation:

```yaml title="acl.yaml"
apiVersion: squid.cdk.clara.net/v1
kind: SquidConfigs
metadata:
  name: allow-localnet
  namespace: healthcare-tools
  annotations:
    squid.ckd.clara.net/instance: squid-sample
spec:
  rules: |
    acl localnet src 10.0.0.0/8
    http_access allow localnet
```

```bash
kubectl apply -f acl.yaml
kubectl get squidconfigs -n healthcare-tools
```

```
NAME             STATE
allow-localnet   Merged
```

Apply as many as you need — each one owns the ConfigMap key `<its-name>.conf`:

```bash
kubectl get cm squid-sample -n healthcare-tools -o jsonpath='{.data}' | jq keys
```

```json
["allow-localnet.conf", "cache-tuning.conf", "squid-init.conf"]
```

`squid-init.conf` is seeded by the operator when it creates the ConfigMap and is not owned by any
`SquidConfigs`.

## Inspect what is deployed

Everything the instance owns shares its name:

```bash
kubectl get deploy,svc,cm,pvc,sa -n healthcare-tools -l app.kubernetes.io/name=squid-sample
```

```bash
kubectl describe squidinstance squid-sample -n healthcare-tools     # events: Created / Updated
kubectl get events -n healthcare-tools --sort-by='.lastTimestamp'
```

The rendered configuration, as Squid sees it:

```bash
kubectl exec -n healthcare-tools deploy/squid-sample -- ls /etc/squid/conf.d/
kubectl exec -n healthcare-tools deploy/squid-sample -- cat /etc/squid/conf.d/allow-localnet.conf
```

## Change configuration

Edit the `SquidConfigs` and re-apply:

```bash
kubectl apply -f acl.yaml
```

The webhook re-validates with `squid -k parse` before anything is written, the controller rewrites
the ConfigMap key, and the instance controller stamps
`squid-operator.kubernetes.io/restartedAt` on the Deployment pod template, rolling the pods.

```bash
kubectl rollout status deploy/squid-sample -n healthcare-tools
```

Re-applying byte-identical rules is detected as a duplicate and does not trigger a rollout.

## Scale

```bash
kubectl patch squidinstance squid-sample -n healthcare-tools \
  --type=merge -p '{"spec":{"replicas":3}}'
```

`kubectl scale` does not work — the CRD has no `scale` subresource. `spec.hpaSpec` exists in the API
but is not reconciled. See [Scaling](scaling.md).

## Use the proxy

```bash
kubectl get svc squid-sample -n healthcare-tools
```

In-cluster clients:

```bash
export http_proxy=http://squid-sample.healthcare-tools.svc.cluster.local:3128
export https_proxy=$http_proxy
```

Quick check:

```bash
kubectl run curl --rm -it --image=curlimages/curl --restart=Never -- \
  curl -sS -x http://squid-sample.healthcare-tools.svc.cluster.local:3128 -I https://example.com
```

A `403 Forbidden` from Squid means it is running and your ACLs denied the request — a configuration
result, not a failure of the operator.

## Read logs

Squid logs go to the PVC mounted at `/var/log/squid`, not to stdout:

```bash
kubectl exec -n healthcare-tools deploy/squid-sample -- tail -f /var/log/squid/access.log
kubectl exec -n healthcare-tools deploy/squid-sample -- tail -f /var/log/squid/cache.log
```

`kubectl logs` shows only what the image writes to stdout. The PVC is `ReadWriteMany` and shared by
all replicas — plan for log rotation (`logfile_rotate` in your rules) so `10Gi` is not filled up.

## Remove things

```bash
kubectl delete squidconfigs allow-localnet -n healthcare-tools   # removes its ConfigMap key, rolls pods
kubectl delete squidinstance squid-sample -n healthcare-tools    # removes all five owned resources
```

Delete fragments **before** their instance: the delete webhook reads the instance ConfigMap, and
without it deletion is refused and the finalizer has to be cleared by hand.

## Common issues

| Symptom                                                   | Cause                                                                 |
| --------------------------------------------------------- | --------------------------------------------------------------------- |
| `missing annotation squid.ckd.clara.net/instance`         | `SquidConfigs` has no target annotation (note the `ckd` spelling)     |
| `validation job failed`                                   | `squid -k parse` rejected the rules — see the Job's pod logs          |
| Apply hangs, then `context deadline exceeded`             | Validation Job could not run within 2 minutes (image pull, quota, RBAC) |
| `HEALTH: Error` on the instance                           | A managed resource failed; check the manager logs                    |
| Pod `Pending`                                             | PVC unbound — the storage class does not support `ReadWriteMany`      |

More in [Troubleshooting](troubleshooting.md).

## Next steps

- [Configuration](configuration.md) — what to write in `spec.rules`
- [Monitoring](monitoring.md) — status, events, metrics
- [API Reference](../reference/api.md) — every field
