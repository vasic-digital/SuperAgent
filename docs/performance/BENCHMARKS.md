# HelixAgent Benchmark Suite

This document catalogs the benchmark tests defined in `tests/performance/` and how to run them.

## Running Benchmarks

All benchmark tests require the `performance` build tag:

```bash
# Run all benchmarks (resource-limited per CLAUDE.md mandate)
GOMAXPROCS=2 nice -n 19 go test -tags performance -bench=. -benchmem \
    -benchtime=3s ./tests/performance/ 2>&1 | tee /tmp/bench-results.txt

# Run a specific benchmark
GOMAXPROCS=2 go test -tags performance -bench=BenchmarkCache_Get \
    -benchmem ./tests/performance/

# Run messaging benchmarks
GOMAXPROCS=2 go test -bench=. -benchmem ./tests/performance/messaging/
```

## Resource Limits

Per the project constitution, ALL benchmark execution MUST be limited to 30-40% of host
resources:

- `GOMAXPROCS=2` — cap Go parallelism
- `nice -n 19` — low OS scheduling priority
- `ionice -c 3` — idle I/O class

## Benchmark Index

### Core / Cache (`benchmark_test.go`)

| Benchmark | Description |
|-----------|-------------|
| `BenchmarkCache_Get` | In-memory cache read latency |
| `BenchmarkCache_Set` | In-memory cache write latency |
| `BenchmarkCache_GetSet_Mixed` | Mixed read/write under load |
| `BenchmarkEventBus_Publish` | EventBus publish throughput |
| `BenchmarkEventBus_PubSub` | EventBus pub/sub round-trip |
| `BenchmarkWorkerPool_Submit` | Worker pool task submission |
| `BenchmarkWorkerPool_SubmitAndWait` | Worker pool submit-and-wait |
| `BenchmarkHTTP_HealthCheck` | HTTP health endpoint latency |
| `BenchmarkHTTP_ChatCompletion` | Chat completion HTTP latency |
| `BenchmarkAllocation_JSONMarshal` | JSON marshal allocation cost |
| `BenchmarkAllocation_EventCreation` | Event struct allocation cost |

### Ensemble (`ensemble_benchmark_test.go`)

| Benchmark | Description |
|-----------|-------------|
| `BenchmarkEnsemble_Semaphore_5Providers` | Semaphore-limited 5-provider ensemble |
| `BenchmarkEnsemble_Semaphore_20Providers` | Semaphore-limited 20-provider ensemble |
| `BenchmarkEnsemble_WithLatency` | Ensemble with simulated provider latency |

### Lazy Loading (`lazy_loading_benchmark_test.go`)

| Benchmark | Description |
|-----------|-------------|
| `BenchmarkDirect_Access` | Direct (non-lazy) provider access |
| `BenchmarkLazy_*` | Lazy-initialized provider access patterns |

### Semaphore (`semaphore_benchmark_test.go`)

| Benchmark | Description |
|-----------|-------------|
| `BenchmarkChannel_Semaphore` | Channel-based semaphore throughput |
| `BenchmarkChannel_HighContention` | Semaphore under high contention |

### Debate (`debate_benchmark_test.go`)

| Benchmark | Description |
|-----------|-------------|
| `BenchmarkAgentRoleAssignment` | Debate agent role assignment cost |
| `BenchmarkBordaVoting` | Borda count voting performance |
| `BenchmarkCondorcetVoting` | Condorcet voting with cycle detection |
| `BenchmarkAccumulatedWisdomStore` | Reflexion wisdom store throughput |
| `BenchmarkAccumulatedWisdomRelevance` | Wisdom relevance scoring |
| `BenchmarkEpisodicMemoryRelevance` | Episodic memory lookup latency |

### Core Internals (`skills_benchmark_test.go`)

| Benchmark | Description |
|-----------|-------------|
| `BenchmarkCLIAgentRegistry_Lookup` | Agent registry lookup latency |
| `BenchmarkCLIAgentRegistry_Enumerate` | Agent registry enumeration |
| `BenchmarkCore_CircuitBreaker_IsOpen` | Circuit breaker state check |
| `BenchmarkCore_CircuitBreaker_GetState` | Circuit breaker full state read |
| `BenchmarkCore_ProviderRegistry_GetProvider` | Provider registry lookup |
| `BenchmarkCore_ProviderRegistry_ListProviders` | Provider registry list |
| `BenchmarkCore_ToolSchema_GetToolSchema` | Tool schema retrieval |
| `BenchmarkCore_ToolSchema_GetToolsByCategory` | Tool schema category filter |
| `BenchmarkCore_ToolSchema_ValidateToolArgs` | Tool argument validation |
| `BenchmarkCore_MCPAdapterRegistry_Get` | MCP adapter registry lookup |
| `BenchmarkCore_MCPAdapterRegistry_GetMetadata` | MCP adapter metadata |
| `BenchmarkCore_DebateOptimizer_ShouldTerminateEarly` | Debate early-termination check |

### Messaging (`messaging/benchmark_test.go`)

| Benchmark | Description |
|-----------|-------------|
| `BenchmarkInMemoryPublish` | In-memory broker publish |
| `BenchmarkInMemoryBatchPublish` | Batch publish throughput |
| `BenchmarkInMemoryPubSub` | Pub/sub round-trip |
| `BenchmarkMessageSerialization` | Message marshal/unmarshal cost |
| `BenchmarkConcurrentPublishers` | N concurrent publishers |
| `BenchmarkMessageThroughput` | Sustained message throughput |
| `BenchmarkLatency` | End-to-end message latency |
| `BenchmarkLargePayload` | Large message handling |
| `BenchmarkTopicRouting` | Topic routing overhead |

## Interpreting Results

Benchmark output format: `BenchmarkName-N   iterations   ns/op   B/op   allocs/op`

- **ns/op** — nanoseconds per operation (lower is better)
- **B/op** — bytes allocated per operation (lower is better)
- **allocs/op** — heap allocations per operation (lower is better)

## Regression Baseline

Store a baseline with:

```bash
GOMAXPROCS=2 go test -tags performance -bench=. -benchmem -count=5 \
    ./tests/performance/ | tee docs/performance/baseline.txt
```

Compare against baseline with `benchstat`:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
benchstat docs/performance/baseline.txt /tmp/bench-results.txt
```
