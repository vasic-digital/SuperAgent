# Database Schema Reference

HelixAgent uses PostgreSQL 15+ with a schema covering users, LLM providers, request/response logging, background tasks, AI debate, knowledge management, and protocol support.

## Schema Overview

| Domain | Tables | SQL File |
|--------|--------|----------|
| Users & Sessions | `users`, `user_sessions` | `sql/schema/users_sessions.sql` |
| LLM Providers | `llm_providers`, `llm_models`, `llm_benchmarks` | `sql/schema/llm_providers.sql` |
| Requests & Responses | `llm_requests`, `llm_responses` | `sql/schema/requests_responses.sql` |
| Background Tasks | `background_tasks`, `dead_letter_queue`, `task_execution_history` | `sql/schema/background_tasks.sql` |
| AI Debate | `debate_logs`, `debate_arguments`, `debate_rounds`, `debate_votes` | `sql/schema/debate_system.sql` |
| Knowledge | `cognee_memories`, `cognee_graphs`, `cognee_entities` | `sql/schema/cognee_memories.sql` |
| Protocols | `mcp_servers`, `mcp_tools`, `acp_agents`, `lsp_servers` | `sql/schema/protocol_support.sql` |
| Performance | Indexes, materialized views | `sql/schema/indexes_views.sql` |

**Consolidated reference**: `sql/schema/complete_schema.sql` contains all tables with comprehensive comments.

**Relationships**: `sql/schema/relationships.sql` documents all foreign key constraints and entity relationships.

## Core Tables

### Users & Sessions

```sql
-- Users table: authentication and profile
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(255) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    role            VARCHAR(50) DEFAULT 'user',
    is_active       BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- User sessions: JWT token tracking
CREATE TABLE user_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

### LLM Providers

```sql
-- Provider registry with scoring
CREATE TABLE llm_providers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) UNIQUE NOT NULL,
    display_name    VARCHAR(255),
    provider_type   VARCHAR(50) NOT NULL,    -- "api_key", "oauth", "free"
    is_enabled      BOOLEAN DEFAULT true,
    verification_score DECIMAL(4,2),          -- LLMsVerifier score (0-10)
    last_verified   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Models available per provider
CREATE TABLE llm_models (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id     UUID NOT NULL REFERENCES llm_providers(id),
    model_name      VARCHAR(255) NOT NULL,
    capabilities    JSONB DEFAULT '{}',
    max_tokens      INTEGER,
    cost_per_token  DECIMAL(10,8),
    is_active       BOOLEAN DEFAULT true
);
```

### Request/Response Logging

```sql
-- All LLM API requests
CREATE TABLE llm_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id),
    provider_id     UUID REFERENCES llm_providers(id),
    model_name      VARCHAR(255),
    prompt_tokens   INTEGER,
    request_body    JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- LLM API responses
CREATE TABLE llm_responses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id      UUID NOT NULL REFERENCES llm_requests(id),
    completion_tokens INTEGER,
    total_tokens    INTEGER,
    response_body   JSONB,
    latency_ms      INTEGER,
    status          VARCHAR(50),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

### Background Tasks

```sql
-- Async task queue
CREATE TABLE background_tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_type       VARCHAR(100) NOT NULL,
    payload         JSONB NOT NULL,
    status          VARCHAR(50) DEFAULT 'pending',  -- pending/queued/running/completed/failed/stuck/cancelled
    priority        INTEGER DEFAULT 0,
    max_retries     INTEGER DEFAULT 3,
    retry_count     INTEGER DEFAULT 0,
    scheduled_at    TIMESTAMPTZ,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Dead letter queue for permanently failed tasks
CREATE TABLE dead_letter_queue (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id         UUID NOT NULL REFERENCES background_tasks(id),
    failure_reason  TEXT,
    original_payload JSONB,
    moved_at        TIMESTAMPTZ DEFAULT NOW()
);
```

### AI Debate System

```sql
-- Debate sessions
CREATE TABLE debate_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic           TEXT NOT NULL,
    status          VARCHAR(50) DEFAULT 'active',
    strategy        VARCHAR(100),
    total_rounds    INTEGER DEFAULT 3,
    consensus_reached BOOLEAN DEFAULT false,
    final_conclusion TEXT,
    confidence_score DECIMAL(4,2),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);

-- Individual debate arguments
CREATE TABLE debate_arguments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    debate_id       UUID NOT NULL REFERENCES debate_logs(id),
    round_number    INTEGER NOT NULL,
    provider_name   VARCHAR(100) NOT NULL,
    model_name      VARCHAR(255),
    position        TEXT NOT NULL,
    argument_text   TEXT NOT NULL,
    confidence      DECIMAL(4,2),
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

## Relationships

Key foreign key relationships:

- `user_sessions.user_id` → `users.id` (CASCADE)
- `llm_models.provider_id` → `llm_providers.id`
- `llm_requests.user_id` → `users.id`
- `llm_requests.provider_id` → `llm_providers.id`
- `llm_responses.request_id` → `llm_requests.id`
- `background_tasks` → `dead_letter_queue.task_id`
- `debate_arguments.debate_id` → `debate_logs.id`

See `sql/schema/relationships.sql` for the complete relationship documentation.

## ER Diagram

Visual entity-relationship diagrams:

- `docs/diagrams/src/database-er.mmd` - Mermaid ER diagram
- `docs/diagrams/src/database-er.puml` - PlantUML class diagram

Generate:
```bash
./scripts/generate-diagrams.sh
```

## Migrations

Migrations are in `migrations/` directory (files 001-014). The consolidated schema in `sql/schema/complete_schema.sql` represents the current state after all migrations.

## SQL Files

| File | Description |
|------|-------------|
| `sql/schema/complete_schema.sql` | Full consolidated schema |
| `sql/schema/users_sessions.sql` | Users and session tables |
| `sql/schema/llm_providers.sql` | LLM provider and model tables |
| `sql/schema/requests_responses.sql` | Request/response logging |
| `sql/schema/background_tasks.sql` | Task queue and dead letter |
| `sql/schema/debate_system.sql` | AI debate tables |
| `sql/schema/cognee_memories.sql` | Knowledge management |
| `sql/schema/protocol_support.sql` | MCP/ACP/LSP tables |
| `sql/schema/indexes_views.sql` | Performance indexes and views |
| `sql/schema/relationships.sql` | FK constraints documentation |
