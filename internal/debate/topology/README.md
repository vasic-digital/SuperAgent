# Debate Topology Package

This package provides communication topology configurations for the AI Debate Orchestrator Framework.

## Overview

The topology package defines how agents communicate during debates. Different topologies suit different debate styles and objectives.

## Supported Topologies

### Mesh Topology

All agents communicate with all others (parallel execution):

```
    ┌───┐
    │ A │
    └─┬─┘
   ╱  │  ╲
┌───┐ │ ┌───┐
│ B │─┼─│ C │
└───┘ │ └───┘
   ╲  │  ╱
    ┌─┴─┐
    │ D │
    └───┘
```

Best for: Maximum diversity, parallel processing

### Star Topology

Hub-and-spoke with central moderator:

```
      ┌───┐
      │ B │
      └─┬─┘
        │
┌───┐ ┌─┴─┐ ┌───┐
│ A │─│HUB│─│ C │
└───┘ └─┬─┘ └───┘
        │
      ┌─┴─┐
      │ D │
      └───┘
```

Best for: Moderated discussions, controlled flow

### Chain Topology

Sequential communication:

```
┌───┐   ┌───┐   ┌───┐   ┌───┐
│ A │ → │ B │ → │ C │ → │ D │
└───┘   └───┘   └───┘   └───┘
```

Best for: Iterative refinement, building on ideas

## Components

### Topology Factory (`factory.go`)

Creates topology configurations:

```go
topo := topology.Create(topology.TypeMesh, agents)
```

### Graph Mesh (`graph_mesh.go`)

Full connectivity implementation:

```go
mesh := topology.NewGraphMesh(agents)
connections := mesh.GetConnections(agent)
```

### Topology Base (`topology.go`)

Common topology interfaces and utilities:

```go
type Topology interface {
    GetConnections(agent Agent) []Agent
    GetCommunicationOrder() [][]Agent
    SupportsParallel() bool
}
```

## Configuration

```go
config := topology.Config{
    Type:           topology.TypeMesh,
    AllowSelfLoop:  false,
    MaxConnections: 0, // unlimited
}
```

## Usage

```go
import "dev.helix.agent/internal/debate/topology"

// Create mesh topology for full connectivity
topo := topology.Create(topology.TypeMesh, agents)

// Get communication order
order := topo.GetCommunicationOrder()
for round, participants := range order {
    fmt.Printf("Round %d: %v\n", round, participants)
}

// Check who an agent can communicate with
connections := topo.GetConnections(agentA)
```

## Topology Comparison

| Topology | Parallel | Complexity | Use Case |
|----------|----------|------------|----------|
| Mesh | Yes | O(n²) | Comprehensive debate |
| Star | Partial | O(n) | Moderated discussion |
| Chain | No | O(n) | Sequential refinement |

## Testing

```bash
go test -v ./internal/debate/topology/...
```

## Files

- `topology.go` - Base types and interfaces
- `factory.go` - Topology creation
- `graph_mesh.go` - Mesh implementation
