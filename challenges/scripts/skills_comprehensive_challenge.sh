#!/bin/bash
# Skills Comprehensive Challenge
# Validates ALL Skills components: Types, Registry, Service, Matcher, Loader, Tracker, Parser, Protocol Adapter, Integration
# Tests: Implementation, Tests, Interface compliance, Protocol adapters

set -e

# Source challenge framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/challenge_framework.sh"

# Initialize challenge
init_challenge "skills_comprehensive" "Skills Comprehensive Verification"
load_env

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    TESTS_TOTAL=$((TESTS_TOTAL + 1))

    if eval "$test_cmd" >> "$LOG_FILE" 2>&1; then
        log_success "PASS: $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        record_assertion "test" "$test_name" "true" ""
        return 0
    else
        log_error "FAIL: $test_name"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        record_assertion "test" "$test_name" "false" "Test command failed"
        return 1
    fi
}

# ============================================================================
# SECTION 1: SKILLS CORE FILES
# ============================================================================
log_info "=============================================="
log_info "SECTION 1: Skills Core Files"
log_info "=============================================="

# Skills component files
SKILLS_COMPONENTS=(
    "types"
    "registry"
    "service"
    "matcher"
    "loader"
    "tracker"
    "parser"
    "protocol_adapter"
    "integration"
)

log_info "Verifying Skills component implementations..."
for component in "${SKILLS_COMPONENTS[@]}"; do
    run_test "Skills: $component implementation exists" \
        "[[ -f '$PROJECT_ROOT/internal/skills/${component}.go' ]]"
done

# Check for tests
for component in "${SKILLS_COMPONENTS[@]}"; do
    run_test "Skills: $component has tests" \
        "[[ -f '$PROJECT_ROOT/internal/skills/${component}_test.go' ]]"
done

# ============================================================================
# SECTION 2: SKILL TYPES
# ============================================================================
log_info "=============================================="
log_info "SECTION 2: Skill Types"
log_info "=============================================="

run_test "Skill struct exists" \
    "grep -q 'type Skill struct' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "SkillExample struct exists" \
    "grep -q 'type SkillExample struct' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "SkillError struct exists" \
    "grep -q 'type SkillError struct' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "SkillCategory struct exists" \
    "grep -q 'type SkillCategory struct' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "SkillMatch struct exists" \
    "grep -q 'type SkillMatch struct' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "MatchType enum exists" \
    "grep -q 'type MatchType string' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "MatchType: exact" \
    "grep -q 'MatchTypeExact' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "MatchType: partial" \
    "grep -q 'MatchTypePartial' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "MatchType: semantic" \
    "grep -q 'MatchTypeSemantic' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "MatchType: fuzzy" \
    "grep -q 'MatchTypeFuzzy' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "SkillUsage struct exists" \
    "grep -q 'type SkillUsage struct' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "SkillResponse struct exists" \
    "grep -q 'type SkillResponse struct' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "RegistryStats struct exists" \
    "grep -q 'type RegistryStats struct' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "SkillConfig struct exists" \
    "grep -q 'type SkillConfig struct' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "AllowedTool struct exists" \
    "grep -q 'type AllowedTool struct' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "ParseAllowedTools function exists" \
    "grep -q 'func ParseAllowedTools' '$PROJECT_ROOT/internal/skills/types.go'"

run_test "DefaultSkillConfig function exists" \
    "grep -q 'func DefaultSkillConfig' '$PROJECT_ROOT/internal/skills/types.go'"

# ============================================================================
# SECTION 3: SKILL REGISTRY
# ============================================================================
log_info "=============================================="
log_info "SECTION 3: Skill Registry"
log_info "=============================================="

run_test "Registry struct exists" \
    "grep -q 'type Registry struct' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "NewRegistry function exists" \
    "grep -q 'func NewRegistry' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: Load method" \
    "grep -q 'func (r \*Registry) Load' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: LoadFromPath method" \
    "grep -q 'func (r \*Registry) LoadFromPath' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: RegisterSkill method" \
    "grep -q 'func (r \*Registry) RegisterSkill' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: Get method" \
    "grep -q 'func (r \*Registry) Get' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: GetByCategory method" \
    "grep -q 'func (r \*Registry) GetByCategory' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: GetByTrigger method" \
    "grep -q 'func (r \*Registry) GetByTrigger' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: GetAll method" \
    "grep -q 'func (r \*Registry) GetAll' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: GetCategories method" \
    "grep -q 'func (r \*Registry) GetCategories' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: GetTriggers method" \
    "grep -q 'func (r \*Registry) GetTriggers' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: Remove method" \
    "grep -q 'func (r \*Registry) Remove' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: Stats method" \
    "grep -q 'func (r \*Registry) Stats' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: EnableHotReload method" \
    "grep -q 'func (r \*Registry) EnableHotReload' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: DisableHotReload method" \
    "grep -q 'func (r \*Registry) DisableHotReload' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "Registry: Search method" \
    "grep -q 'func (r \*Registry) Search' '$PROJECT_ROOT/internal/skills/registry.go'"

run_test "DirectoryWatcher struct exists" \
    "grep -q 'type DirectoryWatcher struct' '$PROJECT_ROOT/internal/skills/registry.go'"

# ============================================================================
# SECTION 4: SKILL SERVICE
# ============================================================================
log_info "=============================================="
log_info "SECTION 4: Skill Service"
log_info "=============================================="

run_test "Service struct exists" \
    "grep -q 'type Service struct' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "NewService function exists" \
    "grep -q 'func NewService' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: Initialize method" \
    "grep -q 'func (s \*Service) Initialize' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: Start method" \
    "grep -q 'func (s \*Service) Start' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: Shutdown method" \
    "grep -q 'func (s \*Service) Shutdown' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: FindSkills method" \
    "grep -q 'func (s \*Service) FindSkills' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: FindBestSkill method" \
    "grep -q 'func (s \*Service) FindBestSkill' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: GetSkill method" \
    "grep -q 'func (s \*Service) GetSkill' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: GetSkillsByCategory method" \
    "grep -q 'func (s \*Service) GetSkillsByCategory' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: GetAllSkills method" \
    "grep -q 'func (s \*Service) GetAllSkills' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: GetCategories method" \
    "grep -q 'func (s \*Service) GetCategories' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: SearchSkills method" \
    "grep -q 'func (s \*Service) SearchSkills' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: StartSkillExecution method" \
    "grep -q 'func (s \*Service) StartSkillExecution' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: RecordToolUse method" \
    "grep -q 'func (s \*Service) RecordToolUse' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: CompleteSkillExecution method" \
    "grep -q 'func (s \*Service) CompleteSkillExecution' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: GetActiveExecutions method" \
    "grep -q 'func (s \*Service) GetActiveExecutions' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: GetUsageStats method" \
    "grep -q 'func (s \*Service) GetUsageStats' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: GetSkillStats method" \
    "grep -q 'func (s \*Service) GetSkillStats' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: GetTopSkills method" \
    "grep -q 'func (s \*Service) GetTopSkills' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: GetUsageHistory method" \
    "grep -q 'func (s \*Service) GetUsageHistory' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: GetRegistryStats method" \
    "grep -q 'func (s \*Service) GetRegistryStats' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: RegisterSkill method" \
    "grep -q 'func (s \*Service) RegisterSkill' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: RemoveSkill method" \
    "grep -q 'func (s \*Service) RemoveSkill' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: SetSemanticMatcher method" \
    "grep -q 'func (s \*Service) SetSemanticMatcher' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: HealthCheck method" \
    "grep -q 'func (s \*Service) HealthCheck' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: ExecuteWithTracking method" \
    "grep -q 'func (s \*Service) ExecuteWithTracking' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "Service: IsRunning method" \
    "grep -q 'func (s \*Service) IsRunning' '$PROJECT_ROOT/internal/skills/service.go'"

run_test "SkillExecutionContext struct exists" \
    "grep -q 'type SkillExecutionContext struct' '$PROJECT_ROOT/internal/skills/service.go'"

# ============================================================================
# SECTION 5: SKILL MATCHER
# ============================================================================
log_info "=============================================="
log_info "SECTION 5: Skill Matcher"
log_info "=============================================="

run_test "Matcher struct exists" \
    "grep -q 'type Matcher struct' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "SemanticMatcher interface exists" \
    "grep -q 'type SemanticMatcher interface' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "NewMatcher function exists" \
    "grep -q 'func NewMatcher' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Matcher: Match method" \
    "grep -q 'func (m \*Matcher) Match' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Matcher: matchExact method" \
    "grep -q 'matchExact' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Matcher: matchPartial method" \
    "grep -q 'matchPartial' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Matcher: matchFuzzy method" \
    "grep -q 'matchFuzzy' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Matcher: matchSemantic method" \
    "grep -q 'matchSemantic' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Matcher: deduplicateAndSort method" \
    "grep -q 'deduplicateAndSort' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Matcher: MatchBest method" \
    "grep -q 'func (m \*Matcher) MatchBest' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Matcher: MatchMultiple method" \
    "grep -q 'func (m \*Matcher) MatchMultiple' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Matcher: SetSemanticMatcher method" \
    "grep -q 'func (m \*Matcher) SetSemanticMatcher' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Utility: normalizeQuery" \
    "grep -q 'func normalizeQuery' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Utility: tokenize" \
    "grep -q 'func tokenize' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Utility: wordOverlap" \
    "grep -q 'func wordOverlap' '$PROJECT_ROOT/internal/skills/matcher.go'"

run_test "Utility: similarity (Jaccard)" \
    "grep -q 'func similarity' '$PROJECT_ROOT/internal/skills/matcher.go'"

# ============================================================================
# SECTION 6: SKILL LOADER
# ============================================================================
log_info "=============================================="
log_info "SECTION 6: Skill Loader"
log_info "=============================================="

run_test "SkillLoader struct exists" \
    "grep -q 'type SkillLoader struct' '$PROJECT_ROOT/internal/skills/loader.go'"

run_test "LoaderConfig struct exists" \
    "grep -q 'type LoaderConfig struct' '$PROJECT_ROOT/internal/skills/loader.go'"

run_test "NewSkillLoader function exists" \
    "grep -q 'func NewSkillLoader' '$PROJECT_ROOT/internal/skills/loader.go'"

run_test "Loader: LoadFromDirectory method" \
    "grep -q 'func (l \*SkillLoader) LoadFromDirectory' '$PROJECT_ROOT/internal/skills/loader.go'"

run_test "Loader: LoadFromConfig method" \
    "grep -q 'func (l \*SkillLoader) LoadFromConfig' '$PROJECT_ROOT/internal/skills/loader.go'"

run_test "Loader: LoadBuiltinSkills method" \
    "grep -q 'func (l \*SkillLoader) LoadBuiltinSkills' '$PROJECT_ROOT/internal/skills/loader.go'"

run_test "Loader: GetLoaded method" \
    "grep -q 'func (l \*SkillLoader) GetLoaded' '$PROJECT_ROOT/internal/skills/loader.go'"

run_test "Loader: GetLoadedCount method" \
    "grep -q 'func (l \*SkillLoader) GetLoadedCount' '$PROJECT_ROOT/internal/skills/loader.go'"

run_test "Loader: GetLoadedByCategory method" \
    "grep -q 'func (l \*SkillLoader) GetLoadedByCategory' '$PROJECT_ROOT/internal/skills/loader.go'"

run_test "Loader: ReloadSkill method" \
    "grep -q 'func (l \*SkillLoader) ReloadSkill' '$PROJECT_ROOT/internal/skills/loader.go'"

run_test "Loader: GetInventory method" \
    "grep -q 'func (l \*SkillLoader) GetInventory' '$PROJECT_ROOT/internal/skills/loader.go'"

run_test "SkillInventory struct exists" \
    "grep -q 'type SkillInventory struct' '$PROJECT_ROOT/internal/skills/loader.go'"

run_test "SkillInfo struct exists" \
    "grep -q 'type SkillInfo struct' '$PROJECT_ROOT/internal/skills/loader.go'"

# ============================================================================
# SECTION 7: SKILL TRACKER
# ============================================================================
log_info "=============================================="
log_info "SECTION 7: Skill Tracker"
log_info "=============================================="

run_test "Tracker struct exists" \
    "grep -q 'type Tracker struct' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "UsageStats struct exists" \
    "grep -q 'type UsageStats struct' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "SkillStats struct exists" \
    "grep -q 'type SkillStats struct' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "CategoryStats struct exists" \
    "grep -q 'type CategoryStats struct' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "NewTracker function exists" \
    "grep -q 'func NewTracker' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "Tracker: StartTracking method" \
    "grep -q 'func (t \*Tracker) StartTracking' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "Tracker: RecordToolUse method" \
    "grep -q 'func (t \*Tracker) RecordToolUse' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "Tracker: CompleteTracking method" \
    "grep -q 'func (t \*Tracker) CompleteTracking' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "Tracker: GetActiveUsage method" \
    "grep -q 'func (t \*Tracker) GetActiveUsage' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "Tracker: GetActiveUsages method" \
    "grep -q 'func (t \*Tracker) GetActiveUsages' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "Tracker: GetHistory method" \
    "grep -q 'func (t \*Tracker) GetHistory' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "Tracker: GetStats method" \
    "grep -q 'func (t \*Tracker) GetStats' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "Tracker: GetSkillStats method" \
    "grep -q 'func (t \*Tracker) GetSkillStats' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "Tracker: GetTopSkills method" \
    "grep -q 'func (t \*Tracker) GetTopSkills' '$PROJECT_ROOT/internal/skills/tracker.go'"

run_test "Tracker: Reset method" \
    "grep -q 'func (t \*Tracker) Reset' '$PROJECT_ROOT/internal/skills/tracker.go'"

# ============================================================================
# SECTION 8: SKILL PARSER
# ============================================================================
log_info "=============================================="
log_info "SECTION 8: Skill Parser"
log_info "=============================================="

run_test "Parser struct exists" \
    "grep -q 'type Parser struct' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "NewParser function exists" \
    "grep -q 'func NewParser' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: ParseFile method" \
    "grep -q 'func (p \*Parser) ParseFile' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: Parse method" \
    "grep -q 'func (p \*Parser) Parse' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: splitFrontmatter method" \
    "grep -q 'splitFrontmatter' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: extractTriggers method" \
    "grep -q 'extractTriggers' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: parseContentSections method" \
    "grep -q 'parseContentSections' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: saveSection method" \
    "grep -q 'saveSection' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: parseExamples method" \
    "grep -q 'parseExamples' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: parseList method" \
    "grep -q 'parseList' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: parseErrorTable method" \
    "grep -q 'parseErrorTable' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: parseRelated method" \
    "grep -q 'parseRelated' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: extractCategory method" \
    "grep -q 'extractCategory' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: extractTags method" \
    "grep -q 'extractTags' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser: ParseDirectory method" \
    "grep -q 'func (p \*Parser) ParseDirectory' '$PROJECT_ROOT/internal/skills/parser.go'"

run_test "Parser uses YAML parsing" \
    "grep -q 'yaml.Unmarshal' '$PROJECT_ROOT/internal/skills/parser.go'"

# ============================================================================
# SECTION 9: PROTOCOL ADAPTER
# ============================================================================
log_info "=============================================="
log_info "SECTION 9: Protocol Adapter"
log_info "=============================================="

run_test "ProtocolType enum exists" \
    "grep -q 'type ProtocolType string' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "ProtocolMCP constant exists" \
    "grep -q 'ProtocolMCP' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "ProtocolACP constant exists" \
    "grep -q 'ProtocolACP' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "ProtocolLSP constant exists" \
    "grep -q 'ProtocolLSP' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "ProtocolSkillAdapter struct exists" \
    "grep -q 'type ProtocolSkillAdapter struct' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "MCPSkillTool struct exists" \
    "grep -q 'type MCPSkillTool struct' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "ACPSkillAction struct exists" \
    "grep -q 'type ACPSkillAction struct' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "LSPSkillCommand struct exists" \
    "grep -q 'type LSPSkillCommand struct' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "SkillToolCall struct exists" \
    "grep -q 'type SkillToolCall struct' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "SkillToolResult struct exists" \
    "grep -q 'type SkillToolResult struct' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "NewProtocolSkillAdapter function exists" \
    "grep -q 'func NewProtocolSkillAdapter' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "Adapter: RegisterAllSkillsAsTools method" \
    "grep -q 'func (a \*ProtocolSkillAdapter) RegisterAllSkillsAsTools' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "Adapter: GetMCPTools method" \
    "grep -q 'func (a \*ProtocolSkillAdapter) GetMCPTools' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "Adapter: GetACPActions method" \
    "grep -q 'func (a \*ProtocolSkillAdapter) GetACPActions' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "Adapter: GetLSPCommands method" \
    "grep -q 'func (a \*ProtocolSkillAdapter) GetLSPCommands' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "Adapter: InvokeMCPTool method" \
    "grep -q 'func (a \*ProtocolSkillAdapter) InvokeMCPTool' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "Adapter: InvokeACPAction method" \
    "grep -q 'func (a \*ProtocolSkillAdapter) InvokeACPAction' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "Adapter: InvokeLSPCommand method" \
    "grep -q 'func (a \*ProtocolSkillAdapter) InvokeLSPCommand' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "Adapter: ToMCPToolList method" \
    "grep -q 'func (a \*ProtocolSkillAdapter) ToMCPToolList' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "Adapter: ToACPActionList method" \
    "grep -q 'func (a \*ProtocolSkillAdapter) ToACPActionList' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "Adapter: ToLSPCommandList method" \
    "grep -q 'func (a \*ProtocolSkillAdapter) ToLSPCommandList' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

run_test "GetSkillUsageHeader function exists" \
    "grep -q 'func GetSkillUsageHeader' '$PROJECT_ROOT/internal/skills/protocol_adapter.go'"

# ============================================================================
# SECTION 10: INTEGRATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 10: Integration"
log_info "=============================================="

run_test "Integration struct exists" \
    "grep -q 'type Integration struct' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "NewIntegration function exists" \
    "grep -q 'func NewIntegration' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "Integration: ProcessRequest method" \
    "grep -q 'func (i \*Integration) ProcessRequest' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "Integration: CompleteRequest method" \
    "grep -q 'func (i \*Integration) CompleteRequest' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "Integration: RecordToolUse method" \
    "grep -q 'func (i \*Integration) RecordToolUse' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "Integration: BuildSkillsUsedSection method" \
    "grep -q 'func (i \*Integration) BuildSkillsUsedSection' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "Integration: EnhancePromptWithSkills method" \
    "grep -q 'func (i \*Integration) EnhancePromptWithSkills' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "Integration: GetService method" \
    "grep -q 'func (i \*Integration) GetService' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "RequestContext struct exists" \
    "grep -q 'type RequestContext struct' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "SkillsUsedMetadata struct exists" \
    "grep -q 'type SkillsUsedMetadata struct' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "SkillUsedInfo struct exists" \
    "grep -q 'type SkillUsedInfo struct' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "ResponseEnhancer struct exists" \
    "grep -q 'type ResponseEnhancer struct' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "NewResponseEnhancer function exists" \
    "grep -q 'func NewResponseEnhancer' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "DebateIntegration struct exists" \
    "grep -q 'type DebateIntegration struct' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "NewDebateIntegration function exists" \
    "grep -q 'func NewDebateIntegration' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "MCPIntegration struct exists" \
    "grep -q 'type MCPIntegration struct' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "NewMCPIntegration function exists" \
    "grep -q 'func NewMCPIntegration' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "ACPIntegration struct exists" \
    "grep -q 'type ACPIntegration struct' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "NewACPIntegration function exists" \
    "grep -q 'func NewACPIntegration' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "LSPIntegration struct exists" \
    "grep -q 'type LSPIntegration struct' '$PROJECT_ROOT/internal/skills/integration.go'"

run_test "NewLSPIntegration function exists" \
    "grep -q 'func NewLSPIntegration' '$PROJECT_ROOT/internal/skills/integration.go'"

# ============================================================================
# SECTION 11: GO TESTS VALIDATION
# ============================================================================
log_info "=============================================="
log_info "SECTION 11: Go Tests Validation"
log_info "=============================================="

log_info "Running Skills package tests..."
run_test "All Skills tests pass" \
    "cd '$PROJECT_ROOT' && go test -v ./internal/skills/... -timeout 120s"

# ============================================================================
# SUMMARY
# ============================================================================
log_info "=============================================="
log_info "CHALLENGE SUMMARY"
log_info "=============================================="

log_info "Total tests: $TESTS_TOTAL"
log_info "Passed: $TESTS_PASSED"
log_info "Failed: $TESTS_FAILED"

if [[ $TESTS_FAILED -eq 0 ]]; then
    log_success "=============================================="
    log_success "ALL SKILLS TESTS PASSED!"
    log_success "=============================================="
    finalize_challenge 0
else
    log_error "=============================================="
    log_error "SOME TESTS FAILED"
    log_error "=============================================="
    finalize_challenge 1
fi
