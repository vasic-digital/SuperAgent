# Changelog

All notable changes to HelixAgent are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Benchmark test coverage for 40/40 LLM providers, 7 debate packages, and 17 core packages (556+ benchmarks)
- Penetration testing expansion: SSRF prevention, API key leakage, rate limit bypass, provider security (37 test functions across 7 files)
- Test type expansion for 23 extracted modules (integration, E2E, security, stress, benchmark)
- Module-specific challenge scripts for all 23 extracted modules (EventBus, Auth, Cache, Concurrency, Embeddings, Formatters, Storage, Streaming, Observability, Optimization, Plugins, Database, Messaging, Security, VectorDB, Memory, RAG, MCP, Agentic, LLMOps, SelfImprove, Planning, Benchmark)
- Architecture documentation (ARCHITECTURE.md) for Agentic, LLMOps, SelfImprove, Planning, Benchmark modules
- LICENSE file (MIT)

### Fixed
- Race conditions in plugins MetricsCollector and DependencyResolver
- Context propagation in ensemble execution, auth middleware, gRPC streaming, model metadata refresh, and tool result processing
- Response writer race in model_metadata.go RefreshModels goroutine
- Crash bugs in comprehensive debate system (nil pointer, timeout, map mutation)

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
