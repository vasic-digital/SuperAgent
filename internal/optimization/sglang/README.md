# SGLang Package

This package provides an HTTP client for the SGLang service, enabling efficient prefix caching with RadixAttention.

## Overview

SGLang (Structured Generation Language) is an LLM serving framework that provides RadixAttention for efficient prefix caching, significantly reducing latency for multi-turn conversations and repeated prompts.

## Features

- **RadixAttention**: Efficient prefix caching for repeated prompts
- **Session Management**: Multi-turn conversation handling
- **Batch Inference**: Parallel request processing
- **Streaming Support**: Real-time response streaming

## Components

### Client (`client.go`)

HTTP client for SGLang service:

```go
config := &sglang.ClientConfig{
    BaseURL: "http://localhost:30000",
    Timeout: 120 * time.Second,
}

client := sglang.NewClient(config)
```

### Session Management

Multi-turn conversation sessions:

```go
session := client.CreateSession("system-prompt")

response1, _ := client.Chat(ctx, session.ID, "Hello!")
response2, _ := client.Chat(ctx, session.ID, "Tell me more")
// Prefix cache reused automatically
```

## Data Types

### ClientConfig

```go
type ClientConfig struct {
    BaseURL string        // SGLang server URL
    Timeout time.Duration // Request timeout
}
```

### Session

```go
type Session struct {
    ID           string    // Unique session ID
    SystemPrompt string    // System prompt for session
    History      []Message // Conversation history
    CreatedAt    time.Time // Session creation time
    LastUsedAt   time.Time // Last activity timestamp
}
```

### Message

```go
type Message struct {
    Role    string // user, assistant, system
    Content string // Message content
}
```

## Usage

### Basic Completion

```go
import "dev.helix.agent/internal/optimization/sglang"

client := sglang.NewClient(nil) // Uses defaults

response, err := client.Complete(ctx, &sglang.CompletionRequest{
    Model:       "meta-llama/Meta-Llama-3.1-8B-Instruct",
    Prompt:      "Explain quantum computing in simple terms",
    MaxTokens:   500,
    Temperature: 0.7,
})
```

### Multi-Turn Conversation

```go
// Create session with system prompt
session := client.CreateSession("You are a helpful coding assistant")

// First turn - full prompt processing
resp1, _ := client.Chat(ctx, session.ID, "Write a function to sort a list")

// Second turn - prefix cached
resp2, _ := client.Chat(ctx, session.ID, "Now add error handling")

// Third turn - more prefix reuse
resp3, _ := client.Chat(ctx, session.ID, "Add documentation")
```

### Batch Processing

```go
requests := []*sglang.CompletionRequest{
    {Prompt: "Question 1", MaxTokens: 100},
    {Prompt: "Question 2", MaxTokens: 100},
    {Prompt: "Question 3", MaxTokens: 100},
}

responses, _ := client.BatchComplete(ctx, requests)
```

### Streaming

```go
stream, _ := client.StreamComplete(ctx, &sglang.CompletionRequest{
    Prompt:    "Tell me a story",
    MaxTokens: 1000,
    Stream:    true,
})

for chunk := range stream {
    fmt.Print(chunk.Content)
}
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SGLANG_BASE_URL` | SGLang server URL | `http://localhost:30000` |
| `SGLANG_TIMEOUT` | Request timeout | `120s` |

### Server Setup

```bash
# Start SGLang server with prefix caching
python -m sglang.launch_server \
    --model-path meta-llama/Meta-Llama-3.1-8B-Instruct \
    --port 30000 \
    --enable-radix-cache
```

## Performance Benefits

| Scenario | Without Cache | With RadixAttention |
|----------|---------------|---------------------|
| First request | 1.0x | 1.0x |
| Similar prompts | 1.0x | 0.3-0.5x |
| Multi-turn (5+ turns) | 1.0x | 0.1-0.2x |

## Testing

```bash
go test -v ./internal/optimization/sglang/...
```

## Files

- `client.go` - HTTP client implementation
