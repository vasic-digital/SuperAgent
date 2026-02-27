# Course-27: Performance Benchmarking

## Course Information
- **Duration:** 50 minutes
- **Level:** Advanced
- **Prerequisites:** Course-26

## Module 1: Benchmarking Basics (10 min)

**Why Benchmark?**
- Measure performance
- Detect regressions
- Optimize code
- Compare implementations

**Go Benchmarking:**
```go
func BenchmarkMyFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        MyFunction()
    }
}
```

**Running:**
```bash
go test -bench=.
go test -bench=. -benchmem
go test -bench=. -cpuprofile=cpu.prof
```

## Module 2: Writing Good Benchmarks (15 min)

**Setup and Teardown:**
```go
func BenchmarkWithSetup(b *testing.B) {
    // Setup (runs once)
    data := setupData()
    b.ResetTimer() // Reset timer after setup
    
    for i := 0; i < b.N; i++ {
        Process(data)
    }
    
    // Cleanup
    cleanup(data)
}
```

**Sub-Benchmarks:**
```go
func BenchmarkAlgorithms(b *testing.B) {
    algorithms := []struct {
        name string
        fn   func([]int)
    }{
        {"sort", sortInts},
        {"heap", heapSort},
    }
    
    for _, alg := range algorithms {
        b.Run(alg.name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                alg.fn(data)
            }
        })
    }
}
```

**Parallel Benchmarks:**
```go
func BenchmarkParallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            MyFunction()
        }
    })
}
```

## Module 3: Analyzing Results (15 min)

**Understanding Output:**
```
BenchmarkFunction-8    1000000    1050 ns/op    256 B/op    5 allocs/op
```

- `1000000`: Number of iterations
- `1050 ns/op`: Time per operation
- `256 B/op`: Bytes allocated per operation
- `5 allocs/op`: Allocations per operation

**Comparing Benchmarks:**
```bash
go test -bench=. > old.txt
# Make changes
go test -bench=. > new.txt

benchcmp old.txt new.txt
```

## Module 4: Continuous Benchmarking (10 min)

**GitHub Actions:**
```yaml
- name: Run benchmarks
  run: go test -bench=. -benchmem | tee benchmark.txt

- name: Compare benchmarks
  uses: benchmark-action/github-action-benchmark@v1
  with:
    tool: 'go'
    output-file-path: benchmark.txt
```

**Alerting on Regression:**
```go
// benchstat analysis
// +10% = Warning
// +25% = Error
```

## Assessment

**Lab:** Benchmark 3 implementations and optimize.
