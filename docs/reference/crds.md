# Custom Resource Definitions (CRDs)

This document describes the Custom Resource Definitions (CRDs) used by the Squid Operator.

## SquidConfig

The `SquidConfig` CRD defines the configuration for Squid proxy instances.

### Example

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: example-config
  namespace: squid-proxy
spec:
  config: |
    http_port 3128
    acl SSL_ports port 443
    acl Safe_ports port 80
    acl Safe_ports port 443
    http_access allow localhost
    http_access deny all
```

### Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `spec.config` | string | Raw Squid configuration file content | Yes |

## SquidInstance

The `SquidInstance` CRD defines a Squid proxy instance deployment.

### Example

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: example-instance
  namespace: squid-proxy
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
  serviceType: ClusterIP
  servicePort: 3128
  monitoring:
    enabled: true
    port: 3129
```

### Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `spec.replicas` | integer | Number of Squid proxy instances to run | Yes |
| `spec.image.repository` | string | Container image repository | Yes |
| `spec.image.tag` | string | Container image tag | Yes |
| `spec.resources.requests.cpu` | string | CPU request (e.g., "100m") | No |
| `spec.resources.requests.memory` | string | Memory request (e.g., "128Mi") | No |
| `spec.resources.limits.cpu` | string | CPU limit (e.g., "500m") | No |
| `spec.resources.limits.memory` | string | Memory limit (e.g., "512Mi") | No |
| `spec.serviceType` | string | Kubernetes service type (ClusterIP, NodePort, LoadBalancer) | No |
| `spec.servicePort` | integer | Service port number | No |
| `spec.monitoring.enabled` | boolean | Enable monitoring | No |
| `spec.monitoring.port` | integer | Monitoring port number | No |

## Status Fields

### SquidInstance Status

| Field | Type | Description |
|-------|------|-------------|
| `status.readyReplicas` | integer | Number of ready replicas |
| `status.conditions` | array | Array of conditions describing the instance state |
| `status.conditions[].type` | string | Type of condition |
| `status.conditions[].status` | string | Status of condition (True, False, Unknown) |
| `status.conditions[].message` | string | Human-readable message |
| `status.conditions[].lastTransitionTime` | string | Last transition time |

## Usage Examples

### Basic Configuration

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: basic-config
spec:
  config: |
    http_port 3128
    acl SSL_ports port 443
    acl Safe_ports port 80
    acl Safe_ports port 443
    http_access allow localhost
    http_access deny all
```

### Production Configuration

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: production-instance
spec:
  replicas: 3
  image:
    repository: squid
    tag: latest
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 1Gi
  serviceType: LoadBalancer
  servicePort: 3128
  monitoring:
    enabled: true
    port: 3129
```

## Validation

The CRDs include validation rules to ensure correct configuration:

- Required fields must be specified
- Resource limits must be greater than or equal to requests
- Port numbers must be within valid ranges
- Service types must be one of: ClusterIP, NodePort, LoadBalancer

## Next Steps

- [API Reference](../reference/api.md) - Detailed API documentation
- [Basic Usage](../user-guide/basic-usage.md) - Getting started guide
- [Configuration Guide](../user-guide/configuration.md) - Advanced configuration options
