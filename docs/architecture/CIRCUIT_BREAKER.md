# Circuit Breaker Pattern in HelixAgent

## Overview

HelixAgent implements the Circuit Breaker pattern to provide fault tolerance and graceful degradation when communicating with LLM providers. This pattern prevents cascading failures, reduces load on failing services, and enables automatic recovery.

## Pattern Explanation

### What is a Circuit Breaker?

The Circuit Breaker pattern is a design pattern used in distributed systems to detect failures and encapsulate the logic of preventing a failure from constantly recurring. It works similarly to an electrical circuit breaker:

- **Closed State**: Normal operation - requests pass through
- **Open State**: Failure detected - requests are immediately rejected
- **Half-Open State**: Recovery testing - limited requests allowed through

### Why Use Circuit Breakers?

1. **Prevent Cascading Failures**: When an LLM provider fails, the circuit breaker prevents the failure from propagating to other parts of the system
2. **Reduce Latency**: Failing fast rather than waiting for timeouts improves response times
3. **Enable Recovery**: Automatic recovery testing allows services to resume when the provider becomes healthy
4. **Resource Protection**: Prevents wasting resources on requests that are likely to fail

## Implementation

### Core Structure

```go
// CircuitBreaker implements the circuit breaker pattern for LLM providers
type CircuitBreaker struct {
    name          string
    state         CircuitState
    failureCount  int64
    successCount  int64
    lastFailure   time.Time
    lastSuccess   time.Time

    // Configuration
    failureThreshold    int64
    successThreshold    int64
    timeout             time.Duration
    halfOpenMaxRequests int64

    // Metrics
    metrics    *CircuitMetrics

    mu sync.RWMutex
}

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
    StateClosed CircuitState = iota
    StateOpen
    StateHalfOpen
)
```

### State Transitions

```
                    +-------------+
                    |             |
         Success    |   Closed    |  Failure < Threshold
         +--------->|   (Normal)  |<---------+
         |          |             |          |
         |          +------+------+          |
         |                 |                 |
         |                 | Failures >= Threshold
         |                 v
         |          +------+------+
         |          |             |
         |          |    Open     |
         |          |  (Failing)  |
         |          |             |
         |          +------+------+
         |                 |
         |                 | Timeout Expired
         |                 v
         |          +------+------+
         |          |             |
         +----------| Half-Open   |
           Success  | (Testing)   |
                    |             |
                    +------+------+
                           |
                           | Failure
                           v
                    (Back to Open)
```

## Configuration Options

### Circuit Breaker Configuration

```yaml
circuit_breaker:
  enabled: true

  # Global defaults
  defaults:
    failure_threshold: 5       # Failures before opening circuit
    success_threshold: 3       # Successes in half-open before closing
    timeout: 30s               # Time before attempting recovery
    half_open_max_requests: 3  # Max requests allowed in half-open state

  # Per-provider overrides
  providers:
    deepseek:
      failure_threshold: 3
      timeout: 60s

    openrouter:
      failure_threshold: 5
      timeout: 45s

    claude:
      failure_threshold: 2
      success_threshold: 5
      timeout: 30s
```

### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `failure_threshold` | int | 5 | Number of consecutive failures before opening |
| `success_threshold` | int | 3 | Successful requests in half-open before closing |
| `timeout` | duration | 30s | Time before transitioning from open to half-open |
| `half_open_max_requests` | int | 3 | Maximum concurrent requests in half-open state |
| `failure_rate_threshold` | float | 0.5 | Failure rate threshold (alternative to count) |
| `sampling_duration` | duration | 60s | Time window for calculating failure rate |

### Advanced Configuration

```yaml
circuit_breaker:
  enabled: true

  # Metrics collection
  metrics:
    enabled: true
    export_interval: 10s
    histogram_buckets: [0.001, 0.01, 0.1, 0.5, 1.0, 5.0]

  # Health check integration
  health_check:
    enabled: true
    interval: 30s
    timeout: 5s

  # Sliding window configuration
  sliding_window:
    type: "count"        # "count" or "time"
    size: 100            # Number of calls or seconds
    minimum_calls: 10    # Minimum calls before calculating failure rate

  # Slow call handling
  slow_call:
    threshold: 5s        # Duration threshold for slow calls
    rate_threshold: 0.8  # Slow call rate threshold
```

## State Transitions

### Closed to Open

The circuit opens when failure conditions are met:

```go
func (cb *CircuitBreaker) recordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.failureCount++
    cb.lastFailure = time.Now()
    cb.metrics.RecordFailure()

    // Check if we should open the circuit
    if cb.failureCount >= cb.failureThreshold {
        cb.state = StateOpen
        cb.openedAt = time.Now()
        cb.metrics.RecordStateChange(StateClosed, StateOpen)
        log.WithFields(logrus.Fields{
            "provider":      cb.name,
            "failures":      cb.failureCount,
            "threshold":     cb.failureThreshold,
        }).Warn("Circuit breaker opened")
    }
}
```

### Open to Half-Open

After the timeout expires, the circuit transitions to half-open:

```go
func (cb *CircuitBreaker) shouldAllowRequest() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()

    switch cb.state {
    case StateClosed:
        return true

    case StateOpen:
        // Check if timeout has expired
        if time.Since(cb.openedAt) >= cb.timeout {
            // Transition to half-open
            cb.mu.RUnlock()
            cb.mu.Lock()
            if cb.state == StateOpen {
                cb.state = StateHalfOpen
                cb.halfOpenRequests = 0
                cb.metrics.RecordStateChange(StateOpen, StateHalfOpen)
            }
            cb.mu.Unlock()
            cb.mu.RLock()
            return true
        }
        return false

    case StateHalfOpen:
        return cb.halfOpenRequests < cb.halfOpenMaxRequests
    }

    return false
}
```

### Half-Open to Closed or Open

Based on the results of test requests:

```go
func (cb *CircuitBreaker) recordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.successCount++
    cb.lastSuccess = time.Now()
    cb.metrics.RecordSuccess()

    if cb.state == StateHalfOpen {
        cb.halfOpenSuccesses++

        if cb.halfOpenSuccesses >= cb.successThreshold {
            // Recovery confirmed - close the circuit
            cb.state = StateClosed
            cb.failureCount = 0
            cb.successCount = 0
            cb.metrics.RecordStateChange(StateHalfOpen, StateClosed)
            log.WithField("provider", cb.name).Info("Circuit breaker closed - provider recovered")
        }
    }
}

func (cb *CircuitBreaker) recordHalfOpenFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    // Half-open failure - reopen the circuit
    cb.state = StateOpen
    cb.openedAt = time.Now()
    cb.halfOpenSuccesses = 0
    cb.metrics.RecordStateChange(StateHalfOpen, StateOpen)
    log.WithField("provider", cb.name).Warn("Circuit breaker reopened - recovery failed")
}
```

## Monitoring and Metrics

### Prometheus Metrics

HelixAgent exports comprehensive circuit breaker metrics:

```prometheus
# Circuit breaker state (0=closed, 1=open, 2=half-open)
helixagent_circuit_breaker_state{provider="deepseek"} 0

# Total state transitions
helixagent_circuit_breaker_state_transitions_total{provider="deepseek",from="closed",to="open"} 5

# Request outcomes
helixagent_circuit_breaker_requests_total{provider="deepseek",result="success"} 1000
helixagent_circuit_breaker_requests_total{provider="deepseek",result="failure"} 50
helixagent_circuit_breaker_requests_total{provider="deepseek",result="rejected"} 200

# Current failure count
helixagent_circuit_breaker_failures{provider="deepseek"} 3

# Time since last state change
helixagent_circuit_breaker_time_in_state_seconds{provider="deepseek",state="closed"} 3600
```

### Metric Types

| Metric | Type | Description |
|--------|------|-------------|
| `circuit_breaker_state` | Gauge | Current state (0=closed, 1=open, 2=half-open) |
| `circuit_breaker_requests_total` | Counter | Total requests by result |
| `circuit_breaker_state_transitions_total` | Counter | State transition count |
| `circuit_breaker_failures` | Gauge | Current failure count |
| `circuit_breaker_latency_seconds` | Histogram | Request latency distribution |
| `circuit_breaker_time_in_state_seconds` | Gauge | Time spent in current state |

### Grafana Dashboard

Example Grafana queries for circuit breaker monitoring:

```promql
# Circuit breaker state timeline
helixagent_circuit_breaker_state{provider=~"$provider"}

# Requests rejected by open circuit
rate(helixagent_circuit_breaker_requests_total{result="rejected"}[5m])

# Recovery time (time from open to closed)
helixagent_circuit_breaker_time_in_state_seconds{state="open"}

# Failure rate by provider
rate(helixagent_circuit_breaker_requests_total{result="failure"}[5m])
  / rate(helixagent_circuit_breaker_requests_total[5m])
```

### Health Check Endpoint

The health check endpoint includes circuit breaker status:

```json
GET /v1/health

{
  "status": "healthy",
  "providers": {
    "deepseek": {
      "status": "healthy",
      "circuit_breaker": {
        "state": "closed",
        "failure_count": 0,
        "success_count": 1000,
        "last_failure": null,
        "last_success": "2024-01-15T10:30:00Z"
      }
    },
    "openrouter": {
      "status": "degraded",
      "circuit_breaker": {
        "state": "half-open",
        "failure_count": 5,
        "success_count": 2,
        "last_failure": "2024-01-15T10:25:00Z",
        "last_success": "2024-01-15T10:29:00Z"
      }
    }
  }
}
```

## Integration Guide

### Using the Circuit Breaker

```go
// Create a new circuit breaker
config := &CircuitBreakerConfig{
    FailureThreshold:    5,
    SuccessThreshold:    3,
    Timeout:             30 * time.Second,
    HalfOpenMaxRequests: 3,
}

cb := NewCircuitBreaker("deepseek", config)

// Execute request with circuit breaker protection
func executeWithCircuitBreaker(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
    // Check if request is allowed
    if !cb.AllowRequest() {
        return nil, ErrCircuitOpen
    }

    // Execute the request
    resp, err := provider.Complete(ctx, req)

    // Record the result
    if err != nil {
        cb.RecordFailure()
        return nil, err
    }

    cb.RecordSuccess()
    return resp, nil
}
```

### Ensemble Integration

The ensemble system integrates circuit breakers for automatic failover:

```go
func (e *Ensemble) Complete(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
    // Get available providers (with open circuits filtered out)
    providers := e.getAvailableProviders()

    if len(providers) == 0 {
        return nil, ErrNoAvailableProviders
    }

    // Execute with fallback
    for _, provider := range providers {
        if !provider.CircuitBreaker.AllowRequest() {
            continue
        }

        resp, err := provider.Complete(ctx, req)
        if err != nil {
            provider.CircuitBreaker.RecordFailure()
            continue
        }

        provider.CircuitBreaker.RecordSuccess()
        return resp, nil
    }

    return nil, ErrAllProvidersFailed
}
```

### Custom Failure Detection

Implement custom failure detection for specific error types:

```go
type FailureDetector interface {
    IsFailure(err error, resp *LLMResponse) bool
}

type DefaultFailureDetector struct{}

func (d *DefaultFailureDetector) IsFailure(err error, resp *LLMResponse) bool {
    if err != nil {
        // Network errors and timeouts are failures
        if errors.Is(err, context.DeadlineExceeded) ||
           errors.Is(err, context.Canceled) ||
           isNetworkError(err) {
            return true
        }

        // Rate limiting is not a failure (temporary)
        if isRateLimitError(err) {
            return false
        }

        return true
    }

    // Check response for errors
    if resp != nil && resp.Error != nil {
        return true
    }

    return false
}
```

### Event Listeners

Subscribe to circuit breaker events:

```go
cb.OnStateChange(func(from, to CircuitState) {
    log.WithFields(logrus.Fields{
        "provider": cb.Name(),
        "from":     from.String(),
        "to":       to.String(),
    }).Info("Circuit breaker state changed")

    // Send alert if circuit opens
    if to == StateOpen {
        alerting.SendAlert(Alert{
            Severity: "warning",
            Title:    fmt.Sprintf("Circuit breaker opened for %s", cb.Name()),
            Message:  "Provider is experiencing failures",
        })
    }
})
```

## Best Practices

### 1. Tune Thresholds Appropriately

- Start with conservative thresholds and adjust based on monitoring
- Consider provider-specific characteristics (latency, reliability)
- Use failure rates rather than counts for high-traffic scenarios

### 2. Implement Graceful Degradation

```go
func handleCircuitOpen(provider string) (*Response, error) {
    // Try fallback provider
    if fallback := getFallbackProvider(provider); fallback != nil {
        return fallback.Complete(ctx, req)
    }

    // Return cached response if available
    if cached := cache.Get(req.Hash()); cached != nil {
        return cached, nil
    }

    // Return graceful error
    return nil, ErrServiceUnavailable
}
```

### 3. Monitor and Alert

- Set up alerts for circuit state changes
- Track time in open state
- Monitor recovery success rate

### 4. Test Circuit Breaker Behavior

```go
func TestCircuitBreakerOpens(t *testing.T) {
    cb := NewCircuitBreaker("test", &CircuitBreakerConfig{
        FailureThreshold: 3,
        Timeout:          1 * time.Second,
    })

    // Record failures
    for i := 0; i < 3; i++ {
        cb.RecordFailure()
    }

    assert.Equal(t, StateOpen, cb.State())
    assert.False(t, cb.AllowRequest())
}
```

## Troubleshooting

### Circuit Opens Too Frequently

1. Increase `failure_threshold`
2. Implement better error classification
3. Add retry logic before recording failure

### Recovery Takes Too Long

1. Decrease `timeout` duration
2. Increase `half_open_max_requests`
3. Implement active health probing

### False Positives

1. Improve failure detection logic
2. Exclude transient errors (rate limiting)
3. Use sliding window for failure rate calculation

---

For more information, see the [HelixAgent documentation](https://dev.helix.agent).
