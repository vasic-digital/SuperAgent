#!/bin/bash

# =============================================================================
# ALL PROTOCOLS COMPREHENSIVE VALIDATION
#
# Master validation script that runs all protocol validations:
# - MCP (Model Context Protocol)
# - LSP (Language Server Protocol)
# - ACP (Agent Communication Protocol)
# - Embeddings
# - Vision
#
# Usage: ./challenges/scripts/all_protocols_validation.sh
# =============================================================================

set -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
WHITE='\033[1;37m'
NC='\033[0m'

# Tracking
TOTAL_PASSED=0
TOTAL_FAILED=0
TOTAL_SKIPPED=0
TOTAL_TESTS=0

declare -A PROTOCOL_RESULTS

echo ""
echo -e "${WHITE}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${WHITE}║             HELIXAGENT ALL PROTOCOLS COMPREHENSIVE VALIDATION               ║${NC}"
echo -e "${WHITE}║                                                                              ║${NC}"
echo -e "${WHITE}║  Testing: MCP • LSP • ACP • Embeddings • Vision • Cognee                   ║${NC}"
echo -e "${WHITE}╚══════════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Function to run a validation script and capture results
run_validation() {
    local name="$1"
    local script="$2"
    local icon="$3"

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}${icon} Running ${name} Validation${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    if [ ! -f "$script" ]; then
        echo -e "${YELLOW}⚠ Script not found: $script${NC}"
        PROTOCOL_RESULTS[$name]="NOT_FOUND"
        return
    fi

    chmod +x "$script" 2>/dev/null

    # Run and capture output
    output=$("$script" 2>&1)
    exit_code=$?

    echo "$output"

    # Parse results from output (using sed for broader compatibility)
    passed=$(echo "$output" | grep -E 'Passed:' | tail -1 | sed 's/.*Passed:[[:space:]]*\([0-9]*\).*/\1/' | tr -d '\033[0-9;m')
    failed=$(echo "$output" | grep -E 'Failed:' | tail -1 | sed 's/.*Failed:[[:space:]]*\([0-9]*\).*/\1/' | tr -d '\033[0-9;m')
    skipped=$(echo "$output" | grep -E 'Skipped:' | tail -1 | sed 's/.*Skipped:[[:space:]]*\([0-9]*\).*/\1/' | tr -d '\033[0-9;m')
    total=$(echo "$output" | grep -E 'Total Tests:' | tail -1 | sed 's/.*Total Tests:[[:space:]]*\([0-9]*\).*/\1/' | tr -d '\033[0-9;m')

    # Set defaults if not found
    passed=${passed:-0}
    failed=${failed:-0}
    skipped=${skipped:-0}
    total=${total:-0}

    # Update totals
    ((TOTAL_PASSED += passed))
    ((TOTAL_FAILED += failed))
    ((TOTAL_SKIPPED += skipped))
    ((TOTAL_TESTS += total))

    # Store result
    if [ $exit_code -eq 0 ]; then
        PROTOCOL_RESULTS[$name]="PASS:$passed/$total"
    else
        PROTOCOL_RESULTS[$name]="FAIL:$passed/$total"
    fi

    echo ""
}

# =============================================================================
# RUN ALL PROTOCOL VALIDATIONS
# =============================================================================

run_validation "MCP" "$SCRIPT_DIR/mcp_validation_comprehensive.sh" "🔌"
run_validation "LSP" "$SCRIPT_DIR/lsp_validation_comprehensive.sh" "📝"
run_validation "ACP" "$SCRIPT_DIR/acp_validation_comprehensive.sh" "🤖"
run_validation "Embeddings" "$SCRIPT_DIR/embeddings_validation_comprehensive.sh" "🧮"
run_validation "Vision" "$SCRIPT_DIR/vision_validation_comprehensive.sh" "👁️"
run_validation "Cognee" "$SCRIPT_DIR/cognee_validation_comprehensive.sh" "🧠"

# =============================================================================
# SUMMARY
# =============================================================================
echo ""
echo -e "${WHITE}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${WHITE}║                    ALL PROTOCOLS VALIDATION SUMMARY                          ║${NC}"
echo -e "${WHITE}╠══════════════════════════════════════════════════════════════════════════════╣${NC}"

# Protocol results table
printf "${WHITE}║${NC}  %-15s │ %-12s │ %-30s ${WHITE}║${NC}\n" "PROTOCOL" "STATUS" "RESULT"
echo -e "${WHITE}╠══════════════════════════════════════════════════════════════════════════════╣${NC}"

for protocol in MCP LSP ACP Embeddings Vision Cognee; do
    result="${PROTOCOL_RESULTS[$protocol]:-N/A}"
    status="${result%%:*}"
    details="${result#*:}"

    case "$status" in
        PASS)
            status_color="${GREEN}✓ PASS${NC}"
            ;;
        FAIL)
            status_color="${RED}✗ FAIL${NC}"
            ;;
        NOT_FOUND)
            status_color="${YELLOW}○ SKIP${NC}"
            details="Script not found"
            ;;
        *)
            status_color="${YELLOW}○ N/A${NC}"
            details="Not run"
            ;;
    esac

    printf "${WHITE}║${NC}  %-15s │ %-22b │ %-30s ${WHITE}║${NC}\n" "$protocol" "$status_color" "$details"
done

echo -e "${WHITE}╠══════════════════════════════════════════════════════════════════════════════╣${NC}"

# Totals
echo -e "${WHITE}║${NC}  ${BLUE}Total Tests:    ${TOTAL_TESTS}${NC}"
echo -e "${WHITE}║${NC}  ${GREEN}Passed:         ${TOTAL_PASSED}${NC}"
echo -e "${WHITE}║${NC}  ${RED}Failed:         ${TOTAL_FAILED}${NC}"
echo -e "${WHITE}║${NC}  ${YELLOW}Skipped:        ${TOTAL_SKIPPED}${NC}"
echo -e "${WHITE}╠══════════════════════════════════════════════════════════════════════════════╣${NC}"

# Overall pass rate
if [ $((TOTAL_PASSED + TOTAL_FAILED)) -gt 0 ]; then
    PASS_RATE=$((TOTAL_PASSED * 100 / (TOTAL_PASSED + TOTAL_FAILED)))
    echo -e "${WHITE}║${NC}  ${GREEN}Pass Rate:      ${PASS_RATE}%${NC} (of non-skipped tests)"
else
    PASS_RATE=100
    echo -e "${WHITE}║${NC}  ${GREEN}Pass Rate:      100%${NC} (no failures)"
fi

echo -e "${WHITE}╚══════════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Final verdict
if [ $TOTAL_FAILED -gt 0 ]; then
    echo -e "${RED}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║                    VALIDATION FAILED - $TOTAL_FAILED test(s) failed                      ║${NC}"
    echo -e "${RED}╚══════════════════════════════════════════════════════════════════════════════╝${NC}"
    exit 1
else
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                    ALL PROTOCOLS VALIDATION PASSED                           ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════════════════╝${NC}"
    exit 0
fi
