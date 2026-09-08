# Installation

The operator ships two installation paths, both from this repository: the Helm chart in
`deploy/squid-operator/`, or the kustomize bases in `config/` through the Makefile. There is no
public Helm repository or published bundle.

## Prerequisites

- A Kubernetes cluster (v1.25+ recommended; the envtest suites target 1.29)
- `kubectl` with cluster-admin rights for the install (CRDs, ClusterRole, webhook configurations)
- [cert-manager](https://cert-manager.io/) — issues the webhook serving certificate
  (`squid-webhook-server-cert`) and injects the CA bundle into the webhook configurations. Without
  it the manager pod stays unready and admission calls fail.
- A `StorageClass` that supports `ReadWriteMany`, for the per-instance log volume
- For the kustomize path: `make`, Go 1.22 and Docker (tool binaries are downloaded into `bin/`)

## Install with Helm

```bash
helm upgrade --install squid-operator deploy/squid-operator \
  --namespace squid-operator --create-namespace \
  --set image.tag=<tag>
```

Chart defaults worth reviewing in `deploy/squid-operator/values.yaml`:

| Value                   | Default                                                                                                | Note                                             |
| ----------------------- | ------------------------------------------------------------------------------------------------------ | ------------------------------------------------ |
| `image.repository`      | `registry.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/squid-operator` | Internal registry                                |
| `image.tag`             | `""` (falls back to `Chart.appVersion`)                                                                | Set explicitly for reproducible installs         |
| `imagePullSecrets`      | `[{name: regcred-gitlab}]`                                                                             | Must exist in the release namespace              |
| `replicaCount`          | `1`                                                                                                    | Leader election is enabled in the Deployment args |
| `autoscaling.enabled`   | `false`                                                                                                | HPA for the **operator**, not for Squid instances |

The chart installs the CRDs (`templates/crds/`), RBAC, the webhook service and configurations, the
cert-manager `Certificate`, and the manager Deployment. The manager is started with
`--metrics-bind-address=0`, i.e. **metrics are disabled by default** in the chart — see
[Monitoring](../user-guide/monitoring.md).

## Install with kustomize / make

```bash
# 1. Build and push the manager image
make docker-build docker-push IMG=<registry>/squid-operator:<tag>

# 2. CRDs only
make install

# 3. Controller, RBAC, webhooks, cert-manager Certificate
make deploy IMG=<registry>/squid-operator:<tag>
```

`make deploy` applies `config/default`, which deploys into the `squid-operator-system` namespace.

### Single-file bundle

```bash
make build-installer IMG=<registry>/squid-operator:<tag>   # writes dist/install.yaml
kubectl apply -f dist/install.yaml
```

## Verify the installation

```bash
kubectl get crd | grep squid.cdk.clara.net
```

```
squidconfigs.squid.cdk.clara.net    2024-01-01T00:00:00Z
squidinstances.squid.cdk.clara.net  2024-01-01T00:00:00Z
```

```bash
kubectl get pods -n squid-operator          # Helm
kubectl get pods -n squid-operator-system   # make deploy
```

```
NAME                                        READY   STATUS    RESTARTS   AGE
squid-operator-7b6cf7c9c4-2xq8v             1/1     Running   0          1m
```

Check that the webhooks got their CA bundle injected:

```bash
kubectl get mutatingwebhookconfigurations,validatingwebhookconfigurations | grep squid
```

If the manager logs `unable to start manager` or the pod never becomes ready, the serving
certificate is usually missing — confirm cert-manager is installed and the `Certificate` is `Ready`.

## Uninstall

```bash
kubectl delete squidconfigs --all -A     # before deleting the operator: finalizers need the webhook
kubectl delete squidinstances --all -A
helm uninstall squid-operator -n squid-operator   # or: make undeploy && make uninstall
```

!!! warning
    Delete the custom resources **before** removing the operator. Both kinds carry the
    `squid.ckd.clara.net/finalizer` finalizer, and `SquidConfigs` deletion is gated by a validating
    webhook. With the manager gone, deletions hang and the finalizers have to be removed by hand:
    `kubectl patch squidconfigs <name> -p '{"metadata":{"finalizers":[]}}' --type=merge`.

## Next steps

- [Quick Start](quickstart.md) — deploy your first Squid instance
- [Configuration](../user-guide/configuration.md) — write and merge Squid rules
