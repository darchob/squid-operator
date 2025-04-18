# Installation

This guide will help you install the Squid Operator in your Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (version 1.16 or later)
- kubectl configured to access your cluster
- Helm 3 (for Helm installation method)

## Installation Methods

### Using Helm

1. Add the Helm repository:
```bash
helm repo add squid-operator https://charts.example.com/squid-operator
```

2. Install the operator:
```bash
helm install squid-operator squid-operator/squid-operator \
  --namespace squid-operator \
  --create-namespace
```

### Using kubectl

1. Apply the CRDs:
```bash
kubectl apply -f https://raw.githubusercontent.com/your-repo/squid-operator/main/deploy/crds/
```

2. Deploy the operator:
```bash
kubectl apply -f https://raw.githubusercontent.com/your-repo/squid-operator/main/deploy/operator.yaml
```

## Verifying the Installation

After installation, verify that the operator is running:

```bash
kubectl get pods -n squid-operator
```

You should see the operator pod running:

```
NAME                              READY   STATUS    RESTARTS   AGE
squid-operator-xxxxxxxxxx-xxxxx   1/1     Running   0          1m
```

## Next Steps

- [Quick Start](quickstart.md) - Deploy your first Squid instance
- [Configuration](../user-guide/configuration.md) - Learn about configuration options
