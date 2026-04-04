# Implementation Specification: Semantic Code Search

**Document ID:** IMPL-001  
**Feature:** Semantic Code Search  
**Priority:** CRITICAL  
**Phase:** 1  
**Estimated Effort:** 3 weeks  
**Source:** Cline, GPTMe

---

## Overview

Implement vector-based semantic code search using embeddings to enable intelligent code retrieval based on meaning rather than just text matching.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Semantic Code Search System                       │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │
│  │   indexer    │  │  embedder    │  │   searcher   │  │   cache     │ │
│  │              │  │              │  │              │  │             │ │
│  │ - File crawl │  │ - Model load │  │ - Vector     │  │ - Redis     │ │
│  │ - AST parse  │  │ - Batch enc  │  │   search     │  │ - TTL       │ │
│  │ - Chunking   │  │ - Normalize  │  │ - Rerank     │  │ - LRU       │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬──────┘ │
│         │                 │                 │                │        │
│         └─────────────────┴─────────────────┴────────────────┘        │
│                                    │                                  │
│                                    ▼                                  │
│                         ┌─────────────────┐                           │
│                         │  Vector Store   │                           │
│                         │  (ChromaDB/     │                           │
│                         │   Qdrant)       │                           │
│                         └─────────────────┘                           │
└─────────────────────────────────────────────────────────────────────────┘
```

## Components

### 1. Code Indexer (`internal/search/indexer/`)

```go
package indexer

// CodeIndexer manages the indexing pipeline
type CodeIndexer struct {
    embedder    *Embedder
    vectorStore VectorStore
    parser      *ASTParser
    config      IndexerConfig
}

type IndexerConfig struct {
    RootPath          string
    IncludePatterns   []string
    ExcludePatterns   []string
    ChunkSize         int
    ChunkOverlap      int
    MaxFileSize       int64
    IndexOnStartup    bool
    WatchFiles        bool
}

// IndexResult represents the outcome of indexing
type IndexResult struct {
    FilesIndexed  int
    ChunksCreated int
    Errors        []error
    Duration      time.Duration
}

func (i *CodeIndexer) Index(ctx context.Context) (*IndexResult, error)
func (i *CodeIndexer) IndexFile(ctx context.Context, path string) error
func (i *CodeIndexer) DeleteFile(ctx context.Context, path string) error
func (i *CodeIndexer) Watch(ctx context.Context) error
```

### 2. Code Chunking Strategy

```go
package chunker

// Chunk represents a code segment with metadata
type Chunk struct {
    ID          string
    Content     string
    FilePath    string
    StartLine   int
    EndLine     int
    Language    string
    Type        ChunkType // function, class, interface, etc.
    Parent      string    // parent scope
    Imports     []string
    Embeddings  []float32
}

type ChunkType string

const (
    ChunkTypeFunction   ChunkType = "function"
    ChunkTypeClass      ChunkType = "class"
    ChunkTypeInterface  ChunkType = "interface"
    ChunkTypeMethod     ChunkType = "method"
    ChunkTypeComment    ChunkType = "comment"
    ChunkTypeImport     ChunkType = "import"
    ChunkTypeGeneral    ChunkType = "general"
)

// Chunker splits code into semantic chunks
type Chunker interface {
    Chunk(content string, language string) ([]Chunk, error)
}

// ASTChunker uses tree-sitter for semantic chunking
type ASTChunker struct {
    parsers map[string]*sitter.Parser
}

func (c *ASTChunker) Chunk(content string, language string) ([]Chunk, error) {
    // Parse AST
    // Extract function boundaries
    // Extract class boundaries
    // Create overlapping chunks for context
}
```

### 3. Embedding Service (`internal/search/embedder/`)

```go
package embedder

// Embedder generates embeddings for code
type Embedder interface {
    Embed(ctx context.Context, texts []string) ([][]float32, error)
    EmbedQuery(ctx context.Context, query string) ([]float32, error)
    Dimensions() int
}

// LocalEmbedder uses local sentence-transformers
type LocalEmbedder struct {
    model     *ort.AdvancedSession
    tokenizer *tokenizer.Tokenizer
    dims      int
}

// RemoteEmbedder uses OpenAI or similar API
type RemoteEmbedder struct {
    client    *openai.Client
    model     string
    dims      int
}

// Supported embedding models
const (
    ModelCodeBERT       = "microsoft/codebert-base"
    ModelCodeT5         = "Salesforce/codet5-base"
    ModelJinaEmbeddings = "jinaai/jina-embeddings-v2-base-code"
    ModelOpenAIAda      = "text-embedding-3-small"
    ModelOpenAILarge    = "text-embedding-3-large"
)
```

### 4. Vector Store Interface (`internal/search/store/`)

```go
package store

// VectorStore abstracts the vector database
type VectorStore interface {
    // Collection management
    CreateCollection(ctx context.Context, name string, dims int) error
    DeleteCollection(ctx context.Context, name string) error
    
    // Document operations
    Upsert(ctx context.Context, collection string, docs []Document) error
    Delete(ctx context.Context, collection string, ids []string) error
    
    // Search
    Search(ctx context.Context, collection string, vector []float32, opts SearchOptions) ([]SearchResult, error)
    SearchByText(ctx context.Context, collection string, text string, opts SearchOptions) ([]SearchResult, error)
    
    // Metadata
    GetCollectionStats(ctx context.Context, collection string) (*CollectionStats, error)
}

type Document struct {
    ID         string
    Vector     []float32
    Metadata   map[string]interface{}
    Content    string
}

type SearchOptions struct {
    TopK        int
    MinScore    float32
    Filters     map[string]interface{}
    IncludeContent bool
}

type SearchResult struct {
    Document
    Score       float32
    Distance    float32
}
```

### 5. Search API Handler (`internal/handlers/search_handler.go`)

```go
package handlers

// SearchRequest represents a semantic search query
type SearchRequest struct {
    Query       string                 `json:"query" binding:"required"`
    Language    string                 `json:"language,omitempty"`
    FilePattern string                 `json:"file_pattern,omitempty"`
    TopK        int                    `json:"top_k" binding:"max=50"`
    MinScore    float32                `json:"min_score,omitempty"`
    Filters     map[string]interface{} `json:"filters,omitempty"`
}

type SearchResponse struct {
    Results     []SearchResult `json:"results"`
    TotalFound  int            `json:"total_found"`
    QueryTime   time.Duration  `json:"query_time_ms"`
}

func (h *SearchHandler) SemanticSearch(c *gin.Context) {
    var req SearchRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Error: err.Error()})
        return
    }
    
    results, err := h.searcher.Search(c.Request.Context(), req)
    if err != nil {
        c.JSON(500, ErrorResponse{Error: err.Error()})
        return
    }
    
    c.JSON(200, results)
}
```

## Implementation Steps

### Week 1: Foundation

**Day 1-2: Setup & Interfaces**
- [ ] Create package structure
- [ ] Define all interfaces
- [ ] Write unit test skeletons

**Day 3-4: Chunking Implementation**
- [ ] Integrate tree-sitter for Go, Python, TypeScript, Rust
- [ ] Implement AST-based chunking
- [ ] Write chunking tests

**Day 5: Embedder Implementation**
- [ ] Implement RemoteEmbedder (OpenAI)
- [ ] Implement LocalEmbedder (ONNX)
- [ ] Add embedding caching

### Week 2: Storage & Indexing

**Day 1-2: Vector Store Implementation**
- [ ] Implement ChromaDB adapter
- [ ] Implement Qdrant adapter
- [ ] Add connection pooling

**Day 3-4: Indexer Implementation**
- [ ] File crawling with gitignore support
- [ ] Parallel indexing pipeline
- [ ] Progress reporting

**Day 5: File Watching**
- [ ] Implement fsnotify-based watching
- [ ] Delta indexing
- [ ] Handle file deletions

### Week 3: API & Integration

**Day 1-2: Search API**
- [ ] Implement search endpoint
- [ ] Add query preprocessing
- [ ] Implement result reranking

**Day 3-4: MCP Integration**
- [ ] Add semantic_search tool
- [ ] Integrate with debate service
- [ ] Add context auto-loading

**Day 5: Testing & Documentation**
- [ ] Integration tests
- [ ] Performance benchmarks
- [ ] API documentation

## Configuration

```yaml
# configs/semantic_search.yaml
semantic_search:
  enabled: true
  
  embedder:
    provider: "openai"  # or "local"
    model: "text-embedding-3-small"
    dimensions: 1536
    batch_size: 100
    
  local_embedder:
    model_path: "models/all-MiniLM-L6-v2.onnx"
    tokenizer_path: "models/tokenizer.json"
    
  vector_store:
    provider: "chroma"  # or "qdrant"
    
    chroma:
      host: "localhost"
      port: 8000
      collection: "code_embeddings"
      
    qdrant:
      host: "localhost"
      port: 6333
      collection: "code_embeddings"
      
  indexer:
    root_path: "."
    include_patterns:
      - "*.go"
      - "*.py"
      - "*.js"
      - "*.ts"
      - "*.rs"
    exclude_patterns:
      - "vendor/"
      - "node_modules/"
      - ".git/"
      - "*.pb.go"
    chunk_size: 512
    chunk_overlap: 128
    max_file_size: 1048576  # 1MB
    index_on_startup: true
    watch_files: true
    
  search:
    default_top_k: 10
    min_score: 0.7
    rerank_enabled: true
    rerank_model: "cross-encoder/ms-marco-MiniLM-L-6-v2"
    cache_ttl: 3600
```

## Database Schema

```sql
-- ChromaDB collections are schemaless, but we define our metadata structure

-- Document metadata structure:
-- {
--   "id": "file:line_start:line_end",
--   "file_path": "relative/path/to/file.go",
--   "start_line": 10,
--   "end_line": 50,
--   "language": "go",
--   "chunk_type": "function",
--   "parent": "ParentFunction",
--   "last_modified": "2026-04-04T10:00:00Z"
-- }

-- Indexing progress tracking
CREATE TABLE IF NOT EXISTS indexing_progress (
    id SERIAL PRIMARY KEY,
    file_path TEXT UNIQUE NOT NULL,
    last_indexed TIMESTAMP WITH TIME ZONE,
    chunks_count INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending',
    error_message TEXT
);

-- Search query logging for analytics
CREATE TABLE IF NOT EXISTS search_queries (
    id SERIAL PRIMARY KEY,
    query TEXT NOT NULL,
    results_count INTEGER,
    query_time_ms INTEGER,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## API Endpoints

```yaml
# OpenAPI spec for semantic search
openapi: 3.0.0
paths:
  /v1/search/semantic:
    post:
      summary: Semantic code search
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/SearchRequest'
      responses:
        200:
          description: Search results
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SearchResponse'
                
  /v1/search/index:
    post:
      summary: Trigger full reindex
      responses:
        200:
          description: Indexing started
          
  /v1/search/index/{path}:
    post:
      summary: Index specific file
      parameters:
        - name: path
          in: path
          required: true
          schema:
            type: string
    delete:
      summary: Remove file from index
      
  /v1/search/stats:
    get:
      summary: Get indexing statistics
      responses:
        200:
          description: Statistics
```

## Testing Strategy

```go
// Unit tests for each component
func TestASTChunker_Chunk(t *testing.T) {}
func TestRemoteEmbedder_Embed(t *testing.T) {}
func TestChromaStore_Search(t *testing.T) {}

// Integration tests
func TestEndToEndSearch(t *testing.T) {
    // Index test codebase
    // Perform search
    // Verify results
}

// Performance benchmarks
func BenchmarkIndexing(b *testing.B) {}
func BenchmarkSearch(b *testing.B) {}
```

## Dependencies

```go
// go.mod additions
go get github.com/go-git/go-git/v5
go get github.com/sashabaranov/go-openai
go get github.com/chromedp/chromedp

go get github.com/qdrant/go-client

go get github.com/amikos-tech/chroma-go
```

## Success Criteria

- [ ] Index 10,000 files in under 5 minutes
- [ ] Search query response <100ms p95
- [ ] Embedding generation <50ms per query
- [ ] Memory usage <2GB for 100K chunks
- [ ] Test coverage >90%
