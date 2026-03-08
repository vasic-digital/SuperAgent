# User Manual 30: Enterprise Architecture

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Enterprise Architecture Diagram](#enterprise-architecture-diagram)
4. [High Availability](#high-availability)
5. [Horizontal Scaling](#horizontal-scaling)
6. [Security Hardening](#security-hardening)
7. [Network Architecture](#network-architecture)
8. [Monitoring and Observability](#monitoring-and-observability)
9. [SLA Management](#sla-management)
10. [Cost Optimization](#cost-optimization)
11. [Capacity Planning](#capacity-planning)
12. [Deployment Pipeline](#deployment-pipeline)
13. [Infrastructure as Code](#infrastructure-as-code)
14. [Compliance and Governance](#compliance-and-governance)
15. [Troubleshooting](#troubleshooting)
16. [Related Resources](#related-resources)

## Overview

This manual describes the enterprise-grade deployment architecture for HelixAgent. It covers high availability, horizontal scaling, security hardening, network architecture, SLA management, and cost optimization for organizations running HelixAgent at scale. The architecture supports thousands of concurrent users, 99.99% uptime targets, and compliance with SOC 2, GDPR, HIPAA, and ISO 27001.

## Prerequisites

- Kubernetes 1.28+ cluster (multi-AZ)
- PostgreSQL 15 with high availability (Patroni, RDS Multi-AZ, or CloudNativePG)
- Redis 7 with Sentinel or Cluster mode
- Container registry (ECR, GCR, ACR, or self-hosted)
- TLS certificates and certificate management (cert-manager)
- Infrastructure as Code tooling (Terraform, Pulumi, or CloudFormation)
- CI/CD pipeline configured

## Enterprise Architecture Diagram

```
                         Internet
                            |
                    +-------v--------+
                    |   WAF / CDN    |
                    | (CloudFlare /  |
                    |  AWS Shield)   |
                    +-------+--------+
                            |
                    +-------v--------+
                    | Global LB      |
                    | (DNS-based     |
                    |  failover)     |
                    +---+--------+---+
                        |        |
            +-----------+        +----------+
            |                               |
   +--------v--------+           +--------v--------+
   |  AZ-1 (Primary) |           |  AZ-2 (Standby) |
   +------------------+           +------------------+
   |                  |           |                  |
   | +------+------+  |           | +------+------+  |
   | |  Ingress    |  |           | |  Ingress    |  |
   | |  Controller |  |           | |  Controller |  |
   | +------+------+  |           | +------+------+  |
   |        |         |           |        |         |
   | +------v------+  |           | +------v------+  |
   | | HelixAgent  |  |           | | HelixAgent  |  |
   | | x3-20 pods  |  |           | | x3-20 pods  |  |
   | | (HPA)       |  |           | | (HPA)       |  |
   | +------+------+  |           | +------+------+  |
   |        |         |           |        |         |
   | +------v------+  |           | +------v------+  |
   | | PostgreSQL  |  |           | | PostgreSQL  |  |
   | | (Primary)   +--+---------->+ | (Replica)   |  |
   | +-------------+  |           | +-------------+  |
   |                  |           |                  |
   | +------+------+  |           | +------+------+  |
   | | Redis       |  |           | | Redis       |  |
   | | (Primary)   +--+---------->+ | (Replica)   |  |
   | +-------------+  |           | +-------------+  |
   +------------------+           +------------------+
```

## High Availability

### Multi-AZ Deployment

Deploy HelixAgent pods across multiple availability zones to survive AZ failures:

```yaml
# k8s/deployment.yaml
spec:
  replicas: 6
  template:
    spec:
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: topology.kubernetes.io/zone
          whenUnsatisfiable: DoNotSchedule
          labelSelector:
            matchLabels:
              app: helixagent
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app: helixagent
                topologyKey: kubernetes.io/hostname
```

### Pod Disruption Budget

Ensure minimum availability during voluntary disruptions (upgrades, node drains):

```yaml
# k8s/pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: helixagent-pdb
  namespace: helixagent
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: helixagent
```

### Health Check Configuration

```yaml
livenessProbe:
  httpGet:
    path: /v1/monitoring/status
    port: 7061
  initialDelaySeconds: 120  # HelixAgent startup takes ~2 min
  periodSeconds: 30
  timeoutSeconds: 10
  failureThreshold: 3
readinessProbe:
  httpGet:
    path: /v1/monitoring/status
    port: 7061
  initialDelaySeconds: 60
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 2
startupProbe:
  httpGet:
    path: /v1/monitoring/status
    port: 7061
  initialDelaySeconds: 30
  periodSeconds: 10
  failureThreshold: 15  # Allow up to 150s for startup verification
```

### Database High Availability

Use CloudNativePG or Patroni for PostgreSQL HA:

```yaml
# CloudNativePG cluster
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: helixagent-db
  namespace: helixagent
spec:
  instances: 3
  primaryUpdateStrategy: unsupervised
  storage:
    size: 100Gi
    storageClass: gp3-encrypted
  postgresql:
    parameters:
      max_connections: "200"
      shared_buffers: "2GB"
      effective_cache_size: "6GB"
  backup:
    barmanObjectStore:
      destinationPath: s3://helixagent-db-backups/
      s3Credentials:
        accessKeyId:
          name: db-backup-creds
          key: ACCESS_KEY_ID
        secretAccessKey:
          name: db-backup-creds
          key: SECRET_ACCESS_KEY
```

## Horizontal Scaling

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
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
        - type: Pods
          value: 2
          periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Pods
          value: 1
          periodSeconds: 120
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
    - type: Pods
      pods:
        metric:
          name: helixagent_active_requests
        target:
          type: AverageValue
          averageValue: "50"
```

### Vertical Pod Autoscaler (Recommendations)

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: helixagent-vpa
  namespace: helixagent
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: helixagent
  updatePolicy:
    updateMode: "Off"  # Recommendation only, do not auto-apply
  resourcePolicy:
    containerPolicies:
      - containerName: helixagent
        minAllowed:
          cpu: "250m"
          memory: "256Mi"
        maxAllowed:
          cpu: "4000m"
          memory: "8Gi"
```

### Connection Pooling

For high-scale deployments, use PgBouncer between HelixAgent and PostgreSQL:

```yaml
# PgBouncer configuration
[databases]
helixagent_db = host=helixagent-db port=5432 dbname=helixagent_db

[pgbouncer]
listen_addr = 0.0.0.0
listen_port = 6432
auth_type = scram-sha-256
pool_mode = transaction
default_pool_size = 50
max_client_conn = 500
reserve_pool_size = 10
```

## Security Hardening

### Web Application Firewall (WAF)

```yaml
# AWS WAF rules for HelixAgent
rules:
  - name: rate-limit-global
    priority: 1
    action: block
    statement:
      rateBasedStatement:
        limit: 2000
        aggregateKeyType: IP

  - name: block-bad-bots
    priority: 2
    action: block
    statement:
      byteMatchStatement:
        fieldToMatch:
          singleHeader:
            name: user-agent
        searchString: "BadBot"

  - name: sql-injection
    priority: 3
    action: block
    statement:
      sqliMatchStatement:
        fieldToMatch:
          body: {}
        textTransformations:
          - priority: 0
            type: URL_DECODE

  - name: request-size-limit
    priority: 4
    action: block
    statement:
      sizeConstraintStatement:
        fieldToMatch:
          body: {}
        comparisonOperator: GT
        size: 10485760  # 10 MB
```

### Network Policies

Restrict pod-to-pod communication:

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
        - port: 7061
    - from:
        - namespaceSelector:
            matchLabels:
              name: monitoring
      ports:
        - port: 9090  # Prometheus scrape
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
    - to: []  # Allow outbound to LLM providers (external)
      ports:
        - port: 443
```

### Secret Management

Use Kubernetes Secrets with external secret operators:

```yaml
# External Secrets (AWS Secrets Manager)
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: helixagent-secrets
  namespace: helixagent
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: helixagent-secrets
  data:
    - secretKey: JWT_SECRET
      remoteRef:
        key: helixagent/jwt-secret
    - secretKey: DB_PASSWORD
      remoteRef:
        key: helixagent/db-password
    - secretKey: DEEPSEEK_API_KEY
      remoteRef:
        key: helixagent/deepseek-api-key
```

### TLS Configuration

```yaml
# cert-manager Certificate
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: helixagent-tls
  namespace: helixagent
spec:
  secretName: helixagent-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
    - helixagent.io
    - "*.helixagent.io"
```

## Network Architecture

### VPN Access for Administration

```
Admin Workstation -> VPN Gateway -> Private Subnet -> Bastion Host -> K8s API Server
                                                                   -> Database
                                                                   -> Redis
```

### DDoS Protection

- CloudFlare or AWS Shield Advanced for L3/L4 protection
- WAF rules for L7 protection
- Rate limiting at CDN edge and application level
- Challenge pages for suspicious traffic

## Monitoring and Observability

### Enterprise Monitoring Stack

| Component | Tool | Purpose |
|---|---|---|
| Metrics | Prometheus + Thanos | Long-term metric storage, global queries |
| Dashboards | Grafana | Visualization and alerting |
| Tracing | Jaeger + OpenTelemetry | Distributed trace collection |
| Logging | ELK Stack / Loki | Centralized log aggregation |
| LLM Observability | Langfuse | Prompt tracking, cost analysis |
| Uptime Monitoring | Datadog / PagerDuty | External availability checks |

### Key Dashboards

1. **Executive Overview** -- Request volume, error rate, SLA compliance
2. **Provider Performance** -- Latency per provider, circuit breaker states, fallback frequency
3. **Debate Analytics** -- Debate duration, consensus rate, topology effectiveness
4. **Infrastructure** -- CPU, memory, disk, network across all pods
5. **Cost Tracking** -- Token usage per provider, estimated API costs

### Alert Tiers

| Tier | Response | Examples |
|---|---|---|
| P1 (Critical) | Immediate page | Service down, all providers failed, data loss |
| P2 (High) | 15 min response | Degraded performance, >1% error rate |
| P3 (Medium) | 1 hour response | Single provider down, high latency |
| P4 (Low) | Next business day | Disk usage warning, certificate expiring |

## SLA Management

### SLA Tiers

| Tier | Uptime | Latency (P95) | Support | Price |
|---|---|---|---|---|
| Standard | 99.9% | < 2s | Business hours | Base |
| Professional | 99.95% | < 1s | 12x5 | 2x Base |
| Enterprise | 99.99% | < 500ms | 24x7 + dedicated | Custom |

### SLA Tracking

```promql
# Monthly uptime calculation
1 - (
    sum(increase(helixagent_requests_total{status=~"5.."}[30d]))
    /
    sum(increase(helixagent_requests_total[30d]))
)

# P95 latency over 30 days
histogram_quantile(0.95, rate(helixagent_request_duration_seconds_bucket[30d]))

# Error budget remaining
(1 - 0.9999) * 30 * 24 * 60  # minutes of allowed downtime per month = 4.32 min
```

### SLA Reporting

```bash
# Generate monthly SLA report
curl "http://localhost:7061/v1/admin/sla-report?period=2026-02" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq .
```

## Cost Optimization

### Compute Optimization

| Strategy | Savings | Risk | Implementation |
|---|---|---|---|
| Reserved instances | 30-60% | Commitment lock-in | 1-3 year reservations for baseline |
| Spot/Preemptible instances | 60-90% | Instance termination | Non-critical workers, with fallback |
| Right-sizing | 10-30% | Under-provisioning | VPA recommendations + monitoring |
| Auto-shutdown | 20-40% | Delayed cold start | Dev/staging environments only |

### LLM Cost Optimization

```yaml
cost_optimization:
  # Cache LLM responses to reduce API calls
  response_cache:
    enabled: true
    ttl: 1h
    backend: redis

  # Use cheaper models for non-critical requests
  model_routing:
    critical:
      model: "deepseek-chat"
      providers: ["deepseek", "openai"]
    standard:
      model: "helixagent-fast"
      providers: ["cerebras", "groq"]

  # Token budget per request
  token_limits:
    max_input_tokens: 4096
    max_output_tokens: 2048

  # Debate optimization
  debate:
    max_rounds: 3          # Limit debate rounds
    early_termination: true # Stop on consensus
    parallel_execution: true # Reduce wall-clock time
```

### Monthly Cost Estimation

| Component | Instances | Unit Cost | Monthly |
|---|---|---|---|
| HelixAgent (m5.xlarge) | 6 | $140/mo | $840 |
| PostgreSQL (db.r5.large) | 2 | $200/mo | $400 |
| Redis (cache.r5.large) | 2 | $150/mo | $300 |
| Load Balancer | 1 | $20/mo | $20 |
| LLM API costs | Variable | ~$0.002/1K tokens | ~$500-5000 |
| Monitoring stack | 3 | $50/mo | $150 |
| **Total (estimated)** | | | **$2,210-6,710** |

## Capacity Planning

### Sizing Guidelines

| User Tier | Concurrent Users | HelixAgent Pods | Database Size | Redis Memory |
|---|---|---|---|---|
| Small | 10-50 | 3 | 2 vCPU, 4 GB | 2 GB |
| Medium | 50-500 | 6-10 | 4 vCPU, 16 GB | 8 GB |
| Large | 500-5000 | 10-20 | 8 vCPU, 32 GB | 16 GB |
| Enterprise | 5000+ | 20+ (multi-region) | 16+ vCPU, 64+ GB | 32+ GB |

### Load Testing

```bash
# Run load test with resource limits
GOMAXPROCS=2 nice -n 19 ionice -c 3 \
    go test -v -p 1 -run TestStress ./tests/stress/...
```

## Deployment Pipeline

### Blue-Green Deployment

```yaml
# Blue (current) and Green (new) deployments
# Step 1: Deploy Green alongside Blue
# Step 2: Run smoke tests against Green
# Step 3: Switch traffic from Blue to Green
# Step 4: Monitor for errors
# Step 5: If errors, switch back to Blue
# Step 6: If stable, tear down Blue
```

### Canary Deployment

```yaml
# k8s/canary.yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: helixagent
spec:
  hosts:
    - helixagent.io
  http:
    - route:
        - destination:
            host: helixagent-stable
          weight: 95
        - destination:
            host: helixagent-canary
          weight: 5
```

## Infrastructure as Code

### Terraform Module

```hcl
module "helixagent" {
  source = "./modules/helixagent"

  environment      = "production"
  region           = "us-east-1"
  instance_type    = "m5.xlarge"
  min_replicas     = 3
  max_replicas     = 20
  db_instance_class = "db.r5.large"
  redis_node_type  = "cache.r5.large"

  tags = {
    Project     = "HelixAgent"
    Environment = "production"
    ManagedBy   = "terraform"
  }
}
```

## Compliance and Governance

See [User Manual 26: Compliance Guide](26-compliance-guide.md) for detailed compliance procedures. Enterprise deployments must additionally:

- Maintain audit logs for 1+ year
- Conduct quarterly access reviews
- Perform annual penetration testing
- Maintain SOC 2 Type II certification
- Implement change management processes
- Document all infrastructure changes

## Troubleshooting

### Pod OOMKilled

**Symptom:** Pods restart with OOMKilled reason.

**Solutions:**
1. Increase memory limits in the deployment manifest
2. Check for memory leaks: `curl http://localhost:7061/debug/pprof/heap`
3. Review VPA recommendations for right-sizing
4. Check if debate sessions are accumulating in memory

### Auto-Scaler Not Scaling Up

**Symptom:** High latency but HPA does not add pods.

**Solutions:**
1. Verify metrics-server is running: `kubectl top pods -n helixagent`
2. Check HPA status: `kubectl describe hpa helixagent-hpa -n helixagent`
3. Verify custom metrics adapter is deployed (for `helixagent_active_requests`)
4. Check if `maxReplicas` is already reached

### Database Connection Pool Exhaustion

**Symptom:** "too many connections" errors in HelixAgent logs.

**Solutions:**
1. Deploy PgBouncer as a connection pooler
2. Reduce `max_idle_conns` in HelixAgent database configuration
3. Increase `max_connections` in PostgreSQL (with corresponding memory increase)
4. Check for connection leaks (connections not returned to pool)

### High Latency Across All Providers

**Symptom:** All LLM providers show high latency simultaneously.

**Solutions:**
1. Check network connectivity from pods to external internet
2. Verify DNS resolution is working: `nslookup api.openai.com` from a pod
3. Check if HTTP/3 (QUIC) is being blocked by network policies or firewalls
4. Review egress proxy or NAT gateway limits
5. Check if all circuit breakers are in half-open state (indicating recovery attempts)

## Related Resources

- [User Manual 18: Performance Monitoring](18-performance-monitoring.md) -- Monitoring at scale
- [User Manual 23: Observability Setup](23-observability-setup.md) -- Enterprise observability
- [User Manual 24: Backup and Recovery](24-backup-recovery.md) -- Enterprise backup strategy
- [User Manual 25: Multi-Region Deployment](25-multi-region-deployment.md) -- Geographic distribution
- [User Manual 26: Compliance Guide](26-compliance-guide.md) -- Enterprise compliance
- [User Manual 29: Disaster Recovery](29-disaster-recovery.md) -- DR procedures
- Kubernetes documentation: https://kubernetes.io/docs/
- Terraform documentation: https://www.terraform.io/docs/
