# Configuration Guide

This guide covers advanced configuration options for the Squid Operator.

## Advanced SquidConfig Options

### Cache Configuration

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: cache-config
spec:
  config: |
    # Cache configuration
    cache_dir ufs /var/spool/squid 10000 16 256
    maximum_object_size 32 MB
    cache_mem 256 MB
    cache_swap_low 90
    cache_swap_high 95
```

### Access Control Lists (ACLs)

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: acl-config
spec:
  config: |
    # ACL definitions
    acl localnet src 10.0.0.0/8
    acl localnet src 172.16.0.0/12
    acl localnet src 192.168.0.0/16
    acl SSL_ports port 443
    acl Safe_ports port 80
    acl Safe_ports port 443
    acl CONNECT method CONNECT
```

### Authentication Configuration

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: auth-config
spec:
  config: |
    # Authentication configuration
    auth_param basic program /usr/lib/squid/basic_ncsa_auth /etc/squid/passwd
    auth_param basic realm Squid proxy-caching web server
    auth_param basic credentialsttl 2 hours
    acl authenticated proxy_auth REQUIRED
    http_access allow authenticated
```

## Advanced SquidInstance Options

### Resource Configuration

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: resource-config
spec:
  replicas: 3
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 2000m
      memory: 2Gi
  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
          - key: app
            operator: In
            values:
            - squid
        topologyKey: kubernetes.io/hostname
```

### Monitoring Configuration

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: monitoring-config
spec:
  monitoring:
    enabled: true
    port: 3129
    metrics:
      enabled: true
      path: /metrics
    logging:
      level: debug
      format: json
```

### Security Configuration

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: security-config
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    runAsGroup: 1000
    fsGroup: 1000
  podSecurityContext:
    seccompProfile:
      type: RuntimeDefault
```

## Environment Variables

### Common Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SQUID_CONFIG_FILE` | Path to Squid configuration file | `/etc/squid/squid.conf` |
| `SQUID_CACHE_DIR` | Path to cache directory | `/var/spool/squid` |
| `SQUID_LOG_DIR` | Path to log directory | `/var/log/squid` |

### Custom Environment Variables

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: env-config
spec:
  env:
    - name: SQUID_CACHE_SIZE
      value: "10000"
    - name: SQUID_MAX_OBJECT_SIZE
      value: "32 MB"
```

## Configuration Best Practices

1. **Resource Management**
   - Set appropriate resource limits and requests
   - Use pod anti-affinity for high availability
   - Monitor resource usage and adjust as needed

2. **Security**
   - Enable security context
   - Use non-root user
   - Implement proper ACLs
   - Enable TLS for secure communication

3. **Monitoring**
   - Enable metrics collection
   - Configure logging
   - Set up alerts for critical events

4. **Performance**
   - Configure cache settings appropriately
   - Use appropriate number of replicas
   - Monitor and tune performance metrics

## Next Steps

- [Basic Usage](../user-guide/basic-usage.md) - Getting started guide
- [Troubleshooting](../user-guide/troubleshooting.md) - Troubleshooting guide
- [API Reference](../reference/api.md) - Detailed API documentation
