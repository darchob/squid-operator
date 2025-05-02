# SquidConfig and SquidInstance Workflow

This document explains how SquidConfigs and SquidInstances interact within the Squid Operator.

## Resource Interaction

The Squid Operator manages two main custom resources:

1. **SquidConfig**: Defines the configuration for Squid proxy instances
2. **SquidInstance**: Manages the actual Squid proxy deployment

## Workflow Diagram

```mermaid
flowchart TD
    %% Styling
    classDef success fill:#d4edda,stroke:#c3e6cb,color:#155724,stroke-width:2px,stroke-dasharray: 0
    classDef warning fill:#fff3cd,stroke:#ffeeba,color:#856404,stroke-width:2px,stroke-dasharray: 0
    classDef danger fill:#f8d7da,stroke:#f5c6cb,color:#721c24,stroke-width:2px,stroke-dasharray: 0
    classDef info fill:#d1ecf1,stroke:#bee5eb,color:#0c5460,stroke-width:2px,stroke-dasharray: 0
    classDef primary fill:#cce5ff,stroke:#b8daff,color:#004085,stroke-width:2px,stroke-dasharray: 0

    %% SquidInstance Flow
    subgraph SquidInstance["SquidInstance Flow"]
        direction TB
        A([SquidInstance Created]):::primary -->|Create| B([Webhook Validation]):::info
        B -->|Validate| C([Create ConfigMap]):::primary
        C -->|Create| D([ConfigMap Ready]):::success
        D -->|Deploy| E([Deploy Squid]):::primary
        E -->|Monitor| F([Monitor Status]):::info
        F -->|Update| G([Update Status]):::success
    end

    %% SquidConfig Flow
    subgraph SquidConfig["SquidConfig Flow"]
        direction TB
        H([SquidConfig Created]):::primary -->|Create| I([Webhook Validation]):::info
        I -->|Validate| J([Create Validation Job]):::primary
        J -->|Run| K([Validate Config]):::info
        K -->|Update| L([Update ConfigMap Data]):::primary
        L -->|Trigger| M([Trigger Reconciliation]):::info
        M -->|Reload| N([Reload Squid]):::primary
        N -->|Monitor| F
    end

    %% Update Flow
    O([SquidConfig Updated]):::warning -.->|Update| I

    %% Delete Flow
    subgraph Deletion["Deletion Flow"]
        direction TB
        P([SquidConfig Deleted]):::danger -->|Validate| Q([Webhook Validation]):::info
        Q -->|Check| R([Check Config Usage]):::info
        R -->|Allow/Deny| S([Allow/Deny Deletion]):::danger
    end

    %% Link styles
    linkStyle default stroke:#666,stroke-width:2px
    linkStyle 0,1,2,3,4,5,6,7,8,9,10,11,12,13,14 stroke:#666,stroke-width:2px
    linkStyle 15 stroke:#666,stroke-width:2px,stroke-dasharray: 5 5
```

## Detailed Workflow

### 1. SquidInstance Creation

When a SquidInstance is created:

1. The webhook validates the instance configuration
   - Checks image configuration
   - Validates storage class
   - Ensures resource requirements
2. The operator creates a ConfigMap for the Squid configuration
3. Creates the necessary Kubernetes resources
4. Deploys the Squid proxy with the configuration

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidInstance
metadata:
  name: my-squid
  namespace: squid-proxy
spec:
  replicas: 1
  image:
    repository: ubuntu/squid
    tag: latest
  storageClassName: standard
```

### 2. SquidConfig Creation/Update

When a SquidConfig is created or updated:

1. The webhook performs validation:
   - Creates a validation job
   - Ensures required instance annotation
   - Validates configuration syntax using squid -k parse
   - Timeout of 2 minutes for validation
2. The operator updates the data in the existing ConfigMap
3. Triggers reconciliation of the affected SquidInstance
4. Reloads the Squid configuration in the running pods

```yaml
apiVersion: squid.claranet.fr/v1
kind: SquidConfig
metadata:
  name: my-squid-config
  namespace: squid-proxy
  annotations:
    squid.ckd.clara.net/instance: my-squid
spec:
  config: |
    http_port 3128
    acl SSL_ports port 443
    acl Safe_ports port 80
```

### 3. SquidConfig Deletion

When a SquidConfig is deleted:

1. The webhook performs pre-deletion validation:
   - Checks if configuration is still in use
   - Verifies configmap references
   - Prevents deletion if configuration is active
2. If validation passes, the resource is deleted
3. The operator updates the affected SquidInstance

### 4. Configuration Updates

When a SquidConfig is updated:

```go
func (r *SquidInstanceReconciler) handleConfigUpdate(ctx context.Context, instance *v1.SquidInstance) error {
    // Get the ConfigMap
    configMap := &corev1.ConfigMap{}
    if err := r.Get(ctx, types.NamespacedName{
        Name:      instance.Spec.ConfigMapName,
        Namespace: instance.Namespace,
    }, configMap); err != nil {
        return err
    }

    // Update Squid configuration
    if err := r.updateSquidConfig(ctx, instance, configMap); err != nil {
        return err
    }

    return nil
}
```

### 5. Status Management

The operator maintains status information for both resources:

```go
type SquidConfigStatus struct {
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    Phase      ConfigPhase       `json:"phase,omitempty"`
}

type SquidInstanceStatus struct {
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    Health     StatusPhase        `json:"health,omitempty"`
}
```

## Best Practices

1. **Configuration Management**
   - Use separate SquidConfigs for different environments
   - Version control your configurations
   - Test configurations before applying
   - Always include the required instance annotation

2. **Instance Management**
   - Create instances in the same namespace as their configs
   - Monitor instance health
   - Use appropriate resource limits
   - Ensure storage class is properly configured

3. **Updates**
   - Plan configuration changes
   - Test updates in staging
   - Monitor for issues
   - Be aware of the 2-minute validation timeout

4. **Deletion**
   - Check for active usage before deleting configurations
   - Ensure proper cleanup of related resources
   - Monitor for any orphaned resources

## Next Steps

- [API Reference](../reference/api.md) - Detailed API documentation
- [Configuration](../user-guide/configuration.md) - Configuration options
- [Monitoring](../user-guide/monitoring.md) - Monitoring setup
