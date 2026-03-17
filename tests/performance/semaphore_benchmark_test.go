//go:build performance
// +build performance

package performance

import (
	"context"
	"runtime"
	"sync"
	"testing"

	"golang.org/x/sync/semaphore"
)

// BenchmarkSemaphore_AcquireRelease benchmarks the weighted semaphore from
// golang.org/x/sync — this is the pattern now used by RunEnsembleWithProviders
// for context-aware concurrency limiting.
func BenchmarkSemaphore_AcquireRelease(b *testing.B) {
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = sem.Acquire(ctx, 1)
			sem.Release(1)
		}
	})
}

// BenchmarkChannel_Semaphore benchmarks the channel-based semaphore pattern
// (previously used by ensemble.go) for comparison.
func BenchmarkChannel_Semaphore(b *testing.B) {
	sem := make(chan struct{}, runtime.NumCPU())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sem <- struct{}{}
			<-sem
		}
	})
}

// BenchmarkUnlimited_Goroutines benchmarks spawning goroutines with no
// concurrency limiting as a baseline to measure the overhead of semaphores.
func BenchmarkUnlimited_Goroutines(b *testing.B) {
	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
		}()
	}
	wg.Wait()
}

// BenchmarkSemaphore_HighContention benchmarks the weighted semaphore under
// high contention (single slot) to measure worst-case scheduling latency.
func BenchmarkSemaphore_HighContention(b *testing.B) {
	sem := semaphore.NewWeighted(1)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = sem.Acquire(ctx, 1)
			sem.Release(1)
		}
	})
}

// BenchmarkChannel_HighContention benchmarks the channel semaphore under
// high contention (single slot) for comparison.
func BenchmarkChannel_HighContention(b *testing.B) {
	sem := make(chan struct{}, 1)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sem <- struct{}{}
			<-sem
		}
	})
}

// BenchmarkSemaphore_TryAcquire benchmarks the non-blocking TryAcquire path
// of the weighted semaphore.
func BenchmarkSemaphore_TryAcquire(b *testing.B) {
	sem := semaphore.NewWeighted(int64(runtime.NumCPU() * 4))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if sem.TryAcquire(1) {
				sem.Release(1)
			}
		}
	})
}
