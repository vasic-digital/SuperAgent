#!/bin/bash
# ============================================================================
# OPENCODE MCP STATUS CHALLENGE
# ============================================================================
# This challenge validates that ALL MCP servers in OpenCode are connected
# and working properly. The challenge FAILS if any MCP shows an error.
#
# This challenge runs `opencode mcp list` and verifies:
# - All MCPs show "connected" status
# - No "failed" or "MCP error" appears in the output
# - No "Connection closed" errors
#
# IMPORTANT: This is a ZERO TOLERANCE challenge - any MCP error = FAIL
# ============================================================================

set -euo pipefail

# Source common utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh" 2>/dev/null || true

# Configuration
RESULTS_DIR="${RESULTS_DIR:-${SCRIPT_DIR}/../results/opencode_mcp_status/$(date +%Y/%m/%d/%Y%m%d_%H%M%S)}"
TIMEOUT="${TIMEOUT:-120}"
VERBOSE="${VERBOSE:-false}"

# Counters
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%H:%M:%S') $*"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $(date '+%H:%M:%S') $*"
    ((TESTS_PASSED++))
    ((TOTAL_TESTS++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $(date '+%H:%M:%S') $*"
    ((TESTS_FAILED++))
    ((TOTAL_TESTS++))
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%H:%M:%S') $*"
}

log_header() {
    echo -e "\n${CYAN}============================================================${NC}"
    echo -e "${CYAN}$*${NC}"
    echo -e "${CYAN}============================================================${NC}\n"
}

# Create results directory
setup_results() {
    mkdir -p "$RESULTS_DIR"
    log_info "Results will be stored in: $RESULTS_DIR"
}

# Check if opencode is available
check_opencode() {
    log_header "CHECKING OPENCODE AVAILABILITY"

    if command -v opencode &> /dev/null; then
        local version=$(opencode --version 2>/dev/null || echo "unknown")
        log_pass "OpenCode is available: $version"
    else
        # Check common locations
        if [[ -f "$HOME/.opencode/bin/opencode" ]]; then
            export PATH="$HOME/.opencode/bin:$PATH"
            log_pass "OpenCode found at ~/.opencode/bin/opencode"
        else
            log_fail "OpenCode not found in PATH or ~/.opencode/bin"
            return 1
        fi
    fi
}

# Check uvx is available (for Python MCPs)
check_uvx() {
    log_header "CHECKING UVX AVAILABILITY"

    if command -v uvx &> /dev/null; then
        local version=$(uvx --version 2>/dev/null || echo "unknown")
        log_pass "uvx is available: $version"
    else
        # Check ~/.local/bin
        if [[ -f "$HOME/.local/bin/uvx" ]]; then
            export PATH="$HOME/.local/bin:$PATH"
            local version=$("$HOME/.local/bin/uvx" --version 2>/dev/null || echo "unknown")
            log_pass "uvx found at ~/.local/bin/uvx: $version"
        else
            log_fail "uvx not found - Python MCPs (fetch, git, time) will fail"
            return 1
        fi
    fi
}

# Run opencode mcp list and capture output
run_mcp_status() {
    log_header "RUNNING OPENCODE MCP STATUS CHECK"

    local output_file="$RESULTS_DIR/mcp_status_output.txt"
    local error_file="$RESULTS_DIR/mcp_status_errors.txt"

    # Ensure uvx is in PATH
    export PATH="$HOME/.local/bin:$HOME/.opencode/bin:$PATH"

    log_info "Running: opencode mcp list"

    # Run opencode mcp list with timeout
    if timeout "$TIMEOUT" opencode mcp list > "$output_file" 2> "$error_file"; then
        log_pass "opencode mcp list completed successfully"
    else
        local exit_code=$?
        if [[ $exit_code -eq 124 ]]; then
            log_fail "opencode mcp list timed out after ${TIMEOUT}s"
        else
            log_fail "opencode mcp list failed with exit code $exit_code"
        fi
        cat "$error_file" 2>/dev/null || true
        return 1
    fi

    # Display the output
    log_info "MCP Status Output:"
    cat "$output_file"
    echo ""
}

# Parse and validate MCP status
validate_mcp_status() {
    log_header "VALIDATING MCP STATUS"

    local output_file="$RESULTS_DIR/mcp_status_output.txt"

    if [[ ! -f "$output_file" ]]; then
        log_fail "MCP status output file not found"
        return 1
    fi

    # Count total MCPs (strip ANSI codes and handle newlines)
    local total_mcps
    total_mcps=$(sed 's/\x1b\[[0-9;]*m//g' "$output_file" | grep -c "●" 2>/dev/null) || total_mcps=0
    log_info "Total MCP servers found: $total_mcps"

    # Check for connected MCPs
    local connected_mcps
    connected_mcps=$(sed 's/\x1b\[[0-9;]*m//g' "$output_file" | grep -c "✓" 2>/dev/null) || connected_mcps=0
    log_info "Connected MCP servers: $connected_mcps"

    # Check for failed MCPs
    local failed_mcps
    failed_mcps=$(sed 's/\x1b\[[0-9;]*m//g' "$output_file" | grep -c "✗" 2>/dev/null) || failed_mcps=0
    log_info "Failed MCP servers: $failed_mcps"

    # Test 1: At least one MCP should be configured
    if [[ "$total_mcps" -gt 0 ]]; then
        log_pass "TEST 1: MCP servers are configured ($total_mcps total)"
    else
        log_fail "TEST 1: No MCP servers configured"
    fi

    # Test 2: No failed MCPs allowed
    if [[ "$failed_mcps" -eq 0 ]]; then
        log_pass "TEST 2: No failed MCP servers"
    else
        log_fail "TEST 2: Found $failed_mcps failed MCP server(s)"
        # List the failed MCPs
        log_info "Failed MCPs:"
        grep -B1 "✗.*failed" "$output_file" | grep "●" || true
    fi

    # Test 3: Check for "MCP error" in output
    if ! grep -q "MCP error" "$output_file"; then
        log_pass "TEST 3: No 'MCP error' messages found"
    else
        log_fail "TEST 3: Found 'MCP error' messages in output"
        grep "MCP error" "$output_file" || true
    fi

    # Test 4: Check for "Connection closed" errors
    if ! grep -q "Connection closed" "$output_file"; then
        log_pass "TEST 4: No 'Connection closed' errors found"
    else
        log_fail "TEST 4: Found 'Connection closed' errors in output"
        grep "Connection closed" "$output_file" || true
    fi

    # Test 5: All MCPs should be connected
    if [[ "$total_mcps" -gt 0 && "$connected_mcps" -eq "$total_mcps" ]]; then
        log_pass "TEST 5: All MCPs connected ($connected_mcps/$total_mcps)"
    else
        log_fail "TEST 5: Not all MCPs connected ($connected_mcps/$total_mcps)"
    fi

    # Test 6: Validate specific required MCPs
    local required_mcps=("filesystem" "memory" "helixagent-mcp")
    for mcp in "${required_mcps[@]}"; do
        if grep -q "✓.*$mcp" "$output_file"; then
            log_pass "TEST 6.$mcp: Required MCP '$mcp' is connected"
        else
            log_fail "TEST 6.$mcp: Required MCP '$mcp' is NOT connected"
        fi
    done

    # Test 7: Validate Python MCPs (using uvx)
    local python_mcps=("fetch" "git" "time")
    for mcp in "${python_mcps[@]}"; do
        if grep -q "✓.*$mcp" "$output_file"; then
            log_pass "TEST 7.$mcp: Python MCP '$mcp' (uvx) is connected"
        else
            log_fail "TEST 7.$mcp: Python MCP '$mcp' (uvx) is NOT connected"
        fi
    done

    # Test 8: Validate HelixAgent remote endpoints
    local helixagent_mcps=("helixagent-acp" "helixagent-lsp" "helixagent-embeddings" "helixagent-vision" "helixagent-cognee")
    for mcp in "${helixagent_mcps[@]}"; do
        if grep -q "✓.*$mcp" "$output_file"; then
            log_pass "TEST 8.$mcp: HelixAgent remote endpoint '$mcp' is connected"
        else
            log_fail "TEST 8.$mcp: HelixAgent remote endpoint '$mcp' is NOT connected"
        fi
    done
}

# Generate summary
generate_summary() {
    log_header "CHALLENGE SUMMARY"

    echo -e "Total Tests: $TOTAL_TESTS"
    echo -e "${GREEN}Passed:${NC}      $TESTS_PASSED"
    echo -e "${RED}Failed:${NC}      $TESTS_FAILED"
    echo ""

    if [[ "$TESTS_FAILED" -eq 0 ]]; then
        echo -e "${GREEN}============================================================${NC}"
        echo -e "${GREEN}  CHALLENGE PASSED: All MCP servers are working!${NC}"
        echo -e "${GREEN}============================================================${NC}"
        return 0
    else
        echo -e "${RED}============================================================${NC}"
        echo -e "${RED}  CHALLENGE FAILED: $TESTS_FAILED test(s) failed!${NC}"
        echo -e "${RED}============================================================${NC}"
        return 1
    fi
}

# Save results to file
save_results() {
    local results_file="$RESULTS_DIR/results.json"

    cat > "$results_file" << EOF
{
    "challenge": "opencode_mcp_status",
    "timestamp": "$(date -Iseconds)",
    "total_tests": $TOTAL_TESTS,
    "passed": $TESTS_PASSED,
    "failed": $TESTS_FAILED,
    "success": $([ "$TESTS_FAILED" -eq 0 ] && echo "true" || echo "false")
}
EOF

    log_info "Results saved to: $results_file"
}

# Main execution
main() {
    log_header "OPENCODE MCP STATUS CHALLENGE"
    log_info "This challenge validates that ALL MCP servers are connected"
    log_info "ZERO TOLERANCE: Any MCP error means the challenge FAILS"
    echo ""

    setup_results

    # Run checks
    check_opencode || true
    check_uvx || true
    run_mcp_status || true
    validate_mcp_status

    # Save and display results
    save_results
    generate_summary

    # Return appropriate exit code
    if [[ "$TESTS_FAILED" -eq 0 ]]; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@"
