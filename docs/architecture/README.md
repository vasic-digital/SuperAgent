# Architecture Documentation

This directory contains comprehensive architecture documentation for HelixAgent, covering system design, service architecture, protocol support, and architectural decision records.

## Overview

HelixAgent is an AI-powered ensemble LLM service that combines responses from multiple language models using intelligent aggregation strategies. The architecture supports 10 LLM providers, dynamic provider selection, and multiple protocol integrations.

## Core Architecture Documents

### System Architecture

| Document | Description |
|----------|-------------|
| [architecture.md](architecture.md) | Core system architecture overview with high-level diagrams, entry points, and layer descriptions |
| [SUPERAGENT_COMPREHENSIVE_ARCHITECTURE.md](SUPERAGENT_COMPREHENSIVE_ARCHITECTURE.md) | Comprehensive architecture explaining HelixAgent as a Virtual LLM Provider with AI Debate Ensemble |
| [SERVICE_ARCHITECTURE.md](SERVICE_ARCHITECTURE.md) | Unified service management architecture with BootManager, HealthChecker, and service dependency graphs |

### Protocol Support

| Document | Description |
|----------|-------------|
| [PROTOCOL_SUPPORT_DOCUMENTATION.md](PROTOCOL_SUPPORT_DOCUMENTATION.md) | Unified protocol support for MCP, LSP, ACP, and Embeddings with API endpoints |
| [README_PROTOCOL_ENHANCED.md](README_PROTOCOL_ENHANCED.md) | Enhanced protocol capabilities and integration details |
| [toon-protocol.md](toon-protocol.md) | Token-Optimized Object Notation (TOON) - custom serialization for AI communication |

### Component Architecture

| Document | Description |
|----------|-------------|
| [AGENTS.md](AGENTS.md) | CLI agent architecture, 48+ agents, build commands, code style guidelines |
| [CIRCUIT_BREAKER.md](CIRCUIT_BREAKER.md) | Circuit breaker pattern implementation for fault tolerance |
| [FORMATTERS_ARCHITECTURE.md](FORMATTERS_ARCHITECTURE.md) | Code formatters integration architecture (118+ formatters, 19 languages) |

### API Architecture

| Document | Description |
|----------|-------------|
| [graphql-api.md](graphql-api.md) | GraphQL API schema with queries, mutations, and subscriptions |

### Messaging & Integration

| Document | Description |
|----------|-------------|
| [messaging-architecture.md](messaging-architecture.md) | Hybrid messaging with RabbitMQ (task queuing) and Kafka (event streaming) |
| [rabbitmq-integration.md](rabbitmq-integration.md) | RabbitMQ configuration and integration details |
| [kafka-integration.md](kafka-integration.md) | Apache Kafka event streaming integration |

## System Component Overview

```
+-------------------------------------------------------------------+
|                          HelixAgent                                 |
|                    LLM Orchestration Platform                       |
+-------------------------------------------------------------------+
|  API Layer: REST (Gin) | gRPC | WebSocket | GraphQL               |
+-------------------------------------------------------------------+
|  Middleware: Auth | Rate Limiting | CORS | Request Validation     |
+-------------------------------------------------------------------+
|                        Core Services                                |
|  +-------------+  +-------------+  +-------------+  +-----------+  |
|  |  Provider   |  |  Ensemble   |  |  Context    |  |  Memory   |  |
|  |  Registry   |  |  Service    |  |  Manager    |  |  Service  |  |
|  +-------------+  +-------------+  +-------------+  +-----------+  |
|  +-------------+  +-------------+  +-------------+  +-----------+  |
|  | MCP Manager |  | LSP Client  |  |    Tool     |  | Security  |  |
|  |             |  |             |  |  Registry   |  |  Sandbox  |  |
|  +-------------+  +-------------+  +-------------+  +-----------+  |
+-------------------------------------------------------------------+
|                       Provider Layer                                |
|  DeepSeek | Gemini | Claude | OpenRouter | Qwen | Mistral | ...   |
+-------------------------------------------------------------------+
|                     Infrastructure Layer                            |
|  PostgreSQL | Redis | Prometheus | Grafana | Docker/K8s           |
+-------------------------------------------------------------------+
```

## Extracted Modules Layer

HelixAgent's functionality is decomposed into **25 independent Go modules** integrated as git
submodules. All modules use `replace` directives in the root `go.mod` for local development. Bridge
adapters in `internal/adapters/<name>/adapter.go` connect internal types to each module.

Full catalog: [`docs/MODULES.md`](../MODULES.md)

| Phase | Modules |
|-------|---------|
| Foundation | EventBus, Concurrency, Observability, Auth, Storage, Streaming |
| Infrastructure | Security, VectorDB, Embeddings, Database, Cache |
| Services | Messaging, Formatters, MCP |
| Integration | RAG, Memory, Optimization, Plugins |
| **AI/ML (Phase 5)** | **Agentic, LLMOps, SelfImprove, Planning, Benchmark** |
| Pre-existing | Containers, Challenges |

### AI/ML Modules (Phase 5)

Five new modules added to support advanced AI/ML capabilities:

| Module | Go Path | Description |
|--------|---------|-------------|
| **Agentic** (`Agentic/`) | `digital.vasic.agentic` | Graph-based workflow orchestration for autonomous AI agents with planning, conditional branching, parallel execution, state management, and retry logic |
| **LLMOps** (`LLMOps/`) | `digital.vasic.llmops` | LLM operations framework: continuous evaluation pipelines, A/B experiment management with statistical significance testing, dataset management, and prompt versioning |
| **SelfImprove** (`SelfImprove/`) | `digital.vasic.selfimprove` | AI self-improvement via RLHF-style feedback collection, reward model training with dimension-weighted scoring, and optimizer that adjusts model parameters based on feedback |
| **Planning** (`Planning/`) | `digital.vasic.planning` | AI planning algorithms: HiPlan (hierarchical milestone-based decomposition), MCTS (Monte Carlo Tree Search for code action optimization), Tree of Thoughts (multi-path reasoning) |
| **Benchmark** (`Benchmark/`) | `digital.vasic.benchmark` | LLM benchmarking framework: industry-standard benchmarks (SWE-bench, HumanEval, MMLU, GSM8K, MATH, MBPP, LMSYS, HellaSwag), custom benchmarks, provider comparison, and leaderboard generation |

## Key Architectural Patterns

### AI Debate Ensemble

HelixAgent presents itself as a single LLM provider with one virtual model (AI Debate Ensemble) that internally:
- Uses 5 primary LLMs selected by LLMsVerifier scores
- Maintains 2-3 fallbacks per primary for resilience
- Employs confidence-weighted voting for response aggregation

### Provider Verification Pipeline

On startup:
1. Discover available providers
2. Verify in parallel (8-test pipeline)
3. Score based on 5 weighted components
4. Rank and select debate team
5. Start server with verified providers

### Circuit Breaker Pattern

Implements fault tolerance with three states:
- **Closed**: Normal operation
- **Open**: Failure detected, requests rejected
- **Half-Open**: Recovery testing with limited requests

### Service Boot Management

BootManager orchestrates infrastructure:
- Groups services by compose file
- Starts via `docker compose up -d`
- Health checks all enabled services
- Supports TCP and HTTP health check types

## Decision Records

Key architectural decisions documented in these files:

| Area | Decision | Rationale |
|------|----------|-----------|
| Provider Selection | Dynamic via LLMsVerifier | Ensures only verified, high-quality providers participate |
| Protocol Support | Unified Protocol Manager | Single interface for MCP, LSP, ACP, Embeddings |
| Messaging | Hybrid RabbitMQ + Kafka | Task queuing + event streaming for different use cases |
| Containerization | Docker/Podman agnostic | Support multiple container runtimes |
| API Design | OpenAI-compatible | Industry-standard interface for easy integration |

## Related Documentation

- [Database Schema](../database/README.md) - Data model and schema reference
- [MCP Systems](../mcp/README.md) - Model Context Protocol adapters
- [Monitoring](../monitoring/README.md) - Observability and health checking
- [Security](../security/README.md) - Security model and sandboxing

## Quick Links

- Main entry point: `cmd/helixagent/main.go`
- Core services: `internal/services/`
- LLM providers: `internal/llm/providers/`
- Protocol handlers: `internal/handlers/`
