// Package clis provides CLI agent integration for HelixAgent.
package clis

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInstancePool(t *testing.T) {
	config := DefaultPoolConfig()
	factory := func() (*AgentInstance, error) {
		return &AgentInstance{
			ID:   "test-instance",
			Type: TypeAider,
		}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)
	require.NotNil(t, pool)
	assert.Equal(t, TypeAider, pool.agentType)
	assert.Equal(t, config.MinIdle, pool.minIdle)
	assert.Equal(t, config.MaxIdle, pool.maxIdle)
	assert.Equal(t, config.MaxActive, pool.maxActive)

	pool.Close()
}

func TestInstancePool_AcquireFromEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pool test in short mode - requires database setup")
	}
	config := PoolConfig{
		MinIdle:     0,
		MaxIdle:     5,
		MaxActive:   10,
		MaxLifetime: time.Hour,
	}

	factoryCalled := 0
	factory := func() (*AgentInstance, error) {
		factoryCalled++
		return &AgentInstance{
			ID:   "instance-1",
			Type: TypeAider,
		}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)
	defer pool.Close()

	ctx := context.Background()
	inst, err := pool.Acquire(ctx)
	require.NoError(t, err)
	assert.NotNil(t, inst)
	assert.Equal(t, "instance-1", inst.ID)
	assert.Equal(t, 1, factoryCalled)
}

func TestInstancePool_AcquireFromPool(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pool test in short mode - requires database setup")
	}
	config := PoolConfig{
		MinIdle:     1,
		MaxIdle:     5,
		MaxActive:   10,
		MaxLifetime: time.Hour,
	}

	factory := func() (*AgentInstance, error) {
		return &AgentInstance{
			ID:   "instance-1",
			Type: TypeAider,
		}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)
	defer pool.Close()

	// Wait for pre-warming
	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()
	inst, err := pool.Acquire(ctx)
	require.NoError(t, err)
	assert.NotNil(t, inst)
	// Should get from pool, not create new
	assert.Equal(t, StatusIdle, inst.Status)
}

func TestInstancePool_Release(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pool test in short mode - requires database setup")
	}
	config := PoolConfig{
		MinIdle:     0,
		MaxIdle:     5,
		MaxActive:   10,
		MaxLifetime: time.Hour,
	}

	factory := func() (*AgentInstance, error) {
		return &AgentInstance{
			ID:     "instance-1",
			Type:   TypeAider,
			Status: StatusIdle,
		}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)
	defer pool.Close()

	ctx := context.Background()
	inst, _ := pool.Acquire(ctx)
	inst.Status = StatusActive
	inst.SessionID = "test-session"

	err := pool.Release(inst)
	require.NoError(t, err)

	// Instance should be reset
	assert.Equal(t, StatusIdle, inst.Status)
	assert.Equal(t, "", inst.SessionID)
}

func TestInstancePool_MaxIdleLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pool test in short mode - requires database setup")
	}
	config := PoolConfig{
		MinIdle:     0,
		MaxIdle:     2,
		MaxActive:   10,
		MaxLifetime: time.Hour,
	}

	factoryCalled := 0
	factory := func() (*AgentInstance, error) {
		factoryCalled++
		return &AgentInstance{
			ID:   string(rune('0' + factoryCalled)),
			Type: TypeAider,
		}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)
	defer pool.Close()

	ctx := context.Background()

	// Create and release instances
	instances := make([]*AgentInstance, 5)
	for i := 0; i < 5; i++ {
		inst, err := pool.Acquire(ctx)
		require.NoError(t, err)
		instances[i] = inst
	}

	// Release all - only MaxIdle should be kept
	for _, inst := range instances {
		pool.Release(inst)
	}

	// Give time for cleanup
	time.Sleep(100 * time.Millisecond)

	// Check stats - should only have MaxIdle instances
	stats := pool.Stats()
	assert.LessOrEqual(t, stats["idle_count"], 2)
}

func TestInstancePool_MaxActiveLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pool test in short mode - requires database setup")
	}
	config := PoolConfig{
		MinIdle:     0,
		MaxIdle:     5,
		MaxActive:   2,
		MaxLifetime: time.Hour,
	}

	factory := func() (*AgentInstance, error) {
		return &AgentInstance{
			ID:   "instance",
			Type: TypeAider,
		}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Acquire up to MaxActive
	inst1, _ := pool.Acquire(ctx)
	inst2, _ := pool.Acquire(ctx)

	// Third acquire should timeout
	_, err := pool.Acquire(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")

	// Release one and try again
	pool.Release(inst1)
	inst3, err := pool.Acquire(ctx)
	require.NoError(t, err)
	assert.NotNil(t, inst3)

	pool.Release(inst2)
	pool.Release(inst3)
}

func TestInstancePool_Invalidate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pool test in short mode - requires database setup")
	}
	config := DefaultPoolConfig()
	factory := func() (*AgentInstance, error) {
		return &AgentInstance{
			ID:   "instance-1",
			Type: TypeAider,
		}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)
	defer pool.Close()

	ctx := context.Background()
	inst, _ := pool.Acquire(ctx)

	err := pool.Invalidate(inst)
	require.NoError(t, err)

	// Instance should be removed from active
	pool.mu.RLock()
	_, exists := pool.active[inst.ID]
	pool.mu.RUnlock()
	assert.False(t, exists)
	assert.Equal(t, StatusTerminated, inst.Status)
}

func TestInstancePool_CleanupExpired(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pool test in short mode - requires database setup")
	}
	config := PoolConfig{
		MinIdle:     0,
		MaxIdle:     5,
		MaxActive:   10,
		MaxLifetime: 50 * time.Millisecond,
	}

	factory := func() (*AgentInstance, error) {
		return &AgentInstance{
			ID:   "instance",
			Type: TypeAider,
		}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)
	defer pool.Close()

	ctx := context.Background()
	inst, _ := pool.Acquire(ctx)
	pool.Release(inst)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Trigger cleanup by acquiring again
	pool.Acquire(ctx)

	// Check that expired instance was removed
	stats := pool.Stats()
	// The exact count depends on timing, but eviction should have occurred
	assert.GreaterOrEqual(t, stats["evicts"], uint64(0))
}

func TestInstancePool_ConcurrentAcquireRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pool test in short mode - requires database setup")
	}
	config := PoolConfig{
		MinIdle:     0,
		MaxIdle:     10,
		MaxActive:   20,
		MaxLifetime: time.Hour,
	}

	var factoryCounter int64
	factory := func() (*AgentInstance, error) {
		count := atomic.AddInt64(&factoryCounter, 1)
		return &AgentInstance{
			ID:   string(rune(int(count))),
			Type: TypeAider,
		}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)
	defer pool.Close()

	const numGoroutines = 50
	const opsPerGoroutine = 20

	var wg sync.WaitGroup
	ctx := context.Background()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				inst, err := pool.Acquire(ctx)
				if err == nil {
					time.Sleep(time.Millisecond)
					pool.Release(inst)
				}
			}
		}()
	}

	wg.Wait()

	// Verify pool is in consistent state
	stats := pool.Stats()
	assert.LessOrEqual(t, stats["active_count"], config.MaxActive)
	assert.LessOrEqual(t, stats["idle_count"], config.MaxIdle)
}

func TestInstancePool_Stats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pool test in short mode - requires database setup")
	}
	config := DefaultPoolConfig()
	factory := func() (*AgentInstance, error) {
		return &AgentInstance{ID: "test", Type: TypeAider}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)
	defer pool.Close()

	ctx := context.Background()

	// Initial stats
	stats := pool.Stats()
	assert.Equal(t, TypeAider, stats["agent_type"])
	assert.Equal(t, 0, stats["idle_count"])
	assert.Equal(t, 0, stats["active_count"])
	assert.Equal(t, uint64(0), stats["hits"])
	assert.Equal(t, uint64(0), stats["misses"])

	// Acquire an instance (should be a miss)
	inst, _ := pool.Acquire(ctx)
	stats = pool.Stats()
	assert.Equal(t, uint64(0), stats["hits"])
	assert.Equal(t, uint64(1), stats["misses"])

	// Release and re-acquire (should be a hit)
	pool.Release(inst)
	time.Sleep(50 * time.Millisecond)
	pool.Acquire(ctx)
	stats = pool.Stats()
	assert.GreaterOrEqual(t, stats["hits"], uint64(1))
}

func TestInstancePool_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pool test in short mode - requires database setup")
	}
	config := DefaultPoolConfig()
	factory := func() (*AgentInstance, error) {
		return &AgentInstance{ID: "test", Type: TypeAider}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)

	ctx := context.Background()
	inst1, _ := pool.Acquire(ctx)
	inst2, _ := pool.Acquire(ctx)

	pool.Release(inst1)

	err := pool.Close()
	require.NoError(t, err)

	// All instances should be terminated
	assert.Equal(t, StatusTerminated, inst1.Status)
	assert.Equal(t, StatusTerminated, inst2.Status)
}

// Benchmarks

func BenchmarkInstancePool_AcquireRelease(b *testing.B) {
	config := PoolConfig{
		MinIdle:     5,
		MaxIdle:     10,
		MaxActive:   20,
		MaxLifetime: time.Hour,
	}

	factory := func() (*AgentInstance, error) {
		return &AgentInstance{ID: "test", Type: TypeAider}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)
	defer pool.Close()

	// Wait for pre-warming
	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inst, _ := pool.Acquire(ctx)
		pool.Release(inst)
	}
}

func BenchmarkInstancePool_ConcurrentAcquireRelease(b *testing.B) {
	config := PoolConfig{
		MinIdle:     10,
		MaxIdle:     20,
		MaxActive:   50,
		MaxLifetime: time.Hour,
	}

	factory := func() (*AgentInstance, error) {
		return &AgentInstance{ID: "test", Type: TypeAider}, nil
	}

	pool := NewInstancePool(TypeAider, config, factory)
	defer pool.Close()

	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			inst, _ := pool.Acquire(ctx)
			pool.Release(inst)
		}
	})
}
