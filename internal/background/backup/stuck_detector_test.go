// Package background provides comprehensive tests for stuck_detector.go —
// DefaultStuckDetector, StuckDetectorConfig, StuckAnalysis, and stuck
// detection algorithms (heartbeat, resource, activity, memory leak).
package background

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// =============================================================================
// StuckDetectorConfig
// =============================================================================

func TestDefaultStuckDetectorConfig_Values(t *testing.T) {
	cfg := DefaultStuckDetectorConfig()
	require.NotNil(t, cfg)

	assert.Equal(t, 5*time.Minute, cfg.DefaultThreshold)
	assert.InDelta(t, 0.1, cfg.CPUActivityThreshold, 0.001)
	assert.InDelta(t, 0.5, cfg.MemoryGrowthThreshold, 0.001)
	assert.Equal(t, int64(1024), cfg.IOActivityThreshold)
	assert.Equal(t, 3, cfg.MinSnapshotsForAnalysis)
}

func TestStuckDetectorConfig_ZeroValue(t *testing.T) {
	var cfg StuckDetectorConfig
	assert.Zero(t, cfg.DefaultThreshold)
	assert.Zero(t, cfg.CPUActivityThreshold)
	assert.Zero(t, cfg.MemoryGrowthThreshold)
	assert.Zero(t, cfg.IOActivityThreshold)
	assert.Zero(t, cfg.MinSnapshotsForAnalysis)
}

func TestStuckDetectorConfig_CustomValues(t *testing.T) {
	cfg := StuckDetectorConfig{
		DefaultThreshold:        10 * time.Minute,
		CPUActivityThreshold:    0.5,
		MemoryGrowthThreshold:   0.8,
		IOActivityThreshold:     2048,
		MinSnapshotsForAnalysis: 5,
	}

	assert.Equal(t, 10*time.Minute, cfg.DefaultThreshold)
	assert.InDelta(t, 0.5, cfg.CPUActivityThreshold, 0.001)
	assert.InDelta(t, 0.8, cfg.MemoryGrowthThreshold, 0.001)
	assert.Equal(t, int64(2048), cfg.IOActivityThreshold)
	assert.Equal(t, 5, cfg.MinSnapshotsForAnalysis)
}

// =============================================================================
// NewDefaultStuckDetector
// =============================================================================

func TestNewDefaultStuckDetector_Creation(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	require.NotNil(t, detector)
	assert.NotNil(t, detector.logger)
	assert.NotNil(t, detector.thresholds)
}

func TestNewDefaultStuckDetector_DefaultThresholds(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	assert.Equal(t, 5*time.Minute, detector.thresholds["default"])
	assert.Equal(t, 3*time.Minute, detector.thresholds["command"])
	assert.Equal(t, 3*time.Minute, detector.thresholds["llm_call"])
	assert.Equal(t, 10*time.Minute, detector.thresholds["debate"])
	assert.Equal(t, 2*time.Minute, detector.thresholds["embedding"])
	assert.Equal(t, time.Duration(0), detector.thresholds["endless"])
}

// =============================================================================
// GetStuckThreshold
// =============================================================================

func TestDefaultStuckDetector_GetStuckThreshold_KnownTypes(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	tests := []struct {
		taskType string
		expected time.Duration
	}{
		{"default", 5 * time.Minute},
		{"command", 3 * time.Minute},
		{"llm_call", 3 * time.Minute},
		{"debate", 10 * time.Minute},
		{"embedding", 2 * time.Minute},
		{"endless", 0},
	}

	for _, tc := range tests {
		t.Run(tc.taskType, func(t *testing.T) {
			assert.Equal(t, tc.expected, detector.GetStuckThreshold(tc.taskType))
		})
	}
}

func TestDefaultStuckDetector_GetStuckThreshold_UnknownType(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	// Unknown type falls back to default
	assert.Equal(t, 5*time.Minute, detector.GetStuckThreshold("unknown_type"))
	assert.Equal(t, 5*time.Minute, detector.GetStuckThreshold("custom_task"))
}

// =============================================================================
// SetThreshold
// =============================================================================

func TestDefaultStuckDetector_SetThreshold_NewType(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	detector.SetThreshold("custom", 15*time.Minute)
	assert.Equal(t, 15*time.Minute, detector.GetStuckThreshold("custom"))
}

func TestDefaultStuckDetector_SetThreshold_OverrideExisting(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	detector.SetThreshold("command", 20*time.Minute)
	assert.Equal(t, 20*time.Minute, detector.GetStuckThreshold("command"))
}

func TestDefaultStuckDetector_SetThreshold_Concurrent(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		taskType := "type_" + string(rune('a'+i%26))
		go func(tt string) {
			defer wg.Done()
			detector.SetThreshold(tt, time.Duration(i)*time.Second)
		}(taskType)
		go func(tt string) {
			defer wg.Done()
			_ = detector.GetStuckThreshold(tt)
		}(taskType)
	}
	wg.Wait()
}

// =============================================================================
// IsStuck — nil task
// =============================================================================

func TestDefaultStuckDetector_IsStuck_NilTaskReturnsNotStuck(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	stuck, reason := detector.IsStuck(ctx, nil, nil)
	assert.False(t, stuck)
	assert.Empty(t, reason)
}

// =============================================================================
// IsStuck — heartbeat timeout
// =============================================================================

func TestDefaultStuckDetector_IsStuck_HeartbeatTimeout(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	staleTime := time.Now().Add(-10 * time.Minute) // 10 minutes ago
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &staleTime,
		Config:        models.TaskConfig{Endless: false},
	}

	stuck, reason := detector.IsStuck(ctx, task, nil)
	assert.True(t, stuck)
	assert.Contains(t, reason, "no heartbeat")
}

func TestDefaultStuckDetector_IsStuck_FreshHeartbeat(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now() // Just now
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	stuck, reason := detector.IsStuck(ctx, task, nil)
	assert.False(t, stuck)
	assert.Empty(t, reason)
}

func TestDefaultStuckDetector_IsStuck_NilHeartbeat(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: nil, // Nil heartbeat is considered stale
		Config:        models.TaskConfig{Endless: false},
	}

	stuck, reason := detector.IsStuck(ctx, task, nil)
	assert.True(t, stuck)
	assert.Contains(t, reason, "no heartbeat ever received")
}

// =============================================================================
// IsStuck — custom stuck threshold
// =============================================================================

func TestDefaultStuckDetector_IsStuck_CustomStuckThreshold(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	// Task with very short custom threshold
	staleTime := time.Now().Add(-2 * time.Second)
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &staleTime,
		Config: models.TaskConfig{
			Endless:            false,
			StuckThresholdSecs: 1, // 1 second threshold
		},
	}

	stuck, reason := detector.IsStuck(ctx, task, nil)
	assert.True(t, stuck)
	assert.Contains(t, reason, "no heartbeat")
}

// =============================================================================
// IsStuck — deadline exceeded
// =============================================================================

func TestDefaultStuckDetector_IsStuck_DeadlineExceeded(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	pastDeadline := time.Now().Add(-1 * time.Hour)
	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Deadline:      &pastDeadline,
		Config:        models.TaskConfig{Endless: false},
	}

	stuck, reason := detector.IsStuck(ctx, task, nil)
	assert.True(t, stuck)
	assert.Contains(t, reason, "exceeded deadline")
}

func TestDefaultStuckDetector_IsStuck_DeadlineNotExceeded(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	futureDeadline := time.Now().Add(1 * time.Hour)
	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Deadline:      &futureDeadline,
		Config:        models.TaskConfig{Endless: false},
	}

	stuck, reason := detector.IsStuck(ctx, task, nil)
	assert.False(t, stuck)
	assert.Empty(t, reason)
}

// =============================================================================
// IsStuck — frozen process
// =============================================================================

func TestDefaultStuckDetector_IsStuck_FrozenProcess(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	// Create snapshots showing zero CPU activity
	snapshots := make([]*models.ResourceSnapshot, 5)
	for i := range snapshots {
		snapshots[i] = &models.ResourceSnapshot{
			CPUPercent:    0.0,
			CPUUserTime:   0.0,
			CPUSystemTime: 0.0,
		}
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.True(t, stuck)
	assert.Contains(t, reason, "frozen")
}

func TestDefaultStuckDetector_IsStuck_ActiveProcess(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	// CPU actively increasing
	snapshots := []*models.ResourceSnapshot{
		{CPUPercent: 25.0, CPUUserTime: 100.0, CPUSystemTime: 20.0},
		{CPUPercent: 20.0, CPUUserTime: 80.0, CPUSystemTime: 15.0},
		{CPUPercent: 15.0, CPUUserTime: 60.0, CPUSystemTime: 10.0},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.False(t, stuck)
	assert.Empty(t, reason)
}

// =============================================================================
// IsStuck — resource exhaustion
// =============================================================================

func TestDefaultStuckDetector_IsStuck_MemoryExhaustion_HighPercent(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	snapshots := []*models.ResourceSnapshot{
		{CPUPercent: 50.0, MemoryPercent: 96.0, CPUUserTime: 100, CPUSystemTime: 20},
		{CPUPercent: 50.0, MemoryPercent: 94.0, CPUUserTime: 80, CPUSystemTime: 15},
		{CPUPercent: 50.0, MemoryPercent: 92.0, CPUUserTime: 60, CPUSystemTime: 10},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.True(t, stuck)
	assert.Contains(t, reason, "memory exhaustion")
}

func TestDefaultStuckDetector_IsStuck_FileDescriptorExhaustion_Over10K(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	snapshots := []*models.ResourceSnapshot{
		{CPUPercent: 50.0, MemoryPercent: 50.0, OpenFDs: 15000, CPUUserTime: 100, CPUSystemTime: 20},
		{CPUPercent: 50.0, MemoryPercent: 50.0, OpenFDs: 14000, CPUUserTime: 80, CPUSystemTime: 15},
		{CPUPercent: 50.0, MemoryPercent: 50.0, OpenFDs: 13000, CPUUserTime: 60, CPUSystemTime: 10},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.True(t, stuck)
	assert.Contains(t, reason, "file descriptor exhaustion")
}

func TestDefaultStuckDetector_IsStuck_ExcessiveThreads(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	snapshots := []*models.ResourceSnapshot{
		{CPUPercent: 50.0, MemoryPercent: 50.0, ThreadCount: 2000, CPUUserTime: 100, CPUSystemTime: 20},
		{CPUPercent: 50.0, MemoryPercent: 50.0, ThreadCount: 1500, CPUUserTime: 80, CPUSystemTime: 15},
		{CPUPercent: 50.0, MemoryPercent: 50.0, ThreadCount: 1200, CPUUserTime: 60, CPUSystemTime: 10},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.True(t, stuck)
	assert.Contains(t, reason, "excessive threads")
}

// =============================================================================
// IsStuck — I/O starvation
// =============================================================================

func TestDefaultStuckDetector_IsStuck_IOStarved(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	// Low CPU but no I/O - I/O starvation
	snapshots := []*models.ResourceSnapshot{
		{CPUPercent: 0.5, IOReadBytes: 1000, IOWriteBytes: 500, MemoryPercent: 50, CPUUserTime: 10, CPUSystemTime: 5},
		{CPUPercent: 0.5, IOReadBytes: 1000, IOWriteBytes: 500, MemoryPercent: 50, CPUUserTime: 8, CPUSystemTime: 4},
		{CPUPercent: 0.5, IOReadBytes: 1000, IOWriteBytes: 500, MemoryPercent: 50, CPUUserTime: 6, CPUSystemTime: 3},
		{CPUPercent: 0.5, IOReadBytes: 1000, IOWriteBytes: 500, MemoryPercent: 50, CPUUserTime: 4, CPUSystemTime: 2},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.True(t, stuck)
	assert.Contains(t, reason, "I/O starved")
}

// =============================================================================
// IsStuck — network hang
// =============================================================================

func TestDefaultStuckDetector_IsStuck_NetworkHung(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	// Active connections but no data transfer, low CPU.
	// I/O bytes differ across snapshots so isIOStarved does NOT trigger first.
	snapshots := []*models.ResourceSnapshot{
		{
			CPUPercent: 0.5, NetConnections: 5,
			NetBytesSent: 1000, NetBytesRecv: 2000,
			IOReadBytes: 400, IOWriteBytes: 400,
			MemoryPercent: 50, CPUUserTime: 10, CPUSystemTime: 5,
		},
		{
			CPUPercent: 0.5, NetConnections: 5,
			NetBytesSent: 1000, NetBytesRecv: 2000,
			IOReadBytes: 300, IOWriteBytes: 300,
			MemoryPercent: 50, CPUUserTime: 8, CPUSystemTime: 4,
		},
		{
			CPUPercent: 0.5, NetConnections: 5,
			NetBytesSent: 1000, NetBytesRecv: 2000,
			IOReadBytes: 200, IOWriteBytes: 200,
			MemoryPercent: 50, CPUUserTime: 6, CPUSystemTime: 3,
		},
		{
			CPUPercent: 0.5, NetConnections: 5,
			NetBytesSent: 1000, NetBytesRecv: 2000,
			IOReadBytes: 100, IOWriteBytes: 100,
			MemoryPercent: 50, CPUUserTime: 4, CPUSystemTime: 2,
		},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.True(t, stuck)
	assert.Contains(t, reason, "network")
}

// =============================================================================
// IsStuck — memory leak
// =============================================================================

func TestDefaultStuckDetector_IsStuck_MemoryLeak(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	// Memory monotonically increasing by >50%
	snapshots := []*models.ResourceSnapshot{
		{CPUPercent: 50, MemoryRSSBytes: 200000000, MemoryPercent: 50, CPUUserTime: 100, CPUSystemTime: 20},
		{CPUPercent: 50, MemoryRSSBytes: 180000000, MemoryPercent: 45, CPUUserTime: 90, CPUSystemTime: 18},
		{CPUPercent: 50, MemoryRSSBytes: 160000000, MemoryPercent: 40, CPUUserTime: 80, CPUSystemTime: 16},
		{CPUPercent: 50, MemoryRSSBytes: 140000000, MemoryPercent: 35, CPUUserTime: 70, CPUSystemTime: 14},
		{CPUPercent: 50, MemoryRSSBytes: 120000000, MemoryPercent: 30, CPUUserTime: 60, CPUSystemTime: 12},
		{CPUPercent: 50, MemoryRSSBytes: 100000000, MemoryPercent: 25, CPUUserTime: 50, CPUSystemTime: 10},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.True(t, stuck)
	assert.Contains(t, reason, "memory leak")
}

func TestDefaultStuckDetector_IsStuck_NoMemoryLeak_Stable(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	// Memory stable
	snapshots := []*models.ResourceSnapshot{
		{CPUPercent: 50, MemoryRSSBytes: 100000000, MemoryPercent: 25, CPUUserTime: 100, CPUSystemTime: 20},
		{CPUPercent: 50, MemoryRSSBytes: 100000000, MemoryPercent: 25, CPUUserTime: 90, CPUSystemTime: 18},
		{CPUPercent: 50, MemoryRSSBytes: 100000000, MemoryPercent: 25, CPUUserTime: 80, CPUSystemTime: 16},
		{CPUPercent: 50, MemoryRSSBytes: 100000000, MemoryPercent: 25, CPUUserTime: 70, CPUSystemTime: 14},
		{CPUPercent: 50, MemoryRSSBytes: 100000000, MemoryPercent: 25, CPUUserTime: 60, CPUSystemTime: 12},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.False(t, stuck)
	assert.Empty(t, reason)
}

// =============================================================================
// IsStuck — endless tasks
// =============================================================================

func TestDefaultStuckDetector_IsStuck_EndlessTask_Healthy(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	task := &models.BackgroundTask{
		TaskType: "endless",
		Config:   models.TaskConfig{Endless: true},
	}

	snapshots := []*models.ResourceSnapshot{
		{CPUPercent: 10, MemoryPercent: 50, ProcessState: "running"},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.False(t, stuck)
	assert.Empty(t, reason)
}

func TestDefaultStuckDetector_IsStuck_EndlessTask_Zombie(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	task := &models.BackgroundTask{
		TaskType: "endless",
		Config:   models.TaskConfig{Endless: true},
	}

	snapshots := []*models.ResourceSnapshot{
		{ProcessState: "zombie", MemoryPercent: 50},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.True(t, stuck)
	assert.Contains(t, reason, "zombie")
}

func TestDefaultStuckDetector_IsStuck_EndlessTask_ZState(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	task := &models.BackgroundTask{
		TaskType: "endless",
		Config:   models.TaskConfig{Endless: true},
	}

	snapshots := []*models.ResourceSnapshot{
		{ProcessState: "Z", MemoryPercent: 50},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.True(t, stuck)
	assert.Contains(t, reason, "zombie")
}

func TestDefaultStuckDetector_IsStuck_EndlessTask_CriticalMemory(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	task := &models.BackgroundTask{
		TaskType: "endless",
		Config:   models.TaskConfig{Endless: true},
	}

	snapshots := []*models.ResourceSnapshot{
		{ProcessState: "running", MemoryPercent: 99.0},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.True(t, stuck)
	assert.Contains(t, reason, "memory exhaustion")
}

func TestDefaultStuckDetector_IsStuck_EndlessTask_NoActivity(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	task := &models.BackgroundTask{
		TaskType: "endless",
		Config:   models.TaskConfig{Endless: true},
	}

	// 5+ snapshots with zero activity
	snapshots := []*models.ResourceSnapshot{
		{CPUPercent: 0, IOReadBytes: 100, IOWriteBytes: 50, MemoryPercent: 30, ProcessState: "S"},
		{CPUPercent: 0, IOReadBytes: 100, IOWriteBytes: 50, MemoryPercent: 30, ProcessState: "S"},
		{CPUPercent: 0, IOReadBytes: 100, IOWriteBytes: 50, MemoryPercent: 30, ProcessState: "S"},
		{CPUPercent: 0, IOReadBytes: 100, IOWriteBytes: 50, MemoryPercent: 30, ProcessState: "S"},
		{CPUPercent: 0, IOReadBytes: 100, IOWriteBytes: 50, MemoryPercent: 30, ProcessState: "S"},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.True(t, stuck)
	assert.Contains(t, reason, "no activity")
}

func TestDefaultStuckDetector_IsStuck_EndlessTask_NoSnapshots(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	task := &models.BackgroundTask{
		TaskType: "endless",
		Config:   models.TaskConfig{Endless: true},
	}

	stuck, reason := detector.IsStuck(ctx, task, nil)
	assert.False(t, stuck)
	assert.Empty(t, reason)
}

func TestDefaultStuckDetector_IsStuck_EndlessTask_WithIOActivity(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	task := &models.BackgroundTask{
		TaskType: "endless",
		Config:   models.TaskConfig{Endless: true},
	}

	// IO changes between snapshots means activity
	snapshots := []*models.ResourceSnapshot{
		{CPUPercent: 0, IOReadBytes: 200, IOWriteBytes: 100, MemoryPercent: 30, ProcessState: "S"},
		{CPUPercent: 0, IOReadBytes: 150, IOWriteBytes: 80, MemoryPercent: 30, ProcessState: "S"},
		{CPUPercent: 0, IOReadBytes: 100, IOWriteBytes: 60, MemoryPercent: 30, ProcessState: "S"},
		{CPUPercent: 0, IOReadBytes: 50, IOWriteBytes: 40, MemoryPercent: 30, ProcessState: "S"},
		{CPUPercent: 0, IOReadBytes: 0, IOWriteBytes: 20, MemoryPercent: 30, ProcessState: "S"},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.False(t, stuck)
	assert.Empty(t, reason)
}

// =============================================================================
// IsStuck — fewer than 3 snapshots
// =============================================================================

func TestDefaultStuckDetector_IsStuck_TooFewSnapshots(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	// Only 2 snapshots (below threshold for deeper analysis)
	snapshots := []*models.ResourceSnapshot{
		{CPUPercent: 0.0, MemoryPercent: 96.0},
		{CPUPercent: 0.0, MemoryPercent: 94.0},
	}

	stuck, reason := detector.IsStuck(ctx, task, snapshots)
	assert.False(t, stuck)
	assert.Empty(t, reason)
}

// =============================================================================
// StuckAnalysis struct
// =============================================================================

func TestStuckAnalysis_ZeroValue(t *testing.T) {
	var sa StuckAnalysis
	assert.False(t, sa.IsStuck)
	assert.Empty(t, sa.Reason)
	assert.Nil(t, sa.Recommendations)
}

func TestHeartbeatStatus_ZeroValue(t *testing.T) {
	var hs HeartbeatStatus
	assert.Nil(t, hs.LastHeartbeat)
	assert.Zero(t, hs.TimeSinceHeartbeat)
	assert.Zero(t, hs.Threshold)
	assert.False(t, hs.IsStale)
}

func TestResourceStatus_ZeroValue(t *testing.T) {
	var rs ResourceStatus
	assert.Zero(t, rs.CPUPercent)
	assert.Zero(t, rs.MemoryPercent)
	assert.Zero(t, rs.MemoryBytes)
	assert.Zero(t, rs.OpenFDs)
	assert.Zero(t, rs.ThreadCount)
	assert.False(t, rs.IsExhausted)
}

func TestActivityStatus_ZeroValue(t *testing.T) {
	var as ActivityStatus
	assert.False(t, as.HasCPUActivity)
	assert.False(t, as.HasIOActivity)
	assert.False(t, as.HasNetActivity)
	assert.Zero(t, as.IOReadBytes)
	assert.Zero(t, as.IOWriteBytes)
	assert.Zero(t, as.NetConnections)
}

// =============================================================================
// AnalyzeTask
// =============================================================================

func TestDefaultStuckDetector_AnalyzeTask_BasicAnalysis(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	analysis := detector.AnalyzeTask(ctx, task, nil)
	require.NotNil(t, analysis)
	assert.False(t, analysis.IsStuck)
	assert.NotNil(t, analysis.Recommendations)
	assert.Equal(t, freshTime, *analysis.HeartbeatStatus.LastHeartbeat)
	assert.False(t, analysis.HeartbeatStatus.IsStale)
}

func TestDefaultStuckDetector_AnalyzeTask_StaleHeartbeat(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	staleTime := time.Now().Add(-10 * time.Minute)
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &staleTime,
		Config:        models.TaskConfig{Endless: false},
	}

	analysis := detector.AnalyzeTask(ctx, task, nil)
	require.NotNil(t, analysis)
	assert.True(t, analysis.IsStuck)
	assert.True(t, analysis.HeartbeatStatus.IsStale)
	assert.Contains(t, analysis.Reason, "no heartbeat")
	// Should recommend restarting
	found := false
	for _, rec := range analysis.Recommendations {
		if rec != "" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestDefaultStuckDetector_AnalyzeTask_NilHeartbeat(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: nil,
		Config:        models.TaskConfig{Endless: false},
	}

	analysis := detector.AnalyzeTask(ctx, task, nil)
	require.NotNil(t, analysis)
	assert.True(t, analysis.HeartbeatStatus.IsStale)
}

func TestDefaultStuckDetector_AnalyzeTask_WithSnapshots(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	snapshots := []*models.ResourceSnapshot{
		{
			CPUPercent: 25.0, MemoryPercent: 60.0, MemoryRSSBytes: 500000000,
			OpenFDs: 100, ThreadCount: 50,
			IOReadBytes: 2000, IOWriteBytes: 1000,
			NetBytesSent: 500, NetBytesRecv: 800, NetConnections: 3,
			CPUUserTime: 100, CPUSystemTime: 20,
		},
		{
			CPUPercent: 20.0, MemoryPercent: 55.0, MemoryRSSBytes: 450000000,
			OpenFDs: 95, ThreadCount: 48,
			IOReadBytes: 1500, IOWriteBytes: 800,
			NetBytesSent: 400, NetBytesRecv: 700, NetConnections: 3,
			CPUUserTime: 80, CPUSystemTime: 15,
		},
	}

	analysis := detector.AnalyzeTask(ctx, task, snapshots)
	require.NotNil(t, analysis)
	assert.False(t, analysis.IsStuck)

	// Check resource status is populated
	assert.InDelta(t, 25.0, analysis.ResourceStatus.CPUPercent, 0.1)
	assert.InDelta(t, 60.0, analysis.ResourceStatus.MemoryPercent, 0.1)
	assert.Equal(t, int64(500000000), analysis.ResourceStatus.MemoryBytes)
	assert.Equal(t, 100, analysis.ResourceStatus.OpenFDs)
	assert.Equal(t, 50, analysis.ResourceStatus.ThreadCount)
	assert.False(t, analysis.ResourceStatus.IsExhausted)

	// Check activity status
	assert.True(t, analysis.ActivityStatus.HasCPUActivity)
	assert.True(t, analysis.ActivityStatus.HasIOActivity)
	assert.True(t, analysis.ActivityStatus.HasNetActivity)
}

func TestDefaultStuckDetector_AnalyzeTask_Exhausted(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	freshTime := time.Now()
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &freshTime,
		Config:        models.TaskConfig{Endless: false},
	}

	snapshots := []*models.ResourceSnapshot{
		{
			CPUPercent: 50.0, MemoryPercent: 95.0, OpenFDs: 15000,
			CPUUserTime: 100, CPUSystemTime: 20,
		},
	}

	analysis := detector.AnalyzeTask(ctx, task, snapshots)
	require.NotNil(t, analysis)
	assert.True(t, analysis.ResourceStatus.IsExhausted)
	assert.NotEmpty(t, analysis.Recommendations)
}

func TestDefaultStuckDetector_AnalyzeTask_IdleStuck(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)
	ctx := context.Background()

	staleTime := time.Now().Add(-10 * time.Minute)
	task := &models.BackgroundTask{
		TaskType:      "command",
		LastHeartbeat: &staleTime,
		Config:        models.TaskConfig{Endless: false},
	}

	snapshots := []*models.ResourceSnapshot{
		{CPUPercent: 0, IOReadBytes: 0, IOWriteBytes: 0},
		{CPUPercent: 0, IOReadBytes: 0, IOWriteBytes: 0},
	}

	analysis := detector.AnalyzeTask(ctx, task, snapshots)
	require.NotNil(t, analysis)
	assert.True(t, analysis.IsStuck)

	// Should recommend deadlock check for completely idle
	hasDeadlockRec := false
	for _, rec := range analysis.Recommendations {
		if rec != "" {
			hasDeadlockRec = true
		}
	}
	assert.True(t, hasDeadlockRec)
}

// =============================================================================
// min3 helper function
// =============================================================================

func TestMin3_Various(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"FirstSmaller", 1, 5, 1},
		{"SecondSmaller", 5, 1, 1},
		{"Equal", 3, 3, 3},
		{"BothZero", 0, 0, 0},
		{"Negative", -5, 5, -5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, min3(tc.a, tc.b))
		})
	}
}

// =============================================================================
// StuckDetector interface compliance
// =============================================================================

func TestDefaultStuckDetector_ImplementsInterface(t *testing.T) {
	var _ StuckDetector = (*DefaultStuckDetector)(nil)
}
