package learning

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"dev.helix.agent/internal/messaging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

type mockBroker struct {
	publishedMessages []*messaging.Message
	subscribeHandler  messaging.MessageHandler
	subscribeTopic    string
	subscribeErr      error
	publishErr        error
}

func (m *mockBroker) Connect(_ context.Context) error     { return nil }
func (m *mockBroker) Close(_ context.Context) error       { return nil }
func (m *mockBroker) HealthCheck(_ context.Context) error { return nil }
func (m *mockBroker) IsConnected() bool                   { return true }
func (m *mockBroker) BrokerType() messaging.BrokerType    { return messaging.BrokerTypeInMemory }
func (m *mockBroker) GetMetrics() *messaging.BrokerMetrics {
	return &messaging.BrokerMetrics{}
}

func (m *mockBroker) Publish(_ context.Context, _ string, msg *messaging.Message, _ ...messaging.PublishOption) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.publishedMessages = append(m.publishedMessages, msg)
	return nil
}

func (m *mockBroker) PublishBatch(_ context.Context, _ string, msgs []*messaging.Message, _ ...messaging.PublishOption) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.publishedMessages = append(m.publishedMessages, msgs...)
	return nil
}

func (m *mockBroker) Subscribe(_ context.Context, topic string, handler messaging.MessageHandler, _ ...messaging.SubscribeOption) (messaging.Subscription, error) {
	if m.subscribeErr != nil {
		return nil, m.subscribeErr
	}
	m.subscribeHandler = handler
	m.subscribeTopic = topic
	return &mockSubscription{topic: topic}, nil
}

type mockSubscription struct {
	topic    string
	active   bool
	unsubErr error
}

func (s *mockSubscription) Unsubscribe() error { s.active = false; return s.unsubErr }
func (s *mockSubscription) IsActive() bool     { return s.active }
func (s *mockSubscription) Topic() string      { return s.topic }
func (s *mockSubscription) ID() string         { return "mock-sub-" + s.topic }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestConfig() CrossSessionConfig {
	return CrossSessionConfig{
		CompletedTopic: "conversations.completed",
		InsightsTopic:  "learning.insights",
		MinConfidence:  0.7,
		MinFrequency:   3,
	}
}

func newTestLogger() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)
	return l
}

func newTestCompletion() ConversationCompletion {
	now := time.Now()
	return ConversationCompletion{
		ConversationID: "conv-1",
		UserID:         "user-1",
		SessionID:      "sess-1",
		StartedAt:      now.Add(-5 * time.Minute),
		CompletedAt:    now,
		Messages: []Message{
			{MessageID: "m1", Role: "user", Content: "Help me fix this error", Timestamp: now.Add(-5 * time.Minute), Tokens: 30},
			{MessageID: "m2", Role: "assistant", Content: "Sure, let me look at that.", Timestamp: now.Add(-4 * time.Minute), Tokens: 40},
			{MessageID: "m3", Role: "user", Content: "Thanks", Timestamp: now.Add(-3 * time.Minute), Tokens: 10},
		},
		Entities: []Entity{
			{EntityID: "e1", Type: "language", Name: "Go", Value: "go", Confidence: 0.95},
			{EntityID: "e2", Type: "framework", Name: "Gin", Value: "gin", Confidence: 0.9},
		},
		Outcome: "successful",
	}
}

// ---------------------------------------------------------------------------
// InsightStore tests
// ---------------------------------------------------------------------------

func TestNewInsightStore(t *testing.T) {
	store := NewInsightStore(newTestLogger())
	assert.NotNil(t, store)
	assert.NotNil(t, store.insights)
	assert.NotNil(t, store.patterns)
	assert.Empty(t, store.insights)
	assert.Empty(t, store.patterns)
}

func TestInsightStore_UpdatePattern_New(t *testing.T) {
	store := NewInsightStore(newTestLogger())
	pattern := Pattern{
		PatternID:   "p1",
		PatternType: PatternUserIntent,
		Description: "test pattern",
		Frequency:   1,
		Confidence:  0.8,
		Metadata:    map[string]interface{}{"key": "val"},
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
	}

	store.UpdatePattern(pattern)

	stored := store.GetPattern("p1")
	assert.NotNil(t, stored)
	assert.Equal(t, "p1", stored.PatternID)
	assert.Equal(t, PatternUserIntent, stored.PatternType)
	assert.Equal(t, 1, stored.Frequency)
	assert.Equal(t, 0.8, stored.Confidence)
}

func TestInsightStore_UpdatePattern_Existing(t *testing.T) {
	store := NewInsightStore(newTestLogger())
	pattern := Pattern{
		PatternID:   "p1",
		PatternType: PatternUserIntent,
		Description: "test pattern",
		Frequency:   1,
		Confidence:  0.8,
		Metadata:    map[string]interface{}{},
		FirstSeen:   time.Now().Add(-time.Hour),
		LastSeen:    time.Now().Add(-time.Hour),
	}
	store.UpdatePattern(pattern)

	beforeUpdate := time.Now()
	update := Pattern{
		PatternID:   "p1",
		PatternType: PatternUserIntent,
		Confidence:  0.6,
	}
	store.UpdatePattern(update)

	stored := store.GetPattern("p1")
	assert.NotNil(t, stored)
	assert.Equal(t, 2, stored.Frequency)
	assert.True(t, stored.LastSeen.After(beforeUpdate) || stored.LastSeen.Equal(beforeUpdate))
	// Average of 0.8 and 0.6 = 0.7
	assert.InDelta(t, 0.7, stored.Confidence, 0.001)
}

func TestInsightStore_GetPattern(t *testing.T) {
	store := NewInsightStore(newTestLogger())
	pattern := Pattern{PatternID: "p1", Description: "found"}
	store.UpdatePattern(pattern)

	result := store.GetPattern("p1")
	assert.NotNil(t, result)
	assert.Equal(t, "found", result.Description)
}

func TestInsightStore_GetPattern_NotFound(t *testing.T) {
	store := NewInsightStore(newTestLogger())
	result := store.GetPattern("nonexistent")
	assert.Nil(t, result)
}

func TestInsightStore_GetPatternsByType(t *testing.T) {
	store := NewInsightStore(newTestLogger())
	store.UpdatePattern(Pattern{PatternID: "p1", PatternType: PatternUserIntent})
	store.UpdatePattern(Pattern{PatternID: "p2", PatternType: PatternDebateStrategy})
	store.UpdatePattern(Pattern{PatternID: "p3", PatternType: PatternUserIntent})

	results := store.GetPatternsByType(PatternUserIntent)
	assert.Len(t, results, 2)
	for _, p := range results {
		assert.Equal(t, PatternUserIntent, p.PatternType)
	}
}

func TestInsightStore_AddInsight(t *testing.T) {
	store := NewInsightStore(newTestLogger())
	insight := Insight{
		InsightID:   "i1",
		UserID:      "user-1",
		InsightType: "user_intent",
		Title:       "Test insight",
		Confidence:  0.9,
		CreatedAt:   time.Now(),
	}

	store.AddInsight(insight)

	result := store.GetInsight("i1")
	assert.NotNil(t, result)
	assert.Equal(t, "Test insight", result.Title)
}

func TestInsightStore_GetInsight(t *testing.T) {
	store := NewInsightStore(newTestLogger())
	store.AddInsight(Insight{InsightID: "i1", Title: "first"})
	store.AddInsight(Insight{InsightID: "i2", Title: "second"})

	result := store.GetInsight("i2")
	assert.NotNil(t, result)
	assert.Equal(t, "second", result.Title)
}

func TestInsightStore_GetInsightsByUser(t *testing.T) {
	store := NewInsightStore(newTestLogger())
	store.AddInsight(Insight{InsightID: "i1", UserID: "user-1"})
	store.AddInsight(Insight{InsightID: "i2", UserID: "user-2"})
	store.AddInsight(Insight{InsightID: "i3", UserID: "user-1"})

	results := store.GetInsightsByUser("user-1")
	assert.Len(t, results, 2)
	for _, ins := range results {
		assert.Equal(t, "user-1", ins.UserID)
	}
}

func TestInsightStore_GetTopPatterns(t *testing.T) {
	store := NewInsightStore(newTestLogger())
	store.UpdatePattern(Pattern{PatternID: "p1", Frequency: 5})
	store.UpdatePattern(Pattern{PatternID: "p2", Frequency: 10})
	store.UpdatePattern(Pattern{PatternID: "p3", Frequency: 3})
	store.UpdatePattern(Pattern{PatternID: "p4", Frequency: 8})

	results := store.GetTopPatterns(2)
	assert.Len(t, results, 2)
	assert.Equal(t, 10, results[0].Frequency)
	assert.Equal(t, 8, results[1].Frequency)
}

func TestInsightStore_GetStats(t *testing.T) {
	store := NewInsightStore(newTestLogger())
	store.UpdatePattern(Pattern{PatternID: "p1", PatternType: PatternUserIntent})
	store.UpdatePattern(Pattern{PatternID: "p2", PatternType: PatternUserIntent})
	store.UpdatePattern(Pattern{PatternID: "p3", PatternType: PatternDebateStrategy})
	store.AddInsight(Insight{InsightID: "i1"})

	stats := store.GetStats()
	assert.Equal(t, 3, stats["total_patterns"])
	assert.Equal(t, 1, stats["total_insights"])

	byType, ok := stats["patterns_by_type"].(map[string]int)
	assert.True(t, ok)
	assert.Equal(t, 2, byType[string(PatternUserIntent)])
	assert.Equal(t, 1, byType[string(PatternDebateStrategy)])
}

// ---------------------------------------------------------------------------
// CrossSessionLearner tests
// ---------------------------------------------------------------------------

func TestNewCrossSessionLearner(t *testing.T) {
	broker := &mockBroker{}
	logger := newTestLogger()
	config := newTestConfig()

	learner := NewCrossSessionLearner(config, broker, logger)

	assert.NotNil(t, learner)
	assert.Equal(t, broker, learner.broker)
	assert.Equal(t, logger, learner.logger)
	assert.Equal(t, "conversations.completed", learner.completedTopic)
	assert.Equal(t, "learning.insights", learner.insightsTopic)
	assert.NotNil(t, learner.insights)
}

func TestNewCrossSessionLearner_NilLogger(t *testing.T) {
	broker := &mockBroker{}
	config := newTestConfig()

	learner := NewCrossSessionLearner(config, broker, nil)

	assert.NotNil(t, learner)
	assert.NotNil(t, learner.logger)
}

func TestCrossSessionLearner_ExtractIntentPattern_HelpSeeking(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		Messages: []Message{
			{Role: "user", Content: "Can you help me with this?"},
		},
	}

	pattern := learner.extractIntentPattern(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, PatternUserIntent, pattern.PatternType)
	assert.Equal(t, "help_seeking", pattern.Metadata["intent"])
}

func TestCrossSessionLearner_ExtractIntentPattern_ExplanationRequest(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		Messages: []Message{
			{Role: "user", Content: "Can you explain this concept?"},
		},
	}

	pattern := learner.extractIntentPattern(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, "explanation_request", pattern.Metadata["intent"])
}

func TestCrossSessionLearner_ExtractIntentPattern_ProblemSolving(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		Messages: []Message{
			{Role: "user", Content: "I need to fix this bug"},
		},
	}

	pattern := learner.extractIntentPattern(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, "problem_solving", pattern.Metadata["intent"])
}

func TestCrossSessionLearner_ExtractIntentPattern_CreationTask(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		Messages: []Message{
			{Role: "user", Content: "Please create a new module"},
		},
	}

	pattern := learner.extractIntentPattern(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, "creation_task", pattern.Metadata["intent"])
}

func TestCrossSessionLearner_ExtractIntentPattern_GeneralInquiry(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		Messages: []Message{
			{Role: "user", Content: "List the logs from yesterday"},
		},
	}

	pattern := learner.extractIntentPattern(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, "general_inquiry", pattern.Metadata["intent"])
}

func TestCrossSessionLearner_ExtractIntentPattern_NoUserMessages(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		Messages: []Message{
			{Role: "assistant", Content: "Hello!"},
			{Role: "system", Content: "System prompt"},
		},
	}

	pattern := learner.extractIntentPattern(completion)

	assert.Nil(t, pattern)
}

func TestCrossSessionLearner_ExtractDebateStrategy(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		DebateRounds: []DebateRound{
			{Round: 1, Position: "advocate", Provider: "claude", Model: "claude-3", Confidence: 0.9, ResponseTimeMs: 500},
			{Round: 1, Position: "critic", Provider: "gemini", Model: "gemini-pro", Confidence: 0.7, ResponseTimeMs: 400},
			{Round: 2, Position: "advocate", Provider: "claude", Model: "claude-3", Confidence: 0.95, ResponseTimeMs: 600},
		},
	}

	pattern := learner.extractDebateStrategy(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, PatternDebateStrategy, pattern.PatternType)
	assert.Equal(t, "claude", pattern.Metadata["best_provider"])
	assert.Equal(t, 3, pattern.Metadata["total_rounds"])
}

func TestCrossSessionLearner_ExtractDebateStrategy_NoRounds(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		DebateRounds: []DebateRound{},
	}

	pattern := learner.extractDebateStrategy(completion)

	assert.Nil(t, pattern)
}

func TestCrossSessionLearner_ExtractEntityCooccurrence(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		Entities: []Entity{
			{EntityID: "e1", Type: "language", Name: "Go"},
			{EntityID: "e2", Type: "framework", Name: "Gin"},
			{EntityID: "e3", Type: "language", Name: "Python"},
		},
	}

	pattern := learner.extractEntityCooccurrence(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, PatternEntityCooccurrence, pattern.PatternType)
	assert.NotEmpty(t, pattern.Metadata["entity_pair"])
}

func TestCrossSessionLearner_ExtractEntityCooccurrence_SingleEntity(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		Entities: []Entity{
			{EntityID: "e1", Type: "language", Name: "Go"},
		},
	}

	pattern := learner.extractEntityCooccurrence(completion)

	assert.Nil(t, pattern)
}

func TestCrossSessionLearner_ExtractUserPreference_Concise(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		Messages: []Message{
			{Role: "user", Content: "Fix bug", Tokens: 20},
			{Role: "user", Content: "Thanks", Tokens: 10},
		},
	}

	pattern := learner.extractUserPreference(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, PatternUserPreference, pattern.PatternType)
	assert.Equal(t, "concise", pattern.Metadata["style"])
}

func TestCrossSessionLearner_ExtractUserPreference_Moderate(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		Messages: []Message{
			{Role: "user", Content: "moderate length message", Tokens: 100},
			{Role: "user", Content: "another moderate message", Tokens: 80},
		},
	}

	pattern := learner.extractUserPreference(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, "moderate", pattern.Metadata["style"])
}

func TestCrossSessionLearner_ExtractUserPreference_Detailed(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		Messages: []Message{
			{Role: "user", Content: "very long detailed message", Tokens: 300},
			{Role: "user", Content: "another very long message", Tokens: 250},
		},
	}

	pattern := learner.extractUserPreference(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, "detailed", pattern.Metadata["style"])
}

func TestCrossSessionLearner_ExtractUserPreference_NoUserMessages(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		Messages: []Message{
			{Role: "assistant", Content: "Hello there"},
		},
	}

	pattern := learner.extractUserPreference(completion)

	assert.Nil(t, pattern)
}

func TestCrossSessionLearner_ExtractConversationFlow_Rapid(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	now := time.Now()
	// 3 messages in 6 seconds = 2s per message (< 5s threshold)
	completion := ConversationCompletion{
		StartedAt:   now,
		CompletedAt: now.Add(6 * time.Second),
		Messages: []Message{
			{Role: "user", Content: "a", Timestamp: now},
			{Role: "assistant", Content: "b", Timestamp: now.Add(2 * time.Second)},
			{Role: "user", Content: "c", Timestamp: now.Add(4 * time.Second)},
		},
	}

	pattern := learner.extractConversationFlow(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, PatternConversationFlow, pattern.PatternType)
	assert.Equal(t, "rapid", pattern.Metadata["flow_type"])
}

func TestCrossSessionLearner_ExtractConversationFlow_Normal(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	now := time.Now()
	// 3 messages in 45 seconds = 15s per message (5s-30s threshold)
	completion := ConversationCompletion{
		StartedAt:   now,
		CompletedAt: now.Add(45 * time.Second),
		Messages: []Message{
			{Role: "user", Content: "a", Timestamp: now},
			{Role: "assistant", Content: "b", Timestamp: now.Add(15 * time.Second)},
			{Role: "user", Content: "c", Timestamp: now.Add(30 * time.Second)},
		},
	}

	pattern := learner.extractConversationFlow(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, "normal", pattern.Metadata["flow_type"])
}

func TestCrossSessionLearner_ExtractConversationFlow_Thoughtful(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	now := time.Now()
	// 3 messages in 3 minutes = 60s per message (> 30s threshold)
	completion := ConversationCompletion{
		StartedAt:   now,
		CompletedAt: now.Add(3 * time.Minute),
		Messages: []Message{
			{Role: "user", Content: "a", Timestamp: now},
			{Role: "assistant", Content: "b", Timestamp: now.Add(1 * time.Minute)},
			{Role: "user", Content: "c", Timestamp: now.Add(2 * time.Minute)},
		},
	}

	pattern := learner.extractConversationFlow(completion)

	assert.NotNil(t, pattern)
	assert.Equal(t, "thoughtful", pattern.Metadata["flow_type"])
}

func TestCrossSessionLearner_ExtractConversationFlow_TooFewMessages(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	now := time.Now()
	completion := ConversationCompletion{
		StartedAt:   now,
		CompletedAt: now.Add(time.Minute),
		Messages: []Message{
			{Role: "user", Content: "a", Timestamp: now},
			{Role: "assistant", Content: "b", Timestamp: now.Add(30 * time.Second)},
		},
	}

	pattern := learner.extractConversationFlow(completion)

	assert.Nil(t, pattern)
}

func TestCrossSessionLearner_GenerateInsights_HighFrequency(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{
		ConversationID: "conv-1",
		UserID:         "user-1",
	}
	patterns := []Pattern{
		{
			PatternID:   "p1",
			PatternType: PatternUserIntent,
			Description: "high frequency pattern",
			Frequency:   5,
			Confidence:  0.85,
		},
	}

	insights := learner.generateInsights(completion, patterns)

	assert.Len(t, insights, 1)
	assert.Equal(t, "user-1", insights[0].UserID)
	assert.Equal(t, string(PatternUserIntent), insights[0].InsightType)
	assert.Contains(t, insights[0].Title, "high frequency pattern")
	assert.NotEmpty(t, insights[0].Impact)
}

func TestCrossSessionLearner_GenerateInsights_LowFrequency(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	completion := ConversationCompletion{UserID: "user-1"}
	patterns := []Pattern{
		{PatternID: "p1", Frequency: 1, Confidence: 0.8},
		{PatternID: "p2", Frequency: 2, Confidence: 0.9},
	}

	insights := learner.generateInsights(completion, patterns)

	assert.Empty(t, insights)
}

func TestCrossSessionLearner_DetermineImpact_High(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	pattern := Pattern{Frequency: 10, Confidence: 0.95}

	impact := learner.determineImpact(pattern)

	assert.Equal(t, "high", impact)
}

func TestCrossSessionLearner_DetermineImpact_Medium(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	pattern := Pattern{Frequency: 5, Confidence: 0.75}

	impact := learner.determineImpact(pattern)

	assert.Equal(t, "medium", impact)
}

func TestCrossSessionLearner_DetermineImpact_Low(t *testing.T) {
	learner := NewCrossSessionLearner(newTestConfig(), &mockBroker{}, newTestLogger())
	pattern := Pattern{Frequency: 2, Confidence: 0.5}

	impact := learner.determineImpact(pattern)

	assert.Equal(t, "low", impact)
}

func TestCrossSessionLearner_StartLearning_Success(t *testing.T) {
	broker := &mockBroker{}
	learner := NewCrossSessionLearner(newTestConfig(), broker, newTestLogger())

	err := learner.StartLearning(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, broker.subscribeHandler)
	assert.Equal(t, "conversations.completed", broker.subscribeTopic)
}

func TestCrossSessionLearner_StartLearning_SubscribeError(t *testing.T) {
	broker := &mockBroker{
		subscribeErr: errors.New("subscribe failed"),
	}
	learner := NewCrossSessionLearner(newTestConfig(), broker, newTestLogger())

	err := learner.StartLearning(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to subscribe to completed topic")
}

func TestCrossSessionLearner_ProcessCompletion(t *testing.T) {
	broker := &mockBroker{}
	learner := NewCrossSessionLearner(newTestConfig(), broker, newTestLogger())
	completion := newTestCompletion()

	payload, err := json.Marshal(completion)
	assert.NoError(t, err)

	msg := &messaging.Message{
		ID:        "msg-1",
		Type:      "conversation.completed",
		Payload:   payload,
		Timestamp: time.Now(),
	}

	err = learner.processCompletion(context.Background(), msg)

	assert.NoError(t, err)
	// Patterns should have been extracted and stored
	allPatterns := learner.GetPatterns("all")
	assert.NotEmpty(t, allPatterns)
}

func TestCrossSessionLearner_GetInsights(t *testing.T) {
	broker := &mockBroker{}
	learner := NewCrossSessionLearner(newTestConfig(), broker, newTestLogger())

	// Add insights directly to the store
	now := time.Now()
	learner.insights.AddInsight(Insight{
		InsightID: "i1",
		Title:     "First",
		CreatedAt: now.Add(-2 * time.Hour),
	})
	learner.insights.AddInsight(Insight{
		InsightID: "i2",
		Title:     "Second",
		CreatedAt: now.Add(-1 * time.Hour),
	})
	learner.insights.AddInsight(Insight{
		InsightID: "i3",
		Title:     "Third",
		CreatedAt: now,
	})

	results := learner.GetInsights(2)
	assert.Len(t, results, 2)
	// Most recent first
	assert.Equal(t, "Third", results[0].Title)
	assert.Equal(t, "Second", results[1].Title)
}

func TestCrossSessionLearner_GetPatterns_All(t *testing.T) {
	broker := &mockBroker{}
	learner := NewCrossSessionLearner(newTestConfig(), broker, newTestLogger())

	learner.insights.UpdatePattern(Pattern{PatternID: "p1", PatternType: PatternUserIntent})
	learner.insights.UpdatePattern(Pattern{PatternID: "p2", PatternType: PatternDebateStrategy})

	results := learner.GetPatterns("all")
	assert.Len(t, results, 2)
}

func TestCrossSessionLearner_GetPatterns_ByType(t *testing.T) {
	broker := &mockBroker{}
	learner := NewCrossSessionLearner(newTestConfig(), broker, newTestLogger())

	learner.insights.UpdatePattern(Pattern{PatternID: "p1", PatternType: PatternUserIntent})
	learner.insights.UpdatePattern(Pattern{PatternID: "p2", PatternType: PatternDebateStrategy})
	learner.insights.UpdatePattern(Pattern{PatternID: "p3", PatternType: PatternUserIntent})

	results := learner.GetPatterns(string(PatternUserIntent))
	assert.Len(t, results, 2)
	for _, p := range results {
		assert.Equal(t, PatternUserIntent, p.PatternType)
	}
}

func TestCrossSessionLearner_GetLearningStats(t *testing.T) {
	broker := &mockBroker{}
	learner := NewCrossSessionLearner(newTestConfig(), broker, newTestLogger())

	learner.insights.UpdatePattern(Pattern{PatternID: "p1", PatternType: PatternUserIntent})
	learner.insights.UpdatePattern(Pattern{PatternID: "p2", PatternType: PatternDebateStrategy})
	learner.insights.AddInsight(Insight{InsightID: "i1"})

	stats := learner.GetLearningStats()

	assert.Equal(t, 2, stats["total_patterns"])
	assert.Equal(t, 1, stats["total_insights"])
	byType, ok := stats["patterns_by_type"].(map[string]int)
	assert.True(t, ok)
	assert.Equal(t, 1, byType[string(PatternUserIntent)])
	assert.Equal(t, 1, byType[string(PatternDebateStrategy)])
}

func TestCrossSessionLearner_PublishInsight_Success(t *testing.T) {
	broker := &mockBroker{}
	learner := NewCrossSessionLearner(newTestConfig(), broker, newTestLogger())

	insight := Insight{
		InsightID:   "i1",
		InsightType: "user_intent",
		Title:       "Test insight",
		Confidence:  0.9,
		CreatedAt:   time.Now(),
	}

	err := learner.publishInsight(context.Background(), insight)

	assert.NoError(t, err)
	assert.Len(t, broker.publishedMessages, 1)
	assert.Equal(t, "learning.insight", broker.publishedMessages[0].Type)

	// Verify the payload contains the insight
	var published Insight
	err = json.Unmarshal(broker.publishedMessages[0].Payload, &published)
	assert.NoError(t, err)
	assert.Equal(t, "i1", published.InsightID)
}

func TestCrossSessionLearner_PublishInsight_Error(t *testing.T) {
	broker := &mockBroker{
		publishErr: errors.New("publish failed"),
	}
	learner := NewCrossSessionLearner(newTestConfig(), broker, newTestLogger())

	insight := Insight{
		InsightID: "i1",
		CreatedAt: time.Now(),
	}

	err := learner.publishInsight(context.Background(), insight)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish insight")
}
