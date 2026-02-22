#!/bin/bash
# ctop Challenge Script
# Validates the ctop (container top) feature works with local and remote containers

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CONTAINERS_DIR="${PROJECT_ROOT}/Containers"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

pass() {
    echo -e "${GREEN}✓${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

skip() {
    echo -e "${YELLOW}⊘${NC} $1 (skipped)"
    TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
}

section() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN} $1 ${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
}

# ─────────────────────────────────────────────────────────────────────────────
# Section 1: Package Structure
# ─────────────────────────────────────────────────────────────────────────────

section "Package Structure"

TEST_NAME="ctop types.go exists"
if [[ -f "${CONTAINERS_DIR}/pkg/ctop/types.go" ]]; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="ctop collector.go exists"
if [[ -f "${CONTAINERS_DIR}/pkg/ctop/collector.go" ]]; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="ctop display.go exists"
if [[ -f "${CONTAINERS_DIR}/pkg/ctop/display.go" ]]; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="ctop tests exist"
if [[ -f "${CONTAINERS_DIR}/pkg/ctop/ctop_test.go" ]]; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

# ─────────────────────────────────────────────────────────────────────────────
# Section 2: Type Definitions
# ─────────────────────────────────────────────────────────────────────────────

section "Type Definitions"

TEST_NAME="ContainerProcess type defined"
if grep -q "type ContainerProcess struct" "${CONTAINERS_DIR}/pkg/ctop/types.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="ContainerProcessList type defined"
if grep -q "type ContainerProcessList struct" "${CONTAINERS_DIR}/pkg/ctop/types.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="SortField constants defined"
if grep -q "SortByCPU" "${CONTAINERS_DIR}/pkg/ctop/types.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="DisplayConfig type defined"
if grep -q "type DisplayConfig struct" "${CONTAINERS_DIR}/pkg/ctop/types.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="CollectorStats type defined"
if grep -q "type CollectorStats struct" "${CONTAINERS_DIR}/pkg/ctop/types.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

# ─────────────────────────────────────────────────────────────────────────────
# Section 3: Collector Functions
# ─────────────────────────────────────────────────────────────────────────────

section "Collector Functions"

TEST_NAME="NewCollector function defined"
if grep -q "func NewCollector" "${CONTAINERS_DIR}/pkg/ctop/collector.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="Collect method defined"
if grep -q "func (c \*Collector) Collect" "${CONTAINERS_DIR}/pkg/ctop/collector.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="collectLocal method defined"
if grep -q "func (c \*Collector) collectLocal" "${CONTAINERS_DIR}/pkg/ctop/collector.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="collectRemote method defined"
if grep -q "func (c \*Collector) collectRemote" "${CONTAINERS_DIR}/pkg/ctop/collector.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="GetStats method defined"
if grep -q "func (c \*Collector) GetStats" "${CONTAINERS_DIR}/pkg/ctop/collector.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

# ─────────────────────────────────────────────────────────────────────────────
# Section 4: Display Functions
# ─────────────────────────────────────────────────────────────────────────────

section "Display Functions"

TEST_NAME="NewDisplay function defined"
if grep -q "func NewDisplay" "${CONTAINERS_DIR}/pkg/ctop/display.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="Display.Run method defined"
if grep -q "func (d \*Display) Run" "${CONTAINERS_DIR}/pkg/ctop/display.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="RenderSnapshot method defined"
if grep -q "func (d \*Display) RenderSnapshot" "${CONTAINERS_DIR}/pkg/ctop/display.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="RenderJSON method defined"
if grep -q "func (d \*Display) RenderJSON" "${CONTAINERS_DIR}/pkg/ctop/display.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="Sort methods (SetSortBy, ToggleSortOrder) defined"
if grep -q "func (d \*Display) SetSortBy" "${CONTAINERS_DIR}/pkg/ctop/display.go" && \
   grep -q "func (d \*Display) ToggleSortOrder" "${CONTAINERS_DIR}/pkg/ctop/display.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

# ─────────────────────────────────────────────────────────────────────────────
# Section 5: ContainerProcess Fields
# ─────────────────────────────────────────────────────────────────────────────

section "ContainerProcess Fields"

REQUIRED_FIELDS=("ID" "Name" "Image" "Runtime" "Host" "Location" "State" "Status"
                  "CPUPercent" "MemoryUsage" "MemoryLimit" "MemoryPercent"
                  "NetworkRx" "NetworkTx" "BlockRead" "BlockWrite" "PIDs" "Ports")

for field in "${REQUIRED_FIELDS[@]}"; do
    TEST_NAME="ContainerProcess has $field field"
    if grep -q "$field" "${CONTAINERS_DIR}/pkg/ctop/types.go" 2>/dev/null; then
        pass "$TEST_NAME"
    else
        fail "$TEST_NAME"
    fi
done

# ─────────────────────────────────────────────────────────────────────────────
# Section 6: List Operations
# ─────────────────────────────────────────────────────────────────────────────

section "List Operations"

TEST_NAME="ContainerProcessList.Sort method defined"
if grep -q "func (list \*ContainerProcessList) Sort" "${CONTAINERS_DIR}/pkg/ctop/collector.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="ContainerProcessList.Filter method defined"
if grep -q "func (list \*ContainerProcessList) Filter" "${CONTAINERS_DIR}/pkg/ctop/collector.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

# ─────────────────────────────────────────────────────────────────────────────
# Section 7: Tests Pass
# ─────────────────────────────────────────────────────────────────────────────

section "Unit Tests"

TEST_NAME="ctop unit tests pass"
cd "${CONTAINERS_DIR}"
if go test -v ./pkg/ctop/... > /tmp/ctop_test_output.txt 2>&1; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
    echo "  Test output:"
    tail -20 /tmp/ctop_test_output.txt | sed 's/^/    /'
fi

# ─────────────────────────────────────────────────────────────────────────────
# Section 8: Integration Check
# ─────────────────────────────────────────────────────────────────────────────

section "Integration Check"

TEST_NAME="ctop package compiles"
cd "${CONTAINERS_DIR}"
if go build ./pkg/ctop/... 2>/dev/null; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="ctop imports from remote package"
if grep -q '"digital.vasic.containers/pkg/remote"' "${CONTAINERS_DIR}/pkg/ctop/collector.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

TEST_NAME="Collector supports HostManager interface"
if grep -q "hostManager remote.HostManager" "${CONTAINERS_DIR}/pkg/ctop/collector.go"; then
    pass "$TEST_NAME"
else
    fail "$TEST_NAME"
fi

# ─────────────────────────────────────────────────────────────────────────────
# Summary
# ─────────────────────────────────────────────────────────────────────────────

echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN} Summary ${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "  ${GREEN}Passed:${NC}  ${TESTS_PASSED}"
echo -e "  ${RED}Failed:${NC}  ${TESTS_FAILED}"
echo -e "  ${YELLOW}Skipped:${NC} ${TESTS_SKIPPED}"
echo ""

if [[ ${TESTS_FAILED} -eq 0 ]]; then
    echo -e "${GREEN}All ctop challenges passed! ✓${NC}"
    exit 0
else
    echo -e "${RED}Some ctop challenges failed! ✗${NC}"
    exit 1
fi
