# HTTP Package

The http package provides a high-performance HTTP client pool with connection reuse, retry logic, and metrics for HelixAgent's external API calls.

## Overview

This package implements an intelligent HTTP client pool that manages connections efficiently across multiple hosts, providing automatic retries with exponential backoff, circuit breaker patterns, and comprehensive metrics collection.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     HTTPClientPool                           │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Host Clients                           ││
│  │  ┌───────────┐  ┌───────────┐  ┌───────────────────┐   ││
│  │  │api.claude │  │api.deepsk │  │ api.openrouter    │   ││
│  │  │  .ai      │  │  eek.com  │  │     .com          │   ││
│  │  └─────┬─────┘  └─────┬─────┘  └─────────┬─────────┘   ││
│  │        │              │                   │             ││
│  │  ┌─────▼──────────────▼───────────────────▼─────────┐  ││
│  │  │              Connection Pool (per host)           │  ││
│  │  │  MaxIdleConns | IdleTimeout | KeepAlive          │  ││
│  │  └──────────────────────────────────────────────────┘  ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    Retry Logic                           ││
│  │  Exponential Backoff | Jitter | Max Retries             ││
│  │  Retryable: 429, 502, 503, 504                          ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                      Metrics                             ││
│  │  Requests | Latency | Errors | Retries | Pool Stats     ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Key Types

### HTTPClientPool

The main connection pool manager.

```go
type HTTPClientPool struct {
    clients    map[string]*HostClient
    mu         sync.RWMutex
    config     PoolConfig
    metrics    *PoolMetrics
    transport  *http.Transport
}
```

### PoolConfig

Configuration for the HTTP client pool.

```go
type PoolConfig struct {
    // Connection settings
    MaxIdleConns        int           // Max idle connections total
    MaxIdleConnsPerHost int           // Max idle connections per host
    MaxConnsPerHost     int           // Max total connections per host
    IdleConnTimeout     time.Duration // Idle connection timeout

    // Timeouts
    DialTimeout         time.Duration // Connection dial timeout
    TLSHandshakeTimeout time.Duration // TLS handshake timeout
    ResponseHeaderTimeout time.Duration // Response header timeout
    RequestTimeout      time.Duration // Total request timeout

    // Retry settings
    MaxRetries          int           // Maximum retry attempts
    InitialRetryDelay   time.Duration // Initial backoff delay
    MaxRetryDelay       time.Duration // Maximum backoff delay
    RetryMultiplier     float64       // Backoff multiplier
    RetryJitter         float64       // Jitter factor (0-1)

    // TLS
    TLSConfig           *tls.Config   // Custom TLS configuration
    InsecureSkipVerify  bool          // Skip TLS verification (dev only)

    // Metrics
    EnableMetrics       bool          // Enable metrics collection
}
```

### HostClient

Per-host client wrapper with default headers.

```go
type HostClient struct {
    client         *http.Client
    defaultHeaders map[string]string
    rateLimiter    *rate.Limiter
    circuitBreaker *CircuitBreaker
    metrics        *HostMetrics
}
```

### PoolMetrics

Tracks pool-wide statistics.

```go
type PoolMetrics struct {
    TotalRequests     int64
    SuccessfulRequests int64
    FailedRequests    int64
    RetriedRequests   int64
    TotalLatency      time.Duration
    AverageLatency    time.Duration
    ActiveConnections int64
    IdleConnections   int64
}
```

## Default Configuration

```go
func DefaultPoolConfig() PoolConfig {
    return PoolConfig{
        MaxIdleConns:          100,
        MaxIdleConnsPerHost:   10,
        MaxConnsPerHost:       100,
        IdleConnTimeout:       90 * time.Second,
        DialTimeout:           30 * time.Second,
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 60 * time.Second,
        RequestTimeout:        120 * time.Second,
        MaxRetries:            3,
        InitialRetryDelay:     1 * time.Second,
        MaxRetryDelay:         30 * time.Second,
        RetryMultiplier:       2.0,
        RetryJitter:           0.1,
        EnableMetrics:         true,
    }
}
```

## Usage Examples

### Basic Usage

```go
import "dev.helix.agent/internal/http"

// Create pool with default config
pool := http.NewClientPool(http.DefaultPoolConfig())

// Make a request
resp, err := pool.Do(ctx, &http.Request{
    Method: "POST",
    URL:    "https://api.anthropic.com/v1/messages",
    Headers: map[string]string{
        "Authorization": "Bearer " + apiKey,
        "Content-Type":  "application/json",
    },
    Body: requestBody,
})
if err != nil {
    return err
}
defer resp.Body.Close()
```

### With Custom Configuration

```go
pool := http.NewClientPool(http.PoolConfig{
    MaxIdleConnsPerHost: 20,
    RequestTimeout:      60 * time.Second,
    MaxRetries:          5,
    InitialRetryDelay:   500 * time.Millisecond,
    EnableMetrics:       true,
})
```

### Per-Host Configuration

```go
// Configure host-specific settings
pool.ConfigureHost("api.anthropic.com", http.HostConfig{
    DefaultHeaders: map[string]string{
        "anthropic-version": "2024-01-01",
    },
    RateLimit:      60, // requests per minute
    CircuitBreaker: http.CircuitBreakerConfig{
        Threshold:   5,
        ResetPeriod: 30 * time.Second,
    },
})
```

### Streaming Requests

```go
// For SSE/streaming responses
resp, err := pool.DoStream(ctx, &http.Request{
    Method: "POST",
    URL:    "https://api.anthropic.com/v1/messages",
    Headers: map[string]string{
        "Accept": "text/event-stream",
    },
    Body: requestBody,
})
if err != nil {
    return err
}

// Process stream
reader := bufio.NewReader(resp.Body)
for {
    line, err := reader.ReadBytes('\n')
    if err == io.EOF {
        break
    }
    processStreamLine(line)
}
```

### With Retry Callback

```go
resp, err := pool.DoWithRetryCallback(ctx, request,
    func(attempt int, err error, delay time.Duration) {
        log.Printf("Retry %d after error: %v, waiting %v",
            attempt, err, delay)
    },
)
```

### Get Metrics

```go
metrics := pool.GetMetrics()
fmt.Printf("Total: %d, Success: %d, Failed: %d\n",
    metrics.TotalRequests,
    metrics.SuccessfulRequests,
    metrics.FailedRequests)
fmt.Printf("Average latency: %v\n", metrics.AverageLatency)

// Per-host metrics
hostMetrics := pool.GetHostMetrics("api.anthropic.com")
fmt.Printf("Anthropic requests: %d\n", hostMetrics.TotalRequests)
```

## Retry Logic

### Retryable Status Codes

The pool automatically retries requests for these status codes:

| Status Code | Description | Retry Behavior |
|-------------|-------------|----------------|
| 429 | Too Many Requests | Retry with backoff, respect Retry-After |
| 502 | Bad Gateway | Retry with backoff |
| 503 | Service Unavailable | Retry with backoff |
| 504 | Gateway Timeout | Retry with backoff |

### Exponential Backoff

```
delay = min(initialDelay * (multiplier ^ attempt) + jitter, maxDelay)
```

Example with defaults:
- Attempt 1: ~1s
- Attempt 2: ~2s
- Attempt 3: ~4s

### Retry-After Header

When the server returns a `Retry-After` header, the pool respects it:

```go
// If Retry-After: 60
// Pool waits 60 seconds before retry
```

## Circuit Breaker

Prevents cascading failures when a host is unhealthy:

```go
type CircuitBreakerConfig struct {
    Threshold   int           // Failures before opening
    ResetPeriod time.Duration // Time before half-open
    HalfOpenMax int           // Requests to allow in half-open
}

// States:
// - Closed: Normal operation
// - Open: All requests fail fast
// - Half-Open: Allow limited requests to test recovery
```

## Integration with HelixAgent

The HTTP pool is used for all external API calls:

```go
// In LLM providers
func (p *ClaudeProvider) Complete(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
    resp, err := p.httpPool.Do(ctx, &http.Request{
        Method: "POST",
        URL:    p.baseURL + "/v1/messages",
        // ...
    })
    // ...
}
```

## Testing

```bash
go test -v ./internal/http/...
go test -bench=. ./internal/http/...  # Benchmark tests
```

### Mocking

```go
func TestWithMockPool(t *testing.T) {
    pool := http.NewMockPool()
    pool.AddResponse("api.example.com", http.MockResponse{
        StatusCode: 200,
        Body:       `{"result": "success"}`,
    })

    // Test code using pool
}
```

## Performance Characteristics

| Metric | Typical Value | Notes |
|--------|---------------|-------|
| Connection Reuse | 90%+ | With keep-alive |
| Retry Overhead | < 100ms | Per attempt |
| Metrics Overhead | < 1ms | Atomic operations |
| Memory per Host | ~10KB | Connection state |

## Best Practices

1. **Reuse the pool**: Create once, use everywhere
2. **Configure per-host**: Set appropriate limits for each API
3. **Monitor metrics**: Track latency and error rates
4. **Handle rate limits**: Configure backoff appropriately
5. **Set timeouts**: Prevent hanging requests
