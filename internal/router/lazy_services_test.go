package router

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLazyService_Get_InitializesOnce(t *testing.T) {
	var callCount int32
	ls := NewLazyService(func() (interface{}, error) {
		atomic.AddInt32(&callCount, 1)
		return "initialized", nil
	})

	// First call triggers factory
	svc, err := ls.Get()
	require.NoError(t, err)
	assert.Equal(t, "initialized", svc)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))

	// Second call returns cached result
	svc2, err2 := ls.Get()
	require.NoError(t, err2)
	assert.Equal(t, "initialized", svc2)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
}

func TestLazyService_Get_PropagatesError(t *testing.T) {
	expectedErr := errors.New("init failed")
	ls := NewLazyService(func() (interface{}, error) {
		return nil, expectedErr
	})

	svc, err := ls.Get()
	assert.Nil(t, svc)
	assert.ErrorIs(t, err, expectedErr)

	// Error is cached — factory not called again
	svc2, err2 := ls.Get()
	assert.Nil(t, svc2)
	assert.ErrorIs(t, err2, expectedErr)
}

func TestLazyService_Get_ConcurrentAccess(t *testing.T) {
	var callCount int32
	ls := NewLazyService(func() (interface{}, error) {
		atomic.AddInt32(&callCount, 1)
		return 42, nil
	})

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			svc, err := ls.Get()
			assert.NoError(t, err)
			assert.Equal(t, 42, svc)
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount),
		"factory must be called exactly once despite concurrent access")
}

func TestLazyService_IsInitialized(t *testing.T) {
	ls := NewLazyService(func() (interface{}, error) {
		return "value", nil
	})

	assert.False(t, ls.IsInitialized(), "should not be initialized before Get()")

	_, _ = ls.Get()
	assert.True(t, ls.IsInitialized(), "should be initialized after Get()")
}

func TestLazyService_IsInitialized_AfterError(t *testing.T) {
	ls := NewLazyService(func() (interface{}, error) {
		return nil, errors.New("boom")
	})

	assert.False(t, ls.IsInitialized())

	_, _ = ls.Get()
	assert.True(t, ls.IsInitialized(),
		"should report initialized even when factory returned an error")
}

func TestLazyServiceRegistry_RegisterAndGet(t *testing.T) {
	reg := NewLazyServiceRegistry()
	reg.Register("svc-a", NewLazyService(func() (interface{}, error) {
		return "alpha", nil
	}))

	svc, ok := reg.Get("svc-a")
	assert.True(t, ok)
	assert.Equal(t, "alpha", svc)
}

func TestLazyServiceRegistry_Get_UnknownName(t *testing.T) {
	reg := NewLazyServiceRegistry()

	svc, ok := reg.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, svc)
}

func TestLazyServiceRegistry_Get_FactoryError(t *testing.T) {
	reg := NewLazyServiceRegistry()
	reg.Register("bad", NewLazyService(func() (interface{}, error) {
		return nil, errors.New("factory failed")
	}))

	svc, ok := reg.Get("bad")
	assert.False(t, ok)
	assert.Nil(t, svc)
}

func TestLazyServiceRegistry_GetLazy(t *testing.T) {
	reg := NewLazyServiceRegistry()
	orig := NewLazyService(func() (interface{}, error) {
		return "lazy-val", nil
	})
	reg.Register("deferred", orig)

	ls, ok := reg.GetLazy("deferred")
	assert.True(t, ok)
	assert.False(t, ls.IsInitialized(), "GetLazy must not trigger initialization")

	// Now trigger it
	svc, err := ls.Get()
	require.NoError(t, err)
	assert.Equal(t, "lazy-val", svc)
	assert.True(t, ls.IsInitialized())
}

func TestLazyServiceRegistry_Names(t *testing.T) {
	reg := NewLazyServiceRegistry()
	reg.Register("x", NewLazyService(func() (interface{}, error) { return 1, nil }))
	reg.Register("y", NewLazyService(func() (interface{}, error) { return 2, nil }))
	reg.Register("z", NewLazyService(func() (interface{}, error) { return 3, nil }))

	names := reg.Names()
	assert.Len(t, names, 3)
	assert.ElementsMatch(t, []string{"x", "y", "z"}, names)
}

func TestLazyServiceRegistry_Count(t *testing.T) {
	reg := NewLazyServiceRegistry()
	assert.Equal(t, 0, reg.Count())

	reg.Register("a", NewLazyService(func() (interface{}, error) { return nil, nil }))
	assert.Equal(t, 1, reg.Count())

	reg.Register("b", NewLazyService(func() (interface{}, error) { return nil, nil }))
	assert.Equal(t, 2, reg.Count())
}

func TestLazyServiceRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewLazyServiceRegistry()
	var callCount int32

	for i := 0; i < 10; i++ {
		name := string(rune('a' + i))
		reg.Register(name, NewLazyService(func() (interface{}, error) {
			atomic.AddInt32(&callCount, 1)
			return name, nil
		}))
	}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			name := string(rune('a' + idx%10))
			svc, ok := reg.Get(name)
			assert.True(t, ok)
			assert.NotNil(t, svc)
		}(i)
	}
	wg.Wait()

	// Each of the 10 services should be initialized exactly once
	assert.Equal(t, int32(10), atomic.LoadInt32(&callCount))
}

func TestLazyServiceRegistry_OverwriteService(t *testing.T) {
	reg := NewLazyServiceRegistry()
	reg.Register("svc", NewLazyService(func() (interface{}, error) {
		return "first", nil
	}))

	// Overwrite with a new factory
	reg.Register("svc", NewLazyService(func() (interface{}, error) {
		return "second", nil
	}))

	svc, ok := reg.Get("svc")
	assert.True(t, ok)
	assert.Equal(t, "second", svc)
}
