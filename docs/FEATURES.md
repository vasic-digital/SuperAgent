# HelixAgent Comprehensive Feature Reference

This document provides a complete reference of all features, providers, protocols, and capabilities supported by HelixAgent.

## Table of Contents

- [1. LLM Providers](#1-llm-providers)
  - [1.1 Premium Providers (Tier 1)](#11-premium-providers-tier-1)
  - [1.2 High-Quality Specialized (Tier 2)](#12-high-quality-specialized-tier-2)
  - [1.3 Fast Inference (Tier 3)](#13-fast-inference-tier-3)
  - [1.4 Alternative Providers (Tier 4)](#14-alternative-providers-tier-4)
  - [1.5 Aggregators & Local (Tier 5-7)](#15-aggregators--local-tier-5-7)
- [2. Embedding Providers](#2-embedding-providers)
  - [2.1 Core Providers](#21-core-providers)
  - [2.2 Extended Providers](#22-extended-providers)
- [3. Protocol Implementations](#3-protocol-implementations)
  - [3.1 MCP (Model Context Protocol)](#31-mcp-model-context-protocol)
  - [3.2 LSP (Language Server Protocol)](#32-lsp-language-server-protocol)
  - [3.3 ACP (Agent Communication Protocol)](#33-acp-agent-communication-protocol)
- [4. Vector Databases](#4-vector-databases)
- [5. Power Features](#5-power-features)
  - [5.1 AI Debate System](#51-ai-debate-system)
  - [5.2 RAG System](#52-rag-system)
  - [5.3 Memory Management](#53-memory-management)
  - [5.4 Semantic Routing](#54-semantic-routing)
  - [5.5 Agentic Workflows](#55-agentic-workflows)
  - [5.6 Security Framework](#56-security-framework)
  - [5.7 Structured Output](#57-structured-output)
  - [5.8 LLM Testing Framework](#58-llm-testing-framework)
  - [5.9 Self-Improvement (RLAIF)](#59-self-improvement-rlaif)
  - [5.10 LLMOps](#510-llmops)
  - [5.11 Benchmark Runner](#511-benchmark-runner)
  - [5.12 Optimization Pipeline](#512-optimization-pipeline)
  - [5.13 Observability](#513-observability)
  - [5.14 Background Tasks](#514-background-tasks)
  - [5.15 Notifications](#515-notifications)
  - [5.16 Plugin System](#516-plugin-system)
  - [5.17 Skills Registry](#517-skills-registry)
  - [5.18 Tools Registry](#518-tools-registry)
  - [5.19 Agents Registry](#519-agents-registry)
- [6. Summary Statistics](#6-summary-statistics)

---

## 1. LLM Providers

HelixAgent supports **21 LLM providers** with automatic discovery and dynamic selection based on LLMsVerifier scores.

### 1.1 Premium Providers (Tier 1)

| Provider | Default Model | Capabilities | Authentication | Priority |
|----------|---------------|--------------|----------------|----------|
| **Claude (Anthropic)** | `claude-sonnet-4-5-20250929` | Streaming, Tools, System prompts | API Key / OAuth2 | 1 |
| **OpenAI** | `gpt-4o` | Streaming, Tools, Response formatting | API Key | 1 |
| **Google Gemini** | `gemini-2.0-flash` | Streaming, Tools, Safety settings, Images | API Key | 2 |

### 1.2 High-Quality Specialized (Tier 2)

| Provider | Default Model | Capabilities | Authentication | Priority |
|----------|---------------|--------------|----------------|----------|
| **DeepSeek** | `deepseek-chat` | Streaming, Tools | API Key | 3 |
| **Mistral** | `mistral-large-latest` | Streaming, Tools, Safe prompt | API Key | 3 |
| **xAI (Grok)** | `grok-3-beta` | Streaming, Tools, Regional support (US/EU) | API Key | 3 |
| **Qwen (Alibaba)** | `qwen-max` | Streaming, Tools | API Key / OAuth2 | 4 |
| **Cohere** | `command-r-plus` | Streaming, Tools, Citations, RAG | API Key | 4 |
| **Perplexity** | `llama-3.1-sonar-large-128k-online` | Streaming, Online search | API Key | 4 |
| **AI21 Labs** | `jamba-1.5-large` | Streaming, Tools | API Key | 5 |

### 1.3 Fast Inference (Tier 3)

| Provider | Default Model | Capabilities | Authentication | Priority |
|----------|---------------|--------------|----------------|----------|
| **Groq** | `llama-3.3-70b-versatile` | Streaming, Tools, Audio transcription | API Key | 5 |
| **Cerebras** | `llama-3.3-70b` | Streaming, Fast inference | API Key | 5 |

### 1.4 Alternative Providers (Tier 4)

| Provider | Default Model | Capabilities | Authentication | Priority |
|----------|---------------|--------------|----------------|----------|
| **Fireworks AI** | `llama-v3p1-70b-instruct` | Streaming, Tools | API Key | 6 |
| **Together AI** | `Llama-3.3-70B-Instruct-Turbo` | Streaming, Tools | API Key | 6 |
| **Replicate** | `meta/llama-2-70b-chat` | Async prediction, Webhooks | API Key | 7 |
| **Hugging Face** | `Meta-Llama-3-8B-Instruct` | Standard/Pro modes, Cache control | API Key | 8 |

### 1.5 Aggregators & Local (Tier 5-7)

| Provider | Default Model | Capabilities | Authentication | Priority |
|----------|---------------|--------------|----------------|----------|
| **OpenRouter** | `anthropic/claude-3.5-sonnet` | 150+ models, Streaming, Tools | API Key | 10 |
| **Zen (OpenCode)** | `grok-code` (free) | Anonymous access, Streaming, Tools | Optional API Key | 4 |
| **Ollama** | `llama3.2` | Local execution, Streaming | None (local) | 20 |

**Note**: Ollama is DEPRECATED (score: 5.0) - only used as last resort fallback.

### Provider Authentication Methods

| Method | Providers | Storage |
|--------|-----------|---------|
| **API Key** | All except OAuth providers | Environment variables |
| **OAuth2** | Claude, Qwen | CLI credential files |
| **Anonymous** | Zen (free models) | Device-ID header |
| **None** | Ollama | Local only |

---

## 2. Embedding Providers

HelixAgent supports **13 embedding providers** with 40+ models.

### 2.1 Core Providers

| Provider | Models | Dimensions | Authentication |
|----------|--------|------------|----------------|
| **OpenAI** | `text-embedding-3-small`, `text-embedding-3-large`, `text-embedding-ada-002` | 1536, 3072 | API Key |
| **Ollama** | `nomic-embed-text`, `mxbai-embed-large`, `all-minilm` | 384, 768, 1024 | None (local) |
| **BGE-M3** | `BAAI/bge-m3` | 1024 | HuggingFace API Key |
| **Nomic** | `nomic-ai/nomic-embed-text-v1.5` | 768 | HuggingFace API Key |
| **CodeBERT** | `microsoft/codebert-base` | 768 | HuggingFace API Key |
| **Qwen3** | `Qwen/Qwen3-Embedding-0.6B` | 768 | HuggingFace API Key |
| **GTE** | `thenlper/gte-large` | 1024 | HuggingFace API Key |
| **E5** | `intfloat/e5-large-v2` | 1024 | HuggingFace API Key |

### 2.2 Extended Providers

| Provider | Models | Dimensions | Authentication |
|----------|--------|------------|----------------|
| **Cohere** | `embed-english-v3.0`, `embed-multilingual-v3.0`, 4 more | 384-4096 | API Key |
| **Voyage AI** | `voyage-3`, `voyage-code-3`, `voyage-finance-2`, 5 more | 512-1536 | API Key |
| **Jina AI** | `jina-embeddings-v3`, `jina-clip-v1`, `jina-colbert-v2`, 6 more | 128-1024 | API Key |
| **Google Vertex AI** | `text-embedding-005`, `text-multilingual-embedding-002`, 3 more | 768 | Service Account |
| **AWS Bedrock** | `amazon.titan-embed-text-v1/v2`, `cohere.embed-english-v3` | 1024-1536 | AWS SigV4 |

---

## 3. Protocol Implementations

### 3.1 MCP (Model Context Protocol)

**Total: 35 implementations (19 adapters + 16 servers)**

#### MCP Adapters (External Service Integrations)

| Category | Adapters | Key Tools |
|----------|----------|-----------|
| **Cloud Storage** | AWS S3, Google Drive | `s3_list_buckets`, `s3_get/put_object`, `gdrive_list/get/create_file` |
| **Project Management** | Jira, Linear, Asana | `jira_get/create_issue`, `linear_create_issue`, `asana_create_task` |
| **Version Control** | GitLab | `gitlab_list_projects`, `gitlab_create_merge_request` |
| **Communication** | Slack | `slack_post_message`, `slack_list_channels` |
| **Design Tools** | Figma, Miro, SVGMaker | `figma_get_file`, `miro_list_boards`, `svgmaker_create_svg` |
| **Infrastructure** | Docker, Kubernetes | `docker_list_containers`, `k8s_list_pods`, `k8s_list_deployments` |
| **Search** | Brave Search | `brave_web_search`, `brave_image_search`, `brave_news_search` |
| **Analytics** | Datadog, Sentry | `datadog_get_metrics`, `sentry_list_issues` |
| **Database** | MongoDB | `mongodb_query_collection`, `mongodb_insert_document` |
| **Knowledge** | Notion | `notion_query_database`, `notion_create_page` |
| **Automation** | Puppeteer | `puppeteer_take_screenshot`, `puppeteer_extract_text` |
| **AI Generation** | Stable Diffusion | `stable_diffusion_generate_image` |

#### MCP Servers (Backend Integrations)

| Category | Servers |
|----------|---------|
| **Vector Stores** | Chroma, Qdrant, Weaviate |
| **Databases** | PostgreSQL, SQLite, Redis |
| **Development** | Git, GitHub, Filesystem |
| **Content** | Fetch, Memory, Replicate |
| **Design** | Figma, Miro, SVGMaker, StableDiffusion |

### 3.2 LSP (Language Server Protocol)

**10 supported language servers**

| Language | Server | Priority | Full LSP Support |
|----------|--------|----------|------------------|
| **Go** | `gopls` | 100 | Yes |
| **Rust** | `rust-analyzer` | 100 | Yes |
| **TypeScript/JS** | `ts-server` | 95 | Yes |
| **Python** | `pylsp` | 90 | Yes |
| **Python** | `pyright` | 85 | Partial |
| **C/C++** | `clangd` | 90 | Yes |
| **Java** | `jdtls` | 80 | Core |
| **PHP** | `intelephense` | 80 | Yes |
| **Ruby** | `solargraph` | 75 | Partial |
| **Lua** | `lua-language-server` | 70 | Partial |

**LSP Capabilities**: Completion, Hover, Definition, References, Diagnostics, Rename, Code Actions, Formatting, Signature Help

### 3.3 ACP (Agent Communication Protocol)

| Component | Purpose |
|-----------|---------|
| **ACPManager** | Server discovery, capability enumeration, action execution |
| **ACPClient** | HTTP/WebSocket transport, JSON-RPC 2.0, retry logic |

**Features**: Multi-transport, exponential backoff, server synchronization, diagnostics

---

## 4. Vector Databases

| Database | Location | Capabilities |
|----------|----------|--------------|
| **Qdrant** | `internal/vectordb/qdrant/` | Dense/sparse vectors, filtering, namespaces |
| **Milvus** | `internal/vectordb/milvus/` | Scalable vector storage, batch operations |
| **Pinecone** | `internal/vectordb/pinecone/` | Cloud-native, metadata filtering |
| **PgVector** | `internal/vectordb/pgvector/` | PostgreSQL native vectors |

---

## 5. Power Features

### 5.1 AI Debate System

**Location**: `internal/debate/`, `internal/services/debate_*`

| Aspect | Description |
|--------|-------------|
| **Purpose** | Multi-round debate between LLM providers with consensus voting |
| **Participants** | 15 LLMs (5 positions x 3 per position) |
| **Topologies** | Mesh, Star, Chain |
| **Phases** | Proposal -> Critique -> Review -> Synthesis |
| **Learning** | Cross-debate lesson extraction and application |
| **Activation** | `POST /v1/debates` |

### 5.2 RAG System

**Location**: `internal/rag/`

| Component | Purpose |
|-----------|---------|
| **Pipeline** | Orchestrates retrieval workflow |
| **Hybrid Search** | Dense + sparse retrieval fusion |
| **HyDE** | Hypothetical document embeddings for query expansion |
| **Reranker** | Multi-stage relevance scoring |
| **Qdrant Integration** | Vector storage and retrieval |

### 5.3 Memory Management

**Location**: `internal/memory/`

| Feature | Description |
|---------|-------------|
| **Memory Types** | Episodic, semantic, procedural, working |
| **Entity Graph** | Entity and relationship storage |
| **Decay** | Automatic memory decay over time |
| **Session Scope** | Cross-session recall |

### 5.4 Semantic Routing

**Location**: `internal/routing/semantic/`

| Feature | Description |
|---------|-------------|
| **Purpose** | Embedding-based request routing |
| **Coverage** | 96.2% test coverage |
| **Matching** | Threshold-based with top-K retrieval |

### 5.5 Agentic Workflows

**Location**: `internal/agentic/`

| Feature | Description |
|---------|-------------|
| **Style** | LangGraph-style DAG execution |
| **Node Types** | Agent, Tool, Condition, Parallel, Human-in-loop, Subgraph |
| **Checkpointing** | Fault-tolerant state saving |
| **Coverage** | 96.5% test coverage |

### 5.6 Security Framework

**Location**: `internal/security/`

| Component | Purpose |
|-----------|---------|
| **Red Team** | 40+ attack patterns |
| **Guardrails** | Output safety constraints |
| **PII Detection** | Sensitive data identification |
| **Secure Fix Agent** | AI-powered vulnerability remediation |
| **Audit Logging** | Security event tracking |
| **OWASP Coverage** | Top 10 vulnerabilities |

### 5.7 Structured Output

**Location**: `internal/structured/`

| Schema Type | Description |
|-------------|-------------|
| **JSON Schema** | Strict JSON output enforcement |
| **Regex** | Pattern-based constraints |
| **Grammar** | Context-free grammar validation |
| **Enum** | Enumeration constraints |

### 5.8 LLM Testing Framework

**Location**: `internal/testing/llm/`

| Feature | Description |
|---------|-------------|
| **Style** | DeepEval-style evaluation |
| **Metrics** | Relevance, faithfulness, hallucination |
| **RAGAS** | RAG evaluation metrics |
| **Coverage** | 96.2% test coverage |

### 5.9 Self-Improvement (RLAIF)

**Location**: `internal/selfimprove/`

| Component | Purpose |
|-----------|---------|
| **AI Reward Model** | LLM-as-judge scoring |
| **Feedback Collector** | Human/AI/debate feedback |
| **Policy Optimizer** | Policy update generation |
| **Constitutional AI** | Principle enforcement |

### 5.10 LLMOps

**Location**: `internal/llmops/`

| Feature | Description |
|---------|-------------|
| **Prompt Registry** | Semantic versioning for prompts |
| **A/B Testing** | Statistical experiment framework |
| **Continuous Eval** | Automated quality monitoring |
| **Alerting** | Regression detection |

### 5.11 Benchmark Runner

**Location**: `internal/benchmark/`

| Benchmark | Type |
|-----------|------|
| **SWE-Bench** | Software engineering |
| **HumanEval** | Code generation |
| **MBPP** | Python programming |
| **MMLU** | Multi-task knowledge |
| **GSM8K** | Math reasoning |
| **MATH** | Advanced math |
| **HellaSwag** | Commonsense reasoning |

### 5.12 Optimization Pipeline

**Location**: `internal/optimization/`

| Component | Purpose |
|-----------|---------|
| **GPTCache** | Semantic caching (90%+ latency reduction) |
| **Outlines** | Structured output constraints |
| **Streaming** | Multi-level buffering and rate limiting |
| **LangChain** | Chain optimization |
| **LlamaIndex** | RAG optimization |
| **SGLang** | Structured generation |
| **Guidance** | Template optimization |
| **LMQL** | Query language optimization |

### 5.13 Observability

**Location**: `internal/observability/`

| Exporter | Type |
|----------|------|
| **Jaeger** | Distributed tracing |
| **Zipkin** | Trace collection |
| **Langfuse** | LLM-specific observability |
| **Prometheus** | Metrics export |

### 5.14 Background Tasks

**Location**: `internal/background/`

| Feature | Description |
|---------|-------------|
| **Task Queue** | PostgreSQL/In-memory queueing |
| **Worker Pool** | Auto-scaling (2-10 workers) |
| **Resource Monitor** | CPU, memory, disk tracking |
| **Stuck Detector** | Timeout and progress detection |
| **States** | pending, queued, running, completed, failed, stuck, cancelled |

### 5.15 Notifications

**Location**: `internal/notifications/`

| Channel | Description |
|---------|-------------|
| **SSE** | Server-Sent Events |
| **WebSocket** | Bidirectional real-time |
| **Webhooks** | HTTP callbacks with HMAC |
| **Polling** | Event storage for polling clients |

### 5.16 Plugin System

**Location**: `internal/plugins/`

| Feature | Description |
|---------|-------------|
| **Discovery** | Automatic plugin scanning |
| **Hot Reload** | File change detection |
| **Dependencies** | Dependency resolution |
| **Security** | Sandboxing and permissions |
| **Versioning** | Semantic versioning |
| **Health** | Plugin health monitoring |

### 5.17 Skills Registry

**Location**: `internal/skills/`

| Feature | Description |
|---------|-------------|
| **Categories** | code, debug, search, git, deploy, docs, test, review |
| **Matching** | Trigger phrase, fuzzy, semantic |
| **Tracking** | Usage analytics |
| **Hot Reload** | YAML configuration reload |

### 5.18 Tools Registry

**Location**: `internal/tools/`

**21 Tools**: Bash, Read, Write, Edit, Glob, Grep, WebFetch, WebSearch, Git, Task, TodoWrite, AskUserQuestion, EnterPlanMode, ExitPlanMode, Skill, NotebookEdit, KillShell, TaskOutput, and more.

### 5.19 Agents Registry

**Location**: `internal/agents/`

**18 Agents**: OpenCode, Crush, HelixCode, Kiro, Aider, ClaudeCode, Cline, CodenameGoose, DeepSeekCLI, Forge, GeminiCLI, GPTEngineer, KiloCode, MistralCode, OllamaCode, Plandex, QwenCode, AmazonQ

---

## 6. Summary Statistics

| Category | Count |
|----------|-------|
| **LLM Providers** | 21 |
| **Embedding Providers** | 13 |
| **MCP Implementations** | 35 (19 adapters + 16 servers) |
| **LSP Language Servers** | 10 |
| **ACP Components** | 2 |
| **Vector Databases** | 4 |
| **Tools** | 21 |
| **CLI Agents** | 18 |
| **Power Features** | 24+ major systems |
| **Security Attack Patterns** | 40+ |
| **Debate Participants** | 15 LLMs |
| **Benchmarks Supported** | 7 |

---

## Quick Reference: API Endpoints

| Endpoint | Protocol | Description |
|----------|----------|-------------|
| `/v1/chat/completions` | OpenAI | Chat completions (ensemble) |
| `/v1/completions` | OpenAI | Text completions |
| `/v1/embeddings` | OpenAI | Vector embeddings |
| `/v1/debates` | HelixAgent | AI debate system |
| `/v1/mcp` | MCP | Model Context Protocol |
| `/v1/lsp` | LSP | Language Server Protocol |
| `/v1/lsp/ws` | LSP | LSP WebSocket |
| `/v1/acp` | ACP | Agent Communication Protocol |
| `/v1/rag/*` | HelixAgent | RAG operations |
| `/v1/cognee` | HelixAgent | Knowledge graph |
| `/v1/vision` | HelixAgent | Image analysis |
| `/v1/tasks` | HelixAgent | Background tasks |
| `/v1/monitoring/*` | HelixAgent | Monitoring endpoints |

---

*Last updated: 2026-01-22*
