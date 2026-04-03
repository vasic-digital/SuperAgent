#!/bin/bash
# Test runner for all provider tests
# Usage: ./scripts/test_providers.sh [provider_name]

set -e

cd "$(dirname "$0")/.."

echo "=== HelixAgent Provider Test Runner ==="
echo

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check for API keys
check_key() {
    local key_name=$1
    local key_value=$2
    if [ -z "$key_value" ]; then
        echo -e "${YELLOW}⚠ $key_name not set - skipping${NC}"
        return 1
    fi
    echo -e "${GREEN}✓ $key_name found${NC}"
    return 0
}

echo "Checking API keys..."
check_key "OPENAI_API_KEY" "$OPENAI_API_KEY"
check_key "ANTHROPIC_API_KEY" "$ANTHROPIC_API_KEY"
check_key "DEEPSEEK_API_KEY" "$DEEPSEEK_API_KEY"
check_key "GROQ_API_KEY" "$GROQ_API_KEY"
check_key "MISTRAL_API_KEY" "$MISTRAL_API_KEY"
check_key "GEMINI_API_KEY" "$GEMINI_API_KEY"
check_key "COHERE_API_KEY" "$COHERE_API_KEY"
check_key "PERPLEXITY_API_KEY" "$PERPLEXITY_API_KEY"
echo

# Run tests for specific provider or all
if [ -n "$1" ]; then
    PROVIDER=$1
    echo "Running tests for: $PROVIDER"
    
    case $PROVIDER in
        openai)
            go test -v ./tests/providers/openai_test.go -timeout 60s
            ;;
        anthropic|claude)
            go test -v ./tests/providers/anthropic_test.go -timeout 60s
            ;;
        deepseek)
            go test -v ./tests/providers/deepseek_test.go -timeout 60s
            ;;
        groq)
            go test -v ./tests/providers/groq_test.go -timeout 60s
            ;;
        mistral)
            go test -v ./tests/providers/mistral_test.go -timeout 60s
            ;;
        gemini)
            go test -v ./tests/providers/gemini_test.go -timeout 60s
            ;;
        cohere)
            go test -v ./tests/providers/cohere_test.go -timeout 60s
            ;;
        perplexity)
            go test -v ./tests/providers/perplexity_test.go -timeout 60s
            ;;
        *)
            echo -e "${RED}Unknown provider: $PROVIDER${NC}"
            echo "Valid providers: openai, anthropic, deepseek, groq, mistral, gemini, cohere, perplexity"
            exit 1
            ;;
    esac
else
    echo "Running all provider tests..."
    echo "This may take several minutes and incur API costs."
    echo "Press Ctrl+C to cancel, or wait 3 seconds to continue..."
    sleep 3
    
    go test -v ./tests/providers/... -timeout 300s
fi

echo
echo -e "${GREEN}=== Tests Complete ===${NC}"
