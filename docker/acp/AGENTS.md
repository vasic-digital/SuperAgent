# AGENTS.md - ACP Manager

## Module Purpose for AI Agents

This is a **standalone HTTP service** that manages agent registration and communication for HelixAgent. It maintains a registry of available CLI agents (Claude Code, OpenCode, etc.) and routes messages between them.

## Agent Guidelines

### DO

✅ **Add new CLI agents** to the pre-registration list  
✅ **Add new endpoints** for agent management  
✅ **Improve filtering** for agent discovery  
✅ **Add tests** for HTTP handlers  
✅ **Document agent capabilities** clearly  

### DON'T

❌ **Add complex business logic** - keep it simple  
❌ **Store sensitive data** - no secrets or keys  
❌ **Break API compatibility** - clients depend on current endpoints  
❌ **Add persistent storage** without discussion  
❌ **Import HelixAgent internal packages** - standalone service  

## Common Tasks

### Adding a New Pre-registered Agent

Add to the `preregisteredAgents` slice in `main.go`:

```go
var preregisteredAgents = []Agent{
    // ... existing agents ...
    {
        ID:           "new-agent-id",
        Name:         "New Agent Name",
        Type:         "cli-agent",
        Capabilities: []string{"code", "chat", "new-capability"},
        Endpoint:     "stdio",
    },
}
```

Guidelines:
- Use lowercase kebab-case for ID
- Choose descriptive capabilities
- Type should be "cli-agent" for CLI tools, "internal" for HelixAgent services

### Adding a New Endpoint

Example: Add `/agents/{id}/ping`:

```go
func handleAgentPing(w http.ResponseWriter, r *http.Request) {
    // Extract agent ID from URL
    agentID := strings.TrimPrefix(r.URL.Path, "/agents/")
    agentID = strings.TrimSuffix(agentID, "/ping")
    
    // Get agent
    registry.mu.RLock()
    agent, exists := registry.agents[agentID]
    registry.mu.RUnlock()
    
    if !exists {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    
    // Ping logic here
    // ...
    
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
        "agent":  agent.Name,
    })
}

// Register in main()
http.HandleFunc("/agents/", func(w http.ResponseWriter, r *http.Request) {
    if strings.HasSuffix(r.URL.Path, "/ping") {
        handleAgentPing(w, r)
    } else {
        handleAgentDetail(w, r)
    }
})
```

### Implementing Capability Filtering

Current filtering is basic. To improve:

```go
func filterAgentsByCapabilities(capabilities []string) []Agent {
    var results []Agent
    
    registry.mu.RLock()
    defer registry.mu.RUnlock()
    
    for _, agent := range registry.agents {
        if hasAllCapabilities(agent, capabilities) {
            results = append(results, agent)
        }
    }
    
    return results
}

func hasAllCapabilities(agent Agent, required []string) bool {
    for _, req := range required {
        found := false
        for _, cap := range agent.Capabilities {
            if cap == req {
                found = true
                break
            }
        }
        if !found {
            return false
        }
    }
    return true
}
```

## Code Structure

### Main File Organization

```go
// main.go

// 1. Imports (lines 1-12)
// 2. Type definitions (lines 15-44)
//    - Agent struct
//    - Message struct
//    - AgentRegistry struct
// 3. Global registry (lines 42-44)
// 4. Pre-registered agents (lines 47-62)
// 5. Init function (lines 64-70)
// 6. HTTP handlers (lines 72+)
//    - handleHealth
//    - handleAgents
//    - handleAgentDetail
//    - handleMessages
// 7. Main function (lines ~150+)
```

### Key Data Structures

**Agent:**
```go
type Agent struct {
    ID           string            // Unique ID (e.g., "claude-code")
    Name         string            // Display name
    Type         string            // "cli-agent" or "internal"
    Capabilities []string          // What it can do
    Endpoint     string            // How to reach it
    Status       string            // "registered", "active", etc.
    LastSeen     time.Time         // For health checking
    Metadata     map[string]string // Extra info
}
```

**Message:**
```go
type Message struct {
    ID        string                 // Unique message ID
    From      string                 // Source agent ID
    To        string                 // Target agent ID
    Type      string                 // "request", "response", "event"
    Payload   map[string]interface{} // Message data
    Timestamp time.Time              // When sent
}
```

**AgentRegistry:**
```go
type AgentRegistry struct {
    agents map[string]Agent  // Map of agent ID to Agent
    mu     sync.RWMutex      // Thread-safe access
}
```

## Testing Patterns

### Handler Test

```go
func TestHandleAgents(t *testing.T) {
    // Setup
    req := httptest.NewRequest("GET", "/agents", nil)
    w := httptest.NewRecorder()
    
    // Execute
    handleAgents(w, req)
    
    // Assert
    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }
    
    var result struct {
        Agents []Agent `json:"agents"`
    }
    json.Unmarshal(w.Body.Bytes(), &result)
    
    if len(result.Agents) == 0 {
        t.Error("Expected agents, got none")
    }
}
```

### Registry Test

```go
func TestAgentRegistry(t *testing.T) {
    registry := &AgentRegistry{
        agents: make(map[string]Agent),
    }
    
    // Test registration
    agent := Agent{ID: "test", Name: "Test Agent"}
    registry.mu.Lock()
    registry.agents["test"] = agent
    registry.mu.Unlock()
    
    // Test retrieval
    registry.mu.RLock()
    retrieved, exists := registry.agents["test"]
    registry.mu.RUnlock()
    
    if !exists {
        t.Error("Agent should exist")
    }
    if retrieved.Name != "Test Agent" {
        t.Errorf("Name mismatch: %s", retrieved.Name)
    }
}
```

## Debugging Tips

### Agents Not Appearing

1. Check if service is running:
```bash
curl http://localhost:8766/health
```

2. Check agent list:
```bash
curl http://localhost:8766/agents | jq
```

3. Check specific agent:
```bash
curl http://localhost:8766/agents/claude-code | jq
```

4. Check logs:
```bash
docker logs helixagent-acp-manager
```

### Message Routing Issues

1. Verify source agent exists:
```bash
curl http://localhost:8766/agents/{from-agent-id}
```

2. Verify destination agent exists:
```bash
curl http://localhost:8766/agents/{to-agent-id}
```

3. Check message format:
```bash
curl -X POST http://localhost:8766/messages \
  -H "Content-Type: application/json" \
  -d '{
    "from": "valid-agent-id",
    "to": "valid-agent-id",
    "type": "request",
    "payload": {}
  }'
```

## Configuration Reference

### Environment Variables

```go
PORT                    // HTTP port (default: 8766)
HEARTBEAT_INTERVAL      // Health check timeout (default: 30s)
MAX_AGENTS              // Max agents (default: 100)
```

### Docker

```dockerfile
ENV PORT=8766
EXPOSE 8766
```

## Integration with HelixAgent

### How HelixAgent Uses ACP

1. **Discovery**: Query `/agents` to find available agents
2. **Filtering**: Use query params to find agents by capability
3. **Messaging**: POST to `/messages` to send tasks
4. **Monitoring**: GET `/health` for service status

### Example: Finding Code Agents

```go
// In HelixAgent service
resp, _ := http.Get("http://acp-manager:8766/agents?capability=code")
var result struct {
    Agents []Agent `json:"agents"`
}
json.NewDecoder(resp.Body).Decode(&result)

for _, agent := range result.Agents {
    log.Printf("Found code agent: %s", agent.Name)
}
```

### Example: Sending Task to Agent

```go
message := map[string]any{
    "from": "helixagent-debate",
    "to":   "claude-code",
    "type": "request",
    "payload": map[string]any{
        "action": "review_code",
        "code":   sourceCode,
    },
}

data, _ := json.Marshal(message)
resp, _ := http.Post(
    "http://acp-manager:8766/messages",
    "application/json",
    bytes.NewBuffer(data),
)
```

## Performance Considerations

### Current Limitations

- In-memory storage (restarts clear data)
- No persistence
- Single instance (no clustering)

### Optimization Tips

- Keep agent count reasonable (< 100)
- Use capability filtering to reduce response size
- Cache agent list if querying frequently
- Consider read replicas if scale needed

## Troubleshooting Checklist

- [ ] Service is running and healthy
- [ ] Port 8766 is accessible
- [ ] Pre-registered agents loaded
- [ ] New agents can register
- [ ] Messages can be sent
- [ ] Filtering works correctly
- [ ] No errors in logs

## Questions?

- Check [README.md](README.md) for usage examples
- Review main.go for implementation
- Look at tests/ for test patterns
- See [docs/protocols/ACP.md](../../docs/protocols/ACP.md) for protocol spec
