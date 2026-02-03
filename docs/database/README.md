# Database Documentation

This directory contains database schema documentation, migration guides, and setup instructions for HelixAgent's data layer.

## Overview

HelixAgent uses PostgreSQL 15+ as its primary data store, with additional support for ClickHouse (time-series analytics) and Neo4j (knowledge graphs) in the Big Data integration.

## Schema Reference

### Core Schema Documents

| Document | Description |
|----------|-------------|
| [schema.md](schema.md) | Core database schema with entity relationship diagrams and table definitions |
| [SCHEMA_REFERENCE.md](SCHEMA_REFERENCE.md) | Comprehensive schema reference covering all 7 domains |
| [BIG_DATA_SCHEMA_REFERENCE.md](BIG_DATA_SCHEMA_REFERENCE.md) | Big Data integration schema (25 tables across PostgreSQL, ClickHouse, Neo4j) |

## Schema Domains

The HelixAgent database is organized into 7 core domains:

| Domain | Tables | SQL File |
|--------|--------|----------|
| Users & Sessions | `users`, `user_sessions` | `sql/schema/users_sessions.sql` |
| LLM Providers | `llm_providers`, `llm_models`, `llm_benchmarks` | `sql/schema/llm_providers.sql` |
| Requests & Responses | `llm_requests`, `llm_responses` | `sql/schema/requests_responses.sql` |
| Background Tasks | `background_tasks`, `dead_letter_queue`, `task_execution_history` | `sql/schema/background_tasks.sql` |
| AI Debate | `debate_logs`, `debate_arguments`, `debate_rounds`, `debate_votes` | `sql/schema/debate_system.sql` |
| Knowledge | `cognee_memories`, `cognee_graphs`, `cognee_entities` | `sql/schema/cognee_memories.sql` |
| Protocols | `mcp_servers`, `mcp_tools`, `acp_agents`, `lsp_servers` | `sql/schema/protocol_support.sql` |

**Consolidated reference**: `sql/schema/complete_schema.sql`

## Big Data Schema

The Big Data integration adds 25 additional tables:

| System | Tables | Purpose |
|--------|--------|---------|
| PostgreSQL | 17 | Conversation context, distributed memory, cross-session learning |
| ClickHouse | 9 | Time-series analytics and metrics |
| Neo4j | 6 types | Knowledge graph nodes and relationships |

### Key Big Data Tables

- `conversation_events` - Event sourcing for conversation replay
- `conversation_compressions` - Context compression tracking
- `distributed_memory_events` - CRDT-based distributed memory
- `cross_session_patterns` - Pattern learning across sessions

## Database Setup

### Prerequisites

- PostgreSQL 15 or higher
- pgx/v5 driver (Go)

### Quick Start

```bash
# Start PostgreSQL via Docker
make infra-core

# Or manually
docker run -d \
  --name helixagent-postgres \
  -e POSTGRES_USER=helixagent \
  -e POSTGRES_PASSWORD=helixagent123 \
  -e POSTGRES_DB=helixagent_db \
  -p 15432:5432 \
  postgres:15

# Apply schema
psql -h localhost -p 15432 -U helixagent -d helixagent_db -f sql/schema/complete_schema.sql
```

### Environment Variables

```bash
DB_HOST=localhost
DB_PORT=15432
DB_USER=helixagent
DB_PASSWORD=helixagent123
DB_NAME=helixagent_db
```

### Connection Pool Settings

Default connection pool configuration:

| Setting | Default | Description |
|---------|---------|-------------|
| MaxOpenConns | 25 | Maximum open connections |
| MaxIdleConns | 10 | Maximum idle connections |
| ConnMaxLifetime | 5m | Connection maximum lifetime |
| ConnMaxIdleTime | 1m | Connection maximum idle time |

## Migrations

### Migration Files Location

```
sql/
├── schema/
│   ├── complete_schema.sql          # All tables consolidated
│   ├── users_sessions.sql           # User authentication
│   ├── llm_providers.sql            # Provider registry
│   ├── requests_responses.sql       # Request logging
│   ├── background_tasks.sql         # Task queue
│   ├── debate_system.sql            # AI debate
│   ├── cognee_memories.sql          # Knowledge storage
│   ├── protocol_support.sql         # MCP/ACP/LSP
│   ├── indexes_views.sql            # Performance indexes
│   ├── relationships.sql            # Foreign key constraints
│   ├── conversation_context.sql     # Big Data: context
│   ├── distributed_memory.sql       # Big Data: memory
│   ├── cross_session_learning.sql   # Big Data: learning
│   └── clickhouse_analytics.sql     # Big Data: analytics
└── migrations/
    └── *.sql                         # Versioned migrations
```

### Running Migrations

```bash
# Apply all migrations
make db-migrate

# Apply specific migration
psql -h localhost -p 15432 -U helixagent -d helixagent_db -f sql/migrations/001_initial.sql
```

## Entity Relationship Diagram

```
users ────────────────── user_sessions
  │                           │
  │                           │
  ├──── api_keys              ├──── completions
  │                           │
  │                           │
  └──────────────────────────────────── providers
                                            │
                                            ├──── provider_configs
                                            │
                                            └──── llm_models

background_tasks ──────── task_events

debates ───────────────── debate_arguments
```

## Indexing Strategy

### Primary Indexes

All tables use UUID primary keys with `gen_random_uuid()`.

### Performance Indexes

```sql
-- User lookups
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);

-- Provider queries
CREATE INDEX idx_providers_enabled ON llm_providers(is_enabled);
CREATE INDEX idx_providers_score ON llm_providers(verification_score DESC);

-- Request logging
CREATE INDEX idx_requests_created ON llm_requests(created_at);
CREATE INDEX idx_requests_user ON llm_requests(user_id);

-- Task queue
CREATE INDEX idx_tasks_status ON background_tasks(status);
CREATE INDEX idx_tasks_priority ON background_tasks(priority DESC);
```

### GIN Indexes (JSONB)

```sql
CREATE INDEX idx_events_data ON conversation_events USING gin(event_data);
CREATE INDEX idx_tasks_payload ON background_tasks USING gin(payload);
```

## Related Documentation

- [Architecture Overview](../architecture/README.md)
- [Service Architecture](../architecture/SERVICE_ARCHITECTURE.md)
- [Big Data Integration](BIG_DATA_SCHEMA_REFERENCE.md)

## Quick Reference

| Connection | Value |
|------------|-------|
| Default Host | localhost |
| Default Port | 15432 |
| Default User | helixagent |
| Default Password | helixagent123 |
| Default Database | helixagent_db |
| Connection String | `postgresql://helixagent:helixagent123@localhost:15432/helixagent_db` |
