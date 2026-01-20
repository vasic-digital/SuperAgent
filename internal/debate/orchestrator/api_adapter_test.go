package orchestrator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/debate/agents"
	"dev.helix.agent/internal/debate/topology"
)

// =============================================================================
// APIAdapter Creation Tests
// =============================================================================

func TestNewAPIAdapter(t *testing.T) {
	orch, _ := createTestOrchestrator()
	adapter := NewAPIAdapter(orch)

	assert.NotNil(t, adapter)
	assert.Equal(t, orch, adapter.orchestrator)
}

// =============================================================================
// Request Conversion Tests
// =============================================================================

func TestAPIAdapter_ConvertAPIRequest(t *testing.T) {
	orch, _ := createTestOrchestrator()
	adapter := NewAPIAdapter(orch)

	apiReq := &APICreateDebateRequest{
		DebateID: "test-debate-123",
		Topic:    "AI Ethics Discussion",
		Participants: []APIParticipantConfig{
			{Name: "Claude", Role: "analyst", LLMProvider: "claude", LLMModel: "claude-3"},
			{Name: "DeepSeek", Role: "coder", LLMProvider: "deepseek", LLMModel: "deepseek-v2"},
		},
		MaxRounds: 5,
		Timeout:   300, // 5 minutes
		Strategy:  "mesh",
		Metadata: map[string]interface{}{
			"source": "api",
		},
	}

	debateReq := adapter.ConvertAPIRequest(apiReq)

	require.NotNil(t, debateReq)
	assert.Equal(t, "test-debate-123", debateReq.ID)
	assert.Equal(t, "AI Ethics Discussion", debateReq.Topic)
	assert.Equal(t, 5, debateReq.MaxRounds)
	assert.Equal(t, 5*time.Minute, debateReq.Timeout)
	assert.Equal(t, topology.TopologyGraphMesh, debateReq.TopologyType)
	assert.Len(t, debateReq.PreferredProviders, 2)
	assert.Contains(t, debateReq.PreferredProviders, "claude")
	assert.Contains(t, debateReq.PreferredProviders, "deepseek")
	assert.NotNil(t, debateReq.Metadata)
	assert.Equal(t, "api", debateReq.Metadata["source"])
}

func TestAPIAdapter_ConvertAPIRequest_MinimalFields(t *testing.T) {
	orch, _ := createTestOrchestrator()
	adapter := NewAPIAdapter(orch)

	apiReq := &APICreateDebateRequest{
		Topic: "Simple Topic",
		Participants: []APIParticipantConfig{
			{Name: "Agent1"},
			{Name: "Agent2"},
		},
	}

	debateReq := adapter.ConvertAPIRequest(apiReq)

	require.NotNil(t, debateReq)
	assert.Equal(t, "Simple Topic", debateReq.Topic)
	assert.Equal(t, 0, debateReq.MaxRounds) // Uses default
	assert.Equal(t, time.Duration(0), debateReq.Timeout)
	assert.Equal(t, topology.TopologyGraphMesh, debateReq.TopologyType) // Default
}

func TestAPIAdapter_ConvertAPIRequest_WithMultiPass(t *testing.T) {
	orch, _ := createTestOrchestrator()
	adapter := NewAPIAdapter(orch)

	apiReq := &APICreateDebateRequest{
		Topic:                     "Multi-pass Topic",
		Participants:              []APIParticipantConfig{{Name: "A"}, {Name: "B"}},
		EnableMultiPassValidation: true,
	}

	debateReq := adapter.ConvertAPIRequest(apiReq)

	require.NotNil(t, debateReq.EnableLearning)
	assert.True(t, *debateReq.EnableLearning)
}

// =============================================================================
// Response Conversion Tests
// =============================================================================

func TestAPIAdapter_ConvertToAPIResponse(t *testing.T) {
	orch, _ := createTestOrchestrator()
	adapter := NewAPIAdapter(orch)

	orchResp := &DebateResponse{
		ID:      "response-123",
		Topic:   "Test Topic",
		Success: true,
		Phases: []*PhaseResponse{
			{
				Phase: "proposal",
				Round: 1,
				Responses: []*AgentResponse{
					{
						AgentID:    "agent-1",
						Provider:   "claude",
						Model:      "claude-3",
						Role:       "proposer",
						Content:    "Test content",
						Confidence: 0.85,
						Score:      9.0,
						Latency:    2 * time.Second,
					},
				},
				ConsensusLevel: 0.8,
			},
		},
		Consensus: &ConsensusResponse{
			Summary:       "Final consensus",
			Confidence:    0.9,
			KeyPoints:     []string{"Point 1", "Point 2"},
			Dissents:      []string{"Dissent 1"},
			VoteBreakdown: map[string]int{"agree": 3, "disagree": 1},
			Method:        "weighted",
		},
		Metrics: &DebateMetrics{
			TotalResponses: 5,
			AvgConfidence:  0.85,
			ConsensusScore: 0.88,
		},
		LessonsLearned:   3,
		PatternsDetected: 2,
		Duration:         5 * time.Minute,
		Metadata:         map[string]interface{}{"key": "value"},
	}

	apiResp := adapter.ConvertToAPIResponse(orchResp)

	require.NotNil(t, apiResp)
	assert.Equal(t, "response-123", apiResp.DebateID)
	assert.Equal(t, "Test Topic", apiResp.Topic)
	assert.Equal(t, "completed", apiResp.Status)
	assert.True(t, apiResp.Success)
	assert.Equal(t, 3, apiResp.LessonsLearned)
	assert.Equal(t, 2, apiResp.PatternsDetected)
	assert.Equal(t, "5m0s", apiResp.Duration)

	// Check responses
	assert.Len(t, apiResp.AllResponses, 1)
	resp := apiResp.AllResponses[0]
	assert.Equal(t, "agent-1", resp.ParticipantID)
	assert.Equal(t, "claude/claude-3", resp.ParticipantName)
	assert.Equal(t, "proposer", resp.Role)
	assert.Equal(t, "Test content", resp.Content)
	assert.Equal(t, 0.85, resp.Confidence)
	assert.Equal(t, 1, resp.Round)
	assert.Equal(t, "proposal", resp.Phase)

	// Check consensus
	require.NotNil(t, apiResp.Consensus)
	assert.True(t, apiResp.Consensus.Reached)
	assert.Equal(t, 0.9, apiResp.Consensus.AgreementLevel)
	assert.Equal(t, "Final consensus", apiResp.Consensus.FinalPosition)
	assert.Len(t, apiResp.Consensus.KeyPoints, 2)
	assert.Len(t, apiResp.Consensus.Disagreements, 1)
	assert.Equal(t, 3, apiResp.Consensus.VoteBreakdown["agree"])

	// Check scores
	assert.Equal(t, 0.85, apiResp.QualityScore)
	assert.Equal(t, 0.88, apiResp.FinalScore)
}

func TestAPIAdapter_ConvertToAPIResponse_Failed(t *testing.T) {
	orch, _ := createTestOrchestrator()
	adapter := NewAPIAdapter(orch)

	orchResp := &DebateResponse{
		ID:      "failed-123",
		Topic:   "Failed Topic",
		Success: false,
		Phases:  []*PhaseResponse{},
	}

	apiResp := adapter.ConvertToAPIResponse(orchResp)

	assert.Equal(t, "failed", apiResp.Status)
	assert.False(t, apiResp.Success)
	assert.Nil(t, apiResp.Consensus)
}

func TestAPIAdapter_ConvertToAPIResponse_NoMetrics(t *testing.T) {
	orch, _ := createTestOrchestrator()
	adapter := NewAPIAdapter(orch)

	orchResp := &DebateResponse{
		ID:      "no-metrics",
		Topic:   "Topic",
		Success: true,
		Phases:  []*PhaseResponse{},
	}

	apiResp := adapter.ConvertToAPIResponse(orchResp)

	assert.Equal(t, 0.0, apiResp.QualityScore)
	assert.Equal(t, 0.0, apiResp.FinalScore)
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestInferDomainFromRole(t *testing.T) {
	testCases := []struct {
		role     string
		expected agents.Domain
	}{
		{"security_analyst", agents.DomainSecurity},
		{"security", agents.DomainSecurity},
		{"architect", agents.DomainArchitecture},
		{"designer", agents.DomainArchitecture},
		{"optimizer", agents.DomainOptimization},
		{"performance", agents.DomainOptimization},
		{"debugger", agents.DomainDebug},
		{"troubleshooter", agents.DomainDebug},
		{"coder", agents.DomainCode},
		{"developer", agents.DomainCode},
		{"programmer", agents.DomainCode},
		{"analyst", agents.DomainReasoning},
		{"researcher", agents.DomainReasoning},
		{"reasoner", agents.DomainReasoning},
		{"unknown", agents.DomainGeneral},
		{"", agents.DomainGeneral},
	}

	for _, tc := range testCases {
		t.Run(tc.role, func(t *testing.T) {
			domain := inferDomainFromRole(tc.role)
			assert.Equal(t, tc.expected, domain)
		})
	}
}

func TestSelectTopologyFromStrategy(t *testing.T) {
	testCases := []struct {
		strategy string
		expected topology.TopologyType
	}{
		{"sequential", topology.TopologyChain},
		{"chain", topology.TopologyChain},
		{"pipeline", topology.TopologyChain},
		{"star", topology.TopologyStar},
		{"hub", topology.TopologyStar},
		{"centralized", topology.TopologyStar},
		{"mesh", topology.TopologyGraphMesh},
		{"parallel", topology.TopologyGraphMesh},
		{"distributed", topology.TopologyGraphMesh},
		{"unknown", topology.TopologyGraphMesh}, // Default
		{"", topology.TopologyGraphMesh},        // Default
	}

	for _, tc := range testCases {
		t.Run(tc.strategy, func(t *testing.T) {
			topo := selectTopologyFromStrategy(tc.strategy)
			assert.Equal(t, tc.expected, topo)
		})
	}
}

// =============================================================================
// API Types Structure Tests
// =============================================================================

func TestAPICreateDebateRequest_Structure(t *testing.T) {
	req := APICreateDebateRequest{
		DebateID: "debate-1",
		Topic:    "Test Topic",
		Participants: []APIParticipantConfig{
			{
				ParticipantID: "p1",
				Name:          "Agent 1",
				Role:          "proposer",
				LLMProvider:   "claude",
				LLMModel:      "claude-3",
			},
		},
		MaxRounds:                 5,
		Timeout:                   300,
		Strategy:                  "mesh",
		EnableCognee:              true,
		EnableMultiPassValidation: true,
		ValidationConfig: &APIValidationConfig{
			EnableValidation:    true,
			EnablePolish:        true,
			ValidationTimeout:   60,
			PolishTimeout:       30,
			MinConfidenceToSkip: 0.9,
			MaxValidationRounds: 3,
			ShowPhaseIndicators: true,
		},
		Metadata: map[string]interface{}{"key": "value"},
	}

	assert.Equal(t, "debate-1", req.DebateID)
	assert.Equal(t, 5, req.MaxRounds)
	assert.True(t, req.EnableMultiPassValidation)
	assert.NotNil(t, req.ValidationConfig)
	assert.True(t, req.ValidationConfig.EnableValidation)
}

func TestAPIDebateResponse_Structure(t *testing.T) {
	resp := APIDebateResponse{
		DebateID: "debate-1",
		Topic:    "Test Topic",
		Status:   "completed",
		Success:  true,
		AllResponses: []APIParticipantResponse{
			{
				ParticipantID:   "p1",
				ParticipantName: "Agent 1",
				Role:            "proposer",
				Content:         "Content",
				Confidence:      0.85,
				Round:           1,
				Phase:           "proposal",
			},
		},
		Consensus: &APIConsensusResult{
			Reached:        true,
			AgreementLevel: 0.9,
			FinalPosition:  "Final position",
			KeyPoints:      []string{"Point 1"},
			Method:         "weighted",
		},
		QualityScore:     0.85,
		FinalScore:       0.9,
		LessonsLearned:   5,
		PatternsDetected: 3,
		Duration:         "5m0s",
	}

	assert.Equal(t, "completed", resp.Status)
	assert.True(t, resp.Success)
	assert.Len(t, resp.AllResponses, 1)
	assert.NotNil(t, resp.Consensus)
	assert.True(t, resp.Consensus.Reached)
}

func TestAPIStatistics_Structure(t *testing.T) {
	stats := APIStatistics{
		ActiveDebates:       5,
		RegisteredAgents:    10,
		TotalLessons:        100,
		TotalPatterns:       50,
		TotalDebatesLearned: 200,
		OverallSuccessRate:  0.85,
	}

	assert.Equal(t, 5, stats.ActiveDebates)
	assert.Equal(t, 10, stats.RegisteredAgents)
	assert.Equal(t, 100, stats.TotalLessons)
	assert.Equal(t, 0.85, stats.OverallSuccessRate)
}

// =============================================================================
// GetDebateStatus and CancelDebate Tests
// =============================================================================

func TestAPIAdapter_GetDebateStatus_NotFound(t *testing.T) {
	orch, _ := createTestOrchestrator()
	adapter := NewAPIAdapter(orch)

	status, found := adapter.GetDebateStatus("non-existent")
	assert.False(t, found)
	assert.Equal(t, "", status)
}

func TestAPIAdapter_CancelDebate_NotFound(t *testing.T) {
	orch, _ := createTestOrchestrator()
	adapter := NewAPIAdapter(orch)

	err := adapter.CancelDebate("non-existent")
	assert.Error(t, err)
}

// =============================================================================
// GetStatistics Tests
// =============================================================================

func TestAPIAdapter_GetStatistics(t *testing.T) {
	orch, _ := createTestOrchestrator()
	adapter := NewAPIAdapter(orch)

	ctx := t.Context()
	stats, err := adapter.GetStatistics(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, 0, stats.ActiveDebates)
	assert.Equal(t, 3, stats.RegisteredAgents)
	assert.GreaterOrEqual(t, stats.TotalLessons, 0)
}
