-- ============================================================================
-- Conversation Context & Compression Schema
-- Infinite context engine with Kafka-backed replay and LLM compression
-- ============================================================================

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- Conversation Events Table
-- Stores conversation event pointers (actual events in Kafka)
-- ============================================================================

CREATE TABLE IF NOT EXISTS conversation_events (
    event_id VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    conversation_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255),
    user_id VARCHAR(255) NOT NULL,
    node_id VARCHAR(255) NOT NULL,
    sequence_number BIGINT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Event payload summary (full payload in Kafka)
    message_id VARCHAR(255),
    entity_count INT DEFAULT 0,
    tokens INT DEFAULT 0,

    -- Kafka metadata
    kafka_topic VARCHAR(255),
    kafka_partition INT,
    kafka_offset BIGINT,

    -- Metadata
    metadata JSONB,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for conversation events
CREATE INDEX IF NOT EXISTS idx_conversation_events_conversation_id ON conversation_events(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conversation_events_user_id ON conversation_events(user_id);
CREATE INDEX IF NOT EXISTS idx_conversation_events_session_id ON conversation_events(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_conversation_events_timestamp ON conversation_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_events_sequence ON conversation_events(conversation_id, sequence_number);
CREATE INDEX IF NOT EXISTS idx_conversation_events_type ON conversation_events(event_type);

-- Composite index for replay queries
CREATE INDEX IF NOT EXISTS idx_conversation_events_replay
    ON conversation_events(conversation_id, sequence_number ASC, timestamp ASC);

-- ============================================================================
-- Conversation Compressions Table
-- Tracks context compression operations
-- ============================================================================

CREATE TABLE IF NOT EXISTS conversation_compressions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    compression_id VARCHAR(255) UNIQUE NOT NULL,
    conversation_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255),

    -- Compression metadata
    compression_type VARCHAR(50) NOT NULL,  -- 'window_summary', 'entity_graph', 'full', 'hybrid'
    compression_strategy VARCHAR(50),

    -- Compression metrics
    original_message_count INT NOT NULL,
    compressed_message_count INT NOT NULL,
    compression_ratio FLOAT NOT NULL,
    original_tokens BIGINT NOT NULL,
    compressed_tokens BIGINT NOT NULL,

    -- Compressed content
    summary_content TEXT,
    preserved_entities TEXT[],
    key_topics TEXT[],

    -- Processing metadata
    llm_model VARCHAR(100),
    compression_duration_ms BIGINT,
    compressed_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Context snapshot
    context_snapshot JSONB,

    -- Indexes
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for compressions
CREATE INDEX IF NOT EXISTS idx_conversation_compressions_conversation ON conversation_compressions(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conversation_compressions_user ON conversation_compressions(user_id);
CREATE INDEX IF NOT EXISTS idx_conversation_compressions_session ON conversation_compressions(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_conversation_compressions_compressed_at ON conversation_compressions(compressed_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_compressions_type ON conversation_compressions(compression_type);

-- GIN indexes for arrays
CREATE INDEX IF NOT EXISTS idx_conversation_compressions_entities ON conversation_compressions USING GIN (preserved_entities);
CREATE INDEX IF NOT EXISTS idx_conversation_compressions_topics ON conversation_compressions USING GIN (key_topics);

-- ============================================================================
-- Conversation Snapshots Table
-- Periodic snapshots for faster recovery
-- ============================================================================

CREATE TABLE IF NOT EXISTS conversation_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255),
    node_id VARCHAR(255) NOT NULL,
    sequence_number BIGINT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Snapshot data
    messages JSONB NOT NULL,
    entities JSONB,
    context JSONB,

    -- Snapshot metadata
    message_count INT NOT NULL DEFAULT 0,
    entity_count INT NOT NULL DEFAULT 0,
    total_tokens BIGINT DEFAULT 0,
    compressed_count INT DEFAULT 0,
    compression_ratio FLOAT,

    -- Snapshot type
    snapshot_type VARCHAR(50) DEFAULT 'periodic',  -- 'periodic', 'compression', 'checkpoint'

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for snapshots
CREATE INDEX IF NOT EXISTS idx_conversation_snapshots_conversation ON conversation_snapshots(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conversation_snapshots_user ON conversation_snapshots(user_id);
CREATE INDEX IF NOT EXISTS idx_conversation_snapshots_timestamp ON conversation_snapshots(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_snapshots_sequence ON conversation_snapshots(conversation_id, sequence_number DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_snapshots_type ON conversation_snapshots(snapshot_type);

-- Composite index for latest snapshot
CREATE INDEX IF NOT EXISTS idx_conversation_snapshots_latest
    ON conversation_snapshots(conversation_id, timestamp DESC, sequence_number DESC);

-- ============================================================================
-- Context Cache Table
-- LRU cache for replayed conversations
-- ============================================================================

CREATE TABLE IF NOT EXISTS conversation_context_cache (
    conversation_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255),

    -- Cached data
    messages JSONB NOT NULL,
    entities JSONB,
    context JSONB,

    -- Cache metadata
    message_count INT NOT NULL DEFAULT 0,
    entity_count INT NOT NULL DEFAULT 0,
    total_tokens BIGINT DEFAULT 0,

    -- Cache stats
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_accessed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    access_count INT DEFAULT 0,
    ttl INT DEFAULT 1800,  -- TTL in seconds (30 minutes)

    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for context cache
CREATE INDEX IF NOT EXISTS idx_context_cache_user ON conversation_context_cache(user_id);
CREATE INDEX IF NOT EXISTS idx_context_cache_session ON conversation_context_cache(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_context_cache_accessed ON conversation_context_cache(last_accessed_at DESC);
CREATE INDEX IF NOT EXISTS idx_context_cache_cached ON conversation_context_cache(cached_at DESC);

-- TTL-based cleanup index
CREATE INDEX IF NOT EXISTS idx_context_cache_expired
    ON conversation_context_cache(cached_at)
    WHERE (EXTRACT(EPOCH FROM (NOW() - cached_at)) > ttl);

-- ============================================================================
-- Conversation Replay Stats Table
-- Tracks replay and compression statistics
-- ============================================================================

CREATE TABLE IF NOT EXISTS conversation_replay_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,

    -- Replay metadata
    replay_type VARCHAR(50) NOT NULL,  -- 'full', 'compressed', 'snapshot'
    events_processed INT DEFAULT 0,
    messages_replayed INT DEFAULT 0,
    entities_extracted INT DEFAULT 0,

    -- Performance metrics
    replay_duration_ms BIGINT,
    cache_hit BOOLEAN DEFAULT FALSE,

    -- Compression applied
    compression_applied BOOLEAN DEFAULT FALSE,
    compression_id VARCHAR(255),

    -- Timestamps
    replayed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for replay stats
CREATE INDEX IF NOT EXISTS idx_replay_stats_conversation ON conversation_replay_stats(conversation_id);
CREATE INDEX IF NOT EXISTS idx_replay_stats_user ON conversation_replay_stats(user_id);
CREATE INDEX IF NOT EXISTS idx_replay_stats_replayed_at ON conversation_replay_stats(replayed_at DESC);
CREATE INDEX IF NOT EXISTS idx_replay_stats_type ON conversation_replay_stats(replay_type);

-- ============================================================================
-- Functions for Context Management
-- ============================================================================

-- Function: Get latest conversation snapshot
CREATE OR REPLACE FUNCTION get_latest_snapshot(p_conversation_id VARCHAR)
RETURNS TABLE (
    snapshot_id UUID,
    sequence_number BIGINT,
    messages JSONB,
    entities JSONB,
    context JSONB,
    timestamp TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cs.snapshot_id,
        cs.sequence_number,
        cs.messages,
        cs.entities,
        cs.context,
        cs.timestamp
    FROM conversation_snapshots cs
    WHERE cs.conversation_id = p_conversation_id
    ORDER BY cs.sequence_number DESC, cs.timestamp DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function: Get conversation events for replay
CREATE OR REPLACE FUNCTION get_conversation_events_for_replay(
    p_conversation_id VARCHAR,
    p_since_sequence BIGINT DEFAULT 0
)
RETURNS TABLE (
    event_id VARCHAR,
    event_type VARCHAR,
    sequence_number BIGINT,
    timestamp TIMESTAMP,
    message_id VARCHAR,
    kafka_topic VARCHAR,
    kafka_partition INT,
    kafka_offset BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ce.event_id,
        ce.event_type,
        ce.sequence_number,
        ce.timestamp,
        ce.message_id,
        ce.kafka_topic,
        ce.kafka_partition,
        ce.kafka_offset
    FROM conversation_events ce
    WHERE ce.conversation_id = p_conversation_id
      AND ce.sequence_number > p_since_sequence
    ORDER BY ce.sequence_number ASC, ce.timestamp ASC;
END;
$$ LANGUAGE plpgsql;

-- Function: Get compression statistics
CREATE OR REPLACE FUNCTION get_compression_stats(p_start_time TIMESTAMP, p_end_time TIMESTAMP)
RETURNS TABLE (
    total_compressions BIGINT,
    avg_compression_ratio FLOAT,
    avg_original_tokens BIGINT,
    avg_compressed_tokens BIGINT,
    compression_types JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*)::BIGINT as total_compressions,
        AVG(compression_ratio) as avg_compression_ratio,
        AVG(original_tokens)::BIGINT as avg_original_tokens,
        AVG(compressed_tokens)::BIGINT as avg_compressed_tokens,
        jsonb_object_agg(
            compression_type,
            COUNT(*)
        ) as compression_types
    FROM conversation_compressions
    WHERE compressed_at >= p_start_time
      AND compressed_at <= p_end_time
    GROUP BY NULL;
END;
$$ LANGUAGE plpgsql;

-- Function: Cleanup expired cache entries
CREATE OR REPLACE FUNCTION cleanup_expired_cache()
RETURNS INT AS $$
DECLARE
    deleted_count INT;
BEGIN
    DELETE FROM conversation_context_cache
    WHERE EXTRACT(EPOCH FROM (NOW() - cached_at)) > ttl;

    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Triggers for Automatic Updates
-- ============================================================================

-- Trigger: Update context cache access timestamp
CREATE OR REPLACE FUNCTION update_cache_access()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_accessed_at = NOW();
    NEW.access_count = NEW.access_count + 1;
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_cache_access
BEFORE UPDATE ON conversation_context_cache
FOR EACH ROW
EXECUTE FUNCTION update_cache_access();

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE conversation_events IS 'Conversation event pointers (full events stored in Kafka)';
COMMENT ON TABLE conversation_compressions IS 'Context compression operations and results';
COMMENT ON TABLE conversation_snapshots IS 'Periodic conversation state snapshots for faster recovery';
COMMENT ON TABLE conversation_context_cache IS 'LRU cache for replayed conversations with TTL';
COMMENT ON TABLE conversation_replay_stats IS 'Statistics for conversation replay and compression operations';

COMMENT ON COLUMN conversation_events.kafka_offset IS 'Kafka offset for event replay from stream';
COMMENT ON COLUMN conversation_compressions.compression_ratio IS 'Ratio of compressed to original tokens (0.0-1.0)';
COMMENT ON COLUMN conversation_snapshots.snapshot_type IS 'Type of snapshot: periodic, compression, checkpoint';
COMMENT ON COLUMN conversation_context_cache.ttl IS 'Time-to-live in seconds before cache expiration';
