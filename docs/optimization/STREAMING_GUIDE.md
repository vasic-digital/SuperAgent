# Enhanced Streaming Guide

The enhanced streaming system provides advanced streaming capabilities with buffering, progress tracking, and rate limiting.

## Overview

Enhanced streaming improves the user experience for LLM responses by:

- Buffering tokens into meaningful units (words, sentences, paragraphs)
- Tracking progress during long generations
- Rate limiting output for controlled delivery
- Server-Sent Events (SSE) support for web clients

## Features

- **Multiple Buffer Types**: Character, word, sentence, line, paragraph, token
- **Progress Tracking**: Estimated time and token progress
- **Rate Limiting**: Token bucket and burst rate limiters
- **SSE Writer**: Server-Sent Events for web streaming
- **Chunk Aggregation**: Combine chunks for analysis

## Configuration

```yaml
optimization:
  streaming:
    enabled: true
    buffer_type: "word"           # character, word, sentence, line, paragraph, token
    progress_interval: "100ms"    # Progress update frequency
    rate_limit: 0                 # Tokens per second (0 = unlimited)
```

## Buffer Types

### Character Buffer

Outputs immediately, one character at a time:

```go
buffer := streaming.NewBuffer(streaming.BufferTypeCharacter)
```

### Word Buffer

Buffers until a complete word is available:

```go
buffer := streaming.NewBuffer(streaming.BufferTypeWord)
buffer.Add("Hel")      // returns ""
buffer.Add("lo ")      // returns "Hello "
buffer.Add("world!")   // returns ""
buffer.Flush()         // returns "world!"
```

### Sentence Buffer

Buffers until a complete sentence (ends with `.`, `!`, `?`):

```go
buffer := streaming.NewBuffer(streaming.BufferTypeSentence)
buffer.Add("Hello world")    // returns ""
buffer.Add(". How are")      // returns "Hello world."
buffer.Add(" you?")          // returns ""
buffer.Flush()               // returns "How are you?"
```

### Line Buffer

Buffers until newline:

```go
buffer := streaming.NewBuffer(streaming.BufferTypeLine)
```

### Paragraph Buffer

Buffers until double newline:

```go
buffer := streaming.NewBuffer(streaming.BufferTypeParagraph)
```

### Token Buffer

Buffers N tokens before outputting:

```go
buffer := streaming.NewBufferWithTokenThreshold(streaming.BufferTypeToken, 5)
```

## Basic Usage

### EnhancedStreamer

```go
import "github.com/helixagent/helixagent/internal/optimization/streaming"

config := &streaming.StreamConfig{
    BufferType:       streaming.BufferTypeWord,
    ProgressInterval: 100 * time.Millisecond,
    RateLimit:        0, // unlimited
}

streamer := streaming.NewEnhancedStreamer(config)
```

### Buffered Streaming

```go
// Input channel from LLM
input := make(chan *streaming.StreamChunk)

// Get buffered output channel
output, getResult := streamer.StreamBuffered(ctx, input, streaming.BufferTypeWord)

// Consume buffered output
for chunk := range output {
    fmt.Print(chunk.Content)
}

// Get aggregated result
result := getResult()
fmt.Printf("\nTotal tokens: %d\n", result.TotalTokens)
```

### With Progress Tracking

```go
progressCallback := func(p *streaming.StreamProgress) {
    if p.EstimatedTokens > 0 {
        pct := float64(p.TokensReceived) / float64(p.EstimatedTokens) * 100
        fmt.Printf("\rProgress: %.1f%% (%d/%d tokens)", pct, p.TokensReceived, p.EstimatedTokens)
    }
}

output, getResult := streamer.StreamWithProgress(ctx, input, progressCallback)
```

### With Rate Limiting

```go
// Limit to 50 tokens per second
output, getResult := streamer.StreamWithRateLimit(ctx, input, 50)
```

### Full Enhanced Streaming

Combines buffering, progress, and rate limiting:

```go
output, getResult := streamer.StreamEnhanced(ctx, input, progressCallback)
```

## Progress Tracking

### ProgressTracker

```go
tracker := streaming.NewProgressTracker()
tracker.SetEstimatedTokens(500)

// Update as tokens arrive
tracker.UpdateTokens(10)

// Get current progress
progress := tracker.GetProgress()
fmt.Printf("Received: %d, Elapsed: %v\n", progress.TokensReceived, progress.Elapsed)
```

### Throttled Callbacks

Prevent callback spam with throttling:

```go
throttled := streaming.ThrottledCallback(callback, 100*time.Millisecond)
// callback will be called at most every 100ms
```

## Rate Limiting

### Token Rate Limiter

```go
limiter := streaming.NewRateLimiter(50) // 50 tokens/second

output := limiter.LimitStream(ctx, input)
// or limit individual chunks
limiter.LimitChunks(ctx, input, output)
```

### Burst Rate Limiter

Allows bursts up to a limit:

```go
limiter := streaming.NewBurstRateLimiter(50, 100) // 50/sec, burst of 100

output := limiter.LimitStream(ctx, input)
```

### Dynamic Rate Adjustment

```go
limiter.SetRate(100) // Change to 100 tokens/sec
limiter.Reset()      // Reset token bucket
```

## Server-Sent Events (SSE)

### SSE Writer

```go
writer := streaming.NewSSEWriter(responseWriter)

// Write events
writer.WriteEvent("message", "Hello world")
writer.WriteData("Plain data")
writer.WriteJSON(map[string]string{"key": "value"})
writer.WriteProgress(50, 100) // 50 of 100 tokens
writer.WriteDone()
writer.WriteError(err)
```

### Stream to SSE

```go
// Stream chunks to SSE
streaming.StreamChunksToSSE(ctx, responseWriter, inputChannel)
```

### Format SSE Event

```go
event := streaming.FormatSSEEvent("message", "data", "id-123", 1000)
// Returns: "event: message\ndata: data\nid: id-123\nretry: 1000\n\n"
```

## Chunk Aggregation

### ChunkAggregator

Aggregate chunks and get final result:

```go
aggregator := streaming.NewChunkAggregator()

output, getResult := aggregator.AggregateChunks(ctx, input)

// Consume output
for chunk := range output {
    fmt.Print(chunk.Content)
}

// Get aggregated result
result := getResult()
fmt.Println("Full content:", result.FullContent)
fmt.Println("Total chunks:", result.ChunkCount)
fmt.Println("Total tokens:", result.TotalTokens)
```

### StreamAggregator

For simple string channels:

```go
aggregator := streaming.NewStreamAggregator(estimatedSize)
output := aggregator.Aggregate(ctx, input)

for chunk := range output {
    fmt.Print(chunk)
}

result := aggregator.Result()
```

## Integration with OptimizationService

```go
svc, _ := optimization.NewService(config)

// LLM returns a stream channel
llmStream := provider.CompleteStream(ctx, prompt)

// Enhance the stream
enhancedStream, getResult := svc.StreamEnhanced(ctx, llmStream, progressCallback)

// Consume enhanced stream
for chunk := range enhancedStream {
    fmt.Print(chunk.Content)
}

// Get metrics
result := getResult()
fmt.Printf("\nStreaming complete: %d tokens in %v\n", result.TotalTokens, result.Duration)
```

## StreamChunk Structure

```go
type StreamChunk struct {
    Content   string    // Text content
    Index     int       // Chunk index
    Timestamp time.Time // When chunk was received
    Done      bool      // True for final chunk
    Error     error     // Any error
    Metadata  map[string]interface{} // Additional data
}
```

## Best Practices

1. **Choose Appropriate Buffer**: Use sentence buffer for readable output, word for responsiveness

2. **Set Progress Interval**: Balance update frequency vs overhead (100-500ms recommended)

3. **Use Rate Limiting Sparingly**: Only when necessary for UX or API limits

4. **Handle Context Cancellation**: All operations respect context cancellation

5. **Close Channels**: Ensure input channels are closed to signal completion

6. **Buffer Final Flush**: Always call `Flush()` or wait for `Done` chunk

## Performance Characteristics

| Buffer Type | Latency | Memory | Use Case |
|-------------|---------|--------|----------|
| Character | Lowest | Lowest | Typewriter effect |
| Word | Low | Low | General use |
| Sentence | Medium | Medium | Readable output |
| Line | Medium | Medium | Log-style output |
| Paragraph | High | Higher | Document generation |
| Token | Variable | Low | Batch processing |
