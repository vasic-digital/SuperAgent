// Package cognitive provides Cognitive Self-Evolving Planning for AI debates.
// Implements the expectation-comparison-refinement loop from ACL 2025 MARBLE framework.
// Research shows +3% improvement through adaptive planning strategies.
package cognitive

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"dev.helix.agent/internal/debate/topology"
)

// PlanningConfig configures the cognitive planning system.
type PlanningConfig struct {
	// EnableLearning enables cross-round learning
	EnableLearning bool `json:"enable_learning"`
	// ExpectationThreshold is the minimum expected improvement
	ExpectationThreshold float64 `json:"expectation_threshold"`
	// AdaptationRate controls how quickly strategy adapts (0-1)
	AdaptationRate float64 `json:"adaptation_rate"`
	// MaxHistorySize limits the learning history
	MaxHistorySize int `json:"max_history_size"`
	// EnableMetaCognition enables reflection on planning effectiveness
	EnableMetaCognition bool `json:"enable_meta_cognition"`
}

// DefaultPlanningConfig returns sensible defaults.
func DefaultPlanningConfig() PlanningConfig {
	return PlanningConfig{
		EnableLearning:       true,
		ExpectationThreshold: 0.05,
		AdaptationRate:       0.3,
		MaxHistorySize:       100,
		EnableMetaCognition:  true,
	}
}

// Expectation represents expected outcomes for a phase.
type Expectation struct {
	Phase              topology.DebatePhase `json:"phase"`
	Round              int                  `json:"round"`
	ExpectedConfidence float64              `json:"expected_confidence"`
	ExpectedConsensus  float64              `json:"expected_consensus"`
	ExpectedInsights   int                  `json:"expected_insights"`
	ExpectedLatency    time.Duration        `json:"expected_latency"`
	KeyGoals           []string             `json:"key_goals"`
	RiskFactors        []string             `json:"risk_factors"`
	Timestamp          time.Time            `json:"timestamp"`
}

// Comparison represents the comparison between expected and actual results.
type Comparison struct {
	Phase topology.DebatePhase `json:"phase"`
	Round int                  `json:"round"`

	// Deltas (actual - expected)
	ConfidenceDelta float64       `json:"confidence_delta"`
	ConsensusDelta  float64       `json:"consensus_delta"`
	InsightsDelta   int           `json:"insights_delta"`
	LatencyDelta    time.Duration `json:"latency_delta"`

	// Scores
	OverallScore       float64  `json:"overall_score"` // 0-1, 1 = exceeded expectations
	GoalsAchieved      []string `json:"goals_achieved"`
	GoalsMissed        []string `json:"goals_missed"`
	RisksRealized      []string `json:"risks_realized"`
	UnexpectedOutcomes []string `json:"unexpected_outcomes"`

	Timestamp time.Time `json:"timestamp"`
}

// Refinement represents strategy adjustments based on comparisons.
type Refinement struct {
	Phase topology.DebatePhase `json:"phase"`
	Round int                  `json:"round"`

	// Adjustments
	ConfidenceAdjustment float64                        `json:"confidence_adjustment"`
	AgentPriorities      map[string]float64             `json:"agent_priorities"`
	RoleEmphasis         map[topology.AgentRole]float64 `json:"role_emphasis"`
	NewGoals             []string                       `json:"new_goals"`
	MitigationStrategies []string                       `json:"mitigation_strategies"`

	// Learning
	Insights        []LearningInsight `json:"insights"`
	SuccessPatterns []string          `json:"success_patterns"`
	FailurePatterns []string          `json:"failure_patterns"`

	AppliedAt time.Time `json:"applied_at"`
}

// LearningInsight represents a learned pattern.
type LearningInsight struct {
	Pattern    string    `json:"pattern"`
	Confidence float64   `json:"confidence"`
	Frequency  int       `json:"frequency"`
	Impact     float64   `json:"impact"` // Positive = helpful
	Source     string    `json:"source"` // Which comparison generated this
	LastSeen   time.Time `json:"last_seen"`
}

// CognitivePlanner implements the expectation-comparison-refinement loop.
type CognitivePlanner struct {
	config PlanningConfig

	// State
	expectations map[topology.DebatePhase]*Expectation
	comparisons  []*Comparison
	refinements  []*Refinement

	// Learning
	learningHistory []LearningInsight
	phaseBaselines  map[topology.DebatePhase]*PhaseBaseline

	// Meta-cognition
	planningMetrics *PlanningMetrics

	mu sync.RWMutex
}

// PhaseBaseline represents baseline expectations for a phase based on history.
type PhaseBaseline struct {
	Phase         topology.DebatePhase
	AvgConfidence float64
	AvgConsensus  float64
	AvgInsights   float64
	AvgLatency    time.Duration
	SampleCount   int
	LastUpdated   time.Time
}

// PlanningMetrics tracks the effectiveness of the planning system.
type PlanningMetrics struct {
	TotalExpectations  int       `json:"total_expectations"`
	TotalComparisons   int       `json:"total_comparisons"`
	TotalRefinements   int       `json:"total_refinements"`
	AccuracyRate       float64   `json:"accuracy_rate"`       // How often expectations were met
	ImprovementRate    float64   `json:"improvement_rate"`    // Improvement from refinements
	LearningEfficiency float64   `json:"learning_efficiency"` // Insights per comparison
	LastUpdated        time.Time `json:"last_updated"`
}

// NewCognitivePlanner creates a new cognitive planner.
func NewCognitivePlanner(config PlanningConfig) *CognitivePlanner {
	return &CognitivePlanner{
		config:          config,
		expectations:    make(map[topology.DebatePhase]*Expectation),
		comparisons:     make([]*Comparison, 0),
		refinements:     make([]*Refinement, 0),
		learningHistory: make([]LearningInsight, 0),
		phaseBaselines:  make(map[topology.DebatePhase]*PhaseBaseline),
		planningMetrics: &PlanningMetrics{},
	}
}

// SetExpectation sets the expected outcomes for a phase.
func (cp *CognitivePlanner) SetExpectation(ctx context.Context, phase topology.DebatePhase, round int, agents []*topology.Agent) *Expectation {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Calculate baseline expectations from history or defaults
	baseline := cp.getOrCreateBaseline(phase)

	// Adjust based on current agents
	avgScore := cp.calculateAverageScore(agents)
	scoreAdjustment := (avgScore - 7.0) / 10.0 // Normalize around 7.0

	// Get learning adjustments
	learningAdjustment := cp.getLearningAdjustment(phase)

	expectation := &Expectation{
		Phase:              phase,
		Round:              round,
		ExpectedConfidence: math.Min(1.0, baseline.AvgConfidence+scoreAdjustment+learningAdjustment),
		ExpectedConsensus:  math.Min(1.0, baseline.AvgConsensus+scoreAdjustment*0.5),
		ExpectedInsights:   int(baseline.AvgInsights * (1 + scoreAdjustment)),
		ExpectedLatency:    baseline.AvgLatency,
		KeyGoals:           cp.generateGoals(phase, round),
		RiskFactors:        cp.identifyRisks(phase, agents),
		Timestamp:          time.Now(),
	}

	cp.expectations[phase] = expectation
	cp.planningMetrics.TotalExpectations++

	return expectation
}

// Compare compares actual results against expectations.
func (cp *CognitivePlanner) Compare(ctx context.Context, phase topology.DebatePhase, round int, actualConfidence, actualConsensus float64, actualInsights int, actualLatency time.Duration, goalsAchieved, unexpectedOutcomes []string) *Comparison {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	expectation := cp.expectations[phase]
	if expectation == nil {
		// Create default expectation if none exists
		expectation = &Expectation{
			Phase:              phase,
			Round:              round,
			ExpectedConfidence: 0.7,
			ExpectedConsensus:  0.6,
			ExpectedInsights:   3,
			ExpectedLatency:    30 * time.Second,
		}
	}

	// Calculate deltas
	confidenceDelta := actualConfidence - expectation.ExpectedConfidence
	consensusDelta := actualConsensus - expectation.ExpectedConsensus
	insightsDelta := actualInsights - expectation.ExpectedInsights
	latencyDelta := actualLatency - expectation.ExpectedLatency

	// Calculate overall score
	overallScore := cp.calculateOverallScore(confidenceDelta, consensusDelta, insightsDelta, latencyDelta)

	// Identify missed goals
	goalsMissed := cp.identifyMissedGoals(expectation.KeyGoals, goalsAchieved)

	// Identify realized risks
	risksRealized := cp.identifyRealizedRisks(expectation.RiskFactors, unexpectedOutcomes)

	comparison := &Comparison{
		Phase:              phase,
		Round:              round,
		ConfidenceDelta:    confidenceDelta,
		ConsensusDelta:     consensusDelta,
		InsightsDelta:      insightsDelta,
		LatencyDelta:       latencyDelta,
		OverallScore:       overallScore,
		GoalsAchieved:      goalsAchieved,
		GoalsMissed:        goalsMissed,
		RisksRealized:      risksRealized,
		UnexpectedOutcomes: unexpectedOutcomes,
		Timestamp:          time.Now(),
	}

	cp.comparisons = append(cp.comparisons, comparison)
	cp.planningMetrics.TotalComparisons++

	// Update baseline with actual results
	cp.updateBaseline(phase, actualConfidence, actualConsensus, float64(actualInsights), actualLatency)

	// Extract learning insights
	if cp.config.EnableLearning {
		cp.extractLearningInsights(comparison)
	}

	// Update accuracy metrics
	cp.updateAccuracyMetrics(comparison)

	return comparison
}

// Refine generates strategy refinements based on comparison results.
func (cp *CognitivePlanner) Refine(ctx context.Context, comparison *Comparison, agents []*topology.Agent) *Refinement {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	refinement := &Refinement{
		Phase:                comparison.Phase,
		Round:                comparison.Round,
		AgentPriorities:      make(map[string]float64),
		RoleEmphasis:         make(map[topology.AgentRole]float64),
		NewGoals:             make([]string, 0),
		MitigationStrategies: make([]string, 0),
		Insights:             make([]LearningInsight, 0),
		SuccessPatterns:      make([]string, 0),
		FailurePatterns:      make([]string, 0),
		AppliedAt:            time.Now(),
	}

	// Adjust confidence expectation for next phase
	refinement.ConfidenceAdjustment = comparison.ConfidenceDelta * cp.config.AdaptationRate

	// Adjust agent priorities based on performance
	cp.adjustAgentPriorities(refinement, agents, comparison)

	// Adjust role emphasis based on phase performance
	cp.adjustRoleEmphasis(refinement, comparison)

	// Generate new goals based on missed goals
	for _, missed := range comparison.GoalsMissed {
		refinement.NewGoals = append(refinement.NewGoals, "Retry: "+missed)
	}

	// Generate mitigation strategies for realized risks
	for _, risk := range comparison.RisksRealized {
		refinement.MitigationStrategies = append(refinement.MitigationStrategies, "Mitigate: "+risk)
	}

	// Extract patterns
	if comparison.OverallScore > 0.7 {
		refinement.SuccessPatterns = cp.extractSuccessPatterns(comparison)
	} else if comparison.OverallScore < 0.3 {
		refinement.FailurePatterns = cp.extractFailurePatterns(comparison)
	}

	// Add recent learning insights
	refinement.Insights = cp.getRecentInsights(5)

	cp.refinements = append(cp.refinements, refinement)
	cp.planningMetrics.TotalRefinements++

	// Update improvement metrics
	cp.updateImprovementMetrics(refinement)

	return refinement
}

// GetNextPhaseStrategy returns optimized strategy for the next phase.
func (cp *CognitivePlanner) GetNextPhaseStrategy(currentPhase topology.DebatePhase) *PhaseStrategy {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	nextPhase := getNextPhase(currentPhase)

	// Get baseline for next phase
	baseline := cp.getOrCreateBaseline(nextPhase)

	// Get recent refinement for current phase
	var lastRefinement *Refinement
	for i := len(cp.refinements) - 1; i >= 0; i-- {
		if cp.refinements[i].Phase == currentPhase {
			lastRefinement = cp.refinements[i]
			break
		}
	}

	// Build strategy
	strategy := &PhaseStrategy{
		Phase:           nextPhase,
		RoleEmphasis:    make(map[topology.AgentRole]float64),
		AgentPriorities: make(map[string]float64),
		Goals:           cp.generateGoals(nextPhase, 1),
		Adjustments:     make([]string, 0),
	}

	// Apply learning from current phase
	if lastRefinement != nil {
		// Carry over role emphasis adjustments
		for role, emphasis := range lastRefinement.RoleEmphasis {
			strategy.RoleEmphasis[role] = emphasis
		}

		// Carry over agent priorities
		for agent, priority := range lastRefinement.AgentPriorities {
			strategy.AgentPriorities[agent] = priority
		}

		// Add adjustments based on success/failure patterns
		for _, pattern := range lastRefinement.SuccessPatterns {
			strategy.Adjustments = append(strategy.Adjustments, "Continue: "+pattern)
		}
		for _, pattern := range lastRefinement.FailurePatterns {
			strategy.Adjustments = append(strategy.Adjustments, "Avoid: "+pattern)
		}
	}

	// Set expectations based on baseline
	strategy.ExpectedConfidence = baseline.AvgConfidence
	strategy.ExpectedConsensus = baseline.AvgConsensus

	return strategy
}

// PhaseStrategy represents the optimized strategy for a phase.
type PhaseStrategy struct {
	Phase              topology.DebatePhase           `json:"phase"`
	RoleEmphasis       map[topology.AgentRole]float64 `json:"role_emphasis"`
	AgentPriorities    map[string]float64             `json:"agent_priorities"`
	Goals              []string                       `json:"goals"`
	Adjustments        []string                       `json:"adjustments"`
	ExpectedConfidence float64                        `json:"expected_confidence"`
	ExpectedConsensus  float64                        `json:"expected_consensus"`
}

// GetPlanningMetrics returns the current planning metrics.
func (cp *CognitivePlanner) GetPlanningMetrics() *PlanningMetrics {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	// Copy metrics
	metrics := *cp.planningMetrics
	return &metrics
}

// GetLearningHistory returns the learning history.
func (cp *CognitivePlanner) GetLearningHistory() []LearningInsight {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	history := make([]LearningInsight, len(cp.learningHistory))
	copy(history, cp.learningHistory)
	return history
}

// ReflectOnPerformance performs meta-cognition on planning effectiveness.
func (cp *CognitivePlanner) ReflectOnPerformance(ctx context.Context) *MetaCognitiveReport {
	if !cp.config.EnableMetaCognition {
		return nil
	}

	cp.mu.RLock()
	defer cp.mu.RUnlock()

	report := &MetaCognitiveReport{
		GeneratedAt:      time.Now(),
		TotalComparisons: len(cp.comparisons),
		TotalRefinements: len(cp.refinements),
		LearningInsights: len(cp.learningHistory),
	}

	// Analyze expectation accuracy
	if len(cp.comparisons) > 0 {
		totalAccuracy := 0.0
		for _, c := range cp.comparisons {
			totalAccuracy += c.OverallScore
		}
		report.ExpectationAccuracy = totalAccuracy / float64(len(cp.comparisons))
	}

	// Analyze refinement effectiveness
	if len(cp.refinements) > 1 {
		improvements := 0
		for i := 1; i < len(cp.refinements); i++ {
			// Check if comparison after refinement improved
			if i < len(cp.comparisons) {
				if cp.comparisons[i].OverallScore > cp.comparisons[i-1].OverallScore {
					improvements++
				}
			}
		}
		report.RefinementEffectiveness = float64(improvements) / float64(len(cp.refinements)-1)
	}

	// Identify top success patterns
	patternCounts := make(map[string]int)
	for _, r := range cp.refinements {
		for _, p := range r.SuccessPatterns {
			patternCounts[p]++
		}
	}

	type patternCount struct {
		pattern string
		count   int
	}
	patterns := make([]patternCount, 0, len(patternCounts))
	for p, c := range patternCounts {
		patterns = append(patterns, patternCount{p, c})
	}
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].count > patterns[j].count
	})

	report.TopSuccessPatterns = make([]string, 0)
	for i := 0; i < 5 && i < len(patterns); i++ {
		report.TopSuccessPatterns = append(report.TopSuccessPatterns, patterns[i].pattern)
	}

	// Generate recommendations
	report.Recommendations = cp.generateRecommendations(report)

	return report
}

// MetaCognitiveReport contains reflection on planning effectiveness.
type MetaCognitiveReport struct {
	GeneratedAt             time.Time `json:"generated_at"`
	TotalComparisons        int       `json:"total_comparisons"`
	TotalRefinements        int       `json:"total_refinements"`
	LearningInsights        int       `json:"learning_insights"`
	ExpectationAccuracy     float64   `json:"expectation_accuracy"`
	RefinementEffectiveness float64   `json:"refinement_effectiveness"`
	TopSuccessPatterns      []string  `json:"top_success_patterns"`
	Recommendations         []string  `json:"recommendations"`
}

// Helper methods

func (cp *CognitivePlanner) getOrCreateBaseline(phase topology.DebatePhase) *PhaseBaseline {
	baseline, exists := cp.phaseBaselines[phase]
	if !exists {
		// Create default baseline based on phase
		baseline = &PhaseBaseline{
			Phase:         phase,
			AvgConfidence: getDefaultConfidence(phase),
			AvgConsensus:  getDefaultConsensus(phase),
			AvgInsights:   getDefaultInsights(phase),
			AvgLatency:    30 * time.Second,
			SampleCount:   0,
			LastUpdated:   time.Now(),
		}
		cp.phaseBaselines[phase] = baseline
	}
	return baseline
}

func (cp *CognitivePlanner) updateBaseline(phase topology.DebatePhase, confidence, consensus, insights float64, latency time.Duration) {
	baseline := cp.getOrCreateBaseline(phase)

	// Exponential moving average
	alpha := cp.config.AdaptationRate
	baseline.AvgConfidence = alpha*confidence + (1-alpha)*baseline.AvgConfidence
	baseline.AvgConsensus = alpha*consensus + (1-alpha)*baseline.AvgConsensus
	baseline.AvgInsights = alpha*insights + (1-alpha)*baseline.AvgInsights
	baseline.AvgLatency = time.Duration(alpha*float64(latency) + (1-alpha)*float64(baseline.AvgLatency))
	baseline.SampleCount++
	baseline.LastUpdated = time.Now()
}

func (cp *CognitivePlanner) calculateAverageScore(agents []*topology.Agent) float64 {
	if len(agents) == 0 {
		return 7.0
	}
	total := 0.0
	for _, a := range agents {
		total += a.Score
	}
	return total / float64(len(agents))
}

func (cp *CognitivePlanner) getLearningAdjustment(phase topology.DebatePhase) float64 {
	adjustment := 0.0

	for _, insight := range cp.learningHistory {
		if insight.Impact > 0 {
			adjustment += insight.Impact * insight.Confidence * 0.01
		}
	}

	return math.Min(0.1, adjustment) // Cap adjustment
}

func (cp *CognitivePlanner) generateGoals(phase topology.DebatePhase, round int) []string {
	switch phase {
	case topology.PhaseProposal:
		return []string{
			"Generate creative solutions",
			"Cover multiple perspectives",
			"Provide clear reasoning",
		}
	case topology.PhaseCritique:
		return []string{
			"Identify logical weaknesses",
			"Find missing considerations",
			"Assess risk factors",
		}
	case topology.PhaseReview:
		return []string{
			"Evaluate proposal quality",
			"Rate viability objectively",
			"Suggest improvements",
		}
	case topology.PhaseOptimization:
		return []string{
			"Synthesize best ideas",
			"Address valid critiques",
			"Improve practicality",
		}
	case topology.PhaseConvergence:
		return []string{
			"Reach consensus",
			"Resolve disagreements",
			"Validate final solution",
		}
	default:
		return []string{"Complete phase successfully"}
	}
}

func (cp *CognitivePlanner) identifyRisks(phase topology.DebatePhase, agents []*topology.Agent) []string {
	risks := make([]string, 0)

	// Agent-based risks
	if len(agents) < 3 {
		risks = append(risks, "Low agent diversity")
	}

	lowConfidenceAgents := 0
	for _, a := range agents {
		if a.Confidence < 0.5 {
			lowConfidenceAgents++
		}
	}
	if lowConfidenceAgents > len(agents)/2 {
		risks = append(risks, "Many low-confidence agents")
	}

	// Phase-specific risks
	switch phase {
	case topology.PhaseProposal:
		risks = append(risks, "Proposals may be too similar")
	case topology.PhaseCritique:
		risks = append(risks, "Critiques may be overly harsh")
	case topology.PhaseConvergence:
		risks = append(risks, "May not reach consensus")
	}

	return risks
}

func (cp *CognitivePlanner) calculateOverallScore(confDelta, consDelta float64, insightDelta int, latencyDelta time.Duration) float64 {
	// Normalize deltas to 0-1 score (0.5 = met expectations)
	confScore := 0.5 + confDelta/2
	consScore := 0.5 + consDelta/2
	insightScore := 0.5 + float64(insightDelta)/10
	latencyScore := 0.5 - float64(latencyDelta.Milliseconds())/60000 // Faster is better

	// Clamp scores
	confScore = math.Max(0, math.Min(1, confScore))
	consScore = math.Max(0, math.Min(1, consScore))
	insightScore = math.Max(0, math.Min(1, insightScore))
	latencyScore = math.Max(0, math.Min(1, latencyScore))

	// Weighted average
	return confScore*0.3 + consScore*0.3 + insightScore*0.2 + latencyScore*0.2
}

func (cp *CognitivePlanner) identifyMissedGoals(expected, achieved []string) []string {
	achievedMap := make(map[string]bool)
	for _, g := range achieved {
		achievedMap[g] = true
	}

	missed := make([]string, 0)
	for _, g := range expected {
		if !achievedMap[g] {
			missed = append(missed, g)
		}
	}
	return missed
}

func (cp *CognitivePlanner) identifyRealizedRisks(risks, outcomes []string) []string {
	realized := make([]string, 0)
	for _, risk := range risks {
		for _, outcome := range outcomes {
			// Simple string matching - could be more sophisticated
			if len(risk) > 5 && len(outcome) > 5 {
				if risk[:5] == outcome[:5] { // Very simplified matching
					realized = append(realized, risk)
					break
				}
			}
		}
	}
	return realized
}

func (cp *CognitivePlanner) adjustAgentPriorities(refinement *Refinement, agents []*topology.Agent, comparison *Comparison) {
	// Increase priority for high-performing agents (based on role contributions)
	for _, agent := range agents {
		priority := 1.0

		// Adjust based on phase performance
		if comparison.OverallScore > 0.7 {
			priority *= 1.1 // Boost all agents in successful phase
		} else if comparison.OverallScore < 0.3 {
			priority *= 0.9 // Slightly reduce for poor phase
		}

		// Adjust based on agent confidence
		priority *= (0.5 + agent.Confidence*0.5)

		refinement.AgentPriorities[agent.ID] = priority
	}
}

func (cp *CognitivePlanner) adjustRoleEmphasis(refinement *Refinement, comparison *Comparison) {
	switch comparison.Phase {
	case topology.PhaseProposal:
		if comparison.OverallScore < 0.5 {
			refinement.RoleEmphasis[topology.RoleProposer] = 1.2
			refinement.RoleEmphasis[topology.RoleArchitect] = 1.1
		}
	case topology.PhaseCritique:
		if comparison.OverallScore < 0.5 {
			refinement.RoleEmphasis[topology.RoleCritic] = 1.2
			refinement.RoleEmphasis[topology.RoleRedTeam] = 1.1
		}
	case topology.PhaseReview:
		if comparison.OverallScore < 0.5 {
			refinement.RoleEmphasis[topology.RoleReviewer] = 1.2
		}
	case topology.PhaseOptimization:
		if comparison.OverallScore < 0.5 {
			refinement.RoleEmphasis[topology.RoleOptimizer] = 1.2
		}
	case topology.PhaseConvergence:
		if comparison.OverallScore < 0.5 {
			refinement.RoleEmphasis[topology.RoleModerator] = 1.2
			refinement.RoleEmphasis[topology.RoleValidator] = 1.1
		}
	}
}

func (cp *CognitivePlanner) extractSuccessPatterns(comparison *Comparison) []string {
	patterns := make([]string, 0)

	if comparison.ConfidenceDelta > 0.1 {
		patterns = append(patterns, "High confidence achieved")
	}
	if comparison.ConsensusDelta > 0.1 {
		patterns = append(patterns, "Strong consensus building")
	}
	if comparison.InsightsDelta > 2 {
		patterns = append(patterns, "Rich insight generation")
	}
	if len(comparison.GoalsAchieved) > len(comparison.GoalsMissed) {
		patterns = append(patterns, "Goal-oriented execution")
	}

	return patterns
}

func (cp *CognitivePlanner) extractFailurePatterns(comparison *Comparison) []string {
	patterns := make([]string, 0)

	if comparison.ConfidenceDelta < -0.2 {
		patterns = append(patterns, "Low confidence responses")
	}
	if comparison.ConsensusDelta < -0.2 {
		patterns = append(patterns, "Poor consensus building")
	}
	if comparison.InsightsDelta < -3 {
		patterns = append(patterns, "Insufficient insights")
	}
	if len(comparison.RisksRealized) > 0 {
		patterns = append(patterns, "Risk materialization")
	}

	return patterns
}

func (cp *CognitivePlanner) extractLearningInsights(comparison *Comparison) {
	insight := LearningInsight{
		Confidence: comparison.OverallScore,
		Frequency:  1,
		Impact:     comparison.OverallScore - 0.5, // Positive if above expectations
		Source:     fmt.Sprintf("%s-round-%d", comparison.Phase, comparison.Round),
		LastSeen:   time.Now(),
	}

	if comparison.OverallScore > 0.7 {
		insight.Pattern = fmt.Sprintf("Success in %s phase", comparison.Phase)
	} else if comparison.OverallScore < 0.3 {
		insight.Pattern = fmt.Sprintf("Struggle in %s phase", comparison.Phase)
	} else {
		insight.Pattern = fmt.Sprintf("Normal performance in %s phase", comparison.Phase)
	}

	// Check for existing similar insight
	found := false
	for i, existing := range cp.learningHistory {
		if existing.Pattern == insight.Pattern {
			cp.learningHistory[i].Frequency++
			cp.learningHistory[i].LastSeen = time.Now()
			cp.learningHistory[i].Confidence = (existing.Confidence + insight.Confidence) / 2
			found = true
			break
		}
	}

	if !found {
		cp.learningHistory = append(cp.learningHistory, insight)

		// Enforce max history size
		if len(cp.learningHistory) > cp.config.MaxHistorySize {
			// Remove oldest, lowest-impact insight
			sort.Slice(cp.learningHistory, func(i, j int) bool {
				return cp.learningHistory[i].Impact*float64(cp.learningHistory[i].Frequency) >
					cp.learningHistory[j].Impact*float64(cp.learningHistory[j].Frequency)
			})
			cp.learningHistory = cp.learningHistory[:cp.config.MaxHistorySize]
		}
	}
}

func (cp *CognitivePlanner) getRecentInsights(n int) []LearningInsight {
	if len(cp.learningHistory) <= n {
		result := make([]LearningInsight, len(cp.learningHistory))
		copy(result, cp.learningHistory)
		return result
	}

	// Sort by recency
	sorted := make([]LearningInsight, len(cp.learningHistory))
	copy(sorted, cp.learningHistory)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LastSeen.After(sorted[j].LastSeen)
	})

	return sorted[:n]
}

func (cp *CognitivePlanner) updateAccuracyMetrics(comparison *Comparison) {
	// Rolling accuracy calculation
	totalScore := 0.0
	for _, c := range cp.comparisons {
		totalScore += c.OverallScore
	}
	cp.planningMetrics.AccuracyRate = totalScore / float64(len(cp.comparisons))
	cp.planningMetrics.LastUpdated = time.Now()
}

func (cp *CognitivePlanner) updateImprovementMetrics(refinement *Refinement) {
	if len(cp.comparisons) < 2 {
		return
	}

	improvements := 0
	for i := 1; i < len(cp.comparisons); i++ {
		if cp.comparisons[i].OverallScore > cp.comparisons[i-1].OverallScore {
			improvements++
		}
	}

	cp.planningMetrics.ImprovementRate = float64(improvements) / float64(len(cp.comparisons)-1)
	cp.planningMetrics.LearningEfficiency = float64(len(cp.learningHistory)) / float64(len(cp.comparisons))
}

func (cp *CognitivePlanner) generateRecommendations(report *MetaCognitiveReport) []string {
	recommendations := make([]string, 0)

	if report.ExpectationAccuracy < 0.5 {
		recommendations = append(recommendations, "Adjust expectations to be more realistic")
	}

	if report.RefinementEffectiveness < 0.3 {
		recommendations = append(recommendations, "Increase adaptation rate for faster learning")
	}

	if report.LearningInsights < report.TotalComparisons/2 {
		recommendations = append(recommendations, "Enable more aggressive insight extraction")
	}

	if len(report.TopSuccessPatterns) > 0 {
		recommendations = append(recommendations, "Reinforce success patterns: "+report.TopSuccessPatterns[0])
	}

	return recommendations
}

// Default values by phase
func getDefaultConfidence(phase topology.DebatePhase) float64 {
	switch phase {
	case topology.PhaseProposal:
		return 0.7
	case topology.PhaseCritique:
		return 0.75
	case topology.PhaseReview:
		return 0.7
	case topology.PhaseOptimization:
		return 0.75
	case topology.PhaseConvergence:
		return 0.8
	default:
		return 0.7
	}
}

func getDefaultConsensus(phase topology.DebatePhase) float64 {
	switch phase {
	case topology.PhaseProposal:
		return 0.5
	case topology.PhaseCritique:
		return 0.6
	case topology.PhaseReview:
		return 0.65
	case topology.PhaseOptimization:
		return 0.7
	case topology.PhaseConvergence:
		return 0.75
	default:
		return 0.6
	}
}

func getDefaultInsights(phase topology.DebatePhase) float64 {
	switch phase {
	case topology.PhaseProposal:
		return 5
	case topology.PhaseCritique:
		return 7
	case topology.PhaseReview:
		return 4
	case topology.PhaseOptimization:
		return 3
	case topology.PhaseConvergence:
		return 2
	default:
		return 3
	}
}

func getNextPhase(current topology.DebatePhase) topology.DebatePhase {
	switch current {
	case topology.PhaseProposal:
		return topology.PhaseCritique
	case topology.PhaseCritique:
		return topology.PhaseReview
	case topology.PhaseReview:
		return topology.PhaseOptimization
	case topology.PhaseOptimization:
		return topology.PhaseConvergence
	default:
		return topology.PhaseProposal
	}
}
