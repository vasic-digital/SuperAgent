#!/bin/bash
# SuperAgent Challenges - Configuration Verification Script
# Usage: ./scripts/verify_config.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_check() { echo -ne "${BLUE}[CHECK]${NC} $1... "; }
print_ok() { echo -e "${GREEN}OK${NC}"; }
print_fail() { echo -e "${RED}FAIL${NC}"; }
print_warn() { echo -e "${YELLOW}WARN${NC}"; }
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }

CONFIGURED=0
NOT_CONFIGURED=0
ERRORS=0

check_env_var() {
    local var_name=$1
    local required=${2:-false}

    print_check "$var_name"

    if [ -n "${!var_name}" ]; then
        # Mask the value for display
        local value="${!var_name}"
        local masked="${value:0:4}****"
        echo -e "${GREEN}configured${NC} ($masked)"
        CONFIGURED=$((CONFIGURED + 1))
        return 0
    else
        if [ "$required" = true ]; then
            echo -e "${RED}not configured (REQUIRED)${NC}"
            ERRORS=$((ERRORS + 1))
        else
            echo -e "${YELLOW}not configured${NC}"
            NOT_CONFIGURED=$((NOT_CONFIGURED + 1))
        fi
        return 1
    fi
}

echo "=========================================="
echo "  SuperAgent Challenges - Config Check"
echo "=========================================="
echo ""

# Load .env if exists
if [ -f "$CHALLENGES_DIR/.env" ]; then
    print_info "Loading .env file..."
    set -a
    source "$CHALLENGES_DIR/.env"
    set +a
else
    print_info "No .env file found. Checking system environment..."
fi

echo ""
echo "--- Primary SuperAgent Providers ---"
check_env_var "ANTHROPIC_API_KEY"
check_env_var "OPENAI_API_KEY"
check_env_var "DEEPSEEK_API_KEY"
check_env_var "GEMINI_API_KEY"
check_env_var "OPENROUTER_API_KEY"
check_env_var "QWEN_API_KEY"
check_env_var "ZAI_API_KEY"
check_env_var "OLLAMA_BASE_URL"

echo ""
echo "--- LLMsVerifier Extended Providers ---"
check_env_var "HUGGINGFACE_API_KEY"
check_env_var "NVIDIA_API_KEY"
check_env_var "CHUTES_API_KEY"
check_env_var "SILICONFLOW_API_KEY"
check_env_var "KIMI_API_KEY"
check_env_var "MISTRAL_API_KEY"
check_env_var "CODESTRAL_API_KEY"
check_env_var "VERCEL_AI_API_KEY"
check_env_var "CEREBRAS_API_KEY"
check_env_var "CLOUDFLARE_API_KEY"
check_env_var "FIREWORKS_API_KEY"
check_env_var "BASETEN_API_KEY"
check_env_var "NOVITA_API_KEY"
check_env_var "UPSTAGE_API_KEY"
check_env_var "NLP_CLOUD_API_KEY"
check_env_var "MODAL_API_KEY"
check_env_var "INFERENCE_API_KEY"

echo ""
echo "--- Challenge Configuration ---"
check_env_var "DEBATE_GROUP_SIZE"
check_env_var "DEBATE_FALLBACKS_PER_MEMBER"
check_env_var "VERIFICATION_TIMEOUT_SECONDS"
check_env_var "API_TEST_TIMEOUT_SECONDS"

echo ""
echo "--- External Tools ---"
print_check "LLMsVerifier path"
if [ -d "$CHALLENGES_DIR/../LLMsVerifier" ]; then
    echo -e "${GREEN}found${NC} ($CHALLENGES_DIR/../LLMsVerifier)"
elif [ -n "$LLMSVERIFIER_PATH" ] && [ -d "$LLMSVERIFIER_PATH" ]; then
    echo -e "${GREEN}found${NC} ($LLMSVERIFIER_PATH)"
else
    echo -e "${YELLOW}not found${NC}"
fi

print_check "Go installation"
if command -v go &> /dev/null; then
    echo -e "${GREEN}$(go version | cut -d' ' -f3)${NC}"
else
    echo -e "${RED}not found${NC}"
    ERRORS=$((ERRORS + 1))
fi

echo ""
echo "=========================================="
echo "  Summary"
echo "=========================================="
echo -e "Configured providers: ${GREEN}$CONFIGURED${NC}"
echo -e "Not configured:       ${YELLOW}$NOT_CONFIGURED${NC}"
echo -e "Errors:               ${RED}$ERRORS${NC}"
echo ""

# Minimum requirement check
if [ $CONFIGURED -ge 1 ] && [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}Configuration valid!${NC}"
    echo "At least one provider is configured. You can run challenges."
    exit 0
elif [ $ERRORS -gt 0 ]; then
    echo -e "${RED}Configuration has errors!${NC}"
    echo "Please fix the required configuration issues."
    exit 1
else
    echo -e "${YELLOW}No providers configured!${NC}"
    echo "Please configure at least one API key in .env"
    exit 1
fi
