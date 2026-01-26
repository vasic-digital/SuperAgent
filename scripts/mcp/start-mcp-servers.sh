#!/bin/bash
# MCP Servers Manager
# Builds and runs MCP servers from local git submodules using Docker/Podman Compose

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/docker/mcp/docker-compose.mcp-servers.yml"

# Detect container runtime
if command -v podman-compose &> /dev/null; then
    COMPOSE_CMD="podman-compose"
elif command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
elif command -v docker &> /dev/null && docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
else
    echo "Error: Neither podman-compose nor docker-compose found"
    exit 1
fi

echo "Using compose command: $COMPOSE_CMD"

# Function to show usage
usage() {
    cat << EOF
MCP Servers Manager - Build and run MCP servers from local git submodules

Usage: $0 [OPTION]

Options:
  --start           Start all MCP servers (builds if needed)
  --start-core      Start only core MCP servers (fetch, git, time, etc.)
  --start-db        Start only database MCP servers (redis, mongodb, qdrant)
  --stop            Stop all MCP servers
  --restart         Restart all MCP servers
  --build           Build all MCP server images
  --rebuild         Force rebuild all images (no cache)
  --status          Show status of MCP servers
  --logs [service]  Show logs (optionally for specific service)
  --ports           Show MCP server ports
  --help            Show this help message

Core MCP Servers (ports 9101-9107):
  - mcp-fetch          :9101  HTTP fetch operations
  - mcp-git            :9102  Git repository operations
  - mcp-time           :9103  Time utilities
  - mcp-filesystem     :9104  File system access
  - mcp-memory         :9105  In-memory storage
  - mcp-everything     :9106  Test/demo MCP
  - mcp-sequentialthinking :9107  Sequential thinking

Database MCP Servers (ports 9201-9301):
  - mcp-redis          :9201  Redis operations (backend: 16379)
  - mcp-mongodb        :9202  MongoDB operations (backend: 27017)
  - mcp-qdrant         :9301  Qdrant vector DB (backend: 6333)

DevOps MCP Servers (ports 9401-9402):
  - mcp-kubernetes     :9401  Kubernetes operations
  - mcp-github         :9402  GitHub operations

Browser MCP Servers (ports 9501):
  - mcp-playwright     :9501  Browser automation

Communication MCP Servers (ports 9601):
  - mcp-slack          :9601  Slack integration

Environment Variables:
  GITHUB_TOKEN     - GitHub personal access token (for mcp-github)
  SLACK_BOT_TOKEN  - Slack bot token (for mcp-slack)
  SLACK_TEAM_ID    - Slack team ID (for mcp-slack)

EOF
}

# Function to show ports
show_ports() {
    cat << EOF
MCP Server Ports:

Core Servers:
  mcp-fetch              localhost:9101
  mcp-git                localhost:9102
  mcp-time               localhost:9103
  mcp-filesystem         localhost:9104
  mcp-memory             localhost:9105
  mcp-everything         localhost:9106
  mcp-sequentialthinking localhost:9107

Database Servers:
  mcp-redis              localhost:9201  (backend: 16379)
  mcp-mongodb            localhost:9202  (backend: 27017)
  mcp-qdrant             localhost:9301  (backend: 6333, 6334)

DevOps Servers:
  mcp-kubernetes         localhost:9401
  mcp-github             localhost:9402

Browser Servers:
  mcp-playwright         localhost:9501

Communication Servers:
  mcp-slack              localhost:9601
EOF
}

# Core services list
CORE_SERVICES="mcp-fetch mcp-git mcp-time mcp-filesystem mcp-memory mcp-everything mcp-sequentialthinking"
DB_SERVICES="mcp-redis mcp-redis-backend mcp-mongodb mcp-mongodb-backend mcp-qdrant mcp-qdrant-backend"

case "${1:-}" in
    --start)
        echo "Starting all MCP servers..."
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$COMPOSE_FILE" up -d
        echo "MCP servers started. Use '$0 --status' to check status."
        ;;
    --start-core)
        echo "Starting core MCP servers..."
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $CORE_SERVICES
        echo "Core MCP servers started."
        ;;
    --start-db)
        echo "Starting database MCP servers..."
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $DB_SERVICES
        echo "Database MCP servers started."
        ;;
    --stop)
        echo "Stopping all MCP servers..."
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$COMPOSE_FILE" down
        echo "MCP servers stopped."
        ;;
    --restart)
        echo "Restarting all MCP servers..."
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$COMPOSE_FILE" restart
        echo "MCP servers restarted."
        ;;
    --build)
        echo "Building MCP server images..."
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$COMPOSE_FILE" build
        echo "Build complete."
        ;;
    --rebuild)
        echo "Rebuilding MCP server images (no cache)..."
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$COMPOSE_FILE" build --no-cache
        echo "Rebuild complete."
        ;;
    --status)
        echo "MCP Server Status:"
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$COMPOSE_FILE" ps
        echo ""
        echo "Port connectivity:"
        for port in 9101 9102 9103 9104 9105 9106 9107 9201 9202 9301 9401 9402 9501 9601; do
            if timeout 1 bash -c "</dev/tcp/localhost/$port" 2>/dev/null; then
                echo "  Port $port: ✅ OPEN"
            else
                echo "  Port $port: ❌ CLOSED"
            fi
        done
        ;;
    --logs)
        cd "$PROJECT_ROOT"
        if [ -n "${2:-}" ]; then
            $COMPOSE_CMD -f "$COMPOSE_FILE" logs -f "$2"
        else
            $COMPOSE_CMD -f "$COMPOSE_FILE" logs -f
        fi
        ;;
    --ports)
        show_ports
        ;;
    --help|"")
        usage
        ;;
    *)
        echo "Unknown option: $1"
        usage
        exit 1
        ;;
esac
