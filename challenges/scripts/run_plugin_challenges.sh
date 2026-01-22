#!/bin/bash
# HelixAgent Plugin Challenges Runner
# Runs all plugin-related challenge scripts
# Total: 90 tests across 4 challenge scripts

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

TOTAL_PASSED=0
TOTAL_FAILED=0
CHALLENGES_PASSED=0
CHALLENGES_FAILED=0

echo "=============================================="
echo "HelixAgent Plugin Challenges Suite"
echo "=============================================="
echo ""
echo "Running 4 plugin challenge scripts (90 tests)"
echo ""

# Run each challenge script
run_challenge() {
    local name="$1"
    local script="$2"

    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}Running: $name${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""

    if bash "$SCRIPT_DIR/$script"; then
        CHALLENGES_PASSED=$((CHALLENGES_PASSED + 1))
        echo ""
        echo -e "${GREEN}$name: PASSED${NC}"
    else
        CHALLENGES_FAILED=$((CHALLENGES_FAILED + 1))
        echo ""
        echo -e "${RED}$name: FAILED${NC}"
    fi
    echo ""
}

# Run all plugin challenges
run_challenge "Plugin Transport Challenge (25 tests)" "plugin_transport_challenge.sh"
run_challenge "Plugin Events Challenge (20 tests)" "plugin_events_challenge.sh"
run_challenge "Plugin UI Challenge (15 tests)" "plugin_ui_challenge.sh"
run_challenge "Plugin Integration Challenge (30 tests)" "plugin_integration_challenge.sh"

# Final summary
echo "=============================================="
echo "Plugin Challenges Suite Results"
echo "=============================================="
echo ""
echo -e "Challenges Passed: ${GREEN}$CHALLENGES_PASSED${NC}/4"
echo -e "Challenges Failed: ${RED}$CHALLENGES_FAILED${NC}/4"
echo ""

if [ $CHALLENGES_FAILED -eq 0 ]; then
    echo -e "${GREEN}=============================================="
    echo "ALL PLUGIN CHALLENGES PASSED!"
    echo -e "==============================================${NC}"
    exit 0
else
    echo -e "${RED}=============================================="
    echo "SOME PLUGIN CHALLENGES FAILED"
    echo -e "==============================================${NC}"
    exit 1
fi
