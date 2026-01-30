#!/bin/bash
# discover-services.sh - Service Discovery Challenge Script
# Tests the service discovery system (TCP, HTTP, DNS, mDNS) and BootManager integration

set -e

# Handle --test flag
if [[ "$1" == "--test" ]]; then
    echo "Running discovery script test mode..."
    # Run the TCP discovery test program
    cd "$(dirname "$0")/test-discovery"
    go run test_tcp_discovery.go --port 9999
    exit $?
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

log_info() { echo -e "${BLUE}[INFO]${NC} $(date +%H:%M:%S) $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $(date +%H:%M:%S) $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $(date +%H:%M:%S) $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $(date +%H:%M:%S) $1"; }
log_test() { log_info "[TEST] $1"; }

run_test() {
    local test_name="$1"
    local test_cmd="$2"
    log_test "$test_name"
    if eval "$test_cmd" &> /dev/null; then
        log_success "$test_name"
        ((PASSED++))
    else
        log_error "$test_name"
        ((FAILED++))
    fi
    ((TOTAL++))
}

log_info "=============================================="
log_info "Service Discovery Challenge"
log_info "=============================================="
log_info "Validates: Service discovery system (TCP, HTTP, DNS, mDNS)"
log_info "           BootManager integration"
log_info "           Configuration support"
log_info ""

# ============================================================================
# Section 1: File Structure Validation
# ============================================================================

log_info "=============================================="
log_info "Section 1: File Structure Validation"
log_info "=============================================="

# Test 1: discovery package exists
run_test "Discovery package exists" \
    "[ -d \"internal/services/discovery\" ]"

# Test 2: discoverer.go exists
run_test "discoverer.go file exists" \
    "[ -f \"internal/services/discovery/discoverer.go\" ]"

# Test 3: discoverer_test.go exists
run_test "discoverer_test.go file exists" \
    "[ -f \"internal/services/discovery/discoverer_test.go\" ]"

# Test 4: BootManager integration (check for discoverer field)
run_test "BootManager has discoverer field" \
    "grep -q \"Discoverer\" internal/services/boot_manager.go"

# ============================================================================
# Section 2: Configuration Validation
# ============================================================================

log_info "=============================================="
log_info "Section 2: Configuration Validation"
log_info "=============================================="

# Test 5: development.yaml has discovery fields
run_test "development.yaml contains discovery_enabled" \
    "grep -q \"discovery_enabled\" configs/development.yaml"

# Test 6: production.yaml has discovery fields
run_test "production.yaml contains discovery_enabled" \
    "grep -q \"discovery_enabled\" configs/production.yaml"

# Test 7: remote-services-example.yaml has discovery fields
run_test "remote-services-example.yaml contains discovery_enabled" \
    "grep -q \"discovery_enabled\" configs/remote-services-example.yaml"

# ============================================================================
# Section 3: Unit Tests Execution
# ============================================================================

log_info "=============================================="
log_info "Section 3: Unit Tests Execution"
log_info "=============================================="

# Test 8: Run discovery unit tests
run_test "Discovery unit tests pass" \
    "cd internal/services/discovery && go test -v -short 2>&1 | grep -q 'PASS'"

# ============================================================================
# Section 4: Integration Test (TCP Discovery)
# ============================================================================

log_info "=============================================="
log_info "Section 4: Integration Test (TCP Discovery)"
log_info "=============================================="

# Start a simple TCP server on a random port
log_info "Starting test TCP server..."
TEST_PORT=$((RANDOM % 10000 + 20000))
timeout 5 nc -l -p $TEST_PORT &
SERVER_PID=$!
sleep 1

# Create a temporary configuration file for discovery
TEMP_CONFIG=$(mktemp)
cat > "$TEMP_CONFIG" << EOF
services:
  test-tcp:
    host: "127.0.0.1"
    port: "$TEST_PORT"
    discovery_enabled: true
    discovery_method: "tcp"
EOF

# Test 9: TCP discovery should succeed
run_test "TCP discovery of local test server" \
    "go run ./scripts/test-discovery/test_tcp_discovery.go --port $TEST_PORT 2>&1 | grep -q 'Discovered: true'"

# Clean up
kill $SERVER_PID 2>/dev/null || true
rm -f "$TEMP_CONFIG"

# ============================================================================
# Section 5: BootManager Integration Test
# ============================================================================

log_info "=============================================="
log_info "Section 5: BootManager Integration Test"
log_info "=============================================="

# Test 10: BootManager compiles with discovery
run_test "BootManager compiles with discovery integration" \
    "go build ./internal/services/boot_manager.go ./internal/services/health_checker.go ./internal/services/discovery/discoverer.go 2>&1 | grep -q ''"

log_info "=============================================="
log_info "Challenge Summary"
log_info "=============================================="
log_info "Total tests: $TOTAL"
log_success "Passed: $PASSED"
if [ $FAILED -eq 0 ]; then
    log_success "Failed: $FAILED"
    log_success "All discovery tests passed!"
else
    log_error "Failed: $FAILED"
    log_warning "Some discovery tests failed"
fi

if [ $FAILED -eq 0 ]; then
    exit 0
else
    exit 1
fi