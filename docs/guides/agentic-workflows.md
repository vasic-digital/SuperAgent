# Agentic Workflows Guide

## Overview

HelixAgent provides a powerful agentic workflow system for orchestrating complex, multi-step AI tasks with checkpointing, branching, and automatic recovery.

## Core Concepts

### Workflow

A workflow is a directed graph of nodes representing AI tasks. Each node can:
- Execute LLM completions
- Make decisions based on outputs
- Branch to different paths
- Loop until conditions are met
- Call external tools

### Node Types

| Type | Description | Use Case |
|------|-------------|----------|
| `completion` | LLM completion | Generate text, answer questions |
| `decision` | Conditional branching | Route based on output |
| `tool` | External tool call | API calls, database queries |
| `parallel` | Parallel execution | Multiple tasks simultaneously |
| `loop` | Iteration | Repeat until condition |
| `checkpoint` | Save state | Recovery point |
| `human` | Human-in-the-loop | Manual approval |

## Creating Workflows

### Basic Workflow

```go
package main

import (
    "context"
    "dev.helix.agent/internal/agentic"
)

func main() {
    // Create workflow
    wf := agentic.NewWorkflow("research-workflow")

    // Add nodes
    wf.AddNode(&agentic.Node{
        ID:   "analyze",
        Type: agentic.NodeTypeCompletion,
        Config: agentic.NodeConfig{
            Provider: "claude",
            Prompt:   "Analyze the following topic: {{.input.topic}}",
        },
    })

    wf.AddNode(&agentic.Node{
        ID:   "synthesize",
        Type: agentic.NodeTypeCompletion,
        Config: agentic.NodeConfig{
            Provider: "gemini",
            Prompt:   "Synthesize the analysis: {{.analyze.output}}",
        },
    })

    // Connect nodes
    wf.AddEdge("analyze", "synthesize")

    // Execute
    ctx := context.Background()
    result, err := wf.Execute(ctx, map[string]interface{}{
        "topic": "Artificial Intelligence in Healthcare",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Result:", result.Output)
}
```

### Workflow with Decisions

```go
wf := agentic.NewWorkflow("classification-workflow")

// Classification node
wf.AddNode(&agentic.Node{
    ID:   "classify",
    Type: agentic.NodeTypeCompletion,
    Config: agentic.NodeConfig{
        Provider: "claude",
        Prompt:   "Classify this text into: positive, negative, neutral\nText: {{.input.text}}",
        OutputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "sentiment": map[string]string{"type": "string"},
            },
        },
    },
})

// Decision node
wf.AddNode(&agentic.Node{
    ID:   "route",
    Type: agentic.NodeTypeDecision,
    Config: agentic.NodeConfig{
        Conditions: []agentic.Condition{
            {
                Expression: "classify.output.sentiment == 'positive'",
                Target:     "positive_handler",
            },
            {
                Expression: "classify.output.sentiment == 'negative'",
                Target:     "negative_handler",
            },
            {
                Expression: "true", // default
                Target:     "neutral_handler",
            },
        },
    },
})

// Handler nodes
wf.AddNode(&agentic.Node{
    ID:   "positive_handler",
    Type: agentic.NodeTypeCompletion,
    Config: agentic.NodeConfig{
        Prompt: "Generate a thank you response for positive feedback: {{.input.text}}",
    },
})

wf.AddNode(&agentic.Node{
    ID:   "negative_handler",
    Type: agentic.NodeTypeCompletion,
    Config: agentic.NodeConfig{
        Prompt: "Generate an empathetic response for negative feedback: {{.input.text}}",
    },
})

wf.AddNode(&agentic.Node{
    ID:   "neutral_handler",
    Type: agentic.NodeTypeCompletion,
    Config: agentic.NodeConfig{
        Prompt: "Generate an informative response: {{.input.text}}",
    },
})

// Connect
wf.AddEdge("classify", "route")
// Decision node handles routing to handlers
```

### Parallel Execution

```go
wf := agentic.NewWorkflow("parallel-research")

// Parallel node executes multiple completions simultaneously
wf.AddNode(&agentic.Node{
    ID:   "research",
    Type: agentic.NodeTypeParallel,
    Config: agentic.NodeConfig{
        Branches: []agentic.Branch{
            {
                ID: "technical",
                Nodes: []agentic.Node{
                    {
                        ID:   "tech_analysis",
                        Type: agentic.NodeTypeCompletion,
                        Config: agentic.NodeConfig{
                            Prompt: "Analyze technical aspects: {{.input.topic}}",
                        },
                    },
                },
            },
            {
                ID: "business",
                Nodes: []agentic.Node{
                    {
                        ID:   "biz_analysis",
                        Type: agentic.NodeTypeCompletion,
                        Config: agentic.NodeConfig{
                            Prompt: "Analyze business implications: {{.input.topic}}",
                        },
                    },
                },
            },
            {
                ID: "ethical",
                Nodes: []agentic.Node{
                    {
                        ID:   "ethics_analysis",
                        Type: agentic.NodeTypeCompletion,
                        Config: agentic.NodeConfig{
                            Prompt: "Analyze ethical considerations: {{.input.topic}}",
                        },
                    },
                },
            },
        },
        JoinStrategy: agentic.JoinAll, // Wait for all branches
    },
})

// Synthesis node combines parallel results
wf.AddNode(&agentic.Node{
    ID:   "synthesize",
    Type: agentic.NodeTypeCompletion,
    Config: agentic.NodeConfig{
        Prompt: `Synthesize the following analyses:
Technical: {{.research.technical.tech_analysis.output}}
Business: {{.research.business.biz_analysis.output}}
Ethical: {{.research.ethical.ethics_analysis.output}}`,
    },
})

wf.AddEdge("research", "synthesize")
```

### Loop Execution

```go
wf := agentic.NewWorkflow("iterative-refinement")

// Initial draft
wf.AddNode(&agentic.Node{
    ID:   "draft",
    Type: agentic.NodeTypeCompletion,
    Config: agentic.NodeConfig{
        Prompt: "Write a draft about: {{.input.topic}}",
    },
})

// Review loop
wf.AddNode(&agentic.Node{
    ID:   "review_loop",
    Type: agentic.NodeTypeLoop,
    Config: agentic.NodeConfig{
        MaxIterations: 3,
        ExitCondition: "review.output.quality_score >= 0.9",
        LoopNodes: []agentic.Node{
            {
                ID:   "review",
                Type: agentic.NodeTypeCompletion,
                Config: agentic.NodeConfig{
                    Prompt: `Review this draft and provide feedback with a quality_score (0-1):
Draft: {{.current_draft}}`,
                    OutputSchema: map[string]interface{}{
                        "type": "object",
                        "properties": map[string]interface{}{
                            "feedback":      map[string]string{"type": "string"},
                            "quality_score": map[string]string{"type": "number"},
                        },
                    },
                },
            },
            {
                ID:   "revise",
                Type: agentic.NodeTypeCompletion,
                Config: agentic.NodeConfig{
                    Prompt: `Revise the draft based on feedback:
Draft: {{.current_draft}}
Feedback: {{.review.output.feedback}}`,
                },
            },
        },
        LoopVariable: "current_draft",
        InitialValue: "{{.draft.output}}",
    },
})

wf.AddEdge("draft", "review_loop")
```

## Tool Integration

### Built-in Tools

```go
wf.AddNode(&agentic.Node{
    ID:   "search",
    Type: agentic.NodeTypeTool,
    Config: agentic.NodeConfig{
        Tool: "web_search",
        Parameters: map[string]interface{}{
            "query": "{{.input.search_query}}",
            "limit": 5,
        },
    },
})
```

### Custom Tools

```go
// Register custom tool
agentic.RegisterTool("database_query", func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    query := params["query"].(string)
    // Execute database query
    rows, err := db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    // Process and return results
    return processRows(rows), nil
})

// Use in workflow
wf.AddNode(&agentic.Node{
    ID:   "fetch_data",
    Type: agentic.NodeTypeTool,
    Config: agentic.NodeConfig{
        Tool: "database_query",
        Parameters: map[string]interface{}{
            "query": "SELECT * FROM users WHERE status = 'active'",
        },
    },
})
```

## Checkpointing & Recovery

### Automatic Checkpoints

```go
wf := agentic.NewWorkflow("long-running-task")
wf.SetCheckpointConfig(agentic.CheckpointConfig{
    Enabled:   true,
    Interval:  time.Minute,
    Storage:   "postgres", // or "redis", "file"
    Retention: 24 * time.Hour,
})

// Nodes automatically checkpoint after completion
wf.AddNode(&agentic.Node{
    ID:   "expensive_computation",
    Type: agentic.NodeTypeCompletion,
    Config: agentic.NodeConfig{
        Checkpoint: true, // Explicit checkpoint after this node
    },
})
```

### Manual Checkpoints

```go
wf.AddNode(&agentic.Node{
    ID:   "checkpoint_1",
    Type: agentic.NodeTypeCheckpoint,
    Config: agentic.NodeConfig{
        Name: "after_initial_processing",
        Data: []string{"analyze.output", "enrich.output"},
    },
})
```

### Resuming from Checkpoint

```go
// Resume failed workflow
executor := agentic.NewExecutor(wf)

// Find latest checkpoint
checkpoint, err := executor.GetLatestCheckpoint(ctx, workflowRunID)
if err != nil {
    log.Fatal(err)
}

// Resume from checkpoint
result, err := executor.ResumeFromCheckpoint(ctx, checkpoint)
if err != nil {
    log.Fatal(err)
}
```

## Human-in-the-Loop

### Approval Gates

```go
wf.AddNode(&agentic.Node{
    ID:   "human_review",
    Type: agentic.NodeTypeHuman,
    Config: agentic.NodeConfig{
        Action:  agentic.HumanActionApprove,
        Timeout: 24 * time.Hour,
        Prompt:  "Please review the generated content:\n\n{{.generate.output}}",
        Options: []string{"approve", "reject", "modify"},
        Webhook: "https://api.example.com/webhooks/approval",
    },
})

// Decision based on human input
wf.AddNode(&agentic.Node{
    ID:   "process_decision",
    Type: agentic.NodeTypeDecision,
    Config: agentic.NodeConfig{
        Conditions: []agentic.Condition{
            {Expression: "human_review.decision == 'approve'", Target: "publish"},
            {Expression: "human_review.decision == 'modify'", Target: "revise"},
            {Expression: "true", Target: "discard"},
        },
    },
})
```

## Error Handling

### Retry Configuration

```go
wf.AddNode(&agentic.Node{
    ID:   "api_call",
    Type: agentic.NodeTypeTool,
    Config: agentic.NodeConfig{
        Tool: "external_api",
        Retry: agentic.RetryConfig{
            MaxAttempts: 3,
            Delay:       time.Second,
            Backoff:     2.0, // Exponential backoff
            RetryOn:     []string{"timeout", "rate_limit"},
        },
    },
})
```

### Fallback Nodes

```go
wf.AddNode(&agentic.Node{
    ID:   "primary",
    Type: agentic.NodeTypeCompletion,
    Config: agentic.NodeConfig{
        Provider: "claude",
        Fallback: &agentic.Node{
            ID:   "fallback",
            Type: agentic.NodeTypeCompletion,
            Config: agentic.NodeConfig{
                Provider: "gemini",
            },
        },
    },
})
```

### Error Handlers

```go
wf.SetErrorHandler(func(ctx context.Context, node *agentic.Node, err error) agentic.ErrorAction {
    // Log error
    log.Printf("Node %s failed: %v", node.ID, err)

    // Decide action
    if isRetryable(err) {
        return agentic.ErrorActionRetry
    }
    if node.Config.Fallback != nil {
        return agentic.ErrorActionFallback
    }
    return agentic.ErrorActionAbort
})
```

## Monitoring & Observability

### Execution Tracing

```go
// Enable tracing
wf.SetTracingConfig(agentic.TracingConfig{
    Enabled:  true,
    Exporter: "jaeger",
    Endpoint: "http://localhost:14268/api/traces",
})

// Node execution generates spans
// workflow:execute
//   └── node:analyze
//   └── node:synthesize
```

### Metrics

```go
// Enable metrics
wf.SetMetricsConfig(agentic.MetricsConfig{
    Enabled:  true,
    Exporter: "prometheus",
})

// Available metrics:
// - workflow_executions_total
// - workflow_duration_seconds
// - node_executions_total
// - node_duration_seconds
// - node_errors_total
```

### Event Streaming

```go
// Subscribe to workflow events
wf.OnEvent(func(event agentic.Event) {
    switch event.Type {
    case agentic.EventNodeStarted:
        log.Printf("Node %s started", event.NodeID)
    case agentic.EventNodeCompleted:
        log.Printf("Node %s completed in %v", event.NodeID, event.Duration)
    case agentic.EventWorkflowCompleted:
        log.Printf("Workflow completed: %v", event.Result)
    }
})
```

## API Endpoints

### Create Workflow

```bash
POST /v1/workflows
Content-Type: application/json

{
  "name": "my-workflow",
  "nodes": [...],
  "edges": [...]
}
```

### Execute Workflow

```bash
POST /v1/workflows/{id}/execute
Content-Type: application/json

{
  "input": {
    "topic": "AI in Healthcare"
  }
}
```

### Get Execution Status

```bash
GET /v1/workflows/executions/{run_id}

Response:
{
  "id": "exec-123",
  "workflow_id": "wf-456",
  "status": "running",
  "current_node": "analyze",
  "progress": 0.45,
  "started_at": "2026-01-23T10:00:00Z",
  "nodes": {
    "analyze": {"status": "completed", "duration_ms": 2340},
    "synthesize": {"status": "running"}
  }
}
```

### Stream Execution Events

```bash
GET /v1/workflows/executions/{run_id}/events
Accept: text/event-stream

event: node_started
data: {"node_id": "analyze", "timestamp": "2026-01-23T10:00:00Z"}

event: node_completed
data: {"node_id": "analyze", "duration_ms": 2340, "output": {...}}
```

## Best Practices

### 1. Design for Failure

- Add checkpoints at critical points
- Configure appropriate retries
- Provide fallback nodes for critical operations

### 2. Optimize Token Usage

- Use smaller models for simple tasks
- Cache intermediate results
- Batch similar operations

### 3. Monitor Performance

- Enable tracing for debugging
- Track execution times
- Set up alerts for failures

### 4. Version Workflows

- Use semantic versioning
- Store workflow definitions in version control
- Test changes before deploying

---

**Document Version**: 1.0
**Last Updated**: January 23, 2026
**Author**: Generated by Claude Code
