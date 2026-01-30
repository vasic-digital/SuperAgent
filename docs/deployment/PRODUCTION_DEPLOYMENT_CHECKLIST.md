# Production Deployment Checklist

**Version**: 1.0
**Last Updated**: 2026-01-30
**Target**: Production deployment of HelixAgent Big Data Integration

---

## Pre-Deployment Checklist

### Infrastructure Requirements

- [ ] **Hardware Provisioned**
  - [ ] Server 1: Messaging & Streaming (8 cores, 16GB RAM, 500GB SSD)
  - [ ] Server 2: Databases (16 cores, 32GB RAM, 1TB SSD)
  - [ ] Server 3: Processing (16 cores, 32GB RAM, 2TB HDD + 500GB SSD)
  - [ ] Total: 40+ cores, 80GB RAM, 3TB+ storage

- [ ] **Network Configuration**
  - [ ] 10 Gbps internal network between servers
  - [ ] Firewall rules configured (see `BIG_DATA_DEPLOYMENT_GUIDE.md`)
  - [ ] DNS records configured
  - [ ] SSL certificates obtained (Let's Encrypt or corporate CA)

- [ ] **Operating System**
  - [ ] Ubuntu 22.04 LTS installed on all servers
  - [ ] System updated (`apt update && apt upgrade`)
  - [ ] Docker 24.0+ installed
  - [ ] Docker Compose 2.20+ installed
  - [ ] Kernel parameters tuned (see deployment guide)

### Software Dependencies

- [ ] **Docker Runtime**
  - [ ] Docker daemon running on all servers
  - [ ] Docker network created (`helixagent-network`)
  - [ ] Docker volumes configured

- [ ] **Kafka CLI Tools**
  - [ ] `kafka-topics.sh` available
  - [ ] `kafka-console-producer.sh` available
  - [ ] `kafka-console-consumer.sh` available

- [ ] **Monitoring Tools**
  - [ ] Prometheus installed
  - [ ] Grafana installed
  - [ ] Alertmanager configured

### Configuration Files

- [ ] **Environment Variables**
  - [ ] `.env.production` created from template
  - [ ] Database credentials configured
  - [ ] API keys configured (providers)
  - [ ] OAuth credentials configured (if using Claude/Qwen CLI)

- [ ] **Performance Configuration**
  - [ ] `configs/bigdata_performance.yaml` reviewed
  - [ ] Resource limits adjusted for hardware
  - [ ] Partition counts calculated based on throughput requirements

- [ ] **Security Configuration**
  - [ ] Strong passwords set for all services
  - [ ] SSL/TLS enabled for Kafka, Neo4j, ClickHouse
  - [ ] Network segmentation configured
  - [ ] Firewall rules applied

---

## Deployment Steps

### Step 1: Clone Repository

```bash
# Clone HelixAgent repository
git clone https://github.com/vasic-digital/SuperAgent.git helixagent
cd helixagent

# Checkout production tag/branch
git checkout main
```

**Verification**:
- [ ] Repository cloned successfully
- [ ] All files present (check `ls -la`)

### Step 2: Configure Environment

```bash
# Copy production environment template
cp .env.example .env.production

# Edit with production values
vim .env.production
```

**Required Variables**:
```bash
# Server
PORT=7061
GIN_MODE=release
JWT_SECRET=<GENERATE_STRONG_SECRET>

# Database
DB_HOST=server2.internal
DB_PORT=5432
DB_USER=helixagent
DB_PASSWORD=<STRONG_PASSWORD>
DB_NAME=helixagent_db

# Redis
REDIS_HOST=server2.internal
REDIS_PORT=6379
REDIS_PASSWORD=<STRONG_PASSWORD>

# Kafka
KAFKA_BOOTSTRAP_SERVERS=server1.internal:9092

# ClickHouse
CLICKHOUSE_HOST=server2.internal
CLICKHOUSE_PORT=9000
CLICKHOUSE_DATABASE=helixagent_analytics
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=<STRONG_PASSWORD>

# Neo4j
NEO4J_URI=bolt://server2.internal:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=<STRONG_PASSWORD>
NEO4J_DATABASE=helixagent

# Big Data
BIGDATA_ENABLED=true
BIGDATA_ENABLE_INFINITE_CONTEXT=true
BIGDATA_ENABLE_DISTRIBUTED_MEMORY=true
BIGDATA_ENABLE_KNOWLEDGE_GRAPH=true
BIGDATA_ENABLE_ANALYTICS=true
BIGDATA_ENABLE_CROSS_LEARNING=true
```

**Verification**:
- [ ] All required variables set
- [ ] Strong passwords generated
- [ ] Server hostnames correct

### Step 3: Create Data Directories

```bash
# Create directories with correct permissions
sudo mkdir -p /data/helixagent/{kafka,clickhouse,neo4j,postgres,minio,spark}
sudo chown -R $USER:$USER /data/helixagent
chmod 755 /data/helixagent
```

**Verification**:
- [ ] Directories created
- [ ] Permissions correct (`ls -la /data/helixagent`)

### Step 4: Configure Docker Compose

```bash
# Review docker-compose.bigdata.yml
vim docker-compose.bigdata.yml

# Update resource limits based on available hardware
# Update volume mounts to /data/helixagent
```

**Verification**:
- [ ] Resource limits appropriate
- [ ] Volume mounts correct
- [ ] Network configuration correct

### Step 5: Start Services

```bash
# Pull latest images
docker-compose -f docker-compose.bigdata.yml pull

# Start all services
docker-compose -f docker-compose.bigdata.yml up -d

# Wait for services to be healthy
./scripts/wait-for-services.sh 600
```

**Verification**:
- [ ] All containers started (`docker ps`)
- [ ] All services healthy (`docker ps --filter "health=healthy"`)
- [ ] No errors in logs (`docker-compose -f docker-compose.bigdata.yml logs`)

### Step 6: Initialize Databases

```bash
# Run database migrations
docker exec helixagent-postgres psql -U helixagent -d helixagent_db -f /sql/schema/complete_schema.sql

# Create ClickHouse tables
for schema in sql/schema/clickhouse_*.sql; do
    cat "$schema" | curl -X POST "http://server2.internal:8123/" --data-binary @-
done

# Verify Neo4j connection
curl -u neo4j:$NEO4J_PASSWORD http://server2.internal:7474/
```

**Verification**:
- [ ] PostgreSQL tables created
- [ ] ClickHouse tables created
- [ ] Neo4j accessible

### Step 7: Create Kafka Topics

```bash
# Create all required topics with production settings
./scripts/create-kafka-topics-production.sh
```

**Topics to create**:
- [ ] `helixagent.memory.events` (12 partitions, 3 replicas)
- [ ] `helixagent.entities.updates` (8 partitions, 3 replicas)
- [ ] `helixagent.relationships.updates` (8 partitions, 3 replicas)
- [ ] `helixagent.analytics.providers` (6 partitions, 2 replicas)
- [ ] `helixagent.analytics.debates` (4 partitions, 2 replicas)
- [ ] `helixagent.conversations` (16 partitions, 3 replicas)

**Verification**:
```bash
kafka-topics.sh --bootstrap-server server1.internal:9092 --list
```

### Step 8: Deploy HelixAgent Application

```bash
# Build Docker image
docker build -t helixagent:latest .

# Run application
docker run -d \
  --name helixagent \
  --network helixagent-network \
  -p 7061:7061 \
  --env-file .env.production \
  --restart unless-stopped \
  helixagent:latest
```

**Verification**:
- [ ] Container started successfully
- [ ] Application logs show no errors
- [ ] Health endpoint accessible (`curl http://localhost:7061/health`)

### Step 9: Verify Big Data Integration

```bash
# Run validation script
./scripts/validate-bigdata-system.sh

# Expected: All tests pass (100%)
```

**Verification**:
- [ ] All 42 validation tests pass
- [ ] No critical errors

### Step 10: Run Performance Benchmarks

```bash
# Run all benchmarks
./scripts/benchmark-bigdata.sh all

# Review results
cat results/benchmarks/*/summary.txt
```

**Target Performance**:
- [ ] Kafka throughput: >10K msg/sec
- [ ] Kafka latency (p95): <10ms
- [ ] ClickHouse insert: >50K rows/sec
- [ ] ClickHouse query (p95): <50ms
- [ ] Neo4j write: >5K nodes/sec
- [ ] Context replay (10K): <5s

### Step 11: Configure Monitoring

```bash
# Start Prometheus
docker run -d \
  --name prometheus \
  --network helixagent-network \
  -p 9090:9090 \
  -v ./configs/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# Start Grafana
docker run -d \
  --name grafana \
  --network helixagent-network \
  -p 3000:3000 \
  grafana/grafana
```

**Verification**:
- [ ] Prometheus scraping metrics
- [ ] Grafana dashboards created
- [ ] Alerts configured

### Step 12: Configure Backups

```bash
# Setup backup cron jobs
crontab -e

# Add daily backup at 2 AM
0 2 * * * /opt/helixagent/scripts/backup-all.sh
```

**Backup Jobs**:
- [ ] PostgreSQL: Daily pg_dump
- [ ] Neo4j: Daily neo4j-admin dump
- [ ] MinIO: Daily mc mirror
- [ ] Kafka topics: Weekly snapshots

**Verification**:
- [ ] Cron jobs scheduled
- [ ] Backup script tested
- [ ] Backup retention configured (7 days, 4 weeks, 12 months)

---

## Post-Deployment Validation

### Functional Testing

- [ ] **Context Replay**
  ```bash
  curl -X POST http://localhost:7061/v1/context/replay \
    -H "Content-Type: application/json" \
    -d '{"conversation_id": "test", "max_tokens": 4000}'
  ```

- [ ] **Memory Sync Status**
  ```bash
  curl http://localhost:7061/v1/memory/sync/status
  ```

- [ ] **Knowledge Graph Search**
  ```bash
  curl -X POST http://localhost:7061/v1/knowledge/search \
    -H "Content-Type: application/json" \
    -d '{"query": "test entity"}'
  ```

- [ ] **Analytics Query**
  ```bash
  curl "http://localhost:7061/v1/analytics/provider/claude?window=24h"
  ```

- [ ] **Big Data Health**
  ```bash
  curl http://localhost:7061/v1/bigdata/health
  # Expected: All components "healthy"
  ```

### Load Testing

- [ ] **Concurrent Requests**
  ```bash
  # Use Apache Bench or similar
  ab -n 1000 -c 10 http://localhost:7061/health
  ```

- [ ] **Sustained Load**
  ```bash
  # Run for 10 minutes
  for i in {1..600}; do
    curl -s http://localhost:7061/health >/dev/null
    sleep 1
  done
  ```

- [ ] **Kafka Throughput Test**
  ```bash
  ./scripts/benchmark-bigdata.sh kafka
  # Verify >10K msg/sec
  ```

### Monitoring Validation

- [ ] **Prometheus Metrics**
  - [ ] `kafka_producer_record_send_total` increasing
  - [ ] `clickhouse_query_duration_seconds` < 0.1
  - [ ] `neo4j_transaction_committed_total` increasing
  - [ ] `context_replay_duration_seconds` < 5

- [ ] **Grafana Dashboards**
  - [ ] Kafka dashboard shows healthy metrics
  - [ ] ClickHouse dashboard shows query performance
  - [ ] Neo4j dashboard shows transaction rate
  - [ ] Context compression dashboard shows hit ratio >80%

- [ ] **Alerts Configured**
  - [ ] Kafka consumer lag > 10K
  - [ ] ClickHouse query > 1s
  - [ ] Neo4j heap > 85%
  - [ ] Memory sync lag > 5s

### Security Validation

- [ ] **Authentication**
  - [ ] Strong passwords enforced
  - [ ] JWT tokens working
  - [ ] OAuth credentials secured

- [ ] **Network Security**
  - [ ] Firewall rules active
  - [ ] Only required ports open
  - [ ] SSL/TLS enabled for all services

- [ ] **Access Control**
  - [ ] Database users have minimum required permissions
  - [ ] Kafka ACLs configured
  - [ ] Neo4j roles configured

### Backup Validation

- [ ] **Backup Execution**
  ```bash
  # Run backup script manually
  ./scripts/backup-all.sh

  # Verify backup files created
  ls -lh /backup/helixagent/
  ```

- [ ] **Restore Test**
  ```bash
  # Test restore on development environment
  ./scripts/restore-backup.sh /backup/helixagent/2026-01-30/
  ```

---

## Production Rollout Plan

### Phase 1: Limited Beta (Week 1)

- [ ] Deploy to 10% of production traffic
- [ ] Monitor all metrics closely
- [ ] Collect user feedback
- [ ] Fix any critical issues

### Phase 2: Gradual Rollout (Weeks 2-3)

- [ ] Increase to 25% traffic
- [ ] Validate performance under load
- [ ] Optimize based on metrics
- [ ] Increase to 50% traffic
- [ ] Increase to 75% traffic

### Phase 3: Full Production (Week 4)

- [ ] 100% production traffic
- [ ] All features enabled
- [ ] Continuous monitoring
- [ ] Regular optimization

---

## Rollback Plan

If critical issues arise:

1. **Immediate Rollback**
   ```bash
   # Stop big data services
   docker-compose -f docker-compose.bigdata.yml down

   # Disable big data in config
   sed -i 's/BIGDATA_ENABLED=true/BIGDATA_ENABLED=false/' .env.production

   # Restart HelixAgent
   docker restart helixagent
   ```

2. **Verify Rollback**
   ```bash
   # Ensure HelixAgent works without big data
   curl http://localhost:7061/health
   ```

3. **Root Cause Analysis**
   - Review logs: `docker-compose -f docker-compose.bigdata.yml logs`
   - Check metrics: Grafana dashboards
   - Identify issue

4. **Fix and Redeploy**
   - Apply fix
   - Test in staging
   - Redeploy to production

---

## Ongoing Maintenance

### Daily Tasks

- [ ] Check service health: `./scripts/check-bigdata-services.sh`
- [ ] Review Grafana dashboards
- [ ] Check for alerts
- [ ] Monitor disk usage

### Weekly Tasks

- [ ] Review performance metrics
- [ ] Analyze slow queries
- [ ] Optimize as needed
- [ ] Review backup logs
- [ ] Update documentation

### Monthly Tasks

- [ ] Review capacity planning
- [ ] Analyze costs
- [ ] Update dependencies
- [ ] Security audit
- [ ] Disaster recovery drill

---

## Support Contacts

- **Infrastructure**: infra@example.com
- **Database**: dba@example.com
- **Application**: dev@example.com
- **Security**: security@example.com
- **On-Call**: oncall@example.com

---

## Sign-Off

**Deployment Lead**: __________________ Date: __________

**Infrastructure Lead**: __________________ Date: __________

**Security Lead**: __________________ Date: __________

**Engineering Manager**: __________________ Date: __________

---

**Deployment Status**: ☐ Planning ☐ In Progress ☐ Completed ☐ Validated

**Production Ready**: ☐ Yes ☐ No (explain: _________________)

**Go-Live Date**: __________
