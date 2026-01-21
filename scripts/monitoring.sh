#!/bin/bash
# Unified Monitoring Stack Management Script for HelixAgent
#
# Usage:
#   ./scripts/monitoring.sh start    # Start all monitoring services
#   ./scripts/monitoring.sh stop     # Stop all monitoring services
#   ./scripts/monitoring.sh status   # Check status of monitoring services
#   ./scripts/monitoring.sh logs     # View aggregated logs
#   ./scripts/monitoring.sh urls     # Show access URLs

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="${PROJECT_DIR}/docker-compose.monitoring.yml"

# Detect container runtime
detect_runtime() {
    if command -v docker &> /dev/null && docker info &> /dev/null 2>&1; then
        echo "docker"
    elif command -v podman &> /dev/null; then
        echo "podman"
    else
        echo "none"
    fi
}

# Detect compose command
detect_compose() {
    local runtime="$1"
    if [ "$runtime" = "docker" ]; then
        if docker compose version &> /dev/null 2>&1; then
            echo "docker compose"
        elif command -v docker-compose &> /dev/null; then
            echo "docker-compose"
        fi
    elif [ "$runtime" = "podman" ]; then
        if command -v podman-compose &> /dev/null; then
            echo "podman-compose"
        fi
    fi
}

RUNTIME=$(detect_runtime)
COMPOSE_CMD=$(detect_compose "$RUNTIME")

if [ -z "$COMPOSE_CMD" ]; then
    echo -e "${RED}Error: No container runtime found. Install Docker or Podman.${NC}"
    exit 1
fi

show_urls() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}HelixAgent Monitoring Stack URLs${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "${GREEN}Grafana:${NC}      http://localhost:3000"
    echo -e "              Username: admin"
    echo -e "              Password: admin123"
    echo ""
    echo -e "${GREEN}Prometheus:${NC}   http://localhost:9090"
    echo ""
    echo -e "${GREEN}Loki:${NC}         http://localhost:3100"
    echo ""
    echo -e "${GREEN}Alertmanager:${NC} http://localhost:9093"
    echo ""
    echo -e "${BLUE}========================================${NC}"
}

start_monitoring() {
    echo -e "${BLUE}Starting HelixAgent Monitoring Stack...${NC}"
    echo -e "${YELLOW}Using: ${COMPOSE_CMD}${NC}"

    cd "$PROJECT_DIR"
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d

    echo ""
    echo -e "${GREEN}Waiting for services to be ready...${NC}"
    sleep 5

    # Check Grafana
    if curl -s "http://localhost:3000/api/health" | grep -q "ok"; then
        echo -e "${GREEN}✓ Grafana is ready${NC}"
    else
        echo -e "${YELLOW}⏳ Grafana is starting...${NC}"
    fi

    # Check Prometheus
    if curl -s "http://localhost:9090/-/ready" | grep -q "Prometheus"; then
        echo -e "${GREEN}✓ Prometheus is ready${NC}"
    else
        echo -e "${YELLOW}⏳ Prometheus is starting...${NC}"
    fi

    # Check Loki
    if curl -s "http://localhost:3100/ready" | grep -q "ready"; then
        echo -e "${GREEN}✓ Loki is ready${NC}"
    else
        echo -e "${YELLOW}⏳ Loki is starting...${NC}"
    fi

    echo ""
    show_urls
}

stop_monitoring() {
    echo -e "${BLUE}Stopping HelixAgent Monitoring Stack...${NC}"
    cd "$PROJECT_DIR"
    $COMPOSE_CMD -f "$COMPOSE_FILE" down
    echo -e "${GREEN}Monitoring stack stopped.${NC}"
}

show_status() {
    echo -e "${BLUE}HelixAgent Monitoring Stack Status${NC}"
    echo -e "${BLUE}========================================${NC}"
    cd "$PROJECT_DIR"
    $COMPOSE_CMD -f "$COMPOSE_FILE" ps
}

show_logs() {
    echo -e "${BLUE}Monitoring Stack Logs (Ctrl+C to exit)${NC}"
    echo -e "${BLUE}========================================${NC}"
    cd "$PROJECT_DIR"
    $COMPOSE_CMD -f "$COMPOSE_FILE" logs -f --tail=100
}

# Main
case "${1:-}" in
    start)
        start_monitoring
        ;;
    stop)
        stop_monitoring
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    urls)
        show_urls
        ;;
    restart)
        stop_monitoring
        sleep 2
        start_monitoring
        ;;
    *)
        echo "Usage: $0 {start|stop|status|logs|urls|restart}"
        echo ""
        echo "Commands:"
        echo "  start   - Start all monitoring services"
        echo "  stop    - Stop all monitoring services"
        echo "  status  - Show status of monitoring services"
        echo "  logs    - View aggregated logs"
        echo "  urls    - Show access URLs"
        echo "  restart - Restart all monitoring services"
        exit 1
        ;;
esac
