// Package knowledge provides the knowledge repository interface for the debate system.
// It bridges the debate protocol with the lesson banking system and provides
// cross-debate learning capabilities.
package knowledge

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"dev.helix.agent/internal/debate"
	"dev.helix.agent/internal/debate/agents"
	"dev.helix.agent/internal/debate/protocol"
	"dev.helix.agent/internal/debate/topology"
)

// Repository provides a unified interface for knowledge management in debates.
type Repository interface {
	// Lesson Management
	ExtractLessons(ctx context.Context, result *protocol.DebateResult) ([]*debate.Lesson, error)
	SearchLessons(ctx context.Context, query string, options SearchOptions) ([]*LessonMatch, error)
	GetRelevantLessons(ctx context.Context, topic string, domain agents.Domain) ([]*LessonMatch, error)
	ApplyLesson(ctx context.Context, lessonID string, debateID string) (*LessonApplication, error)
	RecordOutcome(ctx context.Context, application *LessonApplication, success bool, feedback string) error

	// Cross-Debate Learning
	GetPatterns(ctx context.Context, filter PatternFilter) ([]*DebatePattern, error)
	RecordPattern(ctx context.Context, pattern *DebatePattern) error
	GetSuccessfulStrategies(ctx context.Context, domain agents.Domain) ([]*Strategy, error)

	// Knowledge Retrieval
	GetKnowledgeForAgent(ctx context.Context, agent *agents.SpecializedAgent, topic string) (*AgentKnowledge, error)
	GetDebateHistory(ctx context.Context, filter HistoryFilter) ([]*DebateHistoryEntry, error)

	// Statistics
	GetStatistics(ctx context.Context) (*RepositoryStatistics, error)
}

// SearchOptions configures lesson search.
type SearchOptions struct {
	Categories     []debate.LessonCategory `json:"categories,omitempty"`
	Domain         agents.Domain           `json:"domain,omitempty"`
	MinTier        *debate.LessonTier      `json:"min_tier,omitempty"`
	MinScore       float64                 `json:"min_score"`
	Limit          int                     `json:"limit"`
	IncludeExpired bool                    `json:"include_expired"`
}

// DefaultSearchOptions returns sensible default search options.
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		MinScore: 0.5,
		Limit:    10,
	}
}

// LessonMatch represents a matched lesson with relevance score.
type LessonMatch struct {
	Lesson    *debate.Lesson `json:"lesson"`
	Score     float64        `json:"score"`
	MatchType string         `json:"match_type"` // "semantic", "keyword", "domain"
	Relevance *Relevance     `json:"relevance"`
}

// Relevance explains why a lesson matches.
type Relevance struct {
	TopicMatch   float64  `json:"topic_match"`
	DomainMatch  float64  `json:"domain_match"`
	PatternMatch float64  `json:"pattern_match"`
	Reasons      []string `json:"reasons"`
}

// LessonApplication tracks the application of a lesson to a debate.
type LessonApplication struct {
	ID        string              `json:"id"`
	LessonID  string              `json:"lesson_id"`
	DebateID  string              `json:"debate_id"`
	AppliedAt time.Time           `json:"applied_at"`
	AppliedBy string              `json:"applied_by"` // Agent ID
	Context   string              `json:"context"`
	Outcome   *ApplicationOutcome `json:"outcome,omitempty"`
}

// ApplicationOutcome records the result of applying a lesson.
type ApplicationOutcome struct {
	Success    bool      `json:"success"`
	Feedback   string    `json:"feedback"`
	Impact     float64   `json:"impact"` // -1.0 to 1.0
	RecordedAt time.Time `json:"recorded_at"`
}

// DebatePattern represents a recurring pattern across debates.
type DebatePattern struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	PatternType   PatternType            `json:"pattern_type"`
	Domain        agents.Domain          `json:"domain"`
	Frequency     int                    `json:"frequency"`
	SuccessRate   float64                `json:"success_rate"`
	Triggers      []string               `json:"triggers"`
	Indicators    []PatternIndicator     `json:"indicators"`
	Responses     []PatternResponse      `json:"responses"`
	FirstObserved time.Time              `json:"first_observed"`
	LastObserved  time.Time              `json:"last_observed"`
	Confidence    float64                `json:"confidence"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// PatternType categorizes debate patterns.
type PatternType string

const (
	PatternTypeConsensusBuilding  PatternType = "consensus_building"
	PatternTypeConflictResolution PatternType = "conflict_resolution"
	PatternTypeKnowledgeGap       PatternType = "knowledge_gap"
	PatternTypeExpertise          PatternType = "expertise"
	PatternTypeOptimization       PatternType = "optimization"
	PatternTypeFailure            PatternType = "failure"
)

// PatternIndicator describes what indicates a pattern.
type PatternIndicator struct {
	Type      string  `json:"type"`
	Threshold float64 `json:"threshold"`
	Weight    float64 `json:"weight"`
}

// PatternResponse describes how to respond to a pattern.
type PatternResponse struct {
	Action     string            `json:"action"`
	Priority   int               `json:"priority"`
	Conditions []string          `json:"conditions,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// PatternFilter filters patterns by criteria.
type PatternFilter struct {
	Types        []PatternType `json:"types,omitempty"`
	Domain       agents.Domain `json:"domain,omitempty"`
	MinFrequency int           `json:"min_frequency"`
	MinSuccess   float64       `json:"min_success"`
	Since        *time.Time    `json:"since,omitempty"`
}

// Strategy represents a successful debate strategy.
type Strategy struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	Domain       agents.Domain         `json:"domain"`
	TopologyType topology.TopologyType `json:"topology_type"`
	RoleConfig   []RoleConfiguration   `json:"role_config"`
	Phases       []PhaseStrategy       `json:"phases"`
	SuccessRate  float64               `json:"success_rate"`
	Applications int                   `json:"applications"`
	AvgConsensus float64               `json:"avg_consensus"`
	AvgDuration  time.Duration         `json:"avg_duration"`
}

// RoleConfiguration describes role assignment in a strategy.
type RoleConfiguration struct {
	Role            topology.AgentRole `json:"role"`
	PreferredDomain agents.Domain      `json:"preferred_domain"`
	MinScore        float64            `json:"min_score"`
	Count           int                `json:"count"`
}

// PhaseStrategy describes phase-specific strategy.
type PhaseStrategy struct {
	Phase            topology.DebatePhase `json:"phase"`
	FocusAreas       []string             `json:"focus_areas"`
	PromptHints      []string             `json:"prompt_hints"`
	MinConfidence    float64              `json:"min_confidence"`
	ExpectedInsights int                  `json:"expected_insights"`
}

// AgentKnowledge represents curated knowledge for an agent.
type AgentKnowledge struct {
	AgentID            string           `json:"agent_id"`
	RelevantLessons    []*LessonMatch   `json:"relevant_lessons"`
	ApplicablePatterns []*DebatePattern `json:"applicable_patterns"`
	DomainInsights     []string         `json:"domain_insights"`
	RoleGuidance       []string         `json:"role_guidance"`
	HistoricalContext  []string         `json:"historical_context"`
	GeneratedAt        time.Time        `json:"generated_at"`
}

// DebateHistoryEntry represents a historical debate record.
type DebateHistoryEntry struct {
	ID             string                 `json:"id"`
	Topic          string                 `json:"topic"`
	Domain         agents.Domain          `json:"domain"`
	Success        bool                   `json:"success"`
	ConsensusScore float64                `json:"consensus_score"`
	Duration       time.Duration          `json:"duration"`
	Participants   int                    `json:"participants"`
	LessonsLearned int                    `json:"lessons_learned"`
	Timestamp      time.Time              `json:"timestamp"`
	Summary        string                 `json:"summary"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// HistoryFilter filters debate history.
type HistoryFilter struct {
	Domain       agents.Domain `json:"domain,omitempty"`
	SuccessOnly  bool          `json:"success_only"`
	MinConsensus float64       `json:"min_consensus"`
	Since        *time.Time    `json:"since,omitempty"`
	Limit        int           `json:"limit"`
}

// RepositoryStatistics provides statistics about the knowledge repository.
type RepositoryStatistics struct {
	TotalLessons        int                     `json:"total_lessons"`
	TotalPatterns       int                     `json:"total_patterns"`
	TotalStrategies     int                     `json:"total_strategies"`
	TotalDebates        int                     `json:"total_debates"`
	LessonsByDomain     map[agents.Domain]int   `json:"lessons_by_domain"`
	PatternsByType      map[PatternType]int     `json:"patterns_by_type"`
	OverallSuccessRate  float64                 `json:"overall_success_rate"`
	AvgLessonsPerDebate float64                 `json:"avg_lessons_per_debate"`
	TopCategories       []debate.LessonCategory `json:"top_categories"`
	LastUpdated         time.Time               `json:"last_updated"`
}

// =============================================================================
// Default Implementation
// =============================================================================

// DefaultRepository provides the default implementation of the Repository interface.
type DefaultRepository struct {
	lessonBank    *debate.LessonBank
	patterns      map[string]*DebatePattern
	strategies    map[string]*Strategy
	history       []*DebateHistoryEntry
	applications  map[string]*LessonApplication
	domainLessons map[agents.Domain][]*debate.Lesson

	config RepositoryConfig
	mu     sync.RWMutex
}

// RepositoryConfig configures the default repository.
type RepositoryConfig struct {
	MaxPatterns       int           `json:"max_patterns"`
	MaxStrategies     int           `json:"max_strategies"`
	MaxHistoryEntries int           `json:"max_history_entries"`
	PatternThreshold  float64       `json:"pattern_threshold"`
	StrategyThreshold float64       `json:"strategy_threshold"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
}

// DefaultRepositoryConfig returns sensible defaults.
func DefaultRepositoryConfig() RepositoryConfig {
	return RepositoryConfig{
		MaxPatterns:       1000,
		MaxStrategies:     500,
		MaxHistoryEntries: 10000,
		PatternThreshold:  0.7,
		StrategyThreshold: 0.75,
		CleanupInterval:   24 * time.Hour,
	}
}

// NewDefaultRepository creates a new default repository.
func NewDefaultRepository(lessonBank *debate.LessonBank, config RepositoryConfig) *DefaultRepository {
	return &DefaultRepository{
		lessonBank:    lessonBank,
		patterns:      make(map[string]*DebatePattern),
		strategies:    make(map[string]*Strategy),
		history:       make([]*DebateHistoryEntry, 0),
		applications:  make(map[string]*LessonApplication),
		domainLessons: make(map[agents.Domain][]*debate.Lesson),
		config:        config,
	}
}

// ExtractLessons extracts lessons from a debate result.
func (r *DefaultRepository) ExtractLessons(ctx context.Context, result *protocol.DebateResult) ([]*debate.Lesson, error) {
	if result == nil {
		return nil, fmt.Errorf("debate result is nil")
	}

	// Convert protocol.DebateResult to debate.DebateResult for lesson extraction
	debateResult := r.convertDebateResult(result)

	// Use lesson bank to extract lessons
	lessons, err := r.lessonBank.ExtractLessonsFromDebate(ctx, debateResult)
	if err != nil {
		return nil, fmt.Errorf("failed to extract lessons: %w", err)
	}

	// Index lessons by domain
	r.mu.Lock()
	for _, lesson := range lessons {
		domain := r.inferDomain(lesson)
		r.domainLessons[domain] = append(r.domainLessons[domain], lesson)
	}
	r.mu.Unlock()

	// Record history entry
	r.recordHistory(result, len(lessons))

	return lessons, nil
}

// convertDebateResult converts protocol.DebateResult to debate.DebateResult.
func (r *DefaultRepository) convertDebateResult(result *protocol.DebateResult) *debate.DebateResult {
	// Build rounds from phases
	rounds := make([]debate.DebateRound, 0, len(result.Phases))

	for i, phase := range result.Phases {
		responses := make([]debate.DebateResponse, 0, len(phase.Responses))

		for _, resp := range phase.Responses {
			responses = append(responses, debate.DebateResponse{
				Participant: resp.Provider + "/" + resp.Model,
				Content:     resp.Content,
				Confidence:  resp.Confidence,
				Arguments:   resp.Arguments,
			})
		}

		rounds = append(rounds, debate.DebateRound{
			Number:      i + 1,
			Responses:   responses,
			Summary:     r.buildPhaseSummary(phase),
			KeyInsights: phase.KeyInsights,
		})
	}

	// Build participants list
	participants := make([]string, 0)
	seen := make(map[string]bool)
	for _, phase := range result.Phases {
		for _, resp := range phase.Responses {
			participant := resp.Provider + "/" + resp.Model
			if !seen[participant] {
				participants = append(participants, participant)
				seen[participant] = true
			}
		}
	}

	// Build consensus
	var consensus *debate.DebateConsensus
	if result.FinalConsensus != nil {
		consensus = &debate.DebateConsensus{
			Summary:    result.FinalConsensus.Summary,
			Confidence: result.FinalConsensus.Confidence,
			KeyPoints:  result.FinalConsensus.KeyPoints,
			Dissents:   result.FinalConsensus.Dissents,
		}
	}

	return &debate.DebateResult{
		ID:           result.ID,
		Topic:        result.Topic,
		Rounds:       rounds,
		Consensus:    consensus,
		Participants: participants,
		StartedAt:    result.StartTime,
		EndedAt:      result.EndTime,
	}
}

// buildPhaseSummary creates a summary from a phase result.
func (r *DefaultRepository) buildPhaseSummary(phase *protocol.PhaseResult) string {
	if phase.LeaderResponse != nil {
		return phase.LeaderResponse.Content[:min(500, len(phase.LeaderResponse.Content))]
	}

	if len(phase.Responses) > 0 {
		return fmt.Sprintf("Phase %s completed with %d responses. Consensus level: %.2f",
			phase.Phase, len(phase.Responses), phase.ConsensusLevel)
	}

	return "No responses in phase"
}

// inferDomain infers the domain from a lesson's content.
func (r *DefaultRepository) inferDomain(lesson *debate.Lesson) agents.Domain {
	// Map lesson category to agent domain
	switch lesson.Category {
	case debate.LessonCategorySecurity:
		return agents.DomainSecurity
	case debate.LessonCategoryArchitecture:
		return agents.DomainArchitecture
	case debate.LessonCategoryOptimization:
		return agents.DomainOptimization
	case debate.LessonCategoryDebugging:
		return agents.DomainDebug
	case debate.LessonCategoryPattern, debate.LessonCategoryRefactoring:
		return agents.DomainCode
	default:
		return agents.DomainGeneral
	}
}

// SearchLessons searches for lessons matching the query.
func (r *DefaultRepository) SearchLessons(ctx context.Context, query string, options SearchOptions) ([]*LessonMatch, error) {
	// Use lesson bank search
	searchOpts := debate.SearchOptions{
		Categories: options.Categories,
		MinScore:   options.MinScore,
		Limit:      options.Limit,
	}
	if options.MinTier != nil {
		searchOpts.MinTier = options.MinTier
	}

	results, err := r.lessonBank.SearchLessons(ctx, query, searchOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to search lessons: %w", err)
	}

	// Convert to LessonMatch
	matches := make([]*LessonMatch, 0, len(results))
	for _, result := range results {
		matches = append(matches, &LessonMatch{
			Lesson:    result.Lesson,
			Score:     result.Score,
			MatchType: result.MatchType,
			Relevance: &Relevance{
				TopicMatch: result.Score,
				Reasons:    result.Highlights,
			},
		})
	}

	return matches, nil
}

// GetRelevantLessons finds lessons relevant to a topic and domain.
func (r *DefaultRepository) GetRelevantLessons(ctx context.Context, topic string, domain agents.Domain) ([]*LessonMatch, error) {
	r.mu.RLock()
	domainLessons := r.domainLessons[domain]
	r.mu.RUnlock()

	// Get category for domain
	category := r.domainToCategory(domain)

	// Search with domain filter
	options := SearchOptions{
		Domain:   domain,
		MinScore: 0.3,
		Limit:    20,
	}
	if category != "" {
		options.Categories = []debate.LessonCategory{category}
	}

	matches, err := r.SearchLessons(ctx, topic, options)
	if err != nil {
		return nil, err
	}

	// Add domain-indexed lessons
	for _, lesson := range domainLessons {
		// Check if already in matches
		found := false
		for _, m := range matches {
			if m.Lesson.ID == lesson.ID {
				found = true
				break
			}
		}
		if !found {
			matches = append(matches, &LessonMatch{
				Lesson:    lesson,
				Score:     0.5, // Base relevance for domain match
				MatchType: "domain",
				Relevance: &Relevance{
					DomainMatch: 1.0,
					Reasons:     []string{fmt.Sprintf("Domain match: %s", domain)},
				},
			})
		}
	}

	// Sort by score
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Apply limit
	if len(matches) > options.Limit {
		matches = matches[:options.Limit]
	}

	return matches, nil
}

// domainToCategory maps agent domain to lesson category.
func (r *DefaultRepository) domainToCategory(domain agents.Domain) debate.LessonCategory {
	switch domain {
	case agents.DomainSecurity:
		return debate.LessonCategorySecurity
	case agents.DomainArchitecture:
		return debate.LessonCategoryArchitecture
	case agents.DomainOptimization:
		return debate.LessonCategoryOptimization
	case agents.DomainDebug:
		return debate.LessonCategoryDebugging
	case agents.DomainCode:
		return debate.LessonCategoryPattern
	default:
		return ""
	}
}

// ApplyLesson records the application of a lesson to a debate.
func (r *DefaultRepository) ApplyLesson(ctx context.Context, lessonID string, debateID string) (*LessonApplication, error) {
	// Verify lesson exists
	lesson, err := r.lessonBank.GetLesson(ctx, lessonID)
	if err != nil {
		return nil, fmt.Errorf("lesson not found: %w", err)
	}

	application := &LessonApplication{
		ID:        uuid.New().String(),
		LessonID:  lessonID,
		DebateID:  debateID,
		AppliedAt: time.Now(),
		Context:   lesson.Content.Problem,
	}

	r.mu.Lock()
	r.applications[application.ID] = application
	r.mu.Unlock()

	// Record in lesson bank (uses context string, not debate ID)
	if _, err := r.lessonBank.ApplyLesson(ctx, lessonID, debateID); err != nil {
		return nil, fmt.Errorf("failed to record application: %w", err)
	}

	return application, nil
}

// RecordOutcome records the outcome of applying a lesson.
func (r *DefaultRepository) RecordOutcome(ctx context.Context, application *LessonApplication, success bool, feedback string) error {
	r.mu.Lock()
	app, ok := r.applications[application.ID]
	if !ok {
		r.mu.Unlock()
		return fmt.Errorf("application not found: %s", application.ID)
	}

	app.Outcome = &ApplicationOutcome{
		Success:    success,
		Feedback:   feedback,
		RecordedAt: time.Now(),
	}
	r.mu.Unlock()

	// Record in lesson bank
	outcome := &debate.ApplicationOutcome{
		Success:     success,
		Feedback:    feedback,
		CompletedAt: time.Now(),
	}
	return r.lessonBank.RecordOutcome(ctx, application.LessonID, outcome)
}

// GetPatterns returns patterns matching the filter.
func (r *DefaultRepository) GetPatterns(ctx context.Context, filter PatternFilter) ([]*DebatePattern, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*DebatePattern

	for _, pattern := range r.patterns {
		// Apply filters
		if len(filter.Types) > 0 {
			found := false
			for _, t := range filter.Types {
				if pattern.PatternType == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if filter.Domain != "" && pattern.Domain != filter.Domain {
			continue
		}

		if pattern.Frequency < filter.MinFrequency {
			continue
		}

		if pattern.SuccessRate < filter.MinSuccess {
			continue
		}

		if filter.Since != nil && pattern.LastObserved.Before(*filter.Since) {
			continue
		}

		results = append(results, pattern)
	}

	// Sort by frequency * success rate
	sort.Slice(results, func(i, j int) bool {
		scoreI := float64(results[i].Frequency) * results[i].SuccessRate
		scoreJ := float64(results[j].Frequency) * results[j].SuccessRate
		return scoreI > scoreJ
	})

	return results, nil
}

// RecordPattern records a new debate pattern.
func (r *DefaultRepository) RecordPattern(ctx context.Context, pattern *DebatePattern) error {
	if pattern.ID == "" {
		pattern.ID = uuid.New().String()
	}

	pattern.FirstObserved = time.Now()
	pattern.LastObserved = time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for similar existing pattern
	for id, existing := range r.patterns {
		if r.patternsAreSimilar(existing, pattern) {
			// Update existing pattern
			existing.Frequency++
			existing.LastObserved = time.Now()
			existing.SuccessRate = (existing.SuccessRate*float64(existing.Frequency-1) + pattern.SuccessRate) / float64(existing.Frequency)
			return nil
		}
		_ = id // Silence unused variable
	}

	// Add new pattern
	if len(r.patterns) >= r.config.MaxPatterns {
		r.evictOldestPattern()
	}

	r.patterns[pattern.ID] = pattern
	return nil
}

// patternsAreSimilar checks if two patterns are similar.
func (r *DefaultRepository) patternsAreSimilar(a, b *DebatePattern) bool {
	if a.PatternType != b.PatternType {
		return false
	}
	if a.Domain != b.Domain {
		return false
	}
	// Could add more sophisticated similarity checking
	return a.Name == b.Name
}

// evictOldestPattern removes the oldest, least used pattern.
func (r *DefaultRepository) evictOldestPattern() {
	var oldest *DebatePattern
	var oldestID string

	for id, pattern := range r.patterns {
		if oldest == nil || pattern.LastObserved.Before(oldest.LastObserved) {
			oldest = pattern
			oldestID = id
		}
	}

	if oldestID != "" {
		delete(r.patterns, oldestID)
	}
}

// GetSuccessfulStrategies returns successful strategies for a domain.
func (r *DefaultRepository) GetSuccessfulStrategies(ctx context.Context, domain agents.Domain) ([]*Strategy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*Strategy

	for _, strategy := range r.strategies {
		if domain != "" && strategy.Domain != domain {
			continue
		}

		if strategy.SuccessRate >= r.config.StrategyThreshold {
			results = append(results, strategy)
		}
	}

	// Sort by success rate * applications
	sort.Slice(results, func(i, j int) bool {
		scoreI := results[i].SuccessRate * float64(results[i].Applications)
		scoreJ := results[j].SuccessRate * float64(results[j].Applications)
		return scoreI > scoreJ
	})

	return results, nil
}

// GetKnowledgeForAgent retrieves curated knowledge for an agent.
func (r *DefaultRepository) GetKnowledgeForAgent(ctx context.Context, agent *agents.SpecializedAgent, topic string) (*AgentKnowledge, error) {
	domain := agent.Specialization.PrimaryDomain

	// Get relevant lessons
	lessons, err := r.GetRelevantLessons(ctx, topic, domain)
	if err != nil {
		return nil, err
	}

	// Get applicable patterns
	patterns, err := r.GetPatterns(ctx, PatternFilter{
		Domain:       domain,
		MinFrequency: 2,
		MinSuccess:   0.6,
	})
	if err != nil {
		return nil, err
	}

	// Generate domain insights
	insights := r.generateDomainInsights(domain, lessons)

	// Generate role guidance
	guidance := r.generateRoleGuidance(agent.PrimaryRole, domain)

	// Get historical context
	history := r.getHistoricalContext(topic, domain)

	return &AgentKnowledge{
		AgentID:            agent.ID,
		RelevantLessons:    lessons,
		ApplicablePatterns: patterns,
		DomainInsights:     insights,
		RoleGuidance:       guidance,
		HistoricalContext:  history,
		GeneratedAt:        time.Now(),
	}, nil
}

// generateDomainInsights generates insights based on lessons.
func (r *DefaultRepository) generateDomainInsights(domain agents.Domain, lessons []*LessonMatch) []string {
	insights := make([]string, 0)

	// Extract top insights from high-scoring lessons
	for i, match := range lessons {
		if i >= 5 || match.Score < 0.7 {
			break
		}

		lesson := match.Lesson
		if lesson.Content.Solution != "" {
			insights = append(insights, fmt.Sprintf("From '%s': %s",
				lesson.Title, truncate(lesson.Content.Solution, 200)))
		}
	}

	// Add domain-specific guidance
	switch domain {
	case agents.DomainSecurity:
		insights = append(insights, "Always consider security implications first")
	case agents.DomainArchitecture:
		insights = append(insights, "Focus on scalability and maintainability")
	case agents.DomainOptimization:
		insights = append(insights, "Measure before and after optimization")
	case agents.DomainDebug:
		insights = append(insights, "Reproduce the issue reliably before debugging")
	case agents.DomainCode:
		insights = append(insights, "Follow established patterns and conventions")
	}

	return insights
}

// generateRoleGuidance generates guidance based on role.
func (r *DefaultRepository) generateRoleGuidance(role topology.AgentRole, domain agents.Domain) []string {
	guidance := make([]string, 0)

	switch role {
	case topology.RoleProposer:
		guidance = append(guidance,
			"Generate clear, actionable proposals",
			"Include rationale and trade-offs",
			"Consider alternative approaches")
	case topology.RoleCritic:
		guidance = append(guidance,
			"Identify potential weaknesses",
			"Provide constructive feedback",
			"Suggest improvements")
	case topology.RoleReviewer:
		guidance = append(guidance,
			"Evaluate completeness and correctness",
			"Check for edge cases",
			"Verify alignment with requirements")
	case topology.RoleOptimizer:
		guidance = append(guidance,
			"Focus on efficiency improvements",
			"Balance performance with maintainability",
			"Quantify optimization benefits")
	case topology.RoleModerator:
		guidance = append(guidance,
			"Facilitate productive discussion",
			"Synthesize different perspectives",
			"Drive toward consensus")
	case topology.RoleArchitect:
		guidance = append(guidance,
			"Consider long-term implications",
			"Ensure consistency with overall design",
			"Plan for extensibility")
	}

	return guidance
}

// getHistoricalContext retrieves relevant historical context.
func (r *DefaultRepository) getHistoricalContext(topic string, domain agents.Domain) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	context := make([]string, 0)

	// Find similar past debates
	for _, entry := range r.history {
		if entry.Domain == domain && entry.Success {
			context = append(context, fmt.Sprintf(
				"Previous debate on '%s' achieved %.0f%% consensus in %v",
				truncate(entry.Topic, 50), entry.ConsensusScore*100, entry.Duration.Round(time.Second)))
		}
		if len(context) >= 3 {
			break
		}
	}

	return context
}

// GetDebateHistory retrieves debate history matching the filter.
func (r *DefaultRepository) GetDebateHistory(ctx context.Context, filter HistoryFilter) ([]*DebateHistoryEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*DebateHistoryEntry

	for _, entry := range r.history {
		if filter.Domain != "" && entry.Domain != filter.Domain {
			continue
		}

		if filter.SuccessOnly && !entry.Success {
			continue
		}

		if entry.ConsensusScore < filter.MinConsensus {
			continue
		}

		if filter.Since != nil && entry.Timestamp.Before(*filter.Since) {
			continue
		}

		results = append(results, entry)

		if filter.Limit > 0 && len(results) >= filter.Limit {
			break
		}
	}

	return results, nil
}

// recordHistory records a debate in history.
func (r *DefaultRepository) recordHistory(result *protocol.DebateResult, lessonsLearned int) {
	entry := &DebateHistoryEntry{
		ID:             result.ID,
		Topic:          result.Topic,
		Success:        result.Success,
		Duration:       result.Duration,
		Participants:   result.ParticipantCount,
		LessonsLearned: lessonsLearned,
		Timestamp:      result.EndTime,
	}

	if result.FinalConsensus != nil {
		entry.ConsensusScore = result.FinalConsensus.Confidence
		entry.Summary = result.FinalConsensus.Summary
	}

	// Infer domain from topic
	entry.Domain = r.inferDomainFromTopic(result.Topic)

	r.mu.Lock()
	r.history = append([]*DebateHistoryEntry{entry}, r.history...)

	// Trim history if too long
	if len(r.history) > r.config.MaxHistoryEntries {
		r.history = r.history[:r.config.MaxHistoryEntries]
	}
	r.mu.Unlock()
}

// inferDomainFromTopic infers domain from topic keywords.
func (r *DefaultRepository) inferDomainFromTopic(topic string) agents.Domain {
	topic = toLower(topic)

	if containsAny(topic, "security", "vulnerability", "auth", "encrypt") {
		return agents.DomainSecurity
	}
	if containsAny(topic, "architecture", "design", "system", "scale") {
		return agents.DomainArchitecture
	}
	if containsAny(topic, "performance", "optimize", "speed", "cache") {
		return agents.DomainOptimization
	}
	if containsAny(topic, "debug", "error", "bug", "fix") {
		return agents.DomainDebug
	}
	if containsAny(topic, "code", "implement", "function", "class") {
		return agents.DomainCode
	}

	return agents.DomainGeneral
}

// GetStatistics returns repository statistics.
func (r *DefaultRepository) GetStatistics(ctx context.Context) (*RepositoryStatistics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bankStats := r.lessonBank.GetStatistics()

	stats := &RepositoryStatistics{
		TotalLessons:       bankStats.TotalLessons,
		TotalPatterns:      len(r.patterns),
		TotalStrategies:    len(r.strategies),
		TotalDebates:       len(r.history),
		LessonsByDomain:    make(map[agents.Domain]int),
		PatternsByType:     make(map[PatternType]int),
		OverallSuccessRate: bankStats.OverallSuccessRate,
		LastUpdated:        time.Now(),
	}

	// Count lessons by domain
	for domain, lessons := range r.domainLessons {
		stats.LessonsByDomain[domain] = len(lessons)
	}

	// Count patterns by type
	for _, pattern := range r.patterns {
		stats.PatternsByType[pattern.PatternType]++
	}

	// Calculate avg lessons per debate
	if len(r.history) > 0 {
		totalLessons := 0
		for _, entry := range r.history {
			totalLessons += entry.LessonsLearned
		}
		stats.AvgLessonsPerDebate = float64(totalLessons) / float64(len(r.history))
	}

	// Get top categories
	categoryCount := bankStats.LessonsByCategory
	type catCount struct {
		cat   debate.LessonCategory
		count int
	}
	var counts []catCount
	for cat, count := range categoryCount {
		counts = append(counts, catCount{cat, count})
	}
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})
	for i := 0; i < min(5, len(counts)); i++ {
		stats.TopCategories = append(stats.TopCategories, counts[i].cat)
	}

	return stats, nil
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func toLower(s string) string {
	return strings.ToLower(s)
}

func containsAny(s string, substrs ...string) bool {
	s = strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
