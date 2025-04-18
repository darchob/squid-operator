# Basic Usage Guide

This guide covers the basic usage scenarios for the Squid Operator.

## Basic Configuration

### Creating a Simple Squid Proxy

1. Create a basic SquidConfig:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: basic-squid-config
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

2. Create a basic SquidInstance:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: basic-squid
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

## Common Use Cases

### Basic Proxy Setup

1. Deploy the configuration:

```bash
kubectl apply -f basic-squid-config.yaml
kubectl apply -f basic-squid-instance.yaml
```

2. Verify the deployment:

```bash
kubectl get pods -n squid-proxy
kubectl get svc -n squid-proxy
```

### Scaling the Proxy

To scale the number of Squid proxy instances:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: basic-squid
  namespace: squid-proxy
spec:
  replicas: 3  # Increase the number of replicas
  # ... rest of the configuration
```

### Resource Management

Adjust resource limits and requests:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: basic-squid
  namespace: squid-proxy
spec:
  # ... other configuration
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 1Gi
```

## Service Configuration

### Service Types

Configure different service types:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: basic-squid
  namespace: squid-proxy
spec:
  # ... other configuration
  serviceType: LoadBalancer  # Options: ClusterIP, NodePort, LoadBalancer
  servicePort: 3128
```

## Monitoring and Logging

### Basic Monitoring

Enable basic monitoring:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: basic-squid
  namespace: squid-proxy
spec:
  # ... other configuration
  monitoring:
    enabled: true
    port: 3129
```

### Logging Configuration

Configure basic logging:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: basic-squid-config
  namespace: squid-proxy
spec:
  config: |
    # ... other configuration
    access_log /var/log/squid/access.log
    cache_log /var/log/squid/cache.log
    logformat squid %ts.%03tu %6tr %>a %Ss/%03>Hs %<st %rm %ru %un %Sh/%<A %mt
```

## Troubleshooting

### Common Issues

1. **Pod Not Starting**
   - Check resource limits
   - Verify configuration syntax
   - Check logs: `kubectl logs -n squid-proxy <pod-name>`

2. **Service Not Accessible**
   - Verify service type
   - Check network policies
   - Test connectivity: `kubectl exec -it <pod-name> -n squid-proxy -- curl localhost:3128`

3. **Configuration Issues**
   - Validate SquidConfig syntax
   - Check for missing required fields
   - Verify namespace permissions

## Next Steps

- [Advanced Configuration](../user-guide/configuration.md) - Learn about advanced configuration options
- [Security Guide](../development/security.md) - Security best practices
- [API Reference](../reference/api.md) - Detailed API documentation
