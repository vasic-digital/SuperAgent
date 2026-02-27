# User Manual 29: Disaster Recovery

## Overview
DR procedures for HelixAgent.

## RPO/RTO
- RPO: 5 minutes
- RTO: 30 minutes

## DR Scenarios

### Database Failure
1. Promote replica to primary
2. Update connection strings
3. Verify data consistency

### Complete Region Failure
1. Activate standby region
2. Update DNS
3. Verify services

## Runbooks
Located in: `docs/runbooks/`

### Activation
```bash
./scripts/dr-activate.sh region=us-west-2
```

### Failback
```bash
./scripts/dr-failback.sh region=us-east-1
```

## Testing
Quarterly DR drills scheduled.
