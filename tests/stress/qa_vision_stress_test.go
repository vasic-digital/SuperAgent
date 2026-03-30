//go:build !short

package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	helixqaadapter "dev.helix.agent/internal/adapters/helixqa"
	"digital.vasic.visionengine/pkg/remote"
)

// TestVisionPool_ConcurrentSlotAccess validates thread safety of
// VisionPool under concurrent GetSlot/AssignSlots/Shutdown pressure.
func TestVisionPool_ConcurrentSlotAccess(t *testing.T) {
	pool := remote.NewVisionPool(remote.PoolConfig{
		Host:     "test.local",
		BasePort: 9000,
	})

	targets := make([]remote.SlotTarget, 20)
	for i := 0; i < 20; i++ {
		targets[i] = remote.SlotTarget{
			Platform: "android",
			Device:   fmt.Sprintf("device-%d", i),
		}
	}
	pool.AssignSlots(targets)

	var wg sync.WaitGroup

	// 100 concurrent GetSlot calls
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			device := fmt.Sprintf("device-%d", idx%20)
			slot := pool.GetSlot("android", device)
			if slot != nil {
				slot.Lock()
				slot.RecordCall(time.Millisecond, nil)
				slot.Unlock()
			}
		}(i)
	}

	// Concurrent Size() calls
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pool.Size()
		}()
	}

	wg.Wait()
	assert.Equal(t, 20, pool.Size())

	pool.Shutdown(context.Background())
	assert.Equal(t, 0, pool.Size())
}

// TestVisionPool_SlotRecordCallStress validates VisionSlot
// metrics under high-frequency concurrent recording.
func TestVisionPool_SlotRecordCallStress(t *testing.T) {
	pool := remote.NewVisionPool(remote.PoolConfig{
		Host:     "test.local",
		BasePort: 9000,
	})
	pool.AssignSlots([]remote.SlotTarget{{Platform: "web"}})

	slot := pool.GetSlot("web", "")
	require.NotNil(t, slot)

	var wg sync.WaitGroup
	const numGoroutines = 200

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			slot.Lock()
			if idx%3 == 0 {
				slot.RecordCall(time.Millisecond, fmt.Errorf("test error"))
			} else {
				slot.RecordCall(time.Millisecond, nil)
			}
			slot.Unlock()
		}(i)
	}

	wg.Wait()

	calls, totalTime, errors := slot.Stats()
	assert.Equal(t, numGoroutines, calls)
	assert.Equal(t, time.Duration(numGoroutines)*time.Millisecond, totalTime)
	// ~1/3 should be errors
	assert.True(t, errors > 0)
}

// TestHelixQAAdapter_ConcurrentInit validates thread-safe
// initialization of the HelixQA adapter.
func TestHelixQAAdapter_ConcurrentInit(t *testing.T) {
	adapter := helixqaadapter.New(nil)

	var wg sync.WaitGroup
	dbPath := t.TempDir() + "/stress.db"

	// 50 concurrent Initialize calls
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = adapter.Initialize(dbPath)
		}()
	}

	wg.Wait()

	// Should still work after concurrent init
	platforms := adapter.SupportedPlatforms()
	assert.NotEmpty(t, platforms)

	assert.NoError(t, adapter.Close())
}

// TestHelixQAAdapter_ConcurrentGetFindings validates thread-safe
// findings retrieval under concurrent access.
func TestHelixQAAdapter_ConcurrentGetFindings(t *testing.T) {
	adapter := helixqaadapter.New(nil)
	dbPath := t.TempDir() + "/stress-findings.db"
	require.NoError(t, adapter.Initialize(dbPath))
	defer adapter.Close()

	var wg sync.WaitGroup

	// 100 concurrent GetFindings calls
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			findings, err := adapter.GetFindings("open")
			assert.NoError(t, err)
			_ = findings
		}()
	}

	wg.Wait()
}
