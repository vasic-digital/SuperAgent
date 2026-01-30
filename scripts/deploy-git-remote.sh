#!/bin/bash
# deploy-remote.sh - Deploy HelixAgent services to a remote host via SSH
# Usage: ./deploy-remote.sh <remote-host> [service-name]
# If service-name omitted, deploys all services defined in docker-compose.yml

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
    echo "Usage: $0 [options] <remote-host> [service-name]"
    echo ""
    echo "Deploy HelixAgent services to a remote host via SSH."
    echo ""
    echo "Arguments:"
    echo "  remote-host   SSH host (user@hostname)"
    echo "  service-name  Optional Docker Compose service name (e.g., postgres, redis)"
    echo "                If omitted, deploys all services."
    echo ""
    echo "Options:"
    echo "  --dry-run      Print commands without executing"
    echo "  --help         Show this help message"
    echo "  --services     Specify services (comma-separated) instead of service-name"
    echo "  --compose-file Docker compose file (default: docker-compose.yml)"
    echo "  --remote-dir   Directory on remote host (default: /opt/helixagent)"
    echo "  --ssh-key      Path to SSH private key (default: ~/.ssh/id_rsa)"
    echo "  --git          Use Git clone/update instead of SCP (requires git on remote)"
    echo "  --git-repo     Git repository URL (default: current origin)"
    echo ""
    echo "Environment variables (override with options):"
    echo "  SSH_KEY       Path to SSH private key (default: ~/.ssh/id_rsa)"
    echo "  COMPOSE_FILE  Docker compose file (default: docker-compose.yml)"
    echo "  REMOTE_DIR    Directory on remote host (default: /opt/helixagent)"
    echo "  GIT_MODE      Enable git mode (1) or SCP mode (0)"
    echo "  GIT_REPO      Git repository URL"
    exit 1
}

# Defaults
DRY_RUN=0
SSH_KEY="${SSH_KEY:-$HOME/.ssh/id_rsa}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
REMOTE_DIR="${REMOTE_DIR:-/opt/helixagent}"
SERVICE_NAME=""
GIT_MODE="${GIT_MODE:-0}"
GIT_REPO="${GIT_REPO:-}"

# Parse options
POSITIONAL=()
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=1
            shift
            ;;
        --help)
            usage
            ;;
        --services)
            SERVICE_NAME="$2"
            shift 2
            ;;
        --compose-file)
            COMPOSE_FILE="$2"
            shift 2
            ;;
        --remote-dir)
            REMOTE_DIR="$2"
            shift 2
            ;;
        --ssh-key)
            SSH_KEY="$2"
            shift 2
            ;;
        --git)
            GIT_MODE=1
            shift
            ;;
        --git-repo)
            GIT_REPO="$2"
            shift 2
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
# If SERVICE_NAME not set via --services, use positional argument
if [ -z "$SERVICE_NAME" ]; then
    SERVICE_NAME="${2:-}"
fi

# Validate SSH key exists (skip for dry-run)
if [ $DRY_RUN -eq 0 ]; then
    if [ ! -f "$SSH_KEY" ]; then
        log_error "SSH key not found: $SSH_KEY"
        exit 1
    fi
else
    log_info "[DRY RUN] Skipping SSH key validation"
fi

# Validate compose file exists (skip for dry-run)
if [ $DRY_RUN -eq 0 ]; then
    if [ ! -f "$COMPOSE_FILE" ]; then
        log_error "Docker compose file not found: $COMPOSE_FILE"
        exit 1
    fi
else
    if [ ! -f "$COMPOSE_FILE" ]; then
        log_warning "Docker compose file not found: $COMPOSE_FILE (dry run)"
    fi
fi

log_info "=============================================="
log_info "HelixAgent Remote Deployment"
log_info "=============================================="
log_info "Remote host: $REMOTE_HOST"
log_info "Service: ${SERVICE_NAME:-all services}"
log_info "Remote directory: $REMOTE_DIR"
log_info ""

# Function to run command on remote host
# Function to run command on remote host (respects DRY_RUN)
ssh_cmd() {
    local cmd="ssh -i \"$SSH_KEY\" -o ConnectTimeout=10 -o StrictHostKeyChecking=no \"$REMOTE_HOST\" \"$@\""
    if [ $DRY_RUN -eq 1 ]; then
        log_info "[DRY RUN] Would execute: $cmd"
        return 0
    else
        eval "$cmd"
    fi
}

# Function to copy files via SCP (respects DRY_RUN)
scp_cmd() {
    local src="$1"
    local dst="$2"
    local cmd="scp -i \"$SSH_KEY\" -o StrictHostKeyChecking=no \"$src\" \"$dst\""
    if [ $DRY_RUN -eq 1 ]; then
        log_info "[DRY RUN] Would execute: $cmd"
        return 0
    else
        eval "$cmd"
    fi
}

# Function to copy directory recursively via SCP (respects DRY_RUN)
scp_cmd_r() {
    local src="$1"
    local dst="$2"
    local cmd="scp -i \"$SSH_KEY\" -o StrictHostKeyChecking=no -r \"$src\" \"$dst\""
    if [ $DRY_RUN -eq 1 ]; then
        log_info "[DRY RUN] Would execute: $cmd"
        return 0
    else
        eval "$cmd"
    fi
}

# Check SSH connectivity
log_info "Testing SSH connectivity..."
if [ $DRY_RUN -eq 1 ]; then
    log_success "SSH connectivity confirmed (dry run)"
else
    if ! ssh_cmd "echo 'SSH connection successful'" >/dev/null 2>&1; then
        log_error "SSH connection failed to $REMOTE_HOST"
        exit 1
    fi
    log_success "SSH connectivity confirmed"
fi

# Check Docker availability on remote host
log_info "Checking Docker on remote host..."
if [ $DRY_RUN -eq 1 ]; then
    log_success "Docker is available (dry run)"
else
    if ! ssh_cmd "command -v docker" >/dev/null 2>&1; then
        log_error "Docker not found on remote host"
        exit 1
    fi
    log_success "Docker is available"
fi

# Check Docker Compose availability
log_info "Checking Docker Compose..."
if [ $DRY_RUN -eq 1 ]; then
    COMPOSE_CMD="docker compose"
    log_success "Docker Compose is available (dry run)"
else
    if ! ssh_cmd "command -v docker-compose" >/dev/null 2>&1; then
        # Try Docker Compose Plugin
        if ! ssh_cmd "docker compose version" >/dev/null 2>&1; then
            log_error "Docker Compose not found on remote host"
            exit 1
        fi
        COMPOSE_CMD="docker compose"
    else
        COMPOSE_CMD="docker-compose"
    fi
    log_success "Docker Compose is available"
fi

# Check Git availability (if Git mode)
if [ "$GIT_MODE" = "1" ]; then
    log_info "Checking Git on remote host..."
    if [ $DRY_RUN -eq 1 ]; then
        log_success "Git is available (dry run)"
    else
        if ! ssh_cmd "command -v git" >/dev/null 2>&1; then
            log_error "Git not found on remote host"
            exit 1
        fi
        log_success "Git is available"
    fi
fi

# Create remote directory
log_info "Creating remote directory..."
ssh_cmd "mkdir -p $REMOTE_DIR"

# Copy docker-compose.yml and required files
log_info "Copying compose files..."
scp_cmd "$COMPOSE_FILE" "$REMOTE_HOST:$REMOTE_DIR/docker-compose.yml"

# Copy additional compose files if they exist
for file in docker-compose.production.yml docker-compose.messaging.yml docker-compose.monitoring.yml; do
    if [ -f "$file" ]; then
        log_info "Copying $file..."
        scp_cmd "$file" "$REMOTE_HOST:$REMOTE_DIR/"
    fi
done

# Copy configuration directories
if [ -d "configs" ]; then
    log_info "Copying configs directory..."
    scp_cmd_r "configs" "$REMOTE_HOST:$REMOTE_DIR/"
fi

# Copy scripts directory (contains init-db.sql)
if [ -d "scripts" ]; then
    log_info "Copying scripts directory..."
    scp_cmd_r "scripts" "$REMOTE_HOST:$REMOTE_DIR/"
fi

# Copy .env if exists
if [ -f ".env" ]; then
    log_info "Copying .env file..."
    scp_cmd ".env" "$REMOTE_HOST:$REMOTE_DIR/"
else
    log_warning ".env file not found, remote deployment may require environment variables"
fi

# Determine compose command
if [ -n "$SERVICE_NAME" ]; then
    log_info "Starting service: $SERVICE_NAME"
    ssh_cmd "cd $REMOTE_DIR && $COMPOSE_CMD up -d $SERVICE_NAME"
else
    log_info "Starting all services..."
    ssh_cmd "cd $REMOTE_DIR && $COMPOSE_CMD up -d"
fi

# Wait for services to become healthy
log_info "Waiting for services to start (30 seconds)..."
if [ $DRY_RUN -eq 0 ]; then
    sleep 30
else
    log_info "[DRY RUN] Skipping sleep"
fi

# Check service status
log_info "Checking service status..."
if [ -n "$SERVICE_NAME" ]; then
    ssh_cmd "cd $REMOTE_DIR && $COMPOSE_CMD ps $SERVICE_NAME"
else
    ssh_cmd "cd $REMOTE_DIR && $COMPOSE_CMD ps"
fi

log_success "Deployment completed successfully!"
log_info ""
log_info "Next steps:"
log_info "1. Check logs: ssh -i $SSH_KEY $REMOTE_HOST 'cd $REMOTE_DIR && $COMPOSE_CMD logs -f'"
log_info "2. Stop services: ssh -i $SSH_KEY $REMOTE_HOST 'cd $REMOTE_DIR && $COMPOSE_CMD down'"
log_info "3. Update deployment: re-run this script"
log_info ""
log_info "Remember to configure remote services in configs/remote-services-example.yaml"