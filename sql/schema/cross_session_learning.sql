-- ============================================================================
-- Cross-Session Learning Schema
-- Stores learned patterns, insights, and knowledge accumulation
-- ============================================================================

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- Learned Patterns Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS learned_patterns (
    pattern_id VARCHAR(255) PRIMARY KEY,
    pattern_type VARCHAR(50) NOT NULL,  -- user_intent, debate_strategy, entity_cooccurrence, etc.
    description TEXT NOT NULL,
    frequency INT NOT NULL DEFAULT 1,
    confidence FLOAT NOT NULL,
    examples TEXT[],
    metadata JSONB,
    first_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for learned patterns
CREATE INDEX IF NOT EXISTS idx_learned_patterns_type ON learned_patterns(pattern_type);
CREATE INDEX IF NOT EXISTS idx_learned_patterns_frequency ON learned_patterns(frequency DESC);
CREATE INDEX IF NOT EXISTS idx_learned_patterns_confidence ON learned_patterns(confidence DESC);
CREATE INDEX IF NOT EXISTS idx_learned_patterns_last_seen ON learned_patterns(last_seen DESC);

-- GIN index for metadata
CREATE INDEX IF NOT EXISTS idx_learned_patterns_metadata ON learned_patterns USING GIN (metadata);

-- ============================================================================
-- Learned Insights Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS learned_insights (
    insight_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255),  -- NULL for global insights
    insight_type VARCHAR(50) NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    confidence FLOAT NOT NULL,
    impact VARCHAR(20) NOT NULL,  -- high, medium, low
    patterns JSONB NOT NULL,  -- Array of pattern objects
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for learned insights
CREATE INDEX IF NOT EXISTS idx_learned_insights_user ON learned_insights(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_learned_insights_type ON learned_insights(insight_type);
CREATE INDEX IF NOT EXISTS idx_learned_insights_impact ON learned_insights(impact);
CREATE INDEX IF NOT EXISTS idx_learned_insights_confidence ON learned_insights(confidence DESC);
CREATE INDEX IF NOT EXISTS idx_learned_insights_created ON learned_insights(created_at DESC);

-- GIN indexes for JSONB
CREATE INDEX IF NOT EXISTS idx_learned_insights_patterns ON learned_insights USING GIN (patterns);
CREATE INDEX IF NOT EXISTS idx_learned_insights_metadata ON learned_insights USING GIN (metadata);

-- ============================================================================
-- User Preferences Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    preference_type VARCHAR(50) NOT NULL,  -- communication_style, topic_interest, etc.
    preference_value TEXT NOT NULL,
    confidence FLOAT NOT NULL,
    frequency INT NOT NULL DEFAULT 1,
    last_observed TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(user_id, preference_type)
);

-- Indexes for user preferences
CREATE INDEX IF NOT EXISTS idx_user_preferences_user ON user_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_user_preferences_type ON user_preferences(preference_type);
CREATE INDEX IF NOT EXISTS idx_user_preferences_confidence ON user_preferences(confidence DESC);

-- ============================================================================
-- Entity Cooccurrence Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS entity_cooccurrences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity1_id VARCHAR(255) NOT NULL,
    entity1_type VARCHAR(50) NOT NULL,
    entity2_id VARCHAR(255) NOT NULL,
    entity2_type VARCHAR(50) NOT NULL,
    cooccurrence_count INT NOT NULL DEFAULT 1,
    confidence FLOAT NOT NULL,
    contexts TEXT[],  -- Conversation IDs where they co-occurred
    first_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(entity1_id, entity2_id)
);

-- Indexes for entity cooccurrences
CREATE INDEX IF NOT EXISTS idx_entity_cooccurrences_entity1 ON entity_cooccurrences(entity1_id);
CREATE INDEX IF NOT EXISTS idx_entity_cooccurrences_entity2 ON entity_cooccurrences(entity2_id);
CREATE INDEX IF NOT EXISTS idx_entity_cooccurrences_count ON entity_cooccurrences(cooccurrence_count DESC);
CREATE INDEX IF NOT EXISTS idx_entity_cooccurrences_types ON entity_cooccurrences(entity1_type, entity2_type);

-- ============================================================================
-- Debate Strategy Success Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS debate_strategy_success (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    strategy_name VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100),
    position VARCHAR(50),
    success_count INT NOT NULL DEFAULT 1,
    total_attempts INT NOT NULL DEFAULT 1,
    success_rate FLOAT NOT NULL,
    avg_confidence FLOAT NOT NULL,
    avg_response_time_ms FLOAT,
    last_used TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(strategy_name, provider, position)
);

-- Indexes for debate strategy success
CREATE INDEX IF NOT EXISTS idx_debate_strategy_provider ON debate_strategy_success(provider);
CREATE INDEX IF NOT EXISTS idx_debate_strategy_success_rate ON debate_strategy_success(success_rate DESC);
CREATE INDEX IF NOT EXISTS idx_debate_strategy_confidence ON debate_strategy_success(avg_confidence DESC);

-- ============================================================================
-- Conversation Flow Patterns Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS conversation_flow_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flow_type VARCHAR(50) NOT NULL,  -- rapid, normal, thoughtful
    avg_time_per_message_ms FLOAT NOT NULL,
    message_count_range VARCHAR(50),  -- "1-5", "6-10", "11-20", "21+"
    frequency INT NOT NULL DEFAULT 1,
    success_rate FLOAT,  -- Based on conversation outcome
    avg_satisfaction FLOAT,  -- If available
    metadata JSONB,
    first_seen TIMESTAMP NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(flow_type, message_count_range)
);

-- Indexes for conversation flow patterns
CREATE INDEX IF NOT EXISTS idx_conversation_flow_type ON conversation_flow_patterns(flow_type);
CREATE INDEX IF NOT EXISTS idx_conversation_flow_frequency ON conversation_flow_patterns(frequency DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_flow_success ON conversation_flow_patterns(success_rate DESC);

-- ============================================================================
-- Knowledge Accumulation Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS knowledge_accumulation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_type VARCHAR(50) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    fact TEXT NOT NULL,
    sources TEXT[],  -- Conversation IDs
    confidence FLOAT NOT NULL,
    verification_count INT NOT NULL DEFAULT 1,
    first_learned TIMESTAMP NOT NULL DEFAULT NOW(),
    last_verified TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(knowledge_type, subject, fact)
);

-- Indexes for knowledge accumulation
CREATE INDEX IF NOT EXISTS idx_knowledge_type ON knowledge_accumulation(knowledge_type);
CREATE INDEX IF NOT EXISTS idx_knowledge_subject ON knowledge_accumulation(subject);
CREATE INDEX IF NOT EXISTS idx_knowledge_confidence ON knowledge_accumulation(confidence DESC);
CREATE INDEX IF NOT EXISTS idx_knowledge_verification ON knowledge_accumulation(verification_count DESC);

-- Full-text search index for facts
CREATE INDEX IF NOT EXISTS idx_knowledge_fact_fts ON knowledge_accumulation USING GIN (to_tsvector('english', fact));

-- ============================================================================
-- Learning Statistics Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS learning_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stat_type VARCHAR(50) NOT NULL,
    stat_name VARCHAR(100) NOT NULL,
    stat_value FLOAT NOT NULL,
    aggregation_period VARCHAR(20),  -- hourly, daily, weekly, all_time
    period_start TIMESTAMP,
    period_end TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(stat_type, stat_name, aggregation_period, period_start)
);

-- Indexes for learning statistics
CREATE INDEX IF NOT EXISTS idx_learning_stats_type ON learning_statistics(stat_type);
CREATE INDEX IF NOT EXISTS idx_learning_stats_period ON learning_statistics(aggregation_period, period_start DESC);

-- ============================================================================
-- Helper Functions
-- ============================================================================

-- Function: Get top patterns by frequency
CREATE OR REPLACE FUNCTION get_top_patterns(pattern_type_filter VARCHAR DEFAULT NULL, limit_count INT DEFAULT 10)
RETURNS TABLE (
    pattern_id VARCHAR,
    pattern_type VARCHAR,
    description TEXT,
    frequency INT,
    confidence FLOAT,
    last_seen TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        lp.pattern_id,
        lp.pattern_type,
        lp.description,
        lp.frequency,
        lp.confidence,
        lp.last_seen
    FROM learned_patterns lp
    WHERE (pattern_type_filter IS NULL OR lp.pattern_type = pattern_type_filter)
    ORDER BY lp.frequency DESC, lp.confidence DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Function: Get user preferences summary
CREATE OR REPLACE FUNCTION get_user_preferences_summary(p_user_id VARCHAR)
RETURNS TABLE (
    preference_type VARCHAR,
    preference_value TEXT,
    confidence FLOAT,
    frequency INT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        up.preference_type,
        up.preference_value,
        up.confidence,
        up.frequency
    FROM user_preferences up
    WHERE up.user_id = p_user_id
    ORDER BY up.confidence DESC;
END;
$$ LANGUAGE plpgsql;

-- Function: Get entity cooccurrence network
CREATE OR REPLACE FUNCTION get_entity_cooccurrence_network(entity_id_filter VARCHAR, min_count INT DEFAULT 2)
RETURNS TABLE (
    entity1_id VARCHAR,
    entity1_type VARCHAR,
    entity2_id VARCHAR,
    entity2_type VARCHAR,
    cooccurrence_count INT,
    confidence FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ec.entity1_id,
        ec.entity1_type,
        ec.entity2_id,
        ec.entity2_type,
        ec.cooccurrence_count,
        ec.confidence
    FROM entity_cooccurrences ec
    WHERE (entity_id_filter IS NULL OR ec.entity1_id = entity_id_filter OR ec.entity2_id = entity_id_filter)
      AND ec.cooccurrence_count >= min_count
    ORDER BY ec.cooccurrence_count DESC;
END;
$$ LANGUAGE plpgsql;

-- Function: Get best debate strategies
CREATE OR REPLACE FUNCTION get_best_debate_strategies(position_filter VARCHAR DEFAULT NULL, limit_count INT DEFAULT 5)
RETURNS TABLE (
    strategy_name VARCHAR,
    provider VARCHAR,
    position VARCHAR,
    success_rate FLOAT,
    avg_confidence FLOAT,
    total_attempts INT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        dss.strategy_name,
        dss.provider,
        dss.position,
        dss.success_rate,
        dss.avg_confidence,
        dss.total_attempts
    FROM debate_strategy_success dss
    WHERE (position_filter IS NULL OR dss.position = position_filter)
      AND dss.total_attempts >= 3  -- Minimum sample size
    ORDER BY dss.success_rate DESC, dss.avg_confidence DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Function: Get learning progress over time
CREATE OR REPLACE FUNCTION get_learning_progress(days_back INT DEFAULT 30)
RETURNS TABLE (
    date DATE,
    patterns_learned INT,
    insights_generated INT,
    knowledge_accumulated INT
) AS $$
BEGIN
    RETURN QUERY
    WITH date_range AS (
        SELECT generate_series(
            CURRENT_DATE - (days_back || ' days')::INTERVAL,
            CURRENT_DATE,
            '1 day'::INTERVAL
        )::DATE AS date
    )
    SELECT
        dr.date,
        COALESCE(p.pattern_count, 0)::INT AS patterns_learned,
        COALESCE(i.insight_count, 0)::INT AS insights_generated,
        COALESCE(k.knowledge_count, 0)::INT AS knowledge_accumulated
    FROM date_range dr
    LEFT JOIN (
        SELECT DATE(created_at) AS date, COUNT(*) AS pattern_count
        FROM learned_patterns
        WHERE created_at >= CURRENT_DATE - (days_back || ' days')::INTERVAL
        GROUP BY DATE(created_at)
    ) p ON dr.date = p.date
    LEFT JOIN (
        SELECT DATE(created_at) AS date, COUNT(*) AS insight_count
        FROM learned_insights
        WHERE created_at >= CURRENT_DATE - (days_back || ' days')::INTERVAL
        GROUP BY DATE(created_at)
    ) i ON dr.date = i.date
    LEFT JOIN (
        SELECT DATE(created_at) AS date, COUNT(*) AS knowledge_count
        FROM knowledge_accumulation
        WHERE created_at >= CURRENT_DATE - (days_back || ' days')::INTERVAL
        GROUP BY DATE(created_at)
    ) k ON dr.date = k.date
    ORDER BY dr.date;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Triggers for Automatic Updates
-- ============================================================================

-- Trigger: Update pattern last_seen and updated_at
CREATE OR REPLACE FUNCTION update_pattern_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_seen = NOW();
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_pattern_timestamp
BEFORE UPDATE ON learned_patterns
FOR EACH ROW
EXECUTE FUNCTION update_pattern_timestamp();

-- Trigger: Update user preferences timestamp
CREATE OR REPLACE FUNCTION update_user_preference_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_observed = NOW();
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_user_preference_timestamp
BEFORE UPDATE ON user_preferences
FOR EACH ROW
EXECUTE FUNCTION update_user_preference_timestamp();

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE learned_patterns IS 'Learned patterns from cross-session analysis';
COMMENT ON TABLE learned_insights IS 'Generated insights from pattern analysis';
COMMENT ON TABLE user_preferences IS 'User-specific preference patterns';
COMMENT ON TABLE entity_cooccurrences IS 'Entity co-occurrence tracking for relationship discovery';
COMMENT ON TABLE debate_strategy_success IS 'Success rates of different debate strategies';
COMMENT ON TABLE conversation_flow_patterns IS 'Common conversation flow patterns';
COMMENT ON TABLE knowledge_accumulation IS 'Accumulated knowledge from conversations';
COMMENT ON TABLE learning_statistics IS 'Statistics about the learning process';

COMMENT ON COLUMN learned_patterns.frequency IS 'Number of times this pattern has been observed';
COMMENT ON COLUMN learned_insights.impact IS 'Impact level: high, medium, or low';
COMMENT ON COLUMN debate_strategy_success.success_rate IS 'Ratio of successful debates using this strategy';
COMMENT ON COLUMN knowledge_accumulation.verification_count IS 'Number of times this fact has been verified';
