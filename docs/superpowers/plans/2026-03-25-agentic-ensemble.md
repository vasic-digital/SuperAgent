# AgenticEnsemble Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the AI ensemble behave as a single unified LLM with dual-mode operation: tool-augmented reasoning for queries and autonomous execution via background agents for actionable requests.

**Architecture:** New `AgenticEnsemble` orchestrator wraps existing debate/ensemble infrastructure, adds iterative tool execution per debate phase, background agent worker pool with SpecKit task decomposition, and verification debate. Integrates all power features: MCP, LSP, ACP, RAG, Embeddings, Vision, HelixMemory (4-engine), Formatters, Security, BigData, Plugins, Reflexion, Provenance.

**Tech Stack:** Go 1.25.3, Gin v1.12.0, testify v1.11.1, sync.Once, context.Context, sync.WaitGroup

**Spec:** `docs/superpowers/specs/2026-03-25-agentic-ensemble-design.md`

---

### Task 1: Foundation Types and Interfaces

**Files:**
- Create: `internal/services/agentic_ensemble_types.go`
- Create: `internal/services/agentic_ensemble_types_test.go`

Define all new types that other components depend on.

- [ ] **Step 1: Create type definitions**

```go
// internal/services/agentic_ensemble_types.go
package services

import "time"

// AgenticMode determines how the ensemble processes a request.
type AgenticMode int

const (
    AgenticModeReason  AgenticMode = iota // Tool-augmented debate
    AgenticModeExecute                     // Full autonomous execution
)

func (m AgenticMode) String() string {
    switch m {
    case AgenticModeReason:  return "reason"
    case AgenticModeExecute: return "execute"
    default:                 return "unknown"
    }
}

// AgenticTask is a structured task from SpecKit decomposition.
type AgenticTask struct {
    ID               string
    Description      string
    Dependencies     []string          // IDs of prerequisite tasks
    ToolRequirements []string          // Required: "mcp", "lsp", "rag", "vision", etc.
    Priority         int
    EstimatedSteps   int
    Status           AgenticTaskStatus
}

type AgenticTaskStatus int

const (
    AgenticTaskPending AgenticTaskStatus = iota
    AgenticTaskRunning
    AgenticTaskCompleted
    AgenticTaskFailed
)

// AgenticResult is the output of an agent worker.
type AgenticResult struct {
    TaskID      string
    AgentID     string
    Content     string
    ToolCalls   []ToolExecution
    Duration    time.Duration
    Error       error
}

// ToolExecution records a single tool invocation.
type ToolExecution struct {
    Protocol    string        // "mcp", "lsp", "acp", "rag", "embeddings", "vision"
    Operation   string
    Input       interface{}
    Output      interface{}
    Duration    time.Duration
    Error       error
}

// AgenticEnsembleConfig holds all configuration.
type AgenticEnsembleConfig struct {
    MaxConcurrentAgents  int           // Default: 5
    MaxIterationsPerAgent int          // Default: 20
    MaxToolIterationsPerPhase int      // Default: 5
    AgentTimeout         time.Duration // Default: 5 min
    GlobalTimeout        time.Duration // Default: 15 min
    ToolIterationTimeout time.Duration // Default: 30s
    EnableVision         bool          // Default: true
    EnableMemory         bool          // Default: true
    EnableExecution      bool          // Default: true
}

func DefaultAgenticEnsembleConfig() AgenticEnsembleConfig {
    return AgenticEnsembleConfig{
        MaxConcurrentAgents:       5,
        MaxIterationsPerAgent:     20,
        MaxToolIterationsPerPhase: 5,
        AgentTimeout:              5 * time.Minute,
        GlobalTimeout:             15 * time.Minute,
        ToolIterationTimeout:      30 * time.Second,
        EnableVision:              true,
        EnableMemory:              true,
        EnableExecution:           true,
    }
}
```

- [ ] **Step 2: Write tests for types**

Test: String() methods, default config values, status transitions.

- [ ] **Step 3: Run tests**

Run: `GOMAXPROCS=2 go test ./internal/services/ -run TestAgentic -short -count=1 -v`

- [ ] **Step 4: Commit**

```bash
git add internal/services/agentic_ensemble_types.go internal/services/agentic_ensemble_types_test.go
git commit -m "feat(agentic): add foundation types for AgenticEnsemble

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: Extend ToolIntegration with Vision and HelixMemory

**Files:**
- Modify: `internal/debate/tools/tool_integration.go`
- Create: `internal/services/vision_tool_adapter.go`
- Create: `internal/services/vision_tool_adapter_test.go`

- [ ] **Step 1: Add VisionClient and HelixMemoryClient interfaces to tool_integration.go**

Add after existing interface definitions (after line ~90):

```go
// VisionClient provides access to image/screenshot analysis.
type VisionClient interface {
    AnalyzeImage(ctx context.Context, imageData []byte, prompt string) (interface{}, error)
    AnalyzeURL(ctx context.Context, imageURL string, prompt string) (interface{}, error)
}

// HelixMemoryClient provides full 4-engine memory access.
type HelixMemoryClient interface {
    StoreFact(ctx context.Context, fact string, metadata map[string]string) error
    RecallFacts(ctx context.Context, query string, limit int) ([]string, error)
    BuildGraph(ctx context.Context, content string) error
    QueryGraph(ctx context.Context, query string) ([]interface{}, error)
    CreateAgentSession(ctx context.Context, agentID string) (string, error)
    SendAgentMessage(ctx context.Context, sessionID string, message string) (string, error)
    AddTemporalEdge(ctx context.Context, from, to, relation string, timestamp time.Time) error
    QueryTimeline(ctx context.Context, entity string) ([]interface{}, error)
}
```

- [ ] **Step 2: Add fields to ToolIntegration struct**

Add `visionClient VisionClient` and `memoryClient HelixMemoryClient` fields. Update constructor with optional parameters (backward compatible — nil means disabled).

- [ ] **Step 3: Create VisionToolAdapter**

`internal/services/vision_tool_adapter.go` — wraps the existing `internal/handlers/vision_handler.go` into the `VisionClient` interface.

- [ ] **Step 4: Write adapter tests**

- [ ] **Step 5: Run tests and commit**

```bash
go test ./internal/debate/tools/ -short -count=1
go test ./internal/services/ -run TestVision -short -count=1
git add internal/debate/tools/tool_integration.go internal/services/vision_tool_adapter.go internal/services/vision_tool_adapter_test.go
git commit -m "feat(tools): extend ToolIntegration with Vision and HelixMemory clients

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: Iterative Tool Executor

**Files:**
- Create: `internal/services/iterative_tool_executor.go`
- Create: `internal/services/iterative_tool_executor_test.go`

The core component enabling per-phase iterative tool loops.

- [ ] **Step 1: Implement IterativeToolExecutor**

```go
// internal/services/iterative_tool_executor.go
package services

type IterativeToolExecutor struct {
    toolBridge       *tools.ToolIntegration
    maxIterations    int
    iterationTimeout time.Duration
    logger           *logrus.Logger
}

func NewIterativeToolExecutor(
    toolBridge *tools.ToolIntegration,
    config AgenticEnsembleConfig,
    logger *logrus.Logger,
) *IterativeToolExecutor

// ExecuteWithTools runs an LLM call and iteratively resolves tool calls.
// If the LLM response contains tool_calls, it executes them via ToolIntegration
// and feeds results back for another LLM call. Repeats up to maxIterations.
func (e *IterativeToolExecutor) ExecuteWithTools(
    ctx context.Context,
    provider llm.LLMProvider,
    messages []models.Message,
    availableTools []tools.ToolSchema,
) (*models.LLMResponse, []ToolExecution, error)
```

The loop:
1. Call `provider.Complete(ctx, request)` with tools
2. If response has `tool_calls` and iterations < max:
   - Execute each tool call via appropriate protocol (MCP/LSP/RAG/etc.)
   - Append tool results as messages
   - Go to step 1
3. Return final response + all tool executions

- [ ] **Step 2: Write comprehensive tests**

Test: no tool calls (passthrough), single tool call iteration, multiple iterations, max iterations hit, tool failure with circuit breaker, context cancellation, concurrent safety.

- [ ] **Step 3: Run tests and commit**

```bash
go test ./internal/services/ -run TestIterativeToolExecutor -short -count=1 -v
git add internal/services/iterative_tool_executor.go internal/services/iterative_tool_executor_test.go
git commit -m "feat(agentic): implement IterativeToolExecutor for per-phase tool loops

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 4: Execution Planner (SpecKit Integration)

**Files:**
- Create: `internal/services/execution_planner.go`
- Create: `internal/services/execution_planner_test.go`

Bridges SpecKit debate output to structured AgenticTasks.

- [ ] **Step 1: Implement ExecutionPlanner**

```go
// internal/services/execution_planner.go
package services

type ExecutionPlanner struct {
    specKit          *SpecKitOrchestrator
    intentClassifier *LLMIntentClassifier
    toolExecutor     *IterativeToolExecutor
    logger           *logrus.Logger
}

// DecomposePlan takes a debate decision and decomposes it into executable tasks.
// Uses SpecKit's 7-phase flow for complex plans, or direct LLM decomposition for simple ones.
func (p *ExecutionPlanner) DecomposePlan(
    ctx context.Context,
    decision string,
    provider llm.LLMProvider,
) ([]AgenticTask, error)

// ParseTasksFromDebate extracts structured AgenticTask list from
// SpecKit Tasks phase output (debate result string).
func (p *ExecutionPlanner) ParseTasksFromDebate(
    debateOutput string,
    provider llm.LLMProvider,
) ([]AgenticTask, error)

// BuildDependencyGraph returns tasks in topological execution order.
func (p *ExecutionPlanner) BuildDependencyGraph(
    tasks []AgenticTask,
) ([][]AgenticTask, error) // Returns layers: [independent][dependent_on_layer0]...
```

- [ ] **Step 2: Write tests**

Test: simple 1-task decomposition, multi-task with dependencies, circular dependency detection, empty plan handling, SpecKit integration (mock SpecKit output).

- [ ] **Step 3: Run tests and commit**

```bash
go test ./internal/services/ -run TestExecutionPlanner -short -count=1 -v
git add internal/services/execution_planner.go internal/services/execution_planner_test.go
git commit -m "feat(agentic): implement ExecutionPlanner with SpecKit integration

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 5: Agent Worker and Worker Pool

**Files:**
- Create: `internal/services/agent_worker.go`
- Create: `internal/services/agent_worker_test.go`
- Create: `internal/services/agent_worker_pool.go`
- Create: `internal/services/agent_worker_pool_test.go`

- [ ] **Step 1: Implement AgentWorker**

Individual agent that runs an iterative tool loop for one task.

```go
// internal/services/agent_worker.go
package services

type AgentWorker struct {
    id             string
    task           AgenticTask
    provider       llm.LLMProvider
    toolExecutor   *IterativeToolExecutor
    memoryClient   tools.HelixMemoryClient // For Letta agent state
    lettaSessionID string
    maxIterations  int
    logger         *logrus.Logger
}

// Execute runs the agent's task through iterative reasoning + tool use.
func (w *AgentWorker) Execute(ctx context.Context) AgenticResult
```

- [ ] **Step 2: Implement AgentWorkerPool**

```go
// internal/services/agent_worker_pool.go
package services

type AgentWorkerPool struct {
    semaphore       chan struct{}
    providers       *ProviderRegistry
    toolBridge      *tools.ToolIntegration
    memoryClient    tools.HelixMemoryClient
    eventBus        interface{} // EventBus adapter
    provenance      *audit.ProvenanceTracker
    wg              sync.WaitGroup
    ctx             context.Context
    cancel          context.CancelFunc
    logger          *logrus.Logger
}

// DispatchTasks assigns tasks to agents respecting dependencies.
// Returns a channel that receives results as agents complete.
func (p *AgentWorkerPool) DispatchTasks(
    ctx context.Context,
    taskLayers [][]AgenticTask,
) (<-chan AgenticResult, error)

// Shutdown gracefully stops all agents.
func (p *AgentWorkerPool) Shutdown()

// selectProvider picks best tool-capable provider for a task.
func (p *AgentWorkerPool) selectProvider(task AgenticTask) llm.LLMProvider
```

- [ ] **Step 3: Write tests for AgentWorker**

Test: single-step task, multi-step with tools, max iterations hit, context cancellation, tool failure recovery.

- [ ] **Step 4: Write tests for AgentWorkerPool**

Test: single task dispatch, parallel dispatch, dependency ordering, semaphore limiting, graceful shutdown, provider selection (tool-capable preference).

- [ ] **Step 5: Run tests and commit**

```bash
go test ./internal/services/ -run "TestAgentWorker|TestAgentWorkerPool" -short -count=1 -v
git add internal/services/agent_worker.go internal/services/agent_worker_test.go internal/services/agent_worker_pool.go internal/services/agent_worker_pool_test.go
git commit -m "feat(agentic): implement AgentWorker and AgentWorkerPool with dependency graph

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 6: Verification Debate

**Files:**
- Create: `internal/services/verification_debate.go`
- Create: `internal/services/verification_debate_test.go`

Post-execution verification using a focused debate.

- [ ] **Step 1: Implement VerificationDebate**

```go
// internal/services/verification_debate.go
package services

type VerificationDebate struct {
    debateService     *DebateService
    toolExecutor      *IterativeToolExecutor
    guardrails        interface{} // security.StandardGuardrailPipeline
    formatterRegistry interface{} // formatters.Registry
    logger            *logrus.Logger
}

// Verify runs a verification debate on agent results.
// Uses LSP diagnostics, formatter checks, security scans, and
// a focused debate on completeness/correctness.
func (v *VerificationDebate) Verify(
    ctx context.Context,
    originalRequest string,
    results []AgenticResult,
) (*VerificationResult, error)

type VerificationResult struct {
    Approved     bool
    Confidence   float64
    Issues       []string
    FailedTasks  []string // Task IDs that need re-execution
}
```

- [ ] **Step 2: Write tests**

Test: all results pass, partial failure, security issue detected, formatter applied, re-execution recommended.

- [ ] **Step 3: Run tests and commit**

```bash
go test ./internal/services/ -run TestVerificationDebate -short -count=1 -v
git add internal/services/verification_debate.go internal/services/verification_debate_test.go
git commit -m "feat(agentic): implement VerificationDebate for post-execution validation

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 7: AgenticEnsemble Core Orchestrator

**Files:**
- Create: `internal/services/agentic_ensemble.go`
- Create: `internal/services/agentic_ensemble_test.go`

The main component that ties everything together.

- [ ] **Step 1: Implement AgenticEnsemble**

```go
// internal/services/agentic_ensemble.go
package services

type AgenticEnsemble struct {
    debateService     *DebateService
    specKit           *SpecKitOrchestrator
    toolBridge        *tools.ToolIntegration
    intentClassifier  *LLMIntentClassifier
    workerPool        *AgentWorkerPool
    toolExecutor      *IterativeToolExecutor
    planner           *ExecutionPlanner
    verifier          *VerificationDebate
    providerRegistry  *ProviderRegistry
    provenanceTracker *audit.ProvenanceTracker
    reflexionLoop     *reflexion.ReflexionLoop
    memoryClient      tools.HelixMemoryClient
    eventBus          interface{}
    config            AgenticEnsembleConfig
    logger            *logrus.Logger
}

// Process is the main entry point. Classifies intent and routes to
// Reason mode (tool-augmented debate) or Execute mode (agentic loop).
func (e *AgenticEnsemble) Process(
    ctx context.Context,
    req *models.LLMRequest,
) (*EnsembleResult, error)

// toolAugmentedDebate runs debate with per-phase iterative tool access.
func (e *AgenticEnsemble) toolAugmentedDebate(
    ctx context.Context,
    req *models.LLMRequest,
) (*EnsembleResult, error)

// agenticExecutionLoop runs the full 7-stage execution pipeline.
func (e *AgenticEnsemble) agenticExecutionLoop(
    ctx context.Context,
    req *models.LLMRequest,
) (*EnsembleResult, error)

// classifyMode determines Reason vs Execute based on intent.
func (e *AgenticEnsemble) classifyMode(
    ctx context.Context,
    req *models.LLMRequest,
) (AgenticMode, error)
```

- [ ] **Step 2: Implement toolAugmentedDebate**

Calls existing `debateService.ConductDebate()` but wraps each phase invocation with `toolExecutor.ExecuteWithTools()` so agents can use tools during their phase.

- [ ] **Step 3: Implement agenticExecutionLoop**

The 7 stages: Understand → Plan → Assign → Execute → Verify → Synthesize → Respond.

- [ ] **Step 4: Write comprehensive tests**

Test: reason mode routing, execute mode routing, tool-augmented debate with mock tools, full execution loop with mock agents, verification failure → re-execution, config defaults, graceful shutdown.

- [ ] **Step 5: Run tests and commit**

```bash
go test ./internal/services/ -run TestAgenticEnsemble -short -count=1 -v
git add internal/services/agentic_ensemble.go internal/services/agentic_ensemble_test.go
git commit -m "feat(agentic): implement AgenticEnsemble dual-mode orchestrator

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 8: Wire Into OpenAI-Compatible Handler and Router

**Files:**
- Modify: `internal/handlers/openai_compatible.go:2460`
- Modify: `internal/router/router.go`

- [ ] **Step 1: Add AgenticEnsemble field to UnifiedHandler**

In `openai_compatible.go`, add `agenticEnsemble *services.AgenticEnsemble` field and setter.

- [ ] **Step 2: Update processWithEnsemble()**

At line 2460, add AgenticEnsemble as the primary path:

```go
func (h *UnifiedHandler) processWithEnsemble(ctx context.Context, req *models.LLMRequest, openaiReq *OpenAIChatRequest) (*services.EnsembleResult, error) {
    // Primary: AgenticEnsemble (dual-mode: reason + execute)
    if h.agenticEnsemble != nil {
        logrus.Info("[CODE PATH] Using AgenticEnsemble (unified dual-mode)")
        return h.agenticEnsemble.Process(ctx, req)
    }
    // Fallback: Orchestrator (8-phase debate only)
    if h.orchestratorIntegration != nil {
        logrus.Info("[CODE PATH] Using debate orchestrator fallback")
        return h.processWithOrchestrator(ctx, req, openaiReq)
    }
    // Legacy fallback
    // ... existing code ...
}
```

- [ ] **Step 3: Initialize AgenticEnsemble in router.go**

After the existing orchestrator initialization (around line 863-907), create and wire the AgenticEnsemble with all dependencies.

- [ ] **Step 4: Verify compilation and existing tests**

```bash
go build ./internal/handlers/ ./internal/router/
go test ./internal/router/ -short -count=1
```

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/openai_compatible.go internal/router/router.go
git commit -m "feat(agentic): wire AgenticEnsemble into request pipeline

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 9: Full Test Suite

**Files:**
- Create: `internal/services/agentic_ensemble_integration_test.go`
- Create: `tests/integration/agentic_ensemble_integration_test.go`
- Create: `tests/e2e/agentic_ensemble_e2e_test.go`
- Create: `tests/stress/agentic_ensemble_stress_test.go`
- Create: `tests/security/agentic_ensemble_security_test.go`
- Create: `tests/performance/agentic_ensemble_benchmark_test.go`
- Create: `tests/fuzz/agentic_tool_call_fuzz_test.go`
- Create: `tests/chaos/agentic_ensemble_chaos_test.go`
- Create: `tests/automation/agentic_ensemble_automation_test.go`

- [ ] **Step 1: Integration tests**

Test full pipeline with real ToolIntegration (requires infra). Tag: `//go:build integration`

- [ ] **Step 2: E2E tests**

HTTP request → AgenticEnsemble → response. Tag: `//go:build e2e`

- [ ] **Step 3: Stress tests**

Concurrent requests, agent pool saturation. Tag: `//go:build stress`. GOMAXPROCS=2.

- [ ] **Step 4: Security tests**

Tool injection prevention, agent isolation, output sanitization. Tag: `//go:build security`

- [ ] **Step 5: Benchmark tests**

Mode selection latency, agent spawn overhead, tool iteration throughput. Tag: `//go:build performance`

- [ ] **Step 6: Fuzz tests**

Fuzz tool call parameters, malformed responses. Tag: `//go:build fuzz`

- [ ] **Step 7: Chaos tests**

Provider failures mid-execution, tool timeouts, agent crashes. Tag: `//go:build chaos`

- [ ] **Step 8: Automation tests**

End-to-end automated workflow. Tag: `//go:build automation`

- [ ] **Step 9: Run all and commit**

```bash
go test ./internal/services/ -run TestAgentic -short -count=1
go test ./tests/... -tags integration,stress,security,performance,chaos,automation -run='^$' -count=1
git add internal/services/ tests/
git commit -m "test(agentic): add comprehensive test suite (unit, integration, e2e, stress, security, bench, fuzz, chaos, automation)

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 10: Challenge Scripts

**Files:**
- Create: `challenges/scripts/agentic_ensemble_challenge.sh`
- Create: `challenges/scripts/agentic_tool_loop_challenge.sh`
- Create: `challenges/scripts/agentic_background_agents_challenge.sh`

- [ ] **Step 1: Write agentic_ensemble_challenge.sh**

Validates: AgenticEnsemble struct exists, dual-mode Process method, all power features referenced, config defaults, builds cleanly.

- [ ] **Step 2: Write agentic_tool_loop_challenge.sh**

Validates: IterativeToolExecutor exists, all 6 protocols in ToolIntegration (MCP/LSP/ACP/RAG/Embeddings/Vision), MaxIterations enforcement, tool execution recording.

- [ ] **Step 3: Write agentic_background_agents_challenge.sh**

Validates: AgentWorkerPool exists, semaphore limiting, dependency graph, graceful shutdown (WaitGroup+context), provider selection.

- [ ] **Step 4: Make executable and commit**

```bash
chmod +x challenges/scripts/agentic_ensemble_challenge.sh challenges/scripts/agentic_tool_loop_challenge.sh challenges/scripts/agentic_background_agents_challenge.sh
git add challenges/scripts/
git commit -m "test(challenges): add agentic ensemble, tool loop, and background agents challenges

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 11: Documentation and Content

**Files:**
- Create: `docs/architecture/AGENTIC_ENSEMBLE.md`
- Create: `Website/user-manuals/45-agentic-ensemble-guide.md`
- Create: `Website/video-courses/course-76-agentic-ensemble.md`
- Modify: `CLAUDE.md`
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Write architecture doc**

`docs/architecture/AGENTIC_ENSEMBLE.md` — full architecture with Mermaid diagrams showing request flow, 7-stage pipeline, agent lifecycle.

- [ ] **Step 2: Write user manual**

`Website/user-manuals/45-agentic-ensemble-guide.md` — step-by-step: configuration, reason mode usage, execute mode usage, monitoring via SSE, troubleshooting.

- [ ] **Step 3: Write video course**

`Website/video-courses/course-76-agentic-ensemble.md` — 6 modules covering architecture, tool-augmented debate, execution pipeline, agent workers, verification, production tuning.

- [ ] **Step 4: Update CLAUDE.md**

Add AgenticEnsemble to Architecture section. Update request flow diagram. Add to Architectural Patterns.

- [ ] **Step 5: Update CHANGELOG.md**

Add AgenticEnsemble entry under [Unreleased].

- [ ] **Step 6: Commit**

```bash
git add docs/architecture/AGENTIC_ENSEMBLE.md Website/user-manuals/45-agentic-ensemble-guide.md Website/video-courses/course-76-agentic-ensemble.md CLAUDE.md CHANGELOG.md
git commit -m "docs(agentic): add architecture docs, user manual, video course, update CLAUDE.md

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 12: Final Validation

- [ ] **Step 1: Full build**

Run: `go build ./cmd/... ./internal/...`
Expected: Zero errors

- [ ] **Step 2: Full vet**

Run: `go vet ./internal/...`
Expected: Zero new warnings

- [ ] **Step 3: Unit tests**

Run: `GOMAXPROCS=2 nice -n 19 go test ./internal/services/ -run TestAgentic -short -count=1 -v`
Expected: All pass

- [ ] **Step 4: Run challenges**

```bash
./challenges/scripts/agentic_ensemble_challenge.sh
./challenges/scripts/agentic_tool_loop_challenge.sh
./challenges/scripts/agentic_background_agents_challenge.sh
```

- [ ] **Step 5: Push**

```bash
git push githubhelixdevelopment main
git push upstream main
```
