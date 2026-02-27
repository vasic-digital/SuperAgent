# User Manual 24: Backup and Recovery

## Overview
Data protection and disaster recovery procedures.

## Backup Types

### Database Backup
```bash
# Automated backup
./scripts/backup-database.sh

# Manual backup
pg_dump -h localhost -U helixagent helixagent_db > backup.sql
```

### Configuration Backup
```bash
tar -czf config-backup.tar.gz configs/ .env
```

## Recovery Procedures

### Database Recovery
```bash
psql -h localhost -U helixagent helixagent_db < backup.sql
```

### Point-in-Time Recovery
```bash
# Restore to specific timestamp
./scripts/restore-pit.sh "2026-02-27 10:00:00"
```

## Schedule
- Full backup: Daily at 2 AM
- Incremental: Every 4 hours
- Retention: 30 days
