# User Manual 34: Agentic Workflows Guide

## Overview

HelixAgent's Agentic Workflows provide graph-based workflow orchestration for multi-step AI tasks. Define nodes (processing steps), edges (transitions), and let the engine execute them with checkpointing, state management, and conditional branching.

## API Endpoints

### Create and Execute a Workflow

```
POST /v1/agentic/workflows
```

**Request Body:**

```json
{
  "name": "code-review-workflow",
  "nodes": [
    {
      "id": "analyze",
      "type": "llm",
      "config": {"provider": "claude", "model": "claude-3-sonnet"}
    },
    {
      "id": "review",
      "type": "llm",
      "config": {"provider": "deepseek", "model": "deepseek-coder"}
    },
    {
      "id": "summarize",
      "type": "llm",
      "config": {"provider": "gemini", "model": "gemini-pro"}
    }
  ],
  "edges": [
    {"from": "analyze", "to": "review"},
    {"from": "review", "to": "summarize"}
  ],
  "entry_point": "analyze",
  "end_nodes": ["summarize"],
  "input": {
    "code": "func main() { ... }"
  }
}
```

**Response:**

```json
{
  "id": "wf-abc123",
  "name": "code-review-workflow",
  "status": "completed",
  "result": { ... },
  "execution_time_ms": 2450,
  "nodes_executed": 3
}
```

### Get Workflow Status

```
GET /v1/agentic/workflows/:id
```

## Workflow Concepts

### Nodes
Each node is a processing step. Types: `llm` (AI generation), `tool` (tool execution), `conditional` (branching).

### Edges
Connect nodes. Execution follows edge direction from `entry_point` to `end_nodes`.

### Checkpointing
Workflows support `RestoreFromCheckpoint()` for resuming interrupted executions.

### Configuration
Optional `WorkflowConfig` controls timeout, max retries, and parallelism.

## Best Practices

1. Keep workflows under 10 nodes for maintainability
2. Use conditional nodes for error handling branches
3. Set appropriate timeouts per node
4. Monitor via `/v1/agentic/workflows/:id` for long-running workflows
