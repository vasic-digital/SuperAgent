# User Manual 18: Performance Monitoring

## Overview
Guide for monitoring HelixAgent performance in production.

## Key Metrics
- Response time (p50, p95, p99)
- Throughput (requests/sec)
- Error rates
- Resource usage

## Setup
```yaml
# docker-compose.monitoring.yml
version: '3.8'
services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
  
  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
```

## Dashboards
Import dashboard ID: 12345 for HelixAgent metrics.

## Alerting
Configure alerts for:
- High latency (>500ms p95)
- Error rate (>1%)
- Memory usage (>80%)
