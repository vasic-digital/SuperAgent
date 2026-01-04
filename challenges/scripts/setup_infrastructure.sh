#!/bin/bash
# SuperAgent Challenges - Infrastructure Setup Script
# This script sets up the required infrastructure for running all challenges
#
# REQUIRES: sudo/root access for initial setup

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo "=========================================="
echo "  SuperAgent Infrastructure Setup"
echo "=========================================="

# Check for root/sudo
check_sudo() {
    if [ "$EUID" -ne 0 ]; then
        print_warning "This script requires sudo for some operations"
        print_info "Please run with: sudo $0"
        exit 1
    fi
}

# Setup rootless Podman
setup_podman() {
    print_info "Setting up rootless Podman..."

    local USER_NAME="${SUDO_USER:-$USER}"
    local USER_ID=$(id -u "$USER_NAME")

    # Add subuid/subgid entries
    if ! grep -q "^${USER_NAME}:" /etc/subuid 2>/dev/null; then
        echo "${USER_NAME}:100000:65536" >> /etc/subuid
        print_success "Added subuid entry for $USER_NAME"
    fi

    if ! grep -q "^${USER_NAME}:" /etc/subgid 2>/dev/null; then
        echo "${USER_NAME}:100000:65536" >> /etc/subgid
        print_success "Added subgid entry for $USER_NAME"
    fi

    # Run podman system migrate as the user
    print_info "Running podman system migrate..."
    sudo -u "$USER_NAME" podman system migrate 2>/dev/null || true

    print_success "Podman rootless setup complete"
}

# Start PostgreSQL container
start_postgres() {
    print_info "Starting PostgreSQL container..."

    local USER_NAME="${SUDO_USER:-$USER}"

    # Remove existing container if exists
    sudo -u "$USER_NAME" podman rm -f superagent-postgres 2>/dev/null || true

    # Create network if not exists
    sudo -u "$USER_NAME" podman network create superagent-net 2>/dev/null || true

    # Start PostgreSQL
    sudo -u "$USER_NAME" podman run -d \
        --name superagent-postgres \
        --network superagent-net \
        -e POSTGRES_USER=superagent \
        -e POSTGRES_PASSWORD=superagent123 \
        -e POSTGRES_DB=superagent_db \
        -p 5432:5432 \
        docker.io/postgres:15-alpine

    print_success "PostgreSQL container started"
}

# Start Redis container
start_redis() {
    print_info "Starting Redis container..."

    local USER_NAME="${SUDO_USER:-$USER}"

    # Remove existing container if exists
    sudo -u "$USER_NAME" podman rm -f superagent-redis 2>/dev/null || true

    # Start Redis
    sudo -u "$USER_NAME" podman run -d \
        --name superagent-redis \
        --network superagent-net \
        -p 6379:6379 \
        docker.io/redis:7-alpine

    print_success "Redis container started"
}

# Wait for services to be ready
wait_for_services() {
    print_info "Waiting for services to be ready..."

    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if pg_isready -h localhost -p 5432 -U superagent 2>/dev/null; then
            print_success "PostgreSQL is ready"
            break
        fi
        echo "  Waiting for PostgreSQL... (attempt $attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done

    attempt=1
    while [ $attempt -le $max_attempts ]; do
        if redis-cli -h localhost -p 6379 ping 2>/dev/null | grep -q PONG; then
            print_success "Redis is ready"
            break
        fi
        echo "  Waiting for Redis... (attempt $attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
}

# Start SuperAgent
start_superagent() {
    print_info "Starting SuperAgent..."

    cd "$PROJECT_ROOT"

    export JWT_SECRET="superagent-jwt-secret-for-testing-32chars"
    export GIN_MODE=release
    export DB_HOST=localhost
    export DB_PORT=5432
    export DB_USER=superagent
    export DB_PASSWORD=superagent123
    export DB_NAME=superagent_db
    export REDIS_HOST=localhost
    export REDIS_PORT=6379

    # Build if needed
    if [ ! -f "./bin/superagent" ]; then
        make build
    fi

    # Start in background
    nohup ./bin/superagent --auto-start-docker=false > /tmp/superagent.log 2>&1 &

    sleep 3

    if curl -s http://localhost:8080/health | grep -q healthy; then
        print_success "SuperAgent is running"
    else
        print_error "SuperAgent failed to start"
        cat /tmp/superagent.log
        exit 1
    fi
}

# Main
main() {
    check_sudo
    setup_podman
    start_postgres
    start_redis
    wait_for_services

    print_info ""
    print_success "Infrastructure is ready!"
    print_info ""
    print_info "To start SuperAgent, run as your user:"
    print_info "  export JWT_SECRET='superagent-jwt-secret-for-testing-32chars'"
    print_info "  export DB_HOST=localhost DB_PORT=5432 DB_USER=superagent DB_PASSWORD=superagent123 DB_NAME=superagent_db"
    print_info "  export REDIS_HOST=localhost REDIS_PORT=6379"
    print_info "  ./bin/superagent --auto-start-docker=false"
    print_info ""
    print_info "Then run all challenges:"
    print_info "  cd challenges && ./scripts/run_all_challenges.sh"
}

main "$@"
