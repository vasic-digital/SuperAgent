# HelixAgent Memory System

Mem0-style persistent memory with entity graphs for contextual AI interactions across sessions.

---

## Overview

HelixAgent implements a sophisticated memory system that enables AI to remember and learn from past interactions:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Memory System                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐         │
│  │   Short-term     │  │   Long-term      │  │   Entity Graph   │         │
│  │   Memory         │  │   Memory         │  │   (Knowledge)    │         │
│  │                  │  │                  │  │                  │         │
│  │  Session-scoped  │  │  User-scoped     │  │  Relationships   │         │
│  │  Context, recent │  │  Facts, prefs    │  │  Entities        │         │
│  │  interactions    │  │  instructions    │  │  Inference       │         │
│  └────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘         │
│           │                     │                      │                   │
│           └─────────────────────┼──────────────────────┘                   │
│                                 │                                          │
│                    ┌────────────▼────────────┐                             │
│                    │      Vector Store       │                             │
│                    │  (Qdrant/Pinecone/      │                             │
│                    │   Milvus/pgvector)      │                             │
│                    └─────────────────────────┘                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Key Benefits:**
- **Persistent context** across sessions
- **Personalized responses** based on user history
- **Knowledge graph** for relationship understanding
- **Semantic search** for relevant memory retrieval

---

## Memory Types

### Short-term Memory

Session-scoped memory that provides immediate context:

| Characteristic | Value |
|----------------|-------|
| **Scope** | Session |
| **TTL** | 24 hours (configurable) |
| **Max items** | 100 per session |
| **Use cases** | Current conversation, recent context |

**Example:**

```go
import "dev.helix.agent/internal/memory"

mm := memory.NewManager(config)

// Store short-term memory
err := mm.StoreShortTerm(ctx, &memory.Memory{
    SessionID: sessionID,
    Content:   "User is asking about Go programming",
    Type:      memory.TypeContext,
    Metadata: map[string]interface{}{
        "topic": "golang",
    },
})

// Retrieve session context
memories, err := mm.RetrieveShortTerm(ctx, sessionID, 10)
```

---

### Long-term Memory

Persistent memory that spans sessions:

| Characteristic | Value |
|----------------|-------|
| **Scope** | User |
| **TTL** | Permanent (some types have limits) |
| **Max items** | 10,000 per user |
| **Use cases** | User preferences, facts, instructions |

**Memory Subtypes:**

| Type | Description | TTL | Example |
|------|-------------|-----|---------|
| `fact` | Factual information | Permanent | "User is a software engineer" |
| `preference` | User preferences | Permanent | "User prefers concise responses" |
| `instruction` | User instructions | Permanent | "Always include code examples" |
| `interaction` | Interaction patterns | 30 days | "User often asks about debugging" |
| `feedback` | User feedback | 90 days | "User liked the detailed explanation" |

**Example:**

```go
// Store long-term memory
err := mm.StoreLongTerm(ctx, &memory.Memory{
    UserID:  userID,
    Content: "User is a senior Python developer with 10 years experience",
    Type:    memory.TypeFact,
    Metadata: map[string]interface{}{
        "confidence": 0.95,
        "source":     "explicit_statement",
    },
})

// Retrieve relevant memories
memories, err := mm.RetrieveLongTerm(ctx, &memory.Query{
    UserID:    userID,
    Query:     "What programming languages does the user know?",
    Limit:     5,
    Threshold: 0.7,
})
```

---

### Entity Graph

Knowledge graph storing entities and relationships:

**Entity Types:**

| Type | Description | Examples |
|------|-------------|----------|
| `person` | Individual people | "John Doe", "CEO" |
| `organization` | Companies, teams | "TechCorp", "Engineering Team" |
| `project` | Projects, products | "HelixAgent", "Mobile App" |
| `concept` | Ideas, technologies | "Machine Learning", "REST API" |
| `location` | Places | "San Francisco", "Office A" |
| `event` | Events, meetings | "Product Launch", "Sprint Review" |

**Relationship Types:**

| Type | Description | Example |
|------|-------------|---------|
| `works_at` | Employment | John works at TechCorp |
| `works_on` | Project involvement | Jane works on HelixAgent |
| `knows` | Personal connection | John knows Jane |
| `reports_to` | Hierarchy | Jane reports to John |
| `related_to` | General relation | ML related to AI |
| `part_of` | Composition | API part of Backend |

**Example:**

```go
// Add entity
err := mm.AddEntity(ctx, &memory.Entity{
    ID:   "entity-john",
    Type: memory.EntityTypePerson,
    Name: "John Doe",
    Properties: map[string]interface{}{
        "role":    "CTO",
        "company": "TechCorp",
        "email":   "john@techcorp.com",
    },
})

// Add relationship
err := mm.AddRelationship(ctx, &memory.Relationship{
    FromID: "entity-john",
    ToID:   "entity-techcorp",
    Type:   memory.RelationshipWorksAt,
    Properties: map[string]interface{}{
        "since":    "2020",
        "position": "CTO",
    },
})

// Query relationships
entities, err := mm.QueryEntities(ctx, &memory.EntityQuery{
    Type: memory.EntityTypePerson,
    Filters: map[string]interface{}{
        "role": "CTO",
    },
    IncludeRelationships: true,
    Depth:                2,
})
```

---

## Memory Operations

### Automatic Extraction

Extract memories from conversations automatically:

```go
// Extract memories from conversation
memories, err := mm.ExtractMemories(ctx, &memory.ExtractionRequest{
    UserID:    userID,
    SessionID: sessionID,
    Messages: []memory.Message{
        {Role: "user", Content: "I'm a Python developer specializing in data science"},
        {Role: "assistant", Content: "Great! I'll tailor my responses for Python and data science."},
        {Role: "user", Content: "I work at DataCorp as a senior engineer"},
    },
})

// memories will contain:
// - Fact: "User is a Python developer"
// - Fact: "User specializes in data science"
// - Fact: "User works at DataCorp"
// - Fact: "User is a senior engineer"
```

**Extraction Configuration:**

```yaml
memory:
  extraction:
    enabled: true
    provider: claude
    model: claude-3-5-sonnet
    confidence_threshold: 0.7
    categories:
      - facts
      - preferences
      - instructions
    deduplicate: true
```

---

### Context Building

Build relevant context for prompts:

```go
// Build context from memories
context, err := mm.BuildContext(ctx, &memory.ContextRequest{
    UserID:    userID,
    SessionID: sessionID,
    Query:     "Help me optimize this Python function",
    MaxTokens: 500,
    Types:     []memory.Type{memory.TypeFact, memory.TypePreference},
    Relevance: 0.7,
})

// Use in completion
response, err := provider.Complete(ctx, &llm.Request{
    SystemPrompt: fmt.Sprintf(`You are a helpful assistant.

User Context:
%s`, context.ToString()),
    Prompt: "Help me optimize this Python function",
})
```

**Context Output Example:**

```
User Context:
- User is a senior Python developer with 10 years experience
- User specializes in data science and ML
- User prefers concise, code-focused responses
- User works on performance-critical applications

Recent Session:
- Discussed list comprehension optimization
- Asked about NumPy vectorization
```

---

### Memory Search

Semantic search across memories:

```go
// Search memories
results, err := mm.Search(ctx, &memory.SearchRequest{
    UserID:     userID,
    Query:      "programming experience",
    Types:      []memory.Type{memory.TypeFact},
    Limit:      10,
    Threshold:  0.6,
    DateRange: &memory.DateRange{
        Start: time.Now().AddDate(-1, 0, 0),
        End:   time.Now(),
    },
})

for _, result := range results {
    fmt.Printf("Memory: %s (similarity: %.2f)\n",
        result.Content, result.Similarity)
}
```

---

### Memory Updates

Update and manage memories:

```go
// Update memory
err := mm.Update(ctx, memoryID, &memory.UpdateRequest{
    Content: "User is now a principal Python developer",
    Metadata: map[string]interface{}{
        "updated_at": time.Now(),
        "confidence": 0.98,
        "source":     "explicit_update",
    },
})

// Delete specific memory
err := mm.Delete(ctx, memoryID)

// Delete all user memories
err := mm.DeleteUserMemories(ctx, userID)

// Delete session memories
err := mm.DeleteSessionMemories(ctx, sessionID)
```

---

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

### Search Memories

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
```

**Response:**

```json
{
  "memories": [
    {
      "id": "mem-789",
      "content": "User prefers TypeScript over JavaScript",
      "type": "preference",
      "similarity": 0.92,
      "created_at": "2026-01-20T10:00:00Z",
      "metadata": {
        "confidence": 0.9
      }
    }
  ],
  "total": 1
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
```

**Response:**

```json
{
  "context": "User Context:\n- Prefers TypeScript over JavaScript\n- Senior developer with 10 years experience\n- Works on web applications",
  "memories_used": 5,
  "token_count": 87
}
```

### Entity Management

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
  "include_relationships": true,
  "depth": 2
}
```

---

## Configuration

### Full Configuration

```yaml
memory:
  enabled: true

  # Vector store for embeddings
  vector_store:
    type: qdrant           # qdrant, pinecone, milvus, pgvector
    host: localhost
    port: 6333
    collection: memories
    similarity_metric: cosine

  # Embedding provider
  embedding:
    provider: openai
    model: text-embedding-3-small
    dimensions: 1536
    batch_size: 100

  # Short-term memory settings
  short_term:
    enabled: true
    ttl: 24h
    max_per_session: 100

  # Long-term memory settings
  long_term:
    enabled: true
    max_per_user: 10000

  # Entity graph settings
  entity_graph:
    enabled: true
    max_entities: 50000
    max_relationships: 200000
    store: neo4j           # neo4j, memory

  # Extraction settings
  extraction:
    enabled: true
    provider: claude
    model: claude-3-5-sonnet
    confidence_threshold: 0.7
    deduplicate: true

  # Cache settings
  cache:
    enabled: true
    ttl: 5m
    max_size: 1000
```

### Environment Variables

```bash
# Enable memory system
MEMORY_ENABLED=true

# Vector store
MEMORY_VECTOR_STORE=qdrant
QDRANT_HOST=localhost
QDRANT_PORT=6333

# Embedding
MEMORY_EMBEDDING_PROVIDER=openai
OPENAI_API_KEY=sk-...

# Entity graph
MEMORY_ENTITY_GRAPH_ENABLED=true
NEO4J_URI=bolt://localhost:7687
NEO4J_PASSWORD=...

# Extraction
MEMORY_EXTRACTION_PROVIDER=claude
CLAUDE_API_KEY=...
```

---

## Best Practices

### Memory Extraction

```go
// Good: Use structured extraction with filtering
extraction := &memory.ExtractionConfig{
    Categories:          []string{"facts", "preferences"},
    MinConfidence:       0.7,
    DeduplicateExisting: true,
    MaxPerConversation:  10,
}

// Bad: Extract everything
// This creates noise and wastes storage
```

### Context Building

```go
// Good: Build focused context
context, err := mm.BuildContext(ctx, &memory.ContextRequest{
    Query:     currentQuery,
    Types:     []memory.Type{memory.TypeFact, memory.TypePreference},
    MaxTokens: 500,
    Relevance: 0.8,
})

// Bad: Include all memories
// This wastes tokens and may include irrelevant info
```

### Memory Lifecycle

```go
// Good: Set appropriate TTLs
memory := &memory.Memory{
    Content: "User asked about pricing",
    Type:    memory.TypeInteraction,
    TTL:     7 * 24 * time.Hour, // 7 days
}

// Good: Clean up old memories periodically
mm.Cleanup(ctx, &memory.CleanupConfig{
    OlderThan: 90 * 24 * time.Hour,
    Types:     []memory.Type{memory.TypeInteraction},
})
```

### Privacy

```go
// Good: Respect user privacy
if user.PrivacySettings.DisableMemory {
    return // Skip memory operations
}

// Good: Allow memory deletion
err := mm.DeleteUserMemories(ctx, userID)

// Good: Anonymize before long-term storage
memory.Content = anonymize(memory.Content)
```

### Entity Graph

```go
// Good: Maintain consistency
err := mm.MergeEntities(ctx, duplicateID, canonicalID)

// Good: Prune orphaned entities
err := mm.PruneOrphanedEntities(ctx)

// Good: Validate graph integrity
err := mm.ValidateGraph(ctx)
```

---

## Troubleshooting

### High Memory Usage

```go
// Check memory stats
stats, err := mm.GetStats(ctx, userID)
fmt.Printf("Total memories: %d\n", stats.TotalMemories)
fmt.Printf("Short-term: %d\n", stats.ShortTermCount)
fmt.Printf("Long-term: %d\n", stats.LongTermCount)
fmt.Printf("Entities: %d\n", stats.EntityCount)

// Prune if needed
if stats.TotalMemories > 10000 {
    mm.Prune(ctx, &memory.PruneConfig{
        KeepMostRecent:     5000,
        KeepHighConfidence: true,
    })
}
```

### Slow Retrieval

```go
// Create indexes
mm.CreateIndex(ctx, "user_id", "type", "created_at")

// Use pagination
memories, cursor, err := mm.Retrieve(ctx, &memory.Query{
    Limit:  100,
    Cursor: previousCursor,
})

// Enable caching
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
    Threshold: 0.95, // Cosine similarity
    Strategy:  memory.DedupeKeepNewest,
})

// Manual deduplication
duplicates, err := mm.FindDuplicates(ctx, userID)
for _, dup := range duplicates {
    mm.Delete(ctx, dup.ID)
}
```

---

## Challenges

Validate memory system:

```bash
# Run memory system challenge
./challenges/scripts/memory_system_challenge.sh

# Expected: 14 tests
# - Short-term storage/retrieval
# - Long-term storage/retrieval
# - Memory search
# - Context building
# - Entity graph operations
# - Relationship queries
# - Memory extraction
# - Deduplication
# - Privacy controls
# - Performance benchmarks
```

---

## Related Documentation

- [Architecture](./ARCHITECTURE.md) - System architecture
- [Big Data](./BIGDATA.md) - Distributed memory and knowledge graphs
- [API Reference](/docs/api/README.md) - Full API documentation

---

**Last Updated**: February 2026
**Version**: 1.0.0
