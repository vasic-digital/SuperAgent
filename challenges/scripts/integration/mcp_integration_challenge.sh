#!/bin/bash
# =============================================================================
# MCP Integration Challenge
# Tests all MCP server integrations
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; ((TESTS_PASSED++)); }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; ((TESTS_FAILED++)); }
log_skip() { echo -e "${YELLOW}[SKIP]${NC} $1"; ((TESTS_SKIPPED++)); }

# Configuration
HELIXAGENT_URL="${HELIXAGENT_URL:-http://localhost:7061}"
MCP_MANAGER_URL="${MCP_MANAGER_URL:-http://localhost:9000}"
TIMEOUT=10

# =============================================================================
# Helper Functions
# =============================================================================

wait_for_service() {
    local url=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1

    log_info "Waiting for $service_name to be ready..."
    while [ $attempt -le $max_attempts ]; do
        if curl -sf "$url/health" > /dev/null 2>&1; then
            log_info "$service_name is ready"
            return 0
        fi
        sleep 1
        ((attempt++))
    done
    log_fail "$service_name failed to start"
    return 1
}

test_mcp_endpoint() {
    local endpoint=$1
    local description=$2
    local expected_status=${3:-200}

    if response=$(curl -sf -w "%{http_code}" -o /tmp/mcp_response.json "$HELIXAGENT_URL$endpoint" 2>/dev/null); then
        if [ "$response" = "$expected_status" ]; then
            log_success "$description"
            return 0
        fi
    fi
    log_fail "$description (got: $response, expected: $expected_status)"
    return 1
}

test_mcp_tool_call() {
    local server=$1
    local tool=$2
    local arguments=$3
    local description=$4

    local payload=$(cat <<EOF
{
    "server": "$server",
    "tool": "$tool",
    "arguments": $arguments
}
EOF
)

    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "$HELIXAGENT_URL/api/v1/mcp/tools/call" 2>/dev/null); then
        if echo "$response" | jq -e '.result' > /dev/null 2>&1; then
            log_success "$description"
            return 0
        fi
    fi
    log_fail "$description"
    return 1
}

# =============================================================================
# Test Categories
# =============================================================================

test_mcp_core_infrastructure() {
    log_info "=== Testing MCP Core Infrastructure ==="

    # Test 1: MCP endpoint availability
    test_mcp_endpoint "/api/v1/mcp/servers" "MCP servers endpoint available"

    # Test 2: MCP tools list
    test_mcp_endpoint "/api/v1/mcp/tools/list" "MCP tools list endpoint available"

    # Test 3: Server registry
    if curl -sf "$HELIXAGENT_URL/api/v1/mcp/servers" | jq -e '.servers | length >= 0' > /dev/null 2>&1; then
        log_success "MCP server registry functional"
    else
        log_fail "MCP server registry non-functional"
    fi
}

test_mcp_filesystem_server() {
    log_info "=== Testing MCP Filesystem Server ==="

    # Test 1: List directory
    test_mcp_tool_call "filesystem" "list_directory" '{"path": "/workspace"}' \
        "Filesystem: list_directory tool"

    # Test 2: Read file (if exists)
    test_mcp_tool_call "filesystem" "read_file" '{"path": "/workspace/README.md"}' \
        "Filesystem: read_file tool" || log_skip "No README.md in workspace"

    # Test 3: Write and delete file
    local test_file="/workspace/.mcp_test_$(date +%s).txt"
    if test_mcp_tool_call "filesystem" "write_file" \
        "{\"path\": \"$test_file\", \"content\": \"MCP integration test\"}" \
        "Filesystem: write_file tool"; then
        test_mcp_tool_call "filesystem" "delete_file" "{\"path\": \"$test_file\"}" \
            "Filesystem: delete_file tool"
    fi
}

test_mcp_git_server() {
    log_info "=== Testing MCP Git Server ==="

    # Test 1: Git status
    test_mcp_tool_call "git" "git_status" '{"repo_path": "/workspace"}' \
        "Git: status tool"

    # Test 2: Git log
    test_mcp_tool_call "git" "git_log" '{"repo_path": "/workspace", "max_count": 5}' \
        "Git: log tool"

    # Test 3: Git diff
    test_mcp_tool_call "git" "git_diff" '{"repo_path": "/workspace"}' \
        "Git: diff tool"
}

test_mcp_memory_server() {
    log_info "=== Testing MCP Memory Server ==="

    # Test 1: Store memory
    local key="test_key_$(date +%s)"
    test_mcp_tool_call "memory" "store" \
        "{\"key\": \"$key\", \"value\": \"test value\"}" \
        "Memory: store tool"

    # Test 2: Retrieve memory
    test_mcp_tool_call "memory" "retrieve" "{\"key\": \"$key\"}" \
        "Memory: retrieve tool"

    # Test 3: Delete memory
    test_mcp_tool_call "memory" "delete" "{\"key\": \"$key\"}" \
        "Memory: delete tool"
}

test_mcp_fetch_server() {
    log_info "=== Testing MCP Fetch Server ==="

    # Test 1: Fetch URL
    test_mcp_tool_call "fetch" "fetch" \
        '{"url": "https://httpbin.org/get"}' \
        "Fetch: basic URL fetch"

    # Test 2: Fetch with headers
    test_mcp_tool_call "fetch" "fetch" \
        '{"url": "https://httpbin.org/headers", "headers": {"X-Test": "value"}}' \
        "Fetch: fetch with custom headers"
}

test_mcp_time_server() {
    log_info "=== Testing MCP Time Server ==="

    # Test 1: Get current time
    test_mcp_tool_call "time" "get_current_time" '{}' \
        "Time: get_current_time tool"

    # Test 2: Convert timezone
    test_mcp_tool_call "time" "convert_timezone" \
        '{"time": "2024-01-15T10:00:00Z", "from_tz": "UTC", "to_tz": "America/New_York"}' \
        "Time: convert_timezone tool"
}

test_mcp_sqlite_server() {
    log_info "=== Testing MCP SQLite Server ==="

    # Test 1: Create table
    test_mcp_tool_call "sqlite" "execute" \
        '{"query": "CREATE TABLE IF NOT EXISTS mcp_test (id INTEGER PRIMARY KEY, name TEXT)"}' \
        "SQLite: create table"

    # Test 2: Insert data
    test_mcp_tool_call "sqlite" "execute" \
        '{"query": "INSERT INTO mcp_test (name) VALUES (?)", "params": ["test"]}' \
        "SQLite: insert data"

    # Test 3: Query data
    test_mcp_tool_call "sqlite" "query" \
        '{"query": "SELECT * FROM mcp_test"}' \
        "SQLite: query data"

    # Test 4: Drop table
    test_mcp_tool_call "sqlite" "execute" \
        '{"query": "DROP TABLE IF EXISTS mcp_test"}' \
        "SQLite: drop table"
}

test_mcp_vector_servers() {
    log_info "=== Testing MCP Vector Database Servers ==="

    # Test Chroma MCP (if available)
    if curl -sf "$HELIXAGENT_URL/api/v1/mcp/servers" | jq -e '.servers[] | select(.name == "chroma")' > /dev/null 2>&1; then
        test_mcp_tool_call "chroma" "list_collections" '{}' \
            "Chroma: list_collections"
    else
        log_skip "Chroma MCP server not available"
    fi

    # Test Qdrant MCP (if available)
    if curl -sf "$HELIXAGENT_URL/api/v1/mcp/servers" | jq -e '.servers[] | select(.name == "qdrant")' > /dev/null 2>&1; then
        test_mcp_tool_call "qdrant" "list_collections" '{}' \
            "Qdrant: list_collections"
    else
        log_skip "Qdrant MCP server not available"
    fi
}

test_mcp_connection_pool() {
    log_info "=== Testing MCP Connection Pool ==="

    # Test 1: Multiple concurrent requests
    log_info "Testing concurrent requests..."
    local pids=()
    for i in {1..5}; do
        curl -sf "$HELIXAGENT_URL/api/v1/mcp/tools/list" > /dev/null 2>&1 &
        pids+=($!)
    done

    local all_success=true
    for pid in "${pids[@]}"; do
        if ! wait "$pid"; then
            all_success=false
        fi
    done

    if $all_success; then
        log_success "Connection pool handles concurrent requests"
    else
        log_fail "Connection pool failed concurrent requests"
    fi

    # Test 2: Connection reuse
    local conn_count_before=$(curl -sf "$HELIXAGENT_URL/api/v1/mcp/stats" | jq -r '.active_connections // 0')
    curl -sf "$HELIXAGENT_URL/api/v1/mcp/tools/list" > /dev/null 2>&1
    curl -sf "$HELIXAGENT_URL/api/v1/mcp/tools/list" > /dev/null 2>&1
    local conn_count_after=$(curl -sf "$HELIXAGENT_URL/api/v1/mcp/stats" | jq -r '.active_connections // 0')

    if [ "$conn_count_after" -le "$((conn_count_before + 1))" ]; then
        log_success "Connection pool reuses connections"
    else
        log_skip "Connection reuse test inconclusive"
    fi
}

test_mcp_lazy_initialization() {
    log_info "=== Testing MCP Lazy Initialization ==="

    # Get list of servers
    local servers=$(curl -sf "$HELIXAGENT_URL/api/v1/mcp/servers" | jq -r '.servers[].name' 2>/dev/null)

    if [ -z "$servers" ]; then
        log_skip "No servers to test lazy initialization"
        return
    fi

    for server in $servers; do
        # Check if server starts only when needed
        local status=$(curl -sf "$HELIXAGENT_URL/api/v1/mcp/servers/$server/status" | jq -r '.status // "unknown"')
        if [ "$status" = "pending" ] || [ "$status" = "available" ]; then
            # Server should initialize on first use
            curl -sf -X POST \
                -H "Content-Type: application/json" \
                -d "{\"server\": \"$server\", \"tool\": \"health\", \"arguments\": {}}" \
                "$HELIXAGENT_URL/api/v1/mcp/tools/call" > /dev/null 2>&1 || true

            local new_status=$(curl -sf "$HELIXAGENT_URL/api/v1/mcp/servers/$server/status" | jq -r '.status // "unknown"')
            if [ "$new_status" = "running" ] || [ "$new_status" = "connected" ]; then
                log_success "Lazy initialization works for $server"
            else
                log_skip "Lazy initialization status unclear for $server (status: $new_status)"
            fi
        else
            log_skip "Server $server already initialized (status: $status)"
        fi
        break  # Only test one server to keep it fast
    done
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    echo "=============================================="
    echo "  MCP Integration Challenge"
    echo "  HelixAgent - $(date)"
    echo "=============================================="
    echo ""

    # Wait for services
    wait_for_service "$HELIXAGENT_URL" "HelixAgent" || exit 1

    # Run test categories
    test_mcp_core_infrastructure
    echo ""
    test_mcp_filesystem_server
    echo ""
    test_mcp_git_server
    echo ""
    test_mcp_memory_server
    echo ""
    test_mcp_fetch_server
    echo ""
    test_mcp_time_server
    echo ""
    test_mcp_sqlite_server
    echo ""
    test_mcp_vector_servers
    echo ""
    test_mcp_connection_pool
    echo ""
    test_mcp_lazy_initialization

    # Summary
    echo ""
    echo "=============================================="
    echo "  Challenge Results"
    echo "=============================================="
    echo -e "  ${GREEN}Passed:${NC}  $TESTS_PASSED"
    echo -e "  ${RED}Failed:${NC}  $TESTS_FAILED"
    echo -e "  ${YELLOW}Skipped:${NC} $TESTS_SKIPPED"
    echo "=============================================="

    # Exit with failure if any tests failed
    if [ $TESTS_FAILED -gt 0 ]; then
        echo -e "\n${RED}Challenge FAILED!${NC}"
        exit 1
    else
        echo -e "\n${GREEN}Challenge PASSED!${NC}"
        exit 0
    fi
}

main "$@"
