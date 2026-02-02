# Video Course 15: BigData Analytics

## Course Overview

**Duration**: 3 hours
**Level**: Advanced
**Prerequisites**: Courses 01-05, familiarity with distributed systems and data pipelines

## Course Objectives

By the end of this course, you will be able to:
- Understand the BigData architecture within HelixAgent
- Configure and manage BigData components (Neo4j, ClickHouse, Kafka)
- Build and query knowledge graph streams in real time
- Implement infinite context management for long-running sessions
- Integrate data lake storage with analytics workflows
- Monitor BigData pipeline health and performance

## Module 1: Introduction to BigData in HelixAgent (25 min)

### 1.1 BigData Architecture Overview

**Video: BigData Subsystem Architecture** (15 min)
- Role of BigData in the HelixAgent ecosystem
- Component map: Neo4j, ClickHouse, Kafka, data lake storage
- How BigData integrates with ensemble orchestration and memory systems
- Graceful degradation when dependencies are unavailable
- Key file: `internal/bigdata/integration.go`

**Video: Use Cases and Design Goals** (10 min)
- Infinite context for multi-session conversations
- Distributed memory across provider clusters
- Knowledge graph streaming for entity-rich domains
- Analytics and reporting on LLM usage patterns

### Hands-On Exercise 1
Review the BigData integration entry point and trace the initialization path. Identify which components are optional and how the system behaves when each is missing.

## Module 2: Configuring BigData Components (30 min)

### 2.1 Environment Variables and Feature Flags

**Video: Configuration Deep Dive** (15 min)
- `BIGDATA_ENABLE_*` environment variables
- Per-component feature flags and their defaults
- Service override pattern: `SVC_<SERVICE>_<FIELD>`
- Development vs production configuration profiles
- Relationship to `configs/development.yaml` and `configs/production.yaml`

### 2.2 Container Setup

**Video: Deploying BigData Infrastructure** (15 min)
- Docker Compose configuration for Neo4j, ClickHouse, and Kafka
- Container runtime detection (`./scripts/container-runtime.sh`)
- Resource allocation and tuning for each component
- Network topology and port assignments
- Health check endpoints and readiness probes

### Hands-On Exercise 2
Start the BigData infrastructure stack using `make infra-start`. Verify each component is healthy using the monitoring endpoints. Experiment with disabling individual components and observe graceful degradation.

## Module 3: Knowledge Graph Streaming (30 min)

### 3.1 Graph Data Model

**Video: Knowledge Graph Fundamentals** (15 min)
- Entity and relationship modeling in Neo4j
- How HelixAgent extracts entities from LLM responses
- Graph schema design for multi-provider conversations
- Versioning and temporal annotations on graph nodes

### 3.2 Streaming Pipeline

**Video: Real-Time Graph Updates** (15 min)
- Kafka topics for entity extraction events
- Consumer pipeline: event ingestion to Neo4j writes
- Backpressure handling and batch commit strategies
- Querying the graph during active streaming sessions
- Cross-session entity linking

### Hands-On Exercise 3
Send a series of multi-turn conversation requests through the HelixAgent API. Query the Neo4j graph to verify that entities and relationships have been extracted and stored. Write a Cypher query that retrieves the full entity graph for a conversation.

## Module 4: Infinite Context Management (25 min)

### 4.1 Context Window Extension

**Video: Beyond Token Limits** (15 min)
- The problem of finite context windows across providers
- HelixAgent's approach: context segmentation and retrieval
- Interaction with the RAG subsystem (`internal/rag/`)
- Priority scoring for context segments
- Eviction policies and segment lifecycle

### 4.2 Session Continuity

**Video: Long-Running Session Support** (10 min)
- Session state persistence in PostgreSQL
- Context reconstruction on session resume
- Provider-aware context formatting
- Memory integration for cross-session continuity

### Hands-On Exercise 4
Create a long-running conversation that exceeds a single provider's context window. Verify that HelixAgent transparently manages context segmentation. Inspect the stored context segments in the database and confirm priority ordering.

## Module 5: Data Lake Integration (20 min)

### 5.1 Storage Architecture

**Video: Data Lake Design** (10 min)
- Object storage for raw LLM request and response payloads
- Partitioning strategies by provider, date, and session
- Retention policies and lifecycle management
- Compression and serialization formats

### 5.2 ETL Pipelines

**Video: Extract, Transform, Load** (10 min)
- Kafka-based event sourcing into the data lake
- Transformation stages: normalization, enrichment, deduplication
- Schema evolution and backward compatibility
- Loading data into ClickHouse for analytics

### Hands-On Exercise 5
Configure the data lake storage path and retention policy. Generate traffic through the API and verify that raw payloads are stored correctly. Run a sample ETL pipeline to load data into ClickHouse.

## Module 6: Analytics with ClickHouse (25 min)

### 6.1 Schema and Table Design

**Video: ClickHouse for LLM Analytics** (10 min)
- Table engines: MergeTree, AggregatingMergeTree
- Schema design for request latency, token usage, and cost tracking
- Materialized views for real-time aggregations
- Data types and column encoding for high cardinality fields

### 6.2 Querying and Dashboards

**Video: Building Analytics Queries** (15 min)
- Common analytical queries: provider performance, cost trends, error rates
- Time-series analysis of response latency
- Percentile calculations for SLA monitoring
- Connecting ClickHouse to Grafana for visualization
- Prometheus metrics export from ClickHouse

### Hands-On Exercise 6
Write ClickHouse queries to analyze provider response times over the past 24 hours. Create a materialized view that aggregates token usage by provider per hour. Build a Grafana dashboard panel using the ClickHouse data source.

## Module 7: Monitoring and Health Checks (25 min)

### 7.1 BigData Health Endpoint

**Video: Health and Status Monitoring** (10 min)
- `/v1/bigdata/health` endpoint structure and response format
- Per-component health status reporting
- Circuit breaker integration for BigData dependencies
- Alerting thresholds and escalation policies

### 7.2 Operational Runbooks

**Video: Troubleshooting BigData Issues** (15 min)
- Common failure scenarios: Kafka lag, Neo4j connection exhaustion, ClickHouse memory pressure
- Diagnostic commands and log analysis
- Recovery procedures for each component
- Performance baseline establishment and drift detection
- Integration with OpenTelemetry tracing (`internal/observability/`)

### Hands-On Exercise 7
Simulate a Kafka consumer lag scenario by pausing a consumer group. Observe the health endpoint response and circuit breaker state changes. Follow the runbook to restore normal operation and verify recovery through the monitoring dashboard.

## Course Summary

### Key Takeaways
- BigData in HelixAgent is modular; every component degrades gracefully when unavailable
- Knowledge graph streaming enables real-time entity extraction and cross-session linking
- Infinite context management transparently extends provider token limits
- ClickHouse provides high-performance analytics on LLM usage data
- The `/v1/bigdata/health` endpoint is the single pane of glass for BigData health

### Next Steps
- Course 16: Memory Management for deeper coverage of distributed memory and CRDT resolution
- Course 10: Security Best Practices for securing BigData pipelines
- Course 13: Enterprise Deployment for production-grade BigData infrastructure

### Additional Resources
- `internal/bigdata/integration.go` -- BigData integration entry point
- `internal/observability/` -- OpenTelemetry and tracing configuration
- `configs/production.yaml` -- Production configuration reference
- HelixAgent BigData API reference documentation
