# Memory System (Mem0-Style)

## Overview

The Memory System in HelixAgent provides persistent AI memory capabilities inspired by Mem0. It enables fact extraction, entity recognition, relationship tracking, and intelligent memory retrieval across conversations and sessions.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Memory Manager                              │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Memory Types                           │  │
│  │  ┌───────────┐ ┌───────────┐ ┌────────────┐ ┌─────────┐ │  │
│  │  │ Episodic  │ │ Semantic  │ │ Procedural │ │ Working │ │  │
│  │  │ (Events)  │ │ (Facts)   │ │ (How-to)   │ │ (Short) │ │  │
│  │  └───────────┘ └───────────┘ └────────────┘ └─────────┘ │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                   Entity Graph                            │  │
│  │  ┌─────────┐      ┌─────────┐      ┌─────────┐          │  │
│  │  │ Person  │─────▶│ Project │◀─────│  Org    │          │  │
│  │  └─────────┘      └─────────┘      └─────────┘          │  │
│  │       │                │                │                │  │
│  │       └────────────────┼────────────────┘                │  │
│  │                        ▼                                  │  │
│  │                 ┌───────────┐                            │  │
│  │                 │   Topic   │                            │  │
│  │                 └───────────┘                            │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                   Storage Layer                           │  │
│  │  ├─ Vector Store (embeddings)                            │  │
│  │  ├─ Graph Store (relationships)                          │  │
│  │  └─ Key-Value Store (metadata)                           │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Memory Types

### 1. Episodic Memory

Stores conversation events and experiences:

```go
import "dev.helix.agent/internal/memory"

manager := memory.NewManager(config)

// Store an episodic memory
err := manager.StoreEpisodic(ctx, &memory.EpisodicMemory{
    UserID:     "user-123",
    SessionID:  "session-456",
    Content:    "User discussed implementing a caching layer",
    Timestamp:  time.Now(),
    Importance: 0.8,
    Emotions:   []string{"curious", "focused"},
})
```

### 2. Semantic Memory

Stores facts and knowledge:

```go
// Store a semantic memory (fact)
err := manager.StoreSemantic(ctx, &memory.SemanticMemory{
    UserID:     "user-123",
    Fact:       "User prefers Go over Python for backend development",
    Category:   "preferences",
    Confidence: 0.95,
    Sources:    []string{"conversation-789", "conversation-012"},
})
```

### 3. Procedural Memory

Stores learned procedures and workflows:

```go
// Store a procedural memory
err := manager.StoreProcedural(ctx, &memory.ProceduralMemory{
    UserID:      "user-123",
    Procedure:   "deploy-to-production",
    Steps: []string{
        "Run tests: make test",
        "Build image: make docker-build",
        "Deploy: kubectl apply -f k8s/",
    },
    Conditions: []string{"all tests pass", "no critical vulnerabilities"},
})
```

### 4. Working Memory

Short-term context during conversations:

```go
// Store working memory
err := manager.StoreWorking(ctx, &memory.WorkingMemory{
    SessionID:  "session-456",
    Context:    "Currently debugging a race condition in the cache package",
    Variables: map[string]interface{}{
        "current_file": "cache.go",
        "line_number":  142,
    },
    ExpiresAt:  time.Now().Add(30 * time.Minute),
})
```

## Entity Management

### Entity Extraction

```go
extractor := memory.NewEntityExtractor(llmProvider)

// Extract entities from text
entities, err := extractor.Extract(ctx,
    "John from Acme Corp is working on Project Alpha with Sarah")

// Returns:
// - Person: John (organization: Acme Corp)
// - Organization: Acme Corp
// - Project: Project Alpha
// - Person: Sarah
```

### Entity Graph

```go
graph := memory.NewEntityGraph(storage)

// Add relationship
err := graph.AddRelationship(ctx, &memory.Relationship{
    SourceID:   "entity-john",
    TargetID:   "entity-project-alpha",
    Type:       "works_on",
    Strength:   0.9,
    Properties: map[string]interface{}{
        "role": "lead developer",
        "since": "2024-01-15",
    },
})

// Query relationships
relationships, err := graph.GetRelationships(ctx, "entity-john", &memory.RelationshipFilter{
    Types:       []string{"works_on", "collaborates_with"},
    MinStrength: 0.5,
})
```

### Entity Types

```go
type Entity struct {
    ID          string                 // Unique identifier
    Type        EntityType             // Person, Organization, Project, Topic, etc.
    Name        string                 // Display name
    Aliases     []string               // Alternative names
    Properties  map[string]interface{} // Custom properties
    Embedding   []float64              // Vector embedding
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type EntityType string

const (
    EntityTypePerson       EntityType = "person"
    EntityTypeOrganization EntityType = "organization"
    EntityTypeProject      EntityType = "project"
    EntityTypeTopic        EntityType = "topic"
    EntityTypeLocation     EntityType = "location"
    EntityTypeConcept      EntityType = "concept"
)
```

## Memory Retrieval

### Semantic Search

```go
// Search memories by semantic similarity
memories, err := manager.Search(ctx, &memory.SearchQuery{
    UserID:     "user-123",
    Query:      "previous discussions about caching",
    Types:      []memory.MemoryType{memory.TypeEpisodic, memory.TypeSemantic},
    TopK:       10,
    TimeRange: &memory.TimeRange{
        Start: time.Now().AddDate(0, -1, 0), // Last month
        End:   time.Now(),
    },
})
```

### Contextual Retrieval

```go
// Get relevant context for a new conversation
context, err := manager.GetContext(ctx, &memory.ContextRequest{
    UserID:         "user-123",
    CurrentMessage: "Let's continue working on the cache optimization",
    MaxTokens:      2000,
    IncludeTypes: []memory.MemoryType{
        memory.TypeSemantic,   // User preferences and facts
        memory.TypeProcedural, // Known workflows
        memory.TypeEpisodic,   // Recent relevant conversations
    },
})
```

### Entity-Based Retrieval

```go
// Get memories related to an entity
memories, err := manager.GetByEntity(ctx, "entity-project-alpha", &memory.EntityQuery{
    IncludeRelated: true,  // Include memories about related entities
    Depth:          2,     // Relationship traversal depth
    TopK:           20,
})
```

## Memory Importance and Decay

### Importance Scoring

```go
scorer := memory.NewImportanceScorer(llmProvider)

// Score memory importance
importance, err := scorer.Score(ctx, &memory.ScoringInput{
    Content:    "User mentioned they're the CTO of their company",
    UserID:     "user-123",
    ExistingMemories: existingMemories,
})
// importance.Score = 0.9 (high - role information)
// importance.Reasoning = "Executive role information is important for context"
```

### Memory Decay

```go
// Configure decay settings
decayConfig := &memory.DecayConfig{
    BaseDecayRate:    0.1,    // 10% decay per period
    DecayPeriod:      24 * time.Hour,
    MinImportance:    0.2,    // Minimum importance before deletion
    BoostOnAccess:    0.1,    // Importance boost when accessed
    ImportanceFloor:  0.5,    // Importance above which decay is slower
}

// Apply decay
err := manager.ApplyDecay(ctx, decayConfig)
```

### Memory Consolidation

```go
// Consolidate similar memories
consolidator := memory.NewConsolidator(llmProvider)

consolidated, err := consolidator.Consolidate(ctx, &memory.ConsolidationRequest{
    UserID:           "user-123",
    SimilarityThreshold: 0.85,
    MaxMergeCount:    5,
})

// Merges similar memories into stronger, consolidated memories
```

## Configuration

```yaml
memory:
  # Storage backends
  storage:
    vector_store:
      type: "qdrant"
      url: "http://localhost:6333"
      collection: "memories"
    graph_store:
      type: "neo4j"
      url: "bolt://localhost:7687"
    kv_store:
      type: "redis"
      url: "redis://localhost:6379"

  # Memory settings
  settings:
    max_working_memory: 50
    default_importance: 0.5
    decay_enabled: true
    consolidation_enabled: true

  # Embedding settings
  embedding:
    provider: "openai"
    model: "text-embedding-3-small"
    dimension: 1536
```

## API Usage

### REST Endpoints

```http
# Store a memory
POST /v1/memory
{
  "user_id": "user-123",
  "type": "semantic",
  "content": "User prefers dark mode in all applications",
  "metadata": {
    "category": "preferences"
  }
}

# Search memories
POST /v1/memory/search
{
  "user_id": "user-123",
  "query": "user preferences",
  "top_k": 10
}

# Get context
POST /v1/memory/context
{
  "user_id": "user-123",
  "current_message": "What are my usual settings?",
  "max_tokens": 2000
}
```

## Testing

```bash
# Run memory tests
go test -v ./internal/memory/...

# Run with coverage
go test -cover ./internal/memory/...

# Run integration tests (requires storage backends)
go test -v -tags=integration ./internal/memory/...
```

## Key Files

| File | Description |
|------|-------------|
| `internal/memory/manager.go` | Main memory manager |
| `internal/memory/types.go` | Memory type definitions |
| `internal/memory/entity.go` | Entity extraction and graph |
| `internal/memory/search.go` | Memory search and retrieval |
| `internal/memory/decay.go` | Importance decay logic |
| `internal/memory/consolidation.go` | Memory consolidation |
| `internal/memory/memory_test.go` | Comprehensive tests |

## See Also

- [RAG System](./RAG_SYSTEM.md)
- [Entity Extraction](./ENTITY_EXTRACTION.md)
- [Vector Database](./VECTOR_DATABASE.md)
