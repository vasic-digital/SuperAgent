package bigdata

import (
	"context"
	"time"

	"dev.helix.agent/internal/services"
	"github.com/sirupsen/logrus"
)

// DebateServiceWrapper wraps the debate service to add big data capabilities
type DebateServiceWrapper struct {
	debateService     *services.DebateService
	debateIntegration *DebateIntegration
	analyticsIntegration *AnalyticsIntegration
	entityIntegration *EntityIntegration
	logger            *logrus.Logger
	enableBigData     bool
}

// NewDebateServiceWrapper creates a new debate service wrapper
func NewDebateServiceWrapper(
	debateService *services.DebateService,
	debateIntegration *DebateIntegration,
	analyticsIntegration *AnalyticsIntegration,
	entityIntegration *EntityIntegration,
	logger *logrus.Logger,
	enableBigData bool,
) *DebateServiceWrapper {
	return &DebateServiceWrapper{
		debateService:        debateService,
		debateIntegration:    debateIntegration,
		analyticsIntegration: analyticsIntegration,
		entityIntegration:    entityIntegration,
		logger:               logger,
		enableBigData:        enableBigData,
	}
}

// RunDebate runs a debate with big data integration
func (dsw *DebateServiceWrapper) RunDebate(ctx context.Context, config *services.DebateConfig) (*services.DebateResult, error) {
	startTime := time.Now()

	// Get unlimited context if conversation ID is provided
	if dsw.enableBigData && config.ConversationID != "" {
		conversationCtx, err := dsw.debateIntegration.GetConversationContext(
			ctx,
			config.ConversationID,
			4000, // Default max tokens
		)
		if err != nil {
			dsw.logger.WithError(err).Warn("Failed to get conversation context, continuing without context")
		} else {
			// Inject context into debate configuration
			config.Context = conversationCtx
			dsw.logger.WithFields(logrus.Fields{
				"conversation_id": config.ConversationID,
				"message_count":   len(conversationCtx.Messages),
				"entity_count":    len(conversationCtx.Entities),
				"compressed":      conversationCtx.Compressed,
			}).Info("Loaded conversation context from Kafka")
		}
	}

	// Run the actual debate
	result, err := dsw.debateService.RunDebate(ctx, config)
	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)

	// Publish completion event to big data pipeline
	if dsw.enableBigData {
		go dsw.publishDebateCompletion(context.Background(), config, result, duration)
	}

	return result, nil
}

// publishDebateCompletion publishes debate completion events to big data systems
func (dsw *DebateServiceWrapper) publishDebateCompletion(ctx context.Context, config *services.DebateConfig, result *services.DebateResult, duration time.Duration) {
	// Publish to Kafka for conversation event log
	completion := &DebateCompletion{
		DebateID:       result.DebateID,
		ConversationID: config.ConversationID,
		UserID:         config.UserID,
		SessionID:      config.SessionID,
		Topic:          config.Topic,
		Rounds:         result.TotalRounds,
		Winner:         result.Winner,
		WinnerProvider: dsw.extractProviderFromWinner(result.Winner),
		WinnerModel:    dsw.extractModelFromWinner(result.Winner),
		Confidence:     result.Confidence,
		Duration:       duration,
		StartedAt:      result.StartTime,
		CompletedAt:    result.EndTime,
		Participants:   dsw.convertParticipants(result.Participants),
		Outcome:        dsw.determineOutcome(result),
	}

	if err := dsw.debateIntegration.PublishDebateCompletion(ctx, completion); err != nil {
		dsw.logger.WithError(err).Error("Failed to publish debate completion")
	}

	// Publish debate metrics to analytics
	if dsw.analyticsIntegration != nil {
		totalTokens := dsw.calculateTotalTokens(result.Participants)
		if err := dsw.analyticsIntegration.RecordDebateCompletion(
			ctx,
			result.DebateID,
			config.Topic,
			result.TotalRounds,
			duration,
			len(result.Participants),
			result.Winner,
			completion.WinnerProvider,
			completion.WinnerModel,
			result.Confidence,
			totalTokens,
			completion.Outcome,
		); err != nil {
			dsw.logger.WithError(err).Error("Failed to record debate metrics")
		}
	}

	// Publish entities extracted during debate
	if dsw.entityIntegration != nil && result.Entities != nil {
		if err := dsw.entityIntegration.PublishEntitiesBatch(ctx, result.Entities, config.ConversationID); err != nil {
			dsw.logger.WithError(err).Error("Failed to publish debate entities")
		}
	}
}

// Helper functions

func (dsw *DebateServiceWrapper) extractProviderFromWinner(winner string) string {
	// Winner format: "provider/model" or just "provider"
	// Extract provider name before the slash
	for i, char := range winner {
		if char == '/' {
			return winner[:i]
		}
	}
	return winner
}

func (dsw *DebateServiceWrapper) extractModelFromWinner(winner string) string {
	// Extract model name after the slash
	for i, char := range winner {
		if char == '/' {
			if i+1 < len(winner) {
				return winner[i+1:]
			}
			return ""
		}
	}
	return ""
}

func (dsw *DebateServiceWrapper) convertParticipants(participants []services.DebateParticipant) []DebateParticipant {
	result := make([]DebateParticipant, len(participants))
	for i, p := range participants {
		result[i] = DebateParticipant{
			Provider:     dsw.extractProviderFromWinner(p.Provider),
			Model:        dsw.extractModelFromWinner(p.Provider),
			Position:     p.Position,
			ResponseTime: int(p.ResponseTime.Milliseconds()),
			TokensUsed:   p.TokensUsed,
			Confidence:   p.Confidence,
			Won:          p.Won,
		}
	}
	return result
}

func (dsw *DebateServiceWrapper) determineOutcome(result *services.DebateResult) string {
	if result.Error != nil {
		return "error"
	}
	if result.Abandoned {
		return "abandoned"
	}
	return "successful"
}

func (dsw *DebateServiceWrapper) calculateTotalTokens(participants []services.DebateParticipant) int {
	total := 0
	for _, p := range participants {
		total += p.TokensUsed
	}
	return total
}

// RecordProviderCall records a provider API call for analytics
func (dsw *DebateServiceWrapper) RecordProviderCall(
	ctx context.Context,
	provider, model, requestID string,
	responseTime time.Duration,
	tokensUsed int,
	success bool,
	errorType string,
) {
	if !dsw.enableBigData || dsw.analyticsIntegration == nil {
		return
	}

	if err := dsw.analyticsIntegration.RecordProviderRequest(
		ctx,
		provider,
		model,
		requestID,
		responseTime,
		tokensUsed,
		success,
		errorType,
	); err != nil {
		dsw.logger.WithError(err).Debug("Failed to record provider call")
	}
}

// RecordDebateRound records a single debate round for analytics
func (dsw *DebateServiceWrapper) RecordDebateRound(
	ctx context.Context,
	debateID, provider, model string,
	round int,
	responseTime time.Duration,
	tokensUsed int,
	confidence float64,
) {
	if !dsw.enableBigData || dsw.analyticsIntegration == nil {
		return
	}

	if err := dsw.analyticsIntegration.RecordDebateRound(
		ctx,
		debateID,
		provider,
		model,
		round,
		responseTime,
		tokensUsed,
		confidence,
	); err != nil {
		dsw.logger.WithError(err).Debug("Failed to record debate round")
	}
}
