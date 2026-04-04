-- =============================================================================
-- HelixAgent Complete Database Schema
-- Version: 2.0.0
-- Date: April 4, 2026
-- Compatible with: PostgreSQL 15+ with pgvector extension
-- =============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- For fuzzy text search
CREATE EXTENSION IF NOT EXISTS "vector";   -- For embeddings

-- =============================================================================
-- CORE TABLES
-- =============================================================================

-- Users and authentication
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) UNIQUE,
    role VARCHAR(50) DEFAULT 'user' CHECK (role IN ('user', 'admin', 'service')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_login TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    rate_limit_tier VARCHAR(50) DEFAULT 'standard'
);

CREATE INDEX idx_users_api_key ON users(api_key);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- API Keys table for multiple keys per user
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    scopes JSONB DEFAULT '["read", "write"]',
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true
);

CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);

-- Sessions for conversation tracking
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255),
    model VARCHAR(100),
    provider VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    is_archived BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}',
    context_window_tokens INTEGER DEFAULT 0
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_created_at ON sessions(created_at);
CREATE INDEX idx_sessions_updated_at ON sessions(updated_at);

-- Messages within sessions
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'assistant', 'system', 'tool')),
    content TEXT NOT NULL,
    tokens_used INTEGER,
    tokens_input INTEGER,
    tokens_output INTEGER,
    latency_ms INTEGER,
    model VARCHAR(100),
    provider VARCHAR(100),
    finish_reason VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_role ON messages(role);
CREATE INDEX idx_messages_content_trgm ON messages USING gin (content gin_trgm_ops);

-- =============================================================================
-- AI DEBATE TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS debate_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic TEXT NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'aborted', 'paused')),
    debate_type VARCHAR(50) DEFAULT 'standard' CHECK (debate_type IN ('standard', 'adversarial', 'reflexion', 'comprehensive')),
    max_turns INTEGER DEFAULT 10,
    current_turn INTEGER DEFAULT 0,
    consensus_threshold FLOAT DEFAULT 0.75,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    final_consensus TEXT,
    winner_id UUID,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_debate_sessions_status ON debate_sessions(status);
CREATE INDEX idx_debate_sessions_created_at ON debate_sessions(created_at);

CREATE TABLE IF NOT EXISTS debate_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    debate_id UUID REFERENCES debate_sessions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(100),
    model VARCHAR(100),
    position VARCHAR(50) CHECK (position IN ('pro', 'con', 'neutral', 'judge', 'observer')),
    system_prompt TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_debate_participants_debate ON debate_participants(debate_id);

CREATE TABLE IF NOT EXISTS debate_turns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    debate_id UUID REFERENCES debate_sessions(id) ON DELETE CASCADE,
    participant_id UUID REFERENCES debate_participants(id) ON DELETE CASCADE,
    parent_turn_id UUID REFERENCES debate_turns(id) ON DELETE SET NULL,
    turn_number INTEGER NOT NULL,
    phase VARCHAR(50) CHECK (phase IN ('proposal', 'critique', 'review', 'optimization', 'convergence')),
    content TEXT NOT NULL,
    reasoning TEXT,
    confidence_score FLOAT,
    votes_received INTEGER DEFAULT 0,
    response_time_ms INTEGER,
    tokens_used INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_debate_turns_debate_id ON debate_turns(debate_id);
CREATE INDEX idx_debate_turns_turn_number ON debate_turns(turn_number);
CREATE INDEX idx_debate_turns_participant ON debate_turns(participant_id);

CREATE TABLE IF NOT EXISTS debate_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    debate_id UUID REFERENCES debate_sessions(id) ON DELETE CASCADE,
    turn_id UUID REFERENCES debate_turns(id) ON DELETE CASCADE,
    voter_id UUID REFERENCES debate_participants(id) ON DELETE CASCADE,
    vote_type VARCHAR(50) CHECK (vote_type IN ('upvote', 'downvote', 'neutral')),
    reasoning TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(debate_id, turn_id, voter_id)
);

-- =============================================================================
-- MEMORY SYSTEM TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS memory_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    entity_type VARCHAR(100) NOT NULL CHECK (entity_type IN ('person', 'place', 'organization', 'concept', 'event', 'preference', 'fact')),
    entity_name VARCHAR(255) NOT NULL,
    entity_value TEXT,
    entity_data JSONB NOT NULL,
    confidence FLOAT DEFAULT 1.0 CHECK (confidence >= 0 AND confidence <= 1),
    source VARCHAR(255),  -- Where was this learned from
    expiration_date TIMESTAMPTZ,
    access_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_memory_entities_session ON memory_entities(session_id);
CREATE INDEX idx_memory_entities_user ON memory_entities(user_id);
CREATE INDEX idx_memory_entities_type ON memory_entities(entity_type);
CREATE INDEX idx_memory_entities_name ON memory_entities(entity_name);
CREATE INDEX idx_memory_entities_data ON memory_entities USING gin (entity_data);

-- Episodic memory (conversations/experiences)
CREATE TABLE IF NOT EXISTS episodic_memory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    memory_type VARCHAR(50) CHECK (memory_type IN ('conversation', 'action', 'observation', 'reflection')),
    summary TEXT NOT NULL,
    full_content TEXT,
    emotional_valence FLOAT CHECK (emotional_valence >= -1 AND emotional_valence <= 1),
    importance_score FLOAT DEFAULT 0.5,
    related_entities UUID[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ,
    access_count INTEGER DEFAULT 0
);

CREATE INDEX idx_episodic_memory_session ON episodic_memory(session_id);
CREATE INDEX idx_episodic_memory_user ON episodic_memory(user_id);
CREATE INDEX idx_episodic_memory_type ON episodic_memory(memory_type);

-- Vector embeddings for semantic search
CREATE TABLE IF NOT EXISTS embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_type VARCHAR(100) NOT NULL CHECK (content_type IN ('message', 'document', 'entity', 'memory', 'debate_turn')),
    content_id UUID NOT NULL,
    embedding vector(1536),  -- OpenAI ada-002 dimension
    text_content TEXT,  -- Original text for reference
    model VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_embeddings_vector ON embeddings USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX idx_embeddings_content ON embeddings(content_type, content_id);

-- =============================================================================
-- PROVIDER & MODEL TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    base_url VARCHAR(500),
    api_version VARCHAR(50),
    auth_type VARCHAR(50) CHECK (auth_type IN ('api_key', 'oauth', 'bearer', 'none')),
    is_active BOOLEAN DEFAULT true,
    is_local BOOLEAN DEFAULT false,
    supports_streaming BOOLEAN DEFAULT true,
    supports_tools BOOLEAN DEFAULT false,
    supports_vision BOOLEAN DEFAULT false,
    rate_limit_rpm INTEGER DEFAULT 60,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID REFERENCES providers(id) ON DELETE CASCADE,
    model_id VARCHAR(100) NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    max_tokens INTEGER,
    context_window INTEGER,
    input_cost_per_1k FLOAT,
    output_cost_per_1k FLOAT,
    is_active BOOLEAN DEFAULT true,
    capabilities JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(provider_id, model_id)
);

CREATE INDEX idx_models_provider ON models(provider_id);

-- =============================================================================
-- ANALYTICS TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS provider_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(100) NOT NULL,
    model VARCHAR(100),
    requests_count INTEGER DEFAULT 0,
    tokens_input BIGINT DEFAULT 0,
    tokens_output BIGINT DEFAULT 0,
    latency_avg_ms INTEGER,
    latency_p50_ms INTEGER,
    latency_p95_ms INTEGER,
    latency_p99_ms INTEGER,
    errors_count INTEGER DEFAULT 0,
    error_rate FLOAT DEFAULT 0,
    date DATE NOT NULL,
    hour INTEGER CHECK (hour >= 0 AND hour <= 23),
    UNIQUE(provider, model, date, hour)
);

CREATE INDEX idx_provider_usage_date ON provider_usage(date);
CREATE INDEX idx_provider_usage_provider ON provider_usage(provider);
CREATE INDEX idx_provider_usage_hour ON provider_usage(hour);

CREATE TABLE IF NOT EXISTS request_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id VARCHAR(255) UNIQUE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    provider VARCHAR(100),
    model VARCHAR(100),
    prompt_tokens INTEGER,
    completion_tokens INTEGER,
    total_tokens INTEGER,
    latency_ms INTEGER,
    status_code INTEGER,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_request_logs_user ON request_logs(user_id);
CREATE INDEX idx_request_logs_session ON request_logs(session_id);
CREATE INDEX idx_request_logs_created ON request_logs(created_at);

-- =============================================================================
-- SECURITY TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    ip_address INET,
    user_agent TEXT,
    status VARCHAR(50) CHECK (status IN ('success', 'failure', 'denied')),
    details JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_ip ON audit_logs(ip_address);

-- Rate limiting buckets (for token bucket algorithm)
CREATE TABLE IF NOT EXISTS rate_limit_buckets (
    key VARCHAR(255) PRIMARY KEY,
    tokens FLOAT NOT NULL,
    last_update TIMESTAMPTZ DEFAULT NOW()
);

-- Blocked IPs
CREATE TABLE IF NOT EXISTS blocked_ips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address INET UNIQUE NOT NULL,
    reason VARCHAR(255),
    blocked_until TIMESTAMPTZ,
    blocked_at TIMESTAMPTZ DEFAULT NOW(),
    blocked_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_blocked_ips_address ON blocked_ips(ip_address);

-- =============================================================================
-- MCP & TOOL TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS mcp_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    endpoint_url VARCHAR(500),
    transport_type VARCHAR(50) CHECK (transport_type IN ('stdio', 'sse', 'http')),
    is_active BOOLEAN DEFAULT true,
    is_local BOOLEAN DEFAULT false,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tool_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
    tool_name VARCHAR(255) NOT NULL,
    tool_input JSONB,
    tool_output JSONB,
    execution_time_ms INTEGER,
    status VARCHAR(50) CHECK (status IN ('pending', 'running', 'success', 'error')),
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_tool_executions_session ON tool_executions(session_id);
CREATE INDEX idx_tool_executions_message ON tool_executions(message_id);

-- =============================================================================
-- FUNCTIONS & TRIGGERS
-- =============================================================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to all tables with updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_debate_sessions_updated_at BEFORE UPDATE ON debate_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_providers_updated_at BEFORE UPDATE ON providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_models_updated_at BEFORE UPDATE ON models
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mcp_servers_updated_at BEFORE UPDATE ON mcp_servers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_memory_entities_updated_at BEFORE UPDATE ON memory_entities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to increment access count
CREATE OR REPLACE FUNCTION increment_access_count()
RETURNS TRIGGER AS $$
BEGIN
    NEW.access_count = OLD.access_count + 1;
    NEW.last_accessed_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- =============================================================================
-- VIEWS
-- =============================================================================

-- Daily usage summary
CREATE OR REPLACE VIEW daily_usage_summary AS
SELECT 
    date,
    provider,
    SUM(requests_count) as total_requests,
    SUM(tokens_input) as total_input_tokens,
    SUM(tokens_output) as total_output_tokens,
    AVG(latency_avg_ms) as avg_latency,
    SUM(errors_count) as total_errors,
    AVG(error_rate) as avg_error_rate
FROM provider_usage
GROUP BY date, provider;

-- Active sessions with stats
CREATE OR REPLACE VIEW session_stats AS
SELECT 
    s.id,
    s.title,
    s.user_id,
    s.created_at,
    s.updated_at,
    COUNT(m.id) as message_count,
    SUM(m.tokens_used) as total_tokens,
    AVG(m.latency_ms) as avg_latency
FROM sessions s
LEFT JOIN messages m ON s.id = m.session_id
WHERE s.is_archived = false
GROUP BY s.id, s.title, s.user_id, s.created_at, s.updated_at;

-- Debate leaderboard
CREATE OR REPLACE VIEW debate_leaderboard AS
SELECT 
    d.id,
    d.topic,
    d.status,
    COUNT(DISTINCT p.id) as participant_count,
    COUNT(DISTINCT t.id) as turn_count,
    d.created_at,
    d.completed_at
FROM debate_sessions d
LEFT JOIN debate_participants p ON d.id = p.debate_id
LEFT JOIN debate_turns t ON d.id = t.debate_id
GROUP BY d.id, d.topic, d.status, d.created_at, d.completed_at;

-- Memory usage by user
CREATE OR REPLACE VIEW user_memory_stats AS
SELECT 
    u.id as user_id,
    u.email,
    COUNT(DISTINCT me.id) as entity_count,
    COUNT(DISTINCT em.id) as episodic_memory_count,
    COUNT(DISTINCT s.id) as session_count
FROM users u
LEFT JOIN memory_entities me ON u.id = me.user_id
LEFT JOIN episodic_memory em ON u.id = em.user_id
LEFT JOIN sessions s ON u.id = s.user_id
GROUP BY u.id, u.email;

-- =============================================================================
-- COMMENTS FOR DOCUMENTATION
-- =============================================================================

COMMENT ON TABLE users IS 'Core user accounts for authentication and authorization';
COMMENT ON TABLE sessions IS 'Conversation sessions between users and AI';
COMMENT ON TABLE messages IS 'Individual messages within sessions';
COMMENT ON TABLE debate_sessions IS 'AI debate orchestration sessions';
COMMENT ON TABLE memory_entities IS 'Semantic memory facts and entities extracted from conversations';
COMMENT ON TABLE episodic_memory IS 'Episodic memory of past conversations and experiences';
COMMENT ON TABLE embeddings IS 'Vector embeddings for semantic search using pgvector';
COMMENT ON TABLE provider_usage IS 'Aggregated usage statistics per provider for analytics';
COMMENT ON TABLE audit_logs IS 'Security audit trail of all significant actions';

-- =============================================================================
-- END OF SCHEMA
-- =============================================================================
