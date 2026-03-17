# debate/evaluation - Benchmark Bridge

Connects the debate system with standardized benchmarks (SWE-bench, HumanEval, MMLU) for evaluating debate quality through static code analysis and benchmark task scoring.

## Purpose

The evaluation package bridges the debate orchestrator with the Benchmark module, enabling automated quality assessment of debate outcomes against standardized benchmarks. It performs static code analysis across 5 metrics to score debate-generated code.

## Key Types

### BenchmarkBridge

Evaluates debate outputs against benchmark tasks.

```go
bridge := evaluation.NewBenchmarkBridge(benchmarkRunner, logger)
score, passed, err := bridge.EvaluateResponse(ctx, task, debateConsensus)
```

### Evaluation Metrics

The bridge evaluates debate outputs across 5 metrics:

| Metric | Description |
|--------|-------------|
| Correctness | Does the output match expected behavior? |
| Completeness | Does it address all aspects of the task? |
| Code Quality | Static analysis for style, structure, and patterns |
| Safety | Checks for unsafe patterns and vulnerabilities |
| Efficiency | Evaluates algorithmic complexity and resource usage |

## Usage within Debate System

The benchmark bridge is invoked by the debate CI/CD hooks to validate that debate-generated code meets quality thresholds before acceptance. It can also be used in the debate evaluation phase to score proposals.

## Files

- `benchmark_bridge.go` -- BenchmarkBridge, evaluation metrics, scoring logic
- `benchmark_bridge_test.go` -- Unit tests
