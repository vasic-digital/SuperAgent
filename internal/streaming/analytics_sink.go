package streaming

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// AnalyticsSink writes analytics data to PostgreSQL
type AnalyticsSink struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewAnalyticsSink creates a new analytics sink
func NewAnalyticsSink(logger *zap.Logger) *AnalyticsSink {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &AnalyticsSink{
		logger: logger,
	}
}

// SetDB sets the database connection
func (as *AnalyticsSink) SetDB(db *sql.DB) {
	as.db = db
}

// WriteConversationState writes conversation state snapshot to database
func (as *AnalyticsSink) WriteConversationState(ctx context.Context, state *ConversationState) error {
	if as.db == nil {
		return fmt.Errorf("database connection not set")
	}

	// Marshal entities and context to JSON
	entitiesJSON, err := json.Marshal(state.Entities)
	if err != nil {
		return fmt.Errorf("failed to marshal entities: %w", err)
	}

	compressedContextJSON, err := json.Marshal(state.CompressedContext)
	if err != nil {
		return fmt.Errorf("failed to marshal compressed context: %w", err)
	}

	query := `
		INSERT INTO conversation_state_snapshots (
			conversation_id, user_id, session_id, message_count, entity_count,
			total_tokens, compressed_context, latest_entities, last_updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (conversation_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			session_id = EXCLUDED.session_id,
			message_count = EXCLUDED.message_count,
			entity_count = EXCLUDED.entity_count,
			total_tokens = EXCLUDED.total_tokens,
			compressed_context = EXCLUDED.compressed_context,
			latest_entities = EXCLUDED.latest_entities,
			last_updated_at = EXCLUDED.last_updated_at,
			version = EXCLUDED.version
	`

	_, err = as.db.ExecContext(ctx, query,
		state.ConversationID,
		state.UserID,
		state.SessionID,
		state.MessageCount,
		state.EntityCount,
		state.TotalTokens,
		compressedContextJSON,
		entitiesJSON,
		state.LastUpdatedAt,
		state.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to write conversation state: %w", err)
	}

	as.logger.Debug("Wrote conversation state snapshot",
		zap.String("conversation_id", state.ConversationID),
		zap.Int("message_count", state.MessageCount),
		zap.Int("entity_count", state.EntityCount))

	return nil
}

// WriteWindowedAnalytics writes windowed analytics to database
func (as *AnalyticsSink) WriteWindowedAnalytics(ctx context.Context, analytics *WindowedAnalytics) error {
	if as.db == nil {
		return fmt.Errorf("database connection not set")
	}

	// Marshal provider distribution to JSON
	providerDistJSON, err := json.Marshal(analytics.ProviderDistribution)
	if err != nil {
		return fmt.Errorf("failed to marshal provider distribution: %w", err)
	}

	query := `
		INSERT INTO conversation_analytics (
			id, conversation_id, window_start, window_end, total_messages,
			llm_calls, debate_rounds, avg_response_time_ms, entity_growth,
			knowledge_density, provider_distribution, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err = as.db.ExecContext(ctx, query,
		analytics.ConversationID,
		analytics.WindowStart,
		analytics.WindowEnd,
		analytics.TotalMessages,
		analytics.LLMCalls,
		analytics.DebateRounds,
		analytics.AvgResponseTimeMs,
		analytics.EntityGrowth,
		analytics.KnowledgeDensity,
		providerDistJSON,
		analytics.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to write windowed analytics: %w", err)
	}

	as.logger.Debug("Wrote windowed analytics",
		zap.String("conversation_id", analytics.ConversationID),
		zap.Time("window_start", analytics.WindowStart),
		zap.Time("window_end", analytics.WindowEnd),
		zap.Int("total_messages", analytics.TotalMessages))

	return nil
}

// WriteBatch writes multiple analytics records in a single transaction
func (as *AnalyticsSink) WriteBatch(ctx context.Context, analytics []*WindowedAnalytics) error {
	if as.db == nil {
		return fmt.Errorf("database connection not set")
	}

	if len(analytics) == 0 {
		return nil
	}

	// Start transaction
	tx, err := as.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO conversation_analytics (
			id, conversation_id, window_start, window_end, total_messages,
			llm_calls, debate_rounds, avg_response_time_ms, entity_growth,
			knowledge_density, provider_distribution, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, a := range analytics {
		providerDistJSON, err := json.Marshal(a.ProviderDistribution)
		if err != nil {
			return fmt.Errorf("failed to marshal provider distribution: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			a.ConversationID,
			a.WindowStart,
			a.WindowEnd,
			a.TotalMessages,
			a.LLMCalls,
			a.DebateRounds,
			a.AvgResponseTimeMs,
			a.EntityGrowth,
			a.KnowledgeDensity,
			providerDistJSON,
			a.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	as.logger.Debug("Wrote batch analytics",
		zap.Int("count", len(analytics)))

	return nil
}

// QueryAnalytics queries analytics for a conversation within a time range
func (as *AnalyticsSink) QueryAnalytics(ctx context.Context, conversationID string, start, end time.Time) ([]*WindowedAnalytics, error) {
	if as.db == nil {
		return nil, fmt.Errorf("database connection not set")
	}

	query := `
		SELECT
			conversation_id, window_start, window_end, total_messages,
			llm_calls, debate_rounds, avg_response_time_ms, entity_growth,
			knowledge_density, provider_distribution, created_at
		FROM conversation_analytics
		WHERE conversation_id = $1
		  AND window_start >= $2
		  AND window_end <= $3
		ORDER BY window_start ASC
	`

	rows, err := as.db.QueryContext(ctx, query, conversationID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query analytics: %w", err)
	}
	defer rows.Close()

	var results []*WindowedAnalytics

	for rows.Next() {
		var a WindowedAnalytics
		var providerDistJSON []byte

		err := rows.Scan(
			&a.ConversationID,
			&a.WindowStart,
			&a.WindowEnd,
			&a.TotalMessages,
			&a.LLMCalls,
			&a.DebateRounds,
			&a.AvgResponseTimeMs,
			&a.EntityGrowth,
			&a.KnowledgeDensity,
			&providerDistJSON,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Unmarshal provider distribution
		if err := json.Unmarshal(providerDistJSON, &a.ProviderDistribution); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider distribution: %w", err)
		}

		results = append(results, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	as.logger.Debug("Queried analytics",
		zap.String("conversation_id", conversationID),
		zap.Int("result_count", len(results)))

	return results, nil
}

// AggregateProviderMetrics aggregates provider metrics across all conversations
func (as *AnalyticsSink) AggregateProviderMetrics(ctx context.Context, start, end time.Time) (map[string]interface{}, error) {
	if as.db == nil {
		return nil, fmt.Errorf("database connection not set")
	}

	query := `
		SELECT
			jsonb_object_agg(provider, usage_count) as provider_metrics
		FROM (
			SELECT
				provider_key as provider,
				SUM((provider_value)::int) as usage_count
			FROM conversation_analytics,
				 jsonb_each_text(provider_distribution) as p(provider_key, provider_value)
			WHERE window_start >= $1 AND window_end <= $2
			GROUP BY provider_key
		) sub
	`

	var metricsJSON []byte
	err := as.db.QueryRowContext(ctx, query, start, end).Scan(&metricsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to aggregate provider metrics: %w", err)
	}

	var metrics map[string]interface{}
	if err := json.Unmarshal(metricsJSON, &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	as.logger.Debug("Aggregated provider metrics",
		zap.Int("provider_count", len(metrics)))

	return metrics, nil
}
