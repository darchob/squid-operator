# API Reference

This document provides detailed information about the Squid Operator's API resources.

## SquidConfig

The `SquidConfig` resource defines the configuration for Squid proxy instances.

### Example

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: my-squid-config
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
| config | string | Squid configuration content | Yes |

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

### Fields

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| replicas | integer | Number of Squid proxy replicas | Yes |
| image.repository | string | Squid image repository | Yes |
| image.tag | string | Squid image tag | Yes |
| resources | [ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#resourcerequirements-v1-core) | Resource requirements for the Squid proxy | No |
| serviceType | string | Kubernetes service type (ClusterIP, NodePort, LoadBalancer) | No |
| servicePort | integer | Port number for the Squid proxy service | No |

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
