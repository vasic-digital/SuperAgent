#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
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
echo "  Monitoring Dashboard Challenge"
echo "=========================================="

# Test 1: Monitoring compose file exists
if [ -f "$PROJECT_ROOT/docker-compose.monitoring.yml" ]; then
    record_result "Monitoring compose file exists" "PASS"
else
    record_result "Monitoring compose file exists" "FAIL"
fi

# Test 2: Prometheus service defined
if grep -q "prometheus" "$PROJECT_ROOT/docker-compose.monitoring.yml" 2>/dev/null; then
    record_result "Prometheus service defined" "PASS"
else
    record_result "Prometheus service defined" "FAIL"
fi

# Test 3: Grafana service defined
if grep -q "grafana" "$PROJECT_ROOT/docker-compose.monitoring.yml" 2>/dev/null; then
    record_result "Grafana service defined" "PASS"
else
    record_result "Grafana service defined" "FAIL"
fi

# Test 4: Alertmanager service defined
if grep -q "alertmanager" "$PROJECT_ROOT/docker-compose.monitoring.yml" 2>/dev/null; then
    record_result "Alertmanager service defined" "PASS"
else
    record_result "Alertmanager service defined" "FAIL"
fi

# Test 5: Node exporter defined
if grep -q "node-exporter" "$PROJECT_ROOT/docker-compose.monitoring.yml" 2>/dev/null; then
    record_result "Node exporter service defined" "PASS"
else
    record_result "Node exporter service defined" "FAIL"
fi

# Test 6: Images are version-pinned
LATEST_COUNT=$(grep -c ":latest" "$PROJECT_ROOT/docker-compose.monitoring.yml" 2>/dev/null || true)
if [ "$LATEST_COUNT" -eq 0 ]; then
    record_result "All monitoring images version-pinned" "PASS"
else
    record_result "All monitoring images version-pinned ($LATEST_COUNT use :latest)" "FAIL"
fi

# Test 7: Monitoring tests exist (>= 5 files)
MONITORING_TESTS=$(find "$PROJECT_ROOT/tests/monitoring/" -name "*_test.go" 2>/dev/null | wc -l)
if [ "$MONITORING_TESTS" -ge 5 ]; then
    record_result "Monitoring tests >= 5 files (found: $MONITORING_TESTS)" "PASS"
else
    record_result "Monitoring tests >= 5 files (found: $MONITORING_TESTS)" "FAIL"
fi

# Test 8: Health endpoint test exists
if [ -f "$PROJECT_ROOT/tests/monitoring/health_endpoints_test.go" ]; then
    record_result "Health endpoint test exists" "PASS"
else
    record_result "Health endpoint test exists" "FAIL"
fi

# Test 9: Metrics collection test exists
if [ -f "$PROJECT_ROOT/tests/monitoring/metrics_collection_test.go" ]; then
    record_result "Metrics collection test exists" "PASS"
else
    record_result "Metrics collection test exists" "FAIL"
fi

# Test 10: Circuit breaker monitoring test exists
if [ -f "$PROJECT_ROOT/tests/monitoring/circuit_breaker_transitions_test.go" ]; then
    record_result "Circuit breaker monitoring test exists" "PASS"
else
    record_result "Circuit breaker monitoring test exists" "FAIL"
fi

# Test 11: Provider latency tracking test exists
if [ -f "$PROJECT_ROOT/tests/monitoring/provider_latency_tracking_test.go" ]; then
    record_result "Provider latency tracking test exists" "PASS"
else
    record_result "Provider latency tracking test exists" "FAIL"
fi

# Test 12: Cache hit ratio test exists
if [ -f "$PROJECT_ROOT/tests/monitoring/cache_hit_ratio_test.go" ]; then
    record_result "Cache hit ratio test exists" "PASS"
else
    record_result "Cache hit ratio test exists" "FAIL"
fi

# Test 13: Database query performance test exists
if [ -f "$PROJECT_ROOT/tests/monitoring/database_query_performance_test.go" ]; then
    record_result "Database query performance test exists" "PASS"
else
    record_result "Database query performance test exists" "FAIL"
fi

# Test 14: Monitoring tests compile
cd "$PROJECT_ROOT"
if GOMAXPROCS=2 go test -run=XXX_NOMATCH ./tests/monitoring/ 2>/dev/null; then
    record_result "Monitoring tests compile successfully" "PASS"
else
    record_result "Monitoring tests compile successfully" "FAIL"
fi

# Test 15: Prometheus configuration referenced
if [ -d "$PROJECT_ROOT/configs/prometheus" ] || [ -d "$PROJECT_ROOT/docker/monitoring" ] || grep -rq "prometheus.yml" "$PROJECT_ROOT/docker-compose.monitoring.yml" 2>/dev/null; then
    record_result "Prometheus configuration referenced" "PASS"
else
    record_result "Prometheus configuration referenced" "FAIL"
fi

echo ""
echo "=========================================="
echo "  Results: $PASSED/$TOTAL passed, $FAILED failed"
echo "=========================================="

[ $FAILED -eq 0 ] && exit 0 || exit 1
