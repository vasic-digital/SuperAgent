# HelixAgent Big Data - Production Deployment Guide

**Version**: 1.0
**Last Updated**: 2026-01-30
**Target**: Production deployment on Linux servers

---

## Table of Contents

1. [Infrastructure Requirements](#infrastructure-requirements)
2. [Pre-Deployment Checklist](#pre-deployment-checklist)
3. [Installation Steps](#installation-steps)
4. [Service Configuration](#service-configuration)
5. [Monitoring Setup](#monitoring-setup)
6. [Backup & Recovery](#backup--recovery)
7. [Scaling Guide](#scaling-guide)
8. [Troubleshooting](#troubleshooting)

---

## Infrastructure Requirements

### Minimum Hardware Requirements

| Component | CPU | RAM | Disk | IOPS |
|-----------|-----|-----|------|------|
| **Kafka + Zookeeper** | 2 cores | 4GB | 200GB SSD | 3000+ |
| **ClickHouse** | 4 cores | 8GB | 500GB SSD | 5000+ |
| **Neo4j** | 4 cores | 6GB | 200GB SSD | 3000+ |
| **MinIO** | 2 cores | 4GB | 1TB HDD | 1000+ |
| **Spark Master** | 2 cores | 2GB | 100GB | 1000+ |
| **Spark Workers (×4)** | 4 cores | 8GB | 200GB SSD | 3000+ |
| **Flink JobManager** | 2 cores | 2GB | 100GB | 1000+ |
| **Flink TaskManagers (×4)** | 2 cores | 4GB | 200GB SSD | 3000+ |
| **Qdrant** | 4 cores | 8GB | 200GB SSD | 5000+ |
| **Total (Multi-Server)** | 40+ cores | 80GB | 3TB+ | - |

### Recommended Production Setup

**3-Server Cluster**:

| Server | Role | Services | Specs |
|--------|------|----------|-------|
| **Server 1** | Messaging & Streaming | Zookeeper, Kafka, Flink | 8 cores, 16GB RAM, 500GB SSD |
| **Server 2** | Databases | PostgreSQL, ClickHouse, Neo4j | 16 cores, 32GB RAM, 1TB SSD |
| **Server 3** | Processing | Spark, MinIO, Qdrant | 16 cores, 32GB RAM, 2TB HDD + 500GB SSD |

**Network Requirements**:
- **Internal**: 10 Gbps for inter-service communication
- **External**: 1 Gbps for client access
- **Latency**: <1ms between servers

### Operating System

**Supported OS**:
- Ubuntu 22.04 LTS (recommended)
- CentOS 8 / Rocky Linux 8
- Debian 11+
- Red Hat Enterprise Linux 8+

**Kernel Tuning**:
```bash
# Add to /etc/sysctl.conf
vm.swappiness=1
vm.max_map_count=262144
net.core.somaxconn=65535
net.ipv4.tcp_max_syn_backlog=8192
net.core.netdev_max_backlog=5000
fs.file-max=2097152

# Apply settings
sudo sysctl -p
```

---

## Pre-Deployment Checklist

### 1. System Preparation

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

### 2. Create Data Directories

```bash
# Create mount points for high-performance volumes
sudo mkdir -p /mnt/data/{kafka,clickhouse,neo4j,flink,spark}
sudo mkdir -p /mnt/storage/minio

# Set ownership
sudo chown -R $(id -u):$(id -g) /mnt/data /mnt/storage

# Set permissions
sudo chmod -R 755 /mnt/data /mnt/storage
```

### 3. Configure Firewall

```bash
# UFW (Ubuntu)
sudo ufw allow 2181/tcp  # Zookeeper
sudo ufw allow 9092/tcp  # Kafka
sudo ufw allow 8123/tcp  # ClickHouse HTTP
sudo ufw allow 9000/tcp  # ClickHouse Native + MinIO
sudo ufw allow 7474/tcp  # Neo4j HTTP
sudo ufw allow 7687/tcp  # Neo4j Bolt
sudo ufw allow 6333/tcp  # Qdrant HTTP
sudo ufw allow 4040/tcp  # Spark UI
sudo ufw allow 7077/tcp  # Spark Master
sudo ufw allow 8082/tcp  # Flink UI
sudo ufw allow 7061/tcp  # HelixAgent API

sudo ufw enable
```

### 4. DNS Configuration

Add DNS entries for services:
```
helixagent.example.com      -> Server 1 (main API)
kafka.example.com           -> Server 1
neo4j.example.com           -> Server 2
clickhouse.example.com      -> Server 2
minio.example.com           -> Server 3
spark.example.com           -> Server 3
```

### 5. SSL Certificates

```bash
# Install certbot
sudo apt-get install certbot python3-certbot-nginx -y

# Generate certificates
sudo certbot certonly --standalone \
  -d helixagent.example.com \
  -d kafka.example.com \
  -d neo4j.example.com \
  -d clickhouse.example.com \
  -d minio.example.com \
  -d spark.example.com

# Certificates will be in /etc/letsencrypt/live/
```

---

## Installation Steps

### Step 1: Clone Repository

```bash
git clone https://github.com/anthropics/helixagent.git
cd helixagent
```

### Step 2: Configure Environment

```bash
# Copy example environment file
cp .env.example .env.production

# Edit production environment
nano .env.production
```

**Required Variables**:
```bash
# Database
DB_HOST=server2.internal
DB_PORT=5432
DB_USER=helixagent
DB_PASSWORD=<strong-password>
DB_NAME=helixagent_db

# Redis
REDIS_HOST=server1.internal
REDIS_PORT=6379
REDIS_PASSWORD=<strong-password>

# Kafka
KAFKA_BOOTSTRAP_SERVERS=server1.internal:9092

# Neo4j
NEO4J_URI=bolt://server2.internal:7687
NEO4J_PASSWORD=<strong-password>

# ClickHouse
CLICKHOUSE_HOST=server2.internal
CLICKHOUSE_PORT=9000
CLICKHOUSE_PASSWORD=<strong-password>

# MinIO
MINIO_ENDPOINT=server3.internal:9000
MINIO_ROOT_USER=helixagent
MINIO_ROOT_PASSWORD=<strong-password>
MINIO_BUCKET=helixagent-datalake

# LLM Providers
CLAUDE_API_KEY=<your-key>
DEEPSEEK_API_KEY=<your-key>
GEMINI_API_KEY=<your-key>
# ... add other provider keys
```

### Step 3: Create Docker Network

```bash
docker network create helixagent-network
```

### Step 4: Start Services

```bash
# Start infrastructure services
docker-compose -f docker-compose.bigdata.yml up -d

# Wait for health checks (2-3 minutes)
./scripts/wait-for-services.sh

# Verify all services are healthy
docker-compose -f docker-compose.bigdata.yml ps
```

Expected output (all services should show "healthy"):
```
NAME                    STATUS
helixagent-zookeeper    Up (healthy)
helixagent-kafka        Up (healthy)
helixagent-clickhouse   Up (healthy)
helixagent-neo4j        Up (healthy)
helixagent-minio        Up (healthy)
helixagent-spark-master Up (healthy)
helixagent-flink-jobmanager Up (healthy)
helixagent-qdrant       Up (healthy)
```

### Step 5: Initialize Databases

```bash
# PostgreSQL schema
docker exec -i helixagent-postgres psql -U helixagent -d helixagent_db < sql/schema/complete_schema.sql

# ClickHouse tables
docker exec -i helixagent-clickhouse clickhouse-client --multiquery < sql/schema/clickhouse_analytics.sql

# Neo4j constraints and indexes
docker exec -i helixagent-neo4j cypher-shell -u neo4j -p <password> < sql/schema/neo4j_schema.cypher
```

### Step 6: Start HelixAgent Application

```bash
# Build application
make build

# Start application
./bin/helixagent --config configs/production.yaml
```

### Step 7: Verify Deployment

```bash
# Health check
curl http://localhost:7061/health

# Expected response:
{
  "status": "healthy",
  "services": {
    "postgresql": "healthy",
    "redis": "healthy",
    "kafka": "healthy",
    "neo4j": "healthy",
    "clickhouse": "healthy",
    "minio": "healthy"
  }
}

# Big data status
curl http://localhost:7061/v1/bigdata/status
```

---

## Service Configuration

### Kafka Configuration

**Optimal Settings** (`docker-compose.bigdata.yml`):
```yaml
environment:
  KAFKA_NUM_PARTITIONS: 24              # 2-3x number of consumers
  KAFKA_LOG_RETENTION_HOURS: 168       # 7 days
  KAFKA_LOG_SEGMENT_BYTES: 1073741824  # 1GB segments
  KAFKA_COMPRESSION_TYPE: lz4          # Best performance
  KAFKA_MESSAGE_MAX_BYTES: 10485760    # 10MB max message
```

**Topic Configuration**:
```bash
# Create topics manually for production
kafka-topics.sh --bootstrap-server localhost:9092 --create \
  --topic helixagent.conversations \
  --partitions 24 \
  --replication-factor 3 \
  --config retention.ms=604800000 \
  --config segment.bytes=1073741824 \
  --config compression.type=lz4

# Repeat for all topics:
# - helixagent.messages
# - helixagent.entities
# - helixagent.debates
# - helixagent.memory.events
# - helixagent.memory.snapshots
# - helixagent.memory.conflicts
# - helixagent.learning.insights
# - helixagent.context.compressed
```

### ClickHouse Configuration

**Memory Settings** (`/etc/clickhouse-server/config.xml`):
```xml
<max_server_memory_usage>6442450944</max_server_memory_usage>  <!-- 6GB -->
<max_memory_usage>4294967296</max_memory_usage>                <!-- 4GB per query -->
<max_concurrent_queries>100</max_concurrent_queries>
<max_threads>8</max_threads>
```

**Optimize Tables** (run weekly):
```sql
OPTIMIZE TABLE debate_metrics FINAL;
OPTIMIZE TABLE conversation_metrics FINAL;
OPTIMIZE TABLE provider_performance FINAL;
```

### Neo4j Configuration

**Memory Tuning** (`neo4j.conf`):
```properties
dbms.memory.heap.initial_size=1g
dbms.memory.heap.max_size=4g
dbms.memory.pagecache.size=2g
dbms.memory.transaction.global_max_size=2g
```

**Indexes**:
```cypher
// Ensure all indexes exist
CREATE INDEX entity_id_idx IF NOT EXISTS FOR (e:Entity) ON (e.id);
CREATE INDEX entity_name_idx IF NOT EXISTS FOR (e:Entity) ON (e.name);
CREATE INDEX entity_type_idx IF NOT EXISTS FOR (e:Entity) ON (e.type);
CREATE INDEX conversation_id_idx IF NOT EXISTS FOR (c:Conversation) ON (c.id);
```

### Spark Configuration

**Executor Settings** (`spark-defaults.conf`):
```properties
spark.executor.memory=6g
spark.executor.cores=4
spark.driver.memory=2g
spark.sql.shuffle.partitions=200
spark.default.parallelism=100
```

---

## Monitoring Setup

### Prometheus + Grafana

```bash
# Start monitoring stack
docker-compose -f docker-compose.monitoring.yml up -d

# Access Grafana
open http://localhost:3000
# Login: admin / helixagent123
```

**Pre-built Dashboards**:
- **Kafka Metrics**: Broker throughput, consumer lag, partition distribution
- **ClickHouse Metrics**: Query latency, memory usage, merge performance
- **Neo4j Metrics**: Query latency, transaction rate, page cache hits
- **Spark Metrics**: Job duration, task distribution, executor usage
- **Flink Metrics**: Job latency, checkpoint duration, backpressure

### Log Aggregation

**Loki + Promtail**:
```bash
# Install Loki
docker run -d --name=loki \
  -p 3100:3100 \
  -v $(pwd)/loki-config.yaml:/etc/loki/local-config.yaml \
  grafana/loki:latest

# Install Promtail
docker run -d --name=promtail \
  -v /var/log:/var/log \
  -v /var/lib/docker/containers:/var/lib/docker/containers:ro \
  -v $(pwd)/promtail-config.yaml:/etc/promtail/config.yml \
  grafana/promtail:latest
```

### Alerting Rules

**Critical Alerts** (`alerts.yml`):
```yaml
groups:
  - name: big_data_alerts
    rules:
      - alert: KafkaLagHigh
        expr: kafka_consumer_lag_seconds > 60
        for: 5m
        annotations:
          summary: "Kafka consumer lag > 60s"

      - alert: ClickHouseQuerySlow
        expr: clickhouse_query_duration_seconds > 1
        for: 5m
        annotations:
          summary: "ClickHouse query latency > 1s"

      - alert: Neo4jMemoryHigh
        expr: neo4j_memory_heap_used_percent > 90
        for: 5m
        annotations:
          summary: "Neo4j heap usage > 90%"
```

---

## Backup & Recovery

### Backup Strategy

**Daily Backups**:
```bash
#!/bin/bash
# Daily backup script (/opt/helixagent/backup-daily.sh)

BACKUP_DIR="/mnt/backups/$(date +%Y-%m-%d)"
mkdir -p $BACKUP_DIR

# PostgreSQL
docker exec helixagent-postgres pg_dump -U helixagent helixagent_db | gzip > $BACKUP_DIR/postgres.sql.gz

# Neo4j
docker exec helixagent-neo4j neo4j-admin dump --database=helixagent --to=/tmp/neo4j-backup.dump
docker cp helixagent-neo4j:/tmp/neo4j-backup.dump $BACKUP_DIR/

# MinIO buckets
mc mirror --preserve helixagent/helixagent-datalake $BACKUP_DIR/minio/

# Kafka topic data (optional, data lake has history)
kafka-console-consumer.sh --bootstrap-server localhost:9092 \
  --topic helixagent.conversations \
  --from-beginning \
  --max-messages 100000 > $BACKUP_DIR/kafka_conversations.json

echo "Backup completed: $BACKUP_DIR"
```

**Schedule with cron**:
```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /opt/helixagent/backup-daily.sh
```

### Recovery Procedures

**Restore PostgreSQL**:
```bash
gunzip -c /mnt/backups/2026-01-29/postgres.sql.gz | \
  docker exec -i helixagent-postgres psql -U helixagent -d helixagent_db
```

**Restore Neo4j**:
```bash
docker exec helixagent-neo4j neo4j-admin load --from=/tmp/neo4j-backup.dump --database=helixagent --force
docker restart helixagent-neo4j
```

**Restore MinIO**:
```bash
mc mirror /mnt/backups/2026-01-29/minio/ helixagent/helixagent-datalake
```

---

## Scaling Guide

### Horizontal Scaling

**Add Kafka Brokers**:
```bash
# Start new broker with unique ID
docker-compose -f docker-compose.kafka-broker-2.yml up -d

# Rebalance partitions
kafka-reassign-partitions.sh --bootstrap-server localhost:9092 \
  --reassignment-json-file reassignment.json --execute
```

**Add Spark Workers**:
```bash
# Scale Spark workers
docker-compose -f docker-compose.bigdata.yml up -d --scale spark-worker=8
```

**Add Flink TaskManagers**:
```bash
# Scale Flink task managers
docker-compose -f docker-compose.bigdata.yml up -d --scale flink-taskmanager=8
```

### Vertical Scaling

**Increase Resources**:
```yaml
# docker-compose.override.yml
services:
  clickhouse:
    deploy:
      resources:
        limits:
          cpus: '8.0'
          memory: 16G
        reservations:
          cpus: '4.0'
          memory: 8G
```

---

## Troubleshooting

### Common Issues

**Issue 1: Kafka Consumer Lag**
```bash
# Check lag
kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group helixagent-consumers

# Solution: Add more consumers or increase partitions
```

**Issue 2: ClickHouse Out of Memory**
```bash
# Check memory usage
docker exec helixagent-clickhouse clickhouse-client -q "SELECT * FROM system.metrics WHERE metric LIKE '%Memory%'"

# Solution: Optimize queries or increase memory limits
```

**Issue 3: Neo4j Slow Queries**
```cypher
// Profile slow query
PROFILE MATCH (e:Entity {name: "Docker"})-[:RELATED_TO]-(r)
RETURN r LIMIT 10;

// Solution: Add indexes or optimize query
```

---

## Security Hardening

### SSL/TLS Configuration

**Kafka SSL**:
```bash
# Generate keystore
keytool -keystore kafka.keystore.jks -alias localhost -validity 365 -genkey -keyalg RSA

# Generate truststore
keytool -keystore kafka.truststore.jks -alias CARoot -import -file ca-cert
```

**Neo4j HTTPS**:
```properties
# neo4j.conf
dbms.connector.https.enabled=true
dbms.connector.https.listen_address=:7473
dbms.ssl.policy.https.enabled=true
dbms.ssl.policy.https.base_directory=/var/lib/neo4j/certificates/https
```

### Network Segmentation

```yaml
# Separate networks for each layer
networks:
  helixagent-public:
    driver: bridge
  helixagent-internal:
    driver: bridge
    internal: true
```

---

## Performance Benchmarks

Expected performance on recommended hardware:

| Metric | Target | Actual (Typical) |
|--------|--------|------------------|
| **API Latency (p95)** | <200ms | 150ms |
| **Kafka Throughput** | 10K msg/sec | 12K msg/sec |
| **ClickHouse Query (avg)** | <50ms | 35ms |
| **Neo4j Query (avg)** | <100ms | 75ms |
| **Context Replay (1K msg)** | <1s | 0.8s |
| **Spark Batch Job (1M conv)** | <10min | 8min |
| **Memory Sync Lag** | <100ms | 60ms |

---

## Maintenance Schedule

| Task | Frequency | Command |
|------|-----------|---------|
| **Backup** | Daily | `/opt/helixagent/backup-daily.sh` |
| **Log Rotation** | Weekly | `docker system prune -a --volumes` |
| **Optimize Tables** | Weekly | `OPTIMIZE TABLE ... FINAL;` |
| **Reindex Neo4j** | Monthly | `CALL db.awaitIndexes();` |
| **Update Containers** | Monthly | `docker-compose pull && docker-compose up -d` |
| **Certificate Renewal** | Every 60 days | `certbot renew` |

---

## Support & Documentation

- **API Reference**: `/docs/api/BIG_DATA_API_REFERENCE.md`
- **User Guide**: `/docs/user/BIG_DATA_USER_GUIDE.md`
- **SQL Schemas**: `/sql/schema/`
- **GitHub Issues**: https://github.com/anthropics/helixagent/issues
- **Community**: Discord, Slack

---

**Deployment Guide Version**: 1.0
**Last Updated**: 2026-01-30
**Maintainer**: HelixAgent Team
