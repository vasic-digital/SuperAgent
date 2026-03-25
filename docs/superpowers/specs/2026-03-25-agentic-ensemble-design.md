# AgenticEnsemble: Unified LLM with Autonomous Execution — Design Spec

**Date:** 2026-03-25
**Version:** 1.1.0
**Status:** Approved (v1.1 — corrected per spec review)
**Scope:** Make the AI ensemble behave as a single consistent LLM with full tool access and autonomous execution capability

---

## Table of Contents

1. [Goal](#goal)
2. [Architecture Overview](#architecture-overview)
3. [Dual Operating Modes](#dual-operating-modes)
4. [Tool-Augmented Debate (Reason Mode)](#tool-augmented-debate-reason-mode)
5. [Agentic Execution Loop (Execute Mode)](#agentic-execution-loop-execute-mode)
6. [Background Agent Worker Pool](#background-agent-worker-pool)
7. [Full Power Feature Integration](#full-power-feature-integration)
8. [New Files](#new-files)
9. [Modified Files](#modified-files)
10. [Testing](#testing)
11. [Challenges](#challenges)
12. [Documentation](#documentation)
13. [Error Handling & Safety](#error-handling--safety)
14. [Constraints](#constraints)

---

## Goal

Make the HelixAgent ensemble behave as a **single unified LLM** that can:
- Use all tooling (MCP, LSP, ACP, RAG, Embeddings, Vision) during reasoning
- Autonomously execute multi-step tasks via background worker agents
- Leverage HelixSpecifier for task decomposition and planning
- Integrate every power feature (HelixMemory 4-engine, Formatters, Security, BigData, Plugins, Reflexion, Provenance)
- Operate in dual mode: tool-augmented reasoning for queries, full autonomous execution for actionable requests

---

## Architecture Overview

The `AgenticEnsemble` is a new orchestration component that sits between the intent classifier and the existing debate/ensemble infrastructure. It wraps (not replaces) existing components.

```
User Request → ChatCompletions Handler
    → AgenticEnsemble.Process(ctx, request)
        → intentClassifier.Classify(request)
        → switch mode:
            case Reason:  → toolAugmentedDebate(ctx, request)
            case Execute: → agenticExecutionLoop(ctx, request)
        → return unified OpenAI-compatible response
```

### Core Struct

```go
type AgenticEnsemble struct {
    debateService       *DebateService                          // Existing: internal/services/debate_service.go
    specKit             *SpecKitOrchestrator                    // Existing: internal/services/speckit_orchestrator.go
    toolBridge          *tools.ToolIntegration                  // Existing: internal/debate/tools/tool_integration.go (extended)
    intentClassifier    *LLMIntentClassifier                    // Existing: internal/services/llm_intent_classifier.go
    workerPool          *AgentWorkerPool                        // NEW: internal/services/agent_worker_pool.go
    memory              *memoryadapter.StoreAdapter             // Existing: internal/adapters/memory/adapter.go
    helixMemory         helixmemory.UnifiedMemoryProvider       // Existing: HelixMemory module (Mem0+Cognee+Letta+Graphiti)
    visionAdapter       *VisionToolAdapter                      // NEW: wraps existing internal/handlers/vision_handler.go
    providerRegistry    *ProviderRegistry                       // Existing: internal/services/provider_registry.go
    eventBus            *eventbusadapter.Adapter                // Existing: internal/adapters/eventbus.go
    provenanceTracker   *ProvenanceTracker                      // Existing: internal/debate/audit/provenance_tracker.go
    reflexionLoop       *reflexion.ReflexionLoop                // Existing: internal/debate/reflexion/reflexion_loop.go
    guardrails          *security.StandardGuardrailPipeline     // Existing: internal/security/guardrails.go
    formatterRegistry   *formatters.Registry                    // Existing: internal/formatters/registry.go
    pluginSystem        *ProtocolPluginSystem                   // Existing: internal/services/protocol_plugin_system.go
    bigDataIntegration  *BigDataIntegration                     // Existing: internal/bigdata/integration.go
    config              AgenticEnsembleConfig
}
```

### Key Type Definitions (NEW)

```go
// AgenticTask is the structured task output from SpecKit decomposition.
// The SpecKit Tasks phase produces debate output which is parsed into these.
type AgenticTask struct {
    ID              string            // Unique task identifier
    Description     string            // What needs to be done
    Dependencies    []string          // IDs of tasks that must complete first
    ToolRequirements []string         // Required protocols: "mcp", "lsp", "rag", etc.
    Priority        int               // Higher = execute first among peers
    EstimatedSteps  int               // Hint for iteration budget
    Status          AgenticTaskStatus // pending/running/completed/failed
}

// AgenticTaskStatus tracks task lifecycle.
type AgenticTaskStatus int

const (
    TaskPending AgenticTaskStatus = iota
    TaskRunning
    TaskCompleted
    TaskFailed
)

// VisionClient interface for tool integration.
type VisionClient interface {
    AnalyzeImage(ctx context.Context, imageData []byte, prompt string) (*VisionAnalysis, error)
    AnalyzeURL(ctx context.Context, imageURL string, prompt string) (*VisionAnalysis, error)
}

// HelixMemoryClient interface for full 4-engine access during debate/execution.
type HelixMemoryClient interface {
    // Mem0 — fact storage
    StoreFact(ctx context.Context, fact string, metadata map[string]string) error
    RecallFacts(ctx context.Context, query string, limit int) ([]string, error)
    // Cognee — knowledge graphs
    BuildGraph(ctx context.Context, content string) error
    QueryGraph(ctx context.Context, query string) ([]GraphNode, error)
    // Letta — stateful agent runtime
    CreateAgentSession(ctx context.Context, agentID string) (string, error) // returns sessionID
    SendAgentMessage(ctx context.Context, sessionID string, message string) (string, error)
    // Graphiti — temporal graphs
    AddTemporalEdge(ctx context.Context, from, to, relation string, timestamp time.Time) error
    QueryTimeline(ctx context.Context, entity string, timeRange TimeRange) ([]TemporalEdge, error)
}
```

---

## Dual Operating Modes

Determined by the existing `LLMIntentClassifier`:

| Mode | Trigger | Behavior |
|------|---------|----------|
| **Reason** | Informational queries, analysis, explanations | Tool-augmented debate: each debate phase invokes tools iteratively, producing enriched response |
| **Execute** | Actionable requests (create, modify, fix, implement) | Full agentic loop: debate → SpecKit plan → background agents → verification → response |

The intent classifier already distinguishes between trivial, informational, and actionable intents. The AgenticEnsemble maps these to Reason or Execute mode.

---

## Tool-Augmented Debate (Reason Mode)

Each debate phase becomes an **iterative tool loop** where agents invoke tools and feed results into their reasoning before producing phase output.

### Enhanced Phase Flow

```
For each debate phase (Dehallucination → ... → Convergence):
    1. Agent receives context + prior phase results
    2. Agent produces initial response
    3. If response contains tool_calls:
        a. Execute tool calls via ToolIntegration
        b. Feed tool results back to the SAME agent
        c. Agent produces refined response
        d. Repeat up to MaxToolIterations (default: 5)
    4. Phase output = final refined response
    5. Pass to next phase
```

### IterativeToolExecutor

New component: `internal/services/iterative_tool_executor.go`

```go
type IterativeToolExecutor struct {
    toolBridge       *ToolIntegration
    maxIterations    int           // Default: 5
    iterationTimeout time.Duration // Default: 30s
}

func (e *IterativeToolExecutor) ExecuteWithTools(
    ctx context.Context,
    provider LLMProvider,
    messages []Message,
    availableTools []ToolSchema,
) (*LLMResponse, []ToolExecution, error)
```

Only providers with `SupportsTools: true` (queried dynamically via `provider.GetCapabilities().SupportsTools` at runtime) participate in tool-augmented phases. Currently 23+ providers support tools including Claude, DeepSeek, Gemini, OpenRouter, Qwen, ZAI, Mistral, Zen, Venice, GitHub Models, Together, Junie, Cohere, Groq, xAI, OpenAI, Chutes, Anthropic, AI21, Fireworks, SambaNova, and others. The list is **never hardcoded** — capability detection is dynamic.

---

## Agentic Execution Loop (Execute Mode)

7-stage pipeline integrating every power feature:

### Stage 1: UNDERSTAND
- Tool-augmented debate to understand the request
- RAG: search existing codebase knowledge
- LSP: get diagnostics, definitions, references
- Embeddings: semantic similarity search
- HelixMemory: recall prior decisions (Mem0), entity relationships (Cognee), temporal context (Graphiti)
- Vision: analyze attached images/screenshots/diagrams
- MCP: query external context
- Output: Rich context document

### Stage 2: PLAN (via HelixSpecifier)
- SpecKit 7-phase flow: Constitution → Specify → Clarify → Plan → Tasks → Analyze → Implement
- Each SpecKit phase runs a debate with tool access
- HelixMemory Cognee: build knowledge graph of the plan
- Output: Decomposed task list with dependencies

### Stage 3: ASSIGN
- AgentWorkerPool spawns background agents (goroutines)
- Each agent gets: task, tool access, provider assignment
- Independent tasks assigned to parallel agents
- Dependent tasks queued with dependency graph
- Semaphore limits concurrent agents (default: 5)

### Stage 4: EXECUTE (parallel background agents)
- Each agent runs its own iterative tool loop
- Progress published via EventBus (SSE/WebSocket)
- HelixMemory Letta: maintain stateful agent context
- HelixMemory Graphiti: track temporal relationships

### Stage 5: VERIFY
- Verification debate: all agents present results
- LSP: run diagnostics on changed code
- Formatters: auto-format code output (32+ formatters)
- Security: PII detection, guardrails check
- HelixMemory Mem0: store verified facts
- If verification fails → targeted re-execution

### Stage 6: SYNTHESIZE
- Combine agent outputs into coherent response
- Provenance tracking: full audit trail
- HelixMemory: persist learned patterns
- Reflexion: generate verbal reflection for cross-session learning

### Stage 7: RESPOND
- Format as OpenAI-compatible response
- Include tool execution summary in metadata
- Stream intermediate results via SSE
- Publish completion event via EventBus

---

## Background Agent Worker Pool

### AgentWorkerPool

```go
type AgentWorkerPool struct {
    semaphore         chan struct{}
    providers         *ProviderRegistry
    toolBridge        *ToolIntegration
    memoryAdapter     *HelixMemoryAdapter
    eventBus          EventPublisher
    provenanceTracker *ProvenanceTracker
    wg                sync.WaitGroup
    ctx               context.Context
    cancel            context.CancelFunc
}
```

### AgentWorker

```go
type AgentWorker struct {
    ID              string
    Task            AgenticTask               // Structured task from execution planner
    Provider        LLMProvider
    ToolExecutor    *IterativeToolExecutor
    LettaSessionID  string                    // Letta agent session ID for stateful context
    MaxIterations   int                       // Default: 20 (total across all internal tool loops)
    Results         chan AgentResult
}
```

**Iteration budget clarification:** The agent's `MaxIterations` (20) is the **total** number of LLM reasoning steps across the entire task. Each reasoning step may invoke tools (up to `MaxToolIterations=5` per step is NOT a separate limit — the 5 is per debate phase in Reason mode only). In Execute mode, an agent step is: LLM reasons → optionally calls tools → produces output or next step. 20 such steps per task.

### Lifecycle

```
DispatchTasks(tasks []SpecKitTask)
    → Build dependency graph
    → For each task with no unmet dependencies:
        → semaphore.Acquire()
        → wg.Add(1)
        → go worker.Execute(ctx)
            → Select tool-capable provider
            → Initialize Letta agent state
            → Iterative tool loop (up to MaxIterations)
            → Publish result via eventBus
            → provenanceTracker.Record()
            → semaphore.Release(); wg.Done()
        → On task completion, unblock dependents
```

### Provider Selection

- Prefer tool-capable providers for tool-requiring tasks
- Round-robin across verified providers for load distribution
- Respect circuit breakers (skip Open-state providers)
- Match task complexity to provider capability score

### Concurrency Safety

- ToolIntegration is thread-safe (MCP/LSP clients use sync.Mutex)
- Each agent gets its own Letta agent state (no shared mutable state)
- Results collected via buffered channels
- Dependency graph uses atomic state transitions
- WaitGroup + context.Cancel for graceful shutdown

---

## Full Power Feature Integration

| Feature | Understand | Plan | Execute | Verify | Integration Point |
|---------|:---:|:---:|:---:|:---:|-------|
| MCP (45+ adapters) | Query context | - | Call tools | - | `toolBridge.mcpClient` |
| LSP (Go/Rust/Python/TS) | Definitions, refs | - | Navigate code | Diagnostics | `toolBridge.lspClient` |
| ACP (Agent Communication) | - | Delegate | Agent coordination | - | `toolBridge.acpClient` |
| RAG (Qdrant/Pinecone/Milvus/pgvector) | Knowledge search | Patterns | Retrieve docs | - | `toolBridge.ragClient` |
| Embeddings (6 providers) | Semantic search | Similarity | Relevance | Compare | `toolBridge.embeddingClient` |
| Vision (4 providers) | Analyze images | Diagrams | UI analysis | Visual verify | `toolBridge.visionClient` |
| HelixMemory Mem0 | Recall facts | Prior decisions | Context | Store facts | `memory.StoreFact/RecallFacts` |
| HelixMemory Cognee | Knowledge graph | Build plan graph | - | - | `memory.BuildGraph/QueryGraph` |
| HelixMemory Letta | - | - | Agent runtime | - | `memory.GetAgentState/UpdateState` |
| HelixMemory Graphiti | Temporal context | - | Track changes | Timeline | `memory.AddTemporalEdge/QueryTimeline` |
| Formatters (32+, 19 langs) | - | - | Format code | Auto-format | `formatterRegistry.Format` |
| Security (guardrails, PII) | - | - | - | Scan outputs | `securityGuardrails.Check` |
| Debate (8-phase, 5 positions) | Reasoning | Phase debates | Per-task debate | Verification | `debateService.ConductDebate` |
| SpecKit (7-phase SDD) | - | Full decomposition | - | - | `specKit.Execute` |
| Reflexion (episodic memory) | Prior learnings | - | - | Generate | `reflexionManager.Reflect` |
| Provenance (14 event types) | - | - | Track all calls | Audit trail | `provenanceTracker.Record` |
| Circuit Breakers | - | - | Provider fallback | - | Per-provider CB |
| BigData (infinite context) | Distributed memory | - | Knowledge streaming | - | `bigDataIntegration` |
| Plugins (dynamic loading) | - | - | Plugin execution | - | `pluginRegistry.Execute` |

---

## New Files

| File | Purpose |
|------|---------|
| `internal/services/agentic_ensemble.go` | Core orchestrator — dual-mode Process(), mode selection |
| `internal/services/agentic_ensemble_config.go` | Configuration with sensible defaults |
| `internal/services/iterative_tool_executor.go` | Per-phase iterative tool loop |
| `internal/services/agent_worker_pool.go` | Background agent spawning, dependency graph, lifecycle |
| `internal/services/agent_worker.go` | Individual agent execution loop with tool access |
| `internal/services/execution_planner.go` | SpecKit integration — decompose decisions into executable tasks |
| `internal/services/verification_debate.go` | Post-execution verification debate |
| `internal/services/vision_tool_adapter.go` | Vision integration wrapping existing `internal/handlers/vision_handler.go` into VisionClient interface for ToolIntegration |

---

## Modified Files

| File | Change |
|------|--------|
| `internal/handlers/openai_compatible.go` | Wire AgenticEnsemble into processWithEnsemble() |
| `internal/debate/tools/tool_integration.go` | Add VisionClient, full HelixMemory 4-engine access |
| `internal/router/router.go` | Initialize AgenticEnsemble with all dependencies |
| `internal/services/debate_service.go` | Expose iterative tool hook for per-phase tool execution |

---

## Testing

| Test Type | Files | What's Tested |
|-----------|-------|---------------|
| Unit | `agentic_ensemble_test.go`, `iterative_tool_executor_test.go`, `agent_worker_pool_test.go`, `agent_worker_test.go`, `execution_planner_test.go`, `verification_debate_test.go`, `vision_tool_adapter_test.go` | Each component in isolation |
| Integration | `tests/integration/agentic_ensemble_integration_test.go` | Full pipeline with real tool bridge |
| E2E | `tests/e2e/agentic_ensemble_e2e_test.go` | HTTP request → ensemble → tools → response |
| Stress | `tests/stress/agentic_ensemble_stress_test.go` | Concurrent requests, pool saturation, tool contention |
| Security | `tests/security/agentic_ensemble_security_test.go` | Tool injection, agent isolation, output sanitization |
| Benchmark | `tests/performance/agentic_ensemble_benchmark_test.go` | Mode selection, agent spawn, tool iteration throughput |
| Fuzz | `tests/fuzz/agentic_tool_call_fuzz_test.go` | Tool call params, malformed tool responses |
| Chaos | `tests/chaos/agentic_ensemble_chaos_test.go` | Provider failures mid-execution, tool timeouts, agent crashes |
| Automation | `tests/automation/agentic_ensemble_automation_test.go` | End-to-end automated workflow: request → mode detection → execution → verification → response |

---

## Challenges

| Challenge | Tests | Validates |
|-----------|-------|-----------|
| `agentic_ensemble_challenge.sh` | 30+ | Dual-mode routing, full pipeline, tool integration, agent spawning, verification |
| `agentic_tool_loop_challenge.sh` | 20+ | Iterative tool execution, all 6 protocols, MaxIterations enforcement |
| `agentic_background_agents_challenge.sh` | 20+ | Parallel execution, dependency graph, graceful shutdown, resource limits |

---

## Documentation

| Document | Update |
|----------|--------|
| `CLAUDE.md` | Add AgenticEnsemble to Architecture, update request flow diagram |
| `AGENTS.md` | Add agentic ensemble agent descriptions |
| `docs/architecture/AGENTIC_ENSEMBLE.md` | New — full architecture documentation |
| `Website/user-manuals/45-agentic-ensemble-guide.md` | New — step-by-step user manual |
| `Website/video-courses/course-76-agentic-ensemble.md` | New — video course |
| `CHANGELOG.md` | Add agentic ensemble entry |
| `docs/api/API_REFERENCE.md` | Document new response metadata fields |
| `docs/MODULES.md` | Reference AgenticEnsemble integration |

---

## Streaming Protocol (Execute Mode)

Execute mode streams progress via the existing SSE infrastructure (`internal/notifications/sse_manager.go`):

```
SSE Event Types:
  agentic.stage     → {stage: "UNDERSTAND"|"PLAN"|"ASSIGN"|"EXECUTE"|"VERIFY"|"SYNTHESIZE", status: "started"|"completed"}
  agentic.agent     → {agent_id: "...", task_id: "...", status: "started"|"reasoning"|"tool_call"|"completed"|"failed"}
  agentic.tool      → {agent_id: "...", protocol: "mcp"|"lsp"|"rag"|..., operation: "...", status: "started"|"completed"}
  agentic.progress  → {completed_tasks: N, total_tasks: M, percentage: P}
  content.delta     → {content: "..."} (standard OpenAI SSE format for final response)
```

Clients receive these via the existing `text/event-stream` response on `/v1/chat/completions` when `stream: true`. No new SSE endpoints needed.

## API Endpoints

All agentic execution is accessed through the **existing** `/v1/chat/completions` endpoint. No new REST endpoints are created. However, the response metadata is extended:

```json
{
  "choices": [{"message": {"content": "...", "role": "assistant"}}],
  "usage": {"prompt_tokens": N, "completion_tokens": M},
  "agentic_metadata": {
    "mode": "execute",
    "stages_completed": ["UNDERSTAND", "PLAN", "ASSIGN", "EXECUTE", "VERIFY", "SYNTHESIZE"],
    "agents_spawned": 5,
    "tasks_completed": 8,
    "tools_invoked": [{"protocol": "lsp", "count": 12}, {"protocol": "mcp", "count": 3}],
    "total_duration_ms": 45000,
    "provenance_id": "prov-abc123"
  }
}
```

**Cancellation:** Clients cancel by closing the HTTP connection. The context propagation via `ctx.Done()` through AgenticEnsemble → WorkerPool → individual agents ensures graceful shutdown.

---

## Error Handling & Safety

| Concern | Mechanism | Default |
|---------|-----------|---------|
| Agent timeout | Per-agent context with deadline | 5 minutes |
| Tool failure | Circuit breaker per protocol | 3 failures → skip, continue reasoning-only |
| Infinite loop | MaxIterations per agent + per phase + global | 20/5/15min |
| Resource limits | Semaphore for agents, GOMAXPROCS for tests | 5 agents, GOMAXPROCS=2 |
| Security | All outputs through guardrails + PII detection | Enabled by default |
| Provider failure | Circuit breaker + fallback chain | Existing CB infrastructure |
| Graceful shutdown | WaitGroup + context.Cancel | Drain all agents before exit |
| Audit | Provenance tracker records every decision + tool call | Always on |

---

## Constraints

- All code follows existing Go conventions, Gin patterns, testify testing
- No mocks in production code
- No CI/CD pipelines
- SSH only for git operations
- Resource limits (GOMAXPROCS=2, nice -n 19) for tests/challenges
- HTTP/3 (QUIC) primary transport
- Container-based builds for releases
- Conventional commits

---

*This spec was collaboratively designed through brainstorming on 2026-03-25.*
