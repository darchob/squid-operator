# Networking

This document describes the networking architecture and configuration of the Squid Operator.

## Network Architecture

### Overview

The Squid Operator implements a proxy architecture with the following components:

1. **Client Network**
   - Client applications
   - Network policies
   - Service endpoints

2. **Proxy Network**
   - Squid proxy instances
   - Load balancing
   - Service configuration

3. **Upstream Network**
   - External services
   - DNS resolution
   - Connection handling

## Network Components

### Services

The operator creates Kubernetes services for Squid instances:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: squid-service
  namespace: squid-proxy
spec:
  type: ClusterIP
  ports:
  - port: 3128
    targetPort: 3128
    protocol: TCP
  selector:
    app: squid
```

Service Types:
- ClusterIP (default)
- NodePort
- LoadBalancer

### Network Policies

Network policies control traffic flow:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: squid-network-policy
  namespace: squid-proxy
spec:
  podSelector:
    matchLabels:
      app: squid
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: allowed-client
    ports:
    - protocol: TCP
      port: 3128
```

## Proxy Configuration

### Port Configuration

Default ports:
- HTTP Proxy: 3128
- HTTPS Proxy: 3128
- Monitoring: 3129

### Protocol Support

Supported protocols:
- HTTP
- HTTPS
- FTP
- SSL/TLS

## DNS Configuration

### DNS Resolution

Squid DNS configuration:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: dns-config
spec:
  config: |
    dns_nameservers 8.8.8.8 8.8.4.4
    dns_v4_first on
    dns_defnames on
```

### DNS Cache

DNS caching configuration:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: dns-cache-config
spec:
  config: |
    dns_cache_size 1024
    dns_cache_ttl 3600
    dns_cache_garbage_interval 3600
```

## Security

### TLS Configuration

TLS settings for secure communication:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: tls-config
spec:
  config: |
    https_port 3128 cert=/etc/squid/ssl/cert.pem key=/etc/squid/ssl/key.pem
    ssl_bump server-first all
    sslproxy_cert_error allow all
```

### Authentication

Authentication configuration:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: auth-config
spec:
  config: |
    auth_param basic program /usr/lib/squid/basic_ncsa_auth /etc/squid/passwd
    auth_param basic realm Squid proxy-caching web server
    auth_param basic credentialsttl 2 hours
    acl authenticated proxy_auth REQUIRED
    http_access allow authenticated
```

## Performance

### Connection Handling

Connection management settings:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: connection-config
spec:
  config: |
    client_persistent_connections on
    server_persistent_connections on
    pconn_timeout 1 minute
    read_timeout 15 minutes
    request_timeout 5 minutes
```

### Load Balancing

Load balancing configuration:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: load-balanced-instance
spec:
  replicas: 3
  serviceType: LoadBalancer
  servicePort: 3128
```

## Monitoring

### Network Metrics

Network-related metrics:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: metrics-config
spec:
  config: |
    logformat metrics %ts.%03tu %6tr %>a %Ss/%03>Hs %<st %rm %ru %un %Sh/%<A %mt
    access_log /var/log/squid/metrics.log metrics
```

## Troubleshooting

### Network Issues

Common network issues and solutions:

1. **Connection Timeouts**
   - Check network policies
   - Verify service configuration
   - Check DNS resolution

2. **Performance Issues**
   - Monitor connection counts
   - Check load balancing
   - Verify resource limits

3. **Authentication Failures**
   - Check authentication configuration
   - Verify credentials
   - Check network access

## Next Steps

- [Components](../architecture/components.md) - Component architecture
- [Monitoring](../architecture/monitoring.md) - Monitoring setup
- [Security](../development/security.md) - Security considerations
