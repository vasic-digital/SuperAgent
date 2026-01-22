// Package knowledge provides debate-lesson integration for learning during debates.
package knowledge

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"dev.helix.agent/internal/debate"
	"dev.helix.agent/internal/debate/agents"
	"dev.helix.agent/internal/debate/protocol"
	"dev.helix.agent/internal/debate/topology"
)

// DebateLearningIntegration provides integration between debates and the learning system.
type DebateLearningIntegration struct {
	repository Repository

	// Active debate tracking
	activeDebates map[string]*DebateLearningSession

	config IntegrationConfig
	mu     sync.RWMutex
}

// IntegrationConfig configures the learning integration.
type IntegrationConfig struct {
	// Enable automatic lesson extraction after debates
	AutoExtractLessons bool `json:"auto_extract_lessons"`
	// Enable applying relevant lessons before debates
	AutoApplyLessons bool `json:"auto_apply_lessons"`
	// Minimum consensus for lesson extraction
	MinConsensusForLesson float64 `json:"min_consensus_for_lesson"`
	// Maximum lessons to apply per debate
	MaxLessonsPerDebate int `json:"max_lessons_per_debate"`
	// Enable cognitive planning integration
	EnableCognitiveIntegration bool `json:"enable_cognitive_integration"`
	// Enable pattern detection
	EnablePatternDetection bool `json:"enable_pattern_detection"`
	// Pattern detection threshold
	PatternThreshold float64 `json:"pattern_threshold"`
}

// DefaultIntegrationConfig returns sensible defaults.
func DefaultIntegrationConfig() IntegrationConfig {
	return IntegrationConfig{
		AutoExtractLessons:         true,
		AutoApplyLessons:           true,
		MinConsensusForLesson:      0.7,
		MaxLessonsPerDebate:        5,
		EnableCognitiveIntegration: true,
		EnablePatternDetection:     true,
		PatternThreshold:           0.65,
	}
}

// DebateLearningSession tracks learning within a single debate.
type DebateLearningSession struct {
	DebateID         string                     `json:"debate_id"`
	Topic            string                     `json:"topic"`
	Domain           agents.Domain              `json:"domain"`
	StartTime        time.Time                  `json:"start_time"`
	AppliedLessons   []*LessonApplication       `json:"applied_lessons"`
	PhaseLearning    []*PhaseLearning           `json:"phase_learning"`
	DetectedPatterns []*DebatePattern           `json:"detected_patterns"`
	AgentKnowledge   map[string]*AgentKnowledge `json:"agent_knowledge"`
	CognitiveState   *CognitiveState            `json:"cognitive_state,omitempty"`
}

// PhaseLearning tracks learning during a specific phase.
type PhaseLearning struct {
	Phase           topology.DebatePhase `json:"phase"`
	InsightsGained  []string             `json:"insights_gained"`
	PatternsMatched []string             `json:"patterns_matched"`
	LessonsApplied  []string             `json:"lessons_applied"`
	QualityDelta    float64              `json:"quality_delta"`
	Timestamp       time.Time            `json:"timestamp"`
}

// CognitiveState tracks cognitive planning state during debate.
type CognitiveState struct {
	RefinementCount int        `json:"refinement_count"`
	LastRefinement  *time.Time `json:"last_refinement,omitempty"`
	ImprovementRate float64    `json:"improvement_rate"`
}

// NewDebateLearningIntegration creates a new integration.
func NewDebateLearningIntegration(repository Repository, config IntegrationConfig) *DebateLearningIntegration {
	return &DebateLearningIntegration{
		repository:    repository,
		activeDebates: make(map[string]*DebateLearningSession),
		config:        config,
	}
}

// StartDebateLearning initializes learning for a new debate.
func (dli *DebateLearningIntegration) StartDebateLearning(ctx context.Context, debateID, topic string, participants []*agents.SpecializedAgent) (*DebateLearningSession, error) {
	// Infer domain from topic
	domain := dli.inferDomain(topic)

	session := &DebateLearningSession{
		DebateID:         debateID,
		Topic:            topic,
		Domain:           domain,
		StartTime:        time.Now(),
		AppliedLessons:   make([]*LessonApplication, 0),
		PhaseLearning:    make([]*PhaseLearning, 0),
		DetectedPatterns: make([]*DebatePattern, 0),
		AgentKnowledge:   make(map[string]*AgentKnowledge),
	}

	// Auto-apply lessons if enabled
	if dli.config.AutoApplyLessons {
		lessons, err := dli.repository.GetRelevantLessons(ctx, topic, domain)
		if err == nil && len(lessons) > 0 {
			for i, match := range lessons {
				if i >= dli.config.MaxLessonsPerDebate {
					break
				}

				application, err := dli.repository.ApplyLesson(ctx, match.Lesson.ID, debateID)
				if err == nil {
					session.AppliedLessons = append(session.AppliedLessons, application)
				}
			}
		}
	}

	// Prepare agent knowledge
	for _, agent := range participants {
		knowledge, err := dli.repository.GetKnowledgeForAgent(ctx, agent, topic)
		if err == nil {
			session.AgentKnowledge[agent.ID] = knowledge
		}
	}

	// Initialize cognitive state if enabled
	if dli.config.EnableCognitiveIntegration {
		session.CognitiveState = &CognitiveState{}
	}

	dli.mu.Lock()
	dli.activeDebates[debateID] = session
	dli.mu.Unlock()

	return session, nil
}

// OnPhaseComplete handles learning when a debate phase completes.
func (dli *DebateLearningIntegration) OnPhaseComplete(ctx context.Context, debateID string, phaseResult *protocol.PhaseResult) error {
	dli.mu.Lock()
	session, ok := dli.activeDebates[debateID]
	if !ok {
		dli.mu.Unlock()
		return fmt.Errorf("no active learning session for debate: %s", debateID)
	}
	dli.mu.Unlock()

	// Create phase learning record
	phaseLearning := &PhaseLearning{
		Phase:           phaseResult.Phase,
		InsightsGained:  phaseResult.KeyInsights,
		PatternsMatched: make([]string, 0),
		LessonsApplied:  make([]string, 0),
		Timestamp:       time.Now(),
	}

	// Check for pattern matches
	if dli.config.EnablePatternDetection {
		patterns := dli.detectPatterns(ctx, session, phaseResult)
		for _, pattern := range patterns {
			phaseLearning.PatternsMatched = append(phaseLearning.PatternsMatched, pattern.Name)
			session.DetectedPatterns = append(session.DetectedPatterns, pattern)
		}
	}

	// Update cognitive state if enabled
	if dli.config.EnableCognitiveIntegration && session.CognitiveState != nil {
		dli.updateCognitiveState(session, phaseResult)
	}

	// Calculate quality delta from previous phase
	if len(session.PhaseLearning) > 0 {
		prevPhase := session.PhaseLearning[len(session.PhaseLearning)-1]
		phaseLearning.QualityDelta = phaseResult.ConsensusLevel - prevPhase.QualityDelta
	}

	session.PhaseLearning = append(session.PhaseLearning, phaseLearning)

	return nil
}

// OnDebateComplete handles learning when a debate completes.
func (dli *DebateLearningIntegration) OnDebateComplete(ctx context.Context, result *protocol.DebateResult) (*DebateLearningResult, error) {
	dli.mu.Lock()
	session, ok := dli.activeDebates[result.ID]
	if !ok {
		dli.mu.Unlock()
		// Create a minimal session for lesson extraction
		session = &DebateLearningSession{
			DebateID:  result.ID,
			Topic:     result.Topic,
			StartTime: result.StartTime,
		}
	} else {
		delete(dli.activeDebates, result.ID)
		dli.mu.Unlock()
	}

	learningResult := &DebateLearningResult{
		DebateID:         result.ID,
		SessionDuration:  time.Since(session.StartTime),
		AppliedLessons:   len(session.AppliedLessons),
		DetectedPatterns: len(session.DetectedPatterns),
	}

	// Extract lessons if enabled and consensus is high enough
	if dli.config.AutoExtractLessons {
		if result.FinalConsensus != nil && result.FinalConsensus.Confidence >= dli.config.MinConsensusForLesson {
			lessons, err := dli.repository.ExtractLessons(ctx, result)
			if err == nil {
				learningResult.ExtractedLessons = len(lessons)
				learningResult.Lessons = lessons
			}
		}
	}

	// Record detected patterns
	for _, pattern := range session.DetectedPatterns {
		if pattern.Confidence >= dli.config.PatternThreshold {
			_ = dli.repository.RecordPattern(ctx, pattern)
		}
	}

	// Record outcomes for applied lessons
	for _, application := range session.AppliedLessons {
		success := result.Success && (result.FinalConsensus == nil || result.FinalConsensus.Confidence >= 0.7)
		feedback := fmt.Sprintf("Debate %s with consensus %.2f",
			map[bool]string{true: "succeeded", false: "failed"}[result.Success],
			result.FinalConsensus.Confidence)
		_ = dli.repository.RecordOutcome(ctx, application, success, feedback)
	}

	// Calculate cognitive improvement
	if session.CognitiveState != nil {
		learningResult.CognitiveRefinements = session.CognitiveState.RefinementCount
		learningResult.ImprovementRate = session.CognitiveState.ImprovementRate
	}

	return learningResult, nil
}

// DebateLearningResult summarizes learning from a debate.
type DebateLearningResult struct {
	DebateID             string           `json:"debate_id"`
	SessionDuration      time.Duration    `json:"session_duration"`
	AppliedLessons       int              `json:"applied_lessons"`
	ExtractedLessons     int              `json:"extracted_lessons"`
	DetectedPatterns     int              `json:"detected_patterns"`
	CognitiveRefinements int              `json:"cognitive_refinements"`
	ImprovementRate      float64          `json:"improvement_rate"`
	Lessons              []*debate.Lesson `json:"lessons,omitempty"`
}

// GetAgentKnowledge retrieves knowledge for an agent in an active debate.
func (dli *DebateLearningIntegration) GetAgentKnowledge(debateID, agentID string) (*AgentKnowledge, error) {
	dli.mu.RLock()
	defer dli.mu.RUnlock()

	session, ok := dli.activeDebates[debateID]
	if !ok {
		return nil, fmt.Errorf("no active session for debate: %s", debateID)
	}

	knowledge, ok := session.AgentKnowledge[agentID]
	if !ok {
		return nil, fmt.Errorf("no knowledge for agent: %s", agentID)
	}

	return knowledge, nil
}

// GetLessonsForPrompt returns formatted lessons for inclusion in agent prompts.
func (dli *DebateLearningIntegration) GetLessonsForPrompt(debateID string, agentID string) (string, error) {
	knowledge, err := dli.GetAgentKnowledge(debateID, agentID)
	if err != nil {
		return "", err
	}

	if len(knowledge.RelevantLessons) == 0 {
		return "", nil
	}

	var result string
	result += "## Relevant Knowledge from Previous Debates\n\n"

	for i, match := range knowledge.RelevantLessons {
		if i >= 3 { // Limit to top 3 lessons
			break
		}
		lesson := match.Lesson
		result += fmt.Sprintf("### %s (relevance: %.0f%%)\n", lesson.Title, match.Score*100)
		result += fmt.Sprintf("**Problem:** %s\n", truncate(lesson.Content.Problem, 150))
		result += fmt.Sprintf("**Solution:** %s\n", truncate(lesson.Content.Solution, 200))
		if len(lesson.Content.TradeOffs) > 0 {
			result += fmt.Sprintf("**Key Trade-off:** %s\n", lesson.Content.TradeOffs[0].Aspect)
		}
		result += "\n"
	}

	// Add domain insights
	if len(knowledge.DomainInsights) > 0 {
		result += "## Domain Insights\n"
		for _, insight := range knowledge.DomainInsights[:min(3, len(knowledge.DomainInsights))] {
			result += fmt.Sprintf("- %s\n", insight)
		}
		result += "\n"
	}

	// Add role guidance
	if len(knowledge.RoleGuidance) > 0 {
		result += "## Role Guidance\n"
		for _, guidance := range knowledge.RoleGuidance[:min(3, len(knowledge.RoleGuidance))] {
			result += fmt.Sprintf("- %s\n", guidance)
		}
	}

	return result, nil
}

// detectPatterns detects patterns in the current phase.
func (dli *DebateLearningIntegration) detectPatterns(ctx context.Context, session *DebateLearningSession, phaseResult *protocol.PhaseResult) []*DebatePattern {
	patterns := make([]*DebatePattern, 0)

	// Detect consensus building pattern
	if phaseResult.ConsensusLevel >= 0.8 {
		patterns = append(patterns, &DebatePattern{
			ID:          uuid.New().String(),
			Name:        "High Consensus",
			Description: fmt.Sprintf("Achieved %.0f%% consensus in %s phase", phaseResult.ConsensusLevel*100, phaseResult.Phase),
			PatternType: PatternTypeConsensusBuilding,
			Domain:      session.Domain,
			Frequency:   1,
			SuccessRate: 1.0,
			Confidence:  phaseResult.ConsensusLevel,
			Triggers:    []string{string(phaseResult.Phase)},
		})
	}

	// Detect conflict resolution pattern
	if len(phaseResult.Disagreements) > 0 && phaseResult.ConsensusLevel >= 0.6 {
		patterns = append(patterns, &DebatePattern{
			ID:          uuid.New().String(),
			Name:        "Conflict Resolution",
			Description: fmt.Sprintf("Resolved %d disagreements", len(phaseResult.Disagreements)),
			PatternType: PatternTypeConflictResolution,
			Domain:      session.Domain,
			Frequency:   1,
			SuccessRate: phaseResult.ConsensusLevel,
			Confidence:  0.7,
		})
	}

	// Detect expertise pattern from high-confidence responses
	highConfidenceCount := 0
	for _, resp := range phaseResult.Responses {
		if resp.Confidence >= 0.85 {
			highConfidenceCount++
		}
	}
	if highConfidenceCount >= 2 {
		patterns = append(patterns, &DebatePattern{
			ID:          uuid.New().String(),
			Name:        "Expert Contributions",
			Description: fmt.Sprintf("%d high-confidence responses", highConfidenceCount),
			PatternType: PatternTypeExpertise,
			Domain:      session.Domain,
			Frequency:   1,
			SuccessRate: 0.9,
			Confidence:  0.8,
		})
	}

	return patterns
}

// updateCognitiveState updates the cognitive state based on phase results.
func (dli *DebateLearningIntegration) updateCognitiveState(session *DebateLearningSession, phaseResult *protocol.PhaseResult) {
	if session.CognitiveState == nil {
		return
	}

	// Track refinement
	session.CognitiveState.RefinementCount++
	now := time.Now()
	session.CognitiveState.LastRefinement = &now

	// Calculate improvement rate based on consensus progression
	if len(session.PhaseLearning) > 0 {
		prevConsensus := 0.0
		for _, pl := range session.PhaseLearning {
			prevConsensus = pl.QualityDelta
		}
		improvement := phaseResult.ConsensusLevel - prevConsensus

		// Exponential moving average
		alpha := 0.3
		session.CognitiveState.ImprovementRate = alpha*improvement + (1-alpha)*session.CognitiveState.ImprovementRate
	}
}

// inferDomain infers domain from topic.
func (dli *DebateLearningIntegration) inferDomain(topic string) agents.Domain {
	topic = toLower(topic)

	if containsAny(topic, "security", "vulnerability", "auth", "encrypt", "hack") {
		return agents.DomainSecurity
	}
	if containsAny(topic, "architecture", "design", "system", "scale", "microservice") {
		return agents.DomainArchitecture
	}
	if containsAny(topic, "performance", "optimize", "speed", "cache", "latency") {
		return agents.DomainOptimization
	}
	if containsAny(topic, "debug", "error", "bug", "fix", "trace") {
		return agents.DomainDebug
	}
	if containsAny(topic, "code", "implement", "function", "class", "refactor") {
		return agents.DomainCode
	}
	if containsAny(topic, "logic", "reason", "proof", "algorithm", "math") {
		return agents.DomainReasoning
	}

	return agents.DomainGeneral
}

// GetActiveSession retrieves an active learning session.
func (dli *DebateLearningIntegration) GetActiveSession(debateID string) (*DebateLearningSession, bool) {
	dli.mu.RLock()
	defer dli.mu.RUnlock()
	session, ok := dli.activeDebates[debateID]
	return session, ok
}

// GetActiveSessions returns all active learning sessions.
func (dli *DebateLearningIntegration) GetActiveSessions() []*DebateLearningSession {
	dli.mu.RLock()
	defer dli.mu.RUnlock()

	sessions := make([]*DebateLearningSession, 0, len(dli.activeDebates))
	for _, session := range dli.activeDebates {
		sessions = append(sessions, session)
	}
	return sessions
}

// =============================================================================
// Learning-Enhanced Protocol Adapter
// =============================================================================

// LearningEnhancedProtocol wraps a debate protocol with learning capabilities.
type LearningEnhancedProtocol struct {
	protocol    *protocol.Protocol
	integration *DebateLearningIntegration
	debateID    string
	agents      []*agents.SpecializedAgent
}

// NewLearningEnhancedProtocol creates a learning-enhanced protocol wrapper.
func NewLearningEnhancedProtocol(
	proto *protocol.Protocol,
	integration *DebateLearningIntegration,
	debateAgents []*agents.SpecializedAgent,
) *LearningEnhancedProtocol {
	return &LearningEnhancedProtocol{
		protocol:    proto,
		integration: integration,
		agents:      debateAgents,
	}
}

// Execute runs the debate with learning integration.
// Note: Call ExecuteWithContext for full learning integration.
func (lep *LearningEnhancedProtocol) Execute(ctx context.Context) (*protocol.DebateResult, *DebateLearningResult, error) {
	// Execute the debate first to get result with ID and Topic
	result, err := lep.protocol.Execute(ctx)
	if err != nil {
		return nil, nil, err
	}

	lep.debateID = result.ID

	// Complete learning using result data
	learningResult, _ := lep.integration.OnDebateComplete(ctx, result)

	return result, learningResult, nil
}

// ExecuteWithLearning runs the debate with full learning integration.
func (lep *LearningEnhancedProtocol) ExecuteWithLearning(ctx context.Context, debateID, topic string) (*protocol.DebateResult, *DebateLearningResult, error) {
	lep.debateID = debateID

	// Start learning session
	session, err := lep.integration.StartDebateLearning(ctx, debateID, topic, lep.agents)
	if err != nil {
		// Non-fatal - continue without learning
		session = nil
	}

	// Execute the debate
	result, err := lep.protocol.Execute(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Complete learning
	var learningResult *DebateLearningResult
	if session != nil {
		learningResult, _ = lep.integration.OnDebateComplete(ctx, result)
	}

	return result, learningResult, nil
}

// EnhanceAgentPrompt adds learning context to an agent's prompt.
func (lep *LearningEnhancedProtocol) EnhanceAgentPrompt(agentID, basePrompt string) string {
	lessonsPrompt, err := lep.integration.GetLessonsForPrompt(lep.debateID, agentID)
	if err != nil || lessonsPrompt == "" {
		return basePrompt
	}

	return fmt.Sprintf("%s\n\n---\n\n%s", basePrompt, lessonsPrompt)
}
