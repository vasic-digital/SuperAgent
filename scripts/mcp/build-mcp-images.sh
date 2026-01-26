#!/bin/bash

# Build script for MCP server Docker images
# Uses --network=host to work around podman bridge network issues
#
# Usage: ./scripts/mcp/build-mcp-images.sh [SERVICE_NAME]
#   Without arguments: builds all MCP servers
#   With argument: builds only the specified service

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMPOSE_FILE="$PROJECT_DIR/docker/mcp/docker-compose.mcp-servers.yml"

# Detect container runtime
if command -v podman &> /dev/null; then
    CONTAINER_RUNTIME="podman"
    # podman needs --network=host during build due to bridge network DNS issues
    BUILD_NETWORK="--network=host"
elif command -v docker &> /dev/null; then
    CONTAINER_RUNTIME="docker"
    BUILD_NETWORK=""
else
    echo "Error: Neither docker nor podman found"
    exit 1
fi

echo "Using container runtime: $CONTAINER_RUNTIME"
echo "Project directory: $PROJECT_DIR"

# Define all MCP services and their Dockerfiles
# Format: service_name=dockerfile:BUILD_ARG1=value:BUILD_ARG2=value
declare -A MCP_SERVICES=(
    # Core MCP servers from MCP-Servers monorepo
    ["mcp-fetch"]="Dockerfile.mcp-server:SERVER_NAME=fetch:SOURCE_DIR=MCP-Servers"
    ["mcp-git"]="Dockerfile.mcp-server:SERVER_NAME=git:SOURCE_DIR=MCP-Servers"
    ["mcp-time"]="Dockerfile.mcp-server:SERVER_NAME=time:SOURCE_DIR=MCP-Servers"
    ["mcp-filesystem"]="Dockerfile.mcp-server:SERVER_NAME=filesystem:SOURCE_DIR=MCP-Servers"
    ["mcp-memory"]="Dockerfile.mcp-server:SERVER_NAME=memory:SOURCE_DIR=MCP-Servers"
    ["mcp-everything"]="Dockerfile.mcp-server:SERVER_NAME=everything:SOURCE_DIR=MCP-Servers"
    ["mcp-sequentialthinking"]="Dockerfile.mcp-server:SERVER_NAME=sequentialthinking:SOURCE_DIR=MCP-Servers"

    # GitHub/GitLab MCP servers
    ["mcp-github"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/github-mcp-server"

    # Communication MCP servers
    ["mcp-slack"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/slack-mcp"
    ["mcp-telegram"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/telegram-mcp"

    # Search MCP servers
    ["mcp-brave-search"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/brave-search"

    # Browser automation
    ["mcp-playwright"]="Dockerfile.mcp-playwright:SOURCE_DIR=MCP/submodules/playwright-mcp"
    ["mcp-browserbase"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/browserbase-mcp"

    # Cloud providers
    ["mcp-cloudflare"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/cloudflare-mcp"
    ["mcp-aws"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/aws-mcp"
    ["mcp-heroku"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/heroku-mcp"

    # DevOps/Infrastructure
    ["mcp-kubernetes"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/kubernetes-mcp"
    ["mcp-k8s"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/k8s-mcp-server"

    # Databases
    ["mcp-mongodb"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/mongodb-mcp"
    ["mcp-redis"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/redis-mcp"
    ["mcp-elasticsearch"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/elasticsearch-mcp"
    ["mcp-supabase"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/supabase-mcp"

    # Vector databases
    ["mcp-qdrant"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/qdrant-mcp"

    # Productivity
    ["mcp-notion"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/notion-mcp-server"
    ["mcp-obsidian"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/obsidian-mcp"
    ["mcp-trello"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/trello-mcp"
    ["mcp-airtable"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/airtable-mcp"
    ["mcp-atlassian"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/atlassian-mcp"

    # AI/Search
    ["mcp-perplexity"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/perplexity-mcp"
    ["mcp-firecrawl"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/firecrawl-mcp"
    ["mcp-omnisearch"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/omnisearch-mcp"

    # Monitoring
    ["mcp-sentry"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/sentry-mcp"

    # Microsoft
    ["mcp-microsoft"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/microsoft-mcp"

    # Context/Docs
    ["mcp-context7"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/context7-mcp"
    ["mcp-docs"]="Dockerfile.mcp-submodule:SOURCE_DIR=MCP/submodules/docs-mcp"
)

build_service() {
    local service=$1
    local config=${MCP_SERVICES[$service]}

    if [ -z "$config" ]; then
        echo "Error: Unknown service $service"
        return 1
    fi

    IFS=':' read -ra PARTS <<< "$config"
    local dockerfile="${PARTS[0]}"
    local build_args=""

    for ((i=1; i<${#PARTS[@]}; i++)); do
        build_args="$build_args --build-arg ${PARTS[$i]}"
    done

    # Check if source directory exists
    local source_dir=""
    for part in "${PARTS[@]}"; do
        if [[ "$part" == SOURCE_DIR=* ]]; then
            source_dir="${part#SOURCE_DIR=}"
            break
        fi
    done

    if [ -n "$source_dir" ] && [ ! -d "$PROJECT_DIR/$source_dir" ]; then
        echo "Warning: Source directory $source_dir does not exist, skipping $service"
        return 1
    fi

    echo ""
    echo "=========================================="
    echo "Building $service"
    echo "  Dockerfile: docker/mcp/$dockerfile"
    echo "  Build args: $build_args"
    echo "=========================================="

    if $CONTAINER_RUNTIME build $BUILD_NETWORK \
        -f "$PROJECT_DIR/docker/mcp/$dockerfile" \
        -t "mcp_$service:latest" \
        $build_args \
        "$PROJECT_DIR"; then
        echo "Successfully built $service"
        return 0
    else
        echo "Failed to build $service"
        return 1
    fi
}

# Check if specific service requested
if [ -n "$1" ]; then
    if [[ -v MCP_SERVICES[$1] ]]; then
        build_service "$1"
    else
        echo "Error: Unknown service '$1'"
        echo "Available services:"
        for service in "${!MCP_SERVICES[@]}"; do
            echo "  - $service"
        done | sort
        exit 1
    fi
else
    # Build all services
    echo "Building all MCP server images..."
    echo ""

    BUILT=0
    FAILED=0
    SKIPPED=0

    for service in $(echo "${!MCP_SERVICES[@]}" | tr ' ' '\n' | sort); do
        if build_service "$service"; then
            ((BUILT++))
        else
            ((FAILED++))
            echo "Warning: Failed to build $service"
        fi
    done

    echo ""
    echo "=========================================="
    echo "Build Summary"
    echo "  Built: $BUILT"
    echo "  Failed: $FAILED"
    echo "=========================================="

    if [ $FAILED -gt 0 ]; then
        exit 1
    fi
fi
