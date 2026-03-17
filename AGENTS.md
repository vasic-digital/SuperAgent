# AGENTS.md

Guidance for AI coding agents working in the HelixAgent repository.

## Project Overview

HelixAgent is an AI-powered ensemble LLM service in Go (1.24+) that aggregates responses from multiple language models. It provides OpenAI-compatible APIs with 22+ LLM providers, debate orchestration, MCP adapters, and containerized infrastructure.

## Build Commands

```bash
make build              # Build binary (output in bin/)
make build-debug        # Build with debug symbols
make build-all          # Build for all platforms (Linux, macOS, Windows)
make run                # Run locally
make run-dev            # Development mode (GIN_MODE=debug)
make docker-build       # Build Docker image
make docker-run         # Start services with Docker Compose
```

## Linting & Formatting

```bash
make fmt                # Format with go fmt
make vet                # Run go vet static analysis
make lint               # Run golangci-lint (install: make install-deps)
make security-scan      # Run gosec security scanner
```

**Always run `make fmt vet lint` before committing.**

## Testing Commands

### Running Tests

```bash
make test               # All tests (verbose)
make test-unit          # Unit tests only (./internal/... -short)
make test-integration   # Integration tests with Docker deps
make test-e2e           # End-to-end tests
make test-coverage      # Tests with HTML coverage report
make test-bench         # Benchmark tests
make test-race          # Tests with race detection
```

### Running a Single Test

```bash
# Run specific test function
go test -v -run TestFunctionName ./path/to/package

# Run all tests in a package
go test -v ./internal/llm

# Run tests matching pattern
go test -v -run "Test.*Integration" ./...

# With coverage for single package
go test -v -coverprofile=coverage.out ./internal/llm
go tool cover -html=coverage.out

# Resource-limited test (CRITICAL for host stability)
GOMAXPROCS=2 go test -v -p 1 -run TestName ./path/to/package
```

### Test Infrastructure

**IMPORTANT: Container infrastructure is handled AUTOMATICALLY by HelixAgent during boot.**

The HelixAgent binary (`./bin/helixagent`) orchestrates ALL containers based on `Containers/.env`. There is NO need to manually start test infrastructure.

If you need containers for tests:
1. Run `./bin/helixagent` first - it will deploy containers to local or remote hosts
2. Then run your tests against the running containers

**Legacy commands (may not work without manual container setup):**
```bash
make test-infra-start   # Start PostgreSQL, Redis, Mock LLM containers (DEPRECATED)
make test-infra-stop    # Stop test infrastructure
make test-with-infra    # Run tests with Docker infrastructure
```

**Prefer:** Run `./bin/helixagent` and then execute tests against the running service.

## Code Style Guidelines

### Formatting & Imports
- Use `gofmt` / `goimports` for formatting
- Imports grouped: standard library, third-party, internal (blank line separated)
- Line length: ≤ 100 characters (readability first)

### Naming Conventions
- `camelCase`: local variables, private functions
- `PascalCase`: exported functions, types, constants, fields
- `UPPER_SNAKE_CASE`: exported constants
- Acronyms all caps: `HTTP`, `URL`, `ID`, `JSON`
- Receiver names: 1-2 letters (`s` for service, `c` for client)

### Error Handling
```go
// Always check errors
if err != nil {
    return err
}

// Wrap with context
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Use defer for cleanup
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()
```

### Types & Interfaces
- Use `interface` to define behavior, not data
- Prefer small, focused interfaces (`io.Reader`, `io.Writer`)
- Use struct tags for JSON, YAML, database mapping
- Avoid `any`/`interface{}`; use generics or specific types

### Concurrency
- Always use `context.Context` for cancellation/timeout
- Protect shared data with `sync.Mutex` or `sync.RWMutex`
- Use `sync.WaitGroup` for goroutine coordination

### Testing Patterns
- Write table-driven tests
- Use `testify` assertion library
- Place test files in same package with `_test.go` suffix
- Use `testdata/` directories for fixtures
- Mocks/stubs ONLY in unit tests; integration tests use real services

## Key Conventions

### Tool Schema
All tool parameters use **snake_case** (e.g., `file_path`, `old_string`). See `internal/tools/schema.go`.

### No Comments
**DO NOT ADD COMMENTS** in code unless explicitly requested.

### Git Operations
- **SSH ONLY** for all Git operations - HTTPS is forbidden
- Branch naming: `feat/`, `fix/`, `chore/`, `docs/`, `refactor/`, `test/` + description
- Commits: Conventional Commits (`feat(scope): description`)
- Run `make fmt vet lint` before committing

### Containerization

**CRITICAL: ALL container orchestration is handled AUTOMATICALLY by the HelixAgent binary during its boot process.**

- **DO NOT** manually start, stop, or manage containers via `docker` or `podman` commands
- **DO NOT** run `make test-infra-start` or similar manual container orchestration commands
- **DO NOT** attempt to deploy containers to remote hosts manually via SSH

**The ONLY correct workflow is:**
1. Build the binary: `make build`
2. Run the binary: `./bin/helixagent` (it handles ALL container orchestration automatically)
3. The binary reads `Containers/.env` and orchestrates containers locally OR remotely based on configuration

**Container Orchestration Flow (handled by HelixAgent):**
1. HelixAgent boots and initializes Containers module adapter
2. Adapter reads `Containers/.env` file (NOT project root `.env`)
3. Based on `Containers/.env`:
   - `CONTAINERS_REMOTE_ENABLED=true` → ALL containers to remote host(s) via `CONTAINERS_REMOTE_HOST_*` vars
   - `CONTAINERS_REMOTE_ENABLED=false` → ALL containers locally
4. Health checks performed against configured endpoints
5. Required services failing health check cause boot failure in strict mode

**Key Files:**
- `Containers/.env` - Container orchestration configuration
- `internal/services/boot_manager.go` - Boot orchestration logic
- `tests/precondition/containers_boot_test.go` - Precondition tests for container state

**If you need to run tests that require containers:**
- Simply run `./bin/helixagent` first - it will start all required containers
- Or run tests that use the HelixAgent binary's built-in container management

**Rebuild containers only after code changes:**
```bash
make docker-build && make docker-run  # Only if you changed containerized code
```

## Resource Limits (CRITICAL)

**ALL test execution MUST be limited to 30-40% of host resources:**

```bash
# Pattern for resource-limited execution
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -v -p 1 ./...
```

Host runs mission-critical processes; exceeding limits has caused system crashes.

## Quick Reference

| Task | Command |
|------|---------|
| Build | `make build` |
| Run (starts containers automatically) | `./bin/helixagent` |
| Format | `make fmt` |
| Lint | `make lint` |
| All tests | `make test` |
| Single test | `go test -v -run TestName ./path/to/pkg` |
| Pre-commit | `make fmt vet lint` |

**NOTE:** Container orchestration is AUTOMATIC. Do NOT run manual container commands.

## Extracted Modules (27 Submodules)

HelixAgent's functionality is decomposed into **27 independent Go modules**, each with its own `go.mod`, tests, `CLAUDE.md`, `AGENTS.md`, `README.md`, and `docs/`. All modules are integrated as git submodules with `replace` directives in the root `go.mod` for local development.

### Building and Testing Any Module

```bash
# Build a module
cd <ModuleDir> && go build ./...

# Test a module (resource-limited, with race detection)
cd <ModuleDir> && GOMAXPROCS=2 go test ./... -count=1 -race

# Test all modules
for mod in EventBus Concurrency Observability Auth Storage Streaming \
           Security VectorDB Embeddings Database Cache \
           Messaging Formatters MCP_Module RAG Memory Optimization Plugins \
           Agentic LLMOps SelfImprove Planning Benchmark HelixMemory \
           HelixSpecifier Containers Challenges; do
  echo "Testing $mod..."
  (cd $mod && go test ./... -count=1 -race -short)
done
```

### Phase 1: Foundation (Zero Dependencies)

| Module | Go Module Path | Directory | Description |
|--------|---------------|-----------|-------------|
| EventBus | `digital.vasic.eventbus` | `EventBus/` | Pub/sub event system with synchronous/async dispatch, topic filtering, and middleware chain. 4 packages. |
| Concurrency | `digital.vasic.concurrency` | `Concurrency/` | Worker pools, priority queues, rate limiters (token bucket/sliding window), circuit breakers, semaphores, resource monitoring. 6 packages. |
| Observability | `digital.vasic.observability` | `Observability/` | OpenTelemetry tracing, Prometheus metrics, structured logging, health checks, ClickHouse analytics. 5 packages. |
| Auth | `digital.vasic.auth` | `Auth/` | JWT, API key, OAuth authentication; HTTP middleware; token management. 5 packages. |
| Storage | `digital.vasic.storage` | `Storage/` | Object storage abstraction: S3/MinIO, local filesystem, cloud providers. 4 packages. |
| Streaming | `digital.vasic.streaming` | `Streaming/` | SSE, WebSocket, gRPC streaming, webhooks, HTTP client, transport abstraction. 6 packages. |

```bash
# Example: build and test EventBus
cd EventBus && go build ./... && go test ./... -count=1 -race
```

### Phase 2: Infrastructure (Zero Module Dependencies, Complex)

| Module | Go Module Path | Directory | Description |
|--------|---------------|-----------|-------------|
| Security | `digital.vasic.security` | `Security/` | Guardrails engine, PII detection/redaction, content filtering, policy enforcement, vulnerability scanning. 5 packages. |
| VectorDB | `digital.vasic.vectordb` | `VectorDB/` | Unified vector store: Qdrant, Pinecone, Milvus, pgvector adapters; similarity search, collection management. 5 packages. |
| Embeddings | `digital.vasic.embeddings` | `Embeddings/` | 6 embedding providers (OpenAI, Cohere, Voyage, Jina, Google, Bedrock); batch embedding. 7 packages. |
| Database | `digital.vasic.database` | `Database/` | PostgreSQL (pgx), SQLite, connection pooling, migrations, repository pattern, query builder. 7 packages. |
| Cache | `digital.vasic.cache` | `Cache/` | Redis + in-memory caching, distributed cache, TTL policies (fixed, sliding, adaptive), cache warming. 5 packages. |

```bash
# Example: build and test Database
cd Database && go build ./... && go test ./... -count=1 -race
```

### Phase 3: Services

| Module | Go Module Path | Directory | Description |
|--------|---------------|-----------|-------------|
| Messaging | `digital.vasic.messaging` | `Messaging/` | Kafka + RabbitMQ: unified broker, producer/consumer, dead letter queues, retry policies. 5 packages. |
| Formatters | `digital.vasic.formatters` | `Formatters/` | Code formatter framework: native/service/built-in formatters, registry, executor, caching. 6 packages. |
| MCP | `digital.vasic.mcp` | `MCP_Module/` | Model Context Protocol: adapter framework, client/server, config generation, registry, JSON-RPC protocol. 6 packages. |

```bash
# Example: build and test Formatters
cd Formatters && go build ./... && go test ./... -count=1 -race
```

### Phase 4: Integration

| Module | Go Module Path | Directory | Description |
|--------|---------------|-----------|-------------|
| RAG | `digital.vasic.rag` | `RAG/` | Retrieval-Augmented Generation: chunking, retrieval, reranking, hybrid search, pipeline composition. 5 packages. |
| Memory | `digital.vasic.memory` | `Memory/` | Mem0-style memory: entity graph, semantic search, memory scopes, consolidation. 4 packages. |
| Optimization | `digital.vasic.optimization` | `Optimization/` | GPT-Cache, Outlines structured output, streaming optimization, SGLang, prompt optimization. 6 packages. |
| Plugins | `digital.vasic.plugins` | `Plugins/` | Plugin system: interface + lifecycle, registry, dynamic loading, sandboxing, structured output parsing. 5 packages. |

```bash
# Example: build and test RAG
cd RAG && go build ./... && go test ./... -count=1 -race
```

### Phase 5: AI/ML

| Module | Go Module Path | Directory | Description |
|--------|---------------|-----------|-------------|
| Agentic | `digital.vasic.agentic` | `Agentic/` | Graph-based agentic workflow orchestration: multi-step execution, conditional branching, state management. 1 package. |
| LLMOps | `digital.vasic.llmops` | `LLMOps/` | LLM operations: continuous evaluation, A/B experiment management, dataset management, prompt versioning. 1 package (5 files). |
| SelfImprove | `digital.vasic.selfimprove` | `SelfImprove/` | AI self-improvement: reward modelling, RLHF feedback integration, optimizer, dimension-weighted scoring. 1 package (5 files). |
| Planning | `digital.vasic.planning` | `Planning/` | AI planning algorithms: hierarchical planning (HiPlan), Monte Carlo Tree Search (MCTS), Tree of Thoughts. 1 package (3 files). |
| Benchmark | `digital.vasic.benchmark` | `Benchmark/` | LLM benchmarking: SWE-bench, HumanEval, MMLU and custom benchmarks; leaderboard, provider comparison. 1 package (3 files). |

```bash
# Example: build and test Planning
cd Planning && go build ./... && go test ./... -count=1 -race
```

### Phase 6: Cognitive

| Module | Go Module Path | Directory | Description |
|--------|---------------|-----------|-------------|
| HelixMemory | `digital.vasic.helixmemory` | `HelixMemory/` | Unified cognitive memory engine for HelixAgent and AI debate ensemble. Orchestrates Mem0, Cognee, Letta, Graphiti through 3-stage fusion pipeline. 12 power features, circuit breakers, Prometheus metrics. Active by default; opt out with `-tags nohelixmemory`. 12+ packages. |

```bash
# Example: build and test HelixMemory
cd HelixMemory && go build ./... && go test ./... -count=1 -race
```

### Phase 7: Specification

| Module | Go Module Path | Directory | Description |
|--------|---------------|-----------|-------------|
| HelixSpecifier | `digital.vasic.helixspecifier` | `HelixSpecifier/` | Spec-Driven Development Fusion Engine: 3-pillar architecture (SpecKit + Superpowers + GSD), adaptive ceremony scaling, effort classification, CLI agent adapters, 10 power features, spec memory, DebateFunc injection. Active by default; opt out with `-tags nohelixspecifier`. 27 packages. |

```bash
# Example: build and test HelixSpecifier
cd HelixSpecifier && go build ./... && go test ./... -count=1 -race
```

### Pre-existing Modules

| Module | Go Module Path | Directory | Description |
|--------|---------------|-----------|-------------|
| Containers | `digital.vasic.containers` | `Containers/` | Generic container orchestration: runtime abstraction (Docker/Podman/K8s), health checking, compose orchestration, lifecycle management. 12 packages. |
| Challenges | `digital.vasic.challenges` | `Challenges/` | Generic challenge framework: assertion engine (19 evaluators), registry, runner, reporting, monitoring, metrics, plugin system v2.0.0, userflow testing, Panoptic vision/recorder/testgen/error-analyzer adapters, AI test generation challenges. 15 packages. |

```bash
# Example: build and test Containers
cd Containers && go build ./... && go test ./... -count=1 -race
```

### Module Dependencies in go.mod

All modules use `replace` directives for local development:

```go
replace digital.vasic.eventbus => ./EventBus
replace digital.vasic.concurrency => ./Concurrency
replace digital.vasic.observability => ./Observability
replace digital.vasic.auth => ./Auth
replace digital.vasic.storage => ./Storage
replace digital.vasic.streaming => ./Streaming
replace digital.vasic.security => ./Security
replace digital.vasic.vectordb => ./VectorDB
replace digital.vasic.embeddings => ./Embeddings
replace digital.vasic.database => ./Database
replace digital.vasic.cache => ./Cache
replace digital.vasic.messaging => ./Messaging
replace digital.vasic.formatters => ./Formatters
replace digital.vasic.mcp => ./MCP_Module
replace digital.vasic.rag => ./RAG
replace digital.vasic.memory => ./Memory
replace digital.vasic.optimization => ./Optimization
replace digital.vasic.plugins => ./Plugins
replace digital.vasic.agentic => ./Agentic
replace digital.vasic.llmops => ./LLMOps
replace digital.vasic.selfimprove => ./SelfImprove
replace digital.vasic.planning => ./Planning
replace digital.vasic.benchmark => ./Benchmark
replace digital.vasic.helixmemory => ./HelixMemory
replace digital.vasic.helixspecifier => ./HelixSpecifier
replace digital.vasic.containers => ./Containers
replace digital.vasic.challenges => ./Challenges
```

## Key Files

- `CLAUDE.md` - Detailed project architecture
- `Makefile` - All available commands
- `go.mod` - Module dependencies
- `docs/MODULES.md` - Extracted modules catalog (27 modules)
- `.env.example` - Environment variable templates

## API Endpoints (New in WS1)

Five handler groups connected to the router:

| Endpoint Group | Base Path | Purpose |
|---------------|-----------|---------|
| Background Tasks | `/v1/tasks` | Async task lifecycle (create, pause, resume, cancel, WebSocket updates) |
| Model Discovery | `/v1/discovery` | 3-tier model discovery (Provider API, models.dev, hardcoded fallback) |
| Model Scoring | `/v1/scoring` | 5-component weighted scoring (speed, cost, efficiency, capability, recency) |
| Provider Verification | `/v1/verification` | 8-test verification pipeline for provider health |
| Provider Health | `/v1/health` | Real-time provider health, latency, circuit breaker states |

Protocol Endpoints: MCP `/v1/mcp` | ACP `/v1/acp` | LSP `/v1/lsp` | Embeddings `/v1/embeddings` | Vision `/v1/vision` | Cognee `/v1/cognee` | Startup `/v1/startup/verification` | BigData `/v1/bigdata/health` | Tasks `/v1/tasks` | Discovery `/v1/discovery` | Scoring `/v1/scoring` | Verification `/v1/verification` | Health `/v1/health`

## Additional Testing

### Fuzz Testing

Go native fuzz tests validate parsing robustness:

```bash
# Run fuzz corpus as regression tests
go test -run=Fuzz ./tests/fuzz/

# Run active fuzzing for 30 seconds
GOMAXPROCS=2 nice -n 19 ionice -c 3 \
  go test -fuzz=FuzzJSONRequestParsing -fuzztime=30s -p 1 ./tests/fuzz/
```

Available targets: FuzzJSONRequestParsing, FuzzToolSchemaValidation, FuzzSSEParsing, FuzzModelIDParsing, FuzzHTTPHeaderParsing.

## Additional Challenges (WS1-WS7)

```bash
./challenges/scripts/router_completeness_challenge.sh           # Router handler registration
./challenges/scripts/resource_limits_challenge.sh               # Test resource limit enforcement
./challenges/scripts/documentation_completeness_challenge.sh    # Documentation sync
./challenges/scripts/snyk_automated_scanning_challenge.sh       # 38 tests - Snyk scanning
./challenges/scripts/sonarqube_automated_scanning_challenge.sh  # 45 tests - SonarQube scanning
./challenges/scripts/pprof_memory_profiling_challenge.sh        # Memory leak detection
./challenges/scripts/coverage_gate_challenge.sh                 # Coverage thresholds
./challenges/scripts/lazy_loading_validation_challenge.sh       # sync.Once validation
./challenges/scripts/monitoring_dashboard_challenge.sh          # Prometheus metrics
```
