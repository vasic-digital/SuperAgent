// Package knowledge provides cross-debate learning capabilities.
package knowledge

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"dev.helix.agent/internal/debate"
	"dev.helix.agent/internal/debate/agents"
	"dev.helix.agent/internal/debate/protocol"
	"dev.helix.agent/internal/debate/topology"
)

// CrossDebateLearner learns patterns and strategies across multiple debates.
type CrossDebateLearner struct {
	repository Repository

	// Pattern analysis
	patternAnalyzer *PatternAnalyzer

	// Strategy synthesis
	strategySynthesizer *StrategySynthesizer

	// Knowledge graph
	knowledgeGraph *KnowledgeGraph

	config LearningConfig
	mu     sync.RWMutex
}

// LearningConfig configures cross-debate learning.
type LearningConfig struct {
	// MinDebatesForPattern is minimum debates before pattern is established
	MinDebatesForPattern int `json:"min_debates_for_pattern"`
	// PatternConfidenceThreshold is minimum confidence for pattern
	PatternConfidenceThreshold float64 `json:"pattern_confidence_threshold"`
	// StrategyMinApplications is minimum applications for strategy
	StrategyMinApplications int `json:"strategy_min_applications"`
	// StrategySuccessThreshold is minimum success rate for strategy
	StrategySuccessThreshold float64 `json:"strategy_success_threshold"`
	// EnableKnowledgeGraph enables knowledge graph tracking
	EnableKnowledgeGraph bool `json:"enable_knowledge_graph"`
	// MaxGraphNodes is maximum nodes in knowledge graph
	MaxGraphNodes int `json:"max_graph_nodes"`
	// LearningDecayRate is the decay rate for old learnings
	LearningDecayRate float64 `json:"learning_decay_rate"`
}

// DefaultLearningConfig returns sensible defaults.
func DefaultLearningConfig() LearningConfig {
	return LearningConfig{
		MinDebatesForPattern:       3,
		PatternConfidenceThreshold: 0.7,
		StrategyMinApplications:    5,
		StrategySuccessThreshold:   0.75,
		EnableKnowledgeGraph:       true,
		MaxGraphNodes:              10000,
		LearningDecayRate:          0.05,
	}
}

// NewCrossDebateLearner creates a new cross-debate learner.
func NewCrossDebateLearner(repository Repository, config LearningConfig) *CrossDebateLearner {
	learner := &CrossDebateLearner{
		repository:          repository,
		patternAnalyzer:     NewPatternAnalyzer(),
		strategySynthesizer: NewStrategySynthesizer(),
		config:              config,
	}

	if config.EnableKnowledgeGraph {
		learner.knowledgeGraph = NewKnowledgeGraph(config.MaxGraphNodes)
	}

	return learner
}

// LearnFromDebate extracts learnings from a completed debate.
func (cdl *CrossDebateLearner) LearnFromDebate(ctx context.Context, result *protocol.DebateResult, lessons []*debate.Lesson) (*LearningOutcome, error) {
	cdl.mu.Lock()
	defer cdl.mu.Unlock()

	outcome := &LearningOutcome{
		DebateID:        result.ID,
		LearnedAt:       time.Now(),
		NewPatterns:     make([]*DebatePattern, 0),
		UpdatedPatterns: make([]string, 0),
		NewStrategies:   make([]*Strategy, 0),
		KnowledgeNodes:  make([]string, 0),
	}

	// Analyze for patterns
	patterns := cdl.patternAnalyzer.Analyze(result)
	for _, pattern := range patterns {
		if pattern.Confidence >= cdl.config.PatternConfidenceThreshold {
			if err := cdl.repository.RecordPattern(ctx, pattern); err == nil {
				outcome.NewPatterns = append(outcome.NewPatterns, pattern)
			}
		}
	}

	// Synthesize strategies
	if result.Success && result.FinalConsensus != nil && result.FinalConsensus.Confidence >= cdl.config.StrategySuccessThreshold {
		strategy := cdl.strategySynthesizer.Synthesize(result)
		if strategy != nil {
			outcome.NewStrategies = append(outcome.NewStrategies, strategy)
		}
	}

	// Update knowledge graph
	if cdl.knowledgeGraph != nil {
		nodes := cdl.knowledgeGraph.AddDebate(result, lessons)
		outcome.KnowledgeNodes = nodes
	}

	// Calculate learning quality
	outcome.QualityScore = cdl.calculateLearningQuality(outcome)

	return outcome, nil
}

// LearningOutcome represents the learnings from a debate.
type LearningOutcome struct {
	DebateID        string           `json:"debate_id"`
	LearnedAt       time.Time        `json:"learned_at"`
	NewPatterns     []*DebatePattern `json:"new_patterns"`
	UpdatedPatterns []string         `json:"updated_patterns"`
	NewStrategies   []*Strategy      `json:"new_strategies"`
	KnowledgeNodes  []string         `json:"knowledge_nodes"`
	QualityScore    float64          `json:"quality_score"`
}

// calculateLearningQuality calculates the quality of learnings.
func (cdl *CrossDebateLearner) calculateLearningQuality(outcome *LearningOutcome) float64 {
	score := 0.0

	// Patterns contribute
	score += float64(len(outcome.NewPatterns)) * 0.2
	score += float64(len(outcome.UpdatedPatterns)) * 0.1

	// Strategies contribute more
	score += float64(len(outcome.NewStrategies)) * 0.3

	// Knowledge nodes contribute
	score += float64(len(outcome.KnowledgeNodes)) * 0.05

	// Cap at 1.0
	return math.Min(score, 1.0)
}

// GetRecommendations gets learning-based recommendations for a new debate.
func (cdl *CrossDebateLearner) GetRecommendations(ctx context.Context, topic string, domain agents.Domain) (*DebateRecommendations, error) {
	recommendations := &DebateRecommendations{
		Topic:            topic,
		Domain:           domain,
		GeneratedAt:      time.Now(),
		TopologyAdvice:   make([]string, 0),
		RoleAdvice:       make(map[topology.AgentRole][]string),
		PatternWarnings:  make([]string, 0),
		SuggestedActions: make([]string, 0),
	}

	// Get successful strategies for domain
	strategies, err := cdl.repository.GetSuccessfulStrategies(ctx, domain)
	if err == nil && len(strategies) > 0 {
		recommendations.RecommendedStrategy = strategies[0]
		recommendations.TopologyAdvice = append(recommendations.TopologyAdvice,
			fmt.Sprintf("Use %s topology (%.0f%% success rate from %d debates)",
				strategies[0].TopologyType, strategies[0].SuccessRate*100, strategies[0].Applications))
	}

	// Get relevant patterns
	patterns, err := cdl.repository.GetPatterns(ctx, PatternFilter{
		Domain:       domain,
		MinFrequency: cdl.config.MinDebatesForPattern,
		MinSuccess:   0.6,
	})
	if err == nil {
		for _, pattern := range patterns {
			switch pattern.PatternType {
			case PatternTypeFailure:
				recommendations.PatternWarnings = append(recommendations.PatternWarnings,
					fmt.Sprintf("Warning: '%s' pattern detected - %s", pattern.Name, pattern.Description))
			case PatternTypeConsensusBuilding:
				recommendations.SuggestedActions = append(recommendations.SuggestedActions,
					fmt.Sprintf("Apply '%s' for better consensus", pattern.Name))
			}
		}
	}

	// Get role-specific advice from knowledge graph
	if cdl.knowledgeGraph != nil {
		for _, role := range []topology.AgentRole{
			topology.RoleProposer, topology.RoleCritic, topology.RoleReviewer,
			topology.RoleOptimizer, topology.RoleModerator,
		} {
			advice := cdl.knowledgeGraph.GetRoleAdvice(role, domain)
			if len(advice) > 0 {
				recommendations.RoleAdvice[role] = advice
			}
		}
	}

	// Get relevant lessons
	lessons, err := cdl.repository.GetRelevantLessons(ctx, topic, domain)
	if err == nil {
		recommendations.RelevantLessons = lessons
	}

	return recommendations, nil
}

// DebateRecommendations provides learning-based recommendations.
type DebateRecommendations struct {
	Topic               string                          `json:"topic"`
	Domain              agents.Domain                   `json:"domain"`
	GeneratedAt         time.Time                       `json:"generated_at"`
	RecommendedStrategy *Strategy                       `json:"recommended_strategy,omitempty"`
	TopologyAdvice      []string                        `json:"topology_advice"`
	RoleAdvice          map[topology.AgentRole][]string `json:"role_advice"`
	PatternWarnings     []string                        `json:"pattern_warnings"`
	SuggestedActions    []string                        `json:"suggested_actions"`
	RelevantLessons     []*LessonMatch                  `json:"relevant_lessons,omitempty"`
}

// ApplyDecay applies time-based decay to learnings.
func (cdl *CrossDebateLearner) ApplyDecay(ctx context.Context) error {
	cdl.mu.Lock()
	defer cdl.mu.Unlock()

	// Decay is applied to pattern confidence and strategy scores
	// This ensures recent learnings are weighted more heavily
	// Implementation would iterate through repository and apply decay

	return nil
}

// =============================================================================
// Pattern Analyzer
// =============================================================================

// PatternAnalyzer analyzes debates for recurring patterns.
type PatternAnalyzer struct {
	detectors []PatternDetector
}

// PatternDetector detects specific types of patterns.
type PatternDetector interface {
	Detect(result *protocol.DebateResult) []*DebatePattern
}

// NewPatternAnalyzer creates a new pattern analyzer.
func NewPatternAnalyzer() *PatternAnalyzer {
	return &PatternAnalyzer{
		detectors: []PatternDetector{
			&ConsensusPatternDetector{},
			&ConflictPatternDetector{},
			&ExpertisePatternDetector{},
			&FailurePatternDetector{},
			&OptimizationPatternDetector{},
		},
	}
}

// Analyze analyzes a debate result for patterns.
func (pa *PatternAnalyzer) Analyze(result *protocol.DebateResult) []*DebatePattern {
	patterns := make([]*DebatePattern, 0)

	for _, detector := range pa.detectors {
		detected := detector.Detect(result)
		patterns = append(patterns, detected...)
	}

	return patterns
}

// ConsensusPatternDetector detects consensus-building patterns.
type ConsensusPatternDetector struct{}

func (d *ConsensusPatternDetector) Detect(result *protocol.DebateResult) []*DebatePattern {
	patterns := make([]*DebatePattern, 0)

	if result.FinalConsensus == nil {
		return patterns
	}

	// High early consensus
	if len(result.Phases) > 0 && result.Phases[0].ConsensusLevel >= 0.8 {
		patterns = append(patterns, &DebatePattern{
			ID:          uuid.New().String(),
			Name:        "Early High Consensus",
			Description: "Achieved high consensus early in the debate",
			PatternType: PatternTypeConsensusBuilding,
			Frequency:   1,
			SuccessRate: result.FinalConsensus.Confidence,
			Confidence:  0.8,
			Indicators: []PatternIndicator{
				{Type: "early_consensus", Threshold: 0.8, Weight: 1.0},
			},
		})
	}

	// Progressive consensus building
	if len(result.Phases) >= 3 {
		improving := true
		for i := 1; i < len(result.Phases); i++ {
			if result.Phases[i].ConsensusLevel < result.Phases[i-1].ConsensusLevel {
				improving = false
				break
			}
		}
		if improving {
			patterns = append(patterns, &DebatePattern{
				ID:          uuid.New().String(),
				Name:        "Progressive Consensus",
				Description: "Consensus improved steadily through phases",
				PatternType: PatternTypeConsensusBuilding,
				Frequency:   1,
				SuccessRate: result.FinalConsensus.Confidence,
				Confidence:  0.85,
			})
		}
	}

	return patterns
}

// ConflictPatternDetector detects conflict patterns.
type ConflictPatternDetector struct{}

func (d *ConflictPatternDetector) Detect(result *protocol.DebateResult) []*DebatePattern {
	patterns := make([]*DebatePattern, 0)

	totalDisagreements := 0
	for _, phase := range result.Phases {
		totalDisagreements += len(phase.Disagreements)
	}

	if totalDisagreements > 0 && result.Success {
		patterns = append(patterns, &DebatePattern{
			ID:          uuid.New().String(),
			Name:        "Conflict Resolution",
			Description: fmt.Sprintf("Resolved %d disagreements to reach consensus", totalDisagreements),
			PatternType: PatternTypeConflictResolution,
			Frequency:   1,
			SuccessRate: 1.0,
			Confidence:  0.75,
			Metadata: map[string]interface{}{
				"disagreements_resolved": totalDisagreements,
			},
		})
	}

	return patterns
}

// ExpertisePatternDetector detects expertise patterns.
type ExpertisePatternDetector struct{}

func (d *ExpertisePatternDetector) Detect(result *protocol.DebateResult) []*DebatePattern {
	patterns := make([]*DebatePattern, 0)

	// Track high-confidence contributors
	expertContributions := make(map[string]int)
	for _, phase := range result.Phases {
		for _, resp := range phase.Responses {
			if resp.Confidence >= 0.85 {
				key := resp.Provider + "/" + resp.Model
				expertContributions[key]++
			}
		}
	}

	// Consistent expert contributor
	for contributor, count := range expertContributions {
		if count >= 3 {
			patterns = append(patterns, &DebatePattern{
				ID:          uuid.New().String(),
				Name:        "Expert Contributor",
				Description: fmt.Sprintf("%s contributed %d high-confidence responses", contributor, count),
				PatternType: PatternTypeExpertise,
				Frequency:   1,
				SuccessRate: 0.9,
				Confidence:  0.8,
				Metadata: map[string]interface{}{
					"contributor":   contributor,
					"contributions": count,
				},
			})
		}
	}

	return patterns
}

// FailurePatternDetector detects failure patterns.
type FailurePatternDetector struct{}

func (d *FailurePatternDetector) Detect(result *protocol.DebateResult) []*DebatePattern {
	patterns := make([]*DebatePattern, 0)

	if result.Success {
		return patterns
	}

	// Detect low consensus throughout
	lowConsensusCount := 0
	for _, phase := range result.Phases {
		if phase.ConsensusLevel < 0.5 {
			lowConsensusCount++
		}
	}

	if lowConsensusCount >= len(result.Phases)/2 {
		patterns = append(patterns, &DebatePattern{
			ID:          uuid.New().String(),
			Name:        "Persistent Low Consensus",
			Description: "Low consensus throughout most phases",
			PatternType: PatternTypeFailure,
			Frequency:   1,
			SuccessRate: 0.0,
			Confidence:  0.9,
			Responses: []PatternResponse{
				{
					Action:   "increase_rounds",
					Priority: 1,
				},
				{
					Action:   "add_moderator",
					Priority: 2,
				},
			},
		})
	}

	return patterns
}

// OptimizationPatternDetector detects optimization patterns.
type OptimizationPatternDetector struct{}

func (d *OptimizationPatternDetector) Detect(result *protocol.DebateResult) []*DebatePattern {
	patterns := make([]*DebatePattern, 0)

	// Fast convergence
	if result.EarlyExit && result.Success {
		patterns = append(patterns, &DebatePattern{
			ID:          uuid.New().String(),
			Name:        "Fast Convergence",
			Description: fmt.Sprintf("Reached consensus early: %s", result.EarlyExitReason),
			PatternType: PatternTypeOptimization,
			Frequency:   1,
			SuccessRate: 1.0,
			Confidence:  0.85,
		})
	}

	// High response quality
	if result.Metrics != nil && result.Metrics.AvgConfidence >= 0.8 {
		patterns = append(patterns, &DebatePattern{
			ID:          uuid.New().String(),
			Name:        "High Quality Responses",
			Description: fmt.Sprintf("Average confidence: %.0f%%", result.Metrics.AvgConfidence*100),
			PatternType: PatternTypeOptimization,
			Frequency:   1,
			SuccessRate: result.Metrics.AvgConfidence,
			Confidence:  0.8,
		})
	}

	return patterns
}

// =============================================================================
// Strategy Synthesizer
// =============================================================================

// StrategySynthesizer synthesizes strategies from successful debates.
type StrategySynthesizer struct{}

// NewStrategySynthesizer creates a new strategy synthesizer.
func NewStrategySynthesizer() *StrategySynthesizer {
	return &StrategySynthesizer{}
}

// Synthesize creates a strategy from a successful debate.
func (ss *StrategySynthesizer) Synthesize(result *protocol.DebateResult) *Strategy {
	if !result.Success || result.FinalConsensus == nil {
		return nil
	}

	strategy := &Strategy{
		ID:           uuid.New().String(),
		Name:         fmt.Sprintf("Strategy from debate %s", result.ID[:8]),
		Description:  fmt.Sprintf("Strategy achieving %.0f%% consensus on '%s'", result.FinalConsensus.Confidence*100, truncate(result.Topic, 50)),
		TopologyType: result.TopologyUsed,
		SuccessRate:  result.FinalConsensus.Confidence,
		Applications: 1,
		AvgConsensus: result.FinalConsensus.Confidence,
		AvgDuration:  result.Duration,
		Phases:       make([]PhaseStrategy, 0),
	}

	// Extract role configuration from metrics
	if result.Metrics != nil && result.Metrics.RoleContributions != nil {
		for role, count := range result.Metrics.RoleContributions {
			if count > 0 {
				strategy.RoleConfig = append(strategy.RoleConfig, RoleConfiguration{
					Role:  role,
					Count: count,
				})
			}
		}
	}

	// Extract phase strategies
	for _, phase := range result.Phases {
		phaseStrategy := PhaseStrategy{
			Phase:            phase.Phase,
			FocusAreas:       phase.KeyInsights,
			MinConfidence:    phase.ConsensusLevel,
			ExpectedInsights: len(phase.KeyInsights),
		}
		strategy.Phases = append(strategy.Phases, phaseStrategy)
	}

	return strategy
}

// =============================================================================
// Knowledge Graph
// =============================================================================

// KnowledgeGraph tracks relationships between debate concepts.
type KnowledgeGraph struct {
	nodes    map[string]*KnowledgeNode
	edges    []*KnowledgeEdge
	maxNodes int
	mu       sync.RWMutex
}

// KnowledgeNode represents a concept in the knowledge graph.
type KnowledgeNode struct {
	ID          string                 `json:"id"`
	Type        NodeType               `json:"type"`
	Label       string                 `json:"label"`
	Domain      agents.Domain          `json:"domain,omitempty"`
	Weight      float64                `json:"weight"`
	LastUpdated time.Time              `json:"last_updated"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NodeType categorizes knowledge nodes.
type NodeType string

const (
	NodeTypeTopic   NodeType = "topic"
	NodeTypeConcept NodeType = "concept"
	NodeTypePattern NodeType = "pattern"
	NodeTypeLesson  NodeType = "lesson"
	NodeTypeAgent   NodeType = "agent"
	NodeTypeOutcome NodeType = "outcome"
)

// KnowledgeEdge represents a relationship between nodes.
type KnowledgeEdge struct {
	FromID  string    `json:"from_id"`
	ToID    string    `json:"to_id"`
	Type    EdgeType  `json:"type"`
	Weight  float64   `json:"weight"`
	Created time.Time `json:"created"`
}

// EdgeType categorizes knowledge edges.
type EdgeType string

const (
	EdgeTypeRelatedTo   EdgeType = "related_to"
	EdgeTypeLeadsTo     EdgeType = "leads_to"
	EdgeTypeDerivedFrom EdgeType = "derived_from"
	EdgeTypeContributes EdgeType = "contributes"
	EdgeTypeConflicts   EdgeType = "conflicts"
)

// NewKnowledgeGraph creates a new knowledge graph.
func NewKnowledgeGraph(maxNodes int) *KnowledgeGraph {
	return &KnowledgeGraph{
		nodes:    make(map[string]*KnowledgeNode),
		edges:    make([]*KnowledgeEdge, 0),
		maxNodes: maxNodes,
	}
}

// AddDebate adds a debate's concepts to the knowledge graph.
func (kg *KnowledgeGraph) AddDebate(result *protocol.DebateResult, lessons []*debate.Lesson) []string {
	kg.mu.Lock()
	defer kg.mu.Unlock()

	addedNodes := make([]string, 0)

	// Add topic node
	topicNode := &KnowledgeNode{
		ID:          "topic:" + result.ID,
		Type:        NodeTypeTopic,
		Label:       result.Topic,
		Weight:      1.0,
		LastUpdated: time.Now(),
	}
	kg.nodes[topicNode.ID] = topicNode
	addedNodes = append(addedNodes, topicNode.ID)

	// Add outcome node
	outcomeNode := &KnowledgeNode{
		ID:    "outcome:" + result.ID,
		Type:  NodeTypeOutcome,
		Label: map[bool]string{true: "success", false: "failure"}[result.Success],
		Weight: func() float64 {
			if result.FinalConsensus != nil {
				return result.FinalConsensus.Confidence
			}
			return 0.5
		}(),
		LastUpdated: time.Now(),
	}
	kg.nodes[outcomeNode.ID] = outcomeNode
	addedNodes = append(addedNodes, outcomeNode.ID)

	// Connect topic to outcome
	kg.edges = append(kg.edges, &KnowledgeEdge{
		FromID:  topicNode.ID,
		ToID:    outcomeNode.ID,
		Type:    EdgeTypeLeadsTo,
		Weight:  1.0,
		Created: time.Now(),
	})

	// Add lesson nodes
	for _, lesson := range lessons {
		lessonNode := &KnowledgeNode{
			ID:          "lesson:" + lesson.ID,
			Type:        NodeTypeLesson,
			Label:       lesson.Title,
			Weight:      lesson.Statistics.SuccessRate(),
			LastUpdated: time.Now(),
		}
		kg.nodes[lessonNode.ID] = lessonNode
		addedNodes = append(addedNodes, lessonNode.ID)

		// Connect to topic
		kg.edges = append(kg.edges, &KnowledgeEdge{
			FromID:  topicNode.ID,
			ToID:    lessonNode.ID,
			Type:    EdgeTypeDerivedFrom,
			Weight:  0.8,
			Created: time.Now(),
		})
	}

	// Trim if necessary
	kg.trimIfNecessary()

	return addedNodes
}

// GetRoleAdvice gets advice for a role based on knowledge graph.
func (kg *KnowledgeGraph) GetRoleAdvice(role topology.AgentRole, domain agents.Domain) []string {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	advice := make([]string, 0)

	// Find high-weight lesson nodes
	var lessonNodes []*KnowledgeNode
	for _, node := range kg.nodes {
		if node.Type == NodeTypeLesson && node.Weight >= 0.7 {
			lessonNodes = append(lessonNodes, node)
		}
	}

	// Sort by weight
	sort.Slice(lessonNodes, func(i, j int) bool {
		return lessonNodes[i].Weight > lessonNodes[j].Weight
	})

	// Extract advice from top lessons
	for i, node := range lessonNodes {
		if i >= 3 {
			break
		}
		advice = append(advice, fmt.Sprintf("Consider: %s (%.0f%% success)", node.Label, node.Weight*100))
	}

	return advice
}

// trimIfNecessary removes old nodes if over capacity.
func (kg *KnowledgeGraph) trimIfNecessary() {
	if len(kg.nodes) <= kg.maxNodes {
		return
	}

	// Sort nodes by last updated
	type nodeTime struct {
		id   string
		time time.Time
	}
	var nodes []nodeTime
	for id, node := range kg.nodes {
		nodes = append(nodes, nodeTime{id, node.LastUpdated})
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].time.Before(nodes[j].time)
	})

	// Remove oldest nodes
	toRemove := len(kg.nodes) - kg.maxNodes
	for i := 0; i < toRemove; i++ {
		delete(kg.nodes, nodes[i].id)
	}
}

// GetNode retrieves a node by ID.
func (kg *KnowledgeGraph) GetNode(id string) (*KnowledgeNode, bool) {
	kg.mu.RLock()
	defer kg.mu.RUnlock()
	node, ok := kg.nodes[id]
	return node, ok
}

// GetConnections gets all connections for a node.
func (kg *KnowledgeGraph) GetConnections(nodeID string) []*KnowledgeEdge {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	connections := make([]*KnowledgeEdge, 0)
	for _, edge := range kg.edges {
		if edge.FromID == nodeID || edge.ToID == nodeID {
			connections = append(connections, edge)
		}
	}
	return connections
}

// Size returns the number of nodes in the graph.
func (kg *KnowledgeGraph) Size() int {
	kg.mu.RLock()
	defer kg.mu.RUnlock()
	return len(kg.nodes)
}
