package services

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgenticMode_String(t *testing.T) {
	tests := []struct {
		name     string
		mode     AgenticMode
		expected string
	}{
		{"reason mode", AgenticModeReason, "reason"},
		{"execute mode", AgenticModeExecute, "execute"},
		{"unknown mode", AgenticMode(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mode.String())
		})
	}
}

func TestAgenticMode_Values(t *testing.T) {
	assert.Equal(t, AgenticMode(0), AgenticModeReason)
	assert.Equal(t, AgenticMode(1), AgenticModeExecute)
}

func TestAgenticTaskStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   AgenticTaskStatus
		expected string
	}{
		{"pending", AgenticTaskPending, "pending"},
		{"running", AgenticTaskRunning, "running"},
		{"completed", AgenticTaskCompleted, "completed"},
		{"failed", AgenticTaskFailed, "failed"},
		{"unknown", AgenticTaskStatus(42), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestAgenticTaskStatus_Values(t *testing.T) {
	assert.Equal(t, AgenticTaskStatus(0), AgenticTaskPending)
	assert.Equal(t, AgenticTaskStatus(1), AgenticTaskRunning)
	assert.Equal(t, AgenticTaskStatus(2), AgenticTaskCompleted)
	assert.Equal(t, AgenticTaskStatus(3), AgenticTaskFailed)
}

func TestAgenticTask_StatusTransitions(t *testing.T) {
	task := AgenticTask{
		ID:               "task-001",
		Description:      "Analyze codebase",
		Dependencies:     []string{},
		ToolRequirements: []string{"mcp", "lsp"},
		Priority:         1,
		EstimatedSteps:   3,
		Status:           AgenticTaskPending,
	}

	assert.Equal(t, AgenticTaskPending, task.Status)
	assert.Equal(t, "pending", task.Status.String())

	task.Status = AgenticTaskRunning
	assert.Equal(t, AgenticTaskRunning, task.Status)
	assert.Equal(t, "running", task.Status.String())

	task.Status = AgenticTaskCompleted
	assert.Equal(t, AgenticTaskCompleted, task.Status)
	assert.Equal(t, "completed", task.Status.String())
}

func TestAgenticTask_StatusTransition_ToFailed(t *testing.T) {
	task := AgenticTask{
		ID:          "task-002",
		Description: "Execute tool call",
		Status:      AgenticTaskRunning,
	}

	task.Status = AgenticTaskFailed
	assert.Equal(t, AgenticTaskFailed, task.Status)
	assert.Equal(t, "failed", task.Status.String())
}

func TestAgenticTask_Fields(t *testing.T) {
	task := AgenticTask{
		ID:               "task-abc",
		Description:      "Run semantic search",
		Dependencies:     []string{"task-001", "task-002"},
		ToolRequirements: []string{"embeddings", "vectordb"},
		Priority:         5,
		EstimatedSteps:   10,
		Status:           AgenticTaskPending,
	}

	assert.Equal(t, "task-abc", task.ID)
	assert.Equal(t, "Run semantic search", task.Description)
	assert.Len(t, task.Dependencies, 2)
	assert.Contains(t, task.Dependencies, "task-001")
	assert.Contains(t, task.Dependencies, "task-002")
	assert.Len(t, task.ToolRequirements, 2)
	assert.Equal(t, 5, task.Priority)
	assert.Equal(t, 10, task.EstimatedSteps)
}

func TestAgenticResult_WithoutError(t *testing.T) {
	result := AgenticResult{
		TaskID:   "task-001",
		AgentID:  "agent-alpha",
		Content:  "Analysis complete",
		Duration: 2 * time.Second,
		Error:    nil,
	}

	assert.Equal(t, "task-001", result.TaskID)
	assert.Equal(t, "agent-alpha", result.AgentID)
	assert.Equal(t, "Analysis complete", result.Content)
	assert.Equal(t, 2*time.Second, result.Duration)
	assert.NoError(t, result.Error)
	assert.Empty(t, result.ToolCalls)
}

func TestAgenticResult_WithError(t *testing.T) {
	err := errors.New("provider timeout")
	result := AgenticResult{
		TaskID:   "task-002",
		AgentID:  "agent-beta",
		Content:  "",
		Duration: 5 * time.Minute,
		Error:    err,
	}

	assert.Equal(t, "task-002", result.TaskID)
	assert.Error(t, result.Error)
	assert.Equal(t, "provider timeout", result.Error.Error())
}

func TestAgenticResult_WithToolCalls(t *testing.T) {
	result := AgenticResult{
		TaskID:  "task-003",
		AgentID: "agent-gamma",
		Content: "Tool execution done",
		ToolCalls: []AgenticToolExecution{
			{
				Protocol:  "mcp",
				Operation: "read_file",
				Input:     map[string]string{"path": "/tmp/test.go"},
				Output:    "file contents",
				Duration:  100 * time.Millisecond,
			},
			{
				Protocol:  "lsp",
				Operation: "diagnostics",
				Input:     nil,
				Output:    []string{"no errors"},
				Duration:  50 * time.Millisecond,
				Error:     nil,
			},
		},
		Duration: 500 * time.Millisecond,
	}

	assert.Len(t, result.ToolCalls, 2)
	assert.Equal(t, "mcp", result.ToolCalls[0].Protocol)
	assert.Equal(t, "read_file", result.ToolCalls[0].Operation)
	assert.Equal(t, "lsp", result.ToolCalls[1].Protocol)
	assert.Equal(t, "diagnostics", result.ToolCalls[1].Operation)
}

func TestAgenticToolExecution_WithError(t *testing.T) {
	exec := AgenticToolExecution{
		Protocol:  "mcp",
		Operation: "write_file",
		Input:     map[string]string{"path": "/tmp/out.txt"},
		Duration:  200 * time.Millisecond,
		Error:     errors.New("permission denied"),
	}

	assert.Error(t, exec.Error)
	assert.Equal(t, "permission denied", exec.Error.Error())
	assert.Nil(t, exec.Output)
}

func TestDefaultAgenticEnsembleConfig(t *testing.T) {
	cfg := DefaultAgenticEnsembleConfig()

	assert.Equal(t, 5, cfg.MaxConcurrentAgents)
	assert.Equal(t, 20, cfg.MaxIterationsPerAgent)
	assert.Equal(t, 5, cfg.MaxToolIterationsPerPhase)
	assert.Equal(t, 5*time.Minute, cfg.AgentTimeout)
	assert.Equal(t, 15*time.Minute, cfg.GlobalTimeout)
	assert.Equal(t, 30*time.Second, cfg.ToolIterationTimeout)
	assert.True(t, cfg.EnableVision)
	assert.True(t, cfg.EnableMemory)
	assert.True(t, cfg.EnableExecution)
}

func TestAgenticEnsembleConfig_CustomValues(t *testing.T) {
	cfg := AgenticEnsembleConfig{
		MaxConcurrentAgents:       10,
		MaxIterationsPerAgent:     50,
		MaxToolIterationsPerPhase: 10,
		AgentTimeout:              10 * time.Minute,
		GlobalTimeout:             30 * time.Minute,
		ToolIterationTimeout:      1 * time.Minute,
		EnableVision:              false,
		EnableMemory:              false,
		EnableExecution:           true,
	}

	assert.Equal(t, 10, cfg.MaxConcurrentAgents)
	assert.Equal(t, 50, cfg.MaxIterationsPerAgent)
	assert.Equal(t, 10, cfg.MaxToolIterationsPerPhase)
	assert.Equal(t, 10*time.Minute, cfg.AgentTimeout)
	assert.Equal(t, 30*time.Minute, cfg.GlobalTimeout)
	assert.Equal(t, 1*time.Minute, cfg.ToolIterationTimeout)
	assert.False(t, cfg.EnableVision)
	assert.False(t, cfg.EnableMemory)
	assert.True(t, cfg.EnableExecution)
}

func TestAgenticMetadata_Fields(t *testing.T) {
	meta := AgenticMetadata{
		Mode:            "execute",
		StagesCompleted: []string{"decompose", "assign", "execute", "aggregate"},
		AgentsSpawned:   3,
		TasksCompleted:  5,
		ToolsInvoked: []ToolInvocationSummary{
			{Protocol: "mcp", Count: 4},
			{Protocol: "lsp", Count: 2},
		},
		TotalDurationMs: 12500,
		ProvenanceID:    "prov-abc-123",
	}

	assert.Equal(t, "execute", meta.Mode)
	assert.Len(t, meta.StagesCompleted, 4)
	assert.Equal(t, 3, meta.AgentsSpawned)
	assert.Equal(t, 5, meta.TasksCompleted)
	require.Len(t, meta.ToolsInvoked, 2)
	assert.Equal(t, "mcp", meta.ToolsInvoked[0].Protocol)
	assert.Equal(t, 4, meta.ToolsInvoked[0].Count)
	assert.Equal(t, "lsp", meta.ToolsInvoked[1].Protocol)
	assert.Equal(t, 2, meta.ToolsInvoked[1].Count)
	assert.Equal(t, int64(12500), meta.TotalDurationMs)
	assert.Equal(t, "prov-abc-123", meta.ProvenanceID)
}

func TestToolInvocationSummary_Fields(t *testing.T) {
	summary := ToolInvocationSummary{
		Protocol: "embeddings",
		Count:    7,
	}

	assert.Equal(t, "embeddings", summary.Protocol)
	assert.Equal(t, 7, summary.Count)
}
