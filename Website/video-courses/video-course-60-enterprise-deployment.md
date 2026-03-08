# Video Course 60: Enterprise Deployment

## Course Overview

**Duration**: 2 hours 30 minutes
**Level**: Advanced
**Prerequisites**: Course 01-Fundamentals, Course 03-Deployment, Course 13-Enterprise-Deployment, Kubernetes experience required

This course covers production-grade enterprise deployment of HelixAgent, including multi-region Kubernetes deployment, high availability configuration, security hardening, backup and disaster recovery, SLA management, and a hands-on deployment to a Kubernetes cluster.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Deploy HelixAgent to a multi-region Kubernetes cluster with proper resource management
2. Configure high availability for all components (application, database, cache)
3. Apply security hardening for production environments
4. Implement backup procedures and disaster recovery plans
5. Define and monitor SLAs using the observability stack
6. Execute a complete deployment to a Kubernetes cluster

---

## Module 1: Multi-Region Kubernetes Deployment (30 min)

### 1.1 Architecture Overview

**Video: Multi-Region Design** (10 min)

```
Region A (Primary)                Region B (Secondary)
+------------------+              +------------------+
| K8s Cluster      |              | K8s Cluster      |
| +------+ +-----+|              | +------+ +-----+||
| |Helix | |Helix||  <-- DNS --> | |Helix | |Helix|||
| |Agent | |Agent||  Load Bal.   | |Agent | |Agent|||
| +------+ +-----+|              | +------+ +-----+||
| +------+ +-----+|              | +------+ +-----+||
| |PG Pri| |Redis||              | |PG Rep| |Redis|||
| +------+ +-----+|              | +------+ +-----+||
+------------------+              +------------------+
```

- Active-active or active-passive deployment models
- DNS-based global load balancing across regions
- Database replication from primary to secondary region
- Redis cluster with cross-region replication for cache consistency

### 1.2 Kubernetes Manifests

**Video: Core Deployment Resources** (10 min)

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: helixagent
  namespace: helixagent-prod
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: helixagent
  template:
    metadata:
      labels:
        app: helixagent
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - labelSelector:
                matchLabels:
                  app: helixagent
              topologyKey: kubernetes.io/hostname
      containers:
        - name: helixagent
          image: helixagent:latest
          ports:
            - containerPort: 7061
              name: http
            - containerPort: 9090
              name: metrics
          resources:
            requests:
              cpu: "500m"
              memory: "512Mi"
            limits:
              cpu: "2000m"
              memory: "2Gi"
          livenessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health/ready
              port: http
            initialDelaySeconds: 10
            periodSeconds: 5
          env:
            - name: DB_HOST
              valueFrom:
                secretKeyRef:
                  name: helixagent-db
                  key: host
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: helixagent-db
                  key: password
```

### 1.3 Horizontal Pod Autoscaling

**Video: Scaling Under Load** (10 min)

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: helixagent
  namespace: helixagent-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: helixagent
  minReplicas: 3
  maxReplicas: 12
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Pods
      pods:
        metric:
          name: helixagent_requests_per_second
        target:
          type: AverageValue
          averageValue: "100"
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
        - type: Percent
          value: 50
          periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 25
          periodSeconds: 120
```

### Hands-On Lab 1

Deploy HelixAgent to a Kubernetes cluster:

```bash
# Create namespace
kubectl create namespace helixagent-prod

# Apply secrets
kubectl apply -f k8s/secrets/ -n helixagent-prod

# Deploy database and Redis
kubectl apply -f k8s/infrastructure/ -n helixagent-prod

# Deploy HelixAgent
kubectl apply -f k8s/application/ -n helixagent-prod

# Verify deployment
kubectl get pods -n helixagent-prod
kubectl get svc -n helixagent-prod
```

---

## Module 2: High Availability Configuration (25 min)

### 2.1 PostgreSQL High Availability

**Video: Database HA with Patroni** (10 min)

- Patroni manages PostgreSQL leader election and failover
- Streaming replication from primary to standby nodes
- Connection pooling with PgBouncer for connection management
- Automatic failover completes within 10-30 seconds

```yaml
# PostgreSQL StatefulSet with Patroni
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgresql
spec:
  replicas: 3
  serviceName: postgresql
  template:
    spec:
      containers:
        - name: postgresql
          image: postgres:15-alpine
          resources:
            requests:
              cpu: "1000m"
              memory: "2Gi"
            limits:
              cpu: "4000m"
              memory: "8Gi"
          volumeMounts:
            - name: data
              mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 100Gi
        storageClassName: fast-ssd
```

### 2.2 Redis High Availability

**Video: Redis Sentinel Configuration** (8 min)

- Redis Sentinel monitors master and automatically promotes replicas
- Minimum 3 Sentinel instances for quorum
- Application connects through Sentinel for automatic master discovery
- Failover typically completes in 5-15 seconds

### 2.3 Application-Level HA

**Video: Pod Disruption Budgets and Anti-Affinity** (7 min)

```yaml
# pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: helixagent
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: helixagent
```

- Pod anti-affinity spreads replicas across nodes and zones
- PodDisruptionBudget prevents draining too many pods during maintenance
- Rolling updates with zero downtime (maxUnavailable: 0)

### Hands-On Lab 2

Configure and test high availability:

1. Deploy a 3-node PostgreSQL cluster with Patroni
2. Deploy Redis Sentinel with 3 replicas
3. Verify automatic failover by killing the primary PostgreSQL pod
4. Confirm HelixAgent reconnects to the new primary within 30 seconds
5. Kill a Redis master and verify Sentinel promotes a replica

---

## Module 3: Security Hardening (25 min)

### 3.1 Network Policies

**Video: Restricting Network Access** (8 min)

```yaml
# network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: helixagent-ingress
  namespace: helixagent-prod
spec:
  podSelector:
    matchLabels:
      app: helixagent
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              role: ingress
      ports:
        - port: 7061
          protocol: TCP
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: postgresql
      ports:
        - port: 5432
    - to:
        - podSelector:
            matchLabels:
              app: redis
      ports:
        - port: 6379
    - to: # Allow outbound to LLM provider APIs
        - ipBlock:
            cidr: 0.0.0.0/0
      ports:
        - port: 443
          protocol: TCP
```

### 3.2 Secret Management

**Video: Secure Secret Handling** (8 min)

- Kubernetes Secrets encrypted at rest with EncryptionConfiguration
- External secret operators (Vault, AWS Secrets Manager, GCP Secret Manager)
- Secret rotation without pod restart using volume-mounted secrets
- API key injection via environment variables from sealed secrets

### 3.3 Container Security

**Video: Hardening Container Runtime** (9 min)

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  seccompProfile:
    type: RuntimeDefault
```

- Non-root container execution
- Read-only root filesystem with writable tmpfs mounts for temporary files
- Dropped capabilities prevent privilege escalation
- Security scanning of container images in CI pipeline

### Hands-On Lab 3

Apply security hardening to the deployment:

1. Apply network policies restricting ingress and egress
2. Configure secret encryption at rest
3. Enable non-root container execution
4. Run a Trivy scan on the production container image
5. Verify the application functions correctly with all restrictions

---

## Module 4: Backup and Disaster Recovery (25 min)

### 4.1 Database Backup Strategy

**Video: PostgreSQL Backup and Recovery** (10 min)

```bash
# Continuous WAL archiving for point-in-time recovery
archive_mode = on
archive_command = 'aws s3 cp %p s3://helixagent-backups/wal/%f'

# Scheduled base backup (daily)
pg_basebackup -h postgresql-primary -U replication \
  -D /backup/base --checkpoint=fast --wal-method=stream

# Point-in-time recovery
recovery_target_time = '2026-03-08 12:00:00'
```

- WAL archiving enables point-in-time recovery to any second
- Daily base backups stored in object storage (S3/GCS/MinIO)
- Backup verification: automated restore test weekly
- Retention policy: 7 daily, 4 weekly, 12 monthly backups

### 4.2 Configuration Backup

**Video: Backing Up Non-Database State** (7 min)

- Kubernetes manifests stored in version control (GitOps)
- ConfigMaps and Secrets exported with `velero backup`
- .env files and configuration YAML stored in encrypted vault
- LLMsVerifier verification scores archived for audit trail

### 4.3 Disaster Recovery Plan

**Video: RTO and RPO Definition** (8 min)

| Scenario               | RTO Target | RPO Target | Recovery Method                    |
|------------------------|------------|------------|------------------------------------|
| Single pod failure     | < 30s      | 0          | Kubernetes auto-restart            |
| Single node failure    | < 2 min    | 0          | Pod rescheduling + DB replication  |
| Single region failure  | < 15 min   | < 1 min    | DNS failover to secondary region   |
| Complete data loss     | < 4 hours  | < 1 hour   | Restore from backup + WAL replay   |

- RTO (Recovery Time Objective): maximum acceptable downtime
- RPO (Recovery Point Objective): maximum acceptable data loss
- DR drills: execute recovery procedure monthly
- Runbook documentation for each failure scenario

### Hands-On Lab 4

Implement and test a backup/restore procedure:

1. Configure automated PostgreSQL base backup
2. Perform a base backup to local storage
3. Insert test data after the backup
4. Simulate disaster by stopping the database
5. Restore from backup and verify data integrity
6. Document recovery time and any data gap

---

## Module 5: SLA Management and Monitoring (25 min)

### 5.1 Defining SLIs, SLOs, and SLAs

**Video: Service Level Framework** (10 min)

| SLI (Indicator)           | SLO (Objective)      | SLA (Agreement)       |
|---------------------------|----------------------|-----------------------|
| Request success rate      | 99.9% per month      | 99.5% with credits    |
| P95 response latency      | < 3 seconds          | < 5 seconds           |
| Provider availability     | 3+ providers healthy | 2+ providers healthy  |
| API uptime                | 99.95% per month     | 99.9% per month       |

- SLIs are measurements (what you observe)
- SLOs are targets (what you aim for)
- SLAs are contracts (what you guarantee to customers)

### 5.2 Error Budget Tracking

**Video: Managing Error Budgets** (8 min)

```promql
# Monthly error budget remaining
1 - (
  sum(rate(helixagent_requests_total{status="error"}[30d]))
  / sum(rate(helixagent_requests_total[30d]))
) - 0.999
```

- Error budget = 1 - SLO (e.g., 0.1% for 99.9% SLO)
- Track budget consumption over rolling windows
- Alert when budget consumption rate exceeds sustainable pace
- Freeze deployments when error budget is exhausted

### 5.3 SLA Dashboards

**Video: Executive Visibility** (7 min)

- Monthly uptime percentage with daily granularity
- P95 latency trend with SLO threshold line
- Provider availability heatmap by hour and day
- Error budget burn-down chart
- Incident timeline with impact and duration

### Hands-On Lab 5

Deploy HelixAgent to a Kubernetes cluster end to end:

1. Create the production namespace with resource quotas
2. Deploy secrets, PostgreSQL, Redis infrastructure
3. Deploy HelixAgent with 3 replicas and HPA
4. Apply network policies and security hardening
5. Configure Prometheus scraping and Grafana dashboards
6. Set up alerting rules for SLO violations
7. Run a smoke test to verify the deployment
8. Simulate a pod failure and observe automatic recovery

---

## Course Summary

### Key Takeaways

1. Multi-region Kubernetes deployment provides geographic redundancy and lower latency
2. High availability requires HA at every layer: application (replicas + PDB), database (Patroni), cache (Sentinel)
3. Security hardening includes network policies, secret encryption, non-root containers, and capability dropping
4. Backup strategy must cover database (WAL archiving + base backups), configuration (GitOps), and secrets (vault)
5. SLA management requires clear SLI/SLO/SLA definitions with error budget tracking
6. DR plans must be tested monthly with documented RTO/RPO targets per failure scenario

### Assessment Questions

1. What is the purpose of pod anti-affinity in a Kubernetes deployment?
2. How does a PodDisruptionBudget protect service availability during maintenance?
3. Describe the difference between RTO and RPO with examples.
4. What network policy rules are necessary for HelixAgent to function?
5. How do you calculate and track an error budget for a 99.9% availability SLO?

### Related Courses

- Course 03: Deployment
- Course 13: Enterprise Deployment
- Course 48: Backup and Recovery
- Course 55: Security Scanning Pipeline
- Course 59: Monitoring and Observability

---

**Course Version**: 1.0
**Last Updated**: March 8, 2026
