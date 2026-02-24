# AGENTS.md - Toolkit Module

## Overview

Toolkit is a Go library for building AI-powered applications with multi-provider support,
specialized agents, and reusable infrastructure utilities. It provides unified interfaces for
chat completion, embeddings, reranking, and model discovery across AI providers, plus shared
packages for authentication, configuration, HTTP communication, rate limiting, error handling,
response parsing, and testing.

## Key Files

### Core (`pkg/toolkit/`)

- `toolkit.go` -- `Toolkit` struct: provider and agent registry with register/get/list operations
- `interfaces.go` -- Core interfaces (`Provider`, `Agent`), request/response types (`ChatRequest`, `ChatResponse`, `EmbeddingRequest`, `EmbeddingResponse`, `RerankRequest`, `RerankResponse`), model types (`ModelInfo`, `ModelCapabilities`, `ModelCategory`), `ProviderFactoryRegistry`

### Agents (`pkg/toolkit/agents/`)

- `generic.go` -- `GenericAgent`: general-purpose AI assistant, delegates to provider `Chat()`, configurable model/temperature/max_tokens
- `codereview.go` -- `CodeReviewAgent`: specialized code review with structured analysis (security, performance, maintainability, bugs, best practices)

### Common Utilities (`pkg/toolkit/common/`)

- `discovery/discovery.go` -- `DiscoveryService`: model discovery with caching (5-min TTL), filtering by criteria, sorting; `BaseDiscovery`, `DefaultCategoryInferrer`
- `http/client.go` -- `Client`: HTTP client with retry logic (exponential backoff), auth headers, `Get`/`Post`/`Put`/`Delete`/`DoRequest` methods
- `ratelimit/ratelimit.go` -- `TokenBucket`, `SlidingWindowLimiter`, `PerKeyLimiter`, `CircuitBreaker`, `RateLimiter` interface, HTTP `Middleware`

### Commons Libraries (`Commons/`)

- `auth/auth.go` -- `AuthManager` (API key + OAuth2 token refresh), `OAuth2Refresher` (client credentials/refresh token grants), `AuthInterceptor`, auth `Middleware` (HTTP transport wrapper)
- `config/config.go` -- `Config` map type with typed getters, `Validator` with rules (`Required`, `OneOf`, `MinLength`), `ProviderConfig`, env var loading
- `discovery/discovery.go` -- Generic model discovery framework: `CapabilityInferrer`, `CategoryInferrer`, `ModelFormatter` interfaces; `BaseDiscovery`, `DefaultCapabilityInferrer`, `DefaultCategoryInferrer`, `DefaultModelFormatter`
- `errors/errors.go` -- Error types (`ProviderError`, `APIError`, `RateLimitError`, `AuthenticationError`, `NetworkError`, `TimeoutError`, `ValidationError`), `ErrorHandler`, retryability checks
- `http/client.go` -- Advanced `Client` with rate limiting (`TokenBucket`), request/response interceptors, retry with exponential backoff
- `ratelimit/ratelimit.go` -- Simple `TokenBucket` rate limiter used by `Commons/http`
- `response/response.go` -- `JSONParser`, `StreamingParser` (SSE), `ErrorDetector`, `ResponseValidator`, `PaginationParser`, `ChunkedParser`, `ResponseBuilder`
- `testing/testing.go` -- `MockHTTPClient`, `MockProvider`, `TestServer`, `TestFixtures` (sample requests/responses), assertion helpers (`AssertChatResponse`, `AssertEmbeddingResponse`, `AssertRerankResponse`)

### Providers (`Providers/`)

- `Chutes/chutes.go` -- Chutes provider implementing `Provider` interface; auto-registers via `init()`
- `Chutes/builder.go` -- `ConfigBuilder` with `Build`/`Validate`/`Merge`; `Config` struct (api_key, base_url, timeout, retries, rate_limit)
- `Chutes/client.go` -- Chutes API client: `ChatCompletion`, `CreateEmbeddings`, `CreateRerank`, `GetModels`
- `Chutes/discovery.go` -- Model discovery for Chutes
- `SiliconFlow/siliconflow.go` -- SiliconFlow provider implementing `Provider` interface; auto-registers via `init()`
- `SiliconFlow/builder.go` -- `ConfigBuilder` with `Build`/`Validate`/`Merge`; `Config` struct
- `SiliconFlow/client.go` -- SiliconFlow API client: `ChatCompletion`, `CreateEmbeddings`, `CreateRerank`
- `SiliconFlow/discovery.go` -- Model discovery for SiliconFlow

### CLI (`cmd/toolkit/`)

- `main.go` -- CLI entry point with cobra: `version`, `test`, `chat`, `agent` commands; imports providers via blank imports for auto-registration

### Tests (`tests/`)

- `integration/framework.go` -- Integration test framework
- `integration/provider_integration_test.go` -- Provider integration tests
- `e2e/e2e_test.go` -- End-to-end tests
- `performance/benchmark_test.go` -- Benchmark tests
- `security/security_test.go` -- Security tests
- `chaos/chaos_test.go` -- Chaos engineering tests

## Exported Types Summary

### interfaces.go

- `Provider` -- Interface: `Name()`, `Chat()`, `Embed()`, `Rerank()`, `DiscoverModels()`, `ValidateConfig()`
- `Agent` -- Interface: `Name()`, `Execute()`, `ValidateConfig()`, `Capabilities()`
- `ChatRequest`, `ChatResponse` -- Chat completion request/response
- `EmbeddingRequest`, `EmbeddingResponse` -- Embedding request/response
- `RerankRequest`, `RerankResponse`, `RerankResult` -- Reranking types
- `Message`, `ChatMessage`, `Choice`, `ChatChoice`, `Usage`, `Embedding`, `EmbeddingData`, `RerankData` -- Supporting types
- `ModelInfo`, `ModelCapabilities`, `ModelCategory` -- Model metadata
- `ProviderFactory`, `ProviderFactoryRegistry` -- Factory pattern
- `RegisterProviderFactory()`, `CreateProvider()`, `ListProviders()` -- Global registry functions

### toolkit.go

- `Toolkit` -- Main struct: `RegisterProvider()`, `GetProvider()`, `RegisterAgent()`, `GetAgent()`, `ListProviders()`, `ListAgents()`

### agents/generic.go

- `GenericAgent` -- `NewGenericAgent(name, description, provider)`, `Execute()`, `ValidateConfig()`, `Capabilities()`, `SetConfig()`, `GetConfig()`

### agents/codereview.go

- `CodeReviewAgent` -- `NewCodeReviewAgent(name, provider)`, `Execute()`, `ValidateConfig()`, `Capabilities()`, `SetConfig()`, `GetConfig()`

## Integration with HelixAgent

The Toolkit module is referenced by the main HelixAgent project as a submodule. It provides
the foundational `Provider` and `Agent` interfaces and common utilities used across the
HelixAgent ecosystem for provider communication, configuration management, authentication,
rate limiting, and testing infrastructure.

## Development Standards

- All code must compile and pass `go vet ./...`
- Tests must use table-driven style with `testify`
- No mocks outside unit tests
- Run `make fmt vet lint` before committing
- Follow Conventional Commits: `feat(toolkit): ...`, `fix(toolkit): ...`
