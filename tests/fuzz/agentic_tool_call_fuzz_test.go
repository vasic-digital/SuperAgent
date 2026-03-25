//go:build fuzz

// Package fuzz provides Go native fuzzing tests for the AgenticEnsemble pipeline.
// These tests verify that malformed or adversarial input never causes panics or
// undefined behavior in tool call parameter handling and plan parsing.
package fuzz

import (
	"encoding/json"
	"testing"

	"dev.helix.agent/internal/services"
)

// FuzzAgenticToolCallParameters fuzzes arbitrary byte sequences as tool call
// input parameters to verify the AgenticToolExecution struct handles them safely.
func FuzzAgenticToolCallParameters(f *testing.F) {
	// Seed corpus — valid tool call inputs
	f.Add([]byte(`{"protocol":"mcp","operation":"read_file","input":{"path":"/tmp/test.go"}}`))
	f.Add([]byte(`{"protocol":"lsp","operation":"diagnostics","input":null}`))
	f.Add([]byte(`{"protocol":"embeddings","operation":"search","input":{"query":"test"}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Add([]byte(`{"protocol":"","operation":"","input":{}}`))
	f.Add([]byte(`{"protocol":"mcp","operation":"write","input":{"path":"../../../etc/passwd"}}`))
	f.Add([]byte("\x00\x01\x02\xff\xfe"))
	f.Add([]byte(`{"protocol":"mcp","operation":"exec","input":{"cmd":"$(rm -rf /)"}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Attempt to parse as a tool execution — must never panic
		var exec map[string]interface{}
		_ = json.Unmarshal(data, &exec)

		// Attempt to build an AgenticToolExecution from parsed data (no panic expected)
		if exec != nil {
			toolExec := services.AgenticToolExecution{}

			if proto, ok := exec["protocol"].(string); ok {
				toolExec.Protocol = proto
			}
			if op, ok := exec["operation"].(string); ok {
				toolExec.Operation = op
			}
			if input, ok := exec["input"]; ok {
				toolExec.Input = input
			}

			// Re-marshal and unmarshal to test round-trip safety
			reEncoded, err := json.Marshal(toolExec)
			if err == nil {
				var decoded services.AgenticToolExecution
				_ = json.Unmarshal(reEncoded, &decoded)
			}
		}

		// Also test as a slice of tool executions
		var execSlice []map[string]interface{}
		_ = json.Unmarshal(data, &execSlice)
		for _, item := range execSlice {
			if item == nil {
				continue
			}
			toolExec := services.AgenticToolExecution{}
			if proto, ok := item["protocol"].(string); ok {
				toolExec.Protocol = proto
			}
			_ = toolExec
		}
	})
}

// FuzzAgenticPlanParsing fuzzes the plan decomposition JSON to verify the
// AgenticTask parsing is resilient to malformed input.
func FuzzAgenticPlanParsing(f *testing.F) {
	// Seed corpus — valid plan payloads
	f.Add([]byte(`[{"id":"t1","description":"task one","priority":1,"estimated_steps":3,"status":0}]`))
	f.Add([]byte(`[{"id":"t2","description":"task two","dependencies":["t1"],"tool_requirements":["mcp"],"priority":2,"estimated_steps":5,"status":0}]`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`[{}]`))
	f.Add([]byte(`[{"id":"","description":"","priority":-1,"estimated_steps":0,"status":99}]`))
	f.Add([]byte("\x00\xff"))
	f.Add([]byte(`[{"id":"t1","dependencies":["t1"],"status":0}]`)) // self-dependency
	f.Add([]byte(`[{"id":"t1"},{"id":"t1"}]`))                      // duplicate IDs

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parse as a task slice — must never panic
		var tasks []services.AgenticTask
		_ = json.Unmarshal(data, &tasks)

		for i := range tasks {
			// Access each field safely
			_ = tasks[i].ID
			_ = tasks[i].Description
			_ = tasks[i].Priority
			_ = tasks[i].EstimatedSteps
			_ = tasks[i].Status.String()

			for _, dep := range tasks[i].Dependencies {
				_ = dep
			}
			for _, req := range tasks[i].ToolRequirements {
				_ = req
			}

			// Simulate status transition
			tasks[i].Status = services.AgenticTaskRunning
			_ = tasks[i].Status.String()
		}

		// Re-marshal and unmarshal round-trip
		reEncoded, err := json.Marshal(tasks)
		if err == nil {
			var decoded []services.AgenticTask
			_ = json.Unmarshal(reEncoded, &decoded)
		}

		// Also try parsing as a single task
		var singleTask services.AgenticTask
		_ = json.Unmarshal(data, &singleTask)
		_ = singleTask.Status.String()
	})
}
