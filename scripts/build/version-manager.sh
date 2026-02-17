#!/bin/bash
# HelixAgent Version Manager Library
# Provides version tracking, source hash computation, and change detection.
# Source this file from other build scripts.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
VERSION_DATA_DIR="$PROJECT_ROOT/releases/.version-data"

# App registry: name -> cmd path
declare -A APP_REGISTRY
APP_REGISTRY=(
    [helixagent]="./cmd/helixagent"
    [api]="./cmd/api"
    [grpc-server]="./cmd/grpc-server"
    [cognee-mock]="./cmd/cognee-mock"
    [sanity-check]="./cmd/sanity-check"
    [mcp-bridge]="./cmd/mcp-bridge"
    [generate-constitution]="./cmd/generate-constitution"
)

# Supported platforms
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# Get semantic version from VERSION file
get_semantic_version() {
    local ver
    ver="$(tr -d '[:space:]' < "$PROJECT_ROOT/VERSION")"
    echo "$ver"
}

# Get current version code for an app
get_version_code() {
    local app="$1"
    local file="$VERSION_DATA_DIR/${app}.version-code"
    if [ -f "$file" ]; then
        tr -d '[:space:]' < "$file"
    else
        echo "0"
    fi
}

# Increment and persist version code for an app
increment_version_code() {
    local app="$1"
    local current
    current="$(get_version_code "$app")"
    local next=$((current + 1))
    mkdir -p "$VERSION_DATA_DIR"
    echo "$next" > "$VERSION_DATA_DIR/${app}.version-code"
    echo "$next"
}

# Compute a SHA256 hash of the source files relevant to an app
compute_source_hash() {
    local app="$1"
    local cmd_path="${APP_REGISTRY[$app]}"
    if [ -z "$cmd_path" ]; then
        echo "error: unknown app '$app'" >&2
        return 1
    fi

    (
        cd "$PROJECT_ROOT"
        {
            # App-specific sources
            find "$cmd_path" -name '*.go' -type f 2>/dev/null
            # Shared internal sources
            find internal -name '*.go' -type f 2>/dev/null
            # Dependency files
            echo "go.mod"
            echo "go.sum"
            echo "VERSION"
        } | sort | xargs sha256sum 2>/dev/null | sha256sum | awk '{print $1}'
    )
}

# Get the last recorded source hash for an app
get_last_hash() {
    local app="$1"
    local file="$VERSION_DATA_DIR/${app}.last-hash"
    if [ -f "$file" ]; then
        tr -d '[:space:]' < "$file"
    else
        echo ""
    fi
}

# Save the source hash for an app
save_hash() {
    local app="$1"
    local hash="$2"
    mkdir -p "$VERSION_DATA_DIR"
    echo "$hash" > "$VERSION_DATA_DIR/${app}.last-hash"
}

# Check if source has changed since last build
has_changes() {
    local app="$1"
    local current_hash
    current_hash="$(compute_source_hash "$app")"
    local last_hash
    last_hash="$(get_last_hash "$app")"
    if [ "$current_hash" != "$last_hash" ]; then
        return 0 # has changes
    else
        return 1 # no changes
    fi
}

# Build ldflags string for go build
get_ldflags() {
    local app="$1"
    local version_code="$2"
    local source_hash="$3"
    local builder="${4:-local}"

    local version
    version="$(get_semantic_version)"
    local git_commit
    git_commit="$(git -C "$PROJECT_ROOT" rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
    local git_branch
    git_branch="$(git -C "$PROJECT_ROOT" rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')"
    local build_date
    build_date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

    local pkg="dev.helix.agent/internal/version"
    echo "-w -s \
-X ${pkg}.Version=${version} \
-X ${pkg}.VersionCode=${version_code} \
-X ${pkg}.GitCommit=${git_commit} \
-X ${pkg}.GitBranch=${git_branch} \
-X ${pkg}.BuildDate=${build_date} \
-X ${pkg}.SourceHash=sha256:${source_hash} \
-X ${pkg}.Builder=${builder}"
}

# List all registered apps
list_all_apps() {
    for app in "${!APP_REGISTRY[@]}"; do
        echo "$app"
    done | sort
}

# Get cmd path for an app
get_cmd_path() {
    local app="$1"
    echo "${APP_REGISTRY[$app]}"
}

# Validate that an app name is in the registry
validate_app() {
    local app="$1"
    if [ -z "${APP_REGISTRY[$app]+x}" ]; then
        echo "error: unknown app '$app'. Valid apps:" >&2
        list_all_apps >&2
        return 1
    fi
}
