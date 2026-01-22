# Context Package

This package provides context window management for LLM interactions, handling token limits, message history, and automatic summarization.

## Overview

The Context package manages the limited context window of LLMs by tracking token counts, prioritizing messages, and automatically evicting or summarizing content when limits are reached.

## Features

- **Token Management**: Track and limit context tokens
- **Priority-Based Eviction**: Evict low-priority content first
- **Automatic Summarization**: Condense old messages
- **Pinned Entries**: Protect critical context
- **Chunking**: Split large content into manageable pieces

## Components

### Context Window (`window.go`)

Main context management:

```go
config := &context.WindowConfig{
    MaxTokens:        8000,
    ReserveTokens:    1000, // For response
    EvictionStrategy: context.EvictionLRU,
}

window := context.NewContextWindow(config)
```

### Chunking (`chunking.go`)

Content chunking utilities:

```go
chunker := context.NewChunker(&context.ChunkerConfig{
    ChunkSize:    500,
    ChunkOverlap: 50,
    SplitBy:      context.SplitBySentence,
})

chunks := chunker.Chunk(longDocument)
```

### Summarizer (`summarizer.go`)

Automatic context summarization:

```go
summarizer := context.NewSummarizer(llmProvider, &context.SummarizerConfig{
    MaxSummaryTokens: 200,
    SummarizeThreshold: 0.8, // Summarize when 80% full
})
```

## Data Types

### ContextEntry

```go
type ContextEntry struct {
    ID         string                 // Unique identifier
    Role       string                 // system, user, assistant, tool
    Content    string                 // Entry content
    TokenCount int                    // Tokens in entry
    Timestamp  time.Time              // When added
    Priority   Priority               // Eviction priority
    Metadata   map[string]interface{} // Custom metadata
    Pinned     bool                   // Protected from eviction
}
```

### WindowConfig

```go
type WindowConfig struct {
    MaxTokens        int              // Maximum context tokens
    ReserveTokens    int              // Reserved for response
    EvictionStrategy EvictionStrategy // How to evict
    SummarizeOld     bool             // Auto-summarize old entries
    TokenCounter     TokenCounter     // Token counting function
}
```

### Priority Levels

```go
const (
    PriorityLow      Priority = 0 // Evicted first
    PriorityNormal   Priority = 1 // Default priority
    PriorityHigh     Priority = 2 // Protected longer
    PriorityCritical Priority = 3 // Evicted last
)
```

## Usage

### Basic Context Management

```go
import "dev.helix.agent/internal/optimization/context"

window := context.NewContextWindow(&context.WindowConfig{
    MaxTokens:     4096,
    ReserveTokens: 500,
})

// Add system prompt (pinned)
window.AddPinned("system", "You are a helpful assistant", context.PriorityCritical)

// Add conversation
window.Add("user", "What is Go?", context.PriorityNormal)
window.Add("assistant", "Go is a programming language...", context.PriorityNormal)
window.Add("user", "Tell me more", context.PriorityNormal)

// Get messages for API call
messages := window.GetMessages()
```

### Automatic Eviction

```go
window := context.NewContextWindow(&context.WindowConfig{
    MaxTokens:        4096,
    EvictionStrategy: context.EvictionLRU,
})

// Add many messages...
for _, msg := range conversation {
    err := window.Add(msg.Role, msg.Content, context.PriorityNormal)
    if err == context.ErrContextOverflow {
        // Old messages automatically evicted
    }
}
```

### With Summarization

```go
window := context.NewContextWindow(&context.WindowConfig{
    MaxTokens:    4096,
    SummarizeOld: true,
})

summarizer := context.NewSummarizer(llmProvider, nil)
window.SetSummarizer(summarizer)

// When context is 80% full, old messages are summarized
window.Add("user", longMessage, context.PriorityNormal)
```

### Content Chunking

```go
chunker := context.NewChunker(&context.ChunkerConfig{
    ChunkSize:    1000,
    ChunkOverlap: 100,
    SplitBy:      context.SplitByParagraph,
})

chunks := chunker.Chunk(longDocument)
for i, chunk := range chunks {
    fmt.Printf("Chunk %d (%d tokens): %s...\n", i, chunk.TokenCount, chunk.Content[:50])
}
```

### Priority-Based Management

```go
// System prompts - never evict
window.AddPinned("system", systemPrompt, context.PriorityCritical)

// Recent messages - high priority
window.Add("user", currentQuestion, context.PriorityHigh)

// Old context - low priority (evict first)
window.Add("assistant", oldResponse, context.PriorityLow)
```

## Eviction Strategies

| Strategy | Description |
|----------|-------------|
| `EvictionLRU` | Least Recently Used |
| `EvictionLFU` | Least Frequently Used |
| `EvictionFIFO` | First In First Out |
| `EvictionPriority` | Lowest priority first |

## Testing

```bash
go test -v ./internal/optimization/context/...
```

## Files

- `window.go` - Context window management
- `chunking.go` - Content chunking utilities
- `summarizer.go` - Automatic summarization
