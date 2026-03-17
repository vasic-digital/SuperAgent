#!/bin/bash
# HelixAgent Challenge - SonarQube Automated Scanning
# Validates that SonarQube security scanning infrastructure is operational
# and properly configured for the HelixAgent codebase.
# Tests: config files, compose services, version synchronization, quality gates,
#        resource limits, network configuration, health checks, pinned versions.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source framework if available
if [ -f "$SCRIPT_DIR/challenge_framework.sh" ]; then
    source "$SCRIPT_DIR/challenge_framework.sh"
    init_challenge "sonarqube-automated-scanning" "SonarQube Automated Scanning"
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
echo "  SonarQube Automated Scanning Challenge"
echo "=========================================="
echo ""

# ============================================================================
# SECTION 1: SONARQUBE CONFIGURATION FILES
# ============================================================================
echo -e "${BLUE}--- Section 1: SonarQube Configuration Files ---${NC}"

ROOT_PROPS="$PROJECT_ROOT/sonar-project.properties"
DOCKER_COMPOSE="$PROJECT_ROOT/docker/security/sonarqube/docker-compose.yml"

# Test 1: Root sonar-project.properties exists
if [ -f "$ROOT_PROPS" ]; then
    record_result "Root sonar-project.properties exists" "PASS"
else
    record_result "Root sonar-project.properties exists" "FAIL"
fi

# Test 2: SonarQube Docker Compose exists
if [ -f "$DOCKER_COMPOSE" ]; then
    record_result "SonarQube Docker Compose file exists" "PASS"
else
    record_result "SonarQube Docker Compose file exists" "FAIL"
fi

# Test 3: Compose file is valid YAML
if command -v python3 &>/dev/null; then
    if python3 -c "import yaml; yaml.safe_load(open('$DOCKER_COMPOSE'))" 2>/dev/null; then
        record_result "SonarQube compose file is valid YAML" "PASS"
    else
        record_result "SonarQube compose file is valid YAML" "FAIL"
    fi
elif command -v ruby &>/dev/null; then
    if ruby -ryaml -e "YAML.safe_load(File.read('$DOCKER_COMPOSE'))" 2>/dev/null; then
        record_result "SonarQube compose file is valid YAML" "PASS"
    else
        record_result "SonarQube compose file is valid YAML" "FAIL"
    fi
else
    if grep -q "^version:" "$DOCKER_COMPOSE" && grep -q "^services:" "$DOCKER_COMPOSE"; then
        record_result "SonarQube compose file has valid YAML structure (basic check)" "PASS"
    else
        record_result "SonarQube compose file has valid YAML structure (basic check)" "FAIL"
    fi
fi

# ============================================================================
# SECTION 2: REQUIRED SERVICES
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 2: Required Services ---${NC}"

# Test 4: SonarQube service defined
if grep -q "sonarqube:" "$DOCKER_COMPOSE"; then
    record_result "SonarQube service defined in compose" "PASS"
else
    record_result "SonarQube service defined in compose" "FAIL"
fi

# Test 5: SonarQube database (postgres) service defined
if grep -q "postgres:" "$DOCKER_COMPOSE"; then
    record_result "SonarQube database (postgres) service defined" "PASS"
else
    record_result "SonarQube database (postgres) service defined" "FAIL"
fi

# Test 6: Sonar scanner service defined
if grep -q "sonar-scanner:" "$DOCKER_COMPOSE"; then
    record_result "Sonar scanner service defined in compose" "PASS"
else
    record_result "Sonar scanner service defined in compose" "FAIL"
fi

# Test 7: SonarQube depends on postgres
if grep -A20 "^  sonarqube:" "$DOCKER_COMPOSE" | grep -q "depends_on"; then
    record_result "SonarQube service depends on postgres" "PASS"
else
    record_result "SonarQube service depends on postgres" "FAIL"
fi

# Test 8: Scanner depends on sonarqube with health condition
if grep -A20 "sonar-scanner:" "$DOCKER_COMPOSE" | grep -q "depends_on"; then
    record_result "Scanner service depends on SonarQube" "PASS"
else
    record_result "Scanner service depends on SonarQube" "FAIL"
fi

# ============================================================================
# SECTION 3: VERSION CONFIGURATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 3: Version Configuration ---${NC}"

# Test 9: Root config has project version
if grep -q "sonar.projectVersion=" "$ROOT_PROPS"; then
    ROOT_VERSION=$(grep "sonar.projectVersion=" "$ROOT_PROPS" | cut -d= -f2)
    record_result "Root config has project version ($ROOT_VERSION)" "PASS"
else
    ROOT_VERSION=""
    record_result "Root config has project version" "FAIL"
fi

# Test 10: Root config has project key
if grep -q "sonar.projectKey=helixagent" "$ROOT_PROPS"; then
    record_result "Root config has correct project key (helixagent)" "PASS"
else
    record_result "Root config has correct project key (helixagent)" "FAIL"
fi

# Test 11: Root config has project name
if grep -q "sonar.projectName=HelixAgent" "$ROOT_PROPS"; then
    record_result "Root config has correct project name (HelixAgent)" "PASS"
else
    record_result "Root config has correct project name (HelixAgent)" "FAIL"
fi

# ============================================================================
# SECTION 4: PINNED IMAGE VERSIONS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 4: Pinned Image Versions ---${NC}"

# Test 12: SonarQube uses community edition (not :latest)
if grep -q "sonarqube:community" "$DOCKER_COMPOSE"; then
    record_result "SonarQube uses pinned community edition image" "PASS"
else
    record_result "SonarQube uses pinned community edition image" "FAIL"
fi

# Test 13: PostgreSQL uses pinned version
if grep -q "postgres:.*alpine" "$DOCKER_COMPOSE"; then
    record_result "PostgreSQL uses pinned alpine version" "PASS"
else
    record_result "PostgreSQL uses pinned alpine version" "FAIL"
fi

# Test 14: Scanner image is defined
if grep -q "sonarsource/sonar-scanner-cli" "$DOCKER_COMPOSE"; then
    record_result "Scanner uses official sonarsource/sonar-scanner-cli image" "PASS"
else
    record_result "Scanner uses official sonarsource/sonar-scanner-cli image" "FAIL"
fi

# ============================================================================
# SECTION 5: QUALITY GATE CONFIGURATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 5: Quality Gate Configuration ---${NC}"

# Test 15: Quality gate wait setting configured
if grep -q "sonar.qualitygate.wait" "$ROOT_PROPS"; then
    record_result "Quality gate wait setting configured" "PASS"
else
    record_result "Quality gate wait setting configured" "FAIL"
fi

# Test 16: Issue ignore rules configured
if grep -q "sonar.issue.ignore" "$ROOT_PROPS"; then
    record_result "Issue ignore multicriteria rules configured" "PASS"
else
    record_result "Issue ignore multicriteria rules configured" "FAIL"
fi

# Test 17: Source encoding set to UTF-8
if grep -q "sonar.sourceEncoding=UTF-8" "$ROOT_PROPS"; then
    record_result "Source encoding set to UTF-8" "PASS"
else
    record_result "Source encoding set to UTF-8" "FAIL"
fi

# ============================================================================
# SECTION 6: TEST REPORT PATHS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 6: Test Report Paths ---${NC}"

# Test 18: Go coverage report path configured
if grep -q "sonar.go.coverage.reportPaths" "$ROOT_PROPS"; then
    record_result "Go coverage report path configured" "PASS"
else
    record_result "Go coverage report path configured" "FAIL"
fi

# Test 19: Go test report path configured
if grep -q "sonar.go.tests.reportPaths" "$ROOT_PROPS"; then
    record_result "Go test report path configured" "PASS"
else
    record_result "Go test report path configured" "FAIL"
fi

# Test 20: Source exclusions configured
if grep -q "sonar.exclusions" "$ROOT_PROPS"; then
    record_result "Source exclusions configured (vendor, testdata, etc.)" "PASS"
else
    record_result "Source exclusions configured (vendor, testdata, etc.)" "FAIL"
fi

# Test 21: Test inclusions configured
if grep -q "sonar.test.inclusions" "$ROOT_PROPS"; then
    record_result "Test file inclusions configured (*_test.go)" "PASS"
else
    record_result "Test file inclusions configured (*_test.go)" "FAIL"
fi

# ============================================================================
# SECTION 7: RESOURCE LIMITS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 7: Resource Limits ---${NC}"

# Test 22: SonarQube has memory limit
if grep -q "mem_limit" "$DOCKER_COMPOSE"; then
    record_result "SonarQube containers have memory limits" "PASS"
else
    record_result "SonarQube containers have memory limits" "FAIL"
fi

# Test 23: SonarQube has CPU limit
if grep -q "cpus:" "$DOCKER_COMPOSE"; then
    record_result "SonarQube containers have CPU limits" "PASS"
else
    record_result "SonarQube containers have CPU limits" "FAIL"
fi

# Test 24: PostgreSQL has resource limits
if grep -A20 "postgres:" "$DOCKER_COMPOSE" | grep -q "mem_limit"; then
    record_result "PostgreSQL service has memory limit" "PASS"
else
    record_result "PostgreSQL service has memory limit" "FAIL"
fi

# Test 25: Ulimits configured for SonarQube
if grep -q "ulimits:" "$DOCKER_COMPOSE"; then
    record_result "SonarQube has ulimits configured (nofile, nproc)" "PASS"
else
    record_result "SonarQube has ulimits configured (nofile, nproc)" "FAIL"
fi

# ============================================================================
# SECTION 8: NETWORK CONFIGURATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 8: Network Configuration ---${NC}"

# Test 26: Security network defined
if grep -q "security-network:" "$DOCKER_COMPOSE"; then
    record_result "Security network defined in compose" "PASS"
else
    record_result "Security network defined in compose" "FAIL"
fi

# Test 27: Network uses bridge driver
if grep -q "driver: bridge" "$DOCKER_COMPOSE"; then
    record_result "Security network uses bridge driver" "PASS"
else
    record_result "Security network uses bridge driver" "FAIL"
fi

# Test 28: Network has IPAM subnet configuration
if grep -q "subnet:" "$DOCKER_COMPOSE"; then
    record_result "Network has IPAM subnet configuration" "PASS"
else
    record_result "Network has IPAM subnet configuration" "FAIL"
fi

# ============================================================================
# SECTION 9: HEALTH CHECKS
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 9: Health Checks ---${NC}"

# Test 29: SonarQube has healthcheck
if grep -A30 "^  sonarqube:" "$DOCKER_COMPOSE" | grep -q "healthcheck:"; then
    record_result "SonarQube service has health check" "PASS"
else
    record_result "SonarQube service has health check" "FAIL"
fi

# Test 30: PostgreSQL has healthcheck
if grep -A15 "^  postgres:" "$DOCKER_COMPOSE" | grep -q "healthcheck:"; then
    record_result "PostgreSQL service has health check" "PASS"
else
    record_result "PostgreSQL service has health check" "FAIL"
fi

# Test 31: SonarQube health check uses API endpoint
if grep -q "api/system/status" "$DOCKER_COMPOSE"; then
    record_result "SonarQube health check uses /api/system/status" "PASS"
else
    record_result "SonarQube health check uses /api/system/status" "FAIL"
fi

# Test 32: PostgreSQL health check uses pg_isready
if grep -q "pg_isready" "$DOCKER_COMPOSE"; then
    record_result "PostgreSQL health check uses pg_isready" "PASS"
else
    record_result "PostgreSQL health check uses pg_isready" "FAIL"
fi

# Test 33: Restart policy configured
RESTART_COUNT=$(grep -c "restart:" "$DOCKER_COMPOSE" 2>/dev/null || echo "0")
if [ "$RESTART_COUNT" -ge 2 ]; then
    record_result "Restart policy configured on services ($RESTART_COUNT)" "PASS"
else
    record_result "Restart policy configured on services" "FAIL"
fi

# ============================================================================
# SECTION 10: VOLUME CONFIGURATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 10: Volume Configuration ---${NC}"

# Test 34: SonarQube data volume defined
if grep -q "sonarqube_data:" "$DOCKER_COMPOSE"; then
    record_result "SonarQube data volume defined" "PASS"
else
    record_result "SonarQube data volume defined" "FAIL"
fi

# Test 35: SonarQube extensions volume defined
if grep -q "sonarqube_extensions:" "$DOCKER_COMPOSE"; then
    record_result "SonarQube extensions volume defined" "PASS"
else
    record_result "SonarQube extensions volume defined" "FAIL"
fi

# Test 36: SonarQube logs volume defined
if grep -q "sonarqube_logs:" "$DOCKER_COMPOSE"; then
    record_result "SonarQube logs volume defined" "PASS"
else
    record_result "SonarQube logs volume defined" "FAIL"
fi

# Test 37: PostgreSQL data volume defined
if grep -q "postgres_data:" "$DOCKER_COMPOSE"; then
    record_result "PostgreSQL data volume defined" "PASS"
else
    record_result "PostgreSQL data volume defined" "FAIL"
fi

# ============================================================================
# SECTION 11: ENVIRONMENT AND AUTHENTICATION
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 11: Environment and Authentication ---${NC}"

# Test 38: SONAR_TOKEN environment variable referenced
if grep -q "SONAR_TOKEN" "$DOCKER_COMPOSE"; then
    record_result "SONAR_TOKEN environment variable configured" "PASS"
else
    record_result "SONAR_TOKEN environment variable configured" "FAIL"
fi

# Test 39: SONAR_HOST_URL configured for scanner
if grep -q "SONAR_HOST_URL" "$DOCKER_COMPOSE"; then
    record_result "SONAR_HOST_URL configured for scanner" "PASS"
else
    record_result "SONAR_HOST_URL configured for scanner" "FAIL"
fi

# Test 40: JDBC connection configured for SonarQube
if grep -q "SONAR_JDBC_URL" "$DOCKER_COMPOSE"; then
    record_result "JDBC connection URL configured for SonarQube" "PASS"
else
    record_result "JDBC connection URL configured for SonarQube" "FAIL"
fi

# Test 41: Elasticsearch bootstrap checks disabled for dev
if grep -q "SONAR_ES_BOOTSTRAP_CHECKS_DISABLE" "$DOCKER_COMPOSE"; then
    record_result "Elasticsearch bootstrap checks disabled for dev" "PASS"
else
    record_result "Elasticsearch bootstrap checks disabled for dev" "FAIL"
fi

# Test 42: SonarQube port exposed (9000)
if grep -q "9000:9000" "$DOCKER_COMPOSE"; then
    record_result "SonarQube port 9000 exposed" "PASS"
else
    record_result "SonarQube port 9000 exposed" "FAIL"
fi

# ============================================================================
# SECTION 12: HOST ENVIRONMENT
# ============================================================================
echo ""
echo -e "${BLUE}--- Section 12: Host Environment ---${NC}"

# Test 43: Container runtime available
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

# Test 44: Security test directory exists
if [ -d "$PROJECT_ROOT/tests/security" ]; then
    record_result "Security test directory exists" "PASS"
else
    record_result "Security test directory exists" "FAIL"
fi

# Test 45: Scanning infrastructure test file exists
if [ -f "$PROJECT_ROOT/tests/security/scanning_infrastructure_test.go" ]; then
    record_result "Scanning infrastructure Go test file exists" "PASS"
else
    record_result "Scanning infrastructure Go test file exists" "FAIL"
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
