-- Agentic Workflows Schema
-- Stores workflow definitions, execution state, and results

CREATE TABLE IF NOT EXISTS agentic_workflows (
    id          VARCHAR(64) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    status      VARCHAR(32) NOT NULL DEFAULT 'pending',
    entry_point VARCHAR(64) NOT NULL,
    config      JSONB,
    input       JSONB,
    result      JSONB,
    error       TEXT,
    nodes_executed  INTEGER DEFAULT 0,
    execution_time_ms BIGINT DEFAULT 0,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS agentic_workflow_nodes (
    id          VARCHAR(64) PRIMARY KEY,
    workflow_id VARCHAR(64) NOT NULL REFERENCES agentic_workflows(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    type        VARCHAR(32) NOT NULL,
    config      JSONB,
    status      VARCHAR(32) NOT NULL DEFAULT 'pending',
    result      JSONB,
    error       TEXT,
    started_at  TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    execution_time_ms BIGINT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS agentic_workflow_edges (
    id          SERIAL PRIMARY KEY,
    workflow_id VARCHAR(64) NOT NULL REFERENCES agentic_workflows(id) ON DELETE CASCADE,
    from_node   VARCHAR(64) NOT NULL,
    to_node     VARCHAR(64) NOT NULL,
    condition   JSONB
);

CREATE INDEX idx_workflows_status ON agentic_workflows(status);
CREATE INDEX idx_workflows_created ON agentic_workflows(created_at DESC);
CREATE INDEX idx_workflow_nodes_wf ON agentic_workflow_nodes(workflow_id);
CREATE INDEX idx_workflow_edges_wf ON agentic_workflow_edges(workflow_id);
