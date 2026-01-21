package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/debate"
	"dev.helix.agent/internal/debate/agents"
	"dev.helix.agent/internal/debate/knowledge"
	"dev.helix.agent/internal/debate/protocol"
	"dev.helix.agent/internal/debate/topology"
	"dev.helix.agent/internal/models"
)

// =============================================================================
// Extended Orchestrator Tests for Coverage Improvement
// =============================================================================

func TestOrchestrator_buildResponse_WithMetrics(t *testing.T) {
	orch, _ := createTestOrchestrator()

	result := &protocol.DebateResult{
		ID:      "test-debate",
		Topic:   "Test Topic",
		Success: true,
		Phases: []*protocol.PhaseResult{
			{
				Phase: topology.PhaseProposal,
				Round: 1,
				Responses: []*protocol.PhaseResponse{
					{
						AgentID:    "agent-1",
						Provider:   "claude",
						Model:      "claude-3",
						Role:       topology.RoleProposer,
						Content:    "Test response",
						Confidence: 0.85,
						Score:      8.5,
						Latency:    100 * time.Millisecond,
					},
				},
				ConsensusLevel: 0.8,
				KeyInsights:    []string{"Insight 1"},
				Duration:       1 * time.Second,
			},
		},
		FinalConsensus: &protocol.ConsensusResult{
			Summary:       "Final consensus summary",
			Confidence:    0.9,
			KeyPoints:     []string{"Key point 1", "Key point 2"},
			Dissents:      []string{"Minor dissent"},
			VoteBreakdown: map[string]int{"agree": 3, "disagree": 1},
			WinningVote:   "agree",
			Method:        protocol.ConsensusMethodWeightedVoting,
		},
		Metrics: &protocol.DebateMetrics{
			TotalResponses:     5,
			AvgLatency:         150 * time.Millisecond,
			AvgConfidence:      0.82,
			ConsensusScore:     0.85,
			AgentParticipation: map[string]int{"claude": 2, "deepseek": 3},
		},
		Duration: 5 * time.Minute,
		Metadata: map[string]interface{}{"key": "value"},
	}

	teamAgents := []*agents.SpecializedAgent{
		agents.NewSpecializedAgent("Claude Agent", "claude", "claude-3", agents.DomainReasoning),
	}
	teamAgents[0].Score = 8.5

	response := orch.buildResponse(result, teamAgents, 3, 2)

	assert.Equal(t, "test-debate", response.ID)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Consensus)
	assert.Equal(t, "Final consensus summary", response.Consensus.Summary)
	assert.Equal(t, 0.9, response.Consensus.Confidence)
	assert.Len(t, response.Consensus.KeyPoints, 2)
	assert.Equal(t, "weighted_voting", response.Consensus.Method)
	assert.Len(t, response.Phases, 1)
	assert.Len(t, response.Participants, 1)
	assert.Equal(t, 3, response.LessonsLearned)
	assert.Equal(t, 2, response.PatternsDetected)
	assert.NotNil(t, response.Metrics)
	assert.Equal(t, 5, response.Metrics.TotalResponses)
	assert.Equal(t, 0.82, response.Metrics.AvgConfidence)
}

func TestOrchestrator_buildResponse_WithoutConsensus(t *testing.T) {
	orch, _ := createTestOrchestrator()

	result := &protocol.DebateResult{
		ID:             "test-debate",
		Topic:          "Test Topic",
		Success:        false,
		Phases:         []*protocol.PhaseResult{},
		FinalConsensus: nil, // No consensus
		Duration:       2 * time.Minute,
	}

	response := orch.buildResponse(result, []*agents.SpecializedAgent{}, 0, 0)

	assert.Nil(t, response.Consensus)
	assert.False(t, response.Success)
}

func TestOrchestrator_buildResponse_WithoutMetrics(t *testing.T) {
	orch, _ := createTestOrchestrator()

	result := &protocol.DebateResult{
		ID:      "test-debate",
		Topic:   "Test Topic",
		Success: true,
		Metrics: nil, // No metrics
	}

	response := orch.buildResponse(result, []*agents.SpecializedAgent{}, 0, 0)

	assert.Nil(t, response.Metrics)
}

func TestActiveDebate_Structure(t *testing.T) {
	ad := &ActiveDebate{
		ID:        "debate-1",
		StartTime: time.Now(),
		Status:    DebateStatusRunning,
	}

	assert.Equal(t, "debate-1", ad.ID)
	assert.Equal(t, DebateStatusRunning, ad.Status)
	assert.False(t, ad.StartTime.IsZero())
}

func TestDebateRequest_AllFields(t *testing.T) {
	enableLearning := true
	request := &DebateRequest{
		ID:                 "req-1",
		Topic:              "Test Topic",
		Context:            "Test context",
		Requirements:       []string{"Req1", "Req2"},
		MaxRounds:          5,
		Timeout:            10 * time.Minute,
		TopologyType:       topology.TopologyGraphMesh,
		MinConsensus:       0.8,
		PreferredProviders: []string{"claude", "deepseek"},
		PreferredDomain:    agents.DomainCode,
		EnableLearning:     &enableLearning,
		Metadata:           map[string]interface{}{"key": "value"},
	}

	assert.Equal(t, "req-1", request.ID)
	assert.Equal(t, "Test Topic", request.Topic)
	assert.Len(t, request.Requirements, 2)
	assert.Equal(t, 5, request.MaxRounds)
	assert.True(t, *request.EnableLearning)
}

func TestConsensusResponse_AllFields(t *testing.T) {
	cr := &ConsensusResponse{
		Summary:       "Summary text",
		Confidence:    0.85,
		KeyPoints:     []string{"Point 1", "Point 2"},
		Dissents:      []string{"Dissent 1"},
		VoteBreakdown: map[string]int{"yes": 3, "no": 1},
		WinningVote:   "yes",
		Method:        "weighted",
	}

	assert.Equal(t, "Summary text", cr.Summary)
	assert.Equal(t, 0.85, cr.Confidence)
	assert.Len(t, cr.KeyPoints, 2)
	assert.Len(t, cr.Dissents, 1)
	assert.Equal(t, 3, cr.VoteBreakdown["yes"])
	assert.Equal(t, "yes", cr.WinningVote)
}

func TestPhaseResponse_AllFields(t *testing.T) {
	pr := &PhaseResponse{
		Phase:          "proposal",
		Round:          1,
		Responses:      []*AgentResponse{{AgentID: "a1"}},
		ConsensusLevel: 0.75,
		KeyInsights:    []string{"Insight 1"},
		Duration:       30 * time.Second,
	}

	assert.Equal(t, "proposal", pr.Phase)
	assert.Equal(t, 1, pr.Round)
	assert.Len(t, pr.Responses, 1)
	assert.Equal(t, 0.75, pr.ConsensusLevel)
}

func TestAgentResponse_AllFields(t *testing.T) {
	ar := &AgentResponse{
		AgentID:    "agent-1",
		Provider:   "claude",
		Model:      "claude-3",
		Role:       "proposer",
		Content:    "Response content",
		Confidence: 0.9,
		Score:      8.5,
		Latency:    100 * time.Millisecond,
	}

	assert.Equal(t, "agent-1", ar.AgentID)
	assert.Equal(t, "claude", ar.Provider)
	assert.Equal(t, 0.9, ar.Confidence)
}

func TestParticipantInfo_AllFields(t *testing.T) {
	pi := &ParticipantInfo{
		AgentID:  "agent-1",
		Name:     "Test Agent",
		Provider: "claude",
		Model:    "claude-3",
		Role:     "proposer",
		Domain:   agents.DomainReasoning,
		Score:    8.5,
	}

	assert.Equal(t, "agent-1", pi.AgentID)
	assert.Equal(t, agents.DomainReasoning, pi.Domain)
}

func TestDebateMetrics_AllFields(t *testing.T) {
	dm := &DebateMetrics{
		TotalResponses:    10,
		AvgLatency:        200 * time.Millisecond,
		AvgConfidence:     0.82,
		ConsensusScore:    0.85,
		QualityScore:      0.88,
		ProviderBreakdown: map[string]int{"claude": 5, "deepseek": 5},
	}

	assert.Equal(t, 10, dm.TotalResponses)
	assert.Equal(t, 0.82, dm.AvgConfidence)
	assert.Len(t, dm.ProviderBreakdown, 2)
}

// =============================================================================
// ProviderInvoker Extended Tests
// =============================================================================

func TestProviderInvoker_Invoke_AllPhases(t *testing.T) {
	registry := newMockProviderRegistry()
	registry.AddProvider("claude", newMockLLMProvider("claude"))
	invoker := NewProviderInvoker(registry)

	phases := []topology.DebatePhase{
		topology.PhaseProposal,
		topology.PhaseCritique,
		topology.PhaseReview,
		topology.PhaseOptimization,
		topology.PhaseConvergence,
	}

	for _, phase := range phases {
		t.Run(string(phase), func(t *testing.T) {
			agent := &topology.Agent{
				ID:       "agent-1",
				Provider: "claude",
				Model:    "claude-3",
				Role:     topology.RoleProposer,
				Score:    9.0,
			}

			debateCtx := protocol.DebateContext{
				Topic:        "Test Topic",
				CurrentPhase: phase,
				Round:        1,
			}

			ctx := context.Background()
			response, err := invoker.Invoke(ctx, agent, "Analyze this topic", debateCtx)
			require.NoError(t, err)
			assert.NotNil(t, response)
		})
	}
}

func TestProviderInvoker_Invoke_WithCriticisms(t *testing.T) {
	registry := newMockProviderRegistry()
	registry.AddProvider("claude", newMockLLMProvider("claude"))
	invoker := NewProviderInvoker(registry)

	agent := &topology.Agent{
		ID:       "agent-1",
		Provider: "claude",
		Model:    "claude-3",
		Role:     topology.RoleCritic,
		Score:    9.0,
	}

	debateCtx := protocol.DebateContext{
		Topic:        "Test Topic",
		CurrentPhase: topology.PhaseCritique,
		Round:        1,
	}

	ctx := context.Background()
	response, err := invoker.Invoke(ctx, agent, "Critique this", debateCtx)
	require.NoError(t, err)
	assert.NotNil(t, response)
	// Criticisms should be extracted from the content
	assert.NotNil(t, response.Criticisms)
}

func TestProviderInvoker_Invoke_WithSuggestions(t *testing.T) {
	registry := newMockProviderRegistry()
	registry.AddProvider("claude", newMockLLMProvider("claude"))
	invoker := NewProviderInvoker(registry)

	agent := &topology.Agent{
		ID:       "agent-1",
		Provider: "claude",
		Model:    "claude-3",
		Role:     topology.RoleOptimizer,
		Score:    9.0,
	}

	debateCtx := protocol.DebateContext{
		Topic:        "Test Topic",
		CurrentPhase: topology.PhaseOptimization,
		Round:        1,
	}

	ctx := context.Background()
	response, err := invoker.Invoke(ctx, agent, "Optimize this", debateCtx)
	require.NoError(t, err)
	assert.NotNil(t, response)
	// Suggestions should be extracted
	assert.NotNil(t, response.Suggestions)
}

// =============================================================================
// System Prompt Building Extended Tests
// =============================================================================

func TestBuildSystemPrompt_UnknownRole(t *testing.T) {
	agent := &topology.Agent{
		ID:   "agent-1",
		Role: topology.AgentRole("unknown_role"),
	}

	debateCtx := protocol.DebateContext{
		CurrentPhase: topology.PhaseProposal,
	}

	prompt := buildSystemPrompt(agent, debateCtx)
	assert.Contains(t, prompt, "thoughtful analysis")
}

func TestBuildSystemPrompt_UnknownPhase(t *testing.T) {
	agent := &topology.Agent{
		ID:   "agent-1",
		Role: topology.RoleProposer,
	}

	debateCtx := protocol.DebateContext{
		CurrentPhase: topology.DebatePhase("unknown_phase"),
	}

	prompt := buildSystemPrompt(agent, debateCtx)
	assert.Contains(t, prompt, "analysis and insights")
}

func TestBuildFullPrompt_EmptyContext(t *testing.T) {
	debateCtx := protocol.DebateContext{
		Topic:   "Test Topic",
		Context: "", // Empty context
		Round:   1,
	}

	prompt := buildFullPrompt("Test prompt", debateCtx)
	assert.Contains(t, prompt, "Test Topic")
	assert.NotContains(t, prompt, "Context:")
}

func TestBuildFullPrompt_WithMultiplePreviousPhases(t *testing.T) {
	debateCtx := protocol.DebateContext{
		Topic: "Test Topic",
		Round: 3,
		PreviousPhases: []*protocol.PhaseResult{
			{
				Phase:       topology.PhaseProposal,
				KeyInsights: []string{"Proposal insight 1"},
			},
			{
				Phase:       topology.PhaseCritique,
				KeyInsights: []string{"Critique insight 1", "Critique insight 2"},
			},
		},
	}

	prompt := buildFullPrompt("Continue", debateCtx)
	assert.Contains(t, prompt, "Previous Discussion")
	assert.Contains(t, prompt, "proposal")
	assert.Contains(t, prompt, "critique")
}

func TestBuildFullPrompt_EmptyPreviousPhaseInsights(t *testing.T) {
	debateCtx := protocol.DebateContext{
		Topic: "Test Topic",
		Round: 2,
		PreviousPhases: []*protocol.PhaseResult{
			{
				Phase:       topology.PhaseProposal,
				KeyInsights: []string{}, // Empty insights
			},
		},
	}

	prompt := buildFullPrompt("Continue", debateCtx)
	// Should still work without errors
	assert.Contains(t, prompt, "Test Topic")
}

// =============================================================================
// Confidence Calculation Extended Tests
// =============================================================================

func TestCalculateConfidence_VeryLongContent(t *testing.T) {
	// Test with very long content to ensure capping works
	response := &models.LLMResponse{
		Content:      string(make([]byte, 10000)), // 10000 chars
		FinishReason: "stop",
	}

	conf := calculateConfidence(response)
	assert.LessOrEqual(t, conf, 0.95) // Should be capped
}

func TestCalculateConfidence_VeryShortContent(t *testing.T) {
	response := &models.LLMResponse{
		Content:      "Hi", // Very short
		FinishReason: "length",
	}

	conf := calculateConfidence(response)
	assert.GreaterOrEqual(t, conf, 0.3) // Should have minimum
}

func TestCalculateConfidence_NonStopFinishReason(t *testing.T) {
	response := &models.LLMResponse{
		Content:      string(make([]byte, 500)),
		FinishReason: "length", // Not "stop"
	}

	confWithStop := calculateConfidence(&models.LLMResponse{
		Content:      string(make([]byte, 500)),
		FinishReason: "stop",
	})

	confWithLength := calculateConfidence(response)

	// Stop should give slightly higher confidence
	assert.Greater(t, confWithStop, confWithLength)
}

// =============================================================================
// Argument/Criticism/Suggestion Extraction Tests
// =============================================================================

func TestExtractArguments_WithNumberedList(t *testing.T) {
	content := `Analysis:
1. First numbered point with content
2. Second numbered point
3. Third point here`

	args := extractArguments(content)
	assert.NotEmpty(t, args)
}

func TestExtractArguments_MixedBullets(t *testing.T) {
	content := `Key findings:
- Dash bullet point here
* Asterisk bullet point
1. Numbered item
2. Another numbered item`

	args := extractArguments(content)
	assert.NotEmpty(t, args)
	assert.LessOrEqual(t, len(args), 5) // Max 5
}

func TestExtractArguments_ShortLines(t *testing.T) {
	// Short lines (less than 10 chars) should not be included
	content := `- short
- another
- This is a longer point that should be included`

	args := extractArguments(content)
	// Only the longer one should match
	assert.LessOrEqual(t, len(args), 1)
}

func TestExtractArguments_MoreThanFive(t *testing.T) {
	content := `Points:
- First long point here with content
- Second long point here with content
- Third long point here with content
- Fourth long point here with content
- Fifth long point here with content
- Sixth long point here with content
- Seventh long point here with content`

	args := extractArguments(content)
	assert.Equal(t, 5, len(args)) // Should cap at 5
}

// =============================================================================
// Helper Functions Extended Tests
// =============================================================================

func TestSplitLines_EmptyString(t *testing.T) {
	lines := splitLines("")
	assert.Empty(t, lines)
}

func TestSplitLines_TrailingNewline(t *testing.T) {
	lines := splitLines("line1\nline2\n")
	assert.Len(t, lines, 2) // Trailing newline doesn't add empty string
	assert.Equal(t, "line1", lines[0])
	assert.Equal(t, "line2", lines[1])
}

func TestTrimLine_EmptyString(t *testing.T) {
	result := trimLine("")
	assert.Empty(t, result)
}

func TestTrimLine_OnlyWhitespace(t *testing.T) {
	result := trimLine("   \t  ")
	assert.Empty(t, result)
}

func TestTrimLine_ComplexContent(t *testing.T) {
	result := trimLine("  - * Content with multiple prefixes  \n")
	assert.Equal(t, "Content with multiple prefixes", result)
}

// =============================================================================
// Active Debate Management Extended Tests
// =============================================================================

func TestOrchestrator_ActiveDebateTracking(t *testing.T) {
	orch, _ := createTestOrchestrator()

	// Initially no active debates
	debates := orch.GetActiveDebates()
	assert.Empty(t, debates)

	// Manually add a debate for testing
	orch.mu.Lock()
	orch.activeDebates["test-1"] = &ActiveDebate{
		ID:     "test-1",
		Status: DebateStatusRunning,
	}
	orch.mu.Unlock()

	// Now should have one
	debates = orch.GetActiveDebates()
	assert.Len(t, debates, 1)

	// Check status
	status, found := orch.GetDebateStatus("test-1")
	assert.True(t, found)
	assert.Equal(t, DebateStatusRunning, status)

	// Cancel it
	err := orch.CancelDebate("test-1")
	assert.NoError(t, err)

	status, found = orch.GetDebateStatus("test-1")
	assert.True(t, found)
	assert.Equal(t, DebateStatusCancelled, status)
}

// =============================================================================
// Orchestrator Configuration Tests
// =============================================================================

func TestOrchestratorConfig_AllFields(t *testing.T) {
	config := OrchestratorConfig{
		DefaultMaxRounds:          5,
		DefaultTimeout:            10 * time.Minute,
		DefaultTopology:           topology.TopologyChain,
		DefaultMinConsensus:       0.8,
		MinAgentsPerDebate:        4,
		MaxAgentsPerDebate:        15,
		EnableAgentDiversity:      false,
		EnableLearning:            false,
		EnableCrossDebateLearning: false,
		MinConsensusForLesson:     0.9,
	}

	assert.Equal(t, 5, config.DefaultMaxRounds)
	assert.Equal(t, topology.TopologyChain, config.DefaultTopology)
	assert.False(t, config.EnableLearning)
}

// =============================================================================
// Knowledge Integration Tests
// =============================================================================

func TestOrchestrator_KnowledgeIntegration(t *testing.T) {
	registry := newMockProviderRegistry()
	registry.AddProvider("claude", newMockLLMProvider("claude"))

	lessonBankConfig := debate.DefaultLessonBankConfig()
	lessonBankConfig.EnableSemanticSearch = false
	lessonBank := debate.NewLessonBank(lessonBankConfig, nil, nil)

	config := DefaultOrchestratorConfig()
	config.EnableLearning = true
	config.EnableCrossDebateLearning = true

	orch := NewOrchestrator(registry, lessonBank, config)

	// Verify learning integration is set up
	assert.NotNil(t, orch.learningIntegration)
	assert.NotNil(t, orch.crossDebateLearner)

	// Get knowledge repo
	repo := orch.GetKnowledgeRepository()
	assert.NotNil(t, repo)

	// Get recommendations (should work even without prior debates)
	ctx := context.Background()
	recs, err := orch.GetRecommendations(ctx, "security testing", agents.DomainSecurity)
	require.NoError(t, err)
	assert.NotNil(t, recs)
}

// =============================================================================
// Integration Statistics Tests
// =============================================================================

func TestOrchestrator_StatisticsWithData(t *testing.T) {
	orch, _ := createTestOrchestrator()
	ctx := context.Background()

	// Add some test data to knowledge repo
	repo, ok := orch.GetKnowledgeRepository().(*knowledge.DefaultRepository)
	if ok {
		repo.RecordPattern(ctx, &knowledge.DebatePattern{
			Name:        "Test Pattern",
			PatternType: knowledge.PatternTypeConsensusBuilding,
			Domain:      agents.DomainCode,
		})
	}

	stats, err := orch.GetStatistics(ctx)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 3, stats.RegisteredAgents) // 3 from setup
}
