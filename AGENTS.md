# SuperAgent Development Guidelines

## Build/Lint/Test Commands
```bash
# Build
make build                 # Build SuperAgent binary
make build-debug           # Build with debug symbols
make build-all            # Build for all architectures (linux/darwin/windows, amd64/arm64)

# Run
make run                  # Run SuperAgent locally
make run-dev              # Run in development mode

# Test (6-tier testing strategy)
make test                 # Run all tests
make test-coverage        # Run tests with coverage report (generates coverage.html)
make test-coverage-100    # Run tests with 100% coverage requirement
make test-unit            # Unit tests only: go test -v ./internal/... -short
make test-integration     # Integration tests only: go test -v ./tests/integration
make test-e2e             # E2E tests only: go test -v ./tests/e2e
make test-security        # Security tests only: go test -v ./tests/security
make test-stress          # Stress tests only: go test -v ./tests/stress
make test-chaos           # Chaos tests only: go test -v ./tests/challenge
make test-all-types       # Run all 6 test types sequentially
make test-bench           # Run benchmark tests
make test-race            # Run tests with race detection

# Run single test
go test -v ./internal/handlers -run TestCompletionHandler_Complete_Success
go test -v ./internal/llm/providers -run TestOllamaProvider
go test -v ./internal/handlers -run "TestConvert.*"  # Run matching tests

# Code Quality
make fmt                  # Format Go code (gofmt)
make vet                  # Run go vet
make lint                 # Run golangci-lint
make security-scan        # Run gosec security scan
make all                  # fmt + vet + lint + test + build
```

## Code Style Guidelines

### General
- **Go version**: Go 1.23+ required (toolchain go1.24.11)
- **Formatting**: Always run `make fmt` before committing (uses gofmt)
- **Vetting**: Always run `make vet` to catch potential issues
- **Linting**: Use golangci-lint for comprehensive linting

### Import Organization
```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "time"

    // 2. Third-party packages
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"

    // 3. Internal packages
    "github.com/superagent/superagent/internal/models"
    "github.com/superagent/superagent/internal/services"
)
```

### Naming Conventions
- **Exports**: CamelCase (e.g., `CompletionHandler`, `ProcessRequest`)
- **Private**: camelCase (e.g., `requestService`, `convertToAPIResponse`)
- **Constants**: UPPER_SNAKE_CASE (e.g., `MAX_RETRIES`, `DEFAULT_TIMEOUT`)
- **Interfaces**: Simple, focused names ending in interface type behavior (e.g., `LLMProvider`)
- **Test Functions**: `Test<Struct>_<Method>_<Scenario>` (e.g., `TestCompletionHandler_Complete_Success`)

### Error Handling
```go
// Use structured error wrapping with context
if err != nil {
    return fmt.Errorf("failed to process request: %w", err)
}

// Use AppError for HTTP handler errors
appErr := utils.NewAppError("INVALID_REQUEST", "Missing required field", http.StatusBadRequest, err)
utils.HandleError(c, appErr)

// Always handle errors, never ignore them
result, err := someFunction()
if err != nil {
    logger.WithFields(logrus.Fields{
        "error": err,
        "context": "processing_request",
    }).Error("Failed to process")
    return err
}
```

### Structured Logging
```go
// Always include context in logs
logger.WithFields(logrus.Fields{
    "request_id": req.ID,
    "user_id": userID,
    "provider": providerName,
}).Info("Processing request")

// Use appropriate log levels
logger.Debug("Detailed debugging info")
logger.Info("Normal operation")
logger.Warn("Warning condition")
logger.Error("Error occurred")
```

### Testing
```go
// Use testify for assertions
import "github.com/stretchr/testify/assert"

func TestSomething_SpecificScenario(t *testing.T) {
    // Arrange
    handler := NewCompletionHandler(nil)
    req := &CompletionRequest{Prompt: "test"}

    // Act
    result, err := handler.Process(req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "expected", result.Value)
}

// Test naming: Test<Struct>_<Method>_<Scenario>
// Minimum 95% coverage required
// Use httptest for HTTP testing
```

### Interface Design
- Keep interfaces small and focused
- Design for testability (dependency injection)
- Accept interfaces, return structs
- Example: `LLMProvider` interface with `Complete()`, `CompleteStream()`, `HealthCheck()`

### Type Definitions
```go
// Struct fields should be documented
type CompletionHandler struct {
    requestService *services.RequestService
}

// JSON tags for API types
type CompletionRequest struct {
    Prompt      string `json:"prompt" binding:"required"`
    Model       string `json:"model,omitempty"`
    Temperature float64 `json:"temperature,omitempty"`
}

// Use pointer types for optional fields in structs
```

### Context Usage
- Always pass `context.Context` as first parameter in functions that do I/O
- Use context for request-scoped values, timeouts, cancellation
- Example: `func (h *Handler) Process(ctx context.Context, req *Request) error`

### HTTP Handler Pattern
```go
func (h *Handler) Handle(c *gin.Context) {
    // 1. Parse request
    var req RequestType
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.HandleError(c, err)
        return
    }

    // 2. Process (pass context)
    result, err := h.service.Process(c.Request.Context(), req)
    if err != nil {
        utils.HandleError(c, err)
        return
    }

    // 3. Return response
    c.JSON(http.StatusOK, result)
}
```

### Dependencies
- **Web Framework**: github.com/gin-gonic/gin
- **Logging**: github.com/sirupsen/logrus
- **Testing**: github.com/stretchr/testify
- **Database**: github.com/jackc/pgx/v5 (PostgreSQL)
- **Caching**: github.com/redis/go-redis/v9
- **Metrics**: github.com/prometheus/client_golang

### Documentation
- Document all exported functions/types with examples
- Use godoc comments: `// FunctionName does something...`
- Run `make docs` to serve documentation at http://localhost:6060

### Security
- Never commit secrets or credentials
- Use environment variables for configuration
- Run `make security-scan` to check for security issues
- Validate all user inputs
- Use parameterized queries for database operations
