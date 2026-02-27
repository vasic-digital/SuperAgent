# Course-22: Deadlock Detection and Resolution

## Course Information
- **Duration:** 50 minutes
- **Level:** Advanced
- **Prerequisites:** Course-21

## Module 1: Understanding Deadlocks (12 min)

**Four Coffman Conditions:**
1. Mutual Exclusion
2. Hold and Wait
3. No Preemption
4. Circular Wait

**Visual:**
```
Goroutine A          Goroutine B
    │                    │
    ▼                    ▼
 Lock(X)               Lock(Y)
    │                    │
    ▼                    ▼
 Wait(Y) ───────────► Wait(X)
```

## Module 2: Common Deadlock Scenarios (15 min)

**Scenario 1: Lock Ordering**
```go
// BAD: Different order
go func() { mu1.Lock(); mu2.Lock() }()
go func() { mu2.Lock(); mu1.Lock() }() // Deadlock!

// GOOD: Same order
go func() { mu1.Lock(); mu2.Lock() }()
go func() { mu1.Lock(); mu2.Lock() }() // Safe
```

**Scenario 2: Channel Deadlock**
```go
ch := make(chan int)
ch <- 1 // Blocks forever - no receiver!
```

**Scenario 3: WaitGroup Deadlock**
```go
var wg sync.WaitGroup
wg.Add(1)
wg.Wait() // Never calls wg.Done()!
```

## Module 3: Detection Tools (10 min)

**Runtime Detection:**
```
fatal error: all goroutines are asleep - deadlock!
```

**HelixAgent Deadlock Detector:**
```go
detector := deadlock.NewDetector(5*time.Second, logger)
wrapped := detector.NewLockWrapper(&mu, "resource")
```

## Module 4: Prevention Strategies (10 min)

**Strategy 1: Lock Ordering**
```go
func transfer(from, to *Account) {
    if from.ID < to.ID {
        from.mu.Lock()
        to.mu.Lock()
    } else {
        to.mu.Lock()
        from.mu.Lock()
    }
    defer from.mu.Unlock()
    defer to.mu.Unlock()
}
```

**Strategy 2: Timeout Locks**
```go
if !mu.TryLock(timeout) {
    return errors.New("timeout")
}
```

**Strategy 3: Avoid Nested Locks**

## Assessment

**Lab:** Identify and fix 3 deadlocks in production code.
