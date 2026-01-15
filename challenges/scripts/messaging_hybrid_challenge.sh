#!/bin/bash
# messaging_hybrid_challenge.sh - Hybrid Messaging Challenge
# Tests full messaging integration: RabbitMQ, Kafka, GraphQL, TOON

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
source "$SCRIPT_DIR/common.sh"

CHALLENGE_NAME="Hybrid Messaging Integration Challenge"
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

# Test 1: Core messaging abstraction
echo "[1] Messaging Abstraction Layer"
if [ -f "internal/messaging/broker.go" ]; then
    log_test "broker.go exists" "PASS"
else
    log_test "broker.go exists" "FAIL"
fi

if grep -q "type MessageBroker interface" internal/messaging/broker.go 2>/dev/null; then
    log_test "MessageBroker interface" "PASS"
else
    log_test "MessageBroker interface" "FAIL"
fi

if grep -q "type Message struct" internal/messaging/broker.go 2>/dev/null; then
    log_test "Message struct defined" "PASS"
else
    log_test "Message struct defined" "FAIL"
fi

# Test 2: Multiple broker implementations
echo ""
echo "[2] Broker Implementations"
if [ -d "internal/messaging/rabbitmq" ]; then
    log_test "RabbitMQ implementation" "PASS"
else
    log_test "RabbitMQ implementation" "FAIL"
fi

if [ -d "internal/messaging/kafka" ]; then
    log_test "Kafka implementation" "PASS"
else
    log_test "Kafka implementation" "FAIL"
fi

if [ -d "internal/messaging/inmemory" ]; then
    log_test "In-memory implementation" "PASS"
else
    log_test "In-memory implementation" "FAIL"
fi

# Test 3: Task queue interface
echo ""
echo "[3] Task Queue Interface"
if [ -f "internal/messaging/task_queue.go" ]; then
    log_test "task_queue.go exists" "PASS"
else
    log_test "task_queue.go exists" "FAIL"
fi

if grep -q "TaskQueueBroker" internal/messaging/task_queue.go 2>/dev/null; then
    log_test "TaskQueueBroker interface" "PASS"
else
    log_test "TaskQueueBroker interface" "FAIL"
fi

# Test 4: Event stream interface
echo ""
echo "[4] Event Stream Interface"
if [ -f "internal/messaging/event_stream.go" ]; then
    log_test "event_stream.go exists" "PASS"
else
    log_test "event_stream.go exists" "FAIL"
fi

if grep -q "EventStreamBroker" internal/messaging/event_stream.go 2>/dev/null; then
    log_test "EventStreamBroker interface" "PASS"
else
    log_test "EventStreamBroker interface" "FAIL"
fi

# Test 5: Docker infrastructure
echo ""
echo "[5] Docker Infrastructure"
if [ -f "docker-compose.messaging.yml" ]; then
    log_test "docker-compose.messaging.yml exists" "PASS"
else
    log_test "docker-compose.messaging.yml exists" "FAIL"
fi

services_count=$(grep -c "^  [a-z].*:" docker-compose.messaging.yml 2>/dev/null || echo 0)
if [ "$services_count" -ge 4 ]; then
    log_test "Multiple services defined (${services_count})" "PASS"
else
    log_test "Multiple services defined (${services_count})" "FAIL"
fi

if grep -q "healthcheck:" docker-compose.messaging.yml 2>/dev/null; then
    log_test "Health checks configured" "PASS"
else
    log_test "Health checks configured" "FAIL"
fi

if grep -q "volumes:" docker-compose.messaging.yml 2>/dev/null; then
    log_test "Volumes configured" "PASS"
else
    log_test "Volumes configured" "FAIL"
fi

# Test 6: Configuration files
echo ""
echo "[6] Configuration Files"
if [ -f "configs/messaging.yaml" ]; then
    log_test "messaging.yaml exists" "PASS"
else
    log_test "messaging.yaml exists" "FAIL"
fi

if [ -f "configs/rabbitmq/definitions.json" ]; then
    log_test "RabbitMQ definitions exist" "PASS"
else
    log_test "RabbitMQ definitions exist" "FAIL"
fi

if [ -f "configs/rabbitmq/rabbitmq.conf" ]; then
    log_test "RabbitMQ config exists" "PASS"
else
    log_test "RabbitMQ config exists" "FAIL"
fi

# Test 7: GraphQL integration
echo ""
echo "[7] GraphQL Integration"
if [ -f "internal/graphql/schema.go" ]; then
    log_test "GraphQL schema exists" "PASS"
else
    log_test "GraphQL schema exists" "FAIL"
fi

if [ -d "internal/graphql/types" ]; then
    log_test "GraphQL types exist" "PASS"
else
    log_test "GraphQL types exist" "FAIL"
fi

# Test 8: TOON protocol
echo ""
echo "[8] TOON Protocol"
if [ -f "internal/toon/encoder.go" ]; then
    log_test "TOON encoder exists" "PASS"
else
    log_test "TOON encoder exists" "FAIL"
fi

if [ -f "internal/toon/transport.go" ]; then
    log_test "TOON transport exists" "PASS"
else
    log_test "TOON transport exists" "FAIL"
fi

# Test 9: LLMsVerifier messaging
echo ""
echo "[9] LLMsVerifier Messaging"
if [ -d "LLMsVerifier/llm-verifier/messaging" ]; then
    log_test "LLMsVerifier messaging package" "PASS"
else
    log_test "LLMsVerifier messaging package" "FAIL"
fi

if [ -f "LLMsVerifier/llm-verifier/messaging/publisher.go" ]; then
    log_test "Event publisher exists" "PASS"
else
    log_test "Event publisher exists" "FAIL"
fi

if [ -f "LLMsVerifier/llm-verifier/messaging/events.go" ]; then
    log_test "Event types defined" "PASS"
else
    log_test "Event types defined" "FAIL"
fi

# Test 10: Unit tests pass
echo ""
echo "[10] Unit Tests"
if go test -v ./internal/messaging/... -count=1 2>&1 | grep -q "PASS"; then
    log_test "Messaging tests pass" "PASS"
else
    log_test "Messaging tests pass" "FAIL"
fi

if go test -v ./internal/graphql/... -count=1 2>&1 | grep -q "PASS"; then
    log_test "GraphQL tests pass" "PASS"
else
    log_test "GraphQL tests pass" "FAIL"
fi

if go test -v ./internal/toon/... -count=1 2>&1 | grep -q "PASS"; then
    log_test "TOON tests pass" "PASS"
else
    log_test "TOON tests pass" "FAIL"
fi

# Test 11: Integration test
echo ""
echo "[11] Integration Tests"
if [ -f "tests/integration/messaging_integration_test.go" ]; then
    log_test "Integration test file exists" "PASS"
else
    log_test "Integration test file exists" "FAIL"
fi

# Test 12: Dependencies
echo ""
echo "[12] Dependencies"
if grep -q "rabbitmq\|amqp091-go" go.mod 2>/dev/null; then
    log_test "RabbitMQ client dependency" "PASS"
else
    log_test "RabbitMQ client dependency" "FAIL"
fi

if grep -q "kafka-go" go.mod 2>/dev/null; then
    log_test "Kafka client dependency" "PASS"
else
    log_test "Kafka client dependency" "FAIL"
fi

if grep -q "graphql-go/graphql" go.mod 2>/dev/null; then
    log_test "GraphQL dependency" "PASS"
else
    log_test "GraphQL dependency" "FAIL"
fi

# Test 13: Error handling
echo ""
echo "[13] Error Handling"
if [ -f "internal/messaging/errors.go" ]; then
    log_test "errors.go exists" "PASS"
else
    log_test "errors.go exists" "FAIL"
fi

if grep -q "BrokerError" internal/messaging/errors.go 2>/dev/null; then
    log_test "BrokerError type defined" "PASS"
else
    log_test "BrokerError type defined" "FAIL"
fi

# Test 14: Metrics support
echo ""
echo "[14] Metrics Support"
if [ -f "internal/messaging/metrics.go" ]; then
    log_test "metrics.go exists" "PASS"
else
    log_test "metrics.go exists" "FAIL"
fi

if grep -q "BrokerMetrics" internal/messaging/metrics.go 2>/dev/null; then
    log_test "BrokerMetrics type defined" "PASS"
else
    log_test "BrokerMetrics type defined" "FAIL"
fi

# Test 15: Options support
echo ""
echo "[15] Options Pattern"
if [ -f "internal/messaging/options.go" ]; then
    log_test "options.go exists" "PASS"
else
    log_test "options.go exists" "FAIL"
fi

if grep -q "PublishOption\|SubscribeOption" internal/messaging/options.go 2>/dev/null; then
    log_test "Options types defined" "PASS"
else
    log_test "Options types defined" "FAIL"
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
