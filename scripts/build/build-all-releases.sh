#!/bin/bash
# HelixAgent Build All Releases
# Builds all registered apps for all platforms.
#
# Usage:
#   ./scripts/build/build-all-releases.sh [--force]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Pass through flags
EXTRA_FLAGS=""
if [[ "$*" == *"--force"* ]]; then
    EXTRA_FLAGS="--force"
fi

# Source version manager for app list
source "$SCRIPT_DIR/version-manager.sh"

echo -e "${BLUE}=== HelixAgent: Build All Releases ===${NC}"
echo ""

APPS=($(list_all_apps))
TOTAL_APPS=${#APPS[@]}
SUCCEEDED=0
SKIPPED=0
FAILED=0

for app in "${APPS[@]}"; do
    echo -e "${BLUE}--- Building: $app ---${NC}"
    if "$SCRIPT_DIR/build-release.sh" --app "$app" --all-platforms $EXTRA_FLAGS; then
        SUCCEEDED=$((SUCCEEDED + 1))
    else
        # Exit code 0 with "no changes" message means skipped
        FAILED=$((FAILED + 1))
    fi
    echo ""
done

echo -e "${BLUE}=== All Releases Summary ===${NC}"
echo -e "  Total apps: $TOTAL_APPS"
echo -e "  ${GREEN}Succeeded:${NC} $SUCCEEDED"
if [ "$FAILED" -gt 0 ]; then
    echo -e "  ${RED}Failed:${NC}    $FAILED"
fi
echo ""

if [ "$FAILED" -gt 0 ]; then
    echo -e "${RED}Some builds failed.${NC}"
    exit 1
fi

echo -e "${GREEN}All release builds complete.${NC}"
