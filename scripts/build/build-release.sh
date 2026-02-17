#!/bin/bash
# HelixAgent Release Build Script
# Builds a single app for one or all platforms using a container builder.
#
# Usage:
#   ./scripts/build/build-release.sh --app helixagent [--platform linux/amd64|--all-platforms] [--force]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source dependencies
source "$SCRIPT_DIR/version-manager.sh"

# Detect container runtime
detect_runtime() {
    if command -v docker &>/dev/null && docker info &>/dev/null 2>&1; then
        echo "docker"
    elif command -v podman &>/dev/null; then
        echo "podman"
    else
        echo ""
    fi
}

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Parse arguments
APP=""
PLATFORM=""
ALL_PLATFORMS=false
FORCE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --app)
            APP="$2"
            shift 2
            ;;
        --platform)
            PLATFORM="$2"
            shift 2
            ;;
        --all-platforms)
            ALL_PLATFORMS=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 --app <name> [--platform <os/arch>|--all-platforms] [--force]"
            exit 1
            ;;
    esac
done

if [ -z "$APP" ]; then
    echo -e "${RED}ERROR: --app is required${NC}"
    echo "Usage: $0 --app <name> [--platform <os/arch>|--all-platforms] [--force]"
    echo ""
    echo "Available apps:"
    list_all_apps | sed 's/^/  /'
    exit 1
fi

validate_app "$APP"

# Default to host platform if no platform specified
if [ "$ALL_PLATFORMS" = false ] && [ -z "$PLATFORM" ]; then
    PLATFORM="$(go env GOOS)/$(go env GOARCH)"
fi

# Build platform list
if [ "$ALL_PLATFORMS" = true ]; then
    BUILD_PLATFORMS=("${PLATFORMS[@]}")
else
    BUILD_PLATFORMS=("$PLATFORM")
fi

# Check for changes
if [ "$FORCE" = false ]; then
    if ! has_changes "$APP"; then
        echo -e "${YELLOW}No source changes detected for '$APP'. Use --force to rebuild.${NC}"
        exit 0
    fi
fi

# Detect container runtime
CONTAINER_CMD="$(detect_runtime)"
if [ -z "$CONTAINER_CMD" ]; then
    echo -e "${RED}ERROR: No container runtime found. Install Docker or Podman.${NC}"
    exit 1
fi
echo -e "${BLUE}Using container runtime: $CONTAINER_CMD${NC}"

# Podman-specific options
CONTAINER_OPTS=""
if [ "$CONTAINER_CMD" = "podman" ]; then
    CONTAINER_OPTS="--userns=host"
fi

# Compute hash and increment version code
SOURCE_HASH="$(compute_source_hash "$APP")"
VERSION_CODE="$(increment_version_code "$APP")"
SEMANTIC_VERSION="$(get_semantic_version)"
CMD_PATH="$(get_cmd_path "$APP")"
LDFLAGS="$(get_ldflags "$APP" "$VERSION_CODE" "$SOURCE_HASH" "container")"

BUILDER_IMAGE="helixagent-builder:latest"

echo ""
echo -e "${BLUE}=== Release Build: $APP ===${NC}"
echo "  Version:      $SEMANTIC_VERSION"
echo "  Version Code: $VERSION_CODE"
echo "  Source Hash:  sha256:${SOURCE_HASH:0:16}..."
echo "  Platforms:    ${BUILD_PLATFORMS[*]}"
echo ""

# Ensure builder image exists
if ! $CONTAINER_CMD image inspect "$BUILDER_IMAGE" &>/dev/null; then
    echo -e "${YELLOW}Builder image not found. Building...${NC}"
    $CONTAINER_CMD build \
        -f "$PROJECT_ROOT/docker/build/Dockerfile.builder" \
        -t "$BUILDER_IMAGE" \
        "$PROJECT_ROOT"
fi

BUILDS_OK=0
BUILDS_FAIL=0

for plat in "${BUILD_PLATFORMS[@]}"; do
    OS="${plat%/*}"
    ARCH="${plat#*/}"
    PLATFORM_DIR="$OS-$ARCH"

    BINARY_NAME="$APP"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="${APP}.exe"
    fi

    RELEASE_DIR="$PROJECT_ROOT/releases/$APP/$PLATFORM_DIR/$VERSION_CODE"
    mkdir -p "$RELEASE_DIR"

    echo -e "${BLUE}Building $APP for $OS/$ARCH...${NC}"

    if $CONTAINER_CMD run --rm \
        $CONTAINER_OPTS \
        -v "$PROJECT_ROOT:/build:Z" \
        -v "$RELEASE_DIR:/output:Z" \
        -e BUILD_APP="$APP" \
        -e BUILD_CMD_PATH="$CMD_PATH" \
        -e BUILD_GOOS="$OS" \
        -e BUILD_GOARCH="$ARCH" \
        -e BUILD_LDFLAGS="$LDFLAGS" \
        -e BUILD_OUTPUT="/output/$BINARY_NAME" \
        "$BUILDER_IMAGE"; then

        # Generate build-info.json
        local_commit="$(git -C "$PROJECT_ROOT" rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
        local_branch="$(git -C "$PROJECT_ROOT" rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')"
        local_date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
        local_go="$(go version | awk '{print $3}')"

        cat > "$RELEASE_DIR/build-info.json" <<EOF
{
  "app": "$APP",
  "version": "$SEMANTIC_VERSION",
  "version_code": $VERSION_CODE,
  "git_commit": "$local_commit",
  "git_branch": "$local_branch",
  "build_date": "$local_date",
  "platform": "$OS/$ARCH",
  "go_version": "$local_go",
  "source_hash": "sha256:$SOURCE_HASH",
  "builder": "container"
}
EOF

        # Update latest symlink
        LATEST_LINK="$PROJECT_ROOT/releases/$APP/$PLATFORM_DIR/latest"
        rm -f "$LATEST_LINK"
        ln -s "$VERSION_CODE" "$LATEST_LINK"

        echo -e "  ${GREEN}OK${NC} $OS/$ARCH -> releases/$APP/$PLATFORM_DIR/$VERSION_CODE/$BINARY_NAME"
        BUILDS_OK=$((BUILDS_OK + 1))
    else
        echo -e "  ${RED}FAIL${NC} $OS/$ARCH"
        BUILDS_FAIL=$((BUILDS_FAIL + 1))
    fi
done

# Save hash on success
if [ "$BUILDS_FAIL" -eq 0 ]; then
    save_hash "$APP" "$SOURCE_HASH"
fi

echo ""
echo -e "${BLUE}=== Build Summary ===${NC}"
echo -e "  App:          $APP"
echo -e "  Version:      $SEMANTIC_VERSION (code: $VERSION_CODE)"
echo -e "  ${GREEN}Succeeded:${NC}  $BUILDS_OK"
if [ "$BUILDS_FAIL" -gt 0 ]; then
    echo -e "  ${RED}Failed:${NC}     $BUILDS_FAIL"
    exit 1
fi
echo ""
echo -e "${GREEN}Release build complete.${NC}"
