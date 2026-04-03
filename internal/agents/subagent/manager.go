package subagent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dev.helix.agent/internal/agents"
	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Manager implements SubAgentManager
type Manager struct {
	logger      *zap.Logger
	llmClient   llm.Client
	toolService services.ToolService
	
	agents      map[string]*SubAgent
	tasks       map[string]*SubAgentTask
	running     map[string]context.CancelFunc // taskID -> cancel function
	
	mu          sync.RWMutex
}

// NewManager creates a new sub-agent manager
func NewManager(logger *zap.Logger, llmClient llm.Client, toolService services.ToolService) *Manager {
	m := &Manager{
		logger:      logger,
		llmClient:   llmClient,
		toolService: toolService,
		agents:      make(map[string]*SubAgent),
		tasks:       make(map[string]*SubAgentTask),
		running:     make(map[string]context.CancelFunc),
	}
	
	// Initialize default sub-agents
	for _, agent := range DefaultSubAgents() {
		agent.CreatedAt = time.Now()
		agent.UpdatedAt = time.Now()
		m.agents[agent.ID] = agent
	}
	
	return m
}

// Create creates a new sub-agent
func (m *Manager) Create(ctx context.Context, config SubAgentConfig) (*SubAgent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	agent := &SubAgent{
		ID:        uuid.New().String(),
		Type:      CustomAgent,
		Config:    config,
		Status:    StatusIdle,
		Tools:     []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	m.agents[agent.ID] = agent
	m.logger.Info("Created sub-agent", zap.String("id", agent.ID), zap.String("type", string(agent.Type)))
	
	return agent, nil
}

// Get retrieves a sub-agent by ID
func (m *Manager) Get(ctx context.Context, id string) (*SubAgent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	agent, ok := m.agents[id]
	if !ok {
		return nil, fmt.Errorf("sub-agent not found: %s", id)
	}
	
	return agent, nil
}

// List returns all sub-agents
func (m *Manager) List(ctx context.Context) ([]*SubAgent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	agents := make([]*SubAgent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	
	return agents, nil
}

// Update updates a sub-agent configuration
func (m *Manager) Update(ctx context.Context, id string, config SubAgentConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	agent, ok := m.agents[id]
	if !ok {
		return fmt.Errorf("sub-agent not found: %s", id)
	}
	
	agent.Config = config
	agent.UpdatedAt = time.Now()
	
	m.logger.Info("Updated sub-agent", zap.String("id", id))
	return nil
}

// Delete removes a sub-agent
func (m *Manager) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, ok := m.agents[id]; !ok {
		return fmt.Errorf("sub-agent not found: %s", id)
	}
	
	// Don't allow deletion of built-in agents
	if id == "explore" || id == "plan" || id == "general" {
		return fmt.Errorf("cannot delete built-in sub-agent: %s", id)
	}
	
	delete(m.agents, id)
	m.logger.Info("Deleted sub-agent", zap.String("id", id))
	return nil
}

// Execute runs a task on a sub-agent
func (m *Manager) Execute(ctx context.Context, agentID string, task SubAgentTask) (*TaskResult, error) {
	agent, err := m.Get(ctx, agentID)
	if err != nil {
		return nil, err
	}
	
	// Create task
	task.ID = uuid.New().String()
	task.AgentID = agentID
	task.Status = TaskPending
	task.CreatedAt = time.Now()
	
	m.mu.Lock()
	m.tasks[task.ID] = &task
	m.mu.Unlock()
	
	// Execute synchronously
	return m.executeTask(ctx, agent, &task)
}

// ExecuteAsync runs a task asynchronously
func (m *Manager) ExecuteAsync(ctx context.Context, agentID string, task SubAgentTask) (string, error) {
	agent, err := m.Get(ctx, agentID)
	if err != nil {
		return "", err
	}
	
	// Create task
	task.ID = uuid.New().String()
	task.AgentID = agentID
	task.Status = TaskPending
	task.CreatedAt = time.Now()
	
	m.mu.Lock()
	m.tasks[task.ID] = &task
	m.mu.Unlock()
	
	// Execute asynchronously
	ctx, cancel := context.WithCancel(context.Background())
	
	m.mu.Lock()
	m.running[task.ID] = cancel
	m.mu.Unlock()
	
	go func() {
		defer func() {
			m.mu.Lock()
			delete(m.running, task.ID)
			m.mu.Unlock()
			cancel()
		}()
		
		result, err := m.executeTask(ctx, agent, &task)
		if err != nil {
			m.logger.Error("Async task failed", zap.String("task_id", task.ID), zap.Error(err))
		}
		_ = result
	}()
	
	return task.ID, nil
}

// executeTask executes a task and returns the result
func (m *Manager) executeTask(ctx context.Context, agent *SubAgent, task *SubAgentTask) (*TaskResult, error) {
	// Update status
	m.mu.Lock()
	task.Status = TaskRunning
	now := time.Now()
	task.StartedAt = &now
	agent.Status = StatusRunning
	m.mu.Unlock()
	
	// Build messages
	messages := []llm.Message{
		{Role: "system", Content: agent.Role},
		{Role: "user", Content: task.Prompt},
	}
	
	// Prepare tool definitions
	toolDefs := make([]llm.ToolDefinition, 0, len(agent.Tools))
	for _, toolName := range agent.Tools {
		tool, err := m.toolService.GetTool(toolName)
		if err != nil {
			m.logger.Warn("Tool not found", zap.String("tool", toolName))
			continue
		}
		toolDefs = append(toolDefs, llm.ToolDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		})
	}
	
	// Call LLM
	request := llm.ChatRequest{
		Model:       agent.Config.Model,
		Messages:    messages,
		MaxTokens:   agent.Config.MaxTokens,
		Temperature: agent.Config.Temperature,
		Tools:       toolDefs,
	}
	
	if agent.Config.EnableThinking {
		request.EnableThinking = true
		request.ThinkingBudget = agent.Config.ThinkingBudget
	}
	
	response, err := m.llmClient.Chat(ctx, request)
	if err != nil {
		m.mu.Lock()
		task.Status = TaskFailed
		agent.Status = StatusFailed
		m.mu.Unlock()
		
		return &TaskResult{Error: err.Error()}, err
	}
	
	// Process tool calls if any
	toolCalls := make([]agents.ToolCall, 0)
	for _, tc := range response.ToolCalls {
		toolCalls = append(toolCalls, agents.ToolCall{
			ID:        tc.ID,
			Name:      tc.Name,
			Arguments: tc.Arguments,
		})
	}
	
	// Build result
	result := &TaskResult{
		Content:   response.Content,
		ToolCalls: toolCalls,
		Usage: &agents.TokenUsage{
			InputTokens:  response.Usage.InputTokens,
			OutputTokens: response.Usage.OutputTokens,
			TotalTokens:  response.Usage.TotalTokens,
		},
	}
	
	// Update status
	m.mu.Lock()
	task.Status = TaskCompleted
	task.Result = result
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	agent.Status = StatusCompleted
	m.mu.Unlock()
	
	m.logger.Info("Task completed",
		zap.String("task_id", task.ID),
		zap.String("agent_id", agent.ID),
		zap.Int("input_tokens", result.Usage.InputTokens),
		zap.Int("output_tokens", result.Usage.OutputTokens),
	)
	
	return result, nil
}

// GetTask retrieves task status and result
func (m *Manager) GetTask(ctx context.Context, taskID string) (*SubAgentTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	task, ok := m.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	
	return task, nil
}

// CancelTask cancels a running task
func (m *Manager) CancelTask(ctx context.Context, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	
	if task.Status != TaskRunning {
		return fmt.Errorf("task is not running: %s", taskID)
	}
	
	// Cancel the context
	if cancel, ok := m.running[taskID]; ok {
		cancel()
		delete(m.running, taskID)
	}
	
	task.Status = TaskCancelled
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	
	m.logger.Info("Task cancelled", zap.String("task_id", taskID))
	return nil
}

// SendMessage sends a message to a running sub-agent
func (m *Manager) SendMessage(ctx context.Context, agentID string, message string) error {
	// This would be used for inter-agent communication
	// For now, just log it
	m.logger.Info("Message to sub-agent",
		zap.String("agent_id", agentID),
		zap.String("message", message),
	)
	return nil
}
