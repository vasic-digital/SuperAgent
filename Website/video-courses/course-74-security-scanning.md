# Video Course 74: Security Scanning Pipeline

## Course Overview

**Duration:** 3 hours
**Level:** Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 10 (Security Best Practices)

Build and operate HelixAgent's comprehensive security scanning pipeline. This course covers all 7 scanning tools (gosec, Trivy, Semgrep, Snyk, SonarQube, staticcheck, go vet), their Docker Compose infrastructure, running scans, interpreting reports, fixing findings, and establishing a continuous scanning workflow integrated with the development cycle.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Describe the 7 security scanning tools and their coverage areas
2. Deploy the scanning infrastructure using Docker Compose
3. Execute scans via Makefile targets and interpret JSON/HTML reports
4. Prioritize and fix findings by severity and category
5. Establish a continuous scanning workflow within the development cycle
6. Integrate scan results with Prometheus metrics and alerting

---

## Module 1: Scanner Overview (30 min)

### Video 1.1: The 7 Scanning Tools (15 min)

**Topics:**
- **gosec**: Go-specific static analysis for security issues (SQL injection, hardcoded credentials, weak crypto)
- **Trivy**: Container image vulnerability scanning (CVEs in base images and dependencies)
- **Semgrep**: Pattern-based static analysis with custom rules (OWASP Top 10)
- **Snyk**: Dependency vulnerability scanning (known CVEs in Go modules) and code analysis
- **SonarQube**: Code quality and security analysis (bugs, vulnerabilities, code smells, coverage)
- **staticcheck**: Go static analysis for correctness (beyond `go vet`)
- **go vet**: Standard Go static analysis for suspicious constructs

**Coverage Matrix:**
```
Tool          | Dependencies | Source Code | Containers | Quality |
--------------|-------------|-------------|------------|---------|
gosec         |             | X           |            |         |
Trivy         | X           |             | X          |         |
Semgrep       |             | X           |            |         |
Snyk          | X           | X           |            |         |
SonarQube     |             | X           |            | X       |
staticcheck   |             | X           |            |         |
go vet        |             | X           |            |         |
```

### Video 1.2: When to Use Each Tool (15 min)

**Topics:**
- gosec: first line of defense for Go-specific vulnerabilities
- Trivy: before pushing container images to a registry
- Semgrep: enforcing organization-specific security policies
- Snyk: continuous monitoring of dependency vulnerabilities
- SonarQube: comprehensive quality gate for pull requests
- Layered defense: no single tool catches everything
- False positive management: triage and suppression strategies

---

## Module 2: Docker Compose Setup (30 min)

### Video 2.1: Snyk Infrastructure (15 min)

**Topics:**
- `docker/security/snyk/docker-compose.yml` defines 4 scan profiles
- Services: `snyk-deps` (dependency scan), `snyk-code` (static analysis), `snyk-container`, `snyk-iac`
- Configuration: `SNYK_TOKEN` environment variable for authentication
- Volume mounts: project root as read-only, reports volume for output
- Profiles: `dependencies`, `code`, `container`, `all`

**Compose Structure:**
```yaml
services:
  snyk-deps:
    container_name: helixagent-snyk-deps
    environment:
      SNYK_TOKEN: ${SNYK_TOKEN:-}
    volumes:
      - ../../..:/app:ro
      - snyk-reports:/reports
    command: /scripts/scan-dependencies.sh
    profiles: [dependencies, all]
```

### Video 2.2: SonarQube Infrastructure (15 min)

**Topics:**
- `docker/security/sonarqube/docker-compose.yml` with SonarQube Community Edition 10.7
- Depends on PostgreSQL for persistent storage
- Health check: HTTP probe against `/api/system/status`
- Resource limits: 2 GB memory, 1.0 CPU to protect host system
- `sonar-project.properties` for project-specific configuration
- Start period: 60 seconds for JVM warm-up

---

## Module 3: Running Scans (30 min)

### Video 3.1: Makefile Targets (15 min)

**Topics:**
- `make security-scan`: run all available scanners
- `make security-scan-gosec`: Go security analysis
- `make security-scan-trivy`: container image vulnerability scan
- `make security-scan-semgrep`: pattern-based analysis via container
- `make security-scan-container`: full container security audit
- Auto-detection: Docker vs Podman runtime selection

**Commands:**
```bash
# Run individual scanners
make security-scan-gosec
make security-scan-trivy
make security-scan-semgrep

# Run all scanners
make security-scan

# Snyk via Docker Compose
cd docker/security/snyk
docker compose --profile all up

# SonarQube analysis
cd docker/security/sonarqube
docker compose up -d
sonar-scanner
```

### Video 3.2: Report Generation (15 min)

**Topics:**
- All reports written to `reports/security/` directory
- gosec: `gosec-report.json` with categorized findings
- Trivy: `trivy-report.json` with CVE details and severity
- Semgrep: `semgrep-report.json` with rule matches
- Consolidated report: `consolidated-report.md` aggregating all tools
- Parsing reports with `jq` for quick issue counts

---

## Module 4: Interpreting Reports (30 min)

### Video 4.1: Severity Classification (15 min)

**Topics:**
- CRITICAL: immediate exploitation risk (hardcoded secrets, SQL injection)
- HIGH: significant vulnerability requiring prompt attention (weak crypto, path traversal)
- MEDIUM: moderate risk with mitigating factors (information disclosure, verbose errors)
- LOW: minor issues or best practice violations (code style, unused variables)
- Mapping tool-specific severity to a unified scale

### Video 4.2: Reading Tool-Specific Output (15 min)

**Topics:**
- gosec: rule IDs (G101-G601), confidence levels, file/line references
- Trivy: CVE IDs, CVSS scores, affected packages, fixed versions
- Semgrep: rule names, OWASP categories, matched code snippets
- SonarQube: issue types (bug/vulnerability/code smell), effort estimates
- Cross-referencing findings across tools for validation

**Example gosec Finding:**
```json
{
  "severity": "HIGH",
  "confidence": "HIGH",
  "rule_id": "G101",
  "details": "Potential hardcoded credentials",
  "file": "internal/config/config.go",
  "line": "42",
  "code": "password := \"admin123\""
}
```

---

## Module 5: Fixing Findings (25 min)

### Video 5.1: Prioritization Strategy (10 min)

**Topics:**
- Fix CRITICAL and HIGH findings first
- Group related findings: one root cause may produce multiple alerts
- Distinguish true positives from false positives
- Suppression annotations: `// #nosec G101` with justification comments
- Track suppression debt: review suppressions quarterly

### Video 5.2: Common Fix Patterns (15 min)

**Topics:**
- Hardcoded credentials: move to environment variables or secret managers
- SQL injection: use parameterized queries (`$1` placeholders in pgx)
- Path traversal: validate paths with `filepath.Clean` and base directory checks
- Weak crypto: upgrade to `crypto/rand` and modern hash algorithms
- Dependency CVEs: upgrade to fixed versions, or pin to unaffected versions
- Container CVEs: update base images and minimize image layers

**Fix Example:**
```go
// BEFORE (gosec G101: hardcoded credential)
password := "admin123"

// AFTER
password := os.Getenv("DB_PASSWORD")
if password == "" {
    return fmt.Errorf("DB_PASSWORD environment variable required")
}
```

---

## Module 6: Continuous Scanning Workflow (15 min)

### Video 6.1: Integrating Scans into Development (15 min)

**Topics:**
- Pre-commit: run `make security-scan-gosec` and `go vet` before commits
- Pre-release: run full scan suite including Trivy and Snyk
- Scheduled: weekly SonarQube analysis for trend tracking
- Challenge validation: `snyk_automated_scanning_challenge.sh` (38 tests)
- Challenge validation: `sonarqube_automated_scanning_challenge.sh` (45 tests)
- Prometheus metrics: export scan result counts for dashboard monitoring
- Alert thresholds: trigger notifications when new CRITICAL findings appear

**Workflow:**
```
Code Change --> gosec + go vet (fast, local)
    --> Commit --> Semgrep + Snyk (thorough, containerized)
        --> Release --> Trivy (container scan) + SonarQube (quality gate)
            --> Deploy --> Scheduled weekly full scan
```

---

## Assessment

### Quiz (10 questions)

1. What are the 7 scanning tools in HelixAgent's security pipeline?
2. Which tool scans container images for CVEs?
3. How does Snyk differ from gosec in terms of scanning scope?
4. What Docker Compose profile runs all Snyk scan types?
5. Where are security scan reports written?
6. What is the SonarQube health check endpoint?
7. How do you suppress a gosec finding with justification?
8. What resource limits are set for the SonarQube container?
9. When should Trivy be run in the development workflow?
10. How do you consolidate findings from multiple tools?

### Practical Assessment

Set up and operate the complete scanning pipeline:
1. Deploy Snyk and SonarQube infrastructure via Docker Compose
2. Run all 7 scanners against the HelixAgent codebase
3. Interpret the consolidated report and classify 5 findings by severity
4. Fix 2 findings (one dependency, one source code)
5. Re-run scans and verify the findings are resolved

Deliverables:
1. Docker Compose deployment logs showing healthy services
2. Consolidated security report with finding counts per tool
3. Before/after code diffs for the 2 fixed findings
4. Re-scan report confirming resolution

---

## Resources

- [Snyk Docker Compose](../../docker/security/snyk/docker-compose.yml)
- [SonarQube Docker Compose](../../docker/security/sonarqube/docker-compose.yml)
- [Security Scanning Challenge](../../challenges/scripts/snyk_automated_scanning_challenge.sh)
- [SonarQube Challenge](../../challenges/scripts/sonarqube_automated_scanning_challenge.sh)
- [Security Scanning Guide](../user-manuals/17-security-scanning-guide.md)
- [Course 10: Security Best Practices](course-10-security-best-practices.md)
