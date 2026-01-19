# LSP (Language Server Protocol) Package

The LSP package provides Language Server Protocol integration with AI-powered code intelligence for HelixAgent.

## Overview

This package implements:

- **LSP-AI Integration**: AI-enhanced code completions, hover info, and diagnostics
- **Document Management**: Efficient document storage and change tracking
- **Symbol Indexing**: Code symbol extraction and search
- **Completion Caching**: Intelligent caching for fast responses

## Key Components

### AI Completion Provider

Integrates with LLM providers for intelligent code completions:

```go
provider := lsp.NewAICompletionProvider(config, llmClient, logger)

completions, err := provider.GetCompletions(ctx, &lsp.CompletionRequest{
    Document:   doc,
    Position:   position,
    TriggerKind: lsp.TriggerCharacter,
})
```

### Document Store

Manages open documents and their state:

```go
store := lsp.NewDocumentStore()

// Open document
store.OpenDocument(uri, content, languageID, version)

// Get document
doc, ok := store.GetDocument(uri)

// Apply changes
store.ApplyChanges(uri, changes, newVersion)

// Close document
store.CloseDocument(uri)
```

### Symbol Index

Indexes and searches code symbols:

```go
index := lsp.NewSymbolIndex()

// Index a document
index.IndexDocument(uri, symbols)

// Search symbols
results := index.Search("MyFunction", 10)

// Get symbols for document
symbols := index.GetDocumentSymbols(uri)
```

### Completion Cache

Caches completions for performance:

```go
cache := lsp.NewCompletionCache(config)

// Get cached completions
if items, ok := cache.Get(cacheKey); ok {
    return items
}

// Cache new completions
cache.Set(cacheKey, items, ttl)
```

## Features

### Code Completions

- Context-aware suggestions
- AI-powered semantic completions
- Snippet support
- Documentation integration

### Hover Information

- Type information
- Documentation display
- AI-generated explanations

### Code Actions

- Quick fixes
- Refactoring suggestions
- AI-assisted code improvements

### Diagnostics

- Syntax error detection
- AI-powered issue detection
- Security vulnerability warnings

## Configuration

```go
type LSPAIConfig struct {
    EnableAICompletions bool
    EnableAIHover       bool
    EnableAICodeActions bool
    EnableAIDiagnostics bool
    MaxCompletions      int
    CompletionTimeout   time.Duration
    CacheTTL            time.Duration
    MinConfidence       float64
}
```

## LSP Message Types

| Method | Description |
|--------|-------------|
| `textDocument/completion` | Code completions |
| `textDocument/hover` | Hover information |
| `textDocument/codeAction` | Code actions |
| `textDocument/diagnostics` | Diagnostics |
| `textDocument/definition` | Go to definition |
| `textDocument/references` | Find references |
| `textDocument/rename` | Symbol rename |

## Usage Example

```go
// Create AI-powered LSP server
config := &lsp.LSPAIConfig{
    EnableAICompletions: true,
    EnableAIHover:       true,
    MaxCompletions:      50,
    CompletionTimeout:   5 * time.Second,
}

server := lsp.NewLSPServer(config, llmClient)

// Handle requests
response, err := server.HandleRequest(ctx, request)
```

## Integration

### With IDE/Editor

1. Start LSP server on designated port
2. Connect from IDE using LSP client
3. Configure AI features as needed

### With HelixAgent

```go
// Register LSP handler
router.POST("/v1/lsp", handlers.LSPHandler(lspServer))

// Handle WebSocket connections
router.GET("/v1/lsp/ws", handlers.LSPWebSocketHandler(lspServer))
```

## Testing

```bash
# Run all LSP tests
go test -v ./internal/lsp/...

# Run AI integration tests
go test -v -run TestAI ./internal/lsp/

# Benchmark completions
go test -bench=BenchmarkCompletion ./internal/lsp/
```

## Performance Considerations

- **Caching**: Enable completion caching for frequently used patterns
- **Timeout**: Configure appropriate timeouts for AI calls
- **Debouncing**: Debounce completion requests in client
- **Indexing**: Pre-index large codebases for fast symbol lookup

## See Also

- [LSP Specification](https://microsoft.github.io/language-server-protocol/)
- `internal/services/lsp_manager.go` - LSP manager service
- `internal/handlers/lsp_handler.go` - HTTP handler
