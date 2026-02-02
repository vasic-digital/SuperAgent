package memory

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// helper to build a Memory for benchmarks
func benchMemory(id string) *Memory {
	return &Memory{
		ID:         id,
		UserID:     "user1",
		SessionID:  "session1",
		Content:    "The quick brown fox jumps over the lazy dog.",
		Importance: 0.5,
		Embedding:  []float32{0.1, 0.2, 0.3, 0.4},
		CreatedAt:  time.Now().Add(-time.Hour),
		UpdatedAt:  time.Now().Add(-30 * time.Minute),
		Metadata:   map[string]interface{}{"vector_clock": `{"node1":1}`},
	}
}

// helper to build a MemoryEvent for benchmarks
func benchEvent() *MemoryEvent {
	return &MemoryEvent{
		EventID:     "evt1",
		UserID:      "user1",
		SessionID:   "session1",
		Content:     "A completely new piece of content for the memory.",
		Importance:  0.8,
		Embedding:   []float32{0.5, 0.6, 0.7, 0.8},
		Timestamp:   time.Now(),
		VectorClock: `{"node1":2}`,
		Tags:        []string{"tag1", "tag2"},
		Entities: []MemoryEntity{
			{ID: "e1", Name: "Entity1", Confidence: 0.9},
		},
		Metadata: map[string]interface{}{"key": "value"},
	}
}

// BenchmarkCRDTMergeLWW benchmarks CRDTResolver.Merge with LastWriteWins.
func BenchmarkCRDTMergeLWW(b *testing.B) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)
	local := benchMemory("mem1")
	remote := benchEvent()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cr.Merge(local, remote)
	}
}

// BenchmarkCRDTMergeMergeAll benchmarks CRDTResolver.Merge with MergeAll.
func BenchmarkCRDTMergeMergeAll(b *testing.B) {
	cr := NewCRDTResolver(ConflictStrategyMergeAll)
	local := benchMemory("mem1")
	remote := benchEvent()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cr.Merge(local, remote)
	}
}

// BenchmarkCRDTMergeImportance benchmarks CRDTResolver.Merge with
// Importance strategy.
func BenchmarkCRDTMergeImportance(b *testing.B) {
	cr := NewCRDTResolver(ConflictStrategyImportance)
	local := benchMemory("mem1")
	remote := benchEvent()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cr.Merge(local, remote)
	}
}

// BenchmarkCRDTMergeVectorClock benchmarks CRDTResolver.Merge with
// VectorClock strategy (remote happens after local).
func BenchmarkCRDTMergeVectorClock(b *testing.B) {
	cr := NewCRDTResolver(ConflictStrategyVectorClock)
	local := benchMemory("mem1")
	remote := benchEvent()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cr.Merge(local, remote)
	}
}

// BenchmarkCRDTDetectConflict benchmarks DetectConflict when
// multiple conflict types are present.
func BenchmarkCRDTDetectConflict(b *testing.B) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)
	local := &Memory{
		Content:    "local content differs",
		Importance: 0.3,
		UpdatedAt:  time.Now(),
		Metadata:   map[string]interface{}{"tags": []string{"a"}},
	}
	remote := &MemoryEvent{
		Content:    "remote content differs",
		Importance: 0.9,
		Timestamp:  time.Now().Add(-time.Minute),
		Tags:       []string{"a", "b"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cr.DetectConflict(local, remote)
	}
}

// BenchmarkCRDTResolveWithReport benchmarks ResolveWithReport which
// combines conflict detection and resolution.
func BenchmarkCRDTResolveWithReport(b *testing.B) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)
	local := benchMemory("mem1")
	local.Content = "local content"
	local.Importance = 0.3
	local.UpdatedAt = time.Now()
	local.Metadata = map[string]interface{}{"tags": []string{"a"}}

	remote := benchEvent()
	remote.Content = "remote content"
	remote.Importance = 0.9
	remote.Tags = []string{"a", "b"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cr.ResolveWithReport(local, remote)
	}
}

// BenchmarkCRDTMergeTags benchmarks the mergeTags helper with overlapping
// tag sets.
func BenchmarkCRDTMergeTags(b *testing.B) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)
	local := make([]string, 50)
	remote := make([]string, 50)
	for i := 0; i < 50; i++ {
		local[i] = fmt.Sprintf("tag_%d", i)
		remote[i] = fmt.Sprintf("tag_%d", i+25) // 50% overlap
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cr.mergeTags(local, remote)
	}
}

// BenchmarkCRDTMergeEntities benchmarks mergeEntities with partial overlap.
func BenchmarkCRDTMergeEntities(b *testing.B) {
	cr := NewCRDTResolver(ConflictStrategyLastWriteWins)
	local := make([]MemoryEntity, 20)
	remote := make([]MemoryEntity, 20)
	for i := 0; i < 20; i++ {
		local[i] = MemoryEntity{
			ID:         fmt.Sprintf("e_%d", i),
			Name:       fmt.Sprintf("Entity_%d", i),
			Confidence: 0.5,
		}
		remote[i] = MemoryEntity{
			ID:         fmt.Sprintf("e_%d", i+10), // 50% overlap
			Name:       fmt.Sprintf("Entity_%d", i+10),
			Confidence: 0.7,
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cr.mergeEntities(local, remote)
	}
}

// BenchmarkCRDTMergeConcurrent measures concurrent Merge calls
// across multiple goroutines using b.RunParallel.
func BenchmarkCRDTMergeConcurrent(b *testing.B) {
	cr := NewCRDTResolver(ConflictStrategyMergeAll)

	var counter uint64
	var mu sync.Mutex

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		mu.Lock()
		counter++
		id := counter
		mu.Unlock()

		local := benchMemory(fmt.Sprintf("mem_%d", id))
		remote := benchEvent()

		for pb.Next() {
			cr.Merge(local, remote)
		}
	})
}
