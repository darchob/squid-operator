# Scaling Guide

This guide explains how to scale your Squid instances and configure autoscaling.

## Manual Scaling

### Scaling Replicas

You can manually scale the number of replicas for a Squid instance:

```bash
kubectl scale squidinstance my-squid -n squid-proxy --replicas=3
```

Or update the configuration:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: my-squid
  namespace: squid-proxy
spec:
  replicas: 3
  # ... other configuration ...
```

## Automatic Scaling

### Horizontal Pod Autoscaling

Configure automatic scaling based on CPU or memory usage:

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: my-squid
  namespace: squid-proxy
spec:
  hpaSpec:
    minReplicas: 1
    maxReplicas: 5
    metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 80
```

### Custom Metrics

You can also scale based on custom metrics:

```yaml
spec:
  hpaSpec:
    minReplicas: 1
    maxReplicas: 5
    metrics:
    - type: Pods
      pods:
        metric:
          name: requests-per-second
        target:
          type: AverageValue
          averageValue: 1000
```

## Scaling Considerations

### Resource Requirements

When scaling, consider the following resource requirements:

1. **CPU**
   - Base requirement: 500m per instance
   - Scale up when CPU usage exceeds 80%

2. **Memory**
   - Base requirement: 512Mi per instance
   - Scale up when memory usage exceeds 80%

3. **Storage**
   - Each instance needs persistent storage
   - Consider storage class and size requirements

### Performance Impact

Scaling affects performance in several ways:

1. **Cache Distribution**
   - Each replica maintains its own cache
   - Consider cache synchronization

2. **Load Balancing**
   - Traffic is distributed across replicas
   - Monitor load distribution

3. **Network Overhead**
   - Increased network traffic between replicas
   - Consider network bandwidth

## Best Practices

1. **Scaling Strategy**
   - Start with conservative limits
   - Monitor performance metrics
   - Adjust based on actual usage

2. **Resource Allocation**
   - Set appropriate resource limits
   - Consider burst capacity
   - Monitor resource usage

3. **High Availability**
   - Maintain minimum replicas for HA
   - Use anti-affinity rules
   - Consider zone distribution

## Monitoring Scaling

### Metrics to Monitor

1. **Resource Usage**
   - CPU utilization
   - Memory usage
   - Network traffic

2. **Performance Metrics**
   - Request latency
   - Cache hit ratio
   - Connection count

3. **Scaling Events**
   - Scale-up events
   - Scale-down events
   - Failed scaling attempts

### Alerts

Set up alerts for scaling events:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: squid-scaling-alerts
  namespace: squid-proxy
spec:
  groups:
  - name: squid-scaling
    rules:
    - alert: SquidScalingFailed
      expr: kube_horizontalpodautoscaler_status_condition{condition="ScalingLimited"} == 1
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: Squid instance scaling is limited
        description: Squid instance {{ $labels.instance }} scaling is limited for more than 5 minutes
```

## Next Steps

- [Monitoring](monitoring.md) - Learn how to monitor your Squid instances
- [Configuration](configuration.md) - Advanced configuration options
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
