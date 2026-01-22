# LMQL Package

This package provides an HTTP client for the LMQL service, enabling constrained query execution with variable extraction.

## Overview

LMQL (Language Model Query Language) allows writing SQL-like queries for LLMs, with support for variable extraction, constraints, and structured output.

## Features

- **Query Variables**: Extract structured data from responses
- **Constraint Enforcement**: Enforce output constraints
- **Template Queries**: Reusable query templates
- **Type Safety**: Variable type validation

## Components

### Client (`client.go`)

HTTP client for LMQL service:

```go
config := &lmql.ClientConfig{
    BaseURL: "http://localhost:8014",
    Timeout: 120 * time.Second,
}

client := lmql.NewClient(config)
```

## Data Types

### ClientConfig

```go
type ClientConfig struct {
    BaseURL string        // LMQL server URL
    Timeout time.Duration // Request timeout
}
```

### QueryRequest

```go
type QueryRequest struct {
    Query       string                 // LMQL query string
    Variables   map[string]interface{} // Input variables
    Temperature float64                // Generation temperature
    MaxTokens   int                    // Maximum tokens
}
```

### QueryResponse

```go
type QueryResponse struct {
    Result               map[string]interface{} // Extracted variables
    RawOutput            string                 // Raw LLM output
    ConstraintsSatisfied bool                   // All constraints passed
}
```

## Usage

### Basic Query

```go
import "dev.helix.agent/internal/optimization/lmql"

client := lmql.NewClient(nil)

response, err := client.Query(ctx, &lmql.QueryRequest{
    Query: `
        "Analyze this review: {text}"
        SENTIMENT: [sentiment] where sentiment in ["positive", "negative", "neutral"]
        "Confidence: "
        CONFIDENCE: [confidence] where float(confidence) >= 0.0 and float(confidence) <= 1.0
    `,
    Variables: map[string]interface{}{
        "text": "This product is amazing! Best purchase ever.",
    },
})

fmt.Printf("Sentiment: %s (%.2f confidence)\n",
    response.Result["sentiment"],
    response.Result["confidence"])
```

### Structured Extraction

```go
response, _ := client.Query(ctx, &lmql.QueryRequest{
    Query: `
        "Extract person info from: {input}"
        NAME: [name] where len(name) > 0
        AGE: [age] where int(age) > 0 and int(age) < 150
        OCCUPATION: [occupation]
    `,
    Variables: map[string]interface{}{
        "input": "John Smith, 42, Software Engineer",
    },
})

// response.Result: {"name": "John Smith", "age": "42", "occupation": "Software Engineer"}
```

### Multi-Choice Selection

```go
response, _ := client.Query(ctx, &lmql.QueryRequest{
    Query: `
        "Classify this support ticket: {ticket}"
        CATEGORY: [category] where category in ["billing", "technical", "general", "urgent"]
        PRIORITY: [priority] where priority in ["low", "medium", "high", "critical"]
        "Suggested action: "
        ACTION: [action]
    `,
    Variables: map[string]interface{}{
        "ticket": "My payment was charged twice and I need a refund immediately!",
    },
})
```

### With Temperature Control

```go
response, _ := client.Query(ctx, &lmql.QueryRequest{
    Query:       queryString,
    Variables:   vars,
    Temperature: 0.1, // More deterministic
    MaxTokens:   100,
})
```

## Query Syntax

### Variable Definition
```lmql
[variable_name]                          # Unconstrained variable
[var] where var in ["a", "b", "c"]       # Choice constraint
[num] where int(num) > 0                 # Numeric constraint
[text] where len(text) < 100             # Length constraint
```

### Constraint Operators
- `in [...]` - Choice constraint
- `int(x)` / `float(x)` - Numeric conversion
- `len(x)` - Length check
- `and` / `or` - Logical operators
- `>`, `<`, `>=`, `<=`, `==` - Comparisons

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LMQL_BASE_URL` | LMQL server URL | `http://localhost:8014` |
| `LMQL_TIMEOUT` | Request timeout | `120s` |

### Server Setup

```bash
# Start LMQL service
lmql serve --port 8014
```

## Testing

```bash
go test -v ./internal/optimization/lmql/...
```

## Files

- `client.go` - HTTP client implementation
