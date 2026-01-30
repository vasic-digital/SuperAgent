-- ============================================================================
-- Distributed Memory Schema
-- Event sourcing and CRDT-based multi-node memory synchronization
-- ============================================================================

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- Memory Events Table
-- Stores all memory change events for event sourcing
-- ============================================================================

CREATE TABLE IF NOT EXISTS memory_events (
    event_id VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    node_id VARCHAR(255) NOT NULL,

    -- Memory identification
    memory_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255),

    -- Memory content
    content TEXT,
    embedding vector(1536),  -- Assuming OpenAI embeddings
    importance FLOAT DEFAULT 0.0,
    tags TEXT[],

    -- Entity and relationship data
    entities JSONB,
    relationships JSONB,
    metadata JSONB,

    -- CRDT versioning
    version BIGINT NOT NULL,
    vector_clock JSONB NOT NULL,

    -- Merge tracking
    merged_from TEXT[],

    -- Indexes
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for memory events
CREATE INDEX IF NOT EXISTS idx_memory_events_memory_id ON memory_events(memory_id);
CREATE INDEX IF NOT EXISTS idx_memory_events_user_id ON memory_events(user_id);
CREATE INDEX IF NOT EXISTS idx_memory_events_node_id ON memory_events(node_id);
CREATE INDEX IF NOT EXISTS idx_memory_events_timestamp ON memory_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_memory_events_event_type ON memory_events(event_type);
CREATE INDEX IF NOT EXISTS idx_memory_events_session_id ON memory_events(session_id) WHERE session_id IS NOT NULL;

-- Composite index for common queries
CREATE INDEX IF NOT EXISTS idx_memory_events_user_timestamp
    ON memory_events(user_id, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_memory_events_memory_timestamp
    ON memory_events(memory_id, timestamp DESC);

-- GIN indexes for JSON columns
CREATE INDEX IF NOT EXISTS idx_memory_events_entities ON memory_events USING GIN (entities);
CREATE INDEX IF NOT EXISTS idx_memory_events_relationships ON memory_events USING GIN (relationships);
CREATE INDEX IF NOT EXISTS idx_memory_events_vector_clock ON memory_events USING GIN (vector_clock);
CREATE INDEX IF NOT EXISTS idx_memory_events_metadata ON memory_events USING GIN (metadata);

-- ============================================================================
-- Memory Snapshots Table
-- Periodic snapshots of memory state for faster recovery
-- ============================================================================

CREATE TABLE IF NOT EXISTS memory_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    node_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,

    -- Snapshot data
    memories JSONB NOT NULL,
    entities JSONB,
    relationships JSONB,
    vector_clock JSONB NOT NULL,
    metadata JSONB,

    -- Snapshot metadata
    memory_count INT NOT NULL DEFAULT 0,
    entity_count INT NOT NULL DEFAULT 0,
    relationship_count INT NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for snapshots
CREATE INDEX IF NOT EXISTS idx_memory_snapshots_user_id ON memory_snapshots(user_id);
CREATE INDEX IF NOT EXISTS idx_memory_snapshots_node_id ON memory_snapshots(node_id);
CREATE INDEX IF NOT EXISTS idx_memory_snapshots_timestamp ON memory_snapshots(timestamp DESC);

-- Composite index for latest snapshot queries
CREATE INDEX IF NOT EXISTS idx_memory_snapshots_user_timestamp
    ON memory_snapshots(user_id, timestamp DESC);

-- ============================================================================
-- Conflict Resolution Log
-- Tracks conflicts and their resolutions
-- ============================================================================

CREATE TABLE IF NOT EXISTS memory_conflicts (
    conflict_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    memory_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,

    -- Conflict details
    conflict_type VARCHAR(100) NOT NULL,
    local_event_id VARCHAR(255),
    remote_event_id VARCHAR(255),

    -- Resolution
    resolution_strategy VARCHAR(50) NOT NULL,
    resolved_version JSONB,
    resolved_by VARCHAR(255),  -- Node that resolved

    -- Timestamps
    detected_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP,

    -- Additional context
    details JSONB
);

-- Indexes for conflicts
CREATE INDEX IF NOT EXISTS idx_memory_conflicts_memory_id ON memory_conflicts(memory_id);
CREATE INDEX IF NOT EXISTS idx_memory_conflicts_user_id ON memory_conflicts(user_id);
CREATE INDEX IF NOT EXISTS idx_memory_conflicts_detected_at ON memory_conflicts(detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_memory_conflicts_resolved ON memory_conflicts(resolved_at) WHERE resolved_at IS NOT NULL;

-- ============================================================================
-- Node Registry
-- Tracks active nodes in the distributed system
-- ============================================================================

CREATE TABLE IF NOT EXISTS memory_nodes (
    node_id VARCHAR(255) PRIMARY KEY,
    node_name VARCHAR(255),
    node_type VARCHAR(50),  -- 'primary', 'replica', 'worker'

    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'active',  -- 'active', 'inactive', 'failed'
    last_heartbeat TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Vector clock state
    vector_clock JSONB,

    -- Node metadata
    version VARCHAR(50),
    capabilities JSONB,
    metadata JSONB,

    -- Timestamps
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for nodes
CREATE INDEX IF NOT EXISTS idx_memory_nodes_status ON memory_nodes(status);
CREATE INDEX IF NOT EXISTS idx_memory_nodes_last_heartbeat ON memory_nodes(last_heartbeat DESC);

-- ============================================================================
-- Event Stream Checkpoints
-- Tracks consumer positions in event streams
-- ============================================================================

CREATE TABLE IF NOT EXISTS memory_event_checkpoints (
    checkpoint_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),

    -- Checkpoint position
    last_processed_event_id VARCHAR(255) NOT NULL,
    last_processed_timestamp TIMESTAMP NOT NULL,
    last_processed_version BIGINT NOT NULL,

    -- Vector clock at checkpoint
    vector_clock JSONB,

    -- Checkpoint metadata
    event_count BIGINT DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(node_id, user_id)
);

-- Indexes for checkpoints
CREATE INDEX IF NOT EXISTS idx_memory_checkpoints_node_id ON memory_event_checkpoints(node_id);
CREATE INDEX IF NOT EXISTS idx_memory_checkpoints_user_id ON memory_event_checkpoints(user_id);
CREATE INDEX IF NOT EXISTS idx_memory_checkpoints_updated_at ON memory_event_checkpoints(updated_at DESC);

-- ============================================================================
-- Functions for Event Sourcing
-- ============================================================================

-- Function: Get memory event history
CREATE OR REPLACE FUNCTION get_memory_event_history(p_memory_id VARCHAR)
RETURNS TABLE (
    event_id VARCHAR,
    event_type VARCHAR,
    timestamp TIMESTAMP,
    node_id VARCHAR,
    version BIGINT,
    content TEXT,
    importance FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        me.event_id,
        me.event_type,
        me.timestamp,
        me.node_id,
        me.version,
        me.content,
        me.importance
    FROM memory_events me
    WHERE me.memory_id = p_memory_id
    ORDER BY me.timestamp ASC;
END;
$$ LANGUAGE plpgsql;

-- Function: Get events since timestamp for synchronization
CREATE OR REPLACE FUNCTION get_events_since(p_timestamp TIMESTAMP, p_user_id VARCHAR DEFAULT NULL)
RETURNS TABLE (
    event_id VARCHAR,
    event_type VARCHAR,
    timestamp TIMESTAMP,
    node_id VARCHAR,
    memory_id VARCHAR,
    user_id VARCHAR,
    content TEXT,
    vector_clock JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        me.event_id,
        me.event_type,
        me.timestamp,
        me.node_id,
        me.memory_id,
        me.user_id,
        me.content,
        me.vector_clock
    FROM memory_events me
    WHERE me.timestamp > p_timestamp
      AND (p_user_id IS NULL OR me.user_id = p_user_id)
    ORDER BY me.timestamp ASC;
END;
$$ LANGUAGE plpgsql;

-- Function: Rebuild memory from events
CREATE OR REPLACE FUNCTION rebuild_memory_from_events(p_memory_id VARCHAR)
RETURNS JSONB AS $$
DECLARE
    result JSONB;
BEGIN
    -- Get the latest event for the memory
    SELECT jsonb_build_object(
        'memory_id', me.memory_id,
        'user_id', me.user_id,
        'content', me.content,
        'importance', me.importance,
        'tags', me.tags,
        'entities', me.entities,
        'relationships', me.relationships,
        'metadata', me.metadata,
        'version', me.version,
        'created_at', MIN(me.timestamp),
        'updated_at', MAX(me.timestamp)
    )
    INTO result
    FROM memory_events me
    WHERE me.memory_id = p_memory_id
      AND me.event_type != 'memory.deleted'
    GROUP BY me.memory_id, me.user_id, me.content, me.importance,
             me.tags, me.entities, me.relationships, me.metadata, me.version;

    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Function: Get conflict statistics
CREATE OR REPLACE FUNCTION get_conflict_stats(p_start_time TIMESTAMP, p_end_time TIMESTAMP)
RETURNS TABLE (
    total_conflicts BIGINT,
    resolved_conflicts BIGINT,
    pending_conflicts BIGINT,
    avg_resolution_time INTERVAL,
    conflict_types JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*)::BIGINT as total_conflicts,
        COUNT(CASE WHEN resolved_at IS NOT NULL THEN 1 END)::BIGINT as resolved_conflicts,
        COUNT(CASE WHEN resolved_at IS NULL THEN 1 END)::BIGINT as pending_conflicts,
        AVG(resolved_at - detected_at) as avg_resolution_time,
        jsonb_object_agg(
            conflict_type,
            COUNT(*)
        ) as conflict_types
    FROM memory_conflicts
    WHERE detected_at >= p_start_time
      AND detected_at <= p_end_time
    GROUP BY NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Triggers for Automatic Updates
-- ============================================================================

-- Trigger: Update node last heartbeat
CREATE OR REPLACE FUNCTION update_node_heartbeat()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_node_heartbeat
BEFORE UPDATE ON memory_nodes
FOR EACH ROW
EXECUTE FUNCTION update_node_heartbeat();

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE memory_events IS 'Event sourcing log for distributed memory synchronization';
COMMENT ON TABLE memory_snapshots IS 'Periodic snapshots of memory state for faster recovery and replay';
COMMENT ON TABLE memory_conflicts IS 'Log of detected conflicts and their CRDT-based resolutions';
COMMENT ON TABLE memory_nodes IS 'Registry of active nodes in the distributed memory system';
COMMENT ON TABLE memory_event_checkpoints IS 'Consumer positions in event streams for catch-up synchronization';

COMMENT ON COLUMN memory_events.vector_clock IS 'JSONB vector clock for causal ordering of events';
COMMENT ON COLUMN memory_events.version IS 'Lamport timestamp for event ordering';
COMMENT ON COLUMN memory_events.merged_from IS 'Array of memory IDs that were merged into this memory';
COMMENT ON COLUMN memory_conflicts.resolution_strategy IS 'CRDT strategy used: last_write_wins, merge_all, importance, vector_clock';
