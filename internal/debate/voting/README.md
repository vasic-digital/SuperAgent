# Debate Voting Package

This package provides voting strategies and consensus mechanisms for the AI Debate Orchestrator Framework.

## Overview

The voting package implements various strategies for aggregating agent positions into a final consensus, with confidence weighting based on LLMsVerifier scores.

## Voting Strategies

### Weighted Confidence Voting

Primary strategy using LLMsVerifier scores:

```go
voting := voting.NewWeightedVoting(voting.Config{
    MinConfidence:  0.5,
    WeightByScore:  true,
    TieBreaker:     voting.TieBreakerHighestScore,
})

result := voting.Aggregate(positions)
```

### Majority Voting

Simple majority rules:

```go
voting := voting.NewMajorityVoting()
result := voting.Aggregate(positions)
```

### Unanimous Consensus

All agents must agree:

```go
voting := voting.NewUnanimousVoting(voting.Config{
    AllowAbstention: true,
})
```

## Components

### Weighted Voting (`weighted_voting.go`)

Main implementation with confidence weighting:

```go
type WeightedVoting struct {
    config Config
}

func (v *WeightedVoting) Aggregate(positions []Position) AggregationResult {
    // Weight each position by agent's verification score
    // Normalize weights
    // Calculate weighted consensus
    // Return result with confidence
}
```

## Architecture

```
┌─────────────────────────────────────────────┐
│              Voting System                   │
│                                             │
│  ┌─────────────────────────────────────┐   │
│  │           Input Positions           │   │
│  │  ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐    │   │
│  │  │P1 │ │P2 │ │P3 │ │P4 │ │P5 │    │   │
│  │  │0.9│ │0.8│ │0.7│ │0.85│ │0.75│  │   │
│  │  └───┘ └───┘ └───┘ └───┘ └───┘    │   │
│  └───────────────┬─────────────────────┘   │
│                  ▼                          │
│  ┌─────────────────────────────────────┐   │
│  │       Weighting & Aggregation       │   │
│  │  Score-based weights → Normalize    │   │
│  └───────────────┬─────────────────────┘   │
│                  ▼                          │
│  ┌─────────────────────────────────────┐   │
│  │         Consensus Result            │   │
│  │  Position + Confidence + Breakdown  │   │
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

## Configuration

```go
config := voting.Config{
    Strategy:        voting.StrategyWeighted,
    MinConfidence:   0.5,
    WeightByScore:   true,      // Use LLMsVerifier scores
    NormalizeWeights: true,
    TieBreaker:      voting.TieBreakerHighestScore,
    RequireQuorum:   0.5,       // 50% participation
}
```

## Usage

```go
import "dev.helix.agent/internal/debate/voting"

// Create weighted voting strategy
voter := voting.NewWeightedVoting(voting.DefaultConfig())

// Collect positions from agents
positions := []voting.Position{
    {AgentID: "claude", Position: "approach-a", Confidence: 0.9, Score: 8.5},
    {AgentID: "gemini", Position: "approach-a", Confidence: 0.8, Score: 7.8},
    {AgentID: "deepseek", Position: "approach-b", Confidence: 0.85, Score: 8.0},
}

// Aggregate votes
result := voter.Aggregate(positions)

fmt.Printf("Consensus: %s\n", result.Consensus)
fmt.Printf("Confidence: %.2f\n", result.Confidence)
fmt.Printf("Agreement: %.0f%%\n", result.AgreementRatio*100)
```

## Weight Calculation

Weights are derived from LLMsVerifier scores:

```
weight_i = score_i / Σ(scores)

Final position weight = Σ(weights of supporting agents)
```

## Result Structure

```go
type AggregationResult struct {
    Consensus      string           // Winning position
    Confidence     float64          // Overall confidence (0-1)
    AgreementRatio float64          // Proportion in agreement
    Breakdown      map[string]float64 // Position → weighted support
    Dissenters     []string         // Agents who disagreed
}
```

## Testing

```bash
go test -v ./internal/debate/voting/...
```

## Files

- `weighted_voting.go` - Main voting implementation
