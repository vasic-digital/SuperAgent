//go:build performance
// +build performance

package performance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"dev.helix.agent/internal/debate/reflexion"
	"dev.helix.agent/internal/debate/topology"
	"dev.helix.agent/internal/debate/voting"
)

// =============================================================================
// VOTING BENCHMARKS
// =============================================================================

// BenchmarkWeightedVoting benchmarks the weighted voting calculation with
// the MiniMax formula L* = argmax sum(ci * 1[ai = L]).
func BenchmarkWeightedVoting(b *testing.B) {
	config := voting.VotingConfig{
		MinimumVotes:         3,
		MinimumConfidence:    0.1,
		EnableDiversityBonus: true,
		DiversityWeight:      0.1,
		EnableTieBreaking:    true,
		TieBreakMethod:       voting.TieBreakByHighestConfidence,
		EnableHistoricalWeight: true,
	}

	voteSystem := voting.NewWeightedVotingSystem(config)

	// Pre-populate with votes
	for i := 0; i < 25; i++ {
		_ = voteSystem.AddVote(&voting.Vote{
			AgentID:        fmt.Sprintf("agent-%d", i),
			Choice:         fmt.Sprintf("choice-%d", i%5),
			Confidence:     0.5 + float64(i%50)/100.0,
			Score:          6.0 + float64(i%40)/10.0,
			Specialization: fmt.Sprintf("spec-%d", i%3),
			Role:           fmt.Sprintf("role-%d", i%4),
			Timestamp:      time.Now(),
		})
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = voteSystem.Calculate(ctx)
	}
}

// BenchmarkBordaVoting benchmarks the Borda count voting method.
func BenchmarkBordaVoting(b *testing.B) {
	config := voting.VotingConfig{
		MinimumVotes:      3,
		MinimumConfidence: 0.1,
	}

	voteSystem := voting.NewWeightedVotingSystem(config)

	// Build rankings for Borda count
	rankings := make(map[string][]string)
	choices := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	for i := 0; i < 15; i++ {
		voterID := fmt.Sprintf("voter-%d", i)
		ranking := make([]string, len(choices))
		copy(ranking, choices)
		// Rotate rankings based on voter index
		offset := i % len(choices)
		for j := range ranking {
			ranking[j] = choices[(j+offset)%len(choices)]
		}
		rankings[voterID] = ranking
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = voteSystem.CalculateBordaCount(ctx, rankings)
	}
}

// BenchmarkCondorcetVoting benchmarks the Condorcet voting method with
// pairwise comparisons.
func BenchmarkCondorcetVoting(b *testing.B) {
	config := voting.VotingConfig{
		MinimumVotes:      3,
		MinimumConfidence: 0.1,
	}

	voteSystem := voting.NewWeightedVotingSystem(config)

	// Build rankings for Condorcet
	rankings := make(map[string][]string)
	choices := []string{"alpha", "beta", "gamma", "delta"}
	for i := 0; i < 20; i++ {
		voterID := fmt.Sprintf("voter-%d", i)
		ranking := make([]string, len(choices))
		copy(ranking, choices)
		// Create varied preference orderings
		offset := i % len(choices)
		for j := range ranking {
			ranking[j] = choices[(j+offset)%len(choices)]
		}
		rankings[voterID] = ranking
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = voteSystem.CalculateCondorcet(ctx, rankings)
	}
}

// BenchmarkPluralityVoting benchmarks the plurality (first-past-the-post)
// voting method.
func BenchmarkPluralityVoting(b *testing.B) {
	config := voting.VotingConfig{
		MinimumVotes:      3,
		MinimumConfidence: 0.1,
		EnableTieBreaking: true,
		TieBreakMethod:    voting.TieBreakByMostVotes,
	}

	voteSystem := voting.NewWeightedVotingSystem(config)
	for i := 0; i < 25; i++ {
		_ = voteSystem.AddVote(&voting.Vote{
			AgentID:    fmt.Sprintf("agent-%d", i),
			Choice:     fmt.Sprintf("choice-%d", i%5),
			Confidence: 0.6 + float64(i%40)/100.0,
			Timestamp:  time.Now(),
		})
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = voteSystem.CalculatePlurality(ctx)
	}
}

// BenchmarkUnanimousVoting benchmarks the unanimous voting method.
func BenchmarkUnanimousVoting(b *testing.B) {
	config := voting.VotingConfig{
		MinimumVotes:      3,
		MinimumConfidence: 0.1,
	}

	voteSystem := voting.NewWeightedVotingSystem(config)
	// All votes agree for unanimous
	for i := 0; i < 10; i++ {
		_ = voteSystem.AddVote(&voting.Vote{
			AgentID:    fmt.Sprintf("agent-%d", i),
			Choice:     "consensus",
			Confidence: 0.9,
			Timestamp:  time.Now(),
		})
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = voteSystem.CalculateUnanimous(ctx)
	}
}

// =============================================================================
// EPISODIC MEMORY BENCHMARKS
// =============================================================================

// BenchmarkEpisodicMemoryStore benchmarks storing episodes in the
// episodic memory buffer.
func BenchmarkEpisodicMemoryStore(b *testing.B) {
	memory := reflexion.NewEpisodicMemoryBuffer(10000)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = memory.Store(&reflexion.Episode{
			AgentID:         fmt.Sprintf("agent-%d", i%100),
			SessionID:       fmt.Sprintf("session-%d", i%50),
			TaskDescription: fmt.Sprintf("Benchmark task %d", i),
			AttemptNumber:   i%5 + 1,
			Code:            fmt.Sprintf("func bench%d() {}", i),
			Confidence:      float64(i%100) / 100.0,
			Timestamp:       time.Now(),
		})
	}
}

// BenchmarkEpisodicMemoryRetrieve benchmarks retrieving episodes by agent.
func BenchmarkEpisodicMemoryRetrieve(b *testing.B) {
	memory := reflexion.NewEpisodicMemoryBuffer(10000)

	// Pre-populate
	for i := 0; i < 5000; i++ {
		_ = memory.Store(&reflexion.Episode{
			AgentID:         fmt.Sprintf("agent-%d", i%100),
			SessionID:       fmt.Sprintf("session-%d", i%50),
			TaskDescription: fmt.Sprintf("Task %d", i),
			Confidence:      float64(i%100) / 100.0,
			Timestamp:       time.Now(),
		})
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = memory.GetByAgent(fmt.Sprintf("agent-%d", i%100))
	}
}

// BenchmarkEpisodicMemoryRelevance benchmarks relevance search in
// episodic memory.
func BenchmarkEpisodicMemoryRelevance(b *testing.B) {
	memory := reflexion.NewEpisodicMemoryBuffer(5000)

	// Pre-populate with varied task descriptions
	tasks := []string{
		"Implement binary search tree traversal",
		"Design REST API for user management",
		"Optimize database query performance",
		"Write unit tests for authentication module",
		"Refactor payment processing pipeline",
	}

	for i := 0; i < 1000; i++ {
		_ = memory.Store(&reflexion.Episode{
			AgentID:         fmt.Sprintf("agent-%d", i%10),
			TaskDescription: tasks[i%len(tasks)],
			Confidence:      float64(i%100) / 100.0,
			Timestamp:       time.Now(),
		})
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = memory.GetRelevant("database query optimization", 10)
	}
}

// =============================================================================
// TOPOLOGY BENCHMARKS
// =============================================================================

// BenchmarkTopologyInitialization benchmarks creating and initializing
// a graph mesh topology with agents.
func BenchmarkTopologyInitialization(b *testing.B) {
	agents := make([]*topology.Agent, 25)
	roles := []topology.AgentRole{
		topology.RoleModerator, topology.RoleProposer,
		topology.RoleCritic, topology.RoleReviewer,
		topology.RoleOptimizer,
	}
	for i := 0; i < 25; i++ {
		agents[i] = topology.CreateAgentFromSpec(
			fmt.Sprintf("agent-%d", i),
			roles[i%len(roles)],
			"mock", "mock-model",
			7.0+float64(i%30)/10.0,
			"general",
		)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cfg := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
		topo, err := topology.NewTopology(topology.TopologyGraphMesh, cfg)
		if err != nil {
			b.Fatal(err)
		}
		_ = topo.Initialize(context.Background(), agents)
		_ = topo.Close()
	}
}

// BenchmarkTopologySelection benchmarks the topology selection algorithm
// based on agent count and requirements.
func BenchmarkTopologySelection(b *testing.B) {
	requirements := topology.TopologyRequirements{
		MaxLatency:         100,
		RequireOrdering:    false,
		MaxParallelism:     8,
		EnableDynamicRoles: true,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = topology.SelectTopologyType(i%30+3, requirements)
	}
}

// BenchmarkMessageRouting benchmarks message routing in a graph mesh
// topology with 25 agents.
func BenchmarkMessageRouting(b *testing.B) {
	cfg := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topo, err := topology.NewTopology(topology.TopologyGraphMesh, cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer topo.Close()

	agents := make([]*topology.Agent, 15)
	roles := []topology.AgentRole{
		topology.RoleModerator, topology.RoleProposer,
		topology.RoleCritic, topology.RoleReviewer,
		topology.RoleOptimizer,
	}
	for i := 0; i < 15; i++ {
		agents[i] = topology.CreateAgentFromSpec(
			fmt.Sprintf("agent-%d", i),
			roles[i%len(roles)],
			"mock", "mock-model",
			7.0+float64(i%30)/10.0,
			"general",
		)
	}
	_ = topo.Initialize(context.Background(), agents)

	msg := &topology.Message{
		ID:          "bench-msg",
		FromAgent:   "agent-0",
		ToAgents:    []string{"agent-1", "agent-2"},
		Content:     "Benchmark message",
		MessageType: topology.MessageTypeProposal,
		Phase:       topology.PhaseProposal,
		Round:       1,
		Timestamp:   time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = topo.RouteMessage(msg)
	}
}

// BenchmarkAgentRoleAssignment benchmarks dynamic role assignment in
// a topology.
func BenchmarkAgentRoleAssignment(b *testing.B) {
	cfg := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topo, err := topology.NewTopology(topology.TopologyGraphMesh, cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer topo.Close()

	agents := make([]*topology.Agent, 10)
	for i := 0; i < 10; i++ {
		agents[i] = topology.CreateAgentFromSpec(
			fmt.Sprintf("agent-%d", i),
			topology.RoleProposer,
			"mock", "mock-model", 7.0, "general",
		)
	}
	_ = topo.Initialize(context.Background(), agents)

	roles := []topology.AgentRole{
		topology.RoleProposer, topology.RoleCritic,
		topology.RoleReviewer, topology.RoleOptimizer,
		topology.RoleModerator,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		agentID := fmt.Sprintf("agent-%d", i%10)
		role := roles[i%len(roles)]
		_ = topo.AssignRole(agentID, role)
	}
}

// =============================================================================
// ACCUMULATED WISDOM BENCHMARKS
// =============================================================================

// BenchmarkAccumulatedWisdomStore benchmarks storing wisdom entries.
func BenchmarkAccumulatedWisdomStore(b *testing.B) {
	aw := reflexion.NewAccumulatedWisdom()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = aw.Store(&reflexion.Wisdom{
			Pattern:   fmt.Sprintf("Pattern %d: common error type", i),
			Frequency: i%10 + 1,
			Impact:    float64(i%100) / 100.0,
			Domain:    "code",
			Tags:      []string{"error", "pattern", fmt.Sprintf("tag-%d", i%5)},
		})
	}
}

// BenchmarkAccumulatedWisdomRelevance benchmarks relevance search in
// accumulated wisdom.
func BenchmarkAccumulatedWisdomRelevance(b *testing.B) {
	aw := reflexion.NewAccumulatedWisdom()

	// Pre-populate
	patterns := []string{
		"Nil pointer dereference in concurrent code",
		"SQL injection through string concatenation",
		"Race condition in shared state access",
		"Buffer overflow from unchecked input",
		"Authentication bypass through parameter tampering",
	}
	for i := 0; i < 500; i++ {
		_ = aw.Store(&reflexion.Wisdom{
			Pattern:   patterns[i%len(patterns)],
			Frequency: i%10 + 1,
			Impact:    float64(i%100) / 100.0,
			Domain:    "code",
			Tags:      []string{"security", "bug", fmt.Sprintf("tag-%d", i%10)},
		})
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = aw.GetRelevant("nil pointer concurrent access", 10)
	}
}

// =============================================================================
// VOTE ADD/RESET CYCLE BENCHMARK
// =============================================================================

// BenchmarkVoteAddResetCycle benchmarks a full cycle of adding votes,
// calculating results, and resetting.
func BenchmarkVoteAddResetCycle(b *testing.B) {
	config := voting.DefaultVotingConfig()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		voteSystem := voting.NewWeightedVotingSystem(config)
		for j := 0; j < 10; j++ {
			_ = voteSystem.AddVote(&voting.Vote{
				AgentID:    fmt.Sprintf("agent-%d", j),
				Choice:     fmt.Sprintf("choice-%d", j%3),
				Confidence: 0.7,
				Score:      7.0,
				Timestamp:  time.Now(),
			})
		}
		ctx := context.Background()
		_, _ = voteSystem.Calculate(ctx)
		voteSystem.Reset()
	}
}
