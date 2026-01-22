// Package orchestrator provides the debate orchestrator that bridges
// the new debate framework with existing HelixAgent services.
package orchestrator

import (
	"context"
	"time"

	"dev.helix.agent/internal/debate/agents"
	"dev.helix.agent/internal/debate/topology"
)

// =============================================================================
// API Adapter - Converts between API types and Orchestrator types
// =============================================================================

// APIAdapter provides conversion between API request/response types
// and internal orchestrator types.
type APIAdapter struct {
	orchestrator *Orchestrator
}

// NewAPIAdapter creates a new API adapter.
func NewAPIAdapter(orchestrator *Orchestrator) *APIAdapter {
	return &APIAdapter{orchestrator: orchestrator}
}

// =============================================================================
// API Request Types
// =============================================================================

// APICreateDebateRequest matches the handlers.CreateDebateRequest structure.
type APICreateDebateRequest struct {
	DebateID                  string                 `json:"debate_id,omitempty"`
	Topic                     string                 `json:"topic"`
	Participants              []APIParticipantConfig `json:"participants"`
	MaxRounds                 int                    `json:"max_rounds,omitempty"`
	Timeout                   int                    `json:"timeout,omitempty"` // seconds
	Strategy                  string                 `json:"strategy,omitempty"`
	EnableCognee              bool                   `json:"enable_cognee,omitempty"`
	EnableMultiPassValidation bool                   `json:"enable_multi_pass_validation,omitempty"`
	ValidationConfig          *APIValidationConfig   `json:"validation_config,omitempty"`
	Metadata                  map[string]interface{} `json:"metadata,omitempty"`
}

// APIParticipantConfig matches handlers.ParticipantConfigRequest.
type APIParticipantConfig struct {
	ParticipantID string `json:"participant_id,omitempty"`
	Name          string `json:"name"`
	Role          string `json:"role,omitempty"`
	LLMProvider   string `json:"llm_provider,omitempty"`
	LLMModel      string `json:"llm_model,omitempty"`
}

// APIValidationConfig matches handlers.ValidationConfigRequest.
type APIValidationConfig struct {
	EnableValidation    bool    `json:"enable_validation"`
	EnablePolish        bool    `json:"enable_polish"`
	ValidationTimeout   int     `json:"validation_timeout,omitempty"`
	PolishTimeout       int     `json:"polish_timeout,omitempty"`
	MinConfidenceToSkip float64 `json:"min_confidence_to_skip,omitempty"`
	MaxValidationRounds int     `json:"max_validation_rounds,omitempty"`
	ShowPhaseIndicators bool    `json:"show_phase_indicators,omitempty"`
}

// =============================================================================
// API Response Types
// =============================================================================

// APIDebateResponse matches the expected API response structure.
type APIDebateResponse struct {
	DebateID         string                   `json:"debate_id"`
	Topic            string                   `json:"topic"`
	Status           string                   `json:"status"`
	Success          bool                     `json:"success"`
	AllResponses     []APIParticipantResponse `json:"all_responses"`
	Consensus        *APIConsensusResult      `json:"consensus,omitempty"`
	QualityScore     float64                  `json:"quality_score"`
	FinalScore       float64                  `json:"final_score"`
	LessonsLearned   int                      `json:"lessons_learned"`
	PatternsDetected int                      `json:"patterns_detected"`
	Duration         string                   `json:"duration"`
	Metadata         map[string]interface{}   `json:"metadata,omitempty"`
}

// APIParticipantResponse represents a participant's response in API format.
type APIParticipantResponse struct {
	ParticipantID   string                 `json:"participant_id"`
	ParticipantName string                 `json:"participant_name"`
	Role            string                 `json:"role"`
	LLMProvider     string                 `json:"llm_provider"`
	LLMModel        string                 `json:"llm_model"`
	Content         string                 `json:"content"`
	Confidence      float64                `json:"confidence"`
	QualityScore    float64                `json:"quality_score"`
	ResponseTime    string                 `json:"response_time"`
	Round           int                    `json:"round"`
	Phase           string                 `json:"phase"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// APIConsensusResult represents the consensus in API format.
type APIConsensusResult struct {
	Reached        bool           `json:"reached"`
	AgreementLevel float64        `json:"agreement_level"`
	FinalPosition  string         `json:"final_position"`
	KeyPoints      []string       `json:"key_points"`
	Disagreements  []string       `json:"disagreements,omitempty"`
	VoteBreakdown  map[string]int `json:"vote_breakdown,omitempty"`
	Method         string         `json:"method"`
}

// =============================================================================
// Conversion Methods
// =============================================================================

// ConvertAPIRequest converts an API request to an orchestrator DebateRequest.
func (a *APIAdapter) ConvertAPIRequest(apiReq *APICreateDebateRequest) *DebateRequest {
	debateReq := &DebateRequest{
		ID:        apiReq.DebateID,
		Topic:     apiReq.Topic,
		MaxRounds: apiReq.MaxRounds,
		Metadata:  apiReq.Metadata,
	}

	// Convert timeout (API uses seconds)
	if apiReq.Timeout > 0 {
		debateReq.Timeout = time.Duration(apiReq.Timeout) * time.Second
	}

	// Extract preferred providers and domain from participants
	var preferredProviders []string
	var preferredDomain agents.Domain

	for _, p := range apiReq.Participants {
		if p.LLMProvider != "" {
			preferredProviders = append(preferredProviders, p.LLMProvider)
		}
		// Infer domain from role
		if preferredDomain == "" && p.Role != "" {
			preferredDomain = inferDomainFromRole(p.Role)
		}
	}

	debateReq.PreferredProviders = preferredProviders
	debateReq.PreferredDomain = preferredDomain

	// Select topology based on strategy
	debateReq.TopologyType = selectTopologyFromStrategy(apiReq.Strategy)

	// Enable learning if multi-pass validation is enabled
	if apiReq.EnableMultiPassValidation {
		enableLearning := true
		debateReq.EnableLearning = &enableLearning
	}

	return debateReq
}

// ConvertToAPIResponse converts an orchestrator response to an API response.
func (a *APIAdapter) ConvertToAPIResponse(orchResp *DebateResponse) *APIDebateResponse {
	apiResp := &APIDebateResponse{
		DebateID:         orchResp.ID,
		Topic:            orchResp.Topic,
		Success:          orchResp.Success,
		LessonsLearned:   orchResp.LessonsLearned,
		PatternsDetected: orchResp.PatternsDetected,
		Duration:         orchResp.Duration.String(),
		Metadata:         orchResp.Metadata,
	}

	// Set status
	if orchResp.Success {
		apiResp.Status = "completed"
	} else {
		apiResp.Status = "failed"
	}

	// Convert phases to flat responses
	apiResp.AllResponses = make([]APIParticipantResponse, 0)
	for _, phase := range orchResp.Phases {
		for _, resp := range phase.Responses {
			apiResp.AllResponses = append(apiResp.AllResponses, APIParticipantResponse{
				ParticipantID:   resp.AgentID,
				ParticipantName: resp.Provider + "/" + resp.Model,
				Role:            resp.Role,
				LLMProvider:     resp.Provider,
				LLMModel:        resp.Model,
				Content:         resp.Content,
				Confidence:      resp.Confidence,
				QualityScore:    resp.Score,
				ResponseTime:    resp.Latency.String(),
				Round:           phase.Round,
				Phase:           phase.Phase,
			})
		}
	}

	// Convert consensus
	if orchResp.Consensus != nil {
		apiResp.Consensus = &APIConsensusResult{
			Reached:        orchResp.Success,
			AgreementLevel: orchResp.Consensus.Confidence,
			FinalPosition:  orchResp.Consensus.Summary,
			KeyPoints:      orchResp.Consensus.KeyPoints,
			Disagreements:  orchResp.Consensus.Dissents,
			VoteBreakdown:  orchResp.Consensus.VoteBreakdown,
			Method:         orchResp.Consensus.Method,
		}
	}

	// Set scores from metrics
	if orchResp.Metrics != nil {
		apiResp.QualityScore = orchResp.Metrics.AvgConfidence
		apiResp.FinalScore = orchResp.Metrics.ConsensusScore
	}

	return apiResp
}

// =============================================================================
// High-Level API Methods
// =============================================================================

// ConductDebate runs a complete debate from an API request.
func (a *APIAdapter) ConductDebate(ctx context.Context, apiReq *APICreateDebateRequest) (*APIDebateResponse, error) {
	// Convert request
	debateReq := a.ConvertAPIRequest(apiReq)

	// Run debate
	orchResp, err := a.orchestrator.ConductDebate(ctx, debateReq)
	if err != nil {
		return nil, err
	}

	// Convert response
	return a.ConvertToAPIResponse(orchResp), nil
}

// GetDebateStatus returns the status of an active debate.
func (a *APIAdapter) GetDebateStatus(debateID string) (string, bool) {
	status, found := a.orchestrator.GetDebateStatus(debateID)
	if !found {
		return "", false
	}
	return string(status), true
}

// CancelDebate cancels an active debate.
func (a *APIAdapter) CancelDebate(debateID string) error {
	return a.orchestrator.CancelDebate(debateID)
}

// GetStatistics returns orchestrator statistics in API-friendly format.
func (a *APIAdapter) GetStatistics(ctx context.Context) (*APIStatistics, error) {
	stats, err := a.orchestrator.GetStatistics(ctx)
	if err != nil {
		return nil, err
	}

	return &APIStatistics{
		ActiveDebates:       stats.ActiveDebates,
		RegisteredAgents:    stats.RegisteredAgents,
		TotalLessons:        stats.TotalLessons,
		TotalPatterns:       stats.TotalPatterns,
		TotalDebatesLearned: stats.TotalDebatesLearned,
		OverallSuccessRate:  stats.OverallSuccessRate,
	}, nil
}

// APIStatistics represents statistics in API format.
type APIStatistics struct {
	ActiveDebates       int     `json:"active_debates"`
	RegisteredAgents    int     `json:"registered_agents"`
	TotalLessons        int     `json:"total_lessons"`
	TotalPatterns       int     `json:"total_patterns"`
	TotalDebatesLearned int     `json:"total_debates_learned"`
	OverallSuccessRate  float64 `json:"overall_success_rate"`
}

// =============================================================================
// Helper Functions
// =============================================================================

// inferDomainFromRole maps participant roles to domains.
func inferDomainFromRole(role string) agents.Domain {
	switch role {
	case "security_analyst", "security":
		return agents.DomainSecurity
	case "architect", "designer":
		return agents.DomainArchitecture
	case "optimizer", "performance":
		return agents.DomainOptimization
	case "debugger", "troubleshooter":
		return agents.DomainDebug
	case "coder", "developer", "programmer":
		return agents.DomainCode
	case "analyst", "researcher", "reasoner":
		return agents.DomainReasoning
	default:
		return agents.DomainGeneral
	}
}

// selectTopologyFromStrategy maps strategy names to topology types.
func selectTopologyFromStrategy(strategy string) topology.TopologyType {
	switch strategy {
	case "sequential", "chain", "pipeline":
		return topology.TopologyChain
	case "star", "hub", "centralized":
		return topology.TopologyStar
	case "mesh", "parallel", "distributed":
		return topology.TopologyGraphMesh
	default:
		return topology.TopologyGraphMesh // Default
	}
}
