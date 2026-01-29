-- =============================================================================
-- HelixAgent SQL Schema: LLM Providers & Model Metadata
-- =============================================================================
-- Domain: Provider registration, model catalog, and benchmark tracking.
-- Source migrations: 001_initial_schema.sql, 002_modelsdev_integration.sql
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: llm_providers
-- -----------------------------------------------------------------------------
-- Registry of configured LLM providers (Claude, DeepSeek, Gemini, Mistral,
-- OpenRouter, Qwen, ZAI, Zen, Cerebras, Ollama). Each provider has a weight
-- used in ensemble voting, a health status updated by periodic probes, and
-- an optional Models.dev integration for catalog synchronization.
--
-- Primary Key: id (UUID, auto-generated)
-- Unique Constraints: name
-- Referenced by: llm_responses.provider_id, models_metadata.provider_id (via FK)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS llm_providers (
    id                    UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                  VARCHAR(255)  UNIQUE NOT NULL,     -- Provider identifier (e.g., 'claude', 'deepseek')
    type                  VARCHAR(100)  NOT NULL,             -- Provider type / SDK type
    api_key               VARCHAR(255),                       -- API key (NULL for OAuth/free providers)
    base_url              VARCHAR(500),                       -- Custom API endpoint URL
    model                 VARCHAR(255),                       -- Default model for this provider
    weight                DECIMAL(5,2)  DEFAULT 1.0,          -- Ensemble voting weight (higher = more influence)
    enabled               BOOLEAN       DEFAULT TRUE,         -- Whether this provider is active
    config                JSONB         DEFAULT '{}',          -- Provider-specific configuration (models list, etc.)
    health_status         VARCHAR(50)   DEFAULT 'unknown',    -- Current health: 'healthy', 'degraded', 'unhealthy', 'unknown'
    response_time         BIGINT        DEFAULT 0,            -- Last measured response time in milliseconds

    -- Models.dev integration columns (added in migration 002)
    modelsdev_provider_id VARCHAR(255),                       -- Provider ID on Models.dev catalog
    total_models          INTEGER       DEFAULT 0,            -- Total models available from this provider
    enabled_models        INTEGER       DEFAULT 0,            -- Number of models currently enabled
    last_models_sync      TIMESTAMP WITH TIME ZONE,           -- Last successful Models.dev sync timestamp

    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Basic indexes (migration 001)
CREATE INDEX IF NOT EXISTS idx_llm_providers_name    ON llm_providers(name);
CREATE INDEX IF NOT EXISTS idx_llm_providers_enabled ON llm_providers(enabled);

-- Performance indexes (migration 012)
-- Hot path: routing requests to healthy, enabled providers
CREATE INDEX IF NOT EXISTS idx_providers_healthy_enabled
    ON llm_providers (name, health_status, response_time)
    WHERE enabled = TRUE AND health_status = 'healthy';

-- Weight-based provider selection
CREATE INDEX IF NOT EXISTS idx_providers_by_weight
    ON llm_providers (weight DESC, response_time ASC)
    WHERE enabled = TRUE;

-- Covering index for health-check dashboard queries
CREATE INDEX IF NOT EXISTS idx_providers_health_check
    ON llm_providers (id, name, health_status, response_time, updated_at)
    WHERE enabled = TRUE;

COMMENT ON TABLE llm_providers IS 'Configured LLM providers';

-- -----------------------------------------------------------------------------
-- Table: models_metadata
-- -----------------------------------------------------------------------------
-- Catalog of individual models sourced from Models.dev. Stores capabilities,
-- pricing, benchmark scores, and protocol support (MCP, LSP, ACP, Embeddings).
-- Each model belongs to a provider via provider_id foreign key.
--
-- Primary Key: id (UUID, auto-generated)
-- Unique Constraints: model_id
-- Foreign Keys: provider_id -> llm_providers(id) ON DELETE CASCADE
--               mcp_server_id, lsp_server_id, acp_server_id (logical refs, no FK)
-- Referenced by: model_benchmarks.model_id
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS models_metadata (
    id                UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_id          VARCHAR(255)  UNIQUE NOT NULL,          -- Canonical model identifier
    model_name        VARCHAR(255)  NOT NULL,                  -- Human-readable model name
    provider_id       VARCHAR(255)  NOT NULL,                  -- FK to llm_providers (by id or name)
    provider_name     VARCHAR(255)  NOT NULL,                  -- Provider display name

    -- Model details
    description       TEXT,                                     -- Model description
    context_window    INTEGER,                                  -- Maximum context window (tokens)
    max_tokens        INTEGER,                                  -- Maximum output tokens
    pricing_input     DECIMAL(10,6),                             -- Cost per input token (USD)
    pricing_output    DECIMAL(10,6),                             -- Cost per output token (USD)
    pricing_currency  VARCHAR(10)   DEFAULT 'USD',              -- Pricing currency code

    -- Capability booleans
    supports_vision            BOOLEAN DEFAULT FALSE,           -- Can process images
    supports_function_calling  BOOLEAN DEFAULT FALSE,           -- Supports tool/function calling
    supports_streaming         BOOLEAN DEFAULT FALSE,           -- Supports streaming responses
    supports_json_mode         BOOLEAN DEFAULT FALSE,           -- Supports forced JSON output
    supports_image_generation  BOOLEAN DEFAULT FALSE,           -- Can generate images
    supports_audio             BOOLEAN DEFAULT FALSE,           -- Can process audio
    supports_code_generation   BOOLEAN DEFAULT FALSE,           -- Specialized for code
    supports_reasoning         BOOLEAN DEFAULT FALSE,           -- Chain-of-thought / reasoning model

    -- Performance and quality scores
    benchmark_score   DECIMAL(5,2),                              -- Aggregate benchmark score
    popularity_score  INTEGER,                                   -- Popularity/usage ranking
    reliability_score DECIMAL(5,2),                              -- Reliability score (uptime, error rate)

    -- Classification
    model_type        VARCHAR(100),                              -- e.g., 'chat', 'completion', 'embedding'
    model_family      VARCHAR(100),                              -- e.g., 'claude-3', 'gpt-4', 'gemini-pro'
    version           VARCHAR(50),                               -- Model version string
    tags              JSONB         DEFAULT '[]',                -- Searchable tags array

    -- Models.dev specific fields
    modelsdev_url         TEXT,                                   -- Direct URL on Models.dev
    modelsdev_id          VARCHAR(255),                           -- Models.dev internal ID
    modelsdev_api_version VARCHAR(50),                            -- API version on Models.dev

    -- Raw data
    raw_metadata      JSONB         DEFAULT '{}',                -- Full unprocessed metadata from source

    -- Protocol support columns (added in migration 003)
    protocol_support   JSONB         DEFAULT '[]',               -- Array of supported protocols
    mcp_server_id      VARCHAR(255),                              -- Associated MCP server ID
    lsp_server_id      VARCHAR(255),                              -- Associated LSP server ID
    acp_server_id      VARCHAR(255),                              -- Associated ACP server ID
    embedding_provider VARCHAR(50)   DEFAULT 'pgvector',         -- Embedding provider for this model
    protocol_config    JSONB         DEFAULT '{}',                -- Protocol-specific configuration
    protocol_last_sync TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- Last protocol sync

    -- Timestamps
    last_refreshed_at TIMESTAMP WITH TIME ZONE NOT NULL,         -- Last catalog refresh
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT fk_provider FOREIGN KEY (provider_id)
        REFERENCES llm_providers(id) ON DELETE CASCADE
);

-- Auto-update trigger for updated_at
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

-- Basic indexes (migration 002)
CREATE INDEX IF NOT EXISTS idx_models_metadata_provider_id      ON models_metadata(provider_id);
CREATE INDEX IF NOT EXISTS idx_models_metadata_model_type       ON models_metadata(model_type);
CREATE INDEX IF NOT EXISTS idx_models_metadata_tags             ON models_metadata USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_models_metadata_last_refreshed   ON models_metadata(last_refreshed_at);
CREATE INDEX IF NOT EXISTS idx_models_metadata_model_family     ON models_metadata(model_family);
CREATE INDEX IF NOT EXISTS idx_models_metadata_benchmark_score  ON models_metadata(benchmark_score);

-- Protocol support indexes (migration 003)
CREATE INDEX IF NOT EXISTS idx_models_metadata_protocol_support ON models_metadata USING GIN(protocol_support);

-- Performance indexes (migration 012)
CREATE INDEX IF NOT EXISTS idx_models_by_capabilities
    ON models_metadata (provider_name, supports_streaming, supports_function_calling)
    WHERE provider_name IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_models_scored
    ON models_metadata (benchmark_score DESC, reliability_score DESC)
    WHERE benchmark_score IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_models_tags_gin
    ON models_metadata USING GIN (tags);

CREATE INDEX IF NOT EXISTS idx_models_protocol_gin
    ON models_metadata USING GIN (protocol_support);

COMMENT ON TABLE models_metadata IS 'Model catalog with capabilities, pricing, and protocol support (MCP/LSP/ACP/Embeddings)';

-- -----------------------------------------------------------------------------
-- Table: model_benchmarks
-- -----------------------------------------------------------------------------
-- Individual benchmark results for models. Each row records a specific
-- benchmark test (e.g., MMLU, HumanEval, GSM8K) with raw and normalized
-- scores. A unique constraint prevents duplicate benchmark entries per model.
--
-- Primary Key: id (UUID, auto-generated)
-- Foreign Keys: model_id -> models_metadata(model_id) ON DELETE CASCADE
-- Unique Constraints: (model_id, benchmark_name)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS model_benchmarks (
    id               UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_id         VARCHAR(255)  NOT NULL,                    -- FK to models_metadata.model_id
    benchmark_name   VARCHAR(255)  NOT NULL,                    -- Benchmark name (e.g., 'MMLU', 'HumanEval')
    benchmark_type   VARCHAR(100),                               -- Category (e.g., 'reasoning', 'coding', 'math')
    score            DECIMAL(10,4),                               -- Raw benchmark score
    rank             INTEGER,                                     -- Rank among tested models
    normalized_score DECIMAL(5,2),                                -- Score normalized to 0-100 scale
    benchmark_date   DATE,                                        -- Date the benchmark was run
    metadata         JSONB         DEFAULT '{}',                  -- Additional benchmark metadata
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT fk_model FOREIGN KEY (model_id)
        REFERENCES models_metadata(model_id) ON DELETE CASCADE,
    CONSTRAINT unique_model_benchmark UNIQUE (model_id, benchmark_name)
);

-- Indexes (migration 002)
CREATE INDEX IF NOT EXISTS idx_benchmarks_model_id ON model_benchmarks(model_id);
CREATE INDEX IF NOT EXISTS idx_benchmarks_type     ON model_benchmarks(benchmark_type);
CREATE INDEX IF NOT EXISTS idx_benchmarks_score    ON model_benchmarks(score);

COMMENT ON TABLE model_benchmarks IS 'Individual benchmark results per model';

-- -----------------------------------------------------------------------------
-- Table: models_refresh_history
-- -----------------------------------------------------------------------------
-- Audit log of Models.dev catalog refresh operations. Tracks success/failure
-- counts, duration, and any error messages for operational monitoring.
--
-- Primary Key: id (UUID, auto-generated)
-- No foreign keys. Standalone audit table.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS models_refresh_history (
    id               UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    refresh_type     VARCHAR(50)  NOT NULL,                     -- e.g., 'full', 'incremental', 'provider-specific'
    status           VARCHAR(50)  NOT NULL,                     -- 'success', 'partial', 'failed'
    models_refreshed INTEGER      DEFAULT 0,                    -- Count of models successfully updated
    models_failed    INTEGER      DEFAULT 0,                    -- Count of models that failed to update
    error_message    TEXT,                                       -- Error details if status != 'success'
    started_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at     TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,                                    -- Wall-clock duration of refresh
    metadata         JSONB        DEFAULT '{}'                   -- Additional refresh metadata
);

-- Indexes (migration 002)
CREATE INDEX IF NOT EXISTS idx_refresh_history_started ON models_refresh_history(started_at);
CREATE INDEX IF NOT EXISTS idx_refresh_history_status  ON models_refresh_history(status);

COMMENT ON TABLE models_refresh_history IS 'Audit log of Models.dev catalog refresh operations';
