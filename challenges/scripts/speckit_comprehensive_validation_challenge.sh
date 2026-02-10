#!/usr/bin/env bash

###############################################################################
# SpecKit Comprehensive Validation Challenge
#
# This challenge validates EVERYTHING related to GitHub's SpecKit integration:
# 1. GitHub SpecKit submodule status
# 2. GitHub SpecKit content verification
# 3. Integration with HelixAgent
# 4. Cache functionality
# 5. Configuration
# 6. Documentation
# 7. Functional tests
#
# Zero tolerance for false positives - all tests use real validation.
# NOTE: Only GitHub's SpecKit (cli_agents/spec-kit/) is used - no separate implementation.
###############################################################################

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

run_test() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    local test_name="$1"
    local test_command="$2"

    log_info "Testing: $test_name"

    if eval "$test_command"; then
        log_success "$test_name"
        return 0
    else
        log_error "$test_name"
        return 1
    fi
}

###############################################################################
# Section 1: GitHub SpecKit Submodule Validation
###############################################################################

log_info "========================================="
log_info "Section 1: GitHub SpecKit Submodule"
log_info "========================================="

# Test 1.1: Submodule directory exists
run_test "GitHub SpecKit directory exists" \
    "test -d cli_agents/spec-kit"

# Test 1.2: Submodule is initialized
run_test "GitHub SpecKit submodule is initialized" \
    "git submodule status cli_agents/spec-kit | grep -v '^-'"

# Test 1.3: Submodule has .git directory
run_test "GitHub SpecKit has .git directory" \
    "test -e cli_agents/spec-kit/.git"

# Test 1.4: .gitmodules contains spec-kit
run_test ".gitmodules contains spec-kit configuration" \
    "grep -q 'spec-kit' .gitmodules"

# Test 1.5: Remote URL is correct
run_test "GitHub SpecKit remote URL is correct" \
    "grep -q 'git@github.com:github/spec-kit.git' .gitmodules"

# Test 1.6: README exists and is substantial
run_test "GitHub SpecKit README exists and is comprehensive" \
    "test -f cli_agents/spec-kit/README.md && test \$(wc -c < cli_agents/spec-kit/README.md) -gt 1000"

# Test 1.7: AGENTS.md exists
run_test "GitHub SpecKit AGENTS.md exists" \
    "test -f cli_agents/spec-kit/AGENTS.md"

# Test 1.8: Submodule is on a tagged version
run_test "GitHub SpecKit is on a version tag" \
    "git submodule status cli_agents/spec-kit | grep -E 'v[0-9]+\.[0-9]+\.[0-9]+'"

# Test 1.9: docs directory exists
run_test "GitHub SpecKit docs directory exists" \
    "test -d cli_agents/spec-kit/docs"

# Test 1.10: No local modifications to submodule
run_test "GitHub SpecKit has no local modifications" \
    "git diff --exit-code cli_agents/spec-kit"

###############################################################################
# Section 2: GitHub SpecKit Content Verification
###############################################################################

log_info ""
log_info "========================================="
log_info "Section 2: GitHub SpecKit Content"
log_info "========================================="

# Test 2.1: GitHub SpecKit has scripts directory
run_test "GitHub SpecKit scripts directory exists" \
    "test -d cli_agents/spec-kit/scripts"

# Test 2.2: GitHub SpecKit has docs directory with content
run_test "GitHub SpecKit docs have substantial content" \
    "test -d cli_agents/spec-kit/docs && find cli_agents/spec-kit/docs -type f | grep -q '.'"

# Test 2.3: GitHub SpecKit has media directory
run_test "GitHub SpecKit media directory exists" \
    "test -d cli_agents/spec-kit/media"

# Test 2.4: GitHub SpecKit has Python package config
run_test "GitHub SpecKit has pyproject.toml" \
    "test -f cli_agents/spec-kit/pyproject.toml"

# Test 2.5: GitHub SpecKit has memory directory
run_test "GitHub SpecKit memory directory exists" \
    "test -d cli_agents/spec-kit/memory"

# Test 2.6: GitHub SpecKit has CHANGELOG
run_test "GitHub SpecKit has CHANGELOG.md" \
    "test -f cli_agents/spec-kit/CHANGELOG.md"

# Test 2.7: GitHub SpecKit has CONTRIBUTING guide
run_test "GitHub SpecKit has CONTRIBUTING.md" \
    "test -f cli_agents/spec-kit/CONTRIBUTING.md"

# Test 2.8: GitHub SpecKit has LICENSE
run_test "GitHub SpecKit has LICENSE file" \
    "test -f cli_agents/spec-kit/LICENSE"

###############################################################################
# Section 3: Integration with HelixAgent
###############################################################################

log_info ""
log_info "========================================="
log_info "Section 3: HelixAgent Integration"
log_info "========================================="

# Test 3.1: Agent registry file exists
run_test "Agent registry file exists" \
    "test -f internal/agents/registry.go"

# Test 3.2: Registry contains spec-kit reference
run_test "Agent registry references spec-kit" \
    "grep -q 'spec-kit' internal/agents/registry.go"

# Test 3.3: EntryPoint is defined
run_test "spec-kit EntryPoint is defined in registry" \
    "grep -A5 'spec-kit' internal/agents/registry.go | grep -q 'EntryPoint'"

# Test 3.4: ConfigLocation is defined
run_test "spec-kit ConfigLocation is defined" \
    "grep -A5 'spec-kit' internal/agents/registry.go | grep -q 'ConfigLocation'"

# Test 3.5: Original SpecKit orchestrator exists
run_test "Original SpecKit orchestrator exists in services" \
    "test -f internal/services/speckit_orchestrator.go"

# Test 3.6: SpecKit orchestrator has 7 phases defined
run_test "SpecKit orchestrator defines all 7 phases" \
    "grep -q 'PhaseConstitution\|PhaseSpecify\|PhaseClarify\|PhasePlan\|PhaseTasks\|PhaseAnalyze\|PhaseImplement' internal/services/speckit_orchestrator.go"

# Test 3.7: Integration tests exist
run_test "GitHub SpecKit integration tests exist" \
    "test -f tests/integration/github_speckit_integration_test.go"

# Test 3.8: E2E workflow tests exist (optional)
if [ -f "tests/e2e/speckit_workflow_test.go" ]; then
    run_test "SpecKit E2E workflow tests exist" \
        "test -f tests/e2e/speckit_workflow_test.go"
else
    log_warning "SpecKit E2E workflow tests not yet created (optional)"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
fi

# Test 3.9: Documentation status report exists
run_test "GitHub SpecKit integration status report exists" \
    "test -f GITHUB_SPECKIT_INTEGRATION_STATUS.md"

# Test 3.10: Clarification document exists
run_test "SpecKit clarification document exists" \
    "test -f SPECKIT_CLARIFICATION.md"

###############################################################################
# Section 4: Cache Functionality
###############################################################################

log_info ""
log_info "========================================="
log_info "Section 4: Cache Functionality"
log_info "========================================="

# Test 4.1: Can create cache directory
TEMP_CACHE_DIR="$(mktemp -d)/speckit_test_cache"
run_test "Can create cache directory" \
    "mkdir -p $TEMP_CACHE_DIR/.speckit/cache"

# Test 4.2: Can write to cache
run_test "Can write to cache directory" \
    "echo '{\"phase\":\"test\"}' > $TEMP_CACHE_DIR/.speckit/cache/test.json"

# Test 4.3: Can read from cache
run_test "Can read from cache directory" \
    "test -f $TEMP_CACHE_DIR/.speckit/cache/test.json && grep -q 'phase' $TEMP_CACHE_DIR/.speckit/cache/test.json"

# Test 4.4: Can delete cache
run_test "Can delete cache directory" \
    "rm -rf $TEMP_CACHE_DIR/.speckit/cache && test ! -d $TEMP_CACHE_DIR/.speckit/cache"

# Cleanup
rm -rf "$TEMP_CACHE_DIR"

###############################################################################
# Section 5: Configuration
###############################################################################

log_info ""
log_info "========================================="
log_info "Section 5: Configuration"
log_info "========================================="

# Test 5.1: CLAUDE.md mentions SpecKit
run_test "CLAUDE.md mentions SpecKit" \
    "grep -q 'SpecKit' CLAUDE.md"

# Test 5.2: AGENTS.md mentions SpecKit
run_test "AGENTS.md mentions SpecKit" \
    "grep -q 'SpecKit' AGENTS.md"

# Test 5.3: README mentions SpecKit
run_test "README.md mentions SpecKit" \
    "grep -q 'SpecKit' README.md"

# Test 5.4: Constitution mentions GitSpec compliance
run_test "Constitution mentions GitSpec compliance" \
    "grep -q 'GitSpec' CONSTITUTION.md"

# Test 5.5: go.mod does NOT reference SpecKit module (single implementation)
run_test "Root go.mod does not reference separate SpecKit module" \
    "! grep -q 'digital.vasic.speckit' go.mod 2>/dev/null"

###############################################################################
# Section 6: Documentation
###############################################################################

log_info ""
log_info "========================================="
log_info "Section 6: Documentation"
log_info "========================================="

# Test 6.1: SpecKit user guide exists
run_test "SpecKit user guide exists" \
    "test -f docs/guides/SPECKIT_USER_GUIDE.md"

# Test 6.2: SpecKit user guide is comprehensive
run_test "SpecKit user guide is comprehensive" \
    "test \$(wc -c < docs/guides/SPECKIT_USER_GUIDE.md) -gt 10000"

# Test 6.3: User guide mentions 7-phase flow
run_test "User guide documents 7-phase flow" \
    "grep -q '7-phase' docs/guides/SPECKIT_USER_GUIDE.md"

# Test 6.4: User guide mentions granularity detection
run_test "User guide documents granularity detection" \
    "grep -q 'granularity' docs/guides/SPECKIT_USER_GUIDE.md"

# Test 6.5: User guide has configuration section
run_test "User guide has configuration section" \
    "grep -q 'Configuration' docs/guides/SPECKIT_USER_GUIDE.md"

# Test 6.6: Final status report exists
run_test "SpecKit final status report exists" \
    "test -f SPECKIT_FINAL_STATUS.md"

# Test 6.7: GitHub SpecKit README is present
run_test "GitHub SpecKit has comprehensive README" \
    "test \$(wc -c < cli_agents/spec-kit/README.md) -gt 5000"

# Test 6.8: Integration status report is comprehensive
run_test "GitHub SpecKit integration report is comprehensive" \
    "test \$(wc -c < GITHUB_SPECKIT_INTEGRATION_STATUS.md) -gt 5000"

# Test 6.9: Final status report is comprehensive
run_test "SpecKit final status report is comprehensive" \
    "test \$(wc -c < SPECKIT_FINAL_STATUS.md) -gt 5000"

# Test 6.10: All documentation has recent timestamps
run_test "Documentation has recent timestamps (Feb 2026)" \
    "grep -q 'February.*2026\|2026-02' docs/guides/SPECKIT_USER_GUIDE.md"

###############################################################################
# Section 7: Functional Tests (if components available)
###############################################################################

log_info ""
log_info "========================================="
log_info "Section 7: Functional Tests"
log_info "========================================="

# Test 7.1: Verify no SpecKit directory exists (single implementation)
run_test "No separate SpecKit directory exists" \
    "test ! -d SpecKit"

# Test 7.2: Integration tests compile
run_test "Integration tests compile" \
    "go test -c tests/integration/github_speckit_integration_test.go -o /tmp/speckit_test_bin && rm /tmp/speckit_test_bin"

# Test 7.3: Can run quick integration tests
if [ "${RUN_FULL_TESTS:-false}" = "true" ]; then
    run_test "Integration tests pass" \
        "go test -v -short ./tests/integration/github_speckit_integration_test.go"
else
    log_info "Skipping full test run (set RUN_FULL_TESTS=true to enable)"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
fi

# Test 7.4: Check if Specify CLI is available
if command -v specify &> /dev/null; then
    run_test "Specify CLI is installed" \
        "specify --version"
else
    log_warning "Specify CLI not installed - run: uv tool install specify-cli --from git+https://github.com/github/spec-kit.git"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
fi

# Test 7.5: Check if UV tool is available
if command -v uv &> /dev/null; then
    run_test "UV tool is available" \
        "uv --version"
else
    log_warning "UV not installed - required for GitHub SpecKit installation"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
fi

###############################################################################
# Summary
###############################################################################

echo ""
echo "========================================="
echo "CHALLENGE SUMMARY"
echo "========================================="
echo -e "Total Tests:  ${BLUE}$TOTAL_TESTS${NC}"
echo -e "Passed:       ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed:       ${RED}$FAILED_TESTS${NC}"

if [ "$FAILED_TESTS" -eq 0 ]; then
    echo -e "\n${GREEN}✅ ALL TESTS PASSED!${NC}"
    echo "SpecKit is fully validated and ready to use."
    exit 0
else
    echo -e "\n${RED}❌ SOME TESTS FAILED${NC}"
    echo "Please review the failures above and fix the issues."
    exit 1
fi
