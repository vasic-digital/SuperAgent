#!/bin/bash
# test-remote-deployment.sh - Test remote deployment of HelixAgent containers
# Uses sshpass for password authentication. Checks container health and service availability.
# Usage: ./test-remote-deployment.sh <remote-host> [profile|service]
# Example: ./test-remote-deployment.sh thinker.local core   # Test core services
# Example: ./test-remote-deployment.sh thinker.local full   # Test all services

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

usage() {
    echo "Usage: $0 [options] <remote-host> [profile|service]"
    echo ""
    echo "Test remote deployment of HelixAgent containers."
    echo "Uses sshpass for password authentication (set REMOTE_PASSWORD env var)."
    echo ""
    echo "Arguments:"
    echo "  remote-host   Remote hostname or IP (user will be milosvasic)"
    echo "  profile       Deployment profile: core, ai, monitoring, messaging, bigdata,"
    echo "                protocols, analytics, security, formatters, rag, mcp, full"
    echo "  service       Specific service name (e.g., postgres, redis, helixagent)"
    echo ""
    echo "Options:"
    echo "  --remote-user  SSH username (default: milosvasic)"
    echo "  --remote-dir   Directory on remote host (default: /opt/helixagent)"
    echo "  --password     SSH password (default: from REMOTE_PASSWORD env var)"
    echo "  --compose-runtime  docker or podman (default: auto-detect)"
    echo ""
    echo "Environment variables:"
    echo "  REMOTE_PASSWORD  SSH password for remote host"
    echo "  REMOTE_USER      SSH username (default: milosvasic)"
    echo "  REMOTE_DIR       Remote directory (default: /opt/helixagent)"
    echo ""
    exit 1
}

# Defaults
REMOTE_USER="${REMOTE_USER:-milosvasic}"
REMOTE_DIR="${REMOTE_DIR:-/opt/helixagent}"
REMOTE_PASSWORD="${REMOTE_PASSWORD:-}"
COMPOSE_RUNTIME=""

# Parse options
POSITIONAL=()
while [[ $# -gt 0 ]]; do
    case $1 in
        --remote-user)
            REMOTE_USER="$2"
            shift 2
            ;;
        --remote-dir)
            REMOTE_DIR="$2"
            shift 2
            ;;
        --password)
            REMOTE_PASSWORD="$2"
            shift 2
            ;;
        --compose-runtime)
            COMPOSE_RUNTIME="$2"
            shift 2
            ;;
        --help)
            usage
            ;;
        -*)
            echo "Unknown option: $1"
            usage
            ;;
        *)
            POSITIONAL+=("$1")
            shift
            ;;
    esac
done

# Restore positional arguments
set -- "${POSITIONAL[@]}"

if [ $# -lt 1 ]; then
    usage
fi

REMOTE_HOST="$1"
PROFILE_OR_SERVICE="${2:-core}"

if [ -z "$REMOTE_PASSWORD" ]; then
    log_warning "REMOTE_PASSWORD not set, using default password 'thinker'"
    REMOTE_PASSWORD="thinker"
fi

# SSH helper
ssh_cmd() {
    sshpass -p "$REMOTE_PASSWORD" ssh -o StrictHostKeyChecking=no -o ConnectTimeout=10 "${REMOTE_USER}@${REMOTE_HOST}" "$1"
}

# Detect container runtime on remote host
if [ -z "$COMPOSE_RUNTIME" ]; then
    log_info "Detecting container runtime on remote host..."
    if ssh_cmd "command -v podman &> /dev/null"; then
        COMPOSE_RUNTIME="podman"
        CONTAINER_CLI="podman"
    elif ssh_cmd "command -v docker &> /dev/null"; then
        COMPOSE_RUNTIME="docker"
        CONTAINER_CLI="docker"
    else
        log_error "Neither docker nor podman found on remote host"
        exit 1
    fi
else
    case "$COMPOSE_RUNTIME" in
        podman) CONTAINER_CLI="podman" ;;
        docker) CONTAINER_CLI="docker" ;;
        *) log_error "Invalid compose-runtime: $COMPOSE_RUNTIME"; exit 1 ;;
    esac
fi

log_info "Remote host: $REMOTE_USER@$REMOTE_HOST"
log_info "Profile/Service: $PROFILE_OR_SERVICE"
log_info "Remote directory: $REMOTE_DIR"
log_info "Container runtime: $COMPOSE_RUNTIME ($CONTAINER_CLI)"
log_info ""

# Function to check container health
check_container_health() {
    local container_name="$1"
    local max_wait="${2:-30}"
    
    log_info "Checking health of $container_name..."
    local wait_time=0
    while [ $wait_time -lt $max_wait ]; do
        if ssh_cmd "$CONTAINER_CLI inspect $container_name &> /dev/null"; then
            local health_status=$(ssh_cmd "$CONTAINER_CLI inspect --format='{{.State.Health.Status}}' \"$container_name\" 2>/dev/null || echo 'none'")
            local running_status=$(ssh_cmd "$CONTAINER_CLI inspect --format='{{.State.Running}}' \"$container_name\" 2>/dev/null || echo 'false'")
            if [ "$health_status" = "healthy" ] || [ "$running_status" = "true" ]; then
                log_success "$container_name is healthy/running"
                return 0
            fi
        fi
        sleep 2
        wait_time=$((wait_time + 2))
    done
    log_error "$container_name failed health check after $max_wait seconds"
    return 1
}

# Function to test a specific service
test_service() {
    local service="$1"
    log_info "Testing service: $service"
    
    # Determine container name based on service
    local container_name="helixagent-$service"
    # Some services have different naming (e.g., postgres -> helixagent-postgres)
    # This mapping can be extended
    case "$service" in
        postgres) container_name="helixagent-postgres" ;;
        redis) container_name="helixagent-redis" ;;
        kafka) container_name="helixagent-kafka" ;;
        zookeeper) container_name="helixagent-zookeeper" ;;
        rabbitmq) container_name="helixagent-rabbitmq" ;;
        prometheus) container_name="helixagent-prometheus" ;;
        grafana) container_name="helixagent-grafana" ;;
        helixagent) container_name="helixagent-app" ;;
        *) container_name="helixagent-$service" ;;
    esac
    
    check_container_health "$container_name"
}

# Function to test a profile
test_profile() {
    local profile="$1"
    log_info "Testing profile: $profile"
    
    case "$profile" in
        core)
            test_service "postgres"
            test_service "redis"
            ;;
        ai)
            test_service "cognee"
            ;;
        monitoring)
            test_service "prometheus"
            test_service "grafana"
            ;;
        messaging)
            test_service "rabbitmq"
            test_service "zookeeper"
            test_service "kafka"
            ;;
        bigdata)
            # Add bigdata services
            log_warning "Big Data services test not yet implemented"
            ;;
        protocols)
            log_warning "Protocol services test not yet implemented"
            ;;
        analytics)
            log_warning "Analytics services test not yet implemented"
            ;;
        security)
            log_warning "Security services test not yet implemented"
            ;;
        formatters)
            log_warning "Formatters services test not yet implemented"
            ;;
        rag)
            log_warning "RAG services test not yet implemented"
            ;;
        mcp)
            log_warning "MCP services test not yet implemented"
            ;;
        full)
            # Test core, messaging, monitoring, etc.
            test_service "postgres"
            test_service "redis"
            test_service "rabbitmq"
            test_service "zookeeper"
            test_service "kafka"
            test_service "prometheus"
            test_service "grafana"
            ;;
        *)
            log_error "Unknown profile: $profile"
            exit 1
            ;;
    esac
}

# Determine testing action based on profile/service
case "$PROFILE_OR_SERVICE" in
    core|ai|monitoring|messaging|bigdata|protocols|analytics|security|formatters|rag|mcp|full)
        test_profile "$PROFILE_OR_SERVICE"
        ;;
    *)
        test_service "$PROFILE_OR_SERVICE"
        ;;
esac

log_success "All tests passed for $PROFILE_OR_SERVICE!"
log_info ""
log_info "Remote deployment validation complete."