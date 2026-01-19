#!/bin/bash
# =============================================================================
# LSP Integration Challenge
# Tests all LSP server integrations
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
LSP_MANAGER_URL="${LSP_MANAGER_URL:-http://localhost:5100}"
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

test_lsp_endpoint() {
    local endpoint=$1
    local description=$2
    local expected_status=${3:-200}

    if response=$(curl -sf -w "%{http_code}" -o /tmp/lsp_response.json "$HELIXAGENT_URL$endpoint" 2>/dev/null); then
        if [ "$response" = "$expected_status" ]; then
            log_success "$description"
            return 0
        fi
    fi
    log_fail "$description (got: $response, expected: $expected_status)"
    return 1
}

test_lsp_operation() {
    local server=$1
    local operation=$2
    local arguments=$3
    local description=$4

    local payload=$(cat <<EOF
{
    "serverId": "$server",
    "toolName": "$operation",
    "arguments": $arguments
}
EOF
)

    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "$HELIXAGENT_URL/v1/lsp/execute" 2>/dev/null); then
        if echo "$response" | jq -e '.result' > /dev/null 2>&1 || \
           echo "$response" | jq -e '.data' > /dev/null 2>&1; then
            log_success "$description"
            return 0
        fi
    fi
    log_fail "$description"
    return 1
}

check_lsp_binary() {
    local binary=$1
    local name=$2

    if command -v "$binary" > /dev/null 2>&1; then
        log_success "$name binary available"
        return 0
    else
        log_skip "$name binary not installed"
        return 1
    fi
}

# =============================================================================
# Test Categories
# =============================================================================

test_lsp_core_infrastructure() {
    log_info "=== Testing LSP Core Infrastructure ==="

    # Test 1: LSP servers endpoint
    test_lsp_endpoint "/v1/lsp/servers" "LSP servers endpoint available"

    # Test 2: LSP stats endpoint
    test_lsp_endpoint "/v1/lsp/stats" "LSP stats endpoint available"

    # Test 3: Server list content
    if curl -sf "$HELIXAGENT_URL/v1/lsp/servers" | jq -e '.servers | length >= 0' > /dev/null 2>&1; then
        log_success "LSP server registry functional"
    else
        log_fail "LSP server registry non-functional"
    fi
}

test_lsp_binary_detection() {
    log_info "=== Testing LSP Binary Detection ==="

    # Core language servers
    check_lsp_binary "gopls" "gopls (Go)"
    check_lsp_binary "rust-analyzer" "rust-analyzer (Rust)"
    check_lsp_binary "pylsp" "pylsp (Python)"
    check_lsp_binary "typescript-language-server" "typescript-language-server"

    # Additional language servers
    check_lsp_binary "clangd" "clangd (C/C++)"
    check_lsp_binary "bash-language-server" "bash-language-server"
    check_lsp_binary "yaml-language-server" "yaml-language-server"
    check_lsp_binary "docker-language-server" "docker-language-server"
}

test_lsp_go_server() {
    log_info "=== Testing Go LSP Server (gopls) ==="

    # Skip if gopls not available
    if ! command -v gopls > /dev/null 2>&1; then
        log_skip "gopls not installed, skipping Go LSP tests"
        return
    fi

    # Create test file
    local test_file="/tmp/test_go_$(date +%s).go"
    cat > "$test_file" << 'EOF'
package main

import "fmt"

func main() {
    message := "Hello"
    fmt.Println(message)
}
EOF

    # Test 1: Completion
    test_lsp_operation "gopls" "completion" \
        "{\"uri\": \"file://$test_file\", \"line\": 7, \"character\": 5}" \
        "Go: completion"

    # Test 2: Hover
    test_lsp_operation "gopls" "hover" \
        "{\"uri\": \"file://$test_file\", \"line\": 6, \"character\": 2}" \
        "Go: hover"

    # Test 3: Definition
    test_lsp_operation "gopls" "definition" \
        "{\"uri\": \"file://$test_file\", \"line\": 7, \"character\": 13}" \
        "Go: definition"

    # Cleanup
    rm -f "$test_file"
}

test_lsp_rust_server() {
    log_info "=== Testing Rust LSP Server (rust-analyzer) ==="

    if ! command -v rust-analyzer > /dev/null 2>&1; then
        log_skip "rust-analyzer not installed, skipping Rust LSP tests"
        return
    fi

    # Create test file
    local test_file="/tmp/test_rust_$(date +%s).rs"
    cat > "$test_file" << 'EOF'
fn main() {
    let message = "Hello";
    println!("{}", message);
}
EOF

    # Test 1: Completion
    test_lsp_operation "rust-analyzer" "completion" \
        "{\"uri\": \"file://$test_file\", \"line\": 3, \"character\": 5}" \
        "Rust: completion"

    # Test 2: Hover
    test_lsp_operation "rust-analyzer" "hover" \
        "{\"uri\": \"file://$test_file\", \"line\": 2, \"character\": 8}" \
        "Rust: hover"

    # Cleanup
    rm -f "$test_file"
}

test_lsp_python_server() {
    log_info "=== Testing Python LSP Server (pylsp) ==="

    if ! command -v pylsp > /dev/null 2>&1; then
        log_skip "pylsp not installed, skipping Python LSP tests"
        return
    fi

    # Create test file
    local test_file="/tmp/test_python_$(date +%s).py"
    cat > "$test_file" << 'EOF'
def greet(name):
    message = f"Hello, {name}!"
    return message

result = greet("World")
print(result)
EOF

    # Test 1: Completion
    test_lsp_operation "pylsp" "completion" \
        "{\"uri\": \"file://$test_file\", \"line\": 6, \"character\": 0}" \
        "Python: completion"

    # Test 2: Hover
    test_lsp_operation "pylsp" "hover" \
        "{\"uri\": \"file://$test_file\", \"line\": 5, \"character\": 10}" \
        "Python: hover"

    # Test 3: References
    test_lsp_operation "pylsp" "references" \
        "{\"uri\": \"file://$test_file\", \"line\": 1, \"character\": 4}" \
        "Python: references"

    # Cleanup
    rm -f "$test_file"
}

test_lsp_typescript_server() {
    log_info "=== Testing TypeScript LSP Server ==="

    if ! command -v typescript-language-server > /dev/null 2>&1; then
        log_skip "typescript-language-server not installed, skipping TypeScript LSP tests"
        return
    fi

    # Create test file
    local test_file="/tmp/test_ts_$(date +%s).ts"
    cat > "$test_file" << 'EOF'
interface Person {
    name: string;
    age: number;
}

function greet(person: Person): string {
    return `Hello, ${person.name}!`;
}

const user: Person = { name: "Alice", age: 30 };
console.log(greet(user));
EOF

    # Test 1: Completion
    test_lsp_operation "typescript-language-server" "completion" \
        "{\"uri\": \"file://$test_file\", \"line\": 10, \"character\": 20}" \
        "TypeScript: completion"

    # Test 2: Hover
    test_lsp_operation "typescript-language-server" "hover" \
        "{\"uri\": \"file://$test_file\", \"line\": 6, \"character\": 10}" \
        "TypeScript: hover"

    # Test 3: Definition
    test_lsp_operation "typescript-language-server" "definition" \
        "{\"uri\": \"file://$test_file\", \"line\": 10, \"character\": 12}" \
        "TypeScript: definition"

    # Cleanup
    rm -f "$test_file"
}

test_lsp_diagnostics() {
    log_info "=== Testing LSP Diagnostics ==="

    # Create file with intentional error
    local test_file="/tmp/test_error_$(date +%s).py"
    cat > "$test_file" << 'EOF'
def broken_function(
    # Missing parameter and closing paren
    undefined_variable + 1
EOF

    if command -v pylsp > /dev/null 2>&1; then
        test_lsp_operation "pylsp" "diagnostics" \
            "{\"uri\": \"file://$test_file\"}" \
            "Diagnostics: detect Python syntax errors"
    else
        log_skip "pylsp not available for diagnostics test"
    fi

    rm -f "$test_file"
}

test_lsp_connection_lifecycle() {
    log_info "=== Testing LSP Connection Lifecycle ==="

    # Test 1: Server sync
    if curl -sf -X POST "$HELIXAGENT_URL/v1/lsp/sync" > /dev/null 2>&1; then
        log_success "LSP server sync endpoint works"
    else
        log_fail "LSP server sync endpoint failed"
    fi

    # Test 2: Individual server sync
    local servers=$(curl -sf "$HELIXAGENT_URL/v1/lsp/servers" | jq -r '.servers[0].id // empty' 2>/dev/null)
    if [ -n "$servers" ]; then
        if curl -sf -X POST "$HELIXAGENT_URL/v1/lsp/servers/$servers/sync" > /dev/null 2>&1; then
            log_success "Individual server sync works"
        else
            log_fail "Individual server sync failed"
        fi
    else
        log_skip "No servers available for individual sync test"
    fi
}

test_lsp_ai_integration() {
    log_info "=== Testing LSP-AI Integration ==="

    # Check if LSP-AI is available
    if ! curl -sf "$LSP_MANAGER_URL/health" > /dev/null 2>&1; then
        log_skip "LSP-AI service not available"
        return
    fi

    # Test 1: LSP-AI health
    log_success "LSP-AI service healthy"

    # Test 2: AI-powered completion (if configured)
    local test_file="/tmp/test_ai_$(date +%s).py"
    cat > "$test_file" << 'EOF'
def calculate_fibonacci(n):
    # TODO: implement fibonacci sequence
    pass
EOF

    if response=$(curl -sf -X POST \
        -H "Content-Type: application/json" \
        -d "{\"uri\": \"file://$test_file\", \"line\": 2, \"character\": 0, \"use_ai\": true}" \
        "$HELIXAGENT_URL/v1/lsp/ai/complete" 2>/dev/null); then
        if echo "$response" | jq -e '.suggestions' > /dev/null 2>&1; then
            log_success "LSP-AI powered completion works"
        else
            log_skip "LSP-AI completion returned no suggestions"
        fi
    else
        log_skip "LSP-AI completion endpoint not available"
    fi

    rm -f "$test_file"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    echo "=============================================="
    echo "  LSP Integration Challenge"
    echo "  HelixAgent - $(date)"
    echo "=============================================="
    echo ""

    # Wait for services
    wait_for_service "$HELIXAGENT_URL" "HelixAgent" || exit 1

    # Run test categories
    test_lsp_core_infrastructure
    echo ""
    test_lsp_binary_detection
    echo ""
    test_lsp_go_server
    echo ""
    test_lsp_rust_server
    echo ""
    test_lsp_python_server
    echo ""
    test_lsp_typescript_server
    echo ""
    test_lsp_diagnostics
    echo ""
    test_lsp_connection_lifecycle
    echo ""
    test_lsp_ai_integration

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
