# Components

This document describes the main components of the Squid Operator and their interactions.

## Operator Components

### Controller

The controller is the main component that manages the lifecycle of Squid resources.

```go
type Controller struct {
    client.Client
    Scheme *runtime.Scheme
    Log    logr.Logger
}
```

Responsibilities:
- Watching for changes to SquidConfig and SquidInstance resources
- Reconciling desired state with actual state
- Managing the lifecycle of Squid proxy instances
- Handling configuration updates

### Reconciler

The reconciler handles the reconciliation logic for each resource type.

```go
type SquidInstanceReconciler struct {
    client.Client
    Scheme *runtime.Scheme
    Log    logr.Logger
}
```

Responsibilities:
- Creating and updating Kubernetes resources
- Managing configuration updates
- Handling scaling operations
- Monitoring resource status

## Managed Resources

### SquidConfig

The SquidConfig resource defines the configuration for Squid proxy instances.

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: example-config
spec:
  config: |
    http_port 3128
    acl SSL_ports port 443
    acl Safe_ports port 80
    acl Safe_ports port 443
    http_access allow localhost
    http_access deny all
```

Components:
- Configuration template
- Validation rules
- Default values

### SquidInstance

The SquidInstance resource manages the deployment of Squid proxy instances.

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: example-instance
spec:
  replicas: 1
  image:
    repository: squid
    tag: latest
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
```

Components:
- Deployment specification
- Service configuration
- Resource requirements
- Monitoring setup

## Supporting Components

### Metrics

Metrics collection for monitoring Squid instances.

```go
type MetricsCollector struct {
    client.Client
    Log logr.Logger
}
```

Metrics:
- Request count
- Cache hit ratio
- Response times
- Resource usage

### Webhooks

Validation and mutation webhooks for custom resources.

```go
type ValidatingWebhook struct {
    client.Client
    Log logr.Logger
}
```

Webhooks:
- Validation webhook for SquidConfig
  - Validates configuration syntax using a validation job
  - Checks for required instance annotation
  - Ensures configuration is valid before creation/update
  - Prevents deletion if config is still in use
- Validation webhook for SquidInstance
  - Validates image configuration
  - Ensures storage class is specified
  - Validates resource requirements
- Defaulting webhook for resources
  - Sets default image (ubuntu/squid)
  - Sets default tag
  - Sets default storage class (standard)

### SquidConfig Validation

The SquidConfig webhook performs the following validations:

1. **Pre-Creation/Update Validation**
   - Creates a validation job to test configuration
   - Ensures required instance annotation is present
   - Validates configuration syntax using squid -k parse
   - Timeout of 2 minutes for validation

2. **Pre-Deletion Validation**
   - Checks if configuration is still in use
   - Verifies configmap references
   - Prevents deletion if configuration is active

Example validation error:
```yaml
status:
  conditions:
  - type: Valid
    status: "False"
    reason: ValidationFailed
    message: "Configuration validation failed: squid -k parse returned error"
```

## Component Interactions

### Resource Lifecycle

1. User creates SquidConfig
2. Controller watches for changes
3. Reconciler processes the configuration
4. SquidInstance is created or updated
5. Status is updated

### Configuration Flow

1. SquidConfig is created
2. Configuration is validated
3. ConfigMap is created
4. SquidInstance is updated
5. Pods are restarted if needed

### Monitoring Flow

1. Metrics are collected
2. Prometheus scrapes metrics
3. Alerts are generated
4. Status is updated

## Component Dependencies

### Required Components

- Kubernetes API Server
- Controller Runtime
- Metrics Server
- Prometheus Operator

### Optional Components

- Grafana for visualization
- AlertManager for notifications
- Cert-Manager for TLS

## Next Steps

- [Architecture](../architecture.md) - Overall architecture
- [Networking](../architecture/networking.md) - Network components
- [Monitoring](../architecture/monitoring.md) - Monitoring components
