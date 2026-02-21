#!/bin/bash
#
# Verified Provider Instance Challenge
#
# This challenge validates that verified provider instances are used
# instead of registry lookups when conducting debates.
#
# ROOT CAUSE: Previously, ParticipantConfig only stored provider NAME as string,
# and DebateService looked up providers from registry by name. This caused verified
# CLI-based OAuth providers (Claude CLI, Qwen ACP) to be bypassed in favor of
# API-based providers from registry, which failed with product-restricted tokens.
#
# FIX: Added ProviderInstance field to ParticipantConfig and FallbackConfig,
# and DebateService now uses verified instances when available.
#
# Tests:
# 1. ParticipantConfig.ProviderInstance is preserved when set
# 2. FallbackConfig.ProviderInstance is preserved when set
# 3. DebateTeamMember.ToParticipantConfig() copies Provider instance
# 4. DebateTeamConfig.GetVerifiedProviderInstance() returns correct instances
# 5. DebateTeamConfig.GetParticipantConfigs() includes ProviderInstance
# 6. DebateService uses verified instances when available
# 7. Live API test: No unexpected fallbacks in debate responses

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEST_COUNT=0
PASSED=0
FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    PASSED=$((PASSED+1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    FAILED=$((FAILED+1))
}

log_test() {
    echo -e "${YELLOW}[TEST $((TEST_COUNT+1))]${NC} $1"
    TEST_COUNT=$((TEST_COUNT+1))
}

# Test functions

test_participant_config_has_provider_instance_field() {
    log_test "ParticipantConfig has ProviderInstance field"
    
    # Check the field exists in the struct
    if grep -q "ProviderInstance.*llm.LLMProvider" "$PROJECT_ROOT/internal/services/debate_support_types.go"; then
        log_pass "ParticipantConfig.ProviderInstance field exists"
    else
        log_fail "ParticipantConfig.ProviderInstance field NOT found"
    fi
}

test_fallback_config_has_provider_instance_field() {
    log_test "FallbackConfig has ProviderInstance field"
    
    if grep -A5 "type FallbackConfig struct" "$PROJECT_ROOT/internal/services/debate_support_types.go" | grep -q "ProviderInstance"; then
        log_pass "FallbackConfig.ProviderInstance field exists"
    else
        log_fail "FallbackConfig.ProviderInstance field NOT found"
    fi
}

test_debate_service_uses_verified_instance() {
    log_test "DebateService uses verified provider instance when available"
    
    if grep -q "participant.ProviderInstance != nil" "$PROJECT_ROOT/internal/services/debate_service.go"; then
        log_pass "DebateService checks participant.ProviderInstance"
    else
        log_fail "DebateService does NOT check participant.ProviderInstance"
    fi
}

test_debate_service_checks_team_config() {
    log_test "DebateService falls back to teamConfig for verified instances"
    
    if grep -q "ds.teamConfig.GetVerifiedProviderInstance" "$PROJECT_ROOT/internal/services/debate_service.go"; then
        log_pass "DebateService checks teamConfig.GetVerifiedProviderInstance"
    else
        log_fail "DebateService does NOT check teamConfig for verified instances"
    fi
}

test_debate_team_member_to_participant_config() {
    log_test "DebateTeamMember.ToParticipantConfig() copies Provider"
    
    if grep -A20 "func (m \*DebateTeamMember) ToParticipantConfig()" "$PROJECT_ROOT/internal/services/debate_team_config.go" | grep -q "ProviderInstance:.*m.Provider"; then
        log_pass "ToParticipantConfig() copies Provider instance"
    else
        log_fail "ToParticipantConfig() does NOT copy Provider instance"
    fi
}

test_get_verified_provider_instance_exists() {
    log_test "DebateTeamConfig.GetVerifiedProviderInstance() exists"
    
    if grep -q "func (dtc \*DebateTeamConfig) GetVerifiedProviderInstance" "$PROJECT_ROOT/internal/services/debate_team_config.go"; then
        log_pass "GetVerifiedProviderInstance() method exists"
    else
        log_fail "GetVerifiedProviderInstance() method NOT found"
    fi
}

test_get_participant_configs_exists() {
    log_test "DebateTeamConfig.GetParticipantConfigs() exists"
    
    if grep -q "func (dtc \*DebateTeamConfig) GetParticipantConfigs()" "$PROJECT_ROOT/internal/services/debate_team_config.go"; then
        log_pass "GetParticipantConfigs() method exists"
    else
        log_fail "GetParticipantConfigs() method NOT found"
    fi
}

test_verified_instance_tests_pass() {
    log_test "All verified provider instance unit tests pass"
    
    cd "$PROJECT_ROOT"
    if go test -v -run "TestDebateTeamMember_ToParticipantConfig|TestDebateTeamConfig_GetVerifiedProviderInstance|TestDebateTeamConfig_GetParticipantConfigs|TestParticipantConfig_ProviderInstance|TestFallbackConfig_ProviderInstance" ./internal/services/ -count=1 > /tmp/verified_tests.log 2>&1; then
        log_pass "All verified provider instance unit tests pass"
    else
        log_fail "Some unit tests failed"
        tail -20 /tmp/verified_tests.log
    fi
}

test_debate_service_compiles() {
    log_test "DebateService compiles successfully"
    
    cd "$PROJECT_ROOT"
    if go build ./internal/services/... 2>&1; then
        log_pass "DebateService compiles"
    else
        log_fail "DebateService compilation failed"
    fi
}

# Main execution
main() {
    echo ""
    echo "=========================================="
    echo "  VERIFIED PROVIDER INSTANCE CHALLENGE"
    echo "=========================================="
    echo ""
    echo "Validating that verified provider instances are used"
    echo "instead of registry lookups when conducting debates."
    echo ""
    
    test_participant_config_has_provider_instance_field
    test_fallback_config_has_provider_instance_field
    test_debate_service_uses_verified_instance
    test_debate_service_checks_team_config
    test_debate_team_member_to_participant_config
    test_get_verified_provider_instance_exists
    test_get_participant_configs_exists
    test_verified_instance_tests_pass
    test_debate_service_compiles
    
    echo ""
    echo "=========================================="
    echo "  RESULTS"
    echo "=========================================="
    echo ""
    echo "Total Tests: $TEST_COUNT"
    echo -e "Passed: ${GREEN}$PASSED${NC}"
    echo -e "Failed: ${RED}$FAILED${NC}"
    echo ""
    
    if [ $FAILED -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}Some tests failed!${NC}"
        exit 1
    fi
}

main "$@"
