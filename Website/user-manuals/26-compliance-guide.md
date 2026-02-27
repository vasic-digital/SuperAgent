# User Manual 26: Compliance Guide

## Overview
HelixAgent compliance and audit requirements.

## Standards
- SOC 2
- GDPR
- HIPAA
- ISO 27001

## Data Protection
```yaml
data_retention:
  logs: 90 days
  sessions: 30 days
  backups: 1 year

encryption:
  at_rest: AES-256
  in_transit: TLS 1.3
```

## Audit Logging
All actions logged with:
- Timestamp
- User ID
- Action type
- Resource accessed
- Outcome

## Access Controls
- Role-based access control (RBAC)
- Principle of least privilege
- Regular access reviews

## Compliance Reports
Generate monthly:
```bash
./scripts/generate-compliance-report.sh
```
