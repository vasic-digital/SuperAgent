-- ============================================================================
-- ClickHouse Time-Series Analytics Schema
-- Real-time metrics for debates, conversations, and provider performance
-- ============================================================================

-- ===========================================
-- DEBATE METRICS TABLE
-- ===========================================

CREATE TABLE IF NOT EXISTS debate_metrics (
    debate_id String,
    round UInt8,
    timestamp DateTime,
    provider String,
    model String,
    position String,
    response_time_ms Float32,
    tokens_used UInt32,
    confidence_score Float32,
    error_count UInt8,
    was_winner UInt8  -- Boolean as UInt8
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, debate_id, round);

-- Indexes for faster queries
ALTER TABLE debate_metrics ADD INDEX idx_provider provider TYPE minmax GRANULARITY 4;
ALTER TABLE debate_metrics ADD INDEX idx_model model TYPE minmax GRANULARITY 4;
ALTER TABLE debate_metrics ADD INDEX idx_position position TYPE minmax GRANULARITY 4;

-- ===========================================
-- DEBATE METRICS HOURLY AGGREGATION
-- ===========================================

CREATE MATERIALIZED VIEW IF NOT EXISTS debate_metrics_hourly
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (hour, provider)
AS SELECT
    toStartOfHour(timestamp) AS hour,
    provider,
    COUNT(*) AS total_requests,
    AVG(response_time_ms) AS avg_response_time,
    quantile(0.95)(response_time_ms) AS p95_response_time,
    SUM(tokens_used) AS total_tokens,
    AVG(confidence_score) AS avg_confidence,
    SUM(was_winner) AS total_wins
FROM debate_metrics
GROUP BY hour, provider;

-- ===========================================
-- CONVERSATION METRICS TABLE
-- ===========================================

CREATE TABLE IF NOT EXISTS conversation_metrics (
    conversation_id String,
    user_id String,
    timestamp DateTime,
    message_count UInt32,
    entity_count UInt32,
    total_tokens UInt64,
    duration_ms UInt64,
    debate_rounds UInt8,
    llms_used Array(String)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, user_id, conversation_id);

-- Indexes
ALTER TABLE conversation_metrics ADD INDEX idx_user_id user_id TYPE minmax GRANULARITY 4;
ALTER TABLE conversation_metrics ADD INDEX idx_llms_used llms_used TYPE bloom_filter GRANULARITY 4;

-- ===========================================
-- CONVERSATION METRICS DAILY AGGREGATION
-- ===========================================

CREATE MATERIALIZED VIEW IF NOT EXISTS conversation_metrics_daily
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(day)
ORDER BY (day, user_id)
AS SELECT
    toDate(timestamp) AS day,
    user_id,
    COUNT(*) AS total_conversations,
    AVG(message_count) AS avg_messages,
    AVG(entity_count) AS avg_entities,
    AVG(total_tokens) AS avg_tokens,
    AVG(duration_ms) AS avg_duration_ms,
    AVG(debate_rounds) AS avg_debate_rounds
FROM conversation_metrics
GROUP BY day, user_id;

-- ===========================================
-- PROVIDER PERFORMANCE TABLE
-- ===========================================

CREATE TABLE IF NOT EXISTS provider_performance (
    timestamp DateTime,
    provider String,
    model String,
    total_requests UInt64,
    successful_requests UInt64,
    failed_requests UInt64,
    avg_response_time Float32,
    p50_response_time Float32,
    p95_response_time Float32,
    p99_response_time Float32,
    total_tokens UInt64,
    avg_tokens_per_req Float32,
    total_cost Float64,
    avg_cost_per_req Float64
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, provider, model);

-- ===========================================
-- LLM RESPONSE LATENCY TABLE
-- ===========================================

CREATE TABLE IF NOT EXISTS llm_response_latency (
    timestamp DateTime,
    provider String,
    model String,
    operation String,  -- 'complete', 'stream', 'embed'
    latency_ms Float32,
    tokens_input UInt32,
    tokens_output UInt32,
    cache_hit UInt8  -- Boolean
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, provider, operation);

-- ===========================================
-- ENTITY EXTRACTION METRICS TABLE
-- ===========================================

CREATE TABLE IF NOT EXISTS entity_extraction_metrics (
    timestamp DateTime,
    conversation_id String,
    entity_id String,
    entity_type String,
    extraction_method String,  -- 'llm', 'rule-based', 'hybrid'
    confidence Float32,
    processing_time_ms Float32
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, conversation_id, entity_id);

-- Index for entity type queries
ALTER TABLE entity_extraction_metrics ADD INDEX idx_entity_type entity_type TYPE minmax GRANULARITY 4;

-- ===========================================
-- MEMORY OPERATIONS TABLE
-- ===========================================

CREATE TABLE IF NOT EXISTS memory_operations (
    timestamp DateTime,
    user_id String,
    operation String,  -- 'add', 'update', 'search', 'delete'
    memory_id String,
    duration_ms Float32,
    success UInt8,  -- Boolean
    error_message String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, user_id, operation);

-- ===========================================
-- DEBATE WINNER ANALYSIS TABLE
-- ===========================================

CREATE TABLE IF NOT EXISTS debate_winners (
    debate_id String,
    timestamp DateTime,
    winner_provider String,
    winner_model String,
    winner_position String,
    total_rounds UInt8,
    final_confidence Float32,
    debate_duration_ms UInt64
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, debate_id);

-- ===========================================
-- SYSTEM HEALTH METRICS TABLE
-- ===========================================

CREATE TABLE IF NOT EXISTS system_health (
    timestamp DateTime,
    component String,  -- 'api', 'kafka', 'redis', 'neo4j', 'clickhouse', etc.
    metric_name String,
    metric_value Float64,
    unit String,
    status String  -- 'healthy', 'degraded', 'unhealthy'
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, component, metric_name);

-- ===========================================
-- API REQUEST METRICS TABLE
-- ===========================================

CREATE TABLE IF NOT EXISTS api_requests (
    timestamp DateTime,
    endpoint String,
    method String,  -- GET, POST, PUT, DELETE
    status_code UInt16,
    response_time_ms Float32,
    user_id String,
    session_id String,
    error_message String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, endpoint, method);

-- Index for endpoint queries
ALTER TABLE api_requests ADD INDEX idx_endpoint endpoint TYPE minmax GRANULARITY 4;
ALTER TABLE api_requests ADD INDEX idx_status_code status_code TYPE minmax GRANULARITY 4;

-- ===========================================
-- API REQUESTS MINUTELY AGGREGATION
-- ===========================================

CREATE MATERIALIZED VIEW IF NOT EXISTS api_requests_minutely
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(minute)
ORDER BY (minute, endpoint)
AS SELECT
    toStartOfMinute(timestamp) AS minute,
    endpoint,
    method,
    COUNT(*) AS total_requests,
    AVG(response_time_ms) AS avg_response_time,
    quantile(0.95)(response_time_ms) AS p95_response_time,
    countIf(status_code >= 200 AND status_code < 300) AS success_count,
    countIf(status_code >= 400) AS error_count
FROM api_requests
GROUP BY minute, endpoint, method;

-- ===========================================
-- HELPER FUNCTIONS
-- ===========================================

-- Function to get provider performance summary
CREATE OR REPLACE FUNCTION get_provider_summary(time_window_hours UInt32) AS (
    SELECT
        provider,
        COUNT(*) as total_requests,
        AVG(response_time_ms) as avg_response_time,
        quantile(0.95)(response_time_ms) as p95_response_time,
        AVG(confidence_score) as avg_confidence,
        SUM(tokens_used) as total_tokens,
        SUM(was_winner) as total_wins,
        SUM(was_winner) / COUNT(*) as win_rate
    FROM debate_metrics
    WHERE timestamp >= now() - INTERVAL time_window_hours HOUR
    GROUP BY provider
    ORDER BY avg_confidence DESC
);

-- ===========================================
-- COMMENTS
-- ===========================================

-- Debate metrics store individual debate round performance
COMMENT ON TABLE debate_metrics IS 'Time-series metrics for AI debate performance tracking';

-- Conversation metrics track user conversation statistics
COMMENT ON TABLE conversation_metrics IS 'Conversation-level metrics including messages, entities, and tokens';

-- Provider performance aggregates provider statistics
COMMENT ON TABLE provider_performance IS 'Aggregated provider performance metrics';

-- LLM response latency tracks API call latencies
COMMENT ON TABLE llm_response_latency IS 'LLM API response time tracking for performance analysis';

-- Entity extraction metrics track entity discovery
COMMENT ON TABLE entity_extraction_metrics IS 'Entity extraction performance and confidence tracking';

-- Memory operations track memory system usage
COMMENT ON TABLE memory_operations IS 'Memory system operation logs and performance';

-- Debate winners track debate outcomes
COMMENT ON TABLE debate_winners IS 'Debate winner tracking for analysis';

-- System health tracks infrastructure health
COMMENT ON TABLE system_health IS 'System component health monitoring';

-- API requests track REST API usage
COMMENT ON TABLE api_requests IS 'REST API request logs and performance';
