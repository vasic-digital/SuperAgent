#!/bin/bash
# Circuit Breaker Metrics Challenge
# VALIDATES: Circuit breaker state metrics, state transitions, recovery patterns
# Tests the complete circuit breaker observability infrastructure

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="Circuit Breaker Metrics Challenge"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."

log_info "=============================================="
log_info "$CHALLENGE_NAME"
log_info "=============================================="
log_info "Validates: Circuit breaker observability"
log_info ""

# ============================================================================
# Section 1: Circuit Breaker Monitor Structure
# ============================================================================

log_info "=============================================="
log_info "Section 1: Circuit Breaker Monitor Structure"
log_info "=============================================="

# Test 1: circuit_breaker_monitor.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 1: circuit_breaker_monitor.go exists"
if [ -f "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" ]; then
    log_success "circuit_breaker_monitor.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "circuit_breaker_monitor.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 2: CircuitBreakerMonitor struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 2: CircuitBreakerMonitor struct defined"
if grep -q "type CircuitBreakerMonitor struct" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "CircuitBreakerMonitor struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "CircuitBreakerMonitor struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 3: NewCircuitBreakerMonitor constructor exists
TOTAL=$((TOTAL + 1))
log_info "Test 3: NewCircuitBreakerMonitor constructor exists"
if grep -q "func NewCircuitBreakerMonitor" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "NewCircuitBreakerMonitor constructor exists"
    PASSED=$((PASSED + 1))
else
    log_error "NewCircuitBreakerMonitor constructor NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 4: CircuitBreakerMonitorConfig struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 4: CircuitBreakerMonitorConfig struct defined"
if grep -q "type CircuitBreakerMonitorConfig struct" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "CircuitBreakerMonitorConfig struct defined"
    PASSED=$((PASSED + 1))
else
    log_error "CircuitBreakerMonitorConfig struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 5: DefaultCircuitBreakerMonitorConfig function exists
TOTAL=$((TOTAL + 1))
log_info "Test 5: DefaultCircuitBreakerMonitorConfig function exists"
if grep -q "func DefaultCircuitBreakerMonitorConfig" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "DefaultCircuitBreakerMonitorConfig function exists"
    PASSED=$((PASSED + 1))
else
    log_error "DefaultCircuitBreakerMonitorConfig function NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: Prometheus Metrics Definition
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 2: Prometheus Metrics Definition"
log_info "=============================================="

# Test 6: Circuit state gauge metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 6: Circuit state gauge metric defined"
if grep -q "cbmCircuitStateGauge.*GaugeVec" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Circuit state gauge metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "Circuit state gauge metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 7: Circuit failures counter metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 7: Circuit failures counter metric defined"
if grep -q "cbmCircuitFailuresTotal.*CounterVec" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Circuit failures counter metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "Circuit failures counter metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 8: Open circuits gauge metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 8: Open circuits gauge metric defined"
if grep -q "cbmOpenCircuitsGauge.*Gauge" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Open circuits gauge metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "Open circuits gauge metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 9: Alerts counter metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 9: Alerts counter metric defined"
if grep -q "cbmAlertsTotal.*Counter" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Alerts counter metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "Alerts counter metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 10: Metric name with provider label
TOTAL=$((TOTAL + 1))
log_info "Test 10: Metrics have provider label"
if grep -q '"provider"' "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Metrics have provider label"
    PASSED=$((PASSED + 1))
else
    log_error "Metrics do NOT have provider label!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 3: State Transition Handling
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 3: State Transition Handling"
log_info "=============================================="

# Test 11: CircuitState enumeration used
TOTAL=$((TOTAL + 1))
log_info "Test 11: CircuitState enumeration used"
if grep -q "llm.CircuitClosed\|llm.CircuitHalfOpen\|llm.CircuitOpen" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "CircuitState enumeration used"
    PASSED=$((PASSED + 1))
else
    log_error "CircuitState enumeration NOT used!"
    FAILED=$((FAILED + 1))
fi

# Test 12: State value mapping (0=closed, 1=half_open, 2=open)
TOTAL=$((TOTAL + 1))
log_info "Test 12: State value mapping implemented"
if grep -A30 "func (cbm \*CircuitBreakerMonitor) checkCircuitBreakers" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null | grep -q "stateValue"; then
    log_success "State value mapping implemented"
    PASSED=$((PASSED + 1))
else
    log_error "State value mapping NOT implemented!"
    FAILED=$((FAILED + 1))
fi

# Test 13: checkCircuitBreakers method exists
TOTAL=$((TOTAL + 1))
log_info "Test 13: checkCircuitBreakers method exists"
if grep -q "func (cbm \*CircuitBreakerMonitor) checkCircuitBreakers" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "checkCircuitBreakers method exists"
    PASSED=$((PASSED + 1))
else
    log_error "checkCircuitBreakers method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 14: registerStateChangeListeners method exists
TOTAL=$((TOTAL + 1))
log_info "Test 14: registerStateChangeListeners method exists"
if grep -q "func (cbm \*CircuitBreakerMonitor) registerStateChangeListeners" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "registerStateChangeListeners method exists"
    PASSED=$((PASSED + 1))
else
    log_error "registerStateChangeListeners method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 4: Alert System
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 4: Alert System"
log_info "=============================================="

# Test 15: CircuitBreakerAlert type defined
TOTAL=$((TOTAL + 1))
log_info "Test 15: CircuitBreakerAlert type defined"
if grep -q "type CircuitBreakerAlert struct" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "CircuitBreakerAlert type defined"
    PASSED=$((PASSED + 1))
else
    log_error "CircuitBreakerAlert type NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 16: CircuitBreakerAlertListener type defined
TOTAL=$((TOTAL + 1))
log_info "Test 16: CircuitBreakerAlertListener type defined"
if grep -q "type CircuitBreakerAlertListener func" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "CircuitBreakerAlertListener type defined"
    PASSED=$((PASSED + 1))
else
    log_error "CircuitBreakerAlertListener type NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 17: AddAlertListener method exists
TOTAL=$((TOTAL + 1))
log_info "Test 17: AddAlertListener method exists"
if grep -q "func (cbm \*CircuitBreakerMonitor) AddAlertListener" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "AddAlertListener method exists"
    PASSED=$((PASSED + 1))
else
    log_error "AddAlertListener method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 18: sendAlert method exists
TOTAL=$((TOTAL + 1))
log_info "Test 18: sendAlert method exists"
if grep -q "func (cbm \*CircuitBreakerMonitor) sendAlert" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "sendAlert method exists"
    PASSED=$((PASSED + 1))
else
    log_error "sendAlert method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 19: Alert threshold checking
TOTAL=$((TOTAL + 1))
log_info "Test 19: Alert threshold checking implemented"
if grep -q "alertThreshold" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Alert threshold checking implemented"
    PASSED=$((PASSED + 1))
else
    log_error "Alert threshold checking NOT implemented!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 5: Monitoring Lifecycle
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 5: Monitoring Lifecycle"
log_info "=============================================="

# Test 20: Start method exists
TOTAL=$((TOTAL + 1))
log_info "Test 20: Start method exists"
if grep -q "func (cbm \*CircuitBreakerMonitor) Start" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Start method exists"
    PASSED=$((PASSED + 1))
else
    log_error "Start method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 21: Stop method exists
TOTAL=$((TOTAL + 1))
log_info "Test 21: Stop method exists"
if grep -q "func (cbm \*CircuitBreakerMonitor) Stop" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Stop method exists"
    PASSED=$((PASSED + 1))
else
    log_error "Stop method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 22: Context cancellation support
TOTAL=$((TOTAL + 1))
log_info "Test 22: Context cancellation support"
if grep -q "ctx.Done()" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Context cancellation support"
    PASSED=$((PASSED + 1))
else
    log_error "Context cancellation NOT supported!"
    FAILED=$((FAILED + 1))
fi

# Test 23: Ticker-based periodic checking
TOTAL=$((TOTAL + 1))
log_info "Test 23: Ticker-based periodic checking"
if grep -q "time.NewTicker\|ticker.C" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Ticker-based periodic checking"
    PASSED=$((PASSED + 1))
else
    log_error "Ticker-based periodic checking NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 24: Running flag for state management
TOTAL=$((TOTAL + 1))
log_info "Test 24: Running flag for state management"
if grep -q "running.*bool" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Running flag for state management"
    PASSED=$((PASSED + 1))
else
    log_error "Running flag NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 6: Status Reporting
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 6: Status Reporting"
log_info "=============================================="

# Test 25: GetStatus method exists
TOTAL=$((TOTAL + 1))
log_info "Test 25: GetStatus method exists"
if grep -q "func (cbm \*CircuitBreakerMonitor) GetStatus" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "GetStatus method exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetStatus method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 26: CircuitBreakerStatus type defined
TOTAL=$((TOTAL + 1))
log_info "Test 26: CircuitBreakerStatus type defined"
if grep -q "type CircuitBreakerStatus struct" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "CircuitBreakerStatus type defined"
    PASSED=$((PASSED + 1))
else
    log_error "CircuitBreakerStatus type NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 27: CircuitBreakerProviderStatus type defined
TOTAL=$((TOTAL + 1))
log_info "Test 27: CircuitBreakerProviderStatus type defined"
if grep -q "type CircuitBreakerProviderStatus struct" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "CircuitBreakerProviderStatus type defined"
    PASSED=$((PASSED + 1))
else
    log_error "CircuitBreakerProviderStatus type NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 28: Status includes health flag
TOTAL=$((TOTAL + 1))
log_info "Test 28: Status includes health flag"
if grep -q 'Healthy.*bool.*json:"healthy"' "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Status includes health flag"
    PASSED=$((PASSED + 1))
else
    log_error "Status does NOT include health flag!"
    FAILED=$((FAILED + 1))
fi

# Test 29: Status includes open/half-open/closed counts
TOTAL=$((TOTAL + 1))
log_info "Test 29: Status includes state counts"
if grep -q "OpenCount\|HalfOpenCount\|ClosedCount" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Status includes state counts"
    PASSED=$((PASSED + 1))
else
    log_error "Status does NOT include state counts!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 7: Recovery Operations
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 7: Recovery Operations"
log_info "=============================================="

# Test 30: ResetCircuitBreaker method exists
TOTAL=$((TOTAL + 1))
log_info "Test 30: ResetCircuitBreaker method exists"
if grep -q "func (cbm \*CircuitBreakerMonitor) ResetCircuitBreaker" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "ResetCircuitBreaker method exists"
    PASSED=$((PASSED + 1))
else
    log_error "ResetCircuitBreaker method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 31: ResetAllCircuitBreakers method exists
TOTAL=$((TOTAL + 1))
log_info "Test 31: ResetAllCircuitBreakers method exists"
if grep -q "func (cbm \*CircuitBreakerMonitor) ResetAllCircuitBreakers" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "ResetAllCircuitBreakers method exists"
    PASSED=$((PASSED + 1))
else
    log_error "ResetAllCircuitBreakers method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 32: calculateSuccessRate helper function
TOTAL=$((TOTAL + 1))
log_info "Test 32: calculateSuccessRate helper function exists"
if grep -q "func calculateSuccessRate" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "calculateSuccessRate helper function exists"
    PASSED=$((PASSED + 1))
else
    log_error "calculateSuccessRate helper function NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 8: Thread Safety
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 8: Thread Safety"
log_info "=============================================="

# Test 33: Mutex used for thread safety
TOTAL=$((TOTAL + 1))
log_info "Test 33: Mutex used for thread safety"
if grep -q "mu.*sync.RWMutex\|mu.*sync.Mutex" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Mutex used for thread safety"
    PASSED=$((PASSED + 1))
else
    log_error "Mutex NOT used for thread safety!"
    FAILED=$((FAILED + 1))
fi

# Test 34: Lock/Unlock calls present
TOTAL=$((TOTAL + 1))
log_info "Test 34: Lock/Unlock calls present"
if grep -q "mu.Lock()\|mu.RLock()" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Lock/Unlock calls present"
    PASSED=$((PASSED + 1))
else
    log_error "Lock/Unlock calls NOT present!"
    FAILED=$((FAILED + 1))
fi

# Test 35: sync.Once used for metric initialization
TOTAL=$((TOTAL + 1))
log_info "Test 35: sync.Once used for metric initialization"
if grep -q "cbmMetricsOnce.*sync.Once" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "sync.Once used for metric initialization"
    PASSED=$((PASSED + 1))
else
    log_error "sync.Once NOT used for metric initialization!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 9: Logging and Debugging
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 9: Logging and Debugging"
log_info "=============================================="

# Test 36: Logger field in struct
TOTAL=$((TOTAL + 1))
log_info "Test 36: Logger field in struct"
if grep -q "logger.*logrus.Logger" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Logger field in struct"
    PASSED=$((PASSED + 1))
else
    log_error "Logger field NOT in struct!"
    FAILED=$((FAILED + 1))
fi

# Test 37: Warning log for open circuits
TOTAL=$((TOTAL + 1))
log_info "Test 37: Warning log for open circuits"
if grep -q 'Warn.*"Circuit breaker is OPEN"\|logger.WithFields.*Warn' "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Warning log for open circuits"
    PASSED=$((PASSED + 1))
else
    log_error "Warning log for open circuits NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 38: Debug log for status checks
TOTAL=$((TOTAL + 1))
log_info "Test 38: Debug log for status checks"
if grep -q "Debug.*status check\|Debug.*completed" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Debug log for status checks"
    PASSED=$((PASSED + 1))
else
    log_error "Debug log for status checks NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 39: Info log for start/stop
TOTAL=$((TOTAL + 1))
log_info "Test 39: Info log for start/stop"
if grep -q "Info.*started\|Info.*stopped" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Info log for start/stop"
    PASSED=$((PASSED + 1))
else
    log_error "Info log for start/stop NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 10: Integration with CircuitBreakerManager
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 10: Integration with CircuitBreakerManager"
log_info "=============================================="

# Test 40: CircuitBreakerManager field
TOTAL=$((TOTAL + 1))
log_info "Test 40: CircuitBreakerManager field in struct"
if grep -q "manager.*llm.CircuitBreakerManager" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "CircuitBreakerManager field in struct"
    PASSED=$((PASSED + 1))
else
    log_error "CircuitBreakerManager field NOT in struct!"
    FAILED=$((FAILED + 1))
fi

# Test 41: GetAllStats usage
TOTAL=$((TOTAL + 1))
log_info "Test 41: GetAllStats method used"
if grep -q "manager.GetAllStats()" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "GetAllStats method used"
    PASSED=$((PASSED + 1))
else
    log_error "GetAllStats method NOT used!"
    FAILED=$((FAILED + 1))
fi

# Test 42: ResetAll usage
TOTAL=$((TOTAL + 1))
log_info "Test 42: ResetAll method used"
if grep -q "manager.ResetAll()" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "ResetAll method used"
    PASSED=$((PASSED + 1))
else
    log_error "ResetAll method NOT used!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Results Summary
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Challenge Results Summary"
log_info "=============================================="
log_info "Passed: $PASSED/$TOTAL"
log_info "Failed: $FAILED/$TOTAL"
log_info ""

if [ "$FAILED" -eq 0 ]; then
    log_success "ALL $TOTAL TESTS PASSED!"
    exit 0
else
    log_error "$FAILED TEST(S) FAILED!"
    exit 1
fi
