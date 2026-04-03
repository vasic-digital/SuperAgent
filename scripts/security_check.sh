#!/bin/bash
# Security check script to prevent accidental secret commits
# Run this before committing to verify no secrets are present

set -e

echo "=== Security Check ==="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

FAILED=0

# Check 1: .env is in .gitignore
echo "Checking .gitignore..."
if grep -q "^\.env$" .gitignore || grep -q "^\.env\*" .gitignore; then
    echo -e "${GREEN}✓ .env is properly gitignored${NC}"
else
    echo -e "${RED}✗ .env is NOT in .gitignore!${NC}"
    FAILED=1
fi

# Check 2: No .env files staged
echo ""
echo "Checking staged files..."
STAGED_ENV=$(git diff --cached --name-only | grep -E "\.env$|\.env\." || true)
if [ -z "$STAGED_ENV" ]; then
    echo -e "${GREEN}✓ No .env files staged${NC}"
else
    echo -e "${RED}✗ .env files staged for commit:${NC}"
    echo "$STAGED_ENV"
    FAILED=1
fi

# Check 3: No real API keys in source files
echo ""
echo "Checking for API key patterns in source..."
# Look for patterns like sk-xxxxxxxx (real keys) but not sk-xxxx (placeholders)
KEY_PATTERNS=$(grep -r "sk-[a-zA-Z0-9]\{20,\}" --include="*.go" --include="*.yaml" --include="*.yml" --include="*.json" . 2>/dev/null | grep -v "\.git" | grep -v "vendor" || true)
if [ -z "$KEY_PATTERNS" ]; then
    echo -e "${GREEN}✓ No real API keys found in source${NC}"
else
    echo -e "${YELLOW}⚠ Potential API key patterns found:${NC}"
    echo "$KEY_PATTERNS" | head -10
    echo ""
    echo "Verify these are not real keys before committing!"
fi

# Check 4: .env.example doesn't have real keys
echo ""
echo "Checking .env.example..."
if grep -q "your-\|example\|placeholder\|xxx\|\.\.\." .env.example || [ ! -s .env.example ]; then
    echo -e "${GREEN}✓ .env.example appears to be a template${NC}"
else
    echo -e "${YELLOW}⚠ .env.example may contain real values - verify${NC}"
fi

echo ""
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}=== Security Check PASSED ===${NC}"
    exit 0
else
    echo -e "${RED}=== Security Check FAILED ===${NC}"
    echo "Fix issues before committing!"
    exit 1
fi
