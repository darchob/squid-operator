# Components

Code-level map of the operator. Paths are relative to the repository root.

## Manager — `cmd/main.go`

Builds the controller-runtime manager and registers everything:

- Both reconcilers, each with its own `EventRecorder` (`squidInstances`, `squidConfigs`)
- Both webhook pairs, from `api/v1`, unless `ENABLE_WEBHOOKS=false`
- `healthz` / `readyz` probes on `--health-probe-bind-address` (default `:8081`)
- Metrics on `--metrics-bind-address` (default `:8080`), optionally TLS with `--metrics-secure`
- Leader election under `--leader-elect`, lease ID `aa7cb171.cdk.clara.net`
- HTTP/2 disabled unless `--enable-http2` (mitigates the Rapid Reset CVEs)

## Controllers — `internal/controller/`

### SquidInstanceReconciler

`squidinstance_controller.go`.

```go
type SquidInstanceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}
```

Dispatches on object state: no finalizer → `handlingInit`, `DeletionTimestamp != nil` →
`handlingDeletion`, otherwise `handlingReconciliation`.

The set of resources it manages is a package-level table pairing an empty object with its builder:

```go
var managedResources = []ManagedResource{
	{Type: &appsv1.Deployment{},             Generate: func(si *squidv1.SquidInstance) client.Object { return squid.NewDeployment(si) }},
	{Type: &corev1.ConfigMap{},              Generate: /* squid.ConfigMap */},
	{Type: &corev1.ServiceAccount{},         Generate: /* squid.NewServiceAccount */},
	{Type: &corev1.Service{},                Generate: /* squid.Service */},
	{Type: &corev1.PersistentVolumeClaim{},  Generate: /* squid.PersistentVolumeClaim */},
}
```

Adding a managed resource means adding a builder in `pkg/squid/` and one entry here.

Watches: `Owns` Deployment, ServiceAccount, ConfigMap, PVC, plus an explicit `Watches` on ConfigMaps
that maps a ConfigMap back to the SquidInstance of the same name.

### SquidConfigsReconciler

`squidconfigs_controller.go`.

Adds the finalizer on first sight, then merges. States are an internal enum surfaced as
`status.state`: `Merged`, `Deletion`, `Applying`, `Failed`.

### Shared helpers — `squid_common.go`

- `finalizerName = "squid.ckd.clara.net/finalizer"`
- `handleConfigMap(obj, configs, truncate)` — the merge itself: computes the key
  `<configs.Name>.conf`, deletes it when `truncate` is set (deletion path), otherwise writes
  `spec.rules`; returns a duplicate error if the stored content is already identical.

## Resource builders — `pkg/squid/`

Pure functions: given a `*SquidInstance`, return the desired Kubernetes object. No client, no
context — this is the layer to change what gets deployed.

| File            | Provides                                                                                       |
| --------------- | ---------------------------------------------------------------------------------------------- |
| `defaults.go`   | `Labels`, `Annotations`, `ObjectMeta` — common metadata and owner reference for every child     |
| `workloads.go`  | `NewDeployment` — container `<instance-name>`, image `repository:tag`, `imagePullPolicy: Always`, port `3128` (`http`), TCP liveness + readiness on that port, volumes `configs` (ConfigMap → `/etc/squid/conf.d/`) and `logs` (PVC → `/var/log/squid`) |
| `networking.go` | `Service` — type `LoadBalancer`, port `3128`, selector `app.kubernetes.io/name: <instance-name>` |
| `storages.go`   | `ConfigMap` — seeded with `squid-init.conf`; `PersistentVolumeClaim` — `ReadWriteMany`, `10Gi`   |
| `sa.go`         | `NewServiceAccount`                                                                             |
| `errors.go`     | `IsRulesExistError`                                                                             |

Labels applied to every child object:

```yaml
app.kubernetes.io/name: <instance-name>
app.kubernetes.io/version: <image-tag>       # SquidInstance only
app.kubernetes.io/part-of: squid-<instance-name>
app.kubernetes.io/managed-by: squid-operator
```

No pod resource requests or limits are set — the pods land in the namespace's default
`LimitRange`/`ResourceQuota`, or unbounded if none exists.

## Webhooks — `api/v1/`

Registered by `cmd/main.go` through `SetupWebhookWithManager`, using the legacy
`webhook.Defaulter` / `webhook.Validator` interfaces.

### SquidInstance — `squidinstance_webhook.go`

- **Mutating** (`/mutate-squid-cdk-clara-net-v1-squidinstance`): sets
  `DefaultImage = "ubuntu/squid"`, `DefaultTag = "latest"`, `DefaultStorageClass = "standard"`.
- **Validating** (`/validate-squid-cdk-clara-net-v1-squidinstance`): logs only, returns no error.

!!! warning
    The image default guards on `r.Spec.Image == nil` and then writes through that same nil pointer.
    The CRD marks `image` required, so the API server rejects the object before the webhook is
    reached and the path is unreachable in practice — but it is a latent nil dereference. Always set
    `spec.image` explicitly.

### SquidConfigs — `squidconfigs_webhook.go`

Holds a package-level `client.Client` captured at setup, so the webhook can read cluster state.

- **Mutating**: no-op.
- **Validating create/update**:
    1. Require the `squid.ckd.clara.net/instance` annotation.
    2. Fetch the named `SquidInstance` in the same namespace.
    3. Create a `Job` (`generateName: validate-<name>-`) running the instance's image with
       `echo '<rules>' > /etc/squid/conf.d/00-squid.conf && squid -k parse -f /etc/squid/conf.d/00-squid.conf`,
       `restartPolicy: Never`.
    4. Poll every 5s until success or failure, under a 2-minute context — image pull time counts
       against that budget.
- **Validating delete**: read the target ConfigMap; refuse while `<name>.conf` is still present. The
  controller's deletion path removes the key, so the sequencing matters.

The validation Jobs are not garbage-collected by the operator: set `ttlSecondsAfterFinished` at the
cluster level or clean them up out of band.

## Unused scaffolding — `internal/webhook/v1/`

Kubebuilder v4 `CustomDefaulter` / `CustomValidator` versions of both webhooks. They are **not
registered** in `cmd/main.go` and their `Default`/`Validate*` bodies are still `TODO(user)`. The
migration from `api/v1` is unfinished; changing behaviour means editing `api/v1`.

## Deployment manifests

| Path                              | Contents                                                                     |
| --------------------------------- | ---------------------------------------------------------------------------- |
| `config/crd/bases/`               | Generated CRDs — never hand-edited, run `make manifests`                     |
| `config/rbac/`                    | Generated ClusterRole from `// +kubebuilder:rbac` markers, plus editor/viewer/admin roles |
| `config/webhook/`                 | Webhook `Service` and generated configurations                               |
| `config/certmanager/`             | `Issuer` and `Certificate` for the webhook and metrics serving certs         |
| `config/network-policy/`          | Ingress policies for the metrics (`8443`) and webhook (`443`) ports          |
| `config/prometheus/`              | `ServiceMonitor` for the controller metrics                                  |
| `deploy/squid-operator/`          | Helm chart: CRDs, RBAC, webhooks, cert-manager `Certificate`, Deployment     |

## Next steps

- [Workflow](workflow.md) — how these components interact per operation
- [Networking](networking.md) — ports and policies
- [API Reference](../reference/api.md) — field semantics
