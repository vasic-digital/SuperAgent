# Lab 17: Security Scanning

## Objective
Execute Snyk and SonarQube security scans using containerized infrastructure, analyze findings, and apply fixes.

## Prerequisites
- Docker or Podman available
- HelixAgent built and source code accessible

## Exercise 1: Run Semgrep Scan

```bash
semgrep scan --config auto --severity ERROR --severity WARNING \
  --include="internal/**/*.go" --json > /tmp/semgrep-results.json

# Count findings
cat /tmp/semgrep-results.json | python3 -c "
import json, sys
data = json.load(sys.stdin)
results = data.get('results', [])
print(f'Total findings: {len(results)}')
for r in results:
    sev = r.get('extra',{}).get('severity','?')
    print(f'  [{sev}] {r[\"check_id\"]} @ {r[\"path\"]}:{r[\"start\"][\"line\"]}')"
```

**Expected:** Review each finding. Fix ERROR/CRITICAL, document WARNING false positives.

## Exercise 2: Run Snyk Container Scan

```bash
# Start Snyk scanning infrastructure
docker compose -f docker/security/snyk/docker-compose.yml --profile all up

# Check reports
ls reports/snyk-*.json
cat reports/snyk-deps.json | python3 -m json.tool | head -50
```

## Exercise 3: Run SonarQube Analysis

```bash
# Start SonarQube
docker compose -f docker/security/sonarqube/docker-compose.yml up -d

# Wait for healthy
curl -s http://localhost:9000/api/system/status

# Generate coverage
go test -coverprofile=coverage.out ./internal/...

# Run scanner
docker compose -f docker/security/sonarqube/docker-compose.yml --profile scanner run sonar-scanner
```

**Access dashboard:** http://localhost:9000/dashboard?id=helixagent

## Exercise 4: Fix a Finding

For each finding:
1. Read the description and understand the vulnerability
2. Check if it's a real issue or false positive
3. If real: fix the code, run tests, re-scan
4. If false positive: add `nosemgrep` or `.snyk` policy with justification

## Assessment Questions
1. What is the difference between SAST and SCA scanning?
2. When is `nosemgrep` appropriate vs. actually fixing the code?
3. Why must TLS configs always specify MinVersion?
4. How do you handle dependency vulnerabilities with no available patch?
