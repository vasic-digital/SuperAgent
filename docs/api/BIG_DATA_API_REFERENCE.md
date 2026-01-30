# Big Data Integration - API Reference

**Version**: 1.0
**Base URL**: `http://localhost:7061`
**Last Updated**: 2026-01-30

---

## Table of Contents

1. [Context Replay API](#context-replay-api)
2. [Memory Sync API](#memory-sync-api)
3. [Knowledge Graph API](#knowledge-graph-api)
4. [Analytics API](#analytics-api)
5. [Data Lake API](#data-lake-api)
6. [Learning API](#learning-api)
7. [Health & Status API](#health--status-api)

---

## Context Replay API

### POST /v1/context/replay

Replay conversation from Kafka event log.

**Request**:
```json
{
  "conversation_id": "conv-123",
  "compression_strategy": "hybrid",  // optional: window, entity, full, hybrid
  "max_tokens": 4000,                // optional: trigger compression if exceeded
  "include_metadata": true           // optional: include event metadata
}
```

**Response**:
```json
{
  "conversation_id": "conv-123",
  "messages": [
    {
      "role": "user",
      "content": "Hello",
      "timestamp": "2026-01-30T10:00:00Z"
    },
    {
      "role": "assistant",
      "content": "Hi! How can I help?",
      "timestamp": "2026-01-30T10:00:02Z"
    }
  ],
  "entities": [
    {"id": "ent-1", "name": "Docker", "type": "TECH"}
  ],
  "stats": {
    "total_messages": 1000,
    "events_replayed": 2500,
    "replay_duration_ms": 150,
    "compressed": true,
    "compression_ratio": 0.30
  }
}
```

### GET /v1/context/compression/:conversation_id

Get compression metrics for a conversation.

**Response**:
```json
{
  "conversation_id": "conv-123",
  "compressions": [
    {
      "timestamp": "2026-01-30T10:15:00Z",
      "strategy": "hybrid",
      "original_tokens": 40000,
      "compressed_tokens": 12000,
      "ratio": 0.30,
      "quality_score": 0.96,
      "preserved_entities": 15,
      "llm_model": "claude-3-5-sonnet"
    }
  ]
}
```

---

## Memory Sync API

### GET /v1/memory/sync/status

Get distributed memory synchronization status.

**Response**:
```json
{
  "node_id": "node-us-east-1",
  "status": "active",
  "events_published": 1234,
  "events_consumed": 5678,
  "conflicts_resolved": 12,
  "last_sync_timestamp": "2026-01-30T10:00:00Z",
  "sync_lag_ms": 45,
  "peers": [
    {
      "node_id": "node-us-west-1",
      "status": "active",
      "last_heartbeat": "2026-01-30T09:59:55Z"
    }
  ]
}
```

### GET /v1/memory/conflicts

List recent conflict resolutions.

**Query Parameters**:
- `limit` (int, default: 50): Number of conflicts to return
- `since` (timestamp): Only conflicts after this time

**Response**:
```json
{
  "conflicts": [
    {
      "memory_id": "mem-123",
      "conflict_type": "concurrent_update",
      "node1": {"id": "node-1", "version": 5, "timestamp": "2026-01-30T10:00:00Z"},
      "node2": {"id": "node-2", "version": 5, "timestamp": "2026-01-30T10:00:01Z"},
      "resolution": {
        "strategy": "merge_all",
        "merged_version": 6,
        "timestamp": "2026-01-30T10:00:02Z"
      }
    }
  ]
}
```

---

## Knowledge Graph API

### GET /v1/knowledge/related

Get entities related to a given entity.

**Query Parameters**:
- `entity` (string, required): Entity name
- `limit` (int, default: 10): Max results
- `min_strength` (float, default: 0.5): Minimum relationship strength

**Response**:
```json
{
  "entity": "Docker",
  "related": [
    {
      "name": "Kubernetes",
      "type": "TECH",
      "strength": 0.92,
      "cooccurrence_count": 45,
      "relationship_type": "RELATED_TO"
    },
    {
      "name": "Container",
      "type": "CONCEPT",
      "strength": 0.88,
      "cooccurrence_count": 38
    }
  ]
}
```

### GET /v1/knowledge/path

Find shortest path between two entities.

**Query Parameters**:
- `from` (string, required): Source entity name
- `to` (string, required): Target entity name
- `max_depth` (int, default: 5): Maximum path length

**Response**:
```json
{
  "from": "Docker",
  "to": "Microservices",
  "path": [
    {"entity": "Docker", "type": "TECH"},
    {"relationship": "RELATED_TO", "strength": 0.85},
    {"entity": "Kubernetes", "type": "TECH"},
    {"relationship": "ENABLES", "strength": 0.78},
    {"entity": "Microservices", "type": "CONCEPT"}
  ],
  "path_length": 2,
  "total_strength": 0.815
}
```

### POST /v1/knowledge/query

Execute custom Cypher query.

**Request**:
```json
{
  "cypher": "MATCH (e:Entity {name: $name})-[:RELATED_TO]-(r) RETURN r LIMIT 10",
  "parameters": {
    "name": "Docker"
  }
}
```

**Response**:
```json
{
  "columns": ["r"],
  "rows": [
    {"r": {"id": "ent-2", "name": "Kubernetes", "type": "TECH"}},
    {"r": {"id": "ent-3", "name": "Container", "type": "CONCEPT"}}
  ],
  "row_count": 2,
  "execution_time_ms": 15
}
```

---

## Analytics API

### GET /v1/analytics/providers

Get provider performance metrics.

**Query Parameters**:
- `window` (string, default: "24h"): Time window (1h, 24h, 7d, 30d)
- `provider` (string, optional): Filter by provider

**Response**:
```json
{
  "window": "24h",
  "providers": [
    {
      "provider": "claude",
      "model": "claude-3-5-sonnet",
      "total_requests": 1234,
      "success_rate": 0.98,
      "avg_response_time_ms": 850,
      "p95_response_time_ms": 1200,
      "p99_response_time_ms": 1500,
      "avg_confidence": 0.92,
      "win_rate": 0.45,
      "total_tokens": 1250000
    }
  ]
}
```

### GET /v1/analytics/conversations/trends

Get conversation trends over time.

**Query Parameters**:
- `days` (int, default: 7): Number of days
- `granularity` (string, default: "daily"): daily, hourly

**Response**:
```json
{
  "granularity": "daily",
  "stats": [
    {
      "date": "2026-01-24",
      "conversation_count": 245,
      "message_count": 2940,
      "avg_messages_per_conv": 12,
      "avg_duration_minutes": 8.5,
      "entity_count": 1234,
      "compression_count": 15
    }
  ]
}
```

### GET /v1/analytics/debates/winners

Get debate winner distribution.

**Query Parameters**:
- `position` (string, optional): Filter by position (researcher, critic, etc.)
- `limit` (int, default: 10): Max results

**Response**:
```json
{
  "top_winners": [
    {
      "provider": "claude",
      "model": "claude-3-5-sonnet",
      "position": "researcher",
      "wins": 156,
      "percentage": 45.2,
      "avg_confidence": 0.91
    },
    {
      "provider": "deepseek",
      "model": "deepseek-chat",
      "position": "critic",
      "wins": 132,
      "percentage": 38.3,
      "avg_confidence": 0.87
    }
  ]
}
```

### GET /v1/analytics/compression/stats

Get compression quality statistics.

**Response**:
```json
{
  "total_compressions": 234,
  "strategies": {
    "hybrid": {"count": 150, "avg_ratio": 0.29, "avg_quality": 0.96},
    "window": {"count": 50, "avg_ratio": 0.32, "avg_quality": 0.94},
    "entity": {"count": 20, "avg_ratio": 0.28, "avg_quality": 0.97},
    "full": {"count": 14, "avg_ratio": 0.35, "avg_quality": 0.93}
  },
  "avg_compression_time_ms": 3500,
  "p95_compression_time_ms": 8000
}
```

---

## Data Lake API

### POST /v1/datalake/archive

Archive conversations to data lake.

**Request**:
```json
{
  "conversation_id": "conv-123",  // optional: specific conversation
  "date": "2026-01-29",           // optional: all conversations from date
  "format": "parquet"             // optional: parquet (default), json
}
```

**Response**:
```json
{
  "job_id": "archive-job-456",
  "status": "queued",
  "conversations_to_archive": 145,
  "estimated_size_mb": 280
}
```

### GET /v1/datalake/archive/:job_id/status

Get archive job status.

**Response**:
```json
{
  "job_id": "archive-job-456",
  "status": "completed",
  "progress": 1.0,
  "conversations_archived": 145,
  "files_created": 145,
  "total_size_mb": 278,
  "compression_ratio": 0.68,
  "started_at": "2026-01-30T10:00:00Z",
  "completed_at": "2026-01-30T10:05:23Z",
  "duration_seconds": 323
}
```

### POST /v1/spark/jobs

Submit Spark batch job.

**Request**:
```json
{
  "job_type": "EntityExtraction",  // EntityExtraction, RelationshipMining, TopicModeling, etc.
  "input_path": "s3://helixagent-datalake/conversations/year=2026/month=01/",
  "output_path": "s3://helixagent-datalake/entities/",
  "config": {
    "partitions": 24,
    "min_entity_confidence": 0.7
  }
}
```

**Response**:
```json
{
  "job_id": "spark-job-789",
  "status": "submitted",
  "spark_ui_url": "http://localhost:4040"
}
```

### GET /v1/spark/jobs/:job_id/status

Get Spark job status.

**Response**:
```json
{
  "job_id": "spark-job-789",
  "status": "running",
  "progress": 0.45,
  "rows_processed": 145000,
  "entities_extracted": 12345,
  "started_at": "2026-01-30T10:10:00Z",
  "elapsed_seconds": 180,
  "estimated_remaining_seconds": 220
}
```

---

## Learning API

### GET /v1/learning/patterns/top

Get top learned patterns.

**Query Parameters**:
- `limit` (int, default: 10): Max results
- `pattern_type` (string, optional): Filter by type
- `min_frequency` (int, default: 3): Minimum frequency

**Response**:
```json
{
  "patterns": [
    {
      "pattern_id": "pattern-123",
      "pattern_type": "entity_cooccurrence",
      "description": "Docker and Kubernetes often discussed together",
      "frequency": 45,
      "confidence": 0.92,
      "first_seen": "2026-01-15T10:00:00Z",
      "last_seen": "2026-01-30T09:45:00Z",
      "examples": [
        "User asks about Docker, then Kubernetes",
        "Debate about containerization mentions both"
      ]
    }
  ]
}
```

### GET /v1/learning/preferences

Get user preferences.

**Query Parameters**:
- `user_id` (string, required): User ID

**Response**:
```json
{
  "user_id": "user-123",
  "preferences": [
    {
      "type": "communication_style",
      "value": "concise",
      "confidence": 0.85,
      "frequency": 12,
      "first_observed": "2026-01-10T10:00:00Z",
      "last_observed": "2026-01-30T09:30:00Z"
    },
    {
      "type": "response_format",
      "value": "markdown",
      "confidence": 0.78,
      "frequency": 10
    },
    {
      "type": "code_language",
      "value": "python",
      "confidence": 0.92,
      "frequency": 15
    }
  ]
}
```

### GET /v1/learning/insights

Get generated insights.

**Query Parameters**:
- `user_id` (string, optional): Filter by user
- `min_confidence` (float, default: 0.7): Minimum confidence
- `impact` (string, optional): Filter by impact (high, medium, low)

**Response**:
```json
{
  "insights": [
    {
      "insight_id": "insight-456",
      "user_id": "user-123",
      "type": "personalization",
      "title": "User prefers OAuth2 for authentication",
      "description": "Based on 5 conversations, user consistently asks about OAuth2 when discussing authentication",
      "confidence": 0.87,
      "impact": "high",
      "patterns": [
        {"pattern_id": "pattern-123", "type": "user_intent"},
        {"pattern_id": "pattern-456", "type": "entity_cooccurrence"}
      ],
      "created_at": "2026-01-30T09:00:00Z"
    }
  ]
}
```

### GET /v1/learning/stats

Get learning system statistics.

**Response**:
```json
{
  "total_patterns": 156,
  "patterns_by_type": {
    "user_intent": 45,
    "debate_strategy": 32,
    "entity_cooccurrence": 38,
    "user_preference": 25,
    "conversation_flow": 12,
    "provider_performance": 4
  },
  "total_insights": 67,
  "insights_by_impact": {
    "high": 23,
    "medium": 32,
    "low": 12
  },
  "user_preferences_tracked": 234,
  "entity_relationships": 456,
  "learning_since": "2026-01-15T00:00:00Z",
  "days_learning": 15
}
```

---

## Health & Status API

### GET /health

Overall system health check.

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2026-01-30T10:00:00Z",
  "services": {
    "postgresql": "healthy",
    "redis": "healthy",
    "kafka": "healthy",
    "neo4j": "healthy",
    "clickhouse": "healthy",
    "minio": "healthy"
  },
  "uptime_seconds": 86400,
  "version": "1.0.0"
}
```

### GET /v1/bigdata/status

Big data integration status.

**Response**:
```json
{
  "kafka_streams": {
    "status": "running",
    "topics": 18,
    "consumer_groups": 5,
    "lag_ms": 45
  },
  "distributed_memory": {
    "status": "active",
    "nodes": 3,
    "sync_lag_ms": 45,
    "conflicts_last_hour": 2
  },
  "knowledge_graph": {
    "status": "healthy",
    "entity_count": 1234,
    "relationship_count": 3456,
    "update_rate_per_sec": 15
  },
  "analytics": {
    "status": "healthy",
    "clickhouse_query_avg_ms": 35,
    "data_points_last_hour": 125000
  },
  "data_lake": {
    "status": "healthy",
    "total_size_gb": 45.6,
    "parquet_files": 2345
  },
  "learning": {
    "status": "active",
    "patterns_extracted_today": 23,
    "insights_generated_today": 8
  }
}
```

---

## Error Responses

All endpoints may return the following error structure:

```json
{
  "error": {
    "code": "CONTEXT_REPLAY_FAILED",
    "message": "Failed to replay conversation: Kafka consumer timeout",
    "details": {
      "conversation_id": "conv-123",
      "kafka_error": "timeout after 30s"
    },
    "timestamp": "2026-01-30T10:00:00Z"
  }
}
```

**Common Error Codes**:
- `CONVERSATION_NOT_FOUND` (404)
- `CONTEXT_REPLAY_FAILED` (500)
- `COMPRESSION_FAILED` (500)
- `MEMORY_SYNC_ERROR` (500)
- `KNOWLEDGE_GRAPH_ERROR` (500)
- `ANALYTICS_QUERY_ERROR` (500)
- `DATA_LAKE_ERROR` (500)
- `SPARK_JOB_FAILED` (500)
- `INVALID_REQUEST` (400)
- `UNAUTHORIZED` (401)
- `RATE_LIMIT_EXCEEDED` (429)

---

## Rate Limiting

All endpoints are rate-limited:
- **Default**: 100 requests/minute per IP
- **Analytics**: 300 requests/minute (higher for dashboards)
- **Spark Jobs**: 10 submissions/hour

**Rate Limit Headers**:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1738246800
```

---

## Authentication

**API Key** (recommended for production):
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  http://localhost:7061/v1/analytics/providers
```

**JWT Token** (for user-specific requests):
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:7061/v1/learning/preferences?user_id=user-123
```

---

## Pagination

Endpoints that return lists support pagination:

**Query Parameters**:
- `page` (int, default: 1): Page number
- `limit` (int, default: 50, max: 100): Results per page

**Response Headers**:
```
X-Total-Count: 1234
X-Page: 1
X-Per-Page: 50
X-Total-Pages: 25
Link: <http://localhost:7061/v1/learning/patterns/top?page=2&limit=50>; rel="next"
```

---

## WebSocket API (Real-Time Updates)

### WS /v1/ws/knowledge/updates

Stream real-time knowledge graph updates.

**Connect**:
```javascript
const ws = new WebSocket('ws://localhost:7061/v1/ws/knowledge/updates');

ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  console.log('Entity created:', update);
};
```

**Message Format**:
```json
{
  "type": "entity_created",
  "entity": {
    "id": "ent-123",
    "name": "Docker",
    "type": "TECH"
  },
  "timestamp": "2026-01-30T10:00:00Z"
}
```

---

## Examples

### Example 1: Full Workflow

```bash
# 1. Start conversation
curl -X POST http://localhost:7061/v1/debates \
  -d '{"topic": "Explain Docker", "conversation_id": "conv-abc"}'

# 2. Continue conversation (unlimited history!)
for i in {1..1000}; do
  curl -X POST http://localhost:7061/v1/debates \
    -d "{\"topic\": \"Tell me more\", \"conversation_id\": \"conv-abc\"}"
done

# 3. Replay full context
curl -X POST http://localhost:7061/v1/context/replay \
  -d '{"conversation_id": "conv-abc"}'

# 4. Query knowledge graph
curl "http://localhost:7061/v1/knowledge/related?entity=Docker&limit=10"

# 5. View analytics
curl "http://localhost:7061/v1/analytics/providers?window=24h"

# 6. Archive to data lake
curl -X POST http://localhost:7061/v1/datalake/archive \
  -d '{"conversation_id": "conv-abc"}'

# 7. Run Spark job
curl -X POST http://localhost:7061/v1/spark/jobs \
  -d '{
    "job_type": "EntityExtraction",
    "input_path": "s3://helixagent-datalake/conversations/"
  }'

# 8. View learned patterns
curl "http://localhost:7061/v1/learning/patterns/top?limit=10"
```

---

**API Version**: 1.0
**Documentation**: https://docs.helixagent.ai/api/big-data
**Support**: support@helixagent.ai
