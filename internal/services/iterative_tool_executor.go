package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/models"
	"digital.vasic.debate/tools"
)

// IterativeToolExecutor runs LLM calls with iterative tool resolution.
// When an LLM response contains tool_calls, the executor invokes each tool
// via the ToolIntegration bridge, appends the results to the conversation,
// and re-invokes the LLM. This loop continues until the LLM returns a
// response without tool_calls or maxIterations is reached.
type IterativeToolExecutor struct {
	toolBridge       *tools.ToolIntegration
	maxIterations    int
	iterationTimeout time.Duration
	logger           *logrus.Logger
}

// NewIterativeToolExecutor creates an IterativeToolExecutor.
// maxIterations defaults to 5 when <= 0.
// iterationTimeout defaults to 30s when <= 0.
func NewIterativeToolExecutor(
	toolBridge *tools.ToolIntegration,
	maxIterations int,
	iterationTimeout time.Duration,
	logger *logrus.Logger,
) *IterativeToolExecutor {
	if maxIterations <= 0 {
		maxIterations = 5
	}
	if iterationTimeout <= 0 {
		iterationTimeout = 30 * time.Second
	}
	if logger == nil {
		logger = logrus.New()
	}
	return &IterativeToolExecutor{
		toolBridge:       toolBridge,
		maxIterations:    maxIterations,
		iterationTimeout: iterationTimeout,
		logger:           logger,
	}
}

// CompleteFunc is the signature of an LLM completion call that the executor
// wraps. It accepts the current message history and returns the LLM response.
type CompleteFunc func(ctx context.Context, messages []models.Message) (
	*models.LLMResponse, error,
)

// ExecuteWithTools runs an LLM provider via completeFunc and iteratively
// resolves any tool calls in the response. It returns the final LLM response
// and the full list of tool executions performed across all iterations.
func (e *IterativeToolExecutor) ExecuteWithTools(
	ctx context.Context,
	completeFunc CompleteFunc,
	messages []models.Message,
) (*models.LLMResponse, []AgenticToolExecution, error) {
	var allToolExecs []AgenticToolExecution
	currentMessages := make([]models.Message, len(messages))
	copy(currentMessages, messages)

	for iteration := 0; iteration < e.maxIterations; iteration++ {
		iterCtx, cancel := context.WithTimeout(ctx, e.iterationTimeout)
		resp, err := completeFunc(iterCtx, currentMessages)
		cancel()

		if err != nil {
			return nil, allToolExecs, fmt.Errorf(
				"LLM call failed at iteration %d: %w", iteration, err,
			)
		}

		if resp == nil {
			return nil, allToolExecs, fmt.Errorf(
				"nil response at iteration %d", iteration,
			)
		}

		// If the response has no tool calls we are done.
		if !e.hasToolCalls(resp) {
			return resp, allToolExecs, nil
		}

		// Execute every tool call in the response.
		toolResults, toolExecs := e.executeToolCalls(ctx, resp)
		allToolExecs = append(allToolExecs, toolExecs...)

		// Build the assistant message that contained the tool call request.
		assistantToolCalls := e.toolCallsToMap(resp.ToolCalls)
		currentMessages = append(currentMessages, models.Message{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: assistantToolCalls,
		})

		// Append one tool-result message per tool call.
		for i, result := range toolResults {
			toolCallID := ""
			if i < len(resp.ToolCalls) {
				toolCallID = resp.ToolCalls[i].ID
			}
			name := ""
			if i < len(resp.ToolCalls) {
				name = resp.ToolCalls[i].Function.Name
			}
			currentMessages = append(currentMessages, models.Message{
				Role:    "tool",
				Content: result,
				Name:    strPtr(fmt.Sprintf("%s:%s", toolCallID, name)),
			})
		}

		e.logger.WithFields(logrus.Fields{
			"iteration":  iteration + 1,
			"tool_calls": len(toolExecs),
			"max":        e.maxIterations,
		}).Debug("Tool iteration completed")
	}

	// Max iterations reached — make one final LLM call without tools so the
	// model can summarise what it has learned.
	e.logger.Warn("Max tool iterations reached, performing final LLM call")
	finalResp, err := completeFunc(ctx, currentMessages)
	if err != nil {
		return nil, allToolExecs, fmt.Errorf(
			"final LLM call after max iterations: %w", err,
		)
	}
	return finalResp, allToolExecs, nil
}

// MaxIterations returns the configured maximum iteration count.
func (e *IterativeToolExecutor) MaxIterations() int {
	return e.maxIterations
}

// IterationTimeout returns the configured per-iteration timeout.
func (e *IterativeToolExecutor) IterationTimeout() time.Duration {
	return e.iterationTimeout
}

// hasToolCalls returns true when the LLM response contains at least one tool
// call that needs resolution.
func (e *IterativeToolExecutor) hasToolCalls(resp *models.LLMResponse) bool {
	return resp != nil && len(resp.ToolCalls) > 0
}

// executeToolCalls invokes each tool call via the ToolIntegration bridge and
// returns the string results plus structured AgenticToolExecution records.
func (e *IterativeToolExecutor) executeToolCalls(
	ctx context.Context,
	resp *models.LLMResponse,
) ([]string, []AgenticToolExecution) {
	var results []string
	var execs []AgenticToolExecution

	for _, tc := range resp.ToolCalls {
		start := time.Now()
		result, err := e.invokeTool(ctx, tc)
		duration := time.Since(start)

		exec := AgenticToolExecution{
			Protocol:  e.detectProtocol(tc.Function.Name),
			Operation: tc.Function.Name,
			Input:     tc.Function.Arguments,
			Output:    result,
			Duration:  duration,
			Error:     err,
		}
		execs = append(execs, exec)

		if err != nil {
			results = append(results, fmt.Sprintf(
				"Error executing %s: %v", tc.Function.Name, err,
			))
		} else {
			resultStr, marshalErr := json.Marshal(result)
			if marshalErr != nil {
				results = append(results, fmt.Sprintf("%v", result))
			} else {
				results = append(results, string(resultStr))
			}
		}
	}
	return results, execs
}

// invokeTool routes a single tool call to the appropriate ToolIntegration
// method based on the tool name / detected protocol.
func (e *IterativeToolExecutor) invokeTool(
	ctx context.Context,
	tc models.ToolCall,
) (interface{}, error) {
	if e.toolBridge == nil {
		return nil, fmt.Errorf("tool bridge not configured")
	}
	if !e.toolBridge.IsEnabled() {
		return nil, fmt.Errorf("tool bridge is disabled")
	}

	args, err := parseToolArguments(tc.Function.Arguments)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse arguments for %s: %w", tc.Function.Name, err,
		)
	}

	name := tc.Function.Name
	protocol := e.detectProtocol(name)

	switch protocol {
	case "rag":
		query, _ := args["query"].(string)
		limit := 10
		if l, ok := args["limit"].(float64); ok {
			limit = int(l)
		}
		return e.toolBridge.QueryRAG(ctx, query, limit)

	case "lsp":
		file, _ := args["file"].(string)
		line := 0
		char := 0
		if l, ok := args["line"].(float64); ok {
			line = int(l)
		}
		if c, ok := args["character"].(float64); ok {
			char = int(c)
		}
		return e.toolBridge.GetCodeDefinition(ctx, file, line, char)

	case "embedding":
		text, _ := args["text"].(string)
		return e.toolBridge.GenerateEmbedding(ctx, text)

	case "formatter":
		language, _ := args["language"].(string)
		code, _ := args["code"].(string)
		return e.toolBridge.FormatCode(ctx, language, code)

	default:
		// Default: route through MCP.
		return e.toolBridge.InvokeMCPTool(ctx, name, args)
	}
}

// detectProtocol infers the protocol from the tool name using prefix
// conventions used by ToolIntegration.ListAvailableTools.
func (e *IterativeToolExecutor) detectProtocol(toolName string) string {
	lower := strings.ToLower(toolName)
	switch {
	case strings.HasPrefix(lower, "rag_"):
		return "rag"
	case strings.HasPrefix(lower, "lsp_"):
		return "lsp"
	case strings.HasPrefix(lower, "embed_"):
		return "embedding"
	case strings.HasPrefix(lower, "format_"):
		return "formatter"
	default:
		return "mcp"
	}
}

// toolCallsToMap converts a slice of ToolCall to the map representation
// expected by models.Message.ToolCalls.
func (e *IterativeToolExecutor) toolCallsToMap(
	tcs []models.ToolCall,
) map[string]interface{} {
	if len(tcs) == 0 {
		return nil
	}
	result := make(map[string]interface{}, len(tcs))
	for i, tc := range tcs {
		key := tc.ID
		if key == "" {
			key = fmt.Sprintf("call_%d", i)
		}
		result[key] = map[string]interface{}{
			"id":   tc.ID,
			"type": tc.Type,
			"function": map[string]interface{}{
				"name":      tc.Function.Name,
				"arguments": tc.Function.Arguments,
			},
		}
	}
	return result
}

// parseToolArguments parses the JSON arguments string from a ToolCall into a
// map. Returns an empty map when the input is empty or not valid JSON.
func parseToolArguments(argsJSON string) (map[string]interface{}, error) {
	if argsJSON == "" {
		return make(map[string]interface{}), nil
	}
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return nil, fmt.Errorf("invalid JSON arguments: %w", err)
	}
	return args, nil
}

// strPtr returns a pointer to s.
func strPtr(s string) *string {
	return &s
}
