# Release Process

This guide explains the release process for the Squid Operator.

## Release Types

### 1. Major Releases (X.0.0)

- Breaking changes
- New features
- Significant improvements

### 2. Minor Releases (0.X.0)

- New features
- Backward-compatible changes
- Improvements

### 3. Patch Releases (0.0.X)

- Bug fixes
- Security updates
- Minor improvements

## Release Checklist

### Pre-Release

1. **Code Freeze**
   - Stop merging new features
   - Focus on bug fixes
   - Update documentation

2. **Testing**
   - Run all tests
   - Verify CI/CD pipeline
   - Check code coverage

3. **Documentation**
   - Update changelog
   - Update version numbers
   - Review documentation

4. **Version Update**
   - Update version in code
   - Update version in manifests
   - Update version in documentation

### Release Process

1. **Create Release Branch**
```bash
git checkout -b release/vX.Y.Z
```

2. **Update Version**
```bash
# Update version in code
make update-version VERSION=vX.Y.Z

# Update version in manifests
make update-manifests VERSION=vX.Y.Z
```

3. **Build Artifacts**
```bash
make build
make docker-build
make docker-push
```

4. **Create Release Tag**
```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

5. **Create Release Notes**
```markdown
# Release vX.Y.Z

## Changes
- Feature 1
- Feature 2
- Bug fix 1
- Bug fix 2

## Breaking Changes
- Breaking change 1
- Breaking change 2

## Upgrading
Instructions for upgrading from previous version
```

### Post-Release

1. **Update Default Branch**
   - Merge release branch
   - Update version numbers
   - Update documentation

2. **Announce Release**
   - Update website
   - Send announcements
   - Update social media

3. **Monitor**
   - Watch for issues
   - Respond to feedback
   - Plan next release

## Version Management

### Version Numbers

Follow semantic versioning:
- MAJOR: Breaking changes
- MINOR: New features
- PATCH: Bug fixes

### Version Tags

Format: `vX.Y.Z`
- X: Major version
- Y: Minor version
- Z: Patch version

Example: `v1.2.3`

## Release Automation

### CI/CD Pipeline

The release process is automated through the CI/CD pipeline:

1. **Build**
   - Build artifacts
   - Run tests
   - Generate documentation

2. **Deploy**
   - Push images
   - Update manifests
   - Deploy documentation

3. **Release**
   - Create tag
   - Generate release notes
   - Publish artifacts

### Release Scripts

```bash
#!/bin/bash
# release.sh

VERSION=$1

# Update version
make update-version VERSION=$VERSION

# Build
make build
make docker-build
make docker-push

# Create tag
git tag -a $VERSION -m "Release $VERSION"
git push origin $VERSION
```

## Rollback Procedure

### If Release Fails

1. **Stop Deployment**
```bash
kubectl rollout undo deployment/squid-operator
```

2. **Revert Changes**
```bash
git revert <commit>
git push origin main
```

3. **Clean Up**
```bash
# Remove tag
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z

# Delete release
# (On GitHub/GitLab)
```

## Release Notes

### Format

```markdown
# Release vX.Y.Z

## Changes
- List of changes

## Breaking Changes
- List of breaking changes

## Upgrading
Instructions for upgrading

## Known Issues
List of known issues

## Contributors
List of contributors
```

### Example

```markdown
# Release v1.2.0

## Changes
- Added support for custom Squid configurations
- Improved performance monitoring
- Enhanced security features

## Breaking Changes
- Changed default port from 3128 to 8080
- Updated resource requirements

## Upgrading
To upgrade from v1.1.0:
1. Backup your configurations
2. Update the operator
3. Apply new configurations

## Known Issues
- Issue with cache persistence
- Performance degradation under high load

## Contributors
- John Doe
- Jane Smith
```

## Next Steps

- [Contributing](contributing.md) - Contributing guidelines
- [Testing](testing.md) - Testing guidelines
- [Security](security.md) - Security guidelines
