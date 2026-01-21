# Lab 5: Production Deployment

## Overview

In this lab, you will learn how to deploy HelixAgent to production environments, including Kubernetes deployment, monitoring setup, and security hardening.

**Duration**: 2.5 hours

**Prerequisites**:
- Completed Labs 1-4
- Kubernetes cluster access (or minikube/kind)
- kubectl and helm installed
- Basic understanding of Kubernetes concepts

## Learning Objectives

By the end of this lab, you will be able to:
- Configure HelixAgent for production
- Deploy to Kubernetes using Helm
- Set up monitoring with Prometheus and Grafana
- Implement security best practices
- Configure horizontal pod autoscaling
- Manage secrets and configuration

---

## Part 1: Production Configuration (30 minutes)

### Exercise 1.1: Create Production Config

Create a production-ready configuration file:

```yaml
# configs/production.yaml

server:
  port: 8080
  host: "0.0.0.0"
  gin_mode: release
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  max_header_bytes: 1048576  # 1MB

database:
  host: "${DB_HOST}"
  port: "${DB_PORT}"
  user: "${DB_USER}"
  password: "${DB_PASSWORD}"
  database: "${DB_NAME}"
  ssl_mode: require
  max_connections: 100
  min_connections: 10
  max_conn_lifetime: 1h
  max_conn_idle_time: 30m

redis:
  host: "${REDIS_HOST}"
  port: "${REDIS_PORT}"
  password: "${REDIS_PASSWORD}"
  db: 0
  pool_size: 100
  tls_enabled: true

llm:
  default_timeout: 60s
  max_retries: 3
  retry_backoff: 1s
  circuit_breaker:
    enabled: true
    threshold: 5
    timeout: 30s

security:
  jwt_secret: "${JWT_SECRET}"
  jwt_expiry: 24h
  rate_limit:
    enabled: true
    requests_per_second: 100
    burst: 200
  cors:
    allowed_origins:
      - "https://app.helixagent.ai"
    allowed_methods:
      - GET
      - POST
      - PUT
      - DELETE
    allowed_headers:
      - Authorization
      - Content-Type

observability:
  tracing:
    enabled: true
    exporter: otlp
    endpoint: "${OTEL_EXPORTER_OTLP_ENDPOINT}"
    sample_rate: 0.1
  metrics:
    enabled: true
    port: 9090
  logging:
    level: info
    format: json
```

### Exercise 1.2: Environment Variables

Create a secure environment configuration:

```bash
# .env.production (DO NOT commit this file)

# Database
DB_HOST=postgres.production.svc.cluster.local
DB_PORT=5432
DB_USER=helixagent
DB_PASSWORD=<secure-password-from-secret-manager>
DB_NAME=helixagent

# Redis
REDIS_HOST=redis.production.svc.cluster.local
REDIS_PORT=6379
REDIS_PASSWORD=<secure-password-from-secret-manager>

# Security
JWT_SECRET=<256-bit-secret-from-secret-manager>

# LLM Providers
DEEPSEEK_API_KEY=<from-secret-manager>
GEMINI_API_KEY=<from-secret-manager>
MISTRAL_API_KEY=<from-secret-manager>

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
```

---

## Part 2: Kubernetes Deployment (45 minutes)

### Exercise 2.1: Create Helm Chart

```bash
# Create Helm chart structure
mkdir -p helm/helixagent/templates
cd helm/helixagent
```

Create `Chart.yaml`:
```yaml
# helm/helixagent/Chart.yaml
apiVersion: v2
name: helixagent
description: HelixAgent AI-powered ensemble LLM service
type: application
version: 1.0.0
appVersion: "1.0.0"
```

Create `values.yaml`:
```yaml
# helm/helixagent/values.yaml
replicaCount: 3

image:
  repository: helixagent/helixagent
  tag: "latest"
  pullPolicy: IfNotPresent

imagePullSecrets: []

serviceAccount:
  create: true
  annotations: {}
  name: ""

podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "9090"

podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000

securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
      - ALL

service:
  type: ClusterIP
  port: 8080
  metricsPort: 9090

ingress:
  enabled: true
  className: "nginx"
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: api.helixagent.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: helixagent-tls
      hosts:
        - api.helixagent.example.com

resources:
  requests:
    memory: "512Mi"
    cpu: "250m"
  limits:
    memory: "2Gi"
    cpu: "1000m"

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

nodeSelector: {}

tolerations: []

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

env:
  - name: GIN_MODE
    value: "release"
  - name: CONFIG_FILE
    value: "/etc/helixagent/production.yaml"

envFrom:
  - secretRef:
      name: helixagent-secrets

volumeMounts:
  - name: config
    mountPath: /etc/helixagent
    readOnly: true
  - name: tmp
    mountPath: /tmp

volumes:
  - name: config
    configMap:
      name: helixagent-config
  - name: tmp
    emptyDir: {}

livenessProbe:
  httpGet:
    path: /health
    port: http
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: http
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

### Exercise 2.2: Create Deployment Template

```yaml
# helm/helixagent/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "helixagent.fullname" . }}
  labels:
    {{- include "helixagent.labels" . | nindent 4 }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "helixagent.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      annotations:
        {{- toYaml .Values.podAnnotations | nindent 8 }}
      labels:
        {{- include "helixagent.selectorLabels" . | nindent 8 }}
    spec:
      serviceAccountName: {{ include "helixagent.serviceAccountName" . }}
      securityContext:
        {{- toYaml .Values.podSecurityContext | nindent 8 }}
      containers:
        - name: {{ .Chart.Name }}
          securityContext:
            {{- toYaml .Values.securityContext | nindent 12 }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
            - name: metrics
              containerPort: 9090
              protocol: TCP
          env:
            {{- toYaml .Values.env | nindent 12 }}
          envFrom:
            {{- toYaml .Values.envFrom | nindent 12 }}
          volumeMounts:
            {{- toYaml .Values.volumeMounts | nindent 12 }}
          livenessProbe:
            {{- toYaml .Values.livenessProbe | nindent 12 }}
          readinessProbe:
            {{- toYaml .Values.readinessProbe | nindent 12 }}
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
      volumes:
        {{- toYaml .Values.volumes | nindent 8 }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
```

### Exercise 2.3: Deploy to Kubernetes

```bash
# Create namespace
kubectl create namespace helixagent

# Create secrets (use your secret management solution)
kubectl create secret generic helixagent-secrets \
  --from-env-file=.env.production \
  -n helixagent

# Create ConfigMap
kubectl create configmap helixagent-config \
  --from-file=production.yaml=configs/production.yaml \
  -n helixagent

# Deploy with Helm
helm install helixagent ./helm/helixagent \
  -n helixagent \
  -f configs/production.yaml

# Verify deployment
kubectl get pods -n helixagent
kubectl get services -n helixagent
```

---

## Part 3: Monitoring Setup (45 minutes)

### Exercise 3.1: Prometheus Configuration

Create ServiceMonitor for Prometheus:

```yaml
# helm/helixagent/templates/servicemonitor.yaml
{{- if .Values.metrics.serviceMonitor.enabled }}
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: {{ include "helixagent.fullname" . }}
  labels:
    {{- include "helixagent.labels" . | nindent 4 }}
spec:
  selector:
    matchLabels:
      {{- include "helixagent.selectorLabels" . | nindent 6 }}
  endpoints:
    - port: metrics
      interval: 30s
      path: /metrics
  namespaceSelector:
    matchNames:
      - {{ .Release.Namespace }}
{{- end }}
```

### Exercise 3.2: Grafana Dashboard

Create a Grafana dashboard JSON:

```json
{
  "dashboard": {
    "title": "HelixAgent Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "gridPos": {"x": 0, "y": 0, "w": 12, "h": 8},
        "targets": [
          {
            "expr": "rate(http_requests_total{app=\"helixagent\"}[5m])",
            "legendFormat": "{{method}} {{path}}"
          }
        ]
      },
      {
        "title": "Request Latency (P99)",
        "type": "graph",
        "gridPos": {"x": 12, "y": 0, "w": 12, "h": 8},
        "targets": [
          {
            "expr": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{app=\"helixagent\"}[5m]))",
            "legendFormat": "P99"
          }
        ]
      },
      {
        "title": "LLM Provider Health",
        "type": "stat",
        "gridPos": {"x": 0, "y": 8, "w": 24, "h": 4},
        "targets": [
          {
            "expr": "llm_provider_health_status{app=\"helixagent\"}",
            "legendFormat": "{{provider}}"
          }
        ]
      },
      {
        "title": "Debate System Metrics",
        "type": "graph",
        "gridPos": {"x": 0, "y": 12, "w": 12, "h": 8},
        "targets": [
          {
            "expr": "rate(debate_requests_total{app=\"helixagent\"}[5m])",
            "legendFormat": "Debates/sec"
          },
          {
            "expr": "histogram_quantile(0.95, rate(debate_duration_seconds_bucket{app=\"helixagent\"}[5m]))",
            "legendFormat": "P95 Duration"
          }
        ]
      },
      {
        "title": "Memory & CPU",
        "type": "graph",
        "gridPos": {"x": 12, "y": 12, "w": 12, "h": 8},
        "targets": [
          {
            "expr": "process_resident_memory_bytes{app=\"helixagent\"}",
            "legendFormat": "Memory"
          },
          {
            "expr": "rate(process_cpu_seconds_total{app=\"helixagent\"}[5m])",
            "legendFormat": "CPU"
          }
        ]
      }
    ]
  }
}
```

### Exercise 3.3: Set Up Alerts

Create Prometheus alerting rules:

```yaml
# monitoring/alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: helixagent-alerts
  namespace: helixagent
spec:
  groups:
    - name: helixagent
      rules:
        - alert: HighErrorRate
          expr: |
            rate(http_requests_total{app="helixagent",status=~"5.."}[5m])
            / rate(http_requests_total{app="helixagent"}[5m]) > 0.05
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: High error rate detected
            description: Error rate is {{ $value | humanizePercentage }}

        - alert: HighLatency
          expr: |
            histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{app="helixagent"}[5m])) > 2
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: High latency detected
            description: P95 latency is {{ $value }}s

        - alert: LLMProviderDown
          expr: llm_provider_health_status{app="helixagent"} == 0
          for: 2m
          labels:
            severity: critical
          annotations:
            summary: LLM provider {{ $labels.provider }} is down
```

---

## Part 4: Security Hardening (30 minutes)

### Exercise 4.1: Network Policies

```yaml
# helm/helixagent/templates/networkpolicy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: {{ include "helixagent.fullname" . }}
spec:
  podSelector:
    matchLabels:
      {{- include "helixagent.selectorLabels" . | nindent 6 }}
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - port: 8080
    - from:
        - namespaceSelector:
            matchLabels:
              name: monitoring
      ports:
        - port: 9090
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: database
      ports:
        - port: 5432
    - to:
        - namespaceSelector:
            matchLabels:
              name: cache
      ports:
        - port: 6379
    - to:  # External LLM APIs
        - ipBlock:
            cidr: 0.0.0.0/0
      ports:
        - port: 443
```

### Exercise 4.2: Pod Security Standards

```yaml
# helm/helixagent/templates/podsecuritypolicy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: {{ include "helixagent.fullname" . }}
spec:
  privileged: false
  runAsUser:
    rule: MustRunAsNonRoot
  seLinux:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  volumes:
    - configMap
    - emptyDir
    - secret
  hostNetwork: false
  hostIPC: false
  hostPID: false
```

---

## Verification Checklist

- [ ] Production configuration is valid
- [ ] Helm chart deploys successfully
- [ ] All pods are running and healthy
- [ ] Ingress is accessible
- [ ] Metrics are being collected
- [ ] Grafana dashboard shows data
- [ ] Alerts are configured
- [ ] Network policies are applied
- [ ] HPA is working
- [ ] Secrets are properly managed

## Summary

In this lab, you learned:

1. **Production Configuration**: Creating secure, production-ready configs
2. **Kubernetes Deployment**: Using Helm for deployment
3. **Monitoring**: Setting up Prometheus and Grafana
4. **Security**: Implementing network policies and security contexts
5. **Scalability**: Configuring autoscaling

## Next Steps

- Set up CI/CD pipeline
- Configure disaster recovery
- Implement blue-green deployments
- Set up log aggregation (ELK/Loki)

## Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Helm Documentation](https://helm.sh/docs/)
- [Prometheus Operator](https://prometheus-operator.dev/)
