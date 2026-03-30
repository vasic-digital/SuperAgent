//go:build performance
// +build performance

package performance

import (
	"fmt"
	"path/filepath"
	"testing"

	helixqaadapter "dev.helix.agent/internal/adapters/helixqa"
	"digital.vasic.visionengine/pkg/remote"
)

// BenchmarkVisionPool_AssignSlots measures slot assignment
// performance as the number of devices scales.
func BenchmarkVisionPool_AssignSlots(b *testing.B) {
	for _, numDevices := range []int{1, 5, 10, 50} {
		b.Run(fmt.Sprintf("devices=%d", numDevices), func(b *testing.B) {
			targets := make([]remote.SlotTarget, numDevices)
			for i := 0; i < numDevices; i++ {
				targets[i] = remote.SlotTarget{
					Platform: "android",
					Device:   fmt.Sprintf("device-%d", i),
				}
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pool := remote.NewVisionPool(remote.PoolConfig{
					Host:     "bench.local",
					BasePort: 9000,
				})
				pool.AssignSlots(targets)
			}
		})
	}
}

// BenchmarkVisionPool_GetSlot measures slot lookup performance.
func BenchmarkVisionPool_GetSlot(b *testing.B) {
	pool := remote.NewVisionPool(remote.PoolConfig{
		Host:     "bench.local",
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pool.GetSlot("android", fmt.Sprintf("device-%d", i%20))
	}
}

// BenchmarkHelixQAAdapter_Initialize measures adapter
// initialization performance (SQLite open + migrate).
func BenchmarkHelixQAAdapter_Initialize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		adapter := helixqaadapter.New(nil)
		dbPath := filepath.Join(b.TempDir(), fmt.Sprintf("bench-%d.db", i))
		_ = adapter.Initialize(dbPath)
		_ = adapter.Close()
	}
}

// BenchmarkHelixQAAdapter_SupportedPlatforms measures
// platform listing performance (should be constant time).
func BenchmarkHelixQAAdapter_SupportedPlatforms(b *testing.B) {
	adapter := helixqaadapter.New(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.SupportedPlatforms()
	}
}

// BenchmarkHelixQAAdapter_GetFindings measures findings
// retrieval from an empty store.
func BenchmarkHelixQAAdapter_GetFindings(b *testing.B) {
	adapter := helixqaadapter.New(nil)
	dbPath := filepath.Join(b.TempDir(), "bench.db")
	_ = adapter.Initialize(dbPath)
	defer adapter.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adapter.GetFindings("open")
	}
}
