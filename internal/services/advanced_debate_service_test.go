package services

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/helixagent/helixagent/internal/models"
)

func newAdvancedDebateTestLogger() *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel) // Silence logs in tests
	return log
}

// createAdvancedDebateTestDebateService creates a DebateService with mock providers for testing
func createAdvancedDebateTestDebateService(logger *logrus.Logger) *DebateService {
	// Create mock providers
	mockProvider1 := newDebateMockProvider("openai", &models.LLMResponse{
		Content:      "This is my position on the topic. I present my arguments clearly and thoughtfully.",
		Confidence:   0.85,
		TokensUsed:   100,
		FinishReason: "stop",
	})

	mockProvider2 := newDebateMockProvider("anthropic", &models.LLMResponse{
		Content:      "I offer a different perspective and challenge the previous arguments with new insights.",
		Confidence:   0.90,
		TokensUsed:   120,
		FinishReason: "stop",
	})

	mockProvider3 := newDebateMockProvider("google", &models.LLMResponse{
		Content:      "As a mediator, I synthesize both viewpoints and propose a balanced resolution.",
		Confidence:   0.88,
		TokensUsed:   110,
		FinishReason: "stop",
	})

	// Create registry with mock providers
	registry := createTestProviderRegistry(map[string]*debateMockLLMProvider{
		"openai":    mockProvider1,
		"anthropic": mockProvider2,
		"google":    mockProvider3,
	})

	return NewDebateServiceWithDeps(logger, registry, nil)
}

func TestAdvancedDebateService_ConductAdvancedDebate(t *testing.T) {
	logger := newAdvancedDebateTestLogger()

	ads := NewAdvancedDebateService(
		createAdvancedDebateTestDebateService(logger),
		NewDebateMonitoringService(logger),
		NewDebatePerformanceService(logger),
		NewDebateHistoryService(logger),
		NewDebateResilienceService(logger),
		NewDebateReportingService(logger),
		NewDebateSecurityService(logger),
		logger,
	)

	config := &DebateConfig{
		DebateID:  "test-advanced-debate-1",
		Topic:     "Advanced Topic",
		MaxRounds: 3,
		Timeout:   10 * time.Second,
		Participants: []ParticipantConfig{
			{
				ParticipantID: "participant-1",
				Name:          "Agent 1",
				Role:          "proposer",
				LLMProvider:   "openai",
				LLMModel:      "gpt-4",
			},
		},
		EnableCognee: false,
	}

	result, err := ads.ConductAdvancedDebate(context.Background(), config)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "test-advanced-debate-1", result.DebateID)
	assert.Equal(t, "Advanced Topic", result.Topic)
	assert.True(t, result.Success)

	// Verify report was added to metadata
	require.NotNil(t, result.Metadata)
	assert.Contains(t, result.Metadata, "report")
}

func TestAdvancedDebateService_ConductAdvancedDebate_WithCogneeEnabled(t *testing.T) {
	logger := newAdvancedDebateTestLogger()

	ads := NewAdvancedDebateService(
		createAdvancedDebateTestDebateService(logger),
		NewDebateMonitoringService(logger),
		NewDebatePerformanceService(logger),
		NewDebateHistoryService(logger),
		NewDebateResilienceService(logger),
		NewDebateReportingService(logger),
		NewDebateSecurityService(logger),
		logger,
	)

	config := &DebateConfig{
		DebateID:  "test-advanced-debate-cognee",
		Topic:     "Cognee-Enhanced Debate",
		MaxRounds: 2,
		Timeout:   5 * time.Second,
		Participants: []ParticipantConfig{
			{ParticipantID: "p1", Name: "Agent", Role: "debater", LLMProvider: "openai"},
		},
		EnableCognee: true, // Enable Cognee
	}

	result, err := ads.ConductAdvancedDebate(context.Background(), config)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Cognee enhancement may not be available in test environment
	// Just verify the debate completes successfully with the flag set
	assert.Equal(t, "test-advanced-debate-cognee", result.DebateID)
}

func TestAdvancedDebateService_ConductAdvancedDebate_MultipleParticipants(t *testing.T) {
	logger := newAdvancedDebateTestLogger()

	ads := NewAdvancedDebateService(
		createAdvancedDebateTestDebateService(logger),
		NewDebateMonitoringService(logger),
		NewDebatePerformanceService(logger),
		NewDebateHistoryService(logger),
		NewDebateResilienceService(logger),
		NewDebateReportingService(logger),
		NewDebateSecurityService(logger),
		logger,
	)

	config := &DebateConfig{
		DebateID:  "test-advanced-debate-multi",
		Topic:     "Multi-participant Discussion",
		MaxRounds: 5,
		Timeout:   30 * time.Second,
		Participants: []ParticipantConfig{
			{ParticipantID: "p1", Name: "Alice", Role: "proposer", LLMProvider: "openai"},
			{ParticipantID: "p2", Name: "Bob", Role: "critic", LLMProvider: "anthropic"},
			{ParticipantID: "p3", Name: "Charlie", Role: "mediator", LLMProvider: "google"},
		},
		EnableCognee: false,
	}

	result, err := ads.ConductAdvancedDebate(context.Background(), config)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "test-advanced-debate-multi", result.DebateID)
	// The debate should complete successfully with multiple participants
	assert.True(t, result.Success)
}

func TestAdvancedDebateService_ConductAdvancedDebate_EmptyParticipants(t *testing.T) {
	logger := newAdvancedDebateTestLogger()

	ads := NewAdvancedDebateService(
		createAdvancedDebateTestDebateService(logger),
		NewDebateMonitoringService(logger),
		NewDebatePerformanceService(logger),
		NewDebateHistoryService(logger),
		NewDebateResilienceService(logger),
		NewDebateReportingService(logger),
		NewDebateSecurityService(logger),
		logger,
	)

	config := &DebateConfig{
		DebateID:     "test-advanced-debate-empty",
		Topic:        "Empty Debate",
		MaxRounds:    1,
		Timeout:      5 * time.Second,
		Participants: []ParticipantConfig{},
		EnableCognee: false,
	}

	result, err := ads.ConductAdvancedDebate(context.Background(), config)
	// Empty participants should trigger security validation error
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "security validation failed")
}
