# Video Course 76: Agentic Ensemble -- Dual-Mode LLM Execution

## Course Overview

**Duration:** 3 hours
**Level:** Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 19 (Concurrency Patterns), Course 60 (Enterprise Deployment)

Master HelixAgent's AgenticEnsemble system: dual-mode operation (Reason + Execute), iterative tool loops across 6 protocols, LLM-based task decomposition with dependency graphs, semaphore-limited agent worker pools, and post-execution verification debates. Learn to configure, monitor, and scale the ensemble for production workloads.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Explain the dual-mode architecture and when each mode activates
2. Configure iterative tool loops for debate phases
3. Trace a request through all 7 pipeline stages
4. Build and debug task dependency graphs
5. Monitor agent execution via SSE events and Prometheus metrics
6. Tune the ensemble for latency, throughput, and resource constraints

---

## Module 1: Architecture Overview (30 min)

### Video 1.1: Dual-Mode Design (15 min)

**Topics:**
- Problem statement: why reasoning alone is insufficient for complex tasks
- Reason mode: tool-augmented debate without side effects
- Execute mode: autonomous task execution with verification
- How intent classification selects the mode
- The `AgenticMode` enum and `AgenticEnsembleConfig` struct
- Key file: `internal/services/agentic_ensemble_types.go`

### Video 1.2: The 7-Stage Request Flow (15 min)

**Topics:**
- Stage 1: Intent Classification via `LLMIntentClassifier`
- Stage 2: Mode Selection in `AgenticEnsemble.Process`
- Stage 3a/3b: Tool-Augmented Debate vs. Task Decomposition
- Stage 4: Iterative Tool Loops within debate phases
- Stage 5: Agent Worker Dispatch with dependency layering
- Stage 6: Response Synthesis and metadata assembly
- Stage 7: Verification Debate for quality assurance
- Tracing a request end-to-end with structured logging

---

## Module 2: Tool-Augmented Debate (30 min)

### Video 2.1: ToolIntegration Deep Dive (15 min)

**Topics:**
- The `ToolIntegration` struct and its 8 typed clients
- Nil-safe client pattern: graceful degradation when protocols are unavailable
- `ListAvailableTools` aggregation across MCP, LSP, RAG, Embeddings, Vision, HelixMemory, Formatters, ACP
- Adding a new tool protocol to the integration layer
- Key file: `internal/debate/tools/tool_integration.go`

### Video 2.2: Iterative Tool Loops and the 6 Protocols (15 min)

**Topics:**
- Per-phase tool iteration: `MaxToolIterationsPerPhase` controls loop count
- MCP: `CallTool`, `ListTools`, `GetResource` for external service access
- LSP: `GetDefinition`, `GetReferences`, `GetDiagnostics` for code intelligence
- RAG: `Search`, `Store`, `Rerank` for knowledge retrieval
- Embeddings: `Embed`, `EmbedBatch`, `Similarity` for semantic operations
- ACP: `SendMessage`, `Subscribe` for inter-agent communication
- Vision and HelixMemory: image analysis and 4-engine cognitive memory
- Timeout management with `ToolIterationTimeout`

---

## Module 3: Agentic Execution Pipeline (30 min)

### Video 3.1: ExecutionPlanner and Task Decomposition (15 min)

**Topics:**
- LLM-based plan decomposition in `ExecutionPlanner.DecomposePlan`
- The `AgenticTask` struct: ID, description, dependencies, tool requirements, priority
- JSON extraction from LLM output (handling markdown code blocks)
- UUID generation for task IDs when the LLM omits them
- Key file: `internal/services/execution_planner.go`

### Video 3.2: Dependency Graph Construction (15 min)

**Topics:**
- Kahn's algorithm for topological layering in `BuildDependencyGraph`
- Dependency validation: detecting unknown references
- Circular dependency detection and error reporting
- Layer-based parallel execution: independent tasks run concurrently
- Visualizing dependency graphs for debugging

---

## Module 4: Background Agent Workers (30 min)

### Video 4.1: AgentWorkerPool Design (15 min)

**Topics:**
- Semaphore-limited goroutine pool bounded by `MaxConcurrentAgents`
- Layer-by-layer execution: wait for all tasks in a layer before advancing
- Per-agent timeout enforcement via `context.WithTimeout`
- The `AgenticResult` struct: task ID, agent ID, content, tool calls, duration, error
- Goroutine lifecycle safety: `sync.WaitGroup` for graceful shutdown

### Video 4.2: Agent Lifecycle and Error Handling (15 min)

**Topics:**
- Task status transitions: Pending, Running, Completed, Failed
- Individual agent failure isolation: other agents continue
- Tool call recording in `AgenticToolExecution` (protocol, operation, input, output, duration)
- Global timeout propagation via context cancellation
- Resource monitoring integration with `GOMAXPROCS` limits

---

## Module 5: Verification and Safety (30 min)

### Video 5.1: VerificationDebate (15 min)

**Topics:**
- Post-execution quality gate in `VerificationDebate.Verify`
- Constructing verification prompts from `AgenticResult` summaries
- Three evaluation dimensions: completeness, correctness, coherence
- Parsing structured verification responses (APPROVED/ISSUES format)
- Graceful degradation: low-confidence pass when the LLM is unavailable
- Key file: `internal/services/verification_debate.go`

### Video 5.2: Security Guardrails (15 min)

**Topics:**
- Content filtering and PII detection via Security module integration
- Tool invocation sandboxing: preventing unintended side effects in Reason mode
- Rate limiting on agent worker spawning
- Circuit breaker integration for external tool calls
- Audit trail: `AgenticMetadata` with provenance ID and tool invocation summary

---

## Module 6: Production Deployment (30 min)

### Video 6.1: Configuration and Tuning (15 min)

**Topics:**
- Environment variable reference for all `AgenticEnsembleConfig` fields
- Tuning `MaxConcurrentAgents` for I/O-bound vs. CPU-bound workloads
- Adjusting `MaxToolIterationsPerPhase` for depth vs. latency trade-offs
- Setting `GlobalTimeout` based on task complexity profiles
- Disabling Execute mode with `AGENTIC_ENABLE_EXECUTION=false`

### Video 6.2: Monitoring and Scaling (15 min)

**Topics:**
- SSE event stream for real-time progress tracking
- Prometheus metrics: agent spawn count, task duration histograms, tool invocation rates
- Health check integration at `/v1/monitoring/status`
- Horizontal scaling: multiple HelixAgent instances with shared provider registry
- Debugging with `AgenticMetadata`: stages completed, agents spawned, tasks completed
- Challenge validation: `challenges/scripts/agentic_ensemble_challenge.sh` (32 tests)

---

## Summary

The AgenticEnsemble unifies reasoning and execution into a single pipeline. Reason mode provides tool-augmented debate for analytical queries. Execute mode decomposes decisions into tasks, dispatches them to background workers, and verifies results through an LLM-based quality gate. Both modes share the same tool integration layer, configuration system, and monitoring infrastructure.
