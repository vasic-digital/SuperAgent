# Knowledge Package

The knowledge package implements **GraphRAG** (Graph-based Retrieval Augmented Generation) for intelligent code retrieval and knowledge management in HelixAgent.

## Overview

GraphRAG combines traditional vector similarity search with knowledge graph traversal to provide more contextually relevant retrieval results. It enables HelixAgent to understand code relationships, dependencies, and semantic connections when answering questions or generating code.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       GraphRAG Engine                        │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Retrieval Pipeline                     ││
│  │                                                          ││
│  │  Query → Embed → Search → Traverse → Rerank → Results   ││
│  │                                                          ││
│  │  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────────────┐││
│  │  │ Vector │  │ Graph  │  │Semantic│  │    Hybrid      │││
│  │  │ Search │  │Traverse│  │  Match │  │   Combiner     │││
│  │  └────────┘  └────────┘  └────────┘  └────────────────┘││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Knowledge Graph                        ││
│  │  ┌──────┐    ┌──────┐    ┌──────┐    ┌──────────────┐  ││
│  │  │Nodes │────│Edges │────│Types │────│  Properties  │  ││
│  │  │(Code)│    │(Rels)│    │      │    │              │  ││
│  │  └──────┘    └──────┘    └──────┘    └──────────────┘  ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    LLM Integration                       ││
│  │  Reranker | Query Expansion | Selective Retrieval       ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Key Types

### GraphRAG

The main GraphRAG engine.

```go
type GraphRAG struct {
    vectorStore    VectorStore
    knowledgeGraph *KnowledgeGraph
    embedder       Embedder
    reranker       Reranker
    config         GraphRAGConfig
}
```

### KnowledgeGraph

Represents the code knowledge graph.

```go
type KnowledgeGraph struct {
    nodes map[string]*Node
    edges map[string][]*Edge
    mu    sync.RWMutex
}

type Node struct {
    ID         string
    Type       NodeType
    Content    string
    Embedding  []float32
    Properties map[string]interface{}
    CreatedAt  time.Time
}

type Edge struct {
    Source   string
    Target   string
    Type     EdgeType
    Weight   float64
    Metadata map[string]interface{}
}
```

### Node Types

```go
type NodeType string
const (
    NodeTypeFile      NodeType = "file"
    NodeTypeFunction  NodeType = "function"
    NodeTypeClass     NodeType = "class"
    NodeTypeMethod    NodeType = "method"
    NodeTypeVariable  NodeType = "variable"
    NodeTypeImport    NodeType = "import"
    NodeTypeComment   NodeType = "comment"
    NodeTypeTest      NodeType = "test"
)
```

### Edge Types

```go
type EdgeType string
const (
    EdgeTypeCalls       EdgeType = "calls"
    EdgeTypeImports     EdgeType = "imports"
    EdgeTypeImplements  EdgeType = "implements"
    EdgeTypeInherits    EdgeType = "inherits"
    EdgeTypeContains    EdgeType = "contains"
    EdgeTypeReferences  EdgeType = "references"
    EdgeTypeTests       EdgeType = "tests"
    EdgeTypeDepends     EdgeType = "depends"
)
```

### RetrievalMode

Configures how retrieval is performed.

```go
type RetrievalMode string
const (
    // Only vector similarity search
    RetrievalModeLocal RetrievalMode = "local"

    // Only graph traversal
    RetrievalModeGraph RetrievalMode = "graph"

    // Only semantic matching
    RetrievalModeSemantic RetrievalMode = "semantic"

    // Combination of all modes
    RetrievalModeHybrid RetrievalMode = "hybrid"
)
```

### RetrievalResult

Result from a retrieval query.

```go
type RetrievalResult struct {
    Node       *Node
    Score      float64
    Source     RetrievalMode
    Path       []*Edge          // Graph path if applicable
    Context    []string         // Related context
    Confidence float64
}
```

## Configuration

```go
type GraphRAGConfig struct {
    // Vector search settings
    VectorTopK         int     // Top K results from vector search
    VectorThreshold    float64 // Minimum similarity score

    // Graph traversal settings
    MaxTraversalDepth  int     // Maximum graph hops
    MaxNeighbors       int     // Max neighbors per hop
    EdgeWeightThreshold float64 // Minimum edge weight

    // Hybrid settings
    VectorWeight       float64 // Weight for vector results (0-1)
    GraphWeight        float64 // Weight for graph results (0-1)
    SemanticWeight     float64 // Weight for semantic results (0-1)

    // Reranking
    EnableReranking    bool    // Enable LLM reranking
    RerankTopK         int     // Rerank top K results

    // Caching
    EnableCache        bool
    CacheTTL           time.Duration
}
```

## Usage Examples

### Basic Retrieval

```go
import "dev.helix.agent/internal/knowledge"

// Create GraphRAG engine
graphrag := knowledge.NewGraphRAG(knowledge.GraphRAGConfig{
    VectorTopK:        10,
    MaxTraversalDepth: 3,
    EnableReranking:   true,
})

// Initialize with codebase
err := graphrag.IndexCodebase("/path/to/project")
if err != nil {
    return err
}

// Retrieve relevant code
results, err := graphrag.Retrieve(ctx, knowledge.RetrievalQuery{
    Query:   "How does the authentication middleware work?",
    Mode:    knowledge.RetrievalModeHybrid,
    TopK:    5,
})
if err != nil {
    return err
}

for _, result := range results {
    fmt.Printf("Score: %.2f - %s (%s)\n",
        result.Score, result.Node.ID, result.Node.Type)
}
```

### Graph-Based Retrieval

```go
// Find all functions that call a specific function
results, err := graphrag.TraverseFrom(ctx, knowledge.TraversalQuery{
    StartNode:  "internal/auth/handler.go:Authenticate",
    EdgeTypes:  []knowledge.EdgeType{knowledge.EdgeTypeCalls},
    Direction:  knowledge.DirectionIncoming,
    MaxDepth:   2,
})
```

### Adding Knowledge

```go
// Add a code node
node := &knowledge.Node{
    ID:      "internal/services/user.go:CreateUser",
    Type:    knowledge.NodeTypeFunction,
    Content: functionCode,
    Properties: map[string]interface{}{
        "language":   "go",
        "package":    "services",
        "visibility": "public",
    },
}
graphrag.AddNode(ctx, node)

// Add relationship
edge := &knowledge.Edge{
    Source: "internal/services/user.go:CreateUser",
    Target: "internal/database/user_repo.go:Insert",
    Type:   knowledge.EdgeTypeCalls,
    Weight: 1.0,
}
graphrag.AddEdge(ctx, edge)
```

### LLM Reranking

```go
// Configure LLM reranker
reranker := knowledge.NewLLMReranker(knowledge.RerankerConfig{
    Provider: llmProvider,
    Model:    "claude-3-sonnet",
    Template: `Given the query and candidates, rank by relevance:
Query: {{.Query}}
Candidates:
{{range .Candidates}}
- {{.ID}}: {{.Content | truncate 200}}
{{end}}`,
})

graphrag.SetReranker(reranker)
```

### Selective Retrieval

```go
// Use policy model to decide when to retrieve
selective := knowledge.NewSelectiveRetriever(knowledge.SelectiveConfig{
    PolicyModel: policyLLM,
    GraphRAG:    graphrag,
    Threshold:   0.7, // Only retrieve if confidence > 0.7
})

// Automatically decides whether retrieval is needed
response, retrieved := selective.MaybeRetrieve(ctx, query)
```

### Query Expansion

```go
// Expand query for better retrieval
expander := knowledge.NewQueryExpander(knowledge.ExpanderConfig{
    LLM:          llmProvider,
    MaxExpansions: 3,
})

expandedQueries := expander.Expand(ctx, "auth middleware")
// Returns: ["authentication middleware", "auth handler", "JWT validation"]

// Search with all expansions
results := graphrag.RetrieveMultiQuery(ctx, expandedQueries)
```

## Integration with HelixAgent

GraphRAG is used for contextual code understanding:

```go
// In debate service
func (s *DebateService) GetCodeContext(query string) ([]string, error) {
    results, err := s.graphrag.Retrieve(ctx, knowledge.RetrievalQuery{
        Query: query,
        Mode:  knowledge.RetrievalModeHybrid,
        TopK:  10,
    })
    if err != nil {
        return nil, err
    }

    var contexts []string
    for _, r := range results {
        contexts = append(contexts, r.Node.Content)
    }
    return contexts, nil
}
```

## Indexing Codebases

```go
// Index a codebase
indexer := knowledge.NewCodebaseIndexer(knowledge.IndexerConfig{
    Languages:       []string{"go", "typescript", "python"},
    ExcludePatterns: []string{"vendor/**", "node_modules/**"},
    ChunkSize:       1000,  // Characters per chunk
    ChunkOverlap:    100,   // Overlap between chunks
})

// Index incrementally
stats, err := indexer.IndexDirectory(ctx, "/path/to/project", graphrag)
fmt.Printf("Indexed %d files, %d nodes, %d edges\n",
    stats.Files, stats.Nodes, stats.Edges)
```

## Testing

```bash
go test -v ./internal/knowledge/...
```

### Testing Retrieval Quality

```go
func TestRetrievalQuality(t *testing.T) {
    graphrag := setupTestGraphRAG(t)

    // Test query
    results, err := graphrag.Retrieve(ctx, knowledge.RetrievalQuery{
        Query: "database connection handling",
        TopK:  5,
    })
    require.NoError(t, err)

    // Verify expected results are in top-5
    assert.Contains(t, getIDs(results), "internal/database/pool.go:Connect")
}
```

## Performance

| Operation | Typical Latency | Notes |
|-----------|----------------|-------|
| Vector Search | 10-50ms | Depends on index size |
| Graph Traversal | 1-10ms | Per hop |
| LLM Reranking | 200-500ms | With fast model |
| Full Hybrid | 100-300ms | Combined |

## Best Practices

1. **Index incrementally**: Update on file changes
2. **Tune weights**: Adjust hybrid weights for your codebase
3. **Use selective retrieval**: Don't retrieve for simple queries
4. **Cache embeddings**: Avoid recomputing for unchanged code
5. **Monitor retrieval quality**: Track relevance metrics
