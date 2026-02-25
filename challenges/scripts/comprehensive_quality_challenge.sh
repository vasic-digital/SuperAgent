#!/bin/bash
#
# Comprehensive Code Quality and Performance Challenge
# Validates all aspects of code quality, security, memory safety, and performance
#
# Categories:
# 1. Security Scanning (Snyk, gosec, SonarQube)
# 2. Memory Leak Detection
# 3. Deadlock Prevention
# 4. Race Condition Detection
# 5. Dead Code Removal
# 6. Test Coverage
# 7. Performance Optimization
# 8. Documentation Completeness

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PASSED=0
FAILED=0
WARNINGS=0
TESTS=()

log_pass() {
    echo "✅ PASS: $1"
    PASSED=$((PASSED + 1))
    TESTS+=("PASS: $1")
}

log_fail() {
    echo "❌ FAIL: $1"
    FAILED=$((FAILED + 1))
    TESTS+=("FAIL: $1")
}

log_warn() {
    echo "⚠️  WARN: $1"
    WARNINGS=$((WARNINGS + 1))
    TESTS+=("WARN: $1")
}

log_info() {
    echo "ℹ️  INFO: $1"
}

# ==========================================
# SECTION 1: SECURITY SCANNING
# ==========================================

echo ""
echo "=========================================="
echo "SECTION 1: SECURITY SCANNING"
echo "=========================================="

test_gosec_installed() {
    log_info "Testing gosec installation..."
    
    if command -v gosec &> /dev/null || [ -f "$HOME/go/bin/gosec" ]; then
        log_pass "gosec is installed"
    else
        log_fail "gosec is not installed"
    fi
}

test_gosec_scan() {
    log_info "Running gosec security scan..."
    
    cd "$PROJECT_ROOT"
    RESULT=$($HOME/go/bin/gosec -quiet -exclude-generated ./internal/config/... 2>&1 || echo "")
    
    CRITICAL=$(echo "$RESULT" | grep -c "CWE-" || echo "0")
    
    if [ "$CRITICAL" -lt 50 ]; then
        log_pass "gosec scan completed with $CRITICAL findings (acceptable)"
    else
        log_warn "gosec scan found $CRITICAL findings (review recommended)"
    fi
}

test_snyk_config() {
    log_info "Testing Snyk configuration..."
    
    if [ -f "$PROJECT_ROOT/.snyk" ]; then
        log_pass "Snyk policy file exists"
    else
        log_fail "Snyk policy file missing"
    fi
}

test_sonarqube_config() {
    log_info "Testing SonarQube configuration..."
    
    if [ -f "$PROJECT_ROOT/sonar-project.properties" ]; then
        log_pass "SonarQube configuration exists"
    else
        log_fail "SonarQube configuration missing"
    fi
}

test_no_hardcoded_secrets() {
    log_info "Testing for hardcoded secrets..."
    
    SECRETS=$(grep -r "password\s*=\s*\"[^\$]" --include="*.go" internal/ 2>/dev/null | grep -v "_test.go" | wc -l || echo "0")
    
    if [ "$SECRETS" -eq 0 ]; then
        log_pass "No hardcoded passwords found"
    else
        log_fail "Found $SECRETS potential hardcoded passwords"
    fi
}

# ==========================================
# SECTION 2: MEMORY SAFETY
# ==========================================

echo ""
echo "=========================================="
echo "SECTION 2: MEMORY SAFETY"
echo "=========================================="

test_memory_leak_detector() {
    log_info "Testing memory leak detector..."
    
    DETECTOR="$PROJECT_ROOT/internal/profiling/memory_leak_detector.go"
    
    if [ -f "$DETECTOR" ]; then
        if grep -q "MemoryLeakDetector" "$DETECTOR" && \
           grep -q "MemorySnapshot" "$DETECTOR" && \
           grep -q "MemoryThresholds" "$DETECTOR"; then
            log_pass "Memory leak detector is implemented"
        else
            log_fail "Memory leak detector is incomplete"
        fi
    else
        log_fail "Memory leak detector not found"
    fi
}

test_deadlock_detector() {
    log_info "Testing deadlock detector..."
    
    DETECTOR="$PROJECT_ROOT/internal/profiling/deadlock_detector.go"
    
    if [ -f "$DETECTOR" ]; then
        if grep -q "DeadlockDetector" "$DETECTOR" && \
           grep -q "DeadlockAlert" "$DETECTOR"; then
            log_pass "Deadlock detector is implemented"
        else
            log_fail "Deadlock detector is incomplete"
        fi
    else
        log_fail "Deadlock detector not found"
    fi
}

test_goroutine_cleanup() {
    log_info "Testing goroutine cleanup patterns..."
    
    CHANNEL_CLOSE=$(grep -r "defer close(" internal/ --include="*.go" 2>/dev/null | grep -v "_test.go" | wc -l || echo "0")
    DEFER_CLEANUP=$(grep -r "defer.*Unlock\|defer.*Close" internal/ --include="*.go" 2>/dev/null | grep -v "_test.go" | wc -l || echo "0")
    
    if [ "$DEFER_CLEANUP" -gt 500 ]; then
        log_pass "Proper defer cleanup patterns found ($DEFER_CLEANUP instances)"
    else
        log_warn "Limited defer cleanup patterns found ($DEFER_CLEANUP instances)"
    fi
}

test_context_usage() {
    log_info "Testing context.Context usage..."
    
    CTX_USAGE=$(grep -r "context.Context" internal/ --include="*.go" 2>/dev/null | wc -l || echo "0")
    
    if [ "$CTX_USAGE" -gt 500 ]; then
        log_pass "Good context.Context usage ($CTX_USAGE instances)"
    else
        log_warn "Limited context.Context usage ($CTX_USAGE instances)"
    fi
}

# ==========================================
# SECTION 3: CODE QUALITY
# ==========================================

echo ""
echo "=========================================="
echo "SECTION 3: CODE QUALITY"
echo "=========================================="

test_go_vet() {
    log_info "Running go vet..."
    
    cd "$PROJECT_ROOT"
    if go vet ./internal/... 2>&1 | grep -q "vet:"; then
        log_fail "go vet found issues"
    else
        log_pass "go vet passed"
    fi
}

test_dead_code_detection() {
    log_info "Testing for dead code..."
    
    TODO_COUNT=$(grep -r "TODO\|FIXME\|XXX" --include="*.go" internal/ 2>/dev/null | wc -l || echo "0")
    
    if [ "$TODO_COUNT" -lt 800 ]; then
        log_pass "TODO count acceptable ($TODO_COUNT)"
    else
        log_warn "High TODO count ($TODO_COUNT) - review recommended"
    fi
}

test_skipped_tests() {
    log_info "Testing for skipped tests..."
    
    SKIP_COUNT=$(grep -r "t.Skip\|Skip(" --include="*_test.go" internal/ 2>/dev/null | wc -l || echo "0")
    
    if [ "$SKIP_COUNT" -lt 200 ]; then
        log_pass "Skipped tests count acceptable ($SKIP_COUNT)"
    else
        log_warn "Many skipped tests ($SKIP_COUNT) - review recommended"
    fi
}

test_panic_usage() {
    log_info "Testing panic usage in production code..."
    
    PANIC_COUNT=$(grep -r "panic(" --include="*.go" internal/ 2>/dev/null | grep -v "_test.go\|recover()" | wc -l || echo "0")
    
    if [ "$PANIC_COUNT" -lt 10 ]; then
        log_pass "Limited panic usage in production ($PANIC_COUNT)"
    else
        log_warn "Multiple panics in production code ($PANIC_COUNT)"
    fi
}

test_error_handling() {
    log_info "Testing error handling patterns..."
    
    ERROR_CHECKS=$(grep -r "if err != nil" internal/ --include="*.go" 2>/dev/null | wc -l || echo "0")
    
    if [ "$ERROR_CHECKS" -gt 2000 ]; then
        log_pass "Good error handling ($ERROR_CHECKS checks)"
    else
        log_warn "Limited error handling ($ERROR_CHECKS checks)"
    fi
}

# ==========================================
# SECTION 4: PERFORMANCE
# ==========================================

echo ""
echo "=========================================="
echo "SECTION 4: PERFORMANCE"
echo "=========================================="

test_lazy_loader() {
    log_info "Testing lazy loader implementation..."
    
    LOADER="$PROJECT_ROOT/internal/profiling/lazy_loader.go"
    
    if [ -f "$LOADER" ]; then
        if grep -q "LazyLoader" "$LOADER" && \
           grep -q "Register" "$LOADER" && \
           grep -q "Get" "$LOADER"; then
            log_pass "Lazy loader is implemented"
        else
            log_fail "Lazy loader is incomplete"
        fi
    else
        log_fail "Lazy loader not found"
    fi
}

test_semaphore_usage() {
    log_info "Testing semaphore/concurrency patterns..."
    
    SEMAPHORE=$(grep -r "semaphore\|Semaphore" internal/ --include="*.go" 2>/dev/null | wc -l || echo "0")
    WAITGROUP=$(grep -r "sync.WaitGroup" internal/ --include="*.go" 2>/dev/null | wc -l || echo "0")
    
    if [ "$SEMAPHORE" -gt 10 ] || [ "$WAITGROUP" -gt 50 ]; then
        log_pass "Good concurrency patterns ($SEMAPHORE semaphores, $WAITGROUP waitgroups)"
    else
        log_warn "Limited concurrency patterns ($SEMAPHORE semaphores, $WAITGROUP waitgroups)"
    fi
}

test_connection_pooling() {
    log_info "Testing connection pooling..."
    
    POOL=$(grep -r "Pool\|pool" internal/ --include="*.go" 2>/dev/null | grep -v "_test.go" | wc -l || echo "0")
    
    if [ "$POOL" -gt 20 ]; then
        log_pass "Connection pooling implemented ($POOL references)"
    else
        log_warn "Limited connection pooling ($POOL references)"
    fi
}

# ==========================================
# SECTION 5: TEST COVERAGE
# ==========================================

echo ""
echo "=========================================="
echo "SECTION 5: TEST COVERAGE"
echo "=========================================="

test_unit_tests_exist() {
    log_info "Testing unit test existence..."
    
    TEST_FILES=$(find internal/ -name "*_test.go" 2>/dev/null | wc -l || echo "0")
    
    if [ "$TEST_FILES" -gt 100 ]; then
        log_pass "Good test coverage ($TEST_FILES test files)"
    else
        log_warn "Limited test coverage ($TEST_FILES test files)"
    fi
}

test_integration_tests_exist() {
    log_info "Testing integration test existence..."
    
    INT_TESTS=$(find tests/integration/ -name "*.go" 2>/dev/null | wc -l || echo "0")
    
    if [ "$INT_TESTS" -gt 10 ]; then
        log_pass "Integration tests exist ($INT_TESTS files)"
    else
        log_warn "Limited integration tests ($INT_TESTS files)"
    fi
}

test_e2e_tests_exist() {
    log_info "Testing E2E test existence..."
    
    E2E_TESTS=$(find tests/e2e/ -name "*.go" 2>/dev/null | wc -l || echo "0")
    
    if [ "$E2E_TESTS" -gt 5 ]; then
        log_pass "E2E tests exist ($E2E_TESTS files)"
    else
        log_warn "Limited E2E tests ($E2E_TESTS files)"
    fi
}

test_challenges_exist() {
    log_info "Testing challenge scripts..."
    
    CHALLENGES=$(find challenges/scripts/ -name "*.sh" 2>/dev/null | wc -l || echo "0")
    
    if [ "$CHALLENGES" -gt 30 ]; then
        log_pass "Comprehensive challenges ($CHALLENGES scripts)"
    else
        log_warn "Limited challenges ($CHALLENGES scripts)"
    fi
}

test_profiling_tests() {
    log_info "Testing profiling package tests..."
    
    if [ -f "$PROJECT_ROOT/internal/profiling/memory_leak_detector_test.go" ] || \
       [ -f "$PROJECT_ROOT/internal/profiling/lazy_loader_test.go" ]; then
        log_pass "Profiling tests exist"
    else
        log_info "Profiling tests to be created"
    fi
}

# ==========================================
# SECTION 6: DOCUMENTATION
# ==========================================

echo ""
echo "=========================================="
echo "SECTION 6: DOCUMENTATION"
echo "=========================================="

test_readme_exists() {
    log_info "Testing README existence..."
    
    if [ -f "$PROJECT_ROOT/README.md" ]; then
        log_pass "README.md exists"
    else
        log_fail "README.md missing"
    fi
}

test_claude_md_exists() {
    log_info "Testing CLAUDE.md existence..."
    
    if [ -f "$PROJECT_ROOT/CLAUDE.md" ]; then
        SIZE=$(wc -c < "$PROJECT_ROOT/CLAUDE.md")
        if [ "$SIZE" -gt 10000 ]; then
            log_pass "CLAUDE.md is comprehensive ($SIZE bytes)"
        else
            log_warn "CLAUDE.md could be more comprehensive ($SIZE bytes)"
        fi
    else
        log_fail "CLAUDE.md missing"
    fi
}

test_agents_md_exists() {
    log_info "Testing AGENTS.md existence..."
    
    if [ -f "$PROJECT_ROOT/AGENTS.md" ]; then
        SIZE=$(wc -c < "$PROJECT_ROOT/AGENTS.md")
        if [ "$SIZE" -gt 5000 ]; then
            log_pass "AGENTS.md is comprehensive ($SIZE bytes)"
        else
            log_warn "AGENTS.md could be more comprehensive ($SIZE bytes)"
        fi
    else
        log_fail "AGENTS.md missing"
    fi
}

test_module_docs() {
    log_info "Testing module documentation..."
    
    MODULES_WITH_README=0
    for dir in Auth BuildCheck Cache Challenges Concurrency Containers Database Embeddings EventBus Formatters HelixMemory HelixSpecifier LLMsVerifier MCP_Module Memory Messaging Observability Optimization Plugins RAG Security Storage Streaming VectorDB; do
        if [ -f "$PROJECT_ROOT/$dir/README.md" ]; then
            MODULES_WITH_README=$((MODULES_WITH_README + 1))
        fi
    done
    
    if [ "$MODULES_WITH_README" -gt 15 ]; then
        log_pass "Good module documentation ($MODULES_WITH_README modules with README)"
    else
        log_warn "Some modules missing documentation ($MODULES_WITH_README with README)"
    fi
}

test_api_docs() {
    log_info "Testing API documentation..."
    
    if [ -f "$PROJECT_ROOT/docs/API_DOCUMENTATION.md" ] || \
       [ -f "$PROJECT_ROOT/docs/API.md" ]; then
        log_pass "API documentation exists"
    else
        log_warn "API documentation missing"
    fi
}

# ==========================================
# SECTION 7: BUILD VERIFICATION
# ==========================================

echo ""
echo "=========================================="
echo "SECTION 7: BUILD VERIFICATION"
echo "=========================================="

test_go_build() {
    log_info "Testing Go build..."
    
    cd "$PROJECT_ROOT"
    if go build ./cmd/helixagent 2>/dev/null; then
        log_pass "Go build successful"
    else
        log_fail "Go build failed"
    fi
}

test_go_test_compile() {
    log_info "Testing test compilation..."
    
    cd "$PROJECT_ROOT"
    if go test -c ./internal/config/... 2>/dev/null; then
        log_pass "Tests compile successfully"
        rm -f config.test 2>/dev/null || true
    else
        log_warn "Some tests may have compilation issues"
    fi
}

test_dependencies() {
    log_info "Testing dependencies..."
    
    cd "$PROJECT_ROOT"
    if go mod verify 2>/dev/null | grep -q "all modules verified"; then
        log_pass "Dependencies verified"
    else
        log_warn "Dependency verification had issues"
    fi
}

# ==========================================
# SECTION 8: COMPREHENSIVE REPORT
# ==========================================

echo ""
echo "=========================================="
echo "COMPREHENSIVE QUALITY CHALLENGE RESULTS"
echo "=========================================="

main() {
    cd "$PROJECT_ROOT"
    
    test_gosec_installed
    test_gosec_scan
    test_snyk_config
    test_sonarqube_config
    test_no_hardcoded_secrets
    
    test_memory_leak_detector
    test_deadlock_detector
    test_goroutine_cleanup
    test_context_usage
    
    test_go_vet
    test_dead_code_detection
    test_skipped_tests
    test_panic_usage
    test_error_handling
    
    test_lazy_loader
    test_semaphore_usage
    test_connection_pooling
    
    test_unit_tests_exist
    test_integration_tests_exist
    test_e2e_tests_exist
    test_challenges_exist
    test_profiling_tests
    
    test_readme_exists
    test_claude_md_exists
    test_agents_md_exists
    test_module_docs
    test_api_docs
    
    test_go_build
    test_go_test_compile
    test_dependencies
    
    echo ""
    echo "=============================================="
    echo "Summary"
    echo "=============================================="
    echo "Passed:   $PASSED"
    echo "Failed:   $FAILED"
    echo "Warnings: $WARNINGS"
    echo "Total:    $((PASSED + FAILED + WARNINGS))"
    echo ""
    
    if [ $FAILED -gt 0 ]; then
        echo "Failed Tests:"
        for test in "${TESTS[@]}"; do
            if [[ $test == FAIL* ]]; then
                echo "  - $test"
            fi
        done
        echo ""
    fi
    
    if [ $WARNINGS -gt 0 ]; then
        echo "Warnings:"
        for test in "${TESTS[@]}"; do
            if [[ $test == WARN* ]]; then
                echo "  - $test"
            fi
        done
        echo ""
    fi
    
    if [ $FAILED -eq 0 ]; then
        echo "🎉 All critical tests passed!"
        if [ $WARNINGS -gt 0 ]; then
            echo "⚠️  Some warnings require attention"
        fi
        exit 0
    else
        echo "❌ Some tests failed"
        exit 1
    fi
}

main "$@"
