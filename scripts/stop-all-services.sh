#!/bin/bash
# stop-all-services.sh - Stop ALL HelixAgent Services
# Gracefully shuts down all HelixAgent containers and services

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Detect container runtime
if command -v podman &> /dev/null; then
    RUNTIME="podman"
    COMPOSE_CMD="podman-compose"
elif command -v docker &> /dev/null; then
    RUNTIME="docker"
    COMPOSE_CMD="docker compose"
else
    echo -e "${RED}Error: Neither docker nor podman found!${NC}"
    exit 1
fi

echo "=============================================="
echo "HelixAgent - Complete Services Shutdown"
echo "=============================================="
echo ""
echo "Container Runtime: $RUNTIME"
echo ""

# Function to stop a compose file
stop_compose() {
    local compose_file="$1"
    local description="$2"

    if [ ! -f "$compose_file" ]; then
        return
    fi

    echo -e "${BLUE}▼ Stopping $description...${NC}"
    $COMPOSE_CMD -f "$compose_file" down 2>/dev/null || true
    echo -e "${GREEN}✓ $description stopped${NC}"
}

# Stop in reverse order of startup
echo "Stopping all services..."
echo ""

stop_compose "docker-compose.security.yml" "Security Services"
stop_compose "docker-compose.analytics.yml" "Analytics Services"
stop_compose "docker-compose.bigdata.yml" "Big Data Services"
stop_compose "docker-compose.integration.yml" "Integration Services"
stop_compose "docker-compose.protocols.yml" "Protocol Services"
stop_compose "docker-compose.monitoring.yml" "Monitoring Services"
stop_compose "docker-compose.messaging.yml" "Messaging Services"
stop_compose "docker-compose.yml" "Core Services"

# Check for any remaining helixagent containers
echo ""
echo "Checking for remaining containers..."
REMAINING=$($RUNTIME ps -a --format "{{.Names}}" | grep -c helixagent || true)

if [ "$REMAINING" -gt 0 ]; then
    echo -e "${YELLOW}Found $REMAINING helixagent containers still running${NC}"
    echo ""
    $RUNTIME ps -a --format "table {{.Names}}\t{{.Status}}" | grep helixagent || true
    echo ""
    echo -e "${BLUE}Stopping remaining containers...${NC}"
    $RUNTIME ps -a --format "{{.Names}}" | grep helixagent | xargs -r $RUNTIME stop 2>/dev/null || true
    $RUNTIME ps -a --format "{{.Names}}" | grep helixagent | xargs -r $RUNTIME rm 2>/dev/null || true
    echo -e "${GREEN}✓ Remaining containers stopped${NC}"
else
    echo -e "${GREEN}✓ All helixagent containers stopped${NC}"
fi

echo ""
echo "=========================================="
echo "SHUTDOWN COMPLETE"
echo "=========================================="
echo ""
echo "All HelixAgent services have been stopped."
echo ""
echo "To start services again:"
echo "  ./scripts/start-all-services.sh"
echo ""
