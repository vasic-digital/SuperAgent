#!/bin/bash
# SuperAgent Challenges - Results Cleanup Script
# Usage: ./scripts/cleanup_results.sh [days]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
RESULTS_DIR="$CHALLENGES_DIR/results"
MASTER_DIR="$CHALLENGES_DIR/master_results"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

# Default: clean results older than 30 days
DAYS=${1:-30}
DRY_RUN=false

# Check for --dry-run flag
for arg in "$@"; do
    if [ "$arg" = "--dry-run" ]; then
        DRY_RUN=true
    fi
done

print_info "SuperAgent Challenges - Cleanup"
print_info "Removing results older than $DAYS days"

if [ "$DRY_RUN" = true ]; then
    print_warning "DRY RUN - No files will be deleted"
fi

# Count files to delete
TOTAL_REMOVED=0
TOTAL_SIZE=0

cleanup_old_results() {
    local dir=$1
    local name=$2

    if [ ! -d "$dir" ]; then
        return
    fi

    print_info "Scanning $name..."

    local old_dirs=$(find "$dir" -maxdepth 5 -type d -name "[0-9]*_[0-9]*" -mtime +$DAYS 2>/dev/null)

    for old_dir in $old_dirs; do
        local size=$(du -sh "$old_dir" 2>/dev/null | cut -f1)
        print_info "  Found: $old_dir ($size)"

        if [ "$DRY_RUN" = false ]; then
            rm -rf "$old_dir"
            TOTAL_REMOVED=$((TOTAL_REMOVED + 1))
        fi
    done
}

# Clean challenge results
cleanup_old_results "$RESULTS_DIR" "challenge results"

# Clean master summaries (keep last 10)
print_info "Cleaning old master summaries (keeping last 10)..."
if [ -d "$MASTER_DIR" ]; then
    local summaries=$(ls -1t "$MASTER_DIR"/master_summary_*.md 2>/dev/null | tail -n +11)
    for summary in $summaries; do
        print_info "  Removing: $summary"
        if [ "$DRY_RUN" = false ]; then
            rm -f "$summary"
            TOTAL_REMOVED=$((TOTAL_REMOVED + 1))
        fi
    done
fi

# Clean empty directories
print_info "Removing empty directories..."
if [ "$DRY_RUN" = false ]; then
    find "$RESULTS_DIR" -type d -empty -delete 2>/dev/null || true
fi

echo ""
print_success "Cleanup complete!"
print_info "Items removed: $TOTAL_REMOVED"

if [ "$DRY_RUN" = true ]; then
    echo ""
    print_warning "This was a dry run. Run without --dry-run to delete files."
fi
