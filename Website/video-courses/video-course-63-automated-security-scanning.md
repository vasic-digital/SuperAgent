# Video Course 63: Automated Security Scanning Pipeline

## Course Overview

**Duration:** 3 hours
**Level:** Intermediate to Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 17 (Security Basics), Docker/Podman familiarity

Master HelixAgent's containerized security scanning pipeline using Snyk and SonarQube. Learn to configure, run, and interpret automated scans, enforce quality gates, and integrate scanning into the development workflow.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Deploy and configure Snyk and SonarQube in containers
2. Run automated dependency, code, and container scans
3. Interpret scan results and prioritize remediation
4. Configure quality gates that block insecure builds
5. Write and run challenge scripts that validate scanner correctness
6. Integrate scanning into the manual CI/CD workflow

---

## Module 1: Security Scanning Fundamentals (25 min)

### Video 1.1: Why Containerized Scanning? (10 min)

**Topics:**
- The problem with locally installed scanners (version drift, configuration mismatch)
- Containerized scanning benefits: reproducibility, isolation, no host dependencies
- HelixAgent's 7-tool security stack (SonarQube, Snyk, Gosec, Semgrep, Trivy, KICS, Grype)
- Focus on Snyk and SonarQube as automated, containerized pipelines

### Video 1.2: Scanning Types Overview (15 min)

**Topics:**
- **SAST** (Static Application Security Testing): SonarQube, Gosec, Semgrep -- analyze source code
- **SCA** (Software Composition Analysis): Snyk, Trivy, Grype -- analyze dependencies
- **Container Scanning**: Snyk, Trivy -- analyze Docker images
- **IaC Scanning**: KICS -- analyze infrastructure definitions
- Severity levels: Critical, High, Medium, Low
- CVSS scoring system

---

## Module 2: Snyk Containerized Pipeline (45 min)

### Video 2.1: Setting Up Snyk (15 min)

**Topics:**
- Creating a Snyk account and API token
- Docker Compose configuration for Snyk
- Environment variable setup (`SNYK_TOKEN`, `SNYK_ORG`)
- Container resource limits for scanning

**Demo:**
```bash
# Set up Snyk token
export SNYK_TOKEN=your_token_here

# Run dependency scan
docker compose -f docker/security/snyk/docker-compose.yml \
  --profile deps run --rm snyk-deps
```

### Video 2.2: Running Snyk Scans (15 min)

**Topics:**
- Dependency scanning: `go.mod` analysis for all 27+ modules
- Code analysis: Static analysis of Go source files
- Container scanning: Analyzing the HelixAgent Docker image
- Full scan profile: Running all three scan types

**Demo: Full Scan**
```bash
docker compose -f docker/security/snyk/docker-compose.yml \
  --profile full run --rm snyk-full
```

### Video 2.3: Interpreting Snyk Results (15 min)

**Topics:**
- JSON report structure: vulnerabilities, severity, CVSS, upgrade paths
- Filtering by severity threshold
- Understanding fix recommendations
- Snyk ignore policies (`.snyk` file)
- Exit codes for CI/CD gating (0=clean, 1=vulnerabilities found)

**Example: Reading a Finding**
```json
{
  "id": "SNYK-GOLANG-GITHUBCOMJACKC-PGX5-6241800",
  "title": "SQL Injection",
  "severity": "high",
  "cvssScore": 7.5,
  "fixedIn": ["5.5.4"],
  "upgradePath": ["github.com/jackc/pgx/v5@5.5.4"]
}
```

**Remediation:**
```bash
go get github.com/jackc/pgx/v5@v5.5.4
go mod tidy
```

---

## Module 3: SonarQube Containerized Pipeline (45 min)

### Video 3.1: Deploying SonarQube (15 min)

**Topics:**
- Starting the SonarQube container
- First-time setup and health check
- Changing default credentials
- Project configuration (`sonar-project.properties`)
- Memory requirements (4GB minimum)

**Demo:**
```bash
# Start SonarQube
docker compose -f docker/security/sonarqube/docker-compose.yml up -d sonarqube

# Wait for readiness
until curl -s http://localhost:9000/api/system/status | grep -q '"status":"UP"'; do
  sleep 5
done
echo "Ready!"
```

### Video 3.2: Running SonarQube Analysis (15 min)

**Topics:**
- The sonar-scanner container
- Source exclusions (vendor, third-party)
- Coverage report integration
- Test report integration
- Analysis duration and resource usage

**Demo:**
```bash
# Run analysis
docker compose -f docker/security/sonarqube/docker-compose.yml \
  run --rm sonar-scanner

# View results
open http://localhost:9000/dashboard?id=helixagent
```

### Video 3.3: Quality Gates and Results (15 min)

**Topics:**
- Default quality gate conditions
- Custom quality gate rules
- Issue types: bugs, vulnerabilities, code smells, security hotspots
- Severity levels: blocker, critical, major, minor, info
- Technical debt estimation
- Coverage and duplication metrics

**Quality Gate Thresholds:**
| Metric | Condition | Value |
|--------|-----------|-------|
| New bugs | Greater than | 0 |
| New vulnerabilities | Greater than | 0 |
| Security hotspots reviewed | Less than | 100% |
| Coverage on new code | Less than | 80% |
| Duplicated lines on new code | Greater than | 3% |

---

## Module 4: Challenge Validation (30 min)

### Video 4.1: Snyk Scanning Challenge (15 min)

**Topics:**
- The 38-test Snyk challenge script
- What each test validates (auth, scanning, reports, exit codes)
- Running the challenge
- Fixing challenge failures

**Demo:**
```bash
./challenges/scripts/snyk_automated_scanning_challenge.sh
```

**Key Test Categories:**
1. Container availability (tests 1-5)
2. Authentication (tests 6-10)
3. Dependency scanning (tests 11-18)
4. Code analysis (tests 19-24)
5. Container scanning (tests 25-30)
6. Report validation (tests 31-35)
7. CI/CD gating (tests 36-38)

### Video 4.2: SonarQube Scanning Challenge (15 min)

**Topics:**
- The 45-test SonarQube challenge script
- Server health validation
- Analysis completion verification
- Quality gate evaluation
- API endpoint testing
- Report export validation

**Demo:**
```bash
./challenges/scripts/sonarqube_automated_scanning_challenge.sh
```

**Key Test Categories:**
1. Server startup and health (tests 1-8)
2. Project configuration (tests 9-15)
3. Analysis execution (tests 16-25)
4. Quality gate (tests 26-32)
5. API validation (tests 33-38)
6. Report export (tests 39-42)
7. Metrics collection (tests 43-45)

---

## Module 5: Integration and Automation (20 min)

### Video 5.1: Unified Security Gate (10 min)

**Topics:**
- Aggregating results from Snyk and SonarQube
- Gate logic: pass/fail decision
- Report aggregation script
- Pre-release scanning workflow

**Gate Decision Logic:**
```
PASS = (Snyk critical == 0) AND (Snyk high unpatched == 0)
       AND (SonarQube quality gate == PASS)
       AND (All scanners ran successfully)
       AND (Reports are valid JSON)
```

### Video 5.2: Scheduled and Pre-Release Scanning (10 min)

**Topics:**
- Daily quick scans with crontab
- Weekly full scans
- Pre-release gate enforcement
- Resource limits for scanning containers

---

## Module 6: Hands-On Labs (45 min)

### Lab 1: Run a Full Snyk Scan (15 min)

**Objective:** Execute a full Snyk scan and analyze the results.

**Steps:**
1. Set up `SNYK_TOKEN` environment variable
2. Run the full scan profile
3. Open the JSON report
4. Identify the highest-severity finding
5. Research the remediation path
6. Apply the fix and re-scan

### Lab 2: Deploy SonarQube and Analyze Code (15 min)

**Objective:** Deploy SonarQube, run analysis, and interpret the dashboard.

**Steps:**
1. Start SonarQube container
2. Wait for readiness
3. Run the scanner
4. Open the dashboard at localhost:9000
5. Navigate to the security hotspots view
6. Review and mark at least one hotspot

### Lab 3: Run Both Challenge Scripts (15 min)

**Objective:** Execute both scanning challenges and fix any failures.

**Steps:**
1. Run the Snyk challenge script
2. Note any failures and fix them
3. Run the SonarQube challenge script
4. Note any failures and fix them
5. Achieve 100% pass rate on both challenges

---

## Assessment

### Quiz (10 questions)

1. What is the difference between SAST and SCA?
2. How does Snyk determine if a dependency is vulnerable?
3. What quality gate conditions block a SonarQube analysis?
4. How many tests does the Snyk challenge validate?
5. What CVSS score range corresponds to "Critical" severity?
6. Why must scanning containers respect the 30-40% resource limit?
7. What is the `.snyk` file used for?
8. How does SonarQube differentiate "new code" from "overall code"?
9. What exit code does Snyk return when vulnerabilities are found?
10. How are scanning results aggregated into a unified report?

### Practical Assessment

Configure and run a complete security scanning pipeline:
1. Deploy both Snyk and SonarQube
2. Run full scans on the HelixAgent codebase
3. Generate an aggregated report
4. Identify the top 3 findings by severity
5. Write remediation steps for each finding

---

## Resources

- [User Manual 17: Security Scanning Guide](../user-manuals/17-security-scanning-guide.md)
- [User Manual 32: Automated Security Scanning](../user-manuals/32-automated-security-scanning.md)
- [Snyk Documentation](https://docs.snyk.io/)
- [SonarQube Documentation](https://docs.sonarqube.org/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Docker Compose Security Config](../../docker/security/)
