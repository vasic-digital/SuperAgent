# Changelog

All notable changes to HelixAgent are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - 2026-03-25

### Fixed
- Duplicate GetAgentPool() method in debate/comprehensive/integration.go
- Skills routes registered inside per-request handler closure (moved to startup)
- Channel leaks in Gemini and Qwen ACP providers
- Context cancellation missing in query_optimizer cleanup loop
- 32 broken documentation links across docs/

### Added
- AgenticEnsemble: dual-mode unified LLM (reason + execute) with full tool integration
- IterativeToolExecutor: per-phase iterative tool loops with all 6 protocols
- ExecutionPlanner: SpecKit-based task decomposition with dependency graph
- AgentWorkerPool: background agent spawning with semaphore limiting
- VerificationDebate: post-execution verification with LLM-based quality checks
- VisionClient and HelixMemoryClient interfaces in ToolIntegration
- 3 new challenge scripts for agentic ensemble validation
- Dead code verification challenge
- Test coverage completeness challenge
- 15+ new test files for under-covered packages
- 6 fuzz test targets (JSON, schema, protocol, template, config)
- 3 precondition tests (database, redis, API health)
- goleak goroutine leak detection in 5 critical packages
- 6 stress tests (rate limiter, ensemble, debate, streaming, cache, db pool)
- sync.Once lazy service initialization for router handlers
- Exponential backoff for debate optimizer, HTTP concurrency limiter
- SSE connection caps per client IP, queue depth metrics
- 10 monitoring validation tests, 10 benchmark baselines
- 6 modules added to MODULES.md (41 total)
- 7 new user manuals, 6 new video courses
- SQL schema index and guide
- 11 new challenge scripts across all categories

### Removed
- internal/background/backup/ (stale package duplication, 364KB)
- 5 dead adapter packages (background, observability, events, http, helixqa)

### Changed
- Challenge orchestrator expanded from 64 to 493 scripts
- MODULES.md updated from 33 to 41 modules
- Website updated with new module features and changelog

## [1.2.0] - 2026-03-20

### Added
- Benchmark test coverage for 40/40 LLM providers, 7 debate packages, and 17 core packages (556+ benchmarks)
- Additional 98 benchmark functions across 20 internal packages (benchmark, challenges, concurrency, conversation, features, knowledge, llmops, models, modelsdev, observability, planning, profiling, selfimprove, tools, toon, debate/agents, debate/protocol, debate/reflexion, debate/testing, debate/tools)
- Penetration testing expansion: SSRF prevention, API key leakage, rate limit bypass, provider security (37 test functions across 7 files)
- Test type expansion for 23 extracted modules (integration, E2E, security, stress, benchmark)
- Module-specific challenge scripts for all 23 extracted modules (EventBus, Auth, Cache, Concurrency, Embeddings, Formatters, Storage, Streaming, Observability, Optimization, Plugins, Database, Messaging, Security, VectorDB, Memory, RAG, MCP, Agentic, LLMOps, SelfImprove, Planning, Benchmark)
- Architecture documentation (ARCHITECTURE.md) for Agentic, LLMOps, SelfImprove, Planning, Benchmark, HelixMemory, HelixSpecifier modules
- LICENSE file (MIT)

### Fixed
- Race conditions in plugins MetricsCollector and DependencyResolver
- Context propagation in ensemble execution, auth middleware, gRPC streaming, model metadata refresh, and tool result processing
- Response writer race in model_metadata.go RefreshModels goroutine
- Crash bugs in comprehensive debate system (nil pointer, timeout, map mutation)

### Refactored
- Comprehensive debate system PhaseOrchestrator: replaced all manual placeholder responses with proper agent.Process() calls across all 6 phases (Planning, Generation, Debate, Validation, Refactoring, Integration)
- Wired System to PhaseOrchestrator with pool and orchestrator fields, implemented real convergence checking (quality threshold, consensus level, early stopping)

### Security
- gosec scan completed: all 122 G704 (SSRF) findings are expected for multi-provider LLM service, 5 G117 (exported struct secrets) are informational

### Removed
- Dead services: ProtocolCacheManager, ProviderMetadataService, OptimizedRequestService
- Empty stale directories in debate, graphql, performance, and verifier packages

## [1.0.0] - 2026-03-01

### Added
- **Comprehensive Debate System**: 11-role multi-agent debate with dynamic prompts, real LLM calls, and fallback tracking
- **Provider Verification Reports**: Startup report generation with team headers and per-role model display
- **40 LLM Providers**: Claude, Chutes, DeepSeek, Gemini, Mistral, OpenRouter, Qwen, ZAI, Zen, Cerebras, Ollama, AI21, Anthropic, Cohere, Fireworks, Groq, HuggingFace, OpenAI, Perplexity, Replicate, Together, xAI, Nvidia, SambaNova, Novita, SiliconFlow, Upstage, Kimi, Cloudflare, Codestral, Hyperbolic, Kilo, Modal, Nia, NLPCloud, PublicAI, Sarvam, Vulavula, Zhipu, and generic OpenAI-compatible provider
- **Dynamic Model Discovery**: 3-tier discovery (Provider API, models.dev catalog, hardcoded fallback) with 1-hour TTL cache
- **27 Extracted Modules**: Containers, Challenges, EventBus, Concurrency, Observability, Auth, Storage, Streaming, Security, VectorDB, Embeddings, Database, Cache, Messaging, Formatters, MCP, RAG, Memory, Optimization, Plugins, Agentic, LLMOps, SelfImprove, Planning, Benchmark, HelixMemory, HelixSpecifier
- **Debate Orchestrator**: Multi-topology (mesh/star/chain/tree), 8-phase protocol, cross-debate learning
- **Debate Voting**: 6 methods (Weighted MiniMax, Majority, Borda Count, Condorcet, Plurality, Unanimous)
- **Debate Reflexion**: Episodic memory buffer, verbal reflection, retry-and-learn loop
- **Debate Adversarial Dynamics**: Red/Blue team attack-defend cycles
- **Debate Approval Gates**: Human-in-the-loop REST API (approve/reject)
- **Debate Persistence**: PostgreSQL tables for sessions, turns, code versions
- **Debate Provenance & Audit**: 14 event types, session summaries, JSON export
- **Debate Performance Optimizer**: Parallel LLM execution with semaphore, response caching, early termination
- **SpecKit Auto-Activation**: 7-phase development flow with work granularity detection
- **Constitution Management**: Auto-update on project changes with background watcher
- **48 CLI Agents**: 4 custom (OpenCode, Crush, KiloCode, HelixCode) + 44 generic with 15+ MCP servers each
- **32+ Code Formatters**: 11 native, 14 service, 7 built-in for 19 languages
- **45+ MCP Adapters**: 65+ containerized MCP servers (ports 9101-9999)
- **HelixMemory**: Unified cognitive memory engine (Mem0, Cognee, Letta, Graphiti fusion)
- **HelixSpecifier**: Spec-driven development fusion engine (SpecKit + Superpowers + GSD)
- **Security Framework**: Red team, guardrails, PII detection, content filtering
- **BigData Integration**: Infinite context, distributed memory, knowledge graph streaming
- **OpenAI-compatible API**: `/v1/chat/completions`, `/v1/models` endpoints
- **Protocol Endpoints**: MCP, ACP, LSP, Embeddings, Vision, Cognee
- **HTTP/3 (QUIC)**: Primary transport with Brotli compression, HTTP/2 fallback
- **Container-Based Release Builds**: 7 apps across 5 platforms with change detection
- **LLMsVerifier**: Provider accuracy verification with 5-component scoring
- **Subscription Detection**: 3-tier dynamic detection for provider access levels
- **Circuit Breaker**: Fault tolerance for all external provider dependencies
- **Monitoring**: Prometheus/OpenTelemetry integration, health endpoints, status API

### Infrastructure
- Docker/Podman container orchestration with remote distribution support
- PostgreSQL 15 (pgx/v5), Redis 7 for caching and sessions
- Centralized container management through Containers module
- Automated boot sequence with health checks and strict mode
- 35+ challenge scripts with 800+ validation tests

## [0.1.0] - 2025-12-15

### Added
- Initial project structure with Go 1.24+ module
- Core LLM provider interface and ensemble execution
- Gin-based HTTP server with JWT authentication
- Basic provider implementations (OpenAI, Claude, Gemini, DeepSeek)
- AI Debate configuration system with multi-LLM fallback chains
- Cognee integration for knowledge graph memory
- Toolkit subproject for reusable AI app components
- Docker deployment configuration
