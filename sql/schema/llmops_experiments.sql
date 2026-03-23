-- LLMOps Experiments Schema
-- Stores A/B experiments, evaluations, and prompt versions

CREATE TABLE IF NOT EXISTS llmops_experiments (
    id          VARCHAR(64) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    status      VARCHAR(32) NOT NULL DEFAULT 'created',
    variants    JSONB NOT NULL,
    metrics     JSONB,
    config      JSONB,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at  TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS llmops_evaluations (
    id          VARCHAR(64) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    dataset     VARCHAR(255) NOT NULL,
    status      VARCHAR(32) NOT NULL DEFAULT 'pending',
    metrics     JSONB,
    results     JSONB,
    config      JSONB,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS llmops_prompt_versions (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    version     VARCHAR(32) NOT NULL,
    content     TEXT NOT NULL,
    metadata    JSONB,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(name, version)
);

CREATE INDEX idx_experiments_status ON llmops_experiments(status);
CREATE INDEX idx_experiments_name ON llmops_experiments(name);
CREATE INDEX idx_evaluations_dataset ON llmops_evaluations(dataset);
CREATE INDEX idx_prompts_name ON llmops_prompt_versions(name);
