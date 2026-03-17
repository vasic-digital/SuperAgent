# User Manual 32: Automated Security Scanning

**Version:** 1.0
**Last Updated:** March 17, 2026
**Audience:** Security Engineers, DevOps Engineers, Platform Engineers

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Snyk Containerized Pipeline](#snyk-containerized-pipeline)
4. [SonarQube Containerized Pipeline](#sonarqube-containerized-pipeline)
5. [Unified Security Gate](#unified-security-gate)
6. [Challenge Validation](#challenge-validation)
7. [Report Formats and Aggregation](#report-formats-and-aggregation)
8. [Scheduling and Automation](#scheduling-and-automation)
9. [Troubleshooting](#troubleshooting)

---

## Overview

HelixAgent implements automated, containerized security scanning using Snyk (dependency and code analysis) and SonarQube (code quality and security analysis). Both tools run entirely inside Docker/Podman containers, requiring no local installation. Scanning results are validated by dedicated challenge scripts that enforce zero false positives.

### Key Principles

- **Containerized execution**: All scanners run inside containers for reproducibility
- **Resource-limited**: Scanning containers are capped at 30-40% of host resources
- **Automated validation**: Challenge scripts (38 tests for Snyk, 45 tests for SonarQube) verify scanner correctness
- **Quality gates**: Critical and blocker findings block the build pipeline
- **Report aggregation**: Results from all tools are merged into a unified security report

---

## Architecture

```
HelixAgent Security Scanning Pipeline
======================================

 [Source Code] --> [Snyk Container]      --> [Snyk Report JSON]      --+
                                                                       |
 [Source Code] --> [SonarQube Scanner]   --> [SonarQube Report JSON] --+--> [Aggregator]
                   [SonarQube Server]                                  |       |
                                                                       |    [Unified Report]
 [Source Code] --> [Gosec/Semgrep/Trivy] --> [Tool Reports]          --+       |
                                                                            [Quality Gate]
                                                                               |
                                                                          PASS / FAIL
```

### Container Dependencies

| Container | Image | Port | Purpose |
|-----------|-------|------|---------|
| sonarqube | `sonarqube:community` | 9000 | SonarQube server |
| sonar-scanner | `sonarsource/sonar-scanner-cli` | - | SonarQube analysis runner |
| snyk | `snyk/snyk-cli` | - | Snyk vulnerability scanner |

---

## Snyk Containerized Pipeline

### Prerequisites

- Docker or Podman available
- Snyk API token (free tier available at snyk.io)
- Network access to Snyk API servers

### Configuration

**Environment variables:**

```bash
# Required
export SNYK_TOKEN=your_snyk_api_token

# Optional
export SNYK_ORG=your_organization_id
export SNYK_SEVERITY_THRESHOLD=high  # low, medium, high, critical
```

**Docker Compose configuration:** `docker/security/snyk/docker-compose.yml`

### Running Scans

#### Full Scan (All Analysis Types)

```bash
docker compose -f docker/security/snyk/docker-compose.yml \
  --profile full run --rm snyk-full
```

This runs three analysis types sequentially:
1. **Dependency scan** (`snyk test`): Checks `go.mod` and all submodule dependencies
2. **Code analysis** (`snyk code test`): Static analysis of Go source files
3. **Container scan** (`snyk container test`): Scans the HelixAgent Docker image

#### Individual Scans

```bash
# Dependency vulnerabilities only
docker compose -f docker/security/snyk/docker-compose.yml \
  --profile deps run --rm snyk-deps

# Code analysis only
docker compose -f docker/security/snyk/docker-compose.yml \
  --profile code run --rm snyk-code

# Container image scan only
docker compose -f docker/security/snyk/docker-compose.yml \
  --profile container run --rm snyk-container
```

### Understanding Snyk Results

Snyk categorizes findings by severity:

| Severity | CVSS Range | Action Required | SLA |
|----------|-----------|-----------------|-----|
| Critical | 9.0-10.0 | Immediate fix | 24 hours |
| High | 7.0-8.9 | Urgent fix | 48 hours |
| Medium | 4.0-6.9 | Planned fix | 1 sprint |
| Low | 0.1-3.9 | Backlog | Best effort |

**Example finding:**

```json
{
  "id": "SNYK-GOLANG-GITHUBCOMJACKC-PGX5-6241800",
  "title": "SQL Injection",
  "severity": "high",
  "cvssScore": 7.5,
  "packageName": "github.com/jackc/pgx/v5",
  "version": "5.5.0",
  "fixedIn": ["5.5.4"],
  "upgradePath": ["github.com/jackc/pgx/v5@5.5.4"]
}
```

**Remediation:**

```bash
# Update the vulnerable dependency
go get github.com/jackc/pgx/v5@v5.5.4
go mod tidy

# Verify the fix
docker compose -f docker/security/snyk/docker-compose.yml \
  --profile deps run --rm snyk-deps
```

### Snyk Ignore Policies

For accepted risks, create `.snyk` in the project root:

```yaml
version: v1.25.0
ignore:
  SNYK-GOLANG-EXAMPLE-12345:
    - '*':
        reason: 'Accepted risk: no user input reaches this code path'
        expires: 2026-06-01T00:00:00.000Z
```

---

## SonarQube Containerized Pipeline

### Prerequisites

- Docker or Podman available
- At least 4GB of memory available for the SonarQube container
- Port 9000 available

### Starting the SonarQube Server

```bash
# Start SonarQube (first startup takes 60-90 seconds)
docker compose -f docker/security/sonarqube/docker-compose.yml up -d sonarqube

# Wait for readiness
echo "Waiting for SonarQube to start..."
until curl -s http://localhost:9000/api/system/status | grep -q '"status":"UP"'; do
  sleep 5
  echo "  Still waiting..."
done
echo "SonarQube is ready."
```

### Running Analysis

```bash
# Run the SonarQube scanner
docker compose -f docker/security/sonarqube/docker-compose.yml \
  run --rm sonar-scanner
```

### Configuration

**Project properties:** `docker/security/sonarqube/sonar-project.properties`

```properties
sonar.projectKey=helixagent
sonar.projectName=HelixAgent
sonar.projectVersion=1.0

# Source configuration
sonar.sources=.
sonar.exclusions=vendor/**,cli_agents/**,MCP/**,testdata/**,**/mock_*.go
sonar.tests=.
sonar.test.inclusions=**/*_test.go

# Go-specific settings
sonar.go.coverage.reportPaths=coverage.out
sonar.go.tests.reportPaths=test-report.json

# Encoding
sonar.sourceEncoding=UTF-8
```

### Quality Gates

SonarQube enforces quality gates that block the pipeline:

| Metric | Condition | Threshold |
|--------|-----------|-----------|
| New bugs | Is greater than | 0 |
| New vulnerabilities | Is greater than | 0 |
| New security hotspots reviewed | Is less than | 100% |
| New code coverage | Is less than | 80% |
| New duplicated lines density | Is greater than | 3% |

### Accessing the Dashboard

```bash
# Open SonarQube in browser
open http://localhost:9000

# Default credentials (change immediately)
# Username: admin
# Password: admin
```

Navigate to Projects > helixagent to view:
- Overall code quality rating (A-E)
- Vulnerability count by severity
- Code smell count
- Technical debt estimate
- Coverage percentage

### SonarQube API

Query results programmatically:

```bash
# Get project status
curl -s -u admin:admin \
  "http://localhost:9000/api/qualitygates/project_status?projectKey=helixagent"

# Get issues by severity
curl -s -u admin:admin \
  "http://localhost:9000/api/issues/search?componentKeys=helixagent&severities=CRITICAL,BLOCKER"

# Get metrics
curl -s -u admin:admin \
  "http://localhost:9000/api/measures/component?component=helixagent&metricKeys=bugs,vulnerabilities,code_smells,coverage"
```

### Stopping SonarQube

```bash
docker compose -f docker/security/sonarqube/docker-compose.yml down

# To remove data volumes (clean slate):
docker compose -f docker/security/sonarqube/docker-compose.yml down -v
```

---

## Unified Security Gate

The unified security gate aggregates results from Snyk and SonarQube to make a pass/fail decision:

```bash
# Run full security gate
make ci-security-gate
```

**Gate Logic:**

```
PASS if:
  - Snyk: 0 critical, 0 high (unpatched)
  - SonarQube: Quality gate passes
  - All scanner containers ran successfully
  - Reports are non-empty and valid JSON

FAIL if:
  - Any critical or blocker finding exists
  - Scanner container exited with error
  - Reports are missing or malformed
  - Quality gate status is ERROR
```

---

## Challenge Validation

### Snyk Scanning Challenge

**Script:** `challenges/scripts/snyk_automated_scanning_challenge.sh`
**Tests:** 38

Validates:
1. Snyk container image is available
2. Authentication succeeds with provided token
3. Dependency scan completes without errors
4. Code analysis scan completes
5. Container image scan completes
6. JSON report is generated and valid
7. Findings include expected fields (id, severity, title)
8. Severity threshold filtering works correctly
9. Exit codes match CI/CD gating expectations
10. Resource limits are enforced on the scanner container
11. Report aggregation produces correct summary
12. Multiple submodule scanning covers all 27 modules

```bash
./challenges/scripts/snyk_automated_scanning_challenge.sh
```

### SonarQube Scanning Challenge

**Script:** `challenges/scripts/sonarqube_automated_scanning_challenge.sh`
**Tests:** 45

Validates:
1. SonarQube container starts and reaches UP status
2. Project is created and configured
3. Scanner executes analysis without errors
4. Quality gate evaluates correctly
5. Issues are categorized by type (bug, vulnerability, code smell)
6. Severity mapping is correct (blocker, critical, major, minor, info)
7. API endpoints return expected data
8. Report export works (JSON, HTML)
9. Metrics are collected (coverage, duplications, complexity)
10. Multi-language analysis works (Go, YAML, Docker, SQL)
11. Custom rules are applied
12. Incremental analysis detects new issues

```bash
./challenges/scripts/sonarqube_automated_scanning_challenge.sh
```

---

## Report Formats and Aggregation

### Individual Reports

Each scanner produces reports in `reports/security/`:

```
reports/security/
  snyk-deps-YYYYMMDD_HHMMSS.json      # Snyk dependency scan
  snyk-code-YYYYMMDD_HHMMSS.json      # Snyk code analysis
  snyk-container-YYYYMMDD_HHMMSS.json  # Snyk container scan
  sonarqube-report-YYYYMMDD_HHMMSS.json # SonarQube analysis
```

### Aggregated Report

```bash
# Generate aggregated security report
./scripts/security-aggregate-report.sh
```

Produces:
- `reports/security/aggregated-summary.md` -- Human-readable markdown summary
- `reports/security/aggregated-results.json` -- Machine-readable JSON with all findings

### Report Schema

```json
{
  "scan_timestamp": "2026-03-15T10:30:00Z",
  "tools": {
    "snyk": {"version": "1.1234.0", "scans": 3, "findings": 5},
    "sonarqube": {"version": "10.4", "scans": 1, "findings": 12}
  },
  "summary": {
    "critical": 0,
    "high": 2,
    "medium": 8,
    "low": 7,
    "total": 17
  },
  "gate_status": "PASS",
  "findings": [...]
}
```

---

## Scheduling and Automation

### Daily Quick Scan

```bash
# Add to crontab
0 6 * * * cd /path/to/helixagent && \
  SNYK_TOKEN=$SNYK_TOKEN \
  GOMAXPROCS=2 nice -n 19 ionice -c 3 \
  ./scripts/security-scan-full.sh quick >> /var/log/helixagent-security.log 2>&1
```

### Weekly Full Scan

```bash
# Weekly full scan on Sundays
0 2 * * 0 cd /path/to/helixagent && \
  SNYK_TOKEN=$SNYK_TOKEN \
  docker compose -f docker/security/sonarqube/docker-compose.yml up -d sonarqube && \
  sleep 90 && \
  ./scripts/security-scan-full.sh all >> /var/log/helixagent-security-full.log 2>&1 && \
  docker compose -f docker/security/sonarqube/docker-compose.yml down
```

### Pre-Release Scan

Before every release, run the full security gate:

```bash
# Must pass before make release
make ci-security-gate
```

---

## Troubleshooting

### SonarQube fails to start

**Symptom:** Container exits immediately or health check never succeeds.

**Cause:** Insufficient memory or `vm.max_map_count` too low.

**Solutions:**
```bash
# Check container logs
docker logs helixagent-sonarqube

# Increase vm.max_map_count (Linux)
sudo sysctl -w vm.max_map_count=524288

# Allocate more memory
docker compose -f docker/security/sonarqube/docker-compose.yml down
# Edit docker-compose.yml to increase memory limit to 4G+
docker compose -f docker/security/sonarqube/docker-compose.yml up -d sonarqube
```

### Snyk authentication fails

**Symptom:** "Authentication failed" or empty results.

**Solutions:**
```bash
# Verify token is set
echo $SNYK_TOKEN

# Test authentication
docker run --rm -e SNYK_TOKEN=$SNYK_TOKEN snyk/snyk-cli auth

# Regenerate token at https://app.snyk.io/account
```

### Scanner reports are empty

**Symptom:** JSON reports contain no findings.

**Solutions:**
1. Verify source code is mounted correctly in the container
2. Check scanner logs for errors
3. Ensure the project language is configured correctly
4. Run the scanner outside Docker to compare results

### Quality gate always fails

**Symptom:** SonarQube quality gate reports ERROR even for clean code.

**Solutions:**
1. Check if the quality gate is configured for "new code" vs "overall code"
2. Reset the new code period: SonarQube > Project Settings > New Code
3. Verify coverage reports are generated and found by the scanner

---

## Related Resources

- [User Manual 17: Security Scanning Guide](17-security-scanning-guide.md) -- Overview of all 7 security tools
- [User Manual 31: Fuzz Testing Guide](31-fuzz-testing-guide.md) -- Complementary robustness testing
- [Video Course 63: Automated Security Scanning](../video-courses/video-course-63-automated-security-scanning.md) -- Video walkthrough
- [Snyk Documentation](https://docs.snyk.io/)
- [SonarQube Documentation](https://docs.sonarqube.org/)

---

**Next Manual:** User Manual 33 - Performance Optimization Guide
