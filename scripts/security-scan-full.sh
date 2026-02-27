#!/bin/bash
#
# Comprehensive Security Scanning Script for HelixAgent
# This script orchestrates all security scanning tools: SonarQube, Snyk, Gosec, Semgrep, Trivy, Kics, Grype
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="HelixAgent"
REPORTS_DIR="reports/security"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
SONARQUBE_URL="http://localhost:9000"
SONARQUBE_TOKEN="${SONAR_TOKEN:-}"
SNYK_TOKEN="${SNYK_TOKEN:-}"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Setup function
setup() {
    log_info "Setting up security scanning environment..."
    
    mkdir -p "$REPORTS_DIR"
    
    # Check Docker/Podman availability
    if command -v docker &> /dev/null; then
        CONTAINER_RUNTIME="docker"
        COMPOSE_CMD="docker compose"
    elif command -v podman &> /dev/null; then
        CONTAINER_RUNTIME="podman"
        COMPOSE_CMD="podman-compose"
    else
        log_error "Neither Docker nor Podman found. Cannot run containerized security scans."
        exit 1
    fi
    
    log_success "Using container runtime: $CONTAINER_RUNTIME"
}

# SonarQube functions
start_sonarqube() {
    log_info "Starting SonarQube..."
    
    if ! $COMPOSE_CMD -f docker/security/sonarqube/docker-compose.yml ps | grep -q "healthy"; then
        log_info "Starting SonarQube containers..."
        $COMPOSE_CMD -f docker/security/sonarqube/docker-compose.yml up -d
        
        log_info "Waiting for SonarQube to be ready..."
        for i in {1..30}; do
            if curl -s "$SONARQUBE_URL/api/system/status" | grep -q '"status":"UP"'; then
                log_success "SonarQube is ready"
                return 0
            fi
            echo -n "."
            sleep 10
        done
        
        log_error "SonarQube failed to start within 5 minutes"
        return 1
    else
        log_success "SonarQube is already running"
    fi
}

stop_sonarqube() {
    log_info "Stopping SonarQube..."
    $COMPOSE_CMD -f docker/security/sonarqube/docker-compose.yml down
    log_success "SonarQube stopped"
}

run_sonar_scan() {
    log_info "Running SonarQube scan..."
    
    if [ -z "$SONARQUBE_TOKEN" ]; then
        log_warn "SONAR_TOKEN not set. Skipping SonarQube scan."
        log_warn "To set token: export SONAR_TOKEN=your_token_here"
        return 0
    fi
    
    # Run sonar-scanner
    $COMPOSE_CMD -f docker/security/sonarqube/docker-compose.yml --profile scanner run --rm sonar-scanner \
        -Dsonar.login="$SONARQUBE_TOKEN" \
        -Dsonar.projectKey=helixagent \
        -Dsonar.projectName="HelixAgent" \
        -Dsonar.projectVersion="1.0.0"
    
    log_success "SonarQube scan complete"
    log_info "View results at: $SONARQUBE_URL/dashboard?id=helixagent"
}

# Snyk functions
run_snyk_scan() {
    log_info "Running Snyk security scan..."
    
    if [ -z "$SNYK_TOKEN" ]; then
        log_warn "SNYK_TOKEN not set. Skipping Snyk scan."
        log_warn "To set token: export SNYK_TOKEN=your_token_here"
        return 0
    fi
    
    # Build and run Snyk scanner
    $COMPOSE_CMD -f docker/security/snyk/docker-compose.yml --profile full run --rm snyk-full
    
    log_success "Snyk scan complete"
    log_info "Reports saved to: $REPORTS_DIR/"
}

# Gosec (Go Security Checker)
run_gosec_scan() {
    log_info "Running Gosec security scan..."
    
    if ! command -v gosec &> /dev/null; then
        log_info "Installing gosec..."
        go install github.com/securego/gosec/v2/cmd/gosec@latest
    fi
    
    mkdir -p "$REPORTS_DIR"
    
    gosec -fmt=json -out="$REPORTS_DIR/gosec-report-$TIMESTAMP.json" \
          -severity=medium -confidence=medium ./... 2>/dev/null || true
    
    gosec -fmt=sonarqube -out="$REPORTS_DIR/gosec-sonar-$TIMESTAMP.json" \
          -severity=medium -confidence=medium ./... 2>/dev/null || true
    
    log_success "Gosec scan complete"
}

# Semgrep
run_semgrep_scan() {
    log_info "Running Semgrep scan..."
    
    if command -v semgrep &> /dev/null; then
        semgrep --config=auto --json --output="$REPORTS_DIR/semgrep-report-$TIMESTAMP.json" \
                --metrics=off ./ 2>/dev/null || true
        log_success "Semgrep scan complete"
    else
        log_warn "Semgrep not installed. Running via Docker..."
        $CONTAINER_RUNTIME run --rm -v "$(pwd):/app:ro" \
            -v "$(pwd)/$REPORTS_DIR:/reports" \
            returntocorp/semgrep:latest \
            --config auto --json --output /reports/semgrep-report-$TIMESTAMP.json \
            --metrics off /app 2>/dev/null || true
        log_success "Semgrep scan (Docker) complete"
    fi
}

# Trivy
run_trivy_scan() {
    log_info "Running Trivy vulnerability scan..."
    
    if command -v trivy &> /dev/null; then
        # Filesystem scan
        trivy filesystem --format json --output "$REPORTS_DIR/trivy-fs-$TIMESTAMP.json" \
            --severity HIGH,CRITICAL . 2>/dev/null || true
        
        # If Docker image exists, scan it
        if $CONTAINER_RUNTIME image inspect helixagent:latest &> /dev/null; then
            trivy image --format json --output "$REPORTS_DIR/trivy-image-$TIMESTAMP.json" \
                --severity HIGH,CRITICAL helixagent:latest 2>/dev/null || true
        fi
        
        log_success "Trivy scan complete"
    else
        log_warn "Trivy not installed. Install with: https://aquasecurity.github.io/trivy/"
    fi
}

# Kics (Infrastructure as Code scanner)
run_kics_scan() {
    log_info "Running KICS IaC scan..."
    
    $CONTAINER_RUNTIME run --rm -v "$(pwd):/app:ro" \
        -v "$(pwd)/$REPORTS_DIR:/reports" \
        checkmarx/kics:latest scan \
        -p /app \
        -o /reports \
        --report-formats json \
        --output-name kics-report-$TIMESTAMP 2>/dev/null || true
    
    log_success "KICS scan complete"
}

# Grype
run_grype_scan() {
    log_info "Running Grype vulnerability scan..."
    
    $CONTAINER_RUNTIME run --rm -v "$(pwd):/app:ro" \
        -v "$(pwd)/$REPORTS_DIR:/reports" \
        anchore/grype:latest \
        dir:/app -o json --file /reports/grype-report-$TIMESTAMP.json 2>/dev/null || true
    
    log_success "Grype scan complete"
}

# Generate summary report
generate_summary() {
    log_info "Generating security summary report..."
    
    cat > "$REPORTS_DIR/security-summary-$TIMESTAMP.md" << EOF
# HelixAgent Security Scan Summary

**Date:** $(date -u +%Y-%m-%d\ %H:%M:%S\ UTC)  
**Project:** $PROJECT_NAME  
**Scan ID:** $TIMESTAMP

## Scan Results

### 1. SonarQube
- Status: $(if [ -f "$REPORTS_DIR/sonarqube-report-$TIMESTAMP.json" ]; then "✅ Complete"; else "⏭️  Skipped"; fi)
- URL: $SONARQUBE_URL/dashboard?id=helixagent

### 2. Snyk
- Status: $(if [ -f "$REPORTS_DIR/snyk-deps-$TIMESTAMP.json" ]; then "✅ Complete"; else "⏭️  Skipped"; fi)
- Dependency Scan: $(if [ -f "$REPORTS_DIR/snyk-deps-$TIMESTAMP.json" ]; then "✅" else "❌"; fi)
- Code Scan: $(if [ -f "$REPORTS_DIR/snyk-code-$TIMESTAMP.json" ]; then "✅" else "❌"; fi)

### 3. Gosec (Go Security)
- Status: $(if [ -f "$REPORTS_DIR/gosec-report-$TIMESTAMP.json" ]; then "✅ Complete"; else "⏭️  Skipped"; fi)
- Report: \`$REPORTS_DIR/gosec-report-$TIMESTAMP.json\`

### 4. Semgrep
- Status: $(if [ -f "$REPORTS_DIR/semgrep-report-$TIMESTAMP.json" ]; then "✅ Complete"; else "⏭️  Skipped"; fi)
- Report: \`$REPORTS_DIR/semgrep-report-$TIMESTAMP.json\`

### 5. Trivy
- Status: $(if [ -f "$REPORTS_DIR/trivy-fs-$TIMESTAMP.json" ]; then "✅ Complete"; else "⏭️  Skipped"; fi)
- Filesystem Scan: $(if [ -f "$REPORTS_DIR/trivy-fs-$TIMESTAMP.json" ]; then "✅" else "❌"; fi)
- Container Scan: $(if [ -f "$REPORTS_DIR/trivy-image-$TIMESTAMP.json" ]; then "✅" else "❌"; fi)

### 6. KICS (IaC)
- Status: $(if [ -f "$REPORTS_DIR/kics-report-$TIMESTAMP.json" ]; then "✅ Complete"; else "⏭️  Skipped"; fi)
- Report: \`$REPORTS_DIR/kics-report-$TIMESTAMP.json\`

### 7. Grype
- Status: $(if [ -f "$REPORTS_DIR/grype-report-$TIMESTAMP.json" ]; then "✅ Complete"; else "⏭️  Skipped"; fi)
- Report: \`$REPORTS_DIR/grype-report-$TIMESTAMP.json\`

## Next Steps

1. Review individual scan reports in \`$REPORTS_DIR/\`
2. Address HIGH and CRITICAL severity issues
3. Set up automated scanning in CI/CD pipeline
4. Configure security dashboards

## Reports Location

All detailed reports are available in: \`$REPORTS_DIR/\`

EOF

    log_success "Summary report generated: $REPORTS_DIR/security-summary-$TIMESTAMP.md"
}

# Main execution
main() {
    echo "=========================================="
    echo "🔒 HelixAgent Security Scan Suite"
    echo "=========================================="
    echo ""
    
    setup
    
    case "${1:-all}" in
        sonarqube|sonar)
            start_sonarqube
            run_sonar_scan
            ;;
        snyk)
            run_snyk_scan
            ;;
        gosec)
            run_gosec_scan
            ;;
        semgrep)
            run_semgrep_scan
            ;;
        trivy)
            run_trivy_scan
            ;;
        kics)
            run_kics_scan
            ;;
        grype)
            run_grype_scan
            ;;
        start-sonar)
            start_sonarqube
            ;;
        stop-sonar)
            stop_sonarqube
            ;;
        all)
            log_info "Running all security scans..."
            start_sonarqube
            run_sonar_scan
            run_snyk_scan
            run_gosec_scan
            run_semgrep_scan
            run_trivy_scan
            run_kics_scan
            run_grype_scan
            generate_summary
            ;;
        quick)
            log_info "Running quick security scans (Gosec, Semgrep)..."
            run_gosec_scan
            run_semgrep_scan
            ;;
        *)
            echo "Usage: $0 [sonarqube|snyk|gosec|semgrep|trivy|kics|grype|start-sonar|stop-sonar|all|quick]"
            echo ""
            echo "Commands:"
            echo "  sonarqube   - Run SonarQube scan (starts server if needed)"
            echo "  snyk        - Run Snyk security scan"
            echo "  gosec       - Run Gosec Go security scan"
            echo "  semgrep     - Run Semgrep static analysis"
            echo "  trivy       - Run Trivy vulnerability scan"
            echo "  kics        - Run KICS IaC scan"
            echo "  grype       - Run Grype vulnerability scan"
            echo "  start-sonar - Start SonarQube server only"
            echo "  stop-sonar  - Stop SonarQube server"
            echo "  all         - Run all security scans (default)"
            echo "  quick       - Run quick scans (Gosec + Semgrep)"
            exit 1
            ;;
    esac
    
    echo ""
    echo "=========================================="
    log_success "Security scanning complete!"
    echo "=========================================="
}

# Run main function
main "$@"
