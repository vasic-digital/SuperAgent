# AGENTS.md

Guidance for AI coding agents working in the HelixAgent repository.

## Project Overview

HelixAgent is an AI-powered ensemble LLM service in Go (1.24+) that aggregates responses from multiple language models. It provides OpenAI-compatible APIs with 22+ LLM providers, debate orchestration, MCP adapters, and containerized infrastructure.

## Build Commands

```bash
make build              # Build binary (output in bin/)
make build-debug        # Build with debug symbols
make build-all          # Build for all platforms (Linux, macOS, Windows)
make run                # Run locally
make run-dev            # Development mode (GIN_MODE=debug)
make docker-build       # Build Docker image
make docker-run         # Start services with Docker Compose
```

## Linting & Formatting

```bash
make fmt                # Format with go fmt
make vet                # Run go vet static analysis
make lint               # Run golangci-lint (install: make install-deps)
make security-scan      # Run gosec security scanner
```

**Always run `make fmt vet lint` before committing.**

## Testing Commands

### Running Tests

```bash
make test               # All tests (verbose)
make test-unit          # Unit tests only (./internal/... -short)
make test-integration   # Integration tests with Docker deps
make test-e2e           # End-to-end tests
make test-coverage      # Tests with HTML coverage report
make test-bench         # Benchmark tests
make test-race          # Tests with race detection
```

### Running a Single Test

```bash
# Run specific test function
go test -v -run TestFunctionName ./path/to/package

# Run all tests in a package
go test -v ./internal/llm

# Run tests matching pattern
go test -v -run "Test.*Integration" ./...

# With coverage for single package
go test -v -coverprofile=coverage.out ./internal/llm
go tool cover -html=coverage.out

# Resource-limited test (CRITICAL for host stability)
GOMAXPROCS=2 go test -v -p 1 -run TestName ./path/to/package
```

### Test Infrastructure

**IMPORTANT: Container infrastructure is handled AUTOMATICALLY by HelixAgent during boot.**

The HelixAgent binary (`./bin/helixagent`) orchestrates ALL containers based on `Containers/.env`. There is NO need to manually start test infrastructure.

If you need containers for tests:
1. Run `./bin/helixagent` first - it will deploy containers to local or remote hosts
2. Then run your tests against the running containers

**Legacy commands (may not work without manual container setup):**
```bash
make test-infra-start   # Start PostgreSQL, Redis, Mock LLM containers (DEPRECATED)
make test-infra-stop    # Stop test infrastructure
make test-with-infra    # Run tests with Docker infrastructure
```

**Prefer:** Run `./bin/helixagent` and then execute tests against the running service.

## Code Style Guidelines

### Formatting & Imports
- Use `gofmt` / `goimports` for formatting
- Imports grouped: standard library, third-party, internal (blank line separated)
- Line length: ≤ 100 characters (readability first)

### Naming Conventions
- `camelCase`: local variables, private functions
- `PascalCase`: exported functions, types, constants, fields
- `UPPER_SNAKE_CASE`: exported constants
- Acronyms all caps: `HTTP`, `URL`, `ID`, `JSON`
- Receiver names: 1-2 letters (`s` for service, `c` for client)

### Error Handling
```go
// Always check errors
if err != nil {
    return err
}

// Wrap with context
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Use defer for cleanup
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()
```

### Types & Interfaces
- Use `interface` to define behavior, not data
- Prefer small, focused interfaces (`io.Reader`, `io.Writer`)
- Use struct tags for JSON, YAML, database mapping
- Avoid `any`/`interface{}`; use generics or specific types

### Concurrency
- Always use `context.Context` for cancellation/timeout
- Protect shared data with `sync.Mutex` or `sync.RWMutex`
- Use `sync.WaitGroup` for goroutine coordination

### Testing Patterns
- Write table-driven tests
- Use `testify` assertion library
- Place test files in same package with `_test.go` suffix
- Use `testdata/` directories for fixtures
- Mocks/stubs ONLY in unit tests; integration tests use real services

## Key Conventions

### Tool Schema
All tool parameters use **snake_case** (e.g., `file_path`, `old_string`). See `internal/tools/schema.go`.

### No Comments
**DO NOT ADD COMMENTS** in code unless explicitly requested.

### Git Operations
- **SSH ONLY** for all Git operations - HTTPS is forbidden
- Branch naming: `feat/`, `fix/`, `chore/`, `docs/`, `refactor/`, `test/` + description
- Commits: Conventional Commits (`feat(scope): description`)
- Run `make fmt vet lint` before committing

### Containerization

**CRITICAL: ALL container orchestration is handled AUTOMATICALLY by the HelixAgent binary during its boot process.**

- **DO NOT** manually start, stop, or manage containers via `docker` or `podman` commands
- **DO NOT** run `make test-infra-start` or similar manual container orchestration commands
- **DO NOT** attempt to deploy containers to remote hosts manually via SSH

**The ONLY correct workflow is:**
1. Build the binary: `make build`
2. Run the binary: `./bin/helixagent` (it handles ALL container orchestration automatically)
3. The binary reads `Containers/.env` and orchestrates containers locally OR remotely based on configuration

**Container Orchestration Flow (handled by HelixAgent):**
1. HelixAgent boots and initializes Containers module adapter
2. Adapter reads `Containers/.env` file (NOT project root `.env`)
3. Based on `Containers/.env`:
   - `CONTAINERS_REMOTE_ENABLED=true` → ALL containers to remote host(s) via `CONTAINERS_REMOTE_HOST_*` vars
   - `CONTAINERS_REMOTE_ENABLED=false` → ALL containers locally
4. Health checks performed against configured endpoints
5. Required services failing health check cause boot failure in strict mode

**Key Files:**
- `Containers/.env` - Container orchestration configuration
- `internal/services/boot_manager.go` - Boot orchestration logic
- `tests/precondition/containers_boot_test.go` - Precondition tests for container state

**If you need to run tests that require containers:**
- Simply run `./bin/helixagent` first - it will start all required containers
- Or run tests that use the HelixAgent binary's built-in container management

**Rebuild containers only after code changes:**
```bash
make docker-build && make docker-run  # Only if you changed containerized code
```

## Resource Limits (CRITICAL)

**ALL test execution MUST be limited to 30-40% of host resources:**

```bash
# Pattern for resource-limited execution
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -v -p 1 ./...
```

Host runs mission-critical processes; exceeding limits has caused system crashes.

## Quick Reference

| Task | Command |
|------|---------|
| Build | `make build` |
| Run (starts containers automatically) | `./bin/helixagent` |
| Format | `make fmt` |
| Lint | `make lint` |
| All tests | `make test` |
| Single test | `go test -v -run TestName ./path/to/pkg` |
| Pre-commit | `make fmt vet lint` |

**NOTE:** Container orchestration is AUTOMATIC. Do NOT run manual container commands.

## Key Files

- `CLAUDE.md` - Detailed project architecture
- `Makefile` - All available commands
- `go.mod` - Module dependencies
- `docs/MODULES.md` - Extracted modules catalog (26 modules)
- `.env.example` - Environment variable templates
