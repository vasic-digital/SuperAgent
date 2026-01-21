# Package: agentic

## Overview

The `agentic` package provides graph-based workflow orchestration for autonomous AI agents with planning, execution, and self-correction capabilities. It implements a LangGraph-style approach to building complex AI workflows.

## Architecture

```
agentic/
├── workflow.go       # Core workflow engine and types
└── workflow_test.go  # Unit tests (96.5% coverage)
```

## Key Types

### Workflow

Represents a graph-based agentic workflow with nodes and edges.

```go
type Workflow struct {
    ID          string
    Name        string
    Description string
    Graph       *WorkflowGraph
    State       *WorkflowState
    Config      *WorkflowConfig
}
```

### Node Types

| Type | Description |
|------|-------------|
| `agent` | LLM-based agent node |
| `tool` | Tool execution node |
| `condition` | Conditional branching |
| `parallel` | Parallel execution |
| `human` | Human-in-the-loop |
| `subgraph` | Nested workflow |

## Usage

### Basic Workflow

```go
import "dev.helix.agent/internal/agentic"

// Create workflow builder
builder := agentic.NewWorkflowBuilder("my-workflow")

// Add nodes
builder.AddNode("agent", agentic.NodeTypeAgent, agentHandler)
builder.AddNode("tool", agentic.NodeTypeTool, toolHandler)

// Add edges
builder.AddEdge("agent", "tool")
builder.AddConditionalEdge("tool", "agent", shouldRetry)

// Build and run
workflow := builder.Build()
result, err := workflow.Run(ctx, input)
```

### Checkpointing

```go
// Enable checkpointing for fault tolerance
config := &agentic.WorkflowConfig{
    EnableCheckpointing: true,
    CheckpointInterval:  time.Minute,
}

workflow := builder.WithConfig(config).Build()

// Resume from checkpoint
workflow.ResumeFrom(ctx, checkpointID)
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| MaxIterations | int | 100 | Maximum workflow iterations |
| Timeout | time.Duration | 5m | Overall workflow timeout |
| EnableCheckpointing | bool | false | Enable state checkpointing |
| CheckpointInterval | time.Duration | 30s | Checkpoint frequency |

## Testing

```bash
go test -v ./internal/agentic/...
go test -cover ./internal/agentic/...  # 96.5% coverage
```

## Dependencies

### Internal
- `internal/llm` - LLM provider integration
- `internal/tools` - Tool execution

### External
- `github.com/google/uuid` - Unique IDs
- `github.com/sirupsen/logrus` - Logging

## See Also

- [LangGraph Documentation](https://langchain-ai.github.io/langgraph/)
- [Workflow API Reference](../../docs/api/workflows.md)
