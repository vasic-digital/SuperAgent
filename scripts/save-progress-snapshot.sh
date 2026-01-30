#!/bin/bash
# Save progress snapshot
# Creates a timestamped snapshot of current progress for safe stop/resume

set -e

TIMESTAMP=$(date +'%Y%m%d_%H%M%S')
SNAPSHOT_DIR="progress_snapshots"
SNAPSHOT_FILE="$SNAPSHOT_DIR/snapshot_$TIMESTAMP.md"

# Create snapshot directory
mkdir -p "$SNAPSHOT_DIR"

# Create snapshot
cat > "$SNAPSHOT_FILE" << 'SNAPSHOT'
# Progress Snapshot - $(date +'%Y-%m-%d %H:%M:%S')

## Git Status
```
$(git status)
```

## Recent Commits (Last 5)
```
$(git log --oneline -5)
```

## Current Branch
```
$(git branch --show-current)
```

## Files Changed (Not Committed)
```
$(git status --short)
```

## Build Status
```
$(go build ./internal/... ./cmd/... 2>&1 && echo "✅ Build successful" || echo "❌ Build failed")
```

## Test Status (Quick Check)
```
$(go test -short ./internal/streaming/... 2>&1 | tail -5)
```

## Submodules Status
```
$(git submodule status | head -10)
```

## Task List Status
$(cat PROGRESS.md | grep -A 50 "Phase Completion Status")

## Next Steps
- Review current progress
- Continue from where we left off
- All work is committed and pushed to upstream

## Resume Instructions
1. Pull latest changes: `git pull origin main`
2. Check PROGRESS.md for current phase
3. Review docs/phase*_completion_summary.md for completed phases
4. Continue with next pending task from task list

SNAPSHOT

echo "✅ Progress snapshot saved: $SNAPSHOT_FILE"
echo ""
echo "📊 Snapshot includes:"
echo "   - Git status and recent commits"
echo "   - Build and test status"
echo "   - Task list progress"
echo "   - Resume instructions"
echo ""
echo "To view: cat $SNAPSHOT_FILE"
