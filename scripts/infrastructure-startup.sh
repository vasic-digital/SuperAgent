#!/bin/bash
# HelixAgent Infrastructure Startup Script
# Starts all required services using Docker/Podman Compose

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Detect container runtime
detect_runtime() {
    if command -v podman &> /dev/null && podman info &> /dev/null; then
        echo "podman"
    elif command -v docker &> /dev/null && docker info &> /dev/null; then
        echo "docker"
    else
        echo ""
    fi
}

RUNTIME=$(detect_runtime)
COMPOSE_CMD=""

if [ -z "$RUNTIME" ]; then
    echo -e "${RED}Error: No container runtime available (Docker or Podman)${NC}"
    exit 1
fi

if [ "$RUNTIME" = "podman" ]; then
    COMPOSE_CMD="podman-compose"
    if ! command -v podman-compose &> /dev/null; then
        COMPOSE_CMD="podman compose"
    fi
else
    COMPOSE_CMD="docker compose"
    if ! docker compose version &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    fi
fi

echo -e "${CYAN}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║         HelixAgent Infrastructure Startup                      ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════════╝${NC}"
echo -e "${BLUE}Runtime: ${RUNTIME} | Compose: ${COMPOSE_CMD}${NC}"
echo ""

# Parse arguments
PROFILE="${1:-default}"
WAIT_HEALTHY="${2:-true}"

usage() {
    echo "Usage: $0 [PROFILE] [WAIT_HEALTHY]"
    echo ""
    echo "Profiles:"
    echo "  default    - Core services (postgres, redis, chromadb, cognee)"
    echo "  full       - All services including monitoring"
    echo "  lsp        - LSP servers"
    echo "  rag        - RAG and embedding services"
    echo "  mcp        - MCP servers"
    echo "  all        - Everything"
    echo ""
    echo "Options:"
    echo "  WAIT_HEALTHY - Wait for services to be healthy (true/false, default: true)"
}

# Create network if not exists
create_network() {
    echo -e "${BLUE}Creating network helixagent-network...${NC}"
    if [ "$RUNTIME" = "podman" ]; then
        podman network exists helixagent-network 2>/dev/null || podman network create helixagent-network
    else
        docker network inspect helixagent-network &>/dev/null || docker network create helixagent-network
    fi
}

# Create required volumes
create_volumes() {
    echo -e "${BLUE}Creating volumes...${NC}"
    volumes=(
        "mcp_workspace"
        "lsp_cache"
        "go_cache"
        "cargo_cache"
        "pip_cache"
        "npm_cache"
        "maven_cache"
    )
    for vol in "${volumes[@]}"; do
        if [ "$RUNTIME" = "podman" ]; then
            podman volume exists "$vol" 2>/dev/null || podman volume create "$vol"
        else
            docker volume inspect "$vol" &>/dev/null || docker volume create "$vol"
        fi
    done
}

# Start core services
start_core() {
    echo -e "${GREEN}Starting core services (postgres, redis, chromadb)...${NC}"
    cd "$PROJECT_ROOT"
    $COMPOSE_CMD up -d postgres redis chromadb
}

# Start Cognee
start_cognee() {
    echo -e "${GREEN}Starting Cognee...${NC}"
    cd "$PROJECT_ROOT"
    $COMPOSE_CMD --profile ai up -d cognee
}

# Start LSP servers
start_lsp() {
    echo -e "${GREEN}Starting LSP servers...${NC}"
    cd "$PROJECT_ROOT/docker/lsp"
    $COMPOSE_CMD -f docker-compose.lsp.yml --profile lsp up -d
}

# Start RAG services
start_rag() {
    echo -e "${GREEN}Starting RAG services...${NC}"
    cd "$PROJECT_ROOT/docker/rag"
    $COMPOSE_CMD -f docker-compose.rag.yml --profile rag up -d
}

# Start MCP servers
start_mcp() {
    echo -e "${GREEN}Starting MCP servers...${NC}"
    cd "$PROJECT_ROOT/docker/mcp"
    $COMPOSE_CMD -f docker-compose.mcp.yml up -d
}

# Start monitoring
start_monitoring() {
    echo -e "${GREEN}Starting monitoring (Prometheus, Grafana)...${NC}"
    cd "$PROJECT_ROOT"
    $COMPOSE_CMD --profile monitoring up -d prometheus grafana
}

# Wait for service to be healthy
wait_for_health() {
    local service=$1
    local url=$2
    local max_attempts=${3:-30}
    local attempt=1

    echo -ne "${YELLOW}Waiting for $service..."

    while [ $attempt -le $max_attempts ]; do
        if curl -sf "$url" > /dev/null 2>&1; then
            echo -e "${GREEN} ready!${NC}"
            return 0
        fi
        sleep 2
        attempt=$((attempt + 1))
        echo -n "."
    done

    echo -e "${RED} timeout!${NC}"
    return 1
}

# Check all services
check_services() {
    echo ""
    echo -e "${CYAN}Checking service health...${NC}"

    services=(
        "PostgreSQL|localhost:5432|pg_isready"
        "Redis|localhost:6379|redis-cli ping"
        "ChromaDB|http://localhost:8001/api/v2/heartbeat"
        "Cognee|http://localhost:8000/"
    )

    for svc in "${services[@]}"; do
        IFS='|' read -r name url cmd <<< "$svc"
        if [ -n "$cmd" ] && [ "$cmd" != "http"* ]; then
            # Command-based check
            if eval "$cmd" &>/dev/null; then
                echo -e "  ${GREEN}✓${NC} $name"
            else
                echo -e "  ${RED}✗${NC} $name"
            fi
        else
            # HTTP-based check
            if curl -sf "$url" > /dev/null 2>&1; then
                echo -e "  ${GREEN}✓${NC} $name"
            else
                echo -e "  ${RED}✗${NC} $name"
            fi
        fi
    done
}

# Main execution
main() {
    create_network
    create_volumes

    case "$PROFILE" in
        default)
            start_core
            start_cognee
            ;;
        full)
            start_core
            start_cognee
            start_monitoring
            ;;
        lsp)
            start_lsp
            ;;
        rag)
            start_rag
            ;;
        mcp)
            start_mcp
            ;;
        all)
            start_core
            start_cognee
            start_lsp
            start_rag
            start_mcp
            start_monitoring
            ;;
        help|--help|-h)
            usage
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown profile: $PROFILE${NC}"
            usage
            exit 1
            ;;
    esac

    if [ "$WAIT_HEALTHY" = "true" ]; then
        echo ""
        echo -e "${BLUE}Waiting for services to become healthy...${NC}"
        sleep 5

        case "$PROFILE" in
            default|full|all)
                wait_for_health "PostgreSQL" "http://localhost:5432" 30 || true
                wait_for_health "Redis" "http://localhost:6379" 30 || true
                wait_for_health "ChromaDB" "http://localhost:8001/api/v2/heartbeat" 30 || true
                wait_for_health "Cognee" "http://localhost:8000/" 60 || true
                ;;
        esac
    fi

    check_services

    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║         Infrastructure startup complete!                       ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
}

main
