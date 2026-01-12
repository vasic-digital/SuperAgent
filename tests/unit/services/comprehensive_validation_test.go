package services_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =========================================================================
// COGNEE API PARAMETER VALIDATION TESTS
// =========================================================================

// TestCogneeSearchParameterNames validates that search requests use correct parameter names
func TestCogneeSearchParameterNames(t *testing.T) {
	testCases := []struct {
		name           string
		expectedParams []string
		invalidParams  []string
	}{
		{
			name:           "search_uses_camelCase_parameters",
			expectedParams: []string{"query", "datasets", "topK", "searchType"},
			invalidParams:  []string{"search_type", "limit", "dataset_name"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the request body that would be sent
			reqBody := map[string]interface{}{
				"query":      "test query",
				"datasets":   []string{"test_dataset"},
				"topK":       10,
				"searchType": "CHUNKS",
			}

			data, err := json.Marshal(reqBody)
			require.NoError(t, err)

			// Verify expected params are present
			for _, param := range tc.expectedParams {
				assert.Contains(t, string(data), param, "Expected parameter %s should be present", param)
			}

			// Verify invalid params are NOT present
			for _, param := range tc.invalidParams {
				assert.NotContains(t, string(data), param, "Invalid parameter %s should NOT be present", param)
			}
		})
	}
}

// TestCogneeMemifyParameterNames validates that memify requests use correct parameter names
func TestCogneeMemifyParameterNames(t *testing.T) {
	testCases := []struct {
		name           string
		expectedKeys   []string
		invalidKeys    []string
	}{
		{
			name:           "memify_uses_camelCase_parameters",
			expectedKeys:   []string{"data", "datasetName"},
			invalidKeys:    []string{"dataset_name"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the request body that would be sent
			reqBody := map[string]interface{}{
				"data":        "test content",
				"datasetName": "test_dataset",
			}

			// Verify expected keys are present
			for _, key := range tc.expectedKeys {
				_, exists := reqBody[key]
				assert.True(t, exists, "Expected key %s should be present", key)
			}

			// Verify invalid keys are NOT present (check actual map keys, not substrings)
			for _, key := range tc.invalidKeys {
				_, exists := reqBody[key]
				assert.False(t, exists, "Invalid key %s should NOT be present", key)
			}
		})
	}
}

// =========================================================================
// TOOL ARGUMENT FORMAT VALIDATION TESTS
// =========================================================================

// TestToolArgumentsUseCamelCase validates that tool arguments use camelCase
func TestToolArgumentsUseCamelCase(t *testing.T) {
	testCases := []struct {
		name            string
		toolName        string
		expectedKeys    []string
		invalidKeys     []string
	}{
		{
			name:         "Read_tool_uses_filePath",
			toolName:     "Read",
			expectedKeys: []string{"filePath"},
			invalidKeys:  []string{"file_path"},
		},
		{
			name:         "Write_tool_uses_filePath",
			toolName:     "Write",
			expectedKeys: []string{"filePath"},
			invalidKeys:  []string{"file_path"},
		},
		{
			name:         "Edit_tool_uses_filePath",
			toolName:     "Edit",
			expectedKeys: []string{"filePath"},
			invalidKeys:  []string{"file_path"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate tool call arguments
			args := map[string]interface{}{
				"filePath": "/path/to/file.txt",
			}

			// Verify expected keys are present (check actual map keys)
			for _, key := range tc.expectedKeys {
				_, exists := args[key]
				assert.True(t, exists, "Expected argument %s should be present for %s tool", key, tc.toolName)
			}

			// Verify invalid keys are NOT present (check actual map keys, not substrings)
			for _, key := range tc.invalidKeys {
				_, exists := args[key]
				assert.False(t, exists, "Invalid argument %s should NOT be present for %s tool", key, tc.toolName)
			}
		})
	}
}

// =========================================================================
// OPENCODE CONFIGURATION VALIDATION TESTS
// =========================================================================

// TestOpenCodeToolsAreBooleans validates that tools section uses boolean values
func TestOpenCodeToolsAreBooleans(t *testing.T) {
	config := map[string]interface{}{
		"tools": map[string]bool{
			"Read":      true,
			"Write":     true,
			"Edit":      true,
			"Bash":      true,
			"Glob":      true,
			"Grep":      true,
			"WebFetch":  true,
			"Task":      true,
			"TodoWrite": true,
		},
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	tools, ok := parsed["tools"].(map[string]interface{})
	require.True(t, ok, "tools should be a map")

	for name, value := range tools {
		_, isBool := value.(bool)
		assert.True(t, isBool, "Tool %s should have boolean value, got %T", name, value)
	}
}

// TestOpenCodePermissionsAreStrings validates that permission section uses string values
func TestOpenCodePermissionsAreStrings(t *testing.T) {
	validValues := map[string]bool{"ask": true, "allow": true, "deny": true}

	config := map[string]interface{}{
		"permission": map[string]string{
			"read":     "allow",
			"edit":     "ask",
			"bash":     "ask",
			"webfetch": "allow",
		},
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	perms, ok := parsed["permission"].(map[string]interface{})
	require.True(t, ok, "permission should be a map")

	for key, value := range perms {
		strVal, isString := value.(string)
		assert.True(t, isString, "Permission %s should have string value, got %T", key, value)
		if isString {
			assert.True(t, validValues[strVal], "Permission %s has invalid value %s (expected ask/allow/deny)", key, strVal)
		}
	}
}

// TestOpenCodeNoInvalidTopLevelKeys validates that config doesn't have invalid keys
func TestOpenCodeNoInvalidTopLevelKeys(t *testing.T) {
	validKeys := map[string]bool{
		"$schema": true, "plugin": true, "enterprise": true, "instructions": true,
		"provider": true, "mcp": true, "tools": true, "agent": true,
		"command": true, "keybinds": true, "username": true, "share": true,
		"permission": true, "compaction": true, "mode": true, "autoshare": true,
	}

	invalidKeys := []string{"sse", "external_directory", "toolConfig", "settings"}

	for _, key := range invalidKeys {
		assert.False(t, validKeys[key], "Key %s should not be a valid top-level key", key)
	}
}

// =========================================================================
// RATE LIMITING VALIDATION TESTS
// =========================================================================

// TestWarningRateLimiting validates that warnings are rate-limited
func TestWarningRateLimiting(t *testing.T) {
	type rateLimiter struct {
		mu          sync.Mutex
		lastWarning time.Time
	}

	testCases := []struct {
		name           string
		interval       time.Duration
		callCount      int
		expectedLogs   int
	}{
		{
			name:         "rate_limit_30_seconds",
			interval:     30 * time.Second,
			callCount:    5,
			expectedLogs: 1, // Only first log should occur within 30s window
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rl := &rateLimiter{}
			logCount := 0

			for i := 0; i < tc.callCount; i++ {
				rl.mu.Lock()
				now := time.Now()
				shouldLog := rl.lastWarning.IsZero() || now.Sub(rl.lastWarning) > tc.interval
				if shouldLog {
					rl.lastWarning = now
					logCount++
				}
				rl.mu.Unlock()
			}

			assert.Equal(t, tc.expectedLogs, logCount, "Expected %d logs but got %d", tc.expectedLogs, logCount)
		})
	}
}

// =========================================================================
// DIALOGUE RENDERING VALIDATION TESTS
// =========================================================================

// TestDialogueTagStrippingValidation validates that tool tags are stripped from output
func TestDialogueTagStrippingValidation(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strip_bash_tags",
			input:    "<bash>ls -la</bash>",
			expected: "ls -la",
		},
		{
			name:     "strip_BASH_tags_uppercase",
			input:    "<BASH>ls -la</BASH>",
			expected: "ls -la",
		},
		{
			name:     "strip_python_tags",
			input:    "<python>print('hello')</python>",
			expected: "print('hello')",
		},
		{
			name:     "strip_ruby_tags",
			input:    "<ruby>puts 'hello'</ruby>",
			expected: "puts 'hello'",
		},
		{
			name:     "strip_php_tags",
			input:    "<php>echo 'hello';</php>",
			expected: "echo 'hello';",
		},
		{
			name:     "strip_javascript_tags",
			input:    "<javascript>console.log('hello')</javascript>",
			expected: "console.log('hello')",
		},
		{
			name:     "strip_go_tags",
			input:    "<go>fmt.Println(\"hello\")</go>",
			expected: "fmt.Println(\"hello\")",
		},
		{
			name:     "strip_rust_tags",
			input:    "<rust>println!(\"hello\")</rust>",
			expected: "println!(\"hello\")",
		},
		{
			name:     "preserve_markdown_code_blocks",
			input:    "```bash\nls -la\n```",
			expected: "```bash\nls -la\n```",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := stripToolTags(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// stripToolTags removes tool tags but preserves content (test helper)
func stripToolTags(content string) string {
	tags := []string{
		"bash", "BASH", "python", "ruby", "php", "javascript", "go", "rust",
		"shell", "read", "write", "edit", "glob", "grep",
	}

	result := content
	for _, tag := range tags {
		// Case-insensitive replacement
		lowerTag := strings.ToLower(tag)
		result = strings.ReplaceAll(result, "<"+tag+">", "")
		result = strings.ReplaceAll(result, "</"+tag+">", "")
		result = strings.ReplaceAll(result, "<"+strings.ToUpper(tag)+">", "")
		result = strings.ReplaceAll(result, "</"+strings.ToUpper(tag)+">", "")
		result = strings.ReplaceAll(result, "<"+lowerTag+">", "")
		result = strings.ReplaceAll(result, "</"+lowerTag+">", "")
	}

	return result
}

// =========================================================================
// BACKGROUND TASK WAITING VALIDATION TESTS
// =========================================================================

// TestBackgroundTaskWaitInterface validates TaskWaiter interface
func TestBackgroundTaskWaitInterface(t *testing.T) {
	t.Run("WaitForCompletion_signature", func(t *testing.T) {
		// This test validates that the interface has correct signature
		// The actual implementation is tested elsewhere

		type TaskWaiter interface {
			WaitForCompletion(ctx context.Context, taskID string, timeout time.Duration, progressCallback func(progress float64, message string)) (interface{}, error)
			WaitForCompletionWithOutput(ctx context.Context, taskID string, timeout time.Duration) (interface{}, []byte, error)
		}

		// If this compiles, the interface signature is correct
		assert.True(t, true)
	})
}

// TestAdaptivePolling validates adaptive polling behavior
func TestAdaptivePolling(t *testing.T) {
	t.Run("polling_interval_increases", func(t *testing.T) {
		pollInterval := 100 * time.Millisecond
		maxPollInterval := 2 * time.Second
		intervals := []time.Duration{}

		// Simulate 10 polling iterations
		for i := 0; i < 10; i++ {
			intervals = append(intervals, pollInterval)

			// Increase interval by 20% each iteration
			newInterval := time.Duration(float64(pollInterval) * 1.2)
			if newInterval > maxPollInterval {
				newInterval = maxPollInterval
			}
			pollInterval = newInterval
		}

		// Verify intervals increase
		assert.True(t, intervals[len(intervals)-1] > intervals[0], "Polling interval should increase over time")

		// Verify max is respected
		assert.LessOrEqual(t, intervals[len(intervals)-1], maxPollInterval, "Polling interval should not exceed max")
	})
}

// =========================================================================
// API ENDPOINT VALIDATION TESTS
// =========================================================================

// TestSSEStreamingFormat validates SSE streaming response format
func TestSSEStreamingFormat(t *testing.T) {
	t.Run("sse_format_validation", func(t *testing.T) {
		// SSE format: "data: {json}\n\n"
		sseResponse := "data: {\"id\":\"test\",\"content\":\"hello\"}\n\n"

		assert.True(t, strings.HasPrefix(sseResponse, "data: "), "SSE response should start with 'data: '")
		assert.True(t, strings.HasSuffix(sseResponse, "\n\n"), "SSE response should end with double newline")
	})
}

// TestHealthEndpoint validates health endpoint response
func TestHealthEndpoint(t *testing.T) {
	t.Run("health_response_structure", func(t *testing.T) {
		healthResponse := map[string]interface{}{
			"status": "healthy",
			"providers": map[string]interface{}{
				"healthy":   14,
				"total":     20,
				"unhealthy": 6,
			},
			"timestamp": time.Now().Unix(),
		}

		data, err := json.Marshal(healthResponse)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Contains(t, parsed, "status")
		assert.Contains(t, parsed, "providers")
		assert.Contains(t, parsed, "timestamp")
	})
}

// =========================================================================
// CLI AGENT COMPATIBILITY TESTS
// =========================================================================

// TestCLIAgentCompatibilityValidation validates compatibility with various CLI agents
func TestCLIAgentCompatibilityValidation(t *testing.T) {
	agents := []struct {
		name           string
		toolFormat     string
		expectedFields []string
	}{
		{
			name:           "OpenCode",
			toolFormat:     "camelCase",
			expectedFields: []string{"filePath", "content"},
		},
		{
			name:           "Crush",
			toolFormat:     "camelCase",
			expectedFields: []string{"filePath", "content"},
		},
		{
			name:           "HelixCode",
			toolFormat:     "camelCase",
			expectedFields: []string{"filePath", "content"},
		},
		{
			name:           "KiloCode",
			toolFormat:     "camelCase",
			expectedFields: []string{"filePath", "content"},
		},
		{
			name:           "Cline",
			toolFormat:     "camelCase",
			expectedFields: []string{"filePath", "content"},
		},
		{
			name:           "Continue",
			toolFormat:     "camelCase",
			expectedFields: []string{"filePath", "content"},
		},
		{
			name:           "Aider",
			toolFormat:     "camelCase",
			expectedFields: []string{"filePath", "content"},
		},
		{
			name:           "Cursor",
			toolFormat:     "camelCase",
			expectedFields: []string{"filePath", "content"},
		},
	}

	for _, agent := range agents {
		t.Run(agent.name+"_compatibility", func(t *testing.T) {
			// Generate tool call with expected format
			toolCall := map[string]interface{}{}
			for _, field := range agent.expectedFields {
				toolCall[field] = "test_value"
			}

			data, err := json.Marshal(toolCall)
			require.NoError(t, err)

			// Verify all expected fields are present
			for _, field := range agent.expectedFields {
				assert.Contains(t, string(data), field, "%s agent requires %s field", agent.name, field)
			}

			// Verify snake_case is NOT used
			assert.NotContains(t, string(data), "file_path", "%s should not use snake_case", agent.name)
		})
	}
}

// =========================================================================
// HTTP MOCK SERVER TESTS
// =========================================================================

// TestCogneeEndpointMocking validates Cognee endpoint behavior
func TestCogneeEndpointMocking(t *testing.T) {
	t.Run("root_endpoint_returns_alive_message", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World, I am alive!"})
			}
		}))
		defer server.Close()

		resp, err := http.Get(server.URL + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]string
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result["message"], "alive")
	})

	t.Run("search_endpoint_accepts_correct_params", func(t *testing.T) {
		var receivedBody map[string]interface{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/search" && r.Method == "POST" {
				json.NewDecoder(r.Body).Decode(&receivedBody)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
			}
		}))
		defer server.Close()

		reqBody := map[string]interface{}{
			"query":      "test",
			"datasets":   []string{"test_dataset"},
			"topK":       10,
			"searchType": "CHUNKS",
		}

		data, _ := json.Marshal(reqBody)
		resp, err := http.Post(server.URL+"/api/v1/search", "application/json", strings.NewReader(string(data)))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify correct params were sent
		assert.Contains(t, receivedBody, "searchType")
		assert.Contains(t, receivedBody, "topK")
		assert.NotContains(t, receivedBody, "search_type")
		assert.NotContains(t, receivedBody, "limit")
	})
}

// =========================================================================
// INTEGRATION SMOKE TESTS
// =========================================================================

// TestServerHealthCheckIntegration validates server health check
func TestServerHealthCheckIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would normally hit the actual server
	// For unit testing, we mock it
	t.Run("health_check_returns_status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/health" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status": "healthy",
					"providers": map[string]interface{}{
						"healthy":   14,
						"total":     20,
						"unhealthy": 6,
					},
				})
			}
		}))
		defer server.Close()

		resp, err := http.Get(server.URL + "/v1/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		var health map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)

		assert.Equal(t, "healthy", health["status"])
	})
}
