# User Manual 43: Agentic Ensemble Guide

## Overview

The AgenticEnsemble extends HelixAgent's multi-provider debate with two operating modes: **Reason** (tool-augmented deliberation) and **Execute** (autonomous task execution with background agent workers). This guide covers configuration, usage of both modes, monitoring, and troubleshooting.

## Prerequisites

- HelixAgent binary built and running (`make build && ./bin/helixagent`)
- At least one LLM provider configured with a valid API key in `.env`
- For Execute mode: infrastructure containers running (PostgreSQL, Redis)
- For tool protocols: relevant services enabled (MCP servers, LSP, RAG, etc.)

## Configuration

All settings live in `AgenticEnsembleConfig`. Defaults are production-ready; override via environment variables when needed.

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `AGENTIC_MAX_CONCURRENT_AGENTS` | 5 | Maximum parallel agent workers in Execute mode |
| `AGENTIC_MAX_ITERATIONS_PER_AGENT` | 20 | Maximum LLM reasoning steps per agent |
| `AGENTIC_MAX_TOOL_ITERATIONS_PER_PHASE` | 5 | Tool invocation loops per debate phase |
| `AGENTIC_AGENT_TIMEOUT` | 5m | Per-agent execution timeout |
| `AGENTIC_GLOBAL_TIMEOUT` | 15m | Overall pipeline timeout |
| `AGENTIC_TOOL_ITERATION_TIMEOUT` | 30s | Timeout for a single tool call |
| `AGENTIC_ENABLE_VISION` | true | Enable Vision protocol integration |
| `AGENTIC_ENABLE_MEMORY` | true | Enable HelixMemory (Mem0, Cognee, Letta, Graphiti) |
| `AGENTIC_ENABLE_EXECUTION` | true | Enable Execute mode (false restricts to Reason only) |

## Using Reason Mode

Reason mode is the default. The ensemble debate accesses tools (MCP, LSP, RAG, Embeddings, Vision, HelixMemory) during deliberation but produces no side effects. Ideal for analytical queries, code review, and research.

**Example request:**

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $HELIX_API_KEY" \
  -d '{
    "model": "helixagent/helixagent-debate",
    "messages": [
      {"role": "user", "content": "Analyze the memory leak patterns in our streaming handlers and suggest fixes."}
    ]
  }'
```

The ensemble classifies this as a reasoning task, runs a multi-provider debate with tool access (LSP diagnostics, RAG search over codebase documentation), and returns a synthesized answer. No code is modified.

**What happens internally:**
1. Intent classifier determines Reason mode
2. Multi-provider debate begins with tool integration enabled
3. Each debate phase can invoke up to 5 iterative tool loops
4. Debate converges on a synthesized response
5. Response includes `AgenticMetadata` with `mode: "reason"` and tool invocation summary

## Using Execute Mode

Execute mode decomposes the debate decision into discrete tasks, dispatches them to background agent workers, and verifies results. Suitable for multi-step operations like refactoring, migrations, or automated testing.

**Example request:**

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $HELIX_API_KEY" \
  -d '{
    "model": "helixagent/helixagent-debate",
    "messages": [
      {"role": "user", "content": "Refactor the cache package to use generics and update all call sites."}
    ]
  }'
```

**What happens internally:**
1. Intent classifier determines Execute mode
2. Initial debate produces a plan
3. `ExecutionPlanner` decomposes the plan into tasks with dependencies
4. `AgentWorkerPool` executes tasks layer by layer (parallel within each layer)
5. `VerificationDebate` evaluates results for completeness, correctness, and coherence
6. Final response includes tasks completed, tools invoked, and verification status

To disable Execute mode entirely (restrict to Reason only), set `AGENTIC_ENABLE_EXECUTION=false`.

## Monitoring via SSE Events

When streaming is enabled, the ensemble emits Server-Sent Events for each stage transition:

| Event Type | Payload | When |
|-----------|---------|------|
| `agentic.mode_selected` | `{"mode": "reason"}` or `{"mode": "execute"}` | After intent classification |
| `agentic.tool_invoked` | `{"protocol": "mcp", "tool": "..."}` | Each tool call during debate |
| `agentic.task_started` | `{"task_id": "task-1", "description": "..."}` | Agent begins task execution |
| `agentic.task_completed` | `{"task_id": "task-1", "duration_ms": 1200}` | Agent finishes a task |
| `agentic.verification` | `{"approved": true, "confidence": 0.92}` | Verification debate completes |

Use SSE events to build progress indicators in client applications.

## Troubleshooting

### "provider registry not available"
The ensemble requires at least one verified provider. Check that your `.env` has valid API keys and that startup verification completed successfully. Review `/v1/startup/verification` for provider scores.

### Execute mode returns Reason results
If `AGENTIC_ENABLE_EXECUTION=false` or the intent classifier determines the request does not require execution, the pipeline falls back to Reason mode. Check the `mode` field in the response metadata.

### Individual agent timeouts
When a task exceeds `AGENTIC_AGENT_TIMEOUT`, it fails with a timeout error. Other tasks continue. The verification debate flags the failed task. Increase the timeout or reduce `EstimatedSteps` for complex tasks.

### Verification returns low confidence (0.5)
This indicates the verification LLM was unavailable. The pipeline proceeds with a caveat rather than hard-failing. Check provider health at `/v1/monitoring/status`.

### Circular dependency detected
The `ExecutionPlanner` uses Kahn's algorithm to detect cycles. If the LLM produces tasks with circular dependencies, the pipeline aborts with a descriptive error. Retry or simplify the request.

### Tool protocol unavailable
When a tool client (MCP, LSP, etc.) is not configured, agents receive a descriptive error and continue without that tool. Ensure the required services are running and configured in `.env`.

## Performance Tuning

| Parameter | Guidance |
|-----------|----------|
| `MaxConcurrentAgents` | Increase for I/O-bound tasks (API calls). Keep at 5 or lower for CPU-bound work. |
| `MaxToolIterationsPerPhase` | Reduce to 2-3 for latency-sensitive queries. Increase to 8-10 for deep research. |
| `GlobalTimeout` | Set based on task complexity. Simple queries: 2-5 min. Large refactors: 15-30 min. |
| `ToolIterationTimeout` | Increase if MCP servers or RAG retrieval are slow (default 30s is sufficient for most). |

For resource-limited environments, combine with `GOMAXPROCS` to cap CPU usage:

```bash
GOMAXPROCS=4 AGENTIC_MAX_CONCURRENT_AGENTS=3 ./bin/helixagent
```
