# Production Deployment - Complete Video Course Script

**Total Duration: 75 minutes**
**Level: Advanced**
**Prerequisites: Completion of Course 1 and 2, familiarity with Docker/Kubernetes**

---

## Module 1: Architecture Overview (15 minutes)

### Opening Slide
**Title:** Production Deployment with HelixAgent
**Duration:** 30 seconds

---

### Section 1.1: System Components (5 minutes)

#### Narration Script:

Welcome to the Production Deployment course. Before deploying HelixAgent to production, it's essential to understand its architecture and how the components work together.

HelixAgent is designed as a modern microservices-friendly application with clear separation of concerns. Let me walk you through the key components.

The core application is a Go binary that handles all API requests, provider orchestration, and debate management. It connects to PostgreSQL for persistent storage and Redis for caching and session management.

#### Key Components:

```
HELIXAGENT ARCHITECTURE

                         +-------------------+
                         |   Load Balancer   |
                         |  (nginx/HAProxy)  |
                         +---------+---------+
                                   |
              +--------------------+--------------------+
              |                    |                    |
    +---------v---------+ +--------v--------+ +--------v--------+
    |  HelixAgent Node  | |  HelixAgent Node| |  HelixAgent Node|
    |     (Go App)      | |     (Go App)    | |     (Go App)    |
    +--------+----------+ +--------+--------+ +--------+--------+
             |                     |                    |
             +----------+----------+----------+---------+
                        |                     |
              +---------v---------+  +--------v--------+
              |    PostgreSQL     |  |      Redis      |
              |  (Primary/Replica)|  |    (Cluster)    |
              +-------------------+  +-----------------+
                                            |
                                   +--------v--------+
                                   | LLM Providers   |
                                   | Claude, Gemini, |
                                   | DeepSeek, etc.  |
                                   +-----------------+
```

#### Component Details:

| Component | Purpose | Scaling Strategy |
|-----------|---------|------------------|
| HelixAgent App | API handling, orchestration | Horizontal (stateless) |
| PostgreSQL | Persistent data, sessions | Primary-replica |
| Redis | Caching, rate limiting | Cluster mode |
| Load Balancer | Traffic distribution | Active-passive |

#### Slide Content:
```
SYSTEM COMPONENTS

[Core Application]
- Go 1.23+ binary
- Gin HTTP framework
- gRPC support
- Stateless design for scaling

[Data Layer]
- PostgreSQL 15: User data, configurations, debate history
- Redis 7: Cache, sessions, rate limiting, pub/sub

[External Dependencies]
- LLM Providers (Claude, Gemini, DeepSeek, etc.)
- Cognee (optional knowledge enhancement)
- Prometheus/Grafana (monitoring)

[Key Design Principles]
- Stateless application nodes
- Externalized configuration
- Circuit breaker protection
- Graceful degradation
```

---

### Section 1.2: Data Flow (4 minutes)

#### Narration Script:

Understanding how data flows through HelixAgent helps you optimize performance and troubleshoot issues. Let me trace a typical API request from client to response.

When a request arrives, it first hits your load balancer, which routes it to an available HelixAgent node. The node authenticates the request, checks the cache for recent similar queries, and if needed, forwards the request to the appropriate LLM provider.

#### Request Flow Diagram:

```
CLIENT REQUEST FLOW

1. Client --> Load Balancer
   - SSL termination
   - Health check routing

2. Load Balancer --> HelixAgent Node
   - JWT validation
   - Rate limit check (Redis)

3. HelixAgent --> Cache Check
   - Semantic cache lookup
   - Session context retrieval

4. HelixAgent --> Provider Selection
   - Circuit breaker check
   - Load-based routing

5. HelixAgent --> LLM Provider
   - Request transformation
   - Response streaming

6. Response --> Client
   - Quality scoring
   - Response caching
   - Metrics recording
```

#### Code Example - Request Tracing:
```yaml
# Enable request tracing in production.yaml
tracing:
  enabled: true
  exporter: "jaeger"
  endpoint: "http://jaeger:14268/api/traces"
  sample_rate: 0.1  # 10% sampling in production

# Request includes trace headers
# X-Trace-ID: abc123
# X-Span-ID: def456
```

---

### Section 1.3: Scalability Patterns (3 minutes)

#### Narration Script:

HelixAgent is designed to scale horizontally. Since the application nodes are stateless, you can add more instances to handle increased load. Let me explain the key scalability patterns.

First, horizontal scaling of application nodes. You can run as many HelixAgent instances as needed behind a load balancer. They all share the same PostgreSQL and Redis backends.

Second, database scaling. PostgreSQL supports read replicas for scaling read operations. For write-heavy workloads, consider partitioning strategies.

Third, caching strategy. Redis caching significantly reduces load on LLM providers, which are typically the bottleneck.

#### Slide Content:
```
SCALABILITY PATTERNS

[Horizontal Application Scaling]
+--------+  +--------+  +--------+  +--------+
|  Node  |  |  Node  |  |  Node  |  |  Node  |
|   1    |  |   2    |  |   3    |  |   N    |
+--------+  +--------+  +--------+  +--------+
     All nodes share: PostgreSQL + Redis

[Database Scaling]
Primary --> Replica 1 --> Replica 2
  (writes)    (reads)       (reads)

[Cache Strategy]
1. Semantic cache (similar queries)
2. Response cache (exact matches)
3. Session cache (user context)

[Bottleneck Hierarchy]
LLM Provider (slowest) > Network > Database > CPU
```

---

### Section 1.4: High Availability (3 minutes)

#### Narration Script:

Production systems need high availability. HelixAgent includes several features to ensure uptime even when components fail.

The circuit breaker pattern prevents cascading failures. When a provider starts failing, HelixAgent automatically stops sending requests to it and routes traffic to healthy alternatives.

Health checks continuously monitor all components. If a HelixAgent node becomes unhealthy, the load balancer removes it from rotation.

Failover mechanisms ensure that if one provider is down, traffic automatically routes to alternatives.

#### High Availability Configuration:
```yaml
# production.yaml - HA Settings

high_availability:
  # Health check configuration
  health_check:
    enabled: true
    interval: "10s"
    timeout: "5s"
    unhealthy_threshold: 3

  # Circuit breaker for providers
  circuit_breaker:
    enabled: true
    failure_threshold: 5
    success_threshold: 2
    recovery_timeout: "30s"

  # Failover configuration
  failover:
    enabled: true
    fallback_providers:
      - primary: "claude"
        fallbacks: ["gemini", "deepseek"]
      - primary: "gemini"
        fallbacks: ["claude", "qwen"]

  # Graceful shutdown
  graceful_shutdown:
    timeout: "30s"
    drain_timeout: "15s"
```

#### Slide Content:
```
HIGH AVAILABILITY FEATURES

[Circuit Breaker States]
CLOSED --> (5 failures) --> OPEN
                             |
                        (30s timeout)
                             |
                             v
                         HALF-OPEN
                             |
            (2 successes) <--+--> (failure)
                   |                  |
                   v                  v
                CLOSED              OPEN

[Failover Chain]
Claude (primary)
    |
    +-> Gemini (fallback 1)
           |
           +-> DeepSeek (fallback 2)
                   |
                   +-> Error (all failed)

[Recovery Actions]
1. Automatic retry with backoff
2. Provider rotation
3. Graceful degradation
4. Alert notification
```

---

## Module 2: Deployment Strategies (20 minutes)

### Section 2.1: Docker Deployment (6 minutes)

#### Narration Script:

Docker is the most common way to deploy HelixAgent. Let me show you a production-ready Docker Compose configuration and explain the key settings.

#### Production Docker Compose:
```yaml
# docker-compose.prod.yml
version: "3.8"

services:
  helixagent:
    image: helixagent/helixagent:latest
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: "2"
          memory: "4G"
        reservations:
          cpus: "1"
          memory: "2G"
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
    environment:
      - GIN_MODE=release
      - PORT=8080
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - LOG_LEVEL=info
    env_file:
      - .env.production
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - helixagent-network

  postgres:
    image: postgres:15-alpine
    deploy:
      resources:
        limits:
          cpus: "2"
          memory: "4G"
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./init-scripts:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER}"]
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
          cpus: "1"
          memory: "2G"
    command: redis-server --appendonly yes --maxmemory 1gb --maxmemory-policy allkeys-lru
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
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
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
    depends_on:
      - helixagent
    networks:
      - helixagent-network

volumes:
  postgres-data:
  redis-data:

networks:
  helixagent-network:
    driver: bridge
```

#### Nginx Configuration:
```nginx
# nginx.conf
upstream helixagent {
    least_conn;
    server helixagent:8080 weight=1;
    keepalive 32;
}

server {
    listen 80;
    server_name api.yourcompany.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.yourcompany.com;

    ssl_certificate /etc/nginx/certs/fullchain.pem;
    ssl_certificate_key /etc/nginx/certs/privkey.pem;

    # Security headers
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    location / {
        proxy_pass http://helixagent;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts for long-running requests (debates)
        proxy_connect_timeout 60s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
    }

    location /health {
        proxy_pass http://helixagent/health;
        access_log off;
    }
}
```

---

### Section 2.2: Kubernetes Setup (8 minutes)

#### Narration Script:

For larger deployments, Kubernetes provides advanced orchestration capabilities. Let me show you a production-ready Kubernetes configuration.

#### Kubernetes Manifests:

**Deployment:**
```yaml
# kubernetes/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: helixagent
  labels:
    app: helixagent
spec:
  replicas: 3
  selector:
    matchLabels:
      app: helixagent
  template:
    metadata:
      labels:
        app: helixagent
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      containers:
        - name: helixagent
          image: helixagent/helixagent:v1.0.0
          ports:
            - containerPort: 8080
              name: http
            - containerPort: 9090
              name: metrics
          env:
            - name: GIN_MODE
              value: "release"
            - name: DB_HOST
              valueFrom:
                secretKeyRef:
                  name: helixagent-secrets
                  key: db-host
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: helixagent-secrets
                  key: db-password
            - name: ANTHROPIC_API_KEY
              valueFrom:
                secretKeyRef:
                  name: helixagent-secrets
                  key: anthropic-api-key
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
          volumeMounts:
            - name: config
              mountPath: /app/configs
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: helixagent-config
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

**Service:**
```yaml
# kubernetes/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: helixagent
  labels:
    app: helixagent
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: 8080
      protocol: TCP
      name: http
    - port: 9090
      targetPort: 9090
      protocol: TCP
      name: metrics
  selector:
    app: helixagent
```

**Ingress:**
```yaml
# kubernetes/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: helixagent
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/proxy-read-timeout: "300"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "300"
spec:
  tls:
    - hosts:
        - api.yourcompany.com
      secretName: helixagent-tls
  rules:
    - host: api.yourcompany.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: helixagent
                port:
                  number: 80
```

**Horizontal Pod Autoscaler:**
```yaml
# kubernetes/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: helixagent
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
```

---

### Section 2.3: Load Balancing (3 minutes)

#### Narration Script:

Proper load balancing is crucial for distributing traffic and ensuring high availability. Let me cover the key strategies and configurations.

#### Load Balancing Strategies:

```
LOAD BALANCING OPTIONS

[Round Robin]
Request 1 --> Node 1
Request 2 --> Node 2
Request 3 --> Node 3
Request 4 --> Node 1 (cycle)

Best for: Uniform request patterns

[Least Connections]
Route to node with fewest active connections

Best for: Variable request duration (debates)

[IP Hash]
Same client IP always routes to same node

Best for: Session affinity requirements

[Weighted]
Node A (weight: 3) gets 3x traffic of Node B (weight: 1)

Best for: Mixed hardware capabilities
```

#### AWS Application Load Balancer:
```yaml
# terraform/alb.tf
resource "aws_lb" "helixagent" {
  name               = "helixagent-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = var.public_subnets

  enable_deletion_protection = true
  idle_timeout               = 300  # 5 minutes for debates
}

resource "aws_lb_target_group" "helixagent" {
  name     = "helixagent-tg"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = var.vpc_id

  health_check {
    enabled             = true
    path                = "/health"
    interval            = 30
    timeout             = 10
    healthy_threshold   = 2
    unhealthy_threshold = 3
  }

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 86400  # 1 day
    enabled         = true
  }
}
```

---

### Section 2.4: Auto-Scaling (3 minutes)

#### Narration Script:

Auto-scaling ensures your deployment can handle traffic spikes while minimizing costs during quiet periods. Let me show you how to configure effective auto-scaling.

#### Auto-Scaling Configuration:

```yaml
# AWS Auto Scaling Group
resource "aws_autoscaling_group" "helixagent" {
  name                = "helixagent-asg"
  vpc_zone_identifier = var.private_subnets
  target_group_arns   = [aws_lb_target_group.helixagent.arn]
  health_check_type   = "ELB"

  min_size         = 2
  max_size         = 20
  desired_capacity = 3

  launch_template {
    id      = aws_launch_template.helixagent.id
    version = "$Latest"
  }

  tag {
    key                 = "Name"
    value               = "helixagent"
    propagate_at_launch = true
  }
}

# Scale up policy - CPU
resource "aws_autoscaling_policy" "scale_up_cpu" {
  name                   = "helixagent-scale-up-cpu"
  autoscaling_group_name = aws_autoscaling_group.helixagent.name
  adjustment_type        = "ChangeInCapacity"
  scaling_adjustment     = 2
  cooldown               = 60
}

resource "aws_cloudwatch_metric_alarm" "cpu_high" {
  alarm_name          = "helixagent-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 60
  statistic           = "Average"
  threshold           = 70

  alarm_actions = [aws_autoscaling_policy.scale_up_cpu.arn]
}

# Scale down policy
resource "aws_autoscaling_policy" "scale_down" {
  name                   = "helixagent-scale-down"
  autoscaling_group_name = aws_autoscaling_group.helixagent.name
  adjustment_type        = "ChangeInCapacity"
  scaling_adjustment     = -1
  cooldown               = 300
}
```

#### Scaling Metrics to Monitor:

```
KEY SCALING METRICS

[CPU-Based]
- Scale up: > 70% for 2 minutes
- Scale down: < 30% for 5 minutes

[Memory-Based]
- Scale up: > 80% utilization
- Scale down: < 40% utilization

[Request-Based]
- Scale up: > 1000 req/min per instance
- Scale down: < 200 req/min per instance

[Custom Metrics]
- Active debates > 50 per instance
- Provider error rate > 5%
- Response latency p95 > 5s
```

---

## Module 3: Monitoring and Observability (25 minutes)

### Section 3.1: Prometheus Integration (8 minutes)

#### Narration Script:

Prometheus is the standard for collecting metrics in cloud-native applications. HelixAgent exposes a comprehensive set of metrics that help you understand system health and performance.

#### Prometheus Configuration:
```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093

rule_files:
  - /etc/prometheus/alerts/*.yml

scrape_configs:
  - job_name: 'helixagent'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
    metrics_path: /metrics
    static_configs:
      - targets: ['helixagent:9090']

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
```

#### Key HelixAgent Metrics:

```
# Request metrics
helixagent_http_requests_total{method="POST",path="/v1/completions",status="200"}
helixagent_http_request_duration_seconds{quantile="0.95"}
helixagent_http_requests_in_flight

# Provider metrics
helixagent_provider_requests_total{provider="claude",status="success"}
helixagent_provider_latency_seconds{provider="claude",quantile="0.99"}
helixagent_provider_circuit_breaker_state{provider="claude"} # 0=closed, 1=open, 2=half-open
helixagent_provider_health_status{provider="claude"} # 1=healthy, 0=unhealthy

# Debate metrics
helixagent_debates_total{status="completed"}
helixagent_debate_duration_seconds{quantile="0.95"}
helixagent_debate_quality_score{debate_id="*"}
helixagent_debate_consensus_level{debate_id="*"}

# Cache metrics
helixagent_cache_hits_total
helixagent_cache_misses_total
helixagent_cache_hit_rate

# Resource metrics
helixagent_goroutines_count
helixagent_memory_alloc_bytes
helixagent_gc_duration_seconds
```

#### Alert Rules:
```yaml
# alerts/helixagent.yml
groups:
  - name: helixagent
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(helixagent_http_requests_total{status=~"5.."}[5m]))
          /
          sum(rate(helixagent_http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }}"

      - alert: ProviderCircuitOpen
        expr: helixagent_provider_circuit_breaker_state == 1
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Provider circuit breaker open"
          description: "Circuit breaker for {{ $labels.provider }} is open"

      - alert: HighLatency
        expr: |
          histogram_quantile(0.95, sum(rate(helixagent_http_request_duration_seconds_bucket[5m])) by (le)) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High request latency"
          description: "95th percentile latency is {{ $value }}s"

      - alert: LowCacheHitRate
        expr: helixagent_cache_hit_rate < 0.5
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "Low cache hit rate"
          description: "Cache hit rate is {{ $value | humanizePercentage }}"
```

---

### Section 3.2: Grafana Dashboards (8 minutes)

#### Narration Script:

Grafana provides beautiful visualizations for your Prometheus metrics. Let me show you how to set up comprehensive dashboards for HelixAgent monitoring.

#### Dashboard JSON (Key Panels):
```json
{
  "dashboard": {
    "title": "HelixAgent Production Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(helixagent_http_requests_total[5m]))",
            "legendFormat": "Total Requests/s"
          },
          {
            "expr": "sum(rate(helixagent_http_requests_total{status=~'2..'}[5m]))",
            "legendFormat": "Success"
          },
          {
            "expr": "sum(rate(helixagent_http_requests_total{status=~'5..'}[5m]))",
            "legendFormat": "Errors"
          }
        ]
      },
      {
        "title": "Response Time (p95)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, sum(rate(helixagent_http_request_duration_seconds_bucket[5m])) by (le))",
            "legendFormat": "p95 Latency"
          },
          {
            "expr": "histogram_quantile(0.50, sum(rate(helixagent_http_request_duration_seconds_bucket[5m])) by (le))",
            "legendFormat": "p50 Latency"
          }
        ]
      },
      {
        "title": "Provider Health",
        "type": "stat",
        "targets": [
          {
            "expr": "helixagent_provider_health_status",
            "legendFormat": "{{provider}}"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "mappings": [
              {"value": 1, "text": "Healthy", "color": "green"},
              {"value": 0, "text": "Unhealthy", "color": "red"}
            ]
          }
        }
      },
      {
        "title": "Provider Latency",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, sum(rate(helixagent_provider_latency_seconds_bucket[5m])) by (le, provider))",
            "legendFormat": "{{provider}}"
          }
        ]
      },
      {
        "title": "Active Debates",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(helixagent_debates_active)",
            "legendFormat": "Active"
          }
        ]
      },
      {
        "title": "Debate Quality Distribution",
        "type": "heatmap",
        "targets": [
          {
            "expr": "sum(rate(helixagent_debate_quality_score_bucket[1h])) by (le)"
          }
        ]
      },
      {
        "title": "Cache Performance",
        "type": "gauge",
        "targets": [
          {
            "expr": "helixagent_cache_hit_rate * 100"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "thresholds": {
              "steps": [
                {"value": 0, "color": "red"},
                {"value": 50, "color": "yellow"},
                {"value": 80, "color": "green"}
              ]
            }
          }
        }
      }
    ]
  }
}
```

#### Dashboard Sections:

```
RECOMMENDED DASHBOARD LAYOUT

[Row 1: Overview]
- Total Request Rate (graph)
- Error Rate (stat)
- Active Connections (stat)
- Uptime (stat)

[Row 2: Performance]
- Response Time p50/p95/p99 (graph)
- Request Duration Heatmap
- Requests by Endpoint (table)

[Row 3: Providers]
- Provider Health Status (stat grid)
- Provider Latency (graph)
- Provider Error Rate (graph)
- Circuit Breaker States (stat grid)

[Row 4: Debates]
- Active Debates (stat)
- Debate Duration (graph)
- Quality Score Distribution (histogram)
- Consensus Achievement Rate (stat)

[Row 5: Infrastructure]
- CPU Usage (graph)
- Memory Usage (graph)
- Database Connections (graph)
- Redis Memory (graph)
```

---

### Section 3.3: Log Management (5 minutes)

#### Narration Script:

Centralized logging is essential for troubleshooting production issues. HelixAgent produces structured logs that can be easily parsed and analyzed.

#### Log Configuration:
```yaml
# production.yaml - Logging
logging:
  format: "json"  # JSON for production
  level: "info"
  output: "stdout"

  # Fields included in all logs
  default_fields:
    service: "helixagent"
    environment: "production"
    version: "1.0.0"

  # Structured log levels
  levels:
    request: "info"
    response: "debug"
    error: "error"
    provider: "info"
    debate: "info"
```

#### Log Format Examples:
```json
// Request log
{
  "timestamp": "2024-01-15T10:30:00.000Z",
  "level": "info",
  "msg": "request_completed",
  "request_id": "req-abc123",
  "method": "POST",
  "path": "/v1/completions",
  "status": 200,
  "duration_ms": 1234,
  "provider": "claude",
  "tokens_used": 150,
  "user_id": "user-xyz"
}

// Error log
{
  "timestamp": "2024-01-15T10:30:01.000Z",
  "level": "error",
  "msg": "provider_error",
  "request_id": "req-def456",
  "provider": "gemini",
  "error": "rate_limit_exceeded",
  "retry_after": 30,
  "stack_trace": "..."
}

// Debate log
{
  "timestamp": "2024-01-15T10:30:02.000Z",
  "level": "info",
  "msg": "debate_completed",
  "debate_id": "debate-abc123",
  "duration_ms": 45000,
  "rounds": 3,
  "participants": 3,
  "consensus_reached": true,
  "quality_score": 0.87
}
```

#### ELK Stack Integration:
```yaml
# filebeat.yml
filebeat.inputs:
  - type: container
    paths:
      - /var/lib/docker/containers/*/*.log
    processors:
      - add_kubernetes_metadata:
      - decode_json_fields:
          fields: ["message"]
          target: ""
          overwrite_keys: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  indices:
    - index: "helixagent-logs-%{+yyyy.MM.dd}"
      when.contains:
        kubernetes.labels.app: "helixagent"
```

---

### Section 3.4: Alert Configuration (4 minutes)

#### Narration Script:

Effective alerting notifies you of problems before they impact users. Let me show you how to configure a comprehensive alerting system.

#### AlertManager Configuration:
```yaml
# alertmanager.yml
global:
  resolve_timeout: 5m
  slack_api_url: '${SLACK_WEBHOOK_URL}'
  pagerduty_url: 'https://events.pagerduty.com/v2/enqueue'

route:
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: 'default'

  routes:
    - match:
        severity: critical
      receiver: 'pagerduty-critical'
      continue: true

    - match:
        severity: critical
      receiver: 'slack-critical'

    - match:
        severity: warning
      receiver: 'slack-warnings'

receivers:
  - name: 'default'
    slack_configs:
      - channel: '#helixagent-alerts'

  - name: 'slack-critical'
    slack_configs:
      - channel: '#helixagent-critical'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ .Annotations.description }}'
        send_resolved: true

  - name: 'slack-warnings'
    slack_configs:
      - channel: '#helixagent-warnings'

  - name: 'pagerduty-critical'
    pagerduty_configs:
      - service_key: '${PAGERDUTY_SERVICE_KEY}'
        severity: critical
```

#### Alert Categories:

```
ALERT SEVERITY LEVELS

[Critical - Page Immediately]
- Service down (all instances)
- Database connection failures
- All providers unavailable
- Security breaches

[High - Page During Hours]
- Error rate > 10%
- Response time p95 > 10s
- Single provider down
- Database replication lag > 60s

[Warning - Slack Notification]
- Error rate > 5%
- Circuit breaker opened
- Low cache hit rate
- High memory usage

[Info - Log Only]
- Deployment events
- Configuration changes
- Routine maintenance
```

---

## Module 4: Security and Maintenance (15 minutes)

### Section 4.1: Authentication Setup (5 minutes)

#### Narration Script:

Production deployments require proper authentication. HelixAgent supports multiple authentication methods including JWT tokens, API keys, and OAuth 2.0.

#### Security Configuration:
```yaml
# production.yaml - Security
security:
  # JWT Configuration
  jwt:
    secret: "${JWT_SECRET}"  # Use strong, random secret
    expiration: "1h"
    refresh_expiration: "24h"
    algorithm: "HS256"
    issuer: "helixagent"

  # API Key Configuration
  api_key:
    enabled: true
    header: "X-API-Key"
    length: 32
    hash_algorithm: "sha256"

  # OAuth 2.0 (optional)
  oauth:
    enabled: false
    provider: "auth0"
    client_id: "${OAUTH_CLIENT_ID}"
    client_secret: "${OAUTH_CLIENT_SECRET}"
    domain: "${OAUTH_DOMAIN}"

  # Rate Limiting
  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst: 20
    by_ip: true
    by_api_key: true

  # CORS Configuration
  cors:
    allowed_origins:
      - "https://yourapp.com"
      - "https://admin.yourapp.com"
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]
    allowed_headers: ["Authorization", "Content-Type", "X-API-Key"]
    max_age: 86400

  # TLS Configuration
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/server.crt"
    key_file: "/etc/ssl/certs/server.key"
    min_version: "1.2"
```

#### API Key Management:
```bash
# Generate new API key
curl -X POST http://localhost:8080/admin/api-keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "name": "production-client",
    "permissions": ["read", "write", "debate"],
    "rate_limit": 100,
    "expires_in": "365d"
  }'

# Response
{
  "api_key": "sk-prod-abc123xyz...",
  "name": "production-client",
  "created_at": "2024-01-15T10:00:00Z",
  "expires_at": "2025-01-15T10:00:00Z"
}

# Revoke API key
curl -X DELETE http://localhost:8080/admin/api-keys/sk-prod-abc123xyz \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

---

### Section 4.2: Rate Limiting (3 minutes)

#### Narration Script:

Rate limiting protects your system from abuse and ensures fair resource distribution. HelixAgent implements multiple rate limiting strategies.

#### Rate Limiting Strategies:
```yaml
# Rate limiting configuration
rate_limiting:
  # Global limits
  global:
    requests_per_second: 1000
    burst: 200

  # Per-client limits
  per_client:
    requests_per_minute: 60
    burst: 20

  # Per-endpoint limits
  endpoints:
    "/v1/completions":
      requests_per_minute: 30
      burst: 10
    "/v1/debates":
      requests_per_minute: 10
      burst: 5

  # Rate limit headers
  headers:
    remaining: "X-RateLimit-Remaining"
    limit: "X-RateLimit-Limit"
    reset: "X-RateLimit-Reset"
```

#### Rate Limit Response:
```json
// HTTP 429 Too Many Requests
{
  "error": {
    "type": "rate_limit_error",
    "message": "Rate limit exceeded",
    "retry_after": 30
  }
}

// Headers included:
// X-RateLimit-Limit: 60
// X-RateLimit-Remaining: 0
// X-RateLimit-Reset: 1705320000
// Retry-After: 30
```

---

### Section 4.3: Backup Strategies (4 minutes)

#### Narration Script:

Regular backups are essential for disaster recovery. Let me show you how to configure automated backups for HelixAgent's data.

#### Backup Configuration:
```yaml
# backup/backup-config.yaml
backup:
  enabled: true
  schedule: "0 2 * * *"  # Daily at 2 AM

  postgresql:
    enabled: true
    retention_days: 30
    compression: true
    storage:
      type: "s3"
      bucket: "helixagent-backups"
      prefix: "postgresql/"

  redis:
    enabled: true
    retention_days: 7
    storage:
      type: "s3"
      bucket: "helixagent-backups"
      prefix: "redis/"

  configuration:
    enabled: true
    include:
      - "/app/configs/*.yaml"
      - "/app/secrets/*.env"
    storage:
      type: "s3"
      bucket: "helixagent-backups"
      prefix: "config/"
```

#### Backup Script:
```bash
#!/bin/bash
# scripts/backup.sh

set -e

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/tmp/backups/${DATE}"
S3_BUCKET="helixagent-backups"

mkdir -p ${BACKUP_DIR}

# PostgreSQL backup
echo "Backing up PostgreSQL..."
PGPASSWORD=${DB_PASSWORD} pg_dump \
  -h ${DB_HOST} \
  -U ${DB_USER} \
  -d ${DB_NAME} \
  -F c \
  -f ${BACKUP_DIR}/postgres.dump

# Redis backup
echo "Backing up Redis..."
redis-cli -h ${REDIS_HOST} BGSAVE
sleep 5
cp /var/lib/redis/dump.rdb ${BACKUP_DIR}/redis.rdb

# Compress
echo "Compressing..."
tar -czf ${BACKUP_DIR}.tar.gz -C /tmp/backups ${DATE}

# Upload to S3
echo "Uploading to S3..."
aws s3 cp ${BACKUP_DIR}.tar.gz s3://${S3_BUCKET}/backups/${DATE}.tar.gz

# Cleanup
rm -rf ${BACKUP_DIR} ${BACKUP_DIR}.tar.gz

echo "Backup completed: ${DATE}"
```

#### Kubernetes CronJob:
```yaml
# kubernetes/backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: helixagent-backup
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: backup
              image: helixagent/backup:latest
              env:
                - name: DB_HOST
                  valueFrom:
                    secretKeyRef:
                      name: helixagent-secrets
                      key: db-host
              volumeMounts:
                - name: backup-scripts
                  mountPath: /scripts
          volumes:
            - name: backup-scripts
              configMap:
                name: backup-scripts
          restartPolicy: OnFailure
```

---

### Section 4.4: Updates and Patches (3 minutes)

#### Narration Script:

Keeping HelixAgent updated ensures you have the latest features and security patches. Here's how to manage updates with minimal downtime.

#### Rolling Update Strategy:
```yaml
# kubernetes/deployment.yaml - Update Strategy
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
```

#### Update Procedure:
```bash
# 1. Check current version
kubectl get deployment helixagent -o jsonpath='{.spec.template.spec.containers[0].image}'

# 2. Review release notes
# https://github.com/helixagent/helixagent/releases

# 3. Update to new version
kubectl set image deployment/helixagent \
  helixagent=helixagent/helixagent:v1.1.0

# 4. Monitor rollout
kubectl rollout status deployment/helixagent

# 5. If issues, rollback
kubectl rollout undo deployment/helixagent
```

#### Blue-Green Deployment:
```bash
# Deploy new version as "green"
kubectl apply -f deployment-green.yaml

# Test green deployment
curl http://green.internal/health

# Switch traffic
kubectl patch service helixagent \
  -p '{"spec":{"selector":{"version":"green"}}}'

# Monitor and cleanup old version
kubectl delete deployment helixagent-blue
```

---

## Course Wrap-up (2 minutes)

#### Narration Script:

Congratulations on completing the Production Deployment course! You now have the knowledge to deploy HelixAgent in production with confidence.

We covered the system architecture and components, deployment strategies for Docker and Kubernetes, comprehensive monitoring with Prometheus and Grafana, and security best practices including authentication and backups.

In the final course, we'll cover custom integrations - building plugins, creating custom providers, and extending HelixAgent's capabilities.

#### Slide Content:
```
COURSE COMPLETE!

What You Learned:
- System architecture and data flow
- Docker and Kubernetes deployment
- Load balancing and auto-scaling
- Prometheus metrics and Grafana dashboards
- Log management and alerting
- Security configuration
- Backup and update strategies

Production Checklist:
[ ] TLS/SSL configured
[ ] Authentication enabled
[ ] Rate limiting configured
[ ] Monitoring deployed
[ ] Alerts configured
[ ] Backups scheduled
[ ] Update strategy defined

Next: Course 4 - Custom Integration
```

---

## Supplementary Materials

### Production Checklist
Complete checklist for production deployment readiness.

### Terraform Templates
Infrastructure-as-code templates for AWS and GCP.

### Grafana Dashboard JSON
Pre-built dashboards ready to import.

### Alert Rules Library
Comprehensive alert rules for common scenarios.
