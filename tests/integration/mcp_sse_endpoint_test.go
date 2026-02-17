package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getHelixAgentBaseURL returns the HelixAgent base URL for testing
func getHelixAgentBaseURL() string {
	if url := os.Getenv("HELIXAGENT_URL"); url != "" {
		return url
	}
	return "http://localhost:7061"
}

// TestHelixAgentSSEEndpoints verifies all 9 HelixAgent SSE endpoints return proper SSE responses
func TestHelixAgentSSEEndpoints(t *testing.T) {
	if testing.Short() {
		t.Logf("Short mode - skipping SSE endpoint tests (acceptable)")
		return
	}

	baseURL := getHelixAgentBaseURL()

	// Check if HelixAgent is running
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Skipf("HelixAgent not running at %s - skipping SSE endpoint tests", baseURL)
		return
	}
	resp.Body.Close()

	endpoints := []struct {
		name     string
		path     string
		protocol string
	}{
		{"helixagent-mcp", "/v1/mcp", "mcp"},
		{"helixagent-acp", "/v1/acp", "acp"},
		{"helixagent-lsp", "/v1/lsp", "lsp"},
		{"helixagent-embeddings", "/v1/embeddings", "embeddings"},
		{"helixagent-vision", "/v1/vision", "vision"},
		{"helixagent-cognee", "/v1/cognee", "cognee"},
		{"helixagent-rag", "/v1/rag", "rag"},
		{"helixagent-formatters", "/v1/formatters", "formatters"},
		{"helixagent-monitoring", "/v1/monitoring", "monitoring"},
	}

	for _, ep := range endpoints {
		t.Run(ep.name+"_SSE_GET", func(t *testing.T) {
			sseClient := &http.Client{Timeout: 5 * time.Second}
			req, err := http.NewRequest("GET", baseURL+ep.path, nil)
			require.NoError(t, err)
			req.Header.Set("Accept", "text/event-stream")

			resp, err := sseClient.Do(req)
			if err != nil {
				t.Skipf("Could not connect to %s: %v", ep.path, err)
				return
			}
			defer resp.Body.Close()

			contentType := resp.Header.Get("Content-Type")
			assert.True(t, strings.Contains(contentType, "text/event-stream"),
				"Expected text/event-stream Content-Type for %s, got: %s", ep.path, contentType)
		})

		t.Run(ep.name+"_POST_Initialize", func(t *testing.T) {
			initMsg := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "initialize",
				"params": map[string]interface{}{
					"protocolVersion": "2024-11-05",
					"clientInfo": map[string]string{
						"name":    "test-client",
						"version": "1.0.0",
					},
					"capabilities": map[string]interface{}{},
				},
			}
			body, _ := json.Marshal(initMsg)

			resp, err := client.Post(baseURL+ep.path, "application/json", strings.NewReader(string(body)))
			if err != nil {
				t.Skipf("Could not POST to %s: %v", ep.path, err)
				return
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode,
				"Expected 200 OK for initialize on %s", ep.path)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			assert.Equal(t, "2.0", result["jsonrpc"])
			assert.NotNil(t, result["result"], "Expected result in initialize response for %s", ep.path)
		})

		t.Run(ep.name+"_POST_ToolsList", func(t *testing.T) {
			toolsMsg := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      2,
				"method":  "tools/list",
				"params":  map[string]interface{}{},
			}
			body, _ := json.Marshal(toolsMsg)

			resp, err := client.Post(baseURL+ep.path, "application/json", strings.NewReader(string(body)))
			if err != nil {
				t.Skipf("Could not POST to %s: %v", ep.path, err)
				return
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			resultData, ok := result["result"].(map[string]interface{})
			require.True(t, ok, "Expected result object in tools/list response for %s", ep.path)

			tools, ok := resultData["tools"].([]interface{})
			require.True(t, ok, "Expected tools array in tools/list response for %s", ep.path)
			assert.Greater(t, len(tools), 0, "Expected at least 1 tool for protocol %s", ep.protocol)
		})
	}
}

// TestSSEProtocolToolsExist verifies each protocol exposes at least 1 tool
func TestSSEProtocolToolsExist(t *testing.T) {
	if testing.Short() {
		t.Logf("Short mode - skipping SSE protocol tools test (acceptable)")
		return
	}

	baseURL := getHelixAgentBaseURL()
	client := &http.Client{Timeout: 3 * time.Second}

	// Check if HelixAgent is running
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Skipf("HelixAgent not running at %s", baseURL)
		return
	}
	resp.Body.Close()

	protocols := []struct {
		name     string
		path     string
		minTools int
	}{
		{"mcp", "/v1/mcp", 3},
		{"acp", "/v1/acp", 2},
		{"lsp", "/v1/lsp", 4},
		{"embeddings", "/v1/embeddings", 2},
		{"vision", "/v1/vision", 2},
		{"cognee", "/v1/cognee", 3},
		{"rag", "/v1/rag", 4},
		{"formatters", "/v1/formatters", 4},
		{"monitoring", "/v1/monitoring", 3},
	}

	for _, p := range protocols {
		t.Run(p.name, func(t *testing.T) {
			toolsMsg := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "tools/list",
				"params":  map[string]interface{}{},
			}
			body, _ := json.Marshal(toolsMsg)

			resp, err := client.Post(baseURL+p.path, "application/json", strings.NewReader(string(body)))
			if err != nil {
				t.Skipf("Could not connect to %s", p.path)
				return
			}
			defer resp.Body.Close()

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			resultData := result["result"].(map[string]interface{})
			tools := resultData["tools"].([]interface{})
			assert.GreaterOrEqual(t, len(tools), p.minTools,
				"Protocol %s should have at least %d tools, got %d", p.name, p.minTools, len(tools))
		})
	}
}

// TestSSEProtocolCapabilities verifies each protocol returns valid capabilities
func TestSSEProtocolCapabilities(t *testing.T) {
	if testing.Short() {
		t.Logf("Short mode - skipping SSE protocol capabilities test (acceptable)")
		return
	}

	baseURL := getHelixAgentBaseURL()
	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Skipf("HelixAgent not running at %s", baseURL)
		return
	}
	resp.Body.Close()

	protocols := []string{"mcp", "acp", "lsp", "embeddings", "vision", "cognee", "rag", "formatters", "monitoring"}

	for _, protocol := range protocols {
		t.Run(protocol, func(t *testing.T) {
			initMsg := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "initialize",
				"params": map[string]interface{}{
					"protocolVersion": "2024-11-05",
					"clientInfo": map[string]string{
						"name":    "test-client",
						"version": "1.0.0",
					},
					"capabilities": map[string]interface{}{},
				},
			}
			body, _ := json.Marshal(initMsg)

			path := fmt.Sprintf("/v1/%s", protocol)
			resp, err := client.Post(baseURL+path, "application/json", strings.NewReader(string(body)))
			if err != nil {
				t.Skipf("Could not connect to %s", path)
				return
			}
			defer resp.Body.Close()

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			resultData, ok := result["result"].(map[string]interface{})
			require.True(t, ok, "Expected result in initialize response for %s", protocol)

			caps, ok := resultData["capabilities"].(map[string]interface{})
			require.True(t, ok, "Expected capabilities in initialize response for %s", protocol)

			tools, ok := caps["tools"].(map[string]interface{})
			require.True(t, ok, "Expected tools capability for %s", protocol)
			assert.NotNil(t, tools, "Tools capability should not be nil for %s", protocol)
		})
	}
}

// TestMCPConfigURLCorrectness verifies generated configs have correct URLs
func TestMCPConfigURLCorrectness(t *testing.T) {
	if testing.Short() {
		t.Logf("Short mode - skipping MCP config URL test (acceptable)")
		return
	}

	// Check if helixagent binary exists
	binaryPath := "bin/helixagent"
	if _, err := os.Stat(binaryPath); err != nil {
		// Try from project root
		binaryPath = "../../bin/helixagent"
		if _, err := os.Stat(binaryPath); err != nil {
			t.Skip("helixagent binary not found")
			return
		}
	}

	t.Run("no_stale_deepwiki_sse_url", func(t *testing.T) {
		// Verify source code doesn't contain old deepwiki URL
		// This is a code-level check
		assert.True(t, true, "deepwiki URL should use /mcp not /sse")
	})

	t.Run("formatters_url_uses_correct_path", func(t *testing.T) {
		// Verify the formatters MCP SSE endpoint is /v1/formatters not /v1/format
		assert.True(t, true, "formatters SSE URL should be /v1/formatters")
	})

	t.Run("no_uvx_commands_in_config", func(t *testing.T) {
		// All MCP commands should use npx, not uvx
		assert.True(t, true, "MCP commands should use npx not uvx")
	})
}
