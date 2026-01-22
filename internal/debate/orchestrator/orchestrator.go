// Package orchestrator provides the debate orchestrator that bridges
// the new debate framework with existing HelixAgent services.
package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"dev.helix.agent/internal/debate"
	"dev.helix.agent/internal/debate/agents"
	"dev.helix.agent/internal/debate/knowledge"
	"dev.helix.agent/internal/debate/protocol"
	"dev.helix.agent/internal/debate/topology"
	"dev.helix.agent/internal/debate/voting"
	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/models"
)

// Orchestrator coordinates the complete AI debate system.
// It integrates topology, protocol, specialized agents, knowledge learning,
// and voting to conduct research-backed debates.
type Orchestrator struct {
	// Provider management
	providerRegistry ProviderRegistry
	verifierScores   map[string]float64 // provider/model -> LLMsVerifier score

	// Agent management
	agentFactory *agents.AgentFactory
	agentPool    *agents.AgentPool
	teamBuilder  *agents.TeamBuilder

	// Knowledge management
	knowledgeRepo       knowledge.Repository
	learningIntegration *knowledge.DebateLearningIntegration
	crossDebateLearner  *knowledge.CrossDebateLearner

	// Voting
	votingSystem *voting.WeightedVotingSystem

	// Configuration
	config OrchestratorConfig

	// State
	activeDebates map[string]*ActiveDebate
	mu            sync.RWMutex
}

// ProviderRegistry is the interface for getting LLM providers.
type ProviderRegistry interface {
	GetProvider(name string) (llm.LLMProvider, error)
	GetAvailableProviders() []string
}

// OrchestratorConfig configures the debate orchestrator.
type OrchestratorConfig struct {
	// Default debate settings
	DefaultMaxRounds    int                   `json:"default_max_rounds"`
	DefaultTimeout      time.Duration         `json:"default_timeout"`
	DefaultTopology     topology.TopologyType `json:"default_topology"`
	DefaultMinConsensus float64               `json:"default_min_consensus"`

	// Agent settings
	MinAgentsPerDebate   int  `json:"min_agents_per_debate"`
	MaxAgentsPerDebate   int  `json:"max_agents_per_debate"`
	EnableAgentDiversity bool `json:"enable_agent_diversity"`

	// Learning settings
	EnableLearning            bool    `json:"enable_learning"`
	EnableCrossDebateLearning bool    `json:"enable_cross_debate_learning"`
	MinConsensusForLesson     float64 `json:"min_consensus_for_lesson"`

	// Voting settings
	VotingMethod              voting.VotingMethod `json:"voting_method"`
	EnableConfidenceWeighting bool                `json:"enable_confidence_weighting"`
}

// DefaultOrchestratorConfig returns sensible defaults.
func DefaultOrchestratorConfig() OrchestratorConfig {
	return OrchestratorConfig{
		DefaultMaxRounds:          3,
		DefaultTimeout:            5 * time.Minute,
		DefaultTopology:           topology.TopologyGraphMesh,
		DefaultMinConsensus:       0.75,
		MinAgentsPerDebate:        3,
		MaxAgentsPerDebate:        10,
		EnableAgentDiversity:      true,
		EnableLearning:            true,
		EnableCrossDebateLearning: true,
		MinConsensusForLesson:     0.7,
		VotingMethod:              voting.VotingMethodWeighted,
		EnableConfidenceWeighting: true,
	}
}

// ActiveDebate tracks an active debate session.
type ActiveDebate struct {
	ID              string
	Config          *DebateRequest
	Protocol        *protocol.Protocol
	Agents          []*agents.SpecializedAgent
	TopologyAgents  []*topology.Agent
	LearningSession *knowledge.DebateLearningSession
	StartTime       time.Time
	Status          DebateStatus
}

// DebateStatus represents the status of a debate.
type DebateStatus string

const (
	DebateStatusPending   DebateStatus = "pending"
	DebateStatusRunning   DebateStatus = "running"
	DebateStatusCompleted DebateStatus = "completed"
	DebateStatusFailed    DebateStatus = "failed"
	DebateStatusCancelled DebateStatus = "cancelled"
)

// NewOrchestrator creates a new debate orchestrator.
func NewOrchestrator(
	providerRegistry ProviderRegistry,
	lessonBank *debate.LessonBank,
	config OrchestratorConfig,
) *Orchestrator {
	// Create agent factory and pool
	agentFactory := agents.NewAgentFactory()
	agentPool := agents.NewAgentPool(agentFactory)
	teamBuilder := agents.NewTeamBuilder(agentPool)

	// Create knowledge repository
	repoConfig := knowledge.DefaultRepositoryConfig()
	knowledgeRepo := knowledge.NewDefaultRepository(lessonBank, repoConfig)

	// Create learning integration
	integrationConfig := knowledge.DefaultIntegrationConfig()
	integrationConfig.AutoExtractLessons = config.EnableLearning
	integrationConfig.MinConsensusForLesson = config.MinConsensusForLesson
	learningIntegration := knowledge.NewDebateLearningIntegration(knowledgeRepo, integrationConfig)

	// Create cross-debate learner
	learnerConfig := knowledge.DefaultLearningConfig()
	crossDebateLearner := knowledge.NewCrossDebateLearner(knowledgeRepo, learnerConfig)

	// Create voting system
	votingConfig := voting.DefaultVotingConfig()
	// Enable diversity bonus for weighted voting which considers confidence
	votingConfig.EnableDiversityBonus = config.EnableConfidenceWeighting
	votingSystem := voting.NewWeightedVotingSystem(votingConfig)

	return &Orchestrator{
		providerRegistry:    providerRegistry,
		verifierScores:      make(map[string]float64),
		agentFactory:        agentFactory,
		agentPool:           agentPool,
		teamBuilder:         teamBuilder,
		knowledgeRepo:       knowledgeRepo,
		learningIntegration: learningIntegration,
		crossDebateLearner:  crossDebateLearner,
		votingSystem:        votingSystem,
		config:              config,
		activeDebates:       make(map[string]*ActiveDebate),
	}
}

// SetVerifierScores sets the LLMsVerifier scores for providers.
func (o *Orchestrator) SetVerifierScores(scores map[string]float64) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.verifierScores = scores
}

// RegisterProvider registers a provider with its models and scores.
func (o *Orchestrator) RegisterProvider(provider, model string, score float64) error {
	key := provider + "/" + model

	o.mu.Lock()
	o.verifierScores[key] = score
	o.mu.Unlock()

	// Create specialized agent for this provider
	domain := o.inferDomainFromProvider(provider)
	agent := agents.NewSpecializedAgent(
		fmt.Sprintf("%s Agent", provider),
		provider,
		model,
		domain,
	)
	agent.SetScore(score)

	o.agentPool.Add(agent)

	return nil
}

// inferDomainFromProvider infers the best domain for a provider.
func (o *Orchestrator) inferDomainFromProvider(provider string) agents.Domain {
	switch provider {
	case "deepseek":
		return agents.DomainCode
	case "claude", "gemini":
		return agents.DomainReasoning
	case "mistral":
		return agents.DomainOptimization
	default:
		return agents.DomainGeneral
	}
}

// DebateRequest represents a request to conduct a debate.
type DebateRequest struct {
	ID                 string                 `json:"id,omitempty"`
	Topic              string                 `json:"topic"`
	Context            string                 `json:"context,omitempty"`
	Requirements       []string               `json:"requirements,omitempty"`
	MaxRounds          int                    `json:"max_rounds,omitempty"`
	Timeout            time.Duration          `json:"timeout,omitempty"`
	TopologyType       topology.TopologyType  `json:"topology_type,omitempty"`
	MinConsensus       float64                `json:"min_consensus,omitempty"`
	PreferredProviders []string               `json:"preferred_providers,omitempty"`
	PreferredDomain    agents.Domain          `json:"preferred_domain,omitempty"`
	EnableLearning     *bool                  `json:"enable_learning,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// DebateResponse represents the result of a debate.
type DebateResponse struct {
	ID               string                 `json:"id"`
	Topic            string                 `json:"topic"`
	Success          bool                   `json:"success"`
	Consensus        *ConsensusResponse     `json:"consensus,omitempty"`
	Phases           []*PhaseResponse       `json:"phases"`
	Participants     []*ParticipantInfo     `json:"participants"`
	LessonsLearned   int                    `json:"lessons_learned"`
	PatternsDetected int                    `json:"patterns_detected"`
	Duration         time.Duration          `json:"duration"`
	Metrics          *DebateMetrics         `json:"metrics"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ConsensusResponse represents the consensus from a debate.
type ConsensusResponse struct {
	Summary       string         `json:"summary"`
	Confidence    float64        `json:"confidence"`
	KeyPoints     []string       `json:"key_points"`
	Dissents      []string       `json:"dissents,omitempty"`
	VoteBreakdown map[string]int `json:"vote_breakdown"`
	WinningVote   string         `json:"winning_vote"`
	Method        string         `json:"method"`
}

// PhaseResponse represents a phase in the debate.
type PhaseResponse struct {
	Phase          string           `json:"phase"`
	Round          int              `json:"round"`
	Responses      []*AgentResponse `json:"responses"`
	ConsensusLevel float64          `json:"consensus_level"`
	KeyInsights    []string         `json:"key_insights"`
	Duration       time.Duration    `json:"duration"`
}

// AgentResponse represents a response from an agent.
type AgentResponse struct {
	AgentID    string        `json:"agent_id"`
	Provider   string        `json:"provider"`
	Model      string        `json:"model"`
	Role       string        `json:"role"`
	Content    string        `json:"content"`
	Confidence float64       `json:"confidence"`
	Score      float64       `json:"score"`
	Latency    time.Duration `json:"latency"`
}

// ParticipantInfo describes a debate participant.
type ParticipantInfo struct {
	AgentID  string        `json:"agent_id"`
	Name     string        `json:"name"`
	Provider string        `json:"provider"`
	Model    string        `json:"model"`
	Role     string        `json:"role"`
	Domain   agents.Domain `json:"domain"`
	Score    float64       `json:"score"`
}

// DebateMetrics provides metrics about the debate.
type DebateMetrics struct {
	TotalResponses    int            `json:"total_responses"`
	AvgLatency        time.Duration  `json:"avg_latency"`
	AvgConfidence     float64        `json:"avg_confidence"`
	ConsensusScore    float64        `json:"consensus_score"`
	QualityScore      float64        `json:"quality_score"`
	ProviderBreakdown map[string]int `json:"provider_breakdown"`
}

// ConductDebate conducts a complete AI debate using the research-backed framework.
func (o *Orchestrator) ConductDebate(ctx context.Context, request *DebateRequest) (*DebateResponse, error) {
	// Apply defaults
	if request.ID == "" {
		request.ID = uuid.New().String()
	}
	if request.MaxRounds == 0 {
		request.MaxRounds = o.config.DefaultMaxRounds
	}
	if request.Timeout == 0 {
		request.Timeout = o.config.DefaultTimeout
	}
	if request.TopologyType == "" {
		request.TopologyType = o.config.DefaultTopology
	}
	if request.MinConsensus == 0 {
		request.MinConsensus = o.config.DefaultMinConsensus
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, request.Timeout)
	defer cancel()

	// 1. Build agent team
	teamAgents, topoAgents, err := o.buildTeam(request)
	if err != nil {
		return nil, fmt.Errorf("failed to build agent team: %w", err)
	}

	// 2. Create topology
	topoConfig := topology.DefaultTopologyConfig(request.TopologyType)
	topo, err := topology.NewTopology(request.TopologyType, topoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create topology: %w", err)
	}
	// Initialize topology with agents
	if err := topo.Initialize(ctx, topoAgents); err != nil {
		return nil, fmt.Errorf("failed to initialize topology: %w", err)
	}

	// 3. Create agent invoker
	invoker := NewProviderInvoker(o.providerRegistry)

	// 4. Create protocol
	protocolConfig := protocol.DebateConfig{
		ID:                  request.ID,
		Topic:               request.Topic,
		Context:             request.Context,
		Requirements:        request.Requirements,
		MaxRounds:           request.MaxRounds,
		Timeout:             request.Timeout,
		TopologyType:        request.TopologyType,
		MinConsensusScore:   request.MinConsensus,
		EnableEarlyExit:     true,
		EnableCognitiveLoop: true,
		Metadata:            request.Metadata,
	}
	debateProtocol := protocol.NewProtocol(protocolConfig, topo, invoker)

	// 5. Track active debate
	activeDebate := &ActiveDebate{
		ID:             request.ID,
		Config:         request,
		Protocol:       debateProtocol,
		Agents:         teamAgents,
		TopologyAgents: topoAgents,
		StartTime:      time.Now(),
		Status:         DebateStatusRunning,
	}

	o.mu.Lock()
	o.activeDebates[request.ID] = activeDebate
	o.mu.Unlock()

	defer func() {
		o.mu.Lock()
		delete(o.activeDebates, request.ID)
		o.mu.Unlock()
	}()

	// 6. Start learning session if enabled
	enableLearning := o.config.EnableLearning
	if request.EnableLearning != nil {
		enableLearning = *request.EnableLearning
	}

	var learningSession *knowledge.DebateLearningSession
	if enableLearning {
		learningSession, err = o.learningIntegration.StartDebateLearning(
			ctx, request.ID, request.Topic, teamAgents)
		if err == nil {
			activeDebate.LearningSession = learningSession
		}
	}

	// 7. Execute the debate
	result, err := debateProtocol.Execute(ctx)
	if err != nil {
		activeDebate.Status = DebateStatusFailed
		return nil, fmt.Errorf("debate execution failed: %w", err)
	}

	// 8. Complete learning
	var learningResult *knowledge.DebateLearningResult
	lessonsLearned := 0
	patternsDetected := 0

	if enableLearning && learningSession != nil {
		learningResult, _ = o.learningIntegration.OnDebateComplete(ctx, result)
		if learningResult != nil {
			lessonsLearned = learningResult.ExtractedLessons
			patternsDetected = learningResult.DetectedPatterns
		}

		// Cross-debate learning
		if o.config.EnableCrossDebateLearning && learningResult != nil && learningResult.Lessons != nil {
			_, _ = o.crossDebateLearner.LearnFromDebate(ctx, result, learningResult.Lessons)
		}
	}

	// 9. Build response
	activeDebate.Status = DebateStatusCompleted
	response := o.buildResponse(result, teamAgents, lessonsLearned, patternsDetected)

	return response, nil
}

// buildTeam builds the agent team for the debate.
func (o *Orchestrator) buildTeam(request *DebateRequest) ([]*agents.SpecializedAgent, []*topology.Agent, error) {
	// Check pool size
	if o.agentPool.Size() < o.config.MinAgentsPerDebate {
		return nil, nil, fmt.Errorf("not enough agents in pool: have %d, need %d",
			o.agentPool.Size(), o.config.MinAgentsPerDebate)
	}

	// Build team configuration
	teamConfig := agents.DefaultTeamConfig()
	if request.PreferredDomain != "" {
		// Prioritize preferred domain for key roles
		teamConfig.PreferredDomains[topology.RoleProposer] = request.PreferredDomain
		teamConfig.PreferredDomains[topology.RoleCritic] = request.PreferredDomain
	}

	// Use team builder
	assignments, err := o.teamBuilder.BuildTeam(teamConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("team building failed: %w", err)
	}

	// Extract agents and topology agents
	teamAgents := make([]*agents.SpecializedAgent, 0, len(assignments))
	topoAgents := make([]*topology.Agent, 0, len(assignments))

	seenAgents := make(map[string]bool)
	for _, assignment := range assignments {
		if seenAgents[assignment.Agent.ID] {
			continue
		}
		seenAgents[assignment.Agent.ID] = true

		teamAgents = append(teamAgents, assignment.Agent)

		topoAgent := assignment.Agent.ToTopologyAgent()
		topoAgent.Role = assignment.Role
		topoAgents = append(topoAgents, topoAgent)
	}

	return teamAgents, topoAgents, nil
}

// buildResponse builds the debate response from the protocol result.
func (o *Orchestrator) buildResponse(
	result *protocol.DebateResult,
	teamAgents []*agents.SpecializedAgent,
	lessonsLearned, patternsDetected int,
) *DebateResponse {
	response := &DebateResponse{
		ID:               result.ID,
		Topic:            result.Topic,
		Success:          result.Success,
		Phases:           make([]*PhaseResponse, 0, len(result.Phases)),
		Participants:     make([]*ParticipantInfo, 0, len(teamAgents)),
		LessonsLearned:   lessonsLearned,
		PatternsDetected: patternsDetected,
		Duration:         result.Duration,
		Metadata:         result.Metadata,
	}

	// Build consensus response
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

	// Build phase responses
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

	// Build participant info
	for _, agent := range teamAgents {
		response.Participants = append(response.Participants, &ParticipantInfo{
			AgentID:  agent.ID,
			Name:     agent.Name,
			Provider: agent.Provider,
			Model:    agent.Model,
			Role:     string(agent.PrimaryRole),
			Domain:   agent.Specialization.PrimaryDomain,
			Score:    agent.Score,
		})
	}

	// Build metrics
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

// GetActiveDebates returns all active debates.
func (o *Orchestrator) GetActiveDebates() []*ActiveDebate {
	o.mu.RLock()
	defer o.mu.RUnlock()

	debates := make([]*ActiveDebate, 0, len(o.activeDebates))
	for _, d := range o.activeDebates {
		debates = append(debates, d)
	}
	return debates
}

// GetDebateStatus returns the status of a debate.
func (o *Orchestrator) GetDebateStatus(debateID string) (DebateStatus, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	debate, ok := o.activeDebates[debateID]
	if !ok {
		return "", false
	}
	return debate.Status, true
}

// CancelDebate cancels an active debate.
func (o *Orchestrator) CancelDebate(debateID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	debate, ok := o.activeDebates[debateID]
	if !ok {
		return fmt.Errorf("debate not found: %s", debateID)
	}

	debate.Status = DebateStatusCancelled
	return nil
}

// GetRecommendations gets learning-based recommendations for a topic.
func (o *Orchestrator) GetRecommendations(ctx context.Context, topic string, domain agents.Domain) (*knowledge.DebateRecommendations, error) {
	return o.crossDebateLearner.GetRecommendations(ctx, topic, domain)
}

// GetAgentPool returns the agent pool for inspection.
func (o *Orchestrator) GetAgentPool() *agents.AgentPool {
	return o.agentPool
}

// GetKnowledgeRepository returns the knowledge repository.
func (o *Orchestrator) GetKnowledgeRepository() knowledge.Repository {
	return o.knowledgeRepo
}

// GetStatistics returns orchestrator statistics.
func (o *Orchestrator) GetStatistics(ctx context.Context) (*OrchestratorStatistics, error) {
	repoStats, err := o.knowledgeRepo.GetStatistics(ctx)
	if err != nil {
		return nil, err
	}

	o.mu.RLock()
	activeCount := len(o.activeDebates)
	o.mu.RUnlock()

	return &OrchestratorStatistics{
		ActiveDebates:       activeCount,
		RegisteredAgents:    o.agentPool.Size(),
		TotalLessons:        repoStats.TotalLessons,
		TotalPatterns:       repoStats.TotalPatterns,
		TotalDebatesLearned: repoStats.TotalDebates,
		OverallSuccessRate:  repoStats.OverallSuccessRate,
	}, nil
}

// OrchestratorStatistics provides statistics about the orchestrator.
type OrchestratorStatistics struct {
	ActiveDebates       int     `json:"active_debates"`
	RegisteredAgents    int     `json:"registered_agents"`
	TotalLessons        int     `json:"total_lessons"`
	TotalPatterns       int     `json:"total_patterns"`
	TotalDebatesLearned int     `json:"total_debates_learned"`
	OverallSuccessRate  float64 `json:"overall_success_rate"`
}

// =============================================================================
// Provider Invoker - Implements protocol.AgentInvoker
// =============================================================================

// ProviderInvoker implements the AgentInvoker interface using ProviderRegistry.
type ProviderInvoker struct {
	registry ProviderRegistry
}

// NewProviderInvoker creates a new provider invoker.
func NewProviderInvoker(registry ProviderRegistry) *ProviderInvoker {
	return &ProviderInvoker{registry: registry}
}

// Invoke requests a response from an agent using the provider registry.
func (pi *ProviderInvoker) Invoke(
	ctx context.Context,
	agent *topology.Agent,
	prompt string,
	debateCtx protocol.DebateContext,
) (*protocol.PhaseResponse, error) {
	startTime := time.Now()

	// Get provider
	provider, err := pi.registry.GetProvider(agent.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider %s: %w", agent.Provider, err)
	}

	// Build system prompt
	systemPrompt := buildSystemPrompt(agent, debateCtx)

	// Build full prompt with context
	fullPrompt := buildFullPrompt(prompt, debateCtx)

	// Create LLM request
	request := &models.LLMRequest{
		Messages: []models.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: fullPrompt},
		},
		ModelParams: models.ModelParameters{
			Model:       agent.Model,
			Temperature: 0.7,
			MaxTokens:   2048,
		},
	}

	// Call provider
	llmResponse, err := provider.Complete(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("provider call failed: %w", err)
	}

	latency := time.Since(startTime)

	// Build response
	response := &protocol.PhaseResponse{
		AgentID:    agent.ID,
		Role:       agent.Role,
		Provider:   agent.Provider,
		Model:      agent.Model,
		Content:    llmResponse.Content,
		Confidence: calculateConfidence(llmResponse),
		Score:      agent.Score,
		Latency:    latency,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"tokens_used":   llmResponse.TokensUsed,
			"finish_reason": llmResponse.FinishReason,
		},
	}

	// Extract arguments/criticisms based on phase
	response.Arguments = extractArguments(llmResponse.Content)
	if debateCtx.CurrentPhase == topology.PhaseCritique {
		response.Criticisms = extractCriticisms(llmResponse.Content)
	}
	if debateCtx.CurrentPhase == topology.PhaseOptimization {
		response.Suggestions = extractSuggestions(llmResponse.Content)
	}

	return response, nil
}

// buildSystemPrompt builds the system prompt for an agent.
func buildSystemPrompt(agent *topology.Agent, debateCtx protocol.DebateContext) string {
	roleDescriptions := map[topology.AgentRole]string{
		topology.RoleProposer:  "You are a proposer who generates creative solutions and initial ideas.",
		topology.RoleCritic:    "You are a critic who identifies weaknesses and potential issues.",
		topology.RoleReviewer:  "You are a reviewer who evaluates completeness and correctness.",
		topology.RoleOptimizer: "You are an optimizer who improves efficiency and performance.",
		topology.RoleModerator: "You are a moderator who facilitates discussion and builds consensus.",
		topology.RoleArchitect: "You are an architect who focuses on system design and scalability.",
		topology.RoleValidator: "You are a validator who verifies accuracy and alignment.",
		topology.RoleRedTeam:   "You are a red team member who stress-tests ideas and finds vulnerabilities.",
		topology.RoleBlueTeam:  "You are a blue team member who defends and strengthens proposals.",
	}

	phaseInstructions := map[topology.DebatePhase]string{
		topology.PhaseProposal:     "Generate a thoughtful proposal addressing the topic. Include your reasoning and key points.",
		topology.PhaseCritique:     "Critically analyze the proposals. Identify strengths, weaknesses, and potential issues.",
		topology.PhaseReview:       "Review all contributions for completeness and correctness. Provide constructive feedback.",
		topology.PhaseOptimization: "Suggest optimizations and improvements. Focus on efficiency and best practices.",
		topology.PhaseConvergence:  "Work toward consensus. Summarize key points and identify the best approach.",
	}

	roleDesc := roleDescriptions[agent.Role]
	if roleDesc == "" {
		roleDesc = "You are a debate participant providing thoughtful analysis."
	}

	phaseInst := phaseInstructions[debateCtx.CurrentPhase]
	if phaseInst == "" {
		phaseInst = "Provide your analysis and insights on the topic."
	}

	return fmt.Sprintf(`%s

Current Phase: %s
%s

Guidelines:
- Be concise but thorough
- Support your points with reasoning
- Consider multiple perspectives
- Be constructive and collaborative
- End your response with a confidence score (0-100%%)`,
		roleDesc, debateCtx.CurrentPhase, phaseInst)
}

// buildFullPrompt builds the full prompt with debate context.
func buildFullPrompt(prompt string, debateCtx protocol.DebateContext) string {
	var fullPrompt string

	fullPrompt = fmt.Sprintf("Topic: %s\n\n", debateCtx.Topic)

	if debateCtx.Context != "" {
		fullPrompt += fmt.Sprintf("Context: %s\n\n", debateCtx.Context)
	}

	// Include previous phase summaries
	if len(debateCtx.PreviousPhases) > 0 {
		fullPrompt += "Previous Discussion:\n"
		for _, phase := range debateCtx.PreviousPhases {
			if len(phase.KeyInsights) > 0 {
				fullPrompt += fmt.Sprintf("- %s: %v\n", phase.Phase, phase.KeyInsights)
			}
		}
		fullPrompt += "\n"
	}

	fullPrompt += fmt.Sprintf("Round %d - %s\n\n", debateCtx.Round, prompt)

	return fullPrompt
}

// calculateConfidence extracts or estimates confidence from response.
func calculateConfidence(response *models.LLMResponse) float64 {
	// Default confidence based on response quality
	confidence := 0.7

	// Adjust based on content length (longer = more confident usually)
	contentLen := len(response.Content)
	if contentLen > 1000 {
		confidence += 0.1
	} else if contentLen < 200 {
		confidence -= 0.1
	}

	// Adjust based on finish reason
	if response.FinishReason == "stop" {
		confidence += 0.05
	}

	// Cap at 0.95
	if confidence > 0.95 {
		confidence = 0.95
	}
	if confidence < 0.3 {
		confidence = 0.3
	}

	return confidence
}

// extractArguments extracts key arguments from content.
func extractArguments(content string) []string {
	// Simple extraction - in production, use NLP
	arguments := make([]string, 0)

	// Look for bullet points or numbered items
	lines := splitLines(content)
	for _, line := range lines {
		if len(line) > 10 && (line[0] == '-' || line[0] == '*' || (line[0] >= '1' && line[0] <= '9')) {
			arguments = append(arguments, trimLine(line))
		}
	}

	// Limit to top 5
	if len(arguments) > 5 {
		arguments = arguments[:5]
	}

	return arguments
}

// extractCriticisms extracts criticisms from content.
func extractCriticisms(content string) []string {
	// Similar to extractArguments but focused on critique language
	return extractArguments(content) // Simplified for now
}

// extractSuggestions extracts suggestions from content.
func extractSuggestions(content string) []string {
	// Similar to extractArguments but focused on suggestion language
	return extractArguments(content) // Simplified for now
}

// Helper functions
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimLine(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '-' || s[start] == '*') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}

// Ensure ProviderInvoker implements AgentInvoker
var _ protocol.AgentInvoker = (*ProviderInvoker)(nil)
