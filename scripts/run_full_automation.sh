#!/bin/bash
#
# run_full_automation.sh - Full Automation Test Suite Runner
#
# This script runs the complete automation test suite for HelixAgent.
# It orchestrates all test phases and provides detailed reporting.
#
# Phases:
#   1. Setup - Initialize test environment and mock servers
#   2. Unit Tests - Verify all unit tests pass
#   3. Integration Tests - Run integration tests
#   4. API Tests - Test all API endpoints
#   5. Tool Execution Tests - Test tool execution
#   6. E2E Tests - End-to-end flow tests
#   7. Performance Tests - Load and performance tests
#   8. Cleanup - Stop servers and clean up
#
# Usage:
#   ./scripts/run_full_automation.sh [options]
#
# Options:
#   --verbose          Show detailed output
#   --coverage         Generate coverage report
#   --with-external    Also run external server tests
#   --phase PHASE      Run specific phase only (1-8, or name)
#   --parallel         Run independent tests in parallel
#   --timeout SECONDS  Set test timeout (default: 600)
#   --help             Show this help message
#

set -e

# =============================================================================
# CONFIGURATION
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_DIR="${PROJECT_ROOT}/test-reports"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="${LOG_DIR}/automation_${TIMESTAMP}.log"
REPORT_FILE="${LOG_DIR}/automation_report_${TIMESTAMP}.txt"
COVERAGE_DIR="${LOG_DIR}/coverage"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Default options
VERBOSE=false
COVERAGE=false
WITH_EXTERNAL=false
SPECIFIC_PHASE=""
PARALLEL=false
TEST_TIMEOUT=600
START_INFRA=true
STOP_INFRA=true

# Test results
declare -A PHASE_RESULTS
declare -A PHASE_DURATIONS
TOTAL_TESTS=0
TOTAL_PASSED=0
TOTAL_FAILED=0
TOTAL_SKIPPED=0

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --coverage|-c)
            COVERAGE=true
            shift
            ;;
        --with-external)
            WITH_EXTERNAL=true
            shift
            ;;
        --phase)
            SPECIFIC_PHASE="$2"
            shift 2
            ;;
        --parallel)
            PARALLEL=true
            shift
            ;;
        --timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        --no-infra-start)
            START_INFRA=false
            shift
            ;;
        --no-infra-stop)
            STOP_INFRA=false
            shift
            ;;
        --help|-h)
            echo "HelixAgent Full Automation Test Suite"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --verbose, -v      Show detailed output"
            echo "  --coverage, -c     Generate coverage report"
            echo "  --with-external    Run tests against external server"
            echo "  --phase PHASE      Run specific phase (1-8 or name)"
            echo "  --parallel         Run independent tests in parallel"
            echo "  --timeout SECONDS  Set test timeout (default: 600)"
            echo "  --no-infra-start   Don't start infrastructure"
            echo "  --no-infra-stop    Don't stop infrastructure after tests"
            echo "  --help, -h         Show this help"
            echo ""
            echo "Phases:"
            echo "  1, setup       - Initialize environment"
            echo "  2, unit        - Unit test verification"
            echo "  3, integration - Integration tests"
            echo "  4, api         - API endpoint tests"
            echo "  5, tools       - Tool execution tests"
            echo "  6, e2e         - End-to-end tests"
            echo "  7, performance - Performance tests"
            echo "  8, cleanup     - Cleanup"
            echo ""
            echo "Examples:"
            echo "  $0 --verbose --coverage"
            echo "  $0 --phase api"
            echo "  $0 --with-external --timeout 900"
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

# =============================================================================
# LOGGING FUNCTIONS
# =============================================================================

# Create log directory
mkdir -p "$LOG_DIR"
mkdir -p "$COVERAGE_DIR"

log() {
    local message="[$(date '+%Y-%m-%d %H:%M:%S')] $1"
    echo "$message" >> "$LOG_FILE"
    if [ "$VERBOSE" = true ]; then
        echo "$message"
    fi
}

log_header() {
    echo ""
    echo -e "${CYAN}============================================================${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}============================================================${NC}"
    echo ""
    log "=== $1 ==="
}

log_phase() {
    echo ""
    echo -e "${MAGENTA}------------------------------------------------------------${NC}"
    echo -e "${MAGENTA}  Phase $1: $2${NC}"
    echo -e "${MAGENTA}------------------------------------------------------------${NC}"
    log "--- Phase $1: $2 ---"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
    log "[INFO] $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    log "[PASS] $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    log "[WARN] $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    log "[FAIL] $1"
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    log "[SKIP] $1"
}

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

# Check if a command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# Check if a port is in use
port_in_use() {
    nc -z localhost "$1" 2>/dev/null
}

# Wait for a service to be ready
wait_for_service() {
    local host="$1"
    local port="$2"
    local timeout="$3"
    local count=0

    while [ $count -lt "$timeout" ]; do
        if nc -z "$host" "$port" 2>/dev/null; then
            return 0
        fi
        count=$((count + 1))
        sleep 1
    done
    return 1
}

# Detect container runtime
detect_container_runtime() {
    if command_exists docker && docker info &> /dev/null 2>&1; then
        echo "docker"
    elif command_exists podman && podman info &> /dev/null 2>&1; then
        echo "podman"
    else
        echo "none"
    fi
}

# Get compose command
get_compose_cmd() {
    local runtime="$1"
    if [ "$runtime" = "docker" ]; then
        if docker compose version &> /dev/null 2>&1; then
            echo "docker compose"
        elif command_exists docker-compose; then
            echo "docker-compose"
        fi
    elif [ "$runtime" = "podman" ]; then
        if command_exists podman-compose; then
            echo "podman-compose"
        fi
    fi
}

# =============================================================================
# INFRASTRUCTURE MANAGEMENT
# =============================================================================

start_test_infrastructure() {
    log_info "Starting test infrastructure..."

    CONTAINER_RUNTIME=$(detect_container_runtime)

    if [ "$CONTAINER_RUNTIME" = "none" ]; then
        log_warning "No container runtime found. Some tests may be skipped."
        return 0
    fi

    COMPOSE_CMD=$(get_compose_cmd "$CONTAINER_RUNTIME")

    if [ -z "$COMPOSE_CMD" ]; then
        log_warning "No compose command found. Starting containers directly..."

        # Start PostgreSQL
        if ! port_in_use "${POSTGRES_PORT:-15432}"; then
            $CONTAINER_RUNTIME run -d --name helixagent-test-postgres \
                -p "${POSTGRES_PORT:-15432}:5432" \
                -e POSTGRES_DB=helixagent_db \
                -e POSTGRES_USER=helixagent \
                -e POSTGRES_PASSWORD=helixagent123 \
                docker.io/library/postgres:15-alpine 2>/dev/null || true
        fi

        # Start Redis
        if ! port_in_use "${REDIS_PORT:-16379}"; then
            $CONTAINER_RUNTIME run -d --name helixagent-test-redis \
                -p "${REDIS_PORT:-16379}:6379" \
                docker.io/library/redis:7-alpine redis-server --requirepass helixagent123 2>/dev/null || true
        fi
    else
        # Use docker-compose
        if [ -f "docker-compose.test.yml" ]; then
            $COMPOSE_CMD -f docker-compose.test.yml up -d postgres redis 2>/dev/null || true
        fi
    fi

    # Wait for services
    log_info "Waiting for services to be ready..."
    wait_for_service localhost "${POSTGRES_PORT:-15432}" 30 || log_warning "PostgreSQL may not be ready"
    wait_for_service localhost "${REDIS_PORT:-16379}" 30 || log_warning "Redis may not be ready"

    log_success "Test infrastructure started"
}

stop_test_infrastructure() {
    if [ "$STOP_INFRA" = false ]; then
        log_info "Keeping infrastructure running (--no-infra-stop)"
        return 0
    fi

    log_info "Stopping test infrastructure..."

    CONTAINER_RUNTIME=$(detect_container_runtime)

    if [ "$CONTAINER_RUNTIME" = "none" ]; then
        return 0
    fi

    COMPOSE_CMD=$(get_compose_cmd "$CONTAINER_RUNTIME")

    if [ -n "$COMPOSE_CMD" ] && [ -f "docker-compose.test.yml" ]; then
        $COMPOSE_CMD -f docker-compose.test.yml down -v 2>/dev/null || true
    else
        $CONTAINER_RUNTIME rm -f helixagent-test-postgres helixagent-test-redis 2>/dev/null || true
    fi

    log_success "Test infrastructure stopped"
}

# =============================================================================
# ENVIRONMENT SETUP
# =============================================================================

setup_test_environment() {
    log_info "Setting up test environment..."

    # Database Configuration
    export DB_HOST=localhost
    export DB_PORT=${POSTGRES_PORT:-15432}
    export DB_USER=helixagent
    export DB_PASSWORD=helixagent123
    export DB_NAME=helixagent_db
    export DATABASE_URL="postgres://helixagent:helixagent123@localhost:${DB_PORT}/helixagent_db?sslmode=disable"

    # Redis Configuration
    export REDIS_HOST=localhost
    export REDIS_PORT=${REDIS_PORT:-16379}
    export REDIS_PASSWORD=helixagent123

    # Test Configuration
    export GIN_MODE=test
    export CI=true
    export FULL_TEST_MODE=true

    # Enable external tests if requested
    if [ "$WITH_EXTERNAL" = true ]; then
        export RUN_EXTERNAL_TESTS=true
    fi

    log_success "Test environment configured"
}

# =============================================================================
# TEST EXECUTION FUNCTIONS
# =============================================================================

run_automation_tests() {
    local test_pattern="$1"
    local phase_name="$2"
    local start_time=$(date +%s)
    local exit_code=0

    log_info "Running tests: $test_pattern"

    local test_cmd="go test -v"

    if [ "$COVERAGE" = true ]; then
        test_cmd="$test_cmd -coverprofile=${COVERAGE_DIR}/coverage_${phase_name}.out"
    fi

    test_cmd="$test_cmd -timeout ${TEST_TIMEOUT}s $test_pattern"

    if [ "$VERBOSE" = true ]; then
        eval "$test_cmd" 2>&1 | tee -a "$LOG_FILE" || exit_code=$?
    else
        eval "$test_cmd" >> "$LOG_FILE" 2>&1 || exit_code=$?
    fi

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    PHASE_DURATIONS["$phase_name"]=$duration

    if [ $exit_code -eq 0 ]; then
        PHASE_RESULTS["$phase_name"]="PASS"
        log_success "Phase $phase_name completed in ${duration}s"
    else
        PHASE_RESULTS["$phase_name"]="FAIL"
        log_error "Phase $phase_name failed after ${duration}s"
    fi

    return $exit_code
}

# =============================================================================
# PHASE RUNNERS
# =============================================================================

run_phase_1_setup() {
    log_phase "1" "Setup"

    if [ "$START_INFRA" = true ]; then
        start_test_infrastructure
    fi

    setup_test_environment

    # Run setup tests
    run_automation_tests "./tests/automation/... -run TestPhase1" "setup"
}

run_phase_2_unit() {
    log_phase "2" "Unit Test Verification"
    run_automation_tests "./tests/automation/... -run TestPhase2" "unit"
}

run_phase_3_integration() {
    log_phase "3" "Integration Tests"
    run_automation_tests "./tests/automation/... -run TestPhase3" "integration"
}

run_phase_4_api() {
    log_phase "4" "API Endpoint Tests"
    run_automation_tests "./tests/automation/... -run TestPhase4" "api"
}

run_phase_5_tools() {
    log_phase "5" "Tool Execution Tests"
    run_automation_tests "./tests/automation/... -run TestPhase5" "tools"
}

run_phase_6_e2e() {
    log_phase "6" "End-to-End Flow Tests"
    run_automation_tests "./tests/automation/... -run TestPhase6" "e2e"
}

run_phase_7_performance() {
    log_phase "7" "Performance Tests"
    run_automation_tests "./tests/automation/... -run TestPhase7" "performance"
}

run_phase_8_cleanup() {
    log_phase "8" "Cleanup"
    run_automation_tests "./tests/automation/... -run TestPhase8" "cleanup"
    stop_test_infrastructure
}

run_all_phases() {
    local all_passed=true

    run_phase_1_setup || all_passed=false
    run_phase_2_unit || all_passed=false
    run_phase_3_integration || all_passed=false
    run_phase_4_api || all_passed=false
    run_phase_5_tools || all_passed=false
    run_phase_6_e2e || all_passed=false
    run_phase_7_performance || all_passed=false
    run_phase_8_cleanup || all_passed=false

    if [ "$all_passed" = true ]; then
        return 0
    else
        return 1
    fi
}

run_specific_phase() {
    local phase="$1"

    case "$phase" in
        1|setup)
            run_phase_1_setup
            ;;
        2|unit)
            run_phase_2_unit
            ;;
        3|integration)
            run_phase_3_integration
            ;;
        4|api)
            run_phase_4_api
            ;;
        5|tools)
            run_phase_5_tools
            ;;
        6|e2e)
            run_phase_6_e2e
            ;;
        7|performance)
            run_phase_7_performance
            ;;
        8|cleanup)
            run_phase_8_cleanup
            ;;
        *)
            log_error "Unknown phase: $phase"
            echo "Valid phases: 1-8, setup, unit, integration, api, tools, e2e, performance, cleanup"
            exit 1
            ;;
    esac
}

# =============================================================================
# COVERAGE REPORT GENERATION
# =============================================================================

generate_coverage_report() {
    if [ "$COVERAGE" != true ]; then
        return 0
    fi

    log_info "Generating coverage report..."

    # Find all coverage files
    local coverage_files=""
    for f in "${COVERAGE_DIR}"/coverage_*.out; do
        if [ -f "$f" ]; then
            coverage_files="$coverage_files $f"
        fi
    done

    if [ -z "$coverage_files" ]; then
        log_warning "No coverage files found"
        return 0
    fi

    # Merge coverage files
    echo "mode: atomic" > "${COVERAGE_DIR}/coverage_combined.out"
    for f in $coverage_files; do
        tail -n +2 "$f" >> "${COVERAGE_DIR}/coverage_combined.out" 2>/dev/null || true
    done

    # Generate HTML report
    go tool cover -html="${COVERAGE_DIR}/coverage_combined.out" -o "${COVERAGE_DIR}/coverage.html" 2>/dev/null || true

    # Show coverage summary
    log_info "Coverage summary:"
    go tool cover -func="${COVERAGE_DIR}/coverage_combined.out" 2>/dev/null | tail -1 || true

    log_success "Coverage report generated: ${COVERAGE_DIR}/coverage.html"
}

# =============================================================================
# RESULTS SUMMARY
# =============================================================================

print_results_summary() {
    local total_duration=$(($(date +%s) - START_TIME))

    log_header "Automation Test Results"

    echo ""
    echo "Phase                | Status  | Duration"
    echo "---------------------|---------|----------"

    local passed=0
    local failed=0

    for phase in setup unit integration api tools e2e performance cleanup; do
        local status="${PHASE_RESULTS[$phase]:-N/A}"
        local duration="${PHASE_DURATIONS[$phase]:-0}"

        if [ "$status" = "PASS" ]; then
            echo -e "${phase}              | ${GREEN}PASS${NC}    | ${duration}s"
            passed=$((passed + 1))
        elif [ "$status" = "FAIL" ]; then
            echo -e "${phase}              | ${RED}FAIL${NC}    | ${duration}s"
            failed=$((failed + 1))
        else
            echo -e "${phase}              | ${YELLOW}N/A${NC}     | -"
        fi
    done

    echo ""
    echo "----------------------------------------"
    echo -e "Total: ${GREEN}$passed passed${NC}, ${RED}$failed failed${NC}"
    echo "Total Duration: ${total_duration}s"
    echo ""
    echo "Log file: $LOG_FILE"

    if [ "$COVERAGE" = true ]; then
        echo "Coverage report: ${COVERAGE_DIR}/coverage.html"
    fi

    # Generate report file
    {
        echo "HelixAgent Automation Test Report"
        echo "================================="
        echo ""
        echo "Date: $(date)"
        echo "Duration: ${total_duration}s"
        echo ""
        echo "Results:"
        echo "  Passed: $passed"
        echo "  Failed: $failed"
        echo ""
        echo "Phase Details:"
        for phase in setup unit integration api tools e2e performance cleanup; do
            echo "  $phase: ${PHASE_RESULTS[$phase]:-N/A} (${PHASE_DURATIONS[$phase]:-0}s)"
        done
    } > "$REPORT_FILE"

    echo ""
    echo "Report saved to: $REPORT_FILE"

    if [ $failed -gt 0 ]; then
        return 1
    fi
    return 0
}

# =============================================================================
# CLEANUP HANDLER
# =============================================================================

cleanup() {
    local exit_status=$?

    if [ $exit_status -ne 0 ]; then
        log_warning "Test suite failed with exit code $exit_status"
    fi

    # Always stop infrastructure on exit unless specifically requested not to
    if [ "$STOP_INFRA" = true ]; then
        stop_test_infrastructure
    fi
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    START_TIME=$(date +%s)

    log_header "HelixAgent Full Automation Test Suite"

    log_info "Project root: $PROJECT_ROOT"
    log_info "Log file: $LOG_FILE"
    log_info "Verbose: $VERBOSE"
    log_info "Coverage: $COVERAGE"
    log_info "Timeout: ${TEST_TIMEOUT}s"

    # Register cleanup handler
    trap cleanup EXIT

    # Run tests
    local exit_code=0

    if [ -n "$SPECIFIC_PHASE" ]; then
        log_info "Running specific phase: $SPECIFIC_PHASE"
        run_specific_phase "$SPECIFIC_PHASE" || exit_code=1
    else
        log_info "Running all phases"
        run_all_phases || exit_code=1
    fi

    # Generate coverage report
    generate_coverage_report

    # Print summary
    print_results_summary || exit_code=1

    if [ $exit_code -eq 0 ]; then
        log_header "ALL TESTS PASSED"
    else
        log_header "SOME TESTS FAILED"
    fi

    exit $exit_code
}

# Run main
main
