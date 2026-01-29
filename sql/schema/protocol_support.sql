-- =============================================================================
-- HelixAgent SQL Schema: Protocol Support (MCP, LSP, ACP, Embeddings)
-- =============================================================================
-- Domain: Multi-protocol server configuration, vector search, caching, metrics.
-- Source migrations: 003_protocol_support.sql
--
-- HelixAgent supports four protocols for tool/language/agent integration:
--   MCP (Model Context Protocol)  - Tool server integration
--   LSP (Language Server Protocol) - Code intelligence
--   ACP (Agent Communication Protocol) - Agent-to-agent communication
--   Embeddings - Vector embedding and semantic search
-- =============================================================================

-- Required extension for vector operations
CREATE EXTENSION IF NOT EXISTS pgvector;

-- -----------------------------------------------------------------------------
-- Table: mcp_servers
-- -----------------------------------------------------------------------------
-- Configuration for MCP (Model Context Protocol) servers. MCP servers
-- provide tools that LLMs can invoke (e.g., file operations, database
-- queries, API calls). Servers can be local (spawned via command) or
-- remote (accessed via URL).
--
-- Primary Key: id (VARCHAR, user-defined server identifier)
-- No foreign keys. Referenced by: protocol_metrics.server_id (logical)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mcp_servers (
    id         VARCHAR(255) PRIMARY KEY,                       -- Server identifier (e.g., 'filesystem', 'github')
    name       VARCHAR(255) NOT NULL,                           -- Human-readable server name
    type       VARCHAR(20)  NOT NULL CHECK (type IN ('local', 'remote')), -- 'local' = subprocess, 'remote' = HTTP/SSE
    command    TEXT,                                             -- Launch command for local servers (e.g., 'npx @modelcontextprotocol/server-filesystem')
    url        TEXT,                                             -- Endpoint URL for remote servers
    enabled    BOOLEAN      NOT NULL DEFAULT true,              -- Whether this server is active
    tools      JSONB        NOT NULL DEFAULT '[]',              -- Array of tool definitions provided by this server
    last_sync  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- Last tool list synchronization
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_mcp_servers_enabled ON mcp_servers USING btree (enabled);

-- Performance index (migration 012): active MCP server lookup
CREATE INDEX IF NOT EXISTS idx_mcp_active_servers
    ON mcp_servers (type, enabled, last_sync DESC)
    WHERE enabled = TRUE;

COMMENT ON TABLE mcp_servers IS 'MCP (Model Context Protocol) server configuration';

-- -----------------------------------------------------------------------------
-- Table: lsp_servers
-- -----------------------------------------------------------------------------
-- Configuration for LSP (Language Server Protocol) servers. Each LSP server
-- provides code intelligence for a specific programming language (completions,
-- diagnostics, go-to-definition, etc.).
--
-- Primary Key: id (VARCHAR, user-defined server identifier)
-- No foreign keys.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS lsp_servers (
    id           VARCHAR(255)  PRIMARY KEY,                    -- Server identifier (e.g., 'gopls', 'typescript-language-server')
    name         VARCHAR(255)  NOT NULL,                        -- Human-readable server name
    language     VARCHAR(50)   NOT NULL,                        -- Target language (e.g., 'go', 'typescript', 'python')
    command      VARCHAR(500)  NOT NULL,                        -- Launch command (e.g., 'gopls serve')
    enabled      BOOLEAN       NOT NULL DEFAULT true,           -- Whether this server is active
    workspace    VARCHAR(1000) DEFAULT '/workspace',            -- Default workspace root path
    capabilities JSONB         NOT NULL DEFAULT '[]',           -- LSP capabilities (textDocument/completion, etc.)
    last_sync    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_lsp_servers_enabled ON lsp_servers USING btree (enabled);

COMMENT ON TABLE lsp_servers IS 'LSP (Language Server Protocol) server configuration';

-- -----------------------------------------------------------------------------
-- Table: acp_servers
-- -----------------------------------------------------------------------------
-- Configuration for ACP (Agent Communication Protocol) servers. ACP enables
-- agent-to-agent communication using JSON-RPC 2.0 over stdin/stdout or HTTP.
-- Similar structure to MCP servers.
--
-- Primary Key: id (VARCHAR, user-defined server identifier)
-- No foreign keys.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS acp_servers (
    id         VARCHAR(255) PRIMARY KEY,                       -- Server identifier
    name       VARCHAR(255) NOT NULL,                           -- Human-readable server name
    type       VARCHAR(20)  NOT NULL CHECK (type IN ('local', 'remote')), -- 'local' = subprocess, 'remote' = HTTP
    url        TEXT,                                             -- Endpoint URL for remote servers
    enabled    BOOLEAN      NOT NULL DEFAULT true,              -- Whether this server is active
    tools      JSONB        NOT NULL DEFAULT '[]',              -- Array of tool definitions
    last_sync  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_acp_servers_enabled ON acp_servers USING btree (enabled);

COMMENT ON TABLE acp_servers IS 'ACP (Agent Communication Protocol) server configuration';

-- -----------------------------------------------------------------------------
-- Table: embedding_config
-- -----------------------------------------------------------------------------
-- Configuration for the embedding provider used for vector operations.
-- Stores the active provider, model, dimension, and API credentials.
-- Typically has a single active row, but supports multiple configurations
-- for A/B testing or provider migration.
--
-- Primary Key: id (SERIAL, auto-increment)
-- No foreign keys.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS embedding_config (
    id           SERIAL       PRIMARY KEY,
    provider     VARCHAR(50)  NOT NULL DEFAULT 'pgvector',     -- Provider name (e.g., 'pgvector', 'openai', 'cohere')
    model        VARCHAR(100) NOT NULL DEFAULT 'text-embedding-ada-002', -- Embedding model name
    dimension    INTEGER      NOT NULL DEFAULT 1536,            -- Vector dimension (must match model output)
    api_endpoint TEXT,                                           -- Provider API endpoint URL
    api_key      TEXT,                                           -- Provider API key (encrypted at rest)
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE embedding_config IS 'Embedding provider configuration for vector operations';

-- -----------------------------------------------------------------------------
-- Table: vector_documents
-- -----------------------------------------------------------------------------
-- Stores documents with their vector embeddings for semantic search.
-- Uses pgvector extension for efficient similarity search (cosine distance).
-- Each document has a primary embedding and an optional search_vector for
-- different embedding strategies (e.g., query vs. document embeddings).
--
-- Primary Key: id (UUID, auto-generated via gen_random_uuid)
-- No foreign keys.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vector_documents (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title              TEXT         NOT NULL,                    -- Document title
    content            TEXT         NOT NULL,                    -- Full document content
    metadata           JSONB        DEFAULT '{}',                -- Arbitrary metadata (source, author, tags, etc.)
    embedding_id       UUID,                                     -- Optional reference to external embedding store
    embedding          VECTOR(1536),                              -- Primary document embedding (1536-dim for ada-002)
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    embedding_provider VARCHAR(50)  DEFAULT 'pgvector',          -- Provider that generated the embedding
    search_vector      VECTOR(1536)                               -- Secondary search embedding (optional, for asymmetric search)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_vector_documents_embedding_provider
    ON vector_documents USING btree (embedding_provider);

COMMENT ON TABLE vector_documents IS 'Documents with vector embeddings for semantic search';

-- -----------------------------------------------------------------------------
-- Table: protocol_cache
-- -----------------------------------------------------------------------------
-- Generic cache for protocol operations. Stores serialized response data
-- with TTL-based expiration. Used to cache MCP tool lists, LSP capabilities,
-- and other protocol responses that change infrequently.
--
-- Primary Key: cache_key (VARCHAR, caller-defined cache key)
-- No foreign keys.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS protocol_cache (
    cache_key  VARCHAR(255) PRIMARY KEY,                        -- Cache key (e.g., 'mcp:filesystem:tools')
    cache_data JSONB        NOT NULL,                            -- Cached data (serialized response)
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- Cache entry expiration
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_protocol_cache_expires_at
    ON protocol_cache USING btree (expires_at);

-- Performance index (migration 012): expired cache entry cleanup
CREATE INDEX IF NOT EXISTS idx_cache_expiration
    ON protocol_cache (expires_at)
    WHERE expires_at < NOW();

COMMENT ON TABLE protocol_cache IS 'TTL-based cache for protocol operation responses';

-- -----------------------------------------------------------------------------
-- Table: protocol_metrics
-- -----------------------------------------------------------------------------
-- Operational metrics for all protocol operations. Each row records a single
-- operation (tool call, LSP request, embedding generation) with its status,
-- duration, and optional error details. Used for health monitoring dashboards
-- and materialized view aggregation.
--
-- Primary Key: id (SERIAL, auto-increment)
-- No foreign keys (server_id is a logical reference)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS protocol_metrics (
    id             SERIAL       PRIMARY KEY,
    protocol_type  VARCHAR(20)  NOT NULL CHECK (protocol_type IN ('mcp', 'lsp', 'acp', 'embedding')),
    server_id      VARCHAR(255),                                 -- Logical reference to *_servers.id
    operation      VARCHAR(100) NOT NULL,                        -- Operation name (e.g., 'tool_call', 'completion', 'embed')
    status         VARCHAR(20)  NOT NULL CHECK (status IN ('success', 'error', 'timeout')),
    duration_ms    INTEGER,                                      -- Operation duration in milliseconds
    error_message  TEXT,                                          -- Error details (NULL on success)
    metadata       JSONB        DEFAULT '{}',                     -- Additional operation metadata
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Basic indexes (migration 003)
CREATE INDEX IF NOT EXISTS idx_protocol_metrics_protocol_type
    ON protocol_metrics USING btree (protocol_type);
CREATE INDEX IF NOT EXISTS idx_protocol_metrics_created_at
    ON protocol_metrics USING btree (created_at);

-- Performance indexes (migration 012)
-- Time-series covering index for dashboard queries
CREATE INDEX IF NOT EXISTS idx_protocol_metrics_timeseries
    ON protocol_metrics (protocol_type, created_at DESC)
    INCLUDE (status, duration_ms);

-- Server-scoped aggregation index
CREATE INDEX IF NOT EXISTS idx_protocol_metrics_by_server
    ON protocol_metrics (server_id, operation, status)
    WHERE server_id IS NOT NULL;

-- BRIN index for efficient time-range scans (append-only workload)
CREATE INDEX IF NOT EXISTS idx_protocol_metrics_brin
    ON protocol_metrics USING BRIN (created_at);

COMMENT ON TABLE protocol_metrics IS 'Operational metrics for MCP, LSP, ACP, and embedding protocol operations';
