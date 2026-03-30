#!/bin/bash
# HelixAgent Challenge - Documentation Completeness Validation
# Validates Phase 7 documentation: QA API reference, Website module docs,
# user manuals 39/41/44, video courses 71/73, CLAUDE.md QA endpoints,
# VisionEngine remote test files, and HelixQA adapter test files.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

if [ -f "$SCRIPT_DIR/challenge_framework.sh" ]; then
    source "$SCRIPT_DIR/challenge_framework.sh"
    init_challenge "documentation-completeness-validation" "Documentation Completeness Validation"
    load_env
    FRAMEWORK_LOADED=true
else
    FRAMEWORK_LOADED=false
fi

PASSED=0
FAILED=0

record_result() {
    local name="$1" status="$2"
    if [ "$FRAMEWORK_LOADED" = true ]; then
        if [ "$status" = "PASS" ]; then
            record_assertion "test" "$name" "true" "$name"
        else
            record_assertion "test" "$name" "false" "$name"
        fi
    fi
    if [ "$status" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        echo -e "\033[0;32m[PASS]\033[0m $name"
    else
        FAILED=$((FAILED + 1))
        echo -e "\033[0;31m[FAIL]\033[0m $name"
    fi
}

echo "=== Documentation Completeness Validation Challenge ==="
echo ""

# --- QA API Reference ---
echo "--- QA API Reference ---"

# Test 1: QA_API_REFERENCE.md exists
if [ -f "$PROJECT_ROOT/docs/api/QA_API_REFERENCE.md" ]; then
    record_result "QA_API_REFERENCE.md exists" "PASS"
else
    record_result "QA_API_REFERENCE.md exists" "FAIL"
fi

# Test 2: QA_API_REFERENCE.md is non-empty (has meaningful content)
if [ -f "$PROJECT_ROOT/docs/api/QA_API_REFERENCE.md" ] && \
   [ "$(wc -l < "$PROJECT_ROOT/docs/api/QA_API_REFERENCE.md")" -gt 10 ]; then
    record_result "QA_API_REFERENCE.md is non-empty (>10 lines)" "PASS"
else
    record_result "QA_API_REFERENCE.md is non-empty (>10 lines)" "FAIL"
fi

echo ""
echo "--- Website Module Documentation ---"

# Test 3: Website/README.md exists
if [ -f "$PROJECT_ROOT/Website/README.md" ]; then
    record_result "Website/README.md exists" "PASS"
else
    record_result "Website/README.md exists" "FAIL"
fi

# Test 4: Website/CLAUDE.md exists
if [ -f "$PROJECT_ROOT/Website/CLAUDE.md" ]; then
    record_result "Website/CLAUDE.md exists" "PASS"
else
    record_result "Website/CLAUDE.md exists" "FAIL"
fi

echo ""
echo "--- User Manuals (HelixQA and VisionEngine) ---"

# Test 5: User manual 39 (helixqa) mentions "autonomous" keyword
if [ -f "$PROJECT_ROOT/Website/user-manuals/39-helixqa-guide.md" ] && \
   grep -qi "autonomous" "$PROJECT_ROOT/Website/user-manuals/39-helixqa-guide.md"; then
    record_result "User manual 39 (helixqa) mentions autonomous sessions" "PASS"
else
    record_result "User manual 39 (helixqa) mentions autonomous sessions" "FAIL"
fi

# Test 6: User manual 41 (visionengine) mentions "VisionPool" keyword
if [ -f "$PROJECT_ROOT/Website/user-manuals/41-visionengine-guide.md" ] && \
   grep -q "VisionPool" "$PROJECT_ROOT/Website/user-manuals/41-visionengine-guide.md"; then
    record_result "User manual 41 (visionengine) mentions VisionPool" "PASS"
else
    record_result "User manual 41 (visionengine) mentions VisionPool" "FAIL"
fi

# Test 7: User manual 44 (qa-api-guide) exists
if [ -f "$PROJECT_ROOT/Website/user-manuals/44-qa-api-guide.md" ]; then
    record_result "User manual 44 (qa-api-guide) exists" "PASS"
else
    record_result "User manual 44 (qa-api-guide) exists" "FAIL"
fi

echo ""
echo "--- Video Courses (HelixQA and VisionEngine) ---"

# Test 8: Video course 71 (helixqa) exists and is non-empty
if [ -f "$PROJECT_ROOT/Website/video-courses/course-71-helixqa.md" ] && \
   [ "$(wc -l < "$PROJECT_ROOT/Website/video-courses/course-71-helixqa.md")" -gt 5 ]; then
    record_result "Video course 71 (helixqa) exists and is non-empty" "PASS"
else
    record_result "Video course 71 (helixqa) exists and is non-empty" "FAIL"
fi

# Test 9: Video course 73 (visionengine) exists and mentions VisionEngine
if [ -f "$PROJECT_ROOT/Website/video-courses/course-73-visionengine.md" ] && \
   grep -q "VisionEngine" "$PROJECT_ROOT/Website/video-courses/course-73-visionengine.md"; then
    record_result "Video course 73 (visionengine) exists and mentions VisionEngine" "PASS"
else
    record_result "Video course 73 (visionengine) exists and mentions VisionEngine" "FAIL"
fi

echo ""
echo "--- CLAUDE.md QA Endpoint Coverage ---"

# Test 10: CLAUDE.md mentions /v1/qa/ endpoints
if grep -q "/v1/qa/" "$PROJECT_ROOT/CLAUDE.md"; then
    record_result "CLAUDE.md documents /v1/qa/ endpoints" "PASS"
else
    record_result "CLAUDE.md documents /v1/qa/ endpoints" "FAIL"
fi

echo ""
echo "--- Test File Coverage ---"

# Test 11: VisionEngine pkg/remote/ directory has test files
REMOTE_TEST_COUNT=$(ls "$PROJECT_ROOT/VisionEngine/pkg/remote/"*_test.go 2>/dev/null | wc -l)
if [ "$REMOTE_TEST_COUNT" -gt 0 ]; then
    record_result "VisionEngine pkg/remote/ has test files ($REMOTE_TEST_COUNT found)" "PASS"
else
    record_result "VisionEngine pkg/remote/ has test files" "FAIL"
fi

# Test 12: HelixQA adapter has test files
if [ -f "$PROJECT_ROOT/internal/adapters/helixqa/adapter_test.go" ]; then
    record_result "HelixQA adapter has test file (adapter_test.go)" "PASS"
else
    record_result "HelixQA adapter has test file (adapter_test.go)" "FAIL"
fi

echo ""
echo "=== Results ==="
TOTAL=$((PASSED + FAILED))
echo "Passed: $PASSED/$TOTAL"
echo "Failed: $FAILED/$TOTAL"

if [ "$FRAMEWORK_LOADED" = true ]; then
    finalize_challenge "$PASSED" "$TOTAL"
fi

if [ "$FAILED" -gt 0 ]; then
    exit 1
fi
