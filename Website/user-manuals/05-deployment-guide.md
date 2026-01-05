# SuperAgent Deployment Guide

## Introduction

This guide provides comprehensive instructions for deploying SuperAgent in production environments. It covers Docker, Kubernetes, cloud-native deployments, high availability configurations, and operational best practices for running SuperAgent at scale.

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
     │  SuperAgent-1   │      │  SuperAgent-2   │      │  SuperAgent-3   │
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
  superagent:
    image: superagent/superagent:latest
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
      - DB_NAME=superagent_prod
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
      - superagent-network

  postgres:
    image: postgres:15-alpine
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
    environment:
      - POSTGRES_DB=superagent_prod
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
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d superagent_prod"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - superagent-network

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
      - superagent-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
    depends_on:
      - superagent
    networks:
      - superagent-network

volumes:
  postgres_data:
  redis_data:

networks:
  superagent-network:
    driver: bridge
```

### Nginx Configuration

Create `nginx/nginx.conf`:

```nginx
events {
    worker_connections 4096;
}

http {
    upstream superagent {
        least_conn;
        server superagent:8080 weight=1 max_fails=3 fail_timeout=30s;
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
        server_name api.superagent.example.com;

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

            proxy_pass http://superagent;
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
            proxy_pass http://superagent/health;
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
docker-compose -f docker-compose.prod.yml logs -f superagent

# Scale replicas
docker-compose -f docker-compose.prod.yml up -d --scale superagent=5

# Rolling update
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d --no-deps superagent
```

---

## Kubernetes Deployment

### Namespace and ConfigMap

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: superagent
  labels:
    name: superagent

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: superagent-config
  namespace: superagent
data:
  PORT: "8080"
  GIN_MODE: "release"
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "superagent_prod"
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
  name: superagent-secrets
  namespace: superagent
type: Opaque
stringData:
  JWT_SECRET: "your-jwt-secret-here"
  DB_USER: "superagent"
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
  name: superagent
  namespace: superagent
  labels:
    app: superagent
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: superagent
  template:
    metadata:
      labels:
        app: superagent
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: superagent
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
        - name: superagent
          image: superagent/superagent:v1.0.0
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
              name: http
          envFrom:
            - configMapRef:
                name: superagent-config
            - secretRef:
                name: superagent-secrets
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
                        - superagent
                topologyKey: kubernetes.io/hostname
```

### Service

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: superagent-service
  namespace: superagent
  labels:
    app: superagent
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: 8080
      protocol: TCP
      name: http
  selector:
    app: superagent
```

### Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: superagent-ingress
  namespace: superagent
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  tls:
    - hosts:
        - api.superagent.example.com
      secretName: superagent-tls
  rules:
    - host: api.superagent.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: superagent-service
                port:
                  number: 80
```

### Horizontal Pod Autoscaler

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: superagent-hpa
  namespace: superagent
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: superagent
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
  namespace: superagent
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
              value: superagent_prod
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: superagent-secrets
                  key: DB_USER
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: superagent-secrets
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
  namespace: superagent
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
kubectl -n superagent get pods
kubectl -n superagent get deployments
kubectl -n superagent get services

# View logs
kubectl -n superagent logs -f deployment/superagent

# Scale manually
kubectl -n superagent scale deployment/superagent --replicas=5

# Rolling restart
kubectl -n superagent rollout restart deployment/superagent

# Check rollout status
kubectl -n superagent rollout status deployment/superagent
```

---

## Cloud Provider Deployments

### AWS (EKS)

```bash
# Create EKS cluster
eksctl create cluster \
  --name superagent-prod \
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
  --set clusterName=superagent-prod

# Deploy SuperAgent
kubectl apply -f k8s/
```

### GCP (GKE)

```bash
# Create GKE cluster
gcloud container clusters create superagent-prod \
  --zone us-central1-a \
  --num-nodes 3 \
  --machine-type e2-standard-4 \
  --enable-autoscaling \
  --min-nodes 3 \
  --max-nodes 10

# Get credentials
gcloud container clusters get-credentials superagent-prod --zone us-central1-a

# Deploy SuperAgent
kubectl apply -f k8s/
```

### Azure (AKS)

```bash
# Create AKS cluster
az aks create \
  --resource-group superagent-rg \
  --name superagent-prod \
  --node-count 3 \
  --node-vm-size Standard_D4s_v3 \
  --enable-cluster-autoscaler \
  --min-count 3 \
  --max-count 10

# Get credentials
az aks get-credentials --resource-group superagent-rg --name superagent-prod

# Deploy SuperAgent
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
      - superagent (3 replicas)
      - postgres (primary)
      - redis (cluster)

  - name: us-west-2
    role: secondary
    services:
      - superagent (3 replicas)
      - postgres (replica)
      - redis (cluster)

  - name: eu-west-1
    role: secondary
    services:
      - superagent (2 replicas)
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
  - job_name: 'superagent'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - superagent
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
  - name: superagent-alerts
    rules:
      - alert: HighErrorRate
        expr: rate(superagent_requests_total{status=~"5.."}[5m]) / rate(superagent_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: High error rate detected
          description: Error rate is above 5% for 5 minutes

      - alert: HighLatency
        expr: histogram_quantile(0.99, rate(superagent_request_duration_seconds_bucket[5m])) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: High latency detected
          description: P99 latency is above 10 seconds

      - alert: ProviderUnhealthy
        expr: superagent_provider_health == 0
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
  name: superagent-network-policy
  namespace: superagent
spec:
  podSelector:
    matchLabels:
      app: superagent
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
  name: superagent-psp
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
BACKUP_FILE="$BACKUP_DIR/superagent-$DATE.sql.gz"

# Create backup
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME | gzip > $BACKUP_FILE

# Upload to S3
aws s3 cp $BACKUP_FILE s3://superagent-backups/postgres/

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
  namespace: superagent
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
                    name: superagent-secrets
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
| SuperAgent | 500m | 2000m | 1Gi | 4Gi |
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
kubectl -n superagent get pods -o wide

# Describe pod for events
kubectl -n superagent describe pod <pod-name>

# View container logs
kubectl -n superagent logs <pod-name> -c superagent

# Execute shell in container
kubectl -n superagent exec -it <pod-name> -- /bin/sh

# Check resource usage
kubectl -n superagent top pods

# View events
kubectl -n superagent get events --sort-by='.lastTimestamp'
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
