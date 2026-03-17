# User Manual 17: Security Scanning Guide

**Version:** 1.0  
**Last Updated:** February 27, 2026  
**Audience:** Security Engineers, DevOps, System Administrators

---

## Table of Contents

1. [Overview](#overview)
2. [Security Tools](#security-tools)
3. [Running Security Scans](#running-security-scans)
4. [Interpreting Results](#interpreting-results)
5. [Remediation Workflow](#remediation-workflow)
6. [CI/CD Integration](#cicd-integration)
7. [Troubleshooting](#troubleshooting)

---

## Overview

HelixAgent includes a comprehensive security scanning infrastructure that integrates 7 industry-standard security tools:

- **SonarQube** - Code quality and security analysis
- **Snyk** - Dependency vulnerability scanning
- **Gosec** - Go security checker
- **Semgrep** - Static analysis
- **Trivy** - Container and filesystem scanning
- **KICS** - Infrastructure as Code scanning
- **Grype** - Vulnerability scanner

### Security Coverage

| Category | Tools | Coverage |
|----------|-------|----------|
| Code Quality | SonarQube, Semgrep | 100% |
| Dependencies | Snyk, Trivy, Grype | 100% |
| Go Security | Gosec | 100% |
| Infrastructure | KICS | 100% |
| Containers | Trivy | 100% |

---

## Security Tools

### 1. SonarQube

**Purpose:** Code quality and security analysis

**Features:**
- Static code analysis
- Security vulnerability detection
- Code smell detection
- Technical debt tracking
- Quality gates

**Setup:**

```bash
# Start SonarQube
docker compose -f docker/security/sonarqube/docker-compose.yml up -d

# Access SonarQube
open http://localhost:9000

# Default credentials
Username: admin
Password: admin
```

**Configuration:**

Edit `docker/security/sonarqube/sonar-project.properties`:

```properties
sonar.projectKey=helixagent
sonar.projectName=HelixAgent
sonar.sources=.
sonar.exclusions=vendor/**,cli_agents/**
```

### 2. Snyk

**Purpose:** Dependency and container vulnerability scanning

**Features:**
- Dependency vulnerability detection
- Container image scanning
- Infrastructure as Code scanning
- License compliance

**Setup:**

```bash
# Set API token
export SNYK_TOKEN=your_token_here

# Run Snyk scan
docker compose -f docker/security/snyk/docker-compose.yml --profile full run --rm snyk-full
```

### 3. Gosec

**Purpose:** Go security checker

**Features:**
- CWE mapping
- OWASP compliance
- Security rule enforcement
- Severity classification

**Usage:**

```bash
# Install
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run
gosec -fmt=json -out=reports/gosec-report.json ./...

# With severity filter
gosec -severity=high -confidence=medium ./...
```

### 4. Semgrep

**Purpose:** Lightweight static analysis

**Features:**
- Pattern matching
- Custom rules
- Multiple languages
- CI/CD integration

**Usage:**

```bash
# Run with Docker
docker run --rm -v $(pwd):/app returntocorp/semgrep:latest --config auto /app

# Or with CLI
semgrep --config=auto --json --output=reports/semgrep-report.json ./
```

### 5. Trivy

**Purpose:** Container and filesystem vulnerability scanner

**Features:**
- OS package scanning
- Application dependency scanning
- Misconfiguration detection
- Secret detection

**Usage:**

```bash
# Filesystem scan
trivy filesystem --severity HIGH,CRITICAL .

# Container scan
trivy image helixagent:latest

# Generate report
trivy filesystem --format json --output reports/trivy-report.json .
```

### 6. KICS

**Purpose:** Infrastructure as Code security

**Features:**
- Docker file scanning
- Kubernetes manifest scanning
- Terraform scanning
- CloudFormation scanning

**Usage:**

```bash
# Run with Docker
docker run --rm -v $(pwd):/app checkmarx/kics:latest scan -p /app -o /app/reports
```

### 7. Grype

**Purpose:** Vulnerability scanner for container images and filesystems

**Features:**
- SBOM generation
- Vulnerability matching
- Multiple data sources

**Usage:**

```bash
# Scan directory
docker run --rm -v $(pwd):/app anchore/grype:latest dir:/app

# Scan image
docker run --rm anchore/grype:latest helixagent:latest
```

---

## Running Security Scans

### Quick Scan (2 minutes)

```bash
# Run only Gosec and Semgrep
./scripts/security-scan-full.sh quick
```

### Full Scan (10-15 minutes)

```bash
# Run all security tools
./scripts/security-scan-full.sh all

# With verbose output
./scripts/security-scan-full.sh all 2>&1 | tee security-scan.log
```

### Individual Scans

```bash
# SonarQube only
./scripts/security-scan-full.sh sonarqube

# Snyk only
export SNYK_TOKEN=your_token
./scripts/security-scan-full.sh snyk

# Gosec only
./scripts/security-scan-full.sh gosec

# Semgrep only
./scripts/security-scan-full.sh semgrep

# Trivy only
./scripts/security-scan-full.sh trivy

# KICS only
./scripts/security-scan-full.sh kics

# Grype only
./scripts/security-scan-full.sh grype
```

### Scheduled Scans

Add to crontab for daily scans:

```bash
# Daily security scan at 2 AM
0 2 * * * cd /path/to/helixagent && ./scripts/security-scan-full.sh quick >> /var/log/helixagent-security.log 2>&1

# Weekly full scan on Sundays at 3 AM
0 3 * * 0 cd /path/to/helixagent && ./scripts/security-scan-full.sh all >> /var/log/helixagent-security-full.log 2>&1
```

---

## Interpreting Results

### Report Locations

All reports are saved to `reports/security/`:

```
reports/security/
├── sonarqube-report-YYYYMMDD_HHMMSS.json
├── snyk-deps-YYYYMMDD_HHMMSS.json
├── snyk-code-YYYYMMDD_HHMMSS.json
├── gosec-report-YYYYMMDD_HHMMSS.json
├── semgrep-report-YYYYMMDD_HHMMSS.json
├── trivy-fs-YYYYMMDD_HHMMSS.json
├── kics-report-YYYYMMDD_HHMMSS.json
├── grype-report-YYYYMMDD_HHMMSS.json
└── security-summary-YYYYMMDD_HHMMSS.md
```

### Severity Levels

| Level | Color | Action Required | Timeline |
|-------|-------|----------------|----------|
| Critical | 🔴 Red | Immediate | 24 hours |
| High | 🟠 Orange | Urgent | 48 hours |
| Medium | 🟡 Yellow | Planned | 1 week |
| Low | 🟢 Green | Optional | Backlog |

### SonarQube Quality Gates

**Pass Criteria:**
- 0 Critical vulnerabilities
- 0 Blocker issues
- Code coverage > 80%
- Duplicated lines < 3%

**View Results:**
```bash
open http://localhost:9000/dashboard?id=helixagent
```

### Snyk Results

**Example Output:**
```json
{
  "vulnerabilities": [
    {
      "id": "SNYK-GOLANG-GITHUBCOMEXAMPLE-123456",
      "title": "SQL Injection",
      "severity": "high",
      "cvssScore": 7.5,
      "upgradePath": ["github.com/example/package@1.2.3"]
    }
  ]
}
```

**Remediation:**
```bash
# Update vulnerable dependency
go get -u github.com/example/package@latest

# Verify fix
snyk test
```

---

## Remediation Workflow

### Step 1: Triage

```bash
# View summary
./scripts/security-scan-full.sh all
cat reports/security/security-summary-*.md
```

### Step 2: Prioritize

Address findings in this order:
1. Critical vulnerabilities (remote code execution, SQL injection)
2. High vulnerabilities (authentication bypass, privilege escalation)
3. Medium vulnerabilities (information disclosure)
4. Low vulnerabilities (best practice violations)

### Step 3: Fix

**Example: Fixing SQL Injection**

Before:
```go
query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID)
db.Query(query)
```

After:
```go
query := "SELECT * FROM users WHERE id = $1"
db.Query(query, userID)
```

### Step 4: Verify

```bash
# Re-run security scan
./scripts/security-scan-full.sh quick

# Check specific tool
./scripts/security-scan-full.sh gosec
```

### Step 5: Document

Create security fix commit:
```bash
git add .
git commit -m "security: Fix SQL injection vulnerability

- Use parameterized queries
- Add input validation
- Resolves SNYK-GOLANG-GITHUBCOMEXAMPLE-123456

Risk: High -> None"
```

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Security Scan

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run security scan
        run: ./scripts/security-scan-full.sh quick
      
      - name: Upload reports
        uses: actions/upload-artifact@v3
        with:
          name: security-reports
          path: reports/security/
```

### GitLab CI

```yaml
security_scan:
  stage: test
  script:
    - ./scripts/security-scan-full.sh quick
  artifacts:
    paths:
      - reports/security/
    expire_in: 1 week
  only:
    - merge_requests
    - main
```

### Pre-commit Hook

`.pre-commit-config.yaml`:
```yaml
repos:
  - repo: local
    hooks:
      - id: gosec
        name: Gosec Security Scanner
        entry: gosec
        language: system
        files: '\.go$'
      
      - id: semgrep
        name: Semgrep Static Analysis
        entry: semgrep
        language: system
        files: '\.(go|yaml|yml|json)$'
```

---

## Troubleshooting

### Issue: SonarQube won't start

**Symptoms:**
- Connection refused on port 9000
- Container keeps restarting

**Solution:**
```bash
# Check logs
docker logs helixagent-sonarqube

# Increase memory
docker compose -f docker/security/sonarqube/docker-compose.yml down
docker compose -f docker/security/sonarqube/docker-compose.yml up -d

# Or manually with more memory
docker run -d --name sonarqube \
  -p 9000:9000 \
  -e SONAR_ES_BOOTSTRAP_CHECKS_DISABLE=true \
  -m 4g \
  sonarqube:community
```

### Issue: Snyk authentication failed

**Symptoms:**
- "Authentication failed" error
- Empty results

**Solution:**
```bash
# Verify token
export SNYK_TOKEN=your_actual_token

# Test authentication
docker run --rm -e SNYK_TOKEN=$SNYK_TOKEN snyk/snyk-cli auth
```

### Issue: Gosec reports false positives

**Symptoms:**
- Gosec flags safe code
- Too many warnings

**Solution:**
```go
// Mark line as safe with annotation
// #nosec G101
password := "hardcoded-password"
```

Or exclude rules:
```bash
gosec -exclude=G101,G102 ./...
```

### Issue: Trivy scan is slow

**Symptoms:**
- Scan takes too long
- Timeout errors

**Solution:**
```bash
# Scan specific severity only
trivy filesystem --severity HIGH,CRITICAL .

# Skip specific directories
trivy filesystem --skip-dirs vendor,node_modules .
```

### Issue: High number of vulnerabilities

**Symptoms:**
- Overwhelmed by findings
- Don't know where to start

**Solution:**
1. Focus on Critical and High severity
2. Group by type (SQL injection, XSS, etc.)
3. Fix one category at a time
4. Use automated fixes where possible

```bash
# View only critical and high
./scripts/security-scan-full.sh all
grep -E '"severity": "(critical|high)"' reports/security/*.json
```

---

## Gosec Remediation Results (v1.3.1)

During the v1.3.1 remediation cycle, a full gosec scan of the HelixAgent codebase was performed. The results and fixes are summarized below.

### Scan Summary

| Category | Findings | Fixed | Suppressed |
|---|---|---|---|
| G306 (file permissions) | 19 | 19 | 0 |
| G104 (unhandled errors) | 3 | 3 | 0 |
| G101 (hardcoded credentials) | ~500 | 0 | All (false positives) |
| G304 (file path injection) | ~40 | 0 | Risk-assessed |
| **Total** | **568** | **22** | **546** |

### Common Finding Categories and Fixes

#### G306 -- File Permissions Too Broad

The most actionable finding category. All 19 instances were files created with `0o644` or `0o755` that should use more restrictive permissions:

```go
// BEFORE: world-readable file
os.WriteFile(reportPath, data, 0o644)

// AFTER: owner read/write only
os.WriteFile(reportPath, data, 0o600)
```

```go
// BEFORE: world-executable directory
os.MkdirAll(dataDir, 0o755)

// AFTER: owner full, group read/execute
os.MkdirAll(dataDir, 0o750)
```

**Rule of thumb:**
- Sensitive files (configs, reports, keys): `0o600`
- Data directories: `0o750`
- Log files: `0o640` (allow group read for log aggregators)

#### G104 -- Unhandled Errors

Three instances where error return values were silently discarded:

```go
// BEFORE
file.Close()

// AFTER
if err := file.Close(); err != nil {
    log.Printf("warning: failed to close file: %v", err)
}
```

#### G101 -- Hardcoded Credentials (False Positives)

Most G101 findings are false positives triggered by variable names containing words like `password`, `secret`, `key`, or `token`. Suppress after manual review:

```go
// #nosec G101 -- variable name, not a credential
const PasswordMinLength = 8
```

### File Permission Hardening Guide

When creating files or directories in Go, use the minimum permissions required:

| Use Case | File Mode | Dir Mode | Rationale |
|---|---|---|---|
| API keys, secrets | `0o600` | N/A | Owner-only access |
| Config files | `0o600` | `0o750` | Prevent unauthorized reading |
| Security reports | `0o600` | `0o750` | Sensitive scan data |
| Log files | `0o640` | `0o750` | Allow group read for aggregation |
| Public assets | `0o644` | `0o755` | Intentionally world-readable |
| Temp files | `0o600` | `0o700` | Prevent symlink attacks |

## Best Practices

### 1. Regular Scanning

- Run quick scans daily
- Run full scans weekly
- Scan before every release
- Scan dependencies on update

### 2. Defense in Depth

- Input validation
- Parameterized queries
- Output encoding
- Least privilege
- Security headers

### 3. Secure Coding

- Avoid hardcoded secrets
- Validate all inputs
- Handle errors securely
- Use crypto libraries
- Implement proper logging

### 4. Incident Response

1. Detect vulnerability
2. Assess impact
3. Develop fix
4. Test fix
5. Deploy fix
6. Verify fix
7. Document lessons learned

---

## Resources

### Documentation
- [SonarQube Documentation](https://docs.sonarqube.org/)
- [Snyk Documentation](https://docs.snyk.io/)
- [Gosec Rules](https://securego.io/docs/rules/)
- [Semgrep Registry](https://semgrep.dev/explore)

### Security Standards
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [Go Security Best Practices](https://go.dev/security/best-practices)

### Support
- Security Issues: security@helixagent.dev
- Documentation: docs/security
- Slack: #security-channel

---

## Glossary

| Term | Definition |
|------|------------|
| CWE | Common Weakness Enumeration |
| CVE | Common Vulnerabilities and Exposures |
| CVSS | Common Vulnerability Scoring System |
| SAST | Static Application Security Testing |
| SCA | Software Composition Analysis |
| SBOM | Software Bill of Materials |

---

---

## Automated Containerized Security Scanning

HelixAgent provides fully containerized, automated security scanning pipelines for Snyk and SonarQube. These pipelines run inside Docker/Podman containers for reproducibility and require no local tool installation.

### Containerized Snyk Scanning

The Snyk scanning pipeline runs all scans inside a container with pre-configured profiles.

**Docker Compose Setup:**

```bash
# Full Snyk scan (dependency + code + container)
docker compose -f docker/security/snyk/docker-compose.yml \
  --profile full run --rm snyk-full

# Dependency-only scan
docker compose -f docker/security/snyk/docker-compose.yml \
  --profile deps run --rm snyk-deps

# Code analysis only
docker compose -f docker/security/snyk/docker-compose.yml \
  --profile code run --rm snyk-code

# Container image scan
docker compose -f docker/security/snyk/docker-compose.yml \
  --profile container run --rm snyk-container
```

**Environment Variables:**

| Variable | Required | Description |
|----------|----------|-------------|
| `SNYK_TOKEN` | Yes | Snyk API token from snyk.io |
| `SNYK_ORG` | No | Organization ID for team scans |
| `SNYK_SEVERITY_THRESHOLD` | No | Minimum severity: `low`, `medium`, `high`, `critical` |

**Automated Challenge Validation:**

The Snyk scanning challenge (`challenges/scripts/snyk_automated_scanning_challenge.sh`) validates 38 test points:

- Container image availability and pull
- Authentication with Snyk API
- Dependency vulnerability scanning
- Code analysis scanning
- Container image scanning
- Report generation and format validation
- Severity threshold enforcement
- Exit code correctness for CI/CD gating

```bash
./challenges/scripts/snyk_automated_scanning_challenge.sh
```

### Containerized SonarQube Scanning

SonarQube runs as a persistent container with automated project configuration and analysis.

**Docker Compose Setup:**

```bash
# Start SonarQube server
docker compose -f docker/security/sonarqube/docker-compose.yml up -d sonarqube

# Wait for SonarQube to be ready (may take 60-90 seconds on first start)
until curl -s http://localhost:9000/api/system/status | grep -q '"status":"UP"'; do
  sleep 5
done

# Run analysis
docker compose -f docker/security/sonarqube/docker-compose.yml \
  run --rm sonar-scanner
```

**Configuration Files:**

- `docker/security/sonarqube/sonar-project.properties` -- Project analysis configuration
- `docker/security/sonarqube/docker-compose.yml` -- Container orchestration
- `docker/security/sonarqube/quality-gate.json` -- Custom quality gate rules

**Quality Gate Enforcement:**

SonarQube quality gates block builds when thresholds are exceeded:

| Metric | Threshold | Action |
|--------|-----------|--------|
| Critical vulnerabilities | 0 | Block |
| Blocker issues | 0 | Block |
| Code coverage | > 80% | Warn |
| Duplicated lines | < 3% | Warn |
| Security hotspots reviewed | 100% | Block |

**Automated Challenge Validation:**

The SonarQube scanning challenge (`challenges/scripts/sonarqube_automated_scanning_challenge.sh`) validates 45 test points:

- SonarQube container startup and health
- Project creation and configuration
- Scanner execution and analysis completion
- Quality gate evaluation
- Report export (JSON, HTML)
- Metrics API querying
- Issue categorization and severity mapping
- Custom rule validation

```bash
./challenges/scripts/sonarqube_automated_scanning_challenge.sh
```

### Integrating Automated Scanning into CI/CD

Both scanning pipelines can be integrated into the manual CI/CD workflow:

```bash
# Run both scanners as part of CI validation
make ci-security-scan

# Or individually
make ci-snyk-scan
make ci-sonarqube-scan

# Check results and fail if critical issues found
make ci-security-gate
```

**Resource Limits:**

All scanning containers respect the 30-40% host resource limit:

```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 4G
```

### Scan Result Aggregation

Security scan results from all tools are aggregated into a unified report:

```bash
# Generate aggregated security report
./scripts/security-aggregate-report.sh

# Output
reports/security/
  aggregated-summary.md      # Human-readable summary
  aggregated-results.json    # Machine-readable results
  snyk-report.json          # Raw Snyk results
  sonarqube-report.json     # Raw SonarQube results
```

---

**Next Manual:** User Manual 18 - Performance Monitoring
