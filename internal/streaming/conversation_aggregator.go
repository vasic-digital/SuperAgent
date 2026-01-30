package streaming

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ConversationAggregator aggregates conversation events into state
type ConversationAggregator struct {
	stateStore StateStore
	logger     *zap.Logger
}

// NewConversationAggregator creates a new conversation aggregator
func NewConversationAggregator(stateStore StateStore, logger *zap.Logger) *ConversationAggregator {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &ConversationAggregator{
		stateStore: stateStore,
		logger:     logger,
	}
}

// AddMessage aggregates a message into conversation state
func (ca *ConversationAggregator) AddMessage(ctx context.Context, conversationID string, message *MessageData) (*ConversationState, error) {
	// Get existing state or create new
	state, err := ca.stateStore.GetState(ctx, conversationID)
	if err != nil {
		// Create new state if doesn't exist
		state = &ConversationState{
			ConversationID: conversationID,
			Entities:       make(map[string]EntityData),
			ProviderUsage:  make(map[string]int),
			StartedAt:      time.Now(),
			LastUpdatedAt:  time.Now(),
			Version:        0,
		}

		ca.logger.Debug("Created new conversation state",
			zap.String("conversation_id", conversationID))
	}

	// Update state with message
	state.MessageCount++
	state.TotalTokens += int64(message.Tokens)
	state.LastUpdatedAt = time.Now()

	ca.logger.Debug("Aggregated message",
		zap.String("conversation_id", conversationID),
		zap.String("message_id", message.MessageID),
		zap.Int("message_count", state.MessageCount),
		zap.Int64("total_tokens", state.TotalTokens))

	return state, nil
}

// AddEntity adds an entity to conversation state
func (ca *ConversationAggregator) AddEntity(ctx context.Context, conversationID string, entity EntityData) (*ConversationState, error) {
	state, err := ca.stateStore.GetState(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation state: %w", err)
	}

	// Add or update entity
	state.Entities[entity.EntityID] = entity
	state.EntityCount = len(state.Entities)
	state.LastUpdatedAt = time.Now()

	ca.logger.Debug("Added entity to conversation",
		zap.String("conversation_id", conversationID),
		zap.String("entity_id", entity.EntityID),
		zap.String("entity_name", entity.Name),
		zap.Int("total_entities", state.EntityCount))

	return state, nil
}

// UpdateProviderUsage updates provider usage statistics
func (ca *ConversationAggregator) UpdateProviderUsage(ctx context.Context, conversationID string, provider string) (*ConversationState, error) {
	state, err := ca.stateStore.GetState(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation state: %w", err)
	}

	state.ProviderUsage[provider]++
	state.LastUpdatedAt = time.Now()

	ca.logger.Debug("Updated provider usage",
		zap.String("conversation_id", conversationID),
		zap.String("provider", provider),
		zap.Int("usage_count", state.ProviderUsage[provider]))

	return state, nil
}

// GetState retrieves current conversation state
func (ca *ConversationAggregator) GetState(ctx context.Context, conversationID string) (*ConversationState, error) {
	return ca.stateStore.GetState(ctx, conversationID)
}

// AggregateWindow aggregates state into a windowed analytics snapshot
func (ca *ConversationAggregator) AggregateWindow(ctx context.Context, conversationID string, windowStart, windowEnd time.Time) (*WindowedAnalytics, error) {
	state, err := ca.stateStore.GetState(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation state: %w", err)
	}

	// Calculate knowledge density
	var knowledgeDensity float64
	if state.MessageCount > 0 {
		knowledgeDensity = float64(state.EntityCount) / float64(state.MessageCount)
	}

	// Calculate average response time (simplified - would need more data in practice)
	avgResponseTime := float64(0)
	if state.DebateRoundCount > 0 {
		// This would typically be calculated from actual debate round data
		avgResponseTime = 500.0 // Placeholder
	}

	analytics := &WindowedAnalytics{
		WindowStart:          windowStart,
		WindowEnd:            windowEnd,
		ConversationID:       conversationID,
		TotalMessages:        state.MessageCount,
		LLMCalls:             state.DebateRoundCount,
		DebateRounds:         state.DebateRoundCount,
		AvgResponseTimeMs:    avgResponseTime,
		EntityGrowth:         state.EntityCount,
		KnowledgeDensity:     knowledgeDensity,
		ProviderDistribution: make(map[string]int),
		CreatedAt:            time.Now(),
	}

	// Copy provider distribution
	for provider, count := range state.ProviderUsage {
		analytics.ProviderDistribution[provider] = count
	}

	ca.logger.Debug("Created windowed analytics",
		zap.String("conversation_id", conversationID),
		zap.Time("window_start", windowStart),
		zap.Time("window_end", windowEnd),
		zap.Int("total_messages", analytics.TotalMessages),
		zap.Float64("knowledge_density", analytics.KnowledgeDensity))

	return analytics, nil
}

// MergeStates merges multiple conversation states (for multi-partition aggregation)
func (ca *ConversationAggregator) MergeStates(states ...*ConversationState) *ConversationState {
	if len(states) == 0 {
		return nil
	}

	if len(states) == 1 {
		return states[0]
	}

	merged := &ConversationState{
		ConversationID: states[0].ConversationID,
		UserID:         states[0].UserID,
		SessionID:      states[0].SessionID,
		Entities:       make(map[string]EntityData),
		ProviderUsage:  make(map[string]int),
		StartedAt:      states[0].StartedAt,
		LastUpdatedAt:  time.Now(),
		Version:        0,
	}

	// Merge all states
	for _, state := range states {
		merged.MessageCount += state.MessageCount
		merged.TotalTokens += state.TotalTokens
		merged.DebateRoundCount += state.DebateRoundCount

		// Merge entities
		for id, entity := range state.Entities {
			merged.Entities[id] = entity
		}

		// Merge provider usage
		for provider, count := range state.ProviderUsage {
			merged.ProviderUsage[provider] += count
		}

		// Use latest timestamp
		if state.LastUpdatedAt.After(merged.LastUpdatedAt) {
			merged.LastUpdatedAt = state.LastUpdatedAt
		}

		// Use earliest start time
		if state.StartedAt.Before(merged.StartedAt) {
			merged.StartedAt = state.StartedAt
		}
	}

	merged.EntityCount = len(merged.Entities)

	ca.logger.Debug("Merged conversation states",
		zap.String("conversation_id", merged.ConversationID),
		zap.Int("num_states", len(states)),
		zap.Int("total_messages", merged.MessageCount),
		zap.Int("total_entities", merged.EntityCount))

	return merged
}

// CalculateKnowledgeDensity calculates knowledge density for a conversation
func (ca *ConversationAggregator) CalculateKnowledgeDensity(state *ConversationState) float64 {
	if state.MessageCount == 0 {
		return 0
	}
	return float64(state.EntityCount) / float64(state.MessageCount)
}

// CalculateProviderDistribution calculates provider usage percentage
func (ca *ConversationAggregator) CalculateProviderDistribution(state *ConversationState) map[string]float64 {
	total := 0
	for _, count := range state.ProviderUsage {
		total += count
	}

	if total == 0 {
		return make(map[string]float64)
	}

	distribution := make(map[string]float64)
	for provider, count := range state.ProviderUsage {
		distribution[provider] = float64(count) / float64(total) * 100.0
	}

	return distribution
}
