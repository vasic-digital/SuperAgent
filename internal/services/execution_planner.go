package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/models"
)

// ExecutionPlanner decomposes debate decisions into executable AgenticTasks.
type ExecutionPlanner struct {
	logger *logrus.Logger
}

// NewExecutionPlanner creates a new ExecutionPlanner instance.
func NewExecutionPlanner(logger *logrus.Logger) *ExecutionPlanner {
	return &ExecutionPlanner{logger: logger}
}

// DecomposePlan takes a debate decision string and uses an LLM to decompose
// it into structured tasks with dependencies.
func (p *ExecutionPlanner) DecomposePlan(
	ctx context.Context,
	decision string,
	completeFunc func(ctx context.Context, messages []models.Message) (*models.LLMResponse, error),
) ([]AgenticTask, error) {
	prompt := fmt.Sprintf(`Decompose the following plan into discrete executable tasks.
Return a JSON array where each task has:
- "id": unique short identifier (e.g., "task-1")
- "description": what needs to be done
- "dependencies": array of task IDs that must complete first (empty if independent)
- "tool_requirements": array of required tools ("mcp", "lsp", "rag", "vision", "embeddings", "acp")
- "priority": integer 1-10 (higher = more important)
- "estimated_steps": estimated number of LLM reasoning steps needed (1-20)

Plan to decompose:
%s

Return ONLY the JSON array, no other text.`, decision)

	messages := []models.Message{
		{Role: "system", Content: "You are a task decomposition expert. Return only valid JSON."},
		{Role: "user", Content: prompt},
	}

	resp, err := completeFunc(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM decomposition failed: %w", err)
	}

	return p.parseTasks(resp.Content)
}

// parseTasks extracts AgenticTask list from LLM JSON output.
func (p *ExecutionPlanner) parseTasks(content string) ([]AgenticTask, error) {
	jsonStr := extractJSON(content)

	var rawTasks []struct {
		ID               string   `json:"id"`
		Description      string   `json:"description"`
		Dependencies     []string `json:"dependencies"`
		ToolRequirements []string `json:"tool_requirements"`
		Priority         int      `json:"priority"`
		EstimatedSteps   int      `json:"estimated_steps"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &rawTasks); err != nil {
		return nil, fmt.Errorf("failed to parse task JSON: %w", err)
	}

	tasks := make([]AgenticTask, len(rawTasks))
	for i, rt := range rawTasks {
		id := rt.ID
		if id == "" {
			id = fmt.Sprintf("task-%s", uuid.New().String()[:8])
		}
		tasks[i] = AgenticTask{
			ID:               id,
			Description:      rt.Description,
			Dependencies:     rt.Dependencies,
			ToolRequirements: rt.ToolRequirements,
			Priority:         rt.Priority,
			EstimatedSteps:   rt.EstimatedSteps,
			Status:           AgenticTaskPending,
		}
	}
	return tasks, nil
}

// BuildDependencyGraph returns tasks grouped into layers for parallel
// execution. Layer 0 has no dependencies, layer 1 depends only on layer 0
// tasks, and so on. Returns an error if a circular dependency is detected
// or if a task references an unknown dependency.
func (p *ExecutionPlanner) BuildDependencyGraph(
	tasks []AgenticTask,
) ([][]AgenticTask, error) {
	taskMap := make(map[string]*AgenticTask)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}

	// Validate all dependencies reference existing tasks.
	for _, t := range tasks {
		for _, dep := range t.Dependencies {
			if _, exists := taskMap[dep]; !exists {
				return nil, fmt.Errorf(
					"task %s has unknown dependency %s", t.ID, dep,
				)
			}
		}
	}

	// Kahn's algorithm for topological layering.
	inDegree := make(map[string]int)
	for _, t := range tasks {
		inDegree[t.ID] = len(t.Dependencies)
	}

	remaining := make(map[string]bool)
	for _, t := range tasks {
		remaining[t.ID] = true
	}

	var layers [][]AgenticTask
	for len(remaining) > 0 {
		var layer []AgenticTask
		for id := range remaining {
			if inDegree[id] == 0 {
				layer = append(layer, *taskMap[id])
			}
		}
		if len(layer) == 0 {
			return nil, fmt.Errorf("circular dependency detected")
		}
		for _, t := range layer {
			delete(remaining, t.ID)
			// Decrease in-degree for tasks that depended on this one.
			for id := range remaining {
				for _, dep := range taskMap[id].Dependencies {
					if dep == t.ID {
						inDegree[id]--
					}
				}
			}
		}
		layers = append(layers, layer)
	}

	return layers, nil
}

// extractJSON pulls a JSON array from content that may be wrapped in
// markdown code block markers.
func extractJSON(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}
	return strings.TrimSpace(content)
}
