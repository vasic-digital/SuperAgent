#!/bin/bash
# HelixAgent Challenge: Release Build System
# Tests: ~25 tests across 6 sections
# Validates: Version package, version management, build scripts,
#            container build, release structure, functional validation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

pass() {
    PASSED=$((PASSED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${GREEN}[PASS]${NC} $1"
}

fail() {
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${RED}[FAIL]${NC} $1"
}

section() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

#===============================================================================
# Section 1: Version Package (5 tests)
#===============================================================================
section "Section 1: Version Package"

# Test 1.1: version.go exists with required variables
if [ -f "$PROJECT_ROOT/internal/version/version.go" ] && \
   grep -q 'Version.*=.*"dev"' "$PROJECT_ROOT/internal/version/version.go" && \
   grep -q 'VersionCode.*=.*"0"' "$PROJECT_ROOT/internal/version/version.go" && \
   grep -q 'GitCommit.*=.*"unknown"' "$PROJECT_ROOT/internal/version/version.go" && \
   grep -q 'BuildDate.*=.*"unknown"' "$PROJECT_ROOT/internal/version/version.go" && \
   grep -q 'SourceHash.*=.*"unknown"' "$PROJECT_ROOT/internal/version/version.go"; then
    pass "Version variables with ldflags defaults"
else
    fail "Missing version variables or defaults"
fi

# Test 1.2: Info struct with JSON tags
if grep -q 'type Info struct' "$PROJECT_ROOT/internal/version/version.go" && \
   grep -q 'json:"version"' "$PROJECT_ROOT/internal/version/version.go" && \
   grep -q 'json:"version_code"' "$PROJECT_ROOT/internal/version/version.go" && \
   grep -q 'json:"git_commit"' "$PROJECT_ROOT/internal/version/version.go" && \
   grep -q 'json:"platform"' "$PROJECT_ROOT/internal/version/version.go"; then
    pass "Info struct with JSON tags"
else
    fail "Info struct incomplete or missing JSON tags"
fi

# Test 1.3: Get() function
if grep -q 'func Get() Info' "$PROJECT_ROOT/internal/version/version.go"; then
    pass "Get() function exists"
else
    fail "Get() function missing"
fi

# Test 1.4: Short() function
if grep -q 'func Short() string' "$PROJECT_ROOT/internal/version/version.go"; then
    pass "Short() function exists"
else
    fail "Short() function missing"
fi

# Test 1.5: Version package compiles and tests pass
if (cd "$PROJECT_ROOT" && go test ./internal/version/... >/dev/null 2>&1); then
    pass "Version package tests pass"
else
    fail "Version package tests fail"
fi

#===============================================================================
# Section 2: Version Management (5 tests)
#===============================================================================
section "Section 2: Version Management"

# Test 2.1: VERSION file exists with semver format
if [ -f "$PROJECT_ROOT/VERSION" ]; then
    VER=$(tr -d '[:space:]' < "$PROJECT_ROOT/VERSION")
    if echo "$VER" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
        pass "VERSION file with semver format ($VER)"
    else
        fail "VERSION file not semver format: $VER"
    fi
else
    fail "VERSION file missing"
fi

# Test 2.2: version-manager.sh exists and is executable
if [ -x "$PROJECT_ROOT/scripts/build/version-manager.sh" ]; then
    pass "version-manager.sh executable"
else
    fail "version-manager.sh missing or not executable"
fi

# Test 2.3: App registry has all 7 apps
if source "$PROJECT_ROOT/scripts/build/version-manager.sh" 2>/dev/null; then
    APP_COUNT=$(list_all_apps | wc -l)
    if [ "$APP_COUNT" -eq 7 ]; then
        pass "App registry has 7 apps"
    else
        fail "App registry has $APP_COUNT apps (expected 7)"
    fi
else
    fail "Cannot source version-manager.sh"
fi

# Test 2.4: Hash computation is deterministic
HASH1=$(source "$PROJECT_ROOT/scripts/build/version-manager.sh" 2>/dev/null && compute_source_hash helixagent)
HASH2=$(source "$PROJECT_ROOT/scripts/build/version-manager.sh" 2>/dev/null && compute_source_hash helixagent)
if [ -n "$HASH1" ] && [ "$HASH1" = "$HASH2" ]; then
    pass "Hash computation is deterministic"
else
    fail "Hash computation not deterministic ($HASH1 != $HASH2)"
fi

# Test 2.5: version-data directory exists
if [ -d "$PROJECT_ROOT/releases/.version-data" ]; then
    pass "releases/.version-data directory exists"
else
    fail "releases/.version-data directory missing"
fi

#===============================================================================
# Section 3: Build Scripts (5 tests)
#===============================================================================
section "Section 3: Build Scripts"

# Test 3.1: build-container.sh exists and is executable
if [ -x "$PROJECT_ROOT/scripts/build/build-container.sh" ]; then
    pass "build-container.sh executable"
else
    fail "build-container.sh missing or not executable"
fi

# Test 3.2: build-release.sh exists and is executable
if [ -x "$PROJECT_ROOT/scripts/build/build-release.sh" ]; then
    pass "build-release.sh executable"
else
    fail "build-release.sh missing or not executable"
fi

# Test 3.3: build-all-releases.sh exists and is executable
if [ -x "$PROJECT_ROOT/scripts/build/build-all-releases.sh" ]; then
    pass "build-all-releases.sh executable"
else
    fail "build-all-releases.sh missing or not executable"
fi

# Test 3.4: build-release.sh validates unknown apps
OUTPUT=$(bash "$PROJECT_ROOT/scripts/build/build-release.sh" --app nonexistent 2>&1 || true)
if echo "$OUTPUT" | grep -qi "unknown\|error\|valid"; then
    pass "build-release.sh rejects unknown app names"
else
    fail "build-release.sh does not reject unknown apps"
fi

# Test 3.5: build-release.sh requires --app flag
OUTPUT=$(bash "$PROJECT_ROOT/scripts/build/build-release.sh" 2>&1 || true)
if echo "$OUTPUT" | grep -qi "required\|usage\|app"; then
    pass "build-release.sh shows usage without --app"
else
    fail "build-release.sh does not show usage"
fi

#===============================================================================
# Section 4: Container Build (3 tests)
#===============================================================================
section "Section 4: Container Build"

# Test 4.1: Dockerfile.builder exists
if [ -f "$PROJECT_ROOT/docker/build/Dockerfile.builder" ]; then
    pass "Dockerfile.builder exists"
else
    fail "Dockerfile.builder missing"
fi

# Test 4.2: Dockerfile.builder uses golang base image
if grep -q 'golang:1.24' "$PROJECT_ROOT/docker/build/Dockerfile.builder"; then
    pass "Dockerfile.builder uses golang:1.24 base"
else
    fail "Dockerfile.builder wrong base image"
fi

# Test 4.3: Root Dockerfile has version build args
if grep -q 'ARG BUILD_VERSION' "$PROJECT_ROOT/Dockerfile" && \
   grep -q 'ARG BUILD_COMMIT' "$PROJECT_ROOT/Dockerfile" && \
   grep -q 'ARG BUILD_DATE' "$PROJECT_ROOT/Dockerfile" && \
   grep -q 'internal/version' "$PROJECT_ROOT/Dockerfile"; then
    pass "Root Dockerfile has version build args and ldflags"
else
    fail "Root Dockerfile missing version build args"
fi

#===============================================================================
# Section 5: Release Structure (4 tests)
#===============================================================================
section "Section 5: Release Structure"

# Test 5.1: .gitignore has release binary directories
if grep -q 'releases/helixagent/' "$PROJECT_ROOT/.gitignore" && \
   grep -q 'releases/api/' "$PROJECT_ROOT/.gitignore" && \
   grep -q 'releases/grpc-server/' "$PROJECT_ROOT/.gitignore"; then
    pass ".gitignore excludes release binary directories"
else
    fail ".gitignore missing release binary exclusions"
fi

# Test 5.2: Makefile has release targets
if grep -q 'release:' "$PROJECT_ROOT/Makefile" && \
   grep -q 'release-all:' "$PROJECT_ROOT/Makefile" && \
   grep -q 'release-info:' "$PROJECT_ROOT/Makefile" && \
   grep -q 'release-clean:' "$PROJECT_ROOT/Makefile"; then
    pass "Makefile has release targets"
else
    fail "Makefile missing release targets"
fi

# Test 5.3: Makefile has per-app release targets
if grep -q 'release-helixagent:' "$PROJECT_ROOT/Makefile" && \
   grep -q 'release-api:' "$PROJECT_ROOT/Makefile" && \
   grep -q 'release-grpc-server:' "$PROJECT_ROOT/Makefile"; then
    pass "Makefile has per-app release targets"
else
    fail "Makefile missing per-app release targets"
fi

# Test 5.4: Platforms list includes all 5 targets
if source "$PROJECT_ROOT/scripts/build/version-manager.sh" 2>/dev/null; then
    PLAT_COUNT=${#PLATFORMS[@]}
    if [ "$PLAT_COUNT" -eq 5 ]; then
        pass "5 target platforms defined"
    else
        fail "$PLAT_COUNT platforms defined (expected 5)"
    fi
else
    fail "Cannot source version-manager.sh for platform check"
fi

#===============================================================================
# Section 6: Functional Validation (3 tests)
#===============================================================================
section "Section 6: Functional Validation"

# Test 6.1: cmd/helixagent uses version package
if grep -q 'dev.helix.agent/internal/version' "$PROJECT_ROOT/cmd/helixagent/main.go" && \
   grep -q 'appversion.Get()' "$PROJECT_ROOT/cmd/helixagent/main.go"; then
    pass "cmd/helixagent imports and uses version package"
else
    fail "cmd/helixagent not using version package"
fi

# Test 6.2: cmd/api uses version package
if grep -q '"dev.helix.agent/internal/version"' "$PROJECT_ROOT/cmd/api/main.go" && \
   grep -q 'version.Version' "$PROJECT_ROOT/cmd/api/main.go"; then
    pass "cmd/api imports and uses version package"
else
    fail "cmd/api not using version package"
fi

# Test 6.3: cmd/grpc-server uses version package
if grep -q '"dev.helix.agent/internal/version"' "$PROJECT_ROOT/cmd/grpc-server/main.go" && \
   grep -q 'version.Version' "$PROJECT_ROOT/cmd/grpc-server/main.go"; then
    pass "cmd/grpc-server imports and uses version package"
else
    fail "cmd/grpc-server not using version package"
fi

#===============================================================================
# Summary
#===============================================================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Release Build System Challenge Results${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Total:  $TOTAL"
echo -e "  ${GREEN}Passed: $PASSED${NC}"
if [ "$FAILED" -gt 0 ]; then
    echo -e "  ${RED}Failed: $FAILED${NC}"
    exit 1
else
    echo -e "  Failed: 0"
fi
echo ""
echo -e "${GREEN}All tests passed!${NC}"
