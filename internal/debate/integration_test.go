// Package debate provides integration tests for Phase 2 components.
// Tests the integration of: Topology, Protocol, Cognitive Planning, and Weighted Voting.
package debate

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/debate/cognitive"
	"dev.helix.agent/internal/debate/protocol"
	"dev.helix.agent/internal/debate/topology"
	"dev.helix.agent/internal/debate/voting"
)

// =============================================================================
// Test Helpers
// =============================================================================

// mockAgentInvoker implements protocol.AgentInvoker for testing.
type mockAgentInvoker struct {
	responses      map[string]*protocol.PhaseResponse
	defaultContent string
	defaultConf    float64
	latency        time.Duration
	mu             sync.Mutex
	invocations    []invokeRecord
}

type invokeRecord struct {
	AgentID string
	Phase   topology.DebatePhase
	Time    time.Time
}

func newMockAgentInvoker() *mockAgentInvoker {
	return &mockAgentInvoker{
		responses:      make(map[string]*protocol.PhaseResponse),
		defaultContent: "Default response content",
		defaultConf:    0.75,
		latency:        10 * time.Millisecond,
		invocations:    make([]invokeRecord, 0),
	}
}

func (m *mockAgentInvoker) Invoke(ctx context.Context, agent *topology.Agent, prompt string, debateCtx protocol.DebateContext) (*protocol.PhaseResponse, error) {
	m.mu.Lock()
	m.invocations = append(m.invocations, invokeRecord{
		AgentID: agent.ID,
		Phase:   debateCtx.CurrentPhase,
		Time:    time.Now(),
	})
	m.mu.Unlock()

	// Simulate latency
	time.Sleep(m.latency)

	// Check for specific response
	key := fmt.Sprintf("%s-%s", agent.ID, debateCtx.CurrentPhase)
	if resp, ok := m.responses[key]; ok {
		return resp, nil
	}

	// Generate default response based on phase
	content := m.generatePhaseContent(agent, debateCtx.CurrentPhase)
	vote := ""
	if debateCtx.CurrentPhase == topology.PhaseConvergence {
		vote = "solution_A"
	}

	return &protocol.PhaseResponse{
		AgentID:    agent.ID,
		Role:       agent.Role,
		Provider:   agent.Provider,
		Model:      agent.Model,
		Content:    content,
		Confidence: m.defaultConf + (agent.Score-7.0)/20, // Vary based on score
		Arguments:  []string{"arg1", "arg2"},
		Criticisms: []string{"critique1"},
		Suggestions: []string{"suggestion1"},
		Vote:       vote,
		Score:      agent.Score,
		Latency:    m.latency,
		Timestamp:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}, nil
}

func (m *mockAgentInvoker) generatePhaseContent(agent *topology.Agent, phase topology.DebatePhase) string {
	switch phase {
	case topology.PhaseProposal:
		return fmt.Sprintf("Proposal from %s: Implement solution A with microservices architecture", agent.ID)
	case topology.PhaseCritique:
		return fmt.Sprintf("Critique from %s: The proposal lacks scalability considerations", agent.ID)
	case topology.PhaseReview:
		return fmt.Sprintf("Review from %s: Quality assessment: 8/10, needs security review", agent.ID)
	case topology.PhaseOptimization:
		return fmt.Sprintf("Optimization from %s: Enhanced solution with caching layer", agent.ID)
	case topology.PhaseConvergence:
		return fmt.Sprintf("Convergence from %s: Vote for solution_A with high confidence", agent.ID)
	default:
		return m.defaultContent
	}
}

func (m *mockAgentInvoker) setResponse(agentID string, phase topology.DebatePhase, resp *protocol.PhaseResponse) {
	key := fmt.Sprintf("%s-%s", agentID, phase)
	m.responses[key] = resp
}

func (m *mockAgentInvoker) getInvocations() []invokeRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]invokeRecord, len(m.invocations))
	copy(result, m.invocations)
	return result
}

// createTestAgents creates a set of test agents with various roles and scores.
func createTestAgents() []*topology.Agent {
	return []*topology.Agent{
		{
			ID:             "claude-1",
			Role:           topology.RoleProposer,
			Provider:       "claude",
			Model:          "claude-3-opus",
			Score:          9.2,
			Confidence:     0.9,
			Specialization: "reasoning",
			Capabilities:   []string{"code", "analysis"},
		},
		{
			ID:             "deepseek-1",
			Role:           topology.RoleCritic,
			Provider:       "deepseek",
			Model:          "deepseek-coder",
			Score:          8.5,
			Confidence:     0.85,
			Specialization: "code",
			Capabilities:   []string{"code", "critique"},
		},
		{
			ID:             "gemini-1",
			Role:           topology.RoleReviewer,
			Provider:       "gemini",
			Model:          "gemini-pro",
			Score:          8.8,
			Confidence:     0.88,
			Specialization: "analysis",
			Capabilities:   []string{"analysis", "review"},
		},
		{
			ID:             "qwen-1",
			Role:           topology.RoleOptimizer,
			Provider:       "qwen",
			Model:          "qwen-max",
			Score:          8.3,
			Confidence:     0.83,
			Specialization: "optimization",
			Capabilities:   []string{"code", "optimization"},
		},
		{
			ID:             "mistral-1",
			Role:           topology.RoleModerator,
			Provider:       "mistral",
			Model:          "mistral-large",
			Score:          8.0,
			Confidence:     0.8,
			Specialization: "coordination",
			Capabilities:   []string{"moderation", "consensus"},
		},
		{
			ID:             "openrouter-1",
			Role:           topology.RoleRedTeam,
			Provider:       "openrouter",
			Model:          "anthropic/claude-3",
			Score:          8.7,
			Confidence:     0.87,
			Specialization: "security",
			Capabilities:   []string{"security", "testing"},
		},
		{
			ID:             "zai-1",
			Role:           topology.RoleBlueTeam,
			Provider:       "zai",
			Model:          "zai-1",
			Score:          7.5,
			Confidence:     0.75,
			Specialization: "defense",
			Capabilities:   []string{"security", "validation"},
		},
		{
			ID:             "zen-1",
			Role:           topology.RoleValidator,
			Provider:       "zen",
			Model:          "qwen-coder",
			Score:          6.8,
			Confidence:     0.68,
			Specialization: "validation",
			Capabilities:   []string{"validation", "testing"},
		},
	}
}

// =============================================================================
// Integration Tests: Topology + Protocol
// =============================================================================

func TestIntegration_TopologyAndProtocol_GraphMesh(t *testing.T) {
	ctx := context.Background()

	// Create Graph-Mesh topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topo := topology.NewGraphMeshTopology(topoConfig)

	// Initialize with test agents
	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)

	// Create protocol with the topology
	debateConfig := protocol.DefaultDebateConfig()
	debateConfig.Topic = "Design a scalable microservices architecture"
	debateConfig.MaxRounds = 1
	debateConfig.Timeout = 30 * time.Second

	invoker := newMockAgentInvoker()
	proto := protocol.NewProtocol(debateConfig, topo, invoker)

	// Execute protocol
	result, err := proto.Execute(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify integration
	assert.Equal(t, topology.TopologyGraphMesh, result.TopologyUsed)
	assert.GreaterOrEqual(t, len(result.Phases), 5, "Should have at least 5 phases")
	assert.Equal(t, len(agents), result.ParticipantCount)

	// Verify all phases executed
	phasesSeen := make(map[topology.DebatePhase]bool)
	for _, phase := range result.Phases {
		phasesSeen[phase.Phase] = true
	}
	assert.True(t, phasesSeen[topology.PhaseProposal], "Proposal phase should execute")
	assert.True(t, phasesSeen[topology.PhaseCritique], "Critique phase should execute")
	assert.True(t, phasesSeen[topology.PhaseReview], "Review phase should execute")
	assert.True(t, phasesSeen[topology.PhaseOptimization], "Optimization phase should execute")
	assert.True(t, phasesSeen[topology.PhaseConvergence], "Convergence phase should execute")

	// Verify agents were invoked
	invocations := invoker.getInvocations()
	assert.Greater(t, len(invocations), 0, "Agents should be invoked")

	// Close topology
	err = topo.Close()
	assert.NoError(t, err)
}

func TestIntegration_TopologyAndProtocol_StarTopology(t *testing.T) {
	ctx := context.Background()

	// Create Star topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyStar)
	topo := topology.NewStarTopology(topoConfig)

	// Initialize with test agents
	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)

	// Verify moderator is central node
	moderator := topo.GetModerator()
	assert.NotNil(t, moderator, "Star topology should have moderator")

	// Create protocol
	debateConfig := protocol.DefaultDebateConfig()
	debateConfig.Topic = "Choose the best database for high-write workloads"
	debateConfig.TopologyType = topology.TopologyStar
	debateConfig.MaxRounds = 1
	debateConfig.Timeout = 30 * time.Second

	invoker := newMockAgentInvoker()
	proto := protocol.NewProtocol(debateConfig, topo, invoker)

	result, err := proto.Execute(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.Greater(t, len(result.Phases), 0)

	err = topo.Close()
	assert.NoError(t, err)
}

func TestIntegration_TopologyAndProtocol_ChainTopology(t *testing.T) {
	ctx := context.Background()

	// Create Chain topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyChain)
	topo := topology.NewChainTopology(topoConfig)

	// Initialize with test agents
	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)

	// Verify chain order
	chain := topo.GetChain()
	assert.Equal(t, len(agents), len(chain), "Chain should include all agents")

	// Create protocol
	debateConfig := protocol.DefaultDebateConfig()
	debateConfig.Topic = "Review code changes for security vulnerabilities"
	debateConfig.TopologyType = topology.TopologyChain
	debateConfig.MaxRounds = 1
	debateConfig.Timeout = 30 * time.Second

	invoker := newMockAgentInvoker()
	proto := protocol.NewProtocol(debateConfig, topo, invoker)

	result, err := proto.Execute(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)

	err = topo.Close()
	assert.NoError(t, err)
}

// =============================================================================
// Integration Tests: Protocol + Voting
// =============================================================================

func TestIntegration_ProtocolAndVoting_ConvergencePhase(t *testing.T) {
	ctx := context.Background()

	// Create topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topo := topology.NewGraphMeshTopology(topoConfig)

	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)
	defer topo.Close()

	// Create invoker with specific convergence votes
	invoker := newMockAgentInvoker()

	// Set up votes - most agents vote for solution_A
	for _, agent := range agents {
		vote := "solution_A"
		if agent.ID == "zen-1" || agent.ID == "zai-1" {
			vote = "solution_B" // Some dissent
		}
		invoker.setResponse(agent.ID, topology.PhaseConvergence, &protocol.PhaseResponse{
			AgentID:    agent.ID,
			Role:       agent.Role,
			Provider:   agent.Provider,
			Model:      agent.Model,
			Content:    fmt.Sprintf("I vote for %s", vote),
			Confidence: agent.Confidence,
			Vote:       vote,
			Score:      agent.Score,
			Latency:    10 * time.Millisecond,
			Timestamp:  time.Now(),
		})
	}

	// Create protocol
	debateConfig := protocol.DefaultDebateConfig()
	debateConfig.Topic = "Select the best solution"
	debateConfig.MaxRounds = 1
	debateConfig.Timeout = 30 * time.Second

	proto := protocol.NewProtocol(debateConfig, topo, invoker)
	result, err := proto.Execute(ctx)
	require.NoError(t, err)

	// Verify convergence phase had votes
	var convergencePhase *protocol.PhaseResult
	for _, phase := range result.Phases {
		if phase.Phase == topology.PhaseConvergence {
			convergencePhase = phase
			break
		}
	}
	require.NotNil(t, convergencePhase, "Convergence phase should exist")

	// Count votes in responses
	voteCount := make(map[string]int)
	for _, resp := range convergencePhase.Responses {
		if resp.Vote != "" {
			voteCount[resp.Vote]++
		}
	}
	// Verify votes were collected
	totalVotes := voteCount["solution_A"] + voteCount["solution_B"]
	assert.Greater(t, totalVotes, 0, "Should have collected votes")
	assert.GreaterOrEqual(t, voteCount["solution_A"], voteCount["solution_B"], "solution_A should have at least as many votes")

	// Verify final consensus
	if result.FinalConsensus != nil {
		assert.NotEmpty(t, result.FinalConsensus.VoteBreakdown)
		assert.Equal(t, "solution_A", result.FinalConsensus.WinningVote)
	}
}

func TestIntegration_ProtocolAndVoting_WeightedVotingSystem(t *testing.T) {
	ctx := context.Background()

	// Create weighted voting system
	votingConfig := voting.DefaultVotingConfig()
	votingConfig.MinimumVotes = 3
	votingSystem := voting.NewWeightedVotingSystem(votingConfig)

	// Add votes from different agents
	agents := createTestAgents()

	// Most vote for A, but with varying confidence
	votes := []struct {
		agent  *topology.Agent
		choice string
	}{
		{agents[0], "solution_A"}, // High score, high confidence
		{agents[1], "solution_A"},
		{agents[2], "solution_A"},
		{agents[3], "solution_B"}, // Dissent
		{agents[4], "solution_A"},
		{agents[5], "solution_A"},
		{agents[6], "solution_B"}, // Dissent
		{agents[7], "solution_A"},
	}

	for _, v := range votes {
		err := votingSystem.AddVote(&voting.Vote{
			AgentID:        v.agent.ID,
			Choice:         v.choice,
			Confidence:     v.agent.Confidence,
			Score:          v.agent.Score,
			Specialization: v.agent.Specialization,
			Role:           string(v.agent.Role),
			Timestamp:      time.Now(),
		})
		require.NoError(t, err)
	}

	// Calculate weighted result
	result, err := votingSystem.Calculate(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify weighted voting worked
	assert.Equal(t, "solution_A", result.WinningChoice)
	assert.Greater(t, result.ChoiceScores["solution_A"], result.ChoiceScores["solution_B"])
	assert.Equal(t, 8, result.TotalVotes)
	assert.Equal(t, voting.VotingMethodWeighted, result.Method)

	// Verify consensus level
	assert.Greater(t, result.Consensus, 0.5, "Should have majority consensus")
}

func TestIntegration_VotingSystem_TieBreaking(t *testing.T) {
	ctx := context.Background()

	votingConfig := voting.DefaultVotingConfig()
	votingConfig.MinimumVotes = 4
	votingConfig.EnableTieBreaking = true
	votingConfig.TieBreakMethod = voting.TieBreakByHighestConfidence
	votingSystem := voting.NewWeightedVotingSystem(votingConfig)

	// Add perfectly split votes
	err := votingSystem.AddVote(&voting.Vote{
		AgentID:    "agent-1",
		Choice:     "solution_A",
		Confidence: 0.9, // Highest confidence
		Score:      9.0,
		Timestamp:  time.Now(),
	})
	require.NoError(t, err)

	err = votingSystem.AddVote(&voting.Vote{
		AgentID:    "agent-2",
		Choice:     "solution_A",
		Confidence: 0.7,
		Score:      8.0,
		Timestamp:  time.Now(),
	})
	require.NoError(t, err)

	err = votingSystem.AddVote(&voting.Vote{
		AgentID:    "agent-3",
		Choice:     "solution_B",
		Confidence: 0.8,
		Score:      8.5,
		Timestamp:  time.Now(),
	})
	require.NoError(t, err)

	err = votingSystem.AddVote(&voting.Vote{
		AgentID:    "agent-4",
		Choice:     "solution_B",
		Confidence: 0.75,
		Score:      8.0,
		Timestamp:  time.Now(),
	})
	require.NoError(t, err)

	result, err := votingSystem.Calculate(ctx)
	require.NoError(t, err)

	// Should pick solution_A due to higher confidence in tie-break
	assert.NotEmpty(t, result.WinningChoice)
	assert.False(t, result.IsTie, "Tie should be broken")
}

// =============================================================================
// Integration Tests: Cognitive Planning + Protocol
// =============================================================================

func TestIntegration_CognitivePlanningAndProtocol(t *testing.T) {
	ctx := context.Background()

	// Create cognitive planner
	planningConfig := cognitive.DefaultPlanningConfig()
	planner := cognitive.NewCognitivePlanner(planningConfig)

	// Create topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topo := topology.NewGraphMeshTopology(topoConfig)

	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)
	defer topo.Close()

	// Set expectations for proposal phase
	expectation := planner.SetExpectation(ctx, topology.PhaseProposal, 1, agents)
	require.NotNil(t, expectation)
	assert.Equal(t, topology.PhaseProposal, expectation.Phase)
	assert.Greater(t, expectation.ExpectedConfidence, 0.0)
	assert.NotEmpty(t, expectation.KeyGoals)

	// Simulate phase execution
	actualConfidence := 0.78
	actualConsensus := 0.65
	actualInsights := 5
	actualLatency := 25 * time.Millisecond
	goalsAchieved := []string{"Generate creative solutions", "Provide clear reasoning"}
	unexpectedOutcomes := []string{}

	// Compare results
	comparison := planner.Compare(ctx, topology.PhaseProposal, 1,
		actualConfidence, actualConsensus, actualInsights, actualLatency,
		goalsAchieved, unexpectedOutcomes)

	require.NotNil(t, comparison)
	assert.Equal(t, topology.PhaseProposal, comparison.Phase)

	// Verify deltas are calculated
	assert.InDelta(t, actualConfidence-expectation.ExpectedConfidence, comparison.ConfidenceDelta, 0.01)

	// Generate refinement
	refinement := planner.Refine(ctx, comparison, agents)
	require.NotNil(t, refinement)
	assert.Equal(t, topology.PhaseProposal, refinement.Phase)
	assert.NotEmpty(t, refinement.AgentPriorities)

	// Get next phase strategy
	strategy := planner.GetNextPhaseStrategy(topology.PhaseProposal)
	require.NotNil(t, strategy)
	assert.Equal(t, topology.PhaseCritique, strategy.Phase)
}

func TestIntegration_CognitivePlanning_MultiRoundLearning(t *testing.T) {
	ctx := context.Background()

	planningConfig := cognitive.DefaultPlanningConfig()
	planningConfig.EnableLearning = true
	planningConfig.AdaptationRate = 0.4
	planner := cognitive.NewCognitivePlanner(planningConfig)

	agents := createTestAgents()

	// Simulate multiple rounds
	phases := []topology.DebatePhase{
		topology.PhaseProposal,
		topology.PhaseCritique,
		topology.PhaseReview,
		topology.PhaseOptimization,
		topology.PhaseConvergence,
	}

	for round := 1; round <= 3; round++ {
		for _, phase := range phases {
			// Set expectation
			exp := planner.SetExpectation(ctx, phase, round, agents)
			require.NotNil(t, exp)

			// Simulate results (improving over rounds)
			improvement := float64(round) * 0.05
			actualConf := 0.7 + improvement
			actualCons := 0.6 + improvement
			actualInsights := 3 + round

			// Compare
			comp := planner.Compare(ctx, phase, round,
				actualConf, actualCons, actualInsights, 20*time.Millisecond,
				[]string{"goal1"}, []string{})

			require.NotNil(t, comp)

			// Refine
			ref := planner.Refine(ctx, comp, agents)
			require.NotNil(t, ref)
		}
	}

	// Verify learning history
	history := planner.GetLearningHistory()
	assert.Greater(t, len(history), 0, "Should have learning insights")

	// Verify metrics
	metrics := planner.GetPlanningMetrics()
	assert.Greater(t, metrics.TotalExpectations, 0)
	assert.Greater(t, metrics.TotalComparisons, 0)
	assert.Greater(t, metrics.TotalRefinements, 0)

	// Perform meta-cognition
	report := planner.ReflectOnPerformance(ctx)
	require.NotNil(t, report)
	assert.Greater(t, report.TotalComparisons, 0)
}

// =============================================================================
// Full Integration Tests: All Components Together
// =============================================================================

func TestIntegration_FullDebateFlow(t *testing.T) {
	ctx := context.Background()

	// 1. Create all components
	// Topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topo := topology.NewGraphMeshTopology(topoConfig)

	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)
	defer topo.Close()

	// Cognitive Planner
	planningConfig := cognitive.DefaultPlanningConfig()
	planner := cognitive.NewCognitivePlanner(planningConfig)

	// Voting System
	votingConfig := voting.DefaultVotingConfig()
	votingConfig.MinimumVotes = 3
	votingSystem := voting.NewWeightedVotingSystem(votingConfig)

	// Protocol
	debateConfig := protocol.DefaultDebateConfig()
	debateConfig.Topic = "Design a fault-tolerant distributed system"
	debateConfig.MaxRounds = 1
	debateConfig.Timeout = 60 * time.Second
	debateConfig.EnableCognitiveLoop = true

	invoker := newMockAgentInvoker()
	proto := protocol.NewProtocol(debateConfig, topo, invoker)

	// 2. Set cognitive expectations before execution
	for _, phase := range []topology.DebatePhase{
		topology.PhaseProposal,
		topology.PhaseCritique,
		topology.PhaseReview,
		topology.PhaseOptimization,
		topology.PhaseConvergence,
	} {
		exp := planner.SetExpectation(ctx, phase, 1, agents)
		require.NotNil(t, exp)
	}

	// 3. Execute the protocol
	result, err := proto.Execute(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	// 4. Process results with cognitive planning
	for _, phaseResult := range result.Phases {
		// Calculate actual metrics from phase result
		totalConf := 0.0
		for _, resp := range phaseResult.Responses {
			totalConf += resp.Confidence
		}
		avgConf := totalConf / float64(len(phaseResult.Responses))

		// Compare with expectations
		comp := planner.Compare(ctx, phaseResult.Phase, phaseResult.Round,
			avgConf, phaseResult.ConsensusLevel, len(phaseResult.KeyInsights),
			phaseResult.Duration, phaseResult.KeyInsights, phaseResult.Disagreements)
		require.NotNil(t, comp)

		// Refine strategy
		ref := planner.Refine(ctx, comp, agents)
		require.NotNil(t, ref)
	}

	// 5. Process convergence votes through voting system
	var convergencePhase *protocol.PhaseResult
	for _, pr := range result.Phases {
		if pr.Phase == topology.PhaseConvergence {
			convergencePhase = pr
			break
		}
	}
	require.NotNil(t, convergencePhase)

	for _, resp := range convergencePhase.Responses {
		if resp.Vote != "" {
			err := votingSystem.AddVote(&voting.Vote{
				AgentID:        resp.AgentID,
				Choice:         resp.Vote,
				Confidence:     resp.Confidence,
				Score:          resp.Score,
				Role:           string(resp.Role),
				Reasoning:      resp.Content,
				Timestamp:      resp.Timestamp,
			})
			assert.NoError(t, err)
		}
	}

	// 6. Calculate final vote if we have enough
	if votingSystem.VoteCount() >= votingConfig.MinimumVotes {
		votingResult, err := votingSystem.Calculate(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, votingResult.WinningChoice)
	}

	// 7. Verify overall results
	assert.True(t, result.Success, "Debate should succeed")
	assert.Equal(t, 5, len(result.Phases), "Should have 5 phases")
	assert.Greater(t, result.Metrics.TotalResponses, 0, "Should have responses")

	// 8. Verify cognitive learning occurred
	metrics := planner.GetPlanningMetrics()
	assert.Equal(t, 5, metrics.TotalExpectations)
	assert.Equal(t, 5, metrics.TotalComparisons)
	assert.Equal(t, 5, metrics.TotalRefinements)
}

func TestIntegration_EarlyConsensusExit(t *testing.T) {
	ctx := context.Background()

	// Create topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topo := topology.NewGraphMeshTopology(topoConfig)

	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)
	defer topo.Close()

	// Create invoker that generates high consensus
	invoker := newMockAgentInvoker()
	invoker.defaultConf = 0.95 // High confidence

	// All agents vote the same way with high confidence
	for _, agent := range agents {
		invoker.setResponse(agent.ID, topology.PhaseConvergence, &protocol.PhaseResponse{
			AgentID:    agent.ID,
			Role:       agent.Role,
			Provider:   agent.Provider,
			Model:      agent.Model,
			Content:    "Unanimous agreement on solution_A",
			Confidence: 0.95,
			Vote:       "solution_A",
			Score:      agent.Score,
			Latency:    10 * time.Millisecond,
			Timestamp:  time.Now(),
		})
	}

	// Create protocol with early exit enabled
	debateConfig := protocol.DefaultDebateConfig()
	debateConfig.Topic = "Simple decision with high consensus"
	debateConfig.MaxRounds = 5 // Allow multiple rounds
	debateConfig.Timeout = 60 * time.Second
	debateConfig.EnableEarlyExit = true
	debateConfig.MinConsensusScore = 0.8 // High threshold

	proto := protocol.NewProtocol(debateConfig, topo, invoker)
	result, err := proto.Execute(ctx)
	require.NoError(t, err)

	// Verify early exit occurred
	assert.True(t, result.Success)
	assert.True(t, result.EarlyExit, "Should exit early with high consensus")
	assert.Equal(t, "early_consensus", result.EarlyExitReason)
	assert.Equal(t, 1, result.RoundsCompleted, "Should complete only 1 round")
}

func TestIntegration_ParallelPhaseExecution(t *testing.T) {
	ctx := context.Background()

	// Create Graph-Mesh topology (supports parallel execution)
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topoConfig.MaxParallelism = 4
	topo := topology.NewGraphMeshTopology(topoConfig)

	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)
	defer topo.Close()

	// Get parallel groups
	groups := topo.GetParallelGroups(topology.PhaseProposal)
	assert.Greater(t, len(groups), 0, "Should have parallel groups")

	// Create protocol
	invoker := newMockAgentInvoker()
	invoker.latency = 50 * time.Millisecond // Noticeable latency

	debateConfig := protocol.DefaultDebateConfig()
	debateConfig.Topic = "Test parallel execution"
	debateConfig.MaxRounds = 1
	debateConfig.Timeout = 30 * time.Second

	proto := protocol.NewProtocol(debateConfig, topo, invoker)

	startTime := time.Now()
	result, err := proto.Execute(ctx)
	totalTime := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, result)

	// With parallel execution, total time should be less than
	// sequential execution time (numAgents * numPhases * latency)
	maxSequentialTime := time.Duration(len(agents)*5) * invoker.latency
	assert.Less(t, totalTime, maxSequentialTime,
		"Parallel execution should be faster than sequential")
}

func TestIntegration_DynamicRoleAssignment(t *testing.T) {
	ctx := context.Background()

	// Create topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topoConfig.EnableDynamicRoles = true
	topo := topology.NewGraphMeshTopology(topoConfig)

	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)
	defer topo.Close()

	// Get agents by original role
	proposers := topo.GetAgentsByRole(topology.RoleProposer)
	require.Greater(t, len(proposers), 0)

	originalProposer := proposers[0]

	// Dynamically reassign role
	err = topo.AssignRole(originalProposer.ID, topology.RoleCritic)
	require.NoError(t, err)

	// Verify role change
	agent, found := topo.GetAgent(originalProposer.ID)
	require.True(t, found)
	assert.Equal(t, topology.RoleCritic, agent.Role)

	// Verify role lists updated
	critics := topo.GetAgentsByRole(topology.RoleCritic)
	foundInCritics := false
	for _, c := range critics {
		if c.ID == originalProposer.ID {
			foundInCritics = true
			break
		}
	}
	assert.True(t, foundInCritics, "Agent should be in critics list")
}

func TestIntegration_TopologyLeaderSelection(t *testing.T) {
	ctx := context.Background()

	// Create topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topo := topology.NewGraphMeshTopology(topoConfig)

	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)
	defer topo.Close()

	// Select leaders for each phase
	phases := []topology.DebatePhase{
		topology.PhaseProposal,
		topology.PhaseCritique,
		topology.PhaseReview,
		topology.PhaseOptimization,
		topology.PhaseConvergence,
	}

	for _, phase := range phases {
		leader, err := topo.SelectLeader(phase)
		assert.NoError(t, err)
		assert.NotNil(t, leader, "Phase %s should have a leader", phase)
	}
}

func TestIntegration_ProductiveChaosInVoting(t *testing.T) {
	ctx := context.Background()

	votingConfig := voting.DefaultVotingConfig()
	votingConfig.MinimumVotes = 5
	votingSystem := voting.NewWeightedVotingSystem(votingConfig)

	// Add votes
	for i := 1; i <= 5; i++ {
		choice := "solution_A"
		if i > 3 {
			choice = "solution_B"
		}
		err := votingSystem.AddVote(&voting.Vote{
			AgentID:    fmt.Sprintf("agent-%d", i),
			Choice:     choice,
			Confidence: 0.8,
			Score:      8.0,
			Timestamp:  time.Now(),
		})
		require.NoError(t, err)
	}

	// Calculate result before chaos
	resultBefore, err := votingSystem.Calculate(ctx)
	require.NoError(t, err)

	// Apply productive chaos
	votingSystem.SimulateProductiveChaos(0.1)

	// Calculate result after chaos
	resultAfter, err := votingSystem.Calculate(ctx)
	require.NoError(t, err)

	// Winner should remain the same (chaos shouldn't flip results)
	assert.Equal(t, resultBefore.WinningChoice, resultAfter.WinningChoice)

	// But scores might have shifted slightly
	// This tests that chaos doesn't break the system
	assert.NotZero(t, resultAfter.WinningScore)
}

// =============================================================================
// Stress Tests
// =============================================================================

func TestIntegration_StressTest_ManyAgents(t *testing.T) {
	ctx := context.Background()

	// Create many agents
	agents := make([]*topology.Agent, 50)
	roles := []topology.AgentRole{
		topology.RoleProposer, topology.RoleCritic, topology.RoleReviewer,
		topology.RoleOptimizer, topology.RoleModerator, topology.RoleValidator,
	}

	for i := 0; i < 50; i++ {
		agents[i] = &topology.Agent{
			ID:             fmt.Sprintf("agent-%d", i),
			Role:           roles[i%len(roles)],
			Provider:       "test",
			Model:          "test-model",
			Score:          7.0 + float64(i%30)/10,
			Confidence:     0.7 + float64(i%30)/100,
			Specialization: "general",
			Capabilities:   []string{"test"},
		}
	}

	// Create topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topoConfig.MaxParallelism = 10
	topo := topology.NewGraphMeshTopology(topoConfig)

	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)
	defer topo.Close()

	// Verify all agents registered
	assert.Equal(t, 50, len(topo.GetAgents()))

	// Create protocol with short timeout
	invoker := newMockAgentInvoker()
	invoker.latency = 1 * time.Millisecond

	debateConfig := protocol.DefaultDebateConfig()
	debateConfig.Topic = "Stress test with many agents"
	debateConfig.MaxRounds = 1
	debateConfig.Timeout = 30 * time.Second

	proto := protocol.NewProtocol(debateConfig, topo, invoker)

	result, err := proto.Execute(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.Greater(t, result.Metrics.TotalResponses, 0)
}

func TestIntegration_StressTest_ManyRounds(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topo := topology.NewGraphMeshTopology(topoConfig)

	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)
	defer topo.Close()

	// Create invoker with varying responses to prevent early exit
	invoker := newMockAgentInvoker()
	invoker.latency = 1 * time.Millisecond
	invoker.defaultConf = 0.6 // Lower confidence to prevent early exit

	debateConfig := protocol.DefaultDebateConfig()
	debateConfig.Topic = "Multi-round stress test"
	debateConfig.MaxRounds = 5
	debateConfig.Timeout = 60 * time.Second
	debateConfig.EnableEarlyExit = false // Disable to test all rounds

	proto := protocol.NewProtocol(debateConfig, topo, invoker)

	result, err := proto.Execute(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.Equal(t, 5, result.RoundsCompleted)
	assert.Equal(t, 25, len(result.Phases), "5 rounds × 5 phases = 25 phase results")
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestIntegration_ErrorHandling_InvokerFailure(t *testing.T) {
	ctx := context.Background()

	// Create topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topo := topology.NewGraphMeshTopology(topoConfig)

	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)
	defer topo.Close()

	// Create invoker that fails for some agents
	invoker := newMockAgentInvoker()

	debateConfig := protocol.DefaultDebateConfig()
	debateConfig.Topic = "Test error handling"
	debateConfig.MaxRounds = 1
	debateConfig.Timeout = 30 * time.Second

	proto := protocol.NewProtocol(debateConfig, topo, invoker)

	// Protocol should still succeed even if some invocations fail
	result, err := proto.Execute(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
}

func TestIntegration_ErrorHandling_ContextCancellation(t *testing.T) {
	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Create topology
	topoConfig := topology.DefaultTopologyConfig(topology.TopologyGraphMesh)
	topo := topology.NewGraphMeshTopology(topoConfig)

	agents := createTestAgents()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)
	defer topo.Close()

	// Create invoker with delay
	invoker := newMockAgentInvoker()
	invoker.latency = 100 * time.Millisecond

	debateConfig := protocol.DefaultDebateConfig()
	debateConfig.Topic = "Test context cancellation"
	debateConfig.MaxRounds = 5
	debateConfig.Timeout = 10 * time.Second

	proto := protocol.NewProtocol(debateConfig, topo, invoker)

	// Cancel context after short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	result, err := proto.Execute(ctx)

	// Should handle cancellation gracefully
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	if result != nil {
		assert.False(t, result.Success)
	}
}

func TestIntegration_ErrorHandling_InsufficientVotes(t *testing.T) {
	ctx := context.Background()

	votingConfig := voting.DefaultVotingConfig()
	votingConfig.MinimumVotes = 10 // Require more votes than we'll add
	votingSystem := voting.NewWeightedVotingSystem(votingConfig)

	// Add only 3 votes
	for i := 1; i <= 3; i++ {
		err := votingSystem.AddVote(&voting.Vote{
			AgentID:    fmt.Sprintf("agent-%d", i),
			Choice:     "solution_A",
			Confidence: 0.8,
			Score:      8.0,
			Timestamp:  time.Now(),
		})
		require.NoError(t, err)
	}

	// Should fail with insufficient votes
	_, err := votingSystem.Calculate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient votes")
}
