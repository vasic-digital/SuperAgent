-- =============================================================================
-- HelixAgent: Complete SQL Schema Reference
-- =============================================================================
-- Consolidated from migrations: 001, 002, 003, 011, 012, 013, 014
-- Database: PostgreSQL 15+
-- Extensions: uuid-ossp, pgvector
--
-- This file is a single-file reference of the entire HelixAgent database
-- schema. It merges all tables, indexes, materialized views, stored functions,
-- and triggers into one executable script with comprehensive documentation.
--
-- Table of Contents:
--   1. Extensions
--   2. Custom Types (ENUMs)
--   3. Core Tables (users, sessions, providers, requests, responses, memories)
--   4. Model Catalog Tables (models_metadata, benchmarks, refresh_history)
--   5. Protocol Tables (mcp/lsp/acp servers, embedding, vectors, cache, metrics)
--   6. Background Task Tables (tasks, dead_letter, history, snapshots, webhooks)
--   7. AI Debate Tables (debate_logs)
--   8. Triggers & Functions
--   9. Performance Indexes
--  10. Materialized Views
--  11. Refresh Functions
-- =============================================================================


-- =============================================================================
-- 1. EXTENSIONS
-- =============================================================================

-- UUID generation functions (v4 random UUIDs)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Vector similarity search (pgvector for embedding operations)
CREATE EXTENSION IF NOT EXISTS pgvector;


-- =============================================================================
-- 2. CUSTOM TYPES (ENUMs)
-- =============================================================================

-- Task lifecycle states
DO $$ BEGIN
    CREATE TYPE task_status AS ENUM (
        'pending',      -- Task created, waiting in queue
        'queued',       -- Task assigned to worker pool
        'running',      -- Task currently executing
        'paused',       -- Task paused by user/system
        'completed',    -- Task finished successfully
        'failed',       -- Task failed with error
        'stuck',        -- Task detected as stuck (no heartbeat)
        'cancelled',    -- Task cancelled by user
        'dead_letter'   -- Task moved to dead-letter queue
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Task priority levels for queue ordering
DO $$ BEGIN
    CREATE TYPE task_priority AS ENUM (
        'critical',     -- P0: Execute immediately
        'high',         -- P1: Execute next
        'normal',       -- P2: Default priority
        'low',          -- P3: Execute when resources available
        'background'    -- P4: Execute during low load
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;


-- =============================================================================
-- 3. CORE TABLES (Migration 001)
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: users
-- -----------------------------------------------------------------------------
-- Registered user accounts. Each user has unique username, email, and API key.
-- The role column controls authorization ('user', 'admin').
--
-- PK: id (UUID)
-- Unique: username, email, api_key
-- Referenced by: user_sessions, llm_requests
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    username      VARCHAR(255) UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    api_key       VARCHAR(255) UNIQUE NOT NULL,
    role          VARCHAR(50)  DEFAULT 'user',
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

COMMENT ON TABLE users IS 'User accounts for HelixAgent';

-- -----------------------------------------------------------------------------
-- Table: user_sessions
-- -----------------------------------------------------------------------------
-- Active user sessions with conversation context. Sessions expire after a
-- configurable TTL and track request counts for rate limiting.
--
-- PK: id (UUID)
-- FK: user_id -> users(id) ON DELETE CASCADE
-- Unique: session_token
-- Referenced by: llm_requests, cognee_memories
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_sessions (
    id             UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID         REFERENCES users(id) ON DELETE CASCADE,
    session_token  VARCHAR(255) UNIQUE NOT NULL,
    context        JSONB        DEFAULT '{}',
    memory_id      UUID,                                    -- Logical ref to Cognee memory (external)
    status         VARCHAR(50)  DEFAULT 'active',           -- 'active', 'expired', 'revoked'
    request_count  INTEGER      DEFAULT 0,
    last_activity  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at     TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

COMMENT ON TABLE user_sessions IS 'Active user sessions with context';

-- -----------------------------------------------------------------------------
-- Table: llm_providers
-- -----------------------------------------------------------------------------
-- Registry of configured LLM providers. Weight controls ensemble voting
-- influence. Health status is updated by periodic probes.
--
-- PK: id (UUID)
-- Unique: name
-- Referenced by: llm_responses, models_metadata
-- Extended by migration 002 with Models.dev columns.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS llm_providers (
    id                    UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                  VARCHAR(255)  UNIQUE NOT NULL,
    type                  VARCHAR(100)  NOT NULL,
    api_key               VARCHAR(255),
    base_url              VARCHAR(500),
    model                 VARCHAR(255),
    weight                DECIMAL(5,2)  DEFAULT 1.0,
    enabled               BOOLEAN       DEFAULT TRUE,
    config                JSONB         DEFAULT '{}',
    health_status         VARCHAR(50)   DEFAULT 'unknown',   -- 'healthy', 'degraded', 'unhealthy', 'unknown'
    response_time         BIGINT        DEFAULT 0,           -- Last response time (ms)
    -- Models.dev integration (migration 002)
    modelsdev_provider_id VARCHAR(255),
    total_models          INTEGER       DEFAULT 0,
    enabled_models        INTEGER       DEFAULT 0,
    last_models_sync      TIMESTAMP WITH TIME ZONE,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

COMMENT ON TABLE llm_providers IS 'Configured LLM providers';

-- -----------------------------------------------------------------------------
-- Table: llm_requests
-- -----------------------------------------------------------------------------
-- Every LLM completion request. Carries prompt, message history, model params,
-- and optional ensemble configuration for multi-provider orchestration.
--
-- PK: id (UUID)
-- FK: session_id -> user_sessions(id) ON DELETE CASCADE
--     user_id    -> users(id) ON DELETE CASCADE
-- Referenced by: llm_responses
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS llm_requests (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id      UUID         REFERENCES user_sessions(id) ON DELETE CASCADE,
    user_id         UUID         REFERENCES users(id) ON DELETE CASCADE,
    prompt          TEXT         NOT NULL,
    messages        JSONB        NOT NULL DEFAULT '[]',       -- OpenAI-format message array
    model_params    JSONB        NOT NULL DEFAULT '{}',       -- temperature, top_p, max_tokens, etc.
    ensemble_config JSONB        DEFAULT NULL,                -- Ensemble strategy (NULL = single provider)
    memory_enhanced BOOLEAN      DEFAULT FALSE,               -- Whether Cognee memories were injected
    memory          JSONB        DEFAULT '{}',                -- Injected memory data
    status          VARCHAR(50)  DEFAULT 'pending',           -- 'pending', 'running', 'completed', 'failed'
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at      TIMESTAMP WITH TIME ZONE,
    completed_at    TIMESTAMP WITH TIME ZONE,
    request_type    VARCHAR(50)  DEFAULT 'completion'          -- 'completion', 'chat', 'embedding', 'vision'
);

COMMENT ON TABLE llm_requests IS 'LLM request history';

-- -----------------------------------------------------------------------------
-- Table: llm_responses
-- -----------------------------------------------------------------------------
-- Individual provider responses. In ensemble mode, one request generates
-- multiple responses. The 'selected' flag marks the ensemble winner.
--
-- PK: id (UUID)
-- FK: request_id  -> llm_requests(id) ON DELETE CASCADE
--     provider_id -> llm_providers(id) ON DELETE SET NULL
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS llm_responses (
    id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id      UUID          REFERENCES llm_requests(id) ON DELETE CASCADE,
    provider_id     UUID          REFERENCES llm_providers(id) ON DELETE SET NULL,
    provider_name   VARCHAR(100)  NOT NULL,                   -- Denormalized for query speed
    content         TEXT          NOT NULL,
    confidence      DECIMAL(3,2)  NOT NULL DEFAULT 0.0,        -- 0.00-1.00
    tokens_used     INTEGER       DEFAULT 0,
    response_time   BIGINT        DEFAULT 0,                   -- Milliseconds
    finish_reason   VARCHAR(50)   DEFAULT 'stop',              -- 'stop', 'length', 'content_filter', 'tool_calls'
    metadata        JSONB         DEFAULT '{}',
    selected        BOOLEAN       DEFAULT FALSE,               -- Ensemble winner flag
    selection_score DECIMAL(5,2)  DEFAULT 0.0,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

COMMENT ON TABLE llm_responses IS 'Individual provider responses';

-- -----------------------------------------------------------------------------
-- Table: cognee_memories
-- -----------------------------------------------------------------------------
-- Cognee RAG memory entries. Each entry belongs to a session and dataset,
-- with optional vector and knowledge graph references.
--
-- PK: id (UUID)
-- FK: session_id -> user_sessions(id) ON DELETE CASCADE
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS cognee_memories (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id    UUID         REFERENCES user_sessions(id) ON DELETE CASCADE,
    dataset_name  VARCHAR(255) NOT NULL,
    content_type  VARCHAR(50)  DEFAULT 'text',                -- 'text', 'code', 'structured'
    content       TEXT         NOT NULL,
    vector_id     VARCHAR(255),                                -- External vector store reference
    graph_nodes   JSONB        DEFAULT '{}',                   -- Knowledge graph node references
    search_key    VARCHAR(255),
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

COMMENT ON TABLE cognee_memories IS 'Cognee RAG memory storage';


-- =============================================================================
-- 4. MODEL CATALOG TABLES (Migration 002)
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: models_metadata
-- -----------------------------------------------------------------------------
-- Model catalog from Models.dev. Stores capabilities, pricing, benchmarks,
-- and protocol support. Extended by migration 003 with protocol columns.
--
-- PK: id (UUID)
-- Unique: model_id
-- FK: provider_id -> llm_providers(id) ON DELETE CASCADE
-- Referenced by: model_benchmarks
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS models_metadata (
    id                UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_id          VARCHAR(255)  UNIQUE NOT NULL,
    model_name        VARCHAR(255)  NOT NULL,
    provider_id       VARCHAR(255)  NOT NULL,
    provider_name     VARCHAR(255)  NOT NULL,
    -- Model details
    description       TEXT,
    context_window    INTEGER,
    max_tokens        INTEGER,
    pricing_input     DECIMAL(10,6),
    pricing_output    DECIMAL(10,6),
    pricing_currency  VARCHAR(10)   DEFAULT 'USD',
    -- Capabilities
    supports_vision            BOOLEAN DEFAULT FALSE,
    supports_function_calling  BOOLEAN DEFAULT FALSE,
    supports_streaming         BOOLEAN DEFAULT FALSE,
    supports_json_mode         BOOLEAN DEFAULT FALSE,
    supports_image_generation  BOOLEAN DEFAULT FALSE,
    supports_audio             BOOLEAN DEFAULT FALSE,
    supports_code_generation   BOOLEAN DEFAULT FALSE,
    supports_reasoning         BOOLEAN DEFAULT FALSE,
    -- Scores
    benchmark_score   DECIMAL(5,2),
    popularity_score  INTEGER,
    reliability_score DECIMAL(5,2),
    -- Classification
    model_type        VARCHAR(100),
    model_family      VARCHAR(100),
    version           VARCHAR(50),
    tags              JSONB         DEFAULT '[]',
    -- Models.dev specific
    modelsdev_url         TEXT,
    modelsdev_id          VARCHAR(255),
    modelsdev_api_version VARCHAR(50),
    -- Raw data
    raw_metadata      JSONB         DEFAULT '{}',
    -- Protocol support (migration 003)
    protocol_support   JSONB         DEFAULT '[]',
    mcp_server_id      VARCHAR(255),
    lsp_server_id      VARCHAR(255),
    acp_server_id      VARCHAR(255),
    embedding_provider VARCHAR(50)   DEFAULT 'pgvector',
    protocol_config    JSONB         DEFAULT '{}',
    protocol_last_sync TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- Timestamps
    last_refreshed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT fk_provider FOREIGN KEY (provider_id)
        REFERENCES llm_providers(id) ON DELETE CASCADE
);

COMMENT ON TABLE models_metadata IS 'Model catalog with capabilities, pricing, and protocol support';

-- -----------------------------------------------------------------------------
-- Table: model_benchmarks
-- -----------------------------------------------------------------------------
-- Individual benchmark results per model (MMLU, HumanEval, GSM8K, etc.).
--
-- PK: id (UUID)
-- FK: model_id -> models_metadata(model_id) ON DELETE CASCADE
-- Unique: (model_id, benchmark_name)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS model_benchmarks (
    id               UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_id         VARCHAR(255)  NOT NULL,
    benchmark_name   VARCHAR(255)  NOT NULL,
    benchmark_type   VARCHAR(100),
    score            DECIMAL(10,4),
    rank             INTEGER,
    normalized_score DECIMAL(5,2),
    benchmark_date   DATE,
    metadata         JSONB         DEFAULT '{}',
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT fk_model FOREIGN KEY (model_id)
        REFERENCES models_metadata(model_id) ON DELETE CASCADE,
    CONSTRAINT unique_model_benchmark UNIQUE (model_id, benchmark_name)
);

COMMENT ON TABLE model_benchmarks IS 'Individual benchmark results per model';

-- -----------------------------------------------------------------------------
-- Table: models_refresh_history
-- -----------------------------------------------------------------------------
-- Audit log of Models.dev catalog refresh operations.
--
-- PK: id (UUID)
-- Standalone table (no foreign keys).
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS models_refresh_history (
    id               UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    refresh_type     VARCHAR(50)  NOT NULL,
    status           VARCHAR(50)  NOT NULL,
    models_refreshed INTEGER      DEFAULT 0,
    models_failed    INTEGER      DEFAULT 0,
    error_message    TEXT,
    started_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at     TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    metadata         JSONB        DEFAULT '{}'
);

COMMENT ON TABLE models_refresh_history IS 'Audit log of Models.dev catalog refresh operations';


-- =============================================================================
-- 5. PROTOCOL TABLES (Migration 003)
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: mcp_servers
-- -----------------------------------------------------------------------------
-- MCP (Model Context Protocol) server configuration. Local servers are
-- spawned as subprocesses; remote servers are accessed via HTTP/SSE.
--
-- PK: id (VARCHAR)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mcp_servers (
    id         VARCHAR(255) PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    type       VARCHAR(20)  NOT NULL CHECK (type IN ('local', 'remote')),
    command    TEXT,                                          -- Launch command (local servers)
    url        TEXT,                                          -- Endpoint URL (remote servers)
    enabled    BOOLEAN      NOT NULL DEFAULT true,
    tools      JSONB        NOT NULL DEFAULT '[]',           -- Tool definitions array
    last_sync  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE mcp_servers IS 'MCP server configuration';

-- -----------------------------------------------------------------------------
-- Table: lsp_servers
-- -----------------------------------------------------------------------------
-- LSP (Language Server Protocol) server configuration.
--
-- PK: id (VARCHAR)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS lsp_servers (
    id           VARCHAR(255)  PRIMARY KEY,
    name         VARCHAR(255)  NOT NULL,
    language     VARCHAR(50)   NOT NULL,
    command      VARCHAR(500)  NOT NULL,
    enabled      BOOLEAN       NOT NULL DEFAULT true,
    workspace    VARCHAR(1000) DEFAULT '/workspace',
    capabilities JSONB         NOT NULL DEFAULT '[]',
    last_sync    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE lsp_servers IS 'LSP server configuration';

-- -----------------------------------------------------------------------------
-- Table: acp_servers
-- -----------------------------------------------------------------------------
-- ACP (Agent Communication Protocol) server configuration.
--
-- PK: id (VARCHAR)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS acp_servers (
    id         VARCHAR(255) PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    type       VARCHAR(20)  NOT NULL CHECK (type IN ('local', 'remote')),
    url        TEXT,
    enabled    BOOLEAN      NOT NULL DEFAULT true,
    tools      JSONB        NOT NULL DEFAULT '[]',
    last_sync  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE acp_servers IS 'ACP server configuration';

-- -----------------------------------------------------------------------------
-- Table: embedding_config
-- -----------------------------------------------------------------------------
-- Embedding provider configuration for vector operations.
--
-- PK: id (SERIAL)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS embedding_config (
    id           SERIAL       PRIMARY KEY,
    provider     VARCHAR(50)  NOT NULL DEFAULT 'pgvector',
    model        VARCHAR(100) NOT NULL DEFAULT 'text-embedding-ada-002',
    dimension    INTEGER      NOT NULL DEFAULT 1536,
    api_endpoint TEXT,
    api_key      TEXT,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE embedding_config IS 'Embedding provider configuration';

-- -----------------------------------------------------------------------------
-- Table: vector_documents
-- -----------------------------------------------------------------------------
-- Documents with vector embeddings for semantic search via pgvector.
--
-- PK: id (UUID, gen_random_uuid)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vector_documents (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title              TEXT         NOT NULL,
    content            TEXT         NOT NULL,
    metadata           JSONB        DEFAULT '{}',
    embedding_id       UUID,                                  -- External embedding store ref
    embedding          VECTOR(1536),                           -- Primary embedding (1536-dim)
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    embedding_provider VARCHAR(50)  DEFAULT 'pgvector',
    search_vector      VECTOR(1536)                            -- Secondary search embedding
);

COMMENT ON TABLE vector_documents IS 'Documents with vector embeddings for semantic search';

-- -----------------------------------------------------------------------------
-- Table: protocol_cache
-- -----------------------------------------------------------------------------
-- TTL-based cache for protocol operation responses.
--
-- PK: cache_key (VARCHAR)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS protocol_cache (
    cache_key  VARCHAR(255) PRIMARY KEY,
    cache_data JSONB        NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE protocol_cache IS 'TTL-based cache for protocol responses';

-- -----------------------------------------------------------------------------
-- Table: protocol_metrics
-- -----------------------------------------------------------------------------
-- Operational metrics for all protocol operations (MCP, LSP, ACP, Embedding).
--
-- PK: id (SERIAL)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS protocol_metrics (
    id             SERIAL       PRIMARY KEY,
    protocol_type  VARCHAR(20)  NOT NULL CHECK (protocol_type IN ('mcp', 'lsp', 'acp', 'embedding')),
    server_id      VARCHAR(255),                              -- Logical ref to *_servers.id
    operation      VARCHAR(100) NOT NULL,
    status         VARCHAR(20)  NOT NULL CHECK (status IN ('success', 'error', 'timeout')),
    duration_ms    INTEGER,
    error_message  TEXT,
    metadata       JSONB        DEFAULT '{}',
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE protocol_metrics IS 'Protocol operation metrics';


-- =============================================================================
-- 6. BACKGROUND TASK TABLES (Migration 011)
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: background_tasks
-- -----------------------------------------------------------------------------
-- PostgreSQL-backed task queue with priority scheduling, retry logic,
-- checkpoint/resume, resource requirements, and soft delete.
--
-- PK: id (UUID)
-- FK: parent_task_id -> background_tasks(id) ON DELETE SET NULL (self-ref)
-- Referenced by: task_execution_history, task_resource_snapshots, webhook_deliveries
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS background_tasks (
    id                  UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    -- Identification
    task_type           VARCHAR(100)  NOT NULL,
    task_name           VARCHAR(255)  NOT NULL,
    correlation_id      VARCHAR(255),
    parent_task_id      UUID          REFERENCES background_tasks(id) ON DELETE SET NULL,
    -- Configuration
    payload             JSONB         NOT NULL DEFAULT '{}',
    config              JSONB         NOT NULL DEFAULT '{}',
    priority            task_priority NOT NULL DEFAULT 'normal',
    -- State
    status              task_status   NOT NULL DEFAULT 'pending',
    progress            DECIMAL(5,2)  DEFAULT 0.0,
    progress_message    TEXT,
    checkpoint          JSONB,
    -- Retry
    max_retries         INTEGER       DEFAULT 3,
    retry_count         INTEGER       DEFAULT 0,
    retry_delay_seconds INTEGER       DEFAULT 60,
    last_error          TEXT,
    error_history       JSONB         DEFAULT '[]',
    -- Execution
    worker_id           VARCHAR(100),
    process_pid         INTEGER,
    started_at          TIMESTAMP WITH TIME ZONE,
    completed_at        TIMESTAMP WITH TIME ZONE,
    last_heartbeat      TIMESTAMP WITH TIME ZONE,
    deadline            TIMESTAMP WITH TIME ZONE,
    -- Resources
    required_cpu_cores          INTEGER DEFAULT 1,
    required_memory_mb          INTEGER DEFAULT 512,
    estimated_duration_seconds  INTEGER,
    actual_duration_seconds     INTEGER,
    -- Notifications
    notification_config JSONB         DEFAULT '{}',
    -- User association (logical refs, no FK)
    user_id             UUID,
    session_id          UUID,
    tags                JSONB         DEFAULT '[]',
    metadata            JSONB         DEFAULT '{}',
    -- Timestamps
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    scheduled_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    -- Soft delete
    deleted_at          TIMESTAMP WITH TIME ZONE
);

COMMENT ON TABLE background_tasks IS 'Main table for background task management with PostgreSQL-backed queue';

-- -----------------------------------------------------------------------------
-- Table: background_tasks_dead_letter
-- -----------------------------------------------------------------------------
-- Dead-letter queue for tasks that exhausted all retry attempts.
--
-- PK: id (UUID)
-- No FK (task data is copied for independence).
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS background_tasks_dead_letter (
    id               UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    original_task_id UUID          NOT NULL,
    task_data        JSONB         NOT NULL,
    failure_reason   TEXT          NOT NULL,
    failure_count    INTEGER       DEFAULT 1,
    moved_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    reprocess_after  TIMESTAMP WITH TIME ZONE,
    reprocessed      BOOLEAN       DEFAULT FALSE
);

COMMENT ON TABLE background_tasks_dead_letter IS 'Dead-letter queue for tasks that failed after max retries';

-- -----------------------------------------------------------------------------
-- Table: task_execution_history
-- -----------------------------------------------------------------------------
-- Immutable audit trail of task state transitions.
--
-- PK: id (UUID)
-- FK: task_id -> background_tasks(id) ON DELETE CASCADE
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS task_execution_history (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id     UUID         NOT NULL REFERENCES background_tasks(id) ON DELETE CASCADE,
    event_type  VARCHAR(50)  NOT NULL,
    event_data  JSONB        DEFAULT '{}',
    worker_id   VARCHAR(100),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

COMMENT ON TABLE task_execution_history IS 'Audit trail of task state changes and events';

-- -----------------------------------------------------------------------------
-- Table: task_resource_snapshots
-- -----------------------------------------------------------------------------
-- Time-series resource usage data for running tasks.
--
-- PK: id (UUID)
-- FK: task_id -> background_tasks(id) ON DELETE CASCADE
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS task_resource_snapshots (
    id               UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id          UUID          NOT NULL REFERENCES background_tasks(id) ON DELETE CASCADE,
    -- CPU
    cpu_percent      DECIMAL(5,2),
    cpu_user_time    DECIMAL(12,4),
    cpu_system_time  DECIMAL(12,4),
    -- Memory
    memory_rss_bytes BIGINT,
    memory_vms_bytes BIGINT,
    memory_percent   DECIMAL(5,2),
    -- I/O
    io_read_bytes    BIGINT,
    io_write_bytes   BIGINT,
    io_read_count    BIGINT,
    io_write_count   BIGINT,
    -- Network
    net_bytes_sent   BIGINT,
    net_bytes_recv   BIGINT,
    net_connections  INTEGER,
    -- File descriptors
    open_files       INTEGER,
    open_fds         INTEGER,
    -- Process
    process_state    VARCHAR(20),
    thread_count     INTEGER,
    sampled_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

COMMENT ON TABLE task_resource_snapshots IS 'Time-series resource usage data for monitoring and stuck detection';

-- -----------------------------------------------------------------------------
-- Table: webhook_deliveries
-- -----------------------------------------------------------------------------
-- Webhook notification delivery tracking with retry support.
--
-- PK: id (UUID)
-- FK: task_id -> background_tasks(id) ON DELETE SET NULL
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id         UUID          REFERENCES background_tasks(id) ON DELETE SET NULL,
    webhook_url     TEXT          NOT NULL,
    event_type      VARCHAR(50)  NOT NULL,
    payload         JSONB        NOT NULL,
    status          VARCHAR(20)  DEFAULT 'pending',
    attempts        INTEGER      DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    last_error      TEXT,
    response_code   INTEGER,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    delivered_at    TIMESTAMP WITH TIME ZONE
);

COMMENT ON TABLE webhook_deliveries IS 'Webhook notification delivery tracking with retry support';


-- =============================================================================
-- 7. AI DEBATE TABLES (Migration 014)
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: debate_logs
-- -----------------------------------------------------------------------------
-- Logs every participant action in every debate round. Append-only with
-- time-based retention via expires_at.
--
-- PK: id (SERIAL)
-- No FK (uses string-based identifiers for flexibility).
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS debate_logs (
    id                     SERIAL       PRIMARY KEY,
    debate_id              VARCHAR(255) NOT NULL,
    session_id             VARCHAR(255) NOT NULL,
    participant_id         INTEGER,
    participant_identifier VARCHAR(255),
    participant_name       VARCHAR(255),
    role                   VARCHAR(100),               -- 'proponent', 'opponent', 'moderator', 'synthesizer'
    provider               VARCHAR(100),               -- LLM provider name
    model                  VARCHAR(255),               -- Specific model used
    round                  INTEGER,                    -- Round number (1-based)
    action                 VARCHAR(100),               -- 'response', 'rebuttal', 'summary', 'vote', 'synthesis'
    response_time_ms       BIGINT,
    quality_score          DECIMAL(5,4),                -- 0.0000-1.0000
    tokens_used            INTEGER,
    content_length         INTEGER,
    error_message          TEXT,
    metadata               JSONB        DEFAULT '{}',
    created_at             TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at             TIMESTAMP WITH TIME ZONE     -- NULL = no expiration
);

COMMENT ON TABLE debate_logs IS 'Stores AI debate round logs and participant responses';


-- =============================================================================
-- 8. TRIGGERS & FUNCTIONS
-- =============================================================================

-- Auto-update updated_at on models_metadata
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_models_metadata_updated_at
    BEFORE UPDATE ON models_metadata
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Auto-update updated_at on background_tasks
CREATE OR REPLACE FUNCTION update_background_task_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_background_task_updated_at ON background_tasks;
CREATE TRIGGER trg_background_task_updated_at
    BEFORE UPDATE ON background_tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_background_task_updated_at();

-- Atomic task dequeue with priority ordering and resource filtering
CREATE OR REPLACE FUNCTION dequeue_background_task(
    p_worker_id     VARCHAR(100),
    p_max_cpu_cores INTEGER DEFAULT 0,
    p_max_memory_mb INTEGER DEFAULT 0
)
RETURNS TABLE (
    task_id   UUID,
    task_type VARCHAR(100),
    task_name VARCHAR(255),
    payload   JSONB,
    config    JSONB,
    priority  task_priority
) AS $$
DECLARE
    v_task_id UUID;
BEGIN
    UPDATE background_tasks
    SET status = 'running', worker_id = p_worker_id,
        started_at = NOW(), last_heartbeat = NOW()
    WHERE id = (
        SELECT bt.id FROM background_tasks bt
        WHERE bt.status = 'pending' AND bt.scheduled_at <= NOW() AND bt.deleted_at IS NULL
          AND (p_max_cpu_cores = 0 OR bt.required_cpu_cores <= p_max_cpu_cores)
          AND (p_max_memory_mb = 0 OR bt.required_memory_mb <= p_max_memory_mb)
        ORDER BY
            CASE bt.priority
                WHEN 'critical' THEN 0 WHEN 'high' THEN 1 WHEN 'normal' THEN 2
                WHEN 'low' THEN 3 WHEN 'background' THEN 4
            END,
            bt.created_at ASC
        LIMIT 1 FOR UPDATE SKIP LOCKED
    )
    RETURNING background_tasks.id INTO v_task_id;

    IF v_task_id IS NOT NULL THEN
        RETURN QUERY
        SELECT bt.id, bt.task_type, bt.task_name, bt.payload, bt.config, bt.priority
        FROM background_tasks bt WHERE bt.id = v_task_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Stuck task detection
CREATE OR REPLACE FUNCTION get_stale_tasks(
    p_heartbeat_threshold INTERVAL DEFAULT INTERVAL '5 minutes'
)
RETURNS TABLE (
    task_id UUID, task_type VARCHAR(100), worker_id VARCHAR(100),
    started_at TIMESTAMP WITH TIME ZONE, last_heartbeat TIMESTAMP WITH TIME ZONE,
    time_since_heartbeat INTERVAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT bt.id, bt.task_type, bt.worker_id, bt.started_at, bt.last_heartbeat,
           NOW() - bt.last_heartbeat
    FROM background_tasks bt
    WHERE bt.status = 'running' AND bt.last_heartbeat < NOW() - p_heartbeat_threshold
      AND bt.deleted_at IS NULL;
END;
$$ LANGUAGE plpgsql;

-- Expired debate log cleanup
CREATE OR REPLACE FUNCTION cleanup_expired_debate_logs()
RETURNS INTEGER AS $$
DECLARE deleted_count INTEGER;
BEGIN
    DELETE FROM debate_logs WHERE expires_at IS NOT NULL AND expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;


-- =============================================================================
-- 9. INDEXES (Migrations 001, 002, 003, 011, 012, 014)
-- =============================================================================

-- ---- Core table indexes (001) ----
CREATE INDEX IF NOT EXISTS idx_users_email                  ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_api_key                ON users(api_key);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id        ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at     ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_session_token  ON user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_llm_providers_name           ON llm_providers(name);
CREATE INDEX IF NOT EXISTS idx_llm_providers_enabled        ON llm_providers(enabled);
CREATE INDEX IF NOT EXISTS idx_llm_requests_session_id      ON llm_requests(session_id);
CREATE INDEX IF NOT EXISTS idx_llm_requests_user_id         ON llm_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_llm_requests_status          ON llm_requests(status);
CREATE INDEX IF NOT EXISTS idx_llm_responses_request_id     ON llm_responses(request_id);
CREATE INDEX IF NOT EXISTS idx_llm_responses_provider_id    ON llm_responses(provider_id);
CREATE INDEX IF NOT EXISTS idx_llm_responses_selected       ON llm_responses(selected);
CREATE INDEX IF NOT EXISTS idx_cognee_memories_session_id   ON cognee_memories(session_id);
CREATE INDEX IF NOT EXISTS idx_cognee_memories_dataset_name ON cognee_memories(dataset_name);
CREATE INDEX IF NOT EXISTS idx_cognee_memories_search_key   ON cognee_memories(search_key);

-- ---- Model catalog indexes (002) ----
CREATE INDEX IF NOT EXISTS idx_models_metadata_provider_id     ON models_metadata(provider_id);
CREATE INDEX IF NOT EXISTS idx_models_metadata_model_type      ON models_metadata(model_type);
CREATE INDEX IF NOT EXISTS idx_models_metadata_tags            ON models_metadata USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_models_metadata_last_refreshed  ON models_metadata(last_refreshed_at);
CREATE INDEX IF NOT EXISTS idx_models_metadata_model_family    ON models_metadata(model_family);
CREATE INDEX IF NOT EXISTS idx_models_metadata_benchmark_score ON models_metadata(benchmark_score);
CREATE INDEX IF NOT EXISTS idx_benchmarks_model_id             ON model_benchmarks(model_id);
CREATE INDEX IF NOT EXISTS idx_benchmarks_type                 ON model_benchmarks(benchmark_type);
CREATE INDEX IF NOT EXISTS idx_benchmarks_score                ON model_benchmarks(score);
CREATE INDEX IF NOT EXISTS idx_refresh_history_started         ON models_refresh_history(started_at);
CREATE INDEX IF NOT EXISTS idx_refresh_history_status          ON models_refresh_history(status);

-- ---- Protocol indexes (003) ----
CREATE INDEX IF NOT EXISTS idx_models_metadata_protocol_support ON models_metadata USING GIN(protocol_support);
CREATE INDEX IF NOT EXISTS idx_mcp_servers_enabled              ON mcp_servers USING btree(enabled);
CREATE INDEX IF NOT EXISTS idx_lsp_servers_enabled              ON lsp_servers USING btree(enabled);
CREATE INDEX IF NOT EXISTS idx_acp_servers_enabled              ON acp_servers USING btree(enabled);
CREATE INDEX IF NOT EXISTS idx_vector_documents_embedding_provider ON vector_documents USING btree(embedding_provider);
CREATE INDEX IF NOT EXISTS idx_protocol_cache_expires_at        ON protocol_cache USING btree(expires_at);
CREATE INDEX IF NOT EXISTS idx_protocol_metrics_protocol_type   ON protocol_metrics USING btree(protocol_type);
CREATE INDEX IF NOT EXISTS idx_protocol_metrics_created_at      ON protocol_metrics USING btree(created_at);

-- ---- Background task indexes (011) ----
CREATE INDEX IF NOT EXISTS idx_tasks_status             ON background_tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority_status     ON background_tasks(priority, status, scheduled_at);
CREATE INDEX IF NOT EXISTS idx_tasks_worker              ON background_tasks(worker_id) WHERE status = 'running';
CREATE INDEX IF NOT EXISTS idx_tasks_user                ON background_tasks(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_correlation         ON background_tasks(correlation_id) WHERE correlation_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled           ON background_tasks(scheduled_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_tasks_heartbeat           ON background_tasks(last_heartbeat) WHERE status = 'running';
CREATE INDEX IF NOT EXISTS idx_tasks_deadline            ON background_tasks(deadline) WHERE deadline IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_type                ON background_tasks(task_type);
CREATE INDEX IF NOT EXISTS idx_tasks_created             ON background_tasks(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tasks_not_deleted         ON background_tasks(id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_task_history_task_id       ON task_execution_history(task_id);
CREATE INDEX IF NOT EXISTS idx_task_history_event_type    ON task_execution_history(event_type);
CREATE INDEX IF NOT EXISTS idx_task_history_created       ON task_execution_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_resource_snapshots_task    ON task_resource_snapshots(task_id, sampled_at DESC);
CREATE INDEX IF NOT EXISTS idx_resource_snapshots_recent  ON task_resource_snapshots(sampled_at DESC);
CREATE INDEX IF NOT EXISTS idx_dead_letter_reprocess      ON background_tasks_dead_letter(reprocess_after) WHERE NOT reprocessed;
CREATE INDEX IF NOT EXISTS idx_dead_letter_original       ON background_tasks_dead_letter(original_task_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_task    ON webhook_deliveries(task_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status  ON webhook_deliveries(status) WHERE status != 'delivered';

-- ---- Performance indexes (012) ----
CREATE INDEX IF NOT EXISTS idx_providers_healthy_enabled     ON llm_providers(name, health_status, response_time) WHERE enabled = TRUE AND health_status = 'healthy';
CREATE INDEX IF NOT EXISTS idx_providers_by_weight           ON llm_providers(weight DESC, response_time ASC) WHERE enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_providers_health_check        ON llm_providers(id, name, health_status, response_time, updated_at) WHERE enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_requests_session_status       ON llm_requests(session_id, status, created_at DESC) WHERE status IN ('pending', 'completed');
CREATE INDEX IF NOT EXISTS idx_requests_recent               ON llm_requests(created_at DESC) WHERE created_at > NOW() - INTERVAL '24 hours';
CREATE INDEX IF NOT EXISTS idx_responses_selection           ON llm_responses(request_id, selected, selection_score DESC);
CREATE INDEX IF NOT EXISTS idx_responses_aggregation         ON llm_responses(provider_name, created_at DESC) INCLUDE (response_time, tokens_used, confidence);
CREATE INDEX IF NOT EXISTS idx_sessions_active               ON user_sessions(user_id, status, last_activity DESC) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_sessions_expired              ON user_sessions(expires_at) WHERE status = 'active' AND expires_at < NOW();
CREATE INDEX IF NOT EXISTS idx_mcp_active_servers            ON mcp_servers(type, enabled, last_sync DESC) WHERE enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_protocol_metrics_timeseries   ON protocol_metrics(protocol_type, created_at DESC) INCLUDE (status, duration_ms);
CREATE INDEX IF NOT EXISTS idx_protocol_metrics_by_server    ON protocol_metrics(server_id, operation, status) WHERE server_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_protocol_metrics_brin         ON protocol_metrics USING BRIN(created_at);
CREATE INDEX IF NOT EXISTS idx_cache_expiration              ON protocol_cache(expires_at) WHERE expires_at < NOW();
CREATE INDEX IF NOT EXISTS idx_models_by_capabilities        ON models_metadata(provider_name, supports_streaming, supports_function_calling) WHERE provider_name IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_models_scored                 ON models_metadata(benchmark_score DESC, reliability_score DESC) WHERE benchmark_score IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_models_tags_gin               ON models_metadata USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_models_protocol_gin           ON models_metadata USING GIN(protocol_support);
CREATE INDEX IF NOT EXISTS idx_tasks_queue_order             ON background_tasks(priority, scheduled_at ASC, created_at ASC) WHERE status = 'pending' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_running_heartbeat       ON background_tasks(last_heartbeat, started_at) WHERE status = 'running';
CREATE INDEX IF NOT EXISTS idx_tasks_completion              ON background_tasks(completed_at DESC, status) WHERE status IN ('completed', 'failed');
CREATE INDEX IF NOT EXISTS idx_cognee_content_fts            ON cognee_memories USING GIN(to_tsvector('english', content));
CREATE INDEX IF NOT EXISTS idx_cognee_dataset_recent         ON cognee_memories(dataset_name, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_webhooks_pending_retry        ON webhook_deliveries(created_at, attempts) WHERE status = 'pending' OR status = 'failed';
CREATE INDEX IF NOT EXISTS idx_dead_letter_ready             ON background_tasks_dead_letter(reprocess_after ASC) WHERE NOT reprocessed AND reprocess_after IS NOT NULL;

-- ---- Debate indexes (014) ----
CREATE INDEX IF NOT EXISTS idx_debate_logs_debate_id       ON debate_logs(debate_id);
CREATE INDEX IF NOT EXISTS idx_debate_logs_session_id      ON debate_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_debate_logs_provider        ON debate_logs(provider);
CREATE INDEX IF NOT EXISTS idx_debate_logs_model           ON debate_logs(model);
CREATE INDEX IF NOT EXISTS idx_debate_logs_created_at      ON debate_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_debate_logs_expires_at      ON debate_logs(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_debate_logs_debate_round    ON debate_logs(debate_id, round);
CREATE INDEX IF NOT EXISTS idx_debate_logs_active          ON debate_logs(debate_id) WHERE expires_at IS NULL OR expires_at > NOW();
CREATE INDEX IF NOT EXISTS idx_debate_logs_provider_model  ON debate_logs(provider, model);
CREATE INDEX IF NOT EXISTS idx_debate_logs_metadata        ON debate_logs USING GIN(metadata);

-- ---- Statistics targets (012) ----
ALTER TABLE llm_providers    ALTER COLUMN health_status  SET STATISTICS 500;
ALTER TABLE llm_providers    ALTER COLUMN enabled        SET STATISTICS 500;
ALTER TABLE llm_requests     ALTER COLUMN status         SET STATISTICS 500;
ALTER TABLE llm_responses    ALTER COLUMN provider_name  SET STATISTICS 500;
ALTER TABLE background_tasks ALTER COLUMN status         SET STATISTICS 500;
ALTER TABLE background_tasks ALTER COLUMN priority       SET STATISTICS 500;
ALTER TABLE protocol_metrics ALTER COLUMN protocol_type  SET STATISTICS 500;


-- =============================================================================
-- 10. MATERIALIZED VIEWS (Migration 013)
-- =============================================================================

-- Provider performance (24h window, refresh every 5 min)
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_provider_performance AS
SELECT
    p.name AS provider_name, p.type AS provider_type, p.health_status, p.weight,
    p.response_time AS last_response_time_ms,
    COUNT(DISTINCT r.id) AS total_requests_24h,
    COUNT(DISTINCT r.id) FILTER (WHERE resp.selected = TRUE) AS selected_responses_24h,
    AVG(resp.response_time)::INTEGER AS avg_response_time_ms,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY resp.response_time)::INTEGER AS p50_response_time_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY resp.response_time)::INTEGER AS p95_response_time_ms,
    PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY resp.response_time)::INTEGER AS p99_response_time_ms,
    AVG(resp.confidence)::DECIMAL(3,2) AS avg_confidence,
    SUM(resp.tokens_used) AS total_tokens_24h,
    COUNT(*) FILTER (WHERE resp.finish_reason = 'stop')::FLOAT / NULLIF(COUNT(*), 0) AS success_rate,
    MAX(resp.created_at) AS last_response_at,
    p.updated_at AS provider_updated_at
FROM llm_providers p
LEFT JOIN llm_responses resp ON resp.provider_name = p.name AND resp.created_at > NOW() - INTERVAL '24 hours'
LEFT JOIN llm_requests r ON r.id = resp.request_id
WHERE p.enabled = TRUE
GROUP BY p.id, p.name, p.type, p.health_status, p.weight, p.response_time, p.updated_at
WITH DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_provider_perf_name ON mv_provider_performance(provider_name);
CREATE INDEX IF NOT EXISTS idx_mv_provider_perf_score ON mv_provider_performance(avg_confidence DESC, success_rate DESC);

-- MCP server health (1h window, refresh every 1 min)
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_mcp_server_health AS
SELECT
    s.id AS server_id, s.name AS server_name, s.type AS server_type, s.enabled,
    COUNT(m.id) AS total_operations_1h,
    COUNT(m.id) FILTER (WHERE m.status = 'success') AS successful_operations,
    COUNT(m.id) FILTER (WHERE m.status = 'error') AS failed_operations,
    AVG(m.duration_ms)::INTEGER AS avg_duration_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY m.duration_ms)::INTEGER AS p95_duration_ms,
    MAX(m.created_at) AS last_operation_at,
    COUNT(m.id) FILTER (WHERE m.status = 'success')::FLOAT / NULLIF(COUNT(m.id), 0) AS success_rate,
    s.last_sync, jsonb_array_length(s.tools) AS tool_count
FROM mcp_servers s
LEFT JOIN protocol_metrics m ON m.server_id = s.id AND m.protocol_type = 'mcp' AND m.created_at > NOW() - INTERVAL '1 hour'
GROUP BY s.id, s.name, s.type, s.enabled, s.last_sync, s.tools
WITH DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_mcp_health_id ON mv_mcp_server_health(server_id);
CREATE INDEX IF NOT EXISTS idx_mv_mcp_health_success ON mv_mcp_server_health(success_rate DESC) WHERE enabled = TRUE;

-- Request analytics hourly (7-day window, refresh every 15 min)
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_req_analytics_hour_type ON mv_request_analytics_hourly(hour, request_type);
CREATE INDEX IF NOT EXISTS idx_mv_req_analytics_hour ON mv_request_analytics_hourly(hour DESC);

-- Session stats daily (30-day window, refresh every hour)
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_session_stats_day ON mv_session_stats_daily(day);

-- Task statistics (24h window, refresh every 5 min)
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_task_statistics AS
SELECT
    t.task_type, t.priority,
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
WHERE t.created_at > NOW() - INTERVAL '24 hours' AND t.deleted_at IS NULL
GROUP BY t.task_type, t.priority
WITH DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_task_stats_type_priority ON mv_task_statistics(task_type, priority);

-- Model capabilities (refresh every hour)
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_model_caps_provider ON mv_model_capabilities(provider_name);

-- Protocol metrics aggregation (24h window, refresh every 5 min)
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_protocol_metrics_agg AS
SELECT
    m.protocol_type, m.operation,
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_protocol_agg_key ON mv_protocol_metrics_agg(protocol_type, operation, hour);
CREATE INDEX IF NOT EXISTS idx_mv_protocol_agg_hour ON mv_protocol_metrics_agg(hour DESC);

-- View comments
COMMENT ON MATERIALIZED VIEW mv_provider_performance IS 'Provider performance metrics (24h), refresh every 5 min';
COMMENT ON MATERIALIZED VIEW mv_mcp_server_health IS 'MCP server health (1h), refresh every 1 min';
COMMENT ON MATERIALIZED VIEW mv_request_analytics_hourly IS 'Hourly request analytics (7d), refresh every 15 min';
COMMENT ON MATERIALIZED VIEW mv_session_stats_daily IS 'Daily session statistics (30d), refresh every hour';
COMMENT ON MATERIALIZED VIEW mv_task_statistics IS 'Task statistics by type/priority (24h), refresh every 5 min';
COMMENT ON MATERIALIZED VIEW mv_model_capabilities IS 'Model capabilities by provider, refresh every hour';
COMMENT ON MATERIALIZED VIEW mv_protocol_metrics_agg IS 'Protocol metrics by hour (24h), refresh every 5 min';


-- =============================================================================
-- 11. REFRESH FUNCTIONS (Migration 013)
-- =============================================================================

-- Refresh all materialized views with status reporting
CREATE OR REPLACE FUNCTION refresh_all_materialized_views()
RETURNS TABLE (view_name TEXT, refresh_status TEXT, duration_ms INTEGER) AS $$
DECLARE
    start_time TIMESTAMP; end_time TIMESTAMP;
    views TEXT[] := ARRAY[
        'mv_provider_performance', 'mv_mcp_server_health',
        'mv_request_analytics_hourly', 'mv_session_stats_daily',
        'mv_task_statistics', 'mv_model_capabilities', 'mv_protocol_metrics_agg'
    ];
    v TEXT;
BEGIN
    FOREACH v IN ARRAY views LOOP
        start_time := clock_timestamp();
        BEGIN
            EXECUTE format('REFRESH MATERIALIZED VIEW CONCURRENTLY %I', v);
            end_time := clock_timestamp();
            view_name := v; refresh_status := 'success';
            duration_ms := EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
            RETURN NEXT;
        EXCEPTION WHEN OTHERS THEN
            end_time := clock_timestamp();
            view_name := v; refresh_status := 'error: ' || SQLERRM;
            duration_ms := EXTRACT(MILLISECONDS FROM (end_time - start_time))::INTEGER;
            RETURN NEXT;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Quick refresh of performance-critical views
CREATE OR REPLACE FUNCTION refresh_critical_views()
RETURNS VOID AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_provider_performance;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_mcp_server_health;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_task_statistics;
END;
$$ LANGUAGE plpgsql;

-- Get materialized view statistics
CREATE OR REPLACE FUNCTION get_materialized_view_stats()
RETURNS TABLE (view_name TEXT, row_count BIGINT, size_bytes BIGINT, last_refresh TIMESTAMP WITH TIME ZONE) AS $$
BEGIN
    RETURN QUERY
    SELECT schemaname || '.' || matviewname,
           (SELECT COUNT(*) FROM pg_class c WHERE c.relname = matviewname)::BIGINT,
           pg_relation_size(matviewname::regclass),
           (SELECT MAX(pg_stat_get_last_analyze_time(c.oid)) FROM pg_class c WHERE c.relname = matviewname)
    FROM pg_matviews WHERE schemaname = 'public';
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION refresh_all_materialized_views IS 'Refresh all materialized views with status reporting';
COMMENT ON FUNCTION refresh_critical_views IS 'Quick refresh of performance-critical views';
