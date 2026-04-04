-- CLI Agents Fusion Schema
-- Phase 1: Foundation Layer
-- Date: 2026-04-04

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgvector";

-- ============================================
-- CLI AGENT INSTANCES
-- ============================================

CREATE TABLE cli_agent_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL, -- aider, claude_code, codex, cline, openhands, kiro, continue
    status VARCHAR(20) NOT NULL DEFAULT 'idle',
    config JSONB NOT NULL DEFAULT '{}',
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_cli_agent_instances_session ON cli_agent_instances(session_id);
CREATE INDEX idx_cli_agent_instances_user ON cli_agent_instances(user_id);
CREATE INDEX idx_cli_agent_instances_type ON cli_agent_instances(type);
CREATE INDEX idx_cli_agent_instances_status ON cli_agent_instances(status);

-- ============================================
-- CLI AGENT TASKS
-- ============================================

CREATE TABLE cli_agent_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES cli_agent_instances(id) ON DELETE CASCADE,
    parent_task_id UUID REFERENCES cli_agent_tasks(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL, -- repo_map, diff_apply, git_commit, tool_use, browser_action, etc.
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    priority INTEGER DEFAULT 5, -- 1-10, lower is higher priority
    input JSONB NOT NULL DEFAULT '{}',
    output JSONB,
    error TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    duration_ms INTEGER,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3
);

CREATE INDEX idx_cli_agent_tasks_instance ON cli_agent_tasks(instance_id);
CREATE INDEX idx_cli_agent_tasks_status ON cli_agent_tasks(status);
CREATE INDEX idx_cli_agent_tasks_parent ON cli_agent_tasks(parent_task_id);
CREATE INDEX idx_cli_agent_tasks_created ON cli_agent_tasks(created_at);

-- ============================================
-- REPO MAP DATA
-- ============================================

CREATE TABLE repo_maps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES cli_agent_instances(id) ON DELETE CASCADE,
    repository_path TEXT NOT NULL,
    git_commit_hash VARCHAR(40),
    symbols JSONB NOT NULL DEFAULT '[]', -- Array of symbols with metadata
    file_structure JSONB NOT NULL DEFAULT '{}',
    map_tokens INTEGER DEFAULT 1024,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_accessed_at TIMESTAMP WITH TIME ZONE,
    access_count INTEGER DEFAULT 0
);

CREATE INDEX idx_repo_maps_instance ON repo_maps(instance_id);
CREATE INDEX idx_repo_maps_path ON repo_maps(repository_path);
CREATE INDEX idx_repo_maps_commit ON repo_maps(git_commit_hash);

-- ============================================
-- REPO SYMBOLS (for efficient searching)
-- ============================================

CREATE TABLE repo_symbols (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_map_id UUID REFERENCES repo_maps(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- function, class, variable, method, etc.
    file_path TEXT NOT NULL,
    line_start INTEGER,
    line_end INTEGER,
    signature TEXT,
    documentation TEXT,
    embedding VECTOR(1536), -- For semantic search
    relevance_score FLOAT,
    references_count INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_repo_symbols_repo ON repo_symbols(repo_map_id);
CREATE INDEX idx_repo_symbols_name ON repo_symbols(name);
CREATE INDEX idx_repo_symbols_type ON repo_symbols(type);
CREATE INDEX idx_repo_symbols_file ON repo_symbols(file_path);
CREATE INDEX idx_repo_symbols_embedding ON repo_symbols USING ivfflat (embedding vector_cosine_ops);

-- ============================================
-- GIT OPERATIONS LOG
-- ============================================

CREATE TABLE git_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES cli_agent_instances(id) ON DELETE CASCADE,
    operation VARCHAR(50) NOT NULL, -- commit, diff, merge, branch, etc.
    repository_path TEXT NOT NULL,
    commit_hash VARCHAR(40),
    parent_hash VARCHAR(40),
    message TEXT,
    author_name VARCHAR(255),
    author_email VARCHAR(255),
    files_changed JSONB DEFAULT '[]',
    diff_content TEXT,
    attribution VARCHAR(255), -- HelixAgent or user
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_git_operations_instance ON git_operations(instance_id);
CREATE INDEX idx_git_operations_repo ON git_operations(repository_path);
CREATE INDEX idx_git_operations_commit ON git_operations(commit_hash);
CREATE INDEX idx_git_operations_created ON git_operations(created_at);

-- ============================================
-- DIFF APPLICATIONS
-- ============================================

CREATE TABLE diff_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES cli_agent_instances(id) ON DELETE CASCADE,
    task_id UUID REFERENCES cli_agent_tasks(id) ON DELETE SET NULL,
    file_path TEXT NOT NULL,
    original_content TEXT,
    modified_content TEXT,
    diff_content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, applied, failed, reverted
    search_block TEXT,
    replace_block TEXT,
    error_message TEXT,
    applied_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_diff_applications_instance ON diff_applications(instance_id);
CREATE INDEX idx_diff_applications_task ON diff_applications(task_id);
CREATE INDEX idx_diff_applications_file ON diff_applications(file_path);
CREATE INDEX idx_diff_applications_status ON diff_applications(status);

-- ============================================
-- TERMINAL UI SESSIONS
-- ============================================

CREATE TABLE terminal_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES cli_agent_instances(id) ON DELETE CASCADE,
    session_type VARCHAR(50) NOT NULL, -- claude_code, aider, etc.
    width INTEGER DEFAULT 80,
    height INTEGER DEFAULT 24,
    content JSONB DEFAULT '[]', -- Array of content blocks
    scrollback JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_terminal_sessions_instance ON terminal_sessions(instance_id);

-- ============================================
-- TOOL USE LOG
-- ============================================

CREATE TABLE tool_use_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES cli_agent_instances(id) ON DELETE CASCADE,
    task_id UUID REFERENCES cli_agent_tasks(id) ON DELETE SET NULL,
    tool_name VARCHAR(100) NOT NULL,
    arguments JSONB NOT NULL,
    result JSONB,
    error TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'running',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    required_approval BOOLEAN DEFAULT false,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tool_use_log_instance ON tool_use_log(instance_id);
CREATE INDEX idx_tool_use_log_task ON tool_use_log(task_id);
CREATE INDEX idx_tool_use_log_tool ON tool_use_log(tool_name);
CREATE INDEX idx_tool_use_log_status ON tool_use_log(status);

-- ============================================
-- PROJECT MEMORY
-- ============================================

CREATE TABLE project_memory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    instance_id UUID REFERENCES cli_agent_instances(id) ON DELETE CASCADE,
    entry_type VARCHAR(50) NOT NULL, -- code, conversation, decision, error, insight
    content TEXT NOT NULL,
    embedding VECTOR(1536),
    metadata JSONB DEFAULT '{}',
    importance_score FLOAT DEFAULT 0.5,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    accessed_at TIMESTAMP WITH TIME ZONE,
    access_count INTEGER DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_project_memory_project ON project_memory(project_id);
CREATE INDEX idx_project_memory_instance ON project_memory(instance_id);
CREATE INDEX idx_project_memory_type ON project_memory(entry_type);
CREATE INDEX idx_project_memory_embedding ON project_memory USING ivfflat (embedding vector_cosine_ops);

-- ============================================
-- BROWSER SESSIONS
-- ============================================

CREATE TABLE browser_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES cli_agent_instances(id) ON DELETE CASCADE,
    url TEXT,
    page_title TEXT,
    screenshot BYTEA,
    accessible_tree JSONB,
    actions JSONB DEFAULT '[]',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_browser_sessions_instance ON browser_sessions(instance_id);

-- ============================================
-- SANDBOX ENVIRONMENTS
-- ============================================

CREATE TABLE sandbox_environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES cli_agent_instances(id) ON DELETE CASCADE,
    container_id VARCHAR(255),
    image VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'creating', -- creating, running, stopped, error
    resources JSONB DEFAULT '{}', -- {memory: 512, cpu: 1.0}
    volumes JSONB DEFAULT '[]',
    network_enabled BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    stopped_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_sandbox_env_instance ON sandbox_environments(instance_id);
CREATE INDEX idx_sandbox_env_container ON sandbox_environments(container_id);

-- ============================================
-- PLANNING DATA
-- ============================================

CREATE TABLE task_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES cli_agent_instances(id) ON DELETE CASCADE,
    objective TEXT NOT NULL,
    context JSONB DEFAULT '[]',
    status VARCHAR(20) DEFAULT 'pending', -- pending, running, completed, failed
    tasks JSONB NOT NULL DEFAULT '[]',
    dependencies JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_task_plans_instance ON task_plans(instance_id);
CREATE INDEX idx_task_plans_status ON task_plans(status);

-- ============================================
-- TRIGGERS FOR UPDATED_AT
-- ============================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_cli_agent_instances_updated_at BEFORE UPDATE ON cli_agent_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cli_agent_tasks_updated_at BEFORE UPDATE ON cli_agent_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_repo_maps_updated_at BEFORE UPDATE ON repo_maps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- VIEWS FOR COMMON QUERIES
-- ============================================

-- Active instances with their latest tasks
CREATE VIEW active_cli_instances AS
SELECT 
    i.*,
    COUNT(t.id) FILTER (WHERE t.status = 'running') as running_tasks,
    COUNT(t.id) FILTER (WHERE t.status = 'pending') as pending_tasks,
    MAX(t.created_at) as last_task_at
FROM cli_agent_instances i
LEFT JOIN cli_agent_tasks t ON t.instance_id = i.id
WHERE i.status IN ('idle', 'running')
GROUP BY i.id;

-- Recent git operations with stats
CREATE VIEW recent_git_operations AS
SELECT 
    g.*,
    jsonb_array_length(g.files_changed) as files_changed_count
FROM git_operations g
WHERE g.created_at > NOW() - INTERVAL '24 hours'
ORDER BY g.created_at DESC;

-- High importance project memories
CREATE VIEW important_project_memories AS
SELECT 
    m.*,
    1 - (NOW() - m.created_at) / INTERVAL '30 days' as recency_score
FROM project_memory m
WHERE m.importance_score > 0.7
ORDER BY m.importance_score * recency_score DESC;
