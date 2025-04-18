# Troubleshooting Guide

This guide covers common issues and their solutions when using the Squid Operator.

## Common Issues

### Pod Not Starting

#### Symptoms
- Pods stuck in `Pending` or `ContainerCreating` state
- Pods in `CrashLoopBackOff` state

#### Solutions

1. Check resource availability:
```bash
kubectl describe pod <pod-name> -n squid-proxy
```

2. Check resource limits:
```bash
kubectl get squidinstance <instance-name> -n squid-proxy -o yaml
```

3. Check logs:
```bash
kubectl logs <pod-name> -n squid-proxy
```

### Service Not Accessible

#### Symptoms
- Cannot connect to Squid proxy
- Connection timeouts
- Service not responding

#### Solutions

1. Check service status:
```bash
kubectl get svc -n squid-proxy
```

2. Check endpoints:
```bash
kubectl get endpoints -n squid-proxy
```

3. Test connectivity:
```bash
kubectl exec -it <pod-name> -n squid-proxy -- curl localhost:3128
```

### Configuration Issues

#### Symptoms
- Squid not starting with configuration
- Configuration errors in logs
- Unexpected behavior

#### Solutions

1. Validate configuration:
```bash
kubectl get squidconfig <config-name> -n squid-proxy -o yaml
```

2. Check configuration syntax:
```bash
kubectl exec -it <pod-name> -n squid-proxy -- squid -k parse
```

3. Check configuration file:
```bash
kubectl exec -it <pod-name> -n squid-proxy -- cat /etc/squid/squid.conf
```

## Monitoring and Logging

### Checking Logs

1. Pod logs:
```bash
kubectl logs -f <pod-name> -n squid-proxy
```

2. Previous logs:
```bash
kubectl logs -f --previous <pod-name> -n squid-proxy
```

3. All containers:
```bash
kubectl logs -f <pod-name> -n squid-proxy --all-containers
```

### Checking Metrics

1. Prometheus metrics:
```bash
kubectl port-forward <pod-name> -n squid-proxy 3129:3129
curl localhost:3129/metrics
```

2. Squid metrics:
```bash
kubectl exec -it <pod-name> -n squid-proxy -- squidclient mgr:info
```

## Debugging Tools

### kubectl Commands

1. Describe resources:
```bash
kubectl describe squidinstance <name> -n squid-proxy
kubectl describe squidconfig <name> -n squid-proxy
kubectl describe pod <name> -n squid-proxy
```

2. Get events:
```bash
kubectl get events -n squid-proxy --sort-by='.lastTimestamp'
```

3. Check resource status:
```bash
kubectl get all -n squid-proxy
```

### Squid Debug Commands

1. Check cache:
```bash
kubectl exec -it <pod-name> -n squid-proxy -- squidclient mgr:cache
```

2. Check memory:
```bash
kubectl exec -it <pod-name> -n squid-proxy -- squidclient mgr:mem
```

3. Check network:
```bash
kubectl exec -it <pod-name> -n squid-proxy -- squidclient mgr:netdb
```

## Common Error Messages

### Configuration Errors

1. **Invalid ACL**
```
FATAL: Bungled squid.conf line X: acl
```
Solution: Check ACL syntax and definitions

2. **Port Already in Use**
```
FATAL: Cannot open HTTP Port
```
Solution: Check port configuration and ensure it's not in use

3. **Permission Denied**
```
FATAL: Cannot open log file
```
Solution: Check file permissions and ownership

### Runtime Errors

1. **Cache Directory Issues**
```
FATAL: Failed to verify one of the swap directories
```
Solution: Check cache directory permissions and disk space

2. **Memory Issues**
```
WARNING: Your cache is running out of memory
```
Solution: Adjust memory limits and cache settings

3. **Connection Issues**
```
WARNING: Forwarding loop detected
```
Solution: Check ACLs and forwarding rules

## Recovery Procedures

### Pod Recovery

1. Delete and recreate pod:
```bash
kubectl delete pod <pod-name> -n squid-proxy
```

2. Restart deployment:
```bash
kubectl rollout restart deployment <deployment-name> -n squid-proxy
```

### Configuration Recovery

1. Revert to previous configuration:
```bash
kubectl rollout undo squidconfig <config-name> -n squid-proxy
```

2. Restore from backup:
```bash
kubectl apply -f backup-config.yaml
```

## Next Steps

- [Basic Usage](../user-guide/basic-usage.md) - Getting started guide
- [Configuration Guide](../user-guide/configuration.md) - Advanced configuration options
- [API Reference](../reference/api.md) - Detailed API documentation
