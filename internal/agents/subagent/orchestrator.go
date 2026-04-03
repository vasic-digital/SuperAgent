// Package subagent orchestrator coordinates main workflow with sub-agents
// Inspired by Snow CLI's sub-agent workflow
package subagent

import (
	"context"
	"fmt"
	"strings"

	"dev.helix.agent/internal/agents"
	"go.uber.org/zap"
)

// Orchestrator coordinates sub-agents within the main workflow
type Orchestrator struct {
	logger  *zap.Logger
	manager SubAgentManager
}

// NewOrchestrator creates a new sub-agent orchestrator
func NewOrchestrator(logger *zap.Logger, manager SubAgentManager) *Orchestrator {
	return &Orchestrator{
		logger:  logger,
		manager: manager,
	}
}

// SubAgentDecision represents whether to use a sub-agent
type SubAgentDecision struct {
	UseSubAgent bool           `json:"use_sub_agent"`
	AgentType   SubAgentType   `json:"agent_type,omitempty"`
	Reason      string         `json:"reason"`
}

// Decide determines if a task should use a sub-agent
// Based on Snow CLI's decision logic
func (o *Orchestrator) Decide(ctx context.Context, task string, context map[string]interface{}) (*SubAgentDecision, error) {
	// Simple heuristic-based decision
	// In production, this could use an LLM to decide
	
	// Check for exploration patterns
	if strings.Contains(strings.ToLower(task), "find") ||
		strings.Contains(strings.ToLower(task), "search") ||
		strings.Contains(strings.ToLower(task), "where") ||
		strings.Contains(strings.ToLower(task), "locate") {
		return &SubAgentDecision{
			UseSubAgent: true,
			AgentType:   ExploreAgent,
			Reason:      "Task involves code exploration and searching",
		}, nil
	}
	
	// Check for planning patterns
	if strings.Contains(strings.ToLower(task), "plan") ||
		strings.Contains(strings.ToLower(task), "design") ||
		strings.Contains(strings.ToLower(task), "architecture") ||
		strings.Contains(strings.ToLower(task), "refactor") {
		return &SubAgentDecision{
			UseSubAgent: true,
			AgentType:   PlanAgent,
			Reason:      "Task requires planning and design",
		}, nil
	}
	
	// Check for batch/multi-file patterns
	if strings.Contains(strings.ToLower(task), "all files") ||
		strings.Contains(strings.ToLower(task), "every file") ||
		strings.Contains(strings.ToLower(task), "batch") {
		return &SubAgentDecision{
			UseSubAgent: true,
			AgentType:   GeneralAgent,
			Reason:      "Task involves batch operations across multiple files",
		}, nil
	}
	
	return &SubAgentDecision{
		UseSubAgent: false,
		Reason:      "Task can be handled by main workflow",
	}, nil
}

// ExecuteWithSubAgent executes a task using the appropriate sub-agent
func (o *Orchestrator) ExecuteWithSubAgent(ctx context.Context, agentType SubAgentType, prompt string, context map[string]interface{}) (*TaskResult, error) {
	// Get the appropriate agent
	agents, err := o.manager.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	
	var agentID string
	for _, agent := range agents {
		if agent.Type == agentType {
			agentID = agent.ID
			break
		}
	}
	
	if agentID == "" {
		return nil, fmt.Errorf("no sub-agent found for type: %s", agentType)
	}
	
	o.logger.Info("Executing with sub-agent",
		zap.String("agent_type", string(agentType)),
		zap.String("agent_id", agentID),
	)
	
	// Create task
	task := SubAgentTask{
		Type:    agentType,
		Prompt:  o.buildPrompt(agentType, prompt, context),
		Context: context,
	}
	
	// Execute
	return o.manager.Execute(ctx, agentID, task)
}

// ExecuteAsyncWithSubAgent executes a task asynchronously
func (o *Orchestrator) ExecuteAsyncWithSubAgent(ctx context.Context, agentType SubAgentType, prompt string, context map[string]interface{}) (string, error) {
	agents, err := o.manager.List(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list agents: %w", err)
	}
	
	var agentID string
	for _, agent := range agents {
		if agent.Type == agentType {
			agentID = agent.ID
			break
		}
	}
	
	if agentID == "" {
		return "", fmt.Errorf("no sub-agent found for type: %s", agentType)
	}
	
	task := SubAgentTask{
		Type:    agentType,
		Prompt:  o.buildPrompt(agentType, prompt, context),
		Context: context,
	}
	
	return o.manager.ExecuteAsync(ctx, agentID, task)
}

// buildPrompt builds a specialized prompt for each sub-agent type
func (o *Orchestrator) buildPrompt(agentType SubAgentType, originalPrompt string, context map[string]interface{}) string {
	var sb strings.Builder
	
	switch agentType {
	case ExploreAgent:
		sb.WriteString("EXPLORATION TASK\n")
		sb.WriteString("================\n\n")
		sb.WriteString("Your goal is to explore the codebase and find specific information.\n\n")
		sb.WriteString("Task: ")
		sb.WriteString(originalPrompt)
		sb.WriteString("\n\n")
		sb.WriteString("Instructions:\n")
		sb.WriteString("1. Search thoroughly through the codebase\n")
		sb.WriteString("2. Provide file paths and line numbers\n")
		sb.WriteString("3. Show relevant code snippets\n")
		sb.WriteString("4. Explain your findings\n")
		
	case PlanAgent:
		sb.WriteString("PLANNING TASK\n")
		sb.WriteString("=============\n\n")
		sb.WriteString("Your goal is to create a detailed implementation plan.\n\n")
		sb.WriteString("Task: ")
		sb.WriteString(originalPrompt)
		sb.WriteString("\n\n")
		sb.WriteString("Instructions:\n")
		sb.WriteString("1. Break down the task into steps\n")
		sb.WriteString("2. Identify dependencies\n")
		sb.WriteString("3. Provide a clear implementation roadmap\n")
		sb.WriteString("4. Consider edge cases\n")
		
	case GeneralAgent:
		sb.WriteString("GENERAL CODING TASK\n")
		sb.WriteString("===================\n\n")
		sb.WriteString("Your goal is to complete the coding task efficiently.\n\n")
		sb.WriteString("Task: ")
		sb.WriteString(originalPrompt)
		sb.WriteString("\n\n")
		sb.WriteString("Instructions:\n")
		sb.WriteString("1. Work efficiently across multiple files if needed\n")
		sb.WriteString("2. Maintain code quality\n")
		sb.WriteString("3. Follow existing patterns\n")
		sb.WriteString("4. Test your changes\n")
	}
	
	// Add context if available
	if len(context) > 0 {
		sb.WriteString("\n\nContext:\n")
		for key, value := range context {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
		}
	}
	
	return sb.String()
}

// ProcessTaskResult processes a sub-agent result and integrates it into main workflow
func (o *Orchestrator) ProcessTaskResult(result *TaskResult) (string, error) {
	if result.Error != "" {
		return "", fmt.Errorf("sub-agent failed: %s", result.Error)
	}
	
	// Format the result for main workflow consumption
	var sb strings.Builder
	
	sb.WriteString("## Sub-Agent Results\n\n")
	sb.WriteString(result.Content)
	
	if len(result.ToolCalls) > 0 {
		sb.WriteString("\n\n### Actions Taken\n")
		for _, tc := range result.ToolCalls {
			sb.WriteString(fmt.Sprintf("- **%s**: `%s`\n", tc.Name, tc.Arguments))
		}
	}
	
	if result.Usage != nil {
		sb.WriteString(fmt.Sprintf("\n\n*Tokens used: %d input, %d output*",
			result.Usage.InputTokens,
			result.Usage.OutputTokens,
		))
	}
	
	return sb.String(), nil
}

// Integration with HelixAgent's main agent system
func (o *Orchestrator) ToAgentResult(result *TaskResult) *agents.AgentResult {
	return &agents.AgentResult{
		Content: result.Content,
		ToolCalls: result.ToolCalls,
		Usage: result.Usage,
		Metadata: map[string]interface{}{
			"source": "sub-agent",
		},
	}
}
