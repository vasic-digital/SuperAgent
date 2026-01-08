# LangChain Integration Guide

LangChain provides task decomposition, chain execution, and ReAct agents.

## Overview

LangChain integration enables:

- **Task Decomposition**: Break complex tasks into manageable subtasks
- **Chain Execution**: Run predefined processing chains
- **ReAct Agents**: Reasoning and acting with tool use
- **Text Processing**: Summarization and transformation

## Docker Setup

```bash
# Start LangChain service
docker-compose --profile optimization up -d langchain-server

# Verify
curl http://localhost:8011/health
```

## Configuration

```yaml
optimization:
  langchain:
    enabled: true
    endpoint: "http://localhost:8011"
    timeout: "120s"
    default_chain: "react"
```

## Basic Usage

### Client Initialization

```go
import "dev.helix.agent/internal/optimization/langchain"

config := &langchain.ClientConfig{
    BaseURL: "http://localhost:8011",
    Timeout: 120 * time.Second,
}

client := langchain.NewClient(config)
```

## Task Decomposition

Break complex tasks into subtasks:

```go
response, err := client.Decompose(ctx, &langchain.DecomposeRequest{
    Task:     "Build a REST API with user authentication and database integration",
    MaxSteps: 5,
})

for i, subtask := range response.Subtasks {
    fmt.Printf("%d. %s\n", i+1, subtask.Description)
    fmt.Printf("   Dependencies: %v\n", subtask.Dependencies)
}
```

### Response Structure

```go
type DecomposeResponse struct {
    Subtasks []Subtask `json:"subtasks"`
    Reasoning string   `json:"reasoning"`
}

type Subtask struct {
    ID           string   `json:"id"`
    Description  string   `json:"description"`
    Dependencies []string `json:"dependencies"`
    Complexity   string   `json:"complexity"` // low, medium, high
}
```

## Chain Execution

Execute predefined processing chains:

```go
response, err := client.ExecuteChain(ctx, &langchain.ChainRequest{
    ChainType: "summarize",
    Input:     longDocument,
    Config: map[string]interface{}{
        "max_length": 200,
        "style":      "bullet_points",
    },
})

fmt.Println("Summary:", response.Output)
```

### Available Chains

| Chain Type | Purpose | Config Options |
|------------|---------|----------------|
| summarize | Text summarization | max_length, style |
| translate | Language translation | target_language |
| extract | Information extraction | fields, format |
| qa | Question answering | context |
| transform | Text transformation | instructions |

## ReAct Agents

Run reasoning and acting agents with tools:

```go
response, err := client.RunReActAgent(ctx, &langchain.ReActRequest{
    Goal:           "Find the current weather in New York and summarize it",
    AvailableTools: []string{"search", "calculator", "weather"},
    MaxIterations:  5,
})

fmt.Println("Final answer:", response.FinalAnswer)

// See the reasoning process
for _, step := range response.Steps {
    fmt.Printf("Thought: %s\n", step.Thought)
    fmt.Printf("Action: %s(%s)\n", step.Action, step.ActionInput)
    fmt.Printf("Observation: %s\n\n", step.Observation)
}
```

### Response Structure

```go
type ReActResponse struct {
    FinalAnswer string      `json:"final_answer"`
    Steps       []ReActStep `json:"steps"`
    TotalSteps  int         `json:"total_steps"`
}

type ReActStep struct {
    Thought     string `json:"thought"`
    Action      string `json:"action"`
    ActionInput string `json:"action_input"`
    Observation string `json:"observation"`
}
```

### Available Tools

| Tool | Purpose |
|------|---------|
| search | Web search |
| calculator | Mathematical operations |
| weather | Weather lookup |
| code_executor | Run code snippets |
| file_reader | Read file contents |

## Text Processing

### Summarization

```go
summary, err := client.Summarize(ctx, &langchain.SummarizeRequest{
    Text:      longDocument,
    MaxLength: 200,
    Style:     "concise", // concise, detailed, bullet_points
})
```

### Transformation

```go
transformed, err := client.Transform(ctx, &langchain.TransformRequest{
    Text:         inputText,
    Instructions: "Convert to formal business language",
})
```

## Integration with OptimizationService

```go
config := optimization.DefaultConfig()
config.LangChain.Enabled = true

svc, err := optimization.NewService(config)

// Decompose a complex task
response, err := svc.DecomposeTask(ctx, "Build a machine learning pipeline")

// Run a ReAct agent
agentResp, err := svc.RunReActAgent(ctx, "Research and summarize AI trends", []string{"search"})
```

### Automatic Task Decomposition

The optimization service can automatically decompose complex tasks:

```go
optimized, err := svc.OptimizeRequest(ctx, complexPrompt, embedding)

if len(optimized.DecomposedTasks) > 0 {
    fmt.Println("Task decomposed into:")
    for _, task := range optimized.DecomposedTasks {
        fmt.Println("-", task)
    }
}
```

## Complex Task Detection

Tasks are automatically classified as complex if they:

- Are longer than 500 characters
- Contain phrases like "step by step", "implement", "build", "create"
- Involve multiple distinct operations

```go
// This will trigger decomposition
prompt := "Create a REST API with user authentication. Implement JWT tokens,
           password hashing, and rate limiting. Add database integration with
           PostgreSQL and include comprehensive error handling."

optimized, err := svc.OptimizeRequest(ctx, prompt, nil)
// optimized.DecomposedTasks contains subtasks
```

## Best Practices

1. **Set MaxSteps**: Limit decomposition depth to prevent over-fragmentation

2. **Choose Appropriate Chains**: Match chain type to your processing needs

3. **Limit Agent Iterations**: Set `MaxIterations` to prevent infinite loops

4. **Provide Clear Goals**: Specific goals lead to better agent performance

5. **Use Available Tools**: Only include tools the agent can actually use

## Performance Tuning

| Feature | Latency | Best For |
|---------|---------|----------|
| Decompose | Low | Complex planning |
| Execute Chain | Medium | Text processing |
| ReAct Agent | High | Multi-step tasks |
| Summarize | Medium | Long documents |
| Transform | Low | Simple conversions |

## Error Handling

```go
response, err := client.Decompose(ctx, request)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "timeout"):
        // Retry with longer timeout
    case strings.Contains(err.Error(), "too complex"):
        // Simplify the task
    default:
        log.Error("Decomposition failed:", err)
    }
}
```

## Troubleshooting

### Agent Stuck in Loop

```go
// Set iteration limit
request := &langchain.ReActRequest{
    Goal:          goal,
    MaxIterations: 5, // Prevent infinite loops
}
```

### Poor Decomposition Quality

```go
// Provide more context
request := &langchain.DecomposeRequest{
    Task:     task,
    MaxSteps: 7, // Allow more granularity
    Context:  "This is for a production web application",
}
```

### Service Unavailable

```go
if !client.IsAvailable(ctx) {
    // Skip decomposition, process task directly
    log.Warn("LangChain unavailable, processing without decomposition")
}
```

## API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| /health | GET | Health check |
| /decompose | POST | Task decomposition |
| /chain | POST | Chain execution |
| /react | POST | ReAct agent |
| /summarize | POST | Text summarization |
| /transform | POST | Text transformation |
