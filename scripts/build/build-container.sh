#!/bin/bash
# HelixAgent Container Build Script
# Runs INSIDE the builder container. Receives config via environment variables.
#
# Required env vars:
#   BUILD_APP       - App name (for logging)
#   BUILD_CMD_PATH  - Go package path to build (e.g., ./cmd/helixagent)
#   BUILD_GOOS      - Target OS
#   BUILD_GOARCH    - Target architecture
#   BUILD_LDFLAGS   - Linker flags with version info
#   BUILD_OUTPUT    - Output binary path

set -e

echo "=== HelixAgent Release Builder ==="
echo "App:      ${BUILD_APP}"
echo "Platform: ${BUILD_GOOS}/${BUILD_GOARCH}"
echo "Output:   ${BUILD_OUTPUT}"
echo ""

if [ -z "$BUILD_CMD_PATH" ] || [ -z "$BUILD_GOOS" ] || [ -z "$BUILD_GOARCH" ] || [ -z "$BUILD_OUTPUT" ]; then
    echo "ERROR: Missing required environment variables."
    echo "Required: BUILD_CMD_PATH, BUILD_GOOS, BUILD_GOARCH, BUILD_OUTPUT"
    exit 1
fi

# Create output directory
mkdir -p "$(dirname "$BUILD_OUTPUT")"

echo "Building..."
CGO_ENABLED=0 GOOS="$BUILD_GOOS" GOARCH="$BUILD_GOARCH" \
    go build -mod=mod \
    -ldflags="$BUILD_LDFLAGS" \
    -o "$BUILD_OUTPUT" \
    "$BUILD_CMD_PATH"

echo "Build complete: $BUILD_OUTPUT"
ls -lh "$BUILD_OUTPUT"
