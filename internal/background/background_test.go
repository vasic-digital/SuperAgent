package background

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

// TestInMemoryTaskQueue tests the in-memory task queue implementation
func TestInMemoryTaskQueue_Enqueue(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	task := &models.BackgroundTask{
		TaskType:  "test_task",
		TaskName:  "Test Task",
		Priority:  models.TaskPriorityNormal,
	}

	err := queue.Enqueue(context.Background(), task)
	require.NoError(t, err)

	assert.NotEmpty(t, task.ID)
	assert.Equal(t, models.TaskStatusPending, task.Status)
	assert.NotZero(t, task.CreatedAt)
}

func TestInMemoryTaskQueue_EnqueueWithID(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	task := &models.BackgroundTask{
		ID:        "custom-id-123",
		TaskType:  "test_task",
		TaskName:  "Test Task",
		Priority:  models.TaskPriorityHigh,
	}

	err := queue.Enqueue(context.Background(), task)
	require.NoError(t, err)

	assert.Equal(t, "custom-id-123", task.ID)
}

func TestInMemoryTaskQueue_Dequeue(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	// Enqueue a task
	task := &models.BackgroundTask{
		TaskType:    "test_task",
		TaskName:    "Test Task",
		Priority:    models.TaskPriorityNormal,
		ScheduledAt: time.Now().Add(-time.Second), // Already scheduled
	}
	err := queue.Enqueue(context.Background(), task)
	require.NoError(t, err)

	// Dequeue the task
	dequeued, err := queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{})
	require.NoError(t, err)
	require.NotNil(t, dequeued)

	assert.Equal(t, task.ID, dequeued.ID)
	assert.Equal(t, models.TaskStatusRunning, dequeued.Status)
	assert.NotNil(t, dequeued.WorkerID)
	assert.Equal(t, "worker-1", *dequeued.WorkerID)
	assert.NotNil(t, dequeued.StartedAt)
}

func TestInMemoryTaskQueue_DequeueEmptyQueue(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	// Dequeue from empty queue
	dequeued, err := queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{})
	require.NoError(t, err)
	assert.Nil(t, dequeued)
}

func TestInMemoryTaskQueue_DequeueWithResourceRequirements(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	// Enqueue a task that requires 4 CPU cores
	task := &models.BackgroundTask{
		TaskType:         "heavy_task",
		TaskName:         "Heavy Task",
		Priority:         models.TaskPriorityNormal,
		RequiredCPUCores: 4,
		RequiredMemoryMB: 1024,
		ScheduledAt:      time.Now().Add(-time.Second),
	}
	err := queue.Enqueue(context.Background(), task)
	require.NoError(t, err)

	// Try to dequeue with insufficient resources
	dequeued, err := queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{
		CPUCores: 2, // Not enough
		MemoryMB: 512,
	})
	require.NoError(t, err)
	assert.Nil(t, dequeued, "Should not dequeue task requiring more resources")

	// Dequeue with sufficient resources
	dequeued, err = queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{
		CPUCores: 8,
		MemoryMB: 2048,
	})
	require.NoError(t, err)
	require.NotNil(t, dequeued)
	assert.Equal(t, task.ID, dequeued.ID)
}

func TestInMemoryTaskQueue_Peek(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	// Enqueue multiple tasks
	for i := 0; i < 5; i++ {
		task := &models.BackgroundTask{
			TaskType:    "test_task",
			TaskName:    "Test Task",
			Priority:    models.TaskPriorityNormal,
			ScheduledAt: time.Now().Add(-time.Second),
		}
		err := queue.Enqueue(context.Background(), task)
		require.NoError(t, err)
	}

	// Peek at tasks
	peeked, err := queue.Peek(context.Background(), 3)
	require.NoError(t, err)
	assert.Len(t, peeked, 3)

	// Verify tasks are still pending (not claimed)
	for _, task := range peeked {
		assert.Equal(t, models.TaskStatusPending, task.Status)
	}
}

func TestInMemoryTaskQueue_Requeue(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	// Enqueue and dequeue a task
	task := &models.BackgroundTask{
		TaskType:    "test_task",
		TaskName:    "Test Task",
		Priority:    models.TaskPriorityNormal,
		ScheduledAt: time.Now().Add(-time.Second),
	}
	err := queue.Enqueue(context.Background(), task)
	require.NoError(t, err)

	dequeued, err := queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{})
	require.NoError(t, err)
	require.NotNil(t, dequeued)

	// Requeue the task
	err = queue.Requeue(context.Background(), dequeued.ID, time.Second*5)
	require.NoError(t, err)

	// Verify task is pending again
	requeuedTask := queue.GetTask(dequeued.ID)
	require.NotNil(t, requeuedTask)
	assert.Equal(t, models.TaskStatusPending, requeuedTask.Status)
	assert.Nil(t, requeuedTask.WorkerID)
	assert.Equal(t, 1, requeuedTask.RetryCount)
}

func TestInMemoryTaskQueue_RequeueNonExistent(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	err := queue.Requeue(context.Background(), "non-existent-id", time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestInMemoryTaskQueue_MoveToDeadLetter(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	task := &models.BackgroundTask{
		TaskType:    "test_task",
		TaskName:    "Test Task",
		Priority:    models.TaskPriorityNormal,
		ScheduledAt: time.Now().Add(-time.Second),
	}
	err := queue.Enqueue(context.Background(), task)
	require.NoError(t, err)

	// Move to dead letter
	err = queue.MoveToDeadLetter(context.Background(), task.ID, "max retries exceeded")
	require.NoError(t, err)

	// Verify status
	deadTask := queue.GetTask(task.ID)
	require.NotNil(t, deadTask)
	assert.Equal(t, models.TaskStatusDeadLetter, deadTask.Status)
	assert.NotNil(t, deadTask.LastError)
	assert.Equal(t, "max retries exceeded", *deadTask.LastError)
}

func TestInMemoryTaskQueue_GetPendingCount(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	// Enqueue 3 tasks
	for i := 0; i < 3; i++ {
		task := &models.BackgroundTask{
			TaskType:    "test_task",
			TaskName:    "Test Task",
			Priority:    models.TaskPriorityNormal,
			ScheduledAt: time.Now().Add(-time.Second),
		}
		err := queue.Enqueue(context.Background(), task)
		require.NoError(t, err)
	}

	count, err := queue.GetPendingCount(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestInMemoryTaskQueue_GetRunningCount(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	// Enqueue and dequeue 2 tasks
	for i := 0; i < 2; i++ {
		task := &models.BackgroundTask{
			TaskType:    "test_task",
			TaskName:    "Test Task",
			Priority:    models.TaskPriorityNormal,
			ScheduledAt: time.Now().Add(-time.Second),
		}
		err := queue.Enqueue(context.Background(), task)
		require.NoError(t, err)

		_, err = queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{})
		require.NoError(t, err)
	}

	count, err := queue.GetRunningCount(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestInMemoryTaskQueue_GetQueueDepth(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	// Enqueue tasks with different priorities
	priorities := []models.TaskPriority{
		models.TaskPriorityCritical,
		models.TaskPriorityHigh,
		models.TaskPriorityHigh,
		models.TaskPriorityNormal,
		models.TaskPriorityNormal,
		models.TaskPriorityNormal,
		models.TaskPriorityLow,
	}

	for _, priority := range priorities {
		task := &models.BackgroundTask{
			TaskType:    "test_task",
			TaskName:    "Test Task",
			Priority:    priority,
			ScheduledAt: time.Now().Add(-time.Second),
		}
		err := queue.Enqueue(context.Background(), task)
		require.NoError(t, err)
	}

	depth, err := queue.GetQueueDepth(context.Background())
	require.NoError(t, err)

	assert.Equal(t, int64(1), depth[models.TaskPriorityCritical])
	assert.Equal(t, int64(2), depth[models.TaskPriorityHigh])
	assert.Equal(t, int64(3), depth[models.TaskPriorityNormal])
	assert.Equal(t, int64(1), depth[models.TaskPriorityLow])
}

func TestInMemoryTaskQueue_PriorityOrdering(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	// Enqueue tasks with different priorities (low first)
	task1 := &models.BackgroundTask{
		TaskType:    "test_task",
		TaskName:    "Low Priority",
		Priority:    models.TaskPriorityLow,
		ScheduledAt: time.Now().Add(-time.Second),
	}
	task2 := &models.BackgroundTask{
		TaskType:    "test_task",
		TaskName:    "Critical Priority",
		Priority:    models.TaskPriorityCritical,
		ScheduledAt: time.Now().Add(-time.Second),
	}
	task3 := &models.BackgroundTask{
		TaskType:    "test_task",
		TaskName:    "High Priority",
		Priority:    models.TaskPriorityHigh,
		ScheduledAt: time.Now().Add(-time.Second),
	}

	queue.Enqueue(context.Background(), task1)
	queue.Enqueue(context.Background(), task2)
	queue.Enqueue(context.Background(), task3)

	// Dequeue should return critical first
	dequeued, _ := queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{})
	assert.Equal(t, models.TaskPriorityCritical, dequeued.Priority)

	// Then high
	dequeued, _ = queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{})
	assert.Equal(t, models.TaskPriorityHigh, dequeued.Priority)

	// Then low
	dequeued, _ = queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{})
	assert.Equal(t, models.TaskPriorityLow, dequeued.Priority)
}

// TestDefaultStuckDetector tests
func TestDefaultStuckDetector_NewDefaultStuckDetector(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	assert.NotNil(t, detector)
	assert.NotNil(t, detector.thresholds)
	assert.Equal(t, 5*time.Minute, detector.GetStuckThreshold("default"))
	assert.Equal(t, 3*time.Minute, detector.GetStuckThreshold("command"))
	assert.Equal(t, 10*time.Minute, detector.GetStuckThreshold("debate"))
	assert.Equal(t, time.Duration(0), detector.GetStuckThreshold("endless"))
}

func TestDefaultStuckDetector_GetStuckThreshold(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	// Known task types
	assert.Equal(t, 3*time.Minute, detector.GetStuckThreshold("command"))
	assert.Equal(t, 3*time.Minute, detector.GetStuckThreshold("llm_call"))
	assert.Equal(t, 2*time.Minute, detector.GetStuckThreshold("embedding"))

	// Unknown task type should return default
	assert.Equal(t, 5*time.Minute, detector.GetStuckThreshold("unknown_type"))
}

func TestDefaultStuckDetector_SetThreshold(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	// Set custom threshold
	detector.SetThreshold("custom_task", 15*time.Minute)

	assert.Equal(t, 15*time.Minute, detector.GetStuckThreshold("custom_task"))
}

func TestDefaultStuckDetector_IsStuck_NilTask(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	isStuck, reason := detector.IsStuck(context.Background(), nil, nil)

	assert.False(t, isStuck)
	assert.Empty(t, reason)
}

func TestDefaultStuckDetector_IsStuck_StaleHeartbeat(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	staleTime := time.Now().Add(-10 * time.Minute)
	task := &models.BackgroundTask{
		ID:            "task-1",
		TaskType:      "command",
		Status:        models.TaskStatusRunning,
		LastHeartbeat: &staleTime,
	}

	isStuck, reason := detector.IsStuck(context.Background(), task, nil)

	assert.True(t, isStuck)
	assert.Contains(t, reason, "no heartbeat")
}

func TestDefaultStuckDetector_IsStuck_TaskOverdue(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	recentTime := time.Now()
	deadline := time.Now().Add(-time.Hour) // Passed deadline
	task := &models.BackgroundTask{
		ID:            "task-1",
		TaskType:      "command",
		Status:        models.TaskStatusRunning,
		LastHeartbeat: &recentTime,
		Deadline:      &deadline,
	}

	isStuck, reason := detector.IsStuck(context.Background(), task, nil)

	assert.True(t, isStuck)
	assert.Contains(t, reason, "exceeded deadline")
}

func TestDefaultStuckDetector_IsStuck_ProcessFrozen(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	recentTime := time.Now()
	task := &models.BackgroundTask{
		ID:            "task-1",
		TaskType:      "command",
		Status:        models.TaskStatusRunning,
		LastHeartbeat: &recentTime,
	}

	// Create snapshots showing frozen process
	snapshots := []*models.ResourceSnapshot{
		{TaskID: "task-1", CPUPercent: 0.0, CPUUserTime: 1.0, CPUSystemTime: 0.5},
		{TaskID: "task-1", CPUPercent: 0.0, CPUUserTime: 1.0, CPUSystemTime: 0.5},
		{TaskID: "task-1", CPUPercent: 0.0, CPUUserTime: 1.0, CPUSystemTime: 0.5},
		{TaskID: "task-1", CPUPercent: 0.0, CPUUserTime: 1.0, CPUSystemTime: 0.5},
	}

	isStuck, reason := detector.IsStuck(context.Background(), task, snapshots)

	assert.True(t, isStuck)
	assert.Contains(t, reason, "frozen")
}

func TestDefaultStuckDetector_IsStuck_MemoryExhaustion(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	recentTime := time.Now()
	task := &models.BackgroundTask{
		ID:            "task-1",
		TaskType:      "command",
		Status:        models.TaskStatusRunning,
		LastHeartbeat: &recentTime,
	}

	// Create snapshots showing memory exhaustion
	snapshots := []*models.ResourceSnapshot{
		{TaskID: "task-1", CPUPercent: 50.0, MemoryPercent: 96.0, CPUUserTime: 10.0, CPUSystemTime: 5.0},
		{TaskID: "task-1", CPUPercent: 50.0, MemoryPercent: 95.0, CPUUserTime: 9.0, CPUSystemTime: 4.5},
		{TaskID: "task-1", CPUPercent: 50.0, MemoryPercent: 94.0, CPUUserTime: 8.0, CPUSystemTime: 4.0},
	}

	isStuck, reason := detector.IsStuck(context.Background(), task, snapshots)

	assert.True(t, isStuck)
	assert.Contains(t, reason, "memory exhaustion")
}

func TestDefaultStuckDetector_IsStuck_FileDescriptorExhaustion(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	recentTime := time.Now()
	task := &models.BackgroundTask{
		ID:            "task-1",
		TaskType:      "command",
		Status:        models.TaskStatusRunning,
		LastHeartbeat: &recentTime,
	}

	// Create snapshots showing FD exhaustion
	snapshots := []*models.ResourceSnapshot{
		{TaskID: "task-1", CPUPercent: 50.0, OpenFDs: 15000, CPUUserTime: 10.0, CPUSystemTime: 5.0},
		{TaskID: "task-1", CPUPercent: 50.0, OpenFDs: 14000, CPUUserTime: 9.0, CPUSystemTime: 4.5},
		{TaskID: "task-1", CPUPercent: 50.0, OpenFDs: 13000, CPUUserTime: 8.0, CPUSystemTime: 4.0},
	}

	isStuck, reason := detector.IsStuck(context.Background(), task, snapshots)

	assert.True(t, isStuck)
	assert.Contains(t, reason, "file descriptor")
}

func TestDefaultStuckDetector_IsStuck_EndlessTask(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	recentTime := time.Now()
	task := &models.BackgroundTask{
		ID:            "task-1",
		TaskType:      "endless",
		Status:        models.TaskStatusRunning,
		LastHeartbeat: &recentTime,
		Config:        models.TaskConfig{Endless: true},
	}

	// Even with stale heartbeat, endless tasks use different detection
	staleTime := time.Now().Add(-time.Hour)
	task.LastHeartbeat = &staleTime

	isStuck, reason := detector.IsStuck(context.Background(), task, nil)

	assert.False(t, isStuck, "Endless tasks should not be stuck based on heartbeat alone")
	assert.Empty(t, reason)
}

func TestDefaultStuckDetector_IsStuck_EndlessTaskZombie(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	recentTime := time.Now()
	task := &models.BackgroundTask{
		ID:            "task-1",
		TaskType:      "endless",
		Status:        models.TaskStatusRunning,
		LastHeartbeat: &recentTime,
		Config:        models.TaskConfig{Endless: true},
	}

	snapshots := []*models.ResourceSnapshot{
		{TaskID: "task-1", ProcessState: "zombie"},
	}

	isStuck, reason := detector.IsStuck(context.Background(), task, snapshots)

	assert.True(t, isStuck)
	assert.Contains(t, reason, "zombie")
}

func TestDefaultStuckDetector_AnalyzeTask(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	recentTime := time.Now()
	task := &models.BackgroundTask{
		ID:            "task-1",
		TaskType:      "command",
		Status:        models.TaskStatusRunning,
		LastHeartbeat: &recentTime,
	}

	snapshots := []*models.ResourceSnapshot{
		{
			TaskID:         "task-1",
			CPUPercent:     25.0,
			MemoryPercent:  50.0,
			MemoryRSSBytes: 1024 * 1024 * 100,
			OpenFDs:        100,
			ThreadCount:    10,
			IOReadBytes:    1000,
			IOWriteBytes:   500,
			NetConnections: 5,
		},
		{
			TaskID:         "task-1",
			CPUPercent:     20.0,
			MemoryPercent:  48.0,
			MemoryRSSBytes: 1024 * 1024 * 98,
			OpenFDs:        95,
			ThreadCount:    10,
			IOReadBytes:    800,
			IOWriteBytes:   400,
			NetConnections: 5,
		},
	}

	analysis := detector.AnalyzeTask(context.Background(), task, snapshots)

	assert.NotNil(t, analysis)
	assert.False(t, analysis.IsStuck)
	assert.Equal(t, 25.0, analysis.ResourceStatus.CPUPercent)
	assert.Equal(t, 50.0, analysis.ResourceStatus.MemoryPercent)
	assert.True(t, analysis.ActivityStatus.HasCPUActivity)
	assert.True(t, analysis.ActivityStatus.HasIOActivity)
}

func TestDefaultStuckDetector_AnalyzeTask_StuckTask(t *testing.T) {
	logger := logrus.New()
	detector := NewDefaultStuckDetector(logger)

	staleTime := time.Now().Add(-10 * time.Minute)
	task := &models.BackgroundTask{
		ID:            "task-1",
		TaskType:      "command",
		Status:        models.TaskStatusRunning,
		LastHeartbeat: &staleTime,
	}

	analysis := detector.AnalyzeTask(context.Background(), task, nil)

	assert.NotNil(t, analysis)
	assert.True(t, analysis.IsStuck)
	assert.NotEmpty(t, analysis.Reason)
	assert.True(t, analysis.HeartbeatStatus.IsStale)
	assert.Contains(t, analysis.Recommendations, "Task may need to be cancelled and restarted")
}

// Test utility functions
func TestMin3(t *testing.T) {
	assert.Equal(t, 3, min3(3, 5))
	assert.Equal(t, 3, min3(5, 3))
	assert.Equal(t, 0, min3(0, 5))
	assert.Equal(t, -1, min3(-1, 5))
}

// TestResourceRequirements tests
func TestResourceRequirements_Default(t *testing.T) {
	req := ResourceRequirements{}

	assert.Equal(t, 0, req.CPUCores)
	assert.Equal(t, 0, req.MemoryMB)
	assert.Equal(t, 0, req.DiskMB)
	assert.Equal(t, 0, req.GPUCount)
}

func TestResourceRequirements_WithValues(t *testing.T) {
	req := ResourceRequirements{
		CPUCores: 4,
		MemoryMB: 8192,
		DiskMB:   10240,
		GPUCount: 1,
		Priority: models.TaskPriorityHigh,
	}

	assert.Equal(t, 4, req.CPUCores)
	assert.Equal(t, 8192, req.MemoryMB)
	assert.Equal(t, 10240, req.DiskMB)
	assert.Equal(t, 1, req.GPUCount)
	assert.Equal(t, models.TaskPriorityHigh, req.Priority)
}

// TestSystemResources tests
func TestSystemResources_Default(t *testing.T) {
	res := SystemResources{}

	assert.Equal(t, 0, res.TotalCPUCores)
	assert.Equal(t, float64(0), res.AvailableCPUCores)
	assert.Equal(t, int64(0), res.TotalMemoryMB)
}

func TestSystemResources_WithValues(t *testing.T) {
	res := SystemResources{
		TotalCPUCores:     8,
		AvailableCPUCores: 6.5,
		TotalMemoryMB:     16384,
		AvailableMemoryMB: 12000,
		CPULoadPercent:    18.75,
		MemoryUsedPercent: 26.76,
		DiskUsedPercent:   45.0,
		LoadAvg1:          1.5,
		LoadAvg5:          2.0,
		LoadAvg15:         1.8,
	}

	assert.Equal(t, 8, res.TotalCPUCores)
	assert.Equal(t, 6.5, res.AvailableCPUCores)
	assert.Equal(t, int64(16384), res.TotalMemoryMB)
	assert.Equal(t, int64(12000), res.AvailableMemoryMB)
	assert.Equal(t, 18.75, res.CPULoadPercent)
}

// TestWorkerStatus tests
func TestWorkerStatus_Default(t *testing.T) {
	status := WorkerStatus{}

	assert.Empty(t, status.ID)
	assert.Empty(t, status.Status)
	assert.Nil(t, status.CurrentTask)
}

func TestWorkerStatus_WithValues(t *testing.T) {
	now := time.Now()
	status := WorkerStatus{
		ID:              "worker-1",
		Status:          "busy",
		StartedAt:       now.Add(-time.Hour),
		LastActivity:    now,
		TasksCompleted:  100,
		TasksFailed:     5,
		AvgTaskDuration: 30 * time.Second,
	}

	assert.Equal(t, "worker-1", status.ID)
	assert.Equal(t, "busy", status.Status)
	assert.Equal(t, int64(100), status.TasksCompleted)
	assert.Equal(t, int64(5), status.TasksFailed)
	assert.Equal(t, 30*time.Second, status.AvgTaskDuration)
}

// TestTaskEvent tests
func TestTaskEvent_Creation(t *testing.T) {
	workerID := "worker-1"
	event := TaskEvent{
		TaskID:    "task-1",
		EventType: "started",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"progress": 0.5,
		},
		WorkerID: &workerID,
	}

	assert.Equal(t, "task-1", event.TaskID)
	assert.Equal(t, "started", event.EventType)
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, 0.5, event.Data["progress"])
	assert.Equal(t, "worker-1", *event.WorkerID)
}

// TestExecutionResult tests
func TestExecutionResult_Success(t *testing.T) {
	result := ExecutionResult{
		TaskID:     "task-1",
		Status:     models.TaskStatusCompleted,
		Output:     []byte("success output"),
		Duration:   5 * time.Second,
		RetryCount: 0,
	}

	assert.Equal(t, "task-1", result.TaskID)
	assert.Equal(t, models.TaskStatusCompleted, result.Status)
	assert.Equal(t, "success output", string(result.Output))
	assert.Equal(t, 5*time.Second, result.Duration)
	assert.Equal(t, 0, result.RetryCount)
	assert.Empty(t, result.Error)
}

func TestExecutionResult_Failure(t *testing.T) {
	result := ExecutionResult{
		TaskID:     "task-1",
		Status:     models.TaskStatusFailed,
		Error:      "execution failed: timeout",
		Duration:   30 * time.Second,
		RetryCount: 3,
	}

	assert.Equal(t, models.TaskStatusFailed, result.Status)
	assert.Equal(t, "execution failed: timeout", result.Error)
	assert.Equal(t, 3, result.RetryCount)
}

// TestTaskQueueStats tests
func TestTaskQueueStats_MarshalJSON(t *testing.T) {
	stats := &TaskQueueStats{
		PendingCount: 10,
		RunningCount: 5,
		DepthByPriority: map[models.TaskPriority]int64{
			models.TaskPriorityCritical: 2,
			models.TaskPriorityHigh:     3,
			models.TaskPriorityNormal:   5,
		},
		StatusCounts: map[models.TaskStatus]int64{
			models.TaskStatusPending:   10,
			models.TaskStatusRunning:   5,
			models.TaskStatusCompleted: 100,
		},
	}

	data, err := stats.MarshalJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), `"pending_count":10`)
	assert.Contains(t, string(data), `"running_count":5`)
}

// TestDefaultStuckDetectorConfig tests
func TestDefaultStuckDetectorConfig(t *testing.T) {
	config := DefaultStuckDetectorConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 5*time.Minute, config.DefaultThreshold)
	assert.Equal(t, 0.1, config.CPUActivityThreshold)
	assert.Equal(t, 0.5, config.MemoryGrowthThreshold)
	assert.Equal(t, int64(1024), config.IOActivityThreshold)
	assert.Equal(t, 3, config.MinSnapshotsForAnalysis)
}

// TestStuckAnalysis tests
func TestStuckAnalysis_NotStuck(t *testing.T) {
	analysis := &StuckAnalysis{
		IsStuck:         false,
		Recommendations: []string{},
		HeartbeatStatus: HeartbeatStatus{
			IsStale: false,
		},
		ResourceStatus: ResourceStatus{
			IsExhausted: false,
		},
		ActivityStatus: ActivityStatus{
			HasCPUActivity: true,
			HasIOActivity:  true,
		},
	}

	assert.False(t, analysis.IsStuck)
	assert.Empty(t, analysis.Recommendations)
	assert.False(t, analysis.HeartbeatStatus.IsStale)
	assert.True(t, analysis.ActivityStatus.HasCPUActivity)
}

func TestStuckAnalysis_Stuck(t *testing.T) {
	analysis := &StuckAnalysis{
		IsStuck: true,
		Reason:  "no heartbeat for 10m0s",
		Recommendations: []string{
			"Task may need to be cancelled and restarted",
		},
		HeartbeatStatus: HeartbeatStatus{
			IsStale:   true,
			Threshold: 5 * time.Minute,
		},
	}

	assert.True(t, analysis.IsStuck)
	assert.Contains(t, analysis.Reason, "no heartbeat")
	assert.Len(t, analysis.Recommendations, 1)
	assert.True(t, analysis.HeartbeatStatus.IsStale)
}

// TestWaitResult tests
func TestWaitResult_Success(t *testing.T) {
	task := &models.BackgroundTask{
		ID:     "task-1",
		Status: models.TaskStatusCompleted,
	}
	result := WaitResult{
		Task:     task,
		Output:   []byte("output data"),
		Duration: 5 * time.Second,
		Error:    nil,
	}

	assert.NotNil(t, result.Task)
	assert.Equal(t, models.TaskStatusCompleted, result.Task.Status)
	assert.Equal(t, "output data", string(result.Output))
	assert.Equal(t, 5*time.Second, result.Duration)
	assert.Nil(t, result.Error)
}

// Concurrent access tests
func TestInMemoryTaskQueue_ConcurrentEnqueue(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	const numTasks = 100
	done := make(chan struct{})

	for i := 0; i < numTasks; i++ {
		go func(idx int) {
			defer func() { done <- struct{}{} }()
			task := &models.BackgroundTask{
				TaskType:    "test_task",
				TaskName:    "Test Task",
				Priority:    models.TaskPriorityNormal,
				ScheduledAt: time.Now().Add(-time.Second),
			}
			err := queue.Enqueue(context.Background(), task)
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numTasks; i++ {
		<-done
	}

	count, err := queue.GetPendingCount(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(numTasks), count)
}

func TestInMemoryTaskQueue_ConcurrentDequeue(t *testing.T) {
	logger := logrus.New()
	queue := NewInMemoryTaskQueue(logger)

	const numTasks = 50
	// Enqueue tasks first
	for i := 0; i < numTasks; i++ {
		task := &models.BackgroundTask{
			TaskType:    "test_task",
			TaskName:    "Test Task",
			Priority:    models.TaskPriorityNormal,
			ScheduledAt: time.Now().Add(-time.Second),
		}
		err := queue.Enqueue(context.Background(), task)
		require.NoError(t, err)
	}

	// Concurrent dequeue
	const numWorkers = 10
	results := make(chan *models.BackgroundTask, numTasks)
	done := make(chan struct{})

	for w := 0; w < numWorkers; w++ {
		go func(workerID int) {
			defer func() { done <- struct{}{} }()
			for {
				task, err := queue.Dequeue(context.Background(), "worker-"+string(rune(workerID)), ResourceRequirements{})
				if err != nil || task == nil {
					return
				}
				results <- task
			}
		}(w)
	}

	// Wait for workers
	for i := 0; i < numWorkers; i++ {
		<-done
	}
	close(results)

	// Count dequeued tasks - each task should be dequeued exactly once
	dequeued := make(map[string]bool)
	for task := range results {
		assert.False(t, dequeued[task.ID], "Task %s was dequeued twice", task.ID)
		dequeued[task.ID] = true
	}

	assert.Equal(t, numTasks, len(dequeued))
}
