#!/bin/bash
#
# run_complete_test_suite.sh - Complete Test Suite Runner
#
# This script runs ALL 6 test types with full infrastructure support.
# It automatically starts Docker/Podman containers, runs tests, and cleans up.
#
# Test Types:
#   1. Unit Tests
#   2. Integration Tests
#   3. E2E Tests
#   4. Security Tests
#   5. Stress Tests
#   6. Chaos/Challenge Tests
#
# Usage:
#   ./scripts/run_complete_test_suite.sh [options]
#
# Options:
#   --type TYPE      Run specific test type (unit|integration|e2e|security|stress|chaos|all)
#   --keep           Keep containers running after tests
#   --verbose        Show detailed output
#   --coverage       Generate coverage report
#   --no-cleanup     Skip cleanup on failure (for debugging)
#   --help           Show this help message
#

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="docker-compose.test.yml"
COMPOSE_FILE_INT="docker-compose.integration.yml"
PROJECT_NAME="superagent-test-suite"
MAX_WAIT_TIME=120
LOG_FILE="${PROJECT_ROOT}/test_suite_results.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default options
TEST_TYPE="all"
KEEP_CONTAINERS=false
VERBOSE=false
COVERAGE=false
NO_CLEANUP=false

# Test counters
TOTAL_PASSED=0
TOTAL_FAILED=0
TOTAL_SKIPPED=0

# Track which test types have been run
declare -A TEST_RESULTS

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --type)
            TEST_TYPE="$2"
            shift 2
            ;;
        --keep)
            KEEP_CONTAINERS=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --coverage)
            COVERAGE=true
            shift
            ;;
        --no-cleanup)
            NO_CLEANUP=true
            shift
            ;;
        --help)
            echo "SuperAgent Complete Test Suite Runner"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --type TYPE      Run specific test type:"
            echo "                   unit, integration, e2e, security, stress, chaos, all (default)"
            echo "  --keep           Keep containers running after tests"
            echo "  --verbose        Show detailed output"
            echo "  --coverage       Generate coverage report"
            echo "  --no-cleanup     Skip cleanup on failure (for debugging)"
            echo "  --help           Show this help message"
            echo ""
            echo "Test Types:"
            echo "  1. unit        - Unit tests (./internal/... -short)"
            echo "  2. integration - Integration tests (./tests/integration)"
            echo "  3. e2e         - End-to-end tests (./tests/e2e)"
            echo "  4. security    - Security tests (./tests/security)"
            echo "  5. stress      - Stress tests (./tests/stress)"
            echo "  6. chaos       - Chaos/Challenge tests (./tests/challenge)"
            echo ""
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Change to project directory
cd "$PROJECT_ROOT"

# ============================================================================
# Logging Functions
# ============================================================================

log_header() {
    echo ""
    echo -e "${CYAN}=========================================${NC}"
    echo -e "${CYAN} $1${NC}"
    echo -e "${CYAN}=========================================${NC}"
    echo ""
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_test_type() {
    echo ""
    echo -e "${YELLOW}-------------------------------------------${NC}"
    echo -e "${YELLOW} Running: $1${NC}"
    echo -e "${YELLOW}-------------------------------------------${NC}"
}

# ============================================================================
# Container Runtime Detection
# ============================================================================

detect_container_runtime() {
    if command -v docker &> /dev/null && docker info &> /dev/null 2>&1; then
        echo "docker"
    elif command -v podman &> /dev/null && podman info &> /dev/null 2>&1; then
        echo "podman"
    else
        echo "none"
    fi
}

detect_compose_cmd() {
    local runtime="$1"

    if [ "$runtime" = "docker" ]; then
        if docker compose version &> /dev/null 2>&1; then
            echo "docker compose"
        elif command -v docker-compose &> /dev/null; then
            echo "docker-compose"
        else
            echo "none"
        fi
    elif [ "$runtime" = "podman" ]; then
        if command -v podman-compose &> /dev/null; then
            echo "podman-compose"
        else
            echo "none"
        fi
    else
        echo "none"
    fi
}

# ============================================================================
# Infrastructure Management
# ============================================================================

start_infrastructure() {
    log_info "Starting test infrastructure..."

    # Clean up any existing containers first
    $COMPOSE_CMD -f "$COMPOSE_FILE" down -v --remove-orphans 2>/dev/null || true

    # Build mock LLM server if needed
    log_info "Building mock LLM server..."
    $COMPOSE_CMD -f "$COMPOSE_FILE" build mock-llm 2>/dev/null || {
        log_warning "Mock LLM build had issues, continuing anyway..."
    }

    # Start core services
    log_info "Starting PostgreSQL, Redis, and Mock LLM..."
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d postgres redis mock-llm

    # Wait for services
    wait_for_postgres
    wait_for_redis
    wait_for_mock_llm

    log_success "Test infrastructure is ready!"
}

stop_infrastructure() {
    if [ "$KEEP_CONTAINERS" = true ]; then
        log_warning "Keeping containers running (--keep flag)"
        echo "To stop manually: $COMPOSE_CMD -f $COMPOSE_FILE down -v"
        return 0
    fi

    log_info "Stopping test infrastructure..."
    $COMPOSE_CMD -f "$COMPOSE_FILE" down -v --remove-orphans 2>/dev/null || true
    log_success "Infrastructure stopped and cleaned up"
}

wait_for_postgres() {
    log_info "Waiting for PostgreSQL to be ready..."
    local count=0
    while [ $count -lt $MAX_WAIT_TIME ]; do
        if $COMPOSE_CMD -f "$COMPOSE_FILE" exec -T postgres pg_isready -U superagent -d superagent_db > /dev/null 2>&1; then
            log_success "PostgreSQL is ready!"
            return 0
        fi
        count=$((count + 2))
        sleep 2
        [ "$VERBOSE" = true ] && echo -n "."
    done
    log_error "PostgreSQL failed to start within $MAX_WAIT_TIME seconds"
    return 1
}

wait_for_redis() {
    log_info "Waiting for Redis to be ready..."
    local count=0
    while [ $count -lt $MAX_WAIT_TIME ]; do
        if $COMPOSE_CMD -f "$COMPOSE_FILE" exec -T redis redis-cli -a superagent123 ping 2>/dev/null | grep -q "PONG"; then
            log_success "Redis is ready!"
            return 0
        fi
        count=$((count + 2))
        sleep 2
        [ "$VERBOSE" = true ] && echo -n "."
    done
    log_error "Redis failed to start within $MAX_WAIT_TIME seconds"
    return 1
}

wait_for_mock_llm() {
    log_info "Waiting for Mock LLM server to be ready..."
    local count=0
    while [ $count -lt $MAX_WAIT_TIME ]; do
        if curl -sf "http://localhost:${MOCK_LLM_PORT:-18081}/health" > /dev/null 2>&1; then
            log_success "Mock LLM server is ready!"
            return 0
        fi
        count=$((count + 2))
        sleep 2
        [ "$VERBOSE" = true ] && echo -n "."
    done
    log_warning "Mock LLM server may not be ready, continuing anyway..."
    return 0
}

# ============================================================================
# Environment Setup
# ============================================================================

setup_test_environment() {
    log_info "Setting up test environment variables..."

    # Database Configuration
    export DB_HOST=localhost
    export DB_PORT=${POSTGRES_PORT:-15432}
    export DB_USER=superagent
    export DB_PASSWORD=superagent123
    export DB_NAME=superagent_db
    export DB_SSLMODE=disable
    export DATABASE_URL="postgres://superagent:superagent123@localhost:${POSTGRES_PORT:-15432}/superagent_db?sslmode=disable"

    # Redis Configuration
    export REDIS_HOST=localhost
    export REDIS_PORT=${REDIS_PORT:-16379}
    export REDIS_PASSWORD=superagent123
    export REDIS_URL="redis://:superagent123@localhost:${REDIS_PORT:-16379}"

    # Mock LLM Configuration
    export MOCK_LLM_URL="http://localhost:${MOCK_LLM_PORT:-18081}"
    export MOCK_LLM_ENABLED=true

    # LLM Provider Configurations (all pointing to mock server)
    export CLAUDE_API_KEY=mock-api-key
    export CLAUDE_BASE_URL="http://localhost:${MOCK_LLM_PORT:-18081}/v1"
    export DEEPSEEK_API_KEY=mock-api-key
    export DEEPSEEK_BASE_URL="http://localhost:${MOCK_LLM_PORT:-18081}/v1"
    export GEMINI_API_KEY=mock-api-key
    export GEMINI_BASE_URL="http://localhost:${MOCK_LLM_PORT:-18081}/v1"
    export QWEN_API_KEY=mock-api-key
    export QWEN_BASE_URL="http://localhost:${MOCK_LLM_PORT:-18081}/v1"
    export ZAI_API_KEY=mock-api-key
    export ZAI_BASE_URL="http://localhost:${MOCK_LLM_PORT:-18081}/v1"
    export OLLAMA_BASE_URL="http://localhost:${MOCK_LLM_PORT:-18081}"
    export OLLAMA_ENABLED=true
    export OLLAMA_MODEL=mock-llama

    # JWT and Server Configuration
    export JWT_SECRET=test-jwt-secret-key-for-testing-purposes
    export SUPERAGENT_API_KEY=test-api-key-for-development

    # Test Configuration
    export GIN_MODE=test
    export CI=true
    export FULL_TEST_MODE=true
    export INTEGRATION_TEST=true

    # Cloud Provider Mock Configuration (for cloud tests)
    export AWS_ACCESS_KEY_ID=mock-aws-key
    export AWS_SECRET_ACCESS_KEY=mock-aws-secret
    export AWS_REGION=us-east-1
    export GCP_PROJECT_ID=mock-project
    export GOOGLE_ACCESS_TOKEN=mock-token
    export AZURE_OPENAI_ENDPOINT=http://localhost:${MOCK_LLM_PORT:-18081}
    export AZURE_OPENAI_API_KEY=mock-azure-key
    export AZURE_OPENAI_API_VERSION=2024-02-15-preview

    log_success "Environment variables configured"
}

# ============================================================================
# Test Runners
# ============================================================================

run_unit_tests() {
    log_test_type "Unit Tests"
    local exit_code=0

    local test_cmd="go test -v ./internal/... -short -timeout 300s"
    if [ "$COVERAGE" = true ]; then
        test_cmd="go test -v ./internal/... -short -timeout 300s -coverprofile=coverage_unit.out"
    fi

    if [ "$VERBOSE" = true ]; then
        eval $test_cmd 2>&1 | tee -a "$LOG_FILE" || exit_code=$?
    else
        eval $test_cmd >> "$LOG_FILE" 2>&1 || exit_code=$?
    fi

    TEST_RESULTS["unit"]=$exit_code
    return $exit_code
}

run_integration_tests() {
    log_test_type "Integration Tests"
    local exit_code=0

    local test_cmd="go test -v ./tests/integration/... -timeout 600s"
    if [ "$COVERAGE" = true ]; then
        test_cmd="go test -v ./tests/integration/... -timeout 600s -coverprofile=coverage_integration.out"
    fi

    if [ "$VERBOSE" = true ]; then
        eval $test_cmd 2>&1 | tee -a "$LOG_FILE" || exit_code=$?
    else
        eval $test_cmd >> "$LOG_FILE" 2>&1 || exit_code=$?
    fi

    TEST_RESULTS["integration"]=$exit_code
    return $exit_code
}

run_e2e_tests() {
    log_test_type "E2E Tests"
    local exit_code=0

    local test_cmd="go test -v ./tests/e2e/... -timeout 600s"
    if [ "$COVERAGE" = true ]; then
        test_cmd="go test -v ./tests/e2e/... -timeout 600s -coverprofile=coverage_e2e.out"
    fi

    if [ "$VERBOSE" = true ]; then
        eval $test_cmd 2>&1 | tee -a "$LOG_FILE" || exit_code=$?
    else
        eval $test_cmd >> "$LOG_FILE" 2>&1 || exit_code=$?
    fi

    TEST_RESULTS["e2e"]=$exit_code
    return $exit_code
}

run_security_tests() {
    log_test_type "Security Tests"
    local exit_code=0

    local test_cmd="go test -v ./tests/security/... -timeout 600s"
    if [ "$COVERAGE" = true ]; then
        test_cmd="go test -v ./tests/security/... -timeout 600s -coverprofile=coverage_security.out"
    fi

    if [ "$VERBOSE" = true ]; then
        eval $test_cmd 2>&1 | tee -a "$LOG_FILE" || exit_code=$?
    else
        eval $test_cmd >> "$LOG_FILE" 2>&1 || exit_code=$?
    fi

    TEST_RESULTS["security"]=$exit_code
    return $exit_code
}

run_stress_tests() {
    log_test_type "Stress Tests"
    local exit_code=0

    local test_cmd="go test -v ./tests/stress/... -timeout 900s"
    if [ "$COVERAGE" = true ]; then
        test_cmd="go test -v ./tests/stress/... -timeout 900s -coverprofile=coverage_stress.out"
    fi

    if [ "$VERBOSE" = true ]; then
        eval $test_cmd 2>&1 | tee -a "$LOG_FILE" || exit_code=$?
    else
        eval $test_cmd >> "$LOG_FILE" 2>&1 || exit_code=$?
    fi

    TEST_RESULTS["stress"]=$exit_code
    return $exit_code
}

run_chaos_tests() {
    log_test_type "Chaos/Challenge Tests"
    local exit_code=0

    local test_cmd="go test -v ./tests/challenge/... -timeout 600s"
    if [ "$COVERAGE" = true ]; then
        test_cmd="go test -v ./tests/challenge/... -timeout 600s -coverprofile=coverage_chaos.out"
    fi

    if [ "$VERBOSE" = true ]; then
        eval $test_cmd 2>&1 | tee -a "$LOG_FILE" || exit_code=$?
    else
        eval $test_cmd >> "$LOG_FILE" 2>&1 || exit_code=$?
    fi

    TEST_RESULTS["chaos"]=$exit_code
    return $exit_code
}

run_all_tests() {
    local all_passed=true

    # Run unit tests first (fastest, no infra needed)
    run_unit_tests || all_passed=false

    # Run integration tests
    run_integration_tests || all_passed=false

    # Run E2E tests
    run_e2e_tests || all_passed=false

    # Run security tests
    run_security_tests || all_passed=false

    # Run stress tests
    run_stress_tests || all_passed=false

    # Run chaos tests
    run_chaos_tests || all_passed=false

    if [ "$all_passed" = true ]; then
        return 0
    else
        return 1
    fi
}

# ============================================================================
# Coverage Report Generation
# ============================================================================

generate_combined_coverage() {
    if [ "$COVERAGE" != true ]; then
        return 0
    fi

    log_info "Generating combined coverage report..."

    # Combine all coverage files if they exist
    local coverage_files=""
    for cov_file in coverage_unit.out coverage_integration.out coverage_e2e.out coverage_security.out coverage_stress.out coverage_chaos.out; do
        if [ -f "$cov_file" ]; then
            coverage_files="$coverage_files $cov_file"
        fi
    done

    if [ -n "$coverage_files" ]; then
        # Merge coverage files
        echo "mode: atomic" > coverage_combined.out
        for f in $coverage_files; do
            tail -n +2 "$f" >> coverage_combined.out 2>/dev/null || true
        done

        # Generate HTML report
        go tool cover -html=coverage_combined.out -o coverage_combined.html 2>/dev/null || true

        # Show coverage summary
        log_info "Coverage Summary:"
        go tool cover -func=coverage_combined.out | tail -1

        log_success "Coverage report generated: coverage_combined.html"
    fi
}

# ============================================================================
# Results Summary
# ============================================================================

print_results_summary() {
    log_header "Test Results Summary"

    echo "Test Type        | Status"
    echo "-----------------|--------"

    for test_type in unit integration e2e security stress chaos; do
        local result="${TEST_RESULTS[$test_type]:-N/A}"
        if [ "$result" = "0" ]; then
            echo -e "$test_type         | ${GREEN}PASSED${NC}"
        elif [ "$result" = "N/A" ]; then
            echo -e "$test_type         | ${YELLOW}SKIPPED${NC}"
        else
            echo -e "$test_type         | ${RED}FAILED${NC}"
        fi
    done

    echo ""

    # Count results
    local passed=0
    local failed=0
    local skipped=0

    for test_type in unit integration e2e security stress chaos; do
        local result="${TEST_RESULTS[$test_type]:-N/A}"
        if [ "$result" = "0" ]; then
            passed=$((passed + 1))
        elif [ "$result" = "N/A" ]; then
            skipped=$((skipped + 1))
        else
            failed=$((failed + 1))
        fi
    done

    echo -e "Total: ${GREEN}$passed passed${NC}, ${RED}$failed failed${NC}, ${YELLOW}$skipped skipped${NC}"
    echo ""
    echo "Detailed logs: $LOG_FILE"

    if [ $failed -gt 0 ]; then
        return 1
    fi
    return 0
}

# ============================================================================
# Cleanup Handler
# ============================================================================

cleanup() {
    local exit_status=$?

    if [ "$NO_CLEANUP" = true ] && [ $exit_status -ne 0 ]; then
        log_warning "Skipping cleanup due to --no-cleanup flag"
        log_info "Containers are still running for debugging"
        return
    fi

    stop_infrastructure
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    # Clear log file
    > "$LOG_FILE"

    log_header "SuperAgent Complete Test Suite"

    # Detect container runtime
    CONTAINER_RUNTIME=$(detect_container_runtime)

    if [ "$CONTAINER_RUNTIME" = "none" ]; then
        log_error "No container runtime found!"
        echo ""
        echo "Please install Docker or Podman:"
        echo "  Docker: https://docs.docker.com/get-docker/"
        echo "  Podman: https://podman.io/getting-started/installation"
        exit 1
    fi

    COMPOSE_CMD=$(detect_compose_cmd "$CONTAINER_RUNTIME")

    if [ "$COMPOSE_CMD" = "none" ]; then
        log_error "No compose tool found for $CONTAINER_RUNTIME!"
        if [ "$CONTAINER_RUNTIME" = "podman" ]; then
            echo "Install: pip install podman-compose"
        else
            echo "Install docker-compose or upgrade Docker to use 'docker compose'"
        fi
        exit 1
    fi

    log_info "Using container runtime: $CONTAINER_RUNTIME"
    log_info "Using compose command: $COMPOSE_CMD"
    log_info "Test type: $TEST_TYPE"
    echo ""

    # Register cleanup trap
    trap cleanup EXIT

    # Start infrastructure
    start_infrastructure

    # Setup environment
    setup_test_environment

    # Run tests based on type
    local test_exit_code=0

    case "$TEST_TYPE" in
        unit)
            run_unit_tests || test_exit_code=1
            ;;
        integration)
            run_integration_tests || test_exit_code=1
            ;;
        e2e)
            run_e2e_tests || test_exit_code=1
            ;;
        security)
            run_security_tests || test_exit_code=1
            ;;
        stress)
            run_stress_tests || test_exit_code=1
            ;;
        chaos)
            run_chaos_tests || test_exit_code=1
            ;;
        all)
            run_all_tests || test_exit_code=1
            ;;
        *)
            log_error "Unknown test type: $TEST_TYPE"
            echo "Valid types: unit, integration, e2e, security, stress, chaos, all"
            exit 1
            ;;
    esac

    # Generate coverage report if requested
    generate_combined_coverage

    # Print results summary
    print_results_summary || test_exit_code=1

    if [ $test_exit_code -eq 0 ]; then
        log_success "All tests completed successfully!"
    else
        log_error "Some tests failed!"
    fi

    exit $test_exit_code
}

# Run main function
main
