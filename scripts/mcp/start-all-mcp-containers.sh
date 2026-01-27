#!/bin/bash
# =============================================================================
# MCP Server Container Orchestration Script
# Starts all 65+ MCP containers with proper dependency ordering
#
# Usage:
#   ./scripts/mcp/start-all-mcp-containers.sh           # Start all MCPs
#   ./scripts/mcp/start-all-mcp-containers.sh core      # Start only core MCPs
#   ./scripts/mcp/start-all-mcp-containers.sh stop      # Stop all MCPs
#   ./scripts/mcp/start-all-mcp-containers.sh status    # Show status
#   ./scripts/mcp/start-all-mcp-containers.sh validate  # Run validation
#
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/docker/mcp/docker-compose.mcp-full.yml"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Detect container runtime
detect_runtime() {
    if command -v podman &> /dev/null && podman info &> /dev/null 2>&1; then
        echo "podman"
    elif command -v docker &> /dev/null && docker info &> /dev/null 2>&1; then
        echo "docker"
    else
        echo ""
    fi
}

RUNTIME=$(detect_runtime)

if [[ -z "$RUNTIME" ]]; then
    echo -e "${RED}Error: Neither Docker nor Podman is available${NC}"
    echo "Please install Docker or Podman and ensure the daemon is running"
    exit 1
fi

COMPOSE_CMD="$RUNTIME-compose"
if ! command -v $COMPOSE_CMD &> /dev/null; then
    if [[ "$RUNTIME" == "podman" ]]; then
        COMPOSE_CMD="podman compose"
    else
        COMPOSE_CMD="docker compose"
    fi
fi

echo -e "${BLUE}Using container runtime: ${RUNTIME}${NC}"
echo -e "${BLUE}Using compose command: ${COMPOSE_CMD}${NC}"

# Core MCP services (always started first)
CORE_MCPS="mcp-fetch mcp-git mcp-time mcp-filesystem mcp-memory mcp-everything mcp-sequential-thinking mcp-sqlite mcp-puppeteer mcp-postgres"

# Database MCP services
DATABASE_MCPS="mcp-mongodb mcp-redis mcp-mysql mcp-elasticsearch mcp-supabase"

# Vector DB MCP services
VECTOR_MCPS="mcp-qdrant mcp-chroma mcp-pinecone mcp-weaviate"

# DevOps MCP services
DEVOPS_MCPS="mcp-github mcp-gitlab mcp-sentry mcp-kubernetes mcp-docker mcp-ansible mcp-aws mcp-gcp mcp-heroku mcp-cloudflare mcp-vercel mcp-workers mcp-jetbrains mcp-k8s-alt"

# Browser MCP services
BROWSER_MCPS="mcp-playwright mcp-browserbase mcp-firecrawl mcp-crawl4ai"

# Communication MCP services
COMM_MCPS="mcp-slack mcp-discord mcp-telegram"

# Productivity MCP services
PROD_MCPS="mcp-notion mcp-linear mcp-jira mcp-asana mcp-trello mcp-todoist mcp-monday mcp-airtable mcp-obsidian mcp-atlassian"

# Search/AI MCP services
SEARCH_MCPS="mcp-brave-search mcp-exa mcp-tavily mcp-perplexity mcp-kagi mcp-omnisearch mcp-context7 mcp-llamaindex mcp-langchain mcp-openai"

# Google MCP services
GOOGLE_MCPS="mcp-google-drive mcp-google-calendar mcp-google-maps mcp-youtube mcp-gmail"

# Monitoring MCP services
MONITORING_MCPS="mcp-datadog mcp-grafana mcp-prometheus"

# Finance MCP services
FINANCE_MCPS="mcp-stripe mcp-hubspot mcp-zendesk"

# Design MCP services
DESIGN_MCPS="mcp-figma"

start_core() {
    echo -e "\n${BLUE}Starting Core MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $CORE_MCPS
    echo -e "${GREEN}Core MCP servers started${NC}"
}

start_all() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}Starting All MCP Containers${NC}"
    echo -e "${BLUE}========================================${NC}"

    # Initialize submodules if needed
    echo -e "\n${BLUE}Checking git submodules...${NC}"
    cd "$PROJECT_ROOT"
    git submodule update --init --recursive 2>/dev/null || true

    # Build images
    echo -e "\n${BLUE}Building MCP container images...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" build --parallel 2>/dev/null || $COMPOSE_CMD -f "$COMPOSE_FILE" build

    # Start in tiers
    echo -e "\n${BLUE}Starting Tier 1: Core MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $CORE_MCPS
    sleep 5

    echo -e "\n${BLUE}Starting Tier 2: Database MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $DATABASE_MCPS
    sleep 3

    echo -e "\n${BLUE}Starting Tier 3: Vector DB MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $VECTOR_MCPS
    sleep 3

    echo -e "\n${BLUE}Starting Tier 4: DevOps MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $DEVOPS_MCPS
    sleep 3

    echo -e "\n${BLUE}Starting Tier 5: Browser MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $BROWSER_MCPS
    sleep 3

    echo -e "\n${BLUE}Starting Tier 6: Communication MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $COMM_MCPS
    sleep 2

    echo -e "\n${BLUE}Starting Tier 7: Productivity MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $PROD_MCPS
    sleep 3

    echo -e "\n${BLUE}Starting Tier 8: Search/AI MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $SEARCH_MCPS
    sleep 3

    echo -e "\n${BLUE}Starting Tier 9: Google MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $GOOGLE_MCPS
    sleep 2

    echo -e "\n${BLUE}Starting Tier 10: Monitoring MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $MONITORING_MCPS
    sleep 2

    echo -e "\n${BLUE}Starting Tier 11: Finance MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $FINANCE_MCPS
    sleep 2

    echo -e "\n${BLUE}Starting Tier 12: Design MCP servers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $DESIGN_MCPS

    echo -e "\n${GREEN}========================================${NC}"
    echo -e "${GREEN}All MCP Containers Started${NC}"
    echo -e "${GREEN}========================================${NC}"
}

stop_all() {
    echo -e "\n${BLUE}Stopping all MCP containers...${NC}"
    $COMPOSE_CMD -f "$COMPOSE_FILE" down
    echo -e "${GREEN}All MCP containers stopped${NC}"
}

show_status() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}MCP Container Status${NC}"
    echo -e "${BLUE}========================================${NC}"

    $COMPOSE_CMD -f "$COMPOSE_FILE" ps

    # Count running containers
    running=$($COMPOSE_CMD -f "$COMPOSE_FILE" ps --status running 2>/dev/null | grep -c 'mcp-' || echo 0)
    total=$($COMPOSE_CMD -f "$COMPOSE_FILE" config --services 2>/dev/null | wc -l || echo 0)

    echo -e "\n${BLUE}Summary:${NC}"
    echo -e "  Running: ${GREEN}$running${NC}"
    echo -e "  Total:   ${BLUE}$total${NC}"
}

show_logs() {
    local service=${1:-}
    if [[ -n "$service" ]]; then
        $COMPOSE_CMD -f "$COMPOSE_FILE" logs -f "$service"
    else
        $COMPOSE_CMD -f "$COMPOSE_FILE" logs -f --tail=100
    fi
}

check_health() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}MCP Container Health Check${NC}"
    echo -e "${BLUE}========================================${NC}"

    local healthy=0
    local unhealthy=0
    local unknown=0

    # Check core ports
    core_ports=(9101 9102 9103 9104 9105 9106 9107 9108 9109 9110)
    core_names=("fetch" "git" "time" "filesystem" "memory" "everything" "sequential-thinking" "sqlite" "puppeteer" "postgres")

    for i in "${!core_ports[@]}"; do
        port=${core_ports[$i]}
        name=${core_names[$i]}
        if timeout 2 bash -c "cat < /dev/null > /dev/tcp/localhost/$port" 2>/dev/null; then
            echo -e "  ${GREEN}[OK]${NC} mcp-$name (port $port)"
            ((healthy++))
        else
            echo -e "  ${RED}[FAIL]${NC} mcp-$name (port $port)"
            ((unhealthy++))
        fi
    done

    echo -e "\n${BLUE}Core MCP Health:${NC}"
    echo -e "  Healthy:   ${GREEN}$healthy${NC}"
    echo -e "  Unhealthy: ${RED}$unhealthy${NC}"
}

validate() {
    echo -e "\n${BLUE}Running MCP containerization validation...${NC}"
    RUN_CONTAINER_TESTS=1 "$PROJECT_ROOT/challenges/scripts/mcp_containerized_challenge.sh"
}

regenerate_configs() {
    echo -e "\n${BLUE}Regenerating CLI agent configurations...${NC}"

    # Check if HelixAgent binary exists
    if [[ -x "$PROJECT_ROOT/bin/helixagent" ]]; then
        echo -e "${BLUE}Generating OpenCode config...${NC}"
        "$PROJECT_ROOT/bin/helixagent" --generate-opencode-config --opencode-output="$HOME/.config/opencode/opencode.json"

        echo -e "${BLUE}Generating Crush config...${NC}"
        "$PROJECT_ROOT/bin/helixagent" --generate-agent-config=crush --agent-config-output="$HOME/.config/crush/crush.json"

        echo -e "${GREEN}CLI agent configurations regenerated${NC}"
    else
        echo -e "${YELLOW}HelixAgent binary not found. Run 'make build' first.${NC}"
    fi
}

usage() {
    echo "MCP Container Orchestration Script"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  start       Start all MCP containers (default)"
    echo "  core        Start only core MCP containers"
    echo "  stop        Stop all MCP containers"
    echo "  restart     Restart all MCP containers"
    echo "  status      Show container status"
    echo "  health      Check container health"
    echo "  logs        Show container logs (optionally specify service)"
    echo "  validate    Run containerization validation"
    echo "  regen       Regenerate CLI agent configs"
    echo "  help        Show this help"
    echo ""
    echo "Examples:"
    echo "  $0                  # Start all containers"
    echo "  $0 core             # Start only core MCPs"
    echo "  $0 logs mcp-fetch   # Show logs for mcp-fetch"
    echo "  $0 health           # Check container health"
    echo ""
    echo "Port Ranges:"
    echo "  9101-9120: Core (fetch, git, time, filesystem, etc.)"
    echo "  9201-9220: Database (mongodb, redis, mysql, etc.)"
    echo "  9301-9320: Vector (qdrant, chroma, pinecone, etc.)"
    echo "  9401-9440: DevOps (github, gitlab, kubernetes, etc.)"
    echo "  9501-9520: Browser (playwright, browserbase, etc.)"
    echo "  9601-9620: Communication (slack, discord, telegram)"
    echo "  9701-9740: Productivity (notion, linear, jira, etc.)"
    echo "  9801-9840: Search/AI (brave-search, perplexity, etc.)"
    echo "  9901-9999: Google/Monitoring/Finance/Design"
}

# Main
case "${1:-start}" in
    start)
        start_all
        ;;
    core)
        start_core
        ;;
    stop)
        stop_all
        ;;
    restart)
        stop_all
        sleep 2
        start_all
        ;;
    status)
        show_status
        ;;
    health)
        check_health
        ;;
    logs)
        show_logs "${2:-}"
        ;;
    validate)
        validate
        ;;
    regen|regenerate)
        regenerate_configs
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        usage
        exit 1
        ;;
esac
