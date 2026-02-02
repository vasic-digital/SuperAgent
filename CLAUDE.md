# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

HelixAgent is an AI-powered ensemble LLM service written in Go that combines responses from multiple language models using intelligent aggregation strategies. It provides OpenAI-compatible APIs and supports 10 LLM providers (Claude, DeepSeek, Gemini, Mistral, OpenRouter, Qwen, ZAI, Zen, Cerebras, Ollama) with **dynamic provider selection** based on LLMsVerifier verification scores.

**Module**: `dev.helix.agent` (Go 1.24+, toolchain go1.24.11)

Subprojects: **Toolkit** (`Toolkit/`) — Go library for AI apps. **LLMsVerifier** (`LLMsVerifier/`) — provider accuracy verification.

## Mandatory Development Standards

**These rules are NON-NEGOTIABLE and MUST be followed for every component, service, or feature.**

1. **100% Test Coverage** — Every component MUST have unit, integration, E2E, automation, security/penetration, and benchmark tests. No false positives. Mocks/stubs ONLY in unit tests; all other tests use real data and live services.
2. **Challenge Coverage** — Every component MUST have Challenge scripts (`./challenges/scripts/`) validating real-life use cases. No false success — validate actual behavior, not return codes.
3. **Containerization** — All services MUST run in containers (Docker/Podman/K8s). Must support local default execution AND remote configuration. Auto boot-up before HelixAgent is ready. Remote services need API-based health checks.
4. **Configuration via HelixAgent Only** — CLI agent config export uses only HelixAgent + LLMsVerifier's unified generator (`pkg/cliagents/`). No third-party scripts.
5. **Real Data** — Beyond unit tests, all components MUST use actual API calls, real databases, live services. No simulated success. Fallback chains tested with actual failures.
6. **Health & Observability** — Every service MUST expose health endpoints. Circuit breakers for all external deps. Prometheus/OpenTelemetry integration. Status via `/v1/monitoring/status`.
7. **Documentation & Quality** — Follow existing patterns. Update CLAUDE.md, AGENTS.md, relevant docs. Pass `make fmt vet lint security-scan`. Conventional Commits: `<type>(<scope>): <description>`.
8. **Validation Before Release** — Pass `make ci-validate-all`, `./challenges/scripts/run_all_challenges.sh`, `make test-with-infra`, and benchmark/stress tests.
9. **No Mocks in Production** — Mocks, stubs, fakes, placeholders, TODO implementations STRICTLY FORBIDDEN in production code. All production code must be fully functional with real integrations.
10. **Third-Party Submodules** — `cli_agents/` and `MCP/` are read-only third-party deps; NEVER commit/push changes. Only project-owned submodules (LLMsVerifier, formatters) may be updated. Use `git submodule update --remote`.

## Git Rules

- **SSH only** for cloning — HTTPS may not work. Use SSH URLs in `.gitmodules`.
- **Branch naming**: `feat/`, `fix/`, `chore/`, `docs/`, `refactor/`, `test/` + short description
- **Commits**: Conventional Commits — `feat(llm): add ensemble voting strategy`
- **Always run `make fmt vet lint` before committing**

## Code Style

- Standard Go conventions ([Effective Go](https://go.dev/doc/effective_go)), `gofmt` formatting
- Imports grouped: stdlib, third-party, internal (blank line separated). Use `goimports`.
- Line length ≤ 100 chars (readability first)
- Naming: `camelCase` private, `PascalCase` exported, `UPPER_SNAKE_CASE` constants, acronyms all-caps (`HTTP`, `URL`, `ID`)
- Receivers: 1-2 letters (`s` for service, `c` for client)
- Errors: always check, wrap with `fmt.Errorf("...: %w", err)`, `defer` for cleanup
- Interfaces: small/focused, accept interfaces return structs
- Concurrency: `context.Context` always, `sync.Mutex`/`sync.RWMutex` for shared data
- Tests: table-driven, `testify`, naming `Test<Struct>_<Method>_<Scenario>`

## Build & Run

```bash
make build                # Build binary
make build-debug          # Build with debug symbols
make run                  # Run locally
make run-dev              # Development mode (GIN_MODE=debug)
make docker-build         # Build Docker image
docker-compose up -d      # Start full stack
```

## Testing

```bash
make test                 # All tests (auto-detects infra)
make test-unit            # Unit tests (./internal/... -short)
make test-integration     # Integration tests (./tests/integration)
make test-e2e             # E2E tests (./tests/e2e)
make test-security        # Security tests (./tests/security)
make test-stress          # Stress tests (./tests/stress)
make test-chaos           # Challenge tests (./tests/challenge)
make test-bench           # Benchmarks
make test-race            # Race detection
make test-coverage        # Coverage with HTML report
```

Single test: `go test -v -run TestName ./path/to/package`

With infrastructure:
```bash
make test-infra-start     # Start PostgreSQL, Redis, Mock LLM containers
DB_HOST=localhost DB_PORT=15432 DB_USER=helixagent DB_PASSWORD=helixagent123 DB_NAME=helixagent_db \
REDIS_HOST=localhost REDIS_PORT=16379 REDIS_PASSWORD=helixagent123 \
go test -v -run TestName ./path/to/package
make test-infra-stop      # Stop containers
make test-with-infra      # All tests with Docker infra
```

## Code Quality & CI

```bash
make fmt                  # go fmt
make vet                  # go vet
make lint                 # golangci-lint
make security-scan        # gosec
make install-deps         # Install dev tools
make ci-validate-all      # All validation checks
make ci-pre-commit        # Pre-commit (fmt, vet, fallback)
make ci-pre-push          # Pre-push (includes unit tests)
```

## Infrastructure & Monitoring

```bash
make infra-start          # Start ALL infra (auto-detects Docker/Podman)
make infra-stop / restart / status
make infra-core           # Core: PostgreSQL, Redis, ChromaDB, Cognee
make infra-mcp / lsp / rag
make monitoring-status / circuit-breakers / provider-health / fallback-chain
make monitoring-reset-circuits / force-health-check
```

## Architecture

### Entry Points
- `cmd/helixagent/` — Main app | `cmd/api/` — API server | `cmd/grpc-server/` — gRPC

### Core Packages (`internal/`)
- `llm/providers/` — 10 LLM providers (claude, deepseek, gemini, mistral, openrouter, qwen, zai, zen, cerebras, ollama)
- `llm/ensemble.go` — Ensemble orchestration
- `services/` — Business logic: provider_registry, ensemble, debate_service, debate_team_config, llm_intent_classifier, context_manager, mcp_client, lsp_manager, plugin_system
- `handlers/` — HTTP handlers | `middleware/` — Auth, rate limiting, CORS
- `background/` — Task queue, worker pool, resource monitor | `notifications/` — SSE, WebSocket, Webhooks
- `cache/` — Redis + in-memory | `database/` — PostgreSQL/pgx | `models/` — Data models/enums
- `debate/` — Orchestrator framework: agents, topology, protocol, voting, cognitive, knowledge (8 packages)
- `formatters/` — 32+ code formatters, REST API, middleware executor
- `tools/` — Tool schema registry (21 tools) | `agents/` — CLI agent registry (48 agents)
- `embedding/` — 6 providers (OpenAI, Cohere, Voyage, Jina, Google, Bedrock)
- `vectordb/` — Qdrant, Pinecone, Milvus, pgvector
- `mcp/adapters/` — 45+ MCP adapters | `mcp/config/` — Container config generator
- `rag/` — Hybrid retrieval | `memory/` — Mem0-style with entity graphs
- `routing/semantic/` — Embedding similarity routing
- `agentic/` — Graph-based workflow orchestration
- `security/` — Red team framework, guardrails, PII detection
- `observability/` — OpenTelemetry, Jaeger, Zipkin, Langfuse
- `bigdata/` — Infinite context, distributed memory, knowledge graph streaming
- `optimization/` — gptcache, outlines, streaming, sglang, llamaindex, langchain
- `verifier/` — Startup verification orchestrator and adapters

### Key Interfaces
- `LLMProvider` — Provider contract (Complete, CompleteStream, HealthCheck, GetCapabilities, ValidateConfig)
- `VotingStrategy` — Ensemble voting | `CacheInterface` — Cache abstraction
- `PluginRegistry`/`PluginLoader` — Plugin system | `TaskExecutor`/`TaskQueue` — Background tasks
- `Formatter` — Code formatter interface | Vector stores: `Connect`, `Upsert`, `Search`, `Delete`, `Get`

### Architectural Patterns
- **Provider Registry**: Unified multi-provider interface with credential management
- **Ensemble Strategy**: Confidence-weighted voting, majority vote, parallel execution
- **AI Debate**: Multi-round debate, 5 positions × 3 LLMs = 15 total, multi-pass validation (Initial → Validation → Polish → Final)
- **Debate Orchestrator**: Multi-topology (mesh/star/chain), phase protocol (Proposal → Critique → Review → Synthesis), cross-debate learning, auto-fallback to legacy
- **Circuit Breaker**: Fault tolerance for provider failures
- **Semantic Intent Detection**: LLM-based classification (zero hardcoding), pattern-based fallback
- **Fallback Error Reporting**: Categorized errors (rate_limit, timeout, auth, connection, unavailable, overloaded) in streamed responses

## Startup Verification Pipeline

LLMsVerifier is the **single source of truth**. On startup: discover providers → verify in parallel (8-test pipeline) → score (5 weighted components) → rank → select debate team → start server.

**Provider types**: API Key (DeepSeek, Gemini, Mistral, OpenRouter, ZAI, Cerebras), OAuth (Claude, Qwen), Free (Zen, OpenRouter :free)

**Scoring weights**: ResponseSpeed 25%, CostEffectiveness 25%, ModelEfficiency 20%, Capability 20%, Recency 10%. OAuth +0.5 bonus. Free: 6.0-7.0. Min score: 5.0.

Key files: `internal/verifier/startup.go`, `provider_types.go`, `adapters/oauth_adapter.go`, `adapters/free_adapter.go`

## Provider Access Mechanisms

OAuth/free providers use CLI proxies when direct API access is restricted:
- **Claude**: `claude -p --output-format json` (session continuity) — `internal/llm/providers/claude/claude_cli.go`
- **Qwen**: ACP via `qwen --acp` (JSON-RPC 2.0), fallback CLI — `internal/llm/providers/qwen/qwen_acp.go`
- **Zen**: HTTP server `opencode serve :4096`, fallback CLI — `internal/llm/providers/zen/zen_http.go`

Triggered when: `*_USE_OAUTH_CREDENTIALS=true` + no API key, or no `OPENCODE_API_KEY` for Zen.

**OAuth limitation**: CLI OAuth tokens are product-restricted (cannot use for general API). Get proper API keys from console.anthropic.com / dashscope.aliyuncs.com, or use CLI proxy.

## Configuration

Env vars in `.env.example`: `PORT`, `GIN_MODE`, `JWT_SECRET`, `DB_*`, `REDIS_*`, `*_API_KEY` for each provider, `*_USE_OAUTH_CREDENTIALS`, `COGNEE_ENABLED` (off by default; Mem0 is primary memory).

Service overrides: `SVC_<SERVICE>_<FIELD>` (e.g., `SVC_POSTGRESQL_HOST`, `SVC_REDIS_REMOTE=true`). Config files: `configs/development.yaml`, `configs/production.yaml`.

BigData components configured via `BIGDATA_ENABLE_*` env vars. Missing deps (Neo4j, ClickHouse, Kafka) gracefully degrade. Key file: `internal/bigdata/integration.go`.

## Adding a New LLM Provider

1. Create `internal/llm/providers/<name>/<name>.go` implementing `LLMProvider`
2. Add tool support if applicable (`SupportsTools: true` in GetCapabilities)
3. Register in `internal/services/provider_registry.go`
4. Add env vars to `.env.example`, tests in `internal/llm/providers/<name>/<name>_test.go`

## Tool Schema

All parameters use **snake_case**. Key files: `internal/tools/schema.go`, `internal/tools/handler.go`.

## CLI Agents (48)

Registry: `internal/agents/registry.go`. Generate configs: `./bin/helixagent --generate-agent-config=<name>`. All agents include formatters config. Config generation via LLMsVerifier's `pkg/cliagents/`.

## Code Formatters

32+ formatters (11 native, 14 service, 7 built-in) for 19 languages. REST API: `POST /v1/format`, `GET /v1/formatters`. Service formatters in Docker (ports 9210-9300). Core: `internal/formatters/` (interface, registry, executor, cache, system). Native providers: `internal/formatters/providers/native/`. AI debate integration: `internal/services/debate_formatter_integration.go`.

## MCP Adapters

45+ adapters in `internal/mcp/adapters/`. 65+ containerized MCP servers (ports 9101-9999, zero npx). Container config: `internal/mcp/config/generator_container.go`. Compose: `docker/mcp/docker-compose.mcp-full.yml`.

## Challenges

```bash
./challenges/scripts/run_all_challenges.sh                       # All challenges
./challenges/scripts/unified_verification_challenge.sh           # 15 tests
./challenges/scripts/llms_reevaluation_challenge.sh              # 26 tests
./challenges/scripts/debate_team_dynamic_selection_challenge.sh  # 12 tests
./challenges/scripts/semantic_intent_challenge.sh                # 19 tests
./challenges/scripts/fallback_mechanism_challenge.sh             # 17 tests
./challenges/scripts/integration_providers_challenge.sh          # 47 tests
./challenges/scripts/all_agents_e2e_challenge.sh                 # 102 tests
./challenges/scripts/full_system_boot_challenge.sh               # 53 tests
./challenges/scripts/cli_proxy_challenge.sh                      # 50 tests
```

## LLMsVerifier

```bash
make verifier-init / verifier-build / verifier-test
make verifier-verify MODEL=gpt-4 PROVIDER=openai
```

## Protocol Endpoints

MCP `/v1/mcp` | ACP `/v1/acp` | LSP `/v1/lsp` | Embeddings `/v1/embeddings` | Vision `/v1/vision` | Cognee `/v1/cognee` (optional) | Startup `/v1/startup/verification` | BigData `/v1/bigdata/health`

Fallback: routes to strongest LLM by score, falls back on failure.

## Technology Stack

Gin v1.11.0, PostgreSQL 15 (pgx/v5), Redis 7, testify v1.11.1, Prometheus/Grafana, OpenTelemetry. Supports Docker and Podman (`./scripts/container-runtime.sh`).

## Unified Service Management

`BootManager` (`internal/services/boot_manager.go`): groups services by compose file, starts via `docker compose up -d`, health checks all. `HealthChecker` (`internal/services/health_checker.go`): TCP/HTTP checks with retries. Required services (PostgreSQL, Redis, ChromaDB) fail boot on health failure. Remote services: health check only. SQL schemas: `sql/schema/`.
