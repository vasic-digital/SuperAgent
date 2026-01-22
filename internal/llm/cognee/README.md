# Cognee LLM Client Package

This package provides an HTTP client for the Cognee knowledge graph and memory service.

## Overview

Cognee is a knowledge graph system that enables memory storage, semantic search, and knowledge processing. This client handles communication with the Cognee API for memory operations, dataset management, and knowledge graph visualization.

## Features

- **Memory Management**: Store and retrieve content with vector embeddings
- **Semantic Search**: Search memories using natural language queries
- **Dataset Management**: Create and manage memory datasets
- **Knowledge Processing**: Process and cognify datasets for enhanced retrieval
- **Graph Visualization**: Visualize knowledge graph structures
- **Code Pipeline**: Process code for knowledge extraction

## Data Types

### Client

```go
type Client struct {
    baseURL string
    apiKey  string
    client  *http.Client
}
```

### Memory Operations

```go
type MemoryRequest struct {
    Content     string // Content to store
    DatasetName string // Target dataset
    ContentType string // Content MIME type
}

type MemoryResponse struct {
    VectorID   string                 // Generated vector ID
    GraphNodes map[string]interface{} // Knowledge graph nodes
}
```

### Search Operations

```go
type SearchRequest struct {
    Query       string // Search query
    DatasetName string // Dataset to search
    Limit       int    // Max results
}

type SearchResponse struct {
    Results []models.MemorySource // Search results
}
```

### Cognify Operations

```go
type CognifyRequest struct {
    Datasets []string // Datasets to process
}

type CognifyResponse struct {
    Status string // Processing status
}
```

### Insights Operations

```go
type InsightsRequest struct {
    Query    string   // Insight query
    Datasets []string // Target datasets
    Limit    int      // Max insights
}

type InsightsResponse struct {
    Insights []map[string]interface{} // Generated insights
}
```

### Code Pipeline

```go
type CodePipelineRequest struct {
    Code        string // Source code
    DatasetName string // Target dataset
    Language    string // Programming language
}

type CodePipelineResponse struct {
    Processed bool                   // Processing status
    Results   map[string]interface{} // Extraction results
}
```

### Dataset Management

```go
type DatasetRequest struct {
    Name        string                 // Dataset name
    Description string                 // Description
    Metadata    map[string]interface{} // Custom metadata
}

type DatasetResponse struct {
    ID          string                 // Dataset ID
    Name        string                 // Dataset name
    Description string                 // Description
    CreatedAt   string                 // Creation timestamp
    Metadata    map[string]interface{} // Metadata
}
```

### Visualization

```go
type VisualizeRequest struct {
    DatasetName string // Dataset to visualize
    Format      string // Output format (json, graphml)
}

type VisualizeResponse struct {
    Graph map[string]interface{} // Graph data
}
```

## Usage

### Creating a Client

```go
import "dev.helix.agent/internal/llm/cognee"

client := cognee.NewClient(config)
```

### Storing Memory

```go
resp, err := client.AddMemory(ctx, &cognee.MemoryRequest{
    Content:     "HelixAgent supports 10 LLM providers",
    DatasetName: "documentation",
    ContentType: "text/plain",
})

fmt.Printf("Vector ID: %s\n", resp.VectorID)
```

### Searching Memories

```go
results, err := client.Search(ctx, &cognee.SearchRequest{
    Query:       "What LLM providers are supported?",
    DatasetName: "documentation",
    Limit:       5,
})

for _, result := range results.Results {
    fmt.Printf("Match: %s (score: %.2f)\n", result.Content, result.Score)
}
```

### Processing Datasets (Cognify)

```go
resp, err := client.Cognify(ctx, &cognee.CognifyRequest{
    Datasets: []string{"documentation", "codebase"},
})

fmt.Printf("Status: %s\n", resp.Status)
```

### Getting Insights

```go
insights, err := client.GetInsights(ctx, &cognee.InsightsRequest{
    Query:    "What are the main components?",
    Datasets: []string{"documentation"},
    Limit:    10,
})

for _, insight := range insights.Insights {
    fmt.Printf("Insight: %v\n", insight)
}
```

### Code Processing

```go
resp, err := client.ProcessCode(ctx, &cognee.CodePipelineRequest{
    Code:        sourceCode,
    DatasetName: "codebase",
    Language:    "go",
})
```

### Dataset Management

```go
// Create dataset
dataset, err := client.CreateDataset(ctx, &cognee.DatasetRequest{
    Name:        "my-project",
    Description: "Project documentation",
})

// List datasets
datasets, err := client.ListDatasets(ctx)
```

### Graph Visualization

```go
graph, err := client.Visualize(ctx, &cognee.VisualizeRequest{
    DatasetName: "documentation",
    Format:      "json",
})
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `COGNEE_API_URL` | Cognee API endpoint | `http://localhost:8000` |
| `COGNEE_API_KEY` | API authentication key | - |

### Config File

```yaml
cognee:
  base_url: "http://localhost:8000"
  api_key: "${COGNEE_API_KEY}"
  timeout: 30s
```

## Testing

```bash
go test -v ./internal/llm/cognee/...
```

## Files

- `client.go` - HTTP client implementation
