# Release Build System Guide

## Overview

HelixAgent uses a container-based release build system with auto-incrementing version codes, source hash change detection, and multi-platform support. All release builds run inside Docker/Podman containers for reproducibility.

## Quick Start

```bash
# Build helixagent for all platforms
make release

# Build a specific app
make release-api

# Build all 7 apps for all platforms
make release-all

# Force rebuild (ignores change detection)
make release-force

# Check version status
make release-info
```

## Version Management

### Semantic Version

The project semantic version is stored in the `VERSION` file at the project root. This is a single line containing `X.Y.Z` (e.g., `1.0.0`).

### Version Code

Each app has an independent auto-incrementing integer version code. Version codes start at 0 and increment by 1 on each successful build. Stored in `releases/.version-data/<app>.version-code`.

### Source Hash

Before each build, a SHA256 hash is computed from:
- `cmd/<app>/*.go` (app-specific sources)
- `internal/**/*.go` (shared internal sources)
- `go.mod` and `go.sum` (dependency definitions)
- `VERSION` (semantic version)

The hash is stored in `releases/.version-data/<app>.last-hash` and used for change detection.

## Change Detection

When you run `make release`, the system:
1. Computes the current source hash for the app
2. Compares it to the stored hash from the last successful build
3. Skips the build if no changes are detected
4. Use `--force` to bypass change detection

## Container Builds

All release builds execute inside a builder container (`helixagent-builder:latest`) based on `golang:1.24-alpine`. The container image is built automatically if missing.

```bash
# Manually build the builder image
make release-builder-image
```

The build system auto-detects Docker or Podman. For Podman, `--userns=host` is added automatically.

## App Registry

All 7 buildable applications:

| App | Source | Description |
|-----|--------|-------------|
| `helixagent` | `cmd/helixagent` | Main HelixAgent service |
| `api` | `cmd/api` | Protocol Enhancement REST API |
| `grpc-server` | `cmd/grpc-server` | gRPC LLM Facade service |
| `cognee-mock` | `cmd/cognee-mock` | Cognee mock server |
| `sanity-check` | `cmd/sanity-check` | Sanity check utility |
| `mcp-bridge` | `cmd/mcp-bridge` | MCP bridge service |
| `generate-constitution` | `cmd/generate-constitution` | Constitution generator |

## Target Platforms

| Platform | GOOS | GOARCH |
|----------|------|--------|
| Linux x86_64 | linux | amd64 |
| Linux ARM64 | linux | arm64 |
| macOS x86_64 | darwin | amd64 |
| macOS ARM64 | darwin | arm64 |
| Windows x86_64 | windows | amd64 |

## Directory Structure

```
releases/
  .version-data/                # Git-tracked version metadata
    helixagent.version-code     # e.g., "3"
    helixagent.last-hash        # SHA256 of source at last build
    api.version-code
    api.last-hash
    ...
  helixagent/                   # Git-ignored binary artifacts
    linux-amd64/
      1/
        helixagent
        build-info.json
      2/
        helixagent
        build-info.json
      latest -> 2               # Symlink to latest version code
    linux-arm64/
      ...
    darwin-amd64/
      ...
  api/
    linux-amd64/
      ...
  ...
```

## build-info.json

Each release build generates a `build-info.json` alongside the binary:

```json
{
  "app": "helixagent",
  "version": "1.0.0",
  "version_code": 2,
  "git_commit": "b5f37c3",
  "git_branch": "main",
  "build_date": "2026-02-17T12:00:00Z",
  "platform": "linux/amd64",
  "go_version": "go1.24",
  "source_hash": "sha256:abc123...",
  "builder": "container"
}
```

## Version Injection

Version information is injected at build time via Go `-ldflags -X`:

```
-X dev.helix.agent/internal/version.Version=1.0.0
-X dev.helix.agent/internal/version.VersionCode=2
-X dev.helix.agent/internal/version.GitCommit=b5f37c3
-X dev.helix.agent/internal/version.GitBranch=main
-X dev.helix.agent/internal/version.BuildDate=2026-02-17T12:00:00Z
-X dev.helix.agent/internal/version.SourceHash=sha256:abc123
-X dev.helix.agent/internal/version.Builder=container
```

Without ldflags (dev builds), defaults are: `Version=dev`, `VersionCode=0`, etc.

## Adding a New App

1. Create the app in `cmd/<name>/`
2. Add an entry to the `APP_REGISTRY` associative array in `scripts/build/version-manager.sh`
3. Add `releases/<name>/` to `.gitignore`
4. Add a `release-<name>` target to the Makefile
5. Add the app name to `release-clean` and `release-clean-all` targets

## Adding a New Platform

Add the `OS/ARCH` pair to the `PLATFORMS` array in `scripts/build/version-manager.sh`.

## Makefile Targets

| Target | Description |
|--------|-------------|
| `release` | Build helixagent for all platforms |
| `release-all` | Build all 7 apps for all platforms |
| `release-<app>` | Build a specific app for all platforms |
| `release-force` | Force rebuild all apps (ignore change detection) |
| `release-clean` | Clean release artifacts (keep version data) |
| `release-clean-all` | Clean all release data including version tracking |
| `release-info` | Show version codes and hashes for all apps |
| `release-builder-image` | Build the builder container image |

## Troubleshooting

### Builder image not found
Run `make release-builder-image` to build it manually. The build scripts will also auto-build it when needed.

### Podman permission errors
The build system automatically adds `--userns=host` for Podman. If issues persist, ensure your user has proper subuid/subgid mappings.

### SELinux volume mount issues
Volume mounts use the `:Z` suffix for SELinux compatibility. If builds fail with permission errors, check SELinux contexts.

### "No changes detected" but you want to rebuild
Use `--force` flag or `make release-force` to bypass change detection.

### Cleaning up
- `make release-clean` removes binary artifacts but keeps version codes (so the next build increments properly)
- `make release-clean-all` removes everything including version tracking (resets to version code 0)
