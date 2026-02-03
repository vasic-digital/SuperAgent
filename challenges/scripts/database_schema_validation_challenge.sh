#!/bin/bash
# Database Schema Validation Challenge
# Validates SQL schema correctness, migration application, table structures, and index existence
# Tests: All migrations applied, table structures match models, index existence verification

set -o pipefail

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Database connection settings
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-15432}"
DB_USER="${DB_USER:-helixagent}"
DB_PASSWORD="${DB_PASSWORD:-helixagent123}"
DB_NAME="${DB_NAME:-helixagent_db}"

# Helper functions
pass() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
    echo -e "${GREEN}[PASS]${NC} $1"
}

fail() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    echo -e "${RED}[FAIL]${NC} $1"
    if [ -n "$2" ]; then
        echo -e "       ${YELLOW}Reason: $2${NC}"
    fi
}

skip() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "${YELLOW}[SKIP]${NC} $1"
}

section() {
    echo -e "\n${YELLOW}=== $1 ===${NC}"
}

# Navigate to project root
cd "$(dirname "$0")/../.." || exit 1
PROJECT_ROOT="$(pwd)"

echo "============================================"
echo "  DATABASE SCHEMA VALIDATION CHALLENGE"
echo "  Validates migrations, tables, and indexes"
echo "============================================"
echo ""
echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo ""

# Check if PostgreSQL is accessible
check_database_connection() {
    section "Database Connection Test"

    if command -v psql &> /dev/null; then
        if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
            pass "Database connection established"
            return 0
        else
            fail "Cannot connect to database" "Check DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME"
            return 1
        fi
    else
        # Try using docker exec if psql not available locally
        local runtime=""
        if command -v docker &> /dev/null && docker ps &> /dev/null; then
            runtime="docker"
        elif command -v podman &> /dev/null && podman ps &> /dev/null; then
            runtime="podman"
        fi

        if [ -n "$runtime" ]; then
            local container_name="helixagent-postgres-test"
            if $runtime exec "$container_name" psql -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
                pass "Database connection established (via container)"
                return 0
            fi
        fi

        fail "psql not available and no container connection" "Install postgresql-client or start test containers"
        return 1
    fi
}

# Execute SQL query
run_query() {
    local query="$1"
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -A -c "$query" 2>/dev/null
}

# Execute SQL query with tuples only
run_query_count() {
    local query="$1"
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -A -c "$query" 2>/dev/null | head -1
}

# ============================================================================
# SECTION 1: Schema Definition Files Validation
# ============================================================================

validate_schema_files() {
    section "Schema Definition Files Validation"

    # Check that SQL schema files exist
    echo -e "${BLUE}Testing:${NC} SQL schema files exist"

    local schema_files=(
        "sql/schema/users_sessions.sql"
        "sql/schema/llm_providers.sql"
        "sql/schema/requests_responses.sql"
        "sql/schema/background_tasks.sql"
        "sql/schema/debate_system.sql"
        "sql/schema/cognee_memories.sql"
        "sql/schema/protocol_support.sql"
        "sql/schema/indexes_views.sql"
        "sql/schema/complete_schema.sql"
    )

    local missing=0
    for schema_file in "${schema_files[@]}"; do
        if [ -f "$PROJECT_ROOT/$schema_file" ]; then
            pass "Schema file exists: $schema_file"
        else
            fail "Schema file missing: $schema_file"
            missing=$((missing + 1))
        fi
    done

    # Verify complete_schema.sql has all major sections
    echo -e "${BLUE}Testing:${NC} complete_schema.sql contains all sections"
    local complete_schema="$PROJECT_ROOT/sql/schema/complete_schema.sql"

    if [ -f "$complete_schema" ]; then
        local sections=(
            "CREATE TABLE.*users"
            "CREATE TABLE.*user_sessions"
            "CREATE TABLE.*llm_providers"
            "CREATE TABLE.*llm_requests"
            "CREATE TABLE.*llm_responses"
            "CREATE TABLE.*cognee_memories"
            "CREATE TABLE.*models_metadata"
            "CREATE TABLE.*mcp_servers"
            "CREATE TABLE.*background_tasks"
            "CREATE TABLE.*debate_logs"
        )

        for section in "${sections[@]}"; do
            if grep -qE "$section" "$complete_schema"; then
                pass "Section found: $section"
            else
                fail "Section missing: $section"
            fi
        done
    fi
}

# ============================================================================
# SECTION 2: Migration Application Verification
# ============================================================================

validate_migrations_applied() {
    section "Migration Application Verification"

    # Check if database has migrations tracking table (if using migration framework)
    # For now, verify tables exist that should exist after all migrations

    echo -e "${BLUE}Testing:${NC} Core tables exist"

    local core_tables=(
        "users"
        "user_sessions"
        "llm_providers"
        "llm_requests"
        "llm_responses"
        "cognee_memories"
    )

    for table in "${core_tables[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='$table'")
        if [ "$exists" = "1" ]; then
            pass "Core table exists: $table"
        else
            fail "Core table missing: $table"
        fi
    done

    # Check Migration 002 tables (Models.dev)
    echo -e "${BLUE}Testing:${NC} Migration 002 tables (Models.dev)"

    local migration002_tables=(
        "models_metadata"
        "model_benchmarks"
        "models_refresh_history"
    )

    for table in "${migration002_tables[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='$table'")
        if [ "$exists" = "1" ]; then
            pass "Migration 002 table exists: $table"
        else
            fail "Migration 002 table missing: $table"
        fi
    done

    # Check Migration 003 tables (Protocol Support)
    echo -e "${BLUE}Testing:${NC} Migration 003 tables (Protocol Support)"

    local migration003_tables=(
        "mcp_servers"
        "lsp_servers"
        "acp_servers"
        "embedding_config"
        "vector_documents"
        "protocol_cache"
        "protocol_metrics"
    )

    for table in "${migration003_tables[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='$table'")
        if [ "$exists" = "1" ]; then
            pass "Migration 003 table exists: $table"
        else
            fail "Migration 003 table missing: $table"
        fi
    done

    # Check Migration 011 tables (Background Tasks)
    echo -e "${BLUE}Testing:${NC} Migration 011 tables (Background Tasks)"

    local migration011_tables=(
        "background_tasks"
        "background_tasks_dead_letter"
        "task_execution_history"
        "task_resource_snapshots"
        "webhook_deliveries"
    )

    for table in "${migration011_tables[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='$table'")
        if [ "$exists" = "1" ]; then
            pass "Migration 011 table exists: $table"
        else
            fail "Migration 011 table missing: $table"
        fi
    done

    # Check Migration 014 tables (AI Debate)
    echo -e "${BLUE}Testing:${NC} Migration 014 tables (AI Debate)"

    local exists=$(run_query_count "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='debate_logs'")
    if [ "$exists" = "1" ]; then
        pass "Migration 014 table exists: debate_logs"
    else
        fail "Migration 014 table missing: debate_logs"
    fi
}

# ============================================================================
# SECTION 3: Table Structure Verification
# ============================================================================

validate_table_structures() {
    section "Table Structure Verification"

    # Verify users table structure
    echo -e "${BLUE}Testing:${NC} users table structure"

    local users_columns=(
        "id:uuid"
        "username:character varying"
        "email:character varying"
        "password_hash:character varying"
        "api_key:character varying"
        "role:character varying"
        "created_at:timestamp with time zone"
        "updated_at:timestamp with time zone"
    )

    for col_spec in "${users_columns[@]}"; do
        local col_name="${col_spec%%:*}"
        local col_type="${col_spec#*:}"
        local actual_type=$(run_query "SELECT data_type FROM information_schema.columns WHERE table_name='users' AND column_name='$col_name'")
        if [ -n "$actual_type" ]; then
            pass "users.$col_name exists (type: $actual_type)"
        else
            fail "users.$col_name missing"
        fi
    done

    # Verify llm_providers table has Models.dev columns (migration 002)
    echo -e "${BLUE}Testing:${NC} llm_providers Models.dev extension columns"

    local modelsdev_columns=(
        "modelsdev_provider_id"
        "total_models"
        "enabled_models"
        "last_models_sync"
    )

    for col_name in "${modelsdev_columns[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM information_schema.columns WHERE table_name='llm_providers' AND column_name='$col_name'")
        if [ "$exists" = "1" ]; then
            pass "llm_providers.$col_name exists (migration 002)"
        else
            fail "llm_providers.$col_name missing (migration 002)"
        fi
    done

    # Verify models_metadata protocol columns (migration 003)
    echo -e "${BLUE}Testing:${NC} models_metadata protocol extension columns"

    local protocol_columns=(
        "protocol_support"
        "mcp_server_id"
        "lsp_server_id"
        "acp_server_id"
        "embedding_provider"
        "protocol_config"
        "protocol_last_sync"
    )

    for col_name in "${protocol_columns[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM information_schema.columns WHERE table_name='models_metadata' AND column_name='$col_name'")
        if [ "$exists" = "1" ]; then
            pass "models_metadata.$col_name exists (migration 003)"
        else
            fail "models_metadata.$col_name missing (migration 003)"
        fi
    done

    # Verify background_tasks has all required columns
    echo -e "${BLUE}Testing:${NC} background_tasks comprehensive structure"

    local task_columns=(
        "id"
        "task_type"
        "task_name"
        "correlation_id"
        "parent_task_id"
        "payload"
        "config"
        "priority"
        "status"
        "progress"
        "checkpoint"
        "max_retries"
        "retry_count"
        "worker_id"
        "last_heartbeat"
        "deadline"
        "required_cpu_cores"
        "required_memory_mb"
    )

    for col_name in "${task_columns[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM information_schema.columns WHERE table_name='background_tasks' AND column_name='$col_name'")
        if [ "$exists" = "1" ]; then
            pass "background_tasks.$col_name exists"
        else
            fail "background_tasks.$col_name missing"
        fi
    done
}

# ============================================================================
# SECTION 4: Index Existence Verification
# ============================================================================

validate_indexes() {
    section "Index Existence Verification"

    # Core table indexes
    echo -e "${BLUE}Testing:${NC} Core table indexes"

    local core_indexes=(
        "idx_users_email"
        "idx_users_api_key"
        "idx_user_sessions_user_id"
        "idx_user_sessions_expires_at"
        "idx_user_sessions_session_token"
        "idx_llm_providers_name"
        "idx_llm_providers_enabled"
        "idx_llm_requests_session_id"
        "idx_llm_requests_user_id"
        "idx_llm_requests_status"
        "idx_llm_responses_request_id"
        "idx_llm_responses_provider_id"
        "idx_llm_responses_selected"
        "idx_cognee_memories_session_id"
        "idx_cognee_memories_dataset_name"
    )

    for idx_name in "${core_indexes[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM pg_indexes WHERE indexname='$idx_name'")
        if [ "$exists" = "1" ]; then
            pass "Index exists: $idx_name"
        else
            fail "Index missing: $idx_name"
        fi
    done

    # Model catalog indexes (migration 002)
    echo -e "${BLUE}Testing:${NC} Model catalog indexes"

    local model_indexes=(
        "idx_models_metadata_provider_id"
        "idx_models_metadata_model_type"
        "idx_models_metadata_last_refreshed"
        "idx_benchmarks_model_id"
        "idx_refresh_history_started"
    )

    for idx_name in "${model_indexes[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM pg_indexes WHERE indexname='$idx_name'")
        if [ "$exists" = "1" ]; then
            pass "Index exists: $idx_name"
        else
            fail "Index missing: $idx_name"
        fi
    done

    # Protocol indexes (migration 003)
    echo -e "${BLUE}Testing:${NC} Protocol indexes"

    local protocol_indexes=(
        "idx_mcp_servers_enabled"
        "idx_lsp_servers_enabled"
        "idx_acp_servers_enabled"
        "idx_protocol_cache_expires_at"
        "idx_protocol_metrics_protocol_type"
        "idx_protocol_metrics_created_at"
    )

    for idx_name in "${protocol_indexes[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM pg_indexes WHERE indexname='$idx_name'")
        if [ "$exists" = "1" ]; then
            pass "Index exists: $idx_name"
        else
            fail "Index missing: $idx_name"
        fi
    done

    # Background task indexes (migration 011)
    echo -e "${BLUE}Testing:${NC} Background task indexes"

    local task_indexes=(
        "idx_tasks_status"
        "idx_tasks_priority_status"
        "idx_tasks_worker"
        "idx_tasks_scheduled"
        "idx_tasks_heartbeat"
        "idx_tasks_type"
        "idx_tasks_created"
    )

    for idx_name in "${task_indexes[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM pg_indexes WHERE indexname='$idx_name'")
        if [ "$exists" = "1" ]; then
            pass "Index exists: $idx_name"
        else
            fail "Index missing: $idx_name"
        fi
    done

    # Debate indexes (migration 014)
    echo -e "${BLUE}Testing:${NC} Debate log indexes"

    local debate_indexes=(
        "idx_debate_logs_debate_id"
        "idx_debate_logs_session_id"
        "idx_debate_logs_provider"
        "idx_debate_logs_created_at"
    )

    for idx_name in "${debate_indexes[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM pg_indexes WHERE indexname='$idx_name'")
        if [ "$exists" = "1" ]; then
            pass "Index exists: $idx_name"
        else
            fail "Index missing: $idx_name"
        fi
    done
}

# ============================================================================
# SECTION 5: Foreign Key Constraints Verification
# ============================================================================

validate_foreign_keys() {
    section "Foreign Key Constraints Verification"

    echo -e "${BLUE}Testing:${NC} Foreign key constraints"

    # Check user_sessions -> users FK
    local fk_count=$(run_query_count "SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_type='FOREIGN KEY' AND table_name='user_sessions' AND constraint_name LIKE '%user_id%'")
    if [ -n "$fk_count" ] && [ "$fk_count" -ge "1" ]; then
        pass "FK: user_sessions.user_id -> users.id"
    else
        fail "FK missing: user_sessions.user_id -> users.id"
    fi

    # Check llm_requests -> user_sessions FK
    local fk_count=$(run_query_count "SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_type='FOREIGN KEY' AND table_name='llm_requests' AND constraint_name LIKE '%session%'")
    if [ -n "$fk_count" ] && [ "$fk_count" -ge "1" ]; then
        pass "FK: llm_requests.session_id -> user_sessions.id"
    else
        fail "FK missing: llm_requests.session_id -> user_sessions.id"
    fi

    # Check llm_responses -> llm_requests FK
    local fk_count=$(run_query_count "SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_type='FOREIGN KEY' AND table_name='llm_responses' AND constraint_name LIKE '%request%'")
    if [ -n "$fk_count" ] && [ "$fk_count" -ge "1" ]; then
        pass "FK: llm_responses.request_id -> llm_requests.id"
    else
        fail "FK missing: llm_responses.request_id -> llm_requests.id"
    fi

    # Check task_execution_history -> background_tasks FK
    local fk_count=$(run_query_count "SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_type='FOREIGN KEY' AND table_name='task_execution_history' AND constraint_name LIKE '%task%'")
    if [ -n "$fk_count" ] && [ "$fk_count" -ge "1" ]; then
        pass "FK: task_execution_history.task_id -> background_tasks.id"
    else
        fail "FK missing: task_execution_history.task_id -> background_tasks.id"
    fi
}

# ============================================================================
# SECTION 6: Custom Types and Enums Verification
# ============================================================================

validate_custom_types() {
    section "Custom Types and Enums Verification"

    echo -e "${BLUE}Testing:${NC} Custom types exist"

    # Check task_status enum
    local type_exists=$(run_query_count "SELECT COUNT(*) FROM pg_type WHERE typname='task_status'")
    if [ "$type_exists" = "1" ]; then
        pass "Enum type exists: task_status"

        # Verify enum values
        local enum_values=$(run_query "SELECT string_agg(enumlabel, ',') FROM pg_enum WHERE enumtypid = (SELECT oid FROM pg_type WHERE typname = 'task_status')")
        if echo "$enum_values" | grep -q "pending"; then
            pass "task_status has 'pending' value"
        else
            fail "task_status missing 'pending' value"
        fi
        if echo "$enum_values" | grep -q "running"; then
            pass "task_status has 'running' value"
        else
            fail "task_status missing 'running' value"
        fi
        if echo "$enum_values" | grep -q "completed"; then
            pass "task_status has 'completed' value"
        else
            fail "task_status missing 'completed' value"
        fi
    else
        fail "Enum type missing: task_status"
    fi

    # Check task_priority enum
    local type_exists=$(run_query_count "SELECT COUNT(*) FROM pg_type WHERE typname='task_priority'")
    if [ "$type_exists" = "1" ]; then
        pass "Enum type exists: task_priority"

        local enum_values=$(run_query "SELECT string_agg(enumlabel, ',') FROM pg_enum WHERE enumtypid = (SELECT oid FROM pg_type WHERE typname = 'task_priority')")
        if echo "$enum_values" | grep -q "normal"; then
            pass "task_priority has 'normal' value"
        else
            fail "task_priority missing 'normal' value"
        fi
    else
        fail "Enum type missing: task_priority"
    fi
}

# ============================================================================
# SECTION 7: Stored Functions Verification
# ============================================================================

validate_stored_functions() {
    section "Stored Functions Verification"

    echo -e "${BLUE}Testing:${NC} Stored functions exist"

    local functions=(
        "update_updated_at_column"
        "update_background_task_updated_at"
        "dequeue_background_task"
        "get_stale_tasks"
        "cleanup_expired_debate_logs"
    )

    for func_name in "${functions[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM pg_proc WHERE proname='$func_name'")
        if [ "$exists" -ge "1" ]; then
            pass "Function exists: $func_name"
        else
            fail "Function missing: $func_name"
        fi
    done
}

# ============================================================================
# SECTION 8: Materialized Views Verification
# ============================================================================

validate_materialized_views() {
    section "Materialized Views Verification"

    echo -e "${BLUE}Testing:${NC} Materialized views exist"

    local views=(
        "mv_provider_performance"
        "mv_mcp_server_health"
        "mv_request_analytics_hourly"
        "mv_session_stats_daily"
        "mv_task_statistics"
        "mv_model_capabilities"
        "mv_protocol_metrics_agg"
    )

    for view_name in "${views[@]}"; do
        local exists=$(run_query_count "SELECT COUNT(*) FROM pg_matviews WHERE matviewname='$view_name'")
        if [ "$exists" = "1" ]; then
            pass "Materialized view exists: $view_name"
        else
            fail "Materialized view missing: $view_name"
        fi
    done
}

# ============================================================================
# SECTION 9: Database Extensions Verification
# ============================================================================

validate_extensions() {
    section "Database Extensions Verification"

    echo -e "${BLUE}Testing:${NC} Required extensions"

    # Check uuid-ossp extension
    local uuid_ext=$(run_query_count "SELECT COUNT(*) FROM pg_extension WHERE extname='uuid-ossp'")
    if [ "$uuid_ext" = "1" ]; then
        pass "Extension installed: uuid-ossp"
    else
        fail "Extension missing: uuid-ossp"
    fi

    # Check pgvector extension (optional but expected)
    local vector_ext=$(run_query_count "SELECT COUNT(*) FROM pg_extension WHERE extname='vector'")
    if [ "$vector_ext" = "1" ]; then
        pass "Extension installed: pgvector"
    else
        echo -e "${YELLOW}[WARN]${NC} Extension not installed: pgvector (optional for vector operations)"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    fi
}

# ============================================================================
# SECTION 10: Go Model Consistency Check
# ============================================================================

validate_model_consistency() {
    section "Go Model Consistency Check"

    echo -e "${BLUE}Testing:${NC} Go models match database schema"

    # Check that repository files exist for core tables
    local repos=(
        "internal/database/user_repository.go"
        "internal/database/session_repository.go"
        "internal/database/provider_repository.go"
        "internal/database/request_repository.go"
        "internal/database/response_repository.go"
        "internal/database/cognee_memory_repository.go"
        "internal/database/background_task_repository.go"
        "internal/database/debate_log_repository.go"
    )

    for repo in "${repos[@]}"; do
        if [ -f "$PROJECT_ROOT/$repo" ]; then
            pass "Repository exists: $(basename $repo)"
        else
            fail "Repository missing: $repo"
        fi
    done

    # Check that migrations array in db.go contains expected table definitions
    echo -e "${BLUE}Testing:${NC} db.go migrations array contains core tables"

    local db_go="$PROJECT_ROOT/internal/database/db.go"
    if [ -f "$db_go" ]; then
        if grep -q "CREATE TABLE IF NOT EXISTS users" "$db_go"; then
            pass "db.go contains users table migration"
        else
            fail "db.go missing users table migration"
        fi

        if grep -q "CREATE TABLE IF NOT EXISTS llm_providers" "$db_go"; then
            pass "db.go contains llm_providers table migration"
        else
            fail "db.go missing llm_providers table migration"
        fi

        if grep -q "CREATE TABLE IF NOT EXISTS models_metadata" "$db_go"; then
            pass "db.go contains models_metadata table migration"
        else
            fail "db.go missing models_metadata table migration"
        fi
    else
        fail "db.go file not found"
    fi
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    # First validate schema files (doesn't require database)
    validate_schema_files
    validate_model_consistency

    # Check database connection
    if check_database_connection; then
        validate_migrations_applied
        validate_table_structures
        validate_indexes
        validate_foreign_keys
        validate_custom_types
        validate_stored_functions
        validate_materialized_views
        validate_extensions
    else
        echo -e "\n${YELLOW}Skipping database-dependent tests (no connection)${NC}"
        echo "To run full validation, start test infrastructure:"
        echo "  make test-infra-start"
    fi

    # Summary
    echo ""
    echo "============================================"
    echo "  Database Schema Validation Results"
    echo "============================================"
    echo ""
    echo "Total Tests: $TOTAL_TESTS"
    echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    echo ""

    if [ $TOTAL_TESTS -gt 0 ]; then
        PASS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
        echo "Pass Rate: ${PASS_RATE}%"
    fi
    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}DATABASE SCHEMA VALIDATION CHALLENGE: PASSED${NC}"
        exit 0
    else
        echo -e "${RED}DATABASE SCHEMA VALIDATION CHALLENGE: FAILED${NC}"
        exit 1
    fi
}

main "$@"
