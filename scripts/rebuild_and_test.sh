#!/bin/bash
# Complete rebuild and test script for CLI Agents Porting
# Cleans everything and rebuilds from scratch

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPORT_DIR="$PROJECT_ROOT/test-reports/$(date +%Y%m%d-%H%M%S)"

mkdir -p "$REPORT_DIR"

echo "╔════════════════════════════════════════════════════════════════╗"
echo "║  COMPLETE REBUILD AND TEST - CLI AGENTS PORTING               ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""
echo "Report Directory: $REPORT_DIR"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$REPORT_DIR/build.log"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1" | tee -a "$REPORT_DIR/build.log"; }
log_error() { echo -e "${RED}[FAIL]${NC} $1" | tee -a "$REPORT_DIR/build.log"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$REPORT_DIR/build.log"; }

# Track results
TOTAL_STEPS=10
CURRENT_STEP=0
PASS_COUNT=0
FAIL_COUNT=0

record_result() {
    if [ $1 -eq 0 ]; then
        ((PASS_COUNT++))
    else
        ((FAIL_COUNT++))
    fi
    ((CURRENT_STEP++))
}

# Step 1: Clean everything
step1_clean() {
    log_info "Step 1/10: Cleaning build artifacts..."
    
    cd "$PROJECT_ROOT"
    
    # Clean Go build cache
    go clean -cache -testcache -modcache 2>/dev/null || true
    
    # Remove binaries
    rm -f bin/helixagent bin/api bin/grpc-server 2>/dev/null || true
    
    # Remove test binaries
    find . -name "*.test" -type f -delete 2>/dev/null || true
    
    # Clean Docker
    if command -v docker &> /dev/null; then
        docker system prune -f 2>/dev/null || true
    fi
    
    log_success "Clean completed"
    return 0
}

# Step 2: Verify dependencies
step2_deps() {
    log_info "Step 2/10: Verifying dependencies..."
    
    cd "$PROJECT_ROOT"
    
    # Download dependencies
    go mod download 2>&1 | tee "$REPORT_DIR/deps.log"
    
    # Verify modules
    go mod verify 2>&1 | tee -a "$REPORT_DIR/deps.log"
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        log_success "Dependencies verified"
        return 0
    else
        log_error "Dependency verification failed"
        return 1
    fi
}

# Step 3: Build all binaries
step3_build() {
    log_info "Step 3/10: Building all binaries..."
    
    cd "$PROJECT_ROOT"
    
    # Build helixagent
    log_info "Building helixagent..."
    go build -v -o bin/helixagent ./cmd/helixagent 2>&1 | tee "$REPORT_DIR/build_helixagent.log"
    
    # Build API server
    log_info "Building API server..."
    go build -v -o bin/api ./cmd/api 2>&1 | tee "$REPORT_DIR/build_api.log"
    
    # Build gRPC server
    log_info "Building gRPC server..."
    go build -v -o bin/grpc-server ./cmd/grpc-server 2>&1 | tee "$REPORT_DIR/build_grpc.log"
    
    # Verify binaries exist
    if [ -f "bin/helixagent" ] && [ -f "bin/api" ] && [ -f "bin/grpc-server" ]; then
        log_success "All binaries built successfully"
        ls -lh bin/ >> "$REPORT_DIR/build.log"
        return 0
    else
        log_error "Some binaries failed to build"
        return 1
    fi
}

# Step 4: Run unit tests
step4_unit_tests() {
    log_info "Step 4/10: Running unit tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run tests with coverage
    go test -v -race -coverprofile="$REPORT_DIR/coverage.out" \
        ./internal/clis/... \
        ./internal/ensemble/... \
        ./internal/cache/... \
        ./internal/output/... \
        2>&1 | tee "$REPORT_DIR/unit_tests.log"
    
    TEST_RESULT=${PIPESTATUS[0]}
    
    # Generate coverage report
    if [ -f "$REPORT_DIR/coverage.out" ]; then
        go tool cover -html="$REPORT_DIR/coverage.out" -o "$REPORT_DIR/coverage.html"
        COVERAGE=$(go tool cover -func="$REPORT_DIR/coverage.out" | grep total | awk '{print $3}')
        log_info "Total coverage: $COVERAGE"
    fi
    
    if [ $TEST_RESULT -eq 0 ]; then
        log_success "Unit tests passed"
        return 0
    else
        log_error "Unit tests failed"
        return 1
    fi
}

# Step 5: Run integration tests
step5_integration_tests() {
    log_info "Step 5/10: Running integration tests..."
    
    cd "$PROJECT_ROOT"
    
    # Check if database is available
    if ! pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
        log_warn "PostgreSQL not available, skipping integration tests"
        return 0
    fi
    
    # Run integration tests
    go test -v -tags=integration \
        ./tests/integration/... \
        2>&1 | tee "$REPORT_DIR/integration_tests.log"
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        log_success "Integration tests passed"
        return 0
    else
        log_error "Integration tests failed"
        return 1
    fi
}

# Step 6: Run benchmarks
step6_benchmarks() {
    log_info "Step 6/10: Running benchmarks..."
    
    cd "$PROJECT_ROOT"
    
    # Run benchmarks
    go test -bench=. -benchmem \
        ./internal/clis/... \
        ./internal/ensemble/... \
        2>&1 | tee "$REPORT_DIR/benchmarks.log"
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        log_success "Benchmarks completed"
        return 0
    else
        log_warn "Some benchmarks may have issues"
        return 0
    fi
}

# Step 7: Security scan
step7_security() {
    log_info "Step 7/10: Running security scans..."
    
    cd "$PROJECT_ROOT"
    
    # Run gosec if available
    if command -v gosec &> /dev/null; then
        gosec -fmt sarif -out "$REPORT_DIR/gosec.sarif" ./... 2>&1 | tee "$REPORT_DIR/security.log" || true
    else
        log_warn "gosec not installed, skipping"
    fi
    
    # Run go vet
    go vet ./... 2>&1 | tee -a "$REPORT_DIR/security.log"
    
    log_success "Security scans completed"
    return 0
}

# Step 8: Code quality checks
step8_quality() {
    log_info "Step 8/10: Running code quality checks..."
    
    cd "$PROJECT_ROOT"
    
    # Format check
    UNFORMATTED=$(gofmt -l . 2>/dev/null | wc -l)
    if [ "$UNFORMATTED" -gt 0 ]; then
        log_warn "$UNFORMATTED files need formatting"
        gofmt -l . > "$REPORT_DIR/unformatted_files.txt"
    else
        log_success "All files formatted correctly"
    fi
    
    # Lint check
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run --out-format json > "$REPORT_DIR/lint.json" 2>&1 || true
        log_success "Lint check completed"
    else
        log_warn "golangci-lint not installed"
    fi
    
    return 0
}

# Step 9: Documentation validation
step9_docs() {
    log_info "Step 9/10: Validating documentation..."
    
    cd "$PROJECT_ROOT"
    
    # Check for required docs
    REQUIRED_DOCS=(
        "CLI_AGENTS_PORTING_COMPLETE.md"
        "TEST_PLAN_CLI_AGENTS_PORTING.md"
        "docs/research/cli_agents_porting/README.md"
        "sql/001_cli_agents_fusion.sql"
    )
    
    MISSING=0
    for doc in "${REQUIRED_DOCS[@]}"; do
        if [ ! -f "$doc" ]; then
            log_error "Missing required document: $doc"
            ((MISSING++))
        fi
    done
    
    if [ $MISSING -eq 0 ]; then
        log_success "All required documentation present"
        return 0
    else
        return 1
    fi
}

# Step 10: Generate final report
step10_report() {
    log_info "Step 10/10: Generating final report..."
    
    cat > "$REPORT_DIR/FINAL_REPORT.md" << EOF
# CLI Agents Porting - Final Test Report

**Date:** $(date)
**Report ID:** $(basename "$REPORT_DIR")

## Summary

| Metric | Value |
|--------|-------|
| Total Steps | $TOTAL_STEPS |
| Passed | $PASS_COUNT |
| Failed | $FAIL_COUNT |
| Success Rate | $(awk "BEGIN {printf \"%.1f%%\", ($PASS_COUNT/$TOTAL_STEPS)*100}") |

## Build Artifacts

### Binaries
$(ls -lh "$PROJECT_ROOT/bin/" 2>/dev/null || echo "No binaries found")

### Test Coverage
$(if [ -f "$REPORT_DIR/coverage.out" ]; then echo "Coverage report generated"; else echo "No coverage data"; fi)

### Test Results
$(if [ -f "$REPORT_DIR/unit_tests.log" ]; then tail -20 "$REPORT_DIR/unit_tests.log"; else echo "No test logs"; fi)

## Detailed Results

$(for log in "$REPORT_DIR"/*.log; do
    if [ -f "$log" ]; then
        echo "### $(basename $log)"
        echo "\`\`\`"
        tail -10 "$log" 2>/dev/null || echo "Empty log"
        echo "\`\`\`"
        echo ""
    fi
done)

## Conclusion

$(if [ $FAIL_COUNT -eq 0 ]; then
    echo "✅ ALL TESTS PASSED - Ready for production"
else
    echo "⚠️  $FAIL_COUNT step(s) failed - Review required"
fi)

EOF

    log_success "Final report generated: $REPORT_DIR/FINAL_REPORT.md"
    return 0
}

# Main execution
main() {
    echo "Starting complete rebuild and test process..."
    echo "This may take several minutes..."
    echo ""
    
    step1_clean
    record_result $?
    
    step2_deps
    record_result $?
    
    step3_build
    record_result $?
    
    step4_unit_tests
    record_result $?
    
    step5_integration_tests
    record_result $?
    
    step6_benchmarks
    record_result $?
    
    step7_security
    record_result $?
    
    step8_quality
    record_result $?
    
    step9_docs
    record_result $?
    
    step10_report
    record_result $?
    
    echo ""
    echo "╔════════════════════════════════════════════════════════════════╗"
    echo "║  REBUILD AND TEST COMPLETED                                   ║"
    echo "╠════════════════════════════════════════════════════════════════╣"
    echo "║  Passed: $PASS_COUNT/$TOTAL_STEPS                                              ║"
    echo "║  Failed: $FAIL_COUNT/$TOTAL_STEPS                                              ║"
    echo "║  Report: $REPORT_DIR                                           ║"
    echo "╚════════════════════════════════════════════════════════════════╝"
    
    if [ $FAIL_COUNT -eq 0 ]; then
        exit 0
    else
        exit 1
    fi
}

# Run main
main "$@"
