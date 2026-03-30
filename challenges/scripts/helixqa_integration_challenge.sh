#!/bin/bash
# HelixAgent Challenge - HelixQA Integration
# Validates HelixQA adapter, handler, VisionEngine remote package,
# and API endpoint wiring.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

if [ -f "$SCRIPT_DIR/challenge_framework.sh" ]; then
    source "$SCRIPT_DIR/challenge_framework.sh"
    init_challenge "helixqa-integration" "HelixQA Integration"
    load_env
    FRAMEWORK_LOADED=true
else
    FRAMEWORK_LOADED=false
fi

PASSED=0
FAILED=0

record_result() {
    local name="$1" status="$2"
    if [ "$FRAMEWORK_LOADED" = true ]; then
        if [ "$status" = "PASS" ]; then
            record_assertion "test" "$name" "true" "$name"
        else
            record_assertion "test" "$name" "false" "$name"
        fi
    fi
    if [ "$status" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        echo -e "\033[0;32m[PASS]\033[0m $name"
    else
        FAILED=$((FAILED + 1))
        echo -e "\033[0;31m[FAIL]\033[0m $name"
    fi
}

echo "=== HelixQA Integration Challenge ==="
echo ""

# --- Adapter Structure ---
echo "--- Adapter Structure ---"

# Test 1: Adapter package exists
if [ -d "$PROJECT_ROOT/internal/adapters/helixqa" ]; then
    record_result "Adapter package directory exists" "PASS"
else
    record_result "Adapter package directory exists" "FAIL"
fi

# Test 2: Adapter source file exists
if [ -f "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go" ]; then
    record_result "Adapter source file exists" "PASS"
else
    record_result "Adapter source file exists" "FAIL"
fi

# Test 3: Adapter test file exists
if [ -f "$PROJECT_ROOT/internal/adapters/helixqa/adapter_test.go" ]; then
    record_result "Adapter test file exists" "PASS"
else
    record_result "Adapter test file exists" "FAIL"
fi

# Test 4: Adapter exports New constructor
if grep -q "func New(" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "Adapter exports New constructor" "PASS"
else
    record_result "Adapter exports New constructor" "FAIL"
fi

# Test 5: Adapter exports Initialize method
if grep -q "func (a \*Adapter) Initialize(" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "Adapter exports Initialize method" "PASS"
else
    record_result "Adapter exports Initialize method" "FAIL"
fi

# Test 6: Adapter exports RunAutonomousSession
if grep -q "func (a \*Adapter) RunAutonomousSession(" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "Adapter exports RunAutonomousSession" "PASS"
else
    record_result "Adapter exports RunAutonomousSession" "FAIL"
fi

# Test 7: Adapter exports GetFindings
if grep -q "func (a \*Adapter) GetFindings(" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "Adapter exports GetFindings" "PASS"
else
    record_result "Adapter exports GetFindings" "FAIL"
fi

# Test 8: Adapter exports GetFinding
if grep -q "func (a \*Adapter) GetFinding(" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "Adapter exports GetFinding" "PASS"
else
    record_result "Adapter exports GetFinding" "FAIL"
fi

# Test 9: Adapter exports UpdateFindingStatus
if grep -q "func (a \*Adapter) UpdateFindingStatus(" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "Adapter exports UpdateFindingStatus" "PASS"
else
    record_result "Adapter exports UpdateFindingStatus" "FAIL"
fi

# Test 10: Adapter exports DiscoverCredentials
if grep -q "func (a \*Adapter) DiscoverCredentials(" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "Adapter exports DiscoverCredentials" "PASS"
else
    record_result "Adapter exports DiscoverCredentials" "FAIL"
fi

# Test 11: Adapter exports DiscoverKnowledge
if grep -q "func (a \*Adapter) DiscoverKnowledge(" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "Adapter exports DiscoverKnowledge" "PASS"
else
    record_result "Adapter exports DiscoverKnowledge" "FAIL"
fi

# Test 12: Adapter exports SupportedPlatforms
if grep -q "func (a \*Adapter) SupportedPlatforms(" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "Adapter exports SupportedPlatforms" "PASS"
else
    record_result "Adapter exports SupportedPlatforms" "FAIL"
fi

# Test 13: Adapter exports Close
if grep -q "func (a \*Adapter) Close(" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "Adapter exports Close" "PASS"
else
    record_result "Adapter exports Close" "FAIL"
fi

# --- Handler Structure ---
echo ""
echo "--- Handler Structure ---"

# Test 14: QA handler file exists
if [ -f "$PROJECT_ROOT/internal/handlers/qa_handler.go" ]; then
    record_result "QA handler file exists" "PASS"
else
    record_result "QA handler file exists" "FAIL"
fi

# Test 15: QA handler test file exists
if [ -f "$PROJECT_ROOT/internal/handlers/qa_handler_test.go" ]; then
    record_result "QA handler test file exists" "PASS"
else
    record_result "QA handler test file exists" "FAIL"
fi

# Test 16: QA handler exports NewQAHandler
if grep -q "func NewQAHandler(" "$PROJECT_ROOT/internal/handlers/qa_handler.go"; then
    record_result "Handler exports NewQAHandler" "PASS"
else
    record_result "Handler exports NewQAHandler" "FAIL"
fi

# Test 17: QA handler exports RegisterQARoutes
if grep -q "func RegisterQARoutes(" "$PROJECT_ROOT/internal/handlers/qa_handler.go"; then
    record_result "Handler exports RegisterQARoutes" "PASS"
else
    record_result "Handler exports RegisterQARoutes" "FAIL"
fi

# Test 18: QA handler has StartSession endpoint
if grep -q 'qa.POST("/sessions"' "$PROJECT_ROOT/internal/handlers/qa_handler.go"; then
    record_result "Handler registers POST /sessions" "PASS"
else
    record_result "Handler registers POST /sessions" "FAIL"
fi

# Test 19: QA handler has ListFindings endpoint
if grep -q 'qa.GET("/findings"' "$PROJECT_ROOT/internal/handlers/qa_handler.go"; then
    record_result "Handler registers GET /findings" "PASS"
else
    record_result "Handler registers GET /findings" "FAIL"
fi

# Test 20: QA handler has GetFinding endpoint
if grep -q 'qa.GET("/findings/:id"' "$PROJECT_ROOT/internal/handlers/qa_handler.go"; then
    record_result "Handler registers GET /findings/:id" "PASS"
else
    record_result "Handler registers GET /findings/:id" "FAIL"
fi

# Test 21: QA handler has UpdateFinding endpoint
if grep -q 'qa.PUT("/findings/:id"' "$PROJECT_ROOT/internal/handlers/qa_handler.go"; then
    record_result "Handler registers PUT /findings/:id" "PASS"
else
    record_result "Handler registers PUT /findings/:id" "FAIL"
fi

# Test 22: QA handler has ListPlatforms endpoint
if grep -q 'qa.GET("/platforms"' "$PROJECT_ROOT/internal/handlers/qa_handler.go"; then
    record_result "Handler registers GET /platforms" "PASS"
else
    record_result "Handler registers GET /platforms" "FAIL"
fi

# Test 23: QA handler has DiscoverKnowledge endpoint
if grep -q 'qa.POST("/discover"' "$PROJECT_ROOT/internal/handlers/qa_handler.go"; then
    record_result "Handler registers POST /discover" "PASS"
else
    record_result "Handler registers POST /discover" "FAIL"
fi

# --- Router Wiring ---
echo ""
echo "--- Router Wiring ---"

# Test 24: QA handler wired in router
if grep -q "NewQAHandler" "$PROJECT_ROOT/internal/router/router.go"; then
    record_result "QA handler wired in router" "PASS"
else
    record_result "QA handler wired in router" "FAIL"
fi

# Test 25: QA routes registered in router
if grep -q "RegisterQARoutes" "$PROJECT_ROOT/internal/router/router.go"; then
    record_result "QA routes registered in router" "PASS"
else
    record_result "QA routes registered in router" "FAIL"
fi

# Test 26: QA log message in router
if grep -q "/v1/qa/" "$PROJECT_ROOT/internal/router/router.go"; then
    record_result "QA endpoint log message in router" "PASS"
else
    record_result "QA endpoint log message in router" "FAIL"
fi

# --- VisionEngine Remote Package ---
echo ""
echo "--- VisionEngine Remote Package ---"

# Test 27: Remote package exists
if [ -d "$PROJECT_ROOT/VisionEngine/pkg/remote" ]; then
    record_result "VisionEngine remote package exists" "PASS"
else
    record_result "VisionEngine remote package exists" "FAIL"
fi

# Test 28: Remote package source
if [ -f "$PROJECT_ROOT/VisionEngine/pkg/remote/remote.go" ]; then
    record_result "Remote package source file exists" "PASS"
else
    record_result "Remote package source file exists" "FAIL"
fi

# Test 29: Deployer source
if [ -f "$PROJECT_ROOT/VisionEngine/pkg/remote/deployer.go" ]; then
    record_result "Remote deployer source exists" "PASS"
else
    record_result "Remote deployer source exists" "FAIL"
fi

# Test 30: Remote test file
if [ -f "$PROJECT_ROOT/VisionEngine/pkg/remote/remote_test.go" ]; then
    record_result "Remote package test file exists" "PASS"
else
    record_result "Remote package test file exists" "FAIL"
fi

# Test 31: VisionPool type exported
if grep -q "type VisionPool struct" "$PROJECT_ROOT/VisionEngine/pkg/remote/remote.go"; then
    record_result "VisionPool type exported" "PASS"
else
    record_result "VisionPool type exported" "FAIL"
fi

# Test 32: VisionSlot type exported
if grep -q "type VisionSlot struct" "$PROJECT_ROOT/VisionEngine/pkg/remote/remote.go"; then
    record_result "VisionSlot type exported" "PASS"
else
    record_result "VisionSlot type exported" "FAIL"
fi

# Test 33: LlamaCppDeployer type exported
if grep -q "type LlamaCppDeployer struct" "$PROJECT_ROOT/VisionEngine/pkg/remote/deployer.go"; then
    record_result "LlamaCppDeployer type exported" "PASS"
else
    record_result "LlamaCppDeployer type exported" "FAIL"
fi

# Test 34: Backend constants defined
if grep -q 'BackendLlamaCpp' "$PROJECT_ROOT/VisionEngine/pkg/remote/remote.go"; then
    record_result "BackendLlamaCpp constant defined" "PASS"
else
    record_result "BackendLlamaCpp constant defined" "FAIL"
fi

# --- Compilation ---
echo ""
echo "--- Compilation ---"

# Test 35: VisionEngine compiles
if (cd "$PROJECT_ROOT/VisionEngine" && go build ./... 2>/dev/null); then
    record_result "VisionEngine compiles" "PASS"
else
    record_result "VisionEngine compiles" "FAIL"
fi

# Test 36: HelixQA compiles
if (cd "$PROJECT_ROOT/HelixQA" && go build ./... 2>/dev/null); then
    record_result "HelixQA compiles" "PASS"
else
    record_result "HelixQA compiles" "FAIL"
fi

# Test 37: Adapter compiles
if (cd "$PROJECT_ROOT" && go build ./internal/adapters/helixqa/ 2>/dev/null); then
    record_result "HelixQA adapter compiles" "PASS"
else
    record_result "HelixQA adapter compiles" "FAIL"
fi

# Test 38: Handler compiles
if (cd "$PROJECT_ROOT" && go build ./internal/handlers/ 2>/dev/null); then
    record_result "QA handler compiles" "PASS"
else
    record_result "QA handler compiles" "FAIL"
fi

# Test 39: Router compiles
if (cd "$PROJECT_ROOT" && go build ./internal/router/ 2>/dev/null); then
    record_result "Router compiles with QA wiring" "PASS"
else
    record_result "Router compiles with QA wiring" "FAIL"
fi

# --- Unit Tests ---
echo ""
echo "--- Unit Tests ---"

# Test 40: VisionEngine remote tests pass
if (cd "$PROJECT_ROOT/VisionEngine" && go test ./pkg/remote/ -count=1 -race 2>/dev/null); then
    record_result "VisionEngine remote tests pass" "PASS"
else
    record_result "VisionEngine remote tests pass" "FAIL"
fi

# Test 41: Adapter tests pass
if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 go test -count=1 -short ./internal/adapters/helixqa/ 2>/dev/null); then
    record_result "HelixQA adapter tests pass" "PASS"
else
    record_result "HelixQA adapter tests pass" "FAIL"
fi

# Test 42: Handler tests pass
if (cd "$PROJECT_ROOT" && GOMAXPROCS=2 go test -count=1 -short ./internal/handlers/ -run TestQA 2>/dev/null); then
    record_result "QA handler tests pass" "PASS"
else
    record_result "QA handler tests pass" "FAIL"
fi

# --- Type Mapping ---
echo ""
echo "--- Type Mapping ---"

# Test 43: SessionResult type exported
if grep -q "type SessionResult struct" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "SessionResult type exported" "PASS"
else
    record_result "SessionResult type exported" "FAIL"
fi

# Test 44: Finding type exported
if grep -q "type Finding struct" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "Finding type exported" "PASS"
else
    record_result "Finding type exported" "FAIL"
fi

# Test 45: SessionConfig type exported
if grep -q "type SessionConfig struct" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "SessionConfig type exported" "PASS"
else
    record_result "SessionConfig type exported" "FAIL"
fi

# Test 46: SessionStatus constants
if grep -q "StatusPending" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go" && \
   grep -q "StatusRunning" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go" && \
   grep -q "StatusCompleted" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go" && \
   grep -q "StatusFailed" "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "SessionStatus constants defined" "PASS"
else
    record_result "SessionStatus constants defined" "FAIL"
fi

# --- Integration ---
echo ""
echo "--- Integration ---"

# Test 47: Adapter imports digital.vasic.helixqa
if grep -q 'digital.vasic.helixqa' "$PROJECT_ROOT/internal/adapters/helixqa/adapter.go"; then
    record_result "Adapter imports digital.vasic.helixqa" "PASS"
else
    record_result "Adapter imports digital.vasic.helixqa" "FAIL"
fi

# Test 48: Handler imports adapter
if grep -q 'dev.helix.agent/internal/adapters/helixqa' "$PROJECT_ROOT/internal/handlers/qa_handler.go"; then
    record_result "Handler imports adapter" "PASS"
else
    record_result "Handler imports adapter" "FAIL"
fi

# Test 49: go.mod has helixqa replace directive
if grep -q 'digital.vasic.helixqa' "$PROJECT_ROOT/go.mod"; then
    record_result "go.mod has helixqa dependency" "PASS"
else
    record_result "go.mod has helixqa dependency" "FAIL"
fi

# Test 50: HelixQA submodule on main branch
HELIXQA_BRANCH=$(cd "$PROJECT_ROOT/HelixQA" && git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
if [ "$HELIXQA_BRANCH" = "main" ]; then
    record_result "HelixQA submodule on main branch" "PASS"
else
    record_result "HelixQA submodule on main branch (got: $HELIXQA_BRANCH)" "FAIL"
fi

echo ""
echo "=== Results ==="
TOTAL=$((PASSED + FAILED))
echo "Passed: $PASSED/$TOTAL"
echo "Failed: $FAILED/$TOTAL"

if [ "$FRAMEWORK_LOADED" = true ]; then
    finalize_challenge "$PASSED" "$TOTAL"
fi

if [ "$FAILED" -gt 0 ]; then
    exit 1
fi
