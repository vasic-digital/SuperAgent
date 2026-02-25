package profiling

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type MemoryLeakDetector struct {
	mu sync.RWMutex

	snapshots    []MemorySnapshot
	maxSnapshots int

	thresholds MemoryThresholds
	alerts     chan MemoryAlert

	running bool
	stopCh  chan struct{}
}

type MemorySnapshot struct {
	Timestamp    time.Time
	Alloc        uint64
	TotalAlloc   uint64
	Sys          uint64
	NumGC        uint32
	NumGoroutine int
	HeapAlloc    uint64
	HeapSys      uint64
}

type MemoryThresholds struct {
	MaxAllocMB         uint64
	MaxSysMB           uint64
	MaxGoroutines      int
	GrowthRateMBPerSec float64
	CheckInterval      time.Duration
}

type MemoryAlert struct {
	Type       string
	Timestamp  time.Time
	Current    uint64
	Threshold  uint64
	Message    string
	Goroutines int
}

func DefaultMemoryThresholds() MemoryThresholds {
	return MemoryThresholds{
		MaxAllocMB:         1024,
		MaxSysMB:           2048,
		MaxGoroutines:      10000,
		GrowthRateMBPerSec: 10.0,
		CheckInterval:      10 * time.Second,
	}
}

func NewMemoryLeakDetector(thresholds MemoryThresholds) *MemoryLeakDetector {
	return &MemoryLeakDetector{
		snapshots:    make([]MemorySnapshot, 0, 100),
		maxSnapshots: 100,
		thresholds:   thresholds,
		alerts:       make(chan MemoryAlert, 100),
		stopCh:       make(chan struct{}),
	}
}

func (d *MemoryLeakDetector) Start(ctx context.Context) error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return fmt.Errorf("detector already running")
	}
	d.running = true
	d.mu.Unlock()

	go d.monitorLoop(ctx)
	return nil
}

func (d *MemoryLeakDetector) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return
	}

	close(d.stopCh)
	d.running = false
}

func (d *MemoryLeakDetector) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(d.thresholds.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.check()
		}
	}
}

func (d *MemoryLeakDetector) check() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	snapshot := MemorySnapshot{
		Timestamp:    time.Now(),
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		NumGoroutine: runtime.NumGoroutine(),
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
	}

	d.mu.Lock()
	d.snapshots = append(d.snapshots, snapshot)
	if len(d.snapshots) > d.maxSnapshots {
		d.snapshots = d.snapshots[1:]
	}
	d.mu.Unlock()

	d.checkThresholds(snapshot)
	d.checkGrowthRate()
}

func (d *MemoryLeakDetector) checkThresholds(snapshot MemorySnapshot) {
	allocMB := snapshot.Alloc / 1024 / 1024
	sysMB := snapshot.Sys / 1024 / 1024

	if d.thresholds.MaxAllocMB > 0 && allocMB > d.thresholds.MaxAllocMB {
		d.sendAlert(MemoryAlert{
			Type:       "alloc_exceeded",
			Timestamp:  snapshot.Timestamp,
			Current:    snapshot.Alloc,
			Threshold:  d.thresholds.MaxAllocMB * 1024 * 1024,
			Message:    fmt.Sprintf("Memory allocation exceeded threshold: %dMB > %dMB", allocMB, d.thresholds.MaxAllocMB),
			Goroutines: snapshot.NumGoroutine,
		})
	}

	if d.thresholds.MaxSysMB > 0 && sysMB > d.thresholds.MaxSysMB {
		d.sendAlert(MemoryAlert{
			Type:       "sys_exceeded",
			Timestamp:  snapshot.Timestamp,
			Current:    snapshot.Sys,
			Threshold:  d.thresholds.MaxSysMB * 1024 * 1024,
			Message:    fmt.Sprintf("System memory exceeded threshold: %dMB > %dMB", sysMB, d.thresholds.MaxSysMB),
			Goroutines: snapshot.NumGoroutine,
		})
	}

	if d.thresholds.MaxGoroutines > 0 && snapshot.NumGoroutine > d.thresholds.MaxGoroutines {
		d.sendAlert(MemoryAlert{
			Type:       "goroutines_exceeded",
			Timestamp:  snapshot.Timestamp,
			Current:    uint64(snapshot.NumGoroutine),
			Threshold:  uint64(d.thresholds.MaxGoroutines),
			Message:    fmt.Sprintf("Goroutine count exceeded threshold: %d > %d", snapshot.NumGoroutine, d.thresholds.MaxGoroutines),
			Goroutines: snapshot.NumGoroutine,
		})
	}
}

func (d *MemoryLeakDetector) checkGrowthRate() {
	d.mu.RLock()
	if len(d.snapshots) < 2 {
		d.mu.RUnlock()
		return
	}

	oldest := d.snapshots[0]
	newest := d.snapshots[len(d.snapshots)-1]
	d.mu.RUnlock()

	if newest.Timestamp.Equal(oldest.Timestamp) {
		return
	}

	duration := newest.Timestamp.Sub(oldest.Timestamp).Seconds()
	if duration == 0 {
		return
	}

	allocGrowth := float64(newest.Alloc-oldest.Alloc) / 1024 / 1024
	growthRate := allocGrowth / duration

	if growthRate > d.thresholds.GrowthRateMBPerSec {
		d.sendAlert(MemoryAlert{
			Type:       "growth_rate_exceeded",
			Timestamp:  newest.Timestamp,
			Current:    uint64(growthRate * 1024 * 1024),
			Threshold:  uint64(d.thresholds.GrowthRateMBPerSec * 1024 * 1024),
			Message:    fmt.Sprintf("Memory growth rate suspicious: %.2fMB/s > %.2fMB/s", growthRate, d.thresholds.GrowthRateMBPerSec),
			Goroutines: newest.NumGoroutine,
		})
	}
}

func (d *MemoryLeakDetector) sendAlert(alert MemoryAlert) {
	select {
	case d.alerts <- alert:
	default:
	}
}

func (d *MemoryLeakDetector) Alerts() <-chan MemoryAlert {
	return d.alerts
}

func (d *MemoryLeakDetector) GetSnapshot() MemorySnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemorySnapshot{
		Timestamp:    time.Now(),
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		NumGoroutine: runtime.NumGoroutine(),
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
	}
}

func (d *MemoryLeakDetector) GetHistory() []MemorySnapshot {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]MemorySnapshot, len(d.snapshots))
	copy(result, d.snapshots)
	return result
}

func (d *MemoryLeakDetector) ForceGC() {
	runtime.GC()
}

func (d *MemoryLeakDetector) DetectPotentialLeaks() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var issues []string

	if len(d.snapshots) < 5 {
		return issues
	}

	recent := d.snapshots[len(d.snapshots)-5:]

	allocTrend := make([]uint64, 5)
	for i, s := range recent {
		allocTrend[i] = s.Alloc
	}

	increasing := true
	for i := 1; i < len(allocTrend); i++ {
		if allocTrend[i] <= allocTrend[i-1] {
			increasing = false
			break
		}
	}

	if increasing {
		issues = append(issues, "memory allocation consistently increasing over last 5 samples")
	}

	gcCount := recent[len(recent)-1].NumGC - recent[0].NumGC
	if gcCount == 0 && recent[len(recent)-1].Alloc > recent[0].Alloc {
		issues = append(issues, "no GC triggered despite memory growth")
	}

	goroutineGrowth := recent[len(recent)-1].NumGoroutine - recent[0].NumGoroutine
	if goroutineGrowth > 50 {
		issues = append(issues, fmt.Sprintf("goroutine count grew by %d in recent samples", goroutineGrowth))
	}

	return issues
}
