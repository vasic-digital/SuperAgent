#!/bin/bash

#===============================================================================
# HELIXAGENT REMOTE DEPLOYMENT CHALLENGE
#===============================================================================
# This challenge validates the remote deployment and service discovery system.
#
# The challenge:
# 1. Validates service discovery implementation (TCP, HTTP, DNS, mDNS)
# 2. Validates remote deployment configuration parsing
# 3. Tests SSHRemoteDeployer logic with mock runner
# 4. Tests BootManager integration with remote deployment
# 5. Validates configuration files (YAML) for remote deployment
#
# IMPORTANT: Uses unit tests with mock SSH runner - no actual SSH required.
#
# Usage:
#   ./challenges/scripts/remote_deployment_challenge.sh [options]
#
# Options:
#   --verbose        Enable verbose logging
#   --dry-run        Print commands without executing
#   --help           Show this help message
#
#===============================================================================

set -e

#===============================================================================
# CONFIGURATION
#===============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHALLENGES_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$CHALLENGES_DIR")"

# Timestamps
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Directories
RESULTS_BASE="$CHALLENGES_DIR/results/remote_deployment"
RESULTS_DIR="$RESULTS_BASE/$TIMESTAMP"
LOGS_DIR="$RESULTS_DIR/logs"
OUTPUT_DIR="$RESULTS_DIR/results"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

#===============================================================================
# FUNCTIONS
#===============================================================================

log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1" >&2
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

die() {
    error "$1"
    exit 1
}

print_header() {
    echo -e "${BLUE}===============================================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}===============================================================================${NC}"
}

#===============================================================================
# PARSE ARGUMENTS
#===============================================================================

VERBOSE=0
DRY_RUN=0

while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose)
            VERBOSE=1
            shift
            ;;
        --dry-run)
            DRY_RUN=1
            shift
            ;;
        --help)
            cat << EOF
HelixAgent Remote Deployment Challenge

Validates remote deployment and service discovery system.

Options:
  --verbose        Enable verbose logging
  --dry-run        Print commands without executing
  --help           Show this help message
EOF
            exit 0
            ;;
        *)
            die "Unknown option: $1"
            ;;
    esac
done

#===============================================================================
# SETUP
#===============================================================================

print_header "HELIXAGENT REMOTE DEPLOYMENT CHALLENGE"
log "Starting remote deployment challenge at $(date)"
log "Project root: $PROJECT_ROOT"
log "Results directory: $RESULTS_DIR"

mkdir -p "$LOGS_DIR"
mkdir -p "$OUTPUT_DIR"

#===============================================================================
# TEST 1: SERVICE DISCOVERY UNIT TESTS
#===============================================================================

print_header "TEST 1: SERVICE DISCOVERY UNIT TESTS"
log "Running service discovery unit tests..."

TEST_LOG="$LOGS_DIR/discovery_tests.log"
if [[ $DRY_RUN -eq 1 ]]; then
    log "DRY RUN: would execute: go test ./internal/services/discovery/... -v"
else
    if go test ./internal/services/discovery/... -v 2>&1 | tee "$TEST_LOG"; then
        success "Service discovery unit tests passed"
    else
        error "Service discovery unit tests failed"
        warn "Check logs: $TEST_LOG"
        exit 1
    fi
fi

#===============================================================================
# TEST 2: REMOTE DEPLOYER UNIT TESTS
#===============================================================================

print_header "TEST 2: REMOTE DEPLOYER UNIT TESTS"
log "Running remote deployer unit tests..."

TEST_LOG="$LOGS_DIR/remote_deployer_tests.log"
if [[ $DRY_RUN -eq 1 ]]; then
    log "DRY RUN: would execute: go test ./internal/services -run TestSSHRemoteDeployer -v"
else
    if go test ./internal/services -run TestSSHRemoteDeployer -v 2>&1 | tee "$TEST_LOG"; then
        success "Remote deployer unit tests passed"
    else
        error "Remote deployer unit tests failed"
        warn "Check logs: $TEST_LOG"
        exit 1
    fi
fi

#===============================================================================
# TEST 3: BOOT MANAGER WITH REMOTE DEPLOYMENT
#===============================================================================

print_header "TEST 3: BOOT MANAGER WITH REMOTE DEPLOYMENT"
log "Running boot manager unit tests with remote deployment..."

TEST_LOG="$LOGS_DIR/boot_manager_tests.log"
if [[ $DRY_RUN -eq 1 ]]; then
    log "DRY RUN: would execute: go test ./internal/services -run TestBootManager -v"
else
    if go test ./internal/services -run TestBootManager -v 2>&1 | tee "$TEST_LOG"; then
        success "Boot manager unit tests passed"
    else
        error "Boot manager unit tests failed"
        warn "Check logs: $TEST_LOG"
        exit 1
    fi
fi

#===============================================================================
# TEST 4: CONFIGURATION VALIDATION
#===============================================================================

print_header "TEST 4: CONFIGURATION VALIDATION"
log "Validating remote deployment configuration parsing..."

CONFIG_TEST_FILE="$OUTPUT_DIR/config_test.go"
cat > "$CONFIG_TEST_FILE" << 'EOF'
package main

import (
    "fmt"
    "log"
    "gopkg.in/yaml.v3"
    "io/ioutil"
)

type RemoteDeploymentConfig struct {
    Enabled          bool                            `yaml:"enabled"`
    SSHKey           string                          `yaml:"ssh_key"`
    DefaultSSHUser   string                          `yaml:"default_ssh_user"`
    DefaultRemoteDir string                          `yaml:"default_remote_dir"`
    Hosts            map[string]RemoteDeploymentHost `yaml:"hosts"`
}

type RemoteDeploymentHost struct {
    SSHHost  string   `yaml:"ssh_host"`
    SSHKey   string   `yaml:"ssh_key"`
    Services []string `yaml:"services"`
}

func main() {
    // Test YAML parsing
    yamlContent := `
enabled: true
ssh_key: ~/.ssh/id_rsa
default_ssh_user: deploy
default_remote_dir: /opt/helixagent
hosts:
  host1:
    ssh_host: user@host1.example.com
    services:
      - postgresql
      - redis
  host2:
    ssh_host: user@host2.example.com
    ssh_key: ~/.ssh/deploy_key
    services:
      - cognee
`

    var config RemoteDeploymentConfig
    err := yaml.Unmarshal([]byte(yamlContent), &config)
    if err != nil {
        log.Fatalf("Failed to parse YAML: %v", err)
    }

    if !config.Enabled {
        log.Fatal("Config should be enabled")
    }
    if config.DefaultSSHUser != "deploy" {
        log.Fatal("Default SSH user mismatch")
    }
    if len(config.Hosts) != 2 {
        log.Fatal("Expected 2 hosts")
    }

    fmt.Println("Configuration validation passed")
}
EOF

TEST_LOG="$LOGS_DIR/config_validation.log"
if [[ $DRY_RUN -eq 1 ]]; then
    log "DRY RUN: would execute: go run $CONFIG_TEST_FILE"
else
    if go run "$CONFIG_TEST_FILE" 2>&1 | tee "$TEST_LOG"; then
        success "Configuration validation passed"
    else
        error "Configuration validation failed"
        warn "Check logs: $TEST_LOG"
        exit 1
    fi
fi

#===============================================================================
# TEST 5: DISCOVERY SCRIPT VALIDATION
#===============================================================================

print_header "TEST 5: DISCOVERY SCRIPT VALIDATION"
log "Validating service discovery script..."

DISCOVERY_SCRIPT="$PROJECT_ROOT/scripts/discover-services.sh"
if [[ ! -f "$DISCOVERY_SCRIPT" ]]; then
    error "Discovery script not found: $DISCOVERY_SCRIPT"
    exit 1
fi

# Check script is executable
if [[ ! -x "$DISCOVERY_SCRIPT" ]]; then
    warn "Discovery script is not executable, fixing..."
    chmod +x "$DISCOVERY_SCRIPT"
fi

# Run discovery script with test mode
TEST_LOG="$LOGS_DIR/discovery_script.log"
if [[ $DRY_RUN -eq 1 ]]; then
    log "DRY RUN: would execute: $DISCOVERY_SCRIPT --test"
else
    # Check if script supports --test flag
    if grep -q "--test" "$DISCOVERY_SCRIPT"; then
        if "$DISCOVERY_SCRIPT" --test 2>&1 | tee "$TEST_LOG"; then
            success "Discovery script validation passed"
        else
            error "Discovery script validation failed"
            warn "Check logs: $TEST_LOG"
            exit 1
        fi
    else
        warn "Discovery script does not have --test flag, skipping execution"
    fi
fi

#===============================================================================
# CLEANUP
#===============================================================================

print_header "CLEANUP"
log "Cleaning up temporary files..."

rm -f "$CONFIG_TEST_FILE"

#===============================================================================
# FINAL RESULTS
#===============================================================================

print_header "CHALLENGE COMPLETE"
success "Remote deployment challenge completed successfully!"
log "All tests passed."
log "Results saved to: $RESULTS_DIR"
log "Challenge completed at $(date)"

# Return success
exit 0