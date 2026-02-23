#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0

pass() { PASS=$((PASS+1)); echo "PASS: $1"; }
fail() { FAIL=$((FAIL+1)); echo "FAIL: $1"; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_ROOT="$SCRIPT_DIR/../.."

# go.mod exists
[ -f "$MODULE_ROOT/go.mod" ] && pass "go.mod exists" || fail "go.mod missing"

# go.mod declares correct module name
grep -q "module digital.vasic.planning" "$MODULE_ROOT/go.mod" \
  && pass "go.mod module name is digital.vasic.planning" \
  || fail "go.mod module name incorrect"

# Source files exist
[ -f "$MODULE_ROOT/planning/hiplan.go" ] && pass "hiplan.go exists" || fail "hiplan.go missing"
[ -f "$MODULE_ROOT/planning/mcts.go" ] && pass "mcts.go exists" || fail "mcts.go missing"
[ -f "$MODULE_ROOT/planning/tree_of_thoughts.go" ] && pass "tree_of_thoughts.go exists" || fail "tree_of_thoughts.go missing"

# Documentation files exist
[ -f "$MODULE_ROOT/CLAUDE.md" ] && pass "CLAUDE.md exists" || fail "CLAUDE.md missing"
[ -f "$MODULE_ROOT/README.md" ] && pass "README.md exists" || fail "README.md missing"
[ -f "$MODULE_ROOT/AGENTS.md" ] && pass "AGENTS.md exists" || fail "AGENTS.md missing"

# Package builds
(cd "$MODULE_ROOT" && go build ./... 2>&1) \
  && pass "go build ./... succeeds" \
  || fail "go build ./... failed"

# Key exported types are present in source
grep -q "type HiPlan struct" "$MODULE_ROOT/planning/hiplan.go" \
  && pass "HiPlan struct defined" \
  || fail "HiPlan struct missing"

grep -q "type MCTS struct" "$MODULE_ROOT/planning/mcts.go" \
  && pass "MCTS struct defined" \
  || fail "MCTS struct missing"

grep -q "type TreeOfThoughts struct" "$MODULE_ROOT/planning/tree_of_thoughts.go" \
  && pass "TreeOfThoughts struct defined" \
  || fail "TreeOfThoughts struct missing"

# Key interfaces are present
grep -q "type MilestoneGenerator interface" "$MODULE_ROOT/planning/hiplan.go" \
  && pass "MilestoneGenerator interface defined" \
  || fail "MilestoneGenerator interface missing"

grep -q "type MCTSActionGenerator interface" "$MODULE_ROOT/planning/mcts.go" \
  && pass "MCTSActionGenerator interface defined" \
  || fail "MCTSActionGenerator interface missing"

grep -q "type ThoughtGenerator interface" "$MODULE_ROOT/planning/tree_of_thoughts.go" \
  && pass "ThoughtGenerator interface defined" \
  || fail "ThoughtGenerator interface missing"

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
