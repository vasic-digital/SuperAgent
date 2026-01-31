package learning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dev.helix.agent/internal/messaging"
	"github.com/sirupsen/logrus"
)

// CrossSessionLearner learns patterns across multiple conversations and sessions
type CrossSessionLearner struct {
	broker         messaging.MessageBroker
	insights       *InsightStore
	logger         *logrus.Logger
	completedTopic string
	insightsTopic  string
}

// ConversationCompletion represents a completed conversation
type ConversationCompletion struct {
	ConversationID string                 `json:"conversation_id"`
	UserID         string                 `json:"user_id"`
	SessionID      string                 `json:"session_id"`
	StartedAt      time.Time              `json:"started_at"`
	CompletedAt    time.Time              `json:"completed_at"`
	Messages       []Message              `json:"messages"`
	Entities       []Entity               `json:"entities"`
	DebateRounds   []DebateRound          `json:"debate_rounds,omitempty"`
	Outcome        string                 `json:"outcome"` // "successful", "abandoned", "error"
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Message represents a conversation message
type Message struct {
	MessageID string    `json:"message_id"`
	Role      string    `json:"role"` // "user", "assistant", "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Tokens    int       `json:"tokens"`
}

// Entity represents an extracted entity
type Entity struct {
	EntityID   string                 `json:"entity_id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Value      string                 `json:"value"`
	Confidence float64                `json:"confidence"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// DebateRound represents a debate round
type DebateRound struct {
	Round          int                    `json:"round"`
	Position       string                 `json:"position"`
	Provider       string                 `json:"provider"`
	Model          string                 `json:"model"`
	Response       string                 `json:"response"`
	Confidence     float64                `json:"confidence"`
	ResponseTimeMs int64                  `json:"response_time_ms"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// PatternType defines types of patterns that can be learned
type PatternType string

const (
	// PatternUserIntent represents user intent patterns
	PatternUserIntent PatternType = "user_intent"

	// PatternDebateStrategy represents successful debate strategies
	PatternDebateStrategy PatternType = "debate_strategy"

	// PatternEntityCooccurrence represents entity co-occurrence patterns
	PatternEntityCooccurrence PatternType = "entity_cooccurrence"

	// PatternUserPreference represents user preference patterns
	PatternUserPreference PatternType = "user_preference"

	// PatternConversationFlow represents conversation flow patterns
	PatternConversationFlow PatternType = "conversation_flow"

	// PatternProviderPerformance represents provider performance patterns
	PatternProviderPerformance PatternType = "provider_performance"
)

// Pattern represents a learned pattern
type Pattern struct {
	PatternID   string                 `json:"pattern_id"`
	PatternType PatternType            `json:"pattern_type"`
	Description string                 `json:"description"`
	Frequency   int                    `json:"frequency"`
	Confidence  float64                `json:"confidence"`
	Examples    []string               `json:"examples,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	FirstSeen   time.Time              `json:"first_seen"`
	LastSeen    time.Time              `json:"last_seen"`
}

// Insight represents a learned insight
type Insight struct {
	InsightID   string                 `json:"insight_id"`
	UserID      string                 `json:"user_id,omitempty"` // Empty for global insights
	InsightType string                 `json:"insight_type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Impact      string                 `json:"impact"` // "high", "medium", "low"
	Patterns    []Pattern              `json:"patterns"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// InsightStore manages learned insights
type InsightStore struct {
	insights map[string]*Insight
	patterns map[string]*Pattern
	logger   *logrus.Logger
}

// CrossSessionConfig defines configuration for cross-session learning
type CrossSessionConfig struct {
	CompletedTopic string
	InsightsTopic  string
	MinConfidence  float64
	MinFrequency   int
}

// NewCrossSessionLearner creates a new cross-session learner
func NewCrossSessionLearner(
	config CrossSessionConfig,
	broker messaging.MessageBroker,
	logger *logrus.Logger,
) *CrossSessionLearner {
	if logger == nil {
		logger = logrus.New()
	}

	return &CrossSessionLearner{
		broker:         broker,
		insights:       NewInsightStore(logger),
		logger:         logger,
		completedTopic: config.CompletedTopic,
		insightsTopic:  config.InsightsTopic,
	}
}

// NewInsightStore creates a new insight store
func NewInsightStore(logger *logrus.Logger) *InsightStore {
	return &InsightStore{
		insights: make(map[string]*Insight),
		patterns: make(map[string]*Pattern),
		logger:   logger,
	}
}

// StartLearning begins consuming completed conversations and learning patterns
func (csl *CrossSessionLearner) StartLearning(ctx context.Context) error {
	csl.logger.Info("Starting cross-session learning")

	// Subscribe to completed conversations topic
	handler := func(ctx context.Context, msg *messaging.Message) error {
		return csl.processCompletion(ctx, msg)
	}

	_, err := csl.broker.Subscribe(ctx, csl.completedTopic, handler)
	if err != nil {
		return fmt.Errorf("failed to subscribe to completed topic: %w", err)
	}

	csl.logger.WithField("topic", csl.completedTopic).Info("Subscribed to conversation completions")

	return nil
}

// processCompletion processes a completed conversation
func (csl *CrossSessionLearner) processCompletion(ctx context.Context, msg *messaging.Message) error {
	// Parse completion
	var completion ConversationCompletion
	if err := json.Unmarshal(msg.Payload, &completion); err != nil {
		return fmt.Errorf("failed to unmarshal completion: %w", err)
	}

	csl.logger.WithFields(logrus.Fields{
		"conversation_id": completion.ConversationID,
		"user_id":         completion.UserID,
		"message_count":   len(completion.Messages),
		"entity_count":    len(completion.Entities),
	}).Debug("Processing conversation completion")

	// Extract patterns
	patterns := csl.extractPatterns(completion)

	// Update pattern frequencies
	for _, pattern := range patterns {
		csl.insights.UpdatePattern(pattern)
	}

	// Generate insights
	insights := csl.generateInsights(completion, patterns)

	// Publish insights
	for _, insight := range insights {
		if err := csl.publishInsight(ctx, insight); err != nil {
			csl.logger.WithError(err).Error("Failed to publish insight")
		}
	}

	return nil
}

// extractPatterns extracts patterns from a completed conversation
func (csl *CrossSessionLearner) extractPatterns(completion ConversationCompletion) []Pattern {
	patterns := []Pattern{}

	// Extract user intent patterns
	intentPattern := csl.extractIntentPattern(completion)
	if intentPattern != nil {
		patterns = append(patterns, *intentPattern)
	}

	// Extract debate strategy patterns
	if len(completion.DebateRounds) > 0 {
		debatePattern := csl.extractDebateStrategy(completion)
		if debatePattern != nil {
			patterns = append(patterns, *debatePattern)
		}
	}

	// Extract entity co-occurrence patterns
	entityPattern := csl.extractEntityCooccurrence(completion)
	if entityPattern != nil {
		patterns = append(patterns, *entityPattern)
	}

	// Extract user preference patterns
	prefPattern := csl.extractUserPreference(completion)
	if prefPattern != nil {
		patterns = append(patterns, *prefPattern)
	}

	// Extract conversation flow patterns
	flowPattern := csl.extractConversationFlow(completion)
	if flowPattern != nil {
		patterns = append(patterns, *flowPattern)
	}

	return patterns
}

// extractIntentPattern extracts user intent patterns
func (csl *CrossSessionLearner) extractIntentPattern(completion ConversationCompletion) *Pattern {
	// Analyze user messages to identify intent
	userMessages := []string{}
	for _, msg := range completion.Messages {
		if msg.Role == "user" {
			userMessages = append(userMessages, msg.Content)
		}
	}

	if len(userMessages) == 0 {
		return nil
	}

	// Identify primary intent from first message
	firstMsg := strings.ToLower(userMessages[0])
	var intent string

	if strings.Contains(firstMsg, "help") || strings.Contains(firstMsg, "how") {
		intent = "help_seeking"
	} else if strings.Contains(firstMsg, "explain") || strings.Contains(firstMsg, "what is") {
		intent = "explanation_request"
	} else if strings.Contains(firstMsg, "fix") || strings.Contains(firstMsg, "error") {
		intent = "problem_solving"
	} else if strings.Contains(firstMsg, "create") || strings.Contains(firstMsg, "build") {
		intent = "creation_task"
	} else {
		intent = "general_inquiry"
	}

	return &Pattern{
		PatternID:   fmt.Sprintf("intent-%s-%d", intent, time.Now().UnixNano()),
		PatternType: PatternUserIntent,
		Description: fmt.Sprintf("User intent: %s", intent),
		Frequency:   1,
		Confidence:  0.8,
		Examples:    []string{userMessages[0]},
		Metadata: map[string]interface{}{
			"intent":        intent,
			"message_count": len(userMessages),
		},
		FirstSeen: time.Now(),
		LastSeen:  time.Now(),
	}
}

// extractDebateStrategy extracts successful debate strategies
func (csl *CrossSessionLearner) extractDebateStrategy(completion ConversationCompletion) *Pattern {
	if len(completion.DebateRounds) == 0 {
		return nil
	}

	// Identify which provider/position combinations were most successful
	providerScores := make(map[string]float64)
	positionScores := make(map[string]float64)

	for _, round := range completion.DebateRounds {
		providerScores[round.Provider] += round.Confidence
		positionScores[round.Position] += round.Confidence
	}

	// Find best provider
	var bestProvider string
	var bestScore float64
	for provider, score := range providerScores {
		if score > bestScore {
			bestProvider = provider
			bestScore = score
		}
	}

	return &Pattern{
		PatternID:   fmt.Sprintf("debate-strategy-%d", time.Now().UnixNano()),
		PatternType: PatternDebateStrategy,
		Description: fmt.Sprintf("Successful debate strategy: %s provider", bestProvider),
		Frequency:   1,
		Confidence:  bestScore / float64(len(completion.DebateRounds)),
		Metadata: map[string]interface{}{
			"best_provider":  bestProvider,
			"avg_confidence": bestScore / float64(len(completion.DebateRounds)),
			"total_rounds":   len(completion.DebateRounds),
		},
		FirstSeen: time.Now(),
		LastSeen:  time.Now(),
	}
}

// extractEntityCooccurrence extracts entity co-occurrence patterns
func (csl *CrossSessionLearner) extractEntityCooccurrence(completion ConversationCompletion) *Pattern {
	if len(completion.Entities) < 2 {
		return nil
	}

	// Build co-occurrence map
	cooccurrences := make(map[string]int)
	for i := 0; i < len(completion.Entities); i++ {
		for j := i + 1; j < len(completion.Entities); j++ {
			pair := fmt.Sprintf("%s-%s", completion.Entities[i].Type, completion.Entities[j].Type)
			cooccurrences[pair]++
		}
	}

	// Find most common co-occurrence
	var maxPair string
	var maxCount int
	for pair, count := range cooccurrences {
		if count > maxCount {
			maxPair = pair
			maxCount = count
		}
	}

	if maxCount == 0 {
		return nil
	}

	return &Pattern{
		PatternID:   fmt.Sprintf("entity-cooccurrence-%d", time.Now().UnixNano()),
		PatternType: PatternEntityCooccurrence,
		Description: fmt.Sprintf("Entities often co-occur: %s", maxPair),
		Frequency:   maxCount,
		Confidence:  float64(maxCount) / float64(len(completion.Entities)),
		Metadata: map[string]interface{}{
			"entity_pair": maxPair,
			"count":       maxCount,
		},
		FirstSeen: time.Now(),
		LastSeen:  time.Now(),
	}
}

// extractUserPreference extracts user preference patterns
func (csl *CrossSessionLearner) extractUserPreference(completion ConversationCompletion) *Pattern {
	// Analyze user's interaction style
	userMsgCount := 0
	totalTokens := 0
	for _, msg := range completion.Messages {
		if msg.Role == "user" {
			userMsgCount++
			totalTokens += msg.Tokens
		}
	}

	if userMsgCount == 0 {
		return nil
	}

	avgMsgLength := totalTokens / userMsgCount

	var style string
	if avgMsgLength < 50 {
		style = "concise"
	} else if avgMsgLength < 200 {
		style = "moderate"
	} else {
		style = "detailed"
	}

	return &Pattern{
		PatternID:   fmt.Sprintf("user-pref-%s-%d", completion.UserID, time.Now().UnixNano()),
		PatternType: PatternUserPreference,
		Description: fmt.Sprintf("User prefers %s communication style", style),
		Frequency:   1,
		Confidence:  0.7,
		Metadata: map[string]interface{}{
			"style":          style,
			"avg_msg_length": avgMsgLength,
			"message_count":  userMsgCount,
		},
		FirstSeen: time.Now(),
		LastSeen:  time.Now(),
	}
}

// extractConversationFlow extracts conversation flow patterns
func (csl *CrossSessionLearner) extractConversationFlow(completion ConversationCompletion) *Pattern {
	if len(completion.Messages) < 3 {
		return nil
	}

	// Analyze conversation structure
	duration := completion.CompletedAt.Sub(completion.StartedAt)
	avgTimePerMsg := duration.Milliseconds() / int64(len(completion.Messages))

	var flowType string
	if avgTimePerMsg < 5000 { // < 5 seconds
		flowType = "rapid"
	} else if avgTimePerMsg < 30000 { // < 30 seconds
		flowType = "normal"
	} else {
		flowType = "thoughtful"
	}

	return &Pattern{
		PatternID:   fmt.Sprintf("conv-flow-%d", time.Now().UnixNano()),
		PatternType: PatternConversationFlow,
		Description: fmt.Sprintf("Conversation flow: %s", flowType),
		Frequency:   1,
		Confidence:  0.75,
		Metadata: map[string]interface{}{
			"flow_type":        flowType,
			"avg_time_per_msg": avgTimePerMsg,
			"duration_ms":      duration.Milliseconds(),
		},
		FirstSeen: time.Now(),
		LastSeen:  time.Now(),
	}
}

// generateInsights generates insights from patterns
func (csl *CrossSessionLearner) generateInsights(completion ConversationCompletion, patterns []Pattern) []Insight {
	insights := []Insight{}

	// Generate insight for high-frequency patterns
	for _, pattern := range patterns {
		if pattern.Frequency >= 3 && pattern.Confidence >= 0.7 {
			insight := Insight{
				InsightID:   fmt.Sprintf("insight-%d", time.Now().UnixNano()),
				UserID:      completion.UserID,
				InsightType: string(pattern.PatternType),
				Title:       fmt.Sprintf("Learned Pattern: %s", pattern.Description),
				Description: fmt.Sprintf("Pattern observed %d times with %.2f confidence", pattern.Frequency, pattern.Confidence),
				Confidence:  pattern.Confidence,
				Impact:      csl.determineImpact(pattern),
				Patterns:    []Pattern{pattern},
				Metadata: map[string]interface{}{
					"conversation_id": completion.ConversationID,
				},
				CreatedAt: time.Now(),
			}
			insights = append(insights, insight)
		}
	}

	return insights
}

// determineImpact determines the impact level of a pattern
func (csl *CrossSessionLearner) determineImpact(pattern Pattern) string {
	if pattern.Frequency >= 10 && pattern.Confidence >= 0.9 {
		return "high"
	} else if pattern.Frequency >= 5 && pattern.Confidence >= 0.7 {
		return "medium"
	}
	return "low"
}

// publishInsight publishes an insight to Kafka
func (csl *CrossSessionLearner) publishInsight(ctx context.Context, insight Insight) error {
	payload, err := json.Marshal(insight)
	if err != nil {
		return fmt.Errorf("failed to marshal insight: %w", err)
	}

	msg := &messaging.Message{
		ID:        fmt.Sprintf("insight-%d", time.Now().UnixNano()),
		Type:      "learning.insight",
		Payload:   payload,
		Timestamp: time.Now(),
	}

	if err := csl.broker.Publish(ctx, csl.insightsTopic, msg); err != nil {
		return fmt.Errorf("failed to publish insight: %w", err)
	}

	csl.logger.WithFields(logrus.Fields{
		"insight_id":   insight.InsightID,
		"insight_type": insight.InsightType,
		"confidence":   insight.Confidence,
	}).Debug("Insight published")

	return nil
}

// InsightStore methods

// UpdatePattern updates or adds a pattern to the store
func (is *InsightStore) UpdatePattern(pattern Pattern) {
	existing, exists := is.patterns[pattern.PatternID]
	if exists {
		existing.Frequency++
		existing.LastSeen = time.Now()
		existing.Confidence = (existing.Confidence + pattern.Confidence) / 2
	} else {
		is.patterns[pattern.PatternID] = &pattern
	}
}

// GetPattern retrieves a pattern by ID
func (is *InsightStore) GetPattern(patternID string) *Pattern {
	return is.patterns[patternID]
}

// GetPatternsByType retrieves all patterns of a specific type
func (is *InsightStore) GetPatternsByType(patternType PatternType) []Pattern {
	patterns := []Pattern{}
	for _, pattern := range is.patterns {
		if pattern.PatternType == patternType {
			patterns = append(patterns, *pattern)
		}
	}
	return patterns
}

// AddInsight adds an insight to the store
func (is *InsightStore) AddInsight(insight Insight) {
	is.insights[insight.InsightID] = &insight
}

// GetInsight retrieves an insight by ID
func (is *InsightStore) GetInsight(insightID string) *Insight {
	return is.insights[insightID]
}

// GetInsightsByUser retrieves all insights for a specific user
func (is *InsightStore) GetInsightsByUser(userID string) []Insight {
	insights := []Insight{}
	for _, insight := range is.insights {
		if insight.UserID == userID {
			insights = append(insights, *insight)
		}
	}
	return insights
}

// GetTopPatterns retrieves top patterns by frequency
func (is *InsightStore) GetTopPatterns(limit int) []Pattern {
	patterns := []Pattern{}
	for _, pattern := range is.patterns {
		patterns = append(patterns, *pattern)
	}

	// Sort by frequency (simple bubble sort for small datasets)
	for i := 0; i < len(patterns); i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[j].Frequency > patterns[i].Frequency {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}

	if len(patterns) > limit {
		patterns = patterns[:limit]
	}

	return patterns
}

// GetStats returns statistics about learned patterns and insights
func (is *InsightStore) GetStats() map[string]interface{} {
	patternsByType := make(map[string]int)
	for _, pattern := range is.patterns {
		patternsByType[string(pattern.PatternType)]++
	}

	return map[string]interface{}{
		"total_patterns":   len(is.patterns),
		"total_insights":   len(is.insights),
		"patterns_by_type": patternsByType,
	}
}
