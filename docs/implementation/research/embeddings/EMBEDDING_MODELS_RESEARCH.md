# Embedding Models Research Documentation

## Status: RESEARCHED
**Date**: 2026-01-19

---

## 1. Current Implementation (EXISTING)

### OpenAI Embeddings

**Location**: `internal/services/embedding_manager.go`

| Model | Dimensions | Max Tokens | Cost |
|-------|------------|------------|------|
| text-embedding-ada-002 | 1536 | 8191 | $0.0001/1K tokens |
| text-embedding-3-small | 1536 | 8191 | $0.00002/1K tokens |
| text-embedding-3-large | 3072 | 8191 | $0.00013/1K tokens |

### Local Hash Fallback

**Features**:
- SHA256-based deterministic embeddings
- Configurable dimensions (384-3072)
- Always available (no API dependency)
- Vector normalization

---

## 2. Open Source Embedding Models

### High Priority (Recommended)

#### BGE-M3

**Repository**: BAAI/bge-m3 (Hugging Face)
**License**: MIT

**Features**:
- Multi-functionality: dense + sparse + multi-vector
- Multilingual: 100+ languages
- Long context: 8192 tokens
- High benchmark performance

**Installation**:
```python
from sentence_transformers import SentenceTransformer
model = SentenceTransformer('BAAI/bge-m3')
embeddings = model.encode(texts)
```

**Dimensions**: 1024 (default)

#### all-mpnet-base-v2

**Repository**: sentence-transformers/all-mpnet-base-v2
**License**: MIT

**Features**:
- Best quality for English sentences
- Fast inference
- Well-tested across tasks

**Installation**:
```python
from sentence_transformers import SentenceTransformer
model = SentenceTransformer('all-mpnet-base-v2')
```

**Dimensions**: 768

#### Nomic Embed Text V2

**Repository**: nomic-ai/nomic-embed-text-v2
**License**: Apache 2.0

**Features**:
- Matryoshka embeddings (adjustable dimensions)
- Task-aware embeddings
- 8192 context length

**Installation**:
```python
from sentence_transformers import SentenceTransformer
model = SentenceTransformer('nomic-ai/nomic-embed-text-v2', trust_remote_code=True)
```

**Dimensions**: 64, 128, 256, 512, 768 (selectable)

### Medium Priority

#### Qwen3-Embedding-0.6B

**Repository**: Qwen/Qwen3-Embedding-0.6B
**License**: Apache 2.0

**Features**:
- Multilingual
- Instruction-aware
- Flexible dimensions (32-1024)
- Lightweight (0.6B parameters)

**Installation**:
```python
from transformers import AutoModel, AutoTokenizer
model = AutoModel.from_pretrained("Qwen/Qwen3-Embedding-0.6B")
```

#### EmbeddingGemma-300M

**Repository**: google/embedding-gemma-300m
**License**: Apache 2.0

**Features**:
- On-device optimized
- Multilingual
- 300M parameters

#### gte-multilingual-base

**Repository**: Alibaba-NLP/gte-multilingual-base
**License**: MIT

**Features**:
- General-purpose multilingual
- Good for cross-lingual retrieval

### Specialized Models

#### Jina Embeddings v4

**Repository**: jinaai/jina-embeddings-v4
**License**: CC-BY-NC-4.0 (non-commercial)

**Features**:
- Universal multimodal
- Dense + multi-vector
- Multilingual

**Note**: Non-commercial license restricts production use

#### CodeBERT

**Repository**: microsoft/codebert-base
**License**: MIT

**Features**:
- Code-specific embeddings
- Multi-language programming support
- Good for code search

---

## 3. Local Embedding Runtimes

### sentence-transformers (Python)

**Installation**:
```bash
pip install sentence-transformers
```

**Usage**:
```python
from sentence_transformers import SentenceTransformer

model = SentenceTransformer('all-mpnet-base-v2')
embeddings = model.encode(["text 1", "text 2"])
```

**Features**:
- Easy model loading
- Batch processing
- GPU acceleration
- 100+ pre-trained models

### Ollama Embeddings

**Installation**:
```bash
ollama pull nomic-embed-text
```

**Usage** (API):
```bash
curl http://localhost:11434/api/embeddings \
  -d '{"model": "nomic-embed-text", "prompt": "text to embed"}'
```

**Go Client**:
```go
resp, err := ollamaClient.Embeddings(ctx, &ollama.EmbeddingRequest{
    Model:  "nomic-embed-text",
    Prompt: "text to embed",
})
```

**Supported Models**:
- nomic-embed-text
- mxbai-embed-large
- all-minilm
- bge-m3

### HuggingFace Inference API

**Endpoint**: `https://api-inference.huggingface.co/models/{model_id}`

**Usage**:
```python
import requests

API_URL = "https://api-inference.huggingface.co/models/sentence-transformers/all-mpnet-base-v2"
headers = {"Authorization": f"Bearer {HF_TOKEN}"}

response = requests.post(API_URL, headers=headers, json={"inputs": ["text"]})
embeddings = response.json()
```

**Free Tier**: Rate limited, good for development

### ONNX Runtime

**Installation**:
```bash
pip install onnxruntime  # or onnxruntime-gpu
```

**Usage**:
```python
import onnxruntime as ort

session = ort.InferenceSession("model.onnx")
outputs = session.run(None, {"input_ids": input_ids, "attention_mask": attention_mask})
```

**Benefits**:
- Cross-platform
- Optimized inference
- Quantization support
- CPU/GPU support

---

## 4. Embedding Model Registry Design

### Registry Interface

```go
// internal/embeddings/models/registry.go

type EmbeddingModel interface {
    // Core operations
    Encode(ctx context.Context, texts []string) ([][]float32, error)
    EncodeSingle(ctx context.Context, text string) ([]float32, error)

    // Model info
    Name() string
    Dimensions() int
    MaxTokens() int
    Provider() string

    // Health
    Health(ctx context.Context) error
    Close() error
}

type EmbeddingModelConfig struct {
    Name          string
    Provider      string  // "openai", "ollama", "sentence-transformers", "local"
    ModelID       string
    Dimensions    int
    MaxTokens     int
    BatchSize     int
    Timeout       time.Duration
    CacheEnabled  bool
    CacheTTL      time.Duration
    GPUEnabled    bool
}

type EmbeddingModelRegistry struct {
    models      map[string]EmbeddingModel
    defaultModel string
    mu          sync.RWMutex
}

func (r *EmbeddingModelRegistry) Register(name string, model EmbeddingModel) error
func (r *EmbeddingModelRegistry) Get(name string) (EmbeddingModel, error)
func (r *EmbeddingModelRegistry) GetDefault() (EmbeddingModel, error)
func (r *EmbeddingModelRegistry) List() []string
func (r *EmbeddingModelRegistry) Health(ctx context.Context) map[string]error
```

### Provider Implementations

```go
// OpenAI Provider
type OpenAIEmbeddingModel struct {
    client    *openai.Client
    modelID   string
    dims      int
    maxTokens int
}

// Ollama Provider
type OllamaEmbeddingModel struct {
    client  *ollama.Client
    modelID string
    dims    int
}

// Sentence Transformers (via Python subprocess or gRPC)
type SentenceTransformersModel struct {
    processPool *ProcessPool
    modelID     string
    dims        int
}

// Local Hash Fallback
type LocalHashModel struct {
    dims int
}
```

---

## 5. Integration with HelixAgent

### Configuration

```yaml
# configs/embeddings.yaml
embeddings:
  default_model: "openai-3-small"

  models:
    openai-3-small:
      provider: openai
      model_id: text-embedding-3-small
      dimensions: 1536
      max_tokens: 8191
      batch_size: 100
      timeout: 30s
      cache_enabled: true
      cache_ttl: 1h

    openai-3-large:
      provider: openai
      model_id: text-embedding-3-large
      dimensions: 3072
      max_tokens: 8191

    bge-m3:
      provider: ollama
      model_id: bge-m3
      dimensions: 1024
      max_tokens: 8192
      gpu_enabled: true

    all-mpnet-base-v2:
      provider: sentence-transformers
      model_id: all-mpnet-base-v2
      dimensions: 768

    local-fallback:
      provider: local
      dimensions: 1536

  fallback_chain:
    - openai-3-small
    - bge-m3
    - all-mpnet-base-v2
    - local-fallback
```

### Lazy Initialization

```go
func (r *EmbeddingModelRegistry) GetOrCreate(name string) (EmbeddingModel, error) {
    r.mu.RLock()
    if model, ok := r.models[name]; ok {
        r.mu.RUnlock()
        return model, nil
    }
    r.mu.RUnlock()

    // Lazy initialization
    r.mu.Lock()
    defer r.mu.Unlock()

    // Double-check after acquiring write lock
    if model, ok := r.models[name]; ok {
        return model, nil
    }

    // Create model based on config
    config := r.configs[name]
    model, err := r.createModel(config)
    if err != nil {
        return nil, err
    }

    r.models[name] = model
    return model, nil
}
```

### Fallback Chain

```go
func (r *EmbeddingModelRegistry) EncodeWithFallback(ctx context.Context, texts []string) ([][]float32, string, error) {
    for _, modelName := range r.fallbackChain {
        model, err := r.GetOrCreate(modelName)
        if err != nil {
            continue
        }

        embeddings, err := model.Encode(ctx, texts)
        if err != nil {
            log.Printf("Model %s failed: %v, trying next", modelName, err)
            continue
        }

        return embeddings, modelName, nil
    }

    return nil, "", fmt.Errorf("all embedding models failed")
}
```

---

## 6. Docker Compose Stack

### embedding-models-stack.yml

```yaml
version: '3.8'

services:
  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama_models:/root/.ollama
    deploy:
      resources:
        reservations:
          devices:
            - capabilities: [gpu]
    healthcheck:
      test: ["CMD", "ollama", "list"]
      interval: 30s
      timeout: 10s
      retries: 3

  sentence-transformers:
    build:
      context: ./docker/embeddings/sentence-transformers
      dockerfile: Dockerfile
    ports:
      - "8014:8014"
    environment:
      - TRANSFORMERS_CACHE=/models
    volumes:
      - st_models:/models
    deploy:
      resources:
        limits:
          memory: 4G
        reservations:
          devices:
            - capabilities: [gpu]

volumes:
  ollama_models:
  st_models:
```

### Dockerfile for sentence-transformers service

```dockerfile
# docker/embeddings/sentence-transformers/Dockerfile
FROM python:3.11-slim

WORKDIR /app

RUN pip install --no-cache-dir \
    sentence-transformers \
    flask \
    gunicorn \
    torch --index-url https://download.pytorch.org/whl/cpu

COPY server.py .

# Pre-download models
RUN python -c "from sentence_transformers import SentenceTransformer; \
    SentenceTransformer('all-mpnet-base-v2'); \
    SentenceTransformer('BAAI/bge-m3')"

EXPOSE 8014

CMD ["gunicorn", "-w", "4", "-b", "0.0.0.0:8014", "server:app"]
```

---

## 7. Performance Benchmarks

### Target Metrics

| Model | Batch Size | Target Latency | Target Throughput |
|-------|------------|----------------|-------------------|
| OpenAI 3-small | 100 | < 500ms | 2000 texts/s |
| OpenAI 3-large | 100 | < 800ms | 1500 texts/s |
| BGE-M3 (GPU) | 32 | < 200ms | 500 texts/s |
| all-mpnet (GPU) | 32 | < 150ms | 600 texts/s |
| Local hash | 1000 | < 50ms | 50000 texts/s |

### Optimization Strategies

1. **Batch Processing**: Process texts in optimal batch sizes
2. **Caching**: Cache embeddings with TTL
3. **GPU Acceleration**: Use CUDA when available
4. **Connection Pooling**: Reuse HTTP connections
5. **Quantization**: Use int8 quantized models for faster inference

---

## 8. Testing Requirements

### Unit Tests

```go
func TestEmbeddingModelRegistry_Register(t *testing.T) { ... }
func TestEmbeddingModelRegistry_Get(t *testing.T) { ... }
func TestEmbeddingModelRegistry_EncodeWithFallback(t *testing.T) { ... }
func TestOpenAIEmbeddingModel_Encode(t *testing.T) { ... }
func TestOllamaEmbeddingModel_Encode(t *testing.T) { ... }
func TestLocalHashModel_Encode(t *testing.T) { ... }
```

### Integration Tests

```go
func TestEmbeddingPipeline_EndToEnd(t *testing.T) { ... }
func TestEmbeddingFallback_AllModels(t *testing.T) { ... }
func TestEmbeddingCaching(t *testing.T) { ... }
func TestEmbeddingBatchProcessing(t *testing.T) { ... }
```

### Challenge Script

```bash
#!/bin/bash
# challenges/scripts/embedding_models_challenge.sh

set -euo pipefail

echo "Testing Embedding Models Infrastructure..."

# Test OpenAI embeddings
curl -s http://localhost:8080/v1/embeddings/generate \
  -H "Content-Type: application/json" \
  -d '{"text": "test", "model": "text-embedding-3-small"}' | jq .

# Test Ollama embeddings
curl -s http://localhost:11434/api/embeddings \
  -d '{"model": "nomic-embed-text", "prompt": "test"}' | jq .

# Test sentence-transformers
curl -s http://localhost:8014/encode \
  -H "Content-Type: application/json" \
  -d '{"texts": ["test"]}' | jq .

# Test fallback chain
curl -s http://localhost:8080/v1/embeddings/generate \
  -H "Content-Type: application/json" \
  -d '{"text": "test", "use_fallback": true}' | jq .

echo "All embedding model tests passed!"
```

---

## 9. Security Considerations

1. **API Key Management**: Store keys in secure vault
2. **Rate Limiting**: Respect provider limits
3. **Data Privacy**: Don't log embedding content
4. **Model Isolation**: Run models in sandboxed containers
5. **Input Validation**: Sanitize input texts
