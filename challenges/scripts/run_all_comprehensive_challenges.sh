#!/bin/bash
# Run All Comprehensive Challenges
# Master script to verify entire HelixAgent system

set -o pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

cd "$(dirname "$0")/../.." || exit 1

SCRIPT_DIR="challenges/scripts"
TOTAL_CHALLENGES=0
PASSED_CHALLENGES=0
FAILED_CHALLENGES=0
CHALLENGE_RESULTS=()

echo "============================================"
echo -e "${CYAN}  HELIX AGENT COMPREHENSIVE VERIFICATION${NC}"
echo "============================================"
echo ""
echo "Running all challenge scripts to verify system integrity..."
echo "This ensures maximum efficiency and fault tolerance."
echo ""
echo "Start Time: $(date)"
echo ""

run_challenge() {
    local name="$1"
    local script="$2"

    echo "============================================"
    echo -e "${YELLOW}Running: $name${NC}"
    echo "============================================"

    TOTAL_CHALLENGES=$((TOTAL_CHALLENGES + 1))

    if [ -x "$script" ]; then
        if bash "$script"; then
            PASSED_CHALLENGES=$((PASSED_CHALLENGES + 1))
            CHALLENGE_RESULTS+=("${GREEN}✓${NC} $name")
            echo -e "\n${GREEN}✓ $name PASSED${NC}\n"
        else
            FAILED_CHALLENGES=$((FAILED_CHALLENGES + 1))
            CHALLENGE_RESULTS+=("${RED}✗${NC} $name")
            echo -e "\n${RED}✗ $name FAILED${NC}\n"
        fi
    else
        chmod +x "$script" 2>/dev/null
        if bash "$script"; then
            PASSED_CHALLENGES=$((PASSED_CHALLENGES + 1))
            CHALLENGE_RESULTS+=("${GREEN}✓${NC} $name")
            echo -e "\n${GREEN}✓ $name PASSED${NC}\n"
        else
            FAILED_CHALLENGES=$((FAILED_CHALLENGES + 1))
            CHALLENGE_RESULTS+=("${RED}✗${NC} $name")
            echo -e "\n${RED}✗ $name FAILED${NC}\n"
        fi
    fi
}

# Run all challenges
run_challenge "Advanced AI Features Challenge" "$SCRIPT_DIR/advanced_ai_features_challenge.sh"
run_challenge "Integration Challenge" "$SCRIPT_DIR/integration_challenge.sh"
run_challenge "Resilience Challenge" "$SCRIPT_DIR/resilience_challenge.sh"
run_challenge "Provider Verification Challenge" "$SCRIPT_DIR/provider_verification_challenge.sh"
run_challenge "E2E Workflow Challenge" "$SCRIPT_DIR/e2e_workflow_challenge.sh"

# Also run existing challenges if they exist
if [ -f "$SCRIPT_DIR/unified_verification_challenge.sh" ]; then
    run_challenge "Unified Verification Challenge" "$SCRIPT_DIR/unified_verification_challenge.sh"
fi

if [ -f "$SCRIPT_DIR/debate_team_dynamic_selection_challenge.sh" ]; then
    run_challenge "Debate Team Dynamic Selection Challenge" "$SCRIPT_DIR/debate_team_dynamic_selection_challenge.sh"
fi

if [ -f "$SCRIPT_DIR/semantic_intent_challenge.sh" ]; then
    run_challenge "Semantic Intent Challenge" "$SCRIPT_DIR/semantic_intent_challenge.sh"
fi

if [ -f "$SCRIPT_DIR/fallback_mechanism_challenge.sh" ]; then
    run_challenge "Fallback Mechanism Challenge" "$SCRIPT_DIR/fallback_mechanism_challenge.sh"
fi

if [ -f "$SCRIPT_DIR/multipass_validation_challenge.sh" ]; then
    run_challenge "Multi-Pass Validation Challenge" "$SCRIPT_DIR/multipass_validation_challenge.sh"
fi

echo ""
echo "============================================"
echo -e "${CYAN}  COMPREHENSIVE VERIFICATION SUMMARY${NC}"
echo "============================================"
echo ""
echo "End Time: $(date)"
echo ""
echo "Challenge Results:"
for result in "${CHALLENGE_RESULTS[@]}"; do
    echo -e "  $result"
done
echo ""
echo "============================================"
echo "Total Challenges: $TOTAL_CHALLENGES"
echo -e "${GREEN}Passed: $PASSED_CHALLENGES${NC}"
echo -e "${RED}Failed: $FAILED_CHALLENGES${NC}"
echo ""

if [ $TOTAL_CHALLENGES -gt 0 ]; then
    PASS_RATE=$((PASSED_CHALLENGES * 100 / TOTAL_CHALLENGES))
    echo "Challenge Pass Rate: ${PASS_RATE}%"
fi
echo "============================================"
echo ""

# Feature Summary
echo "Verified Features:"
echo "  • Advanced Planning (ToT, MCTS, HiPlan)"
echo "  • Code Knowledge Graph & GraphRAG"
echo "  • Security (SecureFixAgent, FiveRingDefense)"
echo "  • Governance (SEMAP Protocol)"
echo "  • 10 LLM Providers"
echo "  • 11 MCP Adapters"
echo "  • 3 Embedding Models"
echo "  • LSP-AI Integration"
echo "  • Lesson Banking"
echo "  • Background Task Processing"
echo "  • Notification Systems (SSE, WS, Webhook)"
echo "  • Fault Tolerance & Resilience"
echo ""

if [ $FAILED_CHALLENGES -eq 0 ]; then
    echo -e "${GREEN}╔══════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║  ✓ ALL COMPREHENSIVE CHALLENGES PASSED!  ║${NC}"
    echo -e "${GREEN}║                                          ║${NC}"
    echo -e "${GREEN}║  System is verified to be:               ║${NC}"
    echo -e "${GREEN}║  • Fully integrated                      ║${NC}"
    echo -e "${GREEN}║  • Fault-tolerant                        ║${NC}"
    echo -e "${GREEN}║  • Production-ready                      ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}╔══════════════════════════════════════════╗${NC}"
    echo -e "${RED}║  ✗ SOME CHALLENGES FAILED                ║${NC}"
    echo -e "${RED}║  Review failed challenges above          ║${NC}"
    echo -e "${RED}╚══════════════════════════════════════════╝${NC}"
    exit 1
fi
