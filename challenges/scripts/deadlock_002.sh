#!/bin/bash
#
# Deadlock Detection Challenge - Challenge 002
# This challenge validates understanding and detection of deadlocks
#

set -e

CHALLENGE_NAME="Deadlock Detection 002"
CHALLENGE_ID="deadlock-002"
POINTS=15

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

# Test 1: Detect simple deadlock
test_simple_deadlock() {
    log_info "Test 1: Detecting simple deadlock..."
    
    cat > deadlock_simple.go << 'EOF'
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    var mu1, mu2 sync.Mutex
    
    go func() {
        mu1.Lock()
        fmt.Println("G1: Locked mu1")
        time.Sleep(100 * time.Millisecond)
        mu2.Lock()
        fmt.Println("G1: Locked mu2")
        mu2.Unlock()
        mu1.Unlock()
    }()
    
    go func() {
        mu2.Lock()
        fmt.Println("G2: Locked mu2")
        time.Sleep(100 * time.Millisecond)
        mu1.Lock()
        fmt.Println("G2: Locked mu1")
        mu1.Unlock()
        mu2.Unlock()
    }()
    
    time.Sleep(2 * time.Second)
    fmt.Println("Done")
}
EOF
    
    timeout 3 go run deadlock_simple.go 2>&1 | grep -q "fatal error: all goroutines are asleep - deadlock!"
    if [ $? -eq 0 ]; then
        log_success "✓ Simple deadlock detected"
        return 0
    else
        log_error "✗ Deadlock not detected"
        return 1
    fi
}

# Test 2: Fix deadlock with ordering
test_deadlock_fix() {
    log_info "Test 2: Fixing deadlock with lock ordering..."
    
    cat > deadlock_fixed.go << 'EOF'
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    var mu1, mu2 sync.Mutex
    done := make(chan bool, 2)
    
    // Both goroutines lock in same order
    go func() {
        mu1.Lock()
        fmt.Println("G1: Locked mu1")
        time.Sleep(10 * time.Millisecond)
        mu2.Lock()
        fmt.Println("G1: Locked mu2")
        mu2.Unlock()
        mu1.Unlock()
        done <- true
    }()
    
    go func() {
        mu1.Lock()  // Same order!
        fmt.Println("G2: Locked mu1")
        time.Sleep(10 * time.Millisecond)
        mu2.Lock()
        fmt.Println("G2: Locked mu2")
        mu2.Unlock()
        mu1.Unlock()
        done <- true
    }()
    
    <-done
    <-done
    fmt.Println("Success: No deadlock!")
}
EOF
    
    if timeout 3 go run deadlock_fixed.go 2>&1 | grep -q "Success"; then
        log_success "✓ Deadlock fixed with ordering"
        return 0
    else
        log_error "✗ Fix didn't work"
        return 1
    fi
}

# Test 3: Channel deadlock
test_channel_deadlock() {
    log_info "Test 3: Detecting channel deadlock..."
    
    cat > channel_deadlock.go << 'EOF'
package main

func main() {
    ch := make(chan int)
    ch <- 1  // Blocks forever - no receiver
    <-ch
}
EOF
    
    timeout 2 go run channel_deadlock.go 2>&1 | grep -q "fatal error: all goroutines are asleep - deadlock!"
    if [ $? -eq 0 ]; then
        log_success "✓ Channel deadlock detected"
        return 0
    else
        log_error "✗ Channel deadlock not detected"
        return 1
    fi
}

# Test 4: WaitGroup deadlock
test_waitgroup_deadlock() {
    log_info "Test 4: Detecting WaitGroup deadlock..."
    
    cat > waitgroup_deadlock.go << 'EOF'
package main

import "sync"

func main() {
    var wg sync.WaitGroup
    wg.Add(1)
    // Never call wg.Done() - deadlock!
    wg.Wait()
}
EOF
    
    timeout 2 go run waitgroup_deadlock.go 2>&1 | grep -q "fatal error: all goroutines are asleep - deadlock!"
    if [ $? -eq 0 ]; then
        log_success "✓ WaitGroup deadlock detected"
        return 0
    else
        log_error "✗ WaitGroup deadlock not detected"
        return 1
    fi
}

# Test 5: Select deadlock
test_select_deadlock() {
    log_info "Test 5: Detecting select deadlock..."
    
    cat > select_deadlock.go << 'EOF'
package main

func main() {
    ch1 := make(chan int)
    ch2 := make(chan int)
    
    select {
    case <-ch1:
    case <-ch2:
    }
}
EOF
    
    timeout 2 go run select_deadlock.go 2>&1 | grep -q "fatal error: all goroutines are asleep - deadlock!"
    if [ $? -eq 0 ]; then
        log_success "✓ Select deadlock detected"
        return 0
    else
        log_error "✗ Select deadlock not detected"
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
    
    if test_simple_deadlock; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_deadlock_fix; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_channel_deadlock; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_waitgroup_deadlock; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_select_deadlock; then ((passed++)); else ((failed++)); fi
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
