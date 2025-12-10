-- SuperAgent Database Initialization Script
-- This script sets up the initial database schema for testing

-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS superagent_db;

-- Connect to the database
\c superagent_db;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create user_sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Create llm_requests table
CREATE TABLE IF NOT EXISTS llm_requests (
    id SERIAL PRIMARY KEY,
    session_id INTEGER REFERENCES user_sessions(id) ON DELETE CASCADE,
    prompt TEXT NOT NULL,
    messages JSONB NOT NULL DEFAULT '[]',
    model_params JSONB NOT NULL DEFAULT '{}',
    ensemble_config JSONB DEFAULT NULL,
    memory_enhanced BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) DEFAULT 'pending',
    provider_id VARCHAR(100),
    response_content TEXT,
    tokens_used INTEGER DEFAULT 0,
    response_time_ms INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT
);

-- Create llm_responses table
CREATE TABLE IF NOT EXISTS llm_responses (
    id SERIAL PRIMARY KEY,
    request_id INTEGER REFERENCES llm_requests(id) ON DELETE CASCADE,
    provider_id VARCHAR(100) NOT NULL,
    provider_name VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    confidence DECIMAL(3,2) NOT NULL DEFAULT 0.0,
    tokens_used INTEGER DEFAULT 0,
    response_time_ms INTEGER DEFAULT 0,
    finish_reason VARCHAR(50) DEFAULT 'stop',
    metadata JSONB DEFAULT '{}',
    selected BOOLEAN DEFAULT FALSE,
    selection_score DECIMAL(5,2) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create memory_sources table
CREATE TABLE IF NOT EXISTS memory_sources (
    id SERIAL PRIMARY KEY,
    session_id INTEGER REFERENCES user_sessions(id) ON DELETE CASCADE,
    dataset_name VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    vector_id VARCHAR(255),
    content_type VARCHAR(50) DEFAULT 'text',
    relevance_score DECIMAL(5,2) DEFAULT 1.0,
    search_key VARCHAR(255),
    source_type VARCHAR(50) DEFAULT 'cognee',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expired_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() + INTERVAL '7 days'
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_llm_requests_session_id ON llm_requests(session_id);
CREATE INDEX IF NOT EXISTS idx_llm_responses_request_id ON llm_responses(request_id);
CREATE INDEX IF NOT EXISTS idx_memory_sources_session_id ON memory_sources(session_id);
CREATE INDEX IF NOT EXISTS idx_memory_sources_expires_at ON memory_sources(expired_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);

-- Insert test user for development
INSERT INTO users (username, email, password_hash, api_key, role)
VALUES (
    'testuser',
    'test@example.com',
    '$argon2id$v=19$m=65536,t=1,p=4$test_salt$test_hash',
    'sk-test-api-key-for-development',
    'user'
) ON CONFLICT (username) DO NOTHING;

-- Insert test session
INSERT INTO user_sessions (user_id, session_token, expires_at, metadata)
SELECT
    id,
    'test-session-token-for-development',
    NOW() + INTERVAL '24 hours',
    '{"test": true}'
FROM users
WHERE username = 'testuser'
ON CONFLICT (session_token) DO NOTHING;