# LangChain Package

This package provides an HTTP client for the LangChain service, enabling task decomposition and agent-based workflows.

## Overview

The LangChain client connects to a LangChain service for complex task decomposition, chain-of-thought reasoning, and multi-step agent workflows.

## Features

- **Task Decomposition**: Break complex tasks into subtasks
- **Chain Execution**: Run multi-step processing chains
- **Agent Workflows**: Orchestrate tool-using agents
- **Memory Integration**: Conversation memory support

## Components

### Client (`client.go`)

HTTP client for LangChain service:

```go
config := &langchain.ClientConfig{
    BaseURL: "http://localhost:8011",
    Timeout: 120 * time.Second,
}

client := langchain.NewClient(config)
```

## Data Types

### ClientConfig

```go
type ClientConfig struct {
    BaseURL string        // LangChain server URL
    Timeout time.Duration // Request timeout
}
```

### DecomposeRequest

```go
type DecomposeRequest struct {
    Task     string // Task description
    MaxSteps int    // Maximum subtasks
    Context  string // Additional context
}
```

### Subtask

```go
type Subtask struct {
    ID           int    // Subtask ID
    Description  string // What to do
    Dependencies []int  // Depends on subtask IDs
    Complexity   string // low, medium, high
}
```

## Usage

### Task Decomposition

```go
import "dev.helix.agent/internal/optimization/langchain"

client := langchain.NewClient(nil)

subtasks, err := client.Decompose(ctx, &langchain.DecomposeRequest{
    Task:     "Build a REST API with user authentication",
    MaxSteps: 10,
})

for _, subtask := range subtasks {
    fmt.Printf("%d. %s (deps: %v)\n", subtask.ID, subtask.Description, subtask.Dependencies)
}
```

### Chain Execution

```go
result, err := client.RunChain(ctx, &langchain.ChainRequest{
    ChainType: "summarize_map_reduce",
    Input:     longDocument,
    Config: map[string]interface{}{
        "chunk_size":    1000,
        "chunk_overlap": 100,
    },
})
```

### Agent Workflow

```go
response, err := client.RunAgent(ctx, &langchain.AgentRequest{
    AgentType: "react",
    Task:      "Search for the latest news about AI and summarize it",
    Tools:     []string{"web_search", "summarizer"},
    MaxSteps:  5,
})
```

### With Memory

```go
// Create session with memory
session, _ := client.CreateSession(ctx, &langchain.SessionConfig{
    MemoryType: "buffer_window",
    WindowSize: 10,
})

// Conversation with memory
resp1, _ := client.Chat(ctx, session.ID, "My name is John")
resp2, _ := client.Chat(ctx, session.ID, "What's my name?") // Remembers "John"
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LANGCHAIN_BASE_URL` | LangChain server URL | `http://localhost:8011` |
| `LANGCHAIN_TIMEOUT` | Request timeout | `120s` |

### Server Setup

```bash
# Start LangChain service
python -m langchain.serve --port 8011
```

## Chain Types

| Type | Description |
|------|-------------|
| `summarize` | Document summarization |
| `summarize_map_reduce` | Parallel summarization |
| `qa_retrieval` | Question answering with retrieval |
| `conversation` | Conversational chain |

## Agent Types

| Type | Description |
|------|-------------|
| `react` | ReAct (Reason + Act) agent |
| `structured_chat` | Tool-using chat agent |
| `openai_functions` | OpenAI function calling agent |

## Testing

```bash
go test -v ./internal/optimization/langchain/...
```

## Files

- `client.go` - HTTP client implementation
