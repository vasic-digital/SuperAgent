-- HelixAgent Database Initialization Script
-- This script sets up the initial database schema for testing
-- Note: This runs automatically on the database specified by POSTGRES_DB

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

-- Task status enum
DO $$ BEGIN
    CREATE TYPE task_status AS ENUM (
        'pending', 'queued', 'running', 'paused', 'completed',
        'failed', 'stuck', 'cancelled', 'dead_letter'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Task priority enum
DO $$ BEGIN
    CREATE TYPE task_priority AS ENUM (
        'critical', 'high', 'normal', 'low', 'background'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create background_tasks table
CREATE TABLE IF NOT EXISTS background_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_type VARCHAR(100) NOT NULL,
    task_name VARCHAR(255) NOT NULL,
    correlation_id VARCHAR(255),
    parent_task_id UUID REFERENCES background_tasks(id) ON DELETE SET NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    config JSONB NOT NULL DEFAULT '{}',
    priority task_priority NOT NULL DEFAULT 'normal',
    status task_status NOT NULL DEFAULT 'pending',
    progress DECIMAL(5,2) DEFAULT 0.0,
    progress_message TEXT,
    checkpoint JSONB,
    max_retries INTEGER DEFAULT 3,
    retry_count INTEGER DEFAULT 0,
    retry_delay_seconds INTEGER DEFAULT 60,
    last_error TEXT,
    error_history JSONB DEFAULT '[]',
    worker_id VARCHAR(100),
    process_pid INTEGER,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    deadline TIMESTAMP WITH TIME ZONE,
    required_cpu_cores INTEGER DEFAULT 1,
    required_memory_mb INTEGER DEFAULT 512,
    estimated_duration_seconds INTEGER,
    actual_duration_seconds INTEGER,
    notification_config JSONB DEFAULT '{}',
    user_id UUID,
    session_id UUID,
    tags JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    scheduled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create background_tasks_dead_letter table
CREATE TABLE IF NOT EXISTS background_tasks_dead_letter (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    original_task_id UUID NOT NULL,
    task_data JSONB NOT NULL,
    failure_reason TEXT NOT NULL,
    failure_count INTEGER DEFAULT 1,
    moved_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    reprocess_after TIMESTAMP WITH TIME ZONE,
    reprocessed BOOLEAN DEFAULT FALSE
);

-- Create task_execution_history table
CREATE TABLE IF NOT EXISTS task_execution_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES background_tasks(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB DEFAULT '{}',
    worker_id VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create task_resource_snapshots table
CREATE TABLE IF NOT EXISTS task_resource_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES background_tasks(id) ON DELETE CASCADE,
    cpu_percent DECIMAL(5,2),
    cpu_user_time DECIMAL(12,4),
    cpu_system_time DECIMAL(12,4),
    memory_rss_bytes BIGINT,
    memory_vms_bytes BIGINT,
    memory_percent DECIMAL(5,2),
    io_read_bytes BIGINT,
    io_write_bytes BIGINT,
    io_read_count BIGINT,
    io_write_count BIGINT,
    net_bytes_sent BIGINT,
    net_bytes_recv BIGINT,
    net_connections INTEGER,
    open_files INTEGER,
    open_fds INTEGER,
    process_state VARCHAR(20),
    thread_count INTEGER,
    sampled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create webhook_deliveries table
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID REFERENCES background_tasks(id) ON DELETE SET NULL,
    webhook_url TEXT NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    attempts INTEGER DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    last_error TEXT,
    response_code INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    delivered_at TIMESTAMP WITH TIME ZONE
);

-- Create vector_documents table
CREATE TABLE IF NOT EXISTS vector_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    embedding_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    embedding_provider VARCHAR(50) DEFAULT 'pgvector'
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_session_token ON user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_llm_providers_name ON llm_providers(name);
CREATE INDEX IF NOT EXISTS idx_llm_providers_enabled ON llm_providers(enabled);
CREATE INDEX IF NOT EXISTS idx_llm_requests_session_id ON llm_requests(session_id);
CREATE INDEX IF NOT EXISTS idx_llm_requests_user_id ON llm_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_llm_requests_status ON llm_requests(status);
CREATE INDEX IF NOT EXISTS idx_llm_responses_request_id ON llm_responses(request_id);
CREATE INDEX IF NOT EXISTS idx_llm_responses_provider_id ON llm_responses(provider_id);
CREATE INDEX IF NOT EXISTS idx_llm_responses_selected ON llm_responses(selected);
CREATE INDEX IF NOT EXISTS idx_cognee_memories_session_id ON cognee_memories(session_id);
CREATE INDEX IF NOT EXISTS idx_cognee_memories_dataset_name ON cognee_memories(dataset_name);
CREATE INDEX IF NOT EXISTS idx_cognee_memories_search_key ON cognee_memories(search_key);

-- Background tasks indexes
CREATE INDEX IF NOT EXISTS idx_tasks_status ON background_tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority_status ON background_tasks(priority, status, scheduled_at);
CREATE INDEX IF NOT EXISTS idx_tasks_worker ON background_tasks(worker_id) WHERE status = 'running';
CREATE INDEX IF NOT EXISTS idx_tasks_user ON background_tasks(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_correlation ON background_tasks(correlation_id) WHERE correlation_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled ON background_tasks(scheduled_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_tasks_heartbeat ON background_tasks(last_heartbeat) WHERE status = 'running';
CREATE INDEX IF NOT EXISTS idx_tasks_type ON background_tasks(task_type);
CREATE INDEX IF NOT EXISTS idx_tasks_created ON background_tasks(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_task_history_task_id ON task_execution_history(task_id);
CREATE INDEX IF NOT EXISTS idx_task_history_event_type ON task_execution_history(event_type);

CREATE INDEX IF NOT EXISTS idx_resource_snapshots_task ON task_resource_snapshots(task_id, sampled_at DESC);
CREATE INDEX IF NOT EXISTS idx_dead_letter_original ON background_tasks_dead_letter(original_task_id);

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_task ON webhook_deliveries(task_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status ON webhook_deliveries(status) WHERE status != 'delivered';

CREATE INDEX IF NOT EXISTS idx_vector_documents_title ON vector_documents(title);
CREATE INDEX IF NOT EXISTS idx_vector_documents_provider ON vector_documents(embedding_provider);

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
INSERT INTO user_sessions (user_id, session_token, expires_at, context)
SELECT
    id,
    'test-session-token-for-development',
    NOW() + INTERVAL '24 hours',
    '{"test": true}'::jsonb
FROM users
WHERE username = 'testuser'
ON CONFLICT (session_token) DO NOTHING;