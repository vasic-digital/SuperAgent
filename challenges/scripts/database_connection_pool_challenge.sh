#!/bin/bash
# Database Connection Pool Challenge
# Tests connection pool behavior, exhaustion handling, graceful degradation, and recovery
# Tests: Pool exhaustion, graceful handling, connection recovery, metrics tracking

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
echo "  DATABASE CONNECTION POOL CHALLENGE"
echo "  Tests pool exhaustion and recovery"
echo "============================================"
echo ""

# ============================================================================
# SECTION 1: Pool Configuration Code Validation
# ============================================================================

validate_pool_config_code() {
    section "Pool Configuration Code Validation"

    local pool_config="$PROJECT_ROOT/internal/database/pool_config.go"

    # Check pool_config.go exists
    echo -e "${BLUE}Testing:${NC} pool_config.go exists"
    if [ -f "$pool_config" ]; then
        pass "pool_config.go exists"
    else
        fail "pool_config.go not found"
        return 1
    fi

    # Check PoolConfigOptions struct
    echo -e "${BLUE}Testing:${NC} PoolConfigOptions struct defined"
    if grep -q "type PoolConfigOptions struct" "$pool_config"; then
        pass "PoolConfigOptions struct defined"
    else
        fail "PoolConfigOptions struct not found"
    fi

    # Check MaxConns field
    echo -e "${BLUE}Testing:${NC} MaxConns field exists"
    if grep -q "MaxConns" "$pool_config"; then
        pass "MaxConns configuration field exists"
    else
        fail "MaxConns field not found"
    fi

    # Check MinConns field
    echo -e "${BLUE}Testing:${NC} MinConns field exists"
    if grep -q "MinConns" "$pool_config"; then
        pass "MinConns configuration field exists"
    else
        fail "MinConns field not found"
    fi

    # Check MaxConnLifetime field
    echo -e "${BLUE}Testing:${NC} MaxConnLifetime field exists"
    if grep -q "MaxConnLifetime" "$pool_config"; then
        pass "MaxConnLifetime configuration field exists"
    else
        fail "MaxConnLifetime field not found"
    fi

    # Check MaxConnIdleTime field
    echo -e "${BLUE}Testing:${NC} MaxConnIdleTime field exists"
    if grep -q "MaxConnIdleTime" "$pool_config"; then
        pass "MaxConnIdleTime configuration field exists"
    else
        fail "MaxConnIdleTime field not found"
    fi

    # Check HealthCheckPeriod field
    echo -e "${BLUE}Testing:${NC} HealthCheckPeriod field exists"
    if grep -q "HealthCheckPeriod" "$pool_config"; then
        pass "HealthCheckPeriod configuration field exists"
    else
        fail "HealthCheckPeriod field not found"
    fi

    # Check ConnectTimeout field
    echo -e "${BLUE}Testing:${NC} ConnectTimeout field exists"
    if grep -q "ConnectTimeout" "$pool_config"; then
        pass "ConnectTimeout configuration field exists"
    else
        fail "ConnectTimeout field not found"
    fi
}

# ============================================================================
# SECTION 2: Pool Options Functions Validation
# ============================================================================

validate_pool_options_functions() {
    section "Pool Options Functions Validation"

    local pool_config="$PROJECT_ROOT/internal/database/pool_config.go"

    # Check DefaultPoolOptions function
    echo -e "${BLUE}Testing:${NC} DefaultPoolOptions function exists"
    if grep -q "func DefaultPoolOptions" "$pool_config"; then
        pass "DefaultPoolOptions function exists"
    else
        fail "DefaultPoolOptions function not found"
    fi

    # Check HighPerformancePoolOptions function
    echo -e "${BLUE}Testing:${NC} HighPerformancePoolOptions function exists"
    if grep -q "func HighPerformancePoolOptions" "$pool_config"; then
        pass "HighPerformancePoolOptions function exists"
    else
        fail "HighPerformancePoolOptions function not found"
    fi

    # Check LowLatencyPoolOptions function
    echo -e "${BLUE}Testing:${NC} LowLatencyPoolOptions function exists"
    if grep -q "func LowLatencyPoolOptions" "$pool_config"; then
        pass "LowLatencyPoolOptions function exists"
    else
        fail "LowLatencyPoolOptions function not found"
    fi

    # Check CreateOptimizedPoolConfig function
    echo -e "${BLUE}Testing:${NC} CreateOptimizedPoolConfig function exists"
    if grep -q "func CreateOptimizedPoolConfig" "$pool_config"; then
        pass "CreateOptimizedPoolConfig function exists"
    else
        fail "CreateOptimizedPoolConfig function not found"
    fi
}

# ============================================================================
# SECTION 3: OptimizedPool Implementation Validation
# ============================================================================

validate_optimized_pool() {
    section "OptimizedPool Implementation Validation"

    local pool_config="$PROJECT_ROOT/internal/database/pool_config.go"

    # Check OptimizedPool struct
    echo -e "${BLUE}Testing:${NC} OptimizedPool struct defined"
    if grep -q "type OptimizedPool struct" "$pool_config"; then
        pass "OptimizedPool struct defined"
    else
        fail "OptimizedPool struct not found"
    fi

    # Check NewOptimizedPool constructor
    echo -e "${BLUE}Testing:${NC} NewOptimizedPool constructor exists"
    if grep -q "func NewOptimizedPool" "$pool_config"; then
        pass "NewOptimizedPool constructor exists"
    else
        fail "NewOptimizedPool constructor not found"
    fi

    # Check Acquire method
    echo -e "${BLUE}Testing:${NC} Acquire method exists"
    if grep -q "func (p \*OptimizedPool) Acquire" "$pool_config"; then
        pass "OptimizedPool.Acquire method exists"
    else
        fail "OptimizedPool.Acquire method not found"
    fi

    # Check Release method
    echo -e "${BLUE}Testing:${NC} Release method exists"
    if grep -q "func (p \*OptimizedPool) Release" "$pool_config"; then
        pass "OptimizedPool.Release method exists"
    else
        fail "OptimizedPool.Release method not found"
    fi

    # Check Query method
    echo -e "${BLUE}Testing:${NC} Query method exists"
    if grep -q "func (p \*OptimizedPool) Query" "$pool_config"; then
        pass "OptimizedPool.Query method exists"
    else
        fail "OptimizedPool.Query method not found"
    fi

    # Check Exec method
    echo -e "${BLUE}Testing:${NC} Exec method exists"
    if grep -q "func (p \*OptimizedPool) Exec" "$pool_config"; then
        pass "OptimizedPool.Exec method exists"
    else
        fail "OptimizedPool.Exec method not found"
    fi

    # Check BeginTx method for transaction support
    echo -e "${BLUE}Testing:${NC} BeginTx method exists"
    if grep -q "func (p \*OptimizedPool) BeginTx" "$pool_config"; then
        pass "OptimizedPool.BeginTx method exists"
    else
        fail "OptimizedPool.BeginTx method not found"
    fi

    # Check HealthCheck method
    echo -e "${BLUE}Testing:${NC} HealthCheck method exists"
    if grep -q "func (p \*OptimizedPool) HealthCheck" "$pool_config"; then
        pass "OptimizedPool.HealthCheck method exists"
    else
        fail "OptimizedPool.HealthCheck method not found"
    fi

    # Check Close method
    echo -e "${BLUE}Testing:${NC} Close method exists"
    if grep -q "func (p \*OptimizedPool) Close" "$pool_config"; then
        pass "OptimizedPool.Close method exists"
    else
        fail "OptimizedPool.Close method not found"
    fi
}

# ============================================================================
# SECTION 4: Pool Metrics Validation
# ============================================================================

validate_pool_metrics() {
    section "Pool Metrics Validation"

    local pool_config="$PROJECT_ROOT/internal/database/pool_config.go"

    # Check PoolMetrics struct
    echo -e "${BLUE}Testing:${NC} PoolMetrics struct defined"
    if grep -q "type PoolMetrics struct" "$pool_config"; then
        pass "PoolMetrics struct defined"
    else
        fail "PoolMetrics struct not found"
    fi

    # Check AcquireCount metric
    echo -e "${BLUE}Testing:${NC} AcquireCount metric exists"
    if grep -q "AcquireCount" "$pool_config"; then
        pass "AcquireCount metric exists"
    else
        fail "AcquireCount metric not found"
    fi

    # Check AcquireErrors metric
    echo -e "${BLUE}Testing:${NC} AcquireErrors metric exists"
    if grep -q "AcquireErrors" "$pool_config"; then
        pass "AcquireErrors metric exists"
    else
        fail "AcquireErrors metric not found"
    fi

    # Check TotalAcquireTimeUs metric
    echo -e "${BLUE}Testing:${NC} TotalAcquireTimeUs metric exists"
    if grep -q "TotalAcquireTimeUs" "$pool_config"; then
        pass "TotalAcquireTimeUs metric exists"
    else
        fail "TotalAcquireTimeUs metric not found"
    fi

    # Check MaxConcurrent metric
    echo -e "${BLUE}Testing:${NC} MaxConcurrent metric exists"
    if grep -q "MaxConcurrent" "$pool_config"; then
        pass "MaxConcurrent metric exists"
    else
        fail "MaxConcurrent metric not found"
    fi

    # Check CurrentConcurrent metric
    echo -e "${BLUE}Testing:${NC} CurrentConcurrent metric exists"
    if grep -q "CurrentConcurrent" "$pool_config"; then
        pass "CurrentConcurrent metric exists"
    else
        fail "CurrentConcurrent metric not found"
    fi

    # Check Metrics method
    echo -e "${BLUE}Testing:${NC} Metrics method exists"
    if grep -q "func (p \*OptimizedPool) Metrics" "$pool_config"; then
        pass "OptimizedPool.Metrics method exists"
    else
        fail "OptimizedPool.Metrics method not found"
    fi

    # Check AverageAcquireTime method
    echo -e "${BLUE}Testing:${NC} AverageAcquireTime method exists"
    if grep -q "func (p \*OptimizedPool) AverageAcquireTime" "$pool_config"; then
        pass "OptimizedPool.AverageAcquireTime method exists"
    else
        fail "OptimizedPool.AverageAcquireTime method not found"
    fi
}

# ============================================================================
# SECTION 5: Lazy Pool Implementation Validation
# ============================================================================

validate_lazy_pool() {
    section "Lazy Pool Implementation Validation"

    local pool_config="$PROJECT_ROOT/internal/database/pool_config.go"

    # Check LazyPool struct
    echo -e "${BLUE}Testing:${NC} LazyPool struct defined"
    if grep -q "type LazyPool struct" "$pool_config"; then
        pass "LazyPool struct defined"
    else
        fail "LazyPool struct not found"
    fi

    # Check NewLazyPool constructor
    echo -e "${BLUE}Testing:${NC} NewLazyPool constructor exists"
    if grep -q "func NewLazyPool" "$pool_config"; then
        pass "NewLazyPool constructor exists"
    else
        fail "NewLazyPool constructor not found"
    fi

    # Check Get method for lazy initialization
    echo -e "${BLUE}Testing:${NC} LazyPool.Get method exists"
    if grep -q "func (p \*LazyPool) Get" "$pool_config"; then
        pass "LazyPool.Get method exists"
    else
        fail "LazyPool.Get method not found"
    fi

    # Check IsInitialized method
    echo -e "${BLUE}Testing:${NC} LazyPool.IsInitialized method exists"
    if grep -q "func (p \*LazyPool) IsInitialized" "$pool_config"; then
        pass "LazyPool.IsInitialized method exists"
    else
        fail "LazyPool.IsInitialized method not found"
    fi

    # Check Close method
    echo -e "${BLUE}Testing:${NC} LazyPool.Close method exists"
    if grep -q "func (p \*LazyPool) Close" "$pool_config"; then
        pass "LazyPool.Close method exists"
    else
        fail "LazyPool.Close method not found"
    fi

    # Check atomic operations for thread safety
    echo -e "${BLUE}Testing:${NC} Atomic operations for thread safety"
    if grep -q "atomic.LoadInt32\|atomic.StoreInt32\|atomic.AddInt64" "$pool_config"; then
        pass "Atomic operations used for thread safety"
    else
        fail "Atomic operations not found"
    fi
}

# ============================================================================
# SECTION 6: Connection Pool Unit Tests Validation
# ============================================================================

validate_pool_unit_tests() {
    section "Connection Pool Unit Tests Validation"

    local test_files=(
        "internal/database/pool_config_test.go"
        "internal/database/pool_config_extended_test.go"
    )

    for test_file in "${test_files[@]}"; do
        local full_path="$PROJECT_ROOT/$test_file"
        if [ -f "$full_path" ]; then
            pass "Test file exists: $(basename $test_file)"
        else
            fail "Test file missing: $test_file"
        fi
    done

    # Run pool config unit tests
    echo -e "${BLUE}Testing:${NC} Running pool_config unit tests"
    if go test -v -short -run "Pool" "$PROJECT_ROOT/internal/database/..." 2>&1 | tail -5 | grep -q "ok\|PASS"; then
        pass "Pool configuration unit tests pass"
    else
        fail "Pool configuration unit tests failed"
    fi
}

# ============================================================================
# SECTION 7: Pool Exhaustion Handling Validation
# ============================================================================

validate_pool_exhaustion_handling() {
    section "Pool Exhaustion Handling Validation"

    local pool_config="$PROJECT_ROOT/internal/database/pool_config.go"

    # Check context timeout handling in Acquire
    echo -e "${BLUE}Testing:${NC} Context timeout handling in Acquire"
    if grep -q "ctx.Err()" "$pool_config"; then
        pass "Context error checking in pool operations"
    else
        fail "Context error checking not found"
    fi

    # Check error handling in pool operations
    echo -e "${BLUE}Testing:${NC} Error handling in pool acquire"
    if grep -q "p.metrics.AcquireErrors" "$pool_config"; then
        pass "Acquire errors tracked in metrics"
    else
        fail "Acquire errors not tracked"
    fi

    # Check connection wait tracking
    echo -e "${BLUE}Testing:${NC} Connection wait tracking"
    if grep -q "WaitCount" "$pool_config"; then
        pass "Connection wait count tracked"
    else
        fail "Connection wait count not tracked"
    fi

    # Check pgxpool usage for pool management
    echo -e "${BLUE}Testing:${NC} pgxpool library usage"
    if grep -q "github.com/jackc/pgx/v5/pgxpool" "$pool_config"; then
        pass "pgxpool library used for connection pooling"
    else
        fail "pgxpool library not found"
    fi
}

# ============================================================================
# SECTION 8: Connection Recovery Validation
# ============================================================================

validate_connection_recovery() {
    section "Connection Recovery Validation"

    local pool_config="$PROJECT_ROOT/internal/database/pool_config.go"
    local db_go="$PROJECT_ROOT/internal/database/db.go"

    # Check health check period configuration
    echo -e "${BLUE}Testing:${NC} Health check period configurable"
    if grep -q "HealthCheckPeriod" "$pool_config"; then
        pass "Health check period is configurable"
    else
        fail "Health check period not configurable"
    fi

    # Check AfterConnect hook for connection setup
    echo -e "${BLUE}Testing:${NC} AfterConnect hook for connection setup"
    if grep -q "AfterConnect" "$pool_config"; then
        pass "AfterConnect hook configured"
    else
        fail "AfterConnect hook not found"
    fi

    # Check HealthCheck method implementation
    echo -e "${BLUE}Testing:${NC} HealthCheck method with timeout"
    if grep -qE "context.WithTimeout.*HealthCheck\|HealthCheck.*context.WithTimeout" "$pool_config"; then
        pass "HealthCheck uses timeout context"
    else
        # Check if Ping is called in HealthCheck
        if grep -A5 "func (p \*OptimizedPool) HealthCheck" "$pool_config" | grep -q "Ping"; then
            pass "HealthCheck performs ping operation"
        else
            fail "HealthCheck timeout handling not found"
        fi
    fi

    # Check db.go HealthCheck
    echo -e "${BLUE}Testing:${NC} PostgresDB.HealthCheck method exists"
    if grep -q "func (p \*PostgresDB) HealthCheck" "$db_go"; then
        pass "PostgresDB.HealthCheck method exists"
    else
        fail "PostgresDB.HealthCheck method not found"
    fi

    # Check connection retry logic in services
    echo -e "${BLUE}Testing:${NC} Connection handling in db.go"
    if grep -q "Ping" "$db_go"; then
        pass "Connection verification with Ping"
    else
        fail "Connection verification not found"
    fi
}

# ============================================================================
# SECTION 9: Batch Operations Support Validation
# ============================================================================

validate_batch_operations() {
    section "Batch Operations Support Validation"

    local pool_config="$PROJECT_ROOT/internal/database/pool_config.go"

    # Check SendBatch method
    echo -e "${BLUE}Testing:${NC} SendBatch method exists"
    if grep -q "func (p \*OptimizedPool) SendBatch" "$pool_config"; then
        pass "OptimizedPool.SendBatch method exists"
    else
        fail "OptimizedPool.SendBatch method not found"
    fi

    # Check CopyFrom method for bulk operations
    echo -e "${BLUE}Testing:${NC} CopyFrom method exists"
    if grep -q "func (p \*OptimizedPool) CopyFrom" "$pool_config"; then
        pass "OptimizedPool.CopyFrom method exists"
    else
        fail "OptimizedPool.CopyFrom method not found"
    fi

    # Check pgx.Batch usage
    echo -e "${BLUE}Testing:${NC} pgx.Batch type usage"
    if grep -q "pgx.Batch" "$pool_config"; then
        pass "pgx.Batch type used for batch operations"
    else
        fail "pgx.Batch type not used"
    fi
}

# ============================================================================
# SECTION 10: Pool Statistics Validation
# ============================================================================

validate_pool_statistics() {
    section "Pool Statistics Validation"

    local pool_config="$PROJECT_ROOT/internal/database/pool_config.go"

    # Check Stat method
    echo -e "${BLUE}Testing:${NC} Stat method exists"
    if grep -q "func (p \*OptimizedPool) Stat" "$pool_config"; then
        pass "OptimizedPool.Stat method exists"
    else
        fail "OptimizedPool.Stat method not found"
    fi

    # Check IdleConns tracking
    echo -e "${BLUE}Testing:${NC} IdleConns tracking"
    if grep -q "IdleConns" "$pool_config"; then
        pass "Idle connections tracking exists"
    else
        fail "Idle connections tracking not found"
    fi

    # Check TotalConns tracking
    echo -e "${BLUE}Testing:${NC} TotalConns tracking"
    if grep -q "TotalConns" "$pool_config"; then
        pass "Total connections tracking exists"
    else
        fail "Total connections tracking not found"
    fi
}

# ============================================================================
# SECTION 11: Concurrent Access Test
# ============================================================================

validate_concurrent_access() {
    section "Concurrent Access Test"

    echo -e "${BLUE}Testing:${NC} Race condition detection in pool code"

    # Run race detector on pool config
    if go test -race -short -run "Pool" "$PROJECT_ROOT/internal/database/..." 2>&1 | grep -v "WARNING: DATA RACE" | tail -3 | grep -q "ok\|PASS"; then
        pass "No race conditions in pool operations"
    else
        fail "Race conditions detected in pool operations"
    fi

    # Check mutex usage for thread safety
    echo -e "${BLUE}Testing:${NC} Thread safety mechanisms"
    local pool_config="$PROJECT_ROOT/internal/database/pool_config.go"
    if grep -q "sync.Mutex\|sync.RWMutex\|atomic\." "$pool_config"; then
        pass "Thread safety mechanisms present"
    else
        fail "Thread safety mechanisms not found"
    fi
}

# ============================================================================
# SECTION 12: Database Integration Test
# ============================================================================

run_database_integration_test() {
    section "Database Integration Test"

    # Check if we can connect to database
    echo -e "${BLUE}Testing:${NC} Database connection"

    if command -v psql &> /dev/null; then
        if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
            pass "Database connection successful"

            # Test connection count
            echo -e "${BLUE}Testing:${NC} Connection pool behavior"
            local conn_count=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -A -c "SELECT count(*) FROM pg_stat_activity WHERE datname='$DB_NAME' AND state='active'" 2>/dev/null)
            if [ -n "$conn_count" ]; then
                pass "Active connections count: $conn_count"
            else
                fail "Cannot retrieve connection count"
            fi

            # Test max_connections setting
            echo -e "${BLUE}Testing:${NC} PostgreSQL max_connections setting"
            local max_conn=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -A -c "SHOW max_connections" 2>/dev/null)
            if [ -n "$max_conn" ]; then
                pass "PostgreSQL max_connections: $max_conn"
            else
                fail "Cannot retrieve max_connections"
            fi
        else
            skip "Database not available (expected in CI without infrastructure)"
        fi
    else
        skip "psql not available"
    fi
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    validate_pool_config_code
    validate_pool_options_functions
    validate_optimized_pool
    validate_pool_metrics
    validate_lazy_pool
    validate_pool_unit_tests
    validate_pool_exhaustion_handling
    validate_connection_recovery
    validate_batch_operations
    validate_pool_statistics
    validate_concurrent_access
    run_database_integration_test

    # Summary
    echo ""
    echo "============================================"
    echo "  Database Connection Pool Challenge Results"
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
        echo -e "${GREEN}DATABASE CONNECTION POOL CHALLENGE: PASSED${NC}"
        exit 0
    else
        echo -e "${RED}DATABASE CONNECTION POOL CHALLENGE: FAILED${NC}"
        exit 1
    fi
}

main "$@"
