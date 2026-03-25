#!/bin/bash
# Dead Code Verification Challenge
# Validates that no dead code packages exist in the codebase

set -euo pipefail

PASS=0
FAIL=0
TOTAL=0

check() {
    local desc="$1"
    local result="$2"
    TOTAL=$((TOTAL + 1))
    if [ "$result" = "0" ]; then
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: $desc"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== Dead Code Verification Challenge ==="
echo ""

# 1. No duplicate method declarations
echo "--- Compiler Integrity ---"
DUP=$(go vet ./internal/debate/comprehensive/ 2>&1 | grep -c "already declared" || true)
check "No duplicate method declarations in debate/comprehensive" "$DUP"

# 2. No dead backup directories
echo "--- Dead Directories ---"
BACKUP=$(test -d internal/background/backup && echo 1 || echo 0)
check "No internal/background/backup/ directory" "$BACKUP"

# 3. No unimported adapter packages
echo "--- Dead Adapter Packages ---"
for pkg in background observability events http helixqa; do
    if [ -d "internal/adapters/$pkg" ]; then
        IMPORTS=$(grep -rn "\"dev.helix.agent/internal/adapters/$pkg\"" internal/ --include='*.go' 2>/dev/null | grep -v '_test.go' | wc -l)
        if [ "$IMPORTS" = "0" ]; then
            check "internal/adapters/$pkg/ is imported or removed" "1"
        else
            check "internal/adapters/$pkg/ is imported or removed" "0"
        fi
    else
        check "internal/adapters/$pkg/ is imported or removed" "0"
    fi
done

# 4. Skills routes not inside handler closure
echo "--- Route Registration ---"
SKILLS_IN_HANDLER=$(grep -A2 'response\["healthy"\] = true' internal/router/router.go | grep -c "Skills" || true)
check "Skills routes not inside health handler closure" "$SKILLS_IN_HANDLER"

# 5. Build succeeds
echo "--- Build Verification ---"
if go build ./... 2>/dev/null; then
    check "Full project builds cleanly" "0"
else
    check "Full project builds cleanly" "1"
fi

# 6. go vet passes
if go vet ./internal/... 2>/dev/null; then
    check "go vet passes on internal/" "0"
else
    check "go vet passes on internal/" "1"
fi

echo ""
echo "=== Results: $PASS/$TOTAL passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
