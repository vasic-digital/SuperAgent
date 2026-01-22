# Debate Agents Package

This package provides the agent factory, pooling, and specialization systems for the AI Debate Orchestrator Framework.

## Overview

The agents package manages the lifecycle of debate participants, including creation, pooling, specialization, and configuration based on LLM provider capabilities.

## Components

### Agent Factory (`factory.go`)

Creates and manages debate agents with proper configuration:

```go
factory := agents.NewAgentFactory(providerRegistry)
agent, err := factory.CreateAgent(ctx, agents.AgentConfig{
    ID:           "agent-001",
    Provider:     "claude",
    Model:        "claude-sonnet-4-20250514",
    Specialization: agents.SpecializationAnalyst,
})
```

### Agent Pool

Manages a pool of reusable agents for efficient resource utilization:

```go
pool := agents.NewAgentPool(factory, poolConfig)
agent, err := pool.Acquire(ctx, requirements)
defer pool.Release(agent)
```

### Specialization (`specialization.go`)

Defines agent roles and their capabilities:

| Specialization | Role | Description |
|----------------|------|-------------|
| `Analyst` | Analysis | Deep analytical thinking |
| `Critic` | Critique | Critical evaluation |
| `Synthesizer` | Synthesis | Combining perspectives |
| `Devil's Advocate` | Challenge | Counter-arguments |
| `Moderator` | Facilitation | Guide discussion |

### Templates (`templates.go`)

Pre-configured agent templates for common scenarios:

```go
template := agents.GetTemplate(agents.TemplateDebateAnalyst)
agent := factory.CreateFromTemplate(ctx, template)
```

## Architecture

```
┌─────────────────────────────────────────────┐
│                Agent Factory                 │
│  ┌─────────┐ ┌────────────┐ ┌────────────┐ │
│  │ Create  │ │  Configure │ │ Specialize │ │
│  └────┬────┘ └─────┬──────┘ └─────┬──────┘ │
│       └────────────┼──────────────┘        │
│                    ▼                        │
│  ┌─────────────────────────────────────┐   │
│  │            Agent Pool               │   │
│  │  ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐    │   │
│  │  │ A │ │ B │ │ C │ │ D │ │ E │    │   │
│  │  └───┘ └───┘ └───┘ └───┘ └───┘    │   │
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

## Usage

```go
import "dev.helix.agent/internal/debate/agents"

// Create factory with provider registry
factory := agents.NewAgentFactory(providerRegistry)

// Create specialized agent
config := agents.AgentConfig{
    ID:             "analyst-001",
    Provider:       "gemini",
    Model:          "gemini-1.5-flash",
    Specialization: agents.SpecializationAnalyst,
    Temperature:    0.7,
}

agent, err := factory.CreateAgent(ctx, config)
if err != nil {
    return err
}

// Use agent in debate
response, err := agent.Respond(ctx, debateContext, message)
```

## Testing

```bash
go test -v ./internal/debate/agents/...
```

## Files

- `factory.go` - Agent creation and configuration
- `factory_test.go` - Factory unit tests
- `specialization.go` - Agent role definitions
- `specialization_test.go` - Specialization tests
- `templates.go` - Pre-configured templates
