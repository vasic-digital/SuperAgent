#!/bin/bash
# MCP Servers Manager
# Builds and runs ALL MCP servers from local git submodules using Docker/Podman Compose
# Zero npm/npx dependencies - everything runs from local git submodules

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
MCP Servers Manager - Build and run ALL MCP servers from local git submodules
Zero npm/npx dependencies - everything runs from local git submodules

Usage: $0 [OPTION]

Options:
  --start           Start all MCP servers (builds if needed)
  --start-core      Start only core MCP servers (fetch, git, time, etc.)
  --start-db        Start only database MCP servers (redis, mongodb, qdrant)
  --start-devops    Start only DevOps MCP servers (kubernetes, github, etc.)
  --start-comm      Start only communication MCP servers (slack, telegram)
  --start-prod      Start only productivity MCP servers (notion, trello, etc.)
  --start-search    Start only search/AI MCP servers (brave, perplexity, etc.)
  --stop            Stop all MCP servers
  --restart         Restart all MCP servers
  --build           Build all MCP server images
  --rebuild         Force rebuild all images (no cache)
  --status          Show status of MCP servers
  --logs [service]  Show logs (optionally for specific service)
  --ports           Show MCP server ports
  --help            Show this help message

Port Ranges (32 MCP servers total):
  9101-9199: Core MCP servers (fetch, git, time, filesystem, memory, everything, sequential-thinking)
  9201-9299: Database MCP servers (redis, mongodb, supabase)
  9301-9399: Vector database MCP servers (qdrant)
  9401-9499: DevOps/Infrastructure MCP servers (kubernetes, github, cloudflare, heroku, sentry)
  9501-9599: Browser automation MCP servers (playwright, browserbase, firecrawl)
  9601-9699: Communication MCP servers (slack, telegram)
  9701-9799: Productivity MCP servers (notion, trello, airtable, obsidian, atlassian)
  9801-9899: Search/AI MCP servers (brave-search, perplexity, omnisearch, context7, llamaindex)
  9901-9999: Cloud provider MCP servers (workers)

Environment Variables (set these for MCPs that need API keys):
  GITHUB_TOKEN          - GitHub personal access token (for mcp-github)
  SLACK_BOT_TOKEN       - Slack bot token (for mcp-slack)
  SLACK_TEAM_ID         - Slack team ID (for mcp-slack)
  TELEGRAM_BOT_TOKEN    - Telegram bot token (for mcp-telegram)
  NOTION_API_KEY        - Notion API key (for mcp-notion)
  TRELLO_API_KEY        - Trello API key (for mcp-trello)
  TRELLO_TOKEN          - Trello token (for mcp-trello)
  AIRTABLE_API_KEY      - Airtable API key (for mcp-airtable)
  BRAVE_API_KEY         - Brave Search API key (for mcp-brave-search)
  PERPLEXITY_API_KEY    - Perplexity API key (for mcp-perplexity)
  CLOUDFLARE_API_TOKEN  - Cloudflare API token (for mcp-cloudflare, mcp-workers)
  CLOUDFLARE_ACCOUNT_ID - Cloudflare account ID (for mcp-cloudflare, mcp-workers)
  HEROKU_API_KEY        - Heroku API key (for mcp-heroku)
  SENTRY_AUTH_TOKEN     - Sentry auth token (for mcp-sentry)
  SENTRY_ORG            - Sentry organization (for mcp-sentry)
  BROWSERBASE_API_KEY   - Browserbase API key (for mcp-browserbase)
  BROWSERBASE_PROJECT_ID - Browserbase project ID (for mcp-browserbase)
  FIRECRAWL_API_KEY     - Firecrawl API key (for mcp-firecrawl)
  SUPABASE_URL          - Supabase URL (for mcp-supabase)
  SUPABASE_KEY          - Supabase key (for mcp-supabase)
  ATLASSIAN_URL         - Atlassian URL (for mcp-atlassian)
  ATLASSIAN_EMAIL       - Atlassian email (for mcp-atlassian)
  ATLASSIAN_API_TOKEN   - Atlassian API token (for mcp-atlassian)
  OPENAI_API_KEY        - OpenAI API key (for mcp-llamaindex)

EOF
}

# Function to show ports
show_ports() {
    cat << EOF
MCP Server Ports (32 servers total):

Core Servers (from MCP-Servers monorepo):
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
  mcp-supabase           localhost:9203

Vector Database Servers:
  mcp-qdrant             localhost:9301  (backend: 6333, 6334)

DevOps/Infrastructure Servers:
  mcp-kubernetes         localhost:9401
  mcp-github             localhost:9402
  mcp-cloudflare         localhost:9403
  mcp-heroku             localhost:9404
  mcp-sentry             localhost:9405

Browser Automation Servers:
  mcp-playwright         localhost:9501
  mcp-browserbase        localhost:9502
  mcp-firecrawl          localhost:9503

Communication Servers:
  mcp-slack              localhost:9601
  mcp-telegram           localhost:9602

Productivity Servers:
  mcp-notion             localhost:9701
  mcp-trello             localhost:9702
  mcp-airtable           localhost:9703
  mcp-obsidian           localhost:9704
  mcp-atlassian          localhost:9705

Search/AI Servers:
  mcp-brave-search       localhost:9801
  mcp-perplexity         localhost:9802
  mcp-omnisearch         localhost:9803
  mcp-context7           localhost:9804
  mcp-llamaindex         localhost:9805

Cloud Provider Servers:
  mcp-workers            localhost:9901
EOF
}

# Service groups
CORE_SERVICES="mcp-fetch mcp-git mcp-time mcp-filesystem mcp-memory mcp-everything mcp-sequentialthinking"
DB_SERVICES="mcp-redis mcp-redis-backend mcp-mongodb mcp-mongodb-backend mcp-supabase mcp-qdrant mcp-qdrant-backend"
DEVOPS_SERVICES="mcp-kubernetes mcp-github mcp-cloudflare mcp-heroku mcp-sentry"
BROWSER_SERVICES="mcp-playwright mcp-browserbase mcp-firecrawl"
COMM_SERVICES="mcp-slack mcp-telegram"
PROD_SERVICES="mcp-notion mcp-trello mcp-airtable mcp-obsidian mcp-atlassian"
SEARCH_SERVICES="mcp-brave-search mcp-perplexity mcp-omnisearch mcp-context7 mcp-llamaindex"
CLOUD_SERVICES="mcp-workers"

# All MCP server ports to check
ALL_PORTS="9101 9102 9103 9104 9105 9106 9107 9201 9202 9203 9301 9401 9402 9403 9404 9405 9501 9502 9503 9601 9602 9701 9702 9703 9704 9705 9801 9802 9803 9804 9805 9901"

case "${1:-}" in
    --start)
        echo "Starting all MCP servers (32 servers)..."
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
    --start-devops)
        echo "Starting DevOps MCP servers..."
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $DEVOPS_SERVICES
        echo "DevOps MCP servers started."
        ;;
    --start-comm)
        echo "Starting communication MCP servers..."
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $COMM_SERVICES
        echo "Communication MCP servers started."
        ;;
    --start-prod)
        echo "Starting productivity MCP servers..."
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $PROD_SERVICES
        echo "Productivity MCP servers started."
        ;;
    --start-search)
        echo "Starting search/AI MCP servers..."
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$COMPOSE_FILE" up -d $SEARCH_SERVICES
        echo "Search/AI MCP servers started."
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
        for port in $ALL_PORTS; do
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
