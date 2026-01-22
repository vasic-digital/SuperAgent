# Debate Knowledge Package

This package provides knowledge repository, lesson extraction, and cross-debate learning for the AI Debate Orchestrator Framework.

## Overview

The knowledge package implements a learning system that extracts insights from debates, stores lessons, identifies patterns, and applies learnings to improve future debates.

## Components

### Repository (`repository.go`)

Central storage for debate knowledge:

```go
repo := knowledge.NewRepository(storageConfig)
err := repo.Store(ctx, knowledge.Lesson{
    DebateID:    "debate-001",
    Topic:       "AI Ethics",
    Insight:     "Multi-stakeholder perspectives improve outcomes",
    Confidence:  0.85,
})
```

### Learning System (`learning.go`)

Extracts lessons from completed debates:

```go
learner := knowledge.NewLearner(repo)
lessons, err := learner.ExtractLessons(ctx, debateResult)
```

### Pattern Recognition

Identifies recurring patterns across debates:

```go
patterns := repo.FindPatterns(ctx, knowledge.PatternQuery{
    Topic:      "technology",
    MinOccurrences: 3,
})
```

### Integration (`integration.go`)

Integrates knowledge into active debates:

```go
integrator := knowledge.NewIntegrator(repo)
relevantLessons := integrator.GetRelevantLessons(ctx, currentDebate)
```

## Architecture

```
┌─────────────────────────────────────────────┐
│            Knowledge Repository             │
│  ┌─────────────────────────────────────┐   │
│  │              Storage                │   │
│  │  ┌───────┐ ┌───────┐ ┌───────┐     │   │
│  │  │Lessons│ │Pattern│ │Context│     │   │
│  │  └───────┘ └───────┘ └───────┘     │   │
│  └─────────────────────────────────────┘   │
│  ┌─────────────────────────────────────┐   │
│  │           Learning Layer            │   │
│  │  Extract → Classify → Store         │   │
│  └─────────────────────────────────────┘   │
│  ┌─────────────────────────────────────┐   │
│  │          Integration Layer          │   │
│  │  Query → Match → Apply              │   │
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

## Knowledge Types

| Type | Description | Example |
|------|-------------|---------|
| `Lesson` | Specific insight | "Technical debates benefit from domain experts" |
| `Pattern` | Recurring observation | "Consensus reached faster with moderator" |
| `Strategy` | Effective approach | "Devil's advocate improves thoroughness" |

## Usage

```go
import "dev.helix.agent/internal/debate/knowledge"

// Create knowledge repository
repo := knowledge.NewRepository(knowledge.Config{
    StorageType: knowledge.StorageInMemory,
    MaxLessons:  10000,
})

// After debate completion, extract lessons
learner := knowledge.NewLearner(repo)
lessons, err := learner.ExtractLessons(ctx, debateResult)

// Store lessons
for _, lesson := range lessons {
    repo.Store(ctx, lesson)
}

// In future debates, apply learnings
integrator := knowledge.NewIntegrator(repo)
relevant := integrator.GetRelevantLessons(ctx, knowledge.Query{
    Topic:      newDebate.Topic,
    MinConfidence: 0.7,
})
```

## Testing

```bash
go test -v ./internal/debate/knowledge/...
```

## Files

- `repository.go` - Knowledge storage
- `repository_test.go` - Repository tests
- `learning.go` - Lesson extraction
- `learning_test.go` - Learning tests
- `integration.go` - Cross-debate integration
- `integration_test.go` - Integration tests
- `knowledge_extended_test.go` - Extended tests
