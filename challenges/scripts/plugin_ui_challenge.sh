#!/bin/bash
# HelixAgent Plugin UI Challenge
# Tests debate rendering and progress bar visualization
# 15 tests total

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
TOTAL=15

echo "========================================"
echo "HelixAgent Plugin UI Challenge"
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

# Test 1-5: Debate renderer
echo "--- Debate Renderer ---"
run_test "Debate renderer exists" "test -f '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"
run_test "DebateRenderer class" "grep -q 'class DebateRenderer' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"
run_test "RenderStyle type" "grep -q 'RenderStyle' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"
run_test "renderDebate method" "grep -q 'renderDebate' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"
run_test "renderResponse method" "grep -q 'renderResponse' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"

# Test 6-10: Render styles
echo ""
echo "--- Render Styles ---"
run_test "Theater style" "grep -q 'theater' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"
run_test "Novel style" "grep -q 'novel' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"
run_test "Screenplay style" "grep -q 'screenplay' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"
run_test "Minimal style" "grep -q 'minimal' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"
run_test "Plain style" "grep -q 'plain' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"

# Test 11-15: Progress renderer
echo ""
echo "--- Progress Renderer ---"
run_test "ProgressRenderer class" "grep -q 'class ProgressRenderer' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"
run_test "ASCII progress style" "grep -q 'ascii' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"
run_test "Unicode progress style" "grep -q 'unicode' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"
run_test "Block progress style" "grep -q 'block' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"
run_test "Dots progress style" "grep -q 'dots' '$PROJECT_ROOT/plugins/packages/ui/debate_renderer.ts'"

# Summary
echo ""
echo "========================================"
echo "UI Challenge Results"
echo "========================================"
echo -e "Passed: ${GREEN}$PASSED${NC}/$TOTAL"
echo -e "Failed: ${RED}$FAILED${NC}/$TOTAL"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All UI tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some UI tests failed${NC}"
    exit 1
fi
