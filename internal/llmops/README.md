# Package: llmops

## Overview

The `llmops` package provides enterprise LLMOps capabilities including prompt versioning, A/B testing, continuous evaluation, and alerting for production LLM systems.

## Architecture

```
llmops/
├── types.go        # Core types (prompts, experiments, evaluations)
├── prompts.go      # Prompt registry implementation
├── experiments.go  # A/B testing engine
├── evaluator.go    # Continuous evaluation system
├── integration.go  # System integration
└── llmops_test.go  # Unit tests (49.0% coverage)
```

## Features

- **Prompt Versioning**: Semantic versioning for prompt templates
- **A/B Testing**: Statistical experiment framework
- **Continuous Evaluation**: Automated quality monitoring
- **Alerting**: Regression detection and notifications

## Key Types

### PromptVersion

```go
type PromptVersion struct {
    ID          string
    Name        string
    Version     string  // Semantic version (e.g., "1.0.0")
    Content     string
    Variables   []PromptVariable
    IsActive    bool
}
```

### Experiment

```go
type Experiment struct {
    ID           string
    Name         string
    Variants     []*Variant
    TrafficSplit map[string]float64
    Status       ExperimentStatus
    Metrics      []string
    TargetMetric string
}
```

## Usage

### Prompt Management

```go
import "dev.helix.agent/internal/llmops"

// Create registry
registry := llmops.NewInMemoryPromptRegistry(logger)

// Create prompt version
prompt := &llmops.PromptVersion{
    Name:    "greeting",
    Version: "1.0.0",
    Content: "Hello, {{name}}! How can I help you today?",
    Variables: []llmops.PromptVariable{
        {Name: "name", Type: "string", Required: true},
    },
}
registry.Create(ctx, prompt)

// Render prompt
rendered, _ := registry.Render(ctx, "greeting", "1.0.0", map[string]interface{}{
    "name": "John",
})
// Output: "Hello, John! How can I help you today?"
```

### A/B Testing

```go
// Create experiment manager
manager := llmops.NewInMemoryExperimentManager(logger)

// Create experiment
exp := &llmops.Experiment{
    Name: "Prompt A/B Test",
    Variants: []*llmops.Variant{
        {Name: "Control", IsControl: true, PromptVersion: "1.0.0"},
        {Name: "Treatment", PromptVersion: "2.0.0"},
    },
    Metrics: []string{"quality", "latency"},
}
manager.Create(ctx, exp)

// Start experiment
manager.Start(ctx, exp.ID)

// Assign variant to user
variant, _ := manager.AssignVariant(ctx, exp.ID, userID)

// Record metrics
manager.RecordMetric(ctx, exp.ID, variant.ID, "quality", 0.85)
```

### Continuous Evaluation

```go
evaluator := llmops.NewInMemoryContinuousEvaluator(registry, provider, config, logger)

// Create evaluation dataset
dataset := &llmops.Dataset{
    Name: "Golden Set",
    Type: llmops.DatasetTypeGolden,
}
evaluator.CreateDataset(ctx, dataset)

// Add samples
evaluator.AddSamples(ctx, dataset.ID, []*llmops.DatasetSample{
    {Input: "What is 2+2?", ExpectedOutput: "4"},
})

// Run evaluation
run, _ := evaluator.CreateRun(ctx, &llmops.EvaluationRun{
    Dataset:    dataset.ID,
    PromptName: "math-assistant",
    Metrics:    []string{"accuracy"},
})
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| EnableAutoEvaluation | bool | true | Auto-run evaluations |
| MinSamplesForSignif | int | 100 | Min samples for statistical significance |
| AlertThresholds | map | varies | Threshold for alerts |

## Testing

```bash
go test -v ./internal/llmops/...
go test -cover ./internal/llmops/...
```

## Dependencies

### Internal
- `internal/llm` - LLM providers
- `internal/database` - Data persistence

### External
- Standard library only

## See Also

- [Prompt Engineering Guide](../../docs/guides/prompt-engineering.md)
- [A/B Testing Best Practices](../../docs/guides/ab-testing.md)
