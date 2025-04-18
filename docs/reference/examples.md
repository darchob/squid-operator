# Examples

This section provides practical examples of using the Squid Operator.

## Basic Configuration

### Simple SquidConfig

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
    acl CONNECT method CONNECT
    http_access deny !Safe_ports
    http_access deny CONNECT !SSL_ports
    http_access allow localhost manager
    http_access deny manager
    http_access allow localhost
    http_access deny all
```

### Basic SquidInstance

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

## Advanced Configuration

### SquidConfig with Authentication

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: auth-squid-config
  namespace: squid-proxy
spec:
  config: |
    http_port 3128
    auth_param basic program /usr/lib/squid/basic_ncsa_auth /etc/squid/passwords
    auth_param basic realm proxy
    acl authenticated proxy_auth REQUIRED
    http_access allow authenticated
    http_access deny all
```

### SquidInstance with Authentication

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: auth-squid
  namespace: squid-proxy
spec:
  replicas: 2
  image:
    repository: squid
    tag: latest
  authentication:
    enabled: true
    secretName: squid-auth-secret
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 1Gi
```

## Production Configuration

### High-Availability SquidConfig

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: ha-squid-config
  namespace: squid-proxy
spec:
  config: |
    http_port 3128
    cache_peer squid-1.squid-proxy.svc.cluster.local parent 3128 0 no-query
    cache_peer squid-2.squid-proxy.svc.cluster.local parent 3128 0 no-query
    cache_peer_access squid-1 allow all
    cache_peer_access squid-2 allow all
    never_direct allow all
```

### Production SquidInstance

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: production-squid
  namespace: squid-proxy
spec:
  replicas: 3
  image:
    repository: squid
    tag: latest
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 4Gi
  persistence:
    enabled: true
    size: 10Gi
    storageClass: standard
  security:
    tls:
      enabled: true
      secretName: squid-tls-secret
    networkPolicy:
      enabled: true
      allowedIPs:
        - 10.0.0.0/8
        - 172.16.0.0/12
        - 192.168.0.0/16
```

## Workflow Example

Here's a complete example showing the workflow between SquidConfig and SquidInstance:

1. Create the SquidConfig:
```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: workflow-config
  namespace: squid-proxy
spec:
  config: |
    http_port 3128
    acl SSL_ports port 443
    acl Safe_ports port 80
    acl Safe_ports port 443
    http_access allow localhost
    http_access allow SSL_ports
    http_access deny all
```

2. Create the SquidInstance referencing the config:
```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: workflow-squid
  namespace: squid-proxy
spec:
  replicas: 2
  image:
    repository: squid
    tag: latest
  configRef:
    name: workflow-config
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 1Gi
```

3. Update the SquidConfig:
```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: workflow-config
  namespace: squid-proxy
spec:
  config: |
    http_port 3128
    acl SSL_ports port 443
    acl Safe_ports port 80
    acl Safe_ports port 443
    acl CONNECT method CONNECT
    http_access deny !Safe_ports
    http_access deny CONNECT !SSL_ports
    http_access allow localhost
    http_access allow SSL_ports
    http_access deny all
```

The operator will automatically:
1. Update the ConfigMap with the new configuration
2. Trigger reconciliation of the SquidInstance
3. Reload the Squid configuration in the running pods

## Common Use Cases

### 1. Development Environment

```yaml
# SquidConfig for development
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: dev-config
  namespace: squid-proxy
spec:
  config: |
    http_port 3128
    acl localnet src 10.0.0.0/8
    http_access allow localnet
    http_access deny all

# SquidInstance for development
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: dev-squid
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

### 2. Staging Environment

```yaml
# SquidConfig for staging
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: staging-config
  namespace: squid-proxy
spec:
  config: |
    http_port 3128
    acl localnet src 10.0.0.0/8
    acl SSL_ports port 443
    acl Safe_ports port 80
    acl Safe_ports port 443
    http_access allow localnet
    http_access allow SSL_ports
    http_access deny all

# SquidInstance for staging
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: staging-squid
  namespace: squid-proxy
spec:
  replicas: 2
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
```

### 3. Production Environment

```yaml
# SquidConfig for production
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: prod-config
  namespace: squid-proxy
spec:
  config: |
    http_port 3128
    acl localnet src 10.0.0.0/8
    acl SSL_ports port 443
    acl Safe_ports port 80
    acl Safe_ports port 443
    acl CONNECT method CONNECT
    http_access deny !Safe_ports
    http_access deny CONNECT !SSL_ports
    http_access allow localnet
    http_access allow SSL_ports
    http_access deny all

# SquidInstance for production
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: prod-squid
  namespace: squid-proxy
spec:
  replicas: 3
  image:
    repository: squid
    tag: latest
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 4Gi
  persistence:
    enabled: true
    size: 10Gi
    storageClass: standard
  security:
    tls:
      enabled: true
      secretName: squid-tls-secret
    networkPolicy:
      enabled: true
      allowedIPs:
        - 10.0.0.0/8
        - 172.16.0.0/12
        - 192.168.0.0/16
```

## Next Steps

- [Configuration Guide](../user-guide/configuration.md) - Learn more about configuration options
- [API Reference](../reference/api.md) - Detailed API documentation
- [Architecture](../architecture/overview.md) - Understand the operator's architecture
