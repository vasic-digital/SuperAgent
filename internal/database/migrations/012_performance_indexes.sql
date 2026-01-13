-- Migration: 012_performance_indexes
-- Description: Add performance indexes for hot query paths
-- Date: 2026-01-13

-- ============================================================================
-- Provider Performance Indexes
-- ============================================================================

-- Index for provider lookup by health status and enabled (hot path for routing)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_providers_healthy_enabled
    ON llm_providers (name, health_status, response_time)
    WHERE enabled = TRUE AND health_status = 'healthy';

-- Index for provider weight-based selection
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_providers_by_weight
    ON llm_providers (weight DESC, response_time ASC)
    WHERE enabled = TRUE;

-- Covering index for provider health checks
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_providers_health_check
    ON llm_providers (id, name, health_status, response_time, updated_at)
    WHERE enabled = TRUE;

-- ============================================================================
-- Request/Response Performance Indexes
-- ============================================================================

-- Composite index for request lookup by session with status filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_requests_session_status
    ON llm_requests (session_id, status, created_at DESC)
    WHERE status IN ('pending', 'completed');

-- Index for recent requests (hot path for analytics)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_requests_recent
    ON llm_requests (created_at DESC)
    WHERE created_at > NOW() - INTERVAL '24 hours';

-- Composite index for response selection
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_responses_selection
    ON llm_responses (request_id, selected, selection_score DESC);

-- Covering index for response aggregation
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_responses_aggregation
    ON llm_responses (provider_name, created_at DESC)
    INCLUDE (response_time, tokens_used, confidence);

-- ============================================================================
-- Session Performance Indexes
-- ============================================================================

-- Index for active sessions (hot path)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_active
    ON user_sessions (user_id, status, last_activity DESC)
    WHERE status = 'active';

-- Index for session expiration cleanup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_expired
    ON user_sessions (expires_at)
    WHERE status = 'active' AND expires_at < NOW();

-- ============================================================================
-- MCP/Protocol Server Indexes
-- ============================================================================

-- Index for active MCP servers
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_mcp_active_servers
    ON mcp_servers (type, enabled, last_sync DESC)
    WHERE enabled = TRUE;

-- Index for protocol metrics time-series queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_protocol_metrics_timeseries
    ON protocol_metrics (protocol_type, created_at DESC)
    INCLUDE (status, duration_ms);

-- Index for protocol metrics aggregation by server
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_protocol_metrics_by_server
    ON protocol_metrics (server_id, operation, status)
    WHERE server_id IS NOT NULL;

-- BRIN index for protocol metrics (efficient for time-series)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_protocol_metrics_brin
    ON protocol_metrics USING BRIN (created_at);

-- ============================================================================
-- Cache Indexes
-- ============================================================================

-- Index for cache expiration cleanup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cache_expiration
    ON protocol_cache (expires_at)
    WHERE expires_at < NOW();

-- ============================================================================
-- Model Metadata Performance Indexes
-- ============================================================================

-- Composite index for model lookup by capabilities
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_models_by_capabilities
    ON models_metadata (provider_name, supports_streaming, supports_function_calling)
    WHERE provider_name IS NOT NULL;

-- Index for model search with scoring
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_models_scored
    ON models_metadata (benchmark_score DESC, reliability_score DESC)
    WHERE benchmark_score IS NOT NULL;

-- GIN index for model tags (JSONB array queries)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_models_tags_gin
    ON models_metadata USING GIN (tags);

-- GIN index for protocol support queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_models_protocol_gin
    ON models_metadata USING GIN (protocol_support);

-- ============================================================================
-- Background Tasks Performance Indexes
-- ============================================================================

-- Index for priority queue ordering (critical for task dequeue)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_queue_order
    ON background_tasks (
        priority,
        scheduled_at ASC,
        created_at ASC
    )
    WHERE status = 'pending' AND deleted_at IS NULL;

-- Index for stuck task detection (join with resource snapshots)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_running_heartbeat
    ON background_tasks (last_heartbeat, started_at)
    WHERE status = 'running';

-- Index for task completion tracking
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_completion
    ON background_tasks (completed_at DESC, status)
    WHERE status IN ('completed', 'failed');

-- ============================================================================
-- Cognee Memory Indexes
-- ============================================================================

-- Full-text search index for memory content
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cognee_content_fts
    ON cognee_memories USING GIN (to_tsvector('english', content));

-- Index for memory retrieval by dataset
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cognee_dataset_recent
    ON cognee_memories (dataset_name, created_at DESC);

-- ============================================================================
-- Partial Indexes for Common Filters
-- ============================================================================

-- Partial index for pending webhook deliveries (retry queue)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_webhooks_pending_retry
    ON webhook_deliveries (created_at, attempts)
    WHERE status = 'pending' OR status = 'failed';

-- Partial index for dead letter reprocessing
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_dead_letter_ready
    ON background_tasks_dead_letter (reprocess_after ASC)
    WHERE NOT reprocessed AND reprocess_after IS NOT NULL;

-- ============================================================================
-- Statistics Targets (Improve Query Planner)
-- ============================================================================

-- Increase statistics for frequently queried columns
ALTER TABLE llm_providers ALTER COLUMN health_status SET STATISTICS 500;
ALTER TABLE llm_providers ALTER COLUMN enabled SET STATISTICS 500;
ALTER TABLE llm_requests ALTER COLUMN status SET STATISTICS 500;
ALTER TABLE llm_responses ALTER COLUMN provider_name SET STATISTICS 500;
ALTER TABLE background_tasks ALTER COLUMN status SET STATISTICS 500;
ALTER TABLE background_tasks ALTER COLUMN priority SET STATISTICS 500;
ALTER TABLE protocol_metrics ALTER COLUMN protocol_type SET STATISTICS 500;

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON INDEX idx_providers_healthy_enabled IS 'Hot path index for routing requests to healthy providers';
COMMENT ON INDEX idx_requests_recent IS 'Partial index for analytics queries on recent requests';
COMMENT ON INDEX idx_sessions_active IS 'Hot path index for active session lookups';
COMMENT ON INDEX idx_tasks_queue_order IS 'Priority queue index for task dequeue function';
COMMENT ON INDEX idx_cognee_content_fts IS 'Full-text search on memory content for semantic retrieval';
