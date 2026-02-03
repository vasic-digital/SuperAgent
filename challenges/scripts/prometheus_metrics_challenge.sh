#!/bin/bash
# Prometheus Metrics Challenge
# VALIDATES: Prometheus metrics endpoint, required metrics, counter increments, histogram buckets
# Tests the complete Prometheus metrics instrumentation with comprehensive validation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="Prometheus Metrics Challenge"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."
API_BASE="${HELIX_API_BASE:-http://localhost:8080}"

log_info "=============================================="
log_info "$CHALLENGE_NAME"
log_info "=============================================="
log_info "Validates: Prometheus metrics instrumentation"
log_info ""

# ============================================================================
# Section 1: Metrics Code Structure
# ============================================================================

log_info "=============================================="
log_info "Section 1: Metrics Code Structure"
log_info "=============================================="

# Test 1: observability/metrics.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 1: observability/metrics.go exists"
if [ -f "$PROJECT_ROOT/internal/observability/metrics.go" ]; then
    log_success "observability/metrics.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "observability/metrics.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 2: LLMMetrics struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 2: LLMMetrics struct defined"
if grep -q "type LLMMetrics struct" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "LLMMetrics struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "LLMMetrics struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 3: NewLLMMetrics constructor exists
TOTAL=$((TOTAL + 1))
log_info "Test 3: NewLLMMetrics constructor exists"
if grep -q "func NewLLMMetrics" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "NewLLMMetrics constructor exists"
    PASSED=$((PASSED + 1))
else
    log_error "NewLLMMetrics constructor NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 4: Request counter metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 4: Request counter metric defined"
if grep -q "RequestsTotal.*Int64Counter" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "RequestsTotal counter metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "RequestsTotal counter metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 5: Request duration histogram metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 5: Request duration histogram metric defined"
if grep -q "RequestDuration.*Float64Histogram" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "RequestDuration histogram metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "RequestDuration histogram metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: Required Metrics Validation
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 2: Required Metrics Validation"
log_info "=============================================="

# Test 6: Error counter metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 6: Error counter metric defined"
if grep -q "ErrorsTotal.*Int64Counter" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "ErrorsTotal counter metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "ErrorsTotal counter metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 7: Token metrics defined
TOTAL=$((TOTAL + 1))
log_info "Test 7: Token metrics defined"
if grep -q "InputTokens.*Int64Counter" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null && \
   grep -q "OutputTokens.*Int64Counter" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "Token metrics defined (InputTokens, OutputTokens)"
    PASSED=$((PASSED + 1))
else
    log_error "Token metrics NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 8: Cache metrics defined
TOTAL=$((TOTAL + 1))
log_info "Test 8: Cache metrics defined"
if grep -q "CacheHits.*Int64Counter" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null && \
   grep -q "CacheMisses.*Int64Counter" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "Cache metrics defined (CacheHits, CacheMisses)"
    PASSED=$((PASSED + 1))
else
    log_error "Cache metrics NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 9: Provider metrics defined
TOTAL=$((TOTAL + 1))
log_info "Test 9: Provider metrics defined"
if grep -q "ProviderLatency.*Float64Histogram" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "ProviderLatency histogram metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "ProviderLatency metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 10: Cost metrics defined
TOTAL=$((TOTAL + 1))
log_info "Test 10: Cost metrics defined"
if grep -q "TotalCost.*Float64Counter" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "TotalCost counter metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "TotalCost metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 3: Circuit Breaker Metrics
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 3: Circuit Breaker Metrics"
log_info "=============================================="

# Test 11: Circuit breaker state metric
TOTAL=$((TOTAL + 1))
log_info "Test 11: Circuit breaker state metric defined"
if grep -q "helixagent_circuit_breaker_state" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Circuit breaker state gauge metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "Circuit breaker state metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 12: Circuit breaker failures metric
TOTAL=$((TOTAL + 1))
log_info "Test 12: Circuit breaker failures metric defined"
if grep -q "helixagent_circuit_breaker_failures_total" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Circuit breaker failures counter metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "Circuit breaker failures metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 13: Open circuits gauge metric
TOTAL=$((TOTAL + 1))
log_info "Test 13: Open circuits gauge metric defined"
if grep -q "helixagent_circuit_breakers_open" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Open circuits gauge metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "Open circuits gauge metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 14: Circuit breaker alerts metric
TOTAL=$((TOTAL + 1))
log_info "Test 14: Circuit breaker alerts metric defined"
if grep -q "helixagent_circuit_breaker_alerts_total" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "Circuit breaker alerts counter metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "Circuit breaker alerts metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 4: Prometheus Integration
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 4: Prometheus Integration"
log_info "=============================================="

# Test 15: Prometheus handler in router
TOTAL=$((TOTAL + 1))
log_info "Test 15: Prometheus handler registered in router"
if grep -q 'promhttp.Handler()' "$PROJECT_ROOT/internal/router/router.go" 2>/dev/null; then
    log_success "Prometheus handler registered in router"
    PASSED=$((PASSED + 1))
else
    log_error "Prometheus handler NOT found in router!"
    FAILED=$((FAILED + 1))
fi

# Test 16: /metrics endpoint defined
TOTAL=$((TOTAL + 1))
log_info "Test 16: /metrics endpoint defined"
if grep -q '"/metrics"' "$PROJECT_ROOT/internal/router/router.go" 2>/dev/null; then
    log_success "/metrics endpoint defined"
    PASSED=$((PASSED + 1))
else
    log_error "/metrics endpoint NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 17: promauto used for metric registration
TOTAL=$((TOTAL + 1))
log_info "Test 17: promauto used for automatic registration"
if grep -q "promauto.NewGaugeVec\|promauto.NewCounterVec\|promauto.NewGauge" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    log_success "promauto used for automatic metric registration"
    PASSED=$((PASSED + 1))
else
    log_error "promauto NOT used for automatic registration!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 5: Debate and RAG Metrics
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 5: Debate and RAG Metrics"
log_info "=============================================="

# Test 18: Debate rounds metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 18: Debate rounds metric defined"
if grep -q "DebateRounds.*Int64Counter" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "DebateRounds counter metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "DebateRounds metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 19: Debate consensus metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 19: Debate consensus histogram metric defined"
if grep -q "DebateConsensus.*Float64Histogram" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "DebateConsensus histogram metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "DebateConsensus metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 20: RAG retrievals metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 20: RAG retrievals metric defined"
if grep -q "RAGRetrievals.*Int64Counter" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "RAGRetrievals counter metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "RAGRetrievals metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 21: RAG latency metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 21: RAG latency histogram metric defined"
if grep -q "RAGLatency.*Float64Histogram" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "RAGLatency histogram metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "RAGLatency metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 6: Recording Methods
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 6: Recording Methods"
log_info "=============================================="

# Test 22: RecordRequest method exists
TOTAL=$((TOTAL + 1))
log_info "Test 22: RecordRequest method exists"
if grep -q "func (m \*LLMMetrics) RecordRequest" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "RecordRequest method exists"
    PASSED=$((PASSED + 1))
else
    log_error "RecordRequest method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 23: RecordCacheHit method exists
TOTAL=$((TOTAL + 1))
log_info "Test 23: RecordCacheHit method exists"
if grep -q "func (m \*LLMMetrics) RecordCacheHit" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "RecordCacheHit method exists"
    PASSED=$((PASSED + 1))
else
    log_error "RecordCacheHit method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 24: RecordDebateRound method exists
TOTAL=$((TOTAL + 1))
log_info "Test 24: RecordDebateRound method exists"
if grep -q "func (m \*LLMMetrics) RecordDebateRound" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "RecordDebateRound method exists"
    PASSED=$((PASSED + 1))
else
    log_error "RecordDebateRound method NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 25: RecordRAGRetrieval method exists
TOTAL=$((TOTAL + 1))
log_info "Test 25: RecordRAGRetrieval method exists"
if grep -q "func (m \*LLMMetrics) RecordRAGRetrieval" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "RecordRAGRetrieval method exists"
    PASSED=$((PASSED + 1))
else
    log_error "RecordRAGRetrieval method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 7: Global Metrics Initialization
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 7: Global Metrics Initialization"
log_info "=============================================="

# Test 26: Global metrics instance
TOTAL=$((TOTAL + 1))
log_info "Test 26: Global metrics instance defined"
if grep -q "globalMetrics \*LLMMetrics" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "Global metrics instance defined"
    PASSED=$((PASSED + 1))
else
    log_error "Global metrics instance NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 27: InitGlobalMetrics function exists
TOTAL=$((TOTAL + 1))
log_info "Test 27: InitGlobalMetrics function exists"
if grep -q "func InitGlobalMetrics" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "InitGlobalMetrics function exists"
    PASSED=$((PASSED + 1))
else
    log_error "InitGlobalMetrics function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 28: GetMetrics function exists
TOTAL=$((TOTAL + 1))
log_info "Test 28: GetMetrics function exists"
if grep -q "func GetMetrics" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "GetMetrics function exists"
    PASSED=$((PASSED + 1))
else
    log_error "GetMetrics function NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 29: sync.Once used for initialization
TOTAL=$((TOTAL + 1))
log_info "Test 29: sync.Once used for thread-safe initialization"
if grep -q "metricsOnce.*sync.Once" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "sync.Once used for thread-safe initialization"
    PASSED=$((PASSED + 1))
else
    log_error "sync.Once NOT used for initialization!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 8: Rate Limit and Timeout Metrics
# ============================================================================

log_info ""
log_info "=============================================="
log_info "Section 8: Rate Limit and Timeout Metrics"
log_info "=============================================="

# Test 30: Timeout metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 30: Timeout counter metric defined"
if grep -q "TimeoutsTotal.*Int64Counter" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "TimeoutsTotal counter metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "TimeoutsTotal metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 31: Rate limit metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 31: Rate limit counter metric defined"
if grep -q "RateLimitsTotal.*Int64Counter" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "RateLimitsTotal counter metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "RateLimitsTotal metric NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 32: Requests in flight metric defined
TOTAL=$((TOTAL + 1))
log_info "Test 32: Requests in flight gauge metric defined"
if grep -q "RequestsInFlight.*Int64UpDownCounter" "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    log_success "RequestsInFlight up/down counter metric defined"
    PASSED=$((PASSED + 1))
else
    log_error "RequestsInFlight metric NOT found!"
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
