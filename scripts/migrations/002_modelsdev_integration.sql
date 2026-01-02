-- Models.dev Integration Migration
-- Adds tables for storing model metadata from Models.dev

-- Add new columns to llm_providers
ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS modelsdev_provider_id VARCHAR(255);
ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS total_models INTEGER DEFAULT 0;
ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS enabled_models INTEGER DEFAULT 0;
ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS last_models_sync TIMESTAMP WITH TIME ZONE;

-- Create models_metadata table
CREATE TABLE IF NOT EXISTS models_metadata (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_id VARCHAR(255) UNIQUE NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    provider_name VARCHAR(255) NOT NULL,

    -- Model details
    description TEXT,
    context_window INTEGER,
    max_tokens INTEGER,
    pricing_input DECIMAL(10, 6),
    pricing_output DECIMAL(10, 6),
    pricing_currency VARCHAR(10) DEFAULT 'USD',

    -- Capabilities
    supports_vision BOOLEAN DEFAULT FALSE,
    supports_function_calling BOOLEAN DEFAULT FALSE,
    supports_streaming BOOLEAN DEFAULT FALSE,
    supports_json_mode BOOLEAN DEFAULT FALSE,
    supports_image_generation BOOLEAN DEFAULT FALSE,
    supports_audio BOOLEAN DEFAULT FALSE,
    supports_code_generation BOOLEAN DEFAULT FALSE,
    supports_reasoning BOOLEAN DEFAULT FALSE,

    -- Performance metrics
    benchmark_score DECIMAL(5, 2),
    popularity_score INTEGER,
    reliability_score DECIMAL(5, 2),

    -- Categories and tags
    model_type VARCHAR(100),
    model_family VARCHAR(100),
    version VARCHAR(50),
    tags JSONB DEFAULT '[]',

    -- Models.dev specific
    modelsdev_url TEXT,
    modelsdev_id VARCHAR(255),
    modelsdev_api_version VARCHAR(50),

    -- Metadata
    raw_metadata JSONB DEFAULT '{}',
    last_refreshed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT fk_provider FOREIGN KEY (provider_id) REFERENCES llm_providers(id) ON DELETE CASCADE
);

-- Create model_benchmarks table
CREATE TABLE IF NOT EXISTS model_benchmarks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_id VARCHAR(255) NOT NULL,
    benchmark_name VARCHAR(255) NOT NULL,
    benchmark_type VARCHAR(100),
    score DECIMAL(10, 4),
    rank INTEGER,
    normalized_score DECIMAL(5, 2),
    benchmark_date DATE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT fk_model FOREIGN KEY (model_id) REFERENCES models_metadata(model_id) ON DELETE CASCADE,
    CONSTRAINT unique_model_benchmark UNIQUE (model_id, benchmark_name)
);

-- Create models_refresh_history table
CREATE TABLE IF NOT EXISTS models_refresh_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    refresh_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    models_refreshed INTEGER DEFAULT 0,
    models_failed INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    metadata JSONB DEFAULT '{}'
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_models_metadata_provider_id ON models_metadata(provider_id);
CREATE INDEX IF NOT EXISTS idx_models_metadata_model_type ON models_metadata(model_type);
CREATE INDEX IF NOT EXISTS idx_models_metadata_tags ON models_metadata USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_models_metadata_last_refreshed ON models_metadata(last_refreshed_at);
CREATE INDEX IF NOT EXISTS idx_models_metadata_model_family ON models_metadata(model_family);
CREATE INDEX IF NOT EXISTS idx_models_metadata_benchmark_score ON models_metadata(benchmark_score);

CREATE INDEX IF NOT EXISTS idx_benchmarks_model_id ON model_benchmarks(model_id);
CREATE INDEX IF NOT EXISTS idx_benchmarks_type ON model_benchmarks(benchmark_type);
CREATE INDEX IF NOT EXISTS idx_benchmarks_score ON model_benchmarks(score);

CREATE INDEX IF NOT EXISTS idx_refresh_history_started ON models_refresh_history(started_at);
CREATE INDEX IF NOT EXISTS idx_refresh_history_status ON models_refresh_history(status);

-- Create trigger for updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_models_metadata_updated_at
    BEFORE UPDATE ON models_metadata
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert initial data: Anthropic provider with Models.dev reference
INSERT INTO llm_providers (name, type, model, weight, enabled, config, modelsdev_provider_id)
VALUES (
    'anthropic',
    'anthropic',
    'claude-3-sonnet-20240229',
    1.0,
    TRUE,
    '{"models": ["claude-3-sonnet-20240229", "claude-3-opus-20240229", "claude-3-haiku-20240307"]}'::JSONB,
    'anthropic'
)
ON CONFLICT (name) DO UPDATE SET
    modelsdev_provider_id = EXCLUDED.modelsdev_provider_id;

-- Insert initial data: DeepSeek provider with Models.dev reference
INSERT INTO llm_providers (name, type, model, weight, enabled, config, modelsdev_provider_id)
VALUES (
    'deepseek',
    'deepseek',
    'deepseek-coder',
    1.0,
    TRUE,
    '{"models": ["deepseek-coder", "deepseek-chat"]}'::JSONB,
    'deepseek'
)
ON CONFLICT (name) DO UPDATE SET
    modelsdev_provider_id = EXCLUDED.modelsdev_provider_id;

-- Insert initial data: Google provider with Models.dev reference
INSERT INTO llm_providers (name, type, model, weight, enabled, config, modelsdev_provider_id)
VALUES (
    'google',
    'gemini',
    'gemini-pro',
    1.0,
    TRUE,
    '{"models": ["gemini-pro", "gemini-pro-vision"]}'::JSONB,
    'google'
)
ON CONFLICT (name) DO UPDATE SET
    modelsdev_provider_id = EXCLUDED.modelsdev_provider_id;
