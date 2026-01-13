-- Migration: 013_materialized_views
-- Description: Create materialized views for analytics and performance dashboards
-- Date: 2026-01-13

-- ============================================================================
-- Provider Performance Summary (Refresh: Every 5 minutes)
-- ============================================================================

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

-- ============================================================================
-- MCP Server Health Summary (Refresh: Every 1 minute)
-- ============================================================================

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

-- ============================================================================
-- Request Analytics Hourly (Refresh: Every 15 minutes)
-- ============================================================================

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

-- ============================================================================
-- Session Statistics Daily (Refresh: Every hour)
-- ============================================================================

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

-- ============================================================================
-- Background Task Statistics (Refresh: Every 5 minutes)
-- ============================================================================

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

-- ============================================================================
-- Model Capabilities Summary (Refresh: Every hour)
-- ============================================================================

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

-- ============================================================================
-- Protocol Metrics Aggregation (Refresh: Every 5 minutes)
-- ============================================================================

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

-- ============================================================================
-- Refresh Functions
-- ============================================================================

-- Function to refresh all materialized views
CREATE OR REPLACE FUNCTION refresh_all_materialized_views()
RETURNS TABLE (
    view_name TEXT,
    refresh_status TEXT,
    duration_ms INTEGER
) AS $$
DECLARE
    start_time TIMESTAMP;
    end_time TIMESTAMP;
    views TEXT[] := ARRAY[
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

-- Function to refresh performance-critical views (faster refresh)
CREATE OR REPLACE FUNCTION refresh_critical_views()
RETURNS VOID AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_provider_performance;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_mcp_server_health;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_task_statistics;
END;
$$ LANGUAGE plpgsql;

-- Function to get materialized view statistics
CREATE OR REPLACE FUNCTION get_materialized_view_stats()
RETURNS TABLE (
    view_name TEXT,
    row_count BIGINT,
    size_bytes BIGINT,
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

-- ============================================================================
-- Scheduled Refresh (using pg_cron if available)
-- ============================================================================

-- Note: These require pg_cron extension to be enabled
-- Uncomment and run separately if pg_cron is available

-- SELECT cron.schedule('refresh_critical_views', '*/5 * * * *', 'SELECT refresh_critical_views()');
-- SELECT cron.schedule('refresh_all_views', '0 * * * *', 'SELECT refresh_all_materialized_views()');

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON MATERIALIZED VIEW mv_provider_performance IS 'Provider performance metrics aggregated over 24 hours, refresh every 5 minutes';
COMMENT ON MATERIALIZED VIEW mv_mcp_server_health IS 'MCP server health and operation metrics, refresh every 1 minute';
COMMENT ON MATERIALIZED VIEW mv_request_analytics_hourly IS 'Hourly request analytics for dashboards, refresh every 15 minutes';
COMMENT ON MATERIALIZED VIEW mv_session_stats_daily IS 'Daily session statistics, refresh every hour';
COMMENT ON MATERIALIZED VIEW mv_task_statistics IS 'Background task statistics by type and priority, refresh every 5 minutes';
COMMENT ON MATERIALIZED VIEW mv_model_capabilities IS 'Model capabilities summary by provider, refresh every hour';
COMMENT ON MATERIALIZED VIEW mv_protocol_metrics_agg IS 'Protocol metrics aggregated by hour, refresh every 5 minutes';
COMMENT ON FUNCTION refresh_all_materialized_views IS 'Refresh all materialized views with status reporting';
COMMENT ON FUNCTION refresh_critical_views IS 'Quick refresh of performance-critical views';
