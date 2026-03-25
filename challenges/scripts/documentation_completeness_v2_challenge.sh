#!/bin/bash
# Documentation Completeness V2 Challenge
# Validates comprehensive documentation integrity across 41 modules.
#
# Tests:
# 1. All 41 modules listed in MODULES.md (grep count)
# 2. SQL schema index exists (sql/README.md)
# 3. docs/README.md exists and is non-trivial
# 4. No backup files (MODULES.md.backup*)
# 5. Key doc files exist (AUTHENTICATION.md, RATE_LIMITING.md, OPENTELEMETRY.md)
# 6. All module directories exist on disk
# 7. MODULES.md module table integrity

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

PASS=0
FAIL=0
TOTAL=0

pass_test() {
    PASS=$((PASS + 1))
    TOTAL=$((TOTAL + 1))
    echo "  PASS: $1"
}

fail_test() {
    FAIL=$((FAIL + 1))
    TOTAL=$((TOTAL + 1))
    echo "  FAIL: $1"
}

echo "============================================"
echo "Documentation Completeness V2 Challenge"
echo "============================================"
echo ""

# -----------------------------------------------
# Test 1: MODULES.md lists exactly 41 modules
# -----------------------------------------------
echo "--- Test Group 1: Module Count in MODULES.md ---"

MODULES_FILE="$PROJECT_ROOT/docs/MODULES.md"

if [ ! -f "$MODULES_FILE" ]; then
    fail_test "MODULES.md exists at docs/MODULES.md"
else
    pass_test "MODULES.md exists at docs/MODULES.md"

    # Count rows in the module index table (lines starting with | and a number)
    MODULE_COUNT=$(grep -cE '^\| [0-9]+ \|' "$MODULES_FILE" 2>/dev/null || echo "0")
    if [ "$MODULE_COUNT" -eq 41 ]; then
        pass_test "MODULES.md lists exactly 41 modules (found $MODULE_COUNT)"
    else
        fail_test "MODULES.md lists exactly 41 modules (found $MODULE_COUNT)"
    fi

    # Check the total line
    if grep -q "Total: 41 modules" "$MODULES_FILE" 2>/dev/null; then
        pass_test "MODULES.md contains 'Total: 41 modules' summary"
    else
        fail_test "MODULES.md contains 'Total: 41 modules' summary"
    fi
fi

# -----------------------------------------------
# Test 2: SQL schema index exists
# -----------------------------------------------
echo ""
echo "--- Test Group 2: SQL Schema Index ---"

SQL_README="$PROJECT_ROOT/sql/README.md"
if [ -f "$SQL_README" ]; then
    pass_test "sql/README.md exists"
    LINES=$(wc -l < "$SQL_README")
    if [ "$LINES" -ge 5 ]; then
        pass_test "sql/README.md is non-trivial ($LINES lines)"
    else
        fail_test "sql/README.md is non-trivial ($LINES lines < 5 minimum)"
    fi
else
    fail_test "sql/README.md exists"
    fail_test "sql/README.md is non-trivial (file missing)"
fi

# -----------------------------------------------
# Test 3: docs/README.md exists and is non-trivial
# -----------------------------------------------
echo ""
echo "--- Test Group 3: docs/README.md ---"

DOCS_README="$PROJECT_ROOT/docs/README.md"
if [ -f "$DOCS_README" ]; then
    pass_test "docs/README.md exists"
    LINES=$(wc -l < "$DOCS_README")
    if [ "$LINES" -ge 10 ]; then
        pass_test "docs/README.md is non-trivial ($LINES lines)"
    else
        fail_test "docs/README.md is non-trivial ($LINES lines < 10 minimum)"
    fi
else
    fail_test "docs/README.md exists"
    fail_test "docs/README.md is non-trivial (file missing)"
fi

# -----------------------------------------------
# Test 4: No backup files
# -----------------------------------------------
echo ""
echo "--- Test Group 4: No Backup Files ---"

BACKUP_COUNT=$(find "$PROJECT_ROOT/docs" -maxdepth 1 -name "MODULES.md.backup*" 2>/dev/null | wc -l)
if [ "$BACKUP_COUNT" -eq 0 ]; then
    pass_test "No MODULES.md.backup* files in docs/"
else
    fail_test "No MODULES.md.backup* files in docs/ (found $BACKUP_COUNT)"
fi

BACKUP_ROOT_COUNT=$(find "$PROJECT_ROOT" -maxdepth 1 -name "MODULES.md.backup*" 2>/dev/null | wc -l)
if [ "$BACKUP_ROOT_COUNT" -eq 0 ]; then
    pass_test "No MODULES.md.backup* files in project root"
else
    fail_test "No MODULES.md.backup* files in project root (found $BACKUP_ROOT_COUNT)"
fi

# -----------------------------------------------
# Test 5: Key documentation files exist
# -----------------------------------------------
echo ""
echo "--- Test Group 5: Key Documentation Files ---"

KEY_DOCS=(
    "docs/ARCHITECTURE.md"
    "docs/API_DOCUMENTATION.md"
    "docs/MODULES.md"
    "docs/features/README.md"
)

for doc in "${KEY_DOCS[@]}"; do
    if [ -f "$PROJECT_ROOT/$doc" ]; then
        pass_test "$doc exists"
    else
        fail_test "$doc exists"
    fi
done

# Check for authentication/rate-limiting/opentelemetry docs in any location under docs/
AUTH_DOC=$(find "$PROJECT_ROOT/docs" -iname "*authentication*" -name "*.md" 2>/dev/null | head -1)
if [ -n "$AUTH_DOC" ]; then
    pass_test "Authentication documentation exists ($(basename "$AUTH_DOC"))"
else
    fail_test "Authentication documentation exists (no *authentication*.md found under docs/)"
fi

RATE_DOC=$(find "$PROJECT_ROOT/docs" -iname "*rate_limit*" -o -iname "*rate-limit*" -o -iname "*ratelimit*" 2>/dev/null | grep -i "\.md$" | head -1)
if [ -n "$RATE_DOC" ]; then
    pass_test "Rate limiting documentation exists ($(basename "$RATE_DOC"))"
else
    fail_test "Rate limiting documentation exists (no *rate_limit*.md found under docs/)"
fi

OTEL_DOC=$(find "$PROJECT_ROOT/docs" -iname "*opentelemetry*" -o -iname "*otel*" -o -iname "*observability*" 2>/dev/null | grep -i "\.md$" | head -1)
if [ -n "$OTEL_DOC" ]; then
    pass_test "OpenTelemetry/observability documentation exists ($(basename "$OTEL_DOC"))"
else
    fail_test "OpenTelemetry/observability documentation exists (no *opentelemetry*.md or *observability*.md found under docs/)"
fi

# -----------------------------------------------
# Test 6: All module directories exist on disk
# -----------------------------------------------
echo ""
echo "--- Test Group 6: Module Directories Exist ---"

MODULE_DIRS=(
    "EventBus" "Concurrency" "Observability" "Auth" "Storage" "Streaming"
    "Security" "VectorDB" "Embeddings" "Database" "Cache" "LLMProvider"
    "Messaging" "Formatters" "MCP_Module" "BackgroundTasks"
    "RAG" "ConversationContext" "Memory" "Optimization" "Plugins"
    "Agentic" "LLMOps" "SelfImprove" "Planning" "Benchmark"
    "DebateOrchestrator" "HelixMemory" "HelixSpecifier"
    "Containers" "Challenges" "BuildCheck"
    "ToolSchema" "SkillRegistry" "Models"
    "DocProcessor" "HelixQA" "LLMOrchestrator" "VisionEngine"
    "LLMsVerifier" "MCP-Servers"
)

for mod_dir in "${MODULE_DIRS[@]}"; do
    if [ -d "$PROJECT_ROOT/$mod_dir" ]; then
        pass_test "Module directory $mod_dir/ exists"
    else
        fail_test "Module directory $mod_dir/ exists"
    fi
done

# -----------------------------------------------
# Test 7: MODULES.md table integrity
# -----------------------------------------------
echo ""
echo "--- Test Group 7: MODULES.md Table Integrity ---"

# Verify key modules appear in the table
KEY_MODULES=("DocProcessor" "HelixQA" "LLMOrchestrator" "VisionEngine" "LLMsVerifier" "MCP-Servers")
for mod in "${KEY_MODULES[@]}"; do
    if grep -q "$mod" "$MODULES_FILE" 2>/dev/null; then
        pass_test "$mod is listed in MODULES.md"
    else
        fail_test "$mod is listed in MODULES.md"
    fi
done

# -----------------------------------------------
# Summary
# -----------------------------------------------
echo ""
echo "============================================"
echo "Documentation Completeness V2 Challenge Summary"
echo "============================================"
echo "PASSED: $PASS"
echo "FAILED: $FAIL"
echo "TOTAL:  $TOTAL"
echo ""

if [ "$FAIL" -gt 0 ]; then
    echo "RESULT: FAILED ($FAIL failures)"
    exit 1
else
    echo "RESULT: PASSED (all $PASS tests passed)"
    exit 0
fi
