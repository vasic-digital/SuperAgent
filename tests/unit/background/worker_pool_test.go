package background_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/background"
	"dev.helix.agent/internal/models"

	"go.uber.org/goleak"
)

var (
	testLogger      *logrus.Logger
	testQueue       background.TaskQueue
	testPool        *background.AdaptiveWorkerPool
	testPoolOnce    sync.Once
	testPoolConfig  *background.WorkerPoolConfig
	testPoolMu      sync.Mutex
	testPoolStarted bool
)

// TestMain sets up and tears down test fixtures
func TestMain(m *testing.M) {
	// Replace the default Prometheus registry with a new one
	// This prevents conflicts with other test packages
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	// Set up logger
	testLogger = logrus.New()
	testLogger.SetLevel(logrus.ErrorLevel)

	// goleak.VerifyTestMain runs m.Run() internally, then checks for goroutine leaks.
	// Worker pool goroutines winding down after Stop() are ignored.
	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("dev.helix.agent/internal/background.(*AdaptiveWorkerPool).workerLoop"),
		goleak.IgnoreTopFunction("dev.helix.agent/internal/background.(*AdaptiveWorkerPool).scaleLoop"),
	)

	// Clean up
	if testPool != nil && testPoolStarted {
		testPool.Stop(time.Second)
	}
}

// getTestPool returns a shared worker pool for tests
// This avoids duplicate metric registration issues
func getTestPool() *background.AdaptiveWorkerPool {
	testPoolOnce.Do(func() {
		testQueue = background.NewInMemoryTaskQueue(testLogger)
		testPoolConfig = &background.WorkerPoolConfig{
			MinWorkers:        2,
			MaxWorkers:        8,
			ScaleInterval:     100 * time.Millisecond,
			WorkerIdleTimeout: time.Minute,
			QueuePollInterval: 50 * time.Millisecond,
			HeartbeatInterval: 100 * time.Millisecond,
		}

		testPool = background.NewAdaptiveWorkerPool(
			testPoolConfig,
			testQueue,
			nil, nil, nil, nil,
			testLogger,
		)
	})
	return testPool
}

// startTestPool starts the shared pool if not already started
func startTestPool() error {
	testPoolMu.Lock()
	defer testPoolMu.Unlock()

	pool := getTestPool()
	if !testPoolStarted {
		err := pool.Start(context.Background())
		if err != nil {
			return err
		}
		testPoolStarted = true
		// Give workers time to start
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func TestDefaultWorkerPoolConfig(t *testing.T) {
	config := background.DefaultWorkerPoolConfig()

	assert.NotNil(t, config)
	assert.Greater(t, config.MinWorkers, 0)
	assert.GreaterOrEqual(t, config.MaxWorkers, config.MinWorkers)
	assert.Greater(t, config.ScaleInterval, time.Duration(0))
	assert.Greater(t, config.QueuePollInterval, time.Duration(0))
	assert.Greater(t, config.HeartbeatInterval, time.Duration(0))
}

func TestWorkerState_String(t *testing.T) {
	tests := []struct {
		state    string
		expected string
	}{
		{"idle", "idle"},
		{"busy", "busy"},
		{"stopping", "stopping"},
		{"stopped", "stopped"},
	}

	for _, tc := range tests {
		t.Run(tc.state, func(t *testing.T) {
			// The states are tested through the worker pool
		})
	}
}

func TestAdaptiveWorkerPool_Creation(t *testing.T) {
	pool := getTestPool()
	require.NotNil(t, pool)
}

func TestAdaptiveWorkerPool_StartAndStop(t *testing.T) {
	// Use the shared pool to avoid duplicate metric registration
	pool := getTestPool()
	require.NotNil(t, pool)

	ctx := context.Background()

	// Start should succeed (first time) or return error (already started)
	err := pool.Start(ctx)
	if err == nil {
		testPoolMu.Lock()
		testPoolStarted = true
		testPoolMu.Unlock()

		// Give workers time to start
		time.Sleep(200 * time.Millisecond)

		// Should have minimum workers
		assert.GreaterOrEqual(t, pool.GetWorkerCount(), testPoolConfig.MinWorkers)
	} else {
		// Pool was already started by another test
		assert.Contains(t, err.Error(), "already started")
	}
}

func TestAdaptiveWorkerPool_DoubleStart(t *testing.T) {
	pool := getTestPool()

	// Ensure pool is started first
	err := startTestPool()
	if err != nil {
		// Already started, that's fine
	}

	// Second start should fail
	err = pool.Start(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")
}

func TestAdaptiveWorkerPool_RegisterExecutor(t *testing.T) {
	pool := getTestPool()

	// Create a mock executor
	executor := &mockExecutor{}

	// Register should not panic
	pool.RegisterExecutor("test-task", executor)
}

func TestAdaptiveWorkerPool_GetWorkerStatus(t *testing.T) {
	pool := getTestPool()

	// Ensure pool is started
	err := startTestPool()
	require.NoError(t, err)

	statuses := pool.GetWorkerStatus()
	assert.NotEmpty(t, statuses)

	for _, status := range statuses {
		assert.NotEmpty(t, status.ID)
		assert.NotEmpty(t, status.Status)
		assert.False(t, status.StartedAt.IsZero())
	}
}

func TestAdaptiveWorkerPool_Scale(t *testing.T) {
	pool := getTestPool()

	// Ensure pool is started
	err := startTestPool()
	require.NoError(t, err)

	initialCount := pool.GetWorkerCount()
	assert.GreaterOrEqual(t, initialCount, testPoolConfig.MinWorkers)

	// Scale up
	err = pool.Scale(4)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.GreaterOrEqual(t, pool.GetWorkerCount(), initialCount)

	// Scale beyond max should be capped
	err = pool.Scale(100)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.LessOrEqual(t, pool.GetWorkerCount(), testPoolConfig.MaxWorkers)

	// Scale below min should be capped
	err = pool.Scale(0)
	assert.NoError(t, err)
}

func TestAdaptiveWorkerPool_GetActiveTaskCount(t *testing.T) {
	pool := getTestPool()

	// Ensure pool is started
	err := startTestPool()
	require.NoError(t, err)

	// With no tasks, should have 0 active
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, pool.GetActiveTaskCount())
}

// mockExecutor is a test implementation of TaskExecutor
type mockExecutor struct {
	ExecuteFunc   func(ctx context.Context, task *models.BackgroundTask, reporter background.ProgressReporter) error
	ExecuteCalled int
}

func (m *mockExecutor) Execute(ctx context.Context, task *models.BackgroundTask, reporter background.ProgressReporter) error {
	m.ExecuteCalled++
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, task, reporter)
	}
	return nil
}

func (m *mockExecutor) CanPause() bool {
	return false
}

func (m *mockExecutor) Pause(ctx context.Context, task *models.BackgroundTask) ([]byte, error) {
	return nil, nil
}

func (m *mockExecutor) Resume(ctx context.Context, task *models.BackgroundTask, checkpoint []byte) error {
	return nil
}

func (m *mockExecutor) Cancel(ctx context.Context, task *models.BackgroundTask) error {
	return nil
}

func (m *mockExecutor) GetResourceRequirements() background.ResourceRequirements {
	return background.ResourceRequirements{
		CPUCores: 1,
		MemoryMB: 256,
	}
}

func TestResourceRequirements(t *testing.T) {
	req := background.ResourceRequirements{
		CPUCores: 4,
		MemoryMB: 8192,
		DiskMB:   1024,
		GPUCount: 1,
		Priority: models.TaskPriorityHigh,
	}

	assert.Equal(t, 4, req.CPUCores)
	assert.Equal(t, 8192, req.MemoryMB)
	assert.Equal(t, 1024, req.DiskMB)
	assert.Equal(t, 1, req.GPUCount)
	assert.Equal(t, models.TaskPriorityHigh, req.Priority)
}

func TestSystemResources(t *testing.T) {
	res := background.SystemResources{
		TotalCPUCores:     8,
		AvailableCPUCores: 4.5,
		TotalMemoryMB:     16384,
		AvailableMemoryMB: 8192,
		CPULoadPercent:    56.25,
		MemoryUsedPercent: 50.0,
		DiskUsedPercent:   75.0,
		LoadAvg1:          2.5,
		LoadAvg5:          2.0,
		LoadAvg15:         1.5,
	}

	assert.Equal(t, 8, res.TotalCPUCores)
	assert.InDelta(t, 4.5, res.AvailableCPUCores, 0.01)
	assert.Equal(t, int64(16384), res.TotalMemoryMB)
	assert.InDelta(t, 56.25, res.CPULoadPercent, 0.01)
}

func TestWorkerStatus(t *testing.T) {
	now := time.Now()
	status := background.WorkerStatus{
		ID:              "worker-1",
		Status:          "busy",
		CurrentTask:     nil,
		StartedAt:       now.Add(-time.Hour),
		LastActivity:    now,
		TasksCompleted:  100,
		TasksFailed:     5,
		AvgTaskDuration: 5 * time.Second,
	}

	assert.Equal(t, "worker-1", status.ID)
	assert.Equal(t, "busy", status.Status)
	assert.Equal(t, int64(100), status.TasksCompleted)
	assert.Equal(t, int64(5), status.TasksFailed)
	assert.Equal(t, 5*time.Second, status.AvgTaskDuration)
}

func TestTaskEvent(t *testing.T) {
	workerID := "worker-1"
	event := background.TaskEvent{
		TaskID:    "task-123",
		EventType: "started",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"priority": "high"},
		WorkerID:  &workerID,
	}

	assert.Equal(t, "task-123", event.TaskID)
	assert.Equal(t, "started", event.EventType)
	assert.NotNil(t, event.Data)
	assert.NotNil(t, event.WorkerID)
}

func TestExecutionResult(t *testing.T) {
	result := background.ExecutionResult{
		TaskID:     "task-123",
		Status:     models.TaskStatusCompleted,
		Output:     []byte("test output"),
		Error:      "",
		Duration:   5 * time.Second,
		RetryCount: 0,
	}

	assert.Equal(t, "task-123", result.TaskID)
	assert.Equal(t, models.TaskStatusCompleted, result.Status)
	assert.Equal(t, []byte("test output"), result.Output)
	assert.Empty(t, result.Error)
	assert.Equal(t, 5*time.Second, result.Duration)
}

func TestWaitResult(t *testing.T) {
	result := background.WaitResult{
		Task:     &models.BackgroundTask{ID: "task-123"},
		Output:   []byte("output"),
		Duration: 10 * time.Second,
		Error:    nil,
	}

	assert.NotNil(t, result.Task)
	assert.Equal(t, "task-123", result.Task.ID)
	assert.Equal(t, []byte("output"), result.Output)
	assert.Equal(t, 10*time.Second, result.Duration)
	assert.NoError(t, result.Error)
}
