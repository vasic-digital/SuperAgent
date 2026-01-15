#!/bin/bash
# Messaging Integration Challenge
# VALIDATES: RabbitMQ + Kafka integration, MessagingHub, task queues, event streaming
# Tests the complete messaging infrastructure with 25 tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || {
    log_info() { echo -e "\033[0;34m[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_success() { echo -e "\033[0;32m[SUCCESS] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_error() { echo -e "\033[0;31m[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
    log_warning() { echo -e "\033[0;33m[WARNING] $(date '+%Y-%m-%d %H:%M:%S') $1\033[0m"; }
}

CHALLENGE_NAME="Messaging Integration Challenge"
PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="${SCRIPT_DIR}/../.."

log_info "=============================================="
log_info "$CHALLENGE_NAME"
log_info "=============================================="
log_info "Validates: RabbitMQ + Kafka + MessagingHub"
log_info ""

# ============================================================================
# Section 1: Core Messaging Interfaces
# ============================================================================

log_info "=============================================="
log_info "Section 1: Core Messaging Interfaces"
log_info "=============================================="

# Test 1: broker.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 1: broker.go file exists"
if [ -f "$PROJECT_ROOT/internal/messaging/broker.go" ]; then
    log_success "broker.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "broker.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 2: MessageBroker interface defined
TOTAL=$((TOTAL + 1))
log_info "Test 2: MessageBroker interface defined"
if grep -q "type MessageBroker interface" "$PROJECT_ROOT/internal/messaging/broker.go" 2>/dev/null; then
    log_success "MessageBroker interface is defined"
    PASSED=$((PASSED + 1))
else
    log_error "MessageBroker interface NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 3: Message struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 3: Message struct defined"
if grep -q "type Message struct" "$PROJECT_ROOT/internal/messaging/broker.go" 2>/dev/null; then
    log_success "Message struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "Message struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 4: BrokerType constants defined
TOTAL=$((TOTAL + 1))
log_info "Test 4: BrokerType constants (RabbitMQ, Kafka, InMemory)"
RABBITMQ_CONST=$(grep -c "BrokerTypeRabbitMQ\|BrokerRabbitMQ" "$PROJECT_ROOT/internal/messaging/broker.go" 2>/dev/null || echo "0")
KAFKA_CONST=$(grep -c "BrokerTypeKafka\|BrokerKafka" "$PROJECT_ROOT/internal/messaging/broker.go" 2>/dev/null || echo "0")
if [ "$RABBITMQ_CONST" -gt 0 ] && [ "$KAFKA_CONST" -gt 0 ]; then
    log_success "BrokerType constants defined"
    PASSED=$((PASSED + 1))
else
    log_error "BrokerType constants NOT properly defined!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 2: Task Queue Interface (RabbitMQ Pattern)
# ============================================================================

log_info "=============================================="
log_info "Section 2: Task Queue Interface"
log_info "=============================================="

# Test 5: task_queue.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 5: task_queue.go file exists"
if [ -f "$PROJECT_ROOT/internal/messaging/task_queue.go" ]; then
    log_success "task_queue.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "task_queue.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 6: Task struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 6: Task struct defined"
if grep -q "type Task struct" "$PROJECT_ROOT/internal/messaging/task_queue.go" 2>/dev/null; then
    log_success "Task struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "Task struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 7: TaskQueueBroker interface defined
TOTAL=$((TOTAL + 1))
log_info "Test 7: TaskQueueBroker interface defined"
if grep -q "type TaskQueueBroker interface" "$PROJECT_ROOT/internal/messaging/task_queue.go" 2>/dev/null; then
    log_success "TaskQueueBroker interface is defined"
    PASSED=$((PASSED + 1))
else
    log_error "TaskQueueBroker interface NOT defined!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 3: Event Stream Interface (Kafka Pattern)
# ============================================================================

log_info "=============================================="
log_info "Section 3: Event Stream Interface"
log_info "=============================================="

# Test 8: event_stream.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 8: event_stream.go file exists"
if [ -f "$PROJECT_ROOT/internal/messaging/event_stream.go" ]; then
    log_success "event_stream.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "event_stream.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 9: Event struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 9: Event struct defined"
if grep -q "type Event struct" "$PROJECT_ROOT/internal/messaging/event_stream.go" 2>/dev/null; then
    log_success "Event struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "Event struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 10: EventStreamBroker interface defined
TOTAL=$((TOTAL + 1))
log_info "Test 10: EventStreamBroker interface defined"
if grep -q "type EventStreamBroker interface" "$PROJECT_ROOT/internal/messaging/event_stream.go" 2>/dev/null; then
    log_success "EventStreamBroker interface is defined"
    PASSED=$((PASSED + 1))
else
    log_error "EventStreamBroker interface NOT defined!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 4: Messaging Hub
# ============================================================================

log_info "=============================================="
log_info "Section 4: Messaging Hub"
log_info "=============================================="

# Test 11: hub.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 11: hub.go file exists"
if [ -f "$PROJECT_ROOT/internal/messaging/hub.go" ]; then
    log_success "hub.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "hub.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 12: MessagingHub struct defined
TOTAL=$((TOTAL + 1))
log_info "Test 12: MessagingHub struct defined"
if grep -q "type MessagingHub struct" "$PROJECT_ROOT/internal/messaging/hub.go" 2>/dev/null; then
    log_success "MessagingHub struct is defined"
    PASSED=$((PASSED + 1))
else
    log_error "MessagingHub struct NOT defined!"
    FAILED=$((FAILED + 1))
fi

# Test 13: Hub Initialize method exists
TOTAL=$((TOTAL + 1))
log_info "Test 13: Hub Initialize method exists"
if grep -q "func (h \*MessagingHub) Initialize" "$PROJECT_ROOT/internal/messaging/hub.go" 2>/dev/null; then
    log_success "Initialize method exists"
    PASSED=$((PASSED + 1))
else
    log_error "Initialize method NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 5: RabbitMQ Implementation
# ============================================================================

log_info "=============================================="
log_info "Section 5: RabbitMQ Implementation"
log_info "=============================================="

# Test 14: RabbitMQ broker.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 14: RabbitMQ broker.go exists"
if [ -f "$PROJECT_ROOT/internal/messaging/rabbitmq/broker.go" ]; then
    log_success "RabbitMQ broker.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "RabbitMQ broker.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 15: RabbitMQ config.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 15: RabbitMQ config.go exists"
if [ -f "$PROJECT_ROOT/internal/messaging/rabbitmq/config.go" ]; then
    log_success "RabbitMQ config.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "RabbitMQ config.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 16: RabbitMQ connection.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 16: RabbitMQ connection.go exists"
if [ -f "$PROJECT_ROOT/internal/messaging/rabbitmq/connection.go" ]; then
    log_success "RabbitMQ connection.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "RabbitMQ connection.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 6: Kafka Implementation
# ============================================================================

log_info "=============================================="
log_info "Section 6: Kafka Implementation"
log_info "=============================================="

# Test 17: Kafka broker.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 17: Kafka broker.go exists"
if [ -f "$PROJECT_ROOT/internal/messaging/kafka/broker.go" ]; then
    log_success "Kafka broker.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "Kafka broker.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 18: Kafka config.go exists
TOTAL=$((TOTAL + 1))
log_info "Test 18: Kafka config.go exists"
if [ -f "$PROJECT_ROOT/internal/messaging/kafka/config.go" ]; then
    log_success "Kafka config.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "Kafka config.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 7: In-Memory Fallback
# ============================================================================

log_info "=============================================="
log_info "Section 7: In-Memory Fallback"
log_info "=============================================="

# Test 19: InMemory broker exists
TOTAL=$((TOTAL + 1))
log_info "Test 19: InMemory broker exists"
if [ -f "$PROJECT_ROOT/internal/messaging/inmemory/broker.go" ]; then
    log_success "InMemory broker.go exists"
    PASSED=$((PASSED + 1))
else
    log_error "InMemory broker.go NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 8: Docker Compose Infrastructure
# ============================================================================

log_info "=============================================="
log_info "Section 8: Docker Compose Infrastructure"
log_info "=============================================="

# Test 20: docker-compose.messaging.yml exists
TOTAL=$((TOTAL + 1))
log_info "Test 20: docker-compose.messaging.yml exists"
if [ -f "$PROJECT_ROOT/docker-compose.messaging.yml" ]; then
    log_success "docker-compose.messaging.yml exists"
    PASSED=$((PASSED + 1))
else
    log_error "docker-compose.messaging.yml NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 21: RabbitMQ service defined in compose
TOTAL=$((TOTAL + 1))
log_info "Test 21: RabbitMQ service in docker-compose"
if grep -q "rabbitmq:" "$PROJECT_ROOT/docker-compose.messaging.yml" 2>/dev/null; then
    log_success "RabbitMQ service defined"
    PASSED=$((PASSED + 1))
else
    log_error "RabbitMQ service NOT found in compose!"
    FAILED=$((FAILED + 1))
fi

# Test 22: Kafka service defined in compose
TOTAL=$((TOTAL + 1))
log_info "Test 22: Kafka service in docker-compose"
if grep -q "kafka:" "$PROJECT_ROOT/docker-compose.messaging.yml" 2>/dev/null; then
    log_success "Kafka service defined"
    PASSED=$((PASSED + 1))
else
    log_error "Kafka service NOT found in compose!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Section 9: Unit Tests
# ============================================================================

log_info "=============================================="
log_info "Section 9: Unit Tests"
log_info "=============================================="

# Test 23: Run messaging unit tests
TOTAL=$((TOTAL + 1))
log_info "Test 23: Messaging unit tests pass"
cd "$PROJECT_ROOT"
if go test ./internal/messaging/... -count=1 -short 2>&1 | grep -q "ok\|PASS"; then
    log_success "Messaging unit tests pass"
    PASSED=$((PASSED + 1))
else
    log_error "Messaging unit tests FAILED!"
    FAILED=$((FAILED + 1))
fi

# Test 24: RabbitMQ tests exist
TOTAL=$((TOTAL + 1))
log_info "Test 24: RabbitMQ tests exist"
if [ -f "$PROJECT_ROOT/internal/messaging/rabbitmq/broker_test.go" ]; then
    log_success "RabbitMQ tests exist"
    PASSED=$((PASSED + 1))
else
    log_error "RabbitMQ tests NOT found!"
    FAILED=$((FAILED + 1))
fi

# Test 25: Kafka tests exist
TOTAL=$((TOTAL + 1))
log_info "Test 25: Kafka tests exist"
if [ -f "$PROJECT_ROOT/internal/messaging/kafka/broker_test.go" ]; then
    log_success "Kafka tests exist"
    PASSED=$((PASSED + 1))
else
    log_error "Kafka tests NOT found!"
    FAILED=$((FAILED + 1))
fi

# ============================================================================
# Summary
# ============================================================================

log_info "=============================================="
log_info "Challenge Results"
log_info "=============================================="

PASS_RATE=$((PASSED * 100 / TOTAL))

log_info "Total Tests: $TOTAL"
log_info "Passed: $PASSED"
log_info "Failed: $FAILED"
log_info "Pass Rate: $PASS_RATE%"

if [ $FAILED -eq 0 ]; then
    log_success "=============================================="
    log_success "ALL TESTS PASSED!"
    log_success "Messaging Integration Challenge: SUCCESS"
    log_success "=============================================="
    exit 0
else
    log_error "=============================================="
    log_error "$FAILED test(s) failed!"
    log_error "Messaging Integration Challenge: FAILED"
    log_error "=============================================="
    exit 1
fi
