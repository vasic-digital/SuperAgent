-- =============================================================================
-- HelixAgent SQL Schema: Debate Turns
-- =============================================================================
-- Domain: Granular turn-level state for debate replay and recovery.
-- Each turn captures one agent's action in one phase of one round.
--
-- Debate turns store the complete record of every agent interaction during
-- a debate session, including content, confidence scores, tool calls,
-- test results, and Reflexion episodic memory (reflections). This enables
-- full debate replay, provenance tracking, and failure analysis.
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: debate_turns
-- -----------------------------------------------------------------------------
-- Stores each individual agent action within a debate round and phase.
-- Foreign key to debate_sessions for referential integrity.
-- The reflections column stores Reflexion framework episodic memory entries.
--
-- Primary Key: id (UUID, auto-generated)
-- Foreign Key: session_id -> debate_sessions(id) with CASCADE delete
-- Indexes support: session lookup, round/phase filtering, agent tracking
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS debate_turns (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id       UUID         NOT NULL REFERENCES debate_sessions(id) ON DELETE CASCADE,
    round            INTEGER      NOT NULL,                    -- Debate round number (1-based)
    phase            VARCHAR(50)  NOT NULL,                    -- Protocol phase
    agent_id         VARCHAR(255) NOT NULL,                    -- Agent identifier
    agent_role       VARCHAR(100),                              -- Agent role in this turn
    provider         VARCHAR(100),                              -- LLM provider used
    model            VARCHAR(255),                              -- Specific model used
    content          TEXT,                                      -- Response content
    confidence       DECIMAL(5,4),                              -- Agent confidence (0.0000-1.0000)
    tool_calls       JSONB        DEFAULT '[]',                -- Tool invocations and results
    test_results     JSONB        DEFAULT '{}',                -- Test execution results
    reflections      JSONB        DEFAULT '[]',                -- Reflexion episodic memory entries
    metadata         JSONB        DEFAULT '{}',                -- Additional structured data
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    response_time_ms INTEGER                                    -- Response latency in milliseconds
);

-- Phase constraint: only valid protocol phases
ALTER TABLE debate_turns
    ADD CONSTRAINT chk_debate_turns_phase
    CHECK (phase IN (
        'dehallucination', 'self_evolvement', 'proposal', 'critique',
        'review', 'optimization', 'adversarial', 'convergence'
    ));

-- Column comments
COMMENT ON TABLE debate_turns IS 'Stores individual agent actions within debate rounds for replay/recovery';
COMMENT ON COLUMN debate_turns.id IS 'Unique turn identifier (UUID)';
COMMENT ON COLUMN debate_turns.session_id IS 'References the parent debate session';
COMMENT ON COLUMN debate_turns.round IS 'Debate round number (1-based)';
COMMENT ON COLUMN debate_turns.phase IS 'Protocol phase: dehallucination, self_evolvement, proposal, critique, review, optimization, adversarial, convergence';
COMMENT ON COLUMN debate_turns.agent_id IS 'Identifier of the agent that produced this turn';
COMMENT ON COLUMN debate_turns.agent_role IS 'Role the agent played in this turn';
COMMENT ON COLUMN debate_turns.provider IS 'LLM provider name';
COMMENT ON COLUMN debate_turns.model IS 'Specific model used';
COMMENT ON COLUMN debate_turns.content IS 'Response content produced by the agent';
COMMENT ON COLUMN debate_turns.confidence IS 'Agent confidence score (0.0-1.0)';
COMMENT ON COLUMN debate_turns.tool_calls IS 'Tool invocations and their results as JSONB array';
COMMENT ON COLUMN debate_turns.test_results IS 'Test execution results as JSONB';
COMMENT ON COLUMN debate_turns.reflections IS 'Reflexion framework episodic memory entries as JSONB array';
COMMENT ON COLUMN debate_turns.metadata IS 'Additional metadata as JSONB';
COMMENT ON COLUMN debate_turns.response_time_ms IS 'Time taken to generate response in milliseconds';

-- Session lookup (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_debate_turns_session_id
    ON debate_turns(session_id);

-- Round-level queries within a session
CREATE INDEX IF NOT EXISTS idx_debate_turns_session_round
    ON debate_turns(session_id, round);

-- Phase filtering
CREATE INDEX IF NOT EXISTS idx_debate_turns_phase
    ON debate_turns(phase);

-- Agent tracking
CREATE INDEX IF NOT EXISTS idx_debate_turns_agent
    ON debate_turns(agent_id);

-- Composite: session + round + phase (most specific query)
CREATE INDEX IF NOT EXISTS idx_debate_turns_session_round_phase
    ON debate_turns(session_id, round, phase);

-- Timestamp-based queries
CREATE INDEX IF NOT EXISTS idx_debate_turns_created_at
    ON debate_turns(created_at);

-- GIN index for reflections JSONB (only where reflections exist)
CREATE INDEX IF NOT EXISTS idx_debate_turns_reflections
    ON debate_turns USING GIN (reflections)
    WHERE reflections != '[]'::jsonb;

-- GIN index for tool_calls JSONB (only where tool calls exist)
CREATE INDEX IF NOT EXISTS idx_debate_turns_tool_calls
    ON debate_turns USING GIN (tool_calls)
    WHERE tool_calls != '[]'::jsonb;

-- GIN index for metadata JSONB
CREATE INDEX IF NOT EXISTS idx_debate_turns_metadata
    ON debate_turns USING GIN (metadata);
