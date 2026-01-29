-- =============================================================================
-- HelixAgent SQL Schema: Cognee Memories
-- =============================================================================
-- Domain: RAG memory storage via Cognee knowledge graph integration.
-- Source migrations: 001_initial_schema.sql
--
-- Cognee is the knowledge graph and RAG (Retrieval-Augmented Generation)
-- engine used by HelixAgent. This table stores memory entries that are
-- injected into LLM prompts to provide contextual knowledge. Each memory
-- belongs to a session and dataset, with optional vector and graph references.
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: cognee_memories
-- -----------------------------------------------------------------------------
-- Stores Cognee RAG memory entries. Each entry is a piece of knowledge
-- extracted from conversation context, stored with a dataset name for
-- namespace isolation, and optionally linked to vector embeddings and
-- knowledge graph nodes for semantic retrieval.
--
-- Primary Key: id (UUID, auto-generated)
-- Foreign Keys: session_id -> user_sessions(id) ON DELETE CASCADE
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS cognee_memories (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id    UUID         REFERENCES user_sessions(id) ON DELETE CASCADE,  -- Owning session
    dataset_name  VARCHAR(255) NOT NULL,                    -- Dataset namespace (e.g., 'conversation', 'documents')
    content_type  VARCHAR(50)  DEFAULT 'text',              -- Content type: 'text', 'code', 'structured'
    content       TEXT         NOT NULL,                     -- The memory content itself
    vector_id     VARCHAR(255),                              -- Reference to vector embedding in external store
    graph_nodes   JSONB        DEFAULT '{}',                 -- Knowledge graph node references
    search_key    VARCHAR(255),                              -- Indexed search key for fast lookup
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Basic indexes (migration 001)
CREATE INDEX IF NOT EXISTS idx_cognee_memories_session_id   ON cognee_memories(session_id);
CREATE INDEX IF NOT EXISTS idx_cognee_memories_dataset_name ON cognee_memories(dataset_name);
CREATE INDEX IF NOT EXISTS idx_cognee_memories_search_key   ON cognee_memories(search_key);

-- Performance indexes (migration 012)
-- Full-text search index for semantic retrieval over memory content
CREATE INDEX IF NOT EXISTS idx_cognee_content_fts
    ON cognee_memories USING GIN (to_tsvector('english', content));

-- Composite index for recent memories within a dataset
CREATE INDEX IF NOT EXISTS idx_cognee_dataset_recent
    ON cognee_memories (dataset_name, created_at DESC);

COMMENT ON TABLE cognee_memories IS 'Cognee RAG memory storage';
