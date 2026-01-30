#!/bin/bash
# Auto-commit and push progress
# Usage: ./scripts/auto-commit-progress.sh "Phase X work" [--force]

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
FORCE_PUSH=false
if [[ "$2" == "--force" ]]; then
    FORCE_PUSH=true
fi

MESSAGE="${1:-Auto-commit progress update}"

echo -e "${YELLOW}📝 HelixAgent Auto-Commit Script${NC}"
echo "======================================"
echo ""

# 1. Update PROGRESS.md with current timestamp
echo -e "${GREEN}✓${NC} Updating PROGRESS.md..."
sed -i "s/\*\*Last Updated\*\*:.*/\*\*Last Updated\*\*: $(date +'%Y-%m-%d %H:%M:%S') (Auto-updated on each commit)/" PROGRESS.md

# 2. Check for changes
if git diff --quiet && git diff --cached --quiet; then
    echo -e "${YELLOW}⚠${NC} No changes to commit"
    exit 0
fi

# 3. Stage all changes
echo -e "${GREEN}✓${NC} Staging changes..."
git add -A

# 4. Show status
echo ""
echo -e "${YELLOW}Changes to commit:${NC}"
git status --short | head -20

# 5. Commit
echo ""
echo -e "${GREEN}✓${NC} Creating commit..."
git commit -m "$MESSAGE

Auto-commit by scripts/auto-commit-progress.sh
Timestamp: $(date +'%Y-%m-%d %H:%M:%S')

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

# 6. Get commit hash
COMMIT_HASH=$(git rev-parse --short HEAD)
echo -e "${GREEN}✓${NC} Commit created: ${COMMIT_HASH}"

# 7. Push to remotes
echo ""
echo -e "${GREEN}✓${NC} Pushing to remotes..."

# Push to origin
if git push origin main 2>&1; then
    echo -e "${GREEN}  ✓${NC} Pushed to origin (github.com:vasic-digital/SuperAgent.git)"
else
    echo -e "${RED}  ✗${NC} Failed to push to origin"
fi

# Push to githubhelixdevelopment
if git push githubhelixdevelopment main 2>&1; then
    echo -e "${GREEN}  ✓${NC} Pushed to githubhelixdevelopment (github.com:HelixDevelopment/HelixAgent.git)"
else
    echo -e "${YELLOW}  ⚠${NC} githubhelixdevelopment not accessible or already up-to-date"
fi

# 8. Check submodules
echo ""
echo -e "${YELLOW}📦 Checking submodules...${NC}"
SUBMODULE_CHANGES=$(git submodule foreach --quiet 'git status --porcelain | grep -q . && echo "1" || echo "0"' | grep -c "1" || true)

if [ "$SUBMODULE_CHANGES" -gt 0 ]; then
    echo -e "${YELLOW}  ⚠${NC} $SUBMODULE_CHANGES submodule(s) have uncommitted changes"
    echo "  Run: git submodule foreach 'git status' to see details"
else
    echo -e "${GREEN}  ✓${NC} All submodules clean"
fi

echo ""
echo -e "${GREEN}✅ Auto-commit complete!${NC}"
echo "   Commit: $COMMIT_HASH"
echo "   Message: $MESSAGE"
echo ""
