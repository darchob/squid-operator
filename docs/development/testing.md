# Testing Guide

This guide explains how to test the Squid Operator and its components.

## Testing Types

### 1. Unit Tests

Unit tests test individual components in isolation.

#### Running Unit Tests

```bash
make test
```

#### Example Unit Test

```go
func TestSquidInstanceReconciler_Reconcile(t *testing.T) {
    testCases := []struct {
        name    string
        instance *v1.SquidInstance
        wantErr bool
    }{
        {
            name: "valid instance",
            instance: &v1.SquidInstance{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-instance",
                    Namespace: "default",
                },
                Spec: v1.SquidInstanceSpec{
                    Replicas: pointer.Int32(1),
                },
            },
            wantErr: false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 2. Integration Tests

Integration tests test the interaction between components.

#### Running Integration Tests

```bash
make test-integration
```

#### Example Integration Test

```go
func TestSquidInstanceIntegration(t *testing.T) {
    ctx := context.Background()

    // Setup test environment
    env := &envtest.Environment{
        CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
    }

    cfg, err := env.Start()
    if err != nil {
        t.Fatal(err)
    }
    defer env.Stop()

    // Test implementation
}
```

### 3. End-to-End Tests

E2E tests test the entire system in a real Kubernetes cluster.

#### Running E2E Tests

```bash
make test-e2e
```

#### Example E2E Test

```go
func TestSquidInstanceE2E(t *testing.T) {
    ctx := context.Background()

    // Create test namespace
    ns := &corev1.Namespace{
        ObjectMeta: metav1.ObjectMeta{
            GenerateName: "squid-test-",
        },
    }
    if err := k8sClient.Create(ctx, ns); err != nil {
        t.Fatal(err)
    }

    // Test implementation
}
```

## Test Environment Setup

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Set up test environment:
```bash
make test-env
```

3. Run tests:
```bash
make test
```

### CI/CD Pipeline

The CI/CD pipeline runs tests automatically:

1. Unit tests
2. Integration tests
3. E2E tests
4. Code coverage
5. Linting

## Test Coverage

### Generating Coverage Report

```bash
make test-coverage
```

### Coverage Requirements

- Minimum coverage: 80%
- Critical paths: 100%
- New code: 90%

## Mocking

### Using Mock Objects

```go
type MockClient struct {
    mock.Mock
}

func (m *MockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
    args := m.Called(ctx, key, obj)
    return args.Error(0)
}
```

### Mocking Kubernetes API

```go
func TestWithMockKubernetes(t *testing.T) {
    mockClient := &MockClient{}

    // Setup expectations
    mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).
        Return(nil)

    // Test implementation
}
```

## Test Data

### Test Resources

Store test resources in `testdata/`:

```
testdata/
├── fixtures/
│   ├── valid-instance.yaml
│   └── invalid-instance.yaml
└── expected/
    ├── deployment.yaml
    └── service.yaml
```

### Loading Test Data

```go
func loadTestData(t *testing.T, path string) []byte {
    data, err := ioutil.ReadFile(filepath.Join("testdata", path))
    if err != nil {
        t.Fatal(err)
    }
    return data
}
```

## Best Practices

### 1. Test Organization

- Group related tests
- Use descriptive test names
- Follow the Arrange-Act-Assert pattern

### 2. Test Data Management

- Use fixtures for complex data
- Generate test data programmatically when possible
- Clean up test data after tests

### 3. Error Handling

- Test error conditions
- Verify error messages
- Check error types

### 4. Performance

- Run tests in parallel when possible
- Minimize test setup time
- Use appropriate test scopes

## Troubleshooting Tests

### Common Issues

1. **Test Timeouts**
   - Increase timeout duration
   - Check for deadlocks
   - Verify resource cleanup

2. **Resource Conflicts**
   - Use unique namespaces
   - Clean up resources
   - Check for existing resources

3. **Environment Issues**
   - Verify Kubernetes access
   - Check resource limits
   - Verify network connectivity

### Debugging Tests

1. **Logging**
   ```go
   t.Logf("Debug information: %v", data)
   ```

2. **Breakpoints**
   ```go
   // Add breakpoint
   runtime.Breakpoint()
   ```

3. **Test Output**
   ```bash
   go test -v ./...
   ```

## Next Steps

- [Contributing](contributing.md) - Contributing guidelines
- [Release Process](release.md) - Release management
- [Security](security.md) - Security guidelines
