#!/bin/bash
#
# Memory Leak Detection Challenge - Challenge 003
# This challenge validates understanding of memory leaks in Go
#

set -e

CHALLENGE_NAME="Memory Leak Detection 003"
CHALLENGE_ID="memory-003"
POINTS=20

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

setup() {
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    cd "$TEMP_DIR"
}

# Test 1: Detect goroutine leak
test_goroutine_leak() {
    log_info "Test 1: Detecting goroutine leak..."
    
    cat > goroutine_leak.go << 'EOF'
package main

import (
    "runtime"
    "time"
)

func leaky() {
    ch := make(chan int)
    go func() {
        // This goroutine will block forever
        <-ch
    }()
}

func main() {
    before := runtime.NumGoroutine()
    
    for i := 0; i < 100; i++ {
        leaky()
    }
    
    time.Sleep(100 * time.Millisecond)
    after := runtime.NumGoroutine()
    
    // Should have 100 more goroutines
    if after > before+90 {
        println("LEAK DETECTED: Goroutines increased from", before, "to", after)
    }
}
EOF
    
    output=$(go run goroutine_leak.go 2>&1)
    if echo "$output" | grep -q "LEAK DETECTED"; then
        log_success "✓ Goroutine leak detected"
        return 0
    else
        log_error "✗ Goroutine leak not detected"
        return 1
    fi
}

# Test 2: Detect slice growth leak
test_slice_leak() {
    log_info "Test 2: Detecting slice growth leak..."
    
    cat > slice_leak.go << 'EOF'
package main

import (
    "fmt"
    "runtime"
)

type LargeStruct struct {
    data [1024 * 1024]byte // 1MB
}

func process() []*LargeStruct {
    var results []*LargeStruct
    
    for i := 0; i < 100; i++ {
        results = append(results, &LargeStruct{})
    }
    
    // Only return first 10, but slice keeps reference to all 100
    return results[:10]
}

func main() {
    var m1 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    _ = process()
    
    runtime.GC()
    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)
    
    // Should still hold ~100MB due to slice backing array
    held := (m2.HeapAlloc - m1.HeapAlloc) / 1024 / 1024
    if held > 50 {
        fmt.Printf("LEAK DETECTED: Holding %d MB\n", held)
    }
}
EOF
    
    output=$(go run slice_leak.go 2>&1)
    if echo "$output" | grep -q "LEAK DETECTED"; then
        log_success "✓ Slice leak detected"
        return 0
    else
        log_error "✗ Slice leak not detected"
        return 1
    fi
}

# Test 3: Detect channel leak
test_channel_leak() {
    log_info "Test 3: Detecting channel leak..."
    
    cat > channel_leak.go << 'EOF'
package main

import (
    "fmt"
    "runtime"
    "time"
)

func spawnWorker() chan int {
    ch := make(chan int, 100)
    go func() {
        // Worker processes messages
        for range ch {
            time.Sleep(time.Millisecond)
        }
    }()
    return ch
}

func main() {
    before := runtime.NumGoroutine()
    
    // Create workers but never use or close them
    workers := make([]chan int, 0)
    for i := 0; i < 50; i++ {
        workers = append(workers, spawnWorker())
    }
    
    time.Sleep(100 * time.Millisecond)
    after := runtime.NumGoroutine()
    
    // Workers and goroutines still active
    if after > before+40 {
        fmt.Printf("LEAK DETECTED: %d goroutines leaked\n", after-before)
    }
    
    _ = workers // Suppress unused warning
}
EOF
    
    output=$(go run channel_leak.go 2>&1)
    if echo "$output" | grep -q "LEAK DETECTED"; then
        log_success "✓ Channel leak detected"
        return 0
    else
        log_error "✗ Channel leak not detected"
        return 1
    fi
}

# Test 4: Fix goroutine leak with context
test_goroutine_fix() {
    log_info "Test 4: Fixing goroutine leak with context..."
    
    cat > goroutine_fixed.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "runtime"
    "time"
)

func fixed(ctx context.Context, done chan bool) {
    select {
    case <-done:
        println("Worker: Exiting cleanly")
        return
    case <-ctx.Done():
        println("Worker: Context cancelled")
        return
    }
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    before := runtime.NumGoroutine()
    
    for i := 0; i < 10; i++ {
        done := make(chan bool)
        go fixed(ctx, done)
        close(done) // Signal to exit
    }
    
    time.Sleep(200 * time.Millisecond)
    runtime.GC()
    
    after := runtime.NumGoroutine()
    
    if after <= before+5 {
        fmt.Println("SUCCESS: Goroutines cleaned up properly")
    } else {
        fmt.Printf("LEAK: Still have %d extra goroutines\n", after-before)
    }
}
EOF
    
    output=$(go run goroutine_fixed.go 2>&1)
    if echo "$output" | grep -q "SUCCESS"; then
        log_success "✓ Goroutine leak fixed"
        return 0
    else
        log_error "✗ Fix didn't work"
        return 1
    fi
}

# Test 5: Using pprof for leak detection
test_pprof_detection() {
    log_info "Test 5: Using pprof for memory leak detection..."
    
    cat > pprof_leak.go << 'EOF'
package main

import (
    "fmt"
    "net/http"
    _ "net/http/pprof"
    "runtime"
    "time"
)

func leakyFunction() {
    // Allocate memory and keep reference
    leak := make([]byte, 10*1024*1024) // 10MB
    _ = leak
}

func main() {
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()
    
    var m1 runtime.MemStats
    runtime.ReadMemStats(&m1)
    
    // Simulate leak
    for i := 0; i < 10; i++ {
        leakyFunction()
        time.Sleep(10 * time.Millisecond)
    }
    
    runtime.GC()
    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)
    
    growth := int64(m2.HeapAlloc) - int64(m1.HeapAlloc)
    if growth > 50*1024*1024 {
        fmt.Printf("LEAK DETECTED: Memory grew by %d MB\n", growth/1024/1024)
        fmt.Println("Use: go tool pprof http://localhost:6060/debug/pprof/heap")
    }
}
EOF
    
    timeout 2 go run pprof_leak.go 2>&1 | grep -q "LEAK DETECTED"
    if [ $? -eq 0 ]; then
        log_success "✓ Pprof leak detection working"
        return 0
    else
        log_error "✗ Pprof detection not working"
        return 1
    fi
}

# Run all tests
run_all_tests() {
    echo "=========================================="
    echo "🏁 Challenge: $CHALLENGE_NAME"
    echo "ID: $CHALLENGE_ID"
    echo "Points: $POINTS"
    echo "=========================================="
    echo ""
    
    local passed=0
    local failed=0
    
    setup
    
    if test_goroutine_leak; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_slice_leak; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_channel_leak; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_goroutine_fix; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_pprof_detection; then ((passed++)); else ((failed++)); fi
    echo ""
    
    echo "=========================================="
    echo "📊 Challenge Results"
    echo "=========================================="
    echo -e "${GREEN}Passed: $passed${NC}"
    echo -e "${RED}Failed: $failed${NC}"
    echo "Total: 5"
    echo ""
    
    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}🎉 Challenge Complete! +$POINTS points${NC}"
        return 0
    else
        echo -e "${RED}❌ Challenge Incomplete${NC}"
        return 1
    fi
}

main() {
    run_all_tests
}

main "$@"
