# Package: structured

## Overview

The `structured` package provides constrained output generation capabilities using XGrammar-style techniques. It ensures LLM outputs conform to specified schemas (JSON, regex, grammar) for reliable structured data extraction.

## Architecture

```
structured/
├── types.go          # Schema types and constraints
├── generator.go      # Constrained generation engine
├── grammar.go        # Grammar definitions
└── structured_test.go # Unit tests (52.2% coverage)
```

## Features

- **JSON Schema Validation**: Enforce JSON schema compliance
- **Regex Constraints**: Output matching regex patterns
- **Grammar Constraints**: Context-free grammar rules
- **Type Coercion**: Automatic type conversion

## Key Types

### OutputSchema

```go
type OutputSchema struct {
    Type       SchemaType
    JSONSchema *JSONSchema      // For JSON output
    Regex      string           // For regex constraints
    Grammar    *Grammar         // For CFG constraints
    Examples   []string         // Few-shot examples
}
```

### SchemaType

```go
const (
    SchemaTypeJSON    SchemaType = "json"
    SchemaTypeRegex   SchemaType = "regex"
    SchemaTypeGrammar SchemaType = "grammar"
    SchemaTypeEnum    SchemaType = "enum"
)
```

## Usage

### JSON Schema Constraints

```go
import "dev.helix.agent/internal/structured"

// Define schema
schema := &structured.OutputSchema{
    Type: structured.SchemaTypeJSON,
    JSONSchema: &structured.JSONSchema{
        Type: "object",
        Properties: map[string]*structured.JSONSchema{
            "name": {Type: "string"},
            "age":  {Type: "integer", Minimum: 0},
            "tags": {Type: "array", Items: &structured.JSONSchema{Type: "string"}},
        },
        Required: []string{"name", "age"},
    },
}

// Generate with constraints
generator := structured.NewGenerator(provider, config)
result, err := generator.Generate(ctx, prompt, schema)

// result is guaranteed to match schema
var person Person
json.Unmarshal([]byte(result), &person)
```

### Enum Constraints

```go
schema := &structured.OutputSchema{
    Type: structured.SchemaTypeEnum,
    Enum: []string{"positive", "negative", "neutral"},
}

sentiment, err := generator.Generate(ctx, "Analyze sentiment: I love this!", schema)
// sentiment == "positive"
```

### Regex Constraints

```go
// Email extraction
schema := &structured.OutputSchema{
    Type:  structured.SchemaTypeRegex,
    Regex: `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
}

email, err := generator.Generate(ctx, "Extract email from: Contact me at john@example.com", schema)
// email == "john@example.com"
```

### Grammar Constraints

```go
// SQL query generation
grammar := structured.NewSQLGrammar()
schema := &structured.OutputSchema{
    Type:    structured.SchemaTypeGrammar,
    Grammar: grammar,
}

query, err := generator.Generate(ctx, "Get all users over 18", schema)
// query == "SELECT * FROM users WHERE age > 18"
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| MaxRetries | int | 3 | Retries on validation failure |
| Temperature | float64 | 0.0 | Lower for determinism |
| StrictMode | bool | true | Fail on invalid output |
| ValidateOutput | bool | true | Validate before returning |

## Testing

```bash
go test -v ./internal/structured/...
go test -cover ./internal/structured/...
```

## Dependencies

### Internal
- `internal/llm` - LLM providers

### External
- `github.com/xeipuuv/gojsonschema` - JSON Schema validation

## See Also

- [XGrammar Paper](https://arxiv.org/abs/2312.12345)
- [Outlines Library](https://github.com/outlines-dev/outlines)
