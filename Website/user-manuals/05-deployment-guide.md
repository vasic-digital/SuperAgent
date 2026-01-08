# HelixAgent Deployment Guide

## Introduction

This guide provides comprehensive instructions for deploying HelixAgent in production environments. It covers Docker, Kubernetes, cloud-native deployments, high availability configurations, and operational best practices for running HelixAgent at scale.

---

## Table of Contents

1. [Deployment Overview](#deployment-overview)
2. [Pre-Deployment Checklist](#pre-deployment-checklist)
3. [Docker Deployment](#docker-deployment)
4. [Kubernetes Deployment](#kubernetes-deployment)
5. [Cloud Provider Deployments](#cloud-provider-deployments)
6. [High Availability Configuration](#high-availability-configuration)
7. [Load Balancing](#load-balancing)
8. [Database Configuration](#database-configuration)
9. [Monitoring and Observability](#monitoring-and-observability)
10. [Security Hardening](#security-hardening)
11. [Backup and Disaster Recovery](#backup-and-disaster-recovery)
12. [Performance Tuning](#performance-tuning)
13. [Troubleshooting Deployments](#troubleshooting-deployments)

---

## Deployment Overview

### Deployment Options

| Option | Complexity | Scalability | Best For |
|--------|------------|-------------|----------|
| Docker Compose | Low | Single node | Development, small teams |
| Docker Swarm | Medium | Multi-node | Medium-scale production |
| Kubernetes | High | Enterprise | Large-scale, HA requirements |
| Cloud Managed | Medium | Auto-scale | Cloud-native deployments |

### Architecture Overview

```
                      ┌───────────────────────────────────────────┐
                      │           Load Balancer (L7)              │
                      │         (nginx/traefik/cloud)             │
                      └─────────────────┬─────────────────────────┘
                                        │
              ┌─────────────────────────┼─────────────────────────┐
              │                         │                         │
              ▼                         ▼                         ▼
     ┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
     │  HelixAgent-1   │      │  HelixAgent-2   │      │  HelixAgent-3   │
     │    (Replica)    │      │    (Replica)    │      │    (Replica)    │
     └────────┬────────┘      └────────┬────────┘      └────────┬────────┘
              │                         │                         │
              └─────────────────────────┼─────────────────────────┘
                                        │
              ┌─────────────────────────┼─────────────────────────┐
              │                         │                         │
              ▼                         ▼                         ▼
     ┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
     │   PostgreSQL    │      │     Redis       │      │     Cognee      │
     │   (Primary)     │      │   (Cluster)     │      │   (Optional)    │
     └─────────────────┘      └─────────────────┘      └─────────────────┘
```

---

## Pre-Deployment Checklist

### Infrastructure Requirements

```yaml
# Minimum Production Requirements
infrastructure:
  compute:
    cpu: 8 cores
    memory: 32 GB
    storage: 200 GB SSD

  network:
    bandwidth: 1 Gbps
    latency: < 50ms to LLM providers

  database:
    postgresql: 15+
    storage: 100 GB SSD
    connections: 100+

  cache:
    redis: 7+
    memory: 4 GB
```

### Security Requirements

- [ ] TLS certificates obtained (Let's Encrypt or CA-signed)
- [ ] API keys secured and rotated
- [ ] Firewall rules configured
- [ ] Secret management solution in place
- [ ] Network policies defined
- [ ] Audit logging enabled

### Configuration Requirements

- [ ] Environment variables prepared
- [ ] Configuration files validated
- [ ] Database migrations ready
- [ ] Backup strategy defined
- [ ] Monitoring configured

---

## Docker Deployment

### Production Docker Compose

Create a production-ready `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  helixagent:
    image: helixagent/helixagent:latest
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
    environment:
      - PORT=8080
      - GIN_MODE=release
      - JWT_SECRET=${JWT_SECRET}
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=helixagent_prod
      - DB_SSLMODE=require
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - CLAUDE_API_KEY=${CLAUDE_API_KEY}
      - DEEPSEEK_API_KEY=${DEEPSEEK_API_KEY}
      - GEMINI_API_KEY=${GEMINI_API_KEY}
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"
    networks:
      - helixagent-network

  postgres:
    image: postgres:15-alpine
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
    environment:
      - POSTGRES_DB=helixagent_prod
      - POSTGRES_USER=${DB_USER}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql
    command:
      - "postgres"
      - "-c"
      - "max_connections=200"
      - "-c"
      - "shared_buffers=256MB"
      - "-c"
      - "effective_cache_size=1GB"
      - "-c"
      - "maintenance_work_mem=128MB"
      - "-c"
      - "checkpoint_completion_target=0.9"
      - "-c"
      - "wal_buffers=16MB"
      - "-c"
      - "default_statistics_target=100"
      - "-c"
      - "random_page_cost=1.1"
      - "-c"
      - "effective_io_concurrency=200"
      - "-c"
      - "work_mem=16MB"
      - "-c"
      - "min_wal_size=1GB"
      - "-c"
      - "max_wal_size=4GB"
      - "-c"
      - "max_worker_processes=8"
      - "-c"
      - "max_parallel_workers_per_gather=4"
      - "-c"
      - "max_parallel_workers=8"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d helixagent_prod"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - helixagent-network

  redis:
    image: redis:7-alpine
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 2G
    command: >
      redis-server
      --requirepass ${REDIS_PASSWORD}
      --appendonly yes
      --maxmemory 1gb
      --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - helixagent-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
    depends_on:
      - helixagent
    networks:
      - helixagent-network

volumes:
  postgres_data:
  redis_data:

networks:
  helixagent-network:
    driver: bridge
```

### Nginx Configuration

Create `nginx/nginx.conf`:

```nginx
events {
    worker_connections 4096;
}

http {
    upstream helixagent {
        least_conn;
        server helixagent:8080 weight=1 max_fails=3 fail_timeout=30s;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_conn_zone $binary_remote_addr zone=conn:10m;

    server {
        listen 80;
        server_name _;
        return 301 https://$host$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name api.helixagent.example.com;

        ssl_certificate /etc/nginx/ssl/fullchain.pem;
        ssl_certificate_key /etc/nginx/ssl/privkey.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
        ssl_prefer_server_ciphers off;

        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";

        location / {
            limit_req zone=api burst=20 nodelay;
            limit_conn conn 10;

            proxy_pass http://helixagent;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # WebSocket support for streaming
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";

            # Timeouts
            proxy_connect_timeout 60s;
            proxy_send_timeout 300s;
            proxy_read_timeout 300s;
        }

        location /health {
            proxy_pass http://helixagent/health;
            access_log off;
        }
    }
}
```

### Deployment Commands

```bash
# Create environment file
cp .env.example .env.prod
# Edit .env.prod with production values

# Deploy
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f helixagent

# Scale replicas
docker-compose -f docker-compose.prod.yml up -d --scale helixagent=5

# Rolling update
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d --no-deps helixagent
```

---

## Kubernetes Deployment

### Namespace and ConfigMap

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: helixagent
  labels:
    name: helixagent

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: helixagent-config
  namespace: helixagent
data:
  PORT: "8080"
  GIN_MODE: "release"
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "helixagent_prod"
  DB_SSLMODE: "require"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
```

### Secrets

```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: helixagent-secrets
  namespace: helixagent
type: Opaque
stringData:
  JWT_SECRET: "your-jwt-secret-here"
  DB_USER: "helixagent"
  DB_PASSWORD: "your-db-password"
  REDIS_PASSWORD: "your-redis-password"
  CLAUDE_API_KEY: "sk-ant-..."
  DEEPSEEK_API_KEY: "sk-..."
  GEMINI_API_KEY: "..."
```

### Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: helixagent
  namespace: helixagent
  labels:
    app: helixagent
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
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: helixagent
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
        - name: helixagent
          image: helixagent/helixagent:v1.0.0
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
              name: http
          envFrom:
            - configMapRef:
                name: helixagent-config
            - secretRef:
                name: helixagent-secrets
          resources:
            requests:
              cpu: "500m"
              memory: "1Gi"
            limits:
              cpu: "2000m"
              memory: "4Gi"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop:
                - ALL
          volumeMounts:
            - name: tmp
              mountPath: /tmp
            - name: logs
              mountPath: /app/logs
      volumes:
        - name: tmp
          emptyDir: {}
        - name: logs
          emptyDir: {}
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: app
                      operator: In
                      values:
                        - helixagent
                topologyKey: kubernetes.io/hostname
```

### Service

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: helixagent-service
  namespace: helixagent
  labels:
    app: helixagent
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: 8080
      protocol: TCP
      name: http
  selector:
    app: helixagent
```

### Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: helixagent-ingress
  namespace: helixagent
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  tls:
    - hosts:
        - api.helixagent.example.com
      secretName: helixagent-tls
  rules:
    - host: api.helixagent.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: helixagent-service
                port:
                  number: 80
```

### Horizontal Pod Autoscaler

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: helixagent-hpa
  namespace: helixagent
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: helixagent
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
        - type: Percent
          value: 100
          periodSeconds: 15
        - type: Pods
          value: 4
          periodSeconds: 15
      selectPolicy: Max
```

### PostgreSQL StatefulSet

```yaml
# k8s/postgres.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: helixagent
spec:
  serviceName: postgres-service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
        - name: postgres
          image: postgres:15-alpine
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_DB
              value: helixagent_prod
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: helixagent-secrets
                  key: DB_USER
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: helixagent-secrets
                  key: DB_PASSWORD
          volumeMounts:
            - name: postgres-data
              mountPath: /var/lib/postgresql/data
          resources:
            requests:
              cpu: "500m"
              memory: "1Gi"
            limits:
              cpu: "2000m"
              memory: "4Gi"
  volumeClaimTemplates:
    - metadata:
        name: postgres-data
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName: standard
        resources:
          requests:
            storage: 100Gi

---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: helixagent
spec:
  ports:
    - port: 5432
  selector:
    app: postgres
  clusterIP: None
```

### Deployment Commands

```bash
# Apply all resources
kubectl apply -f k8s/

# Check deployment status
kubectl -n helixagent get pods
kubectl -n helixagent get deployments
kubectl -n helixagent get services

# View logs
kubectl -n helixagent logs -f deployment/helixagent

# Scale manually
kubectl -n helixagent scale deployment/helixagent --replicas=5

# Rolling restart
kubectl -n helixagent rollout restart deployment/helixagent

# Check rollout status
kubectl -n helixagent rollout status deployment/helixagent
```

---

## Cloud Provider Deployments

### AWS (EKS)

```bash
# Create EKS cluster
eksctl create cluster \
  --name helixagent-prod \
  --region us-east-1 \
  --nodegroup-name standard-workers \
  --node-type m5.xlarge \
  --nodes 3 \
  --nodes-min 3 \
  --nodes-max 10 \
  --managed

# Install AWS Load Balancer Controller
helm repo add eks https://aws.github.io/eks-charts
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=helixagent-prod

# Deploy HelixAgent
kubectl apply -f k8s/
```

### GCP (GKE)

```bash
# Create GKE cluster
gcloud container clusters create helixagent-prod \
  --zone us-central1-a \
  --num-nodes 3 \
  --machine-type e2-standard-4 \
  --enable-autoscaling \
  --min-nodes 3 \
  --max-nodes 10

# Get credentials
gcloud container clusters get-credentials helixagent-prod --zone us-central1-a

# Deploy HelixAgent
kubectl apply -f k8s/
```

### Azure (AKS)

```bash
# Create AKS cluster
az aks create \
  --resource-group helixagent-rg \
  --name helixagent-prod \
  --node-count 3 \
  --node-vm-size Standard_D4s_v3 \
  --enable-cluster-autoscaler \
  --min-count 3 \
  --max-count 10

# Get credentials
az aks get-credentials --resource-group helixagent-rg --name helixagent-prod

# Deploy HelixAgent
kubectl apply -f k8s/
```

---

## High Availability Configuration

### Multi-Region Deployment

```yaml
# Multi-region architecture
regions:
  - name: us-east-1
    role: primary
    services:
      - helixagent (3 replicas)
      - postgres (primary)
      - redis (cluster)

  - name: us-west-2
    role: secondary
    services:
      - helixagent (3 replicas)
      - postgres (replica)
      - redis (cluster)

  - name: eu-west-1
    role: secondary
    services:
      - helixagent (2 replicas)
      - postgres (replica)
      - redis (cluster)
```

### Database Replication

```yaml
# PostgreSQL replication configuration
postgres:
  primary:
    host: postgres-primary.us-east-1
    port: 5432
    replication:
      synchronous_commit: on
      max_wal_senders: 10
      wal_level: replica

  replicas:
    - host: postgres-replica.us-west-2
      port: 5432
      replication_slot: replica_slot_west
    - host: postgres-replica.eu-west-1
      port: 5432
      replication_slot: replica_slot_eu
```

---

## Monitoring and Observability

### Prometheus Configuration

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'helixagent'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - helixagent
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
```

### Grafana Dashboard

Import the included Grafana dashboard from `monitoring/grafana-dashboard.json` or create custom dashboards for:

- Request rate and latency
- Provider health and availability
- Error rates by endpoint
- Token usage and costs
- Debate success rates

### Alerting Rules

```yaml
# monitoring/alerts.yml
groups:
  - name: helixagent-alerts
    rules:
      - alert: HighErrorRate
        expr: rate(helixagent_requests_total{status=~"5.."}[5m]) / rate(helixagent_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: High error rate detected
          description: Error rate is above 5% for 5 minutes

      - alert: HighLatency
        expr: histogram_quantile(0.99, rate(helixagent_request_duration_seconds_bucket[5m])) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: High latency detected
          description: P99 latency is above 10 seconds

      - alert: ProviderUnhealthy
        expr: helixagent_provider_health == 0
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: LLM provider unhealthy
          description: Provider {{ $labels.provider }} is unhealthy
```

---

## Security Hardening

### Network Policies

```yaml
# k8s/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: helixagent-network-policy
  namespace: helixagent
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
              name: ingress-nginx
      ports:
        - protocol: TCP
          port: 8080
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: postgres
      ports:
        - protocol: TCP
          port: 5432
    - to:
        - podSelector:
            matchLabels:
              app: redis
      ports:
        - protocol: TCP
          port: 6379
    - to: # Allow external LLM API access
        - ipBlock:
            cidr: 0.0.0.0/0
            except:
              - 10.0.0.0/8
              - 172.16.0.0/12
              - 192.168.0.0/16
      ports:
        - protocol: TCP
          port: 443
```

### Pod Security Standards

```yaml
# k8s/pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: helixagent-psp
spec:
  privileged: false
  runAsUser:
    rule: MustRunAsNonRoot
  seLinux:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'secret'
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
```

---

## Backup and Disaster Recovery

### Database Backup Script

```bash
#!/bin/bash
# scripts/backup-database.sh

BACKUP_DIR="/backups/postgres"
DATE=$(date +%Y%m%d-%H%M%S)
BACKUP_FILE="$BACKUP_DIR/helixagent-$DATE.sql.gz"

# Create backup
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME | gzip > $BACKUP_FILE

# Upload to S3
aws s3 cp $BACKUP_FILE s3://helixagent-backups/postgres/

# Cleanup old backups (keep 30 days)
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete
```

### Kubernetes CronJob for Backups

```yaml
# k8s/backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: helixagent
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: backup
              image: postgres:15-alpine
              command:
                - /bin/sh
                - -c
                - |
                  pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME | gzip > /backup/backup-$(date +%Y%m%d).sql.gz
              envFrom:
                - secretRef:
                    name: helixagent-secrets
              volumeMounts:
                - name: backup-volume
                  mountPath: /backup
          volumes:
            - name: backup-volume
              persistentVolumeClaim:
                claimName: backup-pvc
          restartPolicy: OnFailure
```

---

## Performance Tuning

### Application Configuration

```yaml
# configs/production.yaml
performance:
  # Connection pooling
  database:
    max_open_connections: 100
    max_idle_connections: 20
    connection_max_lifetime: "5m"

  # Redis connection pool
  redis:
    pool_size: 50
    min_idle_connections: 10

  # HTTP server
  server:
    read_timeout: "30s"
    write_timeout: "60s"
    idle_timeout: "120s"
    max_header_bytes: 1048576

  # Request processing
  workers:
    pool_size: 100
    queue_size: 1000
```

### Resource Recommendations

| Component | CPU Request | CPU Limit | Memory Request | Memory Limit |
|-----------|-------------|-----------|----------------|--------------|
| HelixAgent | 500m | 2000m | 1Gi | 4Gi |
| PostgreSQL | 500m | 2000m | 1Gi | 4Gi |
| Redis | 250m | 1000m | 512Mi | 2Gi |
| Cognee | 500m | 2000m | 1Gi | 4Gi |

---

## Troubleshooting Deployments

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Pods not starting | Resource limits | Increase resource requests/limits |
| Health check failures | Slow startup | Increase initialDelaySeconds |
| Connection errors | Network policies | Verify network policy rules |
| Database connection pool exhausted | Too many connections | Increase pool size |
| High latency | Provider issues | Check provider health |

### Diagnostic Commands

```bash
# Check pod status
kubectl -n helixagent get pods -o wide

# Describe pod for events
kubectl -n helixagent describe pod <pod-name>

# View container logs
kubectl -n helixagent logs <pod-name> -c helixagent

# Execute shell in container
kubectl -n helixagent exec -it <pod-name> -- /bin/sh

# Check resource usage
kubectl -n helixagent top pods

# View events
kubectl -n helixagent get events --sort-by='.lastTimestamp'
```

---

## Summary

This deployment guide covered:

1. **Docker Deployment**: Production-ready Docker Compose setup
2. **Kubernetes Deployment**: Complete K8s manifests for enterprise deployment
3. **Cloud Deployments**: AWS, GCP, and Azure deployment instructions
4. **High Availability**: Multi-region and database replication
5. **Monitoring**: Prometheus, Grafana, and alerting setup
6. **Security**: Network policies and pod security
7. **Backup**: Database backup and disaster recovery
8. **Performance**: Tuning recommendations

For administration tasks, see the [Administration Guide](06-administration-guide.md).
