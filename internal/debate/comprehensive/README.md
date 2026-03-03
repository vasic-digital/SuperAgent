# Comprehensive Debate System

The `comprehensive` package implements a full multi-agent debate system with 8 specialized agent roles, 6-phase debate workflow, and quality-gated convergence.

## Architecture

### Agent Roles (8)

| Role | Purpose |
|------|---------|
| Architect | System design and architecture proposals |
| Generator | Code implementation and generation |
| Critic | Code review and issue identification |
| Refactoring | Code improvement suggestions |
| Tester | Test case design and validation |
| Validator | Correctness verification |
| Security | Security analysis and vulnerability detection |
| Performance | Performance optimization and profiling |

### Debate Phases (6)

1. **Planning** - Architects propose system designs
2. **Generation** - Generators produce code implementations
3. **Debate** - Critics critique, generators defend (multi-round)
4. **Validation** - Testers create tests, validators verify correctness
5. **Refactoring** - Refactoring and performance agents optimize code
6. **Integration** - Security review and final validation

### Key Components

- **System** (`system.go`) - Main entry point; orchestrates all phases with convergence checking
- **PhaseOrchestrator** (`phases_orchestrator.go`) - Executes individual phases using specialized agents
- **IntegrationManager** (`integration.go`) - Manages component wiring, streaming, and quality gates
- **DebateEngine** (`engine.go`) - Core debate loop with round management
- **AgentPool** (`agents_pool.go`) - Agent lifecycle, selection, and team management
- **Specialized Agents** (`agents_specialized.go`) - Role-specific agent wrappers with Process() dispatch
- **StreamOrchestrator** (`streaming.go`) - Real-time SSE streaming of debate progress
- **ToolRegistry** (`tools.go`) - Code, search, build, test, security, and analysis tools
- **ConsensusResult** (`consensus.go`) - Consensus tracking and quality scoring
- **Analysis** (`analysis.go`) - Static code analysis for quality metrics

### Convergence Criteria

The system checks three conditions to end debate early:
1. Quality score exceeds threshold (default: 0.95)
2. Consensus level exceeds minimum (default: 0.8)
3. Early stopping: no improvement over N consecutive rounds (default: 3)

## Usage

```go
config := comprehensive.DefaultConfig()
system := comprehensive.NewSystem(config)

result, err := system.ConductDebate(ctx, &comprehensive.DebateRequest{
    ID:    "debate-1",
    Topic: "Implement a rate limiter",
})
```

## Streaming Usage

```go
mgr, _ := comprehensive.NewIntegrationManager(config, logger)
mgr.Initialize(ctx)

result, err := mgr.StreamDebate(ctx, &comprehensive.DebateStreamRequest{
    DebateRequest: req,
    Stream:        true,
    StreamHandler: mySSEHandler,
})
```
