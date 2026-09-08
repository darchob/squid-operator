# Troubleshooting

Failures specific to this operator first, then the generic Kubernetes and Squid ones.

## SquidConfigs is rejected on apply

### `missing annotation squid.ckd.clara.net/instance`

The validating webhook requires the target annotation. Note the **`ckd`** spelling — it does not
match the `cdk.clara.net` API group, and `squid.cdk.clara.net/instance` is silently ignored.

```yaml
metadata:
  annotations:
    squid.ckd.clara.net/instance: squid-sample
```

### `squidinstances.squid.cdk.clara.net "x" not found`

The annotation names an instance that does not exist in the same namespace. The webhook reads it to
pick the image for the validation Job. Cross-namespace targeting is not supported.

### `validation job failed`

`squid -k parse` rejected the rules. The Job is gone from the API response but its pod holds the
reason:

```bash
kubectl get jobs -n healthcare-tools | grep validate-
kubectl logs -n healthcare-tools job/validate-<name>-<suffix>
```

Typical output:

```
FATAL: Bungled /etc/squid/conf.d/00-squid.conf line 3: http_access allow localnet
```

Fix the rules and re-apply. Remember each fragment is parsed **in isolation**: an `http_access`
referencing an `acl` defined in another `SquidConfigs` fails here. Keep the `acl` and its
`http_access` in the same object, or accept that only runtime (`cache.log`) will tell you.

### Apply hangs, then `context deadline exceeded` / `Timeout`

The validation Job never completed within 2 minutes. Look at the pod:

```bash
kubectl get pods -n healthcare-tools -l job-name --sort-by=.metadata.creationTimestamp
kubectl describe pod -n healthcare-tools <validate-pod>
```

Common causes:

- **Image pull** — the Squid image is large or the registry is slow; the pull counts against the
  2-minute budget. Pre-pull it on the nodes.
- **`ImagePullBackOff`** — the namespace lacks the `imagePullSecrets` needed for the instance image.
- **RBAC** — the manager's ServiceAccount cannot create Jobs. The generated ClusterRole does not
  include `batch/jobs`; see [Security](../development/security.md#rbac). Check with:

```bash
kubectl auth can-i create jobs \
  --as=system:serviceaccount:squid-operator:squid-operator -n healthcare-tools
```

- **ResourceQuota** — the namespace cannot admit another pod.

### `configmap <name> still contains the following rules <key>.conf`

The delete webhook refuses while the fragment's key is still in the ConfigMap. Normally the
controller removes the key first, so this means the controller could not run:

```bash
kubectl logs -n squid-operator deploy/squid-operator --tail=100
kubectl get cm <instance-name> -n <ns> -o jsonpath='{.data}' | jq keys
```

If the instance (and its ConfigMap) is already gone, nothing can remove the key — clear the
finalizer by hand:

```bash
kubectl patch squidconfigs <name> -n <ns> --type=merge -p '{"metadata":{"finalizers":[]}}'
```

### Apply fails with `x509` / `no endpoints available for service "…-webhook-service"`

The webhook is unreachable and both configurations are `failurePolicy: Fail`, so admission fails
closed.

```bash
kubectl get pods -n squid-operator
kubectl get certificate -A | grep squid          # cert-manager must report Ready
kubectl get validatingwebhookconfigurations -o yaml | grep -A3 caBundle
```

If `config/network-policy/` is applied, the calling namespace must be labelled `webhook: enabled`.

## SquidInstance problems

### `HEALTH: Error`

At least one managed resource failed. The reason is only in the manager log:

```bash
kubectl logs -n squid-operator deploy/squid-operator | grep -i squidinstance
kubectl describe squidinstance <name> -n <ns>
```

Frequent causes: missing RBAC for one of the five resource types (the generated ClusterRole has
gaps — `deployments` is granted under the wrong API group and `persistentvolumeclaims` is missing),
or an admission/quota rejection in the namespace.

```bash
kubectl auth can-i create deployments --as=system:serviceaccount:squid-operator:squid-operator -n <ns>
kubectl auth can-i create persistentvolumeclaims --as=system:serviceaccount:squid-operator:squid-operator -n <ns>
```

### Pods stuck `Pending`

Almost always the PVC. The operator requests `ReadWriteMany` at `10Gi`:

```bash
kubectl get pvc <instance-name> -n <ns>
kubectl describe pvc <instance-name> -n <ns>
```

`no persistent volumes available for this claim` or a provisioner error means the storage class does
not support `ReadWriteMany`. There is no way to change the access mode through the API — pick a
class that supports it (NFS, CephFS, EFS, Azure Files).

### Pods `CrashLoopBackOff`

Squid rejected the merged configuration at startup, or the image entrypoint failed:

```bash
kubectl logs -n <ns> <pod> --previous
kubectl exec -n <ns> deploy/<name> -- squid -k parse -f /etc/squid/squid.conf
```

Only whole-ConfigMap combinations are unvalidated, so this is usually a cross-fragment problem —
duplicate `acl` names, an `http_access` before its `acl`, or a directive the image does not support.
Fragments load alphabetically by key; rename with numeric prefixes to fix ordering.

### Pods never become `Ready`

Both probes are TCP checks on port `3128`. If your rules changed `http_port`, Squid no longer listens
where the probes look.

```bash
kubectl exec -n <ns> <pod> -- ss -ltnp | grep 3128
```

### New rules do not take effect

The rollout is driven by the ConfigMap watch and the `restartedAt` annotation:

```bash
kubectl get cm <instance-name> -n <ns> -o jsonpath='{.data}' | jq keys        # key present?
kubectl get squidconfigs -n <ns>                                              # STATE Merged?
kubectl rollout status deploy/<instance-name> -n <ns>
kubectl get deploy <instance-name> -n <ns> \
  -o jsonpath='{.spec.template.metadata.annotations.squid-operator\.kubernetes\.io/restartedAt}'
```

If the key is there but `restartedAt` is stale, the instance was not enqueued — force it:

```bash
kubectl annotate squidinstance <name> -n <ns> reconcile="$(date +%s)" --overwrite
```

Re-applying **byte-identical** rules is intentionally a no-op (reported internally as a duplicate),
so it produces no rollout.

### Instance deletion hangs

The finalizer is waiting on the controller. If the operator is not running, either start it or clear
the finalizer:

```bash
kubectl patch squidinstance <name> -n <ns> --type=merge -p '{"metadata":{"finalizers":[]}}'
```

Then delete the leftover Deployment, Service, ServiceAccount, ConfigMap and PVC — all share the
instance's name.

## Inspecting a running proxy

```bash
kubectl exec -n <ns> deploy/<name> -- ls /etc/squid/conf.d/
kubectl exec -n <ns> deploy/<name> -- cat /etc/squid/conf.d/<fragment>.conf
kubectl exec -n <ns> deploy/<name> -- tail -f /var/log/squid/access.log
kubectl exec -n <ns> deploy/<name> -- tail -f /var/log/squid/cache.log
kubectl exec -n <ns> deploy/<name> -- squidclient -h 127.0.0.1 -p 3128 mgr:info
```

`kubectl logs` on a Squid pod shows little: Squid writes to files on the PVC, not stdout.

## Proxy returns errors

| Response                       | Meaning                                                                 |
| ------------------------------ | ----------------------------------------------------------------------- |
| `403 Forbidden`                | Squid ACLs denied the request. Expected if you have `http_access deny all` and no matching allow. |
| `503 Service Unavailable`      | Squid could not reach upstream — DNS or egress network policy.           |
| Connection refused             | No ready endpoints; check pod readiness.                                 |
| Connection times out           | Service `EXTERNAL-IP` pending, or an ingress network policy blocks `3128`. |

```bash
kubectl get endpoints <instance-name> -n <ns>
kubectl run curl --rm -it --image=curlimages/curl --restart=Never -- \
  curl -sS -v -x http://<instance-name>.<ns>.svc.cluster.local:3128 -I https://example.com
```

## Squid messages worth knowing

| Message                                            | Action                                                       |
| -------------------------------------------------- | ------------------------------------------------------------ |
| `FATAL: Bungled ... line N`                        | Syntax error in a fragment at that line                      |
| `FATAL: Cannot open HTTP Port`                     | `http_port` conflicts, or not `3128`                         |
| `FATAL: Cannot open log file`                      | PVC not mounted or full — check `/var/log/squid`              |
| `WARNING: Forwarding loop detected`                | The proxy is resolving to itself; check `cache_peer` / ACLs   |
| `Your cache is running out of memory`              | `cache_mem` too large for the pod's memory                    |

## Escalating

Collect before opening an issue:

```bash
kubectl get squidinstance,squidconfigs -A -o yaml       > squid-resources.yaml
kubectl logs -n squid-operator deploy/squid-operator --tail=500 > operator.log
kubectl get events -A --sort-by='.lastTimestamp' | grep -i squid > squid-events.txt
```

## Next steps

- [Configuration](configuration.md) — rule semantics and limits
- [Monitoring](monitoring.md) — what is observable
- [Workflow](../architecture/workflow.md) — the reconcile paths behind these errors
