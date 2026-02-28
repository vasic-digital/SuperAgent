#!/bin/bash
# Comprehensive Test Execution Script
# Executes all tests with full coverage reporting

set -e

PROJECT_ROOT="/run/media/milosvasic/DATA4TB/Projects/HelixAgent"
LOG_DIR="$PROJECT_ROOT/execution_logs/$(date +%Y%m%d_%H%M%S)"
RESULTS_FILE="$LOG_DIR/test_results.txt"
COVERAGE_FILE="$LOG_DIR/coverage.out"

mkdir -p "$LOG_DIR"

echo "========================================"
echo "HelixAgent Comprehensive Test Execution"
echo "Started: $(date)"
echo "Log Directory: $LOG_DIR"
echo "========================================"

# Test packages in priority order
PACKAGES=(
    "internal/adapters/..."
    "internal/auth/..."
    "internal/cache"
    "internal/concurrency/..."
    "internal/config"
    "internal/database"
    "internal/debate/..."
    "internal/handlers"
    "internal/llm/..."
    "internal/services/..."
    "internal/verifier/..."
    "internal/background"
    "internal/bigdata"
    "internal/challenges"
    "internal/embedding"
    "internal/formatters/..."
    "internal/mcp/..."
    "internal/models"
    "internal/rag"
    "internal/security"
)

TOTAL=${#PACKAGES[@]}
PASSED=0
FAILED=0

for i in "${!PACKAGES[@]}"; do
    PKG="${PACKAGES[$i]}"
    NUM=$((i + 1))
    
    echo "[$NUM/$TOTAL] Testing: $PKG"
    
    if GOMAXPROCS=2 nice -n 19 go test -v -p 1 "$PKG" > "$LOG_DIR/${PKG//\//_}.log" 2>&1; then
        echo "  ✓ PASS"
        echo "PASS: $PKG" >> "$RESULTS_FILE"
        PASSED=$((PASSED + 1))
    else
        echo "  ✗ FAIL"
        echo "FAIL: $PKG" >> "$RESULTS_FILE"
        FAILED=$((FAILED + 1))
    fi
done

# Generate coverage report
echo ""
echo "Generating coverage report..."
GOMAXPROCS=2 nice -n 19 go test -coverprofile="$COVERAGE_FILE" ./internal/... > "$LOG_DIR/coverage.log" 2>&1 || true

echo ""
echo "========================================"
echo "Test Execution Complete"
echo "Passed: $PASSED/$TOTAL"
echo "Failed: $FAILED/$TOTAL"
echo "Results: $RESULTS_FILE"
echo "Coverage: $COVERAGE_FILE"
echo "========================================"
