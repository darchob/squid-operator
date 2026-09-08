# Configuration

All Squid configuration goes through `SquidConfigs.spec.rules`. The operator does not translate
anything: the string is copied verbatim into a ConfigMap key and Squid reads it from
`/etc/squid/conf.d/`. Anything valid in a `squid.conf` include file is valid here.

## How a fragment is wired

```yaml
apiVersion: squid.cdk.clara.net/v1
kind: SquidConfigs
metadata:
  name: cache-tuning                          # (1) becomes the ConfigMap key: cache-tuning.conf
  namespace: healthcare-tools                 # (2) same namespace as the instance
  annotations:
    squid.ckd.clara.net/instance: squid-sample # (3) target SquidInstance — required
spec:
  rules: |                                     # (4) raw squid.conf directives
    cache_mem 256 MB
    maximum_object_size 32 MB
```

1. The object name is the ConfigMap key, so it appears in `squid -k parse` errors. Name fragments
   after what they do.
2. Cross-namespace targeting is not supported.
3. `squid.ckd.clara.net/instance` — note the `ckd` spelling, which does not match the API group
   `squid.cdk.clara.net`. It is what the code reads. Missing annotation ⇒ admission rejected.
4. `spec.rules` is a plain string and the only field in the spec.

Load order across fragments is Squid's include order for `/etc/squid/conf.d/*` — alphabetical by
key. If order matters (an `acl` must be defined before the `http_access` that uses it), either keep
both in one fragment or prefix names: `10-acls`, `20-access`.

## Access control

```yaml
apiVersion: squid.cdk.clara.net/v1
kind: SquidConfigs
metadata:
  name: 10-acls
  namespace: healthcare-tools
  annotations:
    squid.ckd.clara.net/instance: squid-sample
spec:
  rules: |
    acl localnet src 10.0.0.0/8
    acl localnet src 172.16.0.0/12
    acl localnet src 192.168.0.0/16
    acl SSL_ports port 443
    acl Safe_ports port 80
    acl Safe_ports port 443
    acl CONNECT method CONNECT
```

```yaml
apiVersion: squid.cdk.clara.net/v1
kind: SquidConfigs
metadata:
  name: 20-access
  namespace: healthcare-tools
  annotations:
    squid.ckd.clara.net/instance: squid-sample
spec:
  rules: |
    http_access deny !Safe_ports
    http_access deny CONNECT !SSL_ports
    http_access allow localnet
    http_access deny all
```

In Kubernetes, "localnet" is the pod CIDR — check your cluster's ranges rather than copying the
RFC1918 defaults blindly.

## Caching

The PVC is mounted at `/var/log/squid` only; there is **no cache volume**. A `cache_dir` pointing
anywhere else writes to the container's ephemeral filesystem and is lost on restart. Prefer
memory-only caching:

```yaml
spec:
  rules: |
    cache_mem 512 MB
    maximum_object_size_in_memory 512 KB
    cache_dir null /tmp    # requires a Squid build with the null store
```

If you need a real disk cache, add a volume in `pkg/squid/workloads.go` — it is not exposed in the
API today.

## Logging and rotation

Logs land on the shared `ReadWriteMany` PVC (`10Gi`, hardcoded). Rotate them, or the volume fills up
and Squid stops:

```yaml
spec:
  rules: |
    access_log /var/log/squid/access.log
    cache_log /var/log/squid/cache.log
    logfile_rotate 3
    logformat squid %ts.%03tu %6tr %>a %Ss/%03>Hs %<st %rm %ru %un %Sh/%<A %mt
```

All replicas share the volume, so they append to the same files.

## Authentication

Squid authentication helpers need credential files inside the container. The `SquidInstance` spec has
no field for extra volumes or secrets, so `auth_param` rules only work if the credentials are baked
into your image, or if you patch the Deployment out of band (the operator overwrites the pod
template on every reconcile, so patches do not survive).

```yaml
spec:
  rules: |
    auth_param basic program /usr/lib/squid/basic_ncsa_auth /etc/squid/passwd
    auth_param basic realm proxy
    acl authenticated proxy_auth REQUIRED
    http_access allow authenticated
```

Note that `squid -k parse` in the validation webhook only checks syntax — it does not verify that the
helper binary or the password file exists.

## Port considerations

The container exposes `3128` and both probes are TCP checks against it. `http_port` in your rules
must keep Squid listening on `3128`, otherwise the probes fail and the pods never become ready:

```yaml
spec:
  rules: |
    http_port 3128     # do not change
```

Additional listeners on other ports work inside the pod but are not exposed by the `Service`.

## Validation

Every create and update is checked at admission time:

1. The webhook reads the target `SquidInstance` to get its image.
2. It creates a `Job` (`validate-<name>-*`) that writes your rules to
   `/etc/squid/conf.d/00-squid.conf` and runs `squid -k parse -f` on them.
3. It polls for up to 2 minutes. Non-zero exit ⇒ the object is rejected.

Consequences:

- Validation uses the **same Squid build** that will run the rules, so directives unsupported by
  your image are caught.
- The 2-minute budget includes the image pull.
- Only the fragment being applied is parsed, in isolation. An `http_access` referring to an `acl`
  defined in another fragment parses fine here but can still fail at runtime — check `cache.log`
  after a rollout.
- One `Job` per apply. They are not cleaned up by the operator.

## What is not configurable

Hardcoded in `pkg/squid/`, not exposed in the `SquidInstance` API:

| Setting                     | Value                                    |
| --------------------------- | ---------------------------------------- |
| Container / Service port    | `3128`                                   |
| Service type                | `LoadBalancer`                           |
| PVC size / access mode      | `10Gi` / `ReadWriteMany`                 |
| Log mount path              | `/var/log/squid`                         |
| Config mount path           | `/etc/squid/conf.d/`                     |
| `imagePullPolicy`           | `Always`                                 |
| Pod resource requests/limits | none set                                |
| Probes                      | TCP on `3128`, default timings           |
| Security context            | none set on the Squid pods               |
| Extra volumes, env, secrets, `imagePullSecrets`, affinity, tolerations | not supported |

`spec.hpaSpec` is accepted by the API but not reconciled — see [Scaling](scaling.md).

Changing any of these means changing the builders in `pkg/squid/` and, for new fields, the types in
`api/v1/` plus `make manifests`. See [Contributing](../development/contributing.md).

## Best practices

1. **One concern per fragment.** Small objects, meaningful names, independent lifecycles.
2. **Order deliberately.** Numeric name prefixes when definitions must precede use.
3. **Keep ACLs and their `http_access` together** if you cannot guarantee ordering.
4. **End with `http_access deny all`** in the highest-numbered fragment.
5. **Rotate logs.** The volume is finite and shared.
6. **Review `cache.log` after a rollout** — cross-fragment problems only surface at runtime.

## Next steps

- [Examples](../reference/examples.md) — complete manifests
- [Basic Usage](basic-usage.md) — applying and inspecting
- [Troubleshooting](troubleshooting.md) — validation and rollout failures
