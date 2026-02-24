-- =============================================================================
-- HelixAgent SQL Schema: Debate Sessions
-- =============================================================================
-- Domain: Debate lifecycle tracking with full metadata for replay/recovery.
-- Extends the debate_logs system with session-level state management.
--
-- A debate session represents a complete debate lifecycle from initiation
-- through completion. It tracks topology, coordination protocol, configuration,
-- round counts, consensus scores, and final outcomes. Sessions can be paused
-- and resumed, enabling human-in-the-loop approval gates.
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: debate_sessions
-- -----------------------------------------------------------------------------
-- Tracks the lifecycle of each debate session with full metadata.
-- Links to debate_logs via debate_id (string-based, no FK for flexibility).
-- Supports pause/resume via status transitions for approval gates.
--
-- Primary Key: id (UUID, auto-generated)
-- No foreign keys to debate_logs (string-based debate_id for flexibility)
-- Indexes support: debate lookup, status filtering, topology analytics
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS debate_sessions (
    id                    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    debate_id             VARCHAR(255) NOT NULL,              -- Links to debate_logs.debate_id
    topic                 TEXT         NOT NULL,              -- Debate topic or task description
    status                VARCHAR(50)  NOT NULL DEFAULT 'pending', -- Session lifecycle status
    topology_type         VARCHAR(50),                         -- graph_mesh, star, chain, tree
    coordination_protocol VARCHAR(50),                         -- cpde, dpde, adaptive
    config                JSONB        DEFAULT '{}',           -- Max rounds, timeout, consensus threshold, gates config
    initiated_by          VARCHAR(255),                        -- Requester/initiator identifier
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at          TIMESTAMP WITH TIME ZONE,            -- Set when status transitions to completed/failed/cancelled
    total_rounds          INTEGER      DEFAULT 0,              -- Total rounds completed
    final_consensus_score DECIMAL(5,4),                        -- Final consensus level (0.0000-1.0000)
    outcome               JSONB        DEFAULT '{}',           -- Winner, voting method, confidence, summary
    metadata              JSONB        DEFAULT '{}'            -- Additional structured data (audit trail, provenance)
);

-- Status constraint: only valid lifecycle states
ALTER TABLE debate_sessions
    ADD CONSTRAINT chk_debate_sessions_status
    CHECK (status IN ('pending', 'running', 'paused', 'completed', 'failed', 'cancelled'));

-- Column comments
COMMENT ON TABLE debate_sessions IS 'Tracks debate session lifecycle with metadata for replay/recovery';
COMMENT ON COLUMN debate_sessions.id IS 'Unique session identifier (UUID)';
COMMENT ON COLUMN debate_sessions.debate_id IS 'Links to debate_logs.debate_id for log correlation';
COMMENT ON COLUMN debate_sessions.topic IS 'Debate topic or task description';
COMMENT ON COLUMN debate_sessions.status IS 'Session status: pending, running, paused, completed, failed, cancelled';
COMMENT ON COLUMN debate_sessions.topology_type IS 'Debate topology: graph_mesh, star, chain, tree';
COMMENT ON COLUMN debate_sessions.coordination_protocol IS 'Planning style: cpde, dpde, adaptive';
COMMENT ON COLUMN debate_sessions.config IS 'Session configuration as JSONB (max_rounds, timeout, consensus_threshold, gates)';
COMMENT ON COLUMN debate_sessions.initiated_by IS 'Identifier of the session initiator';
COMMENT ON COLUMN debate_sessions.total_rounds IS 'Number of debate rounds completed';
COMMENT ON COLUMN debate_sessions.final_consensus_score IS 'Final consensus score (0.0-1.0)';
COMMENT ON COLUMN debate_sessions.outcome IS 'Final outcome as JSONB (winner, voting_method, confidence)';
COMMENT ON COLUMN debate_sessions.metadata IS 'Additional metadata as JSONB (audit trail, provenance)';

-- Basic lookup indexes
CREATE INDEX IF NOT EXISTS idx_debate_sessions_debate_id
    ON debate_sessions(debate_id);

CREATE INDEX IF NOT EXISTS idx_debate_sessions_status
    ON debate_sessions(status);

CREATE INDEX IF NOT EXISTS idx_debate_sessions_created_at
    ON debate_sessions(created_at);

CREATE INDEX IF NOT EXISTS idx_debate_sessions_topology
    ON debate_sessions(topology_type);

-- Partial index for active sessions (pending, running, paused)
CREATE INDEX IF NOT EXISTS idx_debate_sessions_active
    ON debate_sessions(status)
    WHERE status IN ('pending', 'running', 'paused');

-- GIN index for JSONB metadata queries
CREATE INDEX IF NOT EXISTS idx_debate_sessions_metadata
    ON debate_sessions USING GIN (metadata);

-- GIN index for JSONB config queries
CREATE INDEX IF NOT EXISTS idx_debate_sessions_config
    ON debate_sessions USING GIN (config);

-- Composite index for debate + status queries
CREATE INDEX IF NOT EXISTS idx_debate_sessions_debate_status
    ON debate_sessions(debate_id, status);

-- Auto-update updated_at trigger
CREATE OR REPLACE FUNCTION update_debate_sessions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_debate_sessions_updated_at ON debate_sessions;
CREATE TRIGGER trg_debate_sessions_updated_at
    BEFORE UPDATE ON debate_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_debate_sessions_updated_at();

COMMENT ON FUNCTION update_debate_sessions_updated_at() IS 'Auto-updates updated_at timestamp on debate_sessions row changes';
