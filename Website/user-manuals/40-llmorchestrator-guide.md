# User Manual 40: LLMOrchestrator Guide

## Overview

LLMOrchestrator (`digital.vasic.llmorchestrator`) manages headless CLI agents (OpenCode, Claude Code, Gemini, Junie, Qwen Code) with hybrid pipe and file communication. It provides a thread-safe agent pool, circuit breakers, and structured response parsing.

## Prerequisites

- Go 1.24+
- At least one supported CLI agent binary installed and on `$PATH`:
  - `opencode`, `claude`, `gemini`, `junie`, `qwen-code`
- Corresponding API keys configured in `.env`

## Step 1: Build LLMOrchestrator

```bash
cd LLMOrchestrator
go build ./...
```

Run the standalone orchestrator:

```bash
go run cmd/orchestrator/main.go
```

## Step 2: Configure Agent Paths and Keys

Copy `.env.example` to `.env` and configure:

```bash
# Agent binary paths (auto-detected on $PATH if omitted)
ORCHESTRATOR_OPENCODE_PATH=/usr/local/bin/opencode
ORCHESTRATOR_CLAUDE_PATH=/usr/local/bin/claude
ORCHESTRATOR_GEMINI_PATH=/usr/local/bin/gemini
ORCHESTRATOR_JUNIE_PATH=/usr/local/bin/junie
ORCHESTRATOR_QWEN_PATH=/usr/local/bin/qwen-code

# API keys for each agent
ANTHROPIC_API_KEY=sk-ant-...
GOOGLE_API_KEY=...
OPENAI_API_KEY=sk-...
```

## Step 3: Add CLI Agents to the Pool

The `AgentPool` manages available agents with thread-safe acquire/release semantics:

```go
pool := agent.NewAgentPool()
pool.Add(openCodeAdapter)
pool.Add(claudeCodeAdapter)
pool.Add(geminiAdapter)
```

Acquire an agent matching specific capabilities:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

a, err := pool.Acquire(ctx, agent.Requirements{
    Capabilities: []string{"code-generation", "tool-use"},
})
if err != nil {
    log.Fatal("no matching agent available:", err)
}
defer pool.Release(a)
```

The pool blocks until a matching agent becomes available or the context expires.

## Step 4: Configure Communication Protocols

LLMOrchestrator supports two communication modes:

### Pipe Protocol (Real-Time)

JSON-lines over stdin/stdout for interactive, low-latency exchanges:

```go
transport := protocol.NewPipeTransport(agentProcess)
resp, err := transport.Send(ctx, protocol.Message{
    Role:    "user",
    Content: "Analyze this code for bugs",
})
```

### File Protocol (Artifact Exchange)

Inbox/outbox directories for large payloads (code files, screenshots):

```go
transport := protocol.NewFileTransport(protocol.FileConfig{
    Inbox:  "/tmp/orchestrator/inbox",
    Outbox: "/tmp/orchestrator/outbox",
    Shared: "/tmp/orchestrator/shared",
})
transport.SendFile(ctx, "analysis-request.json", payload)
result, err := transport.WaitForResponse(ctx, "analysis-result.json")
```

## Step 5: Monitor Agent Health

Each agent has an integrated circuit breaker (3 consecutive failures triggers open state for 60 seconds):

```go
monitor := agent.NewHealthMonitor(pool)
status := monitor.Status()
for name, health := range status {
    fmt.Printf("%s: healthy=%v, failures=%d\n", name, health.Healthy, health.ConsecutiveFailures)
}
```

Health states: `healthy` (circuit closed), `degraded` (1-2 failures), `unhealthy` (circuit open).

## Step 6: Handle Failures and Parse Responses

The `ResponseParser` extracts structured data from raw LLM output:

```go
parser := parser.NewResponseParser()
result := parser.Parse(rawOutput)

for _, action := range result.Actions {
    fmt.Printf("Action: %s on %s\n", action.Type, action.Target)
}
for _, issue := range result.Issues {
    fmt.Printf("Issue: [%s] %s\n", issue.Severity, issue.Description)
}
```

When an agent fails, the pool automatically routes to another healthy agent with matching capabilities.

## Step 7: Security Considerations

LLMOrchestrator enforces security boundaries:

- **Path traversal protection**: File transport rejects paths outside configured directories
- **Response length limits**: Oversized responses are truncated to prevent memory exhaustion
- **API key masking**: Keys are never logged or included in error messages
- **Command injection prevention**: Agent arguments are sanitized before process spawn

## Step 8: Run Tests

```bash
make test      # All tests with race detector
make fuzz      # Fuzz tests for parser robustness
make cover     # Coverage report
make check     # vet + tests
```

## Package Reference

| Package | Purpose |
|---------|---------|
| `pkg/agent` | Agent interface, AgentPool, HealthMonitor, CircuitBreaker |
| `pkg/adapter` | BaseAdapter + 5 CLI adapters (opencode, claudecode, gemini, junie, qwencode) |
| `pkg/protocol` | PipeTransport (JSON-lines), FileTransport (inbox/outbox) |
| `pkg/parser` | ResponseParser, action/issue extraction |
| `pkg/config` | .env loading, agent path resolution |

## Troubleshooting

- **"no matching agent available"**: Ensure at least one agent binary is installed and its path is configured
- **Circuit breaker open**: Check agent logs; 3 consecutive failures trigger a 60-second cooldown
- **File transport timeout**: Verify inbox/outbox directories exist and have write permissions
- **Parse errors**: Raw LLM output may lack structure; check that the agent is running in JSON output mode

## Related Resources

- [User Manual 39: HelixQA Guide](39-helixqa-guide.md) -- Uses LLMOrchestrator for autonomous QA
- [User Manual 41: VisionEngine Guide](41-visionengine-guide.md) -- Vision analysis paired with orchestrated agents
- Source: `LLMOrchestrator/README.md`, `LLMOrchestrator/CLAUDE.md`
