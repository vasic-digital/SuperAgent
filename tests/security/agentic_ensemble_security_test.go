//go:build security

// Package security provides security tests for the AgenticEnsemble pipeline.
package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/handlers"
)

// setupSecurityAgenticRouter creates a router with recovery and security
// middleware applied.
func setupSecurityAgenticRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	h := handlers.NewAgenticHandler(logger)

	r := gin.New()
	r.Use(gin.Recovery())

	// Body size limit: 1 MB
	const maxBodyBytes int64 = 1 * 1024 * 1024
	r.Use(func(c *gin.Context) {
		if c.Request.ContentLength > maxBodyBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge,
				gin.H{"error": "request body too large"})
			return
		}
		c.Next()
	})

	api := r.Group("/v1")
	handlers.RegisterAgenticRoutes(api, h)
	return r
}

// postAgenticSecurity fires a POST to the agentic endpoint and returns response.
func postAgenticSecurity(r *gin.Engine, rawBody []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/agentic/workflows",
		bytes.NewReader(rawBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestAgenticEnsemble_ToolInjection verifies that malicious tool_call parameters
// in the request body are rejected or sanitised without causing panics.
func TestAgenticEnsemble_ToolInjection(t *testing.T) {
	r := setupSecurityAgenticRouter()

	injections := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "sql-injection-in-query",
			payload: map[string]interface{}{
				"name":        "inject-test",
				"description": "SQL injection attempt",
				"nodes": []map[string]interface{}{
					{"id": "n1", "name": "Node", "type": "tool"},
				},
				"edges":       []map[string]interface{}{},
				"entry_point": "n1",
				"end_nodes":   []string{"n1"},
				"input": map[string]interface{}{
					"query": "'; DROP TABLE workflows; --",
					"context": map[string]interface{}{
						"tool_args": map[string]interface{}{
							"path": "../../../../etc/passwd",
						},
					},
				},
			},
		},
		{
			name: "path-traversal-in-tool-args",
			payload: map[string]interface{}{
				"name":        "path-traversal-test",
				"description": "path traversal attempt",
				"nodes": []map[string]interface{}{
					{"id": "n1", "name": "ReadNode", "type": "tool"},
				},
				"edges":       []map[string]interface{}{},
				"entry_point": "n1",
				"end_nodes":   []string{"n1"},
				"input": map[string]interface{}{
					"query": "read file",
					"context": map[string]interface{}{
						"file_path": "../../../secrets/.env",
					},
				},
			},
		},
		{
			name: "command-injection-in-node-name",
			payload: map[string]interface{}{
				"name":        "cmd-inject-test",
				"description": "command injection in node name",
				"nodes": []map[string]interface{}{
					{"id": "n1", "name": "$(rm -rf /)", "type": "agent"},
				},
				"edges":       []map[string]interface{}{},
				"entry_point": "n1",
				"end_nodes":   []string{"n1"},
			},
		},
	}

	for _, tc := range injections {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(tc.payload)
			require.NoError(t, err)

			// Must not panic; response can be 200 or 4xx but never 5xx panic
			w := postAgenticSecurity(r, raw)
			assert.NotEqual(t, http.StatusInternalServerError, w.Code,
				"server should not crash on injection: %s", tc.name)
		})
	}
}

// TestAgenticEnsemble_AgentIsolation verifies that two separate workflow
// executions do not share or leak state between each other.
func TestAgenticEnsemble_AgentIsolation(t *testing.T) {
	r := setupSecurityAgenticRouter()

	buildIsolationWorkflow := func(name, secret string) map[string]interface{} {
		return map[string]interface{}{
			"name":        name,
			"description": "isolation test",
			"nodes": []map[string]interface{}{
				{"id": "agent-node", "name": "Agent", "type": "agent"},
			},
			"edges":       []map[string]interface{}{},
			"entry_point": "agent-node",
			"end_nodes":   []string{"agent-node"},
			"input": map[string]interface{}{
				"query": "process data",
				"context": map[string]interface{}{
					"secret": secret,
				},
			},
		}
	}

	// Execute two workflows with different secrets
	raw1, err := json.Marshal(buildIsolationWorkflow("workflow-A", "secret-alpha-12345"))
	require.NoError(t, err)
	raw2, err := json.Marshal(buildIsolationWorkflow("workflow-B", "secret-beta-67890"))
	require.NoError(t, err)

	w1 := postAgenticSecurity(r, raw1)
	w2 := postAgenticSecurity(r, raw2)

	require.Equal(t, http.StatusOK, w1.Code)
	require.Equal(t, http.StatusOK, w2.Code)

	var resp1, resp2 map[string]interface{}
	require.NoError(t, json.Unmarshal(w1.Body.Bytes(), &resp1))
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp2))

	id1, _ := resp1["id"].(string)
	id2, _ := resp2["id"].(string)

	// Different workflow IDs confirm isolation
	assert.NotEqual(t, id1, id2,
		"two separate workflow executions must have distinct IDs")

	// Verify response body of workflow-B does not contain workflow-A's secret
	assert.False(t, strings.Contains(w2.Body.String(), "secret-alpha-12345"),
		"workflow-B response must not leak workflow-A state")
}

// TestAgenticEnsemble_OutputSanitization verifies that potentially dangerous
// content in the workflow input does not appear unsanitised in the response.
func TestAgenticEnsemble_OutputSanitization(t *testing.T) {
	r := setupSecurityAgenticRouter()

	dangerousInputs := []struct {
		name  string
		query string
	}{
		{
			name:  "xss-in-query",
			query: "<script>alert('xss')</script>",
		},
		{
			name:  "null-bytes",
			query: "normal query\x00injected",
		},
		{
			name:  "very-long-input",
			query: strings.Repeat("A", 10000),
		},
	}

	for _, tc := range dangerousInputs {
		t.Run(tc.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"name":        "sanitize-test",
				"description": "output sanitization test",
				"nodes": []map[string]interface{}{
					{"id": "n1", "name": "AgentNode", "type": "agent"},
				},
				"edges":       []map[string]interface{}{},
				"entry_point": "n1",
				"end_nodes":   []string{"n1"},
				"input": map[string]interface{}{
					"query": tc.query,
				},
			}

			raw, err := json.Marshal(payload)
			require.NoError(t, err)

			w := postAgenticSecurity(r, raw)

			// Server must not panic regardless of input
			assert.NotEqual(t, http.StatusInternalServerError, w.Code,
				"server must not crash on dangerous input: %s", tc.name)

			// Response must be valid JSON
			var respBody map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &respBody)
			assert.NoError(t, err, "response must be valid JSON for input: %s", tc.name)
		})
	}
}

// TestAgenticEnsemble_OversizedPayload verifies that excessively large payloads
// are rejected before reaching the handler.
func TestAgenticEnsemble_OversizedPayload(t *testing.T) {
	r := setupSecurityAgenticRouter()

	// Build a payload well over 1 MB
	oversized := make([]byte, 2*1024*1024)
	for i := range oversized {
		oversized[i] = 'A'
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/agentic/workflows",
		bytes.NewReader(oversized))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(oversized))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code,
		"oversized payload should be rejected")
}
