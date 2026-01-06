#!/bin/bash
# SuperAgent Challenges - Run All Challenges in Sequence
# Usage: ./scripts/run_all_challenges.sh [options]
#
# This script automatically:
# 1. Builds SuperAgent binary if needed
# 2. Starts all required infrastructure (SuperAgent server)
# 3. Runs all 40 challenges
# 4. Reports final results

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

# Load environment from project root first (primary location for API keys)
if [ -f "$PROJECT_ROOT/.env" ]; then
    set -a
    source "$PROJECT_ROOT/.env"
    set +a
fi

# Then load challenges-specific .env (can override or add settings)
if [ -f "$CHALLENGES_DIR/.env" ]; then
    set -a
    source "$CHALLENGES_DIR/.env"
    set +a
fi

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }
print_phase() { echo -e "${PURPLE}[PHASE]${NC} $1"; }

# Track started services for cleanup
SUPERAGENT_PID=""
STARTED_SERVICES=()

#===============================================================================
# CLEANUP HANDLER
#===============================================================================
cleanup() {
    print_info "Cleaning up..."

    # Stop SuperAgent if we started it
    if [ -n "$SUPERAGENT_PID" ] && kill -0 "$SUPERAGENT_PID" 2>/dev/null; then
        print_info "Stopping SuperAgent (PID: $SUPERAGENT_PID)..."
        kill "$SUPERAGENT_PID" 2>/dev/null || true
        wait "$SUPERAGENT_PID" 2>/dev/null || true
    fi

    # Remove PID file
    rm -f "$CHALLENGES_DIR/results/superagent_challenges.pid"
}

trap cleanup EXIT

#===============================================================================
# INFRASTRUCTURE FUNCTIONS
#===============================================================================

# Build SuperAgent binary if needed
build_superagent() {
    print_phase "Building SuperAgent Binary"

    local binary=""
    if [ -x "$PROJECT_ROOT/bin/superagent" ]; then
        binary="$PROJECT_ROOT/bin/superagent"
    elif [ -x "$PROJECT_ROOT/superagent" ]; then
        binary="$PROJECT_ROOT/superagent"
    fi

    if [ -n "$binary" ]; then
        print_success "SuperAgent binary found: $binary"
        return 0
    fi

    print_info "Building SuperAgent..."
    if (cd "$PROJECT_ROOT" && make build 2>&1); then
        if [ -x "$PROJECT_ROOT/bin/superagent" ]; then
            print_success "SuperAgent built successfully"
            return 0
        elif [ -x "$PROJECT_ROOT/superagent" ]; then
            print_success "SuperAgent built successfully"
            return 0
        fi
    fi

    print_error "Failed to build SuperAgent"
    return 1
}

# Check if SuperAgent is running
check_superagent() {
    local port="${SUPERAGENT_PORT:-8080}"
    local host="${SUPERAGENT_HOST:-localhost}"

    if curl -s "http://$host:$port/health" > /dev/null 2>&1; then
        return 0
    elif curl -s "http://$host:$port/v1/models" > /dev/null 2>&1; then
        return 0
    fi
    return 1
}

# Start SuperAgent server
start_superagent() {
    print_phase "Starting SuperAgent Server"

    local port="${SUPERAGENT_PORT:-8080}"
    local host="${SUPERAGENT_HOST:-localhost}"

    # Check if already running
    if check_superagent; then
        print_success "SuperAgent already running on $host:$port"
        return 0
    fi

    # Find binary
    local binary=""
    if [ -x "$PROJECT_ROOT/bin/superagent" ]; then
        binary="$PROJECT_ROOT/bin/superagent"
    elif [ -x "$PROJECT_ROOT/superagent" ]; then
        binary="$PROJECT_ROOT/superagent"
    fi

    if [ -z "$binary" ]; then
        print_error "SuperAgent binary not found"
        return 1
    fi

    print_info "Starting SuperAgent from: $binary"

    # Create results directory if needed
    mkdir -p "$CHALLENGES_DIR/results"

    # Start SuperAgent in background
    nohup "$binary" > "$CHALLENGES_DIR/results/superagent_challenges.log" 2>&1 &
    SUPERAGENT_PID=$!
    echo $SUPERAGENT_PID > "$CHALLENGES_DIR/results/superagent_challenges.pid"
    STARTED_SERVICES+=("superagent")

    # Wait for startup
    print_info "Waiting for SuperAgent to start..."
    local max_attempts=30
    local attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if check_superagent; then
            print_success "SuperAgent started successfully (PID: $SUPERAGENT_PID)"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done

    print_error "SuperAgent failed to start within ${max_attempts}s"
    print_error "Check log: $CHALLENGES_DIR/results/superagent_challenges.log"
    cat "$CHALLENGES_DIR/results/superagent_challenges.log" | tail -20
    return 1
}

# Start all required infrastructure
start_infrastructure() {
    print_info "=========================================="
    print_info "  Starting Infrastructure"
    print_info "=========================================="
    echo ""

    # Build SuperAgent if needed
    if ! build_superagent; then
        print_error "Failed to build SuperAgent - cannot continue"
        exit 1
    fi

    # Start SuperAgent
    if ! start_superagent; then
        print_error "Failed to start SuperAgent - cannot continue"
        exit 1
    fi

    echo ""
    print_success "All infrastructure started successfully"
    echo ""
}

# All 40 challenges in dependency order
CHALLENGES=(
    # Infrastructure (no dependencies)
    "health_monitoring"
    "configuration_loading"
    "caching_layer"
    "database_operations"
    "authentication"
    "plugin_system"

    # Security (depends on caching_layer, authentication)
    "rate_limiting"
    "input_validation"

    # Providers (no dependencies)
    "provider_claude"
    "provider_deepseek"
    "provider_gemini"
    "provider_ollama"
    "provider_openrouter"
    "provider_qwen"
    "provider_zai"

    # Core verification
    "provider_verification"

    # Protocols (no dependencies)
    "mcp_protocol"
    "lsp_protocol"
    "acp_protocol"

    # Cloud integrations
    "cloud_aws_bedrock"
    "cloud_gcp_vertex"
    "cloud_azure_openai"

    # Core features (depends on provider_verification)
    "ensemble_voting"
    "embeddings_service"
    "streaming_responses"
    "model_metadata"

    # Debate (depends on provider_verification)
    "ai_debate_formation"
    "ai_debate_workflow"

    # API (depends on provider_verification)
    "openai_compatibility"
    "grpc_api"
    "api_quality_test"

    # Optimization (depends on embeddings)
    "optimization_semantic_cache"
    "optimization_structured_output"

    # Integration
    "cognee_integration"

    # Resilience
    "circuit_breaker"
    "error_handling"
    "concurrent_access"
    "graceful_shutdown"

    # Session (depends on caching, auth)
    "session_management"

    # Validation (depends on main challenge)
    "opencode"
)

# Parse options
VERBOSE=""
STOP_ON_FAILURE=true
SKIP_INFRA=false

while [ $# -gt 0 ]; do
    case "$1" in
        -v|--verbose)
            VERBOSE="--verbose"
            ;;
        --continue-on-failure)
            STOP_ON_FAILURE=false
            ;;
        --skip-infra)
            SKIP_INFRA=true
            ;;
        -h|--help)
            echo "Usage: $0 [--verbose] [--continue-on-failure] [--skip-infra]"
            echo ""
            echo "Options:"
            echo "  --verbose              Enable verbose output"
            echo "  --continue-on-failure  Continue even if a challenge fails"
            echo "  --skip-infra           Skip infrastructure startup (assume already running)"
            exit 0
            ;;
    esac
    shift
done

# Main execution
print_info "=========================================="
print_info "  SuperAgent - Run All Challenges"
print_info "=========================================="
print_info "Start time: $(date)"
print_info "Total challenges: ${#CHALLENGES[@]}"
echo ""

# Start infrastructure unless skipped
if [ "$SKIP_INFRA" = false ]; then
    start_infrastructure
else
    print_warning "Skipping infrastructure startup (--skip-infra)"
    if ! check_superagent; then
        print_error "SuperAgent is not running! Cannot continue."
        print_error "Either start SuperAgent manually or remove --skip-infra"
        exit 1
    fi
    print_success "SuperAgent is running"
fi

echo ""

TOTAL_START=$(date +%s)
PASSED=0
FAILED=0

for challenge in "${CHALLENGES[@]}"; do
    print_info "----------------------------------------"
    print_info "Running: $challenge"
    print_info "----------------------------------------"

    if "$SCRIPT_DIR/run_challenges.sh" "$challenge" $VERBOSE; then
        PASSED=$((PASSED + 1))
        print_success "$challenge completed successfully"
    else
        FAILED=$((FAILED + 1))
        print_error "$challenge failed"

        if [ "$STOP_ON_FAILURE" = true ]; then
            print_error "Stopping due to failure. Use --continue-on-failure to continue."
            break
        fi
    fi
    echo ""
done

TOTAL_END=$(date +%s)
TOTAL_DURATION=$((TOTAL_END - TOTAL_START))

# Generate master summary
print_info "Generating master summary..."
"$SCRIPT_DIR/generate_report.sh" --master-only 2>/dev/null || true

# Final report
echo ""
print_info "=========================================="
print_info "  All Challenges Complete"
print_info "=========================================="
print_info "Total duration: ${TOTAL_DURATION}s"
print_info "Passed: $PASSED / ${#CHALLENGES[@]}"
print_info "Failed: $FAILED / ${#CHALLENGES[@]}"

if [ $FAILED -eq 0 ]; then
    print_success "All challenges passed!"
    exit 0
else
    print_error "$FAILED challenge(s) failed"
    exit 1
fi
