# Release Guide

This document explains how to create releases for taskw so users can install it with `go install github.com/nkaewam/taskw@latest`.

## Prerequisites

1. Ensure your repository is public on GitHub at `github.com/nkaewam/taskw`
2. All changes are committed and pushed to the main branch
3. Tests pass (`go test ./...`)

## Creating a Release

### 1. Tag the release

```bash
# Create and push a version tag
git tag v1.0.0
git push origin v1.0.0
```

### 2. Automated release process

The GitHub Actions workflow (`.github/workflows/release.yml`) will automatically:

- Run tests
- Build binaries for multiple platforms:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
  - Windows (amd64)
- Create a GitHub release with the binaries attached
- Generate release notes

### 3. Verify the release

After the workflow completes:

1. Check the [releases page](https://github.com/nkaewam/taskw/releases)
2. Test installation: `go install github.com/nkaewam/taskw@v1.0.0`
3. Verify: `taskw --version` (if you add version info)

## Version Tagging Convention

Use [Semantic Versioning](https://semver.org/):

- `v1.0.0` - Major release
- `v1.1.0` - Minor release (new features)
- `v1.0.1` - Patch release (bug fixes)

## Manual Release (if needed)

If you need to create a release manually:

```bash
# Build the binary
go build -ldflags="-s -w" -o taskw main.go

# Create a release on GitHub
gh release create v1.0.0 \
  --title "Release v1.0.0" \
  --notes "Release notes here" \
  taskw
```

## Go Install Behavior

When users run `go install github.com/nkaewam/taskw@latest`:

1. Go looks for the main package in the repository root (`main.go`)
2. It builds the binary and installs it to `$GOPATH/bin` or `$HOME/go/bin`
3. The binary name will be `taskw` (based on the repository name)

## Troubleshooting

**Go install doesn't work:**

- Ensure the repository is public
- Verify `main.go` exists in the root directory
- Check that `go.mod` has the correct module name

**Binary doesn't work:**

- Ensure the `Execute()` function is exported from `cmd/taskw`
- Test locally with `go run main.go --help`

**Release workflow fails:**

- Check GitHub Actions logs
- Ensure tests pass locally
- Verify the tag format is correct (`v*.*.*`)
