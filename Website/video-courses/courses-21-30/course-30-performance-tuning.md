# Course-30: Performance Tuning Best Practices

## Course Information
- **Duration:** 45 minutes
- **Level:** Advanced
- **Prerequisites:** Course-29

## Module 1: Measurement First (10 min)

**Golden Rule:** "Measure, don't guess"

**Tools:**
- Benchmarks
- Profiling
- Tracing
- Logging

**Methodology:**
1. Establish baseline
2. Identify bottleneck
3. Optimize
4. Measure improvement
5. Repeat

## Module 2: Common Pitfalls (15 min)

**Pitfall 1: Premature Optimization**
```go
// Don't optimize without profiling!
```

**Pitfall 2: Over-Optimization**
- Complex code
- Hard to maintain
- Minimal gains

**Pitfall 3: Ignoring Real-World Conditions**
- Test with production data
- Simulate real load
- Consider network latency

**Pitfall 4: Cache Thrashing**
```go
// BAD: Cache is too small
// GOOD: Size cache appropriately
```

## Module 3: Production Tuning (12 min)

**GC Tuning:**
```bash
# Adjust GC target
GOGC=100  # Default
GOGC=200  # Less frequent GC
```

**Runtime Settings:**
```go
runtime.GOMAXPROCS(numCPU)
```

**System Tuning:**
- File descriptors: `ulimit -n`
- TCP settings
- Kernel parameters

## Module 4: Monitoring in Production (8 min)

**Key Metrics:**
- Response time (p50, p95, p99)
- Throughput (req/sec)
- Error rate
- Resource usage

**Alerting:**
- Latency spikes
- Error rate increases
- Resource exhaustion

**Dashboard:**
- Real-time metrics
- Historical trends
- Capacity planning

## Assessment

**Lab:** Tune a production-like system.
