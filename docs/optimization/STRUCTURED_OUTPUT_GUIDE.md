# Structured Output Guide

The structured output system provides JSON schema validation and generation, inspired by Outlines.

## Overview

Structured output ensures LLM responses conform to a specified JSON schema, enabling:

- Guaranteed valid JSON responses
- Type-safe data extraction
- Automatic retry on validation failure
- Schema-based prompt enhancement

## Features

- **JSON Schema Validation**: Full JSON Schema draft-07 support
- **Schema Builder**: Fluent API for building schemas
- **Regex Validation**: Pattern-based output validation
- **Choice Constraints**: Limit outputs to predefined options
- **Auto-Retry**: Automatic retry on validation failure

## Configuration

```yaml
optimization:
  structured_output:
    enabled: true
    strict_mode: true        # Fail on invalid output
    retry_on_failure: true   # Retry with validation feedback
    max_retries: 3           # Maximum retry attempts
```

## Basic Usage

### Schema Builder

```go
import "github.com/helixagent/helixagent/internal/optimization/outlines"

// Create a simple schema
schema := outlines.ObjectSchema(map[string]*outlines.JSONSchema{
    "name":  outlines.StringSchema(),
    "age":   outlines.IntegerSchema(),
    "email": outlines.StringSchemaWithFormat("email"),
}, "name", "age") // required fields
```

### Validation

```go
// Create validator
validator, err := outlines.NewSchemaValidator(schema)

// Validate LLM response
result := validator.Validate(`{"name": "John", "age": 30}`)
if result.Valid {
    data := result.Data.(map[string]interface{})
    fmt.Printf("Name: %s, Age: %d\n", data["name"], data["age"])
} else {
    for _, err := range result.Errors {
        fmt.Println("Validation error:", err)
    }
}
```

### With OptimizationService

```go
svc, _ := optimization.NewService(config)

schema := outlines.ObjectSchema(map[string]*outlines.JSONSchema{
    "answer": outlines.StringSchema(),
    "confidence": outlines.NumberSchemaWithRange(0, 1),
}, "answer")

generator := func(prompt string) (string, error) {
    return llmProvider.Complete(ctx, prompt)
}

result, err := svc.GenerateStructured(ctx, prompt, schema, generator)
if result.Valid {
    fmt.Println("Answer:", result.ParsedData)
}
```

## Schema Types

### String

```go
// Basic string
outlines.StringSchema()

// With constraints
outlines.StringSchemaWithConstraints(5, 100)  // min/max length

// With format
outlines.StringSchemaWithFormat("email")      // email, uri, date, uuid, ipv4

// With pattern
outlines.StringSchemaWithPattern(`^\d{3}-\d{4}$`)

// With enum
outlines.StringSchemaWithEnum("low", "medium", "high")
```

### Number

```go
// Integer
outlines.IntegerSchema()

// Number (float)
outlines.NumberSchema()

// With range
outlines.NumberSchemaWithRange(0, 100)

// Integer with range
outlines.IntegerSchemaWithRange(1, 10)
```

### Boolean

```go
outlines.BooleanSchema()
```

### Array

```go
// Array of strings
outlines.ArraySchema(outlines.StringSchema())

// With constraints
outlines.ArraySchemaWithConstraints(outlines.IntegerSchema(), 1, 10) // min/max items
```

### Object

```go
outlines.ObjectSchema(
    map[string]*outlines.JSONSchema{
        "id":    outlines.IntegerSchema(),
        "name":  outlines.StringSchema(),
        "tags":  outlines.ArraySchema(outlines.StringSchema()),
    },
    "id", "name", // required fields
)
```

### Nested Objects

```go
addressSchema := outlines.ObjectSchema(map[string]*outlines.JSONSchema{
    "street": outlines.StringSchema(),
    "city":   outlines.StringSchema(),
    "zip":    outlines.StringSchemaWithPattern(`^\d{5}$`),
}, "city")

personSchema := outlines.ObjectSchema(map[string]*outlines.JSONSchema{
    "name":    outlines.StringSchema(),
    "address": addressSchema,
}, "name")
```

## Advanced Features

### OneOf / AnyOf

```go
// OneOf - exactly one must match
schema := &outlines.JSONSchema{
    OneOf: []*outlines.JSONSchema{
        outlines.StringSchema(),
        outlines.IntegerSchema(),
    },
}

// AnyOf - at least one must match
schema := &outlines.JSONSchema{
    AnyOf: []*outlines.JSONSchema{
        outlines.ObjectSchema(map[string]*outlines.JSONSchema{"type": outlines.StringSchemaWithEnum("text")}, "type"),
        outlines.ObjectSchema(map[string]*outlines.JSONSchema{"type": outlines.StringSchemaWithEnum("image")}, "type"),
    },
}
```

### Choice Generator

Constrain output to specific options:

```go
generator := outlines.NewChoiceGenerator([]string{
    "positive",
    "negative",
    "neutral",
}, true) // case-insensitive

result := generator.Generate(`The sentiment is: positive`)
// result.Match = "positive"
// result.Valid = true
```

### Regex Generator

Validate output against patterns:

```go
generator := outlines.NewRegexGenerator(`\d{4}-\d{2}-\d{2}`)

result := generator.Generate(`The date is 2024-01-15`)
// result.Match = "2024-01-15"
// result.Valid = true
```

### JSON Extraction

Extract JSON from text responses:

```go
text := `Here's the result: {"name": "John", "age": 30} That's all.`
jsonStr, err := outlines.ExtractJSON(text)
// jsonStr = `{"name": "John", "age": 30}`
```

## Structured Generator

For complete structured generation with retry:

```go
config := &outlines.GeneratorConfig{
    MaxRetries:      3,
    StrictMode:      true,
    RepairAttempts:  2,
}

gen := outlines.NewStructuredGenerator(llmProvider, schema, config)

result, err := gen.Generate(ctx, prompt)
if result.Valid {
    fmt.Println("Generated:", result.Content)
    fmt.Println("Parsed:", result.ParsedData)
}
```

## Format Validators

Built-in format validators:

| Format | Pattern | Example |
|--------|---------|---------|
| email | RFC 5322 | user@example.com |
| uri | RFC 3986 | https://example.com |
| date | ISO 8601 | 2024-01-15 |
| uuid | RFC 4122 | 550e8400-e29b-41d4-a716-446655440000 |
| ipv4 | Dotted decimal | 192.168.1.1 |

## Best Practices

1. **Start Simple**: Begin with basic schemas and add complexity as needed

2. **Use Required Fields**: Always specify required fields to catch missing data

3. **Enable Retry**: Use `retry_on_failure` for better reliability

4. **Provide Clear Prompts**: Include schema structure in the prompt for better results:
   ```
   Respond with JSON matching this schema:
   {"name": string, "age": integer}
   ```

5. **Handle Validation Errors**: Always check `result.Valid` before using data

6. **Use Appropriate Types**: Match schema types to expected data

## Error Handling

```go
result := validator.Validate(response)
if !result.Valid {
    for _, err := range result.Errors {
        switch e := err.(type) {
        case *outlines.ValidationError:
            fmt.Printf("Path: %s, Message: %s\n", e.Path, e.Message)
        default:
            fmt.Println("Error:", err)
        }
    }
}
```

## Performance Tips

1. **Cache Validators**: Reuse validators for the same schema
2. **Use Compiled Patterns**: Pre-compile regex patterns for repeated use
3. **Limit Retries**: Set reasonable `max_retries` to avoid infinite loops
4. **Schema Complexity**: Simpler schemas validate faster
