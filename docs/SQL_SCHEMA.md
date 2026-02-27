# HelixAgent SQL Schema Documentation

## Overview

This document provides comprehensive documentation for the HelixAgent PostgreSQL database schema. The schema supports user management, LLM request/response tracking, AI debate orchestration, memory systems, background task processing, and comprehensive analytics.

**Database:** PostgreSQL 15+  
**Extensions:** uuid-ossp, pgvector  
**Total Tables:** 40+  
**Total Indexes:** 100+  
**Materialized Views:** 10+  

---

## Table of Contents

1. [Extensions](#extensions)
2. [Core Tables](#core-tables)
3. [Debate System](#debate-system)
4. [Background Tasks](#background-tasks)
5. [Protocol Support](#protocol-support)
6. [Memory Systems](#memory-systems)
7. [Analytics](#analytics)
8. [Indexes](#indexes)
9. [Entity Relationship Diagram](#entity-relationship-diagram)

---

## Extensions

### Required Extensions

```sql
-- UUID generation for primary keys
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Vector operations for embeddings (similarity search)
CREATE EXTENSION IF NOT EXISTS pgvector;
```

---

## Core Tables

### 1. users
**Purpose:** Registered user accounts with API key management

```sql
CREATE TABLE users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    username      VARCHAR(255) UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    api_key       VARCHAR(255) UNIQUE NOT NULL,
    role          VARCHAR(50)  DEFAULT 'user',
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Columns:**
- `id` (UUID) - Primary key
- `username` (VARCHAR) - Unique username
- `email` (VARCHAR) - Unique email address
- `password_hash` (VARCHAR) - Bcrypt hashed password
- `api_key` (VARCHAR) - Unique API key for authentication
- `role` (VARCHAR) - 'user' or 'admin'
- `created_at` (TIMESTAMP) - Account creation time
- `updated_at` (TIMESTAMP) - Last update time

**Relationships:**
- One-to-Many: user_sessions (user_id)
- One-to-Many: llm_requests (user_id)

**Indexes:**
- PRIMARY KEY (id)
- UNIQUE (username)
- UNIQUE (email)
- UNIQUE (api_key)

---

### 2. user_sessions
**Purpose:** Active user sessions with conversation context

```sql
CREATE TABLE user_sessions (
    id             UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID         REFERENCES users(id) ON DELETE CASCADE,
    session_token  VARCHAR(255) UNIQUE NOT NULL,
    context        JSONB        DEFAULT '{}',
    memory_id      UUID,
    status         VARCHAR(50)  DEFAULT 'active',
    request_count  INTEGER      DEFAULT 0,
    last_activity  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at     TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Columns:**
- `id` (UUID) - Primary key
- `user_id` (UUID) - Foreign key to users
- `session_token` (VARCHAR) - Unique session identifier
- `context` (JSONB) - Session context data
- `memory_id` (UUID) - Reference to Cognee memory
- `status` (VARCHAR) - 'active', 'expired', 'revoked'
- `request_count` (INTEGER) - Number of requests in session
- `last_activity` (TIMESTAMP) - Last request time
- `expires_at` (TIMESTAMP) - Session expiration time
- `created_at` (TIMESTAMP) - Session creation time

**Relationships:**
- Many-to-One: users (user_id)
- One-to-Many: llm_requests (session_id)
- One-to-Many: cognee_memories (session_id)

---

### 3. llm_providers
**Purpose:** Registry of configured LLM providers

```sql
CREATE TABLE llm_providers (
    id                    UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                  VARCHAR(255)  UNIQUE NOT NULL,
    type                  VARCHAR(100)  NOT NULL,
    api_key               VARCHAR(255),
    base_url              VARCHAR(500),
    model                 VARCHAR(255),
    weight                DECIMAL(5,2)  DEFAULT 1.0,
    enabled               BOOLEAN       DEFAULT TRUE,
    config                JSONB         DEFAULT '{}',
    health_status         VARCHAR(50)   DEFAULT 'unknown',
    response_time         BIGINT        DEFAULT 0,
    modelsdev_provider_id VARCHAR(255),
    total_models          INTEGER       DEFAULT 0,
    enabled_models        INTEGER       DEFAULT 0,
    last_models_sync      TIMESTAMP WITH TIME ZONE,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Columns:**
- `id` (UUID) - Primary key
- `name` (VARCHAR) - Provider name (e.g., 'openai', 'anthropic')
- `type` (VARCHAR) - Provider type
- `api_key` (VARCHAR) - API key for provider
- `base_url` (VARCHAR) - Base URL for API
- `model` (VARCHAR) - Default model
- `weight` (DECIMAL) - Ensemble voting weight
- `enabled` (BOOLEAN) - Whether provider is active
- `config` (JSONB) - Additional configuration
- `health_status` (VARCHAR) - 'healthy', 'degraded', 'unhealthy', 'unknown'
- `response_time` (BIGINT) - Last response time in ms
- `modelsdev_provider_id` (VARCHAR) - Models.dev integration ID
- `total_models` (INTEGER) - Total available models
- `enabled_models` (INTEGER) - Enabled model count
- `last_models_sync` (TIMESTAMP) - Last models.dev sync

---

### 4. llm_requests
**Purpose:** Every LLM completion request with full context

```sql
CREATE TABLE llm_requests (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id      UUID         REFERENCES user_sessions(id) ON DELETE CASCADE,
    user_id         UUID         REFERENCES users(id) ON DELETE CASCADE,
    prompt          TEXT         NOT NULL,
    messages        JSONB        NOT NULL DEFAULT '[]',
    model_params    JSONB        NOT NULL DEFAULT '{}',
    ensemble_config JSONB        DEFAULT NULL,
    memory_enhanced BOOLEAN      DEFAULT FALSE,
    memory          JSONB        DEFAULT '{}',
    status          VARCHAR(50)  DEFAULT 'pending',
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at      TIMESTAMP WITH TIME ZONE,
    completed_at    TIMESTAMP WITH TIME ZONE,
    request_type    VARCHAR(50)  DEFAULT 'completion'
);
```

**Columns:**
- `id` (UUID) - Primary key
- `session_id` (UUID) - Foreign key to user_sessions
- `user_id` (UUID) - Foreign key to users
- `prompt` (TEXT) - User prompt
- `messages` (JSONB) - OpenAI-format message array
- `model_params` (JSONB) - Temperature, top_p, max_tokens, etc.
- `ensemble_config` (JSONB) - Ensemble strategy configuration
- `memory_enhanced` (BOOLEAN) - Whether memories were injected
- `memory` (JSONB) - Injected memory data
- `status` (VARCHAR) - 'pending', 'running', 'completed', 'failed'
- Timestamps for tracking request lifecycle

**Relationships:**
- Many-to-One: user_sessions (session_id)
- Many-to-One: users (user_id)
- One-to-Many: llm_responses (request_id)

---

### 5. llm_responses
**Purpose:** Individual provider responses (multiple per request in ensemble mode)

```sql
CREATE TABLE llm_responses (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id      UUID         REFERENCES llm_requests(id) ON DELETE CASCADE,
    provider_id     UUID         REFERENCES llm_providers(id) ON DELETE SET NULL,
    provider_name   VARCHAR(255) NOT NULL,
    content         TEXT         NOT NULL,
    raw_response    JSONB        NOT NULL DEFAULT '{}',
    tokens_prompt   INTEGER      DEFAULT 0,
    tokens_completion INTEGER    DEFAULT 0,
    tokens_total    INTEGER      DEFAULT 0,
    confidence      DECIMAL(5,4) DEFAULT 0.0,
    response_time   BIGINT       DEFAULT 0,
    finish_reason   VARCHAR(50),
    selected        BOOLEAN      DEFAULT FALSE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Columns:**
- `id` (UUID) - Primary key
- `request_id` (UUID) - Foreign key to llm_requests
- `provider_id` (UUID) - Foreign key to llm_providers
- `provider_name` (VARCHAR) - Provider name
- `content` (TEXT) - Response content
- `raw_response` (JSONB) - Complete raw response
- `tokens_*` (INTEGER) - Token counts
- `confidence` (DECIMAL) - Response confidence score
- `response_time` (BIGINT) - Response time in ms
- `finish_reason` (VARCHAR) - Why generation stopped
- `selected` (BOOLEAN) - Whether this is the ensemble winner

---

## Debate System

### 6. debate_sessions
**Purpose:** AI debate session management

```sql
CREATE TABLE debate_sessions (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    topic           TEXT         NOT NULL,
    description     TEXT,
    status          VARCHAR(50)  DEFAULT 'pending',
    current_round   INTEGER      DEFAULT 0,
    total_rounds    INTEGER      NOT NULL,
    voting_method   VARCHAR(50)  DEFAULT 'weighted',
    positions       JSONB        NOT NULL DEFAULT '[]',
    participants    JSONB        NOT NULL DEFAULT '[]',
    final_winner    VARCHAR(255),
    metadata        JSONB        DEFAULT '{}',
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at      TIMESTAMP WITH TIME ZONE,
    completed_at    TIMESTAMP WITH TIME ZONE
);
```

**Columns:**
- `id` (UUID) - Primary key
- `topic` (TEXT) - Debate topic/question
- `description` (TEXT) - Detailed description
- `status` (VARCHAR) - 'pending', 'active', 'completed', 'cancelled'
- `current_round` (INTEGER) - Current debate round
- `total_rounds` (INTEGER) - Total planned rounds
- `voting_method` (VARCHAR) - 'weighted', 'majority', 'borda', 'condorcet'
- `positions` (JSONB) - Array of debate positions
- `participants` (JSONB) - Array of participating agents
- `final_winner` (VARCHAR) - Winning position
- `metadata` (JSONB) - Additional debate metadata
- Timestamps for lifecycle tracking

---

### 7. debate_turns
**Purpose:** Individual turns in a debate session

```sql
CREATE TABLE debate_turns (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id      UUID         REFERENCES debate_sessions(id) ON DELETE CASCADE,
    round_number    INTEGER      NOT NULL,
    turn_number     INTEGER      NOT NULL,
    participant_id  VARCHAR(255) NOT NULL,
    position        VARCHAR(255) NOT NULL,
    content         TEXT         NOT NULL,
    raw_response    JSONB        DEFAULT '{}',
    tokens_used     INTEGER      DEFAULT 0,
    response_time   BIGINT       DEFAULT 0,
    confidence      DECIMAL(5,4) DEFAULT 0.0,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Columns:**
- `id` (UUID) - Primary key
- `session_id` (UUID) - Foreign key to debate_sessions
- `round_number` (INTEGER) - Which round this turn belongs to
- `turn_number` (INTEGER) - Sequential turn number
- `participant_id` (VARCHAR) - Agent/provider ID
- `position` (VARCHAR) - Position being argued
- `content` (TEXT) - Turn content
- `raw_response` (JSONB) - Complete raw response
- `tokens_used` (INTEGER) - Tokens consumed
- `response_time` (BIGINT) - Response time in ms
- `confidence` (DECIMAL) - Confidence score

---

### 8. code_versions
**Purpose:** Track code versions for debate CI/CD integration

```sql
CREATE TABLE code_versions (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id      UUID         REFERENCES debate_sessions(id) ON DELETE CASCADE,
    version_hash    VARCHAR(64)  NOT NULL,
    branch          VARCHAR(255) NOT NULL,
    commit_message  TEXT,
    files_changed   JSONB        DEFAULT '[]',
    diff_summary    TEXT,
    validation_status VARCHAR(50) DEFAULT 'pending',
    test_results    JSONB        DEFAULT '{}',
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Columns:**
- `id` (UUID) - Primary key
- `session_id` (UUID) - Foreign key to debate_sessions
- `version_hash` (VARCHAR) - Git commit hash
- `branch` (VARCHAR) - Git branch name
- `commit_message` (TEXT) - Commit message
- `files_changed` (JSONB) - Array of changed files
- `diff_summary` (TEXT) - Summary of changes
- `validation_status` (VARCHAR) - 'pending', 'passed', 'failed'
- `test_results` (JSONB) - Test execution results

---

## Background Tasks

### 9. background_tasks
**Purpose:** Task queue for async processing

```sql
CREATE TABLE background_tasks (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_type       VARCHAR(100) NOT NULL,
    payload         JSONB        NOT NULL DEFAULT '{}',
    status          task_status  DEFAULT 'pending',
    priority        task_priority DEFAULT 'normal',
    worker_id       VARCHAR(255),
    retry_count     INTEGER      DEFAULT 0,
    max_retries     INTEGER      DEFAULT 3,
    error_message   TEXT,
    result          JSONB,
    scheduled_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at      TIMESTAMP WITH TIME ZONE,
    completed_at    TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Custom Types:**
```sql
CREATE TYPE task_status AS ENUM (
    'pending', 'queued', 'running', 'paused', 
    'completed', 'failed', 'stuck', 'cancelled', 'dead_letter'
);

CREATE TYPE task_priority AS ENUM (
    'critical', 'high', 'normal', 'low', 'background'
);
```

---

### 10. dead_letter_queue
**Purpose:** Failed tasks for manual inspection

```sql
CREATE TABLE dead_letter_queue (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    original_task_id UUID,
    task_type       VARCHAR(100) NOT NULL,
    payload         JSONB        NOT NULL,
    error_message   TEXT         NOT NULL,
    error_stack     TEXT,
    retry_history   JSONB        DEFAULT '[]',
    moved_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

## Memory Systems

### 11. cognee_memories
**Purpose:** Mem0-style memory with entity graphs

```sql
CREATE TABLE cognee_memories (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID         REFERENCES users(id) ON DELETE CASCADE,
    session_id      UUID         REFERENCES user_sessions(id) ON DELETE CASCADE,
    content         TEXT         NOT NULL,
    memory_type     VARCHAR(50)  DEFAULT 'fact',
    entities        JSONB        DEFAULT '[]',
    relationships   JSONB        DEFAULT '[]',
    confidence      DECIMAL(5,4) DEFAULT 0.8,
    access_count    INTEGER      DEFAULT 0,
    last_accessed   TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

### 12. distributed_memory
**Purpose:** Sharded memory for high-scale deployments

```sql
CREATE TABLE distributed_memory (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    shard_key       VARCHAR(64)  NOT NULL,
    user_id         UUID         NOT NULL,
    content         TEXT         NOT NULL,
    embedding       vector(1536),
    metadata        JSONB        DEFAULT '{}',
    node_id         VARCHAR(255) NOT NULL,
    replicated_to   JSONB        DEFAULT '[]',
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

## Protocol Support

### 13. mcp_servers
**Purpose:** MCP (Model Context Protocol) server registry

```sql
CREATE TABLE mcp_servers (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) UNIQUE NOT NULL,
    type            VARCHAR(100) NOT NULL,
    endpoint        VARCHAR(500) NOT NULL,
    port            INTEGER      NOT NULL,
    capabilities    JSONB        NOT NULL DEFAULT '[]',
    config          JSONB        DEFAULT '{}',
    health_status   VARCHAR(50)  DEFAULT 'unknown',
    enabled         BOOLEAN      DEFAULT TRUE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

### 14. lsp_servers
**Purpose:** LSP (Language Server Protocol) server registry

```sql
CREATE TABLE lsp_servers (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) UNIQUE NOT NULL,
    language        VARCHAR(100) NOT NULL,
    command         VARCHAR(500) NOT NULL,
    args            JSONB        DEFAULT '[]',
    file_patterns   JSONB        DEFAULT '[]',
    enabled         BOOLEAN      DEFAULT TRUE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

## Analytics

### 15. clickhouse_analytics
**Purpose:** Analytics export configuration for ClickHouse

```sql
CREATE TABLE clickhouse_analytics (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB        NOT NULL,
    user_id         UUID,
    session_id      UUID,
    timestamp       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    exported        BOOLEAN      DEFAULT FALSE,
    exported_at     TIMESTAMP WITH TIME ZONE
);
```

---

### 16. streaming_analytics
**Purpose:** Real-time streaming metrics

```sql
CREATE TABLE streaming_analytics (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    stream_id       VARCHAR(255) NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    chunk_number    INTEGER      NOT NULL,
    latency_ms      BIGINT       NOT NULL,
    tokens_delta    INTEGER      DEFAULT 0,
    user_id         UUID,
    timestamp       TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

## Indexes

### Performance Indexes

```sql
-- Users
CREATE INDEX idx_users_api_key ON users(api_key);
CREATE INDEX idx_users_email ON users(email);

-- Sessions
CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_sessions_status ON user_sessions(status);

-- Requests
CREATE INDEX idx_requests_user_id ON llm_requests(user_id);
CREATE INDEX idx_requests_session_id ON llm_requests(session_id);
CREATE INDEX idx_requests_status ON llm_requests(status);
CREATE INDEX idx_requests_created_at ON llm_requests(created_at);

-- Responses
CREATE INDEX idx_responses_request_id ON llm_responses(request_id);
CREATE INDEX idx_responses_provider_id ON llm_responses(provider_id);
CREATE INDEX idx_responses_selected ON llm_responses(selected) WHERE selected = TRUE;

-- Debate
CREATE INDEX idx_debate_sessions_status ON debate_sessions(status);
CREATE INDEX idx_debate_sessions_created_by ON debate_sessions(created_by);
CREATE INDEX idx_debate_turns_session_id ON debate_turns(session_id);
CREATE INDEX idx_debate_turns_round ON debate_turns(session_id, round_number);

-- Tasks
CREATE INDEX idx_tasks_status ON background_tasks(status);
CREATE INDEX idx_tasks_priority ON background_tasks(priority);
CREATE INDEX idx_tasks_scheduled_at ON background_tasks(scheduled_at);
CREATE INDEX idx_tasks_worker_id ON background_tasks(worker_id);

-- Memories
CREATE INDEX idx_memories_user_id ON cognee_memories(user_id);
CREATE INDEX idx_memories_session_id ON cognee_memories(session_id);
CREATE INDEX idx_memories_type ON cognee_memories(memory_type);
```

---

## Entity Relationship Diagram

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│     users       │────▶│  user_sessions   │◀────│  llm_requests   │
├─────────────────┤     ├──────────────────┤     ├─────────────────┤
│ PK: id          │     │ PK: id           │     │ PK: id          │
│ username (UQ)   │     │ FK: user_id      │     │ FK: user_id     │
│ email (UQ)      │     │ session_token    │     │ FK: session_id  │
│ api_key (UQ)    │     │ status           │     │ prompt          │
│ role            │     │ expires_at       │     │ messages (JSONB)│
└─────────────────┘     └──────────────────┘     └─────────────────┘
         │                                               │
         │                                               │
         ▼                                               ▼
┌─────────────────┐                           ┌─────────────────┐
│ cognee_memories │                           │  llm_responses  │
├─────────────────┤                           ├─────────────────┤
│ PK: id          │                           │ PK: id          │
│ FK: user_id     │                           │ FK: request_id  │
│ FK: session_id  │                           │ FK: provider_id │
│ content         │                           │ content         │
│ entities (JSONB)│                           │ confidence      │
└─────────────────┘                           └─────────────────┘
                                                        │
                                                        ▼
                                              ┌─────────────────┐
                                              │  llm_providers  │
                                              ├─────────────────┤
                                              │ PK: id          │
                                              │ name (UQ)       │
                                              │ health_status   │
                                              │ weight          │
                                              │ enabled         │
                                              └─────────────────┘

┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ debate_sessions │────▶│   debate_turns   │     │  code_versions  │
├─────────────────┤     ├──────────────────┤     ├─────────────────┤
│ PK: id          │     │ PK: id           │     │ PK: id          │
│ topic           │     │ FK: session_id   │     │ FK: session_id  │
│ status          │     │ round_number     │     │ version_hash    │
│ voting_method   │     │ participant_id   │     │ branch          │
│ positions(JSONB)│     │ content          │     │ validation      │
└─────────────────┘     └──────────────────┘     └─────────────────┘

┌─────────────────┐     ┌──────────────────┐
│background_tasks │────▶│ dead_letter_queue│
├─────────────────┤     ├──────────────────┤
│ PK: id          │     │ PK: id           │
│ task_type       │     │ original_task_id │
│ status (ENUM)   │     │ error_message   │
│ priority (ENUM) │     │ moved_at        │
│ scheduled_at    │     └──────────────────┘
└─────────────────┘
```

---

## Migration History

| Migration | Description | Date |
|-----------|-------------|------|
| 001 | Core tables (users, sessions, providers, requests, responses) | 2025-01-15 |
| 002 | Models.dev integration columns | 2025-01-20 |
| 003 | Cognee memory system | 2025-02-01 |
| 011 | Debate system tables | 2025-02-10 |
| 012 | Background task queue | 2025-02-15 |
| 013 | Protocol support (MCP, LSP, ACP) | 2025-02-20 |
| 014 | Distributed memory sharding | 2025-02-25 |

---

## Best Practices

### 1. Query Optimization
- Use indexes for filtering on status, user_id, created_at
- Use JSONB operators for querying nested data
- Partition large tables by date ranges

### 2. Data Retention
- Archive old llm_requests after 90 days
- Maintain summary statistics in materialized views
- Use dead_letter_queue for failed task investigation

### 3. Security
- Hash passwords with bcrypt (cost factor 12+)
- Encrypt API keys at rest
- Use Row-Level Security (RLS) for multi-tenant scenarios
- Audit sensitive operations

### 4. Performance
- Batch inserts for high-volume tables
- Use connection pooling (recommended: 20-50 connections)
- Enable query plan caching
- Monitor slow queries with pg_stat_statements

---

## Maintenance

### Routine Maintenance Tasks

```sql
-- Analyze tables for query optimization
ANALYZE llm_requests;
ANALYZE llm_responses;

-- Reindex if needed
REINDEX TABLE CONCURRENTLY llm_requests;

-- Vacuum to reclaim space
VACUUM ANALYZE;

-- Update materialized views
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_daily_stats;
```

---

**Last Updated:** February 27, 2026  
**Schema Version:** 14  
**PostgreSQL Version:** 15+
