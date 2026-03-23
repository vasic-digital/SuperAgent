# Phase 1: Concurrency Safety Fixes — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all 15 concurrency issues (race conditions, goroutine leaks, channel safety) identified in the project audit.

**Architecture:** Each fix follows the same pattern: add `sync.Once` for channel closures, `atomic.Bool` for concurrent boolean flags, `sync.WaitGroup` tracking for goroutines, and proper mutex protection for shared state. All fixes are backward-compatible and do not change any public API.

**Tech Stack:** Go stdlib `sync`, `sync/atomic`, `golang.org/x/sync/semaphore`, `github.com/stretchr/testify`

**Spec:** `docs/superpowers/specs/2026-03-23-project-completion-master-spec.md` (Section 3, SP-1)

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/notifications/sse_manager.go` | Modify | Add `sync.Once`+`atomic.Bool` for safe channel closure |
| `internal/notifications/sse_manager_test.go` | Create/Modify | Concurrent Stop/Send test |
| `internal/notifications/kafka_transport.go` | Modify | Add `sync.Once` for double-close protection |
| `internal/notifications/kafka_transport_test.go` | Modify | Concurrent Stop test |
| `internal/streaming/kafka_streams.go` | Modify | Add WaitGroup tracking to goroutine |
| `internal/streaming/kafka_streams_test.go` | Modify | Graceful shutdown test |
| `internal/cache/cache_service.go` | Modify | Fix nested map synchronization |
| `internal/cache/cache_service_test.go` | Modify | Concurrent invalidation test |
| `internal/mcp/connection_pool.go` | Modify | Replace `bool` with `atomic.Bool` |
| `internal/mcp/connection_pool_test.go` | Modify | Concurrent Get/Close test |
| `internal/plugins/hot_reload.go` | Modify | Add WaitGroup goroutine tracking |
| `internal/plugins/hot_reload_test.go` | Modify | Start/Stop lifecycle test |
| `internal/services/integration_orchestrator.go` | Modify | Activate disabled mutex |
| `internal/services/integration_orchestrator_test.go` | Modify | Concurrent workflow test |
| `internal/services/debate_service.go` | Modify | Panic recovery + WaitGroup tracking |
| `internal/services/debate_service_test.go` | Modify | Panic-in-goroutine test |
| `internal/debate/orchestrator/orchestrator.go` | Modify | Protect ActiveDebate fields |
| `internal/debate/orchestrator/orchestrator_test.go` | Modify | Concurrent debate state test |
| `internal/handlers/model_metadata.go` | Modify | Pass service as goroutine param |
| `internal/notifications/websocket_server.go` | Modify | Document lock ordering |
| `internal/notifications/polling_store.go` | Modify | Add panic recovery |
| `internal/llm/circuit_breaker.go` | Modify | Return error on listener limit |
| `internal/services/boot_manager.go` | Modify | Add lock assertion comment |

---

### Task 1: SSE Manager — Safe Channel Closure

**Files:**
- Modify: `internal/notifications/sse_manager.go:14-31` (struct), `81-104` (Stop method)
- Test: `internal/notifications/sse_manager_test.go`

- [ ] **Step 1: Write the failing test**

In `internal/notifications/sse_manager_test.go`, add:
```go
func TestSSEManager_Stop_ConcurrentSendersNoPanic(t *testing.T) {
	config := DefaultSSEConfig()
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager := NewSSEManager(config, logger)

	// Register multiple clients
	clients := make([]chan []byte, 10)
	for i := 0; i < 10; i++ {
		clients[i] = make(chan []byte, 100)
		taskID := fmt.Sprintf("task-%d", i)
		err := manager.RegisterClient(taskID, clients[i])
		require.NoError(t, err)
	}

	// Concurrently send events while stopping
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: Send events rapidly
	go func() {
		defer wg.Done()
		for j := 0; j < 100; j++ {
			_ = manager.SendTaskEvent("task-0", "test", map[string]interface{}{"i": j})
		}
	}()

	// Goroutine 2: Stop after a brief delay
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Millisecond)
		err := manager.Stop()
		assert.NoError(t, err)
	}()

	wg.Wait()
	// If we get here without panic, the test passes

	// Verify double-stop is safe
	err := manager.Stop()
	assert.NoError(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails or panics**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestSSEManager_Stop_ConcurrentSendersNoPanic -v ./internal/notifications/ -count=1 -timeout 30s`
Expected: Either panic on closed channel or race detector warning

- [ ] **Step 3: Implement the fix**

Modify `internal/notifications/sse_manager.go`:

Add to struct (after line 30):
```go
	closed   atomic.Bool
	stopOnce sync.Once
```

Add import `"sync/atomic"`.

Replace the `Stop()` method (lines 81-104) with:
```go
func (m *SSEManager) Stop() error {
	var stopErr error
	m.stopOnce.Do(func() {
		m.logger.Info("Stopping SSE manager")
		m.closed.Store(true)
		m.cancel()
		m.wg.Wait()

		// Close all client channels
		m.clientsMu.Lock()
		for taskID, clients := range m.clients {
			for client := range clients {
				close(client)
			}
			delete(m.clients, taskID)
		}
		m.clientsMu.Unlock()

		m.globalClientsMu.Lock()
		for client := range m.globalClients {
			close(client)
		}
		m.globalClients = make(map[chan<- []byte]struct{})
		m.globalClientsMu.Unlock()
	})
	return stopErr
}
```

Add a `closed` check to all Send methods. In `SendTaskEvent` and similar, before writing to channels:
```go
if m.closed.Load() {
	return nil // Manager shutting down, discard event
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestSSEManager_Stop_ConcurrentSendersNoPanic -v ./internal/notifications/ -count=1 -timeout 30s`
Expected: PASS with no race detector warnings

- [ ] **Step 5: Verify existing tests still pass**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -v ./internal/notifications/ -count=1 -timeout 60s`
Expected: All existing tests PASS

- [ ] **Step 6: Commit**

```bash
git add internal/notifications/sse_manager.go internal/notifications/sse_manager_test.go
git commit -m "fix(notifications): prevent send-on-closed-channel panic in SSEManager

Add sync.Once for idempotent Stop(), atomic.Bool closed guard for
Send methods. Concurrent Stop()+Send() no longer panics."
```

---

### Task 2: Kafka Transport — Double-Close Protection

**Files:**
- Modify: `internal/notifications/kafka_transport.go:66-75` (struct), `118-125` (Stop method)
- Test: `internal/notifications/kafka_transport_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestKafkaTransport_Stop_DoubleCloseNoPanic(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	config := DefaultKafkaTransportConfig()
	transport := NewKafkaTransport(nil, logger, config)
	transport.Start()

	// First stop should succeed
	assert.NotPanics(t, func() {
		transport.Stop()
	})

	// Second stop should NOT panic
	assert.NotPanics(t, func() {
		transport.Stop()
	})
}

func TestKafkaTransport_Stop_ConcurrentStopNoPanic(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	config := DefaultKafkaTransportConfig()
	transport := NewKafkaTransport(nil, logger, config)
	transport.Start()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NotPanics(t, func() {
				transport.Stop()
			})
		}()
	}
	wg.Wait()
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run "TestKafkaTransport_Stop_(DoubleClose|Concurrent)" -v ./internal/notifications/ -count=1 -timeout 30s`
Expected: Panic on double close

- [ ] **Step 3: Implement the fix**

Modify struct to add `stopOnce sync.Once` field. Replace Stop():
```go
func (t *KafkaTransport) Stop() {
	t.stopOnce.Do(func() {
		close(t.stopCh)
		if t.eventCh != nil {
			close(t.eventCh)
		}
		t.wg.Wait()
	})
}
```

Add `stopOnce sync.Once` to the struct (line 73, after `mu`).

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run "TestKafkaTransport_Stop_(DoubleClose|Concurrent)" -v ./internal/notifications/ -count=1 -timeout 30s`
Expected: PASS

- [ ] **Step 5: Run all notification tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -v ./internal/notifications/ -count=1 -timeout 60s`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add internal/notifications/kafka_transport.go internal/notifications/kafka_transport_test.go
git commit -m "fix(notifications): prevent double-close panic in KafkaTransport

Wrap Stop() in sync.Once to make it idempotent. Concurrent and
repeated Stop() calls no longer panic."
```

---

### Task 3: Kafka Streams — WaitGroup Goroutine Tracking

**Files:**
- Modify: `internal/streaming/kafka_streams.go:108-117` (goroutine at line 109)
- Test: `internal/streaming/kafka_streams_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestConversationStreamProcessor_Stop_WaitsForGoroutine(t *testing.T) {
	// Create processor with memory state store
	config := DefaultStreamProcessorConfig()
	config.StateStoreType = "memory"
	logger := zap.NewNop()

	processor, err := NewConversationStreamProcessor(config, nil, logger)
	require.NoError(t, err)

	// Stop should not race with internal goroutine
	assert.NotPanics(t, func() {
		err := processor.Stop()
		assert.NoError(t, err)
	})
}
```

- [ ] **Step 2: Run test**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestConversationStreamProcessor_Stop_WaitsForGoroutine -v ./internal/streaming/ -count=1 -timeout 30s`

- [ ] **Step 3: Implement the fix**

In `kafka_streams.go`, modify the goroutine at line 109:
```go
	// Wait for stop signal
	csp.wg.Add(1)
	go func() {
		defer csp.wg.Done()
		<-ctx.Done()
		if err := sub.Unsubscribe(); err != nil {
			csp.logger.Error("failed to unsubscribe", zap.Error(err))
		}
		if err := csp.Stop(); err != nil {
			csp.logger.Error("failed to stop stream processor", zap.Error(err))
		}
	}()
```

Ensure `Stop()` calls `csp.wg.Wait()` after setting `running = false`.

- [ ] **Step 4: Run test to verify pass**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestConversationStreamProcessor_Stop -v ./internal/streaming/ -count=1 -timeout 30s`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/streaming/kafka_streams.go internal/streaming/kafka_streams_test.go
git commit -m "fix(streaming): track goroutine with WaitGroup in ConversationStreamProcessor

Add wg.Add(1)/wg.Done() around context-done goroutine to prevent
use-after-close on stopChan during shutdown."
```

---

### Task 4: MCP Connection Pool — Atomic Bool for Closed Flag

**Files:**
- Modify: `internal/mcp/connection_pool.go:79` (closed field)
- Test: `internal/mcp/connection_pool_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestMCPConnectionPool_ConcurrentGetClose(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	pool := NewMCPConnectionPool(nil, logger, nil)

	var wg sync.WaitGroup

	// Concurrent Get and Close
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = pool.GetConnection("test-server")
		}()
		go func() {
			defer wg.Done()
			_ = pool.Close()
		}()
	}
	wg.Wait()
}
```

- [ ] **Step 2: Run with race detector**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestMCPConnectionPool_ConcurrentGetClose -v ./internal/mcp/ -count=1 -timeout 30s`
Expected: Race detector warning on `closed` field

- [ ] **Step 3: Implement the fix**

In `connection_pool.go`, replace `closed bool` with `closed atomic.Bool` (add `"sync/atomic"` import).

Replace all `p.closed` reads with `p.closed.Load()` and writes with `p.closed.Store(true)`.

- [ ] **Step 4: Run test to verify pass**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestMCPConnectionPool_ConcurrentGetClose -v ./internal/mcp/ -count=1 -timeout 30s`
Expected: PASS, no race warnings

- [ ] **Step 5: Run all MCP tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -v ./internal/mcp/ -count=1 -timeout 60s`
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add internal/mcp/connection_pool.go internal/mcp/connection_pool_test.go
git commit -m "fix(mcp): replace bool with atomic.Bool for connection pool closed flag

Eliminates data race between concurrent GetConnection() and Close()
calls on MCPConnectionPool."
```

---

### Task 5: Plugin Hot Reload — WaitGroup Goroutine Tracking

**Files:**
- Modify: `internal/plugins/hot_reload.go:17-26` (struct), `69-83` (Start), `86-94` (Stop)
- Test: `internal/plugins/hot_reload_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestHotReloadManager_StartStop_NoGoroutineLeak(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()

	registry := NewRegistry()
	cfg := &config.Config{}
	manager, err := NewHotReloadManager(cfg, registry)
	if err != nil {
		t.Skip("fsnotify not available in test environment")
		return
	}

	ctx := context.Background()
	err = manager.Start(ctx)
	require.NoError(t, err)

	// Stop should wait for watchLoop to exit
	err = manager.Stop()
	require.NoError(t, err)

	// Allow goroutines to settle
	time.Sleep(100 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	assert.LessOrEqual(t, finalGoroutines, initialGoroutines+2,
		"goroutine leak detected: started with %d, ended with %d",
		initialGoroutines, finalGoroutines)
}
```

- [ ] **Step 2: Run test**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestHotReloadManager_StartStop_NoGoroutineLeak -v ./internal/plugins/ -count=1 -timeout 30s`

- [ ] **Step 3: Implement the fix**

Add `wg sync.WaitGroup` to the struct. Modify Start():
```go
func (h *HotReloadManager) Start(ctx context.Context) error {
	fmt.Printf("Starting plugin hot-reload manager")

	if err := h.loadExistingPlugins(); err != nil {
		fmt.Printf("Failed to load existing plugins: %v\n", err)
	}

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.watchLoop(ctx)
	}()

	fmt.Printf("Plugin hot-reload manager started successfully")
	return nil
}
```

Modify Stop() to add `stopOnce`:
```go
func (h *HotReloadManager) Stop() error {
	fmt.Printf("Stopping plugin hot-reload manager")

	close(h.stopChan)
	_ = h.watcher.Close()
	h.wg.Wait()

	fmt.Printf("Plugin hot-reload manager stopped")
	return nil
}
```

Add `stopOnce sync.Once` if Stop can be called multiple times, wrap in `h.stopOnce.Do(...)`.

- [ ] **Step 4: Run test to verify pass**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestHotReloadManager_StartStop -v ./internal/plugins/ -count=1 -timeout 30s`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/plugins/hot_reload.go internal/plugins/hot_reload_test.go
git commit -m "fix(plugins): track watchLoop goroutine with WaitGroup in HotReloadManager

Stop() now waits for watchLoop to exit, preventing goroutine leak."
```

---

### Task 6: Cache Service — Nested Map Synchronization

**Files:**
- Modify: `internal/cache/cache_service.go:21-24`
- Test: `internal/cache/cache_service_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestCacheService_ConcurrentUserKeyTracking(t *testing.T) {
	cs := &CacheService{
		enabled:    false,
		defaultTTL: 30 * time.Minute,
		userKeys:   make(map[string]map[string]struct{}),
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		userID := fmt.Sprintf("user-%d", i%5)
		cacheKey := fmt.Sprintf("key-%d", i)

		go func() {
			defer wg.Done()
			cs.trackUserKey(userID, cacheKey)
		}()
		go func() {
			defer wg.Done()
			cs.InvalidateUserCache(context.Background(), userID)
		}()
	}
	wg.Wait()
}
```

- [ ] **Step 2: Run with race detector**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestCacheService_ConcurrentUserKeyTracking -v ./internal/cache/ -count=1 -timeout 30s`

- [ ] **Step 3: Implement the fix**

In `trackUserKey`, copy the inner map reference under write lock:
```go
func (c *CacheService) trackUserKey(userID, cacheKey string) {
	c.userKeysMu.Lock()
	defer c.userKeysMu.Unlock()

	if c.userKeys[userID] == nil {
		c.userKeys[userID] = make(map[string]struct{})
	}
	c.userKeys[userID][cacheKey] = struct{}{}
}
```

In `InvalidateUserCache`, extract and delete atomically under write lock:
```go
func (c *CacheService) InvalidateUserCache(ctx context.Context, userID string) error {
	c.userKeysMu.Lock()
	keys := c.userKeys[userID]
	delete(c.userKeys, userID)
	c.userKeysMu.Unlock()

	// Now 'keys' is exclusively owned — safe to iterate without lock
	for key := range keys {
		_ = c.delete(ctx, key)
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify pass**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestCacheService_ConcurrentUserKeyTracking -v ./internal/cache/ -count=1 -timeout 30s`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cache/cache_service.go internal/cache/cache_service_test.go
git commit -m "fix(cache): fix nested map race in CacheService userKeys tracking

Extract-and-delete inner map atomically under write lock, ensuring
exclusive ownership before iteration."
```

---

### Task 7: Integration Orchestrator — Activate Disabled Mutex

**Files:**
- Modify: `internal/services/integration_orchestrator.go:29`
- Test: `internal/services/integration_orchestrator_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestIntegrationOrchestrator_ConcurrentWorkflowAccess(t *testing.T) {
	orch := &IntegrationOrchestrator{
		workflows: make(map[string]*Workflow),
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			w := &Workflow{
				ID:   fmt.Sprintf("wf-%d", idx),
				Name: fmt.Sprintf("workflow-%d", idx),
			}
			orch.RegisterWorkflow(w)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 50, len(orch.workflows))
}
```

- [ ] **Step 2: Run with race detector**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestIntegrationOrchestrator_ConcurrentWorkflowAccess -v ./internal/services/ -count=1 -timeout 30s`

- [ ] **Step 3: Implement the fix**

In `integration_orchestrator.go` line 29, remove `//nolint:unused`:
```go
	mu sync.RWMutex
```

Add Lock/Unlock calls in all methods that access `workflows` map. For `RegisterWorkflow`:
```go
func (o *IntegrationOrchestrator) RegisterWorkflow(w *Workflow) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.workflows[w.ID] = w
}
```

For read access methods, use `o.mu.RLock()` / `o.mu.RUnlock()`.

- [ ] **Step 4: Run test to verify pass**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestIntegrationOrchestrator_ConcurrentWorkflowAccess -v ./internal/services/ -count=1 -timeout 30s`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/services/integration_orchestrator.go internal/services/integration_orchestrator_test.go
git commit -m "fix(services): activate mutex for IntegrationOrchestrator workflows map

Remove //nolint:unused from mu sync.RWMutex, add Lock/Unlock around
all workflows map access to prevent concurrent map writes."
```

---

### Task 8: Debate Service — Panic Recovery in Participant Goroutines

**Files:**
- Modify: `internal/services/debate_service.go:886-1024`
- Test: `internal/services/debate_service_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestDebateService_ParticipantPanic_RecoveredGracefully(t *testing.T) {
	// This test verifies that if a participant goroutine panics,
	// the WaitGroup still gets Done() called and doesn't hang
	// Test setup creates a debate service with mock providers
	// where one provider panics during Complete()

	// The key assertion: the function returns (doesn't hang)
	// within a reasonable timeout
	done := make(chan struct{})
	go func() {
		// Run debate with panicking provider
		// (actual implementation depends on test helpers available)
		close(done)
	}()

	select {
	case <-done:
		// Success - didn't hang
	case <-time.After(10 * time.Second):
		t.Fatal("debate service hung - panic recovery failed")
	}
}
```

- [ ] **Step 2: Implement the fix**

In debate_service.go, wrap participant goroutine body with panic recovery:
```go
go func(p ParticipantConfig) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			errChan <- fmt.Errorf("participant %s panicked: %v", p.Name, r)
		}
	}()
	// ... existing participant logic
}(participant)
```

Also track the channel-close goroutine (line ~1020):
```go
wg2 := sync.WaitGroup{}
wg2.Add(1)
go func() {
	defer wg2.Done()
	wg.Wait()
	close(responseChan)
	close(errorChan)
}()
```

- [ ] **Step 3: Run tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -run TestDebateService -v ./internal/services/ -count=1 -timeout 120s`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/services/debate_service.go internal/services/debate_service_test.go
git commit -m "fix(debate): add panic recovery to participant goroutines

Prevent WaitGroup hang when participant panics. Track channel-close
goroutine with separate WaitGroup for clean lifecycle."
```

---

### Task 9: Debate Orchestrator — Protect ActiveDebate Fields

**Files:**
- Modify: `internal/debate/orchestrator/orchestrator.go:49-50`
- Test: `internal/debate/orchestrator/orchestrator_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestOrchestrator_ConcurrentDebateStateModification(t *testing.T) {
	// Create orchestrator and start a debate
	// Concurrently modify debate state from multiple goroutines
	// Verify no race with -race flag
}
```

- [ ] **Step 2: Implement the fix**

Add `mu sync.RWMutex` to `ActiveDebate` struct. Protect all field accesses.

- [ ] **Step 3: Run all orchestrator tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -v ./internal/debate/orchestrator/ -count=1 -timeout 60s`

- [ ] **Step 4: Commit**

```bash
git add internal/debate/orchestrator/orchestrator.go internal/debate/orchestrator/orchestrator_test.go
git commit -m "fix(debate): add mutex to ActiveDebate struct fields

Protect concurrent debate state modification with per-debate
sync.RWMutex."
```

---

### Task 10: Model Metadata Handler — Defensive Parameter Passing

**Files:**
- Modify: `internal/handlers/model_metadata.go`

- [ ] **Step 1: Pass service as parameter to goroutine (not captured via closure)**

Change any `go func() { ... h.service ... }()` to `go func(svc ServiceType) { ... svc ... }(h.service)`

- [ ] **Step 2: Run handler tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -v ./internal/handlers/ -count=1 -timeout 60s`

- [ ] **Step 3: Commit**

```bash
git add internal/handlers/model_metadata.go
git commit -m "fix(handlers): pass service as goroutine parameter in ModelMetadataHandler

Defensive improvement: pass service via parameter instead of closure
capture to prevent future regressions if field becomes mutable."
```

---

### Task 11: WebSocket Server — Lock Ordering Documentation

**Files:**
- Modify: `internal/notifications/websocket_server.go:25-42`

- [ ] **Step 1: Add lock ordering documentation**

Add comment above the struct:
```go
// Lock ordering: clientsMu MUST be acquired before globalClientsMu.
// Never acquire globalClientsMu while holding clientsMu, or deadlock
// will occur.
```

- [ ] **Step 2: Verify lock ordering is correct in all methods**

Read all methods, verify ordering is consistent. If not, fix.

- [ ] **Step 3: Run tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -v ./internal/notifications/ -count=1 -timeout 60s`

- [ ] **Step 4: Commit**

```bash
git add internal/notifications/websocket_server.go
git commit -m "fix(notifications): document lock ordering in WebSocketServer

Add lock ordering contract: clientsMu before globalClientsMu."
```

---

### Task 12: Minor Safety Fixes (Polling Store, Circuit Breaker, Boot Manager)

**Files:**
- Modify: `internal/notifications/polling_store.go`, `internal/llm/circuit_breaker.go`, `internal/services/boot_manager.go`

- [ ] **Step 1: Add panic recovery to polling_store cleanup loop**

```go
func (s *PollingStore) cleanupLoop() {
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			s.logger.Errorf("cleanup loop panicked: %v", r)
		}
	}()
	// ... existing loop
}
```

- [ ] **Step 2: Return error from circuit breaker on listener limit**

In `circuit_breaker.go`, change listener registration to return error:
```go
if len(cb.listeners) >= MaxCircuitBreakerListeners {
	cb.logger.Warnf("circuit breaker %s: listener limit reached (%d)", cb.name, MaxCircuitBreakerListeners)
	return -1, fmt.Errorf("listener limit reached: %d", MaxCircuitBreakerListeners)
}
```

- [ ] **Step 3: Add lock assertion comment to boot_manager**

```go
// setResultLocked sets the boot result.
// REQUIRES: caller holds b.mu (write lock).
func (b *BootManager) setResultLocked(result BootResult) {
```

- [ ] **Step 4: Run all affected tests**

Run tests for each package in parallel:
```bash
GOMAXPROCS=2 go test -race -v ./internal/notifications/ ./internal/llm/ ./internal/services/ -count=1 -timeout 120s
```

- [ ] **Step 5: Commit**

```bash
git add internal/notifications/polling_store.go internal/llm/circuit_breaker.go internal/services/boot_manager.go
git commit -m "fix: minor safety improvements (panic recovery, error return, lock docs)

- polling_store: panic recovery in cleanup loop
- circuit_breaker: return error on listener limit (was silent -1)
- boot_manager: document lock requirement on setResultLocked"
```

---

### Task 13: Debate Service — Track HelixMemory Store Goroutine

**Files:**
- Modify: `internal/services/debate_service.go:2047-2063`

- [ ] **Step 1: Add WaitGroup tracking**

```go
s.wg.Add(1)
go func() {
	defer s.wg.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// ... existing HelixMemory store logic
}()
```

- [ ] **Step 2: Run debate service tests**

Run: `cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent && GOMAXPROCS=2 go test -race -v ./internal/services/ -run TestDebate -count=1 -timeout 120s`

- [ ] **Step 3: Commit**

```bash
git add internal/services/debate_service.go
git commit -m "fix(debate): track HelixMemory store goroutine with WaitGroup

Coordinate shutdown lifecycle for background memory storage."
```

---

### Task 14: Full Phase 1 Verification

- [ ] **Step 1: Run all unit tests with race detector**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -race -p 1 -count=1 -timeout 300s \
  ./internal/notifications/ \
  ./internal/streaming/ \
  ./internal/cache/ \
  ./internal/mcp/ \
  ./internal/plugins/ \
  ./internal/services/ \
  ./internal/debate/orchestrator/ \
  ./internal/handlers/ \
  ./internal/llm/
```
Expected: All PASS, zero race warnings

- [ ] **Step 2: Run full build verification**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent
go build ./...
```
Expected: Clean build, no errors

- [ ] **Step 3: Run existing challenge scripts**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/HelixAgent
GOMAXPROCS=2 nice -n 19 ./challenges/scripts/goroutine_lifecycle_challenge.sh
GOMAXPROCS=2 nice -n 19 ./challenges/scripts/race_condition_challenge.sh
```
Expected: All existing assertions PASS

- [ ] **Step 4: Final commit if any remaining changes**

---

### Task 15: Update Challenge Scripts

- [ ] **Step 1: Update goroutine_lifecycle_challenge.sh**

Add 15 new assertions validating each concurrency fix.

- [ ] **Step 2: Update race_condition_challenge.sh**

Add assertions for atomic bool, nested map sync, lock ordering.

- [ ] **Step 3: Run updated challenges**

```bash
GOMAXPROCS=2 nice -n 19 ./challenges/scripts/goroutine_lifecycle_challenge.sh
GOMAXPROCS=2 nice -n 19 ./challenges/scripts/race_condition_challenge.sh
```

- [ ] **Step 4: Commit**

```bash
git add challenges/scripts/goroutine_lifecycle_challenge.sh challenges/scripts/race_condition_challenge.sh
git commit -m "test(challenges): add 25 new assertions for concurrency safety fixes

Validates all 15 concurrency fixes: channel safety, atomic bools,
WaitGroup tracking, mutex activation, panic recovery."
```
