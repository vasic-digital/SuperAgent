package orchestrator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/debate"
	"dev.helix.agent/internal/debate/agents"
	"dev.helix.agent/internal/debate/protocol"
	"dev.helix.agent/internal/debate/topology"
	"dev.helix.agent/internal/debate/voting"
	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/models"
)

// =============================================================================
// Mock Provider Registry
// =============================================================================

type mockProviderRegistry struct {
	providers map[string]*mockLLMProvider
}

func newMockProviderRegistry() *mockProviderRegistry {
	return &mockProviderRegistry{
		providers: make(map[string]*mockLLMProvider),
	}
}

func (r *mockProviderRegistry) GetProvider(name string) (llm.LLMProvider, error) {
	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

func (r *mockProviderRegistry) GetAvailableProviders() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

func (r *mockProviderRegistry) AddProvider(name string, provider *mockLLMProvider) {
	r.providers[name] = provider
}

// =============================================================================
// Mock LLM Provider
// =============================================================================

type mockLLMProvider struct {
	name      string
	responses map[string]string
	callCount int
}

func newMockLLMProvider(name string) *mockLLMProvider {
	return &mockLLMProvider{
		name:      name,
		responses: make(map[string]string),
	}
}

func (p *mockLLMProvider) Complete(ctx context.Context, request *models.LLMRequest) (*models.LLMResponse, error) {
	p.callCount++

	// Generate a mock response based on the model
	content := "This is a thoughtful analysis of the topic. Key points:\n"
	content += "- First point: Important consideration\n"
	content += "- Second point: Another key insight\n"
	content += "- Third point: Supporting evidence\n"
	content += "\nConfidence: 85%"

	return &models.LLMResponse{
		Content:      content,
		ProviderName: p.name,
		TokensUsed:   150,
		FinishReason: "stop",
	}, nil
}

func (p *mockLLMProvider) CompleteStream(ctx context.Context, request *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	ch := make(chan *models.LLMResponse, 1)
	go func() {
		response, _ := p.Complete(ctx, request)
		ch <- response
		close(ch)
	}()
	return ch, nil
}

func (p *mockLLMProvider) HealthCheck() error {
	return nil
}

func (p *mockLLMProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportsStreaming: true,
		SupportsTools:     true,
	}
}

func (p *mockLLMProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	return true, nil
}

// Ensure mockLLMProvider implements LLMProvider
var _ llm.LLMProvider = (*mockLLMProvider)(nil)

// =============================================================================
// Test Helpers
// =============================================================================

func createTestOrchestrator() (*Orchestrator, *mockProviderRegistry) {
	registry := newMockProviderRegistry()
	registry.AddProvider("claude", newMockLLMProvider("claude"))
	registry.AddProvider("deepseek", newMockLLMProvider("deepseek"))
	registry.AddProvider("gemini", newMockLLMProvider("gemini"))

	lessonBankConfig := debate.DefaultLessonBankConfig()
	lessonBankConfig.EnableSemanticSearch = false
	lessonBank := debate.NewLessonBank(lessonBankConfig, nil, nil)

	config := DefaultOrchestratorConfig()
	config.MinAgentsPerDebate = 3

	orch := NewOrchestrator(registry, lessonBank, config)

	// Register providers with agents
	orch.RegisterProvider("claude", "claude-3", 9.0)
	orch.RegisterProvider("deepseek", "deepseek-coder", 8.5)
	orch.RegisterProvider("gemini", "gemini-pro", 8.0)

	return orch, registry
}

// =============================================================================
// Orchestrator Creation Tests
// =============================================================================

func TestNewOrchestrator(t *testing.T) {
	orch, _ := createTestOrchestrator()

	assert.NotNil(t, orch)
	assert.NotNil(t, orch.agentFactory)
	assert.NotNil(t, orch.agentPool)
	assert.NotNil(t, orch.teamBuilder)
	assert.NotNil(t, orch.knowledgeRepo)
	assert.NotNil(t, orch.learningIntegration)
	assert.NotNil(t, orch.crossDebateLearner)
	assert.NotNil(t, orch.votingSystem)
}

func TestDefaultOrchestratorConfig(t *testing.T) {
	config := DefaultOrchestratorConfig()

	assert.Equal(t, 3, config.DefaultMaxRounds)
	assert.Equal(t, 5*time.Minute, config.DefaultTimeout)
	assert.Equal(t, topology.TopologyGraphMesh, config.DefaultTopology)
	assert.Equal(t, 0.75, config.DefaultMinConsensus)
	assert.Equal(t, 3, config.MinAgentsPerDebate)
	assert.Equal(t, 10, config.MaxAgentsPerDebate)
	assert.True(t, config.EnableAgentDiversity)
	assert.True(t, config.EnableLearning)
	assert.True(t, config.EnableCrossDebateLearning)
	assert.Equal(t, 0.7, config.MinConsensusForLesson)
	assert.Equal(t, voting.VotingMethodWeighted, config.VotingMethod)
	assert.True(t, config.EnableConfidenceWeighting)
}

// =============================================================================
// Provider Registration Tests
// =============================================================================

func TestOrchestrator_RegisterProvider(t *testing.T) {
	orch, _ := createTestOrchestrator()

	err := orch.RegisterProvider("test-provider", "test-model", 7.5)
	require.NoError(t, err)

	// Check agent was added to pool
	assert.Greater(t, orch.agentPool.Size(), 3) // We started with 3

	// Check score was recorded
	score, ok := orch.verifierScores["test-provider/test-model"]
	assert.True(t, ok)
	assert.Equal(t, 7.5, score)
}

func TestOrchestrator_SetVerifierScores(t *testing.T) {
	orch, _ := createTestOrchestrator()

	scores := map[string]float64{
		"claude/claude-3":      9.5,
		"deepseek/deepseek-v2": 8.8,
	}

	orch.SetVerifierScores(scores)

	assert.Equal(t, 9.5, orch.verifierScores["claude/claude-3"])
	assert.Equal(t, 8.8, orch.verifierScores["deepseek/deepseek-v2"])
}

// =============================================================================
// Domain Inference Tests
// =============================================================================

func TestOrchestrator_inferDomainFromProvider(t *testing.T) {
	orch, _ := createTestOrchestrator()

	testCases := []struct {
		provider string
		expected agents.Domain
	}{
		{"deepseek", agents.DomainCode},
		{"claude", agents.DomainReasoning},
		{"gemini", agents.DomainReasoning},
		{"mistral", agents.DomainOptimization},
		{"unknown", agents.DomainGeneral},
	}

	for _, tc := range testCases {
		t.Run(tc.provider, func(t *testing.T) {
			domain := orch.inferDomainFromProvider(tc.provider)
			assert.Equal(t, tc.expected, domain)
		})
	}
}

// =============================================================================
// Active Debate Management Tests
// =============================================================================

func TestOrchestrator_GetActiveDebates(t *testing.T) {
	orch, _ := createTestOrchestrator()

	// Initially empty
	debates := orch.GetActiveDebates()
	assert.Empty(t, debates)
}

func TestOrchestrator_GetDebateStatus_NotFound(t *testing.T) {
	orch, _ := createTestOrchestrator()

	status, found := orch.GetDebateStatus("non-existent")
	assert.False(t, found)
	assert.Equal(t, DebateStatus(""), status)
}

func TestOrchestrator_CancelDebate_NotFound(t *testing.T) {
	orch, _ := createTestOrchestrator()

	err := orch.CancelDebate("non-existent")
	assert.Error(t, err)
}

// =============================================================================
// Statistics Tests
// =============================================================================

func TestOrchestrator_GetStatistics(t *testing.T) {
	orch, _ := createTestOrchestrator()
	ctx := context.Background()

	stats, err := orch.GetStatistics(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, 0, stats.ActiveDebates)
	assert.Equal(t, 3, stats.RegisteredAgents) // We registered 3
	assert.GreaterOrEqual(t, stats.TotalLessons, 0)
	assert.GreaterOrEqual(t, stats.TotalPatterns, 0)
}

// =============================================================================
// Agent Pool Tests
// =============================================================================

func TestOrchestrator_GetAgentPool(t *testing.T) {
	orch, _ := createTestOrchestrator()

	pool := orch.GetAgentPool()
	assert.NotNil(t, pool)
	assert.Equal(t, 3, pool.Size())
}

// =============================================================================
// Knowledge Repository Tests
// =============================================================================

func TestOrchestrator_GetKnowledgeRepository(t *testing.T) {
	orch, _ := createTestOrchestrator()

	repo := orch.GetKnowledgeRepository()
	assert.NotNil(t, repo)
}

// =============================================================================
// Recommendations Tests
// =============================================================================

func TestOrchestrator_GetRecommendations(t *testing.T) {
	orch, _ := createTestOrchestrator()
	ctx := context.Background()

	recommendations, err := orch.GetRecommendations(ctx, "test topic", agents.DomainCode)
	require.NoError(t, err)
	assert.NotNil(t, recommendations)
}

// =============================================================================
// ProviderInvoker Tests
// =============================================================================

func TestNewProviderInvoker(t *testing.T) {
	registry := newMockProviderRegistry()
	invoker := NewProviderInvoker(registry)

	assert.NotNil(t, invoker)
	assert.Equal(t, registry, invoker.registry)
}

func TestProviderInvoker_Invoke(t *testing.T) {
	registry := newMockProviderRegistry()
	registry.AddProvider("claude", newMockLLMProvider("claude"))

	invoker := NewProviderInvoker(registry)

	agent := &topology.Agent{
		ID:       "agent-1",
		Provider: "claude",
		Model:    "claude-3",
		Role:     topology.RoleProposer,
		Score:    9.0,
	}

	debateCtx := protocol.DebateContext{
		Topic:        "Test Topic",
		Context:      "Test context",
		CurrentPhase: topology.PhaseProposal,
		Round:        1,
	}

	ctx := context.Background()
	response, err := invoker.Invoke(ctx, agent, "Analyze this topic", debateCtx)
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.Equal(t, "agent-1", response.AgentID)
	assert.Equal(t, "claude", response.Provider)
	assert.Equal(t, "claude-3", response.Model)
	assert.NotEmpty(t, response.Content)
	assert.Greater(t, response.Confidence, 0.0)
}

func TestProviderInvoker_Invoke_ProviderNotFound(t *testing.T) {
	registry := newMockProviderRegistry()
	invoker := NewProviderInvoker(registry)

	agent := &topology.Agent{
		ID:       "agent-1",
		Provider: "non-existent",
		Model:    "model",
	}

	debateCtx := protocol.DebateContext{
		Topic: "Test",
	}

	ctx := context.Background()
	_, err := invoker.Invoke(ctx, agent, "test", debateCtx)
	assert.Error(t, err)
}

// =============================================================================
// System Prompt Building Tests
// =============================================================================

func TestBuildSystemPrompt(t *testing.T) {
	agent := &topology.Agent{
		ID:   "agent-1",
		Role: topology.RoleProposer,
	}

	debateCtx := protocol.DebateContext{
		CurrentPhase: topology.PhaseProposal,
	}

	prompt := buildSystemPrompt(agent, debateCtx)

	assert.Contains(t, prompt, "proposer")
	assert.Contains(t, prompt, "proposal")
	assert.Contains(t, prompt, "confidence score")
}

func TestBuildSystemPrompt_AllRoles(t *testing.T) {
	roles := []topology.AgentRole{
		topology.RoleProposer,
		topology.RoleCritic,
		topology.RoleReviewer,
		topology.RoleOptimizer,
		topology.RoleModerator,
		topology.RoleArchitect,
		topology.RoleValidator,
		topology.RoleRedTeam,
		topology.RoleBlueTeam,
	}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			agent := &topology.Agent{Role: role}
			debateCtx := protocol.DebateContext{CurrentPhase: topology.PhaseProposal}
			prompt := buildSystemPrompt(agent, debateCtx)
			assert.NotEmpty(t, prompt)
		})
	}
}

func TestBuildSystemPrompt_AllPhases(t *testing.T) {
	phases := []topology.DebatePhase{
		topology.PhaseProposal,
		topology.PhaseCritique,
		topology.PhaseReview,
		topology.PhaseOptimization,
		topology.PhaseConvergence,
	}

	for _, phase := range phases {
		t.Run(string(phase), func(t *testing.T) {
			agent := &topology.Agent{Role: topology.RoleReviewer}
			debateCtx := protocol.DebateContext{CurrentPhase: phase}
			prompt := buildSystemPrompt(agent, debateCtx)
			assert.NotEmpty(t, prompt)
		})
	}
}

// =============================================================================
// Full Prompt Building Tests
// =============================================================================

func TestBuildFullPrompt(t *testing.T) {
	debateCtx := protocol.DebateContext{
		Topic:   "Test Topic",
		Context: "Additional context",
		Round:   2,
	}

	prompt := buildFullPrompt("Analyze this", debateCtx)

	assert.Contains(t, prompt, "Test Topic")
	assert.Contains(t, prompt, "Additional context")
	assert.Contains(t, prompt, "Round 2")
}

func TestBuildFullPrompt_WithPreviousPhases(t *testing.T) {
	debateCtx := protocol.DebateContext{
		Topic: "Test Topic",
		Round: 2,
		PreviousPhases: []*protocol.PhaseResult{
			{
				Phase:       topology.PhaseProposal,
				KeyInsights: []string{"Key insight 1", "Key insight 2"},
			},
		},
	}

	prompt := buildFullPrompt("Continue", debateCtx)

	assert.Contains(t, prompt, "Previous Discussion")
	assert.Contains(t, prompt, "proposal")
}

// =============================================================================
// Confidence Calculation Tests
// =============================================================================

func TestCalculateConfidence(t *testing.T) {
	testCases := []struct {
		name         string
		content      string
		finishReason string
		minConf      float64
		maxConf      float64
	}{
		{
			name:         "Short response",
			content:      "Short",
			finishReason: "stop",
			minConf:      0.3,
			maxConf:      0.7,
		},
		{
			name:         "Long response",
			content:      string(make([]byte, 1500)),
			finishReason: "stop",
			minConf:      0.7,
			maxConf:      0.95,
		},
		{
			name:         "Medium response",
			content:      string(make([]byte, 500)),
			finishReason: "stop",
			minConf:      0.7,
			maxConf:      0.85,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := &models.LLMResponse{
				Content:      tc.content,
				FinishReason: tc.finishReason,
			}
			conf := calculateConfidence(response)
			assert.GreaterOrEqual(t, conf, tc.minConf)
			assert.LessOrEqual(t, conf, tc.maxConf)
		})
	}
}

// =============================================================================
// Argument Extraction Tests
// =============================================================================

func TestExtractArguments(t *testing.T) {
	content := `Here is my analysis:
- First important point about the topic
- Second consideration that matters
* Third bullet with asterisk
1. Numbered point one
2. Numbered point two
Short line`

	args := extractArguments(content)

	assert.NotEmpty(t, args)
	assert.LessOrEqual(t, len(args), 5) // Max 5 arguments
}

func TestExtractArguments_Empty(t *testing.T) {
	args := extractArguments("No bullet points here")
	assert.Empty(t, args)
}

func TestExtractCriticisms(t *testing.T) {
	content := "- Criticism 1\n- Criticism 2"
	criticisms := extractCriticisms(content)
	assert.NotEmpty(t, criticisms)
}

func TestExtractSuggestions(t *testing.T) {
	content := "- Suggestion 1\n- Suggestion 2"
	suggestions := extractSuggestions(content)
	assert.NotEmpty(t, suggestions)
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestSplitLines(t *testing.T) {
	input := "line1\nline2\nline3"
	lines := splitLines(input)

	assert.Len(t, lines, 3)
	assert.Equal(t, "line1", lines[0])
	assert.Equal(t, "line2", lines[1])
	assert.Equal(t, "line3", lines[2])
}

func TestSplitLines_SingleLine(t *testing.T) {
	lines := splitLines("single line")
	assert.Len(t, lines, 1)
	assert.Equal(t, "single line", lines[0])
}

func TestTrimLine(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"  trimmed  ", "trimmed"},
		{"- bullet point", "bullet point"},
		{"* asterisk point", "asterisk point"},
		{"\ttabbed\t", "tabbed"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := trimLine(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// =============================================================================
// Debate Status Constants Tests
// =============================================================================

func TestDebateStatusConstants(t *testing.T) {
	assert.Equal(t, DebateStatus("pending"), DebateStatusPending)
	assert.Equal(t, DebateStatus("running"), DebateStatusRunning)
	assert.Equal(t, DebateStatus("completed"), DebateStatusCompleted)
	assert.Equal(t, DebateStatus("failed"), DebateStatusFailed)
	assert.Equal(t, DebateStatus("cancelled"), DebateStatusCancelled)
}

// =============================================================================
// Request/Response Type Tests
// =============================================================================

func TestDebateRequest_Defaults(t *testing.T) {
	request := &DebateRequest{
		Topic: "Test Topic",
	}

	assert.Empty(t, request.ID)
	assert.Equal(t, 0, request.MaxRounds)
	assert.Equal(t, time.Duration(0), request.Timeout)
}

func TestDebateResponse_Structure(t *testing.T) {
	response := &DebateResponse{
		ID:      "debate-1",
		Topic:   "Test Topic",
		Success: true,
		Consensus: &ConsensusResponse{
			Summary:    "Summary",
			Confidence: 0.85,
			KeyPoints:  []string{"Point 1", "Point 2"},
		},
		Phases: []*PhaseResponse{
			{
				Phase: "Proposal",
				Round: 1,
			},
		},
		Participants: []*ParticipantInfo{
			{
				AgentID:  "agent-1",
				Provider: "claude",
			},
		},
		Metrics: &DebateMetrics{
			TotalResponses: 10,
			AvgConfidence:  0.82,
		},
	}

	assert.Equal(t, "debate-1", response.ID)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Consensus)
	assert.Len(t, response.Phases, 1)
	assert.Len(t, response.Participants, 1)
	assert.NotNil(t, response.Metrics)
}

// =============================================================================
// Build Team Tests
// =============================================================================

func TestOrchestrator_buildTeam_InsufficientAgents(t *testing.T) {
	registry := newMockProviderRegistry()
	lessonBankConfig := debate.DefaultLessonBankConfig()
	lessonBankConfig.EnableSemanticSearch = false
	lessonBank := debate.NewLessonBank(lessonBankConfig, nil, nil)

	config := DefaultOrchestratorConfig()
	config.MinAgentsPerDebate = 5 // Require more than we have

	orch := NewOrchestrator(registry, lessonBank, config)
	// Only register 2 agents
	orch.RegisterProvider("claude", "claude-3", 9.0)
	orch.RegisterProvider("deepseek", "deepseek-coder", 8.5)

	request := &DebateRequest{
		Topic: "Test Topic",
	}

	_, _, err := orch.buildTeam(request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enough agents")
}

// =============================================================================
// OrchestratorStatistics Tests
// =============================================================================

func TestOrchestratorStatistics_Structure(t *testing.T) {
	stats := &OrchestratorStatistics{
		ActiveDebates:       5,
		RegisteredAgents:    10,
		TotalLessons:        50,
		TotalPatterns:       20,
		TotalDebatesLearned: 100,
		OverallSuccessRate:  0.85,
	}

	assert.Equal(t, 5, stats.ActiveDebates)
	assert.Equal(t, 10, stats.RegisteredAgents)
	assert.Equal(t, 50, stats.TotalLessons)
	assert.Equal(t, 20, stats.TotalPatterns)
	assert.Equal(t, 100, stats.TotalDebatesLearned)
	assert.Equal(t, 0.85, stats.OverallSuccessRate)
}
