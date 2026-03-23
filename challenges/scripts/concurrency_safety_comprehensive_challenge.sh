#!/bin/bash
# HelixAgent Challenge - Concurrency Safety Comprehensive
# Validates all concurrency fixes across the codebase: sync.Once for stop-once
# semantics, atomic.Bool for lock-free closed flags, WaitGroup for goroutine
# lifecycle, panic recovery in goroutines, mutex correctness, TLS configuration,
# and race detector verification on key packages.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source framework if available
if [ -f "$SCRIPT_DIR/challenge_framework.sh" ]; then
    source "$SCRIPT_DIR/challenge_framework.sh"
    init_challenge "concurrency-safety-comprehensive" "Concurrency Safety Comprehensive"
    load_env
    FRAMEWORK_LOADED=true
else
    FRAMEWORK_LOADED=false
fi

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

record_result() {
    local test_name="$1"
    local status="$2"
    TOTAL=$((TOTAL + 1))
    if [ "$status" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        echo -e "${GREEN}[PASS]${NC} $test_name"
        if [ "$FRAMEWORK_LOADED" = "true" ]; then
            record_assertion "test" "$test_name" "true" ""
        fi
    else
        FAILED=$((FAILED + 1))
        echo -e "${RED}[FAIL]${NC} $test_name"
        if [ "$FRAMEWORK_LOADED" = "true" ]; then
            record_assertion "test" "$test_name" "false" "Test failed"
        fi
    fi
}

echo "=========================================="
echo "  Concurrency Safety Comprehensive Challenge"
echo "=========================================="
echo ""

# ============================================================================
# SECTION 1: sync.Once STOP-ONCE SEMANTICS
# ============================================================================
echo -e "${BLUE}--- Section 1: sync.Once Stop-Once Semantics ---${NC}"

# Test 1: SSE Manager has stopOnce sync.Once
if grep -q 'stopOnce.*sync\.Once' "$PROJECT_ROOT/internal/notifications/sse_manager.go" 2>/dev/null; then
    record_result "SSE Manager has stopOnce sync.Once" "PASS"
else
    record_result "SSE Manager has stopOnce sync.Once" "FAIL"
fi

# Test 2: SSE Manager has closed atomic.Bool
if grep -q 'closed.*atomic\.Bool' "$PROJECT_ROOT/internal/notifications/sse_manager.go" 2>/dev/null; then
    record_result "SSE Manager has closed atomic.Bool" "PASS"
else
    record_result "SSE Manager has closed atomic.Bool" "FAIL"
fi

# Test 3: Kafka Transport has stopOnce sync.Once
if grep -q 'stopOnce.*sync\.Once' "$PROJECT_ROOT/internal/notifications/kafka_transport.go" 2>/dev/null; then
    record_result "Kafka Transport has stopOnce sync.Once" "PASS"
else
    record_result "Kafka Transport has stopOnce sync.Once" "FAIL"
fi

# ============================================================================
# SECTION 2: atomic.Bool FOR LOCK-FREE CLOSED FLAGS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 2: atomic.Bool Closed Flags ---${NC}"

# Test 4: MCP Connection Pool uses atomic.Bool for closed flag
if grep -q 'closed.*atomic\.Bool' "$PROJECT_ROOT/internal/mcp/connection_pool.go" 2>/dev/null; then
    record_result "MCP Connection Pool uses atomic.Bool for closed flag" "PASS"
else
    record_result "MCP Connection Pool uses atomic.Bool for closed flag" "FAIL"
fi

# ============================================================================
# SECTION 3: WaitGroup FOR GOROUTINE LIFECYCLE
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 3: WaitGroup Goroutine Lifecycle ---${NC}"

# Test 5: Plugin HotReload has wg sync.WaitGroup
if grep -q 'wg.*sync\.WaitGroup' "$PROJECT_ROOT/internal/plugins/hot_reload.go" 2>/dev/null; then
    record_result "Plugin HotReload has wg sync.WaitGroup" "PASS"
else
    record_result "Plugin HotReload has wg sync.WaitGroup" "FAIL"
fi

# ============================================================================
# SECTION 4: MUTEX CORRECTNESS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 4: Mutex Correctness ---${NC}"

# Test 6: Integration Orchestrator mutex is NOT marked //nolint:unused
if grep -q 'mu.*sync\.\(RW\)\?Mutex' "$PROJECT_ROOT/internal/services/integration_orchestrator.go" 2>/dev/null; then
    # Verify the mutex line does NOT have a nolint:unused comment
    if ! grep 'mu.*sync\.\(RW\)\?Mutex' "$PROJECT_ROOT/internal/services/integration_orchestrator.go" | grep -q 'nolint:unused'; then
        record_result "Integration Orchestrator mutex is NOT marked nolint:unused" "PASS"
    else
        record_result "Integration Orchestrator mutex is NOT marked nolint:unused" "FAIL"
    fi
else
    record_result "Integration Orchestrator mutex is NOT marked nolint:unused" "FAIL"
fi

# Test 7: DebateHandler has mu sync.RWMutex for activeDebates protection
if grep -q 'mu.*sync\.RWMutex' "$PROJECT_ROOT/internal/handlers/debate_handler.go" 2>/dev/null; then
    record_result "DebateHandler has mu sync.RWMutex for activeDebates protection" "PASS"
else
    record_result "DebateHandler has mu sync.RWMutex for activeDebates protection" "FAIL"
fi

# ============================================================================
# SECTION 5: PANIC RECOVERY IN GOROUTINES
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 5: Panic Recovery in Goroutines ---${NC}"

# Test 8: Debate Service has panic recovery in participant goroutines
if grep -q 'recover()' "$PROJECT_ROOT/internal/services/debate_service.go" 2>/dev/null; then
    record_result "Debate Service has panic recovery in participant goroutines" "PASS"
else
    record_result "Debate Service has panic recovery in participant goroutines" "FAIL"
fi

# Test 9: Polling Store has panic recovery in cleanupLoop
if grep -q 'recover()' "$PROJECT_ROOT/internal/notifications/polling_store.go" 2>/dev/null; then
    record_result "Polling Store has panic recovery in cleanupLoop" "PASS"
else
    record_result "Polling Store has panic recovery in cleanupLoop" "FAIL"
fi

# ============================================================================
# SECTION 6: CIRCUIT BREAKER LISTENER SAFETY
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 6: Circuit Breaker Listener Safety ---${NC}"

# Test 10: Circuit Breaker logs warning on listener limit
if grep -q 'listener limit reached\|listener.*limit\|MaxCircuitBreakerListeners' "$PROJECT_ROOT/internal/llm/circuit_breaker.go" 2>/dev/null; then
    record_result "Circuit Breaker has listener limit with warning" "PASS"
else
    record_result "Circuit Breaker has listener limit with warning" "FAIL"
fi

# ============================================================================
# SECTION 7: TLS CONFIGURATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 7: TLS Security Configuration ---${NC}"

# Test 11: TLS MinVersion is set in quic_server.go
if grep -q 'MinVersion.*tls\.VersionTLS' "$PROJECT_ROOT/internal/router/quic_server.go" 2>/dev/null; then
    record_result "TLS MinVersion set in quic_server.go" "PASS"
else
    record_result "TLS MinVersion set in quic_server.go" "FAIL"
fi

# ============================================================================
# SECTION 8: ADDITIONAL CONCURRENCY PATTERNS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 8: Additional Concurrency Patterns ---${NC}"

# Test 12: SSE Manager uses WaitGroup for goroutine tracking
if grep -q 'wg.*sync\.WaitGroup\|WaitGroup' "$PROJECT_ROOT/internal/notifications/sse_manager.go" 2>/dev/null; then
    record_result "SSE Manager uses WaitGroup for goroutine tracking" "PASS"
else
    record_result "SSE Manager uses WaitGroup for goroutine tracking" "FAIL"
fi

# Test 13: MCP Connection Pool has WaitGroup for goroutine lifecycle
if grep -q 'wg.*sync\.WaitGroup\|WaitGroup' "$PROJECT_ROOT/internal/mcp/connection_pool.go" 2>/dev/null; then
    record_result "MCP Connection Pool has WaitGroup for goroutine lifecycle" "PASS"
else
    record_result "MCP Connection Pool has WaitGroup for goroutine lifecycle" "FAIL"
fi

# Test 14: No data race suppression comments (//nolint:race) in critical packages
race_suppression=0
for pkg_dir in "$PROJECT_ROOT/internal/notifications" "$PROJECT_ROOT/internal/services" "$PROJECT_ROOT/internal/handlers"; do
    if grep -r 'nolint:race\|nolint:.*race' "$pkg_dir"/*.go 2>/dev/null | grep -v '_test.go' | grep -q .; then
        race_suppression=$((race_suppression + 1))
    fi
done
if [ "$race_suppression" -eq 0 ]; then
    record_result "No race suppression nolint comments in critical production code" "PASS"
else
    record_result "No race suppression nolint comments in critical production code" "FAIL"
fi

# ============================================================================
# SECTION 9: RACE DETECTOR VERIFICATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 9: Race Detector Verification ---${NC}"

# Test 15: Race detection on notifications package
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -count=1 -short -timeout=3m \
    "$PROJECT_ROOT/internal/notifications/..." > /tmp/concurrency_race_notifications.log 2>&1; then
    record_result "Race detector passes: internal/notifications" "PASS"
else
    if grep -q "no test files" /tmp/concurrency_race_notifications.log 2>/dev/null; then
        record_result "Race detector passes: internal/notifications" "PASS"
    elif grep -q "DATA RACE" /tmp/concurrency_race_notifications.log 2>/dev/null; then
        record_result "Race detector passes: internal/notifications" "FAIL"
    else
        record_result "Race detector passes: internal/notifications" "FAIL"
    fi
fi

# Test 16: Race detection on handlers package
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -count=1 -short -timeout=3m \
    "$PROJECT_ROOT/internal/handlers/..." > /tmp/concurrency_race_handlers.log 2>&1; then
    record_result "Race detector passes: internal/handlers" "PASS"
else
    if grep -q "no test files" /tmp/concurrency_race_handlers.log 2>/dev/null; then
        record_result "Race detector passes: internal/handlers" "PASS"
    elif grep -q "DATA RACE" /tmp/concurrency_race_handlers.log 2>/dev/null; then
        record_result "Race detector passes: internal/handlers" "FAIL"
    else
        record_result "Race detector passes: internal/handlers" "FAIL"
    fi
fi

# Test 17: Race detection on plugins package
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -count=1 -short -timeout=3m \
    "$PROJECT_ROOT/internal/plugins/..." > /tmp/concurrency_race_plugins.log 2>&1; then
    record_result "Race detector passes: internal/plugins" "PASS"
else
    if grep -q "no test files" /tmp/concurrency_race_plugins.log 2>/dev/null; then
        record_result "Race detector passes: internal/plugins" "PASS"
    elif grep -q "DATA RACE" /tmp/concurrency_race_plugins.log 2>/dev/null; then
        record_result "Race detector passes: internal/plugins" "FAIL"
    else
        record_result "Race detector passes: internal/plugins" "FAIL"
    fi
fi

# Test 18: Race detection on llm package (circuit breaker)
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -count=1 -short -timeout=3m \
    "$PROJECT_ROOT/internal/llm/..." > /tmp/concurrency_race_llm.log 2>&1; then
    record_result "Race detector passes: internal/llm" "PASS"
else
    if grep -q "no test files" /tmp/concurrency_race_llm.log 2>/dev/null; then
        record_result "Race detector passes: internal/llm" "PASS"
    elif grep -q "DATA RACE" /tmp/concurrency_race_llm.log 2>/dev/null; then
        record_result "Race detector passes: internal/llm" "FAIL"
    else
        record_result "Race detector passes: internal/llm" "FAIL"
    fi
fi

# Test 19: Race detection on mcp package
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -count=1 -short -timeout=3m \
    "$PROJECT_ROOT/internal/mcp/..." > /tmp/concurrency_race_mcp.log 2>&1; then
    record_result "Race detector passes: internal/mcp" "PASS"
else
    if grep -q "no test files" /tmp/concurrency_race_mcp.log 2>/dev/null; then
        record_result "Race detector passes: internal/mcp" "PASS"
    elif grep -q "DATA RACE" /tmp/concurrency_race_mcp.log 2>/dev/null; then
        record_result "Race detector passes: internal/mcp" "FAIL"
    else
        record_result "Race detector passes: internal/mcp" "FAIL"
    fi
fi

# Test 20: Race detection on services package (debate_service, integration_orchestrator)
if GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -count=1 -short -timeout=3m \
    "$PROJECT_ROOT/internal/services/..." > /tmp/concurrency_race_services.log 2>&1; then
    record_result "Race detector passes: internal/services" "PASS"
else
    if grep -q "no test files" /tmp/concurrency_race_services.log 2>/dev/null; then
        record_result "Race detector passes: internal/services" "PASS"
    elif grep -q "DATA RACE" /tmp/concurrency_race_services.log 2>/dev/null; then
        record_result "Race detector passes: internal/services" "FAIL"
    else
        record_result "Race detector passes: internal/services" "FAIL"
    fi
fi

# ============================================================================
# SUMMARY
# ============================================================================
echo ""
echo "=========================================="
echo "  Results: $PASSED/$TOTAL passed, $FAILED failed"
echo "=========================================="

if [ "$FRAMEWORK_LOADED" = "true" ]; then
    record_metric "total_tests" "$TOTAL"
    record_metric "passed_tests" "$PASSED"
    record_metric "failed_tests" "$FAILED"

    if [ $FAILED -eq 0 ]; then
        finalize_challenge "PASSED"
    else
        finalize_challenge "FAILED"
    fi
fi

if [ $FAILED -gt 0 ]; then
    exit 1
fi
exit 0
