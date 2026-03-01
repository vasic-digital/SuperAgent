# ACP Manager - Agent Communication Protocol

**Location:** `docker/acp/`  
**Language:** Go  
**Purpose:** Central hub for agent discovery, registration, and message routing

## Overview

The ACP Manager implements the **Agent Communication Protocol**, enabling seamless communication between different CLI agents (Claude Code, OpenCode, Cline, etc.) and HelixAgent services. It provides a central registry where agents can register their capabilities and exchange messages.

## Features

- **Agent Registry**: Central database of all available agents
- **Pre-registered Agents**: Built-in knowledge of 11 popular CLI agents
- **Capability Discovery**: Query agents by their capabilities
- **Message Routing**: Route messages between agents
- **Health Monitoring**: Track agent availability
- **REST API**: HTTP interface for integration

## Architecture

### Components

```
┌─────────────────────────────────────┐
│         ACP Manager                 │
│  ┌─────────────────────────────┐   │
│  │   Agent Registry            │   │
│  │   (in-memory)               │   │
│  │                             │   │
│  │  ┌────────┐ ┌────────┐     │   │
│  │  │Claude  │ │OpenCode│     │   │
│  │  │ Code   │ │        │     │   │
│  │  └────────┘ └────────┘     │   │
│  └─────────────────────────────┘   │
│  ┌─────────────────────────────┐   │
│  │   HTTP API Server           │   │
│  │   (port 8766)               │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
```

### Pre-registered CLI Agents

The service comes pre-configured with these agents:

| Agent ID | Name | Type | Capabilities |
|----------|------|------|--------------|
| `claude-code` | Claude Code | cli-agent | code, chat, tools |
| `opencode` | OpenCode | cli-agent | code, chat, mcp |
| `cline` | Cline | cli-agent | code, chat, hooks |
| `kilo-code` | Kilo Code | cli-agent | code, chat, multi-platform |
| `aider` | Aider | cli-agent | code, chat, git |
| `goose` | Codename Goose | cli-agent | code, chat, mcp |
| `amazon-q` | Amazon Q | cli-agent | code, chat, aws |
| `kiro` | Kiro | cli-agent | code, chat |
| `gemini-cli` | Gemini CLI | cli-agent | code, chat |
| `deepseek-cli` | DeepSeek CLI | cli-agent | code, chat |

Plus HelixAgent internal agents:

| Agent ID | Name | Type | Capabilities |
|----------|------|------|--------------|
| `helixagent-debate` | HelixAgent Debate | internal | debate, ensemble |
| `helixagent-rag` | HelixAgent RAG | internal | rag, retrieval |
| `helixagent-memory` | HelixAgent Memory | internal | memory, storage |

## Usage

### Running the Service

```bash
# Build the container
docker build -t helixagent/acp-manager docker/acp/

# Run the service
docker run -d \
  -p 8766:8766 \
  -e PORT=8766 \
  helixagent/acp-manager
```

### Docker Compose

```yaml
version: '3.8'
services:
  acp-manager:
    build: docker/acp/
    ports:
      - "8766:8766"
    environment:
      - PORT=8766
    networks:
      - helixagent-network
```

## API Reference

### Health Check

```bash
curl http://localhost:8766/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2026-03-01T15:20:00Z",
  "agents": 11
}
```

### List All Agents

```bash
curl http://localhost:8766/agents
```

Response:
```json
{
  "agents": [
    {
      "id": "claude-code",
      "name": "Claude Code",
      "type": "cli-agent",
      "capabilities": ["code", "chat", "tools"],
      "endpoint": "stdio",
      "status": "registered",
      "last_seen": "2026-03-01T15:20:00Z"
    }
  ]
}
```

### Get Agent by ID

```bash
curl http://localhost:8766/agents/claude-code
```

### Register New Agent

```bash
curl -X POST http://localhost:8766/agents \
  -H "Content-Type: application/json" \
  -d '{
    "id": "my-custom-agent",
    "name": "My Custom Agent",
    "type": "cli-agent",
    "capabilities": ["code", "analysis"],
    "endpoint": "stdio"
  }'
```

### Find Agents by Capability

```bash
curl "http://localhost:8766/agents?capability=code"
```

### Send Message to Agent

```bash
curl -X POST http://localhost:8766/messages \
  -H "Content-Type: application/json" \
  -d '{
    "from": "helixagent-debate",
    "to": "claude-code",
    "type": "request",
    "payload": {
      "action": "analyze_code",
      "code": "func main() {}"
    }
  }'
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8766` | HTTP server port |
| `HEARTBEAT_INTERVAL` | `30s` | Agent heartbeat timeout |
| `MAX_AGENTS` | `100` | Maximum registered agents |

## Data Model

### Agent

```go
type Agent struct {
    ID           string            // Unique identifier
    Name         string            // Human-readable name
    Type         string            // Agent type (cli-agent, internal)
    Capabilities []string          // List of capabilities
    Endpoint     string            // Communication endpoint
    Status       string            // Current status
    LastSeen     time.Time         // Last heartbeat timestamp
    Metadata     map[string]string // Additional metadata
}
```

### Message

```go
type Message struct {
    ID        string                 // Unique message ID
    From      string                 // Source agent ID
    To        string                 // Destination agent ID
    Type      string                 // Message type (request, response, event)
    Payload   map[string]interface{} // Message payload
    Timestamp time.Time              // Message timestamp
}
```

## Integration

### With HelixAgent

HelixAgent uses the ACP Manager to:

1. **Discover CLI Agents**: Find which agents are available
2. **Route Requests**: Send tasks to appropriate agents
3. **Coordinate Workflows**: Orchestrate multi-agent processes
4. **Monitor Health**: Track agent availability

### Example: Agent Discovery

```go
import (
    "encoding/json"
    "net/http"
)

type AgentRegistry struct {
    baseURL string
}

func (r *AgentRegistry) GetAgentsByCapability(capability string) ([]Agent, error) {
    resp, err := http.Get(r.baseURL + "/agents?capability=" + capability)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Agents []Agent `json:"agents"`
    }
    err = json.NewDecoder(resp.Body).Decode(&result)
    return result.Agents, err
}

// Usage
registry := &AgentRegistry{baseURL: "http://acp-manager:8766"}
codeAgents, _ := registry.GetAgentsByCapability("code")
```

### Example: Sending Messages

```go
func (r *AgentRegistry) SendMessage(from, to string, payload map[string]any) error {
    message := map[string]any{
        "from":    from,
        "to":      to,
        "type":    "request",
        "payload": payload,
    }
    
    data, _ := json.Marshal(message)
    _, err := http.Post(
        r.baseURL+"/messages",
        "application/json",
        bytes.NewBuffer(data),
    )
    return err
}

// Usage
registry.SendMessage(
    "helixagent-debate",
    "claude-code",
    map[string]any{
        "action": "review_code",
        "code":   sourceCode,
    },
)
```

## Development

### Building

```bash
cd docker/acp/
go build -o acp-manager .
```

### Testing

```bash
go test ./...
```

### Running Locally

```bash
go run main.go
```

The service will start on port 8766 with pre-registered agents.

## Monitoring

### Health Endpoint

```bash
curl http://localhost:8766/health
```

### Metrics

Future enhancements:
- Prometheus metrics endpoint
- Agent activity tracking
- Message throughput
- Error rates

## Troubleshooting

### Agents Not Found

If agents don't appear in queries:
- Check if service is running: `curl http://localhost:8766/health`
- Verify agent is registered: `curl http://localhost:8766/agents`
- Check capability spelling (case-sensitive)

### Message Delivery Failing

If messages aren't reaching agents:
- Verify source agent is registered
- Verify destination agent exists
- Check agent status is "active"
- Review service logs for errors

## Future Enhancements

- **WebSocket Support**: Real-time bidirectional communication
- **Message Queue**: Persistent message storage
- **Authentication**: API key or mTLS authentication
- **Rate Limiting**: Prevent message flooding
- **Agent Discovery**: Automatic agent detection
- **Metrics Dashboard**: Visual monitoring

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## License

MIT License
