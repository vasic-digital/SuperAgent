package background

import (
	"context"
	"errors"
	"testing"
	"time"

	internalbackground "dev.helix.agent/internal/background"
	internalmodels "dev.helix.agent/internal/models"
	extractedbackground "digital.vasic.background"
	extractedmodels "digital.vasic.models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock implementations for extracted interfaces
// =============================================================================

// mockExtractedTaskQueue implements extractedbackground.TaskQueue
type mockExtractedTaskQueue struct {
	enqueueFn       func(ctx context.Context, task *extractedmodels.BackgroundTask) error
	dequeueFn       func(ctx context.Context, workerID string, req extractedbackground.ResourceRequirements) (*extractedmodels.BackgroundTask, error)
	peekFn          func(ctx context.Context, count int) ([]*extractedmodels.BackgroundTask, error)
	requeueFn       func(ctx context.Context, taskID string, delay time.Duration) error
	moveToDeadFn    func(ctx context.Context, taskID string, reason string) error
	getPendingFn    func(ctx context.Context) (int64, error)
	getRunningFn    func(ctx context.Context) (int64, error)
	getQueueDepthFn func(ctx context.Context) (map[extractedmodels.TaskPriority]int64, error)
}

func (m *mockExtractedTaskQueue) Enqueue(ctx context.Context, task *extractedmodels.BackgroundTask) error {
	if m.enqueueFn != nil {
		return m.enqueueFn(ctx, task)
	}
	return nil
}

func (m *mockExtractedTaskQueue) Dequeue(ctx context.Context, workerID string, req extractedbackground.ResourceRequirements) (*extractedmodels.BackgroundTask, error) {
	if m.dequeueFn != nil {
		return m.dequeueFn(ctx, workerID, req)
	}
	return &extractedmodels.BackgroundTask{ID: "task-1", TaskType: "test"}, nil
}

func (m *mockExtractedTaskQueue) Peek(ctx context.Context, count int) ([]*extractedmodels.BackgroundTask, error) {
	if m.peekFn != nil {
		return m.peekFn(ctx, count)
	}
	return []*extractedmodels.BackgroundTask{{ID: "task-1"}}, nil
}

func (m *mockExtractedTaskQueue) Requeue(ctx context.Context, taskID string, delay time.Duration) error {
	if m.requeueFn != nil {
		return m.requeueFn(ctx, taskID, delay)
	}
	return nil
}

func (m *mockExtractedTaskQueue) MoveToDeadLetter(ctx context.Context, taskID string, reason string) error {
	if m.moveToDeadFn != nil {
		return m.moveToDeadFn(ctx, taskID, reason)
	}
	return nil
}

func (m *mockExtractedTaskQueue) GetPendingCount(ctx context.Context) (int64, error) {
	if m.getPendingFn != nil {
		return m.getPendingFn(ctx)
	}
	return 5, nil
}

func (m *mockExtractedTaskQueue) GetRunningCount(ctx context.Context) (int64, error) {
	if m.getRunningFn != nil {
		return m.getRunningFn(ctx)
	}
	return 3, nil
}

func (m *mockExtractedTaskQueue) GetQueueDepth(ctx context.Context) (map[extractedmodels.TaskPriority]int64, error) {
	if m.getQueueDepthFn != nil {
		return m.getQueueDepthFn(ctx)
	}
	return map[extractedmodels.TaskPriority]int64{
		extractedmodels.TaskPriority("normal"): 10,
	}, nil
}

// mockExtractedWorkerPool implements extractedbackground.WorkerPool
type mockExtractedWorkerPool struct {
	startFn           func(ctx context.Context) error
	stopFn            func(gracePeriod time.Duration) error
	registerFn        func(taskType string, executor extractedbackground.TaskExecutor)
	getWorkerCountFn  func() int
	getActiveCountFn  func() int
	getWorkerStatusFn func() []extractedbackground.WorkerStatus
	scaleFn           func(targetCount int) error
}

func (m *mockExtractedWorkerPool) Start(ctx context.Context) error {
	if m.startFn != nil {
		return m.startFn(ctx)
	}
	return nil
}

func (m *mockExtractedWorkerPool) Stop(gracePeriod time.Duration) error {
	if m.stopFn != nil {
		return m.stopFn(gracePeriod)
	}
	return nil
}

func (m *mockExtractedWorkerPool) RegisterExecutor(taskType string, executor extractedbackground.TaskExecutor) {
	if m.registerFn != nil {
		m.registerFn(taskType, executor)
	}
}

func (m *mockExtractedWorkerPool) GetWorkerCount() int {
	if m.getWorkerCountFn != nil {
		return m.getWorkerCountFn()
	}
	return 4
}

func (m *mockExtractedWorkerPool) GetActiveTaskCount() int {
	if m.getActiveCountFn != nil {
		return m.getActiveCountFn()
	}
	return 2
}

func (m *mockExtractedWorkerPool) GetWorkerStatus() []extractedbackground.WorkerStatus {
	if m.getWorkerStatusFn != nil {
		return m.getWorkerStatusFn()
	}
	return []extractedbackground.WorkerStatus{
		{
			ID:             "worker-1",
			Status:         "idle",
			StartedAt:      time.Now(),
			LastActivity:   time.Now(),
			TasksCompleted: 10,
			TasksFailed:    1,
		},
	}
}

func (m *mockExtractedWorkerPool) Scale(targetCount int) error {
	if m.scaleFn != nil {
		return m.scaleFn(targetCount)
	}
	return nil
}

// mockInternalTaskExecutor implements internalbackground.TaskExecutor
type mockInternalTaskExecutor struct {
	executeFn  func(ctx context.Context, task *internalmodels.BackgroundTask, reporter internalbackground.ProgressReporter) error
	canPauseFn func() bool
	pauseFn    func(ctx context.Context, task *internalmodels.BackgroundTask) ([]byte, error)
	resumeFn   func(ctx context.Context, task *internalmodels.BackgroundTask, checkpoint []byte) error
	cancelFn   func(ctx context.Context, task *internalmodels.BackgroundTask) error
	getReqFn   func() internalbackground.ResourceRequirements
}

func (m *mockInternalTaskExecutor) Execute(ctx context.Context, task *internalmodels.BackgroundTask, reporter internalbackground.ProgressReporter) error {
	if m.executeFn != nil {
		return m.executeFn(ctx, task, reporter)
	}
	return nil
}

func (m *mockInternalTaskExecutor) CanPause() bool {
	if m.canPauseFn != nil {
		return m.canPauseFn()
	}
	return true
}

func (m *mockInternalTaskExecutor) Pause(ctx context.Context, task *internalmodels.BackgroundTask) ([]byte, error) {
	if m.pauseFn != nil {
		return m.pauseFn(ctx, task)
	}
	return []byte("checkpoint"), nil
}

func (m *mockInternalTaskExecutor) Resume(ctx context.Context, task *internalmodels.BackgroundTask, checkpoint []byte) error {
	if m.resumeFn != nil {
		return m.resumeFn(ctx, task, checkpoint)
	}
	return nil
}

func (m *mockInternalTaskExecutor) Cancel(ctx context.Context, task *internalmodels.BackgroundTask) error {
	if m.cancelFn != nil {
		return m.cancelFn(ctx, task)
	}
	return nil
}

func (m *mockInternalTaskExecutor) GetResourceRequirements() internalbackground.ResourceRequirements {
	if m.getReqFn != nil {
		return m.getReqFn()
	}
	return internalbackground.ResourceRequirements{
		CPUCores: 2,
		MemoryMB: 1024,
		Priority: internalmodels.TaskPriorityNormal,
	}
}

// mockExtractedProgressReporter implements extractedbackground.ProgressReporter
type mockExtractedProgressReporter struct {
	progressFn   func(percent float64, message string) error
	heartbeatFn  func() error
	checkpointFn func(data []byte) error
	metricsFn    func(metrics map[string]interface{}) error
	logFn        func(level, message string, fields map[string]interface{}) error
}

func (m *mockExtractedProgressReporter) ReportProgress(percent float64, message string) error {
	if m.progressFn != nil {
		return m.progressFn(percent, message)
	}
	return nil
}

func (m *mockExtractedProgressReporter) ReportHeartbeat() error {
	if m.heartbeatFn != nil {
		return m.heartbeatFn()
	}
	return nil
}

func (m *mockExtractedProgressReporter) ReportCheckpoint(data []byte) error {
	if m.checkpointFn != nil {
		return m.checkpointFn(data)
	}
	return nil
}

func (m *mockExtractedProgressReporter) ReportMetrics(metrics map[string]interface{}) error {
	if m.metricsFn != nil {
		return m.metricsFn(metrics)
	}
	return nil
}

func (m *mockExtractedProgressReporter) ReportLog(level, message string, fields map[string]interface{}) error {
	if m.logFn != nil {
		return m.logFn(level, message, fields)
	}
	return nil
}

// mockExtractedResourceMonitor implements extractedbackground.ResourceMonitor
type mockExtractedResourceMonitor struct {
	getSystemFn   func() (*extractedbackground.SystemResources, error)
	getProcessFn  func(pid int) (*extractedmodels.ResourceSnapshot, error)
	startMonFn    func(taskID string, pid int, interval time.Duration) error
	stopMonFn     func(taskID string) error
	getSnapshotFn func(taskID string) (*extractedmodels.ResourceSnapshot, error)
	isAvailFn     func(req extractedbackground.ResourceRequirements) bool
}

func (m *mockExtractedResourceMonitor) GetSystemResources() (*extractedbackground.SystemResources, error) {
	if m.getSystemFn != nil {
		return m.getSystemFn()
	}
	return &extractedbackground.SystemResources{
		TotalCPUCores:     8,
		AvailableCPUCores: 4.0,
		TotalMemoryMB:     16384,
		AvailableMemoryMB: 8192,
	}, nil
}

func (m *mockExtractedResourceMonitor) GetProcessResources(pid int) (*extractedmodels.ResourceSnapshot, error) {
	if m.getProcessFn != nil {
		return m.getProcessFn(pid)
	}
	return &extractedmodels.ResourceSnapshot{
		ID:         "snap-1",
		TaskID:     "task-1",
		CPUPercent: 25.0,
	}, nil
}

func (m *mockExtractedResourceMonitor) StartMonitoring(taskID string, pid int, interval time.Duration) error {
	if m.startMonFn != nil {
		return m.startMonFn(taskID, pid, interval)
	}
	return nil
}

func (m *mockExtractedResourceMonitor) StopMonitoring(taskID string) error {
	if m.stopMonFn != nil {
		return m.stopMonFn(taskID)
	}
	return nil
}

func (m *mockExtractedResourceMonitor) GetLatestSnapshot(taskID string) (*extractedmodels.ResourceSnapshot, error) {
	if m.getSnapshotFn != nil {
		return m.getSnapshotFn(taskID)
	}
	return &extractedmodels.ResourceSnapshot{ID: "snap-2", TaskID: taskID}, nil
}

func (m *mockExtractedResourceMonitor) IsResourceAvailable(req extractedbackground.ResourceRequirements) bool {
	if m.isAvailFn != nil {
		return m.isAvailFn(req)
	}
	return true
}

// mockExtractedStuckDetector implements extractedbackground.StuckDetector
type mockExtractedStuckDetector struct {
	isStuckFn      func(ctx context.Context, task *extractedmodels.BackgroundTask, snapshots []*extractedmodels.ResourceSnapshot) (bool, string)
	getThresholdFn func(taskType string) time.Duration
	setThresholdFn func(taskType string, threshold time.Duration)
}

func (m *mockExtractedStuckDetector) IsStuck(ctx context.Context, task *extractedmodels.BackgroundTask, snapshots []*extractedmodels.ResourceSnapshot) (bool, string) {
	if m.isStuckFn != nil {
		return m.isStuckFn(ctx, task, snapshots)
	}
	return false, ""
}

func (m *mockExtractedStuckDetector) GetStuckThreshold(taskType string) time.Duration {
	if m.getThresholdFn != nil {
		return m.getThresholdFn(taskType)
	}
	return 5 * time.Minute
}

func (m *mockExtractedStuckDetector) SetThreshold(taskType string, threshold time.Duration) {
	if m.setThresholdFn != nil {
		m.setThresholdFn(taskType, threshold)
	}
}

// mockInternalNotificationService implements internalbackground.NotificationService
type mockInternalNotificationService struct {
	notifyFn      func(ctx context.Context, task *internalmodels.BackgroundTask, event string, data map[string]interface{}) error
	registerSSEFn func(ctx context.Context, taskID string, client chan<- []byte) error
	unregisterFn  func(ctx context.Context, taskID string, client chan<- []byte) error
	registerWSFn  func(ctx context.Context, taskID string, client internalbackground.WebSocketClient) error
	broadcastFn   func(ctx context.Context, taskID string, message []byte) error
}

func (m *mockInternalNotificationService) NotifyTaskEvent(ctx context.Context, task *internalmodels.BackgroundTask, event string, data map[string]interface{}) error {
	if m.notifyFn != nil {
		return m.notifyFn(ctx, task, event, data)
	}
	return nil
}

func (m *mockInternalNotificationService) RegisterSSEClient(ctx context.Context, taskID string, client chan<- []byte) error {
	if m.registerSSEFn != nil {
		return m.registerSSEFn(ctx, taskID, client)
	}
	return nil
}

func (m *mockInternalNotificationService) UnregisterSSEClient(ctx context.Context, taskID string, client chan<- []byte) error {
	if m.unregisterFn != nil {
		return m.unregisterFn(ctx, taskID, client)
	}
	return nil
}

func (m *mockInternalNotificationService) RegisterWebSocketClient(ctx context.Context, taskID string, client internalbackground.WebSocketClient) error {
	if m.registerWSFn != nil {
		return m.registerWSFn(ctx, taskID, client)
	}
	return nil
}

func (m *mockInternalNotificationService) BroadcastToTask(ctx context.Context, taskID string, message []byte) error {
	if m.broadcastFn != nil {
		return m.broadcastFn(ctx, taskID, message)
	}
	return nil
}

// mockExtractedWebSocketClient implements extractedbackground.WebSocketClient
type mockExtractedWebSocketClient struct {
	sendFn  func(data []byte) error
	closeFn func() error
	idFn    func() string
}

func (m *mockExtractedWebSocketClient) Send(data []byte) error {
	if m.sendFn != nil {
		return m.sendFn(data)
	}
	return nil
}

func (m *mockExtractedWebSocketClient) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

func (m *mockExtractedWebSocketClient) ID() string {
	if m.idFn != nil {
		return m.idFn()
	}
	return "ws-client-1"
}

// =============================================================================
// TaskQueueAdapter Tests
// =============================================================================

func TestNewTaskQueueAdapter(t *testing.T) {
	queue := &mockExtractedTaskQueue{}
	adapter := NewTaskQueueAdapter(queue)

	require.NotNil(t, adapter)
	assert.Equal(t, queue, adapter.queue)
}

func TestTaskQueueAdapter_Enqueue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		queue := &mockExtractedTaskQueue{}
		adapter := NewTaskQueueAdapter(queue)

		task := &internalmodels.BackgroundTask{
			ID:       "task-1",
			TaskType: "test",
			Status:   internalmodels.TaskStatusPending,
		}

		err := adapter.Enqueue(context.Background(), task)
		assert.NoError(t, err)
	})

	t.Run("error propagation", func(t *testing.T) {
		expectedErr := errors.New("enqueue failed")
		queue := &mockExtractedTaskQueue{
			enqueueFn: func(ctx context.Context, task *extractedmodels.BackgroundTask) error {
				return expectedErr
			},
		}
		adapter := NewTaskQueueAdapter(queue)

		err := adapter.Enqueue(context.Background(), &internalmodels.BackgroundTask{ID: "task-1"})
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("nil task", func(t *testing.T) {
		queue := &mockExtractedTaskQueue{}
		adapter := NewTaskQueueAdapter(queue)

		// nil task converts to nil extracted task
		err := adapter.Enqueue(context.Background(), nil)
		assert.NoError(t, err)
	})
}

func TestTaskQueueAdapter_Dequeue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		queue := &mockExtractedTaskQueue{}
		adapter := NewTaskQueueAdapter(queue)

		reqs := internalbackground.ResourceRequirements{
			CPUCores: 2,
			MemoryMB: 512,
			Priority: internalmodels.TaskPriorityNormal,
		}

		task, err := adapter.Dequeue(context.Background(), "worker-1", reqs)
		assert.NoError(t, err)
		assert.NotNil(t, task)
	})

	t.Run("error propagation", func(t *testing.T) {
		expectedErr := errors.New("dequeue failed")
		queue := &mockExtractedTaskQueue{
			dequeueFn: func(ctx context.Context, workerID string, req extractedbackground.ResourceRequirements) (*extractedmodels.BackgroundTask, error) {
				return nil, expectedErr
			},
		}
		adapter := NewTaskQueueAdapter(queue)

		task, err := adapter.Dequeue(context.Background(), "worker-1", internalbackground.ResourceRequirements{})
		assert.Error(t, err)
		assert.Nil(t, task)
	})
}

func TestTaskQueueAdapter_Peek(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		queue := &mockExtractedTaskQueue{}
		adapter := NewTaskQueueAdapter(queue)

		tasks, err := adapter.Peek(context.Background(), 5)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
	})

	t.Run("error propagation", func(t *testing.T) {
		expectedErr := errors.New("peek failed")
		queue := &mockExtractedTaskQueue{
			peekFn: func(ctx context.Context, count int) ([]*extractedmodels.BackgroundTask, error) {
				return nil, expectedErr
			},
		}
		adapter := NewTaskQueueAdapter(queue)

		tasks, err := adapter.Peek(context.Background(), 5)
		assert.Error(t, err)
		assert.Nil(t, tasks)
	})
}

func TestTaskQueueAdapter_Requeue(t *testing.T) {
	queue := &mockExtractedTaskQueue{}
	adapter := NewTaskQueueAdapter(queue)

	err := adapter.Requeue(context.Background(), "task-1", 5*time.Second)
	assert.NoError(t, err)
}

func TestTaskQueueAdapter_MoveToDeadLetter(t *testing.T) {
	queue := &mockExtractedTaskQueue{}
	adapter := NewTaskQueueAdapter(queue)

	err := adapter.MoveToDeadLetter(context.Background(), "task-1", "max retries exceeded")
	assert.NoError(t, err)
}

func TestTaskQueueAdapter_GetPendingCount(t *testing.T) {
	queue := &mockExtractedTaskQueue{}
	adapter := NewTaskQueueAdapter(queue)

	count, err := adapter.GetPendingCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestTaskQueueAdapter_GetRunningCount(t *testing.T) {
	queue := &mockExtractedTaskQueue{}
	adapter := NewTaskQueueAdapter(queue)

	count, err := adapter.GetRunningCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestTaskQueueAdapter_GetQueueDepth(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		queue := &mockExtractedTaskQueue{}
		adapter := NewTaskQueueAdapter(queue)

		depth, err := adapter.GetQueueDepth(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, depth)
		assert.Equal(t, int64(10), depth[internalmodels.TaskPriority("normal")])
	})

	t.Run("error propagation", func(t *testing.T) {
		expectedErr := errors.New("depth failed")
		queue := &mockExtractedTaskQueue{
			getQueueDepthFn: func(ctx context.Context) (map[extractedmodels.TaskPriority]int64, error) {
				return nil, expectedErr
			},
		}
		adapter := NewTaskQueueAdapter(queue)

		depth, err := adapter.GetQueueDepth(context.Background())
		assert.Error(t, err)
		assert.Nil(t, depth)
	})
}

// =============================================================================
// WorkerPoolAdapter Tests
// =============================================================================

func TestNewWorkerPoolAdapter(t *testing.T) {
	pool := &mockExtractedWorkerPool{}
	adapter := NewWorkerPoolAdapter(pool)

	require.NotNil(t, adapter)
	assert.Equal(t, pool, adapter.pool)
}

func TestWorkerPoolAdapter_Start(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pool := &mockExtractedWorkerPool{}
		adapter := NewWorkerPoolAdapter(pool)

		err := adapter.Start(context.Background())
		assert.NoError(t, err)
	})

	t.Run("error propagation", func(t *testing.T) {
		expectedErr := errors.New("start failed")
		pool := &mockExtractedWorkerPool{
			startFn: func(ctx context.Context) error {
				return expectedErr
			},
		}
		adapter := NewWorkerPoolAdapter(pool)

		err := adapter.Start(context.Background())
		assert.Error(t, err)
	})
}

func TestWorkerPoolAdapter_Stop(t *testing.T) {
	pool := &mockExtractedWorkerPool{}
	adapter := NewWorkerPoolAdapter(pool)

	err := adapter.Stop(5 * time.Second)
	assert.NoError(t, err)
}

func TestWorkerPoolAdapter_RegisterExecutor(t *testing.T) {
	var registeredType string
	pool := &mockExtractedWorkerPool{
		registerFn: func(taskType string, executor extractedbackground.TaskExecutor) {
			registeredType = taskType
		},
	}
	adapter := NewWorkerPoolAdapter(pool)

	executor := &mockInternalTaskExecutor{}
	adapter.RegisterExecutor("test-type", executor)

	assert.Equal(t, "test-type", registeredType)
}

func TestWorkerPoolAdapter_GetWorkerCount(t *testing.T) {
	pool := &mockExtractedWorkerPool{}
	adapter := NewWorkerPoolAdapter(pool)

	count := adapter.GetWorkerCount()
	assert.Equal(t, 4, count)
}

func TestWorkerPoolAdapter_GetActiveTaskCount(t *testing.T) {
	pool := &mockExtractedWorkerPool{}
	adapter := NewWorkerPoolAdapter(pool)

	count := adapter.GetActiveTaskCount()
	assert.Equal(t, 2, count)
}

func TestWorkerPoolAdapter_GetWorkerStatus(t *testing.T) {
	pool := &mockExtractedWorkerPool{}
	adapter := NewWorkerPoolAdapter(pool)

	statuses := adapter.GetWorkerStatus()
	require.Len(t, statuses, 1)
	assert.Equal(t, "worker-1", statuses[0].ID)
	assert.Equal(t, "idle", statuses[0].Status)
	assert.Equal(t, int64(10), statuses[0].TasksCompleted)
	assert.Equal(t, int64(1), statuses[0].TasksFailed)
}

func TestWorkerPoolAdapter_Scale(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pool := &mockExtractedWorkerPool{}
		adapter := NewWorkerPoolAdapter(pool)

		err := adapter.Scale(8)
		assert.NoError(t, err)
	})

	t.Run("error propagation", func(t *testing.T) {
		expectedErr := errors.New("scale failed")
		pool := &mockExtractedWorkerPool{
			scaleFn: func(targetCount int) error {
				return expectedErr
			},
		}
		adapter := NewWorkerPoolAdapter(pool)

		err := adapter.Scale(8)
		assert.Error(t, err)
	})
}

// =============================================================================
// TaskExecutorAdapter Tests
// =============================================================================

func TestNewTaskExecutorAdapter(t *testing.T) {
	executor := &mockInternalTaskExecutor{}
	adapter := NewTaskExecutorAdapter(executor)

	require.NotNil(t, adapter)
	assert.Equal(t, executor, adapter.executor)
}

func TestTaskExecutorAdapter_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		executor := &mockInternalTaskExecutor{}
		adapter := NewTaskExecutorAdapter(executor)
		reporter := &mockExtractedProgressReporter{}

		task := &extractedmodels.BackgroundTask{
			ID:       "task-1",
			TaskType: "test",
		}

		err := adapter.Execute(context.Background(), task, reporter)
		assert.NoError(t, err)
	})

	t.Run("error propagation", func(t *testing.T) {
		expectedErr := errors.New("execute failed")
		executor := &mockInternalTaskExecutor{
			executeFn: func(ctx context.Context, task *internalmodels.BackgroundTask, reporter internalbackground.ProgressReporter) error {
				return expectedErr
			},
		}
		adapter := NewTaskExecutorAdapter(executor)

		err := adapter.Execute(context.Background(), &extractedmodels.BackgroundTask{ID: "task-1"}, &mockExtractedProgressReporter{})
		assert.Error(t, err)
	})
}

func TestTaskExecutorAdapter_CanPause(t *testing.T) {
	tests := []struct {
		name     string
		canPause bool
	}{
		{"can pause", true},
		{"cannot pause", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &mockInternalTaskExecutor{
				canPauseFn: func() bool { return tt.canPause },
			}
			adapter := NewTaskExecutorAdapter(executor)
			assert.Equal(t, tt.canPause, adapter.CanPause())
		})
	}
}

func TestTaskExecutorAdapter_Pause(t *testing.T) {
	executor := &mockInternalTaskExecutor{}
	adapter := NewTaskExecutorAdapter(executor)

	task := &extractedmodels.BackgroundTask{ID: "task-1"}
	checkpoint, err := adapter.Pause(context.Background(), task)

	assert.NoError(t, err)
	assert.Equal(t, []byte("checkpoint"), checkpoint)
}

func TestTaskExecutorAdapter_Resume(t *testing.T) {
	executor := &mockInternalTaskExecutor{}
	adapter := NewTaskExecutorAdapter(executor)

	task := &extractedmodels.BackgroundTask{ID: "task-1"}
	err := adapter.Resume(context.Background(), task, []byte("checkpoint"))
	assert.NoError(t, err)
}

func TestTaskExecutorAdapter_Cancel(t *testing.T) {
	executor := &mockInternalTaskExecutor{}
	adapter := NewTaskExecutorAdapter(executor)

	task := &extractedmodels.BackgroundTask{ID: "task-1"}
	err := adapter.Cancel(context.Background(), task)
	assert.NoError(t, err)
}

func TestTaskExecutorAdapter_GetResourceRequirements(t *testing.T) {
	executor := &mockInternalTaskExecutor{}
	adapter := NewTaskExecutorAdapter(executor)

	reqs := adapter.GetResourceRequirements()
	assert.Equal(t, 2, reqs.CPUCores)
	assert.Equal(t, 1024, reqs.MemoryMB)
}

// =============================================================================
// ProgressReporterAdapter Tests
// =============================================================================

func TestNewProgressReporterAdapter(t *testing.T) {
	reporter := &mockExtractedProgressReporter{}
	adapter := NewProgressReporterAdapter(reporter)

	require.NotNil(t, adapter)
	assert.Equal(t, reporter, adapter.reporter)
}

func TestProgressReporterAdapter_ReportProgress(t *testing.T) {
	var calledPercent float64
	var calledMsg string

	reporter := &mockExtractedProgressReporter{
		progressFn: func(percent float64, message string) error {
			calledPercent = percent
			calledMsg = message
			return nil
		},
	}
	adapter := NewProgressReporterAdapter(reporter)

	err := adapter.ReportProgress(75.0, "processing")
	assert.NoError(t, err)
	assert.Equal(t, 75.0, calledPercent)
	assert.Equal(t, "processing", calledMsg)
}

func TestProgressReporterAdapter_ReportHeartbeat(t *testing.T) {
	reporter := &mockExtractedProgressReporter{}
	adapter := NewProgressReporterAdapter(reporter)

	err := adapter.ReportHeartbeat()
	assert.NoError(t, err)
}

func TestProgressReporterAdapter_ReportCheckpoint(t *testing.T) {
	reporter := &mockExtractedProgressReporter{}
	adapter := NewProgressReporterAdapter(reporter)

	err := adapter.ReportCheckpoint([]byte("data"))
	assert.NoError(t, err)
}

func TestProgressReporterAdapter_ReportMetrics(t *testing.T) {
	reporter := &mockExtractedProgressReporter{}
	adapter := NewProgressReporterAdapter(reporter)

	metrics := map[string]interface{}{"cpu": 50.0}
	err := adapter.ReportMetrics(metrics)
	assert.NoError(t, err)
}

func TestProgressReporterAdapter_ReportLog(t *testing.T) {
	reporter := &mockExtractedProgressReporter{}
	adapter := NewProgressReporterAdapter(reporter)

	fields := map[string]interface{}{"key": "value"}
	err := adapter.ReportLog("info", "test message", fields)
	assert.NoError(t, err)
}

// =============================================================================
// ResourceMonitorAdapter Tests
// =============================================================================

func TestNewResourceMonitorAdapter(t *testing.T) {
	monitor := &mockExtractedResourceMonitor{}
	adapter := NewResourceMonitorAdapter(monitor)

	require.NotNil(t, adapter)
	assert.Equal(t, monitor, adapter.monitor)
}

func TestResourceMonitorAdapter_GetSystemResources(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		monitor := &mockExtractedResourceMonitor{}
		adapter := NewResourceMonitorAdapter(monitor)

		resources, err := adapter.GetSystemResources()
		assert.NoError(t, err)
		assert.NotNil(t, resources)
		assert.Equal(t, 8, resources.TotalCPUCores)
	})

	t.Run("error propagation", func(t *testing.T) {
		expectedErr := errors.New("system resource error")
		monitor := &mockExtractedResourceMonitor{
			getSystemFn: func() (*extractedbackground.SystemResources, error) {
				return nil, expectedErr
			},
		}
		adapter := NewResourceMonitorAdapter(monitor)

		resources, err := adapter.GetSystemResources()
		assert.Error(t, err)
		assert.Nil(t, resources)
	})
}

func TestResourceMonitorAdapter_GetProcessResources(t *testing.T) {
	monitor := &mockExtractedResourceMonitor{}
	adapter := NewResourceMonitorAdapter(monitor)

	snapshot, err := adapter.GetProcessResources(12345)
	assert.NoError(t, err)
	assert.NotNil(t, snapshot)
}

func TestResourceMonitorAdapter_StartMonitoring(t *testing.T) {
	monitor := &mockExtractedResourceMonitor{}
	adapter := NewResourceMonitorAdapter(monitor)

	err := adapter.StartMonitoring("task-1", 12345, 10*time.Second)
	assert.NoError(t, err)
}

func TestResourceMonitorAdapter_StopMonitoring(t *testing.T) {
	monitor := &mockExtractedResourceMonitor{}
	adapter := NewResourceMonitorAdapter(monitor)

	err := adapter.StopMonitoring("task-1")
	assert.NoError(t, err)
}

func TestResourceMonitorAdapter_GetLatestSnapshot(t *testing.T) {
	monitor := &mockExtractedResourceMonitor{}
	adapter := NewResourceMonitorAdapter(monitor)

	snapshot, err := adapter.GetLatestSnapshot("task-1")
	assert.NoError(t, err)
	assert.NotNil(t, snapshot)
}

func TestResourceMonitorAdapter_IsResourceAvailable(t *testing.T) {
	tests := []struct {
		name      string
		available bool
	}{
		{"resources available", true},
		{"resources unavailable", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &mockExtractedResourceMonitor{
				isAvailFn: func(req extractedbackground.ResourceRequirements) bool {
					return tt.available
				},
			}
			adapter := NewResourceMonitorAdapter(monitor)

			result := adapter.IsResourceAvailable(internalbackground.ResourceRequirements{
				CPUCores: 2,
				MemoryMB: 1024,
			})
			assert.Equal(t, tt.available, result)
		})
	}
}

// =============================================================================
// StuckDetectorAdapter Tests
// =============================================================================

func TestNewStuckDetectorAdapter(t *testing.T) {
	t.Run("with generic detector", func(t *testing.T) {
		detector := &mockExtractedStuckDetector{}
		adapter := NewStuckDetectorAdapter(detector)

		require.NotNil(t, adapter)
		assert.Equal(t, detector, adapter.detector)
		assert.Nil(t, adapter.concreteDetector)
	})
}

func TestStuckDetectorAdapter_IsStuck(t *testing.T) {
	t.Run("not stuck", func(t *testing.T) {
		detector := &mockExtractedStuckDetector{}
		adapter := NewStuckDetectorAdapter(detector)

		task := &internalmodels.BackgroundTask{ID: "task-1", TaskType: "test"}
		isStuck, reason := adapter.IsStuck(context.Background(), task, nil)
		assert.False(t, isStuck)
		assert.Empty(t, reason)
	})

	t.Run("stuck", func(t *testing.T) {
		detector := &mockExtractedStuckDetector{
			isStuckFn: func(ctx context.Context, task *extractedmodels.BackgroundTask, snapshots []*extractedmodels.ResourceSnapshot) (bool, string) {
				return true, "no heartbeat"
			},
		}
		adapter := NewStuckDetectorAdapter(detector)

		task := &internalmodels.BackgroundTask{ID: "task-1", TaskType: "test"}
		isStuck, reason := adapter.IsStuck(context.Background(), task, nil)
		assert.True(t, isStuck)
		assert.Equal(t, "no heartbeat", reason)
	})
}

func TestStuckDetectorAdapter_GetStuckThreshold(t *testing.T) {
	detector := &mockExtractedStuckDetector{}
	adapter := NewStuckDetectorAdapter(detector)

	threshold := adapter.GetStuckThreshold("test-type")
	assert.Equal(t, 5*time.Minute, threshold)
}

func TestStuckDetectorAdapter_SetThreshold(t *testing.T) {
	var setType string
	var setThreshold time.Duration

	detector := &mockExtractedStuckDetector{
		setThresholdFn: func(taskType string, threshold time.Duration) {
			setType = taskType
			setThreshold = threshold
		},
	}
	adapter := NewStuckDetectorAdapter(detector)

	adapter.SetThreshold("test-type", 10*time.Minute)
	assert.Equal(t, "test-type", setType)
	assert.Equal(t, 10*time.Minute, setThreshold)
}

func TestStuckDetectorAdapter_AnalyzeTask_FallbackPath(t *testing.T) {
	detector := &mockExtractedStuckDetector{}
	adapter := NewStuckDetectorAdapter(detector)

	// Without a concrete detector, falls back to basic analysis
	task := &internalmodels.BackgroundTask{ID: "task-1", TaskType: "test"}
	analysis := adapter.AnalyzeTask(context.Background(), task, nil)

	require.NotNil(t, analysis)
	assert.False(t, analysis.IsStuck)
}

// =============================================================================
// NotificationServiceAdapter Tests
// =============================================================================

func TestNewNotificationServiceAdapter(t *testing.T) {
	notifier := &mockInternalNotificationService{}
	adapter := NewNotificationServiceAdapter(notifier)

	require.NotNil(t, adapter)
	assert.Equal(t, notifier, adapter.notifier)
}

func TestNotificationServiceAdapter_NotifyTaskEvent(t *testing.T) {
	notifier := &mockInternalNotificationService{}
	adapter := NewNotificationServiceAdapter(notifier)

	task := &extractedmodels.BackgroundTask{ID: "task-1"}
	data := map[string]interface{}{"key": "value"}
	err := adapter.NotifyTaskEvent(context.Background(), task, "started", data)
	assert.NoError(t, err)
}

func TestNotificationServiceAdapter_RegisterSSEClient(t *testing.T) {
	notifier := &mockInternalNotificationService{}
	adapter := NewNotificationServiceAdapter(notifier)

	ch := make(chan<- []byte)
	err := adapter.RegisterSSEClient(context.Background(), "task-1", ch)
	assert.NoError(t, err)
}

func TestNotificationServiceAdapter_UnregisterSSEClient(t *testing.T) {
	notifier := &mockInternalNotificationService{}
	adapter := NewNotificationServiceAdapter(notifier)

	ch := make(chan<- []byte)
	err := adapter.UnregisterSSEClient(context.Background(), "task-1", ch)
	assert.NoError(t, err)
}

func TestNotificationServiceAdapter_RegisterWebSocketClient(t *testing.T) {
	notifier := &mockInternalNotificationService{}
	adapter := NewNotificationServiceAdapter(notifier)

	client := &mockExtractedWebSocketClient{}
	err := adapter.RegisterWebSocketClient(context.Background(), "task-1", client)
	assert.NoError(t, err)
}

func TestNotificationServiceAdapter_BroadcastToTask(t *testing.T) {
	notifier := &mockInternalNotificationService{}
	adapter := NewNotificationServiceAdapter(notifier)

	err := adapter.BroadcastToTask(context.Background(), "task-1", []byte("hello"))
	assert.NoError(t, err)
}

// =============================================================================
// WebSocketClientAdapter Tests
// =============================================================================

func TestNewWebSocketClientAdapter(t *testing.T) {
	client := &mockExtractedWebSocketClient{}
	adapter := NewWebSocketClientAdapter(client)

	require.NotNil(t, adapter)
	assert.Equal(t, client, adapter.client)
}

func TestWebSocketClientAdapter_Send(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &mockExtractedWebSocketClient{}
		adapter := NewWebSocketClientAdapter(client)

		err := adapter.Send([]byte("test data"))
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("send failed")
		client := &mockExtractedWebSocketClient{
			sendFn: func(data []byte) error { return expectedErr },
		}
		adapter := NewWebSocketClientAdapter(client)

		err := adapter.Send([]byte("test data"))
		assert.Error(t, err)
	})
}

func TestWebSocketClientAdapter_Close(t *testing.T) {
	client := &mockExtractedWebSocketClient{}
	adapter := NewWebSocketClientAdapter(client)

	err := adapter.Close()
	assert.NoError(t, err)
}

func TestWebSocketClientAdapter_ID(t *testing.T) {
	client := &mockExtractedWebSocketClient{
		idFn: func() string { return "custom-id" },
	}
	adapter := NewWebSocketClientAdapter(client)

	assert.Equal(t, "custom-id", adapter.ID())
}

// =============================================================================
// NoOpNotificationService Tests
// =============================================================================

func TestNoOpNotificationService(t *testing.T) {
	noop := &NoOpNotificationService{}

	task := &extractedmodels.BackgroundTask{ID: "task-1"}
	ctx := context.Background()

	assert.NoError(t, noop.NotifyTaskEvent(ctx, task, "event", nil))
	assert.NoError(t, noop.RegisterSSEClient(ctx, "task-1", nil))
	assert.NoError(t, noop.UnregisterSSEClient(ctx, "task-1", nil))
	assert.NoError(t, noop.RegisterWebSocketClient(ctx, "task-1", nil))
	assert.NoError(t, noop.BroadcastToTask(ctx, "task-1", nil))
}
