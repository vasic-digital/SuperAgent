# Benchmark Module User Guide

**Module**: `digital.vasic.benchmark`
**Directory**: `Benchmark/`
**Phase**: 5 (AI/ML)

## Overview

The Benchmark module provides a framework for evaluating LLM providers against industry-standard
and custom benchmarks. It covers:

- **Built-in benchmarks**: SWE-Bench Lite, HumanEval, MMLU Mini, GSM8K, MATH, MBPP, LMSYS,
  HellaSwag, and custom benchmark support.
- **Run management**: Create, start, monitor, and cancel benchmark runs. Each run targets a
  specific provider/model combination.
- **Provider comparison**: Run the same benchmark against multiple providers and collect
  comparative results.
- **Leaderboard generation**: Aggregate completed runs across providers into a scored leaderboard.
- **AI debate evaluation**: Optionally use HelixAgent's AI debate system as the evaluator for
  subjective tasks where pass/fail is not clearly defined.

## Installation

```go
import "digital.vasic.benchmark/benchmark"
```

Add to your `go.mod` (HelixAgent uses a `replace` directive for local development):

```go
require digital.vasic.benchmark v0.0.0

replace digital.vasic.benchmark => ./Benchmark
```

## Key Types and Interfaces

### BenchmarkRunner

The primary interface for interacting with benchmarks.

```go
type BenchmarkRunner interface {
    ListBenchmarks(ctx context.Context) ([]*Benchmark, error)
    GetBenchmark(ctx context.Context, id string) (*Benchmark, error)
    GetTasks(ctx context.Context, benchmarkID string, config *BenchmarkConfig) ([]*BenchmarkTask, error)
    CreateRun(ctx context.Context, run *BenchmarkRun) error
    StartRun(ctx context.Context, runID string) error
    GetRun(ctx context.Context, runID string) (*BenchmarkRun, error)
    ListRuns(ctx context.Context, filter *RunFilter) ([]*BenchmarkRun, error)
    CancelRun(ctx context.Context, runID string) error
    CompareRuns(ctx context.Context, runID1, runID2 string) (*RunComparison, error)
}
```

Concrete implementation: `StandardBenchmarkRunner`.

### BenchmarkSystem

Higher-level orchestration layer that adds provider selection via a `VerifierService` and
leaderboard generation.

```go
type BenchmarkSystem struct { /* ... */ }

func NewBenchmarkSystem(cfg *BenchmarkSystemConfig, logger *logrus.Logger) *BenchmarkSystem

func (s *BenchmarkSystem) Initialize(provider LLMProvider) error
func (s *BenchmarkSystem) SetDebateService(svc DebateServiceForBenchmark)
func (s *BenchmarkSystem) SetVerifierService(svc VerifierServiceForBenchmark)
func (s *BenchmarkSystem) RunBenchmarkWithBestProvider(ctx context.Context, benchmarkType BenchmarkType, config *BenchmarkConfig) (*BenchmarkRun, error)
func (s *BenchmarkSystem) CompareProviders(ctx context.Context, benchmarkType BenchmarkType, providers []string, config *BenchmarkConfig) ([]*BenchmarkRun, error)
func (s *BenchmarkSystem) GenerateLeaderboard(ctx context.Context, benchmarkType BenchmarkType) (*Leaderboard, error)
func (s *BenchmarkSystem) GetRunner() BenchmarkRunner
```

### BenchmarkType

Supported benchmark identifiers:

```go
const (
    BenchmarkTypeSWEBench  BenchmarkType = "swe-bench"
    BenchmarkTypeHumanEval BenchmarkType = "humaneval"
    BenchmarkTypeMBPP      BenchmarkType = "mbpp"
    BenchmarkTypeLMSYS     BenchmarkType = "lmsys"
    BenchmarkTypeHellaSwag BenchmarkType = "hellaswag"
    BenchmarkTypeMMLU      BenchmarkType = "mmlu"
    BenchmarkTypeGSM8K     BenchmarkType = "gsm8k"
    BenchmarkTypeMATH      BenchmarkType = "math"
    BenchmarkTypeCustom    BenchmarkType = "custom"
)
```

### BenchmarkConfig

```go
type BenchmarkConfig struct {
    MaxTasks         int               // 0 = no limit
    Timeout          time.Duration     // per-task timeout
    Concurrency      int               // parallel task execution
    Retries          int
    Temperature      float64
    MaxTokens        int
    SystemPrompt     string
    Difficulties     []DifficultyLevel // filter by easy | medium | hard
    Tags             []string
    SaveResponses    bool
    UseDebateForEval bool              // use AI debate for evaluation
}

func DefaultBenchmarkConfig() *BenchmarkConfig
```

### BenchmarkRun

Tracks a complete benchmark execution.

```go
type BenchmarkRun struct {
    ID            string
    Name          string
    BenchmarkType BenchmarkType
    ProviderName  string
    ModelName     string
    Status        BenchmarkStatus  // pending | running | completed | failed | cancelled
    Config        *BenchmarkConfig
    Results       []*BenchmarkResult
    Summary       *BenchmarkSummary
    StartTime     *time.Time
    EndTime       *time.Time
    CreatedAt     time.Time
}
```

### BenchmarkSummary

```go
type BenchmarkSummary struct {
    TotalTasks     int
    PassedTasks    int
    FailedTasks    int
    ErrorTasks     int
    PassRate       float64
    AverageScore   float64
    AverageLatency time.Duration
    TotalTokens    int
    ByDifficulty   map[DifficultyLevel]*DifficultySummary
    ByTag          map[string]*TagSummary
}
```

### LLMProvider (benchmark interface)

The benchmark module uses its own `LLMProvider` interface:

```go
type LLMProvider interface {
    Complete(ctx context.Context, prompt, systemPrompt string) (string, int, error) // response, tokens, error
    GetName() string
}
```

## Usage Examples

### List Available Benchmarks

```go
package main

import (
    "context"
    "fmt"

    "digital.vasic.benchmark/benchmark"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    runner := benchmark.NewStandardBenchmarkRunner(myLLMProvider, logger)

    benchmarks, _ := runner.ListBenchmarks(context.Background())
    for _, b := range benchmarks {
        fmt.Printf("%s (%s): %d tasks\n", b.Name, b.Type, b.TaskCount)
    }
}
```

### Run a Benchmark

```go
runner := benchmark.NewStandardBenchmarkRunner(myLLMProvider, logger)
ctx := context.Background()

// Get HumanEval benchmark
benchmarks, _ := runner.ListBenchmarks(ctx)
var humanEvalID string
for _, b := range benchmarks {
    if b.Type == benchmark.BenchmarkTypeHumanEval {
        humanEvalID = b.ID
        break
    }
}

cfg := benchmark.DefaultBenchmarkConfig()
cfg.MaxTasks = 20  // run a subset
cfg.Concurrency = 2

run := &benchmark.BenchmarkRun{
    Name:          "humaneval-gpt4-run1",
    BenchmarkType: benchmark.BenchmarkTypeHumanEval,
    ProviderName:  "openai",
    ModelName:     "gpt-4",
    Config:        cfg,
}
_ = runner.CreateRun(ctx, run)
_ = runner.StartRun(ctx, run.ID)

completed, _ := runner.GetRun(ctx, run.ID)
fmt.Printf("Pass rate: %.1f%% (%d/%d tasks)\n",
    completed.Summary.PassRate*100,
    completed.Summary.PassedTasks,
    completed.Summary.TotalTasks,
)
```

### Compare Two Runs

```go
comparison, _ := runner.CompareRuns(ctx, "run-gpt4-id", "run-claude-id")
fmt.Printf("Pass rate change: %+.1f%%\n", comparison.PassRateChange*100)
fmt.Printf("Score change: %+.2f\n", comparison.ScoreChange)
for _, reg := range comparison.Regressions {
    fmt.Println("Regressed task:", reg)
}
```

### Using BenchmarkSystem for Provider Comparison

```go
sysCfg := benchmark.DefaultBenchmarkSystemConfig()
system := benchmark.NewBenchmarkSystem(sysCfg, logger)
_ = system.Initialize(myProvider)
system.SetVerifierService(myVerifierService)

// Run with automatically selected best provider
run, _ := system.RunBenchmarkWithBestProvider(ctx, benchmark.BenchmarkTypeMMLU, nil)
fmt.Printf("Best provider run: %s, pass rate: %.1f%%\n",
    run.ProviderName, run.Summary.PassRate*100)

// Compare multiple providers
runs, _ := system.CompareProviders(ctx, benchmark.BenchmarkTypeGSM8K, []string{"openai", "anthropic", "google"}, nil)
for _, r := range runs {
    fmt.Printf("%s/%s: %.1f%%\n", r.ProviderName, r.ModelName, r.Summary.PassRate*100)
}

// Generate leaderboard
leaderboard, _ := system.GenerateLeaderboard(ctx, benchmark.BenchmarkTypeHumanEval)
for i, entry := range leaderboard.Entries {
    fmt.Printf("#%d %s/%s: %.1f%%\n", i+1, entry.ProviderName, entry.ModelName, entry.PassRate*100)
}
```

## Integration with HelixAgent Adapter

HelixAgent wraps the module through `internal/adapters/benchmark/adapter.go`.

```go
import benchmarkadapter "dev.helix.agent/internal/adapters/benchmark"
import extbenchmark "digital.vasic.benchmark/benchmark"

adapter := benchmarkadapter.New(logger)

// Wire HelixAgent provider and services
_ = adapter.Initialize(myProviderService, "openai", "gpt-4")
adapter.SetDebateService(myDebateService)
adapter.SetVerifierService(myVerifierService)

// List benchmarks
benchmarks, _ := adapter.ListBenchmarks(ctx)

// Run benchmark with best provider (selected via verifier scores)
run, _ := adapter.RunBenchmarkWithBestProvider(ctx, extbenchmark.BenchmarkTypeHumanEval, nil)

// Compare providers
runs, _ := adapter.CompareProviders(ctx, extbenchmark.BenchmarkTypeMMLU,
    []string{"claude", "gemini", "openai"}, nil)

// Generate leaderboard
leaderboard, _ := adapter.GenerateLeaderboard(ctx, extbenchmark.BenchmarkTypeHumanEval)

// Get the underlying runner for low-level access
runner := adapter.GetRunner()
```

The adapter creates and manages a `BenchmarkSystem` internally, injecting the HelixAgent
logger and wiring optional debate/verifier services for enhanced evaluation.

## Build and Test

```bash
cd Benchmark
go build ./...
go test ./... -count=1 -race
```
