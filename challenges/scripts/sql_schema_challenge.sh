#!/bin/bash
# HelixAgent Challenge: SQL Schema Documentation
# Tests: ~15 - SQL file presence, table coverage, relationships
# Usage: ./challenges/scripts/sql_schema_challenge.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

pass() { PASSED=$((PASSED+1)); TOTAL=$((TOTAL+1)); echo -e "${GREEN}[PASS]${NC} $1"; }
fail() { FAILED=$((FAILED+1)); TOTAL=$((TOTAL+1)); echo -e "${RED}[FAIL]${NC} $1"; }
info() { echo -e "${BLUE}[INFO]${NC} $1"; }

check_file() {
    local desc="$1"
    local file="$2"
    if [ -f "$file" ]; then
        pass "$desc"
    else
        fail "$desc - File not found: $file"
    fi
}

check_contains() {
    local desc="$1"
    local file="$2"
    local pattern="$3"
    if grep -qi "$pattern" "$file" 2>/dev/null; then
        pass "$desc"
    else
        fail "$desc - Pattern '$pattern' not found in $file"
    fi
}

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║          SQL SCHEMA DOCUMENTATION CHALLENGE (~15 tests)        ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

SQL_DIR="$PROJECT_ROOT/sql/schema"

# ============================================================================
# SECTION 1: SQL Files Exist
# ============================================================================
info "Section 1: SQL Schema Files Exist"

check_file "Complete schema exists" "$SQL_DIR/complete_schema.sql"
check_file "Users/sessions SQL exists" "$SQL_DIR/users_sessions.sql"
check_file "LLM providers SQL exists" "$SQL_DIR/llm_providers.sql"
check_file "Requests/responses SQL exists" "$SQL_DIR/requests_responses.sql"
check_file "Background tasks SQL exists" "$SQL_DIR/background_tasks.sql"
check_file "Debate system SQL exists" "$SQL_DIR/debate_system.sql"
check_file "Cognee memories SQL exists" "$SQL_DIR/cognee_memories.sql"
check_file "Protocol support SQL exists" "$SQL_DIR/protocol_support.sql"
check_file "Indexes/views SQL exists" "$SQL_DIR/indexes_views.sql"
check_file "Relationships SQL exists" "$SQL_DIR/relationships.sql"

# ============================================================================
# SECTION 2: Complete Schema Coverage
# ============================================================================
info "Section 2: Schema Coverage"

check_contains "Complete schema has users table" "$SQL_DIR/complete_schema.sql" "CREATE TABLE.*users"
check_contains "Complete schema has llm_providers" "$SQL_DIR/complete_schema.sql" "CREATE TABLE.*llm_providers"
check_contains "Complete schema has background_tasks" "$SQL_DIR/complete_schema.sql" "CREATE TABLE.*background_tasks"
check_contains "Complete schema has debate_logs" "$SQL_DIR/complete_schema.sql" "CREATE TABLE.*debate_logs"
check_contains "Relationships file documents FKs" "$SQL_DIR/relationships.sql" "REFERENCES"

# ============================================================================
# SUMMARY
# ============================================================================
echo ""
echo "════════════════════════════════════════════════════════════════════"
echo -e "  Results: ${GREEN}${PASSED} passed${NC} / ${RED}${FAILED} failed${NC} / ${TOTAL} total"
if [ "$FAILED" -eq 0 ]; then
    echo -e "  Status: ${GREEN}ALL TESTS PASSED${NC}"
else
    echo -e "  Status: ${RED}SOME TESTS FAILED${NC}"
fi
echo "════════════════════════════════════════════════════════════════════"

exit $FAILED
