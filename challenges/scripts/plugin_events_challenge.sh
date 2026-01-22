#!/bin/bash
# HelixAgent Plugin Events Challenge
# Tests SSE, WebSocket, and event subscription for CLI agent plugins
# 20 tests total

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=20

echo "========================================"
echo "HelixAgent Plugin Events Challenge"
echo "========================================"
echo ""

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"

    echo -n "Testing: $test_name... "
    if eval "$test_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}PASSED${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}FAILED${NC}"
        FAILED=$((FAILED + 1))
    fi
}

# Test 1-5: Event client structure
echo "--- Event Client Structure ---"
run_test "Event client exists" "test -f '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "SSEEventClient class" "grep -q 'class SSEEventClient' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "WebSocketEventClient class" "grep -q 'class WebSocketEventClient' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "EventClient unified class" "grep -q 'class EventClient' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "Factory functions" "grep -q 'createEventClient' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"

# Test 6-10: Task event types
echo ""
echo "--- Task Event Types ---"
run_test "TaskEventType defined" "grep -q 'TaskEventType' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "task.created event" "grep -q 'task.created' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "task.progress event" "grep -q 'task.progress' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "task.completed event" "grep -q 'task.completed' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "task.failed event" "grep -q 'task.failed' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"

# Test 11-15: Debate event types
echo ""
echo "--- Debate Event Types ---"
run_test "DebateEventType defined" "grep -q 'DebateEventType' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "debate.started event" "grep -q 'debate.started' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "debate.round_started event" "grep -q 'debate.round_started' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "debate.consensus event" "grep -q 'debate.consensus' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "debate.completed event" "grep -q 'debate.completed' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"

# Test 16-20: Reconnection and subscription
echo ""
echo "--- Reconnection & Subscription ---"
run_test "Reconnect interval" "grep -q 'reconnectInterval' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "Max reconnect attempts" "grep -q 'maxReconnectAttempts' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "Subscribe method" "grep -q 'subscribe(' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "Unsubscribe method" "grep -q 'unsubscribe(' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"
run_test "Event filtering" "grep -q 'filterTypes' '$PROJECT_ROOT/plugins/packages/events/event_client.ts'"

# Summary
echo ""
echo "========================================"
echo "Events Challenge Results"
echo "========================================"
echo -e "Passed: ${GREEN}$PASSED${NC}/$TOTAL"
echo -e "Failed: ${RED}$FAILED${NC}/$TOTAL"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All event tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some event tests failed${NC}"
    exit 1
fi
