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

// =============================================================================
// InternalTaskQueueAdapter - Adapts internal TaskQueue to extracted TaskQueue interface
// =============================================================================

// InternalTaskQueueAdapter wraps an internal TaskQueue to implement extracted TaskQueue interface
type InternalTaskQueueAdapter struct {
	queue TaskQueue
}

// NewInternalTaskQueueAdapter creates a new adapter wrapping an internal TaskQueue
func NewInternalTaskQueueAdapter(queue TaskQueue) *InternalTaskQueueAdapter {
	return &InternalTaskQueueAdapter{queue: queue}
}

// Enqueue adds a task to the queue
func (a *InternalTaskQueueAdapter) Enqueue(ctx context.Context, task *extractedmodels.BackgroundTask) error {
	internalTask := convertToInternalTask(task)
	err := a.queue.Enqueue(ctx, internalTask)
	if err != nil {
		return err
	}
	// Copy back fields that may have been set by the internal queue
	if internalTask.ID != "" {
		task.ID = internalTask.ID
	}
	if internalTask.Status != "" {
		task.Status = extractedmodels.TaskStatus(internalTask.Status)
	}
	if internalTask.Priority != "" {
		task.Priority = extractedmodels.TaskPriority(internalTask.Priority)
	}
	if !internalTask.ScheduledAt.IsZero() {
		task.ScheduledAt = internalTask.ScheduledAt
	}
	if !internalTask.CreatedAt.IsZero() {
		task.CreatedAt = internalTask.CreatedAt
	}
	if !internalTask.UpdatedAt.IsZero() {
		task.UpdatedAt = internalTask.UpdatedAt
	}
	return nil
}

// Dequeue atomically retrieves and claims a task from the queue
func (a *InternalTaskQueueAdapter) Dequeue(ctx context.Context, workerID string, requirements extractedbackground.ResourceRequirements) (*extractedmodels.BackgroundTask, error) {
	internalRequirements := ResourceRequirements{
		CPUCores: requirements.CPUCores,
		MemoryMB: requirements.MemoryMB,
		DiskMB:   requirements.DiskMB,
		GPUCount: requirements.GPUCount,
		Priority: models.TaskPriority(requirements.Priority),
	}
	internalTask, err := a.queue.Dequeue(ctx, workerID, internalRequirements)
	if err != nil {
		return nil, err
	}
	return convertToExtractedTask(internalTask), nil
}

// Peek returns tasks without claiming them
func (a *InternalTaskQueueAdapter) Peek(ctx context.Context, count int) ([]*extractedmodels.BackgroundTask, error) {
	internalTasks, err := a.queue.Peek(ctx, count)
	if err != nil {
		return nil, err
	}
	return convertToExtractedTasks(internalTasks), nil
}

// Requeue returns a task to the queue with optional delay
func (a *InternalTaskQueueAdapter) Requeue(ctx context.Context, taskID string, delay time.Duration) error {
	return a.queue.Requeue(ctx, taskID, delay)
}

// MoveToDeadLetter moves a failed task to dead-letter queue
func (a *InternalTaskQueueAdapter) MoveToDeadLetter(ctx context.Context, taskID string, reason string) error {
	return a.queue.MoveToDeadLetter(ctx, taskID, reason)
}

// GetPendingCount returns the number of pending tasks
func (a *InternalTaskQueueAdapter) GetPendingCount(ctx context.Context) (int64, error) {
	return a.queue.GetPendingCount(ctx)
}

// GetRunningCount returns the number of running tasks
func (a *InternalTaskQueueAdapter) GetRunningCount(ctx context.Context) (int64, error) {
	return a.queue.GetRunningCount(ctx)
}

// GetQueueDepth returns counts by priority
func (a *InternalTaskQueueAdapter) GetQueueDepth(ctx context.Context) (map[extractedmodels.TaskPriority]int64, error) {
	internalDepth, err := a.queue.GetQueueDepth(ctx)
	if err != nil {
		return nil, err
	}
	depth := make(map[extractedmodels.TaskPriority]int64)
	for priority, count := range internalDepth {
		depth[extractedmodels.TaskPriority(priority)] = count
	}
	return depth, nil
}

// =============================================================================
// WorkerPool Adapter - Adapts extracted WorkerPool to internal WorkerPool interface
// =============================================================================

// WorkerPoolAdapter wraps the extracted WorkerPool to implement internal WorkerPool interface
type WorkerPoolAdapter struct {
	pool extractedbackground.WorkerPool
}

// NewWorkerPoolAdapter creates a new adapter wrapping an extracted WorkerPool
func NewWorkerPoolAdapter(pool extractedbackground.WorkerPool) *WorkerPoolAdapter {
	return &WorkerPoolAdapter{pool: pool}
}

// Start initializes and starts the worker pool
func (a *WorkerPoolAdapter) Start(ctx context.Context) error {
	return a.pool.Start(ctx)
}

// Stop gracefully stops the worker pool
func (a *WorkerPoolAdapter) Stop(gracePeriod time.Duration) error {
	return a.pool.Stop(gracePeriod)
}

// RegisterExecutor registers a task executor for a task type
func (a *WorkerPoolAdapter) RegisterExecutor(taskType string, executor TaskExecutor) {
	// Adapt internal TaskExecutor to extracted TaskExecutor
	// We need to create an adapter that implements extractedbackground.TaskExecutor
	// For now, we can use TaskExecutorAdapter (which adapts internal to extracted)
	// but that adapter is in internal/adapters/background. We'll create a simple one here.
	extractedExecutor := &taskExecutorAdapter{executor: executor}
	a.pool.RegisterExecutor(taskType, extractedExecutor)
}

// GetWorkerCount returns the current number of workers
func (a *WorkerPoolAdapter) GetWorkerCount() int {
	return a.pool.GetWorkerCount()
}

// GetActiveTaskCount returns the number of currently executing tasks
func (a *WorkerPoolAdapter) GetActiveTaskCount() int {
	return a.pool.GetActiveTaskCount()
}

// GetWorkerStatus returns status information for all workers
func (a *WorkerPoolAdapter) GetWorkerStatus() []WorkerStatus {
	extractedStatus := a.pool.GetWorkerStatus()
	status := make([]WorkerStatus, len(extractedStatus))
	for i, s := range extractedStatus {
		status[i] = *convertToInternalWorkerStatus(&s)
	}
	return status
}

// Scale manually adjusts the worker count
func (a *WorkerPoolAdapter) Scale(targetCount int) error {
	return a.pool.Scale(targetCount)
}

// =============================================================================
// taskExecutorAdapter - Adapts internal TaskExecutor to extracted TaskExecutor (helper)
// =============================================================================

type taskExecutorAdapter struct {
	executor TaskExecutor
}

func (a *taskExecutorAdapter) Execute(ctx context.Context, task *extractedmodels.BackgroundTask, reporter extractedbackground.ProgressReporter) error {
	internalTask := convertToInternalTask(task)
	internalReporter := &progressReporterAdapter{reporter: reporter}
	return a.executor.Execute(ctx, internalTask, internalReporter)
}

func (a *taskExecutorAdapter) CanPause() bool {
	return a.executor.CanPause()
}

func (a *taskExecutorAdapter) Pause(ctx context.Context, task *extractedmodels.BackgroundTask) ([]byte, error) {
	internalTask := convertToInternalTask(task)
	return a.executor.Pause(ctx, internalTask)
}

func (a *taskExecutorAdapter) Resume(ctx context.Context, task *extractedmodels.BackgroundTask, checkpoint []byte) error {
	internalTask := convertToInternalTask(task)
	return a.executor.Resume(ctx, internalTask, checkpoint)
}

func (a *taskExecutorAdapter) Cancel(ctx context.Context, task *extractedmodels.BackgroundTask) error {
	internalTask := convertToInternalTask(task)
	return a.executor.Cancel(ctx, internalTask)
}

func (a *taskExecutorAdapter) GetResourceRequirements() extractedbackground.ResourceRequirements {
	internalReq := a.executor.GetResourceRequirements()
	return convertToExtractedResourceRequirements(internalReq)
}

// =============================================================================
// progressReporterAdapter - Adapts extracted ProgressReporter to internal ProgressReporter (helper)
// =============================================================================

type progressReporterAdapter struct {
	reporter extractedbackground.ProgressReporter
}

func (a *progressReporterAdapter) ReportProgress(percent float64, message string) error {
	return a.reporter.ReportProgress(percent, message)
}

func (a *progressReporterAdapter) ReportHeartbeat() error {
	return a.reporter.ReportHeartbeat()
}

func (a *progressReporterAdapter) ReportCheckpoint(data []byte) error {
	return a.reporter.ReportCheckpoint(data)
}

func (a *progressReporterAdapter) ReportMetrics(metrics map[string]interface{}) error {
	return a.reporter.ReportMetrics(metrics)
}

func (a *progressReporterAdapter) ReportLog(level, message string, fields map[string]interface{}) error {
	return a.reporter.ReportLog(level, message, fields)
}

// =============================================================================
// ResourceMonitor Adapter - Adapts internal ResourceMonitor to extracted ResourceMonitor
// =============================================================================

type resourceMonitorAdapter struct {
	monitor ResourceMonitor
}

func (a *resourceMonitorAdapter) GetSystemResources() (*extractedbackground.SystemResources, error) {
	if a.monitor == nil {
		// Return default system resources with high limits
		return &extractedbackground.SystemResources{
			TotalCPUCores:     64,
			AvailableCPUCores: 64,
			TotalMemoryMB:     1 << 30, // 1 TB in MB
			AvailableMemoryMB: 1 << 30,
			CPULoadPercent:    0.0,
			MemoryUsedPercent: 0.0,
			DiskUsedPercent:   0.0,
			LoadAvg1:          0.0,
			LoadAvg5:          0.0,
			LoadAvg15:         0.0,
		}, nil
	}
	internalResources, err := a.monitor.GetSystemResources()
	if err != nil {
		return nil, err
	}
	return convertToExtractedSystemResources(internalResources), nil
}

func (a *resourceMonitorAdapter) GetProcessResources(pid int) (*extractedmodels.ResourceSnapshot, error) {
	if a.monitor == nil {
		// Return empty snapshot
		return &extractedmodels.ResourceSnapshot{
			ID:             "",
			TaskID:         "",
			CPUPercent:     0.0,
			CPUUserTime:    0.0,
			CPUSystemTime:  0.0,
			MemoryRSSBytes: 0,
			MemoryVMSBytes: 0,
			MemoryPercent:  0.0,
			IOReadBytes:    0,
			IOWriteBytes:   0,
			IOReadCount:    0,
			IOWriteCount:   0,
			NetBytesSent:   0,
			NetBytesRecv:   0,
			NetConnections: 0,
			OpenFiles:      0,
			OpenFDs:        0,
		}, nil
	}
	internalSnapshot, err := a.monitor.GetProcessResources(pid)
	if err != nil {
		return nil, err
	}
	return convertToExtractedResourceSnapshot(internalSnapshot), nil
}

func (a *resourceMonitorAdapter) StartMonitoring(taskID string, pid int, interval time.Duration) error {
	if a.monitor == nil {
		return nil
	}
	return a.monitor.StartMonitoring(taskID, pid, interval)
}

func (a *resourceMonitorAdapter) StopMonitoring(taskID string) error {
	if a.monitor == nil {
		return nil
	}
	return a.monitor.StopMonitoring(taskID)
}

func (a *resourceMonitorAdapter) GetLatestSnapshot(taskID string) (*extractedmodels.ResourceSnapshot, error) {
	if a.monitor == nil {
		// Return empty snapshot
		return &extractedmodels.ResourceSnapshot{
			ID:             "",
			TaskID:         taskID,
			CPUPercent:     0.0,
			CPUUserTime:    0.0,
			CPUSystemTime:  0.0,
			MemoryRSSBytes: 0,
			MemoryVMSBytes: 0,
			MemoryPercent:  0.0,
			IOReadBytes:    0,
			IOWriteBytes:   0,
			IOReadCount:    0,
			IOWriteCount:   0,
			NetBytesSent:   0,
			NetBytesRecv:   0,
			NetConnections: 0,
			OpenFiles:      0,
			OpenFDs:        0,
		}, nil
	}
	internalSnapshot, err := a.monitor.GetLatestSnapshot(taskID)
	if err != nil {
		return nil, err
	}
	return convertToExtractedResourceSnapshot(internalSnapshot), nil
}

func (a *resourceMonitorAdapter) IsResourceAvailable(requirements extractedbackground.ResourceRequirements) bool {
	if a.monitor == nil {
		// If no monitor, assume resources available
		return true
	}
	internalReq := convertToInternalResourceRequirements(requirements)
	return a.monitor.IsResourceAvailable(internalReq)
}

// =============================================================================
// StuckDetector Adapter - Adapts internal StuckDetector to extracted StuckDetector
// =============================================================================

type stuckDetectorAdapter struct {
	detector StuckDetector
}

func (a *stuckDetectorAdapter) IsStuck(ctx context.Context, task *extractedmodels.BackgroundTask, snapshots []*extractedmodels.ResourceSnapshot) (bool, string) {
	if a.detector == nil {
		return false, ""
	}
	internalTask := convertToInternalTask(task)
	internalSnapshots := make([]*models.ResourceSnapshot, len(snapshots))
	for i, snapshot := range snapshots {
		internalSnapshots[i] = convertToInternalResourceSnapshot(snapshot)
	}
	return a.detector.IsStuck(ctx, internalTask, internalSnapshots)
}

func (a *stuckDetectorAdapter) GetStuckThreshold(taskType string) time.Duration {
	if a.detector == nil {
		return 5 * time.Minute // default threshold
	}
	return a.detector.GetStuckThreshold(taskType)
}

func (a *stuckDetectorAdapter) SetThreshold(taskType string, threshold time.Duration) {
	if a.detector == nil {
		return
	}
	a.detector.SetThreshold(taskType, threshold)
}

// =============================================================================
// NotificationService Adapter - Adapts internal NotificationService to extracted NotificationService
// =============================================================================

type notificationServiceAdapter struct {
	service NotificationService
}

func (a *notificationServiceAdapter) NotifyTaskEvent(ctx context.Context, task *extractedmodels.BackgroundTask, event string, data map[string]interface{}) error {
	if a.service == nil {
		return nil
	}
	internalTask := convertToInternalTask(task)
	return a.service.NotifyTaskEvent(ctx, internalTask, event, data)
}

func (a *notificationServiceAdapter) RegisterSSEClient(ctx context.Context, taskID string, client chan<- []byte) error {
	if a.service == nil {
		return nil
	}
	return a.service.RegisterSSEClient(ctx, taskID, client)
}

func (a *notificationServiceAdapter) UnregisterSSEClient(ctx context.Context, taskID string, client chan<- []byte) error {
	if a.service == nil {
		return nil
	}
	return a.service.UnregisterSSEClient(ctx, taskID, client)
}

func (a *notificationServiceAdapter) RegisterWebSocketClient(ctx context.Context, taskID string, client extractedbackground.WebSocketClient) error {
	if a.service == nil {
		return nil
	}
	// Adapt extracted WebSocketClient to internal WebSocketClient
	// For now, we can create a wrapper or use a generic adapter.
	// Since WebSocketClient interface is small, we can create an adapter.
	// However, to keep things simple, we'll assume the internal NotificationService
	// can handle extracted WebSocketClient via an adapter.
	// We'll create a simple adapter:
	internalClient := &webSocketClientAdapter{client: client}
	return a.service.RegisterWebSocketClient(ctx, taskID, internalClient)
}

func (a *notificationServiceAdapter) BroadcastToTask(ctx context.Context, taskID string, message []byte) error {
	if a.service == nil {
		return nil
	}
	return a.service.BroadcastToTask(ctx, taskID, message)
}

// webSocketClientAdapter adapts extracted WebSocketClient to internal WebSocketClient
type webSocketClientAdapter struct {
	client extractedbackground.WebSocketClient
}

func (a *webSocketClientAdapter) Send(message []byte) error {
	return a.client.Send(message)
}

func (a *webSocketClientAdapter) Close() error {
	return a.client.Close()
}

func (a *webSocketClientAdapter) ID() string {
	return a.client.ID()
}
