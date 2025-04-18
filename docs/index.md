# Squid Operator

Welcome to the Squid Operator documentation! This operator helps you manage Squid proxy instances in your Kubernetes cluster.

## Overview

The Squid Operator simplifies the deployment and management of Squid proxy instances in Kubernetes. It provides two main custom resources:

1. **SquidConfig**: Defines the configuration for Squid proxy instances
2. **SquidInstance**: Manages the actual Squid proxy deployment

## Key Features

- **Declarative Configuration**: Manage Squid through Kubernetes custom resources
- **Configuration Management**: Separate configuration from instance deployment
- **Resource Management**: Automatic handling of Kubernetes resources
- **Security**: Built-in support for TLS and network policies

## Quick Start

1. Install the operator
2. Create a SquidConfig
3. Create a SquidInstance
4. Access your proxy

For detailed instructions, see the [Quick Start](getting-started/quickstart.md) guide.

## Documentation

### Getting Started
- [Quick Start](getting-started/quickstart.md) - Deploy your first Squid instance
- [Basic Usage](user-guide/basic-usage.md) - Learn the basics of using the operator

### Architecture
- [Overview](architecture/overview.md) - High-level architecture overview
- [Components](architecture/components.md) - Detailed component descriptions
- [Workflow](architecture/workflow.md) - How components interact

### User Guide
- [Configuration](user-guide/configuration.md) - Configure your Squid instances
- [Troubleshooting](user-guide/troubleshooting.md) - Common issues and solutions

### Reference
- [API Reference](reference/api.md) - Detailed API documentation
- [CRDs](reference/crds.md) - Custom Resource Definitions
- [Examples](reference/examples.md) - Example configurations

### Development
- [Contributing](development/contributing.md) - How to contribute to the project
- [Testing](development/testing.md) - Testing guidelines
- [Release Process](development/release.md) - Release management
- [Security](development/security.md) - Security best practices

## Project Status

The Squid Operator is actively maintained and developed. For the latest updates and releases, check our [GitLab repository](https://git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator).

## Getting Help

If you need help or have questions:
- Check the [troubleshooting guide](user-guide/troubleshooting.md)
- Open an issue in our [GitLab repository](https://git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator/-/issues)
- Contact the maintainers through the repository
