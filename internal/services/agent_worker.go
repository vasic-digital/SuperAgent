package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/models"
)

// AgentWorker is an individual agent that executes one AgenticTask through
// iterative reasoning and tool use. Each iteration the LLM either produces
// a tool call (resolved via the IterativeToolExecutor) or a final answer.
type AgentWorker struct {
	id            string
	task          AgenticTask
	toolExecutor  *IterativeToolExecutor
	maxIterations int
	logger        *logrus.Logger
}

// NewAgentWorker creates an AgentWorker bound to a specific task.
// maxIterations caps the outer reasoning loop (distinct from the per-iteration
// tool loop managed by IterativeToolExecutor). Defaults to 10 when <= 0.
func NewAgentWorker(
	id string,
	task AgenticTask,
	toolExecutor *IterativeToolExecutor,
	maxIterations int,
	logger *logrus.Logger,
) *AgentWorker {
	if maxIterations <= 0 {
		maxIterations = 10
	}
	if logger == nil {
		logger = logrus.New()
	}
	return &AgentWorker{
		id:            id,
		task:          task,
		toolExecutor:  toolExecutor,
		maxIterations: maxIterations,
		logger:        logger,
	}
}

// Execute runs the agent's reasoning loop. On each iteration the LLM is
// invoked via completeFunc. If the IterativeToolExecutor is available, tool
// calls are resolved automatically. The loop terminates when the LLM returns
// a final answer (no tool calls), the maximum iteration count is reached, or
// the context is cancelled.
func (w *AgentWorker) Execute(
	ctx context.Context,
	completeFunc CompleteFunc,
) AgenticResult {
	start := time.Now()

	w.task.Status = AgenticTaskRunning

	messages := []models.Message{
		{
			Role: "system",
			Content: fmt.Sprintf(
				"You are agent %s. Your task: %s. "+
					"Reason step-by-step. When you have a final answer, "+
					"provide it without requesting tool calls.",
				w.id, w.task.Description,
			),
		},
		{
			Role:    "user",
			Content: w.task.Description,
		},
	}

	var allToolExecs []AgenticToolExecution

	for iteration := 0; iteration < w.maxIterations; iteration++ {
		select {
		case <-ctx.Done():
			w.task.Status = AgenticTaskFailed
			return AgenticResult{
				TaskID:    w.task.ID,
				AgentID:   w.id,
				Content:   "",
				ToolCalls: allToolExecs,
				Duration:  time.Since(start),
				Error: fmt.Errorf(
					"context cancelled at iteration %d: %w",
					iteration, ctx.Err(),
				),
			}
		default:
		}

		w.logger.WithFields(logrus.Fields{
			"agent_id":  w.id,
			"task_id":   w.task.ID,
			"iteration": iteration + 1,
			"max":       w.maxIterations,
		}).Debug("Agent worker iteration starting")

		// When a tool executor is available, delegate to it so that tool
		// calls are automatically resolved across multiple sub-iterations.
		if w.toolExecutor != nil {
			resp, toolExecs, err := w.toolExecutor.ExecuteWithTools(
				ctx, completeFunc, messages,
			)
			allToolExecs = append(allToolExecs, toolExecs...)

			if err != nil {
				w.task.Status = AgenticTaskFailed
				return AgenticResult{
					TaskID:    w.task.ID,
					AgentID:   w.id,
					Content:   "",
					ToolCalls: allToolExecs,
					Duration:  time.Since(start),
					Error: fmt.Errorf(
						"tool executor failed at iteration %d: %w",
						iteration, err,
					),
				}
			}

			// If the executor returned a final response (no remaining
			// tool calls), the task is complete.
			if resp != nil && len(resp.ToolCalls) == 0 {
				w.task.Status = AgenticTaskCompleted
				return AgenticResult{
					TaskID:    w.task.ID,
					AgentID:   w.id,
					Content:   resp.Content,
					ToolCalls: allToolExecs,
					Duration:  time.Since(start),
					Error:     nil,
				}
			}

			// The executor hit its own max iterations but the LLM may
			// still be requesting tools. Append the response and loop.
			if resp != nil {
				messages = append(messages, models.Message{
					Role:    "assistant",
					Content: resp.Content,
				})
			}
			continue
		}

		// No tool executor — simple single-shot LLM call per iteration.
		resp, err := completeFunc(ctx, messages)
		if err != nil {
			w.task.Status = AgenticTaskFailed
			return AgenticResult{
				TaskID:    w.task.ID,
				AgentID:   w.id,
				Content:   "",
				ToolCalls: allToolExecs,
				Duration:  time.Since(start),
				Error: fmt.Errorf(
					"LLM call failed at iteration %d: %w", iteration, err,
				),
			}
		}

		if resp == nil {
			w.task.Status = AgenticTaskFailed
			return AgenticResult{
				TaskID:    w.task.ID,
				AgentID:   w.id,
				Content:   "",
				ToolCalls: allToolExecs,
				Duration:  time.Since(start),
				Error: fmt.Errorf(
					"nil response at iteration %d", iteration,
				),
			}
		}

		// No tool calls means the LLM produced a final answer.
		if len(resp.ToolCalls) == 0 {
			w.task.Status = AgenticTaskCompleted
			return AgenticResult{
				TaskID:    w.task.ID,
				AgentID:   w.id,
				Content:   resp.Content,
				ToolCalls: allToolExecs,
				Duration:  time.Since(start),
				Error:     nil,
			}
		}

		// The LLM requested tool calls but we have no executor — record
		// them as unanswered and continue so the model can try again.
		for _, tc := range resp.ToolCalls {
			allToolExecs = append(allToolExecs, AgenticToolExecution{
				Protocol:  "unknown",
				Operation: tc.Function.Name,
				Input:     tc.Function.Arguments,
				Output:    nil,
				Duration:  0,
				Error:     fmt.Errorf("no tool executor configured"),
			})
		}

		messages = append(messages, models.Message{
			Role:    "assistant",
			Content: resp.Content,
		})
		messages = append(messages, models.Message{
			Role:    "user",
			Content: "Tool execution is not available. Please provide your best answer without tools.",
		})
	}

	// Max iterations reached without a final answer.
	w.task.Status = AgenticTaskCompleted
	finalContent := "Max iterations reached without a final answer"
	if len(messages) > 0 {
		last := messages[len(messages)-1]
		if last.Role == "assistant" && last.Content != "" {
			finalContent = last.Content
		}
	}
	return AgenticResult{
		TaskID:    w.task.ID,
		AgentID:   w.id,
		Content:   finalContent,
		ToolCalls: allToolExecs,
		Duration:  time.Since(start),
		Error:     nil,
	}
}

// ID returns the worker's identifier.
func (w *AgentWorker) ID() string {
	return w.id
}

// Task returns a copy of the worker's task.
func (w *AgentWorker) Task() AgenticTask {
	return w.task
}

// MaxIterations returns the configured maximum iteration count.
func (w *AgentWorker) MaxIterations() int {
	return w.maxIterations
}
