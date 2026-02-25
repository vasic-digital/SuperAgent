# Course 18: Security Scanning and Hardening

## Course Overview

**Duration:** 50 minutes  
**Level:** Advanced  
**Prerequisites:** Course 01-Fundamentals, Course 10-Security-Best-Practices

Master the comprehensive security scanning capabilities of HelixAgent including Semgrep, KICS, Trivy, and SonarQube integration.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Run comprehensive security scans
2. Configure custom security rules
3. Integrate scanning into CI/CD pipelines
4. Remediate common vulnerabilities
5. Generate security reports
6. Implement continuous security monitoring

---

## Module 1: Security Scanning Overview (5 min)

### Available Scanners

| Scanner | Purpose | Type |
|---------|---------|------|
| **Gosec** | Go security analysis | Static |
| **Trivy** | Container/FS vulnerabilities | Dynamic |
| **Snyk** | Dependency scanning | Dynamic |
| **Semgrep** | Pattern-based analysis | Static |
| **KICS** | IaC scanning | Static |
| **Grype** | Vulnerability scanning | Dynamic |
| **SonarQube** | Code quality + security | Static |

### Scan Workflow

```
Code Change → Pre-commit Hooks → CI Scan → Report → Remediate → Verify
```

---

## Module 2: Running Security Scans (10 min)

### Basic Scan

```bash
# Run all security scanners
make security-scan

# Run specific scanner
make security-scan-gosec
make security-scan-trivy
make security-scan-semgrep
make security-scan-kics
```

### Using Docker Compose

```bash
# Start security infrastructure
docker compose -f docker-compose.security.yml up -d

# Run scanners
docker compose -f docker-compose.security.yml --profile scan run --rm semgrep-scanner
docker compose -f docker-compose.security.yml --profile scan run --rm kics-scanner
```

### Scan Reports

Reports are generated in `reports/security/`:

```
reports/security/
├── gosec-report.json
├── trivy-report.json
├── semgrep-report.json
├── results.json (KICS)
├── grype-report.json
└── consolidated-report.md
```

---

## Module 3: Semgrep Configuration (10 min)

### Custom Rules

Create `.semgrep.yml`:

```yaml
rules:
  - id: custom-hardcoded-api-key
    message: Potential hardcoded API key
    severity: ERROR
    languages: [go]
    pattern-either:
      - pattern: $VAR = "sk-..."
      - pattern: $VAR = "api_key..."
  
  - id: custom-sql-injection
    message: Potential SQL injection
    severity: ERROR
    languages: [go]
    pattern: fmt.Sprintf("SELECT ... %s ...", $VAR)
```

### Running with Custom Rules

```bash
# Run with custom rules
semgrep --config .semgrep.yml --json --output reports/security/semgrep-custom.json
```

### Pre-built Rule Sets

```bash
# OWASP Top 10
semgrep --config p/owasp-top-ten

# Security best practices
semgrep --config p/security-audit

# Go-specific
semgrep --config p/golang
```

---

## Module 4: Infrastructure as Code Scanning (10 min)

### KICS Configuration

Scan Docker, Kubernetes, Terraform:

```bash
# Scan all IaC files
docker run --rm -v $(pwd):/app checkmarx/kics:latest \
  scan -p /app -o /reports --report-formats json
```

### Common IaC Issues

| Issue | Severity | Fix |
|-------|----------|-----|
| Hardcoded secrets | High | Use environment variables |
| Privileged containers | High | Remove privileged flag |
| Missing resource limits | Medium | Add CPU/memory limits |
| Root user | Medium | Set non-root user |
| Unencrypted volumes | Medium | Enable encryption |

### Trivy Configuration

Create `.trivy.yaml`:

```yaml
severity: HIGH,CRITICAL
ignore-unfixed: true
skip-dirs:
  - vendor
  - node_modules
vulnerability:
  type: os,library
```

---

## Module 5: Integrating with CI/CD (10 min)

### Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
make security-scan-gosec
if [ $? -ne 0 ]; then
    echo "Security scan failed. Fix issues before committing."
    exit 1
fi
```

### GitHub Actions (Manual)

```yaml
name: Security Scan
on: workflow_dispatch

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Gosec
        run: make security-scan-gosec
      - name: Run Semgrep
        run: make security-scan-semgrep
      - name: Run Trivy
        run: make security-scan-trivy
      - name: Upload Reports
        uses: actions/upload-artifact@v4
        with:
          name: security-reports
          path: reports/security/
```

### Makefile Integration

```makefile
ci-security:
	@echo "Running CI security checks..."
	@make security-scan-gosec
	@make security-scan-semgrep
	@make security-scan-trivy
	@make security-report
```

---

## Module 6: Remediation Strategies (5 min)

### Priority-Based Remediation

1. **Critical** - Fix immediately, block deployment
2. **High** - Fix within 24 hours
3. **Medium** - Fix within sprint
4. **Low** - Fix when possible

### Common Fixes

#### SQL Injection

```go
// BAD
query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID)

// GOOD
query := "SELECT * FROM users WHERE id = ?"
db.Query(query, userID)
```

#### Hardcoded Credentials

```go
// BAD
apiKey := "sk-1234567890abcdef"

// GOOD
apiKey := os.Getenv("API_KEY")
```

#### Missing Error Handling

```go
// BAD
result, _ := someFunction()

// GOOD
result, err := someFunction()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

---

## Hands-on Lab

### Exercise 1: Full Security Audit

Run a complete security audit:

1. Start security infrastructure
2. Run all scanners
3. Review consolidated report
4. Fix all Critical and High issues
5. Re-run to verify

### Exercise 2: Custom Semgrep Rules

Create custom rules for your project:

1. Identify project-specific patterns
2. Create `.semgrep.yml`
3. Test rules
4. Add to CI/CD

---

## Quiz

1. Which scanner is best for Go security analysis?
2. How do you configure Trivy severity levels?
3. What is the recommended remediation time for Critical issues?
4. Which scanner handles Infrastructure as Code?

---

## Resources

- [Security Scanning Documentation](../SECURITY.md)
- [Semgrep Documentation](https://semgrep.dev/docs/)
- [KICS Documentation](https://docs.kics.io/)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)

---

**Next Course:** Course 19 - Performance Monitoring and Observability
