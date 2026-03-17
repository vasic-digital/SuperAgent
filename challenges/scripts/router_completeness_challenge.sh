#!/usr/bin/env bash
set -euo pipefail

# Router Completeness Challenge
# Validates that all handler constructors in internal/handlers/*.go are either
# registered in the router or explicitly listed in the exclusion list.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

setup_challenge "router_completeness" "$@"

PASSED=0
FAILED=0
TOTAL=0
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

ROUTER_FILE="$PROJECT_ROOT/internal/router/router.go"
HANDLERS_DIR="$PROJECT_ROOT/internal/handlers"

# Handlers that are intentionally NOT registered in router.go.
# Each entry must have a comment explaining why it is excluded.
EXCLUSION_LIST=(
    "NewCompletionHandler"           # Legacy: replaced by NewUnifiedHandler for OpenAI-compatible routes
    "NewCompletionHandlerWithSkills" # Legacy: replaced by NewUnifiedHandler with skills injection
    "NewDebateHandlerWithSkills"     # Variant of NewDebateHandler; router uses NewDebateHandler directly
    "NewProtocolSSEHandler"          # Router uses NewProtocolSSEHandlerWithACP (superset)
)

print_header "Router Completeness Challenge"
echo "Validates that every handler constructor is wired into the router"
echo ""

# =============================================================================
# Test 1: Router file exists
# =============================================================================
test_start "Router file exists"
TOTAL=$((TOTAL + 1))
if [ -f "$ROUTER_FILE" ]; then
    test_pass
    PASSED=$((PASSED + 1))
else
    test_fail "Router file not found at $ROUTER_FILE"
    FAILED=$((FAILED + 1))
fi

# =============================================================================
# Test 2: Handlers directory exists
# =============================================================================
test_start "Handlers directory exists"
TOTAL=$((TOTAL + 1))
if [ -d "$HANDLERS_DIR" ]; then
    test_pass
    PASSED=$((PASSED + 1))
else
    test_fail "Handlers directory not found at $HANDLERS_DIR"
    FAILED=$((FAILED + 1))
fi

# =============================================================================
# Test 3: Extract all handler constructors and verify registration
# =============================================================================
echo ""
echo "--- Handler Registration Checks ---"

# Extract all New*Handler constructor function names from non-test handler files.
# We match the pattern: func NewXxxHandler or func NewXxxHandlerYyy
constructors=()
while IFS= read -r line; do
    constructors+=("$line")
done < <(grep -rh '^func New[A-Z][A-Za-z]*Handler[A-Za-z]*(' "$HANDLERS_DIR"/*.go 2>/dev/null \
    | sed 's/func \(New[A-Za-z]*Handler[A-Za-z]*\)(.*/\1/' \
    | sort -u)

if [ ${#constructors[@]} -eq 0 ]; then
    test_start "Found handler constructors"
    TOTAL=$((TOTAL + 1))
    test_fail "No handler constructors found"
    FAILED=$((FAILED + 1))
else
    test_start "Found handler constructors (${#constructors[@]} total)"
    TOTAL=$((TOTAL + 1))
    test_pass
    PASSED=$((PASSED + 1))
fi

# Check each constructor is either in router.go or in the exclusion list
unregistered=()
for constructor in "${constructors[@]}"; do
    test_start "Handler $constructor is registered or excluded"
    TOTAL=$((TOTAL + 1))

    # Check exclusion list
    excluded=false
    for excl in "${EXCLUSION_LIST[@]}"; do
        if [ "$constructor" = "$excl" ]; then
            excluded=true
            break
        fi
    done

    if [ "$excluded" = true ]; then
        test_pass
        PASSED=$((PASSED + 1))
        continue
    fi

    # Check if it appears in router.go (either as direct call or via RegisterRoutes pattern)
    if grep -q "$constructor" "$ROUTER_FILE" 2>/dev/null; then
        test_pass
        PASSED=$((PASSED + 1))
    else
        test_fail "$constructor not found in router.go and not in exclusion list"
        FAILED=$((FAILED + 1))
        unregistered+=("$constructor")
    fi
done

# =============================================================================
# Test 4: Verify key route groups exist (either as Group() or via RegisterRoutes)
# =============================================================================
echo ""
echo "--- Key Route Group Existence ---"

# These route groups must appear somewhere in router.go, either as
# Group("/xxx") or as a comment/log message referencing the path.
key_groups=("/discovery" "/scoring" "/verification" "/mcp" "/lsp" "/rag" "/embeddings" "/providers" "/sessions" "/agents" "/protocols" "/skills")

for group in "${key_groups[@]}"; do
    test_start "Route group $group is registered"
    TOTAL=$((TOTAL + 1))
    if grep -q "\"$group\"" "$ROUTER_FILE" 2>/dev/null; then
        test_pass
        PASSED=$((PASSED + 1))
    else
        test_fail "Route group $group not found in router.go"
        FAILED=$((FAILED + 1))
    fi
done

# These groups are registered via handler.RegisterRoutes() which creates
# the group internally. We check for the handler instantiation + RegisterRoutes call
# or for a log message referencing the path.
registered_via_handler=("/tasks" "/monitoring" "/health")
for group in "${registered_via_handler[@]}"; do
    test_start "Route group $group is registered (via handler or group)"
    TOTAL=$((TOTAL + 1))
    # Check either Group("/xxx") or a log/comment mentioning the path or skip list
    if grep -qiE "(\"$group\"|/v1$group|${group#/}.*endpoint|${group#/}.*register)" "$ROUTER_FILE" 2>/dev/null; then
        test_pass
        PASSED=$((PASSED + 1))
    else
        test_fail "Route group $group not found in router.go"
        FAILED=$((FAILED + 1))
    fi
done

# =============================================================================
# Test 5: Verify no dead handler files remain (known dead patterns)
# =============================================================================
echo ""
echo "--- Dead Handler Checks ---"

dead_handlers=("cognee.go" "graphql_handler.go" "openrouter_models.go")
for dead in "${dead_handlers[@]}"; do
    test_start "Dead handler $dead is removed"
    TOTAL=$((TOTAL + 1))
    if [ -f "$HANDLERS_DIR/$dead" ]; then
        test_fail "$dead still exists in handlers directory"
        FAILED=$((FAILED + 1))
    else
        test_pass
        PASSED=$((PASSED + 1))
    fi
done

# =============================================================================
# Summary
# =============================================================================
echo ""
if [ ${#unregistered[@]} -gt 0 ]; then
    echo "Unregistered handlers:"
    for h in "${unregistered[@]}"; do
        echo "  - $h"
    done
    echo ""
fi

print_summary "Router Completeness Challenge" "$PASSED" "$FAILED"

if [ "$FAILED" -gt 0 ]; then
    exit 1
fi
exit 0
