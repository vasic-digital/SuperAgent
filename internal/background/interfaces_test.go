// Package background provides comprehensive tests for interfaces.go —
// compile-time interface compliance checks, struct zero values and defaults,
// WaitResult, ResourceRequirements, SystemResources, WorkerStatus,
// TaskEvent, ExecutionResult types.
package background

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"dev.helix.agent/internal/models"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// =============================================================================
// Interface compliance — compile-time checks
// =============================================================================

// Verify that the interfaces are properly defined by checking that they can
// be used as types. These are compile-time checks; if any interface signature
// is broken the test file will not compile.

func TestTaskExecutor_InterfaceDefinition(t *testing.T) {
	// Compile-time check: the interface exists and has expected methods
	var _ TaskExecutor = (*ifaceTaskExecutor)(nil)
}

func TestProgressReporter_InterfaceDefinition(t *testing.T) {
	var _ ProgressReporter = (*ifaceProgressReporter)(nil)
}

func TestTaskQueue_InterfaceDefinition(t *testing.T) {
	var _ TaskQueue = (*ifaceTaskQueue)(nil)
}

func TestTaskWaiter_InterfaceDefinition(t *testing.T) {
	var _ TaskWaiter = (*ifaceTaskWaiter)(nil)
}

func TestTaskRepository_InterfaceDefinition(t *testing.T) {
	var _ TaskRepository = (*ifaceTaskRepository)(nil)
}

func TestResourceMonitor_InterfaceDefinition(t *testing.T) {
	var _ ResourceMonitor = (*ifaceResourceMonitor)(nil)
}

func TestStuckDetector_InterfaceDefinition(t *testing.T) {
	var _ StuckDetector = (*ifaceStuckDetector)(nil)
}

func TestNotificationService_InterfaceDefinition(t *testing.T) {
	var _ NotificationService = (*ifaceNotificationService)(nil)
}

func TestWebSocketClient_InterfaceDefinition(t *testing.T) {
	var _ WebSocketClient = (*ifaceWebSocketClient)(nil)
}

func TestWorkerPool_InterfaceDefinition(t *testing.T) {
	var _ WorkerPool = (*ifaceWorkerPool)(nil)
}

// =============================================================================
// WaitResult struct
// =============================================================================

func TestWaitResult_ZeroValue(t *testing.T) {
	var wr WaitResult
	assert.Nil(t, wr.Task)
	assert.Nil(t, wr.Output)
	assert.Zero(t, wr.Duration)
	assert.Nil(t, wr.Error)
}

func TestWaitResult_FullyPopulated(t *testing.T) {
	task := &models.BackgroundTask{
		ID:       "task-123",
		TaskType: "test",
		Status:   models.TaskStatusCompleted,
	}

	wr := WaitResult{
		Task:     task,
		Output:   []byte("output data"),
		Duration: 5 * time.Second,
		Error:    nil,
	}

	assert.NotNil(t, wr.Task)
	assert.Equal(t, "task-123", wr.Task.ID)
	assert.Equal(t, []byte("output data"), wr.Output)
	assert.Equal(t, 5*time.Second, wr.Duration)
	assert.Nil(t, wr.Error)
}

func TestWaitResult_WithError_NilTask(t *testing.T) {
	wr := WaitResult{
		Task:     nil,
		Output:   nil,
		Duration: 100 * time.Millisecond,
		Error:    assert.AnError,
	}

	assert.Nil(t, wr.Task)
	assert.Nil(t, wr.Output)
	assert.Equal(t, 100*time.Millisecond, wr.Duration)
	assert.Error(t, wr.Error)
}

// =============================================================================
// ResourceRequirements struct
// =============================================================================

func TestResourceRequirements_ZeroValue(t *testing.T) {
	var rr ResourceRequirements
	assert.Zero(t, rr.CPUCores)
	assert.Zero(t, rr.MemoryMB)
	assert.Zero(t, rr.DiskMB)
	assert.Zero(t, rr.GPUCount)
	assert.Empty(t, rr.Priority)
}

func TestResourceRequirements_FullyPopulated(t *testing.T) {
	rr := ResourceRequirements{
		CPUCores: 4,
		MemoryMB: 2048,
		DiskMB:   10240,
		GPUCount: 1,
		Priority: models.TaskPriorityHigh,
	}

	assert.Equal(t, 4, rr.CPUCores)
	assert.Equal(t, 2048, rr.MemoryMB)
	assert.Equal(t, 10240, rr.DiskMB)
	assert.Equal(t, 1, rr.GPUCount)
	assert.Equal(t, models.TaskPriorityHigh, rr.Priority)
}

func TestResourceRequirements_MinimalTask(t *testing.T) {
	rr := ResourceRequirements{
		CPUCores: 1,
		MemoryMB: 256,
		Priority: models.TaskPriorityBackground,
	}

	assert.Equal(t, 1, rr.CPUCores)
	assert.Equal(t, 256, rr.MemoryMB)
	assert.Zero(t, rr.DiskMB)
	assert.Zero(t, rr.GPUCount)
	assert.Equal(t, models.TaskPriorityBackground, rr.Priority)
}

func TestResourceRequirements_AllPriorities(t *testing.T) {
	priorities := []models.TaskPriority{
		models.TaskPriorityCritical,
		models.TaskPriorityHigh,
		models.TaskPriorityNormal,
		models.TaskPriorityLow,
		models.TaskPriorityBackground,
	}

	for _, p := range priorities {
		t.Run(string(p), func(t *testing.T) {
			rr := ResourceRequirements{Priority: p}
			assert.Equal(t, p, rr.Priority)
		})
	}
}

// =============================================================================
// SystemResources struct
// =============================================================================

func TestSystemResources_ZeroValue_AllFieldsZero(t *testing.T) {
	var sr SystemResources
	assert.Zero(t, sr.TotalCPUCores)
	assert.Zero(t, sr.AvailableCPUCores)
	assert.Zero(t, sr.TotalMemoryMB)
	assert.Zero(t, sr.AvailableMemoryMB)
	assert.Zero(t, sr.CPULoadPercent)
	assert.Zero(t, sr.MemoryUsedPercent)
	assert.Zero(t, sr.DiskUsedPercent)
	assert.Zero(t, sr.LoadAvg1)
	assert.Zero(t, sr.LoadAvg5)
	assert.Zero(t, sr.LoadAvg15)
}

func TestSystemResources_FullyPopulated(t *testing.T) {
	sr := SystemResources{
		TotalCPUCores:     8,
		AvailableCPUCores: 4.5,
		TotalMemoryMB:     16384,
		AvailableMemoryMB: 8000,
		CPULoadPercent:    45.5,
		MemoryUsedPercent: 51.2,
		DiskUsedPercent:   60.0,
		LoadAvg1:          2.5,
		LoadAvg5:          1.8,
		LoadAvg15:         1.2,
	}

	assert.Equal(t, 8, sr.TotalCPUCores)
	assert.InDelta(t, 4.5, sr.AvailableCPUCores, 0.001)
	assert.Equal(t, int64(16384), sr.TotalMemoryMB)
	assert.Equal(t, int64(8000), sr.AvailableMemoryMB)
	assert.InDelta(t, 45.5, sr.CPULoadPercent, 0.001)
	assert.InDelta(t, 51.2, sr.MemoryUsedPercent, 0.001)
	assert.InDelta(t, 60.0, sr.DiskUsedPercent, 0.001)
	assert.InDelta(t, 2.5, sr.LoadAvg1, 0.001)
	assert.InDelta(t, 1.8, sr.LoadAvg5, 0.001)
	assert.InDelta(t, 1.2, sr.LoadAvg15, 0.001)
}

func TestSystemResources_HighLoad(t *testing.T) {
	sr := SystemResources{
		TotalCPUCores:     4,
		AvailableCPUCores: 0.2,
		TotalMemoryMB:     8192,
		AvailableMemoryMB: 100,
		CPULoadPercent:    98.5,
		MemoryUsedPercent: 98.8,
		DiskUsedPercent:   95.0,
		LoadAvg1:          12.0,
		LoadAvg5:          10.5,
		LoadAvg15:         9.0,
	}

	assert.InDelta(t, 98.5, sr.CPULoadPercent, 0.1)
	assert.InDelta(t, 98.8, sr.MemoryUsedPercent, 0.1)
	assert.Greater(t, sr.LoadAvg1, float64(sr.TotalCPUCores))
}

// =============================================================================
// WorkerStatus struct
// =============================================================================

func TestWorkerStatus_ZeroValue(t *testing.T) {
	var ws WorkerStatus
	assert.Empty(t, ws.ID)
	assert.Empty(t, ws.Status)
	assert.Nil(t, ws.CurrentTask)
	assert.True(t, ws.StartedAt.IsZero())
	assert.True(t, ws.LastActivity.IsZero())
	assert.Zero(t, ws.TasksCompleted)
	assert.Zero(t, ws.TasksFailed)
	assert.Zero(t, ws.AvgTaskDuration)
}

func TestWorkerStatus_Idle(t *testing.T) {
	now := time.Now()
	ws := WorkerStatus{
		ID:              "worker-1",
		Status:          "idle",
		CurrentTask:     nil,
		StartedAt:       now.Add(-1 * time.Hour),
		LastActivity:    now.Add(-5 * time.Minute),
		TasksCompleted:  50,
		TasksFailed:     2,
		AvgTaskDuration: 30 * time.Second,
	}

	assert.Equal(t, "worker-1", ws.ID)
	assert.Equal(t, "idle", ws.Status)
	assert.Nil(t, ws.CurrentTask)
	assert.Equal(t, int64(50), ws.TasksCompleted)
	assert.Equal(t, int64(2), ws.TasksFailed)
	assert.Equal(t, 30*time.Second, ws.AvgTaskDuration)
}

func TestWorkerStatus_Busy(t *testing.T) {
	now := time.Now()
	task := &models.BackgroundTask{
		ID:       "task-abc",
		TaskType: "llm_call",
		Status:   models.TaskStatusRunning,
	}

	ws := WorkerStatus{
		ID:           "worker-2",
		Status:       "busy",
		CurrentTask:  task,
		StartedAt:    now.Add(-2 * time.Hour),
		LastActivity: now,
	}

	assert.Equal(t, "busy", ws.Status)
	assert.NotNil(t, ws.CurrentTask)
	assert.Equal(t, "task-abc", ws.CurrentTask.ID)
}

func TestWorkerStatus_AllStatuses(t *testing.T) {
	statuses := []string{"idle", "busy", "stopping", "stopped"}
	for _, s := range statuses {
		t.Run(s, func(t *testing.T) {
			ws := WorkerStatus{Status: s}
			assert.Equal(t, s, ws.Status)
		})
	}
}

// =============================================================================
// TaskEvent struct
// =============================================================================

func TestTaskEvent_ZeroValue(t *testing.T) {
	var te TaskEvent
	assert.Empty(t, te.TaskID)
	assert.Empty(t, te.EventType)
	assert.True(t, te.Timestamp.IsZero())
	assert.Nil(t, te.Data)
	assert.Nil(t, te.WorkerID)
}

func TestTaskEvent_FullyPopulated(t *testing.T) {
	now := time.Now()
	workerID := "worker-1"

	te := TaskEvent{
		TaskID:    "task-123",
		EventType: "task.completed",
		Timestamp: now,
		Data: map[string]interface{}{
			"duration_ms": 5000,
			"output_size": 1024,
		},
		WorkerID: &workerID,
	}

	assert.Equal(t, "task-123", te.TaskID)
	assert.Equal(t, "task.completed", te.EventType)
	assert.Equal(t, now, te.Timestamp)
	assert.NotNil(t, te.Data)
	assert.Contains(t, te.Data, "duration_ms")
	assert.NotNil(t, te.WorkerID)
	assert.Equal(t, "worker-1", *te.WorkerID)
}

func TestTaskEvent_NoWorkerID(t *testing.T) {
	te := TaskEvent{
		TaskID:    "task-456",
		EventType: "task.created",
		Timestamp: time.Now(),
		WorkerID:  nil,
	}

	assert.Nil(t, te.WorkerID)
}

// =============================================================================
// ExecutionResult struct
// =============================================================================

func TestExecutionResult_ZeroValue(t *testing.T) {
	var er ExecutionResult
	assert.Empty(t, er.TaskID)
	assert.Empty(t, er.Status)
	assert.Nil(t, er.Output)
	assert.Empty(t, er.Error)
	assert.Zero(t, er.Duration)
	assert.Zero(t, er.RetryCount)
	assert.Nil(t, er.ResourceMetrics)
}

func TestExecutionResult_Successful(t *testing.T) {
	er := ExecutionResult{
		TaskID:     "task-789",
		Status:     models.TaskStatusCompleted,
		Output:     []byte("success output"),
		Error:      "",
		Duration:   10 * time.Second,
		RetryCount: 0,
		ResourceMetrics: &models.ResourceSnapshot{
			CPUPercent:    25.0,
			MemoryPercent: 40.0,
		},
	}

	assert.Equal(t, "task-789", er.TaskID)
	assert.Equal(t, models.TaskStatusCompleted, er.Status)
	assert.Equal(t, []byte("success output"), er.Output)
	assert.Empty(t, er.Error)
	assert.Equal(t, 10*time.Second, er.Duration)
	assert.Zero(t, er.RetryCount)
	assert.NotNil(t, er.ResourceMetrics)
	assert.InDelta(t, 25.0, er.ResourceMetrics.CPUPercent, 0.1)
}

func TestExecutionResult_Failed(t *testing.T) {
	er := ExecutionResult{
		TaskID:     "task-fail",
		Status:     models.TaskStatusFailed,
		Output:     nil,
		Error:      "connection timeout",
		Duration:   30 * time.Second,
		RetryCount: 3,
	}

	assert.Equal(t, models.TaskStatusFailed, er.Status)
	assert.Nil(t, er.Output)
	assert.Equal(t, "connection timeout", er.Error)
	assert.Equal(t, 3, er.RetryCount)
}

func TestExecutionResult_AllStatuses(t *testing.T) {
	statuses := []models.TaskStatus{
		models.TaskStatusCompleted,
		models.TaskStatusFailed,
		models.TaskStatusCancelled,
		models.TaskStatusStuck,
	}

	for _, s := range statuses {
		t.Run(string(s), func(t *testing.T) {
			er := ExecutionResult{Status: s}
			assert.Equal(t, s, er.Status)
		})
	}
}

// =============================================================================
// Stub implementations for compile-time interface checks
// (prefixed with "iface" to avoid collisions with mocks in other test files)
// =============================================================================

type ifaceTaskExecutor struct{}

func (m *ifaceTaskExecutor) Execute(ctx context.Context, task *models.BackgroundTask, reporter ProgressReporter) error {
	return nil
}
func (m *ifaceTaskExecutor) CanPause() bool { return false }
func (m *ifaceTaskExecutor) Pause(ctx context.Context, task *models.BackgroundTask) ([]byte, error) {
	return nil, nil
}
func (m *ifaceTaskExecutor) Resume(ctx context.Context, task *models.BackgroundTask, checkpoint []byte) error {
	return nil
}
func (m *ifaceTaskExecutor) Cancel(ctx context.Context, task *models.BackgroundTask) error {
	return nil
}
func (m *ifaceTaskExecutor) GetResourceRequirements() ResourceRequirements {
	return ResourceRequirements{}
}

type ifaceProgressReporter struct{}

func (m *ifaceProgressReporter) ReportProgress(percent float64, message string) error { return nil }
func (m *ifaceProgressReporter) ReportHeartbeat() error                               { return nil }
func (m *ifaceProgressReporter) ReportCheckpoint(data []byte) error                   { return nil }
func (m *ifaceProgressReporter) ReportMetrics(metrics map[string]interface{}) error    { return nil }
func (m *ifaceProgressReporter) ReportLog(level, message string, fields map[string]interface{}) error {
	return nil
}

type ifaceTaskQueue struct{}

func (m *ifaceTaskQueue) Enqueue(ctx context.Context, task *models.BackgroundTask) error { return nil }
func (m *ifaceTaskQueue) Dequeue(ctx context.Context, workerID string, requirements ResourceRequirements) (*models.BackgroundTask, error) {
	return nil, nil
}
func (m *ifaceTaskQueue) Peek(ctx context.Context, count int) ([]*models.BackgroundTask, error) {
	return nil, nil
}
func (m *ifaceTaskQueue) Requeue(ctx context.Context, taskID string, delay time.Duration) error {
	return nil
}
func (m *ifaceTaskQueue) MoveToDeadLetter(ctx context.Context, taskID string, reason string) error {
	return nil
}
func (m *ifaceTaskQueue) GetPendingCount(ctx context.Context) (int64, error)  { return 0, nil }
func (m *ifaceTaskQueue) GetRunningCount(ctx context.Context) (int64, error)  { return 0, nil }
func (m *ifaceTaskQueue) GetQueueDepth(ctx context.Context) (map[models.TaskPriority]int64, error) {
	return nil, nil
}

type ifaceTaskWaiter struct{}

func (m *ifaceTaskWaiter) WaitForCompletion(ctx context.Context, taskID string, timeout time.Duration, progressCallback func(progress float64, message string)) (*models.BackgroundTask, error) {
	return nil, nil
}
func (m *ifaceTaskWaiter) WaitForCompletionWithOutput(ctx context.Context, taskID string, timeout time.Duration) (*models.BackgroundTask, []byte, error) {
	return nil, nil, nil
}

type ifaceTaskRepository struct{}

func (m *ifaceTaskRepository) Create(ctx context.Context, task *models.BackgroundTask) error {
	return nil
}
func (m *ifaceTaskRepository) GetByID(ctx context.Context, id string) (*models.BackgroundTask, error) {
	return nil, nil
}
func (m *ifaceTaskRepository) Update(ctx context.Context, task *models.BackgroundTask) error {
	return nil
}
func (m *ifaceTaskRepository) Delete(ctx context.Context, id string) error { return nil }
func (m *ifaceTaskRepository) UpdateStatus(ctx context.Context, id string, status models.TaskStatus) error {
	return nil
}
func (m *ifaceTaskRepository) UpdateProgress(ctx context.Context, id string, progress float64, message string) error {
	return nil
}
func (m *ifaceTaskRepository) UpdateHeartbeat(ctx context.Context, id string) error { return nil }
func (m *ifaceTaskRepository) SaveCheckpoint(ctx context.Context, id string, checkpoint []byte) error {
	return nil
}
func (m *ifaceTaskRepository) GetByStatus(ctx context.Context, status models.TaskStatus, limit, offset int) ([]*models.BackgroundTask, error) {
	return nil, nil
}
func (m *ifaceTaskRepository) GetPendingTasks(ctx context.Context, limit int) ([]*models.BackgroundTask, error) {
	return nil, nil
}
func (m *ifaceTaskRepository) GetStaleTasks(ctx context.Context, threshold time.Duration) ([]*models.BackgroundTask, error) {
	return nil, nil
}
func (m *ifaceTaskRepository) GetByWorkerID(ctx context.Context, workerID string) ([]*models.BackgroundTask, error) {
	return nil, nil
}
func (m *ifaceTaskRepository) CountByStatus(ctx context.Context) (map[models.TaskStatus]int64, error) {
	return nil, nil
}
func (m *ifaceTaskRepository) Dequeue(ctx context.Context, workerID string, maxCPUCores, maxMemoryMB int) (*models.BackgroundTask, error) {
	return nil, nil
}
func (m *ifaceTaskRepository) SaveResourceSnapshot(ctx context.Context, snapshot *models.ResourceSnapshot) error {
	return nil
}
func (m *ifaceTaskRepository) GetResourceSnapshots(ctx context.Context, taskID string, limit int) ([]*models.ResourceSnapshot, error) {
	return nil, nil
}
func (m *ifaceTaskRepository) LogEvent(ctx context.Context, taskID, eventType string, data map[string]interface{}, workerID *string) error {
	return nil
}
func (m *ifaceTaskRepository) GetTaskHistory(ctx context.Context, taskID string, limit int) ([]*models.TaskExecutionHistory, error) {
	return nil, nil
}
func (m *ifaceTaskRepository) MoveToDeadLetter(ctx context.Context, taskID, reason string) error {
	return nil
}

type ifaceResourceMonitor struct{}

func (m *ifaceResourceMonitor) GetSystemResources() (*SystemResources, error) { return nil, nil }
func (m *ifaceResourceMonitor) GetProcessResources(pid int) (*models.ResourceSnapshot, error) {
	return nil, nil
}
func (m *ifaceResourceMonitor) StartMonitoring(taskID string, pid int, interval time.Duration) error {
	return nil
}
func (m *ifaceResourceMonitor) StopMonitoring(taskID string) error { return nil }
func (m *ifaceResourceMonitor) GetLatestSnapshot(taskID string) (*models.ResourceSnapshot, error) {
	return nil, nil
}
func (m *ifaceResourceMonitor) IsResourceAvailable(requirements ResourceRequirements) bool {
	return false
}

type ifaceStuckDetector struct{}

func (m *ifaceStuckDetector) IsStuck(ctx context.Context, task *models.BackgroundTask, snapshots []*models.ResourceSnapshot) (bool, string) {
	return false, ""
}
func (m *ifaceStuckDetector) GetStuckThreshold(taskType string) time.Duration { return 0 }
func (m *ifaceStuckDetector) SetThreshold(taskType string, threshold time.Duration) {}

type ifaceNotificationService struct{}

func (m *ifaceNotificationService) NotifyTaskEvent(ctx context.Context, task *models.BackgroundTask, event string, data map[string]interface{}) error {
	return nil
}
func (m *ifaceNotificationService) RegisterSSEClient(ctx context.Context, taskID string, client chan<- []byte) error {
	return nil
}
func (m *ifaceNotificationService) UnregisterSSEClient(ctx context.Context, taskID string, client chan<- []byte) error {
	return nil
}
func (m *ifaceNotificationService) RegisterWebSocketClient(ctx context.Context, taskID string, client WebSocketClient) error {
	return nil
}
func (m *ifaceNotificationService) BroadcastToTask(ctx context.Context, taskID string, message []byte) error {
	return nil
}

type ifaceWebSocketClient struct{}

func (m *ifaceWebSocketClient) Send(data []byte) error { return nil }
func (m *ifaceWebSocketClient) Close() error           { return nil }
func (m *ifaceWebSocketClient) ID() string              { return "iface-ws-client" }

type ifaceWorkerPool struct{}

func (m *ifaceWorkerPool) Start(ctx context.Context) error                        { return nil }
func (m *ifaceWorkerPool) Stop(gracePeriod time.Duration) error                   { return nil }
func (m *ifaceWorkerPool) RegisterExecutor(taskType string, executor TaskExecutor) {}
func (m *ifaceWorkerPool) GetWorkerCount() int                                     { return 0 }
func (m *ifaceWorkerPool) GetActiveTaskCount() int                                 { return 0 }
func (m *ifaceWorkerPool) GetWorkerStatus() []WorkerStatus                         { return nil }
func (m *ifaceWorkerPool) Scale(targetCount int) error                             { return nil }
