# LLM Optimization Framework

HelixAgent's LLM Optimization Framework provides a unified set of tools for improving performance, reducing costs, and enhancing the quality of LLM responses.

## Overview

The optimization framework consists of:

| Component | Type | Purpose |
|-----------|------|---------|
| **Semantic Cache** | Native Go | Vector similarity-based response caching |
| **Structured Output** | Native Go | JSON schema validation and generation |
| **Enhanced Streaming** | Native Go | Buffered streaming with progress tracking |
| **SGLang** | HTTP Bridge | RadixAttention prefix caching |
| **LlamaIndex** | HTTP Bridge | Document retrieval with Cognee sync |
| **LangChain** | HTTP Bridge | Task decomposition and agents |
| **Guidance** | HTTP Bridge | CFG/regex constrained generation |
| **LMQL** | HTTP Bridge | Query language constraints |

## Quick Start

### Basic Usage

```go
import "dev.helix.agent/internal/optimization"

// Create service with default config
config := optimization.DefaultConfig()
svc, err := optimization.NewService(config)

// Optimize a request (check cache, retrieve context)
optimized, err := svc.OptimizeRequest(ctx, prompt, embedding)
if optimized.CacheHit {
    return optimized.CachedResponse
}

// Process with LLM...

// Optimize and cache the response
result, err := svc.OptimizeResponse(ctx, response, embedding, query, schema)
```

### Docker Services

```bash
# Start optimization services
docker-compose --profile optimization up -d

# With GPU support (SGLang)
docker-compose --profile optimization-gpu up -d
```

### Configuration

Add to your config file:

```yaml
optimization:
  enabled: true

  semantic_cache:
    enabled: true
    similarity_threshold: 0.85
    max_entries: 10000
    ttl: "24h"

  streaming:
    enabled: true
    buffer_type: "word"

  # External services (disabled by default)
  sglang:
    enabled: false
    endpoint: "http://localhost:30000"
```

## Documentation

- [Semantic Cache Guide](./SEMANTIC_CACHE_GUIDE.md) - Vector-based response caching
- [Structured Output Guide](./STRUCTURED_OUTPUT_GUIDE.md) - JSON schema validation
- [Streaming Guide](./STREAMING_GUIDE.md) - Enhanced streaming capabilities
- [SGLang Integration](./SGLANG_INTEGRATION.md) - Prefix caching setup
- [LlamaIndex + Cognee Guide](./LLAMAINDEX_COGNEE_GUIDE.md) - Document retrieval
- [LangChain Guide](./LANGCHAIN_GUIDE.md) - Task decomposition
- [Guidance & LMQL Guide](./GUIDANCE_LMQL_GUIDE.md) - Constrained generation
- [API Reference](./OPTIMIZATION_API_REFERENCE.md) - Full API documentation

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    OptimizationService                       │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ SemanticCache│  │StructuredGen │  │EnhancedStream│      │
│  │  (GPTCache)  │  │  (Outlines)  │  │(llm-streaming)│     │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   SGLang     │  │  LlamaIndex  │  │  LangChain   │      │
│  │  (HTTP)      │  │   (HTTP)     │  │   (HTTP)     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │   Guidance   │  │    LMQL      │                        │
│  │   (HTTP)     │  │   (HTTP)     │                        │
│  └──────────────┘  └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

## Request Flow

1. **Request Optimization**
   - Check semantic cache for similar queries
   - Retrieve relevant context via LlamaIndex
   - Decompose complex tasks via LangChain
   - Warm prefix cache via SGLang

2. **LLM Processing**
   - Standard provider call with optional streaming
   - Structured output validation via Outlines
   - Constrained generation via Guidance/LMQL

3. **Response Optimization**
   - Validate structured output against schema
   - Cache response with embedding
   - Store patterns in Cognee knowledge graph

## Performance Benefits

| Feature | Benefit |
|---------|---------|
| Semantic Cache | 50-90% latency reduction for similar queries |
| Prefix Caching | 30-50% token savings for multi-turn conversations |
| Context Retrieval | More accurate responses with relevant context |
| Task Decomposition | Better handling of complex multi-step tasks |
| Structured Output | Guaranteed valid JSON responses |

## Service Ports

| Service | Port | Purpose |
|---------|------|---------|
| LangChain | 8011 | Task decomposition, agents |
| LlamaIndex | 8012 | Document retrieval |
| Guidance | 8013 | CFG constraints |
| LMQL | 8014 | Query language |
| SGLang | 30000 | Prefix caching (GPU) |
