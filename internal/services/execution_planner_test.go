package services

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

func newTestPlanner() *ExecutionPlanner {
	return NewExecutionPlanner(logrus.New())
}

func TestExecutionPlanner_ParseTasks_ValidJSON(t *testing.T) {
	p := newTestPlanner()

	input := `[
		{
			"id": "task-1",
			"description": "Read source files",
			"dependencies": [],
			"tool_requirements": ["lsp"],
			"priority": 8,
			"estimated_steps": 3
		},
		{
			"id": "task-2",
			"description": "Refactor module",
			"dependencies": ["task-1"],
			"tool_requirements": ["mcp", "lsp"],
			"priority": 6,
			"estimated_steps": 10
		}
	]`

	tasks, err := p.parseTasks(input)
	require.NoError(t, err)
	require.Len(t, tasks, 2)

	assert.Equal(t, "task-1", tasks[0].ID)
	assert.Equal(t, "Read source files", tasks[0].Description)
	assert.Empty(t, tasks[0].Dependencies)
	assert.Equal(t, []string{"lsp"}, tasks[0].ToolRequirements)
	assert.Equal(t, 8, tasks[0].Priority)
	assert.Equal(t, 3, tasks[0].EstimatedSteps)
	assert.Equal(t, AgenticTaskPending, tasks[0].Status)

	assert.Equal(t, "task-2", tasks[1].ID)
	assert.Equal(t, []string{"task-1"}, tasks[1].Dependencies)
}

func TestExecutionPlanner_ParseTasks_MarkdownWrapped(t *testing.T) {
	p := newTestPlanner()

	input := "```json\n" + `[
		{
			"id": "a",
			"description": "Do something",
			"dependencies": [],
			"tool_requirements": ["rag"],
			"priority": 5,
			"estimated_steps": 2
		}
	]` + "\n```"

	tasks, err := p.parseTasks(input)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "a", tasks[0].ID)
	assert.Equal(t, "Do something", tasks[0].Description)
}

func TestExecutionPlanner_ParseTasks_EmptyID(t *testing.T) {
	p := newTestPlanner()

	input := `[{"id":"","description":"auto-id task","dependencies":[],"tool_requirements":[],"priority":1,"estimated_steps":1}]`

	tasks, err := p.parseTasks(input)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.NotEmpty(t, tasks[0].ID, "empty ID should be replaced with generated UUID")
	assert.Contains(t, tasks[0].ID, "task-")
}

func TestExecutionPlanner_ParseTasks_InvalidJSON(t *testing.T) {
	p := newTestPlanner()

	_, err := p.parseTasks("this is not json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse task JSON")
}

func TestExecutionPlanner_BuildDependencyGraph_NoDeps(t *testing.T) {
	p := newTestPlanner()

	tasks := []AgenticTask{
		{ID: "a", Description: "Task A"},
		{ID: "b", Description: "Task B"},
		{ID: "c", Description: "Task C"},
	}

	layers, err := p.BuildDependencyGraph(tasks)
	require.NoError(t, err)
	require.Len(t, layers, 1, "all independent tasks should be in one layer")
	assert.Len(t, layers[0], 3)
}

func TestExecutionPlanner_BuildDependencyGraph_Chain(t *testing.T) {
	p := newTestPlanner()

	// A -> B -> C (linear chain, 3 layers)
	tasks := []AgenticTask{
		{ID: "a", Description: "Task A"},
		{ID: "b", Description: "Task B", Dependencies: []string{"a"}},
		{ID: "c", Description: "Task C", Dependencies: []string{"b"}},
	}

	layers, err := p.BuildDependencyGraph(tasks)
	require.NoError(t, err)
	require.Len(t, layers, 3)

	assert.Len(t, layers[0], 1)
	assert.Equal(t, "a", layers[0][0].ID)

	assert.Len(t, layers[1], 1)
	assert.Equal(t, "b", layers[1][0].ID)

	assert.Len(t, layers[2], 1)
	assert.Equal(t, "c", layers[2][0].ID)
}

func TestExecutionPlanner_BuildDependencyGraph_Parallel(t *testing.T) {
	p := newTestPlanner()

	// A and B are independent; C depends on both
	tasks := []AgenticTask{
		{ID: "a", Description: "Task A"},
		{ID: "b", Description: "Task B"},
		{ID: "c", Description: "Task C", Dependencies: []string{"a", "b"}},
	}

	layers, err := p.BuildDependencyGraph(tasks)
	require.NoError(t, err)
	require.Len(t, layers, 2)

	// Layer 0 should have A and B (order may vary).
	layer0IDs := []string{layers[0][0].ID, layers[0][1].ID}
	sort.Strings(layer0IDs)
	assert.Equal(t, []string{"a", "b"}, layer0IDs)

	// Layer 1 should have only C.
	assert.Len(t, layers[1], 1)
	assert.Equal(t, "c", layers[1][0].ID)
}

func TestExecutionPlanner_BuildDependencyGraph_CircularDep(t *testing.T) {
	p := newTestPlanner()

	tasks := []AgenticTask{
		{ID: "a", Description: "Task A", Dependencies: []string{"b"}},
		{ID: "b", Description: "Task B", Dependencies: []string{"a"}},
	}

	_, err := p.BuildDependencyGraph(tasks)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency detected")
}

func TestExecutionPlanner_BuildDependencyGraph_UnknownDep(t *testing.T) {
	p := newTestPlanner()

	tasks := []AgenticTask{
		{ID: "a", Description: "Task A", Dependencies: []string{"nonexistent"}},
	}

	_, err := p.BuildDependencyGraph(tasks)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown dependency nonexistent")
}

func TestExecutionPlanner_DecomposePlan(t *testing.T) {
	p := newTestPlanner()

	mockComplete := func(
		_ context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		return &models.LLMResponse{
			Content: `[
				{
					"id": "task-1",
					"description": "Analyze codebase",
					"dependencies": [],
					"tool_requirements": ["lsp"],
					"priority": 9,
					"estimated_steps": 5
				}
			]`,
		}, nil
	}

	tasks, err := p.DecomposePlan(context.Background(), "Refactor the auth module", mockComplete)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "task-1", tasks[0].ID)
	assert.Equal(t, "Analyze codebase", tasks[0].Description)
	assert.Equal(t, AgenticTaskPending, tasks[0].Status)
}

func TestExecutionPlanner_DecomposePlan_LLMError(t *testing.T) {
	p := newTestPlanner()

	mockComplete := func(
		_ context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		return nil, fmt.Errorf("provider unavailable")
	}

	_, err := p.DecomposePlan(context.Background(), "Do something", mockComplete)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "LLM decomposition failed")
}

func TestExtractJSON_PlainJSON(t *testing.T) {
	input := `[{"id":"x"}]`
	assert.Equal(t, `[{"id":"x"}]`, extractJSON(input))
}

func TestExtractJSON_MarkdownCodeBlock(t *testing.T) {
	input := "```json\n[{\"id\":\"x\"}]\n```"
	assert.Equal(t, `[{"id":"x"}]`, extractJSON(input))
}

func TestExtractJSON_GenericCodeBlock(t *testing.T) {
	input := "```\n[{\"id\":\"x\"}]\n```"
	assert.Equal(t, `[{"id":"x"}]`, extractJSON(input))
}
