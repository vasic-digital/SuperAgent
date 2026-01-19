# Debate Package

The debate package provides the AI Debate System and Lesson Banking functionality for HelixAgent's ensemble LLM architecture.

## Overview

This package implements:

- **Lesson Bank**: Reusable knowledge capture and retrieval from AI debates
- **Lesson Extraction**: Automatic extraction of insights from debate conclusions
- **Semantic Search**: Vector similarity-based lesson retrieval
- **Lesson Promotion**: Automatic promotion of successful lessons

## Key Components

### Lesson Bank

The core system for storing and retrieving learned knowledge:

```go
config := &debate.LessonBankConfig{
    MaxLessons:           1000,
    MinConfidence:        0.7,
    EnableSemanticSearch: true,
    SimilarityThreshold:  0.85,
    ExpirationDays:       90,
    EnableAutoPromotion:  true,
}

bank := debate.NewLessonBank(config, embedder, storage)
```

### Lesson Structure

```go
type Lesson struct {
    ID              string
    Title           string
    Content         string
    Category        LessonCategory
    Tags            []string
    Confidence      float64
    SuccessRate     float64
    ApplicationCount int
    CreatedAt       time.Time
    LastApplied     time.Time
    SourceDebateID  string
    Embedding       []float64
}
```

### Lesson Categories

| Category | Description |
|----------|-------------|
| `CategoryProgramming` | Programming patterns and best practices |
| `CategoryArchitecture` | System design and architecture decisions |
| `CategoryDebugging` | Debugging strategies and solutions |
| `CategorySecurity` | Security considerations and fixes |
| `CategoryPerformance` | Performance optimization techniques |
| `CategoryBestPractice` | General best practices |
| `CategoryDomainKnowledge` | Domain-specific knowledge |

## Usage

### Adding Lessons

```go
lesson := &debate.Lesson{
    Title:      "Error Handling in Go",
    Content:    "Always wrap errors with context using fmt.Errorf",
    Category:   debate.CategoryBestPractice,
    Tags:       []string{"go", "errors", "best-practice"},
    Confidence: 0.95,
}

err := bank.AddLesson(ctx, lesson)
```

### Searching Lessons

```go
// Text search
lessons, err := bank.Search(ctx, "error handling")

// Semantic search (requires embeddings)
lessons, err := bank.SemanticSearch(ctx, query, 5)

// Category filter
lessons, err := bank.GetByCategory(ctx, debate.CategoryProgramming)
```

### Applying Lessons

```go
// Record successful application
err := bank.RecordOutcome(ctx, lessonID, true)

// Get applicable lessons for context
lessons, err := bank.GetApplicableLessons(ctx, debateContext)
```

### Automatic Extraction

```go
// Extract lessons from debate conclusion
extracted, err := bank.ExtractFromDebate(ctx, debate)

// Auto-promote successful lessons
promoted, err := bank.PromoteSuccessfulLessons(ctx)
```

## Configuration

```go
type LessonBankConfig struct {
    MaxLessons           int     // Maximum stored lessons
    MinConfidence        float64 // Minimum confidence to store (0-1)
    EnableSemanticSearch bool    // Enable vector search
    SimilarityThreshold  float64 // Semantic similarity threshold
    ExpirationDays       int     // Lesson expiration period
    EnableAutoPromotion  bool    // Auto-promote successful lessons
    PromotionThreshold   int     // Applications before promotion
    PromotionSuccessRate float64 // Required success rate
}
```

## Storage Backends

### In-Memory Storage

```go
storage := debate.NewInMemoryLessonStorage()
bank := debate.NewLessonBank(config, embedder, storage)
```

### Persistent Storage (Redis/PostgreSQL)

```go
storage := debate.NewRedisLessonStorage(redisClient)
// or
storage := debate.NewPostgresLessonStorage(db)
```

## Integration with Debate System

The Lesson Bank integrates with the main debate service:

1. **Pre-Debate**: Retrieve relevant lessons to inform debate
2. **During Debate**: Apply lessons to enhance responses
3. **Post-Debate**: Extract new lessons from conclusions
4. **Feedback Loop**: Record outcomes to improve future retrieval

## Testing

```bash
# Run all debate tests
go test -v ./internal/debate/...

# Run lesson bank tests
go test -v -run TestLessonBank ./internal/debate/
```

## Metrics

The package tracks:

- Total lessons stored
- Lessons by category
- Application success rates
- Search hit rates
- Extraction success rates

## See Also

- `internal/services/debate_service.go` - Main debate orchestration
- `internal/services/debate_dialogue.go` - Debate formatting
- `internal/services/debate_team_config.go` - Team configuration
