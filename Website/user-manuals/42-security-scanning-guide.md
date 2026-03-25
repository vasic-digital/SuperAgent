# User Manual 42: Security Scanning Guide

## Overview

HelixAgent integrates 7 security scanners that run inside containers for reproducible, automated vulnerability detection. This guide walks through starting the security infrastructure, running each scanner, interpreting reports, fixing findings, and re-scanning.

## Prerequisites

- Docker 24.0+ or Podman 4.0+
- HelixAgent source repository cloned
- At least 4 GB free RAM for scanner containers
- (Optional) Snyk API token for enhanced vulnerability database access

## Step 1: Start Security Containers

Launch all security scanners with the dedicated compose file:

```bash
docker compose -f docker-compose.security.yml up -d
```

Verify all containers are running:

```bash
docker compose -f docker-compose.security.yml ps
```

For Podman:

```bash
podman-compose -f docker-compose.security.yml up -d
```

## Step 2: Run Snyk (Dependency and Code Analysis)

Snyk scans Go dependencies for known vulnerabilities:

```bash
docker compose -f docker/security/snyk/docker-compose.yml up
```

Review the output for severity levels: `critical`, `high`, `medium`, `low`. Snyk reports include CVE identifiers, affected packages, and recommended upgrade paths.

Key findings to prioritize:
- **Critical/High**: Fix immediately -- these have known exploits
- **Medium**: Fix in the next release cycle
- **Low**: Track and address during maintenance windows

## Step 3: Run SonarQube (Code Quality and Security)

SonarQube performs static analysis for code smells, bugs, and security hotspots:

```bash
docker compose -f docker/security/sonarqube/docker-compose.yml up -d

# Wait for SonarQube to initialize (first run takes 2-3 minutes)
# Access the dashboard at http://localhost:9000 (default: admin/admin)
```

Trigger a scan:

```bash
# From the project root
docker run --rm -v "$(pwd):/usr/src" \
  sonarsource/sonar-scanner-cli \
  -Dsonar.projectKey=helixagent \
  -Dsonar.sources=. \
  -Dsonar.host.url=http://localhost:9000
```

## Step 4: Run Trivy (Container Image Scanning)

Trivy scans container images and filesystem for vulnerabilities:

```bash
# Scan the HelixAgent container image
docker run --rm aquasec/trivy image helixagent:latest

# Scan the filesystem
docker run --rm -v "$(pwd):/project" aquasec/trivy fs /project

# Output as JSON for automated processing
docker run --rm aquasec/trivy image --format json helixagent:latest > trivy-report.json
```

## Step 5: Run Gosec (Go Security Linter)

Gosec identifies security issues specific to Go source code:

```bash
# Via make target
make security-scan

# Or directly
docker run --rm -v "$(pwd):/app" securego/gosec -fmt json /app/...
```

Common findings: hardcoded credentials, SQL injection vectors, insecure random number generation, unhandled errors.

## Step 6: Run Semgrep (Pattern-Based Analysis)

Semgrep finds bugs and security issues using semantic pattern matching:

```bash
docker run --rm -v "$(pwd):/src" returntocorp/semgrep \
  semgrep scan --config auto /src
```

Semgrep rules cover: injection vulnerabilities, authentication bypass, insecure deserialization, and OWASP Top 10 patterns.

## Step 7: Run KICS (Infrastructure as Code Scanning)

KICS scans Docker, Kubernetes, and Terraform configurations:

```bash
docker run --rm -v "$(pwd):/project" checkmarx/kics scan \
  -p /project/docker -p /project/configs \
  --output-path /project/kics-report
```

Focus areas: exposed ports, missing resource limits, privileged containers, missing health checks.

## Step 8: Run Grype (SBOM Vulnerability Matching)

Grype matches software bill of materials against vulnerability databases:

```bash
# Scan container image
docker run --rm anchore/grype helixagent:latest

# Scan a directory
docker run --rm -v "$(pwd):/project" anchore/grype dir:/project

# Generate SBOM first, then scan
docker run --rm anchore/syft helixagent:latest -o json > sbom.json
docker run --rm -v "$(pwd):/data" anchore/grype sbom:/data/sbom.json
```

## Step 9: Interpret and Fix Findings

Aggregate results across all 7 scanners and prioritize:

| Priority | Action | Timeline |
|----------|--------|----------|
| Critical | Patch immediately, redeploy | Same day |
| High | Fix and test | Within 3 days |
| Medium | Schedule fix | Next sprint |
| Low | Track in backlog | Maintenance window |

Common fix patterns:
- **Dependency vulnerability**: Update `go.mod`, run `go mod tidy`, rebuild
- **Code vulnerability**: Apply the fix, add a regression test, verify with `make security-scan`
- **Container vulnerability**: Rebuild from updated base image
- **Configuration issue**: Update compose/Kubernetes manifests

## Step 10: Re-Scan After Fixes

After applying fixes, re-run the relevant scanner to confirm remediation:

```bash
# Re-run all scanners
docker compose -f docker-compose.security.yml up

# Run the challenge to validate zero critical/high findings
./challenges/scripts/snyk_automated_scanning_challenge.sh
./challenges/scripts/sonarqube_automated_scanning_challenge.sh
```

## Troubleshooting

- **SonarQube fails to start**: Increase `vm.max_map_count`: `sysctl -w vm.max_map_count=262144`
- **Snyk rate limited**: Set `SNYK_TOKEN` in your environment for authenticated access
- **Trivy database update slow**: Use `--skip-db-update` after initial download for faster subsequent scans
- **Scanner container OOM**: Increase container memory limits in the compose file

## Related Resources

- [User Manual 10: Security Hardening](10-security-hardening.md) -- Security configuration best practices
- [User Manual 32: Automated Security Scanning](32-automated-security-scanning.md) -- Snyk/SonarQube pipeline details
- Challenge scripts: `challenges/scripts/snyk_automated_scanning_challenge.sh`, `challenges/scripts/sonarqube_automated_scanning_challenge.sh`
