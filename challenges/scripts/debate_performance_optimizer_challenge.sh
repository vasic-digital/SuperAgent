#!/bin/bash
set -uo pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

echo "=================================================="
echo "Debate Performance Optimizer Challenge"
echo "=================================================="
echo ""

check_file() {
    local name="$1"
    local file="$2"
    local pattern="$3"
    
    if [[ -f "$file" ]] && grep -q "$pattern" "$file"; then
        echo -e "${GREEN}✓ PASS${NC}: $name"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: $name"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

echo "Test 1: Performance optimizer file exists"
check_file "Optimizer file exists" "internal/services/debate_performance_optimizer.go" "DebatePerformanceOptimizer"

echo ""
echo "Test 2: Default configuration validation"
check_file "Default config function" "internal/services/debate_performance_optimizer.go" "DefaultDebateOptimizationConfig"

echo ""
echo "Test 3: Verify optimization features in config"
check_file "EnableParallelExecution" "internal/services/debate_performance_optimizer.go" "EnableParallelExecution"
check_file "EnableResponseCaching" "internal/services/debate_performance_optimizer.go" "EnableResponseCaching"
check_file "EnableEarlyTermination" "internal/services/debate_performance_optimizer.go" "EnableEarlyTermination"
check_file "EnableSmartFallback" "internal/services/debate_performance_optimizer.go" "EnableSmartFallback"
check_file "MaxParallelRequests" "internal/services/debate_performance_optimizer.go" "MaxParallelRequests"
check_file "CacheTTL" "internal/services/debate_performance_optimizer.go" "CacheTTL"
check_file "EarlyTerminationThreshold" "internal/services/debate_performance_optimizer.go" "EarlyTerminationThreshold"

echo ""
echo "Test 4: Verify test file exists"
check_file "Test file exists" "internal/services/debate_performance_optimizer_test.go" "TestDebatePerformanceOptimizer"

echo ""
echo "Test 5: Test coverage check"
check_file "CacheHit test" "internal/services/debate_performance_optimizer_test.go" "CacheHit"
check_file "CacheMiss test" "internal/services/debate_performance_optimizer_test.go" "CacheMiss"
check_file "SmartFallback test" "internal/services/debate_performance_optimizer_test.go" "SmartFallback"
check_file "EarlyTermination test" "internal/services/debate_performance_optimizer_test.go" "EarlyTermination"
check_file "Parallel test" "internal/services/debate_performance_optimizer_test.go" "ExecuteParallel"

echo ""
echo "Test 6: Verify integration in DebateService"
check_file "Performance optimizer field" "internal/services/debate_service.go" "performanceOptimizer"

echo ""
echo "Test 7: Verify accessor methods in DebateService"
check_file "GetPerformanceOptimizerStats" "internal/services/debate_service.go" "GetPerformanceOptimizerStats"
check_file "ClearPerformanceOptimizerCache" "internal/services/debate_service.go" "ClearPerformanceOptimizerCache"
check_file "CheckEarlyTermination" "internal/services/debate_service.go" "CheckEarlyTermination"

echo ""
echo "Test 8: Verify caching functions"
check_file "getCachedResponse" "internal/services/debate_performance_optimizer.go" "getCachedResponse"
check_file "cacheResponse" "internal/services/debate_performance_optimizer.go" "cacheResponse"
check_file "generateCacheKey" "internal/services/debate_performance_optimizer.go" "generateCacheKey"
check_file "ClearCache" "internal/services/debate_performance_optimizer.go" "ClearCache"

echo ""
echo "Test 9: Verify smart fallback implementation"
check_file "Smart fallback function" "internal/services/debate_performance_optimizer.go" "executeWithSmartFallback"

echo ""
echo "Test 10: Verify early termination implementation"
check_file "Early termination function" "internal/services/debate_performance_optimizer.go" "ShouldTerminateEarly"

echo ""
echo "Test 11: Verify parallel execution"
check_file "Parallel execution function" "internal/services/debate_performance_optimizer.go" "ExecuteParallel"

echo ""
echo "Test 12: Verify stats tracking"
check_file "CacheHits" "internal/services/debate_performance_optimizer.go" "CacheHits"
check_file "CacheMisses" "internal/services/debate_performance_optimizer.go" "CacheMisses"
check_file "ParallelRequests" "internal/services/debate_performance_optimizer.go" "ParallelRequests"
check_file "EarlyTerminations" "internal/services/debate_performance_optimizer.go" "EarlyTerminations"
check_file "FallbacksTriggered" "internal/services/debate_performance_optimizer.go" "FallbacksTriggered"
check_file "TotalRequests" "internal/services/debate_performance_optimizer.go" "TotalRequests"

echo ""
echo "Test 13: Verify Claude 4.6 models in team config"
check_file "Claude 4.6 Sonnet" "internal/services/debate_team_config.go" "Sonnet46"
check_file "Claude 4.6 Opus" "internal/services/debate_team_config.go" "Opus46"

echo ""
echo "Test 14: Verify Claude CLI timeout"
check_file "Claude CLI timeout" "internal/llm/providers/claude/claude_cli.go" "180"

echo ""
echo "Test 15: Build verification"
if go build ./internal/services/... 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}: Services package builds successfully"
    ((PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}: Services package build failed"
    ((FAILED++))
fi

echo ""
echo "=================================================="
echo "Results: ${GREEN}${PASSED} passed${NC}, ${RED}${FAILED} failed${NC}"
echo "=================================================="

if [[ $FAILED -gt 0 ]]; then
    exit 1
fi

exit 0
