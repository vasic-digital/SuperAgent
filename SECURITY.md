# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

1. **DO NOT** create a public GitHub issue for security vulnerabilities
2. Email security concerns to: security@helix.dev (or contact the maintainer directly)
3. Include as much detail as possible:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- **Acknowledgment**: Within 48 hours of your report
- **Initial Assessment**: Within 1 week
- **Resolution Timeline**: Depends on severity
  - Critical: 24-72 hours
  - High: 1 week
  - Medium: 2-4 weeks
  - Low: Next release cycle

### Disclosure Policy

- We follow coordinated disclosure
- Credit will be given to reporters (unless anonymity is requested)
- Public disclosure occurs after a fix is available

## Security Measures

### Code Security

- **Static Analysis**: Gosec, CodeQL, golangci-lint
- **Dependency Scanning**: Snyk, Trivy, Dependabot
- **Container Security**: Trivy, Hadolint
- **Secrets Detection**: TruffleHog
- **SBOM Generation**: Syft (CycloneDX, SPDX formats)

### Runtime Security

- **Authentication**: JWT with RS256/HS256, API key validation
- **Authorization**: Role-based access control (RBAC)
- **Rate Limiting**: Token bucket and sliding window algorithms
- **Input Validation**: Strict validation on all API inputs
- **Output Sanitization**: XSS prevention, content filtering
- **TLS**: Enforced for all production traffic

### LLM-Specific Security

- **Prompt Injection Protection**: Input sanitization and guardrails
- **PII Detection**: Automatic detection and redaction
- **Content Filtering**: Configurable guardrails for harmful content
- **System Prompt Protection**: Isolation and validation
- **Red Team Framework**: Built-in security testing tools

### Infrastructure Security

- **Container Isolation**: All services run in isolated containers
- **Network Segmentation**: Internal services not exposed externally
- **Secrets Management**: Environment-based, no hardcoded credentials
- **Circuit Breakers**: Fault tolerance for external dependencies
- **Health Monitoring**: Continuous health checks on all services

## Security Testing

### Automated Scanning

```bash
# Run all security scans
make security-scan

# Individual scanners
make security-scan-gosec      # Go security checker
make security-scan-snyk       # Dependency vulnerabilities
make security-scan-trivy      # Container/filesystem scanning
make security-scan-sonarqube  # Code quality and security
```

### Security Tests

```bash
# Security test suites
make test-security            # All security tests
make test-type-security       # Security tests with infrastructure

# Challenge-based validation
./challenges/scripts/security_scanning_challenge.sh
./challenges/scripts/jwt_security_challenge.sh
./challenges/scripts/sql_injection_challenge.sh
./challenges/scripts/xss_prevention_challenge.sh
./challenges/scripts/csrf_protection_challenge.sh
```

### SBOM Generation

```bash
make sbom                     # Generate Software Bill of Materials
```

## Security Configuration

### Gosec Configuration

Located at `.gosec.yml`:
- Configures rule exclusions for documented false positives
- Excludes test fixtures and development-only code
- Full justification for each exclusion

### Snyk Configuration

Located at `.snyk`:
- Dependency analysis policy
- Vulnerability severity thresholds
- Auto-fix preferences

### SonarQube Configuration

Located at `sonar-project.properties`:
- Code quality gates
- Security hotspot detection
- Coverage requirements

## Secure Development Guidelines

### For Contributors

1. **Never commit secrets** - Use environment variables
2. **Validate all inputs** - Trust no external data
3. **Use parameterized queries** - Prevent SQL injection
4. **Sanitize outputs** - Prevent XSS
5. **Keep dependencies updated** - Run `go mod tidy` regularly
6. **Run security scans locally** - Before submitting PRs

### Pre-commit Checks

```bash
# Install pre-commit hooks
make install-hooks

# Manual security check before commit
make ci-pre-commit
```

## Vulnerability Response Process

1. **Triage**: Assess severity and impact
2. **Containment**: Implement temporary mitigations if needed
3. **Fix Development**: Create and test the fix
4. **Review**: Security review of the fix
5. **Deployment**: Roll out fix to all environments
6. **Communication**: Notify affected parties
7. **Post-Mortem**: Document lessons learned

## Third-Party Dependencies

### Audited Dependencies

All dependencies are:
- Scanned weekly via Dependabot and Snyk
- Reviewed for known vulnerabilities
- Updated promptly for security fixes

### Submodule Security

Third-party submodules (`cli_agents/`, `MCP/`) are:
- Read-only - no changes pushed
- Regularly updated to latest versions
- Excluded from certain security scans (as external code)

## Contact

For security-related inquiries:
- Create a private security advisory on GitHub
- Contact the maintainer directly

Thank you for helping keep HelixAgent secure.
