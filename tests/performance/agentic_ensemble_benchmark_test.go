//go:build performance
// +build performance

// Package performance contains benchmark tests for AgenticEnsemble critical paths.
package performance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/handlers"
	"dev.helix.agent/internal/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newBenchAgenticRouter returns a lightweight router for benchmarks.
func newBenchAgenticRouter() *gin.Engine {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	h := handlers.NewAgenticHandler(logger)
	r := gin.New()
	api := r.Group("/v1")
	handlers.RegisterAgenticRoutes(api, h)
	return r
}

// singleNodePayload is the minimum valid workflow body.
var singleNodePayload = map[string]interface{}{
	"name":        "bench-workflow",
	"description": "benchmark test",
	"nodes": []map[string]interface{}{
		{"id": "n1", "name": "Agent", "type": "agent"},
	},
	"edges":       []map[string]interface{}{},
	"entry_point": "n1",
	"end_nodes":   []string{"n1"},
}

// multiNodePayload has a planner + 2 tools + synthesizer.
var multiNodePayload = map[string]interface{}{
	"name":        "bench-plan-workflow",
	"description": "plan benchmark test",
	"nodes": []map[string]interface{}{
		{"id": "plan", "name": "Planner", "type": "agent"},
		{"id": "t1", "name": "Tool1", "type": "tool"},
		{"id": "t2", "name": "Tool2", "type": "tool"},
		{"id": "synth", "name": "Synthesizer", "type": "agent"},
	},
	"edges": []map[string]interface{}{
		{"from": "plan", "to": "t1"},
		{"from": "t1", "to": "t2"},
		{"from": "t2", "to": "synth"},
	},
	"entry_point": "plan",
	"end_nodes":   []string{"synth"},
}

func benchmarkWorkflowRequest(b *testing.B, payload map[string]interface{}) {
	b.Helper()
	r := newBenchAgenticRouter()
	raw, err := json.Marshal(payload)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/agentic/workflows",
			strings.NewReader(string(raw)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected status %d", w.Code)
		}
	}
}

// BenchmarkAgenticEnsemble_ModeClassification measures the cost of determining
// AgenticMode from a string input (intent classification hot path).
func BenchmarkAgenticEnsemble_ModeClassification(b *testing.B) {
	inputs := []string{
		"explain how the caching layer works",
		"implement a new Redis TTL policy",
		"describe the debate orchestration design",
		"create unit tests for the provider registry",
		"what does the ensemble service do",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		var mode services.AgenticMode
		if strings.Contains(input, "implement") ||
			strings.Contains(input, "create") ||
			strings.Contains(input, "execute") {
			mode = services.AgenticModeExecute
		} else {
			mode = services.AgenticModeReason
		}
		_ = mode.String()
	}
}

// BenchmarkAgenticEnsemble_AgentSpawn measures the cost of creating a new
// agentic workflow via the HTTP handler (agent spawn overhead).
func BenchmarkAgenticEnsemble_AgentSpawn(b *testing.B) {
	benchmarkWorkflowRequest(b, singleNodePayload)
}

// BenchmarkAgenticEnsemble_ToolIterationOverhead measures the overhead of
// a multi-tool workflow creation and execution through the handler.
func BenchmarkAgenticEnsemble_ToolIterationOverhead(b *testing.B) {
	benchmarkWorkflowRequest(b, multiNodePayload)
}

// BenchmarkAgenticEnsemble_PlanDecomposition measures the cost of building
// and marshalling a complex multi-task agentic plan.
func BenchmarkAgenticEnsemble_PlanDecomposition(b *testing.B) {
	// Build a plan with 5 tasks
	tasks := make([]services.AgenticTask, 5)
	for i := range tasks {
		tasks[i] = services.AgenticTask{
			ID:               "task-bench-" + string(rune('A'+i)),
			Description:      "Benchmark task description " + string(rune('A'+i)),
			Dependencies:     []string{},
			ToolRequirements: []string{"mcp", "lsp"},
			Priority:         i + 1,
			EstimatedSteps:   i + 2,
			Status:           services.AgenticTaskPending,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		raw, err := json.Marshal(tasks)
		if err != nil {
			b.Fatal(err)
		}

		var decoded []services.AgenticTask
		if err := json.Unmarshal(raw, &decoded); err != nil {
			b.Fatal(err)
		}

		// Simulate status update during decomposition
		for j := range decoded {
			decoded[j].Status = services.AgenticTaskRunning
		}
	}
}

// BenchmarkAgenticEnsemble_ConfigCreation measures DefaultAgenticEnsembleConfig
// creation cost (called on every ensemble init).
func BenchmarkAgenticEnsemble_ConfigCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := services.DefaultAgenticEnsembleConfig()
		_ = cfg.MaxConcurrentAgents
	}
}

// BenchmarkAgenticEnsemble_MetadataAggregation benchmarks building
// AgenticMetadata from a list of results.
func BenchmarkAgenticEnsemble_MetadataAggregation(b *testing.B) {
	results := make([]services.AgenticResult, 5)
	for i := range results {
		results[i] = services.AgenticResult{
			TaskID:  "task-" + string(rune('A'+i)),
			AgentID: "agent-" + string(rune('A'+i)),
			Content: "result content",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		meta := services.AgenticMetadata{
			Mode:            "execute",
			StagesCompleted: []string{"decompose", "assign", "execute", "aggregate"},
			AgentsSpawned:   len(results),
			TasksCompleted:  len(results),
			ProvenanceID:    "bench-prov-001",
		}
		for _, r := range results {
			for _, tc := range r.ToolCalls {
				meta.ToolsInvoked = append(meta.ToolsInvoked,
					services.ToolInvocationSummary{
						Protocol: tc.Protocol,
						Count:    1,
					})
			}
		}
		_ = meta
	}
}
