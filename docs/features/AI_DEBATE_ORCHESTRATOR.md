# AI Debate Orchestrator Framework

## Overview

The AI Debate Orchestrator Framework is an advanced multi-agent debate system that provides sophisticated consensus-building capabilities through structured debate protocols, learning mechanisms, and knowledge management.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     DebateHandler                            │
│  ┌─────────────────┐  ┌──────────────────────────────────┐ │
│  │ Legacy Services │  │ ServiceIntegration (new)         │ │
│  │  DebateService  │  │  ├─ orchestrator                 │ │
│  │  AdvancedDebate │  │  ├─ providerRegistry            │ │
│  └────────┬────────┘  │  └─ config (feature flags)      │ │
│           ↓           └─────────────┬────────────────────┘ │
│      Fallback                       ↓                       │
└───────────────────────┬─────────────┴───────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                    Orchestrator                              │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌─────────────┐ │
│  │Agent Pool │ │Team Build │ │ Protocol  │ │  Knowledge  │ │
│  └───────────┘ └───────────┘ └───────────┘ └─────────────┘ │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌─────────────┐ │
│  │ Topology  │ │  Voting   │ │ Cognitive │ │  Learning   │ │
│  └───────────┘ └───────────┘ └───────────┘ └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. Agent System (`internal/debate/agents/`)

Manages AI agents participating in debates.

```go
// Create an agent from template
agent := agents.NewFromTemplate(agents.TemplateAnalyst, provider)

// Or create a custom agent
agent := &agents.Agent{
    ID:          "custom-analyst",
    Name:        "Custom Analyst",
    Role:        agents.RoleAnalyst,
    Specialties: []string{"data", "statistics"},
    Provider:    provider,
}
```

**Agent Roles:**
- `Moderator` - Guides debate flow
- `Analyst` - Provides data-driven analysis  
- `Critic` - Challenges assumptions
- `Synthesizer` - Combines perspectives
- `Expert` - Domain-specific knowledge

### 2. Topology System (`internal/debate/topology/`)

Defines communication patterns between agents.

**Supported Topologies:**

| Topology | Description | Use Case |
|----------|-------------|----------|
| **Mesh** | All agents communicate with all | Small groups, complex topics |
| **Star** | Central moderator coordinates | Large groups, structured debates |
| **Chain** | Sequential communication | Step-by-step analysis |

```go
// Create mesh topology
topology := topology.NewMesh(agents)

// Create star topology with moderator
topology := topology.NewStar(moderator, participants)

// Create chain topology
topology := topology.NewChain(agents, bidirectional)
```

### 3. Protocol System (`internal/debate/protocol/`)

Manages debate phases and execution.

**Debate Phases:**
1. **Proposal** - Initial positions presented
2. **Critique** - Cross-examination
3. **Review** - Address critiques
4. **Synthesis** - Build consensus

```go
protocol := protocol.NewStandard(&protocol.Config{
    MaxRounds:        5,
    PhaseTimeout:     2 * time.Minute,
    RequireConsensus: true,
})

result, err := protocol.Execute(ctx, topic, agents, topology)
```

### 4. Knowledge System (`internal/debate/knowledge/`)

Persists learnings across debates.

```go
repo := knowledge.NewRepository(storage)

// Store lesson learned
lesson := &knowledge.Lesson{
    Topic:      "API Design",
    Insight:    "REST preferred over GraphQL for simple CRUD",
    Confidence: 0.85,
    Sources:    []string{"debate-123", "debate-456"},
}
repo.AddLesson(ctx, lesson)

// Apply learnings to new debate
learnings := repo.GetRelevantLessons(ctx, newTopic)
```

### 5. Voting System (`internal/debate/voting/`)

Determines consensus from agent positions.

```go
voter := voting.NewWeightedConfidence(&voting.Config{
    MinConfidence: 0.6,
    RequireMajority: true,
})

result := voter.Vote(positions)
// result.Winner, result.Confidence, result.Breakdown
```

## Usage

### Basic Debate

```go
import (
    "dev.helix.agent/internal/debate/orchestrator"
)

// Create orchestrator
orch := orchestrator.New(config, providers, logger)

// Run debate
request := &orchestrator.DebateRequest{
    Topic: "Should we use microservices or monolith?",
    Participants: []string{"claude", "deepseek", "gemini"},
    Config: &orchestrator.DebateConfig{
        Topology:  "mesh",
        MaxRounds: 3,
        Timeout:   5 * time.Minute,
    },
}

result, err := orch.RunDebate(ctx, request)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Consensus: %s\n", result.Consensus)
fmt.Printf("Confidence: %.2f\n", result.Confidence)
```

### With Learning Enabled

```go
config := orchestrator.DefaultServiceIntegrationConfig()
config.EnableNewFramework = true
config.EnableLearning = true

orch := orchestrator.NewWithIntegration(config, providers)

// Debates will now learn from each other
result1, _ := orch.RunDebate(ctx, topic1)
result2, _ := orch.RunDebate(ctx, topic2) // Uses learnings from result1
```

### Multi-Pass Validation

```go
request := &orchestrator.DebateRequest{
    Topic: "Complex technical decision",
    EnableMultiPassValidation: true,
    ValidationConfig: &orchestrator.ValidationConfig{
        EnableValidation:     true,
        EnablePolish:         true,
        MaxValidationRounds:  3,
        MinConfidenceToSkip:  0.9,
    },
}

result, _ := orch.RunDebate(ctx, request)
// result includes validation phases and quality improvements
```

## Configuration

```go
type Config struct {
    // Feature flags
    EnableNewFramework      bool          `yaml:"enable_new_framework"`
    FallbackToLegacy        bool          `yaml:"fallback_to_legacy"`
    EnableLearning          bool          `yaml:"enable_learning"`
    
    // Thresholds
    MinAgentsForNewFramework int          `yaml:"min_agents_for_new_framework"`
    MinConfidenceThreshold   float64      `yaml:"min_confidence_threshold"`
    
    // Timeouts
    DebateTimeout           time.Duration `yaml:"debate_timeout"`
    PhaseTimeout            time.Duration `yaml:"phase_timeout"`
    
    // Topology
    DefaultTopology         string        `yaml:"default_topology"`
}
```

## API Integration

### REST Endpoint

```http
POST /v1/debates
Content-Type: application/json

{
    "topic": "Should we migrate to Kubernetes?",
    "participants": ["claude", "deepseek", "gemini"],
    "config": {
        "topology": "mesh",
        "max_rounds": 3,
        "enable_learning": true
    }
}
```

### Response

```json
{
    "id": "debate-abc123",
    "topic": "Should we migrate to Kubernetes?",
    "consensus": "Yes, Kubernetes migration is recommended...",
    "confidence": 0.87,
    "phases_completed": 4,
    "participants": [...],
    "learnings_applied": 2,
    "learnings_generated": 1
}
```

## Testing

```bash
# Run all debate tests
go test -v ./internal/debate/...

# Run orchestrator tests
go test -v ./internal/debate/orchestrator/...

# Run with coverage
go test -cover ./internal/debate/...
```

## Key Files

| File | Description |
|------|-------------|
| `internal/debate/orchestrator/orchestrator.go` | Main orchestrator |
| `internal/debate/orchestrator/service_integration.go` | Service bridge |
| `internal/debate/agents/factory.go` | Agent creation |
| `internal/debate/knowledge/repository.go` | Knowledge storage |
| `internal/debate/protocol/protocol.go` | Debate protocol |
| `internal/debate/topology/mesh.go` | Mesh topology |
| `internal/debate/voting/weighted.go` | Voting system |

## See Also

- [Debate API Reference](../api/debates.md)
- [Multi-Pass Validation](./MULTI_PASS_VALIDATION.md)
- [Provider Configuration](../guides/providers.md)
