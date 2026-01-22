// Package protocol provides the 5-phase debate protocol implementation.
// Implements: Proposal → Critique → Review → Optimization → Convergence
// Based on research findings from ACL 2025 MARBLE framework and Kimi studies.
package protocol

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"dev.helix.agent/internal/debate/topology"
	"github.com/google/uuid"
)

// DebateConfig configures a debate session.
type DebateConfig struct {
	ID                  string                 `json:"id"`
	Topic               string                 `json:"topic"`
	Context             string                 `json:"context"`
	Requirements        []string               `json:"requirements,omitempty"`
	MaxRounds           int                    `json:"max_rounds"`
	Timeout             time.Duration          `json:"timeout"`
	TopologyType        topology.TopologyType  `json:"topology_type"`
	MinConsensusScore   float64                `json:"min_consensus_score"`
	EnableEarlyExit     bool                   `json:"enable_early_exit"`
	EnableCognitiveLoop bool                   `json:"enable_cognitive_loop"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// DefaultDebateConfig returns a sensible default configuration.
func DefaultDebateConfig() DebateConfig {
	return DebateConfig{
		ID:                  uuid.New().String(),
		MaxRounds:           3,
		Timeout:             5 * time.Minute,
		TopologyType:        topology.TopologyGraphMesh,
		MinConsensusScore:   0.75,
		EnableEarlyExit:     true,
		EnableCognitiveLoop: true,
		Metadata:            make(map[string]interface{}),
	}
}

// PhaseConfig configures a single debate phase.
type PhaseConfig struct {
	Phase           topology.DebatePhase `json:"phase"`
	Timeout         time.Duration        `json:"timeout"`
	MinResponses    int                  `json:"min_responses"`
	MaxParallelism  int                  `json:"max_parallelism"`
	RequiredRoles   []topology.AgentRole `json:"required_roles"`
	Prompt          string               `json:"prompt"`
	ValidationRules []ValidationRule     `json:"validation_rules,omitempty"`
}

// ValidationRule defines a rule for validating phase outputs.
type ValidationRule struct {
	Name        string                             `json:"name"`
	Description string                             `json:"description"`
	Required    bool                               `json:"required"`
	Validator   func(response *PhaseResponse) bool `json:"-"`
}

// PhaseResponse represents a response from an agent in a phase.
type PhaseResponse struct {
	AgentID     string                 `json:"agent_id"`
	Role        topology.AgentRole     `json:"role"`
	Provider    string                 `json:"provider"`
	Model       string                 `json:"model"`
	Content     string                 `json:"content"`
	Confidence  float64                `json:"confidence"`
	Arguments   []string               `json:"arguments,omitempty"`
	Criticisms  []string               `json:"criticisms,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Vote        string                 `json:"vote,omitempty"` // For convergence
	Score       float64                `json:"score"`
	Latency     time.Duration          `json:"latency"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PhaseResult represents the result of a single phase.
type PhaseResult struct {
	Phase          topology.DebatePhase `json:"phase"`
	Round          int                  `json:"round"`
	Responses      []*PhaseResponse     `json:"responses"`
	LeaderResponse *PhaseResponse       `json:"leader_response,omitempty"`
	ConsensusLevel float64              `json:"consensus_level"`
	KeyInsights    []string             `json:"key_insights"`
	Disagreements  []string             `json:"disagreements,omitempty"`
	StartTime      time.Time            `json:"start_time"`
	EndTime        time.Time            `json:"end_time"`
	Duration       time.Duration        `json:"duration"`
}

// DebateResult represents the complete result of a debate.
type DebateResult struct {
	ID               string                 `json:"id"`
	Topic            string                 `json:"topic"`
	Phases           []*PhaseResult         `json:"phases"`
	FinalConsensus   *ConsensusResult       `json:"final_consensus"`
	BestResponse     *PhaseResponse         `json:"best_response"`
	ParticipantCount int                    `json:"participant_count"`
	TotalRounds      int                    `json:"total_rounds"`
	RoundsCompleted  int                    `json:"rounds_completed"`
	Success          bool                   `json:"success"`
	EarlyExit        bool                   `json:"early_exit"`
	EarlyExitReason  string                 `json:"early_exit_reason,omitempty"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	Duration         time.Duration          `json:"duration"`
	TopologyUsed     topology.TopologyType  `json:"topology_used"`
	Metrics          *DebateMetrics         `json:"metrics"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ConsensusResult represents the final consensus from a debate.
type ConsensusResult struct {
	Summary       string          `json:"summary"`
	Confidence    float64         `json:"confidence"`
	KeyPoints     []string        `json:"key_points"`
	Dissents      []string        `json:"dissents,omitempty"`
	VoteBreakdown map[string]int  `json:"vote_breakdown"`
	WinningVote   string          `json:"winning_vote"`
	Contributors  []string        `json:"contributors"`
	Method        ConsensusMethod `json:"method"`
}

// ConsensusMethod identifies how consensus was reached.
type ConsensusMethod string

const (
	ConsensusMethodUnanimous      ConsensusMethod = "unanimous"
	ConsensusMethodMajority       ConsensusMethod = "majority"
	ConsensusMethodWeightedVoting ConsensusMethod = "weighted_voting"
	ConsensusMethodLeaderDecision ConsensusMethod = "leader_decision"
	ConsensusMethodNoConsensus    ConsensusMethod = "no_consensus"
)

// DebateMetrics holds metrics for a debate.
type DebateMetrics struct {
	TotalResponses     int                                    `json:"total_responses"`
	TotalTokens        int                                    `json:"total_tokens"`
	AvgLatency         time.Duration                          `json:"avg_latency"`
	AvgConfidence      float64                                `json:"avg_confidence"`
	ConsensusScore     float64                                `json:"consensus_score"`
	PhaseMetrics       map[topology.DebatePhase]*PhaseMetrics `json:"phase_metrics"`
	AgentParticipation map[string]int                         `json:"agent_participation"`
	RoleContributions  map[topology.AgentRole]int             `json:"role_contributions"`
}

// PhaseMetrics holds metrics for a single phase.
type PhaseMetrics struct {
	ResponseCount  int           `json:"response_count"`
	AvgLatency     time.Duration `json:"avg_latency"`
	AvgConfidence  float64       `json:"avg_confidence"`
	ConsensusLevel float64       `json:"consensus_level"`
}

// AgentInvoker is the interface for invoking agent responses.
type AgentInvoker interface {
	// Invoke requests a response from an agent.
	Invoke(ctx context.Context, agent *topology.Agent, prompt string, context DebateContext) (*PhaseResponse, error)
}

// DebateContext provides context for agent invocations.
type DebateContext struct {
	DebateID       string                 `json:"debate_id"`
	Topic          string                 `json:"topic"`
	Context        string                 `json:"context"`
	CurrentPhase   topology.DebatePhase   `json:"current_phase"`
	Round          int                    `json:"round"`
	PreviousPhases []*PhaseResult         `json:"previous_phases,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Protocol orchestrates the 5-phase debate flow.
type Protocol struct {
	config       DebateConfig
	topology     topology.Topology
	invoker      AgentInvoker
	phaseConfigs map[topology.DebatePhase]*PhaseConfig

	// State
	currentPhase topology.DebatePhase
	currentRound int
	phaseResults []*PhaseResult
	metrics      *DebateMetrics

	mu      sync.RWMutex
	started bool
	stopped bool
}

// NewProtocol creates a new debate protocol.
func NewProtocol(config DebateConfig, topo topology.Topology, invoker AgentInvoker) *Protocol {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}

	p := &Protocol{
		config:       config,
		topology:     topo,
		invoker:      invoker,
		phaseConfigs: make(map[topology.DebatePhase]*PhaseConfig),
		phaseResults: make([]*PhaseResult, 0),
		metrics: &DebateMetrics{
			PhaseMetrics:       make(map[topology.DebatePhase]*PhaseMetrics),
			AgentParticipation: make(map[string]int),
			RoleContributions:  make(map[topology.AgentRole]int),
		},
	}

	// Initialize default phase configs
	p.initializePhaseConfigs()

	return p
}

// initializePhaseConfigs sets up default configurations for each phase.
func (p *Protocol) initializePhaseConfigs() {
	phaseTimeout := p.config.Timeout / 5 // Divide total timeout among phases

	// Phase 1: Proposal
	p.phaseConfigs[topology.PhaseProposal] = &PhaseConfig{
		Phase:          topology.PhaseProposal,
		Timeout:        phaseTimeout,
		MinResponses:   1,
		MaxParallelism: 4,
		RequiredRoles:  []topology.AgentRole{topology.RoleProposer},
		Prompt: `You are in the PROPOSAL phase of an AI debate.

Topic: {{.Topic}}
Context: {{.Context}}

Your role is to propose an initial solution or approach. Be creative but practical.
Consider multiple perspectives and provide clear reasoning for your proposal.

Output your proposal with:
1. A clear solution/approach
2. Key supporting arguments
3. Potential challenges you foresee
4. Your confidence level (0-1)`,
	}

	// Phase 2: Critique
	p.phaseConfigs[topology.PhaseCritique] = &PhaseConfig{
		Phase:          topology.PhaseCritique,
		Timeout:        phaseTimeout,
		MinResponses:   2,
		MaxParallelism: 6,
		RequiredRoles:  []topology.AgentRole{topology.RoleCritic, topology.RoleRedTeam},
		Prompt: `You are in the CRITIQUE phase of an AI debate.

Topic: {{.Topic}}
Previous Proposals:
{{.PreviousContent}}

Your role is to critically analyze the proposals. Identify:
1. Logical weaknesses or flaws
2. Missing considerations
3. Potential risks or failure modes
4. Areas needing clarification

Be constructive but thorough. Don't just criticize - explain why something is problematic.`,
	}

	// Phase 3: Review
	p.phaseConfigs[topology.PhaseReview] = &PhaseConfig{
		Phase:          topology.PhaseReview,
		Timeout:        phaseTimeout,
		MinResponses:   2,
		MaxParallelism: 4,
		RequiredRoles:  []topology.AgentRole{topology.RoleReviewer, topology.RoleArchitect},
		Prompt: `You are in the REVIEW phase of an AI debate.

Topic: {{.Topic}}
Proposals: {{.Proposals}}
Critiques: {{.Critiques}}

Your role is to evaluate the quality of proposals considering the critiques.
1. Which proposals address the critiques best?
2. What strengths remain valid?
3. What modifications would improve the proposals?
4. Rate each proposal's viability (0-1)`,
	}

	// Phase 4: Optimization
	p.phaseConfigs[topology.PhaseOptimization] = &PhaseConfig{
		Phase:          topology.PhaseOptimization,
		Timeout:        phaseTimeout,
		MinResponses:   1,
		MaxParallelism: 4,
		RequiredRoles:  []topology.AgentRole{topology.RoleOptimizer, topology.RoleBlueTeam},
		Prompt: `You are in the OPTIMIZATION phase of an AI debate.

Topic: {{.Topic}}
Original Proposals: {{.Proposals}}
Critiques: {{.Critiques}}
Reviews: {{.Reviews}}

Your role is to synthesize and improve the best ideas:
1. Combine the strongest elements from proposals
2. Address the valid critiques
3. Incorporate review suggestions
4. Produce an optimized solution

Focus on practical improvements, not theoretical perfection.`,
	}

	// Phase 5: Convergence
	p.phaseConfigs[topology.PhaseConvergence] = &PhaseConfig{
		Phase:          topology.PhaseConvergence,
		Timeout:        phaseTimeout,
		MinResponses:   3,
		MaxParallelism: 8,
		RequiredRoles:  []topology.AgentRole{topology.RoleModerator, topology.RoleValidator},
		Prompt: `You are in the CONVERGENCE phase of an AI debate.

Topic: {{.Topic}}
Optimized Solutions: {{.Optimizations}}

All previous context:
{{.FullContext}}

Your role is to vote and build consensus:
1. Evaluate the optimized solutions
2. Vote for the best solution
3. Identify remaining points of disagreement
4. Suggest any final refinements

Provide:
- VOTE: [solution identifier]
- CONFIDENCE: [0-1]
- REASONING: [brief explanation]
- REFINEMENTS: [any final suggestions]`,
	}
}

// SetPhaseConfig sets a custom configuration for a phase.
func (p *Protocol) SetPhaseConfig(phase topology.DebatePhase, config *PhaseConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.phaseConfigs[phase] = config
}

// Execute runs the complete 5-phase debate protocol.
func (p *Protocol) Execute(ctx context.Context) (*DebateResult, error) {
	p.mu.Lock()
	if p.started {
		p.mu.Unlock()
		return nil, fmt.Errorf("protocol already started")
	}
	p.started = true
	p.mu.Unlock()

	startTime := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	defer cancel()

	// Execute rounds
	for round := 1; round <= p.config.MaxRounds; round++ {
		p.currentRound = round

		// Execute all 5 phases
		phases := []topology.DebatePhase{
			topology.PhaseProposal,
			topology.PhaseCritique,
			topology.PhaseReview,
			topology.PhaseOptimization,
			topology.PhaseConvergence,
		}

		for _, phase := range phases {
			select {
			case <-ctx.Done():
				return p.buildResult(startTime, false, "timeout"), ctx.Err()
			default:
			}

			result, err := p.executePhase(ctx, phase)
			if err != nil {
				// Log but continue if possible
				continue
			}

			p.mu.Lock()
			p.phaseResults = append(p.phaseResults, result)
			p.mu.Unlock()

			// Check for early exit after convergence
			if phase == topology.PhaseConvergence && p.config.EnableEarlyExit {
				if result.ConsensusLevel >= p.config.MinConsensusScore {
					return p.buildResult(startTime, true, "early_consensus"), nil
				}
			}
		}
	}

	return p.buildResult(startTime, true, ""), nil
}

// executePhase executes a single debate phase.
func (p *Protocol) executePhase(ctx context.Context, phase topology.DebatePhase) (*PhaseResult, error) {
	config := p.phaseConfigs[phase]
	if config == nil {
		return nil, fmt.Errorf("no configuration for phase: %s", phase)
	}

	startTime := time.Now()

	// Create phase context
	phaseCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	// Update topology phase
	if gm, ok := p.topology.(*topology.GraphMeshTopology); ok {
		gm.SetPhase(phase, "protocol", "phase_transition")
	}

	// Select leader for this phase
	leader, err := p.topology.SelectLeader(phase)
	if err != nil {
		leader = nil // Continue without leader
	}

	// Get agents for this phase
	agents := p.getPhaseAgents(phase, config)

	// Collect responses in parallel
	responses := p.collectResponses(phaseCtx, agents, phase)

	// Get leader's response if available
	var leaderResponse *PhaseResponse
	if leader != nil {
		for _, r := range responses {
			if r.AgentID == leader.ID {
				leaderResponse = r
				break
			}
		}
	}

	// Calculate consensus
	consensusLevel := p.calculateConsensus(responses)

	// Extract key insights
	insights := p.extractInsights(responses)

	// Find disagreements
	disagreements := p.findDisagreements(responses)

	endTime := time.Now()

	result := &PhaseResult{
		Phase:          phase,
		Round:          p.currentRound,
		Responses:      responses,
		LeaderResponse: leaderResponse,
		ConsensusLevel: consensusLevel,
		KeyInsights:    insights,
		Disagreements:  disagreements,
		StartTime:      startTime,
		EndTime:        endTime,
		Duration:       endTime.Sub(startTime),
	}

	// Update metrics
	p.updatePhaseMetrics(phase, result)

	return result, nil
}

// getPhaseAgents returns agents suitable for the given phase.
func (p *Protocol) getPhaseAgents(phase topology.DebatePhase, config *PhaseConfig) []*topology.Agent {
	var agents []*topology.Agent

	// Get agents by required roles
	for _, role := range config.RequiredRoles {
		roleAgents := p.topology.GetAgentsByRole(role)
		agents = append(agents, roleAgents...)
	}

	// If no specific roles found, use parallel groups
	if len(agents) == 0 {
		groups := p.topology.GetParallelGroups(phase)
		for _, group := range groups {
			agents = append(agents, group...)
		}
	}

	// If still empty, use all agents
	if len(agents) == 0 {
		agents = p.topology.GetAgents()
	}

	// Limit to max parallelism
	if config.MaxParallelism > 0 && len(agents) > config.MaxParallelism {
		// Sort by score and take top agents
		sort.Slice(agents, func(i, j int) bool {
			return agents[i].Score > agents[j].Score
		})
		agents = agents[:config.MaxParallelism]
	}

	return agents
}

// collectResponses collects responses from agents in parallel.
func (p *Protocol) collectResponses(ctx context.Context, agents []*topology.Agent, phase topology.DebatePhase) []*PhaseResponse {
	responses := make([]*PhaseResponse, 0, len(agents))
	responseChan := make(chan *PhaseResponse, len(agents))

	var wg sync.WaitGroup

	debateCtx := p.buildDebateContext(phase)

	config := p.phaseConfigs[phase]
	prompt := config.Prompt

	for _, agent := range agents {
		wg.Add(1)
		go func(a *topology.Agent) {
			defer wg.Done()

			resp, err := p.invoker.Invoke(ctx, a, prompt, debateCtx)
			if err != nil {
				// Create error response
				resp = &PhaseResponse{
					AgentID:    a.ID,
					Role:       a.Role,
					Provider:   a.Provider,
					Model:      a.Model,
					Content:    fmt.Sprintf("Error: %v", err),
					Confidence: 0,
					Timestamp:  time.Now(),
					Metadata:   map[string]interface{}{"error": err.Error()},
				}
			}

			select {
			case responseChan <- resp:
			case <-ctx.Done():
				return
			}
		}(agent)
	}

	// Wait for all agents or timeout
	go func() {
		wg.Wait()
		close(responseChan)
	}()

	// Collect responses
	for resp := range responseChan {
		responses = append(responses, resp)

		// Update participation metrics
		p.mu.Lock()
		p.metrics.AgentParticipation[resp.AgentID]++
		p.metrics.RoleContributions[resp.Role]++
		p.mu.Unlock()
	}

	return responses
}

// buildDebateContext creates the context for agent invocations.
func (p *Protocol) buildDebateContext(phase topology.DebatePhase) DebateContext {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return DebateContext{
		DebateID:       p.config.ID,
		Topic:          p.config.Topic,
		Context:        p.config.Context,
		CurrentPhase:   phase,
		Round:          p.currentRound,
		PreviousPhases: p.phaseResults,
		Metadata:       p.config.Metadata,
	}
}

// calculateConsensus calculates the consensus level from responses.
func (p *Protocol) calculateConsensus(responses []*PhaseResponse) float64 {
	if len(responses) == 0 {
		return 0
	}

	// Method 1: Average confidence
	totalConfidence := 0.0
	for _, r := range responses {
		totalConfidence += r.Confidence
	}
	avgConfidence := totalConfidence / float64(len(responses))

	// Method 2: Vote agreement (for convergence phase)
	voteAgreement := p.calculateVoteAgreement(responses)

	// Method 3: Content similarity (simplified)
	contentSimilarity := p.calculateContentSimilarity(responses)

	// Weighted combination
	consensus := avgConfidence*0.4 + voteAgreement*0.4 + contentSimilarity*0.2

	return consensus
}

// calculateVoteAgreement calculates agreement among votes.
func (p *Protocol) calculateVoteAgreement(responses []*PhaseResponse) float64 {
	votes := make(map[string]int)
	totalVotes := 0

	for _, r := range responses {
		if r.Vote != "" {
			votes[r.Vote]++
			totalVotes++
		}
	}

	if totalVotes == 0 {
		return 0.5 // No votes, neutral
	}

	// Find majority
	maxVotes := 0
	for _, count := range votes {
		if count > maxVotes {
			maxVotes = count
		}
	}

	return float64(maxVotes) / float64(totalVotes)
}

// calculateContentSimilarity calculates how similar the responses are.
func (p *Protocol) calculateContentSimilarity(responses []*PhaseResponse) float64 {
	if len(responses) < 2 {
		return 1.0 // Single response is 100% similar to itself
	}

	// Simplified: count common key terms
	termCounts := make(map[string]int)
	for _, r := range responses {
		words := strings.Fields(strings.ToLower(r.Content))
		seen := make(map[string]bool)
		for _, word := range words {
			if len(word) > 3 && !seen[word] {
				termCounts[word]++
				seen[word] = true
			}
		}
	}

	// Calculate overlap
	commonTerms := 0
	for _, count := range termCounts {
		if count > 1 {
			commonTerms++
		}
	}

	if len(termCounts) == 0 {
		return 0
	}

	return float64(commonTerms) / float64(len(termCounts))
}

// extractInsights extracts key insights from responses.
func (p *Protocol) extractInsights(responses []*PhaseResponse) []string {
	insights := make([]string, 0)
	seen := make(map[string]bool)

	for _, r := range responses {
		// Extract arguments as insights
		for _, arg := range r.Arguments {
			if !seen[arg] {
				insights = append(insights, arg)
				seen[arg] = true
			}
		}

		// Extract suggestions as insights
		for _, sug := range r.Suggestions {
			if !seen[sug] {
				insights = append(insights, sug)
				seen[sug] = true
			}
		}
	}

	return insights
}

// findDisagreements identifies points of disagreement.
func (p *Protocol) findDisagreements(responses []*PhaseResponse) []string {
	disagreements := make([]string, 0)
	seen := make(map[string]bool)

	for _, r := range responses {
		for _, crit := range r.Criticisms {
			if !seen[crit] {
				disagreements = append(disagreements, crit)
				seen[crit] = true
			}
		}
	}

	return disagreements
}

// updatePhaseMetrics updates metrics for a phase.
func (p *Protocol) updatePhaseMetrics(phase topology.DebatePhase, result *PhaseResult) {
	p.mu.Lock()
	defer p.mu.Unlock()

	totalLatency := time.Duration(0)
	totalConfidence := 0.0

	for _, r := range result.Responses {
		totalLatency += r.Latency
		totalConfidence += r.Confidence
	}

	n := len(result.Responses)
	if n == 0 {
		n = 1
	}

	p.metrics.PhaseMetrics[phase] = &PhaseMetrics{
		ResponseCount:  len(result.Responses),
		AvgLatency:     totalLatency / time.Duration(n),
		AvgConfidence:  totalConfidence / float64(n),
		ConsensusLevel: result.ConsensusLevel,
	}

	p.metrics.TotalResponses += len(result.Responses)
}

// buildResult builds the final debate result.
func (p *Protocol) buildResult(startTime time.Time, success bool, earlyExitReason string) *DebateResult {
	p.mu.RLock()
	defer p.mu.RUnlock()

	endTime := time.Now()

	// Build final consensus
	consensus := p.buildFinalConsensus()

	// Find best response
	bestResponse := p.findBestResponse()

	// Calculate overall metrics
	p.calculateOverallMetrics()

	return &DebateResult{
		ID:               p.config.ID,
		Topic:            p.config.Topic,
		Phases:           p.phaseResults,
		FinalConsensus:   consensus,
		BestResponse:     bestResponse,
		ParticipantCount: len(p.topology.GetAgents()),
		TotalRounds:      p.config.MaxRounds,
		RoundsCompleted:  p.currentRound,
		Success:          success,
		EarlyExit:        earlyExitReason != "",
		EarlyExitReason:  earlyExitReason,
		StartTime:        startTime,
		EndTime:          endTime,
		Duration:         endTime.Sub(startTime),
		TopologyUsed:     p.config.TopologyType,
		Metrics:          p.metrics,
		Metadata:         p.config.Metadata,
	}
}

// buildFinalConsensus builds the consensus result from convergence phase.
func (p *Protocol) buildFinalConsensus() *ConsensusResult {
	// Find convergence phase results
	var convergencePhases []*PhaseResult
	for _, pr := range p.phaseResults {
		if pr.Phase == topology.PhaseConvergence {
			convergencePhases = append(convergencePhases, pr)
		}
	}

	if len(convergencePhases) == 0 {
		return nil
	}

	// Use last convergence phase
	lastConvergence := convergencePhases[len(convergencePhases)-1]

	// Count votes
	voteBreakdown := make(map[string]int)
	weightedVotes := make(map[string]float64)
	contributors := make([]string, 0)

	for _, r := range lastConvergence.Responses {
		if r.Vote != "" {
			voteBreakdown[r.Vote]++
			weightedVotes[r.Vote] += r.Confidence
		}
		contributors = append(contributors, r.AgentID)
	}

	// Find winning vote
	winningVote := ""
	maxWeight := 0.0
	for vote, weight := range weightedVotes {
		if weight > maxWeight {
			maxWeight = weight
			winningVote = vote
		}
	}

	// Determine consensus method
	method := ConsensusMethodWeightedVoting
	totalVotes := 0
	maxVoteCount := 0
	for _, count := range voteBreakdown {
		totalVotes += count
		if count > maxVoteCount {
			maxVoteCount = count
		}
	}

	if maxVoteCount == totalVotes {
		method = ConsensusMethodUnanimous
	} else if float64(maxVoteCount) > float64(totalVotes)*0.5 {
		method = ConsensusMethodMajority
	}

	// Build summary from best responses
	summary := p.buildConsensusSummary(lastConvergence)

	return &ConsensusResult{
		Summary:       summary,
		Confidence:    lastConvergence.ConsensusLevel,
		KeyPoints:     lastConvergence.KeyInsights,
		Dissents:      lastConvergence.Disagreements,
		VoteBreakdown: voteBreakdown,
		WinningVote:   winningVote,
		Contributors:  contributors,
		Method:        method,
	}
}

// buildConsensusSummary builds a summary from the convergence responses.
func (p *Protocol) buildConsensusSummary(convergence *PhaseResult) string {
	if convergence.LeaderResponse != nil {
		return convergence.LeaderResponse.Content
	}

	// Use highest confidence response
	var bestResp *PhaseResponse
	for _, r := range convergence.Responses {
		if bestResp == nil || r.Confidence > bestResp.Confidence {
			bestResp = r
		}
	}

	if bestResp != nil {
		return bestResp.Content
	}

	return "No consensus summary available"
}

// findBestResponse finds the best response across all phases.
func (p *Protocol) findBestResponse() *PhaseResponse {
	var best *PhaseResponse

	for _, pr := range p.phaseResults {
		for _, r := range pr.Responses {
			if best == nil || r.Score > best.Score || (r.Score == best.Score && r.Confidence > best.Confidence) {
				best = r
			}
		}
	}

	return best
}

// calculateOverallMetrics calculates overall debate metrics.
func (p *Protocol) calculateOverallMetrics() {
	totalLatency := time.Duration(0)
	totalConfidence := 0.0
	count := 0

	for _, pm := range p.metrics.PhaseMetrics {
		totalLatency += pm.AvgLatency * time.Duration(pm.ResponseCount)
		totalConfidence += pm.AvgConfidence * float64(pm.ResponseCount)
		count += pm.ResponseCount
	}

	if count > 0 {
		p.metrics.AvgLatency = totalLatency / time.Duration(count)
		p.metrics.AvgConfidence = totalConfidence / float64(count)
	}

	// Calculate consensus score from convergence phases
	for _, pr := range p.phaseResults {
		if pr.Phase == topology.PhaseConvergence {
			p.metrics.ConsensusScore = pr.ConsensusLevel
		}
	}
}

// Stop stops the protocol execution.
func (p *Protocol) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopped = true
}

// GetCurrentPhase returns the current phase.
func (p *Protocol) GetCurrentPhase() topology.DebatePhase {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentPhase
}

// GetCurrentRound returns the current round.
func (p *Protocol) GetCurrentRound() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentRound
}

// GetPhaseResults returns all phase results so far.
func (p *Protocol) GetPhaseResults() []*PhaseResult {
	p.mu.RLock()
	defer p.mu.RUnlock()

	results := make([]*PhaseResult, len(p.phaseResults))
	copy(results, p.phaseResults)
	return results
}
