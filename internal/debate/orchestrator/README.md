# Debate Orchestrator Package

This package provides the main orchestration logic for the AI Debate Orchestrator Framework.

## Overview

The orchestrator coordinates all aspects of a debate: agent management, protocol execution, topology configuration, voting, and knowledge integration.

## Components

### Orchestrator (`orchestrator.go`)

Main coordination engine:

```go
orch := orchestrator.New(orchestrator.Config{
    AgentFactory:   factory,
    ProtocolRunner: protocol,
    VotingStrategy: voting,
    KnowledgeRepo:  knowledge,
})

result, err := orch.RunDebate(ctx, debateRequest)
```

### Service Integration (`service_integration.go`)

Bridges the orchestrator with HelixAgent services:

```go
integration := orchestrator.NewServiceIntegration(orchestrator.IntegrationConfig{
    EnableNewFramework:      true,
    FallbackToLegacy:        true,
    EnableLearning:          true,
    MinAgentsForNewFramework: 3,
})
```

### Adapter (`adapter.go`)

Adapts debate requests between formats:

```go
adapter := orchestrator.NewAdapter()
internalRequest := adapter.ToInternal(apiRequest)
apiResponse := adapter.ToExternal(internalResult)
```

### API Adapter (`api_adapter.go`)

HTTP API integration:

```go
handler := orchestrator.NewAPIAdapter(orch)
router.POST("/v1/debates", handler.HandleDebate)
```

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Orchestrator                      │
│  ┌───────────────────────────────────────────────┐ │
│  │                Coordination                    │ │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────────────┐ │ │
│  │  │ Agents  │ │Protocol │ │    Knowledge    │ │ │
│  │  └────┬────┘ └────┬────┘ └────────┬────────┘ │ │
│  │       └───────────┼───────────────┘          │ │
│  │                   ▼                          │ │
│  │  ┌─────────────────────────────────────┐    │ │
│  │  │         Execution Engine           │    │ │
│  │  └─────────────────────────────────────┘    │ │
│  └───────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────┐ │
│  │            Service Integration                │ │
│  │  Legacy Services ←→ New Framework             │ │
│  └───────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

## Configuration

```go
config := orchestrator.DefaultConfig()
config.EnableNewFramework = true       // Use new debate framework
config.FallbackToLegacy = true         // Fall back on failure
config.EnableLearning = true           // Enable cross-debate learning
config.MinAgentsForNewFramework = 3    // Minimum agents required
```

## Usage

```go
import "dev.helix.agent/internal/debate/orchestrator"

// Create orchestrator
orch := orchestrator.New(orchestrator.Config{
    AgentFactory:   agents.NewAgentFactory(providerRegistry),
    ProtocolRunner: protocol.NewRunner(protocol.DefaultConfig()),
    VotingStrategy: voting.NewWeightedVoting(voting.DefaultConfig()),
    KnowledgeRepo:  knowledge.NewRepository(knowledge.DefaultConfig()),
})

// Run debate
result, err := orch.RunDebate(ctx, orchestrator.DebateRequest{
    Topic:       "AI Safety Best Practices",
    Topology:    topology.TypeMesh,
    Participants: []string{"claude", "gemini", "deepseek"},
    Rounds:      3,
})
```

## Fallback Behavior

When the new framework fails or isn't available, the orchestrator falls back to legacy debate services:

1. Check if new framework requirements are met
2. Attempt debate with new framework
3. On failure, fall back to legacy services
4. Return result with fallback indicator

## Testing

```bash
go test -v ./internal/debate/orchestrator/...
```

## Files

- `orchestrator.go` - Main orchestration logic
- `service_integration.go` - Service bridge
- `adapter.go` - Request/response adaptation
- `adapter_test.go` - Adapter tests
- `api_adapter.go` - HTTP API integration
