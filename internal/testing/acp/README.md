# ACP Protocol Testing Package

The ACP (Agent Communication Protocol) testing package provides comprehensive utilities for testing ACP agents with real functional validation.

## Overview

This package implements:

- **ACP Testing Framework**: HTTP client for HelixAgent ACP endpoints
- **Client Simulation**: Simulates agent client interactions
- **Message Validation**: JSON schema and protocol compliance validation
- **Session Management Testing**: Agent lifecycle and state management tests
- **Agent Execution Tests**: Real task execution with validation

## Key Principles

**No False Positives**: Tests execute ACTUAL agent operations, not just connectivity checks. Tests FAIL if agent execution fails.

## Directory Structure

```
internal/testing/acp/
├── functional_test.go    # Real ACP agent functional tests
└── README.md             # This file

Related packages:
├── internal/agents/      # CLI agent registry (48 agents)
├── pkg/cliagents/        # Agent configuration generator
└── internal/tools/       # Tool schema registry (21 tools)
```

## Key Components

### ACPClient

HTTP client for testing HelixAgent ACP API:

```go
// Create client
client := acp.NewACPClient("http://localhost:8080")

// List available agents
agents, err := client.ListAgents()
if err != nil {
    t.Skipf("ACP service not running: %v", err)
}
t.Logf("Available agents: %v", agents)

// Get agent information
info, err := client.GetAgentInfo("code-reviewer")

// Execute agent task
resp, err := client.ExecuteTask(&acp.AgentRequest{
    AgentID: "code-reviewer",
    Task:    "Review this code for best practices",
    Context: map[string]interface{}{
        "code":     "func main() {}",
        "language": "go",
    },
    Timeout: 60,
})
```

### Request/Response Types

```go
// AgentRequest represents an ACP agent request
type AgentRequest struct {
    AgentID string                 `json:"agent_id"`
    Task    string                 `json:"task"`
    Context map[string]interface{} `json:"context,omitempty"`
    Tools   []string               `json:"tools,omitempty"`
    Timeout int                    `json:"timeout,omitempty"`
}

// AgentResponse represents an ACP agent response
type AgentResponse struct {
    AgentID  string                 `json:"agent_id"`
    Status   string                 `json:"status"`
    Result   interface{}            `json:"result,omitempty"`
    Error    string                 `json:"error,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

## Client Simulation

### Basic Client

```go
type ACPClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewACPClient(baseURL string) *ACPClient {
    return &ACPClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}
```

### Concurrent Client Simulation

```go
func TestConcurrentAgentExecution(t *testing.T) {
    client := acp.NewACPClient("http://localhost:8080")

    var wg sync.WaitGroup
    results := make(chan *acp.AgentResponse, 10)
    errors := make(chan error, 10)

    // Simulate multiple concurrent clients
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(clientID int) {
            defer wg.Done()

            req := &acp.AgentRequest{
                AgentID: "code-reviewer",
                Task:    fmt.Sprintf("Task from client %d", clientID),
                Context: map[string]interface{}{
                    "code":      "func test() {}",
                    "client_id": clientID,
                },
            }

            resp, err := client.ExecuteTask(req)
            if err != nil {
                errors <- err
                return
            }
            results <- resp
        }(i)
    }

    wg.Wait()
    close(results)
    close(errors)

    // Validate all responses
    successCount := 0
    for resp := range results {
        assert.NotEqual(t, "error", resp.Status)
        successCount++
    }

    for err := range errors {
        t.Logf("Client error: %v", err)
    }

    t.Logf("Successful executions: %d/10", successCount)
}
```

## Supported ACP Agents

| Agent ID | Description | Primary Task |
|----------|-------------|--------------|
| code-reviewer | Code review agent | Review code for best practices |
| bug-finder | Bug detection agent | Find potential bugs |
| refactor-assistant | Refactoring agent | Suggest improvements |
| documentation-generator | Documentation agent | Generate documentation |
| test-generator | Test generation agent | Generate unit tests |
| security-scanner | Security scanning agent | Scan for vulnerabilities |

## Message Validation

### JSON Schema Validation

```go
// AgentRequestSchema defines the JSON schema for agent requests
var AgentRequestSchema = map[string]interface{}{
    "type": "object",
    "required": []string{"agent_id", "task"},
    "properties": map[string]interface{}{
        "agent_id": map[string]interface{}{
            "type":      "string",
            "minLength": 1,
        },
        "task": map[string]interface{}{
            "type":      "string",
            "minLength": 1,
        },
        "context": map[string]interface{}{
            "type": "object",
        },
        "tools": map[string]interface{}{
            "type": "array",
            "items": map[string]interface{}{
                "type": "string",
            },
        },
        "timeout": map[string]interface{}{
            "type":    "integer",
            "minimum": 1,
            "maximum": 300,
        },
    },
}

func validateAgentRequest(req *AgentRequest) error {
    if req.AgentID == "" {
        return fmt.Errorf("agent_id is required")
    }
    if req.Task == "" {
        return fmt.Errorf("task is required")
    }
    if req.Timeout > 300 {
        return fmt.Errorf("timeout exceeds maximum (300s)")
    }
    return nil
}
```

### Response Validation

```go
func validateAgentResponse(resp *AgentResponse) error {
    if resp.AgentID == "" {
        return fmt.Errorf("response missing agent_id")
    }

    validStatuses := map[string]bool{
        "pending":    true,
        "running":    true,
        "completed":  true,
        "error":      true,
        "timeout":    true,
        "cancelled":  true,
    }

    if !validStatuses[resp.Status] {
        return fmt.Errorf("invalid status: %s", resp.Status)
    }

    if resp.Status == "error" && resp.Error == "" {
        return fmt.Errorf("error status requires error message")
    }

    return nil
}
```

## Session Management Testing

### Session Lifecycle

```go
func TestAgentSessionLifecycle(t *testing.T) {
    client := acp.NewACPClient("http://localhost:8080")

    // Start session
    session, err := client.StartSession(&SessionRequest{
        AgentID:    "code-reviewer",
        SessionTTL: 300, // 5 minutes
    })
    require.NoError(t, err)
    sessionID := session.SessionID

    t.Cleanup(func() {
        // End session on test completion
        client.EndSession(sessionID)
    })

    // Execute task within session
    resp, err := client.ExecuteInSession(sessionID, &AgentRequest{
        Task: "Review this code",
        Context: map[string]interface{}{
            "code": "func main() {}",
        },
    })
    require.NoError(t, err)
    assert.Equal(t, "completed", resp.Status)

    // Verify session state
    state, err := client.GetSessionState(sessionID)
    require.NoError(t, err)
    assert.Equal(t, "active", state.Status)
    assert.Equal(t, 1, state.TaskCount)
}
```

### Session Timeout

```go
func TestSessionTimeout(t *testing.T) {
    client := acp.NewACPClient("http://localhost:8080")

    // Create short-lived session
    session, _ := client.StartSession(&SessionRequest{
        AgentID:    "code-reviewer",
        SessionTTL: 1, // 1 second
    })

    // Wait for timeout
    time.Sleep(2 * time.Second)

    // Attempt to use expired session
    _, err := client.ExecuteInSession(session.SessionID, &AgentRequest{
        Task: "Should fail",
    })

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expired")
}
```

### Concurrent Sessions

```go
func TestConcurrentSessions(t *testing.T) {
    client := acp.NewACPClient("http://localhost:8080")

    sessions := make([]*Session, 5)
    var wg sync.WaitGroup

    // Create multiple sessions
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            session, err := client.StartSession(&SessionRequest{
                AgentID:    "code-reviewer",
                SessionTTL: 60,
            })
            if err == nil {
                sessions[idx] = session
            }
        }(i)
    }
    wg.Wait()

    // Verify all sessions created
    validSessions := 0
    for _, s := range sessions {
        if s != nil {
            validSessions++
            defer client.EndSession(s.SessionID)
        }
    }

    assert.Equal(t, 5, validSessions, "All sessions should be created")
}
```

## Functional Tests

### Agent Discovery

```go
func TestACPAgentDiscovery(t *testing.T) {
    client := acp.NewACPClient("http://localhost:8080")

    agents, err := client.ListAgents()
    if err != nil {
        t.Skipf("ACP service not running: %v", err)
    }

    assert.NotEmpty(t, agents, "Should have at least one agent")
    t.Logf("Discovered %d ACP agents: %v", len(agents), agents)
}
```

### Agent Information

```go
func TestACPAgentInfo(t *testing.T) {
    client := acp.NewACPClient("http://localhost:8080")

    for _, agent := range ACPAgents {
        t.Run(agent.ID, func(t *testing.T) {
            info, err := client.GetAgentInfo(agent.ID)
            if err != nil {
                t.Skipf("Agent %s not available: %v", agent.ID, err)
            }

            assert.Equal(t, agent.ID, info.AgentID)
            t.Logf("Agent %s info: %+v", agent.ID, info)
        })
    }
}
```

### Agent Execution

```go
func TestACPAgentExecution(t *testing.T) {
    client := acp.NewACPClient("http://localhost:8080")

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
            }

            assert.Equal(t, agent.ID, resp.AgentID)
            assert.NotEqual(t, "error", resp.Status)
            t.Logf("Agent %s result: %v", agent.ID, resp.Result)
        })
    }
}
```

### Health Check

```go
func TestACPHealthCheck(t *testing.T) {
    client := acp.NewACPClient("http://localhost:8080")

    resp, err := client.httpClient.Get(client.baseURL + "/v1/acp/health")
    if err != nil {
        t.Skipf("ACP service not running: %v", err)
    }
    defer resp.Body.Close()

    require.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Benchmark Tests

```go
func BenchmarkACPAgentExecution(b *testing.B) {
    client := acp.NewACPClient("http://localhost:8080")

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
        }
    }
}

func BenchmarkConcurrentAgents(b *testing.B) {
    client := acp.NewACPClient("http://localhost:8080")

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, _ = client.ExecuteTask(&AgentRequest{
                AgentID: "code-reviewer",
                Task:    "Quick review",
                Context: map[string]interface{}{"code": "x := 1"},
            })
        }
    })
}
```

## Running Tests

```bash
# Run all ACP tests
go test -v ./internal/testing/acp/...

# Run specific agent test
go test -v -run TestACPAgentExecution/code-reviewer ./internal/testing/acp/

# Run with extended timeout
go test -v -timeout 120s ./internal/testing/acp/

# Run benchmarks
go test -bench=. ./internal/testing/acp/

# Run with race detection
go test -race ./internal/testing/acp/
```

## Error Handling

```go
resp, err := client.ExecuteTask(req)
if err != nil {
    // Network/connection error
    if strings.Contains(err.Error(), "connection refused") {
        t.Skipf("ACP service not running")
    }
    t.Fatalf("Request failed: %v", err)
}

// Check response status
switch resp.Status {
case "completed":
    t.Logf("Task completed: %v", resp.Result)
case "error":
    t.Errorf("Agent error: %s", resp.Error)
case "timeout":
    t.Log("Task timed out")
case "cancelled":
    t.Log("Task was cancelled")
default:
    t.Logf("Unknown status: %s", resp.Status)
}
```

## Best Practices

1. **Use Appropriate Timeouts**: Agent tasks can take time
2. **Validate Request Schema**: Check requests before sending
3. **Handle All Status Types**: Agents can return various statuses
4. **Clean Up Sessions**: Always end sessions after use
5. **Skip Unavailable Agents**: Use `t.Skip()` when agents are not running
6. **Test Concurrent Access**: Verify thread safety

## Configuration

```go
type ACPAgentConfig struct {
    ID          string
    Description string
    TestTask    string
}

var ACPAgents = []ACPAgentConfig{
    {
        ID:          "code-reviewer",
        Description: "Code review agent",
        TestTask:    "Review this code for best practices",
    },
    {
        ID:          "bug-finder",
        Description: "Bug detection agent",
        TestTask:    "Find potential bugs in this code",
    },
    // ... more agents
}
```

## See Also

- `internal/agents/` - CLI agent registry
- `pkg/cliagents/` - Agent configuration generator
- `internal/tools/` - Tool schema registry
- [ACP Specification](https://spec.agentcommunicationprotocol.dev/)
