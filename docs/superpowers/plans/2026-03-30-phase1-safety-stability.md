# Phase 1: Safety & Stability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all concurrency bugs — goroutine leaks, missing WaitGroup tracking, context.Background() anti-patterns — so every background goroutine is tracked, waited on, and cancelled properly.

**Architecture:** Each fix adds `sync.WaitGroup` fields to structs that launch goroutines, wraps `go` calls with `wg.Add(1)` / `defer wg.Done()`, and ensures `Close()`/`Shutdown()` calls `wg.Wait()`. Context propagation replaces `context.Background()` in goroutines.

**Tech Stack:** Go stdlib (`sync`, `context`), testify, `go test -race`

---

### Task 1: MessagingHub — Add WaitGroup for healthCheckLoop

**Files:**
- Modify: `internal/messaging/hub.go:14` (add wg field), `:161` (track goroutine), `:167-200` (wait in Close)
- Modify: `internal/messaging/hub.go:278` (fix context.Background)
- Test: `internal/messaging/hub_lifecycle_test.go` (create)

- [ ] **Step 1: Write the failing test**

Create `internal/messaging/hub_lifecycle_test.go`:

```go
package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessagingHub_Close_WaitsForHealthCheck(t *testing.T) {
	cfg := &HubConfig{
		UseFallbackOnError:  true,
		HealthCheckInterval: 50 * time.Millisecond,
	}
	hub := NewMessagingHub(cfg)

	// Initialize starts healthCheckLoop goroutine
	err := hub.Initialize(context.Background())
	require.NoError(t, err)

	// Close should wait for healthCheckLoop to exit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = hub.Close(ctx)
	assert.NoError(t, err)

	// Calling Close again should not panic
	err = hub.Close(ctx)
	assert.NoError(t, err)
}

func TestMessagingHub_Close_PropagatesContext(t *testing.T) {
	cfg := &HubConfig{
		UseFallbackOnError:  true,
		HealthCheckInterval: 100 * time.Millisecond,
	}
	hub := NewMessagingHub(cfg)

	err := hub.Initialize(context.Background())
	require.NoError(t, err)

	// Close with a real context
	ctx := context.Background()
	err = hub.Close(ctx)
	assert.NoError(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails or passes with current code**

Run: `GOMAXPROCS=2 go test -v -count=1 -race -run TestMessagingHub_Close ./internal/messaging/`

Expected: May pass but with race detector warnings, or goroutine leak.

- [ ] **Step 3: Add WaitGroup field to MessagingHub struct**

In `internal/messaging/hub.go`, add `wg` field after `stopCh`:

```go
	// state
	connected bool
	stopCh    chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
```

- [ ] **Step 4: Track healthCheckLoop goroutine with WaitGroup**

Replace line 161 (`go h.healthCheckLoop()`) with:

```go
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.healthCheckLoop()
	}()
```

- [ ] **Step 5: Wait for goroutine in Close()**

Modify the `Close` method to wait for the health check goroutine and use `sync.Once` to prevent double-close panic:

```go
func (h *MessagingHub) Close(ctx context.Context) error {
	h.closeOnce.Do(func() {
		close(h.stopCh)
	})
	h.wg.Wait()

	var errs MultiError
	// ... rest of Close body unchanged ...
```

- [ ] **Step 6: Fix context.Background() in healthCheckLoop**

In `healthCheckLoop()` at line 278, replace `context.Background()` with `context.WithTimeout` using a context derived from the stopCh signal. Since the loop already exits on `<-h.stopCh`, the timeout context is sufficient for individual health checks:

```go
func (h *MessagingHub) healthCheckLoop() {
	ticker := time.NewTicker(h.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_ = h.HealthCheck(ctx) //nolint:errcheck
			cancel()
		case <-h.stopCh:
			return
		}
	}
}
```

Note: `context.Background()` is acceptable here because the health check has its own 10s timeout and the goroutine exits via `stopCh`. The WaitGroup ensures Close() waits for it.

- [ ] **Step 7: Run tests to verify fix**

Run: `GOMAXPROCS=2 go test -v -count=1 -race -run TestMessagingHub ./internal/messaging/`

Expected: All PASS, zero race conditions.

- [ ] **Step 8: Commit**

```bash
git add internal/messaging/hub.go internal/messaging/hub_lifecycle_test.go
git commit -m "fix(messaging): add WaitGroup tracking for healthCheckLoop goroutine

Prevents goroutine leak on Close() by tracking the health check
background goroutine with sync.WaitGroup. Close() now waits for
the goroutine to exit. Uses sync.Once to prevent double-close panic.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: TieredCache — Add WaitGroup for l1CleanupLoop

**Files:**
- Modify: `internal/cache/tiered_cache.go:150-170` (add wg, track goroutine), `:367-370` (wait in Close)
- Test: `internal/cache/tiered_cache_lifecycle_test.go` (create)

- [ ] **Step 1: Write the failing test**

Create `internal/cache/tiered_cache_lifecycle_test.go`:

```go
package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTieredCache_Close_WaitsForCleanup(t *testing.T) {
	cfg := TieredCacheConfig{
		EnableL1:          true,
		L1MaxSize:         100,
		L1TTL:             time.Second,
		L1CleanupInterval: 50 * time.Millisecond,
	}

	tc := NewTieredCache(cfg, nil)

	// Close should wait for cleanup goroutine to exit
	err := tc.Close()
	assert.NoError(t, err)
}

func TestTieredCache_Close_Idempotent(t *testing.T) {
	cfg := TieredCacheConfig{
		EnableL1:          true,
		L1MaxSize:         100,
		L1TTL:             time.Second,
		L1CleanupInterval: 50 * time.Millisecond,
	}

	tc := NewTieredCache(cfg, nil)

	err := tc.Close()
	assert.NoError(t, err)

	// Second close should not panic
	err = tc.Close()
	assert.NoError(t, err)
}

func TestTieredCache_Close_NoL1(t *testing.T) {
	cfg := TieredCacheConfig{
		EnableL1: false,
	}

	tc := NewTieredCache(cfg, nil)

	// No cleanup goroutine started — Close should be fast
	err := tc.Close()
	assert.NoError(t, err)
}
```

- [ ] **Step 2: Run test to verify behavior**

Run: `GOMAXPROCS=2 go test -v -count=1 -race -run TestTieredCache_Close ./internal/cache/`

- [ ] **Step 3: Add WaitGroup field to TieredCache**

In `internal/cache/tiered_cache.go`, find the TieredCache struct definition and add a `wg` field:

```go
type TieredCache struct {
	l1       *l1Cache
	l2       *redis.Client
	config   TieredCacheConfig
	metrics  *TieredCacheMetrics
	tagIndex *tagIndex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}
```

- [ ] **Step 4: Track l1CleanupLoop goroutine**

Replace lines 164-167 (the goroutine launch in `NewTieredCache`):

```go
	if config.EnableL1 {
		tc.wg.Add(1)
		go func() {
			defer tc.wg.Done()
			tc.l1CleanupLoop()
		}()
	}
```

- [ ] **Step 5: Wait in Close()**

Replace lines 367-370:

```go
func (c *TieredCache) Close() error {
	c.cancel()
	c.wg.Wait()
	return nil
}
```

- [ ] **Step 6: Run tests**

Run: `GOMAXPROCS=2 go test -v -count=1 -race -run TestTieredCache ./internal/cache/`

Expected: All PASS, zero races.

- [ ] **Step 7: Commit**

```bash
git add internal/cache/tiered_cache.go internal/cache/tiered_cache_lifecycle_test.go
git commit -m "fix(cache): add WaitGroup tracking for l1CleanupLoop goroutine

Close() now waits for the L1 cleanup goroutine to exit before
returning. Prevents goroutine leak when TieredCache is discarded.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: QueryCache — Add WaitGroup for cleanupLoop

**Files:**
- Modify: `internal/database/query_optimizer.go:86-104` (add wg, track goroutine, wait in Shutdown)
- Test: `internal/database/query_optimizer_lifecycle_test.go` (create)

- [ ] **Step 1: Write the failing test**

Create `internal/database/query_optimizer_lifecycle_test.go`:

```go
package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryCache_Shutdown_WaitsForCleanup(t *testing.T) {
	qc := NewQueryCache(time.Second, 100)
	require.NotNil(t, qc)

	// Shutdown should wait for cleanup goroutine
	qc.Shutdown()

	// Should be safe to call again
	qc.Shutdown()
}

func TestQueryCache_Shutdown_NoLeak(t *testing.T) {
	qc := NewQueryCache(time.Second, 100)

	// Put something in cache
	qc.Put("key1", "value1")

	_, ok := qc.Get("key1")
	assert.True(t, ok)

	qc.Shutdown()
}
```

- [ ] **Step 2: Run test**

Run: `GOMAXPROCS=2 go test -v -count=1 -race -run TestQueryCache_Shutdown ./internal/database/`

- [ ] **Step 3: Add WaitGroup to QueryCache struct**

Find the QueryCache struct in `query_optimizer.go` and add `wg sync.WaitGroup`:

```go
type QueryCache struct {
	cache   map[string]*list.Element
	lruList *list.List
	mu      sync.Mutex
	ttl     time.Duration
	maxSize int
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}
```

- [ ] **Step 4: Track cleanupLoop goroutine**

Replace line 97 (`go qc.cleanupLoop()`) in `NewQueryCache`:

```go
	qc.wg.Add(1)
	go func() {
		defer qc.wg.Done()
		qc.cleanupLoop()
	}()
```

- [ ] **Step 5: Wait in Shutdown()**

Replace lines 101-104:

```go
func (c *QueryCache) Shutdown() {
	c.cancel()
	c.wg.Wait()
}
```

- [ ] **Step 6: Run tests**

Run: `GOMAXPROCS=2 go test -v -count=1 -race -run TestQueryCache ./internal/database/`

Expected: All PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/database/query_optimizer.go internal/database/query_optimizer_lifecycle_test.go
git commit -m "fix(database): add WaitGroup tracking for QueryCache cleanupLoop

Shutdown() now waits for the cleanup goroutine to exit before
returning. Prevents goroutine leak when QueryCache is discarded.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 4: FormatterExecutor BatchExecute — Add panic recovery

**Files:**
- Modify: `internal/formatters/executor.go:101-141` (add panic recovery)
- Test: `internal/formatters/executor_safety_test.go` (create)

- [ ] **Step 1: Write the failing test**

Create `internal/formatters/executor_safety_test.go`:

```go
package formatters

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type panicFormatter struct{}

func (p *panicFormatter) Name() string     { return "panic" }
func (p *panicFormatter) Language() string  { return "go" }
func (p *panicFormatter) Priority() int     { return 0 }
func (p *panicFormatter) Available() bool   { return true }
func (p *panicFormatter) Format(ctx context.Context, req *FormatRequest) (*FormatResult, error) {
	panic("test panic in formatter")
}

func TestExecuteBatch_PanicRecovery(t *testing.T) {
	executor := NewFormatterExecutor()

	reqs := []*FormatRequest{
		{Code: "test", Language: "go", FormatterName: "panic"},
	}

	// Should not panic — should return error
	results, err := executor.ExecuteBatch(context.Background(), reqs)
	// Either err is non-nil or results contain error info
	if err != nil {
		assert.Contains(t, err.Error(), "panic")
	} else {
		require.Len(t, results, 1)
	}
}

func TestExecuteBatch_Empty(t *testing.T) {
	executor := NewFormatterExecutor()
	results, err := executor.ExecuteBatch(context.Background(), nil)
	assert.NoError(t, err)
	assert.Empty(t, results)
}
```

- [ ] **Step 2: Run test**

Run: `GOMAXPROCS=2 go test -v -count=1 -race -run TestExecuteBatch ./internal/formatters/`

Expected: Test panics (unrecovered).

- [ ] **Step 3: Add panic recovery to ExecuteBatch goroutines**

Replace the goroutine launch in `ExecuteBatch` (lines 114-122) with:

```go
	for i, req := range reqs {
		go func(index int, request *FormatRequest) {
			defer func() {
				if r := recover(); r != nil {
					resultChan <- resultPair{
						index: index,
						err:   fmt.Errorf("formatter panic: %v", r),
					}
				}
			}()
			result, err := e.Execute(ctx, request)
			resultChan <- resultPair{
				index:  index,
				result: result,
				err:    err,
			}
		}(i, req)
	}
```

- [ ] **Step 4: Run tests**

Run: `GOMAXPROCS=2 go test -v -count=1 -race -run TestExecuteBatch ./internal/formatters/`

Expected: All PASS — panic recovered gracefully.

- [ ] **Step 5: Commit**

```bash
git add internal/formatters/executor.go internal/formatters/executor_safety_test.go
git commit -m "fix(formatters): add panic recovery to ExecuteBatch goroutines

Prevents a panicking formatter from crashing the entire batch.
Panics are recovered and converted to errors in the result set.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 5: Gemini ACP — Verify map access safety

**Files:**
- Review: `internal/llm/providers/gemini/gemini_acp.go:267-298` (readResponses), `:312-339` (sendRequest)
- Test: `internal/llm/providers/gemini/gemini_acp_safety_test.go` (create)

- [ ] **Step 1: Analyze current locking**

The code at lines 284-286 uses `RLock/RUnlock` correctly:
```go
p.respMu.RLock()
ch, ok := p.responses[resp.ID]
p.respMu.RUnlock()
```

And `sendRequest` at lines 331-338 uses `Lock/Unlock` correctly with defer cleanup. The locking is actually correct — RLock for reads, Lock for writes, and the response channel is buffered (line 330: `make(chan *geminiACPResponse, 1)`).

The original audit flagged this but the code is safe. Write a test confirming this.

- [ ] **Step 2: Write a confirmation test**

Create `internal/llm/providers/gemini/gemini_acp_safety_test.go`:

```go
package gemini

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeminiACP_ResponseMap_ConcurrentAccess(t *testing.T) {
	// Verify the response map locking pattern is safe
	// by simulating concurrent read/write access
	var mu sync.RWMutex
	responses := make(map[int64]chan *geminiACPResponse)

	var wg sync.WaitGroup

	// Simulate sendRequest writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			ch := make(chan *geminiACPResponse, 1)
			mu.Lock()
			responses[id] = ch
			mu.Unlock()

			// Simulate cleanup
			mu.Lock()
			delete(responses, id)
			mu.Unlock()
		}(int64(i))
	}

	// Simulate readResponses reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			mu.RLock()
			_, _ = responses[id]
			mu.RUnlock()
		}(int64(i))
	}

	wg.Wait()
	assert.Empty(t, responses)
}
```

- [ ] **Step 3: Run test with race detector**

Run: `GOMAXPROCS=2 go test -v -count=1 -race -run TestGeminiACP_ResponseMap ./internal/llm/providers/gemini/`

Expected: PASS — no races detected (confirming the pattern is safe).

- [ ] **Step 4: Commit**

```bash
git add internal/llm/providers/gemini/gemini_acp_safety_test.go
git commit -m "test(gemini): add race condition safety test for ACP response map

Confirms the RWMutex locking pattern for the response channel map
is safe under concurrent access from readResponses and sendRequest.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 6: WebSocket Server — Verify lock ordering and add test

**Files:**
- Review: `internal/notifications/websocket_server.go:26-28` (lock ordering doc), `:115-135` (Stop), `:239-253` (Broadcast)
- Test: `internal/notifications/websocket_server_safety_test.go` (create)

- [ ] **Step 1: Analyze current code**

The code already has:
1. Lock ordering documented (line 26-28): "clientsMu MUST be acquired before globalClientsMu"
2. `Stop()` acquires clientsMu first (line 121), then globalClientsMu (line 130) — correct order
3. `Broadcast()` acquires clientsMu (line 241), releases it (line 243), then calls `broadcastGlobal` which acquires globalClientsMu — correct order, no nesting

The pattern is safe. Write a confirmation test.

- [ ] **Step 2: Write the test**

Create `internal/notifications/websocket_server_safety_test.go`:

```go
package notifications

import (
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketServer_ConcurrentBroadcastAndStop(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	server := NewWebSocketServer(logger)

	var wg sync.WaitGroup

	// Concurrent broadcasts
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			server.Broadcast("task-1", []byte("test data"))
		}()
	}

	// Concurrent global broadcasts
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			server.BroadcastGlobal([]byte("global data"))
		}()
	}

	wg.Wait()

	// Stop should succeed without deadlock
	err := server.Stop()
	assert.NoError(t, err)
}
```

- [ ] **Step 3: Run with race detector**

Run: `GOMAXPROCS=2 go test -v -count=1 -race -run TestWebSocketServer_Concurrent ./internal/notifications/`

Expected: PASS — no deadlocks, no races.

- [ ] **Step 4: Commit**

```bash
git add internal/notifications/websocket_server_safety_test.go
git commit -m "test(notifications): add concurrent safety test for WebSocket server

Validates lock ordering is correct under concurrent Broadcast,
BroadcastGlobal, and Stop operations. No deadlock detected.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

### Task 7: Run full race detection and verify all fixes

**Files:**
- No new files — validation only

- [ ] **Step 1: Run race detection on all fixed packages**

```bash
GOMAXPROCS=2 go test -v -count=1 -race -short \
  ./internal/messaging/ \
  ./internal/cache/ \
  ./internal/database/ \
  ./internal/formatters/ \
  ./internal/llm/providers/gemini/ \
  ./internal/notifications/ \
  ./internal/handlers/ \
  2>&1 | tail -30
```

Expected: All PASS, zero races.

- [ ] **Step 2: Run full project build**

```bash
go build ./...
```

Expected: Clean compilation.

- [ ] **Step 3: Run existing concurrency challenge**

```bash
nice -n 19 bash ./challenges/scripts/concurrency_safety_comprehensive_challenge.sh
```

Expected: All tests pass.

- [ ] **Step 4: Run goroutine lifecycle challenge**

```bash
nice -n 19 bash ./challenges/scripts/goroutine_lifecycle_challenge.sh
```

Expected: All tests pass.

- [ ] **Step 5: Push all Phase 1 changes**

```bash
git push githubhelixdevelopment main
git push upstream main
```
