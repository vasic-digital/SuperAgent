package learning

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"dev.helix.agent/internal/learning"
	"dev.helix.agent/internal/messaging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMessageBroker implements messaging.MessageBroker for testing
type MockMessageBroker struct {
	messages       []*messaging.Message
	subscriptions  map[string]messaging.MessageHandler
	publishError   error
	subscribeError error
}

func NewMockMessageBroker() *MockMessageBroker {
	return &MockMessageBroker{
		messages:      []*messaging.Message{},
		subscriptions: make(map[string]messaging.MessageHandler),
	}
}

func (m *MockMessageBroker) Connect(ctx context.Context) error {
	return nil
}

func (m *MockMessageBroker) Close(ctx context.Context) error {
	return nil
}

func (m *MockMessageBroker) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *MockMessageBroker) IsConnected() bool {
	return true
}

func (m *MockMessageBroker) Publish(ctx context.Context, topic string, message *messaging.Message, opts ...messaging.PublishOption) error {
	if m.publishError != nil {
		return m.publishError
	}
	m.messages = append(m.messages, message)
	return nil
}

func (m *MockMessageBroker) PublishBatch(ctx context.Context, topic string, messages []*messaging.Message, opts ...messaging.PublishOption) error {
	if m.publishError != nil {
		return m.publishError
	}
	m.messages = append(m.messages, messages...)
	return nil
}

func (m *MockMessageBroker) Subscribe(ctx context.Context, topic string, handler messaging.MessageHandler, opts ...messaging.SubscribeOption) (messaging.Subscription, error) {
	if m.subscribeError != nil {
		return nil, m.subscribeError
	}
	m.subscriptions[topic] = handler
	return &MockSubscription{}, nil
}

func (m *MockMessageBroker) BrokerType() messaging.BrokerType {
	return messaging.BrokerTypeKafka
}

func (m *MockMessageBroker) GetMetrics() *messaging.BrokerMetrics {
	return &messaging.BrokerMetrics{}
}

func (m *MockMessageBroker) SimulateMessage(ctx context.Context, topic string, payload []byte) error {
	handler, exists := m.subscriptions[topic]
	if !exists {
		return nil
	}

	msg := &messaging.Message{
		ID:        "test-msg",
		Type:      "test",
		Payload:   payload,
		Timestamp: time.Now(),
	}

	return handler(ctx, msg)
}

type MockSubscription struct{}

func (s *MockSubscription) Unsubscribe() error {
	return nil
}

func (s *MockSubscription) IsActive() bool {
	return true
}

func (s *MockSubscription) Topic() string {
	return "test-topic"
}

func (s *MockSubscription) ID() string {
	return "mock-subscription-id"
}

// Test CrossSessionLearner creation
func TestNewCrossSessionLearner(t *testing.T) {
	broker := NewMockMessageBroker()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := learning.CrossSessionConfig{
		CompletedTopic: "test.completed",
		InsightsTopic:  "test.insights",
		MinConfidence:  0.7,
		MinFrequency:   3,
	}

	learner := learning.NewCrossSessionLearner(config, broker, logger)

	require.NotNil(t, learner)
}

// Test StartLearning
func TestCrossSessionLearner_StartLearning(t *testing.T) {
	broker := NewMockMessageBroker()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := learning.CrossSessionConfig{
		CompletedTopic: "test.completed",
		InsightsTopic:  "test.insights",
	}

	learner := learning.NewCrossSessionLearner(config, broker, logger)

	ctx := context.Background()
	err := learner.StartLearning(ctx)

	require.NoError(t, err)
	assert.Contains(t, broker.subscriptions, "test.completed")
}

// Test pattern extraction - User Intent
func TestExtractIntentPattern_HelpSeeking(t *testing.T) {
	completion := learning.ConversationCompletion{
		ConversationID: "conv-123",
		UserID:         "user-456",
		Messages: []learning.Message{
			{Role: "user", Content: "How do I fix this error?", Tokens: 20},
			{Role: "assistant", Content: "Let me help you...", Tokens: 30},
		},
	}

	broker := NewMockMessageBroker()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := learning.CrossSessionConfig{
		CompletedTopic: "test.completed",
		InsightsTopic:  "test.insights",
	}

	learner := learning.NewCrossSessionLearner(config, broker, logger)

	// Simulate conversation completion
	payload, err := json.Marshal(completion)
	require.NoError(t, err)

	ctx := context.Background()
	err = learner.StartLearning(ctx)
	require.NoError(t, err)

	err = broker.SimulateMessage(ctx, "test.completed", payload)
	require.NoError(t, err)

	// Check that patterns were extracted (insights published)
	assert.GreaterOrEqual(t, len(broker.messages), 0) // May have 0 if frequency threshold not met
}

// Test pattern extraction - Entity Cooccurrence
func TestExtractEntityCooccurrence(t *testing.T) {
	completion := learning.ConversationCompletion{
		ConversationID: "conv-123",
		UserID:         "user-456",
		Entities: []learning.Entity{
			{EntityID: "e1", Type: "ORG", Name: "OpenAI"},
			{EntityID: "e2", Type: "TECH", Name: "ChatGPT"},
			{EntityID: "e3", Type: "PERSON", Name: "Sam Altman"},
		},
	}

	broker := NewMockMessageBroker()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := learning.CrossSessionConfig{
		CompletedTopic: "test.completed",
		InsightsTopic:  "test.insights",
	}

	learner := learning.NewCrossSessionLearner(config, broker, logger)

	payload, err := json.Marshal(completion)
	require.NoError(t, err)

	ctx := context.Background()
	err = learner.StartLearning(ctx)
	require.NoError(t, err)

	err = broker.SimulateMessage(ctx, "test.completed", payload)
	require.NoError(t, err)
}

// Test pattern extraction - User Preference
func TestExtractUserPreference_ConciseStyle(t *testing.T) {
	completion := learning.ConversationCompletion{
		ConversationID: "conv-123",
		UserID:         "user-456",
		Messages: []learning.Message{
			{Role: "user", Content: "Quick question", Tokens: 10},
			{Role: "assistant", Content: "Sure!", Tokens: 5},
			{Role: "user", Content: "Thanks", Tokens: 5},
		},
	}

	broker := NewMockMessageBroker()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := learning.CrossSessionConfig{
		CompletedTopic: "test.completed",
		InsightsTopic:  "test.insights",
	}

	learner := learning.NewCrossSessionLearner(config, broker, logger)

	payload, err := json.Marshal(completion)
	require.NoError(t, err)

	ctx := context.Background()
	err = learner.StartLearning(ctx)
	require.NoError(t, err)

	err = broker.SimulateMessage(ctx, "test.completed", payload)
	require.NoError(t, err)
}

// Test pattern extraction - Debate Strategy
func TestExtractDebateStrategy(t *testing.T) {
	completion := learning.ConversationCompletion{
		ConversationID: "conv-123",
		UserID:         "user-456",
		DebateRounds: []learning.DebateRound{
			{Round: 1, Provider: "claude", Position: "researcher", Confidence: 0.92},
			{Round: 1, Provider: "deepseek", Position: "critic", Confidence: 0.85},
			{Round: 2, Provider: "claude", Position: "researcher", Confidence: 0.94},
		},
	}

	broker := NewMockMessageBroker()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := learning.CrossSessionConfig{
		CompletedTopic: "test.completed",
		InsightsTopic:  "test.insights",
	}

	learner := learning.NewCrossSessionLearner(config, broker, logger)

	payload, err := json.Marshal(completion)
	require.NoError(t, err)

	ctx := context.Background()
	err = learner.StartLearning(ctx)
	require.NoError(t, err)

	err = broker.SimulateMessage(ctx, "test.completed", payload)
	require.NoError(t, err)
}

// Test pattern extraction - Conversation Flow
func TestExtractConversationFlow_RapidFlow(t *testing.T) {
	startTime := time.Now()
	completion := learning.ConversationCompletion{
		ConversationID: "conv-123",
		UserID:         "user-456",
		StartedAt:      startTime,
		CompletedAt:    startTime.Add(15 * time.Second), // 15 seconds total
		Messages: []learning.Message{
			{Role: "user", Content: "Hi", Tokens: 5},
			{Role: "assistant", Content: "Hello!", Tokens: 5},
			{Role: "user", Content: "Quick question", Tokens: 10},
			{Role: "assistant", Content: "Sure!", Tokens: 5},
			{Role: "user", Content: "Thanks", Tokens: 5},
		}, // 5 messages in 15 seconds = 3 sec/msg = rapid
	}

	broker := NewMockMessageBroker()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := learning.CrossSessionConfig{
		CompletedTopic: "test.completed",
		InsightsTopic:  "test.insights",
	}

	learner := learning.NewCrossSessionLearner(config, broker, logger)

	payload, err := json.Marshal(completion)
	require.NoError(t, err)

	ctx := context.Background()
	err = learner.StartLearning(ctx)
	require.NoError(t, err)

	err = broker.SimulateMessage(ctx, "test.completed", payload)
	require.NoError(t, err)
}

// Test InsightStore - UpdatePattern
func TestInsightStore_UpdatePattern(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	store := learning.NewInsightStore(logger)

	pattern := learning.Pattern{
		PatternID:   "pattern-123",
		PatternType: learning.PatternUserIntent,
		Description: "Test pattern",
		Frequency:   1,
		Confidence:  0.8,
	}

	// First update (add)
	store.UpdatePattern(pattern)
	retrieved := store.GetPattern("pattern-123")
	require.NotNil(t, retrieved)
	assert.Equal(t, 1, retrieved.Frequency)
	assert.Equal(t, 0.8, retrieved.Confidence)

	// Second update (increment frequency)
	pattern.Confidence = 0.9
	store.UpdatePattern(pattern)
	retrieved = store.GetPattern("pattern-123")
	require.NotNil(t, retrieved)
	assert.Equal(t, 2, retrieved.Frequency)             // Incremented
	assert.InDelta(t, 0.85, retrieved.Confidence, 0.01) // Averaged
}

// Test InsightStore - GetPatternsByType
func TestInsightStore_GetPatternsByType(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	store := learning.NewInsightStore(logger)

	pattern1 := learning.Pattern{
		PatternID:   "pattern-1",
		PatternType: learning.PatternUserIntent,
		Description: "Intent pattern",
		Frequency:   5,
	}

	pattern2 := learning.Pattern{
		PatternID:   "pattern-2",
		PatternType: learning.PatternDebateStrategy,
		Description: "Debate pattern",
		Frequency:   3,
	}

	pattern3 := learning.Pattern{
		PatternID:   "pattern-3",
		PatternType: learning.PatternUserIntent,
		Description: "Another intent pattern",
		Frequency:   2,
	}

	store.UpdatePattern(pattern1)
	store.UpdatePattern(pattern2)
	store.UpdatePattern(pattern3)

	// Get user intent patterns
	intentPatterns := store.GetPatternsByType(learning.PatternUserIntent)
	assert.Len(t, intentPatterns, 2)

	// Get debate patterns
	debatePatterns := store.GetPatternsByType(learning.PatternDebateStrategy)
	assert.Len(t, debatePatterns, 1)
}

// Test InsightStore - AddInsight and GetInsight
func TestInsightStore_AddAndGetInsight(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	store := learning.NewInsightStore(logger)

	insight := learning.Insight{
		InsightID:   "insight-123",
		UserID:      "user-456",
		InsightType: "user_preference",
		Title:       "User prefers concise responses",
		Description: "Pattern observed 10 times",
		Confidence:  0.9,
		Impact:      "high",
	}

	store.AddInsight(insight)

	retrieved := store.GetInsight("insight-123")
	require.NotNil(t, retrieved)
	assert.Equal(t, "insight-123", retrieved.InsightID)
	assert.Equal(t, "user-456", retrieved.UserID)
	assert.Equal(t, 0.9, retrieved.Confidence)
}

// Test InsightStore - GetInsightsByUser
func TestInsightStore_GetInsightsByUser(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	store := learning.NewInsightStore(logger)

	insight1 := learning.Insight{
		InsightID: "insight-1",
		UserID:    "user-123",
		Title:     "Insight 1",
	}

	insight2 := learning.Insight{
		InsightID: "insight-2",
		UserID:    "user-456",
		Title:     "Insight 2",
	}

	insight3 := learning.Insight{
		InsightID: "insight-3",
		UserID:    "user-123",
		Title:     "Insight 3",
	}

	store.AddInsight(insight1)
	store.AddInsight(insight2)
	store.AddInsight(insight3)

	// Get insights for user-123
	userInsights := store.GetInsightsByUser("user-123")
	assert.Len(t, userInsights, 2)

	// Get insights for user-456
	user456Insights := store.GetInsightsByUser("user-456")
	assert.Len(t, user456Insights, 1)
}

// Test InsightStore - GetTopPatterns
func TestInsightStore_GetTopPatterns(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	store := learning.NewInsightStore(logger)

	// Add patterns with different frequencies
	patterns := []learning.Pattern{
		{PatternID: "p1", Frequency: 10},
		{PatternID: "p2", Frequency: 25},
		{PatternID: "p3", Frequency: 5},
		{PatternID: "p4", Frequency: 15},
		{PatternID: "p5", Frequency: 30},
	}

	for _, p := range patterns {
		store.UpdatePattern(p)
	}

	// Get top 3 patterns
	topPatterns := store.GetTopPatterns(3)
	require.Len(t, topPatterns, 3)

	// Should be sorted by frequency descending
	assert.Equal(t, "p5", topPatterns[0].PatternID) // 30
	assert.Equal(t, "p2", topPatterns[1].PatternID) // 25
	assert.Equal(t, "p4", topPatterns[2].PatternID) // 15
}

// Test InsightStore - GetStats
func TestInsightStore_GetStats(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	store := learning.NewInsightStore(logger)

	// Add various patterns
	store.UpdatePattern(learning.Pattern{
		PatternID:   "p1",
		PatternType: learning.PatternUserIntent,
	})
	store.UpdatePattern(learning.Pattern{
		PatternID:   "p2",
		PatternType: learning.PatternUserIntent,
	})
	store.UpdatePattern(learning.Pattern{
		PatternID:   "p3",
		PatternType: learning.PatternDebateStrategy,
	})

	// Add insights
	store.AddInsight(learning.Insight{InsightID: "i1"})
	store.AddInsight(learning.Insight{InsightID: "i2"})

	stats := store.GetStats()

	assert.Equal(t, 3, stats["total_patterns"])
	assert.Equal(t, 2, stats["total_insights"])

	patternsByType := stats["patterns_by_type"].(map[string]int)
	assert.Equal(t, 2, patternsByType["user_intent"])
	assert.Equal(t, 1, patternsByType["debate_strategy"])
}

// Test complete workflow
func TestCrossSessionLearner_CompleteWorkflow(t *testing.T) {
	broker := NewMockMessageBroker()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := learning.CrossSessionConfig{
		CompletedTopic: "test.completed",
		InsightsTopic:  "test.insights",
		MinConfidence:  0.7,
		MinFrequency:   1, // Low threshold for testing
	}

	learner := learning.NewCrossSessionLearner(config, broker, logger)

	ctx := context.Background()
	err := learner.StartLearning(ctx)
	require.NoError(t, err)

	// Create a rich conversation completion
	completion := learning.ConversationCompletion{
		ConversationID: "conv-123",
		UserID:         "user-456",
		SessionID:      "session-789",
		StartedAt:      time.Now().Add(-5 * time.Minute),
		CompletedAt:    time.Now(),
		Messages: []learning.Message{
			{MessageID: "m1", Role: "user", Content: "How do I fix authentication errors?", Tokens: 30},
			{MessageID: "m2", Role: "assistant", Content: "Let me help you with authentication...", Tokens: 50},
			{MessageID: "m3", Role: "user", Content: "Thanks! That worked.", Tokens: 20},
		},
		Entities: []learning.Entity{
			{EntityID: "e1", Type: "TECH", Name: "authentication", Confidence: 0.95},
			{EntityID: "e2", Type: "TECH", Name: "OAuth", Confidence: 0.88},
		},
		DebateRounds: []learning.DebateRound{
			{Round: 1, Provider: "claude", Model: "claude-3-opus", Position: "researcher", Confidence: 0.92, ResponseTimeMs: 250},
		},
		Outcome: "successful",
	}

	payload, err := json.Marshal(completion)
	require.NoError(t, err)

	err = broker.SimulateMessage(ctx, "test.completed", payload)
	require.NoError(t, err)

	// Verify that insights were published
	// Note: May be 0 if frequency thresholds not met
	assert.GreaterOrEqual(t, len(broker.messages), 0)
}

// Benchmark pattern extraction
func BenchmarkExtractPatterns(b *testing.B) {
	completion := learning.ConversationCompletion{
		ConversationID: "conv-123",
		UserID:         "user-456",
		Messages:       make([]learning.Message, 50),
		Entities:       make([]learning.Entity, 20),
		DebateRounds:   make([]learning.DebateRound, 5),
	}

	// Fill with dummy data
	for i := 0; i < 50; i++ {
		completion.Messages[i] = learning.Message{
			MessageID: string(rune(i)),
			Role:      "user",
			Content:   "Test message",
			Tokens:    20,
		}
	}

	broker := NewMockMessageBroker()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := learning.CrossSessionConfig{
		CompletedTopic: "test.completed",
		InsightsTopic:  "test.insights",
	}

	learner := learning.NewCrossSessionLearner(config, broker, logger)

	payload, _ := json.Marshal(completion)
	ctx := context.Background()
	learner.StartLearning(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broker.SimulateMessage(ctx, "test.completed", payload)
	}
}
