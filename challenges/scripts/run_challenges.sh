#!/bin/bash
# SuperAgent Challenges - Challenge Runner Script
# Usage: ./scripts/run_challenges.sh [challenge_name] [options]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
VERBOSE=false
TIMEOUT=600
DRY_RUN=false

# Print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Show usage
usage() {
    cat << EOF
SuperAgent Challenges Runner

Usage: $0 <challenge_name> [options]

Available Challenges:
  provider_verification    Verify all LLM providers and score models
  ai_debate_formation      Form AI debate group from top models
  api_quality_test         Test API quality with assertions

Options:
  -v, --verbose            Enable verbose output
  -t, --timeout SECONDS    Set timeout (default: 600)
  -d, --dry-run            Show what would be done without executing
  -h, --help               Show this help message

Examples:
  $0 provider_verification
  $0 ai_debate_formation --verbose
  $0 api_quality_test --timeout=900

EOF
    exit 0
}

# Parse arguments
parse_args() {
    if [ $# -eq 0 ]; then
        usage
    fi

    CHALLENGE_NAME="$1"
    shift

    while [ $# -gt 0 ]; do
        case "$1" in
            -v|--verbose)
                VERBOSE=true
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift
                ;;
            --timeout=*)
                TIMEOUT="${1#*=}"
                ;;
            -d|--dry-run)
                DRY_RUN=true
                ;;
            -h|--help)
                usage
                ;;
            *)
                print_error "Unknown option: $1"
                usage
                ;;
        esac
        shift
    done
}

# Load environment variables
load_env() {
    local env_loaded=false

    # First try project root .env (primary location for API keys)
    if [ -f "$PROJECT_ROOT/.env" ]; then
        print_info "Loading environment from $PROJECT_ROOT/.env"
        set -a
        source "$PROJECT_ROOT/.env"
        set +a
        env_loaded=true
    fi

    # Then load challenges-specific .env (can override or add settings)
    if [ -f "$CHALLENGES_DIR/.env" ]; then
        print_info "Loading environment from $CHALLENGES_DIR/.env"
        set -a
        source "$CHALLENGES_DIR/.env"
        set +a
        env_loaded=true
    fi

    if [ "$env_loaded" = false ]; then
        print_warning ".env file not found. Using system environment variables."
    fi
}

# Create results directory
create_results_dir() {
    local challenge=$1
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local year=$(date +%Y)
    local month=$(date +%m)
    local day=$(date +%d)

    RESULTS_DIR="$CHALLENGES_DIR/results/$challenge/$year/$month/$day/$timestamp"

    mkdir -p "$RESULTS_DIR/logs"
    mkdir -p "$RESULTS_DIR/results"
    mkdir -p "$RESULTS_DIR/config"

    print_info "Results directory: $RESULTS_DIR"
}

# Find latest results directory for a challenge
find_latest_results() {
    local challenge=$1
    local base_path="$CHALLENGES_DIR/results/$challenge"

    if [ ! -d "$base_path" ]; then
        return 1
    fi

    # Find the most recent timestamp directory that contains a results subdirectory
    local latest=$(find "$base_path" -type d -name "results" 2>/dev/null | \
                   head -n 1 | \
                   xargs -I {} dirname {})

    if [ -n "$latest" ] && [ -d "$latest/results" ]; then
        echo "$latest"
        return 0
    fi

    return 1
}

# Check dependencies
check_dependencies() {
    local challenge=$1

    case "$challenge" in
        ai_debate_formation)
            # Check if provider_verification has run
            PROVIDER_VERIFICATION_RESULTS=$(find_latest_results "provider_verification")
            if [ -z "$PROVIDER_VERIFICATION_RESULTS" ]; then
                print_error "Dependency not met: provider_verification must run first"
                exit 1
            fi
            print_info "Found provider_verification results at: $PROVIDER_VERIFICATION_RESULTS"
            export DEPENDENCY_DIR="$PROVIDER_VERIFICATION_RESULTS"
            ;;
        api_quality_test)
            # Check if ai_debate_formation has run (optional)
            AI_DEBATE_RESULTS=$(find_latest_results "ai_debate_formation")
            if [ -n "$AI_DEBATE_RESULTS" ]; then
                print_info "Found ai_debate_formation results at: $AI_DEBATE_RESULTS"
                export DEPENDENCY_DIR="$AI_DEBATE_RESULTS"
            fi
            ;;
    esac
}

# Run the challenge
run_challenge() {
    local challenge=$1

    print_info "Running challenge: $challenge"
    print_info "Timeout: ${TIMEOUT}s"

    if [ "$DRY_RUN" = true ]; then
        print_warning "DRY RUN - Not executing"
        return 0
    fi

    # Create results directory
    create_results_dir "$challenge"

    # Log start
    local start_time=$(date +%s)
    echo "{\"event\":\"challenge_started\",\"challenge\":\"$challenge\",\"timestamp\":\"$(date -Iseconds)\"}" >> "$RESULTS_DIR/logs/challenge.log"

    # Build and run the Go challenge runner
    local go_file="$CHALLENGES_DIR/codebase/go_files/$challenge/main.go"

    if [ -f "$go_file" ]; then
        print_info "Building challenge runner..."

        cd "$CHALLENGES_DIR/codebase/go_files/$challenge"

        # Build args
        local args="--results-dir=\"$RESULTS_DIR\""
        if [ "$VERBOSE" = true ]; then
            args="$args --verbose"
        fi
        if [ -n "$DEPENDENCY_DIR" ]; then
            args="$args --dependency-dir=\"$DEPENDENCY_DIR\""
        fi

        if [ "$VERBOSE" = true ]; then
            eval "go run main.go $args" 2>&1 | tee "$RESULTS_DIR/logs/output.log"
        else
            eval "go run main.go $args" 2>&1 | tee "$RESULTS_DIR/logs/output.log"
        fi

        local exit_code=${PIPESTATUS[0]}
    else
        print_warning "Go implementation not found. Running shell fallback..."

        local shell_runner="$CHALLENGES_DIR/codebase/challenge_runners/$challenge/run.sh"

        if [ -f "$shell_runner" ]; then
            bash "$shell_runner" --results-dir="$RESULTS_DIR" ${VERBOSE:+--verbose}
            local exit_code=$?
        else
            print_error "No runner found for challenge: $challenge"
            exit 1
        fi
    fi

    # Log completion
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    echo "{\"event\":\"challenge_completed\",\"challenge\":\"$challenge\",\"duration_seconds\":$duration,\"exit_code\":$exit_code,\"timestamp\":\"$(date -Iseconds)\"}" >> "$RESULTS_DIR/logs/challenge.log"

    if [ $exit_code -eq 0 ]; then
        print_success "Challenge completed successfully in ${duration}s"
        print_info "Results: $RESULTS_DIR"
    else
        print_error "Challenge failed with exit code $exit_code"
        exit $exit_code
    fi
}

# Main
main() {
    parse_args "$@"

    print_info "SuperAgent Challenges Runner"
    print_info "Challenge: $CHALLENGE_NAME"

    # Validate challenge name
    case "$CHALLENGE_NAME" in
        provider_verification|ai_debate_formation|api_quality_test)
            ;;
        *)
            print_error "Unknown challenge: $CHALLENGE_NAME"
            usage
            ;;
    esac

    load_env
    check_dependencies "$CHALLENGE_NAME"
    run_challenge "$CHALLENGE_NAME"
}

main "$@"
