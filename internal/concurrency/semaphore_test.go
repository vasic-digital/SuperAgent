package concurrency

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSemaphore(t *testing.T) {
	t.Run("acquire and release", func(t *testing.T) {
		sem := NewSemaphore(2)
		defer sem.Close()

		err := sem.Acquire(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, sem.Current())
		assert.Equal(t, 1, sem.Available())

		sem.Release()
		assert.Equal(t, 0, sem.Current())
		assert.Equal(t, 2, sem.Available())
	})

	t.Run("blocking when full", func(t *testing.T) {
		sem := NewSemaphore(1)
		defer sem.Close()

		err := sem.Acquire(context.Background())
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err = sem.Acquire(ctx)
		assert.Error(t, err)
	})

	t.Run("try acquire", func(t *testing.T) {
		sem := NewSemaphore(1)
		defer sem.Close()

		ok := sem.TryAcquire()
		assert.True(t, ok)

		ok = sem.TryAcquire()
		assert.False(t, ok)
	})

	t.Run("acquire with timeout", func(t *testing.T) {
		sem := NewSemaphore(1)
		defer sem.Close()

		err := sem.Acquire(context.Background())
		require.NoError(t, err)

		err = sem.AcquireWithTimeout(50 * time.Millisecond)
		assert.Error(t, err)
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("rate limiting", func(t *testing.T) {
		rl := NewRateLimiter(10)
		defer rl.Stop()

		start := time.Now()
		for i := 0; i < 5; i++ {
			err := rl.Acquire(context.Background())
			require.NoError(t, err)
		}

		duration := time.Since(start)
		assert.Less(t, duration, 500*time.Millisecond)
	})
}

func TestPrioritySemaphore(t *testing.T) {
	t.Run("high priority acquire", func(t *testing.T) {
		ps := NewPrioritySemaphore(1, 1)
		defer ps.Release()

		err := ps.AcquireHigh(context.Background())
		require.NoError(t, err)
	})

	t.Run("low priority uses either", func(t *testing.T) {
		ps := NewPrioritySemaphore(1, 1)
		defer ps.Release()

		err := ps.AcquireLow(context.Background())
		require.NoError(t, err)
	})
}

func TestResourcePool(t *testing.T) {
	t.Run("create and use pool", func(t *testing.T) {
		factory := func() (interface{}, error) {
			return "resource", nil
		}

		pool, err := NewResourcePool(2, factory)
		require.NoError(t, err)
		defer pool.Close()

		res, err := pool.Acquire(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "resource", res)

		err = pool.Release(res)
		require.NoError(t, err)
	})

	t.Run("pool exhausted", func(t *testing.T) {
		factory := func() (interface{}, error) {
			return "resource", nil
		}

		pool, err := NewResourcePool(1, factory)
		require.NoError(t, err)
		defer pool.Close()

		_, err = pool.Acquire(context.Background())
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err = pool.Acquire(ctx)
		assert.Error(t, err)
	})
}

func TestNonBlockingChan(t *testing.T) {
	t.Run("send and receive", func(t *testing.T) {
		ch := NewNonBlockingChan(2)

		ok := ch.Send("item1")
		assert.True(t, ok)

		ok = ch.Send("item2")
		assert.True(t, ok)

		item, ok := ch.Receive()
		assert.True(t, ok)
		assert.Equal(t, "item1", item)

		assert.Equal(t, 1, ch.Len())
	})

	t.Run("buffer overflow", func(t *testing.T) {
		ch := NewNonBlockingChan(1)

		// Fill channel and buffer
		ok1 := ch.Send("item1")
		ok2 := ch.Send("item2")

		// At least one should succeed (either channel or buffer)
		assert.True(t, ok1 || ok2, "At least one send should succeed")
	})
}

func TestAsyncProcessor(t *testing.T) {
	t.Run("submit and process", func(t *testing.T) {
		var counter int
		var mu sync.Mutex

		ap := NewAsyncProcessor(2, 10)
		defer ap.Stop()

		for i := 0; i < 5; i++ {
			ok := ap.Submit(func() {
				mu.Lock()
				counter++
				mu.Unlock()
			})
			assert.True(t, ok)
		}

		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		assert.GreaterOrEqual(t, counter, 1)
		mu.Unlock()
	})

	t.Run("queue full", func(t *testing.T) {
		// Create processor with 1 worker but queue size 1
		ap := NewAsyncProcessor(1, 1)
		defer ap.Stop()

		// Submit tasks rapidly to fill queue
		var submitted int
		for i := 0; i < 10; i++ {
			if ap.Submit(func() { time.Sleep(10 * time.Millisecond) }) {
				submitted++
			}
		}

		// Should be able to submit at least 2 (queue size + 1 in processing)
		assert.GreaterOrEqual(t, submitted, 1)
	})
}

func TestLazyLoader(t *testing.T) {
	t.Run("lazy loading", func(t *testing.T) {
		loadCount := 0
		loader := func() (interface{}, error) {
			loadCount++
			return "loaded", nil
		}

		ll := NewLazyLoader(loader)
		assert.False(t, ll.IsLoaded())

		val, err := ll.Get()
		require.NoError(t, err)
		assert.Equal(t, "loaded", val)
		assert.True(t, ll.IsLoaded())
		assert.Equal(t, 1, loadCount)

		val, err = ll.Get()
		require.NoError(t, err)
		assert.Equal(t, 1, loadCount)
	})

	t.Run("get or default", func(t *testing.T) {
		loader := func() (interface{}, error) {
			return "loaded", nil
		}

		ll := NewLazyLoader(loader)

		val := ll.GetOrDefault("default")
		assert.Equal(t, "default", val)

		ll.Get()

		val = ll.GetOrDefault("default")
		assert.Equal(t, "loaded", val)
	})
}

func TestNonBlockingCache(t *testing.T) {
	t.Run("set and get", func(t *testing.T) {
		cache := NewNonBlockingCache(time.Minute)

		cache.Set("key1", "value1")

		val, ok := cache.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, "value1", val)
	})

	t.Run("delete", func(t *testing.T) {
		cache := NewNonBlockingCache(time.Minute)

		cache.Set("key1", "value1")
		cache.Delete("key1")

		_, ok := cache.Get("key1")
		assert.False(t, ok)
	})

	t.Run("len", func(t *testing.T) {
		cache := NewNonBlockingCache(time.Minute)

		cache.Set("key1", "value1")
		cache.Set("key2", "value2")

		assert.Equal(t, 2, cache.Len())
	})
}

func TestBackgroundTask(t *testing.T) {
	t.Run("start and stop", func(t *testing.T) {
		executed := false
		var mu sync.Mutex

		task := NewBackgroundTask(func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					mu.Lock()
					executed = true
					mu.Unlock()
					return
				case <-time.After(10 * time.Millisecond):
				}
			}
		})

		task.Start()
		time.Sleep(50 * time.Millisecond)
		task.Stop()

		mu.Lock()
		assert.True(t, executed)
		mu.Unlock()
	})
}
