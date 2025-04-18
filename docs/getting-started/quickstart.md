# Quick Start

This guide will help you quickly deploy a Squid proxy using the Squid Operator.

## Prerequisites

- Kubernetes cluster (version 1.16 or later)
- kubectl configured to communicate with your cluster
- Helm 3 installed

## Installation

1. Add the Helm repository:

```bash
helm repo add squid-operator https://charts.squid-operator.io
helm repo update
```

2. Install the operator:

```bash
helm install squid-operator squid-operator/squid-operator -n squid-proxy --create-namespace
```

## Deploy a Squid Proxy

1. Create a SquidConfig:

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

2. Create a SquidInstance:

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

3. Apply the configurations:

```bash
kubectl apply -f squid-config.yaml
kubectl apply -f squid-instance.yaml
```

## Access the Proxy

Once deployed, you can access the Squid proxy through the Kubernetes service. The service will be available at:

```bash
kubectl get svc my-squid -n squid-proxy
```

## Next Steps

- [Configuration Guide](../user-guide/configuration.md) - Learn how to configure your Squid instances
- [Troubleshooting](../user-guide/troubleshooting.md) - Common issues and solutions
- [Examples](../reference/examples.md) - Example configurations
