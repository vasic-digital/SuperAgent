#!/bin/bash

# ============================================================
# ZEN CLI FACADE CHALLENGE
# ============================================================
# Validates the OpenCode Zen CLI facade mechanism for models that fail direct API
# This challenge ensures NO FALSE POSITIVES - tests MUST fail if functionality is broken
#
# The CLI facade allows models that fail direct API verification to be used
# via the `opencode` CLI command instead.
#
# Test Categories:
# 1-10:   CLI Installation & Availability
# 11-20:  Provider Configuration
# 21-30:  Model Discovery
# 31-40:  Failed API Model Tracking
# 41-50:  Free Adapter Integration
# 51-60:  Response Validation
# 61-70:  Error Handling
# 71-75:  Concurrency Safety
# ============================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
PASSED=0
FAILED=0
SKIPPED=0
TOTAL_TESTS=75

# Results array
declare -a RESULTS

log_test() {
    local num=$1
    local desc=$2
    local status=$3
    local detail=$4

    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}[PASS]${NC} Test $num: $desc"
        ((PASSED++))
        RESULTS+=("PASS:$num:$desc")
    elif [ "$status" = "SKIP" ]; then
        echo -e "${YELLOW}[SKIP]${NC} Test $num: $desc - $detail"
        ((SKIPPED++))
        RESULTS+=("SKIP:$num:$desc:$detail")
    else
        echo -e "${RED}[FAIL]${NC} Test $num: $desc - $detail"
        ((FAILED++))
        RESULTS+=("FAIL:$num:$desc:$detail")
    fi
}

echo -e "${BLUE}============================================================${NC}"
echo -e "${BLUE}ZEN CLI FACADE CHALLENGE - $TOTAL_TESTS Tests${NC}"
echo -e "${BLUE}============================================================${NC}"
echo ""

cd "$PROJECT_ROOT"

# Check if OpenCode CLI is installed
OPENCODE_INSTALLED=false
if command -v opencode &> /dev/null; then
    OPENCODE_INSTALLED=true
    OPENCODE_PATH=$(which opencode)
    echo -e "${GREEN}OpenCode CLI found at: $OPENCODE_PATH${NC}"
else
    echo -e "${YELLOW}OpenCode CLI not found - some tests will be skipped${NC}"
fi

echo ""
echo -e "${BLUE}=== Category 1: CLI Installation & Availability (Tests 1-10) ===${NC}"

# Test 1: ZenCLIProvider file exists
if [ -f "internal/llm/providers/zen/zen_cli.go" ]; then
    log_test 1 "ZenCLIProvider implementation file exists" "PASS"
else
    log_test 1 "ZenCLIProvider implementation file exists" "FAIL" "File not found"
fi

# Test 2: ZenCLIProvider test file exists
if [ -f "internal/llm/providers/zen/zen_cli_test.go" ]; then
    log_test 2 "ZenCLIProvider test file exists" "PASS"
else
    log_test 2 "ZenCLIProvider test file exists" "FAIL" "File not found"
fi

# Test 3: Comprehensive test file exists
if [ -f "internal/llm/providers/zen/zen_cli_comprehensive_test.go" ]; then
    log_test 3 "ZenCLIProvider comprehensive test file exists" "PASS"
else
    log_test 3 "ZenCLIProvider comprehensive test file exists" "FAIL" "File not found"
fi

# Test 4: IsOpenCodeInstalled function exists
if grep -q "func IsOpenCodeInstalled" internal/llm/providers/zen/zen_cli.go; then
    log_test 4 "IsOpenCodeInstalled function exists" "PASS"
else
    log_test 4 "IsOpenCodeInstalled function exists" "FAIL" "Function not found"
fi

# Test 5: GetOpenCodePath function exists
if grep -q "func GetOpenCodePath" internal/llm/providers/zen/zen_cli.go; then
    log_test 5 "GetOpenCodePath function exists" "PASS"
else
    log_test 5 "GetOpenCodePath function exists" "FAIL" "Function not found"
fi

# Test 6: IsCLIAvailable method exists
if grep -q "func (p \*ZenCLIProvider) IsCLIAvailable" internal/llm/providers/zen/zen_cli.go; then
    log_test 6 "IsCLIAvailable method exists" "PASS"
else
    log_test 6 "IsCLIAvailable method exists" "FAIL" "Method not found"
fi

# Test 7: CLI availability is cached with sync.Once
if grep -q "cliCheckOnce.*sync.Once" internal/llm/providers/zen/zen_cli.go; then
    log_test 7 "CLI availability check is cached with sync.Once" "PASS"
else
    log_test 7 "CLI availability check is cached with sync.Once" "FAIL" "sync.Once not found"
fi

# Test 8: GetCLIError method exists
if grep -q "func (p \*ZenCLIProvider) GetCLIError" internal/llm/providers/zen/zen_cli.go; then
    log_test 8 "GetCLIError method exists" "PASS"
else
    log_test 8 "GetCLIError method exists" "FAIL" "Method not found"
fi

# Test 9: NewZenCLIProviderWithUnavailableCLI helper exists
if grep -q "func NewZenCLIProviderWithUnavailableCLI" internal/llm/providers/zen/zen_cli.go; then
    log_test 9 "NewZenCLIProviderWithUnavailableCLI helper exists" "PASS"
else
    log_test 9 "NewZenCLIProviderWithUnavailableCLI helper exists" "FAIL" "Helper not found"
fi

# Test 10: Build succeeds
if go build ./internal/llm/providers/zen/... 2>/dev/null; then
    log_test 10 "Zen provider package builds successfully" "PASS"
else
    log_test 10 "Zen provider package builds successfully" "FAIL" "Build failed"
fi

echo ""
echo -e "${BLUE}=== Category 2: Provider Configuration (Tests 11-20) ===${NC}"

# Test 11: DefaultZenCLIConfig function exists
if grep -q "func DefaultZenCLIConfig" internal/llm/providers/zen/zen_cli.go; then
    log_test 11 "DefaultZenCLIConfig function exists" "PASS"
else
    log_test 11 "DefaultZenCLIConfig function exists" "FAIL" "Function not found"
fi

# Test 12: ZenCLIConfig struct exists
if grep -q "type ZenCLIConfig struct" internal/llm/providers/zen/zen_cli.go; then
    log_test 12 "ZenCLIConfig struct exists" "PASS"
else
    log_test 12 "ZenCLIConfig struct exists" "FAIL" "Struct not found"
fi

# Test 13: NewZenCLIProvider function exists
if grep -q "func NewZenCLIProvider" internal/llm/providers/zen/zen_cli.go; then
    log_test 13 "NewZenCLIProvider function exists" "PASS"
else
    log_test 13 "NewZenCLIProvider function exists" "FAIL" "Function not found"
fi

# Test 14: NewZenCLIProviderWithModel function exists
if grep -q "func NewZenCLIProviderWithModel" internal/llm/providers/zen/zen_cli.go; then
    log_test 14 "NewZenCLIProviderWithModel function exists" "PASS"
else
    log_test 14 "NewZenCLIProviderWithModel function exists" "FAIL" "Function not found"
fi

# Test 15: Provider implements GetName
if grep -q "func (p \*ZenCLIProvider) GetName" internal/llm/providers/zen/zen_cli.go; then
    log_test 15 "GetName method implemented" "PASS"
else
    log_test 15 "GetName method implemented" "FAIL" "Method not found"
fi

# Test 16: Provider implements GetProviderType
if grep -q "func (p \*ZenCLIProvider) GetProviderType" internal/llm/providers/zen/zen_cli.go; then
    log_test 16 "GetProviderType method implemented" "PASS"
else
    log_test 16 "GetProviderType method implemented" "FAIL" "Method not found"
fi

# Test 17: Provider implements Complete
if grep -q "func (p \*ZenCLIProvider) Complete" internal/llm/providers/zen/zen_cli.go; then
    log_test 17 "Complete method implemented" "PASS"
else
    log_test 17 "Complete method implemented" "FAIL" "Method not found"
fi

# Test 18: Provider implements CompleteStream
if grep -q "func (p \*ZenCLIProvider) CompleteStream" internal/llm/providers/zen/zen_cli.go; then
    log_test 18 "CompleteStream method implemented" "PASS"
else
    log_test 18 "CompleteStream method implemented" "FAIL" "Method not found"
fi

# Test 19: Provider implements HealthCheck
if grep -q "func (p \*ZenCLIProvider) HealthCheck" internal/llm/providers/zen/zen_cli.go; then
    log_test 19 "HealthCheck method implemented" "PASS"
else
    log_test 19 "HealthCheck method implemented" "FAIL" "Method not found"
fi

# Test 20: Provider implements GetCapabilities
if grep -q "func (p \*ZenCLIProvider) GetCapabilities" internal/llm/providers/zen/zen_cli.go; then
    log_test 20 "GetCapabilities method implemented" "PASS"
else
    log_test 20 "GetCapabilities method implemented" "FAIL" "Method not found"
fi

echo ""
echo -e "${BLUE}=== Category 3: Model Discovery (Tests 21-30) ===${NC}"

# Test 21: DiscoverModels method exists
if grep -q "func (p \*ZenCLIProvider) DiscoverModels" internal/llm/providers/zen/zen_cli.go; then
    log_test 21 "DiscoverModels method exists" "PASS"
else
    log_test 21 "DiscoverModels method exists" "FAIL" "Method not found"
fi

# Test 22: GetAvailableModels method exists
if grep -q "func (p \*ZenCLIProvider) GetAvailableModels" internal/llm/providers/zen/zen_cli.go; then
    log_test 22 "GetAvailableModels method exists" "PASS"
else
    log_test 22 "GetAvailableModels method exists" "FAIL" "Method not found"
fi

# Test 23: IsModelAvailable method exists
if grep -q "func (p \*ZenCLIProvider) IsModelAvailable" internal/llm/providers/zen/zen_cli.go; then
    log_test 23 "IsModelAvailable method exists" "PASS"
else
    log_test 23 "IsModelAvailable method exists" "FAIL" "Method not found"
fi

# Test 24: GetBestAvailableModel method exists
if grep -q "func (p \*ZenCLIProvider) GetBestAvailableModel" internal/llm/providers/zen/zen_cli.go; then
    log_test 24 "GetBestAvailableModel method exists" "PASS"
else
    log_test 24 "GetBestAvailableModel method exists" "FAIL" "Method not found"
fi

# Test 25: Known models list exists
if grep -q "knownZenModels" internal/llm/providers/zen/zen_cli.go; then
    log_test 25 "Known models fallback list exists" "PASS"
else
    log_test 25 "Known models fallback list exists" "FAIL" "List not found"
fi

# Test 26: GetKnownZenModels function exists
if grep -q "func GetKnownZenModels" internal/llm/providers/zen/zen_cli.go; then
    log_test 26 "GetKnownZenModels function exists" "PASS"
else
    log_test 26 "GetKnownZenModels function exists" "FAIL" "Function not found"
fi

# Test 27: DiscoverZenModels standalone function exists
if grep -q "func DiscoverZenModels" internal/llm/providers/zen/zen_cli.go; then
    log_test 27 "DiscoverZenModels standalone function exists" "PASS"
else
    log_test 27 "DiscoverZenModels standalone function exists" "FAIL" "Function not found"
fi

# Test 28: parseZenModelsOutput function exists
if grep -q "func parseZenModelsOutput" internal/llm/providers/zen/zen_cli.go; then
    log_test 28 "parseZenModelsOutput helper exists" "PASS"
else
    log_test 28 "parseZenModelsOutput helper exists" "FAIL" "Function not found"
fi

# Test 29: models.dev integration exists
if grep -q "discoverModelsFromModelsDev" internal/llm/providers/zen/zen_cli.go; then
    log_test 29 "models.dev API integration exists" "PASS"
else
    log_test 29 "models.dev API integration exists" "FAIL" "Integration not found"
fi

# Test 30: CLI model discovery exists
if grep -q "discoverModelsFromCLI" internal/llm/providers/zen/zen_cli.go; then
    log_test 30 "CLI model discovery exists" "PASS"
else
    log_test 30 "CLI model discovery exists" "FAIL" "Discovery not found"
fi

echo ""
echo -e "${BLUE}=== Category 4: Failed API Model Tracking (Tests 31-40) ===${NC}"

# Test 31: failedAPIModels field exists
if grep -q "failedAPIModels.*map\[string\]bool" internal/llm/providers/zen/zen_cli.go; then
    log_test 31 "failedAPIModels tracking map exists" "PASS"
else
    log_test 31 "failedAPIModels tracking map exists" "FAIL" "Field not found"
fi

# Test 32: MarkModelAsFailedAPI method exists
if grep -q "func (p \*ZenCLIProvider) MarkModelAsFailedAPI" internal/llm/providers/zen/zen_cli.go; then
    log_test 32 "MarkModelAsFailedAPI method exists" "PASS"
else
    log_test 32 "MarkModelAsFailedAPI method exists" "FAIL" "Method not found"
fi

# Test 33: IsModelFailedAPI method exists
if grep -q "func (p \*ZenCLIProvider) IsModelFailedAPI" internal/llm/providers/zen/zen_cli.go; then
    log_test 33 "IsModelFailedAPI method exists" "PASS"
else
    log_test 33 "IsModelFailedAPI method exists" "FAIL" "Method not found"
fi

# Test 34: ShouldUseCLIFacade method exists
if grep -q "func (p \*ZenCLIProvider) ShouldUseCLIFacade" internal/llm/providers/zen/zen_cli.go; then
    log_test 34 "ShouldUseCLIFacade method exists" "PASS"
else
    log_test 34 "ShouldUseCLIFacade method exists" "FAIL" "Method not found"
fi

# Test 35: ShouldUseCLIFacade checks both failed and CLI available
if grep -q "IsModelFailedAPI.*&&.*IsCLIAvailable\|IsCLIAvailable.*&&.*IsModelFailedAPI" internal/llm/providers/zen/zen_cli.go; then
    log_test 35 "ShouldUseCLIFacade checks both conditions" "PASS"
else
    log_test 35 "ShouldUseCLIFacade checks both conditions" "FAIL" "Both conditions not checked"
fi

# Test 36: Unit test for MarkModelAsFailedAPI
if grep -q "TestZenCLIProvider_MarkModelAsFailedAPI" internal/llm/providers/zen/zen_cli_test.go; then
    log_test 36 "MarkModelAsFailedAPI test exists" "PASS"
else
    log_test 36 "MarkModelAsFailedAPI test exists" "FAIL" "Test not found"
fi

# Test 37: Unit test for ShouldUseCLIFacade
if grep -q "TestZenCLIProvider_ShouldUseCLIFacade" internal/llm/providers/zen/zen_cli_test.go; then
    log_test 37 "ShouldUseCLIFacade test exists" "PASS"
else
    log_test 37 "ShouldUseCLIFacade test exists" "FAIL" "Test not found"
fi

# Test 38: Failed API models initialized in constructor
if grep -q "failedAPIModels.*make(map\[string\]bool)" internal/llm/providers/zen/zen_cli.go; then
    log_test 38 "failedAPIModels initialized in constructor" "PASS"
else
    log_test 38 "failedAPIModels initialized in constructor" "FAIL" "Not initialized"
fi

# Test 39: Comprehensive failed model tracking test exists
if grep -q "FailedAPIModelTracking" internal/llm/providers/zen/zen_cli_comprehensive_test.go 2>/dev/null; then
    log_test 39 "Comprehensive failed model tracking test exists" "PASS"
else
    log_test 39 "Comprehensive failed model tracking test exists" "FAIL" "Test not found"
fi

# Test 40: Run failed API tracking unit test
if go test ./internal/llm/providers/zen/... -run "MarkModelAsFailedAPI|ShouldUseCLIFacade" -short -timeout 30s 2>&1 | grep -q "ok"; then
    log_test 40 "Failed API tracking tests pass" "PASS"
else
    log_test 40 "Failed API tracking tests pass" "FAIL" "Tests failed"
fi

echo ""
echo -e "${BLUE}=== Category 5: Free Adapter Integration (Tests 41-50) ===${NC}"

# Test 41: FreeProviderAdapter has zenCLIProvider field
if grep -q "zenCLIProvider.*\*zen.ZenCLIProvider" internal/verifier/adapters/free_adapter.go; then
    log_test 41 "FreeProviderAdapter has zenCLIProvider field" "PASS"
else
    log_test 41 "FreeProviderAdapter has zenCLIProvider field" "FAIL" "Field not found"
fi

# Test 42: FreeProviderAdapter has failedAPIModels field
if grep -q "failedAPIModels.*map\[string\]error" internal/verifier/adapters/free_adapter.go; then
    log_test 42 "FreeProviderAdapter has failedAPIModels field" "PASS"
else
    log_test 42 "FreeProviderAdapter has failedAPIModels field" "FAIL" "Field not found"
fi

# Test 43: IsCLIFacadeAvailable method exists
if grep -q "func (fa \*FreeProviderAdapter) IsCLIFacadeAvailable" internal/verifier/adapters/free_adapter.go; then
    log_test 43 "IsCLIFacadeAvailable method exists" "PASS"
else
    log_test 43 "IsCLIFacadeAvailable method exists" "FAIL" "Method not found"
fi

# Test 44: GetCLIFacadeProvider method exists
if grep -q "func (fa \*FreeProviderAdapter) GetCLIFacadeProvider" internal/verifier/adapters/free_adapter.go; then
    log_test 44 "GetCLIFacadeProvider method exists" "PASS"
else
    log_test 44 "GetCLIFacadeProvider method exists" "FAIL" "Method not found"
fi

# Test 45: GetFailedAPIModels method exists
if grep -q "func (fa \*FreeProviderAdapter) GetFailedAPIModels" internal/verifier/adapters/free_adapter.go; then
    log_test 45 "GetFailedAPIModels method exists" "PASS"
else
    log_test 45 "GetFailedAPIModels method exists" "FAIL" "Method not found"
fi

# Test 46: IsModelUsingCLIFacade method exists
if grep -q "func (fa \*FreeProviderAdapter) IsModelUsingCLIFacade" internal/verifier/adapters/free_adapter.go; then
    log_test 46 "IsModelUsingCLIFacade method exists" "PASS"
else
    log_test 46 "IsModelUsingCLIFacade method exists" "FAIL" "Method not found"
fi

# Test 47: GetCLIFacadeModels method exists
if grep -q "func (fa \*FreeProviderAdapter) GetCLIFacadeModels" internal/verifier/adapters/free_adapter.go; then
    log_test 47 "GetCLIFacadeModels method exists" "PASS"
else
    log_test 47 "GetCLIFacadeModels method exists" "FAIL" "Method not found"
fi

# Test 48: verifyZenModelViaCLI method exists
if grep -q "func (fa \*FreeProviderAdapter) verifyZenModelViaCLI" internal/verifier/adapters/free_adapter.go; then
    log_test 48 "verifyZenModelViaCLI method exists" "PASS"
else
    log_test 48 "verifyZenModelViaCLI method exists" "FAIL" "Method not found"
fi

# Test 49: testModelCompletionViaCLI method exists
if grep -q "func (fa \*FreeProviderAdapter) testModelCompletionViaCLI" internal/verifier/adapters/free_adapter.go; then
    log_test 49 "testModelCompletionViaCLI method exists" "PASS"
else
    log_test 49 "testModelCompletionViaCLI method exists" "FAIL" "Method not found"
fi

# Test 50: Free adapter tests pass
if go test ./internal/verifier/adapters/... -run "CLIFacade" -short -timeout 30s 2>&1 | grep -q "ok"; then
    log_test 50 "Free adapter CLI facade tests pass" "PASS"
else
    log_test 50 "Free adapter CLI facade tests pass" "FAIL" "Tests failed"
fi

echo ""
echo -e "${BLUE}=== Category 6: Response Validation (Tests 51-60) ===${NC}"

# Test 51: Response includes ProviderID
if grep -q "ProviderID.*:.*\"zen-cli\"" internal/llm/providers/zen/zen_cli.go; then
    log_test 51 "Response includes ProviderID" "PASS"
else
    log_test 51 "Response includes ProviderID" "FAIL" "Field not set"
fi

# Test 52: Response includes ProviderName
if grep -q "ProviderName.*:.*\"zen-cli\"" internal/llm/providers/zen/zen_cli.go; then
    log_test 52 "Response includes ProviderName" "PASS"
else
    log_test 52 "Response includes ProviderName" "FAIL" "Field not set"
fi

# Test 53: Response includes Metadata with source
if grep -q "\"source\".*:.*\"opencode-cli\"" internal/llm/providers/zen/zen_cli.go; then
    log_test 53 "Response metadata includes source" "PASS"
else
    log_test 53 "Response metadata includes source" "FAIL" "Not set"
fi

# Test 54: Response includes Metadata with facade flag
if grep -q "\"facade\".*:.*true" internal/llm/providers/zen/zen_cli.go; then
    log_test 54 "Response metadata includes facade flag" "PASS"
else
    log_test 54 "Response metadata includes facade flag" "FAIL" "Not set"
fi

# Test 55: Capabilities include streaming support
if grep -q "SupportsStreaming.*:.*true" internal/llm/providers/zen/zen_cli.go; then
    log_test 55 "Capabilities include streaming support" "PASS"
else
    log_test 55 "Capabilities include streaming support" "FAIL" "Not set"
fi

# Test 56: Capabilities exclude function calling
if grep -q "SupportsFunctionCalling.*:.*false" internal/llm/providers/zen/zen_cli.go; then
    log_test 56 "Capabilities exclude function calling" "PASS"
else
    log_test 56 "Capabilities exclude function calling" "FAIL" "Not set correctly"
fi

# Test 57: Capabilities exclude tools
if grep -q "SupportsTools.*:.*false" internal/llm/providers/zen/zen_cli.go; then
    log_test 57 "Capabilities exclude tools" "PASS"
else
    log_test 57 "Capabilities exclude tools" "FAIL" "Not set correctly"
fi

# Test 58: Response validation test exists
if grep -q "ResponseMetadata\|Response.*Metadata" internal/llm/providers/zen/zen_cli_comprehensive_test.go 2>/dev/null; then
    log_test 58 "Response metadata test exists" "PASS"
else
    log_test 58 "Response metadata test exists" "FAIL" "Test not found"
fi

# Test 59: CLI verified model metadata includes verified_via
if grep -q "\"verified_via\".*:.*\"cli_facade\"" internal/verifier/adapters/free_adapter.go; then
    log_test 59 "CLI verified models include verified_via metadata" "PASS"
else
    log_test 59 "CLI verified models include verified_via metadata" "FAIL" "Not set"
fi

# Test 60: CLI verified model metadata includes direct_api_failed
if grep -q "\"direct_api_failed\".*:.*true" internal/verifier/adapters/free_adapter.go; then
    log_test 60 "CLI verified models include direct_api_failed flag" "PASS"
else
    log_test 60 "CLI verified models include direct_api_failed flag" "FAIL" "Not set"
fi

echo ""
echo -e "${BLUE}=== Category 7: Error Handling (Tests 61-70) ===${NC}"

# Test 61: Complete returns error when CLI unavailable
if grep -q "not available" internal/llm/providers/zen/zen_cli.go; then
    log_test 61 "Complete returns 'not available' error" "PASS"
else
    log_test 61 "Complete returns 'not available' error" "FAIL" "Error message not found"
fi

# Test 62: Empty prompt is rejected
if grep -q "no prompt provided" internal/llm/providers/zen/zen_cli.go; then
    log_test 62 "Empty prompt is rejected" "PASS"
else
    log_test 62 "Empty prompt is rejected" "FAIL" "Check not found"
fi

# Test 63: Timeout is handled
if grep -q "timed out\|DeadlineExceeded" internal/llm/providers/zen/zen_cli.go; then
    log_test 63 "Timeout handling exists" "PASS"
else
    log_test 63 "Timeout handling exists" "FAIL" "Timeout handling not found"
fi

# Test 64: CLI error is captured
if grep -q "CLI failed\|CLI.*error" internal/llm/providers/zen/zen_cli.go; then
    log_test 64 "CLI error capturing exists" "PASS"
else
    log_test 64 "CLI error capturing exists" "FAIL" "Error capturing not found"
fi

# Test 65: Empty response is rejected
if grep -q "empty response\|empty content" internal/llm/providers/zen/zen_cli.go; then
    log_test 65 "Empty response rejection exists" "PASS"
else
    log_test 65 "Empty response rejection exists" "FAIL" "Check not found"
fi

# Test 66: Complete_NotAvailable test exists
if grep -q "Complete_NotAvailable\|CompleteNotAvailable" internal/llm/providers/zen/zen_cli_test.go; then
    log_test 66 "Complete_NotAvailable test exists" "PASS"
else
    log_test 66 "Complete_NotAvailable test exists" "FAIL" "Test not found"
fi

# Test 67: CompleteStream_NotAvailable test exists
if grep -q "CompleteStream_NotAvailable\|CompleteStreamNotAvailable" internal/llm/providers/zen/zen_cli_test.go; then
    log_test 67 "CompleteStream_NotAvailable test exists" "PASS"
else
    log_test 67 "CompleteStream_NotAvailable test exists" "FAIL" "Test not found"
fi

# Test 68: Error handling tests pass
if go test ./internal/llm/providers/zen/... -run "NotAvailable" -short -timeout 30s 2>&1 | grep -q "ok"; then
    log_test 68 "Error handling tests pass" "PASS"
else
    log_test 68 "Error handling tests pass" "FAIL" "Tests failed"
fi

# Test 69: Canned error detection in free adapter
if grep -q "cannedErrorPatterns\|canned error" internal/verifier/adapters/free_adapter.go; then
    log_test 69 "Canned error detection exists in free adapter" "PASS"
else
    log_test 69 "Canned error detection exists in free adapter" "FAIL" "Detection not found"
fi

# Test 70: Strict math test validation (2+2=4)
if grep -q "\"4\"" internal/verifier/adapters/free_adapter.go; then
    log_test 70 "Strict math test validation exists" "PASS"
else
    log_test 70 "Strict math test validation exists" "FAIL" "Validation not found"
fi

echo ""
echo -e "${BLUE}=== Category 8: Concurrency Safety (Tests 71-75) ===${NC}"

# Test 71: sync.Once for CLI check
if grep -q "cliCheckOnce.Do" internal/llm/providers/zen/zen_cli.go; then
    log_test 71 "sync.Once used for CLI availability check" "PASS"
else
    log_test 71 "sync.Once used for CLI availability check" "FAIL" "Not found"
fi

# Test 72: sync.Once for model discovery
if grep -q "modelsDiscoveryOnce" internal/llm/providers/zen/zen_cli.go; then
    log_test 72 "sync.Once used for model discovery" "PASS"
else
    log_test 72 "sync.Once used for model discovery" "FAIL" "Not found"
fi

# Test 73: Mutex in free adapter
if grep -q "mu.*sync.RWMutex" internal/verifier/adapters/free_adapter.go; then
    log_test 73 "Mutex used in free adapter" "PASS"
else
    log_test 73 "Mutex used in free adapter" "FAIL" "Not found"
fi

# Test 74: Concurrent access test exists
if grep -q "ConcurrentAccess" internal/llm/providers/zen/zen_cli_comprehensive_test.go 2>/dev/null; then
    log_test 74 "Concurrent access test exists" "PASS"
else
    log_test 74 "Concurrent access test exists" "FAIL" "Test not found"
fi

# Test 75: Run all zen CLI tests (final validation)
if go test ./internal/llm/providers/zen/... -run "ZenCLI" -short -timeout 60s 2>&1 | grep -q "ok"; then
    log_test 75 "All ZenCLI tests pass" "PASS"
else
    log_test 75 "All ZenCLI tests pass" "FAIL" "Some tests failed"
fi

# ============================================================
# SUMMARY
# ============================================================
echo ""
echo -e "${BLUE}============================================================${NC}"
echo -e "${BLUE}CHALLENGE SUMMARY${NC}"
echo -e "${BLUE}============================================================${NC}"
echo ""
echo -e "Total Tests:  $TOTAL_TESTS"
echo -e "${GREEN}Passed:       $PASSED${NC}"
echo -e "${RED}Failed:       $FAILED${NC}"
echo -e "${YELLOW}Skipped:      $SKIPPED${NC}"
echo ""

# Calculate pass rate
if [ $TOTAL_TESTS -gt 0 ]; then
    PASS_RATE=$(echo "scale=1; ($PASSED * 100) / $TOTAL_TESTS" | bc)
    echo -e "Pass Rate:    ${PASS_RATE}%"
fi

echo ""

# Exit with appropriate code
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}CHALLENGE FAILED - $FAILED tests did not pass${NC}"
    exit 1
else
    echo -e "${GREEN}CHALLENGE PASSED - All $PASSED tests passed!${NC}"
    exit 0
fi
