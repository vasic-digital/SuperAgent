# Guidance & LMQL Integration Guide

Guidance and LMQL provide advanced constrained generation capabilities.

## Overview

| Tool | Purpose | Strength |
|------|---------|----------|
| **Guidance** | CFG-based constraints | Structured templates |
| **LMQL** | Query language | Flexible constraints |

## Guidance

Guidance uses context-free grammars (CFG) for structured generation.

### Docker Setup

```bash
docker-compose --profile optimization up -d guidance-server
curl http://localhost:8013/health
```

### Configuration

```yaml
optimization:
  guidance:
    enabled: true
    endpoint: "http://localhost:8013"
    timeout: "120s"
    cache_programs: true
```

### Client Initialization

```go
import "github.com/helixagent/helixagent/internal/optimization/guidance"

config := &guidance.ClientConfig{
    BaseURL: "http://localhost:8013",
    Timeout: 120 * time.Second,
}

client := guidance.NewClient(config)
```

### Grammar-Based Generation

```go
response, err := client.GenerateWithGrammar(ctx, &guidance.GrammarRequest{
    Prompt: "Generate a JSON object with name and age:",
    Grammar: `
        start: object
        object: "{" pair ("," pair)* "}"
        pair: string ":" value
        string: "\"" /[a-zA-Z]+/ "\""
        value: string | number
        number: /[0-9]+/
    `,
})

fmt.Println("Generated:", response.Output)
// {"name":"John","age":30}
```

### Template-Based Generation

```go
response, err := client.GenerateFromTemplate(ctx, &guidance.TemplateRequest{
    Template: `
        Name: {{gen "name" max_tokens=20}}
        Age: {{gen "age" pattern="[0-9]+"}}
        City: {{select "city" options=["NYC", "LA", "Chicago"]}}
    `,
    Variables: map[string]interface{}{
        "context": "Generate a person profile",
    },
})

fmt.Printf("Name: %s, Age: %s, City: %s\n",
    response.Variables["name"],
    response.Variables["age"],
    response.Variables["city"])
```

### Selection from Options

```go
selected, err := client.SelectOne(ctx, "The best programming language is:", []string{
    "Python",
    "JavaScript",
    "Go",
    "Rust",
})

fmt.Println("Selected:", selected)
```

### Regex-Constrained Generation

```go
response, err := client.GenerateWithRegex(ctx, &guidance.RegexRequest{
    Prompt:  "Generate a phone number:",
    Pattern: `\(\d{3}\) \d{3}-\d{4}`,
})

fmt.Println("Phone:", response.Output)
// (555) 123-4567
```

### JSON Schema Generation

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "name":  map[string]interface{}{"type": "string"},
        "email": map[string]interface{}{"type": "string", "format": "email"},
        "age":   map[string]interface{}{"type": "integer", "minimum": 0},
    },
    "required": []string{"name", "email"},
}

response, err := client.GenerateJSON(ctx, &guidance.JSONSchemaRequest{
    Prompt: "Generate a user profile:",
    Schema: schema,
})

fmt.Println("JSON:", response.Output)
```

## LMQL

LMQL provides a query language for constrained LLM generation.

### Docker Setup

```bash
docker-compose --profile optimization up -d lmql-server
curl http://localhost:8014/health
```

### Configuration

```yaml
optimization:
  lmql:
    enabled: true
    endpoint: "http://localhost:8014"
    timeout: "120s"
    cache_queries: true
```

### Client Initialization

```go
import "github.com/helixagent/helixagent/internal/optimization/lmql"

config := &lmql.ClientConfig{
    BaseURL: "http://localhost:8014",
    Timeout: 120 * time.Second,
}

client := lmql.NewClient(config)
```

### Query Execution

```go
response, err := client.ExecuteQuery(ctx, &lmql.QueryRequest{
    Query: `
        argmax
            "Name: [NAME]\n"
            "Age: [AGE]\n"
        from
            "Generate a person profile:"
        where
            len(NAME) < 20 and
            AGE in ["20", "30", "40", "50"]
    `,
    Variables: map[string]interface{}{
        "context": "business professional",
    },
})

fmt.Printf("NAME: %s, AGE: %s\n",
    response.Result["NAME"],
    response.Result["AGE"])
```

### Constrained Generation

```go
response, err := client.GenerateConstrained(ctx, &lmql.ConstrainedRequest{
    Prompt: "The capital of France is",
    Constraints: []lmql.Constraint{
        {Type: "max_length", Value: "50"},
        {Type: "contains", Value: "Paris"},
        {Type: "not_contains", Value: "London"},
    },
})

fmt.Println("Response:", response.Text)
fmt.Println("All constraints satisfied:", response.AllSatisfied)
```

### Constraint Types

| Type | Description | Example Value |
|------|-------------|---------------|
| max_length | Maximum character length | "100" |
| contains | Must contain string | "Paris" |
| not_contains | Must not contain | "error" |
| regex | Must match pattern | `\d{4}` |
| starts_with | Must start with | "The" |
| ends_with | Must end with | "." |

### Convenience Methods

```go
// Generate with max length
response, err := client.GenerateWithMaxLength(ctx, prompt, 100)

// Generate containing specific text
response, err := client.GenerateContaining(ctx, prompt, []string{"important", "key"})

// Generate matching pattern
response, err := client.GenerateWithPattern(ctx, prompt, `\d{4}-\d{2}-\d{2}`)
```

### Decoding Strategies

```go
// Greedy (argmax)
output, err := client.DecodeGreedy(ctx, prompt)

// Sample multiple outputs
outputs, err := client.DecodeSample(ctx, prompt, 5, 0.7)

// Beam search
outputs, err := client.DecodeBeam(ctx, prompt, 3)
```

### Completion Scoring

```go
completions := []string{
    "Paris is the capital of France.",
    "London is the capital of France.",
    "Berlin is the capital of France.",
}

result, err := client.ScoreCompletions(ctx, "The capital of France is", completions)

fmt.Println("Scores:", result.Scores)
fmt.Println("Best:", result.Ranking[0])
```

### Select Best Completion

```go
best, err := client.SelectBestCompletion(ctx, prompt, completions)
fmt.Println("Best completion:", best)
```

## Integration with OptimizationService

### Using Guidance

```go
config := optimization.DefaultConfig()
config.Guidance.Enabled = true

svc, err := optimization.NewService(config)

selected, err := svc.SelectFromOptions(ctx, "Choose the best approach:", []string{
    "Option A: Fast but expensive",
    "Option B: Slow but cheap",
    "Option C: Balanced approach",
})
```

### Using LMQL

```go
config := optimization.DefaultConfig()
config.LMQL.Enabled = true

svc, err := optimization.NewService(config)

response, err := svc.GenerateConstrained(ctx, prompt, []lmql.Constraint{
    {Type: "max_length", Value: "200"},
    {Type: "contains", Value: "conclusion"},
})
```

## Comparison: Guidance vs LMQL

| Feature | Guidance | LMQL |
|---------|----------|------|
| **Approach** | Templates + Grammars | Query Language |
| **Best For** | Structured forms | Flexible constraints |
| **Learning Curve** | Lower | Higher |
| **Grammar Support** | CFG | Basic |
| **Variable Binding** | Yes | Yes |
| **Selection** | Built-in | Custom |
| **Regex** | Yes | Yes |
| **JSON Schema** | Yes | Via constraints |

## Use Case Recommendations

| Use Case | Recommended Tool |
|----------|------------------|
| Form-like output | Guidance (templates) |
| Multiple choice | Guidance (select) |
| Complex constraints | LMQL |
| JSON generation | Both work well |
| Regex patterns | Both work well |
| Scoring alternatives | LMQL |
| Grammar-based | Guidance |

## Best Practices

### Guidance

1. **Cache Programs**: Enable `cache_programs` for repeated templates
2. **Use Select for Options**: More reliable than open generation
3. **Combine Constraints**: Use templates with regex for complex formats

### LMQL

1. **Cache Queries**: Enable `cache_queries` for repeated constraints
2. **Score When Unsure**: Use scoring to evaluate multiple possibilities
3. **Layer Constraints**: Start broad, add specific constraints

## Troubleshooting

### Guidance: Invalid Grammar

```go
// Test grammar separately
_, err := client.GenerateWithGrammar(ctx, &guidance.GrammarRequest{
    Grammar: simpleGrammar,
    Prompt:  "test",
})
if err != nil {
    // Fix grammar syntax
}
```

### LMQL: Constraints Not Satisfied

```go
response, err := client.GenerateConstrained(ctx, request)
for _, result := range response.ConstraintsChecked {
    if !result.Satisfied {
        fmt.Printf("Failed: %s - %s\n", result.Type, result.Value)
    }
}
```

### Service Unavailable

```go
// Check availability
if !client.IsAvailable(ctx) {
    // Fall back to unconstrained generation
    log.Warn("Guidance unavailable, using unconstrained generation")
}
```

## API Endpoints

### Guidance

| Endpoint | Method | Purpose |
|----------|--------|---------|
| /health | GET | Health check |
| /grammar | POST | Grammar generation |
| /template | POST | Template generation |
| /select | POST | Option selection |
| /regex | POST | Regex generation |
| /json_schema | POST | JSON schema generation |

### LMQL

| Endpoint | Method | Purpose |
|----------|--------|---------|
| /health | GET | Health check |
| /query | POST | Query execution |
| /constrained | POST | Constrained generation |
| /decode | POST | Custom decoding |
| /score | POST | Completion scoring |
