#!/bin/bash

# Skills System Challenge
# Tests the comprehensive Skills system for HelixAgent
# Validates skill parsing, registry, matching, tracking, and protocol integration

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
RESULTS_DIR="$SCRIPT_DIR/../results/skills_challenge"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
PASSED=0
FAILED=0
TOTAL=0

# Functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; ((PASSED++)); ((TOTAL++)); }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; ((FAILED++)); ((TOTAL++)); }
log_section() { echo -e "\n${YELLOW}=== $1 ===${NC}"; }

check_file() {
    if [[ -f "$1" ]]; then
        log_pass "$2 exists"
        return 0
    else
        log_fail "$2 missing: $1"
        return 1
    fi
}

check_contains() {
    if grep -q "$2" "$1" 2>/dev/null; then
        log_pass "$3"
        return 0
    else
        log_fail "$3 - pattern not found: $2"
        return 1
    fi
}

check_go_struct() {
    if grep -q "type $2 struct" "$1" 2>/dev/null; then
        log_pass "$2 struct defined"
        return 0
    else
        log_fail "$2 struct not found in $1"
        return 1
    fi
}

check_go_func() {
    if grep -q "func.*$2" "$1" 2>/dev/null; then
        log_pass "$2 function defined"
        return 0
    else
        log_fail "$2 function not found in $1"
        return 1
    fi
}

# Create results directory
mkdir -p "$RESULTS_DIR"

echo -e "\n${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         HELIXAGENT SKILLS SYSTEM CHALLENGE                      ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}\n"

log_info "Project root: $PROJECT_ROOT"
log_info "Results dir: $RESULTS_DIR"

# ============================================================================
# SECTION 1: Core Files Existence
# ============================================================================
log_section "Section 1: Core Skills Package Files"

SKILLS_DIR="$PROJECT_ROOT/internal/skills"

check_file "$SKILLS_DIR/types.go" "types.go"
check_file "$SKILLS_DIR/parser.go" "parser.go"
check_file "$SKILLS_DIR/registry.go" "registry.go"
check_file "$SKILLS_DIR/matcher.go" "matcher.go"
check_file "$SKILLS_DIR/tracker.go" "tracker.go"
check_file "$SKILLS_DIR/service.go" "service.go"
check_file "$SKILLS_DIR/protocol_adapter.go" "protocol_adapter.go"

# ============================================================================
# SECTION 2: Core Types
# ============================================================================
log_section "Section 2: Core Types Definition"

check_go_struct "$SKILLS_DIR/types.go" "Skill"
check_go_struct "$SKILLS_DIR/types.go" "SkillCategory"
check_go_struct "$SKILLS_DIR/types.go" "SkillMatch"
check_go_struct "$SKILLS_DIR/types.go" "SkillUsage"
check_go_struct "$SKILLS_DIR/types.go" "SkillResponse"
check_go_struct "$SKILLS_DIR/types.go" "RegistryStats"
check_go_struct "$SKILLS_DIR/types.go" "SkillConfig"

check_contains "$SKILLS_DIR/types.go" "MatchType" "MatchType type defined"
check_contains "$SKILLS_DIR/types.go" "MatchTypeExact" "Exact match type"
check_contains "$SKILLS_DIR/types.go" "MatchTypePartial" "Partial match type"
check_contains "$SKILLS_DIR/types.go" "MatchTypeSemantic" "Semantic match type"
check_contains "$SKILLS_DIR/types.go" "MatchTypeFuzzy" "Fuzzy match type"

# ============================================================================
# SECTION 3: Parser Functionality
# ============================================================================
log_section "Section 3: SKILL.md Parser"

check_go_func "$SKILLS_DIR/parser.go" "NewParser"
check_go_func "$SKILLS_DIR/parser.go" "ParseFile"
check_go_func "$SKILLS_DIR/parser.go" "Parse"
check_go_func "$SKILLS_DIR/parser.go" "ParseDirectory"
check_go_func "$SKILLS_DIR/parser.go" "splitFrontmatter"
check_go_func "$SKILLS_DIR/parser.go" "extractTriggers"
check_go_func "$SKILLS_DIR/parser.go" "parseContentSections"
check_go_func "$SKILLS_DIR/parser.go" "extractCategory"

# ============================================================================
# SECTION 4: Registry Functionality
# ============================================================================
log_section "Section 4: Skill Registry"

check_go_func "$SKILLS_DIR/registry.go" "NewRegistry"
check_go_func "$SKILLS_DIR/registry.go" "Load"
check_go_func "$SKILLS_DIR/registry.go" "RegisterSkill"
check_go_func "$SKILLS_DIR/registry.go" "Get"
check_go_func "$SKILLS_DIR/registry.go" "GetByCategory"
check_go_func "$SKILLS_DIR/registry.go" "GetByTrigger"
check_go_func "$SKILLS_DIR/registry.go" "GetAll"
check_go_func "$SKILLS_DIR/registry.go" "Search"
check_go_func "$SKILLS_DIR/registry.go" "Remove"
check_go_func "$SKILLS_DIR/registry.go" "EnableHotReload"
check_go_func "$SKILLS_DIR/registry.go" "Stats"

# ============================================================================
# SECTION 5: Matcher Functionality
# ============================================================================
log_section "Section 5: Skill Matcher"

check_go_func "$SKILLS_DIR/matcher.go" "NewMatcher"
check_go_func "$SKILLS_DIR/matcher.go" "Match"
check_go_func "$SKILLS_DIR/matcher.go" "MatchBest"
check_go_func "$SKILLS_DIR/matcher.go" "MatchMultiple"
check_go_func "$SKILLS_DIR/matcher.go" "matchExact"
check_go_func "$SKILLS_DIR/matcher.go" "matchPartial"
check_go_func "$SKILLS_DIR/matcher.go" "matchFuzzy"
check_go_func "$SKILLS_DIR/matcher.go" "deduplicateAndSort"

check_contains "$SKILLS_DIR/matcher.go" "SemanticMatcher" "Semantic matcher interface"

# ============================================================================
# SECTION 6: Tracker Functionality
# ============================================================================
log_section "Section 6: Usage Tracker"

check_go_func "$SKILLS_DIR/tracker.go" "NewTracker"
check_go_func "$SKILLS_DIR/tracker.go" "StartTracking"
check_go_func "$SKILLS_DIR/tracker.go" "CompleteTracking"
check_go_func "$SKILLS_DIR/tracker.go" "RecordToolUse"
check_go_func "$SKILLS_DIR/tracker.go" "GetActiveUsage"
check_go_func "$SKILLS_DIR/tracker.go" "GetStats"
check_go_func "$SKILLS_DIR/tracker.go" "GetHistory"
check_go_func "$SKILLS_DIR/tracker.go" "GetTopSkills"

check_go_struct "$SKILLS_DIR/tracker.go" "UsageStats"
check_go_struct "$SKILLS_DIR/tracker.go" "SkillStats"
check_go_struct "$SKILLS_DIR/tracker.go" "CategoryStats"

# ============================================================================
# SECTION 7: Service Integration
# ============================================================================
log_section "Section 7: Service Integration"

check_go_func "$SKILLS_DIR/service.go" "NewService"
check_go_func "$SKILLS_DIR/service.go" "Initialize"
check_go_func "$SKILLS_DIR/service.go" "Shutdown"
check_go_func "$SKILLS_DIR/service.go" "FindSkills"
check_go_func "$SKILLS_DIR/service.go" "FindBestSkill"
check_go_func "$SKILLS_DIR/service.go" "StartSkillExecution"
check_go_func "$SKILLS_DIR/service.go" "CompleteSkillExecution"
check_go_func "$SKILLS_DIR/service.go" "GetUsageStats"
check_go_func "$SKILLS_DIR/service.go" "HealthCheck"
check_go_func "$SKILLS_DIR/service.go" "ExecuteWithTracking"
check_go_func "$SKILLS_DIR/service.go" "CreateResponse"

# ============================================================================
# SECTION 8: Protocol Adapter
# ============================================================================
log_section "Section 8: Protocol Adapter"

check_go_struct "$SKILLS_DIR/protocol_adapter.go" "ProtocolSkillAdapter"
check_go_struct "$SKILLS_DIR/protocol_adapter.go" "MCPSkillTool"
check_go_struct "$SKILLS_DIR/protocol_adapter.go" "ACPSkillAction"
check_go_struct "$SKILLS_DIR/protocol_adapter.go" "LSPSkillCommand"
check_go_struct "$SKILLS_DIR/protocol_adapter.go" "SkillToolCall"
check_go_struct "$SKILLS_DIR/protocol_adapter.go" "SkillToolResult"

check_go_func "$SKILLS_DIR/protocol_adapter.go" "NewProtocolSkillAdapter"
check_go_func "$SKILLS_DIR/protocol_adapter.go" "RegisterAllSkillsAsTools"
check_go_func "$SKILLS_DIR/protocol_adapter.go" "GetMCPTools"
check_go_func "$SKILLS_DIR/protocol_adapter.go" "GetACPActions"
check_go_func "$SKILLS_DIR/protocol_adapter.go" "GetLSPCommands"
check_go_func "$SKILLS_DIR/protocol_adapter.go" "InvokeMCPTool"
check_go_func "$SKILLS_DIR/protocol_adapter.go" "InvokeACPAction"
check_go_func "$SKILLS_DIR/protocol_adapter.go" "InvokeLSPCommand"
check_go_func "$SKILLS_DIR/protocol_adapter.go" "ToMCPToolList"
check_go_func "$SKILLS_DIR/protocol_adapter.go" "ToACPActionList"
check_go_func "$SKILLS_DIR/protocol_adapter.go" "ToLSPCommandList"
check_go_func "$SKILLS_DIR/protocol_adapter.go" "GetSkillUsageHeader"

check_contains "$SKILLS_DIR/protocol_adapter.go" "ProtocolMCP" "MCP protocol constant"
check_contains "$SKILLS_DIR/protocol_adapter.go" "ProtocolACP" "ACP protocol constant"
check_contains "$SKILLS_DIR/protocol_adapter.go" "ProtocolLSP" "LSP protocol constant"

# ============================================================================
# SECTION 9: Test Coverage
# ============================================================================
log_section "Section 9: Test Coverage"

check_file "$SKILLS_DIR/parser_test.go" "Parser tests"
check_file "$SKILLS_DIR/registry_test.go" "Registry tests"
check_file "$SKILLS_DIR/matcher_test.go" "Matcher tests"
check_file "$SKILLS_DIR/tracker_test.go" "Tracker tests"
check_file "$SKILLS_DIR/protocol_adapter_test.go" "Protocol adapter tests"

# ============================================================================
# SECTION 10: Run Go Tests
# ============================================================================
log_section "Section 10: Unit Tests Execution"

cd "$PROJECT_ROOT"

log_info "Running skills package tests..."
if go test -v ./internal/skills/... > "$RESULTS_DIR/test_output.log" 2>&1; then
    TEST_COUNT=$(grep -c "^--- PASS" "$RESULTS_DIR/test_output.log" 2>/dev/null || echo "0")
    log_pass "All $TEST_COUNT unit tests passed"
else
    FAIL_COUNT=$(grep -c "^--- FAIL" "$RESULTS_DIR/test_output.log" 2>/dev/null || echo "0")
    log_fail "$FAIL_COUNT tests failed - see $RESULTS_DIR/test_output.log"
fi

# ============================================================================
# SECTION 11: Skills Response Tracking
# ============================================================================
log_section "Section 11: Skills Response Tracking"

check_contains "$SKILLS_DIR/types.go" "SkillsUsed.*SkillUsage" "SkillResponse includes SkillsUsed"
check_contains "$SKILLS_DIR/protocol_adapter.go" "SkillsUsed.*usages" "Protocol result includes skills used"
check_contains "$SKILLS_DIR/protocol_adapter.go" "GetSkillUsageHeader" "HTTP header for skill usage"
check_contains "$SKILLS_DIR/tracker.go" "ToolsInvoked" "Tool invocation tracking"
check_contains "$SKILLS_DIR/types.go" "ProviderUsed" "Provider tracking in response"
check_contains "$SKILLS_DIR/types.go" "ModelUsed" "Model tracking in response"
check_contains "$SKILLS_DIR/types.go" "Protocol" "Protocol tracking in response"

# ============================================================================
# SECTION 12: Build Verification
# ============================================================================
log_section "Section 12: Build Verification"

log_info "Building skills package..."
if go build ./internal/skills/... > "$RESULTS_DIR/build_output.log" 2>&1; then
    log_pass "Skills package builds successfully"
else
    log_fail "Skills package build failed - see $RESULTS_DIR/build_output.log"
fi

log_info "Building full project..."
if go build ./... > "$RESULTS_DIR/full_build_output.log" 2>&1; then
    log_pass "Full project builds successfully"
else
    log_fail "Full project build failed - see $RESULTS_DIR/full_build_output.log"
fi

# ============================================================================
# Summary
# ============================================================================
echo -e "\n${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                    CHALLENGE RESULTS                            ║${NC}"
echo -e "${BLUE}╠════════════════════════════════════════════════════════════════╣${NC}"
echo -e "${BLUE}║${NC}  Total Tests:  ${TOTAL}"
echo -e "${BLUE}║${NC}  ${GREEN}Passed:       ${PASSED}${NC}"
echo -e "${BLUE}║${NC}  ${RED}Failed:       ${FAILED}${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"

# Save results
cat > "$RESULTS_DIR/summary.json" << EOF
{
  "challenge": "skills_system",
  "timestamp": "$(date -Iseconds)",
  "total": $TOTAL,
  "passed": $PASSED,
  "failed": $FAILED,
  "success_rate": $(awk "BEGIN {printf \"%.2f\", $PASSED/$TOTAL*100}")
}
EOF

if [[ $FAILED -eq 0 ]]; then
    echo -e "\n${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║          ALL SKILLS SYSTEM TESTS PASSED!                       ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "\n${RED}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║          SOME TESTS FAILED - SEE DETAILS ABOVE                 ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════════════════════╝${NC}"
    exit 1
fi
