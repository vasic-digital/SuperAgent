#!/bin/bash
#
# Race Condition Detection Challenge - Challenge 001
# This challenge validates understanding and detection of race conditions
#

set -e

CHALLENGE_NAME="Race Condition Detection 001"
CHALLENGE_ID="race-001"
POINTS=10

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Setup temporary directory
setup() {
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    cd "$TEMP_DIR"
}

# Create test file with intentional race condition
create_race_file() {
    cat > race_example.go << 'EOF'
package main

import (
    "fmt"
    "sync"
)

// Counter with race condition
func main() {
    var counter int
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // Race condition: concurrent read/write
            counter++
        }()
    }
    
    wg.Wait()
    fmt.Printf("Counter: %d\n", counter)
}
EOF
}

# Create fixed version
create_fixed_file() {
    cat > race_fixed.go << 'EOF'
package main

import (
    "fmt"
    "sync"
)

// Counter without race condition
func main() {
    var counter int
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mu.Lock()
            counter++
            mu.Unlock()
        }()
    }
    
    wg.Wait()
    fmt.Printf("Counter: %d\n", counter)
}
EOF
}

# Test 1: Detect race in problematic code
test_race_detection() {
    log_info "Test 1: Detecting race condition..."
    
    create_race_file
    
    # Run with race detector
    if go run -race race_example.go 2>&1 | grep -q "DATA RACE"; then
        log_success "✓ Race condition detected by race detector"
        return 0
    else
        log_error "✗ Race condition not detected"
        return 1
    fi
}

# Test 2: Verify fixed code has no race
test_race_fixed() {
    log_info "Test 2: Verifying fixed code has no races..."
    
    create_fixed_file
    
    # Run with race detector - should produce no warnings
    if go run -race race_fixed.go 2>&1 | grep -q "DATA RACE"; then
        log_error "✗ Fixed code still has race condition"
        return 1
    else
        log_success "✓ Fixed code has no race conditions"
        return 0
    fi
}

# Test 3: Map race detection
test_map_race() {
    log_info "Test 3: Detecting map race condition..."
    
    cat > map_race.go << 'EOF'
package main

import (
    "fmt"
    "sync"
)

func main() {
    m := make(map[string]int)
    var wg sync.WaitGroup
    
    // Concurrent writes to map
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            m[fmt.Sprintf("key%d", n)] = n
        }(i)
    }
    
    wg.Wait()
    fmt.Println("Done")
}
EOF
    
    if go run -race map_race.go 2>&1 | grep -q "DATA RACE"; then
        log_success "✓ Map race condition detected"
        return 0
    else
        log_error "✗ Map race condition not detected"
        return 1
    fi
}

# Test 4: Channel race detection
test_channel_race() {
    log_info "Test 4: Detecting channel race condition..."
    
    cat > channel_race.go << 'EOF'
package main

import (
    "fmt"
    "sync"
)

func main() {
    var sum int
    ch := make(chan int, 10)
    var wg sync.WaitGroup
    
    // Producer
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            ch <- n
        }(i)
    }
    
    // Consumer with race
    go func() {
        for val := range ch {
            sum += val // Race here
        }
    }()
    
    wg.Wait()
    close(ch)
    fmt.Printf("Sum: %d\n", sum)
}
EOF
    
    if go run -race channel_race.go 2>&1 | grep -q "DATA RACE"; then
        log_success "✓ Channel race condition detected"
        return 0
    else
        log_error "✗ Channel race condition not detected"
        return 1
    fi
}

# Test 5: Benchmark overhead
test_race_overhead() {
    log_info "Test 5: Measuring race detector overhead..."
    
    cat > benchmark.go << 'EOF'
package main

import (
    "sync"
    "time"
)

func main() {
    var counter int
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    start := time.Now()
    
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mu.Lock()
            counter++
            mu.Unlock()
        }()
    }
    
    wg.Wait()
    elapsed := time.Since(start)
    
    // Just to use the variables
    _ = counter
    _ = elapsed
}
EOF
    
    # Run without race detector
    start=$(date +%s%N)
    go run benchmark.go
    end=$(date +%s%N)
    time_without=$(( (end - start) / 1000000 ))
    
    # Run with race detector
    start=$(date +%s%N)
    go run -race benchmark.go 2>/dev/null
    end=$(date +%s%N)
    time_with=$(( (end - start) / 1000000 ))
    
    overhead=$(( time_with * 100 / time_without - 100 ))
    
    log_info "Time without race detector: ${time_without}ms"
    log_info "Time with race detector: ${time_with}ms"
    log_info "Overhead: ${overhead}%"
    
    if [ $overhead -gt 0 ]; then
        log_success "✓ Race detector overhead measured: ${overhead}%"
        return 0
    else
        log_error "✗ Could not measure overhead"
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
    
    if test_race_detection; then
        ((passed++))
    else
        ((failed++))
    fi
    
    echo ""
    
    if test_race_fixed; then
        ((passed++))
    else
        ((failed++))
    fi
    
    echo ""
    
    if test_map_race; then
        ((passed++))
    else
        ((failed++))
    fi
    
    echo ""
    
    if test_channel_race; then
        ((passed++))
    else
        ((failed++))
    fi
    
    echo ""
    
    if test_race_overhead; then
        ((passed++))
    else
        ((failed++))
    fi
    
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

# Main execution
main() {
    run_all_tests
}

main "$@"
