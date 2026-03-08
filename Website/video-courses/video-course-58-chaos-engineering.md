# Video Course 58: Chaos Engineering for LLM Services

## Course Overview

**Duration**: 2 hours
**Level**: Advanced
**Prerequisites**: Course 01-Fundamentals, Course 06-Testing, Course 57-Stress-Testing, understanding of circuit breakers and failover patterns

This course covers chaos engineering principles applied to LLM services. Topics include fault injection patterns, provider failure simulation, circuit breaker validation, recovery verification, and writing chaos tests for provider failover scenarios.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Apply chaos engineering principles to distributed LLM service architectures
2. Design and implement fault injection patterns for provider failures
3. Simulate realistic provider failure modes (timeout, error, partial response)
4. Validate circuit breaker behavior under controlled fault conditions
5. Verify system recovery after fault injection is removed
6. Write comprehensive chaos tests for provider failover chains

---

## Module 1: Chaos Engineering Principles (20 min)

### 1.1 Why Chaos Engineering for LLM Services

**Video: The Case for Controlled Failure** (8 min)

- LLM services depend on multiple external providers with varying reliability
- Provider outages, rate limits, and degraded responses are routine
- Ensemble systems add complexity: failure in one provider must not cascade
- Chaos testing validates that theoretical fault tolerance works in practice

### 1.2 Core Principles

**Video: Steady State, Hypothesis, Experiment** (7 min)

| Principle            | Application to HelixAgent                        |
|----------------------|--------------------------------------------------|
| Define steady state  | All providers healthy, response latency < 2s     |
| Form hypothesis      | Losing 1 of 3 providers degrades latency < 20%   |
| Introduce failure    | Block one provider's endpoint                     |
| Observe behavior     | Measure latency, error rate, failover speed       |
| Verify or refute     | Compare observed behavior against hypothesis      |

### 1.3 Blast Radius Control

**Video: Safe Chaos Experiments** (5 min)

- Start with the smallest possible failure scope
- Use feature flags to enable/disable fault injection
- Set experiment duration limits with automatic rollback
- Monitor key health metrics during every experiment
- Never run chaos experiments against production without explicit approval

### Hands-On Lab 1

Define a chaos experiment plan:

1. Document the steady state for HelixAgent with 3 providers
2. Form 3 hypotheses about single-provider failure impact
3. Design the fault injection mechanism for each
4. Define success/failure criteria for each experiment
5. Document the rollback procedure

---

## Module 2: Fault Injection Patterns (25 min)

### 2.1 Network-Level Faults

**Video: Simulating Network Failures** (8 min)

```go
type FaultInjector struct {
    targetProvider string
    faultType      FaultType
    enabled        atomic.Bool
}

type FaultType int

const (
    FaultTimeout FaultType = iota
    FaultConnectionRefused
    FaultDNSFailure
    FaultPartialResponse
    FaultHTTP500
    FaultHTTP429RateLimit
)
```

- Timeout: delay response beyond provider timeout threshold
- Connection refused: reject TCP connections to provider endpoint
- DNS failure: return NXDOMAIN for provider hostname
- Partial response: close connection mid-stream

### 2.2 Application-Level Faults

**Video: Injecting Failures at the Provider Layer** (8 min)

```go
type ChaosProvider struct {
    real        LLMProvider
    injector    *FaultInjector
}

func (p *ChaosProvider) Complete(ctx context.Context, req *Request) (*Response, error) {
    if p.injector.enabled.Load() {
        switch p.injector.faultType {
        case FaultTimeout:
            <-ctx.Done()
            return nil, ctx.Err()
        case FaultHTTP500:
            return nil, fmt.Errorf("internal server error (injected)")
        case FaultHTTP429RateLimit:
            return nil, &RateLimitError{RetryAfter: 60}
        case FaultPartialResponse:
            return &Response{Content: "partial..."}, io.ErrUnexpectedEOF
        }
    }
    return p.real.Complete(ctx, req)
}
```

### 2.3 Cascading Failure Simulation

**Video: Multi-Provider Failure Chains** (9 min)

- Inject failures sequentially across providers to test fallback depth
- Simulate correlated failures (e.g., all OpenAI-compatible providers fail)
- Test behavior when all providers are unavailable
- Verify error messages and client-facing responses during total failure

### Hands-On Lab 2

Implement a fault injector for a single provider:

1. Wrap a provider with a `ChaosProvider` that supports fault injection
2. Enable timeout injection and verify the request fails with context deadline
3. Enable HTTP 500 injection and verify the error response
4. Enable rate limit injection and verify the retry-after behavior

---

## Module 3: Provider Failure Simulation (25 min)

### 3.1 Single Provider Failure

**Video: Isolating One Provider** (8 min)

```go
func TestChaos_SingleProviderFailure(t *testing.T) {
    registry := setupTestRegistry(t, 3) // 3 providers
    injector := NewFaultInjector("provider-a", FaultHTTP500)

    // Measure steady state
    steadyLatency := measureAverageLatency(t, registry, 20)

    // Inject fault
    injector.Enable()
    defer injector.Disable()

    // Measure degraded state
    degradedLatency := measureAverageLatency(t, registry, 20)
    successRate := measureSuccessRate(t, registry, 50)

    // Verify hypothesis: latency increase < 50%, success rate > 95%
    latencyIncrease := (degradedLatency - steadyLatency) / steadyLatency
    assert.Less(t, latencyIncrease, 0.50, "latency degradation too severe")
    assert.Greater(t, successRate, 0.95, "success rate too low")
}
```

### 3.2 Majority Provider Failure

**Video: Testing with Most Providers Down** (8 min)

- Disable 2 of 3 providers and verify service continues
- Verify the remaining provider receives all traffic
- Check that circuit breakers open for failed providers
- Measure single-provider performance under full load

### 3.3 Intermittent Failures

**Video: Flapping Provider Simulation** (9 min)

```go
type IntermittentInjector struct {
    failureRate float64 // 0.0 to 1.0
    rng         *rand.Rand
}

func (i *IntermittentInjector) ShouldFail() bool {
    return i.rng.Float64() < i.failureRate
}
```

- 10% failure rate: validate retry logic handles occasional failures
- 50% failure rate: validate circuit breaker transitions correctly
- 90% failure rate: validate circuit breaker opens and stops calling provider

### Hands-On Lab 3

Simulate and verify single provider failure:

1. Set up 3 providers in a test registry
2. Record steady-state latency and success rate baselines
3. Inject HTTP 500 errors into provider A
4. Measure degraded latency and success rate
5. Verify the system routes around the failed provider

---

## Module 4: Circuit Breaker Validation (25 min)

### 4.1 Closed to Open Transition

**Video: Verifying Failure Detection** (8 min)

```go
func TestChaos_CircuitBreaker_OpensOnFailure(t *testing.T) {
    cb := NewCircuitBreaker(CircuitBreakerConfig{
        FailureThreshold: 3,
        RecoveryTimeout:  5 * time.Second,
    })

    // Circuit should be closed initially
    assert.Equal(t, StateClosed, cb.State())

    // Inject 3 failures
    for i := 0; i < 3; i++ {
        cb.RecordFailure()
    }

    // Circuit should be open
    assert.Equal(t, StateOpen, cb.State())

    // Requests should fail fast
    err := cb.Allow()
    assert.ErrorIs(t, err, ErrCircuitOpen)
}
```

### 4.2 Open to Half-Open Transition

**Video: Verifying Recovery Probing** (8 min)

- After recovery timeout, circuit transitions to half-open
- Half-open allows a limited number of probe requests
- Successful probes close the circuit; failures reopen it
- Test timing accuracy of the recovery timeout

### 4.3 Half-Open to Closed Recovery

**Video: Verifying Full Recovery** (9 min)

```go
func TestChaos_CircuitBreaker_RecoveryFlow(t *testing.T) {
    cb := NewCircuitBreaker(CircuitBreakerConfig{
        FailureThreshold: 3,
        SuccessThreshold: 2,
        RecoveryTimeout:  1 * time.Second,
    })

    // Open the circuit
    for i := 0; i < 3; i++ {
        cb.RecordFailure()
    }
    assert.Equal(t, StateOpen, cb.State())

    // Wait for recovery timeout
    time.Sleep(1100 * time.Millisecond)
    assert.Equal(t, StateHalfOpen, cb.State())

    // Record successful probes
    cb.RecordSuccess()
    cb.RecordSuccess()
    assert.Equal(t, StateClosed, cb.State())
}
```

### Hands-On Lab 4

Validate the complete circuit breaker lifecycle:

1. Create a circuit breaker with configurable thresholds
2. Inject failures to open the circuit
3. Verify requests fail fast while open
4. Wait for recovery timeout and verify half-open state
5. Record successes to close the circuit
6. Verify normal operation resumes

---

## Module 5: Recovery Verification and Chaos Tests (25 min)

### 5.1 Recovery Verification

**Video: Confirming System Heals** (8 min)

- After fault removal, system must return to steady state
- Measure time-to-recovery from fault removal to normal metrics
- Verify no permanent state corruption from fault injection
- Check that circuit breakers close and traffic rebalances

### 5.2 Writing a Complete Chaos Test

**Video: End-to-End Chaos Test for Provider Failover** (10 min)

```go
func TestChaos_ProviderFailover_EndToEnd(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping chaos test in short mode")
    }

    registry := setupTestRegistry(t, 3)
    injector := NewFaultInjector("provider-b", FaultTimeout)

    // Phase 1: Steady state baseline (10 requests)
    baseline := collectMetrics(t, registry, 10)
    assert.Greater(t, baseline.SuccessRate, 0.99)

    // Phase 2: Inject fault (20 requests)
    injector.Enable()
    degraded := collectMetrics(t, registry, 20)
    assert.Greater(t, degraded.SuccessRate, 0.90,
        "success rate dropped too much during fault")

    // Phase 3: Remove fault and verify recovery (10 requests)
    injector.Disable()
    time.Sleep(2 * time.Second) // Allow circuit breaker to probe

    recovered := collectMetrics(t, registry, 10)
    assert.Greater(t, recovered.SuccessRate, 0.99,
        "system did not recover after fault removal")

    // Phase 4: Verify latency returned to baseline range
    assert.InDelta(t, baseline.P50Latency.Milliseconds(),
        recovered.P50Latency.Milliseconds(), 500,
        "latency did not return to baseline")
}
```

### 5.3 Chaos Test Best Practices

**Video: Guidelines for Reliable Chaos Tests** (7 min)

- Always measure baseline before injecting faults
- Use deterministic fault injection (avoid random delays in assertions)
- Allow settling time between phases for circuit breaker transitions
- Clean up all injectors in deferred calls
- Run chaos tests with resource limits like stress tests

### Hands-On Lab 5

Write a chaos test for provider failover:

1. Set up a 3-provider ensemble
2. Measure steady-state baseline metrics
3. Inject timeout faults into the highest-ranked provider
4. Verify the system fails over to remaining providers
5. Remove the fault and verify full recovery
6. Assert that latency returns to within 20% of baseline

---

## Course Summary

### Key Takeaways

1. Chaos engineering validates fault tolerance through controlled experiments
2. Fault injection works at both network level (timeout, connection refused) and application level (wrapped providers)
3. Circuit breaker validation requires testing all three state transitions (closed, open, half-open)
4. Recovery verification confirms the system heals after faults are removed
5. Chaos tests follow a 4-phase pattern: baseline, fault injection, degraded observation, recovery verification
6. Blast radius control and automatic rollback are essential safety measures

### Assessment Questions

1. What are the three phases of a chaos engineering experiment?
2. Name three types of network-level faults that can be injected.
3. How does an intermittent fault injector differ from a permanent one?
4. Describe the circuit breaker state transitions that should occur during a chaos test.
5. Why is a settling time needed between fault injection phases?

### Related Courses

- Course 06: Testing
- Course 38: Stress Testing
- Course 56: Performance Optimization
- Course 57: Stress Testing Guide
- Course 59: Monitoring and Observability

---

**Course Version**: 1.0
**Last Updated**: March 8, 2026
