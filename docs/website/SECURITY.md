# HelixAgent Security

Comprehensive security features and practices built into HelixAgent for enterprise-grade protection.

---

## Security Overview

HelixAgent implements defense-in-depth security with multiple layers of protection:

```
┌─────────────────────────────────────────────────────────────────┐
│                        Security Layers                          │
├─────────────────────────────────────────────────────────────────┤
│  Layer 1: Network Security                                      │
│  └── TLS 1.3, Firewall rules, DDoS protection                  │
├─────────────────────────────────────────────────────────────────┤
│  Layer 2: Authentication & Authorization                        │
│  └── JWT, API Keys, OAuth, RBAC                                │
├─────────────────────────────────────────────────────────────────┤
│  Layer 3: Input Validation & Sanitization                       │
│  └── Request validation, Injection prevention                   │
├─────────────────────────────────────────────────────────────────┤
│  Layer 4: Content Security                                      │
│  └── Guardrails, PII detection, Content filtering              │
├─────────────────────────────────────────────────────────────────┤
│  Layer 5: Data Protection                                       │
│  └── Encryption at rest, Encryption in transit                 │
├─────────────────────────────────────────────────────────────────┤
│  Layer 6: Audit & Monitoring                                    │
│  └── Logging, Alerting, Intrusion detection                    │
└─────────────────────────────────────────────────────────────────┘
```

---

## Authentication

### JWT Authentication

JSON Web Tokens for stateless authentication:

```yaml
auth:
  jwt:
    enabled: true
    secret: "${JWT_SECRET}"
    expiration: 24h
    refresh_enabled: true
    refresh_expiration: 168h
    algorithm: HS256
```

**Usage:**

```bash
# Get JWT token
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "pass"}'

# Use token
curl -H "Authorization: Bearer eyJhbGciOi..." \
  http://localhost:8080/v1/chat/completions
```

### API Key Authentication

Static API keys for service-to-service communication:

```yaml
auth:
  api_keys:
    enabled: true
    header: "X-API-Key"
    keys:
      - name: "production-service"
        key: "${API_KEY_PROD}"
        permissions: ["chat", "models"]
        rate_limit: 1000
      - name: "analytics-service"
        key: "${API_KEY_ANALYTICS}"
        permissions: ["metrics", "health"]
        rate_limit: 100
```

### OAuth Integration

OAuth 2.0 support for external identity providers:

```yaml
auth:
  oauth:
    enabled: true
    providers:
      google:
        client_id: "${GOOGLE_CLIENT_ID}"
        client_secret: "${GOOGLE_CLIENT_SECRET}"
        scopes: ["openid", "profile", "email"]
      github:
        client_id: "${GITHUB_CLIENT_ID}"
        client_secret: "${GITHUB_CLIENT_SECRET}"
        scopes: ["user:email"]
```

---

## Authorization

### Role-Based Access Control (RBAC)

```yaml
authorization:
  rbac:
    enabled: true
    roles:
      admin:
        permissions: ["*"]
      developer:
        permissions:
          - "chat:create"
          - "chat:read"
          - "models:list"
          - "embeddings:create"
      viewer:
        permissions:
          - "chat:read"
          - "models:list"
```

### Permission Matrix

| Role | Chat | Models | Admin | Analytics | Memory |
|------|------|--------|-------|-----------|--------|
| admin | Full | Full | Full | Full | Full |
| developer | Create/Read | List | None | Read | Create/Read |
| viewer | Read | List | None | None | Read |

---

## Rate Limiting

### Configuration

```yaml
rate_limiting:
  enabled: true
  strategy: sliding_window

  global:
    requests_per_second: 100
    burst: 200

  per_user:
    requests_per_minute: 60
    burst: 100

  per_api_key:
    requests_per_hour: 1000
    burst: 50

  endpoints:
    "/v1/chat/completions":
      requests_per_minute: 30
    "/v1/embeddings":
      requests_per_minute: 100
```

### Response Headers

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1706886400
Retry-After: 30
```

---

## Content Security

### Guardrails Engine

Protect against harmful content:

```yaml
guardrails:
  enabled: true

  input:
    max_length: 100000
    block_patterns:
      - "(?i)ignore previous instructions"
      - "(?i)system prompt"
    content_filters:
      - harmful_content
      - hate_speech
      - violence

  output:
    max_length: 50000
    content_filters:
      - pii_leakage
      - sensitive_data
```

### PII Detection and Redaction

Automatically detect and protect sensitive information:

```yaml
pii:
  enabled: true
  action: redact  # detect, redact, block

  types:
    - email
    - phone
    - ssn
    - credit_card
    - api_key
    - password
    - ip_address

  patterns:
    custom_id:
      pattern: "CUST-[A-Z0-9]{8}"
      redaction: "[CUSTOMER_ID]"
```

**Detection Example:**

| Input | Output (Redacted) |
|-------|-------------------|
| `Email me at john@example.com` | `Email me at [EMAIL]` |
| `My SSN is 123-45-6789` | `My SSN is [SSN]` |
| `API key: sk-abc123xyz` | `API key: [API_KEY]` |

---

## Vulnerability Scanning

### Built-in Security Scanner

```go
import "dev.helix.agent/internal/security"

scanner := security.NewPatternBasedScanner(config)

// Scan code
vulns, err := scanner.Scan(ctx, code, "go")

// Scan directory
results, err := scanner.ScanDirectory(ctx, "./src", &security.ScanOptions{
    Recursive:   true,
    Languages:   []string{"go", "javascript", "python"},
    IgnorePaths: []string{"vendor/", "node_modules/"},
})
```

### Vulnerability Categories

| Category | CWE | Examples |
|----------|-----|----------|
| **Injection** | CWE-89, CWE-78 | SQL injection, Command injection |
| **XSS** | CWE-79 | Cross-site scripting |
| **Authentication** | CWE-287 | Hardcoded credentials |
| **Cryptography** | CWE-327, CWE-338 | Weak algorithms, Insecure random |
| **Data Exposure** | CWE-200, CWE-532 | Sensitive logging |
| **Configuration** | CWE-16 | Debug mode, Default creds |

### OWASP Top 10 Coverage

| # | Risk | Detection |
|---|------|-----------|
| A01 | Broken Access Control | Authorization checks |
| A02 | Cryptographic Failures | Weak crypto detection |
| A03 | Injection | SQL/Command injection patterns |
| A04 | Insecure Design | Security anti-patterns |
| A05 | Security Misconfiguration | Config validation |
| A06 | Vulnerable Components | Dependency scanning |
| A07 | Authentication Failures | Auth weakness detection |
| A08 | Data Integrity Failures | Deserialization checks |
| A09 | Logging Failures | Sensitive data logging |
| A10 | SSRF | Server-side request patterns |

---

## Encryption

### Data in Transit

TLS 1.3 for all communications:

```yaml
tls:
  enabled: true
  version: "1.3"
  cert_file: "/etc/ssl/certs/helixagent.crt"
  key_file: "/etc/ssl/private/helixagent.key"
  min_version: "1.2"
  ciphers:
    - "TLS_AES_256_GCM_SHA384"
    - "TLS_CHACHA20_POLY1305_SHA256"
```

### Data at Rest

Encryption for stored data:

```yaml
encryption:
  enabled: true
  algorithm: "AES-256-GCM"
  key_management:
    provider: vault  # vault, aws-kms, gcp-kms
    key_rotation: 90d
```

### Secret Management

```yaml
secrets:
  provider: vault
  vault:
    address: "https://vault.example.com"
    token: "${VAULT_TOKEN}"
    path: "secret/helixagent"
    mount: "kv-v2"
```

---

## Audit Logging

### Configuration

```yaml
audit:
  enabled: true
  level: detailed  # basic, detailed, verbose

  events:
    - authentication
    - authorization
    - api_calls
    - configuration_changes
    - security_events

  storage:
    type: file
    path: "/var/log/helixagent/audit.log"
    rotation:
      max_size: 100MB
      max_age: 90d
      compress: true

  export:
    enabled: true
    destination: siem
    format: json
```

### Audit Log Format

```json
{
  "timestamp": "2026-02-03T10:30:00Z",
  "event_type": "api_call",
  "user_id": "user-123",
  "api_key": "prod-***",
  "endpoint": "/v1/chat/completions",
  "method": "POST",
  "source_ip": "192.168.1.100",
  "user_agent": "python-openai/1.0.0",
  "status_code": 200,
  "latency_ms": 1500,
  "request_id": "req-abc123",
  "metadata": {
    "model": "helixagent-debate",
    "tokens_used": 150
  }
}
```

---

## Security Headers

HelixAgent automatically sets security headers:

```http
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
Content-Security-Policy: default-src 'self'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

---

## Network Security

### Firewall Configuration

```yaml
network:
  allowed_ips:
    - "10.0.0.0/8"
    - "172.16.0.0/12"
    - "192.168.0.0/16"

  blocked_ips:
    - "203.0.113.0/24"

  cors:
    enabled: true
    allowed_origins:
      - "https://app.example.com"
    allowed_methods:
      - "GET"
      - "POST"
    allowed_headers:
      - "Authorization"
      - "Content-Type"
```

### DDoS Protection

```yaml
ddos_protection:
  enabled: true
  max_connections: 10000
  connection_rate: 100/s
  request_rate: 1000/s
  body_limit: 10MB
```

---

## Compliance

### Supported Frameworks

| Framework | Coverage |
|-----------|----------|
| **SOC 2 Type II** | Full |
| **GDPR** | Full |
| **CCPA** | Full |
| **HIPAA** | Partial (BAA required) |
| **PCI DSS** | Level 2 |
| **ISO 27001** | Aligned |

### Data Residency

```yaml
data_residency:
  enabled: true
  regions:
    - us-west-2
    - eu-west-1
  default_region: us-west-2
  enforce_region: true
```

### Data Retention

```yaml
data_retention:
  conversations:
    duration: 90d
    action: delete
  audit_logs:
    duration: 365d
    action: archive
  analytics:
    duration: 730d
    action: anonymize
```

---

## Security Best Practices

### API Key Management

1. **Never commit** API keys to version control
2. **Use environment variables** or secret managers
3. **Rotate keys** regularly (90 days recommended)
4. **Scope permissions** to minimum required
5. **Monitor usage** for anomalies

### Deployment Security

1. **Use TLS** for all connections
2. **Enable authentication** for all endpoints
3. **Set rate limits** appropriate to your load
4. **Configure firewalls** to restrict access
5. **Run as non-root** user in containers

### Monitoring

1. **Enable audit logging**
2. **Set up alerts** for security events
3. **Monitor failed authentications**
4. **Track API usage patterns**
5. **Review logs regularly**

---

## Security Challenges

Run security validation:

```bash
# Full security scan
./challenges/scripts/security_scanning_challenge.sh

# Expected output: 10 tests
# - Injection detection
# - XSS detection
# - Credential detection
# - PII detection
# - Configuration validation
# - Authentication tests
# - Authorization tests
# - Rate limiting tests
# - Encryption verification
# - Audit logging verification
```

---

## Incident Response

### Contact

- **Security Team**: security@helixagent.ai
- **Bug Bounty**: [hackerone.com/helixagent](https://hackerone.com/helixagent)
- **Responsible Disclosure**: Follow [security.txt](/.well-known/security.txt)

### Response SLA

| Severity | Response Time | Resolution Target |
|----------|---------------|-------------------|
| Critical | 1 hour | 24 hours |
| High | 4 hours | 72 hours |
| Medium | 24 hours | 7 days |
| Low | 72 hours | 30 days |

---

## Security Checklist

Before going to production:

- [ ] TLS configured and enforced
- [ ] Authentication enabled
- [ ] Rate limiting configured
- [ ] PII detection enabled
- [ ] Audit logging enabled
- [ ] Secrets in secret manager
- [ ] Security headers configured
- [ ] CORS properly restricted
- [ ] Firewall rules in place
- [ ] Monitoring and alerting set up
- [ ] Security challenges passing
- [ ] Incident response plan documented

---

**Last Updated**: February 2026
**Version**: 1.0.0
**Security Contact**: security@helixagent.ai
