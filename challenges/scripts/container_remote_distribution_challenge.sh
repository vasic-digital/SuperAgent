#!/bin/bash
# Container Remote Distribution Challenge
# Validates that remote distribution packages exist, interfaces
# are correct, tests pass, and .env.example is documented.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CONTAINERS_DIR="${PROJECT_ROOT}/Containers"

PASSED=0
FAILED=0
TOTAL=0

pass() {
    PASSED=$((PASSED + 1))
    TOTAL=$((TOTAL + 1))
    echo "  PASS: $1"
}

fail() {
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
    echo "  FAIL: $1"
}

echo "=============================================="
echo " Container Remote Distribution Challenge"
echo "=============================================="
echo ""

# ---- Package Existence ----

echo "--- Package Existence ---"

for pkg in remote scheduler network volume envconfig distribution; do
    if [ -d "${CONTAINERS_DIR}/pkg/${pkg}" ]; then
        pass "pkg/${pkg}/ exists"
    else
        fail "pkg/${pkg}/ not found"
    fi
done

# ---- Key Interface Files ----

echo "--- Key Interface Files ---"

check_file() {
    if [ -f "${CONTAINERS_DIR}/$1" ]; then
        pass "$1 exists"
    else
        fail "$1 not found"
    fi
}

check_file "pkg/remote/types.go"
check_file "pkg/remote/executor.go"
check_file "pkg/remote/ssh_executor.go"
check_file "pkg/remote/connection_pool.go"
check_file "pkg/remote/host_manager.go"
check_file "pkg/remote/runtime.go"
check_file "pkg/remote/probe.go"
check_file "pkg/remote/compose.go"
check_file "pkg/remote/options.go"
check_file "pkg/scheduler/scheduler.go"
check_file "pkg/scheduler/scorer.go"
check_file "pkg/scheduler/strategies.go"
check_file "pkg/scheduler/types.go"
check_file "pkg/network/tunnel.go"
check_file "pkg/network/port_allocator.go"
check_file "pkg/network/overlay.go"
check_file "pkg/volume/manager.go"
check_file "pkg/volume/sshfs.go"
check_file "pkg/volume/nfs.go"
check_file "pkg/volume/rsync.go"
check_file "pkg/envconfig/parser.go"
check_file "pkg/envconfig/dotenv.go"
check_file "pkg/envconfig/template.go"
check_file "pkg/distribution/distributor.go"
check_file "pkg/distribution/workflow.go"
check_file "pkg/distribution/failover.go"
check_file "pkg/distribution/options.go"
check_file "pkg/distribution/types.go"

# ---- Key Interfaces ----

echo "--- Key Interfaces ---"

check_interface() {
    if grep -q "$2" "${CONTAINERS_DIR}/$1" 2>/dev/null; then
        pass "$2 in $1"
    else
        fail "$2 not found in $1"
    fi
}

check_interface "pkg/remote/executor.go" "RemoteExecutor interface"
check_interface "pkg/remote/host_manager.go" "HostManager interface"
check_interface "pkg/scheduler/scheduler.go" "Scheduler interface"
check_interface "pkg/network/tunnel.go" "TunnelManager interface"
check_interface "pkg/volume/manager.go" "VolumeManager interface"
check_interface "pkg/distribution/distributor.go" "Distributor interface"

# ---- Event Types ----

echo "--- Event Types ---"

for event in "remote.host.online" "remote.host.offline" "remote.host.degraded" \
    "distribution.scheduled" "distribution.deployed" "distribution.migrated" \
    "tunnel.created" "tunnel.closed" "volume.mounted" "volume.unmounted"; do
    if grep -q "${event}" "${CONTAINERS_DIR}/pkg/event/events.go"; then
        pass "Event ${event} defined"
    else
        fail "Event ${event} not found"
    fi
done

# ---- .env.example ----

echo "--- .env.example ---"

if [ -f "${CONTAINERS_DIR}/.env.example" ]; then
    pass ".env.example exists"
else
    fail ".env.example not found"
fi

if grep -q "CONTAINERS_REMOTE_ENABLED" "${CONTAINERS_DIR}/.env.example" 2>/dev/null; then
    pass ".env.example has CONTAINERS_REMOTE_ENABLED"
else
    fail ".env.example missing CONTAINERS_REMOTE_ENABLED"
fi

if grep -q "CONTAINERS_REMOTE_HOST_1" "${CONTAINERS_DIR}/.env.example" 2>/dev/null; then
    pass ".env.example has host definitions"
else
    fail ".env.example missing host definitions"
fi

# ---- .gitignore ----

echo "--- .gitignore ---"

if grep -q "^\.env$" "${CONTAINERS_DIR}/.gitignore"; then
    pass ".env in .gitignore"
else
    fail ".env not in .gitignore"
fi

# ---- Documentation ----

echo "--- Documentation ---"

if [ -f "${CONTAINERS_DIR}/docs/REMOTE_DISTRIBUTION.md" ]; then
    pass "REMOTE_DISTRIBUTION.md exists"
else
    fail "REMOTE_DISTRIBUTION.md not found"
fi

if grep -q "remote" "${CONTAINERS_DIR}/CLAUDE.md" && grep -q "distribution" "${CONTAINERS_DIR}/CLAUDE.md"; then
    pass "CLAUDE.md references remote distribution"
else
    fail "CLAUDE.md missing remote distribution"
fi

if grep -q "remote" "${CONTAINERS_DIR}/AGENTS.md" && grep -q "distribution" "${CONTAINERS_DIR}/AGENTS.md"; then
    pass "AGENTS.md references remote distribution"
else
    fail "AGENTS.md missing remote distribution"
fi

if grep -q "Remote distribution" "${CONTAINERS_DIR}/README.md" || grep -q "remote" "${CONTAINERS_DIR}/README.md"; then
    pass "README.md references remote distribution"
else
    fail "README.md missing remote distribution"
fi

# ---- Cluster Snapshot ----

echo "--- Cluster Snapshot ---"

if grep -q "ClusterSnapshot" "${CONTAINERS_DIR}/pkg/monitor/types.go"; then
    pass "ClusterSnapshot type exists"
else
    fail "ClusterSnapshot type not found"
fi

# ---- Boot Manager Distributor Integration ----

echo "--- Boot Manager Integration ---"

if grep -q "WithDistributor" "${CONTAINERS_DIR}/pkg/boot/options.go"; then
    pass "WithDistributor option exists"
else
    fail "WithDistributor option not found"
fi

if grep -q "distributor" "${CONTAINERS_DIR}/pkg/boot/manager.go"; then
    pass "BootManager has distributor integration"
else
    fail "BootManager missing distributor integration"
fi

# ---- Tests Pass ----

echo "--- Tests Pass ---"

if (cd "${CONTAINERS_DIR}" && GOMAXPROCS=2 go test ./... -race -count=1 -timeout=120s -p 1 >/dev/null 2>&1); then
    pass "All Containers module tests pass"
else
    fail "Some Containers module tests failed"
fi

# ---- Build ----

echo "--- Build ---"

if (cd "${CONTAINERS_DIR}" && go build ./... >/dev/null 2>&1); then
    pass "Containers module builds clean"
else
    fail "Containers module build failed"
fi

echo ""
echo "=============================================="
echo " Remote Distribution Challenge Results"
echo "=============================================="
echo " Passed: ${PASSED}/${TOTAL}"
echo " Failed: ${FAILED}/${TOTAL}"
echo "=============================================="

if [ "${FAILED}" -gt 0 ]; then
    exit 1
fi
exit 0
