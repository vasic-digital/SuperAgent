# Video Course 72: LLMOrchestrator Mastery

## Course Overview

**Duration:** 2.5 hours
**Level:** Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 07 (Advanced Providers), Course 69 (Concurrency Safety)

Master the LLMOrchestrator module (`digital.vasic.llmorchestrator`), a standalone Go library for managing headless CLI agents with hybrid pipe and file communication. Learn to configure the agent pool, implement CLI adapters, leverage circuit breakers and health monitoring, and harden the system for production use.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Design an agent pool with capability-based acquisition and blocking wait
2. Implement the hybrid pipe+file transport protocol
3. Build CLI adapters using the BaseAdapter pattern
4. Configure circuit breakers for fault-tolerant agent communication
5. Set up health monitoring for agent pool liveness
6. Apply security hardening: path traversal protection, API key masking, response limits

---

## Module 1: Agent Pool Architecture (30 min)

### Video 1.1: The AgentPool Interface (15 min)

**Topics:**
- `AgentPool` provides thread-safe agent lifecycle management
- Operations: Register, Acquire (blocking), Release, Available, HealthCheck, Shutdown
- `sync.Mutex` + `sync.Cond` for blocking acquisition with context cancellation
- Capability matching: `AgentRequirements` filter agents by features

**Core Interface:**
```go
type AgentPool interface {
    Register(agent Agent) error
    Acquire(ctx context.Context, requirements AgentRequirements) (Agent, error)
    Release(agent Agent)
    Available() []Agent
    HealthCheck(ctx context.Context) []HealthStatus
    Shutdown(ctx context.Context) error
}
```

### Video 1.2: Agent Lifecycle (15 min)

**Topics:**
- The `Agent` interface: ID, Name, Start, Stop, IsRunning, Health, Send, SendStream
- Agent states: registered, acquired, released, shutdown
- The `agentEntry` struct: tracks per-agent acquired/available status
- Error semantics: `ErrNoAvailableAgent`, `ErrPoolShutdown`, `ErrAgentAlreadyRegistered`

**Agent Lifecycle:**
```
Register --> Available --> Acquire --> In Use --> Release --> Available
                                                         --> Shutdown
```

---

## Module 2: Hybrid Protocol (25 min)

### Video 2.1: PipeTransport (JSON-lines) (10 min)

**Topics:**
- `PipeTransport` communicates with CLI processes via stdin/stdout
- JSON-lines protocol: one JSON object per line, newline delimited
- `PipeMessage` struct: type, content, metadata fields
- Timeout handling: context deadlines for send and receive operations

**Sequence:**
```
Caller --> AgentPool.Acquire --> Agent.Send --> PipeTransport.SendPrompt
    --> stdin JSON-line --> CLI Process --> stdout JSON-line --> PipeMessage
```

### Video 2.2: FileTransport (inbox/outbox/shared) (15 min)

**Topics:**
- `FileTransport` uses filesystem directories for asynchronous communication
- Three directories: inbox (agent reads), outbox (agent writes), shared (bidirectional)
- Polling mechanism with configurable interval and timeout
- Use cases: agents that cannot use pipe-based communication
- Path traversal protection: all paths validated against base directory

---

## Module 3: CLI Adapters (25 min)

### Video 3.1: BaseAdapter Pattern (10 min)

**Topics:**
- `BaseAdapter` provides shared process management for all CLI adapters
- Each concrete adapter only implements output parsing logic
- Process lifecycle: start subprocess, attach transport, parse responses, stop
- 5 built-in adapters: OpenCode, ClaudeCode, Gemini, Junie, QwenCode

**Adapter Hierarchy:**
```
Agent Interface
    |
    +-- BaseAdapter (shared process management)
          |-- OpenCodeAgent (OpenCode CLI parsing)
          |-- ClaudeCodeAgent (Claude CLI parsing)
          |-- GeminiAgent (Gemini CLI parsing)
          |-- JunieAgent (Junie CLI parsing)
          |-- QwenCodeAgent (Qwen CLI parsing)
```

### Video 3.2: Building a Custom Adapter (15 min)

**Topics:**
- Extending `BaseAdapter` with custom response parsing
- The `ResponseParser` package: JSON extraction, action parsing, issue detection
- Handling non-standard output formats from different CLI tools
- Testing adapters with mock `CommandRunner` implementations

**Custom Adapter Skeleton:**
```go
type MyAgent struct {
    *BaseAdapter
}

func (a *MyAgent) parseResponse(raw string) (*Response, error) {
    parsed := parser.ParseJSON(raw)
    if parsed == nil {
        parsed = parser.ParsePlainText(raw)
    }
    return &Response{Content: parsed.Content}, nil
}
```

---

## Module 4: Circuit Breaker Integration (20 min)

### Video 4.1: Circuit Breaker Design (10 min)

**Topics:**
- Circuit breaker protects against cascading failures from unresponsive agents
- Three states: closed (normal), open (failing, reject requests), half-open (testing)
- Threshold: 3 consecutive failures opens the circuit for 60 seconds
- `AllowRequest()` check before every agent send operation

### Video 4.2: Integration with Agent Send (10 min)

**Topics:**
- The send sequence: `AllowRequest()` --> `Send()` --> `RecordSuccess/Failure()`
- Open circuit returns `ErrCircuitOpen` immediately without contacting the agent
- Half-open state allows one test request to determine recovery
- Per-agent circuit breakers: each agent has its own failure counter

**Pattern:**
```go
func (a *Agent) Send(ctx context.Context, prompt string) (Response, error) {
    if !a.circuitBreaker.AllowRequest() {
        return Response{}, ErrCircuitOpen
    }
    resp, err := a.transport.SendPrompt(ctx, a.id, prompt, "")
    if err != nil {
        a.circuitBreaker.RecordFailure()
        return Response{}, err
    }
    a.circuitBreaker.RecordSuccess()
    return a.parseResponse(resp)
}
```

---

## Module 5: Health Monitoring (15 min)

### Video 5.1: HealthMonitor and HealthStatus (15 min)

**Topics:**
- `HealthMonitor` runs periodic health checks on all registered agents
- `HealthStatus` struct: agent ID, status (healthy/unhealthy/unknown), latency, timestamp
- Integration with the agent pool: unhealthy agents excluded from acquisition
- Prometheus-compatible metric export for monitoring dashboards
- Configurable check interval and timeout per agent

**Health Check Flow:**
```
HealthMonitor (ticker) --> for each Agent --> agent.Health(ctx)
    --> update HealthStatus map --> expose via HealthCheck() API
```

---

## Module 6: Security Hardening (15 min)

### Video 6.1: Production Security Patterns (15 min)

**Topics:**
- Path traversal protection: validate all file paths against allowed base directories
- Response length limits: truncate oversized agent responses to prevent memory exhaustion
- API key masking: redact sensitive tokens in logs and error messages
- Process isolation: each agent runs as a separate subprocess with limited permissions
- Input sanitization: validate prompts before sending to CLI processes

**Security Checklist:**
1. All file paths validated against base directory (no `../` escape)
2. Response size capped at configurable maximum (default 10 MB)
3. API keys masked in all log output: `sk-****` pattern
4. Subprocess environment variables filtered to allowlist
5. Prompts sanitized: strip shell injection characters

---

## Assessment

### Quiz (10 questions)

1. What synchronization primitives does AgentPool use for blocking acquisition?
2. How does PipeTransport encode messages for CLI communication?
3. What is the role of BaseAdapter in the adapter hierarchy?
4. When does the circuit breaker transition from open to half-open?
5. What are the three FileTransport directories and their purposes?
6. How does HealthMonitor exclude unhealthy agents from the pool?
7. What is the default circuit breaker threshold for consecutive failures?
8. How does path traversal protection work in FileTransport?
9. What does `ResponseParser` extract from raw CLI output?
10. Why does each agent have its own circuit breaker instance?

### Practical Assessment

Build a complete LLMOrchestrator deployment:
1. Register 3 agents with different capabilities
2. Implement a custom adapter for a hypothetical CLI tool
3. Configure circuit breakers and health monitoring
4. Write a test that triggers circuit breaker open and recovery
5. Demonstrate blocking acquisition when all agents are in use

Deliverables:
1. Agent pool configuration and registration code
2. Custom adapter implementation with response parsing
3. Test suite covering circuit breaker states and pool contention
4. Health monitoring output showing agent status over time

---

## Resources

- [LLMOrchestrator Architecture](../../LLMOrchestrator/ARCHITECTURE.md)
- [LLMOrchestrator CLAUDE.md](../../LLMOrchestrator/CLAUDE.md)
- [AgentPool Source](../../LLMOrchestrator/pkg/agent/pool.go)
- [Protocol Package](../../LLMOrchestrator/pkg/protocol/)
- [Adapter Package](../../LLMOrchestrator/pkg/adapter/)
- [Course 69: Concurrency Safety Patterns](course-69-concurrency-safety.md)
