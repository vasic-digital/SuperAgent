#!/bin/bash
# Security Scan Resolution Challenge
# Validates that all security scanner configuration files, compose infrastructure,
# report directories, Makefile targets, and scan scripts are in place.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

PASS=0
FAIL=0
TOTAL=0

check() {
    local desc="$1"
    local result="$2"
    TOTAL=$((TOTAL + 1))
    if [ "$result" = "0" ]; then
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: $desc"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== Security Scan Resolution Challenge ==="
echo ""

# ============================================================================
# SECTION 1: Scanner Configuration Files
# ============================================================================
echo "--- Scanner Configuration Files ---"

for cfg in ".snyk" "sonar-project.properties" ".gosec.yml" ".trivy.yaml" ".semgrep.yml" ".hadolint.yaml" ".secrets.baseline"; do
    if [ -f "$PROJECT_ROOT/$cfg" ]; then
        check "Config file $cfg exists" "0"
    else
        check "Config file $cfg exists" "1"
    fi
done

# ============================================================================
# SECTION 2: Docker Compose Security File
# ============================================================================
echo ""
echo "--- Docker Compose Security Infrastructure ---"

COMPOSE_SEC="$PROJECT_ROOT/docker-compose.security.yml"

# Test: file exists
if [ -f "$COMPOSE_SEC" ]; then
    check "docker-compose.security.yml exists" "0"
else
    check "docker-compose.security.yml exists" "1"
fi

# Test: valid YAML
YAML_OK=1
if [ -f "$COMPOSE_SEC" ]; then
    if command -v python3 &>/dev/null; then
        if python3 -c "import yaml; yaml.safe_load(open('$COMPOSE_SEC'))" 2>/dev/null; then
            YAML_OK=0
        fi
    elif command -v ruby &>/dev/null; then
        if ruby -ryaml -e "YAML.safe_load(File.read('$COMPOSE_SEC'))" 2>/dev/null; then
            YAML_OK=0
        fi
    else
        # Fallback: basic structure markers
        if grep -qE "^services:" "$COMPOSE_SEC" 2>/dev/null; then
            YAML_OK=0
        fi
    fi
fi
check "docker-compose.security.yml is valid YAML" "$YAML_OK"

# ============================================================================
# SECTION 3: Reports Directory
# ============================================================================
echo ""
echo "--- Security Reports Directory ---"

if [ -d "$PROJECT_ROOT/reports/security" ]; then
    check "reports/security/ directory exists" "0"
else
    check "reports/security/ directory exists" "1"
fi

# ============================================================================
# SECTION 4: Makefile Security Scan Targets
# ============================================================================
echo ""
echo "--- Makefile Security Scan Targets ---"

MAKEFILE="$PROJECT_ROOT/Makefile"

for target in "security-scan-snyk" "security-scan-sonarqube" "security-scan-trivy" \
              "security-scan-gosec" "security-scan-semgrep"; do
    if grep -qE "(^|\s)$target" "$MAKEFILE" 2>/dev/null; then
        check "Makefile target '$target' defined" "0"
    else
        check "Makefile target '$target' defined" "1"
    fi
done

# ============================================================================
# SECTION 5: Security Scan Scripts
# ============================================================================
echo ""
echo "--- Security Scan Scripts ---"

for script in "scripts/security-scan.sh" "scripts/security-scan-full.sh"; do
    if [ -f "$PROJECT_ROOT/$script" ]; then
        check "Script $script exists" "0"
    else
        check "Script $script exists" "1"
    fi
done

# Test: scripts are executable
for script in "scripts/security-scan.sh" "scripts/security-scan-full.sh"; do
    if [ -f "$PROJECT_ROOT/$script" ] && [ -x "$PROJECT_ROOT/$script" ]; then
        check "Script $script is executable" "0"
    elif [ ! -f "$PROJECT_ROOT/$script" ]; then
        check "Script $script is executable (file missing)" "1"
    else
        check "Script $script is executable" "1"
    fi
done

echo ""
echo "=== Results: $PASS/$TOTAL passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
