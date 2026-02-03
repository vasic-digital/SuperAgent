#!/bin/bash
# Redis Cache Validation Challenge
# Validates Redis cache operations, TTL behavior, and eviction policies
# Tests: Cache set/get operations, TTL behavior, eviction policies, metrics

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

# Redis connection settings
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-16379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-helixagent123}"

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
echo "  REDIS CACHE VALIDATION CHALLENGE"
echo "  Tests cache operations and TTL behavior"
echo "============================================"
echo ""
echo "Redis: $REDIS_HOST:$REDIS_PORT"
echo ""

# ============================================================================
# SECTION 1: Redis Client Code Validation
# ============================================================================

validate_redis_client_code() {
    section "Redis Client Code Validation"

    local redis_go="$PROJECT_ROOT/internal/cache/redis.go"

    # Check redis.go exists
    echo -e "${BLUE}Testing:${NC} redis.go exists"
    if [ -f "$redis_go" ]; then
        pass "redis.go exists"
    else
        fail "redis.go not found"
        return 1
    fi

    # Check RedisClient struct
    echo -e "${BLUE}Testing:${NC} RedisClient struct defined"
    if grep -q "type RedisClient struct" "$redis_go"; then
        pass "RedisClient struct defined"
    else
        fail "RedisClient struct not found"
    fi

    # Check NewRedisClient constructor
    echo -e "${BLUE}Testing:${NC} NewRedisClient constructor exists"
    if grep -q "func NewRedisClient" "$redis_go"; then
        pass "NewRedisClient constructor exists"
    else
        fail "NewRedisClient constructor not found"
    fi

    # Check Set method
    echo -e "${BLUE}Testing:${NC} Set method exists"
    if grep -q "func (r \*RedisClient) Set" "$redis_go"; then
        pass "RedisClient.Set method exists"
    else
        fail "RedisClient.Set method not found"
    fi

    # Check Get method
    echo -e "${BLUE}Testing:${NC} Get method exists"
    if grep -q "func (r \*RedisClient) Get" "$redis_go"; then
        pass "RedisClient.Get method exists"
    else
        fail "RedisClient.Get method not found"
    fi

    # Check Delete method
    echo -e "${BLUE}Testing:${NC} Delete method exists"
    if grep -q "func (r \*RedisClient) Delete" "$redis_go"; then
        pass "RedisClient.Delete method exists"
    else
        fail "RedisClient.Delete method not found"
    fi

    # Check MGet method (multi-get)
    echo -e "${BLUE}Testing:${NC} MGet method exists"
    if grep -q "func (r \*RedisClient) MGet" "$redis_go"; then
        pass "RedisClient.MGet method exists"
    else
        fail "RedisClient.MGet method not found"
    fi

    # Check Pipeline method
    echo -e "${BLUE}Testing:${NC} Pipeline method exists"
    if grep -q "func (r \*RedisClient) Pipeline" "$redis_go"; then
        pass "RedisClient.Pipeline method exists"
    else
        fail "RedisClient.Pipeline method not found"
    fi

    # Check Ping method for health check
    echo -e "${BLUE}Testing:${NC} Ping method exists"
    if grep -q "func (r \*RedisClient) Ping" "$redis_go"; then
        pass "RedisClient.Ping method exists"
    else
        fail "RedisClient.Ping method not found"
    fi

    # Check Close method
    echo -e "${BLUE}Testing:${NC} Close method exists"
    if grep -q "func (r \*RedisClient) Close" "$redis_go"; then
        pass "RedisClient.Close method exists"
    else
        fail "RedisClient.Close method not found"
    fi

    # Check go-redis library usage
    echo -e "${BLUE}Testing:${NC} go-redis library usage"
    if grep -q "github.com/redis/go-redis/v9" "$redis_go"; then
        pass "go-redis v9 library used"
    else
        fail "go-redis library not found"
    fi

    # Check JSON serialization
    echo -e "${BLUE}Testing:${NC} JSON serialization for cache values"
    if grep -q "json.Marshal\|json.Unmarshal" "$redis_go"; then
        pass "JSON serialization used for cache values"
    else
        fail "JSON serialization not found"
    fi

    # Check TTL/expiration parameter
    echo -e "${BLUE}Testing:${NC} TTL parameter in Set method"
    if grep -q "expiration time.Duration" "$redis_go"; then
        pass "TTL parameter in Set method"
    else
        fail "TTL parameter not found in Set method"
    fi
}

# ============================================================================
# SECTION 2: Tiered Cache Implementation Validation
# ============================================================================

validate_tiered_cache() {
    section "Tiered Cache Implementation Validation"

    local tiered_cache="$PROJECT_ROOT/internal/cache/tiered_cache.go"

    # Check tiered_cache.go exists
    echo -e "${BLUE}Testing:${NC} tiered_cache.go exists"
    if [ -f "$tiered_cache" ]; then
        pass "tiered_cache.go exists"
    else
        fail "tiered_cache.go not found"
        return 1
    fi

    # Check TieredCache struct
    echo -e "${BLUE}Testing:${NC} TieredCache struct defined"
    if grep -q "type TieredCache struct" "$tiered_cache"; then
        pass "TieredCache struct defined"
    else
        fail "TieredCache struct not found"
    fi

    # Check TieredCacheConfig struct
    echo -e "${BLUE}Testing:${NC} TieredCacheConfig struct defined"
    if grep -q "type TieredCacheConfig struct" "$tiered_cache"; then
        pass "TieredCacheConfig struct defined"
    else
        fail "TieredCacheConfig struct not found"
    fi

    # Check L1 (memory) cache configuration
    echo -e "${BLUE}Testing:${NC} L1 (memory) cache configuration"
    if grep -q "L1MaxSize\|L1TTL" "$tiered_cache"; then
        pass "L1 cache configuration exists"
    else
        fail "L1 cache configuration not found"
    fi

    # Check L2 (Redis) cache configuration
    echo -e "${BLUE}Testing:${NC} L2 (Redis) cache configuration"
    if grep -q "L2TTL\|L2Compression\|L2KeyPrefix" "$tiered_cache"; then
        pass "L2 cache configuration exists"
    else
        fail "L2 cache configuration not found"
    fi

    # Check NewTieredCache constructor
    echo -e "${BLUE}Testing:${NC} NewTieredCache constructor exists"
    if grep -q "func NewTieredCache" "$tiered_cache"; then
        pass "NewTieredCache constructor exists"
    else
        fail "NewTieredCache constructor not found"
    fi

    # Check Get method with L1/L2 fallback
    echo -e "${BLUE}Testing:${NC} Get method with tiered lookup"
    if grep -q "func (c \*TieredCache) Get" "$tiered_cache"; then
        pass "TieredCache.Get method exists"
    else
        fail "TieredCache.Get method not found"
    fi

    # Check Set method for both tiers
    echo -e "${BLUE}Testing:${NC} Set method for tiered storage"
    if grep -q "func (c \*TieredCache) Set" "$tiered_cache"; then
        pass "TieredCache.Set method exists"
    else
        fail "TieredCache.Set method not found"
    fi

    # Check Delete method
    echo -e "${BLUE}Testing:${NC} Delete method for tiered cache"
    if grep -q "func (c \*TieredCache) Delete" "$tiered_cache"; then
        pass "TieredCache.Delete method exists"
    else
        fail "TieredCache.Delete method not found"
    fi

    # Check compression support
    echo -e "${BLUE}Testing:${NC} Compression support"
    if grep -q "compress\|gzip\|L2Compression" "$tiered_cache"; then
        pass "Compression support exists"
    else
        fail "Compression support not found"
    fi
}

# ============================================================================
# SECTION 3: Cache Metrics Validation
# ============================================================================

validate_cache_metrics() {
    section "Cache Metrics Validation"

    local tiered_cache="$PROJECT_ROOT/internal/cache/tiered_cache.go"
    local metrics_go="$PROJECT_ROOT/internal/cache/metrics.go"

    # Check TieredCacheMetrics struct
    echo -e "${BLUE}Testing:${NC} TieredCacheMetrics struct defined"
    if grep -q "type TieredCacheMetrics struct" "$tiered_cache"; then
        pass "TieredCacheMetrics struct defined"
    else
        fail "TieredCacheMetrics struct not found"
    fi

    # Check L1Hits metric
    echo -e "${BLUE}Testing:${NC} L1Hits metric exists"
    if grep -q "L1Hits" "$tiered_cache"; then
        pass "L1Hits metric exists"
    else
        fail "L1Hits metric not found"
    fi

    # Check L1Misses metric
    echo -e "${BLUE}Testing:${NC} L1Misses metric exists"
    if grep -q "L1Misses" "$tiered_cache"; then
        pass "L1Misses metric exists"
    else
        fail "L1Misses metric not found"
    fi

    # Check L2Hits metric
    echo -e "${BLUE}Testing:${NC} L2Hits metric exists"
    if grep -q "L2Hits" "$tiered_cache"; then
        pass "L2Hits metric exists"
    else
        fail "L2Hits metric not found"
    fi

    # Check L2Misses metric
    echo -e "${BLUE}Testing:${NC} L2Misses metric exists"
    if grep -q "L2Misses" "$tiered_cache"; then
        pass "L2Misses metric exists"
    else
        fail "L2Misses metric not found"
    fi

    # Check Invalidations metric
    echo -e "${BLUE}Testing:${NC} Invalidations metric exists"
    if grep -q "Invalidations" "$tiered_cache"; then
        pass "Invalidations metric exists"
    else
        fail "Invalidations metric not found"
    fi

    # Check Expirations metric
    echo -e "${BLUE}Testing:${NC} Expirations metric exists"
    if grep -q "Expirations" "$tiered_cache"; then
        pass "Expirations metric exists"
    else
        fail "Expirations metric not found"
    fi

    # Check CompressionSaved metric
    echo -e "${BLUE}Testing:${NC} CompressionSaved metric exists"
    if grep -q "CompressionSaved" "$tiered_cache"; then
        pass "CompressionSaved metric exists"
    else
        fail "CompressionSaved metric not found"
    fi

    # Check Metrics method
    echo -e "${BLUE}Testing:${NC} Metrics method exists"
    if grep -q "func (c \*TieredCache) Metrics" "$tiered_cache"; then
        pass "TieredCache.Metrics method exists"
    else
        fail "TieredCache.Metrics method not found"
    fi

    # Check HitRate method
    echo -e "${BLUE}Testing:${NC} HitRate method exists"
    if grep -q "func (c \*TieredCache) HitRate" "$tiered_cache"; then
        pass "TieredCache.HitRate method exists"
    else
        fail "TieredCache.HitRate method not found"
    fi

    # Check metrics.go for cache metrics service
    if [ -f "$metrics_go" ]; then
        echo -e "${BLUE}Testing:${NC} Dedicated metrics module exists"
        pass "cache/metrics.go exists"
    else
        skip "cache/metrics.go not found (metrics may be inline)"
    fi
}

# ============================================================================
# SECTION 4: TTL Behavior Validation
# ============================================================================

validate_ttl_behavior() {
    section "TTL Behavior Validation"

    local tiered_cache="$PROJECT_ROOT/internal/cache/tiered_cache.go"
    local expiration_go="$PROJECT_ROOT/internal/cache/expiration.go"

    # Check TTL handling in L1 cache
    echo -e "${BLUE}Testing:${NC} L1 TTL handling"
    if grep -q "expiresAt\|L1TTL" "$tiered_cache"; then
        pass "L1 TTL handling exists"
    else
        fail "L1 TTL handling not found"
    fi

    # Check TTL handling in L2 cache
    echo -e "${BLUE}Testing:${NC} L2 TTL handling"
    if grep -q "L2TTL\|l2Set.*ttl" "$tiered_cache"; then
        pass "L2 TTL handling exists"
    else
        fail "L2 TTL handling not found"
    fi

    # Check expiration checking
    echo -e "${BLUE}Testing:${NC} Expiration checking logic"
    if grep -q "time.Now().After(.*expiresAt)\|entry.expiresAt" "$tiered_cache"; then
        pass "Expiration checking logic exists"
    else
        fail "Expiration checking logic not found"
    fi

    # Check negative TTL for not-found caching
    echo -e "${BLUE}Testing:${NC} Negative TTL configuration"
    if grep -q "NegativeTTL" "$tiered_cache"; then
        pass "Negative TTL configuration exists"
    else
        fail "Negative TTL configuration not found"
    fi

    # Check expiration.go for advanced expiration handling
    if [ -f "$expiration_go" ]; then
        echo -e "${BLUE}Testing:${NC} Dedicated expiration module exists"
        pass "cache/expiration.go exists"

        # Check expiration policies
        if grep -qE "ExpirationPolicy|TTLPolicy|ExpirationConfig|ExpirationManager" "$expiration_go"; then
            pass "Expiration policies/config defined"
        else
            fail "Expiration policies not found"
        fi
    else
        skip "cache/expiration.go not found (expiration may be inline)"
    fi

    # Check L1 cleanup goroutine
    echo -e "${BLUE}Testing:${NC} L1 cleanup goroutine"
    if grep -q "l1CleanupLoop\|L1CleanupInterval" "$tiered_cache"; then
        pass "L1 cleanup goroutine exists"
    else
        fail "L1 cleanup goroutine not found"
    fi
}

# ============================================================================
# SECTION 5: Eviction Policy Validation
# ============================================================================

validate_eviction_policies() {
    section "Eviction Policy Validation"

    local tiered_cache="$PROJECT_ROOT/internal/cache/tiered_cache.go"

    # Check L1 max size limit
    echo -e "${BLUE}Testing:${NC} L1 max size limit"
    if grep -q "L1MaxSize\|maxSize" "$tiered_cache"; then
        pass "L1 max size limit exists"
    else
        fail "L1 max size limit not found"
    fi

    # Check eviction logic
    echo -e "${BLUE}Testing:${NC} Eviction logic"
    if grep -q "l1EvictLRU\|L1Evictions\|evict" "$tiered_cache"; then
        pass "Eviction logic exists"
    else
        fail "Eviction logic not found"
    fi

    # Check hit count tracking for LRU approximation
    echo -e "${BLUE}Testing:${NC} Hit count tracking"
    if grep -q "hitCount" "$tiered_cache"; then
        pass "Hit count tracking for LRU approximation"
    else
        fail "Hit count tracking not found"
    fi

    # Check L1 evictions metric
    echo -e "${BLUE}Testing:${NC} L1Evictions metric"
    if grep -q "L1Evictions" "$tiered_cache"; then
        pass "L1Evictions metric exists"
    else
        fail "L1Evictions metric not found"
    fi

    # Check eviction on capacity reached
    echo -e "${BLUE}Testing:${NC} Eviction on capacity reached"
    if grep -q "len.*entries.*>=.*maxSize" "$tiered_cache"; then
        pass "Eviction triggered at capacity"
    else
        fail "Eviction trigger not found"
    fi
}

# ============================================================================
# SECTION 6: Tag-Based Cache Support Validation
# ============================================================================

validate_tag_based_cache() {
    section "Tag-Based Cache Support Validation"

    local tiered_cache="$PROJECT_ROOT/internal/cache/tiered_cache.go"

    # Check tag index structure
    echo -e "${BLUE}Testing:${NC} Tag index structure"
    if grep -q "tagIndex\|type tagIndex struct" "$tiered_cache"; then
        pass "Tag index structure exists"
    else
        fail "Tag index structure not found"
    fi

    # Check InvalidateByTag method
    echo -e "${BLUE}Testing:${NC} InvalidateByTag method"
    if grep -q "func (c \*TieredCache) InvalidateByTag" "$tiered_cache"; then
        pass "InvalidateByTag method exists"
    else
        fail "InvalidateByTag method not found"
    fi

    # Check InvalidateByTags method
    echo -e "${BLUE}Testing:${NC} InvalidateByTags method"
    if grep -q "func (c \*TieredCache) InvalidateByTags" "$tiered_cache"; then
        pass "InvalidateByTags method exists"
    else
        fail "InvalidateByTags method not found"
    fi

    # Check InvalidatePrefix method
    echo -e "${BLUE}Testing:${NC} InvalidatePrefix method"
    if grep -q "func (c \*TieredCache) InvalidatePrefix" "$tiered_cache"; then
        pass "InvalidatePrefix method exists"
    else
        fail "InvalidatePrefix method not found"
    fi

    # Check tag parameter in Set
    echo -e "${BLUE}Testing:${NC} Tag parameter in Set method"
    if grep -qE "tags\s+\.\.\.string" "$tiered_cache"; then
        pass "Set method accepts tags parameter"
    else
        fail "Set method missing tags parameter"
    fi
}

# ============================================================================
# SECTION 7: Cache Service Validation
# ============================================================================

validate_cache_service() {
    section "Cache Service Validation"

    local cache_service="$PROJECT_ROOT/internal/cache/cache_service.go"

    # Check cache_service.go exists
    echo -e "${BLUE}Testing:${NC} cache_service.go exists"
    if [ -f "$cache_service" ]; then
        pass "cache_service.go exists"
    else
        fail "cache_service.go not found"
        return 1
    fi

    # Check CacheService struct
    echo -e "${BLUE}Testing:${NC} CacheService struct defined"
    if grep -q "type CacheService struct\|type Service struct" "$cache_service"; then
        pass "CacheService struct defined"
    else
        fail "CacheService struct not found"
    fi

    # Check cache interface
    echo -e "${BLUE}Testing:${NC} Cache interface defined"
    if grep -q "type.*Cache.*interface\|type CacheInterface interface" "$cache_service"; then
        pass "Cache interface defined"
    else
        skip "Cache interface may be defined elsewhere"
    fi
}

# ============================================================================
# SECTION 8: Provider Cache Validation
# ============================================================================

validate_provider_cache() {
    section "Provider Cache Validation"

    local provider_cache="$PROJECT_ROOT/internal/cache/provider_cache.go"

    # Check provider_cache.go exists
    echo -e "${BLUE}Testing:${NC} provider_cache.go exists"
    if [ -f "$provider_cache" ]; then
        pass "provider_cache.go exists"

        # Check provider-specific caching
        if grep -q "ProviderCache\|providerCache" "$provider_cache"; then
            pass "Provider-specific caching exists"
        else
            fail "Provider-specific caching not found"
        fi
    else
        skip "provider_cache.go not found"
    fi

    # Check MCP cache
    local mcp_cache="$PROJECT_ROOT/internal/cache/mcp_cache.go"
    echo -e "${BLUE}Testing:${NC} mcp_cache.go exists"
    if [ -f "$mcp_cache" ]; then
        pass "mcp_cache.go exists"

        if grep -q "MCPCache" "$mcp_cache"; then
            pass "MCP-specific caching exists"
        else
            fail "MCP-specific caching not found"
        fi
    else
        skip "mcp_cache.go not found"
    fi
}

# ============================================================================
# SECTION 9: Cache Unit Tests Validation
# ============================================================================

validate_cache_unit_tests() {
    section "Cache Unit Tests Validation"

    local test_files=(
        "internal/cache/redis_test.go"
        "internal/cache/tiered_cache_test.go"
        "internal/cache/expiration_test.go"
        "internal/cache/invalidation_test.go"
        "internal/cache/metrics_test.go"
    )

    for test_file in "${test_files[@]}"; do
        local full_path="$PROJECT_ROOT/$test_file"
        if [ -f "$full_path" ]; then
            pass "Test file exists: $(basename $test_file)"
        else
            fail "Test file missing: $test_file"
        fi
    done

    # Run cache unit tests
    echo -e "${BLUE}Testing:${NC} Running cache unit tests"
    if go test -v -short "$PROJECT_ROOT/internal/cache/..." 2>&1 | tail -5 | grep -q "ok\|PASS"; then
        pass "Cache unit tests pass"
    else
        fail "Cache unit tests failed"
    fi
}

# ============================================================================
# SECTION 10: Thread Safety Validation
# ============================================================================

validate_thread_safety() {
    section "Thread Safety Validation"

    local tiered_cache="$PROJECT_ROOT/internal/cache/tiered_cache.go"

    # Check mutex usage
    echo -e "${BLUE}Testing:${NC} Mutex usage in L1 cache"
    if grep -q "sync.RWMutex\|sync.Mutex" "$tiered_cache"; then
        pass "Mutex protection in L1 cache"
    else
        fail "Mutex protection not found"
    fi

    # Check atomic operations
    echo -e "${BLUE}Testing:${NC} Atomic operations for metrics"
    if grep -q "atomic.AddInt64\|atomic.LoadInt64" "$tiered_cache"; then
        pass "Atomic operations for metrics"
    else
        fail "Atomic operations not found"
    fi

    # Run race detection
    echo -e "${BLUE}Testing:${NC} Race condition detection"
    if go test -race -short "$PROJECT_ROOT/internal/cache/..." 2>&1 | grep -v "WARNING: DATA RACE" | tail -3 | grep -q "ok\|PASS"; then
        pass "No race conditions in cache operations"
    else
        fail "Race conditions detected in cache operations"
    fi
}

# ============================================================================
# SECTION 11: Redis Integration Test
# ============================================================================

run_redis_integration_test() {
    section "Redis Integration Test"

    # Check if redis-cli is available
    echo -e "${BLUE}Testing:${NC} Redis connection"

    if command -v redis-cli &> /dev/null; then
        local ping_result
        if [ -n "$REDIS_PASSWORD" ]; then
            ping_result=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" PING 2>/dev/null)
        else
            ping_result=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" PING 2>/dev/null)
        fi

        if [ "$ping_result" = "PONG" ]; then
            pass "Redis connection successful"

            # Test SET/GET operations
            echo -e "${BLUE}Testing:${NC} Redis SET/GET operations"
            local test_key="helixagent:challenge:test:$(date +%s)"
            local test_value="test_value_123"

            if [ -n "$REDIS_PASSWORD" ]; then
                redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" SET "$test_key" "$test_value" EX 60 > /dev/null 2>&1
                local get_result=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" GET "$test_key" 2>/dev/null)
            else
                redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" SET "$test_key" "$test_value" EX 60 > /dev/null 2>&1
                local get_result=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" GET "$test_key" 2>/dev/null)
            fi

            if [ "$get_result" = "$test_value" ]; then
                pass "Redis SET/GET works correctly"
            else
                fail "Redis SET/GET failed"
            fi

            # Test TTL
            echo -e "${BLUE}Testing:${NC} Redis TTL behavior"
            if [ -n "$REDIS_PASSWORD" ]; then
                local ttl_result=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" TTL "$test_key" 2>/dev/null)
            else
                local ttl_result=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" TTL "$test_key" 2>/dev/null)
            fi

            if [ "$ttl_result" -gt 0 ] 2>/dev/null; then
                pass "Redis TTL works correctly (TTL: ${ttl_result}s)"
            else
                fail "Redis TTL not working"
            fi

            # Cleanup
            if [ -n "$REDIS_PASSWORD" ]; then
                redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" DEL "$test_key" > /dev/null 2>&1
            else
                redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" DEL "$test_key" > /dev/null 2>&1
            fi

            # Test INFO command
            echo -e "${BLUE}Testing:${NC} Redis INFO command"
            if [ -n "$REDIS_PASSWORD" ]; then
                local memory_info=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" INFO memory 2>/dev/null | grep "used_memory_human")
            else
                local memory_info=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" INFO memory 2>/dev/null | grep "used_memory_human")
            fi

            if [ -n "$memory_info" ]; then
                pass "Redis INFO works ($memory_info)"
            else
                fail "Redis INFO failed"
            fi
        else
            skip "Redis not available (expected in CI without infrastructure)"
        fi
    else
        skip "redis-cli not available"
    fi
}

# ============================================================================
# SECTION 12: Cache Configuration Validation
# ============================================================================

validate_cache_configuration() {
    section "Cache Configuration Validation"

    local tiered_cache="$PROJECT_ROOT/internal/cache/tiered_cache.go"

    # Check DefaultTieredCacheConfig
    echo -e "${BLUE}Testing:${NC} DefaultTieredCacheConfig exists"
    if grep -q "func DefaultTieredCacheConfig" "$tiered_cache"; then
        pass "DefaultTieredCacheConfig function exists"
    else
        fail "DefaultTieredCacheConfig not found"
    fi

    # Check EnableL1 configuration
    echo -e "${BLUE}Testing:${NC} EnableL1 configuration"
    if grep -q "EnableL1" "$tiered_cache"; then
        pass "L1 enable/disable configuration exists"
    else
        fail "L1 enable/disable configuration not found"
    fi

    # Check EnableL2 configuration
    echo -e "${BLUE}Testing:${NC} EnableL2 configuration"
    if grep -q "EnableL2" "$tiered_cache"; then
        pass "L2 enable/disable configuration exists"
    else
        fail "L2 enable/disable configuration not found"
    fi

    # Check L2KeyPrefix configuration
    echo -e "${BLUE}Testing:${NC} L2KeyPrefix configuration"
    if grep -q "L2KeyPrefix" "$tiered_cache"; then
        pass "L2KeyPrefix configuration exists"
    else
        fail "L2KeyPrefix configuration not found"
    fi
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    validate_redis_client_code
    validate_tiered_cache
    validate_cache_metrics
    validate_ttl_behavior
    validate_eviction_policies
    validate_tag_based_cache
    validate_cache_service
    validate_provider_cache
    validate_cache_unit_tests
    validate_thread_safety
    run_redis_integration_test
    validate_cache_configuration

    # Summary
    echo ""
    echo "============================================"
    echo "  Redis Cache Validation Challenge Results"
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
        echo -e "${GREEN}REDIS CACHE VALIDATION CHALLENGE: PASSED${NC}"
        exit 0
    else
        echo -e "${RED}REDIS CACHE VALIDATION CHALLENGE: FAILED${NC}"
        exit 1
    fi
}

main "$@"
