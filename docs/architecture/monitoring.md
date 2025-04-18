# Monitoring

This document describes the monitoring architecture and configuration of the Squid Operator.

## Monitoring Architecture

### Overview

The monitoring system consists of:

1. **Metrics Collection**
   - Squid metrics
   - Kubernetes metrics
   - Custom metrics

2. **Metrics Storage**
   - Prometheus
   - Time-series database

3. **Visualization**
   - Grafana dashboards
   - Custom visualizations

4. **Alerting**
   - AlertManager
   - Notification channels

## Metrics Collection

### Squid Metrics

Squid provides various metrics through its manager interface:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: metrics-config
spec:
  config: |
    # Enable metrics
    http_port 3129
    acl manager proto cache_object
    acl localhost src 127.0.0.1/32
    http_access allow manager localhost
    http_access deny manager
```

### Prometheus Integration

Prometheus configuration for scraping Squid metrics:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: squid-monitor
  namespace: squid-proxy
spec:
  selector:
    matchLabels:
      app: squid
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

## Metrics Types

### Performance Metrics

Key performance metrics:

- Request rate
- Cache hit ratio
- Response times
- Connection counts
- Memory usage
- CPU usage

### Business Metrics

Business-related metrics:

- User count
- Bandwidth usage
- Top domains
- Error rates
- Authentication attempts

## Visualization

### Grafana Dashboards

Example dashboard configuration:

```yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaDashboard
metadata:
  name: squid-dashboard
  namespace: squid-proxy
spec:
  json: |
    {
      "title": "Squid Proxy Dashboard",
      "panels": [
        {
          "title": "Request Rate",
          "type": "graph",
          "datasource": "Prometheus",
          "targets": [
            {
              "expr": "rate(squid_requests_total[5m])",
              "legendFormat": "{{instance}}"
            }
          ]
        }
      ]
    }
```

## Alerting

### Alert Rules

Prometheus alert rules:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: squid-alerts
  namespace: squid-proxy
spec:
  groups:
  - name: squid
    rules:
    - alert: HighErrorRate
      expr: rate(squid_errors_total[5m]) > 0.1
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: High error rate detected
        description: Squid proxy is experiencing a high error rate
```

### Notification Channels

AlertManager configuration:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: AlertmanagerConfig
metadata:
  name: squid-alerts
  namespace: squid-proxy
spec:
  receivers:
  - name: default
    emailConfigs:
    - to: admin@example.com
      sendResolved: true
  route:
    groupBy: ['alertname']
    groupWait: 30s
    groupInterval: 5m
    repeatInterval: 4h
    receiver: default
```

## Logging

### Log Configuration

Squid log configuration:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: logging-config
spec:
  config: |
    access_log /var/log/squid/access.log
    cache_log /var/log/squid/cache.log
    logformat squid %ts.%03tu %6tr %>a %Ss/%03>Hs %<st %rm %ru %un %Sh/%<A %mt
```

### Log Collection

Fluentd configuration for log collection:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
  namespace: squid-proxy
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/squid/*.log
      pos_file /var/log/fluentd-squid.pos
      tag squid.*
      format none
    </source>
    <match squid.**>
      @type elasticsearch
      host elasticsearch
      port 9200
      logstash_format true
    </match>
```

## Troubleshooting

### Monitoring Issues

Common monitoring issues and solutions:

1. **Metrics Not Showing**
   - Check Prometheus configuration
   - Verify service monitor
   - Check network policies

2. **High Alert Volume**
   - Adjust alert thresholds
   - Review alert rules
   - Check for false positives

3. **Log Collection Issues**
   - Verify log paths
   - Check log permissions
   - Verify log collector configuration

## Next Steps

- [Components](../architecture/components.md) - Component architecture
- [Networking](../architecture/networking.md) - Network configuration
- [Security](../development/security.md) - Security considerations
