# User Manual 10: Security Hardening Guide

## Introduction

This guide covers security best practices for deploying and operating HelixAgent in production environments. It covers authentication, authorization, network security, and protection against common attacks.

## Security Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Security Layers                             │
├─────────────────────────────────────────────────────────────────┤
│  1. Network Security    │ TLS, Firewall, Rate Limiting          │
├─────────────────────────────────────────────────────────────────┤
│  2. Authentication      │ JWT, API Keys, OAuth2                 │
├─────────────────────────────────────────────────────────────────┤
│  3. Authorization       │ RBAC, Scopes, Resource Policies       │
├─────────────────────────────────────────────────────────────────┤
│  4. Input Validation    │ Guardrails, Sanitization              │
├─────────────────────────────────────────────────────────────────┤
│  5. Output Protection   │ PII Detection, Content Filtering      │
├─────────────────────────────────────────────────────────────────┤
│  6. Audit & Monitoring  │ Logging, Alerting, Tracing            │
└─────────────────────────────────────────────────────────────────┘
```

## Authentication

### JWT Authentication

Configure JWT for user authentication:

```yaml
auth:
  jwt:
    enabled: true
    secret: ${JWT_SECRET}  # Use 256-bit minimum
    algorithm: HS256       # Or RS256 for asymmetric
    expiration: 24h
    refresh_enabled: true
    refresh_expiration: 7d
```

**Best Practices:**
- Use strong secrets (256+ bits)
- Set short expiration times
- Implement refresh token rotation
- Store tokens securely (httpOnly cookies)

### API Key Authentication

Configure API keys for service-to-service auth:

```yaml
auth:
  api_keys:
    enabled: true
    header: X-API-Key
    hash_algorithm: sha256
    rate_limit_per_key: true
```

**Creating API Keys:**
```bash
# Via CLI
./helixagent keys create --name "Production App" --scopes "completions,debates"

# Via API
curl -X POST http://localhost:8080/v1/admin/keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name": "Production App", "scopes": ["completions", "debates"]}'
```

### OAuth2 Integration

For enterprise SSO:

```yaml
auth:
  oauth2:
    enabled: true
    providers:
      google:
        client_id: ${GOOGLE_CLIENT_ID}
        client_secret: ${GOOGLE_CLIENT_SECRET}
        allowed_domains: ["company.com"]
      okta:
        issuer: https://company.okta.com
        client_id: ${OKTA_CLIENT_ID}
        client_secret: ${OKTA_CLIENT_SECRET}
```

## Authorization

### Role-Based Access Control (RBAC)

```yaml
auth:
  rbac:
    enabled: true
    roles:
      admin:
        permissions: ["*"]
      developer:
        permissions:
          - "completions:*"
          - "debates:*"
          - "mcp:read"
      viewer:
        permissions:
          - "completions:read"
          - "debates:read"
```

### Scope-Based Authorization

API keys and tokens can have scopes:

```yaml
# Available scopes
scopes:
  - completions        # LLM completions
  - completions:read   # Read-only completions
  - debates            # AI debates
  - debates:read       # Read-only debates
  - mcp                # MCP tool execution
  - mcp:read           # Read MCP data only
  - admin              # Administrative access
```

### Resource Policies

```yaml
policies:
  # Limit model access
  models:
    allowed: ["claude-3-5-sonnet", "gemini-pro"]
    blocked: ["gpt-4"]  # Not available

  # Limit providers
  providers:
    allowed: ["claude", "gemini", "deepseek"]

  # Rate limits by role
  rate_limits:
    admin: 1000/min
    developer: 100/min
    viewer: 10/min
```

## Network Security

### TLS Configuration

Always use TLS in production:

```yaml
server:
  tls:
    enabled: true
    cert_file: /etc/ssl/certs/helixagent.crt
    key_file: /etc/ssl/private/helixagent.key
    min_version: TLS1.2
    cipher_suites:
      - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
      - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
```

### Rate Limiting

Protect against abuse:

```yaml
rate_limit:
  enabled: true

  # Global limits
  global:
    requests_per_minute: 10000
    burst_size: 100

  # Per-IP limits
  per_ip:
    requests_per_minute: 60
    burst_size: 10

  # Per-API-key limits
  per_key:
    requests_per_minute: 1000
    burst_size: 50

  # Endpoint-specific limits
  endpoints:
    "/v1/completions":
      requests_per_minute: 30
    "/v1/debates":
      requests_per_minute: 10
```

### CORS Configuration

```yaml
cors:
  enabled: true
  allowed_origins:
    - "https://app.company.com"
    - "https://admin.company.com"
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
  allowed_headers:
    - Authorization
    - Content-Type
  expose_headers:
    - X-Request-ID
  allow_credentials: true
  max_age: 3600
```

### Firewall Rules

```bash
# Allow only HTTPS
sudo ufw allow 443/tcp

# Allow health checks from load balancer
sudo ufw allow from 10.0.0.0/8 to any port 8080

# Block all other traffic
sudo ufw default deny incoming
```

## Input Validation

### Guardrails

Protect against prompt injection and jailbreaking:

```yaml
guardrails:
  enabled: true

  input:
    # Block prompt injection attempts
    block_injection:
      enabled: true
      patterns:
        - "ignore previous instructions"
        - "disregard your training"
        - "you are now"

    # Block jailbreak attempts
    block_jailbreak:
      enabled: true
      detection_model: claude-3-haiku
      threshold: 0.8

    # Content filters
    content_filters:
      enabled: true
      categories:
        - hate_speech
        - violence
        - sexual_content
      action: block  # or "warn"
```

### Request Validation

```yaml
validation:
  # Size limits
  max_prompt_length: 100000  # characters
  max_messages: 100
  max_file_size: 10MB

  # Content validation
  require_content_type: true
  allowed_content_types:
    - application/json

  # Parameter validation
  strict_parameters: true  # Reject unknown params
```

## Output Protection

### PII Detection

Automatically detect and mask sensitive data:

```yaml
pii:
  enabled: true

  detection:
    types:
      - email
      - phone
      - ssn
      - credit_card
      - api_key
      - password

  action: mask  # or "block", "warn"

  masking:
    email: "***@***.***"
    phone: "***-***-****"
    ssn: "***-**-****"
    credit_card: "****-****-****-****"
```

### Content Filtering

```yaml
output_filters:
  enabled: true

  # Block sensitive topics
  block_topics:
    - illegal_activities
    - harmful_instructions
    - private_information

  # Add disclaimers
  disclaimers:
    medical: "This is not medical advice..."
    legal: "This is not legal advice..."
    financial: "This is not financial advice..."
```

## Audit & Monitoring

### Audit Logging

```yaml
audit:
  enabled: true

  # What to log
  log_events:
    - authentication
    - authorization
    - api_calls
    - admin_actions
    - errors

  # Where to log
  destinations:
    - type: file
      path: /var/log/helixagent/audit.log
      rotation: daily
      retention: 90d

    - type: syslog
      facility: local0

    - type: elasticsearch
      url: http://elasticsearch:9200
      index: helixagent-audit
```

### Security Alerts

```yaml
alerts:
  enabled: true

  rules:
    - name: brute_force
      condition: "failed_auth_count > 10 within 5m"
      severity: high
      actions:
        - block_ip: 1h
        - notify: security@company.com

    - name: suspicious_activity
      condition: "guardrail_blocks > 5 within 1h"
      severity: medium
      actions:
        - notify: security@company.com

    - name: data_exfiltration
      condition: "pii_detected > 100 within 1h"
      severity: critical
      actions:
        - block_user
        - notify: security@company.com
```

### Monitoring

```yaml
monitoring:
  prometheus:
    enabled: true
    port: 9090

  metrics:
    - auth_success_total
    - auth_failure_total
    - guardrail_blocks_total
    - pii_detections_total
    - rate_limit_hits_total
```

## Security Checklist

### Pre-Deployment

- [ ] Generate strong JWT secrets (256+ bits)
- [ ] Configure TLS with valid certificates
- [ ] Set up firewall rules
- [ ] Configure rate limiting
- [ ] Enable audit logging
- [ ] Set up monitoring and alerts
- [ ] Review CORS settings
- [ ] Enable guardrails

### Post-Deployment

- [ ] Run security scan
- [ ] Test rate limiting
- [ ] Verify TLS configuration
- [ ] Test authentication flows
- [ ] Verify audit logs are working
- [ ] Test alert notifications
- [ ] Review access permissions

### Ongoing

- [ ] Rotate secrets quarterly
- [ ] Review audit logs weekly
- [ ] Update dependencies monthly
- [ ] Run penetration tests quarterly
- [ ] Review access permissions monthly
- [ ] Update guardrail patterns as needed

## Red Team Framework

HelixAgent includes a built-in red team framework for testing security:

```bash
# Run security tests
make test-security

# Run specific attack simulations
go test -v ./tests/security/penetration_test.go -run TestPromptInjection
go test -v ./tests/security/penetration_test.go -run TestJailbreaking
go test -v ./tests/security/penetration_test.go -run TestDataExfiltration
```

### Attack Types Tested

| Category | Attacks | Coverage |
|----------|---------|----------|
| Prompt Injection | Direct, Indirect, Nested | 8 tests |
| Jailbreaking | DAN, Roleplay, Encoding | 6 tests |
| Data Exfiltration | PII, Credentials, System | 5 tests |
| Denial of Service | Resource, Token, Loop | 4 tests |
| Privilege Escalation | Admin, System, Provider | 4 tests |

## Incident Response

### Security Incident Procedure

1. **Detect**: Identify the incident through monitoring/alerts
2. **Contain**: Block affected users/IPs, disable compromised keys
3. **Investigate**: Review audit logs, identify scope
4. **Remediate**: Fix vulnerability, rotate credentials
5. **Recover**: Restore service, verify security
6. **Report**: Document incident, notify stakeholders

### Emergency Commands

```bash
# Block IP immediately
curl -X POST http://localhost:8080/v1/admin/security/block-ip \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"ip": "1.2.3.4", "duration": "24h", "reason": "suspicious activity"}'

# Revoke API key
curl -X DELETE http://localhost:8080/v1/admin/keys/{key_id} \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Disable user
curl -X POST http://localhost:8080/v1/admin/users/{user_id}/disable \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## Compliance

### Data Privacy

- Enable PII detection and masking
- Configure data retention policies
- Implement user data deletion
- Maintain audit trails

### SOC 2 Considerations

- Authentication controls
- Access logging
- Encryption at rest and in transit
- Incident response procedures

### GDPR Considerations

- User consent management
- Right to deletion
- Data portability
- Privacy by design

---

**Document Version**: 1.0
**Last Updated**: January 23, 2026
