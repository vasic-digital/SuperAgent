# AGENTS.md

This document provides guidance for AI agents working on the HelixAgent project. It includes essential commands for building, testing, and formatting code, as well as code style guidelines.

## Project Overview

HelixAgent is an AI-powered ensemble LLM service written in Go (1.25.3) that combines responses from multiple language models using intelligent aggregation strategies. It provides OpenAI-compatible APIs and supports **22 dedicated LLM providers** (Claude, Chutes, DeepSeek, Gemini, Mistral, OpenRouter, Qwen, ZAI, Zen, Cerebras, Ollama, AI21, Anthropic, Cohere, Fireworks, Groq, HuggingFace, OpenAI, Perplexity, Replicate, Together, xAI) plus a **generic OpenAI-compatible provider** for 17+ additional providers (NVIDIA, SambaNova, Hyperbolic, Novita, SiliconFlow, Kimi/Moonshot, Upstage, Codestral, DeepInfra, Baseten, NLP Cloud, etc.) with **dynamic provider selection** based on LLMsVerifier verification scores.

## Architecture Overview

HelixAgent uses a **unified startup verification pipeline** where LLMsVerifier acts as the single source of truth for LLM verification and scoring. On startup, the system:

1. **Discovers all providers** (API Key, OAuth, Free)
2. **Verifies all providers in parallel** (8-test pipeline)
3. **Scores verified providers** (5 weighted components: ResponseSpeed, ModelEfficiency, CostEffectiveness, Capability, Recency)
4. **Ranks by score** (OAuth priority when scores close)
5. **Selects AI Debate Team** (25 LLMs: 5 primary + 20 fallback)
6. **Starts server** with verified configuration

### Key Components
- **LLM Provider Registry**: Unified interface for 22 LLM providers with credential management
- **AI Debate System**: Multi-round debate between providers for consensus (5 positions × 5 LLMs = 25 total)
- **SpecKit Orchestrator**: 7-phase development flow (Constitution → Specify → Clarify → Plan → Tasks → Analyze → Implement) with auto-activation based on work granularity detection and phase caching for resumption
- **Constitution Management**: Auto-update Constitution on project changes (new modules, documentation changes, structure changes) with background filesystem monitoring
- **Plugin System**: Hot-reloadable plugins with dependency resolution
- **Formatters System**: 32+ code formatters for 19 programming languages via REST API
- **MCP Adapters**: 45+ Model Context Protocol adapters for external services (Linear, Asana, Jira, etc.)
- **Dynamic Model Discovery**: 3-tier model discovery for all providers — Tier 1: provider API (`/v1/models`), Tier 2: models.dev catalog, Tier 3: hardcoded fallback. Thread-safe caching with 1-hour TTL. Package: `internal/llm/discovery/`
- **Provider Subscription & Access Types**: 3-tier dynamic subscription detection (Tier 1: provider-specific APIs like OpenRouter `/v1/auth/key` and Cohere `/check-api-key`, Tier 2: rate limit header inference, Tier 3: static fallback). 6 subscription types (`free`, `free_credits`, `free_tier`, `pay_as_you_go`, `monthly`, `enterprise`). Per-provider auth mechanism configuration (Bearer, `x-api-key`, `x-goog-api-key`, anonymous). Rate limit header parsing for OpenAI, Anthropic, Groq, OpenRouter, SambaNova, HuggingFace. Subscription info factors into cost scoring. Package: `internal/verifier/`
- **Background Task System**: Task queue, worker pool, resource monitor, stuck detector
- **Notifications**: Real-time notifications via SSE, WebSocket, webhooks, polling

### Provider Types
- **API Key**: DeepSeek, Gemini, Mistral, OpenRouter, ZAI, Cerebras (Bearer token, full verification)
- **OAuth**: Claude, Qwen (OAuth2 tokens from CLI, trust on API failure)
- **Free**: Zen (OpenCode), OpenRouter free models (anonymous/X-Device-ID, reduced verification)

### Generic OpenAI-Compatible Provider
`internal/llm/providers/generic/generic.go` — Lightweight provider implementing `LLMProvider` for any OpenAI-compatible chat completions endpoint. Used by `createProviderForVerification()` as a fallback for providers without dedicated implementations. Supports Complete, CompleteStream, HealthCheck with Bearer auth. Enabled via `SupportedProviders` config in `internal/verifier/provider_types.go`.

### AI Debate Team
Dynamic selection via StartupVerifier: OAuth2 providers first, then LLMsVerifier-scored providers. 5 positions × 5 LLMs (1 primary + 4 fallbacks) = **25 LLMs**. **Provider diversity**: Fallback selection prioritizes unique providers (one model per provider) before filling remaining slots by score. OAuth primaries get non-OAuth fallbacks for redundancy. All verified API key providers now get `Instance` set via `SetInstanceCreator()` for debate participation.

### Multi-Pass Validation
Debate responses undergo re-evaluation, polishing, and improvement before final consensus:
1. **Initial Response** → 2. **Validation** → 3. **Polish & Improve** → 4. **Final Conclusion**

### AI Debate Orchestrator Framework

New framework with multi-topology support (mesh, star, chain), phase-based protocol (Proposal → Critique → Review → Synthesis), learning system (extracts lessons from debates), and automatic fallback to legacy services.

### Verification Failure Tracking
When a provider fails verification, detailed failure information is captured in `UnifiedProvider`:
- **FailureReason**: Human-readable explanation (e.g., "code visibility test failed. 2/3 tests passed (score: 45.0)")
- **FailureCategory**: Categorized failure type (`code_visibility_failed`, `score_below_threshold`, `api_error`, `timeout`, `auth_error`, `canned_response`, `empty_response`)
- **TestDetails**: Per-test breakdown (`ProviderTestDetail` with name, passed, score, details, duration)
- **VerificationMsg**: Summary message from the verification pipeline
- **LastModelResponse**: Truncated (200 chars) last response for debugging

Exposed at `/v1/startup/verification` → `ranked_providers[].failure_reason`, `failure_category`, `test_details`. Key files: `internal/verifier/provider_types.go` (struct), `internal/verifier/startup.go` (helpers: `buildFailureReason`, `categorizeFailure`, `mapTestDetails`).

### Fallback Error Reporting
When an LLM provider fails, detailed error information is included in streamed responses (rate_limit, timeout, auth, connection, unavailable, overloaded).

### Semantic Intent Detection
Uses **LLM-based semantic intent classification** (primary) with pattern-based fallback. Detects confirmation, refusal, question, request, clarification, unclear. **Zero hardcoding** - intent detected by semantic meaning, not exact string matching.

### SpecKit Auto-Activation
Automatically triggers 7-phase development flow for large changes and refactoring. Work granularity detection classifies requests into 5 levels:
- **Single Action** - Small changes (e.g., "Add a log statement")
- **Small Creation** - Minor features (e.g., "Fix typo in README")
- **Big Creation** - Significant features (e.g., "Implement logging system") → **Triggers SpecKit**
- **Whole Functionality** - Complete modules (e.g., "Build payment processing") → **Triggers SpecKit**
- **Refactoring** - Architectural changes (e.g., "Refactor to microservices") → **Triggers SpecKit**

**Phase Caching**: Each phase result cached in `.speckit/cache/` for resumption after interruption. **Flow Resumption**: Automatically detects incomplete flows and resumes from last completed phase.

Key files: `internal/services/speckit_orchestrator.go`, `enhanced_intent_classifier.go`

### Constitution Management
Background service that monitors project changes and auto-updates Constitution:
- **New Modules**: Detects `go.mod` files and adds decoupling rules
- **Documentation Changes**: Syncs Constitution across AGENTS.md/CLAUDE.md when modified
- **Structure Changes**: Tracks new top-level directories
- **Test Coverage**: Flags violations when coverage drops below 100%

Runs with configurable check interval (default: 5 minutes). Enable with `CONSTITUTION_WATCHER_ENABLED=true`.

Key files: `internal/services/constitution_watcher.go`, `constitution_manager.go`, `documentation_sync.go`

### Code Formatters System

32+ formatters (11 native, 14 service, 7 built-in) for 19 programming languages. REST API endpoints: `POST /v1/format`, `GET /v1/formatters`, etc. Service formatters run as Docker containers (ports 9210-9300). Integrated with all 48 CLI agents.

### MCP Adapters & Containerization

45+ MCP adapters for external services (Linear, Asana, Jira, Slack, GitHub, etc.). All MCP servers are containerized (ports 9101-9999), eliminating npm/npx dependencies.


## Quick Start

1. Install Go 1.24+ and Docker.
2. Clone the repository.
3. Run `make install-deps` to install development tools (golangci-lint, gosec).
4. Copy `.env.example` to `.env` and adjust settings if needed.
5. Run `make build` to build the binary.
6. Run `make test` to verify everything works.

## Build Commands

The project uses a Makefile with the following key targets:

### Core Build Commands
- `make build` – Build HelixAgent binary (output in `bin/`)
- `make build-debug` – Build with debug symbols
- `make build-all` – Build for all architectures (Linux, macOS, Windows)
- `make run` – Run HelixAgent locally
- `make run-dev` – Run in development mode (`GIN_MODE=debug`)

### Container Build Commands
- `make docker-build` – Build Docker image
- `make docker-run` – Start services with Docker Compose
- `make docker-stop` – Stop Docker services
- `make docker-clean` – Clean Docker containers and volumes
- `make docker-full` – Start full environment (all profiles)

## Linting & Formatting Commands

- `make fmt` – Format Go code with `go fmt`
- `make vet` – Run `go vet` for static analysis
- `make lint` – Run `golangci-lint` (install with `make install-deps`)
- `make security-scan` – Run `gosec` security scanner
- `make sbom` – Generate Software Bill of Materials (CycloneDX/SPDX)

Always run `make fmt vet lint` before committing.

## Testing Commands

### Basic Testing
- `make test` – Run all tests (verbose)
- `make test-unit` – Run unit tests only (`./internal/... -short`)
- `make test-coverage` – Run tests with coverage report (HTML output)
- `make test-bench` – Run benchmark tests
- `make test-race` – Run tests with race detection

### Specialized Test Suites
- `make test-integration` – Integration tests with Docker dependencies
- `make test-e2e` – End-to-end tests
- `make test-security` – Security tests (LLM penetration testing)
- `make test-stress` – Stress tests
- `make test-chaos` – Chaos/challenge tests (AI debate validation)

### Go Test Suites
- `tests/security/penetration_test.go` – LLM security testing (prompt injection, jailbreaking, data exfiltration)
- `tests/challenge/ai_debate_maximal_challenge_test.go` – AI debate system comprehensive validation
- `tests/integration/llm_mem0_verification_test.go` – All 22 LLM providers + Mem0 memory integration
- `tests/integration/mem0_full_integration_test.go` – Mem0 full integration (store, search, entity graph)
- `tests/integration/mem0_capacity_test.go` – Mem0 capacity and performance tests
- `tests/integration/mem0_resilience_test.go` – Mem0 resilience and recovery tests
- `tests/integration/mem0_llm_integration_test.go` – Mem0 + LLM provider integration tests
- `tests/integration/mem0_ensemble_integration_test.go` – Mem0 + ensemble voting integration
- `tests/integration/mem0_migration_test.go` – Cognee-to-Mem0 migration verification
- `tests/integration/grpc_integration_test.go` – 21 gRPC service integration tests
- `tests/integration/debate_full_flow_test.go` – 7 debate lifecycle integration tests
- `tests/integration/comprehensive_monitoring_test.go` – Full monitoring stack validation
- `tests/stress/bigdata_stress_test.go` – BigData concurrent stress tests
- `tests/stress/formatters_stress_test.go` – Formatters concurrent stress tests
- `tests/stress/memory_stress_test.go` – Memory CRDT/distributed stress tests
- `tests/stress/concurrency_safety_test.go` – Concurrency safety and race condition tests

### Test Infrastructure Management
- `make test-infra-start` – Start PostgreSQL, Redis, Mock LLM containers
- `make test-infra-stop` – Stop test infrastructure
- `make test-infra-clean` – Stop and remove volumes
- `make test-with-infra` – Run all tests with Docker infrastructure

### CI/CD Validation Commands

Pre-commit and pre-push validation targets:

- `make ci-validate-fallback` – Validate reliable fallback mechanism
- `make ci-validate-monitoring` – Validate monitoring systems
- `make ci-validate-all` – Run all validation checks
- `make ci-pre-commit` – Pre-commit validation (fmt, vet, fallback checks)
- `make ci-pre-push` – Pre-push validation (includes unit tests)

### Container Runtime Commands

Support for both Docker and Podman:

- `make container-detect` – Detect container runtime
- `make container-build` – Build container image (auto-detects runtime)
- `make container-start` – Start services (auto-detects runtime)
- `make container-stop` – Stop services
- `make container-logs` – Show logs
- `make container-status` – Check status

### Running a Single Test
To run a specific test or test suite, use the standard `go test` command:

```bash
# Run a single test function
go test -v -run TestFunctionName ./path/to/package

# Run all tests in a package
go test -v ./internal/llm

# Run tests matching a pattern
go test -v -run "Test.*Integration" ./...

# Run tests with coverage for a single package
go test -v -coverprofile=coverage.out ./internal/llm
```

## Infrastructure & Monitoring Commands

### Infrastructure Management
- `make infra-start` – Start ALL infrastructure (auto-detects Docker/Podman)
- `make infra-stop` – Stop ALL infrastructure
- `make infra-restart` – Restart ALL infrastructure
- `make infra-status` – Check ALL infrastructure status
- `make infra-core` – Start core services (PostgreSQL, Redis, ChromaDB, Cognee)
- `make infra-mcp` – Start MCP servers
- `make infra-lsp` – Start LSP servers
- `make infra-rag` – Start RAG services

### Monitoring Endpoints
- `make monitoring-status` – Check monitoring status (curl to /v1/monitoring/status)
- `make monitoring-circuit-breakers` – Check circuit breakers
- `make monitoring-oauth-tokens` – Check OAuth tokens
- `make monitoring-provider-health` – Check provider health
- `make monitoring-fallback-chain` – Check fallback chain
- `make monitoring-reset-circuits` – Reset all circuit breakers
- `make monitoring-validate-fallback` – Validate fallback chain
- `make monitoring-force-health-check` – Force provider health check

## LLMsVerifier Integration Commands

### Initialization & Building
- `make verifier-init` – Initialize LLMsVerifier submodule
- `make verifier-update` – Update LLMsVerifier submodule
- `make verifier-build` – Build verifier CLI
- `make verifier-docker-build` – Build verifier Docker image
- `make verifier-docker-run` – Run verifier in Docker
- `make verifier-docker-stop` – Stop verifier Docker services

### Testing
- `make verifier-test` – Run verifier tests
- `make verifier-test-unit` – Run verifier unit tests
- `make verifier-test-integration` – Run verifier integration tests
- `make verifier-test-e2e` – Run verifier E2E tests
- `make verifier-test-security` – Run verifier security tests
- `make verifier-test-stress` – Run verifier stress tests
- `make verifier-test-chaos` – Run verifier chaos tests
- `make verifier-test-all` – Run ALL verifier tests (6 types)
- `make verifier-test-coverage` – Run verifier tests with coverage
- `make verifier-test-coverage-100` – Check verifier 100% test coverage

### Operations
- `make verifier-run` – Run verifier service
- `make verifier-health` – Check verifier health
- `make verifier-verify MODEL=gpt-4 PROVIDER=openai` – Run model verification
- `make verifier-score MODEL=gpt-4` – Get model score
- `make verifier-providers` – List verified providers
- `make verifier-metrics` – Get verifier metrics
- `make verifier-db-migrate` – Run verifier database migrations
- `make verifier-db-sync` – Sync verifier database
- `make verifier-clean` – Clean verifier artifacts

### SDK Building
- `make verifier-sdk-go` – Build Go SDK for verifier
- `make verifier-sdk-python` – Build Python SDK for verifier
- `make verifier-sdk-js` – Build JavaScript SDK for verifier
- `make verifier-sdk-all` – Build all verifier SDKs
- `make verifier-docs` – Generate verifier documentation
- `make verifier-benchmark` – Run verifier benchmarks

## Challenges & Formatters Commands

### Challenge Scripts
- `make test-challenges` – Run performance challenges
- `make challenges-with-infra` – Run ALL challenges with full auto-started infrastructure
- `make challenge-infra` – Run comprehensive infrastructure challenge
- `make challenge-cli-agents` – Run all CLI agents E2E challenge
- `./challenges/scripts/run_all_challenges.sh` – Run all challenges
- `./challenges/scripts/main_challenge.sh` – Main challenge (generates OpenCode config)
- `./challenges/scripts/unified_verification_challenge.sh` – 15 tests - startup pipeline
- `./challenges/scripts/llms_reevaluation_challenge.sh` – 26 tests - provider re-evaluation on EVERY boot
- `./challenges/scripts/debate_team_dynamic_selection_challenge.sh` – 12 tests - team selection
- `./challenges/scripts/free_provider_fallback_challenge.sh` – 8 tests - Zen/free models
- `./challenges/scripts/semantic_intent_challenge.sh` – 19 tests - intent detection (ZERO hardcoding)
- `./challenges/scripts/fallback_mechanism_challenge.sh` – 17 tests - fallback chain for empty responses
- `./challenges/scripts/integration_providers_challenge.sh` – 47 tests - embedding/vector/MCP integrations
- `./challenges/scripts/all_agents_e2e_challenge.sh` – 102 tests - all 48 CLI agents
- `./challenges/scripts/cli_agent_mcp_challenge.sh` – 26 tests - CLI agent MCP validation (37 MCPs)
- `./challenges/scripts/full_system_boot_challenge.sh` – 53 tests - full system infrastructure validation
- `./challenges/scripts/grpc_service_challenge.sh` – 9 tests - gRPC service validation
- `./challenges/scripts/bigdata_comprehensive_challenge.sh` – 23 tests - BigData components
- `./challenges/scripts/memory_system_challenge.sh` – 14 tests - Memory system (Mem0, CRDT, distributed)
- `./challenges/scripts/mem0_migration_challenge.sh` – Mem0 migration verification (Cognee→Mem0)
- `./challenges/scripts/security_scanning_challenge.sh` – 10 tests - Security scanning tools
- `./challenges/scripts/constitution_watcher_challenge.sh` – 12 tests - Constitution auto-update
- `./challenges/scripts/speckit_auto_activation_challenge.sh` – 15 tests - SpecKit 7-phase flow

### Formatters System
- `make formatters` – (if exists) Start formatters system
- `curl -X POST http://localhost:7061/v1/format` – Format code via REST API
- `curl http://localhost:7061/v1/formatters` – List all formatters
- `curl http://localhost:7061/v1/formatters/detect?file_path=main.py` – Auto-detect formatter
- `./formatters/scripts/init-submodules.sh` – Initialize all formatter submodules
- `./formatters/scripts/build-all.sh` – Build native binaries
- `./formatters/scripts/health-check-all.sh` – Health check all formatters

### Formatter Services (Docker)
- `docker-compose -f docker-compose.formatters.yml up -d` – Start all service formatters (ports 9210-9300)
- `./docker/formatters/build-all.sh` – Build all service formatter containers

## Important Gotchas & Environment Variables

### OAuth Token Restrictions
- **Claude OAuth tokens** from `claude auth login` are **restricted to Claude Code only** and cannot be used for general API calls. Use API key from https://console.anthropic.com/.
- **Qwen OAuth tokens** from Qwen CLI login are for `portal.qwen.ai` only; DashScope API requires separate API key from https://dashscope.aliyuncs.com/.
- **Solution**: Get proper API keys or use CLI proxy mechanisms (enabled via `CLAUDE_CODE_USE_OAUTH_CREDENTIALS=true`, `QWEN_CODE_USE_OAUTH_CREDENTIALS=true`).

### Provider Access Mechanisms
- **Claude**: JSON CLI (`claude -p --output-format json`) with session continuity.
- **Qwen**: ACP (Agent Communication Protocol) via `qwen --acp` (JSON-RPC 2.0 over stdin/stdout), fallback to CLI proxy.
- **Zen/OpenCode**: HTTP server (`opencode serve` on port 4096), fallback to CLI proxy with JSON output.
- **Trigger conditions**: OAuth/free mode when no API key provided.

### Environment Variables
Key environment variables (see `.env.example`):
- Server: `PORT`, `GIN_MODE`, `JWT_SECRET`
- Database: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- Redis: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- LLM providers: `CLAUDE_API_KEY`, `DEEPSEEK_API_KEY`, `GEMINI_API_KEY`, `OPENCODE_API_KEY` (optional), `MISTRAL_API_KEY`, `OPENROUTER_API_KEY`, `ZAI_API_KEY`, `CEREBRAS_API_KEY`, `OLLAMA_BASE_URL`
- OAuth flags: `CLAUDE_CODE_USE_OAUTH_CREDENTIALS`, `QWEN_CODE_USE_OAUTH_CREDENTIALS`
- Cognee (optional, disabled by default): `COGNEE_ENABLED=true`, `COGNEE_AUTH_EMAIL`, `COGNEE_AUTH_PASSWORD`

### Service Configuration
- **Mem0** is the primary memory provider (PostgreSQL backend with entity graphs). Cognee can be enabled but is disabled by default.
- **Service endpoints** configured in `configs/development.yaml` and `configs/production.yaml`. Override via `SVC_<SERVICE>_<FIELD>` environment variables (e.g., `SVC_POSTGRESQL_HOST`, `SVC_REDIS_REMOTE=true`).

### Tool Schema Convention
- All tool parameters use **snake_case** (e.g., `file_path`, `old_string`). See `internal/tools/schema.go`.

### CLI Agent Registry
- **48 CLI agents** supported (original 18 + extended 30). Generate configs with `./bin/helixagent --generate-agent-config=<agent>`.
- All agents include formatters configuration with smart defaults.
- **Config filenames**: `opencode.json` (WITHOUT leading dot), `crush.json`. OpenCode v1.2.6+ does NOT recognize `.opencode.json`.
- **No env var syntax**: CLI agents do NOT support `{env:VAR}` in API keys. Installed configs MUST use real values.
- **Two versions**: Repo examples (`configs/cli-agents/`) use `<YOUR_HELIXAGENT_API_KEY>` placeholder. Installed configs use real API key.
- **Locations**: OpenCode → `~/.config/opencode/opencode.json`. Crush → `~/.config/crush/crush.json`.
- **15+ MCP servers** per agent: 6 HelixAgent remote + 3 extended + 6 local npx + 3 free remote.
- **10+ Plugins** per agent: helixagent-mcp, helixagent-lsp, helixagent-acp, helixagent-embeddings, helixagent-vision, helixagent-rag, helixagent-formatters, helixagent-debate, helixagent-memory, helixagent-monitoring.
- **Extensions** per agent: LSP, ACP, Embeddings, RAG, 8+ Skills (code-review, code-format, semantic-search, vision-analysis, memory-recall, rag-retrieval, lsp-diagnostics, agent-communication).
- **No hardcoding**: All config values come from the unified generator system (`LLMsVerifier/llm-verifier/pkg/cliagents/`).
- **Challenge**: `./challenges/scripts/cli_agent_config_challenge.sh` validates all 48 agents have required features (~60 tests).

### MCP Adapters & Containerization
- **45+ MCP adapters** for external services (Linear, Asana, Jira, Slack, etc.).
- **65+ MCP servers** available as Docker containers (ports 9101-9999), eliminating npm/npx dependencies.
- Containerized MCPs are auto‑configured; use `docker-compose -f docker/mcp/docker-compose.mcp-full.yml up -d`.

### Testing Notes
- Tests auto‑detect infrastructure; run `make test-infra-start` before full test suite.
- Integration tests require PostgreSQL, Redis, and Mock LLM containers.
- Challenge scripts (`./challenges/scripts/`) validate system functionality comprehensively.

### Git Service Access
- **Accessing Git services for cloning can be done only via SSH!** HTTPS cloning may not work due to network restrictions or authentication requirements.
- Ensure SSH keys are properly configured and added to your Git service account (GitHub, GitLab, etc.).
- For submodule initialization and updates, use SSH URLs in `.gitmodules` or configure Git to use SSH for all operations.

## Mandatory Development Standards

**These rules are NON-NEGOTIABLE and MUST be followed for every component, service, or feature added to HelixAgent:**

### 1. Comprehensive Test Coverage (100% Minimum)
- **Every component or service MUST have comprehensive tests with coverage no less than 100%**
- **Multiple test types REQUIRED** for each component:
  - **Unit tests** (isolated, can use mocks/stubs)
  - **Integration tests** (with real dependencies)
  - **End-to-end (E2E) tests** (full user scenarios)
  - **Full automation tests** (CI/CD pipeline validation)
  - **Security & penetration tests** (LLM security testing, prompt injection, jailbreaking, data exfiltration)
  - **Benchmark tests** (performance validation)
- **No false positives allowed** – every successful test execution MUST be verified and validated for real success
- **Mocked or stubbed data ONLY allowed for unit tests** – all other test types MUST use real data, integrations, and components

### 2. Challenge Coverage for Real-Life Validation
- **Every component or service MUST be covered with a variety of Challenges** that verify real-life use cases
- **Challenge scripts** (in `./challenges/scripts/`) must be created for each major feature
- **No false positives in Challenges** – false success is strictly prohibited
- **Challenge execution must validate actual system behavior**, not just return codes

### 3. Containerization & Infrastructure Standards
- **All servers, integration components, services, and servers MUST run inside containers** using Compose technology (Docker, Podman, or Kubernetes)
- **Default local execution**: Every service must run on the current host machine by default
- **Remote configuration capability**: Must support configuration for external services hosted on entirely different physical machines
- **Automatic boot-up**: When HelixAgent starts, all containers must boot up and reach full capacity state before HelixAgent is considered ready to run
- **Remote service health checks**: If services are remote, health checks MUST be executed using API calls with proper error handling
- **Centralized Container Management**: ALL container operations (runtime detection, compose up/down, health checks, remote distribution) MUST go through the Containers module (`digital.vasic.containers`) via `internal/adapters/containers/adapter.go`. No direct `exec.Command` to `docker`/`podman` in production code. The adapter provides: `DetectRuntime`, `ComposeUp`, `ComposeDown`, `HealthCheck`, `HealthCheckHTTP`, `HealthCheckTCP`, `Distribute`, `Shutdown`, `NewServiceOrchestrator`. Remote distribution enabled via `CONTAINERS_REMOTE_*` env vars with SSH-based remote execution, resource-aware scheduling (5 strategies), SSH tunnel networking, and SSHFS/NFS/rsync volume management.
- **Service Orchestrator**: Automatic service discovery and management via `pkg/orchestrator/` in Containers module. Auto-discovers all docker-compose files in `docker/` directory. When `CONTAINERS_REMOTE_ENABLED=true`, ALL services are automatically deployed to remote host(s) with automatic fallback to local. Thread-safe service management with parallel startup.

### 4. Configuration Management
- **Exporting configuration files for all supported CLI agents MUST be done using only HelixAgent**
- **No third-party scripts** are allowed for configuration export
- **Configuration generation** must use the unified generator (`pkg/cliagents/`) in LLMsVerifier

### 5. Real Data & Integration Requirements
- **Real integrations required**: All components beyond unit tests MUST use actual API calls, real databases, and live services
- **No simulated success**: Integration tests must validate actual connectivity and functionality
- **Fallback chains must be real**: Fallback mechanisms must be tested with actual provider failures

### 6. Health Monitoring & Observability
- **Every service MUST expose health endpoints** following the project's health check pattern
- **Circuit breakers must be implemented** for all external dependencies
- **Monitoring integration** with Prometheus, OpenTelemetry, and logging systems
- **Real-time status reporting** through `/v1/monitoring/status` and other monitoring endpoints

### 7. Documentation & Code Quality
- **All new code MUST follow existing project patterns** and architecture
- **Documentation must be updated** in CLAUDE.md, AGENTS.md, and relevant docs
- **Code must pass all linting, formatting, and security scans** (`make fmt vet lint security-scan`)
- **Commit messages MUST follow Conventional Commits** with proper scoping

### 8. HTTP/3 (QUIC) with Brotli Compression (MANDATORY)
- **ALL HTTP communication MUST use HTTP/3 (QUIC) as the primary transport protocol** with Brotli compression
- **HTTP/2 is ONLY allowed as a fallback** when HTTP/3 is not supported by the remote endpoint
- **Compression priority**: Brotli (primary) → gzip (fallback only when Brotli is unavailable)
- **All HTTP clients** (provider API calls, health checks, MCP connections, webhook deliveries, etc.) MUST prefer HTTP/3 with Brotli
- **All HTTP servers** (HelixAgent API, gRPC-gateway, SSE endpoints, etc.) MUST support HTTP/3 and advertise it via `Alt-Svc` headers
- **Go implementation**: Use `quic-go/quic-go` for HTTP/3 transport and `andybalholm/brotli` for compression
- **Testing**: All HTTP integration tests MUST verify HTTP/3 capability where applicable

### 9. Resource Limits for Tests and Challenges (CRITICAL)
- **ALL test and challenge execution MUST be strictly limited to 30-40% of host system resources**
- **This is a CRITICAL constraint** — violating this has caused host machine crashes and full system resets
- **CPU limiting**: Use `GOMAXPROCS=2` (or proportional to available cores), `nice -n 19` for all test processes
- **I/O limiting**: Use `ionice -c 3` (idle class) for all test and challenge processes
- **Sequential execution**: Use `-p 1` flag for `go test` to prevent parallel package execution
- **Container limits**: All test infrastructure containers MUST have CPU and memory limits set
- **Challenge scripts**: MUST NOT spawn unbounded parallel processes; use controlled concurrency
- **Monitoring**: Resource usage MUST be checked before and during test execution
- **Rationale**: The host machine runs mission-critical processes; tests and challenges are secondary workloads

### 10. Validation Before Release
- **All components MUST pass the complete validation suite** (`make ci-validate-all`)
- **Challenge scripts MUST execute successfully** (`./challenges/scripts/run_all_challenges.sh`)
- **Infrastructure MUST be validated** (`make test-with-infra`)
- **Performance MUST meet benchmarks** (stress tests, chaos tests)

### 11. No Mock Implementations in Production Code
- **Mocks, stubs, and fakes are STRICTLY FORBIDDEN in production code**
- **Mocked or stubbed data ONLY allowed for unit tests** – all other components MUST use real implementations
- **No placeholders, no TODO implementations that return mock data** – all production code must be fully functional
- **Real integrations required**: All production components MUST use actual API calls, real databases, and live services
- **Integration tests must validate actual connectivity and functionality** – no simulated success allowed
- **Fallback chains must be real**: Fallback mechanisms must be tested with actual provider failures

### 12. Third-Party Submodule Management
- **Third-party submodules (CLI agents, MCP servers, etc.) MUST NOT be committed or pushed** – they are external dependencies tracked at specific versions
- **Only project-owned submodules (LLMsVerifier, formatters) may be updated** – and only when necessary for project functionality
- **Submodules under `cli_agents/` are third-party code** – never commit changes to these repositories
- **Submodules under `MCP/` are third-party MCP servers** – treat as read-only dependencies
- **When updating submodule references**, use `git submodule update --remote` to pull upstream changes, then commit the updated reference in the main repository

**Failure to comply with ANY of these standards will result in component rejection. These rules ensure HelixAgent maintains enterprise-grade reliability, security, and performance.**

## Code Style Guidelines

### General Principles
- Follow standard Go conventions as described in [Effective Go](https://go.dev/doc/effective_go).
- Write clear, readable, and maintainable code.
- Keep functions small and focused (single responsibility).
- Use comments to explain "why" not "what".
- Avoid premature optimization; profile first.

### Formatting
- Use `gofmt` (or `go fmt`) to format code. The project's Makefile provides `make fmt`.
- Use `goimports` to organize imports (standard library, third‑party, local).
- Imports should be grouped: standard library, external dependencies, internal packages (separated by a blank line).
- Line length: aim for ≤ 100 characters, but readability is more important.

### Naming Conventions
- Use `camelCase` for local variables and private functions.
- Use `PascalCase` for exported functions, types, constants, and fields.
- Use `UPPER_SNAKE_CASE` for exported constants.
- Acronyms should be all caps (e.g., `HTTP`, `URL`, `ID`).
- Use short, descriptive names; avoid abbreviations unless widely understood.
- Receiver names: use one or two letters (e.g., `s` for a service, `c` for a client).

### Error Handling
- Always check errors; do not ignore them.
- Use `if err != nil { return err }` pattern.
- Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- Define custom error types when you need to expose specific error information.
- Use `defer` for cleanup (closing files, releasing resources).

### Types and Interfaces
- Use `interface` to define behavior, not data.
- Prefer small, focused interfaces (e.g., `io.Reader`, `io.Writer`).
- Use `struct` tags for JSON, YAML, database mapping, etc.
- Avoid `any` (`interface{}`); use generics or specific types when possible.
- Use type aliases and embedded structs judiciously.

### Concurrency
- Use goroutines and channels for concurrent tasks.
- Always provide a way to cancel or timeout operations (use `context.Context`).
- Protect shared data with `sync.Mutex` or `sync.RWMutex`.
- Consider using `sync.WaitGroup` to wait for goroutines to finish.

### Testing
- Write table‑driven tests when appropriate.
- Use the `testify` assertion library (already a project dependency).
- Mock external dependencies using interfaces.
- Place test files in the same package as the code being tested (suffix `_test.go`).
- Use `testdata/` directories for fixture files.

## Git & Commit Guidelines

- **Branch naming**: Use prefixes: `feat/`, `fix/`, `chore/`, `docs/`, `refactor/`, `test/`, followed by a short description (e.g., `feat/add-user-auth`).
- **Commit messages**: Follow [Conventional Commits](https://www.conventionalcommits.org/):
  - Format: `<type>(<scope>): <description>`
  - Common types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`.
  - Example: `feat(llm): add ensemble voting strategy`
- Always run `make fmt vet lint test` before committing.

## Additional Resources

- `CLAUDE.md` – Detailed project overview and architecture.
- `Makefile` – Complete list of available commands.
- `go.mod` – Go module dependencies.
- `docs/` – Project documentation.

---

*This document is intended for AI agents working in the HelixAgent repository. Keep it up to date as the project evolves.*

<!-- BEGIN_CONSTITUTION -->
# Project Constitution

**Version:** 1.1.0 | **Updated:** 2026-02-17 12:00

Constitution with 24 rules (24 mandatory) across categories: Quality: 2, Safety: 1, Security: 1, Performance: 2, Containerization: 2, Configuration: 1, Testing: 4, Documentation: 2, Principles: 2, Stability: 1, Observability: 1, GitOps: 1, CI/CD: 1, Architecture: 1, Networking: 1, Resource Management: 1

## Mandatory Principles

**All development MUST adhere to these non-negotiable principles:**

### Architecture

**Comprehensive Decoupling** (Priority: 1)
- Identify all parts and functionalities that can be extracted as separate modules (libraries) and reused in various projects. Perform additional work to make each module fully decoupled and independent. Each module must be a separate project with its own CLAUDE.md, AGENTS.md, README.md, docs/, tests, and challenges.

### Testing

**100% Test Coverage** (Priority: 1)
- Every component MUST have 100% test coverage across ALL test types: unit, integration, E2E, security, stress, chaos, automation, and benchmark tests. No false positives. Use real data and live services (mocks only in unit tests).

**Comprehensive Challenges** (Priority: 1)
- Every component MUST have Challenge scripts validating real-life use cases. No false success - validate actual behavior, not return codes.

**Stress and Integration Tests** (Priority: 2)
- Introduce comprehensive stress and integration tests validating that the system is responsive and not possible to overload or break.

**Infrastructure Before Tests** (Priority: 1)
- ALL infrastructure containers (PostgreSQL, Redis, Mock LLM) MUST be running before executing tests or challenges. Use `make test-infra-start` or `make test-infra-direct-start` (Podman fallback with `--userns=host`). Tests and challenges that require infrastructure WILL FAIL without running containers.

**Comprehensive Verification** (Priority: 1)
- Every fix MUST be verified from all angles: runtime testing (actual HTTP requests), compile verification, code structure checks, npm/dependency existence checks, backward compatibility, and no false positives in tests or challenges. Grep-only validation is NEVER sufficient.

### Documentation

**Complete Documentation** (Priority: 1)
- Every module and feature MUST have complete documentation: README.md, CLAUDE.md, AGENTS.md, user guides, step-by-step manuals, video courses, diagrams, SQL definitions, and website content. No component can remain undocumented.

**Documentation Synchronization** (Priority: 1)
- Anything added to Constitution MUST be present in AGENTS.md and CLAUDE.md, and vice versa. Keep all three synchronized.

### Quality

**No Broken Components** (Priority: 1)
- No module, application, library, or test can remain broken, disabled, or incomplete. Everything must be fully functional and operational.

**No Dead Code** (Priority: 1)
- Identify and remove all 'dead code' - features or functionalities left unconnected with the system. Perform comprehensive research and cleanup.

### Safety

**Memory Safety** (Priority: 1)
- Perform comprehensive research for memory leaks, deadlocks, and race conditions. Apply safety fixes and improvements to prevent these issues.

### Security

**Security Scanning** (Priority: 1)
- Execute Snyk and SonarQube scanning. Analyze findings in depth and resolve everything. Ensure scanning infrastructure is accessible via containerization (Docker/Podman).

### Performance

**Monitoring and Metrics** (Priority: 2)
- Create tests that run and perform monitoring and metrics collection. Use collected data for proper optimizations.

**Lazy Loading and Non-Blocking** (Priority: 2)
- Implement lazy loading and lazy initialization wherever possible. Introduce semaphore mechanisms and non-blocking mechanisms to ensure flawless responsiveness.

### Principles

**Software Principles** (Priority: 2)
- Apply all software principles: KISS, DRY, SOLID, YAGNI, etc. Ensure code is clean, maintainable, and follows best practices.

**Design Patterns** (Priority: 2)
- Use appropriate design patterns: Proxy, Facade, Factory, Abstract Factory, Observer, Mediator, Strategy, etc. Apply patterns where they add value.

### Stability

**Rock-Solid Changes** (Priority: 1)
- All changes must be safe, non-error-prone, and MUST NOT BREAK any existing working functionality. Ensure backward compatibility unless explicitly breaking.

### Containerization

**Full Containerization** (Priority: 2)
- All services MUST run in containers (Docker/Podman/K8s). Support local default execution AND remote configuration. Services must auto-boot before HelixAgent is ready.

**Container-Based Builds** (Priority: 1)
- ALL release builds MUST be performed inside Docker/Podman containers for reproducibility. Use `make release` / `make release-all`. Version info injected via `-ldflags -X`. No release binaries should be built directly on the host unless container build is unavailable.

### Configuration

**Unified Configuration** (Priority: 1)
- **CLI agent configs MUST ONLY be generated using the HelixAgent binary** (`./bin/helixagent --generate-agent-config=<agent>` or `go run ./cmd/helixagent --generate-agent-config=<agent>`).
- **NEVER create, write, or modify CLI agent config files manually or via scripts.** The HelixAgent binary is the sole authority for config generation.
- Config generation uses LLMsVerifier's unified generator (`pkg/cliagents/`). No third-party scripts or manual edits.
- This ensures schema compliance, API key injection, MCP endpoint consistency, and validation for all 48 supported CLI agents.

**Non-Interactive Execution** (Priority: 1)
- **ALL commands MUST be fully non-interactive and automatable via command pipelines.**
- **NEVER prompt for passwords, passphrases, or any user input interactively.**
- SSH connections MUST use key-based authentication with SSH agent (`ssh-add`) or password provided via environment variables/sshpass.
- Container distribution to remote hosts MUST be fully automated through the Containers module's SSH executor with pre-configured credentials.
- All secrets (API keys, passwords, SSH keys) MUST be provided via environment variables or `.env` files, never via interactive prompts.

### Networking

**HTTP/3 (QUIC) with Brotli Compression** (Priority: 1)
- ALL HTTP communication MUST use HTTP/3 (QUIC) as primary transport with Brotli compression. HTTP/2 only as fallback. Compression priority: Brotli → gzip. All HTTP clients and servers MUST prefer HTTP/3. Use `quic-go/quic-go` for transport and `andybalholm/brotli` for compression.

### Resource Management

**Test and Challenge Resource Limits** (Priority: 1)
- ALL test and challenge execution MUST be strictly limited to 30-40% of host system resources. Use GOMAXPROCS=2, nice -n 19, ionice -c 3, and -p 1 for go test. Container limits required. Host machine runs mission-critical processes; exceeding limits has caused system crashes.

### Observability

**Health and Monitoring** (Priority: 2)
- Every service MUST expose health endpoints. Circuit breakers for all external dependencies. Prometheus/OpenTelemetry integration.

### GitOps

**GitSpec Compliance** (Priority: 2)
- Follow GitSpec constitution and all constraints from AGENTS.md and CLAUDE.md.

### CI/CD

**Manual CI/CD Only** (Priority: 1)
- NO GitHub Actions enabled. All CI/CD workflows and pipelines must be executed manually only.

---

*This Constitution is automatically synchronized with AGENTS.md, CLAUDE.md, and CONSTITUTION.json.*

<!-- END_CONSTITUTION -->


