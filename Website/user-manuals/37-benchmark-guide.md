# User Manual 37: LLM Benchmarking Guide

## Overview

HelixAgent's benchmarking system evaluates LLM providers against standard benchmarks (SWE-bench, HumanEval, MMLU) and custom test suites. Compare providers objectively with reproducible results.

## API Endpoints

### Start a Benchmark

```
POST /v1/benchmark/run
```

```json
{
  "benchmark_type": "humaneval",
  "name": "weekly-provider-comparison",
  "provider_name": "deepseek",
  "config": {
    "max_samples": 100,
    "timeout_per_sample": 30
  }
}
```

**Benchmark Types:**
- `humaneval` — Code generation (164 Python problems)
- `swe_bench` — Software engineering tasks (real GitHub issues)
- `mmlu` — Massive Multitask Language Understanding (57 subjects)
- `custom` — User-defined benchmark suite

### List Results

```
GET /v1/benchmark/results
GET /v1/benchmark/results?benchmark_type=humaneval
GET /v1/benchmark/results?provider_name=deepseek
GET /v1/benchmark/results?status=completed
```

### Get Specific Result

```
GET /v1/benchmark/results/:id
```

**Response:**

```json
{
  "id": "bench-abc123",
  "benchmark_type": "humaneval",
  "provider_name": "deepseek",
  "status": "completed",
  "score": 87.5,
  "total_samples": 164,
  "passed_samples": 143,
  "failed_samples": 21,
  "avg_latency_ms": 1250,
  "total_tokens": 45000,
  "started_at": "2026-03-23T10:00:00Z",
  "completed_at": "2026-03-23T10:15:30Z"
}
```

## Interpreting Results

| Metric | Description | Good Value |
|--------|-------------|------------|
| score | Pass rate (0-100) | > 80% |
| avg_latency_ms | Mean response time | < 2000ms |
| total_tokens | Token consumption | Varies |
| pass_rate | passed/total ratio | > 0.8 |

## Best Practices

1. Run benchmarks during off-peak hours to avoid rate limits
2. Compare at least 3 providers for meaningful comparisons
3. Use the same benchmark version across runs for consistency
4. Track results over time to detect provider quality regressions
5. Set `timeout_per_sample` based on task complexity
