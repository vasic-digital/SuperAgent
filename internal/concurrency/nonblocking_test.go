package concurrency

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// NonBlockingChan Tests
// =============================================================================

func TestNewNonBlockingChan(t *testing.T) {
	nbc := NewNonBlockingChan(10)
	assert.NotNil(t, nbc)
	assert.Equal(t, 10, nbc.size)
	assert.Equal(t, 0, nbc.Len())
}

func TestNonBlockingChan_Send_ToChannel(t *testing.T) {
	nbc := NewNonBlockingChan(5)

	ok := nbc.Send("item1")
	assert.True(t, ok)
	assert.Equal(t, 1, nbc.Len())
}

func TestNonBlockingChan_Send_ToBuffer(t *testing.T) {
	// Fill the channel first
	nbc := NewNonBlockingChan(2)

	// Fill channel
	nbc.Send("ch1")
	nbc.Send("ch2")

	// This should go to the buffer
	ok := nbc.Send("buf1")
	assert.True(t, ok)
	assert.Equal(t, 3, nbc.Len())
}

func TestNonBlockingChan_Send_Full(t *testing.T) {
	nbc := NewNonBlockingChan(2)

	// Fill channel
	nbc.Send("ch1")
	nbc.Send("ch2")

	// Fill buffer
	nbc.Send("buf1")
	nbc.Send("buf2")

	// Both channel and buffer full
	ok := nbc.Send("overflow")
	assert.False(t, ok)
}

func TestNonBlockingChan_Receive_FromChannel(t *testing.T) {
	nbc := NewNonBlockingChan(5)
	nbc.Send("item1")

	item, ok := nbc.Receive()
	assert.True(t, ok)
	assert.Equal(t, "item1", item)
}

func TestNonBlockingChan_Receive_FromBuffer(t *testing.T) {
	nbc := NewNonBlockingChan(1)

	// Fill channel
	nbc.Send("ch1")
	// This goes to buffer
	nbc.Send("buf1")

	// Drain channel
	item, ok := nbc.Receive()
	assert.True(t, ok)
	assert.Equal(t, "ch1", item)

	// Next should come from buffer
	item, ok = nbc.Receive()
	assert.True(t, ok)
	assert.Equal(t, "buf1", item)
}

func TestNonBlockingChan_Receive_Empty(t *testing.T) {
	nbc := NewNonBlockingChan(5)

	item, ok := nbc.Receive()
	assert.False(t, ok)
	assert.Nil(t, item)
}

func TestNonBlockingChan_Len(t *testing.T) {
	nbc := NewNonBlockingChan(2)
	assert.Equal(t, 0, nbc.Len())

	nbc.Send("a")
	assert.Equal(t, 1, nbc.Len())

	nbc.Send("b")
	assert.Equal(t, 2, nbc.Len())

	// Goes to buffer
	nbc.Send("c")
	assert.Equal(t, 3, nbc.Len())
}

func TestNonBlockingChan_Concurrent(t *testing.T) {
	nbc := NewNonBlockingChan(100)
	var wg sync.WaitGroup

	// Concurrent sends
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			nbc.Send(idx)
		}(i)
	}

	// Concurrent receives
	var received int32
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, ok := nbc.Receive(); ok {
				atomic.AddInt32(&received, 1)
			}
		}()
	}

	wg.Wait()
	// Should not panic due to concurrent access
}

// =============================================================================
// AsyncProcessor Tests
// =============================================================================

func TestNewAsyncProcessor(t *testing.T) {
	ap := NewAsyncProcessor(2, 10)
	assert.NotNil(t, ap)
	assert.Equal(t, 2, ap.workers)
	ap.Stop()
}

func TestAsyncProcessor_Submit(t *testing.T) {
	ap := NewAsyncProcessor(2, 10)
	defer ap.Stop()

	var executed int32
	done := make(chan struct{})

	ok := ap.Submit(func() {
		atomic.AddInt32(&executed, 1)
		close(done)
	})
	assert.True(t, ok)

	select {
	case <-done:
		assert.Equal(t, int32(1), atomic.LoadInt32(&executed))
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for task execution")
	}
}

func TestAsyncProcessor_Submit_QueueFull(t *testing.T) {
	// Create processor with very small queue and blocking tasks
	ap := NewAsyncProcessor(1, 1)
	defer ap.Stop()

	blocker := make(chan struct{})

	// Submit a task that blocks the worker
	ap.Submit(func() { <-blocker })

	// Fill the queue
	time.Sleep(10 * time.Millisecond) // Let worker pick up first task
	ap.Submit(func() {})              // Fill the queue slot

	// Queue should now be full
	ok := ap.Submit(func() {})
	assert.False(t, ok)

	close(blocker)
}

func TestAsyncProcessor_SubmitWithTimeout_Success(t *testing.T) {
	ap := NewAsyncProcessor(2, 10)
	defer ap.Stop()

	done := make(chan struct{})
	ok := ap.SubmitWithTimeout(func() {
		close(done)
	}, time.Second)
	assert.True(t, ok)

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestAsyncProcessor_SubmitWithTimeout_Timeout(t *testing.T) {
	ap := NewAsyncProcessor(1, 1)
	defer ap.Stop()

	blocker := make(chan struct{})

	// Block the worker
	ap.Submit(func() { <-blocker })
	time.Sleep(10 * time.Millisecond)
	// Fill queue
	ap.Submit(func() {})

	// This should timeout since queue is full
	ok := ap.SubmitWithTimeout(func() {}, 10*time.Millisecond)
	assert.False(t, ok)

	close(blocker)
}

func TestAsyncProcessor_Stop(t *testing.T) {
	ap := NewAsyncProcessor(4, 20)

	var counter int32
	for i := 0; i < 10; i++ {
		ap.Submit(func() {
			atomic.AddInt32(&counter, 1)
		})
	}

	ap.Stop()
	// After stop, all submitted tasks that were picked up should have completed
}

func TestAsyncProcessor_MultipleWorkers(t *testing.T) {
	ap := NewAsyncProcessor(4, 100)
	defer ap.Stop()

	var counter int32
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		ap.Submit(func() {
			defer wg.Done()
			atomic.AddInt32(&counter, 1)
		})
	}

	wg.Wait()
	assert.Equal(t, int32(50), atomic.LoadInt32(&counter))
}

// =============================================================================
// LazyLoader Tests
// =============================================================================

func TestNewLazyLoader(t *testing.T) {
	ll := NewLazyLoader(func() (interface{}, error) {
		return "loaded", nil
	})
	assert.NotNil(t, ll)
	assert.False(t, ll.IsLoaded())
}

func TestLazyLoader_Get_Success(t *testing.T) {
	callCount := 0
	ll := NewLazyLoader(func() (interface{}, error) {
		callCount++
		return "value", nil
	})

	val, err := ll.Get()
	assert.NoError(t, err)
	assert.Equal(t, "value", val)
	assert.True(t, ll.IsLoaded())
	assert.Equal(t, 1, callCount)

	// Second call should not invoke loader again
	val2, err2 := ll.Get()
	assert.NoError(t, err2)
	assert.Equal(t, "value", val2)
	assert.Equal(t, 1, callCount)
}

func TestLazyLoader_Get_Error(t *testing.T) {
	expectedErr := assert.AnError
	ll := NewLazyLoader(func() (interface{}, error) {
		return nil, expectedErr
	})

	val, err := ll.Get()
	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, val)
	assert.True(t, ll.IsLoaded())
}

func TestLazyLoader_IsLoaded(t *testing.T) {
	ll := NewLazyLoader(func() (interface{}, error) {
		return "data", nil
	})

	assert.False(t, ll.IsLoaded())
	_, _ = ll.Get()
	assert.True(t, ll.IsLoaded())
}

func TestLazyLoader_GetOrDefault_BeforeLoad(t *testing.T) {
	ll := NewLazyLoader(func() (interface{}, error) {
		return "loaded-value", nil
	})

	val := ll.GetOrDefault("default-value")
	assert.Equal(t, "default-value", val)
}

func TestLazyLoader_GetOrDefault_AfterLoad(t *testing.T) {
	ll := NewLazyLoader(func() (interface{}, error) {
		return "loaded-value", nil
	})

	_, _ = ll.Get()
	val := ll.GetOrDefault("default-value")
	assert.Equal(t, "loaded-value", val)
}

func TestLazyLoader_Concurrent(t *testing.T) {
	var callCount int32
	ll := NewLazyLoader(func() (interface{}, error) {
		atomic.AddInt32(&callCount, 1)
		return "result", nil
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := ll.Get()
			assert.NoError(t, err)
			assert.Equal(t, "result", val)
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
}

// =============================================================================
// NonBlockingCache Tests
// =============================================================================

func TestNewNonBlockingCache(t *testing.T) {
	cache := NewNonBlockingCache(5 * time.Minute)
	assert.NotNil(t, cache)
	assert.Equal(t, 0, cache.Len())
}

func TestNonBlockingCache_SetAndGet(t *testing.T) {
	cache := NewNonBlockingCache(5 * time.Minute)

	cache.Set("key1", "value1")
	val, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)
}

func TestNonBlockingCache_Get_Missing(t *testing.T) {
	cache := NewNonBlockingCache(5 * time.Minute)

	val, ok := cache.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestNonBlockingCache_Delete(t *testing.T) {
	cache := NewNonBlockingCache(5 * time.Minute)

	cache.Set("key1", "value1")
	assert.Equal(t, 1, cache.Len())

	cache.Delete("key1")
	assert.Equal(t, 0, cache.Len())

	_, ok := cache.Get("key1")
	assert.False(t, ok)
}

func TestNonBlockingCache_Delete_NonExistent(t *testing.T) {
	cache := NewNonBlockingCache(5 * time.Minute)
	cache.Delete("nonexistent") // Should not panic
	assert.Equal(t, 0, cache.Len())
}

func TestNonBlockingCache_Len(t *testing.T) {
	cache := NewNonBlockingCache(5 * time.Minute)

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)
	assert.Equal(t, 3, cache.Len())

	cache.Delete("b")
	assert.Equal(t, 2, cache.Len())
}

func TestNonBlockingCache_Overwrite(t *testing.T) {
	cache := NewNonBlockingCache(5 * time.Minute)

	cache.Set("key", "old")
	cache.Set("key", "new")

	val, ok := cache.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "new", val)
	assert.Equal(t, 1, cache.Len())
}

func TestNonBlockingCache_Concurrent(t *testing.T) {
	cache := NewNonBlockingCache(5 * time.Minute)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := "key"
			cache.Set(key, idx)
			cache.Get(key)
		}(i)
	}

	wg.Wait()
	// Should not panic from concurrent access
}

// =============================================================================
// BackgroundTask Tests
// =============================================================================

func TestNewBackgroundTask(t *testing.T) {
	bt := NewBackgroundTask(func(ctx context.Context) {
		<-ctx.Done()
	})
	assert.NotNil(t, bt)
}

func TestBackgroundTask_StartAndStop(t *testing.T) {
	var executed int32
	bt := NewBackgroundTask(func(ctx context.Context) {
		atomic.StoreInt32(&executed, 1)
		<-ctx.Done()
	})

	bt.Start()
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&executed))

	bt.Stop()

	// Done channel should be closed
	select {
	case <-bt.Done():
		// success
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for Done()")
	}
}

func TestBackgroundTask_Done(t *testing.T) {
	bt := NewBackgroundTask(func(ctx context.Context) {
		// Finish immediately
	})

	bt.Start()

	select {
	case <-bt.Done():
		// success: task completed
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for Done()")
	}
}

func TestBackgroundTask_ContextCancellation(t *testing.T) {
	var ctxDone int32
	bt := NewBackgroundTask(func(ctx context.Context) {
		<-ctx.Done()
		atomic.StoreInt32(&ctxDone, 1)
	})

	bt.Start()
	bt.Stop()

	assert.Equal(t, int32(1), atomic.LoadInt32(&ctxDone))
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkNonBlockingChan_SendReceive(b *testing.B) {
	nbc := NewNonBlockingChan(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nbc.Send(i)
		nbc.Receive()
	}
}

func BenchmarkNonBlockingCache_SetGet(b *testing.B) {
	cache := NewNonBlockingCache(5 * time.Minute)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", i)
		cache.Get("key")
	}
}

func BenchmarkLazyLoader_Get(b *testing.B) {
	ll := NewLazyLoader(func() (interface{}, error) {
		return "value", nil
	})
	_, _ = ll.Get() // Initialize
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ll.Get()
	}
}

func BenchmarkAsyncProcessor_Submit(b *testing.B) {
	ap := NewAsyncProcessor(4, 10000)
	defer ap.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ap.Submit(func() {})
	}
}
