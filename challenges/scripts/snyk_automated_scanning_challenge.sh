#!/bin/bash
# HelixAgent Challenge - Snyk Automated Scanning
# Validates that Snyk security scanning infrastructure is operational
# and properly configured for the HelixAgent codebase.
# Tests: config files, compose services, Dockerfile, scan scripts,
#        container runtime, network config, volume mounts, policy settings.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source framework if available
if [ -f "$SCRIPT_DIR/challenge_framework.sh" ]; then
    source "$SCRIPT_DIR/challenge_framework.sh"
    init_challenge "snyk-automated-scanning" "Snyk Automated Scanning"
    load_env
    FRAMEWORK_LOADED=true
else
    FRAMEWORK_LOADED=false
fi

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

record_result() {
    local test_name="$1"
    local status="$2"
    TOTAL=$((TOTAL + 1))
    if [ "$status" = "PASS" ]; then
        PASSED=$((PASSED + 1))
        echo -e "${GREEN}[PASS]${NC} $test_name"
        if [ "$FRAMEWORK_LOADED" = "true" ]; then
            record_assertion "test" "$test_name" "true" ""
        fi
    else
        FAILED=$((FAILED + 1))
        echo -e "${RED}[FAIL]${NC} $test_name"
        if [ "$FRAMEWORK_LOADED" = "true" ]; then
            record_assertion "test" "$test_name" "false" "Test failed"
        fi
    fi
}

echo "=========================================="
echo "  Snyk Automated Scanning Challenge"
echo "=========================================="
echo ""

# ============================================================================
# SECTION 1: SNYK CONFIGURATION FILES
# ============================================================================
echo -e "${BLUE}--- Section 1: Snyk Configuration Files ---${NC}"

# Test 1: Snyk root configuration exists
if [ -f "$PROJECT_ROOT/.snyk" ]; then
    record_result "Snyk configuration file (.snyk) exists" "PASS"
else
    record_result "Snyk configuration file (.snyk) exists" "FAIL"
fi

# Test 2: Snyk Docker Compose exists
if [ -f "$PROJECT_ROOT/docker/security/snyk/docker-compose.yml" ]; then
    record_result "Snyk Docker Compose file exists" "PASS"
else
    record_result "Snyk Docker Compose file exists" "FAIL"
fi

# Test 3: Snyk Dockerfile exists
if [ -f "$PROJECT_ROOT/docker/security/snyk/Dockerfile" ]; then
    record_result "Snyk Dockerfile exists" "PASS"
else
    record_result "Snyk Dockerfile exists" "FAIL"
fi

# Test 4: Compose file is valid YAML
if command -v python3 &>/dev/null; then
    if python3 -c "import yaml; yaml.safe_load(open('$PROJECT_ROOT/docker/security/snyk/docker-compose.yml'))" 2>/dev/null; then
        record_result "Snyk compose file is valid YAML" "PASS"
    else
        record_result "Snyk compose file is valid YAML" "FAIL"
    fi
elif command -v ruby &>/dev/null; then
    if ruby -ryaml -e "YAML.safe_load(File.read('$PROJECT_ROOT/docker/security/snyk/docker-compose.yml'))" 2>/dev/null; then
        record_result "Snyk compose file is valid YAML" "PASS"
    else
        record_result "Snyk compose file is valid YAML" "FAIL"
    fi
else
    # Fallback: check basic YAML structure markers
    if grep -q "^version:" "$PROJECT_ROOT/docker/security/snyk/docker-compose.yml" && \
       grep -q "^services:" "$PROJECT_ROOT/docker/security/snyk/docker-compose.yml"; then
        record_result "Snyk compose file has valid YAML structure (basic check)" "PASS"
    else
        record_result "Snyk compose file has valid YAML structure (basic check)" "FAIL"
    fi
fi

# ============================================================================
# SECTION 2: REQUIRED SCANNING SERVICES
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 2: Required Scanning Services ---${NC}"

COMPOSE_FILE="$PROJECT_ROOT/docker/security/snyk/docker-compose.yml"

# Test 5-8: Required scanning services defined
for svc in snyk-deps snyk-code snyk-iac snyk-full; do
    if grep -q "$svc" "$COMPOSE_FILE"; then
        record_result "Snyk service '$svc' defined in compose" "PASS"
    else
        record_result "Snyk service '$svc' defined in compose" "FAIL"
    fi
done

# Test 9: Each service has a build context
SVC_COUNT=$(grep -c "build:" "$COMPOSE_FILE" 2>/dev/null || echo "0")
if [ "$SVC_COUNT" -ge 4 ]; then
    record_result "All 4 Snyk services have build configuration" "PASS"
else
    record_result "All 4 Snyk services have build configuration (found $SVC_COUNT)" "FAIL"
fi

# Test 10: Services reference the Snyk Dockerfile
if grep -q "docker/security/snyk/Dockerfile" "$COMPOSE_FILE"; then
    record_result "Services reference correct Snyk Dockerfile" "PASS"
else
    record_result "Services reference correct Snyk Dockerfile" "FAIL"
fi

# ============================================================================
# SECTION 3: SNYK POLICY CONFIGURATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 3: Snyk Policy Configuration ---${NC}"

SNYK_FILE="$PROJECT_ROOT/.snyk"

# Test 11: Snyk config has language settings
if grep -q "language-settings" "$SNYK_FILE"; then
    record_result "Snyk config has language-settings section" "PASS"
else
    record_result "Snyk config has language-settings section" "FAIL"
fi

# Test 12: Snyk config includes Go language
if grep -q "go:" "$SNYK_FILE"; then
    record_result "Snyk config includes Go language" "PASS"
else
    record_result "Snyk config includes Go language" "FAIL"
fi

# Test 13: Snyk config has version field
if grep -q "^version:" "$SNYK_FILE"; then
    record_result "Snyk config has version field" "PASS"
else
    record_result "Snyk config has version field" "FAIL"
fi

# Test 14: Snyk config has ignore section (for managing exceptions)
if grep -q "^ignore:" "$SNYK_FILE"; then
    record_result "Snyk config has ignore section for vulnerability exceptions" "PASS"
else
    record_result "Snyk config has ignore section for vulnerability exceptions" "FAIL"
fi

# Test 15: Snyk config has patch section
if grep -q "^patch:" "$SNYK_FILE"; then
    record_result "Snyk config has patch section" "PASS"
else
    record_result "Snyk config has patch section" "FAIL"
fi

# ============================================================================
# SECTION 4: DOCKERFILE VALIDATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 4: Dockerfile Validation ---${NC}"

DOCKERFILE="$PROJECT_ROOT/docker/security/snyk/Dockerfile"

# Test 16: Dockerfile uses snyk/snyk-cli base image
if grep -q "snyk/snyk-cli" "$DOCKERFILE"; then
    record_result "Dockerfile uses official snyk/snyk-cli base image" "PASS"
else
    record_result "Dockerfile uses official snyk/snyk-cli base image" "FAIL"
fi

# Test 17: Dockerfile installs Go toolchain
if grep -q "go.*linux-amd64" "$DOCKERFILE"; then
    record_result "Dockerfile installs Go toolchain" "PASS"
else
    record_result "Dockerfile installs Go toolchain" "FAIL"
fi

# Test 18: Dockerfile creates scan scripts
if grep -q "scan-dependencies.sh" "$DOCKERFILE" && \
   grep -q "scan-code.sh" "$DOCKERFILE" && \
   grep -q "scan-iac.sh" "$DOCKERFILE" && \
   grep -q "scan-all.sh" "$DOCKERFILE"; then
    record_result "Dockerfile creates all 4 scan scripts" "PASS"
else
    record_result "Dockerfile creates all 4 scan scripts" "FAIL"
fi

# Test 19: Dockerfile creates /reports directory
if grep -q "/reports" "$DOCKERFILE"; then
    record_result "Dockerfile creates /reports output directory" "PASS"
else
    record_result "Dockerfile creates /reports output directory" "FAIL"
fi

# Test 20: Scripts are made executable
if grep -q "chmod +x" "$DOCKERFILE"; then
    record_result "Dockerfile makes scan scripts executable" "PASS"
else
    record_result "Dockerfile makes scan scripts executable" "FAIL"
fi

# ============================================================================
# SECTION 5: COMPOSE INFRASTRUCTURE
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 5: Compose Infrastructure ---${NC}"

# Test 21: SNYK_TOKEN environment variable referenced
if grep -q "SNYK_TOKEN" "$COMPOSE_FILE"; then
    record_result "SNYK_TOKEN environment variable configured in compose" "PASS"
else
    record_result "SNYK_TOKEN environment variable configured in compose" "FAIL"
fi

# Test 22: Reports volume defined
if grep -q "snyk-reports" "$COMPOSE_FILE"; then
    record_result "Snyk reports volume defined in compose" "PASS"
else
    record_result "Snyk reports volume defined in compose" "FAIL"
fi

# Test 23: Network configuration present
if grep -q "security-network" "$COMPOSE_FILE"; then
    record_result "Security network configured in compose" "PASS"
else
    record_result "Security network configured in compose" "FAIL"
fi

# Test 24: Compose uses profiles for selective scanning
PROFILE_COUNT=$(grep -c "profiles:" "$COMPOSE_FILE" 2>/dev/null || echo "0")
if [ "$PROFILE_COUNT" -ge 4 ]; then
    record_result "Compose uses profiles for selective scanning ($PROFILE_COUNT profiles)" "PASS"
else
    record_result "Compose uses profiles for selective scanning (found $PROFILE_COUNT)" "FAIL"
fi

# Test 25: App volume is mounted read-only
if grep -q ":ro" "$COMPOSE_FILE"; then
    record_result "App source volume mounted read-only" "PASS"
else
    record_result "App source volume mounted read-only" "FAIL"
fi

# ============================================================================
# SECTION 6: SCAN SCRIPT CONTENT VALIDATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 6: Scan Script Content ---${NC}"

# Test 26: Dependencies scan uses snyk test
if grep -q "snyk test" "$DOCKERFILE"; then
    record_result "Dependency scan script uses 'snyk test'" "PASS"
else
    record_result "Dependency scan script uses 'snyk test'" "FAIL"
fi

# Test 27: Code scan uses snyk code test
if grep -q "snyk code test" "$DOCKERFILE"; then
    record_result "Code scan script uses 'snyk code test'" "PASS"
else
    record_result "Code scan script uses 'snyk code test'" "FAIL"
fi

# Test 28: IaC scan uses snyk iac test
if grep -q "snyk iac test" "$DOCKERFILE"; then
    record_result "IaC scan script uses 'snyk iac test'" "PASS"
else
    record_result "IaC scan script uses 'snyk iac test'" "FAIL"
fi

# Test 29: Scan scripts produce JSON output
JSON_COUNT=$(grep -c "\-\-json" "$DOCKERFILE" 2>/dev/null || echo "0")
if [ "$JSON_COUNT" -ge 2 ]; then
    record_result "Scan scripts produce JSON output (--json flag)" "PASS"
else
    record_result "Scan scripts produce JSON output (--json flag)" "FAIL"
fi

# Test 30: Severity threshold configured
if grep -q "severity-threshold" "$DOCKERFILE"; then
    record_result "Severity threshold configured in scan scripts" "PASS"
else
    record_result "Severity threshold configured in scan scripts" "FAIL"
fi

# ============================================================================
# SECTION 7: HOST ENVIRONMENT
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 7: Host Environment ---${NC}"

# Test 31: Container runtime available for scanning
RUNTIME=""
if docker info &>/dev/null 2>&1; then
    RUNTIME="docker"
elif podman info &>/dev/null 2>&1; then
    RUNTIME="podman"
fi
if [ -n "$RUNTIME" ]; then
    record_result "Container runtime available ($RUNTIME)" "PASS"
else
    record_result "Container runtime available" "FAIL"
fi

# Test 32: gosec available or installable
if command -v gosec &>/dev/null || [ -f "$HOME/go/bin/gosec" ] || [ -f "$PROJECT_ROOT/bin/gosec" ]; then
    record_result "gosec binary available as complementary scanner" "PASS"
else
    record_result "gosec binary available as complementary scanner" "FAIL"
fi

# Test 33: go vet passes on internal packages
if cd "$PROJECT_ROOT" && GOMAXPROCS=2 nice -n 19 go vet ./internal/... 2>/dev/null; then
    record_result "go vet passes on internal packages" "PASS"
else
    record_result "go vet passes on internal packages" "FAIL"
fi

# Test 34: Makefile has security-scan target
if grep -qE '^security-scan:' "$PROJECT_ROOT/Makefile"; then
    record_result "Makefile has security-scan target" "PASS"
else
    record_result "Makefile has security-scan target" "FAIL"
fi

# Test 35: Makefile has sbom target
if grep -qE '^sbom:' "$PROJECT_ROOT/Makefile"; then
    record_result "Makefile has sbom (Software Bill of Materials) target" "PASS"
else
    record_result "Makefile has sbom (Software Bill of Materials) target" "FAIL"
fi

# ============================================================================
# SECTION 8: INTEGRATION WITH SONARQUBE
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 8: Integration with SonarQube ---${NC}"

# Test 36: Snyk network references SonarQube network
if grep -q "helixagent-sonarqube" "$COMPOSE_FILE"; then
    record_result "Snyk network integrated with SonarQube security network" "PASS"
else
    record_result "Snyk network integrated with SonarQube security network" "FAIL"
fi

# Test 37: SonarQube compose file exists alongside Snyk
if [ -f "$PROJECT_ROOT/docker/security/sonarqube/docker-compose.yml" ]; then
    record_result "SonarQube compose file exists alongside Snyk" "PASS"
else
    record_result "SonarQube compose file exists alongside Snyk" "FAIL"
fi

# Test 38: Security directory has both scanning tools
if [ -d "$PROJECT_ROOT/docker/security/snyk" ] && [ -d "$PROJECT_ROOT/docker/security/sonarqube" ]; then
    record_result "Security directory contains both Snyk and SonarQube" "PASS"
else
    record_result "Security directory contains both Snyk and SonarQube" "FAIL"
fi

# ============================================================================
# SUMMARY
# ============================================================================
echo ""
echo "=========================================="
echo "  Results: $PASSED/$TOTAL passed, $FAILED failed"
echo "=========================================="

if [ "$FRAMEWORK_LOADED" = "true" ]; then
    record_metric "total_tests" "$TOTAL"
    record_metric "passed_tests" "$PASSED"
    record_metric "failed_tests" "$FAILED"

    if [ $FAILED -eq 0 ]; then
        finalize_challenge "PASSED"
    else
        finalize_challenge "FAILED"
    fi
fi

if [ $FAILED -gt 0 ]; then
    exit 1
fi
exit 0
