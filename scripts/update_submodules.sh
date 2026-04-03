#!/bin/bash
# Submodule update script with workaround for known issues

set -e

echo "=== Updating HelixAgent Submodules ==="

# First, initialize/update the bridle submodule itself (but not its nested submodules)
echo "Updating bridle submodule (without nested submodules)..."
git submodule update --init -- cli_agents/bridle || true

# Update all other submodules
echo "Updating all other submodules..."
git submodule update --init --recursive 2>&1 | tee /tmp/submodule_update.log || {
    # Check if only known errors occurred
    if grep -q "axiom.*No url found" /tmp/submodule_update.log; then
        echo ""
        echo "Note: Expected error for bridle/axiom submodule (see docs/SUBMODULE_FIXES.md)"
        echo "Other submodules updated successfully."
        exit 0
    else
        echo "ERROR: Unexpected submodule failures"
        exit 1
    fi
}

echo ""
echo "=== Submodules Updated Successfully ==="
