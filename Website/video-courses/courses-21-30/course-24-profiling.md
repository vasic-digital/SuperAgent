# Course-24: Profiling and Performance Analysis

## Course Information
- **Duration:** 50 minutes
- **Level:** Advanced
- **Prerequisites:** Course-23

## Module 1: Go Profiling Overview (10 min)

**Profile Types:**
- CPU Profile
- Memory Profile (Heap)
- Goroutine Profile
- Block Profile
- Mutex Profile
- Thread Create Profile

## Module 2: CPU Profiling (15 min)

**Code Integration:**
```go
import "runtime/pprof"

func main() {
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
    
    // Your code here
}
```

**Analysis:**
```bash
go tool pprof cpu.prof
(pprof) top
(pprof) list functionName
(pprof) web
```

## Module 3: Memory Profiling (15 min)

**Capture Heap:**
```go
f, _ := os.Create("heap.prof")
pprof.WriteHeapProfile(f)
f.Close()
```

**Analysis:**
```bash
go tool pprof -alloc_space heap.prof
go tool pprof -inuse_objects heap.prof
```

## Module 4: Continuous Profiling (10 min)

**HTTP Endpoint:**
```go
import _ "net/http/pprof"

http.ListenAndServe("localhost:6060", nil)
```

**Access Profiles:**
```
/debug/pprof/heap
/debug/pprof/goroutine
/debug/pprof/profile?seconds=30
```

## Assessment

**Lab:** Profile a slow application and optimize it.
