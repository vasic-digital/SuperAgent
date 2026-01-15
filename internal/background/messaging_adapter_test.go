package background

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

func TestDefaultMessagingTaskQueueConfig(t *testing.T) {
	config := DefaultMessagingTaskQueueConfig()

	require.NotNil(t, config)
	require.NotNil(t, config.PublisherConfig)
	assert.True(t, config.PublisherConfig.Enabled)
	assert.True(t, config.PublisherConfig.AsyncPublish)
	assert.Equal(t, 1000, config.PublisherConfig.BufferSize)
}

func TestNewMessagingTaskQueue(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)

	queue := NewMessagingTaskQueue(delegate, nil, logger, nil)

	require.NotNil(t, queue)
	assert.NotNil(t, queue.delegate)
	assert.NotNil(t, queue.publisher)
	assert.NotNil(t, queue.logger)
	assert.Equal(t, delegate, queue.Delegate())
}

func TestNewMessagingTaskQueue_WithConfig(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	config := &MessagingTaskQueueConfig{
		PublisherConfig: &TaskEventPublisherConfig{
			Enabled:      false,
			AsyncPublish: false,
			BufferSize:   100,
		},
	}

	queue := NewMessagingTaskQueue(delegate, nil, logger, config)

	require.NotNil(t, queue)
	assert.False(t, queue.publisher.IsEnabled())
}

func TestMessagingTaskQueue_Enqueue(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	config := &MessagingTaskQueueConfig{
		PublisherConfig: &TaskEventPublisherConfig{
			Enabled:      false, // Disable to avoid hub dependency
			AsyncPublish: false,
		},
	}

	queue := NewMessagingTaskQueue(delegate, nil, logger, config)

	task := &models.BackgroundTask{
		ID:       "test-task-1",
		TaskType: "test_type",
		TaskName: "Test Task",
		Payload:  json.RawMessage(`{"key": "value"}`),
	}

	err := queue.Enqueue(context.Background(), task)
	assert.NoError(t, err)

	// Verify task was enqueued
	count, err := queue.GetPendingCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestMessagingTaskQueue_Dequeue(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	config := &MessagingTaskQueueConfig{
		PublisherConfig: &TaskEventPublisherConfig{
			Enabled:      false,
			AsyncPublish: false,
		},
	}

	queue := NewMessagingTaskQueue(delegate, nil, logger, config)

	// Enqueue a task
	task := &models.BackgroundTask{
		ID:          "test-task-1",
		TaskType:    "test_type",
		TaskName:    "Test Task",
		ScheduledAt: time.Now().Add(-time.Second),
	}
	err := queue.Enqueue(context.Background(), task)
	require.NoError(t, err)

	// Dequeue the task
	dequeued, err := queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{})
	assert.NoError(t, err)
	require.NotNil(t, dequeued)
	assert.Equal(t, "test-task-1", dequeued.ID)
	assert.Equal(t, models.TaskStatusRunning, dequeued.Status)
}

func TestMessagingTaskQueue_Dequeue_NoTask(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	config := &MessagingTaskQueueConfig{
		PublisherConfig: &TaskEventPublisherConfig{
			Enabled:      false,
			AsyncPublish: false,
		},
	}

	queue := NewMessagingTaskQueue(delegate, nil, logger, config)

	// Dequeue from empty queue
	dequeued, err := queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{})
	assert.NoError(t, err)
	assert.Nil(t, dequeued)
}

func TestMessagingTaskQueue_Peek(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	config := &MessagingTaskQueueConfig{
		PublisherConfig: &TaskEventPublisherConfig{
			Enabled:      false,
			AsyncPublish: false,
		},
	}

	queue := NewMessagingTaskQueue(delegate, nil, logger, config)

	// Enqueue tasks
	for i := 0; i < 3; i++ {
		task := &models.BackgroundTask{
			TaskType: "test_type",
			TaskName: "Test Task",
		}
		err := queue.Enqueue(context.Background(), task)
		require.NoError(t, err)
	}

	// Peek at tasks
	tasks, err := queue.Peek(context.Background(), 2)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestMessagingTaskQueue_Requeue(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	config := &MessagingTaskQueueConfig{
		PublisherConfig: &TaskEventPublisherConfig{
			Enabled:      false,
			AsyncPublish: false,
		},
	}

	queue := NewMessagingTaskQueue(delegate, nil, logger, config)

	// Enqueue a task
	task := &models.BackgroundTask{
		ID:          "test-task-1",
		TaskType:    "test_type",
		TaskName:    "Test Task",
		ScheduledAt: time.Now().Add(-time.Second),
	}
	err := queue.Enqueue(context.Background(), task)
	require.NoError(t, err)

	// Dequeue the task
	dequeued, err := queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{})
	require.NoError(t, err)
	require.NotNil(t, dequeued)

	// Requeue the task
	err = queue.Requeue(context.Background(), dequeued.ID, 100*time.Millisecond)
	assert.NoError(t, err)

	// Verify task is back in pending state
	pendingCount, err := queue.GetPendingCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), pendingCount)
}

func TestMessagingTaskQueue_MoveToDeadLetter(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	config := &MessagingTaskQueueConfig{
		PublisherConfig: &TaskEventPublisherConfig{
			Enabled:      false,
			AsyncPublish: false,
		},
	}

	queue := NewMessagingTaskQueue(delegate, nil, logger, config)

	// Enqueue a task
	task := &models.BackgroundTask{
		ID:          "test-task-1",
		TaskType:    "test_type",
		TaskName:    "Test Task",
		ScheduledAt: time.Now().Add(-time.Second),
	}
	err := queue.Enqueue(context.Background(), task)
	require.NoError(t, err)

	// Move to dead letter
	err = queue.MoveToDeadLetter(context.Background(), task.ID, "test failure")
	assert.NoError(t, err)
}

func TestMessagingTaskQueue_GetPendingCount(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	queue := NewMessagingTaskQueue(delegate, nil, logger, nil)

	// Initially empty
	count, err := queue.GetPendingCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Enqueue tasks
	for i := 0; i < 5; i++ {
		task := &models.BackgroundTask{
			TaskType: "test_type",
			TaskName: "Test Task",
		}
		queue.Enqueue(context.Background(), task)
	}

	// Verify count
	count, err = queue.GetPendingCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestMessagingTaskQueue_GetRunningCount(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	queue := NewMessagingTaskQueue(delegate, nil, logger, nil)

	// Initially empty
	count, err := queue.GetRunningCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Enqueue and dequeue a task
	task := &models.BackgroundTask{
		TaskType:    "test_type",
		TaskName:    "Test Task",
		ScheduledAt: time.Now().Add(-time.Second),
	}
	queue.Enqueue(context.Background(), task)
	queue.Dequeue(context.Background(), "worker-1", ResourceRequirements{})

	// Verify count
	count, err = queue.GetRunningCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestMessagingTaskQueue_GetQueueDepth(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	queue := NewMessagingTaskQueue(delegate, nil, logger, nil)

	// Enqueue tasks with different priorities
	priorities := []models.TaskPriority{
		models.TaskPriorityHigh,
		models.TaskPriorityHigh,
		models.TaskPriorityNormal,
		models.TaskPriorityLow,
	}

	for _, p := range priorities {
		task := &models.BackgroundTask{
			TaskType: "test_type",
			TaskName: "Test Task",
			Priority: p,
		}
		queue.Enqueue(context.Background(), task)
	}

	// Get queue depth
	depth, err := queue.GetQueueDepth(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(2), depth[models.TaskPriorityHigh])
	assert.Equal(t, int64(1), depth[models.TaskPriorityNormal])
	assert.Equal(t, int64(1), depth[models.TaskPriorityLow])
}

func TestMessagingTaskQueue_Publisher(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	queue := NewMessagingTaskQueue(delegate, nil, logger, nil)

	publisher := queue.Publisher()
	assert.NotNil(t, publisher)
}

func TestMessagingTaskQueue_StartStop(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	config := &MessagingTaskQueueConfig{
		PublisherConfig: &TaskEventPublisherConfig{
			Enabled:      true,
			AsyncPublish: true,
			BufferSize:   100,
		},
	}

	queue := NewMessagingTaskQueue(delegate, nil, logger, config)

	// Should not panic
	queue.Start()
	queue.Stop()
}

func TestNewMessagingProgressReporter(t *testing.T) {
	logger := logrus.New()
	config := &TaskEventPublisherConfig{
		Enabled:      false,
		AsyncPublish: false,
	}
	publisher := NewTaskEventPublisher(nil, logger, config)

	task := &models.BackgroundTask{
		ID:       "test-task-1",
		TaskType: "test_type",
	}

	mockReporter := &mockProgressReporter{}
	reporter := NewMessagingProgressReporter(mockReporter, publisher, task, logger)

	assert.NotNil(t, reporter)
	assert.Equal(t, mockReporter, reporter.delegate)
	assert.Equal(t, publisher, reporter.publisher)
	assert.Equal(t, task, reporter.task)
}

func TestMessagingProgressReporter_ReportProgress(t *testing.T) {
	logger := logrus.New()
	config := &TaskEventPublisherConfig{
		Enabled:      false,
		AsyncPublish: false,
	}
	publisher := NewTaskEventPublisher(nil, logger, config)

	task := &models.BackgroundTask{
		ID:       "test-task-1",
		TaskType: "test_type",
	}

	mockReporter := &mockProgressReporter{}
	reporter := NewMessagingProgressReporter(mockReporter, publisher, task, logger)

	err := reporter.ReportProgress(50, "Halfway done")
	assert.NoError(t, err)
	assert.Equal(t, 50.0, task.Progress)
	assert.Equal(t, "Halfway done", *task.ProgressMessage)
	assert.Equal(t, 50.0, mockReporter.lastPercent)
	assert.Equal(t, "Halfway done", mockReporter.lastMessage)
}

func TestMessagingProgressReporter_ReportHeartbeat(t *testing.T) {
	logger := logrus.New()
	config := &TaskEventPublisherConfig{
		Enabled:      false,
		AsyncPublish: false,
	}
	publisher := NewTaskEventPublisher(nil, logger, config)

	task := &models.BackgroundTask{
		ID:       "test-task-1",
		TaskType: "test_type",
	}

	mockReporter := &mockProgressReporter{}
	reporter := NewMessagingProgressReporter(mockReporter, publisher, task, logger)

	err := reporter.ReportHeartbeat()
	assert.NoError(t, err)
	assert.True(t, mockReporter.heartbeatCalled)
}

func TestMessagingProgressReporter_ReportCheckpoint(t *testing.T) {
	logger := logrus.New()
	config := &TaskEventPublisherConfig{
		Enabled:      false,
		AsyncPublish: false,
	}
	publisher := NewTaskEventPublisher(nil, logger, config)

	task := &models.BackgroundTask{
		ID:       "test-task-1",
		TaskType: "test_type",
	}

	mockReporter := &mockProgressReporter{}
	reporter := NewMessagingProgressReporter(mockReporter, publisher, task, logger)

	data := []byte("checkpoint data")
	err := reporter.ReportCheckpoint(data)
	assert.NoError(t, err)
	assert.Equal(t, data, mockReporter.lastCheckpoint)
}

func TestMessagingProgressReporter_ReportMetrics(t *testing.T) {
	logger := logrus.New()
	config := &TaskEventPublisherConfig{
		Enabled:      false,
		AsyncPublish: false,
	}
	publisher := NewTaskEventPublisher(nil, logger, config)

	task := &models.BackgroundTask{
		ID:       "test-task-1",
		TaskType: "test_type",
	}

	mockReporter := &mockProgressReporter{}
	reporter := NewMessagingProgressReporter(mockReporter, publisher, task, logger)

	metrics := map[string]interface{}{"cpu": 50, "memory": 1024}
	err := reporter.ReportMetrics(metrics)
	assert.NoError(t, err)
	assert.Equal(t, metrics, mockReporter.lastMetrics)
}

func TestMessagingProgressReporter_ReportLog(t *testing.T) {
	logger := logrus.New()
	config := &TaskEventPublisherConfig{
		Enabled:      false,
		AsyncPublish: false,
	}
	publisher := NewTaskEventPublisher(nil, logger, config)

	task := &models.BackgroundTask{
		ID:       "test-task-1",
		TaskType: "test_type",
	}

	mockReporter := &mockProgressReporter{}
	reporter := NewMessagingProgressReporter(mockReporter, publisher, task, logger)

	fields := map[string]interface{}{"source": "test"}
	err := reporter.ReportLog("info", "Test message", fields)
	assert.NoError(t, err)
	assert.Equal(t, "info", mockReporter.lastLogLevel)
	assert.Equal(t, "Test message", mockReporter.lastLogMessage)
	assert.Equal(t, fields, mockReporter.lastLogFields)
}

func TestNewMessagingTaskExecutorWrapper(t *testing.T) {
	logger := logrus.New()
	config := &TaskEventPublisherConfig{
		Enabled:      false,
		AsyncPublish: false,
	}
	publisher := NewTaskEventPublisher(nil, logger, config)

	mockExecutor := &mockTaskExecutor{}
	wrapper := NewMessagingTaskExecutorWrapper(mockExecutor, publisher, logger)

	assert.NotNil(t, wrapper)
	assert.Equal(t, mockExecutor, wrapper.executor)
	assert.Equal(t, publisher, wrapper.publisher)
}

func TestMessagingTaskExecutorWrapper_CanPause(t *testing.T) {
	logger := logrus.New()
	config := &TaskEventPublisherConfig{
		Enabled:      false,
		AsyncPublish: false,
	}
	publisher := NewTaskEventPublisher(nil, logger, config)

	mockExecutor := &mockTaskExecutor{canPause: true}
	wrapper := NewMessagingTaskExecutorWrapper(mockExecutor, publisher, logger)

	assert.True(t, wrapper.CanPause())
}

func TestMessagingTaskExecutorWrapper_GetResourceRequirements(t *testing.T) {
	logger := logrus.New()
	config := &TaskEventPublisherConfig{
		Enabled:      false,
		AsyncPublish: false,
	}
	publisher := NewTaskEventPublisher(nil, logger, config)

	expected := ResourceRequirements{
		CPUCores: 2,
		MemoryMB: 1024,
	}
	mockExecutor := &mockTaskExecutor{requirements: expected}
	wrapper := NewMessagingTaskExecutorWrapper(mockExecutor, publisher, logger)

	result := wrapper.GetResourceRequirements()
	assert.Equal(t, expected, result)
}

func TestSetupMessagingForWorkerPool(t *testing.T) {
	logger := logrus.New()
	delegate := NewInMemoryTaskQueue(logger)
	config := &MessagingTaskQueueConfig{
		PublisherConfig: &TaskEventPublisherConfig{
			Enabled:      false,
			AsyncPublish: false,
		},
	}

	msgQueue, err := SetupMessagingForWorkerPool(nil, delegate, nil, logger, config)

	assert.NoError(t, err)
	require.NotNil(t, msgQueue)
	assert.Equal(t, delegate, msgQueue.Delegate())
}

// mockProgressReporter is a mock implementation of ProgressReporter for testing.
type mockProgressReporter struct {
	lastPercent    float64
	lastMessage    string
	heartbeatCalled bool
	lastCheckpoint []byte
	lastMetrics    map[string]interface{}
	lastLogLevel   string
	lastLogMessage string
	lastLogFields  map[string]interface{}
}

func (m *mockProgressReporter) ReportProgress(percent float64, message string) error {
	m.lastPercent = percent
	m.lastMessage = message
	return nil
}

func (m *mockProgressReporter) ReportHeartbeat() error {
	m.heartbeatCalled = true
	return nil
}

func (m *mockProgressReporter) ReportCheckpoint(data []byte) error {
	m.lastCheckpoint = data
	return nil
}

func (m *mockProgressReporter) ReportMetrics(metrics map[string]interface{}) error {
	m.lastMetrics = metrics
	return nil
}

func (m *mockProgressReporter) ReportLog(level, message string, fields map[string]interface{}) error {
	m.lastLogLevel = level
	m.lastLogMessage = message
	m.lastLogFields = fields
	return nil
}

// mockTaskExecutor is a mock implementation of TaskExecutor for testing.
type mockTaskExecutor struct {
	canPause      bool
	requirements  ResourceRequirements
	executeError  error
	pauseData     []byte
	executeResult interface{}
}

func (m *mockTaskExecutor) Execute(ctx context.Context, task *models.BackgroundTask, reporter ProgressReporter) error {
	return m.executeError
}

func (m *mockTaskExecutor) CanPause() bool {
	return m.canPause
}

func (m *mockTaskExecutor) Pause(ctx context.Context, task *models.BackgroundTask) ([]byte, error) {
	return m.pauseData, nil
}

func (m *mockTaskExecutor) Resume(ctx context.Context, task *models.BackgroundTask, checkpoint []byte) error {
	return nil
}

func (m *mockTaskExecutor) Cancel(ctx context.Context, task *models.BackgroundTask) error {
	return nil
}

func (m *mockTaskExecutor) GetResourceRequirements() ResourceRequirements {
	return m.requirements
}
