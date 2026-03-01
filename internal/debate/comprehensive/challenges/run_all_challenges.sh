#!/bin/bash
#
# Master Challenge Script
# Runs all challenges for the comprehensive debate system
#

set -e

echo "========================================="
echo "Comprehensive Debate System Challenges"
echo "========================================="

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../.."

TOTAL_CHALLENGES=0
PASSED_CHALLENGES=0
FAILED_CHALLENGES=0

# Run each challenge
for challenge in "$SCRIPT_DIR"/challenge_*.sh; do
    if [ -f "$challenge" ]; then
        challenge_name=$(basename "$challenge" .sh)
        ((TOTAL_CHALLENGES++))
        
        echo -e "\n${BLUE}=========================================${NC}"
        echo -e "${BLUE}Running: $challenge_name${NC}"
        echo -e "${BLUE}=========================================${NC}"
        
        if bash "$challenge"; then
            ((PASSED_CHALLENGES++))
        else
            ((FAILED_CHALLENGES++))
        fi
    fi
done

# Summary
echo -e "\n${BLUE}=========================================${NC}"
echo "Challenge Summary"
echo -e "${BLUE}=========================================${NC}"
echo -e "Total Challenges: $TOTAL_CHALLENGES"
echo -e "${GREEN}Passed: $PASSED_CHALLENGES${NC}"
echo -e "${RED}Failed: $FAILED_CHALLENGES${NC}"

if [ $FAILED_CHALLENGES -eq 0 ]; then
    echo -e "\n${GREEN}✓ ALL CHALLENGES PASSED${NC}"
    echo -e "${GREEN}✓ Comprehensive Debate System is READY${NC}"
    exit 0
else
    echo -e "\n${RED}✗ SOME CHALLENGES FAILED${NC}"
    exit 1
fi
