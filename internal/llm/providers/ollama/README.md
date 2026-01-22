# Ollama Provider

This package implements the LLM provider interface for locally-hosted Ollama models.

## Overview

The Ollama provider enables HelixAgent to communicate with local Ollama instances, providing access to open-source models like Llama, Mistral, and CodeLlama without external API dependencies.

## Status

**DEPRECATED**: Ollama is assigned a base score of 5.0 and is only used as a fallback provider when other providers are unavailable.

## Supported Models

Any model available in your local Ollama installation:

| Model | Size | Description |
|-------|------|-------------|
| llama3.2 | 3B/11B/90B | Meta's latest Llama |
| codellama | 7B/13B/34B | Code-specialized Llama |
| mistral | 7B | Mistral 7B |
| deepseek-coder | 6.7B/33B | DeepSeek Coder |
| phi3 | 3.8B/14B | Microsoft Phi-3 |

## Prerequisites

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model
ollama pull llama3.2
```

## Configuration

```yaml
providers:
  ollama:
    enabled: true
    base_url: "http://localhost:11434"
    default_model: "llama3.2"
    timeout_seconds: 300
```

No API key required for local instances.

## Features

- **Local Execution**: No external API calls
- **Privacy**: Data stays on your machine
- **Streaming**: Real-time response streaming
- **Custom Models**: Support for GGUF/GGML formats

## Usage

```go
import "dev.helix.agent/internal/llm/providers/ollama"

provider := ollama.NewOllamaProvider(config)
response, err := provider.Complete(ctx, request)
```

## Limitations

- Slower than cloud providers
- Requires local GPU for best performance
- Limited context window compared to cloud models
- No tool calling support in most models

## Testing

```bash
go test -v ./internal/llm/providers/ollama/...
```

## Files

- `ollama.go` - Main provider implementation
- `ollama_test.go` - Unit tests
