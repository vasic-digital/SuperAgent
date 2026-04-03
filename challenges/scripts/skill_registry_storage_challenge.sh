#!/bin/bash

# SkillRegistry Storage Challenge Script
# Validates memory and PostgreSQL storage implementations

set -e

echo "=========================================="
echo "SkillRegistry Storage Challenge"
echo "=========================================="
echo ""

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

# Check if files exist
echo "📁 Checking Storage Files..."
echo "──────────────────────────────────────────"

FILES=(
    "SkillRegistry/storage.go"
    "SkillRegistry/storage_memory.go"
    "SkillRegistry/storage_postgres.go"
    "SkillRegistry/storage_test.go"
)

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        LINES=$(wc -l < "$file")
        echo "  ✅ $file ($LINES lines)"
    else
        echo "  ❌ Missing: $file"
        exit 1
    fi
done

echo ""
echo "🧪 Running Storage Tests..."
echo "──────────────────────────────────────────"

# Run tests with resource limits
cd SkillRegistry
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -v -run "TestMemoryStorage|TestDefaultStorageConfig|TestGenerateSkillID" -timeout 60s . 2>&1 | head -50

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo ""
    echo "  ✅ All storage tests passed"
else
    echo ""
    echo "  ⚠️ Some tests failed (check output above)"
fi

cd "$REPO_ROOT"

echo ""
echo "🔍 Code Analysis..."
echo "──────────────────────────────────────────"

# Check for key interfaces
echo "  Checking SkillStorage interface..."
if grep -q "type SkillStorage interface" SkillRegistry/storage.go; then
    echo "    ✅ SkillStorage interface defined"
else
    echo "    ❌ SkillStorage interface missing"
fi

# Check MemoryStorage methods
MEMORY_METHODS=("Save" "Load" "LoadByName" "Delete" "List" "ListByCategory" "Search" "Update" "Close" "HealthCheck")
for method in "${MEMORY_METHODS[@]}"; do
    if grep -q "func (s \*MemoryStorage) $method" SkillRegistry/storage_memory.go; then
        echo "    ✅ MemoryStorage.$method"
    else
        echo "    ❌ MemoryStorage.$method missing"
    fi
done

# Check PostgresStorage methods  
POSTGRES_METHODS=("Save" "Load" "LoadByName" "Delete" "List" "ListByCategory" "Search" "Update" "Close" "HealthCheck" "InitSchema")
for method in "${POSTGRES_METHODS[@]}"; do
    if grep -q "func (s \*PostgresStorage) $method" SkillRegistry/storage_postgres.go; then
        echo "    ✅ PostgresStorage.$method"
    else
        echo "    ❌ PostgresStorage.$method missing"
    fi
done

echo ""
echo "📊 Code Statistics..."
echo "──────────────────────────────────────────"

SKILLREGISTRY_TOTAL=$(find SkillRegistry -name "*.go" | xargs wc -l | tail -1 | awk '{print $1}')
echo "  Total lines in SkillRegistry: $SKILLREGISTRY_TOTAL"

STORAGE_LINES=$(cat SkillRegistry/storage.go SkillRegistry/storage_memory.go SkillRegistry/storage_postgres.go SkillRegistry/storage_test.go 2>/dev/null | wc -l)
echo "  Storage implementation lines: $STORAGE_LINES"

TEST_FILES=$(ls SkillRegistry/*_test.go 2>/dev/null | wc -l)
echo "  Test files: $TEST_FILES"

echo ""
echo "=========================================="
echo "Challenge Complete"
echo "=========================================="
