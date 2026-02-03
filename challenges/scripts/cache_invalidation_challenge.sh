#!/bin/bash
# Cache Invalidation Challenge
# Tests cache invalidation mechanisms, cascade invalidation, and atomic operations
# Tests: Manual invalidation, cascade invalidation, atomic operations, event-driven invalidation

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
echo "  CACHE INVALIDATION CHALLENGE"
echo "  Tests invalidation mechanisms and atomicity"
echo "============================================"
echo ""

# ============================================================================
# SECTION 1: Invalidation Strategy Interface Validation
# ============================================================================

validate_invalidation_interface() {
    section "Invalidation Strategy Interface Validation"

    local invalidation_go="$PROJECT_ROOT/internal/cache/invalidation.go"

    # Check invalidation.go exists
    echo -e "${BLUE}Testing:${NC} invalidation.go exists"
    if [ -f "$invalidation_go" ]; then
        pass "invalidation.go exists"
    else
        fail "invalidation.go not found"
        return 1
    fi

    # Check InvalidationStrategy interface
    echo -e "${BLUE}Testing:${NC} InvalidationStrategy interface defined"
    if grep -q "type InvalidationStrategy interface" "$invalidation_go"; then
        pass "InvalidationStrategy interface defined"
    else
        fail "InvalidationStrategy interface not found"
    fi

    # Check ShouldInvalidate method in interface
    echo -e "${BLUE}Testing:${NC} ShouldInvalidate method in interface"
    if grep -q "ShouldInvalidate" "$invalidation_go"; then
        pass "ShouldInvalidate method defined"
    else
        fail "ShouldInvalidate method not found"
    fi

    # Check Name method in interface
    echo -e "${BLUE}Testing:${NC} Name method in interface"
    if grep -q "Name() string" "$invalidation_go"; then
        pass "Name method defined in interface"
    else
        fail "Name method not found in interface"
    fi
}

# ============================================================================
# SECTION 2: Manual Invalidation Validation
# ============================================================================

validate_manual_invalidation() {
    section "Manual Invalidation Validation"

    local tiered_cache="$PROJECT_ROOT/internal/cache/tiered_cache.go"
    local invalidation_go="$PROJECT_ROOT/internal/cache/invalidation.go"

    # Check Delete method for single key invalidation
    echo -e "${BLUE}Testing:${NC} Delete method for manual invalidation"
    if grep -q "func (c \*TieredCache) Delete" "$tiered_cache"; then
        pass "TieredCache.Delete method exists"
    else
        fail "TieredCache.Delete method not found"
    fi

    # Check L1 deletion
    echo -e "${BLUE}Testing:${NC} L1 cache deletion"
    if grep -q "l1Delete\|delete.*l1" "$tiered_cache"; then
        pass "L1 cache deletion implemented"
    else
        fail "L1 cache deletion not found"
    fi

    # Check L2 deletion
    echo -e "${BLUE}Testing:${NC} L2 cache deletion"
    if grep -q "l2.Del\|l2.*Del\|redis.*Del" "$tiered_cache"; then
        pass "L2 cache deletion implemented"
    else
        fail "L2 cache deletion not found"
    fi

    # Check invalidation metrics increment
    echo -e "${BLUE}Testing:${NC} Invalidation metrics tracking"
    if grep -q "metrics.Invalidations\|Invalidations" "$tiered_cache"; then
        pass "Invalidation metrics tracked"
    else
        fail "Invalidation metrics not tracked"
    fi
}

# ============================================================================
# SECTION 3: Tag-Based Invalidation Validation
# ============================================================================

validate_tag_invalidation() {
    section "Tag-Based Invalidation Validation"

    local invalidation_go="$PROJECT_ROOT/internal/cache/invalidation.go"

    # Check TagBasedInvalidation struct
    echo -e "${BLUE}Testing:${NC} TagBasedInvalidation struct defined"
    if grep -q "type TagBasedInvalidation struct" "$invalidation_go"; then
        pass "TagBasedInvalidation struct defined"
    else
        fail "TagBasedInvalidation struct not found"
    fi

    # Check NewTagBasedInvalidation constructor
    echo -e "${BLUE}Testing:${NC} NewTagBasedInvalidation constructor"
    if grep -q "func NewTagBasedInvalidation" "$invalidation_go"; then
        pass "NewTagBasedInvalidation constructor exists"
    else
        fail "NewTagBasedInvalidation constructor not found"
    fi

    # Check AddTag method
    echo -e "${BLUE}Testing:${NC} AddTag method"
    if grep -q "func (i \*TagBasedInvalidation) AddTag" "$invalidation_go"; then
        pass "AddTag method exists"
    else
        fail "AddTag method not found"
    fi

    # Check RemoveKey method
    echo -e "${BLUE}Testing:${NC} RemoveKey method"
    if grep -q "func (i \*TagBasedInvalidation) RemoveKey" "$invalidation_go"; then
        pass "RemoveKey method exists"
    else
        fail "RemoveKey method not found"
    fi

    # Check InvalidateByTag method
    echo -e "${BLUE}Testing:${NC} InvalidateByTag method"
    if grep -q "func (i \*TagBasedInvalidation) InvalidateByTag" "$invalidation_go"; then
        pass "InvalidateByTag method exists"
    else
        fail "InvalidateByTag method not found"
    fi

    # Check InvalidateByTags method
    echo -e "${BLUE}Testing:${NC} InvalidateByTags method"
    if grep -q "func (i \*TagBasedInvalidation) InvalidateByTags" "$invalidation_go"; then
        pass "InvalidateByTags method exists"
    else
        fail "InvalidateByTags method not found"
    fi

    # Check tag index structure
    echo -e "${BLUE}Testing:${NC} Tag index data structure"
    if grep -q "tagIndex map\[string\]map\[string\]" "$invalidation_go"; then
        pass "Tag index data structure (tag -> keys mapping)"
    else
        fail "Tag index data structure not found"
    fi

    # Check keyTags mapping
    echo -e "${BLUE}Testing:${NC} Key to tags mapping"
    if grep -q "keyTags.*map\[string\]\[\]string" "$invalidation_go"; then
        pass "Key to tags reverse mapping exists"
    else
        fail "Key to tags mapping not found"
    fi
}

# ============================================================================
# SECTION 4: Cascade Invalidation Validation
# ============================================================================

validate_cascade_invalidation() {
    section "Cascade Invalidation Validation"

    local tiered_cache="$PROJECT_ROOT/internal/cache/tiered_cache.go"
    local invalidation_go="$PROJECT_ROOT/internal/cache/invalidation.go"

    # Check InvalidatePrefix method (cascade by key prefix)
    echo -e "${BLUE}Testing:${NC} InvalidatePrefix method for cascade"
    if grep -q "func (c \*TieredCache) InvalidatePrefix" "$tiered_cache"; then
        pass "InvalidatePrefix method exists"
    else
        fail "InvalidatePrefix method not found"
    fi

    # Check SCAN usage for prefix invalidation in Redis
    echo -e "${BLUE}Testing:${NC} Redis SCAN for prefix invalidation"
    if grep -q "l2.Scan\|Scan.*pattern" "$tiered_cache"; then
        pass "Redis SCAN used for prefix invalidation"
    else
        fail "Redis SCAN not found for prefix invalidation"
    fi

    # Check wildcard handling
    echo -e "${BLUE}Testing:${NC} Wildcard pattern handling"
    if grep -q "containsWildcard\|\\*" "$invalidation_go" "$tiered_cache" 2>/dev/null; then
        pass "Wildcard pattern handling exists"
    else
        fail "Wildcard pattern handling not found"
    fi

    # Check cascade through tag relationships
    echo -e "${BLUE}Testing:${NC} Cascade through tags"
    local tiered_cache="$PROJECT_ROOT/internal/cache/tiered_cache.go"
    if grep -q "InvalidateByTags" "$tiered_cache"; then
        pass "Cascade invalidation through tags"
    else
        fail "Cascade tag invalidation not found"
    fi
}

# ============================================================================
# SECTION 5: Event-Driven Invalidation Validation
# ============================================================================

validate_event_driven_invalidation() {
    section "Event-Driven Invalidation Validation"

    local invalidation_go="$PROJECT_ROOT/internal/cache/invalidation.go"

    # Check EventDrivenInvalidation struct
    echo -e "${BLUE}Testing:${NC} EventDrivenInvalidation struct defined"
    if grep -q "type EventDrivenInvalidation struct" "$invalidation_go"; then
        pass "EventDrivenInvalidation struct defined"
    else
        fail "EventDrivenInvalidation struct not found"
    fi

    # Check NewEventDrivenInvalidation constructor
    echo -e "${BLUE}Testing:${NC} NewEventDrivenInvalidation constructor"
    if grep -q "func NewEventDrivenInvalidation" "$invalidation_go"; then
        pass "NewEventDrivenInvalidation constructor exists"
    else
        fail "NewEventDrivenInvalidation constructor not found"
    fi

    # Check InvalidationRule struct
    echo -e "${BLUE}Testing:${NC} InvalidationRule struct defined"
    if grep -q "type InvalidationRule struct" "$invalidation_go"; then
        pass "InvalidationRule struct defined"
    else
        fail "InvalidationRule struct not found"
    fi

    # Check EventType field in rule
    echo -e "${BLUE}Testing:${NC} EventType field in InvalidationRule"
    if grep -q "EventType" "$invalidation_go"; then
        pass "EventType field in InvalidationRule"
    else
        fail "EventType field not found"
    fi

    # Check KeyPattern field in rule
    echo -e "${BLUE}Testing:${NC} KeyPattern field in InvalidationRule"
    if grep -q "KeyPattern" "$invalidation_go"; then
        pass "KeyPattern field in InvalidationRule"
    else
        fail "KeyPattern field not found"
    fi

    # Check Handler field for custom logic
    echo -e "${BLUE}Testing:${NC} Custom Handler in InvalidationRule"
    if grep -q "Handler.*func" "$invalidation_go"; then
        pass "Custom Handler function in InvalidationRule"
    else
        fail "Custom Handler not found"
    fi

    # Check AddRule method
    echo -e "${BLUE}Testing:${NC} AddRule method"
    if grep -q "func (i \*EventDrivenInvalidation) AddRule" "$invalidation_go"; then
        pass "AddRule method exists"
    else
        fail "AddRule method not found"
    fi

    # Check Start method for event listening
    echo -e "${BLUE}Testing:${NC} Start method for event listening"
    if grep -q "func (i \*EventDrivenInvalidation) Start" "$invalidation_go"; then
        pass "Start method for event listening"
    else
        fail "Start method not found"
    fi

    # Check Stop method
    echo -e "${BLUE}Testing:${NC} Stop method"
    if grep -q "func (i \*EventDrivenInvalidation) Stop" "$invalidation_go"; then
        pass "Stop method exists"
    else
        fail "Stop method not found"
    fi

    # Check event bus integration
    echo -e "${BLUE}Testing:${NC} Event bus integration"
    if grep -q "eventBus\|events.EventBus" "$invalidation_go"; then
        pass "Event bus integration exists"
    else
        fail "Event bus integration not found"
    fi

    # Check handleEvent method
    echo -e "${BLUE}Testing:${NC} handleEvent method"
    if grep -q "handleEvent" "$invalidation_go"; then
        pass "handleEvent method exists"
    else
        fail "handleEvent method not found"
    fi
}

# ============================================================================
# SECTION 6: Composite Invalidation Validation
# ============================================================================

validate_composite_invalidation() {
    section "Composite Invalidation Validation"

    local invalidation_go="$PROJECT_ROOT/internal/cache/invalidation.go"

    # Check CompositeInvalidation struct
    echo -e "${BLUE}Testing:${NC} CompositeInvalidation struct defined"
    if grep -q "type CompositeInvalidation struct" "$invalidation_go"; then
        pass "CompositeInvalidation struct defined"
    else
        fail "CompositeInvalidation struct not found"
    fi

    # Check NewCompositeInvalidation constructor
    echo -e "${BLUE}Testing:${NC} NewCompositeInvalidation constructor"
    if grep -q "func NewCompositeInvalidation" "$invalidation_go"; then
        pass "NewCompositeInvalidation constructor exists"
    else
        fail "NewCompositeInvalidation constructor not found"
    fi

    # Check AddStrategy method
    echo -e "${BLUE}Testing:${NC} AddStrategy method"
    if grep -q "func (c \*CompositeInvalidation) AddStrategy" "$invalidation_go"; then
        pass "AddStrategy method exists"
    else
        fail "AddStrategy method not found"
    fi

    # Check strategies slice
    echo -e "${BLUE}Testing:${NC} Multiple strategies support"
    if grep -q "strategies \[\]InvalidationStrategy" "$invalidation_go"; then
        pass "Multiple strategies support exists"
    else
        fail "Multiple strategies support not found"
    fi
}

# ============================================================================
# SECTION 7: Invalidation Metrics Validation
# ============================================================================

validate_invalidation_metrics() {
    section "Invalidation Metrics Validation"

    local invalidation_go="$PROJECT_ROOT/internal/cache/invalidation.go"

    # Check InvalidationMetrics struct
    echo -e "${BLUE}Testing:${NC} InvalidationMetrics struct defined"
    if grep -q "type InvalidationMetrics struct" "$invalidation_go"; then
        pass "InvalidationMetrics struct defined"
    else
        fail "InvalidationMetrics struct not found"
    fi

    # Check TotalInvalidations metric
    echo -e "${BLUE}Testing:${NC} TotalInvalidations metric"
    if grep -q "TotalInvalidations" "$invalidation_go"; then
        pass "TotalInvalidations metric exists"
    else
        fail "TotalInvalidations metric not found"
    fi

    # Check TagInvalidations metric
    echo -e "${BLUE}Testing:${NC} TagInvalidations metric"
    if grep -q "TagInvalidations" "$invalidation_go"; then
        pass "TagInvalidations metric exists"
    else
        fail "TagInvalidations metric not found"
    fi

    # Check PatternInvalidations metric
    echo -e "${BLUE}Testing:${NC} PatternInvalidations metric"
    if grep -q "PatternInvalidations" "$invalidation_go"; then
        pass "PatternInvalidations metric exists"
    else
        fail "PatternInvalidations metric not found"
    fi

    # Check EventInvalidations metric
    echo -e "${BLUE}Testing:${NC} EventInvalidations metric"
    if grep -q "EventInvalidations" "$invalidation_go"; then
        pass "EventInvalidations metric exists"
    else
        fail "EventInvalidations metric not found"
    fi

    # Check KeysInvalidated metric
    echo -e "${BLUE}Testing:${NC} KeysInvalidated metric"
    if grep -q "KeysInvalidated" "$invalidation_go"; then
        pass "KeysInvalidated metric exists"
    else
        fail "KeysInvalidated metric not found"
    fi

    # Check Metrics method
    echo -e "${BLUE}Testing:${NC} Metrics method"
    if grep -q "func.*Metrics()" "$invalidation_go"; then
        pass "Metrics method exists"
    else
        fail "Metrics method not found"
    fi
}

# ============================================================================
# SECTION 8: Atomic Operations Validation
# ============================================================================

validate_atomic_operations() {
    section "Atomic Operations Validation"

    local invalidation_go="$PROJECT_ROOT/internal/cache/invalidation.go"
    local tiered_cache="$PROJECT_ROOT/internal/cache/tiered_cache.go"

    # Check atomic metric updates
    echo -e "${BLUE}Testing:${NC} Atomic metric updates"
    if grep -q "atomic.AddInt64\|atomic.LoadInt64" "$invalidation_go"; then
        pass "Atomic metric updates in invalidation.go"
    else
        fail "Atomic metric updates not found in invalidation.go"
    fi

    # Check mutex protection for tag index
    echo -e "${BLUE}Testing:${NC} Mutex protection for tag index"
    if grep -qE "mu\s+sync.RWMutex" "$invalidation_go"; then
        pass "RWMutex protection for tag index"
    else
        fail "Mutex protection not found for tag index"
    fi

    # Check lock usage in AddTag
    echo -e "${BLUE}Testing:${NC} Lock usage in AddTag"
    if grep -A10 "func (i \*TagBasedInvalidation) AddTag" "$invalidation_go" | grep -q "i.mu.Lock()"; then
        pass "Lock used in AddTag"
    else
        fail "Lock not used in AddTag"
    fi

    # Check RLock usage in InvalidateByTag
    echo -e "${BLUE}Testing:${NC} RLock usage in InvalidateByTag"
    if grep -A10 "func (i \*TagBasedInvalidation) InvalidateByTag" "$invalidation_go" | grep -q "i.mu.RLock()"; then
        pass "RLock used in InvalidateByTag"
    else
        fail "RLock not used in InvalidateByTag"
    fi

    # Check atomic operations in tiered cache
    echo -e "${BLUE}Testing:${NC} Atomic operations in tiered cache"
    if grep -q "atomic.AddInt64" "$tiered_cache"; then
        pass "Atomic operations in tiered cache"
    else
        fail "Atomic operations not found in tiered cache"
    fi
}

# ============================================================================
# SECTION 9: Default Rules Validation
# ============================================================================

validate_default_rules() {
    section "Default Invalidation Rules Validation"

    local invalidation_go="$PROJECT_ROOT/internal/cache/invalidation.go"

    # Check registerDefaultRules method
    echo -e "${BLUE}Testing:${NC} registerDefaultRules method"
    if grep -q "registerDefaultRules" "$invalidation_go"; then
        pass "registerDefaultRules method exists"
    else
        fail "registerDefaultRules method not found"
    fi

    # Check provider health change rule
    echo -e "${BLUE}Testing:${NC} Provider health change invalidation rule"
    if grep -q "EventProviderHealthChanged" "$invalidation_go"; then
        pass "Provider health change rule exists"
    else
        fail "Provider health change rule not found"
    fi

    # Check MCP server disconnect rule
    echo -e "${BLUE}Testing:${NC} MCP server disconnect invalidation rule"
    if grep -q "EventMCPServerDisconnected" "$invalidation_go"; then
        pass "MCP server disconnect rule exists"
    else
        fail "MCP server disconnect rule not found"
    fi

    # Check cache invalidation event rule
    echo -e "${BLUE}Testing:${NC} Cache invalidation event rule"
    if grep -q "EventCacheInvalidated" "$invalidation_go"; then
        pass "Cache invalidation event rule exists"
    else
        fail "Cache invalidation event rule not found"
    fi
}

# ============================================================================
# SECTION 10: Thread Safety Validation
# ============================================================================

validate_thread_safety() {
    section "Thread Safety Validation"

    local invalidation_go="$PROJECT_ROOT/internal/cache/invalidation.go"

    # Check mutex in TagBasedInvalidation
    echo -e "${BLUE}Testing:${NC} Mutex in TagBasedInvalidation"
    if grep -A5 "type TagBasedInvalidation struct" "$invalidation_go" | grep -q "mu\|sync.RWMutex\|sync.Mutex"; then
        pass "Mutex in TagBasedInvalidation struct"
    else
        fail "Mutex not found in TagBasedInvalidation"
    fi

    # Check mutex in EventDrivenInvalidation
    echo -e "${BLUE}Testing:${NC} Mutex in EventDrivenInvalidation"
    if grep -A10 "type EventDrivenInvalidation struct" "$invalidation_go" | grep -q "mu.*sync"; then
        pass "Mutex in EventDrivenInvalidation struct"
    else
        fail "Mutex not found in EventDrivenInvalidation"
    fi

    # Check context cancellation
    echo -e "${BLUE}Testing:${NC} Context cancellation support"
    if grep -q "ctx.*context.Context\|cancel.*context.CancelFunc" "$invalidation_go"; then
        pass "Context cancellation support exists"
    else
        fail "Context cancellation support not found"
    fi

    # Run race detector
    echo -e "${BLUE}Testing:${NC} Race condition detection"
    if go test -race -short -run "Invalidat" "$PROJECT_ROOT/internal/cache/..." 2>&1 | grep -v "WARNING: DATA RACE" | tail -3 | grep -q "ok\|PASS"; then
        pass "No race conditions in invalidation"
    else
        fail "Race conditions detected in invalidation"
    fi
}

# ============================================================================
# SECTION 11: Invalidation Unit Tests Validation
# ============================================================================

validate_invalidation_tests() {
    section "Invalidation Unit Tests Validation"

    local test_files=(
        "internal/cache/invalidation_test.go"
        "internal/cache/invalidation_extended_test.go"
    )

    for test_file in "${test_files[@]}"; do
        local full_path="$PROJECT_ROOT/$test_file"
        if [ -f "$full_path" ]; then
            pass "Test file exists: $(basename $test_file)"
        else
            fail "Test file missing: $test_file"
        fi
    done

    # Run invalidation unit tests
    echo -e "${BLUE}Testing:${NC} Running invalidation unit tests"
    if go test -v -short -run "Invalidat" "$PROJECT_ROOT/internal/cache/..." 2>&1 | tail -5 | grep -q "ok\|PASS"; then
        pass "Invalidation unit tests pass"
    else
        fail "Invalidation unit tests failed"
    fi
}

# ============================================================================
# SECTION 12: Helper Functions Validation
# ============================================================================

validate_helper_functions() {
    section "Helper Functions Validation"

    local invalidation_go="$PROJECT_ROOT/internal/cache/invalidation.go"

    # Check containsWildcard function
    echo -e "${BLUE}Testing:${NC} containsWildcard helper function"
    if grep -q "func containsWildcard" "$invalidation_go"; then
        pass "containsWildcard helper function exists"
    else
        fail "containsWildcard helper function not found"
    fi

    # Check trimWildcard function
    echo -e "${BLUE}Testing:${NC} trimWildcard helper function"
    if grep -q "func trimWildcard" "$invalidation_go"; then
        pass "trimWildcard helper function exists"
    else
        fail "trimWildcard helper function not found"
    fi
}

# ============================================================================
# SECTION 13: GetTags and GetKeys Methods Validation
# ============================================================================

validate_query_methods() {
    section "Query Methods Validation"

    local invalidation_go="$PROJECT_ROOT/internal/cache/invalidation.go"

    # Check GetTags method
    echo -e "${BLUE}Testing:${NC} GetTags method"
    if grep -q "func (i \*TagBasedInvalidation) GetTags" "$invalidation_go"; then
        pass "GetTags method exists"
    else
        fail "GetTags method not found"
    fi

    # Check GetKeys method
    echo -e "${BLUE}Testing:${NC} GetKeys method"
    if grep -q "func (i \*TagBasedInvalidation) GetKeys" "$invalidation_go"; then
        pass "GetKeys method exists"
    else
        fail "GetKeys method not found"
    fi

    # Check RemoveRules method
    echo -e "${BLUE}Testing:${NC} RemoveRules method"
    if grep -q "func (i \*EventDrivenInvalidation) RemoveRules" "$invalidation_go"; then
        pass "RemoveRules method exists"
    else
        fail "RemoveRules method not found"
    fi
}

# ============================================================================
# SECTION 14: Integration with TieredCache Validation
# ============================================================================

validate_tieredcache_integration() {
    section "TieredCache Integration Validation"

    local tiered_cache="$PROJECT_ROOT/internal/cache/tiered_cache.go"

    # Check tag index in TieredCache
    echo -e "${BLUE}Testing:${NC} Tag index in TieredCache"
    if grep -q "tagIndex" "$tiered_cache"; then
        pass "Tag index integrated in TieredCache"
    else
        fail "Tag index not found in TieredCache"
    fi

    # Check tag removal on delete
    echo -e "${BLUE}Testing:${NC} Tag removal on delete"
    if grep -A20 "func (c \*TieredCache) Delete" "$tiered_cache" | grep -q "tagIndex.Remove\|c.tagIndex.Remove"; then
        pass "Tag index updated on delete"
    else
        fail "Tag index not updated on delete"
    fi

    # Check InvalidateByTag in TieredCache
    echo -e "${BLUE}Testing:${NC} InvalidateByTag in TieredCache"
    if grep -q "func (c \*TieredCache) InvalidateByTag" "$tiered_cache"; then
        pass "InvalidateByTag method in TieredCache"
    else
        fail "InvalidateByTag not found in TieredCache"
    fi

    # Check InvalidateByTags in TieredCache
    echo -e "${BLUE}Testing:${NC} InvalidateByTags in TieredCache"
    if grep -q "func (c \*TieredCache) InvalidateByTags" "$tiered_cache"; then
        pass "InvalidateByTags method in TieredCache"
    else
        fail "InvalidateByTags not found in TieredCache"
    fi

    # Check InvalidatePrefix in TieredCache
    echo -e "${BLUE}Testing:${NC} InvalidatePrefix in TieredCache"
    if grep -q "func (c \*TieredCache) InvalidatePrefix" "$tiered_cache"; then
        pass "InvalidatePrefix method in TieredCache"
    else
        fail "InvalidatePrefix not found in TieredCache"
    fi
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    validate_invalidation_interface
    validate_manual_invalidation
    validate_tag_invalidation
    validate_cascade_invalidation
    validate_event_driven_invalidation
    validate_composite_invalidation
    validate_invalidation_metrics
    validate_atomic_operations
    validate_default_rules
    validate_thread_safety
    validate_invalidation_tests
    validate_helper_functions
    validate_query_methods
    validate_tieredcache_integration

    # Summary
    echo ""
    echo "============================================"
    echo "  Cache Invalidation Challenge Results"
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
        echo -e "${GREEN}CACHE INVALIDATION CHALLENGE: PASSED${NC}"
        exit 0
    else
        echo -e "${RED}CACHE INVALIDATION CHALLENGE: FAILED${NC}"
        exit 1
    fi
}

main "$@"
