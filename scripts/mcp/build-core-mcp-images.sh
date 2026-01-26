#!/bin/bash
# =============================================================================
# Build Core MCP Server Docker Images
# Builds all 7 core MCP servers from the MCP-Servers repository
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$PROJECT_DIR"

echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║         Building Core MCP Server Docker Images                    ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo ""

# Check for container runtime
if command -v podman &> /dev/null; then
    CONTAINER_CMD="podman"
    BUILD_OPTS="--network=host"  # Podman needs host network for DNS
elif command -v docker &> /dev/null; then
    CONTAINER_CMD="docker"
    BUILD_OPTS=""
else
    echo "Error: Neither podman nor docker found!"
    exit 1
fi

echo "Using container runtime: $CONTAINER_CMD"
echo ""

# TypeScript MCP Servers (filesystem, memory, everything, sequentialthinking)
TYPESCRIPT_SERVERS="filesystem memory everything sequentialthinking"

# Python MCP Servers (fetch, git, time)
PYTHON_SERVERS="fetch git time"

build_typescript_server() {
    local server=$1
    echo "━━━ Building TypeScript MCP server: $server ━━━"
    $CONTAINER_CMD build $BUILD_OPTS -t "mcp-$server:latest" \
        -f docker/mcp/Dockerfile.mcp-server \
        --build-arg SERVER_NAME=$server \
        --build-arg SOURCE_DIR=MCP-Servers \
        . 2>&1 | grep -E "(Successfully|STEP|Error)" | tail -5
    echo ""
}

build_python_server() {
    local server=$1
    echo "━━━ Building Python MCP server: $server ━━━"
    $CONTAINER_CMD build $BUILD_OPTS -t "mcp-$server:latest" \
        -f docker/mcp/Dockerfile.mcp-server-python \
        --build-arg SERVER_NAME=$server \
        --build-arg SOURCE_DIR=MCP-Servers \
        . 2>&1 | grep -E "(Successfully|STEP|Error)" | tail -5
    echo ""
}

# Build TypeScript servers
echo "Building TypeScript MCP Servers..."
echo "═══════════════════════════════════════"
for server in $TYPESCRIPT_SERVERS; do
    build_typescript_server "$server"
done

# Build Python servers
echo "Building Python MCP Servers..."
echo "═══════════════════════════════════════"
for server in $PYTHON_SERVERS; do
    build_python_server "$server"
done

echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║                     Build Complete                                 ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo ""
echo "Built images:"
$CONTAINER_CMD images | grep "mcp-" | head -10
echo ""
echo "To start all servers:"
echo "  podman-compose -f docker/mcp/docker-compose.mcp-core.yml up -d"
echo ""
echo "To test servers:"
echo "  ./challenges/scripts/mcp_validation_comprehensive.sh --quick"
