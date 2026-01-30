#!/bin/bash
# deploy-all-remote.sh - Deploy all HelixAgent services to multiple remote hosts
# Usage: ./deploy-all-remote.sh <host1> [host2 ...]
# Alternatively, set HOSTS environment variable (comma-separated)

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
    echo "Usage: $0 <host1> [host2 ...]"
    echo "   or: HOSTS=user@host1,user@host2 $0"
    echo ""
    echo "Deploys all HelixAgent services to each remote host in parallel."
    echo "Each host gets a complete deployment of all services."
    echo ""
    echo "Environment variables:"
    echo "  HOSTS          Comma-separated list of SSH hosts"
    echo "  SSH_KEY        Path to SSH private key (default: ~/.ssh/id_rsa)"
    echo "  COMPOSE_FILE   Docker compose file (default: docker-compose.yml)"
    echo "  REMOTE_DIR     Directory on remote host (default: /opt/helixagent)"
    exit 1
}

# Determine hosts
HOSTS_LIST=()
if [ $# -gt 0 ]; then
    HOSTS_LIST=("$@")
elif [ -n "$HOSTS" ]; then
    IFS=',' read -r -a HOSTS_LIST <<< "$HOSTS"
else
    usage
fi

SSH_KEY="${SSH_KEY:-$HOME/.ssh/id_rsa}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
REMOTE_DIR="${REMOTE_DIR:-/opt/helixagent}"

log_info "=============================================="
log_info "HelixAgent Multi-Host Remote Deployment"
log_info "=============================================="
log_info "Hosts: ${HOSTS_LIST[*]}"
log_info "Total hosts: ${#HOSTS_LIST[@]}"
log_info ""

FAILED_HOSTS=()
SUCCESS_HOSTS=()

for host in "${HOSTS_LIST[@]}"; do
    log_info "Deploying to host: $host"
    # Run deploy-remote.sh in background for parallel deployment
    (
        if ./scripts/deploy-remote.sh "$host"; then
            log_success "Deployment to $host completed successfully"
            echo "$host" >> /tmp/helixagent_deploy_success.$$
        else
            log_error "Deployment to $host failed"
            echo "$host" >> /tmp/helixagent_deploy_failed.$$
        fi
    ) &
done

# Wait for all background jobs
wait

# Collect results
if [ -f /tmp/helixagent_deploy_success.$$ ]; then
    SUCCESS_HOSTS=($(cat /tmp/helixagent_deploy_success.$$))
    rm -f /tmp/helixagent_deploy_success.$$
fi
if [ -f /tmp/helixagent_deploy_failed.$$ ]; then
    FAILED_HOSTS=($(cat /tmp/helixagent_deploy_failed.$$))
    rm -f /tmp/helixagent_deploy_failed.$$
fi

log_info "=============================================="
log_info "Deployment Summary"
log_info "=============================================="
log_info "Total hosts attempted: ${#HOSTS_LIST[@]}"
log_success "Successful: ${#SUCCESS_HOSTS[@]}"
if [ ${#FAILED_HOSTS[@]} -gt 0 ]; then
    log_error "Failed: ${#FAILED_HOSTS[@]}"
    for h in "${FAILED_HOSTS[@]}"; do
        log_error "  - $h"
    done
    exit 1
else
    log_success "All deployments completed successfully!"
    exit 0
fi