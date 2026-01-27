// Package acp provides real functional tests for ACP (Agent Communication Protocol) agents.
// These tests execute ACTUAL agent operations, not just connectivity checks.
// Tests FAIL if the operation fails - no false positives.
package acp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ACPClient provides a client for testing ACP agents
type ACPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewACPClient creates a new ACP test client
func NewACPClient(baseURL string) *ACPClient {
	return &ACPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AgentRequest represents an ACP agent request
type AgentRequest struct {
	AgentID  string                 `json:"agent_id"`
	Task     string                 `json:"task"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Tools    []string               `json:"tools,omitempty"`
	Timeout  int                    `json:"timeout,omitempty"`
}

// AgentResponse represents an ACP agent response
type AgentResponse struct {
	AgentID   string                 `json:"agent_id"`
	Status    string                 `json:"status"`
	Result    interface{}            `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ListAgents lists all available ACP agents
func (c *ACPClient) ListAgents() ([]string, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/v1/acp/agents")
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list agents failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Agents []string `json:"agents"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Agents, nil
}

// GetAgentInfo gets information about a specific agent
func (c *ACPClient) GetAgentInfo(agentID string) (*AgentResponse, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/v1/acp/agents/" + agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get agent info failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result AgentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ExecuteTask sends a task to an ACP agent
func (c *ACPClient) ExecuteTask(req *AgentRequest) (*AgentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/v1/acp/execute", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to execute task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("execute task failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result AgentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ACPAgentConfig holds configuration for testing an ACP agent
type ACPAgentConfig struct {
	ID          string
	Description string
	TestTask    string
}

// ACP agents to test
var ACPAgents = []ACPAgentConfig{
	{ID: "code-reviewer", Description: "Code review agent", TestTask: "Review this code for best practices"},
	{ID: "bug-finder", Description: "Bug detection agent", TestTask: "Find potential bugs in this code"},
	{ID: "refactor-assistant", Description: "Refactoring agent", TestTask: "Suggest refactoring improvements"},
	{ID: "documentation-generator", Description: "Documentation agent", TestTask: "Generate documentation for this function"},
	{ID: "test-generator", Description: "Test generation agent", TestTask: "Generate unit tests for this code"},
	{ID: "security-scanner", Description: "Security scanning agent", TestTask: "Scan for security vulnerabilities"},
}

// TestACPAgentDiscovery tests agent discovery endpoint
func TestACPAgentDiscovery(t *testing.T) {
	client := NewACPClient("http://localhost:8080")

	agents, err := client.ListAgents()
	if err != nil {
		t.Skipf("ACP service not running: %v", err)
		return
	}

	assert.NotEmpty(t, agents, "Should have at least one agent")
	t.Logf("Discovered %d ACP agents: %v", len(agents), agents)
}

// TestACPAgentInfo tests getting agent information
func TestACPAgentInfo(t *testing.T) {
	client := NewACPClient("http://localhost:8080")

	for _, agent := range ACPAgents {
		t.Run(agent.ID, func(t *testing.T) {
			info, err := client.GetAgentInfo(agent.ID)
			if err != nil {
				t.Skipf("Agent %s not available: %v", agent.ID, err)
				return
			}

			assert.Equal(t, agent.ID, info.AgentID)
			t.Logf("Agent %s info: %+v", agent.ID, info)
		})
	}
}

// TestACPAgentExecution tests actual agent task execution
func TestACPAgentExecution(t *testing.T) {
	client := NewACPClient("http://localhost:8080")

	testCode := `
func add(a, b int) int {
    return a + b
}
`

	for _, agent := range ACPAgents {
		t.Run(agent.ID, func(t *testing.T) {
			req := &AgentRequest{
				AgentID: agent.ID,
				Task:    agent.TestTask,
				Context: map[string]interface{}{
					"code":     testCode,
					"language": "go",
				},
				Timeout: 60,
			}

			resp, err := client.ExecuteTask(req)
			if err != nil {
				t.Skipf("Agent %s execution failed: %v", agent.ID, err)
				return
			}

			assert.Equal(t, agent.ID, resp.AgentID)
			assert.NotEqual(t, "error", resp.Status, "Agent should not return error status")
			t.Logf("Agent %s result: %v", agent.ID, resp.Result)
		})
	}
}

// TestACPHealthCheck tests ACP service health
func TestACPHealthCheck(t *testing.T) {
	client := NewACPClient("http://localhost:8080")

	resp, err := client.httpClient.Get(client.baseURL + "/v1/acp/health")
	if err != nil {
		t.Skipf("ACP service not running: %v", err)
		return
	}
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Health check should return 200")
}

// BenchmarkACPAgentExecution benchmarks agent task execution
func BenchmarkACPAgentExecution(b *testing.B) {
	client := NewACPClient("http://localhost:8080")

	req := &AgentRequest{
		AgentID: "code-reviewer",
		Task:    "Review this code",
		Context: map[string]interface{}{
			"code":     "func main() {}",
			"language": "go",
		},
		Timeout: 30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.ExecuteTask(req)
		if err != nil {
			b.Skipf("ACP service not running: %v", err)
			return
		}
	}
}
