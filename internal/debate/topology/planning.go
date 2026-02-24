// Package topology provides planning style implementations for debate coordination.
// Implements CPDE (Centralized Planning, Decentralized Execution) and DPDE
// (Decentralized Planning, Decentralized Execution) planning styles from the
// MARBLE framework research for multi-agent debate orchestration.
package topology

import (
	"fmt"
	"sort"
	"time"
)

// PlanningStyle identifies the coordination planning approach.
type PlanningStyle string

const (
	// PlanningCPDE - Centralized Planning, Decentralized Execution.
	// A lead agent creates a comprehensive plan, distributes tasks to others.
	PlanningCPDE PlanningStyle = "cpde"
	// PlanningDPDE - Decentralized Planning, Decentralized Execution.
	// Each agent plans its own subtask, coordinates through negotiation.
	PlanningDPDE PlanningStyle = "dpde"
	// PlanningAdaptive - Auto-selects between CPDE and DPDE based on task
	// characteristics.
	PlanningAdaptive PlanningStyle = "adaptive"
)

// PlannedTask represents a single task in a debate plan.
type PlannedTask struct {
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	AssignedTo   string   `json:"assigned_to"`
	Dependencies []string `json:"dependencies,omitempty"`
	Priority     int      `json:"priority"`
	Status       string   `json:"status"` // pending, in_progress, completed
}

// TaskPlan represents a complete debate execution plan.
type TaskPlan struct {
	PlannerAgentID string         `json:"planner_agent_id"`
	Style          PlanningStyle  `json:"style"`
	Tasks          []*PlannedTask `json:"tasks"`
	Ordering       []string       `json:"ordering"`
	CreatedAt      time.Time      `json:"created_at"`
}

// PlanningStyleSelector selects the appropriate planning style based on task
// characteristics.
type PlanningStyleSelector struct {
	defaultStyle PlanningStyle
}

// NewPlanningStyleSelector creates a new PlanningStyleSelector with the given
// default style.
func NewPlanningStyleSelector(
	defaultStyle PlanningStyle,
) *PlanningStyleSelector {
	return &PlanningStyleSelector{
		defaultStyle: defaultStyle,
	}
}

// SelectStyle selects the appropriate planning style based on task complexity,
// ambiguity, and agent count.
//
// Selection logic:
//   - High complexity (>0.7) with low ambiguity (<0.3): CPDE (well-defined,
//     needs coordination)
//   - High ambiguity (>0.6): DPDE (exploratory, let agents figure it out)
//   - Small team (<=3): CPDE (central planning efficient)
//   - Large team (>8): DPDE (decentralized scales better)
//   - Otherwise: default style
func (s *PlanningStyleSelector) SelectStyle(
	taskComplexity, ambiguity float64,
	agentCount int,
) PlanningStyle {
	if taskComplexity > 0.7 && ambiguity < 0.3 {
		return PlanningCPDE
	}
	if ambiguity > 0.6 {
		return PlanningDPDE
	}
	if agentCount <= 3 {
		return PlanningCPDE
	}
	if agentCount > 8 {
		return PlanningDPDE
	}
	return s.defaultStyle
}

// NewTaskPlan creates a new empty TaskPlan with the given planner agent ID
// and planning style.
func NewTaskPlan(plannerID string, style PlanningStyle) *TaskPlan {
	return &TaskPlan{
		PlannerAgentID: plannerID,
		Style:          style,
		Tasks:          make([]*PlannedTask, 0),
		Ordering:       make([]string, 0),
		CreatedAt:      time.Now(),
	}
}

// AddTask adds a task to the plan and appends its ID to the ordering.
func (p *TaskPlan) AddTask(
	id, description, assignedTo string,
	priority int,
	dependencies ...string,
) {
	task := &PlannedTask{
		ID:           id,
		Description:  description,
		AssignedTo:   assignedTo,
		Dependencies: dependencies,
		Priority:     priority,
		Status:       "pending",
	}
	p.Tasks = append(p.Tasks, task)
	p.Ordering = append(p.Ordering, id)
}

// GetExecutionOrder returns tasks sorted by dependency order using
// topological sort. Tasks with no dependencies come first, followed by
// tasks whose dependencies are all satisfied. Within the same dependency
// depth, tasks are sorted by priority (higher priority first).
func (p *TaskPlan) GetExecutionOrder() []*PlannedTask {
	if len(p.Tasks) == 0 {
		return nil
	}

	// Build lookup maps
	taskByID := make(map[string]*PlannedTask, len(p.Tasks))
	for _, t := range p.Tasks {
		taskByID[t.ID] = t
	}

	// Kahn's algorithm for topological sort
	inDegree := make(map[string]int, len(p.Tasks))
	dependents := make(map[string][]string, len(p.Tasks))
	for _, t := range p.Tasks {
		if _, exists := inDegree[t.ID]; !exists {
			inDegree[t.ID] = 0
		}
		for _, dep := range t.Dependencies {
			// Only count dependencies that exist in the plan
			if _, exists := taskByID[dep]; exists {
				inDegree[t.ID]++
				dependents[dep] = append(dependents[dep], t.ID)
			}
		}
	}

	// Collect tasks with no incoming edges
	var queue []*PlannedTask
	for _, t := range p.Tasks {
		if inDegree[t.ID] == 0 {
			queue = append(queue, t)
		}
	}

	// Sort initial queue by priority descending
	sort.Slice(queue, func(i, j int) bool {
		return queue[i].Priority > queue[j].Priority
	})

	result := make([]*PlannedTask, 0, len(p.Tasks))

	for len(queue) > 0 {
		// Take the first element (highest priority among ready tasks)
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Reduce in-degree for dependents
		for _, depID := range dependents[current.ID] {
			inDegree[depID]--
			if inDegree[depID] == 0 {
				if t, exists := taskByID[depID]; exists {
					queue = append(queue, t)
				}
			}
		}

		// Re-sort queue by priority after adding new elements
		sort.Slice(queue, func(i, j int) bool {
			return queue[i].Priority > queue[j].Priority
		})
	}

	return result
}

// GetTasksByAgent returns all tasks assigned to the specified agent.
func (p *TaskPlan) GetTasksByAgent(agentID string) []*PlannedTask {
	var tasks []*PlannedTask
	for _, t := range p.Tasks {
		if t.AssignedTo == agentID {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

// GetReadyTasks returns tasks whose dependencies are all completed and
// that are not yet completed themselves.
func (p *TaskPlan) GetReadyTasks() []*PlannedTask {
	// Build a set of completed task IDs
	completed := make(map[string]bool, len(p.Tasks))
	for _, t := range p.Tasks {
		if t.Status == "completed" {
			completed[t.ID] = true
		}
	}

	var ready []*PlannedTask
	for _, t := range p.Tasks {
		if t.Status == "completed" {
			continue
		}

		allDepsCompleted := true
		for _, dep := range t.Dependencies {
			if !completed[dep] {
				allDepsCompleted = false
				break
			}
		}

		if allDepsCompleted {
			ready = append(ready, t)
		}
	}
	return ready
}

// MarkCompleted marks a task as completed by its ID.
// Returns an error if the task is not found.
func (p *TaskPlan) MarkCompleted(taskID string) error {
	for _, t := range p.Tasks {
		if t.ID == taskID {
			t.Status = "completed"
			return nil
		}
	}
	return fmt.Errorf("task %s not found in plan", taskID)
}

// IsComplete returns true if all tasks in the plan are completed.
func (p *TaskPlan) IsComplete() bool {
	for _, t := range p.Tasks {
		if t.Status != "completed" {
			return false
		}
	}
	return true
}
