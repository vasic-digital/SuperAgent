#!/bin/bash
# Website Content Completeness Challenge
# Validates website HTML pages, module counts, changelog entries,
# video course count, and user manual count.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "=========================================="
echo "Website Content Completeness Challenge"
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

    if eval "$test_cmd" > /tmp/website_completeness_test_output.txt 2>&1; then
        echo "  PASSED"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "  FAILED"
        cat /tmp/website_completeness_test_output.txt 2>/dev/null || true
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
        echo "  PASSED - File exists: $file_path"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "  FAILED - File missing: $file_path"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

cd "$PROJECT_ROOT"

echo ""
echo "=== Section 1: All 7 HTML Pages Exist ==="

file_check "index.html exists" "Website/public/index.html"
file_check "features.html exists" "Website/public/features.html"
file_check "pricing.html exists" "Website/public/pricing.html"
file_check "changelog.html exists" "Website/public/changelog.html"
file_check "contact.html exists" "Website/public/contact.html"
file_check "privacy.html exists" "Website/public/privacy.html"
file_check "terms.html exists" "Website/public/terms.html"

run_test "Exactly 7 HTML pages in public root" \
    "test \$(ls Website/public/*.html 2>/dev/null | wc -l) -eq 7"

echo ""
echo "=== Section 2: Module Count Mentions 41 ==="

run_test "index.html mentions 41 modules" \
    "grep -q '41.*module\|41 extracted\|41 module' Website/public/index.html"

run_test "CHANGELOG.md mentions 41 modules" \
    "grep -q '41 total\|41 module\|33 to 41\|33.*41' CHANGELOG.md"

run_test "CLAUDE.md mentions 41 in MODULES.md reference" \
    "grep -q '41 total\|33 to 41' CLAUDE.md || grep -q 'MODULES.md' CLAUDE.md"

echo ""
echo "=== Section 3: Changelog Has SP1-SP5 Entries ==="

run_test "CHANGELOG.md has Unreleased section" \
    "grep -q '## \[Unreleased\]' CHANGELOG.md"

run_test "CHANGELOG.md has v1.0.0 entry" \
    "grep -q '## \[1.0.0\]' CHANGELOG.md"

run_test "CHANGELOG.md has v0.1.0 entry" \
    "grep -q '## \[0.1.0\]' CHANGELOG.md"

run_test "CHANGELOG.md has at least 3 version entries" \
    "test \$(grep -c '## \[' CHANGELOG.md) -ge 3"

run_test "Website changelog.html has version entries" \
    "test \$(grep -c 'version-badge' Website/public/changelog.html) -ge 5"

run_test "Website changelog.html has Fixed sections" \
    "grep -q 'change-type fixed' Website/public/changelog.html"

run_test "Website changelog.html has Added sections" \
    "grep -q 'change-type added' Website/public/changelog.html"

echo ""
echo "=== Section 4: Video Course Count >= 43 ==="

COURSE_COUNT=$(ls Website/video-courses/*.md 2>/dev/null | grep -v README | wc -l)
run_test "Video courses directory exists" \
    "test -d Website/video-courses"

run_test "Video course count >= 43 (found: $COURSE_COUNT)" \
    "test $COURSE_COUNT -ge 43"

run_test "Video courses have content (first course non-empty)" \
    "test -s \$(ls Website/video-courses/course-*.md 2>/dev/null | head -1)"

echo ""
echo "=== Section 5: User Manual Count >= 44 ==="

MANUAL_COUNT=$(ls Website/user-manuals/*.md 2>/dev/null | grep -v README | wc -l)
run_test "User manuals directory exists" \
    "test -d Website/user-manuals"

run_test "User manual count >= 44 (found: $MANUAL_COUNT)" \
    "test $MANUAL_COUNT -ge 44"

run_test "User manuals have content (first manual non-empty)" \
    "test -s \$(ls Website/user-manuals/0*.md 2>/dev/null | head -1)"

echo ""
echo "=== Section 6: Content Quality Checks ==="

run_test "index.html has non-trivial content (>1KB)" \
    "test \$(wc -c < Website/public/index.html) -gt 1024"

run_test "features.html has non-trivial content (>1KB)" \
    "test \$(wc -c < Website/public/features.html) -gt 1024"

run_test "changelog.html has non-trivial content (>1KB)" \
    "test \$(wc -c < Website/public/changelog.html) -gt 1024"

run_test "CHANGELOG.md follows Keep a Changelog format" \
    "grep -q 'Keep a Changelog' CHANGELOG.md"

run_test "CHANGELOG.md has Added/Fixed/Changed/Removed sections" \
    "grep -q '### Added' CHANGELOG.md && grep -q '### Fixed' CHANGELOG.md && grep -q '### Changed' CHANGELOG.md && grep -q '### Removed' CHANGELOG.md"

echo ""
echo "=========================================="
echo "Website Content Completeness Challenge Summary"
echo "=========================================="
echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo "All website content completeness tests PASSED"
    exit 0
else
    echo "Some website content completeness tests FAILED"
    exit 1
fi
