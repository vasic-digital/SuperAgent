# Agentic Module User Guide

**Module**: `digital.vasic.agentic`
**Directory**: `Agentic/`
**Phase**: 5 (AI/ML)

## Overview

The Agentic module provides graph-based workflow orchestration for autonomous AI agents. It models
multi-step AI tasks as directed graphs where nodes represent processing steps and edges represent
transitions between steps.

Key capabilities:

- Define workflows as node/edge graphs with typed nodes (agent, tool, condition, parallel, human,
  subgraph)
- Execute workflows with automatic retries, configurable timeouts, and per-node retry policies
- Thread mutable state (`WorkflowState`) through all nodes for cross-step data sharing
- Automatic checkpointing with configurable intervals and state restoration from any checkpoint
- Context-aware execution with `context.Context` propagation for cancellation and deadlines
- Conditional branching via `ConditionFunc` predicates on both edges and nodes

## Installation

```go
import "digital.vasic.agentic/agentic"
```

Add to your `go.mod` (HelixAgent uses a `replace` directive for local development):

```go
require digital.vasic.agentic v0.0.0

replace digital.vasic.agentic => ./Agentic
```

## Key Types and Interfaces

### Workflow

The top-level orchestration unit. Created via `NewWorkflow`.

```go
type Workflow struct {
    ID          string
    Name        string
    Description string
    Graph       *WorkflowGraph
    State       *WorkflowState
    Config      *WorkflowConfig
    Logger      *logrus.Logger
}
```

Methods: `AddNode`, `AddEdge`, `SetEntryPoint`, `AddEndNode`, `Execute`,
`RestoreFromCheckpoint`.

### WorkflowConfig

Controls execution behaviour.

```go
type WorkflowConfig struct {
    MaxIterations        int
    Timeout              time.Duration
    EnableCheckpoints    bool
    CheckpointInterval   int   // checkpoint every N iterations
    EnableSelfCorrection bool
    MaxRetries           int
    RetryDelay           time.Duration
}
```

Use `DefaultWorkflowConfig()` for sensible defaults (100 iterations, 30-minute timeout,
checkpointing enabled, 3 retries).

### Node

A single step in the workflow.

```go
type Node struct {
    ID          string
    Name        string
    Type        NodeType       // agent | tool | condition | parallel | human | subgraph
    Handler     NodeHandler    // func(ctx, state, input) (*NodeOutput, error)
    Condition   ConditionFunc  // optional guard on this node
    Config      map[string]interface{}
    RetryPolicy *RetryPolicy   // override global retry settings
}
```

### NodeHandler

The execution function for a node:

```go
type NodeHandler func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error)
```

### WorkflowState

Mutable state shared across all node executions.

```go
type WorkflowState struct {
    ID          string
    WorkflowID  string
    CurrentNode string
    Messages    []Message
    Variables   map[string]interface{}
    History     []NodeExecution
    Checkpoints []Checkpoint
    Status      WorkflowStatus  // pending | running | paused | completed | failed
    StartTime   time.Time
    EndTime     *time.Time
    Error       error
}
```

### NodeInput / NodeOutput

```go
type NodeInput struct {
    Query    string
    Messages []Message
    Tools    []Tool
    Context  map[string]interface{}
    Previous *NodeOutput
}

type NodeOutput struct {
    Result    interface{}
    Messages  []Message
    ToolCalls []ToolCall
    NextNode  string  // override next node selection
    ShouldEnd bool    // signal workflow termination
    Error     error
    Metadata  map[string]interface{}
}
```

## Usage Examples

### Simple Linear Workflow

```go
package main

import (
    "context"
    "fmt"

    "digital.vasic.agentic/agentic"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    cfg := agentic.DefaultWorkflowConfig()
    wf := agentic.NewWorkflow("code-review", "Review and fix code", cfg, logger)

    // Node 1: Analyse
    analyseNode := &agentic.Node{
        ID:   "analyse",
        Name: "Analyse Code",
        Type: agentic.NodeTypeAgent,
        Handler: func(ctx context.Context, state *agentic.WorkflowState, input *agentic.NodeInput) (*agentic.NodeOutput, error) {
            state.Variables["issues"] = findIssues(input.Query)
            return &agentic.NodeOutput{Result: state.Variables["issues"]}, nil
        },
    }

    // Node 2: Fix
    fixNode := &agentic.Node{
        ID:   "fix",
        Name: "Apply Fix",
        Type: agentic.NodeTypeAgent,
        Handler: func(ctx context.Context, state *agentic.WorkflowState, input *agentic.NodeInput) (*agentic.NodeOutput, error) {
            issues := state.Variables["issues"]
            fixed := applyFixes(issues)
            return &agentic.NodeOutput{Result: fixed, ShouldEnd: true}, nil
        },
    }

    _ = wf.AddNode(analyseNode)
    _ = wf.AddNode(fixNode)
    _ = wf.AddEdge("analyse", "fix", nil, "proceed")
    _ = wf.SetEntryPoint("analyse")
    _ = wf.AddEndNode("fix")

    finalState, err := wf.Execute(context.Background(), &agentic.NodeInput{
        Query: "func main() { fmt.Println(\"hello\") }",
    })
    if err != nil {
        panic(err)
    }
    fmt.Println("Status:", finalState.Status)
}
```

### Conditional Branching

```go
// Route to different fix strategies based on issue type
wf.AddEdge("analyse", "fix-simple", func(state *agentic.WorkflowState) bool {
    issues, _ := state.Variables["issues"].([]string)
    return len(issues) < 3
}, "few issues")

wf.AddEdge("analyse", "fix-deep", func(state *agentic.WorkflowState) bool {
    issues, _ := state.Variables["issues"].([]string)
    return len(issues) >= 3
}, "many issues")
```

### Checkpoint Restore

```go
// After a failure, restore from the most recent checkpoint and re-run
if len(state.Checkpoints) > 0 {
    lastCP := state.Checkpoints[len(state.Checkpoints)-1]
    if err := wf.RestoreFromCheckpoint(state, lastCP.ID); err != nil {
        panic(err)
    }
    state, _ = wf.Execute(context.Background(), &agentic.NodeInput{})
}
```

## Integration with HelixAgent Adapter

HelixAgent wraps the module through `internal/adapters/agentic/adapter.go`.

```go
import agenticadapter "dev.helix.agent/internal/adapters/agentic"

adapter := agenticadapter.New(logger)

// Create a workflow via the adapter (logger injected automatically)
wf := adapter.NewWorkflow("my-task", "Task description", nil)

// Or run a one-shot fire-and-forget workflow
state, err := adapter.ExecuteWorkflow(ctx, "quick-task", map[string]any{
    "input": "some data",
})
```

The adapter's `ExecuteWorkflow` is a convenience method that builds a single pass-through node.
For real workflows with multiple steps, use `adapter.NewWorkflow` and build the graph manually.

## Build and Test

```bash
cd Agentic
go build ./...
go test ./... -count=1 -race
```
