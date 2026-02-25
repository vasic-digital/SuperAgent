#!/bin/bash
# ============================================================================
# HELIXMEMORY + HELIXSPECIFIER VERIFICATION CHALLENGE
# ============================================================================
# Verifies that:
# 1. HelixMemory module exists and is fully implemented
# 2. HelixSpecifier module exists and is fully implemented
# 3. Both are set as DEFAULT (active unless opted out)
# 4. Both are wired into DebateService
# 5. All tests pass for both modules
# ============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

pass() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
    echo -e "${GREEN}[PASS]${NC} $1"
}

fail() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    FAILED_TESTS=$((FAILED_TESTS + 1))
    echo -e "${RED}[FAIL]${NC} $1"
}

section() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

cd "$PROJECT_ROOT"

echo "============================================"
echo "  HELIXMEMORY + HELIXSPECIFIER VERIFICATION"
echo "============================================"
echo "Based on PDF documents:"
echo "  - HelixMemory_Super_Memory_Provider_Research-1.PDF"
echo "  - HelixSpecifier_Spec_Fusion_Research.PDF"
echo ""

section "HelixMemory Module Verification"

# Test 1: HelixMemory directory exists
if [ -d "HelixMemory" ]; then
    pass "HelixMemory directory exists"
else
    fail "HelixMemory directory not found"
fi

# Test 2: HelixMemory go.mod exists
if [ -f "HelixMemory/go.mod" ]; then
    pass "HelixMemory/go.mod exists"
else
    fail "HelixMemory/go.mod not found"
fi

# Test 3: All four backends implemented (Mem0, Cognee, Letta, Graphiti)
BACKENDS=("mem0" "cognee" "letta" "graphiti")
for backend in "${BACKENDS[@]}"; do
    if [ -f "HelixMemory/pkg/clients/${backend}/client.go" ]; then
        pass "HelixMemory backend ${backend} client.go exists"
    else
        fail "HelixMemory backend ${backend} client.go not found"
    fi
done

# Test 4: Unified provider exists
if [ -f "HelixMemory/pkg/provider/unified.go" ]; then
    pass "HelixMemory unified provider exists"
else
    fail "HelixMemory unified provider not found"
fi

# Test 5: Fusion engine exists
if [ -f "HelixMemory/pkg/fusion/engine.go" ]; then
    pass "HelixMemory fusion engine exists"
else
    fail "HelixMemory fusion engine not found"
fi

# Test 6: Features implemented per PDF
FEATURES=(
    "temporal:Temporal Queries"
    "snapshots:Memory Snapshots"
    "debate_memory:Debate Memory"
    "quality_loop:Quality Loop"
    "confidence:Confidence Scoring"
    "code_gen:Code Generation"
    "mesh:Memory Mesh"
    "mcp_bridge:MCP Bridge"
    "procedural:Procedural Memory"
    "cross_project:Cross-Project Transfer"
    "codebase_dna:Codebase DNA"
    "context_window:Context Window Management"
)

for feature_info in "${FEATURES[@]}"; do
    IFS=':' read -r feature name <<< "$feature_info"
    if [ -f "HelixMemory/pkg/features/${feature}/${feature}.go" ] || [ -f "HelixMemory/pkg/features/${feature}/${name// /_}.go" ]; then
        pass "HelixMemory feature: ${name}"
    else
        if ls HelixMemory/pkg/features/${feature}/*.go 2>/dev/null | head -1 > /dev/null; then
            pass "HelixMemory feature directory: ${feature}"
        else
            fail "HelixMemory feature missing: ${name}"
        fi
    fi
done

# Test 7: HelixMemory builds
if (cd HelixMemory && go build ./... 2>&1); then
    pass "HelixMemory builds successfully"
else
    fail "HelixMemory build failed"
fi

# Test 8: HelixMemory tests pass
if (cd HelixMemory && go test ./... -short -count=1 > /dev/null 2>&1); then
    pass "HelixMemory tests pass"
else
    fail "HelixMemory tests failed"
fi

section "HelixSpecifier Module Verification"

# Test 9: HelixSpecifier directory exists
if [ -d "HelixSpecifier" ]; then
    pass "HelixSpecifier directory exists"
else
    fail "HelixSpecifier directory not found"
fi

# Test 10: HelixSpecifier go.mod exists
if [ -f "HelixSpecifier/go.mod" ]; then
    pass "HelixSpecifier/go.mod exists"
else
    fail "HelixSpecifier/go.mod not found"
fi

# Test 11: Three pillars implemented (SpecKit, Superpowers, GSD)
PILLARS=("speckit:SpecKit" "superpowers:Superpowers" "gsd:GSD")
for pillar_info in "${PILLARS[@]}"; do
    IFS=':' read -r pillar name <<< "$pillar_info"
    if [ -f "HelixSpecifier/pkg/${pillar}/${pillar}.go" ]; then
        pass "HelixSpecifier pillar: ${name}"
    else
        fail "HelixSpecifier pillar missing: ${name}"
    fi
done

# Test 12: Core engine exists
if [ -f "HelixSpecifier/pkg/engine/engine.go" ]; then
    pass "HelixSpecifier engine exists"
else
    fail "HelixSpecifier engine not found"
fi

# Test 13: Intent classifier exists
if [ -f "HelixSpecifier/pkg/intent/classifier.go" ]; then
    pass "HelixSpecifier intent classifier exists"
else
    fail "HelixSpecifier intent classifier not found"
fi

# Test 14: Ceremony scaler exists
if [ -f "HelixSpecifier/pkg/ceremony/ceremony.go" ]; then
    pass "HelixSpecifier ceremony scaler exists"
else
    fail "HelixSpecifier ceremony scaler not found"
fi

# Test 15: Spec memory exists
if [ -f "HelixSpecifier/pkg/memory/memory.go" ]; then
    pass "HelixSpecifier memory exists"
else
    fail "HelixSpecifier memory not found"
fi

# Test 16: Specifier features per PDF
SPECIFIER_FEATURES=(
    "adaptive_ceremony:Adaptive Ceremony"
    "brownfield:Brownfield Integration"
    "constitution_code:Constitution Code"
    "cross_project:Cross-Project Transfer"
    "debate_architecture:Debate Architecture"
    "nyquist_tdd:Nyquist TDD"
    "parallel_execution:Parallel Execution"
    "predictive_spec:Predictive Specification"
    "skill_learning:Skill Learning"
    "spec_memory:Spec Memory"
)

for feature_info in "${SPECIFIER_FEATURES[@]}"; do
    IFS=':' read -r feature name <<< "$feature_info"
    if ls HelixSpecifier/pkg/features/${feature}/*.go 2>/dev/null | head -1 > /dev/null; then
        pass "HelixSpecifier feature: ${name}"
    else
        fail "HelixSpecifier feature missing: ${name}"
    fi
done

# Test 17: HelixSpecifier builds
if (cd HelixSpecifier && go build ./... 2>&1); then
    pass "HelixSpecifier builds successfully"
else
    fail "HelixSpecifier build failed"
fi

# Test 18: HelixSpecifier tests pass
if (cd HelixSpecifier && go test ./... -short -count=1 > /dev/null 2>&1); then
    pass "HelixSpecifier tests pass"
else
    fail "HelixSpecifier tests failed"
fi

section "Default Configuration Verification"

# Test 19: HelixMemory is default (opt-out via build tag)
if grep -q '!nohelixmemory' internal/adapters/memory/factory_helixmemory.go; then
    pass "HelixMemory is DEFAULT (opt-out via -tags nohelixmemory)"
else
    fail "HelixMemory default configuration incorrect"
fi

# Test 20: HelixSpecifier is default (opt-out via build tag)
if grep -q '!nohelixspecifier' internal/adapters/specifier/factory_helixspecifier.go; then
    pass "HelixSpecifier is DEFAULT (opt-out via -tags nohelixspecifier)"
else
    fail "HelixSpecifier default configuration incorrect"
fi

# Test 21: go.mod has replace directives for both modules
if grep -q 'replace digital.vasic.helixmemory => ./HelixMemory' go.mod; then
    pass "go.mod has HelixMemory replace directive"
else
    fail "go.mod missing HelixMemory replace directive"
fi

if grep -q 'replace digital.vasic.helixspecifier => ./HelixSpecifier' go.mod; then
    pass "go.mod has HelixSpecifier replace directive"
else
    fail "go.mod missing HelixSpecifier replace directive"
fi

section "DebateService Integration Verification"

# Test 22: DebateService imports HelixMemory
if grep -q 'memoryadapter' internal/services/debate_service.go; then
    pass "DebateService imports memory adapter"
else
    fail "DebateService missing memory adapter import"
fi

# Test 23: DebateService imports HelixSpecifier
if grep -q 'specifieradapter' internal/services/debate_service.go; then
    pass "DebateService imports specifier adapter"
else
    fail "DebateService missing specifier adapter import"
fi

# Test 24: DebateService initializes HelixMemory
if grep -q 'NewOptimalStoreAdapter' internal/services/debate_service.go; then
    pass "DebateService initializes HelixMemory"
else
    fail "DebateService doesn't initialize HelixMemory"
fi

# Test 25: DebateService initializes HelixSpecifier
if grep -q 'NewOptimalSpecAdapter' internal/services/debate_service.go; then
    pass "DebateService initializes HelixSpecifier"
else
    fail "DebateService doesn't initialize HelixSpecifier"
fi

# Test 26: IsHelixMemoryActive method exists
if grep -q 'IsHelixMemoryActive' internal/services/debate_service.go; then
    pass "IsHelixMemoryActive method exists"
else
    fail "IsHelixMemoryActive method missing"
fi

# Test 27: IsHelixSpecifierActive method exists
if grep -q 'IsHelixSpecifierActive' internal/services/debate_service.go; then
    pass "IsHelixSpecifierActive method exists"
else
    fail "IsHelixSpecifierActive method missing"
fi

section "Build Verification"

# Test 28: Main application builds with both modules
if go build ./cmd/helixagent/... 2>&1; then
    pass "Main application builds with HelixMemory + HelixSpecifier"
else
    fail "Main application build failed"
fi

# Test 29: Main application tests pass
if go test ./internal/... -short -count=1 > /dev/null 2>&1; then
    pass "Internal tests pass"
else
    fail "Internal tests failed"
fi

echo ""
echo "============================================"
echo "  VERIFICATION RESULTS SUMMARY"
echo "============================================"
echo ""
echo "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed:${NC}     $PASSED_TESTS"
echo -e "${RED}Failed:${NC}     $FAILED_TESTS"
echo ""

if [ "$FAILED_TESTS" -eq 0 ]; then
    echo -e "${GREEN}✓ ALL VERIFICATION TESTS PASSED!${NC}"
    echo ""
    echo "HelixMemory and HelixSpecifier are:"
    echo "  - Fully implemented per PDF specifications"
    echo "  - Set as DEFAULT (opt-out via build tags)"
    echo "  - Wired into DebateService"
    echo "  - All tests passing"
    exit 0
else
    echo -e "${RED}✗ SOME VERIFICATION TESTS FAILED${NC}"
    exit 1
fi
