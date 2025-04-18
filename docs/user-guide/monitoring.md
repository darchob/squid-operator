# Monitoring Guide

This guide explains how to monitor your Squid instances and the operator itself.

## Built-in Monitoring

The Squid Operator provides several built-in monitoring capabilities:

### Health Status

Monitor the health status of your Squid instances:

```bash
kubectl get squidinstance -n squid-proxy
```

The output will show the health status of each instance:

```
NAME       HEALTH     AGE
my-squid   Deployed   1h
```

### Detailed Status

Get detailed status information:

```bash
kubectl describe squidinstance my-squid -n squid-proxy
```

## Metrics

### Prometheus Metrics

The operator exposes Prometheus metrics at the `/metrics` endpoint. To enable metrics collection:

1. Create a ServiceMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: squid-operator
  namespace: squid-proxy
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: squid-operator
  endpoints:
  - port: metrics
    interval: 30s
```

### Available Metrics

The operator exposes the following metrics:

- `squid_operator_reconcile_total`: Total number of reconciliations
- `squid_operator_reconcile_errors_total`: Total number of reconciliation errors
- `squid_instance_status`: Current status of Squid instances
- `squid_instance_replicas`: Number of replicas per instance

## Logging

### Operator Logs

View operator logs:

```bash
kubectl logs -n squid-operator deployment/squid-operator
```

### Squid Instance Logs

View logs for a specific Squid instance:

```bash
kubectl logs -n squid-proxy deployment/my-squid
```

## Alerts

### Recommended Alerts

Create the following alerts in your monitoring system:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: squid-operator-alerts
  namespace: squid-proxy
spec:
  groups:
  - name: squid-operator
    rules:
    - alert: SquidInstanceDegraded
      expr: squid_instance_status{status="Error"} == 1
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: Squid instance is in error state
        description: Squid instance {{ $labels.instance }} has been in error state for more than 5 minutes

    - alert: SquidInstanceUnhealthy
      expr: squid_instance_status{status!="Deployed"} == 1
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: Squid instance is unhealthy
        description: Squid instance {{ $labels.instance }} has been unhealthy for more than 10 minutes
```

## Grafana Dashboards

### Recommended Dashboards

Create a Grafana dashboard with the following panels:

1. **Overview**
   - Number of Squid instances
   - Total reconciliation count
   - Error rate

2. **Per Instance**
   - Health status
   - Number of replicas
   - Resource usage
   - Request rate
   - Cache hit ratio

## Troubleshooting

### Common Issues

1. **Instance Not Starting**
   - Check resource limits
   - Verify storage configuration
   - Check logs for errors

2. **High Error Rate**
   - Check network connectivity
   - Verify ACL configuration
   - Monitor resource usage

3. **Performance Issues**
   - Check cache configuration
   - Monitor resource usage
   - Verify autoscaling settings

## Next Steps

- [Scaling](scaling.md) - Learn how to scale your Squid instances
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
- [Configuration](configuration.md) - Advanced configuration options
