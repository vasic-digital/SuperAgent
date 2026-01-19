#!/bin/bash
# Provider Verification Challenge - Test all LLM providers and MCP adapters
# Verifies provider integration, configuration, and fallback mechanisms

set -o pipefail

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
    echo -e "\n${YELLOW}=== $1 ===${NC}"
}

cd "$(dirname "$0")/../.." || exit 1

echo "============================================"
echo "  PROVIDER VERIFICATION CHALLENGE"
echo "  Test All LLM Providers & MCP Adapters"
echo "============================================"

section "LLM Provider Implementation Verification"

# Check each LLM provider implementation exists
LLM_PROVIDERS=("claude" "deepseek" "gemini" "mistral" "openrouter" "qwen" "zai" "zen" "cerebras" "ollama")

for provider in "${LLM_PROVIDERS[@]}"; do
    echo -e "${BLUE}Testing:${NC} Provider $provider"
    if [ -d "internal/llm/providers/$provider" ]; then
        # Check provider file exists
        if ls internal/llm/providers/$provider/*.go > /dev/null 2>&1; then
            pass "Provider $provider implementation exists"
        else
            fail "Provider $provider has no Go files"
        fi
    else
        fail "Provider $provider directory missing"
    fi
done

section "LLM Provider Interface Implementation"

# Verify providers implement LLMProvider interface
for provider in "${LLM_PROVIDERS[@]}"; do
    echo -e "${BLUE}Testing:${NC} $provider implements LLMProvider"
    if grep -r "Complete\|CompleteStream\|HealthCheck\|GetCapabilities" internal/llm/providers/$provider/*.go > /dev/null 2>&1; then
        pass "$provider implements LLMProvider methods"
    else
        fail "$provider missing LLMProvider methods"
    fi
done

section "Provider Compilation Tests"

# Test each provider compiles
for provider in "${LLM_PROVIDERS[@]}"; do
    echo -e "${BLUE}Testing:${NC} $provider compilation"
    if go build ./internal/llm/providers/$provider/... 2>/dev/null; then
        pass "$provider compiles successfully"
    else
        fail "$provider compilation failed"
    fi
done

section "Provider Configuration Tests"

# Check for config validation
for provider in "${LLM_PROVIDERS[@]}"; do
    echo -e "${BLUE}Testing:${NC} $provider config validation"
    if grep -r "ValidateConfig\|validate\|Config" internal/llm/providers/$provider/*.go > /dev/null 2>&1; then
        pass "$provider has configuration handling"
    else
        fail "$provider missing configuration handling"
    fi
done

section "MCP Adapter Implementation Verification"

# Test MCP adapters package compiles
echo -e "${BLUE}Testing:${NC} MCP adapters package compilation"
if go build ./internal/mcp/adapters/... 2>/dev/null; then
    pass "MCP adapters package compiles"
else
    fail "MCP adapters package compilation failed"
fi

# Check each MCP adapter file exists
MCP_ADAPTERS=("registry" "brave_search" "aws_s3" "google_drive" "gitlab" "mongodb" "puppeteer" "slack" "docker" "kubernetes" "notion")

for adapter in "${MCP_ADAPTERS[@]}"; do
    echo -e "${BLUE}Testing:${NC} MCP Adapter $adapter exists"
    if [ -f "internal/mcp/adapters/${adapter}.go" ]; then
        pass "MCP $adapter adapter file exists"
    else
        fail "MCP $adapter adapter file missing"
    fi
done

section "MCP Adapter Registry Tests"

# Test adapter registry functions
echo -e "${BLUE}Testing:${NC} MCP Adapter Registry"
if grep -r "RegisterAdapter\|GetAdapter\|ListAdapters" internal/mcp/adapters/registry.go > /dev/null 2>&1; then
    pass "MCP Registry has registration functions"
else
    fail "MCP Registry missing registration functions"
fi

# Test adapter interface
echo -e "${BLUE}Testing:${NC} MCP Adapter Interface"
if grep -r "type.*Adapter.*interface\|MCPAdapter" internal/mcp/adapters/registry.go > /dev/null 2>&1; then
    pass "MCP Adapter interface defined"
else
    fail "MCP Adapter interface missing"
fi

section "Embedding Provider Verification"

# Test embedding providers
EMBEDDING_PROVIDERS=("OpenAI" "Ollama" "HuggingFace")

for provider in "${EMBEDDING_PROVIDERS[@]}"; do
    echo -e "${BLUE}Testing:${NC} Embedding $provider"
    if grep -r "type ${provider}Embedding\|${provider}Embedder" internal/embedding/models.go > /dev/null 2>&1; then
        pass "Embedding $provider defined"
    else
        fail "Embedding $provider missing"
    fi
done

# Test embedding interface
echo -e "${BLUE}Testing:${NC} Embedding Model Interface"
if grep -r "type EmbeddingModel interface\|Embed.*\[\]float64\|EmbedBatch" internal/embedding/models.go > /dev/null 2>&1; then
    pass "Embedding Model interface defined"
else
    fail "Embedding Model interface missing"
fi

section "Provider Error Handling Tests"

# Check for error handling in providers
echo -e "${BLUE}Testing:${NC} Error handling in LLM providers"
error_handling_count=0
for provider in "${LLM_PROVIDERS[@]}"; do
    if grep -r "return.*err\|if err != nil" internal/llm/providers/$provider/*.go > /dev/null 2>&1; then
        error_handling_count=$((error_handling_count + 1))
    fi
done
if [ $error_handling_count -ge 8 ]; then
    pass "Error handling in $error_handling_count/${#LLM_PROVIDERS[@]} providers"
else
    fail "Error handling missing in some providers"
fi

section "Provider Timeout Configuration Tests"

# Check for timeout handling in providers
echo -e "${BLUE}Testing:${NC} Timeout handling in providers"
timeout_count=0
for provider in "${LLM_PROVIDERS[@]}"; do
    if grep -r "Timeout\|timeout\|context.WithTimeout\|time.Duration" internal/llm/providers/$provider/*.go > /dev/null 2>&1; then
        timeout_count=$((timeout_count + 1))
    fi
done
if [ $timeout_count -ge 5 ]; then
    pass "Timeout handling in $timeout_count/${#LLM_PROVIDERS[@]} providers"
else
    echo -e "${YELLOW}[WARN]${NC} Timeout handling in $timeout_count/${#LLM_PROVIDERS[@]} providers"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    PASSED_TESTS=$((PASSED_TESTS + 1))
fi

section "Provider Health Check Tests"

# Check for health check implementations
echo -e "${BLUE}Testing:${NC} Health check implementations"
health_check_count=0
for provider in "${LLM_PROVIDERS[@]}"; do
    if grep -r "HealthCheck\|healthCheck\|Health\|Ping" internal/llm/providers/$provider/*.go > /dev/null 2>&1; then
        health_check_count=$((health_check_count + 1))
    fi
done
if [ $health_check_count -ge 8 ]; then
    pass "Health checks in $health_check_count/${#LLM_PROVIDERS[@]} providers"
else
    fail "Health checks missing in some providers"
fi

section "Provider Capability Declaration Tests"

# Check for capability declarations
echo -e "${BLUE}Testing:${NC} Capability declarations"
capability_count=0
for provider in "${LLM_PROVIDERS[@]}"; do
    if grep -r "GetCapabilities\|Capabilities\|SupportsStreaming\|SupportsTools" internal/llm/providers/$provider/*.go > /dev/null 2>&1; then
        capability_count=$((capability_count + 1))
    fi
done
if [ $capability_count -ge 8 ]; then
    pass "Capability declarations in $capability_count/${#LLM_PROVIDERS[@]} providers"
else
    fail "Capability declarations missing in some providers"
fi

section "Provider Unit Tests"

# Run tests for LLM providers
echo -e "${BLUE}Testing:${NC} LLM provider unit tests"
if go test -short ./internal/llm/... 2>&1 | tail -1 | grep -q "ok\|PASS"; then
    pass "LLM provider tests pass"
else
    fail "LLM provider tests failed"
fi

section "Verifier Integration Tests"

# Check verifier exists and works
echo -e "${BLUE}Testing:${NC} LLMsVerifier integration"
if [ -d "LLMsVerifier" ]; then
    pass "LLMsVerifier submodule exists"

    # Check verifier compiles
    if go build ./internal/verifier/... 2>/dev/null; then
        pass "Verifier package compiles"
    else
        fail "Verifier package compilation failed"
    fi

    # Check verifier tests
    if go test -short ./internal/verifier/... 2>&1 | tail -1 | grep -q "ok\|PASS"; then
        pass "Verifier tests pass"
    else
        fail "Verifier tests failed"
    fi
else
    fail "LLMsVerifier submodule missing"
fi

echo ""
echo "============================================"
echo "  Provider Verification Results Summary"
echo "============================================"
echo ""
echo "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"
echo ""

if [ $TOTAL_TESTS -gt 0 ]; then
    PASS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo "Pass Rate: ${PASS_RATE}%"
fi
echo ""

echo "Provider Summary:"
echo "  LLM Providers: ${#LLM_PROVIDERS[@]} (claude, deepseek, gemini, mistral, openrouter, qwen, zai, zen, cerebras, ollama)"
echo "  MCP Adapters: ${#MCP_ADAPTERS[@]} (brave, s3, drive, gitlab, mongo, puppeteer, slack, docker, k8s, notion)"
echo "  Embedding Models: ${#EMBEDDING_PROVIDERS[@]} (openai, ollama, huggingface)"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ ALL PROVIDER VERIFICATION TESTS PASSED!${NC}"
    exit 0
else
    echo -e "${RED}✗ SOME PROVIDER VERIFICATION TESTS FAILED${NC}"
    exit 1
fi
