-- =============================================================================
-- HelixAgent SQL Schema: Code Versions
-- =============================================================================
-- Domain: Snapshots of code at debate milestones for version tracking.
--
-- Code versions capture the evolution of solutions through debate rounds.
-- Each version records the code, quality metrics, test pass rates, and
-- diffs from the previous version. This enables solution comparison,
-- rollback capability, and quality trend analysis across debate rounds.
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table: code_versions
-- -----------------------------------------------------------------------------
-- Stores code snapshots at key points during debate sessions.
-- Each version is linked to a debate session and optionally to a specific turn.
-- Version numbers are sequential within a session (enforced by unique constraint).
--
-- Primary Key: id (UUID, auto-generated)
-- Foreign Keys: session_id -> debate_sessions(id), turn_id -> debate_turns(id)
-- Unique: (session_id, version_number) — one version number per session
-- Indexes support: session lookup, quality filtering, language analytics
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS code_versions (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id         UUID         NOT NULL REFERENCES debate_sessions(id) ON DELETE CASCADE,
    turn_id            UUID         REFERENCES debate_turns(id) ON DELETE SET NULL,
    language           VARCHAR(50),                              -- Programming language
    code               TEXT         NOT NULL,                    -- Code snapshot
    version_number     INTEGER      NOT NULL,                    -- Sequential version within session
    quality_score      DECIMAL(5,4),                              -- Overall quality score (0.0000-1.0000)
    test_pass_rate     DECIMAL(5,4),                              -- Percentage of tests passing (0.0000-1.0000)
    metrics            JSONB        DEFAULT '{}',                -- Maintainability, complexity, security scores
    diff_from_previous TEXT,                                      -- Diff from prior version
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Unique constraint: one version number per session
ALTER TABLE code_versions
    ADD CONSTRAINT uq_code_versions_session_version
    UNIQUE (session_id, version_number);

-- Column comments
COMMENT ON TABLE code_versions IS 'Stores code snapshots at debate milestones for version tracking and comparison';
COMMENT ON COLUMN code_versions.id IS 'Unique version identifier (UUID)';
COMMENT ON COLUMN code_versions.session_id IS 'References the parent debate session';
COMMENT ON COLUMN code_versions.turn_id IS 'References the specific debate turn that produced this version';
COMMENT ON COLUMN code_versions.language IS 'Programming language of the code snapshot';
COMMENT ON COLUMN code_versions.code IS 'Full code snapshot at this version';
COMMENT ON COLUMN code_versions.version_number IS 'Sequential version number within the session';
COMMENT ON COLUMN code_versions.quality_score IS 'Overall quality score (0.0-1.0)';
COMMENT ON COLUMN code_versions.test_pass_rate IS 'Percentage of tests passing (0.0-1.0)';
COMMENT ON COLUMN code_versions.metrics IS 'Detailed quality metrics as JSONB (maintainability, complexity, security)';
COMMENT ON COLUMN code_versions.diff_from_previous IS 'Unified diff from the previous version';

-- Session lookup
CREATE INDEX IF NOT EXISTS idx_code_versions_session_id
    ON code_versions(session_id);

-- Turn lookup
CREATE INDEX IF NOT EXISTS idx_code_versions_turn_id
    ON code_versions(turn_id);

-- Session + version ordering (most common query: get versions for a session)
CREATE INDEX IF NOT EXISTS idx_code_versions_session_version
    ON code_versions(session_id, version_number);

-- Language analytics
CREATE INDEX IF NOT EXISTS idx_code_versions_language
    ON code_versions(language);

-- Quality filtering (find high-quality versions)
CREATE INDEX IF NOT EXISTS idx_code_versions_quality
    ON code_versions(quality_score)
    WHERE quality_score IS NOT NULL;

-- Test pass rate filtering
CREATE INDEX IF NOT EXISTS idx_code_versions_test_pass_rate
    ON code_versions(test_pass_rate)
    WHERE test_pass_rate IS NOT NULL;

-- GIN index for metrics JSONB queries
CREATE INDEX IF NOT EXISTS idx_code_versions_metrics
    ON code_versions USING GIN (metrics);
