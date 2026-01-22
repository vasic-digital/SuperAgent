#!/bin/bash
#
# stop-full-test-infra.sh - Stops all test infrastructure
#
# Supports both Docker and Podman
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Stopping all test infrastructure...${NC}"

cd "$PROJECT_ROOT"

# Detect runtime
if command -v docker &> /dev/null && docker info &> /dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
    if ! docker compose version &> /dev/null 2>&1; then
        COMPOSE_CMD="docker-compose"
    fi
elif command -v podman &> /dev/null; then
    COMPOSE_CMD="podman-compose"
else
    echo -e "${RED}No container runtime found${NC}"
    exit 1
fi

# Stop all compose services
$COMPOSE_CMD -f docker-compose.test.yml -f docker-compose.messaging.yml -f docker-compose.bigdata.yml --profile messaging --profile bigdata down --volumes --remove-orphans 2>/dev/null || true

# Also try without profiles for cleanup
$COMPOSE_CMD -f docker-compose.test.yml down --volumes --remove-orphans 2>/dev/null || true

# Remove .env.test
rm -f "$PROJECT_ROOT/.env.test" 2>/dev/null || true

echo -e "${GREEN}All test infrastructure stopped!${NC}"
