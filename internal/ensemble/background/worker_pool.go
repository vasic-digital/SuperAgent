// Package background provides background task processing for HelixAgent.
package background

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"dev.helix.agent/internal/clis"
	"github.com/google/uuid"
)

// WorkerPool manages background agent workers.
type WorkerPool struct {
	db     *sql.DB
	logger *log.Logger

	// Pool configuration
	size        int
	queueSize   int

	// Task queue
	taskQueue   chan *clis.Task

	// Result queue
	resultQueue chan *TaskResult

	// Workers
	workers []*Worker

	// Instance assignment
	instanceAssignments map[string]string // taskID -> instanceID

	// Metrics
	tasksSubmitted   uint64
	tasksCompleted   uint64
	tasksFailed      uint64
	tasksCancelled   uint64

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex

	// Running flag
	running bool
}

// Worker represents a single worker.
type Worker struct {
	id       int
	pool     *WorkerPool
	instance *clis.AgentInstance
	quit     chan struct{}
}

// TaskResult represents the result of a task execution.
type TaskResult struct {
	TaskID   string
	Success  bool
	Result   interface{}
	Error    error
	Duration time.Duration
	WorkerID int
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(size int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		size:                size,
		queueSize:           size * 10,
		taskQueue:           make(chan *clis.Task, size*10),
		resultQueue:         make(chan *TaskResult, size*10),
		workers:             make([]*Worker, 0, size),
		instanceAssignments: make(map[string]string),
		ctx:                 ctx,
		cancel:              cancel,
	}
}

// NewWorkerPoolWithDB creates a pool with database persistence.
func NewWorkerPoolWithDB(db *sql.DB, logger *log.Logger, size int) *WorkerPool {
	pool := NewWorkerPool(size)
	pool.db = db
	pool.logger = logger
	return pool
}

// Start initializes and starts all workers.
func (wp *WorkerPool) Start(ctx context.Context) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.running {
		return fmt.Errorf("worker pool already running")
	}

	// Create workers
	for i := 0; i < wp.size; i++ {
		worker := &Worker{
			id:   i,
			pool: wp,
			quit: make(chan struct{}),
		}
		wp.workers = append(wp.workers, worker)

		wp.wg.Add(1)
		go worker.run()
	}

	// Start result collector
	wp.wg.Add(1)
	go wp.collectResults()

	// Start maintenance loop
	wp.wg.Add(1)
	go wp.maintenanceLoop()

	wp.running = true

	if wp.logger != nil {
		wp.logger.Printf("Worker pool started with %d workers", wp.size)
	}

	return nil
}

// Submit submits a task to the pool.
func (wp *WorkerPool) Submit(ctx context.Context, task *clis.Task) error {
	if !wp.isRunning() {
		return fmt.Errorf("worker pool not running")
	}

	// Set defaults
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	if task.Status == "" {
		task.Status = clis.TaskStatusPending
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	// Persist to database if available
	if wp.db != nil {
		if err := wp.persistTask(ctx, task); err != nil {
			return fmt.Errorf("persist task: %w", err)
		}
	}

	// Submit to queue
	select {
	case wp.taskQueue <- task:
		atomic.AddUint64(&wp.tasksSubmitted, 1)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("task queue full")
	}
}

// SubmitAsync submits a task asynchronously.
func (wp *WorkerPool) SubmitAsync(task *clis.Task) <-chan *TaskResult {
	resultCh := make(chan *TaskResult, 1)

	go func() {
		defer close(resultCh)

		if err := wp.Submit(wp.ctx, task); err != nil {
			resultCh <- &TaskResult{
				TaskID:  task.ID,
				Success: false,
				Error:   err,
			}
			return
		}

		// Wait for result
		for {
			select {
			case result := <-wp.resultQueue:
				if result.TaskID == task.ID {
					resultCh <- result
					return
				}
				// Not our result, put it back
				select {
				case wp.resultQueue <- result:
				default:
				}
			case <-time.After(30 * time.Second):
				resultCh <- &TaskResult{
					TaskID:  task.ID,
					Success: false,
					Error:   fmt.Errorf("timeout waiting for result"),
				}
				return
			}
		}
	}()

	return resultCh
}

// GetTask retrieves a task by ID.
func (wp *WorkerPool) GetTask(ctx context.Context, taskID string) (*clis.Task, error) {
	if wp.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var task clis.Task
	var payloadJSON []byte

	err := wp.db.QueryRowContext(ctx,
		`SELECT id, task_type, task_name, payload, priority, status, 
		        progress_percent, retry_count, max_retries, created_at
		 FROM background_tasks WHERE id = $1`,
		taskID,
	).Scan(
		&task.ID, &task.Type, &task.Name, &payloadJSON, &task.Priority,
		&task.Status, &task.ProgressPercent, &task.RetryCount, &task.MaxRetries,
		&task.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	if err != nil {
		return nil, err
	}

	// Parse payload
	if err := json.Unmarshal(payloadJSON, &task.Payload); err != nil {
		return nil, err
	}

	return &task, nil
}

// CancelTask cancels a pending or running task.
func (wp *WorkerPool) CancelTask(ctx context.Context, taskID string) error {
	if wp.db == nil {
		return fmt.Errorf("database not available")
	}

	_, err := wp.db.ExecContext(ctx,
		"UPDATE background_tasks SET status = $1 WHERE id = $2 AND status IN ($3, $4)",
		clis.TaskStatusCancelled, taskID, clis.TaskStatusPending, clis.TaskStatusAssigned,
	)

	return err
}

// ListTasks returns tasks matching the filter.
func (wp *WorkerPool) ListTasks(
	ctx context.Context,
	status clis.TaskStatus,
	limit int,
) ([]*clis.Task, error) {
	if wp.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	query := `SELECT id, task_type, task_name, payload, priority, status, 
	                 progress_percent, retry_count, max_retries, created_at
	          FROM background_tasks`
	
	var args []interface{}
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, limit)
	}

	rows, err := wp.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*clis.Task
	for rows.Next() {
		var task clis.Task
		var payloadJSON []byte

		err := rows.Scan(
			&task.ID, &task.Type, &task.Name, &payloadJSON, &task.Priority,
			&task.Status, &task.ProgressPercent, &task.RetryCount, &task.MaxRetries,
			&task.CreatedAt,
		)
		if err != nil {
			continue
		}

		json.Unmarshal(payloadJSON, &task.Payload)
		tasks = append(tasks, &task)
	}

	return tasks, rows.Err()
}

// GetStats returns pool statistics.
func (wp *WorkerPool) GetStats() map[string]interface{} {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	return map[string]interface{}{
		"size":             wp.size,
		"running":          wp.running,
		"queue_depth":      len(wp.taskQueue),
		"queue_capacity":   cap(wp.taskQueue),
		"tasks_submitted":  atomic.LoadUint64(&wp.tasksSubmitted),
		"tasks_completed":  atomic.LoadUint64(&wp.tasksCompleted),
		"tasks_failed":     atomic.LoadUint64(&wp.tasksFailed),
		"tasks_cancelled":  atomic.LoadUint64(&wp.tasksCancelled),
	}
}

// Stop shuts down the worker pool.
func (wp *WorkerPool) Stop() error {
	wp.mu.Lock()
	if !wp.running {
		wp.mu.Unlock()
		return nil
	}
	wp.running = false
	wp.mu.Unlock()

	// Signal cancellation
	wp.cancel()

	// Signal all workers to quit
	for _, worker := range wp.workers {
		close(worker.quit)
	}

	// Wait for all goroutines
	wp.wg.Wait()

	// Close channels
	close(wp.taskQueue)
	close(wp.resultQueue)

	if wp.logger != nil {
		wp.logger.Printf("Worker pool stopped")
	}

	return nil
}

// Internal methods

func (wp *WorkerPool) isRunning() bool {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.running
}

func (wp *WorkerPool) persistTask(ctx context.Context, task *clis.Task) error {
	payloadJSON, err := json.Marshal(task.Payload)
	if err != nil {
		return err
	}

	_, err = wp.db.ExecContext(ctx,
		`INSERT INTO background_tasks (
			id, task_type, task_name, payload, priority, status,
			retry_count, max_retries, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			retry_count = EXCLUDED.retry_count,
			progress_percent = EXCLUDED.progress_percent`,
		task.ID, task.Type, task.Name, payloadJSON, task.Priority, task.Status,
		task.RetryCount, task.MaxRetries, task.CreatedAt,
		task.CreatedAt.Add(24*time.Hour), // Default 24h expiration
	)

	return err
}

func (wp *WorkerPool) updateTaskStatus(
	ctx context.Context,
	taskID string,
	status clis.TaskStatus,
	progress int,
	result interface{},
	err error,
) error {
	if wp.db == nil {
		return nil
	}

	resultJSON, _ := json.Marshal(result)
	var errorMsg *string
	if err != nil {
		msg := err.Error()
		errorMsg = &msg
	}

	var startedAt, completedAt interface{}
	if status == clis.TaskStatusRunning {
		startedAt = time.Now()
	}
	if status == clis.TaskStatusCompleted || status == clis.TaskStatusFailed {
		completedAt = time.Now()
	}

	_, dbErr := wp.db.ExecContext(ctx,
		`UPDATE background_tasks SET
			status = $1,
			progress_percent = $2,
			result = $3,
			error_message = $4,
			started_at = COALESCE($5, started_at),
			completed_at = $6
		 WHERE id = $7`,
		status, progress, resultJSON, errorMsg, startedAt, completedAt, taskID,
	)

	return dbErr
}

func (w *Worker) run() {
	defer w.pool.wg.Done()

	for {
		select {
		case task, ok := <-w.pool.taskQueue:
			if !ok {
				return // Channel closed
			}

			// Check if task was cancelled
			if task.Status == clis.TaskStatusCancelled {
				continue
			}

			// Execute task
			result := w.execute(task)

			// Send result
			select {
			case w.pool.resultQueue <- result:
			case <-w.pool.ctx.Done():
				return
			}

			// Update metrics
			if result.Success {
				atomic.AddUint64(&w.pool.tasksCompleted, 1)
			} else {
				atomic.AddUint64(&w.pool.tasksFailed, 1)
			}

		case <-w.quit:
			return

		case <-w.pool.ctx.Done():
			return
		}
	}
}

func (w *Worker) execute(task *clis.Task) *TaskResult {
	start := time.Now()

	// Update status to running
	if w.pool.db != nil {
		w.pool.updateTaskStatus(
			w.pool.ctx,
			task.ID,
			clis.TaskStatusRunning,
			0,
			nil,
			nil,
		)
	}

	// Route to appropriate handler
	var result interface{}
	var err error

	switch task.Type {
	case clis.TaskTypeGitOperation:
		result, err = w.executeGitOperation(task)
	case clis.TaskTypeCodeAnalysis:
		result, err = w.executeCodeAnalysis(task)
	case clis.TaskTypeDocumentation:
		result, err = w.executeDocumentation(task)
	case clis.TaskTypeTesting:
		result, err = w.executeTesting(task)
	case clis.TaskTypeLinting:
		result, err = w.executeLinting(task)
	case clis.TaskTypeBuild:
		result, err = w.executeBuild(task)
	case clis.TaskTypeCodeReview:
		result, err = w.executeCodeReview(task)
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
	}

	duration := time.Since(start)

	// Update final status
	status := clis.TaskStatusCompleted
	if err != nil {
		status = clis.TaskStatusFailed
	}

	if w.pool.db != nil {
		w.pool.updateTaskStatus(
			w.pool.ctx,
			task.ID,
			status,
			100,
			result,
			err,
		)
	}

	return &TaskResult{
		TaskID:   task.ID,
		Success:  err == nil,
		Result:   result,
		Error:    err,
		Duration: duration,
		WorkerID: w.id,
	}
}

// Task execution handlers

func (w *Worker) executeGitOperation(task *clis.Task) (interface{}, error) {
	// Implementation would use Aider git integration
	return map[string]string{"status": "git_operation_executed"}, nil
}

func (w *Worker) executeCodeAnalysis(task *clis.Task) (interface{}, error) {
	// Implementation would use repo map and analysis tools
	return map[string]string{"status": "code_analysis_completed"}, nil
}

func (w *Worker) executeDocumentation(task *clis.Task) (interface{}, error) {
	return map[string]string{"status": "documentation_generated"}, nil
}

func (w *Worker) executeTesting(task *clis.Task) (interface{}, error) {
	return map[string]string{"status": "tests_executed"}, nil
}

func (w *Worker) executeLinting(task *clis.Task) (interface{}, error) {
	return map[string]string{"status": "linting_completed"}, nil
}

func (w *Worker) executeBuild(task *clis.Task) (interface{}, error) {
	return map[string]string{"status": "build_completed"}, nil
}

func (w *Worker) executeCodeReview(task *clis.Task) (interface{}, error) {
	return map[string]string{"status": "code_review_completed"}, nil
}

func (wp *WorkerPool) collectResults() {
	defer wp.wg.Done()

	for result := range wp.resultQueue {
		// Could trigger callbacks here
		if wp.logger != nil && !result.Success {
			wp.logger.Printf("Task %s failed: %v", result.TaskID, result.Error)
		}
	}
}

func (wp *WorkerPool) maintenanceLoop() {
	defer wp.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wp.cleanupExpiredTasks()
			wp.retryFailedTasks()

		case <-wp.ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) cleanupExpiredTasks() {
	if wp.db == nil {
		return
	}

	_, err := wp.db.ExecContext(wp.ctx,
		`UPDATE background_tasks 
		 SET status = $1
		 WHERE status IN ($2, $3) AND expires_at < NOW()`,
		clis.TaskStatusExpired,
		clis.TaskStatusPending, clis.TaskStatusAssigned,
	)

	if err != nil && wp.logger != nil {
		wp.logger.Printf("Error cleaning expired tasks: %v", err)
	}
}

func (wp *WorkerPool) retryFailedTasks() {
	if wp.db == nil {
		return
	}

	rows, err := wp.db.QueryContext(wp.ctx,
		`SELECT id, retry_count, max_retries FROM background_tasks
		 WHERE status = $1 AND retry_count < max_retries`,
		clis.TaskStatusFailed,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var retryCount, maxRetries int
		if err := rows.Scan(&id, &retryCount, &maxRetries); err != nil {
			continue
		}

		// Reset to pending for retry
		wp.db.ExecContext(wp.ctx,
			`UPDATE background_tasks 
			 SET status = $1, retry_count = retry_count + 1
			 WHERE id = $2`,
			clis.TaskStatusPending, id,
		)
	}
}
