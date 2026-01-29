-- =============================================================================
-- HelixAgent SQL Schema: Performance Indexes & Materialized Views
-- =============================================================================
-- Domain: Query optimization, analytics dashboards, and operational monitoring.
-- Source migrations: 012_performance_indexes.sql, 013_materialized_views.sql
--
-- This file consolidates all performance indexes (beyond basic table indexes)
-- and all materialized views used for analytics and dashboards.
-- =============================================================================

-- =============================================================================
-- SECTION 1: PERFORMANCE INDEXES (Migration 012)
-- =============================================================================
-- These indexes are created CONCURRENTLY to avoid blocking writes in
-- production. They target hot query paths identified through query analysis.

-- -----------------------------------------------------------------------------
-- 1.1 Provider Performance Indexes
-- -----------------------------------------------------------------------------

-- Hot path: routing requests to healthy, enabled providers
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_providers_healthy_enabled
    ON llm_providers (name, health_status, response_time)
    WHERE enabled = TRUE AND health_status = 'healthy';

-- Weight-based provider selection for ensemble voting
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_providers_by_weight
    ON llm_providers (weight DESC, response_time ASC)
    WHERE enabled = TRUE;

-- Covering index for health-check dashboard (avoids table access)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_providers_health_check
    ON llm_providers (id, name, health_status, response_time, updated_at)
    WHERE enabled = TRUE;

-- -----------------------------------------------------------------------------
-- 1.2 Request/Response Performance Indexes
-- -----------------------------------------------------------------------------

-- Session-scoped request lookup with status filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_requests_session_status
    ON llm_requests (session_id, status, created_at DESC)
    WHERE status IN ('pending', 'completed');

-- Partial index for analytics on recent requests (last 24 hours)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_requests_recent
    ON llm_requests (created_at DESC)
    WHERE created_at > NOW() - INTERVAL '24 hours';

-- Response selection queries (ensemble voting result lookup)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_responses_selection
    ON llm_responses (request_id, selected, selection_score DESC);

-- Covering index for response aggregation (avoids table lookups)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_responses_aggregation
    ON llm_responses (provider_name, created_at DESC)
    INCLUDE (response_time, tokens_used, confidence);

-- -----------------------------------------------------------------------------
-- 1.3 Session Performance Indexes
-- -----------------------------------------------------------------------------

-- Active session lookups (hot path)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_active
    ON user_sessions (user_id, status, last_activity DESC)
    WHERE status = 'active';

-- Session expiration cleanup job
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_expired
    ON user_sessions (expires_at)
    WHERE status = 'active' AND expires_at < NOW();

-- -----------------------------------------------------------------------------
-- 1.4 MCP/Protocol Server Indexes
-- -----------------------------------------------------------------------------

-- Active MCP server lookup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_mcp_active_servers
    ON mcp_servers (type, enabled, last_sync DESC)
    WHERE enabled = TRUE;

-- Time-series covering index for protocol metrics dashboards
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_protocol_metrics_timeseries
    ON protocol_metrics (protocol_type, created_at DESC)
    INCLUDE (status, duration_ms);

-- Server-scoped aggregation for per-server dashboards
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_protocol_metrics_by_server
    ON protocol_metrics (server_id, operation, status)
    WHERE server_id IS NOT NULL;

-- BRIN index for efficient time-range scans (append-only table)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_protocol_metrics_brin
    ON protocol_metrics USING BRIN (created_at);

-- -----------------------------------------------------------------------------
-- 1.5 Cache Indexes
-- -----------------------------------------------------------------------------

-- Expired cache entry cleanup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cache_expiration
    ON protocol_cache (expires_at)
    WHERE expires_at < NOW();

-- -----------------------------------------------------------------------------
-- 1.6 Model Metadata Performance Indexes
-- -----------------------------------------------------------------------------

-- Capability-based model lookup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_models_by_capabilities
    ON models_metadata (provider_name, supports_streaming, supports_function_calling)
    WHERE provider_name IS NOT NULL;

-- Scored model ranking
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_models_scored
    ON models_metadata (benchmark_score DESC, reliability_score DESC)
    WHERE benchmark_score IS NOT NULL;

-- GIN index for tag-based model search
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_models_tags_gin
    ON models_metadata USING GIN (tags);

-- GIN index for protocol support queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_models_protocol_gin
    ON models_metadata USING GIN (protocol_support);

-- -----------------------------------------------------------------------------
-- 1.7 Background Tasks Performance Indexes
-- -----------------------------------------------------------------------------

-- Priority queue ordering (critical for dequeue_background_task function)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_queue_order
    ON background_tasks (priority, scheduled_at ASC, created_at ASC)
    WHERE status = 'pending' AND deleted_at IS NULL;

-- Stuck task detection
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_running_heartbeat
    ON background_tasks (last_heartbeat, started_at)
    WHERE status = 'running';

-- Task completion tracking
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_completion
    ON background_tasks (completed_at DESC, status)
    WHERE status IN ('completed', 'failed');

-- -----------------------------------------------------------------------------
-- 1.8 Cognee Memory Indexes
-- -----------------------------------------------------------------------------

-- Full-text search on memory content for semantic retrieval
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cognee_content_fts
    ON cognee_memories USING GIN (to_tsvector('english', content));

-- Recent memories within a dataset
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cognee_dataset_recent
    ON cognee_memories (dataset_name, created_at DESC);

-- -----------------------------------------------------------------------------
-- 1.9 Partial Indexes for Common Filters
-- -----------------------------------------------------------------------------

-- Pending/failed webhook retry queue
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhooks_pending_retry
    ON webhook_deliveries (created_at, attempts)
    WHERE status = 'pending' OR status = 'failed';

-- Dead letter reprocessing queue
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_dead_letter_ready
    ON background_tasks_dead_letter (reprocess_after ASC)
    WHERE NOT reprocessed AND reprocess_after IS NOT NULL;

-- -----------------------------------------------------------------------------
-- 1.10 Statistics Targets (Improve Query Planner Accuracy)
-- -----------------------------------------------------------------------------

ALTER TABLE llm_providers    ALTER COLUMN health_status  SET STATISTICS 500;
ALTER TABLE llm_providers    ALTER COLUMN enabled        SET STATISTICS 500;
ALTER TABLE llm_requests     ALTER COLUMN status         SET STATISTICS 500;
ALTER TABLE llm_responses    ALTER COLUMN provider_name  SET STATISTICS 500;
ALTER TABLE background_tasks ALTER COLUMN status         SET STATISTICS 500;
ALTER TABLE background_tasks ALTER COLUMN priority       SET STATISTICS 500;
ALTER TABLE protocol_metrics ALTER COLUMN protocol_type  SET STATISTICS 500;

-- Index comments
COMMENT ON INDEX idx_providers_healthy_enabled IS 'Hot path index for routing requests to healthy providers';
COMMENT ON INDEX idx_requests_recent IS 'Partial index for analytics queries on recent requests';
COMMENT ON INDEX idx_sessions_active IS 'Hot path index for active session lookups';
COMMENT ON INDEX idx_tasks_queue_order IS 'Priority queue index for task dequeue function';
COMMENT ON INDEX idx_cognee_content_fts IS 'Full-text search on memory content for semantic retrieval';


-- =============================================================================
-- SECTION 2: MATERIALIZED VIEWS (Migration 013)
-- =============================================================================
-- Materialized views pre-compute expensive aggregations for dashboards.
-- Each view has a recommended refresh interval and unique index for
-- CONCURRENTLY refresh support.

-- -----------------------------------------------------------------------------
-- 2.1 mv_provider_performance
-- Refresh: Every 5 minutes
-- Purpose: Provider performance metrics aggregated over 24 hours
-- -----------------------------------------------------------------------------
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_provider_performance AS
SELECT
    p.name AS provider_name,
    p.type AS provider_type,
    p.health_status,
    p.weight,
    p.response_time AS last_response_time_ms,
    COUNT(DISTINCT r.id) AS total_requests_24h,
    COUNT(DISTINCT r.id) FILTER (WHERE resp.selected = TRUE) AS selected_responses_24h,
    AVG(resp.response_time)::INTEGER AS avg_response_time_ms,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY resp.response_time)::INTEGER AS p50_response_time_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY resp.response_time)::INTEGER AS p95_response_time_ms,
    PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY resp.response_time)::INTEGER AS p99_response_time_ms,
    AVG(resp.confidence)::DECIMAL(3,2) AS avg_confidence,
    SUM(resp.tokens_used) AS total_tokens_24h,
    COUNT(*) FILTER (WHERE resp.finish_reason = 'stop')::FLOAT /
        NULLIF(COUNT(*), 0) AS success_rate,
    MAX(resp.created_at) AS last_response_at,
    p.updated_at AS provider_updated_at
FROM llm_providers p
LEFT JOIN llm_responses resp ON resp.provider_name = p.name
    AND resp.created_at > NOW() - INTERVAL '24 hours'
LEFT JOIN llm_requests r ON r.id = resp.request_id
WHERE p.enabled = TRUE
GROUP BY p.id, p.name, p.type, p.health_status, p.weight, p.response_time, p.updated_at
WITH DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_provider_perf_name
    ON mv_provider_performance (provider_name);
CREATE INDEX IF NOT EXISTS idx_mv_provider_perf_score
    ON mv_provider_performance (avg_confidence DESC, success_rate DESC);

COMMENT ON MATERIALIZED VIEW mv_provider_performance
    IS 'Provider performance metrics aggregated over 24 hours, refresh every 5 minutes';

-- -----------------------------------------------------------------------------
-- 2.2 mv_mcp_server_health
-- Refresh: Every 1 minute
-- Purpose: MCP server health and operation metrics
-- -----------------------------------------------------------------------------
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_mcp_server_health AS
SELECT
    s.id AS server_id,
    s.name AS server_name,
    s.type AS server_type,
    s.enabled,
    COUNT(m.id) AS total_operations_1h,
    COUNT(m.id) FILTER (WHERE m.status = 'success') AS successful_operations,
    COUNT(m.id) FILTER (WHERE m.status = 'error') AS failed_operations,
    AVG(m.duration_ms)::INTEGER AS avg_duration_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY m.duration_ms)::INTEGER AS p95_duration_ms,
    MAX(m.created_at) AS last_operation_at,
    COUNT(m.id) FILTER (WHERE m.status = 'success')::FLOAT /
        NULLIF(COUNT(m.id), 0) AS success_rate,
    s.last_sync,
    jsonb_array_length(s.tools) AS tool_count
FROM mcp_servers s
LEFT JOIN protocol_metrics m ON m.server_id = s.id
    AND m.protocol_type = 'mcp'
    AND m.created_at > NOW() - INTERVAL '1 hour'
GROUP BY s.id, s.name, s.type, s.enabled, s.last_sync, s.tools
WITH DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_mcp_health_id
    ON mv_mcp_server_health (server_id);
CREATE INDEX IF NOT EXISTS idx_mv_mcp_health_success
    ON mv_mcp_server_health (success_rate DESC) WHERE enabled = TRUE;

COMMENT ON MATERIALIZED VIEW mv_mcp_server_health
    IS 'MCP server health and operation metrics, refresh every 1 minute';

-- -----------------------------------------------------------------------------
-- 2.3 mv_request_analytics_hourly
-- Refresh: Every 15 minutes
-- Purpose: Hourly request analytics for dashboards
-- -----------------------------------------------------------------------------
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_request_analytics_hourly AS
SELECT
    date_trunc('hour', r.created_at) AS hour,
    COUNT(*) AS total_requests,
    COUNT(*) FILTER (WHERE r.status = 'completed') AS completed_requests,
    COUNT(*) FILTER (WHERE r.status = 'failed') AS failed_requests,
    COUNT(DISTINCT r.user_id) AS unique_users,
    COUNT(DISTINCT r.session_id) AS unique_sessions,
    AVG(EXTRACT(EPOCH FROM (r.completed_at - r.started_at)))::INTEGER AS avg_duration_seconds,
    SUM(CASE WHEN r.memory_enhanced THEN 1 ELSE 0 END) AS memory_enhanced_requests,
    SUM(CASE WHEN r.ensemble_config IS NOT NULL THEN 1 ELSE 0 END) AS ensemble_requests,
    r.request_type
FROM llm_requests r
WHERE r.created_at > NOW() - INTERVAL '7 days'
GROUP BY date_trunc('hour', r.created_at), r.request_type
WITH DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_req_analytics_hour_type
    ON mv_request_analytics_hourly (hour, request_type);
CREATE INDEX IF NOT EXISTS idx_mv_req_analytics_hour
    ON mv_request_analytics_hourly (hour DESC);

COMMENT ON MATERIALIZED VIEW mv_request_analytics_hourly
    IS 'Hourly request analytics for dashboards, refresh every 15 minutes';

-- -----------------------------------------------------------------------------
-- 2.4 mv_session_stats_daily
-- Refresh: Every hour
-- Purpose: Daily session statistics
-- -----------------------------------------------------------------------------
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_session_stats_daily AS
SELECT
    date_trunc('day', s.created_at) AS day,
    COUNT(*) AS total_sessions,
    COUNT(*) FILTER (WHERE s.status = 'active') AS active_sessions,
    COUNT(*) FILTER (WHERE s.status = 'expired') AS expired_sessions,
    AVG(s.request_count)::INTEGER AS avg_requests_per_session,
    MAX(s.request_count) AS max_requests_per_session,
    AVG(EXTRACT(EPOCH FROM (s.last_activity - s.created_at)))::INTEGER AS avg_session_duration_seconds,
    COUNT(DISTINCT s.user_id) AS unique_users
FROM user_sessions s
WHERE s.created_at > NOW() - INTERVAL '30 days'
GROUP BY date_trunc('day', s.created_at)
WITH DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_session_stats_day
    ON mv_session_stats_daily (day);

COMMENT ON MATERIALIZED VIEW mv_session_stats_daily
    IS 'Daily session statistics, refresh every hour';

-- -----------------------------------------------------------------------------
-- 2.5 mv_task_statistics
-- Refresh: Every 5 minutes
-- Purpose: Background task statistics by type and priority
-- -----------------------------------------------------------------------------
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_task_statistics AS
SELECT
    t.task_type,
    t.priority,
    COUNT(*) AS total_tasks,
    COUNT(*) FILTER (WHERE t.status = 'pending') AS pending_count,
    COUNT(*) FILTER (WHERE t.status = 'running') AS running_count,
    COUNT(*) FILTER (WHERE t.status = 'completed') AS completed_count,
    COUNT(*) FILTER (WHERE t.status = 'failed') AS failed_count,
    COUNT(*) FILTER (WHERE t.status = 'stuck') AS stuck_count,
    AVG(t.actual_duration_seconds)::INTEGER AS avg_duration_seconds,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY t.actual_duration_seconds)::INTEGER AS p95_duration_seconds,
    AVG(t.retry_count)::DECIMAL(3,1) AS avg_retries,
    MAX(t.created_at) AS last_created,
    MAX(t.completed_at) AS last_completed
FROM background_tasks t
WHERE t.created_at > NOW() - INTERVAL '24 hours'
    AND t.deleted_at IS NULL
GROUP BY t.task_type, t.priority
WITH DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_task_stats_type_priority
    ON mv_task_statistics (task_type, priority);

COMMENT ON MATERIALIZED VIEW mv_task_statistics
    IS 'Background task statistics by type and priority, refresh every 5 minutes';

-- -----------------------------------------------------------------------------
-- 2.6 mv_model_capabilities
-- Refresh: Every hour
-- Purpose: Model capabilities summary by provider
-- -----------------------------------------------------------------------------
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_model_capabilities AS
SELECT
    m.provider_name,
    COUNT(*) AS total_models,
    COUNT(*) FILTER (WHERE m.supports_streaming) AS streaming_models,
    COUNT(*) FILTER (WHERE m.supports_function_calling) AS function_calling_models,
    COUNT(*) FILTER (WHERE m.supports_vision) AS vision_models,
    COUNT(*) FILTER (WHERE m.supports_audio) AS audio_models,
    COUNT(*) FILTER (WHERE m.supports_reasoning) AS reasoning_models,
    AVG(m.benchmark_score)::DECIMAL(5,2) AS avg_benchmark_score,
    MAX(m.benchmark_score) AS max_benchmark_score,
    AVG(m.context_window)::INTEGER AS avg_context_window,
    MAX(m.context_window) AS max_context_window,
    AVG(m.pricing_input)::DECIMAL(10,6) AS avg_input_price,
    AVG(m.pricing_output)::DECIMAL(10,6) AS avg_output_price,
    MAX(m.last_refreshed_at) AS last_sync
FROM models_metadata m
GROUP BY m.provider_name
WITH DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_model_caps_provider
    ON mv_model_capabilities (provider_name);

COMMENT ON MATERIALIZED VIEW mv_model_capabilities
    IS 'Model capabilities summary by provider, refresh every hour';

-- -----------------------------------------------------------------------------
-- 2.7 mv_protocol_metrics_agg
-- Refresh: Every 5 minutes
-- Purpose: Protocol metrics aggregated by hour
-- -----------------------------------------------------------------------------
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_protocol_metrics_agg AS
SELECT
    m.protocol_type,
    m.operation,
    date_trunc('hour', m.created_at) AS hour,
    COUNT(*) AS total_calls,
    COUNT(*) FILTER (WHERE m.status = 'success') AS success_count,
    COUNT(*) FILTER (WHERE m.status = 'error') AS error_count,
    AVG(m.duration_ms)::INTEGER AS avg_duration_ms,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY m.duration_ms)::INTEGER AS p50_duration_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY m.duration_ms)::INTEGER AS p95_duration_ms,
    PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY m.duration_ms)::INTEGER AS p99_duration_ms,
    MIN(m.duration_ms) AS min_duration_ms,
    MAX(m.duration_ms) AS max_duration_ms
FROM protocol_metrics m
WHERE m.created_at > NOW() - INTERVAL '24 hours'
GROUP BY m.protocol_type, m.operation, date_trunc('hour', m.created_at)
WITH DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_protocol_agg_key
    ON mv_protocol_metrics_agg (protocol_type, operation, hour);
CREATE INDEX IF NOT EXISTS idx_mv_protocol_agg_hour
    ON mv_protocol_metrics_agg (hour DESC);

COMMENT ON MATERIALIZED VIEW mv_protocol_metrics_agg
    IS 'Protocol metrics aggregated by hour, refresh every 5 minutes';


-- =============================================================================
-- SECTION 3: REFRESH FUNCTIONS
-- =============================================================================

-- Refresh all materialized views with status reporting
CREATE OR REPLACE FUNCTION refresh_all_materialized_views()
RETURNS TABLE (
    view_name      TEXT,
    refresh_status TEXT,
    duration_ms    INTEGER
) AS $$
DECLARE
    start_time TIMESTAMP;
    end_time   TIMESTAMP;
    views      TEXT[] := ARRAY[
        'mv_provider_performance',
        'mv_mcp_server_health',
        'mv_request_analytics_hourly',
        'mv_session_stats_daily',
        'mv_task_statistics',
        'mv_model_capabilities',
        'mv_protocol_metrics_agg'
    ];
    v TEXT;
BEGIN
    FOREACH v IN ARRAY views
    LOOP
        start_time := clock_timestamp();
        BEGIN
            EXECUTE format('REFRESH MATERIALIZED VIEW CONCURRENTLY %I', v);
            end_time := clock_timestamp();
            view_name := v;
            refresh_status := 'success';
            duration_ms := EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
            RETURN NEXT;
        EXCEPTION WHEN OTHERS THEN
            end_time := clock_timestamp();
            view_name := v;
            refresh_status := 'error: ' || SQLERRM;
            duration_ms := EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
            RETURN NEXT;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Quick refresh of performance-critical views only
CREATE OR REPLACE FUNCTION refresh_critical_views()
RETURNS VOID AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_provider_performance;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_mcp_server_health;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_task_statistics;
END;
$$ LANGUAGE plpgsql;

-- Get statistics about all materialized views (row counts, sizes)
CREATE OR REPLACE FUNCTION get_materialized_view_stats()
RETURNS TABLE (
    view_name    TEXT,
    row_count    BIGINT,
    size_bytes   BIGINT,
    last_refresh TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        schemaname || '.' || matviewname AS view_name,
        (SELECT COUNT(*) FROM pg_class c WHERE c.relname = matviewname)::BIGINT AS row_count,
        pg_relation_size(matviewname::regclass) AS size_bytes,
        (SELECT MAX(pg_stat_get_last_analyze_time(c.oid))
         FROM pg_class c WHERE c.relname = matviewname) AS last_refresh
    FROM pg_matviews
    WHERE schemaname = 'public';
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION refresh_all_materialized_views IS 'Refresh all materialized views with status reporting';
COMMENT ON FUNCTION refresh_critical_views IS 'Quick refresh of performance-critical views';

-- =============================================================================
-- SECTION 4: SCHEDULED REFRESH (pg_cron)
-- =============================================================================
-- Requires pg_cron extension. Uncomment and run separately if available.
--
-- SELECT cron.schedule('refresh_critical_views', '*/5 * * * *', 'SELECT refresh_critical_views()');
-- SELECT cron.schedule('refresh_all_views', '0 * * * *', 'SELECT refresh_all_materialized_views()');
