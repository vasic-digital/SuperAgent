package profiling

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestMemoryLeakDetector_DefaultThresholds(t *testing.T) {
	thresholds := DefaultMemoryThresholds()

	if thresholds.MaxAllocMB == 0 {
		t.Error("MaxAllocMB should not be zero")
	}
	if thresholds.MaxSysMB == 0 {
		t.Error("MaxSysMB should not be zero")
	}
	if thresholds.MaxGoroutines == 0 {
		t.Error("MaxGoroutines should not be zero")
	}
	if thresholds.CheckInterval == 0 {
		t.Error("CheckInterval should not be zero")
	}
}

func TestMemoryLeakDetector_New(t *testing.T) {
	thresholds := DefaultMemoryThresholds()
	detector := NewMemoryLeakDetector(thresholds)

	if detector == nil {
		t.Fatal("detector should not be nil")
	}
	if len(detector.snapshots) != 0 {
		t.Error("snapshots should be empty initially")
	}
}

func TestMemoryLeakDetector_GetSnapshot(t *testing.T) {
	detector := NewMemoryLeakDetector(DefaultMemoryThresholds())

	snapshot := detector.GetSnapshot()

	if snapshot.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
	if snapshot.Alloc == 0 {
		t.Error("alloc should not be zero")
	}
	if snapshot.NumGoroutine == 0 {
		t.Error("NumGoroutine should not be zero")
	}
}

func TestMemoryLeakDetector_GetHistory(t *testing.T) {
	detector := NewMemoryLeakDetector(DefaultMemoryThresholds())

	detector.check()
	detector.check()
	detector.check()

	history := detector.GetHistory()

	if len(history) != 3 {
		t.Errorf("expected 3 snapshots, got %d", len(history))
	}
}

func TestMemoryLeakDetector_DetectPotentialLeaks(t *testing.T) {
	detector := NewMemoryLeakDetector(DefaultMemoryThresholds())

	issues := detector.DetectPotentialLeaks()

	if len(issues) > 0 && len(detector.snapshots) < 5 {
		t.Log("Potential leak detection requires at least 5 samples")
	}
}

func TestMemoryLeakDetector_Alerts(t *testing.T) {
	thresholds := MemoryThresholds{
		MaxAllocMB:         1,
		MaxSysMB:           1,
		MaxGoroutines:      1,
		CheckInterval:      time.Millisecond * 100,
		GrowthRateMBPerSec: 0.001,
	}

	detector := NewMemoryLeakDetector(thresholds)

	alerts := detector.Alerts()
	if alerts == nil {
		t.Error("alerts channel should not be nil")
	}
}

func TestMemoryLeakDetector_ForceGC(t *testing.T) {
	detector := NewMemoryLeakDetector(DefaultMemoryThresholds())

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	detector.ForceGC()

	runtime.ReadMemStats(&m2)

	if m2.NumGC <= m1.NumGC {
		t.Log("GC was triggered")
	}
}

func TestMemoryLeakDetector_StartStop(t *testing.T) {
	thresholds := MemoryThresholds{
		MaxAllocMB:         1024,
		MaxSysMB:           2048,
		MaxGoroutines:      10000,
		CheckInterval:      time.Second * 10,
		GrowthRateMBPerSec: 10.0,
	}

	detector := NewMemoryLeakDetector(thresholds)
	ctx := context.Background()

	err := detector.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start detector: %v", err)
	}

	err = detector.Start(ctx)
	if err == nil {
		t.Error("should return error when already running")
	}

	time.Sleep(100 * time.Millisecond)

	detector.Stop()
	detector.Stop()
}

func TestLazyLoader_Register(t *testing.T) {
	loader := NewLazyLoader()

	loader.Register("test", func() (interface{}, error) {
		return "value", nil
	})

	if !contains(loader.GetRegistered(), "test") {
		t.Error("test should be registered")
	}
}

func TestLazyLoader_Get(t *testing.T) {
	loader := NewLazyLoader()

	callCount := 0
	loader.Register("test", func() (interface{}, error) {
		callCount++
		return "value", nil
	})

	val1, err := loader.Get("test")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if val1 != "value" {
		t.Errorf("expected 'value', got %v", val1)
	}

	val2, _ := loader.Get("test")
	if callCount != 1 {
		t.Errorf("factory should be called once, called %d times", callCount)
	}
	if val1 != val2 {
		t.Error("should return same instance")
	}
}

func TestLazyLoader_IsLoaded(t *testing.T) {
	loader := NewLazyLoader()

	loader.Register("test", func() (interface{}, error) {
		return "value", nil
	})

	if loader.IsLoaded("test") {
		t.Error("should not be loaded initially")
	}

	loader.Get("test")

	if !loader.IsLoaded("test") {
		t.Error("should be loaded after Get")
	}
}

func TestLazyLoader_Unload(t *testing.T) {
	loader := NewLazyLoader()

	loader.Register("test", func() (interface{}, error) {
		return "value", nil
	})

	loader.Get("test")

	if !loader.IsLoaded("test") {
		t.Error("should be loaded")
	}

	loader.Unload("test")

	if loader.IsLoaded("test") {
		t.Error("should not be loaded after Unload")
	}
}

func TestLazyLoader_Preload(t *testing.T) {
	loader := NewLazyLoader()

	callCount := 0
	loader.Register("a", func() (interface{}, error) {
		callCount++
		return "a", nil
	})
	loader.Register("b", func() (interface{}, error) {
		callCount++
		return "b", nil
	})

	err := loader.Preload("a", "b")
	if err != nil {
		t.Fatalf("preload failed: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}

	if !loader.IsLoaded("a") || !loader.IsLoaded("b") {
		t.Error("both should be loaded")
	}
}

func TestLazyLoader_Clear(t *testing.T) {
	loader := NewLazyLoader()

	loader.Register("test", func() (interface{}, error) {
		return "value", nil
	})

	loader.Get("test")
	loader.Clear()

	if len(loader.GetLoaded()) != 0 {
		t.Error("should have no loaded items after Clear")
	}
}

func TestLazyLoader_Concurrent(t *testing.T) {
	loader := NewLazyLoader()

	callCount := 0
	var mu sync.Mutex

	loader.Register("test", func() (interface{}, error) {
		mu.Lock()
		callCount++
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return "value", nil
	})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := loader.Get("test")
			if err != nil {
				t.Errorf("failed to get: %v", err)
			}
		}()
	}

	wg.Wait()

	mu.Lock()
	if callCount != 1 {
		t.Errorf("factory should be called once, called %d times", callCount)
	}
	mu.Unlock()
}

func TestDeadlockDetector_New(t *testing.T) {
	detector := NewDeadlockDetector(30 * time.Second)

	if detector == nil {
		t.Fatal("detector should not be nil")
	}
	if detector.timeout != 30*time.Second {
		t.Error("timeout should be set")
	}
}

func TestDeadlockDetector_RecordLockAcquire(t *testing.T) {
	detector := NewDeadlockDetector(30 * time.Second)

	release := detector.RecordLockAcquire("test-lock")

	info := detector.GetLockInfo()
	if len(info) == 0 {
		t.Error("should have recorded lock")
	}

	release()

	info = detector.GetLockInfo()
	if len(info) != 0 {
		t.Error("should have released lock")
	}
}

func TestDeadlockDetector_Alerts(t *testing.T) {
	detector := NewDeadlockDetector(30 * time.Second)

	alerts := detector.Alerts()
	if alerts == nil {
		t.Error("alerts channel should not be nil")
	}
}

func TestDeadlockDetector_GetGoroutineCount(t *testing.T) {
	detector := NewDeadlockDetector(30 * time.Second)

	count := detector.GetGoroutineCount()
	if count < 1 {
		t.Error("should have at least 1 goroutine")
	}
}

func TestDeadlockDetector_StartStop(t *testing.T) {
	detector := NewDeadlockDetector(time.Second)
	ctx := context.Background()

	err := detector.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start detector: %v", err)
	}

	err = detector.Start(ctx)
	if err == nil {
		t.Error("should return error when already running")
	}

	time.Sleep(100 * time.Millisecond)

	detector.Stop()
	detector.Stop()
}

func TestDeadlockDetector_WrapMutex(t *testing.T) {
	detector := NewDeadlockDetector(30 * time.Second)

	var mu sync.Mutex
	wrapped := detector.WrapMutex(&mu, "wrapped-mutex")

	wrapped.Lock()

	info := detector.GetLockInfo()
	if len(info) == 0 {
		t.Error("should have recorded lock")
	}

	wrapped.Unlock()

	info = detector.GetLockInfo()
	if len(info) != 0 {
		t.Error("should have released lock after Unlock")
	}
}

func TestDeadlockDetector_RecordLockOrder(t *testing.T) {
	detector := NewDeadlockDetector(30 * time.Second)

	detector.RecordLockOrder("a", "b")
	detector.RecordLockOrder("b", "c")

	if len(detector.order) == 0 {
		t.Error("should have recorded lock order")
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
