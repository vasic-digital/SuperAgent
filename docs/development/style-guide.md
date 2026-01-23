# HelixAgent Code Style Guide

## Overview

This document defines the coding standards and best practices for HelixAgent development. Following these guidelines ensures consistency, readability, and maintainability across the codebase.

## General Principles

1. **Clarity over cleverness** - Write code that is easy to understand
2. **Consistency** - Follow existing patterns in the codebase
3. **Simplicity** - Prefer simple solutions over complex ones
4. **Documentation** - Document non-obvious behavior
5. **Testing** - Write tests for all new functionality

## Go Code Style

### Formatting

All Go code must be formatted with `gofmt` or `goimports`:

```bash
# Format code
gofmt -w .

# Format with import organization
goimports -w .
```

### Naming Conventions

#### Packages

- Use short, lowercase names without underscores
- Package name should match directory name
- Avoid generic names like `util`, `common`, `misc`

```go
// Good
package cache
package providers
package middleware

// Bad
package myUtils
package common_helpers
```

#### Variables

- Use camelCase for local variables
- Use descriptive names (avoid single letters except in loops)
- Boolean variables should be positive assertions

```go
// Good
userCount := len(users)
isEnabled := config.Enabled
for i, item := range items { }

// Bad
uc := len(users)
notDisabled := !config.Disabled
for x, y := range items { }
```

#### Functions and Methods

- Use camelCase, starting with uppercase for exported
- Use verbs for action functions
- Use nouns for getter functions (without "Get" prefix)

```go
// Good
func ProcessRequest(req *Request) error { }
func (c *Cache) Size() int { }
func validateInput(input string) error { }

// Bad
func DoTheThing(req *Request) error { }
func (c *Cache) GetSize() int { }  // No "Get" prefix needed
func input_validator(input string) error { }
```

#### Interfaces

- Single-method interfaces use method name + "er" suffix
- Interfaces describe behavior, not implementation

```go
// Good
type Reader interface {
    Read(p []byte) (n int, err error)
}

type LLMProvider interface {
    Complete(ctx context.Context, req *Request) (*Response, error)
}

// Bad
type IReader interface { }  // No "I" prefix
type ProviderImpl interface { }  // Don't mention impl
```

#### Constants

- Use PascalCase for exported constants
- Group related constants with iota when appropriate

```go
// Good
const MaxRetries = 3
const DefaultTimeout = 30 * time.Second

type Priority int

const (
    PriorityLow Priority = iota
    PriorityNormal
    PriorityHigh
)

// Bad
const MAX_RETRIES = 3  // Don't use SCREAMING_CASE
```

### Error Handling

#### Error Creation

- Use lowercase error messages without trailing punctuation
- Wrap errors with context using `fmt.Errorf` with `%w`

```go
// Good
var ErrNotFound = errors.New("not found")
return fmt.Errorf("failed to connect to database: %w", err)

// Bad
var ErrNotFound = errors.New("Not Found.")  // No caps or period
return fmt.Errorf("connection failed")  // No error wrapping
```

#### Error Checking

- Always check errors immediately after function call
- Handle errors at the appropriate level

```go
// Good
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doing something: %w", err)
}

// Bad
result, _ := doSomething()  // Don't ignore errors
if result, err := doSomething(); err != nil {
    // Complex logic here
}
```

#### Custom Errors

- Use sentinel errors for expected conditions
- Use typed errors for errors needing additional context

```go
// Sentinel errors
var (
    ErrNotFound = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
)

// Typed error
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}
```

### Context

- Context should be the first parameter
- Don't store context in structs
- Respect context cancellation

```go
// Good
func ProcessRequest(ctx context.Context, req *Request) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case result := <-process(req):
        return result
    }
}

// Bad
type Service struct {
    ctx context.Context  // Don't store context
}

func ProcessRequest(req *Request, ctx context.Context) error { }  // ctx not first
```

### Concurrency

- Prefer channels over shared memory
- Always close channels from the sender
- Use sync.WaitGroup for goroutine coordination

```go
// Good
func process(items []Item) <-chan Result {
    results := make(chan Result)
    go func() {
        defer close(results)
        for _, item := range items {
            results <- processItem(item)
        }
    }()
    return results
}

// Good - WaitGroup usage
func processAll(items []Item) {
    var wg sync.WaitGroup
    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            process(item)
        }(item)
    }
    wg.Wait()
}
```

### Structs

- Order fields by size for memory efficiency (optional)
- Group related fields together
- Use meaningful zero values

```go
// Good
type Config struct {
    // Required fields
    Host string
    Port int

    // Optional with defaults
    Timeout     time.Duration // Default: 30s
    MaxRetries  int           // Default: 3
    EnableCache bool          // Default: false
}

func NewConfig(host string, port int) *Config {
    return &Config{
        Host:       host,
        Port:       port,
        Timeout:    30 * time.Second,
        MaxRetries: 3,
    }
}
```

### Comments

- Use `//` for single-line comments
- Document all exported symbols
- Explain "why", not "what"

```go
// Good
// ProcessRequest handles incoming LLM requests by distributing them
// across available providers based on their verification scores.
// It returns the best response or an error if all providers fail.
func ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
    // Use a circuit breaker to prevent cascading failures
    // when a provider is experiencing issues.
    if !circuitBreaker.Allow() {
        return nil, ErrCircuitOpen
    }
    // ...
}

// Bad
// This function processes a request
func ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
    // Check if allowed
    if !circuitBreaker.Allow() {
        return nil, ErrCircuitOpen
    }
}
```

### Testing

#### Test Files

- Test files should be in the same package
- Name test files `*_test.go`
- Use table-driven tests

```go
// cache_test.go
package cache

func TestCache_Get(t *testing.T) {
    tests := []struct {
        name     string
        key      string
        want     string
        wantErr  bool
    }{
        {
            name:    "existing key",
            key:     "foo",
            want:    "bar",
            wantErr: false,
        },
        {
            name:    "missing key",
            key:     "missing",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := NewCache()
            c.Set("foo", "bar")

            got, err := c.Get(tt.key)
            if (err != nil) != tt.wantErr {
                t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Get() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### Test Naming

- Use descriptive test names
- Prefix with Test for test functions
- Use subtests for related scenarios

```go
// Good
func TestProvider_Complete_WithValidRequest(t *testing.T) { }
func TestProvider_Complete_WithTimeout(t *testing.T) { }

// With subtests
func TestProvider_Complete(t *testing.T) {
    t.Run("valid request", func(t *testing.T) { })
    t.Run("with timeout", func(t *testing.T) { })
    t.Run("with invalid input", func(t *testing.T) { })
}
```

### Package Organization

```
internal/
├── cache/
│   ├── doc.go           # Package documentation
│   ├── cache.go         # Main implementation
│   ├── cache_test.go    # Tests
│   ├── redis.go         # Redis implementation
│   ├── memory.go        # In-memory implementation
│   └── options.go       # Configuration options
├── llm/
│   ├── doc.go
│   ├── provider.go      # Interface definition
│   ├── ensemble.go      # Ensemble implementation
│   └── providers/
│       ├── claude/
│       ├── deepseek/
│       └── ...
```

## API Design

### HTTP Endpoints

- Use RESTful conventions
- Use plural nouns for resources
- Use HTTP methods correctly

```
GET    /v1/completions       # List completions
POST   /v1/completions       # Create completion
GET    /v1/completions/:id   # Get specific completion
DELETE /v1/completions/:id   # Delete completion

POST   /v1/debates           # Create debate
GET    /v1/debates/:id       # Get debate status
```

### Request/Response

- Use consistent JSON field naming (snake_case)
- Include pagination for list endpoints
- Return appropriate HTTP status codes

```go
type CompletionRequest struct {
    Prompt      string  `json:"prompt" binding:"required"`
    Model       string  `json:"model,omitempty"`
    MaxTokens   int     `json:"max_tokens,omitempty"`
    Temperature float64 `json:"temperature,omitempty"`
}

type CompletionResponse struct {
    ID         string `json:"id"`
    Content    string `json:"content"`
    Model      string `json:"model"`
    TokensUsed int    `json:"tokens_used"`
    CreatedAt  string `json:"created_at"`
}
```

## Commit Messages

Follow conventional commits format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `test`: Tests
- `chore`: Maintenance

Examples:
```
feat(providers): add Mistral provider support

- Implement Complete and CompleteStream methods
- Add rate limiting and retry logic
- Include comprehensive tests

Closes #123
```

```
fix(cache): resolve race condition in concurrent access

Use sync.RWMutex instead of sync.Mutex for better
read performance while maintaining write safety.
```

## Code Review Checklist

- [ ] Code follows style guide
- [ ] Tests are included and pass
- [ ] Documentation is updated
- [ ] Error handling is appropriate
- [ ] No security vulnerabilities
- [ ] Performance is acceptable
- [ ] Backward compatibility maintained

---

**Document Version**: 1.0
**Last Updated**: January 23, 2026
