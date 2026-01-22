# Debate Protocol Package

This package provides the phase-based debate execution protocol for the AI Debate Orchestrator Framework.

## Overview

The protocol package defines and executes the structured phases of a debate, ensuring orderly progression from proposal through synthesis.

## Debate Phases

| Phase | Icon | Description |
|-------|------|-------------|
| **Proposal** | 📝 | Initial positions presented |
| **Critique** | 🔍 | Critical evaluation of proposals |
| **Review** | ✓ | Cross-review and validation |
| **Synthesis** | 📜 | Final consensus building |

## Components

### Protocol Runner (`protocol.go`)

Executes debates according to the defined protocol:

```go
runner := protocol.NewRunner(protocol.Config{
    Phases:        []protocol.Phase{Proposal, Critique, Review, Synthesis},
    PhaseTimeout:  60 * time.Second,
    MaxIterations: 3,
})

result, err := runner.Execute(ctx, debateContext, agents)
```

### Phase Handlers

Each phase has specific handling logic:

```go
// Proposal phase - agents present initial positions
proposalResults := runner.RunPhase(ctx, protocol.PhaseProposal, agents)

// Critique phase - agents evaluate each other's positions
critiqueResults := runner.RunPhase(ctx, protocol.PhaseCritique, agents)

// Review phase - validation and refinement
reviewResults := runner.RunPhase(ctx, protocol.PhaseReview, agents)

// Synthesis phase - build consensus
synthesisResult := runner.RunPhase(ctx, protocol.PhaseSynthesis, agents)
```

## Architecture

```
┌─────────────────────────────────────────────┐
│            Protocol Execution               │
│                                             │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐ │
│  │Proposal │ →  │Critique │ →  │ Review  │ │
│  │   📝    │    │   🔍    │    │   ✓     │ │
│  └─────────┘    └─────────┘    └─────────┘ │
│                                     │       │
│                                     ▼       │
│                              ┌─────────┐   │
│                              │Synthesis│   │
│                              │   📜    │   │
│                              └─────────┘   │
│                                     │       │
│                                     ▼       │
│                              ┌─────────┐   │
│                              │ Result  │   │
│                              └─────────┘   │
└─────────────────────────────────────────────┘
```

## Configuration

```go
config := protocol.Config{
    Phases: []protocol.Phase{
        protocol.PhaseProposal,
        protocol.PhaseCritique,
        protocol.PhaseReview,
        protocol.PhaseSynthesis,
    },
    PhaseTimeout:      60 * time.Second,
    MaxIterations:     3,
    AllowEarlyConsensus: true,
    MinConfidenceThreshold: 0.8,
}
```

## Usage

```go
import "dev.helix.agent/internal/debate/protocol"

// Create protocol runner
runner := protocol.NewRunner(protocol.DefaultConfig())

// Execute debate
result, err := runner.Execute(ctx, protocol.ExecutionContext{
    Topic:      "AI Governance",
    Agents:     agents,
    Topology:   topology,
    Knowledge:  relevantLessons,
})

// Access phase results
for _, phase := range result.Phases {
    fmt.Printf("Phase %s: %d responses\n", phase.Name, len(phase.Responses))
}
```

## Phase Details

### Proposal Phase
- Each agent presents initial position
- Independent responses (no cross-influence)
- Output: Collection of initial positions

### Critique Phase
- Agents review others' proposals
- Identify strengths and weaknesses
- Output: Critique of each proposal

### Review Phase
- Cross-validation of critiques
- Refinement of positions
- Output: Validated positions

### Synthesis Phase
- Build consensus from validated positions
- Aggregate insights
- Output: Final synthesis with confidence

## Testing

```bash
go test -v ./internal/debate/protocol/...
```

## Files

- `protocol.go` - Protocol execution engine
- `protocol_test.go` - Unit tests
