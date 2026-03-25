#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

record_result() {
    TOTAL=$((TOTAL + 1))
    if [ "$2" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        echo -e "${GREEN}[PASS]${NC} $1"
    else
        FAILED=$((FAILED + 1))
        echo -e "${RED}[FAIL]${NC} $1"
    fi
}

echo "=========================================="
echo "  Monitoring Metrics Accuracy Challenge"
echo "=========================================="
echo ""

# --------------------------------------------------------------------------
# Test 1: circuit_breaker_monitor.go exists in internal/services/
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" ]; then
    record_result "circuit_breaker_monitor.go exists" "PASS"
else
    record_result "circuit_breaker_monitor.go exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 2: provider_health_monitor.go exists in internal/services/
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/internal/services/provider_health_monitor.go" ]; then
    record_result "provider_health_monitor.go exists" "PASS"
else
    record_result "provider_health_monitor.go exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 3: concurrency_monitor.go exists in internal/services/
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/internal/services/concurrency_monitor.go" ]; then
    record_result "concurrency_monitor.go exists" "PASS"
else
    record_result "concurrency_monitor.go exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 4: oauth_token_monitor.go exists in internal/services/
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/internal/services/oauth_token_monitor.go" ]; then
    record_result "oauth_token_monitor.go exists" "PASS"
else
    record_result "oauth_token_monitor.go exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 5: sync.Once (metricsOnce) pattern used in circuit_breaker_monitor.go
# --------------------------------------------------------------------------
if grep -q "metricsOnce\|sync\.Once" "$PROJECT_ROOT/internal/services/circuit_breaker_monitor.go" 2>/dev/null; then
    record_result "sync.Once pattern in circuit_breaker_monitor.go" "PASS"
else
    record_result "sync.Once pattern in circuit_breaker_monitor.go" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 6: sync.Once (metricsOnce) pattern used in provider_health_monitor.go
# --------------------------------------------------------------------------
if grep -q "metricsOnce\|sync\.Once" "$PROJECT_ROOT/internal/services/provider_health_monitor.go" 2>/dev/null; then
    record_result "sync.Once pattern in provider_health_monitor.go" "PASS"
else
    record_result "sync.Once pattern in provider_health_monitor.go" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 7: OpenTelemetry metrics file exists (internal/observability/metrics.go)
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/internal/observability/metrics.go" ]; then
    record_result "OpenTelemetry metrics.go exists" "PASS"
else
    record_result "OpenTelemetry metrics.go exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 8: Extended metrics file exists (metrics_extended.go)
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/internal/observability/metrics_extended.go" ]; then
    record_result "metrics_extended.go exists" "PASS"
else
    record_result "metrics_extended.go exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 9: Monitoring test files exist in tests/monitoring/
# --------------------------------------------------------------------------
MONITORING_TEST_COUNT=$(find "$PROJECT_ROOT/tests/monitoring/" -name "*_test.go" 2>/dev/null | wc -l)
if [ "$MONITORING_TEST_COUNT" -ge 3 ]; then
    record_result "Monitoring test files >= 3 (found: $MONITORING_TEST_COUNT)" "PASS"
else
    record_result "Monitoring test files >= 3 (found: $MONITORING_TEST_COUNT)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 10: metrics_collection_test.go exists in tests/monitoring/
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/monitoring/metrics_collection_test.go" ]; then
    record_result "metrics_collection_test.go exists" "PASS"
else
    record_result "metrics_collection_test.go exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 11: circuit_breaker_transitions_test.go exists in tests/monitoring/
# --------------------------------------------------------------------------
if [ -f "$PROJECT_ROOT/tests/monitoring/circuit_breaker_transitions_test.go" ]; then
    record_result "circuit_breaker_transitions_test.go exists" "PASS"
else
    record_result "circuit_breaker_transitions_test.go exists" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 12: At least 5 metric registrations (promauto/prometheus/otel) in codebase
# --------------------------------------------------------------------------
METRIC_REG_COUNT=$(grep -r "promauto\.New\|prometheus\.New\|metric\.Int64\|metric\.Float64" \
    "$PROJECT_ROOT/internal/" --include="*.go" | grep -v "_test\.go" | wc -l)
if [ "$METRIC_REG_COUNT" -ge 5 ]; then
    record_result "At least 5 metric registrations (found: $METRIC_REG_COUNT)" "PASS"
else
    record_result "At least 5 metric registrations (found: $METRIC_REG_COUNT)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 13: metricsOnce pattern used codebase-wide (>= 3 occurrences)
# --------------------------------------------------------------------------
METRICS_ONCE_COUNT=$(grep -r "metricsOnce\|MetricsOnce" "$PROJECT_ROOT/internal/" --include="*.go" \
    | grep -v "_test\.go" | wc -l)
if [ "$METRICS_ONCE_COUNT" -ge 3 ]; then
    record_result "metricsOnce pattern >= 3 occurrences (found: $METRICS_ONCE_COUNT)" "PASS"
else
    record_result "metricsOnce pattern >= 3 occurrences (found: $METRICS_ONCE_COUNT)" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 14: OpenTelemetry or Prometheus import in observability package
# --------------------------------------------------------------------------
if grep -q "prometheus\|opentelemetry\|go\.opentelemetry\.io\|otel" \
    "$PROJECT_ROOT/internal/observability/metrics.go" 2>/dev/null; then
    record_result "Metrics library import in observability/metrics.go" "PASS"
else
    record_result "Metrics library import in observability/metrics.go" "FAIL"
fi

# --------------------------------------------------------------------------
# Test 15: Monitoring tests compile successfully
# --------------------------------------------------------------------------
cd "$PROJECT_ROOT"
if GOMAXPROCS=2 go test -run=XXX_NOMATCH ./tests/monitoring/ 2>/dev/null; then
    record_result "Monitoring tests compile successfully" "PASS"
else
    record_result "Monitoring tests compile successfully" "FAIL"
fi

echo ""
echo "=========================================="
echo "  Results: $PASSED/$TOTAL passed, $FAILED failed"
echo "=========================================="

[ $FAILED -eq 0 ] && exit 0 || exit 1
