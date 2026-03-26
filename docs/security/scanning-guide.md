# Security Scanning Guide

## Supported Scanners

| Scanner | Purpose | Config File |
|---------|---------|-------------|
| Gosec | Go source code security | `.gosec.yml` |
| Snyk | Dependency vulnerabilities | `.snyk` |
| SonarQube | Code quality + security | `sonar-project.properties` |
| Trivy | Container vulnerability scanning | N/A |

## Running Scans

### All Scanners
```bash
make security-scan-all
```

### Individual Scanners

**Gosec** — Go security rules:
```bash
make security-scan-gosec
```

**Snyk** — Dependency scanning (containerized):
```bash
make security-scan-snyk
# Or via Docker directly:
cd docker/security/snyk && docker compose --profile all up
```

**SonarQube** — Code quality analysis (containerized):
```bash
make security-scan-sonarqube
# SonarQube UI available at http://localhost:9000
```

**Trivy** — Container image scanning:
```bash
make security-scan-trivy
```

## Exclusions

See `.gosec.yml` for Go-specific exclusions (e.g., G404 for retry jitter randomness).
See `.snyk` for dependency-level ignores with expiration dates.
See `sonar-project.properties` for SonarQube exclusion patterns.

## CI Integration

All scans can be run via `make security-scan-all`. Reports are output to `reports/`.

## Resolution Process

1. Run scan
2. Classify findings: Critical / High / Medium / Low / False Positive
3. Fix actionable items
4. Document suppressions in `docs/security/scan-resolutions.md`
5. Re-run scan to verify
