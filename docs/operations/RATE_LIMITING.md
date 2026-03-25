# Rate Limiting Configuration

HelixAgent enforces rate limiting to protect both the service and upstream LLM providers from overload.

## Overview

Rate limiting is implemented in the HTTP middleware layer (`internal/middleware/`) and uses a combination of token bucket and sliding window algorithms from the Concurrency module (`digital.vasic.concurrency`).

## Configuration

Rate limits are set per API key and can be configured in `configs/production.yaml`:

```yaml
rate_limiting:
  enabled: true
  default_rpm: 60          # Requests per minute
  default_tpm: 100000      # Tokens per minute
  burst_multiplier: 1.5    # Allow short bursts above limit
```

Environment variable overrides:

| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMIT_ENABLED` | `true` | Enable/disable rate limiting |
| `RATE_LIMIT_RPM` | `60` | Default requests per minute |
| `RATE_LIMIT_TPM` | `100000` | Default tokens per minute |

## Response Headers

When rate limiting is active, responses include:

| Header | Description |
|--------|-------------|
| `X-RateLimit-Limit` | Maximum requests allowed |
| `X-RateLimit-Remaining` | Requests remaining in window |
| `X-RateLimit-Reset` | Unix timestamp when limit resets |
| `Retry-After` | Seconds to wait (only on 429 responses) |

## Provider-Level Limits

Each LLM provider has its own upstream rate limits. HelixAgent detects these via rate limit response headers and adjusts request pacing automatically. The subscription detection system (`internal/verifier/subscription_detector.go`) identifies provider tier limits.

## Circuit Breaker Integration

When a provider's rate limit is exceeded repeatedly, the circuit breaker trips to prevent further requests until the provider recovers. See `internal/architecture/CIRCUIT_BREAKER.md` for details.

## Related Documentation

- [Circuit Breaker](../architecture/CIRCUIT_BREAKER.md)
- [Monitoring](../monitoring/README.md)
- [Configuration Guide](../guides/configuration-guide.md)
