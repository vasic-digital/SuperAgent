-- =============================================================================
-- HelixAgent SQL Schema: Relationships & Entity-Relationship Documentation
-- =============================================================================
-- This file documents all foreign key relationships, logical references,
-- and entity-relationship descriptions across the HelixAgent database.
-- It contains no DDL -- only documentation comments.
-- =============================================================================

-- =============================================================================
-- FOREIGN KEY RELATIONSHIPS (Enforced by Database)
-- =============================================================================

-- -----------------------------------------------------------------------------
-- users (1) --< user_sessions (N)
-- -----------------------------------------------------------------------------
-- A user can have many sessions. Deleting a user cascades to all sessions.
-- FK: user_sessions.user_id -> users.id ON DELETE CASCADE

-- -----------------------------------------------------------------------------
-- users (1) --< llm_requests (N)
-- -----------------------------------------------------------------------------
-- A user can make many LLM requests. Deleting a user cascades to all requests.
-- FK: llm_requests.user_id -> users.id ON DELETE CASCADE

-- -----------------------------------------------------------------------------
-- user_sessions (1) --< llm_requests (N)
-- -----------------------------------------------------------------------------
-- A session can have many LLM requests. Deleting a session cascades to its requests.
-- FK: llm_requests.session_id -> user_sessions.id ON DELETE CASCADE

-- -----------------------------------------------------------------------------
-- user_sessions (1) --< cognee_memories (N)
-- -----------------------------------------------------------------------------
-- A session can accumulate many Cognee memories. Deleting a session cascades
-- to its memories.
-- FK: cognee_memories.session_id -> user_sessions.id ON DELETE CASCADE

-- -----------------------------------------------------------------------------
-- llm_requests (1) --< llm_responses (N)
-- -----------------------------------------------------------------------------
-- Each request generates one or more responses (one per provider in ensemble mode).
-- Deleting a request cascades to all its responses.
-- FK: llm_responses.request_id -> llm_requests.id ON DELETE CASCADE

-- -----------------------------------------------------------------------------
-- llm_providers (1) --< llm_responses (N)
-- -----------------------------------------------------------------------------
-- Each response comes from one provider. If a provider is deleted, the response
-- retains its data but the provider_id is set to NULL. The provider_name column
-- (denormalized) preserves the provider identity.
-- FK: llm_responses.provider_id -> llm_providers.id ON DELETE SET NULL

-- -----------------------------------------------------------------------------
-- llm_providers (1) --< models_metadata (N)
-- -----------------------------------------------------------------------------
-- Each provider can have many models in the catalog. Deleting a provider
-- cascades to all its model metadata entries.
-- FK: models_metadata.provider_id -> llm_providers.id ON DELETE CASCADE

-- -----------------------------------------------------------------------------
-- models_metadata (1) --< model_benchmarks (N)
-- -----------------------------------------------------------------------------
-- Each model can have many benchmark results. Deleting a model cascades to
-- its benchmarks. The FK uses model_id (VARCHAR), not the UUID primary key.
-- FK: model_benchmarks.model_id -> models_metadata.model_id ON DELETE CASCADE
-- Unique constraint: (model_id, benchmark_name) prevents duplicate benchmarks

-- -----------------------------------------------------------------------------
-- background_tasks (1) --< background_tasks (N)  [Self-referential]
-- -----------------------------------------------------------------------------
-- Tasks can have parent-child relationships for hierarchical task decomposition.
-- If a parent task is deleted, child tasks retain their data with parent_task_id
-- set to NULL.
-- FK: background_tasks.parent_task_id -> background_tasks.id ON DELETE SET NULL

-- -----------------------------------------------------------------------------
-- background_tasks (1) --< task_execution_history (N)
-- -----------------------------------------------------------------------------
-- Each task has an immutable audit trail of state transitions. Deleting a task
-- cascades to its execution history.
-- FK: task_execution_history.task_id -> background_tasks.id ON DELETE CASCADE

-- -----------------------------------------------------------------------------
-- background_tasks (1) --< task_resource_snapshots (N)
-- -----------------------------------------------------------------------------
-- Each running task has periodic resource usage snapshots. Deleting a task
-- cascades to its resource snapshots.
-- FK: task_resource_snapshots.task_id -> background_tasks.id ON DELETE CASCADE

-- -----------------------------------------------------------------------------
-- background_tasks (1) --< webhook_deliveries (N)
-- -----------------------------------------------------------------------------
-- Each task can trigger multiple webhook notifications. If a task is deleted,
-- webhook delivery records are preserved with task_id set to NULL.
-- FK: webhook_deliveries.task_id -> background_tasks.id ON DELETE SET NULL


-- =============================================================================
-- LOGICAL REFERENCES (Not Enforced by FK Constraints)
-- =============================================================================

-- -----------------------------------------------------------------------------
-- user_sessions.memory_id -> (external Cognee memory system)
-- -----------------------------------------------------------------------------
-- Optional reference to a Cognee memory ID. Not enforced by FK because
-- the memory may live in an external system (vector store, graph DB).

-- -----------------------------------------------------------------------------
-- models_metadata.mcp_server_id -> mcp_servers.id
-- models_metadata.lsp_server_id -> lsp_servers.id
-- models_metadata.acp_server_id -> acp_servers.id
-- -----------------------------------------------------------------------------
-- Protocol server associations for models. Not enforced by FK because
-- models and servers have independent lifecycles and may be configured
-- at different times.

-- -----------------------------------------------------------------------------
-- protocol_metrics.server_id -> mcp_servers.id / lsp_servers.id / acp_servers.id
-- -----------------------------------------------------------------------------
-- Metrics reference the server that handled the operation. Not enforced by FK
-- because metrics span multiple server tables (polymorphic reference) and
-- metrics must survive server deletion for historical analysis.

-- -----------------------------------------------------------------------------
-- background_tasks.user_id -> users.id
-- background_tasks.session_id -> user_sessions.id
-- -----------------------------------------------------------------------------
-- Task-user and task-session associations. Not enforced by FK for flexibility:
-- tasks may be system-initiated (no user), and task records should survive
-- user/session deletion for operational auditing.

-- -----------------------------------------------------------------------------
-- background_tasks_dead_letter.original_task_id -> background_tasks.id
-- -----------------------------------------------------------------------------
-- Dead-letter entries reference the original task. Not enforced by FK because
-- the task data is copied into the dead-letter table (task_data JSONB) and
-- the original task may have been deleted.

-- -----------------------------------------------------------------------------
-- debate_logs.session_id -> user_sessions (logical, uses VARCHAR not UUID)
-- debate_logs.debate_id -> (internal debate system, not stored in a table)
-- -----------------------------------------------------------------------------
-- Debate logs use string identifiers for flexibility and decoupling from
-- the session/debate lifecycle.

-- -----------------------------------------------------------------------------
-- vector_documents.embedding_id -> (external embedding store)
-- -----------------------------------------------------------------------------
-- Optional reference to an embedding in an external vector database
-- (Qdrant, Pinecone, Milvus). Not enforced by FK.

-- -----------------------------------------------------------------------------
-- cognee_memories.vector_id -> (external vector store)
-- -----------------------------------------------------------------------------
-- Optional reference to a vector embedding in an external store.
-- Not enforced by FK.


-- =============================================================================
-- ENTITY-RELATIONSHIP SUMMARY
-- =============================================================================
--
-- Core Domain:
--   users (1) --<< user_sessions (N) --<< llm_requests (N) --<< llm_responses (N)
--                                     \                        /
--                                      \-- cognee_memories    /-- llm_providers (1)
--
-- Model Catalog:
--   llm_providers (1) --<< models_metadata (N) --<< model_benchmarks (N)
--                                              \-- models_refresh_history (standalone)
--
-- Protocol Infrastructure:
--   mcp_servers, lsp_servers, acp_servers (independent)
--   embedding_config (standalone)
--   vector_documents (standalone with pgvector)
--   protocol_cache (standalone KV store)
--   protocol_metrics (references servers logically)
--
-- Background Tasks:
--   background_tasks (1) --<< task_execution_history (N)
--                        --<< task_resource_snapshots (N)
--                        --<< webhook_deliveries (N)
--                        --<< background_tasks (N) [self-ref: parent/child]
--   background_tasks_dead_letter (standalone archive)
--
-- AI Debate:
--   debate_logs (standalone, string-based references)
--
-- CASCADE Behavior Summary:
--   DELETE user       -> cascades to: sessions, requests
--   DELETE session    -> cascades to: requests, cognee_memories
--   DELETE request    -> cascades to: responses
--   DELETE provider   -> cascades to: models_metadata; SET NULL on: responses.provider_id
--   DELETE model      -> cascades to: model_benchmarks
--   DELETE task       -> cascades to: execution_history, resource_snapshots; SET NULL on: webhooks
--   DELETE parent_task -> SET NULL on: child tasks
