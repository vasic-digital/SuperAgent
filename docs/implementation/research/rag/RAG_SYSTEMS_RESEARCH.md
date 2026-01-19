# RAG Systems Research Documentation

## Status: RESEARCHED
**Date**: 2026-01-19

---

## 1. Vector Databases

### PostgreSQL pgvector (EXISTING)

**Status**: Production
**Features**:
- ACID transactions
- SQL compatibility
- Cosine similarity index
- JSONB metadata
- Full-text search integration

**Current Integration**: `internal/services/embedding_manager.go`

### ChromaDB

**Repository**: https://github.com/chroma-core/chroma
**License**: Apache 2.0

**Installation**:
```bash
pip install chromadb
chroma run --path /data/chroma
```

**Features**:
- In-memory or persistent
- Default: all-minilm-l6-v2 embeddings
- OpenAI, Cohere embedding support
- Metadata filtering with operators
- LangChain/LlamaIndex native integration

**API**:
```python
client = chromadb.Client()
collection = client.create_collection("docs")
collection.add(documents=["..."], embeddings=[[...]], ids=["id1"])
results = collection.query(query_embeddings=[[...]], n_results=10)
```

**Docker**:
```yaml
chromadb:
  image: chromadb/chroma:latest
  ports:
    - "8000:8000"
  volumes:
    - chroma_data:/chroma/chroma
```

### Qdrant

**Repository**: https://github.com/qdrant/qdrant
**License**: Apache 2.0

**Installation**:
```bash
docker run -p 6333:6333 qdrant/qdrant
pip install qdrant-client
```

**Features**:
- REST + gRPC APIs
- JSON payload filtering
- Vector quantization (97% RAM savings)
- Horizontal scaling with sharding
- SIMD acceleration
- Write-ahead logging

**Current Integration**: `internal/vectordb/qdrant/client.go`

### Weaviate

**Repository**: https://github.com/weaviate/weaviate
**License**: BSD-3-Clause

**Installation**:
```bash
docker run -p 8080:8080 semitechnologies/weaviate:latest
pip install weaviate-client
```

**Features**:
- Hybrid search (vector + BM25)
- GraphQL API
- Module system for embeddings
- Multi-tenancy
- Horizontal scaling

**API**:
```python
import weaviate
client = weaviate.Client("http://localhost:8080")
client.schema.create_class({
    "class": "Document",
    "vectorizer": "text2vec-openai"
})
```

### Pinecone

**Type**: Managed cloud service
**Free Tier**: 1 GB storage, 2M writes/month

**Features**:
- Serverless deployment
- Auto-scaling
- Metadata filtering
- Namespace isolation

### MongoDB Atlas Vector Search

**Type**: Managed cloud service
**Free Tier**: 500 MB (M0 cluster)

**Features**:
- Vector + document storage
- Aggregation pipeline
- Atlas Search integration

### FAISS

**Repository**: https://github.com/facebookresearch/faiss
**License**: MIT

**Installation**:
```bash
pip install faiss-cpu  # or faiss-gpu
```

**Features**:
- CPU/GPU optimized
- Multiple index types (Flat, IVF, HNSW, PQ)
- Billion-scale support
- No persistence (wrap with storage)

---

## 2. Managed RAG Services

### Ragie

**Type**: RAG-as-a-Service
**Free Tier**: Developer tools

**Features**:
- Full document ingestion
- Real-time indexing
- Retrieval with citations
- Agent support

### Pinecone Assistant

**Type**: Full RAG API
**Free Tier**: First assistant free

**Features**:
- Document upload
- Q&A with grounding
- Built-in generation

### Vectara

**Type**: Full RAG platform
**Free Tier**: 30 days, 10,000 credits

**Features**:
- Data processing pipeline
- Chunking and embedding
- LLM integration
- Enterprise security

### Cohere RAG

**Type**: LLM with RAG
**Free Tier**: Limited API calls

**Features**:
- Pass documents to Chat API
- Grounded generation
- Multi-lingual support

---

## 3. RAG Frameworks

### LlamaIndex (EXISTING)

**Current Integration**: `internal/optimization/llamaindex/client.go`

**Features**:
- HyDE (hypothetical documents)
- Query decomposition
- Step-back prompting
- Document reranking
- Query fusion (RRF)

### LangChain (EXISTING)

**Current Integration**: `internal/optimization/langchain/client.go`

**Features**:
- Task decomposition
- Chain execution
- ReAct agent
- Text summarization
- Custom transformations

### RAGatouille (ColBERT)

**Repository**: https://github.com/bclavie/RAGatouille
**License**: Apache 2.0

**Installation**:
```bash
pip install ragatouille
```

**Features**:
- ColBERT late-interaction retrieval
- Token-level matching (better than dense embeddings)
- Superior domain generalization
- Data-efficient training
- Multilingual support

**Usage**:
```python
from ragatouille import RAGPretrainedModel

RAG = RAGPretrainedModel.from_pretrained("colbert-ir/colbertv2.0")
index_path = RAG.index(index_name="my_index", collection=documents)
results = RAG.search(query, k=10)
```

**Integrations**:
- LangChain (via Vespa)
- LlamaIndex
- FastRAG
- Flask server

---

## 4. Embedding Models

### OpenAI Models (EXISTING)

| Model | Dimensions | Max Tokens |
|-------|------------|------------|
| text-embedding-ada-002 | 1536 | 8191 |
| text-embedding-3-small | 1536 | 8191 |
| text-embedding-3-large | 3072 | 8191 |

### Open Source Models

| Model | Dimensions | Features | License |
|-------|------------|----------|---------|
| Qwen3-Embedding-0.6B | 32-1024 | Multilingual, instruction-aware | Apache 2.0 |
| EmbeddingGemma-300M | Variable | On-device optimized | Apache 2.0 |
| Jina Embeddings v4 | Variable | Multimodal, multilingual | CC-BY-NC |
| BGE-M3 | 8192 tokens | Dense + sparse + multi-vector | MIT |
| all-mpnet-base-v2 | 768 | Sentence-transformers | MIT |
| gte-multilingual-base | Variable | General-purpose | MIT |
| Nomic Embed Text V2 | Matryoshka | Adjustable dimensionality | Apache 2.0 |

### Local Embedding Runtimes

1. **sentence-transformers**: Python library
2. **Ollama**: Local via API
3. **HuggingFace Inference**: Cloud API
4. **ONNX Runtime**: Optimized local

---

## 5. Advanced RAG Techniques

### Query Enhancement

| Technique | Description | Status |
|-----------|-------------|--------|
| HyDE | Generate hypothetical answer, use for retrieval | EXISTING |
| Multi-Query | Expand to multiple related queries | NOT_STARTED |
| Step-Back | Abstract to conceptual level | EXISTING |
| Query Decomposition | Break into sub-queries | EXISTING |
| Query Fusion (RRF) | Combine results from variations | EXISTING |

### Retrieval Enhancement

| Technique | Description | Status |
|-----------|-------------|--------|
| LLM Reranking | Cross-encoder rescoring | EXISTING |
| ColBERT Reranking | Token-level matching | NOT_STARTED |
| Contextual Compression | Reduce retrieved context | NOT_STARTED |
| Parent Document | Retrieve parent chunks | NOT_STARTED |
| Self-Query | Auto-generate filters | NOT_STARTED |

### Graph-Based RAG

| Technique | Description | Status |
|-----------|-------------|--------|
| Knowledge Graph RAG | Entity traversal | EXISTING (Cognee) |
| Microsoft GraphRAG | Community detection | NOT_STARTED |
| Entity Extraction | NER + linking | PARTIAL |

---

## 6. Public Code RAG Datasets

### Large-Scale Repositories

| Dataset | Size | Languages | Access |
|---------|------|-----------|--------|
| codeparrot/github-code | ~1 TB | 32 | Hugging Face |
| bigcode/the-stack | 6 TB | 358 | Hugging Face |
| bigcode/the-stack-v2 | Multi-TB | 600+ | Hugging Face |

### Q&A Datasets

| Dataset | Size | Format | Access |
|---------|------|--------|--------|
| Stack Overflow | Millions | Q&A pairs | Kaggle/BigQuery |
| Stack Exchange | Hundreds GB | XML dump | Archive.org |
| CodeSearchNet | ~3.5 GB | (comment, code) pairs | GitHub/HuggingFace |

---

## 7. Integration Architecture

### Unified Vector Interface

```go
type VectorDatabase interface {
    // Collection management
    CreateCollection(ctx context.Context, name string, config CollectionConfig) error
    DeleteCollection(ctx context.Context, name string) error

    // Document operations
    Upsert(ctx context.Context, collection string, docs []Document) error
    Delete(ctx context.Context, collection string, ids []string) error

    // Search operations
    Search(ctx context.Context, collection string, query SearchQuery) ([]SearchResult, error)
    HybridSearch(ctx context.Context, collection string, query HybridQuery) ([]SearchResult, error)

    // Health
    Health(ctx context.Context) error
}
```

### RAG Pipeline Interface

```go
type RAGPipeline interface {
    // Query transformation
    TransformQuery(ctx context.Context, query string, technique string) ([]string, error)

    // Retrieval
    Retrieve(ctx context.Context, queries []string, options RetrievalOptions) ([]Document, error)

    // Reranking
    Rerank(ctx context.Context, query string, docs []Document) ([]Document, error)

    // Generation
    Generate(ctx context.Context, query string, context []Document) (string, error)
}
```

---

## 8. Docker Compose Stack

### rag-vector-stack.yml

```yaml
version: '3.8'

services:
  postgres:
    image: pgvector/pgvector:pg16
    environment:
      POSTGRES_DB: helixagent_db
      POSTGRES_USER: helixagent
      POSTGRES_PASSWORD: helixagent123
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  chromadb:
    image: chromadb/chroma:latest
    ports:
      - "8000:8000"
    volumes:
      - chroma_data:/chroma/chroma

  qdrant:
    image: qdrant/qdrant:latest
    ports:
      - "6333:6333"
      - "6334:6334"
    volumes:
      - qdrant_data:/qdrant/storage

  weaviate:
    image: semitechnologies/weaviate:latest
    ports:
      - "8080:8080"
    environment:
      QUERY_DEFAULTS_LIMIT: 25
      AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: 'true'
      PERSISTENCE_DATA_PATH: '/var/lib/weaviate'
    volumes:
      - weaviate_data:/var/lib/weaviate

volumes:
  postgres_data:
  chroma_data:
  qdrant_data:
  weaviate_data:
```

### rag-services-stack.yml

```yaml
version: '3.8'

services:
  llamaindex:
    build: ./docker/rag/llamaindex
    ports:
      - "8012:8012"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}

  langchain:
    build: ./docker/rag/langchain
    ports:
      - "8011:8011"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}

  ragatouille:
    build: ./docker/rag/ragatouille
    ports:
      - "8013:8013"
    volumes:
      - ragatouille_index:/app/indexes
    deploy:
      resources:
        reservations:
          devices:
            - capabilities: [gpu]

  cognee:
    image: cognee/cognee:latest
    ports:
      - "7100:7100"
    environment:
      - COGNEE_AUTH_EMAIL=admin@helixagent.ai
      - COGNEE_AUTH_PASSWORD=HelixAgentPass123

volumes:
  ragatouille_index:
```

---

## 9. Testing Requirements

### Unit Tests
- Vector database connections
- Embedding generation
- Search operations
- RAG pipeline stages

### Integration Tests
- End-to-end RAG flow
- Multi-database failover
- Embedding model switching
- Context window management

### Challenge Scripts

```bash
# rag_vector_challenge.sh
#!/bin/bash
set -euo pipefail

echo "Testing RAG Vector Infrastructure..."

# Test PostgreSQL pgvector
psql -h localhost -U helixagent -d helixagent_db \
  -c "SELECT * FROM pg_extension WHERE extname = 'vector';"

# Test ChromaDB
curl -s http://localhost:8000/api/v1/heartbeat | jq .

# Test Qdrant
curl -s http://localhost:6333/collections | jq .

# Test Weaviate
curl -s http://localhost:8080/v1/meta | jq .

echo "All vector DB tests passed!"
```

---

## 10. Performance Benchmarks

### Target Metrics

| Operation | Target Latency | Target Throughput |
|-----------|----------------|-------------------|
| Embedding (batch 100) | < 500ms | 1000/s |
| Vector search (top-10) | < 50ms | 100/s |
| Hybrid search | < 100ms | 50/s |
| RAG pipeline | < 2s | 10/s |

### Optimization Strategies

1. **Batch Processing**: Process embeddings in batches of 50-100
2. **Caching**: Cache embeddings (TTL: 1 hour)
3. **Index Optimization**: Use HNSW for approximate search
4. **Quantization**: Use product quantization for large indexes
5. **Parallel Search**: Query multiple DBs concurrently
