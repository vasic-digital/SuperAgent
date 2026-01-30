-- ============================================================================
-- Streaming Analytics Schema
-- Kafka Streams real-time conversation processing and analytics
-- ============================================================================

-- Enable pgcrypto for UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- Conversation State Snapshots
-- Real-time aggregated state from Kafka Streams
-- ============================================================================

CREATE TABLE IF NOT EXISTS conversation_state_snapshots (
    conversation_id UUID PRIMARY KEY,
    user_id UUID,
    session_id UUID,
    message_count INT NOT NULL DEFAULT 0,
    entity_count INT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    compressed_context JSONB,          -- Compressed conversation history
    latest_entities JSONB,             -- Entity graph snapshot
    last_updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 0, -- Optimistic locking
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for conversation state snapshots
CREATE INDEX IF NOT EXISTS idx_conv_state_user_id ON conversation_state_snapshots(user_id);
CREATE INDEX IF NOT EXISTS idx_conv_state_session_id ON conversation_state_snapshots(session_id);
CREATE INDEX IF NOT EXISTS idx_conv_state_updated_at ON conversation_state_snapshots(last_updated_at);
CREATE INDEX IF NOT EXISTS idx_conv_state_entities ON conversation_state_snapshots USING GIN (latest_entities);

-- ============================================================================
-- Conversation Analytics
-- Windowed analytics from stream processing
-- ============================================================================

CREATE TABLE IF NOT EXISTS conversation_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID,
    window_start TIMESTAMP NOT NULL,
    window_end TIMESTAMP NOT NULL,
    total_messages INT NOT NULL DEFAULT 0,
    llm_calls INT NOT NULL DEFAULT 0,
    debate_rounds INT NOT NULL DEFAULT 0,
    avg_response_time_ms FLOAT,
    entity_growth INT NOT NULL DEFAULT 0,           -- New entities discovered in window
    knowledge_density FLOAT,                        -- Entities per message
    provider_distribution JSONB,                    -- Provider usage breakdown
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for conversation analytics
CREATE INDEX IF NOT EXISTS idx_conv_analytics_conversation ON conversation_analytics(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conv_analytics_window_start ON conversation_analytics(window_start);
CREATE INDEX IF NOT EXISTS idx_conv_analytics_window_end ON conversation_analytics(window_end);
CREATE INDEX IF NOT EXISTS idx_conv_analytics_created_at ON conversation_analytics(created_at);
CREATE INDEX IF NOT EXISTS idx_conv_analytics_provider_dist ON conversation_analytics USING GIN (provider_distribution);

-- Composite index for time range queries
CREATE INDEX IF NOT EXISTS idx_conv_analytics_time_range
    ON conversation_analytics(conversation_id, window_start, window_end);

-- ============================================================================
-- TimescaleDB Support (if available)
-- Convert to hypertable for better time-series performance
-- ============================================================================

-- This will only execute if TimescaleDB extension is installed
DO $$
BEGIN
    -- Check if TimescaleDB is available
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        -- Create hypertable for conversation_analytics
        PERFORM create_hypertable(
            'conversation_analytics',
            'window_start',
            if_not_exists => TRUE,
            migrate_data => TRUE
        );

        RAISE NOTICE 'Created TimescaleDB hypertable for conversation_analytics';
    ELSE
        RAISE NOTICE 'TimescaleDB not available, using regular table';
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'TimescaleDB hypertable creation skipped: %', SQLERRM;
END
$$;

-- ============================================================================
-- Entity Tracking
-- Track entities extracted from conversations
-- ============================================================================

CREATE TABLE IF NOT EXISTS stream_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL,
    message_id VARCHAR(255),
    entity_id VARCHAR(255) NOT NULL,
    entity_name VARCHAR(500) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    properties JSONB,
    importance FLOAT DEFAULT 0.0,
    extracted_at TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(conversation_id, entity_id)
);

-- Indexes for entity tracking
CREATE INDEX IF NOT EXISTS idx_stream_entities_conversation ON stream_entities(conversation_id);
CREATE INDEX IF NOT EXISTS idx_stream_entities_type ON stream_entities(entity_type);
CREATE INDEX IF NOT EXISTS idx_stream_entities_name ON stream_entities(entity_name);
CREATE INDEX IF NOT EXISTS idx_stream_entities_extracted_at ON stream_entities(extracted_at);
CREATE INDEX IF NOT EXISTS idx_stream_entities_importance ON stream_entities(importance DESC);
CREATE INDEX IF NOT EXISTS idx_stream_entities_properties ON stream_entities USING GIN (properties);

-- ============================================================================
-- Provider Metrics
-- Aggregated provider performance metrics
-- ============================================================================

CREATE TABLE IF NOT EXISTS provider_stream_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(100) NOT NULL,
    model VARCHAR(200),
    window_start TIMESTAMP NOT NULL,
    window_end TIMESTAMP NOT NULL,
    request_count INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    failure_count INT NOT NULL DEFAULT 0,
    avg_response_time_ms FLOAT,
    p50_response_time_ms FLOAT,
    p95_response_time_ms FLOAT,
    p99_response_time_ms FLOAT,
    total_tokens BIGINT DEFAULT 0,
    total_cost DECIMAL(10, 4) DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for provider metrics
CREATE INDEX IF NOT EXISTS idx_provider_metrics_provider ON provider_stream_metrics(provider);
CREATE INDEX IF NOT EXISTS idx_provider_metrics_model ON provider_stream_metrics(model);
CREATE INDEX IF NOT EXISTS idx_provider_metrics_window_start ON provider_stream_metrics(window_start);
CREATE INDEX IF NOT EXISTS idx_provider_metrics_window_end ON provider_stream_metrics(window_end);
CREATE INDEX IF NOT EXISTS idx_provider_metrics_created_at ON provider_stream_metrics(created_at);

-- Composite index for provider performance queries
CREATE INDEX IF NOT EXISTS idx_provider_metrics_performance
    ON provider_stream_metrics(provider, model, window_start, window_end);

-- ============================================================================
-- Debate Round Analytics
-- Detailed debate round performance tracking
-- ============================================================================

CREATE TABLE IF NOT EXISTS debate_round_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL,
    debate_id VARCHAR(255),
    round_number INT NOT NULL,
    position INT NOT NULL,
    role VARCHAR(50),
    provider VARCHAR(100) NOT NULL,
    model VARCHAR(200) NOT NULL,
    response_time_ms BIGINT NOT NULL,
    tokens_used INT NOT NULL,
    confidence_score FLOAT,
    error_count INT DEFAULT 0,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for debate round analytics
CREATE INDEX IF NOT EXISTS idx_debate_analytics_conversation ON debate_round_analytics(conversation_id);
CREATE INDEX IF NOT EXISTS idx_debate_analytics_debate_id ON debate_round_analytics(debate_id);
CREATE INDEX IF NOT EXISTS idx_debate_analytics_provider ON debate_round_analytics(provider);
CREATE INDEX IF NOT EXISTS idx_debate_analytics_timestamp ON debate_round_analytics(timestamp);

-- Composite index for debate performance analysis
CREATE INDEX IF NOT EXISTS idx_debate_analytics_performance
    ON debate_round_analytics(provider, model, round_number, response_time_ms);

-- ============================================================================
-- Materialized Views for Common Queries
-- ============================================================================

-- View: Recent conversation activity (last 24 hours)
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_recent_conversation_activity AS
SELECT
    conversation_id,
    MAX(window_end) as last_activity,
    SUM(total_messages) as total_messages,
    SUM(debate_rounds) as total_debate_rounds,
    AVG(avg_response_time_ms) as avg_response_time,
    AVG(knowledge_density) as avg_knowledge_density
FROM conversation_analytics
WHERE window_start >= NOW() - INTERVAL '24 hours'
GROUP BY conversation_id;

-- Index on materialized view
CREATE INDEX IF NOT EXISTS idx_mv_recent_activity_conversation
    ON mv_recent_conversation_activity(conversation_id);

-- View: Provider performance summary (last 7 days)
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_provider_performance_summary AS
SELECT
    provider,
    model,
    COUNT(*) as total_requests,
    AVG(avg_response_time_ms) as avg_response_time,
    SUM(request_count) as total_request_count,
    SUM(success_count) as total_success_count,
    SUM(failure_count) as total_failure_count,
    CASE
        WHEN SUM(request_count) > 0
        THEN (SUM(success_count)::FLOAT / SUM(request_count)::FLOAT * 100)
        ELSE 0
    END as success_rate_percent
FROM provider_stream_metrics
WHERE window_start >= NOW() - INTERVAL '7 days'
GROUP BY provider, model;

-- Index on materialized view
CREATE INDEX IF NOT EXISTS idx_mv_provider_summary_provider
    ON mv_provider_performance_summary(provider, model);

-- ============================================================================
-- Functions for Analytics
-- ============================================================================

-- Function: Refresh all materialized views
CREATE OR REPLACE FUNCTION refresh_streaming_analytics_views()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_recent_conversation_activity;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_provider_performance_summary;
END;
$$ LANGUAGE plpgsql;

-- Function: Get conversation summary
CREATE OR REPLACE FUNCTION get_conversation_summary(p_conversation_id UUID)
RETURNS TABLE (
    total_messages BIGINT,
    total_entities BIGINT,
    total_tokens BIGINT,
    debate_rounds BIGINT,
    duration_minutes FLOAT,
    knowledge_density FLOAT,
    top_providers JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COALESCE(MAX(css.message_count)::BIGINT, 0) as total_messages,
        COALESCE(MAX(css.entity_count)::BIGINT, 0) as total_entities,
        COALESCE(MAX(css.total_tokens), 0) as total_tokens,
        COALESCE(SUM(ca.debate_rounds)::BIGINT, 0) as debate_rounds,
        EXTRACT(EPOCH FROM (MAX(ca.window_end) - MIN(ca.window_start))) / 60.0 as duration_minutes,
        COALESCE(AVG(ca.knowledge_density), 0.0) as knowledge_density,
        jsonb_object_agg(
            provider_key,
            provider_value::int
        ) FILTER (WHERE provider_key IS NOT NULL) as top_providers
    FROM conversation_state_snapshots css
    LEFT JOIN conversation_analytics ca ON ca.conversation_id = css.conversation_id
    LEFT JOIN LATERAL (
        SELECT provider_key, provider_value
        FROM jsonb_each_text(ca.provider_distribution)
        ORDER BY provider_value::int DESC
        LIMIT 5
    ) providers ON true
    WHERE css.conversation_id = p_conversation_id
    GROUP BY css.conversation_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE conversation_state_snapshots IS 'Real-time conversation state from Kafka Streams processing';
COMMENT ON TABLE conversation_analytics IS 'Windowed analytics aggregated from conversation events';
COMMENT ON TABLE stream_entities IS 'Entities extracted from conversation messages in real-time';
COMMENT ON TABLE provider_stream_metrics IS 'Provider performance metrics from stream processing';
COMMENT ON TABLE debate_round_analytics IS 'Detailed analytics for individual debate rounds';

COMMENT ON COLUMN conversation_state_snapshots.compressed_context IS 'Compressed conversation history for infinite context';
COMMENT ON COLUMN conversation_state_snapshots.version IS 'Optimistic locking version for concurrent updates';
COMMENT ON COLUMN conversation_analytics.knowledge_density IS 'Ratio of entities to messages (higher = more information-dense)';
COMMENT ON COLUMN conversation_analytics.provider_distribution IS 'JSON object mapping provider names to usage counts';
