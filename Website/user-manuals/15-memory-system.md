# HelixAgent Memory System Guide

## Introduction

HelixAgent implements a Mem0-style memory system that persists user preferences, facts, and context across conversations. The system supports entity graph extraction, CRDT-based conflict resolution for multi-node deployments, distributed synchronization via Kafka, and event sourcing for full auditability. Memory is the primary persistence layer for user context (Cognee integration is optional and off by default).

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Memory Manager](#memory-manager)
3. [Memory Store](#memory-store)
4. [Entity Graphs](#entity-graphs)
5. [CRDT Conflict Resolution](#crdt-conflict-resolution)
6. [Distributed Memory Manager](#distributed-memory-manager)
7. [Event Sourcing](#event-sourcing)
8. [Configuration](#configuration)
9. [Usage Examples](#usage-examples)

---

## Architecture Overview

```
Conversation Messages
        |
        v
  Memory Manager (Mem0-style)
        |
   +----+----+
   |         |
   v         v
MemoryStore  EntityGraph
   |         (entities + relationships)
   v
CRDT Resolver <-- MemoryEvents
   |
   v
Distributed Memory Manager
   |
   +---> Kafka (event propagation)
   +---> EventLog (audit trail)
   +---> VectorClock (causal ordering)
```

The core packages reside in `internal/memory/`:

| File | Purpose |
|------|---------|
| `manager.go` | Mem0-style memory manager |
| `store_memory.go` | In-memory store with search |
| `crdt.go` | Conflict resolution strategies |
| `distributed_manager.go` | Multi-node synchronization |

---

## Memory Manager

The `Manager` (`internal/memory/manager.go`) provides Mem0-style capabilities: extracting memories from conversation messages, searching by query, and organizing by user and session.

### Core Operations

| Method | Description |
|--------|-------------|
| `AddMemory` | Store a new memory entry |
| `AddFromMessages` | Extract and store memories from conversation messages |
| `Search` | Search memories by text query with options |
| `GetByUser` | Retrieve all memories for a user |
| `GetBySession` | Retrieve memories from a specific session |

### Memory Structure

Each memory entry contains an ID, user ID, session ID, content text, category (preference, fact, instruction, context), importance score, arbitrary metadata, and timestamps (created, updated, last accessed).

---

## Memory Store

The `InMemoryStore` provides the default storage backend with full-text search, sorting, and pagination. Search accepts options for user filtering, category, minimum score, limit, sort field, and sort order. Match scores are calculated based on content similarity.

---

## Entity Graphs

The memory system extracts entities and relationships from conversations, building a knowledge graph per user.

### Entity Operations

| Method | Description |
|--------|-------------|
| `AddEntity` | Add a named entity (person, place, concept) |
| `GetEntity` | Retrieve entity by ID |
| `SearchEntities` | Search entities by name or type |
| `AddRelationship` | Create a relationship between entities |
| `GetRelationships` | Get all relationships for an entity |

Entities have an ID, name, and type (person, place, concept). Relationships link two entities with a typed edge (e.g., `works_with`, `knows`).

---

## CRDT Conflict Resolution

When multiple nodes update the same memory concurrently, the `CRDTResolver` (`internal/memory/crdt.go`) determines the winning state. Five strategies are available:

| Strategy | Constant | Behavior |
|----------|----------|----------|
| Last Write Wins | `last_write_wins` | Most recent timestamp wins |
| Merge All | `merge_all` | Intelligently merges all fields |
| Importance | `importance` | Highest importance score wins |
| Vector Clock | `vector_clock` | Uses causal ordering for consistency |
| Custom | `custom` | User-defined resolution function |

### Creating a Resolver

```go
// Use last-write-wins strategy
resolver := memory.NewCRDTResolver(memory.ConflictStrategyLastWriteWins)

// Or use a custom resolver
resolver := memory.NewCRDTResolver(memory.ConflictStrategyCustom)
resolver.WithCustomResolver(func(local *memory.Memory, remote *memory.MemoryEvent) *memory.Memory {
    // Custom merge logic
    if remote.Importance > local.Importance {
        local.Content = remote.Content
    }
    return local
})
```

### Merge Operation

The `Merge` method takes a local memory state and a remote event, returning the resolved memory:

```go
resolved := resolver.Merge(localMemory, remoteEvent)
```

---

## Distributed Memory Manager

The `DistributedMemoryManager` (`internal/memory/distributed_manager.go`) coordinates memory across multiple HelixAgent nodes.

### Components

| Component | Role |
|-----------|------|
| Local Manager | Handles single-node memory operations |
| Vector Clock | Tracks causal ordering across nodes |
| Event Log | Persists all memory events for replay |
| CRDT Resolver | Resolves concurrent updates |
| Kafka Publisher | Broadcasts events to other nodes |

Each node receives a unique ID (auto-generated UUID if not specified). When a memory is added, updated, or deleted, the local manager applies the change, creates a `MemoryEvent` with the node's vector clock, appends it to the event log, and publishes it to Kafka. Other nodes receive the event and apply it through the CRDT resolver. Use `GetVectorClock()` and `GetSyncStatus()` to monitor synchronization state.

---

## Event Sourcing

All memory mutations are captured as `MemoryEvent` records in an append-only event log. This enables:

- **Full audit trail**: Every change is recorded with timestamp, node ID, and vector clock
- **State reconstruction**: Replay events to rebuild memory state from scratch
- **Debugging**: Trace the history of any memory entry

### Event Log Interface

| Method | Description |
|--------|-------------|
| `Append` | Add a new event |
| `GetEvents` | Get all events for a memory ID |
| `GetEventsSince` | Get events after a timestamp |
| `GetEventsForUser` | Get all events for a user |
| `GetEventsFromNode` | Get events originating from a specific node |

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MEMORY_ENABLED` | `true` | Enable the memory system |
| `MEMORY_STORE_TYPE` | `memory` | Storage backend (memory, postgres) |
| `BIGDATA_ENABLE_DISTRIBUTED_MEMORY` | `false` | Enable distributed synchronization |
| `KAFKA_BOOTSTRAP_SERVERS` | `localhost:9092` | Kafka brokers for event propagation |
| `COGNEE_ENABLED` | `false` | Enable Cognee (Mem0 is primary) |

### Conflict Strategy

Set the default conflict resolution strategy in the configuration:

```yaml
memory:
  conflict_strategy: "last_write_wins"  # or merge_all, importance, vector_clock
```

---

## Usage Examples

### Storing Memories from a Conversation

```go
manager := memory.NewManager(store, logger)

messages := []memory.Message{
    {Role: "user", Content: "I prefer Python for data science work."},
    {Role: "assistant", Content: "Noted. I will use Python examples for data science topics."},
}

memories, err := manager.AddFromMessages(ctx, messages, "user-123", "session-456")
```

### Searching Memories

```go
results, err := manager.Search(ctx, "programming preferences", &memory.SearchOptions{
    UserID: "user-123",
    Limit:  5,
})

for _, mem := range results {
    fmt.Printf("[%s] %s (importance: %.2f)\n", mem.Category, mem.Content, mem.Importance)
}
```
