package messaging

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskState_Values(t *testing.T) {
	assert.Equal(t, TaskState("pending"), TaskStatePending)
	assert.Equal(t, TaskState("queued"), TaskStateQueued)
	assert.Equal(t, TaskState("running"), TaskStateRunning)
	assert.Equal(t, TaskState("completed"), TaskStateCompleted)
	assert.Equal(t, TaskState("failed"), TaskStateFailed)
	assert.Equal(t, TaskState("canceled"), TaskStateCanceled)
	assert.Equal(t, TaskState("dead_lettered"), TaskStateDeadLettered)
}

func TestTaskPriority_Values(t *testing.T) {
	assert.Equal(t, TaskPriority(1), TaskPriorityLow)
	assert.Equal(t, TaskPriority(5), TaskPriorityNormal)
	assert.Equal(t, TaskPriority(8), TaskPriorityHigh)
	assert.Equal(t, TaskPriority(10), TaskPriorityCritical)
}

func TestNewTask(t *testing.T) {
	payload := []byte(`{"key": "value"}`)
	task := NewTask("test.task", payload)

	assert.NotEmpty(t, task.ID)
	assert.Contains(t, task.ID, "task-")
	assert.Equal(t, "test.task", task.Type)
	assert.Equal(t, payload, task.Payload)
	assert.Equal(t, TaskPriorityNormal, task.Priority)
	assert.Equal(t, TaskStatePending, task.State)
	assert.Equal(t, 3, task.MaxRetries)
	assert.Equal(t, 0, task.RetryCount)
	assert.False(t, task.CreatedAt.IsZero())
	assert.NotNil(t, task.Metadata)
}

func TestNewTaskWithID(t *testing.T) {
	payload := []byte(`{"key": "value"}`)
	task := NewTaskWithID("custom-task-id", "test.task", payload)

	assert.Equal(t, "custom-task-id", task.ID)
	assert.Equal(t, "test.task", task.Type)
	assert.Equal(t, payload, task.Payload)
}

func TestTask_WithPriority(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	task.WithPriority(TaskPriorityHigh)

	assert.Equal(t, TaskPriorityHigh, task.Priority)
}

func TestTask_WithMaxRetries(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	task.WithMaxRetries(5)

	assert.Equal(t, 5, task.MaxRetries)
}

func TestTask_WithDeadline(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	deadline := time.Now().Add(1 * time.Hour)
	task.WithDeadline(deadline)

	assert.Equal(t, deadline, task.Deadline)
}

func TestTask_WithTTL(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	before := time.Now().UTC()
	task.WithTTL(1 * time.Hour)
	after := time.Now().UTC()

	assert.True(t, task.Deadline.After(before.Add(59*time.Minute)))
	assert.True(t, task.Deadline.Before(after.Add(61*time.Minute)))
}

func TestTask_WithScheduledAt(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	scheduledAt := time.Now().Add(1 * time.Hour)
	task.WithScheduledAt(scheduledAt)

	assert.Equal(t, scheduledAt, task.ScheduledAt)
}

func TestTask_WithDelay(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	before := time.Now().UTC()
	task.WithDelay(1 * time.Hour)
	after := time.Now().UTC()

	assert.True(t, task.ScheduledAt.After(before.Add(59*time.Minute)))
	assert.True(t, task.ScheduledAt.Before(after.Add(61*time.Minute)))
}

func TestTask_WithMetadata(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	task.WithMetadata("key1", "value1").WithMetadata("key2", "value2")

	assert.Equal(t, "value1", task.Metadata["key1"])
	assert.Equal(t, "value2", task.Metadata["key2"])
}

func TestTask_WithTraceID(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	task.WithTraceID("trace-123")

	assert.Equal(t, "trace-123", task.TraceID)
}

func TestTask_WithCorrelationID(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	task.WithCorrelationID("corr-123")

	assert.Equal(t, "corr-123", task.CorrelationID)
}

func TestTask_WithParentTaskID(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	task.WithParentTaskID("parent-123")

	assert.Equal(t, "parent-123", task.ParentTaskID)
}

func TestTask_GetMetadata(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	task.Metadata["key"] = "value"

	assert.Equal(t, "value", task.GetMetadata("key"))
	assert.Equal(t, "", task.GetMetadata("nonexistent"))

	// Test with nil metadata
	task.Metadata = nil
	assert.Equal(t, "", task.GetMetadata("key"))
}

func TestTask_IsExpired(t *testing.T) {
	task := NewTask("test", []byte("payload"))

	// No deadline
	assert.False(t, task.IsExpired())

	// Expired
	task.Deadline = time.Now().Add(-1 * time.Hour)
	assert.True(t, task.IsExpired())

	// Not expired
	task.Deadline = time.Now().Add(1 * time.Hour)
	assert.False(t, task.IsExpired())
}

func TestTask_IsScheduled(t *testing.T) {
	task := NewTask("test", []byte("payload"))

	// Not scheduled
	assert.False(t, task.IsScheduled())

	// Scheduled for future
	task.ScheduledAt = time.Now().Add(1 * time.Hour)
	assert.True(t, task.IsScheduled())

	// Scheduled time passed
	task.ScheduledAt = time.Now().Add(-1 * time.Hour)
	assert.False(t, task.IsScheduled())
}

func TestTask_CanRetry(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	task.MaxRetries = 3

	assert.True(t, task.CanRetry())

	task.RetryCount = 2
	assert.True(t, task.CanRetry())

	task.RetryCount = 3
	assert.False(t, task.CanRetry())
}

func TestTask_IncrementRetry(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	assert.Equal(t, 0, task.RetryCount)

	task.IncrementRetry()
	assert.Equal(t, 1, task.RetryCount)

	task.IncrementRetry()
	assert.Equal(t, 2, task.RetryCount)
}

func TestTask_SetState(t *testing.T) {
	task := NewTask("test", []byte("payload"))

	task.SetState(TaskStateRunning)
	assert.Equal(t, TaskStateRunning, task.State)
	assert.False(t, task.StartedAt.IsZero())

	task.SetState(TaskStateCompleted)
	assert.Equal(t, TaskStateCompleted, task.State)
	assert.False(t, task.CompletedAt.IsZero())
}

func TestTask_SetResult(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	result := []byte(`{"result": "success"}`)

	task.SetResult(result)

	assert.Equal(t, result, task.Result)
	assert.Equal(t, TaskStateCompleted, task.State)
}

func TestTask_SetError(t *testing.T) {
	task := NewTask("test", []byte("payload"))

	task.SetError("something went wrong")

	assert.Equal(t, "something went wrong", task.Error)
	assert.Equal(t, TaskStateFailed, task.State)
}

func TestTask_Duration(t *testing.T) {
	task := NewTask("test", []byte("payload"))

	// Not started
	assert.Equal(t, time.Duration(0), task.Duration())

	// Running
	task.StartedAt = time.Now().Add(-100 * time.Millisecond)
	duration := task.Duration()
	assert.True(t, duration >= 100*time.Millisecond)

	// Completed
	task.CompletedAt = task.StartedAt.Add(50 * time.Millisecond)
	assert.Equal(t, 50*time.Millisecond, task.Duration())
}

func TestTask_Clone(t *testing.T) {
	task := NewTask("test", []byte("payload"))
	task.Metadata["key"] = "value"
	task.TraceID = "trace-123"
	task.Result = []byte("result")

	clone := task.Clone()

	assert.Equal(t, task.ID, clone.ID)
	assert.Equal(t, task.Type, clone.Type)
	assert.Equal(t, task.Payload, clone.Payload)
	assert.Equal(t, task.Metadata, clone.Metadata)
	assert.Equal(t, task.TraceID, clone.TraceID)
	assert.Equal(t, task.Result, clone.Result)

	// Verify deep copy
	task.Payload[0] = 'X'
	assert.NotEqual(t, task.Payload, clone.Payload)

	task.Metadata["key"] = "modified"
	assert.NotEqual(t, task.Metadata["key"], clone.Metadata["key"])
}

func TestTask_ToMessage(t *testing.T) {
	task := NewTask("test.task", []byte(`{"data":"test"}`))
	task.TraceID = "trace-123"
	task.CorrelationID = "corr-123"
	task.Priority = TaskPriorityHigh
	task.Deadline = time.Now().Add(1 * time.Hour)

	msg := task.ToMessage()

	assert.Equal(t, task.ID, msg.ID)
	assert.Equal(t, task.Type, msg.Type)
	assert.Equal(t, task.TraceID, msg.TraceID)
	assert.Equal(t, task.CorrelationID, msg.CorrelationID)
	assert.Equal(t, MessagePriority(task.Priority), msg.Priority)
}

func TestTaskFromMessage(t *testing.T) {
	originalTask := NewTask("test.task", []byte(`{"data":"test"}`))
	originalTask.TraceID = "trace-123"

	msg := originalTask.ToMessage()
	msg.DeliveryTag = 123

	recoveredTask, err := TaskFromMessage(msg)
	require.NoError(t, err)

	assert.Equal(t, originalTask.ID, recoveredTask.ID)
	assert.Equal(t, originalTask.Type, recoveredTask.Type)
	assert.Equal(t, uint64(123), recoveredTask.DeliveryTag)
}

func TestTaskFromMessage_InvalidJSON(t *testing.T) {
	msg := NewMessage("test", []byte("invalid json"))

	_, err := TaskFromMessage(msg)
	assert.Error(t, err)
}

func TestDefaultTaskQueueConfig(t *testing.T) {
	cfg := DefaultTaskQueueConfig()

	assert.Equal(t, "helixagent.tasks.default", cfg.DefaultQueue)
	assert.Equal(t, "helixagent.tasks.dlq", cfg.DeadLetterQueue)
	assert.Equal(t, "helixagent.tasks.retry", cfg.RetryQueue)
	assert.Equal(t, 10, cfg.PrefetchCount)
	assert.True(t, cfg.PublisherConfirm)
	assert.Equal(t, 10, cfg.MaxPriority)
	assert.Equal(t, 24*time.Hour, cfg.TaskTTL)
	assert.Equal(t, 5*time.Second, cfg.RetryDelay)
	assert.Equal(t, 3, cfg.MaxRetries)
}

func TestDefaultTaskWorkerConfig(t *testing.T) {
	cfg := DefaultTaskWorkerConfig()

	assert.Contains(t, cfg.WorkerID, "worker-")
	assert.Equal(t, QueueBackgroundTasks, cfg.Queue)
	assert.Equal(t, 10, cfg.Concurrency)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 5*time.Second, cfg.RetryDelay)
	assert.Equal(t, 5*time.Minute, cfg.Timeout)
	assert.Equal(t, 30*time.Second, cfg.GracefulShutdown)
}

func TestNewTaskResult(t *testing.T) {
	// Success result
	result := NewTaskResult("task-123", map[string]string{"key": "value"}, nil, 100*time.Millisecond)
	assert.Equal(t, "task-123", result.TaskID)
	assert.True(t, result.Success)
	assert.NotNil(t, result.Result)
	assert.Empty(t, result.Error)
	assert.Equal(t, 100*time.Millisecond, result.Duration)

	// Failed result
	result2 := NewTaskResult("task-456", nil, ErrHandlerError, 50*time.Millisecond)
	assert.Equal(t, "task-456", result2.TaskID)
	assert.False(t, result2.Success)
	assert.Nil(t, result2.Result)
	assert.NotEmpty(t, result2.Error)
}

func TestTaskRegistry(t *testing.T) {
	registry := NewTaskRegistry()

	handler := func(ctx context.Context, task *Task) error { return nil }

	// Register
	registry.Register("task.type1", handler)
	registry.Register("task.type2", handler)

	// Get
	h, ok := registry.Get("task.type1")
	assert.True(t, ok)
	assert.NotNil(t, h)

	h, ok = registry.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, h)

	// Types
	types := registry.Types()
	assert.Len(t, types, 2)
	assert.Contains(t, types, "task.type1")
	assert.Contains(t, types, "task.type2")

	// Unregister
	registry.Unregister("task.type1")
	_, ok = registry.Get("task.type1")
	assert.False(t, ok)
}

func TestQueueConstants(t *testing.T) {
	assert.Equal(t, "helixagent.tasks.background", QueueBackgroundTasks)
	assert.Equal(t, "helixagent.tasks.llm", QueueLLMRequests)
	assert.Equal(t, "helixagent.tasks.debate", QueueDebateRounds)
	assert.Equal(t, "helixagent.tasks.verification", QueueVerification)
	assert.Equal(t, "helixagent.tasks.notifications", QueueNotifications)
	assert.Equal(t, "helixagent.tasks.dlq", QueueDeadLetter)
	assert.Equal(t, "helixagent.tasks.retry", QueueRetry)
}

func TestExchangeConstants(t *testing.T) {
	assert.Equal(t, "helixagent.tasks", ExchangeTasks)
	assert.Equal(t, "helixagent.events", ExchangeEvents)
	assert.Equal(t, "helixagent.notifications", ExchangeNotifications)
	assert.Equal(t, "helixagent.dlx", ExchangeDeadLetter)
}

func TestTask_JSONSerialization(t *testing.T) {
	task := NewTask("test.task", []byte(`{"data":"test"}`))
	task.Priority = TaskPriorityHigh
	task.TraceID = "trace-123"
	task.Metadata["key"] = "value"

	// Marshal
	data, err := json.Marshal(task)
	require.NoError(t, err)

	// Unmarshal
	var decoded Task
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, task.ID, decoded.ID)
	assert.Equal(t, task.Type, decoded.Type)
	assert.Equal(t, task.Priority, decoded.Priority)
	assert.Equal(t, task.TraceID, decoded.TraceID)
}
