-- Planning Sessions Schema
-- Stores HiPlan, MCTS, and Tree of Thoughts planning results

CREATE TABLE IF NOT EXISTS planning_sessions (
    id              VARCHAR(64) PRIMARY KEY,
    algorithm       VARCHAR(32) NOT NULL, -- 'hiplan', 'mcts', 'tot'
    status          VARCHAR(32) NOT NULL DEFAULT 'pending',
    input           JSONB NOT NULL,
    config          JSONB,
    result          JSONB,
    error           TEXT,
    execution_time_ms BIGINT DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at    TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS planning_hiplan_milestones (
    id          SERIAL PRIMARY KEY,
    session_id  VARCHAR(64) NOT NULL REFERENCES planning_sessions(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    order_index INTEGER NOT NULL,
    status      VARCHAR(32) NOT NULL DEFAULT 'pending',
    steps       JSONB
);

CREATE TABLE IF NOT EXISTS planning_mcts_nodes (
    id          SERIAL PRIMARY KEY,
    session_id  VARCHAR(64) NOT NULL REFERENCES planning_sessions(id) ON DELETE CASCADE,
    parent_id   INTEGER REFERENCES planning_mcts_nodes(id),
    action      VARCHAR(255),
    state       JSONB,
    visits      INTEGER DEFAULT 0,
    reward      DOUBLE PRECISION DEFAULT 0,
    depth       INTEGER DEFAULT 0
);

CREATE INDEX idx_planning_algorithm ON planning_sessions(algorithm);
CREATE INDEX idx_planning_status ON planning_sessions(status);
CREATE INDEX idx_planning_created ON planning_sessions(created_at DESC);
CREATE INDEX idx_milestones_session ON planning_hiplan_milestones(session_id);
CREATE INDEX idx_mcts_nodes_session ON planning_mcts_nodes(session_id);
