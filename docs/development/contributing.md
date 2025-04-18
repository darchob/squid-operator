# Contributing Guide

This guide explains how to contribute to the Squid Operator project.

## Development Environment Setup

### Prerequisites

- Go 1.16 or later
- Docker
- kubectl
- kind or minikube
- operator-sdk

### Setting Up the Development Environment

1. Clone the repository:
```bash
git clone https://git.fr.clara.net/claranet/healthcare/buildops/projects/kubernetes/operators/squid-operator
cd squid-operator
```

2. Install dependencies:
```bash
go mod download
```

3. Set up a local Kubernetes cluster:
```bash
kind create cluster --name squid-operator
```

4. Install the operator:
```bash
make install
make deploy
```

## Development Workflow

### 1. Code Structure

The project follows a standard Go project layout:

```
.
├── api/              # API definitions
├── controllers/      # Controller implementations
├── internal/         # Internal packages
├── pkg/             # Public packages
├── config/          # Configuration files
└── deploy/          # Deployment manifests
```

### 2. Making Changes

1. Create a new branch:
```bash
git checkout -b feature/your-feature
```

2. Make your changes

3. Run tests:
```bash
make test
```

4. Build the operator:
```bash
make build
```

5. Deploy to your local cluster:
```bash
make deploy
```

### 3. Testing

#### Unit Tests
```bash
make test
```

#### Integration Tests
```bash
make test-integration
```

#### E2E Tests
```bash
make test-e2e
```

### 4. Code Style

The project follows the following coding standards:

1. **Go Code**
   - Use `gofmt` for formatting
   - Follow Go best practices
   - Write unit tests for new code

2. **Documentation**
   - Document all public APIs
   - Update README for significant changes
   - Add comments for complex logic

3. **Git Commits**
   - Use conventional commits
   - Write clear commit messages
   - Reference issues in commits

## Pull Request Process

1. Fork the repository
2. Create your feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

### Pull Request Checklist

- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Code follows style guidelines
- [ ] All tests pass
- [ ] Branch is up to date

## Documentation

### Updating Documentation

1. Documentation is written in Markdown
2. Located in the `docs/` directory
3. Follow the existing structure
4. Use proper formatting

### Adding New Documentation

1. Create a new file in the appropriate directory
2. Add the file to `mkdocs.yml`
3. Follow the existing style
4. Include examples where appropriate

## Release Process

### Versioning

The project follows semantic versioning:

- MAJOR version for incompatible API changes
- MINOR version for backward-compatible functionality
- PATCH version for backward-compatible bug fixes

### Creating a Release

1. Update version numbers
2. Update changelog
3. Create release tag
4. Build and publish artifacts

## Community Guidelines

### Code of Conduct

- Be respectful
- Be constructive
- Be collaborative
- Be professional

### Communication

- Use the issue tracker for bugs and features
- Use pull requests for code changes
- Use discussions for questions and ideas

## Getting Help

If you need help:

1. Check the documentation
2. Search existing issues
3. Open a new issue
4. Contact the maintainers

## Next Steps

- [Testing](testing.md) - Testing guidelines
- [Release Process](release.md) - Release management
- [Security](security.md) - Security guidelines
