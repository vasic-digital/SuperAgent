package orchestrator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/debate/agents"
	"dev.helix.agent/internal/debate/protocol"
	"dev.helix.agent/internal/debate/topology"
)

// =============================================================================
// Legacy Config Conversion Tests
// =============================================================================

func TestConvertFromLegacyConfig(t *testing.T) {
	legacy := &LegacyDebateConfig{
		DebateID:  "debate-123",
		Topic:     "Test Topic",
		MaxRounds: 5,
		Timeout:   300, // 5 minutes in seconds
		Participants: []LegacyParticipant{
			{
				Name:        "Agent 1",
				Role:        "proposer",
				LLMProvider: "claude",
				LLMModel:    "claude-3",
			},
			{
				Name:        "Agent 2",
				Role:        "critic",
				LLMProvider: "deepseek",
				LLMModel:    "deepseek-coder",
			},
		},
		EnableCognee: true,
		Metadata: map[string]interface{}{
			"custom_key": "custom_value",
		},
	}

	request := ConvertFromLegacyConfig(legacy)

	require.NotNil(t, request)
	assert.Equal(t, "debate-123", request.ID)
	assert.Equal(t, "Test Topic", request.Topic)
	assert.Equal(t, 5, request.MaxRounds)
	assert.Equal(t, 5*time.Minute, request.Timeout)
	assert.Len(t, request.PreferredProviders, 2)
	assert.Contains(t, request.PreferredProviders, "claude")
	assert.Contains(t, request.PreferredProviders, "deepseek")
	assert.NotNil(t, request.Metadata)
	assert.Equal(t, "custom_value", request.Metadata["custom_key"])
}

func TestConvertFromLegacyConfig_EmptyParticipants(t *testing.T) {
	legacy := &LegacyDebateConfig{
		DebateID:  "debate-456",
		Topic:     "Empty Participants Topic",
		MaxRounds: 3,
		Timeout:   120,
	}

	request := ConvertFromLegacyConfig(legacy)

	require.NotNil(t, request)
	assert.Empty(t, request.PreferredProviders)
}

// =============================================================================
// Legacy Result Conversion Tests
// =============================================================================

func TestConvertToLegacyResult(t *testing.T) {
	startTime := time.Now().Add(-5 * time.Minute)
	response := &DebateResponse{
		ID:       "debate-789",
		Topic:    "Test Topic",
		Success:  true,
		Duration: 5 * time.Minute,
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
			Summary:    "Final consensus summary",
			Confidence: 0.9,
			KeyPoints:  []string{"Point 1", "Point 2"},
			Dissents:   []string{"Dissent 1"},
		},
		Metrics: &DebateMetrics{
			TotalResponses: 5,
			AvgConfidence:  0.85,
			ConsensusScore: 0.88,
		},
		Metadata: map[string]interface{}{
			"test_key": "test_value",
		},
	}

	result := ConvertToLegacyResult(response, startTime)

	require.NotNil(t, result)
	assert.Equal(t, "debate-789", result.DebateID)
	assert.Equal(t, "Test Topic", result.Topic)
	assert.Equal(t, startTime, result.StartTime)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.AllResponses)

	// Check first response
	assert.Len(t, result.AllResponses, 1)
	resp := result.AllResponses[0]
	assert.Equal(t, "claude/claude-3", resp.ParticipantName)
	assert.Equal(t, "proposer", resp.Role)
	assert.Equal(t, "Test content", resp.Content)
	assert.Equal(t, 0.85, resp.Confidence)
	assert.Equal(t, 9.0, resp.QualityScore)
	assert.Equal(t, 1, resp.Round)

	// Check consensus
	require.NotNil(t, result.Consensus)
	assert.True(t, result.Consensus.Reached)
	assert.Equal(t, 0.9, result.Consensus.AgreementLevel)
	assert.Equal(t, "Final consensus summary", result.Consensus.FinalPosition)
	assert.Len(t, result.Consensus.KeyPoints, 2)
	assert.Len(t, result.Consensus.Disagreements, 1)

	// Check metrics
	assert.Equal(t, 0.85, result.QualityScore)
	assert.Equal(t, 0.88, result.FinalScore)
}

func TestConvertToLegacyResult_NoConsensus(t *testing.T) {
	startTime := time.Now()
	response := &DebateResponse{
		ID:      "debate-no-consensus",
		Topic:   "Topic",
		Success: false,
		Phases:  []*PhaseResponse{},
	}

	result := ConvertToLegacyResult(response, startTime)

	require.NotNil(t, result)
	assert.Nil(t, result.Consensus)
	assert.False(t, result.Success)
}

func TestConvertToLegacyResult_NoMetrics(t *testing.T) {
	startTime := time.Now()
	response := &DebateResponse{
		ID:      "debate-no-metrics",
		Topic:   "Topic",
		Success: true,
		Phases:  []*PhaseResponse{},
	}

	result := ConvertToLegacyResult(response, startTime)

	require.NotNil(t, result)
	assert.Equal(t, 0.0, result.QualityScore)
	assert.Equal(t, 0.0, result.FinalScore)
}

// =============================================================================
// Protocol Result Conversion Tests
// =============================================================================

func TestConvertProtocolResultToResponse(t *testing.T) {
	protocolResult := &protocol.DebateResult{
		ID:      "protocol-debate-1",
		Topic:   "Protocol Topic",
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
						Content:    "Protocol response content",
						Confidence: 0.88,
						Score:      8.5,
						Latency:    1500 * time.Millisecond,
					},
				},
				ConsensusLevel: 0.75,
				KeyInsights:    []string{"Insight 1", "Insight 2"},
				Duration:       30 * time.Second,
			},
		},
		FinalConsensus: &protocol.ConsensusResult{
			Summary:       "Protocol consensus summary",
			Confidence:    0.92,
			KeyPoints:     []string{"Key 1", "Key 2"},
			Dissents:      []string{"Dissent 1"},
			VoteBreakdown: map[string]int{"agree": 3, "disagree": 1},
			WinningVote:   "agree",
			Method:        protocol.ConsensusMethodWeightedVoting,
		},
		Metrics: &protocol.DebateMetrics{
			TotalResponses:     10,
			AvgLatency:         2 * time.Second,
			AvgConfidence:      0.85,
			ConsensusScore:     0.9,
			AgentParticipation: map[string]int{"claude": 5, "deepseek": 5},
		},
		Duration: 3 * time.Minute,
		Metadata: map[string]interface{}{
			"protocol_key": "protocol_value",
		},
	}

	response := ConvertProtocolResultToResponse(protocolResult)

	require.NotNil(t, response)
	assert.Equal(t, "protocol-debate-1", response.ID)
	assert.Equal(t, "Protocol Topic", response.Topic)
	assert.True(t, response.Success)
	assert.Equal(t, 3*time.Minute, response.Duration)

	// Check phases
	require.Len(t, response.Phases, 1)
	phase := response.Phases[0]
	assert.Equal(t, "proposal", phase.Phase)
	assert.Equal(t, 1, phase.Round)
	assert.Equal(t, 0.75, phase.ConsensusLevel)
	assert.Len(t, phase.KeyInsights, 2)
	assert.Len(t, phase.Responses, 1)

	phaseResp := phase.Responses[0]
	assert.Equal(t, "agent-1", phaseResp.AgentID)
	assert.Equal(t, "claude", phaseResp.Provider)
	assert.Equal(t, "proposer", phaseResp.Role)
	assert.Equal(t, 0.88, phaseResp.Confidence)

	// Check consensus
	require.NotNil(t, response.Consensus)
	assert.Equal(t, "Protocol consensus summary", response.Consensus.Summary)
	assert.Equal(t, 0.92, response.Consensus.Confidence)
	assert.Len(t, response.Consensus.KeyPoints, 2)
	assert.Len(t, response.Consensus.Dissents, 1)
	assert.Equal(t, "agree", response.Consensus.WinningVote)
	assert.Equal(t, "weighted_voting", response.Consensus.Method)

	// Check metrics
	require.NotNil(t, response.Metrics)
	assert.Equal(t, 10, response.Metrics.TotalResponses)
	assert.Equal(t, 2*time.Second, response.Metrics.AvgLatency)
	assert.Equal(t, 0.85, response.Metrics.AvgConfidence)
	assert.Equal(t, 0.9, response.Metrics.ConsensusScore)
}

func TestConvertProtocolResultToResponse_NoConsensus(t *testing.T) {
	protocolResult := &protocol.DebateResult{
		ID:      "no-consensus",
		Topic:   "Topic",
		Success: false,
		Phases:  []*protocol.PhaseResult{},
	}

	response := ConvertProtocolResultToResponse(protocolResult)

	require.NotNil(t, response)
	assert.Nil(t, response.Consensus)
	assert.Nil(t, response.Metrics)
}

// =============================================================================
// Role Mapping Tests
// =============================================================================

func TestMapLegacyRole(t *testing.T) {
	testCases := []struct {
		legacyRole string
		expected   topology.AgentRole
	}{
		{"proposer", topology.RoleProposer},
		{"critic", topology.RoleCritic},
		{"debater", topology.RoleReviewer},
		{"moderator", topology.RoleModerator},
		{"analyst", topology.RoleCritic},
		{"synthesis", topology.RoleModerator},
		{"synthesizer", topology.RoleModerator},
		{"mediator", topology.RoleModerator},
		{"architect", topology.RoleArchitect},
		{"optimizer", topology.RoleOptimizer},
		{"validator", topology.RoleValidator},
		{"red-team", topology.RoleRedTeam},
		{"red_team", topology.RoleRedTeam},
		{"blue-team", topology.RoleBlueTeam},
		{"blue_team", topology.RoleBlueTeam},
		{"unknown", topology.RoleReviewer}, // Default
		{"", topology.RoleReviewer},        // Empty string
	}

	for _, tc := range testCases {
		t.Run(tc.legacyRole, func(t *testing.T) {
			role := MapLegacyRole(tc.legacyRole)
			assert.Equal(t, tc.expected, role)
		})
	}
}

func TestMapRoleToLegacy(t *testing.T) {
	testCases := []struct {
		role     topology.AgentRole
		expected string
	}{
		{topology.RoleProposer, "proposer"},
		{topology.RoleCritic, "critic"},
		{topology.RoleReviewer, "debater"},
		{topology.RoleModerator, "moderator"},
		{topology.RoleArchitect, "architect"},
		{topology.RoleOptimizer, "optimizer"},
		{topology.RoleValidator, "validator"},
		{topology.RoleRedTeam, "red-team"},
		{topology.RoleBlueTeam, "blue-team"},
		{topology.AgentRole("unknown"), "debater"}, // Default
	}

	for _, tc := range testCases {
		t.Run(string(tc.role), func(t *testing.T) {
			legacy := MapRoleToLegacy(tc.role)
			assert.Equal(t, tc.expected, legacy)
		})
	}
}

// =============================================================================
// Domain Mapping Tests
// =============================================================================

func TestMapTopicToDomain(t *testing.T) {
	testCases := []struct {
		topic    string
		expected agents.Domain
	}{
		// Security domain
		{"Security vulnerability assessment", agents.DomainSecurity},
		{"Authentication best practices", agents.DomainSecurity},
		{"Encryption algorithms comparison", agents.DomainSecurity},
		{"Hack prevention strategies", agents.DomainSecurity},

		// Architecture domain
		{"Architecture design patterns", agents.DomainArchitecture},
		{"System scalability design", agents.DomainArchitecture},
		{"Microservice architecture", agents.DomainArchitecture},

		// Optimization domain
		{"Performance optimization techniques", agents.DomainOptimization},
		{"Cache strategies", agents.DomainOptimization},
		{"Latency reduction", agents.DomainOptimization},
		{"Speed improvements", agents.DomainOptimization},

		// Debug domain
		{"Debug production error", agents.DomainDebug},
		{"Bug fixing strategies", agents.DomainDebug},
		{"Error handling", agents.DomainDebug},
		{"Trace analysis", agents.DomainDebug},

		// Code domain
		{"Code implementation approach", agents.DomainCode},
		{"Function refactoring", agents.DomainCode},
		{"Class structure", agents.DomainCode}, // Note: "design" would match architecture first

		// Reasoning domain
		{"Logic problem solving", agents.DomainReasoning},
		{"Algorithm complexity", agents.DomainReasoning}, // Note: "algorithm design" would match architecture first
		{"Mathematical proof", agents.DomainReasoning},

		// General domain (no keywords)
		{"Random topic without keywords", agents.DomainGeneral},
		{"", agents.DomainGeneral},
	}

	for _, tc := range testCases {
		t.Run(tc.topic, func(t *testing.T) {
			domain := MapTopicToDomain(tc.topic)
			assert.Equal(t, tc.expected, domain)
		})
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestToLower(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"Hello World", "hello world"},
		{"hello", "hello"},
		{"123ABC", "123abc"},
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := toLower(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestContainsKeyword(t *testing.T) {
	assert.True(t, containsKeyword("hello world", "hello", "test"))
	assert.True(t, containsKeyword("hello world", "world"))
	assert.False(t, containsKeyword("hello world", "foo", "bar"))
	assert.False(t, containsKeyword("", "test"))
}

func TestContains(t *testing.T) {
	assert.True(t, contains("hello world", "hello"))
	assert.True(t, contains("hello world", "world"))
	assert.True(t, contains("hello world", "lo wo"))
	assert.False(t, contains("hello world", "xyz"))
	assert.False(t, contains("short", "longer substring"))
}

// =============================================================================
// Legacy Type Structure Tests
// =============================================================================

func TestLegacyDebateConfig_Structure(t *testing.T) {
	config := LegacyDebateConfig{
		DebateID:  "debate-1",
		Topic:     "Test Topic",
		MaxRounds: 3,
		Timeout:   180,
		Participants: []LegacyParticipant{
			{
				Name:         "Agent 1",
				Role:         "proposer",
				LLMProvider:  "claude",
				LLMModel:     "claude-3",
				Temperature:  0.7,
				SystemPrompt: "Be creative",
				Fallbacks: []LegacyFallback{
					{LLMProvider: "deepseek", LLMModel: "deepseek-v2"},
				},
			},
		},
		EnableCognee: true,
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	assert.Equal(t, "debate-1", config.DebateID)
	assert.Equal(t, 3, config.MaxRounds)
	assert.Len(t, config.Participants, 1)
	assert.Len(t, config.Participants[0].Fallbacks, 1)
}

func TestLegacyDebateResult_Structure(t *testing.T) {
	result := LegacyDebateResult{
		DebateID:  "result-1",
		Topic:     "Result Topic",
		StartTime: time.Now().Add(-10 * time.Minute),
		EndTime:   time.Now(),
		AllResponses: []LegacyParticipantResponse{
			{
				ParticipantName: "Agent 1",
				Role:            "proposer",
				Content:         "Response content",
				Confidence:      0.85,
				QualityScore:    0.9,
				Round:           1,
			},
		},
		Consensus: &LegacyConsensus{
			Reached:        true,
			AgreementLevel: 0.88,
			FinalPosition:  "Final position",
			KeyPoints:      []string{"Point 1"},
			Disagreements:  []string{"Disagreement 1"},
		},
		QualityScore: 0.85,
		FinalScore:   0.9,
		Success:      true,
	}

	assert.Equal(t, "result-1", result.DebateID)
	assert.True(t, result.Success)
	assert.NotNil(t, result.Consensus)
	assert.True(t, result.Consensus.Reached)
}
