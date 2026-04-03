-- Migration: CLI Agents Fusion Schema
-- Description: Database schema for multi-instance ensemble system with CLI agent integrations
-- Version: 1.0.0
-- Created: 2026-04-03

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- For fuzzy text search

-- ============================================
-- CORE TABLES
-- ============================================

-- Agent instances table - tracks running CLI agent instances
CREATE TABLE agent_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_type VARCHAR(50) NOT NULL,
    instance_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'creating',
    
    -- Configuration (stored as JSON for flexibility)
    config JSONB NOT NULL DEFAULT '{}',
    provider_config JSONB DEFAULT '{}',
    
    -- Resource limits
    max_memory_mb INTEGER,
    max_cpu_percent INTEGER,
    max_disk_mb INTEGER,
    
    -- Current state
    current_session_id UUID,
    current_task_id UUID,
    current_workspace VARCHAR(500),
    
    -- Health tracking
    last_health_check TIMESTAMP WITH TIME ZONE,
    health_status VARCHAR(20) DEFAULT 'unknown',
    health_details JSONB DEFAULT '{}',
    
    -- Metrics
    requests_processed INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,
    total_execution_time_ms BIGINT DEFAULT 0,
    
    -- Lifecycle timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    terminated_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_agent_type CHECK (agent_type IN (
        'aider', 'claude_code', 'codex', 'cline', 'openhands', 
        'kiro', 'continue', 'supermaven', 'cursor', 'windsurf',
        'augment', 'sourcegraph', 'codeium', 'tabnine', 'codegpt',
        'twin', 'devin', 'devika', 'swe_agent', 'gpt_pilot',
        'metamorph', 'junie', 'amazon_q', 'github_copilot', 'jetbrains_ai',
        'codegemma', 'starcoder', 'qwen_coder', 'mistral_code', 'gemini_assist',
        'codey', 'llama_code', 'deepseek_coder', 'wizardcoder', 'phind',
        'cody', 'cursor_sh', 'trae', 'blackbox', 'lovable',
        'v0', 'tempo', 'bolt', 'replit_agent', 'idx',
        'firebase_studio', 'cascade', 'helixagent'
    )),
    CONSTRAINT valid_status CHECK (status IN (
        'creating', 'idle', 'active', 'background', 'degraded',
        'recovering', 'terminating', 'terminated', 'failed'
    )),
    CONSTRAINT valid_health CHECK (health_status IN (
        'healthy', 'degraded', 'unhealthy', 'unknown'
    ))
);

-- Ensemble sessions table - tracks multi-instance ensemble executions
CREATE TABLE ensemble_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Strategy configuration
    strategy VARCHAR(50) NOT NULL,
    strategy_config JSONB DEFAULT '{}',
    
    -- Participants
    participant_types TEXT[] NOT NULL,
    primary_instance_id UUID,
    critique_instance_ids UUID[],
    verification_instance_ids UUID[],
    fallback_instance_ids UUID[],
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'creating',
    
    -- Context and results
    context JSONB DEFAULT '{}',
    task_definition JSONB NOT NULL,
    intermediate_results JSONB DEFAULT '{}',
    final_result JSONB,
    
    -- Consensus tracking
    consensus_reached BOOLEAN,
    confidence_score FLOAT,
    voting_results JSONB DEFAULT '{}',
    
    -- Performance metrics
    total_duration_ms BIGINT,
    tokens_consumed INTEGER DEFAULT 0,
    api_calls INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_strategy CHECK (strategy IN (
        'voting', 'debate', 'consensus', 'pipeline', 
        'parallel', 'sequential', 'expert_panel'
    )),
    CONSTRAINT valid_session_status CHECK (status IN (
        'creating', 'active', 'paused', 'completed', 
        'failed', 'cancelled'
    ))
);

-- Feature registry table - tracks ported CLI agent features
CREATE TABLE feature_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Feature identification
    feature_name VARCHAR(100) NOT NULL UNIQUE,
    feature_category VARCHAR(50) NOT NULL,
    feature_description TEXT,
    
    -- Source information
    source_agent VARCHAR(50) NOT NULL,
    source_file VARCHAR(500),
    source_url TEXT,
    
    -- Implementation details
    implementation_type VARCHAR(50) NOT NULL,
    internal_path VARCHAR(500),
    interface_definition JSONB DEFAULT '{}',
    
    -- Dependencies
    dependencies TEXT[],
    external_deps TEXT[],
    required_providers TEXT[],
    
    -- Status tracking
    status VARCHAR(20) DEFAULT 'planned',
    priority INTEGER DEFAULT 3,
    complexity VARCHAR(20),
    estimated_effort_hours INTEGER,
    
    -- Port tracking
    porting_notes TEXT,
    porting_challenges TEXT,
    test_coverage FLOAT DEFAULT 0.0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_category CHECK (feature_category IN (
        'core_llm', 'code_understanding', 'git_operations', 'project_management',
        'ui_ux', 'tool_integration', 'security', 'extensibility',
        'performance', 'collaboration', 'deployment', 'ai_features'
    )),
    CONSTRAINT valid_impl_type CHECK (implementation_type IN (
        'port', 'adapt', 'implement', 'wrap', 'bridge'
    )),
    CONSTRAINT valid_feature_status CHECK (status IN (
        'planned', 'in_progress', 'implemented', 'tested', 
        'deployed', 'deprecated'
    )),
    CONSTRAINT valid_complexity CHECK (complexity IN ('low', 'medium', 'high')),
    CONSTRAINT valid_priority CHECK (priority BETWEEN 1 AND 5)
);

-- ============================================
-- COMMUNICATION TABLES
-- ============================================

-- Agent communication log - tracks inter-agent messages
CREATE TABLE agent_communication_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Message metadata
    message_type VARCHAR(50) NOT NULL,
    session_id UUID NOT NULL,
    
    -- Sender/receiver
    sender_instance_id UUID,
    sender_type VARCHAR(50),
    receiver_instance_id UUID,
    receiver_type VARCHAR(50),
    broadcast BOOLEAN DEFAULT FALSE,
    
    -- Content
    payload JSONB NOT NULL,
    payload_size_bytes INTEGER,
    
    -- Status
    status VARCHAR(20) DEFAULT 'pending',
    error_message TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE,
    received_at TIMESTAMP WITH TIME ZONE,
    processed_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_message_type CHECK (message_type IN (
        'request', 'response', 'event', 'heartbeat', 
        'command', 'result', 'error'
    )),
    CONSTRAINT valid_comm_status CHECK (status IN (
        'pending', 'sent', 'received', 'processed', 'failed'
    ))
);

-- Event bus log - tracks system events
CREATE TABLE event_bus_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Event details
    event_type VARCHAR(100) NOT NULL,
    event_source VARCHAR(100) NOT NULL,
    session_id UUID,
    instance_id UUID,
    
    -- Content
    payload JSONB NOT NULL,
    
    -- Routing
    topic VARCHAR(100),
    priority INTEGER DEFAULT 3,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_event_priority CHECK (priority BETWEEN 1 AND 5)
);

-- ============================================
-- SYNCHRONIZATION TABLES
-- ============================================

-- Distributed locks table - for distributed locking
CREATE TABLE distributed_locks (
    name VARCHAR(255) PRIMARY KEY,
    owner VARCHAR(255) NOT NULL,
    acquired_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Context
    session_id UUID,
    instance_id UUID,
    lock_context JSONB DEFAULT '{}'
);

-- CRDT state table - for conflict-free replicated data types
CREATE TABLE crdt_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Identification
    crdt_type VARCHAR(50) NOT NULL,
    crdt_key VARCHAR(255) NOT NULL,
    session_id UUID,
    
    -- State
    state JSONB NOT NULL,
    vector_clock JSONB NOT NULL,
    
    -- Metadata
    instance_id UUID,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_crdt_type CHECK (crdt_type IN (
        'g_counter', 'pn_counter', 'g_set', 'or_set', 
        'lww_register', 'mv_register'
    )),
    CONSTRAINT unique_crdt_key UNIQUE (crdt_type, crdt_key, session_id)
);

-- ============================================
-- WORKER & TASK TABLES
-- ============================================

-- Background tasks table - for worker pool tasks
CREATE TABLE background_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Task definition
    task_type VARCHAR(50) NOT NULL,
    task_name VARCHAR(200),
    payload JSONB NOT NULL,
    priority INTEGER DEFAULT 3,
    
    -- Assignment
    assigned_instance_id UUID,
    
    -- Status
    status VARCHAR(20) DEFAULT 'pending',
    progress_percent INTEGER DEFAULT 0,
    result JSONB,
    error_message TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Retry logic
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    
    -- Constraints
    CONSTRAINT valid_task_type CHECK (task_type IN (
        'git_operation', 'code_analysis', 'documentation',
        'testing', 'linting', 'build', 'deploy',
        'code_review', 'refactoring', 'optimization'
    )),
    CONSTRAINT valid_task_status CHECK (status IN (
        'pending', 'assigned', 'running', 'completed', 
        'failed', 'cancelled', 'expired'
    )),
    CONSTRAINT valid_priority CHECK (priority BETWEEN 1 AND 5),
    CONSTRAINT valid_progress CHECK (progress_percent BETWEEN 0 AND 100)
);

-- ============================================
-- PERFORMANCE & MONITORING TABLES
-- ============================================

-- Instance metrics table - for performance tracking
CREATE TABLE instance_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL,
    
    -- Metrics
    metric_name VARCHAR(100) NOT NULL,
    metric_value FLOAT NOT NULL,
    metric_unit VARCHAR(50),
    
    -- Context
    labels JSONB DEFAULT '{}',
    
    -- Timestamp
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Cache statistics table - for cache performance
CREATE TABLE cache_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Cache identification
    cache_type VARCHAR(50) NOT NULL,
    cache_name VARCHAR(100) NOT NULL,
    
    -- Statistics
    hits BIGINT DEFAULT 0,
    misses BIGINT DEFAULT 0,
    evictions BIGINT DEFAULT 0,
    size_bytes BIGINT DEFAULT 0,
    entry_count INTEGER DEFAULT 0,
    
    -- Timestamp
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_cache_type CHECK (cache_type IN (
        'semantic', 'embedding', 'repo_map', 'provider', 'instance'
    ))
);

-- ============================================
-- INDEXES
-- ============================================

-- Agent instances indexes
CREATE INDEX idx_agent_instances_type ON agent_instances(agent_type);
CREATE INDEX idx_agent_instances_status ON agent_instances(status);
CREATE INDEX idx_agent_instances_health ON agent_instances(health_status);
CREATE INDEX idx_agent_instances_session ON agent_instances(current_session_id) 
    WHERE current_session_id IS NOT NULL;
CREATE INDEX idx_agent_instances_created ON agent_instances(created_at DESC);

-- Ensemble sessions indexes
CREATE INDEX idx_ensemble_sessions_status ON ensemble_sessions(status);
CREATE INDEX idx_ensemble_sessions_strategy ON ensemble_sessions(strategy);
CREATE INDEX idx_ensemble_sessions_created ON ensemble_sessions(created_at DESC);
CREATE INDEX idx_ensemble_sessions_primary ON ensemble_sessions(primary_instance_id);

-- Feature registry indexes
CREATE INDEX idx_feature_registry_category ON feature_registry(feature_category);
CREATE INDEX idx_feature_registry_source ON feature_registry(source_agent);
CREATE INDEX idx_feature_registry_status ON feature_registry(status);
CREATE INDEX idx_feature_registry_priority ON feature_registry(priority);

-- Communication log indexes
CREATE INDEX idx_comm_log_session ON agent_communication_log(session_id);
CREATE INDEX idx_comm_log_sender ON agent_communication_log(sender_instance_id);
CREATE INDEX idx_comm_log_receiver ON agent_communication_log(receiver_instance_id);
CREATE INDEX idx_comm_log_type ON agent_communication_log(message_type);
CREATE INDEX idx_comm_log_created ON agent_communication_log(created_at DESC);
CREATE INDEX idx_comm_log_status ON agent_communication_log(status);

-- Event bus indexes
CREATE INDEX idx_event_log_type ON event_bus_log(event_type);
CREATE INDEX idx_event_log_session ON event_bus_log(session_id);
CREATE INDEX idx_event_log_source ON event_bus_log(event_source);
CREATE INDEX idx_event_log_topic ON event_bus_log(topic);
CREATE INDEX idx_event_log_created ON event_bus_log(created_at DESC);

-- Distributed locks indexes
CREATE INDEX idx_locks_expires ON distributed_locks(expires_at);
CREATE INDEX idx_locks_owner ON distributed_locks(owner);

-- Background tasks indexes
CREATE INDEX idx_bg_tasks_status ON background_tasks(status);
CREATE INDEX idx_bg_tasks_type ON background_tasks(task_type);
CREATE INDEX idx_bg_tasks_assigned ON background_tasks(assigned_instance_id);
CREATE INDEX idx_bg_tasks_priority ON background_tasks(priority);
CREATE INDEX idx_bg_tasks_created ON background_tasks(created_at);

-- Metrics indexes
CREATE INDEX idx_metrics_instance ON instance_metrics(instance_id);
CREATE INDEX idx_metrics_name ON instance_metrics(metric_name);
CREATE INDEX idx_metrics_recorded ON instance_metrics(recorded_at DESC);

-- ============================================
-- VIEWS
-- ============================================

-- Active instances view
CREATE VIEW v_active_instances AS
SELECT 
    id,
    agent_type,
    instance_name,
    status,
    health_status,
    current_session_id,
    requests_processed,
    errors_count,
    created_at,
    EXTRACT(EPOCH FROM (NOW() - created_at))/3600 as uptime_hours
FROM agent_instances
WHERE status IN ('idle', 'active', 'background')
  AND (terminated_at IS NULL OR terminated_at > NOW() - INTERVAL '1 hour');

-- Ensemble session summary view
CREATE VIEW v_ensemble_summary AS
SELECT 
    s.id,
    s.strategy,
    s.status,
    s.consensus_reached,
    s.confidence_score,
    s.created_at,
    s.started_at,
    s.completed_at,
    EXTRACT(EPOCH FROM (COALESCE(s.completed_at, NOW()) - s.started_at))/1000 as duration_seconds,
    array_length(s.participant_types, 1) as participant_count
FROM ensemble_sessions s;

-- Feature implementation status view
CREATE VIEW v_feature_status AS
SELECT 
    feature_category,
    status,
    COUNT(*) as count,
    SUM(CASE WHEN priority = 1 THEN 1 ELSE 0 END) as p1_count,
    SUM(CASE WHEN priority = 2 THEN 1 ELSE 0 END) as p2_count,
    SUM(CASE WHEN priority = 3 THEN 1 ELSE 0 END) as p3_count,
    SUM(estimated_effort_hours) as total_effort_hours
FROM feature_registry
GROUP BY feature_category, status
ORDER BY feature_category, status;

-- ============================================
-- FUNCTIONS & TRIGGERS
-- ============================================

-- Update timestamp trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update triggers
CREATE TRIGGER update_agent_instances_updated_at 
    BEFORE UPDATE ON agent_instances 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ensemble_sessions_updated_at 
    BEFORE UPDATE ON ensemble_sessions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_feature_registry_updated_at 
    BEFORE UPDATE ON feature_registry 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_background_tasks_updated_at 
    BEFORE UPDATE ON background_tasks 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Clean expired locks function
CREATE OR REPLACE FUNCTION clean_expired_locks()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM distributed_locks WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- INITIAL DATA
-- ============================================

-- Populate feature registry with critical features
INSERT INTO feature_registry 
    (feature_name, feature_category, source_agent, implementation_type, priority, complexity, estimated_effort_hours, status, feature_description)
VALUES
    -- Foundation features (Priority 1)
    ('instance_management', 'core_llm', 'helixagent', 'implement', 1, 'high', 120, 'planned', 'Manages lifecycle of CLI agent instances'),
    ('ensemble_coordination', 'core_llm', 'helixagent', 'implement', 1, 'high', 140, 'planned', 'Coordinates multiple agent instances for ensemble execution'),
    ('distributed_synchronization', 'core_llm', 'helixagent', 'implement', 1, 'high', 100, 'planned', 'Distributed locks and CRDTs for state synchronization'),
    ('event_bus', 'core_llm', 'helixagent', 'implement', 1, 'medium', 60, 'planned', 'Event-driven communication between instances'),
    ('instance_pooling', 'performance', 'helixagent', 'implement', 1, 'medium', 80, 'planned', 'Pool pattern for efficient instance reuse'),
    
    -- Aider features (Priority 2)
    ('repo_map', 'code_understanding', 'aider', 'port', 2, 'high', 120, 'planned', 'AST-based repository understanding and symbol ranking'),
    ('diff_format', 'code_understanding', 'aider', 'port', 2, 'medium', 40, 'planned', 'SEARCH/REPLACE block editing format'),
    ('git_integration', 'git_operations', 'aider', 'port', 2, 'high', 80, 'planned', 'Deep git integration for code changes'),
    
    -- Claude Code features (Priority 2)
    ('terminal_ui', 'ui_ux', 'claude_code', 'adapt', 2, 'high', 100, 'planned', 'Rich terminal UI with syntax highlighting'),
    ('tool_use_framework', 'tool_integration', 'claude_code', 'adapt', 2, 'medium', 60, 'planned', 'LLM tool use and function calling framework'),
    
    -- Codex features (Priority 3)
    ('code_interpreter', 'ai_features', 'codex', 'adapt', 3, 'high', 80, 'planned', 'Code execution and interpretation'),
    ('reasoning_display', 'ui_ux', 'codex', 'adapt', 3, 'low', 20, 'planned', 'Display reasoning steps and chain-of-thought'),
    
    -- Cline features (Priority 3)
    ('browser_automation', 'tool_integration', 'cline', 'port', 3, 'high', 160, 'planned', 'Browser automation and web interaction'),
    ('computer_use', 'tool_integration', 'cline', 'port', 3, 'high', 200, 'planned', 'General computer control capabilities'),
    ('autonomy_framework', 'ai_features', 'cline', 'adapt', 3, 'high', 120, 'planned', 'Autonomous task execution framework'),
    
    -- OpenHands features (Priority 3-4)
    ('docker_sandbox', 'security', 'openhands', 'port', 3, 'high', 100, 'planned', 'Docker-based secure code execution'),
    ('security_isolation', 'security', 'openhands', 'adapt', 4, 'medium', 60, 'planned', 'Multi-layer security isolation'),
    
    -- Kiro features (Priority 4)
    ('project_memory', 'ai_features', 'kiro', 'adapt', 4, 'medium', 80, 'planned', 'Persistent project memory with embeddings'),
    
    -- Continue features (Priority 4)
    ('lsp_client', 'tool_integration', 'continue', 'adapt', 4, 'high', 100, 'planned', 'LSP client for IDE integration'),
    
    -- Output system (Priority 2)
    ('streaming_pipeline', 'performance', 'helixagent', 'implement', 2, 'medium', 80, 'planned', 'Optimized streaming output pipeline'),
    ('semantic_caching', 'performance', 'helixagent', 'implement', 3, 'high', 100, 'planned', 'Semantic similarity-based response caching'),
    ('background_workers', 'performance', 'helixagent', 'implement', 3, 'medium', 60, 'planned', 'Background task processing pool'),
    ('load_balancing', 'performance', 'helixagent', 'implement', 3, 'medium', 80, 'planned', 'Request distribution across instances')
ON CONFLICT (feature_name) DO NOTHING;

-- ============================================
-- COMMENTS
-- ============================================

COMMENT ON TABLE agent_instances IS 'Tracks running CLI agent instances with their configuration and state';
COMMENT ON TABLE ensemble_sessions IS 'Manages multi-instance ensemble execution sessions';
COMMENT ON TABLE feature_registry IS 'Registry of CLI agent features being ported to HelixAgent';
COMMENT ON TABLE agent_communication_log IS 'Audit log of inter-agent communication';
COMMENT ON TABLE event_bus_log IS 'Event bus message log for debugging and replay';
COMMENT ON TABLE distributed_locks IS 'Distributed locking mechanism for cluster coordination';
COMMENT ON TABLE crdt_state IS 'CRDT state storage for conflict-free replication';
COMMENT ON TABLE background_tasks IS 'Task queue for background worker processing';
COMMENT ON TABLE instance_metrics IS 'Performance metrics for agent instances';
COMMENT ON TABLE cache_statistics IS 'Cache performance statistics';

-- Migration complete
SELECT 'Migration 001_cli_agents_fusion.sql completed successfully' AS status;

-- ============================================
-- ADDITIONAL TABLES FOR COMPLETE IMPLEMENTATION
-- ============================================

-- Semantic cache table for caching LLM responses
CREATE TABLE semantic_cache (
    key VARCHAR(255) PRIMARY KEY,
    query TEXT NOT NULL,
    embedding vector(1536),  -- Adjust dimension based on embedding model
    response JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    hit_count INTEGER DEFAULT 0
);

-- Index for similarity search
CREATE INDEX idx_semantic_cache_embedding ON semantic_cache USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX idx_semantic_cache_expires ON semantic_cache(expires_at);

-- Git operations log
CREATE TABLE git_operations_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operation VARCHAR(50) NOT NULL,
    repository_path VARCHAR(500) NOT NULL,
    branch VARCHAR(100),
    commit_hash VARCHAR(40),
    files_changed TEXT[],
    execution_time_ms INTEGER,
    success BOOLEAN,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_git_ops_repo ON git_operations_log(repository_path);
CREATE INDEX idx_git_ops_created ON git_operations_log(created_at DESC);

-- LSP sessions table
CREATE TABLE lsp_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    language_server VARCHAR(100) NOT NULL,
    root_path VARCHAR(500) NOT NULL,
    process_id INTEGER,
    capabilities JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_lsp_sessions_status ON lsp_sessions(status);

-- Comments for new tables
COMMENT ON TABLE semantic_cache IS 'Semantic cache for LLM responses with embedding-based similarity search';
COMMENT ON TABLE git_operations_log IS 'Audit log of git operations performed by agents';
COMMENT ON TABLE lsp_sessions IS 'Active Language Server Protocol sessions';
