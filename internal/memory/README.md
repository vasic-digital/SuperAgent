# Package: memory

## Overview

The `memory` package provides a Mem0-style memory system for persistent AI memory with fact extraction, entity graphs, and cross-session recall. It enables AI agents to remember user preferences, facts, and context across conversations.

## Architecture

```
memory/
├── types.go         # Memory types and interfaces
├── manager.go       # Memory manager implementation
├── store_memory.go  # In-memory storage
└── memory_test.go   # Unit tests (94.6% coverage)
```

## Memory Types

| Type | Description |
|------|-------------|
| `episodic` | Conversation and event memories |
| `semantic` | Facts and knowledge |
| `procedural` | How-to knowledge |
| `working` | Short-term context |

## Key Types

### Memory

```go
type Memory struct {
    ID          string
    UserID      string
    SessionID   string
    Content     string
    Type        MemoryType
    Embedding   []float32
    Importance  float64
    AccessCount int
    LastAccess  time.Time
}
```

### Entity

```go
type Entity struct {
    ID         string
    Name       string
    Type       string  // person, place, thing, concept
    Properties map[string]interface{}
    Aliases    []string
}
```

### Relationship

```go
type Relationship struct {
    SourceID   string
    TargetID   string
    Type       string  // knows, works_at, located_in
    Strength   float64 // 0-1 confidence
}
```

## Usage

### Basic Memory Operations

```go
import "dev.helix.agent/internal/memory"

// Create manager
manager := memory.NewManager(config, logger)

// Add memory
mem := &memory.Memory{
    UserID:  "user-123",
    Content: "User prefers dark mode and concise responses",
    Type:    memory.MemoryTypeSemantic,
}
manager.Add(ctx, mem)

// Search memories
results, _ := manager.Search(ctx, "user preferences", &memory.SearchOptions{
    UserID: "user-123",
    TopK:   5,
})
```

### Entity Graph

```go
// Add entity
entity := &memory.Entity{
    Name: "John Smith",
    Type: "person",
    Properties: map[string]interface{}{
        "role": "engineer",
    },
}
manager.AddEntity(ctx, entity)

// Add relationship
rel := &memory.Relationship{
    SourceID: entity.ID,
    TargetID: companyEntity.ID,
    Type:     "works_at",
    Strength: 0.95,
}
manager.AddRelationship(ctx, rel)
```

### Conversation Memory

```go
// Process conversation for memory extraction
messages := []memory.Message{
    {Role: "user", Content: "I work at Acme Corp as a senior developer"},
    {Role: "assistant", Content: "Great! How can I help you today?"},
}

// Extract and store memories
manager.ProcessConversation(ctx, userID, messages)
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| MaxMemoriesPerUser | int | 1000 | Memory limit per user |
| EmbeddingDimension | int | 1536 | Vector embedding size |
| MinImportance | float64 | 0.3 | Threshold for storage |
| DecayRate | float64 | 0.01 | Memory decay rate |

## Testing

```bash
go test -v ./internal/memory/...
go test -cover ./internal/memory/...  # 94.6% coverage
```

## Dependencies

### Internal
- `internal/embedding` - Text embeddings
- `internal/vectordb` - Vector storage

### External
- Standard library only

## See Also

- [Mem0 Documentation](https://mem0.ai/)
- [Memory API Reference](../../docs/api/memory.md)
