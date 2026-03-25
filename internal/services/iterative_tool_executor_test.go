package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// mockIterativeLLMFunc builds a CompleteFunc that returns responses from the
// provided slice in order. Each invocation pops the next response. When
// the slice is exhausted it returns an error.
func mockIterativeLLMFunc(
	responses []*models.LLMResponse,
	errs []error,
) CompleteFunc {
	idx := 0
	return func(_ context.Context, _ []models.Message) (
		*models.LLMResponse, error,
	) {
		if idx >= len(responses) {
			return nil, fmt.Errorf("no more mock responses (call %d)", idx)
		}
		resp := responses[idx]
		var err error
		if idx < len(errs) {
			err = errs[idx]
		}
		idx++
		return resp, err
	}
}

// callCounter wraps a CompleteFunc and counts invocations.
type callCounter struct {
	fn    CompleteFunc
	count int
}

func (c *callCounter) call(
	ctx context.Context, msgs []models.Message,
) (*models.LLMResponse, error) {
	c.count++
	return c.fn(ctx, msgs)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestIterativeToolExecutor_NoToolCalls_Passthrough(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	exec := NewIterativeToolExecutor(nil, 5, 30*time.Second, logger)

	resp := &models.LLMResponse{
		Content:    "Hello, world!",
		ToolCalls:  nil,
		ProviderID: "test-provider",
	}

	counter := &callCounter{fn: mockIterativeLLMFunc(
		[]*models.LLMResponse{resp}, nil,
	)}

	result, toolExecs, err := exec.ExecuteWithTools(
		context.Background(), counter.call,
		[]models.Message{{Role: "user", Content: "Hi"}},
	)

	require.NoError(t, err)
	assert.Equal(t, "Hello, world!", result.Content)
	assert.Empty(t, toolExecs)
	assert.Equal(t, 1, counter.count, "should call LLM exactly once")
}

func TestIterativeToolExecutor_SingleToolIteration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	// First response has a tool call; second response is final.
	respWithTool := &models.LLMResponse{
		Content: "",
		ToolCalls: []models.ToolCall{
			{
				ID:   "call_1",
				Type: "function",
				Function: models.ToolCallFunction{
					Name:      "mcp_test_tool",
					Arguments: `{"key": "value"}`,
				},
			},
		},
	}
	finalResp := &models.LLMResponse{
		Content:   "Final answer after tool use",
		ToolCalls: nil,
	}

	counter := &callCounter{fn: mockIterativeLLMFunc(
		[]*models.LLMResponse{respWithTool, finalResp}, nil,
	)}

	// nil toolBridge — tool invocation will fail but executor should still
	// proceed and append the error as a tool result.
	exec := NewIterativeToolExecutor(nil, 5, 30*time.Second, logger)

	result, toolExecs, err := exec.ExecuteWithTools(
		context.Background(), counter.call,
		[]models.Message{{Role: "user", Content: "Use a tool"}},
	)

	require.NoError(t, err)
	assert.Equal(t, "Final answer after tool use", result.Content)
	assert.Len(t, toolExecs, 1)
	assert.Equal(t, "mcp_test_tool", toolExecs[0].Operation)
	assert.NotNil(t, toolExecs[0].Error, "should record error for nil bridge")
	assert.Equal(t, 2, counter.count, "one for tool call, one for final")
}

func TestIterativeToolExecutor_MultipleIterations(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	resp1 := &models.LLMResponse{
		Content: "",
		ToolCalls: []models.ToolCall{
			{
				ID:   "call_1",
				Type: "function",
				Function: models.ToolCallFunction{
					Name:      "rag_search",
					Arguments: `{"query": "test"}`,
				},
			},
		},
	}
	resp2 := &models.LLMResponse{
		Content: "",
		ToolCalls: []models.ToolCall{
			{
				ID:   "call_2",
				Type: "function",
				Function: models.ToolCallFunction{
					Name:      "lsp_get_definition",
					Arguments: `{"file": "main.go", "line": 10}`,
				},
			},
		},
	}
	resp3 := &models.LLMResponse{
		Content:   "Done after two tool rounds",
		ToolCalls: nil,
	}

	counter := &callCounter{fn: mockIterativeLLMFunc(
		[]*models.LLMResponse{resp1, resp2, resp3}, nil,
	)}

	exec := NewIterativeToolExecutor(nil, 10, 30*time.Second, logger)

	result, toolExecs, err := exec.ExecuteWithTools(
		context.Background(), counter.call,
		[]models.Message{{Role: "user", Content: "Multi-step"}},
	)

	require.NoError(t, err)
	assert.Equal(t, "Done after two tool rounds", result.Content)
	assert.Len(t, toolExecs, 2)
	assert.Equal(t, "rag_search", toolExecs[0].Operation)
	assert.Equal(t, "rag", toolExecs[0].Protocol)
	assert.Equal(t, "lsp_get_definition", toolExecs[1].Operation)
	assert.Equal(t, "lsp", toolExecs[1].Protocol)
	assert.Equal(t, 3, counter.count)
}

func TestIterativeToolExecutor_MaxIterationsReached(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	// Every response will contain a tool call — we want to hit the limit.
	makeToolResp := func(id string) *models.LLMResponse {
		return &models.LLMResponse{
			Content: "",
			ToolCalls: []models.ToolCall{
				{
					ID:   id,
					Type: "function",
					Function: models.ToolCallFunction{
						Name:      "mcp_loop_tool",
						Arguments: `{}`,
					},
				},
			},
		}
	}

	maxIter := 3
	// We need maxIter tool responses + 1 final response after the loop.
	responses := make([]*models.LLMResponse, 0, maxIter+1)
	for i := 0; i < maxIter; i++ {
		responses = append(responses, makeToolResp(
			fmt.Sprintf("call_%d", i),
		))
	}
	// The final response (after max reached) has no tool calls.
	responses = append(responses, &models.LLMResponse{
		Content:   "Forced final after max iterations",
		ToolCalls: nil,
	})

	counter := &callCounter{fn: mockIterativeLLMFunc(responses, nil)}
	exec := NewIterativeToolExecutor(nil, maxIter, 30*time.Second, logger)

	result, toolExecs, err := exec.ExecuteWithTools(
		context.Background(), counter.call,
		[]models.Message{{Role: "user", Content: "loop"}},
	)

	require.NoError(t, err)
	assert.Equal(t, "Forced final after max iterations", result.Content)
	assert.Len(t, toolExecs, maxIter)
	// maxIter iterations + 1 final call
	assert.Equal(t, maxIter+1, counter.count)
}

func TestIterativeToolExecutor_ContextCancellation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	failFunc := func(
		ctx context.Context, _ []models.Message,
	) (*models.LLMResponse, error) {
		return nil, ctx.Err()
	}

	exec := NewIterativeToolExecutor(nil, 5, 30*time.Second, logger)

	result, toolExecs, err := exec.ExecuteWithTools(
		ctx, failFunc,
		[]models.Message{{Role: "user", Content: "cancelled"}},
	)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Empty(t, toolExecs)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestIterativeToolExecutor_NilToolBridge_GracefulHandling(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	// Response with a tool call but no bridge configured.
	respWithTool := &models.LLMResponse{
		Content: "",
		ToolCalls: []models.ToolCall{
			{
				ID:   "call_1",
				Type: "function",
				Function: models.ToolCallFunction{
					Name:      "mcp_some_tool",
					Arguments: `{"x": 1}`,
				},
			},
		},
	}
	finalResp := &models.LLMResponse{
		Content:   "Recovered after tool error",
		ToolCalls: nil,
	}

	counter := &callCounter{fn: mockIterativeLLMFunc(
		[]*models.LLMResponse{respWithTool, finalResp}, nil,
	)}

	exec := NewIterativeToolExecutor(nil, 5, 30*time.Second, logger)

	result, toolExecs, err := exec.ExecuteWithTools(
		context.Background(), counter.call,
		[]models.Message{{Role: "user", Content: "test"}},
	)

	require.NoError(t, err)
	assert.Equal(t, "Recovered after tool error", result.Content)
	assert.Len(t, toolExecs, 1)
	assert.Error(t, toolExecs[0].Error)
	assert.Contains(t, toolExecs[0].Error.Error(), "tool bridge not configured")
}

func TestIterativeToolExecutor_LLMError(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	llmErr := errors.New("provider unavailable")
	fn := mockIterativeLLMFunc(
		[]*models.LLMResponse{nil},
		[]error{llmErr},
	)

	exec := NewIterativeToolExecutor(nil, 5, 30*time.Second, logger)

	result, toolExecs, err := exec.ExecuteWithTools(
		context.Background(), fn,
		[]models.Message{{Role: "user", Content: "fail"}},
	)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Empty(t, toolExecs)
	assert.Contains(t, err.Error(), "provider unavailable")
	assert.Contains(t, err.Error(), "iteration 0")
}

func TestIterativeToolExecutor_NilResponse(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	fn := mockIterativeLLMFunc(
		[]*models.LLMResponse{nil},
		[]error{nil},
	)

	exec := NewIterativeToolExecutor(nil, 5, 30*time.Second, logger)

	result, _, err := exec.ExecuteWithTools(
		context.Background(), fn,
		[]models.Message{{Role: "user", Content: "nil resp"}},
	)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "nil response")
}

func TestIterativeToolExecutor_DefaultValues(t *testing.T) {
	exec := NewIterativeToolExecutor(nil, 0, 0, nil)

	assert.Equal(t, 5, exec.MaxIterations())
	assert.Equal(t, 30*time.Second, exec.IterationTimeout())
	assert.NotNil(t, exec.logger)
}

func TestIterativeToolExecutor_DetectProtocol(t *testing.T) {
	exec := NewIterativeToolExecutor(nil, 1, time.Second, nil)

	tests := []struct {
		name     string
		expected string
	}{
		{"rag_search", "rag"},
		{"lsp_get_definition", "lsp"},
		{"embed_text", "embedding"},
		{"format_go", "formatter"},
		{"some_mcp_tool", "mcp"},
		{"unknown", "mcp"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, exec.detectProtocol(tc.name))
		})
	}
}

func TestParseToolArguments(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		args, err := parseToolArguments(`{"key": "value", "num": 42}`)
		require.NoError(t, err)
		assert.Equal(t, "value", args["key"])
		assert.Equal(t, float64(42), args["num"])
	})

	t.Run("empty string", func(t *testing.T) {
		args, err := parseToolArguments("")
		require.NoError(t, err)
		assert.NotNil(t, args)
		assert.Empty(t, args)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := parseToolArguments("not json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON")
	})
}

func TestIterativeToolExecutor_HasToolCalls(t *testing.T) {
	exec := NewIterativeToolExecutor(nil, 1, time.Second, nil)

	assert.False(t, exec.hasToolCalls(nil))
	assert.False(t, exec.hasToolCalls(&models.LLMResponse{}))
	assert.False(t, exec.hasToolCalls(&models.LLMResponse{
		ToolCalls: []models.ToolCall{},
	}))
	assert.True(t, exec.hasToolCalls(&models.LLMResponse{
		ToolCalls: []models.ToolCall{
			{ID: "1", Function: models.ToolCallFunction{Name: "test"}},
		},
	}))
}

func TestIterativeToolExecutor_ToolCallsToMap(t *testing.T) {
	exec := NewIterativeToolExecutor(nil, 1, time.Second, nil)

	t.Run("nil input", func(t *testing.T) {
		result := exec.toolCallsToMap(nil)
		assert.Nil(t, result)
	})

	t.Run("empty input", func(t *testing.T) {
		result := exec.toolCallsToMap([]models.ToolCall{})
		assert.Nil(t, result)
	})

	t.Run("with tool calls", func(t *testing.T) {
		tcs := []models.ToolCall{
			{
				ID:   "call_abc",
				Type: "function",
				Function: models.ToolCallFunction{
					Name:      "test_tool",
					Arguments: `{"a": 1}`,
				},
			},
		}
		result := exec.toolCallsToMap(tcs)
		assert.Len(t, result, 1)
		assert.Contains(t, result, "call_abc")
	})

	t.Run("empty ID uses index", func(t *testing.T) {
		tcs := []models.ToolCall{
			{
				ID:   "",
				Type: "function",
				Function: models.ToolCallFunction{
					Name: "tool",
				},
			},
		}
		result := exec.toolCallsToMap(tcs)
		assert.Contains(t, result, "call_0")
	})
}

func TestIterativeToolExecutor_MultipleToolCallsPerIteration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	// Single response with two tool calls.
	respWithTools := &models.LLMResponse{
		Content: "",
		ToolCalls: []models.ToolCall{
			{
				ID:   "call_a",
				Type: "function",
				Function: models.ToolCallFunction{
					Name:      "mcp_tool_a",
					Arguments: `{"a": 1}`,
				},
			},
			{
				ID:   "call_b",
				Type: "function",
				Function: models.ToolCallFunction{
					Name:      "rag_search",
					Arguments: `{"query": "test"}`,
				},
			},
		},
	}
	finalResp := &models.LLMResponse{
		Content:   "Done with both tools",
		ToolCalls: nil,
	}

	counter := &callCounter{fn: mockIterativeLLMFunc(
		[]*models.LLMResponse{respWithTools, finalResp}, nil,
	)}

	exec := NewIterativeToolExecutor(nil, 5, 30*time.Second, logger)

	result, toolExecs, err := exec.ExecuteWithTools(
		context.Background(), counter.call,
		[]models.Message{{Role: "user", Content: "multi-tool"}},
	)

	require.NoError(t, err)
	assert.Equal(t, "Done with both tools", result.Content)
	assert.Len(t, toolExecs, 2)
	assert.Equal(t, "mcp_tool_a", toolExecs[0].Operation)
	assert.Equal(t, "mcp", toolExecs[0].Protocol)
	assert.Equal(t, "rag_search", toolExecs[1].Operation)
	assert.Equal(t, "rag", toolExecs[1].Protocol)
	assert.Equal(t, 2, counter.count)
}

func TestStrPtr(t *testing.T) {
	s := "hello"
	p := strPtr(s)
	require.NotNil(t, p)
	assert.Equal(t, s, *p)
}
