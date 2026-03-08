# internal/version

## Overview

The version package is the single source of truth for HelixAgent's build
version information. All fields are injected at build time via
`-ldflags -X` and default to development values when built without flags.

## Build-Time Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `Version` | `"dev"` | Semantic version string |
| `VersionCode` | `"0"` | Numeric build counter |
| `GitCommit` | `"unknown"` | Git commit SHA |
| `GitBranch` | `"unknown"` | Git branch name |
| `BuildDate` | `"unknown"` | ISO 8601 build timestamp |
| `SourceHash` | `"unknown"` | SHA256 of source files |
| `Builder` | `"local"` | Build environment identifier |

## Key Types

### Info

```go
type Info struct {
    Version     string `json:"version"`
    VersionCode string `json:"version_code"`
    GitCommit   string `json:"git_commit"`
    GitBranch   string `json:"git_branch"`
    BuildDate   string `json:"build_date"`
    SourceHash  string `json:"source_hash"`
    Builder     string `json:"builder"`
    GoVersion   string `json:"go_version"`
    Platform    string `json:"platform"`
}
```

## Key Functions

| Function | Description |
|----------|-------------|
| `Get()` | Returns the full `Info` struct with all build metadata |
| `Short()` | Returns a one-line summary: `HelixAgent v1.0.0 (build 42, commit abc123, 2026-01-01)` |
| `(Info) String()` | Returns a multi-line version description |
| `(Info) JSON()` | Returns JSON-encoded version info |

## Build Example

```bash
go build -ldflags "\
  -X dev.helix.agent/internal/version.Version=1.0.0 \
  -X dev.helix.agent/internal/version.VersionCode=42 \
  -X dev.helix.agent/internal/version.GitCommit=$(git rev-parse --short HEAD) \
  -X dev.helix.agent/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -X dev.helix.agent/internal/version.Builder=container" \
  ./cmd/helixagent
```

Release builds use `make release` which performs builds inside Docker
containers and injects all version fields automatically.
