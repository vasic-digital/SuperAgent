# Debate Cognitive Package

This package provides reasoning patterns and cognitive strategies for the AI Debate Orchestrator Framework.

## Overview

The cognitive package implements various reasoning patterns that agents use to analyze topics, construct arguments, and evaluate positions during debates.

## Components

### Cognitive Planning (`cognitive_planning.go`)

Implements structured reasoning approaches:

```go
planner := cognitive.NewCognitivePlanner(config)
plan := planner.CreatePlan(ctx, topic, constraints)
```

### Reasoning Patterns

| Pattern | Description | Use Case |
|---------|-------------|----------|
| `Analytical` | Break down complex topics | Deep analysis |
| `Comparative` | Compare and contrast | Multiple options |
| `Causal` | Cause and effect | Impact analysis |
| `Evaluative` | Assess quality/value | Judgment tasks |
| `Creative` | Generate new ideas | Brainstorming |

### Analysis Strategies

```go
analyzer := cognitive.NewAnalyzer(pattern)
analysis := analyzer.Analyze(ctx, input)
```

## Architecture

```
┌─────────────────────────────────────────────┐
│              Cognitive Engine               │
│  ┌─────────────────────────────────────┐   │
│  │         Reasoning Patterns          │   │
│  │  ┌──────┐ ┌──────┐ ┌──────────┐    │   │
│  │  │Analyt│ │Compar│ │Evaluative│    │   │
│  │  └──────┘ └──────┘ └──────────┘    │   │
│  └─────────────────────────────────────┘   │
│  ┌─────────────────────────────────────┐   │
│  │          Planning Layer             │   │
│  │  Strategy → Steps → Execution       │   │
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

## Usage

```go
import "dev.helix.agent/internal/debate/cognitive"

// Create cognitive planner
planner := cognitive.NewCognitivePlanner(cognitive.Config{
    DefaultPattern: cognitive.PatternAnalytical,
    MaxDepth:       3,
})

// Generate reasoning plan
plan, err := planner.CreatePlan(ctx, cognitive.PlanRequest{
    Topic:       "AI governance frameworks",
    Constraints: []string{"Consider ethical implications"},
    Goals:       []string{"Identify key stakeholders"},
})

// Execute reasoning
result := planner.Execute(ctx, plan)
```

## Testing

```bash
go test -v ./internal/debate/cognitive/...
```

## Files

- `cognitive_planning.go` - Core planning implementation
- `cognitive_planning_test.go` - Unit tests
