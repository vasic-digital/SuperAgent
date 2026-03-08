# Video Course 53: HelixMemory Deep Dive

## Course Overview

**Duration**: 2 hours 15 minutes
**Level**: Advanced
**Prerequisites**: Course 01-Fundamentals, Course 16-Memory-Management, familiarity with Go interfaces and concurrency patterns

HelixMemory is the unified cognitive memory engine for HelixAgent and the AI debate ensemble. This course covers its architecture, the 3-stage fusion pipeline, backend configuration, circuit breakers, and hands-on memory operations.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Explain the HelixMemory architecture and its role in the HelixAgent ecosystem
2. Configure and operate the four memory backends (Mem0, Cognee, Letta, Graphiti)
3. Understand and trace the 3-stage fusion pipeline from ingestion to retrieval
4. Monitor HelixMemory health using Prometheus metrics and circuit breakers
5. Store, retrieve, and manage memories through the HelixMemory API
6. Disable HelixMemory via build tags when not required

---

## Module 1: Architecture Overview (25 min)

### 1.1 What Is HelixMemory?

**Video: Unified Cognitive Memory Engine** (10 min)

- HelixMemory orchestrates four distinct memory backends into a single coherent system
- Active by default in HelixAgent; opt out with `-tags nohelixmemory`
- Module path: `digital.vasic.helixmemory` located in `HelixMemory/`
- 12+ internal packages covering fusion, backends, metrics, and infrastructure bridge

**Key Concepts:**

| Backend   | Purpose                           | Data Model          |
|-----------|-----------------------------------|---------------------|
| Mem0      | Factual memory (key-value facts)  | Flat fact store     |
| Cognee    | Knowledge graph construction      | Entity-relationship |
| Letta     | Stateful agent runtime memory     | Session state       |
| Graphiti  | Temporal knowledge graph          | Time-stamped edges  |

### 1.2 Module Structure

**Video: Package Layout and Dependencies** (8 min)

- `HelixMemory/fusion/` -- 3-stage pipeline orchestration
- `HelixMemory/backends/` -- Mem0, Cognee, Letta, Graphiti adapters
- `HelixMemory/metrics/` -- Prometheus counters, histograms, gauges
- `HelixMemory/circuitbreaker/` -- Per-backend fault isolation
- `HelixMemory/bridge/` -- Infrastructure bridge to HelixAgent internals

### 1.3 Integration with AI Debate

**Video: Memory in the Debate Ensemble** (7 min)

- Debate sessions store intermediate reasoning in HelixMemory
- Cross-session learning uses Graphiti temporal edges
- Reflexion framework reads episodic memory buffer from Mem0
- Knowledge graph from Cognee informs dehallucination phase

### Hands-On Lab 1

Inspect the HelixMemory module structure and verify build tag behavior:

```bash
# Build with HelixMemory (default)
make build

# Build without HelixMemory
go build -tags nohelixmemory ./cmd/helixagent

# Compare binary sizes
ls -lh bin/helixagent
```

---

## Module 2: The 3-Stage Fusion Pipeline (30 min)

### 2.1 Stage 1 -- Ingestion

**Video: Memory Ingestion and Routing** (10 min)

- Incoming memory items are classified by type (fact, entity, session state, temporal event)
- Router dispatches to one or more backends based on classification
- Deduplication checks run before backend writes
- Write operations are parallel with per-backend timeouts

### 2.2 Stage 2 -- Consolidation

**Video: Cross-Backend Consolidation** (10 min)

- Periodic consolidation merges overlapping facts across backends
- Entity resolution links Cognee graph nodes to Mem0 facts
- Temporal ordering from Graphiti annotates Letta session states
- Conflict resolution uses recency and confidence scoring

### 2.3 Stage 3 -- Retrieval and Ranking

**Video: Unified Retrieval** (10 min)

- Query fans out to all backends in parallel
- Results are scored by relevance, recency, and source reliability
- Fusion ranker merges and deduplicates cross-backend results
- Final ranked list returned to the caller with provenance metadata

### Hands-On Lab 2

Trace a memory item through the fusion pipeline using debug logging:

```bash
# Enable debug logging for HelixMemory
export HELIXMEMORY_LOG_LEVEL=debug

# Start HelixAgent and observe fusion pipeline logs
./bin/helixagent 2>&1 | grep -E "helixmemory|fusion|backend"
```

---

## Module 3: Backend Configuration (25 min)

### 3.1 Mem0 Configuration

**Video: Configuring Mem0 for Fact Storage** (7 min)

- Connection settings via environment variables
- Memory scope configuration (user, session, global)
- TTL policies for fact expiration
- Batch operations for bulk ingestion

### 3.2 Cognee Configuration

**Video: Knowledge Graph with Cognee** (6 min)

- Cognee container setup via `docker-compose.cognee.yml`
- Entity extraction configuration
- Relationship type definitions
- Graph traversal depth limits

### 3.3 Letta and Graphiti Configuration

**Video: Stateful Runtime and Temporal Graph** (6 min)

- Letta session persistence settings
- Graphiti temporal resolution (second, minute, hour granularity)
- Edge retention policies
- Snapshot intervals for state checkpointing

### 3.4 Environment Variables Reference

**Video: Configuration Reference** (6 min)

```bash
# Mem0
MEM0_HOST=localhost
MEM0_PORT=8090

# Cognee
COGNEE_ENABLED=true
COGNEE_HOST=localhost
COGNEE_PORT=8091

# Letta
LETTA_HOST=localhost
LETTA_PORT=8092

# Graphiti
GRAPHITI_HOST=localhost
GRAPHITI_PORT=8093

# Global
HELIXMEMORY_FUSION_TIMEOUT=30s
HELIXMEMORY_CONSOLIDATION_INTERVAL=5m
```

### Hands-On Lab 3

Configure all four backends and verify connectivity:

```bash
# Start memory infrastructure
make infra-core

# Verify backend health
curl http://localhost:7061/v1/helixmemory/health
```

---

## Module 4: Circuit Breakers and Observability (20 min)

### 4.1 Per-Backend Circuit Breakers

**Video: Fault Isolation** (10 min)

- Each backend has an independent circuit breaker
- Failure threshold, success threshold, and recovery timeout are configurable
- When a backend circuit opens, fusion pipeline degrades gracefully
- Remaining backends continue serving requests without interruption

### 4.2 Prometheus Metrics

**Video: Monitoring HelixMemory** (10 min)

- `helixmemory_store_duration_seconds` -- write latency histogram per backend
- `helixmemory_retrieve_duration_seconds` -- read latency histogram per backend
- `helixmemory_circuit_breaker_state` -- gauge (0=closed, 1=half-open, 2=open)
- `helixmemory_fusion_items_total` -- counter of items processed through pipeline
- `helixmemory_consolidation_runs_total` -- counter of consolidation cycles

### Hands-On Lab 4

Create a Grafana dashboard for HelixMemory metrics:

1. Import the Prometheus data source
2. Create panels for write/read latency per backend
3. Add circuit breaker state indicators
4. Set alert thresholds for degraded backends

---

## Module 5: Hands-On Operations (25 min)

### 5.1 Storing Memories

**Video: Write Operations** (8 min)

```bash
# Store a fact via the HelixMemory API
curl -X POST http://localhost:7061/v1/helixmemory/store \
  -H "Content-Type: application/json" \
  -d '{
    "type": "fact",
    "scope": "user",
    "user_id": "user-123",
    "content": "User prefers Python for data analysis",
    "metadata": {"confidence": 0.95, "source": "conversation"}
  }'
```

### 5.2 Retrieving Memories

**Video: Read and Search Operations** (8 min)

```bash
# Retrieve memories by query
curl -X POST http://localhost:7061/v1/helixmemory/retrieve \
  -H "Content-Type: application/json" \
  -d '{
    "query": "programming language preferences",
    "scope": "user",
    "user_id": "user-123",
    "limit": 10
  }'
```

### 5.3 Managing Memory Lifecycle

**Video: Update, Delete, and Consolidation** (9 min)

- Updating existing memories with new confidence scores
- Deleting memories by ID or by scope
- Forcing manual consolidation runs
- Exporting memory snapshots for backup

### Hands-On Lab 5

Complete end-to-end memory workflow:

1. Store 5 facts about a user via the API
2. Store 3 entities in the knowledge graph
3. Query for related memories using semantic search
4. Verify fusion ranking returns cross-backend results
5. Delete a specific memory and confirm removal

---

## Course Summary

### Key Takeaways

1. HelixMemory unifies four memory backends (Mem0, Cognee, Letta, Graphiti) through a 3-stage fusion pipeline
2. The fusion pipeline handles ingestion routing, cross-backend consolidation, and ranked retrieval
3. Per-backend circuit breakers ensure graceful degradation when individual backends fail
4. Prometheus metrics provide full observability into memory operations
5. HelixMemory is active by default; disable with `-tags nohelixmemory`

### Assessment Questions

1. Name the four memory backends and describe the primary purpose of each.
2. What are the three stages of the fusion pipeline and what happens in each?
3. How does a circuit breaker opening on one backend affect the overall memory system?
4. Which Prometheus metric would you use to detect backend latency degradation?
5. How does the AI debate ensemble use HelixMemory for cross-session learning?

### Related Courses

- Course 16: Memory Management
- Course 02: AI Debate System
- Course 59: Monitoring and Observability
- Course 15: BigData Analytics

---

**Course Version**: 1.0
**Last Updated**: March 8, 2026
