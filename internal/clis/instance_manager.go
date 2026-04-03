// Package clis provides CLI agent integration for HelixAgent.
package clis

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// InstanceManager manages the lifecycle of CLI agent instances.
type InstanceManager struct {
	db     *sql.DB
	logger *log.Logger

	// Instance pools by type for efficient reuse
	pools map[AgentType]*InstancePool

	// Active instances (both idle and active)
	instances map[string]*AgentInstance
	mu        sync.RWMutex

	// Event bus for instance events
	eventBus *EventBus

	// Background workers
	workerPool *WorkerPool

	// Health check control
	healthCheckStop chan struct{}

	// Metrics
	createdCount   uint64
	destroyedCount uint64
}

// NewInstanceManager creates a new instance manager.
func NewInstanceManager(db *sql.DB, logger *log.Logger) (*InstanceManager, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection required")
	}

	if logger == nil {
		logger = log.Default()
	}

	im := &InstanceManager{
		db:              db,
		logger:          logger,
		pools:           make(map[AgentType]*InstancePool),
		instances:       make(map[string]*AgentInstance),
		eventBus:        NewEventBus(),
		workerPool:      NewWorkerPool(100),
		healthCheckStop: make(chan struct{}),
	}

	// Start background health checks
	go im.healthCheckLoop()

	// Recover existing instances from database
	if err := im.recoverInstances(context.Background()); err != nil {
		logger.Printf("Warning: failed to recover instances: %v", err)
	}

	return im, nil
}

// CreateInstance creates a new agent instance.
func (m *InstanceManager) CreateInstance(
	ctx context.Context,
	agentType AgentType,
	config InstanceConfig,
	provider ProviderConfig,
) (*AgentInstance, error) {
	// Check if this agent type is available
	if !m.IsAgentTypeAvailable(agentType) {
		return nil, fmt.Errorf("agent type %s is not available", agentType)
	}

	// Generate unique ID and name
	instanceID := uuid.New().String()
	instanceName := fmt.Sprintf("%s-%s", agentType, generateShortID())

	// Create instance object
	instance := &AgentInstance{
		ID:        instanceID,
		Type:      agentType,
		Name:      instanceName,
		Status:    StatusCreating,
		Config:    config,
		Provider:  provider,
		Resources: ResourceLimits{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		RequestCh:  make(chan *Request, 10),
		ResponseCh: make(chan *Response, 10),
		EventCh:    make(chan *Event, 10),
	}

	// Initialize type-specific components
	if err := m.initializeInstance(ctx, instance); err != nil {
		return nil, fmt.Errorf("initialize instance: %w", err)
	}

	// Persist to database
	if err := m.persistInstance(ctx, instance); err != nil {
		return nil, fmt.Errorf("persist instance: %w", err)
	}

	// Register in memory
	m.mu.Lock()
	m.instances[instance.ID] = instance
	m.mu.Unlock()

	// Update metrics
	atomic.AddUint64(&m.createdCount, 1)

	// Start event loops
	go m.instanceEventLoop(instance)
	go m.instanceHealthLoop(instance)

	// Mark as idle once initialized
	instance.Status = StatusIdle
	instance.Health = HealthHealthy
	instance.UpdatedAt = time.Now()
	now := time.Now()
	instance.StartedAt = &now

	// Update database status
	_, err := m.db.ExecContext(ctx,
		"UPDATE agent_instances SET status = $1, health_status = $2, started_at = NOW() WHERE id = $3",
		StatusIdle, HealthHealthy, instance.ID,
	)
	if err != nil {
		m.logger.Printf("Warning: failed to update instance status: %v", err)
	}

	// Publish event
	m.eventBus.Publish(&Event{
		ID:        uuid.New().String(),
		Type:      EventTypeStatus,
		Source:    instance.ID,
		Payload:   map[string]string{"status": string(StatusIdle)},
		Timestamp: time.Now(),
	})

	m.logger.Printf("Created instance %s of type %s", instance.ID, agentType)

	return instance, nil
}

// AcquireInstance gets an instance from the pool or creates a new one.
func (m *InstanceManager) AcquireInstance(
	ctx context.Context,
	agentType AgentType,
) (*AgentInstance, error) {
	// Try to get from pool first
	if pool, ok := m.pools[agentType]; ok {
		if instance, err := pool.Acquire(ctx); err == nil {
			m.logger.Printf("Acquired instance %s from pool", instance.ID)
			return instance, nil
		}
	}

	// Create new instance if pool is empty
	m.logger.Printf("Pool empty for %s, creating new instance", agentType)
	return m.CreateInstance(ctx, agentType, DefaultInstanceConfig(), ProviderConfig{})
}

// ReleaseInstance returns an instance to the pool or terminates it.
func (m *InstanceManager) ReleaseInstance(ctx context.Context, instance *AgentInstance) error {
	if instance == nil {
		return nil
	}

	// Reset instance state
	instance.SessionID = ""
	instance.TaskID = ""
	instance.Status = StatusIdle
	instance.UpdatedAt = time.Now()

	// Try to return to pool
	if pool, ok := m.pools[instance.Type]; ok {
		if err := pool.Release(instance); err == nil {
			m.logger.Printf("Released instance %s to pool", instance.ID)
			return nil
		}
	}

	// Terminate if pool is full or doesn't exist
	return m.TerminateInstance(ctx, instance.ID)
}

// GetInstance retrieves an instance by ID.
func (m *InstanceManager) GetInstance(id string) (*AgentInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, ok := m.instances[id]
	if !ok {
		return nil, fmt.Errorf("instance %s not found", id)
	}

	return instance, nil
}

// ListInstances returns all instances matching the filter.
func (m *InstanceManager) ListInstances(status InstanceStatus, agentType AgentType) []*AgentInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*AgentInstance
	for _, instance := range m.instances {
		if status != "" && instance.Status != status {
			continue
		}
		if agentType != "" && instance.Type != agentType {
			continue
		}
		result = append(result, instance)
	}

	return result
}

// TerminateInstance terminates an instance.
func (m *InstanceManager) TerminateInstance(ctx context.Context, id string) error {
	m.mu.Lock()
	instance, exists := m.instances[id]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("instance %s not found", id)
	}

	m.logger.Printf("Terminating instance %s", id)

	// Update status
	instance.Status = StatusTerminating
	instance.UpdatedAt = time.Now()

	// Perform type-specific cleanup
	if err := m.cleanupInstance(ctx, instance); err != nil {
		m.logger.Printf("Warning: cleanup error for %s: %v", id, err)
	}

	// Close channels
	close(instance.RequestCh)
	close(instance.ResponseCh)
	close(instance.EventCh)

	// Update database
	_, err := m.db.ExecContext(ctx,
		`UPDATE agent_instances 
		 SET status = $1, terminated_at = NOW(), updated_at = NOW() 
		 WHERE id = $2`,
		StatusTerminated, id,
	)
	if err != nil {
		m.logger.Printf("Warning: failed to update termination status: %v", err)
	}

	// Remove from memory
	m.mu.Lock()
	delete(m.instances, id)
	m.mu.Unlock()

	// Update metrics
	atomic.AddUint64(&m.destroyedCount, 1)

	m.logger.Printf("Terminated instance %s", id)

	return nil
}

// SendRequest sends a request to an instance and waits for response.
func (m *InstanceManager) SendRequest(
	ctx context.Context,
	instanceID string,
	req *Request,
) (*Response, error) {
	instance, err := m.GetInstance(instanceID)
	if err != nil {
		return nil, err
	}

	if !instance.CanAcceptWork() {
		return nil, fmt.Errorf("instance %s cannot accept work (status: %s, health: %s)",
			instanceID, instance.Status, instance.Health)
	}

	// Set instance as active
	instance.Status = StatusActive
	instance.UpdatedAt = time.Now()

	// Send request
	select {
	case instance.RequestCh <- req:
		// Request sent
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout sending request to instance")
	}

	// Wait for response with timeout
	ctx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	select {
	case resp := <-instance.ResponseCh:
		// Update metrics
		instance.RequestsProcessed++
		instance.TotalExecTimeMs += int64(resp.Duration.Milliseconds())
		if !resp.Success {
			instance.ErrorsCount++
		}

		// Mark idle after processing
		instance.Status = StatusIdle
		instance.UpdatedAt = time.Now()

		return resp, nil

	case <-ctx.Done():
		instance.ErrorsCount++
		return nil, fmt.Errorf("request timeout: %w", ctx.Err())
	}
}

// BroadcastRequest sends a request to all instances of a specific type.
func (m *InstanceManager) BroadcastRequest(
	ctx context.Context,
	agentType AgentType,
	req *Request,
) map[string]*Response {
	instances := m.ListInstances(StatusIdle, agentType)
	
	results := make(map[string]*Response)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, instance := range instances {
		wg.Add(1)
		go func(inst *AgentInstance) {
			defer wg.Done()
			
			resp, err := m.SendRequest(ctx, inst.ID, req)
			mu.Lock()
			if err != nil {
				results[inst.ID] = &Response{
					RequestID: req.ID,
					Success:   false,
					Error: &ErrorDetail{
						Code:    "BROADCAST_ERROR",
						Message: err.Error(),
					},
				}
			} else {
				results[inst.ID] = resp
			}
			mu.Unlock()
		}(instance)
	}

	wg.Wait()
	return results
}

// IsAgentTypeAvailable checks if an agent type is available.
func (m *InstanceManager) IsAgentTypeAvailable(agentType AgentType) bool {
	// Check if there's a pool for this type
	// For now, allow all types defined in the enum
	switch agentType {
	case TypeAider, TypeClaudeCode, TypeCodex, TypeCline,
		TypeOpenHands, TypeKiro, TypeContinue, TypeHelixAgent:
		return true
	default:
		// Could check database or configuration here
		return false
	}
}

// GetMetrics returns manager metrics.
func (m *InstanceManager) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	activeCount := len(m.instances)
	m.mu.RUnlock()

	return map[string]interface{}{
		"created_total":   atomic.LoadUint64(&m.createdCount),
		"destroyed_total": atomic.LoadUint64(&m.destroyedCount),
		"active_count":    activeCount,
		"pool_count":      len(m.pools),
	}
}

// Close shuts down the instance manager.
func (m *InstanceManager) Close() error {
	// Stop health checks
	close(m.healthCheckStop)

	// Terminate all instances
	ctx := context.Background()
	m.mu.RLock()
	instances := make([]*AgentInstance, 0, len(m.instances))
	for _, inst := range m.instances {
		instances = append(instances, inst)
	}
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, instance := range instances {
		wg.Add(1)
		go func(inst *AgentInstance) {
			defer wg.Done()
			if err := m.TerminateInstance(ctx, inst.ID); err != nil {
				m.logger.Printf("Error terminating instance %s: %v", inst.ID, err)
			}
		}(instance)
	}

	wg.Wait()
	return nil
}

// Internal methods

func (m *InstanceManager) initializeInstance(ctx context.Context, inst *AgentInstance) error {
	// Type-specific initialization
	switch inst.Type {
	case TypeAider:
		// Initialize Aider-specific components
		inst.State = map[string]interface{}{
			"repo_map_enabled": true,
			"diff_format":      "search_replace",
		}

	case TypeClaudeCode:
		// Initialize Claude Code-specific components
		inst.State = map[string]interface{}{
			"terminal_enabled": true,
			"tool_use_enabled": true,
		}

	case TypeCodex:
		inst.State = map[string]interface{}{
			"interpreter_enabled": true,
			"reasoning_enabled":   true,
		}

	case TypeCline:
		inst.State = map[string]interface{}{
			"browser_enabled": true,
			"autonomy_enabled": true,
		}

	case TypeOpenHands:
		inst.State = map[string]interface{}{
			"sandbox_enabled":  true,
			"security_level":   "high",
		}

	case TypeKiro:
		inst.State = map[string]interface{}{
			"memory_enabled": true,
		}

	case TypeContinue:
		inst.State = map[string]interface{}{
			"lsp_enabled": true,
		}

	case TypeHelixAgent:
		// Native HelixAgent instance
		inst.State = map[string]interface{}{
			"native": true,
		}

	default:
		return fmt.Errorf("unknown agent type: %s", inst.Type)
	}

	return nil
}

func (m *InstanceManager) cleanupInstance(ctx context.Context, inst *AgentInstance) error {
	// Type-specific cleanup
	switch inst.Type {
	case TypeAider:
		// Cleanup Aider resources

	case TypeClaudeCode:
		// Cleanup terminal resources

	case TypeCline:
		// Cleanup browser resources

	case TypeOpenHands:
		// Stop sandbox containers

	default:
		// Generic cleanup
	}

	return nil
}

func (m *InstanceManager) persistInstance(ctx context.Context, inst *AgentInstance) error {
	configJSON, err := json.Marshal(inst.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	providerJSON, err := json.Marshal(inst.Provider)
	if err != nil {
		return fmt.Errorf("marshal provider: %w", err)
	}

	stateJSON, err := json.Marshal(inst.State)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	_, err = m.db.ExecContext(ctx,
		`INSERT INTO agent_instances (
			id, agent_type, instance_name, status, config, provider_config,
			max_memory_mb, max_cpu_percent, current_session_id, current_task_id,
			health_status, requests_processed, errors_count, total_execution_time_ms,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		 ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			config = EXCLUDED.config,
			health_status = EXCLUDED.health_status,
			requests_processed = EXCLUDED.requests_processed,
			errors_count = EXCLUDED.errors_count,
			total_execution_time_ms = EXCLUDED.total_execution_time_ms,
			updated_at = EXCLUDED.updated_at`,
		inst.ID, inst.Type, inst.Name, inst.Status, configJSON, providerJSON,
		inst.Config.MaxMemoryMB, inst.Config.MaxCPUPercent,
		sql.NullString{String: inst.SessionID, Valid: inst.SessionID != ""},
		sql.NullString{String: inst.TaskID, Valid: inst.TaskID != ""},
		inst.Health, inst.RequestsProcessed, inst.ErrorsCount, inst.TotalExecTimeMs,
		inst.CreatedAt, inst.UpdatedAt,
	)

	return err
}

func (m *InstanceManager) recoverInstances(ctx context.Context) error {
	// Query instances that should be active
	rows, err := m.db.QueryContext(ctx,
		`SELECT id, agent_type, instance_name, status, config, provider_config,
		        max_memory_mb, max_cpu_percent, current_session_id, current_task_id,
		        health_status, requests_processed, errors_count, total_execution_time_ms,
		        created_at, updated_at
		 FROM agent_instances
		 WHERE status IN ('idle', 'active', 'background')`,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var inst AgentInstance
		var configJSON, providerJSON []byte
		var sessionID, taskID sql.NullString

		err := rows.Scan(
			&inst.ID, &inst.Type, &inst.Name, &inst.Status, &configJSON, &providerJSON,
			&inst.Config.MaxMemoryMB, &inst.Config.MaxCPUPercent,
			&sessionID, &taskID,
			&inst.Health, &inst.RequestsProcessed, &inst.ErrorsCount, &inst.TotalExecTimeMs,
			&inst.CreatedAt, &inst.UpdatedAt,
		)
		if err != nil {
			m.logger.Printf("Error scanning instance: %v", err)
			continue
		}

		if sessionID.Valid {
			inst.SessionID = sessionID.String
		}
		if taskID.Valid {
			inst.TaskID = taskID.String
		}

		// Parse config
		if err := json.Unmarshal(configJSON, &inst.Config); err != nil {
			m.logger.Printf("Error parsing config for %s: %v", inst.ID, err)
		}
		if err := json.Unmarshal(providerJSON, &inst.Provider); err != nil {
			m.logger.Printf("Error parsing provider for %s: %v", inst.ID, err)
		}

		// Initialize channels
		inst.RequestCh = make(chan *Request, 10)
		inst.ResponseCh = make(chan *Response, 10)
		inst.EventCh = make(chan *Event, 10)

		// Register in memory
		m.mu.Lock()
		m.instances[inst.ID] = &inst
		m.mu.Unlock()

		// Restart event loops
		go m.instanceEventLoop(&inst)
		go m.instanceHealthLoop(&inst)

		m.logger.Printf("Recovered instance %s of type %s", inst.ID, inst.Type)
	}

	return rows.Err()
}

func (m *InstanceManager) instanceEventLoop(inst *AgentInstance) {
	m.logger.Printf("Started event loop for instance %s", inst.ID)

	for {
		select {
		case req, ok := <-inst.RequestCh:
			if !ok {
				return // Channel closed
			}
			resp := m.handleRequest(inst, req)
			inst.ResponseCh <- resp

		case event, ok := <-inst.EventCh:
			if !ok {
				return
			}
			m.eventBus.Publish(event)

		case <-m.healthCheckStop:
			return
		}
	}
}

func (m *InstanceManager) instanceHealthLoop(inst *AgentInstance) {
	ticker := time.NewTicker(inst.Config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if inst.Status == StatusTerminating || inst.Status == StatusTerminated {
				return
			}

			result := m.performHealthCheck(inst)
			inst.Health = result.Status
			inst.HealthDetails = result.Details
			now := time.Now()
			inst.LastHealthCheck = &now

			// Update database
			_, err := m.db.Exec(
				"UPDATE agent_instances SET health_status = $1, last_health_check = NOW() WHERE id = $2",
				inst.Health, inst.ID,
			)
			if err != nil {
				m.logger.Printf("Error updating health check for %s: %v", inst.ID, err)
			}

		case <-m.healthCheckStop:
			return
		}
	}
}

func (m *InstanceManager) handleRequest(inst *AgentInstance, req *Request) *Response {
	start := time.Now()

	// Route to type-specific handler
	var result interface{}
	var err error

	switch req.Type {
	case RequestTypeExecute:
		result, err = m.handleExecute(inst, req.Payload)
	case RequestTypeQuery:
		result, err = m.handleQuery(inst, req.Payload)
	case RequestTypeHealth:
		result = m.performHealthCheck(inst)
	case RequestTypeCancel:
		// Handle cancellation
		result = map[string]bool{"cancelled": true}
	default:
		err = fmt.Errorf("unknown request type: %s", req.Type)
	}

	duration := time.Since(start)

	if err != nil {
		return &Response{
			RequestID: req.ID,
			Success:   false,
			Error: &ErrorDetail{
				Code:    "REQUEST_ERROR",
				Message: err.Error(),
			},
			Duration: duration,
		}
	}

	return &Response{
		RequestID: req.ID,
		Success:   true,
		Result:    result,
		Duration:  duration,
	}
}

func (m *InstanceManager) handleExecute(inst *AgentInstance, payload interface{}) (interface{}, error) {
	// Type-specific execution
	switch inst.Type {
	case TypeAider:
		// Execute Aider command
		return m.executeAider(inst, payload)
	case TypeClaudeCode:
		return m.executeClaudeCode(inst, payload)
	case TypeCodex:
		return m.executeCodex(inst, payload)
	default:
		return nil, fmt.Errorf("execution not implemented for type: %s", inst.Type)
	}
}

func (m *InstanceManager) handleQuery(inst *AgentInstance, payload interface{}) (interface{}, error) {
	// Type-specific query
	return map[string]interface{}{
		"status":  inst.Status,
		"health":  inst.Health,
		"metrics": m.GetMetrics(),
	}, nil
}

func (m *InstanceManager) performHealthCheck(inst *AgentInstance) *HealthCheckResult {
	result := &HealthCheckResult{
		CheckedAt: time.Now(),
	}

	// Basic health checks
	if inst.Status == StatusFailed || inst.Status == StatusTerminating {
		result.Healthy = false
		result.Status = HealthUnhealthy
		result.Message = "Instance in failed/terminating state"
		return result
	}

	// Check error rate
	if inst.RequestsProcessed > 0 {
		errorRate := float64(inst.ErrorsCount) / float64(inst.RequestsProcessed)
		if errorRate > 0.5 {
			result.Status = HealthDegraded
			result.Message = "High error rate detected"
			result.Details = map[string]interface{}{
				"error_rate": errorRate,
			}
			return result
		}
	}

	result.Healthy = true
	result.Status = HealthHealthy
	result.Message = "Instance is healthy"
	return result
}

func (m *InstanceManager) healthCheckLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Clean up expired locks
			_, err := m.db.Exec("DELETE FROM distributed_locks WHERE expires_at < NOW()")
			if err != nil {
				m.logger.Printf("Error cleaning locks: %v", err)
			}

		case <-m.healthCheckStop:
			return
		}
	}
}

// Type-specific execution methods (stubs for now)
func (m *InstanceManager) executeAider(inst *AgentInstance, payload interface{}) (interface{}, error) {
	return map[string]string{"status": "executed", "type": "aider"}, nil
}

func (m *InstanceManager) executeClaudeCode(inst *AgentInstance, payload interface{}) (interface{}, error) {
	return map[string]string{"status": "executed", "type": "claude_code"}, nil
}

func (m *InstanceManager) executeCodex(inst *AgentInstance, payload interface{}) (interface{}, error) {
	return map[string]string{"status": "executed", "type": "codex"}, nil
}

// Helper functions

func generateShortID() string {
	return uuid.New().String()[:8]
}

// WorkerPool is a simple worker pool for background tasks.
type WorkerPool struct {
	size int
	sem  chan struct{}
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(size int) *WorkerPool {
	return &WorkerPool{
		size: size,
		sem:  make(chan struct{}, size),
	}
}

// Submit submits a task to the pool.
func (p *WorkerPool) Submit(ctx context.Context, fn func()) error {
	select {
	case p.sem <- struct{}{}:
		go func() {
			defer func() { <-p.sem }()
			fn()
		}()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
