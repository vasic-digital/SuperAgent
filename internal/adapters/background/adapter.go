// Package background provides adapters bridging HelixAgent's internal background
// types to the generic digital.vasic.background module.
//
// The adapter package maintains backward compatibility with code using
// dev.helix.agent/internal/background while delegating core operations
// to digital.vasic.background.
package background

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	internalbackground "dev.helix.agent/internal/background"
	internalmodels "dev.helix.agent/internal/models"
	extractedbackground "digital.vasic.background"
	extractedmodels "digital.vasic.models"
)

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
func (a *TaskQueueAdapter) Enqueue(ctx context.Context, task *internalmodels.BackgroundTask) error {
	extractedTask := convertToExtractedTask(task)
	return a.queue.Enqueue(ctx, extractedTask)
}

// Dequeue atomically retrieves and claims a task from the queue
func (a *TaskQueueAdapter) Dequeue(ctx context.Context, workerID string, requirements internalbackground.ResourceRequirements) (*internalmodels.BackgroundTask, error) {
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
func (a *TaskQueueAdapter) Peek(ctx context.Context, count int) ([]*internalmodels.BackgroundTask, error) {
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
func (a *TaskQueueAdapter) GetQueueDepth(ctx context.Context) (map[internalmodels.TaskPriority]int64, error) {
	extractedDepth, err := a.queue.GetQueueDepth(ctx)
	if err != nil {
		return nil, err
	}
	depth := make(map[internalmodels.TaskPriority]int64)
	for priority, count := range extractedDepth {
		depth[internalmodels.TaskPriority(priority)] = count
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
func (a *WorkerPoolAdapter) RegisterExecutor(taskType string, executor internalbackground.TaskExecutor) {
	// Adapt internal TaskExecutor to extracted TaskExecutor
	extractedExecutor := NewTaskExecutorAdapter(executor)
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
func (a *WorkerPoolAdapter) GetWorkerStatus() []internalbackground.WorkerStatus {
	extractedStatus := a.pool.GetWorkerStatus()
	status := make([]internalbackground.WorkerStatus, len(extractedStatus))
	for i, s := range extractedStatus {
		status[i] = internalbackground.WorkerStatus{
			ID:              s.ID,
			Status:          s.Status,
			CurrentTask:     convertToInternalTask(s.CurrentTask),
			StartedAt:       s.StartedAt,
			LastActivity:    s.LastActivity,
			TasksCompleted:  s.TasksCompleted,
			TasksFailed:     s.TasksFailed,
			AvgTaskDuration: s.AvgTaskDuration,
		}
	}
	return status
}

// Scale manually adjusts the worker count
func (a *WorkerPoolAdapter) Scale(targetCount int) error {
	return a.pool.Scale(targetCount)
}

// =============================================================================
// TaskExecutor Adapter - Adapts internal TaskExecutor to extracted TaskExecutor
// =============================================================================

// TaskExecutorAdapter wraps internal TaskExecutor to implement extracted TaskExecutor interface
type TaskExecutorAdapter struct {
	executor internalbackground.TaskExecutor
}

// NewTaskExecutorAdapter creates a new adapter wrapping an internal TaskExecutor
func NewTaskExecutorAdapter(executor internalbackground.TaskExecutor) *TaskExecutorAdapter {
	return &TaskExecutorAdapter{executor: executor}
}

// Execute runs the task with context and progress reporting
func (a *TaskExecutorAdapter) Execute(ctx context.Context, task *extractedmodels.BackgroundTask, reporter extractedbackground.ProgressReporter) error {
	internalTask := convertToInternalTask(task)
	internalReporter := NewProgressReporterAdapter(reporter)
	return a.executor.Execute(ctx, internalTask, internalReporter)
}

// CanPause returns whether this task type supports pause/resume
func (a *TaskExecutorAdapter) CanPause() bool {
	return a.executor.CanPause()
}

// Pause saves checkpoint for later resume
func (a *TaskExecutorAdapter) Pause(ctx context.Context, task *extractedmodels.BackgroundTask) ([]byte, error) {
	internalTask := convertToInternalTask(task)
	return a.executor.Pause(ctx, internalTask)
}

// Resume restores from checkpoint
func (a *TaskExecutorAdapter) Resume(ctx context.Context, task *extractedmodels.BackgroundTask, checkpoint []byte) error {
	internalTask := convertToInternalTask(task)
	return a.executor.Resume(ctx, internalTask, checkpoint)
}

// Cancel handles graceful cancellation
func (a *TaskExecutorAdapter) Cancel(ctx context.Context, task *extractedmodels.BackgroundTask) error {
	internalTask := convertToInternalTask(task)
	return a.executor.Cancel(ctx, internalTask)
}

// GetResourceRequirements returns resource needs for this executor
func (a *TaskExecutorAdapter) GetResourceRequirements() extractedbackground.ResourceRequirements {
	internalReq := a.executor.GetResourceRequirements()
	return extractedbackground.ResourceRequirements{
		CPUCores: internalReq.CPUCores,
		MemoryMB: internalReq.MemoryMB,
		DiskMB:   internalReq.DiskMB,
		GPUCount: internalReq.GPUCount,
		Priority: convertTaskPriority(internalReq.Priority),
	}
}

// =============================================================================
// ProgressReporter Adapter - Adapts extracted ProgressReporter to internal ProgressReporter
// =============================================================================

// ProgressReporterAdapter wraps extracted ProgressReporter to implement internal ProgressReporter interface
type ProgressReporterAdapter struct {
	reporter extractedbackground.ProgressReporter
}

// NewProgressReporterAdapter creates a new adapter wrapping an extracted ProgressReporter
func NewProgressReporterAdapter(reporter extractedbackground.ProgressReporter) *ProgressReporterAdapter {
	return &ProgressReporterAdapter{reporter: reporter}
}

// ReportProgress reports task progress (0-100 percentage)
func (a *ProgressReporterAdapter) ReportProgress(percent float64, message string) error {
	return a.reporter.ReportProgress(percent, message)
}

// ReportHeartbeat sends a heartbeat to indicate the task is still alive
func (a *ProgressReporterAdapter) ReportHeartbeat() error {
	return a.reporter.ReportHeartbeat()
}

// ReportCheckpoint saves a checkpoint for pause/resume capability
func (a *ProgressReporterAdapter) ReportCheckpoint(data []byte) error {
	return a.reporter.ReportCheckpoint(data)
}

// ReportMetrics reports custom metrics from the task
func (a *ProgressReporterAdapter) ReportMetrics(metrics map[string]interface{}) error {
	return a.reporter.ReportMetrics(metrics)
}

// ReportLog reports a log entry from the task
func (a *ProgressReporterAdapter) ReportLog(level, message string, fields map[string]interface{}) error {
	return a.reporter.ReportLog(level, message, fields)
}

// =============================================================================
// ResourceMonitor Adapter - Adapts extracted ResourceMonitor to internal ResourceMonitor interface
// =============================================================================

// ResourceMonitorAdapter wraps the extracted ResourceMonitor to implement internal ResourceMonitor interface
type ResourceMonitorAdapter struct {
	monitor extractedbackground.ResourceMonitor
}

// NewResourceMonitorAdapter creates a new adapter wrapping an extracted ResourceMonitor
func NewResourceMonitorAdapter(monitor extractedbackground.ResourceMonitor) *ResourceMonitorAdapter {
	return &ResourceMonitorAdapter{monitor: monitor}
}

// GetSystemResources returns current system resource usage
func (a *ResourceMonitorAdapter) GetSystemResources() (*internalbackground.SystemResources, error) {
	extractedResources, err := a.monitor.GetSystemResources()
	if err != nil {
		return nil, err
	}
	return convertToInternalSystemResources(extractedResources), nil
}

// GetProcessResources returns resource usage for a specific process
func (a *ResourceMonitorAdapter) GetProcessResources(pid int) (*internalmodels.ResourceSnapshot, error) {
	extractedSnapshot, err := a.monitor.GetProcessResources(pid)
	if err != nil {
		return nil, err
	}
	return convertToInternalResourceSnapshot(extractedSnapshot), nil
}

// StartMonitoring begins periodic monitoring of a process
func (a *ResourceMonitorAdapter) StartMonitoring(taskID string, pid int, interval time.Duration) error {
	return a.monitor.StartMonitoring(taskID, pid, interval)
}

// StopMonitoring stops monitoring a process
func (a *ResourceMonitorAdapter) StopMonitoring(taskID string) error {
	return a.monitor.StopMonitoring(taskID)
}

// GetLatestSnapshot returns the most recent snapshot for a task
func (a *ResourceMonitorAdapter) GetLatestSnapshot(taskID string) (*internalmodels.ResourceSnapshot, error) {
	extractedSnapshot, err := a.monitor.GetLatestSnapshot(taskID)
	if err != nil {
		return nil, err
	}
	return convertToInternalResourceSnapshot(extractedSnapshot), nil
}

// IsResourceAvailable checks if system has enough resources
func (a *ResourceMonitorAdapter) IsResourceAvailable(requirements internalbackground.ResourceRequirements) bool {
	// Convert internal ResourceRequirements to extracted ResourceRequirements
	extractedRequirements := extractedbackground.ResourceRequirements{
		CPUCores: requirements.CPUCores,
		MemoryMB: requirements.MemoryMB,
		DiskMB:   requirements.DiskMB,
		GPUCount: requirements.GPUCount,
		Priority: convertTaskPriority(requirements.Priority),
	}
	return a.monitor.IsResourceAvailable(extractedRequirements)
}

// =============================================================================
// StuckDetector Adapter - Adapts extracted StuckDetector to internal StuckDetector interface
// =============================================================================

// StuckDetectorAdapter wraps the extracted StuckDetector to implement internal StuckDetector interface
type StuckDetectorAdapter struct {
	detector         extractedbackground.StuckDetector
	concreteDetector *extractedbackground.DefaultStuckDetector
}

// NewStuckDetectorAdapter creates a new adapter wrapping an extracted StuckDetector
func NewStuckDetectorAdapter(detector extractedbackground.StuckDetector) *StuckDetectorAdapter {
	adapter := &StuckDetectorAdapter{detector: detector}
	// Try to store concrete detector for AnalyzeTask method
	if concrete, ok := detector.(*extractedbackground.DefaultStuckDetector); ok {
		adapter.concreteDetector = concrete
	}
	return adapter
}

// IsStuck determines if a task is stuck based on various criteria
func (a *StuckDetectorAdapter) IsStuck(ctx context.Context, task *internalmodels.BackgroundTask, snapshots []*internalmodels.ResourceSnapshot) (bool, string) {
	extractedTask := convertToExtractedTask(task)
	extractedSnapshots := make([]*extractedmodels.ResourceSnapshot, len(snapshots))
	for i, snapshot := range snapshots {
		extractedSnapshots[i] = convertToExtractedResourceSnapshot(snapshot)
	}
	return a.detector.IsStuck(ctx, extractedTask, extractedSnapshots)
}

// GetStuckThreshold returns the stuck detection threshold for a task type
func (a *StuckDetectorAdapter) GetStuckThreshold(taskType string) time.Duration {
	return a.detector.GetStuckThreshold(taskType)
}

// SetThreshold sets a custom threshold for a task type
func (a *StuckDetectorAdapter) SetThreshold(taskType string, threshold time.Duration) {
	a.detector.SetThreshold(taskType, threshold)
}

// AnalyzeTask performs detailed stuck analysis on a task
func (a *StuckDetectorAdapter) AnalyzeTask(ctx context.Context, task *internalmodels.BackgroundTask, snapshots []*internalmodels.ResourceSnapshot) *internalbackground.StuckAnalysis {
	if a.concreteDetector == nil {
		// Fallback: create a basic analysis using IsStuck
		isStuck, reason := a.IsStuck(ctx, task, snapshots)
		analysis := &internalbackground.StuckAnalysis{
			IsStuck:         isStuck,
			Reason:          reason,
			HeartbeatStatus: internalbackground.HeartbeatStatus{},
			ResourceStatus:  internalbackground.ResourceStatus{},
			ActivityStatus:  internalbackground.ActivityStatus{},
			Recommendations: []string{},
		}
		return analysis
	}
	extractedTask := convertToExtractedTask(task)
	extractedSnapshots := make([]*extractedmodels.ResourceSnapshot, len(snapshots))
	for i, snapshot := range snapshots {
		extractedSnapshots[i] = convertToExtractedResourceSnapshot(snapshot)
	}
	extractedAnalysis := a.concreteDetector.AnalyzeTask(ctx, extractedTask, extractedSnapshots)
	return convertToInternalStuckAnalysis(extractedAnalysis)
}

// =============================================================================
// NotificationService Adapter - Adapts internal NotificationService to extracted NotificationService interface
// =============================================================================

// NotificationServiceAdapter wraps internal NotificationService to implement extracted NotificationService interface
type NotificationServiceAdapter struct {
	notifier internalbackground.NotificationService
}

// NewNotificationServiceAdapter creates a new adapter wrapping an internal NotificationService
func NewNotificationServiceAdapter(notifier internalbackground.NotificationService) *NotificationServiceAdapter {
	return &NotificationServiceAdapter{notifier: notifier}
}

// NotifyTaskEvent sends notifications for a task event
func (a *NotificationServiceAdapter) NotifyTaskEvent(ctx context.Context, task *extractedmodels.BackgroundTask, event string, data map[string]interface{}) error {
	internalTask := convertToInternalTask(task)
	return a.notifier.NotifyTaskEvent(ctx, internalTask, event, data)
}

// RegisterSSEClient registers a client for SSE notifications
func (a *NotificationServiceAdapter) RegisterSSEClient(ctx context.Context, taskID string, client chan<- []byte) error {
	return a.notifier.RegisterSSEClient(ctx, taskID, client)
}

// UnregisterSSEClient removes an SSE client
func (a *NotificationServiceAdapter) UnregisterSSEClient(ctx context.Context, taskID string, client chan<- []byte) error {
	return a.notifier.UnregisterSSEClient(ctx, taskID, client)
}

// RegisterWebSocketClient registers a WebSocket client
func (a *NotificationServiceAdapter) RegisterWebSocketClient(ctx context.Context, taskID string, client extractedbackground.WebSocketClient) error {
	// Adapt extracted WebSocketClient to internal WebSocketClient
	internalClient := NewWebSocketClientAdapter(client)
	return a.notifier.RegisterWebSocketClient(ctx, taskID, internalClient)
}

// BroadcastToTask broadcasts a message to all clients watching a task
func (a *NotificationServiceAdapter) BroadcastToTask(ctx context.Context, taskID string, message []byte) error {
	return a.notifier.BroadcastToTask(ctx, taskID, message)
}

// =============================================================================
// WebSocketClient Adapter - Adapts extracted WebSocketClient to internal WebSocketClient interface
// =============================================================================

// WebSocketClientAdapter wraps extracted WebSocketClient to implement internal WebSocketClient interface
type WebSocketClientAdapter struct {
	client extractedbackground.WebSocketClient
}

// NewWebSocketClientAdapter creates a new adapter wrapping an extracted WebSocketClient
func NewWebSocketClientAdapter(client extractedbackground.WebSocketClient) *WebSocketClientAdapter {
	return &WebSocketClientAdapter{client: client}
}

// Send sends data to the client
func (a *WebSocketClientAdapter) Send(data []byte) error {
	return a.client.Send(data)
}

// Close closes the connection
func (a *WebSocketClientAdapter) Close() error {
	return a.client.Close()
}

// ID returns the client identifier
func (a *WebSocketClientAdapter) ID() string {
	return a.client.ID()
}

// =============================================================================
// NoOpNotificationService provides a no-operation implementation of NotificationService
// =============================================================================

type NoOpNotificationService struct{}

func (n *NoOpNotificationService) NotifyTaskEvent(ctx context.Context, task *extractedmodels.BackgroundTask, event string, data map[string]interface{}) error {
	return nil
}

func (n *NoOpNotificationService) RegisterSSEClient(ctx context.Context, taskID string, client chan<- []byte) error {
	return nil
}

func (n *NoOpNotificationService) UnregisterSSEClient(ctx context.Context, taskID string, client chan<- []byte) error {
	return nil
}

func (n *NoOpNotificationService) RegisterWebSocketClient(ctx context.Context, taskID string, client extractedbackground.WebSocketClient) error {
	return nil
}

func (n *NoOpNotificationService) BroadcastToTask(ctx context.Context, taskID string, message []byte) error {
	return nil
}

// =============================================================================
// Factory functions for creating extracted module instances
// =============================================================================

// NewPostgresTaskQueue creates a new PostgreSQL-backed task queue using extracted module
func NewPostgresTaskQueue(repository internalbackground.TaskRepository, logger interface{}) *TaskQueueAdapter {
	// Adapt internal repository to extracted repository
	adaptedRepo := NewTaskRepositoryAdapter(repository)
	// Create extracted PostgresTaskQueue using constructor
	var logrusLogger *logrus.Logger
	if l, ok := logger.(*logrus.Logger); ok {
		logrusLogger = l
	} else {
		// Create a default logger if not provided
		logrusLogger = logrus.New()
	}
	extractedQueue := extractedbackground.NewPostgresTaskQueue(adaptedRepo, logrusLogger)
	return NewTaskQueueAdapter(extractedQueue)
}

// NewProcessResourceMonitor creates a new process resource monitor using extracted module
func NewProcessResourceMonitor(repository internalbackground.TaskRepository, logger interface{}) *ResourceMonitorAdapter {
	adaptedRepo := NewTaskRepositoryAdapter(repository)
	var logrusLogger *logrus.Logger
	if l, ok := logger.(*logrus.Logger); ok {
		logrusLogger = l
	} else {
		logrusLogger = logrus.New()
	}
	extractedMonitor := extractedbackground.NewProcessResourceMonitor(adaptedRepo, logrusLogger)
	return NewResourceMonitorAdapter(extractedMonitor)
}

// NewDefaultStuckDetector creates a new stuck detector using extracted module
func NewDefaultStuckDetector(logger interface{}) *StuckDetectorAdapter {
	var logrusLogger *logrus.Logger
	if l, ok := logger.(*logrus.Logger); ok {
		logrusLogger = l
	} else {
		logrusLogger = logrus.New()
	}
	extractedDetector := extractedbackground.NewDefaultStuckDetector(logrusLogger)
	return NewStuckDetectorAdapter(extractedDetector)
}
