# Security Best Practices in HelixAgent

## Authentication & Authorization

- JWT tokens with configurable expiry
- API key validation with rate limiting
- OAuth2 integration for CLI providers (Claude, Qwen, Gemini)

## Input Validation

- All API inputs validated at handler level
- Path traversal prevention in `internal/utils/path_validation.go`
- SQL parameterized queries via pgx (no string concatenation)

## Secure Communication

- HTTP/3 (QUIC) as primary transport with TLS
- Brotli compression with gzip fallback
- No plain HTTP in production

## Dependency Management

- Vendored dependencies (`vendor/`) for reproducible builds
- Snyk monitoring for vulnerability alerts
- Minimal dependency surface

## Container Security

- Non-root container execution
- Read-only filesystem where possible
- Resource limits on all containers
- Health checks for all services

## Secret Management

- API keys via environment variables or `.env` files
- No secrets in source code (gosec G101 enforced)
- Credential rotation support
