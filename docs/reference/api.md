# API Reference

This document provides detailed information about the Squid Operator's API resources.

## SquidConfig

The `SquidConfig` resource defines the configuration for Squid proxy instances.

### Example

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: example-config
  annotations:
    squid.ckd.clara.net/instance: my-squid-instance
spec:
  config: |
    http_port 3128
    acl SSL_ports port 443
    acl Safe_ports port 80
    acl Safe_ports port 443
    http_access allow localhost
    http_access deny all
```

### Validation

The SquidConfig resource is validated by a webhook that:

1. **Pre-Creation/Update Validation**
   - Creates a validation job to test configuration
   - Ensures required instance annotation is present
   - Validates configuration syntax using squid -k parse
   - Timeout of 2 minutes for validation

2. **Pre-Deletion Validation**
   - Checks if configuration is still in use
   - Verifies configmap references
   - Prevents deletion if configuration is active

### Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `spec.config` | string | Squid configuration in squid.conf format | Yes |
| `metadata.annotations.squid.ckd.clara.net/instance` | string | Reference to the SquidInstance using this config | Yes |

### Validation Rules

1. **Port Configuration**
   - Port numbers must be between 1 and 65535
   - Each port must be unique
   - Required ports must be defined

2. **ACL Configuration**
   - ACL names must be unique
   - ACL types must be valid
   - ACL values must be properly formatted

3. **Access Control**
   - Must have at least one access rule
   - Rules must reference valid ACLs
   - Default deny rule should be present

4. **Resource Limits**
   - Cache size must be positive
   - Memory limits must be reasonable
   - Connection limits must be positive

## SquidInstance

The `SquidInstance` resource defines a Squid proxy instance deployment.

### Example

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: my-squid
  namespace: squid-proxy
spec:
  replicas: 1
  image:
    repository: ubuntu/squid
    tag: latest
  storageClassName: standard
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
```

### Fields

| Field | Type | Description | Required | Default |
|-------|------|-------------|----------|---------|
| `spec.replicas` | integer | Number of Squid proxy replicas | Yes | - |
| `spec.image.repository` | string | Squid image repository | No | ubuntu/squid |
| `spec.image.tag` | string | Squid image tag | No | latest |
| `spec.storageClassName` | string | Storage class for persistent volume | No | standard |
| `spec.resources` | [ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core) | Resource requirements for the Squid proxy | No | - |
| serviceType | string | Kubernetes service type (ClusterIP, NodePort, LoadBalancer) | No | - |
| servicePort | integer | Port number for the Squid proxy service | No | - |

## Status Conditions

The operator uses standard Kubernetes conditions to report the status of resources:

- `Ready`: The resource is ready and fully operational
- `Progressing`: The resource is being created or updated
- `Degraded`: The resource is in a degraded state
- `Stalled`: The resource is stalled and cannot progress

## Events

The operator emits Kubernetes events for important state changes:

- `Created`: Resource was created
- `Updated`: Resource was updated
- `Deleted`: Resource was deleted
- `Error`: An error occurred
- `Warning`: A warning condition was detected
