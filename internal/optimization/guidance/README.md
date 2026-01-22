# Guidance Package

This package provides constrained generation capabilities for LLM outputs, ensuring generated text follows specified patterns and constraints.

## Overview

The Guidance package enables precise control over LLM output through constraints like regex patterns, choices, length limits, and custom validators.

## Features

- **Pattern Constraints**: Regex-based output validation
- **Choice Constraints**: Select from predefined options
- **Length Constraints**: Min/max character limits
- **Composition**: Combine multiple constraints
- **Retry Logic**: Automatic retry on constraint violation

## Components

### Generator (`generator.go`)

Constrained text generation:

```go
generator := guidance.NewConstrainedGenerator(llmBackend, nil)

result, err := generator.Generate(ctx, "Generate a UUID", guidance.PatternConstraint(`[a-f0-9-]{36}`))
// Result guaranteed to match UUID pattern
```

### Constraints (`constraints.go`)

Built-in constraint types:

```go
// Pattern constraint
pattern := guidance.PatternConstraint(`\d{3}-\d{3}-\d{4}`)

// Choice constraint
choice := guidance.ChoiceConstraint("yes", "no", "maybe")

// Length constraint
length := guidance.LengthConstraint(10, 100)

// Composite constraint
combined := guidance.And(pattern, length)
```

## Data Types

### Constraint

```go
type Constraint interface {
    Validate(output string) []string // Returns validation errors
    Description() string             // Human-readable description
    Type() ConstraintType           // Constraint type identifier
}
```

### GenerationResult

```go
type GenerationResult struct {
    Output           string            // Generated text
    Valid            bool              // Passed all constraints
    Attempts         int               // Generation attempts
    ValidationErrors []string          // Any validation errors
    Metadata         *GenerationMetadata // Generation metadata
}
```

### GenerationMetadata

```go
type GenerationMetadata struct {
    Model          string         // Model used
    Provider       string         // Provider name
    TokensUsed     int            // Tokens consumed
    LatencyMs      int64          // Generation latency
    ConstraintType ConstraintType // Constraint applied
    Timestamp      time.Time      // Completion time
}
```

## Usage

### Pattern-Based Generation

```go
import "dev.helix.agent/internal/optimization/guidance"

generator := guidance.NewConstrainedGenerator(llmBackend, nil)

// Generate email address
result, _ := generator.Generate(ctx,
    "Generate a valid email address",
    guidance.PatternConstraint(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
)
// result.Output: "user@example.com"
```

### Choice-Based Generation

```go
// Force specific category selection
result, _ := generator.Generate(ctx,
    "Classify this text sentiment: 'I love this product!'",
    guidance.ChoiceConstraint("positive", "negative", "neutral"),
)
// result.Output: "positive"
```

### Composite Constraints

```go
// Combine multiple constraints
constraint := guidance.And(
    guidance.LengthConstraint(50, 200),
    guidance.PatternConstraint(`^[A-Z].*\.$`), // Start caps, end period
    guidance.Not(guidance.ContainsConstraint("TODO")),
)

result, _ := generator.Generate(ctx, "Write a summary paragraph", constraint)
```

### With Retry Logic

```go
config := &guidance.GeneratorConfig{
    MaxRetries:     5,
    IncludeFeedback: true,
}

generator := guidance.NewConstrainedGenerator(llmBackend, config)

result, _ := generator.GenerateWithRetry(ctx, prompt, constraint, 5)
fmt.Printf("Valid after %d attempts\n", result.Attempts)
```

### JSON Structure Constraint

```go
// Ensure valid JSON output
jsonConstraint := guidance.JSONConstraint(&guidance.JSONSchema{
    Type: "object",
    Properties: map[string]*guidance.JSONSchema{
        "name": {Type: "string"},
        "age":  {Type: "integer"},
    },
    Required: []string{"name"},
})

result, _ := generator.Generate(ctx, "Extract person info from: John is 30", jsonConstraint)
```

## Constraint Types

| Type | Description | Example |
|------|-------------|---------|
| `Pattern` | Regex match | `\d{4}-\d{2}-\d{2}` for dates |
| `Choice` | Select from options | `["yes", "no"]` |
| `Length` | Character limits | Min 10, Max 100 |
| `Contains` | Must include text | Contains "APPROVED" |
| `JSON` | Valid JSON structure | Object with required fields |
| `And` | All must pass | Pattern AND Length |
| `Or` | Any must pass | Choice OR Pattern |
| `Not` | Must not match | Not Contains "error" |

## Testing

```bash
go test -v ./internal/optimization/guidance/...
```

## Files

- `generator.go` - Constrained generator implementation
- `constraints.go` - Constraint type implementations
- `types.go` - Type definitions
- `client.go` - External guidance service client
