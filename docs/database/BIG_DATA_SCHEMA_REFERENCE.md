# Big Data Integration - SQL Schema Reference

**Version**: 1.0
**Last Updated**: 2026-01-30

---

## Table of Contents

1. [Overview](#overview)
2. [Conversation Context Schema](#conversation-context-schema)
3. [Distributed Memory Schema](#distributed-memory-schema)
4. [Cross-Session Learning Schema](#cross-session-learning-schema)
5. [ClickHouse Analytics Schema](#clickhouse-analytics-schema)
6. [Neo4j Graph Schema](#neo4j-graph-schema)
7. [Helper Functions](#helper-functions)
8. [Indexes & Constraints](#indexes--constraints)

---

## Overview

The Big Data Integration adds **25 new tables** across PostgreSQL, ClickHouse, and Neo4j:
- **PostgreSQL**: 17 tables (conversation context, distributed memory, learning)
- **ClickHouse**: 9 tables (time-series analytics)
- **Neo4j**: 3 node types, 3 relationship types

### Schema Files

| File | Purpose | Tables |
|------|---------|--------|
| `sql/schema/conversation_context.sql` | Infinite context engine | 5 |
| `sql/schema/distributed_memory.sql` | Event sourcing, CRDT | 6 |
| `sql/schema/cross_session_learning.sql` | Pattern learning | 8 |
| `sql/schema/clickhouse_analytics.sql` | Time-series metrics | 9 |
| `sql/schema/neo4j_schema.cypher` | Knowledge graph | 6 types |

---

## Conversation Context Schema

### Tables (5)

#### 1. conversation_events

**Purpose**: Store all conversation events for replay (event sourcing).

```sql
CREATE TABLE conversation_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES user_sessions(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,  -- message.added, entity.extracted, etc.
    sequence_number BIGINT NOT NULL,   -- Monotonic sequence
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    event_data JSONB NOT NULL,         -- Event payload
    metadata JSONB DEFAULT '{}'::jsonb,
    kafka_offset BIGINT,               -- Kafka offset for correlation
    kafka_partition INT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_conversation_events_conv_id ON conversation_events(conversation_id);
CREATE INDEX idx_conversation_events_type ON conversation_events(event_type);
CREATE INDEX idx_conversation_events_sequence ON conversation_events(conversation_id, sequence_number);
CREATE INDEX idx_conversation_events_timestamp ON conversation_events(timestamp);
CREATE INDEX idx_conversation_events_kafka ON conversation_events(kafka_partition, kafka_offset);
CREATE INDEX idx_conversation_events_data ON conversation_events USING gin(event_data);
```

**Event Types**:
- `message.added` - New message added to conversation
- `message.updated` - Message edited
- `message.deleted` - Message deleted
- `entity.extracted` - Entity extracted from message
- `relationship.created` - Entity relationship discovered
- `debate.started` - Debate round started
- `debate.completed` - Debate round completed
- `context.compressed` - Context compressed
- `snapshot.created` - Snapshot saved
- `conversation.paused` - Conversation paused
- `conversation.resumed` - Conversation resumed

**Example**:
```sql
INSERT INTO conversation_events (
    conversation_id, event_type, sequence_number, event_data
) VALUES (
    'conv-123',
    'message.added',
    1,
    '{"role": "user", "content": "Hello", "timestamp": "2026-01-30T10:00:00Z"}'::jsonb
);
```

#### 2. conversation_compressions

**Purpose**: Track context compression operations and metrics.

```sql
CREATE TABLE conversation_compressions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES user_sessions(id) ON DELETE CASCADE,
    compression_type VARCHAR(50) NOT NULL,  -- window_summary, entity_graph, full, hybrid
    original_message_count INT NOT NULL,
    compressed_message_count INT NOT NULL,
    compression_ratio FLOAT NOT NULL,        -- compressed / original
    original_tokens BIGINT NOT NULL,
    compressed_tokens BIGINT NOT NULL,
    quality_score FLOAT,                     -- 0-1, LLM-evaluated
    summary_content JSONB,                   -- Compressed summary
    preserved_entities JSONB,                -- Critical entities kept
    preserved_topics JSONB,                  -- Main topics kept
    compression_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    llm_model VARCHAR(100),                  -- Model used for compression
    compression_duration_ms INT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_compressions_conv_id ON conversation_compressions(conversation_id);
CREATE INDEX idx_compressions_type ON conversation_compressions(compression_type);
CREATE INDEX idx_compressions_timestamp ON conversation_compressions(compression_timestamp);
CREATE INDEX idx_compressions_ratio ON conversation_compressions(compression_ratio);
```

**Compression Types**:
- `window_summary` - Sliding window with LLM summaries
- `entity_graph` - Preserve entities + key messages
- `full` - Single comprehensive summary
- `hybrid` - Window + entity + recent messages (default)

**Example**:
```sql
INSERT INTO conversation_compressions (
    conversation_id,
    compression_type,
    original_message_count,
    compressed_message_count,
    compression_ratio,
    original_tokens,
    compressed_tokens,
    quality_score,
    llm_model
) VALUES (
    'conv-123',
    'hybrid',
    10000,
    3000,
    0.30,
    40000,
    12000,
    0.96,
    'claude-3-5-sonnet'
);
```

#### 3. conversation_snapshots

**Purpose**: Periodic snapshots of full conversation state for fast recovery.

```sql
CREATE TABLE conversation_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES user_sessions(id) ON DELETE CASCADE,
    snapshot_type VARCHAR(50) NOT NULL,  -- periodic, on_demand, pre_compression
    sequence_number BIGINT NOT NULL,     -- Event sequence at snapshot time
    message_count INT NOT NULL,
    entity_count INT NOT NULL,
    snapshot_data JSONB NOT NULL,        -- Full conversation state
    size_bytes BIGINT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_snapshots_conv_id ON conversation_snapshots(conversation_id);
CREATE INDEX idx_snapshots_sequence ON conversation_snapshots(conversation_id, sequence_number DESC);
CREATE INDEX idx_snapshots_created ON conversation_snapshots(created_at DESC);
```

#### 4. conversation_context_cache

**Purpose**: LRU cache for fast context retrieval (in-memory alternative to DB reads).

```sql
CREATE TABLE conversation_context_cache (
    conversation_id UUID PRIMARY KEY REFERENCES user_sessions(id) ON DELETE CASCADE,
    cached_context JSONB NOT NULL,
    message_count INT NOT NULL,
    entity_count INT NOT NULL,
    last_message_timestamp TIMESTAMP,
    cache_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    ttl_seconds INT DEFAULT 1800,        -- 30 minutes default
    access_count INT DEFAULT 0,
    last_accessed_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_cache_timestamp ON conversation_context_cache(cache_timestamp);
CREATE INDEX idx_cache_accessed ON conversation_context_cache(last_accessed_at);
```

#### 5. conversation_replay_stats

**Purpose**: Statistics about replay operations (monitoring and optimization).

```sql
CREATE TABLE conversation_replay_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES user_sessions(id) ON DELETE CASCADE,
    replay_type VARCHAR(50) NOT NULL,    -- full, partial, compressed
    events_replayed INT NOT NULL,
    replay_duration_ms INT NOT NULL,
    cache_hit BOOLEAN DEFAULT false,
    compression_triggered BOOLEAN DEFAULT false,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_replay_stats_conv_id ON conversation_replay_stats(conversation_id);
CREATE INDEX idx_replay_stats_duration ON conversation_replay_stats(replay_duration_ms);
CREATE INDEX idx_replay_stats_created ON conversation_replay_stats(created_at DESC);
```

---

## Distributed Memory Schema

### Tables (6)

#### 1. memory_events

**Purpose**: Event sourcing for distributed memory synchronization.

```sql
CREATE TABLE memory_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id VARCHAR(100) UNIQUE NOT NULL,   -- Globally unique event ID
    event_type VARCHAR(50) NOT NULL,         -- Created, Updated, Deleted, Merged
    node_id VARCHAR(100) NOT NULL,           -- Source node
    memory_id UUID NOT NULL,
    user_id UUID,
    session_id UUID,
    content TEXT,
    embedding VECTOR(1536),                  -- pgvector
    entities JSONB DEFAULT '[]'::jsonb,
    relationships JSONB DEFAULT '[]'::jsonb,
    version BIGINT NOT NULL,                 -- CRDT version
    timestamp TIMESTAMP NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    kafka_offset BIGINT,
    kafka_partition INT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_memory_events_event_id ON memory_events(event_id);
CREATE INDEX idx_memory_events_memory_id ON memory_events(memory_id);
CREATE INDEX idx_memory_events_node_id ON memory_events(node_id);
CREATE INDEX idx_memory_events_version ON memory_events(memory_id, version);
CREATE INDEX idx_memory_events_timestamp ON memory_events(timestamp);
```

**Event Types**:
- `Created` - New memory created
- `Updated` - Memory content/embedding updated
- `Deleted` - Memory soft-deleted
- `Merged` - Conflict resolution (CRDT merge)

#### 2. memory_conflicts

**Purpose**: Log of conflict resolutions (for auditing and debugging).

```sql
CREATE TABLE memory_conflicts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    memory_id UUID NOT NULL,
    conflict_type VARCHAR(50) NOT NULL,      -- version_conflict, concurrent_update
    node1_id VARCHAR(100) NOT NULL,
    node1_version BIGINT NOT NULL,
    node1_timestamp TIMESTAMP NOT NULL,
    node2_id VARCHAR(100) NOT NULL,
    node2_version BIGINT NOT NULL,
    node2_timestamp TIMESTAMP NOT NULL,
    resolution_strategy VARCHAR(50) NOT NULL, -- last_write_wins, merge_all, custom
    merged_version BIGINT NOT NULL,
    merged_content TEXT,
    resolution_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_conflicts_memory_id ON memory_conflicts(memory_id);
CREATE INDEX idx_conflicts_timestamp ON memory_conflicts(resolution_timestamp DESC);
```

#### 3. memory_snapshots

**Purpose**: Periodic snapshots for fast state recovery.

```sql
CREATE TABLE memory_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id VARCHAR(100) UNIQUE NOT NULL,
    node_id VARCHAR(100) NOT NULL,
    snapshot_type VARCHAR(50) NOT NULL,      -- periodic, on_demand, pre_migration
    memory_count INT NOT NULL,
    total_size_bytes BIGINT,
    snapshot_data JSONB NOT NULL,            -- Compressed memory state
    last_event_id VARCHAR(100),              -- Last event included
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_snapshots_node_id ON memory_snapshots(node_id);
CREATE INDEX idx_snapshots_created ON memory_snapshots(created_at DESC);
```

#### 4. memory_sync_status

**Purpose**: Track synchronization status across nodes.

```sql
CREATE TABLE memory_sync_status (
    node_id VARCHAR(100) PRIMARY KEY,
    last_event_consumed VARCHAR(100),
    last_event_timestamp TIMESTAMP,
    events_published BIGINT DEFAULT 0,
    events_consumed BIGINT DEFAULT 0,
    conflicts_resolved INT DEFAULT 0,
    sync_lag_ms BIGINT,                      -- Replication lag
    last_heartbeat TIMESTAMP DEFAULT NOW(),
    status VARCHAR(50) DEFAULT 'active',     -- active, lagging, failed
    metadata JSONB DEFAULT '{}'::jsonb
);
```

#### 5. crdt_versions

**Purpose**: Track CRDT version vectors for each memory.

```sql
CREATE TABLE crdt_versions (
    memory_id UUID PRIMARY KEY,
    version_vector JSONB NOT NULL,           -- {node1: 5, node2: 3, node3: 7}
    current_version BIGINT NOT NULL,
    last_updated_by VARCHAR(100) NOT NULL,
    last_updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_crdt_current_version ON crdt_versions(current_version);
```

#### 6. distributed_memory_config

**Purpose**: Configuration for distributed memory system.

```sql
CREATE TABLE distributed_memory_config (
    config_key VARCHAR(100) PRIMARY KEY,
    config_value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Default config
INSERT INTO distributed_memory_config VALUES
('crdt_strategy', '"merge_all"'::jsonb, 'CRDT conflict resolution strategy'),
('snapshot_interval_seconds', '300'::jsonb, 'Snapshot every 5 minutes'),
('sync_lag_threshold_ms', '1000'::jsonb, 'Alert if lag > 1 second'),
('enable_compression', 'true'::jsonb, 'Compress snapshot data');
```

---

## Cross-Session Learning Schema

### Tables (8)

#### 1. learned_patterns

**Purpose**: Store patterns learned from conversation analysis.

```sql
CREATE TABLE learned_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern_id VARCHAR(100) UNIQUE NOT NULL,
    pattern_type VARCHAR(50) NOT NULL,       -- user_intent, debate_strategy, etc.
    description TEXT NOT NULL,
    frequency INT DEFAULT 1,                 -- How many times observed
    confidence FLOAT DEFAULT 0.0,            -- 0-1 confidence score
    examples JSONB DEFAULT '[]'::jsonb,      -- Example instances
    metadata JSONB DEFAULT '{}'::jsonb,
    first_seen TIMESTAMP DEFAULT NOW(),
    last_seen TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_patterns_type ON learned_patterns(pattern_type);
CREATE INDEX idx_patterns_frequency ON learned_patterns(frequency DESC);
CREATE INDEX idx_patterns_confidence ON learned_patterns(confidence DESC);
CREATE INDEX idx_patterns_last_seen ON learned_patterns(last_seen DESC);
CREATE INDEX idx_patterns_metadata ON learned_patterns USING gin(metadata);
```

**Pattern Types**:
- `user_intent` - User intent patterns (help_seeking, implementation, comparison)
- `debate_strategy` - Successful debate strategies
- `entity_cooccurrence` - Entities that appear together
- `user_preference` - User communication preferences
- `conversation_flow` - Conversation timing patterns
- `provider_performance` - Provider success patterns

#### 2. learned_insights

**Purpose**: Generated insights from pattern analysis.

```sql
CREATE TABLE learned_insights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    insight_id VARCHAR(100) UNIQUE NOT NULL,
    user_id UUID,                            -- NULL for global insights
    insight_type VARCHAR(50) NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    confidence FLOAT NOT NULL,
    impact VARCHAR(20) NOT NULL,             -- high, medium, low
    patterns JSONB DEFAULT '[]'::jsonb,      -- Contributing patterns
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_insights_user_id ON learned_insights(user_id);
CREATE INDEX idx_insights_type ON learned_insights(insight_type);
CREATE INDEX idx_insights_confidence ON learned_insights(confidence DESC);
CREATE INDEX idx_insights_impact ON learned_insights(impact);
CREATE INDEX idx_insights_created ON learned_insights(created_at DESC);
CREATE INDEX idx_insights_patterns ON learned_insights USING gin(patterns);
```

#### 3. user_preferences

**Purpose**: User-specific preference patterns.

```sql
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    preference_type VARCHAR(50) NOT NULL,    -- communication_style, response_format, etc.
    preference_value TEXT NOT NULL,
    confidence FLOAT DEFAULT 0.0,
    frequency INT DEFAULT 1,
    examples JSONB DEFAULT '[]'::jsonb,
    first_observed TIMESTAMP DEFAULT NOW(),
    last_observed TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, preference_type)
);

CREATE INDEX idx_user_prefs_user_id ON user_preferences(user_id);
CREATE INDEX idx_user_prefs_confidence ON user_preferences(confidence DESC);
```

**Preference Types**:
- `communication_style` - concise, detailed, technical
- `response_format` - markdown, code-heavy, visual
- `code_language` - python, javascript, go
- `explanation_depth` - brief, moderate, comprehensive

#### 4. entity_cooccurrences

**Purpose**: Track which entities appear together.

```sql
CREATE TABLE entity_cooccurrences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity1_id VARCHAR(200) NOT NULL,
    entity1_name VARCHAR(200) NOT NULL,
    entity1_type VARCHAR(50) NOT NULL,
    entity2_id VARCHAR(200) NOT NULL,
    entity2_name VARCHAR(200) NOT NULL,
    entity2_type VARCHAR(50) NOT NULL,
    cooccurrence_count INT DEFAULT 1,
    confidence FLOAT DEFAULT 0.0,
    contexts JSONB DEFAULT '[]'::jsonb,      -- Conversation IDs
    first_seen TIMESTAMP DEFAULT NOW(),
    last_seen TIMESTAMP DEFAULT NOW(),
    UNIQUE(entity1_id, entity2_id)
);

CREATE INDEX idx_cooccur_entity1 ON entity_cooccurrences(entity1_id);
CREATE INDEX idx_cooccur_entity2 ON entity_cooccurrences(entity2_id);
CREATE INDEX idx_cooccur_count ON entity_cooccurrences(cooccurrence_count DESC);
```

#### 5. debate_strategy_success

**Purpose**: Track success rates of debate strategies.

```sql
CREATE TABLE debate_strategy_success (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    strategy_id VARCHAR(100) UNIQUE NOT NULL,
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    position VARCHAR(50) NOT NULL,           -- researcher, critic, synthesizer, etc.
    success_count INT DEFAULT 0,
    total_attempts INT DEFAULT 0,
    success_rate FLOAT DEFAULT 0.0,
    avg_confidence FLOAT DEFAULT 0.0,
    avg_response_time_ms INT DEFAULT 0,
    last_updated TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_strategy_provider ON debate_strategy_success(provider);
CREATE INDEX idx_strategy_position ON debate_strategy_success(position);
CREATE INDEX idx_strategy_success_rate ON debate_strategy_success(success_rate DESC);
```

#### 6. conversation_flow_patterns

**Purpose**: Common conversation flow patterns.

```sql
CREATE TABLE conversation_flow_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flow_type VARCHAR(50) NOT NULL,          -- rapid, normal, thoughtful
    avg_time_per_message_ms INT NOT NULL,
    message_count_min INT NOT NULL,
    message_count_max INT NOT NULL,
    success_rate FLOAT DEFAULT 0.0,
    frequency INT DEFAULT 1,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_flow_type ON conversation_flow_patterns(flow_type);
CREATE INDEX idx_flow_frequency ON conversation_flow_patterns(frequency DESC);
```

#### 7. knowledge_accumulation

**Purpose**: Accumulated knowledge from conversations.

```sql
CREATE TABLE knowledge_accumulation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_type VARCHAR(50) NOT NULL,     -- fact, procedure, concept
    subject VARCHAR(200) NOT NULL,
    fact TEXT NOT NULL,
    sources JSONB DEFAULT '[]'::jsonb,       -- Conversation IDs
    verification_count INT DEFAULT 1,
    confidence FLOAT DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_knowledge_type ON knowledge_accumulation(knowledge_type);
CREATE INDEX idx_knowledge_subject ON knowledge_accumulation(subject);
CREATE INDEX idx_knowledge_confidence ON knowledge_accumulation(confidence DESC);
CREATE INDEX idx_knowledge_fact_search ON knowledge_accumulation USING gin(to_tsvector('english', fact));
```

#### 8. learning_statistics

**Purpose**: Statistics about learning process.

```sql
CREATE TABLE learning_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stat_type VARCHAR(50) NOT NULL,
    stat_name VARCHAR(100) NOT NULL,
    stat_value NUMERIC NOT NULL,
    aggregation_period VARCHAR(20) NOT NULL, -- hourly, daily, weekly, all_time
    period_start TIMESTAMP,
    period_end TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(stat_name, aggregation_period, period_start)
);

CREATE INDEX idx_learning_stats_type ON learning_statistics(stat_type);
CREATE INDEX idx_learning_stats_period ON learning_statistics(aggregation_period, period_start DESC);
```

---

## ClickHouse Analytics Schema

### Tables (9)

#### 1. debate_metrics

```sql
CREATE TABLE debate_metrics (
    debate_id UUID,
    round UInt8,
    timestamp DateTime,
    provider String,
    model String,
    position String,
    response_time_ms Float32,
    tokens_used UInt32,
    confidence_score Float32,
    winner Boolean,
    error_count UInt8
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, debate_id, round);
```

#### 2. conversation_metrics

```sql
CREATE TABLE conversation_metrics (
    conversation_id UUID,
    user_id UUID,
    session_id UUID,
    timestamp DateTime,
    message_count UInt32,
    total_tokens UInt64,
    entity_count UInt32,
    debate_rounds UInt8,
    duration_seconds UInt32,
    compression_triggered Boolean
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, conversation_id);
```

#### 3. provider_performance

```sql
CREATE TABLE provider_performance (
    timestamp DateTime,
    provider String,
    model String,
    request_count UInt32,
    success_count UInt32,
    error_count UInt32,
    avg_response_time_ms Float32,
    p95_response_time_ms Float32,
    p99_response_time_ms Float32,
    avg_confidence Float32,
    total_tokens UInt64
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, provider, model);
```

*(Remaining 6 tables follow similar patterns for llm_response_latency, entity_extraction_metrics, memory_operations, debate_winners, system_health, api_requests)*

---

## Neo4j Graph Schema

### Node Types

```cypher
// Entity nodes
(:Entity {
    id: String,
    name: String,
    type: String,              // TECH, PERSON, ORG, CONCEPT, etc.
    importance: Float,
    mention_count: Integer,
    properties: Map,
    created_at: DateTime
})

// Conversation nodes
(:Conversation {
    id: String,
    user_id: String,
    session_id: String,
    message_count: Integer,
    started_at: DateTime,
    completed_at: DateTime
})

// User nodes
(:User {
    id: String,
    preferences: Map,
    conversation_count: Integer,
    created_at: DateTime
})
```

### Relationship Types

```cypher
// Entity relationships
(:Entity)-[:RELATED_TO {
    strength: Float,
    cooccurrence_count: Integer,
    contexts: List<String>
}]->(:Entity)

// Entity in conversation
(:Entity)-[:MENTIONED_IN {
    conversation_id: String,
    timestamp: DateTime,
    message_index: Integer
}]->(:Conversation)

// User conversations
(:User)-[:STARTED]->(:Conversation)
```

### Indexes & Constraints

```cypher
// Unique constraints
CREATE CONSTRAINT entity_id_unique IF NOT EXISTS
FOR (e:Entity) REQUIRE e.id IS UNIQUE;

CREATE CONSTRAINT conversation_id_unique IF NOT EXISTS
FOR (c:Conversation) REQUIRE c.id IS UNIQUE;

// Indexes
CREATE INDEX entity_name_idx IF NOT EXISTS
FOR (e:Entity) ON (e.name);

CREATE INDEX entity_type_idx IF NOT EXISTS
FOR (e:Entity) ON (e.type);

CREATE INDEX conversation_user_idx IF NOT EXISTS
FOR (c:Conversation) ON (c.user_id);
```

---

## Helper Functions

### PostgreSQL Helper Functions (10)

```sql
-- Get top patterns by frequency
CREATE OR REPLACE FUNCTION get_top_patterns(pattern_limit INT DEFAULT 10)
RETURNS TABLE (
    pattern_id VARCHAR,
    pattern_type VARCHAR,
    description TEXT,
    frequency INT,
    confidence FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT lp.pattern_id, lp.pattern_type, lp.description,
           lp.frequency, lp.confidence
    FROM learned_patterns lp
    ORDER BY lp.frequency DESC, lp.confidence DESC
    LIMIT pattern_limit;
END;
$$ LANGUAGE plpgsql;

-- Get user preferences summary
CREATE OR REPLACE FUNCTION get_user_preferences_summary(p_user_id UUID)
RETURNS TABLE (
    preference_type VARCHAR,
    preference_value TEXT,
    confidence FLOAT,
    frequency INT
) AS $$
BEGIN
    RETURN QUERY
    SELECT up.preference_type, up.preference_value,
           up.confidence, up.frequency
    FROM user_preferences up
    WHERE up.user_id = p_user_id
    ORDER BY up.confidence DESC;
END;
$$ LANGUAGE plpgsql;

-- Additional 8 helper functions follow similar patterns...
```

---

## Indexes & Constraints

### Primary Indexes

All tables have:
- Primary key index (UUID or composite)
- Foreign key indexes (for relationships)
- Timestamp indexes (for time-based queries)

### Performance Indexes

- **JSONB GIN Indexes**: For metadata and event_data columns
- **Partial Indexes**: For active/non-deleted records
- **Covering Indexes**: For common query patterns

### Maintenance

```sql
-- Reindex all tables (monthly)
REINDEX DATABASE helixagent_db;

-- Analyze for query optimization
ANALYZE learned_patterns;
ANALYZE conversation_events;

-- Vacuum for cleanup
VACUUM ANALYZE learned_patterns;
```

---

**Schema Version**: 1.0
**Total Tables**: 25 (17 PostgreSQL + 8 ClickHouse)
**Total Indexes**: 100+
**Total Constraints**: 20+
