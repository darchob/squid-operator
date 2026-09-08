# Quick Start

Deploy a Squid proxy and add a configuration fragment to it. Assumes the operator is already
installed — see [Installation](installation.md).

## Prerequisites

- Operator running, CRDs present (`kubectl get crd | grep squid.cdk.clara.net`)
- A `StorageClass` supporting `ReadWriteMany`
- A namespace to work in (`healthcare-tools` in the examples below)

## 1. Create a SquidInstance

```yaml title="squid-instance.yaml"
apiVersion: squid.cdk.clara.net/v1
kind: SquidInstance
metadata:
  name: squid-sample
  namespace: healthcare-tools
spec:
  replicas: 1
  image:
    repository: ubuntu/squid
    tag: latest
  storageClassName: standard   # must support ReadWriteMany
```

```bash
kubectl apply -f squid-instance.yaml
```

The instance controller creates five resources, all named `squid-sample` in the same namespace:

| Resource                | Detail                                                                 |
| ----------------------- | ---------------------------------------------------------------------- |
| `Deployment`            | Squid container, port `3128` (`http`), TCP liveness and readiness probes |
| `ConfigMap`             | Mounted at `/etc/squid/conf.d/`, seeded with `squid-init.conf`           |
| `PersistentVolumeClaim` | `ReadWriteMany`, `10Gi`, mounted at `/var/log/squid`                    |
| `Service`               | Type `LoadBalancer`, port `3128`                                        |
| `ServiceAccount`        | Used by the pods                                                        |

```bash
kubectl get squidinstance -n healthcare-tools
```

```
NAME           HEALTH
squid-sample   Done
```

`HEALTH` is `Done` once every owned resource reconciled, `Error` otherwise.

## 2. Add a configuration fragment

`SquidConfigs` selects its target instance through the **`squid.ckd.clara.net/instance`
annotation** — there is no spec field for it. The target must be a `SquidInstance` in the same
namespace.

```yaml title="squid-configs.yaml"
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
    acl localnet src 172.16.0.0/12
    acl localnet src 192.168.0.0/16
    http_access allow localnet
```

```bash
kubectl apply -f squid-configs.yaml
```

On admission the validating webhook creates a `Job` named `validate-allow-localnet-<suffix>` from
the target instance's Squid image, writes the rules to `/etc/squid/conf.d/00-squid.conf` and runs
`squid -k parse`. Admission waits up to 2 minutes: if the parse fails, `kubectl apply` is rejected
with the error and nothing reaches the ConfigMap.

Once merged:

```bash
kubectl get squidconfigs -n healthcare-tools
```

```
NAME             STATE
allow-localnet   Merged
```

```bash
kubectl get configmap squid-sample -n healthcare-tools -o jsonpath='{.data}' | jq keys
```

```json
["allow-localnet.conf", "squid-init.conf"]
```

Each `SquidConfigs` owns the key `<its-name>.conf`. The instance controller notices the ConfigMap
change and stamps `squid-operator.kubernetes.io/restartedAt` on the Deployment pod template, so the
pods roll and Squid reloads with the new rules.

## 3. Use the proxy

```bash
kubectl get svc squid-sample -n healthcare-tools
```

```
NAME           TYPE           CLUSTER-IP     EXTERNAL-IP     PORT(S)          AGE
squid-sample   LoadBalancer   10.96.14.201   10.0.12.34      3128:31128/TCP   2m
```

From inside the cluster:

```bash
kubectl run curl --rm -it --image=curlimages/curl --restart=Never -- \
  curl -x http://squid-sample.healthcare-tools.svc.cluster.local:3128 -I https://example.com
```

Once the load balancer reports an address, it is mirrored to `status.loadBalancer` on the instance.

## 4. Clean up

```bash
kubectl delete -f squid-configs.yaml    # webhook refuses while the key is still in the ConfigMap
kubectl delete -f squid-instance.yaml   # finalizer removes the five owned resources
```

## Next steps

- [Configuration Guide](../user-guide/configuration.md) — what to put in `spec.rules`
- [Basic Usage](../user-guide/basic-usage.md) — everyday operations
- [Troubleshooting](../user-guide/troubleshooting.md) — when admission or rollout fails
- [Examples](../reference/examples.md) — more manifests
