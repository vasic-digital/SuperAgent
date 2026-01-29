-- =============================================================================
-- HelixAgent SQL Schema: Background Task System
-- =============================================================================
-- Domain: Asynchronous task execution, dead-letter queue, monitoring, webhooks.
-- Source migrations: 011_background_tasks.sql
--
-- The background task system uses PostgreSQL as a durable task queue with
-- atomic dequeue (SELECT ... FOR UPDATE SKIP LOCKED), priority scheduling,
-- retry logic, stuck detection, and resource monitoring.
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Custom ENUM Types
-- -----------------------------------------------------------------------------

-- Task lifecycle states
-- Flow: pending -> queued -> running -> completed | failed | stuck | cancelled
-- Special: dead_letter (moved to dead-letter table after max retries)
DO $$ BEGIN
    CREATE TYPE task_status AS ENUM (
        'pending',      -- Task created, waiting in queue
        'queued',       -- Task assigned to worker pool
        'running',      -- Task currently executing
        'paused',       -- Task paused by user/system
        'completed',    -- Task finished successfully
        'failed',       -- Task failed with error
        'stuck',        -- Task detected as stuck (no heartbeat)
        'cancelled',    -- Task cancelled by user
        'dead_letter'   -- Task moved to dead-letter queue
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Task priority levels for queue ordering
DO $$ BEGIN
    CREATE TYPE task_priority AS ENUM (
        'critical',     -- P0: Execute immediately, skip queue
        'high',         -- P1: Execute next available slot
        'normal',       -- P2: Default priority
        'low',          -- P3: Execute when resources available
        'background'    -- P4: Execute during low-load periods
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- -----------------------------------------------------------------------------
-- Table: background_tasks
-- -----------------------------------------------------------------------------
-- Main table for the PostgreSQL-backed task queue. Supports hierarchical tasks
-- (parent_task_id), checkpoint/resume, resource requirements, and soft delete.
--
-- Primary Key: id (UUID, auto-generated)
-- Foreign Keys: parent_task_id -> background_tasks(id) ON DELETE SET NULL (self-ref)
-- Referenced by: task_execution_history, task_resource_snapshots, webhook_deliveries
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS background_tasks (
    id                  UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Task identification
    task_type           VARCHAR(100)  NOT NULL,                -- Category (e.g., 'llm_batch', 'model_sync', 'report')
    task_name           VARCHAR(255)  NOT NULL,                -- Human-readable task name
    correlation_id      VARCHAR(255),                           -- External correlation ID for tracing
    parent_task_id      UUID          REFERENCES background_tasks(id) ON DELETE SET NULL, -- Parent task (hierarchical)

    -- Task configuration
    payload             JSONB         NOT NULL DEFAULT '{}',    -- Task input data
    config              JSONB         NOT NULL DEFAULT '{}',    -- Execution configuration
    priority            task_priority NOT NULL DEFAULT 'normal', -- Queue priority

    -- State management
    status              task_status   NOT NULL DEFAULT 'pending', -- Current lifecycle state
    progress            DECIMAL(5,2)  DEFAULT 0.0,               -- Completion percentage (0.00-100.00)
    progress_message    TEXT,                                      -- Human-readable progress description
    checkpoint          JSONB,                                     -- Checkpoint data for resume after failure

    -- Retry configuration
    max_retries         INTEGER       DEFAULT 3,                  -- Maximum retry attempts
    retry_count         INTEGER       DEFAULT 0,                  -- Current retry count
    retry_delay_seconds INTEGER       DEFAULT 60,                 -- Delay between retries (seconds)
    last_error          TEXT,                                      -- Most recent error message
    error_history       JSONB         DEFAULT '[]',               -- Array of all error messages with timestamps

    -- Execution tracking
    worker_id           VARCHAR(100),                              -- Worker that claimed this task
    process_pid         INTEGER,                                   -- OS process ID of the worker
    started_at          TIMESTAMP WITH TIME ZONE,                  -- When execution started
    completed_at        TIMESTAMP WITH TIME ZONE,                  -- When execution completed
    last_heartbeat      TIMESTAMP WITH TIME ZONE,                  -- Last worker heartbeat (for stuck detection)
    deadline            TIMESTAMP WITH TIME ZONE,                  -- Hard deadline for task completion

    -- Resource requirements (used for scheduling decisions)
    required_cpu_cores          INTEGER DEFAULT 1,                 -- Minimum CPU cores needed
    required_memory_mb          INTEGER DEFAULT 512,               -- Minimum memory in MB
    estimated_duration_seconds  INTEGER,                            -- Estimated run time
    actual_duration_seconds     INTEGER,                            -- Actual run time (set on completion)

    -- Notification configuration
    notification_config JSONB         DEFAULT '{}',                -- Webhook/SSE/WebSocket notification settings

    -- User association
    user_id             UUID,                                       -- Requesting user (no FK for flexibility)
    session_id          UUID,                                       -- Requesting session (no FK for flexibility)
    tags                JSONB         DEFAULT '[]',                 -- Searchable tags array
    metadata            JSONB         DEFAULT '{}',                 -- Arbitrary metadata

    -- Timestamps
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    scheduled_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),    -- Earliest execution time

    -- Soft delete
    deleted_at          TIMESTAMP WITH TIME ZONE                    -- NULL = not deleted
);

-- Auto-update trigger for updated_at
CREATE OR REPLACE FUNCTION update_background_task_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_background_task_updated_at ON background_tasks;
CREATE TRIGGER trg_background_task_updated_at
    BEFORE UPDATE ON background_tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_background_task_updated_at();

-- Basic indexes (migration 011)
CREATE INDEX IF NOT EXISTS idx_tasks_status          ON background_tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority_status  ON background_tasks(priority, status, scheduled_at);
CREATE INDEX IF NOT EXISTS idx_tasks_worker           ON background_tasks(worker_id) WHERE status = 'running';
CREATE INDEX IF NOT EXISTS idx_tasks_user             ON background_tasks(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_correlation      ON background_tasks(correlation_id) WHERE correlation_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled        ON background_tasks(scheduled_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_tasks_heartbeat        ON background_tasks(last_heartbeat) WHERE status = 'running';
CREATE INDEX IF NOT EXISTS idx_tasks_deadline         ON background_tasks(deadline) WHERE deadline IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_type             ON background_tasks(task_type);
CREATE INDEX IF NOT EXISTS idx_tasks_created          ON background_tasks(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tasks_not_deleted      ON background_tasks(id) WHERE deleted_at IS NULL;

-- Performance indexes (migration 012)
-- Priority queue ordering (critical for dequeue_background_task function)
CREATE INDEX IF NOT EXISTS idx_tasks_queue_order
    ON background_tasks (priority, scheduled_at ASC, created_at ASC)
    WHERE status = 'pending' AND deleted_at IS NULL;

-- Stuck task detection (join with resource_snapshots)
CREATE INDEX IF NOT EXISTS idx_tasks_running_heartbeat
    ON background_tasks (last_heartbeat, started_at)
    WHERE status = 'running';

-- Task completion tracking
CREATE INDEX IF NOT EXISTS idx_tasks_completion
    ON background_tasks (completed_at DESC, status)
    WHERE status IN ('completed', 'failed');

COMMENT ON TABLE background_tasks IS 'Main table for background task management with PostgreSQL-backed queue';

-- -----------------------------------------------------------------------------
-- Table: background_tasks_dead_letter
-- -----------------------------------------------------------------------------
-- Dead-letter queue for tasks that exhausted all retry attempts. Stores the
-- full task snapshot for manual inspection and optional reprocessing.
--
-- Primary Key: id (UUID, auto-generated)
-- No foreign keys (task data is copied, not referenced, to survive task deletion)
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS background_tasks_dead_letter (
    id               UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    original_task_id UUID          NOT NULL,                   -- Original task ID (for reference only)
    task_data        JSONB         NOT NULL,                    -- Full task snapshot at time of failure
    failure_reason   TEXT          NOT NULL,                    -- Final failure reason
    failure_count    INTEGER       DEFAULT 1,                   -- Total failure count
    moved_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),   -- When moved to dead-letter
    reprocess_after  TIMESTAMP WITH TIME ZONE,                  -- Earliest reprocessing time (NULL = manual only)
    reprocessed      BOOLEAN       DEFAULT FALSE                -- Whether this has been reprocessed
);

-- Indexes (migration 011)
CREATE INDEX IF NOT EXISTS idx_dead_letter_reprocess ON background_tasks_dead_letter(reprocess_after) WHERE NOT reprocessed;
CREATE INDEX IF NOT EXISTS idx_dead_letter_original  ON background_tasks_dead_letter(original_task_id);

-- Performance index (migration 012)
CREATE INDEX IF NOT EXISTS idx_dead_letter_ready
    ON background_tasks_dead_letter (reprocess_after ASC)
    WHERE NOT reprocessed AND reprocess_after IS NOT NULL;

COMMENT ON TABLE background_tasks_dead_letter IS 'Dead-letter queue for tasks that failed after max retries';

-- -----------------------------------------------------------------------------
-- Table: task_execution_history
-- -----------------------------------------------------------------------------
-- Immutable audit trail of task state transitions and events. Every status
-- change, retry, checkpoint, and error is recorded here for debugging.
--
-- Primary Key: id (UUID, auto-generated)
-- Foreign Keys: task_id -> background_tasks(id) ON DELETE CASCADE
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS task_execution_history (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id     UUID         NOT NULL REFERENCES background_tasks(id) ON DELETE CASCADE,
    event_type  VARCHAR(50)  NOT NULL,                     -- e.g., 'created', 'started', 'completed', 'failed', 'retried', 'stuck'
    event_data  JSONB        DEFAULT '{}',                  -- Event-specific data (error details, progress, etc.)
    worker_id   VARCHAR(100),                                -- Worker that generated this event
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes (migration 011)
CREATE INDEX IF NOT EXISTS idx_task_history_task_id    ON task_execution_history(task_id);
CREATE INDEX IF NOT EXISTS idx_task_history_event_type ON task_execution_history(event_type);
CREATE INDEX IF NOT EXISTS idx_task_history_created    ON task_execution_history(created_at DESC);

COMMENT ON TABLE task_execution_history IS 'Audit trail of task state changes and events';

-- -----------------------------------------------------------------------------
-- Table: task_resource_snapshots
-- -----------------------------------------------------------------------------
-- Time-series resource usage data sampled from running tasks. Used by the
-- stuck detector to identify tasks that have stopped making progress (flat
-- CPU, no I/O) and by the resource monitor for capacity planning.
--
-- Primary Key: id (UUID, auto-generated)
-- Foreign Keys: task_id -> background_tasks(id) ON DELETE CASCADE
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS task_resource_snapshots (
    id               UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id          UUID          NOT NULL REFERENCES background_tasks(id) ON DELETE CASCADE,

    -- CPU metrics
    cpu_percent      DECIMAL(5,2),                              -- CPU usage percentage (0-100)
    cpu_user_time    DECIMAL(12,4),                              -- User-space CPU time (seconds)
    cpu_system_time  DECIMAL(12,4),                              -- Kernel CPU time (seconds)

    -- Memory metrics
    memory_rss_bytes BIGINT,                                     -- Resident Set Size (physical memory)
    memory_vms_bytes BIGINT,                                     -- Virtual Memory Size
    memory_percent   DECIMAL(5,2),                               -- Memory usage percentage

    -- I/O metrics
    io_read_bytes    BIGINT,                                     -- Total bytes read from disk
    io_write_bytes   BIGINT,                                     -- Total bytes written to disk
    io_read_count    BIGINT,                                     -- Number of read operations
    io_write_count   BIGINT,                                     -- Number of write operations

    -- Network metrics
    net_bytes_sent   BIGINT,                                     -- Total network bytes sent
    net_bytes_recv   BIGINT,                                     -- Total network bytes received
    net_connections  INTEGER,                                     -- Active network connections

    -- File descriptors
    open_files       INTEGER,                                     -- Open file handles
    open_fds         INTEGER,                                     -- Open file descriptors

    -- Process state
    process_state    VARCHAR(20),                                  -- OS process state (R, S, D, Z, T)
    thread_count     INTEGER,                                      -- Active thread count

    sampled_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW()       -- Snapshot timestamp
);

-- Indexes (migration 011)
CREATE INDEX IF NOT EXISTS idx_resource_snapshots_task   ON task_resource_snapshots(task_id, sampled_at DESC);
CREATE INDEX IF NOT EXISTS idx_resource_snapshots_recent ON task_resource_snapshots(sampled_at DESC);

COMMENT ON TABLE task_resource_snapshots IS 'Time-series resource usage data for monitoring and stuck detection';

-- -----------------------------------------------------------------------------
-- Table: webhook_deliveries
-- -----------------------------------------------------------------------------
-- Tracks webhook notification delivery attempts for task lifecycle events.
-- Supports retry with exponential backoff and records HTTP response codes.
--
-- Primary Key: id (UUID, auto-generated)
-- Foreign Keys: task_id -> background_tasks(id) ON DELETE SET NULL
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id              UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id         UUID          REFERENCES background_tasks(id) ON DELETE SET NULL,
    webhook_url     TEXT          NOT NULL,                    -- Target webhook URL
    event_type      VARCHAR(50)  NOT NULL,                    -- Event type (e.g., 'task.completed', 'task.failed')
    payload         JSONB        NOT NULL,                     -- Webhook payload (task data snapshot)
    status          VARCHAR(20)  DEFAULT 'pending',            -- Delivery status: 'pending', 'delivered', 'failed'
    attempts        INTEGER      DEFAULT 0,                    -- Delivery attempt count
    last_attempt_at TIMESTAMP WITH TIME ZONE,                  -- Last delivery attempt timestamp
    last_error      TEXT,                                       -- Last delivery error message
    response_code   INTEGER,                                    -- HTTP response code from last attempt
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    delivered_at    TIMESTAMP WITH TIME ZONE                    -- Successful delivery timestamp
);

-- Indexes (migration 011)
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_task   ON webhook_deliveries(task_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status ON webhook_deliveries(status) WHERE status != 'delivered';

-- Performance index (migration 012)
-- Partial index for pending/failed webhook retry queue
CREATE INDEX IF NOT EXISTS idx_webhooks_pending_retry
    ON webhook_deliveries (created_at, attempts)
    WHERE status = 'pending' OR status = 'failed';

COMMENT ON TABLE webhook_deliveries IS 'Webhook notification delivery tracking with retry support';

-- -----------------------------------------------------------------------------
-- Stored Functions
-- -----------------------------------------------------------------------------

-- Atomic task dequeue: selects the highest-priority pending task and assigns
-- it to the given worker in a single transaction using FOR UPDATE SKIP LOCKED.
CREATE OR REPLACE FUNCTION dequeue_background_task(
    p_worker_id     VARCHAR(100),
    p_max_cpu_cores INTEGER DEFAULT 0,
    p_max_memory_mb INTEGER DEFAULT 0
)
RETURNS TABLE (
    task_id   UUID,
    task_type VARCHAR(100),
    task_name VARCHAR(255),
    payload   JSONB,
    config    JSONB,
    priority  task_priority
) AS $$
DECLARE
    v_task_id UUID;
BEGIN
    UPDATE background_tasks
    SET
        status = 'running',
        worker_id = p_worker_id,
        started_at = NOW(),
        last_heartbeat = NOW()
    WHERE id = (
        SELECT bt.id
        FROM background_tasks bt
        WHERE bt.status = 'pending'
          AND bt.scheduled_at <= NOW()
          AND bt.deleted_at IS NULL
          AND (p_max_cpu_cores = 0 OR bt.required_cpu_cores <= p_max_cpu_cores)
          AND (p_max_memory_mb = 0 OR bt.required_memory_mb <= p_max_memory_mb)
        ORDER BY
            CASE bt.priority
                WHEN 'critical'   THEN 0
                WHEN 'high'       THEN 1
                WHEN 'normal'     THEN 2
                WHEN 'low'        THEN 3
                WHEN 'background' THEN 4
            END,
            bt.created_at ASC
        LIMIT 1
        FOR UPDATE SKIP LOCKED
    )
    RETURNING background_tasks.id INTO v_task_id;

    IF v_task_id IS NOT NULL THEN
        RETURN QUERY
        SELECT bt.id, bt.task_type, bt.task_name, bt.payload, bt.config, bt.priority
        FROM background_tasks bt
        WHERE bt.id = v_task_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Returns running tasks whose last heartbeat exceeds the threshold (stuck detection).
CREATE OR REPLACE FUNCTION get_stale_tasks(
    p_heartbeat_threshold INTERVAL DEFAULT INTERVAL '5 minutes'
)
RETURNS TABLE (
    task_id              UUID,
    task_type            VARCHAR(100),
    worker_id            VARCHAR(100),
    started_at           TIMESTAMP WITH TIME ZONE,
    last_heartbeat       TIMESTAMP WITH TIME ZONE,
    time_since_heartbeat INTERVAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        bt.id, bt.task_type, bt.worker_id, bt.started_at, bt.last_heartbeat,
        NOW() - bt.last_heartbeat AS time_since_heartbeat
    FROM background_tasks bt
    WHERE bt.status = 'running'
      AND bt.last_heartbeat < NOW() - p_heartbeat_threshold
      AND bt.deleted_at IS NULL;
END;
$$ LANGUAGE plpgsql;
