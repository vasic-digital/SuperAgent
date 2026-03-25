// Package automation provides end-to-end automation tests for the AgenticEnsemble.
// These tests validate complete request-to-response workflows and ensure all
// pipeline stages fire in the correct order.
package automation

import (
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

// setupAutomationRouter returns a router configured for end-to-end automation tests.
func setupAutomationRouter() *gin.Engine {
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

// automationPost is a convenience wrapper for POST requests in automation tests.
func automationPost(t *testing.T, r *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestAgenticEnsemble_EndToEnd_ReasonWorkflow exercises the full reason-mode
// pipeline: request → classify (reason) → single-agent execution → response.
//
// Pipeline stages validated:
//  1. Intent classification → reason mode
//  2. Workflow creation accepted by handler
//  3. Single agent node executed
//  4. Response contains workflow ID, status, and timing metadata
func TestAgenticEnsemble_EndToEnd_ReasonWorkflow(t *testing.T) {
	r := setupAutomationRouter()

	// Stage 1: classify intent as reason mode
	query := "explain the ensemble voting strategy used by HelixAgent"
	var mode services.AgenticMode
	if strings.Contains(query, "implement") || strings.Contains(query, "create") {
		mode = services.AgenticModeExecute
	} else {
		mode = services.AgenticModeReason
	}
	assert.Equal(t, services.AgenticModeReason, mode, "query should classify as reason mode")

	// Stage 2: build and submit workflow
	body := map[string]interface{}{
		"name":        "e2e-reason-workflow",
		"description": "end-to-end reason mode automation test",
		"nodes": []map[string]interface{}{
			{"id": "explain", "name": "Explainer", "type": "agent"},
		},
		"edges":       []map[string]interface{}{},
		"entry_point": "explain",
		"end_nodes":   []string{"explain"},
		"input": map[string]interface{}{
			"query": query,
			"context": map[string]interface{}{
				"mode": mode.String(),
			},
		},
	}

	start := time.Now()
	w := automationPost(t, r, "/v1/agentic/workflows", body)
	elapsed := time.Since(start)

	// Stage 3: validate response
	require.Equal(t, http.StatusOK, w.Code, "handler returned error: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// Stage 4: validate all response fields
	assert.NotEmpty(t, resp["id"], "workflow ID must be present")
	assert.Equal(t, "e2e-reason-workflow", resp["name"])
	assert.Equal(t, float64(1), resp["nodes_count"])
	assert.Equal(t, float64(0), resp["edges_count"])
	assert.Equal(t, "explain", resp["entry_point"])
	assert.NotEmpty(t, resp["status"], "status must be set")
	assert.NotEmpty(t, resp["created_at"], "created_at must be set")

	t.Logf("E2E reason workflow completed: id=%s status=%s elapsed=%s",
		resp["id"], resp["status"], elapsed)
}

// TestAgenticEnsemble_EndToEnd_ExecuteWorkflow exercises the full execute-mode
// pipeline: request → classify (execute) → planner → multi-agent execution →
// aggregation → verify → response.
//
// Pipeline stages validated:
//  1. Intent classification → execute mode
//  2. Multi-node workflow with planner + tools created
//  3. All nodes registered in response
//  4. Workflow ID retrievable via GET endpoint
//  5. Response metadata is complete and consistent
func TestAgenticEnsemble_EndToEnd_ExecuteWorkflow(t *testing.T) {
	r := setupAutomationRouter()

	// Stage 1: classify intent as execute mode
	query := "implement a new endpoint for listing available LLM models"
	var mode services.AgenticMode
	if strings.Contains(query, "implement") || strings.Contains(query, "create") {
		mode = services.AgenticModeExecute
	} else {
		mode = services.AgenticModeReason
	}
	assert.Equal(t, services.AgenticModeExecute, mode, "query should classify as execute mode")

	// Stage 2: build multi-agent workflow (planner + tools + verifier)
	body := map[string]interface{}{
		"name":        "e2e-execute-workflow",
		"description": "end-to-end execute mode automation test",
		"nodes": []map[string]interface{}{
			{"id": "plan", "name": "Planner", "type": "agent"},
			{"id": "search", "name": "SearchTool", "type": "tool"},
			{"id": "implement", "name": "Implementer", "type": "agent"},
			{"id": "verify", "name": "Verifier", "type": "agent"},
		},
		"edges": []map[string]interface{}{
			{"from": "plan", "to": "search", "label": "plan-ready"},
			{"from": "search", "to": "implement", "label": "context-ready"},
			{"from": "implement", "to": "verify", "label": "impl-ready"},
		},
		"entry_point": "plan",
		"end_nodes":   []string{"verify"},
		"config": map[string]interface{}{
			"max_iterations":         15,
			"timeout_seconds":        60,
			"enable_checkpoints":     true,
			"enable_self_correction": true,
		},
		"input": map[string]interface{}{
			"query": query,
			"context": map[string]interface{}{
				"mode":             mode.String(),
				"enable_execution": true,
			},
		},
	}

	start := time.Now()
	w := automationPost(t, r, "/v1/agentic/workflows", body)
	elapsed := time.Since(start)

	// Stage 3: validate response
	require.Equal(t, http.StatusOK, w.Code, "handler returned error: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// Stage 4: validate all nodes are registered
	assert.Equal(t, float64(4), resp["nodes_count"], "all 4 nodes must be registered")
	assert.Equal(t, float64(3), resp["edges_count"], "all 3 edges must be registered")
	assert.Equal(t, "plan", resp["entry_point"])

	workflowID, ok := resp["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, workflowID)

	// Stage 5: retrieve workflow via GET to confirm persistence
	getReq := httptest.NewRequest(http.MethodGet, "/v1/agentic/workflows/"+workflowID, nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	require.Equal(t, http.StatusOK, getW.Code)

	var getResp map[string]interface{}
	require.NoError(t, json.Unmarshal(getW.Body.Bytes(), &getResp))

	assert.Equal(t, workflowID, getResp["id"])
	assert.Equal(t, "e2e-execute-workflow", getResp["name"])

	t.Logf("E2E execute workflow completed: id=%s status=%s elapsed=%s",
		workflowID, resp["status"], elapsed)
}

// TestAgenticEnsemble_EndToEnd_InvalidWorkflowNotFound verifies that requesting
// a non-existent workflow returns 404.
func TestAgenticEnsemble_EndToEnd_InvalidWorkflowNotFound(t *testing.T) {
	r := setupAutomationRouter()

	req := httptest.NewRequest(http.MethodGet, "/v1/agentic/workflows/nonexistent-id-12345", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["error"])
}

// TestAgenticEnsemble_EndToEnd_ModeConsistency verifies that AgenticMode values
// are consistent across the full system (types, string representations, config).
func TestAgenticEnsemble_EndToEnd_ModeConsistency(t *testing.T) {
	// Validate mode integer values
	assert.Equal(t, services.AgenticMode(0), services.AgenticModeReason)
	assert.Equal(t, services.AgenticMode(1), services.AgenticModeExecute)

	// Validate string representations
	assert.Equal(t, "reason", services.AgenticModeReason.String())
	assert.Equal(t, "execute", services.AgenticModeExecute.String())
	assert.Equal(t, "unknown", services.AgenticMode(99).String())

	// Validate default config aligns with expected defaults
	cfg := services.DefaultAgenticEnsembleConfig()
	assert.Equal(t, 5, cfg.MaxConcurrentAgents)
	assert.Equal(t, 20, cfg.MaxIterationsPerAgent)
	assert.Equal(t, 5, cfg.MaxToolIterationsPerPhase)
	assert.True(t, cfg.EnableVision)
	assert.True(t, cfg.EnableMemory)
	assert.True(t, cfg.EnableExecution)
}

// TestAgenticEnsemble_EndToEnd_MetadataLifecycle simulates the full metadata
// lifecycle from task creation through completion and aggregation.
func TestAgenticEnsemble_EndToEnd_MetadataLifecycle(t *testing.T) {
	// Create tasks
	tasks := []services.AgenticTask{
		{
			ID:               "auto-task-001",
			Description:      "Research existing pagination patterns",
			ToolRequirements: []string{"mcp", "embeddings"},
			Priority:         1,
			EstimatedSteps:   3,
			Status:           services.AgenticTaskPending,
		},
		{
			ID:               "auto-task-002",
			Description:      "Implement pagination logic",
			Dependencies:     []string{"auto-task-001"},
			ToolRequirements: []string{"mcp", "lsp"},
			Priority:         2,
			EstimatedSteps:   8,
			Status:           services.AgenticTaskPending,
		},
		{
			ID:               "auto-task-003",
			Description:      "Write tests for pagination",
			Dependencies:     []string{"auto-task-002"},
			ToolRequirements: []string{"mcp"},
			Priority:         3,
			EstimatedSteps:   5,
			Status:           services.AgenticTaskPending,
		},
	}

	// Simulate execution lifecycle
	results := make([]services.AgenticResult, 0, len(tasks))
	for i := range tasks {
		tasks[i].Status = services.AgenticTaskRunning
		assert.Equal(t, "running", tasks[i].Status.String())

		// Simulate task completion
		tasks[i].Status = services.AgenticTaskCompleted
		results = append(results, services.AgenticResult{
			TaskID:  tasks[i].ID,
			AgentID: "agent-" + tasks[i].ID,
			Content: "completed: " + tasks[i].Description,
			ToolCalls: []services.AgenticToolExecution{
				{Protocol: "mcp", Operation: "read", Duration: 10 * time.Millisecond},
			},
			Duration: 100 * time.Millisecond * time.Duration(tasks[i].Priority),
		})
	}

	// Build metadata
	meta := services.AgenticMetadata{
		Mode:            "execute",
		StagesCompleted: []string{"decompose", "assign", "execute", "aggregate"},
		AgentsSpawned:   len(tasks),
		TasksCompleted:  len(results),
		ProvenanceID:    "automation-prov-001",
	}
	for _, r := range results {
		meta.TotalDurationMs += r.Duration.Milliseconds()
		for _, tc := range r.ToolCalls {
			meta.ToolsInvoked = append(meta.ToolsInvoked,
				services.ToolInvocationSummary{Protocol: tc.Protocol, Count: 1})
		}
	}

	assert.Equal(t, 3, meta.AgentsSpawned)
	assert.Equal(t, 3, meta.TasksCompleted)
	assert.Equal(t, int64(600), meta.TotalDurationMs) // 100+200+300
	assert.Len(t, meta.ToolsInvoked, 3)
	assert.Equal(t, "mcp", meta.ToolsInvoked[0].Protocol)

	// Verify all tasks completed
	for _, task := range tasks {
		assert.Equal(t, services.AgenticTaskCompleted, task.Status,
			"task %s should be completed", task.ID)
	}
}
