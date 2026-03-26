# HelixAgent Security Documentation

## Overview

HelixAgent implements defense-in-depth security with automated scanning, runtime guardrails, and secure-by-default configurations.

## Contents

- [Scanning Guide](scanning-guide.md) — Running Snyk, SonarQube, gosec, and Trivy scans
- [Vulnerability Disclosure](vulnerability-disclosure.md) — How to report security issues
- [Best Practices](best-practices.md) — Go security best practices applied in HelixAgent
- [Threat Model](threat-model.md) — System threat model and mitigations

## Related Documentation

- [Security Scanning Setup](../SECURITY_SCANNING.md) — Detailed scanner configuration
- [Security Hardening Guide](../guides/security-hardening.md) — Production hardening
- [PII Detection](../features/pii-detection.md) — Content filtering and redaction

## Quick Start

```bash
# Run all security scanners
make security-scan-all

# Run specific scanner
make security-scan-snyk
make security-scan-sonarqube
make security-scan-gosec
```
