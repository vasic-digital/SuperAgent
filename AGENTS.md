# AGENTS.md

Guidance for AI coding agents working in the HelixAgent repository.

## Project Overview

HelixAgent is an AI-powered ensemble LLM service in Go (1.24+, toolchain go1.24.11) that aggregates responses from multiple language models. It provides OpenAI-compatible APIs with 22+ LLM providers, debate orchestration, MCP adapters, and containerized infrastructure.

**Module**: `dev.helix.agent`

**Subprojects**: Toolkit (`Toolkit/`) — Go library for AI apps. LLMsVerifier (`LLMsVerifier/`) — provider accuracy verification. **27 extracted modules** (see below).

## Mandatory Development Standards (NON-NEGOTIABLE)

1. **100% Test Coverage** — Unit, integration, E2E, security, stress, chaos, automation, benchmark tests. Mocks ONLY in unit tests.
2. **Challenge Coverage** — Every component MUST have Challenge scripts (`./challenges/scripts/`).
3. **Containerization** — All services in containers. Auto boot-up via HelixAgent binary.
4. **Centralized Container Management** — ALL container ops via `internal/adapters/containers/adapter.go`. No direct `docker`/`podman` commands.
5. **Configuration via HelixAgent Only** — CLI agent configs generated ONLY by `./bin/helixagent --generate-agent-config=<name>`.
6. **Real Data** — Beyond unit tests, use actual API calls, real databases, live services.
7. **Health & Observability** — Health endpoints, circuit breakers, Prometheus/OpenTelemetry.
8. **Documentation & Quality** — Update CLAUDE.md, AGENTS.md. Pass `make fmt vet lint security-scan`.
9. **No Mocks in Production** — Mocks, stubs, TODO implementations FORBIDDEN in production.
10. **Third-Party Submodules** — `cli_agents/` and `MCP/` are read-only. Use `git submodule update --remote`.
11. **Container-Based Builds** — ALL release builds inside Docker/Podman. `make release` / `make release-all`.
12. **Mandatory Container Rebuild** — Rebuild containers after code changes: `make docker-build && make docker-run`.
13. **Infrastructure Before Tests** — Start containers before tests: `make test-infra-start`.
14. **HTTP/3 (QUIC) with Brotli** — Primary transport. HTTP/2 fallback. Brotli → gzip compression.
15. **Resource Limits** — ALL tests limited to 30-40% host resources: `GOMAXPROCS=2 nice -n 19 ionice -c 3`.

## Build Commands

```bash
make build              # Build binary (output in bin/)
make build-debug        # Build with debug symbols
make build-all          # Build for all platforms
make run                # Run locally
make run-dev            # Development mode (GIN_MODE=debug)
make docker-build       # Build Docker image
make docker-run         # Start services with Docker Compose
```

### Release Builds (Container-Based)

```bash
make release              # Build helixagent for all platforms
make release-all          # Build ALL 7 apps for all platforms
make release-<app>        # Build specific app
make release-force        # Force rebuild (ignore change detection)
make release-info         # Show version codes and source hashes
make release-clean        # Clean release artifacts
```

## Linting & Formatting

```bash
make fmt                # Format with go fmt
make vet                # Run go vet
make lint               # Run golangci-lint
make security-scan      # Run gosec
make ci-validate-all    # All validation checks
make ci-pre-commit      # Pre-commit (fmt, vet)
make ci-pre-push        # Pre-push (includes unit tests)
```

**Always run `make fmt vet lint` before committing.**

## Testing Commands

### Running Tests

```bash
make test               # All tests (verbose)
make test-unit          # Unit tests only (./internal/... -short)
make test-integration   # Integration tests
make test-e2e           # End-to-end tests
make test-security      # Security tests
make test-stress        # Stress tests
make test-chaos         # Challenge tests
make test-bench         # Benchmark tests
make test-fuzz          # Fuzz tests (corpus replay)
make test-race          # Race detection
make test-coverage      # Coverage with HTML report
```

### Running a Single Test

```bash
go test -v -run TestFunctionName ./path/to/package
go test -v ./internal/llm
go test -v -run "Test.*Integration" ./...

# Resource-limited (CRITICAL)
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -v -p 1 -run TestName ./path/to/package
```

### Test Infrastructure

```bash
make test-infra-start   # Start PostgreSQL, Redis, Mock LLM containers
make test-infra-stop    # Stop test infrastructure
make test-with-infra    # Run tests with Docker infra
```

**IMPORTANT:** Containers MUST be running before tests/challenges.

## Infrastructure & Monitoring

```bash
make infra-start        # Start ALL infra (auto-detects Docker/Podman)
make infra-stop / restart / status
make infra-core         # Core: PostgreSQL, Redis, ChromaDB, Cognee
make monitoring-status / circuit-breakers / provider-health / fallback-chain
make monitoring-reset-circuits / force-health-check
```

## Code Style Guidelines

### Formatting & Imports
- Use `gofmt` / `goimports`
- Imports grouped: stdlib, third-party, internal (blank line separated)
- Line length: ≤ 100 characters

### Naming Conventions
- `camelCase`: local variables, private functions
- `PascalCase`: exported functions, types, constants, fields
- `UPPER_SNAKE_CASE`: exported constants
- Acronyms all caps: `HTTP`, `URL`, `ID`, `JSON`
- Receiver names: 1-2 letters (`s` for service, `c` for client)

### Error Handling
```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()
```

### Types & Interfaces
- Use `interface` to define behavior, not data
- Prefer small, focused interfaces
- Avoid `any`/`interface{}`; use generics

### Concurrency
- Always use `context.Context`
- Protect shared data with `sync.Mutex` or `sync.RWMutex`
- Use `sync.WaitGroup` for goroutine coordination

### Testing Patterns
- Write table-driven tests
- Use `testify` assertion library
- Mocks/stubs ONLY in unit tests

## Key Conventions

### Tool Schema
All tool parameters use **snake_case**. See `internal/tools/schema.go`.

### No Comments
**DO NOT ADD COMMENTS** unless explicitly requested.

### Git Operations
- **SSH ONLY** — HTTPS is forbidden
- Branch naming: `feat/`, `fix/`, `chore/`, `docs/`, `refactor/`, `test/` + description
- Commits: Conventional Commits (`feat(scope): description`)
- Run `make fmt vet lint` before committing

### Containerization

**CRITICAL: ALL container orchestration handled AUTOMATICALLY by HelixAgent binary.**

**FORBIDDEN:**
- Manual `docker`/`podman` commands
- `make test-infra-start` (use HelixAgent binary)
- Manual SSH for container deployment

**ONLY ACCEPTABLE WORKFLOW:**
1. `make build` — Build binary
2. `./bin/helixagent` — Run (orchestrates containers automatically)
3. Binary reads `Containers/.env` and orchestrates everything

**Key Files:**
- `Containers/.env` — Container orchestration config
- `internal/services/boot_manager.go` — Boot orchestration
- `tests/precondition/containers_boot_test.go` — Precondition tests

## Resource Limits (CRITICAL)

ALL test/challenge execution limited to 30-40% host resources:
```bash
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -v -p 1 ./...
```

## Quick Reference

| Task | Command |
|------|---------|
| Build | `make build` |
| Run | `./bin/helixagent` |
| Format | `make fmt` |
| Lint | `make lint` |
| All tests | `make test` |
| Single test | `go test -v -run TestName ./path/to/pkg` |
| Pre-commit | `make fmt vet lint` |
| Release | `make release` |

## Extracted Modules (27 Submodules)

Each module is independent with its own `go.mod`, tests, `CLAUDE.md`, `AGENTS.md`, `README.md`, `docs/`.

### Building and Testing Modules

```bash
cd <ModuleDir> && go build ./... && GOMAXPROCS=2 go test ./... -count=1 -race

# Test all modules
for mod in EventBus Concurrency Observability Auth Storage Streaming \
           Security VectorDB Embeddings Database Cache \
           Messaging Formatters MCP_Module RAG Memory Optimization Plugins \
           Agentic LLMOps SelfImprove Planning Benchmark HelixMemory \
           HelixSpecifier Containers Challenges; do
  (cd $mod && go test ./... -count=1 -race -short)
done
```

### Phase 1: Foundation

| Module | Path | Description |
|--------|------|-------------|
| EventBus | `EventBus/` | Pub/sub, topic filtering, middleware |
| Concurrency | `Concurrency/` | Worker pools, rate limiters, circuit breakers |
| Observability | `Observability/` | OpenTelemetry, Prometheus, logging |
| Auth | `Auth/` | JWT, API key, OAuth, middleware |
| Storage | `Storage/` | S3/MinIO, local, cloud abstraction |
| Streaming | `Streaming/` | SSE, WebSocket, gRPC, webhooks |

### Phase 2: Infrastructure

| Module | Path | Description |
|--------|------|-------------|
| Security | `Security/` | Guardrails, PII detection, filtering |
| VectorDB | `VectorDB/` | Qdrant, Pinecone, Milvus, pgvector |
| Embeddings | `Embeddings/` | 6 providers, batch embedding |
| Database | `Database/` | PostgreSQL/pgx, SQLite, migrations |
| Cache | `Cache/` | Redis + in-memory, distributed cache |

### Phase 3: Services

| Module | Path | Description |
|--------|------|-------------|
| Messaging | `Messaging/` | Kafka + RabbitMQ unified broker |
| Formatters | `Formatters/` | 32+ code formatters |
| MCP | `MCP_Module/` | Model Context Protocol framework |

### Phase 4: Integration

| Module | Path | Description |
|--------|------|-------------|
| RAG | `RAG/` | Chunking, retrieval, reranking |
| Memory | `Memory/` | Mem0-style with entity graphs |
| Optimization | `Optimization/` | GPT-Cache, Outlines, SGLang |
| Plugins | `Plugins/` | Plugin lifecycle, sandboxing |

### Phase 5: AI/ML

| Module | Path | Description |
|--------|------|-------------|
| Agentic | `Agentic/` | Graph-based workflow orchestration |
| LLMOps | `LLMOps/` | Evaluation, A/B testing, prompt versioning |
| SelfImprove | `SelfImprove/` | RLHF, reward modeling |
| Planning | `Planning/` | HiPlan, MCTS, Tree of Thoughts |
| Benchmark | `Benchmark/` | SWE-bench, HumanEval, MMLU |

### Phase 6: Cognitive

| Module | Path | Description |
|--------|------|-------------|
| HelixMemory | `HelixMemory/` | Unified memory engine (Mem0, Cognee, Letta, Graphiti fusion). Active by default. |

### Phase 7: Specification

| Module | Path | Description |
|--------|------|-------------|
| HelixSpecifier | `HelixSpecifier/` | Spec-Driven Development Engine (SpecKit + Superpowers + GSD). Active by default. |

### Pre-existing

| Module | Path | Description |
|--------|------|-------------|
| Containers | `Containers/` | Docker/Podman/K8s orchestration |
| Challenges | `Challenges/` | Challenge framework, 19 evaluators, userflow testing |

## Architecture

### Entry Points
- `cmd/helixagent/` — Main app
- `cmd/api/` — API server
- `cmd/grpc-server/` — gRPC

### Core Packages (`internal/`)
- `llm/providers/` — 22+ dedicated providers + generic OpenAI-compatible
- `llm/discovery/` — 3-tier model discovery
- `llm/ensemble.go` — Ensemble orchestration
- `services/` — Business logic (provider_registry, debate_service, etc.)
- `handlers/` — HTTP handlers (BackgroundTask, Discovery, Scoring, Verification, Health)
- `debate/` — 13 packages (agents, topology, voting, reflexion, gates, etc.)
- `formatters/` — 32+ code formatters
- `tools/` — Tool schema registry (21 tools)
- `agents/` — CLI agent registry (46 agents)
- `mcp/adapters/` — 45+ MCP adapters
- `verifier/` — Startup verification
- `adapters/` — Bridge to extracted modules (20+ files)

### Key Interfaces
- `LLMProvider` — Complete, CompleteStream, HealthCheck, GetCapabilities
- `VotingStrategy` — Ensemble voting
- `Formatter` — Code formatter
- Vector stores: Connect, Upsert, Search, Delete, Get

### Goroutine Lifecycle Safety
All handlers with background goroutines use `sync.WaitGroup` lifecycle tracking. Pattern: `WaitGroup.Add(1)` before launch, `defer WaitGroup.Done()`, `Shutdown()` calls `cancel()` + `WaitGroup.Wait()`.

### Architectural Patterns
- **Provider Registry** — Multi-provider interface with credentials
- **Ensemble Strategy** — Confidence-weighted voting, parallel execution
- **AI Debate** — Multi-round, 5 positions × 5 LLMs, 8-phase protocol
- **Debate Voting** — 6 methods (Weighted, Majority, Borda, Condorcet, Plurality, Unanimous)
- **Circuit Breaker** — Fault tolerance for providers
- **Dynamic Model Discovery** — 3-tier (API → models.dev → fallback)

## Startup Verification Pipeline

LLMsVerifier is the single source of truth. On startup: discover → verify (8-test pipeline) → score → rank → select debate team → start.

**Provider types**: API Key, OAuth, Free

**Scoring**: ResponseSpeed 25%, CostEffectiveness 25%, ModelEfficiency 20%, Capability 20%, Recency 10%

## Provider Access Mechanisms

- **Claude**: `claude -p --output-format json` (CLI proxy)
- **Qwen**: `qwen --acp` (JSON-RPC 2.0)
- **Zen**: HTTP server `opencode serve :4096`
- **Junie**: CLI mode with `--output-format json` and ACP mode via `junie --acp`

## CLI Agents (46)

Registry: `internal/agents/registry.go`

Generate configs: `./bin/helixagent --generate-agent-config=<name>`

### CLI Agent Config Rules (MANDATORY)

1. **Config filenames**: `opencode.json` (NO leading dot)
2. **No env var syntax**: Real API key values, not `{env:VAR_NAME}`
3. **Two versions**: Repo examples use placeholder, installed use real keys
4. **Config locations**: `~/.config/opencode/opencode.json`
5. **Model ID format**: `helixagent/helixagent-debate`
6. **15+ MCP servers**: 6 HelixAgent remote + 3 extended + 6 local + 3 free remote
7. **10+ Plugins**: helixagent-mcp, helixagent-lsp, helixagent-acp, etc.
8. **Extensions**: LSP, ACP, Embeddings, RAG, 8+ Skills

## Code Formatters

32+ formatters (11 native, 14 service, 7 built-in) for 19 languages.

REST API: `POST /v1/format`, `GET /v1/formatters`

## MCP Adapters

45+ adapters in `internal/mcp/adapters/`. 65+ containerized MCP servers (ports 9101-9999).

## CI/CD Container Build System

Five-phase system running inside Docker/Podman:

```bash
make ci-all              # All five phases + report
make ci-go               # Phase 1: Go builds + tests
make ci-mobile           # Phase 2: Flutter/RN
make ci-web              # Phase 3: Angular + Playwright
make ci-desktop          # Phase 4: Electron/Tauri
make ci-integration      # Phase 5: Full-stack integration
make ci-report           # Aggregate reports
CI_RESOURCE_LIMIT=medium make ci-all
```

## Challenges

```bash
./challenges/scripts/run_all_challenges.sh                       # All challenges
./challenges/scripts/release_build_challenge.sh                  # 25 tests
./challenges/scripts/debate_orchestrator_challenge.sh            # 61 tests
./challenges/scripts/helixmemory_challenge.sh                    # 80+ tests
./challenges/scripts/helixspecifier_challenge.sh                 # 138 tests
./challenges/scripts/cli_agent_config_challenge.sh               # 60 tests
./challenges/scripts/all_agents_e2e_challenge.sh                 # 102 tests
./challenges/scripts/full_system_boot_challenge.sh               # 53 tests
./challenges/scripts/ci_container_build_challenge.sh             # 87 tests
./challenges/scripts/snyk_automated_scanning_challenge.sh        # 38 tests
./challenges/scripts/sonarqube_automated_scanning_challenge.sh   # 45 tests
# ... and 460+ more challenge scripts
```

**IMPORTANT:** Start containers before challenges with `make test-infra-start`.

## LLMsVerifier

```bash
make verifier-init / verifier-build / verifier-test
make verifier-verify MODEL=gpt-4 PROVIDER=openai
```

## Protocol Endpoints

MCP `/v1/mcp` | ACP `/v1/acp` | LSP `/v1/lsp` | Embeddings `/v1/embeddings` | Vision `/v1/vision` | Cognee `/v1/cognee` | Startup `/v1/startup/verification` | BigData `/v1/bigdata/health` | Tasks `/v1/tasks` | Discovery `/v1/discovery` | Scoring `/v1/scoring` | Verification `/v1/verification` | Health `/v1/health`

## Configuration

Env vars in `.env.example`: `PORT`, `GIN_MODE`, `JWT_SECRET`, `DB_*`, `REDIS_*`, `*_API_KEY`, `*_USE_OAUTH_CREDENTIALS`, `COGNEE_ENABLED`, `CONSTITUTION_WATCHER_ENABLED`.

Service overrides: `SVC_<SERVICE>_<FIELD>` (e.g., `SVC_POSTGRESQL_HOST`).

## Technology Stack

Gin v1.11.0, PostgreSQL 15 (pgx/v5), Redis 7, testify v1.11.1, Prometheus/Grafana, OpenTelemetry. Docker and Podman support.

## Key Files

- `CLAUDE.md` — Detailed architecture
- `Makefile` — All commands
- `go.mod` — Module dependencies
- `docs/MODULES.md` — Modules catalog
- `.env.example` — Environment templates

## Adding a New LLM Provider

1. Create `internal/llm/providers/<name>/<name>.go` implementing `LLMProvider`
2. Add tool support if applicable (`SupportsTools: true`)
3. Register in `internal/services/provider_registry.go`
4. Add env vars to `.env.example`, tests in `internal/llm/providers/<name>/<name>_test.go`
