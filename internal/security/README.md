# Security Package

The security package provides security scanning, vulnerability detection, and secure code generation for HelixAgent.

## Overview

This package implements:

- **Security Scanner**: Pattern-based vulnerability detection
- **Secure Fix Agent**: AI-powered security remediation
- **Code Analysis**: Static analysis for security issues
- **Vulnerability Database**: Known vulnerability patterns

## Key Components

### Pattern-Based Scanner

Detects security vulnerabilities using regex patterns:

```go
scanner := security.NewPatternBasedScanner(config)

// Scan code content
vulns, err := scanner.Scan(ctx, code, "go")

// Scan file
vulns, err := scanner.ScanFile(ctx, "/path/to/file.go")

for _, vuln := range vulns {
    fmt.Printf("[%s] %s at line %d\n", vuln.Severity, vuln.Title, vuln.Line)
}
```

### Secure Fix Agent

AI-powered security vulnerability remediation:

```go
agent := security.NewSecureFixAgent(config, llmClient)

fix, err := agent.FixVulnerability(ctx, &security.FixRequest{
    Code:          vulnerableCode,
    Vulnerability: vuln,
    Language:      "go",
})
```

## Vulnerability Types

| Category | Examples |
|----------|----------|
| **Injection** | SQL injection, Command injection, XSS |
| **Authentication** | Hardcoded credentials, Weak passwords |
| **Cryptography** | Weak algorithms, Insecure random |
| **Data Exposure** | Sensitive data logging, PII leaks |
| **Configuration** | Debug mode, Default credentials |
| **Dependencies** | Vulnerable packages |

## Security Patterns

### SQL Injection Detection

```go
patterns := []security.Pattern{
    {
        Name:     "SQL Injection",
        Pattern:  `(?i)(execute|query)\s*\(\s*["\'].*\+.*["\']`,
        Severity: security.SeverityHigh,
        CWE:      "CWE-89",
    },
}
```

### XSS Detection

```go
patterns := []security.Pattern{
    {
        Name:     "Cross-Site Scripting",
        Pattern:  `(?i)innerHTML\s*=|document\.write\(`,
        Severity: security.SeverityMedium,
        CWE:      "CWE-79",
    },
}
```

## Configuration

```go
type ScannerConfig struct {
    Patterns        []Pattern
    IgnoreFiles     []string
    IgnorePaths     []string
    MaxFileSize     int64
    EnableCWE       bool
    EnableOWASP     bool
    SeverityFilter  []Severity
}
```

## Severity Levels

| Level | Description |
|-------|-------------|
| `Critical` | Immediate exploitation risk |
| `High` | Significant security impact |
| `Medium` | Moderate security concern |
| `Low` | Minor security issue |
| `Info` | Informational finding |

## Usage Examples

### Full Security Scan

```go
results, err := scanner.ScanDirectory(ctx, "./src", &security.ScanOptions{
    Recursive:   true,
    IncludeTest: false,
    Languages:   []string{"go", "javascript"},
})

for _, result := range results.Vulnerabilities {
    fmt.Printf("%s: %s (%s)\n", result.File, result.Title, result.Severity)
}
```

### Generate Security Report

```go
report, err := scanner.GenerateReport(ctx, results, &security.ReportOptions{
    Format:      "json",
    IncludeFix:  true,
    GroupByCWE:  true,
})
```

### Integrate with CI/CD

```go
results, _ := scanner.ScanDirectory(ctx, ".", nil)

if results.HasCritical() || results.HasHigh() {
    os.Exit(1) // Fail build
}
```

## Language Support

| Language | Coverage |
|----------|----------|
| Go | Full |
| JavaScript | Full |
| TypeScript | Full |
| Python | Full |
| Java | Partial |
| Rust | Partial |
| PHP | Partial |

## Testing

```bash
# Run security tests
go test -v ./internal/security/...

# Test vulnerability detection
go test -v -run TestVulnerability ./internal/security/

# Test fix generation
go test -v -run TestSecureFix ./internal/security/
```

## OWASP Top 10 Coverage

- A01: Broken Access Control
- A02: Cryptographic Failures
- A03: Injection
- A04: Insecure Design
- A05: Security Misconfiguration
- A06: Vulnerable Components
- A07: Authentication Failures
- A08: Data Integrity Failures
- A09: Logging Failures
- A10: Server-Side Request Forgery

## See Also

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Database](https://cwe.mitre.org/)
- `tests/security/` - Security test suite
