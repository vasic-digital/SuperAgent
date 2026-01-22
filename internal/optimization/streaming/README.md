# Streaming Package

This package provides enhanced streaming capabilities with progress tracking, rate limiting, and buffering for LLM responses.

## Overview

The Streaming package extends basic LLM streaming with features like word-level buffering, progress tracking, rate limiting, and Server-Sent Events (SSE) support.

## Components

### Enhanced Streamer (`enhanced_streamer.go`)

Main streaming orchestrator with progress tracking:

```go
config := &streaming.StreamConfig{
    BufferType:       streaming.BufferTypeWord,
    ProgressInterval: 100 * time.Millisecond,
    RateLimit:        50.0, // tokens per second
    EstimatedTokens:  500,
}

streamer := streaming.NewEnhancedStreamer(config)
output := streamer.StreamWithProgress(ctx, inputStream, progressCallback)
```

### Buffer Types (`buffer.go`)

Content buffering strategies:
- **Character**: Emit each character
- **Word**: Buffer until word boundary
- **Sentence**: Buffer until sentence boundary
- **Token**: Buffer by token count

### Progress Tracking (`progress.go`)

Real-time progress monitoring:

```go
tracker := streaming.NewProgressTracker(estimatedTokens)
tracker.Start()

// Get progress updates
progress := tracker.GetProgress()
fmt.Printf("%.1f%% complete, ETA: %v\n", progress.Percentage, progress.ETA)
```

### Rate Limiter (`rate_limiter.go`)

Token rate limiting:

```go
limiter := streaming.NewRateLimiter(50.0) // 50 tokens/second
throttled := limiter.Throttle(ctx, inputStream)
```

### SSE Support (`sse.go`)

Server-Sent Events formatting:

```go
sseWriter := streaming.NewSSEWriter(httpResponseWriter)
for chunk := range stream {
    sseWriter.WriteEvent("message", chunk.Content)
}
sseWriter.Close()
```

## Data Types

### StreamChunk

```go
type StreamChunk struct {
    Content     string                 // Chunk content
    TokenCount  int                    // Tokens in chunk
    IsComplete  bool                   // Final chunk flag
    FinishReason string               // stop, length, error
    Metadata    map[string]interface{} // Custom metadata
}
```

### StreamProgress

```go
type StreamProgress struct {
    TokensReceived int           // Tokens received so far
    TotalTokens    int           // Estimated total tokens
    Percentage     float64       // Completion percentage
    ElapsedTime    time.Duration // Time since start
    ETA            time.Duration // Estimated time remaining
    TokensPerSec   float64       // Current throughput
}
```

### StreamConfig

```go
type StreamConfig struct {
    BufferType           BufferType    // Buffering strategy
    ProgressInterval     time.Duration // Progress update frequency
    RateLimit            float64       // Max tokens per second
    EstimatedTokens      int           // Expected total tokens
    TokenBufferThreshold int           // Token buffer size
}
```

## Usage

### Basic Streaming with Progress

```go
import "dev.helix.agent/internal/optimization/streaming"

config := streaming.DefaultStreamConfig()
config.EstimatedTokens = 1000

streamer := streaming.NewEnhancedStreamer(config)

// Progress callback
onProgress := func(p *streaming.StreamProgress) {
    fmt.Printf("\rProgress: %.1f%% (%.0f tokens/sec)", p.Percentage, p.TokensPerSec)
}

// Stream with progress
output := streamer.StreamWithProgress(ctx, llmStream, onProgress)
for chunk := range output {
    fmt.Print(chunk.Content)
}
```

### Word-Level Buffering

```go
config := &streaming.StreamConfig{
    BufferType: streaming.BufferTypeWord,
}

buffer := streaming.NewWordBuffer()
for chunk := range rawStream {
    for _, word := range buffer.Add(chunk.Content) {
        fmt.Print(word + " ")
    }
}
```

### Rate-Limited Streaming

```go
limiter := streaming.NewRateLimiter(30.0) // 30 tokens/second max
throttled := limiter.Throttle(ctx, fastStream)

for chunk := range throttled {
    // Chunks arrive at max 30 tokens/second
    process(chunk)
}
```

### SSE Endpoint

```go
func StreamHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")

    sse := streaming.NewSSEWriter(w)
    defer sse.Close()

    for chunk := range llmStream {
        sse.WriteEvent("chunk", chunk.Content)
        w.(http.Flusher).Flush()
    }
    sse.WriteEvent("done", "")
}
```

## Buffer Types

| Type | Description | Use Case |
|------|-------------|----------|
| `BufferTypeCharacter` | No buffering | Real-time typing effect |
| `BufferTypeWord` | Buffer until whitespace | Natural word display |
| `BufferTypeSentence` | Buffer until punctuation | Sentence-by-sentence |
| `BufferTypeToken` | Buffer by token count | Batch processing |

## Testing

```bash
go test -v ./internal/optimization/streaming/...
```

## Files

- `enhanced_streamer.go` - Main streamer implementation
- `buffer.go` - Buffering implementations
- `progress.go` - Progress tracking
- `rate_limiter.go` - Rate limiting
- `sse.go` - Server-Sent Events support
- `chunker.go` - Content chunking utilities
- `aggregator.go` - Stream aggregation
- `types.go` - Type definitions
