// Package orchestrator provides adapters between the new debate framework
// and existing HelixAgent service types.
package orchestrator

import (
	"time"

	"dev.helix.agent/internal/debate/agents"
	"dev.helix.agent/internal/debate/protocol"
	"dev.helix.agent/internal/debate/topology"
)

// =============================================================================
// Service Type Adapters
// =============================================================================

// LegacyDebateConfig represents the existing debate configuration format.
type LegacyDebateConfig struct {
	DebateID     string                 `json:"debate_id"`
	Topic        string                 `json:"topic"`
	MaxRounds    int                    `json:"max_rounds"`
	Timeout      int                    `json:"timeout"` // seconds
	Participants []LegacyParticipant    `json:"participants"`
	EnableCognee bool                   `json:"enable_cognee"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// LegacyParticipant represents the existing participant format.
type LegacyParticipant struct {
	Name         string           `json:"name"`
	Role         string           `json:"role"` // "proposer", "critic", "debater"
	LLMProvider  string           `json:"llm_provider"`
	LLMModel     string           `json:"llm_model"`
	Temperature  float64          `json:"temperature,omitempty"`
	SystemPrompt string           `json:"system_prompt,omitempty"`
	Fallbacks    []LegacyFallback `json:"fallbacks,omitempty"`
}

// LegacyFallback represents a fallback provider configuration.
type LegacyFallback struct {
	LLMProvider string `json:"llm_provider"`
	LLMModel    string `json:"llm_model"`
}

// LegacyDebateResult represents the existing debate result format.
type LegacyDebateResult struct {
	DebateID     string                      `json:"debate_id"`
	Topic        string                      `json:"topic"`
	StartTime    time.Time                   `json:"start_time"`
	EndTime      time.Time                   `json:"end_time"`
	AllResponses []LegacyParticipantResponse `json:"all_responses"`
	Consensus    *LegacyConsensus            `json:"consensus,omitempty"`
	QualityScore float64                     `json:"quality_score"`
	FinalScore   float64                     `json:"final_score"`
	Success      bool                        `json:"success"`
	Metadata     map[string]interface{}      `json:"metadata,omitempty"`
}

// LegacyParticipantResponse represents the existing participant response format.
type LegacyParticipantResponse struct {
	ParticipantName string                 `json:"participant_name"`
	Role            string                 `json:"role"`
	LLMProvider     string                 `json:"llm_provider"`
	LLMModel        string                 `json:"llm_model"`
	Content         string                 `json:"content"`
	Confidence      float64                `json:"confidence"`
	QualityScore    float64                `json:"quality_score"`
	ResponseTime    time.Duration          `json:"response_time"`
	Round           int                    `json:"round"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// LegacyConsensus represents the existing consensus format.
type LegacyConsensus struct {
	Reached        bool     `json:"reached"`
	AgreementLevel float64  `json:"agreement_level"`
	FinalPosition  string   `json:"final_position"`
	KeyPoints      []string `json:"key_points"`
	Disagreements  []string `json:"disagreements,omitempty"`
}

// =============================================================================
// Conversion Functions
// =============================================================================

// ConvertFromLegacyConfig converts a legacy config to the new DebateRequest format.
func ConvertFromLegacyConfig(legacy *LegacyDebateConfig) *DebateRequest {
	request := &DebateRequest{
		ID:        legacy.DebateID,
		Topic:     legacy.Topic,
		MaxRounds: legacy.MaxRounds,
		Timeout:   time.Duration(legacy.Timeout) * time.Second,
		Metadata:  legacy.Metadata,
	}

	// Extract preferred providers from participants
	for _, p := range legacy.Participants {
		request.PreferredProviders = append(request.PreferredProviders, p.LLMProvider)
	}

	return request
}

// ConvertToLegacyResult converts a DebateResponse to the legacy result format.
func ConvertToLegacyResult(response *DebateResponse, startTime time.Time) *LegacyDebateResult {
	result := &LegacyDebateResult{
		DebateID:     response.ID,
		Topic:        response.Topic,
		StartTime:    startTime,
		EndTime:      startTime.Add(response.Duration),
		AllResponses: make([]LegacyParticipantResponse, 0),
		Success:      response.Success,
		Metadata:     response.Metadata,
	}

	// Convert phase responses to flat list
	for _, phase := range response.Phases {
		for _, agentResp := range phase.Responses {
			legacyResp := LegacyParticipantResponse{
				ParticipantName: agentResp.Provider + "/" + agentResp.Model,
				Role:            agentResp.Role,
				LLMProvider:     agentResp.Provider,
				LLMModel:        agentResp.Model,
				Content:         agentResp.Content,
				Confidence:      agentResp.Confidence,
				QualityScore:    agentResp.Score,
				ResponseTime:    agentResp.Latency,
				Round:           phase.Round,
				Metadata: map[string]interface{}{
					"phase": phase.Phase,
				},
			}
			result.AllResponses = append(result.AllResponses, legacyResp)
		}
	}

	// Convert consensus
	if response.Consensus != nil {
		result.Consensus = &LegacyConsensus{
			Reached:        response.Success,
			AgreementLevel: response.Consensus.Confidence,
			FinalPosition:  response.Consensus.Summary,
			KeyPoints:      response.Consensus.KeyPoints,
			Disagreements:  response.Consensus.Dissents,
		}
	}

	// Calculate scores
	if response.Metrics != nil {
		result.QualityScore = response.Metrics.AvgConfidence
		result.FinalScore = response.Metrics.ConsensusScore
	}

	return result
}

// ConvertProtocolResultToResponse converts a protocol.DebateResult to DebateResponse.
func ConvertProtocolResultToResponse(result *protocol.DebateResult) *DebateResponse {
	response := &DebateResponse{
		ID:       result.ID,
		Topic:    result.Topic,
		Success:  result.Success,
		Phases:   make([]*PhaseResponse, 0, len(result.Phases)),
		Duration: result.Duration,
		Metadata: result.Metadata,
	}

	// Convert consensus
	if result.FinalConsensus != nil {
		response.Consensus = &ConsensusResponse{
			Summary:       result.FinalConsensus.Summary,
			Confidence:    result.FinalConsensus.Confidence,
			KeyPoints:     result.FinalConsensus.KeyPoints,
			Dissents:      result.FinalConsensus.Dissents,
			VoteBreakdown: result.FinalConsensus.VoteBreakdown,
			WinningVote:   result.FinalConsensus.WinningVote,
			Method:        string(result.FinalConsensus.Method),
		}
	}

	// Convert phases
	for _, phase := range result.Phases {
		phaseResp := &PhaseResponse{
			Phase:          string(phase.Phase),
			Round:          phase.Round,
			Responses:      make([]*AgentResponse, 0, len(phase.Responses)),
			ConsensusLevel: phase.ConsensusLevel,
			KeyInsights:    phase.KeyInsights,
			Duration:       phase.Duration,
		}

		for _, resp := range phase.Responses {
			phaseResp.Responses = append(phaseResp.Responses, &AgentResponse{
				AgentID:    resp.AgentID,
				Provider:   resp.Provider,
				Model:      resp.Model,
				Role:       string(resp.Role),
				Content:    resp.Content,
				Confidence: resp.Confidence,
				Score:      resp.Score,
				Latency:    resp.Latency,
			})
		}

		response.Phases = append(response.Phases, phaseResp)
	}

	// Convert metrics
	if result.Metrics != nil {
		response.Metrics = &DebateMetrics{
			TotalResponses:    result.Metrics.TotalResponses,
			AvgLatency:        result.Metrics.AvgLatency,
			AvgConfidence:     result.Metrics.AvgConfidence,
			ConsensusScore:    result.Metrics.ConsensusScore,
			ProviderBreakdown: result.Metrics.AgentParticipation,
		}
	}

	return response
}

// =============================================================================
// Role Mapping
// =============================================================================

// MapLegacyRole maps legacy role strings to topology.AgentRole.
func MapLegacyRole(legacyRole string) topology.AgentRole {
	switch legacyRole {
	case "proposer":
		return topology.RoleProposer
	case "critic":
		return topology.RoleCritic
	case "debater":
		return topology.RoleReviewer
	case "moderator":
		return topology.RoleModerator
	case "analyst":
		return topology.RoleCritic // Analyst maps to critic behavior
	case "synthesis", "synthesizer":
		return topology.RoleModerator // Synthesis maps to moderator
	case "mediator":
		return topology.RoleModerator
	case "architect":
		return topology.RoleArchitect
	case "optimizer":
		return topology.RoleOptimizer
	case "validator":
		return topology.RoleValidator
	case "red-team", "red_team":
		return topology.RoleRedTeam
	case "blue-team", "blue_team":
		return topology.RoleBlueTeam
	default:
		return topology.RoleReviewer // Default to reviewer
	}
}

// MapRoleToLegacy maps topology.AgentRole to legacy role string.
func MapRoleToLegacy(role topology.AgentRole) string {
	switch role {
	case topology.RoleProposer:
		return "proposer"
	case topology.RoleCritic:
		return "critic"
	case topology.RoleReviewer:
		return "debater"
	case topology.RoleModerator:
		return "moderator"
	case topology.RoleArchitect:
		return "architect"
	case topology.RoleOptimizer:
		return "optimizer"
	case topology.RoleValidator:
		return "validator"
	case topology.RoleRedTeam:
		return "red-team"
	case topology.RoleBlueTeam:
		return "blue-team"
	default:
		return "debater"
	}
}

// =============================================================================
// Domain Mapping
// =============================================================================

// MapTopicToDomain maps a topic to a likely domain.
func MapTopicToDomain(topic string) agents.Domain {
	// Use the knowledge package's domain inference
	// This is a simplified version - in production, use the full inference
	topicLower := toLower(topic)

	if containsKeyword(topicLower, "security", "vulnerability", "auth", "encrypt", "hack") {
		return agents.DomainSecurity
	}
	if containsKeyword(topicLower, "architecture", "design", "system", "scale", "microservice") {
		return agents.DomainArchitecture
	}
	if containsKeyword(topicLower, "performance", "optimize", "speed", "cache", "latency") {
		return agents.DomainOptimization
	}
	if containsKeyword(topicLower, "debug", "error", "bug", "fix", "trace") {
		return agents.DomainDebug
	}
	if containsKeyword(topicLower, "code", "implement", "function", "class", "refactor") {
		return agents.DomainCode
	}
	if containsKeyword(topicLower, "logic", "reason", "proof", "algorithm", "math") {
		return agents.DomainReasoning
	}

	return agents.DomainGeneral
}

// Helper functions for adapter
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

func containsKeyword(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if contains(s, kw) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
