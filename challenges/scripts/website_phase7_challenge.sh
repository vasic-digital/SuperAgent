#!/bin/bash
# Phase 7: Website Update Challenge
# Validates website content and build

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "=========================================="
echo "Phase 7: Website Update Challenge"
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
    
    if eval "$test_cmd" > /tmp/phase7_test_output.txt 2>&1; then
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

cd "$PROJECT_ROOT"

echo ""
echo "=== Section 1: Website Structure ==="

run_test "Website directory exists" "test -d $PROJECT_ROOT/Website"
file_check "Website build script" "Website/build.sh"
file_check "Website package.json" "Website/package.json"

echo ""
echo "=== Section 2: Website Content ==="

file_check "User manuals directory" "Website/user-manuals/README.md"
file_check "Video courses directory" "Website/video-courses/README.md"

echo ""
echo "=== Section 3: Website Assets ==="

run_test "Website public directory exists" "test -d $PROJECT_ROOT/Website/public"
run_test "Website styles directory exists" "test -d $PROJECT_ROOT/Website/styles"
run_test "Website scripts directory exists" "test -d $PROJECT_ROOT/Website/scripts"

echo ""
echo "=== Section 4: Documentation Website ==="

file_check "Docs architecture" "docs/website/ARCHITECTURE.md"

echo ""
echo "=== Section 5: Content Counts ==="

run_test "User manuals count >= 16" "test $(ls Website/user-manuals/*.md 2>/dev/null | wc -l) -ge 16"
run_test "Video courses count >= 16" "test $(ls Website/video-courses/*.md 2>/dev/null | wc -l) -ge 16"

echo ""
echo "=== Section 6: Website Build Verification ==="

run_test "Build script is executable" "test -x Website/build.sh"

echo ""
echo "=== Section 7: Content Quality ==="

run_test "User manual 01 has content" "test -s Website/user-manuals/01-getting-started.md"
run_test "Video course 01 has content" "test -s Website/video-courses/course-01-fundamentals.md"

echo ""
echo "=========================================="
echo "Phase 7 Website Challenge Summary"
echo "=========================================="
echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo "✓ All Phase 7 website tests PASSED"
    exit 0
else
    echo "✗ Some Phase 7 website tests FAILED"
    exit 1
fi
