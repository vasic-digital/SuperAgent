//go:build performance

package performance

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	helixqaadapter "dev.helix.agent/internal/adapters/helixqa"
	"digital.vasic.visionengine/pkg/remote"
)

func TestVisionSlot_MetricsAccumulation(t *testing.T) {
	pool := remote.NewVisionPool(remote.PoolConfig{
		Host:     "metrics.local",
		BasePort: 9000,
	})
	pool.AssignSlots([]remote.SlotTarget{{Platform: "android", Device: "dev1"}})

	slot := pool.GetSlot("android", "dev1")
	require.NotNil(t, slot)

	// Record 10 successful calls
	for i := 0; i < 10; i++ {
		slot.Lock()
		slot.RecordCall(50*time.Millisecond, nil)
		slot.Unlock()
	}

	// Record 3 failed calls
	for i := 0; i < 3; i++ {
		slot.Lock()
		slot.RecordCall(100*time.Millisecond, assert.AnError)
		slot.Unlock()
	}

	calls, totalTime, errors := slot.Stats()
	assert.Equal(t, 13, calls, "should track all calls")
	assert.Equal(t, 800*time.Millisecond, totalTime, "should accumulate total time")
	assert.Equal(t, 3, errors, "should track error count")
}

func TestVisionPool_MultiSlotMetrics(t *testing.T) {
	pool := remote.NewVisionPool(remote.PoolConfig{
		Host:     "metrics.local",
		BasePort: 9000,
	})
	pool.AssignSlots([]remote.SlotTarget{
		{Platform: "android", Device: "dev1"},
		{Platform: "android", Device: "dev2"},
		{Platform: "web"},
	})

	// Exercise each slot differently
	for _, target := range []struct {
		platform string
		device   string
		calls    int
	}{
		{"android", "dev1", 5},
		{"android", "dev2", 10},
		{"web", "", 3},
	} {
		slot := pool.GetSlot(target.platform, target.device)
		require.NotNil(t, slot)
		for i := 0; i < target.calls; i++ {
			slot.Lock()
			slot.RecordCall(time.Millisecond, nil)
			slot.Unlock()
		}
	}

	// Verify per-slot metrics
	s1 := pool.GetSlot("android", "dev1")
	calls1, _, _ := s1.Stats()
	assert.Equal(t, 5, calls1)

	s2 := pool.GetSlot("android", "dev2")
	calls2, _, _ := s2.Stats()
	assert.Equal(t, 10, calls2)

	s3 := pool.GetSlot("web", "")
	calls3, _, _ := s3.Stats()
	assert.Equal(t, 3, calls3)
}

func TestHelixQAAdapter_OperationLatency(t *testing.T) {
	adapter := helixqaadapter.New(nil)
	dbPath := filepath.Join(t.TempDir(), "metrics.db")
	require.NoError(t, adapter.Initialize(dbPath))
	defer adapter.Close()

	// Measure GetFindings latency
	start := time.Now()
	for i := 0; i < 100; i++ {
		_, _ = adapter.GetFindings("open")
	}
	elapsed := time.Since(start)

	// 100 queries should complete in under 5s
	assert.Less(t, elapsed, 5*time.Second,
		"100 GetFindings queries should complete quickly")

	// Average should be under 50ms
	avgMs := elapsed.Milliseconds() / 100
	assert.Less(t, avgMs, int64(50),
		"average GetFindings latency should be under 50ms")
}

func TestHelixQAAdapter_InitializeLatency(t *testing.T) {
	start := time.Now()
	adapter := helixqaadapter.New(nil)
	dbPath := filepath.Join(t.TempDir(), "init-latency.db")
	require.NoError(t, adapter.Initialize(dbPath))
	elapsed := time.Since(start)
	defer adapter.Close()

	// Initialize should complete in under 500ms
	assert.Less(t, elapsed, 500*time.Millisecond,
		"adapter initialization should be fast")
}
