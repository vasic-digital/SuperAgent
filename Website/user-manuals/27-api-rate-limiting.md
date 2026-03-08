# User Manual 27: API Rate Limiting

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Rate Limiting Architecture](#rate-limiting-architecture)
4. [Token Bucket Algorithm](#token-bucket-algorithm)
5. [Sliding Window Algorithm](#sliding-window-algorithm)
6. [Default Rate Limits](#default-rate-limits)
7. [Per-Provider Rate Limits](#per-provider-rate-limits)
8. [Middleware Configuration](#middleware-configuration)
9. [Response Headers](#response-headers)
10. [Redis-Backed Distributed Limiting](#redis-backed-distributed-limiting)
11. [Custom Rate Limit Rules](#custom-rate-limit-rules)
12. [Rate Limit Bypass and Whitelisting](#rate-limit-bypass-and-whitelisting)
13. [Monitoring Rate Limits](#monitoring-rate-limits)
14. [Configuration Reference](#configuration-reference)
15. [Troubleshooting](#troubleshooting)
16. [Related Resources](#related-resources)

## Overview

HelixAgent implements multi-layer rate limiting to protect both the service itself and upstream LLM providers from overload. Rate limits operate at three levels: per-client (based on API key or IP), per-endpoint (different limits for chat completions vs. model listing), and per-provider (respecting each LLM provider's rate limits). Two algorithms are available: token bucket (for burst tolerance) and sliding window (for strict per-interval enforcement).

The Concurrency module (`digital.vasic.concurrency`) provides the rate limiter implementations used by the HelixAgent middleware.

## Prerequisites

- HelixAgent running on port 7061
- Redis for distributed rate limiting (optional; in-memory works for single instances)
- Understanding of the client authentication model (API keys, JWT)

## Rate Limiting Architecture

```
+-------------------+
|  Client Request   |
+--------+----------+
         |
+--------v----------+
|  IP-Based Limiter |  (anonymous requests)
|  60 req/min       |
+--------+----------+
         |
+--------v----------+
|  API Key Limiter  |  (authenticated requests)
|  600 req/min      |
+--------+----------+
         |
+--------v----------+
|  Endpoint Limiter |  (per-route limits)
|  varies by path   |
+--------+----------+
         |
+--------v----------+
|  Provider Limiter |  (upstream protection)
|  per-provider     |
+--------+----------+
         |
+--------v----------+
|  Handler Logic    |
+-------------------+
```

## Token Bucket Algorithm

The token bucket algorithm allows bursty traffic while enforcing a sustained rate. Tokens are added at a fixed rate and consumed per request. Unused tokens accumulate up to a maximum burst size.

```go
import "dev.helix.agent/Concurrency/ratelimiter"

limiter := ratelimiter.NewTokenBucket(ratelimiter.TokenBucketConfig{
    Rate:      100,  // 100 tokens per second (sustained rate)
    BurstSize: 200,  // Allow bursts up to 200 requests
})

// In the request handler
if !limiter.Allow() {
    c.JSON(http.StatusTooManyRequests, gin.H{
        "error": "rate limit exceeded",
        "retry_after": limiter.RetryAfter().Seconds(),
    })
    return
}
```

### How It Works

```
Bucket capacity: 200 tokens
Refill rate: 100 tokens/second

Time 0s:   [200/200] -- bucket full
Time 0s:   150 requests arrive, 150 tokens consumed -> [50/200]
Time 0.5s: 50 tokens refilled -> [100/200]
Time 0.5s: 120 requests arrive, only 100 allowed -> [0/200], 20 rejected
Time 1s:   100 tokens refilled -> [100/200]
```

## Sliding Window Algorithm

The sliding window algorithm provides precise rate enforcement without burst tolerance. It counts requests within a sliding time window.

```go
limiter := ratelimiter.NewSlidingWindow(ratelimiter.SlidingWindowConfig{
    Window:    time.Minute,
    MaxCount:  600,          // 600 requests per minute
    Precision: time.Second,  // 1-second resolution
})

if !limiter.Allow() {
    // Rate limited
}
```

### How It Works

```
Window: 1 minute, Max: 600

Time 10:00:00 - 10:01:00: 500 requests -> allowed (500/600)
Time 10:00:30 - 10:01:30: sliding window recalculates
  Requests in 10:00:00-10:00:30: 300 (weighted 50%)
  Requests in 10:00:30-10:01:30: 400 (weighted 100%)
  Estimated: 150 + 400 = 550 -> allowed
```

## Default Rate Limits

```yaml
rate_limits:
  # Anonymous requests (no API key, identified by IP)
  anonymous:
    requests_per_minute: 60
    requests_per_hour: 1000
    burst_size: 10
    algorithm: sliding_window

  # Authenticated requests (valid API key or JWT)
  authenticated:
    requests_per_minute: 600
    requests_per_hour: 10000
    burst_size: 100
    algorithm: token_bucket

  # Per-endpoint overrides
  endpoints:
    "/v1/chat/completions":
      requests_per_minute: 120
      burst_size: 20
    "/v1/models":
      requests_per_minute: 300
      burst_size: 50
    "/v1/embeddings":
      requests_per_minute: 200
      burst_size: 30
    "/v1/monitoring/status":
      requests_per_minute: 600   # Higher limit for health checks
      burst_size: 100
```

## Per-Provider Rate Limits

Each LLM provider has its own rate limits. HelixAgent respects these to prevent 429 responses from upstream:

| Provider | Requests/Min | Tokens/Min | Notes |
|---|---|---|---|
| OpenAI | 60-10000 | 40000-300000 | Varies by tier |
| Claude | 60-4000 | 100000 | Varies by plan |
| DeepSeek | 60 | N/A | Free tier |
| Gemini | 60-1000 | N/A | Varies by model |
| Mistral | 120 | N/A | Standard tier |
| OpenRouter | 200 | N/A | Free models lower |
| Cerebras | 30 | N/A | Fast inference |

Provider rate limits are configured in the provider registry and enforced per-provider:

```go
// Per-provider limiter in the provider registry
type ProviderConfig struct {
    Name           string
    RateLimit      int           // requests per minute
    TokenLimit     int           // tokens per minute
    BurstSize      int           // burst allowance
    CooldownPeriod time.Duration // wait after 429 response
}
```

## Middleware Configuration

### Gin Middleware Setup

```go
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
    // Create per-IP and per-API-key limiters
    ipLimiters := sync.Map{}
    keyLimiters := sync.Map{}

    return func(c *gin.Context) {
        var limiter *ratelimiter.TokenBucket
        var key string

        apiKey := c.GetHeader("Authorization")
        if apiKey != "" {
            key = apiKey
            l, _ := keyLimiters.LoadOrStore(key,
                ratelimiter.NewTokenBucket(config.Authenticated))
            limiter = l.(*ratelimiter.TokenBucket)
        } else {
            key = c.ClientIP()
            l, _ := ipLimiters.LoadOrStore(key,
                ratelimiter.NewTokenBucket(config.Anonymous))
            limiter = l.(*ratelimiter.TokenBucket)
        }

        if !limiter.Allow() {
            setRateLimitHeaders(c, limiter)
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": gin.H{
                    "message": "Rate limit exceeded. Please retry after the reset time.",
                    "type":    "rate_limit_error",
                },
            })
            return
        }

        setRateLimitHeaders(c, limiter)
        c.Next()
    }
}
```

### Registering the Middleware

```go
router := gin.New()
router.Use(RateLimitMiddleware(RateLimitConfig{
    Authenticated: ratelimiter.TokenBucketConfig{Rate: 10, BurstSize: 100},
    Anonymous:     ratelimiter.TokenBucketConfig{Rate: 1, BurstSize: 10},
}))
```

## Response Headers

All responses include standard rate limit headers:

```http
HTTP/1.1 200 OK
X-RateLimit-Limit: 600
X-RateLimit-Remaining: 542
X-RateLimit-Reset: 1741459260
X-RateLimit-Policy: token_bucket
Retry-After: 60
```

| Header | Description |
|---|---|
| `X-RateLimit-Limit` | Maximum requests allowed in the current window |
| `X-RateLimit-Remaining` | Remaining requests in the current window |
| `X-RateLimit-Reset` | Unix timestamp when the limit resets |
| `X-RateLimit-Policy` | Algorithm in use (token_bucket or sliding_window) |
| `Retry-After` | Seconds to wait before retrying (only on 429 responses) |

### 429 Response Body

```json
{
    "error": {
        "message": "Rate limit exceeded. Please retry after the reset time.",
        "type": "rate_limit_error",
        "code": "rate_limit_exceeded",
        "retry_after": 12.5
    }
}
```

## Redis-Backed Distributed Limiting

For multi-instance deployments, use Redis as the rate limit backend to share state across instances:

```yaml
rate_limiter:
  backend: redis
  redis_url: redis://:helixagent123@localhost:16379/1
  key_prefix: "rl:"
  sync_interval: 100ms
```

### Redis Implementation

```go
func (r *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
    pipe := r.client.Pipeline()

    now := time.Now().UnixMilli()
    windowStart := now - r.windowMs

    // Remove expired entries
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

    // Count current window
    pipe.ZCard(ctx, key)

    // Add current request
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})

    // Set key expiry
    pipe.Expire(ctx, key, r.window+time.Second)

    results, err := pipe.Exec(ctx)
    if err != nil {
        return false, fmt.Errorf("redis pipeline: %w", err)
    }

    count := results[1].(*redis.IntCmd).Val()
    return count < int64(r.maxCount), nil
}
```

## Custom Rate Limit Rules

### Per-User Tier Limits

```yaml
rate_limit_tiers:
  free:
    requests_per_minute: 20
    requests_per_day: 500
  pro:
    requests_per_minute: 200
    requests_per_day: 10000
  enterprise:
    requests_per_minute: 2000
    requests_per_day: 100000
```

### Per-Model Limits

```yaml
model_rate_limits:
  "helixagent-debate":
    requests_per_minute: 30   # Debates are expensive
  "helixagent-fast":
    requests_per_minute: 300  # Single-provider, fast
```

## Rate Limit Bypass and Whitelisting

### IP Whitelist

```yaml
rate_limit_whitelist:
  ips:
    - "10.0.0.0/8"       # Internal network
    - "172.16.0.0/12"    # Docker network
  api_keys:
    - "admin-key-hash"   # Admin bypass
```

### Health Check Exemption

The `/v1/monitoring/status` and `/metrics` endpoints are exempt from rate limiting to ensure monitoring always works:

```go
func isExemptPath(path string) bool {
    exemptPaths := []string{
        "/v1/monitoring/status",
        "/metrics",
        "/health",
    }
    for _, p := range exemptPaths {
        if path == p {
            return true
        }
    }
    return false
}
```

## Monitoring Rate Limits

### Prometheus Metrics

```promql
# Rate limit rejection rate
rate(helixagent_rate_limit_rejected_total[5m])

# Rate limit hit ratio by client
helixagent_rate_limit_remaining / helixagent_rate_limit_limit

# Top rate-limited clients
topk(10, sum by (client) (rate(helixagent_rate_limit_rejected_total[5m])))
```

### API Endpoint

```bash
# View current rate limit status for a client
curl -s http://localhost:7061/v1/admin/rate-limits?client=api-key-123 \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq .
```

## Configuration Reference

| Setting | Default | Description |
|---|---|---|
| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |
| `RATE_LIMIT_BACKEND` | `memory` | Backend (memory or redis) |
| `RATE_LIMIT_ANON_RPM` | `60` | Anonymous requests per minute |
| `RATE_LIMIT_AUTH_RPM` | `600` | Authenticated requests per minute |
| `RATE_LIMIT_BURST` | `100` | Token bucket burst size |
| `RATE_LIMIT_ALGORITHM` | `token_bucket` | Algorithm (token_bucket or sliding_window) |
| `RATE_LIMIT_REDIS_URL` | `""` | Redis URL for distributed limiting |
| `RATE_LIMIT_KEY_PREFIX` | `rl:` | Redis key prefix |

## Troubleshooting

### Clients Receiving 429 Too Frequently

**Symptom:** Legitimate clients are being rate limited.

**Solutions:**
1. Check current rate limit settings: `curl -v http://localhost:7061/v1/models` (inspect headers)
2. Increase per-client limits for authenticated users
3. Switch from sliding window to token bucket to allow bursts
4. Add the client's IP or API key to the whitelist if appropriate
5. Verify the client is sending the `Authorization` header (otherwise treated as anonymous)

### Rate Limits Not Working in Multi-Instance Setup

**Symptom:** Each instance enforces its own limits independently.

**Solutions:**
1. Enable Redis backend: `RATE_LIMIT_BACKEND=redis`
2. Verify all instances connect to the same Redis instance
3. Check Redis connectivity: `redis-cli -h localhost -p 16379 -a helixagent123 PING`

### Memory Usage Growing from Rate Limiter

**Symptom:** In-memory rate limiter consumes increasing memory.

**Solutions:**
1. Enable cleanup of expired limiter entries (default: every 5 minutes)
2. Switch to Redis backend for long-running deployments
3. Reduce the number of unique keys tracked (aggregate by subnet instead of IP)

### Provider 429 Errors Despite Client Rate Limiting

**Symptom:** LLM providers return 429 even though client-facing limits are not exceeded.

**Solutions:**
1. Configure per-provider rate limits to match provider quotas
2. Enable the circuit breaker to back off after 429 responses
3. Check the fallback chain to route to alternative providers
4. Review provider subscription tier and upgrade if needed

## Related Resources

- [User Manual 19: Concurrency Patterns](19-concurrency-patterns.md) -- Rate limiter implementations
- [User Manual 28: Custom Middleware](28-custom-middleware.md) -- Middleware integration
- [User Manual 18: Performance Monitoring](18-performance-monitoring.md) -- Monitoring rate limit metrics
- [User Manual 26: Compliance Guide](26-compliance-guide.md) -- Rate limiting for abuse prevention
- Concurrency module: `Concurrency/`
- Middleware: `internal/middleware/`
- Rate limiter: `Concurrency/ratelimiter/`
