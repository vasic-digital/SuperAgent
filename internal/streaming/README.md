# Streaming Package

The streaming package provides response streaming capabilities for HelixAgent.

## Overview

Implements:
- Server-Sent Events (SSE) streaming
- WebSocket streaming
- Token-by-token streaming
- Progress tracking

## Key Components

```go
streamer := streaming.NewEnhancedStreamer(config)

// Stream response
err := streamer.StreamResponse(ctx, writer, func(chunk string) {
    // Process each chunk
})
```

## See Also

- `internal/optimization/streaming/` - Advanced streaming optimization
- `internal/handlers/` - HTTP streaming handlers
