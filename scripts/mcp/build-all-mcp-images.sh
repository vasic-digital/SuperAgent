#!/bin/bash
# =============================================================================
# Build ALL MCP Server Docker Images
# Builds all 45+ MCP servers from submodules and the main MCP-Servers repo
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$PROJECT_DIR"

echo "========================================================================"
echo "               Building ALL MCP Server Docker Images"
echo "========================================================================"
echo ""

# Check for container runtime
if command -v podman &> /dev/null; then
    CONTAINER_CMD="podman"
    BUILD_OPTS="--network=host"
elif command -v docker &> /dev/null; then
    CONTAINER_CMD="docker"
    BUILD_OPTS=""
else
    echo "Error: Neither podman nor docker found!"
    exit 1
fi

echo "Using container runtime: $CONTAINER_CMD"
echo ""

# Counter
BUILT=0
FAILED=0
SKIPPED=0

build_image() {
    local name="$1"
    local dockerfile="$2"
    local context="$3"
    local build_args="$4"

    echo "━━━ Building: $name ━━━"

    if $CONTAINER_CMD build $BUILD_OPTS -t "$name:latest" -f "$dockerfile" $build_args "$context" 2>&1 | grep -E "(Successfully|Error|fatal)" | tail -2; then
        ((BUILT++))
        echo "✓ $name built successfully"
    else
        ((FAILED++))
        echo "✗ $name build failed"
    fi
    echo ""
}

# ===========================================================================
# TIER 1: Core MCP Servers (from MCP-Servers repo)
# ===========================================================================
echo "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
echo "┃  TIER 1: Core MCP Servers (MCP-Servers repo)                       ┃"
echo "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
echo ""

# TypeScript Core Servers
for server in filesystem memory everything sequentialthinking; do
    build_image "mcp-$server" \
        "docker/mcp/Dockerfile.mcp-server" \
        "." \
        "--build-arg SERVER_NAME=$server --build-arg SOURCE_DIR=MCP-Servers"
done

# Python Core Servers
for server in fetch git time; do
    build_image "mcp-$server" \
        "docker/mcp/Dockerfile.mcp-server-python" \
        "." \
        "--build-arg SERVER_NAME=$server --build-arg SOURCE_DIR=MCP-Servers"
done

# ===========================================================================
# TIER 2: Enterprise Integration Servers
# ===========================================================================
echo "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
echo "┃  TIER 2: Enterprise Integration Servers                            ┃"
echo "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
echo ""

# MongoDB (TypeScript)
if [ -d "MCP/submodules/mongodb-mcp" ]; then
    build_image "mcp-mongodb" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/mongodb-mcp"
fi

# Redis (Python)
if [ -d "MCP/submodules/redis-mcp" ]; then
    build_image "mcp-redis" \
        "docker/mcp/Dockerfile.mcp-python" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/redis-mcp"
fi

# GitHub (Go)
if [ -d "MCP/submodules/github-mcp-server" ]; then
    build_image "mcp-github" \
        "docker/mcp/Dockerfile.mcp-go" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/github-mcp-server"
fi

# Slack (Go)
if [ -d "MCP/submodules/slack-mcp" ]; then
    build_image "mcp-slack" \
        "docker/mcp/Dockerfile.mcp-go" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/slack-mcp"
fi

# Notion (TypeScript)
if [ -d "MCP/submodules/notion-mcp-server" ]; then
    build_image "mcp-notion" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/notion-mcp-server"
fi

# Trello (TypeScript/Bun)
if [ -d "MCP/submodules/trello-mcp" ]; then
    build_image "mcp-trello" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/trello-mcp"
fi

# Kubernetes (Go)
if [ -d "MCP/submodules/kubernetes-mcp" ]; then
    build_image "mcp-kubernetes" \
        "docker/mcp/Dockerfile.mcp-go" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/kubernetes-mcp"
fi

# ===========================================================================
# TIER 3: Data & Vector Stores
# ===========================================================================
echo "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
echo "┃  TIER 3: Data & Vector Stores                                      ┃"
echo "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
echo ""

# Qdrant (Python)
if [ -d "MCP/submodules/qdrant-mcp" ]; then
    build_image "mcp-qdrant" \
        "docker/mcp/Dockerfile.mcp-python" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/qdrant-mcp"
fi

# Supabase (TypeScript)
if [ -d "MCP/submodules/supabase-mcp" ]; then
    build_image "mcp-supabase" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/supabase-mcp"
fi

# Atlassian (Python)
if [ -d "MCP/submodules/atlassian-mcp" ]; then
    build_image "mcp-atlassian" \
        "docker/mcp/Dockerfile.mcp-python" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/atlassian-mcp"
fi

# ===========================================================================
# TIER 4: Browser & Web
# ===========================================================================
echo "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
echo "┃  TIER 4: Browser & Web                                             ┃"
echo "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
echo ""

# Browserbase (TypeScript)
if [ -d "MCP/submodules/browserbase-mcp" ]; then
    build_image "mcp-browserbase" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/browserbase-mcp"
fi

# Firecrawl (TypeScript)
if [ -d "MCP/submodules/firecrawl-mcp" ]; then
    build_image "mcp-firecrawl" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/firecrawl-mcp"
fi

# Brave Search (TypeScript)
if [ -d "MCP/submodules/brave-search" ]; then
    build_image "mcp-brave-search" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/brave-search"
fi

# Playwright (TypeScript)
if [ -d "MCP/submodules/playwright-mcp" ]; then
    build_image "mcp-playwright" \
        "docker/mcp/Dockerfile.mcp-playwright" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/playwright-mcp"
fi

# ===========================================================================
# TIER 5: Communication
# ===========================================================================
echo "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
echo "┃  TIER 5: Communication                                             ┃"
echo "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
echo ""

# Telegram (Python)
if [ -d "MCP/submodules/telegram-mcp" ]; then
    build_image "mcp-telegram" \
        "docker/mcp/Dockerfile.mcp-python" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/telegram-mcp"
fi

# ===========================================================================
# TIER 6: Productivity
# ===========================================================================
echo "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
echo "┃  TIER 6: Productivity                                              ┃"
echo "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
echo ""

# Airtable (TypeScript)
if [ -d "MCP/submodules/airtable-mcp" ]; then
    build_image "mcp-airtable" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/airtable-mcp"
fi

# Obsidian (TypeScript)
if [ -d "MCP/submodules/obsidian-mcp" ]; then
    build_image "mcp-obsidian" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/obsidian-mcp"
fi

# ===========================================================================
# TIER 7: Cloud & Deployment
# ===========================================================================
echo "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
echo "┃  TIER 7: Cloud & Deployment                                        ┃"
echo "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
echo ""

# Heroku (TypeScript)
if [ -d "MCP/submodules/heroku-mcp" ]; then
    build_image "mcp-heroku" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/heroku-mcp"
fi

# Cloudflare (TypeScript)
if [ -d "MCP/submodules/cloudflare-mcp" ]; then
    build_image "mcp-cloudflare" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/cloudflare-mcp"
fi

# Workers (TypeScript)
if [ -d "MCP/submodules/workers-mcp" ]; then
    build_image "mcp-workers" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/workers-mcp"
fi

# ===========================================================================
# TIER 8: AI & Search
# ===========================================================================
echo "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
echo "┃  TIER 8: AI & Search                                               ┃"
echo "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
echo ""

# Perplexity (TypeScript)
if [ -d "MCP/submodules/perplexity-mcp" ]; then
    build_image "mcp-perplexity" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/perplexity-mcp"
fi

# Omnisearch (TypeScript)
if [ -d "MCP/submodules/omnisearch-mcp" ]; then
    build_image "mcp-omnisearch" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/omnisearch-mcp"
fi

# Context7 (TypeScript)
if [ -d "MCP/submodules/context7-mcp" ]; then
    build_image "mcp-context7" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/context7-mcp"
fi

# LlamaIndex (TypeScript)
if [ -d "MCP/submodules/llamaindex-mcp" ]; then
    build_image "mcp-llamaindex" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/llamaindex-mcp"
fi

# LangChain (Python)
if [ -d "MCP/submodules/langchain-mcp" ]; then
    build_image "mcp-langchain" \
        "docker/mcp/Dockerfile.mcp-python" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/langchain-mcp"
fi

# ===========================================================================
# TIER 9: DevOps & Monitoring
# ===========================================================================
echo "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
echo "┃  TIER 9: DevOps & Monitoring                                       ┃"
echo "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
echo ""

# Sentry (TypeScript)
if [ -d "MCP/submodules/sentry-mcp" ]; then
    build_image "mcp-sentry" \
        "docker/mcp/Dockerfile.mcp-submodule" \
        "." \
        "--build-arg SOURCE_DIR=MCP/submodules/sentry-mcp"
fi

# Microsoft (TypeScript/C#)
if [ -d "MCP/submodules/microsoft-mcp" ]; then
    echo "Skipping microsoft-mcp (requires .NET build)"
    ((SKIPPED++))
fi

# ===========================================================================
# Summary
# ===========================================================================
echo ""
echo "========================================================================"
echo "                         Build Summary                                  "
echo "========================================================================"
echo ""
echo "  Built:   $BUILT"
echo "  Failed:  $FAILED"
echo "  Skipped: $SKIPPED"
echo ""

echo "MCP images:"
$CONTAINER_CMD images | grep "mcp-" | head -30
echo ""

echo "To start all servers:"
echo "  podman-compose -f docker/mcp/docker-compose.mcp-all-servers.yml up -d"
