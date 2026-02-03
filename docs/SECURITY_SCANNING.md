# Security Scanning Guide

This document describes the security scanning infrastructure and best practices for HelixAgent.

## Overview

HelixAgent implements a multi-layered security scanning approach:

1. **Static Analysis** - Code-level security checks
2. **Dependency Scanning** - Vulnerability detection in dependencies
3. **Container Security** - Docker image and runtime security
4. **Secrets Detection** - Prevent credential leaks
5. **Dynamic Testing** - Runtime security validation

## Security Tools

### 1. Gosec (Go Security Checker)

**Purpose**: Static analysis specifically for Go code security issues.

**Configuration**: `.gosec.yml`

```bash
# Run locally
make security-scan-gosec

# Or directly
gosec -conf .gosec.yml ./...
```

**Key Rules Checked**:
- G101: Hardcoded credentials
- G102: Bind to all interfaces
- G104: Unhandled errors
- G107: URL manipulation
- G201-G204: SQL injection
- G301-G307: File permissions
- G401-G405: Cryptographic issues
- G501-G505: Blocklisted imports
- G601: Implicit memory aliasing

**Exclusions** (documented in `.gosec.yml`):
- G404: Math/rand for non-crypto (jitter, shuffling)
- G115: Safe integer conversions with bounds checking
- G101: False positives in test fixtures
- G402: Development-only TLS configurations

### 2. Snyk

**Purpose**: Dependency vulnerability scanning and license compliance.

**Configuration**: `.snyk`

```bash
# Run locally (requires SNYK_TOKEN)
make security-scan-snyk

# Or in CI
snyk test --all-projects
```

**Features**:
- Known vulnerability detection (CVE database)
- License compliance checking
- Remediation suggestions
- Continuous monitoring

### 3. Trivy

**Purpose**: Comprehensive vulnerability scanner for code, dependencies, and containers.

**Configuration**: Command-line flags in Makefile

```bash
# Filesystem scan
make security-scan-trivy

# Container image scan
trivy image helixagent:latest

# Configuration scan
trivy config .
```

**Scan Types**:
- `fs`: Filesystem (source code, dependencies)
- `image`: Container images
- `config`: IaC misconfigurations
- `rootfs`: Root filesystem
- `sbom`: Software Bill of Materials

### 4. SonarQube

**Purpose**: Comprehensive code quality and security analysis.

**Configuration**: `sonar-project.properties`

```bash
# Start SonarQube server
make start-sonarqube

# Run analysis
make security-scan-sonarqube
```

**Features**:
- Security hotspots
- Code smells
- Bug detection
- Coverage integration
- Technical debt tracking

### 5. Hadolint

**Purpose**: Dockerfile linting and best practices.

**Configuration**: `.hadolint.yaml`

```bash
# Run locally
hadolint docker/Dockerfile

# With config
hadolint --config .hadolint.yaml docker/Dockerfile
```

**Key Rules**:
- DL3000-DL3059: Dockerfile best practices
- SC1000-SC2999: Shell script issues in RUN commands

### 6. CodeQL

**Purpose**: Semantic code analysis for security vulnerabilities.

**Configuration**: `.github/workflows/security.yml`

Runs automatically in GitHub Actions with:
- Security-extended query suite
- Go-specific security checks
- SARIF output to GitHub Security tab

### 7. TruffleHog

**Purpose**: Secrets detection in code and git history.

**Configuration**: Command-line in CI workflow

```bash
# Scan current directory
trufflehog filesystem .

# Scan git history
trufflehog git file://. --since-commit HEAD~10
```

### 8. detect-secrets

**Purpose**: Pre-commit secrets detection.

**Configuration**: `.secrets.baseline`, `.pre-commit-config.yaml`

```bash
# Generate baseline
detect-secrets scan > .secrets.baseline

# Audit findings
detect-secrets audit .secrets.baseline
```

## CI/CD Integration

### GitHub Actions Workflow

The security workflow (`.github/workflows/security.yml`) runs:

| Job | Trigger | Output |
|-----|---------|--------|
| gosec | push, PR | SARIF to Security tab |
| snyk | push, PR | SARIF to Security tab |
| trivy | push, PR | SARIF to Security tab |
| trivy-container | main push | SARIF to Security tab |
| hadolint | push, PR | SARIF to Security tab |
| codeql | push, PR | Security tab |
| secrets-scan | push, PR | Workflow fail on secrets |
| dependency-review | PR only | PR comment |
| sbom | main push | Artifact upload |

### Dependabot

Configured in `.github/dependabot.yml`:

- **Go modules**: Weekly updates, grouped by minor/patch
- **GitHub Actions**: Weekly updates
- **Docker**: Weekly updates
- **Submodules**: Weekly updates

## Local Development

### Pre-commit Hooks

Install and configure pre-commit:

```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

Hooks configured:
- Go formatting (gofmt, goimports)
- Go vetting (go vet)
- Secrets detection (detect-secrets)
- Dockerfile linting (hadolint)
- YAML linting (yamllint)
- Shell linting (shellcheck)

### Running All Scans

```bash
# Run all security scans
make security-scan

# Individual scans
make security-scan-gosec
make security-scan-snyk
make security-scan-trivy
make security-scan-sonarqube
make security-scan-go

# Stop security services
make security-scan-stop
```

### Security Reports

Reports are generated in `reports/security/`:

```
reports/security/
├── gosec-report.json         # Gosec findings
├── gosec-report.html         # Gosec HTML report
├── snyk-report.json          # Snyk findings
├── trivy-report.json         # Trivy findings
├── go-analysis.txt           # Go vet/staticcheck
├── sonar-scan.log            # SonarQube scan log
└── SECURITY_SCAN_SUMMARY.md  # Combined summary
```

## Security Testing

### Unit Tests

Security-focused unit tests in `tests/security/`:

```bash
make test-security
```

Tests cover:
- Input validation (SQL injection, XSS, path traversal)
- Authentication (JWT validation, API keys)
- Authorization (RBAC, permissions)
- Cryptographic operations

### Penetration Tests

LLM-specific security tests:

```bash
go test -v ./tests/security/penetration_test.go -tags=security
```

Tests cover:
- Prompt injection
- Jailbreaking attempts
- System prompt leakage
- Data exfiltration
- Model manipulation

### Challenge Scripts

Security validation challenges:

```bash
# Security infrastructure validation
./challenges/scripts/security_scanning_challenge.sh

# JWT security
./challenges/scripts/jwt_security_challenge.sh

# Injection protection
./challenges/scripts/sql_injection_challenge.sh
./challenges/scripts/xss_prevention_challenge.sh

# CSRF protection
./challenges/scripts/csrf_protection_challenge.sh
```

## SBOM Generation

Software Bill of Materials for supply chain security:

```bash
make sbom
```

Generates:
- CycloneDX format (JSON)
- SPDX format (JSON)
- Go module dependency graph

Output in `reports/sbom/`.

## Vulnerability Response

### Severity Levels

| Severity | Response Time | Action |
|----------|---------------|--------|
| Critical | 24-72 hours | Immediate patch, notify users |
| High | 1 week | Prioritized fix |
| Medium | 2-4 weeks | Scheduled fix |
| Low | Next release | Track and fix |

### Response Process

1. **Triage**: Assess severity and impact
2. **Containment**: Temporary mitigations if needed
3. **Fix**: Develop and test the fix
4. **Review**: Security review of changes
5. **Deploy**: Roll out to all environments
6. **Communicate**: Notify affected parties
7. **Post-mortem**: Document lessons learned

## Best Practices

### For Developers

1. **Run security scans before commits**
   ```bash
   make ci-pre-commit
   ```

2. **Never commit secrets**
   - Use environment variables
   - Use `.env` files (gitignored)
   - Use secret management services

3. **Keep dependencies updated**
   ```bash
   go mod tidy
   go get -u ./...
   ```

4. **Review security warnings**
   - Don't ignore warnings
   - Document false positives in `.gosec.yml`
   - Fix issues promptly

5. **Write security tests**
   - Test input validation
   - Test authentication/authorization
   - Test error handling

### For Reviewers

1. **Check for new dependencies**
   - Review license compatibility
   - Check for known vulnerabilities
   - Verify necessity

2. **Look for security anti-patterns**
   - Hardcoded credentials
   - SQL/command injection
   - Improper error handling
   - Missing input validation

3. **Verify test coverage**
   - Security tests for new features
   - Edge case handling
   - Error path testing

## Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://golang.org/doc/security)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)

## Contact

For security concerns, see [SECURITY.md](../SECURITY.md).
