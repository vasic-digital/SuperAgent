-- =============================================================================
-- HelixAgent SQL Schema: AI Debate System
-- =============================================================================
-- Domain: AI debate round logging, participant tracking, and quality analysis.
-- Source migrations: 014_debate_logs.sql
--
-- The AI Debate system orchestrates multi-round debates between LLM providers.
-- Each debate has multiple participants (proponent, opponent, moderator) who
-- take turns responding. This table logs every action in every round for
-- analytics, quality scoring, and retention management.
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: debate_logs
-- -----------------------------------------------------------------------------
-- Stores a log entry for each participant action in each debate round.
-- Designed for high-volume append-only writes with time-based retention
-- (expires_at). Participant fields are denormalized for query performance.
--
-- Primary Key: id (SERIAL, auto-increment)
-- No foreign keys (uses string-based debate_id and session_id for flexibility)
-- Indexes support: debate lookup, session lookup, provider analytics, retention
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS debate_logs (
    id                     SERIAL       PRIMARY KEY,
    debate_id              VARCHAR(255) NOT NULL,              -- Unique debate identifier
    session_id             VARCHAR(255) NOT NULL,              -- User session identifier
    participant_id         INTEGER,                             -- Numeric participant ID within the debate
    participant_identifier VARCHAR(255),                        -- String identifier for the participant
    participant_name       VARCHAR(255),                        -- Display name (e.g., 'Analyst', 'Critic')
    role                   VARCHAR(100),                        -- Debate role: 'proponent', 'opponent', 'moderator', 'synthesizer'
    provider               VARCHAR(100),                        -- LLM provider name (e.g., 'claude', 'deepseek', 'gemini')
    model                  VARCHAR(255),                        -- Specific model used (e.g., 'claude-3-opus-20240229')
    round                  INTEGER,                             -- Debate round number (1-based)
    action                 VARCHAR(100),                        -- Action type: 'response', 'rebuttal', 'summary', 'vote', 'synthesis'
    response_time_ms       BIGINT,                              -- Time to generate response (milliseconds)
    quality_score          DECIMAL(5,4),                         -- Quality score (0.0000-1.0000)
    tokens_used            INTEGER,                              -- Tokens consumed for this action
    content_length         INTEGER,                              -- Response content length in characters
    error_message          TEXT,                                  -- Error message if the action failed
    metadata               JSONB        DEFAULT '{}',            -- Additional metadata (confidence, reasoning chain, etc.)
    created_at             TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Log entry creation time
    expires_at             TIMESTAMP WITH TIME ZONE               -- Optional expiration for log retention
);

-- Column comments
COMMENT ON TABLE debate_logs IS 'Stores AI debate round logs and participant responses';
COMMENT ON COLUMN debate_logs.debate_id IS 'Unique identifier for the debate session';
COMMENT ON COLUMN debate_logs.session_id IS 'User session identifier';
COMMENT ON COLUMN debate_logs.participant_id IS 'Numeric ID of the debate participant';
COMMENT ON COLUMN debate_logs.participant_identifier IS 'String identifier for the participant';
COMMENT ON COLUMN debate_logs.participant_name IS 'Display name of the participant';
COMMENT ON COLUMN debate_logs.role IS 'Role of the participant in the debate (proponent, opponent, moderator, etc.)';
COMMENT ON COLUMN debate_logs.provider IS 'LLM provider name (claude, deepseek, gemini, etc.)';
COMMENT ON COLUMN debate_logs.model IS 'Specific model used for this response';
COMMENT ON COLUMN debate_logs.round IS 'Debate round number';
COMMENT ON COLUMN debate_logs.action IS 'Type of action (response, rebuttal, summary, etc.)';
COMMENT ON COLUMN debate_logs.response_time_ms IS 'Time taken to generate response in milliseconds';
COMMENT ON COLUMN debate_logs.quality_score IS 'Quality score of the response (0.0-1.0)';
COMMENT ON COLUMN debate_logs.tokens_used IS 'Number of tokens used in the response';
COMMENT ON COLUMN debate_logs.content_length IS 'Length of the response content in characters';
COMMENT ON COLUMN debate_logs.error_message IS 'Error message if the response failed';
COMMENT ON COLUMN debate_logs.metadata IS 'Additional metadata as JSONB';
COMMENT ON COLUMN debate_logs.created_at IS 'Timestamp when the log was created';
COMMENT ON COLUMN debate_logs.expires_at IS 'Optional expiration time for log retention';

-- Basic indexes (migration 014)
CREATE INDEX IF NOT EXISTS idx_debate_logs_debate_id  ON debate_logs(debate_id);
CREATE INDEX IF NOT EXISTS idx_debate_logs_session_id ON debate_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_debate_logs_provider   ON debate_logs(provider);
CREATE INDEX IF NOT EXISTS idx_debate_logs_model      ON debate_logs(model);
CREATE INDEX IF NOT EXISTS idx_debate_logs_created_at ON debate_logs(created_at);

-- Expiration index for retention cleanup
CREATE INDEX IF NOT EXISTS idx_debate_logs_expires_at
    ON debate_logs(expires_at) WHERE expires_at IS NOT NULL;

-- Composite index for debate round queries
CREATE INDEX IF NOT EXISTS idx_debate_logs_debate_round
    ON debate_logs(debate_id, round);

-- Partial index for active (non-expired) debates
CREATE INDEX IF NOT EXISTS idx_debate_logs_active
    ON debate_logs(debate_id)
    WHERE expires_at IS NULL OR expires_at > NOW();

-- Composite index for provider performance analysis
CREATE INDEX IF NOT EXISTS idx_debate_logs_provider_model
    ON debate_logs(provider, model);

-- GIN index for JSONB metadata queries
CREATE INDEX IF NOT EXISTS idx_debate_logs_metadata
    ON debate_logs USING GIN (metadata);

-- Retention cleanup function
CREATE OR REPLACE FUNCTION cleanup_expired_debate_logs()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM debate_logs
    WHERE expires_at IS NOT NULL AND expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_debate_logs() IS 'Removes expired debate logs based on expires_at column';
