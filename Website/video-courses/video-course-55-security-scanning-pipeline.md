# Video Course 55: Security Scanning Pipeline

## Course Overview

**Duration**: 2 hours
**Level**: Intermediate to Advanced
**Prerequisites**: Course 01-Fundamentals, Course 10-Security-Best-Practices, Course 18-Security-Scanning, basic Docker/Podman experience

This course covers the full security scanning pipeline for HelixAgent, including Gosec, Snyk, Trivy, Semgrep, and SonarQube. You will learn how to set up the scanning infrastructure, run scans, interpret results, manage triage workflows, and handle suppression policies.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Deploy the complete security scanning infrastructure using `docker-compose.security.yml`
2. Run individual and combined security scans across the codebase
3. Interpret scan results and prioritize findings by severity
4. Implement a triage workflow with documented suppression management
5. Integrate scans into pre-commit hooks and manual CI pipelines
6. Fix common vulnerability categories found in Go codebases

---

## Module 1: Scanning Infrastructure Setup (25 min)

### 1.1 Overview of Scanners

**Video: Scanner Landscape** (8 min)

| Scanner    | Category         | Focus Area                         |
|------------|------------------|------------------------------------|
| Gosec      | Static analysis  | Go-specific security patterns      |
| Snyk       | Dependency scan  | Known CVEs in dependencies         |
| Trivy      | Vulnerability    | Container images, filesystems, IaC |
| Semgrep    | Pattern matching | Custom rules, OWASP patterns       |
| SonarQube  | Quality + security | Code smells, bugs, vulnerabilities |

### 1.2 Docker Compose Setup

**Video: Standing Up the Infrastructure** (10 min)

```bash
# Start the security scanning infrastructure
docker compose -f docker/security/docker-compose.security.yml up -d

# Verify all services are running
docker compose -f docker/security/docker-compose.security.yml ps
```

Key services deployed:
- SonarQube server with PostgreSQL backend
- Trivy server for caching vulnerability database
- Semgrep scan runner
- Snyk CLI container

### 1.3 Configuration Files

**Video: Scanner Configuration** (7 min)

- `.gosecrc` -- Gosec rule inclusions and exclusions
- `.trivy.yaml` -- Severity filters, skip directories, vulnerability types
- `.semgrep.yml` -- Custom rule definitions
- `sonar-project.properties` -- SonarQube project binding

### Hands-On Lab 1

Deploy the security infrastructure and verify each scanner is operational:

```bash
# Start infrastructure
docker compose -f docker/security/docker-compose.security.yml up -d

# Verify SonarQube is ready (may take 60-90 seconds)
curl -s http://localhost:9000/api/system/status | grep -o '"status":"[^"]*"'

# Run a quick Gosec check
make security-scan-gosec
```

---

## Module 2: Running Security Scans (25 min)

### 2.1 Individual Scanner Execution

**Video: Running Each Scanner** (10 min)

```bash
# Gosec -- Go static security analysis
make security-scan-gosec

# Trivy -- filesystem and container scanning
make security-scan-trivy

# Semgrep -- pattern-based analysis
make security-scan-semgrep

# Snyk -- dependency vulnerability scanning
snyk test --all-projects

# SonarQube -- comprehensive quality and security
make security-scan-sonarqube
```

### 2.2 Combined Scan Execution

**Video: Full Pipeline Run** (8 min)

```bash
# Run all scanners in sequence
make security-scan

# Output locations
# reports/security/gosec-report.json
# reports/security/trivy-report.json
# reports/security/semgrep-report.json
# reports/security/snyk-report.json
# reports/security/consolidated-report.md
```

### 2.3 Scan Profiles

**Video: Configuring Scan Depth** (7 min)

- Quick scan: Gosec only (30 seconds)
- Standard scan: Gosec + Trivy + Semgrep (2-3 minutes)
- Full scan: All five scanners including SonarQube (5-10 minutes)
- IaC-focused scan: Trivy + Semgrep on Docker/K8s files only

### Hands-On Lab 2

Run a full pipeline scan and collect all reports:

```bash
# Execute full scan
make security-scan

# Count findings by severity
cat reports/security/consolidated-report.md | grep -c "CRITICAL"
cat reports/security/consolidated-report.md | grep -c "HIGH"
```

---

## Module 3: Interpreting Results (25 min)

### 3.1 Understanding Severity Levels

**Video: Severity Classification** (8 min)

| Severity | Response Time     | Examples                                    |
|----------|-------------------|---------------------------------------------|
| Critical | Immediate         | SQL injection, RCE, hardcoded secrets       |
| High     | Within 24 hours   | Unvalidated input, privilege escalation     |
| Medium   | Within sprint     | Missing error handling, weak crypto config  |
| Low      | When convenient   | Informational, style-related security hints |

### 3.2 Reading Gosec Reports

**Video: Gosec Output Analysis** (5 min)

```json
{
  "severity": "HIGH",
  "confidence": "HIGH",
  "cwe": {"id": "89", "url": "https://cwe.mitre.org/data/definitions/89.html"},
  "rule_id": "G201",
  "details": "SQL string formatting",
  "file": "internal/database/queries.go",
  "line": "42"
}
```

### 3.3 Reading Trivy Reports

**Video: Trivy Output Analysis** (5 min)

- Vulnerability entries with CVE identifiers
- Fixed version information when available
- Package path and installed version
- CVSS score and attack vector

### 3.4 Cross-Scanner Correlation

**Video: Correlating Findings** (7 min)

- Same issue may appear in multiple scanners with different identifiers
- Deduplication strategies for the consolidated report
- CWE identifiers as the common taxonomy
- Priority override when scanners disagree on severity

### Hands-On Lab 3

Analyze a real scan report and classify findings:

1. Open the consolidated report
2. Identify the top 5 findings by severity
3. Check for cross-scanner duplicates
4. Map each finding to its CWE identifier

---

## Module 4: Triage and Suppression Management (20 min)

### 4.1 Triage Workflow

**Video: Structured Triage Process** (10 min)

1. Review finding with full context (file, line, rule)
2. Classify: true positive, false positive, or accepted risk
3. True positive: create fix task with priority matching severity
4. False positive: document reason and add suppression
5. Accepted risk: document justification with approval and review date

### 4.2 Suppression Policies

**Video: Managing Suppressions** (10 min)

```go
// Gosec inline suppression with justification
// #nosec G104 -- error return value intentionally ignored for cleanup logging
defer file.Close()
```

```yaml
# .semgrep.yml suppression
rules:
  - id: custom-rule
    paths:
      exclude:
        - "vendor/"
        - "generated/"
```

```yaml
# .trivyignore
# CVE-2024-XXXXX -- accepted risk, no fix available, mitigated by network policy
CVE-2024-XXXXX
```

- All suppressions must include justification comments
- Suppressions reviewed quarterly for continued validity
- Suppression count tracked as a metric

### Hands-On Lab 4

Practice the triage workflow:

1. Select 3 findings from the scan report
2. Classify each as true positive, false positive, or accepted risk
3. For the true positive, write a fix
4. For the false positive, add a suppression with justification
5. Re-run the scanner to confirm the suppression is effective

---

## Module 5: Fixing Common Findings (20 min)

### 5.1 Hardcoded Credentials

**Video: Credential Management Fixes** (5 min)

```go
// FINDING: Hardcoded API key
apiKey := "sk-1234567890abcdef"

// FIX: Read from environment
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    return fmt.Errorf("API_KEY environment variable is required")
}
```

### 5.2 SQL Injection

**Video: Parameterized Queries** (5 min)

```go
// FINDING: String concatenation in SQL
query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", userID)

// FIX: Parameterized query
query := "SELECT * FROM users WHERE id = $1"
row := db.QueryRow(query, userID)
```

### 5.3 Insecure TLS Configuration

**Video: TLS Hardening** (5 min)

```go
// FINDING: Insecure TLS minimum version
tlsConfig := &tls.Config{}

// FIX: Enforce TLS 1.3 minimum
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS13,
}
```

### 5.4 Missing Error Handling

**Video: Error Handling Patterns** (5 min)

```go
// FINDING: Ignored error return
result, _ := riskyOperation()

// FIX: Handle and wrap errors
result, err := riskyOperation()
if err != nil {
    return fmt.Errorf("risky operation failed: %w", err)
}
```

### Hands-On Lab 5

Fix a real finding from the scan:

1. Select a HIGH severity finding from the Gosec report
2. Read the affected source code
3. Apply the appropriate fix pattern
4. Run the scanner again to verify the fix resolves the finding
5. Ensure existing tests still pass

---

## Course Summary

### Key Takeaways

1. Five scanners (Gosec, Snyk, Trivy, Semgrep, SonarQube) provide layered security coverage
2. The `docker-compose.security.yml` deploys the full scanning infrastructure
3. Severity classification drives response timelines from immediate to best-effort
4. Suppressions require documented justification and quarterly review
5. Cross-scanner correlation using CWE identifiers eliminates duplicate work
6. Common Go security fixes follow well-established patterns

### Assessment Questions

1. Which scanner is best suited for finding known CVEs in Go dependencies?
2. What is the recommended response time for a Critical severity finding?
3. How should false positives be handled in the triage workflow?
4. What is the purpose of CWE identifiers in cross-scanner correlation?
5. Write a Gosec suppression comment with proper justification.

### Related Courses

- Course 10: Security Best Practices
- Course 18: Security Scanning and Hardening
- Course 34: Penetration Testing
- Course 39: Security Testing

---

**Course Version**: 1.0
**Last Updated**: March 8, 2026
