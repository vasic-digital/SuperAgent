#!/bin/bash
set -e  # Exit on error

set -euo pipefail

# List of submodules (directories with .git files)
module_dirs=(
    "Auth"
    "BackgroundTasks"
    "Benchmark"
    "Cache"
    "Challenges"
    "Concurrency"
    "Containers"
    "Database"
    "Embeddings"
    "EventBus"
    "Formatters"
    "LLMOps"
    "MCP_Module"
    "Memory"
    "Messaging"
    "Models"
    "Observability"
    "Optimization"
    "Plugins"
    "Planning"
    "RAG"
    "Security"
    "SelfImprove"
    "Storage"
    "Streaming"
    "VectorDB"
    "Agentic"
    "HelixMemory"
    "HelixSpecifier"
)

echo "=== Processing Submodules ==="

# Process each submodule
for module in "${module_dirs[@]}"; do
    if [ -d "$module" ]; then
        echo ""
        echo "Processing $module..."
        cd "$module"
        
        # Check if it's a git repository
        if [ ! -d ".git" ]; then
            echo "  - Not a git repository, skipping"
            cd ..
            continue
        fi
        
        # Get current branch
        branch=$(git rev-parse --abbrev-ref HEAD)
        
        # Check for changes
        has_changes=$(git status --porcelain)
        
        if [ -z "$has_changes" ]; then
            echo "  - No changes to $module"
            cd ..
            continue
        fi
        
        echo "  - Changes detected in $module"
        
        # Add all changes
        git add -A .
        
        # Commit changes
        if git diff --cached --quiet; then
            git commit -m "chore($module): update from HelixAgent integration" --allow-empty
        else
            # If nothing to commit, create empty commit
            git commit --allow-empty -m "chore($module): no changes" --allow-empty
        fi
        
        # Push to all remotes
        remotes=$(git remote | awk '{print $1}')
        for remote in $remotes; do
            echo "  - Pushing to $remote..."
            if git push "$remote" "$branch" 2>&1; then
                echo "    ✓ Pushed to $remote"
            else
                echo "    ✗ Failed to push to $remote"
            fi
        done
        
        cd ..
    else
        echo "  - Directory does not exist, skipping"
    fi
done

