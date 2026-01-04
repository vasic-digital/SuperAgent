#!/bin/bash
# SuperAgent Challenges - Run All Challenges in Sequence
# Usage: ./scripts/run_all_challenges.sh [options]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Challenges in dependency order
CHALLENGES=(
    "provider_verification"
    "ai_debate_formation"
    "api_quality_test"
)

# Parse options
VERBOSE=""
STOP_ON_FAILURE=true

while [ $# -gt 0 ]; do
    case "$1" in
        -v|--verbose)
            VERBOSE="--verbose"
            ;;
        --continue-on-failure)
            STOP_ON_FAILURE=false
            ;;
        -h|--help)
            echo "Usage: $0 [--verbose] [--continue-on-failure]"
            exit 0
            ;;
    esac
    shift
done

# Main execution
print_info "=========================================="
print_info "  SuperAgent - Run All Challenges"
print_info "=========================================="
print_info "Start time: $(date)"
echo ""

TOTAL_START=$(date +%s)
PASSED=0
FAILED=0

for challenge in "${CHALLENGES[@]}"; do
    print_info "----------------------------------------"
    print_info "Running: $challenge"
    print_info "----------------------------------------"

    if "$SCRIPT_DIR/run_challenges.sh" "$challenge" $VERBOSE; then
        PASSED=$((PASSED + 1))
        print_success "$challenge completed successfully"
    else
        FAILED=$((FAILED + 1))
        print_error "$challenge failed"

        if [ "$STOP_ON_FAILURE" = true ]; then
            print_error "Stopping due to failure. Use --continue-on-failure to continue."
            break
        fi
    fi
    echo ""
done

TOTAL_END=$(date +%s)
TOTAL_DURATION=$((TOTAL_END - TOTAL_START))

# Generate master summary
print_info "Generating master summary..."
"$SCRIPT_DIR/generate_report.sh" --master-only 2>/dev/null || true

# Final report
echo ""
print_info "=========================================="
print_info "  All Challenges Complete"
print_info "=========================================="
print_info "Total duration: ${TOTAL_DURATION}s"
print_info "Passed: $PASSED / ${#CHALLENGES[@]}"
print_info "Failed: $FAILED / ${#CHALLENGES[@]}"

if [ $FAILED -eq 0 ]; then
    print_success "All challenges passed!"
    exit 0
else
    print_error "$FAILED challenge(s) failed"
    exit 1
fi
