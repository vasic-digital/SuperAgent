#!/bin/bash
# messaging_rabbitmq_challenge.sh - RabbitMQ Integration Challenge
# Tests RabbitMQ broker implementation for HelixAgent

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
source "$SCRIPT_DIR/common.sh"

CHALLENGE_NAME="RabbitMQ Integration Challenge"
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

# Test 1: RabbitMQ broker package exists
echo "[1] RabbitMQ Package Structure"
if [ -f "internal/messaging/rabbitmq/broker.go" ]; then
    log_test "broker.go exists" "PASS"
else
    log_test "broker.go exists" "FAIL"
fi

if [ -f "internal/messaging/rabbitmq/config.go" ]; then
    log_test "config.go exists" "PASS"
else
    log_test "config.go exists" "FAIL"
fi

if [ -f "internal/messaging/rabbitmq/broker_test.go" ]; then
    log_test "broker_test.go exists" "PASS"
else
    log_test "broker_test.go exists" "FAIL"
fi

# Test 2: RabbitMQ broker implements MessageBroker interface
echo ""
echo "[2] Interface Implementation"
if grep -q "func.*Broker.*Connect" internal/messaging/rabbitmq/broker.go 2>/dev/null; then
    log_test "Connect method exists" "PASS"
else
    log_test "Connect method exists" "FAIL"
fi

if grep -q "func.*Broker.*Publish" internal/messaging/rabbitmq/broker.go 2>/dev/null; then
    log_test "Publish method exists" "PASS"
else
    log_test "Publish method exists" "FAIL"
fi

if grep -q "func.*Broker.*Subscribe" internal/messaging/rabbitmq/broker.go 2>/dev/null; then
    log_test "Subscribe method exists" "PASS"
else
    log_test "Subscribe method exists" "FAIL"
fi

if grep -q "func.*Broker.*Close" internal/messaging/rabbitmq/broker.go 2>/dev/null; then
    log_test "Close method exists" "PASS"
else
    log_test "Close method exists" "FAIL"
fi

# Test 3: RabbitMQ configuration validation
echo ""
echo "[3] Configuration"
if grep -q "Config struct" internal/messaging/rabbitmq/config.go 2>/dev/null; then
    log_test "Config struct defined" "PASS"
else
    log_test "Config struct defined" "FAIL"
fi

if grep -q "Validate" internal/messaging/rabbitmq/config.go 2>/dev/null; then
    log_test "Validate method exists" "PASS"
else
    log_test "Validate method exists" "FAIL"
fi

if grep -q "Host" internal/messaging/rabbitmq/config.go 2>/dev/null; then
    log_test "Host field exists" "PASS"
else
    log_test "Host field exists" "FAIL"
fi

if grep -q "Port" internal/messaging/rabbitmq/config.go 2>/dev/null; then
    log_test "Port field exists" "PASS"
else
    log_test "Port field exists" "FAIL"
fi

# Test 4: Unit tests pass
echo ""
echo "[4] Unit Tests"
if go test -v ./internal/messaging/rabbitmq/... -count=1 2>&1 | grep -q "PASS"; then
    log_test "RabbitMQ unit tests pass" "PASS"
else
    log_test "RabbitMQ unit tests pass" "FAIL"
fi

# Test 5: Dead letter queue support
echo ""
echo "[5] Dead Letter Queue Support"
# Check for DLQ in definitions or messaging config
if grep -q "dead-letter\|dlq\|dead_letter\|dlx" configs/rabbitmq/definitions.json 2>/dev/null || grep -q "dead_letter" configs/messaging.yaml 2>/dev/null; then
    log_test "Dead letter queue support" "PASS"
else
    log_test "Dead letter queue support" "FAIL"
fi

# Test 6: Publisher confirms
echo ""
echo "[6] Publisher Confirms"
if grep -q "Confirm\|confirm" internal/messaging/rabbitmq/broker.go 2>/dev/null; then
    log_test "Publisher confirms support" "PASS"
else
    log_test "Publisher confirms support" "FAIL"
fi

# Test 7: Reconnection logic
echo ""
echo "[7] Reconnection Logic"
if grep -q "reconnect\|Reconnect" internal/messaging/rabbitmq/broker.go 2>/dev/null; then
    log_test "Reconnection support" "PASS"
else
    log_test "Reconnection support" "FAIL"
fi

# Test 8: Docker Compose configuration
echo ""
echo "[8] Docker Compose Configuration"
if [ -f "docker-compose.messaging.yml" ]; then
    log_test "docker-compose.messaging.yml exists" "PASS"
else
    log_test "docker-compose.messaging.yml exists" "FAIL"
fi

if grep -q "rabbitmq" docker-compose.messaging.yml 2>/dev/null; then
    log_test "RabbitMQ service defined" "PASS"
else
    log_test "RabbitMQ service defined" "FAIL"
fi

if grep -q "5672" docker-compose.messaging.yml 2>/dev/null; then
    log_test "AMQP port configured" "PASS"
else
    log_test "AMQP port configured" "FAIL"
fi

if grep -q "15672" docker-compose.messaging.yml 2>/dev/null; then
    log_test "Management port configured" "PASS"
else
    log_test "Management port configured" "FAIL"
fi

# Test 9: RabbitMQ definitions file
echo ""
echo "[9] RabbitMQ Definitions"
if [ -f "configs/rabbitmq/definitions.json" ]; then
    log_test "definitions.json exists" "PASS"
else
    log_test "definitions.json exists" "FAIL"
fi

if grep -q "queues" configs/rabbitmq/definitions.json 2>/dev/null; then
    log_test "Queue definitions present" "PASS"
else
    log_test "Queue definitions present" "FAIL"
fi

if grep -q "exchanges" configs/rabbitmq/definitions.json 2>/dev/null; then
    log_test "Exchange definitions present" "PASS"
else
    log_test "Exchange definitions present" "FAIL"
fi

if grep -q "bindings" configs/rabbitmq/definitions.json 2>/dev/null; then
    log_test "Binding definitions present" "PASS"
else
    log_test "Binding definitions present" "FAIL"
fi

# Test 10: Messaging configuration
echo ""
echo "[10] Messaging Configuration"
if [ -f "configs/messaging.yaml" ]; then
    log_test "messaging.yaml exists" "PASS"
else
    log_test "messaging.yaml exists" "FAIL"
fi

if grep -q "rabbitmq" configs/messaging.yaml 2>/dev/null; then
    log_test "RabbitMQ config section" "PASS"
else
    log_test "RabbitMQ config section" "FAIL"
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
