-- =============================================================================
-- HelixAgent SQL Schema: Users & Sessions
-- =============================================================================
-- Domain: Authentication, authorization, and session management.
-- Source migrations: 001_initial_schema.sql
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: users
-- -----------------------------------------------------------------------------
-- Stores registered user accounts for HelixAgent. Each user has a unique
-- username, email, and API key for programmatic access. The role column
-- controls authorization levels (e.g., 'user', 'admin').
--
-- Primary Key: id (UUID, auto-generated via uuid_generate_v4)
-- Unique Constraints: username, email, api_key
-- Referenced by: user_sessions.user_id, llm_requests.user_id
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    username      VARCHAR(255) UNIQUE NOT NULL,          -- Unique login name
    email         VARCHAR(255) UNIQUE NOT NULL,          -- Unique email address
    password_hash VARCHAR(255) NOT NULL,                 -- Bcrypt/argon2 hashed password
    api_key       VARCHAR(255) UNIQUE NOT NULL,          -- API key for programmatic access
    role          VARCHAR(50)  DEFAULT 'user',            -- Authorization role: 'user', 'admin'
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Performance indexes for users
CREATE INDEX IF NOT EXISTS idx_users_email   ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);

COMMENT ON TABLE users IS 'User accounts for HelixAgent';

-- -----------------------------------------------------------------------------
-- Table: user_sessions
-- -----------------------------------------------------------------------------
-- Tracks active user sessions. Each session belongs to a user and carries
-- a JSONB context object that accumulates conversation state. Sessions
-- expire after a configurable TTL and track request counts for rate limiting.
--
-- Primary Key: id (UUID, auto-generated)
-- Foreign Keys: user_id -> users(id) ON DELETE CASCADE
-- Unique Constraints: session_token
-- Referenced by: llm_requests.session_id, cognee_memories.session_id
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_sessions (
    id             UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID         REFERENCES users(id) ON DELETE CASCADE,  -- Owning user
    session_token  VARCHAR(255) UNIQUE NOT NULL,          -- Opaque bearer token
    context        JSONB        DEFAULT '{}',              -- Accumulated conversation context
    memory_id      UUID,                                   -- Optional Cognee memory reference
    status         VARCHAR(50)  DEFAULT 'active',          -- Session state: 'active', 'expired', 'revoked'
    request_count  INTEGER      DEFAULT 0,                 -- Total requests in this session
    last_activity  TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Last request timestamp
    expires_at     TIMESTAMP WITH TIME ZONE NOT NULL,      -- Session expiration time
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Performance indexes for user_sessions
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id       ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at    ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_session_token ON user_sessions(session_token);

-- Migration 012: Hot-path index for active session lookups
CREATE INDEX IF NOT EXISTS idx_sessions_active
    ON user_sessions (user_id, status, last_activity DESC)
    WHERE status = 'active';

-- Migration 012: Index for session expiration cleanup jobs
CREATE INDEX IF NOT EXISTS idx_sessions_expired
    ON user_sessions (expires_at)
    WHERE status = 'active' AND expires_at < NOW();

COMMENT ON TABLE user_sessions IS 'Active user sessions with context';
