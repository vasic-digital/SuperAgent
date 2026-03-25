package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

// agentWorkerMockComplete returns a CompleteFunc that yields a fixed response on
// every call. When toolCalls is non-empty the first call returns them; all
// subsequent calls return a plain content response.
func agentWorkerMockComplete(
	content string,
	toolCalls []models.ToolCall,
) CompleteFunc {
	calls := 0
	return func(
		_ context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		calls++
		if calls == 1 && len(toolCalls) > 0 {
			return &models.LLMResponse{
				Content:   "",
				ToolCalls: toolCalls,
			}, nil
		}
		return &models.LLMResponse{
			Content:   content,
			ToolCalls: nil,
		}, nil
	}
}

// agentWorkerErrComplete returns a CompleteFunc that always errors.
func agentWorkerErrComplete(err error) CompleteFunc {
	return func(
		_ context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		return nil, err
	}
}

func TestAgentWorker_SingleStep(t *testing.T) {
	task := AgenticTask{
		ID:          "task-single",
		Description: "Summarize the code",
		Status:      AgenticTaskPending,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	worker := NewAgentWorker(
		"agent-1", task, nil, 10, logger,
	)

	complete := agentWorkerMockComplete("The code is well structured.", nil)

	result := worker.Execute(context.Background(), complete)

	assert.Equal(t, "task-single", result.TaskID)
	assert.Equal(t, "agent-1", result.AgentID)
	assert.Equal(t, "The code is well structured.", result.Content)
	assert.NoError(t, result.Error)
	assert.Empty(t, result.ToolCalls)
	assert.True(t, result.Duration > 0)
	assert.Equal(t, AgenticTaskCompleted, worker.Task().Status)
}

func TestAgentWorker_MaxIterations(t *testing.T) {
	task := AgenticTask{
		ID:          "task-maxiter",
		Description: "Infinite tool loop",
		Status:      AgenticTaskPending,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// completeFunc always returns tool calls, forcing the loop to hit
	// maxIterations without the tool executor.
	alwaysToolCalls := func(
		_ context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		return &models.LLMResponse{
			Content: "need tools",
			ToolCalls: []models.ToolCall{
				{
					ID:   "call-1",
					Type: "function",
					Function: models.ToolCallFunction{
						Name:      "read_file",
						Arguments: `{"path":"/tmp/x"}`,
					},
				},
			},
		}, nil
	}

	maxIter := 3
	worker := NewAgentWorker(
		"agent-max", task, nil, maxIter, logger,
	)

	result := worker.Execute(context.Background(), alwaysToolCalls)

	assert.Equal(t, "task-maxiter", result.TaskID)
	assert.NoError(t, result.Error)
	// Should have recorded one unanswered tool exec per iteration.
	assert.Len(t, result.ToolCalls, maxIter)
	for _, tc := range result.ToolCalls {
		assert.Error(t, tc.Error)
		assert.Equal(t, "read_file", tc.Operation)
	}
	assert.Equal(t, maxIter, worker.MaxIterations())
}

func TestAgentWorker_ContextCancellation(t *testing.T) {
	task := AgenticTask{
		ID:          "task-cancel",
		Description: "Long running task",
		Status:      AgenticTaskPending,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	worker := NewAgentWorker("agent-cancel", task, nil, 10, logger)

	result := worker.Execute(ctx, agentWorkerMockComplete("done", nil))

	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "context cancelled")
	assert.Equal(t, AgenticTaskFailed, worker.Task().Status)
}

func TestAgentWorker_LLMError(t *testing.T) {
	task := AgenticTask{
		ID:          "task-err",
		Description: "Will fail",
		Status:      AgenticTaskPending,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	worker := NewAgentWorker("agent-err", task, nil, 5, logger)

	result := worker.Execute(
		context.Background(),
		agentWorkerErrComplete(errors.New("provider unavailable")),
	)

	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "provider unavailable")
	assert.Equal(t, AgenticTaskFailed, worker.Task().Status)
}

func TestAgentWorker_NilResponse(t *testing.T) {
	task := AgenticTask{
		ID:          "task-nil",
		Description: "Nil response test",
		Status:      AgenticTaskPending,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	nilFunc := func(
		_ context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		return nil, nil
	}

	worker := NewAgentWorker("agent-nil", task, nil, 5, logger)

	result := worker.Execute(context.Background(), nilFunc)

	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "nil response")
	assert.Equal(t, AgenticTaskFailed, worker.Task().Status)
}

func TestAgentWorker_DefaultMaxIterations(t *testing.T) {
	task := AgenticTask{
		ID:          "task-default",
		Description: "Default iter",
		Status:      AgenticTaskPending,
	}

	worker := NewAgentWorker("agent-def", task, nil, 0, nil)

	assert.Equal(t, 10, worker.MaxIterations())
	assert.Equal(t, "agent-def", worker.ID())
	assert.Equal(t, "task-default", worker.Task().ID)
}

func TestAgentWorker_Accessors(t *testing.T) {
	task := AgenticTask{
		ID:             "task-acc",
		Description:    "Accessor test",
		Dependencies:   []string{"dep-1"},
		Priority:       3,
		EstimatedSteps: 5,
		Status:         AgenticTaskPending,
	}

	worker := NewAgentWorker("agent-acc", task, nil, 7, nil)

	assert.Equal(t, "agent-acc", worker.ID())
	assert.Equal(t, 7, worker.MaxIterations())

	got := worker.Task()
	assert.Equal(t, "task-acc", got.ID)
	assert.Equal(t, "Accessor test", got.Description)
	assert.Equal(t, 3, got.Priority)
}

func TestAgentWorker_ToolCallsWithoutExecutor(t *testing.T) {
	task := AgenticTask{
		ID:          "task-notool",
		Description: "Tool calls but no executor",
		Status:      AgenticTaskPending,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	callCount := 0
	complete := func(
		_ context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		callCount++
		if callCount == 1 {
			return &models.LLMResponse{
				Content: "",
				ToolCalls: []models.ToolCall{
					{
						ID:   "call-1",
						Type: "function",
						Function: models.ToolCallFunction{
							Name:      "search",
							Arguments: `{"q":"hello"}`,
						},
					},
				},
			}, nil
		}
		// Second call: model gives final answer after being told
		// tools are unavailable.
		return &models.LLMResponse{
			Content: "Here is my best answer without tools.",
		}, nil
	}

	worker := NewAgentWorker("agent-notool", task, nil, 5, logger)

	result := worker.Execute(context.Background(), complete)

	assert.NoError(t, result.Error)
	assert.Equal(t, "Here is my best answer without tools.", result.Content)
	require.Len(t, result.ToolCalls, 1)
	assert.Error(t, result.ToolCalls[0].Error)
	assert.Equal(t, "search", result.ToolCalls[0].Operation)
	assert.Equal(t, AgenticTaskCompleted, worker.Task().Status)
}

func TestAgentWorker_DurationIsPositive(t *testing.T) {
	task := AgenticTask{
		ID:          "task-dur",
		Description: "Duration check",
		Status:      AgenticTaskPending,
	}

	worker := NewAgentWorker("agent-dur", task, nil, 1, nil)

	result := worker.Execute(
		context.Background(),
		agentWorkerMockComplete("fast answer", nil),
	)

	assert.NoError(t, result.Error)
	assert.True(
		t, result.Duration > 0,
		"duration should be positive, got %v", result.Duration,
	)
	assert.True(
		t, result.Duration < 5*time.Second,
		"mock should complete quickly, got %v", result.Duration,
	)
}
