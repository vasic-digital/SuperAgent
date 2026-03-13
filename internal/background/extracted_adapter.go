package background

import (
	"context"
	"encoding/json"
	"time"

	"dev.helix.agent/internal/models"
	extractedbackground "digital.vasic.background"
	extractedmodels "digital.vasic.models"
)

// TaskRepositoryAdapter adapts internal TaskRepository to extracted TaskRepository interface
type TaskRepositoryAdapter struct {
	repo TaskRepository
}

// NewTaskRepositoryAdapter creates a new adapter wrapping an internal TaskRepository
func NewTaskRepositoryAdapter(repo TaskRepository) *TaskRepositoryAdapter {
	return &TaskRepositoryAdapter{repo: repo}
}

// Create implements extracted TaskRepository.Create
func (a *TaskRepositoryAdapter) Create(ctx context.Context, task *extractedmodels.BackgroundTask) error {
	internalTask := convertToInternalTask(task)
	return a.repo.Create(ctx, internalTask)
}

// GetByID implements extracted TaskRepository.GetByID
func (a *TaskRepositoryAdapter) GetByID(ctx context.Context, id string) (*extractedmodels.BackgroundTask, error) {
	internalTask, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertToExtractedTask(internalTask), nil
}

// Update implements extracted TaskRepository.Update
func (a *TaskRepositoryAdapter) Update(ctx context.Context, task *extractedmodels.BackgroundTask) error {
	internalTask := convertToInternalTask(task)
	return a.repo.Update(ctx, internalTask)
}

// Delete implements extracted TaskRepository.Delete
func (a *TaskRepositoryAdapter) Delete(ctx context.Context, id string) error {
	return a.repo.Delete(ctx, id)
}

// UpdateStatus implements extracted TaskRepository.UpdateStatus
func (a *TaskRepositoryAdapter) UpdateStatus(ctx context.Context, id string, status extractedmodels.TaskStatus) error {
	internalStatus := models.TaskStatus(status)
	return a.repo.UpdateStatus(ctx, id, internalStatus)
}

// UpdateProgress implements extracted TaskRepository.UpdateProgress
func (a *TaskRepositoryAdapter) UpdateProgress(ctx context.Context, id string, progress float64, message string) error {
	return a.repo.UpdateProgress(ctx, id, progress, message)
}

// UpdateHeartbeat implements extracted TaskRepository.UpdateHeartbeat
func (a *TaskRepositoryAdapter) UpdateHeartbeat(ctx context.Context, id string) error {
	return a.repo.UpdateHeartbeat(ctx, id)
}

// SaveCheckpoint implements extracted TaskRepository.SaveCheckpoint
func (a *TaskRepositoryAdapter) SaveCheckpoint(ctx context.Context, id string, checkpoint []byte) error {
	return a.repo.SaveCheckpoint(ctx, id, checkpoint)
}

// GetByStatus implements extracted TaskRepository.GetByStatus
func (a *TaskRepositoryAdapter) GetByStatus(ctx context.Context, status extractedmodels.TaskStatus, limit, offset int) ([]*extractedmodels.BackgroundTask, error) {
	internalStatus := models.TaskStatus(status)
	internalTasks, err := a.repo.GetByStatus(ctx, internalStatus, limit, offset)
	if err != nil {
		return nil, err
	}
	return convertToExtractedTasks(internalTasks), nil
}

// GetPendingTasks implements extracted TaskRepository.GetPendingTasks
func (a *TaskRepositoryAdapter) GetPendingTasks(ctx context.Context, limit int) ([]*extractedmodels.BackgroundTask, error) {
	internalTasks, err := a.repo.GetPendingTasks(ctx, limit)
	if err != nil {
		return nil, err
	}
	return convertToExtractedTasks(internalTasks), nil
}

// GetStaleTasks implements extracted TaskRepository.GetStaleTasks
func (a *TaskRepositoryAdapter) GetStaleTasks(ctx context.Context, threshold time.Duration) ([]*extractedmodels.BackgroundTask, error) {
	internalTasks, err := a.repo.GetStaleTasks(ctx, threshold)
	if err != nil {
		return nil, err
	}
	return convertToExtractedTasks(internalTasks), nil
}

// GetByWorkerID implements extracted TaskRepository.GetByWorkerID
func (a *TaskRepositoryAdapter) GetByWorkerID(ctx context.Context, workerID string) ([]*extractedmodels.BackgroundTask, error) {
	internalTasks, err := a.repo.GetByWorkerID(ctx, workerID)
	if err != nil {
		return nil, err
	}
	return convertToExtractedTasks(internalTasks), nil
}

// CountByStatus implements extracted TaskRepository.CountByStatus
func (a *TaskRepositoryAdapter) CountByStatus(ctx context.Context) (map[extractedmodels.TaskStatus]int64, error) {
	internalCounts, err := a.repo.CountByStatus(ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[extractedmodels.TaskStatus]int64)
	for status, count := range internalCounts {
		counts[extractedmodels.TaskStatus(status)] = count
	}
	return counts, nil
}

// Dequeue implements extracted TaskRepository.Dequeue
func (a *TaskRepositoryAdapter) Dequeue(ctx context.Context, workerID string, maxCPUCores, maxMemoryMB int) (*extractedmodels.BackgroundTask, error) {
	internalTask, err := a.repo.Dequeue(ctx, workerID, maxCPUCores, maxMemoryMB)
	if err != nil {
		return nil, err
	}
	return convertToExtractedTask(internalTask), nil
}

// SaveResourceSnapshot implements extracted TaskRepository.SaveResourceSnapshot
func (a *TaskRepositoryAdapter) SaveResourceSnapshot(ctx context.Context, snapshot *extractedmodels.ResourceSnapshot) error {
	if a.repo == nil || snapshot == nil {
		return nil
	}
	// Convert extracted ResourceSnapshot to internal ResourceSnapshot
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	var internalSnapshot models.ResourceSnapshot
	if err := json.Unmarshal(data, &internalSnapshot); err != nil {
		return err
	}
	return a.repo.SaveResourceSnapshot(ctx, &internalSnapshot)
}

// GetResourceSnapshots implements extracted TaskRepository.GetResourceSnapshots
func (a *TaskRepositoryAdapter) GetResourceSnapshots(ctx context.Context, taskID string, limit int) ([]*extractedmodels.ResourceSnapshot, error) {
	internalSnapshots, err := a.repo.GetResourceSnapshots(ctx, taskID, limit)
	if err != nil {
		return nil, err
	}
	// Convert slice
	extractedSnapshots := make([]*extractedmodels.ResourceSnapshot, len(internalSnapshots))
	for i, snapshot := range internalSnapshots {
		data, err := json.Marshal(snapshot)
		if err != nil {
			return nil, err
		}
		var extractedSnapshot extractedmodels.ResourceSnapshot
		if err := json.Unmarshal(data, &extractedSnapshot); err != nil {
			return nil, err
		}
		extractedSnapshots[i] = &extractedSnapshot
	}
	return extractedSnapshots, nil
}

// LogEvent implements extracted TaskRepository.LogEvent
func (a *TaskRepositoryAdapter) LogEvent(ctx context.Context, taskID, eventType string, data map[string]interface{}, workerID *string) error {
	return a.repo.LogEvent(ctx, taskID, eventType, data, workerID)
}

// GetTaskHistory implements extracted TaskRepository.GetTaskHistory
func (a *TaskRepositoryAdapter) GetTaskHistory(ctx context.Context, taskID string, limit int) ([]*extractedmodels.TaskExecutionHistory, error) {
	internalHistory, err := a.repo.GetTaskHistory(ctx, taskID, limit)
	if err != nil {
		return nil, err
	}
	// Convert slice
	extractedHistory := make([]*extractedmodels.TaskExecutionHistory, len(internalHistory))
	for i, history := range internalHistory {
		data, err := json.Marshal(history)
		if err != nil {
			return nil, err
		}
		var extracted extractedmodels.TaskExecutionHistory
		if err := json.Unmarshal(data, &extracted); err != nil {
			return nil, err
		}
		extractedHistory[i] = &extracted
	}
	return extractedHistory, nil
}

// MoveToDeadLetter implements extracted TaskRepository.MoveToDeadLetter
func (a *TaskRepositoryAdapter) MoveToDeadLetter(ctx context.Context, taskID, reason string) error {
	return a.repo.MoveToDeadLetter(ctx, taskID, reason)
}

// =============================================================================
// TaskQueue Adapter - Adapts extracted TaskQueue to internal TaskQueue interface
// =============================================================================

// TaskQueueAdapter wraps the extracted TaskQueue to implement internal TaskQueue interface
type TaskQueueAdapter struct {
	queue extractedbackground.TaskQueue
}

// NewTaskQueueAdapter creates a new adapter wrapping an extracted TaskQueue
func NewTaskQueueAdapter(queue extractedbackground.TaskQueue) *TaskQueueAdapter {
	return &TaskQueueAdapter{queue: queue}
}

// Enqueue adds a task to the queue
func (a *TaskQueueAdapter) Enqueue(ctx context.Context, task *models.BackgroundTask) error {
	extractedTask := convertToExtractedTask(task)
	err := a.queue.Enqueue(ctx, extractedTask)
	if err != nil {
		return err
	}
	// Copy back fields that may have been set by the extracted queue
	if extractedTask.ID != "" {
		task.ID = extractedTask.ID
	}
	if extractedTask.Status != "" {
		task.Status = models.TaskStatus(extractedTask.Status)
	}
	if extractedTask.Priority != "" {
		task.Priority = models.TaskPriority(extractedTask.Priority)
	}
	if !extractedTask.ScheduledAt.IsZero() {
		task.ScheduledAt = extractedTask.ScheduledAt
	}
	if !extractedTask.CreatedAt.IsZero() {
		task.CreatedAt = extractedTask.CreatedAt
	}
	if !extractedTask.UpdatedAt.IsZero() {
		task.UpdatedAt = extractedTask.UpdatedAt
	}
	return nil
}

// Dequeue atomically retrieves and claims a task from the queue
func (a *TaskQueueAdapter) Dequeue(ctx context.Context, workerID string, requirements ResourceRequirements) (*models.BackgroundTask, error) {
	extractedRequirements := extractedbackground.ResourceRequirements{
		CPUCores: requirements.CPUCores,
		MemoryMB: requirements.MemoryMB,
		DiskMB:   requirements.DiskMB,
		GPUCount: requirements.GPUCount,
		Priority: convertTaskPriority(requirements.Priority),
	}
	extractedTask, err := a.queue.Dequeue(ctx, workerID, extractedRequirements)
	if err != nil {
		return nil, err
	}
	return convertToInternalTask(extractedTask), nil
}

// Peek returns tasks without claiming them
func (a *TaskQueueAdapter) Peek(ctx context.Context, count int) ([]*models.BackgroundTask, error) {
	extractedTasks, err := a.queue.Peek(ctx, count)
	if err != nil {
		return nil, err
	}
	return convertToInternalTasks(extractedTasks), nil
}

// Requeue returns a task to the queue with optional delay
func (a *TaskQueueAdapter) Requeue(ctx context.Context, taskID string, delay time.Duration) error {
	return a.queue.Requeue(ctx, taskID, delay)
}

// MoveToDeadLetter moves a failed task to dead-letter queue
func (a *TaskQueueAdapter) MoveToDeadLetter(ctx context.Context, taskID string, reason string) error {
	return a.queue.MoveToDeadLetter(ctx, taskID, reason)
}

// GetPendingCount returns the number of pending tasks
func (a *TaskQueueAdapter) GetPendingCount(ctx context.Context) (int64, error) {
	return a.queue.GetPendingCount(ctx)
}

// GetRunningCount returns the number of running tasks
func (a *TaskQueueAdapter) GetRunningCount(ctx context.Context) (int64, error) {
	return a.queue.GetRunningCount(ctx)
}

// GetQueueDepth returns counts by priority
func (a *TaskQueueAdapter) GetQueueDepth(ctx context.Context) (map[models.TaskPriority]int64, error) {
	extractedDepth, err := a.queue.GetQueueDepth(ctx)
	if err != nil {
		return nil, err
	}
	depth := make(map[models.TaskPriority]int64)
	for priority, count := range extractedDepth {
		depth[models.TaskPriority(priority)] = count
	}
	return depth, nil
}
