# User Manual 29: Disaster Recovery

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [RTO and RPO Targets](#rto-and-rpo-targets)
4. [Disaster Recovery Architecture](#disaster-recovery-architecture)
5. [DR Scenarios and Runbooks](#dr-scenarios-and-runbooks)
6. [Database Failover](#database-failover)
7. [Redis Failover](#redis-failover)
8. [Complete Region Failure](#complete-region-failure)
9. [LLM Provider Outage](#llm-provider-outage)
10. [Data Replication](#data-replication)
11. [Backup Verification](#backup-verification)
12. [DR Activation Procedures](#dr-activation-procedures)
13. [Failback Procedures](#failback-procedures)
14. [DR Testing](#dr-testing)
15. [Communication Plan](#communication-plan)
16. [Troubleshooting](#troubleshooting)
17. [Related Resources](#related-resources)

## Overview

This manual defines disaster recovery (DR) procedures for HelixAgent across all failure scenarios: database failure, cache failure, complete region failure, LLM provider outages, and infrastructure corruption. The procedures are designed to meet the defined RTO (Recovery Time Objective) and RPO (Recovery Point Objective) targets.

HelixAgent's architecture provides inherent resilience through circuit breakers, provider fallback chains, and ensemble voting. However, infrastructure-level failures require explicit DR procedures.

## Prerequisites

- Multi-region deployment configured (see [User Manual 25](25-multi-region-deployment.md))
- Backup system operational (see [User Manual 24](24-backup-recovery.md))
- DNS provider with failover support
- Monitoring and alerting configured (see [User Manual 18](18-performance-monitoring.md))
- DR runbooks accessible to the operations team
- SSH access to all infrastructure hosts

## RTO and RPO Targets

| Scenario | RPO | RTO | Priority |
|---|---|---|---|
| Single provider failure | 0 (no data loss) | Immediate (circuit breaker) | Low |
| Database primary failure | 5 minutes | 15 minutes | Critical |
| Redis failure | 0 (cache rebuild) | 5 minutes | Medium |
| Single AZ failure | 5 minutes | 15 minutes | High |
| Complete region failure | 5 minutes | 30 minutes | Critical |
| Data corruption | Point-in-time | 60 minutes | Critical |
| Full infrastructure loss | Last backup | 2 hours | Critical |

## Disaster Recovery Architecture

```
                    Primary Region (US-East)
+-------------------------------------------------------+
|  +----------+  +-----------+  +----------+            |
|  |HelixAgent|  |HelixAgent |  |HelixAgent|            |
|  |  Pod 1   |  |  Pod 2    |  |  Pod 3   |            |
|  +-----+----+  +-----+-----+  +-----+----+            |
|        |              |              |                  |
|  +-----v--------------v--------------v-----+           |
|  |          PostgreSQL Primary              |           |
|  |          (WAL streaming)                 +--------+  |
|  +------------------------------------------+        |  |
|  +-----+----+                                        |  |
|  |   Redis   |                                       |  |
|  |  Primary  |                                       |  |
|  +-----+----+                                        |  |
+--------|---------------------------------------------+  |
         |                                                |
         | Replication                     WAL Streaming  |
         |                                                |
+--------v---------------------------------------------+  |
|                DR Region (US-West)                    |  |
|  +----------+  +-----------+  +----------+           |  |
|  |HelixAgent|  |HelixAgent |  |HelixAgent|           |  |
|  | Standby  |  |  Standby  |  |  Standby | (scaled  |  |
|  +----------+  +-----------+  +----------+  to 0)    |  |
|                                                      |  |
|  +------------------------------------------+        |  |
|  |          PostgreSQL Replica              <--------+  |
|  |          (hot standby)                   |           |
|  +------------------------------------------+           |
|  +-----+----+                                           |
|  |   Redis   |                                          |
|  |  Replica  |                                          |
|  +----------+                                           |
+----------------------------------------------------------+
```

## DR Scenarios and Runbooks

### Scenario Classification

| Level | Description | Example | Response |
|---|---|---|---|
| L1 | Single component degradation | One provider down | Automatic (circuit breaker) |
| L2 | Service degradation | Database slow, high latency | Investigate, possible manual intervention |
| L3 | Partial outage | Database primary down | Execute failover runbook |
| L4 | Regional outage | Complete AZ or region failure | Activate DR region |
| L5 | Catastrophic failure | Data corruption, full infrastructure loss | Restore from backup |

## Database Failover

### Automatic Failover (Patroni/Stolon)

If using a PostgreSQL HA solution like Patroni:

```bash
# Check cluster status
patronictl -c /etc/patroni/patroni.yml list

# The cluster automatically promotes a replica if the primary fails
# Verify the new primary
patronictl -c /etc/patroni/patroni.yml switchover
```

### Manual Failover

```bash
# Step 1: Verify primary is truly unreachable
pg_isready -h primary-host -p 5432
# Returns: primary-host:5432 - no response

# Step 2: Check replica lag before promoting
psql -h replica-host -p 5432 -U helixagent -c \
    "SELECT pg_last_wal_receive_lsn(), pg_last_wal_replay_lsn(),
     pg_last_wal_receive_lsn() - pg_last_wal_replay_lsn() AS lag;"

# Step 3: Promote the replica
psql -h replica-host -p 5432 -U helixagent -c "SELECT pg_promote();"
# Or: pg_ctl promote -D /var/lib/postgresql/data

# Step 4: Update HelixAgent configuration
# Update DB_HOST environment variable or Kubernetes secret
kubectl -n helixagent set env deployment/helixagent DB_HOST=replica-host

# Step 5: Verify connectivity
curl -s http://localhost:7061/v1/monitoring/status | jq '.components.postgresql'

# Step 6: Restart HelixAgent pods for clean connection pool
kubectl -n helixagent rollout restart deployment/helixagent
```

### Post-Failover: Rebuild Old Primary as Replica

```bash
# On the old primary (after it recovers)
pg_basebackup -h new-primary-host -U replicator -D /var/lib/postgresql/data -Fp -Xs -P -R
systemctl start postgresql
```

## Redis Failover

### With Redis Sentinel

Sentinel automatically handles failover:

```bash
# Check Sentinel status
redis-cli -p 26379 SENTINEL masters

# Check which node is the current master
redis-cli -p 26379 SENTINEL get-master-addr-by-name helixagent-master

# Force manual failover if needed
redis-cli -p 26379 SENTINEL failover helixagent-master
```

### Without Sentinel (Manual)

```bash
# Step 1: Verify primary is down
redis-cli -h redis-primary -p 6379 -a helixagent123 PING
# Returns: Could not connect

# Step 2: Promote replica
redis-cli -h redis-replica -p 6379 -a helixagent123 REPLICAOF NO ONE

# Step 3: Update HelixAgent configuration
kubectl -n helixagent set env deployment/helixagent REDIS_HOST=redis-replica

# Step 4: HelixAgent will rebuild cache on demand (cache misses are acceptable)
```

### Cache Rebuild

Redis data loss is acceptable since HelixAgent uses Redis primarily for caching. On cache miss, data is regenerated from the source (LLM providers, database). The only impact is temporarily increased latency during cache warming.

## Complete Region Failure

### Activation Procedure

```bash
# Step 1: Confirm regional failure
# Check monitoring dashboards, cloud provider status page
curl -s https://us-east.helixagent.io/v1/monitoring/status
# Returns: timeout or error

# Step 2: Activate DR region
./scripts/dr-activate.sh region=us-west-2

# This script performs:
# - Promotes PostgreSQL replica in DR region
# - Promotes Redis replica in DR region
# - Scales up HelixAgent pods in DR region (from 0 to 3+)
# - Waits for readiness probes to pass
# - Updates DNS to point to DR region

# Step 3: Verify DR region is serving traffic
curl -s https://helixagent.io/v1/monitoring/status | jq .

# Step 4: Monitor error rates during transition
curl -s https://helixagent.io/v1/monitoring/provider-health | jq .
```

### DR Activation Script

```bash
#!/bin/bash
# scripts/dr-activate.sh
set -euo pipefail

REGION=${1#region=}
echo "[$(date)] Activating DR region: ${REGION}"

echo "[Step 1] Promoting PostgreSQL replica..."
ssh deploy@${REGION}-db "psql -c 'SELECT pg_promote();'"

echo "[Step 2] Promoting Redis replica..."
ssh deploy@${REGION}-redis "redis-cli -a helixagent123 REPLICAOF NO ONE"

echo "[Step 3] Scaling up HelixAgent pods..."
kubectl --context=${REGION} -n helixagent scale deployment/helixagent --replicas=3

echo "[Step 4] Waiting for pods to be ready..."
kubectl --context=${REGION} -n helixagent rollout status deployment/helixagent --timeout=120s

echo "[Step 5] Updating DNS..."
# Update DNS via cloud provider API
./scripts/dns-update.sh target=${REGION}

echo "[Step 6] Verifying..."
sleep 10
STATUS=$(curl -s -o /dev/null -w "%{http_code}" https://helixagent.io/v1/monitoring/status)
if [ "$STATUS" = "200" ]; then
    echo "[$(date)] DR activation complete. Region ${REGION} is serving traffic."
else
    echo "[$(date)] WARNING: Health check returned ${STATUS}. Manual verification required."
fi
```

## LLM Provider Outage

HelixAgent handles provider outages automatically through circuit breakers and the fallback chain:

```
Provider fails -> Circuit breaker opens -> Fallback to next provider -> Ensemble reconfigures

Timeline:
  T+0:   Provider starts returning errors
  T+30s: Circuit breaker opens (after threshold failures)
  T+30s: Requests automatically routed to fallback providers
  T+5m:  Circuit breaker enters half-open state (probes the provider)
  T+5m:  If provider recovers, circuit breaker closes
  T+5m:  If still failing, circuit breaker remains open
```

### Manual Provider Recovery

```bash
# View circuit breaker states
curl -s http://localhost:7061/v1/monitoring/circuit-breakers | jq .

# Force reset all circuit breakers
curl -s -X POST http://localhost:7061/v1/monitoring/reset-circuits

# Force health check
curl -s -X POST http://localhost:7061/v1/monitoring/force-health-check
```

## Data Replication

### PostgreSQL Streaming Replication

Ensure replication is healthy:

```bash
# On the primary: check replication status
psql -U helixagent -c "SELECT client_addr, state, sent_lsn, write_lsn, replay_lsn,
    pg_wal_lsn_diff(sent_lsn, replay_lsn) AS lag_bytes
    FROM pg_stat_replication;"

# On the replica: check replication delay
psql -U helixagent -c "SELECT NOW() - pg_last_xact_replay_timestamp() AS replication_delay;"
```

### Acceptable Lag Thresholds

| Metric | Warning | Critical |
|---|---|---|
| Replication lag (bytes) | > 16 MB | > 64 MB |
| Replication lag (time) | > 1 minute | > 5 minutes |
| WAL archive lag | > 10 files | > 50 files |

## Backup Verification

### Automated Verification

Run backup verification weekly to ensure recoverability:

```bash
#!/bin/bash
# scripts/verify-dr-readiness.sh

echo "=== DR Readiness Check ==="

# 1. Verify backups exist and are recent
LATEST_BACKUP=$(ls -t /backups/postgres/*.dump | head -1)
BACKUP_AGE=$(( ($(date +%s) - $(stat -c %Y "$LATEST_BACKUP")) / 3600 ))
if [ "$BACKUP_AGE" -gt 24 ]; then
    echo "[FAIL] Latest backup is ${BACKUP_AGE} hours old (threshold: 24h)"
else
    echo "[PASS] Latest backup is ${BACKUP_AGE} hours old"
fi

# 2. Verify replication is current
LAG=$(psql -h replica-host -U helixagent -t -c \
    "SELECT EXTRACT(EPOCH FROM NOW() - pg_last_xact_replay_timestamp())::int;")
if [ "$LAG" -gt 300 ]; then
    echo "[FAIL] Replication lag is ${LAG} seconds (threshold: 300s)"
else
    echo "[PASS] Replication lag is ${LAG} seconds"
fi

# 3. Verify DR region pods can start
DR_PODS=$(kubectl --context=us-west-2 -n helixagent get pods --no-headers 2>/dev/null | wc -l)
echo "[INFO] DR region has ${DR_PODS} pods defined"

# 4. Test backup restore (to a temp database)
echo "[INFO] Testing backup restore..."
./scripts/verify-backup.sh "$LATEST_BACKUP"

echo "=== DR Readiness Check Complete ==="
```

## Failback Procedures

After the primary region recovers, failback to restore normal operations:

```bash
#!/bin/bash
# scripts/dr-failback.sh
set -euo pipefail

PRIMARY_REGION=${1#region=}
echo "[$(date)] Starting failback to ${PRIMARY_REGION}"

# Step 1: Verify primary region infrastructure is healthy
echo "[Step 1] Verifying primary region infrastructure..."
ssh deploy@${PRIMARY_REGION}-db "pg_isready"
ssh deploy@${PRIMARY_REGION}-redis "redis-cli -a helixagent123 PING"

# Step 2: Resync data from DR to primary
echo "[Step 2] Resyncing PostgreSQL data..."
ssh deploy@${PRIMARY_REGION}-db "pg_basebackup -h dr-region-db -U replicator -D /var/lib/postgresql/data -Fp -Xs -P -R"

# Step 3: Wait for replication to catch up
echo "[Step 3] Waiting for replication catchup..."
sleep 60

# Step 4: Promote primary region database
echo "[Step 4] Promoting primary database..."
ssh deploy@${PRIMARY_REGION}-db "psql -c 'SELECT pg_promote();'"

# Step 5: Scale up primary region HelixAgent
echo "[Step 5] Scaling up primary region..."
kubectl --context=${PRIMARY_REGION} -n helixagent scale deployment/helixagent --replicas=3
kubectl --context=${PRIMARY_REGION} -n helixagent rollout status deployment/helixagent --timeout=120s

# Step 6: Update DNS back to primary
echo "[Step 6] Updating DNS..."
./scripts/dns-update.sh target=${PRIMARY_REGION}

# Step 7: Scale down DR region
echo "[Step 7] Scaling down DR region..."
kubectl --context=us-west-2 -n helixagent scale deployment/helixagent --replicas=0

# Step 8: Reconfigure DR replica
echo "[Step 8] Reconfiguring DR replica..."
ssh deploy@dr-region-db "pg_basebackup -h ${PRIMARY_REGION}-db -U replicator -D /var/lib/postgresql/data -Fp -Xs -P -R"

echo "[$(date)] Failback to ${PRIMARY_REGION} complete."
```

## DR Testing

### Quarterly DR Drills

Schedule quarterly disaster recovery drills:

1. **Tabletop exercise** -- Walk through runbooks without executing
2. **Component failover** -- Fail individual components (database, Redis) in staging
3. **Regional failover** -- Full DR activation in staging environment
4. **Production failover** -- Annual production DR drill during maintenance window

### DR Drill Checklist

- [ ] Notify stakeholders of the planned DR drill
- [ ] Verify monitoring alerts trigger correctly
- [ ] Execute activation runbook
- [ ] Verify service availability in DR region
- [ ] Measure actual RTO (compare to target)
- [ ] Test data integrity in DR region
- [ ] Execute failback runbook
- [ ] Verify service restoration in primary region
- [ ] Document lessons learned
- [ ] Update runbooks with any corrections

## Communication Plan

### Escalation Matrix

| Severity | Response Time | Notification Channel | Approver |
|---|---|---|---|
| L1 (component) | 15 minutes | Slack #ops | On-call engineer |
| L2 (degradation) | 10 minutes | Slack #ops + PagerDuty | Engineering lead |
| L3 (partial outage) | 5 minutes | PagerDuty + SMS | Engineering manager |
| L4 (regional outage) | Immediate | PagerDuty + phone | VP Engineering |
| L5 (catastrophic) | Immediate | All channels | CTO |

### Status Page Updates

During a DR event, post updates every 15 minutes:

```
[2026-03-08 10:00] INVESTIGATING: Elevated error rates in US-East region
[2026-03-08 10:15] IDENTIFIED: US-East database primary is unreachable
[2026-03-08 10:20] ACTING: Activating DR failover to US-West region
[2026-03-08 10:30] MONITORING: DR region is serving traffic, verifying data integrity
[2026-03-08 10:45] RESOLVED: Service fully restored via DR region
```

## Troubleshooting

### DR Region Pods Fail to Start

**Symptom:** Pods crash or fail readiness probes in DR region.

**Solutions:**
1. Check pod logs: `kubectl -n helixagent logs deployment/helixagent`
2. Verify database connectivity from DR pods
3. Ensure environment variables/secrets are configured for DR region
4. Increase `initialDelaySeconds` on readiness probe (startup verification takes ~2 minutes)

### Data Inconsistency After Failover

**Symptom:** Missing or stale data after database failover.

**Solutions:**
1. Check replication lag at time of failure
2. Compare row counts between backup and restored database
3. Run integrity checks on critical tables (debate_sessions, debate_turns)
4. Accept RPO-bounded data loss and notify affected users

### DNS Failover Too Slow

**Symptom:** Clients still connect to the failed region after DNS update.

**Solutions:**
1. Lower DNS TTL to 60 seconds (must be set before the failure occurs)
2. Use DNS providers with instant propagation (Cloudflare, Route 53)
3. Clients should respect TTL and not cache DNS longer
4. Consider client-side failover configuration as a backup

### Failback Causes Replication Conflicts

**Symptom:** PostgreSQL replication fails after failback with timeline divergence.

**Solutions:**
1. Rebuild the old primary from scratch using `pg_basebackup`
2. Ensure only one node is the primary at any time
3. Use `recovery_target_timeline = 'latest'` in recovery configuration

## Related Resources

- [User Manual 24: Backup and Recovery](24-backup-recovery.md) -- Backup procedures and restore
- [User Manual 25: Multi-Region Deployment](25-multi-region-deployment.md) -- Multi-region infrastructure
- [User Manual 18: Performance Monitoring](18-performance-monitoring.md) -- Monitoring during DR events
- [User Manual 30: Enterprise Architecture](30-enterprise-architecture.md) -- Enterprise HA architecture
- Runbooks: `docs/runbooks/`
- DR scripts: `scripts/dr-activate.sh`, `scripts/dr-failback.sh`
