#!/bin/bash
# messaging_migration_challenge.sh - Migration Challenge
# Tests migration capabilities and backward compatibility

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
source "$SCRIPT_DIR/common.sh"

CHALLENGE_NAME="Messaging Migration Challenge"
PASSED=0
FAILED=0
TOTAL=0

log_test() {
    local test_name="$1"
    local status="$2"
    TOTAL=$((TOTAL + 1))
    if [ "$status" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        echo -e "  \e[32m✓\e[0m $test_name"
    else
        FAILED=$((FAILED + 1))
        echo -e "  \e[31m✗\e[0m $test_name"
    fi
}

echo "=============================================="
echo "  $CHALLENGE_NAME"
echo "=============================================="
echo ""

cd "$PROJECT_ROOT"

# Test 1: Fallback broker exists
echo "[1] Fallback Mechanism"
if [ -d "internal/messaging/inmemory" ]; then
    log_test "In-memory fallback broker exists" "PASS"
else
    log_test "In-memory fallback broker exists" "FAIL"
fi

if grep -q "BrokerTypeInMemory" internal/messaging/broker.go 2>/dev/null; then
    log_test "In-memory broker type defined" "PASS"
else
    log_test "In-memory broker type defined" "FAIL"
fi

# Test 2: Multiple broker implementations
echo ""
echo "[2] Multiple Broker Implementations"
broker_count=0
for dir in internal/messaging/*/; do
    if [ -f "${dir}broker.go" ]; then
        broker_count=$((broker_count + 1))
    fi
done

if [ "$broker_count" -ge 2 ]; then
    log_test "Multiple broker implementations (${broker_count})" "PASS"
else
    log_test "Multiple broker implementations (${broker_count})" "FAIL"
fi

# Test 3: Message broker interface
echo ""
echo "[3] Unified Broker Interface"
if grep -q "type MessageBroker interface" internal/messaging/broker.go 2>/dev/null; then
    log_test "MessageBroker interface defined" "PASS"
else
    log_test "MessageBroker interface defined" "FAIL"
fi

if grep -q "BrokerType()" internal/messaging/broker.go 2>/dev/null; then
    log_test "BrokerType method in interface" "PASS"
else
    log_test "BrokerType method in interface" "FAIL"
fi

# Test 4: Configuration validation
echo ""
echo "[4] Configuration Validation"
if grep -q "func.*Validate" internal/messaging/broker.go 2>/dev/null; then
    log_test "Config validation method" "PASS"
else
    log_test "Config validation method" "FAIL"
fi

if grep -q "func.*Validate" internal/messaging/rabbitmq/config.go 2>/dev/null; then
    log_test "RabbitMQ config validation" "PASS"
else
    log_test "RabbitMQ config validation" "FAIL"
fi

if grep -q "func.*Validate" internal/messaging/kafka/config.go 2>/dev/null; then
    log_test "Kafka config validation" "PASS"
else
    log_test "Kafka config validation" "FAIL"
fi

# Test 5: Error handling
echo ""
echo "[5] Error Handling"
if [ -f "internal/messaging/errors.go" ]; then
    log_test "Errors file exists" "PASS"
else
    log_test "Errors file exists" "FAIL"
fi

if grep -q "IsRetryable" internal/messaging/errors.go 2>/dev/null; then
    log_test "Retryable error support" "PASS"
else
    log_test "Retryable error support" "FAIL"
fi

# Test 6: Metrics support
echo ""
echo "[6] Metrics Support"
if [ -f "internal/messaging/metrics.go" ]; then
    log_test "Metrics file exists" "PASS"
else
    log_test "Metrics file exists" "FAIL"
fi

if grep -q "BrokerMetrics" internal/messaging/metrics.go 2>/dev/null; then
    log_test "BrokerMetrics type defined" "PASS"
else
    log_test "BrokerMetrics type defined" "FAIL"
fi

# Test 7: CLI agents still work
echo ""
echo "[7] Backward Compatibility - CLI Agents"
if [ -f "internal/agents/registry.go" ]; then
    log_test "Agent registry exists" "PASS"
else
    log_test "Agent registry exists" "FAIL"
fi

# Check that agent registry is not broken
if grep -q "OpenCode\|Crush\|HelixCode\|Kiro" internal/agents/registry.go 2>/dev/null; then
    log_test "CLI agents still registered" "PASS"
else
    log_test "CLI agents still registered" "FAIL"
fi

# Test 8: Background tasks compatibility
echo ""
echo "[8] Backward Compatibility - Background Tasks"
if [ -f "internal/background/task_queue.go" ]; then
    log_test "Background task queue exists" "PASS"
else
    log_test "Background task queue exists" "FAIL"
fi

if [ -f "internal/background/worker_pool.go" ]; then
    log_test "Worker pool exists" "PASS"
else
    log_test "Worker pool exists" "FAIL"
fi

# Test 9: Zero-downtime capability
echo ""
echo "[9] Zero-Downtime Migration Support"
# Check for connection recovery/reconnection support
if grep -q "reconnect\|Reconnect" internal/messaging/rabbitmq/broker.go 2>/dev/null; then
    log_test "RabbitMQ reconnection support" "PASS"
else
    log_test "RabbitMQ reconnection support" "FAIL"
fi

# Check for health checks
if grep -q "HealthCheck" internal/messaging/broker.go 2>/dev/null; then
    log_test "Health check method in interface" "PASS"
else
    log_test "Health check method in interface" "FAIL"
fi

# Test 10: All existing tests still pass
echo ""
echo "[10] Existing Tests Compatibility"
if go test ./internal/messaging/... -count=1 2>&1 | grep -q "^ok"; then
    log_test "Messaging tests pass" "PASS"
else
    log_test "Messaging tests pass" "FAIL"
fi

if go test ./internal/toon/... -count=1 2>&1 | grep -q "^ok"; then
    log_test "TOON tests pass" "PASS"
else
    log_test "TOON tests pass" "FAIL"
fi

if go test ./internal/graphql/... -count=1 2>&1 | grep -q "^ok"; then
    log_test "GraphQL tests pass" "PASS"
else
    log_test "GraphQL tests pass" "FAIL"
fi

echo ""
echo "=============================================="
echo "  Results: $PASSED/$TOTAL tests passed"
echo "=============================================="

if [ $FAILED -gt 0 ]; then
    echo -e "\e[31m$FAILED test(s) failed\e[0m"
    exit 1
else
    echo -e "\e[32mAll tests passed!\e[0m"
    exit 0
fi
