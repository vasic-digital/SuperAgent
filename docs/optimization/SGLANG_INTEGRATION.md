# SGLang Integration Guide

SGLang provides RadixAttention for efficient prefix caching in multi-turn conversations.

## Overview

SGLang (Structured Generation Language) implements RadixAttention, which enables:

- **Prefix Caching**: Reuse KV-cache for common prefixes
- **Multi-Turn Efficiency**: 30-50% token savings in conversations
- **Session Management**: Maintain conversation context efficiently

## Prerequisites

- NVIDIA GPU with CUDA support (recommended: 24GB+ VRAM)
- Docker with NVIDIA Container Toolkit
- SGLang Docker image

## Docker Setup

### Start SGLang Service

```bash
# Start with optimization profile
docker-compose --profile optimization-gpu up -d sglang

# Or run directly
docker run -d --gpus all \
  -p 30000:30000 \
  --name sglang \
  lmsysorg/sglang:latest \
  python -m sglang.launch_server \
  --model-path meta-llama/Llama-2-7b-chat-hf \
  --port 30000
```

### Verify Service

```bash
curl http://localhost:30000/health
# Should return: {"status": "healthy"}
```

## Configuration

```yaml
optimization:
  sglang:
    enabled: true
    endpoint: "http://localhost:30000"
    timeout: "120s"
    fallback_on_unavailable: true
```

## Basic Usage

### Client Initialization

```go
import "dev.helix.agent/internal/optimization/sglang"

config := &sglang.ClientConfig{
    BaseURL: "http://localhost:30000",
    Timeout: 120 * time.Second,
}

client := sglang.NewClient(config)
```

### Health Check

```go
health, err := client.Health(ctx)
if err != nil {
    log.Fatal("SGLang unavailable:", err)
}
fmt.Println("Status:", health.Status)
```

### Simple Completion

```go
response, err := client.CompleteSimple(ctx, "What is machine learning?")
fmt.Println(response)
```

### Completion with System Prompt

```go
response, err := client.CompleteWithSystem(ctx,
    "You are a helpful assistant.",
    "Explain quantum computing.",
)
```

### Full Completion Request

```go
resp, err := client.Complete(ctx, &sglang.CompletionRequest{
    Messages: []sglang.Message{
        {Role: "system", Content: "You are an expert programmer."},
        {Role: "user", Content: "Write a Python function to sort a list."},
    },
    Temperature: 0.7,
    MaxTokens:   500,
    TopP:        0.9,
})

if len(resp.Choices) > 0 {
    fmt.Println(resp.Choices[0].Message.Content)
}
```

## Session Management

SGLang's power comes from session-based prefix caching.

### Create Session

```go
session, err := client.CreateSession(ctx, "session-123", "You are a helpful coding assistant.")
// System prompt is cached for reuse
```

### Continue Session

```go
// First turn
response1, err := client.ContinueSession(ctx, "session-123", "What is Python?")
// Prefix cache is used

// Second turn - prefix is already cached!
response2, err := client.ContinueSession(ctx, "session-123", "How do I install it?")
// Much faster due to prefix caching
```

### Delete Session

```go
err := client.DeleteSession(ctx, "session-123")
```

### Cleanup Stale Sessions

```go
// Remove sessions unused for 1 hour
removed := client.CleanupSessions(ctx, 1 * time.Hour)
fmt.Printf("Removed %d stale sessions\n", removed)
```

## Prefix Caching

### Warm Prefix

Pre-cache common prefixes for faster responses:

```go
resp, err := client.WarmPrefix(ctx, "You are an expert in machine learning and AI.")
if resp.Cached {
    fmt.Printf("Prefix cached, ~%d tokens\n", resp.TokenSize)
}
```

### Warm Multiple Prefixes

```go
prefixes := []string{
    "You are a helpful coding assistant.",
    "You are an expert in data science.",
    "You are a creative writing assistant.",
}

err := client.WarmPrefixes(ctx, prefixes)
// All prefixes are warmed in parallel
```

## Integration with OptimizationService

```go
config := optimization.DefaultConfig()
config.SGLang.Enabled = true
config.SGLang.Endpoint = "http://localhost:30000"

svc, err := optimization.NewService(config)

// Create a conversation session
err = svc.CreateSession(ctx, "user-123", "You are a helpful assistant.")

// Continue the conversation
response, err := svc.ContinueSession(ctx, "user-123", "Hello!")
```

## Request Flow with Prefix Caching

```
Request 1: System prompt + "What is ML?"
┌─────────────────────────────────────────┐
│ System Prompt │ User Query │ Response   │
│   (Cached)    │  (New)     │   (Gen)    │
└─────────────────────────────────────────┘

Request 2: System prompt + History + "Tell me more"
┌─────────────────────────────────────────────────────┐
│ System + History │ New Query │ Response            │
│    (Cached!)     │   (New)   │   (Gen)             │
└─────────────────────────────────────────────────────┘
         ↑
   Prefix Reused - Faster!
```

## Performance Benefits

| Scenario | Without Caching | With SGLang |
|----------|-----------------|-------------|
| First message | 100% tokens | 100% tokens |
| Second message | 100% tokens | ~60% tokens |
| Fifth message | 100% tokens | ~30% tokens |
| System prompt reuse | Each time | Once |

## GPU Memory Considerations

| Model Size | Min VRAM | Recommended |
|------------|----------|-------------|
| 7B params | 16GB | 24GB |
| 13B params | 32GB | 40GB |
| 70B params | 80GB+ | 160GB+ |

## Troubleshooting

### Service Not Starting

```bash
# Check logs
docker logs sglang

# Verify GPU access
docker run --gpus all nvidia/cuda:11.0-base nvidia-smi
```

### Out of Memory

```bash
# Reduce model size or use quantization
docker run --gpus all lmsysorg/sglang:latest \
  python -m sglang.launch_server \
  --model-path meta-llama/Llama-2-7b-chat-hf \
  --quantization awq
```

### Connection Refused

```go
// Check if service is available
if !client.IsAvailable(ctx) {
    // Use fallback
    config.SGLang.Enabled = false
}
```

## Best Practices

1. **Warm Common Prefixes**: Pre-cache frequently used system prompts at startup

2. **Use Sessions for Conversations**: Leverage session management for multi-turn chats

3. **Cleanup Stale Sessions**: Periodically remove unused sessions to free memory

4. **Enable Fallback**: Set `fallback_on_unavailable: true` for graceful degradation

5. **Monitor GPU Memory**: Watch memory usage to prevent OOM errors

## API Reference

### Endpoints Used

| Endpoint | Method | Purpose |
|----------|--------|---------|
| /health | GET | Health check |
| /v1/chat/completions | POST | Chat completion |

### Response Structure

```go
type CompletionResponse struct {
    ID      string             `json:"id"`
    Object  string             `json:"object"`
    Created int64              `json:"created"`
    Model   string             `json:"model"`
    Choices []CompletionChoice `json:"choices"`
    Usage   Usage              `json:"usage"`
}

type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```
