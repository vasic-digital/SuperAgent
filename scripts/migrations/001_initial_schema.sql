-- Migration: 001_initial_schema
-- Description: Initial HelixAgent database schema
-- Date: 2026-03-17

BEGIN;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    context JSONB DEFAULT '{}',
    memory_id UUID,
    status VARCHAR(50) DEFAULT 'active',
    request_count INTEGER DEFAULT 0,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create llm_providers table
CREATE TABLE IF NOT EXISTS llm_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(100) NOT NULL,
    api_key VARCHAR(255),
    base_url VARCHAR(500),
    model VARCHAR(255),
    weight DECIMAL(5,2) DEFAULT 1.0,
    enabled BOOLEAN DEFAULT TRUE,
    config JSONB DEFAULT '{}',
    health_status VARCHAR(50) DEFAULT 'unknown',
    response_time BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create llm_requests table
CREATE TABLE IF NOT EXISTS llm_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES user_sessions(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    prompt TEXT NOT NULL,
    messages JSONB NOT NULL DEFAULT '[]',
    model_params JSONB NOT NULL DEFAULT '{}',
    ensemble_config JSONB DEFAULT NULL,
    memory_enhanced BOOLEAN DEFAULT FALSE,
    memory JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    request_type VARCHAR(50) DEFAULT 'completion'
);

-- Create llm_responses table
CREATE TABLE IF NOT EXISTS llm_responses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id UUID REFERENCES llm_requests(id) ON DELETE CASCADE,
    provider_id UUID REFERENCES llm_providers(id) ON DELETE SET NULL,
    provider_name VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    confidence DECIMAL(3,2) NOT NULL DEFAULT 0.0,
    tokens_used INTEGER DEFAULT 0,
    response_time BIGINT DEFAULT 0,
    finish_reason VARCHAR(50) DEFAULT 'stop',
    metadata JSONB DEFAULT '{}',
    selected BOOLEAN DEFAULT FALSE,
    selection_score DECIMAL(5,2) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create cognee_memories table
CREATE TABLE IF NOT EXISTS cognee_memories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES user_sessions(id) ON DELETE CASCADE,
    dataset_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(50) DEFAULT 'text',
    content TEXT NOT NULL,
    vector_id VARCHAR(255),
    graph_nodes JSONB DEFAULT '{}',
    search_key VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create model_metadata table for model catalog
CREATE TABLE IF NOT EXISTS model_metadata (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_id VARCHAR(255) UNIQUE NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    provider VARCHAR(100) NOT NULL,
    model_type VARCHAR(100),
    description TEXT,
    capabilities JSONB DEFAULT '[]',
    context_length INTEGER,
    max_tokens INTEGER,
    input_price DECIMAL(20,10),
    output_price DECIMAL(20,10),
    supports_streaming BOOLEAN DEFAULT FALSE,
    supports_function_calling BOOLEAN DEFAULT FALSE,
    supports_vision BOOLEAN DEFAULT FALSE,
    supports_json_mode BOOLEAN DEFAULT FALSE,
    quality_score DECIMAL(3,2),
    speed_score DECIMAL(3,2),
    cost_score DECIMAL(3,2),
    enabled BOOLEAN DEFAULT TRUE,
    metadata JSONB DEFAULT '{}',
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_session_token ON user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_llm_providers_name ON llm_providers(name);
CREATE INDEX IF NOT EXISTS idx_llm_requests_session_id ON llm_requests(session_id);
CREATE INDEX IF NOT EXISTS idx_llm_responses_request_id ON llm_responses(request_id);
CREATE INDEX IF NOT EXISTS idx_model_metadata_provider ON model_metadata(provider);
CREATE INDEX IF NOT EXISTS idx_model_metadata_model_type ON model_metadata(model_type);

COMMIT;
