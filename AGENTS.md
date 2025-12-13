# SuperAgent Development Guidelines

## Build/Lint/Test Commands
```bash
# Build
make build                 # Build SuperAgent binary
make build-all            # Build for all architectures

# Test (6-tier testing strategy)
make test                 # Run all tests
make test-unit            # Unit tests only: go test -v ./internal/... -short
make test-integration     # Integration tests only: go test -v ./tests/integration
make test-e2e             # E2E tests only: go test -v ./tests/e2e
make test-security        # Security tests only: go test -v ./tests/security
make test-stress          # Stress tests only: go test -v ./tests/stress
make test-chaos           # Chaos tests only: go test -v ./tests/challenge
make test-coverage        # Run tests with coverage report

# Code Quality
make fmt                  # Format Go code (gofmt)
make vet                  # Run go vet
make lint                 # Run golangci-lint
make security-scan        # Run gosec security scan
```

## Code Style Guidelines
- **Go 1.23+ required** - Use standard Go conventions (gofmt, go vet)
- **Imports** - Group imports (stdlib, third-party, internal), no unused imports
- **Naming** - Use CamelCase for exports, camelCase for private, constants UPPER_SNAKE_CASE
- **Error Handling** - Always handle errors, use structured error wrapping with context
- **Logging** - Use structured logging with logrus, include context in all logs
- **Testing** - Minimum 95% coverage, use testify for assertions, include benchmarks
- **Interfaces** - Design for testability, keep interfaces small and focused
- **Documentation** - Document all exported functions/types with examples
