# AGENTS.md

## MANDATORY: No CI/CD Pipelines

**NO GitHub Actions, GitLab CI/CD, or any automated pipeline may exist in this repository!**

- No `.github/workflows/` directory
- No `.gitlab-ci.yml` file
- No Jenkinsfile, .travis.yml, .circleci, or any other CI configuration
- **NO Git hooks (pre-commit, pre-push, post-commit, etc.)** may be installed or configured
- All builds and tests are run manually or via Makefile targets
- This rule is permanent and non-negotiable

Guidance for AI coding agents working in the HelixAgent repository.

## Project Overview

HelixAgent is an AI-powered ensemble LLM service in Go (1.24+) that aggregates responses from multiple language models. It provides OpenAI-compatible APIs with 22+ LLM providers, debate orchestration, MCP adapters, and containerized infrastructure.

**Module**: `dev.helix.agent`

**Subprojects**: Toolkit (`Toolkit/`) — Go library for AI apps. LLMsVerifier (`LLMsVerifier/`) — provider accuracy verification. 27 extracted modules.

## Mandatory Development Standards (NON-NEGOTIABLE)

1. **100% Test Coverage** — Unit, integration, E2E, security, stress, chaos, automation, benchmark tests. Mocks ONLY in unit tests.
2. **Challenge Coverage** — Every component MUST have Challenge scripts (`./challenges/scripts/`).
3. **Containerization** — All services in containers. Auto boot-up via HelixAgent binary. ALL container ops via `internal/adapters/containers/adapter.go`. No direct `docker`/`podman` commands.
4. **Configuration via HelixAgent Only** — CLI agent configs generated ONLY by `./bin/helixagent --generate-agent-config=<name>`.
5. **Real Data** — Beyond unit tests, use actual API calls, real databases, live services.
6. **No Mocks in Production** — Mocks, stubs, TODO implementations FORBIDDEN in production.
7. **Resource Limits** — ALL tests limited to 30-40% host resources: `GOMAXPROCS=2 nice -n 19 ionice -c 3`.

## Build Commands

```bash
make build              # Build binary (output in bin/)
make build-debug        # Build with debug symbols
make run                # Run locally
make run-dev            # Development mode (GIN_MODE=debug)
make docker-build       # Build Docker image
make docker-run         # Start services with Docker Compose
make release            # Build helixagent for all platforms (container-based)
make release-all        # Build ALL 7 apps for all platforms
make release-<app>      # Build specific app
```

## Linting & Formatting

```bash
make fmt                # Format with go fmt
make vet                # Run go vet
make lint               # Run golangci-lint
make security-scan      # Run gosec
make ci-pre-commit      # Pre-commit (fmt, vet)
make ci-pre-push        # Pre-push (includes unit tests)
```

**Always run `make fmt vet lint` before committing.**

## Testing Commands

### All Tests
```bash
make test               # All tests (verbose)
make test-unit          # Unit tests only (./internal/... -short)
make test-integration   # Integration tests
make test-e2e           # End-to-end tests
make test-security      # Security tests
make test-stress        # Stress tests
make test-chaos         # Challenge tests
make test-bench         # Benchmark tests
make test-fuzz          # Fuzz tests (corpus replay)
make test-race          # Race detection
make test-coverage      # Coverage with HTML report
```

### Single Test
```bash
go test -v -run TestFunctionName ./path/to/package
go test -v ./internal/llm
go test -v -run "Test.*Integration" ./...

# Resource-limited (CRITICAL)
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -v -p 1 -run TestName ./path/to/package
```

### Infrastructure
```bash
make test-infra-start   # Start PostgreSQL, Redis, Mock LLM containers
make test-infra-stop    # Stop test infrastructure
make test-with-infra    # Run tests with Docker infra
```

**IMPORTANT:** Containers MUST be running before tests/challenges.

## Code Style Guidelines

- **Formatting**: Use `gofmt` / `goimports`. Imports grouped: stdlib, third-party, internal (blank line separated). Line length ≤ 100.
- **Naming**: `camelCase` private, `PascalCase` exported, `UPPER_SNAKE_CASE` constants. Acronyms all caps (`HTTP`, `URL`, `ID`, `JSON`). Receiver names: 1-2 letters (`s` for service, `c` for client).
- **Error Handling**: Always check errors, wrap with `fmt.Errorf("...: %w", err)`. Use `defer` for cleanup.
- **Types & Interfaces**: Use `interface` to define behavior, not data. Prefer small, focused interfaces. Avoid `any`/`interface{}`; use generics.
- **Concurrency**: Always use `context.Context`. Protect shared data with `sync.Mutex`/`sync.RWMutex`. Use `sync.WaitGroup` for goroutine coordination.
- **Testing**: Write table-driven tests. Use `testify` assertion library. Mocks/stubs ONLY in unit tests.

## Key Conventions

- **Tool Schema**: All tool parameters use **snake_case**. See `internal/tools/schema.go`.
- **No Comments**: **DO NOT ADD COMMENTS** unless explicitly requested.
- **Git Operations**: **SSH ONLY** — HTTPS forbidden. Branch naming: `feat/`, `fix/`, `chore/`, `docs/`, `refactor/`, `test/` + description. Commits: Conventional Commits (`feat(scope): description`). Run `make fmt vet lint` before committing.
- **Containerization**: **ALL container orchestration handled AUTOMATICALLY by HelixAgent binary.** Forbidden: manual `docker`/`podman` commands, `make test-infra-start`. Acceptable workflow: `make build` → `./bin/helixagent` (reads `Containers/.env`, orchestrates everything).

## Quick Reference

| Task | Command |
|------|---------|
| Build | `make build` |
| Run | `./bin/helixagent` |
| Format | `make fmt` |
| Lint | `make lint` |
| All tests | `make test` |
| Single test | `go test -v -run TestName ./path/to/pkg` |
| Pre-commit | `make fmt vet lint` |
| Release | `make release` |

## Adding a New LLM Provider

1. Create `internal/llm/providers/<name>/<name>.go` implementing `LLMProvider`
2. Add tool support if applicable (`SupportsTools: true`)
3. Register in `internal/services/provider_registry.go`
4. Add env vars to `.env.example`, tests in `internal/llm/providers/<name>/<name>_test.go`