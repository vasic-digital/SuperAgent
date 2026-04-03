#!/bin/bash
# Add remaining CLI agent submodules
# Run this after current git operations complete

cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent

echo "=== Removing lock files ==="
rm -f .git/index.lock
rm -f .git/modules/*/index.lock 2>/dev/null

echo "=== Adding remaining submodules ==="

# 1. crush - Charm's CLI agent
echo "1. Adding crush..."
rm -rf cli_agents/crush .git/modules/cli_agents/crush 2>/dev/null
git rm --cached cli_agents/crush 2>/dev/null
git submodule add --force https://github.com/charmbracelet/crush.git cli_agents/crush

# 2. x-cmd - Modular toolkit
echo "2. Adding x-cmd..."
git submodule add https://github.com/x-cmd/x-cmd.git cli_agents/x-cmd

# 3. pi - Minimal coding harness
echo "3. Adding pi..."
git submodule add https://github.com/pi-mono/pi.git cli_agents/pi

# 4. roo-code - VS Code + CLI
echo "4. Adding roo-code..."
git submodule add https://github.com/RooVetGit/Roo-Code.git cli_agents/roo-code

# 5. continue - IDE + CLI
echo "5. Adding continue..."
git submodule add https://github.com/continuedev/continue.git cli_agents/continue

# 6. open-interpreter - General purpose
echo "6. Adding open-interpreter..."
git submodule add https://github.com/OpenInterpreter/open-interpreter.git cli_agents/open-interpreter

# 7. swe-agent - Academic/research
echo "7. Adding swe-agent..."
git submodule add https://github.com/SWE-agent/SWE-agent.git cli_agents/swe-agent

echo "=== Initializing all new submodules ==="
git submodule update --init --recursive

echo "=== Final status ==="
echo "Total CLI agent submodules:"
git submodule status | grep "cli_agents/" | wc -l

echo ""
echo "=== Committing changes ==="
git add .gitmodules
git commit -m "Add remaining CLI agent submodules: crush, x-cmd, pi, roo-code, continue, open-interpreter, swe-agent"

echo "Done!"
