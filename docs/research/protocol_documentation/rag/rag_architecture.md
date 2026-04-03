# RAG (Retrieval-Augmented Generation) - Complete Specification

**Protocol:** RAG Pipeline  
**Version:** 1.0  
**Status:** Draft  
**HelixAgent Implementation:** [internal/rag/](../../../internal/rag/)  
**Analysis Date:** 2026-04-03  

---

## Executive Summary

RAG (Retrieval-Augmented Generation) enhances LLM responses by retrieving relevant context from a knowledge base before generation. This reduces hallucinations and grounds responses in factual data.

**Key Components:**
1. Document ingestion and chunking
2. Embedding generation
3. Vector storage
4. Context retrieval
5. Response generation with context

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         RAG PIPELINE ARCHITECTURE                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────┐  │
│  │   Documents  │───►│   Chunking   │───►│  Embeddings  │───►│  Vector  │  │
│  │   (PDF, MD,  │    │   Strategy   │    │   (Text →   │    │   DB     │  │
│  │    TXT, etc) │    │              │    │   Vectors)   │    │          │  │
│  └──────────────┘    └──────────────┘    └──────────────┘    └────┬─────┘  │
│                                                                    │        │
│                                                                    │        │
│  Query: "What is our refund policy?"                              │        │
│       │                                                            │        │
│       ▼                                                            ▼        │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────┐  │
│  │   Query      │───►│  Similarity  │◄───┤   Vector     │◄───┤  Context │  │
│  │   Embedding  │    │    Search    │    │   Database   │    │  Store   │  │
│  └──────────────┘    └──────────────┘    └──────────────┘    └──────────┘  │
│                             │                                                │
│                             ▼                                                │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                        CONTEXT ASSEMBLY                                 │ │
│  │  Retrieved chunks (top-k):                                              │ │
│  │  • "Our refund policy allows returns within 30 days..." (score: 0.92)  │ │
│  │  • "To initiate a refund, contact support@..." (score: 0.85)          │ │
│  │  • "Refunds processed within 5-7 business days..." (score: 0.78)      │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                    │                                        │
│                                    ▼                                        │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                    AUGMENTED GENERATION                                 │ │
│  │                                                                         │ │
│  │  Prompt:                                                                │ │
│  │  Context:                                                               │ │
│  │  [1] Our refund policy allows returns within 30 days...                │ │
│  │  [2] To initiate a refund, contact support@...                         │ │
│  │                                                                         │ │
│  │  Question: What is our refund policy?                                  │ │
│  │                                                                         │ │
│  │  Answer:                                                                │ │
│  │  Based on our policy, you can return items within 30 days...           │ │
│  │                                                                         │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## RAG Pipeline Stages

### Stage 1: Document Ingestion

```typescript
interface Document {
  id: string;
  content: string;
  metadata: {
    source: string;        // File path or URL
    title?: string;
    author?: string;
    createdAt?: Date;
    modifiedAt?: Date;
    type: 'pdf' | 'markdown' | 'text' | 'html' | 'code';
    tags?: string[];
    language?: string;
  };
}

// Supported document formats
const SUPPORTED_FORMATS = [
  'application/pdf',
  'text/markdown',
  'text/plain',
  'text/html',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/json',
  'text/csv'
];
```

### Stage 2: Chunking Strategies

```typescript
interface ChunkingConfig {
  strategy: 'fixed' | 'semantic' | 'recursive' | 'code-aware';
  chunkSize: number;        // Target chunk size (tokens)
  chunkOverlap: number;     // Overlap between chunks (tokens)
  separator?: string;       // Custom separator
}

// Fixed-size chunking
function fixedChunking(text: string, size: number, overlap: number): Chunk[] {
  const chunks: Chunk[] = [];
  let start = 0;
  
  while (start < text.length) {
    const end = Math.min(start + size, text.length);
    chunks.push({
      text: text.slice(start, end),
      startIndex: start,
      endIndex: end
    });
    start = end - overlap;
  }
  
  return chunks;
}

// Semantic chunking (by paragraphs/sections)
function semanticChunking(text: string): Chunk[] {
  // Split on semantic boundaries
  const separators = ['\n\n', '\n', '. ', '? ', '! '];
  // Implementation...
}

// Recursive chunking (hierarchical)
function recursiveChunking(text: string, config: ChunkingConfig): Chunk[] {
  // Try separators in order of preference
  // If chunk too large, recursively split
  // Implementation...
}

// Code-aware chunking
function codeChunking(code: string, language: string): Chunk[] {
  // Respect code structure (functions, classes, etc.)
  // Implementation...
}
```

**Chunk Size Guidelines:**

| Use Case | Chunk Size | Overlap | Strategy |
|----------|-----------|---------|----------|
| General Q&A | 512 tokens | 50 tokens | Semantic |
| Code search | 256 tokens | 25 tokens | Code-aware |
| Legal docs | 1024 tokens | 100 tokens | Recursive |
| Conversational | 384 tokens | 40 tokens | Semantic |

### Stage 3: Embedding Generation

```typescript
interface EmbeddingConfig {
  model: string;           // e.g., 'text-embedding-3-large'
  batchSize: number;       // Default: 100
  normalize: boolean;      // Default: true
  dimensions?: number;     // For dimension reduction
}

async function generateEmbeddings(
  chunks: Chunk[],
  config: EmbeddingConfig
): Promise<EmbeddedChunk[]> {
  // Implementation using embeddings provider
}
```

### Stage 4: Vector Storage

```typescript
interface VectorStoreConfig {
  type: 'chromadb' | 'qdrant' | 'pinecone' | 'weaviate';
  collection: string;
  distance: 'cosine' | 'euclidean' | 'dot';
}

interface EmbeddedChunk {
  id: string;
  text: string;
  embedding: number[];
  metadata: {
    documentId: string;
    chunkIndex: number;
    startIndex: number;
    endIndex: number;
  };
}
```

### Stage 5: Context Retrieval

```typescript
interface RetrievalConfig {
  topK: number;           // Number of chunks to retrieve
  threshold: number;      // Minimum similarity score
  rerank: boolean;        // Use reranking
  expandContext: boolean; // Include neighboring chunks
}

interface RetrievedChunk {
  chunk: EmbeddedChunk;
  score: number;
  rank: number;
}

async function retrieveContext(
  query: string,
  config: RetrievalConfig
): Promise<RetrievedChunk[]> {
  // 1. Generate query embedding
  const queryEmbedding = await generateEmbedding(query);
  
  // 2. Similarity search
  const results = await vectorSearch(queryEmbedding, config.topK);
  
  // 3. Filter by threshold
  const filtered = results.filter(r => r.score >= config.threshold);
  
  // 4. Rerank (optional)
  if (config.rerank) {
    return await rerankResults(query, filtered);
  }
  
  return filtered;
}
```

### Stage 6: Context Assembly

```typescript
interface ContextAssemblyConfig {
  maxTokens: number;      // Maximum context tokens
  format: 'numbered' | 'bulleted' | 'citations';
  includeMetadata: boolean;
}

function assembleContext(
  chunks: RetrievedChunk[],
  config: ContextAssemblyConfig
): string {
  let context = '';
  let tokens = 0;
  
  for (let i = 0; i < chunks.length; i++) {
    const chunk = chunks[i];
    const chunkText = formatChunk(chunk, i + 1, config.format);
    const chunkTokens = estimateTokens(chunkText);
    
    if (tokens + chunkTokens > config.maxTokens) {
      break;
    }
    
    context += chunkText + '\n\n';
    tokens += chunkTokens;
  }
  
  return context.trim();
}

function formatChunk(
  chunk: RetrievedChunk,
  index: number,
  format: string
): string {
  switch (format) {
    case 'numbered':
      return `[${index}] ${chunk.chunk.text}`;
    case 'citations':
      return `${chunk.chunk.text} [Source: ${chunk.chunk.metadata.documentId}]`;
    default:
      return `• ${chunk.chunk.text}`;
  }
}
```

### Stage 7: Augmented Generation

```typescript
interface GenerationConfig {
  model: string;
  temperature: number;
  maxTokens: number;
  systemPrompt?: string;
}

async function generateResponse(
  query: string,
  context: string,
  config: GenerationConfig
): Promise<string> {
  const prompt = buildRAGPrompt(query, context);
  
  const response = await llm.generate({
    model: config.model,
    messages: [
      {
        role: 'system',
        content: config.systemPrompt || 'You are a helpful assistant. Use the provided context to answer questions accurately. If the context doesn\'t contain the answer, say so.'
      },
      {
        role: 'user',
        content: prompt
      }
    ],
    temperature: config.temperature,
    max_tokens: config.maxTokens
  });
  
  return response.content;
}

function buildRAGPrompt(query: string, context: string): string {
  return `Context:
${context}

Question: ${query}

Answer:`;
}
```

---

## HelixAgent RAG Implementation

### Architecture

**Source:** [`internal/rag/`](../../../internal/rag/)

```
internal/rag/
├── pipeline.go             # Main RAG pipeline
├── ingestion/
│   ├── loader.go          # Document loading
│   ├── chunker.go         # Text chunking
│   └── parser.go          # Document parsing
├── embedding/
│   ├── embedder.go        # Embedding generation
│   └── batch.go           # Batch processing
├── retrieval/
│   ├── retriever.go       # Context retrieval
│   ├── reranker.go        # Result reranking
│   └── filters.go         # Query filters
├── generation/
│   ├── augmenter.go       # Context assembly
│   └── prompt.go          # Prompt building
└── storage/
    ├── document_store.go  # Document metadata
    └── vector_store.go    # Vector storage
```

### Pipeline Implementation

**Source:** [`internal/rag/pipeline.go`](../../../internal/rag/pipeline.go)

```go
package rag

// RAGPipeline implements complete RAG workflow
// Source: internal/rag/pipeline.go#L1-250

type RAGPipeline struct {
    loader      *DocumentLoader
    chunker     *TextChunker
    embedder    *EmbeddingGenerator
    retriever   *ContextRetriever
    generator   *ResponseGenerator
    docStore    DocumentStore
    vectorStore VectorStore
}

// IngestDocument processes document through pipeline
// Source: internal/rag/pipeline.go#L45-95
func (p *RAGPipeline) IngestDocument(ctx context.Context, doc *Document) error {
    // 1. Parse document
    content, err := p.loader.Load(ctx, doc)
    if err != nil {
        return fmt.Errorf("load document: %w", err)
    }
    
    // 2. Chunk document
    chunks := p.chunker.Chunk(content, ChunkingConfig{
        Strategy:     "semantic",
        ChunkSize:    512,
        ChunkOverlap: 50,
    })
    
    // 3. Generate embeddings
    embeddedChunks, err := p.embedder.EmbedChunks(ctx, chunks)
    if err != nil {
        return fmt.Errorf("generate embeddings: %w", err)
    }
    
    // 4. Store in vector DB
    if err := p.vectorStore.Upsert(ctx, embeddedChunks); err != nil {
        return fmt.Errorf("store vectors: %w", err)
    }
    
    // 5. Store document metadata
    return p.docStore.Save(ctx, doc)
}

// Query performs RAG query
// Source: internal/rag/pipeline.go#L97-180
func (p *RAGPipeline) Query(ctx context.Context, req *RAGQueryRequest) (*RAGQueryResponse, error) {
    // 1. Generate query embedding
    queryEmbedding, err := p.embedder.EmbedText(ctx, req.Query)
    if err != nil {
        return nil, fmt.Errorf("embed query: %w", err)
    }
    
    // 2. Retrieve relevant chunks
    retrieved, err := p.retriever.Retrieve(ctx, &RetrieveRequest{
        QueryEmbedding: queryEmbedding,
        TopK:          req.TopK,
        Threshold:     req.Threshold,
        Filters:       req.Filters,
    })
    if err != nil {
        return nil, fmt.Errorf("retrieve context: %w", err)
    }
    
    // 3. Rerank if enabled
    if req.Rerank {
        retrieved, err = p.retriever.Rerank(ctx, req.Query, retrieved)
        if err != nil {
            return nil, fmt.Errorf("rerank: %w", err)
        }
    }
    
    // 4. Assemble context
    context := p.generator.AssembleContext(retrieved, ContextAssemblyConfig{
        MaxTokens:       req.MaxContextTokens,
        Format:          req.ContextFormat,
        IncludeMetadata: req.IncludeMetadata,
    })
    
    // 5. Generate response
    answer, err := p.generator.Generate(ctx, &GenerateRequest{
        Query:       req.Query,
        Context:     context,
        Model:       req.Model,
        Temperature: req.Temperature,
    })
    if err != nil {
        return nil, fmt.Errorf("generate response: %w", err)
    }
    
    return &RAGQueryResponse{
        Answer:    answer,
        Context:   context,
        Sources:   p.extractSources(retrieved),
        Metadata:  p.buildMetadata(retrieved),
    }, nil
}
```

### Chunking Implementation

**Source:** [`internal/rag/ingestion/chunker.go`](../../../internal/rag/ingestion/chunker.go)

```go
package ingestion

// TextChunker implements multiple chunking strategies
// Source: internal/rag/ingestion/chunker.go#L1-180

type TextChunker struct {
    tokenizer *Tokenizer
}

// Chunk splits text using configured strategy
// Source: internal/rag/ingestion/chunker.go#L45-95
func (c *TextChunker) Chunk(text string, config ChunkingConfig) []Chunk {
    switch config.Strategy {
    case "fixed":
        return c.fixedChunking(text, config)
    case "semantic":
        return c.semanticChunking(text, config)
    case "recursive":
        return c.recursiveChunking(text, config)
    case "code-aware":
        return c.codeChunking(text, config)
    default:
        return c.semanticChunking(text, config)
    }
}

// semanticChunking splits on semantic boundaries
// Source: internal/rag/ingestion/chunker.go#L97-150
func (c *TextChunker) semanticChunking(text string, config ChunkingConfig) []Chunk {
    // Split on semantic boundaries (paragraphs, sentences)
    separators := []string{"\n\n", "\n", ". ", "? ", "! "}
    
    var chunks []Chunk
    var currentChunk strings.Builder
    var currentTokens int
    var startIndex int
    
    sentences := c.splitSentences(text)
    
    for i, sentence := range sentences {
        sentenceTokens := c.tokenizer.Count(sentence)
        
        if currentTokens+sentenceTokens > config.ChunkSize && currentTokens > 0 {
            // Save current chunk
            chunks = append(chunks, Chunk{
                Text:       currentChunk.String(),
                StartIndex: startIndex,
                EndIndex:   startIndex + len(currentChunk.String()),
                Index:      len(chunks),
            })
            
            // Start new chunk with overlap
            overlapStart := max(0, len(chunks)-config.ChunkOverlap)
            currentChunk.Reset()
            currentTokens = 0
            startIndex = chunks[overlapStart].StartIndex
        }
        
        currentChunk.WriteString(sentence)
        currentTokens += sentenceTokens
    }
    
    // Don't forget last chunk
    if currentTokens > 0 {
        chunks = append(chunks, Chunk{
            Text:       currentChunk.String(),
            StartIndex: startIndex,
            EndIndex:   startIndex + len(currentChunk.String()),
            Index:      len(chunks),
        })
    }
    
    return chunks
}
```

---

## API Endpoints

### Document Management

```
POST /v1/rag/documents              # Upload document
GET  /v1/rag/documents              # List documents
GET  /v1/rag/documents/{id}         # Get document
DELETE /v1/rag/documents/{id}       # Delete document
POST /v1/rag/documents/{id}/ingest  # Trigger ingestion
```

### Query

```
POST /v1/rag/query                  # RAG query
POST /v1/rag/retrieve               # Retrieve only (no generation)
```

### Example: Upload and Query

```bash
# Upload document
curl -X POST http://localhost:7061/v1/rag/documents \
  -F "file=@company-handbook.pdf" \
  -F "metadata={\"type\":\"pdf\",\"category\":\"hr\"}"

# Response
{"id": "doc-123", "status": "processing"}

# Wait for ingestion...

# Query
curl -X POST http://localhost:7061/v1/rag/query \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What is the vacation policy?",
    "top_k": 5,
    "rerank": true,
    "model": "claude-3-5-sonnet"
  }'

# Response
{
  "answer": "According to the company handbook, employees receive 20 days of paid vacation per year...",
  "sources": [
    {
      "document_id": "doc-123",
      "chunk_index": 5,
      "score": 0.92,
      "text": "Employees are entitled to 20 days of paid vacation..."
    }
  ],
  "metadata": {
    "retrieval_time_ms": 45,
    "generation_time_ms": 1200,
    "total_tokens": 450
  }
}
```

---

## Advanced Features

### Hybrid Search

Combine vector similarity with keyword matching:

```go
// Source: internal/rag/retrieval/hybrid.go

func (r *HybridRetriever) Retrieve(ctx context.Context, req *RetrieveRequest) ([]RetrievedChunk, error) {
    // Vector search
    vectorResults, err := r.vectorSearch(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // Keyword search (BM25)
    keywordResults, err := r.keywordSearch(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // Fuse results
    return r.fuseResults(vectorResults, keywordResults, 0.7) // 70% vector, 30% keyword
}
```

### Query Rewriting

Improve retrieval with query expansion:

```go
// Source: internal/rag/retrieval/rewriter.go

func (r *QueryRewriter) Rewrite(ctx context.Context, query string) ([]string, error) {
    // Generate variations
    variations := []string{query}
    
    // Hypothetical document embedding
    hypothetical, err := r.generateHypotheticalAnswer(ctx, query)
    if err == nil {
        variations = append(variations, hypothetical)
    }
    
    // Sub-queries
    subQueries, err := r.generateSubQueries(ctx, query)
    if err == nil {
        variations = append(variations, subQueries...)
    }
    
    return variations, nil
}
```

### Reranking

Use cross-encoder for better relevance:

```go
// Source: internal/rag/retrieval/reranker.go

func (r *CrossEncoderReranker) Rerank(ctx context.Context, query string, chunks []RetrievedChunk) ([]RetrievedChunk, error) {
    // Prepare pairs
    pairs := make([][]string, len(chunks))
    for i, chunk := range chunks {
        pairs[i] = []string{query, chunk.Text}
    }
    
    // Score with cross-encoder
    scores, err := r.model.Predict(ctx, pairs)
    if err != nil {
        return nil, err
    }
    
    // Sort by new scores
    for i := range chunks {
        chunks[i].Score = scores[i]
    }
    
    sort.Slice(chunks, func(i, j int) bool {
        return chunks[i].Score > chunks[j].Score
    })
    
    return chunks, nil
}
```

---

## Source Code Reference

### RAG Core Files

| Component | Source File | Lines | Description |
|-----------|-------------|-------|-------------|
| Pipeline | `internal/rag/pipeline.go` | 250 | Main RAG workflow |
| Chunker | `internal/rag/ingestion/chunker.go` | 180 | Text chunking |
| Loader | `internal/rag/ingestion/loader.go` | 120 | Document loading |
| Retriever | `internal/rag/retrieval/retriever.go` | 150 | Context retrieval |
| Reranker | `internal/rag/retrieval/reranker.go` | 100 | Result reranking |
| Generator | `internal/rag/generation/augmenter.go` | 130 | Response generation |
| Handler | `internal/handlers/rag.go` | 90 | HTTP handler |

---

## Conclusion

HelixAgent's RAG implementation provides:

- ✅ End-to-end RAG pipeline
- ✅ Multiple chunking strategies
- ✅ Multi-provider embeddings
- ✅ Vector database integration
- ✅ Advanced retrieval (hybrid, reranking)
- ✅ Query rewriting

**Recommendation:** Use RAG for knowledge base Q&A, document analysis, and grounded generation.

---

*Specification Version: 1.0*  
*Last Updated: 2026-04-03*  
*HelixAgent Commit: aa960946*
