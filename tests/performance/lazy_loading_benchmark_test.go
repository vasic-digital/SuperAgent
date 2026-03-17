//go:build performance
// +build performance

package performance

import (
	"sync"
	"testing"
)

// BenchmarkSync_Once_Initialization benchmarks sync.Once overhead for lazy
// initialization — the primary pattern used by LazyProvider and all lazy
// singletons in the codebase. After the first call, sync.Once is essentially
// a single atomic load, so this measures the fast-path cost.
func BenchmarkSync_Once_Initialization(b *testing.B) {
	var once sync.Once
	var value int

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			once.Do(func() {
				value = 42
			})
			_ = value
		}
	})
}

// BenchmarkDirect_Access benchmarks direct variable access as a baseline
// to compare against sync.Once overhead.
func BenchmarkDirect_Access(b *testing.B) {
	value := 42

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = value
		}
	})
}

// BenchmarkMutex_Protected_Access benchmarks RWMutex-protected access to
// show the cost relative to sync.Once (which is lock-free on the fast path).
func BenchmarkMutex_Protected_Access(b *testing.B) {
	var mu sync.RWMutex
	value := 42

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.RLock()
			_ = value
			mu.RUnlock()
		}
	})
}

// BenchmarkSync_Once_MultipleInstances benchmarks many independent sync.Once
// instances to simulate the pattern used by LazyProviderRegistry where each
// provider has its own sync.Once.
func BenchmarkSync_Once_MultipleInstances(b *testing.B) {
	const numInstances = 20
	onces := make([]sync.Once, numInstances)
	values := make([]int, numInstances)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			idx := i % numInstances
			onces[idx].Do(func() {
				values[idx] = idx + 1
			})
			_ = values[idx]
			i++
		}
	})
}

// BenchmarkLazyInit_WithClosure benchmarks a closure-based lazy init pattern
// that closely mirrors the LazyProvider.Get() hot path.
func BenchmarkLazyInit_WithClosure(b *testing.B) {
	type lazyValue struct {
		once  sync.Once
		value interface{}
	}

	lv := &lazyValue{}
	factory := func() interface{} {
		return map[string]int{"key": 42}
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lv.once.Do(func() {
				lv.value = factory()
			})
			_ = lv.value
		}
	})
}
