-- =============================================================================
-- HelixAgent SQL Schema: LLM Requests & Responses
-- =============================================================================
-- Domain: Request lifecycle tracking and provider response storage.
-- Source migrations: 001_initial_schema.sql
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: llm_requests
-- -----------------------------------------------------------------------------
-- Records every LLM completion request made through HelixAgent. Each request
-- belongs to a session and user, carries the prompt and message history, and
-- optionally specifies ensemble configuration for multi-provider orchestration.
-- The memory_enhanced flag indicates whether Cognee RAG memories were injected.
--
-- Primary Key: id (UUID, auto-generated)
-- Foreign Keys:
--   session_id -> user_sessions(id) ON DELETE CASCADE
--   user_id    -> users(id) ON DELETE CASCADE
-- Referenced by: llm_responses.request_id
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS llm_requests (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id      UUID         REFERENCES user_sessions(id) ON DELETE CASCADE,  -- Owning session
    user_id         UUID         REFERENCES users(id) ON DELETE CASCADE,           -- Requesting user
    prompt          TEXT         NOT NULL,                   -- Raw user prompt text
    messages        JSONB        NOT NULL DEFAULT '[]',      -- Full message history (OpenAI format)
    model_params    JSONB        NOT NULL DEFAULT '{}',      -- Model parameters (temperature, top_p, etc.)
    ensemble_config JSONB        DEFAULT NULL,               -- Ensemble strategy config (NULL = single provider)
    memory_enhanced BOOLEAN      DEFAULT FALSE,              -- Whether Cognee memories were injected
    memory          JSONB        DEFAULT '{}',               -- Injected memory data
    status          VARCHAR(50)  DEFAULT 'pending',          -- Lifecycle: 'pending', 'running', 'completed', 'failed'
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),  -- Request creation time
    started_at      TIMESTAMP WITH TIME ZONE,                -- When processing began
    completed_at    TIMESTAMP WITH TIME ZONE,                -- When processing finished
    request_type    VARCHAR(50)  DEFAULT 'completion'         -- Type: 'completion', 'chat', 'embedding', 'vision'
);

-- Basic indexes (migration 001)
CREATE INDEX IF NOT EXISTS idx_llm_requests_session_id ON llm_requests(session_id);
CREATE INDEX IF NOT EXISTS idx_llm_requests_user_id    ON llm_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_llm_requests_status     ON llm_requests(status);

-- Performance indexes (migration 012)
-- Composite index for session-scoped request lookup with status filtering
CREATE INDEX IF NOT EXISTS idx_requests_session_status
    ON llm_requests (session_id, status, created_at DESC)
    WHERE status IN ('pending', 'completed');

-- Partial index for recent requests (analytics hot path, last 24 hours)
CREATE INDEX IF NOT EXISTS idx_requests_recent
    ON llm_requests (created_at DESC)
    WHERE created_at > NOW() - INTERVAL '24 hours';

COMMENT ON TABLE llm_requests IS 'LLM request history';

-- -----------------------------------------------------------------------------
-- Table: llm_responses
-- -----------------------------------------------------------------------------
-- Stores individual provider responses for each request. In ensemble mode,
-- a single request generates multiple responses (one per provider). The
-- ensemble voting system marks the winning response via the 'selected' flag
-- and records the selection_score used to choose it.
--
-- Primary Key: id (UUID, auto-generated)
-- Foreign Keys:
--   request_id  -> llm_requests(id) ON DELETE CASCADE
--   provider_id -> llm_providers(id) ON DELETE SET NULL
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS llm_responses (
    id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id      UUID          REFERENCES llm_requests(id) ON DELETE CASCADE,   -- Parent request
    provider_id     UUID          REFERENCES llm_providers(id) ON DELETE SET NULL,  -- Source provider (NULL if provider deleted)
    provider_name   VARCHAR(100)  NOT NULL,                  -- Provider name (denormalized for query speed)
    content         TEXT          NOT NULL,                   -- Response content / completion text
    confidence      DECIMAL(3,2)  NOT NULL DEFAULT 0.0,       -- Provider self-reported confidence (0.00-1.00)
    tokens_used     INTEGER       DEFAULT 0,                  -- Total tokens consumed (input + output)
    response_time   BIGINT        DEFAULT 0,                  -- Response latency in milliseconds
    finish_reason   VARCHAR(50)   DEFAULT 'stop',             -- Completion reason: 'stop', 'length', 'content_filter', 'tool_calls'
    metadata        JSONB         DEFAULT '{}',                -- Provider-specific response metadata
    selected        BOOLEAN       DEFAULT FALSE,               -- TRUE if chosen by ensemble voting
    selection_score DECIMAL(5,2)  DEFAULT 0.0,                 -- Ensemble selection score
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Basic indexes (migration 001)
CREATE INDEX IF NOT EXISTS idx_llm_responses_request_id  ON llm_responses(request_id);
CREATE INDEX IF NOT EXISTS idx_llm_responses_provider_id ON llm_responses(provider_id);
CREATE INDEX IF NOT EXISTS idx_llm_responses_selected    ON llm_responses(selected);

-- Performance indexes (migration 012)
-- Composite index for response selection queries
CREATE INDEX IF NOT EXISTS idx_responses_selection
    ON llm_responses (request_id, selected, selection_score DESC);

-- Covering index for response aggregation (avoids table lookups)
CREATE INDEX IF NOT EXISTS idx_responses_aggregation
    ON llm_responses (provider_name, created_at DESC)
    INCLUDE (response_time, tokens_used, confidence);

COMMENT ON TABLE llm_responses IS 'Individual provider responses';
