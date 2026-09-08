# Squid Operator

Kubernetes operator that deploys and manages [Squid](http://www.squid-cache.org/) forward-proxy
instances from custom resources, and lets the proxy configuration be composed from many small,
independently-owned fragments instead of one monolithic `squid.conf`.

Built with [Kubebuilder](https://book.kubebuilder.io/) v4 / controller-runtime.
API group: **`squid.cdk.clara.net/v1`**.

## Why

Squid configuration in a shared cluster tends to be owned by several teams: one wants an ACL for its
CI runners, another a cache directive, another a proxy peer. Editing a single ConfigMap by hand makes
those changes conflict and unreviewable.

The operator splits the problem in two:

| Kind            | Owns                                                                                        |
| --------------- | ------------------------------------------------------------------------------------------- |
| `SquidInstance` | The runtime: replica count, image, storage class. Creates Deployment, Service, ServiceAccount, ConfigMap and PVC. |
| `SquidConfigs`  | One fragment of `squid.conf`. Merged into the target instance's ConfigMap under its own key. |

Any number of `SquidConfigs` can target the same `SquidInstance`. Each owns exactly one ConfigMap key
(`<squidconfigs-name>.conf`), so fragments are added and removed independently and never overwrite
each other.

## Key features

- **Declarative Squid** — instances and configuration as Kubernetes custom resources.
- **Composable configuration** — one ConfigMap key per `SquidConfigs`, merged by the operator.
- **Admission-time syntax check** — the `SquidConfigs` validating webhook runs `squid -k parse` in a
  short-lived Job built from the target instance's own image, so a broken fragment is rejected
  instead of crash-looping the proxy.
- **Safe deletion** — finalizers clean up owned resources, and a `SquidConfigs` cannot be deleted
  while its rules are still present in the ConfigMap.
- **Automatic rollout** — a ConfigMap change re-annotates the Deployment pod template, triggering a
  rolling restart so Squid picks up the new rules.

## Quick start

1. [Install the operator](getting-started/installation.md) (needs cert-manager for the webhook certificates).
2. Create a `SquidInstance`.
3. Create one or more `SquidConfigs` annotated with `squid.ckd.clara.net/instance: <instance-name>`.
4. Reach the proxy on port `3128` through the instance `Service`.

See the [Quick Start](getting-started/quickstart.md) guide for the manifests.

## Documentation

### Getting started
- [Installation](getting-started/installation.md) — Helm chart, kustomize and prerequisites
- [Quick Start](getting-started/quickstart.md) — deploy your first instance
- [Basic Usage](user-guide/basic-usage.md) — day-to-day operations

### Architecture
- [Overview](architecture/overview.md) — what the operator reconciles
- [Components](architecture/components.md) — controllers, builders, webhooks
- [Workflow](architecture/workflow.md) — create / update / delete flows
- [Networking](architecture/networking.md) — proxy, webhook and metrics traffic

### User guide
- [Configuration](user-guide/configuration.md) — writing `SquidConfigs` rules
- [Scaling](user-guide/scaling.md) — replicas and the current autoscaling limits
- [Monitoring](user-guide/monitoring.md) — status, events and controller metrics
- [Troubleshooting](user-guide/troubleshooting.md) — common failures

### Reference
- [API Reference](reference/api.md) — field-by-field API description
- [CRDs](reference/crds.md) — generated CRD schemas and printer columns
- [Examples](reference/examples.md) — copy-pasteable manifests

### Development
- [Contributing](development/contributing.md) — environment, code generation, workflow
- [Testing](development/testing.md) — envtest and e2e suites
- [Release Process](development/release.md) — tags, images, changelog
- [Security](development/security.md) — RBAC, webhook certificates, hardening notes

## Current limitations

Known and intentional gaps, documented so you do not discover them in production:

- `spec.hpaSpec` on `SquidInstance` is accepted by the API but **not reconciled** — no HPA is created.
- No Ingress support. Only a `LoadBalancer` `Service` on port `3128` is created.
- Pod resource requests/limits, the container port, the PVC size (`10Gi`) and its access mode
  (`ReadWriteMany`) are hardcoded in `pkg/squid/`, not exposed in the spec.
- The annotation and finalizer domain is `squid.ckd.clara.net`, while the API group is
  `squid.cdk.clara.net`. The `ckd` spelling is what the code reads; fixing it would break existing
  resources.
- The generated ClusterRole does not grant `batch/jobs`, which the `SquidConfigs` validating webhook
  needs. See [Security](development/security.md#rbac).

## Project status

Actively developed. For the latest updates and releases, check the
[GitLab repository](https://git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator).

## Getting help

- Check the [troubleshooting guide](user-guide/troubleshooting.md)
- Open an issue in the [issue tracker](https://git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/-/issues)
- Contact the maintainers through the repository
