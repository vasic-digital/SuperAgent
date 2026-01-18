-- Migration: 014_debate_logs
-- Description: Create debate_logs table for AI debate history tracking
-- Date: 2026-01-18
-- Author: HelixAgent Team

-- Create debate_logs table for storing AI debate round history
CREATE TABLE IF NOT EXISTS debate_logs (
    id SERIAL PRIMARY KEY,
    debate_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    participant_id INTEGER,
    participant_identifier VARCHAR(255),
    participant_name VARCHAR(255),
    role VARCHAR(100),
    provider VARCHAR(100),
    model VARCHAR(255),
    round INTEGER,
    action VARCHAR(100),
    response_time_ms BIGINT,
    quality_score DECIMAL(5, 4),
    tokens_used INTEGER,
    content_length INTEGER,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Add comments for documentation
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

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_debate_logs_debate_id ON debate_logs(debate_id);
CREATE INDEX IF NOT EXISTS idx_debate_logs_session_id ON debate_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_debate_logs_provider ON debate_logs(provider);
CREATE INDEX IF NOT EXISTS idx_debate_logs_model ON debate_logs(model);
CREATE INDEX IF NOT EXISTS idx_debate_logs_created_at ON debate_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_debate_logs_expires_at ON debate_logs(expires_at) WHERE expires_at IS NOT NULL;

-- Composite index for debate queries
CREATE INDEX IF NOT EXISTS idx_debate_logs_debate_round ON debate_logs(debate_id, round);

-- Partial index for active debates (not expired)
CREATE INDEX IF NOT EXISTS idx_debate_logs_active ON debate_logs(debate_id)
    WHERE expires_at IS NULL OR expires_at > NOW();

-- Index for provider performance analysis
CREATE INDEX IF NOT EXISTS idx_debate_logs_provider_model ON debate_logs(provider, model);

-- GIN index for JSONB metadata queries
CREATE INDEX IF NOT EXISTS idx_debate_logs_metadata ON debate_logs USING GIN (metadata);

-- Function to clean up expired debate logs
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

-- Grant permissions (adjust based on your user setup)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON debate_logs TO helixagent;
-- GRANT USAGE, SELECT ON SEQUENCE debate_logs_id_seq TO helixagent;
