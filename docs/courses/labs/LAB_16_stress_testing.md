# Lab 16: Stress Testing

## Objective
Run stress tests to validate system resilience under high concurrency and load.

## Prerequisites
- HelixAgent built (`make build`)
- Go test toolchain available

## Exercise 1: Run Concurrency Stress Tests

```bash
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -v -count=1 -timeout 60s ./tests/stress/ -run TestAPI
```

**Observe:** Goroutine count before/after, response times, panic count (should be 0).

## Exercise 2: Race Condition Detection

```bash
GOMAXPROCS=2 go test -race -count=1 -timeout 120s -short \
  ./internal/notifications/ \
  ./internal/mcp/ \
  ./internal/plugins/ \
  ./internal/cache/
```

**Expected:** All packages pass with zero data race warnings.

## Exercise 3: Provider Registry Under Load

```bash
GOMAXPROCS=2 go test -v -count=1 -timeout 60s ./tests/stress/ -run TestProvider
```

**Verify:** Concurrent GetProvider/ListProviders calls return consistent results.

## Exercise 4: Resource Limit Validation

Monitor system resources during test execution:

```bash
# In terminal 1: monitor resources
top -b -n 60 -d 1 | grep "go test" > /tmp/resource-log.txt &

# In terminal 2: run stress tests
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -p 1 -count=1 -timeout 120s ./tests/stress/

# Verify: CPU never exceeds 40% sustained
```

## Assessment Questions
1. Why is GOMAXPROCS=2 mandatory for stress tests in this project?
2. What does `nice -n 19` do and why is it important here?
3. How do you detect goroutine leaks in a stress test?
4. What is the difference between `-race` and stress testing?
