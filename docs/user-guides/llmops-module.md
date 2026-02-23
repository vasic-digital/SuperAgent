# LLMOps Module User Guide

**Module**: `digital.vasic.llmops`
**Directory**: `LLMOps/`
**Phase**: 5 (AI/ML)

## Overview

The LLMOps module provides the operational infrastructure for managing LLMs in production. It covers
four core concerns:

1. **Continuous Evaluation** — Run evaluation pipelines against golden/regression/benchmark datasets
   on a schedule and compare runs to detect regressions.
2. **A/B Experiment Management** — Define experiments with multiple model/prompt variants, assign
   traffic splits, record per-variant metrics, and determine statistical significance.
3. **Dataset Management** — Create and version datasets (golden, regression, benchmark, user-
   generated) with typed samples.
4. **Prompt Versioning** — Store, activate, and render versioned prompt templates with typed
   variable substitution.

## Installation

```go
import "digital.vasic.llmops/llmops"
```

Add to your `go.mod` (HelixAgent uses a `replace` directive for local development):

```go
require digital.vasic.llmops v0.0.0

replace digital.vasic.llmops => ./LLMOps
```

## Key Types and Interfaces

### ContinuousEvaluator

Manages the lifecycle of evaluation runs.

```go
type ContinuousEvaluator interface {
    CreateRun(ctx context.Context, run *EvaluationRun) error
    StartRun(ctx context.Context, runID string) error
    GetRun(ctx context.Context, runID string) (*EvaluationRun, error)
    ListRuns(ctx context.Context, filter *EvaluationFilter) ([]*EvaluationRun, error)
    ScheduleRun(ctx context.Context, run *EvaluationRun, schedule string) error
    CompareRuns(ctx context.Context, runID1, runID2 string) (*RunComparison, error)
}
```

The concrete in-memory implementation is `InMemoryContinuousEvaluator` (used in tests and via
the adapter).

### EvaluationRun

Describes a single evaluation against a named dataset.

```go
type EvaluationRun struct {
    ID            string
    Name          string
    Dataset       string           // Dataset identifier
    PromptName    string
    PromptVersion string
    ModelName     string
    Metrics       []string
    Status        EvaluationStatus // pending | running | completed | failed
    Results       *EvaluationResults
    StartTime     *time.Time
    EndTime       *time.Time
    CreatedAt     time.Time
}
```

### ExperimentManager

Manages A/B experiments end-to-end.

```go
type ExperimentManager interface {
    Create(ctx context.Context, exp *Experiment) error
    Get(ctx context.Context, id string) (*Experiment, error)
    List(ctx context.Context, status ExperimentStatus) ([]*Experiment, error)
    Start(ctx context.Context, id string) error
    Pause(ctx context.Context, id string) error
    Complete(ctx context.Context, id string, winner string) error
    Cancel(ctx context.Context, id string) error
    AssignVariant(ctx context.Context, experimentID, userID string) (*Variant, error)
    RecordMetric(ctx context.Context, experimentID, variantID, metric string, value float64) error
    GetResults(ctx context.Context, experimentID string) (*ExperimentResult, error)
}
```

### Experiment / Variant

```go
type Experiment struct {
    ID           string
    Name         string
    Variants     []*Variant
    TrafficSplit map[string]float64  // variantID -> percentage (0-100)
    Status       ExperimentStatus    // draft | running | paused | completed | cancelled
    Metrics      []string
    TargetMetric string
    Winner       string
    Significance float64
    Confidence   float64
}

type Variant struct {
    ID            string
    Name          string
    PromptName    string
    PromptVersion string
    ModelName     string
    Parameters    map[string]interface{}  // temperature, max_tokens, etc.
    IsControl     bool
}
```

### Dataset / DatasetSample

```go
type Dataset struct {
    ID          string
    Name        string
    Type        DatasetType   // golden | regression | benchmark | user
    SampleCount int
    CreatedAt   time.Time
}

type DatasetSample struct {
    ID             string
    Input          string
    ExpectedOutput string
    Context        string
    Metadata       map[string]interface{}
}
```

### PromptRegistry

Manages versioned prompt templates.

```go
type PromptRegistry interface {
    Create(ctx context.Context, prompt *PromptVersion) error
    Get(ctx context.Context, name, version string) (*PromptVersion, error)
    GetLatest(ctx context.Context, name string) (*PromptVersion, error)
    Activate(ctx context.Context, name, version string) error
    Render(ctx context.Context, name, version string, vars map[string]interface{}) (string, error)
}
```

## Usage Examples

### Continuous Evaluation Run

```go
package main

import (
    "context"
    "fmt"

    "digital.vasic.llmops/llmops"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    evaluator := llmops.NewInMemoryContinuousEvaluator(nil, nil, nil, logger)
    ctx := context.Background()

    // Create a golden dataset
    ds := &llmops.Dataset{
        Name: "my-golden-set",
        Type: llmops.DatasetTypeGolden,
    }
    _ = evaluator.CreateDataset(ctx, ds)

    // Add samples
    samples := []*llmops.DatasetSample{
        {Input: "What is 2+2?", ExpectedOutput: "4"},
        {Input: "Capital of France?", ExpectedOutput: "Paris"},
    }
    _ = evaluator.AddSamples(ctx, ds.ID, samples)

    // Create and start an evaluation run
    run := &llmops.EvaluationRun{
        Name:          "baseline-eval",
        Dataset:       ds.ID,
        PromptName:    "qa-prompt",
        PromptVersion: "1.0.0",
        ModelName:     "gpt-4",
        Metrics:       []string{"exact_match", "latency"},
    }
    _ = evaluator.CreateRun(ctx, run)
    _ = evaluator.StartRun(ctx, run.ID)

    result, _ := evaluator.GetRun(ctx, run.ID)
    fmt.Printf("Pass rate: %.2f%%\n", result.Results.PassRate*100)
}
```

### A/B Experiment

```go
mgr := llmops.NewInMemoryExperimentManager(logger)
ctx := context.Background()

exp := &llmops.Experiment{
    Name: "prompt-a-vs-b",
    Variants: []*llmops.Variant{
        {ID: "control", Name: "Variant A", PromptName: "prompt-v1", IsControl: true},
        {ID: "treatment", Name: "Variant B", PromptName: "prompt-v2"},
    },
    TrafficSplit:  map[string]float64{"control": 50, "treatment": 50},
    Metrics:       []string{"quality", "latency"},
    TargetMetric:  "quality",
}
_ = mgr.Create(ctx, exp)
_ = mgr.Start(ctx, exp.ID)

// Assign a variant for a user request
variant, _ := mgr.AssignVariant(ctx, exp.ID, "user-123")
fmt.Println("Assigned variant:", variant.Name)

// Record quality metric after response
_ = mgr.RecordMetric(ctx, exp.ID, variant.ID, "quality", 0.87)

// Get statistical results
results, _ := mgr.GetResults(ctx, exp.ID)
fmt.Printf("Winner: %s (significance: %.2f)\n", results.Winner, results.Significance)
```

### Run Comparison (Regression Detection)

```go
comparison, err := evaluator.CompareRuns(ctx, "run-v1-id", "run-v2-id")
if err != nil {
    panic(err)
}
fmt.Printf("Pass rate change: %+.1f%%\n", comparison.PassRateChange*100)
for _, reg := range comparison.Regressions {
    fmt.Println("Regression:", reg)
}
```

## Integration with HelixAgent Adapter

HelixAgent wraps the module through `internal/adapters/llmops/adapter.go`.

```go
import llmopsadapter "dev.helix.agent/internal/adapters/llmops"

adapter := llmopsadapter.New(logger)

// Create an evaluator (dependencies injected from HelixAgent)
evaluator := adapter.NewEvaluator()

// Create an experiment manager
mgr := adapter.NewExperimentManager()

// Convenience: create a dataset in one call
ds, err := adapter.CreateDataset(ctx, evaluator, "regression-suite", llmopsmod.DatasetTypeRegression)
```

## Build and Test

```bash
cd LLMOps
go build ./...
go test ./... -count=1 -race
```
