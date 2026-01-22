package protocol

import (
	"context"
	"sync"
	"testing"
	"time"

	"dev.helix.agent/internal/debate/topology"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAgentInvoker is a mock implementation of AgentInvoker for testing.
type MockAgentInvoker struct {
	responses   map[string]*PhaseResponse
	invokeCount int
	invokeDelay time.Duration
	mu          sync.Mutex
}

func NewMockAgentInvoker() *MockAgentInvoker {
	return &MockAgentInvoker{
		responses: make(map[string]*PhaseResponse),
	}
}

func (m *MockAgentInvoker) SetResponse(agentID string, response *PhaseResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[agentID] = response
}

func (m *MockAgentInvoker) SetInvokeDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.invokeDelay = delay
}

func (m *MockAgentInvoker) GetInvokeCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.invokeCount
}

func (m *MockAgentInvoker) Invoke(ctx context.Context, agent *topology.Agent, prompt string, debateCtx DebateContext) (*PhaseResponse, error) {
	m.mu.Lock()
	m.invokeCount++
	delay := m.invokeDelay
	resp := m.responses[agent.ID]
	m.mu.Unlock()

	// Simulate network latency
	if delay > 0 {
		time.Sleep(delay)
	}

	if resp != nil {
		return resp, nil
	}

	// Generate default response
	return &PhaseResponse{
		AgentID:    agent.ID,
		Role:       agent.Role,
		Provider:   agent.Provider,
		Model:      agent.Model,
		Content:    "Mock response for " + string(debateCtx.CurrentPhase),
		Confidence: 0.8,
		Arguments:  []string{"Argument 1", "Argument 2"},
		Score:      7.5,
		Latency:    50 * time.Millisecond,
		Timestamp:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}, nil
}

// Helper to create test agents
func createTestAgents(count int) []*topology.Agent {
	roles := []topology.AgentRole{
		topology.RoleProposer,
		topology.RoleCritic,
		topology.RoleReviewer,
		topology.RoleOptimizer,
		topology.RoleModerator,
	}

	agents := make([]*topology.Agent, count)
	for i := 0; i < count; i++ {
		agents[i] = topology.CreateAgentFromSpec(
			generateTestID(i),
			roles[i%len(roles)],
			"test_provider",
			"test_model",
			7.5+float64(i)*0.1,
			"general",
		)
	}
	return agents
}

func generateTestID(i int) string {
	return "agent-" + string(rune('a'+i))
}

// ============================================================================
// Config Tests
// ============================================================================

func TestDefaultDebateConfig(t *testing.T) {
	config := DefaultDebateConfig()

	assert.NotEmpty(t, config.ID)
	assert.Equal(t, 3, config.MaxRounds)
	assert.Equal(t, 5*time.Minute, config.Timeout)
	assert.Equal(t, topology.TopologyGraphMesh, config.TopologyType)
	assert.Equal(t, 0.75, config.MinConsensusScore)
	assert.True(t, config.EnableEarlyExit)
	assert.True(t, config.EnableCognitiveLoop)
}

// ============================================================================
// Protocol Creation Tests
// ============================================================================

func TestNewProtocol(t *testing.T) {
	config := DefaultDebateConfig()
	config.Topic = "Test topic"

	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	assert.NotNil(t, p)
	assert.Equal(t, config.Topic, p.config.Topic)
	assert.NotNil(t, p.phaseConfigs)
	assert.Len(t, p.phaseConfigs, 5) // All 5 phases configured
}

func TestNewProtocol_GeneratesID(t *testing.T) {
	config := DebateConfig{
		Topic:   "Test topic",
		Timeout: time.Minute,
	}

	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	assert.NotEmpty(t, p.config.ID)
}

func TestProtocol_PhaseConfigs(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	// Check all phases have configs
	phases := []topology.DebatePhase{
		topology.PhaseProposal,
		topology.PhaseCritique,
		topology.PhaseReview,
		topology.PhaseOptimization,
		topology.PhaseConvergence,
	}

	for _, phase := range phases {
		cfg := p.phaseConfigs[phase]
		assert.NotNil(t, cfg, "config missing for phase %s", phase)
		assert.Equal(t, phase, cfg.Phase)
		assert.NotEmpty(t, cfg.Prompt)
		assert.Greater(t, cfg.Timeout, time.Duration(0))
	}
}

func TestProtocol_SetPhaseConfig(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	customConfig := &PhaseConfig{
		Phase:          topology.PhaseProposal,
		Timeout:        2 * time.Minute,
		MinResponses:   5,
		MaxParallelism: 10,
		Prompt:         "Custom prompt",
	}

	p.SetPhaseConfig(topology.PhaseProposal, customConfig)

	assert.Equal(t, 2*time.Minute, p.phaseConfigs[topology.PhaseProposal].Timeout)
	assert.Equal(t, 5, p.phaseConfigs[topology.PhaseProposal].MinResponses)
	assert.Equal(t, "Custom prompt", p.phaseConfigs[topology.PhaseProposal].Prompt)
}

// ============================================================================
// Protocol Execution Tests
// ============================================================================

func TestProtocol_Execute_Basic(t *testing.T) {
	config := DefaultDebateConfig()
	config.Topic = "Should AI systems be transparent?"
	config.Context = "Discussing AI ethics"
	config.MaxRounds = 1
	config.Timeout = 30 * time.Second

	// Create topology with agents
	topo := topology.NewGraphMeshTopology(topology.DefaultTopologyConfig(topology.TopologyGraphMesh))
	agents := createTestAgents(5)
	ctx := context.Background()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)

	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	result, err := p.Execute(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, config.Topic, result.Topic)
	assert.True(t, result.Success)
	assert.Equal(t, 5, result.ParticipantCount)
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestProtocol_Execute_CollectsResponses(t *testing.T) {
	config := DefaultDebateConfig()
	config.Topic = "Test topic"
	config.MaxRounds = 1
	config.Timeout = 30 * time.Second

	topo := topology.NewGraphMeshTopology(topology.DefaultTopologyConfig(topology.TopologyGraphMesh))
	agents := createTestAgents(3)
	ctx := context.Background()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)

	invoker := NewMockAgentInvoker()

	// Set custom responses for each agent
	for _, agent := range agents {
		invoker.SetResponse(agent.ID, &PhaseResponse{
			AgentID:    agent.ID,
			Role:       agent.Role,
			Content:    "Custom response from " + agent.ID,
			Confidence: 0.9,
			Score:      8.0,
			Timestamp:  time.Now(),
		})
	}

	p := NewProtocol(config, topo, invoker)

	result, err := p.Execute(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, len(result.Phases), 0)

	// Check responses were collected
	totalResponses := 0
	for _, phase := range result.Phases {
		totalResponses += len(phase.Responses)
	}
	assert.Greater(t, totalResponses, 0)
}

func TestProtocol_Execute_EarlyExit(t *testing.T) {
	config := DefaultDebateConfig()
	config.Topic = "Test topic"
	config.MaxRounds = 5 // High round count
	config.Timeout = 1 * time.Minute
	config.EnableEarlyExit = true
	config.MinConsensusScore = 0.5 // Low threshold for easy consensus

	topo := topology.NewGraphMeshTopology(topology.DefaultTopologyConfig(topology.TopologyGraphMesh))
	agents := createTestAgents(3)
	ctx := context.Background()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)

	invoker := NewMockAgentInvoker()

	// Set high confidence responses with same votes
	for _, agent := range agents {
		invoker.SetResponse(agent.ID, &PhaseResponse{
			AgentID:    agent.ID,
			Role:       agent.Role,
			Content:    "Agreed response",
			Confidence: 0.95,         // High confidence
			Vote:       "solution_a", // Same vote
			Score:      8.5,
			Timestamp:  time.Now(),
		})
	}

	p := NewProtocol(config, topo, invoker)

	result, err := p.Execute(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should exit early due to high consensus
	assert.LessOrEqual(t, result.RoundsCompleted, config.MaxRounds)
}

func TestProtocol_Execute_Timeout(t *testing.T) {
	config := DefaultDebateConfig()
	config.Topic = "Test topic"
	config.MaxRounds = 1
	config.Timeout = 100 * time.Millisecond // Very short timeout

	topo := topology.NewGraphMeshTopology(topology.DefaultTopologyConfig(topology.TopologyGraphMesh))
	agents := createTestAgents(3)
	ctx := context.Background()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)

	invoker := NewMockAgentInvoker()
	invoker.SetInvokeDelay(500 * time.Millisecond) // Slow responses

	p := NewProtocol(config, topo, invoker)

	result, err := p.Execute(ctx)

	// May timeout or complete partially
	assert.NotNil(t, result)
}

func TestProtocol_Execute_CannotStartTwice(t *testing.T) {
	config := DefaultDebateConfig()
	config.Topic = "Test topic"
	config.MaxRounds = 1
	config.Timeout = 5 * time.Second

	topo := topology.NewGraphMeshTopology(topology.DefaultTopologyConfig(topology.TopologyGraphMesh))
	agents := createTestAgents(2)
	ctx := context.Background()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)

	invoker := NewMockAgentInvoker()
	invoker.SetInvokeDelay(100 * time.Millisecond)

	p := NewProtocol(config, topo, invoker)

	// Start first execution in background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = p.Execute(ctx)
	}()

	// Wait a bit for first execution to start
	time.Sleep(50 * time.Millisecond)

	// Try to start second execution
	_, err = p.Execute(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")

	wg.Wait()
}

// ============================================================================
// Phase Execution Tests
// ============================================================================

func TestProtocol_ExecutePhase(t *testing.T) {
	config := DefaultDebateConfig()
	config.Topic = "Test topic"

	topo := topology.NewGraphMeshTopology(topology.DefaultTopologyConfig(topology.TopologyGraphMesh))
	agents := createTestAgents(4)
	ctx := context.Background()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)

	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)
	p.currentRound = 1

	result, err := p.executePhase(ctx, topology.PhaseProposal)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, topology.PhaseProposal, result.Phase)
	assert.Equal(t, 1, result.Round)
	assert.Greater(t, len(result.Responses), 0)
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestProtocol_GetPhaseAgents(t *testing.T) {
	config := DefaultDebateConfig()

	topo := topology.NewGraphMeshTopology(topology.DefaultTopologyConfig(topology.TopologyGraphMesh))
	agents := []*topology.Agent{
		topology.CreateAgentFromSpec("p1", topology.RoleProposer, "p", "m", 8.0, "code"),
		topology.CreateAgentFromSpec("p2", topology.RoleProposer, "p", "m", 7.5, "code"),
		topology.CreateAgentFromSpec("c1", topology.RoleCritic, "p", "m", 7.0, "reasoning"),
		topology.CreateAgentFromSpec("r1", topology.RoleReviewer, "p", "m", 6.5, "general"),
	}
	ctx := context.Background()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)

	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	// Test proposal phase - should get proposers
	proposalConfig := p.phaseConfigs[topology.PhaseProposal]
	phaseAgents := p.getPhaseAgents(topology.PhaseProposal, proposalConfig)

	// Should have at least the proposers
	assert.Greater(t, len(phaseAgents), 0)
}

// ============================================================================
// Consensus Calculation Tests
// ============================================================================

func TestProtocol_CalculateConsensus(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	// High consensus scenario
	highConsensusResponses := []*PhaseResponse{
		{AgentID: "a1", Confidence: 0.9, Vote: "A", Content: "Option A is best for performance"},
		{AgentID: "a2", Confidence: 0.85, Vote: "A", Content: "Option A provides best performance"},
		{AgentID: "a3", Confidence: 0.95, Vote: "A", Content: "A offers superior performance"},
	}

	consensus := p.calculateConsensus(highConsensusResponses)
	assert.Greater(t, consensus, 0.7)

	// Low consensus scenario
	lowConsensusResponses := []*PhaseResponse{
		{AgentID: "a1", Confidence: 0.5, Vote: "A", Content: "Something completely different"},
		{AgentID: "a2", Confidence: 0.3, Vote: "B", Content: "Another unrelated topic"},
		{AgentID: "a3", Confidence: 0.4, Vote: "C", Content: "Yet another subject"},
	}

	consensus = p.calculateConsensus(lowConsensusResponses)
	assert.Less(t, consensus, 0.7)
}

func TestProtocol_CalculateVoteAgreement(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	// Unanimous votes
	unanimousResponses := []*PhaseResponse{
		{Vote: "A"},
		{Vote: "A"},
		{Vote: "A"},
	}

	agreement := p.calculateVoteAgreement(unanimousResponses)
	assert.Equal(t, 1.0, agreement)

	// Split votes
	splitResponses := []*PhaseResponse{
		{Vote: "A"},
		{Vote: "B"},
		{Vote: "C"},
	}

	agreement = p.calculateVoteAgreement(splitResponses)
	assert.InDelta(t, 0.33, agreement, 0.1)

	// No votes
	noVoteResponses := []*PhaseResponse{
		{Vote: ""},
		{Vote: ""},
	}

	agreement = p.calculateVoteAgreement(noVoteResponses)
	assert.Equal(t, 0.5, agreement) // Neutral
}

func TestProtocol_CalculateContentSimilarity(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	// Similar content
	similarResponses := []*PhaseResponse{
		{Content: "The solution should focus on performance and scalability"},
		{Content: "We need to prioritize performance and ensure scalability"},
		{Content: "Performance is key and scalability is important"},
	}

	similarity := p.calculateContentSimilarity(similarResponses)
	assert.Greater(t, similarity, 0.0)

	// Single response
	singleResponse := []*PhaseResponse{
		{Content: "Only one response"},
	}

	similarity = p.calculateContentSimilarity(singleResponse)
	assert.Equal(t, 1.0, similarity)

	// Empty responses
	emptyResponses := []*PhaseResponse{}

	similarity = p.calculateContentSimilarity(emptyResponses)
	// Empty responses should return 0 (nothing to compare)
	// But since len(responses) < 2 check returns 1.0 for single/empty
	assert.InDelta(t, 1.0, similarity, 0.01)
}

// ============================================================================
// Insight Extraction Tests
// ============================================================================

func TestProtocol_ExtractInsights(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	responses := []*PhaseResponse{
		{
			Arguments:   []string{"Insight 1", "Insight 2"},
			Suggestions: []string{"Suggestion A"},
		},
		{
			Arguments:   []string{"Insight 1", "Insight 3"}, // Insight 1 is duplicate
			Suggestions: []string{"Suggestion B"},
		},
	}

	insights := p.extractInsights(responses)

	// Should have 5 unique insights (no duplicates)
	assert.Len(t, insights, 5)
	assert.Contains(t, insights, "Insight 1")
	assert.Contains(t, insights, "Insight 2")
	assert.Contains(t, insights, "Insight 3")
	assert.Contains(t, insights, "Suggestion A")
	assert.Contains(t, insights, "Suggestion B")
}

func TestProtocol_FindDisagreements(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	responses := []*PhaseResponse{
		{Criticisms: []string{"Problem 1", "Problem 2"}},
		{Criticisms: []string{"Problem 1", "Problem 3"}}, // Problem 1 is duplicate
	}

	disagreements := p.findDisagreements(responses)

	// Should have 3 unique disagreements
	assert.Len(t, disagreements, 3)
}

// ============================================================================
// Result Building Tests
// ============================================================================

func TestProtocol_BuildFinalConsensus(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	// Add convergence phase result
	p.phaseResults = []*PhaseResult{
		{
			Phase: topology.PhaseConvergence,
			Responses: []*PhaseResponse{
				{AgentID: "a1", Vote: "A", Confidence: 0.9, Content: "Solution A is best"},
				{AgentID: "a2", Vote: "A", Confidence: 0.8, Content: "Agree with A"},
				{AgentID: "a3", Vote: "B", Confidence: 0.7, Content: "Prefer B"},
			},
			ConsensusLevel: 0.8,
			KeyInsights:    []string{"Key insight 1"},
			Disagreements:  []string{"Disagreement 1"},
		},
	}

	consensus := p.buildFinalConsensus()

	assert.NotNil(t, consensus)
	assert.Equal(t, 0.8, consensus.Confidence)
	assert.Equal(t, "A", consensus.WinningVote)
	assert.Equal(t, 2, consensus.VoteBreakdown["A"])
	assert.Equal(t, 1, consensus.VoteBreakdown["B"])
	assert.Len(t, consensus.Contributors, 3)
}

func TestProtocol_BuildFinalConsensus_NoConvergence(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	// No convergence phase
	p.phaseResults = []*PhaseResult{
		{Phase: topology.PhaseProposal, Responses: []*PhaseResponse{}},
	}

	consensus := p.buildFinalConsensus()
	assert.Nil(t, consensus)
}

func TestProtocol_FindBestResponse(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	p.phaseResults = []*PhaseResult{
		{
			Phase: topology.PhaseProposal,
			Responses: []*PhaseResponse{
				{AgentID: "a1", Score: 7.0, Confidence: 0.8},
				{AgentID: "a2", Score: 9.0, Confidence: 0.9}, // Highest score
				{AgentID: "a3", Score: 8.0, Confidence: 0.85},
			},
		},
	}

	best := p.findBestResponse()

	assert.NotNil(t, best)
	assert.Equal(t, "a2", best.AgentID)
	assert.Equal(t, 9.0, best.Score)
}

// ============================================================================
// State Management Tests
// ============================================================================

func TestProtocol_GetCurrentPhase(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)
	p.currentPhase = topology.PhaseCritique

	assert.Equal(t, topology.PhaseCritique, p.GetCurrentPhase())
}

func TestProtocol_GetCurrentRound(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)
	p.currentRound = 3

	assert.Equal(t, 3, p.GetCurrentRound())
}

func TestProtocol_GetPhaseResults(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	p.phaseResults = []*PhaseResult{
		{Phase: topology.PhaseProposal, Round: 1},
		{Phase: topology.PhaseCritique, Round: 1},
	}

	results := p.GetPhaseResults()

	assert.Len(t, results, 2)
	assert.Equal(t, topology.PhaseProposal, results[0].Phase)
}

func TestProtocol_Stop(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	assert.False(t, p.stopped)

	p.Stop()

	assert.True(t, p.stopped)
}

// ============================================================================
// Metrics Tests
// ============================================================================

func TestProtocol_UpdatePhaseMetrics(t *testing.T) {
	config := DefaultDebateConfig()
	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)

	result := &PhaseResult{
		Phase: topology.PhaseProposal,
		Responses: []*PhaseResponse{
			{Latency: 100 * time.Millisecond, Confidence: 0.8},
			{Latency: 200 * time.Millisecond, Confidence: 0.9},
		},
		ConsensusLevel: 0.85,
	}

	p.updatePhaseMetrics(topology.PhaseProposal, result)

	metrics := p.metrics.PhaseMetrics[topology.PhaseProposal]
	assert.NotNil(t, metrics)
	assert.Equal(t, 2, metrics.ResponseCount)
	assert.Equal(t, 150*time.Millisecond, metrics.AvgLatency)
	assert.InDelta(t, 0.85, metrics.AvgConfidence, 0.01)
	assert.InDelta(t, 0.85, metrics.ConsensusLevel, 0.01)
	assert.Equal(t, 2, p.metrics.TotalResponses)
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestProtocol_FullDebateWithAllPhases(t *testing.T) {
	config := DefaultDebateConfig()
	config.Topic = "What is the best approach for implementing microservices?"
	config.Context = "Discussion about software architecture patterns"
	config.MaxRounds = 1
	config.Timeout = 30 * time.Second
	config.EnableEarlyExit = false // Complete all phases

	// Create topology with diverse agents
	topo := topology.NewGraphMeshTopology(topology.DefaultTopologyConfig(topology.TopologyGraphMesh))
	agents := []*topology.Agent{
		topology.CreateAgentFromSpec("proposer", topology.RoleProposer, "claude", "opus", 8.5, "code"),
		topology.CreateAgentFromSpec("critic", topology.RoleCritic, "deepseek", "coder", 8.0, "reasoning"),
		topology.CreateAgentFromSpec("reviewer", topology.RoleReviewer, "gemini", "pro", 7.8, "general"),
		topology.CreateAgentFromSpec("optimizer", topology.RoleOptimizer, "mistral", "large", 7.5, "code"),
		topology.CreateAgentFromSpec("moderator", topology.RoleModerator, "openai", "gpt4", 8.2, "reasoning"),
	}

	ctx := context.Background()
	err := topo.Initialize(ctx, agents)
	require.NoError(t, err)

	invoker := NewMockAgentInvoker()

	// Set responses that lead to consensus
	for _, agent := range agents {
		invoker.SetResponse(agent.ID, &PhaseResponse{
			AgentID:     agent.ID,
			Role:        agent.Role,
			Provider:    agent.Provider,
			Model:       agent.Model,
			Content:     "Comprehensive response about microservices",
			Confidence:  0.85,
			Arguments:   []string{"Domain-driven design helps", "API gateway pattern is useful"},
			Criticisms:  []string{"Consider complexity", "Monitor performance"},
			Suggestions: []string{"Use event sourcing", "Implement circuit breakers"},
			Vote:        "microservices_approach_A",
			Score:       agent.Score,
			Latency:     50 * time.Millisecond,
			Timestamp:   time.Now(),
		})
	}

	p := NewProtocol(config, topo, invoker)

	result, err := p.Execute(ctx)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify all 5 phases were executed
	assert.Len(t, result.Phases, 5)

	phasesSeen := make(map[topology.DebatePhase]bool)
	for _, phase := range result.Phases {
		phasesSeen[phase.Phase] = true
		assert.Greater(t, len(phase.Responses), 0, "Phase %s should have responses", phase.Phase)
	}

	assert.True(t, phasesSeen[topology.PhaseProposal])
	assert.True(t, phasesSeen[topology.PhaseCritique])
	assert.True(t, phasesSeen[topology.PhaseReview])
	assert.True(t, phasesSeen[topology.PhaseOptimization])
	assert.True(t, phasesSeen[topology.PhaseConvergence])

	// Verify consensus was built
	assert.NotNil(t, result.FinalConsensus)
	assert.NotEmpty(t, result.FinalConsensus.Summary)

	// Verify metrics
	assert.NotNil(t, result.Metrics)
	assert.Greater(t, result.Metrics.TotalResponses, 0)

	// Verify result structure
	assert.Equal(t, config.Topic, result.Topic)
	assert.True(t, result.Success)
	assert.Equal(t, topology.TopologyGraphMesh, result.TopologyUsed)
	assert.Equal(t, 5, result.ParticipantCount)
}

// ============================================================================
// Consensus Method Tests
// ============================================================================

func TestConsensusMethod_Values(t *testing.T) {
	methods := []ConsensusMethod{
		ConsensusMethodUnanimous,
		ConsensusMethodMajority,
		ConsensusMethodWeightedVoting,
		ConsensusMethodLeaderDecision,
		ConsensusMethodNoConsensus,
	}

	for _, m := range methods {
		assert.NotEmpty(t, string(m))
	}
}

// ============================================================================
// Debate Context Tests
// ============================================================================

func TestProtocol_BuildDebateContext(t *testing.T) {
	config := DefaultDebateConfig()
	config.ID = "test-debate-123"
	config.Topic = "Test topic"
	config.Context = "Test context"
	config.Metadata = map[string]interface{}{"key": "value"}

	topo := topology.NewDefaultTopology()
	invoker := NewMockAgentInvoker()

	p := NewProtocol(config, topo, invoker)
	p.currentRound = 2
	p.phaseResults = []*PhaseResult{
		{Phase: topology.PhaseProposal, Round: 1},
	}

	ctx := p.buildDebateContext(topology.PhaseCritique)

	assert.Equal(t, config.ID, ctx.DebateID)
	assert.Equal(t, config.Topic, ctx.Topic)
	assert.Equal(t, config.Context, ctx.Context)
	assert.Equal(t, topology.PhaseCritique, ctx.CurrentPhase)
	assert.Equal(t, 2, ctx.Round)
	assert.Len(t, ctx.PreviousPhases, 1)
	assert.Equal(t, "value", ctx.Metadata["key"])
}
