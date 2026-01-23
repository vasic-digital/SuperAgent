# Memory Management Guide

## Overview

HelixAgent implements a Mem0-style memory management system that enables persistent, contextual memory across conversations and sessions. This guide covers memory storage, retrieval, entity graphs, and best practices.

## Memory Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Memory Manager                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ Short-term   │  │  Long-term   │  │   Entity Graph       │  │
│  │ Memory       │  │  Memory      │  │   (Knowledge Graph)  │  │
│  │ (Session)    │  │ (Persistent) │  │                      │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘  │
│         │                 │                      │              │
│         ▼                 ▼                      ▼              │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Vector Store                           │  │
│  │  (Qdrant / Pinecone / pgvector)                          │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Memory Types

### Short-term Memory

Session-scoped memory that persists during a conversation:

```go
import "dev.helix.agent/internal/memory"

// Create memory manager
mm := memory.NewManager(&memory.Config{
    VectorStore:    vectorStore,
    EmbeddingModel: embeddingProvider,
    SessionTTL:     24 * time.Hour,
})

// Store short-term memory
err := mm.StoreShortTerm(ctx, &memory.Memory{
    SessionID: sessionID,
    Content:   "User prefers concise responses",
    Type:      memory.TypePreference,
    Metadata: map[string]interface{}{
        "source": "user_feedback",
    },
})
```

### Long-term Memory

Persistent memory that spans sessions:

```go
// Store long-term memory
err := mm.StoreLongTerm(ctx, &memory.Memory{
    UserID:  userID,
    Content: "User is a software engineer working on Go projects",
    Type:    memory.TypeFact,
    Metadata: map[string]interface{}{
        "confidence": 0.95,
        "source":     "conversation",
    },
})

// Retrieve long-term memories
memories, err := mm.RetrieveLongTerm(ctx, &memory.Query{
    UserID:    userID,
    Query:     "What does the user work on?",
    Limit:     5,
    Threshold: 0.7, // Similarity threshold
})
```

### Entity Graph

Knowledge graph of entities and relationships:

```go
// Add entity
err := mm.AddEntity(ctx, &memory.Entity{
    ID:    "entity-123",
    Type:  memory.EntityTypePerson,
    Name:  "John Doe",
    Properties: map[string]interface{}{
        "role":    "CTO",
        "company": "TechCorp",
    },
})

// Add relationship
err := mm.AddRelationship(ctx, &memory.Relationship{
    FromID:   "entity-123",
    ToID:     "entity-456",
    Type:     memory.RelationshipWorksAt,
    Properties: map[string]interface{}{
        "since": "2020",
    },
})

// Query entity graph
entities, err := mm.QueryEntities(ctx, &memory.EntityQuery{
    Type: memory.EntityTypePerson,
    Filters: map[string]interface{}{
        "role": "CTO",
    },
    IncludeRelationships: true,
})
```

## Memory Operations

### Storing Memories

```go
// Automatic memory extraction from conversation
memories, err := mm.ExtractMemories(ctx, &memory.ExtractionRequest{
    UserID:    userID,
    SessionID: sessionID,
    Messages: []memory.Message{
        {Role: "user", Content: "I'm a Python developer"},
        {Role: "assistant", Content: "Great! I'll tailor my responses for Python."},
    },
})

// Store extracted memories
for _, mem := range memories {
    err := mm.Store(ctx, mem)
    if err != nil {
        log.Printf("Failed to store memory: %v", err)
    }
}
```

### Retrieving Memories

```go
// Retrieve relevant memories for context
context, err := mm.BuildContext(ctx, &memory.ContextRequest{
    UserID:    userID,
    SessionID: sessionID,
    Query:     "Help me with a Python script",
    MaxTokens: 1000, // Limit context size
})

// Use in completion request
response, err := provider.Complete(ctx, &llm.Request{
    SystemPrompt: fmt.Sprintf(`You are a helpful assistant.

User Context:
%s`, context.ToString()),
    Prompt: "Help me with a Python script",
})
```

### Updating Memories

```go
// Update existing memory
err := mm.Update(ctx, memoryID, &memory.UpdateRequest{
    Content: "User is a senior Python developer with 10 years experience",
    Metadata: map[string]interface{}{
        "updated_at": time.Now(),
        "confidence": 0.98,
    },
})
```

### Deleting Memories

```go
// Delete specific memory
err := mm.Delete(ctx, memoryID)

// Delete all memories for user
err := mm.DeleteUserMemories(ctx, userID)

// Delete session memories
err := mm.DeleteSessionMemories(ctx, sessionID)
```

## Memory Types

| Type | Description | TTL | Scope |
|------|-------------|-----|-------|
| `fact` | Factual information about user | Permanent | User |
| `preference` | User preferences | Permanent | User |
| `context` | Conversation context | Session | Session |
| `instruction` | User instructions | Permanent | User |
| `interaction` | Interaction patterns | 30 days | User |
| `feedback` | User feedback | 90 days | User |

## Entity Types

| Type | Description | Example |
|------|-------------|---------|
| `person` | A person | "John Doe", "CEO" |
| `organization` | A company/org | "TechCorp", "OpenAI" |
| `project` | A project | "HelixAgent", "Website Redesign" |
| `concept` | A concept/topic | "Machine Learning", "API Design" |
| `location` | A place | "San Francisco", "Office A" |
| `event` | An event | "Product Launch", "Team Meeting" |

## Relationship Types

| Type | Description | Example |
|------|-------------|---------|
| `works_at` | Employment | John → TechCorp |
| `works_on` | Project involvement | John → HelixAgent |
| `knows` | Personal connection | John → Jane |
| `reports_to` | Hierarchy | Jane → John |
| `related_to` | General relation | AI → ML |
| `part_of` | Composition | API → System |

## Configuration

### Memory Manager Configuration

```yaml
memory:
  enabled: true

  # Vector store settings
  vector_store:
    type: qdrant  # or pinecone, pgvector
    host: localhost
    port: 6333
    collection: memories

  # Embedding settings
  embedding:
    provider: openai
    model: text-embedding-3-small
    dimensions: 1536

  # Memory settings
  short_term:
    enabled: true
    ttl: 24h
    max_per_session: 100

  long_term:
    enabled: true
    max_per_user: 10000

  entity_graph:
    enabled: true
    max_entities: 50000
    max_relationships: 200000

  # Extraction settings
  extraction:
    enabled: true
    provider: claude
    model: claude-3-5-sonnet
    confidence_threshold: 0.7
```

### Environment Variables

```bash
# Vector store
MEMORY_VECTOR_STORE=qdrant
QDRANT_HOST=localhost
QDRANT_PORT=6333

# Embedding
MEMORY_EMBEDDING_PROVIDER=openai
OPENAI_API_KEY=sk-...

# Extraction
MEMORY_EXTRACTION_PROVIDER=claude
CLAUDE_API_KEY=...
```

## API Endpoints

### Store Memory

```bash
POST /v1/memory
Content-Type: application/json
Authorization: Bearer <token>

{
  "user_id": "user-123",
  "session_id": "session-456",
  "content": "User prefers TypeScript over JavaScript",
  "type": "preference",
  "metadata": {
    "confidence": 0.9,
    "source": "explicit_statement"
  }
}
```

### Retrieve Memories

```bash
POST /v1/memory/search
Content-Type: application/json
Authorization: Bearer <token>

{
  "user_id": "user-123",
  "query": "programming language preferences",
  "limit": 5,
  "threshold": 0.7,
  "types": ["preference", "fact"]
}

Response:
{
  "memories": [
    {
      "id": "mem-789",
      "content": "User prefers TypeScript over JavaScript",
      "type": "preference",
      "similarity": 0.92,
      "created_at": "2026-01-20T10:00:00Z"
    }
  ]
}
```

### Build Context

```bash
POST /v1/memory/context
Content-Type: application/json
Authorization: Bearer <token>

{
  "user_id": "user-123",
  "session_id": "session-456",
  "query": "Help me with a coding task",
  "max_tokens": 1000
}

Response:
{
  "context": "User Context:\n- Prefers TypeScript over JavaScript\n- Senior developer with 10 years experience\n- Works on web applications\n\nRecent Session:\n- Discussed React components\n- Asked about testing strategies",
  "memories_used": 5,
  "token_count": 87
}
```

### Manage Entities

```bash
# Add entity
POST /v1/memory/entities
{
  "type": "person",
  "name": "John Doe",
  "properties": {
    "role": "CTO",
    "company": "TechCorp"
  }
}

# Add relationship
POST /v1/memory/relationships
{
  "from_id": "entity-123",
  "to_id": "entity-456",
  "type": "works_at"
}

# Query entities
POST /v1/memory/entities/search
{
  "type": "person",
  "query": "engineers at TechCorp",
  "include_relationships": true
}
```

## Best Practices

### 1. Memory Extraction

```go
// Good: Use structured extraction prompts
extraction := &memory.ExtractionConfig{
    Categories: []string{"facts", "preferences", "instructions"},
    MinConfidence: 0.7,
    DeduplicateExisting: true,
}

// Bad: Extract everything without filtering
// This leads to noise and high storage costs
```

### 2. Context Building

```go
// Good: Build focused context
context, err := mm.BuildContext(ctx, &memory.ContextRequest{
    Query:     currentQuery,
    Types:     []memory.Type{memory.TypeFact, memory.TypePreference},
    MaxTokens: 500, // Leave room for response
    Relevance: 0.8, // High relevance threshold
})

// Bad: Include all memories
// This wastes tokens and may include irrelevant info
```

### 3. Memory Lifecycle

```go
// Good: Set appropriate TTLs
memory := &memory.Memory{
    Content: "User asked about pricing",
    Type:    memory.TypeInteraction,
    TTL:     7 * 24 * time.Hour, // 7 days for interactions
}

// Good: Clean up old memories periodically
mm.Cleanup(ctx, &memory.CleanupConfig{
    OlderThan: 90 * 24 * time.Hour,
    Types:     []memory.Type{memory.TypeInteraction},
})
```

### 4. Privacy Considerations

```go
// Good: Respect user privacy settings
if user.PrivacySettings.DisableMemory {
    // Skip memory operations
    return
}

// Good: Allow memory deletion
err := mm.DeleteUserMemories(ctx, userID)

// Good: Anonymize sensitive data
memory.Content = anonymize(memory.Content)
```

### 5. Entity Graph Management

```go
// Good: Maintain entity consistency
err := mm.MergeEntities(ctx, duplicateID, canonicalID)

// Good: Prune orphaned entities
err := mm.PruneOrphanedEntities(ctx)

// Good: Validate relationships
err := mm.ValidateGraph(ctx)
```

## Troubleshooting

### High Memory Usage

```go
// Check memory counts
stats, err := mm.GetStats(ctx, userID)
log.Printf("Total memories: %d", stats.TotalMemories)
log.Printf("Entity count: %d", stats.EntityCount)

// Prune if needed
if stats.TotalMemories > 10000 {
    mm.Prune(ctx, &memory.PruneConfig{
        KeepMostRecent: 5000,
        KeepHighConfidence: true,
    })
}
```

### Slow Retrieval

```go
// Use indexes
mm.CreateIndex(ctx, "user_id", "type")

// Use pagination
memories, cursor, err := mm.Retrieve(ctx, &memory.Query{
    Limit:  100,
    Cursor: previousCursor,
})

// Use caching
mm.EnableCache(&memory.CacheConfig{
    TTL:     5 * time.Minute,
    MaxSize: 1000,
})
```

### Duplicate Memories

```go
// Enable deduplication
mm.SetDeduplicationConfig(&memory.DeduplicationConfig{
    Enabled:   true,
    Threshold: 0.95, // Similarity threshold
    Strategy:  memory.DedupeKeepNewest,
})
```

---

**Document Version**: 1.0
**Last Updated**: January 23, 2026
**Author**: Generated by Claude Code
