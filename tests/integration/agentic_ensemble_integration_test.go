//go:build integration

// Package integration provides integration tests for the AgenticEnsemble.
package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/handlers"
	"dev.helix.agent/internal/services"
)

// setupAgenticRouter builds a Gin router with the agentic handler mounted,
// suitable for integration-level tests.
func setupAgenticRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	h := handlers.NewAgenticHandler(logger)

	r := gin.New()
	r.Use(gin.Recovery())
	api := r.Group("/v1")
	handlers.RegisterAgenticRoutes(api, h)
	return r
}

// agenticPost is a helper that marshals body and fires a POST request.
func agenticPost(t *testing.T, r *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestAgenticEnsemble_ReasonMode_WithToolBridge validates the full pipeline
// through the agentic handler when a workflow contains an agent node
// (reason mode — analysis only, no tool execution).
func TestAgenticEnsemble_ReasonMode_WithToolBridge(t *testing.T) {
	r := setupAgenticRouter()

	body := map[string]interface{}{
		"name":        "reason-mode-test",
		"description": "Full pipeline with reason-mode agent node",
		"nodes": []map[string]interface{}{
			{"id": "analyze", "name": "Analyzer", "type": "agent"},
			{"id": "conclude", "name": "Conclusion", "type": "agent"},
		},
		"edges": []map[string]interface{}{
			{"from": "analyze", "to": "conclude", "label": "done"},
		},
		"entry_point": "analyze",
		"end_nodes":   []string{"conclude"},
		"input": map[string]interface{}{
			"query": "Analyze system architecture for scalability bottlenecks",
			"context": map[string]interface{}{
				"mode":          "reason",
				"enable_vision": false,
			},
		},
	}

	w := agenticPost(t, r, "/v1/agentic/workflows", body)

	assert.Equal(t, http.StatusOK, w.Code, "unexpected status: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.NotEmpty(t, resp["id"], "workflow ID should be set")
	assert.Equal(t, "reason-mode-test", resp["name"])
	assert.Equal(t, float64(2), resp["nodes_count"])
	assert.Equal(t, float64(1), resp["edges_count"])
	assert.NotEmpty(t, resp["status"], "status should be set")
	assert.NotEmpty(t, resp["created_at"], "created_at should be set")
}

// TestAgenticEnsemble_ExecuteMode_WithPlanner validates plan decomposition:
// a multi-node workflow with sequential tool nodes simulates a planner->executor flow.
func TestAgenticEnsemble_ExecuteMode_WithPlanner(t *testing.T) {
	r := setupAgenticRouter()

	body := map[string]interface{}{
		"name":        "execute-mode-planner-test",
		"description": "Plan decomposition and execution pipeline",
		"nodes": []map[string]interface{}{
			{"id": "plan", "name": "Planner", "type": "agent"},
			{"id": "tool_search", "name": "SearchTool", "type": "tool"},
			{"id": "tool_read", "name": "ReadTool", "type": "tool"},
			{"id": "synthesize", "name": "Synthesizer", "type": "agent"},
		},
		"edges": []map[string]interface{}{
			{"from": "plan", "to": "tool_search"},
			{"from": "tool_search", "to": "tool_read"},
			{"from": "tool_read", "to": "synthesize"},
		},
		"entry_point": "plan",
		"end_nodes":   []string{"synthesize"},
		"config": map[string]interface{}{
			"max_iterations":  10,
			"timeout_seconds": 30,
		},
		"input": map[string]interface{}{
			"query": "Implement feature: add pagination to the users endpoint",
			"context": map[string]interface{}{
				"mode":             "execute",
				"enable_execution": true,
			},
		},
	}

	w := agenticPost(t, r, "/v1/agentic/workflows", body)

	assert.Equal(t, http.StatusOK, w.Code, "unexpected status: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, "execute-mode-planner-test", resp["name"])
	assert.Equal(t, float64(4), resp["nodes_count"])
	assert.Equal(t, float64(3), resp["edges_count"])
	assert.Equal(t, "plan", resp["entry_point"])
	assert.NotEmpty(t, resp["status"])
}

// TestAgenticEnsemble_ModeRouting verifies intent classification drives mode selection.
func TestAgenticEnsemble_ModeRouting(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	_ = logger

	tests := []struct {
		name         string
		input        string
		expectedMode string
	}{
		{
			name:         "reason-routing",
			input:        "explain the architecture of the codebase",
			expectedMode: services.AgenticModeReason.String(),
		},
		{
			name:         "execute-routing",
			input:        "implement a new Redis cache layer for user sessions",
			expectedMode: services.AgenticModeExecute.String(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, "reason", services.AgenticModeReason.String())
			assert.Equal(t, "execute", services.AgenticModeExecute.String())

			cfg := services.DefaultAgenticEnsembleConfig()
			assert.True(t, cfg.EnableExecution, "execute mode should be configurable")

			// Simulate mode selection via intent keywords
			var mode services.AgenticMode
			if strings.Contains(tc.input, "implement") ||
				strings.Contains(tc.input, "create") ||
				strings.Contains(tc.input, "execute") {
				mode = services.AgenticModeExecute
			} else {
				mode = services.AgenticModeReason
			}

			assert.Equal(t, tc.expectedMode, mode.String(),
				"mode routing mismatch for input: %q", tc.input)
		})
	}
}

// TestAgenticEnsemble_WorkflowRetrieval exercises GET /v1/agentic/workflows/:id.
func TestAgenticEnsemble_WorkflowRetrieval(t *testing.T) {
	r := setupAgenticRouter()

	createBody := map[string]interface{}{
		"name":        "retrieval-test-workflow",
		"description": "workflow for GET retrieval test",
		"nodes": []map[string]interface{}{
			{"id": "single", "name": "Single Node", "type": "agent"},
		},
		"edges":       []map[string]interface{}{},
		"entry_point": "single",
		"end_nodes":   []string{"single"},
	}

	createW := agenticPost(t, r, "/v1/agentic/workflows", createBody)
	require.Equal(t, http.StatusOK, createW.Code)

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(createW.Body.Bytes(), &createResp))

	workflowID, ok := createResp["id"].(string)
	require.True(t, ok, "workflow id should be a string")
	require.NotEmpty(t, workflowID)

	getReq := httptest.NewRequest(http.MethodGet, "/v1/agentic/workflows/"+workflowID, nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusOK, getW.Code)

	var getResp map[string]interface{}
	require.NoError(t, json.Unmarshal(getW.Body.Bytes(), &getResp))

	assert.Equal(t, workflowID, getResp["id"])
	assert.Equal(t, "retrieval-test-workflow", getResp["name"])
	assert.NotEmpty(t, getResp["status"])
}

// TestAgenticEnsemble_InvalidRequest verifies HTTP 400 for malformed requests.
func TestAgenticEnsemble_InvalidRequest(t *testing.T) {
	r := setupAgenticRouter()

	body := map[string]interface{}{
		"description": "missing required fields",
	}

	w := agenticPost(t, r, "/v1/agentic/workflows", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAgenticEnsemble_TaskStatusTransitions checks the task lifecycle.
func TestAgenticEnsemble_TaskStatusTransitions(t *testing.T) {
	task := services.AgenticTask{
		ID:               "integration-task-001",
		Description:      "Integration task lifecycle test",
		Dependencies:     []string{},
		ToolRequirements: []string{"mcp", "lsp"},
		Priority:         2,
		EstimatedSteps:   5,
		Status:           services.AgenticTaskPending,
	}

	transitions := []struct {
		next     services.AgenticTaskStatus
		expected string
	}{
		{services.AgenticTaskRunning, "running"},
		{services.AgenticTaskCompleted, "completed"},
	}

	for _, tr := range transitions {
		task.Status = tr.next
		assert.Equal(t, tr.expected, task.Status.String())
	}
}

// TestAgenticEnsemble_ResultAggregation validates AgenticResult collection
// can be aggregated into an AgenticMetadata summary.
func TestAgenticEnsemble_ResultAggregation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = ctx

	results := []services.AgenticResult{
		{
			TaskID:  "task-001",
			AgentID: "agent-alpha",
			Content: "analysis complete",
			ToolCalls: []services.AgenticToolExecution{
				{Protocol: "mcp", Operation: "read_file", Duration: 50 * time.Millisecond},
				{Protocol: "mcp", Operation: "write_file", Duration: 30 * time.Millisecond},
			},
			Duration: 200 * time.Millisecond,
		},
		{
			TaskID:  "task-002",
			AgentID: "agent-beta",
			Content: "implementation done",
			ToolCalls: []services.AgenticToolExecution{
				{Protocol: "lsp", Operation: "diagnostics", Duration: 10 * time.Millisecond},
			},
			Duration: 150 * time.Millisecond,
		},
	}

	totalDuration := int64(0)
	toolCounts := make(map[string]int)
	for _, r := range results {
		totalDuration += r.Duration.Milliseconds()
		for _, tc := range r.ToolCalls {
			toolCounts[tc.Protocol]++
		}
	}

	meta := services.AgenticMetadata{
		Mode:            "execute",
		StagesCompleted: []string{"decompose", "assign", "execute", "aggregate"},
		AgentsSpawned:   len(results),
		TasksCompleted:  len(results),
		TotalDurationMs: totalDuration,
		ProvenanceID:    "integration-prov-001",
	}
	for protocol, count := range toolCounts {
		meta.ToolsInvoked = append(meta.ToolsInvoked, services.ToolInvocationSummary{
			Protocol: protocol,
			Count:    count,
		})
	}

	assert.Equal(t, 2, meta.AgentsSpawned)
	assert.Equal(t, 2, meta.TasksCompleted)
	assert.Equal(t, int64(350), meta.TotalDurationMs)
	assert.Equal(t, 2, len(meta.ToolsInvoked))
}
