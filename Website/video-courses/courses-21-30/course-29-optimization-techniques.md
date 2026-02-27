# Course-29: Optimization Techniques

## Course Information
- **Duration:** 50 minutes
- **Level:** Advanced
- **Prerequisites:** Course-28

## Module 1: Algorithm Optimization (12 min)

**Big O Analysis:**
- O(1) - Constant time
- O(log n) - Logarithmic
- O(n) - Linear
- O(n log n) - Linearithmic
- O(n²) - Quadratic

**Example Optimization:**
```go
// O(n²) - Slow
func hasDuplicateSlow(items []int) bool {
    for i := 0; i < len(items); i++ {
        for j := i + 1; j < len(items); j++ {
            if items[i] == items[j] {
                return true
            }
        }
    }
    return false
}

// O(n) - Fast
func hasDuplicateFast(items []int) bool {
    seen := make(map[int]bool)
    for _, item := range items {
        if seen[item] {
            return true
        }
        seen[item] = true
    }
    return false
}
```

## Module 2: Memory Optimization (15 min)

**Reduce Allocations:**
```go
// BAD: Allocates on every call
func process(items []int) []int {
    result := make([]int, 0)
    for _, item := range items {
        result = append(result, item*2)
    }
    return result
}

// GOOD: Pre-allocate
func processFast(items []int) []int {
    result := make([]int, 0, len(items))
    for _, item := range items {
        result = append(result, item*2)
    }
    return result
}
```

**Object Pool:**
```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

func process() {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    
    // Use buffer
}
```

## Module 3: Concurrency Optimization (13 min)

**Worker Pools:**
```go
func workerPool(jobs <-chan Job, workers int) <-chan Result {
    results := make(chan Result)
    var wg sync.WaitGroup
    
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- process(job)
            }
        }()
    }
    
    go func() {
        wg.Wait()
        close(results)
    }()
    
    return results
}
```

**Batch Processing:**
```go
func processBatch(items []Item, batchSize int) {
    for i := 0; i < len(items); i += batchSize {
        end := i + batchSize
        if end > len(items) {
            end = len(items)
        }
        batch := items[i:end]
        process(batch)
    }
}
```

## Module 4: Database Optimization (10 min)

**Connection Pooling:**
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
```

**Query Optimization:**
- Use indexes
- Select only needed columns
- Use prepared statements
- Batch inserts

## Assessment

**Lab:** Optimize a slow service by 50%.
