package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sirupsen/logrus"
)

// ClickHouseAnalytics manages time-series analytics in ClickHouse
type ClickHouseAnalytics struct {
	conn   *sql.DB
	logger *logrus.Logger
}

// DebateMetrics represents metrics for a single debate round
type DebateMetrics struct {
	DebateID       string
	Round          int
	Timestamp      time.Time
	Provider       string
	Model          string
	Position       string
	ResponseTimeMs float32
	TokensUsed     int
	ConfidenceScore float32
	ErrorCount     int
	WasWinner      bool
	Metadata       map[string]string
}

// ProviderStats represents aggregated provider statistics
type ProviderStats struct {
	Provider          string
	TotalRequests     int64
	AvgResponseTime   float32
	P95ResponseTime   float32
	P99ResponseTime   float32
	AvgConfidence     float32
	TotalTokens       int64
	AvgTokensPerReq   float32
	ErrorRate         float32
	WinRate           float32
	Period            string
}

// ConversationMetrics represents metrics for a conversation
type ConversationMetrics struct {
	ConversationID string
	UserID         string
	Timestamp      time.Time
	MessageCount   int
	EntityCount    int
	TotalTokens    int64
	DurationMs     int64
	DebateRounds   int
	LLMsUsed       []string
	Metadata       map[string]string
}

// ClickHouseConfig defines ClickHouse configuration
type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	TLS      bool
}

// NewClickHouseAnalytics creates a new ClickHouse analytics client
func NewClickHouseAnalytics(config ClickHouseConfig, logger *logrus.Logger) (*ClickHouseAnalytics, error) {
	if logger == nil {
		logger = logrus.New()
	}

	// Build connection string
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	if !config.TLS {
		dsn += "?secure=false"
	}

	// Connect to ClickHouse
	conn, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open ClickHouse connection: %w", err)
	}

	// Verify connectivity
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"host":     config.Host,
		"port":     config.Port,
		"database": config.Database,
	}).Info("ClickHouse analytics initialized")

	return &ClickHouseAnalytics{
		conn:   conn,
		logger: logger,
	}, nil
}

// StoreDebateMetrics stores metrics for a debate round
func (cha *ClickHouseAnalytics) StoreDebateMetrics(ctx context.Context, metrics DebateMetrics) error {
	query := `
		INSERT INTO debate_metrics (
			debate_id, round, timestamp, provider, model, position,
			response_time_ms, tokens_used, confidence_score, error_count, was_winner
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := cha.conn.ExecContext(ctx, query,
		metrics.DebateID,
		metrics.Round,
		metrics.Timestamp,
		metrics.Provider,
		metrics.Model,
		metrics.Position,
		metrics.ResponseTimeMs,
		metrics.TokensUsed,
		metrics.ConfidenceScore,
		metrics.ErrorCount,
		metrics.WasWinner,
	)

	if err != nil {
		return fmt.Errorf("failed to insert debate metrics: %w", err)
	}

	cha.logger.WithFields(logrus.Fields{
		"debate_id": metrics.DebateID,
		"round":     metrics.Round,
		"provider":  metrics.Provider,
	}).Debug("Debate metrics stored")

	return nil
}

// StoreDebateMetricsBatch stores multiple debate metrics in batch
func (cha *ClickHouseAnalytics) StoreDebateMetricsBatch(ctx context.Context, metricsList []DebateMetrics) error {
	if len(metricsList) == 0 {
		return nil
	}

	tx, err := cha.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO debate_metrics (
			debate_id, round, timestamp, provider, model, position,
			response_time_ms, tokens_used, confidence_score, error_count, was_winner
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, metrics := range metricsList {
		_, err := stmt.ExecContext(ctx,
			metrics.DebateID,
			metrics.Round,
			metrics.Timestamp,
			metrics.Provider,
			metrics.Model,
			metrics.Position,
			metrics.ResponseTimeMs,
			metrics.TokensUsed,
			metrics.ConfidenceScore,
			metrics.ErrorCount,
			metrics.WasWinner,
		)
		if err != nil {
			return fmt.Errorf("failed to insert batch item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	cha.logger.WithField("count", len(metricsList)).Debug("Batch debate metrics stored")

	return nil
}

// GetProviderPerformance retrieves aggregated provider performance statistics
func (cha *ClickHouseAnalytics) GetProviderPerformance(ctx context.Context, window time.Duration) ([]ProviderStats, error) {
	query := `
		SELECT
			provider,
			COUNT(*) as total_requests,
			AVG(response_time_ms) as avg_response_time,
			quantile(0.95)(response_time_ms) as p95_response_time,
			quantile(0.99)(response_time_ms) as p99_response_time,
			AVG(confidence_score) as avg_confidence,
			SUM(tokens_used) as total_tokens,
			AVG(tokens_used) as avg_tokens_per_req,
			AVG(error_count) as error_rate,
			AVG(CAST(was_winner AS Float32)) as win_rate
		FROM debate_metrics
		WHERE timestamp >= now() - INTERVAL ? SECOND
		GROUP BY provider
		ORDER BY avg_confidence DESC
	`

	rows, err := cha.conn.QueryContext(ctx, query, int64(window.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("failed to query provider performance: %w", err)
	}
	defer rows.Close()

	var stats []ProviderStats
	for rows.Next() {
		var s ProviderStats
		err := rows.Scan(
			&s.Provider,
			&s.TotalRequests,
			&s.AvgResponseTime,
			&s.P95ResponseTime,
			&s.P99ResponseTime,
			&s.AvgConfidence,
			&s.TotalTokens,
			&s.AvgTokensPerReq,
			&s.ErrorRate,
			&s.WinRate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		s.Period = fmt.Sprintf("last_%s", window.String())
		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return stats, nil
}

// GetProviderTrends retrieves provider performance trends over time
func (cha *ClickHouseAnalytics) GetProviderTrends(ctx context.Context, provider string, interval string, periods int) ([]ProviderStats, error) {
	query := fmt.Sprintf(`
		SELECT
			provider,
			toStartOfInterval(timestamp, INTERVAL 1 %s) as period_start,
			COUNT(*) as total_requests,
			AVG(response_time_ms) as avg_response_time,
			quantile(0.95)(response_time_ms) as p95_response_time,
			quantile(0.99)(response_time_ms) as p99_response_time,
			AVG(confidence_score) as avg_confidence,
			SUM(tokens_used) as total_tokens,
			AVG(tokens_used) as avg_tokens_per_req,
			AVG(error_count) as error_rate,
			AVG(CAST(was_winner AS Float32)) as win_rate
		FROM debate_metrics
		WHERE provider = ?
		  AND timestamp >= now() - INTERVAL ? %s
		GROUP BY provider, period_start
		ORDER BY period_start DESC
		LIMIT ?
	`, interval, interval)

	rows, err := cha.conn.QueryContext(ctx, query, provider, periods, periods)
	if err != nil {
		return nil, fmt.Errorf("failed to query provider trends: %w", err)
	}
	defer rows.Close()

	var trends []ProviderStats
	for rows.Next() {
		var s ProviderStats
		var periodStart time.Time
		err := rows.Scan(
			&s.Provider,
			&periodStart,
			&s.TotalRequests,
			&s.AvgResponseTime,
			&s.P95ResponseTime,
			&s.P99ResponseTime,
			&s.AvgConfidence,
			&s.TotalTokens,
			&s.AvgTokensPerReq,
			&s.ErrorRate,
			&s.WinRate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		s.Period = periodStart.Format("2006-01-02 15:04:05")
		trends = append(trends, s)
	}

	return trends, nil
}

// StoreConversationMetrics stores conversation metrics
func (cha *ClickHouseAnalytics) StoreConversationMetrics(ctx context.Context, metrics ConversationMetrics) error {
	query := `
		INSERT INTO conversation_metrics (
			conversation_id, user_id, timestamp, message_count, entity_count,
			total_tokens, duration_ms, debate_rounds, llms_used
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := cha.conn.ExecContext(ctx, query,
		metrics.ConversationID,
		metrics.UserID,
		metrics.Timestamp,
		metrics.MessageCount,
		metrics.EntityCount,
		metrics.TotalTokens,
		metrics.DurationMs,
		metrics.DebateRounds,
		metrics.LLMsUsed,
	)

	if err != nil {
		return fmt.Errorf("failed to insert conversation metrics: %w", err)
	}

	return nil
}

// GetConversationTrends retrieves conversation trends
func (cha *ClickHouseAnalytics) GetConversationTrends(ctx context.Context, interval string, periods int) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		SELECT
			toStartOfInterval(timestamp, INTERVAL 1 %s) as period_start,
			COUNT(*) as total_conversations,
			AVG(message_count) as avg_messages,
			AVG(entity_count) as avg_entities,
			AVG(total_tokens) as avg_tokens,
			AVG(duration_ms) as avg_duration_ms,
			AVG(debate_rounds) as avg_debate_rounds
		FROM conversation_metrics
		WHERE timestamp >= now() - INTERVAL ? %s
		GROUP BY period_start
		ORDER BY period_start DESC
		LIMIT ?
	`, interval, interval)

	rows, err := cha.conn.QueryContext(ctx, query, periods, periods)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversation trends: %w", err)
	}
	defer rows.Close()

	var trends []map[string]interface{}
	for rows.Next() {
		var periodStart time.Time
		var totalConversations int64
		var avgMessages, avgEntities float64
		var avgTokens, avgDuration, avgDebateRounds float64

		err := rows.Scan(
			&periodStart,
			&totalConversations,
			&avgMessages,
			&avgEntities,
			&avgTokens,
			&avgDuration,
			&avgDebateRounds,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		trends = append(trends, map[string]interface{}{
			"period":               periodStart.Format("2006-01-02 15:04:05"),
			"total_conversations":  totalConversations,
			"avg_messages":         avgMessages,
			"avg_entities":         avgEntities,
			"avg_tokens":           avgTokens,
			"avg_duration_ms":      avgDuration,
			"avg_debate_rounds":    avgDebateRounds,
		})
	}

	return trends, nil
}

// GetTopProviders retrieves top performing providers
func (cha *ClickHouseAnalytics) GetTopProviders(ctx context.Context, limit int, sortBy string) ([]ProviderStats, error) {
	validSortFields := map[string]bool{
		"confidence": true,
		"winrate":    true,
		"speed":      true,
		"requests":   true,
	}

	if !validSortFields[sortBy] {
		sortBy = "confidence" // Default
	}

	var orderClause string
	switch sortBy {
	case "confidence":
		orderClause = "avg_confidence DESC"
	case "winrate":
		orderClause = "win_rate DESC"
	case "speed":
		orderClause = "avg_response_time ASC"
	case "requests":
		orderClause = "total_requests DESC"
	}

	query := fmt.Sprintf(`
		SELECT
			provider,
			COUNT(*) as total_requests,
			AVG(response_time_ms) as avg_response_time,
			quantile(0.95)(response_time_ms) as p95_response_time,
			quantile(0.99)(response_time_ms) as p99_response_time,
			AVG(confidence_score) as avg_confidence,
			SUM(tokens_used) as total_tokens,
			AVG(tokens_used) as avg_tokens_per_req,
			AVG(error_count) as error_rate,
			AVG(CAST(was_winner AS Float32)) as win_rate
		FROM debate_metrics
		WHERE timestamp >= now() - INTERVAL 1 DAY
		GROUP BY provider
		ORDER BY %s
		LIMIT ?
	`, orderClause)

	rows, err := cha.conn.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top providers: %w", err)
	}
	defer rows.Close()

	var providers []ProviderStats
	for rows.Next() {
		var s ProviderStats
		err := rows.Scan(
			&s.Provider,
			&s.TotalRequests,
			&s.AvgResponseTime,
			&s.P95ResponseTime,
			&s.P99ResponseTime,
			&s.AvgConfidence,
			&s.TotalTokens,
			&s.AvgTokensPerReq,
			&s.ErrorRate,
			&s.WinRate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		s.Period = "last_24h"
		providers = append(providers, s)
	}

	return providers, nil
}

// GetDebateAnalytics retrieves debate-specific analytics
func (cha *ClickHouseAnalytics) GetDebateAnalytics(ctx context.Context, debateID string) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(DISTINCT round) as total_rounds,
			AVG(response_time_ms) as avg_response_time,
			MAX(response_time_ms) as max_response_time,
			SUM(tokens_used) as total_tokens,
			AVG(confidence_score) as avg_confidence,
			SUM(error_count) as total_errors
		FROM debate_metrics
		WHERE debate_id = ?
	`

	var totalRounds int64
	var avgResponseTime, maxResponseTime float32
	var totalTokens int64
	var avgConfidence float32
	var totalErrors int

	err := cha.conn.QueryRowContext(ctx, query, debateID).Scan(
		&totalRounds,
		&avgResponseTime,
		&maxResponseTime,
		&totalTokens,
		&avgConfidence,
		&totalErrors,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query debate analytics: %w", err)
	}

	return map[string]interface{}{
		"debate_id":         debateID,
		"total_rounds":      totalRounds,
		"avg_response_time": avgResponseTime,
		"max_response_time": maxResponseTime,
		"total_tokens":      totalTokens,
		"avg_confidence":    avgConfidence,
		"total_errors":      totalErrors,
	}, nil
}

// Close closes the ClickHouse connection
func (cha *ClickHouseAnalytics) Close() error {
	if cha.conn != nil {
		return cha.conn.Close()
	}
	return nil
}
