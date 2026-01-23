#!/bin/bash
# MCP Servers Build Script
# Handles both Docker and Podman with proper network configuration

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=============================================="
echo "  MCP Servers Container Build"
echo "=============================================="
echo ""

# Detect container runtime
detect_runtime() {
    if command -v podman &> /dev/null; then
        echo "podman"
    elif command -v docker &> /dev/null; then
        echo "docker"
    else
        echo ""
    fi
}

RUNTIME=$(detect_runtime)

if [ -z "$RUNTIME" ]; then
    echo -e "${RED}ERROR: No container runtime (Docker/Podman) found${NC}"
    exit 1
fi

echo "Container Runtime: $RUNTIME"
echo ""

# Pre-flight network check
echo "Checking network connectivity..."
if ! curl -s --connect-timeout 5 -I https://dl-cdn.alpinelinux.org/alpine/v3.23/main/x86_64/APKINDEX.tar.gz > /dev/null 2>&1; then
    echo -e "${RED}ERROR: Cannot reach Alpine package repository${NC}"
    echo "Please check your network connection."
    exit 1
fi
echo -e "${GREEN}Network OK${NC}"
echo ""

# Test container DNS resolution
echo "Testing container DNS resolution..."
if ! $RUNTIME run --rm --network=host alpine:latest sh -c "apk update > /dev/null 2>&1" 2>/dev/null; then
    echo -e "${YELLOW}WARNING: Container DNS may have issues, using host network mode${NC}"
fi
echo -e "${GREEN}Container DNS OK with host network${NC}"
echo ""

# Build the image
echo "Building MCP Servers container..."
echo ""

# Use host network for build to avoid DNS issues
if [ "$RUNTIME" = "podman" ]; then
    $RUNTIME build --network=host -t helixagent-mcp-servers:latest .
else
    # Docker also supports --network=host for builds
    DOCKER_BUILDKIT=1 $RUNTIME build --network=host -t helixagent-mcp-servers:latest .
fi

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}=============================================="
    echo "  Build Successful!"
    echo "==============================================${NC}"
    echo ""
    echo "To start the MCP servers:"
    echo "  $RUNTIME run -d --name helixagent-mcp-servers \\"
    echo "    -p 3001-3020:3001-3020 \\"
    echo "    helixagent-mcp-servers:latest"
    echo ""
    echo "Or use docker-compose/podman-compose:"
    echo "  ${RUNTIME}-compose up -d"
else
    echo ""
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
