#!/bin/bash
# graphql_integration_challenge.sh - GraphQL Integration Challenge
# Tests GraphQL API implementation for HelixAgent

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
source "$SCRIPT_DIR/common.sh"

CHALLENGE_NAME="GraphQL Integration Challenge"
PASSED=0
FAILED=0
TOTAL=0

log_test() {
    local test_name="$1"
    local status="$2"
    TOTAL=$((TOTAL + 1))
    if [ "$status" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        echo -e "  \e[32m✓\e[0m $test_name"
    else
        FAILED=$((FAILED + 1))
        echo -e "  \e[31m✗\e[0m $test_name"
    fi
}

echo "=============================================="
echo "  $CHALLENGE_NAME"
echo "=============================================="
echo ""

cd "$PROJECT_ROOT"

# Test 1: GraphQL package structure
echo "[1] GraphQL Package Structure"
if [ -f "internal/graphql/schema.go" ]; then
    log_test "schema.go exists" "PASS"
else
    log_test "schema.go exists" "FAIL"
fi

if [ -f "internal/graphql/schema_test.go" ]; then
    log_test "schema_test.go exists" "PASS"
else
    log_test "schema_test.go exists" "FAIL"
fi

if [ -d "internal/graphql/types" ]; then
    log_test "types directory exists" "PASS"
else
    log_test "types directory exists" "FAIL"
fi

if [ -f "internal/graphql/types/types.go" ]; then
    log_test "types/types.go exists" "PASS"
else
    log_test "types/types.go exists" "FAIL"
fi

# Test 2: Schema types defined
echo ""
echo "[2] Schema Types"
if grep -q "QueryType" internal/graphql/schema.go 2>/dev/null; then
    log_test "QueryType defined" "PASS"
else
    log_test "QueryType defined" "FAIL"
fi

if grep -q "MutationType" internal/graphql/schema.go 2>/dev/null; then
    log_test "MutationType defined" "PASS"
else
    log_test "MutationType defined" "FAIL"
fi

if grep -q "providerType" internal/graphql/schema.go 2>/dev/null; then
    log_test "providerType defined" "PASS"
else
    log_test "providerType defined" "FAIL"
fi

if grep -q "debateType" internal/graphql/schema.go 2>/dev/null; then
    log_test "debateType defined" "PASS"
else
    log_test "debateType defined" "FAIL"
fi

if grep -q "taskType" internal/graphql/schema.go 2>/dev/null; then
    log_test "taskType defined" "PASS"
else
    log_test "taskType defined" "FAIL"
fi

# Test 3: Query fields
echo ""
echo "[3] Query Fields"
if grep -q '"providers"' internal/graphql/schema.go 2>/dev/null; then
    log_test "providers query field" "PASS"
else
    log_test "providers query field" "FAIL"
fi

if grep -q '"debates"' internal/graphql/schema.go 2>/dev/null; then
    log_test "debates query field" "PASS"
else
    log_test "debates query field" "FAIL"
fi

if grep -q '"tasks"' internal/graphql/schema.go 2>/dev/null; then
    log_test "tasks query field" "PASS"
else
    log_test "tasks query field" "FAIL"
fi

if grep -q '"verificationResults"' internal/graphql/schema.go 2>/dev/null; then
    log_test "verificationResults query field" "PASS"
else
    log_test "verificationResults query field" "FAIL"
fi

if grep -q '"providerScores"' internal/graphql/schema.go 2>/dev/null; then
    log_test "providerScores query field" "PASS"
else
    log_test "providerScores query field" "FAIL"
fi

# Test 4: Mutation fields
echo ""
echo "[4] Mutation Fields"
if grep -q '"createDebate"' internal/graphql/schema.go 2>/dev/null; then
    log_test "createDebate mutation" "PASS"
else
    log_test "createDebate mutation" "FAIL"
fi

if grep -q '"submitDebateResponse"' internal/graphql/schema.go 2>/dev/null; then
    log_test "submitDebateResponse mutation" "PASS"
else
    log_test "submitDebateResponse mutation" "FAIL"
fi

if grep -q '"createTask"' internal/graphql/schema.go 2>/dev/null; then
    log_test "createTask mutation" "PASS"
else
    log_test "createTask mutation" "FAIL"
fi

if grep -q '"cancelTask"' internal/graphql/schema.go 2>/dev/null; then
    log_test "cancelTask mutation" "PASS"
else
    log_test "cancelTask mutation" "FAIL"
fi

if grep -q '"refreshProvider"' internal/graphql/schema.go 2>/dev/null; then
    log_test "refreshProvider mutation" "PASS"
else
    log_test "refreshProvider mutation" "FAIL"
fi

# Test 5: Input types
echo ""
echo "[5] Input Types"
if grep -q "providerFilterInput" internal/graphql/schema.go 2>/dev/null; then
    log_test "ProviderFilter input type" "PASS"
else
    log_test "ProviderFilter input type" "FAIL"
fi

if grep -q "createDebateInput" internal/graphql/schema.go 2>/dev/null; then
    log_test "CreateDebateInput type" "PASS"
else
    log_test "CreateDebateInput type" "FAIL"
fi

if grep -q "createTaskInput" internal/graphql/schema.go 2>/dev/null; then
    log_test "CreateTaskInput type" "PASS"
else
    log_test "CreateTaskInput type" "FAIL"
fi

# Test 6: Resolvers
echo ""
echo "[6] Resolvers"
if grep -q "ResolveProviders" internal/graphql/schema.go 2>/dev/null; then
    log_test "ResolveProviders resolver" "PASS"
else
    log_test "ResolveProviders resolver" "FAIL"
fi

if grep -q "ResolveDebates" internal/graphql/schema.go 2>/dev/null; then
    log_test "ResolveDebates resolver" "PASS"
else
    log_test "ResolveDebates resolver" "FAIL"
fi

if grep -q "ResolveTasks" internal/graphql/schema.go 2>/dev/null; then
    log_test "ResolveTasks resolver" "PASS"
else
    log_test "ResolveTasks resolver" "FAIL"
fi

if grep -q "ResolveCreateDebate" internal/graphql/schema.go 2>/dev/null; then
    log_test "ResolveCreateDebate resolver" "PASS"
else
    log_test "ResolveCreateDebate resolver" "FAIL"
fi

# Test 7: Schema initialization
echo ""
echo "[7] Schema Initialization"
if grep -q "InitSchema" internal/graphql/schema.go 2>/dev/null; then
    log_test "InitSchema function" "PASS"
else
    log_test "InitSchema function" "FAIL"
fi

if grep -q "ExecuteQuery" internal/graphql/schema.go 2>/dev/null; then
    log_test "ExecuteQuery function" "PASS"
else
    log_test "ExecuteQuery function" "FAIL"
fi

# Test 8: Types package
echo ""
echo "[8] Types Package"
if grep -q "type Provider struct" internal/graphql/types/types.go 2>/dev/null; then
    log_test "Provider struct" "PASS"
else
    log_test "Provider struct" "FAIL"
fi

if grep -q "type Model struct" internal/graphql/types/types.go 2>/dev/null; then
    log_test "Model struct" "PASS"
else
    log_test "Model struct" "FAIL"
fi

if grep -q "type Debate struct" internal/graphql/types/types.go 2>/dev/null; then
    log_test "Debate struct" "PASS"
else
    log_test "Debate struct" "FAIL"
fi

if grep -q "type Task struct" internal/graphql/types/types.go 2>/dev/null; then
    log_test "Task struct" "PASS"
else
    log_test "Task struct" "FAIL"
fi

if grep -q "type VerificationResults struct" internal/graphql/types/types.go 2>/dev/null; then
    log_test "VerificationResults struct" "PASS"
else
    log_test "VerificationResults struct" "FAIL"
fi

# Test 9: Unit tests
echo ""
echo "[9] Unit Tests"
if go test -v ./internal/graphql/... -count=1 2>&1 | grep -q "PASS"; then
    log_test "GraphQL unit tests pass" "PASS"
else
    log_test "GraphQL unit tests pass" "FAIL"
fi

# Test 10: graphql-go dependency
echo ""
echo "[10] Dependencies"
if grep -q "graphql-go/graphql" go.mod 2>/dev/null; then
    log_test "graphql-go in go.mod" "PASS"
else
    log_test "graphql-go in go.mod" "FAIL"
fi

echo ""
echo "=============================================="
echo "  Results: $PASSED/$TOTAL tests passed"
echo "=============================================="

if [ $FAILED -gt 0 ]; then
    echo -e "\e[31m$FAILED test(s) failed\e[0m"
    exit 1
else
    echo -e "\e[32mAll tests passed!\e[0m"
    exit 0
fi
