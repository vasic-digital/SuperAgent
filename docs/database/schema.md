# HelixAgent Database Schema

## Overview

HelixAgent uses PostgreSQL 15 as its primary data store. This document describes the database schema, relationships, and indexing strategies.

## Schema Diagram

```
┌────────────────────┐     ┌────────────────────┐     ┌────────────────────┐
│      users         │     │    api_keys        │     │    sessions        │
├────────────────────┤     ├────────────────────┤     ├────────────────────┤
│ id (PK)            │◀────│ user_id (FK)       │     │ id (PK)            │
│ email              │     │ id (PK)            │     │ user_id (FK)       │◀──┐
│ password_hash      │     │ key_hash           │     │ provider_id (FK)   │   │
│ role               │     │ name               │     │ created_at         │   │
│ created_at         │     │ scopes             │     │ expires_at         │   │
│ updated_at         │     │ created_at         │     │ last_activity      │   │
│ status             │     │ expires_at         │     │ metadata           │   │
└────────────────────┘     │ last_used_at       │     └────────────────────┘   │
         │                 │ rate_limit         │              │               │
         │                 └────────────────────┘              │               │
         │                                                     │               │
         ▼                                                     ▼               │
┌────────────────────┐     ┌────────────────────┐     ┌────────────────────┐   │
│    providers       │     │  provider_configs  │     │    completions     │   │
├────────────────────┤     ├────────────────────┤     ├────────────────────┤   │
│ id (PK)            │◀────│ provider_id (FK)   │     │ id (PK)            │   │
│ name               │     │ id (PK)            │     │ session_id (FK)    │───┘
│ type               │     │ key                │     │ provider_id (FK)   │
│ enabled            │     │ value              │     │ prompt             │
│ priority           │     │ encrypted          │     │ response           │
│ created_at         │     │ updated_at         │     │ tokens_in          │
│ verification_score │     └────────────────────┘     │ tokens_out         │
│ last_verified      │                                │ latency_ms         │
│ health_status      │                                │ created_at         │
└────────────────────┘                                │ cached             │
         │                                            │ error              │
         │                                            └────────────────────┘
         │
         ▼
┌────────────────────┐     ┌────────────────────┐     ┌────────────────────┐
│  background_tasks  │     │    task_events     │     │    debates         │
├────────────────────┤     ├────────────────────┤     ├────────────────────┤
│ id (PK)            │◀────│ task_id (FK)       │     │ id (PK)            │
│ type               │     │ id (PK)            │     │ topic              │
│ payload            │     │ event_type         │     │ status             │
│ status             │     │ data               │     │ participants       │
│ priority           │     │ created_at         │     │ rounds             │
│ created_at         │     └────────────────────┘     │ consensus          │
│ started_at         │                                │ confidence         │
│ completed_at       │                                │ created_at         │
│ error              │                                │ completed_at       │
│ retry_count        │                                │ metadata           │
│ worker_id          │                                └────────────────────┘
└────────────────────┘                                         │
                                                               │
                                                               ▼
                                                ┌────────────────────┐
                                                │  debate_arguments  │
                                                ├────────────────────┤
                                                │ id (PK)            │
                                                │ debate_id (FK)     │
                                                │ position           │
                                                │ provider_id (FK)   │
                                                │ round              │
                                                │ argument           │
                                                │ confidence         │
                                                │ created_at         │
                                                └────────────────────┘
```

## Table Definitions

### users

Stores user account information.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_role ON users(role);
```

**Fields:**
| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| email | VARCHAR(255) | User email (unique) |
| password_hash | VARCHAR(255) | Bcrypt password hash |
| role | VARCHAR(50) | User role (user, admin, service) |
| status | VARCHAR(20) | Account status (active, suspended, deleted) |
| metadata | JSONB | Additional user metadata |

### api_keys

Stores API keys for authentication.

```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    scopes TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    rate_limit INTEGER DEFAULT 1000,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_status ON api_keys(status);
```

**Fields:**
| Field | Type | Description |
|-------|------|-------------|
| key_hash | VARCHAR(255) | SHA-256 hash of API key |
| scopes | TEXT[] | Allowed scopes (completions, debates, admin) |
| rate_limit | INTEGER | Requests per minute limit |

### providers

Stores LLM provider configurations.

```sql
CREATE TABLE providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 5,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verification_score DECIMAL(4,2) DEFAULT 0.0,
    last_verified TIMESTAMP WITH TIME ZONE,
    health_status VARCHAR(20) DEFAULT 'unknown',
    capabilities JSONB DEFAULT '{}'::jsonb,
    rate_limits JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_providers_name ON providers(name);
CREATE INDEX idx_providers_enabled ON providers(enabled);
CREATE INDEX idx_providers_type ON providers(type);
CREATE INDEX idx_providers_score ON providers(verification_score DESC);
```

**Fields:**
| Field | Type | Description |
|-------|------|-------------|
| type | VARCHAR(50) | Provider type (apikey, oauth, free) |
| verification_score | DECIMAL(4,2) | LLMsVerifier score (0-10) |
| health_status | VARCHAR(20) | Current health (healthy, degraded, unhealthy, unknown) |
| capabilities | JSONB | Provider capabilities (streaming, tools, vision) |

### provider_configs

Stores encrypted provider credentials.

```sql
CREATE TABLE provider_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    key VARCHAR(100) NOT NULL,
    value TEXT NOT NULL,
    encrypted BOOLEAN DEFAULT true,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(provider_id, key)
);

CREATE INDEX idx_provider_configs_provider ON provider_configs(provider_id);
```

### sessions

Stores user sessions for stateful interactions.

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    provider_id UUID REFERENCES providers(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,
    context JSONB DEFAULT '[]'::jsonb
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
CREATE INDEX idx_sessions_last_activity ON sessions(last_activity);
```

### completions

Stores completion requests and responses for auditing.

```sql
CREATE TABLE completions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    provider_id UUID NOT NULL REFERENCES providers(id),
    prompt TEXT NOT NULL,
    response TEXT,
    tokens_in INTEGER DEFAULT 0,
    tokens_out INTEGER DEFAULT 0,
    latency_ms INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    cached BOOLEAN DEFAULT false,
    error TEXT,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_completions_session ON completions(session_id);
CREATE INDEX idx_completions_provider ON completions(provider_id);
CREATE INDEX idx_completions_created ON completions(created_at DESC);
CREATE INDEX idx_completions_cached ON completions(cached);
```

### background_tasks

Stores background task queue.

```sql
CREATE TABLE background_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    priority INTEGER DEFAULT 5,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    worker_id VARCHAR(100),
    result JSONB,
    scheduled_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tasks_status ON background_tasks(status);
CREATE INDEX idx_tasks_priority ON background_tasks(priority DESC);
CREATE INDEX idx_tasks_created ON background_tasks(created_at);
CREATE INDEX idx_tasks_type ON background_tasks(type);
CREATE INDEX idx_tasks_scheduled ON background_tasks(scheduled_at) WHERE scheduled_at IS NOT NULL;
CREATE INDEX idx_tasks_pending ON background_tasks(priority DESC, created_at) WHERE status = 'pending';
```

**Task Status Values:**
- `pending` - Task is waiting to be processed
- `queued` - Task has been picked up by a worker
- `running` - Task is currently executing
- `completed` - Task finished successfully
- `failed` - Task failed after all retries
- `stuck` - Task hasn't progressed (detected by stuck detector)
- `cancelled` - Task was cancelled

### task_events

Stores task lifecycle events for real-time notifications.

```sql
CREATE TABLE task_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES background_tasks(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    data JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_task_events_task ON task_events(task_id);
CREATE INDEX idx_task_events_type ON task_events(event_type);
CREATE INDEX idx_task_events_created ON task_events(created_at DESC);
```

**Event Types:**
- `task.created` - Task was submitted
- `task.started` - Task started executing
- `task.progress` - Task progress update
- `task.completed` - Task finished successfully
- `task.failed` - Task failed
- `task.cancelled` - Task was cancelled

### debates

Stores AI debate sessions.

```sql
CREATE TABLE debates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    participants JSONB NOT NULL DEFAULT '[]'::jsonb,
    rounds INTEGER DEFAULT 0,
    max_rounds INTEGER DEFAULT 5,
    consensus TEXT,
    confidence DECIMAL(4,3) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'::jsonb,
    validation_config JSONB DEFAULT '{}'::jsonb,
    multipass_result JSONB
);

CREATE INDEX idx_debates_status ON debates(status);
CREATE INDEX idx_debates_created ON debates(created_at DESC);
```

**Debate Status Values:**
- `pending` - Debate not yet started
- `running` - Debate in progress
- `validating` - Multi-pass validation phase
- `completed` - Debate finished
- `failed` - Debate failed

### debate_arguments

Stores individual arguments from debate participants.

```sql
CREATE TABLE debate_arguments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    debate_id UUID NOT NULL REFERENCES debates(id) ON DELETE CASCADE,
    position VARCHAR(100) NOT NULL,
    provider_id UUID NOT NULL REFERENCES providers(id),
    round INTEGER NOT NULL,
    argument TEXT NOT NULL,
    confidence DECIMAL(4,3) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_primary BOOLEAN DEFAULT true,
    validation_status VARCHAR(20),
    validation_feedback TEXT
);

CREATE INDEX idx_debate_args_debate ON debate_arguments(debate_id);
CREATE INDEX idx_debate_args_position ON debate_arguments(position);
CREATE INDEX idx_debate_args_round ON debate_arguments(round);
CREATE INDEX idx_debate_args_provider ON debate_arguments(provider_id);
```

## Migrations

Migrations are managed using golang-migrate. Files are located in `migrations/`.

```bash
# Run migrations
make migrate-up

# Rollback last migration
make migrate-down

# Create new migration
make migrate-create NAME=add_new_table
```

## Indexes Strategy

### Primary Lookups
- All primary keys use UUID for distributed systems
- Unique constraints on business identifiers (email, name)

### Query Optimization
- Status-based indexes for queue polling
- Composite indexes for frequent queries
- Partial indexes for specific conditions (e.g., pending tasks)

### Timestamp Indexes
- DESC ordering for recent-first queries
- Used for audit trails and history

## Connection Pooling

```yaml
database:
  host: localhost
  port: 5432
  user: helixagent
  password: secret
  name: helixagent_db
  pool:
    max_open: 25
    max_idle: 5
    max_lifetime: 1h
    max_idle_time: 15m
```

## Backup Strategy

```bash
# Daily full backup
pg_dump -Fc helixagent_db > backup_$(date +%Y%m%d).dump

# Point-in-time recovery with WAL archiving
archive_mode = on
archive_command = 'cp %p /archive/%f'
```

---

**Document Version**: 1.0
**Last Updated**: January 23, 2026
**Author**: Generated by Claude Code
