# Outlines Package

This package provides structured output generation with JSON schema validation for LLM responses.

## Overview

The Outlines package ensures LLM outputs conform to predefined JSON schemas, enabling reliable structured data extraction from natural language responses.

## Components

### Structured Generator (`generator.go`)

Generates structured output conforming to a schema:

```go
schema := &outlines.JSONSchema{
    Type: "object",
    Properties: map[string]*outlines.JSONSchema{
        "name":  {Type: "string"},
        "age":   {Type: "integer"},
        "email": {Type: "string", Format: "email"},
    },
    Required: []string{"name", "age"},
}

generator, _ := outlines.NewStructuredGenerator(llmProvider, schema, nil)
result, _ := generator.Generate(ctx, "Extract user info from: John is 30 years old")
```

### Schema Validation (`validator.go`)

Validates generated output against JSON schemas:

```go
validator, _ := outlines.NewSchemaValidator(schema)
errors := validator.Validate(jsonData)
if len(errors) > 0 {
    // Handle validation errors
}
```

### Schema Definition (`schema.go`)

JSON Schema type definitions:

```go
type JSONSchema struct {
    Type        string                 // string, integer, number, boolean, array, object
    Properties  map[string]*JSONSchema // Object properties
    Items       *JSONSchema            // Array item schema
    Required    []string               // Required fields
    Enum        []interface{}          // Allowed values
    Format      string                 // email, date-time, uri, etc.
    MinLength   *int                   // String minimum length
    MaxLength   *int                   // String maximum length
    Minimum     *float64               // Number minimum
    Maximum     *float64               // Number maximum
    Pattern     string                 // Regex pattern
}
```

## Data Types

### StructuredResponse

```go
type StructuredResponse struct {
    Content    string      // Raw generated content
    ParsedData interface{} // Parsed JSON data
    Valid      bool        // Validation passed
    Errors     []string    // Validation errors
    Retries    int         // Number of retry attempts
}
```

### GeneratorConfig

```go
type GeneratorConfig struct {
    MaxRetries            int  // Maximum retry attempts (default: 3)
    StrictMode            bool // Fail immediately on error
    IncludeSchemaInPrompt bool // Include schema in prompt
    IncludeErrorFeedback  bool // Include errors in retry prompts
}
```

## Usage

### Basic Structured Generation

```go
import "dev.helix.agent/internal/optimization/outlines"

// Define the expected output schema
schema := &outlines.JSONSchema{
    Type: "object",
    Properties: map[string]*outlines.JSONSchema{
        "sentiment": {
            Type: "string",
            Enum: []interface{}{"positive", "negative", "neutral"},
        },
        "confidence": {
            Type:    "number",
            Minimum: ptr(0.0),
            Maximum: ptr(1.0),
        },
    },
    Required: []string{"sentiment", "confidence"},
}

// Create generator
generator, err := outlines.NewStructuredGenerator(llmProvider, schema, nil)

// Generate structured output
result, err := generator.Generate(ctx, "Analyze sentiment: I love this product!")
if result.Valid {
    // Access parsed data
    data := result.ParsedData.(map[string]interface{})
    sentiment := data["sentiment"].(string)
    confidence := data["confidence"].(float64)
}
```

### With Retry Logic

```go
config := &outlines.GeneratorConfig{
    MaxRetries:           5,
    IncludeErrorFeedback: true,
}

generator, _ := outlines.NewStructuredGenerator(llmProvider, schema, config)
result, _ := generator.Generate(ctx, prompt)
// Generator automatically retries on validation failure
```

### Complex Nested Schemas

```go
schema := &outlines.JSONSchema{
    Type: "object",
    Properties: map[string]*outlines.JSONSchema{
        "users": {
            Type: "array",
            Items: &outlines.JSONSchema{
                Type: "object",
                Properties: map[string]*outlines.JSONSchema{
                    "id":    {Type: "string"},
                    "roles": {Type: "array", Items: &outlines.JSONSchema{Type: "string"}},
                },
            },
        },
    },
}
```

## Validation Features

- **Type Checking**: Validates JSON types match schema
- **Required Fields**: Ensures required properties exist
- **Enum Validation**: Restricts values to allowed set
- **Format Validation**: email, date-time, uri, uuid
- **Range Validation**: min/max for numbers and strings
- **Pattern Matching**: Regex pattern validation

## Testing

```bash
go test -v ./internal/optimization/outlines/...
```

## Files

- `generator.go` - Structured generator implementation
- `schema.go` - JSON schema type definitions
- `validator.go` - Schema validation logic
