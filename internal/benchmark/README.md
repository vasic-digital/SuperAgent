# Package: benchmark

## Overview

The `benchmark` package provides comprehensive LLM benchmark runners for evaluating model performance across industry-standard benchmarks including SWE-Bench, HumanEval, MBPP, MMLU, GSM8K, and more.

## Architecture

```
benchmark/
├── types.go          # Benchmark types and models
├── runner.go         # Benchmark execution engine
├── integration.go    # LLM provider integration
└── benchmark_test.go # Unit tests (73.3% coverage)
```

## Supported Benchmarks

| Benchmark | Type | Description |
|-----------|------|-------------|
| SWE-Bench | Code | Software engineering tasks |
| HumanEval | Code | Python code generation |
| MBPP | Code | Mostly basic Python problems |
| LMSYS | Chat | Chatbot arena evaluation |
| HellaSwag | Reasoning | Common sense reasoning |
| MMLU | Knowledge | Massive multitask language understanding |
| GSM8K | Math | Grade school math word problems |
| MATH | Math | Competition math problems |

## Key Types

### BenchmarkTask

```go
type BenchmarkTask struct {
    ID           string
    BenchmarkID  string
    Type         BenchmarkType
    Name         string
    Prompt       string
    Expected     string
    TestCases    []*TestCase
    Difficulty   DifficultyLevel
    TimeLimit    time.Duration
}
```

### BenchmarkResult

```go
type BenchmarkResult struct {
    TaskID       string
    ProviderName string
    ModelName    string
    Response     string
    Passed       bool
    Score        float64  // 0.0 to 1.0
    Latency      time.Duration
    TokensUsed   int
}
```

## Usage

### Running Benchmarks

```go
import "dev.helix.agent/internal/benchmark"

// Create runner
runner := benchmark.NewRunner(provider, config)

// Run HumanEval benchmark
results, err := runner.RunBenchmark(ctx, benchmark.BenchmarkTypeHumanEval)

// Get summary
summary := results.Summary()
fmt.Printf("Pass@1: %.2f%%\n", summary.PassAt1 * 100)
```

### Custom Benchmarks

```go
tasks := []*benchmark.BenchmarkTask{
    {
        ID:     "custom-1",
        Type:   benchmark.BenchmarkTypeCustom,
        Prompt: "Write a function to sort an array",
        TestCases: []*benchmark.TestCase{
            {Input: "[3,1,2]", Expected: "[1,2,3]"},
        },
    },
}

results, err := runner.RunTasks(ctx, tasks)
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| Concurrency | int | 4 | Parallel task execution |
| Timeout | time.Duration | 60s | Per-task timeout |
| RetryCount | int | 3 | Retries on failure |
| SaveResults | bool | true | Persist results |

## Testing

```bash
go test -v ./internal/benchmark/...
go test -cover ./internal/benchmark/...
```

## Dependencies

### Internal
- `internal/llm` - LLM providers
- `internal/database` - Result storage

### External
- Standard library only

## See Also

- [SWE-Bench](https://www.swebench.com/)
- [HumanEval Paper](https://arxiv.org/abs/2107.03374)
