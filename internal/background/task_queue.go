package background

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/models"
	extractedbackground "digital.vasic.background"
)

// PostgresTaskQueue implements TaskQueue using PostgreSQL
type PostgresTaskQueue struct {
	// adapter wraps the extracted module's TaskQueue implementation
	*TaskQueueAdapter
	// repository is kept for backward compatibility but unused
	repository TaskRepository
	logger     *logrus.Logger
	mu         sync.RWMutex //nolint:unused
}

// NewPostgresTaskQueue creates a new PostgreSQL-backed task queue
func NewPostgresTaskQueue(repository TaskRepository, logger *logrus.Logger) *PostgresTaskQueue {
	// Adapt internal repository to extracted repository interface
	adaptedRepo := NewTaskRepositoryAdapter(repository)
	// Create extracted PostgresTaskQueue
	extractedQueue := extractedbackground.NewPostgresTaskQueue(adaptedRepo, logger)
	// Wrap extracted queue with adapter to internal TaskQueue interface
	queueAdapter := NewTaskQueueAdapter(extractedQueue)

	return &PostgresTaskQueue{
		TaskQueueAdapter: queueAdapter,
		repository:       repository,
		logger:           logger,
	}
}

// Enqueue adds a task to the queue

// Dequeue atomically retrieves and claims a task from the queue

// Peek returns tasks without claiming them

// Requeue returns a task to the queue with optional delay

// MoveToDeadLetter moves a failed task to dead-letter queue

// GetQueueDepth returns counts by priority

// TaskQueueStats holds queue statistics
type TaskQueueStats struct {
	PendingCount    int64                         `json:"pending_count"`
	RunningCount    int64                         `json:"running_count"`
	DepthByPriority map[models.TaskPriority]int64 `json:"depth_by_priority"`
	StatusCounts    map[models.TaskStatus]int64   `json:"status_counts"`
}

// GetStats returns queue statistics
func (q *PostgresTaskQueue) GetStats(ctx context.Context) (*TaskQueueStats, error) {
	statusCounts, err := q.repository.CountByStatus(ctx)
	if err != nil {
		return nil, err
	}

	depth, err := q.GetQueueDepth(ctx)
	if err != nil {
		depth = make(map[models.TaskPriority]int64)
	}

	return &TaskQueueStats{
		PendingCount:    statusCounts[models.TaskStatusPending],
		RunningCount:    statusCounts[models.TaskStatusRunning],
		DepthByPriority: depth,
		StatusCounts:    statusCounts,
	}, nil
}

// InMemoryTaskQueue provides an in-memory queue implementation for testing
type InMemoryTaskQueue struct {
	// adapter wraps the extracted module's InMemoryTaskQueue implementation
	*TaskQueueAdapter
	// extractedQueue is the concrete extracted InMemoryTaskQueue for test helpers
	extractedQueue *extractedbackground.InMemoryTaskQueue
	logger         *logrus.Logger
}

// NewInMemoryTaskQueue creates a new in-memory task queue
func NewInMemoryTaskQueue(logger *logrus.Logger) *InMemoryTaskQueue {
	// Create extracted InMemoryTaskQueue
	extractedQueue := extractedbackground.NewInMemoryTaskQueue(logger)
	// Wrap extracted queue with adapter to internal TaskQueue interface
	queueAdapter := NewTaskQueueAdapter(extractedQueue)

	return &InMemoryTaskQueue{
		TaskQueueAdapter: queueAdapter,
		extractedQueue:   extractedQueue,
		logger:           logger,
	}
}

// GetTask returns a task by ID (for testing)
func (q *InMemoryTaskQueue) GetTask(taskID string) *models.BackgroundTask {
	extractedTask := q.extractedQueue.GetTask(taskID)
	return convertToInternalTask(extractedTask)
}

// UpdateTask updates a task (for testing)
func (q *InMemoryTaskQueue) UpdateTask(task *models.BackgroundTask) {
	extractedTask := convertToExtractedTask(task)
	q.extractedQueue.UpdateTask(extractedTask)
}

// MarshalJSON for TaskQueueStats
func (s *TaskQueueStats) MarshalJSON() ([]byte, error) {
	type Alias TaskQueueStats
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})
}
