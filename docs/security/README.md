# Security Documentation

This directory contains security documentation for HelixAgent, including sandboxing guides, security scanning procedures, and vulnerability management.

## Overview

HelixAgent implements a comprehensive security model with defense-in-depth principles, including process isolation, container sandboxing, input validation, and multi-tool security scanning.

## Documentation Index

| Document | Description |
|----------|-------------|
| [SANDBOXING.md](SANDBOXING.md) | Security sandboxing for plugin execution, tool operations, and external integrations |
| [../SECURITY_SCANNING.md](../SECURITY_SCANNING.md) | Security scanning infrastructure and best practices |

## Security Model

### Defense in Depth

HelixAgent employs a multi-layered security approach:

```
+--------------------------------------------------+
|              Application Layer                    |
|  - Input validation                               |
|  - Authentication & Authorization                 |
|  - Rate limiting                                  |
+--------------------------------------------------+
|              Sandbox Layer                        |
|  - Process isolation                              |
|  - Resource limits                                |
|  - Capability restrictions                        |
+--------------------------------------------------+
|              Container Layer                      |
|  - Docker/Podman isolation                        |
|  - Network namespaces                             |
|  - Filesystem restrictions                        |
+--------------------------------------------------+
|              System Layer                         |
|  - SELinux/AppArmor profiles                      |
|  - Kernel security modules                        |
|  - Audit logging                                  |
+--------------------------------------------------+
```

### Security Principles

1. **Least Privilege**: Components receive only required permissions
2. **Isolation**: Plugins and tools run in isolated environments
3. **Validation**: All inputs validated before processing
4. **Audit**: All security-relevant operations logged
5. **Fail Secure**: Failures default to denying access

## Sandboxing Guide

### Process Isolation

Plugins and tools can run in isolated processes with:
- Linux namespaces
- Seccomp filtering
- Cgroups resource limits

### Container Isolation

For maximum isolation, Docker/Podman containers provide:
- Read-only filesystems
- Dropped capabilities
- Resource limits (memory, CPU, PIDs)
- Network isolation

### Configuration Example

```go
type ProcessSandbox struct {
    Enabled         bool
    UseNamespaces   bool
    UseSeccomp      bool
    UseCgroups      bool
    MaxMemoryBytes  int64
    MaxCPUPercent   int
    MaxProcesses    int
    MaxOpenFiles    int
    NetworkMode     string  // "none", "host", "bridge"
    AllowedHosts    []string
    AllowedPorts    []int
}
```

## Vulnerability Scanning

### Security Tools

| Tool | Purpose | Configuration |
|------|---------|---------------|
| Gosec | Go static analysis | `.gosec.yml` |
| Snyk | Dependency scanning | `.snyk` |
| Trivy | Comprehensive scanner | Makefile flags |
| SonarQube | Code quality & security | `sonar-project.properties` |

### Gosec Rules

Key security rules checked:
- G101: Hardcoded credentials
- G102: Bind to all interfaces
- G104: Unhandled errors
- G107: URL manipulation
- G201-G204: SQL injection
- G301-G307: File permissions
- G401-G405: Cryptographic issues
- G501-G505: Blocklisted imports

### Running Security Scans

```bash
# Run all security scans
make security-scan

# Individual tools
make security-scan-gosec     # Go security checker
make security-scan-snyk      # Dependency vulnerabilities
make security-scan-trivy     # Comprehensive scan

# Container image scan
trivy image helixagent:latest

# Configuration scan
trivy config .
```

## Security Best Practices

### API Security

| Practice | Implementation |
|----------|----------------|
| Authentication | JWT tokens with configurable expiry |
| Authorization | Role-based access control (RBAC) |
| Rate Limiting | Token bucket algorithm per client |
| Input Validation | Schema validation on all endpoints |
| CORS | Configurable origin whitelist |

### Secrets Management

- Environment variables for API keys
- No hardcoded credentials in code
- Secrets detection in CI pipeline
- Encrypted storage for sensitive data

### Network Security

- TLS for all external connections
- Network namespaces for container isolation
- Configurable allowed hosts/ports
- Firewall rules for production

## Authentication & Authorization

### JWT Configuration

```bash
JWT_SECRET=your-secure-secret
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h
```

### API Key Authentication

```bash
# Per-provider API keys
CLAUDE_API_KEY=...
DEEPSEEK_API_KEY=...
OPENROUTER_API_KEY=...
```

### OAuth Credentials

For providers using OAuth:
```bash
CLAUDE_USE_OAUTH_CREDENTIALS=true
QWEN_USE_OAUTH_CREDENTIALS=true
```

## Audit Logging

All security-relevant operations are logged:
- Authentication attempts
- Authorization decisions
- Plugin executions
- External API calls
- Configuration changes

### Log Configuration

```yaml
logging:
  level: info
  format: json
  audit:
    enabled: true
    path: /var/log/helixagent/audit.log
```

## Security Testing

### Challenge Scripts

```bash
# Run security challenge
./challenges/scripts/security_scanning_challenge.sh

# Full security validation
make test-security
```

### Penetration Testing

Security tests are located in `tests/security/`:
- SQL injection tests
- XSS prevention tests
- Authentication bypass tests
- Authorization boundary tests

## Incident Response

### Security Alerts

Configure alerts for:
- Failed authentication attempts
- Unusual API patterns
- Resource exhaustion attacks
- Privilege escalation attempts

### Response Procedures

1. Detect: Monitoring alerts trigger
2. Contain: Rate limiting, IP blocking
3. Investigate: Log analysis, forensics
4. Remediate: Patch, configuration update
5. Report: Document incident and response

## Related Documentation

- [Architecture Overview](../architecture/README.md)
- [Circuit Breaker](../architecture/CIRCUIT_BREAKER.md)
- [Monitoring](../monitoring/README.md)
- [Service Architecture](../architecture/SERVICE_ARCHITECTURE.md)

## Make Targets

```bash
make security-scan        # Run all security scans
make security-scan-gosec  # Go security checker
make security-scan-snyk   # Dependency vulnerabilities
make security-scan-trivy  # Comprehensive scan
make test-security        # Security tests
```

## Configuration Files

| File | Purpose |
|------|---------|
| `.gosec.yml` | Gosec exclusions and rules |
| `.snyk` | Snyk configuration |
| `sonar-project.properties` | SonarQube settings |
| `configs/security/*.yaml` | Runtime security config |
