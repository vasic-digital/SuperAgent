#!/bin/bash
# Phase 6: Documentation Completion Challenge
# Validates all documentation requirements

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "=========================================="
echo "Phase 6: Documentation Completion Challenge"
echo "=========================================="

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

run_test() {
    local test_name="$1"
    local test_cmd="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo "Test $TOTAL_TESTS: $test_name"
    
    if eval "$test_cmd" > /tmp/phase6_test_output.txt 2>&1; then
        echo "✓ PASSED"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "✗ FAILED"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

file_check() {
    local test_name="$1"
    local file_path="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo "File Check $TOTAL_TESTS: $test_name"
    
    if [ -f "$PROJECT_ROOT/$file_path" ]; then
        echo "✓ PASSED - File exists: $file_path"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "✗ FAILED - File missing: $file_path"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

count_check() {
    local test_name="$1"
    local dir_path="$2"
    local min_count="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo "Count Check $TOTAL_TESTS: $test_name"
    
    actual_count=$(find "$PROJECT_ROOT/$dir_path" -name "*.md" -type f 2>/dev/null | wc -l)
    if [ "$actual_count" -ge "$min_count" ]; then
        echo "✓ PASSED - Found $actual_count files (min: $min_count)"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "✗ FAILED - Found $actual_count files (min: $min_count)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

cd "$PROJECT_ROOT"

echo ""
echo "=== Section 1: Core Documentation ==="

file_check "Root README.md" "README.md"
file_check "Root CLAUDE.md" "CLAUDE.md"
file_check "Root AGENTS.md" "AGENTS.md"
file_check "CONSTITUTION.json" "CONSTITUTION.json"

echo ""
echo "=== Section 2: API Documentation ==="

file_check "API Reference" "docs/api/API_REFERENCE.md"
file_check "OpenAPI Spec" "docs/api/openapi.yaml"
file_check "gRPC Documentation" "docs/api/grpc.md"
file_check "Big Data API Reference" "docs/api/BIG_DATA_API_REFERENCE.md"

echo ""
echo "=== Section 3: User Manuals (16 required) ==="

count_check "User manuals count" "Website/user-manuals" 16

run_test "User manual 01 exists" "ls $PROJECT_ROOT/Website/user-manuals/01-*.md >/dev/null 2>&1"
run_test "User manual 02 exists" "ls $PROJECT_ROOT/Website/user-manuals/02-*.md >/dev/null 2>&1"
run_test "User manual 03 exists" "ls $PROJECT_ROOT/Website/user-manuals/03-*.md >/dev/null 2>&1"
run_test "User manual 04 exists" "ls $PROJECT_ROOT/Website/user-manuals/04-*.md >/dev/null 2>&1"
run_test "User manual 05 exists" "ls $PROJECT_ROOT/Website/user-manuals/05-*.md >/dev/null 2>&1"
run_test "User manual 06 exists" "ls $PROJECT_ROOT/Website/user-manuals/06-*.md >/dev/null 2>&1"
run_test "User manual 07 exists" "ls $PROJECT_ROOT/Website/user-manuals/07-*.md >/dev/null 2>&1"
run_test "User manual 08 exists" "ls $PROJECT_ROOT/Website/user-manuals/08-*.md >/dev/null 2>&1"

echo ""
echo "=== Section 4: Video Courses (16 required) ==="

count_check "Video courses count" "Website/video-courses" 16

run_test "Video course 01 exists" "ls $PROJECT_ROOT/Website/video-courses/course-01-*.md >/dev/null 2>&1"
run_test "Video course 02 exists" "ls $PROJECT_ROOT/Website/video-courses/course-02-*.md >/dev/null 2>&1"
run_test "Video course 03 exists" "ls $PROJECT_ROOT/Website/video-courses/course-03-*.md >/dev/null 2>&1"
run_test "Video course 04 exists" "ls $PROJECT_ROOT/Website/video-courses/course-04-*.md >/dev/null 2>&1"
run_test "Video course 05 exists" "ls $PROJECT_ROOT/Website/video-courses/course-05-*.md >/dev/null 2>&1"
run_test "Video course 06 exists" "ls $PROJECT_ROOT/Website/video-courses/course-06-*.md >/dev/null 2>&1"
run_test "Video course 07 exists" "ls $PROJECT_ROOT/Website/video-courses/course-07-*.md >/dev/null 2>&1"
run_test "Video course 08 exists" "ls $PROJECT_ROOT/Website/video-courses/course-08-*.md >/dev/null 2>&1"

echo ""
echo "=== Section 5: Extracted Modules Documentation ==="

for module in Auth Cache Concurrency Containers Database Embeddings EventBus Formatters Memory Messaging Observability Optimization Plugins RAG Security Storage Streaming VectorDB; do
    file_check "$module README.md" "${module}/README.md"
    file_check "$module CLAUDE.md" "${module}/CLAUDE.md"
    file_check "$module AGENTS.md" "${module}/AGENTS.md"
done

echo ""
echo "=== Section 6: Architecture Documentation ==="

file_check "Architecture overview" "docs/ARCHITECTURE.md"
file_check "Modules documentation" "docs/MODULES.md"
file_check "Website architecture" "docs/website/ARCHITECTURE.md"

echo ""
echo "=== Section 7: Guides ==="

file_check "Deployment guide" "docs/guides/deployment-guide.md"
file_check "Contributing guide" "docs/CONTRIBUTING.md"
file_check "SpecKit user guide" "docs/guides/SPECKIT_USER_GUIDE.md"

echo ""
echo "=== Section 8: Test Documentation ==="

file_check "Test coverage plan" "docs/TEST_COVERAGE_REMEDIATION_PLAN.md"
file_check "Test validation report" "docs/TEST_VALIDATION_REPORT.md"

echo ""
echo "=== Section 9: Security Documentation ==="

file_check "Security scanning report" "docs/security/PHASE3_SECURITY_SCAN_REPORT.md"
file_check "Memory safety report" "docs/memory_safety/PHASE4_MEMORY_SAFETY_REPORT.md"

echo ""
echo "=== Section 10: Performance Documentation ==="

file_check "Performance report" "docs/performance/PHASE5_PERFORMANCE_REPORT.md"

echo ""
echo "=== Section 11: Website Build ==="

run_test "Website build script exists" "test -x $PROJECT_ROOT/Website/build.sh"
run_test "Website package.json exists" "test -f $PROJECT_ROOT/Website/package.json"

echo ""
echo "=========================================="
echo "Phase 6 Documentation Challenge Summary"
echo "=========================================="
echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo "✓ All Phase 6 documentation tests PASSED"
    exit 0
else
    echo "✗ Some Phase 6 documentation tests FAILED"
    exit 1
fi
